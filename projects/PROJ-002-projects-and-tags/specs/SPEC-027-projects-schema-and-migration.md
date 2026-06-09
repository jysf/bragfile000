---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-027
  type: story                      # epic | story | task | bug | chore
  cycle: build
  blocked: false
  priority: high
  complexity: M                    # S | M | L  (L means split it) — L-risk, held to M (see Context)

project:
  id: PROJ-002
  stage: STAGE-007
repo:
  id: bragfile

agents:
  architect: claude-opus-4-8
  implementer: claude-opus-4-8     # usually same Claude, different session
  created_at: 2026-06-08

references:
  decisions: [DEC-017, DEC-002, DEC-005, DEC-015]
  constraints:
    - migrations-are-append-only
    - storage-tests-use-tempdir
    - timestamps-in-utc-rfc3339
    - test-before-implementation
    - errors-wrap-with-context
    - no-sql-in-cli-layer
    - no-new-top-level-deps-without-decision
  related_specs: [SPEC-025, SPEC-028, SPEC-029, SPEC-030, SPEC-031, SPEC-032]
---

# SPEC-027: Projects schema + `0004_*` migration + DEC-017

## Context

This is the **foundation spec of STAGE-007** — every later spec in the
stage (the `brag project` CLI in SPEC-028/029, the status dashboard in
SPEC-030, the `here` resolver in SPEC-031, the `brag add` auto-fill in
SPEC-032) builds on the schema and the read primitives laid down here.
It is the one spec the stage flagged as **L-risk, held to M** by
splitting the Store *mutation* methods out to SPEC-029 and the cwd
*resolver* out to SPEC-031 (STAGE-007 Spec Backlog).

The single load-bearing question the stage exists to resolve — **how the
existing free-text `entries.project` column relates to the new `projects`
entity** — is decided here and emitted as **DEC-017**. That answer
determines whether the `0004_*` migration backfills anything, what
`brag project delete` does to entries (SPEC-029), and how the status
dashboard's recent-brag count is computed (SPEC-030). It is also the
**split-watch trigger**: if DEC-017 had required a non-trivial
`entries.project` backfill, that backfill would peel into its own spec
rather than push SPEC-027 to L. **It does not** — DEC-017 chooses the
soft-string-match model, whose defining property is **zero backfill**
(see DEC-017 and `## Notes for the Implementer`). That choice is what
keeps this spec at M.

Parent stage: `STAGE-007` (projects core), PROJ-002 (`projects-and-tags`).
Prior foundation: `SPEC-025` (the `0003_normalize_tags` migration + the
Store transactional patterns this spec mirrors).

## Goal

Lay down the `projects` + `project_locations` schema as a forward-only
`0004_*` migration, emit **DEC-017** (the `entries.project` ↔ `projects`
relationship + the status enum + the state-note model), and ship the
Store **read/foundation primitives** (`CreateProject`, `GetProject`,
`ListProjects`, `AddLocation`). No `brag project` CLI command ships in
this spec.

## Inputs

- **Files to read:**
  - `internal/storage/migrations/0003_normalize_tags.sql` — the migration
    style + the "runs inside the runner's per-migration transaction, no
    `BEGIN`/`COMMIT`" rule the `0004_*` file must follow.
  - `internal/storage/store.go` — the transactional `Add`/`Update`
    patterns (`BeginTx` + `defer tx.Rollback()` + hydrate-and-return) the
    new Store methods mirror; `ErrTagExists`/`ErrTagNotFound` lookup shape.
  - `internal/storage/entry.go` — the `Entry` struct + `ListFilter`
    (the free-text `Project` field DEC-017 preserves).
  - `internal/storage/errors.go` — sentinel-error style for the two new
    errors.
  - `internal/storage/migrate.go` — confirms migrations apply in lexical
    order from `migrations/*.sql` via `embed.FS`; the runner adds the
    `schema_migrations` row, so the `0004_*` file must not.
  - `internal/storage/store_test.go` + `internal/storage/fts_test.go` —
    the four count-bump sites (see `## Outputs`).
  - `docs/data-model.md` — gains the two new tables.
- **Related code paths:** `internal/storage/` only. No `internal/cli/`
  changes in this spec.

## Outputs

