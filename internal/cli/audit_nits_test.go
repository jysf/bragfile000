package cli

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestWriteFileAtomic_ReplacesContentAndLeavesNoTemp pins the happy path of the
// STAGE-018 atomic-write nit: the file ends up with exactly the new bytes, at
// the requested mode, and the temp file used to get there is gone.
func TestWriteFileAtomic_ReplacesContentAndLeavesNoTemp(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "config.json")
	if err := os.WriteFile(target, []byte(`{"old":true}`), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}

	want := []byte(`{"new":true,"servers":{"brag":{}}}`)
	if err := writeFileAtomic(target, want, 0o644); err != nil {
		t.Fatalf("writeFileAtomic: %v", err)
	}

	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if string(got) != string(want) {
		t.Errorf("content = %q, want %q", got, want)
	}
	fi, err := os.Stat(target)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if perm := fi.Mode().Perm(); perm != 0o644 {
		t.Errorf("mode = %o, want 644 — CreateTemp makes 0600, so the chmod "+
			"before rename is load-bearing", perm)
	}

	// No .brag-tmp-* residue in the directory.
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".brag-tmp-") {
			t.Errorf("temp file %q left behind", e.Name())
		}
	}
}

// TestWriteFileAtomic_FailureLeavesOriginalIntact is the point of the nit.
//
// `brag mcp install` rewrites a config file the user did not author — a
// client's .mcp.json can hold every other MCP server they have registered. The
// previous os.WriteFile truncates first and writes second, so a failure between
// those steps leaves an EMPTY file. Here the rename is made to fail (the target
// path is a directory, which cannot be replaced by a file), and the original
// content must survive untouched.
func TestWriteFileAtomic_FailureLeavesOriginalIntact(t *testing.T) {
	dir := t.TempDir()
	// A directory at the target path: os.Rename(file, dir) fails on every
	// supported platform, exercising the failure path without needing
	// permission games that behave differently under root/CI.
	target := filepath.Join(dir, "config.json")
	if err := os.Mkdir(target, 0o755); err != nil {
		t.Fatalf("mkdir target: %v", err)
	}
	sentinel := filepath.Join(target, "keepme")
	if err := os.WriteFile(sentinel, []byte("other servers"), 0o644); err != nil {
		t.Fatalf("seed sentinel: %v", err)
	}

	err := writeFileAtomic(target, []byte("new content"), 0o644)
	if err == nil {
		t.Fatal("want error when the target cannot be replaced, got nil")
	}

	// The pre-existing content must be untouched — nothing truncated, nothing
	// half-written.
	got, readErr := os.ReadFile(sentinel)
	if readErr != nil {
		t.Fatalf("original content disappeared after a failed write: %v", readErr)
	}
	if string(got) != "other servers" {
		t.Errorf("original content = %q, want it untouched", got)
	}
	// And no temp file survives the failure.
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".brag-tmp-") {
			t.Errorf("temp file %q left behind after failure", e.Name())
		}
	}
}

// TestOpenStoreErrorIsNotDoubleWrapped is the STAGE-018 regression for the
// doubled prefix.
//
// storage.Open already wraps its failures with "open store: …", and all 31 CLI
// call sites wrapped that AGAIN with the identical words, so every command's
// db-path failure read "brag: open store: open store: mkdir …". The fix was to
// drop the redundant CLI wrap; this pins that the prefix appears exactly once
// on a real command path.
func TestOpenStoreErrorIsNotDoubleWrapped(t *testing.T) {
	// A path whose parent cannot be created — storage.Open fails at mkdir.
	bad := filepath.Join(string(os.PathSeparator), "nonexistent-brag-root", "x", "db.sqlite")
	_, _, err := runListCmd(t, bad)
	if err == nil {
		t.Skip("expected an open failure for an unwritable root; environment allows it")
	}
	if n := strings.Count(err.Error(), "open store:"); n != 1 {
		t.Errorf("%q contains %d occurrences of \"open store:\", want exactly 1 — "+
			"storage.Open already supplies that context, so the CLI must not re-wrap", err, n)
	}
}

// TestFlagParseErrorIsAUserError pins that a malformed flag exits 1, not 2.
//
// main.go maps errors.Is(err, ErrUser) to exit 1 and everything else to exit 2,
// the code reserved for INTERNAL faults. Cobra's flag-parse errors reached it
// unwrapped, so `brag search -foo` — an entirely user-actionable mistake —
// reported the same exit status as a corrupt database. Fixed once at the root
// via SetFlagErrorFunc rather than per subcommand. (STAGE-018 audit nit.)
func TestFlagParseErrorIsAUserError(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	_, _, err := runListCmd(t, dbPath, "-foo")
	if err == nil {
		t.Fatal("want an error for an unknown shorthand flag, got nil")
	}
	if !errors.Is(err, ErrUser) {
		t.Errorf("a malformed flag must be ErrUser (exit 1), not an internal fault "+
			"(exit 2); got %v", err)
	}
	if !strings.Contains(err.Error(), "unknown shorthand flag") {
		t.Errorf("want cobra's own diagnostic preserved, got %q", err)
	}
}
