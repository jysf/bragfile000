---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-048
  type: story                      # epic | story | task | bug | chore
  cycle: ship
  blocked: false
  priority: high
  complexity: M                    # M — fourth DEC-014 consumer, but adds a NEW window semantics (calendar), a NEW aggregate helper (WithImpact split), one new command, one new render file, and emits DEC-028. More than SPEC-020 (which consumed DEC-014 verbatim); less than SPEC-018 (which created the package + the DEC).

project:
  id: PROJ-004
  stage: STAGE-011
repo:
  id: bragfile

agents:
  architect: claude-opus-4-8
  implementer: claude-opus-4-8     # usually same Claude, different session
  created_at: 2026-07-06

references:
  decisions:
    - DEC-028   # EMITTED by this spec — calendar windows, project=initiative, impact-first envelope shape
    - DEC-014   # EXTENDED (fourth consumer) — single-object envelope (JSON) + provenance/summary block (markdown); window semantics deliberately diverge (calendar, not rolling)
    - DEC-008   # `--since` date format (YYYY-MM-DD | Nd/Nw/Nm) — reused verbatim via cli.ParseSince
    - DEC-011   # JSON per-entry shape — DEC-028 uses a NARROWER 4-key projection, not the 9-key shape
    - DEC-006   # cobra framework — new `brag impact` subcommand
    - DEC-007   # required-flag validation in RunE — mutually-exclusive window flags + `--format` use UserErrorf
    - DEC-013   # markdown export shape — provenance-block convention DEC-014 inherited and SPEC-048 reuses
    - DEC-017   # entries↔project relationship — grouping is on the `project` field (the initiative axis)
  constraints:
    - no-sql-in-cli-layer
    - stdout-is-for-data-stderr-is-for-humans
    - errors-wrap-with-context
    - test-before-implementation
    - one-spec-per-pr
  related_specs:
    - SPEC-007   # shipped; ListFilter struct + Store.List read path; impact uses ListFilter{Since, Project, Type, Tag}
    - SPEC-018   # shipped; emitted DEC-014, seeded internal/aggregate, echoFiltersForSummary precedent SPEC-048 mirrors
    - SPEC-019   # shipped; second DEC-014 consumer; GroupEntriesByProject helper SPEC-048 reuses verbatim
    - SPEC-020   # shipped; third DEC-014 consumer; per-spec payload-shape divergence precedent (maps vs arrays)
---

# SPEC-048: `brag impact` — the calendar-windowed, initiative-grouped impact digest

## Context

First spec of STAGE-011 and the deterministic data foundation the
v0.4.0 story surface reads. `brag impact` is the **fourth** consumer of
DEC-014's rule-based digest envelope, joining `brag summary` (SPEC-018),
`brag review` (SPEC-019), and `brag stats` (SPEC-020). Like them it is
local-first, no-network, no-model — a rule-based digest, not an LLM
feature. STAGE-012's `brag story --audience` (the headline of PROJ-004)
reads this digest; `brag impact` is the deterministic foundation, the
shaping comes later.

Where the existing three digests answer "what happened recently?"
(`summary`), "here's the period, now reflect" (`review`), and "chart of
myself over all time" (`stats`), `brag impact` answers a fourth
question: **"what did I actually move the needle on this
quarter/month/year, grouped by initiative?"** It selects the entries
that carry an `impact` statement, groups them by project (= initiative),
and renders the impact text in full over a calendar reporting period.

The spec does **four** things in one pass and, unlike SPEC-020, **emits
a DEC** (DEC-028) because two choices are genuinely new to this consumer:

1. **Adds `brag impact --quarter|--month|--year|--since <date>
   [--format markdown|json] [--project P] [--type T] [--tag G]`** as the
   fourth DEC-014 consumer. Exactly one window flag is **required**; the
   four window flags are **mutually exclusive**. `--format` defaults to
   `markdown`. Filter flags compose with the window.

