package cli

import (
	"fmt"
	"os"
	"path/filepath"
)

// writeFileAtomic writes data to path via a temp file in the SAME directory
// followed by a rename, so a reader never observes a partial file and an
// interrupted write cannot truncate what was already there.
//
// STAGE-018 audit nit. `brag mcp install` rewrites a config file the user did
// not author and may not have backed up — a client's .mcp.json or
// claude_desktop_config.json, which can hold every other MCP server they have
// registered. A bare os.WriteFile truncates first and writes second, so a crash
// or a full disk between those two steps leaves the file empty. That is the one
// failure mode worth engineering against here: not losing OUR entry, but losing
// THEIRS.
//
// The temp file is created in filepath.Dir(path) rather than os.TempDir()
// because rename(2) is only atomic within a filesystem; a temp file on another
// device would fall back to a copy and reintroduce the partial-write window.
func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	f, err := os.CreateTemp(dir, ".brag-tmp-*")
	if err != nil {
		return fmt.Errorf("create temp file in %s: %w", dir, err)
	}
	tmp := f.Name()
	// Best-effort cleanup on every failure path below; a successful rename
	// makes this a no-op that returns an ignorable error.
	defer func() { _ = os.Remove(tmp) }()

	if _, err := f.Write(data); err != nil {
		f.Close()
		return fmt.Errorf("write temp file: %w", err)
	}
	// Sync before rename: rename only orders the directory entry, not the file
	// contents, so without this a crash can leave a renamed-but-empty file.
	if err := f.Sync(); err != nil {
		f.Close()
		return fmt.Errorf("sync temp file: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}
	// CreateTemp always makes 0600; restore the caller's intended mode before
	// the file becomes visible under its real name.
	if err := os.Chmod(tmp, perm); err != nil {
		return fmt.Errorf("chmod temp file: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("rename temp file into place: %w", err)
	}
	return nil
}
