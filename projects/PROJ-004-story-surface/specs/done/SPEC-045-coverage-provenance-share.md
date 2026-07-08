---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-045
  type: story
  cycle: ship
  blocked: false
  priority: medium
  complexity: M

project:
  id: PROJ-004
  stage: STAGE-013                 # re-homed here from the never-activated PROJ-003/STAGE-010; STAGE-013 adopts it
repo:
  id: bragfile

agents:
  architect: claude-opus-4-8
  implementer: claude-opus-4-8     # usually same Claude, different session
  created_at: 2026-07-05

references:
  decisions:
    - DEC-033   # EMITTED by this spec — the coverage metric's definition + surface (IsAgentAuthored classifier unification, monthly agent-share, self-reference density)
    - DEC-024   # reserved agent:/model: provenance namespace — the classifier reads it
    - DEC-014   # EXTENDED (sixth consumer) — single-object envelope (JSON) + provenance/summary block (markdown)
    - DEC-028   # calendar-window semantics — reuses windowCutoff/selectedWindow verbatim
    - DEC-032   # --previous window modifier — coverage joins the calendar-windowed family
    - DEC-031   # sparkline primitive — the monthly agent-share trend renders via spark.Line
    - DEC-015   # normalized tags/taggings join the SQL classifier uses; the Go predicate mirrors it
    - DEC-007   # required/invalid-flag validation via UserErrorf
    - DEC-006   # cobra framework — new `brag coverage` subcommand
  constraints:
    - no-sql-in-cli-layer
    - stdout-is-for-data-stderr-is-for-humans
    - errors-wrap-with-context
    - test-before-implementation
    - one-spec-per-pr
  related_specs:
    - SPEC-043   # shipped; the agent:/model: SQL classifier (provenanceExistsClause) this unifies with a Go predicate
    - SPEC-048   # shipped; brag impact — the calendar-window core (windowCutoff/selectedWindow/nowFunc) coverage reuses
    - SPEC-051   # shipped; brag wrapped — the monthly-cadence + DEC-014 renderer coverage mirrors
    - SPEC-052   # shipped; internal/spark — the sparkline primitive the trend renders through
    - SPEC-053   # shipped; --previous — coverage joins the shared calendar-windowed family
---

# SPEC-045: `brag coverage` — agent-vs-human provenance share over time

## Context

Re-homed from the never-activated PROJ-003/STAGE-010 frame stub; STAGE-013
adopts it as the last feature spec before the v0.4.0 cut (SPEC-054). This is
action-register **P3** — the "read/measure" half of the agent-native thesis
that SPEC-043 (`brag list --author`, the read filter) unblocked.

v0.3.0 made the corpus *able* to distinguish agent- from human-authored
entries via the reserved `agent:`/`model:` provenance namespace (DEC-024), and
SPEC-043 shipped the read filter that classifies them (`provenanceExistsClause`
in `internal/storage/store.go`). The thesis is now **measurable but not yet
measured over time**. At the v0.3.0 cut the baseline is **0% agent-authored**
(189 human / 0 agent) — the trend starts accruing now, as the MCP write path
gets used, and this command surfaces it.

`brag coverage` is the **sixth consumer** of DEC-014's rule-based digest
envelope, joining `summary`/`review`/`stats`/`impact`/`wrapped`. Like them it
is local-first, no-network, no-model — a rule-based read over existing data,
**no schema change**. It reports **provenance share** (agent vs human counts +
share, windowed by month), a **monthly agent-share trend** rendered as a
sparkline (DEC-031), and a **self-reference density** measure (entries
mentioning `brag`/`bragfile`).

The spec does three things and **emits DEC-033** because two choices are
genuinely new: (a) the **classifier unification** — factoring a pure
`IsAgentAuthored(storage.Entry) bool` into `internal/aggregate` shared with
storage's SQL clause, closing the SPEC-043 drift-coupling WATCH; and (b) the
**coverage metric definition + surface** (the per-month agent/human/share
bucketing, the agent-share sparkline, and the self-reference density measure).

DEC-014 is **extended, not relitigated**: envelope shape, empty-state rules
(numbers `0`, arrays `[]`, objects `{}`), markdown provenance-then-body
convention, 2-space JSON indent are inherited. The calendar-window semantics
and the `--previous` modifier are reused verbatim from DEC-028/DEC-032. The
monthly bucketing reuses `aggregate.Cadence`'s month-labeling approach; the
sparkline reuses `spark.Line` (DEC-031) with JSON kept raw (DEC-031 choice f).

Parent stage:
[`STAGE-013-polish-and-v0-4-0-cut.md`](../stages/STAGE-013-polish-and-v0-4-0-cut.md) —
Success Criteria ("the P3 agent-assist metric — adopt the drafted SPEC-045").
Project: PROJ-004 (the story surface, v0.4.0).

## Goal

Ship `brag coverage --quarter|--month|--year|--since <date> [--previous]
[--format markdown|json] [--project|--type|--tag]` as the sixth DEC-014
consumer: over a **calendar** reporting window, classify each entry as
agent-authored (carries a reserved `agent:`/`model:` tag) or human-authored,
report the overall agent/human split + share, a **per-month agent/human/share
series** with an agent-share **sparkline**, and a **self-reference density**
measure (entries whose title or description mentions `brag`/`bragfile`). The
classifier is **single-sourced** with SPEC-043: one `IsAgentAuthored` predicate
in `internal/aggregate`, kept in agreement with the SQL clause by a
cross-package test.

## Inputs

- **Files to read:**
  - `/AGENTS.md` — §6 cycle rules; §9 testing conventions (golden style,
    load-bearing-golden-first, literal-artifact-as-spec); §12 design rules
    (decide-at-design-time, NOT-contains self-audit, flag-default explicitness);
    §14 confidence.
  - `/guidance/constraints.yaml` — the five referenced constraints.
  - `/decisions/DEC-033-coverage-metric-definition-and-surface.md` — the choices
    this spec implements (emitted by this spec, drafted in this design cycle).
  - `/decisions/DEC-024-...` — the reserved `agent:`/`model:` namespace.
  - `/decisions/DEC-014-...` — the envelope extended.
  - `/decisions/DEC-028-...` / `DEC-032-...` — the calendar window + `--previous`
    reused verbatim.
  - `/decisions/DEC-031-...` — the sparkline primitive + JSON-stays-raw rule.
  - `internal/storage/store.go` — `provenanceExistsClause` (the SQL classifier
    the Go predicate must agree with), `ListFilter{Since, Tag, Project, Type,
    Author}`, `Store.List`.
  - `internal/aggregate/aggregate.go` — `Cadence` (the month-labeling +
    bucketing precedent), `MostCommon`, `Span`; add `IsAgentAuthored` +
    `CoverageBucket` + `CoverageByMonth` here.
  - `internal/export/wrapped.go` — the closest sibling renderer (DEC-014
    envelope, `*string` null fields, `spark.Line` call, empty-state); coverage
    mirrors its shape.
  - `internal/export/stats.go` — `extractTags`/`extractProjects` precedent (the
    comma-split pattern `IsAgentAuthored` mirrors), `corpusSpanRecord`.
  - `internal/cli/impact.go` + `internal/cli/window.go` — `runImpact`,
    `nowFunc`, `windowCutoff`/`selectedWindow`, the `--previous` bounded-window
    filter, `echoFiltersForImpact`; coverage's CLI mirrors it.
  - `internal/mcpserver/provenance.go` — `reservedTag`/`stampProvenance` (the
    write side the classifier reads; for the drift-guard test).
- **Data read:** `Store.List(ListFilter{Since, Tag, Project, Type})` — the
  existing read path. Coverage does NOT set `ListFilter.Author` (it needs BOTH
  classes to compute a share); classification happens in Go via
  `IsAgentAuthored` so both counts come from one query. The `--previous`
  upper-bound filter runs in Go (DEC-032), SQL-free.
