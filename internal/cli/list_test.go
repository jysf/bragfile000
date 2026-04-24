package cli

import (
	"bytes"
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jysf/bragfile000/internal/storage"
	"github.com/jysf/bragfile000/internal/storage/storagetest"
	"github.com/spf13/cobra"
)

// newListTestRoot builds a fresh root command with the list subcommand
// attached, wires separate stdout/stderr buffers, and clears
// BRAGFILE_DB so host env doesn't leak in. The caller drives args via
// cmd.SetArgs.
func newListTestRoot(t *testing.T) (*cobra.Command, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	t.Setenv("BRAGFILE_DB", "")
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

	root, outBuf, errBuf := newListTestRoot(t)
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

	root, outBuf, errBuf := newListTestRoot(t)
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

	root, outBuf, errBuf := newListTestRoot(t)
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

	root, outBuf, errBuf := newListTestRoot(t)
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

	root, outBuf, errBuf := newListTestRoot(t)
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

	root, outBuf, errBuf := newListTestRoot(t)
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

// seedListEntry inserts one entry with only the fields the filter tests
// need populated. Returns the inserted entry.
func seedListEntry(t *testing.T, dbPath, title, tags, project, typ string) storage.Entry {
	t.Helper()
	s, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("storage.Open: %v", err)
	}
	defer s.Close()
	e, err := s.Add(storage.Entry{Title: title, Tags: tags, Project: project, Type: typ})
	if err != nil {
		t.Fatalf("Add(%q): %v", title, err)
	}
	return e
}

// mustBackdate forwards to storagetest.Backdate and t.Fatals on error;
// CLI tests cannot import database/sql per the no-sql-in-cli-layer
// constraint, so the SQL UPDATE lives in the storagetest sub-package.
func mustBackdate(t *testing.T, dbPath string, id int64, at time.Time) {
	t.Helper()
	if err := storagetest.Backdate(dbPath, id, at); err != nil {
		t.Fatalf("storagetest.Backdate: %v", err)
	}
}

func runListCmd(t *testing.T, dbPath string, args ...string) (stdout, stderr string, runErr error) {
	t.Helper()
	root, outBuf, errBuf := newListTestRoot(t)
	full := append([]string{"--db", dbPath, "list"}, args...)
	root.SetArgs(full)
	runErr = root.Execute()
	return outBuf.String(), errBuf.String(), runErr
}

func TestListCmd_FilterByTag(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	seedListEntry(t, dbPath, "alpha-auth", "auth", "", "")
	seedListEntry(t, dbPath, "bravo-other", "backend", "", "")

	out, errOut, err := runListCmd(t, dbPath, "--tag", "auth")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errOut != "" {
		t.Fatalf("expected empty stderr, got %q", errOut)
	}
	if !strings.Contains(out, "alpha-auth") {
		t.Errorf("expected %q in stdout, got %q", "alpha-auth", out)
	}
	if strings.Contains(out, "bravo-other") {
		t.Errorf("expected %q NOT in stdout, got %q", "bravo-other", out)
	}
}

func TestListCmd_TagFilterNoFalsePositive(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	seedListEntry(t, dbPath, "exact-match", "auth", "", "")
	seedListEntry(t, dbPath, "looks-like", "authoring", "", "")

	out, errOut, err := runListCmd(t, dbPath, "--tag", "auth")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errOut != "" {
		t.Fatalf("expected empty stderr, got %q", errOut)
	}
	if !strings.Contains(out, "exact-match") {
		t.Errorf("expected %q in stdout, got %q", "exact-match", out)
	}
	if strings.Contains(out, "looks-like") {
		t.Errorf("expected %q NOT in stdout (substring false-positive), got %q", "looks-like", out)
	}
}

