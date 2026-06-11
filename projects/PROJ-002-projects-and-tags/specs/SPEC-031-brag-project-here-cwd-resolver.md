---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-031
  type: story                      # epic | story | task | bug | chore
  cycle: build                     # frame | design | build | verify | ship
  blocked: false
  priority: medium
  complexity: S                    # S | M | L  (L means split it)

project:
  id: PROJ-002
  stage: STAGE-007
repo:
  id: bragfile

agents:
  architect: claude-sonnet-4-6
  implementer: claude-sonnet-4-6   # usually same Claude, different session
  created_at: 2026-06-10

references:
  decisions: [DEC-017, DEC-019]
  constraints:
    - no-sql-in-cli-layer
    - stdout-is-for-data-stderr-is-for-humans
    - storage-tests-use-tempdir
    - errors-wrap-with-context
    - test-before-implementation
    - no-new-top-level-deps-without-decision
  related_specs: [SPEC-027, SPEC-028, SPEC-029, SPEC-030, SPEC-032]
---

# SPEC-031: brag project here — cwd resolver

## Context

This is the **sixth spec of STAGE-007** (Projects core). It adds
`brag project here`: run it from inside a registered project directory
(or any subdirectory of one) and the command prints which project you're
in. The command is S-complexity because its entire new behavior is one
Store method (`ProjectForPath`) plus one thin CLI wrapper; there is no
schema change and no migration.

Two things make this spec load-bearing beyond its size:
1. **SPEC-032 reuses the resolver.** `brag add --project` auto-fill
   (the next spec) calls `ProjectForPath` directly from its `RunE`.
   The Store method must be clean, tested, and well-specified before
   SPEC-032 can build on it.
2. **The resolution policy is a non-trivial design decision** (how does
   a cwd map to a project when the user may be deep in the tree?),
   locked as **DEC-019** and applicable to both SPEC-031 and SPEC-032.

Parent stage: `STAGE-007-projects-core.md`; PROJ-002 brief governs
dev/prod DB isolation (see Implementation Context). Prior foundation:
`SPEC-027` (the `project_locations` schema + `AddLocation`),
`SPEC-028` (CLI conventions: `renderProjectPlain`, `resolveProjectByNameOrID`,
`--format ""` default), and `SPEC-030` (the `brag project status`
dashboard — confirms SPEC-031 is the next pending spec).

## Goal

Add `brag project here`, which reads `os.Getwd()` and prints the
matching registered project (nearest-ancestor match, DEC-019), and
expose the resolution logic as `Store.ProjectForPath(cwd string)
(*Project, error)` so SPEC-032 can reuse it without duplicating the
algorithm.

## Inputs

- **Files to read:**
  - `internal/storage/project.go` — `Project` struct, existing Store
    methods (`CreateProject`, `GetProject`, `ListProjects`, `AddLocation`),
    `locationsForProject`; `ProjectForPath` is added here.
  - `internal/storage/errors.go` — confirm no new sentinel needed
    (`ProjectForPath` uses nil return, not `ErrNotFound`).
  - `internal/cli/project.go` — `NewProjectCmd()` (add `here`),
    `renderProjectPlain`, `resolveProjectByNameOrID`, `--format` flag
    default pattern, `getCwd` variable (add it); `runProjectHere` is
    added here.
  - `internal/export/project.go` — `ToProjectJSON` (reused for
    `--format json` output).
  - `decisions/DEC-017` — global-uniqueness of `project_locations.path`
    (the guarantee that makes the longest-prefix tie-break well-defined);
    soft string match (not relevant to `here`, but confirms `here` does
    NOT join on `entries.project`).
  - `decisions/DEC-019` — the resolution policy (emitted by this spec;
    read before build).
  - `internal/cli/project_test.go` — test patterns
    (`newProjectTestRoot`, `runProjectCmd`, `TestProjectCmd_BarePrintsHelp`);
    the `getCwd` override approach used here mirrors the package-level
    variable pattern used in the same package.
- **Related code paths:** `internal/storage/` (new method), `internal/cli/`
  (new subcommand + `getCwd` var), `docs/api-contract.md` (status-change
  update).

## Outputs

- **Files modified:**
  - `internal/storage/project.go` — add `ProjectForPath(cwd string)
    (*Project, error)` Store method. No new imports needed (already
    imports `context`, `database/sql` indirectly via the package,
    `fmt`, `strings`, `time`; adds `path/filepath`).
  - `internal/cli/project.go` — add `var getCwd = os.Getwd` (package
    level; overridable in tests), `newProjectHereCmd()`,
    `runProjectHere()`, and `cmd.AddCommand(newProjectHereCmd())` in
    `NewProjectCmd()`. Add `"os"` + `"path/filepath"` to stdlib imports.
  - `docs/api-contract.md` — **status-change UPDATE**: add
    `### brag project here` section (see Notes for the literal).
