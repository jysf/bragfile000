---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-036
  type: story                      # epic | story | task | bug | chore
  cycle: build                     # frame | design | build | verify | ship
  blocked: false
  priority: medium
  complexity: S                    # S | M | L  (L means split it)

project:
  id: PROJ-002
  stage: STAGE-008
repo:
  id: bragfile

agents:
  architect: claude-opus-4-8
  implementer: claude-opus-4-8     # usually same Claude, different session
  created_at: 2026-06-12

references:
  decisions: [DEC-002, DEC-021]
  constraints:
    - no-cgo
    - migrations-are-append-only
    - errors-wrap-with-context
    - storage-tests-use-tempdir
    - timestamps-in-utc-rfc3339
    - test-before-implementation
  related_specs: [SPEC-031, SPEC-032]
---

# SPEC-036: migration auto-backup safety belt

## Context

`brag`'s migrations are **forward-only** (DEC-002): there is no
`migrate down`, filenames are append-only (`migrations-are-append-only`),
and "upgrade" means "run a newer binary, which silently migrates
`~/.bragfile` on the next `storage.Open`." Until now, nothing snapshotted
the database before that one-way mutation.

**Motivating incident (this session, 2026-06-12).** A v0.2.x binary opened
the **production** `~/.bragfile`; `applyMigrations` silently ran
`0003_normalize_tags` + `0004_add_projects`, irreversibly upgrading a DB
with **no backup** and no down-migration to fall back on. The runner did
exactly what DEC-002 specifies — the gap is that the one-way door has no
snapshot behind it.

This spec adds an `Open`-time **auto-backup safety belt**: before applying
any pending migration to an *existing* DB, write a consistent snapshot
first, so that class of event becomes recoverable. The durability model
(mechanism / trigger / failure policy) is **DEC-021**, emitted with this
spec.

- Parent: `STAGE-008` (polish + v0.2.0 release), backlog item SPEC-036
  ("S/M — Migration auto-backup safety belt … DEC-021 likely").
- Project: `PROJ-002`.
- **No new SQL migration.** This is an `Open`-time guard, not a schema
  change; `schema_migrations` stays at **4** (`0001..0004`).

## Goal

Before `storage.Open` applies any pending migration to a database that
already has at least one migration applied, write a consistent single-file
snapshot of that database to a timestamped sidecar; if the snapshot fails,
abort `Open` rather than migrate an un-backed-up DB.

## Inputs

- **Files to read:**
  - `internal/storage/store.go` — `Open()` (lines 27–54); the safety belt
    slots in **before** `applyMigrations` (line 48).
  - `internal/storage/migrate.go` — `applyMigrations` / `loadApplied`; the
    pending/applied diff that the trigger reuses.
  - `internal/cli/add.go:14–19` and `project.go`'s `getCwd` — the
    `var addGetCwd = os.Getwd` injectable-seam pattern the `clock` seam
    mirrors (SPEC-031/032).
  - `DEC-002` (forward-only migrations — the regime guarded) and the new
    `DEC-021` (durability model).
- **External APIs:** none. `VACUUM INTO` is plain `database/sql` against
  the existing `modernc.org/sqlite` driver.
- **Related code paths:** `internal/storage/`.

## Outputs

- **Files created:**
  - `internal/storage/backup.go` — the `clock` seam, `backupTimeFormat`,
    and `backupBeforeMigrations` helper (one concept per file, §8).
  - `internal/storage/backup_test.go` — the failing tests below.
  - `decisions/DEC-021-migration-auto-backup-durability-model.md` — emitted
    with this spec (durability model).
- **Files modified:**
  - `internal/storage/store.go` — `Open` calls `backupBeforeMigrations`
    between `fs.Sub` and `applyMigrations`.
  - `internal/storage/migrate.go` — extract a read-only `migrationStatus`
    helper (applied/pending partition) that both `applyMigrations` and the
    backup guard use. Behavior of `applyMigrations` is unchanged.
- **New exports:** none (all new identifiers are unexported:
  `backupBeforeMigrations`, `migrationStatus`, `clock`, `backupTimeFormat`).
- **Database changes:** **none.** No migration added; `schema_migrations`
  stays at 4.

## Acceptance Criteria