func TestListCmd_FilterByProject(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	seedListEntry(t, dbPath, "hit-platform", "", "platform", "")
	seedListEntry(t, dbPath, "miss-growth", "", "growth", "")

	out, errOut, err := runListCmd(t, dbPath, "--project", "platform")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errOut != "" {
		t.Fatalf("expected empty stderr, got %q", errOut)
	}
	if !strings.Contains(out, "hit-platform") {
		t.Errorf("expected %q in stdout, got %q", "hit-platform", out)
	}
	if strings.Contains(out, "miss-growth") {
		t.Errorf("expected %q NOT in stdout, got %q", "miss-growth", out)
	}
}

func TestListCmd_FilterByType(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	seedListEntry(t, dbPath, "hit-shipped", "", "", "shipped")
	seedListEntry(t, dbPath, "miss-learned", "", "", "learned")

	out, errOut, err := runListCmd(t, dbPath, "--type", "shipped")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errOut != "" {
		t.Fatalf("expected empty stderr, got %q", errOut)
	}
	if !strings.Contains(out, "hit-shipped") {
		t.Errorf("expected %q in stdout, got %q", "hit-shipped", out)
	}
	if strings.Contains(out, "miss-learned") {
		t.Errorf("expected %q NOT in stdout, got %q", "miss-learned", out)
	}
}

func TestListCmd_FilterBySince_ISODate(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	old := seedListEntry(t, dbPath, "old-entry", "", "", "")
	newEntry := seedListEntry(t, dbPath, "new-entry", "", "", "")

	mustBackdate(t, dbPath, old.ID, time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	mustBackdate(t, dbPath, newEntry.ID, time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC))

	out, errOut, err := runListCmd(t, dbPath, "--since", "2026-01-01")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errOut != "" {
		t.Fatalf("expected empty stderr, got %q", errOut)
	}
	if !strings.Contains(out, "new-entry") {
		t.Errorf("expected %q in stdout, got %q", "new-entry", out)
	}
	if strings.Contains(out, "old-entry") {
		t.Errorf("expected %q NOT in stdout, got %q", "old-entry", out)
	}
}

func TestListCmd_FilterBySince_Days(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	old := seedListEntry(t, dbPath, "ten-days-ago", "", "", "")
	newEntry := seedListEntry(t, dbPath, "two-days-ago", "", "", "")

	now := time.Now().UTC()
	mustBackdate(t, dbPath, old.ID, now.Add(-10*24*time.Hour))
	mustBackdate(t, dbPath, newEntry.ID, now.Add(-2*24*time.Hour))

	out, errOut, err := runListCmd(t, dbPath, "--since", "7d")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errOut != "" {
		t.Fatalf("expected empty stderr, got %q", errOut)
	}
	if !strings.Contains(out, "two-days-ago") {
		t.Errorf("expected %q in stdout, got %q", "two-days-ago", out)
	}
	if strings.Contains(out, "ten-days-ago") {
		t.Errorf("expected %q NOT in stdout, got %q", "ten-days-ago", out)
	}
}

func TestListCmd_FilterBySince_Weeks(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	old := seedListEntry(t, dbPath, "thirty-days-ago", "", "", "")
	newEntry := seedListEntry(t, dbPath, "ten-days-ago", "", "", "")

	now := time.Now().UTC()
	mustBackdate(t, dbPath, old.ID, now.Add(-30*24*time.Hour))
	mustBackdate(t, dbPath, newEntry.ID, now.Add(-10*24*time.Hour))

	out, errOut, err := runListCmd(t, dbPath, "--since", "2w")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errOut != "" {
		t.Fatalf("expected empty stderr, got %q", errOut)
	}
	if !strings.Contains(out, "ten-days-ago") {
		t.Errorf("expected %q in stdout, got %q", "ten-days-ago", out)
	}
	if strings.Contains(out, "thirty-days-ago") {
		t.Errorf("expected %q NOT in stdout, got %q", "thirty-days-ago", out)
	}
}

