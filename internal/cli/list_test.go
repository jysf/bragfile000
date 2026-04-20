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

// newListTestRoot builds a fresh root command with the list subcommand
// attached, wires separate stdout/stderr buffers, and clears
// BRAGFILE_DB so host env doesn't leak in. The caller drives args via
// cmd.SetArgs.
func newListTestRoot(t *testing.T, dbPath string) (*cobra.Command, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	t.Setenv("BRAGFILE_DB", "")
	_ = dbPath
	root := NewRootCmd("test")
	root.AddCommand(NewListCmd())
	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	root.SetOut(outBuf)
	root.SetErr(errBuf)
	return root, outBuf, errBuf
}

func TestListCmd_EmptyPrintsNothing(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")

	// Open and immediately close so the DB file + schema exist but no rows are present.
	s, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("storage.Open: %v", err)
	}
	if err := s.Close(); err != nil {
		t.Fatalf("storage.Close: %v", err)
	}

	root, outBuf, errBuf := newListTestRoot(t, dbPath)
	root.SetArgs([]string{"--db", dbPath, "list"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if outBuf.Len() != 0 {
		t.Fatalf("expected empty stdout, got %q", outBuf.String())
	}
	if errBuf.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", errBuf.String())
	}
}

func TestListCmd_PrintsReverseChronological(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")

	s, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("storage.Open: %v", err)
	}
	for _, title := range []string{"first", "second", "third"} {
		if _, err := s.Add(storage.Entry{Title: title}); err != nil {
			t.Fatalf("Add(%q): %v", title, err)
		}
	}
	if err := s.Close(); err != nil {
		t.Fatalf("storage.Close: %v", err)
	}

	root, outBuf, errBuf := newListTestRoot(t, dbPath)
	root.SetArgs([]string{"--db", dbPath, "list"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errBuf.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", errBuf.String())
	}

	lines := strings.Split(strings.TrimRight(outBuf.String(), "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d: %q", len(lines), outBuf.String())
	}
	want := []string{"third", "second", "first"}
	for i, line := range lines {
		fields := strings.Split(line, "\t")
		if len(fields) != 3 {
			t.Fatalf("line %d: want 3 tab-separated fields, got %d: %q", i, len(fields), line)
		}
		if fields[2] != want[i] {
			t.Errorf("line %d: want title %q, got %q", i, want[i], fields[2])
		}
	}
}

func TestListCmd_TabSeparatedFormat(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")

	s, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("storage.Open: %v", err)
	}
	inserted, err := s.Add(storage.Entry{Title: "solo"})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	if err := s.Close(); err != nil {
		t.Fatalf("storage.Close: %v", err)
	}

	root, outBuf, errBuf := newListTestRoot(t, dbPath)
	root.SetArgs([]string{"--db", dbPath, "list"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errBuf.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", errBuf.String())
	}

	line := strings.TrimRight(outBuf.String(), "\n")
	if strings.Contains(line, "\n") {
		t.Fatalf("expected a single line, got %q", outBuf.String())
	}
	if got := strings.Count(line, "\t"); got != 2 {
		t.Fatalf("expected exactly 2 tab characters, got %d: %q", got, line)
	}
	fields := strings.Split(line, "\t")
	if _, err := strconv.ParseInt(fields[0], 10, 64); err != nil {
		t.Errorf("field 0 (id): %q: %v", fields[0], err)
	}
	ts, err := time.Parse(time.RFC3339, fields[1])
	if err != nil {
		t.Errorf("field 1 (created_at): %q: %v", fields[1], err)
	} else if loc := ts.Location().String(); loc != "UTC" {
		t.Errorf("field 1 (created_at): expected UTC location, got %q", loc)
	}
	if fields[2] != "solo" {
		t.Errorf("field 2 (title): want %q, got %q", "solo", fields[2])
	}
	if fields[0] != strconv.FormatInt(inserted.ID, 10) {
		t.Errorf("field 0 (id): want %d, got %q", inserted.ID, fields[0])
	}
}