- **Files created:**
  - `internal/storage/migrations/0004_add_projects.sql` — the new
    forward-only migration (DEC-002). Creates `projects` +
    `project_locations` + their index. **No backfill** (DEC-017
    soft-string-match → `entries.project` is untouched).
  - `internal/storage/project.go` — the `Project` struct + the four
    Store methods (`CreateProject`, `GetProject`, `ListProjects`,
    `AddLocation`) + two sentinel errors (`ErrProjectExists`,
    `ErrLocationExists`). (One concept per file under `internal/storage/`,
    per AGENTS.md §8; `project.go` mirrors `entry.go` + the project
    methods on `*Store`.)
  - `internal/storage/project_test.go` — the failing tests below.
  - `decisions/DEC-017-entries-project-relationship.md` — the DEC.
- **Files modified:**
  - `internal/storage/store_test.go` — **count-bump** (see audit below).
  - `internal/storage/fts_test.go` — **count-bump** (see audit below).
  - `docs/data-model.md` — add the `projects` + `project_locations`
    entity tables; add the `0004_*` line to "Schema Evolution" /
    "Indexes"; strike the "Projects normalization" line from "Future
    schema shapes (deferred)" (status-change). Reference DEC-017.
  - `guidance/questions.yaml` — file `project-state-note-shape`
    (sub-choice < 0.8; see DEC-017 and §14).
- **New exports:**
  - `storage.Project` struct (`ID`, `Name`, `Status`, `StateNote`,
    `Locations []string`, `CreatedAt`, `UpdatedAt`).
  - `func (s *Store) CreateProject(p Project) (Project, error)`
  - `func (s *Store) GetProject(id int64) (Project, error)`
  - `func (s *Store) ListProjects() ([]Project, error)`
  - `func (s *Store) AddLocation(projectID int64, path string) error`
  - `storage.ErrProjectExists`, `storage.ErrLocationExists`
- **Database changes:** `0004_add_projects.sql` (forward-only).

### Premise audit (`projects/_templates/premise-audit.md`), run at design

Run at design **2026-06-08**; all greps below were executed and reconciled.

- [x] **Addition / count-bump — greps run, literal-count assertions
  listed as planned bumps.** Adding the fourth migration breaks the
  literal-count assertions that run against the **real** `migrationsFS`.
  Grep: `grep -rn '0003_normalize_tags\|count = %d, want 3' internal/storage/*_test.go`.
  **Four sites, two files, all planned bumps:**
  - `internal/storage/store_test.go:172` — `want := []string{"0001_initial",
    "0002_add_fts", "0003_normalize_tags"}` → append `"0004_add_projects"`;
    the line-173 index comparison gains a `versions[3] != want[3]` term and
    the slug must match the file stem.
  - `internal/storage/store_test.go:206-208` — comment + `if count != 3` →
    `4` (and update the `// 0001…0003` comment to name `0004_add_projects`).
  - `internal/storage/fts_test.go:149` — same `want` list → append
    `"0004_add_projects"`; the line-150 index comparison gains a
    `got[3] != want[3]` term.
  - `internal/storage/fts_test.go:266-270` — `if count != 3` → `4`.
  - **Stays untouched (verified):** `internal/storage/migrate_test.go:146`
    (`count … want 2`) runs against in-test `fstest.MapFS` fixtures, **not**
    the real FS — restated from STAGE-006's verified correction. Grep:
    `grep -n 'MapFS\|want 2' internal/storage/migrate_test.go` → all hits are
    MapFS-scoped.
  - §12(a): the `want` lists are lexically ordered and `0004_add_projects`
    appends last. Confirmed at design by a scratch lexical assertion
    (`"0003_normalize_tags" < "0004_add_projects"`) — see §12(b) below.
- [x] **Inversion / removal — greps run, invalidated tests enumerated.**
  DEC-017 chooses **soft string match**: `entries.project` stays free
  text, `ListFilter.Project` keeps its `e.project = ?` equality, and **no
  existing behavior is inverted or removed.** Grep:
  `grep -rn 'ListFilter\|\.Project\|ByProject\|GroupEntriesByProject' --include='*.go' internal/`
  → hits in `internal/cli/list.go:59`, `internal/cli/export.go:83`
  (`filter.Project = v`), `internal/aggregate/aggregate.go`
  (`ByProject`, `GroupEntriesByProject`), and their tests.
  **Reconciliation: zero planned rewrites.** Every one of these reads or
  groups on the free-text `entries.project` string; DEC-017 preserves that
  string and its semantics verbatim, so all of them keep passing
  unchanged. This is the central reason the soft-match model holds the
  spec at M (an FK or optional-link model **would** have invalidated these
  premises — see DEC-017 Alternatives). Enumerated here explicitly per the
  inversion case so it is a *design conclusion*, not a build-time silence.
