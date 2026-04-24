---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-013
  type: story                      # epic | story | task | bug | chore
  cycle: ship
  blocked: false
  priority: medium
  complexity: S                    # S | M | L  (L means split it)

project:
  id: PROJ-001
  stage: STAGE-003
repo:
  id: bragfile

agents:
  architect: claude-opus-4-7
  implementer: claude-opus-4-7     # usually same Claude, different session
  created_at: 2026-04-23

references:
  decisions:
    - DEC-006  # cobra framework
    - DEC-007  # RunE-validated flag handling (applies if -P is ever mis-valued; currently a Bool so N/A)
  constraints:
    - no-sql-in-cli-layer
    - stdout-is-for-data-stderr-is-for-humans
    - errors-wrap-with-context
    - test-before-implementation
    - one-spec-per-pr
  related_specs:
    - SPEC-004  # shipped; original `brag list` command shape
    - SPEC-007  # shipped; list filter flags (--tag/--project/--type/--since/--limit) — same command, this spec adds an orthogonal flag
---

# SPEC-013: `brag list --show-project` / `-P` — surface project at scan time

## Context

First spec in STAGE-003. The `project` field is populated per-entry but
invisible at `brag list` scan time; today a user has to `brag show <id>`
or filter by a specific `--project` to see which initiative a row
belongs to. The user flagged this mid-STAGE-002 and the brief
subsequently elevated it to a STAGE-003 must-have ("I want to see
quickly what projects I have been working on") — "a scan-time nicety
that's become load-bearing now that corpus size makes visual scanning
useful" (brief, "Why Now"). SPEC-013 adds a standalone opt-in flag
(`-P`/`--show-project`) that inserts a project column into `brag list`
output without touching the byte-stable default shape that existing
scripts depend on.

Parent stage: `STAGE-003-reports-and-ai-friendly-i-o.md`. Project:
PROJ-001 (MVP). This spec is fully independent of the export trio
(SPEC-014/015/016) and the stdin-JSON spec (SPEC-017) — it can land
first or last within STAGE-003.

## Goal

Ship an opt-in `-P` / `--show-project` boolean flag on `brag list`
that renders a fourth tab-separated column `<id>\t<created_at>\t
<project>\t<title>` (empty project → literal `-`), while leaving
flag-off output byte-identical to what STAGE-002 shipped. The flag
composes freely with every existing STAGE-002 filter
(`--tag`/`--project`/`--type`/`--since`/`--limit`).

## Inputs

- **Files to read:**
  - `docs/api-contract.md` — `brag list` section (lines 78–92): current
    synopsis, output shape, ordering contract.
  - `docs/tutorial.md` — §4 "Read them back" (lines 123–168): current
    tab-separated docs + `cut -f3` examples.
  - `AGENTS.md` §9 (testing conventions — separate buffers, tie-break,
    fail-first, assertion specificity, locked-decisions-need-tests,
    premise-audit status-change case), §12 (CLI test harness rules).
  - `projects/PROJ-001-mvp/brief.md` — "Project column" detail section
    under "Detail on individual ideas" (options catalogue; option 2 is
    the one the stage selected).
  - `projects/PROJ-001-mvp/stages/STAGE-003-reports-and-ai-friendly-i-o.md`
    — "Success Criteria" bullet 1, "Scope: In scope" bullet 1,
    "Design Notes" → "Project column rendering (SPEC-013)"
    (authoritative design lock), "Premise-audit hot spots" → SPEC-013.
  - `/decisions/DEC-004-tags-comma-joined-for-mvp.md` — tags stay
    comma-joined TEXT; informs that `project` is a plain
    single-value TEXT field with no normalization needed for rendering.
  - `/decisions/DEC-006-cobra-cli-framework.md`
  - `/decisions/DEC-007-required-flag-validation-in-runE.md` — does
    not fire for a bool flag with a default, but the pattern stays
    honored (no `MarkFlagRequired`).
  - `/guidance/constraints.yaml`
  - `internal/cli/list.go` — existing `NewListCmd`/`runList` that this
    spec extends (36 lines today; adds one flag declaration, one
    dispatch check, two `Fprintf` branches).
  - `internal/cli/list_test.go` — existing tests; the shared
    `newListTestRoot`, `seedListEntry`, and `runListCmd` helpers are
    the pattern for the new tests.
  - `projects/PROJ-001-mvp/specs/done/SPEC-004-brag-list-command.md`
    — original `brag list` spec (output format lock that SPEC-013
    preserves under flag-off).
  - `projects/PROJ-001-mvp/specs/done/SPEC-007-list-filter-flags.md`
    — filter-flag spec whose test patterns (Failing Tests → `internal/
    cli/list_test.go`) are the template for SPEC-013's tests.
- **External APIs:** none.
- **Related code paths:** `internal/cli/`, `docs/`, `README.md`.

## Outputs

- **Files created:** none.
- **Files modified:**
  - `internal/cli/list.go` — declare `cmd.Flags().BoolP("show-project",
    "P", false, "include project in output (adds column between
    created_at and title)")`; in `runList`, read the flag once; choose
    the `Fprintf` format string based on the flag; render empty
    project as `-`.
  - `internal/cli/list_test.go` — add the seven tests listed in
    Failing Tests below; existing tests stay unchanged.
  - `docs/api-contract.md` — `brag list` synopsis gains
    `[-P|--show-project]`; add a sub-bullet documenting the flag-on
    output shape and the empty-project-renders-as-`-` rule; keep the
    default-shape bullet unchanged.
  - `docs/tutorial.md` — §4 "Read them back" gains a small block
    introducing `-P` (one example; note that `-P` output has project
    as field 3 and title as field 4, so scripts using `cut -f3` on
    plain output continue to work but `-P` consumers use `cut -f3`
    for project / `cut -f4` for title).
  - `README.md` — line 52's filter-flag list gains a mention of `-P`
    alongside the existing filter list (one-line status-claim
    update).
- **Files deleted:**
  - `projects/PROJ-001-mvp/specs/SPEC-014-brag-list-show-project-column.md`
    — duplicate empty-template file left behind when `just new-spec`
    ran twice (same slug). SPEC-014 is reserved in the stage backlog
    for `brag list --format json|tsv + export --format json + DEC-011`;
    removing the stray empty template clears the slot for the real
    SPEC-014 to scaffold later.
- **New exports:** none. The flag is local to the `list` command; no
  new function or type.
- **Database changes:** none.

## Locked design decisions

Reproduced here so build / verify don't re-litigate. Each is paired
with a failing test below per AGENTS.md §9 (SPEC-009 ship lesson).

1. **Flag name + shorthand:** `-P, --show-project`. Bool flag, default
   `false`. Exactly one flag, exactly these two forms; both must work
   and must produce identical output. Chosen over `-p` (collides with
   `add`'s `--project` shorthand; see SPEC-005). *Pair: `TestListCmd_
   ShowProject_ShortFormEquivalent` + `TestListCmd_ShowProject_
   HelpShowsFlag`.*
2. **Column ordering when flag on:** `<id>\t<created_at>\t<project>\t
   <title>`. Four columns, three tabs, project slotted between
   created_at and title. *Pair: `TestListCmd_ShowProject_
   AddsFourthColumn`.*
3. **Empty project → literal `-`:** when a row's `project` field is
   empty string (the Go zero for the field — `storage.Entry.Project`
   is a plain string, no NULL distinction at this layer), render the
   third column as `-`, not empty. *Pair: `TestListCmd_ShowProject_
   EmptyProjectRendersAsDash`.*
4. **Flag-off byte stability:** plain `brag list` output is byte-
   identical to STAGE-002 (`<id>\t<created_at>\t<title>`, exactly
   two tabs per line, no `-` filler, nothing changes). *Pair:
   `TestListCmd_PlainOutputByteIdenticalToSTAGE002`.*
5. **Orthogonal to STAGE-002 filters:** `-P` composes freely with
   `--tag`/`--project`/`--type`/`--since`/`--limit`. The filters pick
   rows; `-P` picks columns. *Pair: `TestListCmd_ShowProject_
   WithProjectFilter` (the redundant-looking-but-real case) and
   `TestListCmd_ShowProject_ComposedWithAllFilters`.*

**Out of scope (by design):**

- **`--pretty` bundle.** STAGE-003 ships standalone `-P`; STAGE-004's
  `--pretty` (emoji + bundled project column) can default `-P` on
  inside its bundle. Do NOT introduce `--pretty` in this spec.
- **`--columns <list>`.** Per-column selection was option 3 in the
  brief's four-option catalogue; rejected as overkill for MVP.
- **Changing the default to always include project.** Rejected by the
  brief (option 4) — would break existing scripts that parse 3
  tab-separated columns.
- **Literal-dash-project collision.** An entry whose `project` field
  is the literal string `"-"` will render identically to an
  empty-project row. Accepted as a silent cosmetic collision; not
  mentioned in user-facing flag help. User's explicit call (Q3
  during design).

## Premise audit (AGENTS.md §9 — status-change case)

The `brag list` output shape is a status claim documented in multiple
places. Default-mode output is unchanged (decision 4), so no existing
asserts or docs are invalidated — but the `-P` variant adds a new
shape worth announcing. Audit findings:

- `docs/tutorial.md` §4 (lines 123–168): "Prints tab-separated
  `<id>\t<created_at>\t<title>`" is still accurate for the default.
  Action: add a `-P` sub-block with one example and a note that the
  new column arrives between created_at and title (so `cut -f3` on
  default output still yields titles; `-P` users shift to `cut -f4`
  for titles / `cut -f3` for project).
- `docs/api-contract.md` §`brag list` (lines 78–92): synopsis line 81
  needs `[-P|--show-project]`; the "Output" bullet (line 89) stays
  correct for default; add a follow-on sub-bullet describing the
  flag-on shape + empty-rendering rule.
- `README.md` line 52: feature list mentions filters but not `-P`.
  Action: add a short `-P` mention alongside the filter list.
- Existing tests in `internal/cli/list_test.go`: none asserts that
  `brag list` has *only* the five filter flags, so no count-bump
  needed. `TestListCmd_HelpShowsFilters`'s needle list is a lower-
  bound check (`strings.Contains`), not an exhaustive match — adding
  `-P` to the command does not break it. Per Q1 during design, a
  dedicated `TestListCmd_ShowProject_HelpShowsFlag` is added rather
  than extending that list.

Additive-case and inversion-case scans both come up empty: no
tracked-collection count is coupled to the list command's flag count,
and no existing behavior is inverted by adding an opt-in flag with a
default-off state.

## Acceptance Criteria

Every criterion is testable; paired test name in italics.

- [ ] Plain `brag list` (no `-P`) produces output byte-identical to
      STAGE-002: exactly `<id>\t<created_at>\t<title>\n` per row, two
      tab characters per line, no project column, no `-` filler. No
      regression for existing scripts. *TestListCmd_
      PlainOutputByteIdenticalToSTAGE002*
- [ ] `brag list --show-project` produces
      `<id>\t<created_at>\t<project>\t<title>\n` per row — exactly
      three tab characters per line, project slotted between
      created_at and title. *TestListCmd_ShowProject_AddsFourthColumn*
- [ ] An entry with empty `project` field renders with `-` in the
      project column (third field), not an empty string.
      *TestListCmd_ShowProject_EmptyProjectRendersAsDash*
- [ ] `brag list -P` and `brag list --show-project` produce
      byte-identical output. *TestListCmd_ShowProject_ShortFormEquivalent*
- [ ] `brag list -P --project platform` filters to rows with
      `project == "platform"` AND renders the project column (showing
      `platform`). The flags are orthogonal — filtering does not
      suppress the column, and the column does not suppress filtering.
      *TestListCmd_ShowProject_WithProjectFilter*
- [ ] `brag list -P --tag T --project P --type K --since 7d
      --limit N` combines with every STAGE-002 filter in a single
      invocation without error, and the resulting rows all render
      with the four-column shape. *TestListCmd_ShowProject_
      ComposedWithAllFilters*
- [ ] `brag list --help` output includes the string `--show-project`
      AND the string `-P` AND a distinctive substring of the flag's
      one-line description (the literal `include project in output`
      — an §9 assertion-specificity choice, not the generic word
      "project" which already appears for `--project`).
      *TestListCmd_ShowProject_HelpShowsFlag*
- [ ] Every existing `TestListCmd_*` test in `list_test.go` stays
      green unchanged — no modifications to existing tests. *[manual:
      `go test ./internal/cli/ -run TestListCmd`]*
- [ ] `docs/api-contract.md`, `docs/tutorial.md`, and `README.md`
      mention `-P` / `--show-project` per the Outputs section above.
      *[manual grep: `grep -n 'show-project\|\-P' docs/api-contract.md
      docs/tutorial.md README.md` returns hits in all three]*
- [ ] `gofmt -l .` empty, `go vet ./...` clean, `CGO_ENABLED=0 go
      build ./...` succeeds, `go test ./...` green.
- [ ] The duplicate `SPEC-014-brag-list-show-project-column.md`
      template file is deleted. *[manual: `ls projects/PROJ-001-mvp/
      specs/ | grep SPEC-014` returns nothing]*

## Failing Tests

Written now. All follow AGENTS.md §9: separate `outBuf` / `errBuf`
with a no-cross-leakage assert on the empty side; fail-first run
before implementation; assertion-specificity on help substrings; one
failing test paired with each locked decision in the section above.

Reuse the existing helpers in `internal/cli/list_test.go`:

- `newListTestRoot(t)` — returns `(*cobra.Command, *outBuf, *errBuf)`.
- `seedListEntry(t, dbPath, title, tags, project, typ)` — inserts and
  returns a `storage.Entry`.
- `runListCmd(t, dbPath, args...)` — returns `(stdout, stderr, err)`.

No helper changes; no new helper needed. All new tests live in the
existing `internal/cli/list_test.go` file.

### `internal/cli/list_test.go` (seven new tests)

- **`TestListCmd_PlainOutputByteIdenticalToSTAGE002`** — seed three
  entries with known titles and mixed project values (one with
  `"platform"`, one with `""`, one with `"growth"`). Run `brag list`
  (no `-P`). Compute `expected` as a string concatenating, for each
  entry in reverse-insertion order, `fmt.Sprintf("%d\t%s\t%s\n",
  e.ID, e.CreatedAt.UTC().Format(time.RFC3339), e.Title)`. Assert
  `stdout == expected` exactly (byte-for-byte), `stderr == ""`, `err
  == nil`. This is the byte-stability lock from decision 4 — the
  strongest form, because it catches any accidental column addition
  or reformat.

- **`TestListCmd_ShowProject_AddsFourthColumn`** — seed one entry
  with `project: "platform"` and title `"solo"`. Run `brag list -P`.
  Trim trailing newline; assert:
  - `strings.Count(line, "\t") == 3` (exactly three tabs).
  - Splitting on `\t` yields four fields; field[0] parses as int64;
    field[1] parses as RFC3339 UTC; `fields[2] == "platform"`;
    `fields[3] == "solo"`.
  - `stderr == ""`.

- **`TestListCmd_ShowProject_EmptyProjectRendersAsDash`** — seed one
  entry with `project: ""` and title `"no-project"`. Run `brag list
  -P`. Assert the single line splits into four fields with
  `fields[2] == "-"` and `fields[3] == "no-project"`. (This test
  also passes if someone wrongly renders empty as empty — the
  critical assertion is the literal `-`.)

- **`TestListCmd_ShowProject_ShortFormEquivalent`** — seed two
  entries with distinct projects. Run `brag list --show-project` and
  `brag list -P` in sequence against the same DB path. Assert the
  two stdout strings are byte-identical, and both stderrs are empty.

- **`TestListCmd_ShowProject_WithProjectFilter`** — seed two entries:
  `("hit-platform", "", "platform", "")` and `("miss-growth", "", "growth",
  "")`. Run `brag list -P --project platform`. Assert:
  - `stdout` contains `"hit-platform"` and does NOT contain
    `"miss-growth"` (filter worked).
  - The single remaining line has exactly three tabs and
    `fields[2] == "platform"` (column is rendered).
  - `stderr == ""`.

- **`TestListCmd_ShowProject_ComposedWithAllFilters`** — seed three
  entries varying across tag/project/type so that exactly one row
  survives the combined filter:
  - `("hit", "auth", "platform", "shipped")` — survives.
  - `("miss-tag", "backend", "platform", "shipped")` — wrong tag.
  - `("miss-project", "auth", "growth", "shipped")` — wrong project.

  Run `brag list -P --tag auth --project platform --type shipped
  --since 1900-01-01 --limit 5`. (Using a far-past `--since` avoids
  backdating; all fresh rows pass the date filter.) Assert:
  - `stdout` contains `"hit"` only — not `"miss-tag"` or
    `"miss-project"`.
  - The single line has exactly three tabs; `fields[2] == "platform"`
    and `fields[3] == "hit"`.
  - `stderr == ""`.

- **`TestListCmd_ShowProject_HelpShowsFlag`** — `root.SetArgs([]
  string{"list", "--help"})`; execute; assert `outBuf.String()`
  contains each of `"--show-project"`, `"-P"`, and the literal
  substring `"include project in output"` (the distinctive part of
  the flag's one-line help string — NOT the generic `"project"`
  which already appears for `--project`). Assert `errBuf.Len() == 0`.

Notes for the implementer on testing patterns:

- **Fail-first run.** Before touching `list.go`, add all seven tests
  and run `go test ./internal/cli/ -run ShowProject` (plus the
  `PlainOutputByte...` test). Every test must fail for the expected
  reason: the flag doesn't exist yet (so `cobra` will reject `-P`
  and `--show-project` with an unknown-flag error); the
  `PlainOutputByte...` test already passes under today's code
  (because the default shape is already correct) — THIS IS FINE and
  expected, and serves as the regression lock. Document this
  unexpectedly-passing test in build completion's Deviations if
  you'd normally call out a passing-before-implementation test; the
  §9 SPEC-003 lesson is about *silently-passing because assertion
  too weak*, not *correctly-passing because implementation already
  exists for the default branch*. The byte-exact assertion in that
  test is specifically the lock preventing future regressions.
- **Test ordering inside the file.** Place the seven new tests
  immediately after `TestListCmd_HelpShape` (at the bottom of the
  file before any existing trailing utilities), grouped together and
  prefixed with a comment block naming SPEC-013. Keeps the diff
  reviewable.
- **Reuse `seedListEntry`'s return value.** For the byte-exact test,
  each call returns the hydrated `storage.Entry` with `ID` and
  `CreatedAt` set — use those to compute the expected output string
  rather than round-tripping through a second `List` call.
- **Do NOT modify `TestListCmd_HelpShowsFilters`.** Per Q1, the new
  help test is additive. Leave the existing one's needle list
  alone.

## Implementation Context

*Read before starting build. Self-contained handoff.*

### Decisions that apply

- `DEC-006` — Cobra framework. Use `cmd.Flags().BoolP(name, shorthand,
  default, usage)` to declare both `-P` and `--show-project` with
  one call; read via `cmd.Flags().GetBool("show-project")` in
  `runList`. Do NOT declare two separate flags.
- `DEC-007` — Required-flag validation in `RunE`. Does not fire here
  (the flag is a bool with a default; no "required" semantics). The
  pattern carries forward as a negative: do NOT reach for
  `MarkFlagRequired`, and do NOT introduce user-error paths for a
  bool flag that cannot be "invalid". Bool flag parse errors
  (`--show-project=notabool`) are cobra's job, not this spec's.
- `DEC-004` — Tags are comma-joined TEXT. Irrelevant to rendering
  (tags aren't rendered by `list`), but informs: `project` is also a
  plain single-value TEXT column with no normalization at the render
  layer. A future tag-normalization change would not affect SPEC-013.

### Constraints that apply

For `internal/cli/**`, `docs/**`, `README.md`:

- `no-sql-in-cli-layer` — blocking. `list.go` must not import
  `database/sql`. SPEC-013 changes only argv parsing and output
  formatting; no SQL touched.
- `stdout-is-for-data-stderr-is-for-humans` — blocking. The extra
  column goes to stdout (where rows already go). Nothing is added to
  stderr. Every happy-path test asserts `errBuf.Len() == 0`.
- `errors-wrap-with-context` — warning. No new error paths
  introduced; existing wraps stay unchanged.
- `test-before-implementation` — blocking. Write the seven new
  tests first, confirm six fail for the expected reason (unknown
  flag) and the byte-exact default-shape test passes as a regression
  lock, then implement.
- `one-spec-per-pr` — blocking. Branch
  `feat/spec-013-brag-list-show-project-column`. Diff touches only
  the files in Outputs.

### AGENTS.md lessons that apply

- §9 separate `outBuf` / `errBuf` in CLI tests (SPEC-001).
- §9 fail-first before implementation (SPEC-003).
- §9 assertion specificity — help test uses `"include project in
  output"`, not generic `"project"` (SPEC-005).
- §9 every locked design decision paired with a failing test
  (SPEC-009). Checked above: five locked decisions, five (of seven)
  test pairings explicitly enumerated in the Locked-design-decisions
  section; the remaining two tests cover byte-stability and
  composition-with-all-filters.
- §9 premise-audit three cases. Applied above under "Premise
  audit" — this spec is a status-change (documented output shape
  changes under opt-in flag) with neither existing-test inversion
  nor tracked-collection count-bump. Doc updates enumerated in
  Outputs rather than discovered at build time.

### Prior related work

- **SPEC-004** (shipped 2026-04-20). Original `brag list` command +
  the 3-column tab-separated output this spec extends. The byte-
  exact expected-output construction in
  `TestListCmd_PlainOutputByteIdenticalToSTAGE002` mirrors the
  `fmt.Fprintf("%d\t%s\t%s\n", ...)` line in `runList` — keep them
  in sync.
- **SPEC-007** (shipped 2026-04-20). Filter flags on `brag list`.
  `storage.ListFilter` and the existing `cmd.Flags().Changed()`
  dispatch pattern in `runList` are reused; SPEC-013 adds nothing
  to that filter-detection block. The `seedListEntry` / `runListCmd`
  / `mustBackdate` test helpers established there are what SPEC-013
  consumes.
- **SPEC-011 / SPEC-012** (shipped 2026-04-22). FTS5 + `brag
  search`. Unrelated to rendering, but a reminder that `brag search`
  produces the same 3-column output and is NOT in SPEC-013's scope
  — `-P` is a `brag list` feature only. If `brag search` ever grows
  a `-P` of its own, that's a separate spec.

### Out of scope (for this spec specifically)

If any of these feels necessary during build, write a new spec.

- **`-P` on `brag search`.** Out of scope. `search` has its own
  output rendering; SPEC-013 touches `list` only.
- **`--pretty` mode / emoji decoration.** STAGE-004 polish pass.
- **`--columns <list>`** per-column selection. Rejected by the brief.
- **Default-on project column.** Rejected by the brief.
- **Project-field normalization / trimming / case-folding.** The
  rendered string is the raw column value. If a project contains a
  tab, the output breaks column alignment — accepted trade-off for
  MVP, same as SPEC-004's treatment of tabs in titles.
- **Literal-dash-project cosmetic collision handling.** Accepted
  silently per Q3 during design.
- **Any change to `storage.Entry`, `storage.Store`, or
  `storage.ListFilter`.** This spec is pure CLI-layer.
- **Any new helper, sub-package, or exported function.** The
  implementation fits in ~10 lines of `list.go` plus a conditional
  `Fprintf`.

## Notes for the Implementer

- **Flag declaration.** In `NewListCmd`, after the existing five
  filter flag declarations, add:
  ```go
  cmd.Flags().BoolP("show-project", "P", false,
      "include project in output (adds column between created_at and title)")
  ```
  The usage string is load-bearing: `TestListCmd_ShowProject_
  HelpShowsFlag` asserts on the literal substring `"include project
  in output"`. Pick any phrasing, but make sure the test's needle
  matches what you write (update the spec if you change the
  wording — write it here first, then copy into code).

- **`runList` dispatch.** Read the flag once, just before the
  output loop:
  ```go
  showProject, _ := cmd.Flags().GetBool("show-project")
  ```
  Keep the existing `for _, e := range entries` loop; inside the
  loop, branch on the flag:
  ```go
  if showProject {
      project := e.Project
      if project == "" {
          project = "-"
      }
      fmt.Fprintf(out, "%d\t%s\t%s\t%s\n",
          e.ID,
          e.CreatedAt.UTC().Format(time.RFC3339),
          project,
          e.Title)
  } else {
      fmt.Fprintf(out, "%d\t%s\t%s\n",
          e.ID,
          e.CreatedAt.UTC().Format(time.RFC3339),
          e.Title)
  }
  ```
  Do NOT refactor the flag-off branch; it must remain the exact
  string `%d\t%s\t%s\n` with the three existing args. The byte-
  stability test locks this.

- **No helper extraction.** It is tempting to write a
  `formatListLine(e Entry, showProject bool) string` helper. Resist
  for SPEC-013 — the two-branch conditional is load-bearing,
  reads as a diff of exactly the new behavior, and adding a helper
  would obscure the "flag-off is byte-identical" contract. If the
  STAGE-004 `--pretty` spec wants to extract, it can do so then.

- **`Long` string update.** The existing `Long` string's Examples
  block has five lines covering the filter flags. Add one more line
  before the closing backtick:
  ```
    brag list -P                                    # include project column
  ```
  Two-space indent to match the existing rows; emoji-free; one-line
  description. `TestListCmd_ShowProject_HelpShowsFlag` does not
  assert on this example line (it asserts on the flag
  declaration's usage string), so the exact wording here is
  cosmetic — match the tone of the existing examples.

- **Doc updates (execute in this order).**
  1. `docs/api-contract.md` line 81 synopsis — change to
     `brag list [-P|--show-project] [--tag T] [--project P] [--type
     T] [--since 2026-01-01] [--limit N]`. Add a sub-bullet after
     the existing "Output: one line per entry, tab-separated
     columns" line:
     `- With \`-P\` / \`--show-project\`: output becomes \`<id>\t
     <created_at>\t<project>\t<title>\` (four tab-separated
     columns). Empty project renders as \`-\`.`
  2. `docs/tutorial.md` §4 — add a new subsection between "Filter
     flags" and "Search your entries":
     `### See project at scan time` with one example
     (`brag list -P`) and a note that the project column sits
     between created_at and title (scripts using `cut -f3` for
     titles should switch to `cut -f4` under `-P`).
  3. `README.md` line 52 — update the one-line feature summary to
     include `-P` alongside the existing filter mention. A minimal
     edit: after the parenthesized filter list, add
     `; add \`-P\` to include the project in output`.

- **SPEC-014 cleanup.** Before opening the PR, `rm projects/
  PROJ-001-mvp/specs/SPEC-014-brag-list-show-project-column.md`
  (the duplicate empty template). Do this in the same PR; mention
  it under Deviations if it surprises verify. The SPEC-014 slot is
  reserved in the stage backlog for the `brag list --format
  json|tsv` spec — leaving the stray empty template risks
  confusion when someone runs `just new-spec` for the real
  SPEC-014.

- **No new exports, no new files, no new packages.** The entire
  implementation diff should be ~15 lines across `list.go`,
  ~180 lines across `list_test.go` (seven test functions), and a
  handful of one-line doc edits.

- **Branch:** `feat/spec-013-brag-list-show-project-column`.

---

## Build Completion

*Filled in at the end of the **build** cycle, before advancing to verify.*

- **Branch:** `feat/spec-013-brag-list-show-project-column`
- **PR (if applicable):** (opened after `just advance-cycle`)
- **All acceptance criteria met?** yes
  - Fail-first confirmed: 6 of 7 new tests failed on unknown `-P` /
    missing `--show-project` substrings; `TestListCmd_PlainOutput
    ByteIdenticalToSTAGE002` passed immediately as the spec
    anticipated (regression lock on the already-correct default
    shape).
  - After implementation all 7 new tests pass, every prior
    `TestListCmd_*` is still green, `gofmt -l .` empty, `go vet
    ./...` clean, `CGO_ENABLED=0 go build ./...` succeeds,
    `go test ./...` and `just test` both green.
  - `docs/api-contract.md`, `docs/tutorial.md`, and `README.md` all
    mention `-P` / `--show-project`. Premise-audit grep
    (`grep -rn 'brag list' docs/ README.md`) verified: every
    default-mode mention correctly describes flag-off behavior that
    `-P` preserves; no stale "three columns" claims anywhere in the
    repo.
  - `SPEC-014-brag-list-show-project-column.md` duplicate template
    deleted; the slot is clear for the real SPEC-014 (JSON/TSV
    formats + DEC-011) to scaffold when STAGE-003 reaches it.
- **New decisions emitted:**
  - None. Spec anticipated no DEC and none proved necessary.
    Implementation is a pure cobra flag declaration
    (DEC-006 pattern) plus a two-branch `Fprintf` in `runList`;
    DEC-007 doesn't fire for a bool flag with a default, and the
    spec explicitly said the pattern carries forward as a negative
    (no `MarkFlagRequired`).
- **Deviations from spec:**
  - None. Followed the spec's literal code sketch for the flag
    declaration and the two-branch conditional. Usage string
    matches the spec's prescribed wording verbatim
    (`"include project in output (adds column between created_at
    and title)"`). No helper extracted. Doc edits match the three
    prescribed locations exactly. SPEC-014 template deletion
    handled as the spec instructed; noted here (not as a
    surprise) per the spec's "mention it under Deviations if it
    surprises verify" conditional.
- **Follow-up work identified:**
  - None. Out-of-scope items (`-P` on `brag search`, `--pretty`
    bundle, `--columns`, default-on, helper extraction) are
    already captured in the spec's Locked design decisions and
    Out-of-scope sections.

### Build-phase reflection (3 questions, short answers)

Process-focused: how did the build go? What friction did the spec create?

1. **What was unclear in the spec that slowed you down?**
   — Nothing, honestly. The spec handed over a literal code
   sketch for both the `BoolP` declaration and the two-branch
   `Fprintf`, named the exact usage-string substring the help
   test asserts on, listed the file-by-file doc edits in order,
   and called out the fail-first expectation that
   `PlainOutputByteIdenticalToSTAGE002` would pass immediately
   (with an explicit "this is NOT a SPEC-003 §9 silently-passing
   violation" note so I wouldn't misclassify it under
   Deviations). The premise audit under "status-change case" was
   pre-walked, so the doc sweep was a three-edit exercise rather
   than a grep-then-deliberate exercise. Zero build-time
   ambiguity.

2. **Was there a constraint or decision that should have been listed but wasn't?**
   — No. The five constraints (no-sql-in-cli-layer, stdout-is-
   for-data, errors-wrap-with-context, test-before-
   implementation, one-spec-per-pr) and two decisions (DEC-006
   cobra, DEC-007 validation pattern carried forward as a
   negative) were the exact levers this implementation needed.
   DEC-004 was cited as informational (tags unrelated to render)
   and that framing was correct.

3. **If you did this task again, what would you do differently?**
   — Nothing substantive. One tiny observation: when I wrote
   `TestListCmd_ShowProject_ComposedWithAllFilters`, my assertion
   `strings.Contains(out, "hit")` technically allows a false
   positive if a future miss-title happens to contain `hit` as a
   substring. The spec's seeded titles (`"hit"`, `"miss-tag"`,
   `"miss-project"`) don't collide, so the assertion is sound
   given the fixture, and the line-count + field-index asserts
   already lock the surviving-row shape tightly. Swapping the
   three Contains asserts for a full-line equality would be
   marginally stronger but also harder to read without an ID
   round-trip. I'd keep it as-is on a replay.

---

## Reflection (Ship)

*Appended during the **ship** cycle. Outcome-focused reflection, distinct
from the process-focused build reflection above.*

1. **What would I do differently next time?**
   — Almost nothing — SPEC-013 is the cleanest build-verify cycle so
   far. The one process nit: the build commit body claimed "Also
   removes the duplicate empty SPEC-014 template file…", but the
   diffstat showed no deletion because that file was never tracked in
   git (scaffolded and `rm`'d pre-first-commit). The end state is
   correct and verifiable by `ls`, but the commit message mildly
   misleads a reader grepping `git log -p` for the deletion. Next
   time: if you claim a cleanup in a commit body, make sure the
   cleanup actually appears in the diff — either by committing the
   scaffolding noise first in a throwaway `chore:` commit and then
   deleting it in the build commit, or by not mentioning it as a
   "removal" at all and just noting it under Deviations. Tiny hygiene
   item; everything else executed verbatim against the spec.

2. **Does any template, constraint, or decision need updating?**
   — No. The spec's premise-audit status-change walk (pre-enumerated
   doc locations) was the SPEC-012 ship lesson paying off exactly as
   intended: zero doc discoveries at build time, zero stale-status
   grep hits at verify. The five-decisions-five-tests pairing worked;
   the fact that two of the seven tests
   (`PlainOutputByteIdenticalToSTAGE002`, `ComposedWithAllFilters`)
   attach to decisions 4/5 as secondary pairs rather than 1:1
   primaries isn't a template defect — the spec's prose called it out
   and the decisions are still each covered. `_templates/spec.md`,
   `constraints.yaml`, and `AGENTS.md` §9 all survived unchanged;
   nothing new to lift out. If the clean-run streak continues
   (SPEC-013 makes three-in-a-row: 011 ship-lesson → 012 ship-lesson
   → 013 applied both), the §9 premise-audit trilogy is probably
   load-bearing in its current form.

3. **Is there a follow-up spec I should write now before I forget?**
   — No. The out-of-scope list (`-P` on `brag search`, `--pretty`
   bundle, `--columns`, default-on project column, `formatListLine`
   helper extraction) covers every tangent that surfaced. Of those,
   `--pretty` is already earmarked for STAGE-004 polish and
   `--columns` / default-on are already rejected by the brief —
   neither needs a backlog entry. Nothing new surfaced during build
   or verify that doesn't already have a home. Build Completion said
   "None" and I concur.