func TestListCmd_FilterBySince_Months(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	old := seedListEntry(t, dbPath, "one-hundred-days", "", "", "")
	newEntry := seedListEntry(t, dbPath, "sixty-days", "", "", "")

	now := time.Now().UTC()
	mustBackdate(t, dbPath, old.ID, now.Add(-100*24*time.Hour))
	mustBackdate(t, dbPath, newEntry.ID, now.Add(-60*24*time.Hour))

	// 3m = 90 days per DEC-008 approximation.
	out, errOut, err := runListCmd(t, dbPath, "--since", "3m")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errOut != "" {
		t.Fatalf("expected empty stderr, got %q", errOut)
	}
	if !strings.Contains(out, "sixty-days") {
		t.Errorf("expected %q in stdout, got %q", "sixty-days", out)
	}
	if strings.Contains(out, "one-hundred-days") {
		t.Errorf("expected %q NOT in stdout, got %q", "one-hundred-days", out)
	}
}

func TestListCmd_FilterByLimit(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	for _, title := range []string{"one", "two", "three", "four"} {
		seedListEntry(t, dbPath, title, "", "", "")
	}

	out, errOut, err := runListCmd(t, dbPath, "--limit", "2")
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
}

func TestListCmd_CombinedFilters(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	hit := seedListEntry(t, dbPath, "combined-hit", "", "platform", "")
	missP := seedListEntry(t, dbPath, "combined-missP", "", "growth", "")
	old := seedListEntry(t, dbPath, "combined-old", "", "platform", "")

	now := time.Now().UTC()
	mustBackdate(t, dbPath, hit.ID, now.Add(-2*24*time.Hour))
	mustBackdate(t, dbPath, missP.ID, now.Add(-2*24*time.Hour))
	mustBackdate(t, dbPath, old.ID, now.Add(-20*24*time.Hour))

	out, errOut, err := runListCmd(t, dbPath,
		"--project", "platform",
		"--since", "7d",
		"--limit", "5",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errOut != "" {
		t.Fatalf("expected empty stderr, got %q", errOut)
	}
	if !strings.Contains(out, "combined-hit") {
		t.Errorf("expected %q in stdout, got %q", "combined-hit", out)
	}
	if strings.Contains(out, "combined-missP") {
		t.Errorf("expected %q NOT in stdout, got %q", "combined-missP", out)
	}
	if strings.Contains(out, "combined-old") {
		t.Errorf("expected %q NOT in stdout, got %q", "combined-old", out)
	}
}

func TestListCmd_FilterPreservesOrder(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	a := seedListEntry(t, dbPath, "first-x", "x", "", "")
	b := seedListEntry(t, dbPath, "second-x", "x", "", "")
	c := seedListEntry(t, dbPath, "third-x", "x", "", "")

	out, errOut, err := runListCmd(t, dbPath, "--tag", "x")
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
	wantIDs := []int64{c.ID, b.ID, a.ID}
	for i, line := range lines {
		gotID := strings.Split(line, "\t")[0]
		if gotID != strconv.FormatInt(wantIDs[i], 10) {
			t.Errorf("line %d: id=%s, want %d", i, gotID, wantIDs[i])
		}
	}
}

func TestListCmd_InvalidSinceIsUserError(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	// Need the DB to exist so runList doesn't fail earlier; but the
	// since validation happens before storage open anyway. Still safe.
	out, _, err := runListCmd(t, dbPath, "--since", "notadate")
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

func TestListCmd_InvalidLimitIsUserError(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	for _, arg := range []string{"0", "-5"} {
		t.Run("limit="+arg, func(t *testing.T) {
			out, _, err := runListCmd(t, dbPath, "--limit", arg)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !errors.Is(err, ErrUser) {
				t.Errorf("expected errors.Is(err, ErrUser); got %v", err)
			}
			if out != "" {
				t.Errorf("expected empty stdout, got %q", out)
			}
		})
	}
}

func TestListCmd_EmptyFilterValueIsUserError(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	out, _, err := runListCmd(t, dbPath, "--tag", "")
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

func TestListCmd_HelpShowsFilters(t *testing.T) {
	root, outBuf, errBuf := newListTestRoot(t)
	root.SetArgs([]string{"list", "--help"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errBuf.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", errBuf.String())
	}
	out := outBuf.String()
	for _, needle := range []string{"--tag", "--project", "--type", "--since", "--limit", "Examples:"} {
		if !strings.Contains(out, needle) {
			t.Errorf("expected help to contain %q, got:\n%s", needle, out)
		}
	}
}

func TestListCmd_HelpShape(t *testing.T) {
	root, outBuf, errBuf := newListTestRoot(t)
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

// -----------------------------------------------------------------------------
// SPEC-013: `brag list --show-project` / `-P`
//
// Seven tests covering the five locked design decisions in the spec plus the
// byte-stability regression lock (decision 4 → PlainOutputByteIdenticalToSTAGE002)
// and the composition-with-all-filters case (decision 5 → ComposedWithAllFilters).
// -----------------------------------------------------------------------------

func TestListCmd_PlainOutputByteIdenticalToSTAGE002(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	entries := []storage.Entry{
		seedListEntry(t, dbPath, "first-title", "", "platform", ""),
		seedListEntry(t, dbPath, "second-title", "", "", ""),
		seedListEntry(t, dbPath, "third-title", "", "growth", ""),
	}

	out, errOut, err := runListCmd(t, dbPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errOut != "" {
		t.Fatalf("expected empty stderr, got %q", errOut)
	}

	var expected strings.Builder
	for i := len(entries) - 1; i >= 0; i-- {
		e := entries[i]
		fmt.Fprintf(&expected, "%d\t%s\t%s\n",
			e.ID,
			e.CreatedAt.UTC().Format(time.RFC3339),
			e.Title)
	}
	if out != expected.String() {
		t.Fatalf("plain list output not byte-identical to STAGE-002 shape\nwant: %q\ngot:  %q", expected.String(), out)
	}
}

func TestListCmd_ShowProject_AddsFourthColumn(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	seedListEntry(t, dbPath, "solo", "", "platform", "")

	out, errOut, err := runListCmd(t, dbPath, "-P")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errOut != "" {
		t.Fatalf("expected empty stderr, got %q", errOut)
	}

	line := strings.TrimRight(out, "\n")
	if strings.Contains(line, "\n") {
		t.Fatalf("expected a single line, got %q", out)
	}
	if got := strings.Count(line, "\t"); got != 3 {
		t.Fatalf("expected exactly 3 tab characters, got %d: %q", got, line)
	}
	fields := strings.Split(line, "\t")
	if len(fields) != 4 {
		t.Fatalf("expected 4 fields, got %d: %q", len(fields), line)
	}
	if _, err := strconv.ParseInt(fields[0], 10, 64); err != nil {
		t.Errorf("field 0 (id) %q: %v", fields[0], err)
	}
	ts, err := time.Parse(time.RFC3339, fields[1])
	if err != nil {
		t.Errorf("field 1 (created_at) %q: %v", fields[1], err)
	} else if loc := ts.Location().String(); loc != "UTC" {
		t.Errorf("field 1 (created_at): expected UTC location, got %q", loc)
	}
	if fields[2] != "platform" {
		t.Errorf("field 2 (project): want %q, got %q", "platform", fields[2])
	}
	if fields[3] != "solo" {
		t.Errorf("field 3 (title): want %q, got %q", "solo", fields[3])
	}
}

func TestListCmd_ShowProject_EmptyProjectRendersAsDash(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	seedListEntry(t, dbPath, "no-project", "", "", "")

	out, errOut, err := runListCmd(t, dbPath, "-P")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errOut != "" {
		t.Fatalf("expected empty stderr, got %q", errOut)
	}
	line := strings.TrimRight(out, "\n")
	fields := strings.Split(line, "\t")
	if len(fields) != 4 {
		t.Fatalf("expected 4 fields, got %d: %q", len(fields), line)
	}
	if fields[2] != "-" {
		t.Errorf("empty project should render as %q, got %q", "-", fields[2])
	}
	if fields[3] != "no-project" {
		t.Errorf("field 3 (title): want %q, got %q", "no-project", fields[3])
	}
}

func TestListCmd_ShowProject_ShortFormEquivalent(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	seedListEntry(t, dbPath, "alpha", "", "platform", "")
	seedListEntry(t, dbPath, "bravo", "", "growth", "")

	outLong, errLong, err := runListCmd(t, dbPath, "--show-project")
	if err != nil {
		t.Fatalf("--show-project: unexpected error: %v", err)
	}
	if errLong != "" {
		t.Fatalf("--show-project: expected empty stderr, got %q", errLong)
	}

	outShort, errShort, err := runListCmd(t, dbPath, "-P")
	if err != nil {
		t.Fatalf("-P: unexpected error: %v", err)
	}
	if errShort != "" {
		t.Fatalf("-P: expected empty stderr, got %q", errShort)
	}

	if outLong != outShort {
		t.Fatalf("--show-project and -P produced different output\nlong:  %q\nshort: %q", outLong, outShort)
	}
}

func TestListCmd_ShowProject_WithProjectFilter(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	seedListEntry(t, dbPath, "hit-platform", "", "platform", "")
	seedListEntry(t, dbPath, "miss-growth", "", "growth", "")

	out, errOut, err := runListCmd(t, dbPath, "-P", "--project", "platform")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errOut != "" {
		t.Fatalf("expected empty stderr, got %q", errOut)
	}
	if !strings.Contains(out, "hit-platform") {
		t.Errorf("expected %q in stdout, got %q", "hit-platform", out)
	}
	if strings.Contains(out, "miss-growth") {
		t.Errorf("expected %q NOT in stdout, got %q", "miss-growth", out)
	}
	line := strings.TrimRight(out, "\n")
	if strings.Contains(line, "\n") {
		t.Fatalf("expected a single line, got %q", out)
	}
	if got := strings.Count(line, "\t"); got != 3 {
		t.Fatalf("expected exactly 3 tab characters, got %d: %q", got, line)
	}
	fields := strings.Split(line, "\t")
	if fields[2] != "platform" {
		t.Errorf("field 2 (project): want %q, got %q", "platform", fields[2])
	}
}

func TestListCmd_ShowProject_ComposedWithAllFilters(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	seedListEntry(t, dbPath, "hit", "auth", "platform", "shipped")
	seedListEntry(t, dbPath, "miss-tag", "backend", "platform", "shipped")
	seedListEntry(t, dbPath, "miss-project", "auth", "growth", "shipped")

	out, errOut, err := runListCmd(t, dbPath,
		"-P",
		"--tag", "auth",
		"--project", "platform",
		"--type", "shipped",
		"--since", "1900-01-01",
		"--limit", "5",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errOut != "" {
		t.Fatalf("expected empty stderr, got %q", errOut)
	}
	if !strings.Contains(out, "hit") {
		t.Errorf("expected %q in stdout, got %q", "hit", out)
	}
	if strings.Contains(out, "miss-tag") {
		t.Errorf("expected %q NOT in stdout, got %q", "miss-tag", out)
	}
	if strings.Contains(out, "miss-project") {
		t.Errorf("expected %q NOT in stdout, got %q", "miss-project", out)
	}
	line := strings.TrimRight(out, "\n")
	if strings.Contains(line, "\n") {
		t.Fatalf("expected exactly one surviving row, got %q", out)
	}
	if got := strings.Count(line, "\t"); got != 3 {
		t.Fatalf("expected exactly 3 tab characters, got %d: %q", got, line)
	}
	fields := strings.Split(line, "\t")
	if fields[2] != "platform" {
		t.Errorf("field 2 (project): want %q, got %q", "platform", fields[2])
	}
	if fields[3] != "hit" {
		t.Errorf("field 3 (title): want %q, got %q", "hit", fields[3])
	}
}

func TestListCmd_ShowProject_HelpShowsFlag(t *testing.T) {
	root, outBuf, errBuf := newListTestRoot(t)
	root.SetArgs([]string{"list", "--help"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errBuf.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", errBuf.String())
	}
	out := outBuf.String()
	for _, needle := range []string{"--show-project", "-P", "include project in output"} {
		if !strings.Contains(out, needle) {
			t.Errorf("expected help to contain %q, got:\n%s", needle, out)
		}
	}
}
