package mcpserver

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNoSQLImport enforces no-sql-in-cli-layer's architecture line for this
// package by convention: internal/mcpserver, like internal/cli, must not
// import database/sql directly — all persistence goes through
// *storage.Store. The constraint's path glob covers internal/cli/** only;
// this test covers the gap for the new package.
func TestNoSQLImport(t *testing.T) {
	err := filepath.WalkDir(".", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || filepath.Base(path) == "import_audit_test.go" {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if strings.Contains(string(b), `"database/sql"`) {
			t.Errorf("%s imports database/sql; SQL must stay in internal/storage", path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