- [x] **Status change — greps run, every doc hit listed as updates/stays.**
  The new tables change a feature's shipping status in the docs. Grep:
  `grep -rln -i 'project' docs/ README.md` (17 files). The per-spec doc
  scope here is narrow (the comprehensive tutorial/architecture/api sweep
  is **STAGE-008**); only `docs/data-model.md` carries schema-status claims
  this spec invalidates:
  - `docs/data-model.md` — **UPDATE.** Add the two entity tables; add the
    `0004_*` evolution + index lines; **strike** the "Projects
    normalization … Deferred" bullet under "Future schema shapes (deferred,
    not in PROJ-001)" (it is now realized, not deferred); add DEC-017 to
    References.
  - **STAYS (STAGE-008, not this spec):** `docs/tutorial.md`,
    `docs/api-contract.md`, `docs/architecture.md` — these describe the
    *command surface* and *user workflow*, neither of which ships in
    SPEC-027 (no `brag project` command here). The remaining hits
    (`docs/brag-entry.schema.json`, `docs/blog/*`, `docs/reports/*`,
    `docs/framework-feedback/*`, `docs/macos-notarization-checklist.md`,
    `docs/CONTEXTCORE_ALIGNMENT.md`, `README.md`) reference the *existing*
    free-text `entries.project` or are historical/quoted prose — DEC-017
    preserves that field, so none carries an invalidated status claim.
- [x] **Cross-check:** actual grep hits reconciled against the lists above
  at design; no un-enumerated hit remained.

## Acceptance Criteria

Testable outcomes. Happy path, error cases, edge cases.

- [ ] **Migration applies.** After `storage.Open` on a fresh DB, tables
  `projects` and `project_locations` exist with the columns DEC-017
  specifies, and `schema_migrations` contains exactly four rows ending in
  `0004_add_projects`.
- [ ] **Migration is idempotent.** A second `Open` on the same path does
  not re-apply `0004` (still exactly four `schema_migrations` rows).
- [ ] **`entries.project` is lossless and unchanged (DEC-017).** The
  `0004_*` migration touches neither `entries` nor its data; an entry's
  `project` string and `brag list --project foo` behavior are byte-stable
  before and after the migration. (No `entries.project` value is read,
  rewritten, or backfilled.)
- [ ] **`CreateProject` round-trips.** A created project reads back via
  `GetProject` with the same `Name`/`Status`/`StateNote`, a positive
  generated `ID`, and `CreatedAt == UpdatedAt` (UTC, RFC3339,
  second-truncated) on insert.
- [ ] **`status` defaults to `active`; `state_note` defaults to empty.**
  `CreateProject` with an empty `Status` persists `"active"`; an empty
  `StateNote` persists `""` (both via the column `DEFAULT` and the Store's
  normalization — they must agree).
- [ ] **Duplicate name is a typed error.** A second `CreateProject` with
  an existing `Name` returns an error wrapping `ErrProjectExists`; no
  partial row is left behind (transactional).
- [ ] **`GetProject` miss is `ErrNotFound`.** `GetProject` on an unknown
  id returns an error wrapping `storage.ErrNotFound` (reusing the existing
  sentinel, matching `Get`).
- [ ] **`AddLocation` + location round-trip.** After `AddLocation` for two
  paths, `GetProject(id).Locations` returns both, ordered by insertion
  (location `id` ASC).
- [ ] **Location paths are globally unique.** `AddLocation` of a path
  already attached (to the same *or* a different project) returns an error
  wrapping `ErrLocationExists`. (This is the schema guarantee SPEC-031's
  `here` resolver relies on: a path maps to at most one project.)
