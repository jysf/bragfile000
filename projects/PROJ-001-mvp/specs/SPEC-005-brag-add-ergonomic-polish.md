---
task:
  id: SPEC-005
  type: chore
  cycle: build
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
  created_at: 2026-04-20

references:
  decisions:
    - DEC-006  # cobra framework — StringP is the shorthand API
    - DEC-007  # RunE-validated required flags, not MarkFlagRequired
  constraints:
    - stdout-is-for-data-stderr-is-for-humans
    - errors-wrap-with-context
    - test-before-implementation
    - one-spec-per-pr
  related_specs:
    - SPEC-001  # shipped; root command with --db persistent flag
    - SPEC-003  # shipped; add command with all long-form flags
---

# SPEC-005: `brag add` ergonomic polish (shorthands + help)

## Context

First spec in STAGE-002. Dogfooding SPEC-003 immediately surfaced two
friction points: every `brag add` invocation requires typing six
long-form flag names (no single-letter shorthands), and new users
don't always discover that `brag <command> --help` (not `brag --help`)
shows per-subcommand flags. Both are pure ergonomics — nothing
architectural — but they gate the 10-second-capture goal in PROJ-001's
success criteria.

This spec is purely additive: existing long-form flags keep working
byte-identically. No storage changes, no schema changes, no DEC
changes. Small enough to ship in a single tight build session; first
spec in STAGE-002 and explicitly meant to validate the stage's open.

Parent stage: `STAGE-002-capture-and-retrieval.md`. Project: PROJ-001.

## Goal

Add single-letter shorthands to every `brag add` flag, refresh the
`add` command's `Long` description with concrete usage examples, and
add one line to the root command's `Long` pointing users at
`brag <command> --help` for per-subcommand flag detail.

## Inputs

- **Files to read:**
  - `internal/cli/add.go` — current flag declarations (all
    `.String(name, default, usage)` — no shorthands).
  - `internal/cli/root.go` — current root `Long` description.
  - `internal/cli/add_test.go` — existing `newRootWithAdd` test
    harness; new tests reuse it.
  - `internal/cli/errors.go` — `ErrUser` sentinel + `UserErrorf` (no
    change, but referenced by the empty-title shorthand test).
  - `AGENTS.md` §9 (separate `outBuf`/`errBuf`, tie-break,
    fail-first test run).
  - `/decisions/DEC-006-cobra-cli-framework.md` — StringP is the
    cobra API for shorthand + long-form pairs.
  - `/decisions/DEC-007-required-flag-validation-in-runE.md` —
    confirms: do NOT reach for `MarkFlagRequired`. The existing
    `TrimSpace` check in `runAdd` is the only way `--title`/`-t`
    empty values become `ErrUser`.
- **External APIs:** none.
- **Related code paths:** `internal/cli/`.

## Outputs

- **Files modified:**
  - `internal/cli/add.go` — each of the six `.String()` calls
    becomes `.StringP()` with a shorthand letter; `Long` gains a
    short Examples block.
  - `internal/cli/root.go` — `Long` gains one line pointing users
    at `brag <command> --help`.
  - `internal/cli/add_test.go` — add new tests (see Failing Tests).
  - `internal/cli/root_test.go` — one new test asserting the Long
    description mentions subcommand help.
  - `docs/tutorial.md` — update the full-metadata example to show
    the shorthand form alongside the long form.
- **Files created:** none.
- **New exports:** none (signatures unchanged).
- **Database changes:** none.

## Shorthand assignments

Mapping locked here so verify / build sessions don't re-litigate:

| Long | Short | Rationale |
|---|---|---|
| `--title` | `-t` | First letter; most common flag. |
| `--description` | `-d` | First letter; unique. |
| `--tags` | `-T` | `-t` is taken by title; capital T follows the convention (e.g. `docker build -t` vs some tools using `-T` for plural). |
| `--project` | `-p` | First letter; unique. |
| `--type` | `-k` | `-t` taken; `-k` reads as "kind" and avoids collision with `-T` (tags). Mnemonic: `k`ind. |
| `--impact` | `-i` | First letter; unique. |

Cobra's built-in `-h` (help) and root `-v` (version) are the only
pre-existing short flags; no collision with any of the above.

## Acceptance Criteria

- [ ] `brag add -t "x"` inserts an entry identically to
      `brag add --title "x"` (same resulting row, same stdout ID,
      same zero stderr). *[TestAdd_ShorthandTitleEquivalentToLong]*
- [ ] Each of `-d`, `-T`, `-p`, `-k`, `-i` persists its value to the
      corresponding column. One test per shorthand field, using
      `Store.List` on the same DB to read back and assert.
      *[TestAdd_Shorthand{Description,Tags,Project,Type,Impact}]*
- [ ] Mixed form works: `brag add -t "x" --tags "a,b" -p "platform"`
      persists all three fields correctly.
      *[TestAdd_ShorthandAndLongFormMix]*
