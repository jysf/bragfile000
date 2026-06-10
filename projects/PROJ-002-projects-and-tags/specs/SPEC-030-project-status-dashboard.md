---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-030
  type: story                      # epic | story | task | bug | chore
  cycle: design                    # frame | design | build | verify | ship
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
  decisions: [DEC-017, DEC-011, DEC-014, DEC-006, DEC-007, DEC-005]
  constraints:
    - no-sql-in-cli-layer
    - stdout-is-for-data-stderr-is-for-humans
    - storage-tests-use-tempdir
    - errors-wrap-with-context
    - timestamps-in-utc-rfc3339
    - test-before-implementation
    - one-spec-per-pr
    - no-cgo
    - no-new-top-level-deps-without-decision
  related_specs: [SPEC-027, SPEC-028, SPEC-029, SPEC-031, SPEC-032, SPEC-026]
---

# SPEC-030: project status dashboard

## Context

This spec ships `brag project status` — the **"scannable" success
criterion** of STAGE-007: *"`brag project status` lists active projects
by recency, each showing its state/next-action note and a recent-brag
count."* It is the read-only dashboard half of the Projects surface,
built on the SPEC-027 schema (`projects` + `project_locations`,
`0004_add_projects`), the SPEC-028 CLI scaffolding (`NewProjectCmd`
parent + `export.projectRecord` JSON family), and the SPEC-029 mutation
methods (`UpdateProject`/`ArchiveProject` — which is what makes
`updated_at` recency ordering observable and what produces the
`archived` projects this dashboard excludes).

It is also the **first consumer of DEC-017's soft string join.** SPEC-027
chose soft string match (`entries.project` stays free text, joined to
`projects.name` opportunistically at query time, zero backfill) and
DEC-017's own Validation section names this spec as the proof: *"SPEC-030's
recent-brag count via `entries.project = projects.name` reads naturally
and needs no schema change."* This spec realizes exactly that join.

Two design questions STAGE-007 surfaced are resolved here (both as
**localized Locked Design Decisions, no new DEC** — see Implementation
Context):

1. **Q1 — what is the "recent-brag count"?** Resolved: a **total
   (lifetime) count** of entries whose `project` string equals the
   project name. Recency is already carried by the dashboard's
   `updated_at DESC` ordering and each project's `status`; the count's
   honest job is lifetime capture volume. (LD1.)
2. **Q2 — `project-state-note-shape` (open question in
   `questions.yaml`).** This dashboard was filed as the consumer that
   would decide whether `state_note` needs structure (`state` +
   `next_action`) or stays a single free-text column. Resolved: **stays a
   single free-text column.** A one-row-per-project dashboard needs *less*
   structure than `brag project show` (which already renders the note as
   one labeled line and reads fine), not more. The open question is closed
   in `questions.yaml` (`status: answered`); **no migration `0005`, no
   DEC-019.** (LD3.)

Parent stage: `STAGE-007-projects-core.md` (SPEC-030 is the fourth of
seven specs; SPEC-027/028/029 shipped). Project: `PROJ-002`
(`projects-and-tags`).

## Goal

Add a `brag project status` subcommand to the existing `brag project`
parent that lists every **non-archived** project ordered
`updated_at DESC, id DESC`, each row carrying its status, a **total brag
count** (entries whose free-text `project` equals the project name —
the DEC-017 soft string join), and its `state_note`; in plain
(tab-separated) form by default and as a DEC-011 naked JSON array under
`--format json`. Read-only: **no migration, no schema change, no
mutation** — one new Store read method, one export helper, one cobra
subcommand.

## Inputs

- **Files to read:**
  - `internal/storage/project.go` — the `Project` struct + the SPEC-027/029
    methods; the new `ProjectStatus` struct + `ProjectStatuses()` method
    sit here (one concept per file). Mirror `ListProjects`' scan/parse
    shape; the brag-count join is the one new piece.
  - `internal/storage/store.go` — `List`'s `e.project = ?` filter (the
    same free-text column the soft join reads) and the simple-query style;
    `Add` for the `entries` columns.
  - `internal/storage/entry.go` — `Entry.Project` (the free-text column)
    + `ListFilter.Project` (`e.project = ?`).
  - `internal/cli/project.go` — the `NewProjectCmd` parent, the
    `newProjectListCmd`/`runProjectList` `--format`-switch shape, and the
    open/resolve pattern (`getFlagString` → `config.ResolveDBPath` →
    `storage.Open` → `defer s.Close()`) the new `status` subcommand clones.
  - `internal/export/project.go` — `projectRecord` + `ToProjectsJSON`
    (the DEC-011 naked-array, 2-space-indent, `[]`-not-`null` marshal);
    the new `projectStatusRecord` + `ToProjectStatusesJSON` sit beside them.
  - `internal/cli/project_test.go` — the `newProjectTestRoot`/`runProjectCmd`
    + buffer-separation CLI test patterns; `TestProjectCmd_BarePrintsHelp`
    (gets `"status"` added — see Outputs).
  - `internal/storage/project_test.go` — `newTestStore(t)`; the
    `ListProjects` ordering test pattern the status-ordering test mirrors.
  - `decisions/DEC-017` (the soft string match the count realizes),
    `DEC-011` (naked array), `DEC-014` (single-document family — context
    for why `status` is an array like `list`, not a single object like
    `show`), `DEC-006`/`DEC-007` (cobra + `RunE` validation),
    `DEC-005` (autoincrement PK → the `id DESC` tie-break).
  - `guidance/questions.yaml` — `project-state-note-shape` (resolved here).
- **External APIs:** none. Plain Store calls; no new dependency.
- **Related code paths:** `internal/storage/project.go` (one read method),
  `internal/export/project.go` (one JSON helper), `internal/cli/project.go`
  (one subcommand), `docs/api-contract.md` (one new section + one stale
  forward-reference fix). **No `cmd/brag/main.go` change** — `status`
  rides the already-registered `NewProjectCmd` parent (SPEC-028).

## Outputs

