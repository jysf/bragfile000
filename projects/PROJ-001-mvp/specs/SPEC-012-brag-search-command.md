---
task:
  id: SPEC-012
  type: story
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
  created_at: 2026-04-22

references:
  decisions:
    - DEC-005  # integer autoincrement IDs (tie-break for relevance ties)
    - DEC-006  # cobra framework
    - DEC-007  # RunE-validated flags AND positional args
    - DEC-010  # search query syntax (tokenize + phrase-quote + AND-join)
  constraints:
    - no-sql-in-cli-layer
    - storage-tests-use-tempdir
    - stdout-is-for-data-stderr-is-for-humans
    - errors-wrap-with-context
    - test-before-implementation
    - one-spec-per-pr
  related_specs:
    - SPEC-004  # shipped; list command shape (tab-separated output)
    - SPEC-011  # shipped; FTS5 index + triggers (entries_fts exists)
---

# SPEC-012: `brag search "query"` command

## Context

Last spec in STAGE-002. Consumes the FTS5 infrastructure shipped in
SPEC-011 via a new `Store.Search` method + a thin `brag search`
cobra subcommand. Query semantics are locked by DEC-010 (auto-
tokenize + phrase-quote user input + AND-join) — the decision
earned its own DEC because SPEC-011 build empirically surfaced
FTS5's hyphen-as-NOT surprise and the rejected alternatives (raw
FTS5 pass-through, single-phrase quote, hybrid detection) are
non-trivial to re-evaluate later.

Output matches `brag list` byte-for-byte: tab-separated
`<id>\t<created_at>\t<title>`, newline-terminated. Users piping
`brag list | cut -f3` and `brag search "x" | cut -f3` get
identical behavior. Relevance-sorted by default (FTS5's `rank`
pseudo-column) with `id DESC` as the tie-break so results are
deterministic even when multiple rows share the same rank score.

Premise audit per AGENTS.md §9 (SPEC-010 rule + SPEC-011
extension):

- **Inversion/removal**: none. Search is a pure addition; no
  existing behavior inverts.
- **Additive-to-tracked-collection**: adds DEC-010 to the decisions
  collection. Grep `grep -rn "decisions" internal/**/*_test.go`
  → no existing tests assert on DEC count or list decisions by
  literal name. Clean; no existing test updates needed.

Parent stage: `STAGE-002-capture-and-retrieval.md`. Project:
PROJ-001. When this spec ships, STAGE-002 closes at 8/8.

## Goal

Ship `brag search "query"` as a cobra subcommand that parses user
input per DEC-010 (tokenize + phrase-quote + AND-join), calls a
new `Store.Search(query string) ([]Entry, error)` method which
issues a single FTS5 `MATCH ... ORDER BY rank, id DESC` query,
and prints results in the same tab-separated format as `brag list`
to stdout. Missing / empty / quote-containing queries return
`ErrUser` (exit 1); storage failures surface as internal error
(exit 2); zero-result queries return cleanly with no stdout and
no stderr (exit 0 — valid empty result is not an error).

## Inputs

- **Files to read:**
  - `docs/api-contract.md` — `brag search` section defines the
    shape; this spec refines the exit-code contract and references
    DEC-010 for query semantics.
  - `docs/data-model.md` — `entries_fts` virtual table (shipped in
    SPEC-011).
  - `AGENTS.md` §8 (coding conventions), §9 (testing: separate
    buffers, fail-first, assertion specificity, locked-decisions-
    need-tests, premise audit both cases), §12 "During design".
  - `/decisions/DEC-005-integer-autoincrement-ids.md` — `id DESC`
    tie-break when FTS5 rank is equal.
  - `/decisions/DEC-006-cobra-cli-framework.md`
  - `/decisions/DEC-007-required-flag-validation-in-runE.md` —
    applies to the single positional query arg (same pattern as
    `show`/`delete`/`edit`).
  - `/decisions/DEC-010-search-query-syntax.md` — the canonical
    query-building behavior.
  - `/guidance/constraints.yaml`
  - `internal/cli/list.go` — reference shape for a cobra command
    that calls a Store method and prints tab-separated rows.
  - `internal/cli/list_test.go` — reference for the happy-path
    test shape.
  - `internal/cli/errors.go` — `ErrUser` + `UserErrorf`.
  - `internal/storage/store.go` — existing `Store.List` / `Get`
    / `Add` / `Update` / `Delete`; `Search` lives alongside.
  - `internal/storage/fts_test.go` (shipped in SPEC-011) — prior
    art for FTS5 test patterns; `Store.Search` tests extend this
    file rather than adding a new one (keeps FTS concerns
    consolidated).
  - `internal/storage/migrations/0002_add_fts.sql` — reference for
    the column list; `Store.Search` joins back to `entries` to
    hydrate the full row, not just ids.
