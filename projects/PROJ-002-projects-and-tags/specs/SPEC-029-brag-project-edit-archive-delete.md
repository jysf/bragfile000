---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-029
  type: story                      # epic | story | task | bug | chore
  cycle: build                     # frame | design | build | verify | ship
  blocked: false
  priority: medium
  complexity: M                    # S | M | L  (L means split it)

project:
  id: PROJ-002
  stage: STAGE-007
repo:
  id: bragfile

agents:
  architect: claude-opus-4-8
  implementer: claude-opus-4-8     # usually same Claude, different session
  created_at: 2026-06-09

references:
  decisions: [DEC-018, DEC-017, DEC-006, DEC-007, DEC-011, DEC-005, DEC-002]
  constraints:
    - no-sql-in-cli-layer
    - stdout-is-for-data-stderr-is-for-humans
    - storage-tests-use-tempdir
    - errors-wrap-with-context
    - test-before-implementation
    - one-spec-per-pr
    - timestamps-in-utc-rfc3339
    - no-cgo
    - no-new-top-level-deps-without-decision
  related_specs: [SPEC-027, SPEC-028, SPEC-030, SPEC-031, SPEC-032, SPEC-033]
---

# SPEC-029: brag project edit / archive / delete

## Context

This is the **mutation half** of the `brag project` command surface and
the third of STAGE-007's six specs (SPEC-027 shipped the schema + read
primitives; SPEC-028 shipped the `brag project` parent + `new`/`list`/
`show`). It adds the three commands that *change* a registered project —
`edit` (scalar fields), `archive` (a non-destructive status flip), and
`delete` (destructive) — plus the Store mutation methods they call
(`UpdateProject`, `ArchiveProject`, `DeleteProject`).

It is where two latent properties of the schema become real:

1. **`updated_at` actually moves.** SPEC-027 set `created_at == updated_at`
   on insert and never bumped it, so `ListProjects`' `updated_at DESC`
   ordering was only ever exercised by the `id`-DESC tie-break. The
   mutations here are the first writes that advance `updated_at` past
   `created_at`, making the recency ordering observable (and testable
   without a sleep — see §9 below).
2. **`brag project delete` has a defined blast radius.** STAGE-007's
   Success Criteria flag delete's blast radius as a thing that must be
   "defined, tested, and consciously chosen — not incidental." That
   choice is emitted here as **DEC-018**, building directly on DEC-017's
   soft-string-match model (entries are deliberately left untouched) and
   on SPEC-027's shipped forward note (FK enforcement is OFF, so
   `project_locations` must be deleted manually in-transaction — no
   cascade fires).

