package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jysf/bragfile000/internal/storage"
	"github.com/spf13/cobra"
)

// newSparkTestRoot builds a fresh root with the spark subcommand attached and
// isolates BRAGFILE_DB. Mirrors newCoverageTestRoot. Separate outBuf/errBuf so
// stdout/stderr cross-leakage is assertable (AGENTS.md §9).
func newSparkTestRoot(t *testing.T) (*cobra.Command, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	t.Setenv("BRAGFILE_DB", "")
	root := NewRootCmd("test")
	root.AddCommand(NewSparkCmd())
	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	root.SetOut(outBuf)
	root.SetErr(errBuf)
	return root, outBuf, errBuf
}

func runSparkCmd(t *testing.T, dbPath string, args ...string) (stdout, stderr string, runErr error) {
	t.Helper()
	root, outBuf, errBuf := newSparkTestRoot(t)
	full := append([]string{"--db", dbPath, "spark"}, args...)
	root.SetArgs(full)
	runErr = root.Execute()
	return outBuf.String(), errBuf.String(), runErr
}

// sparkGlyphs is the DEC-031 block-glyph ladder. containsGlyph reports whether
// s carries any of them — used to assert glyphs present (markdown default-on)
// and absent (--no-spark / NO_COLOR / JSON).
const sparkGlyphs = "▁▂▃▄▅▆▇█"

func containsGlyph(s string) bool {
	for _, r := range s {
		if strings.ContainsRune(sparkGlyphs, r) {
			return true
		}
	}
	return false
}

// TestSparkCmd_DefaultWindowIsMonth pins DEC-037 choice 1: no window flag →
// --month, and the scope echoes it.
func TestSparkCmd_DefaultWindowIsMonth(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	now := time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC)
	withNowFunc(t, now)
	seedImpactEntry(t, dbPath, storage.Entry{Title: "recent", Project: "alpha"}, now.Add(-3*24*time.Hour))

	out, errStr, err := runSparkCmd(t, dbPath)
	if err != nil {
		t.Fatalf("unexpected error: %v (stderr: %s)", err, errStr)
	}
	if !strings.Contains(out, "# Bragfile Spark") {
		t.Errorf("expected markdown header:\n%s", out)
	}
	if !strings.Contains(out, "Scope: month") {
		t.Errorf("expected default scope month:\n%s", out)
	}
}

// TestSparkCmd_WindowsMutuallyExclusive pins DEC-037 choice 1: --week/--month/
// --quarter are mutually exclusive; two → UserError with empty stdout.
func TestSparkCmd_WindowsMutuallyExclusive(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	withNowFunc(t, time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC))

	root, outBuf, _ := newSparkTestRoot(t)
	root.SetArgs([]string{"--db", dbPath, "spark", "--week", "--quarter"})
	err := root.Execute()
	if err == nil || !errors.Is(err, ErrUser) {
		t.Fatalf("expected UserError, got %v", err)
	}
	if outBuf.Len() != 0 {
		t.Errorf("expected empty stdout, got %q", outBuf.String())
	}
}

// TestSparkCmd_TotalAndByProjectRows pins DEC-037 choice 3: a Total row plus
// by-project rows, with (no project) handled via aggregate.NoProjectKey.
func TestSparkCmd_TotalAndByProjectRows(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	now := time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC)
	withNowFunc(t, now)
	seedImpactEntry(t, dbPath, storage.Entry{Title: "a1", Project: "alpha"}, now.Add(-2*24*time.Hour))
	seedImpactEntry(t, dbPath, storage.Entry{Title: "a2", Project: "alpha"}, now.Add(-5*24*time.Hour))
	seedImpactEntry(t, dbPath, storage.Entry{Title: "b1", Project: "beta"}, now.Add(-1*24*time.Hour))
	seedImpactEntry(t, dbPath, storage.Entry{Title: "n1"}, now.Add(-3*24*time.Hour)) // no project

	out, errStr, err := runSparkCmd(t, dbPath)
	if err != nil {
		t.Fatalf("unexpected error: %v (stderr: %s)", err, errStr)
	}
	for _, needle := range []string{"Total (", "alpha (", "beta (", "(no project) ("} {
		if !strings.Contains(out, needle) {
			t.Errorf("expected row %q in output:\n%s", needle, out)
		}
	}
}