- [ ] `brag add -t ""` returns an `ErrUser` identical to `brag add
      --title ""` (empty-title validation applies to both forms;
      main.go maps to exit code 1).
      *[TestAdd_EmptyShorthandTitleIsUserError]*
- [ ] `brag add --help` stdout contains each shorthand in cobra's
      standard `-X, --long` format (e.g. `-t, --title`). Help goes
      to stdout with empty stderr.
      *[TestAdd_HelpShowsShorthands]*
- [ ] `brag add --help` stdout contains at least one example line
      (the refreshed `Long` description with Examples).
      *[TestAdd_HelpShowsExamples]*
- [ ] `brag --help` stdout mentions `brag <command> --help` (exact
      phrasing not pinned — assert the substring `<command> --help`
      or `brag <cmd> --help`).
      *[TestRoot_HelpPointsAtSubcommandHelp]*
- [ ] All SPEC-001/003/004 tests remain green. No existing test is
      modified; new tests are add-only.
      *[manual: go test ./...]*
- [ ] `gofmt -l .` empty, `go vet ./...` clean, `CGO_ENABLED=0 go
      build ./...` succeeds, `go test ./...` green.

## Failing Tests

Written now. All tests reuse the existing `newRootWithAdd` helper in
`internal/cli/add_test.go`. Every test uses separate `outBuf` /
`errBuf` per AGENTS.md §9 with no-cross-leakage asserts on the empty
side.

### New tests in `internal/cli/add_test.go`

- **`TestAdd_ShorthandTitleEquivalentToLong`** — run `brag add -t
  "hello"` against a temp DB. Assert: exit nil, stdout is a single
  numeric ID + newline, stderr empty. Then open the same DB via
  `storage.Open` and list entries; assert the single row has
  `Title == "hello"`.

- **`TestAdd_ShorthandDescription`** — `brag add -t "x" -d "body"`;
  after listing, assert returned entry's `Description == "body"`.

- **`TestAdd_ShorthandTags`** — `brag add -t "x" -T "auth,perf"`;
  assert entry's `Tags == "auth,perf"`.

- **`TestAdd_ShorthandProject`** — `brag add -t "x" -p "platform"`;
  assert entry's `Project == "platform"`.

- **`TestAdd_ShorthandType`** — `brag add -t "x" -k "shipped"`;
  assert entry's `Type == "shipped"`.

- **`TestAdd_ShorthandImpact`** — `brag add -t "x" -i "mobile v3"`;
  assert entry's `Impact == "mobile v3"`.

- **`TestAdd_ShorthandAndLongFormMix`** — `brag add -t "x" --tags
  "a,b" -p "proj"`. Assert all three fields persisted.

- **`TestAdd_EmptyShorthandTitleIsUserError`** — run `brag add -t ""`.
  Assert: `err != nil`, `errors.Is(err, ErrUser) == true`, `outBuf`
  empty (no ID printed). Mirrors the existing long-form empty-title
  test exactly; verify parity.

- **`TestAdd_HelpShowsShorthands`** — run `brag add --help`. Assert:
  err nil, `errBuf.Len() == 0`, `outBuf.String()` contains each of
  `"-t, --title"`, `"-d, --description"`, `"-T, --tags"`, `"-p,
  --project"`, `"-k, --type"`, `"-i, --impact"`. (Cobra's `--help`
  renders the `StringP` pair in that comma-separated form.)

- **`TestAdd_HelpShowsExamples`** — run `brag add --help`. Assert
  `outBuf.String()` contains `"Examples:"` or at least one literal
  `"brag add"` invocation in the help body (the refreshed `Long`).

### New test in `internal/cli/root_test.go`

