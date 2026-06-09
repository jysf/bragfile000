---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-028
  type: story                      # epic | story | task | bug | chore
  cycle: verify
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
  created_at: 2026-06-08

references:
  decisions: [DEC-017, DEC-011, DEC-013, DEC-014, DEC-006, DEC-007, DEC-005]
  constraints:
    - no-sql-in-cli-layer
    - stdout-is-for-data-stderr-is-for-humans
    - storage-tests-use-tempdir
    - errors-wrap-with-context
    - test-before-implementation
    - one-spec-per-pr
    - no-cgo
    - no-new-top-level-deps-without-decision
  related_specs: [SPEC-027, SPEC-029, SPEC-030, SPEC-031, SPEC-032, SPEC-026]
---

# SPEC-028: brag project new / list / show

## Context

This is the **first CLI surface** of STAGE-007 (Projects core), built
directly on the Store read/create primitives SPEC-027 shipped
(`CreateProject` / `GetProject` / `ListProjects` / `AddLocation`; the
`Project` struct; `ErrProjectExists` / `ErrLocationExists`; reuse of
`ErrNotFound`). SPEC-027 laid down the `projects` + `project_locations`
schema (`0004_add_projects`), emitted **DEC-017** (soft string match;
the `active|paused|done|archived` status enum; the single free-text
`state_note`), and stopped deliberately short of any `brag project`
command. SPEC-028 adds the **read+create** half of that command surface:

- `brag project` — a **parent** command (prints help; no `RunE`),
  mirroring SPEC-026's `brag tag` parent. SPEC-029 slots `edit` /
  `archive` / `delete` into this same parent.
- `brag project new <name> --path <dir>` — `CreateProject` + `AddLocation`.
- `brag project list` — `ListProjects`; plain + `--format json` (DEC-011).
- `brag project show <name|id>` — `GetProject` (+ name lookup); plain +
  `--format json`. Renders name, status, state_note, locations — **not**
  a recent-brag count (that is the SPEC-030 dashboard).

It is a **read+create** spec: no mutation of existing projects
(`edit`/`archive`/`delete` are SPEC-029), no `here` cwd resolver
(SPEC-031), no `brag add` auto-fill (SPEC-032), **no migration, no schema
change**. The complexity is **M**: one parent + three thin cobra
subcommands, one tiny Store read helper (`GetProjectByName`), two
`internal/export` JSON helpers, and a status-change doc update to
`docs/api-contract.md` — purely additive, no inversion, no count-bump.