Parent stage: `STAGE-007-projects-core.md` (PROJ-002, `projects-and-tags`).
Prior shipped work this extends: `SPEC-028` (PR #41, `85c5173`) — the
`brag project` parent + the `RunE`/open-resolve/confirmation-to-stderr
patterns mirrored here; `SPEC-027` (PR #40, `7a67834`) — the `Project`
struct + the transactional Store method shape.

### Complexity: held to M by peeling location-editing (the L-watch)

STAGE-007 flagged this spec to **watch for L**: `edit` (scalar fields
*and* `--add-path`/`--remove-path` location editing) + `archive` +
`delete` (blast radius) read **L** together. Per the stage's prescribed
split, **location editing is peeled into its own spec (SPEC-033)** and
`edit` here is **scalar-only** (`--name` / `--status` / `--state-note`).
That keeps SPEC-029 at M: three thin Store methods, three thin cobra
subcommands on the existing parent, one DEC, no migration. The peel is
**flagged, not silently absorbed** — see `## Outputs` (SPEC-033 added to
the backlog) and `## Out of scope`. The `here` resolver (SPEC-031), the
status dashboard (SPEC-030), and `brag add` auto-fill (SPEC-032) are
unchanged and remain later specs.

## Goal

Add `brag project edit <name|id>` (scalar fields: `--name` rename,
`--status`, `--state-note`), `brag project archive <name|id>` (status →
`archived`, non-destructive, recoverable), and `brag project delete
<name|id>` (destructive, confirmation-guarded) to the existing `brag
project` parent, backed by three transactional Store methods
(`UpdateProject` / `ArchiveProject` / `DeleteProject`) that **bump
`updated_at`** on every mutation. Emit **DEC-018** locking `delete`'s
blast radius. **No migration, no schema change.**

## Inputs

- **Files to read:**
  - `internal/storage/project.go` — the `Project` struct, the SPEC-027
    primitives (`CreateProject` / `GetProject` / `GetProjectByName` /
    `ListProjects` / `AddLocation` / `locationsForProject`); the new
    mutation methods sit beside them, mirroring their tx shape.
  - `internal/storage/store.go` — the canonical mutation shapes
    (`Update`: `BeginTx` + `defer tx.Rollback()` + `UPDATE` +
    `RowsAffected()==0 → ErrNotFound` + hydrate-via-`Get`; `Delete`:
    delete dependents first, then the row; `RenameTag`'s
    name-uniqueness pre-check inside the tx).
  - `internal/storage/errors.go` — sentinel-error style (`ErrNotFound`,
    `ErrProjectExists` reused; one new sentinel `ErrInvalidStatus`).
  - `internal/cli/project.go` — the parent (`NewProjectCmd`), the
    open/resolve pattern, the name-first/id-fallback resolution in
    `runProjectShow` (LD2), `renderProjectPlain`, and the confirmation-
    to-stderr style (`Created project %q.`) the new RunEs mirror.
  - `internal/cli/delete.go` — the **y/N confirmation + `--yes`/`-y`**
    pattern (`bufio.NewReader(cmd.InOrStdin())`, decline → `Aborted.` on
    stderr, exit 0) that `project delete` matches verbatim.
  - `internal/cli/errors.go` — `UserErrorf` / `ErrUser`.
  - `internal/cli/delete_test.go` + `internal/cli/project_test.go` — the
    CLI test patterns (`newProjectTestRoot` / `runProjectCmd`, out/err
    buffer separation, `root.SetIn(strings.NewReader("y\n"))` for the
    prompt).
  - `internal/cli/list_test.go` — `seedListEntry(t, dbPath, title, tags,
    project, typ)` used by the "delete leaves entries' project string
    untouched" test.
  - `decisions/DEC-017` (the relationship + status enum + state_note the
    mutations honor), `DEC-007` (inline `RunE` validation), `DEC-011`
    (output family — relevant only as the family these confirmation-only
    commands sit beside; they emit **no** `--format` data), `DEC-005`
    (autoincrement PKs → the `id`-DESC tie-break).
- **External APIs:** none. Plain Store calls; no new dependency.
- **Related code paths:** `internal/storage/project.go` (+`errors.go`),
  `internal/cli/project.go` (3 subcommands on the existing parent),
  `docs/api-contract.md` (3 command sections + a DEC-018 References row),
  `decisions/DEC-018-*.md` (new).

## Outputs

- **Files created:**
  - `decisions/DEC-018-project-delete-blast-radius.md` — the delete
    blast-radius decision (entries untouched / locations manual-delete /
    `'project'` taggings cleaned in-tx; archive-vs-delete distinction).
- **Files modified:**
  - `internal/storage/project.go` — add `UpdateProject(id int64, p
    Project) (Project, error)`, `ArchiveProject(id int64) error`,
    `DeleteProject(id int64) error`, and the package-level
    `validProjectStatuses` enum map.
  - `internal/storage/errors.go` — add `ErrInvalidStatus`.
  - `internal/storage/project_test.go` — add the storage mutation tests
    below (**additive**; no SPEC-027/028 test rewritten).
  - `internal/cli/project.go` — add `newProjectEditCmd` /
    `newProjectArchiveCmd` / `newProjectDeleteCmd` + their RunEs + the
    shared `resolveProjectByNameOrID` helper; register all three on the
    existing `NewProjectCmd` parent; widen the parent's `Short`.
  - `internal/cli/project_test.go` — add the CLI mutation tests below
    (**additive**).
  - `docs/api-contract.md` — **status-change UPDATE**: add three command
    sections after `### brag project show <name|id>` and before `### brag
    completion <shell>`; add a `DEC-018` row to the References list
    (after the `DEC-017` row at line 559).
  - `projects/PROJ-002-projects-and-tags/stages/STAGE-007-projects-core.md`
    — update the SPEC-029 backlog line (cycle) and **add SPEC-033** (the
    peeled location-editing spec); bump the Count line.
- **New exports:**
  - `func (s *Store) UpdateProject(id int64, p Project) (Project, error)`
  - `func (s *Store) ArchiveProject(id int64) error`
  - `func (s *Store) DeleteProject(id int64) error`
  - `storage.ErrInvalidStatus`
  - `func cli.NewProjectCmd()` gains `edit`/`archive`/`delete` (the
    constructor is already exported + registered in `main.go`).
- **Database changes:** **NONE.** No migration. The mutations are
  `UPDATE`/`DELETE` on tables `0004_add_projects` already created. This
  is load-bearing (it is why the count-bump audit case is empty).

### Premise audit (`projects/_templates/premise-audit.md`), run at design

Greps were **run** at design **2026-06-09** and reconciled against the
lists below.

```
- [x] Inversion/removal: greps run — NONE existing inverted (the updated_at-bump
      ordering case is an ADDITION, not an inversion; confirmed below)
- [x] Addition/count-bump: greps run — NONE (no migration; no count-coupled assertion)
- [x] Status-change: greps run, every doc hit listed as updates/stays
- [x] Cross-check: actual grep hits reconciled against the lists above
```

**1. Inversion / removal — NONE existing inverted.** The mutations add
behavior; they invert nothing already shipped. Greps run:
- `grep -rn "OrderedByUpdatedAt\|CreatedAt.Equal\|UpdatedAt" internal/storage/project_test.go`
  → `TestListProjects_OrderedByUpdatedAtThenIDDesc` (`:205`) creates three
  projects **in one second** and asserts the **`id`-DESC tie-break**
  (their `updated_at` are equal); `TestCreateProject_RoundTrip` (`:67-68`)
  asserts `CreatedAt.Equal(UpdatedAt)` on a **fresh insert**. SPEC-029
  mutates **neither** of those fixtures and changes **neither**
  `CreateProject` nor the equal-`updated_at` insert path — so both
  premises hold verbatim. Enabling `updated_at` bumps makes the
  *distinct*-`updated_at` ordering newly testable; that is a **new test**
  (`TestUpdateProject_BumpReordersListProjects`), an **ADDITION**, not a
  rewrite of the existing tie-break test. Confirmed: zero planned
  rewrites/deletions.
- `grep -rn "func.*CreateProject\|status ==" internal/storage/project.go`
  → `CreateProject` keeps its `"" → "active"` default and is **not**
  retrofitted with enum validation here (it never receives a non-default
  status — `brag project new` always creates `active`). The new enum
  validation lives in `UpdateProject` only. No existing `CreateProject`
  test premise changes.

**2. Addition / count-bump — NONE.** No migration is added;
`schema_migrations` stays at **4**. Grep run:
`grep -rn "0004_add_projects\|want 4\|count != 4" internal/` → the only
hits are the SPEC-027 sites (`store_test.go:172,206-208`,
`project_migration_test.go:116,150-151`, `fts_test.go:149,269-270`), all
already at 4 and **untouched** by SPEC-029. Grep for command-set
coupling: `grep -rn "Commands()\|len(.*Commands" internal/cli cmd` → the
only hit is `list_test.go:629`, a **test-local** root iterated to read a
`Short`; **no test enumerates or counts the production root or the
`project` parent's subcommand set**, so attaching three subcommands to
`NewProjectCmd` couples to no assertion. (Same registration-gap finding
as SPEC-026/028; restated.) No bumps.

**3. Status change — the three new commands.** Grep run:
`grep -rln -i "project" docs/ README.md` (17 files). Per-spec doc scope is
narrow (the comprehensive tutorial/architecture sweep is **STAGE-008**).
Disposition:
- **Updates (this spec):**
  - `docs/api-contract.md` — **UPDATE.** Add `### brag project edit
    <name|id>`, `### brag project archive <name|id>`, `### brag project
    delete <name|id>` after the existing `### brag project show` section
    (line 502) and before `### brag completion` (line 503). Add a
    `DEC-018` row to References after the `DEC-017` row (line 559). (The
    DEC-017/011 rows already present **stay**.)
- **Stays here (STAGE-008, or no status claim invalidated):**
  - `docs/data-model.md` — **STAYS.** SPEC-027 added the `projects` +
    `project_locations` tables already; SPEC-029 adds **no schema, no
    column** (only `UPDATE`/`DELETE` DML), so no data-model status claim
    is invalidated.
  - `docs/tutorial.md`, `docs/architecture.md`, `README.md` — **STAY**
    (the full projects+tags tutorial + architecture diagram refresh is
    the **STAGE-008** sweep; remaining `project` hits reference the
    *existing* `entries.project` string, untouched by DEC-017/018).
  - `docs/brag-entry.schema.json`, `docs/CONTEXTCORE_ALIGNMENT.md`,
    `docs/macos-notarization-checklist.md`, `docs/development.md`,
    `docs/blog/**`, `docs/framework-feedback/**`, `docs/reports/**` —
    **STAY.** Historical / process / `entries.project` input-contract
    prose; no shipped-behavior status claim about `brag project edit/
    archive/delete`.

**4. Cross-check.** Actual grep hits reconciled against the lists above
at design; no un-enumerated hit remained. (Build-side: re-run
`grep -rln -i "project" docs/ README.md` before the doc edit and treat
any delta as a question, per the premise-audit cross-check rule.)

## Acceptance Criteria

Testable outcomes. Happy path, error cases, edge cases.

- [ ] **`edit` changes a scalar field and bumps `updated_at`.** `brag
  project edit <p> --status paused` exits 0, prints `Edited project
  "<p>".` to **stderr** (stdout empty); afterward `show` reports `Status:
  paused`, and the project's `updated_at` is strictly later than it was
  before the edit (Store-level; observable without a sleep — see §9).
- [ ] **`edit --name` renames; entries keep their string (DEC-017).**
  `edit <p> --name <q>` makes `show <q>` succeed and `show <p>` fail; an
  existing brag entry whose `project` was `<p>` **still reads `<p>`**
  (no entry string is rewritten — the accepted DEC-017 rename tradeoff).
- [ ] **`edit` is additive across unspecified fields.** Editing only
  `--state-note` leaves `name` and `status` unchanged; editing only
  `--status` leaves `name` and `state_note` unchanged.
- [ ] **`edit` requires at least one field flag.** `edit <p>` with no
  `--name`/`--status`/`--state-note` is a user error (exit 1); nothing
  changes.
- [ ] **`edit --status` validates the enum.** A value outside
  `active|paused|done|archived` is a user error (exit 1) naming the
  accepted set; the row is unchanged (`ErrInvalidStatus` mapped).
- [ ] **`edit --name` to an existing *other* name is a clean user
  error.** Renaming `b` to `a` when `a` already exists exits 1
  (`ErrProjectExists` mapped); `b` is unchanged. Renaming a project to
  **its own current name** is allowed (no-op, self-excluded from the
  uniqueness check).
- [ ] **`archive` flips status non-destructively and is recoverable.**
  `brag project archive <p>` exits 0, prints `Archived project "<p>".`
  to stderr; `show` reports `Status: archived`, and the `state_note` and
  **all locations are preserved**. `edit <p> --status active` restores
  it (archive is recoverable; delete is not).
- [ ] **`delete` removes the project and its locations (DEC-018).**
  `brag project delete <p> --yes` exits 0, prints `Deleted project
  "<p>".` to stderr; afterward `show <p>` is a user error (gone), and the
  path that was registered to `<p>` can be **re-registered to a new
  project** (the `project_locations` rows were deleted in-tx — FK is OFF,
  so no cascade; the manual delete is what frees the path).
- [ ] **`delete` leaves entries' `project` string untouched (DEC-017/
  DEC-018).** After deleting project `<p>`, a brag entry captured with
  `--project <p>` still exists and `brag list --project <p>` still
  returns it. Delete's blast radius on `entries` is **none**.
- [ ] **`delete` requires confirmation unless `--yes`/`-y`.** Without
  `--yes`, `delete` prompts `… [y/N]` on stderr and reads stdin; `y`/`Y`
  proceeds; anything else prints `Aborted.` to stderr, exits 0, and
  **does not delete** (matches `brag delete` for entries).
- [ ] **Unknown project → clean user error on all three.** `edit` /
  `archive` / `delete` against a name-and-id miss each exit 1
  (`ErrNotFound` mapped to `UserErrorf`).
- [ ] **`<name|id>` resolution mirrors `show`.** All three resolve the
  positional arg as a **name first**, then as a positive-integer **id**
  if no project is named that; a project literally named an integer is
  reached by name (name precedence), same as SPEC-028 LD2.
- [ ] **stdout/stderr discipline.** Every confirmation
  (`Edited.`/`Archived.`/`Deleted.`/`Aborted.`) and every error goes to
  **stderr**; these commands write **nothing to stdout** (a success test
  asserts `outBuf` empty per §9).
- [ ] **No SQL under `internal/cli/`**; no new migration; `go test
  ./...`, `gofmt -l .`, `go vet ./...`, `CGO_ENABLED=0 go build ./...`
  clean; `brag project --help` lists `edit`/`archive`/`delete` in the
  **real binary** (registration confirmed at build — see Notes).

## Failing Tests

Written at **design**; build makes them pass. All are **new/additive**
(the premise audit found zero inversions). Storage tests use `t.TempDir()`
(`storage-tests-use-tempdir`) via `newTestStore(t)`; CLI tests use
`newProjectTestRoot` / `runProjectCmd` and out/err buffer separation.

### `internal/storage/project_test.go` (modify — additive)

- `"TestUpdateProject_EditsScalarFieldsAndBumpsUpdatedAt"` — create a
  project; capture `before := GetProject(id).UpdatedAt`; **backdate** the
  row's `created_at` *and* `updated_at` to a fixed past instant (e.g.
  `2020-01-01T00:00:00Z`) via a second `sql.Open(path)` handle (the
  pattern `store_test.go` already uses to inspect `schema_migrations`);
  `UpdateProject(id, Project{Name:…, Status:"paused", StateNote:"x"})`;
  assert the returned/`GetProject` row has the new `Status`/`StateNote`
  and `UpdatedAt.After(CreatedAt)` (now `2026…` > the backdated
  `2020…`) — the bump, asserted **without a sleep** (§9: `now > backdated`
  is deterministic).
