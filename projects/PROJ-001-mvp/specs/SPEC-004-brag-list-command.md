---
task:
  id: SPEC-004
  type: story
  cycle: verify
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
    - DEC-005  # integer autoincrement IDs (we print them)
    - DEC-006  # cobra framework
    - DEC-007  # required-flag validation in RunE (pattern carries forward)
  constraints:
    - no-sql-in-cli-layer
    - stdout-is-for-data-stderr-is-for-humans
    - errors-wrap-with-context
    - test-before-implementation
    - one-spec-per-pr
  related_specs:
    - SPEC-001  # shipped; root command + --db flag
    - SPEC-002  # shipped; Store.List(ListFilter{}) stub
    - SPEC-003  # shipped; add command, ErrUser pattern, main.go exit mapping
---

# SPEC-004: `brag list` command

## Context

Last spec in STAGE-001. SPEC-003 made `brag add` work end-to-end;
SPEC-004 closes the loop by giving the user a way to read what was
written. When this spec ships, the minimum capture + retrieve loop
(`brag add --title "..."` → `brag list`) works, which is STAGE-001's
stated outcome. No filter flags in this spec — the plumbing for
filtering (`ListFilter` struct) exists from SPEC-002; flag wiring is
deferred to STAGE-002 per the stage backlog.

Parent stage: `STAGE-001-foundations.md`. Project: PROJ-001 (MVP).

## Goal

Ship `brag list` as a cobra subcommand that opens the store, calls
`Store.List(ListFilter{})`, and prints every entry to stdout on its
own line in the form `<id>\t<created_at>\t<title>` in reverse-
chronological order — matching the API contract in
`docs/api-contract.md` and applying every lesson accumulated in
AGENTS.md §9 and §10 through SPEC-003.

## Inputs

- **Files to read:**
  - `docs/api-contract.md` — `brag list` section (format, order, no
    filter flags in STAGE-001).
  - `docs/architecture.md` — `internal/cli` row of Responsibilities;
    Data Flow section's "`list` is the mirror image of `add`".
  - `AGENTS.md` §8 (coding conventions), §9 (testing conventions —
    separate buffers, time-based ordering tie-break), §10 (git/PR
    conventions including `/`-anchored gitignore).
  - `/decisions/DEC-003-config-resolution-order.md`
  - `/decisions/DEC-005-integer-autoincrement-ids.md`
  - `/decisions/DEC-006-cobra-cli-framework.md`
  - `/decisions/DEC-007-required-flag-validation-in-runE.md`
  - `/guidance/constraints.yaml`
  - `internal/cli/add.go` + `internal/cli/add_test.go` (from SPEC-003)
    — the reference shape for a subcommand with store access.
  - `internal/cli/errors.go` (from SPEC-003) — `ErrUser` sentinel +
    `UserErrorf` helper; not used by `list` but don't duplicate.
  - `internal/storage/store.go` + `entry.go` (from SPEC-002) —
    `Store.List`, `Entry`, `ListFilter{}`.
  - `cmd/brag/main.go` (from SPEC-003) — existing command
    registration and exit-code mapping.
- **External APIs:** none.
- **Related code paths:** `internal/cli/`, `cmd/brag/main.go`.

## Outputs

- **Files created:**
  - `internal/cli/list.go` — defines `NewListCmd() *cobra.Command`
    plus unexported `runList` RunE handler.
  - `internal/cli/list_test.go` — tests for the command's output
    format, ordering, empty case, and error propagation.
- **Files modified:**
  - `cmd/brag/main.go` — one added line:
    `root.AddCommand(cli.NewListCmd())` alongside the existing
    `root.AddCommand(cli.NewAddCmd())`. Nothing else.
- **New exports:**
  - `cli.NewListCmd() *cobra.Command`
- **Database changes:** none.

## Acceptance Criteria

- [ ] `brag list` on a fresh DB (no entries) prints nothing to stdout,
      nothing to stderr, and exits 0.
      *[TestListCmd_EmptyPrintsNothing]*
- [ ] `brag list` after three `Store.Add` calls prints three lines to
      stdout, one per entry, in reverse-chronological order (most
      recent first). *[TestListCmd_PrintsReverseChronological]*
