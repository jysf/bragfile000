---
# Maps to ContextCore insight.* semantic conventions.

insight:
  id: DEC-021
  type: decision
  confidence: 0.86
  audience:
    - developer
    - agent

agent:
  id: claude-opus-4-8
  session_id: null

# Decisions are repo-level, but it's useful to track which project
# caused them to be emitted.
project:
  id: PROJ-002
repo:
  id: bragfile

created_at: 2026-06-12
supersedes: null
superseded_by: null

tags:
  - storage
  - migrations
  - backup
  - durability
  - safety
---

# DEC-021: migration auto-backup safety belt — durability model

## Decision

Before `storage.Open` applies any pending migration to an **existing**
database, it writes a consistent, single-file snapshot of that database
to a timestamped sidecar next to the DB file. The model has three fixed
parts:

1. **Mechanism — `VACUUM INTO` through the open `*sql.DB`.** The snapshot
   is produced by `db.ExecContext(ctx, "VACUUM INTO ?", dest)` on the
   already-open driver connection. No external process, no file copy.
   This is the only correct mechanism given the constraints (below).

2. **Trigger — `applied > 0 AND pending > 0`.** A backup is written
   **only** when the DB already has at least one applied migration *and*
   at least one embedded migration is still pending. A brand-new DB
   (nothing applied yet — everything is "pending" because the file was
   just created) has nothing to lose and is **not** backed up. A DB
   already at head (nothing pending) has nothing about to change and is
   **not** backed up. "Existing/non-empty" is defined as *applied > 0*
   (an established schema), not as *has table rows*.

3. **Failure policy — ABORT.** If the snapshot write fails, `Open`
   returns a wrapped error and applies **no** migration. A migration that
   ran after a failed backup would recreate the exact unrecoverable
   situation this guard exists to prevent.

Backups are timestamped sidecars (`<dbpath>.pre-<highestPendingVersion>.<UTC>.backup`)
and **all are kept** — no pruning in v0.2.0. The guard is silent (it is a
library call with no stdout/stderr contract); the backup file's existence
is the signal.

## Context

`brag`'s migrations are **forward-only** (DEC-002): there is no
`migrate down`, and migration filenames are an append-only contract
(`migrations-are-append-only`). The only recovery from a bad or unwanted
migration is to restore a copy of the pre-migration database — and until
now no such copy was ever made.

**Motivating incident (this session, 2026-06-12).** A v0.2.x `brag`
binary opened the **production** `~/.bragfile` database. `applyMigrations`
silently ran `0003_normalize_tags` + `0004_add_projects`, irreversibly
upgrading a database that had no backup. Nothing was malicious and nothing
was buggy — the runner did exactly what DEC-002 specifies. The gap is that
"upgrade = run a newer binary, which silently migrates on first Open" is a
one-way door with no snapshot behind it. STAGE-008 promotes the brief's
"optional" migration-safety belt to an in-scope stage success criterion on
the strength of this incident.

The discriminator must be precise: the backup exists to protect data that
*already lives* in an *established* DB and is *about to be mutated*. The
overwhelmingly common path — a fresh install creating `~/.bragfile` for the
first time — must stay backup-free (there is nothing to lose, and a sidecar
on every first run is noise). That is exactly `applied == 0`.

### Constraints that shaped the mechanism

- **`no-cgo` / DEC-001 — pure-Go `modernc.org/sqlite`.** The build has no
  C toolchain and ships no `sqlite3` CLI; we cannot shell out to
  `sqlite3 db ".backup"`. The snapshot must go *through the Go driver*.
- **WAL-safety.** A naive `cp`/`io.Copy` of the DB file is not guaranteed
  consistent if a WAL/journal sidecar is live. (`brag` does not currently
  enable WAL, so a quiescent `cp` happens to be safe *today* — but relying
  on that is fragile and future-hostile.) `VACUUM INTO` reads a consistent
  snapshot through the driver regardless of journal mode.

`VACUUM INTO` satisfies both: it is a single driver call that emits one
consistent `.sqlite` file, with no external dependency and no journal-mode
assumption. Verified supported on `modernc.org/sqlite v1.51.0` at design
time (a bound `?` parameter works; a pre-existing destination fails with
`output file already exists`, confirming the timestamped name must be
unique).

## Alternatives Considered

- **Mechanism: file copy (`io.Copy` / `cp`) of the DB file.**
  Rejected: not guaranteed consistent under a live WAL/journal, and it
  duplicates whatever transient sidecar state exists. `VACUUM INTO` is
  strictly safer for equal effort.

