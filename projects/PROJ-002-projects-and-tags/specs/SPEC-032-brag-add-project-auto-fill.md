---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-032
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
  created_at: 2026-06-11

references:
  decisions: [DEC-017, DEC-019]
  constraints:
    - no-sql-in-cli-layer
    - stdout-is-for-data-stderr-is-for-humans
    - storage-tests-use-tempdir
    - errors-wrap-with-context
    - test-before-implementation
    - no-new-top-level-deps-without-decision
  related_specs: [SPEC-027, SPEC-028, SPEC-029, SPEC-030, SPEC-031, SPEC-033]
---

# SPEC-032: brag add — `--project` auto-fill from cwd

## Context

This is the **seventh spec of STAGE-007** (Projects core) and the last
one that touches capture. It delivers the brief's "**Capture knows where
it is**" success criterion: when `brag add` runs from inside a registered
project's directory and the user did *not* supply a project explicitly,
the entry's project is auto-filled from the cwd.

The work is small and entirely confined to `internal/cli/add.go`: no new
Store method, no migration, no new command. It reuses the resolver shipped
by **SPEC-031** — `Store.ProjectForPath(cwd)` (nearest-ancestor /
longest-prefix matching, **DEC-019**) — and replicates the `getCwd`
indirection pattern that SPEC-031 introduced in `internal/cli/project.go`,
under a separate `addGetCwd` var so the two files' tests inject
independently.

The write-path semantics agree with **DEC-017**: `entries.project` stays
free text. Auto-fill writes the *matched project's `Name`* into that
free-text column exactly as if the user had typed `--project <name>`. No
foreign key, no link column — the auto-filled value is just a string, and
`brag project status`'s soft-string-match count (`entries.project =
projects.name`, DEC-017/SPEC-030) will count it because the string equals
the project name by construction.

Parent stage: `STAGE-007-projects-core.md` (Spec Backlog → SPEC-032;
Design Notes question #2, which DEC-019 already resolved). PROJ-002 brief
governs dev/prod DB isolation. Prior foundation: **SPEC-031** (the
resolver + the `getCwd` precedent), **SPEC-027** (the
`project_locations` schema + global `UNIQUE(path)`), **SPEC-028** (the
`brag project new --path` registration path that creates the locations
this command resolves against).

## Goal

When `brag add` runs without an explicitly-provided project and the cwd
resolves to a registered project location, auto-fill the entry's project
from that project's name; otherwise leave it empty. An explicit project
(from a flag, the JSON body, or the editor buffer) always wins, and
auto-fill never fails the add.

## Inputs

- **Files to read:**
  - `internal/cli/add.go` — `runAdd` dispatcher, `runAddFlags`,
    `runAddEditor`, `getFlagString`. The three entry-building sites
    that gain auto-fill; `addGetCwd` is declared here.
  - `internal/cli/add_json.go` — `runAddJSON` and `parseAddJSON` /
    `addJSONInput`. The JSON path builds its `storage.Entry` from the
    stdin body (`in.Project`), **not** from the `--project` flag.
  - `internal/cli/project.go` — the `getCwd` precedent
    (`var getCwd = os.Getwd`, line 21) to mirror; `runProjectHere`'s
    `ProjectForPath` call pattern (LD5 nil-on-no-match).
  - `internal/storage/project.go` — `ProjectForPath(cwd string)
    (*Project, error)` signature + the nil-on-no-match contract (already
    shipped, SPEC-031); `Project.Name`.
  - `decisions/DEC-019` — the resolution policy this command reuses
    (read before build; do NOT re-derive the algorithm here).
  - `decisions/DEC-017` — `entries.project` soft string match; confirms
    the auto-filled value is a plain free-text string, no link.
  - `internal/cli/add_test.go` — the test harness (`newRootWithAdd`,
    `installAddEditFunc`, the `outBuf`/`errBuf` split); the existing
    explicit-project tests whose premise must be preserved.
  - `internal/cli/add_json_test.go` — JSON-path test harness
    (`root.SetIn(strings.NewReader(...))`).
- **External APIs:** none.
- **Related code paths:** `internal/cli/` only (no `internal/storage/`
  change), `docs/api-contract.md` (status-change doc update).

## Outputs

- **Files modified:**
  - `internal/cli/add.go`:
    - Add `var addGetCwd = os.Getwd` at package level (mirrors
      `project.go`'s `getCwd`; separate var so add-package tests inject
      independently — SPEC-031 precedent).
    - Add the shared helper `autoFillProject(s *storage.Store, explicit
      string, explicitSet bool) string` (the only new function).
    - In `runAddFlags`: set `Project:` via `autoFillProject(s,
      getFlagString(cmd, "project"), cmd.Flags().Changed("project"))`.
    - In `runAddEditor`: set `Project:` via `autoFillProject(s,
      parsed.Project, parsed.Project != "")`.
    - Add `"os"` to the stdlib import group (the file imports `"fmt"`,
      `"strings"` today; `"os"` is new).
    - Update the cobra `Long` with one sentence documenting auto-fill
      (locked LD6; paired with a help test).
  - `internal/cli/add_json.go`:
    - In `runAddJSON`: after the store opens and before `s.Add(entry)`,
      set `entry.Project = autoFillProject(s, entry.Project,
      entry.Project != "")`.
    - No new import (uses the package-level `autoFillProject`).
  - `docs/api-contract.md` — **status-change UPDATE**: add a
    `STAGE-007 (cwd --project auto-fill)` note to the `### brag add`
    section (see Notes for the literal).