- `"TestUpdateProject_BumpReordersListProjects"` — create `p1`,`p2`,`p3`;
  backdate their `created_at`/`updated_at` to **distinct descending**
  past instants so `ListProjects` order is fixed `p1,p2,p3` (or whichever
  ordering the backdate implies); `UpdateProject` the **oldest**; assert
  it now sorts **first** in `ListProjects` (its bumped `updated_at` =
  now beats every backdated value). Locks the recency-reordering the bump
  enables — the case SPEC-027 explicitly deferred (no sleep).
- `"TestUpdateProject_RenameDoesNotRewriteEntries"` — `Add(Entry{Project:
  "old"})`; `CreateProject(Project{Name:"old"})`; `UpdateProject` rename
  `old → new`; assert `List(ListFilter{Project:"old"})` still returns the
  entry and `List(ListFilter{Project:"new"})` returns none — the DEC-017
  rename tradeoff (entries keep their captured string).
- `"TestUpdateProject_DuplicateNameErrProjectExists"` — create `a` and
  `b`; `UpdateProject(b.ID, Project{Name:"a", …})` returns
  `errors.Is(err, ErrProjectExists)`; `GetProject(b.ID).Name == "b"`
  (unchanged).
- `"TestUpdateProject_SameNameSelfRenameAllowed"` — create `a` (note `n`);
  `UpdateProject(a.ID, Project{Name:"a", StateNote:"n2"})` succeeds (the
  uniqueness check self-excludes via `id != ?`); `StateNote` is now `n2`.
- `"TestUpdateProject_InvalidStatusErrInvalidStatus"` —
  `UpdateProject(id, Project{Name:…, Status:"bogus"})` returns
  `errors.Is(err, ErrInvalidStatus)`; the row's status is unchanged.
- `"TestUpdateProject_NotFound"` — `UpdateProject(99999, Project{Name:
  "fresh-unused", Status:"active"})` returns `errors.Is(err, ErrNotFound)`
  (the name is unused, so the uniqueness pre-check passes and the
  zero-rows-affected `UPDATE` yields `ErrNotFound`).
- `"TestArchiveProject_FlipsStatusNonDestructive"` — create an `active`
  project with `StateNote:"keep"` and two locations; `ArchiveProject(id)`;
  `GetProject(id)` has `Status == "archived"`, `StateNote == "keep"`, and
  **both locations still present** (non-destructive). Backdate first and
  assert `UpdatedAt.After(CreatedAt)` (the bump).
- `"TestArchiveProject_RecoverableViaUpdate"` — archive, then
  `UpdateProject(id, Project{Name:…, Status:"active"})`; `GetProject`
  status is `active` again (archive is recoverable — the archive/delete
  distinction).
- `"TestArchiveProject_NotFound"` — `ArchiveProject(99999)` returns
  `errors.Is(err, ErrNotFound)`.
- `"TestDeleteProject_RemovesProjectAndLocations"` — create a project +
  two locations; `DeleteProject(id)`; `GetProject(id)` →
  `errors.Is(err, ErrNotFound)`; `ListProjects()` is empty; a second
  `sql.Open(path)` handle confirms `SELECT COUNT(*) FROM
  project_locations WHERE project_id = id` is `0` (manual in-tx delete —
  FK is OFF, no cascade).
- `"TestDeleteProject_FreesPathForReuse"` — create `a` with path `/p`;
  `DeleteProject(a.ID)`; `CreateProject(b)` + `AddLocation(b.ID, "/p")`
  succeeds (no `ErrLocationExists` — the path row was freed).
- `"TestDeleteProject_LeavesEntriesUntouched"` — `Add(Entry{Project:
  "bragfile"})`; `CreateProject(Project{Name:"bragfile"})`;
  `DeleteProject(id)`; assert `List(ListFilter{Project:"bragfile"})`
  still returns the entry (DEC-017/DEC-018: entries blast radius = none).
- `"TestDeleteProject_RemovesProjectTaggings"` — create a project;
  via a second `sql.Open(path)` handle insert a `'project'` tagging row
  (`INSERT INTO tags(name) …` then `INSERT INTO taggings(tag_id,
  taggable_type, taggable_id, position) VALUES (?, 'project', id, 0)`);
  `DeleteProject(id)`; confirm `SELECT COUNT(*) FROM taggings WHERE
  taggable_type='project' AND taggable_id=id` is `0` (the forward-proof
  in-tx cleanup, even though no command writes `'project'` taggings yet).
- `"TestDeleteProject_NotFound"` — `DeleteProject(99999)` returns
  `errors.Is(err, ErrNotFound)`.

### `internal/cli/project_test.go` (modify — additive)

- `"TestProjectEdit_ChangesStatusAndConfirms"` — `new bragfile --path
  /tmp/x`; `edit bragfile --status paused` exits 0, stderr contains
  `Edited project "bragfile".`, **stdout empty**; then `show bragfile`
  contains `Status: paused`. (stdout/stderr separation asserted.)
- `"TestProjectEdit_SetsStateNote"` — `edit bragfile --state-note "next:
  cut v0.2.0"`; `show` contains `State note: next: cut v0.2.0`.
- `"TestProjectEdit_Rename"` — `edit bragfile --name brag-cli`; `show
  brag-cli` succeeds (`Name: brag-cli`); `show bragfile` →
  `errors.Is(err, ErrUser)`.
