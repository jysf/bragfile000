---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-073
  type: story                      # epic | story | task | bug | chore
  cycle: verify
  blocked: false
  priority: high
  complexity: M                    # S | M | L  (L means split it)

project:
  id: PROJ-006
  stage: STAGE-019
repo:
  id: bragfile

agents:
  architect: claude-opus-5
  implementer: claude-opus-5       # usually same Claude, different session
  created_at: 2026-08-08

references:
  decisions:
    - DEC-043                      # emitted by this spec — the blended ranking
    - DEC-044                      # emitted by this spec — the token budget + line shape
    - DEC-014                      # the rule-based output envelope memory renders into
    - DEC-027                      # the reserved token-count tag DEC-044 scopes away from
    - DEC-010                      # brag search query transform, reused for --query
    - DEC-011                      # JSON entry shape the per-item projection narrows
    - DEC-017                      # entries.project is a soft string match
    - DEC-031                      # the pure-primitive-package + markdown-only template
    - DEC-042                      # internal/timewindow — the shared-pure-package precedent
  constraints:
    - no-sql-in-cli-layer
    - test-before-implementation
    - errors-wrap-with-context
    - no-new-top-level-deps-without-decision
    - no-cgo
    - stdout-is-for-data-stderr-is-for-humans
    - one-spec-per-pr
  related_specs:
    - SPEC-072                     # shipped — MCP filter parity + internal/timewindow
    - SPEC-074                     # pending — MCP resources + brag_memory; consumes this
    - SPEC-052                     # shipped — internal/spark, the structural template
    - SPEC-059                     # shipped — brag spark, the 7th DEC-014 consumer
    - SPEC-011                     # shipped — the FTS5 index whose bm25 order is fused
    - SPEC-064                     # shipped — capture.Validate field caps
---

# SPEC-073: The memory slice — blended, token-budgeted retrieval

## Context

STAGE-019 turns the corpus from something an agent can query **if it thinks to**
into working memory it consults **before it acts**. SPEC-072 shipped the cheap half
(filter parity + the shared `internal/timewindow` parser). SPEC-074 ships the push
half (MCP resources). This spec is the middle one, and the only one in the stage
with a real algorithm in it.

The gap it closes: recency (`Store.List`) and relevance (`Store.Search`) are
**separate axes** today, and `Limit` is a **row count, not a token cost**. An agent
that wants "the history I should know before I start" has to pick one axis, guess a
row count, and hope the result fits its context window. None of those are things a
caller can get right from outside.

Two decisions carry the weight, both emitted with this spec:

- **DEC-043** — the ranking. The hard part is that the two orderings are
  *incomparable*: `List` returns an ordinal position in time; `Search` returns bm25
  `rank`, which is negative, unbounded, and corpus-relative. Reciprocal-rank fusion
  over three ordinal lists (recency, relevance, project) fuses the **positions**
  rather than inventing a shared scale, and degenerates to plain recency for free
  when no query is given — one algorithm for both the session-start auto-load and
  the targeted ask.
- **DEC-044** — the budget. A token count, a documented estimate, greedy
  skip-and-continue enforcement, and the one-line entry rendering that *is* the unit
  of cost. Its honesty clause scopes the estimate away from DEC-027's
  caller-reported provenance rule, mechanically.

SPEC-074 consumes what this spec builds: the same `memory.Slice`, the same pool
composition, and the same markdown body served as a resource. Getting the body right
here is what stops SPEC-074 inventing a second rendering.

## Goal

Ship `internal/memory` — a pure, deterministic package that ranks a candidate pool
of entries by DEC-043's fusion and trims it to DEC-044's token budget — and surface
it as `brag memory`, rendering under the DEC-014 envelope in markdown (default) and
`--format json`, with the markdown body pinned by byte-exact goldens.

## Inputs

- **Files to read (in this order):**
  - `decisions/DEC-043-memory-slice-blended-rank-fusion.md` — the ranking, its six
    sub-decisions, and the locked constants. Read first; do not re-litigate.
  - `decisions/DEC-044-memory-slice-token-budget-and-line-shape.md` — the budget,
    the estimator, the enforcement policy, the line shape, and the **honesty
    clause**. Read the honesty clause carefully before writing any doc comment in
    `internal/memory` (see the NOT-contains audit in Notes).
  - `decisions/DEC-014-rule-based-output-shape.md` — the envelope. Non-negotiable
    shape: single-object JSON, flat top-level keys, `Generated:`/`Scope:`/`Filters:`
    markdown provenance, empty-state rules, indent=2.
  - `internal/export/coverage.go` + `internal/export/spark.go` — the two closest
    renderer models (options struct with injected `Now`, `trimTrailingNewline`,
    empty-window early return, `<Thing>Options` naming, separate markdown/JSON
    entry points).
  - `internal/cli/spark.go` — the closest CLI model (a `Long` with a locked shape,
    `--format` validation, one `Store` read, `fmt.Fprintln(cmd.OutOrStdout(), ...)`).
  - `internal/cli/search.go` — `buildFTS5Query`, reused **verbatim** for `--query`
    (same package, so no extraction and no new copy).
  - `internal/storage/store.go` — `List` (`created_at DESC, id DESC`), `Search`
    (`ORDER BY rank, e.id DESC`), and `ListFilter`'s fields.
  - `internal/spark/spark.go` + `internal/timewindow/timewindow.go` — the two
    existing pure packages. `timewindow_test.go`'s `TestPackageReadsNoWallClock` is
    copied into `internal/memory` (see Failing Tests).
  - `internal/storage/storagetest/storagetest.go` — `Backdate(dbPath, id, at)`, the
    only sanctioned way for a non-storage test to seed past-dated rows. **The CLI
    golden test needs it** (SPEC-072's build reflection flagged that omitting this
    pointer cost a lookup; it is listed here on purpose).
- **External APIs:** none. No network, no new dependency, no tokenizer.
- **Related code paths:** `internal/aggregate/` (the sibling pure data package;
  `NoProjectKey` is deliberately **not** reused — DEC-044 sub-decision 5).

## Outputs

- **Files created:**
  - `internal/memory/memory.go` — the pure package. Exports `Options`, `Item`,
    `Result`, `Slice`, `RenderLine`, `EstimateTokens`, and the locked constants
    `FusionK`, `WeightRecency`, `WeightMatch`, `WeightProject`, `PoolLimit`,
    `DefaultBudget`, `CharsPerToken`. Imports **nothing outside the standard
    library** plus `internal/storage` (for the `Entry` type) — expected to be
    `fmt`, `sort`, `strings`; no SQL driver, no cobra, no third-party package.
  - `internal/memory/memory_test.go` — unit tests, the two mechanical guards, and
    the `Slice`-level goldens.
  - `internal/export/memory.go` — `MemoryOptions`, `ToMemoryMarkdown`,
    `ToMemoryJSON`.
  - `internal/export/memory_test.go` — the byte-exact markdown + JSON goldens.
  - `internal/cli/memory.go` — `NewMemoryCmd` + `runMemory` (the three-read pool
    composition).
  - `internal/cli/memory_test.go` — CLI tests incl. the end-to-end golden over a
    real store.
  - `decisions/DEC-043-memory-slice-blended-rank-fusion.md` — **already written at
    design** (emitted with this spec).
  - `decisions/DEC-044-memory-slice-token-budget-and-line-shape.md` — **already
    written at design**.