- **Files modified (tests — additive):**
  - `internal/cli/add_test.go` — new auto-fill tests for the flag and
    editor paths plus the best-effort and explicit-wins guards (see
    Failing Tests). No rewrites of existing tests.
  - `internal/cli/add_json_test.go` — new auto-fill tests for the JSON
    path. No rewrites.
- **New exports:** none. `autoFillProject` and `addGetCwd` are
  unexported, package-internal.
- **Database changes:** **NONE.** No migration, no schema change, no new
  Store method. The resolver (`ProjectForPath`) and the schema it queries
  already shipped in SPEC-031 / SPEC-027.

### Premise audit (`projects/_templates/premise-audit.md`), run at design

Greps **run at design 2026-06-11** and reconciled below.

**1. Inversion / removal — NONE (the load-bearing guarantee).** SPEC-032
must not change `brag add` for a user who supplies a project explicitly.
The explicit-project signal is checked *before* any cwd resolution in all
three paths (flag `Changed("project")`; non-empty JSON `project`;
non-empty editor `Project:`), so every existing explicit-project test is
preserved verbatim. Greps run:
- `grep -rn '\-\-project\|"-p"\|Project:' internal/cli/add_test.go
  internal/cli/add_json_test.go` → the explicit-project tests
  (`TestAdd_AllOptionalFieldsPersisted` `:136/:167`, `TestAdd_ShorthandProject`
  `:358/:379`, the JSON `"project":"p"` case `add_json_test.go:173/:200`,
  the editor `Project: platform` case `:676/:720`) all pass an explicit
  project → `explicitSet == true` → value returned verbatim → **unchanged**.
- The no-project tests (`TestAdd_SuccessPrintsIDToStdoutOnly`,
  `TestAdd_OutputIsPipeable`, …) run against a **fresh `t.TempDir()` DB
  with zero `project_locations` rows**, so `ProjectForPath` returns `nil`
  regardless of the process's real cwd → project stays `""` → **unchanged**.
  (This is why no existing test needs `addGetCwd` injection: an empty
  `project_locations` table makes auto-fill inert.)
- **Zero planned rewrites; zero planned deletions.**

**2. Addition / count-bump — NONE.** No migration is added, so the
`schema_migrations` literal-count assertions are untouched. Grep run:
`grep -rn '0004_add_projects\|want 4\|count != 4\|schema_migrations'
internal/cli/` → no hits in the CLI package (those assertions live in
`internal/storage/*_test.go`, which this spec does not touch). The add
tests assert per-add entry counts (`len(entries) != 1`), but auto-fill
never changes *how many* entries an add creates — only the `project`
field of the one row — so no entry-count assertion is affected.

**3. Status change — `docs/api-contract.md` only.** SPEC-032 changes the
behavior of an existing command (`brag add`). Grep run:
`grep -rn "brag add" docs/ README.md` → reconciled:
- `docs/api-contract.md:31` `### brag add` — **UPDATE**: add the
  `STAGE-007 (cwd --project auto-fill)` note (literal in Notes).
