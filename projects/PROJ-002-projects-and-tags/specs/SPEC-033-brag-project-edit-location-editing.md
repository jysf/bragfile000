---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-033
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
  created_at: 2026-06-11

references:
  decisions: [DEC-020, DEC-017, DEC-018, DEC-006, DEC-007]
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
  related_specs: [SPEC-027, SPEC-029, SPEC-031]
---

# SPEC-033: brag project edit — location editing

## Context

This is the **last spec of STAGE-007** (Projects core). It was **peeled
from SPEC-029** at that spec's design when the L-watch fired: `edit`
(scalar fields *and* location editing) + `archive` + `delete` read **L**
together, so SPEC-029 shipped scalar-only `edit` and location editing was
split here (STAGE-007 Spec Backlog; SPEC-029 `## Out of scope`).

It adds `--add-path` and `--remove-path` to the existing
`brag project edit <name|id>` command, backed by a new Store method
**`RemoveLocation`** — the single-row counterpart to SPEC-027's shipped
`AddLocation` — plus a transactional batch method `EditLocations` that
makes a multi-path edit atomic. It emits **DEC-020** fixing the
location-editing semantics (not-attached → error, other-project → error,
verbatim matching, atomic removes-before-adds, no `updated_at` bump on
location edits).

It does **not** touch `AddLocation`'s contract, add a migration, or close
the stage. With it shipped, STAGE-007's "full CRUD" success criterion is
complete (`new`/`list`/`show`/`edit`/`archive`/`delete` all ship, and
`edit` now covers the full `project_locations` surface).

Parent stage: `STAGE-007-projects-core.md` (PROJ-002, `projects-and-tags`).
Prior shipped work this extends: **SPEC-029** (PR with `edit`/`archive`/
`delete` + the scalar `runProjectEdit` this widens), **SPEC-027** (the
`project_locations` schema + `UNIQUE(path)` + `AddLocation` + the
`locationsForProject` hydrator), **SPEC-031** (`ProjectForPath`/DEC-019 —
the read-time normalization this spec's verbatim storage-side matching
complements).

## Goal

Add repeatable `--add-path` / `--remove-path` flags to `brag project edit`
and a Store `RemoveLocation(projectID, path)` method (counterpart to
`AddLocation`), applying a single invocation's location changes atomically
(removes before adds) and refusing to remove a path that is not attached
to this project. Emit **DEC-020** for the semantics. **No migration, no
schema change** — `project_locations` already exists (SPEC-027).

## Inputs

- **Files to read:**
  - `internal/storage/project.go` — the `Project` struct; `AddLocation`
    (the method `RemoveLocation` mirrors/inverts); `locationsForProject`
    (the hydrator the round-trip tests assert through); `UpdateProject` /
    `DeleteProject` (the `BeginTx` + `defer tx.Rollback()` + `Commit` tx
    shape `EditLocations` copies).
  - `internal/storage/errors.go` — `ErrLocationExists` (reused by the add
    path); the two new sentinels go beside it.
  - `internal/cli/project.go` — `runProjectEdit` / `newProjectEditCmd`
    (lines ~364-455 — what this spec widens), `resolveProjectByNameOrID`,
    the open/resolve and confirmation-to-stderr style.
  - `internal/cli/project_test.go` — `newProjectTestRoot` / `runProjectCmd`
    helpers; the existing `TestProjectEdit_*` family the new tests sit
    beside; `TestProjectEdit_NoFlagsErrUser` (`:547`) — the guard test
    that asserts `ErrUser` only (see premise audit).
  - `internal/storage/project_test.go` — `newTestStore(t)`,
    `backdateProject` (the §9 no-sleep helper, reused by the no-bump
    test).
  - `decisions/DEC-020` (emitted here), `DEC-017` (soft match — entries
    independent of locations), `DEC-018` (delete cascade — the in-tx
    counterpart `RemoveLocation` mirrors), `DEC-019` (cwd normalization —
    why storage matching stays verbatim).
- **External APIs:** none. Plain Store calls; no new dependency.
- **Related code paths:** `internal/storage/project.go` (+`errors.go`),
  `internal/cli/project.go` (`runProjectEdit` + `newProjectEditCmd`),
  `docs/api-contract.md` (the `brag project edit` section + a DEC-020
  References row), `decisions/DEC-020-*.md` (new).

## Outputs

- **Files created:**
  - `decisions/DEC-020-project-location-editing-semantics.md` — the
    location-editing semantics decision.
- **Files modified:**
  - `internal/storage/errors.go` — add `ErrLocationNotFound` and
    `ErrLocationOtherProject`.
  - `internal/storage/project.go` — add `RemoveLocation(projectID int64,
    path string) error` and `EditLocations(projectID int64, remove, add
    []string) error`. `AddLocation` is **unchanged**.
  - `internal/storage/project_test.go` — add the storage tests below
    (**additive**; no SPEC-027/029/031 test rewritten).
  - `internal/cli/project.go` — widen `newProjectEditCmd` (two
    `StringArray` flags + a rewritten `Long`) and `runProjectEdit` (guard
    accepts the new flags; scalar update, then atomic location batch).
  - `internal/cli/project_test.go` — add the CLI tests below
    (**additive**; `TestProjectEdit_NoFlagsErrUser` is **not** rewritten).
  - `docs/api-contract.md` — **status-change UPDATE**: extend the
    `### brag project edit <name|id>` section with the two flags + the
    location semantics; add a `DEC-020` row to the References list.
  - `projects/PROJ-002-projects-and-tags/stages/STAGE-007-projects-core.md`
    — update the SPEC-033 backlog line (cycle) and the Count line at ship.
- **New exports:**
  - `func (s *Store) RemoveLocation(projectID int64, path string) error`
  - `func (s *Store) EditLocations(projectID int64, remove, add []string) error`
  - `storage.ErrLocationNotFound`, `storage.ErrLocationOtherProject`
- **Database changes:** **NONE.** No migration. `project_locations`
  (incl. `UNIQUE(path)`) was created by `0004_add_projects` (SPEC-027);
  this spec only `INSERT`/`DELETE`s rows. This is load-bearing — it is
  why the count-bump audit case is empty.