- **Files created:** none. Every change folds into an existing file.
- **Files modified:**
  - `internal/storage/project.go` — add the `ProjectStatus` struct + the
    `ProjectStatuses()` read method (the soft-join dashboard query).
  - `internal/storage/project_test.go` — add the storage failing tests
    below (additive; no rewrite of SPEC-027/028/029 tests).
  - `internal/export/project.go` — add `projectStatusRecord` +
    `ToProjectStatusesJSON` (beside `projectRecord`/`ToProjectsJSON`).
  - `internal/export/project_test.go` — add the export failing tests below
    (additive).
  - `internal/cli/project.go` — add `newProjectStatusCmd()` +
    `runProjectStatus` + the `truncateStateNote` plain-render helper;
    register `status` on the parent in `NewProjectCmd()` and add it to the
    parent `Short`.
  - `internal/cli/project_test.go` — add the CLI failing tests below; and
    **modify `TestProjectCmd_BarePrintsHelp`** to add `"status"` to its
    subcommand-presence list (additive — the test asserts presence, not
    exhaustiveness, so this is the dashboard's registration assertion).
  - `docs/api-contract.md` — **status-change UPDATE** (see audit): add the
    `### brag project status` section after the `show` section; fix the
    now-stale "a later STAGE-007 spec" forward-reference in the `show`
    section.
- **New exports:**
  - `storage.ProjectStatus` struct (`ID`, `Name`, `Status`, `StateNote`,
    `BragCount`, `CreatedAt`, `UpdatedAt`).
  - `func (s *Store) ProjectStatuses() ([]ProjectStatus, error)`
  - `func export.ToProjectStatusesJSON(statuses []storage.ProjectStatus) ([]byte, error)`
  - `func cli.NewProjectCmd()` gains a `status` subcommand (the
    constructor signature is unchanged).
- **Database changes:** **NONE.** No migration. The brag count is a
  read-time join on the existing `entries.project` and `projects.name`
  columns (DEC-017 soft string match). This is a load-bearing property
  (it is why the count-bump premise-audit case is empty), not an omission.

### Premise audit (`projects/_templates/premise-audit.md`), run at design

Greps were **run** at design **2026-06-09** and reconciled against the
lists below.

```
- [x] Inversion/removal: greps run — NONE (new read-only command + new Store method; no schema/behavior change)
- [x] Addition/count-bump: greps run — NONE (no migration; no count-coupled assertion broken)
- [x] Status-change: greps run, every doc hit listed as updates/stays
- [x] Cross-check: actual grep hits reconciled against the lists above
```

**1. Inversion / removal — NONE.** `brag project status` is a brand-new
command and `ProjectStatuses()` a brand-new Store method; no existing
behavior is inverted, no flag/column removed. DEC-017's soft string
match leaves `entries.project` free text, so the new `e.project = p.name`
join **reads** the column exactly as `List`'s `e.project = ?` filter
already does — it changes nothing about `list --project` / `ByProject` /
`GroupEntriesByProject`. Grep run:
`grep -rn 'e.project = \|ListFilter\|ByProject\|func.*ProjectStatus' internal/`
→ hits are `store.go:306` (`e.project = ?`, unchanged), the aggregate
`ByProject`/`GroupEntriesByProject` (untouched), and their tests.
**Reconciliation: zero rewrites, zero deletions.**

**2. Addition / count-bump — NONE.** No migration is added, so
`schema_migrations` is untouched (SPEC-027's count of **4** stands).
Grep run: `grep -rn '0004_add_projects\|count != 4\|want 4' internal/storage/*_test.go`
→ all hits (`store_test.go`, `fts_test.go:149,269`,
`project_migration_test.go:116,150`) are the SPEC-027 sites already at 4;
none is touched by SPEC-030. **Subcommand-set coupling:** grep run
`grep -rn 'Commands()\|\[\]string{"new", "list", "show"' internal/cli`
→ (a) `list_test.go:629` iterates a **test-local** root built with only
`NewListCmd()` to read the `list` Short — not coupled to the `project`
subcommand set; (b) `project_test.go:49` (`TestProjectCmd_BarePrintsHelp`)
asserts the **presence** of `new`/`list`/`show` (not exhaustiveness), so
adding `status` does not break it — this spec *adds* `"status"` to that
list as the dashboard's registration assertion (see Outputs). No
literal-count assertion is broken.

**3. Status change — the new command.** Grep run:
`grep -rln -i 'project status\|brag project' docs/ README.md` → two files
carry a claim this spec touches:

- **Updates (this spec):**
  - `docs/api-contract.md` — **UPDATE.**
    - **Add** a `### brag project status` section after `### brag project
      show <name|id>` (line ~501) and before `### brag project edit`
      (line ~503) — house style matches the adjacent project sections
      (fenced example + bullets); see the literal in Notes.
    - **Fix the stale forward-reference** in the `show` section
      (line ~494): `"that is `brag project status`, a later STAGE-007
      spec.)"` → `"that is `brag project status` (below).)"` — `status`
      now ships, so "a later spec" is a stale status claim.
    - References list (~line 611): `DEC-017` row already present — **stays**
      (no new DEC; no new reference needed).
- **Stays here (STAGE-008, or no status claim invalidated):**
  - `docs/data-model.md` — **STAYS.** SPEC-027 added the `projects`/
    `project_locations` tables and the `state_note` description already
    forward-references `brag project status (SPEC-030)`. SPEC-030 adds
    **no schema and no column**, so data-model carries no status claim
    this spec invalidates. (The `state_note` row's "single free-text
    state/next-action note" wording is **confirmed correct** by Q2's
    resolution — no edit needed.)
  - `docs/tutorial.md`, `docs/architecture.md`, `README.md` — **STAY.**
    The comprehensive projects+tags tutorial and the architecture
    diagram/responsibilities refresh are the **STAGE-008** sweep (brief +
    stage Scope: "only per-spec, premise-audit-driven doc updates fold in
    here"). Remaining `project` hits reference the existing free-text
    `entries.project` (untouched by DEC-017) or are historical/narrative.

**4. Cross-check.** Actual grep hits reconciled against the lists above
at design; no un-enumerated hit remained. (Build-side: re-run
`grep -rln -i 'project status\|brag project' docs/ README.md` before the
doc edit and treat any delta as a question, per the premise-audit
cross-check rule.)

## Acceptance Criteria

Testable outcomes. Happy path, error cases, edge cases.

- [ ] **Counts via the DEC-017 soft join.** `ProjectStatuses()` returns,
  for each project, `BragCount` = the number of `entries` rows whose
  `project` string **exactly equals** the project name. A project with
  zero matching entries has `BragCount == 0` (not omitted). (AC ↔
  `TestProjectStatuses_CountsAndExcludesArchived`,
  `TestProjectStatuses_SoftJoinIgnoresUnregisteredAndNull`.)
- [ ] **Archived projects are excluded.** A project with
  `status == "archived"` does **not** appear in `ProjectStatuses()` even
  if entries reference its name; `active`/`paused`/`done` all appear.
  (↔ `TestProjectStatuses_CountsAndExcludesArchived`.)
- [ ] **Ordering is recency, then id tie-break.** `ProjectStatuses()`
  returns rows ordered `updated_at DESC, id DESC`. Editing a project
  (bumping `updated_at` via `UpdateProject`) floats it to the top; among
  same-`updated_at` rows the newest `id` is first. (↔
  `TestProjectStatuses_OrderedByUpdatedAtThenIDDesc`.)
- [ ] **Empty is a non-nil empty slice.** With no non-archived projects,
  `ProjectStatuses()` returns a non-nil slice of length 0 (so JSON renders
  `[]`). (↔ `TestProjectStatuses_EmptyReturnsNonNilSlice`,
  `TestProjectStatus_EmptyJSONIsBrackets`.)
- [ ] **`status` plain shape + ordering.** `brag project status` prints
  one tab-separated `<name>\t<status>\t<brag_count>\t<state_note>` row per
  non-archived project to **stdout**, ordered `updated_at DESC, id DESC`;
  stderr empty; exit 0. (↔ `TestProjectStatus_ListsActiveWithCount`,
  `TestProjectStatus_OrderedByRecency`,
  `TestProjectStatus_ExcludesArchived`.)
- [ ] **State note truncated in plain, blank when empty.** In plain
  output a `state_note` longer than 50 runes is truncated to 50 runes +
  `…`; an empty note renders as an empty final column. (↔
  `TestProjectStatus_StateNoteTruncatedInPlain`,
  `TestProjectStatus_ListsActiveWithCount`.)
- [ ] **`status --format json`.** Emits a **naked JSON array** (DEC-011;
  2-space indent) of status objects, keys `id, name, status, state_note,
  brag_count, created_at, updated_at` in that order; `brag_count` a JSON
  number; `state_note` carried **in full** (never truncated); timestamps
  RFC3339; same order as plain; an empty dashboard emits `[]` (never
  `null`). (↔ `TestProjectStatus_JSON`,
  `TestProjectStatus_EmptyJSONIsBrackets`,
  `TestToProjectStatusesJSON_*`.)
- [ ] **`--format` default is plain; unknown is a user error.** No
  `--format` (default `""`) → plain rows; `--format json` → the array;
  any other value (e.g. `xml`) → user error (exit 1). (↔
  `TestProjectStatus_UnknownFormatErrUser`.)
- [ ] **stdout/stderr discipline.** Dashboard data (plain rows or JSON)
  goes to **stdout**; nothing routine goes to stderr (a success test
  asserts no cross-leakage, per §9). (↔
  `TestProjectStatus_StdoutDiscipline`.)
- [ ] **Help renders.** `brag project status --help` contains `Examples:`
  and the distinctive token `brag project status`. (↔
  `TestProjectStatus_HelpShowsExamples`.)
- [ ] **Registration.** `brag project` bare help lists `status` among its
  subcommands (the subcommand is wired into `NewProjectCmd`). (↔
  `TestProjectCmd_BarePrintsHelp`, with `"status"` added.)
- [ ] **No regressions.** No SQL under `internal/cli/`; no migration;
  `go test ./...`, `gofmt -l .`, `go vet ./...`, and
  `CGO_ENABLED=0 go build ./...` are clean; output shapes
  (DEC-011/013/014) and search (DEC-010) stay byte-stable (untouched).

## Failing Tests

Written at **design**; build makes them pass. All are **new/additive**
(the premise audit found zero inversions). Storage tests use `t.TempDir()`
(`storage-tests-use-tempdir`). CLI tests use the `newProjectTestRoot` +
`runProjectCmd` patterns; entries are seeded by opening a second
`storage.Open(dbPath)` handle and calling `Add` (the same file-backed
temp DB the CLI command opens).

### `internal/storage/project_test.go` (modify — additive)

- `"TestProjectStatuses_CountsAndExcludesArchived"` — create projects
  `bragfile`(active), `platform`(paused), `sideproj`(done),
  `oldthing`(set `archived` via `ArchiveProject`); `Add` entries with
  `Project` = `bragfile`×2, `platform`×1, `oldthing`×1. **Asserts:**
  `ProjectStatuses()` returns exactly 3 rows (no `oldthing`); the row for
  `bragfile` has `BragCount == 2`, `platform == 1`, `sideproj == 0`; no
  returned row has `Status == "archived"`.
- `"TestProjectStatuses_SoftJoinIgnoresUnregisteredAndNull"` — create one
  project `bragfile`; `Add` entries with `Project` = `bragfile`,
  `bragfile-old` (no project row), `""` (empty), and one entry with no
  `Project` field set. **Asserts:** the `bragfile` row's `BragCount == 1`
  — only the exact-string match counts; the unregistered/empty/absent
  project strings inflate nothing (the soft join is exact-match per
  DEC-017).
- `"TestProjectStatuses_OrderedByUpdatedAtThenIDDesc"` — create `p1`,
  `p2`, `p3` (same second → `created_at == updated_at`). **Sub-case A
  (tie-break):** with no mutation, assert the returned order is
  `p3, p2, p1` (id DESC, the §9 monotonic tie-break under an RFC3339
  same-second tie). **Sub-case B (recency):** `UpdateProject(p1.ID, …)`
  to bump `p1`'s `updated_at`, then assert `p1` is now **first** (recency
  ordering is observable because SPEC-029's `UpdateProject` advances
  `updated_at` — the case SPEC-027's ordering test deferred to a
  mutation-bearing spec). *Avoids `sleep`: the `updated_at` advance comes
  from the Store stamping a fresh `now` on update, and the tie-break is
  proven by id, not by a timestamp inequality (§9).*
- `"TestProjectStatuses_EmptyReturnsNonNilSlice"` — fresh store, no
  projects. **Asserts:** `ProjectStatuses()` returns `err == nil` and a
  **non-nil** slice with `len == 0` (so the JSON layer renders `[]`).

### `internal/export/project_test.go` (modify — additive)

- `"TestToProjectStatusesJSON_NakedArrayShape"` — one `ProjectStatus`
  with `BragCount: 7`. **Asserts:** marshals to a naked array
  (`HasPrefix(trim, "[")`) whose single element has the seven keys in
  order `id, name, status, state_note, brag_count, created_at,
  updated_at`; `brag_count` renders as the JSON number `7` (not a
  string); 2-space indent (`strings.Contains(out, "\n  {")` and
  `"    \"id\""`).
- `"TestToProjectStatusesJSON_EmptyIsBracketsNotNull"` —
  `ToProjectStatusesJSON(nil)` and `ToProjectStatusesJSON([]storage.ProjectStatus{})`
  both yield exactly `[]` (DEC-011 empty discipline).
- `"TestToProjectStatusesJSON_StateNoteNotTruncated"` — a
  `ProjectStatus` whose `StateNote` is 80 characters long. **Asserts:**
  the marshaled JSON contains the **full** 80-char note (truncation is a
  plain-output concern only; JSON is the scripting path and carries the
  complete value).

### `internal/cli/project_test.go` (modify — additive + one assertion add)

- `"TestProjectCmd_BarePrintsHelp"` (**modify**) — add `"status"` to the
  existing `[]string{"new", "list", "show"}` presence list so the bare
  `brag project` help is asserted to list the `status` subcommand. (The
  test already asserts presence, not exhaustiveness; this is purely
  additive.)
- `"TestProjectStatus_ListsActiveWithCount"` — `project new bragfile
  --path /x`; seed 2 entries with `Project:"bragfile"` via a second
  `storage.Open(dbPath)` handle; `project status`. **Asserts:** stdout
  has exactly one row, tab-separated, fields `bragfile`, `active`, `2`,
  and an **empty** final state-note column (the project was created with
  the default empty `state_note`); stderr empty; exit 0.
- `"TestProjectStatus_ExcludesArchived"` — `new a --path /a`, `new b
  --path /b`, then `project archive b`; `project status`. **Asserts:**
  stdout lists `a` and **not** `b`.
- `"TestProjectStatus_OrderedByRecency"` — `new p1`, `new p2`, `new p3`
  (distinct paths); then `project edit p1 --state-note "touched"` (bumps
  `p1.updated_at`); `project status`. **Asserts:** the first stdout row is
  `p1` (most-recently-updated first); without the edit the order would be
  `p3, p2, p1` (id DESC) — the edit floats `p1` to the top.
- `"TestProjectStatus_StateNoteTruncatedInPlain"` — `new bragfile
  --path /x`; `project edit bragfile --state-note "<a 60-char note>"`;
  `project status`. **Asserts:** the plain row's state-note column is the
  first 50 runes + `…` (length 51 display runes), **not** the full
  60-char note.
- `"TestProjectStatus_JSON"` — `new bragfile --path /x`; seed 3 entries
  with `Project:"bragfile"`; set a long state note via `project edit
  bragfile --state-note "<70-char note>"`; `project status --format json`.
  **Asserts:** stdout is a naked array (`HasPrefix(trim,"[")`); the
  element unmarshals with `brag_count == 3` and the **full 70-char**
  `state_note` (untruncated); 2-space indent
  (`strings.Contains(out, "    \"id\"")`); stderr empty.
- `"TestProjectStatus_EmptyJSONIsBrackets"` — no projects;
  `project status --format json` → stdout is exactly `[]`;
  `project status` plain → empty stdout. (Both exit 0, stderr empty.)
- `"TestProjectStatus_UnknownFormatErrUser"` — `project status --format
  xml` → `errors.Is(err, ErrUser)`.
- `"TestProjectStatus_StdoutDiscipline"` — on a populated dashboard, the
  rows are on **stdout** and stderr is empty (§9 no-cross-leakage).
- `"TestProjectStatus_HelpShowsExamples"` — `project status --help`
  contains `Examples:` and the distinctive token `brag project status`
  (positive substring; the §12 NOT-contains caveat does not apply).

> **Locked-decision ↔ test traceability (§9).** Each Locked Design
> Decision (## Implementation Context) has a paired failing test:
> **LD1** (total brag count via soft join) →
> `TestProjectStatuses_CountsAndExcludesArchived` +
> `TestProjectStatuses_SoftJoinIgnoresUnregisteredAndNull`;
> **LD2** (non-archived filter + recency order) →
> `TestProjectStatuses_CountsAndExcludesArchived` (exclusion) +
> `TestProjectStatuses_OrderedByUpdatedAtThenIDDesc` +
> `TestProjectStatus_ExcludesArchived` + `TestProjectStatus_OrderedByRecency`;
> **LD3** (single free-text note; plain truncates, JSON full) →
> `TestProjectStatus_StateNoteTruncatedInPlain` +
> `TestToProjectStatusesJSON_StateNoteNotTruncated` +
> `TestProjectStatus_JSON`;
> **LD4** (array shape / `brag_count` number / locations omitted /
> `--format ""` default) → `TestProjectStatus_JSON` +
> `TestToProjectStatusesJSON_NakedArrayShape` +
> `TestProjectStatus_EmptyJSONIsBrackets` +
> `TestProjectStatus_UnknownFormatErrUser`;
> **LD5** (single LEFT JOIN GROUP BY method, no N+1, no location hydrate)
> → `TestProjectStatuses_*` (the storage tests exercise the method
> directly).

> **§12(a)/(b) note for build:** the dashboard SQL was executed against
> the real driver at design (see "§12(b) design-time verification" in
> Notes): archived-excluded, `COUNT(e.id)` → 0 for no-match, exact-string
> soft join, `updated_at DESC, id DESC` ordering all confirmed. The JSON
> element shape follows the SPEC-028 `projectRecord` pattern (already
> §12(b)-validated). Transcribe the literals; do not re-derive.

## Implementation Context

*Read this section and the files it points to before build. This is the
whole handoff — the build session has only this spec.*

### Decisions that apply

- **`DEC-017`** (shipped, SPEC-027) — **the soft string match this spec
  realizes.** `entries.project` is free text; "entries of a project" is
  `entries.project = projects.name` at query time, **no FK, no link
  column, no backfill.** The brag count is exactly this join. DEC-017's
  Validation names SPEC-030 as the proof the count "reads naturally and
  needs no schema change." Consequence to honor: the join is
  **exact-match only** (no case-folding, no fuzzy) — accepted at personal
  scale. The `status` enum is `active|paused|done|archived`; this
  dashboard excludes `archived`.
- **`DEC-011`** — naked JSON array, field-names-match-columns, 2-space
  indent, `[]`-not-`null` on empty. `brag project status --format json`
  follows it verbatim; the element adds `brag_count` (a computed number)
  to the project's identity fields.
- **`DEC-014`** (context, not borrowed-from) — the single-document
  envelope is for the aggregating digests; `status` is a **multi-row
  index** like `brag project list`, so it is a DEC-011 **array**, not a
  DEC-014 single object. (`show` is the single-object case; `status` and
  `list` are arrays.)
- **`DEC-006`** — cobra: `status` is a `*cobra.Command` built by a
  `new…Cmd()` constructor and attached to the `NewProjectCmd` parent (no
  `main.go` change — the parent is already registered, SPEC-028).
- **`DEC-007`** — `--format` value validation lives **inline in `RunE`**
  via `UserErrorf` (mirror `runProjectList`'s
  `format != "" && format != "json"` check).
- **`DEC-005`** — INTEGER autoincrement PK gives the `id DESC` tie-break
  the dashboard orders on under a same-second `updated_at` tie.

### Locked design decisions

The two STAGE-007 design questions surfaced for this spec, plus three
output/shape choices, decided here with reasoning. **No new DEC** — each
is an application of DEC-017 + the brief + existing output-shape DECs,
localized to SPEC-030; honest confidence on each is ≥ 0.8, so no
`/guidance/questions.yaml` entry is filed (§14). Q2's resolution
*closes* the pre-existing `project-state-note-shape` question (see
"Resolving the open question" below).

- **LD1 — the brag count is a TOTAL (lifetime) count, not time-windowed.**
  Confidence 0.85. The count is
  `COUNT(entries WHERE entries.project = projects.name)` over all time.
  *Rationale:* this is a **status dashboard for self-assessment, not
  analytics**. "Is this project active at a glance?" is answered by three
  signals the dashboard already carries — the `updated_at DESC`
  **ordering** (project-record recency), the **`status`** field
  (`active`/`paused`/`done`), and the **brag count** (how much I have
  captured against it). The count's honest job is lifetime capture
  *volume*; recency is not its job and is already covered twice over.
  *Rejected alternatives (build-time):* **(B) a time-windowed count**
  (e.g. last 30 days) — rejected on two grounds: a hardcoded window
  silently shows `0` for a project you logged ten brags against two
  months ago, which is *misleading* for "is this active" (the project's
  own `updated_at`/`status` already encode staleness), and a configurable
  `--since` flag turns a glance-dashboard into an analytics query
  (out of scope; STAGE-007 brief excludes analytics/time-series). **(C)
  both total + windowed** ("12 total, 3 in 30d") — rejected as
  double-encoding recency (already in the ordering) and cluttering a
  one-row-per-project scannable view with a second number that needs a
  legend. Total is the one honest number. *Exact SQL is in LD5 / Notes.*

- **LD2 — the dashboard shows NON-ARCHIVED projects (`status != 'archived'`),
  ordered `updated_at DESC, id DESC`.** Confidence 0.85. "Active" in the
  success criterion means **non-archived** — `active`, `paused`, and
  `done` all appear; only `archived` (the SPEC-029 `ArchiveProject` flip,
  the explicit "set aside / out of view" state) is hidden. *Rationale:*
  `archived` is the one status whose entire purpose is "remove from the
  working view" (DEC-018 frames archive as the recoverable
  out-of-the-way flip); `paused`/`done` are still things you may want to
  scan (a paused project is a candidate to resume; a done project is
  recent context). Ordering is recency-first with the §9 monotonic
  `id DESC` tie-break — the SPEC-029 `UpdateProject`/`ArchiveProject`
  methods are what make `updated_at` advance, so recency ordering is now
  *observable* (it was laid down but untestable in SPEC-027). *Rejected
  alternative:* `status == 'active'` only — rejected as too narrow; it
  would hide `paused`/`done` projects the user is plausibly scanning for,
  and "active projects" in the brief is the colloquial "not archived,"
  not the literal enum value (the enum has four values; the dashboard
  hides exactly one).

- **LD3 — `state_note` stays a SINGLE free-text column; plain output
  truncates it, JSON carries it in full.** Confidence 0.85. This
  **resolves the `project-state-note-shape` open question to "keep
  single."** *Rationale:* a one-row-per-project dashboard needs *less*
  structure than `brag project show`, which already renders the note as a
  single labeled `State note:` line (SPEC-028) and reads fine; a dashboard
  row is strictly terser, so if free text suffices for `show` it suffices
  here. Splitting into `state` + `next_action` would either need two
  columns (cramping a terminal row) or be concatenated back to one display
  string (defeating the split). The user's "Shipped tags; next: cut
  v0.2.0" convention is a free-text habit, not something the dashboard
  parses. Under forward-only migrations (DEC-002) a later split is a cheap
  additive column + one-shot copy, not a rebuild (DEC-017 + questions.yaml
  both note this) — so holding the single column is YAGNI-correct until a
  renderer genuinely needs structure, and this renderer does not.
  *Plain truncation:* a `state_note` longer than **50 runes** is shown as
  its first 50 runes + `…` (rune-based, so multibyte notes are never split
  mid-character); an empty note renders as an empty final column.
  *JSON is never truncated* — it is the scripting path and must carry the
  exact stored value. *Rejected alternative (B):* split into structured
  columns via a `0005` migration amending DEC-017 — rejected because the
  dashboard, the consumer that was meant to force the question, does not
  need structure; building it would be speculative schema work. **No
  migration, no DEC-019.**

- **LD4 — output shapes.** Confidence 0.85.
  - **Plain** (`brag project status`, default): one
    `<name>\t<status>\t<brag_count>\t<state_note>` row per non-archived
    project, tab-separated like `brag project list`/`brag list`/`brag
    tags`, in `ProjectStatuses()` order (`updated_at DESC, id DESC`).
    `brag_count` is the decimal integer; `state_note` is the truncated
    note (LD3), empty column when blank. No header row (matches every
    other plain command). No `--format` → plain.
  - **`--format json`**: a **naked JSON array** of status objects
    (DEC-011); `[]` on empty, never `null`. **Element shape:** keys
    `id, name, status, state_note, brag_count, created_at, updated_at` in
    that order; `brag_count` a JSON **number**; `state_note` the **full**
    value; timestamps RFC3339. Field names match the `projects` columns
    (plus `brag_count`).
  - **`locations` is intentionally OMITTED** from both plain and JSON.
    The dashboard answers "what am I working on / what state / how much
    have I logged" — *where* a project lives is `brag project show`'s job.
    Omitting locations also keeps `ProjectStatuses()` a single query with
    **no per-row location hydration** (LD5). *Rejected alternative:* reuse
    SPEC-028's `projectRecord` (which carries `locations`) and add
    `brag_count` — rejected because it would force the N+1 location
    hydration `ListProjects` does, for data the dashboard does not show.
  - **`--format` default is `""`** (empty → plain); the only accepted
    non-empty value is `json`; anything else → `UserErrorf`. *Stated
    explicitly per the STAGE-006/SPEC-026 flag-default WATCH item: each
    flag's DEFAULT is locked, not just its accepted set.* (This is the
    third CLI `--format` spec to state its default explicitly — SPEC-026,
    SPEC-028, now SPEC-030; see "Flag-default WATCH" in Notes. **Do not
    codify mid-stage** — it is a stage-close note.)

- **LD5 — `ProjectStatuses()` is ONE `LEFT JOIN … GROUP BY` query, not
  `ListProjects` + per-row counts.** Confidence 0.85. The brag count is a
  genuine aggregation, so a single grouped query is the natural and
  correct shape — *unlike* the one-to-many `Locations` hydration, where
  SPEC-027 deliberately chose per-row queries. A `LEFT JOIN` (so
  zero-brag projects still appear) with `COUNT(e.id)` (which yields `0`
  for the null-join row, where `COUNT(*)` would wrongly yield `1`) is the
  exact, validated form. The method returns a dedicated **flat**
  `ProjectStatus` struct (not the `Project` struct) so there is no
  unpopulated `Locations` field to mislead a caller. *Rejected
  alternative:* `ListProjects()` then a per-name `CountEntries(name)`
  call — rejected as N+1 for a query SQLite expresses in one statement,
  and it would carry the unused `Locations`. The single query is in Notes
  and was §12(b)-pre-flighted.

#### Resolving the open question (`project-state-note-shape`)

LD3 resolves the `project-state-note-shape` question (filed at SPEC-027
design, DEC-017, the one DEC-017 sub-choice below 0.8). Per the
questions.yaml header convention (and matching the resolved
`tags-storage-model` precedent), this spec sets the question to
`status: answered` with `answered_by: SPEC-030 (2026-06-09)` and a
resolution note: *the dashboard — the named consumer — renders the note
fine as a single free-text column (truncated in plain, full in JSON);
no split needed; a later split stays a cheap additive migration if a
future renderer ever needs structure.* No DEC is emitted (DEC-017 already
records the single-column choice; this confirms it rather than amending
it).

### Constraints that apply

(see `/guidance/constraints.yaml` for full text)

- `no-sql-in-cli-layer` (**blocking**) — the new SQL (the dashboard join)
  lives **only** in `internal/storage/project.go`. `project.go` (CLI)
  imports `storage`/`config`/`export`/`cobra` (+ stdlib); **never**
  `database/sql`. The CLI subcommand calls `ProjectStatuses()` only.
- `stdout-is-for-data-stderr-is-for-humans` (**blocking**) — dashboard
  rows + JSON go to **stdout**; there is no routine stderr output (no
  confirmation line — `status` is read-only). Errors go to stderr via the
  root error path.
- `storage-tests-use-tempdir` (**blocking**) — every storage test uses
  `t.TempDir()` (via `newTestStore`); CLI tests open their seed handle on
  the same `t.TempDir()` path; never touch `~/.bragfile`.
- `errors-wrap-with-context` (warning) — `ProjectStatuses()` wraps like
  `ListProjects` (`fmt.Errorf("project statuses: %w", err)`); the CLI
  wraps non-user errors (`fmt.Errorf("open store: %w", err)`).
- `timestamps-in-utc-rfc3339` (**blocking**) — `created_at`/`updated_at`
  are parsed UTC and re-emitted RFC3339 exactly like `ListProjects` /
  `ToProjectsJSON`. The dashboard reads timestamps; it never writes them.
- `test-before-implementation` (blocking) — the Failing Tests above are
  the design deliverable.
- `one-spec-per-pr` (blocking) — the PR references SPEC-030 only.
- `no-cgo` / `no-new-top-level-deps-without-decision` — pure-Go path; no
  new dependency.

### Design-time pre-flight (§12(b)) — run at design 2026-06-09, results below

- **Dashboard SQL** — the exact query in Notes was executed against the
  real driver (`modernc.org/sqlite`) in a scratch `internal/storage`
  test on a populated DB (4 projects: active/paused/done/archived;
  entries matching by free-text project string, plus unregistered-string,
  empty, and absent-project entries). Confirmed: **archived excluded**
  (3 rows, no `archived`); `COUNT(e.id)` → `bragfile=2, platform=1,
  sideproj=0` (LEFT JOIN keeps the zero-brag project; unregistered/empty/
  absent project strings inflate nothing); **ordering** `updated_at DESC,
  id DESC` with same-second `created_at` resolved by the `id DESC`
  tie-break (`sideproj`, `platform`, `bragfile`). ✅ The scratch test was
  removed after the pre-flight; `go test ./internal/storage/` stayed
  green.
- **JSON shape** — `ToProjectStatusesJSON` is a near-clone of SPEC-028's
  §12(b)-validated `ToProjectsJSON` (naked array, 2-space indent,
  `[]`-not-`null`); the only element difference is `brag_count` (an `int`
  → JSON number) replacing `locations`. The empty-`[]` and 2-space-indent
  discipline carry over verbatim.
- **Plain render literals** — the row (`%s\t%s\t%d\t%s\n`) and the
  truncation (`[]rune`, 50 + `…`) are this spec's own deterministic
  `fmt`/string output; the failing-test assertions match them exactly
  (§12(a) self-check).

### Dev/prod DB isolation (PROJ-002 brief) — still mandatory this stage

The schema is at v0.2.x (post-`0004`) from SPEC-027. SPEC-030 adds **no
migration**, but the rule stands: build/run the dev binary against a dev
DB (`BRAGFILE_DB=~/.bragfile-dev/db.sqlite` or `--db`); **never open the
production `~/.bragfile/db.sqlite`** with a v0.2.x binary. All tests use
`t.TempDir()`.

### Prior related work

- `SPEC-027` (shipped, PR #40) — the schema, `Project` struct, the four
  read primitives, **DEC-017** (the soft join this spec is the first
  consumer of). Its ordering test deferred the recency case to a
  mutation-bearing spec; SPEC-030's `TestProjectStatuses_OrderedByUpdatedAtThenIDDesc`
  sub-case B is that case (using SPEC-029's `UpdateProject`).
- `SPEC-028` (shipped, PR #41) — `NewProjectCmd` parent (already
  registered in `main.go` — so `status` needs **no** `main.go` change),
  the `export.projectRecord`/`ToProjectsJSON` family the new helper
  clones, the `--format`-switch + buffer-separation test patterns.
- `SPEC-029` (shipped, PR #42) — `UpdateProject` (bumps `updated_at`,
  making recency ordering observable) and `ArchiveProject` (produces the
  `archived` rows this dashboard excludes). The flag-default WATCH
  reached N=3 candidacy at SPEC-029 ship; SPEC-030 is another `--format`
  data point (stage-close note, not codified here).

### Out of scope (for this spec specifically)

If any of these feels necessary during build, **stop and flag** — do not
expand this spec.

- **`brag project here` cwd auto-detect** → **SPEC-031** (the cwd
  resolver against `project_locations`).
- **`brag add` `--project` auto-fill from cwd** → **SPEC-032**.
- **`brag project edit` location editing (`--add-path`/`--remove-path`)**
  → **SPEC-033**.
- **Any time-windowed / `--since` count, per-type breakdown, or
  time-series** → out of scope (LD1; the brief excludes analytics).
  Revisit only if real dogfooding shows the lifetime count is unhelpful.
- **Splitting `state_note` into `state` + `next_action`** → resolved
  *against* (LD3). If a future renderer needs structure it is an additive
  `0005` migration superseding DEC-017's note sub-choice — not this spec.
- **Archived-project views / a `--all` or `--archived` flag** → not a
  STAGE-007 criterion; the dashboard is non-archived by definition (LD2).
  A future "show archived too" flag would be its own small spec.
- **Showing `locations` or a recent-brag *list*** → the dashboard is a
  count + state-note index; `brag project show` is the detail view.
- **Any migration or schema change** → there is none.

## Notes for the Implementer

### `storage` — the dashboard struct + query (`internal/storage/project.go`)

Add after `ListProjects` (or beside the other read methods). The query is
the §12(b)-validated literal — transcribe it verbatim. `COUNT(e.id)` (not
`COUNT(*)`) is load-bearing: with the `LEFT JOIN`, a project with no
matching entries produces one row whose `e.*` columns are `NULL`, and
`COUNT(e.id)` counts non-null → `0`, where `COUNT(*)` would wrongly count
the null-join row as `1`.

```go
// ProjectStatus is one row of the `brag project status` dashboard: a
// non-archived project plus its brag count — the number of entries whose
// free-text project string equals the project name (the DEC-017
// soft-string-match join, entries.project = projects.name). Locations are
// intentionally omitted: the dashboard is about activity, not where a
// project lives (that is `brag project show`).
type ProjectStatus struct {
	ID        int64
	Name      string
	Status    string
	StateNote string
	BragCount int
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ProjectStatuses returns every non-archived project (status != 'archived')
// ordered updated_at DESC, id DESC, each with its total brag count via the
// DEC-017 soft-string-match join (entries.project = projects.name). A
// project with no matching entries has BragCount 0 (LEFT JOIN + COUNT(e.id)).
// Returns a non-nil empty slice when no non-archived projects exist.
func (s *Store) ProjectStatuses() ([]ProjectStatus, error) {
	ctx := context.Background()
	rows, err := s.db.QueryContext(ctx,
		`SELECT p.id, p.name, p.status, p.state_note, p.created_at, p.updated_at,
		        COUNT(e.id) AS brag_count
		   FROM projects p
		   LEFT JOIN entries e ON e.project = p.name
		  WHERE p.status != 'archived'
		  GROUP BY p.id
		  ORDER BY p.updated_at DESC, p.id DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("project statuses: %w", err)
	}
	defer rows.Close()

	out := make([]ProjectStatus, 0)
	for rows.Next() {
		var (
			st                         ProjectStatus
			createdAtRaw, updatedAtRaw string
		)
		if err := rows.Scan(&st.ID, &st.Name, &st.Status, &st.StateNote,
			&createdAtRaw, &updatedAtRaw, &st.BragCount); err != nil {
			return nil, fmt.Errorf("project statuses: %w", err)
		}
		created, err := time.Parse(time.RFC3339, createdAtRaw)
		if err != nil {
			return nil, fmt.Errorf("project statuses: parse created_at %q: %w", createdAtRaw, err)
		}
		updated, err := time.Parse(time.RFC3339, updatedAtRaw)
		if err != nil {
			return nil, fmt.Errorf("project statuses: parse updated_at %q: %w", updatedAtRaw, err)
		}
		st.CreatedAt = created.UTC()
		st.UpdatedAt = updated.UTC()
		out = append(out, st)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("project statuses: %w", err)
	}
	return out, nil
}
```

(`project.go` already imports `context`, `fmt`, `time` — no new imports.)

### `export` — the JSON helper (`internal/export/project.go`)

Add beside `projectRecord`/`ToProjectsJSON`, mirroring its DEC-011
discipline. `brag_count` is an `int` (JSON number); `state_note` is the
**full** stored value (truncation is plain-output only):

```go
// projectStatusRecord is the DEC-011-family serialization for a `brag
// project status` dashboard row: the project's identity fields plus
// brag_count. state_note is carried in full (the plain-output truncation
// does not apply to JSON). Locations are omitted (see storage.ProjectStatus).
type projectStatusRecord struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	StateNote string `json:"state_note"`
	BragCount int    `json:"brag_count"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// ToProjectStatusesJSON renders dashboard rows as a naked JSON array
// (DEC-011 shape; 2-space indent). Empty/nil input renders "[]", never "null".
func ToProjectStatusesJSON(statuses []storage.ProjectStatus) ([]byte, error) {
	out := make([]projectStatusRecord, 0, len(statuses))
	for _, st := range statuses {
		out = append(out, projectStatusRecord{
			ID:        st.ID,
			Name:      st.Name,
			Status:    st.Status,
			StateNote: st.StateNote,
			BragCount: st.BragCount,
			CreatedAt: st.CreatedAt.UTC().Format(time.RFC3339),
			UpdatedAt: st.UpdatedAt.UTC().Format(time.RFC3339),
		})
	}
	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal project statuses json: %w", err)
	}
	return b, nil
}
```

(`project.go` (export) already imports `encoding/json`, `fmt`, `time`,
and `storage` — no new imports.)

### `cli` — the `status` subcommand (`internal/cli/project.go`)

Register `status` on the parent (in `NewProjectCmd`, after `show` — read
commands `list`/`show`/`status` first, then the mutations) and update the
parent `Short`:

```go
	Short: "Manage projects (new, list, show, status, edit, archive, delete)",
	...
	cmd.AddCommand(newProjectShowCmd())
	cmd.AddCommand(newProjectStatusCmd())
	cmd.AddCommand(newProjectEditCmd())
```

The subcommand (clones `newProjectListCmd`'s `--format` shape):

```go
func newProjectStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show active projects by recency with brag counts",
		Long: `Show every non-archived project, most-recently-updated first, as a scannable
dashboard: one row per project with its status, total brag count, and state
note. The brag count is the number of brag entries whose project matches the
project name (DEC-017 soft string match). Archived projects are not shown.

Output is plain tab-separated rows (default) or a naked JSON array of status
objects (--format json) per DEC-011. In plain output a long state note is
truncated; the JSON carries it in full.

Examples:
  brag project status                 # name<TAB>status<TAB>count<TAB>state note
  brag project status --format json   # naked JSON array of status objects`,
		RunE: runProjectStatus,
	}
	cmd.Flags().String("format", "", "output format (one of: json); default is plain tab-separated")
	return cmd
}

