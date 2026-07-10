package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/jysf/bragfile000/internal/storage"
	"github.com/spf13/cobra"

	_ "modernc.org/sqlite"
)

// newStoryTestRoot builds a fresh root with the story subcommand attached
// and isolates BRAGFILE_DB so the host env can't leak in.
func newStoryTestRoot(t *testing.T) (*cobra.Command, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	t.Setenv("BRAGFILE_DB", "")
	root := NewRootCmd("test")
	root.AddCommand(NewStoryCmd())
	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	root.SetOut(outBuf)
	root.SetErr(errBuf)
	return root, outBuf, errBuf
}

func runStoryCmd(t *testing.T, dbPath string, args ...string) (stdout, stderr string, runErr error) {
	t.Helper()
	root, outBuf, errBuf := newStoryTestRoot(t)
	full := append([]string{"--db", dbPath, "story"}, args...)
	root.SetArgs(full)
	runErr = root.Execute()
	return outBuf.String(), errBuf.String(), runErr
}

// withStoryNowFunc swaps the story clock seam for the duration of a test.
func withStoryNowFunc(t *testing.T, now time.Time) {
	t.Helper()
	prev := storyNowFunc
	storyNowFunc = func() time.Time { return now }
	t.Cleanup(func() { storyNowFunc = prev })
}

// seedStoryEntry mirrors impact's seedImpactEntry: add via the store, then
// rewrite created_at through a raw connection (Add always stamps now()).
func seedStoryEntry(t *testing.T, dbPath string, e storage.Entry, createdAt time.Time) {
	t.Helper()
	seedImpactEntry(t, dbPath, e, createdAt)
}

// seedStoryFixture seeds the six-entry story fixture (matching the
// internal/story fixture) into a temp DB.
func seedStoryFixture(t *testing.T, dbPath string) {
	t.Helper()
	entries := []struct {
		e  storage.Entry
		at time.Time
	}{
		{storage.Entry{Title: "alpha-early", Project: "alpha", Type: "shipped", Tags: "perf", Impact: "cut p95 login latency 40%"}, time.Date(2026, 2, 1, 10, 0, 0, 0, time.UTC)},
		{storage.Entry{Title: "alpha-messy", Project: "alpha", Type: "learned"}, time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC)},
		{storage.Entry{Title: "beta-one", Project: "beta", Type: "shipped", Impact: "onboarding time down to 1 day"}, time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC)},
		{storage.Entry{Title: "beta-two", Project: "beta", Type: "shipped", Impact: "removed the nightly cron entirely"}, time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC)},
		{storage.Entry{Title: "loose-note", Type: "fixed"}, time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)},
		{storage.Entry{Title: "perf-sweep", Project: "gamma", Type: "shipped", Tags: "perf", Impact: "shaved 200ms off cold start"}, time.Date(2026, 6, 15, 10, 0, 0, 0, time.UTC)},
	}
	for _, en := range entries {
		seedStoryEntry(t, dbPath, en.e, en.at)
	}
}

// Test 13 — TestStoryCmd_RequiresAudience.
func TestStoryCmd_RequiresAudience(t *testing.T) {
	t.Run("missing", func(t *testing.T) {
		dbPath := filepath.Join(t.TempDir(), "test.db")
		out, _, err := runStoryCmd(t, dbPath)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, ErrUser) {
			t.Errorf("expected errors.Is(err, ErrUser); got %v", err)
		}
		if !bytes.Contains([]byte(err.Error()), []byte("--audience")) {
			t.Errorf("error should mention --audience, got %q", err.Error())
		}
		if out != "" {
			t.Errorf("expected empty stdout, got %q", out)
		}
	})

	t.Run("unknown", func(t *testing.T) {
		dbPath := filepath.Join(t.TempDir(), "test.db")
		out, _, err := runStoryCmd(t, dbPath, "--audience", "nope")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, ErrUser) {
			t.Errorf("expected errors.Is(err, ErrUser); got %v", err)
		}
		if !bytes.Contains([]byte(err.Error()), []byte("nope")) {
			t.Errorf("error should name the audience, got %q", err.Error())
		}
		if out != "" {
			t.Errorf("expected empty stdout, got %q", out)
		}
	})
}