- **No schema change.** Classifies existing `agent:`/`model:` tags. No migration.

## Outputs

- **New file `internal/export/coverage.go`** — `CoverageOptions`,
  `ToCoverageMarkdown`, `ToCoverageJSON`, the `coverageEnvelope` /
  `selfReferenceRecord` JSON structs.
- **New file `internal/export/coverage_test.go`** — the renderer failing tests
  below.
- **New file `internal/cli/coverage.go`** — `NewCoverageCmd`, `runCoverage`,
  `echoFiltersForCoverage`, `monthLabelsBetween`.
- **New file `internal/cli/coverage_test.go`** — the CLI-layer failing tests.
- **Edit `internal/aggregate/aggregate.go`** — add `IsAgentAuthored`,
  `CoverageBucket`, `CoverageByMonth`, `SelfReferenceCount`, and the
  `shareRound` helper (see Implementation Context for exact signatures).
- **Edit `internal/aggregate/aggregate_test.go`** — append `TestIsAgentAuthored_*`,
  `TestCoverageByMonth_*`, `TestSelfReferenceCount_*`.
- **Edit `internal/storage/store_test.go`** — add the classifier-agreement test
  `TestProvenanceClassifier_GoPredicateMatchesSQLClause` (the drift guard). It
  lives in `internal/storage` because it exercises the real `Store.List` SQL
  path and compares it to `aggregate.IsAgentAuthored` over the same corpus.
- **Edit `cmd/brag/main.go`** — register `NewCoverageCmd()` (mirror
  `NewWrappedCmd()` at line 38).
- **New file `/decisions/DEC-033-coverage-metric-definition-and-surface.md`** —
  emitted by this spec (drafted in this design cycle).
- **Edit `/guidance/questions.yaml`** — the `coverage-sparkline-metric-choice`
  question (confidence < 0.8 on the "sparkline the share vs the agent count"
  sub-choice; added in this design cycle).
- **Edit `docs/api-contract.md`** — add a `brag coverage` section (mirror the
  `brag impact` section; window flags, envelope, DEC-033 cross-link).
- **Edit `docs/tutorial.md` + `README.md`** *(new-command doc references, §9)* —
  add `brag coverage` wherever the digest command family is enumerated. See
  Premise Audit for the greps to run.
- **Edit `AGENTS.md`** — add the `coverage` §11 glossary term (the digest
  family is enumerated there).
- **New exports:**
  - `aggregate.IsAgentAuthored(e storage.Entry) bool`
  - `aggregate.CoverageBucket{Period string; Agent int; Human int; Share float64}`
  - `aggregate.CoverageByMonth(entries []storage.Entry, months []string) []CoverageBucket`
  - `aggregate.SelfReferenceCount(entries []storage.Entry) int`
  - `export.CoverageOptions`, `export.ToCoverageMarkdown`, `export.ToCoverageJSON`
  - `cli.NewCoverageCmd`
- **Database changes:** none.

## Acceptance Criteria

1. `brag coverage --year` prints a markdown coverage digest for the current
   calendar year: `# Bragfile Coverage`, the provenance block (`Generated:` /
   `Scope: year` / `Filters: (none)` / `Entries: <total>`), then
   `## Provenance share` (agent count + %, human count + %),
   `## Monthly trend` (an `Agent share: <glyphs>` sparkline line, then one
   `- <YYYY-MM>: <a> agent / <h> human (<pct>%)` line per scope month,
   zero-filled), and `## Self-reference` (entries mentioning `brag`/`bragfile`
   + %).
2. `--quarter`/`--month`/`--since <date>` behave per DEC-028 (calendar
   periods); `--previous` shifts to the last-completed period per DEC-032.
   Exactly one window flag is **required** and the four are **mutually
   exclusive** (reuses `selectedWindow`). `--previous` + `--since` is a
   `UserError`; `--previous` with no window is a `UserError` (`selectedWindow`
   requires one).
3. **Classification is single-sourced.** An entry carrying `agent:*` or
   `model:*` counts as agent; one carrying neither counts as human; a
   false-positive topic tag (`agentic`, `modeling` — no colon) counts as human.
   The Go `aggregate.IsAgentAuthored` predicate and storage's SQL
   `provenanceExistsClause` classify the same corpus **identically** (the
   drift-guard test).
4. `--format` defaults to `markdown`; `--format json` emits the DEC-014
   envelope; any other value is a `UserError`.
5. `--format json` emits the envelope with flat keys `generated_at`, `scope`,
   `filters`, `total_entries`, `agent_entries`, `human_entries`, `agent_share`,
   `by_month` (array of `{period, agent, human, share}`), `self_reference`
   (`{count, share}`), 2-space indent. `scope` echoes the DEC-028/DEC-032 token
   (`"year"`, `"quarter:previous"`, `"since:<raw>"`, …).
6. **The monthly series is present in both renderers** — one bucket per scope
   month, zero-filled — so the trend is always fully shaped. The `share` is
   `agent / (agent+human)` for that month, rounded to 4 decimals; a month with
   zero entries has `share == 0`.
7. **The agent-share sparkline renders in markdown only** (DEC-031 choice f):
   `spark.Line` over the per-month agent-share (share×100, integer-scaled). JSON
   carries raw counts/shares, **no glyphs**. The sparkline is suppressed by
   `--no-spark` OR a present `NO_COLOR` env var (reuses the SPEC-052
   `lookupSparkEnv` posture).
8. **Self-reference density** counts entries whose `Title` OR `Description`
   contains `brag` (case-insensitive) — which subsumes `bragfile`. Reported as
   count + share of total. (Substring, not word-boundary — see LD5.)
9. Empty window (no entries in scope): provenance renders (through
   `Entries: 0`); the `## Provenance share`/`## Monthly trend`/`## Self-reference`
   body sections are OMITTED from markdown (DEC-014 part 4). JSON renders every
   key: `total_entries`/`agent_entries`/`human_entries` `0`, `agent_share` `0`,
   `by_month` still the full zero-filled month series, `self_reference`
   `{count:0, share:0}`, `filters` `{}`.
10. Filter flags `--project`/`--type`/`--tag` compose with the window and echo
    into `Filters:` / `filters` exactly as `impact` does.
11. `internal/cli/coverage.go` imports no `database/sql` / SQL driver
    (`no-sql-in-cli-layer`); the digest body goes to stdout, all errors to
    stderr (`stdout-is-for-data-stderr-is-for-humans`); errors wrap with
    context.

## Failing Tests

Written during **design** (this spec), made to pass during **build**.
Load-bearing goldens are written FIRST (§9). All goldens below were computed at
design time against the REAL `internal/spark.Line` (a scratch program inside the
module tree, since removed) and hand-verified arithmetic — they are faithful,
not hand-typed (the SPEC-049 wrong-golden lesson). All fixtures use injected
`Now` so the `Generated:` line is deterministic.

### Shared renderer fixture (`internal/export/coverage_test.go`)

