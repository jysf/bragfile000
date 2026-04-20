package cli

import (
	"bytes"
	"errors"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/jysf/bragfile000/internal/storage"
	"github.com/spf13/cobra"
)

// newRootWithAdd builds a fresh root command with the add subcommand
// attached and returns the root plus a t.TempDir()-backed DB path.
// Caller is responsible for setting args, out, err.
func newRootWithAdd(t *testing.T) (*cobra.Command, string) {
	t.Helper()
	root := NewRootCmd("test")
	root.AddCommand(NewAddCmd())
	dbPath := filepath.Join(t.TempDir(), "test.db")
	return root, dbPath
}

func TestAdd_SuccessPrintsIDToStdoutOnly(t *testing.T) {
	root, dbPath := newRootWithAdd(t)
	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"--db", dbPath, "add", "--title", "first brag"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errBuf.Len() != 0 {
		t.Fatalf("expected stderr to be empty, got %q", errBuf.String())
	}
	idStr := strings.TrimSpace(outBuf.String())
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		t.Fatalf("expected stdout to parse as integer, got %q: %v", idStr, err)
	}
	if id <= 0 {
		t.Fatalf("expected positive id, got %d", id)
	}

	s, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("storage.Open: %v", err)
	}
	defer s.Close()
	entries, err := s.List(storage.ListFilter{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Title != "first brag" {
		t.Errorf("expected title %q, got %q", "first brag", entries[0].Title)
	}
}

func TestAdd_OutputIsPipeable(t *testing.T) {
	root, dbPath := newRootWithAdd(t)
	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"--db", dbPath, "add", "--title", "pipeable"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errBuf.Len() != 0 {
		t.Fatalf("expected stderr to be empty, got %q", errBuf.String())
	}
	re := regexp.MustCompile(`^\d+\n$`)
	if !re.MatchString(outBuf.String()) {
		t.Fatalf("expected stdout to match %q, got %q", `^\d+\n$`, outBuf.String())
	}
}

func TestAdd_MissingTitleIsUserError(t *testing.T) {
	root, dbPath := newRootWithAdd(t)
	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"--db", dbPath, "add"})

	err := root.Execute()
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, ErrUser) {
		t.Fatalf("expected errors.Is(err, ErrUser) to be true; got %v", err)
	}
}

func TestAdd_EmptyTitleIsUserError(t *testing.T) {
	root, dbPath := newRootWithAdd(t)
	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"--db", dbPath, "add", "--title", ""})

	err := root.Execute()
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, ErrUser) {
		t.Fatalf("expected errors.Is(err, ErrUser) to be true; got %v", err)
	}
}

func TestAdd_WhitespaceTitleIsUserError(t *testing.T) {
	root, dbPath := newRootWithAdd(t)
	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"--db", dbPath, "add", "--title", "   "})

	err := root.Execute()
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, ErrUser) {
		t.Fatalf("expected errors.Is(err, ErrUser) to be true; got %v", err)
	}
}

func TestAdd_AllOptionalFieldsPersisted(t *testing.T) {
	root, dbPath := newRootWithAdd(t)
	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{
		"--db", dbPath, "add",
		"--title", "x",
		"--description", "why",
		"--tags", "a,b",
		"--project", "p",
		"--type", "t",
		"--impact", "i",
	})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	s, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("storage.Open: %v", err)
	}
	defer s.Close()
	entries, err := s.List(storage.ListFilter{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	e := entries[0]
	if e.Title != "x" {
		t.Errorf("title: got %q want %q", e.Title, "x")
	}
	if e.Description != "why" {
		t.Errorf("description: got %q want %q", e.Description, "why")
	}
	if e.Tags != "a,b" {
		t.Errorf("tags: got %q want %q", e.Tags, "a,b")
	}
	if e.Project != "p" {
		t.Errorf("project: got %q want %q", e.Project, "p")
	}
	if e.Type != "t" {
		t.Errorf("type: got %q want %q", e.Type, "t")
	}
	if e.Impact != "i" {
		t.Errorf("impact: got %q want %q", e.Impact, "i")
	}
}

func TestAdd_TwoAddsDistinctIDs(t *testing.T) {
	root, dbPath := newRootWithAdd(t)
	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"--db", dbPath, "add", "--title", "same"})
	if err := root.Execute(); err != nil {
		t.Fatalf("first add: %v", err)
	}

	// Fresh root for the second invocation so cobra re-parses args cleanly.
	root2 := NewRootCmd("test")
	root2.AddCommand(NewAddCmd())
	root2.SetOut(&outBuf)
	root2.SetErr(&errBuf)
	root2.SetArgs([]string{"--db", dbPath, "add", "--title", "same"})
	if err := root2.Execute(); err != nil {
		t.Fatalf("second add: %v", err)
	}
	if errBuf.Len() != 0 {
		t.Fatalf("expected stderr to be empty, got %q", errBuf.String())
	}

	lines := strings.Split(strings.TrimRight(outBuf.String(), "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 ID lines, got %d: %q", len(lines), outBuf.String())
	}
	id1, err := strconv.ParseInt(lines[0], 10, 64)
	if err != nil {
		t.Fatalf("parse id1 %q: %v", lines[0], err)
	}
	id2, err := strconv.ParseInt(lines[1], 10, 64)
	if err != nil {
		t.Fatalf("parse id2 %q: %v", lines[1], err)
	}
	if id1 == id2 {
		t.Fatalf("expected distinct IDs, got %d == %d", id1, id2)
	}

	s, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("storage.Open: %v", err)
	}
	defer s.Close()
	entries, err := s.List(storage.ListFilter{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
}

func TestAdd_HelpListsAllFlags(t *testing.T) {
	root, _ := newRootWithAdd(t)
	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"add", "--help"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errBuf.Len() != 0 {
		t.Fatalf("expected stderr to be empty, got %q", errBuf.String())
	}
	out := outBuf.String()
	for _, flag := range []string{"--title", "--description", "--tags", "--project", "--type", "--impact", "--db"} {
		if !strings.Contains(out, flag) {
			t.Errorf("expected help to mention %q, got %q", flag, out)
		}
	}
}