- **Files modified (tests — additive):**
  - `internal/storage/project_test.go` — add `ProjectForPath` tests
    (seven new tests; no rewrites of SPEC-027/028/029/030 tests).
  - `internal/cli/project_test.go` — add `here` CLI tests using `getCwd`
    override (six new tests; one additive update to
    `TestProjectCmd_BarePrintsHelp`).
- **New exports:**
  - `func (s *Store) ProjectForPath(cwd string) (*Project, error)`
- **Database changes:** **NONE.** No migration. The schema is already
  correct (the `project_locations` table + its global `UNIQUE(path)`
  guarantee from SPEC-027 are exactly what the resolver needs).

### Premise audit (`projects/_templates/premise-audit.md`), run at design

Greps **run at design 2026-06-10** and reconciled below.

**1. Inversion / removal — NONE.** `brag project here` is a brand-new
subcommand. No existing Store method, flag, column, or CLI behavior is
changed. The `NewProjectCmd()` parent gains one `AddCommand` call;
existing subcommands are unaffected. Grep run:
`grep -rn 'ProjectForPath\|project.*here' internal/` → hits are only
comments referencing SPEC-031 in `project.go` and
`migrations/0004_add_projects.sql` (both forward-looking, not
behavior). Zero planned rewrites.

**2. Addition / count-bump.** Two sites:

- **`schema_migrations` count — NONE.** No migration added; the count
  stays at 4. Grep run:
  `grep -rn '0004_add_projects\|want 4\|count != 4' internal/` →
  all hits are SPEC-027 sites; untouched by SPEC-031.
- **`TestProjectCmd_BarePrintsHelp` subcommand list — ADDITIVE UPDATE.**
  The test at `internal/cli/project_test.go:51` iterates
  `[]string{"new", "list", "show", "status"}` and asserts each is
  present in `brag project` help. Adding `here` to `NewProjectCmd()`
  does **not** break this check (it uses `strings.Contains`, not an
  exact list). But per the additive-case premise-audit rule, the
  registration of `here` should have a paired failing test — the
  cleanest place is to add `"here"` to this slice. **Planned additive
  update:** extend the slice to
  `[]string{"new", "list", "show", "status", "here"}`.

**3. Status change — `docs/api-contract.md` only.** Grep run:
`grep -rn "brag project here\|SPEC-031" docs/ README.md` → two hits:
- `docs/api-contract.md:458` — `"path normalization is \`brag project
  here\`'s concern, STAGE-007"` — a forward-reference in the
  `brag project new` section; **stays** (correct after we ship;
  no status claim invalidated).
- `docs/data-model.md:84` — `"SPEC-031 owns normalization"` in the
  `project_locations.path` column description; **stays** (already
  correct).
- `docs/api-contract.md` — **UPDATE**: add a
  `### brag project here` section (see Notes; the command did not
  exist before this spec). The comprehensive tutorial/architecture
  sweep is STAGE-008.

**4. Cross-check.** All grep hits reconciled against the lists above;
no un-enumerated hit remained.

## Acceptance Criteria

Testable outcomes. Happy path, error cases, edge cases.

- [ ] **`ProjectForPath`: exact match.** `ProjectForPath(path)` where
  `path` is a registered location returns a non-nil `*Project` with the
  correct `Name`.
- [ ] **`ProjectForPath`: subdirectory match (LD1).** A cwd that is a
  subdirectory of a registered path returns the owning project (not nil).
- [ ] **`ProjectForPath`: no match.** A cwd with no registered ancestor
  returns `nil, nil` — specifically `nil` not an error wrapping
  `ErrNotFound`.
- [ ] **`ProjectForPath`: longest-prefix wins (LD1/DEC-019).** When
  project A registers `/a/b` and project B registers `/a/b/sub`, a cwd
  of `/a/b/sub/deep` returns project B (more specific).
- [ ] **`ProjectForPath`: separator guard (LD1).** A cwd of
  `/home/user/worker` does NOT match a registered path of
  `/home/user/work` (the guard prevents partial-dirname false-positives).
- [ ] **`ProjectForPath`: normalizes with `filepath.Clean` (LD1).**
  A cwd with redundant `.` segments that cleans to a registered path
  is still matched.
- [ ] **`brag project here`: matched → stdout one-liner, exit 0.**
  With a project registered and cwd inside its location, `brag project
  here` prints `<name>\t<status>\t<state_note>` to stdout (empty
  state_note → `-`), exit 0, stderr empty (LD2, §9).