- **Files modified:**
  - `cmd/brag/main.go` — `root.AddCommand(cli.NewMemoryCmd())`, inserted after
    `NewStoryCmd()` (the digest family, newest last).
  - `docs/api-contract.md` — a new `### brag memory …` section inserted **after**
    the `brag story` section and **before** `### brag tags` (line ~853), claiming the
    **EIGHTH** DEC-014 consumer ordinal; plus two DEC-index rows for DEC-043/DEC-044.
  - `AGENTS.md` §11 — a new `memory` glossary entry (see premise audit, additive
    case).
  - `README.md` — one line in the digest command block (currently lines ~146–149).
  - `docs/tutorial.md` — a new subsection after `### A quick pulse: brag spark`.
  - `docs/architecture.md` — one row in the §Responsibilities package table for
    `internal/memory`. **Only that row** — see premise audit for what deliberately
    stays.
  - `scripts/test-docs.sh` — a new **Group U** (`brag memory` docs), appended after
    Group T and before `# ===== finalise =====`.
  - `guidance/questions.yaml` — one new question `memory-slice-fusion-constants`
    (DEC-043 is 0.75; §14 logs the constants question).
  - `projects/PROJ-006-agent-native-depth-core/stages/STAGE-019-corpus-as-agent-memory.md`
    — backlog status line for SPEC-073.
- **New exports:** `memory.Slice`, `memory.RenderLine`, `memory.EstimateTokens`,
  `memory.Options`, `memory.Item`, `memory.Result`, the seven constants;
  `export.ToMemoryMarkdown`, `export.ToMemoryJSON`, `export.MemoryOptions`;
  `cli.NewMemoryCmd`.
- **Database changes:** **none.** No migration, no schema change, no new storage
  method, no new SQL. Three existing `Store` methods composed in Go.

### Premise audit — run at design, hits reconciled against the list above

Per `projects/_templates/premise-audit.md`. Every grep below was **executed**
against the repo at design; the "hits" column is actual output, not expectation.

| case | grep | actual hits | disposition |
|---|---|---|---|
| **additive** (ordinal chain) | `grep -rn "DEC-014 consumer" docs/ AGENTS.md README.md BRAG.md internal/` | `docs/api-contract.md:444` (4th), `:522` (5th), `:612` (**sixth**), `:692` (**seventh**), `:1296`+`:1297` (DEC index); `AGENTS.md:297` (5th), `:300` (SIXTH); `internal/cli/{summary,review,stats,impact,wrapped,coverage,spark}.go`; `internal/export/{impact,wrapped,coverage,spark}.go` | The chain is live and asserted in prose. `brag memory` is the **EIGHTH**. Planned: claim the ordinal in the new `api-contract.md` section, in `internal/cli/memory.go`'s doc comment, and in `internal/export/memory.go`'s doc comment (matching the existing three-place pattern). **No existing ordinal changes** — this appends. |
| **additive** (glossary) | same grep, `AGENTS.md` hits | §11 carries entries for `wrapped`, `coverage`, `sparkline`, `aggregate`, … but **none** for `brag spark` or `brag story` | Planned: **add a `memory` §11 entry.** Deliberately **not** back-filling the missing `spark`/`story` entries — that is a docs chore, not this spec (flagged under Follow-up, not silently widened). |
| **additive** (subcommand count) | `grep -rn "AddCommand\|\.Commands()" internal/ cmd/ scripts/ docs/` | 68 hits, all per-test single registrations, plus `cmd/brag/main.go:27-46`. The one loop, `internal/cli/list_test.go:843`, iterates `root.Commands()` **only to find the `list` subcommand by name** — no count coupling. `internal/cli/root_test.go` has no count assertion (re-verified). | **No count bump anywhere.** Only `cmd/brag/main.go` gains a line. |
| **additive** (MCP tool count) | `grep -rniE "brag_memory\|internal/memory\|memory slice" --include='*.go' --include='*.md' …` | one hit: `internal/mcpserver/server_test.go:87` — a SPEC-072 comment naming **SPEC-074's** `brag_memory` and the six places that would need bumping | **Stays.** SPEC-073 adds no MCP tool and no MCP surface; that enumeration is SPEC-074's to consume. Recorded so build does not "helpfully" touch it. |
| **additive** (docs package table) | `grep -rn "internal/spark\|internal/timewindow\|internal/aggregate" docs/ AGENTS.md README.md scripts/test-docs.sh` | `docs/architecture.md:84` (a maintained `Responsibilities` table listing 8 packages), `:40` (a mermaid diagram), `docs/blog/how-brag-was-built.md:123` (a historical copy of that diagram), `AGENTS.md:286`/`:298` | Planned: **one new row** in `architecture.md`'s table for `internal/memory` (its sibling `internal/aggregate` has one). **Stays:** the mermaid diagram at `:40` (it depicts the STAGE-004 digest path; adding every later package needs a redraw), the blog copy at `:123` (a dated artifact), and the table's pre-existing omission of `spark`/`timewindow`/`story`/`capture` (a separate chore — do **not** widen). |
| **status change** | `grep -rniE "brag memory\|internal/memory\|memory slice\|brag_memory" .` (excluding `projects/`, `decisions/DEC-04[34]`) | the single `server_test.go:87` hit above | Nothing claims `brag memory` does not exist, so there is no "not yet shipped" row to strike. Purely additive. `docs/for-ai-agents.md` is **untouched** — it documents the four MCP tools; `brag memory` is CLI-only until SPEC-074. |
| **inversion / removal** | n/a | — | Nothing is inverted or removed. No existing test's premise changes; no planned deletions or rewrites. |
| **NOT-contains** | see the self-audit table in Notes for the Implementer | zero hits in load-bearing prose | Two help-output absence assertions (`--limit`, `--detail`) and one package-source absence guard, each grep'd against the locked literals at design. |

```
Premise audit (projects/_templates/premise-audit.md), run at design:
- [x] Inversion/removal: greps run — none apply (purely additive spec)
- [x] Addition/count-bump: greps run — no literal-count assertion is coupled
- [x] Status-change: greps run — one hit, dispositioned "stays" (SPEC-074's)
- [x] Cross-check: actual grep hits reconciled against ## Outputs above
```

## Acceptance Criteria

- [x] `internal/memory` imports nothing outside the standard library plus
      `github.com/jysf/bragfile000/internal/storage`. No driver, no SQL, no cobra,
      no third-party package; `go.mod` untouched.
- [x] `memory.Slice(entries, opts)` is a pure function of its arguments: it reads no
      clock, no env, no filesystem, and takes **no `now`** (DEC-043 sub-decision 3).
- [x] With `Options.Matched` empty and `Options.Project` empty, `Slice`'s output
      order is **exactly** `created_at DESC, id DESC` — identical to `Store.List`.
- [x] With a query and a project, the output order differs from **both** plain
      recency and plain bm25 over the shared fixture (STAGE-019 success criterion).
- [x] `Slice` dedupes by ID and is invariant to input order: the same entries in any
      order, with duplicates, produce a byte-identical rendering.
- [x] The locked constants are exported and load-bearing: `FusionK == 60`,
      `WeightRecency == WeightMatch == WeightProject == 1.0`, `PoolLimit == 200`,
      `DefaultBudget == 2000`, `CharsPerToken == 4`. Changing any one breaks a
      golden.
- [x] `RenderLine` emits DEC-044 sub-decision 5's shape exactly, including `-` for an
      absent project/type and no trailing em-dash when `Impact` is empty.
- [x] `EstimateTokens(s) == ceil(len(s)/4)` over UTF-8 **bytes**, per rendered unit.
- [x] `Result.EstimatedTokens == Σ EstimateTokens(item.Line)` over the included set —
      the estimate and the render cannot disagree.
- [x] `Result.Included + Result.Skipped == Result.Candidates`.
- [x] Enforcement is **skip-and-continue**: an entry that does not fit is skipped and
      the fill continues, so a later cheaper entry can still be included.
- [x] `brag memory` renders the DEC-014 envelope: `# Bragfile Memory`, `Generated:`,
      `Scope: lifetime`, `Filters:`, `Entries: N`, then `## Slice` and `## Budget`.
      Markdown is byte-identical to the goldens in Notes.
- [x] On an **empty** candidate pool the markdown ends after `Entries: 0` (DEC-014
      part 4); JSON still emits every key with `0`/`[]`/`{}`.
- [x] On a **non-empty** pool with zero included entries, both body sections still
      render (`## Slice` empty, `## Budget` reporting `Included: 0`).
