package cli

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jysf/bragfile000/internal/editor"
	"github.com/jysf/bragfile000/internal/storage"
	"github.com/spf13/cobra"
)

// newRootWithEdit builds a fresh root command with the edit subcommand
// attached, installs editFn as testEditFunc (nil-safe), and returns the
// root plus a t.TempDir()-backed DB path. Cleanup resets testEditFunc.
func newRootWithEdit(t *testing.T, editFn editor.EditFunc) (*cobra.Command, string) {
	t.Helper()
	testEditFunc = editFn
	t.Cleanup(func() { testEditFunc = nil })

	root := NewRootCmd("test")
	root.AddCommand(NewEditCmd())
	dbPath := filepath.Join(t.TempDir(), "test.db")
	return root, dbPath
}

func seedEditEntry(t *testing.T, dbPath string, e storage.Entry) storage.Entry {
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
	return inserted
}

func getEntry(t *testing.T, dbPath string, id int64) storage.Entry {
	t.Helper()
	s, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("storage.Open: %v", err)
	}
	defer s.Close()
	got, err := s.Get(id)
	if err != nil {
		t.Fatalf("Get(%d): %v", id, err)
	}
	return got
}

// initEmptyStore ensures the DB file exists (migrations applied) without
// seeding any rows. Used by not-found tests so the "store not found"
// path doesn't mask "entry not found".
func initEmptyStore(t *testing.T, dbPath string) {
	t.Helper()
	s, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("storage.Open: %v", err)
	}
	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

func TestEditCmd_HappyPath(t *testing.T) {
	editFn := func(path string) error {
		return os.WriteFile(path, []byte("Title: NEW TITLE\n\nNEW BODY\n"), 0o600)
	}
	root, dbPath := newRootWithEdit(t, editFn)
	inserted := seedEditEntry(t, dbPath, storage.Entry{Title: "orig", Description: "orig body"})

	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"--db", dbPath, "edit", "1"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if outBuf.Len() != 0 {
		t.Errorf("expected stdout empty, got %q", outBuf.String())
	}
	if !strings.Contains(errBuf.String(), "Updated.") {
		t.Errorf("expected stderr to contain %q, got %q", "Updated.", errBuf.String())
	}
	got := getEntry(t, dbPath, inserted.ID)
	if got.Title != "NEW TITLE" {
		t.Errorf("Title after edit = %q, want %q", got.Title, "NEW TITLE")
	}
	if !strings.Contains(got.Description, "NEW BODY") {
		t.Errorf("Description after edit = %q; expected to contain %q", got.Description, "NEW BODY")
	}
}

func TestEditCmd_UnchangedBufferPrintsNoChanges(t *testing.T) {
	editFn := func(path string) error { return nil }
	root, dbPath := newRootWithEdit(t, editFn)
	inserted := seedEditEntry(t, dbPath, storage.Entry{Title: "keep me"})

	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"--db", dbPath, "edit", "1"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if outBuf.Len() != 0 {
		t.Errorf("expected stdout empty, got %q", outBuf.String())
	}
	if !strings.Contains(errBuf.String(), "No changes.") {
		t.Errorf("expected stderr to contain %q, got %q", "No changes.", errBuf.String())
	}
	got := getEntry(t, dbPath, inserted.ID)
	if got.Title != "keep me" {
		t.Errorf("Title changed unexpectedly: %q", got.Title)
	}
	delta := got.UpdatedAt.Sub(inserted.UpdatedAt)
	if delta < -time.Second || delta > time.Second {
		t.Errorf("UpdatedAt drifted: pre=%v post=%v (delta=%v)", inserted.UpdatedAt, got.UpdatedAt, delta)
	}
}

func TestEditCmd_NotFoundIsUserError(t *testing.T) {
	called := false
	editFn := func(path string) error {
		called = true
		return nil
	}
	root, dbPath := newRootWithEdit(t, editFn)
	initEmptyStore(t, dbPath)

	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"--db", dbPath, "edit", "999"})

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
	if called {
		t.Errorf("editor should not be invoked when entry is not found")
	}
}

func TestEditCmd_ParseErrorIsUserError(t *testing.T) {
	// Invalid buffer: no Title header at all.
	editFn := func(path string) error {
		return os.WriteFile(path, []byte("Tags: a\n\nbody\n"), 0o600)
	}
	root, dbPath := newRootWithEdit(t, editFn)
	inserted := seedEditEntry(t, dbPath, storage.Entry{Title: "orig"})

	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"--db", dbPath, "edit", "1"})

	err := root.Execute()
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, ErrUser) {
		t.Fatalf("expected errors.Is(err, ErrUser); got %v", err)
	}
	got := getEntry(t, dbPath, inserted.ID)
	if got.Title != "orig" {
		t.Errorf("Title = %q, want %q (row must be unchanged after parse error)", got.Title, "orig")
	}
}

func TestEditCmd_EditorErrorIsInternal(t *testing.T) {
	// Fake writes modified bytes before erroring so Launch takes the
	// "error + changed" path (per DEC #6, error + unchanged is an abort).
	editFn := func(path string) error {
		if err := os.WriteFile(path, []byte("Title: partial\n\n"), 0o600); err != nil {
			return err
		}
		return errors.New("editor exited non-zero")
	}
	root, dbPath := newRootWithEdit(t, editFn)
	inserted := seedEditEntry(t, dbPath, storage.Entry{Title: "orig"})

	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"--db", dbPath, "edit", "1"})

	err := root.Execute()
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if errors.Is(err, ErrUser) {
		t.Fatalf("editor failure must NOT be a user error; got %v", err)
	}
	got := getEntry(t, dbPath, inserted.ID)
	if got.Title != "orig" {
		t.Errorf("Title = %q, want %q (row must be unchanged after editor failure)", got.Title, "orig")
	}
}

func TestEditCmd_NoArgsIsUserError(t *testing.T) {
	root, dbPath := newRootWithEdit(t, nil)
	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"--db", dbPath, "edit"})

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

func TestEditCmd_TooManyArgsIsUserError(t *testing.T) {
	root, dbPath := newRootWithEdit(t, nil)
	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"--db", dbPath, "edit", "1", "2"})

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

func TestEditCmd_NonNumericArgIsUserError(t *testing.T) {
	root, dbPath := newRootWithEdit(t, nil)
	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"--db", dbPath, "edit", "abc"})

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

func TestEditCmd_NonPositiveArgIsUserError(t *testing.T) {
	root, dbPath := newRootWithEdit(t, nil)
	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"--db", dbPath, "edit", "0"})

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

func TestEditCmd_HelpShape(t *testing.T) {
	root, _ := newRootWithEdit(t, nil)
	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"edit", "--help"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errBuf.Len() != 0 {
		t.Fatalf("expected stderr empty, got %q", errBuf.String())
	}
	if !strings.Contains(outBuf.String(), "Examples:") {
		t.Errorf("expected help to contain %q, got %q", "Examples:", outBuf.String())
	}
}
