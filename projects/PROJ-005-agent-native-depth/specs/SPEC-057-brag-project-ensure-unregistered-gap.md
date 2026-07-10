---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-057
  type: story                      # epic | story | task | bug | chore
  cycle: verify
  blocked: false
  priority: medium
  complexity: S                    # S | M | L  (L means split it)

project:
  id: PROJ-005
  stage: STAGE-015
repo:
  id: bragfile

agents:
  architect: claude-opus-4-8
  implementer: claude-opus-4-8     # usually same Claude, different session
  created_at: 2026-07-10

references:
  decisions: [DEC-036, DEC-017, DEC-018]
  constraints:
    - no-sql-in-cli-layer
    - stdout-is-for-data-stderr-is-for-humans
    - errors-wrap-with-context
    - timestamps-in-utc-rfc3339
    - test-before-implementation
    - storage-tests-use-tempdir
  related_specs: [SPEC-027, SPEC-028, SPEC-029, SPEC-031, SPEC-055, SPEC-058]
---

# SPEC-057: `brag project ensure` — close the unregistered-project gap

## Context

Entries store `project` as a **free-text string** with no referential check
against the `projects` table (DEC-017's soft string match). The MCP `brag_add`
does **not** auto-fill `project` from cwd the way the CLI `brag add` does
(SPEC-032), so an agent logging over MCP can write `project: "standup"` when
`standup` was never registered. The result is an **orphan project string**:
`brag project list` / `brag project status` (which read the `projects` table)
never see it, and downstream consumers that map entries→repos (the `standup`
portfolio tracker consumes `brag list --format json` + `brag project list
--format json`) silently miss it.

STAGE-015 makes `brag mcp serve` a first-class path for agents. This spec closes
the referential half of that gap by giving agents (and humans) an **explicit,
idempotent primitive** to register a project by name before or after logging
against it — safe to run repeatedly, safe to script into an agent's setup — and
documents the two soft-link facts downstream mappers depend on. It is the
second spec of STAGE-015, sitting between SPEC-055 (`brag mcp install`) and
SPEC-058 (the agent-facing MCP docs).

This spec deliberately does **not** change `brag_add` / `brag add` to
auto-register unknown projects (see Locked design decisions #4 and Out of
scope) — silent auto-registration would pollute the `projects` table with every
typo. `ensure` is the safe, explicit alternative.

## Goal

Add `brag project ensure <name> [--location PATH]`: an idempotent create-or-no-op
that registers a project by name (status `active`) if absent and no-ops if
present, optionally attaching a filesystem location idempotently — always safe to
re-run — and document the two soft-link facts (project-list locations are
authoritative-but-incomplete; a project may have multiple locations) where the
project commands are documented.

## Inputs

- **Files to read:**
  - `internal/storage/project.go` — `CreateProject` (returns `ErrProjectExists`
    on dup name), `GetProjectByName`, `AddLocation` (globally-unique paths,
    `ErrLocationExists`), `EditLocations` (the owner-query pattern
    `EnsureLocation` mirrors), `ProjectStatuses` (the `LEFT JOIN entries e ON
    e.project = p.name` soft-link).
  - `internal/storage/errors.go` — `ErrProjectExists`, `ErrLocationExists`,
    `ErrLocationOtherProject` (reused, not re-defined).
  - `internal/cli/project.go` — `NewProjectCmd`, `runProjectNew` (the
    create-then-attach + ListProjects pre-check pattern; the STDERR confirmation
    convention to mirror).
  - `internal/cli/errors.go` — `ErrUser` / `UserErrorf`.
- **Related code paths:** `internal/storage/`, `internal/cli/`.
- **No external APIs.**

## Outputs

- **Files modified:**
  - `internal/storage/project.go` — add two methods:
    - `func (s *Store) EnsureProject(name string) (Project, bool, error)`
    - `func (s *Store) EnsureLocation(projectID int64, path string) (bool, error)`
  - `internal/cli/project.go` — add `newProjectEnsureCmd()` + `runProjectEnsure`,
    register it in `NewProjectCmd` (add `cmd.AddCommand(newProjectEnsureCmd())`),
    and add `ensure` to the parent command's `Short` string.
  - `docs/api-contract.md` — add a `### brag project ensure <name> [--location
    PATH]` section (placed after `brag project new`), and add the **two soft-link
    facts** as a short note in the project-commands area (see Notes for the
    Implementer for the literal text).
  - `docs/tutorial.md` — one short line pointing at `brag project ensure` for
    idempotent registration, only if it fits the existing project-registry
    paragraph (§ around line 234 / 672); do not force it.
- **Files created:** none (tests are appended to existing files).
- **New exports:** `Store.EnsureProject`, `Store.EnsureLocation` (signatures
  above).
- **Database changes:** none. No migration. `ensure` reuses the `projects` /
  `project_locations` tables from `0004_add_projects.sql`.

### Premise-audit results (RUN at design, §9)

Greps run against the repo at design time; **no breakages found**, so there are
no planned test deletions or count-bumps:

- `grep -n 'new", "list"\|Commands()' internal/cli/project_test.go` →
  `project_test.go:51` — `TestProjectCmd_BarePrintsHelp` asserts a **subset**
  `{"new","list","show","status","here"}` via `strings.Contains`, not an exact
  set and not the absence of others. Adding `ensure` leaves it green; no update
  needed. (Cobra auto-lists `ensure` in help, which the subset check ignores.)
- `grep -rn "Manage projects" internal/cli/*_test.go docs/` → no hits. No test
  or doc asserts the parent command's exact `Short` string, so adding `ensure`
  to it breaks nothing.
- `grep -rn "ensure" internal/ docs/` → no pre-existing `brag project ensure`
  references to reconcile (only unrelated prose uses of the word "ensure").
- No literal-count assertion (à la `schema_migrations` / `ListFilter` field
  count) is coupled to the project subcommand set.

## Acceptance Criteria

- [ ] `Store.EnsureProject(name)` creates an `active` project when absent and
      returns `(project, true, nil)`.
- [ ] `Store.EnsureProject(name)` on an existing name returns the existing row,
      `(project, false, nil)` — **no error**, no `ErrProjectExists`, no
      duplicate row, `updated_at` unchanged.
- [ ] `Store.EnsureLocation(projectID, path)` attaches a free path and returns
      `(true, nil)`; a path already attached **to this project** is an
      idempotent no-op returning `(false, nil)` with no duplicate row.
- [ ] `Store.EnsureLocation(projectID, path)` on a path attached to a
      **different** project returns `(false, ErrLocationOtherProject)`.
- [ ] `brag project ensure <name>` creates the project (visible in `brag project
      list`) and prints `Created project "<name>".` to **stderr**; stdout empty;
      exit 0.
- [ ] Re-running `brag project ensure <name>` exits 0 with no error, prints
      `Project "<name>" already exists.` to stderr, and creates no duplicate.
- [ ] `brag project ensure <name> --location PATH` attaches PATH and prints
      `Attached location "<path>".`; re-running prints `Location "<path>"
      already attached.` and adds no duplicate; both exit 0.
- [ ] `brag project ensure <name> --location PATH` where PATH belongs to a
      different project returns a `UserError` naming the path; exit 1.
- [ ] Empty name → `UserError`; name > 64 chars → `UserError` mentioning `64`.
- [ ] `--location` defaults to `""` (unset); no location is attached and no cwd
      is registered when the flag is omitted.
- [ ] `docs/api-contract.md` documents `brag project ensure` and the two
      soft-link facts.

## Failing Tests

Written during **design**, made to pass in **build**. These are appended to the
existing test files and will not compile until `EnsureProject` / `EnsureLocation`
and the `ensure` subcommand exist — the intended fail-first state.

- **`internal/storage/project_test.go`** (use `newTestStore(t)` / `t.TempDir()`):
  - `TestEnsureProject_CreatesWhenAbsent` — asserts `created == true`, `Status ==
    "active"`, non-zero `CreatedAt`, and the row is persisted (`GetProjectByName`).
    *(Pairs LD1.)*
  - `TestEnsureProject_NoOpWhenExists` — second call returns the same `ID`,
    `created == false`, `err == nil`, and `ListProjects` length stays 1 (no
    duplicate). *(Pairs LD1 — the never-errors-on-dup guarantee.)*
  - `TestEnsureLocation_AttachesIdempotently` — first call `attached == true`;
    second call of the same path to the same project `attached == false`,
    `err == nil`, exactly one location row. *(Pairs LD2 — idempotency.)*
  - `TestEnsureLocation_DifferentProjectErrLocationOtherProject` — a path owned
    by another project returns `ErrLocationOtherProject`, `attached == false`.
    *(Pairs LD2 — cross-project conflict.)*
- **`internal/cli/project_test.go`** (separate `outBuf`/`errBuf`; assert no
  cross-leakage):
  - `TestProjectEnsure_CreatesWhenAbsent` — stdout empty; stderr contains
    `Created project "standup".`; the project appears in `project list`.
    *(Pairs LD3.)*
  - `TestProjectEnsure_IdempotentReRunNoError` — second run exits 0, stderr
    contains `Project "standup" already exists.`, exactly one `standup` row.
    *(Pairs LD3 + LD1.)*
  - `TestProjectEnsure_AttachesLocation` — stderr contains `Attached location
    "/repo/standup".`; `project show` lists the path. *(Pairs LD2/LD3.)*
  - `TestProjectEnsure_LocationIdempotentReRun` — second run stderr contains
    `Location "/repo/standup" already attached.`, no duplicate row. *(Pairs LD2.)*
  - `TestProjectEnsure_LocationOnDifferentProjectErrUser` — cross-project path →
    `ErrUser`, message names the path. *(Pairs LD2.)*
  - `TestProjectEnsure_EmptyNameErrUser` — empty name → `ErrUser`. *(Pairs LD3.)*
  - `TestProjectEnsure_OverLongNameErrUser` — 65-char name → `ErrUser` mentioning
    `64`. *(Pairs LD3.)*
  - `TestProjectEnsure_StdoutStderrSeparation` — `outBuf.Len() == 0` and
    `errBuf.Len() != 0` on a create. *(Pairs the stdout/stderr constraint.)*
  - `TestProjectEnsure_HelpShowsExamples` — `ensure --help` contains
    `Examples:`, `brag project ensure`, and `--location`. *(Pairs LD3 literal
    artifact.)*

## Implementation Context

*Read this section (and the files it points to) before starting the build cycle.*

### Locked design decisions

**LD1 — `EnsureProject(name string) (Project, bool, error)` is the idempotent
upsert.** The bool is `created` (true = a row was inserted, false = the name
already existed). Implementation (explicit-over-clever, house style):

1. `GetProjectByName(name)`. On success → return `(existing, false, nil)`.
2. On `ErrNotFound` → `CreateProject(Project{Name: name})` (status defaults to
   `active`; `created_at`/`updated_at` set UTC RFC3339 by `CreateProject`). On
   success → return `(created, true, nil)`.
3. Defensive: if `CreateProject` returns `ErrProjectExists` (a TOCTOU that
   cannot really occur in the single-user CLI, but keep the method total),
   re-`GetProjectByName` and return `(existing, false, nil)`.

`EnsureProject` **never** returns `ErrProjectExists`. It does **not** bump
`updated_at` on the no-op path (it changed nothing — it returns the existing row
verbatim via `GetProjectByName`). Wrap all internal errors with
`fmt.Errorf("ensure project %q: %w", name, err)`.

*Rejected alternative (build-time):* insert-first-then-fetch-on-dup as the
primary path. Rejected — the read-first shape returns the fully hydrated
existing project (locations included) on the common re-run path, reads more
clearly, and reserves the insert path for the genuinely-absent case; the dup
branch is kept only as a defensive backstop, not the primary flow.

**LD2 — `EnsureLocation(projectID int64, path string) (bool, error)` is the
idempotent location attach; cross-project is `ErrLocationOtherProject`.** The
bool is `attached` (true = a row was inserted, false = already attached to this
project). A project may hold **multiple** locations, so this only guards the
*same* path, not the count. Implementation, mirroring `EditLocations`'
owner-query:

1. Query the current owner of `path`: `SELECT project_id FROM project_locations
   WHERE path = ?`.
2. `sql.ErrNoRows` (path free) → `INSERT INTO project_locations (project_id,
   path) VALUES (?, ?)`; return `(true, nil)`.
3. owner `== projectID` → idempotent no-op; return `(false, nil)`.
4. owner `!= projectID` → return `(false, ErrLocationOtherProject)` (reuse the
   existing sentinel; paths are globally UNIQUE — the guarantee SPEC-031 relies
   on). Do **not** silently retarget the path.

Wrap internal errors with `fmt.Errorf("ensure location %q: %w", path, err)`.
This keeps all SQL in `internal/storage` (`no-sql-in-cli-layer`) and puts the
same-vs-different-project decision behind a typed error the CLI maps to a
`UserError`.

*Rejected alternative (build-time):* have the CLI iterate `ListProjects` to
detect the owner (the `runProjectNew` pre-check pattern). Rejected here — a
storage method with a typed error is cleaner to test at the storage layer
(the spec's storage-level location tests), avoids re-hydrating every project in
the CLI, and reuses `EditLocations`' existing owner-query shape. `AddLocation`
stays unchanged (it cannot distinguish same- from other-project — it returns
`ErrLocationExists` for both — which is why `EnsureLocation` exists).

**LD3 — CLI `brag project ensure <name> [--location PATH]`.**
- Positional `name`: required, `strings.TrimSpace`d; empty → `UserErrorf(...)`.
  Length limit **≤ 64 characters** (rune count) → over-length is
  `UserErrorf("project name exceeds 64-character limit")` (matches the 64-char
  cap `brag add --json` / MCP enforce on the `project` field, so an ensured name
  can always soft-match a normally-added entry; `runProjectNew`'s current
  lack of a length check is a pre-existing asymmetry left out of scope here).
- `--location` flag: `String`, **default `""` (unset)**. A location is attached
  **only** when the flag is non-empty. It is deliberately NOT defaulted to cwd —
  ensure must never register a surprise cwd (that would re-introduce the
  accidental-registration class of bug). State the default explicitly
  (§12 flag-default-explicitness).
- Sequence: resolve the db path (`config.ResolveDBPath` + `storage.Open`, mirror
  the other subcommands) → `EnsureProject(name)` → if `--location` non-empty,
  `EnsureLocation(proj.ID, path)`.
- **Output discipline (mirror `runProjectNew`): stdout stays empty; the human
  confirmation goes to `cmd.ErrOrStderr()`. Exit 0 on both create and no-op.**
  Two independent, composable stderr lines (each testable in isolation):
  - Project line — `Created project %q.` when `created`, else `Project %q
    already exists.`
  - Location line (**only when `--location` given**) — `Attached location %q.`
    when `attached`, else `Location %q already attached.`
  On `EnsureLocation` returning `ErrLocationOtherProject`, return
  `UserErrorf("path %q is already registered to another project", path)`. Note
  the project is still (idempotently) ensured before the location step; a
  cross-project location conflict is a genuine user error and the re-run after
  fixing it is safe — this is consistent with ensure's idempotent contract.

**LD4 — scope boundaries (locked).**
- `brag add` / `brag_add` stay **free-text**; this spec does **not** add
  silent auto-registration of unknown projects on add. *Rejected alternative:*
  auto-register (or warn-and-register) unknown `project` values at add time —
  rejected because it pollutes the `projects` table with every typo and
  mis-spelling, exactly the noise `ensure` avoids by being explicit. `ensure` is
  the safe primitive; add stays a pure capture.
- Whether the **MCP server** exposes an `ensure`-equivalent tool is **out of
  scope** (candidate follow-up — see Out of scope). This spec adds the CLI
  command and the storage primitives only.

### Decisions that apply

- `DEC-036` — this spec's decision: the `project ensure` idempotent-upsert
  semantics (`EnsureProject`/`EnsureLocation` signatures + return shapes), the
  keep-`brag_add`-free-text boundary, and the two documented soft-link facts.
- `DEC-017` — `entries.project` ↔ `projects` soft string match; the reason an
  unregistered project string is invisible to `project list`/`status` and why
  registration is purely additive (no backfill, no entry rewrite).
- `DEC-018` — project delete blast radius; house context for the project model
  (ensure is the additive counterpart to delete — neither touches entries).

### Constraints that apply

- `no-sql-in-cli-layer` — the ensure/attach SQL lives in `EnsureProject` /
  `EnsureLocation` in `internal/storage`; the CLI only calls them.
- `stdout-is-for-data-stderr-is-for-humans` — confirmations go to stderr;
  stdout stays empty (tests assert no cross-leakage).
- `errors-wrap-with-context` — wrap storage/CLI errors with an operation prefix.
- `timestamps-in-utc-rfc3339` — `EnsureProject`'s create path writes
  `created_at`/`updated_at` via `CreateProject`, which already formats UTC
  RFC3339; no new timestamp code.
- `test-before-implementation`, `storage-tests-use-tempdir` — the failing tests
  above.

### Prior related work

- `SPEC-027` (shipped) — the `0004_add_projects.sql` migration + `CreateProject`
  / `AddLocation` primitives this spec composes.
- `SPEC-028`/`SPEC-029` (shipped) — the `project` CRUD CLI whose `runProjectNew`
  output convention and ListProjects pre-check pattern this mirrors.
- `SPEC-031` (shipped) — `brag project here`; established that
  `project_locations.path` is globally UNIQUE (the guarantee `EnsureLocation`'s
  cross-project error protects).
- `SPEC-055` (STAGE-015 sibling) — `brag mcp install`.
- `SPEC-058` (STAGE-015 sibling) — the agent-facing MCP docs spec that carries
  the full tool contract; this spec documents only the command it adds + the two
  soft-link facts.

### Out of scope (for this spec specifically)

- Auto-registering unknown projects on `brag add` / `brag_add` (rejected, LD4).
- Exposing an `ensure` tool over the MCP server — candidate follow-up; if
  wanted, a new spec adds a `brag_project_ensure` tool to `internal/mcpserver`.
- The big agent-facing MCP docs (full tool schemas, gotchas, impact-framing
  convention) — that is SPEC-058.
- Retrofitting a 64-char name check onto `brag project new` (pre-existing
  asymmetry; if desired, a separate tidy spec).
- Any path normalization — paths are stored verbatim (SPEC-031/DEC-019 own
  normalization at cwd-resolve time only). `EnsureLocation` matches verbatim,
  exactly like `AddLocation`.

## Notes for the Implementer

### Literal artifact — the cobra command (transcribe verbatim)

```go
func newProjectEnsureCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ensure <name> [--location <dir>]",
		Short: "Idempotently register a project (create if absent, no-op if present)",
		Long: `Ensure a project is registered, safely and repeatably. If a project with
this name already exists, ensure does nothing (exit 0); if not, it creates it
with status "active". With --location, the given path is attached if it is not
already registered (a project may have multiple locations); re-attaching the
same path is a no-op. A path already registered to a DIFFERENT project is
rejected (paths are globally unique).

Unlike 'brag project new', ensure never errors when the project already exists,
so it is safe to run before every capture or from an agent's setup script. The
project name is stored verbatim and must be 64 characters or fewer.

--location is optional and defaults to unset: no location (and no current
directory) is registered unless you pass it.

Examples:
  brag project ensure standup
  brag project ensure standup --location ~/code/standup
  brag project ensure platform --location /srv/platform`,
		RunE: runProjectEnsure,
	}
	cmd.Flags().String("location", "", "filesystem directory to attach to the project (optional; default: unset)")
	return cmd
}
```

Register it in `NewProjectCmd` with `cmd.AddCommand(newProjectEnsureCmd())` and
update the parent `Short` to
`"Manage projects (new, ensure, list, show, status, edit, archive, delete, here)"`.

### Exact stderr strings (literal — the tests diff against these)

- `Created project %q.` / `Project %q already exists.`
- `Attached location %q.` / `Location %q already attached.`
- cross-project location: `UserErrorf("path %q is already registered to another project", path)`
- empty name: `UserErrorf("project name must not be empty")`
- over-length: `UserErrorf("project name exceeds 64-character limit")`

### Docs — the two soft-link facts (literal text for `docs/api-contract.md`)

Add near the project commands (after the `brag project ensure` section is a good
home) a short note, e.g.:

```
> **Two soft-link facts for consumers that map entries → projects/repos.**
> (1) `brag project list` (and `brag project status`) read the `projects`
> table, which is **authoritative but incomplete**: an entry's `project` is
> free text (DEC-017) and MCP `brag_add` does not auto-fill it from cwd, so an
> entry may reference a project name that was never registered. Use
> `brag project ensure <name>` to register such a name. (2) A single project
> may have **multiple** on-disk `locations`, so any mapping from an entry (or a
> path) to a repo must handle a one-project-many-directories shape — see the
> `locations` array in `brag project list --format json` / `brag project show
> --format json`.
```

Add the `### brag project ensure <name> [--location PATH]` section itself in the
same style as the existing `### brag project new` section (usage examples,
behavior, exit codes). If a one-line pointer fits the `docs/tutorial.md`
project-registry paragraph, add it; otherwise leave tutorial untouched (note the
decision in Build Completion).

### Gotchas / reuse

- Reuse `ErrLocationOtherProject`; do **not** add a new sentinel.
- Fail-first (§9/§12 build step): after appending nothing new (tests already
  written), the first `go test ./...` in build must fail to **compile** on the
  missing `EnsureProject`/`EnsureLocation`/`ensure` symbols — that is the
  expected fail-first, not a stray error. Then implement until green.
- Do not bump `updated_at` on the ensure no-op path.
- Mirror `runProjectNew`'s db-open/`defer s.Close()` boilerplate exactly.

---

## Build Completion

*Filled in at the end of the **build** cycle, before advancing to verify.*

- **Branch:** feat/spec-057-project-ensure
- **PR (if applicable):** (opened at end of build — see PR link in the ship notes)
- **All acceptance criteria met?** yes
- **New decisions emitted:**
  - none (DEC-036 was emitted at design; no new decision needed at build)
- **Deviations from spec:**
  - none. `EnsureProject`/`EnsureLocation` implemented per LD1/LD2 verbatim;
    the cobra command and stderr strings transcribed from the Notes section as
    written; the two soft-link facts and the `brag project ensure` doc section
    added to `docs/api-contract.md`; a one-line pointer added to
    `docs/tutorial.md` (it fit the project-registry paragraph cleanly).
- **Follow-up work identified:**
  - none new. The two candidate follow-ups already named in the spec still
    stand (an MCP `brag_project_ensure` tool; retrofitting a 64-char name check
    onto `brag project new`), both explicitly out of scope here.

### Build-phase reflection (3 questions, short answers)

Process-focused: how did the build go? What friction did the spec create?

1. **What was unclear in the spec that slowed you down?**
   — Nothing. The spec was unusually complete: exact signatures, the verbatim
   cobra command, exact stderr strings, and the literal doc text were all
   provided, and the failing tests were already authored. Build was a
   transcription-plus-wire-up exercise, which is the point.

2. **Was there a constraint or decision that should have been listed but wasn't?**
   — No. Every constraint that bit (no-sql-in-cli-layer, the stdout/stderr
   split, errors-wrap-with-context) was listed, and DEC-036's LD1–LD4 covered
   every design question I would otherwise have had to raise.

3. **If you did this task again, what would you do differently?**
   — Nothing on the implementation. The only friction was environmental (the
   isolated worktree vs. shared-checkout path for edits), not spec-related.

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