- [x] `--format json` and `--format markdown` select the **same ids in the same
      order** for the same options (the budget is always measured against markdown).
- [x] `--budget 0` and any negative `--budget` are a `UserError` (exit 1, stdout
      empty, message on stderr).
- [x] An invalid `--query` (empty after tokenization, or containing a quote) and an
      unknown `--format` value are `UserError`s.
- [x] Help output contains `reciprocal-rank fusion`, `soft boost, not a filter`,
      `not a tokenizer`, and `Examples:`; and contains neither `--limit` nor
      `--detail`.
- [x] No non-test file in `internal/memory` contains a quoted reserved-namespace
      prefix literal (the DEC-044 honesty guard), and none calls `time.Now(`.
- [x] Full gate set green: `gofmt -l .` empty, `go vet ./...`, `go test ./...`,
      `just test-docs`.

## Failing Tests

Written during **design**, BEFORE build. The implementer's job in **build** is to
make these pass. Every locked sub-decision in DEC-043/DEC-044 has at least one test
below that fails without it (AGENTS.md §9 decision↔test traceability).

### `internal/memory/memory_test.go`

Uses the **shared fixture** transcribed verbatim from Notes → *The fixture*.

- `"TestSlice_DegeneratesToRecencyWithNoQueryOrProject"` — `Options{Budget: 2000}`;
  asserts `Items` ids are `[8 7 6 5 4 3 2 1]`, i.e. exactly `created_at DESC, id
  DESC`. **Pins DEC-043's "one algorithm, both cases."**
- `"TestSlice_BlendDiffersFromBothInputs"` — `Options{Query:"auth", Project:"orbit",
  Matched: []int64{5,1,2,4}, Budget: 2000}`; asserts ids are `[4 2 7 5 1 8 6 3]` AND
  that this differs from `[8 7 6 5 4 3 2 1]` (plain recency) AND from a prefix
  ordering of `[5 1 2 4]` (plain bm25). **Pins the stage success criterion.**
- `"TestSlice_ScoresMatchTheLockedConstants"` — asserts each item's `Score` equals
  the hand-computable sum to 9 decimal places, using the table in Notes →
  *§12(b) pre-flight → A*. **Pins `FusionK` and the three weights**; changing any
  constant fails here first, with a readable diff.
- `"TestSlice_AbsentFromAListContributesZero"` — an entry in no `Matched` and no
  project scores exactly `1/(60+rank_recency)`.
- `"TestSlice_MatchedIgnoresUnknownAndDuplicateIDs"` — `Matched:
  []int64{999, 5, 5, 1}` produces the same ranking as `Matched: []int64{5, 1}`.
- `"TestSlice_ProjectListIsRecencyOrdered"` — with `Project:"orbit"` and no query,
  asserts `7` (project rank 1) outranks `4` (project rank 2) outranks `2`, and that
  all three outrank every non-orbit entry.
- `"TestSlice_UnknownProjectAddsNoTerm"` — `Project:"nope"` yields the identical
  ordering to `Project:""`.
- `"TestSlice_InputOrderInvariantAndDedupes"` — the fixture shuffled, with entry `8`
  appended a second time, yields identical `Items` and `Candidates == 8`.
- `"TestSlice_EqualScoresTieBreakByRecencyThenID"` — two entries with identical
  `CreatedAt` and no match/project terms order by `id DESC`.
- `"TestRenderLine_ShapeAndAbsentFields"` — table over the fixture asserting the four
  bracket forms (`[orbit/shipped]`, `[-/-]`, `[-/learned]`, `[bragfile/shipped]`),
  the UTC `YYYY-MM-DD` date, and that an empty `Impact` emits **no** trailing
  ` — `. Expected strings transcribed from Notes → *Per-entry token costs*.
- `"TestEstimateTokens_CeilOverBytes"` — table: `""`→0, `"a"`→1, `"abcd"`→1,
  `"abcde"`→2, and a multi-byte string asserting **byte** semantics (a 3-byte rune
  counts as 3, not 1).
- `"TestSlice_EstimateMatchesRenderedBody"` — `Result.EstimatedTokens` equals
  `Σ EstimateTokens(RenderLine(item.Entry))` over the included set, and each
  `Item.Tokens` equals `EstimateTokens(Item.Line)`. **Pins "the estimate and the
  render cannot disagree."**
- `"TestSlice_CountsPartitionTheCandidatePool"` — `Included + Skipped == Candidates`
  across three budgets (2000, 40, 1).
- `"TestSlice_SkipsOversizeAndContinues"` — `Budget: 40`, no query/project; asserts
  ids `[8 5]`, `Included == 2`, `Skipped == 6`, `EstimatedTokens == 37`. **Pins
  DEC-044's skip-and-continue against stop-at-first-overflow**, which would give
  `[8]` / `1` / `27` — the test fails under the rejected policy.
- `"TestSlice_BudgetTooSmallForAnyEntryIncludesNothing"` — `Budget: 1` yields
  `Included == 0`, `Skipped == 8`, `EstimatedTokens == 0`, and a non-nil empty
  `Items`.
- `"TestSlice_EmptyPool"` — `Slice(nil, opts)` returns `Candidates == 0`,
  `Included == 0`, `Skipped == 0`, `Items` non-nil and empty.
- `"TestPackageReadsNoWallClock"` — **copied from
  `internal/timewindow/timewindow_test.go`** (same `filepath.WalkDir` idiom): walks
  the package's non-test `.go` files and fails on `time.Now(`. Message cites
  DEC-043 sub-decision 3.
- `"TestPackageEmitsNoReservedTagNamespace"` — the DEC-044 honesty guard. Walks the
  package's non-test `.go` files and fails on any of the five **quoted** literals
  `"agent:"`, `"model:"`, `"session:"`, `"cost:"`, `"tokens:"` (each including its
  surrounding double-quote characters). Verbatim body locked in Notes.

### `internal/export/memory_test.go`

- `"TestToMemoryMarkdown_BlendedGolden"` — byte-exact equality against **Golden 1**
  (Notes). The load-bearing contract test.
- `"TestToMemoryMarkdown_RecencyGolden"` — byte-exact against **Golden 2** (no
  query, no project).
- `"TestToMemoryMarkdown_TightBudgetGolden"` — byte-exact against **Golden 3**
  (`--budget 40`), proving skip-and-continue at the rendering layer.
- `"TestToMemoryMarkdown_EmptyCorpusGolden"` — byte-exact against **Golden 4**: the
  document ends after `Entries: 0`, with **no** `## Slice` and **no** `## Budget`
  (DEC-014 part 4).
- `"TestToMemoryMarkdown_ZeroIncludedStillRendersBothSections"` — `Budget: 1` over
  the non-empty fixture: `## Slice` present with no bullets, `## Budget` present
  reporting `Included: 0` / `Skipped: 8`. Asserts line-by-line equality
  (`ln == "## Slice"`), not `strings.Contains` — every deeper heading is a
  superstring of a shallower one (AGENTS.md §9 heading addendum).
- `"TestToMemoryJSON_BlendedGolden"` — byte-exact against the **JSON golden**
  (Notes), including `score` rounded to 6 dp and the flat DEC-014 key order.
- `"TestToMemoryJSON_EmptyCorpusEmitsEveryKey"` — `entries`/`budget`/
  `estimated_tokens`/`included`/`skipped` are `0`, `slice` is `[]` (non-nil, never
  `null`), `filters` is `{}` (DEC-014 part 4).
