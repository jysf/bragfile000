---
task:
  id: SPEC-002
  type: story
  cycle: ship
  blocked: false
  priority: high
  complexity: M

project:
  id: PROJ-001
  stage: STAGE-001
repo:
  id: bragfile

agents:
  architect: claude-opus-4-7
  implementer: claude-opus-4-7
  created_at: 2026-04-20

references:
  decisions:
    - DEC-001  # pure-Go sqlite driver
    - DEC-002  # embedded migrations, no external lib
    - DEC-004  # tags as comma-joined TEXT (MVP)
    - DEC-005  # INTEGER AUTOINCREMENT IDs
  constraints:
    - no-cgo
    - no-new-top-level-deps-without-decision
    - storage-tests-use-tempdir
    - migrations-are-append-only
    - timestamps-in-utc-rfc3339
    - errors-wrap-with-context
    - test-before-implementation
    - one-spec-per-pr
  related_specs:
    - SPEC-001  # shipped; root command + config path resolver
---

# SPEC-002: SQLite storage + embedded migrations

## Context

SPEC-001 shipped the Cobra skeleton and `internal/config` path
resolver. Nothing in the binary persists yet. This spec builds the
persistence layer that SPEC-003 (`brag add`) and SPEC-004 (`brag
list`) both depend on. When this spec ships:

- A `*storage.Store` type exists that owns a `*sql.DB` against an
  embedded-driver SQLite file.
- Calling `storage.Open(path)` auto-creates the parent directory,
  auto-applies embedded migrations, and returns a ready store.
- `Store.Add(Entry)` inserts a row; `Store.List(ListFilter)` returns
  all rows in reverse-chronological order. (Filter fields land in
  STAGE-002 — MVP ships with the field list empty.)
- All of the above is covered by hermetic unit tests using
  `t.TempDir()`.

Parent stage: `STAGE-001-foundations.md`. Project: PROJ-001 (MVP).

## Goal

Ship `internal/storage` — a `Store` type plus the initial `entries` +
`schema_migrations` schema as `internal/storage/migrations/0001_initial.sql`
embedded via `embed.FS` — such that `storage.Open` on a fresh temp
directory creates the DB, applies the migration, and supports `Add` and
`List` round-trips with RFC3339-UTC timestamps and INTEGER autoincrement
IDs.

## Inputs

- **Files to read:**
  - `docs/architecture.md` — `internal/storage` row in the
    Responsibilities table; Data Flow section's Open sequence.
  - `docs/data-model.md` — `entries` and `schema_migrations` field
    lists; indexes; Data Lifecycle section.
  - `AGENTS.md` §3 (stack), §8 (coding conventions), §9 (testing
    conventions).
  - `/decisions/DEC-001-pure-go-sqlite-driver.md`
  - `/decisions/DEC-002-embedded-migrations.md`
  - `/decisions/DEC-004-tags-comma-joined-for-mvp.md`
  - `/decisions/DEC-005-integer-autoincrement-ids.md`
  - `/guidance/constraints.yaml` — all the storage-facing constraints.
  - SPEC-001 (shipped, in `specs/done/`) — for existing package layout
    and test patterns.
- **External APIs:** none.
- **Related code paths:**
  - `internal/config/` (exists from SPEC-001) — consumed by callers but
    not imported by this package.
  - `cmd/brag/main.go` (exists) — **do not modify in this spec**;
    SPEC-003 wires the first subcommand that uses the store.

## Outputs

