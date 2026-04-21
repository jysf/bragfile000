package editor

import (
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// EditFunc opens the file at path in an editor and returns once the
// user has saved (or aborted). The production implementation is
// Default; tests inject fakes so no real editor is spawned.
type EditFunc func(path string) error

// Default resolves $EDITOR → $VISUAL → vi and execs it against the
// given path, wired to the current stdio so interactive editors work.
var Default EditFunc = func(path string) error {
	argv := append(resolveEditor(), path)
	c := exec.Command(argv[0], argv[1:]...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		return fmt.Errorf("editor exited: %w", err)
	}
	return nil
}

// Launch writes initial to a temp .md file, invokes edit on it, and
// returns the resulting bytes plus whether the content actually
// changed (SHA-256 of the raw bytes — stricter than semantic-equal but
// simpler). The temp file is removed on return.
func Launch(initial []byte, edit EditFunc) ([]byte, bool, error) {
	f, err := os.CreateTemp("", "brag-edit-*.md")
	if err != nil {
		return nil, false, fmt.Errorf("create temp: %w", err)
	}
	path := f.Name()
	defer os.Remove(path)
	if _, err := f.Write(initial); err != nil {
		_ = f.Close()
		return nil, false, fmt.Errorf("write temp: %w", err)
	}
	if err := f.Close(); err != nil {
		return nil, false, fmt.Errorf("close temp: %w", err)
	}

	initialHash := sha256.Sum256(initial)
	if err := edit(path); err != nil {
		return nil, false, fmt.Errorf("edit: %w", err)
	}
	edited, err := os.ReadFile(path)
	if err != nil {
		return nil, false, fmt.Errorf("read edited temp: %w", err)
	}
	editedHash := sha256.Sum256(edited)
	return edited, initialHash != editedHash, nil
}

// resolveEditor returns the editor command (with any arguments split
// on whitespace) to invoke. Precedence matches the spec: $EDITOR
// wins, $VISUAL is the fallback, vi is the final default.
func resolveEditor() []string {
	for _, env := range []string{"EDITOR", "VISUAL"} {
		if v := strings.TrimSpace(os.Getenv(env)); v != "" {
			return strings.Fields(v)
		}
	}
	return []string{"vi"}
}