- `"TestProjectEdit_NoFlagsErrUser"` — `edit bragfile` (no field flags) →
  `errors.Is(err, ErrUser)`.
- `"TestProjectEdit_InvalidStatusErrUser"` — `edit bragfile --status
  bogus` → `errors.Is(err, ErrUser)`; message mentions the accepted set;
  `show` still `Status: active`.
- `"TestProjectEdit_DuplicateNameErrUser"` — create `a` and `b`; `edit b
  --name a` → `errors.Is(err, ErrUser)`; `show b` still `Name: b`.
- `"TestProjectEdit_UnknownProjectErrUser"` — `edit nope --status paused`
  → `errors.Is(err, ErrUser)`.
- `"TestProjectArchive_FlipsStatusAndRecoverable"` — `new`; `archive
  bragfile` exits 0, stderr `Archived project "bragfile".`, stdout empty;
  `show` → `Status: archived`; then `edit bragfile --status active`;
  `show` → `Status: active` (recoverable).
- `"TestProjectArchive_UnknownProjectErrUser"` — `archive nope` →
  `errors.Is(err, ErrUser)`.
- `"TestProjectDelete_RemovesAndConfirms"` — `new bragfile --path /tmp/x`;
  `delete bragfile --yes` exits 0, stderr `Deleted project "bragfile".`,
  stdout empty; `show bragfile` → `errors.Is(err, ErrUser)`.
- `"TestProjectDelete_PromptConfirmsWithY"` — `new`; run `delete bragfile`
  with `root.SetIn(strings.NewReader("y\n"))`; project is deleted, exit 0.
- `"TestProjectDelete_PromptDeclineAborts"` — `new`; run `delete bragfile`
  with stdin `"n\n"`; stderr contains `Aborted.`, exit 0, and `show
  bragfile` still succeeds (not deleted).
- `"TestProjectDelete_FreesPathForReuse"` — `new a --path /p`; `delete a
  --yes`; `new b --path /p` succeeds (exit 0, `Created project "b".`) —
  the location row was freed (DEC-018 manual-cascade, CLI-observable).
- `"TestProjectDelete_LeavesEntryProjectString"` — `seedListEntry(t,
  dbPath, "did a thing", "", "bragfile", "feature")`; `new bragfile
  --path /p`; `delete bragfile --yes`; then `brag list --project
  bragfile` (via the list command on the same db) still returns the entry
  (DEC-017/DEC-018 entries-untouched, CLI-observable).
- `"TestProjectDelete_UnknownProjectErrUser"` — `delete nope --yes` →
  `errors.Is(err, ErrUser)`.
- `"TestProjectMutations_HelpShowsExamples"` — `edit --help`, `archive
  --help`, and `delete --help` each contain `Examples:` and a distinctive
  token unique to their locked `Long` (`brag project edit`, `--status
  active`, `brag project delete`); `archive --help` contains
  `recoverable` (or `edit … --status active`) and `delete --help`
  contains `irreversible`/`cannot be undone` — the archive-vs-delete
  distinction is in the help text. (Positive substring asserts; the §12
  NOT-contains self-audit is N/A.)