- [ ] **`brag project here`: no match → stderr + exit 1.**
  With no registered project matching cwd, stderr contains
  `"not inside any registered project"`, exit 1, stdout empty (LD2, §9).
- [ ] **`brag project here --format json`: matched → single JSON object.**
  `--format json` emits a single JSON object (same shape as
  `brag project show --format json`) to stdout, with `locations` hydrated
  (not `[]`), exit 0, stderr empty (LD3).
- [ ] **`brag project here --format <unknown>` → user error.**
  Any non-empty, non-`json` value exits 1 (user error); `errors.Is(err,
  ErrUser)` (LD4).
- [ ] **`--format` default is `""` (plain).** With no flag, output is
  the plain one-liner (LD4).
- [ ] **No SQL under `internal/cli/`.** `runProjectHere` calls Store
  methods only (`ProjectForPath`, `GetProject` for JSON); no
  `database/sql` import in `project.go`.
- [ ] **No migration, no schema change.**

## Failing Tests

Written at **design**; build makes them pass. All are **new/additive** —
the premise audit found zero inversions and no count-bump rewrites
beyond the `TestProjectCmd_BarePrintsHelp` additive update. Storage tests
use `t.TempDir()` (`storage-tests-use-tempdir`). CLI tests override
`getCwd` for cwd injection (see Implementation Context).

### `internal/storage/project_test.go` (modify — additive)

- `"TestProjectForPath_ExactMatch"` — create project `"bragfile"`,
  `AddLocation(id, "/home/user/work")`; call
  `s.ProjectForPath("/home/user/work")` → non-nil result with
  `Name == "bragfile"`, err == nil. Confirms exact-match (the
  degenerate ancestor case).

- `"TestProjectForPath_SubdirectoryMatch"` — same setup; call
  `s.ProjectForPath("/home/user/work/internal/cli")` → non-nil
  `Name == "bragfile"`, err == nil. Confirms LD1 nearest-ancestor.

- `"TestProjectForPath_NoMatch"` — no locations registered; call
  `s.ProjectForPath("/some/random/path")` → result is nil (not
  `ErrNotFound`), err == nil. Confirms nil-on-no-match contract (LD5).

- `"TestProjectForPath_LongestPrefixWins"` — create project A, add
  location `/a/b`; create project B, add location `/a/b/sub`; call
  `s.ProjectForPath("/a/b/sub/deep")` → `Name == "B"` (project B
  registered at `/a/b/sub`, length 7, beats `/a/b`, length 4). Locks
  DEC-019 longest-prefix behavior.

- `"TestProjectForPath_SeparatorGuard"` — create project `"work"`,
  add location `/home/user/work`; call
  `s.ProjectForPath("/home/user/worker")` → nil (the cwd starts with
  `/home/user/work` as a byte-string but NOT as a path ancestor because
  `/home/user/work/` is not a prefix of `/home/user/worker`). Locks
  the separator guard (DEC-019).

- `"TestProjectForPath_NormalizesCleanPath"` — create project `"bragfile"`,
  add location `/home/user/work`; call
  `s.ProjectForPath("/home/user/work/./src")` → non-nil
  `Name == "bragfile"` (filepath.Clean normalizes the cwd to
  `/home/user/work/src` which has `/home/user/work` as ancestor).

