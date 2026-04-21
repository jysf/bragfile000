---
task:
  id: SPEC-006
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
    - DEC-005  # integer autoincrement IDs (we parse and print them)
    - DEC-006  # cobra framework
    - DEC-007  # RunE-validated required args (extends to positional args)
  constraints:
    - no-sql-in-cli-layer
    - storage-tests-use-tempdir
    - stdout-is-for-data-stderr-is-for-humans
    - errors-wrap-with-context
    - timestamps-in-utc-rfc3339
    - test-before-implementation
    - one-spec-per-pr
  related_specs:
    - SPEC-001  # shipped; root command + --db flag
    - SPEC-002  # shipped; storage layer, Entry type
    - SPEC-003  # shipped; add command, ErrUser pattern, main.go exit mapping
    - SPEC-005  # shipped; help-assertion specificity lesson
---

# SPEC-006: `brag show <id>` command

## Context

Second shipped spec in STAGE-002 (after SPEC-005). Today, `brag list`
prints only `<id>\t<created_at>\t<title>` per entry — the actual
description, tags, project, type, and impact fields are stored but
invisible through the CLI. Users have to reach for
`sqlite3 ~/.bragfile/db.sqlite` to see the full row. `brag show <id>`
closes that gap: thin cobra subcommand + one new `Store.Get(id)`
method + a small markdown renderer. When this spec ships, the
read-back loop is complete for the end user — they can capture with
`brag add`, scan with `brag list`, and drill into any row with
`brag show <id>`.

Parent stage: `STAGE-002-capture-and-retrieval.md`. Project: PROJ-001.

## Goal

Ship `brag show <id>` as a cobra subcommand that reads a single entry
via a new `Store.Get(id int64) (Entry, error)` method and prints the
entry as markdown (title, metadata table, optional description body)
to stdout. Non-existent IDs surface as `ErrUser` (exit code 1); any
other failure surfaces as internal error (exit code 2).

## Inputs

- **Files to read:**
  - `docs/api-contract.md` — "`brag show <id>`" section defines the
    expected markdown output shape.
  - `docs/data-model.md` — nullability of `description/tags/project/
    type/impact` drives the "omit empty metadata rows" behavior.
  - `AGENTS.md` §8 (conventions), §9 (testing: separate buffers,
    fail-first, assertion specificity).
  - `/decisions/DEC-005-integer-autoincrement-ids.md`
  - `/decisions/DEC-006-cobra-cli-framework.md`
  - `/decisions/DEC-007-required-flag-validation-in-runE.md`
  - `/guidance/constraints.yaml`
  - `internal/cli/add.go` + `internal/cli/list.go` (both shipped) —
    reference shape for a cobra subcommand that opens a store,
    queries it, and prints one thing to stdout.
  - `internal/cli/errors.go` — `ErrUser` sentinel + `UserErrorf`.
  - `internal/storage/store.go` — existing `Open`, `Add`, `List`
    methods; `Get` follows the same error-wrapping and timestamp-
    parsing pattern.
  - `internal/storage/entry.go` — `Entry` struct; `Store.Get` returns
    this type hydrated.
- **External APIs:** none.
- **Related code paths:** `internal/cli/`, `internal/storage/`,
  `cmd/brag/main.go`.

## Outputs

- **Files created:**
  - `internal/cli/show.go` — `NewShowCmd() *cobra.Command` +
    unexported `runShow` RunE handler + unexported `renderEntry`
    helper.
  - `internal/cli/show_test.go` — tests for argument parsing,
    output shape, missing-ID handling, help shape.
  - `internal/storage/errors.go` — new file; exports
    `ErrNotFound = errors.New("entry not found")`. (Could
    alternatively live at the top of `store.go` or `entry.go` — the
    new file keeps the domain-error surface tidy and gives us a
    home for future sentinels like `ErrDuplicate` if we ever need
    them.)
- **Files modified:**
  - `internal/storage/store.go` — add `Store.Get(id int64) (Entry,
    error)`.
  - `internal/storage/store_test.go` — add `TestGet_*` tests (or
    split into `get_test.go` — builder's call; the test content is
    what matters).
  - `cmd/brag/main.go` — register the subcommand with one added
    line: `root.AddCommand(cli.NewShowCmd())`. Nothing else.
