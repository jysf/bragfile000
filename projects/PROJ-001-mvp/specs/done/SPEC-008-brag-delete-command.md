---
task:
  id: SPEC-008
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
  created_at: 2026-04-20

references:
  decisions:
    - DEC-005  # integer autoincrement IDs (parsed from positional arg)
    - DEC-006  # cobra framework
    - DEC-007  # RunE-validated flags AND positional args (+SPEC-006 extension)
  constraints:
    - no-sql-in-cli-layer
    - storage-tests-use-tempdir
    - stdout-is-for-data-stderr-is-for-humans
    - errors-wrap-with-context
    - test-before-implementation
    - one-spec-per-pr
  related_specs:
    - SPEC-002  # shipped; Store, Entry, scan patterns
    - SPEC-003  # shipped; add command, ErrUser, main.go exit mapping
    - SPEC-005  # shipped; shorthand flag convention, help-assertion specificity
    - SPEC-006  # shipped; show command, Store.Get, storage.ErrNotFound,
                # DEC-007 positional-arg extension
---

# SPEC-008: `brag delete <id>` command

## Context

Fourth spec in STAGE-002. `brag add`, `list`, and `show` are shipped;
the user can capture, scan, and drill in, but there's no way to remove
a typo-entry short of `sqlite3 ~/.bragfile/db.sqlite "DELETE FROM
entries WHERE id = N"`. SPEC-008 closes the last CRUD gap: a thin
cobra subcommand + one new `Store.Delete(id)` method, with a
confirmation prompt by default and a `--yes`/`-y` bypass for scripting.

Reuses every pattern SPEC-006 established for ID-based commands
(positional-arg validation via DEC-007 extension, `storage.ErrNotFound`
sentinel mapped to `UserErrorf`, `Long` with Examples). Adds one new
shape: interactive stdin confirmation.

Parent stage: `STAGE-002-capture-and-retrieval.md`. Project: PROJ-001.

## Goal

Ship `brag delete <id>` as a cobra subcommand that hard-deletes an
entry after user confirmation (or immediately with `--yes`/`-y`).
Invalid IDs, missing IDs, and malformed args all exit 1 via `ErrUser`.
Declining the confirmation prompt exits **0** (a user choice, not a
user error).

## Inputs

- **Files to read:**
  - `docs/api-contract.md` — "`brag delete <id>`" section. **Note:**
    this spec UPDATES the doc's exit-code clause — declining the
    confirmation prompt is reclassified from exit 1 to exit 0 (see
    Design rationale below). The update lands in this spec's PR.
  - `docs/data-model.md` — Data Lifecycle section confirms hard
    delete is the only mode; no soft-delete column.
  - `AGENTS.md` §8 (conventions), §9 (testing: separate buffers,
    fail-first, assertion specificity), §12 "During design".
  - `/decisions/DEC-005-integer-autoincrement-ids.md`
  - `/decisions/DEC-006-cobra-cli-framework.md`
  - `/decisions/DEC-007-required-flag-validation-in-runE.md` — the
    SPEC-006 extension on positional-arg validation applies
    identically here.
  - `internal/cli/show.go` — reference shape for an ID-taking cobra
    subcommand.
  - `internal/cli/show_test.go` — reference shape for the
    positional-arg ErrUser test battery.
  - `internal/cli/errors.go` — `ErrUser` + `UserErrorf`.
  - `internal/storage/store.go` — existing `Store.Get` to fetch the
    entry for the confirmation message.
  - `internal/storage/errors.go` — `ErrNotFound`; `Store.Delete`
    returns it for missing IDs, same pattern as `Store.Get`.
  - `internal/storage/storagetest/` — test helper sub-package; not
    needed for this spec's tests but a good reminder of the layering.
- **External APIs:** none.
- **Related code paths:** `internal/cli/`, `internal/storage/`,
  `cmd/brag/main.go`.

## Outputs

- **Files created:**
  - `internal/cli/delete.go` — `NewDeleteCmd() *cobra.Command` +
    unexported `runDelete` RunE handler.
  - `internal/cli/delete_test.go` — arg-parsing, confirmation,
    happy path, stream-discipline, help-shape tests.