- **Files created:**
  - `internal/storage/entry.go` — defines the `Entry` struct
    (ID int64, Title/Description/Tags/Project/Type/Impact string,
    CreatedAt/UpdatedAt time.Time) and the `ListFilter` struct
    (empty for MVP — fields added in STAGE-002).
  - `internal/storage/store.go` — defines `Store`, `Open`, `Close`,
    `Add`, `List`.
  - `internal/storage/migrate.go` — internal migration runner
    (`applyMigrations(ctx, db, src) error`). Reads the embedded FS,
    diffs against `schema_migrations`, applies each missing migration
    in its own transaction.
  - `internal/storage/migrations/0001_initial.sql` — creates the
    `entries` table and the two indexes (`idx_entries_created_at`,
    `idx_entries_project`). **Does NOT create `schema_migrations`**
    — the runner creates that table before reading it.
  - `internal/storage/store_test.go` — tests for `Open`, `Add`, `List`,
    `Close` (hermetic, `t.TempDir()`).
  - `internal/storage/migrate_test.go` — tests for migration runner
    idempotency, ordering, and rollback-on-failure.
  - `go.mod` / `go.sum` — gain `modernc.org/sqlite` as a dependency
    (justified by DEC-001).
- **Files modified:** none. (SPEC-001's files are not touched.)
- **New exports:**
  - `storage.Entry` (struct)
  - `storage.ListFilter` (struct — empty for MVP)
  - `storage.Store` (struct)
  - `storage.Open(path string) (*Store, error)`
  - `(*Store).Close() error`
  - `(*Store).Add(e Entry) (Entry, error)` — returns the inserted
    entry hydrated with `ID`, `CreatedAt`, `UpdatedAt`.
  - `(*Store).List(f ListFilter) ([]Entry, error)` — returns all
    entries in `created_at DESC` order; filter fields are ignored
    for MVP (the stub).
- **Database changes:**
  - Initial schema via `0001_initial.sql`:
    - `CREATE TABLE entries (id INTEGER PRIMARY KEY AUTOINCREMENT,
      title TEXT NOT NULL, description TEXT, tags TEXT, project TEXT,
      type TEXT, impact TEXT, created_at TEXT NOT NULL, updated_at
      TEXT NOT NULL);`
    - `CREATE INDEX idx_entries_created_at ON entries(created_at DESC);`
    - `CREATE INDEX idx_entries_project ON entries(project);`
  - `schema_migrations (version TEXT PRIMARY KEY, applied_at TEXT NOT NULL)`
    is created by the migration runner (not a migration file) so it
    exists before the first migration applies.

## Acceptance Criteria

- [ ] `go build ./...` succeeds with no CGO. `CGO_ENABLED=0 go build
      ./...` also succeeds (smoke test of the no-cgo constraint).
      *[manual]*
- [ ] `storage.Open` on a fresh `t.TempDir()` path succeeds and the
      DB file exists on disk afterward. *[TestOpen_CreatesDBFile]*
- [ ] `storage.Open` on a nested path like `<tmp>/nested/sub/db.sqlite`
      auto-creates the parent directories and succeeds.
      *[TestOpen_CreatesParentDir]*
- [ ] After `Open`, the DB contains `entries` and `schema_migrations`
      tables plus both indexes. *[TestOpen_SchemaExists]*
- [ ] After `Open`, `schema_migrations` contains exactly one row with
      `version = '0001_initial'`. *[TestOpen_MigrationsTracked]*
- [ ] Calling `Open` → `Close` → `Open` on the same path is idempotent
      — no error, no duplicate `schema_migrations` row, no schema
      change. *[TestOpen_Idempotent]*
- [ ] `Store.Add(Entry{Title: "x"})` returns an `Entry` with `ID > 0`,
      `CreatedAt` in UTC (non-zero), and `UpdatedAt.Equal(CreatedAt)`.
      *[TestAdd_BasicInsert]*
- [ ] `Store.Add` persists every field — `Title, Description, Tags,
      Project, Type, Impact` — so that subsequent `List` returns them
      intact. *[TestAdd_PersistsAllFields]*
- [ ] Two `Add` calls with the same `Title` produce two distinct rows
      with distinct IDs (no implicit dedup). *[TestAdd_Duplicates]*
- [ ] Raw `created_at` strings in the DB parse cleanly with
      `time.Parse(time.RFC3339, ...)` and report `Location` as UTC.
      *[TestAdd_TimestampsAreRFC3339UTC]*
- [ ] `Store.List(ListFilter{})` on an empty DB returns a non-nil,
      zero-length slice and nil error. *[TestList_EmptyReturnsEmpty]*
- [ ] `Store.List(ListFilter{})` returns entries in `created_at DESC`
      order. *[TestList_ReverseChronological]*
- [ ] `Store.Close()` closes the underlying `*sql.DB` without error.
      *[TestStore_CloseNoError]*
- [ ] Migration runner applies files in lexical order and records each
      in `schema_migrations`. *[TestMigrate_AppliesInOrder]*
- [ ] A failing migration rolls back its own transaction and leaves
      `schema_migrations` consistent with only the successful
      predecessors. *[TestMigrate_FailedMigrationRollsBack]*
- [ ] `gofmt -l .` empty, `go vet ./...` clean, `go test ./...` green.

## Failing Tests

Written now. All tests use `t.TempDir()` for the DB path
(`storage-tests-use-tempdir` constraint). Use `t.Helper()` in shared
helpers.

### `internal/storage/store_test.go`

Imports: `testing`, `context`, `database/sql`, `os`, `path/filepath`,
`strings`, `time`, package under test (`storage`).

Shared helper (unexported, inside the test file):

- `func newTestStore(t *testing.T) (*Store, string)` — takes
  `t.TempDir()`, builds `filepath.Join(dir, "db.sqlite")`, calls
  `Open`, registers `t.Cleanup(func() { s.Close() })`, returns both
  the store and the DB path so tests can open a raw `*sql.DB` for
  introspection.

Tests:

- **`TestOpen_CreatesDBFile`** — open on a fresh temp path, assert no
  error; `os.Stat(path)` returns nil error.
- **`TestOpen_CreatesParentDir`** — nested path
  `filepath.Join(tmp, "nested", "sub", "db.sqlite")`, assert no error
  and the file exists after `Open`.
- **`TestOpen_SchemaExists`** — after `Open`, open a raw `*sql.DB`
  against the same path and assert each of `entries`,
  `schema_migrations`, `idx_entries_created_at`, `idx_entries_project`
  appears in `sqlite_master` via a small helper
  `objectExists(t, db, kind, name) bool`.
- **`TestOpen_MigrationsTracked`** — after `Open`, raw-query
  `SELECT version FROM schema_migrations` and assert exactly one row
  with `"0001_initial"`.
- **`TestOpen_Idempotent`** — `Open`, `Close`, `Open` on the same
  path. Assert second open succeeds, `schema_migrations` still has
  exactly one row, no schema change.
- **`TestAdd_BasicInsert`** — `Add(Entry{Title: "x"})`. Assert
  `got.ID > 0`, `!got.CreatedAt.IsZero()`, `got.UpdatedAt.Equal(
  got.CreatedAt)`, `got.CreatedAt.Location().String() == "UTC"`.
- **`TestAdd_PersistsAllFields`** — `Add` an `Entry` with
  `Title/Description/Tags/Project/Type/Impact` all set.
  Call `List(ListFilter{})`. Assert the single returned entry equals
  the input across all string fields (ignoring ID/timestamps).
- **`TestAdd_Duplicates`** — two `Add(Entry{Title: "same"})` calls.
  Assert the two returned IDs differ and `List` returns two entries.
- **`TestAdd_TimestampsAreRFC3339UTC`** — `Add` an entry, then raw-
  query `SELECT created_at FROM entries`, read the string, parse
  with `time.Parse(time.RFC3339, raw)`; assert no error and
  `parsed.Location().String() == "UTC"`.
- **`TestList_EmptyReturnsEmpty`** — fresh store, `List(ListFilter{})`
  returns `len == 0`, the slice is non-nil, err is nil.
- **`TestList_ReverseChronological`** — `Add` three entries with
  titles `"a"`, `"b"`, `"c"`, sleeping `10 * time.Millisecond` between
  each so RFC3339 timestamps strictly differ. Assert `List` returns
  titles in order `["c", "b", "a"]`.
- **`TestStore_CloseNoError`** — `Open`, `Close`, assert no error.

### `internal/storage/migrate_test.go`

These drive the migration runner directly, bypassing `Open`, so the
runner should accept an `fs.FS` parameter (see Notes for the
Implementer). Use `testing/fstest.MapFS` for inputs.

Imports: `testing`, `context`, `database/sql`, `strings`,
`testing/fstest`, `_ "modernc.org/sqlite"`, package under test.

Tests:

- **`TestMigrate_AppliesInOrder`** — build a `fstest.MapFS` with
  `0001_a.sql` (`CREATE TABLE a(x INTEGER);`) and `0002_b.sql`
  (`CREATE TABLE b(x INTEGER);`). Open a fresh in-memory-or-temp
  sqlite DB, call `applyMigrations(ctx, db, mapFS)`, assert
  `schema_migrations` has rows `"0001_a"` and `"0002_b"` in that
  order and both tables exist.
- **`TestMigrate_FailedMigrationRollsBack`** — build a MapFS with
  `0001_good.sql` (`CREATE TABLE good(x INTEGER);`) and
  `0002_bad.sql` (`CREATE TABLE if not a valid syntax;;`). Assert the
  runner returns a non-nil error, `schema_migrations` contains only
  `"0001_good"`, `good` table exists, no `bad` table exists.
- **`TestMigrate_Idempotent`** — same inputs as the first test; call
  the runner twice. Assert no error, still exactly two rows in
  `schema_migrations`, no "already exists" SQL error bubbles up.

## Implementation Context

*Read before starting build. Self-contained handoff.*

### Decisions that apply

- `DEC-001` — Pure-Go SQLite driver `modernc.org/sqlite`. Import it
  anonymously (`_ "modernc.org/sqlite"`). Open with
  `sql.Open("sqlite", path)`. **Do not pull `mattn/go-sqlite3`** — it
  requires CGO and violates `no-cgo`.
- `DEC-002` — Embedded migrations via `embed.FS`. Apply in lexical
  filename order. Each migration runs in its own transaction
  alongside the `INSERT INTO schema_migrations` so a mid-migration
  failure rolls back cleanly.
- `DEC-004` — `entries.tags` is a single `TEXT` column holding a
  comma-joined list. `Entry.Tags` is a plain Go `string`. No
  validation beyond "whatever the caller passes is stored verbatim".
- `DEC-005` — `entries.id` is `INTEGER PRIMARY KEY AUTOINCREMENT`;
  `Entry.ID` is `int64`.

### Constraints that apply

For `internal/storage/**`, `internal/storage/migrations/**`, `go.mod`:

- `no-cgo` — blocking. `modernc.org/sqlite` is the only sqlite dep.
- `no-new-top-level-deps-without-decision` — warning.
  `modernc.org/sqlite` is justified by DEC-001. Any other new
  top-level dep requires a DEC before you add it.
- `storage-tests-use-tempdir` — blocking. Every storage test uses
  `t.TempDir()`. Never touch `~/.bragfile`.
- `migrations-are-append-only` — blocking. Once `0001_initial.sql` is
  merged, never edit or rename it. Corrections land as new forward
  migrations in a later spec.
- `timestamps-in-utc-rfc3339` — blocking. Write
  `time.Now().UTC().Format(time.RFC3339)`. Do not use
  `DEFAULT CURRENT_TIMESTAMP` in the schema.
- `errors-wrap-with-context` — warning. Wrap every returned error:
  `fmt.Errorf("open store: %w", err)`, etc.
- `test-before-implementation` — blocking. Write failing tests first.
- `one-spec-per-pr` — blocking. Branch
  `feat/spec-002-sqlite-storage-and-migrations`.

### Prior related work

- **SPEC-001** (shipped on 2026-04-20, archived at
  `specs/done/SPEC-001-go-module-and-cobra-scaffold.md`). PR #1
  merged as `ff4a446`; ship commit `3883a42` added the AGENTS.md §9
  rule on CLI test buffer splitting. That rule does **not** apply to
  SPEC-002 (no CLI tests here) but carry it into SPEC-003+.
- Package layout already established: `cmd/brag/main.go`,
  `internal/cli/`, `internal/config/`. You are adding
  `internal/storage/`.

### Out of scope (for this spec specifically)

If any of these feels necessary during build, write a new spec.

- **FTS5 virtual table + triggers** — STAGE-002. Do NOT add any FTS
  setup to `0001_initial.sql`; it belongs in `0002_add_fts.sql` in a
  later spec.
- **`ListFilter` fields / WHERE-clause logic** — STAGE-002. Keep the
  struct as a named empty struct (`type ListFilter struct{}`).
- **`Get`, `Update`, `Delete`, `Search` methods** — STAGE-002.
- **CLI commands (`brag add`, `brag list`)** — SPEC-003 wires `add`
  over `Store.Add`; SPEC-004 wires `list` over `Store.List`.
  SPEC-002 must not import `cobra` and must not touch `cmd/brag/` or
  `internal/cli/`.
- **Connection pooling, WAL mode, busy-timeout tuning** — the default
  `sql.Open("sqlite", path)` behavior is fine for a single-user CLI.
  Revisit only if a bug demands it.
- **`Close` idempotency beyond `*sql.DB.Close()`** — no finalizer,
  no guard beyond what `database/sql` provides.
- **CI configuration** — STAGE-004.

## Notes for the Implementer

- **Driver import.** Anonymous at the top of `store.go`:
  ```go
  import (
      _ "modernc.org/sqlite"
  )
  ```
  Then `sql.Open("sqlite", path)`.
- **Embed directive.** In `migrate.go`:
  ```go
  //go:embed migrations/*.sql
  var migrationsFS embed.FS
  ```
  And pass `migrationsFS` into the runner from `Open`.
- **Migration runner signature.** Make it
  `applyMigrations(ctx context.Context, db *sql.DB, src fs.FS) error`.
  This lets the tests inject a `fstest.MapFS`; production code passes
  `migrationsFS` rooted at the `migrations` subdirectory via
  `fs.Sub(migrationsFS, "migrations")`.
- **Parent directory.** `os.MkdirAll(filepath.Dir(path), 0o755)`
  before `sql.Open`. If `filepath.Dir(path) == "."` or `"/"`,
  `MkdirAll` is a no-op.
- **`schema_migrations` bootstrap.** Before reading applied versions
  run `CREATE TABLE IF NOT EXISTS schema_migrations (version TEXT
  PRIMARY KEY, applied_at TEXT NOT NULL)` outside any migration
  transaction.
- **Per-migration transaction.**
  ```go
  tx, err := db.BeginTx(ctx, nil)
  // exec migration SQL
  // insert row into schema_migrations
  tx.Commit() // or tx.Rollback() on error
  ```
- **Multi-statement SQL.** `modernc.org/sqlite`'s `ExecContext`
  accepts multi-statement SQL. If a weird edge case forces it, split
  on `;` followed by whitespace/EOL; keep the split logic inside
  `applyMigrations`.
- **`Entry` timestamps.** Keep as `time.Time` in-struct; marshal as
  `time.Now().UTC().Format(time.RFC3339)` on write; parse with
  `time.Parse(time.RFC3339, raw)` on read. A parse failure during
  `List` means the DB is corrupt — return the wrapped error, don't
  silently zero the field.
- **Empty `List` slice.** Return `make([]Entry, 0)` for the empty
  case, not `nil` — the test asserts non-nil.
- **No `init()` functions** (AGENTS.md §8).
- **Version name in `schema_migrations`.** Use the filename stem
  without the `.sql` extension: `"0001_initial"`, not `"0001_initial.sql"`.

---

## Build Completion

*Filled in at the end of the **build** cycle, before advancing to verify.*

- **Branch:** `feat/spec-002-sqlite-storage-and-migrations`
- **PR (if applicable):** opened at end of build (link added on push)
- **All acceptance criteria met?** yes — all 15 named tests pass;
  `gofmt -l .` empty; `go vet ./...` clean; `go build ./...` and
  `CGO_ENABLED=0 go build ./...` both succeed.
- **New decisions emitted:**
  - (none; all storage decisions were pre-existing: DEC-001, DEC-002,
    DEC-004, DEC-005)
- **Deviations from spec:**
  - `List` uses a secondary `ORDER BY id DESC` tie-break on top of
    `created_at DESC`. `time.RFC3339` only has second precision, so
    three `Add`s with 10ms spacing share the same `created_at` string;
    without a tie-break `TestList_ReverseChronological` would be
    non-deterministic. AUTOINCREMENT IDs (DEC-005) are monotone, so
    `id DESC` produces the insertion-order reversal the spec expects.
  - `Add` truncates `time.Now().UTC()` to the second before returning
    it on the hydrated `Entry`, so the in-memory `CreatedAt` matches
    the RFC3339 string that ends up in the DB (the
    `UpdatedAt.Equal(CreatedAt)` assertion stays true on round-trip).
- **Follow-up work identified:**
  - STAGE-002: populate `ListFilter` fields and WHERE-clause logic;
    add FTS5 virtual table via `0002_add_fts.sql`; revisit the
    `created_at DESC, id DESC` tie-break if anything relies on a
    finer-than-second ordering (it probably shouldn't).

### Build-phase reflection (3 questions, short answers)

1. **What was unclear in the spec that slowed you down?**
   — `TestList_ReverseChronological` sleeps 10 ms between `Add` calls
   "so RFC3339 timestamps strictly differ", but RFC3339 is
   second-precision — 10 ms is not enough. The spec didn't call out a
   tie-break ordering. I added a secondary `id DESC`, which is
   consistent with DEC-005 but is an implicit assumption worth making
   explicit in the spec (or bumping the sleep to 1.1 s, which would
   slow the suite for no real benefit).

2. **Was there a constraint or decision that should have been listed but wasn't?**
   — Nothing missing. The spec's Implementation Context plus DEC-001/
   002/004/005 covered every question that came up. Minor nit: the
   `entries` columns are nullable per the data-model doc, so the `List`
   scan has to use `sql.NullString`; the spec's `Outputs` section shows
   them as plain `TEXT` without calling that out. Not a blocker — just
   a thing I noticed when scanning rows.

3. **If you did this task again, what would you do differently?**
   — Write the scan-side of `List` before the insert-side of `Add`.
   Doing it in insert-first order meant I built the hydrated `Entry`
   return value, then discovered (when writing `List`) that nullable
   columns need `sql.NullString`; the return shape in `Add` happened to
   be fine because I never round-trip through a nullable scan there,
   but working read-first would have surfaced the null-handling
   question earlier and kept the two code paths symmetric from the
   start.

---

## Reflection (Ship)

*Appended 2026-04-20 during the **ship** cycle. Outcome-focused,
distinct from the process-focused build reflection above.*

1. **What would I do differently next time?**
   Prescribe the timestamp tie-break in the spec itself.
   `TestList_ReverseChronological` used a 10ms sleep between `Add`
   calls to differentiate timestamps, but RFC3339 is second-precision
   — the sleeps were irrelevant. Build session caught it and added
   `ORDER BY created_at DESC, id DESC`, which is correct and
   consistent with DEC-005 (monotonic `INTEGER AUTOINCREMENT`). Next
   spec I write that involves time-ordering will either bump sleeps
   beyond one second (slow the suite, rarely worth it) or explicitly
   prescribe the monotonic tie-break up front.

2. **Does any template, constraint, or decision need updating?**
   Yes — one AGENTS.md §9 addition: time-based ordering tests must
   use a monotonic tie-break column (e.g., `id DESC` under DEC-005)
   because RFC3339 is second-precision; sleep-based timestamp
   separation alone is insufficient. Applied in the same ship commit
   so any future ordering tests in SPEC-003/004 or beyond inherit it.

3. **Is there a follow-up spec I should write now before I forget?**
   No. SPEC-003 (`brag add`) and SPEC-004 (`brag list`) are already
   in STAGE-001's backlog. The tie-break lesson and the
   nullable-columns observation (from build-phase Q2) apply
   prospectively via the AGENTS.md §9 update and via data-model.md
   (which already marks each column's nullability — the lesson is
   for spec authors to mirror that into the spec's Outputs section).