// Test 14 — TestStoryCmd_WindowResolutionAndSharedHelper (LOAD-BEARING).
func TestStoryCmd_WindowResolutionAndSharedHelper(t *testing.T) {
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)

	scopeFromJSON := func(t *testing.T, out string) string {
		t.Helper()
		var env struct {
			Scope string `json:"scope"`
		}
		if err := json.Unmarshal([]byte(out), &env); err != nil {
			t.Fatalf("json unmarshal: %v\n%s", err, out)
		}
		return env.Scope
	}

	t.Run("default-window-from-profile", func(t *testing.T) {
		dbPath := filepath.Join(t.TempDir(), "test.db")
		withStoryNowFunc(t, now)
		out, _, err := runStoryCmd(t, dbPath, "--audience", "me", "--format", "json")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got := scopeFromJSON(t, out); got != "year" {
			t.Errorf("me default scope: got %q, want year", got)
		}
	})

	t.Run("explicit-window-overrides", func(t *testing.T) {
		dbPath := filepath.Join(t.TempDir(), "test.db")
		withStoryNowFunc(t, now)
		out, _, err := runStoryCmd(t, dbPath, "--audience", "me", "--quarter", "--format", "json")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got := scopeFromJSON(t, out); got != "quarter" {
			t.Errorf("explicit scope: got %q, want quarter", got)
		}
	})

	t.Run("mutually-exclusive", func(t *testing.T) {
		dbPath := filepath.Join(t.TempDir(), "test.db")
		withStoryNowFunc(t, now)
		out, _, err := runStoryCmd(t, dbPath, "--audience", "me", "--quarter", "--month")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, ErrUser) {
			t.Errorf("expected errors.Is(err, ErrUser); got %v", err)
		}
		if !bytes.Contains([]byte(err.Error()), []byte("mutually exclusive")) {
			t.Errorf("expected mutual-exclusion error, got %q", err.Error())
		}
		if out != "" {
			t.Errorf("expected empty stdout, got %q", out)
		}
	})
}

// Test 15 — TestStoryCmd_MeVsExecDivergenceLive (LOAD-BEARING headline).
func TestStoryCmd_MeVsExecDivergenceLive(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	withStoryNowFunc(t, now)
	seedStoryFixture(t, dbPath)

	meOut, _, err := runStoryCmd(t, dbPath, "--audience", "me", "--year")
	if err != nil {
		t.Fatalf("me run: %v", err)
	}
	execOut, _, err := runStoryCmd(t, dbPath, "--audience", "exec", "--year")
	if err != nil {
		t.Fatalf("exec run: %v", err)
	}

	// me keeps impact-less beats + the (no project) thread.
	for _, needle := range []string{"alpha-messy", "loose-note", "### (no project)"} {
		if !bytes.Contains([]byte(meOut), []byte(needle)) {
			t.Errorf("me output should contain %q:\n%s", needle, meOut)
		}
	}
	// exec drops impact-less beats + folds the (no project) thread.
	for _, needle := range []string{"alpha-messy", "loose-note", "(no project)"} {
		if bytes.Contains([]byte(execOut), []byte(needle)) {
			t.Errorf("exec output must NOT contain %q:\n%s", needle, execOut)
		}
	}
	// exec leads with beta (impact-beat-count DESC).
	if !bytes.Contains([]byte(execOut), []byte("### beta")) {
		t.Fatalf("exec output missing beta thread:\n%s", execOut)
	}
	betaIdx := bytes.Index([]byte(execOut), []byte("### beta"))
	alphaIdx := bytes.Index([]byte(execOut), []byte("### alpha"))
	if alphaIdx >= 0 && betaIdx > alphaIdx {
		t.Errorf("exec first thread should be beta (before alpha):\n%s", execOut)
	}

	// Both carry their own directive; me's differs from exec's.
	if !bytes.Contains([]byte(meOut), []byte("messy middle")) {
		t.Errorf("me output should carry the me directive (messy middle)")
	}
	if !bytes.Contains([]byte(execOut), []byte("business impact")) {
		t.Errorf("exec output should carry the exec directive (business impact)")
	}
}