- [ ] Each printed line matches the format
      `<id>\t<created_at>\t<title>` — exactly two tab characters per
      line, `created_at` in RFC3339 UTC, `id` as plain decimal
      integer. *[TestListCmd_TabSeparatedFormat]*
- [ ] When ordering is otherwise ambiguous (same `created_at` second),
      the tie-break from `Store.List` (id DESC) propagates through
      unchanged — SPEC-004 adds nothing to the ordering logic.
      *[TestListCmd_TieBreakIsIDDescending]*
- [ ] `brag list --db <path>` honors the flag: running against a
      different DB file returns that DB's entries, not the default.
      *[TestListCmd_RespectsDBFlag]*
- [ ] A `Store.Open` failure (e.g., `--db` pointing at an unreadable
      path or a directory) surfaces as an error from `RunE`. The
      returned error is **not** an `ErrUser` (so `main.go` maps it to
      exit code 2, internal error). *[TestListCmd_StorageOpenErrorIsInternal]*
- [ ] `brag list --help` prints help that includes the command
      description. No required flags advertised.
      *[TestListCmd_HelpShape]*
- [ ] Existing tests from SPEC-001/002/003 remain green. Integration
      of `list` does not regress any prior test.
      *[manual: go test ./...]*
- [ ] `gofmt -l .` empty, `go vet ./...` clean, `CGO_ENABLED=0 go
      build ./...` succeeds, `go test ./...` green.

## Failing Tests

Written now. All tests use the lessons accumulated in AGENTS.md §9:
separate `outBuf` / `errBuf` buffers with no-cross-leakage asserts on
the empty-stderr side; time-based ordering tests don't rely on sleeps
alone (the tie-break from `Store.List` already handles second-
precision collisions).

### `internal/cli/list_test.go`

Imports: `testing`, `bytes`, `path/filepath`, `strings`, `errors`,
`github.com/jysf/bragfile000/internal/cli`,
`github.com/jysf/bragfile000/internal/storage`, `github.com/spf13/cobra`.

Shared helper (unexported, mirrors the SPEC-003 add_test.go pattern —
reuse the same helper name if `add_test.go` already has one; otherwise
define locally):

- `func newListTestRoot(t *testing.T, dbPath string) (*cobra.Command,
  *bytes.Buffer, *bytes.Buffer)` — builds a fresh root command with
  the `--db` persistent flag set to `dbPath`, registers
  `cli.NewListCmd()` as a subcommand, wires separate `outBuf` and
  `errBuf` via `root.SetOut(outBuf)` / `root.SetErr(errBuf)`, returns
  all three. Does NOT `.Execute()` — the test drives args.

Tests:

- **`TestListCmd_EmptyPrintsNothing`** — `Store.Open` a fresh
  `t.TempDir()` DB and immediately close it. Build the root+list
  command against that path, `cmd.SetArgs([]string{"list"})`, execute.
  Assert: no error, `outBuf.Len() == 0`, `errBuf.Len() == 0`.

- **`TestListCmd_PrintsReverseChronological`** — open a store, `Add`
  three entries with titles `"first"`, `"second"`, `"third"` in that
  insertion order. Close. Run `list`. Assert: `errBuf.Len() == 0`;
  splitting `outBuf.String()` on `"\n"` yields (after trimming a
  trailing empty string) three lines in order `["third", "second",
  "first"]` — check by splitting each line on `"\t"` and comparing
  the 3rd field.

- **`TestListCmd_TabSeparatedFormat`** — add one entry. Run `list`.
  Take the single non-empty output line. Assert: `strings.Count(line,
  "\t") == 2` (exactly two tabs). Split on `"\t"`; assert the
  fields are, in order: a decimal string parseable via
  `strconv.ParseInt`, a string parseable via
  `time.Parse(time.RFC3339, …)` whose `Location().String() == "UTC"`,
  and the entry's title.

