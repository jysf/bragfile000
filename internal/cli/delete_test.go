package cli

import (
	"bytes"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jysf/bragfile000/internal/storage"
	"github.com/spf13/cobra"
)

// newRootWithDelete builds a fresh root command with the delete subcommand
// attached and returns the root plus a t.TempDir()-backed DB path.
func newRootWithDelete(t *testing.T) (*cobra.Command, string) {
	t.Helper()
	root := NewRootCmd("test")
	root.AddCommand(NewDeleteCmd())
	dbPath := filepath.Join(t.TempDir(), "test.db")
	return root, dbPath
}

// seedDeleteEntry opens the store at dbPath, adds e, closes, and returns
// the inserted entry's ID.
func seedDeleteEntry(t *testing.T, dbPath string, e storage.Entry) int64 {
	t.Helper()
	s, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("storage.Open: %v", err)
	}
	inserted, err := s.Add(e)
	if err != nil {
		_ = s.Close()
		t.Fatalf("Add: %v", err)
	}
	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	return inserted.ID
}

// entryExists opens the store at dbPath and returns whether id is
// retrievable. Used to verify row presence/absence after delete.
func entryExists(t *testing.T, dbPath string, id int64) bool {
	t.Helper()
	s, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("storage.Open: %v", err)
	}
	defer s.Close()
	_, err = s.Get(id)
	if err == nil {
		return true
	}
	if errors.Is(err, storage.ErrNotFound) {
		return false
	}
	t.Fatalf("entryExists: unexpected error: %v", err)
	return false
}

func TestDeleteCmd_ConfirmY(t *testing.T) {
	root, dbPath := newRootWithDelete(t)
	id := seedDeleteEntry(t, dbPath, storage.Entry{Title: "shipped auth refactor"})

	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetIn(strings.NewReader("y\n"))
	root.SetArgs([]string{"--db", dbPath, "delete", "1"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if outBuf.Len() != 0 {
		t.Errorf("expected stdout to be empty, got %q", outBuf.String())
	}
	errOut := errBuf.String()
	for _, want := range []string{"Delete entry 1", `"shipped auth refactor"`, "Deleted."} {
		if !strings.Contains(errOut, want) {
			t.Errorf("expected stderr to contain %q, got %q", want, errOut)
		}
	}
	if entryExists(t, dbPath, id) {
		t.Errorf("row %d should have been deleted", id)
	}
}

func TestDeleteCmd_DeclineN(t *testing.T) {
	root, dbPath := newRootWithDelete(t)
	id := seedDeleteEntry(t, dbPath, storage.Entry{Title: "keep me"})

	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetIn(strings.NewReader("n\n"))
	root.SetArgs([]string{"--db", dbPath, "delete", "1"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if outBuf.Len() != 0 {
		t.Errorf("expected stdout to be empty, got %q", outBuf.String())
	}
	if !strings.Contains(errBuf.String(), "Aborted.") {
		t.Errorf("expected stderr to contain %q, got %q", "Aborted.", errBuf.String())
	}
	if !entryExists(t, dbPath, id) {
		t.Errorf("row %d should still exist after decline", id)
	}
}

func TestDeleteCmd_DeclineEmpty(t *testing.T) {
	root, dbPath := newRootWithDelete(t)
	id := seedDeleteEntry(t, dbPath, storage.Entry{Title: "keep me"})

	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetIn(strings.NewReader("\n"))
	root.SetArgs([]string{"--db", dbPath, "delete", "1"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if outBuf.Len() != 0 {
		t.Errorf("expected stdout to be empty, got %q", outBuf.String())
	}
	if !strings.Contains(errBuf.String(), "Aborted.") {
		t.Errorf("expected stderr to contain %q, got %q", "Aborted.", errBuf.String())
	}
	if !entryExists(t, dbPath, id) {
		t.Errorf("row %d should still exist after empty decline", id)
	}
}

func TestDeleteCmd_DeclineOther(t *testing.T) {
	root, dbPath := newRootWithDelete(t)
	id := seedDeleteEntry(t, dbPath, storage.Entry{Title: "keep me"})

	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetIn(strings.NewReader("maybe\n"))
	root.SetArgs([]string{"--db", dbPath, "delete", "1"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if outBuf.Len() != 0 {
		t.Errorf("expected stdout to be empty, got %q", outBuf.String())
	}
	if !strings.Contains(errBuf.String(), "Aborted.") {
		t.Errorf("expected stderr to contain %q, got %q", "Aborted.", errBuf.String())
	}
	if !entryExists(t, dbPath, id) {
		t.Errorf("row %d should still exist after non-y response", id)
	}
}

func TestDeleteCmd_YesFlagLongAndShort(t *testing.T) {
	// Run once with --yes, once with -y. Seed two entries so each run
	// has its own row to delete.
	root1, dbPath := newRootWithDelete(t)
	id1 := seedDeleteEntry(t, dbPath, storage.Entry{Title: "first"})
	id2 := seedDeleteEntry(t, dbPath, storage.Entry{Title: "second"})

	var out1, err1 bytes.Buffer
	root1.SetOut(&out1)
	root1.SetErr(&err1)
	root1.SetArgs([]string{"--db", dbPath, "delete", "1", "--yes"})
	if err := root1.Execute(); err != nil {
		t.Fatalf("--yes: unexpected error: %v", err)
	}
	if out1.Len() != 0 {
		t.Errorf("--yes: expected stdout empty, got %q", out1.String())
	}
	if !strings.Contains(err1.String(), "Deleted.") {
		t.Errorf("--yes: expected stderr to contain %q, got %q", "Deleted.", err1.String())
	}
	if entryExists(t, dbPath, id1) {
		t.Errorf("--yes: row %d should have been deleted", id1)
	}

	// Fresh root for the -y case (cobra caches parsed args per Execute call).
	root2 := NewRootCmd("test")
	root2.AddCommand(NewDeleteCmd())

	var out2, err2 bytes.Buffer
	root2.SetOut(&out2)
	root2.SetErr(&err2)
	root2.SetArgs([]string{"--db", dbPath, "delete", "2", "-y"})
	if err := root2.Execute(); err != nil {
		t.Fatalf("-y: unexpected error: %v", err)
	}
	if out2.Len() != 0 {
		t.Errorf("-y: expected stdout empty, got %q", out2.String())
	}
	if !strings.Contains(err2.String(), "Deleted.") {
		t.Errorf("-y: expected stderr to contain %q, got %q", "Deleted.", err2.String())
	}
	if entryExists(t, dbPath, id2) {
		t.Errorf("-y: row %d should have been deleted", id2)
	}
}

func TestDeleteCmd_NotFoundIsUserError(t *testing.T) {
	root, dbPath := newRootWithDelete(t)
	// Open + close the store so migrations run and the file exists.
	s, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("storage.Open: %v", err)
	}
	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"--db", dbPath, "delete", "999"})

	err = root.Execute()
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, ErrUser) {
		t.Fatalf("expected errors.Is(err, ErrUser); got %v", err)
	}
	if outBuf.Len() != 0 {
		t.Errorf("expected stdout empty, got %q", outBuf.String())
	}
	if strings.Contains(errBuf.String(), "Delete entry") {
		t.Errorf("expected no prompt to reach stderr before Get fails, got %q", errBuf.String())
	}
}

func TestDeleteCmd_NotFoundWithYesIsUserError(t *testing.T) {
	root, dbPath := newRootWithDelete(t)
	s, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("storage.Open: %v", err)
	}
	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"--db", dbPath, "delete", "999", "--yes"})

	err = root.Execute()
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, ErrUser) {
		t.Fatalf("expected errors.Is(err, ErrUser); got %v", err)
	}
	if outBuf.Len() != 0 {
		t.Errorf("expected stdout empty, got %q", outBuf.String())
	}
}

