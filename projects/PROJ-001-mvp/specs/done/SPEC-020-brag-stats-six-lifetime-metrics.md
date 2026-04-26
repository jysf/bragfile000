---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-020
  type: story                      # epic | story | task | bug | chore
  cycle: ship
  blocked: false
  priority: medium
  complexity: S                    # S — consumes DEC-014 verbatim, three new helpers, one new command, one new render file. No DEC emission, no new package.

project:
  id: PROJ-001
  stage: STAGE-004
repo:
  id: bragfile

agents:
  architect: claude-opus-4-7
  implementer: claude-opus-4-7     # usually same Claude, different session
  created_at: 2026-04-25

references:
  decisions:
    - DEC-006   # cobra framework — new `brag stats` subcommand
    - DEC-007   # required-flag validation in RunE — `--format` uses UserErrorf
    - DEC-013   # markdown export shape — provenance-block convention DEC-014 inherited and SPEC-020 reuses
    - DEC-014   # CONSUMED VERBATIM (third consumer) — single-object envelope (JSON) + provenance/summary block (markdown)
    - DEC-004   # tags comma-joined TEXT — splitting at the count-by-tag aggregation boundary
  constraints:
    - no-sql-in-cli-layer
    - stdout-is-for-data-stderr-is-for-humans
    - errors-wrap-with-context
    - test-before-implementation
    - one-spec-per-pr
  related_specs:
    - SPEC-007   # shipped; ListFilter struct + Store.List read path; stats uses ListFilter{} (lifetime, no filters)
    - SPEC-014   # shipped; emitted DEC-011 + created internal/export package; structural mirror for goldens-locked DEC pattern
    - SPEC-018   # shipped; emitted DEC-014, seeded internal/aggregate, rangeCutoff helper precedent SPEC-020 does NOT use (no --range)
    - SPEC-019   # shipped; second DEC-014 consumer; Long-vs-help self-audit deviation captured in build reflection; SPEC-020 applies the same watch-pattern proactively at design (the second confirming case)
---

# SPEC-020: `brag stats` — six lifetime metrics consuming DEC-014

## Context

Third (and last) of three specs in STAGE-004. SPEC-018 shipped 2026-04-25
with the load-bearing pieces of the stage in place: DEC-014 (rule-based
output shape — single-object JSON envelope plus provenance/summary
markdown convention), the `internal/aggregate` package (`ByType`,
`ByProject`, `GroupForHighlights`, `rangeCutoff`), and the first
DEC-014 consumer (`brag summary`). SPEC-019 shipped the second consumer
(`brag review`) on the same day, extending `internal/aggregate` by one
helper. SPEC-020 is the third and final consumer.