// TestSparkCmd_TopEightCap pins DEC-037 choice 3: at most the top-8 projects by
// in-window entry volume are shown. Eight projects tie at count 2 (all kept via
// alpha-ASC); two projects at count 1 are dropped.
func TestSparkCmd_TopEightCap(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	now := time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC)
	withNowFunc(t, now)
	for i := 1; i <= 8; i++ {
		p := fmt.Sprintf("keep%d", i)
		seedImpactEntry(t, dbPath, storage.Entry{Title: p + "-1", Project: p}, now.Add(-2*24*time.Hour))
		seedImpactEntry(t, dbPath, storage.Entry{Title: p + "-2", Project: p}, now.Add(-4*24*time.Hour))
	}
	seedImpactEntry(t, dbPath, storage.Entry{Title: "d1", Project: "drop1"}, now.Add(-1*24*time.Hour))
	seedImpactEntry(t, dbPath, storage.Entry{Title: "d2", Project: "drop2"}, now.Add(-1*24*time.Hour))

	out, errStr, err := runSparkCmd(t, dbPath)
	if err != nil {
		t.Fatalf("unexpected error: %v (stderr: %s)", err, errStr)
	}
	if !strings.Contains(out, "keep1 (") {
		t.Errorf("expected a top-8 project row (keep1):\n%s", out)
	}
	for _, dropped := range []string{"drop1 (", "drop2 ("} {
		if strings.Contains(out, dropped) {
			t.Errorf("expected %q excluded by the top-8 cap:\n%s", dropped, out)
		}
	}
}

// TestSparkCmd_ProjectSelector pins DEC-037 choice 3: --project narrows the
// by-project rows to that one project (Total still spans the whole corpus);
// other projects' rows are absent.
func TestSparkCmd_ProjectSelector(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	now := time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC)
	withNowFunc(t, now)
	seedImpactEntry(t, dbPath, storage.Entry{Title: "a1", Project: "alpha"}, now.Add(-2*24*time.Hour))
	seedImpactEntry(t, dbPath, storage.Entry{Title: "b1", Project: "beta"}, now.Add(-2*24*time.Hour))

	out, errStr, err := runSparkCmd(t, dbPath, "--project", "alpha")
	if err != nil {
		t.Fatalf("unexpected error: %v (stderr: %s)", err, errStr)
	}
	if !strings.Contains(out, "Total (") {
		t.Errorf("expected Total row to remain:\n%s", out)
	}
	if !strings.Contains(out, "alpha (") {
		t.Errorf("expected the selected project row (alpha):\n%s", out)
	}
	if strings.Contains(out, "beta (") {
		t.Errorf("expected other project (beta) row absent under --project alpha:\n%s", out)
	}
}

// TestSparkCmd_SparkDefaultOnAndSuppressed pins DEC-037 choice 6: markdown
// renders glyph rows by default; --no-spark and a present NO_COLOR both
// suppress the glyphs (falling back to raw per-bucket counts, so no glyph runes
// remain).
func TestSparkCmd_SparkDefaultOnAndSuppressed(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	now := time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC)
	withNowFunc(t, now)
	seedImpactEntry(t, dbPath, storage.Entry{Title: "a1", Project: "alpha"}, now.Add(-2*24*time.Hour))
	seedImpactEntry(t, dbPath, storage.Entry{Title: "a2", Project: "alpha"}, now.Add(-20*24*time.Hour))

	t.Run("default-on", func(t *testing.T) {
		withLookupSparkEnv(t, "", false) // NO_COLOR unset
		out, _, err := runSparkCmd(t, dbPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !containsGlyph(out) {
			t.Errorf("default-on must render glyphs:\n%s", out)
		}
	})

	t.Run("no-spark-flag", func(t *testing.T) {
		withLookupSparkEnv(t, "", false)
		out, _, err := runSparkCmd(t, dbPath, "--no-spark")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if containsGlyph(out) {
			t.Errorf("--no-spark must suppress glyphs:\n%s", out)
		}
	})

	t.Run("no-color-env", func(t *testing.T) {
		withLookupSparkEnv(t, "", true) // NO_COLOR present (empty value)
		out, _, err := runSparkCmd(t, dbPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if containsGlyph(out) {
			t.Errorf("NO_COLOR present must suppress glyphs:\n%s", out)
		}
	})
}

// TestSparkCmd_JSONRawCountsNoGlyphs pins DEC-037 choice 6 / DEC-031 choice f:
// --format json is valid JSON carrying raw per-bucket counts (total.series +
// by_project) and contains NO glyph runes.
func TestSparkCmd_JSONRawCountsNoGlyphs(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	now := time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC)
	withNowFunc(t, now)
	seedImpactEntry(t, dbPath, storage.Entry{Title: "a1", Project: "alpha"}, now.Add(-2*24*time.Hour))

	out, errStr, err := runSparkCmd(t, dbPath, "--format", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v (stderr: %s)", err, errStr)
	}
	if containsGlyph(out) {
		t.Errorf("JSON must not contain glyph runes:\n%s", out)
	}
	var env map[string]any
	if err := json.Unmarshal([]byte(out), &env); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, out)
	}
	if _, ok := env["total"]; !ok {
		t.Errorf("JSON missing total key:\n%s", out)
	}
	if _, ok := env["by_project"]; !ok {
		t.Errorf("JSON missing by_project key:\n%s", out)
	}
	if _, ok := env["scope"]; !ok {
		t.Errorf("JSON missing scope key (DEC-014 envelope):\n%s", out)
	}
}

