---
# Maps to ContextCore insight.* semantic conventions.

insight:
  id: DEC-030
  type: decision
  confidence: 0.72
  audience:
    - developer
    - agent

agent:
  id: claude-opus-4-8
  session_id: null

project:
  id: PROJ-004
repo:
  id: bragfile

created_at: 2026-07-06
supersedes: null
superseded_by: null

tags:
  - cli
  - io-contract
  - aggregation
  - time
  - digest
  - wrapped
  - human-consumer
  - ai-consumer
---

# DEC-030: `brag wrapped` — named-calendar-period selection, bounded window, and the celebratory section taxonomy

## Decision

`brag wrapped` is the fifth consumer of DEC-014's rule-based digest envelope
(after `summary`/`review`/`stats`/`impact`). It is a shareable, celebratory
year- or quarter-in-review. Five choices are locked; the envelope, empty-state
rules, markdown conventions, and JSON indent are inherited from DEC-014
verbatim, and the calendar-period *concept* is inherited from DEC-028 with one
deliberate divergence (the upper bound).

1. **The period is named positionally: `brag wrapped [<year>] [Q<n>]`.**
   `brag wrapped 2026` = calendar year 2026. `brag wrapped 2026 Q3` = calendar
   quarter Q3 2026 (case-insensitive `q<1..4>`). The `scope` field echoes
   `"2026"` / `"2026-Q3"`. This is a *positional* surface, deliberately NOT
   `impact`'s `--year`/`--quarter` bool flags — those carry the current-period-
   up-to-now (`[cutoff, now]`) meaning; `wrapped` names a **bounded** period a
   human says out loud. Year is 4 digits, plausibility-bounded [2000,2999];
   `Q0`/`Q5`/non-`Q`/extra tokens are a `UserError`.

2. **The default period (no argument) is the CURRENT calendar year.** Not the
   last-completed year. A "wrapped" is usually a retrospective of a finished
   period, but the completed-period variant is exactly what the planned
   `--previous` modifier (SPEC-053, foreseen by DEC-028) will add; baking
   last-completed into the *default* now would preempt SPEC-053 and split the
   "which period" logic across two specs. Current-year default mirrors DEC-028
   choice 1's current-in-progress choice, keeping the digest family consistent;
   `--previous` then layers on cleanly.