- `"TestProjectForPath_MultipleProjectsOneMatch"` — create projects A
  (`/alpha/repo`) and B (`/beta/repo`); call
  `s.ProjectForPath("/alpha/repo/cmd")` → `Name == "A"` (only A
  matches; B's path is `/beta/repo`, a non-ancestor). Confirms no
  cross-contamination between unrelated projects.

> **§12(b) note for build:** these expected values are design-verified
> against the algorithm in DEC-019. Transcribe them; do not re-derive.
> Key literals: `/a/b/sub` (len=7) > `/a/b` (len=4); `filepath.Clean`
> of `/home/user/work/./src` is `/home/user/work/src`; and
> `strings.HasPrefix("/home/user/worker", "/home/user/work/")` is
> false because `"worker"` does not start with `"work/"`.

### `internal/cli/project_test.go` (modify — additive + one inlined update)

All `here` CLI tests override `getCwd` for hermetic control over the cwd
(see Implementation Context § `getCwd` injection). The test structure
mirrors the existing `runProjectCmd` pattern.

- **Additive update to `TestProjectCmd_BarePrintsHelp` (`:51`)** —
  extend the subcommand slice from
  `[]string{"new", "list", "show", "status"}` to
  `[]string{"new", "list", "show", "status", "here"}`. One element
  added; all existing assertions still hold. This is a planned additive
  update (premise audit §2), not an inversion.

- `"TestProjectHere_MatchedProject"` — create project `"bragfile"` with
  state note `"next: cut v0.2.0"` at `t.TempDir()` dir `D`;
  override `getCwd → D`; run `["here"]`; assert:
  - `outBuf` contains `"bragfile\tactive\tnext: cut v0.2.0"` (the
    plain one-liner, LD2)
  - `errBuf == ""` (§9 no cross-leakage)
  - err == nil (exit 0)

- `"TestProjectHere_EmptyStateNote"` — project with empty state_note at
  `D`; override `getCwd → D`; `["here"]`; assert `outBuf` contains
  `"bragfile\tactive\t-"` (empty note renders `-`, LD2).

- `"TestProjectHere_NoMatch"` — no project registered; override
  `getCwd → t.TempDir()`; `["here"]`; assert:
  - `errors.Is(err, ErrUser)` (exit 1)
  - `strings.Contains(errBuf, "not inside any registered project")` (LD2)
  - `outBuf == ""` (§9 no cross-leakage)

- `"TestProjectHere_JSONFormat"` — create project `"bragfile"` with
  location `D`; override `getCwd → D/sub` (a subdirectory of D,
  confirming nearest-ancestor in CLI too); run `["here", "--format",
  "json"]`; assert:
  - `outBuf` is parseable JSON, begins with `{` (single object, LD3)
  - unmarshalled object has `"name": "bragfile"` and
    `"locations"` is a non-empty JSON array (hydrated via GetProject,
    NOT `[]`)
  - `errBuf == ""`

- `"TestProjectHere_UnknownFormatErrUser"` — override `getCwd → D`
  (any registered path); `["here", "--format", "xml"]`;
  assert `errors.Is(err, ErrUser)`. (LD4)

- `"TestProjectHere_HelpShowsExamples"` — run `["here", "--help"]`;
  assert `outBuf` contains `"Examples:"` AND `"brag project here
  --format json"` (a token unique to the locked `Long`, per SPEC-005
  lesson; this is a positive assertion only — §12 NOT-contains self-audit
  N/A since no Failing Test asserts the ABSENCE of any token).

> **Locked-decision ↔ test traceability (§9 rule).** LD1 (nearest-ancestor
> + separator guard) → `TestProjectForPath_SubdirectoryMatch` +
> `TestProjectForPath_LongestPrefixWins` + `TestProjectForPath_SeparatorGuard` +
> `TestProjectHere_JSONFormat` (subdirectory cwd in CLI). LD2 (plain
> one-liner, nil=user-error) → `TestProjectHere_MatchedProject` +
> `TestProjectHere_EmptyStateNote` + `TestProjectHere_NoMatch`. LD3
> (JSON via `ToProjectJSON`, locations hydrated) →
> `TestProjectHere_JSONFormat`. LD4 (`--format` default `""`) →
> `TestProjectHere_UnknownFormatErrUser` + help test. LD5 (nil not
> ErrNotFound) → `TestProjectForPath_NoMatch`.

## Implementation Context

*Read this section (and the files it points to) before the build cycle.*

### Decisions that apply

- **`DEC-019` (emitted by this spec)** — nearest-ancestor (longest-prefix)
  resolution policy; the exact algorithm; nil-on-no-match; separator
  guard; `filepath.Clean` normalization; `filepath.EvalSymlinks` out of
  scope. **Read DEC-019 before writing `ProjectForPath`.**
- **`DEC-017`** — `entries.project` stays free text; the
  `project_locations.path` global `UNIQUE` guarantee is what makes the
  longest-prefix tie-break well-defined (no two rows can share a path).
  `ProjectForPath` does NOT join on `entries`; it queries
  `project_locations JOIN projects` only.
- **`DEC-011` / `DEC-013`** — output shapes; `--format json` reuses
  `ToProjectJSON` (single object, same shape as `brag project show`).
- **`DEC-006` / `DEC-007`** — cobra command constructor pattern; inline
  `RunE` validation via `UserErrorf`; `--format` default `""`.

### Locked design decisions

**LD1 — Resolution policy: nearest-ancestor, longest-prefix (DEC-019).**
Confidence 0.90. See DEC-019 for full rationale; the short version:
"I'm deep in a project's tree; what am I in?" is the primary use case,
and exact-match-only fails it. Separator guard prevents
`/home/user/work` from matching `/home/user/worker`. `filepath.Clean` on
both sides handles minor path untidiness. `filepath.EvalSymlinks` is out
of scope (v0.2.0). *Rejected alternatives:* exact-match-only (Option A
— fails the primary use case), optional exact flag (Option C — YAGNI,
LD1 subsumes A).

**LD2 — Plain output: one-liner `<name>\t<status>\t<state_note or "-">`.
No-match → `UserErrorf("not inside any registered project")`.** Confidence
0.85. `brag project here` is a quick lookup command — "which project am I
in?" — not a full project view. A one-liner is more script-friendly and
appropriately minimal; `brag project show` already provides the full
multi-line view. `state_note` empty → `-` for consistency with how
`renderProjectPlain` renders empty notes. No-match is a user error (exit
1, stderr message) because "I'm inside a registered project" is a
precondition for the command to be meaningful; reporting it as an error
(not a silent empty output) is more predictable for scripts. Stdout empty
on no-match preserves pipeability. *Rejected alternative:* reuse
`renderProjectPlain` (multi-line) — rejected because `here`'s narrow
purpose is one line of orientation, not a full project view.

**LD3 — `--format json`: call `GetProject(p.ID)` to hydrate locations,
then `ToProjectJSON`.** Confidence 0.88. `ProjectForPath` does not
hydrate `Locations` (the resolver only needs scalar fields; SPEC-032
only needs the project `Name`). But `--format json` should return the
complete project shape (same as `brag project show --format json`),
including locations. The extra `GetProject` round-trip is one query and
is invisible at personal scale. *Rejected alternative:* hydrate locations
in `ProjectForPath` — rejected because SPEC-032 doesn't need them and
adding N+1 hydration or a more complex JOIN to the resolver complicates
the method for no benefit to the primary consumer.

**LD4 — `--format` default is `""`.** Confidence 0.90. Exactly as in
`brag project list`, `show`, `status`, `tag`, `tags` — the STAGE-006
flag-default WATCH item (N=3 confirmed; codify at stage close). The
format check is `format != "" && format != "json"` (matches all other
project commands). *Stated explicitly per the flag-default-explicitness
lesson.*

**LD5 — `ProjectForPath` returns `(*Project, error)`, nil on no match.**
Confidence 0.90. "Not inside any registered project" is a normal state
that the CLI renders as a user message — it is NOT the same as
"internal error" or "project not found by ID". Using nil (not
`ErrNotFound`) makes the caller's logic simple: `if p == nil { /* no
match */ }`. SPEC-032 follows the same pattern:
`p, err := s.ProjectForPath(cwd); if p != nil { /* auto-fill */ }`.
*Rejected alternative:* return `(Project, error)` wrapping `ErrNotFound`
— rejected because it conflates "no project here" (a normal navigation
state) with "you asked for a project that doesn't exist" (an error in
commands like `GetProject(id)`).

**LD6 — No optional path argument.** Confidence 0.92. Keeping
`brag project here` argument-free maintains simplicity (S complexity)
and focuses the command. Scripting use cases that need to resolve an
arbitrary path are better served by `brag project show` (by name/id).
SPEC-032's auto-fill also only needs `os.Getwd()`. *Rejected:* optional
`[path]` arg — YAGNI for v0.2.0.

### Constraints that apply

- `no-sql-in-cli-layer` (**blocking**) — `runProjectHere` calls
  `s.ProjectForPath(cwd)` and `s.GetProject(p.ID)` only; no
  `database/sql` import in `project.go`.
- `stdout-is-for-data-stderr-is-for-humans` (**blocking**) — matched
  project → data to stdout; `"not inside any registered project"`
  user error → stderr; all test assertions check both buffers.
- `storage-tests-use-tempdir` (**blocking**) — all new storage tests
  use `t.TempDir()`; paths used in `ProjectForPath` tests are
  controlled strings (not real filesystem paths at `~/.bragfile`).
- `errors-wrap-with-context` (warning) — `ProjectForPath` wraps:
  `fmt.Errorf("project for path %q: %w", cwd, err)`.
- `test-before-implementation` (blocking) — failing tests above are
  the design deliverable.
- `no-new-top-level-deps-without-decision` — `path/filepath` and `os`
  are stdlib; **no new top-level dependency**.

### `getCwd` injection for CLI tests

`runProjectHere` reads the cwd via a package-level variable:

```go
// getCwd is the function called to get the current working directory.
// Package-level so tests in the cli package can override it without
// the production binary carrying any test-only surface.
var getCwd = os.Getwd
```

In CLI tests, override it with `t.Cleanup` to restore the original:

```go
dir := t.TempDir()
orig := getCwd
getCwd = func() (string, error) { return dir, nil }
t.Cleanup(func() { getCwd = orig })
```

Tests do not call `t.Parallel()` (the package has none), so the global
override is safe under sequential execution. This is the same pattern
the `no-sql-in-cli-layer` constraint requires: no direct OS state in
tests.

### Prior related work

- `SPEC-027` (shipped, PR #40) — the `project_locations` schema,
  `AddLocation`, the global `UNIQUE(path)` guarantee.
- `SPEC-028` (shipped, PR #41) — the `--format ""` default pattern,
  `renderProjectPlain`, `resolveProjectByNameOrID`, the `runProjectCmd`
  CLI test helper; `internal/cli/project.go` already imports all needed
  packages except `"os"` and `"path/filepath"`.
- `SPEC-030` (shipped) — the `brag project status` dashboard confirms
  the tab-separated plain output style used by `here`.

### Out of scope (for this spec specifically)

- **Symlink resolution.** `filepath.EvalSymlinks` is explicitly deferred
  in DEC-019; do not add it.
- **Optional path argument** (`brag project here /some/path`). LD6.
- **`brag add --project` auto-fill.** SPEC-032.
- **Location editing** (`--add-path` / `--remove-path`). SPEC-033.

## Notes for the Implementer

### `ProjectForPath` (add to `internal/storage/project.go`)

Two new stdlib imports for `internal/storage/project.go`:
`"path/filepath"` (for `filepath.Clean`, `filepath.Separator`) and
`"strings"` (for `strings.HasPrefix`). Neither is currently in the
file — it imports `context`, `database/sql`, `errors`, `fmt`, `time`
only. Add both to the stdlib import group. The query loads all
location+project pairs in one round-trip; the Go loop applies DEC-019
policy. `Locations` is intentionally NOT hydrated (SPEC-032 doesn't
need it; the CLI JSON path calls `GetProject(p.ID)` separately).

```go
// ProjectForPath resolves cwd against all registered project_locations
// using nearest-ancestor (longest-prefix) matching (DEC-019). Both cwd
// and each stored path are cleaned with filepath.Clean before comparison.
// Returns nil, nil — not ErrNotFound — when no location is an ancestor
// of cwd; callers distinguish "no project here" from a real error by the
// nil check. Locations is not hydrated on the returned Project.
func (s *Store) ProjectForPath(cwd string) (*Project, error) {
	cwd = filepath.Clean(cwd)
	sep := string(filepath.Separator)

	ctx := context.Background()
	rows, err := s.db.QueryContext(ctx,
		`SELECT pl.path, p.id, p.name, p.status, p.state_note,
		        p.created_at, p.updated_at
		   FROM project_locations pl
		   JOIN projects p ON p.id = pl.project_id`,
	)
	if err != nil {
		return nil, fmt.Errorf("project for path %q: %w", cwd, err)
	}
	defer rows.Close()

	var (
		bestPath    string
		bestProject *Project
	)

	for rows.Next() {
		var (
			locPath                    string
			p                          Project
			createdAtRaw, updatedAtRaw string
		)
		if err := rows.Scan(&locPath, &p.ID, &p.Name, &p.Status, &p.StateNote,
			&createdAtRaw, &updatedAtRaw); err != nil {
			return nil, fmt.Errorf("project for path %q: %w", cwd, err)
		}
		cleanLoc := filepath.Clean(locPath)

		// Nearest-ancestor check: cwd must equal cleanLoc (exact match) or
		// start with cleanLoc + separator (cleanLoc is a parent directory).
		// The separator suffix prevents /home/user/work matching /home/user/worker.
		if cwd != cleanLoc && !strings.HasPrefix(cwd, cleanLoc+sep) {
			continue
		}

		// Longest prefix wins (most-specific registered ancestor, DEC-019).
		if bestProject == nil || len(cleanLoc) > len(bestPath) {
			created, err := time.Parse(time.RFC3339, createdAtRaw)
			if err != nil {
				return nil, fmt.Errorf("project for path %q: parse created_at %q: %w",
					cwd, createdAtRaw, err)
			}
			updated, err := time.Parse(time.RFC3339, updatedAtRaw)
			if err != nil {
				return nil, fmt.Errorf("project for path %q: parse updated_at %q: %w",
					cwd, updatedAtRaw, err)
			}
			p.CreatedAt = created.UTC()
			p.UpdatedAt = updated.UTC()
			// Locations intentionally nil — callers that need the full
			// location list call GetProject(p.ID).
			bestPath = cleanLoc
			cp := p
			bestProject = &cp
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("project for path %q: %w", cwd, err)
	}
	return bestProject, nil
}
```

### `here` subcommand (add to `internal/cli/project.go`)

Add at package level (before or after existing package-level vars):

```go
// getCwd is the function called to get the current working directory.
// Package-level so tests in the cli package can override it without
// the production binary carrying any test-only surface.
var getCwd = os.Getwd
```

Add `"os"` to the stdlib import group (alongside `"bufio"`, `"errors"`,
`"fmt"`, `"io"`, `"strconv"`, `"strings"`; add `"path/filepath"` too if
not already present — check before adding).

Add to `NewProjectCmd()` (after the existing `cmd.AddCommand` calls):

```go
cmd.AddCommand(newProjectHereCmd())
```

Also update `NewProjectCmd`'s `Short` to include `here`:

```go
Short: "Manage projects (new, list, show, status, edit, archive, delete, here)",
```

New command constructors (transcribe verbatim — `Long` is §12(a)-self-audited):

```go
func newProjectHereCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "here",
		Short: "Show which project the current directory belongs to",
		Long: `Resolve the current working directory against registered project locations.
Prints the matching project if the cwd is inside a registered location
(nearest-ancestor match — you may be in any subdirectory, not just the exact
registered root). If no registered project matches, exits 1.

Output is a single tab-separated line (default) or a JSON object (--format json)
with the full project shape including locations (same as 'brag project show').

Examples:
  brag project here                 # name<TAB>status<TAB>state-note
  brag project here --format json   # single project JSON object`,
		RunE: runProjectHere,
	}
	cmd.Flags().String("format", "", "output format (one of: json); default is plain")
	return cmd
}