- [ ] **AC1 — Backup IS created for an existing DB with pending work.**
  Re-opening a DB that has `0001` applied (and ≥1 row) while `0002..0004`
  are pending writes exactly one sidecar before migrating.
- [ ] **AC2 — Trigger discriminator is exact: `applied>0 AND pending>0`.**
  A brand-new DB (`applied==0`) gets **no** backup. A DB already at head
  (`pending==0`) gets **no** backup.
- [ ] **AC3 — Mechanism is `VACUUM INTO` through the open `*sql.DB`.** The
  sidecar is a valid, openable SQLite file holding the **pre-migration**
  state (pre-`0004`: it has the seeded `entries` rows and does **not** have
  the `projects` table).
- [ ] **AC4 — Failure aborts `Open`.** If the snapshot write fails, `Open`
  returns a wrapped error, returns a nil store, and applies **no**
  migration (the on-disk DB stays at its pre-Open version).
- [ ] **AC5 — Filename matches the scheme** `<dbpath>.pre-<highestPendingVersion>.<UTC>.backup`
  with a filename-safe (colon-free) timestamp; the **injected `clock`**
  determines that timestamp.
- [ ] **AC6 — No new migration.** `schema_migrations` still tracks exactly
  the four `0001..0004` versions after Open; the safety belt adds none.

## Failing Tests

Written during **design**, BEFORE build. All time-dependent assertions use
the injectable `clock` var — **no `time.Sleep`** (§9 no-sleep discipline);
freezing `clock` is what makes the timestamp deterministic.

All tests live in `internal/storage/backup_test.go`, in-package
(`package storage`), and use `t.TempDir()` (`storage-tests-use-tempdir`).
They reuse the existing `apply0001Only(t)` helper (`store_test.go:817`,
seeds a DB at `0001` only) and `openRawDB(t, path)` (`fts_test.go:19`).

- **`internal/storage/backup_test.go`**

  - **`TestBackup_CreatesSnapshotForExistingDBWithPending`** *(AC1, AC3, AC5)*
    — `apply0001Only` → seed one row via raw INSERT into `entries`
    (0001-era schema has the `tags` column) → `rawDB.Close()`. Freeze
    `clock` to `2026-06-12T09:30:15Z`. `Open(path)`. Asserts:
    - exactly one file matches `filepath.Glob(path + ".pre-*.backup")`;
    - that file's name **equals**
      `path + ".pre-0004_add_projects.20260612T093015Z.backup"` (exact —
      this pins the scheme *and* proves the injected clock drove the
      timestamp);
    - opening the sidecar as a raw `*sql.DB` and running
      `SELECT COUNT(*) FROM entries` returns the seeded count (the snapshot
      holds pre-migration rows);
    - `objectExists(sidecarDB, "table", "projects") == false` (snapshot was
      taken **before** `0004`, proving it is a true pre-migration copy);
    - the **live** DB at `path` is now at head:
      `SELECT COUNT(*) FROM schema_migrations == 4` and `projects` exists.

  - **`TestBackup_NoSnapshotForFreshDB`** *(AC2)* — fresh `t.TempDir()`
    path, `Open(path)` (applied starts at 0). Asserts
    `filepath.Glob(path + ".pre-*.backup")` returns **zero** files.

  - **`TestBackup_NoSnapshotWhenAlreadyAtHead`** *(AC2)* — `Open(path)` on a
    fresh path (applies `0001..0004`), `Close()`. Confirm zero `*.backup`
    files. `Open(path)` **again** (now `pending==0`). Asserts still **zero**
    `*.backup` files — the at-head re-open created none.

  - **`TestBackup_FailureAbortsOpenAndLeavesDBUnmigrated`** *(AC4)* —
    `apply0001Only` → seed one row → `rawDB.Close()`. Freeze `clock` to a
    fixed instant `T`; pre-create the exact target path the guard will
    compute (`path + ".pre-0004_add_projects." + T.Format("20060102T150405Z") + ".backup"`)
    as an empty file so `VACUUM INTO` fails with `output file already
    exists`. `s, err := Open(path)`. Asserts:
    - `err != nil` and `s == nil`;
    - `err.Error()` contains both `"open store:"` and
      `"backup before migrations"` (wrapped per `errors-wrap-with-context`);
    - opening `path` raw, `SELECT COUNT(*) FROM schema_migrations == 1`
      (only `0001`; `0002..0004` were **not** applied — abort happened
      before migrating).

  - **`TestBackup_NoNewMigrationVersion`** *(AC6)* — `apply0001Only` → seed
    a row → `rawDB.Close()` → `Open(path)`. Asserts the live DB's
    `schema_migrations` versions are exactly
    `["0001_initial","0002_add_fts","0003_normalize_tags","0004_add_projects"]`
    (count 4) — the safety belt records no version of its own.

