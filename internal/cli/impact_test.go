package cli

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/jysf/bragfile000/internal/storage"
	"github.com/spf13/cobra"

	_ "modernc.org/sqlite"
)

// newImpactTestRoot builds a fresh root command with the impact
// subcommand attached and isolates BRAGFILE_DB so the host env can't
// leak in. The caller drives args via cmd.SetArgs.
func newImpactTestRoot(t *testing.T) (*cobra.Command, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	t.Setenv("BRAGFILE_DB", "")
	root := NewRootCmd("test")
	root.AddCommand(NewImpactCmd())
	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	root.SetOut(outBuf)
	root.SetErr(errBuf)
	return root, outBuf, errBuf
}

func runImpactCmd(t *testing.T, dbPath string, args ...string) (stdout, stderr string, runErr error) {
	t.Helper()
	root, outBuf, errBuf := newImpactTestRoot(t)
	full := append([]string{"--db", dbPath, "impact"}, args...)
	root.SetArgs(full)
	runErr = root.Execute()
	return outBuf.String(), errBuf.String(), runErr
}

// seedImpactEntry adds an entry via the store, then rewrites its
// created_at to the given instant via a raw connection (the store's
// Add always stamps time.Now()). Raw SQL is confined to the test file;
// the production CLI layer stays SQL-free (no-sql-in-cli-layer).
func seedImpactEntry(t *testing.T, dbPath string, e storage.Entry, createdAt time.Time) {
	t.Helper()
	s, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	added, err := s.Add(e)
	s.Close()
	if err != nil {
		t.Fatalf("add entry %q: %v", e.Title, err)
	}
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open raw db: %v", err)
	}
	defer db.Close()
	ts := createdAt.UTC().Format(time.RFC3339)
	if _, err := db.Exec(
		`UPDATE entries SET created_at = ?, updated_at = ? WHERE id = ?`,
		ts, ts, added.ID,
	); err != nil {
		t.Fatalf("rewrite created_at for id %d: %v", added.ID, err)
	}
}

// withNowFunc swaps the package clock seam for the duration of a test.
func withNowFunc(t *testing.T, now time.Time) {
	t.Helper()
	prev := nowFunc
	nowFunc = func() time.Time { return now }
	t.Cleanup(func() { nowFunc = prev })
}

// Test 9 — TestImpactCmd_RequiresExactlyOneWindow pairs locked
// decision 2 (exactly one window flag, mutually exclusive, required).
func TestImpactCmd_RequiresExactlyOneWindow(t *testing.T) {
	t.Run("none", func(t *testing.T) {
		dbPath := filepath.Join(t.TempDir(), "test.db")
		out, errStr, err := runImpactCmd(t, dbPath)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, ErrUser) {
			t.Errorf("expected errors.Is(err, ErrUser); got %v", err)
		}
		if out != "" {
			t.Errorf("expected empty stdout, got %q", out)
		}
		for _, needle := range []string{"--quarter", "--month", "--year", "--since"} {
			if !bytes.Contains([]byte(err.Error()), []byte(needle)) {
				t.Errorf("expected error to name %q, got %q", needle, err.Error())
			}
		}
		_ = errStr
	})

	t.Run("two", func(t *testing.T) {
		dbPath := filepath.Join(t.TempDir(), "test.db")
		out, _, err := runImpactCmd(t, dbPath, "--quarter", "--month")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, ErrUser) {
			t.Errorf("expected errors.Is(err, ErrUser); got %v", err)
		}
		if out != "" {
			t.Errorf("expected empty stdout, got %q", out)
		}
		if !bytes.Contains([]byte(err.Error()), []byte("mutually exclusive")) {
			t.Errorf("expected error to mention mutual exclusion, got %q", err.Error())
		}
	})
}

