---
task:
  id: SPEC-010
  type: story
  cycle: ship
  blocked: false
  priority: high
  complexity: S

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
    - DEC-006  # cobra framework
    - DEC-007  # RunE-validated flags AND positional args
    - DEC-009  # editor buffer format (net/textproto header + markdown body)
  constraints:
    - no-sql-in-cli-layer
    - stdout-is-for-data-stderr-is-for-humans
    - errors-wrap-with-context
    - timestamps-in-utc-rfc3339
    - test-before-implementation
    - one-spec-per-pr
  related_specs:
    - SPEC-003  # shipped; add command, flag-mode path
    - SPEC-005  # shipped; add shorthand flags, assertion-specificity lesson
    - SPEC-009  # shipped; internal/editor package, testEditFunc hook pattern
---

# SPEC-010: `brag add` no-args editor launch

## Context

Sixth spec in STAGE-002. Extends `brag add` with a second invocation
mode: if the user runs `brag add` with no field flags set, open
`$EDITOR` on a hint template, then parse + persist the saved buffer
via `Store.Add`. Flag mode (`brag add -t "x"`) continues to work
byte-identically.

This spec is the editor-launch's first user-visible payoff beyond
`edit`: it lets users compose rich narrative entries (multi-paragraph
descriptions, all five optional fields) in their real editor without
fighting shell quoting. The entire infrastructure — `internal/editor`
package with `Render` / `Parse` / `Launch` / `Default`, the
`testEditFunc` hook for subprocess-free CLI tests — landed in
SPEC-009. This spec is almost entirely integration work plus one new
helper (`editor.EmptyTemplate`).

Parent stage: `STAGE-002-capture-and-retrieval.md`. Project: PROJ-001.

## Goal

Dispatch `brag add` into one of two paths at runtime:

1. **Flag mode** (any of `--title/-t`, `--description/-d`,
   `--tags/-T`, `--project/-p`, `--type/-k`, `--impact/-i` set):
   existing path, byte-identical to SPEC-003/SPEC-005 behavior.
2. **Editor mode** (none of those flags set): render a hint template
   via `editor.EmptyTemplate()`, spawn `$EDITOR` via `editor.Launch`,
   parse on change, call `Store.Add`, print the new ID to stdout.
   Unchanged save aborts with exit 0 and `Aborted.` on stderr.
   Parse failure (e.g. empty Title) is `UserErrorf`, exit 1. Editor
   failure is wrapped internal error, exit 2.

The root `--db` persistent flag does NOT trigger flag mode (it's a
path override, not an entry field). Flag-mode / editor-mode
detection inspects only the six entry-field flags via
`cmd.Flags().Changed(name)`.

## Inputs

- **Files to read:**
  - `docs/api-contract.md` — `brag add` section has a "STAGE-002
    (editor-launch form)" block that needs updating now that the
    feature is shipping. Current text describes "YAML front-matter"
    which is wrong post-DEC-009; this spec corrects it.
  - `docs/tutorial.md` — §9 "What's NOT there yet" lists `brag add`
    with no args. This spec strikes that row.
  - `AGENTS.md` §8 (coding conventions), §9 (testing: separate
    buffers, fail-first, assertion specificity, locked-decision-
    needs-test), §12 "During design".
  - `/decisions/DEC-006-cobra-cli-framework.md`
  - `/decisions/DEC-007-required-flag-validation-in-runE.md`
  - `/decisions/DEC-009-editor-buffer-format.md` — the buffer
    format; `editor.EmptyTemplate()` in this spec produces the
    all-empty-headers variant of the DEC-009 shape.
  - `/guidance/constraints.yaml`
  - `internal/cli/add.go` — current `runAdd`; this spec splits it
    into dispatcher + `runAddFlags` (existing path, renamed) +
    `runAddEditor` (new path).
  - `internal/cli/add_test.go` — existing `newRootWithAdd` helper
    + all SPEC-003/SPEC-005 tests (must stay green — regression
    guard).
  - `internal/cli/edit.go` — reference for editor-launch + parse +
    store-write flow (exactly the shape `runAddEditor` mirrors).
  - `internal/cli/edit_test.go` — reference for `testEditFunc` hook
    usage in tests.
  - `internal/editor/editor.go` — `Render`, `Parse`, `Fields`;
    this spec adds `EmptyTemplate`.
  - `internal/editor/editor_test.go` — append one new test for
    `EmptyTemplate`.
  - `internal/storage/store.go` — `Store.Add` is unchanged; editor
    mode calls it with a hydrated `Entry` just like flag mode does.