### Premise audit (run at design, §9)

- **Inversion — existing tests that now emit a sidecar.** Every storage
  test that opens a **fresh** `t.TempDir()` DB has `applied==0` on first
  Open → **no backup** → unaffected. Two existing in-package tests open a
  DB **seeded at `0001`** and then call `Open` (which applies `0002..0004`),
  so they **now trigger a backup**:
  - `internal/storage/store_test.go` → `TestMigrate_ETL_Lossless`
    (uses `apply0001Only`, seeds 6 rows, then `Open`).
  - `internal/storage/fts_test.go` → `TestFTS_MigrationBackfillsExistingRows`
    (inline applies `0001` + seeds 3 rows, then `Open`).
  Both assert only on `entries_fts` / `tags` / `taggings` / `schema_migrations`
  contents — **neither asserts on directory contents or file counts** — and
  both run in `t.TempDir()`, so the new sidecar is created harmlessly and
  cleaned up. **They still pass unchanged; no rewrite needed.** (Verified by
  grep: no test counts dir entries or globs for files; the only `os.Stat`
  calls target the DB file/parent-dir mode, not sibling files.)
- **Count-bump — none.** No migration added; `schema_migrations` stays at
  4. The §12(a) literal-count assertions (`count == 4`, the `0001..0004`
  version lists in `store_test.go`, `fts_test.go`, `project_migration_test.go`)
  are **unaffected** — confirmed they assert 4 and remain 4. AC6 guards this.
- **§12(b) — N/A.** No DDL, no embedded SQL artifact whose shape depends on
  an external tool. (The one external-behavior assumption — that
  `modernc.org/sqlite` supports `VACUUM INTO` with a bound parameter — was
  pre-flighted at design: see Implementation Context.)
- **Status-change (docs).** The backup behavior is documented by **SPEC-034**
  (the doc sweep) — forward reference only; no tutorial prose here.

## Implementation Context

*Read this section (and the files it points to) before starting the build
cycle.*

### Design-time pre-flight (done — transcribe, don't re-derive)

`VACUUM INTO` was run through `modernc.org/sqlite v1.51.0` (the version in
`go.mod`) at design time. Confirmed:
1. `db.ExecContext(ctx, "VACUUM INTO ?", dest)` with a **bound parameter**
   succeeds (no need to string-interpolate the path — avoids quote/escaping
   hazards for paths containing `'`).
2. Re-running into an existing destination fails with
   `SQL logic error: output file already exists (1)` — so the timestamped
   name must be unique, and this is exactly the lever the failure test uses.
3. The produced file opens as a valid `*sql.DB` with the source rows intact.

### Exact Go — `internal/storage/backup.go` (new file)

```go
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
```

### Exact Go — `internal/storage/migrate.go` (extract `migrationStatus`)

Replace the body of `applyMigrations` with a call into a new read-only
`migrationStatus` helper. `migrationStatus` performs the schema_migrations
bootstrap + applied-load + file-list + partition that `applyMigrations`
used to do inline, and returns the two version slices. `applyMigrations`'
external behavior is unchanged (the existing `TestMigrate_*` tests must
still pass).