- **Mechanism: shell out to the `sqlite3` CLI `.backup`.**
  Rejected: violates the spirit of `no-cgo`/DEC-001 — the CLI is not part
  of the pure-Go build and is not guaranteed present on the user's machine.

- **Trigger: "always back up on Open."**
  Rejected: writes a sidecar on every first-run install and on every
  no-op re-open, which is pure noise. The `applied>0 AND pending>0`
  discriminator backs up exactly when there is both something to lose and
  something about to change.

- **Trigger: "back up only if `entries` has rows."**
  Rejected: couples the guard to one table's contents and would skip a DB
  whose value is its schema/tag/project state rather than its entry rows.
  `applied > 0` ("an established schema exists") is the correct, cheap
  proxy for "existing DB" and needs no row scan.

- **Failure policy: proceed-and-warn (migrate anyway, log a warning).**
  Rejected: a migration that runs after a failed backup is precisely the
  un-backed-up irreversible migration the incident was about. "Never
  migrate without a backup" is the entire point, so a failed backup must
  be fatal to the Open.

- **UX: prompt-and-confirm (`[y/N]`) before a schema-bumping migration.**
  Rejected (see STAGE-008 Design Notes DQ1): breaks non-interactive
  callers (`brag add --json`, cron, CI, non-TTY pipelines), and a prompt
  is exactly the kind of guard a hurried user dismisses — the incident's
  isolation discipline *was* a "prompt" that got bypassed. Auto-backup is
  non-interactive, unbypassable, and incident-fit.

- **Retention: prune to keep-last-N.**
  Deferred, not chosen. Keep-all is simplest and strictly recoverable for
  v0.2.0; the backups are small single files next to a small DB. Pruning
  is possible future work, out of scope here.

## Consequences

- **Positive:** The motivating incident becomes recoverable: any forward
  migration of an established DB now leaves a consistent pre-migration
  snapshot beside it. The guard is non-interactive, so no pipeline breaks.
- **Positive:** Zero new dependencies — `VACUUM INTO` is plain
  `database/sql` against the existing driver.
- **Negative:** A backup file accumulates on every version bump of an
  existing DB (keep-all). For a personal-scale CLI with small DBs this is
  negligible; pruning is deferred.
- **Negative:** `Open` now fails (rather than silently migrating) if the
  snapshot cannot be written (e.g. a read-only or full disk). This is the
  intended trade — refusing to migrate is the safe failure.
- **Neutral:** The backup is a point-in-time `.sqlite` at the *pre-pending*
  schema version. Restoring it means "you are back to before this binary's
  migrations ran," which is exactly the downgrade story DEC-002 otherwise
  lacks.
- **Neutral:** Two existing in-package tests that seed a DB at `0001` and
  re-`Open` (`TestMigrate_ETL_Lossless`, `TestFTS_MigrationBackfillsExistingRows`)
  now emit a sidecar into their `t.TempDir()`. Neither asserts on directory
  contents, so both still pass unchanged (noted in SPEC-036's premise audit).

## Validation

Right if:
- After upgrading an existing `~/.bragfile`, a `*.pre-*.backup` sidecar
  appears and opens cleanly as the pre-migration database.
- A fresh `brew install` + first `brag` run produces **no** sidecar.
- A second `brag` invocation against an at-head DB produces **no** new
  sidecar.
- A failed snapshot aborts `Open` with the DB left un-migrated.

Revisit if:
- Backup files accumulate enough to warrant retention/pruning (then add a
  follow-up DEC for the retention policy).
- `brag` ever enables WAL or grows a second DB backend (re-check the
  snapshot mechanism, though `VACUUM INTO` should still hold).
- A future migration is large enough that a full-DB snapshot is materially
  expensive (unlikely at personal scale).

## References

- Related specs: SPEC-036 (the safety belt itself)
- Related decisions:
  - DEC-002 (forward-only embedded migrations — the regime this guard
    protects; supplies "there is no down-migration")
  - DEC-001 (pure-Go sqlite driver — forces the `VACUUM INTO`-through-the-
    driver mechanism over a `sqlite3` CLI call)
- Related constraints: `no-cgo`, `migrations-are-append-only`,
  `errors-wrap-with-context`, `timestamps-in-utc-rfc3339`,
  `storage-tests-use-tempdir`
- External docs:
  - https://www.sqlite.org/lang_vacuum.html#vacuuminto