// Test 10 — TestImpactCmd_CalendarNotRolling (LOAD-BEARING divergence).
// The quarter-start entry is IN; the day-before entry is OUT. This
// fails under a rolling now-90d reading when now is early in the quarter.
func TestImpactCmd_CalendarNotRolling(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")

	// now = 2026-07-06: calendar Q3 start is 2026-07-01T00:00:00Z.
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	withNowFunc(t, now)

	qStart := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	dayBefore := qStart.AddDate(0, 0, -1)

	seedImpactEntry(t, dbPath, storage.Entry{
		Title: "in-quarter", Project: "alpha", Type: "shipped",
		Impact: "landed at the boundary",
	}, qStart)
	seedImpactEntry(t, dbPath, storage.Entry{
		Title: "before-quarter", Project: "alpha", Type: "shipped",
		Impact: "should be excluded",
	}, dayBefore)

	out, errStr, err := runImpactCmd(t, dbPath, "--quarter")
	if err != nil {
		t.Fatalf("unexpected error: %v (stderr: %s)", err, errStr)
	}
	if !bytes.Contains([]byte(out), []byte("in-quarter")) {
		t.Errorf("quarter-start entry must be IN the calendar window:\n%s", out)
	}
	if bytes.Contains([]byte(out), []byte("before-quarter")) {
		t.Errorf("day-before-quarter entry must be OUT (calendar, not rolling):\n%s", out)
	}
}

// Test 11 — TestImpactCmd_SinceReusesParseSince pairs locked decision 1
// (--since reuses ParseSince; scope echoes "since:<raw>").
func TestImpactCmd_SinceReusesParseSince(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	withNowFunc(t, now)

	seedImpactEntry(t, dbPath, storage.Entry{
		Title: "after-cutoff", Project: "alpha", Type: "shipped",
		Impact: "in range",
	}, time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC))
	seedImpactEntry(t, dbPath, storage.Entry{
		Title: "before-cutoff", Project: "alpha", Type: "shipped",
		Impact: "out of range",
	}, time.Date(2025, 12, 31, 10, 0, 0, 0, time.UTC))

	t.Run("valid", func(t *testing.T) {
		out, _, err := runImpactCmd(t, dbPath, "--since", "2026-01-01", "--format", "json")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		var env struct {
			Scope string `json:"scope"`
		}
		if err := json.Unmarshal([]byte(out), &env); err != nil {
			t.Fatalf("json unmarshal: %v\n%s", err, out)
		}
		if env.Scope != "since:2026-01-01" {
			t.Errorf("scope: got %q, want %q", env.Scope, "since:2026-01-01")
		}
		if !bytes.Contains([]byte(out), []byte("after-cutoff")) {
			t.Errorf("expected 'after-cutoff' (>= 2026-01-01) in output:\n%s", out)
		}
		if bytes.Contains([]byte(out), []byte("before-cutoff")) {
			t.Errorf("did not expect 'before-cutoff' (< 2026-01-01) in output:\n%s", out)
		}
	})

	t.Run("bad", func(t *testing.T) {
		out, _, err := runImpactCmd(t, dbPath, "--since", "not-a-date")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, ErrUser) {
			t.Errorf("expected errors.Is(err, ErrUser); got %v", err)
		}
		if out != "" {
			t.Errorf("expected empty stdout, got %q", out)
		}
	})
}

// Test 12 — TestImpactCmd_FormatDefaultAndUnknown pairs locked decision
// 7 (--format defaults to markdown; unknown → UserError).
func TestImpactCmd_FormatDefaultAndUnknown(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	withNowFunc(t, now)

	t.Run("default-markdown", func(t *testing.T) {
		out, _, err := runImpactCmd(t, dbPath, "--quarter")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !bytes.Contains([]byte(out), []byte("# Bragfile Impact")) {
			t.Errorf("expected markdown header in default output:\n%s", out)
		}
	})

	t.Run("json", func(t *testing.T) {
		out, _, err := runImpactCmd(t, dbPath, "--quarter", "--format", "json")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		var env map[string]json.RawMessage
		if err := json.Unmarshal([]byte(out), &env); err != nil {
			t.Fatalf("expected JSON output, unmarshal failed: %v\n%s", err, out)
		}
		if _, ok := env["scope"]; !ok {
			t.Errorf("expected a 'scope' key in JSON output:\n%s", out)
		}
	})

	t.Run("unknown", func(t *testing.T) {
		out, _, err := runImpactCmd(t, dbPath, "--quarter", "--format", "xml")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, ErrUser) {
			t.Errorf("expected errors.Is(err, ErrUser); got %v", err)
		}
		if out != "" {
			t.Errorf("expected empty stdout, got %q", out)
		}
	})
}

