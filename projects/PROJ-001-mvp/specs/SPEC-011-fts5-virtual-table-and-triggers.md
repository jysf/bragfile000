---
task:
  id: SPEC-011
  type: story
  cycle: build
  blocked: false
  priority: high
  complexity: M

project:
  id: PROJ-001
  stage: STAGE-002
repo:
  id: bragfile

agents:
  architect: claude-opus-4-7
  implementer: claude-opus-4-7
  created_at: 2026-04-21

references:
  decisions:
    - DEC-001  # pure-Go sqlite driver (modernc.org/sqlite)
    - DEC-002  # embedded migrations, append-only
    - DEC-004  # tags comma-joined TEXT
  constraints:
    - no-cgo
    - no-sql-in-cli-layer
    - storage-tests-use-tempdir
    - migrations-are-append-only
    - timestamps-in-utc-rfc3339
    - errors-wrap-with-context
    - test-before-implementation
    - one-spec-per-pr
  related_specs:
    - SPEC-002  # shipped; storage layer, migrations runner, 0001_initial.sql
---

# SPEC-011: FTS5 virtual table + triggers

## Context

Seventh spec in STAGE-002. First schema change since SPEC-002's
`0001_initial.sql`. Introduces SQLite's Full-Text Search v5 (FTS5)
as a virtual table that mirrors the searchable fields of `entries`
(title, description, tags, project, impact). Triggers keep the FTS
index in sync automatically on every INSERT / UPDATE / DELETE.
Backfill happens as part of the migration so users with existing
entries (the author has 10 at time of writing) get indexed
transparently on upgrade.

**No Go code changes in this spec.** Pure SQL in a new migration
file plus SQL-level tests that prove the sync machinery works. The
user-visible `brag search "query"` command lives in SPEC-012 as a
thin wrapper over a new `Store.Search` method.

**Premise audit** (per AGENTS.md §9 SPEC-010 rule): this spec is
purely additive at the SQL layer. No existing behavior is inverted
or removed. Existing tests remain untouched. The `0001_initial.sql`
migration stays byte-identical (migrations-are-append-only
constraint). Existing `Store.Add/Get/Update/Delete/List` methods
are unchanged.

Parent stage: `STAGE-002-capture-and-retrieval.md`. Project: PROJ-001.

## Goal

Ship `internal/storage/migrations/0002_add_fts.sql` — a single
migration file that, inside one transaction, (1) creates the
`entries_fts` virtual table as an external-content FTS5 index over
`entries`, (2) backfills existing rows into the index, and (3)
installs AFTER INSERT / UPDATE / DELETE triggers on `entries` that
keep `entries_fts` synchronized. Verified by six new tests covering:
fresh-DB migration, migration-on-non-empty-DB backfill, trigger-
driven sync through each of Store's CRUD methods, and a raw FTS
MATCH query returning the expected ids.

## Inputs

