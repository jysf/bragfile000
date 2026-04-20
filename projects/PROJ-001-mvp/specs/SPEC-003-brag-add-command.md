---
task:
  id: SPEC-003
  type: story
  cycle: build
  blocked: false
  priority: high
  complexity: S

project:
  id: PROJ-001
  stage: STAGE-001
repo:
  id: bragfile

agents:
  architect: claude-opus-4-7
  implementer: claude-opus-4-7
  created_at: 2026-04-20

references:
  decisions:
    - DEC-003  # config resolution order
    - DEC-004  # tags as comma-joined TEXT (user passes verbatim)
    - DEC-006  # cobra as the CLI framework
  constraints:
    - no-cgo
    - no-sql-in-cli-layer
    - stdout-is-for-data-stderr-is-for-humans
    - errors-wrap-with-context
    - test-before-implementation
    - one-spec-per-pr
  related_specs:
    - SPEC-001  # shipped; root command + config resolver
    - SPEC-002  # shipped; storage layer
---

# SPEC-003: `brag add` command

## Context

SPEC-002 shipped the storage layer. The binary compiles but has no
subcommands — users can't actually capture a brag. This spec ships
the first real user-visible command: `brag add`. When this spec
ships:

- `brag add --title "..."` writes one row to `~/.bragfile/db.sqlite`
  (or the `--db`-specified path) and prints the inserted ID to
  stdout, nothing else.
- `brag add` with no args or empty `--title` exits 1 with an error
  message on stderr. Storage failures exit 2.
- All optional fields from the api-contract (`--description`,
  `--tags`, `--project`, `--type`, `--impact`) are accepted and
  persisted.
- The output is pipe-friendly: `id=$(brag add --title "...")` works.
- A new error sentinel + error-formatting pattern gets established
  that SPEC-004 and every later subcommand will inherit.

Parent stage: `STAGE-001-foundations.md`. Project: PROJ-001 (MVP).
Both SPEC-001 and SPEC-002 ship cleanly as of 2026-04-20.

## Goal

Ship `internal/cli/add.go` + a small error-handling helper in
`internal/cli/errors.go`, wire the subcommand into
`cmd/brag/main.go`, and update `main.go` to format errors as
`brag: <message>` on stderr with exit codes 1 (user error) / 2
(internal error). After this spec, `brag add --title "x"` works
end-to-end and prints an integer ID.

## Inputs

- **Files to read:**
  - `docs/architecture.md` — `internal/cli` row of the
    Responsibilities table; the Data Flow "happy path" section
    describes exactly this command.
  - `docs/api-contract.md` — `brag add` section (flags, required
    vs optional, exit codes, stdout shape, error formatting rules).
  - `AGENTS.md` §8 (coding conventions), §9 (testing conventions —
    especially the separate `outBuf`/`errBuf` rule from SPEC-001 and
    the monotonic tie-break rule from SPEC-002).
  - `/decisions/DEC-003-config-resolution-order.md`
  - `/decisions/DEC-004-tags-comma-joined-for-mvp.md`
  - `/decisions/DEC-006-cobra-cli-framework.md`
  - `/guidance/constraints.yaml`.
  - `projects/PROJ-001-mvp/specs/done/SPEC-001-go-module-and-cobra-scaffold.md`
    — shape of `NewRootCmd` and existing test pattern.
  - `projects/PROJ-001-mvp/specs/done/SPEC-002-sqlite-storage-and-migrations.md`
    — `Store.Add` signature and returned `Entry` shape.
- **External APIs:** none.
- **Related code paths:**
  - `internal/cli/root.go` (exists) — read it; the subcommand will
    rely on the `--db` persistent flag defined there.
  - `internal/config/config.go` (exists) — `ResolveDBPath` is used
    to expand the `--db` value into an absolute path.
  - `internal/storage` (exists) — `storage.Open`, `storage.Entry`,
    `(*Store).Add`, `(*Store).Close`.
  - `cmd/brag/main.go` (exists) — gets the subcommand registration
    and the error-formatting/exit-code mapping.

## Outputs