- **External APIs:** none.
- **Related code paths:** `internal/cli/`, `internal/editor/`.

## Outputs

- **Files modified:**
  - `internal/cli/add.go` — `runAdd` becomes a dispatcher:
    ```go
    func runAdd(cmd *cobra.Command, args []string) error {
        for _, name := range []string{"title", "description", "tags",
            "project", "type", "impact"} {
            if cmd.Flags().Changed(name) {
                return runAddFlags(cmd, args)
            }
        }
        return runAddEditor(cmd)
    }
    ```
    The existing body becomes `runAddFlags` (same logic; no behavior
    change). A new `runAddEditor` is added that mirrors `runEdit`'s
    flow (render template → Launch → parse → Add).
  - `internal/cli/add_test.go` — append the new editor-mode tests
    (see Failing Tests); existing flag-mode tests stay untouched.
  - `internal/editor/editor.go` — add `EmptyTemplate() []byte`
    exported helper that returns a hint buffer with all five
    editable headers pre-listed with empty values. (Distinct from
    `Render(Fields{})` which omits empty headers per DEC-009.)
  - `internal/editor/editor_test.go` — append `TestEmptyTemplate_*`
    tests (see Failing Tests).
  - `docs/api-contract.md` — `brag add` "editor-launch form" block
    rewritten: correct format pointer (DEC-009, NOT YAML), exit
    codes, behavior on unchanged save.
  - `docs/tutorial.md` — §2 or §3 gains a short "editor-launch
    form" subsection; §9 strikes `brag add` with no args from
    "What's NOT there yet".
- **Files created:** none.
- **New exports:**
  - `editor.EmptyTemplate() []byte`
- **Database changes:** none.

## Locked design decisions (inline — no new DEC needed)

Each decision has a mapped failing test per the AGENTS.md §9
locked-decisions-need-tests rule (SPEC-009 ship lesson).

1. **No-flags-set triggers editor mode.** Dispatcher inspects the
   six entry-field flags only. If none was `Changed()`, editor mode.
   *Test: TestAddCmd_NoFlagsOpensEditor.*

2. **`--db` alone does NOT trigger flag mode.** Running `brag add
   --db /tmp/x.db` with no other flags still opens the editor.
   `--db` is a root-level persistent flag, not an entry field.
   *Test: TestAddCmd_DbFlagAloneStillOpensEditor.*

3. **Any single entry-field flag triggers flag mode.** Setting only
   `--description "x"` (without `--title`) must take the flag path,
   which then fails with `UserErrorf` because Title is empty — NOT
   silently open an editor. This keeps the flag-mode contract
   unambiguous.
   *Test: TestAddCmd_SingleFieldFlagForcesFlagMode.*

4. **Flag mode is unchanged.** Every SPEC-003/SPEC-005 test still
   passes byte-identically.
   *Test: all existing TestAdd_* tests remain green — regression
   guard is the existing suite, no new explicit test.*

5. **Editor-mode unchanged save: exit 0, "Aborted." to stderr.**
   Same shape as `brag add --yes`-declined in SPEC-008's delete
   flow: a deliberate user choice is not an error.
   *Test: TestAddCmd_EditorUnchangedBufferAborts.*

6. **Editor-mode parse failure: exit 1, ErrUser.** If the user
   saves a buffer without a valid `Title:` header, return
   `UserErrorf(...)`. The DB is unchanged (no row inserted).
   *Test: TestAddCmd_EditorParseErrorIsUserError.*