func runProjectHere(cmd *cobra.Command, _ []string) error {
	format, _ := cmd.Flags().GetString("format")
	if format != "" && format != "json" {
		return UserErrorf("unknown --format value %q (accepted: json)", format)
	}

	cwd, err := getCwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
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

	p, err := s.ProjectForPath(cwd)
	if err != nil {
		return fmt.Errorf("resolve project: %w", err)
	}
	if p == nil {
		return UserErrorf("not inside any registered project")
	}

	out := cmd.OutOrStdout()
	if format == "json" {
		full, err := s.GetProject(p.ID)
		if err != nil {
			return fmt.Errorf("get project: %w", err)
		}
		body, err := export.ToProjectJSON(full)
		if err != nil {
			return fmt.Errorf("render project json: %w", err)
		}
		fmt.Fprintln(out, string(body))
		return nil
	}
	// Plain one-liner: name<TAB>status<TAB>state_note (LD2).
	note := p.StateNote
	if note == "" {
		note = "-"
	}
	fmt.Fprintf(out, "%s\t%s\t%s\n", p.Name, p.Status, note)
	return nil
}
```

**§12(a) self-audit on the `Long`:** the only test that checks `Long`
content is `TestProjectHere_HelpShowsExamples`, which asserts the
PRESENCE of `"Examples:"` and `"brag project here --format json"`.
Both tokens appear in the `Long` above. No test asserts the ABSENCE
of any token, so the §12 NOT-contains check is N/A.

### `docs/api-contract.md` — status-change UPDATE (literal)

Insert this section **after** `### brag project status` (around line
524, before `### brag project edit`). House style mirrors the adjacent
project sections:

```
### `brag project here` — show the project for the current directory (STAGE-007)

​```
brag project here
brag project here --format json
​```

Resolves `os.Getwd()` against registered project locations using
nearest-ancestor (longest-prefix) matching (DEC-019): you may be anywhere
inside a registered location's directory tree, not just at the exact root.
When multiple registered paths are ancestors of the cwd, the most specific
(longest) path wins.

- Plain output (default): a single tab-separated line
  `<name>\t<status>\t<state_note>` on stdout (`-` when state note is
  empty); stderr empty; exit 0.
- `--format json` — a single JSON object with the full project shape
  (same as `brag project show --format json`; `locations` hydrated).
- Not inside any registered project → stderr:
  `not inside any registered project`, exit 1, stdout empty.
- Unknown `--format` exits 1 (user error). No positional arguments;
  reads `os.Getwd()` only.
```

(The `​` zero-width marks above are only to escape the nested fences in
this spec — the real doc uses plain triple-backtick fences.)

### Gotchas

- **Two files named `project.go` are modified — keep them straight.**
  - `internal/storage/project.go` gets `ProjectForPath`. New imports:
    `"path/filepath"` and `"strings"` (neither is in the file today —
    its imports are `context`, `database/sql`, `errors`, `fmt`, `time`
    only). Add both to the stdlib import group.
  - `internal/cli/project.go` gets `getCwd` + the `here` subcommand.
    New imports: `"os"` and `"path/filepath"`. `"strings"` is already
    imported there (`strings.TrimSpace`, `strings.Join`) — do NOT
    add it again. Updated stdlib group:
    `bufio, errors, fmt, io, os, path/filepath, strconv, strings`.
