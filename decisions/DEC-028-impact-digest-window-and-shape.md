---
# Maps to ContextCore insight.* semantic conventions.

insight:
  id: DEC-028
  type: decision
  confidence: 0.78
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
  - impact
  - human-consumer
  - ai-consumer
---

# DEC-028: `brag impact` — calendar time windows, project=initiative grouping, and the impact-first envelope shape

## Decision

`brag impact` is the fourth consumer of DEC-014's rule-based digest
envelope. Six choices are locked, four of which extend (never relax)
DEC-014:

1. **Time windows are CALENDAR periods, not rolling.** This is a
   *deliberate divergence* from DEC-014 choice (6) (which locks
   *rolling* windows for `brag summary`/`review`). `--quarter` =
   the calendar quarter containing "now" (Jan–Mar / Apr–Jun /
   Jul–Sep / Oct–Dec), `[quarter-start 00:00:00 UTC, now]`.
   `--month` = the calendar month containing "now"
   (`[first-of-month 00:00:00 UTC, now]`). `--year` = the calendar
   year containing "now" (`[Jan-1 00:00:00 UTC, now]`). Each is the
   **current** (in-progress) period — not the previous complete one.
   `--since <date>` reuses DEC-008's `ParseSince` verbatim
   (`YYYY-MM-DD` or `Nd`/`Nw`/`Nm`), lower-bounded at the parsed
   instant, upper-bounded at now. The four window flags are
   **mutually exclusive**; passing two is a `UserError`. Exactly one
   is **required** — there is no default window (an unbounded impact
   digest over the whole corpus is `brag stats`' job, not this
   command's). The `scope` field in the envelope echoes the literal
   window token: `"quarter"` / `"month"` / `"year"` / `"since:<raw>"`.

2. **"Initiative" IS the `project` field.** The brief says
   "initiative/project"; this locks them as the same axis — an
   initiative is a `project` value, not a distinct tag namespace or a
   new column. Grouping reuses `aggregate.GroupEntriesByProject`
   verbatim (alpha-ASC by project, `(no project)` last, chrono-ASC +
   ID-tiebreak within group). No schema change, no new aggregate
   helper for grouping.

3. **The digest is IMPACT-FIRST: only entries with a non-empty
   `impact` field appear in the grouped body.** Entries in-window
   whose `impact` is empty are **excluded** from the grouped body
   (they have no impact statement to surface) but are **counted** and
   surfaced as a provenance tally so the number is never silently
   dropped. The markdown provenance block gains an
   `Entries: <shown>/<in-window> with impact` line; the JSON envelope
   gains flat `entries_in_window` and `entries_with_impact` integer
   keys. Grouping, counts, and the rendered body all operate on the
   **with-impact subset**.

4. **Per-entry rendering surfaces the impact text as the payload.**
   Markdown renders each shown entry as `- <id>: <title>` followed by
   an indented `  <impact>` line (the outcome/progression text is the
   point of the command, so it is rendered in full, not elided the way
   `summary`/`review` elide descriptions). JSON renders each entry as a
   4-key object `{id, title, project, impact}` inside the grouped
   payload — a deliberately **narrower** projection than DEC-011's
   9-key shape, because `impact` is the one field this digest exists to
   surface and the consumer is a narrative-synthesis pipe, not a
   round-trip.

5. **The envelope extends DEC-014 verbatim.** Single-object JSON,
   flat top-level keys, `scope`/`filters` provenance, markdown
   provenance-then-summary-then-payload, 2-space JSON indent,
   empty-state rules (numbers `0`, arrays `[]`, dates `null`/`-`,
   objects `{}`). Per-spec payload keys at top level:
   `entries_in_window`, `entries_with_impact`, `counts_by_project`
   (a `map[string]int` over the with-impact subset, matching
   SPEC-018's `counts_by_type` map shape and ordering asymmetry), and
   `impact_by_project` (an array of `{project, entries:[{id, title,
   project, impact}]}` groups). Empty-window and no-impact behavior
   follow DEC-014 part (4): provenance always renders (including the
   two tally counts, which are `0`); the `## Impact` body and the
   per-spec payload sections are OMITTED from markdown when there are
   zero with-impact entries.

6. **Filter flags compose with the window, same as summary.**
   `--project` / `--type` / `--tag` narrow the in-window set before the
   impact-first split, and echo into the `filters` object exactly as
   `brag summary` does (reusing the same per-flag echo shape). Passing
   `--project` and grouping-by-project is not contradictory — the
   filter narrows to one initiative, the grouping then renders that one
   group.

JSON envelope skeleton (per-spec keys; provenance shared with DEC-014):

```json
{
  "generated_at": "2026-07-06T12:00:00Z",
  "scope": "quarter",
  "filters": {},
  "entries_in_window": 7,
  "entries_with_impact": 4,
  "counts_by_project": { "alpha": 3, "beta": 1 },
  "impact_by_project": [
    {
      "project": "alpha",
      "entries": [
        { "id": 12, "title": "Ship auth", "project": "alpha",
          "impact": "cut p95 login latency 40%" }
      ]
    }
  ]
}
```

Markdown envelope skeleton:

```
# Bragfile Impact

Generated: 2026-07-06T12:00:00Z
Scope: quarter
Filters: (none)
Entries: 4/7 with impact

## Impact

### alpha

- 12: Ship auth
  cut p95 login latency 40%
```

## Context

`brag impact` is the deterministic data foundation the v0.4.0 story
surface (STAGE-012's `brag story --audience`) reads. It is a rule-based,
time-windowed, initiative-grouped digest that surfaces entries' `impact`
fields — a sibling of `brag summary`/`review`/`stats`, local-first, no
network, no model.

Four choices needed a decision that the three existing DEC-014 consumers
did not settle:

- **Window semantics.** `summary`/`review` use rolling windows
  (DEC-014 choice 6). But the story surface's altitudes (skip-level,
  exec) map to *reporting periods* humans and orgs actually name:
  "this quarter," "this month," "this year." A rolling-30-day
  "month" would silently disagree with the calendar month a manager
  means when they ask "what did you ship this month?" The digest that
  feeds a manager/exec narrative should align to the boundary the
  audience thinks in. Hence calendar, not rolling — and this needs to
  be an explicit, tested divergence from DEC-014 so a future reader
  doesn't file it as drift.
- **Initiative vs project.** The brief writes "initiative/project."
  Left unresolved, build would guess. Collapsing them to the existing
  `project` axis reuses `GroupEntriesByProject`, requires no schema
  change, and matches how every other digest already groups.
- **What "surfaces the impact field" means.** An impact digest that
  padded its body with impact-less entries would bury the signal it
  exists to surface. Impact-first (with a visible tally of what was
  excluded) keeps the body pure without silently hiding rows.
- **Per-entry projection.** DEC-011's 9-key shape carries fields
  (description, tags, timestamps) that the narrative pipe does not need
  and that would bloat the bundle. A 4-key `{id, title, project,
  impact}` projection is the minimum that lets a downstream synthesizer
  attribute an impact statement to an entry and its initiative.

DEC-014 is extended, not relitigated: the envelope, empty-state rules,
markdown conventions, and JSON indent are inherited. Only the window
semantics diverge, and that divergence is scoped to this command's new
flags (`--quarter`/`--month`/`--year`/`--since`), which DEC-014's
`summary`/`review` never had.

## Alternatives Considered

- **Option A: Rolling windows (inherit DEC-014 choice 6 verbatim).**
  - What it is: `--month` = `now - 30 days`, `--quarter` = `now - 90
    days`, `--year` = `now - 365 days`, matching `brag summary`.
  - Why rejected: The story surface's whole point is audience-shaped
    reporting, and the audiences (manager/skip/exec) think in calendar
    periods. "What did you ship this quarter?" means the calendar
    quarter, not a trailing 90 days that straddles two quarters. A
    rolling reading would make the exec digest subtly disagree with the
    org's actual quarter boundary. Rolling is simpler to implement, but
    correctness-to-audience-intent wins here. (Rolling stays right for
    `summary`/`review`, whose framing is "recent activity," not
    "reporting period" — so DEC-014 choice 6 is untouched for them.)

- **Option B: Previous-complete period instead of current-in-progress.**
  - What it is: `--quarter` = the last *finished* quarter; `--month` =
    last finished month.
  - Why rejected: The dominant workflow is "help me write my update
    *now*, mid-period." An engineer preparing a weekly/1:1/quarterly
    note wants the period they're currently in, up to today. "Previous
    complete quarter" is a real but secondary need; deferred to a
    possible future `--previous` modifier rather than made the default.

- **Option C: Initiative as a distinct tag namespace (e.g. `init:*`
  reserved tags).**
  - What it is: Treat "initiative" as a higher grouping than project —
    a reserved tag prefix that spans multiple projects.
  - Why rejected: No such concept exists in the schema, the corpus has
    zero `init:*` history, and introducing one is a data-model change
    that belongs to a dedicated spec if ever justified. The brief's
    "initiative/project" reads as "the project axis, called an
    initiative in narrative framing," not "a new grouping level." Reuse
    beats invention here.

- **Option D: Include impact-less entries in the body (grey/blank
  impact line).**
  - What it is: Render every in-window entry; show `(no impact
    recorded)` for the empty ones.
  - Why rejected: Buries the signal. The command exists to surface
    impact statements; padding the body with rows that have none
    dilutes exactly the thing a downstream narrative pipe is selecting
    on. The tally (`4/7 with impact`) preserves honesty about what was
    excluded without polluting the body. A user who wants all in-window
    entries has `brag list --since` / `brag summary` already.

- **Option E: Reuse DEC-011's full 9-key per-entry JSON shape (as
  `brag review` does).**
  - What it is: Serialize each shown entry with the complete
    description/tags/timestamps shape via `toEntryRecord`.
  - Why rejected: The narrative pipe needs id + title + project +
    impact to attribute an impact statement to an initiative; the other
    five keys are dead weight in the bundle STAGE-012 emits to an LLM.
    A narrow projection keeps the bundle lean and signals intent
    (this digest is *about* impact). `review` keeps the 9-key shape
    because its job is a faithful entry dump for reflection; `impact`'s
    job is selection, so the projection differs by design — the same
    kind of deliberate per-consumer payload divergence DEC-014 already
    tolerates (SPEC-018 maps vs SPEC-020 arrays).

- **Option F (chosen): calendar windows + project=initiative +
  impact-first body + 4-key projection, all over the DEC-014 envelope.**
  - What it is: The six choices above applied together.
  - Why selected: Each divergence from the existing consumers is
    grounded in the story-surface workflow (calendar for reporting
    intent; impact-first for signal purity; narrow projection for a
    lean synthesis bundle), and everything that *can* be inherited from
    DEC-014 is (envelope, empty-state, markdown convention, indent).
    `brag impact` reads as a natural fourth sibling, not a bespoke path.

## Consequences

- **Positive:** The story surface (STAGE-012) reads one more DEC-014
  digest with a stable envelope; `jq .scope`, `.filters`,
  `.entries_with_impact`, `.impact_by_project[].entries[].impact`
  transfer cleanly. Grouping and provenance reuse existing
  `internal/aggregate` + `internal/export` machinery. No schema change.
- **Negative:** A second window semantics now exists in the codebase
  (rolling for `summary`/`review`, calendar for `impact`). A consumer
  reasoning about "what does `--month` mean" must know which command
  they're in. Mitigated by: the `scope` echo is command-specific and
  documented; the divergence is explicit here and enforced by a
  dedicated failing test (calendar-boundary vs rolling).
- **Negative:** Calendar-quarter/month/year math must be correct across
  month lengths and year boundaries (no rolling-day shortcut).
  Mitigated by computing period starts with `time.Date(...)` calendar
  constructors (never day subtraction), unit-tested at boundaries.
- **Neutral:** The current-in-progress period choice means a digest run
  on the 2nd of a month returns a near-empty window; that is correct
  ("this month so far"). A `--previous` modifier is a clean future
  addition if the complete-period workflow materializes.
- **Neutral:** The impact-first exclusion means `counts_by_project`
  counts the *with-impact* subset, not the raw in-window set. The
  `entries_in_window` tally preserves the raw count so no information
  is lost.

## Validation

Right if:
- `brag impact --quarter --format json | jq .scope` returns
  `"quarter"`; `--since 2026-01-01` returns `"since:2026-01-01"`.
- A digest run mid-quarter includes an entry created on the 1st day of
  the calendar quarter and excludes one created the day before it —
  i.e. the boundary is the calendar quarter start, not `now - 90d`.
  (This is the load-bearing calendar-vs-rolling assertion.)
- An in-window entry with empty `impact` increments `entries_in_window`
  but not `entries_with_impact`, and does not appear in
  `impact_by_project`.
- A shown entry's `impact` text renders in full in both markdown (the
  indented line) and JSON (the `impact` key) — never elided.