- **Files modified:**
  - `internal/storage/store.go` — add `Store.Delete(id int64)
    error`. Returns `ErrNotFound` (wrapped) when no rows are
    affected.
  - `internal/storage/store_test.go` — add two tests (happy path,
    not-found) following the `TestDelete_*` naming convention.
  - `cmd/brag/main.go` — register the subcommand with one added
    line: `root.AddCommand(cli.NewDeleteCmd())`.
  - `docs/api-contract.md` — under "`brag delete <id>`": change the
    exit-code clause from "Exit 1 if the ID does not exist or the
    user declines" to "Exit 1 if the ID does not exist or the arg
    is invalid; decline at the confirmation prompt exits 0
    (cleanly aborted)."
  - `docs/tutorial.md` — strike `brag delete <id>` from §9
    "What's NOT there yet"; add a short example in §4 or a new
    section near `show`.
- **New exports:**
  - `(*storage.Store).Delete(id int64) error`
  - `cli.NewDeleteCmd() *cobra.Command`
- **Database changes:** none. The existing `entries` schema
  supports deletion without schema changes.

## Design decisions inline (no new DEC needed)

The following choices are locked in the spec so build / verify don't
re-litigate. Each was evaluated; none felt weighty enough to warrant
a standalone DEC file.

1. **Decline → exit 0, not exit 1.** User said "n" at a confirmation
   prompt. That's a deliberate choice, not a user error. `rm -i`,
   `kubectl delete`, `git rebase --abort` all exit 0 on user-aborted
   operations. `main.go`'s `ErrUser` → exit 1 formatting ("brag:
   user error: ...") would misclassify a deliberate abort as an
   error message. `docs/api-contract.md` gets amended in this spec's
   PR to match.

2. **All feedback to stderr; stdout stays empty.** This command has
   no structured data to return on success — the caller knows they
   asked to delete, they don't need a "Deleted." string to parse.
   Prompt, "Deleted.", and "Aborted." all go to stderr
   (`cmd.ErrOrStderr()`), honoring
   `stdout-is-for-data-stderr-is-for-humans` strictly.

3. **Confirmation prompt format.**
   ```
   Delete entry 42 ("shipped auth refactor")? [y/N]
   ```
   Shows ID + quoted title so the user can double-check which entry
   they're about to delete. `[y/N]` (capital N) indicates "No is
   the default" — enter, empty input, any non-`y`/`Y` response
   declines. Only lowercase `y` or uppercase `Y` confirms.

4. **`-y` shorthand for `--yes`.** Convention (`apt -y`, `rm -f`).
   No flag-letter collision — delete has only this one flag and
   `-y` is unique in its subcommand scope.

5. **Hard delete only.** Per DEC-004-adjacent decisions and data-
   model.md Data Lifecycle: no soft-delete column, no trash bin,
   no undo. If we ever want recoverability, it's a follow-up spec.
   SQLite file is the backup.

6. **No dry-run flag (e.g., `--dry-run`).** Out of scope — the
   confirmation prompt already serves the "am I sure?" function.
   `--yes` is the only modifier.

7. **Fetch-then-delete.** `runDelete` fetches the entry via
   `Store.Get` first (for the confirmation-prompt title), then
   deletes. This means a disappearing-between-fetch-and-delete row
   is theoretically possible (single-user local DB, so unlikely).
   Documented here; `Store.Delete`'s `ErrNotFound` path surfaces
   cleanly if it happens.

## Acceptance Criteria

- [ ] `brag delete <id>` on an existing entry prompts
      `Delete entry <id> ("<title>")? [y/N] ` to stderr, reads from
      stdin, and:
      - on `y` or `Y` + newline: deletes the row, prints `Deleted.`
        to stderr, exits 0. *[TestDeleteCmd_ConfirmY]*
      - on `n`, empty line, or any other input: leaves the row in
        place, prints `Aborted.` to stderr, exits 0.
        *[TestDeleteCmd_DeclineN, TestDeleteCmd_DeclineEmpty,
        TestDeleteCmd_DeclineOther]*
