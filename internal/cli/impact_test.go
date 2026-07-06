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
		cutoff, scope, err := windowCutoff("quarter", "", now)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
		if !cutoff.Equal(want) {
			t.Errorf("quarter cutoff: got %v, want %v", cutoff, want)
		}
		if scope != "quarter" {
			t.Errorf("quarter scope: got %q, want %q", scope, "quarter")
		}
	})

	t.Run("month", func(t *testing.T) {
		cutoff, scope, err := windowCutoff("month", "", now)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
		if !cutoff.Equal(want) {
			t.Errorf("month cutoff: got %v, want %v", cutoff, want)
		}
		if scope != "month" {
			t.Errorf("month scope: got %q, want %q", scope, "month")
		}
	})

	t.Run("year", func(t *testing.T) {
		cutoff, scope, err := windowCutoff("year", "", now)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		if !cutoff.Equal(want) {
			t.Errorf("year cutoff: got %v, want %v", cutoff, want)
		}
		if scope != "year" {
			t.Errorf("year scope: got %q, want %q", scope, "year")
		}
	})

	t.Run("since", func(t *testing.T) {
		cutoff, scope, err := windowCutoff("since", "2026-01-01", now)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		if !cutoff.Equal(want) {
			t.Errorf("since cutoff: got %v, want %v", cutoff, want)
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
		cutoff, _, err := windowCutoff("quarter", "", tc.now)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !cutoff.Equal(tc.want) {
			t.Errorf("quarter cutoff for %v: got %v, want %v", tc.now, cutoff, tc.want)
		}
	}
}
