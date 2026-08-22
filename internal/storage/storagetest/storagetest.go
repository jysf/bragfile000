// Package storagetest exposes test-only helpers that need raw SQL
// access to a Bragfile database. Living under internal/storage/ keeps
// the database/sql dependency — and the schema knowledge that rides
// with it — inside the storage layer. A CLI test may open a database
// directly if it wants to (no-sql-in-cli-layer binds production files
// only, DEC-047); this package is the preferred route because it keeps
// that SQL in one place, not because the constraint forbids the other.
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
