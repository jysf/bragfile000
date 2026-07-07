---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-051
  type: story
  cycle: verify
  blocked: false
  priority: high
  complexity: M

project:
  id: PROJ-004
  stage: STAGE-013
repo:
  id: bragfile

agents:
  architect: claude-opus-4-8
  implementer: claude-opus-4-8     # usually same Claude, different session
  created_at: 2026-07-06

references:
  decisions: [DEC-014, DEC-028, DEC-030, DEC-022, DEC-008, DEC-007, DEC-013]
  constraints:
    - stdout-is-for-data-stderr-is-for-humans
    - no-sql-in-cli-layer
    - no-cgo
    - test-before-implementation
    - errors-wrap-with-context
  related_specs: [SPEC-048, SPEC-020, SPEC-018, SPEC-019, SPEC-049]
---

# SPEC-051: `brag wrapped [year|quarter]` — the shareable year/quarter-in-review digest

## Context

`brag wrapped` is the headline polish feature of STAGE-013 (PROJ-004's closing
stage) — a **shareable, celebratory year- or quarter-in-review** digest: "your
year in brags." It is the fifth consumer of the DEC-014 rule-based digest
envelope, joining `summary`/`review`/`stats`/`impact`. It curates a *selection*
of the existing `internal/aggregate` toolbox into a retrospective highlight
reel over a **named calendar period**, and it reuses the calendar-window
infrastructure that `brag impact` established (DEC-028) — a year or a quarter
is a calendar period.

- Parent stage: `STAGE-013` (Polish + v0.4.0 cut). First spec in its backlog;
  the headline feature.
- Project: `PROJ-004` (the story surface). `wrapped` is the shareable/celebratory
  counterpart to the analytical `impact` digest.
- **Text-first by design.** The visual/sparklines pass is a SEPARATE later spec
  (**SPEC-052**). This spec designs `wrapped` as a clean text digest whose
  cadence data lives in a **named, sparkline-ready slot** (`cadence.series`) so
  SPEC-052 can render it as a sparkline WITHOUT reshaping the envelope. No
  sparklines here.
- **Quarterly is first-class** alongside the annual view — companies report by
  the quarter.
- New decision this spec emits: **DEC-030** (wrapped period selection + the
  section taxonomy / envelope). Confidence 0.72.

## Goal

Ship `brag wrapped [<year>] [Q<n>]` — a rule-based, deterministic (no LLM),
local-first digest that renders a celebratory review-to-share of the corpus
over a named calendar year or quarter, as markdown (default) or the DEC-014
JSON envelope (`--format json`). The default period (no argument) is the
**current calendar year**.

## Inputs