- **External APIs:** none.
- **Related code paths:** `internal/cli/`, `internal/storage/`,
  `cmd/brag/main.go`.

## Outputs

- **Files created:**
  - `internal/cli/search.go` — `NewSearchCmd() *cobra.Command` +
    unexported `runSearch` + unexported `buildFTS5Query(raw
    string) (string, error)` helper implementing DEC-010.
  - `internal/cli/search_test.go` — CLI + `buildFTS5Query`
    tests (see Failing Tests).
- **Files modified:**
  - `internal/storage/store.go` — add `Store.Search(query string,
    limit int) ([]Entry, error)`. `limit <= 0` → no LIMIT
    applied (unlimited, matches `Store.List(ListFilter{})`
    convention). `query == ""` → `ErrUser`-compatible error
    wrapping a sentinel so CLI can map cleanly.
  - `internal/storage/fts_test.go` — append `TestSearch_*` tests
    alongside existing FTS tests.
  - `cmd/brag/main.go` — register the subcommand with one added
    line: `root.AddCommand(cli.NewSearchCmd())`.
  - `docs/api-contract.md` — update `brag search` section:
    (a) reference DEC-010 for query semantics, (b) document
    exit codes (0 on success OR zero-results; 1 on user error;
    2 on internal error), (c) note `--limit N` flag.
  - `docs/tutorial.md` — strike `brag search` from §9 "What's NOT
    there yet"; add a short "Search your entries" mini-section
    near the existing `brag list` filter examples.
- **New exports:**
  - `(*storage.Store).Search(query string, limit int) ([]Entry, error)`
  - `cli.NewSearchCmd() *cobra.Command`
- **New sentinel (optional but cleaner):**
  - `storage.ErrEmptyQuery = errors.New("empty search query")` —
    returned by `Store.Search` when the query argument is empty.
    CLI layer wraps as `UserErrorf`. If the builder prefers to
    inline validation in `runSearch` before calling the store, no
    sentinel is needed. Builder's call.
- **Database changes:** none. SPEC-011's `entries_fts` is the
  index; SPEC-012 only reads from it.

## Locked design decisions (inline — DEC-010 covers the non-trivial part)

Each has a paired failing test per §9.

1. **Query semantics per DEC-010.** `buildFTS5Query` tokenizes the
   user input, phrase-quotes each non-empty token, joins with
   `" "`. Empty / whitespace-only / quote-containing input returns
   an error that `runSearch` maps to `UserErrorf`.
   *Tests: TestBuildFTS5Query_SingleWord,
   TestBuildFTS5Query_MultiWordAnd,
   TestBuildFTS5Query_HyphenatedLiteral,
   TestBuildFTS5Query_EmptyIsError,
   TestBuildFTS5Query_WhitespaceOnlyIsError,
   TestBuildFTS5Query_QuoteInQueryIsError.*

2. **Output format mirrors `list` exactly.** Tab-separated
   `<id>\t<created_at>\t<title>\n`. No ranking column, no snippet,
   no decoration. Pipeability parity with `brag list`.
   *Test: TestSearchCmd_TabSeparatedOutput.*

3. **Ordering: `ORDER BY rank, id DESC`.** FTS5's `rank` pseudo-
   column (lower = more relevant in the default bm25 setup).
   `id DESC` breaks ties for determinism. Tests must produce
   multiple rows with indistinguishable rank to exercise the
   tie-break.
   *Test: TestSearch_OrdersByRelevanceThenIdDesc.*