- **`ProjectForPath` does not hydrate Locations.** The plain output renders
  `p.StateNote`, not locations. The JSON path calls `GetProject(p.ID)` to
  hydrate — two Store calls in the JSON path is intentional and fine at
  personal scale.
- **nil check, not ErrNotFound.** `if p == nil { return UserErrorf(...) }`
  — do NOT call `errors.Is(err, ErrNotFound)`.
- **`getCwd` is a package-level var**, not a constant. Tests override it;
  the production binary always gets `os.Getwd`. Declare it once, before
  the first function that uses it.
- **`NewProjectCmd()` Short string update.** Add `"here"` to the comma
  list so `brag project --help` shows it in the one-liner.
- **`gofmt -w .` + `go vet ./...`** before the PR; confirm
  `./brag project here --help` in the real binary.

### §12(b) design-time verification (run at design 2026-06-10)

The resolution algorithm (DEC-019) was verified against the test cases
by tracing the Go logic directly:

- `/home/user/work/./src` → `filepath.Clean` → `/home/user/work/src`;
  `strings.HasPrefix("/home/user/work/src", "/home/user/work/")` = true
  → matches `/home/user/work`. ✅
- `/home/user/worker`; `cwd != "/home/user/work"` and
  `strings.HasPrefix("/home/user/worker", "/home/user/work/")` = false
  (the string `/home/user/worker` does NOT start with `/home/user/work/`)
  → no match. ✅
