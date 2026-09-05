package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/jysf/bragfile000/internal/storage"
)

// newRootWithLearn mirrors newRootWithAdd: milestone TTY seam pinned off by
// default, learn + add + list registered so the milestone PAIR (below) can
// exercise both verbs against one corpus.
func newRootWithLearn(t *testing.T) (*cobra.Command, string) {
	t.Helper()
	addStderrIsTTY = func() bool { return false }
	t.Cleanup(func() { addStderrIsTTY = defaultStderrIsTTY })
	root := NewRootCmd("test")
	root.AddCommand(NewLearnCmd())
	root.AddCommand(NewAddCmd())
	root.AddCommand(NewListCmd())
	dbPath := filepath.Join(t.TempDir(), "test.db")
	return root, dbPath
}

// TestLearnCmd_PinsFailedType is the core claim: the verb writes the reserved
// value, and `brag list --type failed` gets it back (STAGE-023 criterion 2,
// satisfied by an existing flag).
func TestLearnCmd_PinsFailedType(t *testing.T) {
	root, dbPath := newRootWithLearn(t)
	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"--db", dbPath, "learn", "--title", "shared-worker pool did not cut cold starts"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	id, err := strconv.ParseInt(strings.TrimSpace(outBuf.String()), 10, 64)
	if err != nil {
		t.Fatalf("stdout should be the ID alone, got %q", outBuf.String())
	}
	s, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("storage.Open: %v", err)
	}
	defer s.Close()
	got, err := s.Get(id)
	if err != nil {
		t.Fatalf("Get(%d): %v", id, err)
	}
	if got.Type != FailureType {
		t.Errorf("Type = %q, want %q", got.Type, FailureType)
	}
	if FailureType != "failed" {
		t.Errorf("FailureType = %q, want %q", FailureType, "failed")
	}
}

// TestLearnCmd_NoTypeFlag: the value is pinned, so the flag must not exist.
// A structural check on the flag set, not a substring check on help text —
// the Long deliberately DOES contain "--type" (it teaches `brag list --type
// failed`), so a NOT-contains here would be wrong.
func TestLearnCmd_NoTypeFlag(t *testing.T) {
	cmd := NewLearnCmd()
	if f := cmd.Flags().Lookup("type"); f != nil {
		t.Errorf("brag learn must not define --type, got %v", f)
	}
	if f := cmd.Flags().ShorthandLookup("k"); f != nil {
		t.Errorf("brag learn must not define -k (add's type shorthand), got %v", f)
	}
	for _, name := range []string{"title", "description", "tags", "project", "impact"} {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("brag learn must define --%s", name)
		}
	}
}

// TestLearnCmd_MilestoneSuppressed_AddStillFires is the PAIR. The negative
// alone proves nothing: stderr is also empty when no milestone would have
// crossed. So the same corpus, the same forced-TTY, the same threshold is
// driven through BOTH verbs — add fires, learn is silent.
func TestLearnCmd_MilestoneSuppressed_AddStillFires(t *testing.T) {
	// negative: learn is the 10th entry, TTY on, nothing on stderr
	root, dbPath := newRootWithLearn(t)
	seedEntries(t, dbPath, 9, "")
	setStderrIsTTY(t, true)
	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"--db", dbPath, "learn", "--title", "tenth, and it did not work"})
	if err := root.Execute(); err != nil {
		t.Fatalf("learn: unexpected error: %v", err)
	}
	if errBuf.Len() != 0 {
		t.Errorf("brag learn must not congratulate; stderr = %q", errBuf.String())
	}

	// positive: same threshold, same TTY, via add — the milestone DOES fire,
	// which is what makes the silence above evidence of suppression.
	root2, dbPath2 := newRootWithLearn(t)
	seedEntries(t, dbPath2, 9, "")
	setStderrIsTTY(t, true)
	var outBuf2, errBuf2 bytes.Buffer
	root2.SetOut(&outBuf2)
	root2.SetErr(&errBuf2)
	root2.SetArgs([]string{"--db", dbPath2, "add", "--title", "tenth"})
	if err := root2.Execute(); err != nil {
		t.Fatalf("add: unexpected error: %v", err)
	}
	if !strings.Contains(errBuf2.String(), "🎉 10 brags and counting") {
		t.Fatalf("control failed: add should still fire the milestone, got %q", errBuf2.String())
	}
}

// TestLearnCmd_EditorModeOverwritesUserType: the template omits Type, but a
// user who re-adds the header does not get to redirect the pinned value.
func TestLearnCmd_EditorModeOverwritesUserType(t *testing.T) {
	root, dbPath := newRootWithLearn(t)
	installAddEditFunc(t, func(path string) error {
		return os.WriteFile(path, []byte("Title: bloom filter on the tag join\nType: shipped\n\ntried it; the join was never the bottleneck\n"), 0o600)
	})
	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"--db", dbPath, "learn"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	id, err := strconv.ParseInt(strings.TrimSpace(outBuf.String()), 10, 64)
	if err != nil {
		t.Fatalf("stdout should be the ID alone, got %q", outBuf.String())
	}
	s, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("storage.Open: %v", err)
	}
	defer s.Close()
	got, err := s.Get(id)
	if err != nil {
		t.Fatalf("Get(%d): %v", id, err)
	}
	if got.Type != FailureType {
		t.Errorf("editor-mode Type = %q, want %q (user's header must be overwritten)", got.Type, FailureType)
	}
}

// TestLearnCmd_EmptyTitleIsUserError: flag mode requires --title, same as add.
func TestLearnCmd_EmptyTitleIsUserError(t *testing.T) {
	root, dbPath := newRootWithLearn(t)
	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"--db", dbPath, "learn", "--impact", "cost two days"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected a user error, got nil")
	}
	if !strings.Contains(err.Error(), "--title is required") {
		t.Errorf("error = %v, want it to mention --title is required", err)
	}
}