### Premise audit (`projects/_templates/premise-audit.md`), run at design

Greps were **run at design 2026-06-11** and reconciled against the lists
below.

```
- [x] Inversion/removal: greps run — the edit-guard logic broadens and a
      forward-reference in the edit Long becomes false; NO test asserts the
      old guard message (reconciled below), so ZERO test rewrites
- [x] Addition/count-bump: greps run — NONE (no migration; no count-coupled
      assertion; no test counts Store methods or sentinels)
- [x] Status-change: greps run, every doc hit listed as updates/stays
- [x] Cross-check: actual grep hits reconciled against the lists above
```

**1. Inversion / removal — guard broadens; the "exact message" premise is
FALSE (caught by running the grep).** Adding `--add-path`/`--remove-path`
to `edit` changes two pieces of existing behavior:

- **(a) The at-least-one-flag guard must also accept the new flags.**
  Today `runProjectEdit` (`internal/cli/project.go:400-402`) errors when
  none of `--name`/`--status`/`--state-note` changed. After this spec a
  location-only edit (`edit foo --add-path /x`) must be *accepted*, so the
  guard gains `addChanged`/`removeChanged` terms. This is a genuine
  behavior change to existing logic — the inversion the audit exists to
  catch.
- **(b) The guard message string changes** (it now lists five flags).
  **The design instruction assumed "a test asserts that exact error
  message"** and required identifying and updating it. **Running the grep
  reconciled this assumption against reality and it is FALSE:**
  - `grep -rn "edit requires at least one" internal/` → **one hit only**:
    `internal/cli/project.go:401` (the guard string itself). **No test**
    asserts the message.
  - The only test exercising the guard is
    `TestProjectEdit_NoFlagsErrUser` (`internal/cli/project_test.go:547`),
    which runs `edit bragfile` (no flags) and asserts **only**
    `errors.Is(err, ErrUser)`. A truly-empty edit is **still** a user
    error after this spec, so that assertion holds verbatim. **Zero test
    rewrites.** (This is the §9 audit-grep cross-check working: *enumerate
    AND run* the grep, then reconcile — the run corrected the design's
    starting assumption. A new test is *added*, not a rewrite — see
    `TestProjectEdit_NoFlagsMessageListsLocationFlags` below — to pair the
    new five-flag message with an assertion per the locked-decision↔test
    rule.)
- **(c) A forward-reference in the edit `Long` is now false.**
  `internal/cli/project.go:373-374` reads *"Editing locations is a
  separate command (a later STAGE-007 spec); this edits scalar fields
  only."* This spec **is** that command, so the `Long` is rewritten (the
  status-change Long-text fix). Grep: `grep -n "separate command"
  internal/cli/project.go` → one hit (`:373`); no **test** asserts that
  sentence (`TestProjectMutations_HelpShowsExamples` asserts only
  `"Examples:"` + `"brag project edit"`, both preserved by the rewrite —
  verified). Zero test rewrites from (c) either.

Other inversion check: `grep -rn "AddLocation\|locationsForProject"
internal/` → `AddLocation` is called by `runProjectNew`
(`internal/cli/project.go`) and tested in `project_test.go`; this spec
**does not change `AddLocation`** (its verbatim-store contract is reused
as-is and is explicitly out of scope), so those premises are preserved.
`EditLocations`' add path duplicates `AddLocation`'s existence-check logic
inside its transaction rather than calling it (so the batch shares one
tx); `AddLocation` itself is untouched.

**2. Addition / count-bump — NONE.** No migration is added;
`schema_migrations` stays at **4**. Grep run:
`grep -rln "want 4\|count != 4\|0004_add_projects" internal/storage` →
the only hits are the SPEC-027 migration sites
(`store_test.go`, `project_migration_test.go`, `fts_test.go`), all already
at 4 and **untouched** here. Method/sentinel counting:
`grep -rn "Commands()\|len(.*Commands\|NumMethod\|NumField" internal/cli
internal/storage cmd` → the only hit is
`internal/cli/list_test.go:629`, a **test-local** root iterated to read
the `list` subcommand's `Short`; it does not enumerate or count the
`project` parent's subcommands, the Store's methods, or the sentinel-error
set — so adding two Store methods and two sentinels couples to **no**
assertion. (Same registration-gap finding as SPEC-028/029; restated.)
No bumps.

**3. Status change — the `brag project edit` location flags.** Grep run:
`grep -rln -i "project" docs/ README.md` (17 files). Per-spec doc scope is
narrow (the comprehensive tutorial/architecture sweep is **STAGE-008**).
Disposition:
- **Updates (this spec):**
  - `docs/api-contract.md` — **UPDATE.** Extend the existing
    `### brag project edit <name|id>` section (lines 567-586) with the
    `--add-path`/`--remove-path` flags and their semantics; update the
    "Pass at least one of …" line to list all five flags. Add a `DEC-020`
    row to the References list (after the `DEC-018` row at line 676). (The
    DEC-017/018 rows already present **stay**.)
  - `internal/cli/project.go` edit `Long` — the forward-reference fix
    (1(c) above); not a doc file but the same status-change class.