```go
// coverageYearFixture: 10 entries across calendar 2026, modeling the
// post-v0.3.0 agent-adoption ramp — 0 agent-authored in H1, then agent
// entries appearing Jul/Sep/Nov/Dec. Exercises: agent: only (id 6),
// model: only (id 7), both (ids 4, 9), plain human (ids 1,2,3,5,10), the
// FALSE-POSITIVE guard (id 8: tags "agentic,modeling" — no colon → human),
// and self-reference (ids 1,4,9 mention "brag"). Totals: 4 agent / 6 human,
// agent_share 0.4; self-reference 3 (0.3).
var coverageYearFixture = []storage.Entry{
    {ID: 1, Title: "bragfile MVP retro", Description: "shipped the CLI", Tags: "process",
        CreatedAt: time.Date(2026, 2, 10, 10, 0, 0, 0, time.UTC)},
    {ID: 2, Title: "auth refactor", Description: "cleaned up login", Tags: "auth",
        CreatedAt: time.Date(2026, 3, 5, 10, 0, 0, 0, time.UTC)},
    {ID: 3, Title: "docs pass", Description: "rewrote the tutorial", Tags: "docs",
        CreatedAt: time.Date(2026, 5, 20, 10, 0, 0, 0, time.UTC)},
    {ID: 4, Title: "MCP server for brag", Description: "agent-native write spine",
        Tags: "mcp,agent:claude-code,model:claude-opus-4-8",
        CreatedAt: time.Date(2026, 7, 4, 10, 0, 0, 0, time.UTC)},
    {ID: 5, Title: "hotfix streak bug", Description: "local-day streak", Tags: "fix",
        CreatedAt: time.Date(2026, 7, 5, 10, 0, 0, 0, time.UTC)},
    {ID: 6, Title: "impact digest", Description: "calendar windows", Tags: "agent:claude-code",
        CreatedAt: time.Date(2026, 9, 12, 10, 0, 0, 0, time.UTC)},
    {ID: 7, Title: "story surface", Description: "audience shaping", Tags: "model:claude-opus-4-8,narrative",
        CreatedAt: time.Date(2026, 11, 3, 10, 0, 0, 0, time.UTC)},
    {ID: 8, Title: "modeling notes", Description: "agentic patterns essay", Tags: "agentic,modeling",
        CreatedAt: time.Date(2026, 11, 20, 10, 0, 0, 0, time.UTC)},
    {ID: 9, Title: "wrapped + sparklines", Description: "shareable year in brags",
        Tags: "agent:claude-code,model:claude-opus-4-8,visual",
        CreatedAt: time.Date(2026, 12, 15, 10, 0, 0, 0, time.UTC)},
    {ID: 10, Title: "release cut", Description: "v0.4.0 to homebrew", Tags: "release",
        CreatedAt: time.Date(2026, 12, 20, 10, 0, 0, 0, time.UTC)},
}
var coverageYearNow = time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC)
var coverageYearMonths = []string{ // the 12 ordered scope labels
    "2026-01","2026-02","2026-03","2026-04","2026-05","2026-06",
    "2026-07","2026-08","2026-09","2026-10","2026-11","2026-12"}
```

> Note: the renderer receives the already-in-window slice + the ordered
> `ScopeMonths` labels (the CLI does the windowing), exactly like `wrapped`.

#### Test 1 — `TestToCoverageMarkdown_DEC014FullDocumentGolden` (LOAD-BEARING — write FIRST)

Byte-exact. `CoverageOptions{Scope:"2026", Filters:"(none)", FiltersJSON:nil,
ScopeMonths: coverageYearMonths, Now: coverageYearNow, Spark:true}`. Expected
document (computed against real `spark.Line`; the trend glyphs are the share×100
sparkline):

```
# Bragfile Coverage

Generated: 2026-12-31T23:59:59Z
Scope: 2026
Filters: (none)
Entries: 10

## Provenance share

- Agent-authored: 4 (40.0%)
- Human-authored: 6 (60.0%)

## Monthly trend

Agent share: ▁▁▁▁▁▁▅▁█▁▅▅

- 2026-01: 0 agent / 0 human (0%)
- 2026-02: 0 agent / 1 human (0%)
- 2026-03: 0 agent / 1 human (0%)
- 2026-04: 0 agent / 0 human (0%)
- 2026-05: 0 agent / 1 human (0%)
- 2026-06: 0 agent / 0 human (0%)
- 2026-07: 1 agent / 1 human (50%)
- 2026-08: 0 agent / 0 human (0%)
- 2026-09: 1 agent / 0 human (100%)
- 2026-10: 0 agent / 0 human (0%)
- 2026-11: 1 agent / 1 human (50%)
- 2026-12: 1 agent / 1 human (50%)

## Self-reference

- Entries mentioning brag/bragfile: 3 (30.0%)
```

Locks: provenance block + `Entries: 10`; overall share `40.0%`/`60.0%`
(one-decimal `%.1f%%`); the sparkline glyph string `▁▁▁▁▁▁▅▁█▁▅▅` (share×100 →
`[0,0,0,0,0,0,50,0,100,0,50,50]` → min→max normalization: 0→`▁`, 50→`▅`,
100→`█`, verified against real `spark.Line`); per-month lines zero-filled with
the whole-percent `(%.0f%%)` share; self-reference `3 (30.0%)`; trailing newline
stripped (`trimTrailingNewline` contract).

#### Test 2 — `TestToCoverageJSON_DEC033ShapeGolden` (LOAD-BEARING — write SECOND)

Byte-exact JSON on the same fixture/opts (`FiltersJSON:nil → {}`; `Spark`
ignored by JSON). Expected:

```json
{
  "generated_at": "2026-12-31T23:59:59Z",
  "scope": "2026",
  "filters": {},
  "total_entries": 10,
  "agent_entries": 4,
  "human_entries": 6,
  "agent_share": 0.4,
  "by_month": [
    {
      "period": "2026-01",
      "agent": 0,
      "human": 0,
      "share": 0
    },
    {
      "period": "2026-02",
      "agent": 0,
      "human": 1,
      "share": 0
    },
    {
      "period": "2026-03",
      "agent": 0,
      "human": 1,
      "share": 0
    },
    {
      "period": "2026-04",
      "agent": 0,
      "human": 0,
      "share": 0
    },
    {
      "period": "2026-05",
      "agent": 0,
      "human": 1,
      "share": 0
    },
    {
      "period": "2026-06",
      "agent": 0,
      "human": 0,
      "share": 0
    },
    {
      "period": "2026-07",
      "agent": 1,
      "human": 1,
      "share": 0.5
    },
    {
      "period": "2026-08",
      "agent": 0,
      "human": 0,
      "share": 0
    },
    {
      "period": "2026-09",
      "agent": 1,
      "human": 0,
      "share": 1
    },
    {
      "period": "2026-10",
      "agent": 0,
      "human": 0,
      "share": 0
    },
    {
      "period": "2026-11",
      "agent": 1,
      "human": 1,
      "share": 0.5
    },
    {
      "period": "2026-12",
      "agent": 1,
      "human": 1,
      "share": 0.5
    }
  ],
  "self_reference": {
    "count": 3,
    "share": 0.3
  }
}
```

Locks: top-level flat keys in declared order; the 4-key per-month projection
(`period`, `agent`, `human`, `share`); `share` as a JSON number (`0`, `0.5`,
`1`, not a string, not `%`-scaled); `agent_share` `0.4`; `self_reference`
`{count, share}`; NO glyph field anywhere (DEC-031 choice f); 2-space indent.
**Design-time expected-value check (§12b):** every count/share was computed
against the fixture (agent = ids 4,6,7,9 = 4; human = 6; the false-positive id 8
classified human; self-reference = ids 1,4,9 = 3) — not hand-typed blindly.

#### Test 3 — `TestToCoverage_EmptyWindowShape`

Input: `nil` entries, `ScopeMonths: ["2026-07","2026-08","2026-09"]`,
`Scope:"2026-Q3"`. Markdown = header + provenance only (`Entries: 0`), no
`## Provenance share`/`## Monthly trend`/`## Self-reference` body (assert
`!strings.Contains(md, "## Monthly trend")`). JSON: `total_entries`/
`agent_entries`/`human_entries` `0`, `agent_share` `0`, `by_month` the 3
zero-filled buckets (`share` `0` each), `self_reference` `{count:0, share:0}`,
`filters` `{}`.