- cwd `/a/b/sub/deep`: `/a/b` (len=4) matches, `/a/b/sub` (len=8)
  also matches; `len("/a/b/sub") > len("/a/b")` → `/a/b/sub` wins. ✅
- `strings.HasPrefix("/a/b/sub/deep", "/a/b/")` = true ✅;
  `strings.HasPrefix("/a/b/sub/deep", "/a/b/sub/")` = true ✅.

`§12(b)` note: no external tool validation needed (no SQL DDL being
introduced; the query is a plain SELECT JOIN; the algorithm is pure Go
traceable at design). The plain-output literal `fmt.Fprintf(out,
"%s\t%s\t%s\n", p.Name, p.Status, note)` is design-decided against the
test assertion `strings.Contains(out, "bragfile\tactive\tnext: cut v0.2.0")`
— both include the same three fields in the same order. ✅

---

## Build Completion

*Filled in at the end of the **build** cycle, before advancing to verify.*

- **Branch:** `feat/spec-031-brag-project-here-cwd-resolver`
- **PR (if applicable):** opened below
- **All acceptance criteria met?** yes
- **New decisions emitted:**
  - `DEC-019` — nearest-ancestor (longest-prefix) cwd-to-project
    resolution policy *(emitted at design; no build deviation)*
- **Deviations from spec:**
  - The `p == nil` no-match path uses `fmt.Fprintln(cmd.ErrOrStderr(), "not inside any registered project")` + `return fmt.Errorf("%w", ErrUser)` instead of `return UserErrorf(...)`. This is required because the test asserts `strings.Contains(errBuf, "not inside any registered project")`: with `SilenceErrors: true`, cobra does not write to errBuf — the function must write explicitly. The bare `fmt.Errorf("%w", ErrUser)` return gives exit 1 without double-printing in production. This is the spec's "SilentErr or equivalent" pattern, deferred to build-time decision since the test constraint surface wasn't fully analysed at design.
- **Follow-up work identified:** none

### Build-phase reflection (3 questions, short answers)

1. **What was unclear in the spec that slowed you down?**
   — The phrase "cobra RunE returning a SilentErr or equivalent" in the stdout/stderr discipline note required inference: with `SilenceErrors: true`, cobra never writes to errBuf, so the "not inside" message must be written explicitly before returning a bare `ErrUser` wrapper. The test asserting BOTH `errors.Is(err, ErrUser)` AND `strings.Contains(errBuf, msg)` is what made the pattern unambiguous — but that took a test-run failure to surface.

2. **Was there a constraint or decision that should have been listed but wasn't?**
   — The interaction between `SilenceErrors: true` and `errBuf`-asserting tests for user-error messages is a non-obvious pattern. The spec covered it via the "SilentErr or equivalent" hint, but a concrete example (parallel to "Aborted." in `runProjectDelete`) would have made it instantly clear. No new DEC needed; a sentence in the implementation context would have been sufficient.

3. **If you did this task again, what would you do differently?**
   — At design, trace through the test assertion for `TestProjectHere_NoMatch` against the cobra execution model (`SilenceErrors: true` + `root.SetErr(errBuf)`) to pre-decide the "write to errBuf then return bare ErrUser" implementation pattern. That would have made the build a pure transcription rather than a debug cycle.

---

## Reflection (Ship)

*Appended during the **ship** cycle. Outcome-focused reflection.*

1. **What would I do differently next time?**
   — <answer>

2. **Does any template, constraint, or decision need updating?**
   — <answer>

3. **Is there a follow-up spec I should write now before I forget?**
   — <answer>