4. **`--limit N` flag, same shape as `list`.** `--limit 0` →
   unlimited (matches `Store.List`'s zero-is-no-limit convention).
   `--limit -5` or other non-positive non-zero values → ErrUser.
   Included in this spec (not deferred) because it's trivially
   `LIMIT ?` in SQL and users expect parity with `brag list`.
   Other filter flags (`--tag`, `--project`, `--since`) are
   **deferred** to a future polish spec to keep SPEC-012 S-sized.
   *Tests: TestSearchCmd_LimitRespected,
   TestSearchCmd_InvalidLimitIsUserError.*

5. **Zero results is not an error.** Query matches nothing →
   stdout empty, stderr empty, exit 0. Matches `brag list` on an
   empty DB.
   *Test: TestSearchCmd_ZeroResultsExitsZero.*

6. **Single positional query argument.** `brag search` (no arg)
   and `brag search a b` (two args) both return `ErrUser`. Users
   quote multi-word queries: `brag search "cut latency"`.
   Enforces DEC-007's RunE-validation pattern (no
   `cobra.ExactArgs(1)`).
   *Tests: TestSearchCmd_NoArgsIsUserError,
   TestSearchCmd_TooManyArgsIsUserError.*

7. **Store.Search signature.**
   `(*Store).Search(query string, limit int) ([]Entry, error)`.
   Hydrates full `Entry` rows (not just ids or titles) by
   joining `entries_fts` back to `entries`. Mirrors `Store.List`'s
   shape so CLI layer's rendering is identical.
   *Test: TestSearch_ReturnsHydratedEntries.*

8. **Empty/invalid query handled at CLI layer, not storage.**
   `buildFTS5Query` returns an error on empty/quote input;
   `Store.Search` assumes the query is already validated. Keeps
   storage dumb, CLI responsible for input shape.
   *Test: TestBuildFTS5Query_* series (decisions #1 + #8
   collapse to the same tests).*

## Acceptance Criteria

- [ ] `brag search "shipped"` returns entries containing "shipped"
      in any FTS5-indexed field
      (title/description/tags/project/impact).
      *[TestSearchCmd_HappyPath]*
- [ ] `brag search "auth-refactor"` finds entries containing both
      tokens (the hyphen does NOT trigger FTS5 NOT-operator
      interpretation). *[TestBuildFTS5Query_HyphenatedLiteral +
      TestSearchCmd_HyphenatedQuery]*
- [ ] `brag search "cut latency"` returns entries containing BOTH
      words anywhere (AND semantics). A row with only "cut" or
      only "latency" is not returned.
      *[TestSearchCmd_MultiWordAndSemantics]*
- [ ] `brag search "nonesuch_query_token_zzz"` returns no rows;
      stdout empty, stderr empty, exit 0.
      *[TestSearchCmd_ZeroResultsExitsZero]*
- [ ] `brag search ""` returns `ErrUser`; stdout empty.
      *[TestSearchCmd_EmptyQueryIsUserError]*
- [ ] `brag search '  '` (all whitespace) returns `ErrUser`.
      *[TestSearchCmd_WhitespaceOnlyQueryIsUserError]*
- [ ] `brag search 'with "quote"'` returns `ErrUser`.
      *[TestSearchCmd_QuoteInQueryIsUserError]*
- [ ] `brag search` (no arg) and `brag search a b` (too many args)
      both return `ErrUser`. *[TestSearchCmd_NoArgsIsUserError,
      TestSearchCmd_TooManyArgsIsUserError]*
- [ ] Output format: tab-separated `<id>\t<created_at>\t<title>\n`,
      identical byte shape to `brag list`.
      *[TestSearchCmd_TabSeparatedOutput]*
- [ ] Results ordered by FTS5 `rank` ascending, with `id DESC` as
      tie-break. Two rows with identical rank come back in id-
      descending order. *[TestSearch_OrdersByRelevanceThenIdDesc]*
- [ ] `brag search "query" --limit 3` caps the result count at 3.
      *[TestSearchCmd_LimitRespected]*
- [ ] `brag search "query" --limit -5` returns `ErrUser`.
      *[TestSearchCmd_InvalidLimitIsUserError]*