// Test 13 — TestImpactCmd_StdoutStderrSeparation pairs locked decision
// 9 (stdout-is-for-data-stderr-is-for-humans).
func TestImpactCmd_StdoutStderrSeparation(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	withNowFunc(t, now)

	t.Run("success-to-stdout-only", func(t *testing.T) {
		root, outBuf, errBuf := newImpactTestRoot(t)
		root.SetArgs([]string{"--db", dbPath, "impact", "--quarter"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if outBuf.Len() == 0 {
			t.Errorf("expected digest on stdout, got empty")
		}
		if errBuf.Len() != 0 {
			t.Errorf("expected empty stderr on success, got %q", errBuf.String())
		}
	})

	t.Run("usererror-keeps-stdout-clean", func(t *testing.T) {
		// The root sets SilenceErrors (main.go owns stderr formatting +
		// exit-code mapping), so the UserError rides the returned error
		// rather than errBuf here. The stdout/stderr contract this test
		// enforces is: a failed run writes NOTHING to stdout — the digest
		// stream stays clean, and the error is a UserError main.go routes
		// to stderr.
		root, outBuf, _ := newImpactTestRoot(t)
		root.SetArgs([]string{"--db", dbPath, "impact"}) // no window → UserError
		err := root.Execute()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, ErrUser) {
			t.Errorf("expected errors.Is(err, ErrUser); got %v", err)
		}
		if outBuf.Len() != 0 {
			t.Errorf("expected empty stdout on error, got %q", outBuf.String())
		}
	})
}

// TestWindowCutoff_CalendarArithmetic is the pure-helper unit test for
// the calendar window math (mirrors TestRangeCutoff_* for summary).
func TestWindowCutoff_CalendarArithmetic(t *testing.T) {
	now := time.Date(2026, 8, 15, 12, 0, 0, 0, time.UTC) // Q3, August

	t.Run("quarter", func(t *testing.T) {
		cutoff, end, scope, err := windowCutoff("quarter", "", now, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
		if !cutoff.Equal(want) {
			t.Errorf("quarter cutoff: got %v, want %v", cutoff, want)
		}
		if !end.IsZero() {
			t.Errorf("quarter (current) end must be the zero sentinel, got %v", end)
		}
		if scope != "quarter" {
			t.Errorf("quarter scope: got %q, want %q", scope, "quarter")
		}
	})

	t.Run("month", func(t *testing.T) {
		cutoff, end, scope, err := windowCutoff("month", "", now, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
		if !cutoff.Equal(want) {
			t.Errorf("month cutoff: got %v, want %v", cutoff, want)
		}
		if !end.IsZero() {
			t.Errorf("month (current) end must be the zero sentinel, got %v", end)
		}
		if scope != "month" {
			t.Errorf("month scope: got %q, want %q", scope, "month")
		}
	})

	t.Run("year", func(t *testing.T) {
		cutoff, end, scope, err := windowCutoff("year", "", now, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		if !cutoff.Equal(want) {
			t.Errorf("year cutoff: got %v, want %v", cutoff, want)
		}
		if !end.IsZero() {
			t.Errorf("year (current) end must be the zero sentinel, got %v", end)
		}
		if scope != "year" {
			t.Errorf("year scope: got %q, want %q", scope, "year")
		}
	})

	t.Run("since", func(t *testing.T) {
		cutoff, end, scope, err := windowCutoff("since", "2026-01-01", now, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		if !cutoff.Equal(want) {
			t.Errorf("since cutoff: got %v, want %v", cutoff, want)
		}
		if !end.IsZero() {
			t.Errorf("since end must be the zero sentinel, got %v", end)
		}
		if scope != "since:2026-01-01" {
			t.Errorf("since scope: got %q, want %q", scope, "since:2026-01-01")
		}
	})

	// Q1/Q2/Q4 boundaries exercise the qStartMonth formula.
	for _, tc := range []struct {
		now  time.Time
		want time.Time
	}{
		{time.Date(2026, 2, 10, 0, 0, 0, 0, time.UTC), time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
		{time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC), time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)},
		{time.Date(2026, 11, 10, 0, 0, 0, 0, time.UTC), time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)},
	} {
		cutoff, _, _, err := windowCutoff("quarter", "", tc.now, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !cutoff.Equal(tc.want) {
			t.Errorf("quarter cutoff for %v: got %v, want %v", tc.now, cutoff, tc.want)
		}
	}
}

// spec53Now is the canonical frozen instant for SPEC-053 (--previous):
// mid-Q3-2026 (July), so every window has a well-defined, non-degenerate
// previous period within 2026 except --year (which lands in 2025).
var spec53Now = time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)

// TestWindowCutoff_PreviousBoundaries (LOAD-BEARING — the boundary-math
// core, SPEC-053 / DEC-032). Asserts windowCutoff(..., previous=true)
// returns exactly the [start, end) pairs and "<window>:previous" scope
// tokens at both spec53Now and the January year-boundary instant (the
// AddDate roll), and that previous=false is unchanged (bounded=false).
func TestWindowCutoff_PreviousBoundaries(t *testing.T) {
	janRoll := time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)

	cases := []struct {
		name      string
		window    string
		now       time.Time
		wantStart time.Time
		wantEnd   time.Time
		wantScope string
	}{
		// spec53Now (mid-Q3-2026).
		{"quarter@Q3", "quarter", spec53Now, date(2026, 4, 1), date(2026, 7, 1), "quarter:previous"},
		{"month@Jul", "month", spec53Now, date(2026, 6, 1), date(2026, 7, 1), "month:previous"},
		{"year@2026", "year", spec53Now, date(2025, 1, 1), date(2026, 1, 1), "year:previous"},
		// January year-boundary roll.
		{"quarter@Jan", "quarter", janRoll, date(2025, 10, 1), date(2026, 1, 1), "quarter:previous"},
		{"month@Jan", "month", janRoll, date(2025, 12, 1), date(2026, 1, 1), "month:previous"},
		{"year@Jan", "year", janRoll, date(2025, 1, 1), date(2026, 1, 1), "year:previous"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			start, end, scope, err := windowCutoff(tc.window, "", tc.now, true)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !start.Equal(tc.wantStart) {
				t.Errorf("start: got %v, want %v", start, tc.wantStart)
			}
			if !end.Equal(tc.wantEnd) {
				t.Errorf("end: got %v, want %v", end, tc.wantEnd)
			}
			if end.IsZero() {
				t.Errorf("previous window must be bounded (non-zero end)")
			}
			if scope != tc.wantScope {
				t.Errorf("scope: got %q, want %q", scope, tc.wantScope)
			}
		})
	}

	// previous=false is unchanged: current-period cutoff, zero (open) end.
	for _, window := range []string{"quarter", "month", "year"} {
		t.Run("additive/"+window, func(t *testing.T) {
			_, end, scope, err := windowCutoff(window, "", spec53Now, false)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !end.IsZero() {
				t.Errorf("current-period end must be the zero sentinel, got %v", end)
			}
			if scope != window {
				t.Errorf("current scope: got %q, want %q", scope, window)
			}
		})
	}
}

// date is a UTC midnight helper for the boundary tables.
func date(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// TestWindowCutoff_PreviousSinceIsUserError pins the helper-level guard:
// --previous is undefined for --since (LD3). The CLI also rejects the flag
// combo, but this backs it up so a future caller can't bypass it.
func TestWindowCutoff_PreviousSinceIsUserError(t *testing.T) {
	_, _, _, err := windowCutoff("since", "2026-01-01", spec53Now, true)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrUser) {
		t.Errorf("expected errors.Is(err, ErrUser); got %v", err)
	}
}

// TestImpactCmd_PreviousQuarterBounded (LOAD-BEARING — the bounded-window
// divergence). At spec53Now, --quarter --previous includes ONLY the two
// prev-Q2 entries; the current-Q3 entry (created "now-ish") is EXCLUDED,
// proving the upper bound is the current period start, not now.
func TestImpactCmd_PreviousQuarterBounded(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	withNowFunc(t, spec53Now)

	seedImpactEntry(t, dbPath, storage.Entry{
		Title: "q1-before", Project: "alpha", Type: "shipped", Impact: "before prev-Q2",
	}, time.Date(2026, 3, 31, 10, 0, 0, 0, time.UTC))
	seedImpactEntry(t, dbPath, storage.Entry{
		Title: "q2-start", Project: "alpha", Type: "shipped", Impact: "prev-Q2 start",
	}, time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC))
	seedImpactEntry(t, dbPath, storage.Entry{
		Title: "q2-end", Project: "alpha", Type: "shipped", Impact: "prev-Q2 end",
	}, time.Date(2026, 6, 30, 23, 0, 0, 0, time.UTC))
	seedImpactEntry(t, dbPath, storage.Entry{
		Title: "q3-start", Project: "alpha", Type: "shipped", Impact: "current Q3, excluded",
	}, time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC))

	out, errStr, err := runImpactCmd(t, dbPath, "--quarter", "--previous", "--format", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v (stderr: %s)", err, errStr)
	}
	var env struct {
		Scope           string `json:"scope"`
		EntriesInWindow int    `json:"entries_in_window"`
	}
	if err := json.Unmarshal([]byte(out), &env); err != nil {
		t.Fatalf("json unmarshal: %v\n%s", err, out)
	}
	if env.Scope != "quarter:previous" {
		t.Errorf("scope: got %q, want %q", env.Scope, "quarter:previous")
	}
	if env.EntriesInWindow != 2 {
		t.Errorf("entries_in_window: got %d, want 2\n%s", env.EntriesInWindow, out)
	}
	for _, in := range []string{"q2-start", "q2-end"} {
		if !bytes.Contains([]byte(out), []byte(in)) {
			t.Errorf("expected prev-Q2 entry %q IN the window:\n%s", in, out)
		}
	}
	for _, out2 := range []string{"q1-before", "q3-start"} {
		if bytes.Contains([]byte(out), []byte(out2)) {
			t.Errorf("entry %q must be OUT of the bounded prev-Q2 window:\n%s", out2, out)
		}
	}
}