func TestListCmd_TieBreakIsIDDescending(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")

	s, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("storage.Open: %v", err)
	}
	// Two rapid inserts should share a created_at second; the tie
	// break in Store.List (id DESC) must carry through.
	a, err := s.Add(storage.Entry{Title: "earlier-id"})
	if err != nil {
		t.Fatalf("Add a: %v", err)
	}
	b, err := s.Add(storage.Entry{Title: "later-id"})
	if err != nil {
		t.Fatalf("Add b: %v", err)
	}
	if err := s.Close(); err != nil {
		t.Fatalf("storage.Close: %v", err)
	}
	if b.ID <= a.ID {
		t.Fatalf("expected b.ID > a.ID; got a=%d b=%d", a.ID, b.ID)
	}

	root, outBuf, errBuf := newListTestRoot(t, dbPath)
	root.SetArgs([]string{"--db", dbPath, "list"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errBuf.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", errBuf.String())
	}

	lines := strings.Split(strings.TrimRight(outBuf.String(), "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %q", len(lines), outBuf.String())
	}
	firstID := strings.Split(lines[0], "\t")[0]
	secondID := strings.Split(lines[1], "\t")[0]
	if firstID != strconv.FormatInt(b.ID, 10) {
		t.Errorf("expected higher id %d first, got %q", b.ID, firstID)
	}
	if secondID != strconv.FormatInt(a.ID, 10) {
		t.Errorf("expected lower id %d second, got %q", a.ID, secondID)
	}
}

func TestListCmd_RespectsDBFlag(t *testing.T) {
	dbA := filepath.Join(t.TempDir(), "a.db")
	dbB := filepath.Join(t.TempDir(), "b.db")

	for _, tc := range []struct {
		path, title string
	}{
		{dbA, "in-A"},
		{dbB, "in-B"},
	} {
		s, err := storage.Open(tc.path)
		if err != nil {
			t.Fatalf("storage.Open(%s): %v", tc.path, err)
		}
		if _, err := s.Add(storage.Entry{Title: tc.title}); err != nil {
			t.Fatalf("Add(%s): %v", tc.title, err)
		}
		if err := s.Close(); err != nil {
			t.Fatalf("storage.Close(%s): %v", tc.path, err)
		}
	}

	root, outBuf, errBuf := newListTestRoot(t, dbA)
	root.SetArgs([]string{"--db", dbA, "list"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errBuf.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", errBuf.String())
	}
	out := outBuf.String()
	if !strings.Contains(out, "in-A") {
		t.Errorf("expected output to contain %q, got %q", "in-A", out)
	}
	if strings.Contains(out, "in-B") {
		t.Errorf("expected output NOT to contain %q, got %q", "in-B", out)
	}
}

func TestListCmd_StorageOpenErrorIsInternal(t *testing.T) {
	// Pointing --db at a directory (not a file) should fail inside
	// storage.Open. The returned error must NOT be an ErrUser so
	// main.go maps it to exit 2 (internal).
	dir := t.TempDir()

	root, outBuf, errBuf := newListTestRoot(t, dir)
	root.SetArgs([]string{"--db", dir, "list"})
	err := root.Execute()
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if errors.Is(err, ErrUser) {
		t.Fatalf("expected !errors.Is(err, ErrUser); got %v", err)
	}
	if outBuf.Len() != 0 {
		t.Fatalf("expected empty stdout, got %q", outBuf.String())
	}
	_ = errBuf // stderr may carry cobra/main prefixing, not asserted here
}

func TestListCmd_HelpShape(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	root, outBuf, errBuf := newListTestRoot(t, dbPath)
	root.SetArgs([]string{"list", "--help"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errBuf.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", errBuf.String())
	}
	out := outBuf.String()
	if !strings.Contains(out, "Usage:") {
		t.Errorf("expected help to contain %q, got %q", "Usage:", out)
	}

	// Find the list subcommand's declared Short string; help must include it.
	var listShort string
	for _, c := range root.Commands() {
		if c.Name() == "list" {
			listShort = c.Short
			break
		}
	}
	if listShort == "" {
		t.Fatalf("list subcommand not registered")
	}
	if !strings.Contains(out, listShort) {
		t.Errorf("expected help to contain Short %q, got %q", listShort, out)
	}
}