- **Files to read:**
  - `docs/data-model.md` — "Virtual: `entries_fts` (arrives in
    STAGE-002)" already describes the intended shape. This spec
    ships that shape; the doc gets updated from planned to shipped.
  - `docs/architecture.md` — storage-layer section.
  - `AGENTS.md` §8 (conventions), §9 (testing — separate buffers
    N/A here since no CLI tests, but fail-first + assertion
    specificity + locked-decisions-need-tests apply).
  - `/decisions/DEC-001-pure-go-sqlite-driver.md` — the pure-Go
    driver ships FTS5 (needs smoke-test confirmation in this
    spec's first test).
  - `/decisions/DEC-002-embedded-migrations.md` — migrations are
    append-only; each runs in its own transaction.
  - `/decisions/DEC-004-tags-comma-joined-for-mvp.md` — tags
    stored as comma-joined TEXT; FTS5's default unicode61
    tokenizer will split on commas, which is what we want for
    tag search (each tag becomes a token).
  - `/guidance/constraints.yaml` — `migrations-are-append-only`
    + `no-cgo`.
  - `internal/storage/migrations/0001_initial.sql` — the first
    migration; 0002 extends it without modification.
  - `internal/storage/migrate.go` — runs embedded migrations in
    order inside transactions. No change needed.
  - `internal/storage/store.go` — `Store.Add/Update/Delete` exist
    from SPEC-002/009; triggers keep FTS in sync for all three
    without Go-side changes.
  - `internal/storage/store_test.go` + `get_test.go` + the new
    `storagetest` sub-package — prior-art test patterns.
  - External: [SQLite FTS5 docs](https://sqlite.org/fts5.html),
    specifically §3 (external content tables) and §4.4.3 (triggers
    for updating the FTS index).
- **External APIs:** none.
- **Related code paths:** `internal/storage/migrations/`,
  `internal/storage/`.

## Outputs

- **Files created:**
  - `internal/storage/migrations/0002_add_fts.sql` — the migration.
    Contains `CREATE VIRTUAL TABLE entries_fts USING fts5(...)`,
    the backfill `INSERT ... SELECT`, and the three triggers.
  - `internal/storage/fts_test.go` — new test file with the six
    FTS-specific tests. Kept separate from `store_test.go` to
    keep the FTS concern self-contained and avoid bloating the
    existing file.
- **Files modified:**
  - `docs/data-model.md` — promote the "Virtual: `entries_fts`
    (arrives in STAGE-002)" section from planned to shipped.
    Include the actual migration SQL as a reference and document
    the external-content pattern.
  - `docs/architecture.md` — update the storage-layer section to
    mention the FTS index as an additional ride-along on top of
    `entries`.
- **Files NOT modified** (explicit premise-audit result — SPEC-010
  rule):
  - `internal/storage/store.go` — no Store method changes.
    Triggers handle sync.
  - `internal/storage/migrations/0001_initial.sql` — frozen per
    `migrations-are-append-only`.
  - `internal/storage/migrate.go` — the runner already handles
    multi-statement migrations in a transaction.
  - Any CLI layer file — no CLI surface in this spec.
- **New exports:** none.
- **Database changes:**
  - New virtual table `entries_fts` (external content, references
    `entries`). See Locked Design Decisions and Notes for shape.
  - Three new triggers on `entries`: `entries_ai`, `entries_ad`,
    `entries_au` (after-insert, after-delete, after-update).
  - New row in `schema_migrations`: `version = '0002_add_fts'`.

## Locked design decisions (inline — no new DEC needed)

Each decision has a mapped failing test per AGENTS.md §9 "locked-
decisions-need-tests" (SPEC-009 ship rule).

1. **External-content FTS5** — `USING fts5(..., content='entries',
   content_rowid='id')`. The FTS table does NOT store the content
   itself; it stores an inverted index referencing `entries` by
   `id` (via SQLite's `rowid`). Rejected standalone FTS (doubles
   storage) and contentless FTS with manual `INSERT INTO ft(ft,
   ...) VALUES ('delete', ...)` for every update (triggers become
   harder to reason about).
   *Test: TestFTS_VirtualTableShape.*

2. **Indexed fields: title, description, tags, project, impact.**
   `id`, `created_at`, `updated_at` are NOT in the FTS index —
   they're numeric/timestamp, not useful for text search. If the
   user wants date-range filtering they use `brag list --since`,
   not `brag search`.
   *Test: TestFTS_IndexedFieldsOnly (implicit via MATCH
   behaviour — a search for a created_at timestamp fragment
   returns no results).*

3. **Default tokenizer: unicode61.** SQLite's default. Lowercases
   input, strips non-alphanumerics. This means a tag value
   `"auth,perf"` tokenizes as `auth` and `perf` (comma is a
   separator), which is exactly what we want for tag search
   semantics. No `tokenize='...'` clause in the `CREATE VIRTUAL
   TABLE` statement.
   *Test: TestFTS_UnicodeTokenizerSplitsOnPunctuation (search
   for `"auth"` matches a row with `tags='auth,perf'`).*

4. **Backfill inside the migration transaction.** The migration
   executes `CREATE VIRTUAL TABLE`, then `INSERT INTO entries_fts
   (rowid, ...) SELECT id, ... FROM entries`, then creates the
   three triggers. If any step fails, the whole thing rolls back
   (DEC-002's per-migration-transaction contract holds). Critical
   for the author's DB which already has 10 entries at upgrade
   time.
   *Test: TestFTS_MigrationBackfillsExistingRows.*

5. **Trigger shape uses FTS5's INSERT-as-command syntax for
   delete/update.** DELETE trigger issues `INSERT INTO
   entries_fts(entries_fts, rowid, title, ...) VALUES('delete',
   old.id, old.title, ...)` — the `entries_fts` column as first
   arg with value `'delete'` is FTS5's documented command mode.
   UPDATE trigger issues delete-old + insert-new in the same
   block.
   *Test: TestFTS_TriggerDeleteRemovesFromIndex,
   TestFTS_TriggerUpdateReplacesIndexedRow.*

6. **FTS5 is compiled into modernc.org/sqlite.** Spec assumes
   this; first test verifies empirically by running
   `CREATE VIRTUAL TABLE ... USING fts5(...)` against a fresh
   `*sql.DB` and checking for no error. If this fails, the whole
   spec is blocked and verify should reject with a follow-up spec
   to evaluate alternatives (e.g., switching to an FTS-equipped
   driver or a standalone index implementation).
   *Test: TestFTS_SmokeCreateUnderPureGoDriver.*

7. **No CLI, no Store method changes.** `brag search` and
   `Store.Search` land in SPEC-012. Triggers handle write-side
   sync automatically; no Go-side `Store.Add/Update/Delete` code
   changes. Migration is the entire Go-code diff for this spec
   (beyond tests).
   *Test: existing Store tests remain green (regression guard).*

## Acceptance Criteria

- [ ] `CREATE VIRTUAL TABLE ... USING fts5(...)` succeeds under
      the `modernc.org/sqlite` driver (proves FTS5 is available
      in the pure-Go SQLite we're shipping).
      *[TestFTS_SmokeCreateUnderPureGoDriver]*
- [ ] After `storage.Open` on a fresh DB, the `entries_fts` virtual
      table exists (`SELECT name FROM sqlite_master WHERE type =
      'table' AND name = 'entries_fts'` returns a row).
      *[TestFTS_VirtualTableShape]*
- [ ] After `storage.Open` on a fresh DB, all three triggers
      (`entries_ai`, `entries_ad`, `entries_au`) exist in
      `sqlite_master WHERE type='trigger'`.
      *[TestFTS_TriggersExistAfterMigration]*
- [ ] After `storage.Open` on a fresh DB, `schema_migrations`
      contains both `0001_initial` and `0002_add_fts`.
      *[TestFTS_BothMigrationsTracked]*
- [ ] Migration on a non-empty DB backfills existing rows: `Add`
      three entries against a DB that has ONLY 0001 applied, then
      run the migration runner again to apply 0002; `entries_fts`
      is populated with three rows matching the entries' indexed
      fields. *[TestFTS_MigrationBackfillsExistingRows]*
- [ ] `Store.Add` on a migrated DB causes the new row to appear in
      `entries_fts` (via the `entries_ai` trigger). Verified by
      raw SQL query on `entries_fts WHERE rowid = ?`.
      *[TestFTS_TriggerInsertAddsToIndex]*
- [ ] `Store.Update` replaces the indexed content for that rowid
      in `entries_fts`. After update, MATCH on the OLD title
      returns no hits; MATCH on the NEW title returns the row.
      *[TestFTS_TriggerUpdateReplacesIndexedRow]*
- [ ] `Store.Delete` removes the row from `entries_fts`. After
      delete, `SELECT COUNT(*) FROM entries_fts WHERE rowid = ?`
      returns 0.
      *[TestFTS_TriggerDeleteRemovesFromIndex]*
- [ ] `SELECT rowid FROM entries_fts WHERE entries_fts MATCH 'foo'`
      returns the ids of entries containing `foo` in any indexed
      field, including within a `tags='auth,perf'`-style
      comma-joined value.
      *[TestFTS_MatchQueryReturnsExpectedIds,
      TestFTS_UnicodeTokenizerSplitsOnPunctuation]*
- [ ] Existing SPEC-001..010 tests remain green. No existing test
      is modified; no file under `internal/cli/` is touched.
      *[manual: go test ./...]*
- [ ] `gofmt -l .` empty, `go vet ./...` clean, `CGO_ENABLED=0 go
      build ./...` succeeds, `go test ./...` green.
- [ ] `docs/data-model.md` — `entries_fts` section promoted from
      planned to shipped; actual migration SQL included or
      referenced.
- [ ] `docs/architecture.md` — storage-layer description mentions
      the FTS index as an additional ride-along (one-line update).

## Failing Tests

Written now. All tests use `t.TempDir()` for the DB path
(`storage-tests-use-tempdir` constraint). Use `t.Helper()` in
shared helpers. Fail-first run before implementation (§9 SPEC-003
lesson). Every locked design decision has at least one paired
failing test (§9 SPEC-009 lesson).

All tests live in `internal/storage/fts_test.go` (new file)
except where noted. Reuse the existing `newTestStore(t)` helper
from `store_test.go` where appropriate.

### `internal/storage/fts_test.go` (new)

Imports: `testing`, `context`, `database/sql`, `path/filepath`,
`time`, `_ "modernc.org/sqlite"`, package under test.

Shared helper (unexported, inside the test file):

- `func openRawDB(t *testing.T, path string) *sql.DB` — opens a
  raw `*sql.DB` against the given path for direct queries against
  `sqlite_master`, `entries_fts`, `schema_migrations`. Used by
  tests that need to inspect DB state without going through the
  `Store` abstraction. Registers `t.Cleanup(func() { db.Close() })`.

Tests:

- **`TestFTS_SmokeCreateUnderPureGoDriver`** — open a fresh
  `*sql.DB` at `t.TempDir()`, execute `CREATE VIRTUAL TABLE fts5_smoke
  USING fts5(content)`. Assert no error. (If modernc.org/sqlite ever
  ships without FTS5, this is the single test that fails loudly.)

- **`TestFTS_VirtualTableShape`** — `storage.Open`, then raw-query
  `SELECT name, sql FROM sqlite_master WHERE type = 'table' AND
  name = 'entries_fts'`. Assert:
  - A row is returned.
  - The captured `sql` contains each of: `"title"`,
    `"description"`, `"tags"`, `"project"`, `"impact"`.
  - The `sql` contains `"content='entries'"` AND
    `"content_rowid='id'"` (locks decision #1).
  - The `sql` does NOT contain `"tokenize"` (confirms default
    tokenizer per decision #3).

- **`TestFTS_TriggersExistAfterMigration`** — `storage.Open`, raw-
  query `SELECT name FROM sqlite_master WHERE type = 'trigger'`.
  Assert the result contains exactly: `entries_ai`, `entries_ad`,
  `entries_au` (and no others).

- **`TestFTS_BothMigrationsTracked`** — `storage.Open`, raw-query
  `SELECT version FROM schema_migrations ORDER BY version`.
  Assert exactly two rows: `0001_initial` then `0002_add_fts`.

- **`TestFTS_MigrationBackfillsExistingRows`** — this is the most
  interesting test. Build a DB manually:
  1. Open a raw `*sql.DB`, apply ONLY `0001_initial.sql`
     (via direct `db.Exec(content)` — not the Store.Open path,
     which would apply all migrations). `INSERT` three rows
     directly into `entries` with distinctive titles
     (`"alpha-backfill"`, `"beta-backfill"`, `"gamma-backfill"`)
     and representative values in tags/project/etc.
  2. Close the raw db.
  3. Call `storage.Open(path)` — the migration runner should see
     0001 is already applied and now apply 0002.
  4. Raw-query `SELECT rowid, title FROM entries_fts ORDER BY
     rowid`. Assert three rows with titles matching the three
     backfilled entries.
  5. Raw-query `SELECT COUNT(*) FROM schema_migrations` — assert
     2.

- **`TestFTS_TriggerInsertAddsToIndex`** — `storage.Open`, call
  `Store.Add(Entry{Title: "insert-trigger-xyz"})`. Raw-query
  `SELECT rowid, title FROM entries_fts WHERE rowid = ?` with the
  returned id. Assert one row, title matches exactly.

- **`TestFTS_TriggerUpdateReplacesIndexedRow`** — `storage.Open`,
  Add an entry with title `"old-trigger-phrase"`. Update it to
  `"new-trigger-phrase"` via `Store.Update`. Run two MATCH
  queries:
  1. `SELECT rowid FROM entries_fts WHERE entries_fts MATCH
     'old-trigger-phrase'` — assert no rows.
  2. `SELECT rowid FROM entries_fts WHERE entries_fts MATCH
     'new-trigger-phrase'` — assert exactly one row, matching
     id.

- **`TestFTS_TriggerDeleteRemovesFromIndex`** — Add an entry,
  confirm it's in `entries_fts` via raw count. `Store.Delete(id)`.
  Raw-query `SELECT COUNT(*) FROM entries_fts WHERE rowid = ?` —
  assert 0.

- **`TestFTS_MatchQueryReturnsExpectedIds`** — Add three entries
  with distinctive, non-overlapping title words (e.g.,
  `"zebrafish"`, `"platypus"`, `"quokka"`). Run `SELECT rowid FROM
  entries_fts WHERE entries_fts MATCH ? ORDER BY rowid` with
  `"zebrafish"`. Assert exactly the zebrafish id. Same for
  `"platypus"` and `"quokka"`.

- **`TestFTS_UnicodeTokenizerSplitsOnPunctuation`** — Add an entry
  with `Tags: "auth,perf,backend"`. MATCH on `"auth"` — assert
  the row is returned. MATCH on `"perf"` — same. MATCH on
  `"xxx_missing_tag"` — assert no rows. Explicitly locks decision
  #3: default tokenizer treats commas as separators.

Notes for the implementer on testing patterns:

- Fail-first: write all 10 tests, run `go test ./...` once BEFORE
  any SQL change. Expected failure modes:
  - `TestFTS_SmokeCreateUnderPureGoDriver` — should PASS if
    modernc.org/sqlite has FTS5 (expected). If it fails, STOP and
    raise a question in `/guidance/questions.yaml` before
    proceeding.
  - All other FTS tests — should fail because `0002_add_fts.sql`
    doesn't exist yet, so `entries_fts` / triggers / second
    migration row don't exist.
- `TestFTS_MigrationBackfillsExistingRows` manually seeds a
  0001-only DB. The implementer can either read
  `0001_initial.sql` via `embed.FS` in the test, or duplicate the
  CREATE TABLE statements inline (small, stable). Read-from-embed
  is cleaner; duplication is simpler. Either is acceptable —
  document the choice inline.
- Use `context.Background()` for any `*sql.DB.ExecContext` /
  `QueryContext` calls — matches storage package conventions.

## Implementation Context

*Read before starting build. Self-contained handoff.*

### Decisions that apply

- `DEC-001` — Pure-Go SQLite driver. FTS5 is available in
  modernc.org/sqlite; `TestFTS_SmokeCreateUnderPureGoDriver` is
  the empirical proof.
- `DEC-002` — Embedded migrations, append-only, each in its own
  transaction. `0002_add_fts.sql` is the second file. Named
  strictly `0002_add_fts.sql` so it sorts after `0001_initial.sql`
  in the lexical-order iteration. Its `schema_migrations.version`
  value is `"0002_add_fts"` (filename stem).
- `DEC-004` — Tags comma-joined. FTS5's default unicode61
  tokenizer treats commas as separators, which makes tag search
  work naturally without any special handling. If the tokenizer
  ever changes, DEC-004 and this spec both need revisit — but at
  MVP scale the default wins.

### Constraints that apply

For `internal/storage/migrations/**`,
`internal/storage/fts_test.go`, `docs/**`:

- `no-cgo` — blocking. The pure-Go SQLite driver ships FTS5; no
  CGO is introduced.
- `no-sql-in-cli-layer` — blocking. This spec touches no CLI
  files. Triggers live in SQL, in the storage layer.
- `storage-tests-use-tempdir` — blocking. Every new test uses
  `t.TempDir()`.
- `migrations-are-append-only` — blocking. `0001_initial.sql` is
  never edited or renamed. `0002_add_fts.sql` is a new file.
- `timestamps-in-utc-rfc3339` — not relevant here; FTS doesn't
  touch timestamps.
- `errors-wrap-with-context` — warning. No new Go code that
  returns errors.
- `test-before-implementation` — blocking.
- `one-spec-per-pr` — blocking. Branch
  `feat/spec-011-fts5-virtual-table-and-triggers`.

### AGENTS.md lessons that apply

- §9 fail-first (SPEC-003). Applies as noted above.
- §9 locked-decisions-need-tests (SPEC-009). 7 decisions mapped
  to 10 paired tests (decisions 1–3 and 5 get one test each;
  decision 4 gets the backfill test specifically; decision 6
  gets the smoke test; decision 7 is guarded by "existing tests
  still pass").
- §9 premise-audit-for-removed-behavior (SPEC-010). Explicitly
  performed: this spec removes or inverts no existing behavior.
  `Store.Add/Update/Delete` continue to work byte-identically —
  the only difference is triggers fire automatically on their
  writes. No tests need deletion.
- §12 "During design". Every implementation option in Notes
  below passes every blocking constraint. No "either is
  acceptable" language.

### Prior related work

- **SPEC-002** (shipped). Storage layer, migration runner,
  `0001_initial.sql` creating the `entries` and
  `schema_migrations` tables. This spec builds on top without
  modification.
- **SPEC-006** (shipped). `Store.Get` — used in test assertions
  that check read-back.
- **SPEC-007** (shipped). `storage.ErrNotFound` — not directly
  used here but the spec's handlers stay consistent with that
  pattern if any new error emerges (none expected).
- **SPEC-009** (shipped). `Store.Update` — the update trigger
  test verifies sync through this method.

### Out of scope (for this spec specifically)

If any of these feels necessary during build, write a new spec.

- **`brag search "query"` command** — SPEC-012. Thin wrapper over
  `Store.Search`.
- **`Store.Search(query string) ([]Entry, error)` method** —
  SPEC-012. One `database/sql` call wrapping `SELECT ... FROM
  entries_fts WHERE entries_fts MATCH ?`. Depends on this spec's
  table existing.
- **FTS5 query syntax features** — phrase search, prefix search
  (`auth*`), NEAR(), column filters. All supported by default
  FTS5; the design of how `brag search` exposes them to users is
  SPEC-012's concern.
- **BM25 ranking / ORDER BY rank** — SPEC-012.
- **Custom tokenizer** (porter stemmer, trigram). Default
  unicode61 is sufficient for MVP.
- **`brag db reindex` command** — future polish spec if index
  drift ever becomes a real problem.
- **Performance tuning** — partition sizes, page caching, etc.
  Premature at O(hundreds) of rows.
- **Fuzzy search / typo tolerance** — out of MVP scope.
- **Search result highlighting** — future polish.

## Notes for the Implementer

- **`0002_add_fts.sql` shape.** Keep it tight; everything inside
  the migration runner's implicit transaction.
  ```sql
  -- External-content FTS5 index over entries. rowid maps to
  -- entries.id via content_rowid='id'.
  CREATE VIRTUAL TABLE entries_fts USING fts5(
      title, description, tags, project, impact,
      content='entries',
      content_rowid='id'
  );

  -- Backfill existing rows. On a fresh DB this is a no-op
  -- (SELECT returns zero rows); on an upgraded DB it populates
  -- the index from whatever is already in `entries`.
  INSERT INTO entries_fts(rowid, title, description, tags, project, impact)
  SELECT id, title, description, tags, project, impact FROM entries;

  -- Keep the index in sync on future writes. FTS5 uses
  -- INSERT INTO entries_fts(entries_fts, ...) VALUES('delete', ...)
  -- for deletes from the index (see SQLite FTS5 docs §4.4.3).
  CREATE TRIGGER entries_ai AFTER INSERT ON entries BEGIN
      INSERT INTO entries_fts(rowid, title, description, tags, project, impact)
      VALUES (new.id, new.title, new.description, new.tags, new.project, new.impact);
  END;

  CREATE TRIGGER entries_ad AFTER DELETE ON entries BEGIN
      INSERT INTO entries_fts(entries_fts, rowid, title, description, tags, project, impact)
      VALUES ('delete', old.id, old.title, old.description, old.tags, old.project, old.impact);
  END;

  CREATE TRIGGER entries_au AFTER UPDATE ON entries BEGIN
      INSERT INTO entries_fts(entries_fts, rowid, title, description, tags, project, impact)
      VALUES ('delete', old.id, old.title, old.description, old.tags, old.project, old.impact);
      INSERT INTO entries_fts(rowid, title, description, tags, project, impact)
      VALUES (new.id, new.title, new.description, new.tags, new.project, new.impact);
  END;
  ```

  Note the explicit column list in `INSERT INTO entries_fts(rowid,
  title, ...)` — critical because `entries_fts` has an implicit
  `rowid` that must be set explicitly to align with `entries.id`.
  The `SELECT id, title, ...` backfill also explicitly uses `id`
  as the rowid.

- **Trigger NULL handling.** Entries columns `description`, `tags`,
  `project`, `type`, `impact` are nullable TEXT. The triggers
  pass `new.<col>` through verbatim. FTS5 treats NULL values as
  empty content (no tokens indexed). That's the right behavior —
  a row with NULL description doesn't contribute tokens from
  that column.

- **Column name quoting.** `type` is a SQL keyword in some
  dialects but not in SQLite (and not reserved in SQLite's FTS5
  column list). The `0001_initial.sql` already uses unquoted
  `type` successfully. `0002_add_fts.sql` does the same.

- **Multi-statement SQL in one migration file.** The existing
  migration runner (SPEC-002) executes the entire `.sql` file
  contents via `db.Exec(string(fileBytes))` inside a transaction.
  SQLite accepts multi-statement SQL that way. The four
  statements (CREATE VIRTUAL TABLE + INSERT SELECT + three CREATE
  TRIGGER) run as one unit.

- **Backfill-on-fresh-DB semantics.** On a fresh DB with no
  entries, the `INSERT ... SELECT` selects zero rows and is a
  no-op. No special-casing needed.

- **`docs/data-model.md` update.** The existing "Virtual:
  `entries_fts` (arrives in STAGE-002)" block becomes:
  ```
  ### Virtual: `entries_fts` (shipped in SPEC-011)

  External-content FTS5 virtual table indexing title, description,
  tags, project, impact from `entries`. Kept in sync via
  AFTER INSERT / UPDATE / DELETE triggers on `entries`.

  [actual CREATE VIRTUAL TABLE statement]
  ```
  Keep it close to what's actually in the migration file so
  readers don't have to cross-reference.

- **`docs/architecture.md` update.** The storage-layer description
  mentions the `entries` table and the `schema_migrations`
  tracker. Add one line: "Plus an FTS5 ride-along table
  `entries_fts` that indexes title, description, tags, project,
  impact and stays in sync via SQL triggers (SPEC-011)."

- **No `init()` functions** (§8). No Go code changes anyway.

---

## Build Completion

*Filled in at the end of the **build** cycle, before advancing to verify.*

- **Branch:**
- **PR (if applicable):**
- **All acceptance criteria met?** yes/no
- **New decisions emitted:**
  - (none expected; design decisions are inline)
- **Deviations from spec:**
  - [list]
- **Follow-up work identified:**
  - [any new specs for the stage's backlog]

### Build-phase reflection (3 questions, short answers)

1. **What was unclear in the spec that slowed you down?**
   — <answer>

2. **Was there a constraint or decision that should have been listed but wasn't?**
   — <answer>

3. **If you did this task again, what would you do differently?**
   — <answer>

---

## Reflection (Ship)

*Appended during the **ship** cycle.*

1. **What would I do differently next time?**
   — <answer>

2. **Does any template, constraint, or decision need updating?**
   — <answer>

3. **Is there a follow-up spec I should write now before I forget?**
   — <answer>