func TestDeleteCmd_NoArgsIsUserError(t *testing.T) {
	root, dbPath := newRootWithDelete(t)
	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"--db", dbPath, "delete"})

	err := root.Execute()
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, ErrUser) {
		t.Fatalf("expected errors.Is(err, ErrUser); got %v", err)
	}
	if outBuf.Len() != 0 {
		t.Errorf("expected stdout empty, got %q", outBuf.String())
	}
}

func TestDeleteCmd_TooManyArgsIsUserError(t *testing.T) {
	root, dbPath := newRootWithDelete(t)
	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"--db", dbPath, "delete", "1", "2"})

	err := root.Execute()
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, ErrUser) {
		t.Fatalf("expected errors.Is(err, ErrUser); got %v", err)
	}
	if outBuf.Len() != 0 {
		t.Errorf("expected stdout empty, got %q", outBuf.String())
	}
}

func TestDeleteCmd_NonNumericArgIsUserError(t *testing.T) {
	root, dbPath := newRootWithDelete(t)
	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"--db", dbPath, "delete", "abc"})

	err := root.Execute()
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, ErrUser) {
		t.Fatalf("expected errors.Is(err, ErrUser); got %v", err)
	}
	if outBuf.Len() != 0 {
		t.Errorf("expected stdout empty, got %q", outBuf.String())
	}
}

func TestDeleteCmd_NonPositiveArgIsUserError(t *testing.T) {
	root, dbPath := newRootWithDelete(t)
	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"--db", dbPath, "delete", "0"})

	err := root.Execute()
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, ErrUser) {
		t.Fatalf("expected errors.Is(err, ErrUser); got %v", err)
	}
	if outBuf.Len() != 0 {
		t.Errorf("expected stdout empty, got %q", outBuf.String())
	}
}

func TestDeleteCmd_HelpShape(t *testing.T) {
	root, _ := newRootWithDelete(t)
	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"delete", "--help"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errBuf.Len() != 0 {
		t.Fatalf("expected stderr empty, got %q", errBuf.String())
	}
	out := outBuf.String()
	for _, want := range []string{"Examples:", "-y, --yes"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected help to contain %q, got %q", want, out)
		}
	}
}