- **Files to read:**
  - `internal/cli/window.go` — the shared calendar-window core (`windowCutoff`,
    `selectedWindow`, `windowFlagNames`). `wrapped` reuses the calendar-period
    *concept* but needs a BOUNDED period (start AND end), so it adds its own
    period parser (`parseWrappedPeriod`) alongside — see Implementation Context.
  - `internal/cli/impact.go` — the `nowFunc` clock seam + `runImpact` shape
    (`wrapped` mirrors the CLI wiring: resolve period → `Store.List` → filter →
    render).
  - `internal/export/impact.go`, `internal/export/stats.go`,
    `internal/export/summary.go` — the DEC-014 renderer pattern (`ToXMarkdown` /
    `ToXJSON`, the `XOptions{Now, Scope, Filters, FiltersJSON}` struct,
    `trimTrailingNewline`, the struct-tag JSON envelope).
  - `internal/aggregate/aggregate.go` — the toolbox `wrapped` curates:
    `MostCommon`, `Streak`, `Span`, `WithImpact`, `GroupEntriesByProject`.
  - `DEC-014` (envelope), `DEC-028` (calendar-window semantics), `DEC-030`
    (this spec's decision).
- **Related code paths:** `internal/cli/`, `internal/export/`,
  `internal/aggregate/`, `internal/storage/entry.go` (`ListFilter.Since`).

## Outputs

- **Files created:**
  - `internal/export/wrapped.go` — `WrappedOptions`, `ToWrappedMarkdown`,
    `ToWrappedJSON`, plus the `Cadence` helper (bucketing + busiest-month).
  - `internal/export/wrapped_test.go` — the goldens + unit tests below.
  - `internal/cli/wrapped.go` — `NewWrappedCmd`, `runWrapped`,
    `parseWrappedPeriod`.
  - `internal/cli/wrapped_test.go` — the CLI-level tests below.
- **Files modified:**
  - `cmd/brag/main.go` (or wherever subcommands are registered) — register
    `NewWrappedCmd()`.
  - `internal/aggregate/aggregate.go` — add `CadenceBucket` + `Cadence(entries,
    months []string) ([]CadenceBucket, busiestMonth string)` (or keep the
    cadence bucketing in `export`; see LD8 — DECISION LOCKED to `aggregate`).
  - `docs/api-contract.md` — add the `brag wrapped` section.
  - `docs/tutorial.md` — surface `wrapped` in the digest family (§ that lists
    `summary`/`review`/`stats`/`impact`).
  - `README.md` — add `wrapped` to the command list if one is enumerated.
  - The STAGE-013 backlog line for SPEC-051 flips to a build state at build.
- **New exports:**
  - `aggregate.CadenceBucket{Period string; Count int}`
  - `aggregate.Cadence(entries []storage.Entry, months []string) ([]CadenceBucket, string)`
  - `export.WrappedOptions`, `export.ToWrappedMarkdown`, `export.ToWrappedJSON`
  - `cli.NewWrappedCmd`
- **Database changes:** none. Read-only over existing schema
  (`Store.List(ListFilter{Since})` + an in-CLI upper-bound filter).

## Acceptance Criteria

- [ ] `brag wrapped` (no argument) covers the **current calendar year**;
      `scope` renders as `"<year>"` (e.g. `"2026"`).
- [ ] `brag wrapped 2026` covers calendar year 2026; `brag wrapped 2026 Q3`
      covers calendar quarter Q3 2026; `scope` renders `"2026"` / `"2026-Q3"`.
- [ ] The period is **bounded on both ends**: an entry created the day before
      the period start is excluded; an entry created the day after the period
      end is excluded. (For a completed year/quarter the upper bound is NOT
      "now" — it is the period end. This is the load-bearing bounded-window
      assertion, distinct from `impact`'s `[cutoff, now]`.)
- [ ] Output is markdown by default, DEC-014 JSON envelope under
      `--format json`; an unknown `--format` is a `UserError` on stderr,
      stdout empty, non-zero exit.
- [ ] The markdown digest renders these sections in order: provenance (with a
      headline `Entries: N` line), `## Cadence` (busiest month + per-month
      bucket series), `## Top initiatives` (top-5 projects by count, excluding
      `(no project)`), `## Impact moments` (with-impact entries grouped by
      project, impact text in full), `## Rhythm` (longest streak, top-5 tags,
      top-3 types), `## Span` (first/last entry date + active days).
- [ ] The JSON envelope carries the DEC-014 provenance keys plus per-spec keys
      `total_entries`, `cadence` (`{busiest_month, series:[{period,count}]}`),
      `top_initiatives`, `impact_moments`, `longest_streak`, `top_tags`,
      `top_types`, `span`. Empty-state per DEC-014 part (4): arrays `[]`,
      objects `{}`, `busiest_month`/date fields `null`, numbers `0`.
- [ ] The **cadence series is present as data** (one bucket per month in scope,
      zero-filled) in BOTH renderers — text now, sparkline-ready for SPEC-052
      without reshaping.
- [ ] `--tag` / `--project` / `--type` filters compose with the period and echo
      into `Filters:` / `filters` exactly as `impact` does.
- [ ] Empty period (no entries in scope): provenance only (through `Entries: 0`);
      the `## Cadence`/`## Top initiatives`/… body sections are OMITTED from
      markdown (DEC-014 part 4). JSON still renders every key with empty-state
      values.
- [ ] Malformed period argument (bad year, `Q0`/`Q5`, extra tokens) is a
      `UserError` on stderr, stdout empty, non-zero exit.
- [ ] The digest is deterministic given `(entries, scope, now)` — the injected
      `nowFunc` seam makes goldens byte-stable.

## Failing Tests

Written during **design**, BEFORE build. Every golden below was computed at
design time against the REAL `internal/aggregate` helpers (via a scratch program
inside the module tree, since removed) — they are faithful, not hand-typed.

The shared fixture, `wrappedYearFixture` (7 entries over calendar 2026), is
defined once in `wrapped_test.go`:

```go
// wrappedYearFixture: 7 entries across calendar 2026. Exercises every
// section — cadence spread (Jan/Feb/Apr/Jul/Nov), projects alpha(3)/beta(2)/
// gamma(1) + one (no project), 3 with-impact entries (alpha x2, beta x1),
// a 2-day local streak (Jul 4 + Jul 5) as the longest run, varied tags/types.
var wrappedYearFixture = []storage.Entry{
	{ID: 1, Title: "kickoff", Project: "alpha", Type: "shipped", Tags: "auth,api",
		Impact:    "cut p95 login latency 40%",
		CreatedAt: time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)},
	{ID: 2, Title: "docs pass", Project: "beta", Type: "shipped", Tags: "docs",
		CreatedAt: time.Date(2026, 2, 3, 10, 0, 0, 0, time.UTC)},
	{ID: 3, Title: "migration", Project: "alpha", Type: "shipped", Tags: "db,api",
		Impact:    "removed the nightly cron entirely",
		CreatedAt: time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC)},
	{ID: 4, Title: "reflection", Project: "", Type: "learned", Tags: "process",
		CreatedAt: time.Date(2026, 4, 22, 10, 0, 0, 0, time.UTC)},
	{ID: 5, Title: "launch", Project: "beta", Type: "shipped", Tags: "api",
		Impact:    "onboarding time down to 1 day",
		CreatedAt: time.Date(2026, 7, 4, 10, 0, 0, 0, time.UTC)},
	{ID: 6, Title: "hotfix", Project: "alpha", Type: "fixed", Tags: "auth",
		CreatedAt: time.Date(2026, 7, 5, 10, 0, 0, 0, time.UTC)},
	{ID: 7, Title: "retro notes", Project: "gamma", Type: "learned", Tags: "process,docs",
		CreatedAt: time.Date(2026, 11, 30, 10, 0, 0, 0, time.UTC)},
}
var wrappedYearNow = time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC)
```

### `internal/export/wrapped_test.go`

- **`TestToWrappedMarkdown_DEC014FullDocumentGolden`** (LOAD-BEARING, written
  FIRST). Byte-exact assertion of the full markdown document over
  `wrappedYearFixture` with `WrappedOptions{Scope:"2026", Filters:"(none)",
  FiltersJSON:nil, ScopeMonths: <the 12 month labels 2026-01..2026-12>,
  Now:wrappedYearNow}`. The renderer receives the already-in-period slice (the
  CLI does the windowing) plus the ordered `ScopeMonths` labels. `want`:

```
# Bragfile Wrapped

Generated: 2026-12-31T23:59:59Z
Scope: 2026
Filters: (none)
Entries: 7

## Cadence

Busiest month: 2026-04 (2)

- 2026-01: 1
- 2026-02: 1
- 2026-03: 0
- 2026-04: 2
- 2026-05: 0
- 2026-06: 0
- 2026-07: 2
- 2026-08: 0
- 2026-09: 0
- 2026-10: 0
- 2026-11: 1
- 2026-12: 0

## Top initiatives

- alpha: 3
- beta: 2
- gamma: 1

## Impact moments

### alpha

- 1: kickoff
  cut p95 login latency 40%
- 3: migration
  removed the nightly cron entirely

### beta

- 5: launch
  onboarding time down to 1 day

## Rhythm

Longest streak: 2 days

**Top tags**
- api: 3
- auth: 2
- docs: 2
- process: 2
- db: 1

**Top types**
- shipped: 4
- learned: 2
- fixed: 1

## Span

- First entry: 2026-01-15
- Last entry: 2026-11-30
- Active days: 320
```

  (Returned bytes have the trailing `\n` stripped, matching every other
  renderer's contract.)

- **`TestToWrappedJSON_DEC030ShapeGolden`** (LOAD-BEARING). Byte-exact JSON
  over the same fixture/opts (`FiltersJSON:nil → {}`). `want`:

```json
{
  "generated_at": "2026-12-31T23:59:59Z",
  "scope": "2026",
  "filters": {},
  "total_entries": 7,
  "cadence": {
    "busiest_month": "2026-04",
    "series": [
      {
        "period": "2026-01",
        "count": 1
      },
      {
        "period": "2026-02",
        "count": 1
      },
      {
        "period": "2026-03",
        "count": 0
      },
      {
        "period": "2026-04",
        "count": 2
      },
      {
        "period": "2026-05",
        "count": 0
      },
      {
        "period": "2026-06",
        "count": 0
      },
      {
        "period": "2026-07",
        "count": 2
      },
      {
        "period": "2026-08",
        "count": 0
      },
      {
        "period": "2026-09",
        "count": 0
      },
      {
        "period": "2026-10",
        "count": 0
      },
      {
        "period": "2026-11",
        "count": 1
      },
      {
        "period": "2026-12",
        "count": 0
      }
    ]
  },
  "top_initiatives": [
    {
      "project": "alpha",
      "count": 3
    },
    {
      "project": "beta",
      "count": 2
    },
    {
      "project": "gamma",
      "count": 1
    }
  ],
  "impact_moments": [
    {
      "project": "alpha",
      "entries": [
        {
          "id": 1,
          "title": "kickoff",
          "project": "alpha",
          "impact": "cut p95 login latency 40%"
        },
        {
          "id": 3,
          "title": "migration",
          "project": "alpha",
          "impact": "removed the nightly cron entirely"
        }
      ]
    },
    {
      "project": "beta",
      "entries": [
        {
          "id": 5,
          "title": "launch",
          "project": "beta",
          "impact": "onboarding time down to 1 day"
        }
      ]
    }
  ],
  "longest_streak": 2,
  "top_tags": [
    {
      "name": "api",
      "count": 3
    },
    {
      "name": "auth",
      "count": 2
    },
    {
      "name": "docs",
      "count": 2
    },
    {
      "name": "process",
      "count": 2
    },
    {
      "name": "db",
      "count": 1
    }
  ],
  "top_types": [
    {
      "name": "shipped",
      "count": 4
    },
    {
      "name": "learned",
      "count": 2
    },
    {
      "name": "fixed",
      "count": 1
    }
  ],
  "span": {
    "first_entry_date": "2026-01-15",
    "last_entry_date": "2026-11-30",
    "active_days": 320
  }
}
```

- **`TestToWrappedMarkdown_QuarterGolden`** (LOAD-BEARING for the quarter path).
  Fixture `[]storage.Entry{wrappedYearFixture[4], wrappedYearFixture[5]}` (launch
  + hotfix, both in Jul), `WrappedOptions{Scope:"2026-Q3",
  ScopeMonths:["2026-07","2026-08","2026-09"], Filters:"(none)",
  Now: time.Date(2026,9,30,23,59,59,0,time.UTC)}`. `want`:

```
# Bragfile Wrapped

Generated: 2026-09-30T23:59:59Z
Scope: 2026-Q3
Filters: (none)
Entries: 2

## Cadence

Busiest month: 2026-07 (2)

- 2026-07: 2
- 2026-08: 0
- 2026-09: 0

## Top initiatives

- alpha: 1
- beta: 1

## Impact moments

### beta

- 5: launch
  onboarding time down to 1 day

## Rhythm

Longest streak: 2 days

**Top tags**
- api: 1
- auth: 1

**Top types**
- fixed: 1
- shipped: 1

## Span

- First entry: 2026-07-04
- Last entry: 2026-07-05
- Active days: 2
```

  (Note: `Top types` shows `fixed` before `shipped` — both count 1, so
  `MostCommon` breaks the tie alpha-ASC. Faithful to the helper.)

- **`TestToWrapped_EmptyPeriodShape`**. `WrappedOptions{Scope:"2026-Q3",
  ScopeMonths:["2026-07","2026-08","2026-09"], Filters:"(none)",
  Now: time.Date(2026,9,30,23,59,59,0,time.UTC)}` over `nil` entries.
  - Markdown `want` (body omitted per DEC-014 part 4):

```
# Bragfile Wrapped

Generated: 2026-09-30T23:59:59Z
Scope: 2026-Q3
Filters: (none)
Entries: 0
```

  - Also assert `!strings.Contains(md, "## Cadence")`.
  - JSON: unmarshal to `map[string]json.RawMessage`; assert `total_entries=="0"`,
    `cadence` is `{"busiest_month":null,"series":[...3 zero buckets...]}` (the
    scope months still render, zero-filled — the sparkline slot is present even
    when empty), `top_initiatives=="[]"`, `impact_moments=="[]"`,
    `longest_streak=="0"`, `top_tags=="[]"`, `top_types=="[]"`,
    `span=={"first_entry_date":null,"last_entry_date":null,"active_days":0}`,
    `filters=="{}"`. (Empty-period `busiest_month` is `null`; the series is
    still fully present so SPEC-052 renders a flat sparkline, not a missing one.)

- **`TestToWrapped_ImpactTextRenderedInFull`**. A single with-impact entry with
  a long impact string; assert the full impact text appears in both markdown
  (the indented line) and JSON (the `impact` key) — never elided (mirrors
  `impact`'s Test 5).

- **`TestToWrapped_FiltersEchoed`**. `Filters:"--project alpha"`,
  `FiltersJSON:{"project":"alpha"}`; assert `Filters: --project alpha` in
  markdown and `filters.project=="alpha"` (and only that key) in JSON.

- **`TestCadence_ZeroFilledAndBusiest`** (aggregate). `aggregate.Cadence` over a
  3-entry input across two of three scope months returns one bucket per scope
  month in order (zero-filled for the empty month) and the busiest-month label;
  on empty input returns the zero-filled series + `""` busiest.

### `internal/cli/wrapped_test.go`

(Harness mirrors `impact_test.go`: `newWrappedTestRoot`, `runWrappedCmd`,
`seedWrappedEntry` — raw SQL confined to the test file so
`no-sql-in-cli-layer` holds — and `withNowFunc` swapping the `nowFunc` seam.)

- **`TestWrappedCmd_DefaultsToCurrentYear`**. With `nowFunc` frozen at
  2026-07-06 and a seeded entry in Jan 2026, `brag wrapped` (no arg) emits
  `Scope: 2026`. Asserts the default period is the current calendar year.

- **`TestWrappedCmd_NamedYearAndQuarter`**. `brag wrapped 2026` → `Scope: 2026`;
  `brag wrapped 2026 Q3` → `Scope: 2026-Q3`. Case-insensitive `q3` accepted
  (LD3).

- **`TestWrappedCmd_BoundedWindow`** (LOAD-BEARING — the bounded-window
  divergence from `impact`). Seed three entries: one on 2025-12-31 (before the
  2026 period), one on 2026-06-15 (in period), one on 2027-01-01 (after the
  period). `brag wrapped 2026` includes ONLY the mid entry: `Entries: 1`, and
  the 2025 + 2027 titles are absent. This is the assertion that distinguishes
  `wrapped`'s bounded `[period-start, period-end]` from `impact`'s
  `[cutoff, now]`. `nowFunc` is frozen at 2027-06-01 so the run is "after" the
  named year and the upper bound is proven to be the period end, not now.

- **`TestWrappedCmd_QuarterBoundaries`**. Seed entries on 2026-06-30 (Q2),
  2026-07-01 (Q3 start), 2026-09-30 (Q3 end), 2026-10-01 (Q4). `brag wrapped
  2026 Q3` includes exactly the Jul-1 and Sep-30 entries (`Entries: 2`);
  excludes the Jun-30 and Oct-1 ones.

- **`TestWrappedCmd_MalformedPeriodIsUserError`**. Table: `["999"]` (year too
  short / out of range per LD3), `["2026", "Q0"]`, `["2026", "Q5"]`,
  `["2026", "Q3", "extra"]`, `["notayear"]`. Each → non-nil error that is a
  `UserError` (assert via `errors.As`), stdout empty, message on stderr.

- **`TestWrappedCmd_UnknownFormatIsUserError`**. `brag wrapped 2026 --format
  yaml` → `UserError`, stdout empty.

- **`TestWrappedCmd_JSONWiring`**. `brag wrapped 2026 --format json` over a
  seeded corpus round-trips through `encoding/json` and carries `scope=="2026"`
  and a `cadence.series` of length 12. (Light wiring check; the byte-goldens
  live in the export test.)

- **`TestWrappedCmd_HelpShowsExamples`**. `brag wrapped --help` contains the
  literal `Examples:` label and the exact example line `brag wrapped 2026 Q3`
  (unique-token discipline, AGENTS.md §9).

- **`TestWrappedCmd_StdoutStderrSeparation`**. A successful run writes the
  digest to stdout and leaves stderr empty (`errBuf.Len()==0`); the malformed
  case writes only to stderr (per `stdout-is-for-data-stderr-is-for-humans`).

## Implementation Context

*Read this section (and the files it points to) before starting the build
cycle.*

### Decisions that apply

- `DEC-030` (this spec) — locks the wrapped **period selection** (positional
  `[<year>] [Q<n>]`; default = current calendar year; bounded `[period-start,
  period-end]`), the **section taxonomy** (Cadence / Top initiatives / Impact
  moments / Rhythm / Span), and the **envelope keys** + the sparkline-ready
  `cadence.series` slot.
- `DEC-014` — the rule-based digest envelope `wrapped` extends verbatim
  (single-object JSON, flat top-level keys, `scope`/`filters` provenance,
  markdown provenance-then-body, 2-space indent, empty-state part 4).
- `DEC-028` — the calendar-window semantics `wrapped` reuses the *concept* of
  (project=initiative grouping; impact-first `WithImpact`+`GroupEntriesByProject`
  for the Impact-moments section; calendar periods via `time.Date` constructors,
  never day subtraction). `wrapped` DIVERGES from `impact` on the upper bound:
  `impact` is `[cutoff, now]`; `wrapped` is `[period-start, period-end]`
  (bounded). DEC-030 records this as the deliberate extension.
- `DEC-022` — `Streak` buckets by local calendar day off the injected `now`.
  `wrapped` surfaces only **longest** streak (period-scoped over the passed
  entries), which is independent of `now` — so the `now` value affects only the
  `Generated:` line here, not the streak number. Note this in `wrapped.go`.
- `DEC-008` — `ParseSince` is NOT used by `wrapped` (no `--since`); the period
  parser is positional. Listed for contrast.
- `DEC-007` — required/invalid-flag validation via `UserErrorf` (malformed
  period, unknown `--format`).
- `DEC-013` — count-ordering (DESC by count, alpha-ASC tiebreak, `(no project)`
  last) inherited transitively via the aggregate helpers.

### Constraints that apply

- `stdout-is-for-data-stderr-is-for-humans` — digest body to stdout; UserErrors
  to stderr. Tested by `TestWrappedCmd_StdoutStderrSeparation`.
- `no-sql-in-cli-layer` — `internal/cli/wrapped.go` must not import
  `database/sql`. The read path is `Store.List(ListFilter{Since: periodStart})`;
  the **upper-bound** filter (`created_at <= periodEnd`) is applied in Go in the
  CLI layer (`ListFilter` has no `Until` field — see Notes). Raw SQL in the test
  file (for `seedWrappedEntry`'s `created_at` rewrite) is confined to the test,
  matching `impact_test.go`.
- `no-cgo` — no new deps; pure-Go stdlib only.
- `test-before-implementation` — the Failing Tests above are written first.
- `errors-wrap-with-context` — storage/config errors wrapped
  `fmt.Errorf("...: %w", err)`; period/format user errors use `UserErrorf`.

### Prior related work

- `SPEC-048` (shipped) — `brag impact`; the fourth DEC-014 consumer, DEC-028,
  `internal/export/impact.go`, `internal/cli/impact.go`, the `nowFunc` seam and
  the `windowCutoff` calendar core. `wrapped` is the closest structural sibling —
  copy its CLI + renderer shape.
- `SPEC-020` (shipped) — `brag stats`; the `Streak`/`MostCommon`/`Span` helpers
  and the `extractTags`/`extractProjects` (non-empty) helpers `wrapped` reuses,
  plus the `corpusSpanRecord` date-or-null pattern the `span` sub-object mirrors.
- `SPEC-018`/`SPEC-019` — the first DEC-014 consumers and `GroupEntriesByProject`.

### Out of scope (for this spec specifically)

- **Sparklines / any visual rendering** — SPEC-052. This spec ONLY guarantees
  the `cadence.series` data slot exists and is sparkline-ready.
- **`--previous`** (last-completed period) — SPEC-053. `wrapped`'s default is
  the *current* calendar year; the completed-period variant is `--previous`'s
  job, deliberately not baked into the default (mirrors DEC-028's current-vs-
  previous deferral).
- **`--out <file>`** — the whole digest family defers file output; `> file`
  works.
- Any LLM call-out; any network; any schema change.
- Weekly / monthly `wrapped` scopes — year + quarter only (the stage brief's
  scope). If a monthly wrapped is wanted later, it is a new spec.

## Notes for the Implementer

- **Period parser (`parseWrappedPeriod(args []string, now time.Time)`)** returns
  `(scope string, months []string, start, end time.Time, err error)`:
  - 0 args → current calendar year (`now.Year()`).
  - 1 arg → a 4-digit year (`^\d{4}$`, plausibility-bounded e.g. 2000–2999);
    anything else is a `UserError`.
  - 2 args → year + quarter token `Q<n>` (case-insensitive, `n∈1..4`); `Q0`/`Q5`/
    non-`Q` is a `UserError`.
  - 3+ args → `UserError` (extra tokens).
  - `start` = `time.Date(year, startMonth, 1, 0,0,0,0, UTC)`; `end` = the
    instant just before the next period starts (`start.AddDate(1,0,0)` for a
    year, `start.AddDate(0,3,0)` for a quarter) — compute the **exclusive** next
    boundary and filter `created_at < nextBoundary` to avoid a last-second
    off-by-one. `months` = the ordered `YYYY-MM` labels in scope (12 for a year,
    3 for a quarter). `scope` = `"2026"` or `"2026-Q3"`.
- **Windowing** (CLI layer, no SQL): `entries, _ := s.List(ListFilter{Since:
  start})` then keep only those with `e.CreatedAt.Before(nextBoundary)`. Pass the
  filtered slice + `months` to the renderer. (`ListFilter` has only a `Since`
  lower bound; the upper bound is a trivial Go filter and stays SQL-free.)
- **`aggregate.Cadence(entries, months)`** buckets by `created_at.UTC().Format(
  "2006-01")`, then emits one `CadenceBucket{Period, Count}` per label in
  `months` order (zero-filled), plus the busiest-month label (first max wins;
  `""` when the series is all-zero). This is the sparkline-ready slot — SPEC-052
  reads `series[].count` to render `▁▂▃▄▅▆▇█`. LD8: this lives in `aggregate`
  (SQL-free, pure), NOT `export`, so SPEC-052 and any future `stats` cadence
  reuse it.
- **Renderer** (`ToWrappedMarkdown` / `ToWrappedJSON`, mirroring
  `impact.go`/`stats.go`): `WrappedOptions{Scope, Filters, FiltersJSON,
  ScopeMonths []string, Now time.Time}`. Curate: `aggregate.Cadence(entries,
  ScopeMonths)`; `aggregate.MostCommon(extractProjects(entries), 5)` for
  Top initiatives (excludes `(no project)`, exactly like `stats`' top_projects);
  `aggregate.WithImpact` + `aggregate.GroupEntriesByProject` for Impact moments
  (impact text in full, `impact.go` shape); `aggregate.Streak(entries, Now)` →
  use ONLY `longest`; `aggregate.MostCommon(extractTags(entries),5)` and
  `aggregate.MostCommon(extractTypes(entries),3)` for Rhythm; `aggregate.Span`
  for Span. `extractTags`/`extractProjects` already exist in `export/stats.go` —
  reuse; add `extractTypes` (non-empty `Type` values) next to them.
- **Empty-state**: when `len(entries)==0`, markdown emits provenance through
  `Entries: 0` and returns (no body) — exactly DEC-014 part 4. JSON always emits
  every key; `cadence.series` is still the full zero-filled month series (the
  slot is present so SPEC-052 renders a flat line, not a gap); `busiest_month`
  and the two span dates are `null`; arrays `[]`; numbers `0`.
- **JSON key order** (locked by struct-tag declaration order): `generated_at`,
  `scope`, `filters`, `total_entries`, `cadence`, `top_initiatives`,
  `impact_moments`, `longest_streak`, `top_tags`, `top_types`, `span`. Nested:
  `cadence{busiest_month, series}`, each bucket `{period, count}`; each
  top-initiative `{project, count}`; each impact group `{project,
  entries:[{id,title,project,impact}]}`; each tag/type `{name, count}`; span
  `{first_entry_date, last_entry_date, active_days}`. Use `*string` for
  `busiest_month` and the two span dates so empty renders `null`.
- **Registration**: add `NewWrappedCmd()` alongside the other subcommands where
  they are registered.
- **Docs sweep** (AGENTS.md §9 status-change rule): grep `docs/ README.md
  AGENTS.md` for the digest-family list (`summary`/`review`/`stats`/`impact`)
  and add `wrapped` wherever the family is enumerated; add the `wrapped` §11
  glossary term and a `docs/api-contract.md` section. Enumerate the actual hits
  in `## Outputs` at build after running the grep (design lists the files;
  build re-verifies the exact lines per §12 audit-grep cross-check).

### Locked design decisions (build-time)

1. **LD1 — Positional period, not flags.** `brag wrapped [<year>] [Q<n>]`, not
   `--year/--quarter`. A "wrapped" names a period the way a human says it
   ("2026", "2026 Q3"); positional reads cleaner for the headline shareable
   command and avoids overloading `impact`'s `--year/--quarter` bool flags
   (which mean "current year/quarter up to now" — a DIFFERENT semantics).
   *Rejected:* reusing `impact`'s `--quarter/--year` bool flags — they carry
   the `[cutoff, now]` current-period meaning; `wrapped` names a bounded past
   period. Overloading them would conflate two window semantics on one flag.
2. **LD2 — Default = current calendar year.** No argument → the current year,
   `[Jan-1, now-ish-but-actually-Dec-31-exclusive-next-year]`... precisely: the
   full current calendar year bounded `[Jan-1 00:00, next-Jan-1 00:00)`.
   *Rejected:* default = last-completed year/quarter. That is the retrospective
   framing, but it is exactly what `--previous` (SPEC-053) is planned to add;
   baking last-completed into the default now would preempt SPEC-053 and split
   the "which period" logic across two specs. Current-year default mirrors
   DEC-028's current-in-progress choice, keeping the digest family consistent;
   `--previous` layers cleanly on top later.
3. **LD3 — Year is 4 digits, plausibility-bounded; quarter is `Q<1..4>`,
   case-insensitive.** `^\d{4}$` in `[2000,2999]`; `q3`/`Q3` both accepted;
   `Q0`/`Q5`/`3`/`QQ` are `UserError`. *Rejected:* accepting 2-digit years
   (`26`) — ambiguous; accepting bare `3` for quarter — collides with a
   year-shaped arg.
4. **LD4 — Bounded `[period-start, next-boundary)` window, upper bound is the
   period end, NOT now.** The load-bearing divergence from `impact`. Filter
   `Since: start` in SQL, then `created_at < nextBoundary` in Go. *Rejected:*
   `[start, now]` like `impact` — wrong for a named past period (would include
   nothing after "now" but ALSO would wrongly include everything up to now when
   the named period is the current year and the run happens next year; the
   bounded upper edge is the correct, testable semantics).
5. **LD5 — Section taxonomy: Cadence → Top initiatives → Impact moments →
   Rhythm → Span.** A celebratory arc: how often you showed up (cadence),
   where (initiatives), what mattered (impact moments), your habits (rhythm),
   the bookends (span). *Rejected:* leading with raw totals (that is `stats`'
   analytical framing); leading with impact (that is `impact`'s job). Cadence-
   first is the "wrapped" feel — the shape of your year.
6. **LD6 — Top initiatives excludes `(no project)`; caps at 5.** Reuses
   `MostCommon(extractProjects,5)` (same helper + exclusion as `stats`'
   top_projects) — a "top initiatives" reel with `(no project)` in it reads as
   noise in a shareable digest. *Rejected:* `ByProject` (includes `(no
   project)` last) — right for `summary`'s completeness, wrong for a curated
   reel.
7. **LD7 — Rhythm surfaces LONGEST streak only, not current.** A wrapped is a
   retrospective of a *finished/named* period; "current streak" is a live-corpus
   metric (`stats`' job) and would be meaningless scoped to a past period.
   *Rejected:* showing current streak — semantically wrong for a period digest.
8. **LD8 — `Cadence` lives in `internal/aggregate`, not `internal/export`.**
   It is pure, SQL-free, and SPEC-052 (+ a future `stats` cadence) will reuse
   it. *Rejected:* inlining the bucketing in `export/wrapped.go` — would force
   SPEC-052 to either duplicate or refactor it out; put it in the toolbox now.
9. **LD9 — `--format` default is `"markdown"`** (states the default explicitly
   per §12 flag-default rule), accepted values `markdown`/`json`, unknown →
   `UserError`. Mirrors `impact`/`stats`.

### Rejected alternatives (build-time)

- **A `Cadence` in months for a year but weeks for a quarter.** Rejected:
  keep the bucket unit uniform (calendar month) across both scopes so
  `series[].period` is always `YYYY-MM` and SPEC-052 renders one sparkline
  shape. A quarter is 3 monthly buckets. (A future weekly cadence is a
  SPEC-052/own-spec call, not baked here.)
- **Including a `by_type` count block like `summary`.** Rejected: `Rhythm`'s
  top-3 types covers the celebratory need; a full by-type block is `summary`/
  `stats` analytical surface. Keep `wrapped` curated, not a metrics dump.
- **Reusing `ByProject` for Top initiatives to get `(no project)` last.**
  Covered in LD6 — excluded deliberately.
- **A `--year`/`--quarter` flag surface.** Covered in LD1.

---

## Build Completion

*Filled in at the end of the **build** cycle, before advancing to verify.*

- **Branch:** feat/spec-051-brag-wrapped
- **PR (if applicable):** #89 (draft → ready-for-review at build close)
- **All acceptance criteria met?** Yes. All 10 acceptance criteria verified by
  the Failing Tests (all passing) plus a live smoke of `brag wrapped 2026` /
  `brag wrapped 2026 Q3` (year JSON + malformed `Q5` → user error). Every golden
  matched the REAL `aggregate`/renderer output on FIRST implementation — no
  golden was edited, no escalation needed (a pre-build scratch test confirmed the
  span=320, longest-streak=2, top-tags/types/initiatives orders, quarter span=2,
  and `fixed`-before-`shipped` tie-break all come from the real helpers before a
  line of renderer was written).
- **New decisions emitted:**
  - `DEC-030` — wrapped period selection + section taxonomy + envelope (emitted
    at design; no new DEC needed at build — DEC-031 stays free).
- **Files changed:**
  - `internal/aggregate/aggregate.go` — `CadenceBucket{Period,Count}` (with
    `json:"period"/"count"` tags — see Deviations) + `Cadence(entries, months)`.
  - `internal/export/wrapped.go` (new) — `WrappedOptions`, `ToWrappedMarkdown`,
    `ToWrappedJSON`, `extractTypes`.
  - `internal/export/wrapped_test.go` (new) — the two byte-goldens + quarter +
    empty + impact-in-full + filters-echoed + `TestCadence_ZeroFilledAndBusiest`.
  - `internal/cli/wrapped.go` (new) — `NewWrappedCmd`, `runWrapped`,
    `parseWrappedPeriod`, `parseYearArg`, `parseQuarterArg`, `monthLabels`,
    `echoFiltersForWrapped`.
  - `internal/cli/wrapped_test.go` (new) — the 9 CLI-level tests incl.
    `TestWrappedCmd_BoundedWindow`.
  - `cmd/brag/main.go` — registered `NewWrappedCmd()`.
  - `docs/api-contract.md` — added the `brag wrapped` section.
  - `docs/tutorial.md` — added the `### Your year in brags: brag wrapped` section.
  - `README.md` — added `brag wrapped` to the digest command list.
  - `AGENTS.md` — added the `wrapped` §11 glossary term.
  - `guidance/questions.yaml` — resolved `wrapped-default-current-vs-last-completed`
    (current-calendar-year, per orchestrator sign-off).
- **Docs sweep (§9 status-change / §12 audit-grep cross-check):** grepped
  `docs/ README.md AGENTS.md` for the digest-family enumeration. The command-list
  hits that get `wrapped` added: `README.md` command block, `docs/tutorial.md`
  digest section (new subsection between `impact` and `story`),
  `docs/api-contract.md` command sections (new `### brag wrapped` section),
  `AGENTS.md` §11 glossary. NOT touched: the DEC-014 *inventory* sentences in
  `docs/api-contract.md:916` and `docs/data-model.md:217` enumerate the DEC-014
  consumers as of DEC-014's authorship (`summary/review/stats/impact`) — those
  are historical DEC-provenance lines, not a live "current command family" list,
  so adding `wrapped` there would misattribute it to DEC-014's era. Left as-is;
  `test-docs`/`test-hook` both green.
- **Deviations from spec:**
  - The JSON `{period,count}` key casing lives as struct tags on
    `aggregate.CadenceBucket` itself, not on a separate `export`-side projection.
    The spec (LD8 / Notes) locks the bucket in `aggregate` AND locks the golden
    to lowercase `period`/`count`; since `export` embeds the aggregate struct
    directly (rather than reprojecting), the tags must ride on the aggregate
    struct for the golden to hold. This is the minimal faithful reading of both
    locks together; the alternative (a parallel export-side bucket type) would
    duplicate the shape SPEC-052 is meant to reuse, defeating LD8. Not a
    semantic deviation — same wire shape the golden specifies.
  - No other deviations. Section order, envelope keys, empty-state, bounded
    window, and both byte-goldens are exactly as specified.
- **Follow-up work identified:**
  - None new. SPEC-052 (sparklines over `cadence.series`) and SPEC-053
    (`--previous`) were already foreseen; the slot + bounded-window machinery
    they need are in place.

### Build-phase reflection (3 questions, short answers)

1. **What was unclear in the spec that slowed you down?** — Almost nothing. The
   one thing that surfaced at first `go test` was the JSON key casing: the spec
   locks the cadence bucket in `aggregate` (LD8) and locks the golden to
   lowercase `period`/`count`, but doesn't say where the json tags live. Because
   the renderer embeds the aggregate struct directly, the tags had to go on the
   aggregate struct — a two-second fix, but the spec could have named it.
2. **Was there a constraint or decision that should have been listed but
   wasn't?** — No. The reference set (DEC-014/028/030/022/007/013, the five
   constraints) was complete; `impact.go`/`stats.go` gave the exact renderer +
   CLI shape to mirror, and the `nowFunc`/`seedImpactEntry`/`withNowFunc` test
   harness copied over cleanly.
3. **If you did this task again, what would you do differently?** — Add the json
   struct tags to `CadenceBucket` in the same edit that creates it, instead of
   discovering the casing at the first golden run. The design-time faithfulness
   pass (the removed scratch program) had already computed the values; running a
   scratch *marshal* of the bucket too would have caught the casing before build.

---

## Reflection (Ship)

*Appended during the **ship** cycle.*

1. **What would I do differently next time?** — <answer>
2. **Does any template, constraint, or decision need updating?** — <answer>
3. **Is there a follow-up spec I should write now before I forget?** — <answer>