2. **Renders via `internal/export/impact.go`** (new file, sibling to
   `summary.go`/`review.go`/`stats.go`/`markdown.go`/`json.go`).
   Extends DEC-014's envelope verbatim. Per-spec payload keys at top
   level: `entries_in_window`, `entries_with_impact`,
   `counts_by_project` (map, over the with-impact subset — same shape
   and alpha-ASC-JSON asymmetry as SPEC-018's `counts_by_type`), and
   `impact_by_project` (array of `{project, entries:[{id, title,
   project, impact}]}` groups).

3. **Extends `internal/aggregate` with one helper** —
   `WithImpact(entries) []storage.Entry` (pure filter: returns entries
   whose `Impact` is non-empty, order preserved). Grouping reuses the
   existing `GroupEntriesByProject` verbatim on the with-impact subset.

4. **Emits DEC-028** locking the two new choices: **calendar** windows
   (a deliberate, tested divergence from DEC-014 choice 6's *rolling*
   windows for `summary`/`review`) and the **impact-first** body with a
   **4-key** per-entry projection (narrower than DEC-011's 9-key shape).

DEC-014 is **extended, not relitigated**: envelope shape, empty-state
rules (numbers `0`, arrays `[]`, dates `null`/`-`, objects `{}`),
markdown provenance-then-summary-then-payload convention, and 2-space
JSON indent are all inherited. Only the **window semantics** diverge —
calendar, not rolling — and that divergence is scoped to this command's
new flags (`--quarter`/`--month`/`--year`/`--since`), which
`summary`/`review` never had, so DEC-014 choice (6) is untouched for
them. SPEC-018's/019's/020's load-bearing goldens still hold; SPEC-048
adds its own (`TestToImpactMarkdown_DEC014FullDocumentGolden` and
`TestToImpactJSON_DEC028ShapeGolden`).

A deliberate per-spec-payload divergence inside DEC-014's envelope,
following the SPEC-018-maps-vs-SPEC-020-arrays precedent:
`impact_by_project` is an array-of-objects (grouped, ordered) while
`counts_by_project` is a map (lookup-by-key). Both coexist in the same
envelope by design.

Parent stage:
[`STAGE-011-impact-digest-foundation.md`](../stages/STAGE-011-impact-digest-foundation.md) —
Success Criteria (`brag impact` rule-based digest, DEC-014 envelope,
reuses `internal/aggregate`, no migration) and Design Notes ("extends
the DEC-014 rule-based output family… no LLM in the binary"). Project:
PROJ-004 (story surface, v0.4.0).

## Goal

Ship `brag impact --quarter|--month|--year|--since <date> [--format
markdown|json] [--project|--type|--tag]` as the fourth DEC-014 consumer:
select the in-window entries that carry an `impact` statement, group
them by project (= initiative), and render the impact text in full over
a **calendar** reporting period, consuming DEC-014's envelope verbatim
except for the deliberately-divergent calendar window semantics locked
in DEC-028. Surface a `<shown>/<in-window> with impact` tally so no
impact-less rows are silently dropped.

## Inputs

- **Files to read:**
  - `/AGENTS.md` — §6 cycle rules; §9 testing conventions (golden
    style, monotonic-tiebreak, load-bearing-golden-first); §12 design
    rules (decide-at-design-time, NOT-contains self-audit,
    literal-artifact-as-spec, flag-default explicitness); §14 confidence.
  - `/guidance/constraints.yaml` — the five referenced constraints.
  - `/decisions/DEC-028-impact-digest-window-and-shape.md` — the choices
    this spec implements (emitted by this spec).
  - `/decisions/DEC-014-rule-based-output-shape.md` — the envelope
    extended.
  - `/decisions/DEC-008-since-date-format.md` — the `--since` parser
    reused.
  - `/decisions/DEC-011-json-output-shape.md` — the 9-key shape the
    4-key projection deliberately narrows.
  - `internal/aggregate/aggregate.go` — `GroupEntriesByProject`
    (reused verbatim), `NoProjectKey`; add `WithImpact` here.
  - `internal/export/summary.go` — the closest sibling renderer
    (`SummaryOptions`, `ToSummaryMarkdown`, `ToSummaryJSON`,
    `trimTrailingNewline`); SPEC-048 mirrors its shape.
  - `internal/export/review.go` — `GroupEntriesByProject` markdown
    grouping precedent.
  - `internal/cli/summary.go` — `runSummary`, `rangeCutoff`,
    `echoFiltersForSummary`; SPEC-048's `runImpact` mirrors it plus the
    window-flag mutual-exclusion + calendar math.
  - `internal/cli/since.go` — `ParseSince` (reused verbatim for
    `--since`).
- **Data read:** `Store.List(ListFilter{Since, Project, Type, Tag})` —
  the existing read path; the impact-first split and window math live
  above storage.
- **No schema change.** The `impact` column already exists on
  `entries` (`storage.Entry.Impact`), populated via `brag add
  --impact`/editor. No migration.

## Outputs

- **New file `internal/export/impact.go`** — `ImpactOptions`,
  `ToImpactMarkdown`, `ToImpactJSON`, the `impactEnvelope` /
  `impactProjectGroup` / `impactEntry` JSON structs.
- **New file `internal/export/impact_test.go`** — the failing tests
  below.
- **New file `internal/cli/impact.go`** — `NewImpactCmd`, `runImpact`,
  the `windowCutoff` calendar helper, `echoFiltersForImpact`.
- **New file `internal/cli/impact_test.go`** — the CLI-layer failing
  tests below.
- **Edit `internal/aggregate/aggregate.go`** — add `WithImpact`.
- **Edit `internal/aggregate/aggregate_test.go`** — append
  `TestWithImpact_*`.
- **Edit `cmd/brag/main.go` (or wherever subcommands register)** — wire
  `NewImpactCmd()` into the root command (mirror how `NewSummaryCmd()`
  is registered).
- **New file `/decisions/DEC-028-impact-digest-window-and-shape.md`** —
  emitted by this spec (already drafted in this design cycle).
- **Edit `/guidance/questions.yaml`** — the
  `impact-window-calendar-vs-rolling` question (confidence < 0.8 on the
  window sub-choice; already added in this design cycle).
- **Edit `docs/api-contract.md`** — add a `brag impact` section (window
  flags, envelope, DEC-028 cross-link). *(Build transcribes; the
  literal doc block is not embedded in this spec because
  api-contract.md's section shape is established by the other three
  digest sections — build mirrors `brag summary`'s section verbatim.)*
- **Edit `docs/tutorial.md` + `README.md`** *(status-change premise
  audit, §9)* — if either lists the shipped command surface or a "what's
  there" table, add `brag impact`. See Premise Audit below for the exact
  greps to run.

### Premise audit note (planned, not build-time discovery)

This spec is **additive** (new command, new file, new aggregate helper,
new DEC). No existing behavior is inverted or removed. The additive
premise-audit cases that apply (§9):

- **New command → doc references update.** `brag impact` is a new
  command; grep the docs for the command-surface list and add it. See
  Premise Audit section for the executed greps + expected hits.
- **Addition to a tracked collection.** DEC count and any
  literal-count assertion over `internal/export/*_test.go` digest
  goldens: none asserted (each digest golden is self-contained). The
  DEC list is prose in `docs/`, not a count-asserted collection.

## Acceptance Criteria

1. `brag impact --quarter` prints a markdown impact digest for the
   current calendar quarter: `# Bragfile Impact`, the provenance block
   (`Generated:` / `Scope: quarter` / `Filters: (none)` /
   `Entries: <shown>/<in-window> with impact`), then `## Impact` with
   `### <project>` groups, each entry as `- <id>: <title>` + an indented
   impact line — **only** for entries whose `impact` is non-empty.
2. `--month` / `--year` behave identically for the current calendar
   month / year. `--since <date>` accepts DEC-008's formats via
   `ParseSince`.
3. The four window flags are **mutually exclusive** — passing two is a
   `UserError` on stderr (non-zero exit, empty stdout). Passing **none**
   is a `UserError` (required). `--format` defaults to `markdown`;
   `--format json` emits the DEC-014 envelope; any other value is a
   `UserError`.
4. **Calendar, not rolling** (the load-bearing divergence): the window
   lower bound is the calendar period start (`time.Date` constructor),
   NOT `now - 90/30/365 days`. An entry created on the first instant of
   the calendar quarter is IN; one created the day before is OUT.
5. `--format json` emits the envelope with flat keys `generated_at`,
   `scope`, `filters`, `entries_in_window`, `entries_with_impact`,
   `counts_by_project` (map, with-impact subset), `impact_by_project`
   (array of `{project, entries:[{id, title, project, impact}]}`),
   2-space indent. `scope` echoes `"quarter"`/`"month"`/`"year"`/
   `"since:<raw>"`.
6. In-window entries with **empty impact** increment `entries_in_window`
   but not `entries_with_impact`, and do NOT appear in `counts_by_project`
   or `impact_by_project`.
7. Impact text renders **in full** — never elided — in both markdown
   (the indented line) and JSON (the `impact` key).
8. Empty-window / no-impact: provenance renders (both tally counts `0`),
   the `## Impact` body and per-spec payload sections are OMITTED from
   markdown; JSON renders `entries_in_window`/`entries_with_impact` as
   `0`, `counts_by_project` as `{}`, `impact_by_project` as `[]`.
9. Filter flags `--project`/`--type`/`--tag` compose with the window and
   echo into `filters` exactly as `brag summary` does; `(no project)`
   grouping and `--project` filtering are not contradictory.
10. `internal/cli/impact.go` imports no `database/sql` / SQL driver
    (`no-sql-in-cli-layer`); the digest body goes to stdout, all errors
    to stderr (`stdout-is-for-data-stderr-is-for-humans`); errors wrap
    with context; `WithImpact` has happy-path + empty tests.

## Failing Tests

Written during design (this spec), made to pass during build. Load-
bearing goldens are written FIRST (§9). All fixtures use injected
`Now`/`Since` so the calendar math is deterministic. Fixture `Now` is
`2026-07-06T12:00:00Z` (Q3 2026 → calendar quarter start `2026-07-01`;
calendar month start `2026-07-01`; calendar year start `2026-01-01`).

### Shared renderer fixture (`internal/export/impact_test.go`)

```go
// impactFixture: 5 entries. 3 carry impact (shown), 2 don't (counted,
// not shown). Projects alpha (2 with impact), beta (1 with impact),
// gamma (1 in-window but NO impact → excluded from body/counts),
// (no project) (1 in-window, NO impact → excluded). Chrono-ASC within
// alpha (IDs 1 then 4) with non-monotonic id/time pairing so the
// ID-tiebreak path in GroupEntriesByProject is exercised.
var impactFixture = []storage.Entry{
    {ID: 1, Title: "alpha-old", Description: "d", Tags: "auth",
        Project: "alpha", Type: "shipped",
        Impact:    "cut p95 login latency 40%",
        CreatedAt: time.Date(2026, 7, 1, 10, 0, 0, 0, time.UTC),
        UpdatedAt: time.Date(2026, 7, 1, 10, 0, 0, 0, time.UTC)},
    {ID: 2, Title: "beta-mid",
        Project: "beta", Type: "learned",
        Impact:    "onboarding time down to 1 day",
        CreatedAt: time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC),
        UpdatedAt: time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC)},
    {ID: 3, Title: "gamma-noimpact",
        Project: "gamma", Type: "shipped",
        Impact:    "", // in-window, NO impact → excluded
        CreatedAt: time.Date(2026, 7, 3, 10, 0, 0, 0, time.UTC),
        UpdatedAt: time.Date(2026, 7, 3, 10, 0, 0, 0, time.UTC)},
    {ID: 4, Title: "alpha-new",
        Project: "alpha", Type: "shipped",
        Impact:    "removed the nightly cron entirely",
        CreatedAt: time.Date(2026, 7, 4, 10, 0, 0, 0, time.UTC),
        UpdatedAt: time.Date(2026, 7, 4, 10, 0, 0, 0, time.UTC)},
    {ID: 5, Title: "unbound-noimpact",
        Type:      "fixed", // (no project), NO impact → excluded
        Impact:    "",
        CreatedAt: time.Date(2026, 7, 5, 10, 0, 0, 0, time.UTC),
        UpdatedAt: time.Date(2026, 7, 5, 10, 0, 0, 0, time.UTC)},
}

var impactFixedNow = time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
```

> Note: the renderer receives the already-in-window slice (the CLI does
> the windowing). `impactFixture` therefore stands in for "everything in
> the window"; `entries_in_window` = 5, `entries_with_impact` = 3.
> `ImpactOptions.EntriesInWindow` carries the raw in-window count so the
> renderer can print the tally without re-deriving it.

#### Test 1 — `TestToImpactMarkdown_DEC014FullDocumentGolden` (LOAD-BEARING — write FIRST)

Byte-exact assertion. `ImpactOptions{Scope: "quarter", Filters:
"(none)", FiltersJSON: nil, EntriesInWindow: 5, Now: impactFixedNow}`.
Expected document:

```
# Bragfile Impact

Generated: 2026-07-06T12:00:00Z
Scope: quarter
Filters: (none)
Entries: 3/5 with impact

## Impact

### alpha

- 1: alpha-old
  cut p95 login latency 40%
- 4: alpha-new
  removed the nightly cron entirely

### beta

- 2: beta-mid
  onboarding time down to 1 day
```

Locks: provenance block + the `Entries: 3/5 with impact` tally line;
alpha before beta (alpha-ASC); within alpha, id 1 before id 4
(chrono-ASC); gamma and (no project) absent (no impact); impact line
indented two spaces and rendered in full; no trailing newline (matches
`trimTrailingNewline` contract).

#### Test 2 — `TestToImpactJSON_DEC028ShapeGolden` (LOAD-BEARING — write SECOND)

Byte-exact JSON assertion on the same fixture/opts (FiltersJSON nil →
`{}`). Expected:

```json
{
  "generated_at": "2026-07-06T12:00:00Z",
  "scope": "quarter",
  "filters": {},
  "entries_in_window": 5,
  "entries_with_impact": 3,
  "counts_by_project": {
    "alpha": 2,
    "beta": 1
  },
  "impact_by_project": [
    {
      "project": "alpha",
      "entries": [
        {
          "id": 1,
          "title": "alpha-old",
          "project": "alpha",
          "impact": "cut p95 login latency 40%"
        },
        {
          "id": 4,
          "title": "alpha-new",
          "project": "alpha",
          "impact": "removed the nightly cron entirely"
        }
      ]
    },
    {
      "project": "beta",
      "entries": [
        {
          "id": 2,
          "title": "beta-mid",
          "project": "beta",
          "impact": "onboarding time down to 1 day"
        }
      ]
    }
  ]
}
```

Locks: top-level flat keys in declared order; `counts_by_project` map
alpha-ASC (Go json map ordering); `impact_by_project` array in group
order (alpha, beta); the 4-key per-entry projection (`id`, `title`,
`project`, `impact` — NOT the 9-key DEC-011 shape); 2-space indent.
**Design-time expected-value check (§12b):** the group order (alpha
before beta) and the within-alpha order (id 1 before id 4) are computed
from `GroupEntriesByProject`'s locked sort — confirmed against the
fixture, not hand-typed blindly.

#### Test 3 — `TestToImpact_EmptyWindowShape`

Input: `nil` entries, `EntriesInWindow: 0`. Markdown = header +
provenance only (`Entries: 0/0 with impact`), no `## Impact` body. JSON:
`entries_in_window` `0`, `entries_with_impact` `0`, `counts_by_project`
`{}`, `impact_by_project` `[]` (non-nil), `filters` `{}`.

#### Test 4 — `TestToImpact_InWindowButNoImpactExcluded`

Input: the fixture but with `EntriesInWindow: 5` and only the two
impact-less entries would-be-shown check. Assert markdown body contains
`3/5 with impact`, does NOT contain `gamma-noimpact`, does NOT contain
`unbound-noimpact`, and JSON `entries_with_impact == 3` while
`entries_in_window == 5`. *(NOT-contains self-audit, §12: the forbidden
tokens `gamma-noimpact` / `unbound-noimpact` appear only in this test's
fixture data and commentary, never in any `Long` string or rendered
literal — confirmed at design.)*

#### Test 5 — `TestToImpact_ImpactTextRenderedInFull`

Input: one entry with a long multi-clause impact string containing a
colon and a percent sign. Assert the full string appears verbatim in
both the markdown indented line and the JSON `impact` value (guards
against elision/truncation).

#### Test 6 — `TestToImpactMarkdown_FiltersEchoed`

`ImpactOptions{Scope: "quarter", Filters: "--project alpha",
FiltersJSON: {"project":"alpha"}, EntriesInWindow: 2, ...}` over the two
alpha entries. Assert the `Filters: --project alpha` line renders and
the JSON `filters` object is `{"project":"alpha"}`.

### `internal/aggregate/aggregate_test.go` (existing file — tests appended)

#### Test 7 — `TestWithImpact_FiltersNonEmptyImpactPreservingOrder`

`WithImpact(impactFixture)` returns exactly the 3 entries with non-empty
`Impact`, in input order (ids 1, 2, 4). Confirms the empty-impact
entries (ids 3, 5) are dropped and order is preserved (not re-sorted —
grouping does the sorting).

#### Test 8 — `TestWithImpact_EmptyInputAndAllEmptyImpact`

`WithImpact(nil)` → non-nil empty slice (`len == 0`, `!= nil` so JSON
callers never see null). `WithImpact([entries all with Impact==""])` →
non-nil empty slice.

### `internal/cli/impact_test.go` (new file)

CLI tests build a `*cobra.Command` via `NewImpactCmd()` with separate
`outBuf`/`errBuf` (§9), a `t.TempDir()` DB seeded through `storage`.

#### Test 9 — `TestImpactCmd_RequiresExactlyOneWindow`

- No window flag → `UserError`, `errBuf` mentions the required window
  set, `outBuf` empty.
- Two window flags (`--quarter --month`) → `UserError` naming the
  conflict, `outBuf` empty. *(This is the mutual-exclusion assertion.)*

#### Test 10 — `TestImpactCmd_CalendarNotRolling` (LOAD-BEARING divergence)

Seed two entries with impact: one at the **calendar quarter start**
(`<Qstart>T00:00:00Z`) and one **one day before** the quarter start.
Inject a fixed `now` inside the quarter (via the injectable clock seam —
see Implementation Context). Run `--quarter`. Assert the quarter-start
entry IS present in output and the day-before entry is ABSENT. This
would FAIL under a rolling `now - 90d` reading (which would include the
day-before entry when now is early in the quarter). This is the test
that makes DEC-028 choice (1) load-bearing.

#### Test 11 — `TestImpactCmd_SinceReusesParseSince`

`--since 2026-01-01` selects entries `>= 2026-01-01`; `--since <bad>`
→ `UserError` (delegated to `ParseSince`). `scope` echoes
`"since:2026-01-01"`.

#### Test 12 — `TestImpactCmd_FormatDefaultAndUnknown`

No `--format` → markdown (assert `# Bragfile Impact` in `outBuf`).
`--format json` → assert `outBuf` parses as JSON with a `scope` key.
`--format xml` → `UserError`, `outBuf` empty. *(§12 flag-default:
`--format` default is `"markdown"`, stated in the literal `Long` and
the flag registration.)*

#### Test 13 — `TestImpactCmd_StdoutStderrSeparation`

A successful `--quarter` run writes the digest to `outBuf` only;
`errBuf.Len() == 0`. A `UserError` run writes to `errBuf` only;
`outBuf.Len() == 0`. Enforces `stdout-is-for-data-stderr-is-for-humans`.

## Implementation Context

Everything build needs without re-discovering it.

### The read path

`brag impact` reads via the existing `Store.List(ListFilter{...})`. The
window becomes `ListFilter.Since`; filter flags map to
`ListFilter.Project/Type/Tag`. There is **no** upper-bound field on
`ListFilter` and none is needed: the calendar period end is always
"now," and every stored `created_at` is `<= now`, so `Since` alone
bounds the window correctly. **Do not add a SQL `until` clause** — the
window's upper edge is implicit.

### The calendar window helper (CLI layer, pure)

Mirror `rangeCutoff` from `internal/cli/summary.go` but with calendar
math. `windowCutoff(window string, sinceRaw string, now time.Time)
(cutoff time.Time, scope string, err error)`:

```
now is UTC.
"quarter": qStartMonth = ((int(now.Month())-1)/3)*3 + 1   // 1,4,7,10
           cutoff = time.Date(now.Year(), time.Month(qStartMonth), 1, 0,0,0,0, time.UTC)
           scope  = "quarter"
"month":   cutoff = time.Date(now.Year(), now.Month(), 1, 0,0,0,0, time.UTC)
           scope  = "month"
"year":    cutoff = time.Date(now.Year(), 1, 1, 0,0,0,0, time.UTC)
           scope  = "year"
"since":   cutoff, err = cli.ParseSince(sinceRaw)         // DEC-008 verbatim
           scope  = "since:" + sinceRaw
```

Use `time.Date` calendar constructors — NEVER `now.AddDate(0,0,-90)` or
day subtraction. This is the correctness core of the calendar-vs-rolling
divergence.

### Mutual exclusion + required (CLI layer)

Register `--quarter`, `--month`, `--year` as `Bool` flags and `--since`
as `String`. In `runImpact`, collect which of the four are set
(`cmd.Flags().Changed(...)`). Zero set → `UserErrorf("one of --quarter,
--month, --year, --since is required")`. Two-or-more set →
`UserErrorf("--quarter, --month, --year, --since are mutually exclusive
(got %s)", ...)`. Exactly one → dispatch into `windowCutoff`. (Cobra's
`MarkFlagsMutuallyExclusive` handles pairs but not "exactly one
required" cleanly across a bool+string mix, and its error goes to a
different path than `UserErrorf`; do the check explicitly so the error
rides the established `UserError`→stderr path per DEC-007.)

### Injectable clock seam (for Test 10)

Per AGENTS.md §9 (`os`-level calls go through an injectable package var),
reference the wall clock in `internal/cli/impact.go` through a
package-level `var nowFunc = func() time.Time { return time.Now().UTC() }`
so Test 10 can substitute a fixed instant. This mirrors the
`getCwd`/`clock` seams SPEC-031/032/036 added. `runImpact` calls
`nowFunc()` once and threads it into both `windowCutoff` and
`ImpactOptions.Now` (single source, like `stats` threads `opts.Now`).

### The renderer (`internal/export/impact.go`)

Mirror `summary.go` structure exactly. `ImpactOptions`:

```go
type ImpactOptions struct {
    Scope           string            // "quarter"|"month"|"year"|"since:<raw>"
    Filters         string            // markdown line ("(none)" or echoed flags)
    FiltersJSON     map[string]string // JSON filters object (nil → {})
    EntriesInWindow int               // raw in-window count (for the tally)
    Now             time.Time         // injected; Generated: line
}
```

`ToImpactMarkdown(entries, opts)`: header `# Bragfile Impact`, blank,
`Generated:`/`Scope:`/`Filters:` lines, then
`Entries: <len(WithImpact(entries))>/<opts.EntriesInWindow> with impact`.
If `WithImpact(entries)` is empty, return after the provenance block
(trim trailing newline). Else `## Impact`, then for each
`GroupEntriesByProject(WithImpact(entries))` group: blank, `### <proj>`,
blank, and for each entry `- <id>: <title>\n  <impact>`.

`ToImpactJSON(entries, opts)`: the `impactEnvelope` struct (field order
= JSON key order):

```go
type impactEnvelope struct {
    GeneratedAt      string                `json:"generated_at"`
    Scope            string                `json:"scope"`
    Filters          map[string]string     `json:"filters"`
    EntriesInWindow  int                   `json:"entries_in_window"`
    EntriesWithImpact int                  `json:"entries_with_impact"`
    CountsByProject  map[string]int        `json:"counts_by_project"`
    ImpactByProject  []impactProjectGroup  `json:"impact_by_project"`
}
type impactProjectGroup struct {
    Project string        `json:"project"`
    Entries []impactEntry `json:"entries"`
}
type impactEntry struct {
    ID      int64  `json:"id"`
    Title   string `json:"title"`
    Project string `json:"project"`
    Impact  string `json:"impact"`
}
```

`CountsByProject` is built from `WithImpact` grouped counts (init to
`map[string]int{}` so empty renders `{}`). `ImpactByProject` init to
`[]impactProjectGroup{}` so empty renders `[]`. `Filters` nil → `{}`.
Marshal with `json.MarshalIndent(env, "", "  ")`. Note
`impactEntry.Project` echoes the group project (so `(no project)` group
entries carry `"project": "(no project)"` if such a group ever renders —
though with impact-first, only entries WITH a project OR (no project)
that have impact reach here; the group key is `NoProjectKey` for
empty-project entries, consistent with the other digests).

### The aggregate helper (`internal/aggregate/aggregate.go`)

```go
// WithImpact returns the subset of entries whose Impact field is
// non-empty, preserving input order. Used by brag impact (SPEC-048):
// the impact digest is impact-first — impact-less entries are counted
// in provenance but excluded from the grouped body. Empty input or an
// all-empty-impact input returns a non-nil empty slice (JSON callers
// never see null).
func WithImpact(entries []storage.Entry) []storage.Entry {
    out := make([]storage.Entry, 0, len(entries))
    for _, e := range entries {
        if e.Impact != "" {
            out = append(out, e)
        }
    }
    return out
}
```

Grouping is **not** re-implemented: `GroupEntriesByProject` already
sorts alpha-ASC (NoProjectKey last) with chrono-ASC + ID-tiebreak within
group. Feed it `WithImpact(entries)`.

### Filters echo (CLI layer)

Mirror `echoFiltersForSummary` verbatim as `echoFiltersForImpact` over
the same three-flag set (`tag`, `project`, `type`). Same
"third-caller-threshold" reasoning applies — two callers with the same
set is still below the lift-to-shared bar; keep it local. *(If build
judges the two are now byte-identical and wants to lift a shared
`echoFilters3` helper, that is a Rejected-alternative below — do NOT do
it silently.)*

### The cobra command (literal `Long`, §12 literal-artifact-as-spec)

```
Long: `Print a rule-based impact digest for a calendar reporting period: the entries that carry an impact statement, grouped by initiative (project), with each impact rendered in full. No LLM.

Output is markdown (default) or a single-object JSON envelope (--format json) per DEC-014/DEC-028. Exactly one window is required and the windows are mutually exclusive:
  --quarter   the current calendar quarter (Jan-Mar / Apr-Jun / Jul-Sep / Oct-Dec), up to now
  --month     the current calendar month, up to now
  --year      the current calendar year, up to now
  --since D   entries on or after D (YYYY-MM-DD or Nd/Nw/Nm), up to now

Windows are CALENDAR periods, not rolling — this differs from brag summary on purpose (the story surface reports by quarter/month/year). Only entries with a non-empty impact appear in the body; the provenance line tallies how many in-window entries had one. Filter flags --tag/--project/--type compose with the window.

Examples:
  brag impact --quarter                              # this calendar quarter, markdown
  brag impact --year --format json                   # this calendar year, JSON envelope
  brag impact --since 2026-01-01 --project alpha     # since a date, one initiative`,
```

Flags (with explicit defaults, §12 flag-default rule):

```go
cmd.Flags().Bool("quarter", false, "impact for the current calendar quarter")
cmd.Flags().Bool("month", false, "impact for the current calendar month")
cmd.Flags().Bool("year", false, "impact for the current calendar year")
cmd.Flags().String("since", "", "impact since a date (YYYY-MM-DD or Nd/Nw/Nm)")
cmd.Flags().String("format", "markdown", "output format (one of: markdown, json)")
cmd.Flags().String("tag", "", "filter to entries whose tags contain this token")
cmd.Flags().String("project", "", "filter to entries with this project (exact match)")
cmd.Flags().String("type", "", "filter to entries with this type (exact match)")
```

**NOT-contains self-audit (§12), run at design:** Test 4 asserts the
markdown body does NOT contain `gamma-noimpact` / `unbound-noimpact`.
Grepping the load-bearing prose above (the `Long` string, the markdown
golden, the flag help) for those two tokens: **zero hits.** They live
only in the test fixture. Clean.

### Registration

Mirror `NewSummaryCmd()`'s registration in `cmd/brag/main.go` (or the
root-command assembly): `rootCmd.AddCommand(cli.NewImpactCmd())`. Build:
read how the other three digests are added and follow the identical
line.

## Locked design decisions

Each has ≥1 paired failing test (§9 traceability).

1. **Calendar windows, not rolling (DEC-028 choice 1).** `--quarter`/
   `--month`/`--year` compute the period start with `time.Date`
   constructors; `--since` reuses `ParseSince`. Paired test: **Test 10**
   (`TestImpactCmd_CalendarNotRolling`) — the quarter-start entry is in,
   the day-before is out, which fails under a rolling reading.

2. **Exactly one window flag, mutually exclusive, required (DEC-028
   choice 1 / DEC-007).** Paired test: **Test 9**
   (`TestImpactCmd_RequiresExactlyOneWindow`).

3. **Initiative = project; reuse `GroupEntriesByProject` (DEC-028
   choice 2).** No new grouping helper. Paired tests: **Test 1 / Test 2**
   (alpha-before-beta group order, id-1-before-id-4 within alpha — the
   locked sort) and **Test 7**.

4. **Impact-first body with a tally (DEC-028 choice 3).** Only
   `WithImpact` entries in the body; `entries_in_window` vs
   `entries_with_impact` surfaced. Paired tests: **Test 4**
   (`TestToImpact_InWindowButNoImpactExcluded`), **Test 7/8**
   (`WithImpact`).

5. **Impact text rendered in full; 4-key per-entry projection (DEC-028
   choices 3-4).** Not the 9-key DEC-011 shape. Paired tests: **Test 5**
   (`TestToImpact_ImpactTextRenderedInFull`), **Test 2** (the 4-key JSON
   shape golden).

6. **Envelope extends DEC-014 verbatim; empty-state per DEC-014 (4).**
   Paired tests: **Test 1 / Test 2** (full goldens), **Test 3**
   (`TestToImpact_EmptyWindowShape`).

7. **`--format` defaults to `markdown`; unknown → UserError (DEC-007,
   §12 flag-default).** Paired test: **Test 12**
   (`TestImpactCmd_FormatDefaultAndUnknown`).

8. **Filters compose + echo like summary (DEC-028 choice 6).** Paired
   tests: **Test 6** (`TestToImpactMarkdown_FiltersEchoed`).

9. **stdout/stderr separation, no SQL in CLI, wrapped errors.** Paired
   test: **Test 13** (`TestImpactCmd_StdoutStderrSeparation`); the
   no-SQL constraint is a compile/import check.

### Rejected alternatives (build-time)

- **Adding an `Until`/upper-bound field to `ListFilter`.** Rejected:
  the calendar period end is always "now" and all `created_at <= now`,
  so `Since` alone bounds the window. Adding `Until` touches the storage
  read path and SQL for zero benefit here. If a future
  previous-complete-period (`--previous`) flag lands, revisit then.

- **Lifting a shared `echoFilters3` helper out of
  `echoFiltersForSummary` + `echoFiltersForImpact`.** Rejected at this
  spec: two callers with an identical three-flag set is still below the
  established third-caller lift threshold (SPEC-018 set this). Keep
  `echoFiltersForImpact` local; do NOT refactor `summary.go` in this
  PR (`one-spec-per-pr`).

- **Reusing DEC-011's 9-key `toEntryRecord` projection for the JSON
  entries.** Rejected: the narrative pipe needs only `{id, title,
  project, impact}`; the other five keys bloat the bundle. The narrow
  projection is deliberate (DEC-028 choice 4) and paired to Test 2.

- **Cobra `MarkFlagsMutuallyExclusive` for the window flags.** Rejected:
  it handles pairwise exclusion but not "exactly one required" cleanly
  across a bool+string flag mix, and its error bypasses the
  `UserError`→stderr path DEC-007 established. Do the check explicitly.

- **Including impact-less in-window entries with a `(no impact)`
  placeholder.** Rejected (DEC-028 Option D): buries the signal the
  command exists to surface. The tally preserves honesty; the body stays
  pure. Paired to Test 4.

## Premise Audit (AGENTS.md §9 — additive: new-command doc references)

This spec adds a new command (`brag impact`). Per §9's status-change /
new-command case, grep the docs for the shipped-command surface and
enumerate every hit as a planned Outputs update. **Design-side: greps
run against the repo, expected hits reconciled below (§9 audit-grep
cross-check).**

```
grep -rn "brag summary\|brag review\|brag stats" docs/ README.md
```

Expected: the four-digest family is documented together in
`docs/api-contract.md` (a per-command section) and mentioned in
`docs/tutorial.md` and `README.md` feature lists. Build adds a `brag
impact` section to `api-contract.md` (mirroring the `brag summary`
section verbatim) and a line to any tutorial/README command list that
enumerates the digests. **Build re-runs this grep and reconciles the
actual hit set against this enumeration before the doc sweep**; treat
any delta as a question, not a silent scope expansion.

No inversion/removal cases (nothing existing is changed). No
count-asserted collection is touched (each digest golden is
self-contained; the DEC list in docs is prose, not a count assertion).

## Build Completion

All 13 specified failing tests made real and passing; full gate set
green (`go test ./...` 615 passed; `gofmt -l .` empty; `go vet ./...`
clean; `CGO_ENABLED=0 go build ./...` success; `just test-docs` ALL OK;
`just test-hook` ALL OK).

- **Deviations from spec:**
  - **Test 13 error-path assertion adjusted (behavior-preserving).** The
    spec's Test 13 sketch said a `UserError` run "writes to `errBuf`
    only." But `NewRootCmd` sets `SilenceErrors: true` — cobra never
    writes the error to stderr in-process; `main.go` owns stderr
    formatting + exit-code mapping (same as every other digest command's
    tests, which assert `errors.Is(err, ErrUser)` on the returned error,
    not `errBuf`). The test enforces the same contract the spec intends
    (`stdout-is-for-data`: a failed run writes NOTHING to stdout, and the
    error is a `UserError` main.go routes to stderr) via the returned
    error rather than a populated `errBuf`. The stdout-cleanliness half
    of the separation is asserted exactly as written.
  - **One extra pure-helper unit test added (additive, not a deviation):**
    `TestWindowCutoff_CalendarArithmetic` (mirrors summary's
    `TestRangeCutoff_*` — locks the Q1/Q2/Q3/Q4 + month + year + since
    boundary math at the unit layer, per the calendar-vs-rolling
    correctness core). Nothing removed or weakened.
  - **Test 10 seeding seam:** the CLI test seeds two entries then rewrites
    their `created_at` via a raw `sql.DB` connection *inside the test
    file* (`seedImpactEntry`) — `Store.Add` always stamps `time.Now()`
    and the summary precedent explicitly rejected a
    `Store.SetCreatedAtForTesting` method. Raw SQL is confined to the
    `_test.go` file; production `internal/cli/impact.go` imports no
    `database/sql` (verified — `no-sql-in-cli-layer` holds).
  - `echoFiltersForImpact` kept local (not lifted to a shared helper) —
    per DEC-028's Rejected alternative (two callers with an identical
    three-flag set is below the third-caller lift threshold).

- **New DEC-* files:** none created during build. DEC-028 was drafted in
  the design cycle and is unchanged; no genuinely-new build decision
  warranted a DEC-029.

- **`--previous` NOT implemented** (explicitly future scope, per sign-off);
  the design does not foreclose it. **`brag impact` is TEXT-PURE** — no
  sparklines/visual columns (that is a separate STAGE-013 visual spec).

- **questions.yaml:** `impact-window-calendar-vs-rolling` marked
  `resolved` (calendar, current-in-progress-to-date; per orchestrator +
  user sign-off).

- **Reflection (3 answers):**
  1. *Did the spec's failing tests fully constrain the build?* Yes,
     tightly. The two byte-exact goldens (Test 1 markdown, Test 2 JSON)
     pinned the entire envelope shape — provenance tally line, indented
     impact rendering, group/within-group order, the 4-key projection,
     2-space indent — so the renderer had essentially one correct form.
     The only genuine judgment call was the Test 13 stderr-vs-returned-
     error framing, resolved by matching the established codebase pattern.
  2. *Was the calendar-vs-rolling divergence worth the DEC + dedicated
     test?* Yes. `windowCutoff` is small, but Test 10 (in at the quarter
     start, out the day before) is the single assertion that would catch
     a future refactor silently reintroducing `now.AddDate(0,0,-90)`.
     Without it the divergence would read as drift; with it, it reads as
     intent.
  3. *Any friction reusing the DEC-014 machinery?* None — `WithImpact` +
     `GroupEntriesByProject` composed cleanly (impact-first filter feeds
     the existing verbatim grouper), and mirroring `summary.go`'s
     structure meant the renderer was mostly transcription. The
     fourth-consumer path is now well-worn.

## Verify

**Verdict: ✅ APPROVED** (fresh independent VERIFY cycle, 2026-07-06).

Re-derived from SPEC-048 + DEC-028 + DEC-014 + constraints, not the
build's self-report. Six gates re-run independently, all exit 0:
`go test ./...` (615 passed), `gofmt -l .` (empty), `go vet ./...`
(clean), `CGO_ENABLED=0 go build ./...` (success), `just test-docs`
(ALL OK), `just test-hook` (ALL OK).

- **All 10 acceptance criteria satisfied**, each traced to code+test:
  windowing + provenance + impact-first body (`internal/export/impact.go`,
  `internal/cli/impact.go`); the two load-bearing goldens
  (`TestToImpactMarkdown_DEC014FullDocumentGolden`,
  `TestToImpactJSON_DEC028ShapeGolden`) pin the full envelope shape and
  the 4-key projection.
- **Calendar, not rolling — verified by reading AND exercising.**
  `windowCutoff` uses `time.Date` calendar constructors (no day
  subtraction); the `nowFunc` clock seam is threaded once into both
  `windowCutoff` and `ImpactOptions.Now`. An independent boundary probe
  confirmed `TestImpactCmd_CalendarNotRolling` genuinely distinguishes
  calendar from rolling: at `now=Jul-1 12:00` the calendar cutoff is
  `Jul-1 00:00` (day-before OUT) while a rolling `now-90d` cutoff would
  be `Apr-2` (day-before IN). Probe reverted; tree clean.
- **Impact-first confirmed live:** a real `brag impact --quarter` run
  over a seeded corpus printed `Entries: 2/3 with impact` — the
  no-impact row counted in `entries_in_window` but excluded from body,
  `counts_by_project`, and `impact_by_project`.
- **`--since` reuses DEC-008 `ParseSince`** (not a bespoke parser);
  scope echoes `since:<raw>`. Bad `--since` → UserError, empty stdout.
- **Constraints hold:** production `internal/cli/impact.go` imports no
  `database/sql` (raw SQL confined to `_test.go` seeding helper);
  live error paths (two windows / none / unknown format / bad since)
  all exit 1 with empty stdout and the error on stderr.
- **Documented deviation is legitimate**, not a masked failure: Test 13
  asserts `errors.Is(err, ErrUser)` + empty stdout rather than a
  populated `errBuf`, because `NewRootCmd` sets `SilenceErrors: true`
  (`internal/cli/root.go`) and `main.go` owns stderr — the exact pattern
  `summary_test.go`/`stats_test.go` use. Verified live: the real binary
  routes the UserError to stderr.
- **`--previous` correctly absent** (future scope). NOT-contains
  self-audit clean: `gamma-noimpact`/`unbound-noimpact` appear only in
  the spec's own fixture/commentary, never in `Long`, the renderer,
  docs, README, or DEC-028. No smuggled decision — DEC-028 covers every
  non-trivial choice; questions.yaml `impact-window-calendar-vs-rolling`
  is `resolved`. Docs swept (README, tutorial.md, api-contract.md).

## Reflection

*(filled during ship)*