- **Files created:**
  - `internal/cli/add.go` — exports
    `NewAddCmd() *cobra.Command`. Constructs the subcommand with
    flags, `MarkFlagRequired("title")`, and a `RunE` that opens the
    store, builds the `Entry`, inserts it, prints the ID, and
    closes.
  - `internal/cli/errors.go` — exports a sentinel
    `var ErrUser = errors.New("user error")` used to flag
    user-input errors so `main.go` can map them to exit 1. Also
    exports a small helper `UserErrorf(format string, args ...any)
    error` returning `fmt.Errorf("%w: "+format, append([]any{ErrUser},
    args...)...)` for ergonomic wrapping at the call site.
  - `internal/cli/add_test.go` — tests per the "Failing Tests"
    section below. Uses separate `outBuf` / `errBuf` (AGENTS.md §9).
  - `internal/cli/errors_test.go` — one or two tiny tests confirming
    `ErrUser` is detectable via `errors.Is`.
- **Files modified:**
  - `internal/cli/root.go` — add `SilenceErrors = true` and
    `SilenceUsage = true` on the returned `*cobra.Command` so
    main.go owns error formatting and usage-on-failure doesn't
    dump on every user error. Verify existing SPEC-001 tests still
    pass.
  - `cmd/brag/main.go` — register the subcommand
    (`root.AddCommand(cli.NewAddCmd())`) and replace the current
    naive `os.Exit(1)` error path with a helper that:
      - prints the error to stderr with prefix `brag: `
      - maps `errors.Is(err, cli.ErrUser)` → exit code 1
      - maps anything else → exit code 2
      - nil → exit code 0
- **New exports:**
  - `cli.NewAddCmd() *cobra.Command`
  - `cli.ErrUser` (sentinel error)
  - `cli.UserErrorf(format string, args ...any) error`
- **Database changes:** none. This spec only consumes SPEC-002's
  existing schema.

## Acceptance Criteria

- [ ] `brag add --title "first brag"` on a fresh temp DB inserts one
      row and writes exactly the inserted ID (+ trailing newline) to
      stdout. Stderr is empty. Exit 0.
      *[TestAdd_SuccessPrintsIDToStdoutOnly]*
- [ ] `id=$(./brag --db /tmp/x.db add --title "pipeable")` captures
      an integer ID cleanly (no prefix, no whitespace beyond the
      trailing newline the shell strips).
      *[TestAdd_OutputIsPipeable]*
- [ ] `brag add` with no `--title` flag returns a user error
      (detectable via `errors.Is(err, cli.ErrUser)`). Intended exit
      code: 1. *[TestAdd_MissingTitleIsUserError]*
- [ ] `brag add --title ""` (explicit empty string) returns a user
      error. `MarkFlagRequired` alone does not catch this; the
      command must validate explicitly.
      *[TestAdd_EmptyTitleIsUserError]*
- [ ] `brag add --title "   "` (whitespace-only title) returns a
      user error. Implementation uses `strings.TrimSpace` to
      validate. *[TestAdd_WhitespaceTitleIsUserError]*
- [ ] `brag add --title "x" --description "why" --tags "a,b"
      --project "p" --type "t" --impact "i"` persists every optional
      field verbatim. A follow-up `storage.Store.List(...)` call
      returns an `Entry` with exactly those string values.
      *[TestAdd_AllOptionalFieldsPersisted]*
- [ ] Two successive `brag add --title "same"` calls produce rows
      with distinct IDs; both appear in a subsequent `List(...)`.
      *[TestAdd_TwoAddsDistinctIDs]*
- [ ] `brag add --help` lists all six flags (`--title`,
      `--description`, `--tags`, `--project`, `--type`, `--impact`)
      plus inherited `--db`/`--version`/`-h,--help` from root. Help
      goes to stdout; stderr is empty.
      *[TestAdd_HelpListsAllFlags]*
- [ ] `cli.ErrUser` is detectable via `errors.Is` when a user error
      is returned. `cli.UserErrorf("bad flag %q", name)` produces an
      error that both (a) matches `errors.Is(err, ErrUser)` and (b)
      contains the formatted message.
      *[TestErrUser_IsDetectable, TestUserErrorf_FormatsAndWraps]*