func runProjectStatus(cmd *cobra.Command, _ []string) error {
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

	statuses, err := s.ProjectStatuses()
	if err != nil {
		return fmt.Errorf("project statuses: %w", err)
	}

	out := cmd.OutOrStdout()
	if format == "json" {
		body, err := export.ToProjectStatusesJSON(statuses)
		if err != nil {
			return fmt.Errorf("render project statuses json: %w", err)
		}
		fmt.Fprintln(out, string(body))
		return nil
	}
	for _, st := range statuses {
		fmt.Fprintf(out, "%s\t%s\t%d\t%s\n",
			st.Name, st.Status, st.BragCount, truncateStateNote(st.StateNote))
	}
	return nil
}

// truncateStateNote shortens a state note for the plain dashboard so a
// long note doesn't blow out the row. Rune-based so a multibyte note is
// never split mid-character. JSON output is never truncated.
func truncateStateNote(note string) string {
	const max = 50
	r := []rune(note)
	if len(r) <= max {
		return note
	}
	return string(r[:max]) + "…"
}
```

Imports for `project.go` (CLI) are unchanged — `errors`, `fmt`, `io`,
`strconv`, `strings`, `bufio`, plus `config`/`export`/`storage`/`cobra`
are all already present. **No `database/sql`** (`no-sql-in-cli-layer`).

### `docs/api-contract.md` — the status-change UPDATE

**(a)** Add this section **after** `### brag project show <name|id>`
(ends ~line 501) and **before** `### brag project edit` (~line 503).
House style mirrors the adjacent `brag project list`/`show` sections:

```
### `brag project status` — active-project dashboard (STAGE-007)

​```
brag project status                 # name<TAB>status<TAB>count<TAB>state note
brag project status --format json   # naked JSON array of status objects
​```

Shows every **non-archived** project (status `active`, `paused`, or `done`),
most-recently-updated first (`updated_at DESC, id DESC`), as a scannable
dashboard. Each row carries the project name, status, a **brag count** (the
number of entries whose `project` string equals the project name — the DEC-017
soft string match, counted over all time), and the state note.

- Plain output: tab-separated `<name>\t<status>\t<brag_count>\t<state_note>`
  rows on stdout (a long state note is truncated; an empty note prints empty).
- `--format json` — naked JSON array of status objects (DEC-011; 2-space
  indent; `[]` on empty, never `null`). Object keys: `id, name, status,
  state_note, brag_count, created_at, updated_at` (`brag_count` a number;
  `state_note` carried in full, never truncated; timestamps RFC3339).
- Default (no `--format`) — plain rows. Unknown `--format` exits 1 (user
  error). stdout carries data; stderr empty.
```

(The `​` zero-width marks above only escape the nested fences in this
spec — the real doc uses plain triple-backtick fences.)

**(b)** Fix the now-stale forward-reference in the `show` section
(~line 494): change
`(No recent-brag count — that is `brag project status`, a later STAGE-007 spec.)`
to
`(No recent-brag count — that is `brag project status` (below).)`
— `status` ships in this spec, so "a later spec" is a stale status claim.