- Parent stage: `STAGE-007-projects-core.md` (SPEC-028 is the second of
  six specs; SPEC-027 shipped first, PR #40).
- Project: `PROJ-002` brief (dev/prod DB isolation governs this stage —
  see Implementation Context).
- Builds on `DEC-017` (the relationship model + the data the renderers
  display) and the DEC-011/013/014 output-shape family.

## Goal

Add the `brag project` parent command with `new`, `list`, and `show`
subcommands on top of SPEC-027's Store primitives — `new <name> --path
<dir>` registers a project with one initial location; `list` and `show`
read projects in plain and `--format json` form per the DEC-011 output
family — implemented as thin cobra commands that call **Store methods
only** (no SQL in the CLI layer), with **no schema change and no
migration**.

## Inputs

- **Files to read:**
  - `internal/storage/project.go` — the SPEC-027 primitives
    (`CreateProject` / `GetProject` / `ListProjects` / `AddLocation`),
    the `Project` struct, `locationsForProject`; the new
    `GetProjectByName` sits beside them, cloning `GetProject`.
  - `internal/storage/errors.go` — `ErrProjectExists`,
    `ErrLocationExists`, `ErrNotFound` (reused, not re-defined).
  - `internal/cli/tag.go` + `internal/cli/tags.go` — the SPEC-026 parent
    command (`NewTagCmd`: `Use`, no `RunE`, `AddCommand`) and the
    `--format` handling / unknown-format `UserErrorf` you mirror.
  - `internal/cli/list.go` + `internal/cli/show.go` — the open/resolve
    pattern (`getFlagString("db")` → `config.ResolveDBPath` →
    `storage.Open` → `defer s.Close()`), inline positional-arg
    validation (DEC-007), and the `--format` switch shape.
  - `internal/cli/delete.go` — the **confirmation-to-stderr** discipline
    (`fmt.Fprintln(cmd.ErrOrStderr(), "Deleted.")`) that `new` mirrors.
  - `internal/cli/errors.go` — `UserErrorf` / `ErrUser`.
  - `internal/export/json.go` — `ToJSON` / `ToTagsJSON` (the DEC-011
    naked-array, 2-space-indent, `[]`-not-`null` marshal); the new
    project helpers sit in a sibling `internal/export/project.go`.
  - `cmd/brag/main.go` — command registration
    (`root.AddCommand(cli.NewProjectCmd())`).
  - `internal/cli/tags_test.go` + `internal/cli/list_test.go` — the CLI
    test patterns (`newXxxTestRoot`, `seedListEntry`, `runXxxCmd`,
    out/err buffer separation).
  - `decisions/DEC-017` (relationship + status enum + state_note),
    `DEC-011` / `DEC-013` / `DEC-014` (output shapes), `DEC-006` /
    `DEC-007` (cobra + RunE validation), `DEC-005` (autoincrement PKs).
- **External APIs:** none. Plain Store calls; no new dependency (would
  need a DEC under `no-new-top-level-deps-without-decision`).
- **Related code paths:** `internal/cli/` (the new parent + three
  subcommands), `internal/storage/project.go` (one read helper),
  `internal/export/` (two JSON helpers), `cmd/brag/main.go`
  (registration), `docs/api-contract.md` (three command sections).

## Outputs

- **Files created:**
  - `internal/cli/project.go` — `NewProjectCmd()` (the `brag project`
    parent, no `RunE`) + `newProjectNewCmd()` / `newProjectListCmd()` /
    `newProjectShowCmd()` and their `RunE`s. (All three subcommands live
    under the one parent, so they share one file — mirroring `tag.go`,
    which holds the `tag` parent + `rename` + `merge`.)
  - `internal/cli/project_test.go` — the CLI failing tests below.
  - `internal/export/project.go` — `projectRecord` + `ToProjectsJSON`
    (naked array, for `list`) + `ToProjectJSON` (single object, for
    `show`).
- **Files modified:**
  - `internal/storage/project.go` — add `GetProjectByName(name string)
    (Project, error)` (a **small SPEC-028 Store addition**, not a
    SPEC-027 change; see Notes). Reuses `ErrNotFound`.
  - `internal/storage/project_test.go` — add the two `GetProjectByName`
    tests below (additive; no rewrite of SPEC-027 tests).
  - `cmd/brag/main.go` — register `cli.NewProjectCmd()`.
  - `docs/api-contract.md` — **status-change UPDATE**: add three command
    sections + a DEC-017 References row (see audit + Notes).
- **New exports:**
  - `func (s *Store) GetProjectByName(name string) (Project, error)`
  - `func cli.NewProjectCmd() *cobra.Command`
  - `func export.ToProjectsJSON(projects []storage.Project) ([]byte, error)`
  - `func export.ToProjectJSON(p storage.Project) ([]byte, error)`
- **Database changes:** **NONE.** No migration. SPEC-028 is CLI + one
  Store read helper only. This is a load-bearing property of the spec
  (it is why the count-bump premise-audit case is empty), not an
  omission.

### Premise audit (`projects/_templates/premise-audit.md`), run at design

Greps were **run** at design **2026-06-08** and reconciled against the
lists below.

```
- [x] Inversion/removal: greps run — NONE (purely additive commands; no schema change)
- [x] Addition/count-bump: greps run — NONE (no migration; no count-coupled assertion)
- [x] Status-change: greps run, every doc hit listed as updates/stays
- [x] Cross-check: actual grep hits reconciled against the lists above
```

**1. Inversion / removal — NONE.** `brag project new`/`list`/`show` are
brand-new commands; no existing behavior is inverted, no flag or column
is removed, and **DEC-017's soft-string-match leaves `entries.project`
untouched** — so nothing about `list --project` / `ByProject` /
`GroupEntriesByProject` changes. Greps run:
`grep -rn 'func Test.*[Pp]roject' internal/` surfaces the SPEC-027
storage tests (`TestCreateProject_*`, `TestGetProject_NotFound`,
`TestAddLocation_*`, `TestListProjects_*`, `TestOpen_ProjectsTablesExist`,
…) and the SPEC-007/SPEC-019 `entries.project` filter/digest tests
(`TestList_FilterByProject`, `TestListCmd_ShowProject`, the `ByProject`
aggregate tests). **Every one is unchanged in premise** — the SPEC-027
storage tests keep passing (SPEC-028 only *adds* `GetProjectByName`
beside them), and the `entries.project` tests are untouched by DEC-017.
No planned rewrites or deletions.

**2. Addition / count-bump — NONE.** No migration is added, so
`schema_migrations` is untouched (SPEC-027's count of **4** stands). Grep
run: `grep -rn '0004_add_projects\|want 4\|count != 4' internal/` →
hits are the SPEC-027 sites (`store_test.go:172,206-208`,
`project_migration_test.go`, `fts_test.go`) — all already at 4, none
touched by SPEC-028. Grep run for command-set coupling:
`grep -rn 'Commands()\|len(.*Commands' internal/cli cmd` → the only
`root.Commands()` use is `list_test.go:629`, which iterates a
**test-local** root built with only `NewListCmd()` to read the `list`
subcommand's `Short` — **no test enumerates or counts the production
root subcommand set** (`cmd/brag/main.go`). Registering `project` in
`main.go` therefore couples to no assertion. No bumps. (This is the same
SPEC-026 registration-gap finding, restated; see "Registration gap"
under Notes.)

**3. Status change — the new commands.** Grep run:
`grep -rln -i 'project' docs/ README.md` (17 files). The per-spec doc
scope here is narrow — the comprehensive tutorial/architecture sweep is
**STAGE-008** (brief + stage Scope). Disposition of each:

- **Updates (this spec):**
  - `docs/api-contract.md` — **UPDATE.**
    - **Add** three command sections after `### brag tags`/`### brag tag
      *` and before `### brag completion`: `### brag project new <name>
      --path <dir>`, `### brag project list`, `### brag project show
      <name|id>` (see ## Notes for the literal).
    - References list (~line 504): **add** a `DEC-017` row. (The
      DEC-011/013/014 rows already exist and **stay**.)
- **Stays here (STAGE-008, or no status claim invalidated):**
  - `docs/data-model.md` — **STAYS.** SPEC-027 already added the
    `projects` + `project_locations` entity tables and struck the
    "Projects normalization … Deferred" bullet (`grep -n 'project'
    docs/data-model.md` confirms the tables + DEC-017 reference are
    present). SPEC-028 adds **no schema and no column**, so data-model
    carries no status claim this spec invalidates.
  - `docs/tutorial.md`, `docs/architecture.md`, `README.md` — **STAY.**
    These describe the user workflow / the `internal/` package diagram;
    the full projects+tags tutorial and the `architecture.md` diagram +
    `internal/projects`-style responsibilities refresh are the
    **STAGE-008** sweep (brief, stage Scope: "Comprehensive doc sweep …
    is STAGE-008; only per-spec, premise-audit-driven doc updates fold
    in here"). The remaining `project` hits there reference the
    *existing* free-text `entries.project` (untouched by DEC-017) or are
    historical/narrative.
  - `docs/brag-entry.schema.json`, `docs/CONTEXTCORE_ALIGNMENT.md`,
    `docs/macos-notarization-checklist.md`, `docs/development.md`,
    `docs/blog/**`, `docs/framework-feedback/**`, `docs/reports/**` —
    **STAY.** Historical / process / `entries.project` input-contract
    prose; no shipped-behavior status claim about the `brag project`
    command surface.

**4. Cross-check.** Actual grep hits reconciled against the lists above
at design; no un-enumerated hit remained. (Build-side: re-run
`grep -rln -i 'project' docs/ README.md` before the doc edit and treat
any delta as a question, per the premise-audit cross-check rule.)

## Acceptance Criteria

Testable outcomes. Happy path, error cases, edge cases.

- [ ] **`brag project` (bare) prints help.** `brag project` with no
  subcommand prints usage (containing `Usage:` and the `new`/`list`/`show`
  subcommand names) to **stdout**, exit 0, stderr empty — cobra's default
  for a parent with subcommands and no `RunE` (mirrors `brag tag`).
- [ ] **`new` creates and attaches.** `brag project new bragfile --path
  /tmp/x` exits 0, prints `Created project "bragfile".` to **stderr**,
  stdout empty; afterward the project exists with status `active`,
  empty `state_note`, and `/tmp/x` as its single location (verified via
  `brag project show bragfile`).
- [ ] **`new` requires `--path`.** `brag project new bragfile` (no
  `--path`, or `--path ""`) returns a **user error** (exit 1); no
  project is created.
- [ ] **`new` requires a non-empty name.** `brag project new "" --path
  /tmp/x` (or whitespace-only) returns a user error (exit 1); nothing
  created.
- [ ] **`new` duplicate name → clean user error.** A second `new` with an
  existing name maps `ErrProjectExists` to a user error (exit 1) naming
  the project; no second row.
- [ ] **`new` path already registered → clean user error, no orphan.**
  `new projA --path /p` then `new projB --path /p` returns a user error
  (exit 1) stating the path is already registered, **and `projB` is NOT
  created** (the path is pre-checked before `CreateProject`, so no
  location-less orphan project is left behind). `--path` is stored
  **verbatim** (SPEC-031 owns normalization), matching `AddLocation`.
- [ ] **`list` plain ordering + shape.** `brag project list` prints one
  tab-separated `<name>\t<status>\t<locations>` row per project to
  stdout, ordered `updated_at DESC, id DESC` (newest first, the
  `ListProjects` order), locations comma-joined (`-` when none); stderr
  empty; exit 0.
- [ ] **`list --format json`.** Emits a **naked JSON array** of project
  objects (2-space indent, DEC-011), keys `id, name, status, state_note,
  locations, created_at, updated_at` in that order, `locations` a JSON
  array of strings, timestamps RFC3339; same order as plain; an empty DB
  emits `[]` (never `null`).
- [ ] **`list` default `--format` is plain.** With no `--format` flag
  (default `""`), output is the plain rows; an unknown `--format` value
  (e.g. `xml`) exits 1 (user error).
- [ ] **`show <name|id>` resolves name-first, id-fallback.** `brag
  project show bragfile` finds it by name; `brag project show <id>` finds
  it by id when no project is *named* that integer; a project literally
  named `"1"` is returned by `show 1` (name takes precedence over id).
- [ ] **`show` plain renders name/status/state_note/locations.** Plain
  output contains `Name: <name>`, `Status: <status>`, `State note:
  <state_note or "-">`, and a `Locations:` block listing each path
  (or `Locations: (none)`). **No recent-brag count** (that is SPEC-030).
- [ ] **`show --format json`** emits a **single JSON object** (not an
  array) with the project element shape (`locations` a JSON array;
  `[]` not `null` when empty); 2-space indent.
- [ ] **`show` miss → clean user error.** `brag project show
  nonexistent` (no name and, if numeric, no such id) maps `ErrNotFound`
  to a user error (exit 1). Unknown `--format` exits 1.
- [ ] **stdout/stderr discipline.** `list`/`show` data goes to stdout;
  `new`'s `Created project …` confirmation and all errors go to stderr
  (a success `new`/`list`/`show` test asserts no cross-leakage, per §9).
- [ ] **No SQL under `internal/cli/`** (commands call Store methods only);
  no new migration file; `go test ./...`, `gofmt -l .`, `go vet ./...`,
  and `CGO_ENABLED=0 go build ./...` are clean; `brag project --help`
  works in the **real binary** (registration confirmed — see Notes).

## Failing Tests

Written at **design**; build makes them pass. All are **new/additive**
(no rewrites — the premise audit found zero inversions). Storage tests
use `t.TempDir()` (`storage-tests-use-tempdir`). CLI tests use the
`newProjectTestRoot` + buffer-separation patterns mirrored from
`tags_test.go`; each builds its own root with only `NewProjectCmd()`.

### `internal/storage/project_test.go` (modify — additive)

- `"TestGetProjectByName_RoundTrip"` — `CreateProject({Name:"bragfile",
  StateNote:"n"})`, `AddLocation(id,"/a")`; `GetProjectByName("bragfile")`
  returns the same `ID`/`Name`/`Status`/`StateNote` and
  `Locations == []string{"/a"}` (hydrated like `GetProject`).
- `"TestGetProjectByName_NotFound"` — `GetProjectByName("nope")` returns
  `errors.Is(err, ErrNotFound)` (reuses the existing sentinel; **no**
  new `ErrProjectNotFound`).

### `internal/export/project_test.go` (new)

- `"TestToProjectsJSON_NakedArrayShape"` — one project with one location
  marshals to a naked array whose single element has the seven keys in
  order (`id, name, status, state_note, locations, created_at,
  updated_at`), `locations` a JSON array; 2-space indented
  (`strings.Contains(out, "\n  {")` and `"    \"id\""`).
- `"TestToProjectsJSON_EmptyIsBracketsNotNull"` — `ToProjectsJSON(nil)`
  and `ToProjectsJSON([]storage.Project{})` both yield exactly `[]`
  (DEC-011 empty discipline; pre-flighted §12(b)).
- `"TestToProjectJSON_SingleObjectAndEmptyLocations"` — `ToProjectJSON`
  of a project yields a single JSON **object** (`HasPrefix(trim, "{")`);
  a project with no locations yields `"locations": []` (not `null`).

### `internal/cli/project_test.go` (new)

- `"TestProjectCmd_BarePrintsHelp"` — `brag project` (no subcommand) →
  stdout contains `Usage:` and the subcommand names `new`, `list`,
  `show`; exit 0; stderr empty. (Parent-with-no-RunE behavior, mirrors
  `tag`.)
- `"TestProjectNew_CreatesAndAttaches"` — `project new bragfile --path
  /tmp/x` exits 0, stderr contains `Created project "bragfile".`, stdout
  empty; then `project show bragfile` plain shows `Status: active` and a
  `/tmp/x` location line. (LD3/LD5)
- `"TestProjectNew_RequiresPath"` — `project new bragfile` (no `--path`)
  → `errors.Is(err, ErrUser)`; a follow-up `project list` is empty
  (nothing created). (LD1)
- `"TestProjectNew_EmptyNameErrUser"` — `project new "" --path /tmp/x` →
  `errors.Is(err, ErrUser)`; nothing created.
- `"TestProjectNew_DuplicateNameErrUser"` — create `bragfile` twice
  (distinct paths); the second → `errors.Is(err, ErrUser)`, message
  names `bragfile`; `project list` shows exactly one row.
- `"TestProjectNew_PathAlreadyRegisteredErrUser_NoOrphan"` — `new projA
  --path /p`, then `new projB --path /p` → `errors.Is(err, ErrUser)`,
  message mentions the path; **`project show projB` → `ErrUser`
  (not created)** and `project list` shows exactly one row (`projA`).
  Locks the orphan-prevention pre-check. (LD3)
- `"TestProjectNew_StdoutStderrSeparation"` — on success the `Created
  project …` line is on **stderr only**; stdout is empty (§9). (LD6)
- `"TestProjectList_PlainOrderingAndShape"` — create `p1`, `p2`, `p3`
  (each via `new` with distinct paths); `project list` plain rows are
  `<name>\t<status>\t<path>` in `updated_at DESC, id DESC` order —
  i.e. **newest-created first** (`p3, p2, p1`); stderr empty.
- `"TestProjectList_JSON"` — `project list --format json` → naked array
  (`HasPrefix(trim,"[")`), unmarshals to `[]{id,name,status,state_note,
  locations[],created_at,updated_at}` in the same order; 2-space indent
  (`strings.Contains(out, "    \"id\"")`).
- `"TestProjectList_EmptyJSONIsBrackets"` — no projects → `--format json`
  prints exactly `[]`; plain prints empty stdout. (LD4)
- `"TestProjectList_UnknownFormatErrUser"` — `project list --format xml`
  → `errors.Is(err, ErrUser)`.
- `"TestProjectShow_ByName"` — after `new bragfile --path /tmp/x` (and a
  project whose state_note is set via the Store seed helper), `project
  show bragfile` plain contains `Name: bragfile`, `Status: active`,
  `State note: <note or ->`, `Locations:` and `  /tmp/x`. (LD5)
- `"TestProjectShow_ById"` — `project show <id>` (the numeric id of an
  existing project, looked up via the Store) renders the same block.
  (LD2)
- `"TestProjectShow_NameFirstResolution"` — create a project named `"1"`
  (path `/one`) and another (id ≥ 2); `project show 1` returns the
  project **named** `"1"`, not project id 1 (name precedence). (LD2)
- `"TestProjectShow_JSONSingleObject"` — `project show bragfile --format
  json` → a single JSON **object** (`HasPrefix(trim,"{")`), `locations`
  a JSON array; 2-space indent.
- `"TestProjectShow_NotFoundErrUser"` — `project show nonexistent` →
  `errors.Is(err, ErrUser)`; `project show 99999` (numeric, no such id,
  no such name) → `errors.Is(err, ErrUser)`.
- `"TestProjectShow_UnknownFormatErrUser"` — `project show bragfile
  --format xml` → `errors.Is(err, ErrUser)`.
- `"TestProjectNew_HelpShowsExamples"` — `project new --help` (and
  `project list --help`) contain `Examples:` and a distinctive token
  (`brag project new` / `brag project list --format json`) — a token
  unique to the locked `Long` (SPEC-005 lesson; the §12 NOT-contains
  caveat does not apply — these are positive substring asserts).

> **Locked-decision ↔ test traceability (§9).** Each locked design
> decision (## Implementation Context → Locked design decisions) has a
> paired failing test: **LD1** (`--path` required) →
> `TestProjectNew_RequiresPath`; **LD2** (show name-first/id-fallback) →
> `TestProjectShow_NameFirstResolution` + `TestProjectShow_ById`; **LD3**
> (orphan-prevention path pre-check) →
> `TestProjectNew_PathAlreadyRegisteredErrUser_NoOrphan`; **LD4**
> (list=array / show=object / locations=array / `--format` default `""`)
> → `TestProjectList_JSON` + `TestProjectList_EmptyJSONIsBrackets` +
> `TestProjectShow_JSONSingleObject` + the two `*_UnknownFormatErrUser`
> tests + the export-helper tests; **LD5** (plain render literals) →
> `TestProjectList_PlainOrderingAndShape` + `TestProjectShow_ByName`;
> **LD6** (confirmation→stderr) → `TestProjectNew_StdoutStderrSeparation`.

> **§12(a) note for build:** the JSON element key set/order, the empty
> `[]` / single-`{}` shapes, and the plain-render line literals were
> validated at design (§12(b) below). Transcribe them; do not re-derive.

## Implementation Context

*Read this section and the files it points to before build. This is the
whole handoff — the build session has only this spec.*

### Decisions that apply

- **`DEC-017`** (shipped, SPEC-027) — the data these commands render and
  the relationship they honor. **Soft string match:** `entries.project`
  stays free text, so SPEC-028 writes/reads **nothing** about
  `entries.project` — `project show`/`list` render only the `projects`
  row + its locations, with **no recent-brag count** (that join is
  SPEC-030). The `status` enum is `active|paused|done|archived` (default
  `active`, Store-validated, no DB CHECK); `state_note` is a single
  free-text column. `new` does **not** let the user set status/state_note
  (status defaults `active`, state_note defaults `""`); editing them is
  SPEC-029.
- **`DEC-011`** — naked JSON array, field-names-match-columns, 2-space
  indent, `[]`-not-`null` on empty. `brag project list --format json`
  follows it verbatim. The **element shape** adds `locations` as a JSON
  **array of strings** (projects' locations are a genuine list, not a
  DEC-004-style comma-joined legacy string) — a within-family extension,
  not a divergence (see LD4).
- **`DEC-013` / `DEC-014`** — the output-shape family `show` borrows
  from: `show` of a **single** object emits a single JSON **object**
  (like DEC-014's single-document envelope), distinct from `list`'s
  DEC-011 array (like `brag list` vs `brag show`). The plain `show` block
  is a lightweight labeled render (not the DEC-014 markdown envelope —
  that envelope is for the aggregating digests, and `brag show` of one
  entry is the closer precedent).
- **`DEC-006`** — cobra: the parent + each subcommand is a
  `*cobra.Command` built by a `New…`/`new…Cmd()` constructor; the parent
  is registered in `cmd/brag/main.go`.
- **`DEC-007`** — required/positional validation lives **inline in
  `RunE`** via `UserErrorf` (cobra's built-in arg/flag validators return
  unwrappable plain errors that would exit 2). `--path` required, name
  non-empty, arg count, and `--format` value are all checked in `RunE`.
  Mirror `delete.go`/`show.go`/`tags.go`.
- **`DEC-005`** — INTEGER autoincrement PKs give the `id DESC` tie-break
  `ListProjects` (hence `project list`) orders on.

### Locked design decisions

The four design questions STAGE-007 surfaced for this spec, decided here
with reasoning. **No new DEC** — each is an application of the brief +
existing DECs, localized to SPEC-028's CLI surface; honest confidence on
each is ≥ 0.8, so no `/guidance/questions.yaml` entry is filed (§14).

- **LD1 — `new` REQUIRES `--path` (single).** Confidence 0.85. The
  brief's canonical flow is `new <name> --path <dir>`, and a path-less
  project defeats the headline "where does it live on this machine"
  feature. SPEC-028 ships **no other location-adding surface** (the
  `here` resolver is SPEC-031; `edit` is SPEC-029), so a path-less
  project created here would be permanently location-less — bad. Enforced
  in `RunE` via `UserErrorf` (DEC-007). *Rejected alternative
  (build-time):* optional/location-less project — rejected because it
  strands the project with no way to attach a location within this
  spec's surface. `--path` is **single** (not repeatable): multi-location
  is a one-project-many-directories capability, but the canonical `new`
  attaches one initial location; adding more is an edit-time concern
  (SPEC-029) — a repeatable `--path` is YAGNI here.

- **LD2 — `show <name|id>` resolves NAME-first, then integer-id
  fallback.** Confidence 0.82. Names are the user-facing, UNIQUE handle
  (the brief writes `<name|id>` name-first; users type names);
  id is the stable escape hatch. Resolution: look up by name via the new
  `GetProjectByName`; **if that misses AND the arg parses as a positive
  int64**, fall back to `GetProject(id)`; otherwise the name miss is the
  error. This makes the common case (typing a name) never ambiguous, and
  the only unreachable edge is a project *literally named* an integer
  shadowing that id-lookup — acceptable and documented, and exercised by
  `TestProjectShow_NameFirstResolution`. *Rejected alternative:* id-first
  (parse-int-then-name) — rejected because a project named `"42"` would
  be unreachable by `show`, the worse edge. The numeric-name shadow under
  name-first is the rarer, less surprising case.

- **LD3 — `new` is CreateProject + AddLocation, with a path pre-check to
  prevent an orphan.** Confidence 0.83. `CreateProject` commits before
  `AddLocation` runs, so a path conflict *after* a successful create
  would leave a location-less orphan project (and the name is now taken).
  To guarantee **nothing is created on a path conflict**, `new`
  pre-checks path availability by scanning `ListProjects()` Locations for
  an **exact** match (the same verbatim-string, global-uniqueness basis
  `AddLocation` uses) **before** `CreateProject`; on a hit it returns a
  clean `UserErrorf` and creates nothing. `AddLocation`'s
  `ErrLocationExists` is still mapped as a defensive backstop (single-user
  CLI has no real TOCTOU race). The pre-check uses only the SPEC-027
  primitive `ListProjects` — **no new Store method, no SQL in the CLI**
  (iterating returned structs is plain Go). *Rejected alternatives:* (a)
  accept the orphan + document — rejected, surprising ("error, but it
  created a half-project"); (b) a combined transactional
  `CreateProjectWithLocation` Store method — rejected as duplicating
  `CreateProject`/`AddLocation` SQL for a rare conflict; revisit only if
  multi-location `new` or a real race emerges. Path is stored
  **verbatim** (SPEC-031 owns normalization), so the pre-check, the
  insert, and `AddLocation`'s uniqueness are all consistently
  exact-string.

- **LD4 — output shapes.** Confidence 0.85.
  - `project list` **plain**: one `<name>\t<status>\t<locations>` row per
    project (locations comma-joined; `-` when none), tab-separated like
    `brag list`/`brag tags`, in `ListProjects` order
    (`updated_at DESC, id DESC`). State_note is **not** in `list` — the
    state-note + recent-brag dashboard is `brag project status`
    (SPEC-030); `list` stays a lean index.
  - `project list` **`--format json`**: naked array of project objects
    (DEC-011); `[]` on empty.
  - `project show` **plain**: a labeled block (Name / Status / State note
    / Locations) — see the literal in Notes.
  - `project show` **`--format json`**: a **single JSON object** (not an
    array) — `show` is one thing, mirroring `brag show` (one entry) vs
    `brag list` (many).
  - **Element shape** (shared by list elements and the show object):
    keys `id, name, status, state_note, locations, created_at,
    updated_at` in that order; `locations` a JSON **array of strings**
    (`[]` not `null`); timestamps RFC3339. Field names match the
    `projects` columns (plus `locations`).
  - **`--format` default is `""`** (empty → plain). The only accepted
    non-empty value is `json`; anything else → `UserErrorf` (mirrors
    `brag tags`'s `format != "" && format != "json"` check). *Stated
    explicitly per the STAGE-006 flag-default WATCH item (SPEC-026's
    `--format ""`-default lesson): each flag's DEFAULT is locked, not
    just its accepted set — see "Flag defaults" in Notes.*

### Constraints that apply

(see `/guidance/constraints.yaml` for full text)

- `no-sql-in-cli-layer` (**blocking**) — `project.go` imports only
  `storage` / `config` / `export` / `cobra` (+ stdlib `strconv` /
  `strings`); **never** `database/sql`. All persistence is Store calls,
  including the LD3 path pre-check (which iterates `ListProjects()`
  results — plain Go, no SQL).
- `stdout-is-for-data-stderr-is-for-humans` (**blocking**) — `list`/`show`
  rows + JSON go to **stdout**; `new`'s `Created project …` confirmation
  and all errors go to **stderr** (mirror `delete.go`'s `Deleted.`).
- `storage-tests-use-tempdir` (**blocking**) — the two `GetProjectByName`
  tests use `t.TempDir()` (via the package's existing helper); never
  touch `~/.bragfile`.
- `errors-wrap-with-context` (warning) — `GetProjectByName` wraps like
  `GetProject` (`fmt.Errorf("get project by name %q: %w", name, err)`);
  CLI wraps non-user errors (`fmt.Errorf("open store: %w", err)`).
- `test-before-implementation` (blocking) — the Failing Tests above are
  the design deliverable.
- `one-spec-per-pr` (blocking) — the PR references SPEC-028 only.
- `no-cgo` / `no-new-top-level-deps-without-decision` — pure-Go path; no
  new dependency.

### Design-time pre-flight (§12(b)) — run at design 2026-06-08, results below

- **JSON shapes** — the `projectRecord` marshal was run through a scratch
  `encoding/json` program (the helper is a near-clone of the shipped
  `ToTagsJSON`): `ToProjectsJSON([1 project])` → naked array with the
  seven keys in column order, `locations` a JSON array, 2-space indent;
  `ToProjectsJSON(empty)` → exactly `"[]"` (not `null`);
  `ToProjectJSON(p)` → a single `{…}` object; a project with
  `Locations: []string{}` → `"locations": []` (not `null`). ✅ (Output
  captured at design; transcribe the helper verbatim from Notes.)
- **Plain render literals** — the `list` row (`%s\t%s\t%s\n`) and the
  `show` block are this spec's own `fmt.Fprintf` output (deterministic,
  no external tool); the failing-test line assertions match them exactly
  (§12(a) self-check: `Name: `, `Status: `, `State note: `, `Locations:`
  and the `<name>\t<status>\t<loc>` row were reconciled against the
  locked `fmt` literals).
- **cobra help rendering** — `Examples:` + the distinctive example tokens
  follow the proven SPEC-026 `brag tags`/`brag tag` help pattern (cobra
  v1.10.2 renders the `Long` verbatim into `--help`); the help tests
  assert positive substrings only, so the §12 NOT-contains self-audit is
  N/A (no Failing Test asserts a token *absent* from a `Long`).

### Dev/prod DB isolation (PROJ-002 brief) — still mandatory this stage

The schema is at v0.2.x (post-`0004`) from SPEC-027. While v0.2.x is in
flight:

- Build/run the dev binary against a **dev DB**:
  `BRAGFILE_DB=~/.bragfile-dev/db.sqlite` (or `--db`), via `just install`
  → `~/go/bin/brag`. **Never open the production `~/.bragfile/db.sqlite`**
  with a v0.2.x binary (it carries the `0004` migration). SPEC-028 adds
  no migration, but the rule stands.
- Production stays brew-installed at v0.1.0; the documented upgrade is
  STAGE-008.
- All tests use `t.TempDir()` regardless.

### Prior related work

- `SPEC-027` (shipped, PR #40, `7a67834`) — the schema, the four Store
  primitives, `DEC-017`. Its forward design note flagged FK enforcement
  is OFF (decorative `REFERENCES`) — irrelevant to SPEC-028 (no delete
  here) but worth knowing. `GetProjectByName` is a near-clone of its
  `GetProject` (same `Scan` + `time.Parse` + `locationsForProject`
  hydrate).
- `SPEC-026` (shipped, PR #37) — the `brag tag` parent-command pattern
  (`NewTagCmd`: `Use`, no `RunE`, `AddCommand`), the `--format ""`-default
  handling, the `newXxxTestRoot`/buffer-separation CLI test patterns, and
  the **registration-gap** finding (no test enumerates the root set)
  carried forward here verbatim.

### Out of scope (for this spec specifically)

If any of these feels necessary during build, **stop and flag** — do not
expand this spec.

- **`edit` / `archive` / `delete`** → **SPEC-029** (Store
  `UpdateProject` / `ArchiveProject` / `DeleteProject` + the mutation
  CLI). SPEC-028 ships **read+create** only. Setting `status` /
  `state_note` to non-default values, adding/removing locations after
  `new`, and any project mutation are SPEC-029.
- **`brag project status` dashboard + recent-brag count** → **SPEC-030**.
  The recent-brag count is the DEC-017 `entries.project = projects.name`
  join; **not computed here**. `show`/`list` render no count.
- **`brag project here` cwd resolver + path normalization** →
  **SPEC-031**. `new --path` stores the path **verbatim**; the
  exact/nearest-ancestor/longest-prefix resolution and any
  absolute-path normalization are SPEC-031's.
- **`brag add` `--project` auto-fill** → **SPEC-032**.
- **Writing `'project'` taggings / `brag project tag`** → schema-ready
  only (DEC-015); the project-tagging command is a STAGE-008/PROJ-003
  candidate (STAGE-007 design question #4 default). No `'project'`
  taggings are written here.
- **Repeatable `--path` / multi-location `new` / `--status` / `--note`
  flags on `new`** — YAGNI for SPEC-028 (LD1/LD4). A project gets one
  initial location and the default status/note; the rest is SPEC-029.
- **Any migration or schema change** — there is none.

## Notes for the Implementer

### `storage` — the one read helper (`internal/storage/project.go`)

Add beside `GetProject`, cloning its shape exactly (Scan + `time.Parse`
UTC + `locationsForProject` hydrate); map `sql.ErrNoRows` → `ErrNotFound`
(reuse the sentinel; **do not** add `ErrProjectNotFound`):

```go
// GetProjectByName returns the project with the given name (names are
// globally UNIQUE), with its Locations hydrated in insertion order.
// Returns an error wrapping ErrNotFound if no row matches. Mirrors
// GetProject; used by `brag project show <name|id>` (SPEC-028).
func (s *Store) GetProjectByName(name string) (Project, error) {
	ctx := context.Background()

	var (
		p                          Project
		createdAtRaw, updatedAtRaw string
	)
	err := s.db.QueryRowContext(ctx,
		`SELECT id, name, status, state_note, created_at, updated_at
		 FROM projects WHERE name = ?`, name,
	).Scan(&p.ID, &p.Name, &p.Status, &p.StateNote, &createdAtRaw, &updatedAtRaw)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Project{}, fmt.Errorf("get project %q: %w", name, ErrNotFound)
		}
		return Project{}, fmt.Errorf("get project %q: %w", name, err)
	}

	created, err := time.Parse(time.RFC3339, createdAtRaw)
	if err != nil {
		return Project{}, fmt.Errorf("get project %q: parse created_at %q: %w", name, createdAtRaw, err)
	}
	updated, err := time.Parse(time.RFC3339, updatedAtRaw)
	if err != nil {
		return Project{}, fmt.Errorf("get project %q: parse updated_at %q: %w", name, updatedAtRaw, err)
	}
	p.CreatedAt = created.UTC()
	p.UpdatedAt = updated.UTC()

	locs, err := s.locationsForProject(ctx, p.ID)
	if err != nil {
		return Project{}, fmt.Errorf("get project %q: %w", name, err)
	}
	p.Locations = locs
	return p, nil
}
```

(`project.go` already imports `context`, `database/sql`, `errors`,
`fmt`, `time` — no new imports.)

### `export` — the two JSON helpers (`internal/export/project.go`, new)

Mirror `json.go`'s `ToTagsJSON` discipline (DEC-011 naked array, 2-space
indent, `[]`-not-`null`). `locations` is a JSON array; normalize a `nil`
slice to `[]string{}` so it renders `[]`:

```go
package export

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/jysf/bragfile000/internal/storage"
)

// projectRecord is the DEC-011-family serialization shape for a project:
// keys in `projects`-column order, locations as a JSON array of strings,
// timestamps pre-formatted RFC3339. Shared by ToProjectsJSON (list) and
// ToProjectJSON (show) so the array elements and the single-show object
// are byte-identical in shape.
type projectRecord struct {
	ID        int64    `json:"id"`
	Name      string   `json:"name"`
	Status    string   `json:"status"`
	StateNote string   `json:"state_note"`
	Locations []string `json:"locations"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
}

func toProjectRecord(p storage.Project) projectRecord {
	locs := p.Locations
	if locs == nil {
		locs = []string{}
	}
	return projectRecord{
		ID:        p.ID,
		Name:      p.Name,
		Status:    p.Status,
		StateNote: p.StateNote,
		Locations: locs,
		CreatedAt: p.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: p.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

// ToProjectsJSON renders projects as a naked JSON array (DEC-011 shape;
// 2-space indent). Empty/nil input renders "[]", never "null".
func ToProjectsJSON(projects []storage.Project) ([]byte, error) {
	out := make([]projectRecord, 0, len(projects))
	for _, p := range projects {
		out = append(out, toProjectRecord(p))
	}
	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal projects json: %w", err)
	}
	return b, nil
}

// ToProjectJSON renders a single project as a JSON object (not an array)
// for `brag project show --format json`. 2-space indent; an empty
// Locations renders "[]", never "null".
func ToProjectJSON(p storage.Project) ([]byte, error) {
	b, err := json.MarshalIndent(toProjectRecord(p), "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal project json: %w", err)
	}
	return b, nil
}
```

### `cli` — `internal/cli/project.go` (new)

`NewProjectCmd()` returns the parent (`Use: "project"`, a `Short`, **no
`RunE`** — a bare `brag project` then prints help, like `brag tag`); it
attaches the three subcommands. Each subcommand does the standard
open/resolve (`getFlagString(cmd, "db")` → `config.ResolveDBPath` →
`storage.Open` → `defer s.Close()`), inline `RunE` validation (DEC-007),
and maps sentinels to `UserErrorf`.

**Parent + Long literals (transcribe; §12(a)-self-audited):**

```go
// NewProjectCmd returns the `brag project` parent command with new,
// list, and show subcommands. A bare `brag project` prints help (cobra
// default for a command with subcommands and no RunE). SPEC-029 adds
// edit/archive/delete to this same parent.
func NewProjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Manage projects (new, list, show)",
	}
	cmd.AddCommand(newProjectNewCmd())
	cmd.AddCommand(newProjectListCmd())
	cmd.AddCommand(newProjectShowCmd())
	return cmd
}
```

`new` subcommand — `--path` is a **required** string flag (default `""`,
validated in `RunE`; LD1). Confirmation to **stderr** (LD6). Orphan
pre-check via `ListProjects` (LD3):

```go
func newProjectNewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "new <name> --path <dir>",
		Short: "Register a new project with an initial location",
		Long: `Register a new project with a filesystem location. The project starts with
status "active" and an empty state note; use 'brag project edit' to change them.
The --path is required and is stored verbatim; a path already registered to
another project is rejected.

Examples:
  brag project new bragfile --path ~/code/bragfile
  brag project new platform --path /srv/platform`,
		RunE: runProjectNew,
	}
	cmd.Flags().String("path", "", "filesystem directory for the project (required)")
	return cmd
}

func runProjectNew(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return UserErrorf("new requires exactly one <name> argument")
	}
	name := strings.TrimSpace(args[0])
	if name == "" {
		return UserErrorf("project name must not be empty")
	}
	path, _ := cmd.Flags().GetString("path")
	if strings.TrimSpace(path) == "" {
		return UserErrorf("--path is required (the project's directory)")
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

	// LD3: pre-check the path is free so a conflict creates no orphan
	// project. ListProjects hydrates Locations; iterating its result is
	// plain Go (no SQL in the CLI layer). Exact-string match — the same
	// verbatim basis AddLocation enforces; SPEC-031 owns normalization.
	existing, err := s.ListProjects()
	if err != nil {
		return fmt.Errorf("list projects: %w", err)
	}
	for _, p := range existing {
		for _, loc := range p.Locations {
			if loc == path {
				return UserErrorf("path %q is already registered to project %q", path, p.Name)
			}
		}
	}

	created, err := s.CreateProject(storage.Project{Name: name})
	if err != nil {
		if errors.Is(err, storage.ErrProjectExists) {
			return UserErrorf("project %q already exists", name)
		}
		return fmt.Errorf("create project: %w", err)
	}
	if err := s.AddLocation(created.ID, path); err != nil {
		if errors.Is(err, storage.ErrLocationExists) {
			// Defensive backstop for the TOCTOU window (no real race in a
			// single-user CLI); the pre-check above is the primary guard.
			return UserErrorf("path %q is already registered to another project", path)
		}
		return fmt.Errorf("add location: %w", err)
	}
	fmt.Fprintf(cmd.ErrOrStderr(), "Created project %q.\n", name)
	return nil
}
```

`list` subcommand — plain rows or `--format json`; default `""` (LD4):

```go
func newProjectListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List projects (most-recently-updated first)",
		Long: `List every registered project, most-recently-updated first, one per line as
<name>` + "\t" + `<status>` + "\t" + `<locations> (comma-joined; "-" when none).

Output is plain tab-separated rows (default) or a naked JSON array of project
objects (--format json) per DEC-011.

Examples:
  brag project list                 # name<TAB>status<TAB>locations
  brag project list --format json   # naked JSON array of project objects`,
		RunE: runProjectList,
	}
	cmd.Flags().String("format", "", "output format (one of: json); default is plain tab-separated")
	return cmd
}

func runProjectList(cmd *cobra.Command, _ []string) error {
	format, _ := cmd.Flags().GetString("format")
	if format != "" && format != "json" {
		return UserErrorf("unknown --format value %q (accepted: json)", format)
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

	projects, err := s.ListProjects()
	if err != nil {
		return fmt.Errorf("list projects: %w", err)
	}

	out := cmd.OutOrStdout()
	if format == "json" {
		body, err := export.ToProjectsJSON(projects)
		if err != nil {
			return fmt.Errorf("render projects json: %w", err)
		}
		fmt.Fprintln(out, string(body))
		return nil
	}
	for _, p := range projects {
		loc := "-"
		if len(p.Locations) > 0 {
			loc = strings.Join(p.Locations, ",")
		}
		fmt.Fprintf(out, "%s\t%s\t%s\n", p.Name, p.Status, loc)
	}
	return nil
}
```

`show` subcommand — name-first/id-fallback resolution (LD2); plain block
or single JSON object (LD4):

```go
func newProjectShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <name|id>",
		Short: "Show a single project by name or id",
		Long: `Show one project's name, status, state note, and locations. The argument is
resolved as a name first; if no project has that name and the argument is a
positive integer, it is resolved as a project id.

Examples:
  brag project show bragfile         # by name
  brag project show 3                # by id (when no project is named "3")
  brag project show bragfile --format json`,
		RunE: runProjectShow,
	}
	cmd.Flags().String("format", "", "output format (one of: json); default is plain")
	return cmd
}

func runProjectShow(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return UserErrorf("show requires exactly one <name|id> argument")
	}
	key := strings.TrimSpace(args[0])
	if key == "" {
		return UserErrorf("project name or id must not be empty")
	}
	format, _ := cmd.Flags().GetString("format")
	if format != "" && format != "json" {
		return UserErrorf("unknown --format value %q (accepted: json)", format)
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

	// LD2: name first, then integer-id fallback.
	project, err := s.GetProjectByName(key)
	if err != nil {
		if !errors.Is(err, storage.ErrNotFound) {
			return fmt.Errorf("get project: %w", err)
		}
		// Name miss: try id iff the key is a positive integer.
		id, convErr := strconv.ParseInt(key, 10, 64)
		if convErr != nil || id <= 0 {
			return UserErrorf("no project named %q", key)
		}
		project, err = s.GetProject(id)
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				return UserErrorf("no project named or with id %q", key)
			}
			return fmt.Errorf("get project: %w", err)
		}
	}

	out := cmd.OutOrStdout()
	if format == "json" {
		body, err := export.ToProjectJSON(project)
		if err != nil {
			return fmt.Errorf("render project json: %w", err)
		}
		fmt.Fprintln(out, string(body))
		return nil
	}
	renderProjectPlain(out, project)
	return nil
}
```

**Plain `show` render literal** (transcribe; the failing tests assert
these exact line prefixes — §12(a)). `state_note` empty → `-`;
locations empty → `Locations: (none)`:

```go
func renderProjectPlain(out io.Writer, p storage.Project) {
	fmt.Fprintf(out, "Name: %s\n", p.Name)
	fmt.Fprintf(out, "Status: %s\n", p.Status)
	note := p.StateNote
	if note == "" {
		note = "-"
	}
	fmt.Fprintf(out, "State note: %s\n", note)
	if len(p.Locations) == 0 {
		fmt.Fprintln(out, "Locations: (none)")
		return
	}
	fmt.Fprintln(out, "Locations:")
	for _, l := range p.Locations {
		fmt.Fprintf(out, "  %s\n", l)
	}
}
```

Imports for `project.go`: `errors`, `fmt`, `io`, `strconv`, `strings`,
plus `config`, `export`, `storage`, `cobra`. **No `database/sql`**
(`no-sql-in-cli-layer`).

### Registration (and the registration gap)

In `cmd/brag/main.go`, after the existing `root.AddCommand(...)` calls
(e.g. beside `NewTagCmd()`):

```go
root.AddCommand(cli.NewProjectCmd())
```

**No test enumerates the production root command set** (premise audit
case 2; the only `root.Commands()` use is a test-local root in
`list_test.go`). So registration is **not** covered by any CLI unit test
assertion — build MUST confirm it in the **real binary**: `just build`
(or `CGO_ENABLED=0 go build ./...`) then `./brag project --help` and
`./brag project new --help` render without error and list the
subcommands. State this confirmation in the Build Completion notes.

### `docs/api-contract.md` — the status-change UPDATE (literal)

Add these three sections **after** `### brag tag merge <src> <dst>` and
**before** `### brag completion <shell>`. House style mirrors the
adjacent `brag tags` / `brag tag *` sections (fenced example block + flag
bullets + exit/IO notes):

```
### `brag project new <name> --path <dir>` — register a project (STAGE-007)

​```
brag project new bragfile --path ~/code/bragfile
​```

Registers a new project named `<name>` with one initial filesystem location
`<dir>`. The project starts with status `active` and an empty state note
(use `brag project edit` to change them — STAGE-007 later spec). `--path` is
required and stored verbatim (path normalization is `brag project here`'s
concern, STAGE-007).

- Exits 0 on success; stderr: `Created project "<name>".` (stdout empty).
- Exit 1 (user error) if `<name>` is empty, `--path` is missing/empty, the
  name already exists, or the path is already registered to another project
  (in which case nothing is created — the path is checked first).

### `brag project list` — list projects (STAGE-007)

​```
brag project list                 # name<TAB>status<TAB>locations
brag project list --format json   # naked JSON array of project objects
​```

Lists every registered project, most-recently-updated first
(`updated_at DESC, id DESC`), one per line as `<name>\t<status>\t<locations>`
(locations comma-joined; `-` when none).

- `--format json` — naked JSON array of project objects (DEC-011; 2-space
  indent; `[]` on empty, never `null`). Object keys: `id, name, status,
  state_note, locations, created_at, updated_at` (locations a JSON array of
  strings; timestamps RFC3339).
- Default (no `--format`) — plain tab-separated rows on stdout.
- Unknown `--format` exits 1 (user error). stdout carries data; stderr empty.

### `brag project show <name|id>` — show one project (STAGE-007)

​```
brag project show bragfile
brag project show 3 --format json
​```

Shows one project's name, status, state note, and locations. The argument is
resolved as a **name first**; if no project has that name and the argument is
a positive integer, it is resolved as a project **id**. (No recent-brag count
— that is `brag project status`, a later STAGE-007 spec.)

- Plain output is a labeled block (`Name:`, `Status:`, `State note:`,
  `Locations:`).
- `--format json` — a single JSON object (not an array) with the same element
  shape as `brag project list`.
- Exit 1 (user error) if no project matches the name or id, or on unknown
  `--format`.
```

(The `​` zero-width marks above are only to escape the nested fences in
this spec — the real doc uses plain triple-backtick fences.)

Then add to the **References** list (~line 504, beside the `DEC-016`
row):

```
- `DEC-017` — `entries.project` ↔ `projects` relationship (soft string match) + `projects.status` enum + single `state_note`; the data `brag project show`/`list` render.
```

### Gotchas

- **No SQL in `project.go`.** The LD3 path pre-check iterates
  `ListProjects()` *results* (plain Go); the id/name parse is `strconv`
  (allowed, like `delete.go`). Everything touching the DB is a Store call.
- **`--format` default is `""`, not `"json"`.** The check is
  `format != "" && format != "json"` (the SPEC-026 `--format` lesson) —
  an empty default must pass. Stated in LD4; do not default it to `json`.
- **Confirmation on stderr, data on stdout.** `new` → `Created project
  …` on **stderr**; `list`/`show` data on **stdout**. A success test
  asserts `errBuf`/`outBuf` separation (§9). `new` writes **nothing** to
  stdout.
- **`show` resolves name-first.** Try `GetProjectByName`; only on a
  `ErrNotFound` AND a positive-integer key fall back to `GetProject`.
  Other `GetProjectByName` errors are internal (exit 2), not user errors.
- **`ListProjects` order is the `list` order** — don't re-sort. It is
  `updated_at DESC, id DESC`; SPEC-028 never bumps `updated_at`, so for
  projects created here `created_at == updated_at` and the id tie-break
  decides (newest first).
- **`gofmt -w .` + `go vet ./...`** before the PR; confirm
  `./brag project --help` in the real binary (registration gap).

---

## Build Completion

*Filled in at the end of the **build** cycle, before advancing to verify.*

- **Branch:** `feat/spec-028-brag-project-new-list-show`
- **PR (if applicable):** pending
- **All acceptance criteria met?** yes
- **New decisions emitted:**
  - none (LD1–LD4 are localized CLI decisions; confidence ≥ 0.82; no DEC-018)
- **Deviations from spec:**
  - none — all literals transcribed verbatim; no unexpected implementation choices
- **Follow-up work identified:**
  - none beyond the already-listed SPEC-029..032 backlog

**Real-binary registration confirmed:**
`CGO_ENABLED=0 go build -o /tmp/brag-028 ./cmd/brag` → success.
`/tmp/brag-028 project --help` renders `Available Commands: list / new / show`.
`/tmp/brag-028 project new --help` shows `Examples:` + `--path` flag.
`/tmp/brag-028 project list --help` shows `Examples:` + `--format` flag.
`/tmp/brag-028 project show --help` shows `Examples:` + `--format` flag.
All render correctly; `NewProjectCmd()` is wired in `cmd/brag/main.go`.

### Build-phase reflection (3 questions, short answers)

Process-focused: how did the build go? What friction did the spec create?

1. **What was unclear in the spec that slowed you down?**
   — Nothing was unclear. The spec was a canonical literal-artifact-as-spec build: `GetProjectByName`, the two export helpers, and all four CLI functions were embedded verbatim under "Notes for the Implementer" with explicit imports, and the `--format` default/check pattern was called out explicitly in both LD4 and Gotchas. The one test I wrote with ad-hoc string manipulation (`TestProjectShow_ById`) was self-introduced complexity I immediately replaced with `strconv.FormatInt` — that was not a spec ambiguity, just a code quality reflex.

2. **Was there a constraint or decision that should have been listed but wasn't?**
   — No. The `no-sql-in-cli-layer`, stdout/stderr discipline, `--format ""` default, and orphan pre-check were all listed and exercised. The `getFlagString` helper being package-private in `add.go` (accessible to `project.go` as same-package) is the kind of "go read this file" detail the spec correctly points to rather than re-explaining.

3. **If you did this task again, what would you do differently?**
   — Write `TestProjectShow_ById` with `strconv.FormatInt` from the start rather than drafting a hand-rolled `formatInt64`. Otherwise nothing: the spec's literal-artifact approach made the build mechanical and fast.

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