- **`TestRoot_HelpPointsAtSubcommandHelp`** — build the root with
  `--help` args (no subcommand attached for the test — just the
  root's `Long` matters). Assert `outBuf.String()` contains `"--help"`
  and one of `"<command>"`, `"<cmd>"`, or `"brag <"`. (Phrasing not
  pinned to exact wording; substring assert keeps the test
  implementable across small rewordings.)

Notes for the implementer on testing patterns:

- Run `go test ./...` once after writing the new tests and BEFORE any
  implementation change — confirm the new tests fail for the
  expected reason (unknown shorthand flag, or missing substring in
  help text). That's the SPEC-003 Q3 / SPEC-004 habit now codified
  in AGENTS.md §9.
- Do NOT modify any existing test — purely additive.
- Reuse `newRootWithAdd(t)` for add tests. The root-help test
  doesn't need it; build a minimal `NewRootCmd("test")` directly.

## Implementation Context

*Read before starting build. Self-contained handoff.*

### Decisions that apply

- `DEC-006` — Cobra is the framework. Use `cmd.Flags().StringP(long,
  short, default, usage)` to add shorthand + long-form pairs in a
  single call. This is the canonical cobra API — don't invent
  anything around it.
- `DEC-007` — Required-flag validation lives in `RunE`, not via
  `MarkFlagRequired`. SPEC-005 does NOT change that. The existing
  `strings.TrimSpace(title) == ""` check in `runAdd` handles both
  `--title ""` and `-t ""` because cobra maps both to the same
  string flag; no code path change needed.

### Constraints that apply

For `internal/cli/**` and `docs/tutorial.md`:

- `stdout-is-for-data-stderr-is-for-humans` — blocking. Every new
  test asserts `errBuf.Len() == 0` on the happy-path side.
- `errors-wrap-with-context` — warning. No new error paths in this
  spec; existing wrapping in `runAdd` stays unchanged.
- `test-before-implementation` — blocking. Write failing tests
  first, run `go test ./...`, confirm they fail for the right reason
  (undefined shorthand / missing substring), THEN edit `add.go` and
  `root.go`.
- `one-spec-per-pr` — blocking. Branch
  `feat/spec-005-brag-add-ergonomic-polish`. Diff should touch only
  `internal/cli/add.go`, `internal/cli/add_test.go`,
  `internal/cli/root.go`, `internal/cli/root_test.go`, and
  `docs/tutorial.md`.

### AGENTS.md lessons that apply

- §9 separate `outBuf`/`errBuf` in every CLI test (SPEC-001).
- §9 fail-first test run before implementation (SPEC-003 ship
  lesson, validated in SPEC-004).
- §10 `/`-anchored gitignore — not touched by this spec; the rule
  stands.

### Prior related work

- **SPEC-001** (shipped). Root command `Long` is currently one
  sentence; SPEC-005 appends one line.
- **SPEC-003** (shipped). `internal/cli/add.go` has all six flags
  declared via `.String()`; SPEC-005 converts each to `.StringP()`
  with a shorthand letter. Existing `runAdd` logic is unchanged —
  cobra binds both forms to the same flag target.

### Out of scope (for this spec specifically)

If any of these feels necessary during build, create a new spec.

- **Shorthand flags on other subcommands** (`list`, future
  `show`/`edit`/`delete`/`search`). `list` has no mandatory flags
  today and filter flags arrive in SPEC-006 with their own
  shorthand decisions. Other subcommands haven't been written yet.
  SPEC-005's scope is strictly the `add` surface.
- **Shell completion scripts** (`brag completion zsh` etc.). Cobra
  supports this for free but enabling it is post-MVP polish.
- **Global-flag shorthand on `--db`** (e.g. `-D`). Deliberately
  skipped — `--db` is rarely typed in daily use (default DB path
  covers 99% of cases); adding a shorthand just burns a letter for
  no real gain.
- **Removing the long-form flags.** NEVER. Backwards compatibility
  forever; both forms stay.
- **Auto-completion in ~/.zshrc.** Out of scope.

## Notes for the Implementer

- **`StringP` signature.** `cmd.Flags().StringP("title", "t", "",
  "short headline (required)")`. Positional args are:
  name, shorthand, default, usage.
- **Do not touch `runAdd`.** Cobra maps `-t` and `--title` to the
  same underlying flag target; `getFlagString(cmd, "title")` keeps
  working. No other Go code changes.
- **Existing comment about `MarkFlagRequired`.** Leave it. It
  still applies and explains DEC-007 to future readers.
- **Add command `Long` refresh.** Keep the current first sentence;
  add an Examples block. Suggested shape (verify session will sign
  off on the wording):
  ```go
  Long: `Add a new brag entry. Title is required; other fields are optional.

  Examples:
    brag add -t "shipped the auth refactor"
    brag add -t "cut p99 latency" -T "auth,perf" -p "platform" \
             -i "unblocked mobile v3 release"
    brag add --title "..." --description "..." --tags "..." \
             --project "..." --type "..." --impact "..."

  Short forms: -t title, -d description, -T tags, -p project,
  -k type, -i impact.`,
  ```
- **Root `Long` refresh.** Append one line to the existing
  description:
  ```go
  Long: `Bragfile — a local-first CLI for engineers to capture and retrieve career accomplishments for retros, reviews, and resumes.

  Run 'brag <command> --help' for command-specific flags and usage.`,
  ```
- **Tutorial update.** In `docs/tutorial.md` §3 (Capture with full
  metadata), add a second example block showing the shorthand form.
  Keep the long-form block above it — the doc teaches both. Also
  add a tiny note in §2 that `-t` is available.
- **No new `init()` functions** (AGENTS.md §8). Still applies.
- **Cobra shorthand help rendering.** `StringP("title", "t", ...)`
  makes `--help` render the flag as `-t, --title string`. The
  `TestAdd_HelpShowsShorthands` test relies on this exact format.
  If cobra's output ever drifts (unlikely), the test fails loudly —
  that's acceptable.

---

## Build Completion

*Filled in at the end of the **build** cycle, before advancing to verify.*

- **Branch:**
- **PR (if applicable):**
- **All acceptance criteria met?** yes/no
- **New decisions emitted:**
  - (none expected; pure ergonomic polish)
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