> **Locked-decision ↔ test traceability (§9).** Each locked decision has
> a paired failing test: **DEC-018 / LD-DELETE** (blast radius) →
> `TestDeleteProject_RemovesProjectAndLocations` +
> `TestDeleteProject_FreesPathForReuse` +
> `TestDeleteProject_LeavesEntriesUntouched` +
> `TestDeleteProject_RemovesProjectTaggings` +
> `TestProjectDelete_FreesPathForReuse` +
> `TestProjectDelete_LeavesEntryProjectString`; **LD-EDIT-SCALAR**
> (scalar fields, rename-doesn't-rewrite, enum validation, uniqueness) →
> the `TestUpdateProject_*` + `TestProjectEdit_*` families;
> **LD-ARCHIVE** (non-destructive, recoverable) →
> `TestArchiveProject_FlipsStatusNonDestructive` +
> `TestArchiveProject_RecoverableViaUpdate` +
> `TestProjectArchive_FlipsStatusAndRecoverable`; **LD-CONFIRM** (y/N +
> `--yes`) → `TestProjectDelete_Prompt*`; **LD-BUMP** (`updated_at`
> advances + reorders) → `TestUpdateProject_EditsScalarFieldsAndBumpsUpdatedAt`
> + `TestUpdateProject_BumpReordersListProjects`; **LD-STDERR**
> (confirmation→stderr) → the stdout-empty assertions in the
> `*_ChangesStatusAndConfirms` / `*_RemovesAndConfirms` tests.

> **§9 no-sleep note for build.** Every `updated_at`-bump and reordering
> assertion is made deterministic by **backdating** rows to a fixed past
> instant (`2020…`) via a second `sql.Open(path)` handle, then asserting
> the post-mutation `updated_at` (= `now`, `2026…`) beats it. Do **not**
> add a `time.Sleep` to separate timestamps — RFC3339 is second-precision
> and the backdate makes the inequality robust regardless of test speed.

## Implementation Context

*Read this section and the files it points to before build. This is the
whole handoff — the build session has only this spec.*

### Decisions that apply

- **`DEC-018` (emitted by this spec)** — `brag project delete`'s blast
  radius: **entries untouched** (DEC-017 soft match → an entry keeps its
  free-text project string; cascade = none), **`project_locations`
  deleted manually in-tx** (FK enforcement is OFF — `PRAGMA
  foreign_keys` is never set ON — so the `REFERENCES projects(id)` clause
  does **not** cascade; verified at SPEC-027), and **`'project'` taggings
  deleted in-tx** (none are written yet — schema-ready per DEC-015 — but
  the cleanup ships now so a future project-tag surface needs no delete
  change). Also records the **archive-vs-delete** contract: `archive` is
  a recoverable status flip, `delete` is irreversible. Read it before
  build — it is the why behind `DeleteProject`'s three DML statements.
- **`DEC-017`** (shipped, SPEC-027) — soft string match. The mutations
  honor it: `UpdateProject` rename changes only `projects.name` and
  **never** rewrites `entries.project`; `DeleteProject` leaves `entries`
  alone. The `status` enum (`active|paused|done|archived`, default
  `active`, Store-validated, no DB CHECK) is the set `UpdateProject`
  validates against and `archive` sets `archived` within. The single
  free-text `state_note` is the column `edit --state-note` writes.
- **`DEC-007`** — required/positional/flag validation lives **inline in
  `RunE`** via `UserErrorf` (cobra's built-in validators return
  unwrappable plain errors that exit 2). Arg count, empty-key, the
  "at-least-one-field" rule, and the `--status` enum surface are all
  checked in `RunE` (the enum is validated in the Store and the sentinel
  mapped). Mirror `delete.go` / `project.go`'s existing RunEs.
- **`DEC-006`** — cobra: each subcommand is a `*cobra.Command` from a
  `new…Cmd()` constructor, attached to the already-registered
  `NewProjectCmd` parent. **No `cmd/brag/main.go` change** — the parent
  is registered (SPEC-028); the new subcommands ride it.
- **`DEC-005`** — INTEGER autoincrement PKs give the `id`-DESC tie-break
  `ListProjects` orders on after `updated_at DESC`.
- **`DEC-011`** — relevant only as the family these commands sit beside.
  `edit`/`archive`/`delete` are **confirmation-only** (no `--format`, no
  data to stdout); they do **not** emit DEC-011 JSON. (Reading a project
  as JSON is `show`; SPEC-028.)

### Locked design decisions

The design questions STAGE-029 was asked to resolve, decided here with
reasoning. Confidence on each is ≥ 0.8, so no `/guidance/questions.yaml`
entry is filed (§14). Only the delete blast radius is durable+cross-spec
enough to warrant a DEC (**DEC-018**); the rest are localized LDs in the
SPEC-028 style.

- **LD-DELETE — delete blast radius (→ DEC-018).** Confidence 0.85.
  `DeleteProject` removes, in one transaction and **in this order**:
  (1) `'project'` taggings, (2) `project_locations` rows, (3) the
  `projects` row. **`entries` are not touched** — under DEC-017 an entry
  owns its free-text project string and a project's deletion must not
  rewrite history (the cleanest possible blast radius on entries: none).
  `project_locations` **must** be deleted manually because FK enforcement
  is OFF (SPEC-027 forward note) — the `REFERENCES` clause is decorative,
  so without the explicit `DELETE` the location rows would dangle (and
  worse, keep the path globally reserved, blocking re-registration). The
  `'project'`-taggings delete is a **no-op today** (nothing writes them)
  but is laid down now so the eventual `brag project tag` surface
  (STAGE-008/PROJ-003) inherits correct cleanup for free — mirroring how
  `Store.Delete` already deletes an entry's taggings first. *Rejected
  alternatives:* (a) rely on the `REFERENCES` cascade — **wrong**, FK is
  OFF, nothing cascades; (b) `SET NULL`/cascade onto `entries` — rejected,
  it would rewrite captured history, violating DEC-017; (c) skip the
  `'project'`-taggings delete until the tag surface exists — rejected as a
  latent orphan-row bug deferred into a future spec for no saving (the
  statement is one line and harmless now). This is durable (SPEC-030's
  recent-brag count and any future project-tag spec depend on the
  contract) and cross-spec — hence **DEC-018**, not a localized LD.

- **LD-EDIT-SCALAR — `edit` covers scalar fields only; location editing
  is peeled to SPEC-033.** Confidence 0.85. Editable: `--name` (rename),
  `--status` (enum-validated), `--state-note`. **Locations are NOT
  edited here.** Reason: `edit` + `archive` + `delete` with location
  editing folded in reads **L** (a new `RemoveLocation` Store method,
  repeatable `--add-path`/`--remove-path` multi-operation semantics, and
  their tests) — exactly the L the stage told this spec to watch for. The
  prescribed split keeps `edit` to the three scalar columns and peels
  locations into **SPEC-033**. `edit` requires **at least one** field
  flag (a no-flag `edit` is a user error — nothing to do); unspecified
  fields are preserved by reading the current project and overriding only
  `cmd.Flags().Changed(...)` fields. Rename honors DEC-017 (the
  `projects.name` changes; `entries.project` strings do **not**), and the
  name-uniqueness check **self-excludes** (`id != ?`) so renaming to the
  current name is a legal no-op. `--status` is validated against the
  DEC-017 enum **in the Store** (`ErrInvalidStatus`), keeping the enum a
  single source of truth (DEC-017: "validated in the Store"); the CLI
  maps the sentinel to a `UserErrorf` naming the accepted set. *Rejected
  alternative:* absorb `--add-path`/`--remove-path` here — rejected as
  the L the stage flagged; peeled to SPEC-033 (flagged in Outputs +
  backlog, not silently dropped).

- **LD-ARCHIVE — `archive` is a non-destructive status flip, recoverable
  via `edit --status active`.** Confidence 0.88. `ArchiveProject` sets
  `status='archived'` and bumps `updated_at`; it touches **nothing
  else** (name, state_note, locations all preserved), so it is fully
  reversible. No confirmation prompt (it destroys nothing). Help text
  states the recovery command explicitly and contrasts with `delete`.
  *Rejected alternative:* make `archive` a flavor of `delete` (soft
  delete with a hidden flag) — rejected; the stage wants `archive` and
  `delete` "clearly distinct," and a plain status value the user can flip
  back is the simplest recoverable primitive.

- **LD-CONFIRM — `delete` matches `brag delete`'s y/N + `--yes`/`-y`
  precedent.** Confidence 0.90. `delete` prompts `Delete project "<p>"
  and its locations? This cannot be undone. [y/N] ` on stderr and reads
  stdin; `y`/`Y` proceeds, anything else → `Aborted.` (exit 0);
  `--yes`/`-y` skips the prompt. This is `delete.go` verbatim in shape —
  the established, consistent guard for a destructive command. *Rejected
  alternatives:* unguarded delete (rejected — destructive, removes
  locations; inconsistent with `brag delete`); typed-name confirmation
  (rejected — heavier than the precedent, no other command does it).

- **LD-BUMP — every mutation bumps `updated_at`.** Confidence 0.9.
  `UpdateProject`, `ArchiveProject` set `updated_at = now` (UTC, RFC3339,
  second-truncated — `timestamps-in-utc-rfc3339`). This is the first code
  to advance `updated_at` past `created_at`, realizing `ListProjects`'
  recency ordering. SPEC-027 deferred the distinct-`updated_at` ordering
  test for exactly this reason; it is added here (no sleep — see §9).

- **LD-STDERR — confirmation-to-stderr, nothing to stdout.** Confidence
  0.9. `Edited project "<p>".` / `Archived project "<p>".` / `Deleted
  project "<p>".` / `Aborted.` all go to **stderr** (mirroring
  `delete.go` and SPEC-028's `Created project …`); these commands emit
  **no stdout**. No `--format` flag (they are not data-emitting). *Per
  the STAGE-006 flag-default WATCH item, each flag's default is stated:*
  `--name`/`--status`/`--state-note` default `""` (detected via
  `Changed`, not value); `--yes`/`-y` defaults `false`; **no `--format`
  on any of the three.** (This is the THIRD CLI spec with flags —
  SPEC-026, SPEC-028, SPEC-029 — see the WATCH note in Notes.)

### Constraints that apply

(see `/guidance/constraints.yaml` for full text)

- `no-sql-in-cli-layer` (**blocking**) — `project.go` calls Store methods
  only; the `bufio`/`strconv`/`strings` it adds are not SQL. All
  persistence (resolution, mutation) is a Store call. The delete-prompt
  read is `bufio.NewReader(cmd.InOrStdin())` (no DB).
- `stdout-is-for-data-stderr-is-for-humans` (**blocking**) — all
  confirmations + errors → **stderr**; **no stdout** from these commands.
  A success test asserts `outBuf` empty (§9).
- `storage-tests-use-tempdir` (**blocking**) — every new storage test
  uses `newTestStore(t)` / `t.TempDir()`; second raw `sql.Open` handles
  point at the same temp path; never `~/.bragfile`.
- `timestamps-in-utc-rfc3339` (**blocking**) — `updated_at` written as
  `time.Now().UTC().Truncate(time.Second).Format(time.RFC3339)`, exactly
  like `Store.Update`.
- `errors-wrap-with-context` (warning) — `fmt.Errorf("update project %d:
  %w", id, err)` etc., matching the existing method style.
- `test-before-implementation` (blocking) — the Failing Tests above are
  the design deliverable.
- `one-spec-per-pr` (blocking) — the PR references SPEC-029 only.
- `no-cgo` / `no-new-top-level-deps-without-decision` — pure-Go path; no
  new dependency (`bufio` is stdlib, already used by `delete.go`).

### Design-time verification (run at design 2026-06-09)

- **Blast-radius SQL** — `DeleteProject`'s three DML statements were
  reasoned against the live schema: `taggings` (taggable polymorphic, has
  a `(taggable_type, taggable_id)` shape — `0003`), `project_locations`
  (`project_id` column, no ON DELETE behavior because FK is OFF — `0004`),
  `projects`. The `entries` table is **not** referenced. The path-reuse
  test (`FreesPathForReuse`) is the observable proof the location row is
  actually gone (the `UNIQUE(path)` would otherwise reject re-registration).
- **§9 no-sleep determinism** — confirmed the backdate approach: a second
  `sql.Open(path)` handle runs `UPDATE projects SET created_at=?,
  updated_at=? WHERE id=?` with a fixed `2020…` instant; the subsequent
  mutation stamps `now` (`2026…`); `After`/ordering inequalities are
  robust without any sleep. (Same family as `storagetest.Backdate` for
  entries, kept inline here since storage-package tests may touch raw SQL,
  mirroring `store_test.go`'s second-handle `schema_migrations` reads.)
- **Enum validation placement** — `UpdateProject` validates `Status`
  against `validProjectStatuses` and returns `ErrInvalidStatus`; confirmed
  `CreateProject` is **not** retrofitted (no inversion), since `brag
  project new` never supplies a non-default status.

### Dev/prod DB isolation (PROJ-002 brief) — still mandatory this stage

The schema is at v0.2.x (post-`0004`) from SPEC-027. Run the dev binary
against a **dev DB** (`BRAGFILE_DB=~/.bragfile-dev/db.sqlite` or `--db`,
via `just install`). **Never open the production `~/.bragfile/db.sqlite`**
with a v0.2.x binary. SPEC-029 adds no migration, but the rule stands.
All tests use `t.TempDir()` regardless.

### Prior related work

- `SPEC-028` (shipped, PR #41, `85c5173`) — the `brag project` parent +
  `new`/`list`/`show`; the open/resolve pattern, the name-first/id-fallback
  resolution (LD2) this spec factors into `resolveProjectByNameOrID`, the
  `renderProjectPlain` used by the verify-via-`show` tests, and the
  confirmation-to-stderr discipline. Its ship reflection flagged the
  flag-default WATCH at **N=2**; SPEC-029 is the natural **N=3** (see
  Notes).
- `SPEC-027` (shipped, PR #40, `7a67834`) — the schema + Store primitives
  + DEC-017. **Its ship reflection's forward note for SPEC-029 is the
  load-bearing input:** FK enforcement is OFF, so `DeleteProject` must
  manually delete `project_locations` rows (no cascade) and define its
  `'project'`-taggings blast radius. That note is realized here as DEC-018.
- `internal/storage/store.go` `Update`/`Delete`/`RenameTag` — the exact
  transactional shapes `UpdateProject`/`DeleteProject` mirror (incl.
  `RowsAffected()==0 → ErrNotFound` and the in-tx name-uniqueness
  pre-check).

### Out of scope (for this spec specifically)

If any of these feels necessary during build, **stop and flag** — do not
expand this spec.

- **Location editing (`--add-path` / `--remove-path` + a `RemoveLocation`
  Store method)** → **SPEC-033** (peeled to hold SPEC-029 at M; the L the
  stage told this spec to watch for). `edit` here is scalar-only.
- **The `brag project status` dashboard + recent-brag count** →
  **SPEC-030** (the `entries.project = projects.name` join under DEC-017).
- **The `here` cwd resolver + path normalization** → **SPEC-031**.
- **`brag add` `--project` auto-fill** → **SPEC-032**.
- **Writing `'project'` taggings / a `brag project tag` surface** →
  schema-ready only (DEC-015); STAGE-008/PROJ-003 candidate. `DeleteProject`
  *cleans up* `'project'` taggings defensively but **nothing writes them**.
- **Retrofitting `CreateProject` with enum validation** — out of scope
  (no inversion); `new` never supplies a non-default status. If a future
  spec lets `new` set status, fold the validation in then.
- **`--format`/JSON output on `edit`/`archive`/`delete`** — they are
  confirmation-only; reading a project as data is `show` (SPEC-028).
- **Any migration or schema change** — there is none.

## Notes for the Implementer

### `storage` — the sentinel (`internal/storage/errors.go`)

Add beside the existing project sentinels:

```go
// ErrInvalidStatus is returned (wrapped) by UpdateProject when the given
// status is not one of the DEC-017 enum values (active|paused|done|
// archived). Validated in the Store (not a DB CHECK) so adding a value
// later is an additive change.
var ErrInvalidStatus = errors.New("invalid project status")
```

### `storage` — the three mutation methods (`internal/storage/project.go`)

Add beside the SPEC-027/028 methods. Mirror `store.go`'s `Update`/`Delete`
exactly (`context.Background()`, `BeginTx` + `defer tx.Rollback()`,
`time.Now().UTC().Truncate(time.Second)`, `RowsAffected()==0 →
ErrNotFound`, `fmt.Errorf("…: %w", err)`). `project.go` already imports
`context`, `database/sql`, `errors`, `fmt`, `time` — no new imports.

```go
// validProjectStatuses is the DEC-017 status enum. Validated in the Store
// (not a DB CHECK) so adding a value later is an additive change under the
// forward-only migration regime (DEC-002), mirroring entries.type's
// free-text column.
var validProjectStatuses = map[string]bool{
	"active": true, "paused": true, "done": true, "archived": true,
}

// UpdateProject replaces the editable scalar fields (Name, Status,
// StateNote) on the project with id and bumps updated_at. Locations are
// NOT edited here (SPEC-033). Returns the hydrated project (id and
// created_at preserved). Errors:
//   - ErrInvalidStatus if Status is not a DEC-017 enum value
//   - ErrProjectExists if Name is already taken by a *different* project
//   - ErrNotFound if no row matches id
// This is the first Store method to advance updated_at past created_at,
// making ListProjects' recency ordering observable.
func (s *Store) UpdateProject(id int64, p Project) (Project, error) {
	if p.Status == "" {
		p.Status = "active"
	}
	if !validProjectStatuses[p.Status] {
		return Project{}, fmt.Errorf("update project %d: status %q: %w", id, p.Status, ErrInvalidStatus)
	}
	now := time.Now().UTC().Truncate(time.Second)

	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Project{}, fmt.Errorf("update project %d: %w", id, err)
	}
	defer tx.Rollback()

	// Name is UNIQUE; a rename collides only with a *different* project.
	// Self-exclude (id != ?) so renaming to the current name is a no-op.
	var exists int
	if err := tx.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM projects WHERE name = ? AND id != ?`, p.Name, id,
	).Scan(&exists); err != nil {
		return Project{}, fmt.Errorf("update project %d: %w", id, err)
	}
	if exists > 0 {
		return Project{}, fmt.Errorf("update project %d name %q: %w", id, p.Name, ErrProjectExists)
	}

	res, err := tx.ExecContext(ctx,
		`UPDATE projects SET name = ?, status = ?, state_note = ?, updated_at = ?
		 WHERE id = ?`,
		p.Name, p.Status, p.StateNote, now.Format(time.RFC3339), id,
	)
	if err != nil {
		return Project{}, fmt.Errorf("update project %d: %w", id, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return Project{}, fmt.Errorf("update project %d: %w", id, err)
	}
	if n == 0 {
		return Project{}, fmt.Errorf("update project %d: %w", id, ErrNotFound)
	}

	if err := tx.Commit(); err != nil {
		return Project{}, fmt.Errorf("update project %d: %w", id, err)
	}
	return s.GetProject(id)
}

// ArchiveProject sets the project's status to "archived" (the DEC-017
// non-destructive flip) and bumps updated_at. Name, state_note, and
// locations are preserved — archive is recoverable via
// UpdateProject(..., Status:"active"). Returns ErrNotFound if no row matches.
func (s *Store) ArchiveProject(id int64) error {
	now := time.Now().UTC().Truncate(time.Second)
	res, err := s.db.ExecContext(context.Background(),
		`UPDATE projects SET status = 'archived', updated_at = ? WHERE id = ?`,
		now.Format(time.RFC3339), id,
	)
	if err != nil {
		return fmt.Errorf("archive project %d: %w", id, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("archive project %d: %w", id, err)
	}
	if n == 0 {
		return fmt.Errorf("archive project %d: %w", id, ErrNotFound)
	}
	return nil
}

// DeleteProject permanently removes the project with id and its full blast
// radius (DEC-018), all in one transaction and in this order:
//   1. any 'project' taggings (none are written yet — schema-ready per
//      DEC-015 — but cleaned now so a future project-tag surface needs no
//      delete change)
//   2. the project's project_locations rows — FK enforcement is OFF in
//      bragfile, so the REFERENCES clause does NOT cascade; deleting these
//      manually is also what frees the globally-UNIQUE path for reuse
//   3. the projects row itself
// entries are deliberately UNTOUCHED (DEC-017 soft string match): an entry
// keeps the free-text project string it was captured with. Returns
// ErrNotFound if no row matches id.
func (s *Store) DeleteProject(id int64) error {
	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("delete project %d: %w", id, err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx,
		`DELETE FROM taggings WHERE taggable_type = 'project' AND taggable_id = ?`, id,
	); err != nil {
		return fmt.Errorf("delete project %d: remove taggings: %w", id, err)
	}
	if _, err := tx.ExecContext(ctx,
		`DELETE FROM project_locations WHERE project_id = ?`, id,
	); err != nil {
		return fmt.Errorf("delete project %d: remove locations: %w", id, err)
	}

	res, err := tx.ExecContext(ctx, `DELETE FROM projects WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete project %d: %w", id, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete project %d: %w", id, err)
	}
	if n == 0 {
		return fmt.Errorf("delete project %d: %w", id, ErrNotFound)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("delete project %d: %w", id, err)
	}
	return nil
}
```

### `cli` — the resolution helper + three subcommands (`internal/cli/project.go`)

Add `bufio` to the imports (already used by `delete.go`; the rest —
`errors`, `fmt`, `io`, `strconv`, `strings`, `config`, `export`,
`storage`, `cobra` — are present). Register the three on the existing
parent and widen its `Short`:

```go
func NewProjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Manage projects (new, list, show, edit, archive, delete)",
	}
	cmd.AddCommand(newProjectNewCmd())
	cmd.AddCommand(newProjectListCmd())
	cmd.AddCommand(newProjectShowCmd())
	cmd.AddCommand(newProjectEditCmd())
	cmd.AddCommand(newProjectArchiveCmd())
	cmd.AddCommand(newProjectDeleteCmd())
	return cmd
}
```

Shared resolver (mirrors `runProjectShow`'s LD2 logic, factored so
`edit`/`archive`/`delete` resolve identically — name first, then a
positive-integer id):

```go
// resolveProjectByNameOrID resolves key as a project name first, then —
// if no project has that name and key is a positive integer — as a
// project id (mirroring `brag project show`, SPEC-028 LD2). Returns the
// resolved project, or an error wrapping storage.ErrNotFound on a miss.
func resolveProjectByNameOrID(s *storage.Store, key string) (storage.Project, error) {
	project, err := s.GetProjectByName(key)
	if err == nil {
		return project, nil
	}
	if !errors.Is(err, storage.ErrNotFound) {
		return storage.Project{}, err
	}
	id, convErr := strconv.ParseInt(key, 10, 64)
	if convErr != nil || id <= 0 {
		return storage.Project{}, fmt.Errorf("resolve project %q: %w", key, storage.ErrNotFound)
	}
	return s.GetProject(id)
}
```

`edit` — scalar fields, `Changed`-gated, at-least-one required:

```go
func newProjectEditCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit <name|id>",
		Short: "Edit a project's name, status, or state note",
		Long: `Edit a project's scalar fields. The project is resolved by name first, then
by id. Pass at least one of --name, --status, or --state-note; unspecified
fields are left unchanged.

Renaming a project does NOT rewrite the project string on existing brag entries
— they keep what they were captured with (DEC-017). Editing locations is a
separate command (a later STAGE-007 spec); this edits scalar fields only.

Examples:
  brag project edit bragfile --status paused
  brag project edit bragfile --state-note "shipped tags; next: cut v0.2.0"
  brag project edit bragfile --name brag-cli`,
		RunE: runProjectEdit,
	}
	cmd.Flags().String("name", "", "new project name (rename; must be unique)")
	cmd.Flags().String("status", "", "new status (one of: active, paused, done, archived)")
	cmd.Flags().String("state-note", "", "new state/next-action note")
	return cmd
}

func runProjectEdit(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return UserErrorf("edit requires exactly one <name|id> argument")
	}
	key := strings.TrimSpace(args[0])
	if key == "" {
		return UserErrorf("project name or id must not be empty")
	}

	nameChanged := cmd.Flags().Changed("name")
	statusChanged := cmd.Flags().Changed("status")
	noteChanged := cmd.Flags().Changed("state-note")
	if !nameChanged && !statusChanged && !noteChanged {
		return UserErrorf("edit requires at least one of --name, --status, --state-note")
	}

	dbFlag := getFlagString(cmd, "db")
	dbPath, err := config.ResolveDBPath(dbFlag)
	if err != nil {
		return fmt.Errorf("resolve db path: %w", err)
	}
	s, err := storage.Open(dbPath)
	if err != nil {
		return fmt.Errorf("open store: %w", err)
	}
	defer s.Close()

	current, err := resolveProjectByNameOrID(s, key)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return UserErrorf("no project named or with id %q", key)
		}
		return fmt.Errorf("resolve project: %w", err)
	}

	next := current
	if nameChanged {
		name, _ := cmd.Flags().GetString("name")
		name = strings.TrimSpace(name)
		if name == "" {
			return UserErrorf("--name must not be empty")
		}
		next.Name = name
	}
	if statusChanged {
		status, _ := cmd.Flags().GetString("status")
		next.Status = strings.TrimSpace(status)
	}
	if noteChanged {
		note, _ := cmd.Flags().GetString("state-note")
		next.StateNote = note
	}

	updated, err := s.UpdateProject(current.ID, next)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrProjectExists):
			return UserErrorf("project %q already exists", next.Name)
		case errors.Is(err, storage.ErrInvalidStatus):
			return UserErrorf("invalid status %q (accepted: active, paused, done, archived)", next.Status)
		case errors.Is(err, storage.ErrNotFound):
			return UserErrorf("no project named or with id %q", key)
		}
		return fmt.Errorf("update project: %w", err)
	}
	fmt.Fprintf(cmd.ErrOrStderr(), "Edited project %q.\n", updated.Name)
	return nil
}
```

`archive` — recoverable status flip, no prompt:

```go
func newProjectArchiveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "archive <name|id>",
		Short: "Archive a project (non-destructive status flip; recoverable)",
		Long: `Archive a project by setting its status to "archived". This is a
non-destructive flip: the project, its state note, and its locations are all
preserved, and it can be restored at any time with:

  brag project edit <name|id> --status active

Archive is NOT delete. To permanently remove a project and its locations, use
'brag project delete', which is irreversible.

Examples:
  brag project archive bragfile
  brag project archive 3`,
		RunE: runProjectArchive,
	}
	return cmd
}

func runProjectArchive(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return UserErrorf("archive requires exactly one <name|id> argument")
	}
	key := strings.TrimSpace(args[0])
	if key == "" {
		return UserErrorf("project name or id must not be empty")
	}

	dbFlag := getFlagString(cmd, "db")
	dbPath, err := config.ResolveDBPath(dbFlag)
	if err != nil {
		return fmt.Errorf("resolve db path: %w", err)
	}
	s, err := storage.Open(dbPath)
	if err != nil {
		return fmt.Errorf("open store: %w", err)
	}
	defer s.Close()

	project, err := resolveProjectByNameOrID(s, key)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return UserErrorf("no project named or with id %q", key)
		}
		return fmt.Errorf("resolve project: %w", err)
	}

	if err := s.ArchiveProject(project.ID); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return UserErrorf("no project named or with id %q", key)
		}
		return fmt.Errorf("archive project: %w", err)
	}
	fmt.Fprintf(cmd.ErrOrStderr(), "Archived project %q.\n", project.Name)
	return nil
}
```

`delete` — destructive, y/N + `--yes`/`-y` (mirrors `delete.go`):

```go
func newProjectDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <name|id>",
		Short: "Permanently delete a project (irreversible)",
		Long: `Permanently delete a project and its locations. Prompts for confirmation
unless --yes is passed.

This is IRREVERSIBLE and distinct from 'brag project archive' (a recoverable
status flip). Delete removes the project row and every filesystem location
attached to it. Existing brag entries are NOT touched — an entry keeps the
project string it was captured with (DEC-017), so 'brag list --project <name>'
still finds those entries after the project is deleted.

Examples:
  brag project delete bragfile        # prompts y/N on stdin
  brag project delete bragfile --yes  # skip the prompt
  brag project delete 3 -y            # by id, no prompt`,
		RunE: runProjectDelete,
	}
	cmd.Flags().BoolP("yes", "y", false, "skip confirmation prompt")
	return cmd
}

func runProjectDelete(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return UserErrorf("delete requires exactly one <name|id> argument")
	}
	key := strings.TrimSpace(args[0])
	if key == "" {
		return UserErrorf("project name or id must not be empty")
	}

	dbFlag := getFlagString(cmd, "db")
	dbPath, err := config.ResolveDBPath(dbFlag)
	if err != nil {
		return fmt.Errorf("resolve db path: %w", err)
	}
	s, err := storage.Open(dbPath)
	if err != nil {
		return fmt.Errorf("open store: %w", err)
	}
	defer s.Close()

	project, err := resolveProjectByNameOrID(s, key)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return UserErrorf("no project named or with id %q", key)
		}
		return fmt.Errorf("resolve project: %w", err)
	}

	yes, _ := cmd.Flags().GetBool("yes")
	if !yes {
		fmt.Fprintf(cmd.ErrOrStderr(),
			"Delete project %q and its locations? This cannot be undone. [y/N] ", project.Name)
		reader := bufio.NewReader(cmd.InOrStdin())
		line, _ := reader.ReadString('\n')
		line = strings.TrimRight(line, "\r\n")
		if line != "y" && line != "Y" {
			fmt.Fprintln(cmd.ErrOrStderr(), "Aborted.")
			return nil
		}
	}

	if err := s.DeleteProject(project.ID); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return UserErrorf("no project named or with id %q", key)
		}
		return fmt.Errorf("delete project: %w", err)
	}
	fmt.Fprintf(cmd.ErrOrStderr(), "Deleted project %q.\n", project.Name)
	return nil
}
```

### Registration (and the registration gap)

**No `cmd/brag/main.go` change** — `NewProjectCmd()` is already registered
(SPEC-028). The three subcommands attach inside `NewProjectCmd`. **No test
enumerates the parent's subcommand set** (premise audit case 2), so
registration is uncovered by unit tests — build MUST confirm it in the
**real binary**: `CGO_ENABLED=0 go build -o /tmp/brag-029 ./cmd/brag`,
then `/tmp/brag-029 project --help` lists `edit`/`archive`/`delete`, and
`/tmp/brag-029 project edit --help` / `archive --help` / `delete --help`
render with `Examples:`. State this in Build Completion.

### `docs/api-contract.md` — the status-change UPDATE (literal)

Add these three sections **after** `### brag project show <name|id>`
(ending line 502, the `Exit 1 …` bullet) and **before** `### brag
completion <shell>` (line 503). House style mirrors the adjacent `brag
project *` sections (fenced example + prose + exit/IO bullets):