`brag stats` is the **chart-of-yourself panorama** of the stage. Where
`brag summary` is the lightweight digest ("what happened this period?")
and `brag review` is the reflection ritual ("here's last week, now
think about it"), `brag stats` is the lifetime-corpus aggregation:
"how often have I been logging? how long is my current streak? what
am I working on most?" Six metrics rendered as either markdown
(default) or the DEC-014 JSON envelope, no LLM, no filters, no
`--range` — just the entire corpus.

The spec does **three** things in one pass — same shape as SPEC-019, no
DEC emission:

1. **Adds `brag stats [--format markdown|json]`** as the third DEC-014
   consumer. Only flag is `--format` (default markdown). NO filter
   flags (`--tag`/`--project`/`--type` deliberately rejected per stage
   Design Notes "Filter flag reuse"). NO `--range` (lifetime corpus
   only per stage Design Notes "Six metrics, lifetime corpus"). NO
   `--out` (stdout only per stage Design Notes "Output destination").

2. **Renders via `internal/export/stats.go`** (new file, sibling to
   `summary.go` + `review.go` + `markdown.go` + `json.go`). Consumes
   DEC-014's envelope verbatim. Per-spec payload keys at top level:
   `total_count`, `entries_per_week`, `current_streak`,
   `longest_streak`, `top_tags`, `top_projects`, `corpus_span` (the
   last is a sub-object with `first_entry_date`, `last_entry_date`,
   `days`).

3. **Extends `internal/aggregate` with three helpers** —
   `Streak(entries, now) (current, longest int)` (test-injectable
   `now` for date-boundary determinism per stage Design Notes
   "Time-of-computation determinism"), `MostCommon(values, n)
   []NameCount` (pure top-N counter; renderer adapts the result for
   `top_tags` and `top_projects`), `Span(entries) CorpusSpan` (first
   entry, last entry, days inclusive).

DEC-014 is **consumed verbatim** — no DEC emission this spec, no
re-litigation of the six locked choices. SPEC-018's load-bearing
goldens (`TestToSummaryMarkdown_DEC014FullDocumentGolden` and
`TestToSummaryJSON_DEC014ShapeGolden`) prove the envelope works on
the summary side; SPEC-019's mirror goldens
(`TestToReviewMarkdown_DEC014FullDocumentGolden` and
`TestToReviewJSON_DEC014ShapeGolden`) prove the review side composes
the same envelope; SPEC-020's goldens
(`TestToStatsMarkdown_DEC014FullDocumentGolden` and
`TestToStatsJSON_DEC014ShapeGolden`) prove the stats side composes
the same envelope. If the goldens ever fail in a way that implies
the envelope changed, that's a DEC-014 violation — fix the code,
not the test.

A deliberate per-spec-payload-shape divergence within DEC-014's
envelope: `top_tags` and `top_projects` are arrays-of-objects (NOT
maps) because their semantic ordering is DESC-by-count — Go's
`encoding/json` would alpha-sort a map key set at marshal time and
lose that ordering. SPEC-018's `counts_by_type` / `counts_by_project`
map shape works because lookup-by-key is the consumer's purpose
there; for top-N consumers the order IS the meaning, not metadata.
Two map-shape payloads (SPEC-018) coexist with two array-shape
payloads (SPEC-020) inside the same DEC-014 envelope — documented
here so future readers see it as a deliberate per-spec call, not a
shape drift.

Parent stage:
[`STAGE-004-rule-based-polish-summary-review-stats.md`](../stages/STAGE-004-rule-based-polish-summary-review-stats.md) —
Spec Backlog → SPEC-020 entry (lines ~225–232); Design Notes →
"DEC-014" / "`internal/aggregate` package" / "Filter flag reuse" /
"Output destination" / "SPEC-020-specific (`brag stats`)" (lines
~395–419). Project: PROJ-001 (MVP).

## Goal

Ship `brag stats [--format markdown|json]` as the third DEC-014
consumer: render six lifetime aggregations (total count, entries/week
rolling average, current streak, longest streak, top-5 most-common
tags, top-5 most-common projects, corpus span) over the entire
corpus, consuming DEC-014's envelope verbatim, with `top_tags` /
`top_projects` as arrays-of-objects to preserve DESC-by-count
ordering and `corpus_span` as a sub-object with date-or-null fields.

## Inputs

- **Files to read:**
  - `/AGENTS.md` — §6 cycle rules; §7 spec anatomy; §8 DEC discipline
    (note: SPEC-020 does NOT emit a DEC; consumed-verbatim DEC-014);
    §9 premise-audit family (SPEC-020 is **addition + status-change**
    — apply the §9 audit-grep cross-check addendum SPEC-018 earned and
    SPEC-019 first-validated; SPEC-015 substring-trap addendum
    applies to heading-level asserts; SPEC-002 monotonic-tiebreak
    rule applies to span/streak ordering tests); §11 Domain Glossary
    `stats` entry already exists from SPEC-018 (line 255); SPEC-020
    VERIFIES wording matches what ships and updates only if the entry
    misnames anything; §12 "decide at design time when decidable"
    — SPEC-020 applies it for entries_per_week's decimal-vs-integer
    weeks computation, top-5 tie-cap behavior, and the SPEC-019-earned
    Long-vs-help self-audit watch-pattern (the second confirming
    case).
  - `/projects/PROJ-001-mvp/brief.md` — STAGE-004 stage-plan + sketch.
  - `/projects/PROJ-001-mvp/stages/STAGE-004-rule-based-polish-summary-review-stats.md`
    — THE authoritative scope. Spec Backlog → SPEC-020 entry (lines
    ~225–232); Design Notes → "DEC-014" (lines ~250–267) consumed
    verbatim; "`internal/aggregate` package" (lines ~269–280) —
    extension is allowed and locked here; "Filter flag reuse" (lines
    ~282–288) — `brag stats` does NOT accept filter flags in MVP
    (locked NO); "Output destination" (lines ~290–295) — stdout only,
    no `--out` (locked NO); "SPEC-020-specific (`brag stats`)" (lines
    ~395–419).
  - `/projects/PROJ-001-mvp/session-log.md` — recent state.
  - `/projects/PROJ-001-mvp/backlog.md` — NOT for scope; for awareness
    of out-of-scope siblings (`brag remind`, emoji passes, `brag add
    --at` backdating, configurable stats metrics if a user ever asks,
    `--out` deferral, `--compact` JSON, calendar-week semantics).
  - `/decisions/DEC-014-rule-based-output-shape.md` — THE shape
    contract this spec consumes verbatim. Choices (1)–(6) all apply;
    `scope` field echoes `"lifetime"` for stats (per DEC-014 part 6
    + Option D rejection rationale: stats has no range, scope is
    always "lifetime").
  - `/decisions/DEC-013-markdown-export-shape.md` — provenance/summary
    block convention DEC-014 inherits from; stats markdown follows
    `Generated:` / `Scope:` / `Filters:` lines.
  - `/decisions/DEC-011-json-output-shape.md` — naked-array shape for
    list/export. DEC-014 envelope is intentionally different. Stats
    does NOT use the DEC-011 9-key per-entry shape because stats
    doesn't render entries — only aggregations.
  - `/decisions/DEC-006-cobra-cli-framework.md` — new `brag stats` is
    a cobra subcommand following the same pattern.
  - `/decisions/DEC-007-required-flag-validation-in-runE.md` — applies
    to `--format` validation; goes through `UserErrorf` in `RunE`,
    never `MarkFlagRequired`.
  - `/decisions/DEC-004-tags-comma-joined-for-mvp.md` — tags stored as
    comma-joined TEXT; `aggregate.MostCommon` for tags requires
    splitting on `,` and trimming at the aggregation boundary.
  - `/guidance/constraints.yaml` — full constraint list.
  - `/projects/PROJ-001-mvp/specs/done/SPEC-019-brag-review-week-and-month-flags.md`
    — DIRECT precedent. SPEC-019 created `internal/export/review.go`,
    extended `internal/aggregate` by one helper. SPEC-020 follows the
    same shape: `internal/export/stats.go` (sibling), extends
    `internal/aggregate` by THREE helpers (`Streak`, `MostCommon`,
    `Span`). Read its Notes for the Implementer for code-sketch
    patterns AND its Build Completion → Reflection (Ship) for the
    Long-vs-help negative-substring watch-pattern this spec applies
    proactively at design.
  - `/projects/PROJ-001-mvp/specs/done/SPEC-018-brag-summary-aggregate-package-and-shape-dec.md`
    — DEC-014 origin spec; structural precedent for the
    Rejected-alternatives-build-time discipline this spec applies a
    third time.
  - `/internal/aggregate/aggregate.go` — exists post-SPEC-018/019.
    Read to understand the `ByType` / `ByProject` /
    `GroupForHighlights` / `GroupEntriesByProject` signatures + the
    `NoProjectKey` constant. SPEC-020 ADDS three helpers (`Streak`,
    `MostCommon`, `Span`) plus three new types (`NameCount`,
    `CorpusSpan`, and reuses the existing `TypeCount` / `ProjectCount`
    pattern via `NameCount`).
  - `/internal/aggregate/aggregate_test.go` — sibling tests file;
    SPEC-020's new helper tests append after the existing six tests
    (NOT a separate file, since it's the same package and the
    existing file is the convention SPEC-018/019 established).
  - `/internal/export/summary.go` — sibling renderer; stats.go follows
    the same shape (`StatsOptions`-style options struct with
    injectable `Now`; markdown returns bytes with trailing `\n`
    stripped; JSON via `MarshalIndent` indent=2).
  - `/internal/export/review.go` — closest sibling renderer; stats.go
    follows the same shape (no `Filters` field on options struct
    because stats accepts no filter flags; the markdown line is
    hard-coded `Filters: (none)`; the JSON envelope is hard-coded
    `"filters": {}`).
  - `/internal/export/json.go` — `entryRecord` lives here; SPEC-020
    does NOT use it (stats doesn't render per-entry shapes).
    Untouched.
  - `/internal/export/markdown.go` — sibling; precedent for the
    `noProjectKey` constant + provenance line format. Stats uses
    `aggregate.NoProjectKey` (capital-N) for the data-layer sentinel.
  - `/internal/cli/summary.go` — direct precedent for the CLI command
    structure (`--format` validation in `RunE`, `UserErrorf` patterns).
    Stats's CLI structure is symmetric with three differences: (a) NO
    `--range` flag; (b) NO filter flags; (c) the `Now` passed to the
    renderer drives BOTH the `Generated:` line AND the streak
    today-reference (single source so determinism is identical).
  - `/internal/cli/review.go` — closest sibling CLI command (no filter
    flags, stdout only). Stats mirrors review's NO-filter-NO-out
    posture.
  - `/internal/cli/list_test.go` — `seedListEntry` helper at line
    ~258, same package, direct reuse for stats CLI tests.
  - `/internal/cli/errors.go` — `ErrUser` sentinel + `UserErrorf`
    helper.
  - `/internal/cli/summary_test.go` — direct precedent for CLI test
    harness (`newSummaryTestRoot`, `runSummaryCmd` helpers; line-based
    equality for heading-level asserts; regex strip for `Generated:`
    line). Stats's test file follows the same shape.
  - `/cmd/brag/main.go` — gains one
    `root.AddCommand(cli.NewStatsCmd())` line after the existing nine
    `AddCommand` calls (line 25 currently calls `NewReviewCmd`;
    SPEC-020 appends a tenth).
- **External APIs:** none. stdlib `encoding/json`, `time`, `sort`,
  `strings`, `bytes`, `fmt`, `math` cover the needs. No new Go module
  dependencies. Per `no-new-top-level-deps-without-decision`, any
  proposed dep needs its own DEC.
- **Related code paths:** `internal/aggregate/` (three helpers + three
  types added); `internal/export/` (new `stats.go` file); `internal/cli/`
  (new `stats.go` file); `cmd/brag/main.go`; `docs/`; `README.md`;
  `AGENTS.md`.

## Outputs

- **Files created:**
  - `/internal/export/stats.go` — new file in the existing
    `internal/export` package. Exports:
    - `type StatsOptions struct { Now time.Time }` — `Now` is injected
      for deterministic `Generated:` lines AND for the streak
      today-reference (single source: same value passes to
      `aggregate.Streak`'s `now` parameter). NO `Scope` field — stats
      always renders `Scope: lifetime` (hard-coded). NO `Filters` /
      `FiltersJSON` fields — stats accepts no filter flags; the
      markdown line is hard-coded `Filters: (none)` and the JSON
      envelope is hard-coded `"filters": {}`.
    - `func ToStatsMarkdown(entries []storage.Entry, opts StatsOptions) ([]byte, error)`
      — renders the markdown digest per DEC-014. Returns bytes with
      trailing `\n` stripped (matches `ToJSON` / `ToMarkdown` /
      `ToSummaryMarkdown` / `ToReviewMarkdown` byte contract). Empty
      entries slice returns header + provenance block only (no
      metric body), per DEC-014 part (4) "per-spec payload sections
      OMITTED on empty entries." Document structure: `# Bragfile
      Stats` heading + provenance block (`Generated:` / `Scope:
      lifetime` / `Filters: (none)`) + `## Stats` wrapper + five
      bold sub-headers (`**Activity**`, `**Streaks**`, `**Top tags**`,
      `**Top projects**`, `**Corpus span**`) with bullets under each.
    - `func ToStatsJSON(entries []storage.Entry, opts StatsOptions) ([]byte, error)`
      — renders the JSON envelope per DEC-014. Single object, top-
      level keys: `generated_at`, `scope` (always `"lifetime"`),
      `filters` (always `{}`), `total_count` (int), `entries_per_week`
      (float, rounded to 2 decimals), `current_streak` (int),
      `longest_streak` (int), `top_tags` (array-of-objects;
      `[{"tag", "count"}, ...]`), `top_projects` (array-of-objects;
      `[{"project", "count"}, ...]`), `corpus_span` (sub-object with
      `first_entry_date` (string `"YYYY-MM-DD"` or null when empty),
      `last_entry_date` (same), `days` (int)). Pretty-printed with
      2-space indent.
  - `/internal/export/stats_test.go` — new file. Six tests against a
    fixed `[]storage.Entry` fixture + explicit `StatsOptions`. Two
    load-bearing goldens (markdown + JSON). See Failing Tests.
  - `/internal/cli/stats.go` — new file. Exports
    `func NewStatsCmd() *cobra.Command` plus unexported `runStats`.
    Declares ONLY `--format` (default `markdown`; RunE-validated;
    accepted: `markdown`, `json`). NO filter flags (`--tag`,
    `--project`, `--type`), NO `--range`, NO `--out`, NO `--since`.
    Calls `Store.List(storage.ListFilter{})` (zero-value filter →
    every row, lifetime corpus), renders via `ToStatsMarkdown` /
    `ToStatsJSON` to stdout. Shares one `time.Now().UTC()` value
    between the renderer's `Generated:` line and the streak
    today-reference (single source — see Notes for the Implementer).
  - `/internal/cli/stats_test.go` — new file. Five tests using
    `t.TempDir()` for DB paths, seeding entries via the package-local
    `seedListEntry` helper. See Failing Tests.
- **Files modified:**
  - `/internal/aggregate/aggregate.go` — adds:
    - `type NameCount struct { Name string; Count int }` — generic
      top-N count; `Name` is the value (a tag string or a project
      string), `Count` is the occurrence count. Renamed-for-clarity
      sibling of `TypeCount` / `ProjectCount`; the renderer wraps
      `[]NameCount` into the per-spec `top_tags`/`top_projects`
      shapes with semantic JSON keys (`tag` / `project`).
    - `type CorpusSpan struct { First, Last time.Time; Days int }` —
      first entry's `CreatedAt` (UTC), last entry's `CreatedAt` (UTC),
      and the count of UTC calendar days from `First`'s date to
      `Last`'s date INCLUSIVE (single-day corpus → `Days == 1`; same
      day on multiple entries → `Days == 1`; empty corpus → all-zero
      struct). Inclusive-on-both-ends matches the natural reading of
      "the corpus exists for N days."
    - `func Streak(entries []storage.Entry, now time.Time) (current, longest int)`
      — computes both streaks in one pass over a UTC-date set. Test-
      injectable `now` for date-boundary determinism (mirrors
      `rangeCutoff(scope, now)` precedent from SPEC-018). `current`
      counts back from `now`'s UTC date: if that date has at least
      one entry, count it (1) and walk back until a date without
      entries; if `now`'s UTC date has zero entries, `current = 0`
      (NOT "the streak that ended yesterday — was N"; per stage
      Design Notes "Streak edge cases" + locked decision §6 below).
      `longest` is the longest consecutive UTC-date run anywhere in
      the corpus. Empty entries → returns `(0, 0)`.
    - `func MostCommon(values []string, n int) []NameCount` — pure
      top-N counter. Counts occurrences of each non-empty string in
      `values`; returns the top `n` by DESC count with alpha-ASC
      tiebreak. STRICT cap at `n` (when 6+ values tie at the boundary
      count, alpha-ASC tiebreak determines which `n` are returned —
      no overflow). Empty input → non-nil empty slice. Fewer than
      `n` distinct values → returns however many exist (no padding).
      The renderer extracts `[]string` from entries (split tags on
      `,` per DEC-004; project field with empty-string entries
      EXCLUDED) and calls this helper.
    - `func Span(entries []storage.Entry) CorpusSpan` — returns the
      `CorpusSpan` for `entries`. Empty input → zero-value struct
      (`First.IsZero()` true, `Last.IsZero()` true, `Days == 0`).
      `Days` computation: `int(lastDay.Sub(firstDay).Hours()/24) + 1`
      where `firstDay` and `lastDay` are the UTC-truncated calendar
      dates of `First` and `Last` (`time.Date(y, m, d, 0, 0, 0, 0,
      time.UTC)`).
  - `/internal/aggregate/aggregate_test.go` — adds four tests at the
    end of the file (same package convention; not a separate file):
    - `TestStreak_CurrentAndLongest` (subtests covering: today-with-
      entries → current ≥ 1; today-without-entries → current = 0;
      gap mid-corpus → longest preserved; single-entry → both = 1)
    - `TestMostCommon_TopNCapAlphaTiebreakAndEmpty` (subtests covering:
      strict cap at n; alpha-ASC tiebreak at the cap boundary; fewer
      than n distinct → returns all; empty input → non-nil empty
      slice)
    - `TestSpan_FirstLastAndInclusiveDays` (subtests covering: multi-
      entry span; single-entry span → days = 1; same-day-multiple-
      entries → days = 1; empty → zero-value struct)
    - `TestStatsAggregate_EmptyInputContract` (one consolidated
      assertion that `Streak(nil, t)`, `MostCommon(nil, 5)`,
      `Span(nil)` each return the empty contract — `(0, 0)`,
      non-nil empty slice, zero-value struct)
  - `/cmd/brag/main.go` — one added line:
    `root.AddCommand(cli.NewStatsCmd())` after the existing nine
    `AddCommand` calls (line 25 calls `NewReviewCmd`; SPEC-020 appends
    a line 26 calling `NewStatsCmd`).
  - `/docs/api-contract.md` — adds a new `### brag stats` section
    after the `### brag review --week | --month` section (line ~287
    today, post-SPEC-019). Section content (drafted in detail under
    Notes for the Implementer):
    - synopsis with `--format markdown|json` (default markdown);
    - prose describes the six metrics by name (total entries,
      entries/week rolling average, current/longest streak, top-5
      tags, top-5 projects, corpus span);
    - cross-link to DEC-014;
    - mention markdown is default; JSON is the single-object
      envelope with `top_tags` / `top_projects` as
      arrays-of-objects (rationale: DESC-by-count ordering would be
      lost under map encoding); `corpus_span` is a sub-object with
      date-or-null fields;
    - lifetime-corpus only; filter flags NOT accepted (rolling-window
      digests live in `brag summary`);
    - stdout only (no `--out`).
    The end-of-file References list (line ~362) is updated: the
    DEC-014 row currently says "and `brag stats` (arriving later in
    STAGE-004)" — UPDATE to drop the forward-reference (stats now
    shipped). New text: `DEC-014 — rule-based output shape for brag
    summary, brag review, and brag stats: single-object JSON envelope
    with generated_at / scope / filters provenance + per-spec payload
    keys; markdown convention reuses DEC-013's provenance +
    summary-block style.`
  - `/docs/tutorial.md` — (a) line 3 Scope blurb: `\`brag stats\`
    arrives in a later STAGE-004 spec.` → drop the deferral
    sentence entirely (stats now shipped). The Scope blurb shrinks
    by one line. (b) Optional but recommended: §4 "Read them back"
    gains a `### Lifetime stats: brag stats` subsection after the
    `### Weekly reflection: brag review` subsection (added by
    SPEC-019). Author judgment call on placement; keep prose short
    (5–10 lines). Mirrors SPEC-018/019's optional-but-recommended
    paste-into-AI subsections.
  - `/docs/data-model.md` — line 149 update: the existing DEC-014 row
    mentions "`brag stats` (later STAGE-004 spec)". UPDATE to drop
    the forward-reference: "`brag stats` (this spec)" or merge into
    the prose so all three commands appear as shipped consumers
    without forward-references. No schema change.
  - `/README.md` — line 64 update: the existing wording mentions
    "`brag stats` arrives in a later STAGE-004 spec." UPDATE to drop
    the forward-reference (stats now shipped). New wording rolls all
    three rule-based commands together as shipped: "`brag summary
    --range week|month`, `brag review --week|--month`, and `brag
    stats` are shipped (STAGE-004; markdown default, `--format json`
    for the DEC-014 envelope)."
  - `/AGENTS.md` — §11 Domain Glossary updates:
    - **`stats` entry (line 255)** — currently reads "**stats** —
      `brag stats`: six lifetime aggregations (total entries,
      entries/week rolling average, current streak, longest streak,
      top-5 most-common tags, top-5 most-common projects, corpus
      span). STAGE-004 (SPEC-020)." VERIFY this wording is accurate
      for what ships (per the user prompt "SPEC-020 should verify
      the `stats` entry's wording is accurate for the shipped
      command"). Audit result: WORDING ACCURATE — the six metrics
      named match the shipped set (the user prompt counts "six"
      framing metrics by treating corpus span as one composite;
      glossary line follows the same framing). NO change to the
      entry text needed; verified by inspection at design time.
    - **`aggregate` entry (line 246)** — currently mentions "and
      SPEC-020's `Streak` / `MostCommon` / `Span`". Helper names
      ALREADY MATCH this spec's locked names. NO change to the
      entry text needed; verified by inspection at design time.
- **New exports:**
  - `aggregate.NameCount`, `aggregate.CorpusSpan`, `aggregate.Streak`,
    `aggregate.MostCommon`, `aggregate.Span`.
  - `export.StatsOptions`, `export.ToStatsMarkdown`,
    `export.ToStatsJSON`.
  - `cli.NewStatsCmd`.
- **Database changes:** none. Pure read path; uses existing
  `Store.List(ListFilter{})` from SPEC-007 with a zero-value filter
  (lifetime corpus). No migration.

## Acceptance Criteria

Every criterion is testable. Paired failing test name in italics
where applicable. SPEC-020 has **13 failing tests** across **2 new
files** plus **4 tests appended to `internal/aggregate/aggregate_test.go`**.
Load-bearing goldens written FIRST per SPEC-014 / SPEC-015 / SPEC-018
/ SPEC-019 ship lessons.

- [ ] `aggregate.Streak(entries, now)` returns `(current, longest)`
      where `current` is the count of consecutive UTC days with ≥1
      entry counting back from `now`'s UTC date (or 0 if `now`'s UTC
      date has zero entries — locks "today-with-zero → 0, NOT 'the
      streak that ended yesterday'") and `longest` is the longest
      consecutive UTC-date run anywhere in the corpus. Single-entry
      corpus on `now`'s UTC date → both = 1. Empty corpus → both = 0.
      *TestStreak_CurrentAndLongest* (subtests `today_has_entries`,
      `today_zero_entries_yields_zero`, `gap_mid_corpus_longest`,
      `single_entry`, `empty_corpus`).
- [ ] `aggregate.MostCommon(values, n)` returns up to `n`
      `NameCount` entries ordered DESC by count with alpha-ASC
      tiebreak. STRICT cap at `n` even when 6+ values tie at the
      boundary (alpha-ASC tiebreak determines which `n`); FEWER than
      `n` distinct values returns however many exist (no padding).
      Empty-string values are EXCLUDED from counting (callers strip
      empty before calling). Empty input → non-nil empty slice.
      *TestMostCommon_TopNCapAlphaTiebreakAndEmpty* (subtests
      `cap_at_n`, `boundary_tie_alpha_resolves`, `fewer_than_n`,
      `empty_excluded`, `empty_input_nonnil_slice`).
- [ ] `aggregate.Span(entries)` returns a `CorpusSpan` with `First`
      = the earliest `CreatedAt` (UTC), `Last` = the latest
      `CreatedAt` (UTC), `Days` = the count of UTC calendar days
      from `First`'s date to `Last`'s date INCLUSIVE (single-day
      corpus → `Days == 1`; same-day multiple-entries → `Days == 1`;
      empty corpus → zero-value struct, `Days == 0`).
      *TestSpan_FirstLastAndInclusiveDays* (subtests `multi_day`,
      `single_day`, `same_day_multiple_entries`, `empty_corpus`).
- [ ] `aggregate.Streak(nil, t)`, `aggregate.MostCommon(nil, 5)`,
      `aggregate.Span(nil)` each return the empty contract — `(0,
      0)`, non-nil empty slice (`len == 0`, distinct from nil), and
      zero-value struct respectively. Mirrors SPEC-018's empty-input
      contract test for the existing helpers.
      *TestStatsAggregate_EmptyInputContract*
- [ ] `export.ToStatsMarkdown(fixture, opts{Now: <fixed>})` emits a
      byte-identical markdown document locking DEC-014 markdown
      choices on the fixture: `# Bragfile Stats` heading +
      provenance block (`Generated:` / `Scope: lifetime` / `Filters:
      (none)`) + `## Stats` wrapper + five bold sub-headers
      (`**Activity**`, `**Streaks**`, `**Top tags**`, `**Top
      projects**`, `**Corpus span**`) with bullets under each.
      *(LOAD-BEARING — write FIRST per SPEC-014/15/18/19 ship
      lessons.)*
      *TestToStatsMarkdown_DEC014FullDocumentGolden*
- [ ] `export.ToStatsJSON(fixture, opts{Now: <fixed>})` emits a
      byte-identical JSON envelope locking DEC-014 JSON choices on
      the same fixture: `generated_at` (RFC3339), `scope:
      "lifetime"`, `filters: {}`, `total_count` (int),
      `entries_per_week` (float, 2-decimal), `current_streak` (int),
      `longest_streak` (int), `top_tags` (array of `{tag, count}`
      objects in DESC-by-count order with alpha-ASC tiebreak),
      `top_projects` (same shape with `{project, count}`),
      `corpus_span` (sub-object with `first_entry_date`,
      `last_entry_date`, `days`). Pretty-printed indent=2.
      *(LOAD-BEARING — write SECOND.)*
      *TestToStatsJSON_DEC014ShapeGolden*