7. **Editor-mode editor exec failure: exit 2, not ErrUser.** If
   `editor.Launch` returns a non-nil error, wrap and return. The
   DB is unchanged.
   *Test: TestAddCmd_EditorErrorIsInternal.*

8. **Editor-mode success: stdout receives inserted ID, stderr
   empty.** Mirrors flag-mode's stdout contract so scripting works
   identically across modes: `id=$(brag add)` works if the user
   saves a valid buffer.
   *Test: TestAddCmd_EditorHappyPathPrintsIDToStdout.*

9. **`editor.EmptyTemplate()` contains all five header names with
   empty values.** Distinct from `editor.Render(Fields{})` which
   omits empty fields per DEC-009. The template is a UX hint, not
   a renderer output.
   *Test: TestEmptyTemplate_ContainsAllHeaders.*

10. **`editor.EmptyTemplate()` round-trips through `Parse` to an
    `ErrTitleRequired`-style error.** I.e., if the user saves the
    template unchanged, the SHA-256 check aborts BEFORE Parse
    runs. But if the user touches the buffer while leaving Title
    empty, Parse runs and returns a user-error. Both paths return
    exit 0 (unchanged) or exit 1 (parse error) respectively —
    never silently insert a Title-less row.
    *Tests: TestEmptyTemplate_ParsesToMissingTitleError,
    TestAddCmd_EditorUnchangedBufferAborts (covers the SHA path),
    TestAddCmd_EditorParseErrorIsUserError (covers the parse
    path).*

## Acceptance Criteria

- [ ] `brag add` with no flags invokes the (test-injected) editor
      function. *[TestAddCmd_NoFlagsOpensEditor]*
- [ ] `brag add --db /tmp/x.db` (only `--db` set) still opens the
      editor. *[TestAddCmd_DbFlagAloneStillOpensEditor]*
- [ ] `brag add --description "x"` (any single field flag) takes
      the flag path; Title-empty triggers `ErrUser` (DEC-007
      already enforced). Editor is NOT opened.
      *[TestAddCmd_SingleFieldFlagForcesFlagMode]*
- [ ] Editor mode, fake edit writes a valid buffer: inserts the
      row, prints the new ID to stdout, stderr empty.
      *[TestAddCmd_EditorHappyPathPrintsIDToStdout]*
- [ ] Editor mode, fake edit leaves the buffer unchanged: no row
      inserted; stderr contains `Aborted.`; stdout empty; exit 0.
      *[TestAddCmd_EditorUnchangedBufferAborts]*
- [ ] Editor mode, fake edit writes a buffer with empty `Title:`:
      no row inserted; `errors.Is(err, ErrUser)`; stdout empty.
      *[TestAddCmd_EditorParseErrorIsUserError]*
- [ ] Editor mode, fake edit returns an error: no row inserted;
      `err != nil`; `!errors.Is(err, ErrUser)`; stdout empty.
      *[TestAddCmd_EditorErrorIsInternal]*
- [ ] `editor.EmptyTemplate()` returns bytes containing each of
      `"Title:"`, `"Tags:"`, `"Project:"`, `"Type:"`, `"Impact:"`,
      in that order, each with no trailing value on its line.
      *[TestEmptyTemplate_ContainsAllHeaders]*