// Test 16 — TestStoryCmd_PrintDirectiveOnly.
func TestStoryCmd_PrintDirectiveOnly(t *testing.T) {
	t.Run("directive-only-no-db", func(t *testing.T) {
		// A non-existent DB path proves --print-directive does not open it.
		dbPath := filepath.Join(t.TempDir(), "does-not-exist.db")
		out, _, err := runStoryCmd(t, dbPath, "--audience", "exec", "--print-directive")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if bytes.Contains([]byte(out), []byte("# Bragfile Story")) {
			t.Errorf("--print-directive must not emit the bundle header:\n%s", out)
		}
		if bytes.Contains([]byte(out), []byte("## Threads")) {
			t.Errorf("--print-directive must not emit the threads section:\n%s", out)
		}
		if !bytes.Contains([]byte(out), []byte("business impact")) {
			t.Errorf("--print-directive should print the exec directive:\n%s", out)
		}
	})

	t.Run("unknown-audience", func(t *testing.T) {
		dbPath := filepath.Join(t.TempDir(), "test.db")
		out, _, err := runStoryCmd(t, dbPath, "--audience", "nope", "--print-directive")
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

// Test 17 — TestStoryCmd_FormatDefaultAndUnknown.
func TestStoryCmd_FormatDefaultAndUnknown(t *testing.T) {
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)

	t.Run("default-markdown", func(t *testing.T) {
		dbPath := filepath.Join(t.TempDir(), "test.db")
		withStoryNowFunc(t, now)
		out, _, err := runStoryCmd(t, dbPath, "--audience", "me", "--year")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !bytes.Contains([]byte(out), []byte("# Bragfile Story")) {
			t.Errorf("expected markdown header in default output:\n%s", out)
		}
	})

	t.Run("json", func(t *testing.T) {
		dbPath := filepath.Join(t.TempDir(), "test.db")
		withStoryNowFunc(t, now)
		out, _, err := runStoryCmd(t, dbPath, "--audience", "me", "--year", "--format", "json")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		var env map[string]json.RawMessage
		if err := json.Unmarshal([]byte(out), &env); err != nil {
			t.Fatalf("expected JSON, unmarshal failed: %v\n%s", err, out)
		}
		if _, ok := env["audience"]; !ok {
			t.Errorf("expected an 'audience' key in JSON output:\n%s", out)
		}
	})

	t.Run("unknown", func(t *testing.T) {
		dbPath := filepath.Join(t.TempDir(), "test.db")
		withStoryNowFunc(t, now)
		out, _, err := runStoryCmd(t, dbPath, "--audience", "me", "--year", "--format", "xml")
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

// Test 18 — TestStoryCmd_StdoutStderrSeparation.
func TestStoryCmd_StdoutStderrSeparation(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	withStoryNowFunc(t, now)

	t.Run("success-to-stdout-only", func(t *testing.T) {
		root, outBuf, errBuf := newStoryTestRoot(t)
		root.SetArgs([]string{"--db", dbPath, "story", "--audience", "me", "--year"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if outBuf.Len() == 0 {
			t.Errorf("expected bundle on stdout, got empty")
		}
		if errBuf.Len() != 0 {
			t.Errorf("expected empty stderr on success, got %q", errBuf.String())
		}
	})

	t.Run("usererror-keeps-stdout-clean", func(t *testing.T) {
		root, outBuf, _ := newStoryTestRoot(t)
		root.SetArgs([]string{"--db", dbPath, "story"}) // no audience → UserError
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

// storyScopeFromJSON unmarshals the scope from a story JSON envelope.
func storyScopeFromJSON(t *testing.T, out string) string {
	t.Helper()
	var env struct {
		Scope string `json:"scope"`
	}
	if err := json.Unmarshal([]byte(out), &env); err != nil {
		t.Fatalf("json unmarshal: %v\n%s", err, out)
	}
	return env.Scope
}

// TestStoryCmd_PreviousExplicitWindowBounded: story --quarter --previous at
// spec53Now → scope "quarter:previous", a prev-Q2 entry IN and a current-Q3
// entry OUT. Proves --previous shifts an explicit window flag AND applies the
// bounded upper edge on the story path (SPEC-053).
func TestStoryCmd_PreviousExplicitWindowBounded(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	withStoryNowFunc(t, time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC))

	seedStoryEntry(t, dbPath, storage.Entry{
		Title: "prev-q2-may", Project: "alpha", Type: "shipped", Impact: "in prev Q2",
	}, time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC))
	seedStoryEntry(t, dbPath, storage.Entry{
		Title: "cur-q3-jul", Project: "alpha", Type: "shipped", Impact: "current Q3",
	}, time.Date(2026, 7, 3, 10, 0, 0, 0, time.UTC))

	out, _, err := runStoryCmd(t, dbPath, "--audience", "exec", "--quarter", "--previous", "--format", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := storyScopeFromJSON(t, out); got != "quarter:previous" {
		t.Errorf("scope: got %q, want %q", got, "quarter:previous")
	}
	if !bytes.Contains([]byte(out), []byte("prev-q2-may")) {
		t.Errorf("prev-Q2 entry must be IN the bundle:\n%s", out)
	}
	if bytes.Contains([]byte(out), []byte("cur-q3-jul")) {
		t.Errorf("current-Q3 entry must be OUT (bounded upper edge):\n%s", out)
	}
}

// TestStoryCmd_PreviousShiftsProfileDefault: story --audience me --previous
// (no window flag) → the me profile default (year) shifted back one, bounded;
// scope "year:previous". A 2025 entry IN, a current-2026 entry OUT.
func TestStoryCmd_PreviousShiftsProfileDefault(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	withStoryNowFunc(t, time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC))

	seedStoryEntry(t, dbPath, storage.Entry{
		Title: "y2025-jun", Project: "alpha", Type: "shipped", Impact: "in prev year",
	}, time.Date(2025, 6, 1, 10, 0, 0, 0, time.UTC))
	seedStoryEntry(t, dbPath, storage.Entry{
		Title: "y2026-feb", Project: "alpha", Type: "shipped", Impact: "current year",
	}, time.Date(2026, 2, 1, 10, 0, 0, 0, time.UTC))

	out, _, err := runStoryCmd(t, dbPath, "--audience", "me", "--previous", "--format", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := storyScopeFromJSON(t, out); got != "year:previous" {
		t.Errorf("scope: got %q, want %q", got, "year:previous")
	}
	if !bytes.Contains([]byte(out), []byte("y2025-jun")) {
		t.Errorf("prev-year entry must be IN the bundle:\n%s", out)
	}
	if bytes.Contains([]byte(out), []byte("y2026-feb")) {
		t.Errorf("current-year entry must be OUT (bounded upper edge):\n%s", out)
	}
}

// TestStoryCmd_PreviousWithSinceIsUserError: --since + --previous → UserError,
// stdout empty (LD3, story path).
func TestStoryCmd_PreviousWithSinceIsUserError(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	withStoryNowFunc(t, time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC))
	out, _, err := runStoryCmd(t, dbPath, "--audience", "me", "--since", "2026-01-01", "--previous")
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

// TestStoryCmd_HelpShowsPrevious: --help contains --previous and a
// distinctive example line.
func TestStoryCmd_HelpShowsPrevious(t *testing.T) {
	root, outBuf, _ := newStoryTestRoot(t)
	root.SetArgs([]string{"story", "--help"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := outBuf.String()
	if !bytes.Contains([]byte(out), []byte("--previous")) {
		t.Errorf("expected --previous in help:\n%s", out)
	}
	if !bytes.Contains([]byte(out), []byte("brag story --audience me --previous")) {
		t.Errorf("expected example line 'brag story --audience me --previous' in help:\n%s", out)
	}
}
