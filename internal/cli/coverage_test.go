package cli

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jysf/bragfile000/internal/storage"
	"github.com/spf13/cobra"

	_ "modernc.org/sqlite"
)

// newCoverageTestRoot builds a fresh root with the coverage subcommand
// attached and isolates BRAGFILE_DB. Mirrors newImpactTestRoot.
func newCoverageTestRoot(t *testing.T) (*cobra.Command, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	t.Setenv("BRAGFILE_DB", "")
	root := NewRootCmd("test")
	root.AddCommand(NewCoverageCmd())
	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	root.SetOut(outBuf)
	root.SetErr(errBuf)
	return root, outBuf, errBuf
}

func runCoverageCmd(t *testing.T, dbPath string, args ...string) (stdout, stderr string, runErr error) {
	t.Helper()
	root, outBuf, errBuf := newCoverageTestRoot(t)
	full := append([]string{"--db", dbPath, "coverage"}, args...)
	root.SetArgs(full)
	runErr = root.Execute()
	return outBuf.String(), errBuf.String(), runErr
}

// seedCoverageEntry adds an entry then rewrites its created_at (the store
// stamps time.Now()). Raw SQL confined to the test file. Mirrors
// seedImpactEntry.
func seedCoverageEntry(t *testing.T, dbPath string, e storage.Entry, createdAt time.Time) {
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

// seedCoverageMix seeds a small mixed agent/human corpus in the current year
// (relative to the injected now used by the caller).
func seedCoverageMix(t *testing.T, dbPath string, now time.Time) {
	t.Helper()
	seedCoverageEntry(t, dbPath, storage.Entry{Title: "human work", Tags: "perf"},
		time.Date(now.Year(), 2, 1, 10, 0, 0, 0, time.UTC))
	seedCoverageEntry(t, dbPath, storage.Entry{Title: "agent work", Tags: "agent:claude-code"},
		time.Date(now.Year(), 3, 1, 10, 0, 0, 0, time.UTC))
}

// Test 10 — TestCoverageCmd_RequiresExactlyOneWindow.
func TestCoverageCmd_RequiresExactlyOneWindow(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	withNowFunc(t, time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC))

	t.Run("no-window", func(t *testing.T) {
		root, outBuf, _ := newCoverageTestRoot(t)
		root.SetArgs([]string{"--db", dbPath, "coverage"})
		err := root.Execute()
		if err == nil || !errors.Is(err, ErrUser) {
			t.Fatalf("expected UserError, got %v", err)
		}
		if outBuf.Len() != 0 {
			t.Errorf("expected empty stdout, got %q", outBuf.String())
		}
	})

	t.Run("two-windows", func(t *testing.T) {
		root, outBuf, _ := newCoverageTestRoot(t)
		root.SetArgs([]string{"--db", dbPath, "coverage", "--year", "--month"})
		err := root.Execute()
		if err == nil || !errors.Is(err, ErrUser) {
			t.Fatalf("expected UserError naming the conflict, got %v", err)
		}
		if outBuf.Len() != 0 {
			t.Errorf("expected empty stdout, got %q", outBuf.String())
		}
	})
}

// Test 11 — TestCoverageCmd_CalendarWindowAndScope.
func TestCoverageCmd_CalendarWindowAndScope(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	withNowFunc(t, now)
	seedCoverageMix(t, dbPath, now)

	cases := []struct {
		name      string
		args      []string
		wantScope string
	}{
		{"year", []string{"--year"}, "Scope: year"},
		{"since", []string{"--since", "2026-01-01"}, "Scope: since:2026-01-01"},
		{"quarter-previous", []string{"--quarter", "--previous"}, "Scope: quarter:previous"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out, errStr, err := runCoverageCmd(t, dbPath, c.args...)
			if err != nil {
				t.Fatalf("unexpected error: %v (stderr: %s)", err, errStr)
			}
			if !strings.Contains(out, c.wantScope) {
				t.Errorf("expected %q in output:\n%s", c.wantScope, out)
			}
		})
	}

	t.Run("previous-since-is-usererror", func(t *testing.T) {
		root, outBuf, _ := newCoverageTestRoot(t)
		root.SetArgs([]string{"--db", dbPath, "coverage", "--since", "2026-01-01", "--previous"})
		err := root.Execute()
		if err == nil || !errors.Is(err, ErrUser) {
			t.Fatalf("expected UserError, got %v", err)
		}
		if outBuf.Len() != 0 {
			t.Errorf("expected empty stdout, got %q", outBuf.String())
		}
	})
}