```
### `brag project edit <name|id>` — edit a project's scalar fields (STAGE-007)

​```
brag project edit bragfile --status paused
brag project edit bragfile --state-note "shipped tags; next: cut v0.2.0"
brag project edit bragfile --name brag-cli
​```

Edits a project's scalar fields. The argument resolves as a **name first**,
then as a positive-integer **id**. Pass at least one of `--name`, `--status`,
or `--state-note`; unspecified fields are unchanged.

- `--name` — rename (must be unique). Renaming does **not** rewrite the project
  string on existing brag entries (DEC-017); they keep their captured string.
- `--status` — one of `active`, `paused`, `done`, `archived` (validated).
- `--state-note` — the free-text state/next-action note.
- Bumps `updated_at` (so the project rises in `brag project list` recency order).
- Exits 0 on success; stderr: `Edited project "<name>".` (stdout empty).
- Exit 1 (user error) if no field flag is given, the project is not found, the
  new name is already taken, or `--status` is outside the enum.

### `brag project archive <name|id>` — archive a project (STAGE-007)

​```
brag project archive bragfile
​```

Sets a project's status to `archived` — a **non-destructive, recoverable**
flip. The project, its state note, and its locations are all preserved. Restore
it with `brag project edit <name|id> --status active`. This is **not** delete.

- Exits 0 on success; stderr: `Archived project "<name>".` (stdout empty).
- Exit 1 (user error) if the project is not found.

### `brag project delete <name|id>` — permanently delete a project (STAGE-007)

​```
brag project delete bragfile        # prompts y/N on stdin
brag project delete bragfile --yes  # skip the prompt
​```

Permanently removes a project and its `project_locations` rows. **Irreversible**
(distinct from `archive`). Prompts for `y/N` confirmation on stderr unless
`--yes`/`-y` is passed; a non-`y` answer prints `Aborted.` and exits 0 without
deleting. Existing brag entries are **not** touched — an entry keeps its
project string (DEC-017), so `brag list --project <name>` still finds those
entries afterward (blast radius on entries: none — DEC-018).

- `--yes`, `-y` — skip the confirmation prompt.
- Exits 0 on success; stderr: `Deleted project "<name>".` (stdout empty).
- Exit 1 (user error) if the project is not found.
```