- [ ] `brag delete <id> --yes` and `brag delete <id> -y` skip the
      prompt entirely, delete the row, print `Deleted.` to stderr,
      exit 0. Stdin is not read. *[TestDeleteCmd_YesFlagLongAndShort]*
- [ ] `brag delete 999` where id 999 does not exist returns an error
      such that `errors.Is(err, cli.ErrUser)` matches (exit 1).
      Tested both with and without `--yes`.
      *[TestDeleteCmd_NotFoundIsUserError,
      TestDeleteCmd_NotFoundWithYesIsUserError]*
- [ ] `brag delete` with no positional arg returns `ErrUser`.
      *[TestDeleteCmd_NoArgsIsUserError]*
- [ ] `brag delete 1 2` returns `ErrUser` (too many args).
      *[TestDeleteCmd_TooManyArgsIsUserError]*
- [ ] `brag delete abc` returns `ErrUser` (non-numeric).
      *[TestDeleteCmd_NonNumericArgIsUserError]*
- [ ] `brag delete 0` returns `ErrUser` (non-positive).
      *[TestDeleteCmd_NonPositiveArgIsUserError]*
- [ ] On every happy/decline/error path, `outBuf.Len() == 0`
      (stdout remains empty). Asserted in every test.
- [ ] `brag delete --help` shows usage with the `--yes, -y` flag
      and an `Examples:` block.
      *[TestDeleteCmd_HelpShape]*
- [ ] `Store.Delete(id)` removes the row from `entries`. After
      deletion, `Store.Get(id)` returns `storage.ErrNotFound`.
      *[TestDelete_RemovesRow]*
- [ ] `Store.Delete(id)` on a missing ID returns
      `storage.ErrNotFound` wrapped. *[TestDelete_NotFoundReturnsErrNotFound]*
- [ ] Existing tests (SPEC-001..007) remain green; no regressions.
      *[manual: go test ./...]*
- [ ] `gofmt -l .` empty, `go vet ./...` clean, `CGO_ENABLED=0 go
      build ./...` succeeds, `go test ./...` green.
- [ ] `docs/api-contract.md` updated for the exit-code amendment.
- [ ] `docs/tutorial.md` updated: `delete <id>` struck from "What's
      NOT there yet" and demonstrated with at least one example.

## Failing Tests

Written now. Every CLI test uses separate `outBuf` / `errBuf` and
asserts on both (§9). Every test that prompts for confirmation also
sets `cmd.SetIn(strings.NewReader(...))` to feed stdin
deterministically. Every help/output assertion uses distinctive
tokens, not cobra boilerplate (§9 SPEC-005 lesson). Fail-first run
before implementation (§9 SPEC-003 lesson).

### New tests in `internal/storage/store_test.go`

Reuse the `newTestStore` helper. Follow existing `TestAdd_*` /
`TestGet_*` style.

- **`TestDelete_RemovesRow`** — `Add` an entry, call `Delete(id)`,
  assert no error; call `Get(id)`, assert `errors.Is(err,
  storage.ErrNotFound)`.
- **`TestDelete_NotFoundReturnsErrNotFound`** — fresh store.
  `Delete(999)`. Assert `errors.Is(err, storage.ErrNotFound)`.

### New tests in `internal/cli/delete_test.go`

Use a new `newRootWithDelete(t)` helper modeled on `newRootWithShow`
in `show_test.go` — builds root + delete subcommand, returns root +
temp DB path. Tests use separate `outBuf` / `errBuf` via
`root.SetOut` / `root.SetErr`, and feed stdin via `root.SetIn` when
the test exercises the confirmation path.