// Test 12 — TestCoverageCmd_FormatDefaultAndUnknown.
func TestCoverageCmd_FormatDefaultAndUnknown(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	withNowFunc(t, now)
	seedCoverageMix(t, dbPath, now)

	t.Run("default-markdown", func(t *testing.T) {
		out, errStr, err := runCoverageCmd(t, dbPath, "--year")
		if err != nil {
			t.Fatalf("unexpected error: %v (stderr: %s)", err, errStr)
		}
		if !strings.Contains(out, "# Bragfile Coverage") {
			t.Errorf("expected markdown header:\n%s", out)
		}
	})

	t.Run("json", func(t *testing.T) {
		out, errStr, err := runCoverageCmd(t, dbPath, "--year", "--format", "json")
		if err != nil {
			t.Fatalf("unexpected error: %v (stderr: %s)", err, errStr)
		}
		var env map[string]any
		if err := json.Unmarshal([]byte(out), &env); err != nil {
			t.Fatalf("output is not valid JSON: %v\n%s", err, out)
		}
		if _, ok := env["agent_share"]; !ok {
			t.Errorf("JSON missing agent_share key:\n%s", out)
		}
	})

	t.Run("unknown-format", func(t *testing.T) {
		root, outBuf, _ := newCoverageTestRoot(t)
		root.SetArgs([]string{"--db", dbPath, "coverage", "--year", "--format", "yaml"})
		err := root.Execute()
		if err == nil || !errors.Is(err, ErrUser) {
			t.Fatalf("expected UserError, got %v", err)
		}
		if outBuf.Len() != 0 {
			t.Errorf("expected empty stdout, got %q", outBuf.String())
		}
	})
}

// Test 13 — TestCoverageCmd_NoSparkAndNoColorSuppress.
func TestCoverageCmd_NoSparkAndNoColorSuppress(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	withNowFunc(t, now)
	seedCoverageMix(t, dbPath, now)

	t.Run("default-on", func(t *testing.T) {
		withLookupSparkEnv(t, "", false) // NO_COLOR unset
		out, errStr, err := runCoverageCmd(t, dbPath, "--year")
		if err != nil {
			t.Fatalf("unexpected error: %v (stderr: %s)", err, errStr)
		}
		if !strings.Contains(out, "Agent share:") {
			t.Errorf("default-on must render 'Agent share:' line:\n%s", out)
		}
	})

	t.Run("no-spark-flag", func(t *testing.T) {
		withLookupSparkEnv(t, "", false)
		out, errStr, err := runCoverageCmd(t, dbPath, "--year", "--no-spark")
		if err != nil {
			t.Fatalf("unexpected error: %v (stderr: %s)", err, errStr)
		}
		if strings.Contains(out, "Agent share:") {
			t.Errorf("--no-spark must remove the 'Agent share:' line:\n%s", out)
		}
	})

	t.Run("no-color-env", func(t *testing.T) {
		withLookupSparkEnv(t, "", true) // NO_COLOR present (empty value)
		out, errStr, err := runCoverageCmd(t, dbPath, "--year")
		if err != nil {
			t.Fatalf("unexpected error: %v (stderr: %s)", err, errStr)
		}
		if strings.Contains(out, "Agent share:") {
			t.Errorf("NO_COLOR present must suppress 'Agent share:' line:\n%s", out)
		}
	})
}

// Test 14 — TestCoverageCmd_StdoutStderrSeparation.
func TestCoverageCmd_StdoutStderrSeparation(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	withNowFunc(t, now)
	seedCoverageMix(t, dbPath, now)

	t.Run("success-to-stdout-only", func(t *testing.T) {
		root, outBuf, errBuf := newCoverageTestRoot(t)
		root.SetArgs([]string{"--db", dbPath, "coverage", "--year"})
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
		root, outBuf, _ := newCoverageTestRoot(t)
		root.SetArgs([]string{"--db", dbPath, "coverage", "--year", "--format", "yaml"})
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

// Test 15 — TestCoverageCmd_HelpShowsExamples.
func TestCoverageCmd_HelpShowsExamples(t *testing.T) {
	root, outBuf, _ := newCoverageTestRoot(t)
	root.SetArgs([]string{"coverage", "--help"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	help := outBuf.String()
	if !strings.Contains(help, "Examples:") {
		t.Errorf("help missing 'Examples:' label:\n%s", help)
	}
	if !strings.Contains(help, "brag coverage --year") {
		t.Errorf("help missing example line 'brag coverage --year':\n%s", help)
	}
}