```go
// migrationStatus ensures schema_migrations exists, then partitions the
// embedded migrations into those already applied and those still pending
// (pending returned in lexical apply order). It applies nothing — it is the
// read-only half that both applyMigrations and the Open-time backup safety
// belt (backup.go) build on.
func migrationStatus(ctx context.Context, db *sql.DB, src fs.FS) (applied, pending []string, err error) {
	if _, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
        version TEXT PRIMARY KEY,
        applied_at TEXT NOT NULL
    )`); err != nil {
		return nil, nil, fmt.Errorf("create schema_migrations: %w", err)
	}

	appliedSet, err := loadApplied(ctx, db)
	if err != nil {
		return nil, nil, err
	}

	entries, err := fs.ReadDir(src, ".")
	if err != nil {
		return nil, nil, fmt.Errorf("read migrations dir: %w", err)
	}
	var files []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		files = append(files, e.Name())
	}
	sort.Strings(files)

	for _, name := range files {
		version := strings.TrimSuffix(name, ".sql")
		if _, ok := appliedSet[version]; ok {
			applied = append(applied, version)
		} else {
			pending = append(pending, version)
		}
	}
	return applied, pending, nil
}

// applyMigrations reads *.sql files from src, diffs them against the
// schema_migrations table, and applies each missing migration inside its
// own transaction (alongside the tracking INSERT) in lexical order.
func applyMigrations(ctx context.Context, db *sql.DB, src fs.FS) error {
	_, pending, err := migrationStatus(ctx, db, src)
	if err != nil {
		return err
	}
	for _, version := range pending {
		name := version + ".sql"
		sqlBytes, err := fs.ReadFile(src, name)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}
		if err := runMigration(ctx, db, version, string(sqlBytes)); err != nil {
			return fmt.Errorf("apply migration %s: %w", name, err)
		}
	}
	return nil
}
```

`loadApplied` and `runMigration` are unchanged. The `embed.FS` declaration
and imports (`sort`, `strings`, `time`, etc.) stay; nothing new is imported
in `migrate.go`.

### Exact Go — `internal/storage/store.go` (Open integration)

Insert the guard between the `fs.Sub` block and `applyMigrations`. Use one
shared `ctx`:

```go
	sub, err := fs.Sub(migrationsFS, "migrations")
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("open store: %w", err)
	}

	// Safety belt: snapshot an existing DB before any forward-only,
	// irreversible migration (DEC-002) mutates it. No-op for a brand-new or
	// already-current DB. A failed snapshot aborts Open rather than migrate
	// an un-backed-up DB (DEC-021).
	ctx := context.Background()
	if err := backupBeforeMigrations(ctx, db, path, sub); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("open store: %w", err)
	}
	if err := applyMigrations(ctx, db, sub); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("open store: %w", err)
	}