- **`TestDeleteCmd_ConfirmY`** — seed one entry. SetIn `"y\n"`.
  `SetArgs([]string{"delete", "1"})`. Execute. Assert:
  - err nil
  - `outBuf.Len() == 0` (strict stdout discipline)
  - `errBuf` contains the distinctive literal `"Delete entry 1"`
    AND the quoted title substring AND `"Deleted."`
  - Open a second `*sql.DB` via `storagetest` (or a small new
    helper) and confirm the row is gone. Or simply re-`Open` the
    store via `storage.Open` and call `Store.Get(1)`, asserting
    `errors.Is(err, storage.ErrNotFound)`.

- **`TestDeleteCmd_DeclineN`** — SetIn `"n\n"`. Execute. Assert:
  err nil, `outBuf.Len() == 0`, `errBuf` contains `"Aborted."`,
  `Store.Get(1)` still returns the entry (row not deleted).

- **`TestDeleteCmd_DeclineEmpty`** — SetIn `"\n"` (just newline).
  Assert: same shape as DeclineN — aborted, row present.

- **`TestDeleteCmd_DeclineOther`** — SetIn `"maybe\n"` (any non-
  `y`/`Y` response). Assert: aborted, row present.

- **`TestDeleteCmd_YesFlagLongAndShort`** — run twice (two entries).
  Once with `--yes`, once with `-y`. No SetIn needed. Both should
  delete without prompt and print `"Deleted."` to stderr. Stdout
  empty for both.

- **`TestDeleteCmd_NotFoundIsUserError`** — fresh store.
  `SetArgs([]string{"delete", "999"})`. No `--yes`. Execute.
  Assert: `errors.Is(err, ErrUser)`, `outBuf.Len() == 0`, no prompt
  reached stderr (the implementation should fail at `Store.Get`
  before prompting).

- **`TestDeleteCmd_NotFoundWithYesIsUserError`** — same as above
  but with `--yes`. Assert: `errors.Is(err, ErrUser)`, stdout empty.

- **`TestDeleteCmd_NoArgsIsUserError`** — `delete` alone. `ErrUser`.

- **`TestDeleteCmd_TooManyArgsIsUserError`** — `delete 1 2`. `ErrUser`.

- **`TestDeleteCmd_NonNumericArgIsUserError`** — `delete abc`.
  `ErrUser`.

- **`TestDeleteCmd_NonPositiveArgIsUserError`** — `delete 0`.
  `ErrUser`. (Skip `delete -5` — cobra parses it as a flag.)

- **`TestDeleteCmd_HelpShape`** — `delete --help`. Assert nil error,
  `errBuf.Len() == 0`, `outBuf` contains `"Examples:"` (distinctive
  label) AND `"-y, --yes"` (cobra's rendered shorthand form for
  `StringVarP`-like bindings; actually for `BoolVarP` the form is
  the same — `-y, --yes` — so the substring assert holds).

Notes for the implementer on testing patterns:

- Fail-first: run `go test ./...` once after writing these tests,
  before any implementation. Confirm each fails for the expected
  reason (undefined `NewDeleteCmd`, `Store.Delete`, or missing
  substring). If any unexpectedly passes, tighten the assertion.
- Use `cmd.SetIn(strings.NewReader(...))` to feed stdin. Reading
  via `bufio.NewReader(cmd.InOrStdin()).ReadString('\n')` (or
  `fmt.Fscanln`) works with the test-injected reader.

## Implementation Context

*Read before starting build. Self-contained handoff.*

### Decisions that apply

- `DEC-005` — IDs are int64, parsed from `strconv.ParseInt(arg, 10,
  64)`, validated `> 0`.
- `DEC-006` — Cobra. `NewDeleteCmd` mirrors `NewShowCmd` (SPEC-006)
  and `NewAddCmd` (SPEC-003). Use `BoolVarP` for `--yes`/`-y`.
- `DEC-007` — Required-flag + positional-arg validation lives in
  `RunE`. No `cobra.ExactArgs`. Parse and validate in `runDelete`
  directly; return `UserErrorf(...)` on every invalid-arg path.

### Constraints that apply

For `internal/cli/**`, `internal/storage/**`, `cmd/brag/**`, `docs/**`:

- `no-sql-in-cli-layer` — blocking. `delete.go` imports only
  `config` + `storage` (+ stdlib). Any SQL stays in `store.go`.