- [ ] `editor.Parse(editor.EmptyTemplate())` returns an error
      mentioning `"Title"` (empty title is a parse error, same as
      SPEC-009's existing contract).
      *[TestEmptyTemplate_ParsesToMissingTitleError]*
- [ ] All existing SPEC-003/SPEC-005 `TestAdd_*` tests pass
      unchanged. *[manual: go test ./...]*
- [ ] `brag add --help` still shows all six field flags + their
      shorthands (no regression to SPEC-005's help shape).
      Existing help test is sufficient guard; no new assertion
      needed, but if the `Long` description changes to mention the
      editor-mode, that change must preserve the existing
      `"Examples:"` label and per-flag substrings.
- [ ] `gofmt -l .` empty, `go vet ./...` clean, `CGO_ENABLED=0 go
      build ./...` succeeds, `go test ./...` green.
- [ ] `docs/api-contract.md` "`brag add`" editor-launch block
      rewritten to reference DEC-009 (not YAML).
- [ ] `docs/tutorial.md` struck the "`brag add` no args" row from
      §9; short subsection showing the editor-mode flow added.

## Failing Tests

Written now. Every new CLI test uses separate `outBuf` / `errBuf`
with no-cross-leakage asserts (§9 SPEC-001). Every output-shape
assertion targets distinctive content (§9 SPEC-005). Fail-first run
before implementation (§9 SPEC-003). Every locked design decision
above has at least one paired failing test (§9 SPEC-009).

### `internal/editor/editor_test.go` (append)

Imports: existing — `testing`, `strings`, package under test.

- **`TestEmptyTemplate_ContainsAllHeaders`** — call
  `editor.EmptyTemplate()`. Assert the returned bytes, as a string,
  contains exactly these five substrings in order:
  `"Title: \n"`, `"Tags: \n"`, `"Project: \n"`, `"Type: \n"`,
  `"Impact: \n"`. Use `strings.Index` to confirm each appears AFTER
  the previous (ordered presence).

- **`TestEmptyTemplate_EndsWithBlankLine`** — the header block must
  be followed by a blank line (per DEC-009, header / body separator).
  Assert `strings.HasSuffix(string(tpl), "Impact: \n\n")`.

- **`TestEmptyTemplate_ParsesToMissingTitleError`** — call
  `editor.Parse(editor.EmptyTemplate())`. Assert non-nil error;
  `err.Error()` contains `"title"` (case-insensitive match
  acceptable via `strings.Contains(strings.ToLower(err.Error()),
  "title")`).

### `internal/cli/add_test.go` (append)

Reuse the existing `newRootWithAdd(t) (*cobra.Command, string)`
helper. Tests that exercise editor mode set `testEditFunc` at the
top of the test and register a `t.Cleanup(func() { testEditFunc =
nil })` reset (mirrors the SPEC-009 `edit_test.go` pattern).

Imports: existing — `bytes`, `errors`, `strings`, `testing`,
`strconv`, storage, cli.

- **`TestAddCmd_NoFlagsOpensEditor`** — set `testEditFunc` to a
  fake that writes `"Title: hello\n\n"` to the temp file and
  returns nil. `SetArgs([]string{"add"})`. Execute. Assert:
  - err nil
  - `outBuf` contains a single numeric line (the new ID)
  - `errBuf.Len() == 0`
  - `Store.List(ListFilter{})` on the same DB has one entry with
    `Title == "hello"`.

- **`TestAddCmd_DbFlagAloneStillOpensEditor`** — same setup, but
  `SetArgs([]string{"add", "--db", dbPath})`. Same assertions.
  (Structurally, `newRootWithAdd` already wires `--db` via the
  persistent flag; passing it explicitly confirms that `--db` is
  not treated as a field flag by the dispatcher.)

- **`TestAddCmd_SingleFieldFlagForcesFlagMode`** — do NOT set
  `testEditFunc` (leave nil; if dispatcher routes to editor mode
  incorrectly, the real `editor.Default` would try to spawn a
  subprocess and the test would hang or fail). `SetArgs(
  []string{"add", "--description", "only a description"})`.
  Execute. Assert:
  - err != nil (flag mode sees empty Title, returns `ErrUser`)
  - `errors.Is(err, ErrUser)`
  - No rows in the DB.
  (If this test hangs instead of returning `ErrUser`, the
  dispatcher is incorrectly routing `--description`-only to editor
  mode. Hanging is the fail-first signal.)

- **`TestAddCmd_EditorHappyPathPrintsIDToStdout`** — fake editor
  writes a full buffer:
  ```
  Title: cut p99 latency
  Tags: auth,perf
  Project: platform
  Type: shipped
  Impact: unblocked mobile v3

  redis-backed session cache.
  ```
  Execute. Assert nil error; `outBuf` is a single numeric line
  followed by newline; parse it via `strconv.ParseInt`; open the
  store and `Get(id)` returns the entry with the matching title,
  description, etc.

- **`TestAddCmd_EditorUnchangedBufferAborts`** — fake editor is a
  no-op (returns nil without modifying the file). Execute. Assert:
  - err nil
  - `outBuf.Len() == 0` (no ID printed)
  - `errBuf` contains the distinctive literal `"Aborted."`
  - DB is empty (no row inserted).

- **`TestAddCmd_EditorParseErrorIsUserError`** — fake editor writes
  a buffer with the `Title:` header set but empty (e.g. the
  template with no changes except a whitespace touch to trigger
  the changed=true path, then kept empty). Actually simpler:
  fake writes `"Tags: foo\n\n"` (no Title header at all). Execute.
  Assert `errors.Is(err, ErrUser)`; `outBuf.Len() == 0`; DB empty.

- **`TestAddCmd_EditorErrorIsInternal`** — fake editor returns
  `errors.New("boom")`. Execute. Assert `err != nil`, `!errors.Is(
  err, ErrUser)`, `outBuf.Len() == 0`, DB empty.

Notes for the implementer on testing patterns:

- Fail-first: after writing the new tests, run `go test ./...` once
  BEFORE any implementation change. The editor-mode tests should
  fail because the dispatcher doesn't exist yet (flag mode is still
  the only path). `TestEmptyTemplate_*` should fail because
  `EmptyTemplate` doesn't exist yet. If any unexpectedly passes,
  tighten the assertion.
- Every existing `TestAdd_*` test should continue to pass
  throughout; if any turn red during the refactor (splitting
  `runAdd` into dispatcher + `runAddFlags`), the split introduced
  a regression. The whole point of the split is behavioral parity.
- Use `t.Cleanup(func() { testEditFunc = nil })` in every test
  that sets `testEditFunc`. Missing cleanup would leak hooks into
  subsequent tests.

## Implementation Context

*Read before starting build. Self-contained handoff.*

### Decisions that apply

- `DEC-006` — Cobra. Dispatcher lives in `runAdd`; cobra wiring
  (flags, Use, Short, Long) does not change.
- `DEC-007` — RunE-validated args. Flag-mode Title-empty path is
  unchanged (DEC-007 already covers it). Editor-mode parse-error
  path returns via `UserErrorf`, consistent with DEC-007's
  principle.
- `DEC-009` — Editor buffer format. `editor.EmptyTemplate()`
  produces a buffer in the DEC-009 shape with all headers present
  but empty-valued. Parser already tolerates empty-valued headers
  (returns them as empty strings in `Fields`); only empty `Title`
  triggers an error. No format change.

### Constraints that apply

For `internal/cli/**`, `internal/editor/**`, `docs/**`:

- `no-sql-in-cli-layer` — blocking. `add.go` imports only
  `config`, `storage`, `editor` (+ stdlib). Unchanged from SPEC-003.
- `stdout-is-for-data-stderr-is-for-humans` — blocking. Editor
  mode: new ID → stdout; "Aborted." → stderr. Flag mode: unchanged.
  Every test asserts both streams.
- `errors-wrap-with-context` — warning. Every returned error is
  wrapped: `fmt.Errorf("resolve db path: %w", err)`,
  `fmt.Errorf("open store: %w", err)`, `fmt.Errorf("launch
  editor: %w", err)`, `fmt.Errorf("add entry: %w", err)`.
- `timestamps-in-utc-rfc3339` — blocking. `Store.Add` handles this
  already; no new timestamp code in `runAddEditor`.
- `test-before-implementation` — blocking. Fail-first run as noted.
- `one-spec-per-pr` — blocking. Branch
  `feat/spec-010-brag-add-no-args-editor-launch`.

### AGENTS.md lessons that apply

- §9 separate `outBuf` / `errBuf` (SPEC-001).
- §9 assertion specificity (SPEC-005) — tests assert on
  `"Aborted."` literal, not generic "brag add" strings.
- §9 fail-first (SPEC-003).
- §9 locked-decisions-need-tests (SPEC-009) — the 10 locked design
  decisions above each map to a failing test (decisions 2 and 3 —
  `--db`-alone and single-field-flag — each get their own test
  rather than being inferred).
- §12 "During design" — every implementation option below passes
  every blocking constraint. No "either is acceptable."

### Prior related work

- **SPEC-003** (shipped). Current `runAdd` that becomes
  `runAddFlags`. Zero behavior change to this path — the only
  source-code change is renaming the function and adding a new
  dispatcher `runAdd` above it.
- **SPEC-005** (shipped). Shorthand flags. Preserved unchanged.
- **SPEC-009** (shipped). `internal/editor` package: `Render`,
  `Parse`, `Fields`, `EditFunc`, `Launch`, `Default`. The
  `testEditFunc` hook pattern in `internal/cli/edit.go` — this
  spec adds a matching pattern in `add.go`. Decide: same package-
  level var (`testEditFunc` already exists and is accessible from
  `add.go` since both live in the `cli` package) or a new var
  (`testAddEditFunc`)? See Notes.

### Out of scope (for this spec specifically)

If any of these feels necessary during build, write a new spec.

- **Flag-based partial update** (`brag update <id> -t "new"`). Still
  deferred per user decision.
- **Template customization** — users cannot (yet) pre-seed the
  template with default values. If someone wants `--project
  platform` always prefilled in editor mode, that's a future spec
  (and possibly a shell alias / function in the meantime).
- **Pre-save hooks** — e.g., refusing to save if Impact is empty.
  Only Title is required per DEC-009.
- **Alternate buffer formats** — DEC-009 is the format.
- **Per-user template file** — e.g., `~/.bragfile/add-template.md`
  overriding `editor.EmptyTemplate()`. Future polish.
- **TUI mode** — still indefinitely deferred.

## Notes for the Implementer

- **Share or duplicate `testEditFunc`?** `edit.go` already declares
  `var testEditFunc editor.EditFunc` at the package level in the
  `cli` package. `add.go` is in the same package. **Sharing is
  correct:** both commands funnel through the same test hook.
  Test files that set `testEditFunc` for `add` tests should use
  the same name — one lock means one reset, simpler mental model.
  If build prefers a distinct `testAddEditFunc`, call it out in
  Build Completion deviations. Author leans share.

- **Dispatcher shape in `add.go`.**
  ```go
  func runAdd(cmd *cobra.Command, args []string) error {
      fieldFlags := []string{"title", "description", "tags",
          "project", "type", "impact"}
      for _, name := range fieldFlags {
          if cmd.Flags().Changed(name) {
              return runAddFlags(cmd, args)
          }
      }
      return runAddEditor(cmd)
  }
  ```
  The existing body of `runAdd` (from SPEC-003/SPEC-005) becomes
  `runAddFlags(cmd *cobra.Command, args []string) error` —
  identical logic, renamed.

- **`runAddEditor` shape.**
  ```go
  func runAddEditor(cmd *cobra.Command) error {
      editFn := testEditFunc
      if editFn == nil {
          editFn = editor.Default
      }
      edited, changed, err := editor.Launch(editor.EmptyTemplate(), editFn)
      if err != nil {
          return fmt.Errorf("launch editor: %w", err)
      }
      if !changed {
          fmt.Fprintln(cmd.ErrOrStderr(), "Aborted.")
          return nil
      }
      f, err := editor.Parse(edited)
      if err != nil {
          return UserErrorf("invalid buffer: %v", err)
      }
      dbFlag := getFlagString(cmd, "db")
      path, err := config.ResolveDBPath(dbFlag)
      if err != nil {
          return fmt.Errorf("resolve db path: %w", err)
      }
      s, err := storage.Open(path)
      if err != nil {
          return fmt.Errorf("open store: %w", err)
      }
      defer s.Close()
      inserted, err := s.Add(storage.Entry{
          Title: f.Title, Description: f.Description,
          Tags: f.Tags, Project: f.Project,
          Type: f.Type, Impact: f.Impact,
      })
      if err != nil {
          return fmt.Errorf("add entry: %w", err)
      }
      fmt.Fprintln(cmd.OutOrStdout(), inserted.ID)
      return nil
  }
  ```

- **`editor.EmptyTemplate()` shape.**
  ```go
  // EmptyTemplate returns a buffer shaped per DEC-009 with all five
  // editable headers pre-listed but empty. Distinct from
  // Render(Fields{}) which omits empty-valued headers — the template
  // is a UX hint for first-time editor-mode use, not a renderer
  // output.
  func EmptyTemplate() []byte {
      return []byte("Title: \nTags: \nProject: \nType: \nImpact: \n\n")
  }
  ```
  Keep the trailing double-newline: the first `\n` terminates the
  `Impact: ` header line; the second `\n` is the blank line
  separator between the header block and the (empty) body.

- **`Long` description for `add`.** The SPEC-005 `Long` lists
  flags and an Examples block. Add a short paragraph at the start
  (before the existing Examples) explaining the two modes:
  ```go
  Long: `Add a new brag entry, either via flags or by opening $EDITOR.

  Flag mode (any of --title, -d, -T, -p, -k, -i set): inserts directly
  from flag values. --title is required.

  Editor mode (no flags set): opens $EDITOR on a template buffer.
  Save a valid entry to insert it. Save unchanged to abort cleanly.

  Examples: ...
  `
  ```
  Preserve the existing `Examples:` block and the existing
  shorthand legend line — the SPEC-005 `TestAdd_HelpShowsExamples`
  and `TestAdd_HelpShowsShorthands` assertions must continue to
  pass.

- **`docs/api-contract.md` rewrite.** The current "STAGE-002
  (editor-launch form)" block describes "YAML front-matter" which
  is wrong post-DEC-009. Replace with:
  ```
  brag add            # no args → opens $EDITOR on a template buffer
  ```
  and a prose block pointing at DEC-009 for the format, plus exit
  codes (0 on success or abort; 1 on parse error; 2 on editor
  exec failure). Match the structure of the `brag edit <id>`
  section that landed in SPEC-009.

- **`docs/tutorial.md` update.**
  1. §9 "What's NOT there yet": strike the `brag add` with no args
     row.
  2. Somewhere in §2 or §3 (where full-metadata capture is shown),
     add a short subsection showing editor-mode invocation + the
     template preview. Tutorial's voice: minimal, practical.

- **No `init()` functions** (§8).

---

## Build Completion

*Filled in at the end of the **build** cycle, before advancing to verify.*

- **Branch:** `feat/spec-010-brag-add-no-args-editor-launch`
- **PR (if applicable):** opened post-build (link in PR description)
- **All acceptance criteria met?** yes (with one clarified deviation; see below)
- **New decisions emitted:**
  - (none; DEC-009 from SPEC-009 covers the format)
- **Deviations from spec:**
  - **Deleted `TestAdd_MissingTitleIsUserError`** (SPEC-003).
    The spec's acceptance criteria included "all existing
    SPEC-003/SPEC-005 `TestAdd_*` tests pass unchanged", but locked
    design decision #1 ("no field flags → editor mode") directly
    obsoletes that test's premise — `brag add` (no args) is no longer
    a user error, it now opens the editor. Under the new dispatcher
    the test would call `editor.Default` (no `testEditFunc`
    installed) and hang spawning `vi`. Coverage of the
    title-required-in-flag-mode contract is preserved by
    `TestAdd_EmptyTitleIsUserError`,
    `TestAdd_WhitespaceTitleIsUserError`, and
    `TestAdd_EmptyShorthandTitleIsUserError`. The deleted test was
    replaced with an explanatory comment in `add_test.go`.
  - **`runAddFlags` accepts `args []string` parameter** (matches
    `cobra.RunE` shape) — the spec's pseudocode in "Notes for the
    Implementer" used `runAddFlags(cmd, args)` for the dispatcher
    call but originally hinted at `runAddFlags(cmd *cobra.Command,
    args []string)` only on the call site; both shapes match the
    implementation, so this is a non-deviation but worth noting.
  - **Tutorial §8 wrapper note** updated to drop the "Until
    editor-launch ships in STAGE-002" qualifier (the wrapper is
    still useful for 10-second flag-mode capture, but the
    "until ..." framing is now wrong). Disclosed per stage-level
    drive-by guidance.