- [ ] After this spec, SPEC-001's three root-command tests
      (`TestRootCmd_VersionFlag`, `TestRootCmd_HelpFlag`,
      `TestRootCmd_NoArgs`) still pass — `SilenceErrors` /
      `SilenceUsage` do not regress them. *[existing tests green]*
- [ ] `gofmt -l .` empty, `go vet ./...` clean, `go test ./...`
      green, `go build ./cmd/brag` succeeds, `CGO_ENABLED=0 go build
      ./cmd/brag` succeeds.

## Failing Tests

All CLI tests use separate `outBuf` and `errBuf` (AGENTS.md §9).
All tests that touch storage use `t.TempDir()` for the DB path. Use
`t.Helper()` in shared helpers.

### `internal/cli/add_test.go`

Imports: `testing`, `bytes`, `errors`, `path/filepath`, `strconv`,
`strings`, the package under test, and `internal/storage` for
verification reads.

Shared helper (inside the test file, unexported):

- `func newRootWithAdd(t *testing.T) (*cobra.Command, string)` —
  builds a fresh root (`NewRootCmd("test")`), attaches the add
  subcommand (`root.AddCommand(NewAddCmd())`), returns the root and
  a `t.TempDir()`-backed DB path. Caller is responsible for setting
  args, out, err.

Tests:

- **`TestAdd_SuccessPrintsIDToStdoutOnly`** — arrange root+add, set
  args `[]string{"--db", dbPath, "add", "--title", "first brag"}`,
  attach separate `outBuf` and `errBuf`. Execute. Assert:
  - `err == nil`
  - `errBuf.Len() == 0`
  - `strings.TrimSpace(outBuf.String())` parses as a positive
    integer.
  - A `storage.Open(dbPath)` + `List(ListFilter{})` returns one
    entry with `Title == "first brag"`.
- **`TestAdd_OutputIsPipeable`** — same setup. Assert the entire
  `outBuf.String()` equals `<id>\n` (no prefix, no extra
  whitespace). Regex `^\d+\n$` is fine.
- **`TestAdd_MissingTitleIsUserError`** — args `[]string{"--db",
  dbPath, "add"}`, no `--title` flag. Assert:
  - `err != nil`
  - `errors.Is(err, ErrUser)` is true.
- **`TestAdd_EmptyTitleIsUserError`** — args `[]string{"--db",
  dbPath, "add", "--title", ""}`. Assert same two properties as
  above.
- **`TestAdd_WhitespaceTitleIsUserError`** — args `[]string{"--db",
  dbPath, "add", "--title", "   "}`. Same assertions.
- **`TestAdd_AllOptionalFieldsPersisted`** — args include every
  optional flag. Execute. Open the DB via `storage.Open`, `List`,
  assert the single returned entry has every field set to the
  passed value.
- **`TestAdd_TwoAddsDistinctIDs`** — execute add twice against the
  same DB with the same title. Assert `outBuf` contains two
  distinct integer IDs on two lines (parse & compare). `List`
  returns two entries.