- **Stays here (already correct / STAGE-008):**
  - `docs/data-model.md` — **STAYS.** The `project_locations.path`
    description already reads "stored verbatim (SPEC-031 owns
    normalization). Globally unique" — both still true after this spec
    (`RemoveLocation` matches verbatim; the `UNIQUE` invariant is
    preserved). No schema, no column change.
  - The SPEC-029/030/031/032 "Out of scope → SPEC-033" references and the
    STAGE-007 backlog `--add-path`/`--remove-path` lines — **STAY**: they
    correctly point forward at this spec; shipping it does not invalidate
    them (the stage close updates the backlog status line, not these).
  - `docs/tutorial.md`, `docs/architecture.md`, `README.md`,
    `docs/brag-entry.schema.json`, `docs/CONTEXTCORE_ALIGNMENT.md`,
    `docs/macos-notarization-checklist.md`, `docs/development.md`,
    `docs/blog/**`, `docs/framework-feedback/**`, `docs/reports/**` —
    **STAY** (STAGE-008 sweep / historical / `entries.project` prose; no
    shipped-status claim about `edit`'s location flags).

**4. Cross-check.** Actual grep hits reconciled against the lists above at
design; no un-enumerated hit remained. (Build-side: re-run
`grep -rn "edit requires at least one" internal/` and
`grep -rln -i "project" docs/ README.md` before editing and treat any
delta as a question, per the premise-audit cross-check rule.)

## Acceptance Criteria

Testable outcomes. Happy path, error cases, edge cases.

- [ ] **`edit --add-path` attaches a path.** `brag project edit <p>
  --add-path <dir>` exits 0, prints `Edited project "<p>".` to **stderr**
  (stdout empty); afterward `show <p>` lists `<dir>` among its locations.
- [ ] **`edit --remove-path` detaches a path.** `edit <p> --remove-path
  <dir>` (where `<dir>` is attached to `<p>`) exits 0; afterward `show
  <p>` no longer lists `<dir>`, and `<dir>` is free to register to a new
  project.
- [ ] **Both flags are repeatable.** `edit <p> --add-path /a --add-path
  /b` attaches both `/a` and `/b` in one invocation (DEC-020 / cobra
  `StringArray`).
- [ ] **Removing a not-attached path is a user error.** `edit <p>
  --remove-path <q>` where `<q>` is registered to **no** project exits 1
  (`ErrLocationNotFound` mapped); nothing changes.
- [ ] **Removing another project's path is a user error.** `edit <p>
  --remove-path <q>` where `<q>` is attached to a **different** project
  exits 1 (`ErrLocationOtherProject` mapped); `<q>` stays attached to that
  other project.
- [ ] **Adding an already-registered path is a user error.** `edit <p>
  --add-path <q>` where `<q>` is attached to another project exits 1
  (`ErrLocationExists` mapped); nothing changes.
- [ ] **A multi-path edit is atomic (rolls back).** An invocation whose
  later operation fails (e.g. `--add-path /free --add-path /occupied`,
  `/occupied` owned by another project) applies **none** of its changes:
  `/free` is **not** attached afterward.
- [ ] **Removes apply before adds.** `edit <p> --remove-path /a --add-path
  /a` (same `/a`, attached to `<p>`) succeeds and leaves `/a` attached
  once (no transient `UNIQUE(path)` collision).
- [ ] **Location edits compose with scalar edits.** `edit <p> --status
  paused --add-path /y` sets status `paused` **and** attaches `/y` in one
  call.
- [ ] **The guard accepts a path-only edit.** `edit <p> --add-path /y`
  (no `--name`/`--status`/`--state-note`) is **not** rejected by the
  at-least-one-flag guard.
- [ ] **A no-flag edit still errors, naming all five flags.** `edit <p>`
  with no flags exits 1 (`ErrUser`) and the message lists
  `--add-path`/`--remove-path` alongside the scalar flags.
- [ ] **Paths match verbatim.** A `--remove-path` value that differs from
  the stored string only by normalization (e.g. a trailing `/`) does
  **not** match (reported as not-attached); the exact stored string does.
- [ ] **Location edits do not bump `updated_at`.** `EditLocations`
  leaves `projects.updated_at` unchanged (scalar recency is unaffected by
  a location-only edit).
- [ ] **No SQL under `internal/cli/`**; no new migration; `go test
  ./...`, `gofmt -l .`, `go vet ./...`, `CGO_ENABLED=0 go build ./...`
  clean; `brag project edit --help` shows `--add-path`/`--remove-path` in
  the **real binary**.

## Failing Tests

Written at **design**; build makes them pass. All are **new/additive** —
the premise audit found zero test rewrites (the guard test asserts
`ErrUser`, which still holds). Storage tests use `t.TempDir()`
(`storage-tests-use-tempdir`) via `newTestStore(t)`; CLI tests use
`newProjectTestRoot` / `runProjectCmd` and out/err buffer separation.

### `internal/storage/project_test.go` (modify — additive)

- `"TestRemoveLocation_RemovesAttachedPath"` — `CreateProject{Name:"p"}`;
  `AddLocation(id,"/a")`; `RemoveLocation(id,"/a")` returns nil;
  `GetProject(id).Locations` is empty.
- `"TestRemoveLocation_NotAttachedErrLocationNotFound"` — create `p` (no
  locations); `RemoveLocation(id,"/nope")` returns
  `errors.Is(err, ErrLocationNotFound)`.
- `"TestRemoveLocation_OtherProjectErrLocationOtherProject"` — create
  `p1`, `AddLocation(p1,"/a")`; create `p2`; `RemoveLocation(p2.ID,"/a")`
  returns `errors.Is(err, ErrLocationOtherProject)`; `GetProject(p1.ID).
  Locations` still `["/a"]` (the other project's location untouched).
- `"TestRemoveLocation_VerbatimMatch"` — `AddLocation(id,"/a/b")`;
  `RemoveLocation(id,"/a/b/")` (trailing slash) returns
  `errors.Is(err, ErrLocationNotFound)` (verbatim, **not** `filepath.Clean`d);
  then `RemoveLocation(id,"/a/b")` succeeds. Locks the verbatim-matching
  sub-decision (DEC-020).
- `"TestEditLocations_RepeatableAdds"` — create `p`; `EditLocations(id,
  nil, []string{"/a","/b"})` returns nil; `GetProject(id).Locations`
  deep-equals `["/a","/b"]` (insertion order).
- `"TestEditLocations_RemovesBeforeAdds"` — create `p`,
  `AddLocation(id,"/a")`; `EditLocations(id, []string{"/a"},
  []string{"/a"})` returns nil; `GetProject(id).Locations` is exactly
  `["/a"]` (removed then re-added — no transient `UNIQUE` collision).
- `"TestEditLocations_AddDuplicateErrLocationExists"` — create `p1`,
  `AddLocation(p1,"/a")`; create `p2`; `EditLocations(p2.ID, nil,
  []string{"/a"})` returns `errors.Is(err, ErrLocationExists)`.
- `"TestEditLocations_RollsBackOnOccupiedAdd"` — create `owner`,
  `AddLocation(owner,"/occupied")`; create `p`; `EditLocations(p.ID, nil,
  []string{"/free","/occupied"})` returns `errors.Is(err,
  ErrLocationExists)`; assert `GetProject(p.ID).Locations` is **empty**
  (`/free` was rolled back — the atomicity contract). A second
  `sql.Open(path)` handle confirms `SELECT COUNT(*) FROM project_locations
  WHERE path='/free'` is `0`.
- `"TestEditLocations_RollsBackOnNotAttachedRemove"` — create `p`,
  `AddLocation(p,"/a")`; `EditLocations(p.ID, []string{"/a","/typo"},
  nil)` returns `errors.Is(err, ErrLocationNotFound)`; assert
  `GetProject(p.ID).Locations` is still `["/a"]` (the first remove rolled
  back).
- `"TestEditLocations_DoesNotBumpUpdatedAt"` — create `p`; `backdateProject(
  t, path, p.ID, time.Date(2020,1,1,0,0,0,0,time.UTC))`; `EditLocations(
  p.ID, nil, []string{"/a"})`; assert `GetProject(p.ID).UpdatedAt` equals
  the backdated 2020 instant (a location edit leaves `updated_at`
  unmoved — DEC-020; no `time.Sleep`, §9).

### `internal/cli/project_test.go` (modify — additive)

- `"TestProjectEdit_AddPath"` — `new bragfile --path /tmp/x`; `edit
  bragfile --add-path /tmp/y` exits 0, stderr contains `Edited project
  "bragfile".`, **stdout empty**; then `show bragfile` contains `/tmp/y`
  (and `/tmp/x`).
- `"TestProjectEdit_RemovePath"` — `new bragfile --path /tmp/x`; `edit
  bragfile --add-path /tmp/y`; `edit bragfile --remove-path /tmp/x` exits
  0; `show bragfile` no longer contains `/tmp/x` but still contains
  `/tmp/y`.
- `"TestProjectEdit_RepeatableAddPath"` — `new bragfile --path /tmp/x`;
  `edit bragfile --add-path /a --add-path /b`; `show bragfile` contains
  both `/a` and `/b` (repeatable `StringArray`).
- `"TestProjectEdit_RemoveNotAttachedErrUser"` — `new bragfile --path
  /tmp/x`; `edit bragfile --remove-path /nope` → `errors.Is(err, ErrUser)`;
  message mentions the path is not registered; `show bragfile` still lists
  `/tmp/x`.
- `"TestProjectEdit_RemoveOtherProjectErrUser"` — `new a --path /pa`; `new
  b --path /pb`; `edit b --remove-path /pa` → `errors.Is(err, ErrUser)`;
  `show a` still lists `/pa` (the other project's location intact).
- `"TestProjectEdit_AddDuplicateErrUser"` — `new a --path /pa`; `new b
  --path /pb`; `edit b --add-path /pa` → `errors.Is(err, ErrUser)`
  (already registered); `show a` still lists `/pa`.
- `"TestProjectEdit_AtomicBatchRollsBack"` — `new bragfile --path /x`;
  `new other --path /occupied`; `edit bragfile --add-path /free --add-path
  /occupied` → `errors.Is(err, ErrUser)`; `show bragfile` does **not**
  list `/free` (CLI-observable rollback — DEC-020 atomicity).
- `"TestProjectEdit_ComposesScalarAndLocation"` — `new bragfile --path /x`;
  `edit bragfile --status paused --add-path /y` exits 0; `show bragfile`
  contains `Status: paused` **and** `/y` (scalar + location in one call).
- `"TestProjectEdit_PathOnlyEditAccepted"` — `new bragfile --path /x`;
  `edit bragfile --add-path /y` exits 0 (the guard no longer rejects a
  location-only edit — the broadened-guard pairing).
- `"TestProjectEdit_NoFlagsMessageListsLocationFlags"` — `new bragfile
  --path /x`; `edit bragfile` (no flags) → `errors.Is(err, ErrUser)` AND
  `err.Error()` contains `--add-path` and `--remove-path` (the updated
  five-flag guard message). *(New test; the existing
  `TestProjectEdit_NoFlagsErrUser` — which asserts `ErrUser` only — is
  left as-is and still passes.)*
- `"TestProjectEdit_HelpShowsLocationFlags"` — `edit --help` contains
  `Examples:`, `brag project edit`, AND `--add-path` (the widened `Long`).
  Positive substring asserts; the §12 NOT-contains self-audit is N/A (no
  Failing Test asserts the ABSENCE of any token).

> **Locked-decision ↔ test traceability (§9).** **DEC-020 / verbatim
> match** → `TestRemoveLocation_VerbatimMatch`. **DEC-020 / not-attached
> → error** → `TestRemoveLocation_NotAttachedErrLocationNotFound` +
> `TestProjectEdit_RemoveNotAttachedErrUser`. **DEC-020 / other-project →
> error** → `TestRemoveLocation_OtherProjectErrLocationOtherProject` +
> `TestProjectEdit_RemoveOtherProjectErrUser`. **DEC-020 / atomic
> removes-before-adds + rollback** →
> `TestEditLocations_RemovesBeforeAdds` +
> `TestEditLocations_RollsBackOnOccupiedAdd` +
> `TestEditLocations_RollsBackOnNotAttachedRemove` +
> `TestProjectEdit_AtomicBatchRollsBack`. **DEC-020 / add-duplicate** →
> `TestEditLocations_AddDuplicateErrLocationExists` +
> `TestProjectEdit_AddDuplicateErrUser`. **DEC-020 / no `updated_at`
> bump** → `TestEditLocations_DoesNotBumpUpdatedAt`. **LD-REPEATABLE
> (`StringArray`)** → `TestEditLocations_RepeatableAdds` +
> `TestProjectEdit_RepeatableAddPath`. **LD-COMPOSE (scalar+location
> sequential)** → `TestProjectEdit_ComposesScalarAndLocation`.
> **Guard broadened** → `TestProjectEdit_PathOnlyEditAccepted` +
> `TestProjectEdit_NoFlagsMessageListsLocationFlags`. **Happy paths** →
> `TestRemoveLocation_RemovesAttachedPath` + `TestProjectEdit_AddPath` +
> `TestProjectEdit_RemovePath`.

> **§9 no-sleep note.** The only timestamp assertion is
> `TestEditLocations_DoesNotBumpUpdatedAt`, which uses `backdateProject`
> (the existing second-`sql.Open` helper) and asserts the post-edit
> `updated_at` **equals** the backdated 2020 instant. No `time.Sleep`.

## Implementation Context

*Read this section and the files it points to before build. This is the
whole handoff — the build session has only this spec.*

### Decisions that apply

- **`DEC-020` (emitted by this spec)** — the location-editing semantics:
  not-attached remove → `ErrLocationNotFound`; other-project remove →
  `ErrLocationOtherProject`; verbatim path matching; a single invocation's
  location changes apply in **one transaction, removes before adds**,
  all-or-nothing; scalar and location edits are **sequential** (scalar
  first), each atomic but not jointly atomic; **location edits do not bump
  `updated_at`**. Read it before build — it is the why behind every choice
  here.
- **`DEC-017`** (shipped, SPEC-027) — soft string match: `entries.project`
  is free text, independent of `project_locations`. Location edits touch
  **no** `entries` row. (Relevant as the boundary: this spec never reads
  or writes `entries`.)
- **`DEC-018`** (shipped, SPEC-029) — delete blast radius: `DeleteProject`
  removes `project_locations` rows **manually in-tx** because FK
  enforcement is OFF. `RemoveLocation`/`EditLocations` are the **manual
  single-row / single-batch counterpart** to that cascade, and freeing a
  `UNIQUE(path)` for reuse works the same way (the row is gone, the path
  is free).
- **`DEC-006` / `DEC-007`** — cobra: the flags + the at-least-one-flag
  guard + the error→`UserErrorf` mapping live inline in `RunE` (cobra's
  built-in validators return unwrappable plain errors that exit 2). The
  two new flags are added on the existing `newProjectEditCmd`; **no
  `cmd/brag/main.go` change** (the `project` parent is registered).

### Locked design decisions

Confidence on each is ≥ 0.8, so no `/guidance/questions.yaml` entry is
filed (§14). Only the cross-cutting location semantics are durable enough
to warrant a DEC (**DEC-020**); the rest are localized LDs.

- **LD-REMOVE-METHOD — `RemoveLocation` is a thin public counterpart over
  the transactional `EditLocations` engine.** Confidence 0.88.
  `RemoveLocation(projectID, path)` is the named inverse of `AddLocation`
  the stage backlog asked for; it is implemented as
  `s.EditLocations(projectID, []string{path}, nil)` so a single remove and
  a batch share one validated, transactional code path (no duplicated
  not-attached / other-project logic). `AddLocation` stays a separate
  shipped method (its contract is out of scope); `EditLocations`'
  **add** path duplicates `AddLocation`'s existence-check + insert inside
  its own transaction (so the whole batch is one tx) rather than calling
  the non-transactional `AddLocation`. *Rejected alternative:* a
  self-contained single-row `RemoveLocation` plus sequential
  `AddLocation`/`RemoveLocation` calls from the CLI — rejected because
  sequential calls each commit independently (no rollback across the
  batch), and a pure validation pass in the CLI would need raw `SELECT`s,
  violating `no-sql-in-cli-layer`. Keeping the transaction in **one Store
  method** keeps all SQL in `internal/storage` and delivers real
  rollback.

- **LD-REPEATABLE — `--add-path`/`--remove-path` are repeatable
  `StringArray`.** Confidence 0.88. Each flag is
  `cmd.Flags().StringArray(...)`, so `--add-path /a --add-path /b`
  collects `["/a","/b"]`. **`StringArray`, not `StringSlice`:**
  `StringSlice` comma-splits each value, which would corrupt a path
  containing a comma and break the verbatim-storage contract;
  `StringArray` treats each occurrence as one literal value. Default
  `nil`; presence detected via `cmd.Flags().Changed(...)`. *Rejected
  alternative:* single `String` flags — rejected for asymmetry (one add
  per call) and because batch edits are the natural multi-directory use
  case the brief calls out.

- **LD-COMPOSE — scalar edit first, then the atomic location batch.**
  Confidence 0.82 (the DEC-020 companion contract). `runProjectEdit`
  calls `UpdateProject` (only if a scalar flag changed) then
  `EditLocations` (only if a location flag changed). Scalar-first makes an
  invalid `--status` / duplicate `--name` abort inside `UpdateProject`
  *before* any location write. The two are not joined in one transaction
  (that would unify `UpdateProject` with the location methods); the only
  partial window is valid-scalar + failing-location, which leaves the
  scalar change applied and is re-runnable. *Rejected alternatives:*
  one combined transaction (disproportionate method merge); location-first
  (a valid location batch could apply before a scalar validation failure —
  worse fail-fast). See DEC-020.

- **LD-NO-BUMP — location edits do not move `updated_at`.** Confidence
  0.80. `EditLocations` writes only `project_locations`; it never updates
  the `projects` row, so `updated_at` (the scalar-field recency signal,
  SPEC-029) is unaffected by a location-only edit. *Rejected alternative:*
  bump `updated_at` on location edits — rejected as conflating structural
  location changes with scalar recency and complicating the location-only
  path. (Folded into DEC-020.)

### Constraints that apply

(see `/guidance/constraints.yaml` for full text)

- `no-sql-in-cli-layer` (**blocking**) — `runProjectEdit` calls Store
  methods only (`UpdateProject`, `EditLocations`, `resolveProjectByNameOrID`);
  the atomic location validation+mutation lives entirely in
  `EditLocations` in `internal/storage`. No `database/sql` import in
  `internal/cli/project.go`.
- `stdout-is-for-data-stderr-is-for-humans` (**blocking**) — the
  `Edited project "<p>".` confirmation and every error → **stderr**; these
  commands write **nothing to stdout** (success tests assert `outBuf`
  empty).
- `storage-tests-use-tempdir` (**blocking**) — every new storage test
  uses `newTestStore(t)` / `t.TempDir()`; the rollback test's second
  `sql.Open` handle points at the same temp path; never `~/.bragfile`.
- `errors-wrap-with-context` (warning) — `fmt.Errorf("edit locations:
  remove %q: %w", path, err)` etc., matching the existing method style.
- `test-before-implementation` (blocking) — the Failing Tests above are
  the design deliverable.
- `one-spec-per-pr` (blocking) — the PR references SPEC-033 only.
- `timestamps-in-utc-rfc3339` — no new timestamp writes (location edits
  don't touch `updated_at`); the constraint is satisfied vacuously.
- `no-cgo` / `no-new-top-level-deps-without-decision` — pure-Go path; no
  new dependency. `internal/storage/project.go` already imports
  everything `EditLocations` needs (`context`, `database/sql`, `errors`,
  `fmt`); `internal/cli/project.go` needs no new import (`StringArray` is
  on the already-used `cmd.Flags()`).

### Dev/prod DB isolation (PROJ-002 brief) — still mandatory this stage

The schema is at v0.2.x (post-`0004`) from SPEC-027. Run the dev binary
against a **dev DB** (`BRAGFILE_DB=~/.bragfile-dev/db.sqlite` or `--db`).
**Never open the production `~/.bragfile/db.sqlite`** with a v0.2.x
binary. SPEC-033 adds no migration, but the rule stands. All tests use
`t.TempDir()` regardless.

### Prior related work

- `SPEC-029` (shipped) — the scalar `runProjectEdit` + `newProjectEditCmd`
  this widens; the `resolveProjectByNameOrID` resolver and the
  confirmation-to-stderr discipline reused verbatim. Its `## Out of scope`
  explicitly peeled location editing + `RemoveLocation` here.
- `SPEC-027` (shipped) — `project_locations` (`UNIQUE(path)`),
  `AddLocation` (the verbatim-store contract reused), `locationsForProject`
  (the hydrator the round-trip tests assert through). Its FK-off forward
  note (realized in DEC-018) is why a removed row is gone for good and the
  path frees immediately.
- `SPEC-031` (shipped) — `ProjectForPath`/DEC-019: read-time
  `filepath.Clean` normalization. This spec's verbatim storage-side
  matching is the deliberate counterpart (normalize at resolve, store
  verbatim).
- `internal/storage/store.go` / `project.go` `UpdateProject` /
  `DeleteProject` — the `BeginTx` + `defer tx.Rollback()` + `Commit` shape
  `EditLocations` copies.

### Out of scope (for this spec specifically)

If any of these feels necessary during build, **stop and flag** — do not
expand this spec.

- **Any change to `AddLocation`'s contract** (signature, verbatim-store
  behavior, error). It is reused as-is; `EditLocations` re-implements the
  add check inside its tx rather than altering `AddLocation`.
- **cwd auto-detection for `--add-path`** (e.g. `--add-path` defaulting to
  `os.Getwd()`). That is SPEC-031/DEC-019 territory; `--add-path` takes an
  explicit path here.
- **Path normalization at storage time.** Verbatim (DEC-020); normalization
  stays at resolve time (DEC-019).
- **Bumping `updated_at` on location edits.** Out (LD-NO-BUMP / DEC-020).
- **Joining scalar + location edits in one transaction.** Out (LD-COMPOSE
  / DEC-020); they are sequential.
- **A `--clear-paths` / replace-all flag, or location reordering.** Not in
  the brief; the add/remove primitives suffice.
- **The STAGE-007 close itself** (stage-ship reflection, backlog Count
  finalization beyond the SPEC-033 status line). That is the Stage Ship
  prompt, run separately after this spec ships.

## Notes for the Implementer

### `storage` — the two sentinels (`internal/storage/errors.go`)

Add beside the existing `ErrLocationExists`:

```go
// ErrLocationNotFound is returned (wrapped) by RemoveLocation/EditLocations
// when the path to remove is attached to no project (a typo guard — removing
// a path that was never registered is a user error, not a silent no-op).
var ErrLocationNotFound = errors.New("location not found")

// ErrLocationOtherProject is returned (wrapped) by RemoveLocation/EditLocations
// when the path to remove is attached to a DIFFERENT project. Paths are
// globally unique (UNIQUE(path)); removing another project's location through
// this project is refused rather than silently deleting it.
var ErrLocationOtherProject = errors.New("location attached to a different project")
```

### `storage` — `RemoveLocation` + `EditLocations` (`internal/storage/project.go`)

Add beside `AddLocation`. **No new imports** — `project.go` already imports
`context`, `database/sql`, `errors`, `fmt` (and `path/filepath`, `strings`,
`time`). Transcribe verbatim:

```go
// RemoveLocation detaches path from the project identified by projectID. It is
// the single-path counterpart to AddLocation. Paths match VERBATIM against the
// stored value (storage is verbatim end to end; SPEC-031/DEC-019 own
// normalization at cwd-resolve time only). Errors:
//   - ErrLocationNotFound      if path is attached to no project
//   - ErrLocationOtherProject  if path is attached to a different project
// (DEC-020). Implemented over the same transactional engine as EditLocations
// so a single remove and a batch share one validated code path.
func (s *Store) RemoveLocation(projectID int64, path string) error {
	return s.EditLocations(projectID, []string{path}, nil)
}

// EditLocations applies a set of location removals and additions to the project
// identified by projectID in ONE transaction, all-or-nothing (DEC-020). Removes
// are applied before adds, so a path may be removed and re-added in the same
// call without a transient UNIQUE(path) collision. Any failure rolls the whole
// set back, leaving project_locations unchanged. Per-path rules:
//   - remove: path must be attached to projectID
//             (else ErrLocationNotFound / ErrLocationOtherProject)
//   - add:    path must be free after the removes
//             (else ErrLocationExists — the same global-uniqueness guard
//             AddLocation enforces; the in-tx COUNT also catches in-batch dups)
// Paths match verbatim. updated_at is NOT bumped: location editing is a
// structural change to project_locations, distinct from the scalar-field
// recency UpdateProject tracks (DEC-020).
func (s *Store) EditLocations(projectID int64, remove, add []string) error {
	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("edit locations: %w", err)
	}
	defer tx.Rollback()

	// Removes first — frees any UNIQUE path that a later add re-registers.
	for _, path := range remove {
		var ownerID int64
		err := tx.QueryRowContext(ctx,
			`SELECT project_id FROM project_locations WHERE path = ?`, path,
		).Scan(&ownerID)
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("edit locations: remove %q: %w", path, ErrLocationNotFound)
		}
		if err != nil {
			return fmt.Errorf("edit locations: remove %q: %w", path, err)
		}
		if ownerID != projectID {
			return fmt.Errorf("edit locations: remove %q: %w", path, ErrLocationOtherProject)
		}
		if _, err := tx.ExecContext(ctx,
			`DELETE FROM project_locations WHERE path = ?`, path,
		); err != nil {
			return fmt.Errorf("edit locations: remove %q: %w", path, err)
		}
	}

	// Adds — path must be free; the in-tx COUNT also backstops in-batch dups.
	for _, path := range add {
		var exists int
		if err := tx.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM project_locations WHERE path = ?`, path,
		).Scan(&exists); err != nil {
			return fmt.Errorf("edit locations: add %q: %w", path, err)
		}
		if exists > 0 {
			return fmt.Errorf("edit locations: add %q: %w", path, ErrLocationExists)
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO project_locations (project_id, path) VALUES (?, ?)`,
			projectID, path,
		); err != nil {
			return fmt.Errorf("edit locations: add %q: %w", path, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("edit locations: %w", err)
	}
	return nil
}
```

Notes on the engine:
- **Removes-before-adds** is what makes `--remove-path /a --add-path /a`
  legal (the `DELETE` frees `/a` before the `INSERT` re-checks `COUNT`).
- **In-batch duplicate adds** (`--add-path /a --add-path /a`) fail on the
  second iteration's `COUNT` (the first add is visible in the same tx) and
  roll back — correct.
- `projectID` is assumed valid (the CLI resolves the project first, same
  as `AddLocation`); the method does not re-verify project existence.

### `cli` — widen `newProjectEditCmd` (`internal/cli/project.go`)

Add the two `StringArray` flags and rewrite the `Long` (drop the now-false
"separate command" forward-reference; document the location flags). Keep
`"Examples:"` and `"brag project edit"` in the body (the help test asserts
them).

```go
func newProjectEditCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit <name|id>",
		Short: "Edit a project's name, status, state note, or locations",
		Long: `Edit a project's fields. The project is resolved by name first, then by id.
Pass at least one of --name, --status, --state-note, --add-path, or --remove-path;
unspecified fields are left unchanged.

Locations are edited with --add-path and --remove-path (both repeatable). Paths
are matched verbatim against what was registered. --remove-path errors if the
path is not registered to this project; --add-path errors if the path is already
registered to another project. All location changes in one invocation apply
atomically (removes before adds), so a path can be removed and re-added at once.

Renaming a project does NOT rewrite the project string on existing brag entries
— they keep what they were captured with (DEC-017).

Examples:
  brag project edit bragfile --status paused
  brag project edit bragfile --state-note "shipped tags; next: cut v0.2.0"
  brag project edit bragfile --name brag-cli
  brag project edit bragfile --add-path ~/code/bragfile --add-path /srv/bragfile
  brag project edit bragfile --remove-path /srv/old-location`,
		RunE: runProjectEdit,
	}
	cmd.Flags().String("name", "", "new project name (rename; must be unique)")
	cmd.Flags().String("status", "", "new status (one of: active, paused, done, archived)")
	cmd.Flags().String("state-note", "", "new state/next-action note")
	cmd.Flags().StringArray("add-path", nil, "attach a filesystem path to the project (repeatable)")
	cmd.Flags().StringArray("remove-path", nil, "detach a filesystem path from the project (repeatable)")
	return cmd
}
```

### `cli` — widen `runProjectEdit` (`internal/cli/project.go`)

Broaden the guard to accept the location flags, do the scalar update (only
when a scalar flag changed) first, then the atomic location batch (only
when a location flag changed). Transcribe verbatim:

```go
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
	addChanged := cmd.Flags().Changed("add-path")
	removeChanged := cmd.Flags().Changed("remove-path")
	if !nameChanged && !statusChanged && !noteChanged && !addChanged && !removeChanged {
		return UserErrorf("edit requires at least one of --name, --status, --state-note, --add-path, --remove-path")
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

	finalName := current.Name

	// Scalar fields first: UpdateProject validates --status/--name and returns
	// before any write on a bad value, so a typo'd status aborts the whole edit
	// before any location is touched (DEC-020 scalar-then-locations order).
	if nameChanged || statusChanged || noteChanged {
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
		finalName = updated.Name
	}

	// Location edits next, applied atomically (removes before adds,
	// all-or-nothing — DEC-020).
	if addChanged || removeChanged {
		adds, _ := cmd.Flags().GetStringArray("add-path")
		removes, _ := cmd.Flags().GetStringArray("remove-path")
		if err := s.EditLocations(current.ID, removes, adds); err != nil {
			switch {
			case errors.Is(err, storage.ErrLocationNotFound):
				return UserErrorf("cannot remove a path that is not registered to project %q", finalName)
			case errors.Is(err, storage.ErrLocationOtherProject):
				return UserErrorf("cannot remove a path registered to a different project")
			case errors.Is(err, storage.ErrLocationExists):
				return UserErrorf("cannot add a path that is already registered")
			}
			return fmt.Errorf("edit locations: %w", err)
		}
	}

	fmt.Fprintf(cmd.ErrOrStderr(), "Edited project %q.\n", finalName)
	return nil
}
```

Notes:
- `finalName` preserves the SPEC-029 behavior of confirming with the
  *new* name after a rename, while a location-only edit confirms with the
  current name.
- The `ErrLocationExists` message says "already registered" (not "to
  another project") so it reads correctly whether the collision is with
  another project or a re-add of a path already on this one.

### `docs/api-contract.md` — status-change UPDATE (literal)

Replace the body of the existing `### brag project edit <name|id>` section
(lines 575-586, from "Edits a project's scalar fields." through the last
bullet) with this (keep the heading + the fenced example block at the top;
add the two location examples to that block):

```
brag project edit bragfile --status paused
brag project edit bragfile --state-note "shipped tags; next: cut v0.2.0"
brag project edit bragfile --name brag-cli
brag project edit bragfile --add-path ~/code/bragfile
brag project edit bragfile --remove-path /srv/old-location
```

Edits a project's fields. The argument resolves as a **name first**, then as
a positive-integer **id**. Pass at least one of `--name`, `--status`,
`--state-note`, `--add-path`, or `--remove-path`; unspecified fields are
unchanged.

- `--name` — rename (must be unique). Renaming does **not** rewrite the project
  string on existing brag entries (DEC-017); they keep their captured string.
- `--status` — one of `active`, `paused`, `done`, `archived` (validated).
- `--state-note` — the free-text state/next-action note.
- `--add-path` / `--remove-path` (both repeatable) — attach/detach filesystem
  locations. Paths match **verbatim** against what was registered.
  `--remove-path` exits 1 if the path is not registered to this project or is
  registered to a **different** project; `--add-path` exits 1 if the path is
  already registered. All location changes in one invocation apply
  **atomically** (removes before adds); a failure leaves locations unchanged.
  Location edits do **not** change `updated_at` (DEC-020).
- A scalar edit bumps `updated_at` (so the project rises in `brag project list`
  recency order); a location-only edit does not.
- Exits 0 on success; stderr: `Edited project "<name>".` (stdout empty).
- Exit 1 (user error) if no flag is given, the project is not found, the new
  name is already taken, `--status` is outside the enum, or a location operation
  is rejected.

(Then add to the References list at the bottom, after the `DEC-018` row:)

- `DEC-020` — `brag project edit` location editing: `RemoveLocation`/`EditLocations`; remove-not-attached and remove-other-project are user errors; verbatim path matching; one invocation's location changes are atomic (removes before adds); location edits don't bump `updated_at`.

### Gotchas

- **Two files named `project.go`.** `RemoveLocation`/`EditLocations` go in
  `internal/storage/project.go` (no new imports). The flags + guard +
  batch call go in `internal/cli/project.go` (no new imports — `StringArray`
  is on the already-used `cmd.Flags()`, and `errors`/`fmt`/`strings`/
  `storage` are present).
- **Do not touch `AddLocation`.** `EditLocations` re-implements the add
  check inside its tx on purpose; changing `AddLocation` is out of scope
  and would risk SPEC-027/028 test premises.
- **`StringArray`, not `StringSlice`** (no comma-splitting — verbatim
  paths). See LD-REPEATABLE.
- **The guard test is not rewritten.** `TestProjectEdit_NoFlagsErrUser`
  asserts `ErrUser` only and still passes; add the *new*
  `TestProjectEdit_NoFlagsMessageListsLocationFlags` for the message.
- **`gofmt -w .` + `go vet ./...`** before the PR; confirm
  `./brag project edit --help` shows `--add-path`/`--remove-path` in the
  real binary, and that a manual `edit … --remove-path <not-attached>`
  exits 1.

### §12(b) design-time verification (run at design 2026-06-11)

No external tool (no migration / DDL). The location semantics were traced
against the live `project_locations` schema (`UNIQUE(path)`, FK-off) and
the `modernc.org/sqlite` driver behavior already exercised by SPEC-027/029:

- `SELECT project_id FROM project_locations WHERE path = ?` on a missing
  path → `sql.ErrNoRows` → `ErrLocationNotFound`. ✅ (mirrors `GetProject`'s
  `sql.ErrNoRows` mapping.)
- Removes-before-adds: `DELETE … WHERE path='/a'` then
  `SELECT COUNT(*) … WHERE path='/a'` returns 0 within the same tx →
  the re-add `INSERT` succeeds. ✅
- In-batch duplicate add: first `INSERT` is visible to the second
  iteration's `COUNT` within the same tx → second add → `ErrLocationExists`
  → `tx.Rollback()` undoes the first. ✅ (the atomicity contract.)
- Verbatim: `WHERE path = '/a/b/'` does not equal stored `'/a/b'` →
  no row → `ErrLocationNotFound` (no `filepath.Clean` in the storage
  layer). ✅
- The guard-message grep (`grep -rn "edit requires at least one"
  internal/`) returns the source line only — **no test** asserts the
  message; reconciled in the premise audit. ✅

---

## Build Completion

*Filled in at the end of the **build** cycle, before advancing to verify.*

- **Branch:** `feat/spec-033-project-edit-location-editing`
- **PR (if applicable):** opened after this file is updated
- **All acceptance criteria met?** yes
- **New decisions emitted:**
  - `DEC-020` — `brag project edit` location-editing semantics *(emitted
    at design; confirmed no build deviation — verbatim transcription of all
    implementation literals)*
- **Deviations from spec:**
  - None. All 20 failing tests written first; all implementation literals
    transcribed verbatim from `## Notes for the Implementer`. The 11th CLI
    test (`TestProjectEdit_HelpShowsLocationFlags`) matches the spec's list;
    the spec heading says "20" but lists 21 — implementation matches the
    enumerated list, not the summary count.
- **Follow-up work identified:**
  - None beyond the STAGE-007 close (which is the Stage Ship prompt, run
    separately).

### Build-phase reflection (3 questions, short answers)

1. **What was unclear in the spec that slowed you down?**
   — Nothing slowed the build. The spec's `## Notes for the Implementer`
   carried verbatim code for every piece; reading the existing `DeleteProject`
   and `AddLocation` patterns first made the `EditLocations` shape obvious
   before the transcription confirmed it.

2. **Was there a constraint or decision that should have been listed but wasn't?**
   — No missing constraints. The spec explicitly noted "no new imports" and
   confirmed the `no-sql-in-cli-layer` + `StringArray`-not-`StringSlice`
   choices; both held cleanly at build.

3. **If you did this task again, what would you do differently?**
   — Nothing substantial. The literal-artifact-as-spec pattern made this
   the cleanest build in STAGE-007: read spec → write tests → transcribe
   code → run suite. The premise audit's "grep and reconcile" already caught
   the guard-message-test false assumption at design, so build had zero
   surprises.

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