- [ ] **`ListProjects` ordering.** `ListProjects` returns every project
  ordered `updated_at DESC, id DESC` (recency, with the monotonic
  id tie-break per AGENTS.md §9), each hydrated with its `Locations`.
- [ ] **No regressions.** `go test ./...`, `gofmt -l .`, `go vet ./...`,
  and `CGO_ENABLED=0 go build ./...` are clean; output shapes
  (DEC-011/013/014) and search (DEC-010) stay byte-stable (untouched).

## Failing Tests

Written during **design**, BEFORE build. Build makes these pass. All
storage tests use `t.TempDir()` (`storage-tests-use-tempdir`); none touch
`~/.bragfile`. Reuse the package's `newTestStore(t)` helper where it fits;
where a test needs the raw `*sql.DB` to inspect `schema_migrations`, open
a second `sql.Open("sqlite", path)` handle as `store_test.go` already does.

- **`internal/storage/project_test.go`** (new)
  - `"TestCreateProject_RoundTrip"` — `CreateProject({Name:"bragfile",
    Status:"active", StateNote:"next: cut v0.2.0"})` returns `ID > 0` and
    `CreatedAt == UpdatedAt` (non-zero, UTC); `GetProject(id)` returns the
    same `Name`/`Status`/`StateNote`. Asserts freshness via `ID > 0`, **not**
    a timestamp inequality (AGENTS.md §9 addendum — no `sleep`).
  - `"TestCreateProject_StatusDefaultsActive"` — `CreateProject` with
    `Status:""` reads back `Status == "active"`; with `StateNote:""` reads
    back `StateNote == ""`.
  - `"TestCreateProject_DuplicateNameErrProjectExists"` — second create
    with the same name returns `errors.Is(err, ErrProjectExists)`; a
    follow-up `ListProjects()` shows exactly one row (no partial insert).
  - `"TestGetProject_NotFound"` — `GetProject(99999)` returns
    `errors.Is(err, ErrNotFound)`.
  - `"TestAddLocation_RoundTripOrderedByID"` — attach `/a` then `/b` to a
    project; `GetProject(id).Locations` deep-equals `[]string{"/a","/b"}`.
  - `"TestAddLocation_DuplicatePathErrLocationExists_SameProject"` —
    re-attaching `/a` to the same project returns
    `errors.Is(err, ErrLocationExists)`.
  - `"TestAddLocation_DuplicatePathErrLocationExists_DifferentProject"` —
    attaching `/a` to a *second* project (after it is attached to the
    first) returns `errors.Is(err, ErrLocationExists)` — the global-uniqueness
    guarantee SPEC-031 depends on.
  - `"TestListProjects_OrderedByUpdatedAtThenIDDesc"` — create three
    projects; assert the returned slice is ordered `updated_at DESC,
    id DESC`. To avoid an RFC3339 second-precision tie producing flaky
    order (AGENTS.md §9), the test sets identical `created_at`/`updated_at`
    across all three (insert via the Store, which stamps one `now`) and
    asserts the **id DESC** tie-break deterministically (newest id first);
    a separate sub-case with distinct `updated_at` values — set by reading
    rows back and not relying on sleep — is **out of scope** (mutation that
    bumps `updated_at` is SPEC-029). So this test locks the tie-break only.
  - `"TestListProjects_HydratesLocations"` — a project with two locations
    comes back from `ListProjects` with both in `Locations`.
- **`internal/storage/project_migration_test.go`** (new — migration-level)
  - `"TestOpen_ProjectsTablesExist"` — after `Open`, `projects` and
    `project_locations` exist (`sqlite_master` lookup), and a raw insert
    honoring the schema (name UNIQUE, path UNIQUE, status default)
    succeeds; a duplicate name and a duplicate path each fail at the DB
    level.
  - `"TestOpen_MigrationsTracked_Includes0004"` — `schema_migrations`
    ordered = `["0001_initial","0002_add_fts","0003_normalize_tags",
    "0004_add_projects"]` (count `== 4`). *(This is the new migration-list
    assertion; it duplicates the bumped `store_test.go`/`fts_test.go` sites
    intentionally as the project-side anchor — keep all three in sync.)*
  - `"TestOpen_0004Idempotent"` — open twice; `schema_migrations` count
    stays `4`.
  - `"TestMigration0004_DoesNotTouchEntries"` — open a DB, `Add` an entry
    with `Project:"platform"`, close, re-open (no-op 0004 already applied),
    and assert the entry's `project` is still `"platform"` and
    `List(ListFilter{Project:"platform"})` still returns it. Locks the
    DEC-017 lossless-and-unchanged contract for `entries.project`.
