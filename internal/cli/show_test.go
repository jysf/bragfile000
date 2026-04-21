package cli

import (
	"bytes"
	"errors"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/jysf/bragfile000/internal/storage"
	"github.com/spf13/cobra"
)

// newRootWithShow builds a fresh root command with the show subcommand
// attached and returns the root plus a t.TempDir()-backed DB path.
func newRootWithShow(t *testing.T) (*cobra.Command, string) {
	t.Helper()
	root := NewRootCmd("test")
	root.AddCommand(NewShowCmd())
	dbPath := filepath.Join(t.TempDir(), "test.db")
	return root, dbPath
}

// seedEntry opens the store at dbPath, adds e, closes, and returns the
// inserted entry's ID.
func seedEntry(t *testing.T, dbPath string, e storage.Entry) int64 {
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

func TestShowCmd_FullEntryRendersAllSections(t *testing.T) {
	root, dbPath := newRootWithShow(t)
	id := seedEntry(t, dbPath, storage.Entry{
		Title:       "cut p99 latency",
		Description: "redis lookup",
		Tags:        "auth,perf",
		Project:     "platform",
		Type:        "shipped",
		Impact:      "unblocked mobile v3",
	})

	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"--db", dbPath, "show", strconv.FormatInt(id, 10)})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errBuf.Len() != 0 {
		t.Fatalf("expected stderr to be empty, got %q", errBuf.String())
	}
	out := outBuf.String()
	for _, want := range []string{
		"# cut p99 latency",
		"auth,perf",
		"platform",
		"shipped",
		"unblocked mobile v3",
		"## Description",
		"redis lookup",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("expected stdout to contain %q, got %q", want, out)
		}
	}
}

func TestShowCmd_EmptyMetadataRowsOmitted(t *testing.T) {
	root, dbPath := newRootWithShow(t)
	id := seedEntry(t, dbPath, storage.Entry{Title: "only title"})

	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"--db", dbPath, "show", strconv.FormatInt(id, 10)})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errBuf.Len() != 0 {
		t.Fatalf("expected stderr to be empty, got %q", errBuf.String())
	}
	out := outBuf.String()
	if !strings.Contains(out, "# only title") {
		t.Errorf("expected title line %q, got %q", "# only title", out)
	}
	for _, label := range []string{"| tags", "| project", "| type", "| impact"} {
		if strings.Contains(out, label) {
			t.Errorf("expected stdout NOT to contain %q (empty row leaked), got %q", label, out)
		}
	}
}

func TestShowCmd_EmptyDescriptionSectionOmitted(t *testing.T) {
	root, dbPath := newRootWithShow(t)
	id := seedEntry(t, dbPath, storage.Entry{Title: "only title"})

	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"--db", dbPath, "show", strconv.FormatInt(id, 10)})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errBuf.Len() != 0 {
		t.Fatalf("expected stderr to be empty, got %q", errBuf.String())
	}
	if strings.Contains(outBuf.String(), "## Description") {
		t.Errorf("expected stdout NOT to contain %q, got %q", "## Description", outBuf.String())
	}
}

func TestShowCmd_NotFoundIsUserError(t *testing.T) {
	root, dbPath := newRootWithShow(t)
	// Open + close the store so migrations run and the file exists, but
	// the entries table is empty.
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
	root.SetArgs([]string{"--db", dbPath, "show", "999"})

	err = root.Execute()
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, ErrUser) {
		t.Fatalf("expected errors.Is(err, ErrUser); got %v", err)
	}
	if outBuf.Len() != 0 {
		t.Errorf("expected stdout to be empty, got %q", outBuf.String())
	}
	if errBuf.Len() != 0 {
		t.Errorf("expected stderr to be empty (main.go formats user errors), got %q", errBuf.String())
	}
}

func TestShowCmd_NoArgsIsUserError(t *testing.T) {
	root, dbPath := newRootWithShow(t)
	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"--db", dbPath, "show"})

	err := root.Execute()
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, ErrUser) {
		t.Fatalf("expected errors.Is(err, ErrUser); got %v", err)
	}
}

func TestShowCmd_TooManyArgsIsUserError(t *testing.T) {
	root, dbPath := newRootWithShow(t)
	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"--db", dbPath, "show", "1", "2"})

	err := root.Execute()
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, ErrUser) {
		t.Fatalf("expected errors.Is(err, ErrUser); got %v", err)
	}
}

func TestShowCmd_NonNumericArgIsUserError(t *testing.T) {
	root, dbPath := newRootWithShow(t)
	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"--db", dbPath, "show", "abc"})

	err := root.Execute()
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, ErrUser) {
		t.Fatalf("expected errors.Is(err, ErrUser); got %v", err)
	}
}

func TestShowCmd_NonPositiveArgIsUserError(t *testing.T) {
	root, dbPath := newRootWithShow(t)
	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"--db", dbPath, "show", "0"})

	err := root.Execute()
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, ErrUser) {
		t.Fatalf("expected errors.Is(err, ErrUser); got %v", err)
	}
}

func TestShowCmd_HelpShape(t *testing.T) {
	root, _ := newRootWithShow(t)
	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"show", "--help"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errBuf.Len() != 0 {
		t.Fatalf("expected stderr to be empty, got %q", errBuf.String())
	}
	if !strings.Contains(outBuf.String(), "Examples:") {
		t.Errorf("expected help to contain distinctive %q label, got %q", "Examples:", outBuf.String())
	}
}