- [ ] `ToStatsMarkdown` AND `ToStatsJSON` on empty entries each emit
      the DEC-014-prescribed empty-corpus shape. Markdown: header +
      provenance only (no `## Stats` wrapper, no bold sub-headers,
      no metric bullets) — document ends after the `Filters: (none)`
      line. JSON: full envelope with `total_count: 0`,
      `entries_per_week: 0`, `current_streak: 0`, `longest_streak:
      0`, `top_tags: []`, `top_projects: []`, `corpus_span:
      {first_entry_date: null, last_entry_date: null, days: 0}`.
      *TestToStats_EmptyCorpusShape* (subtests `markdown` and
      `json`).
- [ ] Top-5 cap is enforced even when more than five tags/projects
      exist or when 6+ tie at the boundary count. Renderer emits
      exactly 5 entries in `top_tags` and 5 in `top_projects`; the
      five chosen are deterministic (DESC count + alpha-ASC at
      boundary). *TestToStats_TopFiveCapEnforcedAtBoundary*
      (subtests `tags` and `projects`; both use a 6-element fixture
      with 6 tied counts, alpha-ASC determines which 5 ship).
- [ ] `entries_per_week` calculation is decimal-weeks (NOT integer
      weeks): `weeks := float64(span.Days) / 7.0`; if `weeks < 1.0`
      (i.e., `span.Days < 7`) → `0.0`; else
      `math.Round((float64(total)/weeks)*100) / 100`.
      *TestEntriesPerWeek_DecimalWeeksAndSubWeekZero* (in
      `internal/export/stats_test.go`; subtests `sub_week_zero`,
      `exactly_one_week`, `partial_weeks_two_decimals`).
- [ ] `brag stats` (no flags) emits markdown to stdout (default
      format), starts with `# Bragfile Stats`, and contains a `Scope:
      lifetime` line. errBuf empty. *TestStatsCmd_BareDefaultsToMarkdown*
- [ ] `brag stats --format json` writes the JSON envelope per
      DEC-014 to stdout: top-level keys `generated_at`, `scope` =
      `"lifetime"`, `filters` = `{}`, plus the seven payload keys
      (`total_count`, `entries_per_week`, `current_streak`,
      `longest_streak`, `top_tags`, `top_projects`, `corpus_span`).
      Trailing newline from `fmt.Fprintln`; stderr empty.
      *TestStatsCmd_FormatJSONShape*
- [ ] `brag stats --format yaml` (unknown format) exits 1 (user
      error) with a message naming `yaml` AND the accepted set
      (`markdown`, `json`).
      *TestStatsCmd_UnknownFormatIsUserError*
- [ ] `brag stats --help` output contains needles `--format`,
      `markdown`, `json` (each as distinctive needles per AGENTS.md
      §9 assertion-specificity rule). The help text DOES NOT
      advertise `--tag`, `--project`, `--type`, `--out`, `--range`,
      `--since`, `--week`, `--month` (none of these flags are
      declared on this command). Stderr empty.
      *TestStatsCmd_HelpShowsFormatOnly*
- [ ] `brag stats --tag X` (and the corresponding cases for
      `--project`, `--type`, `--out`, `--range`, `--since`,
      `--week`, `--month`) exit with cobra's "unknown flag" error
      (NOT a `RunE` user-error path). Locks the "lifetime corpus,
      no filters, no range" decision per stage Design Notes — these
      flag names are genuinely undeclared, not declared-and-rejected.
      *TestStatsCmd_UndeclaredFlagsRejectedAsUnknown* (subtests for
      each undeclared flag name).
- [ ] `brag --help` lists `stats` as a subcommand (cobra auto-
      registers it via `cmd/brag/main.go` AddCommand).
      *[manual: `go build ./cmd/brag && ./brag --help` shows
      `stats` in the command list; `./brag stats --help` shows the
      synopsis.]*
- [ ] `gofmt -l .` empty; `go vet ./...` clean; `CGO_ENABLED=0
      go build ./...` succeeds; `go test ./...` and `just test`
      green. Existing SPEC-018 goldens
      (`TestToSummaryMarkdown_DEC014FullDocumentGolden`,
      `TestToSummaryJSON_DEC014ShapeGolden`) AND existing SPEC-019
      goldens (`TestToReviewMarkdown_DEC014FullDocumentGolden`,
      `TestToReviewJSON_DEC014ShapeGolden`) AND existing
      `TestToJSON_*` goldens stay byte-identical (proves the
      aggregate-package extension was non-breaking).
- [ ] Doc sweep: `docs/api-contract.md` gains a `### brag stats`
      section after the `### brag review` section; the DEC-014
      References row at line ~362 drops the "and `brag stats`
      (arriving later in STAGE-004)" forward-reference;
      `docs/tutorial.md` line 3 Scope blurb drops the stats
      deferral sentence; `docs/data-model.md` line 149 drops the
      stats forward-reference; `README.md` line 64 drops the stats
      deferral sentence; `AGENTS.md` §11 Domain Glossary `stats`
      and `aggregate` entries verified accurate (NO text changes
      required — both already name what ships).
      *[manual greps listed under Premise audit below.]*

## Locked design decisions

Reproduced here so build / verify don't re-litigate. Each is paired
with at least one failing test below per AGENTS.md §9 SPEC-009 ship
lesson + the SPEC-018 §12 "decide at design time when decidable"
discipline (third spec applying it proactively).

