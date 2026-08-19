// Package config resolves the bragfile database path. ResolveDBPath applies
// DEC-003's fixed order — an explicit flag value, then the BRAGFILE_DB env
// var, then DefaultDBPath's ~/.bragfile/db.sqlite — expanding a leading
// `~/` and returning an absolute path. internal/cli's commands call it to
// resolve the --db flag; internal/storage's dev/prod migration guard
// (DEC-026) calls it independently to find the real database regardless of
// what path the caller passed. It has no dependency on internal/storage, so
// both can import it without a cycle.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DefaultDBPath returns the default database path: ~/.bragfile/db.sqlite,
// expanded to an absolute path.
func DefaultDBPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve db path: %w", err)
	}
	return filepath.Join(home, ".bragfile", "db.sqlite"), nil
}

// ResolveDBPath resolves the database path using the order defined in DEC-003:
// flag value (if non-empty) → BRAGFILE_DB env var (if set) → default.
func ResolveDBPath(flagValue string) (string, error) {
	path := flagValue

	if path == "" {
		path = os.Getenv("BRAGFILE_DB")
	}

	if path == "" {
		return DefaultDBPath()
	}

	// Expand tilde to home directory
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve db path: %w", err)
		}
		path = filepath.Join(home, path[2:])
	}

	path, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve db path: %w", err)
	}

	return path, nil
}