(The `​` zero-width marks above only escape the nested fences in this spec
— the real doc uses plain triple-backtick fences.)

Then add to the **References** list (after the `DEC-017` row, line 559):

```
- `DEC-018` — `brag project delete` blast radius: entries untouched (soft match), project_locations deleted manually in-tx (FK off → no cascade), `'project'` taggings cleaned in-tx; archive is the recoverable status flip, delete is irreversible.
```

### Flag-default WATCH item (N=3 — note, do NOT codify mid-spec)

This is the **third** CLI spec with flags whose defaults are stated
explicitly (SPEC-026 `--format ""`; SPEC-028 `--format ""` + LD4;
SPEC-029 `--name`/`--status`/`--state-note` default `""` detected via
`Changed`, `--yes`/`-y` default `false`, **no `--format`**). If verify/
ship agrees this is a clean **N=3 same-outcome**, **NOTE it for STAGE-007
close** as a candidate to fold into the literal-artifact-as-spec guidance
(§12) — do **not** codify it mid-spec (the codification meta-rule needs
the documented trigger at a close, not a unilateral mid-stage edit).

### Gotchas

- **No SQL in `project.go`.** Resolution and mutation are Store calls; the
  prompt read is `bufio` on `cmd.InOrStdin()`; id parsing is `strconv`.
  Nothing imports `database/sql`.