1. **DEC-014 consumed verbatim — no DEC emission, no re-litigation
   (third consumer).** All six locked choices apply: (1) JSON
   single-object envelope, (2) top-level flat keys (`generated_at`,
   `scope`, `filters` plus per-spec payload keys at top level — no
   nested `payload` wrapper), (3) markdown provenance-block convention
   (`# <Doc Title>` heading + `Generated:` / `Scope:` / `Filters:`
   lines), (4) empty-state values (numeric → 0; arrays → `[]`;
   objects → `{}`; date fields → null in JSON / `-` in markdown),
   (5) JSON pretty-printed indent=2, (6) `scope` echoes the literal
   range value — for stats the value is hard-coded `"lifetime"` per
   DEC-014 Option D rejection rationale ("stats has no range; the
   neutral term scope covers all three commands"). *Pair: load-
   bearing `TestToStatsMarkdown_DEC014FullDocumentGolden` +
   `TestToStatsJSON_DEC014ShapeGolden` cover (1)–(5);
   `TestStatsCmd_FormatJSONShape` covers (6) at the CLI plumbing
   level.*

2. **`top_tags` and `top_projects` are arrays-of-objects (NOT maps).**
   Each element is `{"tag": "<name>", "count": <n>}` for top_tags and
   `{"project": "<name>", "count": <n>}` for top_projects. Group order:
   DESC by count with alpha-ASC tiebreak; STRICT cap at 5 (per locked
   decision §3 below). The semantic-name keys (`tag` / `project`)
   match review's `entries_grouped[].project` semantic-name pattern
   from SPEC-019. ARRAY shape preserves DESC-by-count ordering — a
   map keyed by name would lose order under Go's `encoding/json`
   alpha-sort at marshal time. This is a deliberate per-spec payload
   shape divergence within DEC-014's envelope: SPEC-018's
   `counts_by_type` / `counts_by_project` use map shape (lookup-by-
   key is the consumer's purpose); SPEC-020's `top_tags` /
   `top_projects` use array-of-objects shape (DESC-by-count ordering
   IS the meaning). Two map payloads coexist with two array payloads
   inside the same DEC-014 envelope — documented intentionally.
   *Pair: `TestToStatsJSON_DEC014ShapeGolden` byte-locks the array-
   of-objects shape; `TestToStats_TopFiveCapEnforcedAtBoundary`
   exercises the cap + alpha-ASC at the boundary.*

3. **Top-5 cap is STRICT — exactly 5 (or fewer, when fewer distinct
   values exist).** When 6+ tags or projects tie at the boundary
   count, alpha-ASC tiebreak determines which 5 are returned. Fixed-
   shape contract beats include-all-ties unpredictability;
   downstream tooling (markdown column widths, JSON consumer array-
   handling assumptions) gets a deterministic upper bound. When fewer
   than 5 distinct values exist (small corpus, or after exclusion of
   empty-string projects), return however many exist — NO padding
   with empty placeholders. Empty corpus → empty array `[]`. *Pair:
   `TestMostCommon_TopNCapAlphaTiebreakAndEmpty` covers all four
   cases at the data-layer; `TestToStats_TopFiveCapEnforcedAtBoundary`
   covers the renderer side.*

4. **`entries_per_week` is decimal-weeks divided into total, rounded
   to 2 decimals; sub-1-week corpus → 0.0.** Lock the calculation:
   `weeks := float64(span.Days) / 7.0`; if `weeks < 1.0` (i.e.,
   `span.Days < 7`) → return `0.0`; else return
   `math.Round((float64(total)/weeks)*100) / 100`. Worked examples:
   - 14 entries spanning 14 days → weeks = 2.0 → 14/2.0 = 7.00.
   - 5 entries spanning 11 days → weeks ≈ 1.571 → 5/1.571 ≈ 3.18.
   - 3 entries on the same UTC day → span.Days = 1 → weeks ≈ 0.143
     < 1 → 0.0.
   - 7 entries spanning 7 days → weeks = 1.0 → 7/1.0 = 7.00.
   - 1 entry → span.Days = 1 → weeks ≈ 0.143 < 1 → 0.0.
   Decimal weeks (not integer weeks) preserves sub-week granularity;
   integer weeks would give 5/1 = 5.0 for the 11-day case, which
   is misleading. SPEC-018-precedent for "decide at design time
   when decidable" applies — building this in the build session
   would be off-loading. *Pair:
   `TestEntriesPerWeek_DecimalWeeksAndSubWeekZero` (subtests
   `sub_week_zero`, `exactly_one_week`,
   `partial_weeks_two_decimals`); the load-bearing JSON golden
   asserts the 2-decimal float output on the fixture.*

5. **`corpus_span` is a sub-object with date-or-null fields, not a
   flat triplet.** JSON shape: `"corpus_span": {"first_entry_date":
   "YYYY-MM-DD" or null, "last_entry_date": "YYYY-MM-DD" or null,
   "days": <int>}`. Date format is ISO 8601 calendar date
   (`YYYY-MM-DD`), UTC. Empty corpus → both date fields are `null`
   in JSON (`-` in markdown), `days` is `0`. Sub-object groups the
   three semantically-related fields under one key (consumer reads
   `jq .corpus_span.days` rather than `jq .span_days` floating at
   top level alongside seven other metrics). The asymmetry vs
   flat-top-level metrics (`total_count`, etc.) is intentional —
   `corpus_span` is a structured tuple; the others are scalars.
   *Pair: `TestToStatsJSON_DEC014ShapeGolden` byte-locks the sub-
   object shape; `TestToStats_EmptyCorpusShape` (json subtest)
   asserts `null` for both dates on empty input.*

6. **Streak edge cases — locked verbatim from stage Design Notes.**
   (a) Streak boundaries are UTC calendar days (matches storage's
   `time.Now().UTC()`). Time-zone handling deferred to backlog.
   (b) `now`'s UTC date with zero entries → `current_streak == 0`
   (NOT "the streak that ended yesterday — was N"). The contract
   is "consecutive UTC days WITH entries up to AND INCLUDING today";
   today-without-entries breaks the streak immediately. (c) `now`'s
   UTC date with ≥ 1 entry → count it (1) and walk back UTC day
   by UTC day until a day with zero entries appears; that day
   itself is NOT counted. (d) Multiple entries on the same UTC
   day count as one streak day (de-duplication by UTC date is the
   data layer's first pass before counting). (e) Test-injectable
   `now` (the second arg to `Streak(entries, now)`) so date-
   boundary cases are deterministic. *Pair: `TestStreak_CurrentAndLongest`
   exercises (b)–(e); (a) is implicit in the implementation
   (UTC-truncate via `e.CreatedAt.UTC()`).*

7. **`Span.Days` is INCLUSIVE on both endpoints.** Single-day corpus
   → `Days == 1` (the corpus exists for one day). Two-day corpus
   (Mon–Tue, regardless of how many entries on each) → `Days == 2`.
   Empty corpus → `Days == 0`. The inclusive-on-both-ends reading
   matches the natural "the corpus exists for N days" intuition;
   exclusive-on-end would give a 7-day Mon–Sun corpus a `Days` of 6,
   which surprises consumers. The same `span.Days` value drives
   `entries_per_week` per locked decision §4. *Pair:
   `TestSpan_FirstLastAndInclusiveDays` covers all three cases.*

8. **`--format markdown|json` default markdown, validated in `RunE`
   (DEC-007).** Unknown values → `UserErrorf` naming the offending
   value and the accepted list. NEVER `MarkFlagRequired`. Same set
   as summary and review; consistent help-surface vocabulary.
   *Pair: `TestStatsCmd_UnknownFormatIsUserError` +
   `TestStatsCmd_HelpShowsFormatOnly`.*

9. **NO filter flags, NO `--range`, NO `--out`, NO `--since`, NO
   `--week`/`--month`.** Stats accepts ONLY `--format`. Per stage
   Design Notes "Filter flag reuse" (stats is corpus-wide) +
   "Output destination" (stdout only) + "Six metrics, lifetime
   corpus" (no range). All eight unsupported flag names are
   GENUINELY UNDECLARED on the cobra command — cobra's auto-parser
   surfaces them as `unknown flag: --X` errors, which is the
   desired behavior (explicitly undeclared, not declared-and-
   rejected). *Pair: `TestStatsCmd_UndeclaredFlagsRejectedAsUnknown`
   (eight subtests, one per undeclared flag); `TestStatsCmd_HelpShowsFormatOnly`
   asserts `--help` advertises `--format` AND `markdown` AND `json`
   AND DOES NOT contain any of the eight undeclared flag tokens
   (line-based assertion per the SPEC-015 substring-trap addendum
   plus the SPEC-019 Long-vs-help self-audit watch-pattern locked
   here proactively — see Premise audit below).*

10. **Single `Now` source: the CLI layer takes `time.Now().UTC()`
    ONCE and passes the same value to BOTH the renderer's `Generated:`
    line AND `aggregate.Streak`'s `now` parameter (via
    `StatsOptions.Now`).** Mirrors SPEC-018's `MarkdownOptions.Now`
    /`SummaryOptions.Now` injectable-clock pattern. Avoids a subtle
    bug where two `time.Now()` calls bracket the streak computation
    and cross a midnight UTC boundary mid-render. The renderer
    function passes `opts.Now` straight through to
    `aggregate.Streak(entries, opts.Now)`. *Pair: load-bearing
    JSON/markdown goldens use a fixed `opts.Now`; the CLI tests
    (which CAN'T fix `time.Now()`) assert on the SHAPE of the
    `Generated:` line via the RFC3339 regex, not on a specific
    timestamp.*

11. **`internal/aggregate` extension: THREE new helpers (`Streak`,
    `MostCommon`, `Span`) plus two new types (`NameCount`,
    `CorpusSpan`).** `MostCommon` is generic over `[]string` (renderer
    extracts tags/projects from entries before calling), keeping the
    helper independent of `storage.Entry`'s shape. `Streak` and
    `Span` operate on `[]storage.Entry` directly because they read
    the timestamp field. The single-helper-per-spec growth pattern
    SPEC-019 established becomes a three-helper-per-spec growth here,
    proportional to SPEC-020's three lifetime aggregations that don't
    map to existing helpers. NO refactor of `ByType` / `ByProject` /
    `GroupForHighlights` / `GroupEntriesByProject` — those stay
    untouched. *Pair: `TestStreak_*`, `TestMostCommon_*`, `TestSpan_*`
    each cover one new helper; `TestStatsAggregate_EmptyInputContract`
    consolidates the empty-input contract.*

12. **NO LLM piping; NO `brag ai-stats`; NO summary-vs-stats
    cross-command logic.** Stats stands alone — it doesn't consume
    summary's output, doesn't compose with review's output, doesn't
    feed any LLM. The user's downstream AI workflow is the same
    paste-into-Claude/GPT shape across all three rule-based
    commands per DEC-014's value thesis. *No paired test — this is
    a scope lock by absence (the spec contains zero LLM-related
    code paths).*

**Out of scope (by design — backlog entries exist or are explicitly
deferred):**

- `--out <path>` flag on stats. Backlog deferral noted in stage
  Design Notes "Output destination". Same backlog entry covers
  summary, review, and stats.
- `--range <window>` on stats. Lifetime corpus only for MVP per
  stage Design Notes "Six metrics, lifetime corpus"; backlog if a
  user ever asks for windowed stats.
- Filter flags (`--tag`/`--project`/`--type`) on stats. Stage
  Design Notes "Filter flag reuse" locks NO; backlog if real
  workflow emerges (likely never — the value of stats is the
  unfiltered lifetime view).
- `--since` / arbitrary date filters on stats. Same scope lock.
- `--week`/`--month` on stats (review's flag set). Review owns
  windowed reflection; stats owns lifetime panorama. Distinct
  commands, distinct flag sets.
- Configurable metric list (a way to opt out of, e.g., streaks).
  Six are baked; deterministic spec scope. Backlog if a user
  ever asks for customization.
- A seventh+ metric (median-entries-per-week, busiest-day-of-week,
  type-distribution-percentages, etc.). Six is the locked count;
  expanding would be a backlog item with revisit trigger "user
  requests after using stats for several weeks."
- `--compact` / non-pretty JSON for stats (or any of the three
  STAGE-004 commands). Inherits DEC-011's pretty-default; same
  backlog entry covers them all.
- Time-zone configuration for streak/span boundaries. UTC-only
  for MVP (matches storage's `time.Now().UTC()`); revisit if a
  user notices a streak break across timezone changes.
- Calendar-week semantics for `entries_per_week` (e.g., compute
  weeks as Mon–Sun chunks rather than rolling 7-day blocks).
  Decimal-weeks-from-span is the locked formula; calendar-week
  alternative is in the backlog under DEC-014's "calendar-week
  semantics for `--range`" entry (same revisit trigger applies
  here).
- `brag remind` / habit-enforcement. Backlog with revisit trigger
  "first week with zero entries."
- LLM piping / AI integration — PROJ-002 territory.
- Any change to summary's or review's existing behavior. SPEC-020
  adds to `internal/aggregate` but does not modify any existing
  helper's signature or return shape. Existing goldens stay
  byte-identical.

**Rejected alternatives (build-time):**

These are choices the build agent might consider, with the
prescribed path locked here so the call doesn't off-load to
build-time and slip into Deviations later. Per SPEC-018 ship
reflection: "either-is-fine off-loads to build" is the
anti-pattern these locks deliberately avoid. SPEC-018 was the
first proactive application; SPEC-019 the second; SPEC-020 the
third.

1. **Map-keyed `top_tags` / `top_projects` JSON shape — REJECTED.**
   The path would render `top_tags` as a JSON object keyed by tag
   name (`{"auth": 12, "security": 8, ...}`) instead of an array
   of `{tag, count}` objects.

   *Why rejected:*
   - **Loses DESC-by-count ordering.** Go's `encoding/json` sorts
     `map[string]int` keys alphabetical-ASC when marshaling
     (deterministic since 1.12). The "top" semantic of `top_tags`
     IS DESC-by-count — alpha-sort destroys it. Consumers reading
     `jq '.top_tags | to_entries | sort_by(.value) | reverse'` to
     reconstruct the order would have to repeat the sort the data
     layer already did. The array-of-objects shape preserves the
     locked ordering by-construction.
   - **Asymmetric with locked decision §2.** `top_tags` semantically
     differs from `counts_by_type` (SPEC-018's map-keyed shape) — a
     count-by-type is "lookup the count for type X" (map-natural);
     a top-N is "give me the items in DESC count order" (array-
     natural). The same envelope can carry both shapes without
     drift; the per-spec choice is locked.
   - **JSON consumers writing AI prompts.** A prompt asking
     "summarize my top tags" reads naturally over an array of
     `{tag, count}` objects in order. Over a map the prompt has to
     instruct the LLM to "sort the values DESC and treat as a
     ranking" — extra cognitive load for zero gain.

2. **Integer-weeks `entries_per_week` formula — REJECTED.** The path
   would compute `weeks := span.Days / 7` (integer division) and
   `entries_per_week := total / weeks`. For an 11-day corpus with
   5 entries: 11/7 = 1, 5/1 = 5.0, vs the locked decimal-weeks
   formula's 5/1.571 = 3.18.

   *Why rejected:*
   - **Loses sub-week granularity in the most common range.** A
     corpus that's 1.5 weeks old should NOT report the same
     entries/week as a 1-week corpus. Integer weeks rounds the
     denominator down, inflating the average.
   - **Surprises low-volume users.** A user with 3 entries spanning
     12 days would see "3.0 entries/week" under integer-weeks
     (12/7 = 1, 3/1 = 3.0) when the natural reading is "I logged
     about 1.75 entries/week" (3 / (12/7) = 3 / 1.71 = 1.75).
     Decimal-weeks gives the natural answer.
   - **Per stage Design Notes lock.** "Decide at design time" — the
     formula choice is determinable from the metric semantics
     ("rolling average") and shouldn't off-load to build.

3. **Include-all-ties at the top-5 boundary — REJECTED.** The path
   would return all elements at the boundary count when 6+ tie at
   the cap (so `top_tags` could legitimately have length 6, 7, or
   more on edge fixtures).

   *Why rejected:*
   - **Variable output size surprises consumers.** The contract name
     is "top 5"; consumers reading the spec or the help text expect
     5. Variable-length arrays under a fixed name surface invisible
     edge cases (markdown column widths, JSON-to-CSV exporters,
     downstream tooling that pre-allocates buffers).
   - **The "fairness" reading is weak.** Alpha-ASC tiebreak is
     already deterministic and well-defined; extending it from
     "which N when N tie" to "all who tie" is a different contract.
     The strict-cap contract is documented and consumers can lean
     on it; include-all-ties is not.
   - **Future "top 10" / "top 20" knob.** A backlog entry can
     promote configurable N if a user asks; the strict-cap formula
     extends naturally (just change the constant). Include-all-ties
     would couple any future knob to the size of the underlying
     ties, which is opaque to the user.

4. **`MostCommon` operating on `[]storage.Entry` directly — REJECTED.**
   The path would expose two helper functions
   (`MostCommonTags(entries, n) []TagCount`, `MostCommonProjects(entries,
   n) []ProjectCount`) that handle DEC-004 splitting + (no project)
   exclusion internally, and skip the generic `MostCommon` helper.

   *Why rejected:*
   - **Two functions duplicate the top-N counting logic.** The
     "count, sort DESC by count with alpha-ASC tiebreak, cap at n"
     algorithm runs in both — twice the surface for the same
     algorithm.
   - **`[]string` is the natural input.** Tags and projects both
     reduce to "a list of strings"; the splitting/exclusion logic
     is renderer-side adapter code, not aggregation logic. Keeping
     `MostCommon` generic over `[]string` makes future use cases
     (e.g., a hypothetical "top 5 most-common verbs in titles" if
     a future spec wants it) trivial — pass any string slice.
   - **`NameCount` is the right name for the data type.** A neutral
     `NameCount{Name, Count}` works cleanly in the renderer (which
     re-keys to `{tag, count}` / `{project, count}` via struct tags
     on the JSON-side wrapping types). Specific types
     (`TagCount`, `ProjectCount`) inside `internal/aggregate` would
     pollute the package surface and force the renderer to switch
     on which helper produced the result.

5. **Inline `Streak` / `MostCommon` / `Span` in
   `internal/export/stats.go` (no `internal/aggregate` extension) —
   REJECTED.** The path would skip the aggregate-package growth and
   compute all six metrics inside the renderer.

   *Why rejected:*
   - **Violates the SPEC-018-locked aggregate/render seam.**
     `internal/aggregate` is the data layer; `internal/export` is
     the bytes layer. SPEC-018's locked decision §2 ("aggregate is
     the SINGLE data-layer source") is binding for SPEC-019/020 by
     parent-stage scope.
   - **Loses unit-testability of the algorithm.** `Streak` and
     `Span` have non-trivial date-boundary edge cases (UTC midnight,
     same-day multiple entries, today-with-zero, single-day corpus).
     Testing each through the renderer + cobra layer would couple
     the algorithm tests to the rendering byte contract, multiplying
     test brittleness for zero gain.
   - **Future PROJ-002 LLM-backed siblings.** A `brag ai-stats`
     spec in PROJ-002 would want to wrap the same six aggregations
     and feed them into an LLM prompt. The aggregate-package shape
     is the reusable surface; the renderer's bytes are not.

6. **Computing `Span.Days` exclusive (subtract-and-don't-add-one) —
   REJECTED.** The path would compute
   `days := int(lastDay.Sub(firstDay).Hours() / 24)` and treat a
   single-day corpus as `Days == 0`.

   *Why rejected:*
   - **Surprises the natural reading.** "How many days does my
     corpus span?" — a one-day corpus spans ONE day, not zero.
     Exclusive math gives 0 for single-day corpora, which makes
     `entries_per_week` divide-by-zero (or skip with sub-week
     check, which then triggers when it shouldn't have).
   - **Empty corpus already has its own zero contract.** `len(entries)
     == 0` → zero-value `CorpusSpan` (all three fields zero,
     including `Days`). Inclusive math for non-empty corpora gives
     `Days >= 1`, which composes cleanly with `entries_per_week`'s
     sub-week check.
   - **Forward-compatible with calendar-week-aware variants.**
     Future calendar-week semantics (if ever promoted from backlog)
     would re-derive the day count from a different basis; the
     inclusive-on-both-ends contract is the right invariant to lock
     for the rolling-window MVP.

## Premise audit (AGENTS.md §9 — addition + status-change, with
audit-grep cross-check applied at design)

SPEC-020 is an **addition** case (new command, three new aggregate
helpers, new export functions, new exports list, +1 to root
AddCommand call count) AND a **status-change** case (api-contract.md
+ tutorial.md + README.md + data-model.md all forward-reference
`brag stats` as deferred behavior; SPEC-020 supersedes those
references). Both AGENTS.md §9 heuristics apply.

This spec is the **second proactive validation case for the
SPEC-018-earned audit-grep cross-check addendum** ("design enumerates
→ design verifies its enumeration → build re-verifies and questions
deltas"; SPEC-019 was the first). It is also the **second proactive
validation case for the SPEC-019-earned Long-vs-help negative-
substring self-audit watch-pattern** ("for each `NOT contains`
assertion in Failing Tests, GREP THE SPEC TEXT for the forbidden
token; if found, decide whether to fix the prose or the test BEFORE
locking"; SPEC-019 had to self-catch this at build, SPEC-020 catches
it at design). If both proactive applications hold clean through
build + verify, ship reflection should propose §12 codification of
both addenda together.

Each grep below was actually executed at design time against the
current working tree; the enumerated hits below match the actual
`rg` / `grep` output as of 2026-04-25.

**Addition heuristics** (SPEC-011 ship lesson — grep tracked
collections for count coupling):

- Root command list: `cmd/brag/main.go` has nine `AddCommand` calls
  today (verified 2026-04-25, post-SPEC-019 ship — lines 17–25 in
  `main.go`). SPEC-020 makes it ten. Grep:

  ```
  grep -n 'AddCommand' cmd/brag/main.go internal/cli/*.go
  ```

  Audit each hit:
  - `cmd/brag/main.go:17–25`: the nine existing calls. Adding a
    tenth doesn't break any test.
  - `internal/cli/*_test.go`: tests use ad-hoc roots that
    `AddCommand` only the subcommand under test (e.g.,
    `newSummaryTestRoot` calls `root.AddCommand(NewSummaryCmd())`);
    they do NOT iterate the global subcommand list, so they're
    unaffected by the new entry.
  - No test asserts `len(root.Commands()) == 9` or similar; verified
    by `grep -rn 'NumCommand\|len.*Commands()' internal/cli/` returning
    zero hits.

- DEC collection: SPEC-020 adds NO new DEC. Verified 2026-04-25:
  the grep `find decisions -name 'DEC-*' | wc -l` returns 14 today
  (DEC-001 through DEC-014); SPEC-020 leaves it at 14.

- `internal/aggregate` exports: SPEC-020 adds five (the three
  function exports `Streak`, `MostCommon`, `Span` plus the two
  type exports `NameCount`, `CorpusSpan`). No test asserts on the
  package's exported-symbol count.

- `internal/export` exports: SPEC-020 adds three (`StatsOptions`,
  `ToStatsMarkdown`, `ToStatsJSON`). No test asserts on the package's
  exported-symbol count.

- `--format` accepted values: distinct flag per command. The
  list/export/summary/review tests asserting on their respective
  `--format` accepted sets are unaffected by stats's separate
  `(accepted: markdown, json)` list (same accepted set as summary
  and review, but the tests are command-scoped — verified by reading
  `internal/cli/{summary,review}_test.go`'s help-test functions
  which only assert on their command's help output).

- `internal/cli/list_test.go` `seedListEntry` helper: SPEC-020
  reuses it (same package, source-level reuse). Helper signature
  stable.

- `entryRecord` shape (SPEC-019's extracted `toEntryRecord`):
  SPEC-020 does NOT use it (stats doesn't render per-entry shapes).
  Existing `internal/export/json_test.go` and
  `internal/export/review_test.go` goldens stay byte-identical;
  no risk.

**Status-change heuristics** (SPEC-012 ship lesson — grep feature
name across docs) — explicit grep commands AND actual hits as of
2026-04-25:

```
grep -rn "brag stats" docs/ README.md AGENTS.md
```

Actual hits (verified 2026-04-25 via the grep run earlier this
session):

- `README.md:64` — "`brag stats` arrives in a later STAGE-004
  spec." → UPDATE: stats now shipped; drop the deferral sentence
  and re-roll the three rule-based commands together.
- `docs/api-contract.md:362` — DEC-014 References row mentions
  "and `brag stats` (arriving later in STAGE-004)". UPDATE: drop
  the forward-reference; rewrite as "`DEC-014` — rule-based output
  shape for `brag summary`, `brag review`, and `brag stats`: ..."
- `docs/tutorial.md:3` — "`brag stats` arrives in a later STAGE-004
  spec." → UPDATE: drop the deferral sentence (the Scope blurb
  shrinks).
- `docs/data-model.md:149` — DEC-014 row mentions "`brag stats`
  (later STAGE-004 spec)". UPDATE: drop the forward-reference.
- `AGENTS.md:246` — `aggregate` glossary entry already names the
  three SPEC-020 helpers (`Streak` / `MostCommon` / `Span`). NO
  change needed; verified at design.
- `AGENTS.md:249` — `digest` glossary entry already names all three
  rule-based commands. NO change needed.
- `AGENTS.md:255` — `stats` glossary entry already names the six
  metrics with wording that matches what ships. NO change needed;
  verified at design per the user prompt's instruction "SPEC-020
  should verify the `stats` entry's wording is accurate for the
  shipped command."

**Long-vs-help negative-substring self-audit (SPEC-019-earned
watch-pattern, second proactive validation case):**

For every `NOT contains "X"` assertion in this spec's Failing Tests,
the grep below finds occurrences of `X` in this spec's
load-bearing text (the `Long` description sketch under Notes for
the Implementer + the `## Outputs` prose that drives what the
implementer types into the binary). If a forbidden token appears
in load-bearing text, fix the prose OR the test BEFORE locking.
Notes for the Implementer that contain `--`-prefixed flag names
in DESIGN PROSE (not the Long sketch itself) are fine — those
tokens never reach the cobra `--help` rendering.

Forbidden tokens for `TestStatsCmd_HelpShowsFormatOnly` and
`TestStatsCmd_UndeclaredFlagsRejectedAsUnknown`: `--tag`,
`--project`, `--type`, `--out`, `--range`, `--since`, `--week`,
`--month`. Each must NOT appear in the locked `Long` sketch (under
Notes for the Implementer → `internal/cli/stats.go` skeleton →
`Long` field).

Grep at design time (the `Long` block extracted to a temp file or
inspected line-by-line):

```
grep -nE -- '--tag|--project|--type|--out|--range|--since|--week|--month' \
  <Long-sketch text only>
```

**Audit result:** The locked `Long` sketch (drafted under Notes for
the Implementer below) names rejected flags WITHOUT the `--` prefix
where mentioned in prose ("Filter flags (tag, project, type) and
the range / since / out flags that summary and review use are NOT
accepted on stats — use `brag summary` for windowed digests."). The
grep against the `Long` block returns zero hits for any forbidden
token. ✅ Audit clean at design time. The SPEC-019 build-time
deviation (where `--`-prefixed names in the Long sketch tripped
the help-test) does NOT recur here.

Spec-prose hits (the `## Locked design decisions` section and this
Premise audit section both name `--tag` / `--project` / etc. with
the `--` prefix as design references) — these are NOT load-bearing
because they don't get rendered into the binary's `--help` output.
The watch-pattern targets the spec→binary→test path, not the
spec-prose→reader path. The greps are scoped accordingly.

**Doc-sweep symmetric action enumeration** (from the `brag stats`
greps above — every hit maps to a planned `## Outputs` modification):

| Grep hit | Action under `## Outputs` |
| --- | --- |
| `README.md:64` | Drop deferral sentence; re-roll three commands as shipped |
| `docs/api-contract.md:362` (DEC-014 row) | Drop forward-reference; rewrite row |
| `docs/api-contract.md` (no `### brag stats` section yet) | ADD new section after `### brag review` |
| `docs/tutorial.md:3` | Drop deferral sentence from Scope blurb |
| `docs/tutorial.md:§4` (no stats subsection yet) | OPTIONAL: add `### Lifetime stats: brag stats` subsection after the SPEC-019 review subsection |
| `docs/data-model.md:149` | Drop forward-reference |
| `AGENTS.md:246` (aggregate entry) | NO change (already accurate) |
| `AGENTS.md:249` (digest entry) | NO change (already accurate) |
| `AGENTS.md:255` (stats entry) | NO change (verified accurate) |

No discoveries expected at build time. Build re-runs all the
greps above and questions any delta from this enumerated list
(per the audit-grep cross-check both-sides discipline; SPEC-019
ran clean both times — SPEC-020 is the second confirming case
for the addendum).

**Existing-test audit** (addition-case doesn't add tracked-count
coupling beyond the AddCommand list verified above; verify nothing
breaks):

- `internal/cli/summary_test.go`, `internal/cli/review_test.go` —
  command-scoped help tests; unaffected by stats's separate flag
  set. The `--format markdown|json` accepted-set assertion in
  summary's and review's help tests is summary-/review-scoped;
  stats's help test asserts the same accepted set on stats's own
  help output.
- `internal/cli/list_test.go`, `export_test.go`, `show_test.go`,
  `add_test.go`, `add_json_test.go`, `edit_test.go`, `delete_test.go`,
  `search_test.go` — no overlap with stats surface.
- `internal/export/summary_test.go`, `review_test.go`,
  `markdown_test.go`, `json_test.go` — sibling files in the package;
  new `stats_test.go` lands alongside, no cross-coupling. Existing
  goldens (`TestToSummaryMarkdown_*`, `TestToSummaryJSON_*`,
  `TestToReviewMarkdown_*`, `TestToReviewJSON_*`, `TestToJSON_*`)
  stay byte-identical.
- `internal/aggregate/aggregate_test.go` — gains four tests at the
  end of the file; existing six tests
  (`TestByType_DESCByCountAlphaTiebreak`,
  `TestByProject_NoProjectKeyForcedLast`,
  `TestGroupForHighlights_ChronoASCWithNoProjectLast`,
  `TestAggregate_EmptyInputReturnsNonNilEmptySlice`,
  `TestGroupEntriesByProject_OrderingAndIDTiebreak`,
  `TestGroupEntriesByProject_EmptyInputReturnsNonNilEmptySlice`)
  stay byte-identical.
- `internal/storage/*_test.go` — read-only path; stats uses existing
  `Store.List` with a zero-value filter. Unaffected.

## Failing Tests

Written now, during **design**. Thirteen tests total across **2 new
files** plus **4 tests appended to `internal/aggregate/aggregate_test.go`**.
All follow AGENTS.md §9: separate `outBuf` / `errBuf` with no-cross-
leakage asserts; fail-first run before implementation; assertion
specificity on help/error substrings; every locked decision paired
with at least one failing test; line-based equality (not
`strings.Contains`) for any heading-level assertion (SPEC-015
substring-trap addendum); ID-based (not timestamp-based) distinctness
for any freshness or ordering tie-break (SPEC-017 freshness-assertion
addendum); test-injectable `now` for streak determinism (SPEC-018
`rangeCutoff` precedent).

Goldens reuse a single fixture so all renderer choices anchor to the
same canonical entries. Aggregate tests use literal `[]storage.Entry`
slices.

### Shared renderer fixture (used by tests 5, 6, 8)

```go
// 8 entries spanning 14 UTC days, exercising:
//   * top-5 cap at 5 (six tags total: auth, security, backend, db,
//     perf, refactor — 6th sorted alpha-ASC drops out)
//   * (no project) excluded from top_projects
//   * span: first 2026-04-12, last 2026-04-25, days 14
//   * streak: at fixedNow=2026-04-25T12:00, current ≥ 1 because
//     2026-04-25 has an entry; longest = 3 because 2026-04-12 to
//     2026-04-14 are consecutive
//   * entries_per_week: total=8, span_days=14, weeks=2.0, value=4.00
var fixture = []storage.Entry{
    {ID: 1, Title: "alpha-old", Tags: "auth,security",
     Project: "alpha", Type: "shipped",
     CreatedAt: time.Date(2026, 4, 12, 10, 0, 0, 0, time.UTC),
     UpdatedAt: time.Date(2026, 4, 12, 10, 0, 0, 0, time.UTC)},
    {ID: 2, Title: "alpha-mid", Tags: "auth,backend",
     Project: "alpha", Type: "shipped",
     CreatedAt: time.Date(2026, 4, 13, 10, 0, 0, 0, time.UTC),
     UpdatedAt: time.Date(2026, 4, 13, 10, 0, 0, 0, time.UTC)},
    {ID: 3, Title: "alpha-new", Tags: "auth",
     Project: "alpha", Type: "learned",
     CreatedAt: time.Date(2026, 4, 14, 10, 0, 0, 0, time.UTC),
     UpdatedAt: time.Date(2026, 4, 14, 10, 0, 0, 0, time.UTC)},
    {ID: 4, Title: "beta-mid", Tags: "db",
     Project: "beta", Type: "shipped",
     CreatedAt: time.Date(2026, 4, 18, 10, 0, 0, 0, time.UTC),
     UpdatedAt: time.Date(2026, 4, 18, 10, 0, 0, 0, time.UTC)},
    {ID: 5, Title: "unbound-1", Tags: "perf",
     Project: "", Type: "shipped",  // (no project)
     CreatedAt: time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC),
     UpdatedAt: time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC)},
    {ID: 6, Title: "gamma-1", Tags: "refactor",
     Project: "gamma", Type: "fixed",
     CreatedAt: time.Date(2026, 4, 22, 10, 0, 0, 0, time.UTC),
     UpdatedAt: time.Date(2026, 4, 22, 10, 0, 0, 0, time.UTC)},
    {ID: 7, Title: "beta-late", Tags: "auth,db",
     Project: "beta", Type: "shipped",
     CreatedAt: time.Date(2026, 4, 24, 10, 0, 0, 0, time.UTC),
     UpdatedAt: time.Date(2026, 4, 24, 10, 0, 0, 0, time.UTC)},
    {ID: 8, Title: "alpha-last", Tags: "security",
     Project: "alpha", Type: "shipped",
     CreatedAt: time.Date(2026, 4, 25, 10, 0, 0, 0, time.UTC),
     UpdatedAt: time.Date(2026, 4, 25, 10, 0, 0, 0, time.UTC)},
}

var fixedNow = time.Date(2026, 4, 25, 12, 0, 0, 0, time.UTC)
```

Tag occurrence counts (from the fixture, after DEC-004 splitting):
- `auth`: 4 (entries 1, 2, 3, 7)
- `security`: 2 (entries 1, 8)
- `db`: 2 (entries 4, 7)
- `backend`: 1 (entry 2)
- `perf`: 1 (entry 5)
- `refactor`: 1 (entry 6)

Top-5 expected (DESC count + alpha-ASC tiebreak): `auth` (4),
`db` (2), `security` (2), `backend` (1), `perf` (1). The 6th
candidate `refactor` (1) drops out because 4 tags tie at count=1
and only 5 slots exist; alpha-ASC at the boundary keeps `backend`
+ `perf` and excludes `refactor`. That's the cap-with-alpha-tiebreak
case captured in the load-bearing JSON golden.

Project counts (after empty-string exclusion):
- `alpha`: 4 (entries 1, 2, 3, 8)
- `beta`: 2 (entries 4, 7)
- `gamma`: 1 (entry 6)

Top-5 expected: `alpha` (4), `beta` (2), `gamma` (1). Three
distinct projects → three entries returned (no padding to 5).
`(no project)` is EXCLUDED before counting; entry 5 contributes
nothing to top_projects.

Span: First = 2026-04-12T10:00:00Z, Last = 2026-04-25T10:00:00Z,
Days = 14 (inclusive of both calendar dates).

Streak at fixedNow = 2026-04-25T12:00:00Z UTC: current = 1 (only
2026-04-25 has an entry; 2026-04-24 is also in the corpus, so the
walk-back continues — entry 7 dates to 2026-04-24 → current = 2;
walk back to 2026-04-23 → no entry → streak ends. So current = 2.).
Longest = 3 (2026-04-12 → 2026-04-13 → 2026-04-14 are three
consecutive days with entries; no other run of 3+).

Entries/week: total = 8, span_days = 14, weeks = 14/7.0 = 2.0,
entries_per_week = 8 / 2.0 = 4.00.

### `internal/aggregate/aggregate_test.go` (existing file — 4 tests appended)

The new helper tests append after the existing six tests in the
same file, sharing imports and any package-local helpers (the
existing `sharedFixture` in `aggregate_test.go` is reused only
where the SPEC-018 fixture happens to fit; SPEC-020's tests
introduce separate compact fixtures for each case to keep
date-boundary scenarios crisp).

#### Test 1 — `TestStreak_CurrentAndLongest`

Five subtests covering the streak edge cases locked in §6. Build
session: see SPEC-019 `aggregate_test.go` test-construction patterns
for how to lay these out.

- **`today_has_entries`** — three consecutive UTC dates ending on
  `now`'s date (e.g., Apr 23/24/25 with `now` = Apr 25 12:00 UTC).
  Want `current == 3`, `longest == 3`.
- **`today_zero_entries_yields_zero`** — three consecutive UTC dates
  ending on `now - 1` day (e.g., Apr 22/23/24 with `now` = Apr 25).
  Want `current == 0`, `longest == 3`. Pairs locked decision §6
  part (b) — today-with-zero → 0, NOT "the streak that ended
  yesterday was N".
- **`gap_mid_corpus_longest`** — a run of 5 (e.g., Apr 10–14), an
  8-day gap, then a run of 2 (Apr 23–24); `now` = Apr 25 with no
  entry on it. Want `current == 0`, `longest == 5`. Pairs §6 (a)
  + (b) + longest-walks-the-whole-corpus.
- **`single_entry`** — one entry on `now`'s UTC date. Want
  `current == 1`, `longest == 1` (natural answer for a user's
  first day with `brag`).
- **`multiple_entries_same_day`** — three entries on the same UTC
  date (different hours). Want `current == 1`, `longest == 1`.
  Pairs locked decision §6 part (d) — same-day de-duplication.

#### Test 2 — `TestMostCommon_TopNCapAlphaTiebreakAndEmpty`

Five subtests covering the top-N contract. Use `reflect.DeepEqual`
on `[]NameCount` for each case (mirrors SPEC-018 ByType test
structure).

- **`cap_at_n`** — input `["a","a","a","b","b","c","c","d","e"]`
  with `n=3`. Want `[{a,3}, {b,2}, {c,2}]` — strict cap at 3,
  alpha-ASC between b and c, d/e excluded.
- **`boundary_tie_alpha_resolves`** — input
  `["zebra","yak","x-ray","wolf","vulture","umbrella"]` (each
  count=1) with `n=5`. Want the 5 alphabetically-first values
  (`umbrella, vulture, wolf, x-ray, yak`); `zebra` excluded.
  Pairs locked decision §3.
- **`fewer_than_n`** — input `["a","a","b"]` with `n=5`. Want
  `[{a,2}, {b,1}]` — 2 elements returned, NO padding.
- **`empty_strings_excluded`** — input `["a","","a","","b"]` with
  `n=5`. Want `[{a,2}, {b,1}]` — empty-string values excluded
  from counting.
- **`empty_input_nonnil_slice`** — `MostCommon(nil, 5)` and
  `MostCommon([]string{}, 5)` each return non-nil empty slices
  (`got != nil && len(got) == 0`). Pairs DEC-014 part (4).

#### Test 3 — `TestSpan_FirstLastAndInclusiveDays`

Four subtests covering the span contract.

- **`multi_day`** — three entries on Apr 12, 18, 25 (UTC). Want
  `First` = Apr 12 10:00 UTC, `Last` = Apr 25 10:00 UTC, `Days =
  14` (inclusive of both endpoints). Use `.Equal()` for time
  comparisons.
- **`single_day`** — one entry. Want `Days == 1` (single-day
  corpus exists for one day, NOT zero). Pairs locked decision §7.
- **`same_day_multiple_entries`** — three entries at different
  hours on the same UTC date. Want `Days == 1`.
- **`empty_corpus`** — `Span(nil)`. Want `First.IsZero() == true`,
  `Last.IsZero() == true`, `Days == 0`.

#### Test 4 — `TestStatsAggregate_EmptyInputContract`

Consolidated assertion that all three new helpers honor the empty-
input contract — `Streak(nil, fixedNow)` returns `(0, 0)`,
`MostCommon(nil, 5)` returns a non-nil empty slice (`len == 0`,
distinct from nil), `Span(nil)` returns a zero-value struct
(`First.IsZero()`, `Last.IsZero()`, `Days == 0`). Mirrors SPEC-018's
`TestAggregate_EmptyInputReturnsNonNilEmptySlice`. Pairs locked
decision §11.

### `internal/export/stats_test.go` (new file — 5 tests)

Tests against the shared fixture + explicit `StatsOptions`. No
cobra, no DB.

#### Test 5 — `TestToStatsMarkdown_DEC014FullDocumentGolden` (LOAD-BEARING — write FIRST)

Build `opts := StatsOptions{Now: fixedNow}`. Call `got, err :=
ToStatsMarkdown(fixture, opts)`. Assert `err == nil` and
`bytes.Equal(got, []byte(want))` where `want` is the literal string
below. If this fails, DEC-014 has been violated — fix code, not
test.

Expected output (byte-exact, no trailing newline — the CLI layer
adds one via `fmt.Fprintln`):

```
# Bragfile Stats

Generated: 2026-04-25T12:00:00Z
Scope: lifetime
Filters: (none)

## Stats

**Activity**
- Total entries: 8
- Entries/week: 4.00

**Streaks**
- Current: 2 days
- Longest: 3 days

**Top tags**
- auth: 4
- db: 2
- security: 2
- backend: 1
- perf: 1

**Top projects**
- alpha: 4
- beta: 2
- gamma: 1

**Corpus span**
- First entry: 2026-04-12
- Last entry: 2026-04-25
- Days: 14
```

Notes on the layout:
- `# Bragfile Stats` is the document title (parallels `# Bragfile
  Summary` and `# Bragfile Review`).
- Provenance block reuses DEC-014's three lines (`Generated:`,
  `Scope:`, `Filters:`). `Filters: (none)` is hard-coded — stats
  doesn't accept filter flags so the value is constant. `Scope:
  lifetime` is hard-coded — stats has no range.
- `## Stats` is the wrapper around the five bold sub-headers.
  Parallels summary's `## Summary` block + `## Highlights`
  wrapper depth.
- Five bold sub-headers: `**Activity**`, `**Streaks**`, `**Top
  tags**`, `**Top projects**`, `**Corpus span**`. Bullets under
  each. Mirrors DEC-014's markdown convention reuse of DEC-013's
  `**By type**` / `**By project**` count style.
- Tag and project bullets are DESC by count with alpha-ASC
  tiebreak (locked decisions §2 + §3): `auth (4)` first,
  `db (2)` and `security (2)` tied alpha-ASC, then `backend (1)`
  and `perf (1)` tied alpha-ASC. `refactor` (1) dropped at the
  cap.
- Corpus span dates rendered as `YYYY-MM-DD` (ISO 8601 calendar
  date, UTC).
- `Days: 14` rendered as plain integer.
- `Entries/week: 4.00` rendered as 2-decimal float (per locked
  decision §4 — even when the result is a round number, format
  consistently to 2 decimals).
- Streak metric labels include `days` suffix (`2 days`, `3 days`)
  for human readability — parallels the human-facing metric
  framing.

Assertions:
1. `bytes.Equal(got, []byte(want))` — single exact-byte compare.
2. On failure, print both `want` and `got` for diffability (copy
   the SPEC-014 / SPEC-015 / SPEC-018 / SPEC-019 helper pattern).
3. Line-based assertion on heading levels per AGENTS.md §9
   substring-trap addendum (SPEC-015 lesson): split `got` into
   lines and assert `lines[0] == "# Bragfile Stats"` and that
   `"## Stats"` appears as a standalone line, not as a substring
   of a deeper heading.

Pairs locked decisions §1, §2, §3, §4, §5, §7.

#### Test 6 — `TestToStatsJSON_DEC014ShapeGolden` (LOAD-BEARING — write SECOND)

Build `opts := StatsOptions{Now: fixedNow}`. Call `got, err :=
ToStatsJSON(fixture, opts)`. Assert `err == nil` and
`bytes.Equal(got, []byte(want))` where `want` is the literal JSON
below.

Expected output (byte-exact, no trailing newline):

```json
{
  "generated_at": "2026-04-25T12:00:00Z",
  "scope": "lifetime",
  "filters": {},
  "total_count": 8,
  "entries_per_week": 4,
  "current_streak": 2,
  "longest_streak": 3,
  "top_tags": [
    {
      "tag": "auth",
      "count": 4
    },
    {
      "tag": "db",
      "count": 2
    },
    {
      "tag": "security",
      "count": 2
    },
    {
      "tag": "backend",
      "count": 1
    },
    {
      "tag": "perf",
      "count": 1
    }
  ],
  "top_projects": [
    {
      "project": "alpha",
      "count": 4
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
  "corpus_span": {
    "first_entry_date": "2026-04-12",
    "last_entry_date": "2026-04-25",
    "days": 14
  }
}
```

Notes on the JSON shape:
- Top-level keys appear in struct-tag declaration order
  (`generated_at`, `scope`, `filters`, `total_count`,
  `entries_per_week`, `current_streak`, `longest_streak`,
  `top_tags`, `top_projects`, `corpus_span`) — Go's
  `encoding/json` preserves struct-tag order in `MarshalIndent`.
  Locked decision §1 part (2).
- `filters: {}` is an empty object (stats never accepts filter
  flags; the value is constant). Locked decision §1 part (4).
- `total_count: 8` — int.
- `entries_per_week: 4` — float64 with value 4.00. Go's
  `MarshalIndent` of `float64(4.0)` emits `4` (no decimal point)
  by default. The byte-exact golden bakes in this Go-stdlib
  behavior; consumers reading `jq '.entries_per_week'` get the
  numeric value `4` regardless of how it's serialized. The
  2-decimal "rounding" applies in computation (so 3.183 becomes
  3.18, not 3.183); the SERIALIZATION may drop trailing zeros.
  This is the load-bearing call: trust Go's stdlib JSON encoder
  for float formatting. Document explicitly in the spec so build
  doesn't try to add custom MarshalJSON.
- `top_tags` / `top_projects` are arrays of `{tag, count}` /
  `{project, count}` objects in DESC count + alpha-ASC tiebreak
  order (locked decision §2). Cap at 5 enforced — `top_tags`
  has exactly 5 elements; `top_projects` has 3 (only 3 distinct
  projects exist after `(no project)` exclusion).
- `corpus_span` is a sub-object (locked decision §5):
  `first_entry_date` is `"YYYY-MM-DD"` (ISO 8601 calendar date,
  UTC), `last_entry_date` same, `days` int. Both date fields
  populated for non-empty corpus.

Assertions:
1. `bytes.Equal(got, []byte(want))` — single exact-byte compare.
2. Parse `got` via `json.Decoder` and verify object's top-level
   keys appear in struct-tag declaration order (`generated_at`,
   `scope`, `filters`, `total_count`, `entries_per_week`,
   `current_streak`, `longest_streak`, `top_tags`, `top_projects`,
   `corpus_span`). This is the load-bearing key-order assertion
   that DEC-014 part (2) rests on.
3. ORDERING lock: parse `got` and assert
   `m["top_tags"][0]["tag"] == "auth"` (DESC count puts the
   highest-count tag first).
4. CAP lock: parse `got` and assert `len(m["top_tags"]) == 5`
   (cap enforced even though 6 distinct tags exist in the
   fixture).

Pairs locked decisions §1, §2, §3, §4, §5, §7.

#### Test 7 — `TestToStats_EmptyCorpusShape`

Subtests `markdown` and `json`. The DEC-014 empty-state contract
applied to stats.

**Subtest `markdown`:** render `ToStatsMarkdown([]storage.Entry{},
StatsOptions{Now: fixedNow})`. Expected bytes (byte-exact, no
trailing newline):

```
# Bragfile Stats

Generated: 2026-04-25T12:00:00Z
Scope: lifetime
Filters: (none)
```

NO `## Stats` wrapper. NO bold sub-headers. NO metric bullets.
Document ends after the `Filters:` line.

Assertions:
1. `bytes.Equal(got, []byte(want))`.
2. Line-split + walk: assert NO line equals `"## Stats"` or any
   of `**Activity**`, `**Streaks**`, `**Top tags**`, `**Top
   projects**`, `**Corpus span**` (line-based, not
   `strings.Contains`, per SPEC-015 substring-trap addendum). The
   negative assertions are the load-bearing empty-corpus contract.

**Subtest `json`:** render `ToStatsJSON([]storage.Entry{},
StatsOptions{Now: fixedNow})`. Expected bytes (byte-exact):

```json
{
  "generated_at": "2026-04-25T12:00:00Z",
  "scope": "lifetime",
  "filters": {},
  "total_count": 0,
  "entries_per_week": 0,
  "current_streak": 0,
  "longest_streak": 0,
  "top_tags": [],
  "top_projects": [],
  "corpus_span": {
    "first_entry_date": null,
    "last_entry_date": null,
    "days": 0
  }
}
```

Assertions:
1. `bytes.Equal(got, []byte(want))`.
2. Parse via `json.Unmarshal` into `map[string]any` and:
   - assert `m["total_count"] == float64(0)` (Go decodes JSON
     numbers as float64).
   - assert `m["top_tags"]` is a `[]any` with `len == 0` (NOT
     nil, NOT `null`) — locks empty-state non-nil rule.
   - assert `m["top_projects"]` same.
   - assert `m["corpus_span"].(map[string]any)["first_entry_date"]
     == nil` (JSON `null` decodes to Go nil).
   - assert `m["corpus_span"].(map[string]any)["days"] == float64(0)`.

Pairs locked decision §1 part (4) (empty-state arrays = `[]`,
date fields = null in JSON), §5 (corpus_span sub-object shape on
empty input).

#### Test 8 — `TestToStats_TopFiveCapEnforcedAtBoundary`

Subtests `tags` and `projects`. Each builds a 6-entry fixture
(distinct single-tag or single-project, all same UTC day) so 6
values tie at count=1; assert the renderer returns exactly 5
with alpha-ASC tiebreak determining which 5.

- **`tags`** — fixture entries with `Tags` ∈ `{zebra, yak, x-ray,
  wolf, vulture, umbrella}`. Render `ToStatsJSON`, parse via
  `json.Unmarshal`. Assert `len(m["top_tags"]) == 5`; the 5
  returned values are `umbrella, vulture, wolf, x-ray, yak` in
  order; `zebra` does NOT appear in any element's `tag` field.
- **`projects`** — same shape with the `Project` field instead
  of `Tags`. Same 5-cap + alpha-ASC assertion on `top_projects`
  using `project` as the inner key.

Pairs locked decisions §2 (array shape), §3 (strict cap + alpha
tiebreak).

#### Test 9 — `TestEntriesPerWeek_DecimalWeeksAndSubWeekZero`

Three subtests covering the `entries_per_week` computation locked
at §4. Tests render `ToStatsJSON`, parse, and assert on the
parsed `entries_per_week` value (NOT byte-golden — too brittle
across fixtures for float formatting). Use `math.Abs(got - want)
< 0.001` for float comparison.

- **`sub_week_zero`** — 3 entries spanning 2 UTC days (Apr 25 →
  Apr 26 inclusive, `span_days = 2`). Weeks = 2/7.0 ≈ 0.286 < 1
  → assert `entries_per_week == 0.0`.
- **`exactly_one_week`** — 7 entries, one per day across 7
  consecutive UTC dates (`span_days = 7`, weeks = 1.0). Assert
  `entries_per_week == 7.0`.
- **`partial_weeks_two_decimals`** — 5 entries spanning 11 days
  (weeks ≈ 1.5714, value ≈ 3.1818 → rounded to 3.18). Assert
  `entries_per_week ≈ 3.18`.

Pairs locked decision §4 (decimal-weeks formula, sub-week zero,
2-decimal rounding).

### `internal/cli/stats_test.go` (new file — 4 tests)

Reuse `seedListEntry` patterns from `internal/cli/list_test.go`
(same package, so direct reuse). Use `t.TempDir()` for DB paths;
never touch `~/.bragfile`.

#### Test 10 — `TestStatsCmd_BareDefaultsToMarkdown`

Run `brag stats` (no flags). Seed two fresh entries first via
`seedListEntry` so the corpus is non-empty (the markdown body
exists; we check the scope-line and headers).

Assertions:
1. `err == nil`; `errBuf.Len() == 0`.
2. `outBuf.String()` starts with `"# Bragfile Stats\n\n"`
   (markdown default).
3. Line-walk `outBuf.String()`:
   - one line equals `"Scope: lifetime"`.
   - one line equals `"Filters: (none)"`.
   - one line equals `"## Stats"` (entries section present
     because corpus is non-empty).
4. `Generated:` line matches the RFC3339 regex
   (`^Generated: \d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`) — reuse
   the `lineMatches` + `regexp.MustCompile` helpers from
   `summary_test.go` (same package, source-level reuse).

Pairs locked decisions §1 (default markdown + scope hard-code),
§8 (--format default).

#### Test 11 — `TestStatsCmd_FormatJSONShape`

Seed three fresh entries via `seedListEntry`:
- `{Title: "a", Tags: "auth", Project: "alpha", Type: "shipped"}`
- `{Title: "b", Tags: "auth,security", Project: "alpha", Type: "shipped"}`
- `{Title: "c", Tags: "perf", Project: "beta", Type: "learned"}`

Run `brag stats --format json`. Parse via `json.Unmarshal` into
`map[string]any`. Assertions:

1. `err == nil`; `errBuf.Len() == 0`; trailing newline on stdout.
2. All ten top-level keys present (`generated_at`, `scope`,
   `filters`, `total_count`, `entries_per_week`,
   `current_streak`, `longest_streak`, `top_tags`,
   `top_projects`, `corpus_span`).
3. `m["scope"] == "lifetime"`; `m["filters"]` is empty map;
   `m["total_count"] == float64(3)`.
4. `m["top_tags"]` first element has `tag == "auth"` with `count
   == 2` (DESC by count).
5. `m["top_projects"]` first element has `project == "alpha"`
   with `count == 2`; second is `beta` with count 1.
6. `m["corpus_span"]` has `first_entry_date`, `last_entry_date`,
   `days` keys present.

Pairs locked decisions §1, §2, §5, §8.

#### Test 12 — `TestStatsCmd_UnknownFormatIsUserError`

Run `brag stats --format yaml`. Assertions:
1. `err != nil`; `errors.Is(err, ErrUser) == true`.
2. `outBuf.String() == ""`.
3. Error message contains `yaml` AND `markdown` AND `json`.

Pairs locked decision §8 (DEC-007 RunE validation; --format
accepted set).

#### Test 13 — `TestStatsCmd_HelpShowsFormatOnly`

Run `brag stats --help`. Assertions:

1. `err == nil`; `errBuf.Len() == 0`.
2. `outBuf.String()` contains needles `--format`, `markdown`,
   `json` (positive substring assertions per AGENTS.md §9).
3. `outBuf.String()` DOES NOT contain any of `--tag`, `--project`,
   `--type`, `--out`, `--range`, `--since`, `--week`, `--month`
   — eight literal flag tokens that are NOT declared on stats.
   Iterate the forbidden list with `strings.Contains` per token
   and `t.Errorf` on any hit.

Locks decision §9 at the help-surface level. The SPEC-019-earned
Long-vs-help watch-pattern fires here — the design-time grep run
under Premise audit confirmed the locked Long sketch contains
none of these forbidden tokens, so the assertion passes without
build-time prose adjustment. Pairs locked decisions §8 + §9 + the
self-audit watch-pattern (second proactive validation case).

#### Test 14 — `TestStatsCmd_UndeclaredFlagsRejectedAsUnknown`

Subtests for each undeclared flag: `--tag`, `--project`, `--type`,
`--out`, `--range`, `--since`, `--week`, `--month`. For each:

- Run `brag stats --<flag> X` (or `--<flag>` for boolean ones).
- Assert `err != nil`.
- Assert `errors.Is(err, ErrUser) == false` — cobra's
  `unknown flag` error is NOT an `ErrUser` (it's cobra's own
  error type wrapped through). Assert by negation rather than
  positive identity, since the exact error type cobra returns
  is implementation detail.
- Assert `err.Error()` contains the literal `unknown flag` AND
  the offending flag name (e.g., `--tag`).

Locks decision §9 — undeclared flags surface as cobra unknown-
flag errors rather than the command rejecting them in `RunE`.

> **Note on test count:** The acceptance criteria reference 13
> failing tests; tests 1–4 land in `aggregate_test.go`, tests 5–9
> in `stats_test.go` (export package), tests 10–14 in
> `stats_test.go` (cli package). The fourteenth test is
> `TestStatsCmd_UndeclaredFlagsRejectedAsUnknown`; "13 tests" in
> the criteria refers to top-level tests (sub-tests count once
> per parent). Total subtest count is ~30.

## Implementation Context

### Decisions that apply

- `DEC-014` — rule-based output shape (envelope JSON + markdown
  provenance) consumed verbatim. SPEC-020 is the third consumer;
  goldens prove conformance. NO new DEC; if you think you need
  one, STOP and ask.
- `DEC-013` — markdown export shape; provenance + summary block
  conventions DEC-014 inherits. Stats's `## Stats` section + bold
  sub-headers parallel DEC-013's `## Summary` + `**By type**` /
  `**By project**` style.
- `DEC-011` — naked-array per-entry shape. Stats does NOT use it
  (no per-entry rendering in stats output).
- `DEC-007` — required-flag validation in `RunE`. Applies to
  `--format` validation. NEVER use `MarkFlagRequired`; always
  `UserErrorf` in `RunE`.
- `DEC-006` — cobra framework. New `brag stats` is a cobra
  subcommand following the pattern of every other.
- `DEC-004` — tags comma-joined TEXT. The renderer splits
  `e.Tags` on `,` (with trim) before passing to
  `aggregate.MostCommon`; no schema change.

### Constraints that apply

These constraints apply to the paths touched by this task (see
`/guidance/constraints.yaml` for full text):

- `no-sql-in-cli-layer` — the new CLI file `internal/cli/stats.go`
  MUST NOT import `database/sql`. Storage access goes through
  `storage.Open(...).List(...)`. Mirrors summary's structure.
- `stdout-is-for-data-stderr-is-for-humans` — markdown/JSON bodies
  to stdout via `cmd.OutOrStdout()`. Errors via the cobra return
  path; `main.go` writes them to stderr. Tests assert
  `errBuf.Len() == 0` on success paths.
- `errors-wrap-with-context` — `fmt.Errorf("...: %w", err)` for
  wrapped errors. `UserErrorf` for user-error paths.
- `test-before-implementation` — write the thirteen failing tests
  first, run `go test ./...`, confirm the expected failure modes
  (compile errors for missing types/funcs, assertion failures
  for the goldens), THEN implement.
- `one-spec-per-pr` — one feature branch + one PR for SPEC-020.

### Prior related work

- `SPEC-019` (shipped 2026-04-25) — direct precedent. SPEC-019
  created `internal/export/review.go`, extended
  `internal/aggregate` by ONE helper. SPEC-020 follows the same
  shape: `internal/export/stats.go` (sibling), extends
  `internal/aggregate` by THREE helpers (`Streak`, `MostCommon`,
  `Span`). Also: SPEC-019's build reflection captured the
  Long-vs-help self-audit watch-pattern; SPEC-020 applies it
  proactively at design.
- `SPEC-018` (shipped 2026-04-25) — emitted DEC-014, seeded
  `internal/aggregate`, established the Rejected-alternatives-
  build-time discipline this spec re-applies a third time.
- `SPEC-014` (shipped) — emitted DEC-011, created the
  `internal/export` package. Structural precedent for the
  goldens-locked DEC pattern.
- `SPEC-015` (shipped) — markdown export with DEC-013 provenance/
  summary conventions DEC-014 inherits. Substring-trap addendum
  applies to this spec's heading-level asserts.
- `SPEC-007` (shipped) — `Store.List(ListFilter{})` is the read
  path stats uses with a zero-value filter (lifetime corpus).

### Out of scope (for this spec specifically)

Explicit list of what this spec does NOT include. If any of these
feel necessary during build, create a new spec rather than
expanding this one.

- DEC emission. DEC-014 is consumed verbatim. Any "I think we
  need a DEC for X" thought during build is a STOP-and-ask.
- Refactor of any existing aggregate helper (`ByType`,
  `ByProject`, `GroupForHighlights`, `GroupEntriesByProject`).
  Sound future refactors out-of-scope.
- Modification of `internal/export/json.go`'s `entryRecord` /
  `toEntryRecord` helper. Stats doesn't render per-entry shapes.
- Any change to summary's or review's rendering or CLI behavior.
  Existing goldens MUST stay byte-identical.
- Modification of the parent stage doc.
- Any expansion of `--format` accepted values beyond `markdown`
  and `json`.
- Any addition of `--tag` / `--project` / `--type` / `--out` /
  `--range` / `--since` / `--week` / `--month`. Locked NO.
- `brag remind` / habit-enforcement command. Backlog.
- Backdating support (`brag add --at <date>`). Backlog.
- Configurable metric list. Six are baked.
- Calendar-week semantics for `entries_per_week`. Decimal-weeks
  is the locked formula.
- Time-zone configuration for streak/span. UTC-only for MVP.

## Notes for the Implementer

Gotchas, style preferences, reuse opportunities.

### Aggregate helpers

`internal/aggregate/aggregate.go` gains three functions and two
types. Mirror SPEC-018's `ByType`/`ByProject`/`GroupForHighlights`
construction patterns (sort.Slice with multi-key comparator;
non-nil empty slice on empty input; package-private buckets map).

**Types** (exact signatures):

```go
type NameCount struct {
    Name  string
    Count int
}

type CorpusSpan struct {
    First time.Time
    Last  time.Time
    Days  int
}
```

`NameCount` stays neutral (renamed-for-clarity sibling of
`TypeCount` / `ProjectCount`); the renderer wraps `[]NameCount`
into per-spec JSON shapes with semantic keys.

**Function signatures** (exact):

```go
func Streak(entries []storage.Entry, now time.Time) (current, longest int)
func MostCommon(values []string, n int) []NameCount
func Span(entries []storage.Entry) CorpusSpan
```

**Invariants per function:**

- **`Streak`** — guarantees: (a) empty entries → `(0, 0)`; (b)
  `now`'s UTC date with zero entries → `current = 0` (per locked
  decision §6, NOT "the streak that ended yesterday"); (c) walks
  back UTC-date by UTC-date from `now`'s UTC date while each
  date has ≥1 entry; (d) `longest` is the longest consecutive
  UTC-date run across the entire corpus, computed in one pass
  over a sorted unique-date set; (e) multiple entries on the
  same UTC date count as one streak day (de-dup by formatted
  date string). Implementation hint: use `e.CreatedAt.UTC().Format("2006-01-02")`
  as the date key in a `map[string]struct{}`; for the walk-back,
  start with `time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)`
  and decrement via `cursor.AddDate(0, 0, -1)`.
- **`MostCommon`** — guarantees: (a) empty input → non-nil
  empty slice (`make([]NameCount, 0)`); (b) empty-string values
  excluded from counting; (c) DESC by count with alpha-ASC
  tiebreak (`sort.Slice` comparator: count differs → DESC count;
  else name ASC); (d) STRICT cap at `n` via `out = out[:n]`
  AFTER sort (so the alpha-ASC tiebreak determines which `n`
  at the boundary); (e) fewer than `n` distinct values returns
  however many exist (no padding).
- **`Span`** — guarantees: (a) empty input → zero-value struct;
  (b) `First` / `Last` are UTC times of the earliest / latest
  `CreatedAt`; (c) `Days` is INCLUSIVE on both endpoints —
  compute via UTC-truncation:
  `firstDay := time.Date(first.Year(), first.Month(), first.Day(), 0, 0, 0, 0, time.UTC)`
  (same for `lastDay`), then
  `days := int(lastDay.Sub(firstDay).Hours()/24) + 1`. The `+1`
  is the inclusive-on-both-ends contract from locked decision §7.

Three helpers added in one pass (vs SPEC-019's one) because
SPEC-020's three lifetime aggregations don't map to existing
helpers.

### `internal/export/stats.go` (renderer)

Mirror SPEC-018's `summary.go` and SPEC-019's `review.go` patterns:
options struct with injectable `Now`; `ToXMarkdown` / `ToXJSON` pair;
`bytes.Buffer` + `fmt.Fprintln/Fprintf` for markdown; struct-tagged
envelope + `json.MarshalIndent` for JSON; `trimTrailingNewline` on
the markdown bytes.

**Public surface** (exact):

```go
type StatsOptions struct {
    Now time.Time
}

func ToStatsMarkdown(entries []storage.Entry, opts StatsOptions) ([]byte, error)
func ToStatsJSON(entries []storage.Entry, opts StatsOptions) ([]byte, error)
```

`StatsOptions` has NO `Scope` field (always renders `Scope:
lifetime`) and NO `Filters` / `FiltersJSON` fields (stats accepts
no filter flags; `Filters: (none)` in markdown and `"filters": {}`
in JSON are hard-coded constants in the renderer).

**Markdown invariants** (`ToStatsMarkdown`):
- Always emit: `# Bragfile Stats` + blank line + `Generated:
  <RFC3339>` + `Scope: lifetime` + `Filters: (none)`.
- Empty `entries` → return after the `Filters:` line (no `## Stats`
  wrapper, no sub-headers, no metric bullets) per DEC-014 part (4).
- Non-empty: blank line + `## Stats` + blank line + five sub-
  blocks separated by blank lines, in fixed order:
  `**Activity**` (Total entries: %d, Entries/week: %.2f),
  `**Streaks**` (Current: %d days, Longest: %d days),
  `**Top tags**` (bullets `- <tag>: <count>`),
  `**Top projects**` (bullets `- <project>: <count>`),
  `**Corpus span**` (First entry / Last entry as `YYYY-MM-DD`,
  Days as int). Trailing `\n` stripped via `trimTrailingNewline`.
- Format `entries_per_week` with `%.2f` so markdown shows `4.00`
  even when the value is a round integer (mirrors locked
  decision §4 — 2-decimal display in markdown). Markdown's
  formatter is independent of JSON's Go-stdlib float encoder.

**JSON invariants** (`ToStatsJSON`):
- Wire shape via private struct types with `json:` tags in the
  exact key order `generated_at`, `scope`, `filters`,
  `total_count`, `entries_per_week`, `current_streak`,
  `longest_streak`, `top_tags`, `top_projects`, `corpus_span`.
- `top_tags` element struct: `{Tag string \`json:"tag"\`; Count
  int \`json:"count"\`}`. `top_projects` element struct: same
  shape with `Project` / `\`json:"project"\`` (locked decision §2
  — semantic-name keys, not neutral `name`).
- `corpus_span` element struct uses `*string` for the two date
  fields so JSON marshals empty-corpus as `null` (and non-empty
  as `"YYYY-MM-DD"` after assigning address-of-local-string).
  Days stays `int` — `0` on empty corpus is the valid value.
- Initialize empty slices as non-nil (`[]tagCount{}`,
  `[]projCount{}`) so empty corpus marshals as `[]` not `null`
  per DEC-014 part (4).
- Filters initialized as `map[string]string{}` (always).
- Trust Go's stdlib `MarshalIndent` (indent="  ") for float
  formatting — `float64(4.0)` serializes as `4`; `float64(3.18)`
  as `3.18`. Do NOT add a custom `MarshalJSON`; the byte-exact
  goldens bake the stdlib behavior.

**Renderer-local helpers** (private; live in `stats.go`):
- `computeEntriesPerWeek(total, spanDays int) float64` — decimal-
  weeks formula: `weeks := float64(spanDays) / 7.0`; if `weeks <
  1.0` return `0.0`; else `math.Round((float64(total)/weeks)*100)
  / 100`. See "Decimal-weeks formula + worked examples" below.
- `extractTags(entries []storage.Entry) []string` — splits each
  entry's `Tags` on `,` per DEC-004, `strings.TrimSpace`s each
  token, drops empties, returns flat list for `MostCommon`.
- `extractProjects(entries []storage.Entry) []string` — returns
  each entry's non-empty `Project` field; empty `Project` is
  EXCLUDED before counting (locked decision §3 — `(no project)`
  excluded from `top_projects`).

These adapter helpers stay in `internal/export` (not
`internal/aggregate`) because they couple to `storage.Entry`'s
field layout — a renderer-side concern, not pure data aggregation.
The aggregate layer stays generic over `[]string`.

**Decimal-weeks formula + worked examples** (load-bearing,
locked decision §4):

```
weeks := float64(spanDays) / 7.0
if weeks < 1.0 { return 0.0 }
return math.Round((float64(total)/weeks)*100) / 100
```

| total | spanDays | weeks  | entries_per_week |
| ----- | -------- | ------ | ---------------- |
| 8     | 14       | 2.000  | 4.00             |
| 5     | 11       | 1.571… | 3.18             |
| 3     | 1        | 0.143… | 0.0  (sub-week)  |
| 7     | 7        | 1.000  | 7.00             |
| 1     | 1        | 0.143… | 0.0  (sub-week)  |
| 14    | 14       | 2.000  | 7.00             |

The `< 1.0` threshold (NOT `<= 1.0`) means span_days = 7 → weeks
= 1.0 → entries_per_week = total. Sub-week corpora always return
0.0.

### `internal/cli/stats.go` (CLI command)

Mirror SPEC-018's `summary.go` and SPEC-019's `review.go` patterns:
`NewStatsCmd() *cobra.Command` constructor + private `runStats`;
`--format` validated in `RunE` per DEC-007; storage opened via
`config.ResolveDBPath` + `storage.Open`; renderer dispatched via
switch on format.

**Public surface** (exact):

```go
func NewStatsCmd() *cobra.Command
```

**Flags** (only one declared):

```go
cmd.Flags().String("format", "markdown", "output format (one of: markdown, json)")
```

NO other flags — `--tag` / `--project` / `--type` / `--out` /
`--range` / `--since` / `--week` / `--month` are GENUINELY
UNDECLARED so cobra surfaces them as `unknown flag` errors per
locked decision §9.

**`runStats` invariants:**
- Validate `--format` against `{markdown, json}`; unknown value
  → `UserErrorf` naming the value + accepted set (DEC-007).
- Storage call: `s.List(storage.ListFilter{})` (zero-value filter
  → every row, lifetime corpus).
- **Single `Now` source** (locked decision §10, load-bearing for
  testability): take `time.Now().UTC()` ONCE in `runStats`,
  store it on `export.StatsOptions.Now`, pass that value through
  to the renderer. The renderer in turn passes the SAME value to
  `aggregate.Streak(entries, opts.Now)`. This avoids a subtle
  midnight-UTC race where two separate `time.Now()` calls
  bracket the streak computation and disagree on "today's" UTC
  date. Mirrors SPEC-018's `MarkdownOptions.Now` /
  `SummaryOptions.Now` injectable-clock pattern.
- Body to stdout via `fmt.Fprintln(cmd.OutOrStdout(), string(body))`
  — adds the trailing newline (matches SPEC-018/019 stdout
  convention).

**Long string content** (the cobra `Long` field, load-bearing
for the help-vs-Long self-audit):

```
Print six lifetime aggregations over the entire corpus: total entries, entries/week (rolling average over the corpus span), current and longest streak (consecutive UTC days with entries), top-5 most-common tags, top-5 most-common projects, plus the corpus span (first entry, last entry, days). No LLM ships in the binary.

Output is markdown (default) or a single-object JSON envelope (--format json) per DEC-014. The JSON shape uses arrays of objects for top_tags and top_projects to preserve DESC-by-count ordering; corpus_span is a sub-object with date-or-null fields.

Stats covers the lifetime corpus only — no time window, no filters. Use brag summary for windowed digests; brag review for reflection over the last 7 or 30 days. Stdout only; redirect with > if you want a file.

Examples:
  brag stats                        # lifetime corpus, markdown
  brag stats --format json          # lifetime corpus, JSON envelope
```

The `Short` field is: `"Lifetime stats: six aggregations over
the entire corpus"`.

**Self-audit grep on the Long block above** (the SPEC-019 watch-
pattern applied proactively at design — load-bearing):

```
grep -nE -- '--tag|--project|--type|--out|--range|--since|--week|--month' \
  <the Long block above, exactly as written>
```

Result: **zero hits.** The Long mentions `--format` (the only
declared flag) and references "windowed digests" / "reflection
over the last 7 or 30 days" as descriptive phrases, but uses
none of the eight forbidden `--`-prefixed tokens. The
`TestStatsCmd_HelpShowsFormatOnly` negative-substring assertions
will pass on this Long without build-time prose adjustment. ✅
This is the SPEC-019-earned watch-pattern catching at design what
SPEC-019 had to catch at build — second proactive validation
case.

**Implementer warning:** if you change the Long string during
build (tightening prose, adding examples, etc.), RE-RUN the
self-audit grep on the modified Long before declaring build
complete. Any hit on a forbidden token means either the help
test will fail OR you've reverted to the exact SPEC-019 build-
time deviation.

### `cmd/brag/main.go` update

One line added after the existing nine `AddCommand` calls (line
25 currently calls `NewReviewCmd`):

```go
root.AddCommand(cli.NewStatsCmd())
```

The full block becomes:

```go
root.AddCommand(cli.NewAddCmd())
root.AddCommand(cli.NewListCmd())
root.AddCommand(cli.NewShowCmd())
root.AddCommand(cli.NewDeleteCmd())
root.AddCommand(cli.NewEditCmd())
root.AddCommand(cli.NewSearchCmd())
root.AddCommand(cli.NewExportCmd())
root.AddCommand(cli.NewSummaryCmd())
root.AddCommand(cli.NewReviewCmd())
root.AddCommand(cli.NewStatsCmd())
```

### Test file harness

`internal/cli/stats_test.go` follows `summary_test.go`'s pattern
exactly — `newStatsTestRoot(t)` builds a fresh root with the stats
subcommand attached, isolates `BRAGFILE_DB`, returns
`*cobra.Command` + `outBuf` + `errBuf`. `runStatsCmd(t, dbPath,
args...)` builds the args slice with `--db <dbPath> stats` prefix.

Reuse `firstChars` and `lineMatches` helpers from
`summary_test.go` (same package, source-level reuse).

### Doc updates checklist

Run these greps before AND after the doc sweep to confirm the
audit-grep cross-check both-sides discipline:

```
grep -rn "brag stats" docs/ README.md AGENTS.md
grep -rn "internal/aggregate" .
grep -rn "Streak\|MostCommon\|Span" internal/
```

Expected post-sweep state:
- `grep -rn "brag stats" docs/ README.md AGENTS.md` returns hits
  in api-contract.md (new section + DEC-014 row updated), tutorial.md
  (Scope blurb updated, optional subsection added), README.md
  (updated Scope blurb), AGENTS.md (existing glossary entries
  unchanged), data-model.md (DEC-014 row updated). NO hit
  references stats as deferred or forward-referenced.
- `grep -rn "Streak\|MostCommon\|Span" internal/` returns hits in
  `internal/aggregate/aggregate.go` (the three new functions),
  `internal/aggregate/aggregate_test.go` (the four new tests),
  `internal/export/stats.go` (the renderer calls).

If any post-sweep grep returns an unexpected hit, audit it before
declaring the sweep done.

### Optional tutorial subsection (recommended)

Under `docs/tutorial.md` §4 "Read them back", add (after the
SPEC-019 `### Weekly reflection: brag review` subsection):

```markdown
### Lifetime stats: brag stats

Run `brag stats` for the lifetime panorama: total entries,
entries-per-week rolling average, current and longest streak,
top-5 most-common tags and projects, plus your corpus span:

    brag stats

Or pipe the JSON form into your favorite LLM for a "what does my
year look like?" prompt:

    brag stats --format json | claude "summarize my brag history"

Stats is corpus-wide — there are no filter or range flags. Use
`brag summary` for windowed digests, or `brag review` for
reflection over the last 7 or 30 days.
```

Keep the prose short — the tutorial value is the *paste-into-AI*
pattern, the same shape as the SPEC-018/019 subsections.

---

## Build Completion

*Filled in at the end of the **build** cycle, before advancing to verify.*

- **Branch:** `feat/spec-020-brag-stats-six-lifetime-metrics`
- **PR (if applicable):** opened post-`just advance-cycle SPEC-020 verify`
- **All acceptance criteria met?** yes — 17/17 (corrected from
  build-time miscount of 16; verify caught). Aggregate, export
  goldens (markdown + JSON byte-exact), empty-corpus contract, top-5
  cap with alpha-ASC tiebreak, decimal-weeks formula, CLI defaults,
  `--format json` envelope, unknown-format ErrUser, help shows
  `--format` only, eight undeclared flags rejected as cobra unknown
  flags, `brag --help` lists `stats`, all gates green, doc sweep
  complete.
- **New decisions emitted:**
  - none — DEC-014 consumed verbatim per Locked design decisions §1.
    SPEC-020 is the third (and last) DEC-014 consumer.
- **Deviations from spec:**
  - none. Six rejected-alternative paths held: (1) array-of-objects
    for `top_tags`/`top_projects` (not map-keyed); (2) decimal-weeks
    `entries_per_week` (not integer); (3) strict cap at 5 with
    alpha-ASC tiebreak (not include-all-ties); (4) `MostCommon`
    operating on `[]string` (not `[]storage.Entry`); (5) three
    aggregate helpers added (not inlined in renderer); (6) `Span.Days`
    inclusive on both endpoints. Both load-bearing goldens passed
    byte-exact on the first build run; existing SPEC-014/018/019
    goldens stayed byte-identical.
- **Follow-up work identified:**
  - STAGE-004 ships when SPEC-020 ships (last spec in the backlog).
    Stage-level reflection runs before the stage is closed.
  - `revew1.md`, `status-after-nine-specs.md`, and the
    `framework-feedback/` directory are tracked-as-untracked in the
    git status; left untouched (out of scope for SPEC-020).

### Build-phase reflection (3 questions, short answers)

Process-focused: how did the build go? What friction did the spec create?

1. **What was unclear in the spec that slowed you down?**
   — Nothing material. The 17%-trim experiment (signatures +
   invariants instead of full skeletons) was sufficient. Three
   places where the spec leaned on precedent rather than full code:
   (a) the `Streak` longest-run loop was implementable directly from
   the invariant ("longest consecutive UTC-date run anywhere in the
   corpus") + the implementation hint about UTC-truncation; (b) the
   `corpusSpanRecord.*string` pattern for nullable JSON dates was
   stated explicitly enough; (c) the `extractTags`/`extractProjects`
   adapter design was stated explicitly. The shared fixture's
   pre-computed expected values (top-5 ordering, span days, streak
   counts, entries/week) made the goldens trivial to reproduce
   byte-exact on the first try. Net: signatures + invariants were
   enough; no missing sketch slowed the build.

2. **Was there a constraint or decision that should have been listed but wasn't?**
   — No. The SPEC-019-earned Long-vs-help negative-substring self-
   audit watch-pattern caught nothing in the locked Long (design-time
   audit clean; build-time re-grep of the rendered `--help` clean —
   second proactive validation case). The §9 audit-grep cross-check
   addendum (SPEC-018-earned) ran clean both times — design-side
   enumeration matched the actual `grep "brag stats" docs/ README.md
   AGENTS.md` output verbatim; build-side re-run before doc-sweep
   surfaced no deltas. Both addenda earned their second confirming
   case here; ship reflection should propose §12 codification for
   both together.

3. **If you did this task again, what would you do differently?**
   — Almost nothing. The fail-first run produced exactly the
   expected `undefined: Streak / NameCount / StatsOptions /
   ToStatsMarkdown / ToStatsJSON / NewStatsCmd` symbol errors before
   any implementation; the goldens passed byte-exact on first run
   for both markdown and JSON; the eight-flag undeclared-flag
   subtests passed without prose adjustment. The only minor friction:
   I added a Go-comment-line `// No filter flags, no --range, no
   --out.` above `NewStatsCmd` that contained two of the forbidden
   tokens — harmless because comments don't reach `--help`, but if
   future iterations of the watch-pattern grep narrow scope to "any
   line in the CLI source file naming the forbidden tokens" rather
   than "the rendered Long string only," that comment would surface
   as a false positive. The watch-pattern's scope (spec→binary→test
   path, NOT spec-prose→reader path) is correctly stated in the
   spec's Premise audit; build held to that scope.

---

## Reflection (Ship)

*Appended during the **ship** cycle. Outcome-focused reflection, distinct
from the process-focused build reflection above.*

1. **What would I do differently next time?**
   — The trim experiment was the load-bearing data point of this
   build, and it validated cleanly: zero thrash, both goldens
   byte-exact on first run, no missing skeleton bits surfaced at
   build time. The actionable lesson is more specific than
   "compress when precedent exists" — the precondition that
   mattered is **structural analogy**, not just that prior specs
   are in `done/`. SPEC-020 worked because it followed SPEC-019's
   exact shape (consume DEC-014 verbatim, extend
   `internal/aggregate`, new sibling renderer, new sibling CLI
   file), and SPEC-019 followed SPEC-018's shape. The construction
   was already proven twice; the spec just needed to point at the
   proof.

   For STAGE-005: the **first spec in any new stage should keep a
   fuller Notes-for-the-Implementer skeleton; specs 2+ in a stage
   can compress to signatures + invariants only when the first
   spec's shape is the construction precedent.** STAGE-005's first
   spec (likely homebrew distribution / goreleaser) has no in-stage
   precedent and shouldn't trim. The second STAGE-005 spec, if it
   shares a shape with the first, can. Worth carrying into SPEC-021
   design as a heuristic to apply, not yet a rule to codify (N=1
   for the trim itself; SPEC-018 + SPEC-019 both used fuller
   skeletons).

2. **Does any template, constraint, or decision need updating?**
   — Yes. **Codify the SPEC-019 negative-substring self-audit
   watch-pattern as a §12 addendum at this ship.** Two confirming
   cases (SPEC-019 build self-catch + SPEC-020 design pre-empt),
   the rule is concrete and greppable, and the scope clause
   (spec→binary→test path, NOT spec-prose→reader path) earned its
   place when SPEC-020 build self-caught a Go comment containing
   forbidden tokens that wouldn't reach `--help`. Suggested
   addendum text — keep tight:

   > **NOT-contains assertions need a self-audit grep against
   > load-bearing prose.** When a Failing Test asserts output
   > DOES NOT contain "X", grep the spec's load-bearing text
   > (the Long / help-rendering / doc-rendering prose that the
   > implementer types into the binary) for X at design time.
   > Hits in load-bearing prose are bugs to fix BEFORE locking.
   > Hits in commentary or design-justification prose are fine
   > — those don't reach the binary. Lesson earned in SPEC-019
   > build reflection (caught at build) and SPEC-020 design
   > (caught proactively).

   **Hold the trim heuristic for one more confirming case.**
   STAGE-005's shape is sufficiently different (distribution, CI,
   packaging — not the aggregate/render symmetry STAGE-004 had)
   that the trim might not transfer cleanly. Mention as a flag
   for SPEC-021 design; don't codify yet.

   **§9 audit-grep cross-check addendum stands as written** (per
   verify N1) — second confirming case ran clean unchanged; no
   refinement surface area appeared.

3. **Is there a follow-up spec I should write now before I forget?**
   — No spec. STAGE-005's six workstreams are already queued in
   `brief.md`; SPEC-021 emerges naturally when STAGE-005 starts.
   The `README.md:79` STAGE-004→STAGE-005 typo is a one-line
   chore — bundled into the stage-ship commit.

   **One small flag worth carrying into STAGE-005's first spec
   design:** doc sweeps should grep for stale stage-references
   (`STAGE-00N` mentions) at the start of each stage as a routine
   sanity check; the SPEC-020 verify accidentally surfaced that
   the discipline isn't currently routine.