// TestImpactCmd_PreviousMonthAndYear covers --month/--year --previous at
// spec53Now with a below/in/above entry each.
func TestImpactCmd_PreviousMonthAndYear(t *testing.T) {
	t.Run("month", func(t *testing.T) {
		dbPath := filepath.Join(t.TempDir(), "test.db")
		withNowFunc(t, spec53Now)
		seedImpactEntry(t, dbPath, storage.Entry{Title: "may-out", Project: "a", Type: "shipped", Impact: "x"}, time.Date(2026, 5, 31, 10, 0, 0, 0, time.UTC))
		seedImpactEntry(t, dbPath, storage.Entry{Title: "jun-in", Project: "a", Type: "shipped", Impact: "x"}, time.Date(2026, 6, 15, 10, 0, 0, 0, time.UTC))
		seedImpactEntry(t, dbPath, storage.Entry{Title: "jul-out", Project: "a", Type: "shipped", Impact: "x"}, time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC))
		out, _, err := runImpactCmd(t, dbPath, "--month", "--previous", "--format", "json")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertScope(t, out, "month:previous")
		assertContainsAll(t, out, "jun-in")
		assertContainsNone(t, out, "may-out", "jul-out")
	})

	t.Run("year", func(t *testing.T) {
		dbPath := filepath.Join(t.TempDir(), "test.db")
		withNowFunc(t, spec53Now)
		seedImpactEntry(t, dbPath, storage.Entry{Title: "y2024-out", Project: "a", Type: "shipped", Impact: "x"}, time.Date(2024, 12, 31, 10, 0, 0, 0, time.UTC))
		seedImpactEntry(t, dbPath, storage.Entry{Title: "y2025-in", Project: "a", Type: "shipped", Impact: "x"}, time.Date(2025, 7, 1, 10, 0, 0, 0, time.UTC))
		seedImpactEntry(t, dbPath, storage.Entry{Title: "y2026-out", Project: "a", Type: "shipped", Impact: "x"}, time.Date(2026, 1, 2, 10, 0, 0, 0, time.UTC))
		out, _, err := runImpactCmd(t, dbPath, "--year", "--previous", "--format", "json")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertScope(t, out, "year:previous")
		assertContainsAll(t, out, "y2025-in")
		assertContainsNone(t, out, "y2024-out", "y2026-out")
	})
}