- `docs/tutorial.md` `brag add` hits (`:505` `bragit() { brag add … -p
  "work"; }`, the flag/editor/json walkthroughs) — **STAY**. They are
  explicit-project or generic examples that auto-fill does not change;
  the comprehensive projects+tags tutorial walkthrough is **STAGE-008**
  (stage scope rule — only per-spec api-contract/data-model updates fold
  in here).
- `README.md`, `docs/architecture.md`, `docs/blog/*`,
  `docs/brag-entry.schema.json`, `docs/data-model.md` `brag add` hits —
  **STAY** (no status claim about cwd auto-fill is invalidated;
  `data-model.md`'s `brag add` line describes the insert, unaffected).

**4. §12(b) design-time pre-flight — N/A for migrations (none).** No DDL
is introduced, so there is no external-tool pre-flight (`goreleaser
check`, a migration driver run, etc.) to run. The only design-decidable
literals are (a) the Go helper, traceable by inspection (see §12(a) note
in Failing Tests), and (b) the `Long`/api-contract prose, audited under
the §12(a) self-audit in Failing Tests. The resolver algorithm itself was
already §12(b)-verified at SPEC-031 design; this spec only *calls* it.

**5. Cross-check.** All grep hits above reconciled against the enumerated
lists; no un-enumerated hit remained.

## Acceptance Criteria

Testable outcomes. Happy path, error cases, edge cases.

- [ ] **Flag path, auto-fill fires.** `brag add -t "x"` (no `--project`)
  run with `addGetCwd` resolving inside a registered location writes the
  entry with `project == <registered project name>` (LD1/LD2).
- [ ] **Flag path, explicit wins.** `brag add -t "x" -p "explicit"` from
  inside a registered location writes `project == "explicit"` — the cwd
  match is ignored (LD3).
- [ ] **Flag path, explicit empty is honored (NOT auto-filled).**
  `brag add -t "x" -p ""` from inside a registered location writes
  `project == ""` — `Changed("project")` is true, so the empty value is
  the user's intent and auto-fill does not fire (LD3, the
  `Changed`-vs-empty distinction).
- [ ] **JSON path, auto-fill fires.** `echo '{"title":"x"}' | brag add
  --json` from inside a registered location writes `project ==
  <project name>` (LD1/LD4).
- [ ] **JSON path, explicit wins.** `{"title":"x","project":"explicit"}`
  from inside a registered location writes `project == "explicit"` (LD4).
- [ ] **Editor path, auto-fill fires.** Editor mode with a buffer whose
  `Project:` header is empty, run from inside a registered location,
  writes `project == <project name>` (LD1/LD4).
- [ ] **Editor path, explicit wins.** Editor buffer with `Project:
  explicit`, from inside a registered location, writes `project ==
  "explicit"` (LD4).
- [ ] **Auto-fill is silent.** When auto-fill fires, stderr is empty
  (no "(project: …)" announcement); only the entry ID is on stdout
  (LD5, §9).
- [ ] **No-match → empty, silent, no error.** With no registered location
  matching the cwd, `brag add -t "x"` writes `project == ""`, exits 0,
  stderr empty (LD1, LD5).
- [ ] **Best-effort on resolver failure.** When `addGetCwd` returns an
  error, `brag add -t "x"` still succeeds (exit 0), writes `project ==
  ""`, and prints no error — auto-fill is suppressed, the add proceeds
  (LD2).
- [ ] **Help documents auto-fill.** `brag add --help` output contains the
  auto-fill sentence (LD6).
- [ ] **No SQL under `internal/cli/`.** `autoFillProject` calls
  `s.ProjectForPath` only; `add.go` / `add_json.go` import no
  `database/sql`.
- [ ] **No migration, no new Store method, no new command.**

## Failing Tests