- **`internal/storage/store_test.go`** (modify — count-bump)
  - `TestOpen_MigrationsTracked` (`:172`) — extend `want` to four entries
    ending `"0004_add_projects"`; extend the index comparison.
  - `TestOpen_Idempotent` (`:206-208`) — `want 3` → `want 4`; update the
    naming comment.
- **`internal/storage/fts_test.go`** (modify — count-bump)
  - the `want` list (`:149`) — extend to four entries.
  - the `count … want 3` site (`:266-270`) — → `4`.

> **§12(a) note for build:** every expected literal above is
> design-decided against the real migration set — the `want` list order
> (`0004` appends last lexically), the `count == 4`, and the migration
> slug `0004_add_projects` were all confirmed at design (§12(b)).
> Transcribe them; do not re-derive by hand.

## Implementation Context

*Read this section (and the files it points to) before starting the
build cycle. It is the handoff document, folded into the spec.*

### Decisions that apply

- **`DEC-017` (emitted by this spec)** — `entries.project` relates to
  `projects` by **soft string match**: `entries.project` stays free text,
  joins to `projects.name` opportunistically, **no backfill, no FK, no
  link column**. Also fixes the `projects.status` enum
  (`active`/`paused`/`done`/`archived`, default `active`, **not**
  DB-CHECK-enforced — validated in the Store, matching how `entries.type`
  is free text) and the **single** free-text `state_note` column (the
  state/next-action note). Read it before build — it is the why behind
  every schema choice here.
- `DEC-002` — embedded forward-only migrations; the `0004_*` file is the
  mechanism. It runs **inside the runner's per-migration transaction** —
  do **not** add `BEGIN`/`COMMIT`, and do **not** insert into
  `schema_migrations` (the runner does that). Mirror the header-comment
  style of `0003_normalize_tags.sql`.
- `DEC-005` — `projects.id` and `project_locations.id` are
  `INTEGER PRIMARY KEY AUTOINCREMENT` (monotonic; gives the id-DESC
  tie-break the ordering tests rely on).
- `DEC-015` — the polymorphic `taggings` shape. **Relevant only as a
  forward guarantee:** projects become a second taggable type with **no
  schema change**, but SPEC-027 writes **no `'project'` taggings** (schema-
  ready only; STAGE-007 design question #4 default confirmed — see "Out of
  scope"). Design question #6 (position base for `'project'` taggings) is
  therefore **not triggered** by this spec.

### Constraints that apply

(see `/guidance/constraints.yaml` for full text)

- `migrations-are-append-only` — `0004_add_projects.sql` is new; never
  edit a shipped migration. (0001–0003 are untouched.)
- `storage-tests-use-tempdir` — every new test uses `t.TempDir()`.
- `timestamps-in-utc-rfc3339` — `created_at`/`updated_at` written by the
  Go layer as `time.Now().UTC().Truncate(time.Second).Format(time.RFC3339)`,
  exactly like `Store.Add`.
- `test-before-implementation` — the Failing Tests above are written first.
- `errors-wrap-with-context` — `fmt.Errorf("create project: %w", err)`,
  matching the existing method style.
- `no-sql-in-cli-layer` — N/A in practice (no CLI here), but the new SQL
  lives only in `internal/storage/`.
- `no-new-top-level-deps-without-decision` — none added; `database/sql` +
  `modernc.org/sqlite` only.

### Prior related work

- `SPEC-025` (shipped 2026-06-07) — `0003_normalize_tags.sql` + the Store
  transactional cutover. This spec mirrors its migration header style and
  its `BeginTx`/`defer Rollback`/hydrate-and-return method shape. The §12(a)
  lesson codified at STAGE-006 close (run a test's embedded expected
  literals against their source at design) is applied here to the four
  count-bump sites and the new migration-list assertion.
- `STAGE-007` Design Notes — the five surfaced design questions; #1
  (entries.project relationship) and #3 (status enum + state note) are
  resolved here as DEC-017; #4 (project tagging surface) confirmed
  schema-ready-only; #2 (here resolution) and #6 (tagging position base)
  are **not** this spec's to decide.

### Out of scope (for this spec specifically)

If any of these feels necessary during build, **stop and flag** — do not
expand this spec.

- **Any `brag project` CLI command.** `new`/`list`/`show` are SPEC-028;
  `edit`/`archive`/`delete` are SPEC-029. No file under `internal/cli/`
  changes here.
- **Store mutation methods.** `UpdateProject` / `ArchiveProject` /
  `DeleteProject` are SPEC-029. This spec ships read/create primitives
  only (`CreateProject` creates; it does not update). `updated_at` is set
  equal to `created_at` on insert and is not bumped by anything in this
  spec.
- **The `here` cwd resolver + path normalization policy.** SPEC-031.
  `AddLocation` stores the path **verbatim** as given; the
  exact/nearest-ancestor/longest-prefix resolution policy and any
  absolute-path normalization are SPEC-031's to decide. (The schema's
  global `UNIQUE(path)` is laid here because SPEC-031's resolver needs the
  "one path → one project" guarantee, but the *resolution* is not built.)