- **New exports:**
  - `storage.ErrNotFound` (sentinel)
  - `(*storage.Store).Get(id int64) (storage.Entry, error)`
  - `cli.NewShowCmd() *cobra.Command`
- **Database changes:** none.

## Acceptance Criteria

- [ ] `brag show <id>` on an existing entry prints:
      - A `# <title>` line as the first stdout content.
      - A markdown table with `id`, `created_at`, `updated_at` rows
        (always) and `tags`, `project`, `type`, `impact` rows (only
        when those fields are non-empty).
      - A `## Description` section followed by the description body
        iff the description is non-empty.
      - Nothing to stderr.
      *[TestShowCmd_FullEntryRendersAllSections]*
- [ ] Metadata rows with empty string values are OMITTED entirely —
      the output never contains `| tags        |  |` for an entry
      whose tags field is empty. *[TestShowCmd_EmptyMetadataRowsOmitted]*
- [ ] `## Description` section is OMITTED entirely when the entry's
      description is empty. *[TestShowCmd_EmptyDescriptionSectionOmitted]*
- [ ] `brag show 999` on a DB that has no entry 999 returns an error
      that `errors.Is(err, cli.ErrUser)` matches (so `main.go` maps
      to exit code 1, user error). `outBuf` stays empty; stderr stays
      empty (`main.go` formats the error). *[TestShowCmd_NotFoundIsUserError]*
- [ ] `brag show` with no positional argument returns an `ErrUser`.
      *[TestShowCmd_NoArgsIsUserError]*
- [ ] `brag show` with more than one positional argument returns an
      `ErrUser`. *[TestShowCmd_TooManyArgsIsUserError]*
- [ ] `brag show abc` (non-numeric arg) returns an `ErrUser`.
      *[TestShowCmd_NonNumericArgIsUserError]*
- [ ] `brag show 0` returns an `ErrUser` (IDs are 1+ under
      AUTOINCREMENT per DEC-005). Same for `brag show -5`, though
      cobra may parse `-5` as a flag — if it does, document the
      cobra behavior in Build Completion; the spec's contract is
      "non-positive integer → ErrUser".
      *[TestShowCmd_NonPositiveArgIsUserError]*