3. **The window is BOUNDED on both ends: `[period-start, next-boundary)`.**
   This is the deliberate divergence from `impact`, whose window is
   `[cutoff, now]` (the upper bound is implicit — every stored `created_at <=
   now`). A named calendar period has a real END that is NOT "now": `brag
   wrapped 2026` run in Feb 2027 must cover Jan–Dec 2026, not spill into 2027.
   `start = time.Date(year, startMonth, 1, 0,0,0,0, UTC)`; the exclusive upper
   boundary is the next period's start (`start.AddDate(1,0,0)` for a year,
   `start.AddDate(0,3,0)` for a quarter). All period math uses `time.Date`
   calendar constructors, never day subtraction (inheriting DEC-028's rule).
   `storage.ListFilter` has only a `Since` lower bound, so the read is
   `List(ListFilter{Since:start})` and the `created_at < nextBoundary` upper
   filter is applied in Go in the CLI layer — SQL-free, `no-sql-in-cli-layer`
   intact.

4. **The section taxonomy is a celebratory arc: Cadence → Top initiatives →
   Impact moments → Rhythm → Span.** Curated from the existing
   `internal/aggregate` toolbox, deliberately DISTINCT from `stats` (lifetime
   metrics dump) and `impact` (impact-only):
   - **Cadence** — total entries + busiest month + a per-month bucket **series**
     (12 buckets for a year, 3 for a quarter, zero-filled). This is the
     **sparkline-ready data slot** (choice 5).
   - **Top initiatives** — top-5 projects by count, EXCLUDING `(no project)`
     (reuses `MostCommon(extractProjects,5)`, same helper+exclusion as `stats`'
     top_projects — a shareable reel should not surface `(no project)` as an
     "initiative").
   - **Impact moments** — with-impact entries grouped by project, impact text
     in full (reuses `WithImpact` + `GroupEntriesByProject`, the `impact`
     rendering shape).
   - **Rhythm** — LONGEST streak only (not current — current is a live-corpus
     metric, meaningless scoped to a named past period), top-5 tags, top-3 types.
   - **Span** — first/last entry date + active days (`aggregate.Span`).

5. **The `cadence.series` slot is present as DATA in both renderers now, so the
   SPEC-052 visual pass drops a sparkline into it WITHOUT reshaping the
   envelope.** The JSON `cadence` object is
   `{busiest_month: <"YYYY-MM"|null>, series: [{period:"YYYY-MM", count:N}, ...]}`;
   the markdown renders the same series as a `- YYYY-MM: N` bullet list under
   `## Cadence`. Even on an empty period the series is fully present (zero-filled)
   so SPEC-052 renders a flat sparkline, not a gap. This is the text-first /
   visual-later boundary the stage brief draws between SPEC-051 and SPEC-052.

JSON envelope skeleton (per-spec keys; provenance shared with DEC-014):

```json
{
  "generated_at": "2026-12-31T23:59:59Z",
  "scope": "2026",
  "filters": {},
  "total_entries": 7,
  "cadence": {
    "busiest_month": "2026-04",
    "series": [ { "period": "2026-01", "count": 1 }, ... ]
  },
  "top_initiatives": [ { "project": "alpha", "count": 3 } ],
  "impact_moments": [
    { "project": "alpha",
      "entries": [ { "id": 1, "title": "kickoff", "project": "alpha",
                     "impact": "cut p95 login latency 40%" } ] }
  ],
  "longest_streak": 2,
  "top_tags": [ { "name": "api", "count": 3 } ],
  "top_types": [ { "name": "shipped", "count": 4 } ],
  "span": { "first_entry_date": "2026-01-15",
            "last_entry_date": "2026-11-30", "active_days": 320 }
}
```

Markdown envelope skeleton:

```
# Bragfile Wrapped

Generated: 2026-12-31T23:59:59Z
Scope: 2026
Filters: (none)
Entries: 7

## Cadence

Busiest month: 2026-04 (2)

- 2026-01: 1
...

## Top initiatives

- alpha: 3

## Impact moments

### alpha

- 1: kickoff
  cut p95 login latency 40%

## Rhythm

Longest streak: 2 days

**Top tags**
- api: 3

**Top types**
- shipped: 4

## Span

- First entry: 2026-01-15
- Last entry: 2026-11-30
- Active days: 320
```

## Context

`brag wrapped` is STAGE-013's headline polish feature — the user-requested
"fun"/shareable surface that makes the story surface land. It reads the same
corpus the analytical digests read, but shapes it as a **retrospective
highlight reel** over a named calendar period, not a metrics table.

Four choices needed a decision the existing DEC-014/DEC-028 consumers did not
settle:

- **How the period is named + the default.** `impact` takes current-period bool
  flags; a "wrapped" names a specific past period ("2026", "2026 Q3"). Positional
  args read the way a human says it and avoid overloading `impact`'s flags with a
  second semantics. The default is current-year (not last-completed) so the
  last-completed variant stays clean for `--previous` (SPEC-053) rather than being
  split across two specs.
- **The upper bound.** Every existing calendar digest is `[cutoff, now]` — the
  upper edge is implicit. A named period has a real END. This is the one place
  `wrapped` must diverge from `impact`, and it needs an explicit, tested bounded
  window so a future reader files it as intentional, not drift.
- **The section taxonomy.** Left unresolved, build would guess which of the
  toolbox's helpers to surface and in what order. Locking Cadence→Initiatives→
  Impact→Rhythm→Span gives the digest a celebratory arc distinct from `stats`
  and `impact`, and pins each section to a specific existing helper (no new
  aggregation except cadence bucketing).
- **The sparkline seam.** SPEC-052 is a separate spec, but its slot must exist
  now or SPEC-052 reshapes the envelope. Locking `cadence.series` as present-as-
  data (zero-filled, sparkline-ready) draws the text-first/visual-later boundary
  cleanly.

DEC-014 and DEC-028 are extended, not relitigated: envelope, empty-state,
markdown convention, JSON indent, project=initiative grouping, impact-first
`WithImpact` selection, and `time.Date` calendar math are all inherited. Only the
period-naming surface, the default, the bounded upper edge, and the section
taxonomy are new.

## Alternatives Considered

- **Option A: `--year`/`--quarter` flags (reuse `impact`'s surface).**
  - Why rejected: those flags mean "current year/quarter up to now"
    (`[cutoff, now]`). `wrapped` names a bounded past period. Overloading the
    same flags with two window semantics is the confusion DEC-028 already warns
    about ("a consumer reasoning about what `--month` means must know which
    command they're in"). A distinct positional surface keeps the two clean.

- **Option B: Default = last-completed period.**
  - Why rejected: it is the truer "wrapped" framing, but it is precisely what
    `--previous` (SPEC-053, foreseen by DEC-028) exists to add. Baking it into
    the default now preempts that spec and splits the period-selection logic.
    Current-year default is consistent with DEC-028's current-in-progress choice
    and leaves `--previous` a clean, additive layer.

- **Option C: `[period-start, now]` window (inherit `impact`'s upper bound).**
  - Why rejected: wrong for a named past period. `brag wrapped 2026` run in 2027
    must not include 2027 entries, and must include all of Dec 2026. The bounded
    `[start, next-boundary)` upper edge is the only correct, testable semantics
    for a named calendar period.

- **Option D: A metrics-dump taxonomy (totals-first, by-type block, current
  streak).**
  - Why rejected: that is `stats`' analytical framing. A shareable wrapped leads
    with cadence (the shape of your year), curates initiatives + impact moments,
    and shows longest (not current) streak — a retrospective, not a live
    dashboard. Leading with impact is `impact`'s job.

- **Option E: Defer the cadence data slot to SPEC-052.**
  - Why rejected: SPEC-052 would then have to reshape the envelope to add the
    series, breaking the "extend, never reshape" discipline every DEC-014
    consumer has held. Present-as-data-now / rendered-as-sparkline-later is the
    boundary the stage brief draws; the slot must exist here.

- **Option F (chosen): positional named period + current-year default + bounded
  window + celebratory taxonomy + sparkline-ready cadence slot, over the DEC-014
  envelope.**
  - Why selected: each new choice is grounded in the shareable-retrospective
    workflow, and everything inheritable from DEC-014/DEC-028 is inherited.
    `wrapped` reads as a natural fifth digest sibling with exactly one deliberate
    divergence (the bounded upper edge), tested explicitly.

## Consequences

- **Positive:** Reuses the entire `internal/aggregate` toolbox and the DEC-014
  envelope; no schema change; no new dependency. `jq .scope`, `.cadence.series`,
  `.impact_moments`, `.longest_streak` transfer cleanly. SPEC-052 drops a
  sparkline into `cadence.series` with zero envelope change. SPEC-053's
  `--previous` layers onto the bounded-window machinery already built here.
- **Negative:** A THIRD window semantics now exists across the family: rolling
  (`summary`/`review`), calendar-to-now (`impact`), and bounded-calendar
  (`wrapped`). Mitigated: the `scope` echo is command-specific and documented;
  the bounded divergence is explicit here and enforced by a dedicated failing
  test (`TestWrappedCmd_BoundedWindow`).
- **Negative:** The upper-bound filter runs in Go (the CLI layer) because
  `ListFilter` has no `Until`. Mitigated: it is a one-line `created_at <
  nextBoundary` filter, stays SQL-free, and keeps `no-sql-in-cli-layer` intact;
  adding an `Until` to `ListFilter` was considered heavier than warranted for
  one caller (revisit if a second bounded-window consumer appears).
- **Neutral:** Cadence buckets are monthly for both year (12) and quarter (3)
  scopes — uniform `YYYY-MM` period labels so SPEC-052 renders one sparkline
  shape. A weekly cadence for the quarter view is a future call, not baked here.
- **Neutral:** `wrapped` surfaces longest-not-current streak, so the `now` seam
  affects only the `Generated:` line, not any metric — `now` stays injectable
  for deterministic goldens but does not couple to the streak number.

## Validation

Right if:
- `brag wrapped 2026 --format json | jq .scope` returns `"2026"`;
  `brag wrapped 2026 Q3` returns `"2026-Q3"`; bare `brag wrapped` returns the
  current calendar year.
- `brag wrapped 2026` run when "now" is in 2027 includes an entry from
  2026-06-15 and excludes one from 2025-12-31 AND one from 2027-01-01 — i.e. the
  upper bound is the period end, not now. (The load-bearing bounded-window
  assertion.)
- `brag wrapped 2026 Q3` includes 2026-07-01 and 2026-09-30 and excludes
  2026-06-30 and 2026-10-01.
- The `cadence.series` has one bucket per month in scope (12 / 3), zero-filled,
  in BOTH renderers, even on an empty period; `busiest_month` is `null` when the
  period is empty.
- `Top initiatives` never contains `(no project)`; `Rhythm` shows longest, never
  current, streak.
- A malformed period (`Q5`, `notayear`, extra token) and an unknown `--format`
  are `UserError`s on stderr with a non-zero exit and empty stdout.

Revisit if:
- A monthly or weekly `wrapped` scope is wanted (new spec; keep the bounded-
  window machinery, add the scope).
- A second bounded-window consumer appears (then promote the Go upper-bound
  filter into a `ListFilter.Until` field with a small DEC).
- SPEC-052 needs a cadence field the `{period,count}` bucket omits (widen the
  bucket deliberately, with a test).

Confidence: 0.72. The envelope-inheritance and taxonomy choices (1 partial, 3,
4, 5) are strong (0.8–0.9) — they reuse proven DEC-014/DEC-028/aggregate
machinery and pin each section to an existing helper. The composite is dragged
by choice 2 (current-year-vs-last-completed default): a "wrapped" idiomatically
means the finished period, and a user could reasonably expect `brag wrapped`
alone to show last year. The choice to default to current-year (leaving
last-completed to `--previous`/SPEC-053) is a workflow-consistency call, not an
obvious one — it is the first revisit trigger and is recorded as an open question
in `/guidance/questions.yaml` (above §14's 0.7 hard threshold, but filed
proactively as the softest sub-choice, matching DEC-028's precedent). The
positional-vs-flag
surface (choice 1) is the second soft spot but lower-risk (it does not change
what data is shown, only how the period is named).

## References

- Related specs:
  - SPEC-051 (emits this DEC; wires `brag wrapped [<year>] [Q<n>]`; adds
    `internal/export/wrapped.go`, `internal/cli/wrapped.go`,
    `aggregate.Cadence`; reuses `MostCommon`/`Streak`/`Span`/`WithImpact`/
    `GroupEntriesByProject`).
  - SPEC-048 (shipped) — `brag impact`; DEC-028; the `nowFunc` seam, the
    calendar core in `window.go`, the impact renderer shape `wrapped` mirrors.
  - SPEC-020 (shipped) — `brag stats`; `Streak`/`MostCommon`/`Span` and the
    `extractProjects` non-empty exclusion `wrapped`'s Top initiatives reuses.
  - SPEC-052 (planned) — the sparklines/visual pass; consumes `cadence.series`.
  - SPEC-053 (planned) — `--previous`; layers the last-completed period onto
    `wrapped`'s bounded-window machinery.
- Related decisions:
  - DEC-014 (rule-based output envelope) — EXTENDED verbatim (envelope,
    empty-state, markdown convention, indent).
  - DEC-028 (`brag impact` calendar windows) — the calendar-period concept and
    `time.Date` math reused; `wrapped` DIVERGES on the upper bound (bounded
    `[start, next-boundary)` vs `impact`'s `[cutoff, now]`), and reuses
    project=initiative grouping + impact-first `WithImpact` selection.
  - DEC-022 (local-day streak) — `Streak` reused; `wrapped` surfaces only
    longest (period-scoped), decoupling the streak number from the `now` seam.
  - DEC-013 (count-ordering) — inherited transitively via the aggregate helpers.
  - DEC-007 (required/invalid-flag validation via `UserErrorf`).
- Related constraints: `stdout-is-for-data-stderr-is-for-humans`,
  `no-sql-in-cli-layer` (the upper-bound filter runs in Go, SQL-free),
  `test-before-implementation`.
- Related docs:
  - `docs/api-contract.md` (gains a `brag wrapped` section — SPEC-051).
  - `projects/PROJ-004-story-surface/stages/STAGE-013-polish-and-v0-4-0-cut.md`.