- **`TestAdd_HelpListsAllFlags`** — args `[]string{"add", "--help"}`
  (no `--db` needed; help doesn't touch storage). Assert:
  - `err == nil`
  - `errBuf.Len() == 0`
  - `outBuf.String()` contains every one of the six flags.
  - `outBuf.String()` contains `"--db"` (inherited persistent flag).

### `internal/cli/errors_test.go`

- **`TestErrUser_IsDetectable`** — `err := ErrUser`; assert
  `errors.Is(err, ErrUser)` is true.
- **`TestUserErrorf_FormatsAndWraps`** —
  `err := UserErrorf("bad %q", "x")`. Assert:
  - `errors.Is(err, ErrUser)` is true.
  - `err.Error()` contains `"bad \"x\""`.

Notes for the implementer on testing patterns:

- Reuse the `newRootWithAdd` helper across tests to keep each test
  body small. Mark it with `t.Helper()` so failures point at the
  caller.
- Don't test exit-code-mapping behavior here — that lives in
  `main.go`, which is not covered by tests in this spec (there is
  no `main_test.go` in Go stdlib convention for a `main` package
  entrypoint). The contract is: add returns `ErrUser`-wrapped errors
  for bad input, other errors for storage failures. Mapping is
  `main.go`'s job; tests cover the "command returns correct error
  shape" half.
- If a test opens `storage.Open` on the same path the command
  already wrote to, close both with `defer s.Close()` to avoid
  dangling file handles on Windows (not currently a target, but
  harmless on macOS/Linux too).

## Implementation Context

*Read before starting build. Self-contained handoff.*

### Decisions that apply

- `DEC-003` — `--db` flag → `BRAGFILE_DB` env → default
  `~/.bragfile/db.sqlite`. `NewAddCmd`'s `RunE` reads the
  inherited `--db` flag via `cmd.Flags().GetString("db")` and
  passes it to `config.ResolveDBPath`. Do NOT re-implement the
  resolution.
- `DEC-004` — Tags are comma-joined `TEXT`. This spec stores
  whatever string the user passes to `--tags` verbatim. No
  splitting, no trimming, no validation. If the user passes
  `" a , b "` we persist `" a , b "`. Normalization is a future
  concern (revisit in STAGE-002 once `list --tag` needs it).
- `DEC-006` — Cobra. Use `cmd.Flags().String/StringP(...)`,
  `cmd.MarkFlagRequired("title")`, and cobra's
  `cmd.OutOrStdout() / cmd.ErrOrStderr()` for output (not
  `os.Stdout` / `os.Stderr` directly) so tests can redirect.

### Constraints that apply

For `internal/cli/**` and `cmd/brag/**`:

- `no-cgo` — blocking. No new CGO deps.
- `no-sql-in-cli-layer` — blocking. `internal/cli/add.go` must not
  import `database/sql` or any SQL driver. All persistence goes
  through `internal/storage`.
- `stdout-is-for-data-stderr-is-for-humans` — blocking. The ID goes
  to stdout. The `brag: ` error prefix goes to stderr. Tests assert
  separate buffers (AGENTS.md §9).
- `errors-wrap-with-context` — warning. Wrap storage errors:
  `fmt.Errorf("add entry: %w", err)`. User-input errors use
  `cli.UserErrorf(...)`.
- `test-before-implementation` — blocking. Write failing tests
  first.
- `one-spec-per-pr` — blocking. Branch
  `feat/spec-003-brag-add-command`.

### Prior related work

- **SPEC-001** (shipped on 2026-04-20, archived) — PR #1, merged at
  `ff4a446`; ship commit `3883a42` added the AGENTS.md §9
  separate-buffer rule. This spec is the first to actually use it
  on a subcommand.
- **SPEC-002** (shipped on 2026-04-20, archived) — PR #2, merged at
  `02dcd0e`; ship commit `b5f7ca8` added the AGENTS.md §9
  tie-break rule (does not apply here — add doesn't do ordering).
  `Store.Add(Entry)` returns `(Entry, error)` with the inserted
  `Entry.ID` populated. Deviations in build noted the `id DESC`
  tie-break and second-truncated timestamps — both are storage-
  layer concerns; this spec just calls `Add` and reads `.ID`.

### Out of scope (for this spec specifically)

If any of these feels necessary during build, write a new spec.

- **`brag add` with no args → opens `$EDITOR`.** STAGE-002. This
  spec is flags-only.
- **Tag normalization / splitting / validation.** Store whatever
  the user passes. Normalization lands in STAGE-002 alongside
  `list --tag`.
- **Duplicate-title detection or `--dedupe` flag.** Explicitly out
  per STAGE-001 success criteria: "Running the command twice with
  the same title creates two distinct entries (no implicit
  dedup)."
- **Confirmation prompt, colored output, fancy formatting.** Plain
  integer + newline to stdout. Nothing else.
- **`brag list` or any other subcommand.** SPEC-004 handles list;
  `show` / `edit` / `delete` / `search` belong to STAGE-002.
- **Shell completion for the new flags.** Cobra gives it for free
  later (`brag completion zsh`) but that command isn't added in
  PROJ-001.
- **Exit codes other than 0/1/2.** api-contract says three codes;
  don't invent a fourth.

## Notes for the Implementer

- **Subcommand constructor shape.** Match SPEC-001's pattern:
  ```go
  func NewAddCmd() *cobra.Command {
      cmd := &cobra.Command{
          Use:   "add",
          Short: "Add a new brag entry",
          Long:  "Add a new brag entry. Title is required; other fields are optional.",
          RunE:  runAdd,
      }
      cmd.Flags().String("title", "", "short headline (required)")
      cmd.Flags().String("description", "", "free-form body")
      cmd.Flags().String("tags", "", "comma-joined tag list (e.g. \"auth,perf\")")
      cmd.Flags().String("project", "", "project / initiative this brag belongs to")
      cmd.Flags().String("type", "", "free-form category (shipped, learned, mentored, ...)")
      cmd.Flags().String("impact", "", "impact statement (metric, quote, outcome)")
      _ = cmd.MarkFlagRequired("title")
      return cmd
  }
  ```
- **`runAdd` body sketch.**
  ```go
  func runAdd(cmd *cobra.Command, args []string) error {
      title, _ := cmd.Flags().GetString("title")
      if strings.TrimSpace(title) == "" {
          return UserErrorf("--title is required and must not be empty")
      }
      dbFlag, _ := cmd.Flags().GetString("db")       // inherited persistent flag
      path, err := config.ResolveDBPath(dbFlag)
      if err != nil {
          return fmt.Errorf("resolve db path: %w", err)
      }
      s, err := storage.Open(path)
      if err != nil {
          return fmt.Errorf("open store: %w", err)
      }
      defer s.Close()

      entry := storage.Entry{
          Title:       title,
          Description: getFlagString(cmd, "description"),
          Tags:        getFlagString(cmd, "tags"),
          Project:     getFlagString(cmd, "project"),
          Type:        getFlagString(cmd, "type"),
          Impact:      getFlagString(cmd, "impact"),
      }
      inserted, err := s.Add(entry)
      if err != nil {
          return fmt.Errorf("add entry: %w", err)
      }
      fmt.Fprintln(cmd.OutOrStdout(), inserted.ID)
      return nil
  }

  func getFlagString(cmd *cobra.Command, name string) string {
      v, _ := cmd.Flags().GetString(name)
      return v
  }
  ```
- **Error helper (`errors.go`).**
  ```go
  package cli

  import (
      "errors"
      "fmt"
  )

  var ErrUser = errors.New("user error")

  func UserErrorf(format string, args ...any) error {
      return fmt.Errorf("%w: "+format, append([]any{ErrUser}, args...)...)
  }
  ```
- **`main.go` error mapping.** Replace the current
  `if err := root.Execute(); err != nil { os.Exit(1) }` with:
  ```go
  if err := root.Execute(); err != nil {
      fmt.Fprintf(os.Stderr, "brag: %s\n", err.Error())
      if errors.Is(err, cli.ErrUser) {
          os.Exit(1)
      }
      os.Exit(2)
  }
  ```
  Import `"errors"` and the cli package. Note: `cmd.SilenceErrors =
  true` and `cmd.SilenceUsage = true` on the root prevent cobra's
  default "Error: ..." line and usage dump so main's formatting is
  the only stderr output on failure.
- **`SilenceErrors` / `SilenceUsage`.** Add both to the
  `*cobra.Command` returned by `NewRootCmd` in `internal/cli/root.go`.
  Re-run SPEC-001's tests to confirm no regression — the existing
  tests assert on stdout only so this shouldn't break them, but
  check.
- **Don't panic, don't `log.Fatal`.** Every error returns via
  `RunE`. `main.go` is the only place `os.Exit` lives.
- **The `add --help` test.** Cobra exits its `Help()` via stdout,
  not via the `RunE`. When `--help` is passed, cobra short-circuits
  and writes to the command's `OutOrStdout`. The test should assert
  the written content contains `"--title"`, `"--description"`,
  `"--tags"`, `"--project"`, `"--type"`, `"--impact"`, and `"--db"`.

---

## Build Completion

*Filled in at the end of the **build** cycle, before advancing to verify.*

- **Branch:**
- **PR (if applicable):**
- **All acceptance criteria met?** yes/no
- **New decisions emitted:**
  - (none expected)
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