// TestImpactCmd_PreviousYearBoundaryRoll: at 2026-01-15, --month --previous
// is [2025-12-01, 2026-01-01) — AddDate rolls the month back across the year
// boundary (Dec of the prior year), not "month 0".
func TestImpactCmd_PreviousYearBoundaryRoll(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	withNowFunc(t, time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC))
	seedImpactEntry(t, dbPath, storage.Entry{Title: "dec-in", Project: "a", Type: "shipped", Impact: "x"}, time.Date(2025, 12, 20, 10, 0, 0, 0, time.UTC))
	seedImpactEntry(t, dbPath, storage.Entry{Title: "jan-out", Project: "a", Type: "shipped", Impact: "x"}, time.Date(2026, 1, 5, 10, 0, 0, 0, time.UTC))
	out, _, err := runImpactCmd(t, dbPath, "--month", "--previous", "--format", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertScope(t, out, "month:previous")
	assertContainsAll(t, out, "dec-in")
	assertContainsNone(t, out, "jan-out")
}

// TestImpactCmd_PreviousWithSinceIsUserError: --since + --previous → UserError.
func TestImpactCmd_PreviousWithSinceIsUserError(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	withNowFunc(t, spec53Now)
	out, _, err := runImpactCmd(t, dbPath, "--since", "2026-01-01", "--previous")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrUser) {
		t.Errorf("expected errors.Is(err, ErrUser); got %v", err)
	}
	for _, needle := range []string{"--previous", "--since"} {
		if !bytes.Contains([]byte(err.Error()), []byte(needle)) {
			t.Errorf("expected error to name %q, got %q", needle, err.Error())
		}
	}
	if out != "" {
		t.Errorf("expected empty stdout, got %q", out)
	}
}