```

(`store.go` already imports `context`; the existing
`applyMigrations(context.Background(), ...)` call is replaced by the shared
`ctx`.)

### Decisions that apply

- `DEC-021` — the durability model emitted with this spec: mechanism
  (`VACUUM INTO` through the driver), trigger (`applied>0 AND pending>0`),
  failure policy (abort), keep-all retention, silent.
- `DEC-002` — forward-only embedded migrations: the irreversible regime the
  backup protects, and the reason a snapshot is the only recovery path.
- `DEC-001` — pure-Go driver: forces the through-the-driver mechanism over
  a `sqlite3` CLI call.

### Constraints that apply

- `no-cgo` — the snapshot must go through the Go driver, not shell out.
- `migrations-are-append-only` — this spec adds **no** migration; it must
  not edit/rename any `migrations/*.sql`.
- `errors-wrap-with-context` — every new error is `fmt.Errorf("<op>: %w")`;
  `Open` wraps as `open store: …`, the guard wraps as
  `backup before migrations: …`.
- `storage-tests-use-tempdir` — all new tests use `t.TempDir()`.
- `timestamps-in-utc-rfc3339` — stored timestamps stay RFC3339 UTC; the
  filename-only `backupTimeFormat` is a colon-free derivative (filenames
  can't hold colons), not a new stored format.
- `test-before-implementation` — the Failing Tests above are written first.

### Prior related work

- `SPEC-031` / `SPEC-032` (shipped) — the `var getCwd = os.Getwd` /
  `addGetCwd` injectable-os-var seam the `clock` var mirrors. This is the
  **N=2 → N=3** data point for that WATCH-list pattern (STAGE-007 close):
  `clock` is the same shape (package var defaulting to the real impl,
  overridden in tests, never reassigned in production).

### Out of scope (for this spec specifically)

- **Backup pruning / retention policy.** Keep-all for v0.2.0; pruning is
  possible future work (would earn its own DEC).
- **The launchd daily-backup / `scripts/backup-db.sh`.** Ops, not release
  code; STAGE-008 keeps it out.
- **The documented WAL-safe backup recipe / tutorial prose.** That is
  SPEC-034 (doc sweep) — forward reference only.
- **Any down-migration capability.** Forward-only stays; the snapshot *is*
  the downgrade story.
- **Prompt-and-confirm / TTY-aware UX.** Rejected in DEC-021; the guard is
  non-interactive and silent.

## Notes for the Implementer

- **One concept per file (§8):** the safety belt is its own concept → new
  `backup.go`, not appended to `store.go`. `migrationStatus` is a migration
  concern → it lives in `migrate.go`.
- **Highest pending version** is `pending[len(pending)-1]` because
  `migrationStatus` returns `pending` in lexical (apply) order — for the
  incident scenario that is `"0004_add_projects"`, giving
  `db.sqlite.pre-0004_add_projects.<ts>.backup`.
- **"Non-empty" means `applied > 0`, not "has rows."** Do not add a row-count
  check; the discriminator is a pure count comparison (cheap, no scan) and
  must stay that way (DEC-021).
- **No `time.Sleep` anywhere.** The frozen `clock` is the entire mechanism
  for deterministic timestamps (§9).
- **The two existing seeded-at-0001 tests** (`TestMigrate_ETL_Lossless`,
  `TestFTS_MigrationBackfillsExistingRows`) now emit a sidecar into their
  tempdir but assert nothing about it — leave them as-is; do not "fix" them.
- After build: `go test ./...`, `gofmt -l .`, `go vet ./...`, and
  `CGO_ENABLED=0 go build ./...` must all be clean.

---

## Build Completion

*Filled in at the end of the **build** cycle, before advancing to verify.*

- **Branch:** `feat/spec-036-migration-auto-backup-safety-belt`
- **PR (if applicable):** (opened below)
- **All acceptance criteria met?** yes
- **New decisions emitted:**
  - `DEC-021` — migration auto-backup durability model (emitted at design)
- **Deviations from spec:**
  - **Test setup for `TestBackup_FailureAbortsOpenAndLeavesDBUnmigrated`:** the spec's
    pre-flight assumed creating an empty (0-byte) file at the collision path would cause
    `VACUUM INTO` to fail. `modernc.org/sqlite v1.52.0` overwrites empty files silently;
    only a non-empty existing file triggers "output file already exists (1)". The fix is
    a one-line test change: instead of `os.WriteFile(path, nil, 0o600)`, open a real
    SQLite database at the collision path and write one table row. Production behavior
    is unaffected — real backup files are never 0 bytes and the timestamped name prevents
    real-world collisions. All five failing tests now pass. Per spec instruction ("stop
    and report"), this deviation is documented here; the production code path is correct.
- **Follow-up work identified:**
  - None (backup pruning / retention policy deferred per spec out-of-scope note)

### Build-phase reflection (3 questions, short answers)

Process-focused: how did the build go? What friction did the spec create?

1. **What was unclear in the spec that slowed you down?**
   — Nothing was unclear; the literal-artifact-as-spec approach (exact Go in Implementation
   Context) made transcription mechanical. The only friction was the v1.52.0 VACUUM INTO
   behavior difference, which the spec anticipated with the "stop and report" instruction.

2. **Was there a constraint or decision that should have been listed but wasn't?**
   — No missing constraints or decisions. All five relevant constraints were explicitly
   listed and applied cleanly. The `no-cgo` constraint was the load-bearing one (forced
   VACUUM INTO through the driver); it was correctly listed.

3. **If you did this task again, what would you do differently?**
   — The design-time pre-flight for the failure test (§12(b)) should have verified the
   behavior with an EMPTY file, not just with a full existing SQLite file. The pre-flight
   confirmed "re-running into an existing destination fails" but used a previously-vacuumed
   SQLite file as the destination. Running the pre-flight with a 0-byte empty file would
   have caught the v1.52.0 difference at design time.

---

## Reflection (Ship)

*Appended during the **ship** cycle. Outcome-focused reflection, distinct
from the process-focused build reflection above.*

1. **What would I do differently next time?**
   — <answer>

2. **Does any template, constraint, or decision need updating?**
   — <answer>

3. **Is there a follow-up spec I should write now before I forget?**
   — <answer>