- **`edit` preserves unspecified fields** by copying `current` and
  overriding only `Changed` flags — do not pass a zero-valued `Project`
  to `UpdateProject` (it would blank `name`/`state_note`).
- **Self-rename is legal.** The uniqueness check is `name = ? AND id != ?`
  — renaming a project to its own name must not error.
- **`archive` has no prompt; `delete` does.** Archive destroys nothing.
  Delete mirrors `delete.go` exactly: `--yes`/`-y`, decline → `Aborted.`
  on stderr, exit 0.
- **Confirmations on stderr, nothing on stdout.** A success test asserts
  `outBuf` is empty (§9). These commands never write stdout.
- **§9 no sleep.** Backdate rows to a fixed `2020…` instant via a second
  `sql.Open(path)` handle to make every `updated_at`-bump/reorder
  assertion deterministic; never `time.Sleep`.
- **`gofmt -w .` + `go vet ./...`** before the PR; confirm `./brag project
  --help` lists the three new subcommands in the real binary.

---

## Build Completion

*Filled in at the end of the **build** cycle, before advancing to verify.*

- **Branch:**
- **PR (if applicable):**
- **All acceptance criteria met?** yes/no
- **New decisions emitted:**
  - `DEC-018` — `brag project delete` blast radius (emitted at design)
- **Deviations from spec:**
  - [list]
- **Follow-up work identified:**
  - [SPEC-033 already added to the backlog at design]

### Build-phase reflection (3 questions, short answers)

Process-focused: how did the build go? What friction did the spec create?

1. **What was unclear in the spec that slowed you down?**
   — <answer>

2. **Was there a constraint or decision that should have been listed but wasn't?**
   — <answer>

3. **If you did this task again, what would you do differently?**
   — <answer>

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