- **Follow-up work identified:**
  - None new. SPEC-011 (FTS5) and SPEC-012 (`brag search`) remain
    as planned.

### Build-phase reflection (3 questions, short answers)

1. **What was unclear in the spec that slowed you down?**
   — The "all existing SPEC-003/SPEC-005 `TestAdd_*` tests pass
   unchanged" acceptance criterion conflicted with locked design
   decision #1 (no-flags → editor mode); the conflicting test
   (`TestAdd_MissingTitleIsUserError`) wasn't called out as
   needing removal/update. I had to discover the conflict by
   running the suite, watching `vi` get spawned in the test
   process list, and reading the spec carefully to confirm the
   numbered design decision should win.

2. **Was there a constraint or decision that should have been listed
   but wasn't?**
   — No new constraint. But the spec's §"Locked design decisions"
   item 4 ("Flag mode is unchanged") would have benefited from an
   explicit carve-out: "existing tests that *exercise editor mode
   under the new semantics* (i.e. `add` with no field flags) need
   updating". The pattern would help future specs that flip
   default-route semantics.

3. **If you did this task again, what would you do differently?**
   — Skim every existing test in the touched file BEFORE running
   fail-first, looking specifically for tests whose *premise*
   (not just assertion) is invalidated by the new behavior. The
   `TestAdd_MissingTitleIsUserError` conflict was visible in
   `add_test.go` line ~85 the moment I read it; I noticed only
   after the test hung in CI-style execution. A two-minute pre-run
   "premise audit" of the existing suite would have caught it
   immediately.