- [ ] `brag search --help` output contains the literal
      `"Examples:"` label (distinctive per SPEC-005 assertion-
      specificity rule) and references DEC-010 shape via an
      example. *[TestSearchCmd_HelpShape]*
- [ ] `Store.Search("query", 0)` returns all matching rows (no
      limit applied when `limit <= 0`).
      *[TestSearch_ZeroLimitMeansUnlimited]*
- [ ] `Store.Search` returns hydrated `Entry` values (all fields
      populated, not just id/title).
      *[TestSearch_ReturnsHydratedEntries]*
- [ ] Existing SPEC-001..011 tests remain green. No existing test
      modified. *[manual: go test ./...]*
- [ ] `gofmt -l .` empty, `go vet ./...` clean, `CGO_ENABLED=0 go
      build ./...` succeeds, `go test ./...` green.
- [ ] `docs/api-contract.md` `brag search` section updated
      (DEC-010 reference, exit codes, `--limit`).
- [ ] `docs/tutorial.md` §9 strikes `brag search`; a short
      "Search your entries" subsection added near the list-filter
      examples.

## Failing Tests

Written now. All CLI tests use separate `outBuf` / `errBuf` with
no-cross-leakage asserts (§9 SPEC-001). Every output assertion
targets distinctive content (§9 SPEC-005). Fail-first run before
implementation (§9 SPEC-003). Every locked design decision above
has at least one paired failing test (§9 SPEC-009, SPEC-010,
SPEC-011 rules).

### `internal/cli/search_test.go` (new file)

Pure-function tests for `buildFTS5Query` live at the top of this
file (no DB, no cobra). CLI tests use the now-standard
`newRootWith<Cmd>(t)` helper pattern.

Imports: `testing`, `bytes`, `errors`, `strings`, `strconv`, cli,
storage, cobra.

**`buildFTS5Query` tests (pure function):**

- **`TestBuildFTS5Query_SingleWord`** — input `"latency"`; expect
  output `` `"latency"` `` and nil error.
- **`TestBuildFTS5Query_MultiWordAnd`** — input `"cut latency"`;
  expect output `` `"cut" "latency"` ``.
- **`TestBuildFTS5Query_HyphenatedLiteral`** — input
  `"auth-refactor"`; expect output `` `"auth-refactor"` `` (hyphen
  preserved inside the phrase-quote). A follow-up MATCH using
  this output should find rows with adjacent `auth` + `refactor`
  tokens; that end-to-end path is covered in the CLI test below.
- **`TestBuildFTS5Query_EmptyIsError`** — input `""`; expect
  non-nil error; output empty string.
- **`TestBuildFTS5Query_WhitespaceOnlyIsError`** — input `"   "`
  (spaces and tabs); expect non-nil error.
- **`TestBuildFTS5Query_QuoteInQueryIsError`** — input
  `` `with "quote"` ``; expect non-nil error. Locks the DEC-010
  decision to reject rather than escape.

**CLI tests:**

Use a new `newRootWithSearch(t) (*cobra.Command, string)` helper
that builds root + search subcommand, attaches them to a
`t.TempDir()`-backed DB, returns root + DB path. Seeds entries
via a pre-open `Store.Add` dance when a test needs data.

- **`TestSearchCmd_HappyPath`** — seed three entries with
  distinctive, non-overlapping content:
  - id 1: title=`"alpha distinctive"`, other fields unrelated
  - id 2: title=`"beta"`, description=`"alpha distinctive"`
  - id 3: title=`"gamma"`, unrelated content
  Run `search "alpha distinctive"`. Assert: err nil, two rows in
  stdout (id 2 then id 1 by rank; both contain the phrase — but
  because DEC-010 builds `"alpha" "distinctive"` as AND, both
  rows qualify). Assert each row is tab-separated and the titles
  match expected.

- **`TestSearchCmd_HyphenatedQuery`** — seed an entry with
  description `"the auth refactor landed cleanly"`. Run
  `search "auth-refactor"`. Assert err nil, stdout contains the
  row. The default unicode61 tokenizer indexes `auth` and
  `refactor` as separate tokens; the phrase-quoted query
  `"auth-refactor"` also tokenizes to those two tokens adjacent.
  Rows with the two tokens adjacent match.

