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
//
// The fractional seconds are load-bearing, not decoration (STAGE-018 audit
// nit). At the original second resolution, two migrating opens inside the same
// second produced the SAME filename — and because VACUUM INTO refuses an
// existing destination and DEC-021 turns a failed backup into an aborted open,
// the observable symptom was "the second process cannot open the database at
// all". Not an overwrite, but worse than it sounds, and reachable whenever
// anything opens brag twice in quick succession (a hook plus a script, a test
// harness, two shells).
//
// Deliberately fixed by widening the timestamp rather than by probing for a
// free name and adding a `-2` suffix on collision. Dodging would work around a
// state DEC-021 exists to STOP on, and would produce a filename that DEC-021's
// own recovery instructions — which quote the timestamped sidecar name — do not
// describe. Widening keeps the abort semantics exactly as decided and makes the
// collision practically unreachable instead.
const backupTimeFormat = "20060102T150405.000000000Z"

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

	// VACUUM INTO requires the destination not already exist, and DEC-021 makes
	// that refusal ABORT the migration rather than proceed un-backed-up. The
	// timestamp is what keeps the name unique; see backupTimeFormat for why it
	// carries sub-second precision. Bind the path as a parameter (verified
	// supported on modernc.org/sqlite v1.51.0) so a path containing a quote can
	// never break the statement.
	if _, err := db.ExecContext(ctx, `VACUUM INTO ?`, dest); err != nil {
		return fmt.Errorf("backup before migrations: vacuum into %s: %w", dest, err)
	}
	return nil
}