---

## Reflection (Ship)

*Appended 2026-04-21 during the **ship** cycle. Outcome-focused,
distinct from the process-focused build reflection above.*

1. **What would I do differently next time?**
   When a locked design decision inverts or removes existing
   behavior, explicitly enumerate the existing tests whose
   premise is invalidated by the change. SPEC-010's decision #1
   (no-flags → editor mode) inverted
   `TestAdd_MissingTitleIsUserError`'s entire premise — that test
   previously asserted "no-args is a user error" because Title
   was required; after this spec, no-args opens the editor
   instead. Build correctly deleted the test, but the deletion
   was a build-time *discovery*, not a planned spec action. A
   two-minute "premise audit" during design — walk each locked
   decision against existing tests in the affected files — would
   have made the deletion a planned output listed under Outputs
   alongside the files modified.

2. **Does any template, constraint, or decision need updating?**
   Yes — extend the AGENTS.md §9 "locked-decisions-need-tests"
   rule (earned in SPEC-009 ship) with its inverse: when a locked
   decision inverts or removes existing behavior, enumerate the
   tests whose premise becomes invalid so their deletion is a
   planned action listed in spec Outputs, not a build-time
   discovery disclosed under Deviations. Both halves of the rule
   (new behavior ↔ new test; removed behavior ↔ removed test)
   make design-to-test traceability symmetric. Applied in this
   ship commit.

3. **Is there a follow-up spec to write now before I forget?**
   No. SPEC-011 (FTS5 virtual table + triggers, M) is the last
   M-sized spec in STAGE-002 and the next pending item. SPEC-012
   (`brag search`, S) follows it as the final STAGE-002 spec.
   After SPEC-012 ships, STAGE-002 closes and the project moves
   to STAGE-003 (export + summary) framing — or a dogfooding
   pause, user's call then.
