package storage

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"time"
)

// backupTimeFormat is a filename-safe, compact UTC timestamp for the
// pre-migration sidecar. RFC3339 is unusable in a filename because its
// colons are illegal on some filesystems (and awkward on all), so the
// sidecar uses this colon-free form. Stored timestamps remain RFC3339
// elsewhere (timestamps-in-utc-rfc3339); this format is for filenames only.
const backupTimeFormat = "20060102T150405Z"

// clock returns the current time used to stamp the backup filename. It is
// a package-level var so tests can freeze it and assert on the exact
// sidecar name — the same injectable-seam pattern as add.go's addGetCwd
// and project.go's getCwd (SPEC-031/032). Production never reassigns it.
var clock = func() time.Time { return time.Now().UTC() }

// backupBeforeMigrations snapshots the DB at path to a timestamped sidecar
// before any pending migration runs — but ONLY for an existing DB (>=1
// migration already applied) that has work pending. Two no-op cases:
//
//   - applied == 0: a brand-new DB. Everything is "pending" only because
//     the file was just created; there is nothing to lose. No backup.
//   - pending == 0: the DB is already at head. Nothing is about to change.
//     No backup.
//
// Only applied>0 AND pending>0 — an established DB about to be mutated by a
// forward-only, irreversible migration (DEC-002) — earns a backup.
//
// The snapshot goes through the open *sql.DB via VACUUM INTO, which writes
// a single-file, transaction-consistent copy with no external tooling. The
// build is CGO-off pure Go (DEC-001), so neither the sqlite3 CLI nor a
// WAL-unsafe file copy is available or correct; the driver is the only
// correct path. If the snapshot fails, the caller (Open) aborts rather than
// migrate an un-backed-up DB (DEC-021).
func backupBeforeMigrations(ctx context.Context, db *sql.DB, path string, src fs.FS) error {
	applied, pending, err := migrationStatus(ctx, db, src)
	if err != nil {
		return fmt.Errorf("backup before migrations: %w", err)
	}
	if len(applied) == 0 || len(pending) == 0 {
		return nil // brand-new DB, or already at head: nothing to back up.
	}

	highest := pending[len(pending)-1] // pending is in lexical apply order.
	dest := fmt.Sprintf("%s.pre-%s.%s.backup", path, highest, clock().Format(backupTimeFormat))

	// VACUUM INTO requires the destination not already exist; the
	// timestamped name guarantees that in practice. Bind the path as a
	// parameter (verified supported on modernc.org/sqlite v1.51.0) so a
	// path containing a quote can never break the statement.
	if _, err := db.ExecContext(ctx, `VACUUM INTO ?`, dest); err != nil {
		return fmt.Errorf("backup before migrations: vacuum into %s: %w", dest, err)
	}
	return nil
}
