package cli

import (
	"bytes"
	"errors"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jysf/bragfile000/internal/storage"
	"github.com/spf13/cobra"
)

// --- buildFTS5Query tests (pure function, no DB, no cobra) ---------

func TestBuildFTS5Query_SingleWord(t *testing.T) {
	got, err := buildFTS5Query("latency")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := `"latency"`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestBuildFTS5Query_MultiWordAnd(t *testing.T) {
	got, err := buildFTS5Query("cut latency")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := `"cut" "latency"`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestBuildFTS5Query_HyphenatedLiteral(t *testing.T) {
	got, err := buildFTS5Query("auth-refactor")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := `"auth-refactor"`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestBuildFTS5Query_EmptyIsError(t *testing.T) {
	got, err := buildFTS5Query("")
	if err == nil {
		t.Fatalf("expected error, got nil (output=%q)", got)
	}
	if got != "" {
		t.Errorf("expected empty output on error, got %q", got)
	}
}

func TestBuildFTS5Query_WhitespaceOnlyIsError(t *testing.T) {
	_, err := buildFTS5Query("   \t  ")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestBuildFTS5Query_QuoteInQueryIsError(t *testing.T) {
	_, err := buildFTS5Query(`with "quote"`)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- CLI helpers ----------------------------------------------------

// newRootWithSearch builds a fresh root + search subcommand, separate
// stdout/stderr buffers, and a t.TempDir()-backed DB path. Callers
// seed entries with seedSearchEntries before invoking the command.
func newRootWithSearch(t *testing.T) (root *cobra.Command, outBuf, errBuf *bytes.Buffer, dbPath string) {
	t.Helper()
	t.Setenv("BRAGFILE_DB", "")
	dbPath = filepath.Join(t.TempDir(), "search.db")
	root = NewRootCmd("test")
	root.AddCommand(NewSearchCmd())
	outBuf = &bytes.Buffer{}
	errBuf = &bytes.Buffer{}
	root.SetOut(outBuf)
	root.SetErr(errBuf)
	return root, outBuf, errBuf, dbPath
}

// seedSearchEntries opens a Store at dbPath, Adds each Entry in order,
// and closes. Returns the hydrated inserted entries in insertion order.
func seedSearchEntries(t *testing.T, dbPath string, es ...storage.Entry) []storage.Entry {
	t.Helper()
	s, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("storage.Open: %v", err)
	}
	defer s.Close()
	out := make([]storage.Entry, 0, len(es))
	for _, e := range es {
		ins, err := s.Add(e)
		if err != nil {
			t.Fatalf("Add(%q): %v", e.Title, err)
		}
		out = append(out, ins)
	}
	return out
}

// runSearchCmd is the common "invoke brag search <args>" wrapper.
func runSearchCmd(t *testing.T, dbPath string, searchArgs ...string) (stdout, stderr string, runErr error) {
	t.Helper()
	root, outBuf, errBuf, _ := newRootWithSearchAt(t, dbPath)
	full := append([]string{"--db", dbPath, "search"}, searchArgs...)
	root.SetArgs(full)
	runErr = root.Execute()
	return outBuf.String(), errBuf.String(), runErr
}

// newRootWithSearchAt is a variant that uses an already-chosen dbPath
// so the caller can seed it first.
func newRootWithSearchAt(t *testing.T, dbPath string) (*cobra.Command, *bytes.Buffer, *bytes.Buffer, string) {
	t.Helper()
	t.Setenv("BRAGFILE_DB", "")
	root := NewRootCmd("test")
	root.AddCommand(NewSearchCmd())
	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	root.SetOut(outBuf)
	root.SetErr(errBuf)
	return root, outBuf, errBuf, dbPath
}

// --- CLI tests ------------------------------------------------------

func TestSearchCmd_HappyPath(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "search.db")
	seeded := seedSearchEntries(t, dbPath,
		storage.Entry{Title: "alpha distinctive"},                      // id 1
		storage.Entry{Title: "beta", Description: "alpha distinctive"}, // id 2
		storage.Entry{Title: "gamma", Description: "unrelated stuff"},  // id 3
	)
	if len(seeded) != 3 {
		t.Fatalf("seed count = %d, want 3", len(seeded))
	}

	out, errOut, err := runSearchCmd(t, dbPath, "alpha distinctive")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errOut != "" {
		t.Fatalf("expected empty stderr, got %q", errOut)
	}

	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("want 2 lines, got %d: %q", len(lines), out)
	}
	for i, line := range lines {
		fields := strings.Split(line, "\t")
		if len(fields) != 3 {
			t.Fatalf("line %d: want 3 tab-separated fields, got %d: %q", i, len(fields), line)
		}
	}
	// Both row 1 and row 2 match; row 3 does not. We don't assert on
	// ordering here (rank-ordering is covered separately) — just that
	// the included/excluded set is correct.
	titles := []string{
		strings.Split(lines[0], "\t")[2],
		strings.Split(lines[1], "\t")[2],
	}
	haveAlpha, haveBeta := false, false
	for _, t0 := range titles {
		switch t0 {
		case "alpha distinctive":
			haveAlpha = true
		case "beta":
			haveBeta = true
		}
	}
	if !haveAlpha || !haveBeta {
		t.Errorf("expected titles to include both %q and %q, got %v",
			"alpha distinctive", "beta", titles)
	}
	for _, line := range lines {
		if strings.Contains(line, "gamma") {
			t.Errorf("unexpected match for %q in %q", "gamma", line)
		}
	}
}

func TestSearchCmd_HyphenatedQuery(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "search.db")
	seedSearchEntries(t, dbPath,
		storage.Entry{
			Title:       "landing-notes",
			Description: "the auth refactor landed cleanly",
		},
	)

	out, errOut, err := runSearchCmd(t, dbPath, "auth-refactor")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errOut != "" {
		t.Fatalf("expected empty stderr, got %q", errOut)
	}
	if !strings.Contains(out, "landing-notes") {
		t.Errorf("expected stdout to contain %q, got %q", "landing-notes", out)
	}
}

func TestSearchCmd_MultiWordAndSemantics(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "search.db")
	seedSearchEntries(t, dbPath,
		storage.Entry{Title: "cut latency"},
		storage.Entry{Title: "only cut"},
		storage.Entry{Title: "only latency"},
	)

	out, errOut, err := runSearchCmd(t, dbPath, "cut latency")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errOut != "" {
		t.Fatalf("expected empty stderr, got %q", errOut)
	}
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 1 {
		t.Fatalf("want 1 line, got %d: %q", len(lines), out)
	}
	if title := strings.Split(lines[0], "\t")[2]; title != "cut latency" {
		t.Errorf("title = %q, want %q", title, "cut latency")
	}
}