#### Test 4 — `TestToCoverage_SparklineMarkdownOnlyAndEscaped`

- With `Spark:true`, markdown contains the `Agent share: ` line followed by a
  string of block glyphs (assert the line is present and every rune after the
  label is in `▁▂▃▄▅▆▇█`).
- With `Spark:false`, markdown does NOT contain `Agent share:` (the
  `--no-spark`/`NO_COLOR` path).
- JSON output NEVER contains any block glyph and never a `sparkline` key,
  regardless of `Spark` — DEC-031 choice f (NOT-contains self-audit: the glyph
  runes appear only in the markdown golden + this assertion, never in any `Long`
  string or the JSON envelope; confirmed at design).

#### Test 5 — `TestToCoverage_FiltersEchoed`

`CoverageOptions{Filters:"--project alpha", FiltersJSON:{"project":"alpha"}}`;
assert `Filters: --project alpha` in markdown and `filters` object is
`{"project":"alpha"}` (only that key) in JSON.

### `internal/aggregate/aggregate_test.go` (existing file — tests appended)

#### Test 6 — `TestIsAgentAuthored_ClassifiesReservedNamespace`

Table over single entries: `agent:x` → true; `model:x` → true; `a,agent:x,b` →
true (mid-list); `agent:x,model:y` → true; `agentic,modeling` → **false**
(no colon — the false-positive guard); `""` (no tags) → false;
`auth,api` → false; `agent:` prefix with any suffix → true (a bare `agent:` tag
cannot occur — `reservedTag` drops empty values — but the predicate is
prefix-anchored regardless). Confirms the predicate matches the SQL `LIKE
'agent:%'/'model:%'`.

#### Test 7 — `TestCoverageByMonth_BucketsAndShareZeroFilled`

`aggregate.CoverageByMonth(coverageYearFixture, coverageYearMonths)` returns 12
buckets in month order; assert the agent/human counts and `share` per bucket
match the Test-2 JSON `by_month` values exactly (e.g. `2026-07` = `{1,1,0.5}`,
`2026-09` = `{1,0,1}`, `2026-03` = `{0,1,0}`, `2026-01` = `{0,0,0}`).
Zero-entry months are present with `{0,0,0}`. On empty input over the same
months → 12 `{0,0,0}` buckets. Confirms zero-filling + the 4-decimal share
rounding + one bucket per scope month.

#### Test 8 — `TestSelfReferenceCount_SubstringCaseInsensitive`

`aggregate.SelfReferenceCount(coverageYearFixture)` → `3` (ids 1,4,9 mention
"brag"/"bragfile" in title/description). A table also asserts: title-only match,
description-only match, mixed-case (`BragFile`), and a non-match (no "brag"
substring) → not counted; empty input → `0`.

### `internal/storage/store_test.go` (existing file — the drift guard)

#### Test 9 — `TestProvenanceClassifier_GoPredicateMatchesSQLClause` (LOAD-BEARING — the agreement test)

The cross-package agreement test that closes the SPEC-043 drift-coupling WATCH.
Seeds a `t.TempDir()` corpus covering every classification edge (agent-only,
model-only, both, plain-human, no-tags, the `agentic,modeling` false-positive,
an agent tag mid-list). Runs BOTH classifiers over the same rows:

- SQL side: `Store.List(ListFilter{Author:"agent"})` → the set the
  `provenanceExistsClause` selects.
- Go side: `Store.List(ListFilter{})` (all rows) → partition by
  `aggregate.IsAgentAuthored`.

Assert the two agent-sets are **identical** (same entry IDs), and the two
human-sets (`Author:"human"` vs `!IsAgentAuthored`) are identical. If either
classifier ever changes without the other, this test fails — the single-source
guarantee (AC 3). *(Design-time verification: this exact comparison was run
against the real `Store.List` SQL path at design and passed — 4 agent of 7,
false-positives correctly excluded — so the two definitions are byte-for-byte in
agreement at the moment the spec is locked.)*

### `internal/cli/coverage_test.go` (new file)

CLI tests build a `*cobra.Command` via `NewCoverageCmd()` with separate
`outBuf`/`errBuf` (§9), a `t.TempDir()` DB seeded through `storage`, and the
`withNowFunc`/`seed*Entry` harness copied from `impact_test.go`/`wrapped_test.go`.

#### Test 10 — `TestCoverageCmd_RequiresExactlyOneWindow`

Reuses `selectedWindow` (as `impact` does): no window flag → `UserError`,
`outBuf` empty; two window flags → `UserError` naming the conflict, `outBuf`
empty.

#### Test 11 — `TestCoverageCmd_CalendarWindowAndScope`

`--year` over a seeded corpus emits `Scope: year`; `--since 2026-01-01` emits
`Scope: since:2026-01-01`; `--quarter --previous` emits `Scope: quarter:previous`
(reuses `windowCutoff`; light wiring — the calendar math is already tested by
`impact`'s `TestWindowCutoff_*`). `--previous --since` → `UserError`.

#### Test 12 — `TestCoverageCmd_FormatDefaultAndUnknown`

No `--format` → markdown (assert `# Bragfile Coverage` in `outBuf`). `--format
json` → `outBuf` parses as JSON with an `agent_share` key. `--format yaml` →
`UserError`, `outBuf` empty. *(§12 flag-default: `--format` default is
`"markdown"`, stated in the literal `Long` and the flag registration.)*

#### Test 13 — `TestCoverageCmd_NoSparkAndNoColorSuppress`

`--year` (default) → `outBuf` contains `Agent share:`. `--year --no-spark` →
`outBuf` does NOT contain `Agent share:`. With `lookupSparkEnv` stubbed so
`NO_COLOR` is present (any value, including empty) → no `Agent share:` line
either. Mirrors `wrapped`'s spark-escape tests (reuses the `lookupSparkEnv`
seam).

#### Test 14 — `TestCoverageCmd_StdoutStderrSeparation`