- **`brag add` `--project` auto-fill.** SPEC-032.
- **The `brag project status` dashboard + recent-brag count.** SPEC-030.
  The recent-brag count will be a `COUNT(entries WHERE entries.project =
  projects.name)` join under DEC-017's soft-match model, but it is **not**
  computed here.
- **Writing `'project'` taggings.** Schema-ready only (DEC-015); deferred
  (STAGE-007 design question #4 default; the `brag project tag` surface is
  a STAGE-008/PROJ-003 candidate).
- **Any backfill of `entries.project`.** DEC-017's whole point: there is
  none. (This is the L-split-watch trigger that did **not** fire.)

## Notes for the Implementer

### The `0004_add_projects.sql` literal (transcribe verbatim — §12(b)-validated)

This exact body was run through `modernc.org/sqlite` at design (§12(b)
below): it applies inside the runner's transaction, the two `UNIQUE`
constraints fire, and the `status`/`state_note` defaults apply. Mirror
`0003`'s header-comment convention.

```sql
-- 0004_add_projects.sql — SPEC-027 (PROJ-002 / STAGE-007)
-- First-class projects entity (DEC-017). Adds the `projects` table and a
-- `project_locations` join supporting one-project-many-directories.
-- Forward-only (DEC-002); runs inside the migration runner's per-migration
-- transaction — do NOT add BEGIN/COMMIT or a schema_migrations insert here.
--
-- DEC-017 (soft string match): entries.project stays free text and is NOT
-- touched by this migration — no FK, no link column, no backfill. The
-- relationship is an opportunistic join on projects.name at query time.
-- Validated at design (§12(b)) against modernc.org/sqlite 1.51.0
-- (SQLite 3.53.1): tables create, name/path UNIQUE enforced, status and
-- state_note defaults applied, entries untouched.

-- The projects entity. status is a Store-validated enum (active | paused |
-- done | archived), not a DB CHECK — mirroring entries.type's free-text
-- column, so a future status value is an additive Store change, not a
-- table rebuild under forward-only migrations. state_note is the single
-- free-text state/next-action note rendered by brag project status (SPEC-030).
CREATE TABLE projects (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    name       TEXT NOT NULL UNIQUE,
    status     TEXT NOT NULL DEFAULT 'active',
    state_note TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

-- One project, many directories. path is globally UNIQUE so a cwd resolves
-- to at most one project (the guarantee SPEC-031's `here` resolver relies on).
CREATE TABLE project_locations (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL REFERENCES projects(id),
    path       TEXT NOT NULL UNIQUE
);

CREATE INDEX idx_project_locations_project ON project_locations(project_id);
```

### The Store methods (`internal/storage/project.go`)

Mirror `store.go` exactly: `context.Background()`, `BeginTx` +
`defer tx.Rollback()` for the writes, `fmt.Errorf("…: %w", err)` wrapping,
`time.Now().UTC().Truncate(time.Second)` for timestamps.

