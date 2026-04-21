// Package storagetest exposes test-only helpers that need raw SQL
// access to a Bragfile database. Living under internal/storage/ keeps
// the database/sql dependency inside the storage layer, which lets CLI
// tests use these helpers without violating the no-sql-in-cli-layer
// constraint.
package storagetest

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// Backdate rewrites entries.created_at on the row with the given id to
// at (formatted as UTC RFC3339). Store.Add always stamps time.Now(), so
// tests covering --since filters use this helper to seed past-dated
// rows out-of-band.
func Backdate(dbPath string, id int64, at time.Time) error {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("storagetest.Backdate: open %s: %w", dbPath, err)
	}
	defer db.Close()
	ts := at.UTC().Format(time.RFC3339)
	if _, err := db.Exec("UPDATE entries SET created_at = ? WHERE id = ?", ts, id); err != nil {
		return fmt.Errorf("storagetest.Backdate: id=%d: %w", id, err)
	}
	return nil
}