A successful `--year` run writes the digest to `outBuf` only; the returned error
is nil. A `UserError` run (`--format yaml`) returns a `UserError` (assert
`errors.As`) and writes NOTHING to `outBuf` (matches the `impact`/`wrapped`
convention: `NewRootCmd` sets `SilenceErrors`, so `main.go` owns stderr; assert
on the returned error + empty stdout, per SPEC-048's Test 13 deviation).

#### Test 15 — `TestCoverageCmd_HelpShowsExamples`

`brag coverage --help` contains the literal `Examples:` label and the exact
example line `brag coverage --year` (unique-token discipline, §9).

## Implementation Context

Everything build needs without re-discovering it.

### The classifier unification (the WATCH this closes)

Two definitions of "agent-authored" exist today and must become one meaning:

- **SQL (storage):** `provenanceExistsClause` in `internal/storage/store.go` —
  `EXISTS (... t.name LIKE 'agent:%' OR t.name LIKE 'model:%')` over the
  taggings join (DEC-015). Prefix-anchored on the reserved namespace.
- **Go (aggregate, NEW):** `IsAgentAuthored` splits `Entry.Tags` (the
  comma-joined projection reconstructed from the same join) and prefix-matches
  each token. Because `Entry.Tags` is exactly the `GROUP_CONCAT` of `t.name`
  values, the two operate on the same token set; prefix `agent:`/`model:` on a
  token is the Go equivalent of `LIKE 'agent:%'/'model:%'` on `t.name`.

```go
// IsAgentAuthored reports whether e carries a reserved provenance tag
// (agent:<name> or model:<id>, DEC-024) — the SINGLE Go-side definition of
// "agent-authored", kept in agreement with storage's provenanceExistsClause
// SQL predicate by TestProvenanceClassifier_GoPredicateMatchesSQLClause. It
// splits Entry.Tags (the comma-joined projection of the same taggings join
// the SQL clause queries) and prefix-matches each token, mirroring the
// LIKE 'agent:%' / 'model:%' anchoring: a topic tag like "agentic" or
// "modeling" (no colon) is NOT provenance. This is the classifier SPEC-043's
// --author filter reads in SQL; brag coverage reads it in Go so it can count
// BOTH classes from one query.
func IsAgentAuthored(e storage.Entry) bool {
    for _, raw := range strings.Split(e.Tags, ",") {
        t := strings.TrimSpace(raw)
        if strings.HasPrefix(t, "agent:") || strings.HasPrefix(t, "model:") {
            return true
        }
    }
    return false
}
```

Do **not** re-implement the membership test a third time. `brag list --author`
(SQL) and `brag coverage` (Go) now share one *meaning*, pinned by Test 9.

### The monthly bucketer (aggregate, pure)

`CoverageByMonth` mirrors `Cadence`'s month-labeling exactly but splits by
provenance and computes a per-month share:

```go
// CoverageBucket is one month's provenance split: Period is the "YYYY-MM"
// label, Agent/Human the classified counts, Share = Agent/(Agent+Human)
// rounded to 4 decimals (0 when the month is empty). SPEC-045. The json tags
// are the on-the-wire shape the coverage envelope embeds directly (LD8 of
// SPEC-051): by_month[].period / .agent / .human / .share.
type CoverageBucket struct {
    Period string  `json:"period"`
    Agent  int     `json:"agent"`
    Human  int     `json:"human"`
    Share  float64 `json:"share"`
}

// CoverageByMonth buckets entries by UTC calendar month, classifies each via
// IsAgentAuthored, and emits one CoverageBucket per label in months order
// (zero-filled). months is the ordered "YYYY-MM" set the CLI derives from the
// window (12 for a year, 3 for a quarter, N for --since) so the series is
// always fully present, even on an empty window. Mirrors Cadence (SPEC-051).
func CoverageByMonth(entries []storage.Entry, months []string) []CoverageBucket {
    agent := map[string]int{}
    human := map[string]int{}
    for _, e := range entries {
        lbl := e.CreatedAt.UTC().Format("2006-01")
        if IsAgentAuthored(e) {
            agent[lbl]++
        } else {
            human[lbl]++
        }
    }
    out := make([]CoverageBucket, 0, len(months))
    for _, lbl := range months {
        a, h := agent[lbl], human[lbl]
        out = append(out, CoverageBucket{Period: lbl, Agent: a, Human: h, Share: shareRound(a, a+h)})
    }
    return out
}

// shareRound returns num/den rounded to 4 decimals (half-away-from-zero via
// math.Round), or 0 when den == 0. Used for per-month and overall shares so
// the JSON number is stable and the goldens are byte-exact.
func shareRound(num, den int) float64 {
    if den == 0 {
        return 0
    }
    return math.Round(float64(num)/float64(den)*10000) / 10000
}
```

`aggregate` gains `math` + `strings` imports if not already present (both
stdlib; `no-cgo`/`no-new-top-level-deps` unaffected).

### Self-reference density (aggregate, pure)

```go
// SelfReferenceCount returns how many entries mention "brag" (case-insensitive)
// in Title or Description — a proxy for dogfooding density (the corpus talking
// about the tool itself). Substring match: "brag" subsumes "bragfile" (LD5).
func SelfReferenceCount(entries []storage.Entry) int {
    n := 0
    for _, e := range entries {
        hay := strings.ToLower(e.Title + " " + e.Description)
        if strings.Contains(hay, "brag") {
            n++
        }
    }
    return n
}
```

### The CLI command (`internal/cli/coverage.go`)

Mirror `runImpact` structure exactly. Reuse `nowFunc`, `selectedWindow`,
`windowCutoff`, the `--previous` bounded-window Go filter, `echoFiltersFor*`,
and `lookupSparkEnv`. Coverage does NOT set `ListFilter.Author` — it reads all
in-window rows once and classifies in Go (needs both classes). Derive
`ScopeMonths` from `[cutoff, upperEdge)`:

```
// After windowCutoff returns (cutoff, end, scope, err) and the in-window
// entries are collected (with the --previous end filter applied in Go, per
// impact.go), derive the ordered month labels covering the window. The upper
// edge is `end` when non-zero (--previous), else `now`. Labels run from the
// cutoff's month to the upper edge's month inclusive, stepping AddDate(0,1,0)
// and formatting "2006-01" — the same shape wrapped's monthLabels produces,
// so an empty window still yields a fully-present zero-filled series.
```

A small `monthLabelsBetween(start, upperInclusive time.Time) []string` helper
(CLI layer, pure) produces the labels; it is coverage-local (no third caller
yet — `wrapped`'s `monthLabels` takes an explicit count from a known
year/quarter; coverage derives from an arbitrary `[cutoff, upper]`, so a shared
lift is a Rejected alternative below). The renderer receives the in-window slice
+ `ScopeMonths`. (For `--previous`, the upper edge is `end.Add(-time.Nanosecond)`
so the exclusive boundary month is not spuriously included; a completed period's
last in-scope month is the month before `end`.)

Command surface (literal `Long`, §12 literal-artifact-as-spec):

```
Long: `Print a rule-based coverage digest: how much of your work over a calendar reporting period was agent-authored vs human-authored, and how that share is trending month by month. Provenance is read from the reserved agent:/model: tags the MCP write path stamps (DEC-024). No LLM.

Output is markdown (default) or a single-object JSON envelope (--format json) per DEC-014/DEC-033. Exactly one window is required and the windows are mutually exclusive:
  --quarter   the current calendar quarter (Jan-Mar / Apr-Jun / Jul-Sep / Oct-Dec), up to now
  --month     the current calendar month, up to now
  --year      the current calendar year, up to now
  --since D   entries on or after D (YYYY-MM-DD or Nd/Nw/Nm), up to now

--previous shifts the selected window to the last-completed period (bounded on both ends). It requires a window flag and is incompatible with --since. Windows are CALENDAR periods, not rolling.

The monthly trend is a per-month agent-share sparkline (markdown only; --no-spark or NO_COLOR suppresses it). Filter flags --tag/--project/--type compose with the window.

Examples:
  brag coverage --year                               # this calendar year, markdown
  brag coverage --quarter --previous                 # the whole previous quarter
  brag coverage --since 2026-01-01 --format json     # since a date, JSON envelope`,
```

Flags (with explicit defaults, §12 flag-default rule):

```go
cmd.Flags().Bool("quarter", false, "coverage for the current calendar quarter")
cmd.Flags().Bool("month", false, "coverage for the current calendar month")
cmd.Flags().Bool("year", false, "coverage for the current calendar year")
cmd.Flags().String("since", "", "coverage since a date (YYYY-MM-DD or Nd/Nw/Nm)")
cmd.Flags().Bool("previous", false, "shift the window to the last-completed period")
cmd.Flags().String("format", "markdown", "output format (one of: markdown, json)")
cmd.Flags().String("tag", "", "filter to entries whose tags contain this token")
cmd.Flags().String("project", "", "filter to entries with this project (exact match)")
cmd.Flags().String("type", "", "filter to entries with this type (exact match)")
cmd.Flags().Bool("no-spark", false, "suppress the in-terminal agent-share sparkline")
```

### The renderer (`internal/export/coverage.go`)

Mirror `wrapped.go`. `CoverageOptions{Scope, Filters, FiltersJSON, ScopeMonths
[]string, Now time.Time, Spark bool}`. Markdown: header `# Bragfile Coverage`,
provenance block through `Entries: <len(entries)>`; if empty, return (DEC-014
part 4). Else `## Provenance share` (agent/human counts computed by partitioning
`entries` via `aggregate.IsAgentAuthored`, `%.1f%%` shares); `## Monthly trend`
— if `opts.Spark`, an `Agent share: <spark.Line(shareInts)>` line where
`shareInts[i] = int(math.Round(bucket.Share*100))`, then the per-month
`- <period>: <a> agent / <h> human (<%.0f%%>)` lines; `## Self-reference` line.
JSON: the `coverageEnvelope` struct (field order = key order):

```go
type coverageEnvelope struct {
    GeneratedAt   string                     `json:"generated_at"`
    Scope         string                     `json:"scope"`
    Filters       map[string]string          `json:"filters"`
    TotalEntries  int                        `json:"total_entries"`
    AgentEntries  int                        `json:"agent_entries"`
    HumanEntries  int                        `json:"human_entries"`
    AgentShare    float64                    `json:"agent_share"`
    ByMonth       []aggregate.CoverageBucket `json:"by_month"`
    SelfReference selfReferenceRecord        `json:"self_reference"`
}
type selfReferenceRecord struct {
    Count int     `json:"count"`
    Share float64 `json:"share"`
}
```

`ByMonth` is `aggregate.CoverageByMonth(entries, opts.ScopeMonths)` (embed the
aggregate struct directly, like `wrapped` embeds `CadenceBucket` — LD8 of
SPEC-051). Overall agent/human counts are computed by partitioning `entries`
via `aggregate.IsAgentAuthored` (NOT by summing the buckets — an entry outside
`ScopeMonths` cannot occur because the CLI derives the labels from the window,
but partitioning the slice is the single source of the totals). `AgentShare` and
`SelfReference.Share` use `aggregate.shareRound` over the total. `Filters` nil →
`{}`. `by_month` is always the full zero-filled series (never `null`). Marshal
with `json.MarshalIndent(env, "", "  ")`. **JSON never contains glyphs** — the
sparkline is a markdown-only rendering (DEC-031 choice f).

### Registration

`cmd/brag/main.go`: add `root.AddCommand(cli.NewCoverageCmd())` next to
`NewWrappedCmd()` (line 38).

### Decisions that apply

- `DEC-033` (this spec) — the coverage metric definition (per-month
  agent/human/share, agent-share sparkline, self-reference density) + the
  `IsAgentAuthored` classifier unification + the standalone-command surface.
- `DEC-024` — the reserved `agent:`/`model:` namespace the classifier reads.
- `DEC-014` — the envelope extended verbatim (sixth consumer).
- `DEC-028` + `DEC-032` — the calendar window + `--previous`, reused via the
  shared `windowCutoff`/`selectedWindow` core with no change.
- `DEC-031` — `spark.Line` for the trend; JSON stays raw (choice f); the
  `--no-spark`/`NO_COLOR` escape via `lookupSparkEnv`.
- `DEC-015` — the tags/taggings join both classifiers read.
- `DEC-007`/`DEC-006` — flag validation + cobra subcommand.

### Constraints that apply

- `no-sql-in-cli-layer` — `internal/cli/coverage.go` imports no `database/sql`.
  The read path is `Store.List(ListFilter{Since,...})`; classification +
  bucketing + the `--previous` upper-bound filter run in Go. Raw SQL in the CLI
  test file (for `created_at` rewrites) is confined to `_test.go`, matching
  `impact_test.go`. **Test 9 lives in `internal/storage`** precisely because it
  needs the SQL path — that is the storage layer's own test, not the CLI's.
- `stdout-is-for-data-stderr-is-for-humans` — digest to stdout, UserErrors to
  stderr (Test 14).
- `errors-wrap-with-context` — storage/config errors wrapped; user errors via
  `UserErrorf`.
- `test-before-implementation` — the Failing Tests above are written first.
- `one-spec-per-pr` — this PR is SPEC-045 only.

### Prior related work

- `SPEC-043` (shipped) — the SQL classifier + `--author` filter this unifies
  with. Its ship reflection named P3 as the natural next spec: "provenance share
  over time, windowed by month" — exactly this.
- `SPEC-048`/`SPEC-051`/`SPEC-052`/`SPEC-053` (shipped) — the calendar-window
  core, the DEC-014 renderer + monthly cadence, the sparkline primitive, and
  `--previous`. Coverage is assembled almost entirely from these.

### Out of scope (for this spec specifically)

- **First-class `agent`/`model` columns** — the DEC-024 "later, if earned"
  promotion. The tag convention stays the classifier; this spec is a read.
- **A provenance breakdown inside `brag stats`/`wrapped`** — would reshape those
  locked DEC-014 envelopes + goldens. Coverage is its own command (LD1).
- **Cost / token economics** (the exec-ROI dimension) — PROJ-005; the seed
  capture (SPEC-046) accrues it, but reporting it is a later metric.
- **Weekly buckets / a configurable bucket unit** — monthly only (matches
  `wrapped`'s cadence unit); a future call.

## Verify

**Verdict: ✅ APPROVED** (independent verify cycle, fresh session, re-derived
from spec + DEC-033 + constraints — build self-report NOT trusted).

- **Six gates (re-run):** `go test ./...` PASS (762 tests, 11 packages);
  `gofmt -l .` empty; `go vet ./...` clean; `CGO_ENABLED=0 go build ./...`
  clean; `just test-docs` ALL OK; `just test-hook` ALL OK. All exit 0.
- **No import cycle in SHIPPED state:** `go vet ./internal/storage/...
  ./internal/aggregate/...` clean; the agreement test is `package storage_test`
  (external) as the spec's deviation records — the correct Go resolution.
- **Classifier agreement (key check):**
  `TestProvenanceClassifier_GoPredicateMatchesSQLClause` PASS — SQL
  `provenanceExistsClause` and Go `aggregate.IsAgentAuthored` partition the same
  seeded corpus identically (4 agent / 3 human; `agentic,modeling` false-positive
  excluded by both). `brag list --author` unregressed:
  `TestList_FilterByAuthor` + author-compose tests green; the filter still uses
  the SQL clause (coverage does NOT set `ListFilter.Author`).
- **Live end-to-end:** built the binary, seeded a throwaway DB reproducing the
  year fixture (4 agent ids 4/6/7/9, 6 human, self-ref 3). Confirmed: overall
  share + per-month zero-filled series + self-reference; agent-share sparkline in
  markdown, suppressed by `--no-spark` and by `NO_COLOR`; JSON glyph-free with
  the flat keys + 4-key per-month projection + 2-space indent; `--previous` /
  window flags + scope tokens (`year`, `quarter:previous`, `since:<raw>`);
  `--project` filter echo; empty window omits markdown body but renders every
  JSON key. Validation: no-window / two-window / `--previous --since` /
  `--format yaml` all `UserError` with empty stdout (stdout/stderr separation).
- **Golden faithfulness (re-derived):** `spark.Line([0,0,0,0,0,0,50,0,100,0,50,50])`
  → min0/max100/span100 → 0=▁, 50=round(3.5)=4=▅, 100=7=█ → `▁▁▁▁▁▁▅▁█▁▅▅`,
  matches Test 1's markdown golden. Both load-bearing byte-goldens (Test 1 md,
  Test 2 JSON) PASS unmodified. No mis-computed golden.
- **Constraints:** `internal/cli/coverage.go` + `internal/export/coverage.go`
  import no `database/sql`/driver; `aggregate` has no `database/sql` (SQL only in
  `internal/storage`). `go.mod`/`go.sum` unchanged vs `main` (no new dependency).
- **No smuggled decision:** DEC-033 covers the metric + surface; no DEC-034.

## Locked design decisions

Each has ≥1 paired failing test (§9 traceability).

1. **LD1 — Standalone `brag coverage`, not a `stats`/`wrapped` section or a
   flag.** *Rejected:* (a) a `brag stats --provenance` flag or a stats section —
   reshapes the locked DEC-014 stats envelope + its byte-goldens (SPEC-043
   explicitly deferred "a provenance breakdown in brag stats" as needing its own
   spec); (b) a `wrapped` section — buries a *candid* self-metric inside a
   *celebratory shareable* digest (wrong audience; the story brief separates
   "reflect/for me" from "promote/for company"). Provenance-share-over-time is a
   distinct question deserving a distinct command; a standalone command extends
   DEC-014 additively (new keys in a new envelope) without touching any shipped
   golden. The draft framed it as "its own query"; "fold in" is satisfied by
   folding it into the *digest family* (same envelope, same window core, same
   sparkline), not into one existing command. Paired tests: **Test 1/2** (the
   `# Bragfile Coverage` standalone envelope), **Test 12** (own command).

2. **LD2 — Classifier single-sourced via `aggregate.IsAgentAuthored`, kept in
   agreement with the SQL clause by a cross-package test.** No third definition.
   `brag list --author` stays SQL (it composes with `LIMIT`); `brag coverage`
   uses the Go predicate (it needs both classes from one query). Agreement is
   the contract, not shared code (SQL and Go cannot literally share the
   membership expression). Paired tests: **Test 6** (the predicate), **Test 9**
   (the load-bearing agreement/drift guard).

3. **LD3 — Calendar windows + `--previous`, reused verbatim from
   DEC-028/DEC-032.** Coverage adds no window semantics; it calls
   `selectedWindow`/`windowCutoff` unchanged and applies the `--previous`
   bounded filter in Go exactly as `impact` does. Paired test: **Test 10/11**.

4. **LD4 — Sparkline the per-month AGENT SHARE (share×100), markdown-only; JSON
   raw.** The trend the command exists to show is *adoption* (share), not
   *volume* (count) — sparkline-ing the raw agent count would conflate a busy
   month with an agent-heavy one. `spark.Line` over `[]int` of share×100 gives a
   0–100 range that reads as "% agent-authored, month over month." JSON stays
   raw counts+shares (DEC-031 choice f); consumers re-render if they want a
   sparkline. *Rejected:* sparkline the agent COUNT (conflates volume with
   adoption); put glyphs in JSON (DEC-031 choice f forbids). This sub-choice is
   the reason DEC-033's confidence sits below 0.8 and gets a question (§14).
   Paired tests: **Test 1** (the exact glyph string over shares), **Test 4**
   (markdown-only + escaped + JSON-glyph-free).

5. **LD5 — Self-reference = substring `brag` (case-insensitive) in Title OR
   Description.** `brag` subsumes `bragfile`; a substring match is the simplest
   honest proxy for "the corpus talking about the tool." *Rejected:* word-
   boundary/regex matching (over-engineered for a density proxy; "bragging" is a
   near-zero false-positive in this corpus and not worth a regex), and matching
   Tags (tags are the classifier's domain; self-reference is about prose). Paired
   test: **Test 8**.

6. **LD6 — Overall shares render one-decimal (`40.0%`); per-month render whole-
   percent (`50%`); JSON shares are 4-decimal numbers.** The markdown precisions
   differ on purpose: the two headline shares deserve a decimal; twelve per-month
   lines read cleaner as whole percents. JSON is the precise machine value.
   Paired tests: **Test 1** (both markdown precisions), **Test 2** (`0.5`/`1`/
   `0.4` JSON numbers).

7. **LD7 — Coverage reads all in-window rows once and classifies in Go; it does
   NOT set `ListFilter.Author`.** A share needs both classes; two filtered
   queries would double the read and risk the two counts disagreeing about the
   window. One query + Go partition is simpler and correct. Paired test:
   **Test 7** (the bucketer over the full slice) + Test 9 (agreement).

8. **LD8 — `--format` default `"markdown"`, unknown → `UserError`; empty-state
   omits the body (DEC-014 part 4).** Mirrors every digest. Paired tests:
   **Test 12** (format), **Test 3** (empty).

### Rejected alternatives (build-time)

- **Lifting a shared `monthLabelsBetween` / merging with `wrapped`'s
  `monthLabels`.** Rejected at this spec: `wrapped.monthLabels(year, start,
  count)` takes an explicit count from a known year/quarter; coverage derives
  labels from an arbitrary `[cutoff, upperEdge]` window (since/previous make the
  span variable). Two different shapes; a shared helper is a premature lift (the
  third-caller threshold, SPEC-018). Keep `monthLabelsBetween` coverage-local;
  do NOT refactor `wrapped.go` in this PR (`one-spec-per-pr`).
- **Setting `ListFilter.Author` and issuing two queries.** Covered in LD7 —
  rejected for double-read + window-consistency risk.
- **Putting the sparkline glyphs in the JSON envelope.** Rejected — DEC-031
  choice f: a sparkline is a lossy visual of `by_month[].share`, not data.
- **Reusing `aggregate.Cadence` directly for the trend.** Rejected: `Cadence`
  counts all entries per month (one number); coverage needs the agent/human
  split + a share per month. `CoverageByMonth` is the provenance-aware sibling;
  it mirrors `Cadence`'s labeling but is a distinct, tested helper.
- **A `stats`-style `--provenance` flag.** Covered in LD1.

## Premise Audit (AGENTS.md §9 — additive: new-command doc references)

This spec adds a new command (`brag coverage`). Per §9's new-command case, grep
the docs for the digest-command surface and enumerate every hit as a planned
Outputs update. **Design-side: greps run against the repo, expected hits
reconciled below.**

```
grep -rn "brag summary\|brag review\|brag stats\|brag impact\|brag wrapped" docs/ README.md AGENTS.md
```

Expected: the digest family is documented together in `docs/api-contract.md` (a
per-command section), `docs/tutorial.md` (the digest section), `README.md`
(command list), and `AGENTS.md` §11 glossary. Build adds a `brag coverage`
section to `api-contract.md` (mirroring the `brag impact` section), a tutorial
line, a README command-list line, and the `coverage` §11 glossary term.
**Build re-runs this grep and reconciles the actual hit set against this
enumeration before the doc sweep** (§12 audit-grep cross-check); treat any delta
as a question, not silent scope expansion. Do NOT touch the DEC-014 *inventory*
sentences that enumerate consumers as-of-DEC-014's-authorship (historical
provenance lines, not a live command list — the SPEC-051 build precedent).

No inversion/removal cases (nothing existing changes). No count-asserted
collection is touched (each digest golden is self-contained).

## Notes for the Implementer

- The two byte-goldens (Test 1 markdown, Test 2 JSON) were computed at design
  against the real `spark.Line` and hand-verified. Before writing the renderer,
  re-run a scratch marshal + `spark.Line([]int{0,0,0,0,0,0,50,0,100,0,50,50})`
  and confirm `▁▁▁▁▁▁▅▁█▁▅▅` (the SPEC-051 build lesson: compute the golden
  values, including the sparkline, before the first `go test`).
- Test 9 is the whole point — write it first among the aggregate/storage tests
  and confirm it fails only because `aggregate.IsAgentAuthored` does not exist
  yet, then passes once added, then would fail if either classifier drifted.
- NOT-contains self-audit (§12): the block glyphs `▁▂▃▄▅▆▇█` must appear in the
  markdown golden and Test 4 only — NEVER in the `Long` string or the JSON
  envelope. Confirmed clean at design.
- `shareRound` is unexported in `aggregate`; `CoverageByMonth` and the renderer's
  overall-share both call it, so it lives in `aggregate` next to the bucketer
  (the renderer imports `aggregate` already). If build prefers, an exported
  `aggregate.Share(num, den int) float64` is acceptable — but keep ONE rounding
  definition so per-month and overall shares round identically.

## Build Completion

*Filled in at the end of the **build** cycle, before advancing to verify.*

- **Branch:** feat/spec-045-coverage
- **PR:** #92 (ready for review, not merged).
- **All acceptance criteria met?** Yes — all eleven AC. `brag coverage --year`
  renders the full markdown digest (provenance block → `## Provenance share`
  → `## Monthly trend` with the agent-share sparkline → `## Self-reference`);
  `--quarter`/`--month`/`--since`/`--previous` behave per DEC-028/DEC-032
  (exactly-one-required + mutually-exclusive via `selectedWindow`; `--previous
  --since` and `--previous`-with-no-window are `UserError`s); classification is
  single-sourced and the Go predicate matches the SQL clause (the drift-guard
  test); `--format` defaults to markdown, unknown → `UserError`; the JSON
  envelope carries the flat keys in order with a 4-key per-month projection and
  2-space indent; the monthly series is present + zero-filled in both renderers;
  the sparkline is markdown-only over share×100, escaped by `--no-spark`/
  `NO_COLOR`, and glyph-free in JSON; self-reference is case-insensitive
  substring `brag`; the empty window omits the markdown body but renders every
  JSON key; filters echo like `impact`; the CLI layer is SQL-free with the
  digest on stdout and errors on stderr.
- **GOLDEN FAITHFULNESS (first step, per the build brief):** both load-bearing
  goldens matched real output **byte-exact on the first `go test`** — the
  markdown golden (including the trend sparkline `▁▁▁▁▁▁▅▁█▁▅▅`, independently
  reconfirmed against real `spark.Line([]int{0,0,0,0,0,0,50,0,100,0,50,50})`
  before writing the renderer) and the JSON DEC-033 shape golden. No golden was
  edited; no code was contorted to fit a golden.
- **Classifier agreement + no `--author` regression:** the Go
  `aggregate.IsAgentAuthored` predicate and storage's SQL
  `provenanceExistsClause` classify the same seeded corpus **identically** —
  confirmed by `TestProvenanceClassifier_GoPredicateMatchesSQLClause`
  (`internal/storage`, external `storage_test` package to avoid the
  aggregate→storage import cycle): 4 agent of 7, the `agentic,modeling`
  false-positive correctly excluded, both agent-sets and human-sets equal by
  ID. `brag list --author` is unchanged — it still uses the SQL clause;
  `internal/storage`'s `TestList_FilterByAuthor` stays green.
- **New decisions emitted:** DEC-033 (emitted at design; implemented verbatim).
- **Files changed:**
  - `internal/aggregate/aggregate.go` — added `IsAgentAuthored`,
    `CoverageBucket`, `CoverageByMonth`, `SelfReferenceCount`, the unexported
    `shareRound` + its exported alias `Share` (one rounding definition); added
    `math` + `strings` imports.
  - `internal/aggregate/aggregate_test.go` — appended `TestIsAgentAuthored_*`,
    `TestCoverageByMonth_*`, `TestSelfReferenceCount_*` + a local fixture.
  - `internal/export/coverage.go` (new) — `CoverageOptions`,
    `ToCoverageMarkdown`, `ToCoverageJSON`, the `coverageEnvelope` /
    `selfReferenceRecord` structs, `partitionProvenance`, `pct`.
  - `internal/export/coverage_test.go` (new) — Tests 1–5 (the two byte-goldens
    + empty-shape + sparkline-markdown-only-and-escaped + filters-echoed).
  - `internal/cli/coverage.go` (new) — `NewCoverageCmd`, `runCoverage`,
    `monthLabelsBetween` (coverage-local), `echoFiltersForCoverage`.
  - `internal/cli/coverage_test.go` (new) — Tests 10–15.
  - `internal/storage/provenance_agreement_test.go` (new, `package
    storage_test`) — the drift-guard (Test 9).
  - `cmd/brag/main.go` — registered `NewCoverageCmd()` after `NewWrappedCmd()`.
  - `docs/api-contract.md` (new `brag coverage` section mirroring `brag
    impact` + a DEC-033 line), `docs/tutorial.md` (new subsection),
    `README.md` (command-list line), `AGENTS.md` §11 (`coverage` glossary
    term + refreshed `sparkline`/`--previous` terms).
  - `guidance/questions.yaml` — `coverage-sparkline-metric-choice` resolved
    (status → answered, share per the orchestrator sign-off).
- **Docs sweep (§9 / §12 audit-grep cross-check):** re-ran the Premise-Audit
  grep (`brag summary\|review\|stats\|impact\|wrapped`) — the digest family is
  documented in `docs/api-contract.md`, `docs/tutorial.md`, `README.md`, and
  `AGENTS.md` §11 exactly as enumerated; each got a `brag coverage` addition.
  The DEC-014 *inventory* sentence (api-contract line 1109, "for summary,
  review, stats, and impact") was left untouched — it is a historical
  provenance line for DEC-014's authorship, not a live command list (the
  SPEC-051 build precedent). NOT-contains self-audit: the block glyphs
  `▁▂▃▄▅▆▇█` appear only in the markdown golden + Test 4 — never in the `Long`
  string, the JSON envelope, or any other production string (grep-confirmed).
- **Deviations from spec:** One structural deviation, in-spirit: Test 9 lives
  in an **external** `package storage_test` file
  (`provenance_agreement_test.go`) rather than appended to the in-package
  `store_test.go`. The spec placed it "in `internal/storage`", which it is —
  but an in-package (`package storage`) test importing `internal/aggregate`
  (which imports `internal/storage`) forms an import cycle Go rejects. The
  external test package is the standard Go resolution and uses only the
  exported API + the literal `"agent"`/`"human"` author tokens; the test's
  intent (real SQL path vs Go predicate over one seeded corpus) is unchanged.
  Also, per the spec's own offered option, the renderer's overall-share uses
  the exported `aggregate.Share` alias of `shareRound` (kept as ONE rounding
  definition) since `shareRound` is unexported.
- **Gate results:** `go test ./...` PASS; `gofmt -l .` empty; `go vet ./...`
  clean; `CGO_ENABLED=0 go build ./...` clean; `just test-docs` ALL OK; `just
  test-hook` ALL OK.
- **Follow-up work identified:** The `coverage-sparkline-metric-choice`
  residual (the single-entry-month = 100% spike) is parked in questions.yaml
  for a post-v0.4.0 dogfooding re-check — no wire change needed to re-render
  the trend (JSON already carries raw counts+shares). No other follow-up
  surfaced during build.

## Reflection (Ship)

*Appended at ship (2026-07-07). Merged to main; ships in the v0.4.0 cut.
Last feature spec of STAGE-013 — only the release cut (SPEC-054) remains.*

1. **What would I do differently next time?** — Nothing notable; adopting a
   frame-stub (re-homing SPEC-045 from the closed PROJ-003 into STAGE-013) was
   clean, and the metric assembled almost entirely from shipped machinery
   (window core, wrapped's monthly renderer, `spark`, `--previous`) — a sign the
   digest family's shared primitives are paying compounding dividends.
2. **Does any template, constraint, or decision need updating?** — No. The
   real win to remember: the `IsAgentAuthored` Go predicate is now the single
   twin of storage's SQL `provenanceExistsClause`, pinned by a cross-package
   agreement test — closing the SPEC-043 drift WATCH. That drift-guard pattern
   (Go-predicate ↔ SQL-clause agreement) is the reusable move whenever a
   classifier must live on both sides of the `no-sql-in-cli-layer` boundary.
3. **Is there a follow-up spec I should write now before I forget?** — No. The
   `coverage-sparkline-metric-choice` residual (single-entry-month = 100% spike)
   is parked for post-v0.4.0 dogfooding, no wire change needed. STAGE-013's only
   remaining backlog item is SPEC-054, the v0.4.0 release cut.
