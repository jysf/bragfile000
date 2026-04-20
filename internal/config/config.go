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