- **`TestSearchCmd_MultiWordAndSemantics`** — seed three entries:
  - title=`"cut latency"`
  - title=`"only cut"` (no latency)
  - title=`"only latency"` (no cut)
  Run `search "cut latency"`. Assert exactly one row returned
  (the first one); the other two are excluded by AND semantics.

- **`TestSearchCmd_ZeroResultsExitsZero`** — seed one entry; run
  `search "xyznomatchxyzxyz"`. Assert err nil, `outBuf.Len() == 0`,
  `errBuf.Len() == 0`.

- **`TestSearchCmd_EmptyQueryIsUserError`** — `search ""`. Assert
  `errors.Is(err, ErrUser)`, `outBuf.Len() == 0`.

- **`TestSearchCmd_WhitespaceOnlyQueryIsUserError`** —
  `search "   "`. Assert `errors.Is(err, ErrUser)`.

- **`TestSearchCmd_QuoteInQueryIsUserError`** —
  `search 'with "quote"'`. Assert `errors.Is(err, ErrUser)`.

- **`TestSearchCmd_NoArgsIsUserError`** — `search` with no
  positional. Assert `errors.Is(err, ErrUser)`.

- **`TestSearchCmd_TooManyArgsIsUserError`** — `search a b`.
  Assert `errors.Is(err, ErrUser)`.

- **`TestSearchCmd_TabSeparatedOutput`** — seed one entry with a
  known title; run `search "<title-word>"`. Assert the single
  output line has exactly two tabs, and splitting on `"\t"`
  yields three fields in the shape `[id, rfc3339, title]`.

- **`TestSearchCmd_LimitRespected`** — seed five matching entries;
  run `search "matchword" --limit 3`. Assert three rows returned.

- **`TestSearchCmd_InvalidLimitIsUserError`** — `search "x"
  --limit -5`. Assert `errors.Is(err, ErrUser)`.

- **`TestSearchCmd_HelpShape`** — `search --help`. Assert nil
  error, `errBuf.Len() == 0`, `outBuf` contains the distinctive
  `"Examples:"` label.

### `internal/storage/fts_test.go` (append existing file)

- **`TestSearch_OrdersByRelevanceThenIdDesc`** — seed three
  entries such that two have genuinely identical FTS5 rank (e.g.,
  all three have the same query word exactly once in the same
  field). Call `Store.Search("word", 0)`. Because ranks match,
  `id DESC` tie-break applies: assert the returned slice is in
  id-descending order (highest id first among the tied-rank
  matches).

- **`TestSearch_ReturnsHydratedEntries`** — seed one entry with
  every field populated. Call `Store.Search("uniquetitle", 0)`.
  Assert the single returned Entry has all fields hydrated, not
  just `ID` / `Title` / `CreatedAt`.