- [ ] `brag show --help` prints usage to stdout with empty stderr;
      help contains the literal string `Examples:` (SPEC-005 lesson:
      assert on distinctive content, not a generic `brag show`
      substring that cobra's Usage line already carries).
      *[TestShowCmd_HelpShape]*
- [ ] `Store.Get(id)` on a fresh DB returns an error such that
      `errors.Is(err, storage.ErrNotFound)` matches.
      *[TestGet_NotFoundReturnsErrNotFound]*
- [ ] `Store.Get(id)` on an inserted entry returns the `Entry` with
      all fields populated, timestamps parsed as UTC.
      *[TestGet_RoundTripsAllFields]*
- [ ] Existing SPEC-001/002/003/004/005 tests remain green. No
      existing test is modified. *[manual: go test ./...]*
- [ ] `gofmt -l .` empty, `go vet ./...` clean, `CGO_ENABLED=0 go
      build ./...` succeeds, `go test ./...` green.

## Failing Tests

Written now. Every happy-path test uses separate `outBuf` / `errBuf`
with a no-cross-leakage assert on the empty side (§9). Every help
assert targets distinctive content, not cobra boilerplate (§9, from
SPEC-005 lesson). Fail-first run required before implementation (§9,
from SPEC-003 lesson); if an "unexpectedly passing" test appears,
tighten the assertion before proceeding.

### New tests in `internal/storage/store_test.go` (or new `get_test.go`)

- **`TestGet_RoundTripsAllFields`** — `Add` an `Entry` with every
  string field populated. Call `Store.Get(returned.ID)`. Assert the
  hydrated `Entry` equals the inserted one on `Title`, `Description`,
  `Tags`, `Project`, `Type`, `Impact`; assert `ID == returned.ID`;
  assert `CreatedAt.Location().String() == "UTC"` and
  `UpdatedAt.Equal(CreatedAt)`.

- **`TestGet_NotFoundReturnsErrNotFound`** — fresh store, no rows.
  Call `Store.Get(42)`. Assert `errors.Is(err, storage.ErrNotFound)`
  and the returned `Entry` is the zero value.

- **`TestGet_PartiallyEmptyFieldsHydrateAsEmptyStrings`** — `Add` an
  `Entry{Title: "only title"}`. `Get` it. Assert `Description == ""`,
  `Tags == ""`, etc. (verifies the `sql.NullString` scan path
  correctly produces empty strings rather than leaking `"<nil>"` or
  panicking).

### New tests in `internal/cli/show_test.go`

Use a new helper modeled on `newRootWithAdd` in `add_test.go`:

- `newRootWithShow(t *testing.T) (*cobra.Command, string)` — root
  with the show subcommand attached; returns root plus a
  `t.TempDir()`-backed DB path.

Tests:

- **`TestShowCmd_FullEntryRendersAllSections`** — open store at the
  temp path, `Add` an entry with every field populated (title
  `"cut p99 latency"`, description `"redis lookup"`, tags
  `"auth,perf"`, project `"platform"`, type `"shipped"`, impact
  `"unblocked mobile v3"`). Close. Run `show <id>` through the root
  command with separate out/err buffers. Assert: nil error,
  `errBuf.Len() == 0`, and `outBuf.String()` contains each of:
    - `"# cut p99 latency"` (title line)
    - `"| id          | 1"` (or similar — assert on the id value)
    - `"auth,perf"` (distinctive tags content, not the label "tags")
    - `"platform"` (distinctive project content)
    - `"shipped"` (distinctive type content)
    - `"unblocked mobile v3"` (distinctive impact content)
    - `"## Description"` (distinctive section label)
    - `"redis lookup"` (distinctive description body)

- **`TestShowCmd_EmptyMetadataRowsOmitted`** — `Add` an `Entry{Title:
  "only title"}`. Close. Run `show <id>`. Assert: `outBuf` contains
  `"# only title"`; `outBuf` does NOT contain `"| tags"`, `"| project"`,
  `"| type"`, or `"| impact"` (no empty metadata rows leak through).

- **`TestShowCmd_EmptyDescriptionSectionOmitted`** — same entry
  (no description). Assert: `outBuf` does NOT contain
  `"## Description"`.

- **`TestShowCmd_NotFoundIsUserError`** — fresh store, no rows. Run
  `show 999`. Assert: `err != nil`, `errors.Is(err, ErrUser)`,
  `outBuf.Len() == 0`, `errBuf.Len() == 0` (main.go formats user
  errors; the command itself only returns the error).

- **`TestShowCmd_NoArgsIsUserError`** — run `show` with no
  positional args. Assert: `err != nil`, `errors.Is(err, ErrUser)`.

- **`TestShowCmd_TooManyArgsIsUserError`** — run `show 1 2`.
  Assert: `err != nil`, `errors.Is(err, ErrUser)`.

- **`TestShowCmd_NonNumericArgIsUserError`** — run `show abc`.
  Assert: `err != nil`, `errors.Is(err, ErrUser)`.

- **`TestShowCmd_NonPositiveArgIsUserError`** — run `show 0`.
  Assert: `err != nil`, `errors.Is(err, ErrUser)`. (Do not test
  `show -5` — cobra may parse `-5` as a flag; that's cobra's domain
  to document, not ours.)

- **`TestShowCmd_HelpShape`** — run `show --help`. Assert: nil error,
  `errBuf.Len() == 0`, `outBuf.String()` contains `"Examples:"`
  (distinctive label), NOT just `"brag show"` (cobra's Usage line
  would carry that for free — SPEC-005 lesson).

Notes for the implementer on testing patterns:

- Run `go test ./...` once after the new tests exist and BEFORE any
  implementation change. Confirm each new test fails for the
  expected reason (missing `NewShowCmd`, missing `ErrNotFound`,
  missing `Store.Get`, or — for output-shape tests — missing expected
  substring). Exactly per AGENTS.md §9.
- If any test "unexpectedly passes" in that run, tighten the
  assertion before proceeding (§9 lesson from SPEC-005).
- Every CLI test uses separate `outBuf` / `errBuf` and asserts on
  both (§9, SPEC-001 lesson).

## Implementation Context

*Read before starting build. Self-contained handoff.*

### Decisions that apply

- `DEC-005` — `entries.id` is `INTEGER PRIMARY KEY AUTOINCREMENT`;
  `Entry.ID` is `int64`. `Store.Get` takes `id int64`; `show` parses
  `strconv.ParseInt(arg, 10, 64)`. IDs are always positive.
- `DEC-006` — Cobra is the framework. `NewShowCmd()` follows the
  same shape as `NewAddCmd()` and `NewListCmd()`. No new cobra
  features introduced.
- `DEC-007` — Required-flag validation lives in `RunE`, not in
  cobra's `MarkFlagRequired`. **This spec extends the principle to
  positional-argument validation** — do NOT use `cobra.ExactArgs(1)`
  or other `Args` validators that return unwrappable plain errors.
  Validate `len(args) == 1` and `strconv.ParseInt` success directly
  in `runShow`, returning `UserErrorf(...)` on failure. (If this
  extension feels weighty enough to deserve its own DEC, the build
  session may emit one — the author leans "no", since the
  underlying principle is already captured by DEC-007.)

### Constraints that apply

For `internal/cli/**`, `internal/storage/**`, `cmd/brag/**`:

- `no-sql-in-cli-layer` — blocking. `show.go` must not import
  `database/sql` or any driver. All data access goes through
  `Store.Get`.
- `storage-tests-use-tempdir` — blocking. Every storage test
  (`TestGet_*`) uses `t.TempDir()`.
- `stdout-is-for-data-stderr-is-for-humans` — blocking. The
  rendered entry goes to `cmd.OutOrStdout()`. The command itself
  emits nothing on stderr; `main.go` formats any returned error to
  stderr. Every happy-path test asserts `errBuf.Len() == 0`.
- `errors-wrap-with-context` — warning. Every returned error from
  `runShow` is wrapped: `fmt.Errorf("resolve db path: %w", err)`,
  `fmt.Errorf("open store: %w", err)`, `fmt.Errorf("get entry: %w",
  err)`. User-error returns via `UserErrorf` are their own thing and
  don't need additional wrapping.
- `timestamps-in-utc-rfc3339` — blocking. `Store.Get` parses
  `created_at` and `updated_at` via `time.Parse(time.RFC3339, raw)`.
  The renderer formats via `.UTC().Format(time.RFC3339)` (calling
  `.UTC()` is defensive idempotence — the value is already UTC after
  parse).
- `test-before-implementation` — blocking. Tests first, then
  implementation.
- `one-spec-per-pr` — blocking. Branch
  `feat/spec-006-brag-show-command`. Diff touches only the files
  listed in Outputs.

### AGENTS.md lessons that apply

- §9 separate `outBuf`/`errBuf` in every CLI test (SPEC-001).
- §9 monotonic tie-break for time-based ordering — not directly
  relevant here (no ordering in show), but `Store.Get` is by primary
  key (SPEC-002).
- §9 fail-first test run before implementation (SPEC-003 ship).
- §9 **assertion specificity** (SPEC-005 ship) — every output-shape
  assert targets distinctive content, not labels cobra already
  produces or words that appear in the test setup.
- §10 `/`-anchored gitignore — no new ignore patterns in this spec.

### Prior related work

- **SPEC-002** (shipped). Provides `Entry` struct, `Store.List`
  (which uses the same `sql.NullString` scan pattern `Store.Get`
  will), `timestamps-in-utc-rfc3339` behavior, `storage-tests-use-
  tempdir` pattern.
- **SPEC-003** (shipped). `internal/cli/add.go` establishes the
  per-command shape: cobra constructor, `RunE` that resolves `--db`,
  opens store, does the thing, prints to stdout. `internal/cli/
  errors.go` defines `ErrUser` + `UserErrorf`.
- **SPEC-004** (shipped). `internal/cli/list.go` shows how to
  iterate a storage query result set and render one line per row —
  `show` is simpler (one row, richer rendering).
- **SPEC-005** (shipped). `Long` descriptions with an `Examples:`
  block and the help-assertion-specificity lesson apply here.

### Out of scope (for this spec specifically)

If any of these feels necessary during build, write a new spec.

- **Styled / colored terminal output.** Straight markdown to stdout.
  Users can pipe through `glow`, `bat`, or a renderer of their
  choice.
- **`--format json` or other output formats.** STAGE-003 (export)
  covers alternate serialization.
- **Editing after showing.** SPEC-009 ships `brag edit <id>`.
- **Auto-opening in `$PAGER`.** Out of MVP scope.
- **Showing multiple IDs at once** (`brag show 1 2 3`). Not in
  `api-contract.md`. Reject as `ErrUser` per the too-many-args test.
- **Partial-field display flags** (`brag show 1 --field impact`).
  Premature; `awk -F'|'` or `jq` on a future json format covers it.
- **Backward-compatible column width / alignment.** Markdown renders
  without alignment; don't reach for `text/tabwriter`.
- **New DEC for positional-arg validation.** DEC-007 already covers
  the principle; Implementation Context flags this but doesn't
  require a new DEC. Build session may emit one if it disagrees.

## Notes for the Implementer

- **`Store.Get` shape.** Mirror `Store.List`'s scan path for
  nullable columns:
  ```go
  var desc, tags, proj, typ, imp sql.NullString
  var createdAtRaw, updatedAtRaw string
  row := s.db.QueryRow(q, id)
  if err := row.Scan(&e.ID, &e.Title, &desc, &tags, &proj, &typ, &imp,
      &createdAtRaw, &updatedAtRaw); err != nil {
      if errors.Is(err, sql.ErrNoRows) {
          return Entry{}, fmt.Errorf("get entry %d: %w", id, ErrNotFound)
      }
      return Entry{}, fmt.Errorf("get entry %d: %w", id, err)
  }
  ```
  Then assign the `sql.NullString.String` values to the `Entry`
  fields and parse the two timestamps.

- **`ErrNotFound` location.** Put it in a new
  `internal/storage/errors.go` to keep the domain-error surface
  tidy. If the builder prefers adding it to `entry.go`, that's also
  fine; call out the choice in Build Completion deviations.

- **Argument validation in RunE.**
  ```go
  func runShow(cmd *cobra.Command, args []string) error {
      if len(args) != 1 {
          return UserErrorf("show requires exactly one <id> argument")
      }
      id, err := strconv.ParseInt(args[0], 10, 64)
      if err != nil {
          return UserErrorf("invalid id %q: must be a positive integer", args[0])
      }
      if id <= 0 {
          return UserErrorf("invalid id %d: must be positive", id)
      }
      // ... resolve db path, open store, Get, render
  }
  ```
  Do NOT set `cmd.Args = cobra.ExactArgs(1)` — see DEC-007
  extension above.

- **Renderer shape.** Keep it as an unexported helper inside
  `show.go`:
  ```go
  func renderEntry(w io.Writer, e storage.Entry) {
      fmt.Fprintf(w, "# %s\n\n", e.Title)
      fmt.Fprintln(w, "| field       | value |")
      fmt.Fprintln(w, "| ----------- | ----- |")
      fmt.Fprintf(w, "| id          | %d |\n", e.ID)
      fmt.Fprintf(w, "| created_at  | %s |\n", e.CreatedAt.UTC().Format(time.RFC3339))
      fmt.Fprintf(w, "| updated_at  | %s |\n", e.UpdatedAt.UTC().Format(time.RFC3339))
      if e.Tags != "" {
          fmt.Fprintf(w, "| tags        | %s |\n", e.Tags)
      }
      if e.Project != "" {
          fmt.Fprintf(w, "| project     | %s |\n", e.Project)
      }
      if e.Type != "" {
          fmt.Fprintf(w, "| type        | %s |\n", e.Type)
      }
      if e.Impact != "" {
          fmt.Fprintf(w, "| impact      | %s |\n", e.Impact)
      }
      if e.Description != "" {
          fmt.Fprintf(w, "\n## Description\n\n%s\n", e.Description)
      }
  }
  ```
  Exact whitespace/alignment isn't load-bearing — markdown viewers
  don't care. Tests should assert on distinctive content substrings,
  not exact-byte equality.

- **Command `Long` with Examples block.** Mirror the SPEC-005
  pattern:
  ```go
  Long: `Show a single brag entry as markdown.

  Examples:
    brag show 42              # print entry 42
    brag show 42 | glow       # render in a markdown viewer`,
  ```

- **Register in `main.go`.** One added line next to the existing
  `AddCommand` calls:
  ```go
  root.AddCommand(cli.NewShowCmd())
  ```
  Nothing else in `main.go` changes.

- **No `init()` functions** (AGENTS.md §8).

---

## Build Completion

*Filled in at the end of the **build** cycle, before advancing to verify.*

- **Branch:** `feat/spec-006-brag-show-command`
- **PR (if applicable):** <opened after advance-cycle>
- **All acceptance criteria met?** yes
- **New decisions emitted:**
  - None. DEC-007's positional-arg extension was handled inline per
    the spec's Implementation Context guidance — the principle is
    identical to the flag case, so a new DEC would just duplicate
    DEC-007's reasoning. Verify can push back if it disagrees.
- **Deviations from spec:**
  - None of substance. `storage.ErrNotFound` landed in the new
    `internal/storage/errors.go` file as the spec preferred. Used
    `errors.Is(err, sql.ErrNoRows)` (imported `errors`) in
    `store.go` to map the not-found path; this matches the
    Implementation Context pseudocode verbatim.
  - Smoke-tested `brag show 1` end-to-end against a real temp DB
    before advancing. Output matches the documented shape; the
    `ErrUser` path surfaces `brag: user error: no entry with id 999`
    on stderr with exit 1 (`main.go` does the formatting, as
    required by `stdout-is-for-data-stderr-is-for-humans`). The
    `user error:` prefix falls out of `UserErrorf`'s `%w`
    wrapping — consistent with how `add` already surfaces its user
    errors, so no deviation from prior behavior.
- **Follow-up work identified:**
  - None new. SPEC-007 (`list` filter flags) remains the next
    pending item in STAGE-002's backlog and is unblocked by this
    ship.
- **Cobra `show -5` behavior note (acceptance criterion clause):**
  cobra's pflag parser consumes a leading `-5` as an unknown short
  flag and returns its own error from argv parsing, not from
  `runShow`. That surface (unknown-flag error, not `ErrUser`) is
  cobra's domain; the spec explicitly flagged this and the test
  suite intentionally only covers `show 0`, not `show -5`. No
  change required.

### Build-phase reflection (3 questions, short answers)

1. **What was unclear in the spec that slowed you down?**
   — Nothing material. The Implementation Context was unusually
   thorough: it listed the exact scan-path, the exact `runShow`
   skeleton, and the exact `renderEntry` layout, so the only
   decisions left to me were trivial (import grouping, where to
   place `errors.Is` imports, what to name the test helper).

2. **Was there a constraint or decision that should have been listed but wasn't?**
   — No. The seven referenced constraints plus DEC-005/006/007
   covered every real decision. The one thing worth capturing as
   a stage-level pattern (not a new constraint): the fail-first
   run really did earn its keep here — it surfaced the three
   missing-symbol errors before I wrote any implementation code,
   which is exactly what §9 promises. Worth leaving in AGENTS.md
   as-is.

3. **If you did this task again, what would you do differently?**
   — Reach for `strconv.FormatInt` at the call site from the
   start. I briefly wrote a one-line `itoa` wrapper in
   `show_test.go`, then immediately deleted it before the first
   commit — a wrapper that costs a reader more than the inline
   call it replaces is just noise. Noted for future test-helper
   instincts.

---

## Reflection (Ship)

*Appended 2026-04-20 during the **ship** cycle. Outcome-focused,
distinct from the process-focused build reflection above.*

1. **What would I do differently next time?**
   Keep investing in the same pattern that just worked: thorough
   Implementation Context, prescriptive test list, distinctive-token
   assertions. SPEC-006's build was unusually clean, and the build
   session explicitly credited the spec's Implementation Context for
   the lack of ambiguity. Not "I would change X" but "this is now
   the proven recipe for an S spec." Keep the bar this high for
   SPEC-007 / SPEC-008 even though smaller specs tempt shorter specs.

2. **Does any template, constraint, or decision need updating?**
   One small documentation amendment: append a note to DEC-007's
   References section that the principle has been extended
   successfully to positional-arg validation in SPEC-006 (no
   `cobra.ExactArgs`). Not a supersession — the pattern is unchanged;
   the DEC's underlying reasoning generalizes. This turns DEC-007
   from a one-use decision into the documented canonical pattern for
   "cobra built-in validators return unwrappable errors → validate
   in RunE." Makes it discoverable for future specs that run into
   the same concern (SPEC-008 delete will). Applied in this ship
   commit.

3. **Is there a follow-up spec I should write now before I forget?**
   No. SPEC-007 (`list` filter flags, M) is next pending in STAGE-002
   and unblocked. SPEC-008 (`delete`) is queued after. Nothing
   slipping.