// TestSparkCmd_UnknownFormat pins the --format guard (DEC-007 pattern).
func TestSparkCmd_UnknownFormat(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	withNowFunc(t, time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC))

	root, outBuf, _ := newSparkTestRoot(t)
	root.SetArgs([]string{"--db", dbPath, "spark", "--format", "yaml"})
	err := root.Execute()
	if err == nil || !errors.Is(err, ErrUser) {
		t.Fatalf("expected UserError, got %v", err)
	}
	if outBuf.Len() != 0 {
		t.Errorf("expected empty stdout, got %q", outBuf.String())
	}
}

// TestSparkCmd_EmptyCorpusHeaderOnly pins DEC-014 part 4: an empty window emits
// the header/provenance block only (through "Entries: 0"); no ## Pulse body.
func TestSparkCmd_EmptyCorpusHeaderOnly(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	withNowFunc(t, time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC))

	out, errStr, err := runSparkCmd(t, dbPath)
	if err != nil {
		t.Fatalf("unexpected error: %v (stderr: %s)", err, errStr)
	}
	if !strings.Contains(out, "Entries: 0") {
		t.Errorf("expected 'Entries: 0' on empty corpus:\n%s", out)
	}
	if strings.Contains(out, "## Pulse") {
		t.Errorf("expected no ## Pulse body on empty corpus:\n%s", out)
	}
	if containsGlyph(out) {
		t.Errorf("expected no glyphs on empty corpus:\n%s", out)
	}
}

// TestSparkCmd_StdoutStderrSeparation pins AGENTS.md §9: success → stdout only;
// UserError → clean stdout.
func TestSparkCmd_StdoutStderrSeparation(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	now := time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC)
	withNowFunc(t, now)
	seedImpactEntry(t, dbPath, storage.Entry{Title: "a1", Project: "alpha"}, now.Add(-2*24*time.Hour))

	t.Run("success-to-stdout-only", func(t *testing.T) {
		root, outBuf, errBuf := newSparkTestRoot(t)
		root.SetArgs([]string{"--db", dbPath, "spark"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if outBuf.Len() == 0 {
			t.Errorf("expected pulse on stdout, got empty")
		}
		if errBuf.Len() != 0 {
			t.Errorf("expected empty stderr on success, got %q", errBuf.String())
		}
	})

	t.Run("usererror-keeps-stdout-clean", func(t *testing.T) {
		root, outBuf, _ := newSparkTestRoot(t)
		root.SetArgs([]string{"--db", dbPath, "spark", "--week", "--month"})
		err := root.Execute()
		if err == nil || !errors.Is(err, ErrUser) {
			t.Fatalf("expected UserError, got %v", err)
		}
		if outBuf.Len() != 0 {
			t.Errorf("expected empty stdout on error, got %q", outBuf.String())
		}
	})
}

// TestSparkCmd_HelpShowsExamples pins the literal-artifact Long: an Examples:
// block naming `brag spark`.
func TestSparkCmd_HelpShowsExamples(t *testing.T) {
	root, outBuf, _ := newSparkTestRoot(t)
	root.SetArgs([]string{"spark", "--help"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	help := outBuf.String()
	if !strings.Contains(help, "Examples:") {
		t.Errorf("help missing 'Examples:' label:\n%s", help)
	}
	if !strings.Contains(help, "brag spark") {
		t.Errorf("help missing 'brag spark' example:\n%s", help)
	}
}
