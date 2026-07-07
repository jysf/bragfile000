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

// newWrappedTestRoot builds a fresh root command with the wrapped
// subcommand attached and isolates BRAGFILE_DB so the host env can't
// leak in. Mirrors newImpactTestRoot.
func newWrappedTestRoot(t *testing.T) (*cobra.Command, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	t.Setenv("BRAGFILE_DB", "")
	root := NewRootCmd("test")
	root.AddCommand(NewWrappedCmd())
	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	root.SetOut(outBuf)
	root.SetErr(errBuf)
	return root, outBuf, errBuf
}

func runWrappedCmd(t *testing.T, dbPath string, args ...string) (stdout, stderr string, runErr error) {
	t.Helper()
	root, outBuf, errBuf := newWrappedTestRoot(t)
	full := append([]string{"--db", dbPath, "wrapped"}, args...)
	root.SetArgs(full)
	runErr = root.Execute()
	return outBuf.String(), errBuf.String(), runErr
}

// seedWrappedEntry adds an entry via the store, then rewrites its
// created_at to the given instant via a raw connection (the store's Add
// always stamps time.Now()). Raw SQL is confined to the test file; the
// production CLI layer stays SQL-free (no-sql-in-cli-layer). Mirrors
// seedImpactEntry.
func seedWrappedEntry(t *testing.T, dbPath string, e storage.Entry, createdAt time.Time) {
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

func TestWrappedCmd_DefaultsToCurrentYear(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	withNowFunc(t, time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC))

	seedWrappedEntry(t, dbPath, storage.Entry{
		Title: "jan", Project: "alpha", Type: "shipped",
	}, time.Date(2026, 1, 10, 10, 0, 0, 0, time.UTC))

	out, errStr, err := runWrappedCmd(t, dbPath)
	if err != nil {
		t.Fatalf("unexpected error: %v (stderr: %s)", err, errStr)
	}
	if !bytes.Contains([]byte(out), []byte("Scope: 2026")) {
		t.Errorf("default period must be the current calendar year (Scope: 2026):\n%s", out)
	}
}

func TestWrappedCmd_NamedYearAndQuarter(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	withNowFunc(t, time.Date(2027, 6, 1, 12, 0, 0, 0, time.UTC))

	t.Run("year", func(t *testing.T) {
		out, _, err := runWrappedCmd(t, dbPath, "2026")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !bytes.Contains([]byte(out), []byte("Scope: 2026\n")) {
			t.Errorf("expected Scope: 2026:\n%s", out)
		}
	})

	t.Run("quarter", func(t *testing.T) {
		out, _, err := runWrappedCmd(t, dbPath, "2026", "Q3")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !bytes.Contains([]byte(out), []byte("Scope: 2026-Q3")) {
			t.Errorf("expected Scope: 2026-Q3:\n%s", out)
		}
	})

	t.Run("quarter-lowercase", func(t *testing.T) {
		out, _, err := runWrappedCmd(t, dbPath, "2026", "q3")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !bytes.Contains([]byte(out), []byte("Scope: 2026-Q3")) {
			t.Errorf("case-insensitive q3 must be accepted:\n%s", out)
		}
	})
}

// TestWrappedCmd_BoundedWindow (LOAD-BEARING — the bounded-window
// divergence from impact). The upper bound is the period END, not now.
func TestWrappedCmd_BoundedWindow(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	// now is AFTER the named year, so [start, now] would wrongly sweep in
	// everything up to 2027-06-01. The bounded window must exclude it.
	withNowFunc(t, time.Date(2027, 6, 1, 12, 0, 0, 0, time.UTC))

	seedWrappedEntry(t, dbPath, storage.Entry{
		Title: "before-period", Project: "alpha", Type: "shipped",
	}, time.Date(2025, 12, 31, 10, 0, 0, 0, time.UTC))
	seedWrappedEntry(t, dbPath, storage.Entry{
		Title: "in-period", Project: "alpha", Type: "shipped",
	}, time.Date(2026, 6, 15, 10, 0, 0, 0, time.UTC))
	seedWrappedEntry(t, dbPath, storage.Entry{
		Title: "after-period", Project: "alpha", Type: "shipped",
	}, time.Date(2027, 1, 1, 10, 0, 0, 0, time.UTC))

	out, errStr, err := runWrappedCmd(t, dbPath, "2026")
	if err != nil {
		t.Fatalf("unexpected error: %v (stderr: %s)", err, errStr)
	}
	if !bytes.Contains([]byte(out), []byte("Entries: 1")) {
		t.Errorf("bounded 2026 window must contain exactly the mid entry:\n%s", out)
	}
	if bytes.Contains([]byte(out), []byte("before-period")) {
		t.Errorf("entry before the period start must be excluded:\n%s", out)
	}
	if bytes.Contains([]byte(out), []byte("after-period")) {
		t.Errorf("entry after the period end must be excluded (upper bound is period end, not now):\n%s", out)
	}
}