// TestImpactCmd_PreviousWithoutWindowIsUserError: --previous with no window
// flag → the existing "exactly one window required" UserError (a modifier is
// not a window).
func TestImpactCmd_PreviousWithoutWindowIsUserError(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	withNowFunc(t, spec53Now)
	out, _, err := runImpactCmd(t, dbPath, "--previous")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrUser) {
		t.Errorf("expected errors.Is(err, ErrUser); got %v", err)
	}
	if out != "" {
		t.Errorf("expected empty stdout, got %q", out)
	}
}

// TestImpactCmd_NoPreviousUnchanged (regression guard): --quarter (no
// --previous) at spec53Now still reports the CURRENT quarter up to now, and
// scope echoes plain "quarter".
func TestImpactCmd_NoPreviousUnchanged(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	withNowFunc(t, spec53Now)
	seedImpactEntry(t, dbPath, storage.Entry{Title: "cur-q3", Project: "a", Type: "shipped", Impact: "x"}, time.Date(2026, 7, 5, 10, 0, 0, 0, time.UTC))
	seedImpactEntry(t, dbPath, storage.Entry{Title: "prev-q2", Project: "a", Type: "shipped", Impact: "x"}, time.Date(2026, 6, 30, 10, 0, 0, 0, time.UTC))
	out, _, err := runImpactCmd(t, dbPath, "--quarter", "--format", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertScope(t, out, "quarter")
	assertContainsAll(t, out, "cur-q3")
	assertContainsNone(t, out, "prev-q2")
}

// TestImpactCmd_HelpShowsPrevious: --help contains --previous and a
// distinctive example line.
func TestImpactCmd_HelpShowsPrevious(t *testing.T) {
	root, outBuf, _ := newImpactTestRoot(t)
	root.SetArgs([]string{"impact", "--help"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := outBuf.String()
	if !bytes.Contains([]byte(out), []byte("--previous")) {
		t.Errorf("expected --previous in help:\n%s", out)
	}
	if !bytes.Contains([]byte(out), []byte("brag impact --quarter --previous")) {
		t.Errorf("expected example line 'brag impact --quarter --previous' in help:\n%s", out)
	}
}

// TestImpactCmd_StdoutStderrSeparation_Previous: a successful --previous run
// writes only stdout; the --since --previous combo writes only the returned
// UserError (main.go routes it to stderr) with empty stdout.
func TestImpactCmd_StdoutStderrSeparation_Previous(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	withNowFunc(t, spec53Now)

	t.Run("success-to-stdout-only", func(t *testing.T) {
		root, outBuf, errBuf := newImpactTestRoot(t)
		root.SetArgs([]string{"--db", dbPath, "impact", "--quarter", "--previous"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if outBuf.Len() == 0 {
			t.Errorf("expected digest on stdout, got empty")
		}
		if errBuf.Len() != 0 {
			t.Errorf("expected empty stderr on success, got %q", errBuf.String())
		}
	})

	t.Run("incoherent-combo-keeps-stdout-clean", func(t *testing.T) {
		root, outBuf, _ := newImpactTestRoot(t)
		root.SetArgs([]string{"--db", dbPath, "impact", "--since", "2026-01-01", "--previous"})
		err := root.Execute()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, ErrUser) {
			t.Errorf("expected errors.Is(err, ErrUser); got %v", err)
		}
		if outBuf.Len() != 0 {
			t.Errorf("expected empty stdout on error, got %q", outBuf.String())
		}
	})
}

// assertScope unmarshals the JSON envelope and checks scope.
func assertScope(t *testing.T, out, want string) {
	t.Helper()
	var env struct {
		Scope string `json:"scope"`
	}
	if err := json.Unmarshal([]byte(out), &env); err != nil {
		t.Fatalf("json unmarshal: %v\n%s", err, out)
	}
	if env.Scope != want {
		t.Errorf("scope: got %q, want %q\n%s", env.Scope, want, out)
	}
}

func assertContainsAll(t *testing.T, out string, needles ...string) {
	t.Helper()
	for _, n := range needles {
		if !bytes.Contains([]byte(out), []byte(n)) {
			t.Errorf("expected %q IN output:\n%s", n, out)
		}
	}
}

func assertContainsNone(t *testing.T, out string, needles ...string) {
	t.Helper()
	for _, n := range needles {
		if bytes.Contains([]byte(out), []byte(n)) {
			t.Errorf("expected %q OUT of output:\n%s", n, out)
		}
	}
}