(The References list already carries a `DEC-017` row from SPEC-028 — **no
change**; this spec emits no new DEC.)

### `guidance/questions.yaml` — resolve `project-state-note-shape`

Set the `project-state-note-shape` question to `status: answered`, add
`answered_by: SPEC-030 (2026-06-09)`, and append a resolution note (the
single free-text column renders fine on the dashboard; no split; a later
split stays a cheap additive migration). Matches the `tags-storage-model`
resolved-question precedent (the file's documented mechanism: status →
`answered` + link to the resolving spec). **This is a design deliverable
of this spec, done now** (not at build).

### Flag-default WATCH (carry-forward, not codified here)

`status`'s `--format ""` default is stated explicitly (LD4) per the
STAGE-006/SPEC-026 flag-default-explicitness WATCH item. This is the
**third** CLI `--format` spec to do so (SPEC-026, SPEC-028, SPEC-030) —
the N=3 same-outcome bar. **Do not codify mid-stage** (§12 codification
discipline): note it for STAGE-007 close as a candidate to fold a
one-liner into the literal-artifact-as-spec guidance ("state each
embedded flag's default, not just its accepted values").

### Gotchas

- **`COUNT(e.id)`, not `COUNT(*)`.** With the `LEFT JOIN`, `COUNT(*)`
  would count the null-join row as `1` for a zero-brag project.
  `COUNT(e.id)` counts only non-null → `0`. §12(b)-confirmed.
- **`status != 'archived'`** is the filter (not `status = 'active'`).
  `paused` and `done` projects appear; only `archived` is hidden (LD2).
- **No SQL in `project.go` (CLI).** The join lives in
  `storage.ProjectStatuses()`; the CLI calls the method.
- **Truncation is plain-only.** JSON carries the full `state_note`; do
  not truncate in `ToProjectStatusesJSON`.
- **No `main.go` change.** `status` rides the already-registered
  `NewProjectCmd` parent — adding it to the parent's `AddCommand` chain is
  enough. The functional `status` tests (via `newProjectTestRoot`, which
  builds the parent) cover its registration; still confirm
  `./brag project status --help` in the real binary.
- **`gofmt -w .` + `go vet ./...`** before the PR.

---

## Build Completion

*Filled in at the end of the **build** cycle, before advancing to verify.*

- **Branch:**
- **PR (if applicable):**
- **All acceptance criteria met?** yes/no
- **New decisions emitted:**
  - none expected (LD1–LD5 are localized; confidence ≥ 0.85; no DEC-019)
- **Deviations from spec:**
  - [list]
- **Follow-up work identified:**
  - [any new specs for the stage's backlog]

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