func TestWrappedCmd_QuarterBoundaries(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	withNowFunc(t, time.Date(2027, 6, 1, 12, 0, 0, 0, time.UTC))

	seedWrappedEntry(t, dbPath, storage.Entry{
		Title: "q2-end", Project: "alpha", Type: "shipped",
	}, time.Date(2026, 6, 30, 10, 0, 0, 0, time.UTC))
	seedWrappedEntry(t, dbPath, storage.Entry{
		Title: "q3-start", Project: "alpha", Type: "shipped", Impact: "in",
	}, time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC))
	seedWrappedEntry(t, dbPath, storage.Entry{
		Title: "q3-end", Project: "alpha", Type: "shipped", Impact: "in",
	}, time.Date(2026, 9, 30, 23, 0, 0, 0, time.UTC))
	seedWrappedEntry(t, dbPath, storage.Entry{
		Title: "q4-start", Project: "alpha", Type: "shipped",
	}, time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC))

	out, errStr, err := runWrappedCmd(t, dbPath, "2026", "Q3")
	if err != nil {
		t.Fatalf("unexpected error: %v (stderr: %s)", err, errStr)
	}
	if !bytes.Contains([]byte(out), []byte("Entries: 2")) {
		t.Errorf("Q3 window must contain exactly the Jul-1 and Sep-30 entries:\n%s", out)
	}
	if bytes.Contains([]byte(out), []byte("q2-end")) || bytes.Contains([]byte(out), []byte("q4-start")) {
		t.Errorf("Q2-end and Q4-start entries must be excluded from Q3:\n%s", out)
	}
}

func TestWrappedCmd_MalformedPeriodIsUserError(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	withNowFunc(t, time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC))

	cases := [][]string{
		{"999"},
		{"2026", "Q0"},
		{"2026", "Q5"},
		{"2026", "Q3", "extra"},
		{"notayear"},
	}
	for _, args := range cases {
		t.Run(joinArgs(args), func(t *testing.T) {
			out, _, err := runWrappedCmd(t, dbPath, args...)
			if err == nil {
				t.Fatalf("expected error for args %v, got nil", args)
			}
			if !errors.Is(err, ErrUser) {
				t.Errorf("expected a UserError for args %v, got %v", args, err)
			}
			if out != "" {
				t.Errorf("expected empty stdout for args %v, got %q", args, out)
			}
		})
	}
}

func joinArgs(args []string) string {
	if len(args) == 0 {
		return "empty"
	}
	s := args[0]
	for _, a := range args[1:] {
		s += "_" + a
	}
	return s
}

func TestWrappedCmd_UnknownFormatIsUserError(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	withNowFunc(t, time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC))

	out, _, err := runWrappedCmd(t, dbPath, "2026", "--format", "yaml")
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

func TestWrappedCmd_JSONWiring(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	withNowFunc(t, time.Date(2027, 6, 1, 12, 0, 0, 0, time.UTC))

	seedWrappedEntry(t, dbPath, storage.Entry{
		Title: "one", Project: "alpha", Type: "shipped", Tags: "api",
	}, time.Date(2026, 3, 10, 10, 0, 0, 0, time.UTC))

	out, _, err := runWrappedCmd(t, dbPath, "2026", "--format", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var env struct {
		Scope   string `json:"scope"`
		Cadence struct {
			Series []struct {
				Period string `json:"period"`
				Count  int    `json:"count"`
			} `json:"series"`
		} `json:"cadence"`
	}
	if err := json.Unmarshal([]byte(out), &env); err != nil {
		t.Fatalf("json unmarshal: %v\n%s", err, out)
	}
	if env.Scope != "2026" {
		t.Errorf("scope: got %q, want %q", env.Scope, "2026")
	}
	if len(env.Cadence.Series) != 12 {
		t.Errorf("cadence.series length: got %d, want 12", len(env.Cadence.Series))
	}
}

func TestWrappedCmd_HelpShowsExamples(t *testing.T) {
	root, outBuf, _ := newWrappedTestRoot(t)
	root.SetArgs([]string{"wrapped", "--help"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := outBuf.String()
	if !bytes.Contains([]byte(out), []byte("Examples:")) {
		t.Errorf("expected an Examples: label in help:\n%s", out)
	}
	if !bytes.Contains([]byte(out), []byte("brag wrapped 2026 Q3")) {
		t.Errorf("expected the example line 'brag wrapped 2026 Q3' in help:\n%s", out)
	}
}

func TestWrappedCmd_StdoutStderrSeparation(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	withNowFunc(t, time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC))

	t.Run("success-to-stdout-only", func(t *testing.T) {
		seedWrappedEntry(t, dbPath, storage.Entry{
			Title: "jan", Project: "alpha", Type: "shipped",
		}, time.Date(2026, 1, 10, 10, 0, 0, 0, time.UTC))
		root, outBuf, errBuf := newWrappedTestRoot(t)
		root.SetArgs([]string{"--db", dbPath, "wrapped", "2026"})
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
		root, outBuf, _ := newWrappedTestRoot(t)
		root.SetArgs([]string{"--db", dbPath, "wrapped", "notayear"})
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