- `CreateProject(p Project) (Project, error)`: normalize `Status` (`"" →
  "active"`) before insert so the Go value and the DB default agree; stamp
  `now` into both `created_at` and `updated_at`; `INSERT INTO projects(...)`;
  on a `UNIQUE` failure on `name`, return `ErrProjectExists` (detect by
  pre-checking `SELECT COUNT(*) FROM projects WHERE name = ?` inside the tx,
  matching `RenameTag`'s existence-check shape — cleaner than string-matching
  the driver error). Hydrate and return `p` with `ID`, `CreatedAt`,
  `UpdatedAt`, and the normalized `Status`. `Locations` is empty on a fresh
  create (caller attaches via `AddLocation`).
- `GetProject(id int64) (Project, error)`: `SELECT id, name, status,
  state_note, created_at, updated_at FROM projects WHERE id = ?`; map
  `sql.ErrNoRows` → `ErrNotFound` (reuse the existing sentinel, like `Get`);
  then hydrate `Locations` via `SELECT path FROM project_locations WHERE
  project_id = ? ORDER BY id`.
- `ListProjects() ([]Project, error)`: `SELECT … FROM projects ORDER BY
  updated_at DESC, id DESC`; for each row hydrate `Locations` (a second
  per-row query is fine at personal scale — mirror the simple-query style;
  do not prematurely optimize into a join). Return a non-nil empty slice on
  no rows (`out := make([]Project, 0)`, like `List`).
- `AddLocation(projectID int64, path string) error`: `INSERT INTO
  project_locations(project_id, path) VALUES (?, ?)`; on a `UNIQUE(path)`
  failure return `ErrLocationExists` (pre-check `SELECT COUNT(*) FROM
  project_locations WHERE path = ?`). Store `path` **verbatim** — no
  normalization here (SPEC-031 owns that). Optionally verify `projectID`
  exists; not required by the tests, keep it minimal.

### Errors (`internal/storage/errors.go` or `project.go`)

Two new sentinels in the `errors.go` style:

```go
// ErrProjectExists is returned (wrapped) by CreateProject when name is taken.
var ErrProjectExists = errors.New("project already exists")

// ErrLocationExists is returned (wrapped) by AddLocation when the path is
// already attached to some project (paths are globally unique).
var ErrLocationExists = errors.New("location already exists")
```

`GetProject` reuses `ErrNotFound` (do **not** add an `ErrProjectNotFound`).

### Gotchas

- **The migration runner inserts the `schema_migrations` row** (see
  `migrate.go:99`). The `0004_*` file must not — adding it would double-write.
- **Lexical ordering is load-bearing.** `0004_add_projects` must sort after
  `0003_normalize_tags`; the count-bump `want` lists append it last.
- **Don't bump `updated_at` anywhere.** Nothing in this spec mutates a
  project after create; `created_at == updated_at` for every row this spec
  can produce, which is what the round-trip test asserts.
- **`gofmt` the SQL-adjacent Go and run `go vet ./...`** before opening the PR.

### §12(b) design-time verification (run at design 2026-06-08 — PASSED)

The `0004_add_projects.sql` body above was executed against the real driver
(`modernc.org/sqlite`) in a scratch `internal/storage` test, on top of a DB
with 0001–0003 already applied:

- Both tables create inside a single transaction (the runner's shape). ✅
- `INSERT` honoring the schema succeeds; `status`/`state_note` omitted →
  read back `"active"` / `""` (defaults apply). ✅
- Duplicate `name` → `UNIQUE` violation. ✅
- Duplicate `path` (same and different project) → `UNIQUE` violation. ✅
- Two locations round-trip ordered by `id`. ✅
- `entries` rows are untouched by the migration (DEC-017 lossless). ✅
- Lexical assertion `"0003_normalize_tags" < "0004_add_projects"` → true,
  confirming the count-bump `want`-list append position (§12(a)). ✅

The scratch test was removed after the pre-flight; `go test ./internal/storage/`
remained green. Build transcribes the literal; verify diffs against it.

---

## Build Completion

*Filled in at the end of the **build** cycle, before advancing to verify.*

- **Branch:**
- **PR (if applicable):**
- **All acceptance criteria met?** yes/no
- **New decisions emitted:**
  - `DEC-017` — entries.project ↔ projects relationship (soft string match)
    + status enum + state-note model *(emitted at design; confirm no build
    deviation)*
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