- **`TestListCmd_TieBreakIsIDDescending`** — `Store.Add` two entries
  *within the same second* (call `Add` twice in rapid succession with
  no sleep). The second insert should get a higher ID via
  AUTOINCREMENT. Run `list`. Assert the higher ID appears first,
  even though both rows share a `created_at` second. (This test
  guards against any future accidental `ORDER BY` change in the CLI
  layer — `Store.List` already does the right thing; we just verify
  the CLI doesn't rewrap.)

- **`TestListCmd_RespectsDBFlag`** — open two different temp DBs.
  Add entry `"in-A"` to DB A and `"in-B"` to DB B. Close both.
  Invoke `list` with `--db` pointing at A. Assert only `"in-A"`
  appears in stdout, not `"in-B"`.

- **`TestListCmd_StorageOpenErrorIsInternal`** — invoke `list --db`
  pointing at a path that is a directory, not a file (use
  `t.TempDir()` itself as the `--db` value — opening it as a SQLite
  DB will fail). Assert: `err != nil`, `!errors.Is(err, cli.ErrUser)`,
  `outBuf.Len() == 0`.

- **`TestListCmd_HelpShape`** — build root+list with buffers,
  `cmd.SetArgs([]string{"list", "--help"})`, execute. Assert: nil
  error; `outBuf` contains `"Usage:"` and a one-line description
  string matching `cmd.Short`; `errBuf.Len() == 0` (help goes to
  stdout per SPEC-001's pattern).

Notes for the implementer on testing patterns:

- Follow SPEC-003's `add_test.go` exactly for buffer setup and root-
  with-subcommand harness. If `add_test.go` already exports or
  defines a helper that constructs a root-with-subcommands for
  testing, reuse it rather than duplicating. (If none exists, the
  `newListTestRoot` helper above belongs in `list_test.go`.)
- `t.TempDir()` for every DB path.
- Use `t.Setenv("BRAGFILE_DB", "")` at the top of any test that
  cares about env precedence, so host env doesn't leak in.

## Implementation Context

*Read before starting build. Self-contained handoff.*

### Decisions that apply

- `DEC-003` — Config resolution order is flag → env → default. `list`
  reads the `--db` persistent flag off the command; the resolution
  logic lives in `internal/config.ResolveDBPath` from SPEC-001. Do
  not duplicate.
- `DEC-005` — IDs are `INTEGER AUTOINCREMENT`, printed as plain
  decimal integers. No zero-padding, no prefix, no ULID.
- `DEC-006` — Cobra is the framework; one subcommand per file under
  `internal/cli/` following the `add.go` shape from SPEC-003.
- `DEC-007` — Required-flag validation in `RunE`, not
  `MarkFlagRequired`. SPEC-004 has **no required flags**, so this
  decision doesn't alter any code path here, but **do not reach for
  `MarkFlagRequired` if you find yourself adding a required flag to
  `list` during build**. Flag any such need in /guidance/questions.yaml
  and stop.

### Constraints that apply

For `internal/cli/**` and `cmd/brag/**`:

- `no-sql-in-cli-layer` — blocking. `list.go` must not import
  `database/sql` or any SQL driver. All persistence goes through
  `storage.Store`.
- `stdout-is-for-data-stderr-is-for-humans` — blocking. The listed
  entries go to stdout (`cmd.OutOrStdout()`). Any human-readable
  prefix (e.g., for future errors) would go to stderr — but this
  command has no stderr output on success. Enforce via the test
  pattern: every happy-path test asserts `errBuf.Len() == 0`.
- `errors-wrap-with-context` — warning. Wrap every returned error:
  `fmt.Errorf("resolve db path: %w", err)`, `fmt.Errorf("open
  store: %w", err)`, `fmt.Errorf("list entries: %w", err)`.
- `test-before-implementation` — blocking. Write failing tests
  first; run them and confirm they fail for the expected reason
  (per SPEC-003 Q3 ship reflection — don't skip this step).
- `one-spec-per-pr` — blocking. Branch
  `feat/spec-004-brag-list-command`. The diff should only touch the
  files listed in Outputs. No drive-bys unless they're load-bearing
  AND disclosed under Deviations.

### AGENTS.md lessons that apply

- §9: **Separate `outBuf` and `errBuf`** in every test with a no-
  cross-leakage assert on the empty side. (From SPEC-001.)
- §9: **Monotonic tie-break for time-based ordering** is already in
  `Store.List`; SPEC-004 should not rewrap that. The
  `TestListCmd_TieBreakIsIDDescending` test guards against
  accidental re-sort in the CLI. (From SPEC-002.)
- §10: **`/`-anchor gitignore for binary names.** SPEC-003 already
  fixed this by narrowing `brag` → `/brag`. SPEC-004 should not add
  any new gitignore entries, but if it does, anchor them. (From
  SPEC-003.)

### Prior related work

- **SPEC-001** (shipped 2026-04-20, archived). Root command +
  `internal/config`. `list` inherits the persistent `--db` flag from
  the root; no new flag plumbing.
- **SPEC-002** (shipped 2026-04-20, archived). `Store.List(ListFilter
  {}) ([]Entry, error)` returns entries in `created_at DESC, id DESC`
  order. `ListFilter` is a named empty struct.
- **SPEC-003** (shipped 2026-04-20, archived). `internal/cli/add.go`
  establishes the per-command shape: flag declarations, `RunE`
  resolves `--db`, opens store, performs the operation, prints one
  thing to stdout. `internal/cli/errors.go` defines `ErrUser` +
  `UserErrorf`. `cmd/brag/main.go` registers `NewAddCmd()` and maps
  errors via `errors.Is(err, cli.ErrUser)` to exit code 1, everything
  else to 2.

### Out of scope (for this spec specifically)

If any of these feels necessary during build, write a new spec.

- **Filter flags** (`--tag`, `--project`, `--type`, `--since`,
  `--limit`). STAGE-002. The stage backlog pins this: "no filter
  flags yet". Do NOT extend `ListFilter` or add any `cmd.Flags().
  String("tag", …)` call.
- **Alternate output formats** (JSON, CSV, styled). STAGE-002 or
  STAGE-003. MVP output is the tab-separated form in
  `api-contract.md`.
- **Paging / `--limit`.** STAGE-002. Unlimited default, suffices for
  MVP at personal scale.
- **Coloring / terminal detection.** Out of MVP scope entirely.
- **`Store.Get(id)`, `Store.Delete(id)`, `Store.Update(...)`** —
  STAGE-002 subcommands (`show`, `edit`, `delete`) own those.
- **FTS5 / `brag search`** — STAGE-002.
- **Escaping titles containing tabs or newlines** — accept the
  trade-off for MVP. Titles are typed by the user; tabs/newlines in
  titles are user-generated pathologies, not MVP bugs. A follow-up
  question belongs in `questions.yaml` if it comes up in real use.

## Notes for the Implementer

- **Command shape.** Mirror `add.go` exactly: `NewListCmd()` builds
  the `*cobra.Command` with `Use: "list"`, `Short` and `Long`
  describing the command, `RunE: runList`. No flags declared on the
  subcommand itself (the persistent `--db` comes from root). Return
  the command.
- **RunE body.**
  ```go
  func runList(cmd *cobra.Command, _ []string) error {
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

      entries, err := s.List(storage.ListFilter{})
      if err != nil {
          return fmt.Errorf("list entries: %w", err)
      }

      out := cmd.OutOrStdout()
      for _, e := range entries {
          fmt.Fprintf(out, "%d\t%s\t%s\n",
              e.ID,
              e.CreatedAt.UTC().Format(time.RFC3339),
              e.Title)
      }
      return nil
  }
  ```
  Reuse `getFlagString` from `add.go`.
- **Printing on empty.** The `for` loop is a no-op if `entries` is
  empty. Nothing prints. No "no entries" message anywhere. Empty is
  a valid successful state.
- **Time formatting.** `Entry.CreatedAt` is already `time.Time` in
  UTC (DEC-002 + `timestamps-in-utc-rfc3339` constraint). Calling
  `.UTC()` again is idempotent and defensive against any future
  round-trip that might introduce a non-UTC zone on read.
- **Registration.** In `cmd/brag/main.go`, add exactly one line next
  to the existing `root.AddCommand(cli.NewAddCmd())`:
  `root.AddCommand(cli.NewListCmd())`. Do not touch anything else in
  `main.go`.
- **Error propagation.** Do not introduce new error types. Storage
  errors come back wrapped and propagate; `main.go`'s
  `errors.Is(err, cli.ErrUser)` check will correctly NOT match,
  landing on exit code 2. This is what the spec's
  `TestListCmd_StorageOpenErrorIsInternal` asserts.
- **No SQL here.** `list.go` imports only `fmt`, `time`,
  `github.com/spf13/cobra`, `github.com/jysf/bragfile000/internal/
  config`, `github.com/jysf/bragfile000/internal/storage`.
- **No `init()` functions.**
- **If you need a required flag**, STOP and ask (DEC-007). SPEC-004
  does not anticipate any; this note is defensive.

---

## Build Completion

*Filled in at the end of the **build** cycle, before advancing to verify.*

- **Branch:** `feat/spec-004-brag-list-command`
- **PR (if applicable):** (opened after `just advance-cycle`)
- **All acceptance criteria met?** yes
  - Confirmed: `go test ./...` green (7 new `TestListCmd_*` tests + all
    prior tests still pass), `gofmt -l .` empty, `go vet ./...` clean,
    `CGO_ENABLED=0 go build ./...` succeeds, smoke-test via built
    binary produces the exact tab-separated, reverse-chronological,
    id-DESC-tie-break output the api-contract specifies.
- **New decisions emitted:**
  - None. Implementation is pure reuse of DEC-003 (config resolution),
    DEC-005 (integer IDs as plain decimals), DEC-006 (cobra), and the
    `add.go` shape. DEC-007 did not need to fire — `list` has no
    required flags.
- **Deviations from spec:**
  - `newListTestRoot` helper receives `dbPath` but does not consume it;
    tests pass `dbPath` into `root.SetArgs([]string{"--db", dbPath,
    "list", …})`. Kept the spec's signature for consistency; the
    parameter reads as documentation of "the DB this root is going to
    talk to" without binding the helper to any particular arg shape.
    Explicit `_ = dbPath` marks it intentional.
  - `TestListCmd_PrintsReverseChronological` does NOT sleep between
    `Add` calls. Three rapid inserts share a created_at second; the
    `id DESC` tie-break in `Store.List` already produces the required
    `third, second, first` order. Adding sleeps would have made the
    test slow (~3s) for no invariant gain — and the whole point of the
    SPEC-002 lesson is that sleeps alone aren't the mechanism. The
    ordering invariant is covered without them here; the explicit
    tie-break case lives in `TestListCmd_TieBreakIsIDDescending`.
- **Follow-up work identified:**
  - None new. Stage backlog items (filter flags, alternate output
    formats, paging) are already captured in STAGE-002 scope and in
    this spec's Out-of-scope section.

### Build-phase reflection (3 questions, short answers)

1. **What was unclear in the spec that slowed you down?**
   — The `newListTestRoot` helper signature includes a `dbPath`
   parameter with no described use inside the body ("test drives
   args"). I spent a few minutes deciding whether the helper should
   stash it, `SetArgs` with it, or leave it as a hint to callers. The
   spec's intent became clear once I accepted that it's just a
   documentation hook — but a one-liner clarification ("the helper
   doesn't consume dbPath; the caller uses it when calling SetArgs")
   would have saved the deliberation.

2. **Was there a constraint or decision that should have been listed but wasn't?**
   — No. The spec's `## Implementation Context` was
   self-contained: DEC-003/005/006/007, `no-sql-in-cli-layer`,
   `stdout-is-for-data-stderr-is-for-humans`, `errors-wrap-with-
   context`, `test-before-implementation`, and `one-spec-per-pr` were
   the exact levers this implementation needed. The `add.go` code
   sample in "Notes for the Implementer" was load-bearing — it
   removed the last ambiguity about `RunE` body shape.

3. **If you did this task again, what would you do differently?**
   — Write the tests without the between-Add sleeps from the
   beginning rather than adding and then removing them. The SPEC-002
   lesson is *exactly* "sleeps don't establish order under RFC3339
   second precision; monotonic tie-break does", and `Store.List`
   already provides the tie-break. Three rapid inserts with the
   `id DESC` tie-break IS the correct way to assert reverse-chron
   order — faster, stronger invariant, no flake window.

---

## Reflection (Ship)

*Appended during the **ship** cycle.*

1. **What would I do differently next time?**
   — <answer>

2. **Does any template, constraint, or decision need updating?**
   — <answer>

3. **Is there a follow-up spec I should write now before I forget?**
   — <answer>