- Passing two window flags, or none, is a `UserError` on stderr with a
  non-zero exit; stdout stays empty.

Revisit if:
- A real workflow needs the previous *complete* period (add
  `--previous`, don't change the default).
- Initiative genuinely needs to span multiple projects (then it becomes
  a data-model spec, not a grouping tweak here).
- The narrative pipe (STAGE-012) needs a field the 4-key projection
  omits (widen the projection deliberately, with a test).

Confidence: 0.78. The envelope-inheritance choices (2,3,4,5,6) are
strong (0.85–0.9) — they reuse proven DEC-014/aggregate machinery. The
composite is dragged by choice (1): calendar-vs-rolling is a genuine
judgment call. Rolling (matching DEC-014) is defensible and simpler;
calendar is chosen because the story surface's audiences think in
reporting periods, but a user could reasonably expect either. This is
the first revisit trigger and is recorded as an open question in
`/guidance/questions.yaml` per §14 (confidence < 0.8 on the window
semantics sub-choice specifically). The current-vs-previous period
sub-choice (choice 1, second half) is the second soft spot.

## References

- Related specs:
  - SPEC-048 (emits this DEC; wires `brag impact
    --quarter|--month|--year|--since`; adds `internal/export/impact.go`;
    reuses `aggregate.GroupEntriesByProject`).
  - SPEC-018 (shipped; emitted DEC-014, seeded `internal/aggregate`,
    `echoFiltersForSummary` filter-echo precedent SPEC-048 mirrors).
  - SPEC-019 (shipped; second DEC-014 consumer; `GroupEntriesByProject`
    which SPEC-048 reuses).
  - SPEC-020 (shipped; third DEC-014 consumer; per-spec payload-shape
    divergence precedent — maps vs arrays — SPEC-048 follows for its
    narrow 4-key projection).
- Related decisions:
  - DEC-014 (rule-based output envelope) — EXTENDED verbatim except for
    the window semantics, which SPEC-048 deliberately diverges from
    (calendar, not rolling) for its own new flags. DEC-014 choice (6)
    is untouched for `summary`/`review`.
  - DEC-008 (`--since` date format) — reused verbatim via
    `cli.ParseSince` for the `--since` window.
  - DEC-011 (JSON per-entry shape) — DEC-028 uses a deliberately
    narrower 4-key projection, not the 9-key shape, for the narrative
    pipe.
  - DEC-007 (required-flag validation in RunE) — the mutually-exclusive
    window flags and `--format` use `UserErrorf`.
  - DEC-017 (entries↔project relationship) — grouping is on the
    `project` field this DEC establishes as the initiative axis.
- Related constraints: `stdout-is-for-data-stderr-is-for-humans`,
  `no-sql-in-cli-layer` (window math + impact-first split live in the
  CLI/export layers; the read path is `Store.List(ListFilter{Since})`),
  `test-before-implementation`.
- Related docs:
  - `docs/api-contract.md` (gains a `brag impact` section — SPEC-048).
  - `projects/PROJ-004-story-surface/stages/STAGE-011-impact-digest-foundation.md`.