Written at **design**; build makes them pass. All are **new/additive** —
the premise audit found zero inversions and no count-bump, so there are
**no rewrites or deletions** of existing tests. Storage helpers seeded in
these tests use `t.TempDir()` DBs (`storage-tests-use-tempdir`). The CLI
tests override the package-level `addGetCwd` for hermetic cwd control
(mirrors SPEC-031's `getCwd` override; see Implementation Context).

**Shared test setup pattern** (transcribe once as a local helper or
inline): open the test DB at `dbPath`, `CreateProject({Name:"bragfile"})`,
`AddLocation(id, dir)` for a `dir` under `t.TempDir()`, close; then
override `addGetCwd → func() (string, error) { return dir, nil }` with
`t.Cleanup` restoring the original.

### `internal/cli/add_test.go` (modify — additive)

- `"TestAdd_AutoFillFromCwd_FlagPath"` — register project `"bragfile"` at
  `dir`; override `addGetCwd → dir`; run `["--db", dbPath, "add", "-t",
  "x"]` (no `--project`); assert the persisted entry has
  `Project == "bragfile"`, `errBuf == ""` (silent, LD5/§9), exit 0.
  Covers LD1 (auto-fill fires) + LD2 (write-path) + LD5 (silent).

- `"TestAdd_AutoFillFromSubdir_FlagPath"` — register `"bragfile"` at
  `dir`; override `addGetCwd → filepath.Join(dir, "internal", "cli")`
  (a subdirectory); run `add -t "x"`; assert `Project == "bragfile"`.
  Confirms the nearest-ancestor reuse (DEC-019) flows through auto-fill,
  not just exact-dir.

- `"TestAdd_ExplicitProjectWins_FlagPath"` — register `"bragfile"` at
  `dir`; override `addGetCwd → dir`; run `add -t "x" -p "explicit"`;
  assert `Project == "explicit"` (the registered match is ignored). LD3.

- `"TestAdd_ExplicitEmptyProjectNotAutoFilled"` — register `"bragfile"`
  at `dir`; override `addGetCwd → dir`; run `add -t "x" -p ""`; assert
  `Project == ""` — the `Changed("project")` guard treats an explicit
  empty string as the user's intent, NOT a trigger for auto-fill. LD3.
  *(This is the precise reason the spec mandates `Changed("project")` over
  `getFlagString(cmd,"project") != ""`.)*

- `"TestAdd_NoMatchLeavesProjectEmpty_FlagPath"` — register NO project;
  override `addGetCwd → t.TempDir()` (some unregistered dir); run
  `add -t "x"`; assert `Project == ""`, `errBuf == ""`, exit 0
  (`ProjectForPath` returns nil → silent empty). LD1/LD5.

- `"TestAdd_AutoFillBestEffortOnCwdError"` — register `"bragfile"` at
  `dir` (so a match *would* exist); override `addGetCwd → func()
  (string, error) { return "", errors.New("boom") }`; run `add -t "x"`;
  assert the add **succeeds** (`err == nil`, exit 0), the entry persists
  with `Project == ""`, and `errBuf == ""` (no leaked error). Exercises
  the helper's `getCwd`-error suppression branch (LD2 best-effort). The
  symmetric `ProjectForPath`-error branch shares the same `err != nil →
  return ""` clause; it is not separately fault-injectable without a new
  Store seam (out of scope — "no new Store method"), so it is covered by
  inspection of the single suppression clause this test exercises.

- `"TestAdd_AutoFillFromCwd_EditorPath"` — register `"bragfile"` at
  `dir`; override `addGetCwd → dir`; `installAddEditFunc` writing a valid
  buffer with **no** `Project:` header (e.g. `"Title: shipped x\n\nbody\n"`);
  run `["--db", dbPath, "add"]` (editor mode); assert the persisted entry
  has `Project == "bragfile"`. LD4 (editor path).

- `"TestAdd_ExplicitProjectWins_EditorPath"` — register `"bragfile"` at
  `dir`; override `addGetCwd → dir`; `installAddEditFunc` writing a buffer
  with `Project: explicit`; run `add`; assert `Project == "explicit"`
  (the buffer's non-empty project wins). LD4.

- `"TestAdd_HelpMentionsAutoFill"` — run `["add", "--help"]`; assert
  `outBuf` contains a token unique to the locked `Long` auto-fill
  sentence — `"auto-fills --project"` (per the SPEC-005 unique-token
  lesson; positive assertion only — no NOT-contains, so the §12
  NOT-contains self-audit is N/A). LD6.

### `internal/cli/add_json_test.go` (modify — additive)

- `"TestAddJSON_AutoFillFromCwd"` — register `"bragfile"` at `dir`;
  override `addGetCwd → dir`; `root.SetIn(strings.NewReader(`{"title":"x"}`))`;
  run `["--db", dbPath, "add", "--json"]`; assert the persisted entry has
  `Project == "bragfile"`, `errBuf == ""`, exit 0. LD4 (JSON path) — note
  `--project` cannot be passed in JSON mode (mutually exclusive), so the
  explicit signal is the absent/empty body field.

- `"TestAddJSON_ExplicitProjectWins"` — register `"bragfile"` at `dir`;
  override `addGetCwd → dir`;
  `root.SetIn(strings.NewReader(`{"title":"x","project":"explicit"}`))`;
  run `add --json`; assert `Project == "explicit"` (non-empty body field
  is explicit, wins over cwd). LD4.

> **Locked-decision ↔ test traceability (§9 rule).** LD1 (auto-fill fires
> on no-explicit + match) → `TestAdd_AutoFillFromCwd_FlagPath` +
> `TestAdd_AutoFillFromSubdir_FlagPath` + `TestAdd_AutoFillFromCwd_EditorPath`
> + `TestAddJSON_AutoFillFromCwd`. LD2 (write project Name; best-effort) →
> `TestAdd_AutoFillFromCwd_FlagPath` (write) + `TestAdd_AutoFillBestEffortOnCwdError`
> (best-effort). LD3 (flag `Changed` guard, not emptiness) →
> `TestAdd_ExplicitProjectWins_FlagPath` + `TestAdd_ExplicitEmptyProjectNotAutoFilled`.
> LD4 (per-path explicit signal across all three modes) → the editor and
> JSON explicit-wins + auto-fill tests. LD5 (silent; no-match empty) →
> the `errBuf == ""` assertions + `TestAdd_NoMatchLeavesProjectEmpty_FlagPath`.
> LD6 (help documents it) → `TestAdd_HelpMentionsAutoFill`.

> **§12(a) note for build:** the expected `project` values in these tests
> are design-decidable — each equals the `Name` passed to `CreateProject`
> (`"bragfile"`) or the explicit literal (`"explicit"` / `""`). Transcribe
> them; they are not derived from a sort or a migration.

## Implementation Context

*Read this section (and the files it points to) before the build cycle.*

### Decisions that apply

- **`DEC-019`** — nearest-ancestor (longest-prefix) cwd resolution; the
  exact algorithm lives in `Store.ProjectForPath` (SPEC-031, already
  shipped). SPEC-032 **calls** it; do not re-implement or wrap the
  algorithm. The nil-on-no-match contract is what makes the helper's
  `if p == nil { return "" }` correct.
- **`DEC-017`** — `entries.project` is free text joined to `projects.name`
  by soft string match. Auto-fill writes the matched project's `Name`
  into the free-text column; because the written string *equals* the
  project name, SPEC-030's `entries.project = projects.name` count picks
  it up with no further work. No FK, no link column.
- **`DEC-011` / `DEC-012`** — the entry/JSON shapes are unchanged; only
  the `project` field's *value* is affected, and only when the user left
  it unspecified.

### Locked design decisions

**LD1 — Auto-fill triggers only when no explicit project is provided,
and only on a positive cwd match.** Confidence 0.95. The order in the
helper is: explicit-set short-circuit first, then cwd resolution. A nil
`ProjectForPath` result (no registered ancestor) leaves the project
empty. Reuses DEC-019 wholesale via `ProjectForPath`. *Rejected
alternative:* re-deriving the resolution inline in `add.go` — rejected
because it would duplicate DEC-019 logic and drift from `brag project
here`; the whole point of SPEC-031 exposing `ProjectForPath` was reuse.

**LD2 — Auto-fill writes `project.Name` and is best-effort.** Confidence
0.92. The matched `*Project`'s `Name` is the value written (a plain
string, DEC-017). "Best-effort" means: if `addGetCwd()` errors **or**
`ProjectForPath` errors, the helper returns `""` and the add proceeds
normally — a project we couldn't resolve must never fail a capture
(capture-first ethos; `brag add` is meant to be fast and reliable).
*Rejected alternative:* surfacing the resolver error to the user —
rejected because a transient cwd/db hiccup turning `brag add` into a
failure is a worse outcome than silently capturing with an empty project.

**LD3 — The flag-path explicit signal is `cmd.Flags().Changed("project")`,
NOT `value != ""`.** Confidence 0.95. `Changed` distinguishes "user
passed `-p ""`" (explicit empty — honor it, write `""`) from "user passed
no `-p`" (default empty — auto-fill). `getFlagString(cmd,"project") != ""`
cannot make that distinction and would auto-fill over an explicit empty
string. This is the accepted-constraint test, locked. *Rejected
alternative:* emptiness test — rejected as it conflates the two cases.

**LD4 — The "explicit project" question has a per-path answer; the
`Changed` test is the flag-path instance of one general principle ("did
the user already say what the project is?").** Confidence 0.90.
- **flag path:** `cmd.Flags().Changed("project")`.
- **JSON path:** `entry.Project != ""`. In JSON mode `--project` is
  *structurally impossible* (the dispatcher in `runAdd` rejects `--json`
  combined with any field flag, and `project` is a field flag), so the
  only explicit source is the stdin object's `"project"` field. A
  non-empty value is explicit; absent/empty triggers auto-fill. The
  decoder cannot distinguish `{"project":""}` from an absent key (both
  yield `""`), so an explicit empty JSON project is treated as
  "unspecified" → auto-fill — an accepted edge case (see Out of scope).
- **editor path:** `parsed.Project != ""`. Editor mode is entered only
  when *no* field flag is set, so `--project` is likewise structurally
  unset; the buffer's `Project:` header is the explicit source. Same
  empty-equals-unspecified treatment as JSON.

  Why this is not a contradiction of the accepted constraint: the
  constraint names `Changed("project")` as the test for "was `--project`
  explicitly provided," which is exactly right *where `--project` is a
  reachable input* — the flag path. JSON and editor reach the project
  field through a different door; the principle ("don't override an
  explicit project") is constant, the mechanism is per-door. *Rejected
  alternative:* applying `Changed("project")` uniformly — rejected
  because in JSON/editor it is *always* false, which would make auto-fill
  clobber an explicit `{"project":"foo"}` or `Project: foo` buffer. That
  is the exact inversion the premise audit forbids.

**LD5 — Auto-fill is silent (no stderr, no stdout noise).** Confidence
0.90. `brag add` is fast and scriptable; its only stdout is the new
entry's ID (pipeability, §9), and adding a "(project: myapp)" line to
stderr on every invocation from inside a project dir would be persistent
noise for zero functional gain — the user sees the result via `brag
list` / `brag project status`. The silent choice keeps every existing
`errBuf.Len() == 0` assertion valid and adds no new output contract.
This Q2 choice is ≥0.85 and confined to this command, so **no DEC** is
emitted. *Rejected alternative:* announce on stderr — rejected as
unwanted noise on the hot path; a `--verbose`/`-v` affordance is YAGNI
for v0.2.0 (revisit only if a user reports surprise at the auto-filled
value).

**LD6 — Document auto-fill in the `brag add` `Long` and
`docs/api-contract.md`.** Confidence 0.88. A behavior that silently
populates a field needs to be discoverable; the repo documents command
behavior in the cobra `Long` (cf. `brag project here`) and per-spec
api-contract entries (cf. SPEC-031). One sentence in each, paired with
`TestAdd_HelpMentionsAutoFill`. *Rejected alternative:* leave it
undocumented (silent *and* hidden) — rejected; silent-in-output is fine,
silent-in-docs is a discoverability bug.

**LD7 — Single shared helper `autoFillProject`, called once per path
(Q1 = Option C).** Confidence 0.90. The three paths source the project
from three different places (flag, JSON body, editor buffer), so there is
no single pre-dispatch value to set — **Option A (resolve once in
`runAdd` before dispatch) does not work**: `runAddJSON` and
`runAddEditor` never read the `--project` flag, so setting it in `runAdd`
would not reach them, and mutating flag state to smuggle a value into the
flag path is a hack. The auto-fill *logic* (cwd → `ProjectForPath` →
best-effort suppression), however, is identical across paths, so it
belongs in one helper that each path calls with its own `(explicit,
explicitSet)` pair. This is one new function and one line per call site.
*Rejected alternatives:* Option A (pre-dispatch — broken, above);
Option B (inline the logic in each path — three copies of the
best-effort suppression, more drift surface).

### Constraints that apply

- `no-sql-in-cli-layer` (**blocking**) — `autoFillProject` calls
  `s.ProjectForPath` (a Store method) only; neither `add.go` nor
  `add_json.go` imports `database/sql`. ✓ by construction.
- `stdout-is-for-data-stderr-is-for-humans` (**blocking**) — auto-fill
  writes nothing to either stream (LD5); the only stdout remains the
  entry ID; tests assert `errBuf == ""` when auto-fill fires.
- `storage-tests-use-tempdir` (**blocking**) — the project/location
  seeding in the new tests uses `t.TempDir()` DBs and `t.TempDir()`
  directories; no `~/.bragfile`, no real registered location.
- `errors-wrap-with-context` (warning) — `autoFillProject` returns no
  error (best-effort, LD2), so there is nothing to wrap; the surrounding
  `s.Add` / `config.ResolveDBPath` error wrapping is unchanged.
- `test-before-implementation` (blocking) — the Failing Tests above are
  the design deliverable.
- `no-new-top-level-deps-without-decision` — `"os"` is stdlib;
  `path/filepath` is not needed in `add.go` (cleaning happens inside
  `ProjectForPath`). **No new top-level dependency.**

### `addGetCwd` injection for CLI tests

Mirror SPEC-031's `getCwd` precedent, but as a **separate** package-level
var so the two `internal/cli` files inject independently (the SPEC-031
ship reflection explicitly flagged "future specs testing os-dependent
calls should call this out"):

```go
// addGetCwd is the function used to read the current working directory
// for --project auto-fill. Package-level so cli tests can inject a cwd
// without the production binary carrying test-only surface. Separate
// from project.go's getCwd (SPEC-031) so each file's tests override
// independently.
var addGetCwd = os.Getwd
```

Override pattern in tests (sequential package, no `t.Parallel()`):

```go
orig := addGetCwd
addGetCwd = func() (string, error) { return dir, nil }
t.Cleanup(func() { addGetCwd = orig })
```

### Prior related work

- `SPEC-031` (shipped, PR #44) — `Store.ProjectForPath`, DEC-019, and the
  `getCwd` indirection precedent this spec mirrors.
- `SPEC-027` (shipped) — `project_locations` + global `UNIQUE(path)`; the
  schema the resolver queries.
- `SPEC-028` (shipped) — `brag project new --path`, which registers the
  locations auto-fill resolves against.

### Out of scope (for this spec specifically)

- **`brag add --project ""` beyond the `Changed()` guard.** The flag
  path honors an explicit `-p ""` as empty (LD3). For JSON/editor,
  `{"project":""}` / an empty `Project:` header are indistinguishable
  from "absent" and are treated as unspecified → auto-fill (LD4). No
  further special-casing.
- **Announcing the auto-filled value** (a `--verbose` flag, a stderr
  note). Silent is locked (LD5); revisit only on a real user report.
- **Symlink resolution / any change to the resolution policy.** Owned by
  DEC-019; `filepath.EvalSymlinks` stays out (SPEC-031 scope).
- **Location editing** (`--add-path` / `--remove-path`). SPEC-033.
- **Any change to `brag list` / `brag show` / digest filtering.** Those
  read `entries.project` as before; this spec only affects how the
  `project` value is chosen at capture time.
- **A new Store method, a migration, or a new command.** None.

## Notes for the Implementer

Gotchas, style preferences, reuse opportunities.

### `add.go` — the helper + `addGetCwd` (transcribe verbatim)

Add `"os"` to the stdlib import group. Add the var + helper at package
level (place `addGetCwd` near the top, the helper anywhere after it):

```go
// addGetCwd is the function used to read the current working directory
// for --project auto-fill. Package-level so cli tests can inject a cwd
// without the production binary carrying test-only surface. Separate
// from project.go's getCwd (SPEC-031) so each file's tests override
// independently.
var addGetCwd = os.Getwd

// autoFillProject returns the project name to record on a new entry.
// When the user provided a project explicitly (explicitSet), that value
// is returned verbatim — even when empty, which is the user's intent
// (DEC-017 keeps entries.project free text). Otherwise the cwd is
// resolved against registered project locations (nearest-ancestor,
// DEC-019, via Store.ProjectForPath / SPEC-031) and the matching
// project's name is auto-filled. Auto-fill is best-effort: a getCwd
// error, a resolver error, or no match all yield "" and never fail the
// add. Silent by design (LD5) — it writes nothing to stdout/stderr.
func autoFillProject(s *storage.Store, explicit string, explicitSet bool) string {
	if explicitSet {
		return explicit
	}
	cwd, err := addGetCwd()
	if err != nil {
		return ""
	}
	p, err := s.ProjectForPath(cwd)
	if err != nil || p == nil {
		return ""
	}
	return p.Name
}
```

In `runAddFlags`, change the entry's `Project` field:

```go
	entry := storage.Entry{
		Title:       title,
		Description: getFlagString(cmd, "description"),
		Tags:        getFlagString(cmd, "tags"),
		Project:     autoFillProject(s, getFlagString(cmd, "project"), cmd.Flags().Changed("project")),
		Type:        getFlagString(cmd, "type"),
		Impact:      getFlagString(cmd, "impact"),
	}
```

In `runAddEditor`, change the inserted entry's `Project` field:

```go
	inserted, err := s.Add(storage.Entry{
		Title:       parsed.Title,
		Description: parsed.Description,
		Tags:        parsed.Tags,
		Project:     autoFillProject(s, parsed.Project, parsed.Project != ""),
		Type:        parsed.Type,
		Impact:      parsed.Impact,
	})
```

Add one sentence to the cobra `Long` (LD6). Insert it after the
JSON-mode paragraph and before the `Examples:` block — a clean spot that
keeps `Examples:` last so the existing help layout is preserved:

```
When --project is not given, brag auto-fills it from the current directory
if you are inside a registered project's location (see 'brag project here').
An explicit --project always wins, and auto-fill never fails the add.
```

(The unique help token the test asserts is `"auto-fills --project"`, so
keep that exact phrase. Adjust the sentence wording only if it still
contains that literal substring.)

### `add_json.go` — one line in `runAddJSON`

After `defer s.Close()` and **before** `s.Add(entry)`:

```go
	entry.Project = autoFillProject(s, entry.Project, entry.Project != "")
```

No new import — `autoFillProject` and `storage` are already in the
`cli` package / imported here.

### `docs/api-contract.md` — status-change UPDATE (literal)

Insert this block in the `### brag add` section, **after** the
`**STAGE-003 (JSON stdin form):**` block (ends ~line 105) and **before**
`### brag list` (~line 107). House style mirrors the existing
`**STAGE-NNN (...):**` sub-headers in this section:

```
**STAGE-007 (cwd `--project` auto-fill):**

When `--project` is not supplied, `brag add` resolves the current working
directory against registered project locations (nearest-ancestor match,
DEC-019) and auto-fills the entry's project from the matching project's
name. This applies to all three input modes:

- flag mode — fires only when `--project`/`-p` is not passed at all;
  passing `-p` (even `-p ""`) is an explicit choice and is recorded
  verbatim.
- JSON mode — fires only when the stdin object has no non-empty
  `"project"` field.
- editor mode — fires only when the buffer's `Project:` header is empty.

Auto-fill is silent (no stderr) and best-effort: if the cwd cannot be
resolved or matches no registered project, the entry is saved with an
empty project, exactly as before. An explicit project always wins over
the cwd. See `brag project here` (SPEC-031) for the shared resolver.
```

### Gotchas

- **`Changed("project")` only matters in the flag path.** In JSON and
  editor modes `--project` is structurally unreachable (dispatcher
  rules), so those paths pass `value != ""` as `explicitSet`. Do NOT try
  to read `cmd.Flags().Changed("project")` in `runAddJSON` /
  `runAddEditor` — it is always false there and would clobber an explicit
  body/buffer project (the inversion the premise audit forbids).
- **Place auto-fill AFTER the store opens, BEFORE `s.Add`.** The helper
  needs `s`. In the editor path this is also after the `!changed` abort
  check, so an aborted editor never resolves a cwd (correct — nothing is
  inserted).
- **Existing no-project tests stay green without `addGetCwd` injection.**
  Their temp DBs have no `project_locations` rows, so `ProjectForPath`
  returns nil regardless of the real cwd. Do not "fix" them.
- **Separate var, not a shared `getCwd`.** Declare `addGetCwd` in
  `add.go`; do not reuse `project.go`'s `getCwd` (per-file injection
  independence, SPEC-031 ship reflection).
- **`gofmt -w .` + `go vet ./...`** before the PR; confirm
  `./brag add --help` shows the auto-fill sentence and that
  `cd <a registered dir>; ./brag add -t test` records the project (smoke
  against the dev DB at `~/.bragfile-dev`, never the prod DB).

---

## Build Completion

*Filled in at the end of the **build** cycle, before advancing to verify.*

- **Branch:**
- **PR (if applicable):**
- **All acceptance criteria met?** yes/no
- **New decisions emitted:**
  - none expected (Q2 silent ≥0.85; resolution policy is DEC-019)
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