- `storage-tests-use-tempdir` — blocking.
- `stdout-is-for-data-stderr-is-for-humans` — blocking. Strictest
  application in this spec: stdout stays empty on every path.
  Every test asserts `outBuf.Len() == 0`.
- `errors-wrap-with-context` — warning. Wrap storage errors:
  `fmt.Errorf("get entry: %w", err)`,
  `fmt.Errorf("delete entry: %w", err)`.
  `UserErrorf` returns are their own thing.
- `test-before-implementation` — blocking.
- `one-spec-per-pr` — blocking. Branch
  `feat/spec-008-brag-delete-command`.

### AGENTS.md lessons that apply

- §9 separate `outBuf`/`errBuf` in every CLI test (SPEC-001).
- §9 monotonic tie-break — not directly relevant (no list ordering
  here).
- §9 fail-first test run before implementation (SPEC-003).
- §9 assertion specificity (SPEC-005) — help test asserts
  `"Examples:"`, confirmation-prompt test asserts on the quoted
  title substring, not generic labels.
- §12 "During design" (SPEC-007) — every implementation choice in
  this spec's Notes passes `no-sql-in-cli-layer` and every other
  blocking constraint. No "either is acceptable" language
  anywhere. Confirmed.

### Prior related work

- **SPEC-002** (shipped). Store/Entry/scan patterns; storage
  `newTestStore` helper; error-wrapping style.
- **SPEC-003** (shipped). `add.go` establishes the cobra-command
  shape; `main.go` handles `ErrUser` → exit 1 mapping.
- **SPEC-005** (shipped). Shorthand flag convention (`-y` for
  `--yes` here). Help-assertion specificity (`"Examples:"`
  substring).
- **SPEC-006** (shipped). `show.go` is the canonical reference for
  an ID-taking command: positional-arg validation, `Store.Get`
  integration, `ErrNotFound` → `UserErrorf` mapping. SPEC-008
  mirrors all of it; the only new surface is the interactive
  confirmation prompt.
- **SPEC-007** (shipped). `internal/storage/storagetest` sub-package
  exists if a test needs to inspect DB state independently of
  `Store` — not strictly needed here (`Store.Get` suffices for
  post-delete verification).

### Out of scope (for this spec specifically)

If any of these feels necessary during build, write a new spec.

- **Soft delete** (`deleted_at` column, `--restore` command).
  Deliberately not in MVP. Data Lifecycle doc is clear: hard delete.
- **Bulk delete** (`brag delete 1 2 3` or `brag delete --tag auth`).
  Single-ID only. Filter-based bulk delete is a future polish spec.
- **`--dry-run`** (preview without deleting). The confirmation
  prompt already serves this purpose.
- **Trash bin / undo.** Out of MVP scope.
- **`--force` flag** (synonym or alternate name for `--yes`).
  `--yes` and `-y` only; conventional enough.
- **Re-confirmation for `--yes`** on high-risk deletions
  (e.g., "are you sure?" twice). Single confirmation is enough.
- **Any UI beyond stdin/stderr.** No TUI, no interactive menu.

## Notes for the Implementer

