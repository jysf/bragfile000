package cli

import (
	"bytes"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func newTagTestRoot(t *testing.T) (*cobra.Command, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	t.Setenv("BRAGFILE_DB", "")
	root := NewRootCmd("test")
	root.AddCommand(NewTagCmd())
	// Also attach list and tags so we can assert post-state.
	root.AddCommand(NewListCmd())
	root.AddCommand(NewTagsCmd())
	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	root.SetOut(outBuf)
	root.SetErr(errBuf)
	return root, outBuf, errBuf
}

func runTagSubCmd(t *testing.T, dbPath string, subArgs ...string) (stdout, stderr string, runErr error) {
	t.Helper()
	root, outBuf, errBuf := newTagTestRoot(t)
	full := append([]string{"--db", dbPath, "tag"}, subArgs...)
	root.SetArgs(full)
	runErr = root.Execute()
	return outBuf.String(), errBuf.String(), runErr
}

// runListForTag runs `brag list --tag <tag>` against the test root and
// returns stdout lines (trimmed). CLI layer only — no direct storage import.
func runListForTag(t *testing.T, dbPath, tag string) []string {
	t.Helper()
	root, outBuf, errBuf := newTagTestRoot(t)
	root.SetArgs([]string{"--db", dbPath, "list", "--tag", tag})
	if err := root.Execute(); err != nil {
		t.Fatalf("list --tag %q: %v (stderr=%q)", tag, err, errBuf.String())
	}
	out := strings.TrimRight(outBuf.String(), "\n")
	if out == "" {
		return nil
	}
	return strings.Split(out, "\n")
}

func TestTagCmd_Rename(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	seedListEntry(t, dbPath, "e1", "auth", "", "")
	seedListEntry(t, dbPath, "e2", "auth", "", "")

	_, errOut, err := runTagSubCmd(t, dbPath, "rename", "auth", "authz")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(errOut, "Renamed.") {
		t.Errorf("expected 'Renamed.' in stderr, got %q", errOut)
	}

	if rows := runListForTag(t, dbPath, "authz"); len(rows) != 2 {
		t.Errorf("list --tag authz: got %d rows, want 2", len(rows))
	}
	if rows := runListForTag(t, dbPath, "auth"); len(rows) != 0 {
		t.Errorf("list --tag auth: got %d rows, want 0", len(rows))
	}
}

func TestTagCmd_RenameIntoExistingErrors(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	seedListEntry(t, dbPath, "e1", "auth", "", "")
	seedListEntry(t, dbPath, "e2", "perf", "", "")

	_, _, err := runTagSubCmd(t, dbPath, "rename", "auth", "perf")
	if !errors.Is(err, ErrUser) {
		t.Fatalf("expected ErrUser, got %v", err)
	}
	// Error message must mention "merge".
	if err != nil && !strings.Contains(err.Error(), "merge") {
		t.Errorf("error message should mention 'merge', got %q", err.Error())
	}
}

func TestTagCmd_RenameMissingErrors(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")

	_, _, err := runTagSubCmd(t, dbPath, "rename", "nope", "x")
	if !errors.Is(err, ErrUser) {
		t.Fatalf("expected ErrUser for missing tag, got %v", err)
	}
	if err != nil && !strings.Contains(err.Error(), "nope") {
		t.Errorf("error message should name the missing tag, got %q", err.Error())
	}
}

func TestTagCmd_RenameSameNameErrors(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	_, _, err := runTagSubCmd(t, dbPath, "rename", "auth", "auth")
	if !errors.Is(err, ErrUser) {
		t.Fatalf("expected ErrUser for same name, got %v", err)
	}
}

func TestTagCmd_RenameArgCountErrors(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	_, _, err := runTagSubCmd(t, dbPath, "rename", "auth")
	if !errors.Is(err, ErrUser) {
		t.Fatalf("expected ErrUser for too few args, got %v", err)
	}
}

func TestTagCmd_Merge(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	seedListEntry(t, dbPath, "e1", "auth,perf", "", "")
	seedListEntry(t, dbPath, "e3", "auth", "", "")

	_, errOut, err := runTagSubCmd(t, dbPath, "merge", "auth", "perf")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(errOut, "Merged.") {
		t.Errorf("expected 'Merged.' in stderr, got %q", errOut)
	}

	if rows := runListForTag(t, dbPath, "perf"); len(rows) != 2 {
		t.Errorf("list --tag perf: got %d rows, want 2", len(rows))
	}
	if rows := runListForTag(t, dbPath, "auth"); len(rows) != 0 {
		t.Errorf("list --tag auth: got %d rows, want 0", len(rows))
	}
}

func TestTagCmd_MergeDeDups(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	// e1 is tagged both auth and perf; after merge auth→perf, e1 counted once.
	seedListEntry(t, dbPath, "e1", "auth,perf", "", "")
	seedListEntry(t, dbPath, "e2", "auth", "", "")

	_, _, err := runTagSubCmd(t, dbPath, "merge", "auth", "perf")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// `brag tags` shows perf with de-duped count (e1 once + e2 once = 2).
	root, outBuf, _ := newTagTestRoot(t)
	root.SetArgs([]string{"--db", dbPath, "tags"})
	if err := root.Execute(); err != nil {
		t.Fatalf("brag tags: %v", err)
	}
	tagsOut := outBuf.String()
	if !strings.Contains(tagsOut, "perf\t2") {
		t.Errorf("expected 'perf\\t2' in tags output, got %q", tagsOut)
	}
	if strings.Contains(tagsOut, "auth") {
		t.Errorf("expected 'auth' to be gone from tags output, got %q", tagsOut)
	}
}

func TestTagCmd_MergeMissingErrors(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	seedListEntry(t, dbPath, "e", "perf", "", "")

	// Missing src.
	_, _, err := runTagSubCmd(t, dbPath, "merge", "nope", "perf")
	if !errors.Is(err, ErrUser) {
		t.Fatalf("missing src: expected ErrUser, got %v", err)
	}

	// Missing dst — message should mention "rename".
	_, _, err2 := runTagSubCmd(t, dbPath, "merge", "perf", "nope")
	if !errors.Is(err2, ErrUser) {
		t.Fatalf("missing dst: expected ErrUser, got %v", err2)
	}
	if err2 != nil && !strings.Contains(err2.Error(), "rename") {
		t.Errorf("missing dst error should mention 'rename', got %q", err2.Error())
	}
}

func TestTagCmd_MergeSameNameErrors(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	_, _, err := runTagSubCmd(t, dbPath, "merge", "auth", "auth")
	if !errors.Is(err, ErrUser) {
		t.Fatalf("expected ErrUser for same name, got %v", err)
	}
}

func TestTagCmd_MergeArgCountErrors(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	_, _, err := runTagSubCmd(t, dbPath, "merge", "auth")
	if !errors.Is(err, ErrUser) {
		t.Fatalf("expected ErrUser for wrong arg count, got %v", err)
	}
}