func TestSearchCmd_ZeroResultsExitsZero(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "search.db")
	seedSearchEntries(t, dbPath, storage.Entry{Title: "unrelated"})

	out, errOut, err := runSearchCmd(t, dbPath, "xyznomatchxyzxyz")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "" {
		t.Errorf("expected empty stdout, got %q", out)
	}
	if errOut != "" {
		t.Errorf("expected empty stderr, got %q", errOut)
	}
}

func TestSearchCmd_EmptyQueryIsUserError(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "search.db")
	out, _, err := runSearchCmd(t, dbPath, "")
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

func TestSearchCmd_WhitespaceOnlyQueryIsUserError(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "search.db")
	_, _, err := runSearchCmd(t, dbPath, "   ")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrUser) {
		t.Errorf("expected errors.Is(err, ErrUser); got %v", err)
	}
}

func TestSearchCmd_QuoteInQueryIsUserError(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "search.db")
	_, _, err := runSearchCmd(t, dbPath, `with "quote"`)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrUser) {
		t.Errorf("expected errors.Is(err, ErrUser); got %v", err)
	}
}

func TestSearchCmd_NoArgsIsUserError(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "search.db")
	root, outBuf, _, _ := newRootWithSearchAt(t, dbPath)
	root.SetArgs([]string{"--db", dbPath, "search"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrUser) {
		t.Errorf("expected errors.Is(err, ErrUser); got %v", err)
	}
	if outBuf.Len() != 0 {
		t.Errorf("expected empty stdout, got %q", outBuf.String())
	}
}

func TestSearchCmd_TooManyArgsIsUserError(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "search.db")
	_, _, err := runSearchCmd(t, dbPath, "a", "b")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrUser) {
		t.Errorf("expected errors.Is(err, ErrUser); got %v", err)
	}
}

func TestSearchCmd_TabSeparatedOutput(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "search.db")
	seeded := seedSearchEntries(t, dbPath,
		storage.Entry{Title: "solitary-pinecone"},
	)
	inserted := seeded[0]

	out, errOut, err := runSearchCmd(t, dbPath, "solitary-pinecone")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errOut != "" {
		t.Fatalf("expected empty stderr, got %q", errOut)
	}

	line := strings.TrimRight(out, "\n")
	if strings.Contains(line, "\n") {
		t.Fatalf("expected single output line, got %q", out)
	}
	if got := strings.Count(line, "\t"); got != 2 {
		t.Fatalf("expected exactly 2 tab chars, got %d: %q", got, line)
	}
	fields := strings.Split(line, "\t")
	if len(fields) != 3 {
		t.Fatalf("expected 3 fields, got %d: %q", len(fields), line)
	}
	if fields[0] != strconv.FormatInt(inserted.ID, 10) {
		t.Errorf("id field = %q, want %d", fields[0], inserted.ID)
	}
	if _, err := time.Parse(time.RFC3339, fields[1]); err != nil {
		t.Errorf("created_at field %q is not RFC3339: %v", fields[1], err)
	}
	if fields[2] != "solitary-pinecone" {
		t.Errorf("title field = %q, want %q", fields[2], "solitary-pinecone")
	}
}

func TestSearchCmd_LimitRespected(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "search.db")
	seedSearchEntries(t, dbPath,
		storage.Entry{Title: "matchword one"},
		storage.Entry{Title: "matchword two"},
		storage.Entry{Title: "matchword three"},
		storage.Entry{Title: "matchword four"},
		storage.Entry{Title: "matchword five"},
	)

	out, errOut, err := runSearchCmd(t, dbPath, "matchword", "--limit", "3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errOut != "" {
		t.Fatalf("expected empty stderr, got %q", errOut)
	}
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("want 3 lines, got %d: %q", len(lines), out)
	}
}

func TestSearchCmd_InvalidLimitIsUserError(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "search.db")
	_, _, err := runSearchCmd(t, dbPath, "x", "--limit", "-5")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrUser) {
		t.Errorf("expected errors.Is(err, ErrUser); got %v", err)
	}
}

func TestSearchCmd_HelpShape(t *testing.T) {
	root, outBuf, errBuf, _ := newRootWithSearch(t)
	root.SetArgs([]string{"search", "--help"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errBuf.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", errBuf.String())
	}
	out := outBuf.String()
	if !strings.Contains(out, "Examples:") {
		t.Errorf("expected help to contain %q, got:\n%s", "Examples:", out)
	}
}