- **Command shape.** Mirror `internal/cli/show.go` closely:
  ```go
  func NewDeleteCmd() *cobra.Command {
      cmd := &cobra.Command{
          Use:   "delete <id>",
          Short: "Delete a brag entry",
          Long: `Delete a brag entry. Prompts for confirmation unless --yes is passed.

  Examples:
    brag delete 42              # prompts y/N on stdin
    brag delete 42 --yes        # skip prompt
    brag delete 42 -y           # same via shorthand`,
          RunE: runDelete,
      }
      cmd.Flags().BoolP("yes", "y", false, "skip confirmation prompt")
      return cmd
  }
  ```

- **`runDelete` structure.**
  ```go
  func runDelete(cmd *cobra.Command, args []string) error {
      if len(args) != 1 {
          return UserErrorf("delete requires exactly one <id> argument")
      }
      id, err := strconv.ParseInt(args[0], 10, 64)
      if err != nil {
          return UserErrorf("invalid id %q: must be a positive integer", args[0])
      }
      if id <= 0 {
          return UserErrorf("invalid id %d: must be positive", id)
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

      entry, err := s.Get(id)
      if err != nil {
          if errors.Is(err, storage.ErrNotFound) {
              return UserErrorf("no entry with id %d", id)
          }
          return fmt.Errorf("get entry: %w", err)
      }

      yes, _ := cmd.Flags().GetBool("yes")
      if !yes {
          fmt.Fprintf(cmd.ErrOrStderr(),
              "Delete entry %d (%q)? [y/N] ", id, entry.Title)
          reader := bufio.NewReader(cmd.InOrStdin())
          line, _ := reader.ReadString('\n')
          line = strings.TrimRight(line, "\r\n")
          if line != "y" && line != "Y" {
              fmt.Fprintln(cmd.ErrOrStderr(), "Aborted.")
              return nil  // exit 0 — deliberate user choice
          }
      }

      if err := s.Delete(id); err != nil {
          // Unlikely path given the Get above, but handle defensively.
          if errors.Is(err, storage.ErrNotFound) {
              return UserErrorf("no entry with id %d", id)
          }
          return fmt.Errorf("delete entry: %w", err)
      }
      fmt.Fprintln(cmd.ErrOrStderr(), "Deleted.")
      return nil
  }
  ```

- **`Store.Delete` implementation.**
  ```go
  func (s *Store) Delete(id int64) error {
      res, err := s.db.Exec(`DELETE FROM entries WHERE id = ?`, id)
      if err != nil {
          return fmt.Errorf("delete entry %d: %w", id, err)
      }
      n, err := res.RowsAffected()
      if err != nil {
          return fmt.Errorf("delete entry %d: %w", id, err)
      }
      if n == 0 {
          return fmt.Errorf("delete entry %d: %w", id, ErrNotFound)
      }
      return nil
  }
  ```

- **Importing `bufio` and `strings`.** New in `delete.go`. Keep
  imports tidy (`gofmt` handles grouping).

- **Stream discipline.** Every write from `runDelete` goes to
  `cmd.ErrOrStderr()`. Nothing to `cmd.OutOrStdout()`. Verify with
  a `grep OutOrStdout internal/cli/delete.go` — should return zero
  hits.

- **`docs/api-contract.md` amendment.** Find the `brag delete <id>`
  section. Replace the existing bullet:
  ```
  Prompts for confirmation on stdin unless `--yes` is passed. Exit 1 if the ID does not exist or the user declines.
  ```
  with:
  ```
  Prompts for confirmation on stdin unless `--yes` (`-y`) is passed. Exit 1 if the ID does not exist, the arg is invalid, or missing. Exit 0 (no error) if the user declines the confirmation prompt — a deliberate choice, not an error.
  ```

- **`docs/tutorial.md` updates.**
  1. In §9 "What's NOT there yet", remove the
     `brag show <id>` / `brag edit <id>` / `brag delete <id>` line,
     or strike just the `delete` part if the line was previously
     compound. (Check current state; `show` already shipped so the
     table may need editing anyway — leave `edit <id>` in the
     STAGE-002 bucket.)
  2. Optionally add a short example in §4 or a new mini-section
     showing `brag delete 42` with the prompt and `brag delete 42
     -y`. Short; no need to belabor.

- **No `init()` functions** (§8).

---

## Build Completion

*Filled in at the end of the **build** cycle, before advancing to verify.*

- **Branch:** `feat/spec-008-brag-delete-command`
- **PR (if applicable):** (opened after advance-cycle)
- **All acceptance criteria met?** yes
- **New decisions emitted:**
  - None. Every design decision was inline in the spec; none rose to
    DEC-level weight during implementation.
- **Deviations from spec:**
  - Drive-by (disclosed): `docs/tutorial.md` §9 "What's NOT there
    yet" had a compound `show / edit / delete` line. Striking just
    `delete` would have left `show` — already shipped in SPEC-006 —
    incorrectly advertised as not-yet-shipped. I removed both `show`
    and `delete`, leaving the row as `edit <id>` only. This matches
    the spec's own parenthetical ("`show` already shipped so the
    table may need editing anyway — leave `edit <id>` in the
    STAGE-002 bucket"). Not a surprise, but disclosing here for
    completeness.
- **Follow-up work identified:**
  - `docs/tutorial.md` §4 still has the stale line "`brag show <id>`
    arrives in STAGE-002 to dump a single entry in full" even though
    SPEC-006 shipped it. Left untouched — out of scope here. A
    tutorial-refresh spec could sweep this once STAGE-002 finishes.

### Build-phase reflection (3 questions, short answers)

1. **What was unclear in the spec that slowed you down?**
   — Nothing slowed me down meaningfully. The spec was one of the
   cleanest so far: exact struct for `runDelete`, exact struct for
   `Store.Delete`, a full failing-test list with per-test assertion
   hints, and prior-art pointers to `show.go`/`show_test.go` for
   shape. The only soft spot was the `TestDeleteCmd_YesFlagLongAndShort`
   hint ("run twice, two entries") — cobra's parsed-args state is
   sticky on a single `*cobra.Command`, so I used two fresh roots to
   guarantee the second `SetArgs` took effect. Not a spec defect,
   just a cobra quirk worth noting for future multi-invocation tests.

2. **Was there a constraint or decision that should have been listed but wasn't?**
   — No. DEC-005/006/007, `no-sql-in-cli-layer`,
   `storage-tests-use-tempdir`, `stdout-is-for-data-stderr-is-for-
   humans`, `errors-wrap-with-context`, `test-before-implementation`,
   `one-spec-per-pr` — all already cited. Every path in `delete.go`
   stayed inside those rails without fighting them.

3. **If you did this task again, what would you do differently?**
   — I'd codify the "separate-root-per-SetArgs-invocation" pattern
   into a tiny `freshRoot(t, cmd) *cobra.Command` helper rather than
   inlining `NewRootCmd(...) + AddCommand(...)` a second time inside
   `TestDeleteCmd_YesFlagLongAndShort`. Not worth introducing in this
   spec (one test, two call sites), but the next time a test exercises
   two cobra invocations in the same test body it'd be nicer to have
   a named helper. Could land in a future test-harness polish spec.

---

## Reflection (Ship)

*Appended 2026-04-20 during the **ship** cycle. Outcome-focused,
distinct from the process-focused build reflection above.*

1. **What would I do differently next time?**
   Keep writing prescriptive specs. SPEC-008's build completed
   cleanly and quickly because the Implementation Context hand-fed
   the `runDelete` and `Store.Delete` code blocks — the same recipe
   SPEC-005 and SPEC-006 validated. Verify flagged a yellow
   observation that a "too-prescriptive" spec means the build
   session does "almost no design work," which is technically true
   but is actually the intended framework behavior. Not a change
   to make, but a framework-level nuance worth filing: the
   spec-as-code-prescription pattern is the thing paying interest,
   not an anti-pattern. Logged the nuance to
   `framework-feedback/process-feedback.md` as an addendum.

2. **Does any template, constraint, or decision need updating?**
   One small template-clarity note: the spec.md template's
   `agents.architect` and `agents.implementer` fields both default
   to the same model id for every spec, so they're informational
   noise under the `claude-only` variant, not a contamination
   signal. Verify's yellow flag on SPEC-008 misread "architect ==
   implementer == claude-opus-4-7" as evidence of same-session
   contamination, when it's actually just the template default.
   Adding a one-line note to AGENTS.md §2 ("Work Hierarchy")
   clarifying the fields' semantics under `claude-only`. This
   ship commit applies that note.

3. **Is there a follow-up spec to write now before I forget?**
   No. SPEC-009 (`internal/editor` package + `brag edit <id>` +
   `Store.Update`, M) is next pending in STAGE-002 — editor-launch
   capture, which is a genuine shape change from the ID-taking
   commands we've been shipping. SPEC-010/011/012 remain queued
   after.