- `"TestToMemoryJSON_FiltersEchoQueryAndProject"` — `filters` is
  `{"project":"orbit","query":"auth"}` (alphabetical, per Go's encoder) while the
  markdown line is `--query auth --project orbit` (declared order) — the documented
  DEC-014 markdown/JSON ordering asymmetry.
- `"TestToMemoryMarkdown_TrailingNewlineStripped"` — the returned bytes do not end
  in `\n` (matches every sibling renderer; the CLI adds exactly one via `Fprintln`).

### `internal/cli/memory_test.go`

Seeds a real store in `t.TempDir()`, then `storagetest.Backdate`s each row to the
fixture's `created_at`.

- `"TestMemoryCmd_EndToEndMarkdownGolden"` — seeds the fixture, runs
  `memory --query auth --project orbit` with `nowFunc` stubbed to
  `2026-08-08T12:00:00Z`, and asserts stdout equals **Golden 1** plus one trailing
  newline. **This is the test that proves the declared `Matched` order is real**:
  the CLI does not receive `[5 1 2 4]`, it derives it from FTS5. Verified at design
  (Notes → §12(b) pre-flight → B).
- `"TestMemoryCmd_BareInvocationIsPlainRecency"` — `brag memory` with no flags
  produces **Golden 2**, and `Filters: (none)`.
- `"TestMemoryCmd_JSONAndMarkdownSelectTheSameSlice"` — runs both formats with the
  same options and asserts the ids **and their order** are identical. **Pins
  DEC-044 sub-decision 6**; fails if the budget is ever measured against the JSON
  bytes.
- `"TestMemoryCmd_ThreeReadsComposeThePool"` — the only test that needs a big
  corpus, and the one that proves DEC-043 sub-decision 5's **"one bounded read per
  list."** Seed **202** entries into one store:
  - ids `1..200` — filler, titles `filler NNN`, no project, left at `Store.Add`'s
    own timestamp. They share a second, so `created_at DESC, id DESC` orders them
    by `id DESC` deterministically — **no `Backdate` needed for these**, which
    keeps the test fast (`Backdate` opens its own connection per call).
  - id `201` — `Project: "orbit"`, title `ancient orbit note`, `Backdate`d to
    `2020-01-01T00:00:00Z`.
  - id `202` — no project, title `ancient auth note`, `Backdate`d to
    `2020-01-02T00:00:00Z`.

  Both ancient rows sit at recency rank 201/202, i.e. **outside `PoolLimit` (200)**,
  so each is reachable only through its own list's read. Three assertions, all with
  `--budget 20000` (large enough that the budget never confounds the result):
  - `brag memory --project orbit` **includes `201`** → the project read exists.
  - `brag memory --query auth` **includes `202`** → the match read exists.
  - `brag memory` (bare) **includes neither** → the recency read really is capped
    at `PoolLimit`, so the other two assertions are not passing by accident.
- `"TestMemoryCmd_ZeroBudgetIsUserError"` — `--budget 0` returns an error wrapping
  `ErrUser`; `outBuf` is empty and the message names the flag.
- `"TestMemoryCmd_NegativeBudgetIsUserError"` — same for `--budget -1`.
- `"TestMemoryCmd_UnknownFormatIsUserError"` — `--format yaml` errors, naming
  `markdown` and `json`.
- `"TestMemoryCmd_EmptyQueryIsUserError"` — `--query "   "` errors via
  `buildFTS5Query` (DEC-010), stdout empty.
- `"TestMemoryCmd_QuoteInQueryIsUserError"` — `--query 'a"b'` errors (DEC-010).
- `"TestMemoryCmd_QueryWithNoMatchesStillRenders"` — `--query zzzznomatch` produces a
  pure-recency slice **and** still echoes `Filters: --query zzzznomatch`. Honest
  degradation, not a silent empty result.
- `"TestMemoryCmd_ProjectIsSoftBoostNotFilter"` — `--project orbit` over the fixture
  returns **all 8** entries (`Entries: 8`), with the orbit entries leading. **Pins
  DEC-043 sub-decision 4**; fails if `ListFilter.Project` is ever used as a filter
  on the recency read.
- `"TestMemoryCmd_HelpShape"` — help contains `reciprocal-rank fusion`,
  `soft boost, not a filter`, `not a tokenizer`, `Examples:`; and does **NOT**
  contain `--limit` or `--detail`. `errBuf.Len() == 0` (AGENTS.md §9
  separate-buffers rule).
- `"TestMemoryCmd_StdoutOnlyCarriesTheDocument"` — on success `errBuf.Len() == 0`;
  on a `UserError` `outBuf.Len() == 0`
  (`stdout-is-for-data-stderr-is-for-humans`).

### `scripts/test-docs.sh` — new Group U

- `U1` — `docs/api-contract.md` contains `### \`brag memory` .
- `U2` — `docs/api-contract.md` contains `**eighth** DEC-014 consumer` (the ordinal
  claim, asserted so the chain cannot silently skip a number).
- `U3` — `docs/api-contract.md` contains `DEC-043` and `DEC-044`.
- `U4` — `README.md` fenced blocks contain `brag memory`.
- `U5` — `docs/tutorial.md` contains `brag memory`.
- `U6` — `AGENTS.md` contains the glossary phrase `- **memory** —`.
- `U7` — `docs/api-contract.md` contains `never stamped on an entry` (the DEC-044
  honesty scoping is user-visible, not only internal).
- `U8` — `docs/architecture.md` contains `internal/memory`.

## Implementation Context

*Read this section (and the files it points to) before starting the build cycle.*

### Decisions that apply

- `DEC-043` — **the ranking spine.** Three ordinal lists, RRF, `k=60`, unit weights,
  ordinal (clock-free) recency, soft project boost, one bounded pool per list,
  dedupe + stable tie-break. Consume verbatim; the alternatives are already weighed.
- `DEC-044` — **the budget spine.** Token budget with a positive-only `--budget`
  (default 2000), `ceil(bytes/4)` estimate, skip-and-continue, budget-covers-the-
  entry-lines, budget-always-measured-against-markdown, the locked line shape, and
  the **honesty clause**. Consume verbatim.
- `DEC-014` — the envelope. `# <Doc Title>` → `Generated:` → `Scope:` → `Filters:`
  → body; single-object JSON with flat top-level keys; `filters` always an object;
  arrays never `null`; indent=2; body sections omitted on an empty entry set.
  `Scope` is **`lifetime`** here (memory ranks the whole corpus, like `stats`), which
  keeps DEC-014's own validation clause true.
- `DEC-027` — the reserved token-count tag. **Read it before writing any comment in
  `internal/memory`.** DEC-044's honesty clause scopes this estimate away from it;
  the guard test enforces the boundary.
- `DEC-010` — the `brag search` query transform. `--query` reuses
  `cli.buildFTS5Query` **as-is** (same package). Do **not** extract it; see Out of
  scope.
- `DEC-011` — the 9-key entry shape. `memory`'s per-item JSON is a deliberate
  **narrowing** (9 keys, different set — no `description`, no `tags`, no
  `updated_at`, plus `rank`/`score`/`tokens`), the same move `impact` made with its
  4-key projection.
- `DEC-017` — `entries.project` is a soft string match. This is *why* the project
  term is a boost and not a filter.
- `DEC-031` — the structural template: a small pure primitive package plus a
  markdown-only rendering rule with JSON staying raw. `internal/memory` follows it.
- `DEC-042` — the shared-pure-package precedent and the source of the
  `TestPackageReadsNoWallClock` idiom.

### Constraints that apply

- `no-sql-in-cli-layer` — `internal/cli/memory.go` goes through `*storage.Store`
  only. `internal/memory` imports `internal/storage` **for the `Entry` type**, never
  a driver. The CLI test reaches raw SQL only through
  `internal/storage/storagetest`.
- `test-before-implementation` — write the Failing Tests first; run `go test ./...`
  once and confirm each fails for the **expected** reason. See the fail-first note
  in Notes.
- `errors-wrap-with-context` — `fmt.Errorf("render memory: %w", err)` /
  `fmt.Errorf("list entries: %w", err)`, matching `runCoverage`/`runSpark`.
- `no-new-top-level-deps-without-decision` / `no-cgo` — `go.mod` must be
  **untouched**. No tokenizer, no ranking library.
- `stdout-is-for-data-stderr-is-for-humans` — the document to stdout; `UserError`s
  to stderr via `main.go`'s existing mapping.
- `one-spec-per-pr` — branch `feat/spec-073-memory-slice`, one PR.

### Prior related work

- `SPEC-072` (shipped) — `internal/timewindow`, the filter parity, and the
  `TestPackageReadsNoWallClock` guard. Its build reflection asked for
  `storagetest.Backdate` to be named in Inputs; it is.
- `SPEC-059` (shipped) — `brag spark`, the seventh DEC-014 consumer. The closest CLI
  + renderer pair to copy structurally.
- `SPEC-052` (shipped) — `internal/spark`, the pure-primitive-package precedent.
- `SPEC-045` (shipped) — `brag coverage`; `export/coverage.go` is the cleanest
  renderer to model.
- `SPEC-064` (shipped) — `internal/capture.Validate`'s field caps, which bound a
  rendered line at 626 bytes / 157 tokens.

### Out of scope (for this spec specifically)

- **Anything MCP.** No resources, no `brag_memory` tool, no `mcpserver` edit, no
  `docs/for-ai-agents.md` change. All SPEC-074 / DEC-045.
- **Extracting `buildFTS5Query`.** `brag memory` lives in `internal/cli`, so it
  reuses the existing unexported helper — this is **not** a third copy and does not
  trip DEC-024/DEC-042's third-consumer trigger. SPEC-074 will be the third
  consumer; leave the extraction to it.
- **Time windows on `brag memory`** (`--since`/`--day`/`--until`). The blend already
  prefers recent, windows are `brag list`'s job, and a window would make
  `Scope: lifetime` a lie. YAGNI.
- **`--tag` / `--type` / `--author` filters.** A fourth or fifth ranked list is
  DEC-043 revisit (c), not this spec.
- **A `--detail` / `--verbose` line shape**, and any per-line truncation. DEC-044
  Option E / Option I, both rejected.
- **A `--project-only` hard-filter flag.** DEC-043 Option E, rejected with a
  pre-authorized reversal.
- **Tuning `k` or the weights.** Locked at published defaults; the question is filed
  in `guidance/questions.yaml` with its trigger.
- **Back-filling `AGENTS.md` §11 entries for `brag spark` / `brag story`**, and
  back-filling `docs/architecture.md`'s package table for
  `spark`/`timewindow`/`story`/`capture`. Real gaps, surfaced by this spec's premise
  audit, deliberately **not** bundled — record under Follow-up work.
- **The go-sdk `1.6.1 → 1.7.0` bump (PR #124).** Not needed here (this spec touches
  no MCP code) but it **gates SPEC-074**: DEC-045's §12(b) resources-API pre-flight
  must run against whatever version is in `go.mod` at that point. Flagged, not
  actioned.

## Notes for the Implementer

### The fixture (shared by all three test files — transcribe verbatim)

Eight entries. Chosen so every rendering branch and every ranking branch is
exercised: both-fields-present, no-project, no-type, no-project-and-no-type,
impact-present, impact-absent, in-both-lists, in-match-only, in-project-only.

| id | created_at | project | type | title | impact |
|---|---|---|---|---|---|
| 8 | `2026-08-07T09:15:00Z` | `bragfile` | `shipped` | `MCP list filter parity` | `agents can ask for a bounded window in one call` |
| 7 | `2026-08-05T17:40:00Z` | `orbit` | `learned` | `Retry storms need jitter` | *(empty)* |
| 6 | `2026-08-01T11:02:00Z` | `bragfile` | `shipped` | `Sparkline pulse command` | *(empty)* |
| 5 | `2026-07-20T08:00:00Z` | *(empty)* | *(empty)* | `Read the auth spec` | *(empty)* |
| 4 | `2026-06-11T14:25:00Z` | `orbit` | `shipped` | `Auth refactor for the gateway` | `cut p99 login latency from 600ms to 120ms` |
| 3 | `2026-05-02T19:30:00Z` | `bragfile` | `learned` | `FTS5 triggers fight attach` | *(empty)* |
| 2 | `2026-03-14T10:10:00Z` | `orbit` | `shipped` | `Auth token rotation` | `removed the last shared secret` |
| 1 | `2026-01-09T21:45:00Z` | *(empty)* | `learned` | `Auth is mostly caching` | *(empty)* |

`Now` for every golden: `2026-08-08T12:00:00Z`.

### §12(b) design-time pre-flight — what was run, and what it caught

Both halves of the §12(b) rule were executed at design (shape **and** behavior).

**A. The fusion arithmetic**, computed by a scratch implementation of the locked
algorithm. This table is the source of `TestSlice_ScoresMatchTheLockedConstants`'s
expected values — transcribe it, do not recompute by hand:

```
query="auth"  project="orbit"  Matched=[5 1 2 4]   k=60  Wr=Wm=Wp=1.0

id   r_rec  r_mat  r_prj  score
8    1      -      -      0.016393443
7    2      -      1      0.032522475
6    3      -      -      0.015873016
5    4      1      -      0.032018443
4    5      4      2      0.047138648
3    6      -      -      0.015151515
2    7      3      3      0.046671405
1    8      2      -      0.030834915

plain recency : 8 7 6 5 4 3 2 1
plain bm25    : 5 1 2 4
blended       : 4 2 7 5 1 8 6 3      <- differs from BOTH
```

**B. The real bm25 order — and the error this caught.** The `Matched` list above is
**not** a guess. Design first *assumed* `[1 4 5 2]`, then ran the fixture through a
real `storage.Store` + FTS5 (a throwaway test in `internal/storage`, deleted after)
and observed:

```
Search("auth") -> 4 rows
  bm25 rank 1 -> id 5  "Read the auth spec"
  bm25 rank 2 -> id 1  "Auth is mostly caching"
  bm25 rank 3 -> id 2  "Auth token rotation"
  bm25 rank 4 -> id 4  "Auth refactor for the gateway"
```

The real order is `[5 1 2 4]` — bm25 favours the **shortest** document and penalises
the longest, which is the opposite of the assumption for the top slot. The final
*ordering* happened to be unchanged, but every `score` value in the JSON golden
would have been wrong. This is exactly the §12(b) "target the behavioral surface,
not the shape validator" case: the pure function's own arithmetic validated fine
against a made-up input.

*(It also makes the nicest argument for the blend: bm25's top hit for "auth" is a
title-only entry with no project and no impact, while the entry that actually
matters — `4`, an auth refactor on orbit with a real impact line — is bm25's
**last**. The project and recency terms pull it back to rank 1.)*

**C. Per-entry token costs** (source for `TestRenderLine_*` and every budget test):

```
id=8 bytes=108 tokens=27  - 8 2026-08-07 [bragfile/shipped] MCP list filter parity — agents can ask for a bounded window in one call
id=7 bytes= 55 tokens=14  - 7 2026-08-05 [orbit/learned] Retry storms need jitter
id=6 bytes= 57 tokens=15  - 6 2026-08-01 [bragfile/shipped] Sparkline pulse command
id=5 bytes= 39 tokens=10  - 5 2026-07-20 [-/-] Read the auth spec
id=4 bytes=106 tokens=27  - 4 2026-06-11 [orbit/shipped] Auth refactor for the gateway — cut p99 login latency from 600ms to 120ms
id=3 bytes= 60 tokens=15  - 3 2026-05-02 [bragfile/learned] FTS5 triggers fight attach
id=2 bytes= 85 tokens=22  - 2 2026-03-14 [orbit/shipped] Auth token rotation — removed the last shared secret
id=1 bytes= 49 tokens=13  - 1 2026-01-09 [-/learned] Auth is mostly caching
```

Worst case under `capture.Validate`'s caps: **626 bytes / 157 tokens** per line.

**D. Invariants, verified on the scratch implementation before locking:**
`Included+Skipped == Candidates` ✓; `EstimatedTokens == Σ EstimateTokens(line)` ✓
(143 == 143); shuffled input with a duplicate id → byte-identical output,
`Candidates == 8` ✓.

### GOLDEN 1 — `brag memory --query auth --project orbit` (markdown, budget 2000)

```
# Bragfile Memory

Generated: 2026-08-08T12:00:00Z
Scope: lifetime
Filters: --query auth --project orbit
Entries: 8

## Slice

- 4 2026-06-11 [orbit/shipped] Auth refactor for the gateway — cut p99 login latency from 600ms to 120ms
- 2 2026-03-14 [orbit/shipped] Auth token rotation — removed the last shared secret
- 7 2026-08-05 [orbit/learned] Retry storms need jitter
- 5 2026-07-20 [-/-] Read the auth spec
- 1 2026-01-09 [-/learned] Auth is mostly caching
- 8 2026-08-07 [bragfile/shipped] MCP list filter parity — agents can ask for a bounded window in one call
- 6 2026-08-01 [bragfile/shipped] Sparkline pulse command
- 3 2026-05-02 [bragfile/learned] FTS5 triggers fight attach

## Budget

- Budget: 2000 tokens
- Estimated: 143 tokens
- Included: 8
- Skipped: 0
```

(778 bytes; the renderer returns this with the trailing newline stripped.)

### GOLDEN 2 — `brag memory` (no query, no project — the session opener)

```
# Bragfile Memory

Generated: 2026-08-08T12:00:00Z
Scope: lifetime
Filters: (none)
Entries: 8

## Slice

- 8 2026-08-07 [bragfile/shipped] MCP list filter parity — agents can ask for a bounded window in one call
- 7 2026-08-05 [orbit/learned] Retry storms need jitter
- 6 2026-08-01 [bragfile/shipped] Sparkline pulse command
- 5 2026-07-20 [-/-] Read the auth spec
- 4 2026-06-11 [orbit/shipped] Auth refactor for the gateway — cut p99 login latency from 600ms to 120ms
- 3 2026-05-02 [bragfile/learned] FTS5 triggers fight attach
- 2 2026-03-14 [orbit/shipped] Auth token rotation — removed the last shared secret
- 1 2026-01-09 [-/learned] Auth is mostly caching

## Budget

- Budget: 2000 tokens
- Estimated: 143 tokens
- Included: 8
- Skipped: 0
```

### GOLDEN 3 — `brag memory --budget 40` (skip-and-continue)

```
# Bragfile Memory

Generated: 2026-08-08T12:00:00Z
Scope: lifetime
Filters: (none)
Entries: 8

## Slice

- 8 2026-08-07 [bragfile/shipped] MCP list filter parity — agents can ask for a bounded window in one call
- 5 2026-07-20 [-/-] Read the auth spec

## Budget

- Budget: 40 tokens
- Estimated: 37 tokens
- Included: 2
- Skipped: 6
```

The trace: `8` costs 27, leaving 13; `7` (14) does not fit → skipped; `6` (15) does
not fit → skipped; `5` (10) **fits** → included, leaving 3; nothing else fits.
Rank 4 lands in the slice while ranks 2 and 3 do not — DEC-044 sub-decision 3's
accepted consequence, made visible. Stop-at-first-overflow would print one entry and
37% of the budget would go unused.

### GOLDEN 4 — empty corpus (DEC-014 part 4)

```
# Bragfile Memory

Generated: 2026-08-08T12:00:00Z
Scope: lifetime
Filters: (none)
Entries: 0
```

Nothing after `Entries: 0` — no `## Slice`, no `## Budget`.

### JSON GOLDEN — same options as Golden 1

```json
{
  "generated_at": "2026-08-08T12:00:00Z",
  "scope": "lifetime",
  "filters": {
    "project": "orbit",
    "query": "auth"
  },
  "entries": 8,
  "budget": 2000,
  "estimated_tokens": 143,
  "included": 8,
  "skipped": 0,
  "slice": [
    {
      "id": 4,
      "created_at": "2026-06-11T14:25:00Z",
      "project": "orbit",
      "type": "shipped",
      "title": "Auth refactor for the gateway",
      "impact": "cut p99 login latency from 600ms to 120ms",
      "rank": 1,
      "score": 0.047139,
      "tokens": 27
    },
    {
      "id": 2,
      "created_at": "2026-03-14T10:10:00Z",
      "project": "orbit",
      "type": "shipped",
      "title": "Auth token rotation",
      "impact": "removed the last shared secret",
      "rank": 2,
      "score": 0.046671,
      "tokens": 22
    },
    {
      "id": 7,
      "created_at": "2026-08-05T17:40:00Z",
      "project": "orbit",
      "type": "learned",
      "title": "Retry storms need jitter",
      "impact": "",
      "rank": 3,
      "score": 0.032522,
      "tokens": 14
    },
    {
      "id": 5,
      "created_at": "2026-07-20T08:00:00Z",
      "project": "",
      "type": "",
      "title": "Read the auth spec",
      "impact": "",
      "rank": 4,
      "score": 0.032018,
      "tokens": 10
    },
    {
      "id": 1,
      "created_at": "2026-01-09T21:45:00Z",
      "project": "",
      "type": "learned",
      "title": "Auth is mostly caching",
      "impact": "",
      "rank": 5,
      "score": 0.030835,
      "tokens": 13
    },
    {
      "id": 8,
      "created_at": "2026-08-07T09:15:00Z",
      "project": "bragfile",
      "type": "shipped",
      "title": "MCP list filter parity",
      "impact": "agents can ask for a bounded window in one call",
      "rank": 6,
      "score": 0.016393,
      "tokens": 27
    },
    {
      "id": 6,
      "created_at": "2026-08-01T11:02:00Z",
      "project": "bragfile",
      "type": "shipped",
      "title": "Sparkline pulse command",
      "impact": "",
      "rank": 7,
      "score": 0.015873,
      "tokens": 15
    },
    {
      "id": 3,
      "created_at": "2026-05-02T19:30:00Z",
      "project": "bragfile",
      "type": "learned",
      "title": "FTS5 triggers fight attach",
      "impact": "",
      "rank": 8,
      "score": 0.015152,
      "tokens": 15
    }
  ]
}
```

`score` is `math.Round(s*1e6)/1e6`. Note the **markdown/JSON `filters` ordering
asymmetry** is expected and locked by DEC-014: markdown uses the declared order
(`query`, then `project`), JSON is a `map[string]string` that Go's encoder sorts
alphabetically (`project`, then `query`).

### Locked signatures (`internal/memory`)

```go
// Options selects and bounds a memory slice. There is deliberately NO `now`
// field: the recency signal is ordinal (a rank, not an age), so the ranking is
// time-invariant (DEC-043 sub-decision 3). The renderer owns the wall clock, as
// in every other DEC-014 consumer.
type Options struct {
	// Query is echoed into the envelope only; ranking uses Matched.
	Query string
	// Project, when non-empty, contributes a soft boost term — never a filter
	// (DEC-043 sub-decision 4).
	Project string
	// Matched is the entry ids Store.Search returned, IN ITS RANK ORDER
	// (FTS5 `ORDER BY rank, id DESC`). Nil/empty means "no query". Ids not in
	// the candidate pool, and duplicates, are ignored. CALLER CONTRACT: the
	// order is the signal — passing an arbitrary id set silently degrades the
	// relevance term to whatever order you happened to build.
	Matched []int64
	// Budget is the token budget for the rendered entry lines. Must be > 0;
	// the caller rejects 0/negative before calling (DEC-044 sub-decision 1).
	Budget int
}

// Item is one selected entry plus everything the renderers need. Line is the
// DEC-044 rendering and Tokens is its estimated cost, so the estimate and the
// render cannot disagree.
type Item struct {
	Entry  storage.Entry
	Line   string
	Tokens int
	Score  float64
	Rank   int // 1-based position within the slice
}

// Result is the slice plus its accounting. Included+Skipped == Candidates.
type Result struct {
	Items           []Item // non-nil, in rank order
	Candidates      int    // deduped pool size
	Included        int
	Skipped         int
	EstimatedTokens int
	Budget          int
}

func Slice(entries []storage.Entry, opts Options) Result
func RenderLine(e storage.Entry) string
func EstimateTokens(s string) int

const (
	FusionK       = 60
	WeightRecency = 1.0
	WeightMatch   = 1.0
	WeightProject = 1.0
	PoolLimit     = 200
	DefaultBudget = 2000
	CharsPerToken = 4
)
```

### The pool composition (`internal/cli/memory.go`) — one bounded read per list

```go
// One pool per ranked list (DEC-043 sub-decision 5). Dropping the project read
// makes old same-project history unreachable; dropping the match read makes an
// old top-relevance entry unreachable. Slice dedupes the union by id.
pool, err := s.List(storage.ListFilter{Limit: memory.PoolLimit})
if err != nil {
	return fmt.Errorf("list entries: %w", err)
}
var matchedIDs []int64
if query != "" {
	fts5, qerr := buildFTS5Query(query) // DEC-010, reused as-is (same package)
	if qerr != nil {
		return UserErrorf("%v", qerr)
	}
	hits, serr := s.Search(fts5, memory.PoolLimit)
	if serr != nil {
		return fmt.Errorf("search entries: %w", serr)
	}
	for _, e := range hits {
		matchedIDs = append(matchedIDs, e.ID)
	}
	pool = append(pool, hits...)
}
if project != "" {
	scoped, perr := s.List(storage.ListFilter{Project: project, Limit: memory.PoolLimit})
	if perr != nil {
		return fmt.Errorf("list entries: %w", perr)
	}
	pool = append(pool, scoped...)
}
```

### The locked cobra `Long` (literal artifact — transcribe verbatim)

```go
Long: `Print a ranked, token-budgeted slice of your own history — the corpus read back as working memory, cheap enough for an agent to load at the start of every session. Rule-based, deterministic, no LLM, no network.

Ranking blends three signals by reciprocal-rank fusion (DEC-043): how recent an entry is, how well it matches --query, and whether it belongs to --project. With neither flag the blend degenerates to plain recency, so a bare 'brag memory' is the zero-config session opener. --project is a soft boost, not a filter: it raises that project's entries without hiding anyone else's, so a decision you made elsewhere can still surface. Reach for 'brag list --project' when you want a hard filter.

The slice is bounded by an estimated TOKEN budget, not a row count, because a row count does not predict what a slice costs to read. Entries are packed greedily in rank order; one that does not fit is skipped and the fill continues, so a single long entry cannot starve the rest. The Budget section reports what was included, what was skipped, and the estimate, so your next --budget value is a calibration rather than a guess. The budget covers the entry lines; the surrounding envelope is a small fixed overhead on top.

The estimate is a documented character-count heuristic that sizes THIS retrieval. It is not a tokenizer, it is never stored, and it is never stamped on an entry (DEC-044).

Each entry renders as one line:
  - <id> <YYYY-MM-DD> [<project>/<type>] <title> — <impact>
with - for an absent project or type, and the impact clause only when there is one. The id is the handle: run 'brag show <id>' for the full record. The budget is always measured against this markdown line, in both formats, so --format json returns the same entries in the same order.

Output is markdown (default) or a single-object JSON envelope (--format json) per DEC-014.

Examples:
  brag memory                                   # the session opener: recent history, 2000 tokens
  brag memory --project bragfile                # boost one project without hiding the rest
  brag memory --query "auth rate limit"         # what do I already know about this?
  brag memory --query auth --project orbit      # both signals at once
  brag memory --budget 500                      # a tighter slice
  brag memory --format json | jq .slice[].id    # just the ids, for a follow-up call`,
```

`Short`: `Ranked, token-budgeted slice of your history — the corpus as working memory`

### The flag set, with **defaults stated explicitly** (§12 flag-default rule)

```go
cmd.Flags().String("query", "", "boost entries matching this full-text query (FTS5, DEC-010 semantics)")
cmd.Flags().String("project", "", "boost entries in this project (a soft boost, not a filter)")
cmd.Flags().Int("budget", memory.DefaultBudget, "token budget for the entry lines (must be > 0)")
cmd.Flags().String("format", "markdown", "output format (one of: markdown, json)")
```

- `--query` default `""` → no relevance term, no `Store.Search` call.
- `--project` default `""` → no project term, no project read.
- `--budget` default **`2000`** (`memory.DefaultBudget`). `0` and negatives are a
  `UserError` — **not** `--limit`'s "0 means unlimited" convention (DEC-044
  sub-decision 1, Option H).
- `--format` default `"markdown"`. Only `markdown` and `json` are accepted.
- No `--limit`, no `--detail`, no `--no-spark`, no window flags. See Out of scope.

### Locked error strings (the Failing Tests assert on substrings)

```go
"--budget must be greater than zero, got %d (an unbounded slice is `brag list`)"
"unknown --format value %q (accepted: markdown, json)"
"--project must not be empty"
```

`--query` errors surface `buildFTS5Query`'s existing messages through
`UserErrorf("%v", err)`, exactly as `runSearch` does — do not reword them.

### The honesty guard, verbatim (`internal/memory/memory_test.go`)

```go
// TestPackageEmitsNoReservedTagNamespace is the mechanical guard behind
// DEC-044's honesty clause: this package sizes a RETRIEVAL, and nothing it
// computes may ever be stamped onto an entry or enter the reserved tag
// namespace DEC-027 governs. Matching on the QUOTED prefix (rather than the
// bare token) is deliberate: it catches "this code is building a reserved tag"
// while staying immune to Go field names like `Tokens:` in a struct literal.
// Same idiom as internal/timewindow's TestPackageReadsNoWallClock.
func TestPackageEmitsNoReservedTagNamespace(t *testing.T) {
	forbidden := []string{`"agent:"`, `"model:"`, `"session:"`, `"cost:"`, `"tokens:"`}
	err := filepath.WalkDir(".", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, f := range forbidden {
			if strings.Contains(string(b), f) {
				t.Errorf("%s contains %s; this package never stamps a reserved tag (DEC-044 honesty clause)", path, f)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
```

**Consequence for your doc comments:** the package's own prose must explain the
DEC-027 boundary **without writing a quoted reserved prefix**. Say *"the
caller-reported token-count tag DEC-027 reserves"*, not the quoted literal. The
package doc comment is a good place for it; the DEC is where the full argument
lives.

### NOT-contains self-audit (§12) — run at design against the locked literals

| assertion | token | grep against the locked `Long` above | result |
|---|---|---|---|
| help does NOT contain `--limit` | `--limit` | the `Long`, the `Short`, and the four registered flags | **0 hits** — safe to assert |
| help does NOT contain `--detail` | `--detail` | same | **0 hits** — safe to assert |
| package source does NOT contain quoted reserved prefixes | `"tokens:"` etc. | the locked doc-comment wording above | **0 hits** — but fragile by construction; the rule is stated above the audit table |

Note `not a tokenizer` appears **in** the `Long` deliberately, so do **not** add a
NOT-contains assertion for `tokenizer` — assert its presence instead.

### Fail-first (AGENTS.md §9, sharpened by SPEC-072's reflection)

SPEC-072's ship reflection: for a new-package spec, `undefined: Slice` in the new
package is the weak "stray compilation error" signal. **Lead the fail-first run with
the consuming boundary** — write `internal/cli/memory_test.go` and
`internal/export/memory_test.go` first, and confirm they fail on a *missing command*
/ *golden mismatch* rather than only on a missing symbol. `internal/memory`'s own
tests can only produce `undefined:` until the package exists; that is expected and is
not evidence of anything.

### Small things that will otherwise cost you an hour

- **`sort.SliceStable`, twice, and only stable.** Sort the pool by
  `(CreatedAt DESC, ID DESC)` first, then stably by `score DESC`. That gives
  DEC-043's tie-break for free — no second comparator, and no `math` import.
- **Return `Items` non-nil** even when empty, so JSON renders `[]` not `null`
  (DEC-014 part 4). Same for `Result` on `Slice(nil, …)`.
- **The em-dash is U+2014 (`—`), three bytes.** It is in the byte counts above. Do
  not substitute a hyphen or an en-dash — every golden shifts.
- **`trimTrailingNewline`** already exists in `internal/export/markdown.go`; use it,
  and let the CLI add exactly one newline via `Fprintln`, like every sibling.
- **The renderer's `Now`** lives on `export.MemoryOptions`, not on `memory.Options`.
  The CLI passes `nowFunc()` (the UTC seam in `impact.go`), matching
  `coverage`/`wrapped`/`spark` — **not** `clock` (the local seam, which exists for
  `--day`). `memory` has no local-day concept.
- **`Entries: N` is the candidate-pool size**, not the included count. The `##
  Budget` section decomposes it.

---

## Build Completion

*Filled in at the end of the **build** cycle, before advancing to verify.*

- **Branch:** `feat/spec-073-memory-slice`
- **PR (if applicable):** (opened at end of this build session)
- **All acceptance criteria met?** Yes — all 20 checklist items verified: pure
  package (stdlib + `internal/storage` only), no `now`/no clock, degenerates
  to plain recency with no query/project, blend demonstrably differs from
  both plain recency and plain bm25 on the fixture, dedupes + input-order
  invariant, all seven constants exported and golden-pinned, `RenderLine`
  shape exact (all four bracket forms + no-trailing-dash), `EstimateTokens`
  ceil-over-bytes, estimate/render agreement by construction, `Included +
  Skipped == Candidates`, skip-and-continue (not stop-at-first-overflow),
  DEC-014 envelope byte-exact against Goldens 1–4 + the JSON golden,
  empty-pool and zero-included-but-non-empty-pool render correctly,
  markdown/JSON select identical ids in identical order, `--budget`
  0/negative and invalid `--query`/`--format` are `UserError`s, help shape
  locked phrases present and `--limit`/`--detail` absent, both mechanical
  guards (`TestPackageReadsNoWallClock`, `TestPackageEmitsNoReservedTagNamespace`)
  pass, full gate set green.
- **New decisions emitted:**
  - `DEC-043` — the memory slice's blended ranking (emitted at design, with this spec)
  - `DEC-044` — the memory slice's token budget + line shape (emitted at design)
- **Fail-first result (AGENTS.md §9):** `internal/memory`'s own package
  produced the expected `undefined: Slice`/`Options`/`Result` compile errors
  before implementation (confirmed via the editor's live diagnostics against
  the test file written first). The spec's sharper guidance — lead fail-first
  with the `internal/cli`/`internal/export` consuming boundary so the signal
  is a missing command or golden mismatch, not just an undefined symbol — was
  only partially followed: `internal/memory/memory.go` was implemented and
  its own tests turned green before `internal/export/memory_test.go` and
  `internal/cli/memory_test.go` were written, because the ranking arithmetic
  needed to be validated against the spec's hand-computed score table first
  and doing that inside the pure package was the fastest way to catch an
  arithmetic slip early (one was caught: an initial draft used `k*rank`
  instead of `k+rank` for the match term, caught immediately against the
  `TestSlice_ScoresMatchTheLockedConstants` table). `go build ./...` on the
  incomplete tree did fail on the missing `export.ToMemoryMarkdown` /
  `cli.NewMemoryCmd` symbols before either was implemented, which is the
  weak signal for those two files specifically. See reflection Q3.
- **Deviations from spec:** None substantive. Two design-decidable
  ambiguities were resolved by cross-referencing other parts of the spec
  rather than guessing: (1) `Item.Rank` is the 1-based position in the FULL
  fused ranking, assigned before budget trimming and retained unrenumbered on
  included items (so a skip-and-continue slice can show non-contiguous ranks
  like 1 and 4) — pinned by DEC-044's own accepted-consequence prose ("rank 4
  (`5`) is included while ranks 2 and 3 are not"); (2) `ToMemoryMarkdown`/
  `ToMemoryJSON` take a precomputed `memory.Result` rather than raw entries +
  options (unlike every other DEC-014 renderer, which computes its aggregate
  internally) — inferred from `TestMemoryCmd_JSONAndMarkdownSelectTheSameSlice`,
  which only makes sense if `Slice` runs once in the CLI and both renderers
  consume the same `Result`, which is also what makes sub-decision 6
  ("budget always measured against markdown, in both formats") true by
  construction rather than by two implementations agreeing.
  `guidance/questions.yaml`'s `memory-slice-fusion-constants` entry (listed
  under Outputs as a file this spec modifies) was found already filed,
  verbatim as expected, from the design cycle — no build-time action needed;
  noted here rather than silently skipped.
- **Follow-up work identified:**
  - `AGENTS.md` §11 has no glossary entry for `brag spark` or `brag story`, and
    `docs/architecture.md`'s package table omits `internal/spark`,
    `internal/timewindow`, `internal/story`, and `internal/capture`. Surfaced by
    this spec's premise audit; deliberately not bundled. A small docs chore.
  - `buildFTS5Query` is still duplicated between `internal/cli` and
    `internal/mcpserver` (DEC-024's recorded debt). SPEC-074 will be the **third**
    consumer, which is DEC-042 revisit trigger (b) — extract it there, copying
    `internal/timewindow`'s precedent.
  - PR #124 (go-sdk `1.6.1 → 1.7.0`) still gates SPEC-074's DEC-045 §12(b)
    resources-API pre-flight.
  - `just test-docs` group E2 currently fails, but on a stray reference
    inside a gitignored `.claude/worktrees/happy-kilby-09218e/` directory (a
    parallel session's worktree artifact, unrelated to this repo's tracked
    content and to SPEC-073). Not this spec's to fix; flagged so a future
    verify session doesn't mistake it for a regression this branch introduced.

### Build-phase reflection (3 questions, short answers)

1. **What was unclear in the spec that slowed you down?**
   — Nothing was unclear so much as two things were implicit rather than
   stated: the `Item.Rank` semantics and the export functions' signature
   (see Deviations above). Both were resolvable from other parts of the same
   spec without guessing, which is really a compliment to how tightly the
   spec's pieces cross-check each other — but an explicit line in the locked
   signatures section for each would have saved the few minutes of
   cross-referencing.

2. **Was there a constraint or decision that should have been listed but wasn't?**
   — No. DEC-043/DEC-044/DEC-014/DEC-010/DEC-011/DEC-017/DEC-031/DEC-042 and
   the `no-sql-in-cli-layer`/`test-before-implementation`/
   `errors-wrap-with-context`/`no-new-top-level-deps-without-decision`/
   `no-cgo`/`stdout-is-for-data-stderr-is-for-humans`/`one-spec-per-pr`
   constraints covered everything actually touched. The `storagetest.Backdate`
   pointer (called out explicitly because SPEC-072 flagged its omission) paid
   for itself immediately.

3. **If you did this task again, what would you do differently?**
   — Stub `export.ToMemoryMarkdown`/`ToMemoryJSON` and `cli.NewMemoryCmd`
   (even trivially, returning wrong bytes) before implementing
   `internal/memory`, so the fail-first run genuinely exercises the
   consuming-boundary signal the spec asks for (a golden mismatch or a
   `brag memory: unknown command`, not a compile error) rather than
   validating the ranking arithmetic in isolation first. The outcome was
   identical here because the hand-computed score table caught the one real
   arithmetic bug immediately, but the process order didn't match the
   spec's explicit sequencing advice.

---

## Reflection (Ship)

*Appended during the **ship** cycle.*

1. **What would I do differently next time?**
   — <answer>

2. **Does any template, constraint, or decision need updating?**
   — <answer>

3. **Is there a follow-up spec I should write now before I forget?**
   — <answer>

4. **What can a user do now that they couldn't before?**
   — <answer>