- **`TestSearch_ZeroLimitMeansUnlimited`** — seed seven matching
  entries. Call `Store.Search("word", 0)`. Assert all seven
  returned. Also call `Store.Search("word", -1)`: assert all
  seven returned (negative limit also means "no limit"; matches
  storage package's zero/negative-as-unset convention).

Notes for the implementer on testing patterns:

- Fail-first: write all ~15 tests, run `go test ./...` once
  BEFORE any implementation. Expected failures: `search.go` and
  `buildFTS5Query` don't exist yet; `Store.Search` doesn't exist
  yet. Confirm each fails for the expected symbol-missing reason
  before writing code.
- FTS5 rank semantics: lower rank = more relevant (bm25 default
  returns negative numbers for matches; lower absolute value =
  stronger match). Assertions should use "is strictly
  less/greater" rather than hard-coded rank values.
- Use `strings.Fields(input)` for whitespace tokenization in
  `buildFTS5Query`. It splits on any whitespace run and returns
  no empty tokens. Simpler than a manual loop.

## Implementation Context

*Read before starting build. Self-contained handoff.*

### Decisions that apply

- `DEC-005` — Integer IDs; `id DESC` tie-break for deterministic
  ordering when FTS5 ranks are identical.
- `DEC-006` — Cobra. `NewSearchCmd` mirrors `NewListCmd` shape.
- `DEC-007` — Validate args in RunE, not via `cobra.ExactArgs`.
  Applies to both the positional query arg and `--limit` flag.
- `DEC-010` — **The query-syntax decision.** Read the full DEC
  before implementing `buildFTS5Query`; the alternatives
  considered (raw pass-through, single-phrase wrap, hybrid) are
  explicitly rejected and the builder should not revisit them
  without a new DEC.

### Constraints that apply

For `internal/cli/**`, `internal/storage/**`, `cmd/brag/**`,
`docs/**`:

- `no-sql-in-cli-layer` — blocking. `search.go` imports only
  `config`, `storage`, `editor` (not needed here), and stdlib.
  The `buildFTS5Query` helper is pure string manipulation — no
  SQL construction, just FTS5 syntax transformation.
- `storage-tests-use-tempdir` — blocking. All new storage tests
  use `t.TempDir()`.
- `stdout-is-for-data-stderr-is-for-humans` — blocking. Matching
  rows → stdout. Zero-result silence → both empty. Errors →
  handled by main.go (→ stderr). Tests assert both streams.
- `errors-wrap-with-context` — warning. Wrap every returned
  error path: `fmt.Errorf("resolve db path: %w", err)`,
  `fmt.Errorf("open store: %w", err)`,
  `fmt.Errorf("search: %w", err)`.
- `test-before-implementation` — blocking.
- `one-spec-per-pr` — blocking. Branch
  `feat/spec-012-brag-search-command`.

### AGENTS.md lessons that apply

- §9 separate `outBuf` / `errBuf` (SPEC-001).
- §9 fail-first (SPEC-003).
- §9 assertion specificity — help test targets `"Examples:"`
  literal, not generic `"brag search"` that cobra's Usage line
  would render anyway (SPEC-005).
- §9 locked-decisions-need-tests — each of the 8 locked decisions
  above has ≥1 paired test in Failing Tests (SPEC-009).
- §9 premise audit (both branches) — explicitly performed.
  **Inversion/removal**: none, search is purely additive
  (SPEC-010 rule). **Addition-to-tracked-collection**: adds
  DEC-010; no existing tests assert on DEC counts (SPEC-011
  rule).
- §12 "During design" (SPEC-007) — every implementation option
  in Notes passes every blocking constraint. No "either is
  acceptable" language.

### Prior related work

- **SPEC-004** (shipped). `internal/cli/list.go` +
  `Store.List(ListFilter{})` — `brag search` mirrors both
  shapes (tab-separated stdout, hydrated Entry slice). `list`
  orders by `created_at DESC, id DESC`; `search` orders by
  `rank, id DESC`. Different primary sort, same tie-break
  philosophy.
- **SPEC-007** (shipped). `Store.List` WHERE clause + filter
  flags. SPEC-012 borrows the `--limit` shape but defers
  `--tag / --project / --type / --since` to a future polish
  spec — keeping SPEC-012 S-sized.
- **SPEC-011** (shipped). `entries_fts` virtual table + triggers
  + backfill. Search reads from this table; triggers keep it
  fresh on every Add / Update / Delete automatically.

### Out of scope (for this spec specifically)

Write a new spec rather than expanding SPEC-012 if any of these
pull.

- **Filter flags on `search`** (`--tag`, `--project`, `--type`,
  `--since`). Deferred to a future polish spec. Requires
  composing WHERE clauses on top of MATCH, which doubles the
  SQL-building surface.
- **`--raw` / `--advanced` flag** to pass query through to FTS5
  verbatim. DEC-010's "Revisit if users demand FTS5 operators"
  criterion.
- **Snippet / highlighting** (`snippet()` FTS5 function).
  Future polish.
- **Column-specific search** (`brag search "auth" --field tags`).
  Future polish.
- **Fuzzy / typo-tolerant search.** Would require a different
  FTS5 tokenizer (trigram) — out of scope.
- **Result count in stderr footer** (`3 matches`). Future polish
  and easy to add via a `--count` flag.
- **JSON output on search.** STAGE-003's `brag export --format
  json` will handle structured consumers. `brag search` stays
  line-based.
- **Updating `Store.List` to also filter by FTS5.** Separate
  axes of retrieval: `list` = filtered, `search` = ranked. Don't
  merge them.

## Notes for the Implementer

- **`buildFTS5Query` shape.**
  ```go
  // buildFTS5Query converts a user-typed search argument into an
  // FTS5 MATCH-compatible string per DEC-010. Returns an error
  // for empty, whitespace-only, or quote-containing input.
  func buildFTS5Query(raw string) (string, error) {
      if strings.ContainsRune(raw, '"') {
          return "", fmt.Errorf("search query must not contain quotes")
      }
      tokens := strings.Fields(raw)
      if len(tokens) == 0 {
          return "", fmt.Errorf("search query must not be empty")
      }
      parts := make([]string, len(tokens))
      for i, t := range tokens {
          parts[i] = `"` + t + `"`
      }
      return strings.Join(parts, " "), nil
  }
  ```

- **`runSearch` shape.**
  ```go
  func runSearch(cmd *cobra.Command, args []string) error {
      if len(args) != 1 {
          return UserErrorf("search requires exactly one query argument")
      }
      fts5, err := buildFTS5Query(args[0])
      if err != nil {
          return UserErrorf("%v", err)
      }
      limit, _ := cmd.Flags().GetInt("limit")
      if cmd.Flags().Changed("limit") && limit < 0 {
          return UserErrorf("invalid --limit %d: must be zero or positive", limit)
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
      entries, err := s.Search(fts5, limit)
      if err != nil {
          return fmt.Errorf("search: %w", err)
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

- **`Store.Search` SQL.**
  ```go
  const q = `
      SELECT e.id, e.title, e.description, e.tags, e.project,
             e.type, e.impact, e.created_at, e.updated_at
      FROM entries_fts
      JOIN entries e ON e.id = entries_fts.rowid
      WHERE entries_fts MATCH ?
      ORDER BY rank, e.id DESC
  `
  ```
  Append `" LIMIT ?"` and bind the limit parameter only when
  `limit > 0`. Iterate rows, scan into `Entry` using the same
  `sql.NullString` pattern as `Store.List` (SPEC-002 / SPEC-006).

- **Rank semantics.** FTS5's `rank` pseudo-column uses bm25 by
  default. Lower rank values = more relevant matches. `ORDER BY
  rank` (ascending) puts best matches first; `id DESC` breaks
  ties.

- **Search cobra shape.**
  ```go
  func NewSearchCmd() *cobra.Command {
      cmd := &cobra.Command{
          Use:   "search <query>",
          Short: "Search brag entries via FTS5",
          Long: `Search entries by content across title, description, tags, project, and impact.

  Query semantics (DEC-010):
    - Tokens are whitespace-separated
    - Each token is treated as a literal string (no FTS5 operators)
    - Multiple tokens: AND semantics (entries must contain all words)

  Examples:
    brag search "auth"                   # find entries mentioning auth
    brag search "cut latency"            # entries with both "cut" AND "latency"
    brag search "auth-refactor"          # literal match; hyphen is not NOT-operator
    brag search "redis" --limit 5        # top 5 matches`,
          RunE: runSearch,
      }
      cmd.Flags().Int("limit", 0, "cap result count (0 = unlimited)")
      return cmd
  }
  ```

- **`docs/api-contract.md` amendment.** The existing
  `brag search` section is stubbed from the project-design phase.
  Rewrite it with:
  - DEC-010 reference for query semantics.
  - Exit code 0 on success OR zero-results.
  - Exit code 1 on invalid query (empty, quote-containing,
    no-args, too-many-args) or bad `--limit`.
  - Exit code 2 on storage failure.
  - `--limit N` flag, 0 = unlimited.

- **`docs/tutorial.md` update.**
  1. §9 "What's NOT there yet": strike the `brag search` row.
  2. Add a short subsection under §4 (or near the filter-flags
     subsection) showing `brag search "latency"` and
     `brag search "auth-refactor"` with a brief note that
     multi-word queries are AND and hyphens are fine.

- **No `init()` functions** (§8).

---

## Build Completion

*Filled in at the end of the **build** cycle, before advancing to verify.*

- **Branch:**
- **PR (if applicable):**
- **All acceptance criteria met?** yes/no
- **New decisions emitted:**
  - (none expected; DEC-010 was emitted during this spec's design)
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
