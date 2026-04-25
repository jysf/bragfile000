---
# Maps to ContextCore insight.* semantic conventions.

insight:
  id: DEC-014
  type: decision
  confidence: 0.80
  audience:
    - developer
    - agent

agent:
  id: claude-opus-4-7
  session_id: null

project:
  id: PROJ-001
repo:
  id: bragfile

created_at: 2026-04-25
supersedes: null
superseded_by: null

tags:
  - cli
  - io-contract
  - json
  - markdown
  - aggregation
  - human-consumer
  - ai-consumer
---

# DEC-014: Rule-based output shape — single-object envelope (JSON) + provenance/summary block (markdown) for `summary`, `review`, `stats`

## Decision

The three rule-based STAGE-004 commands — `brag summary`,
`brag review`, `brag stats` — emit a shared output shape. Six
choices are locked:

1. **JSON output is a single-object envelope, not an array.**
   This intentionally diverges from DEC-011 (naked JSON array for
   `brag list --format json` and `brag export --format json`)
   because aggregations carry per-document metadata
   (`generated_at`, `scope`, `filters`) that does not fit a
   per-entry row. Where DEC-011's consumers want
   `jq '.[] | .field'`, DEC-014's consumers want `jq .scope`,
   `jq .counts_by_type`, `jq .highlights[0].entries`.

2. **Top-level keys are flat, not nested under a `payload`
   wrapper.** Provenance keys (`generated_at`, `scope`,
   `filters`) sit at the same level as per-spec payload keys
   (`counts_by_type`, `counts_by_project`, `highlights` for
   summary; `entries_grouped`, `reflection_questions` for
   review; the six metric keys for stats). Best `jq`
   ergonomics — consumers don't write `.payload.X`. Per-spec
   payload keys are documented in each consuming spec and stay
   stable across that spec's lifetime.

3. **Markdown output uses a provenance-then-summary-then-payload
   convention.** `# <Doc Title>` level-1 heading, then a
   provenance block with `Generated: <RFC3339>`, `Scope: <range
   value>`, `Filters: <echoed flag string or "(none)">`, then a
   `## Summary` block (where applicable — summary and stats
   render it; review may omit it), reusing DEC-013's
   `**By type**` / `**By project**` bulleted-counts convention,
   then per-spec body. Inherits DEC-013's count-ordering rule:
   DESC by count, alphabetical-ASC tiebreak,
   `(no project)` last in the by-project list regardless of
   count.

4. **Empty-state values follow a fixed-shape rule.** Numeric
   fields render as `0`. Arrays render as `[]` (non-nil; never
   `null` in JSON). Date fields render as `null` in JSON and
   `-` in markdown. Object fields render as `{}` (non-nil;
   `filters` is always an object, never `null`, even when no
   filter flags were echoed). Provenance always renders;
   `## Summary` and per-spec payload sections are OMITTED
   from the markdown when the entry set is empty (the document
   ends after the `Filters:` line). Mirrors DEC-013's empty-
   state precedent.

5. **JSON is pretty-printed with 2-space indentation.** Matches
   DEC-011 and DEC-013. No `--compact` flag in MVP for any of
   the three rule-based commands. Inherited backlog entry
   covers all JSON-emitting commands.

6. **`--range` (where present) is rolling-window, not
   calendar.** `brag summary --range week` = entries with
   `created_at >= time.Now().UTC() - 7 days`; `--range month`
   = `now - 30 days`. NOT calendar-week (Mon–Sun) or calendar-
   month (1st–end). `brag stats` operates on the lifetime
   corpus and has no `--range`. `brag review` uses `--week` /
   `--month` named flags with the same rolling semantics. The
   `scope` field in the JSON envelope echoes the literal range
   value (`"week"` / `"month"` / `"lifetime"`) so consumers
   don't need to derive it.

JSON envelope skeleton (per-spec keys vary; provenance is
shared):

```json
{
  "generated_at": "2026-04-25T12:00:00Z",
  "scope": "week",
  "filters": {},
  "<per_spec_payload_keys>": ...
}
```

Markdown envelope skeleton:

```
# <Doc Title>

Generated: 2026-04-25T12:00:00Z
Scope: week
Filters: (none)

## Summary

**By type**
- shipped: 3
- learned: 1

**By project**
- alpha: 2
- beta: 1
- (no project): 1

<per-spec payload sections>
```

Counts-as-objects asymmetry (locked, not bug): the markdown
side renders counts DESC by count (human-scannable). The JSON
side renders counts as `map[string]int` objects, which Go's
`encoding/json` sorts alphabetical-ASC by key when marshaling
(deterministic across runs since Go 1.12). That means
markdown and JSON show counts in different orders. JSON
consumers re-sort if they care. This trade-off is captured
explicitly so future readers don't see it as drift.

## Context

STAGE-004 ships three rule-based digests that need to agree on
provenance and on the structural skeleton consumers can rely
on across the three commands:

- `brag summary --range week|month` (SPEC-018) — counts by
  type/project + grouped highlights.
- `brag review --week | --month` (SPEC-019) — entries grouped
  by project + three reflection questions appended.
- `brag stats` (SPEC-020) — six lifetime aggregations.

Without one shape lock, each consumer drifts — provenance
fields differ between commands, `jq` recipes don't transfer,
and AI consumers downstream have to case-analyze per command.
The shape is small but load-bearing enough to warrant a DEC
rather than scattering choices across three spec files.

The structural mirror is SPEC-014 / DEC-011: same "one shape,
one helper, one byte-identical golden" pattern. The shape
itself differs from DEC-011 because DEC-011's consumers
emit lists (`list`, `export`), and DEC-014's consumers emit
single-document digests with metadata that doesn't belong on
each row.

Six choices needed locking; each is stated above.

## Alternatives Considered

- **Option A: Reuse DEC-011's naked-array shape verbatim.**
  - What it is: Make summary/review/stats emit an array of
    something — perhaps a wrapper object inside an array of
    one element, or a flattened representation.
  - Why rejected: Doesn't fit aggregations. A digest carries
    per-document metadata (`generated_at`, `scope`,
    `filters`) that has no natural per-row home. Wrapping
    in `[{...}]` is dishonest about the shape (one-element
    arrays signal "this is a list" to consumers, but it
    isn't). Forcing aggregations into DEC-011's naked-array
    form would make `jq .` recipes look the same shape across
    list/export/summary, but at the cost of misleading every
    consumer about whether they're holding a list. The two
    shapes serve two different jobs and earning their right
    to differ is the simpler outcome.

- **Option B: Per-spec independent shapes.**
  - What it is: Let each of summary/review/stats define its
    own envelope without shared keys. Summary picks
    `{generated_at, scope, ...}`; review picks
    `{when, period, ...}`; stats picks `{computed_at, ...}`.
  - Why rejected: Defeats the point of a stage-level DEC.
    The whole reason these three commands ship as one stage
    is to give the user-and-agent a consistent paste-into-AI
    surface. Three different envelopes mean three different
    `jq` recipes for the same metadata; consumers writing
    AI prompts to summarize across the three would case-
    analyze each command's shape. The cost of one shared
    DEC is small; the payoff (consistent surface) is the
    stage's value thesis.

- **Option C: Nested `payload` key.**
  - What it is: Provenance keys at top level; everything per-
    spec under a `payload` key. E.g. `{generated_at, scope,
    filters, payload: {counts_by_type, ...}}`.
  - Why rejected: Extra `jq` overhead. Consumers writing
    `.payload.counts_by_type` everywhere is uglier than
    `.counts_by_type`. The cleanliness payoff (clean
    metadata vs. content separation) is real but small —
    consumers can already filter on `generated_at`/`scope`
    by name without confusing them with payload keys, and
    schema documentation lives in this DEC + each consuming
    spec, not in the JSON shape. Top-level flat wins on
    ergonomics.

- **Option D: `range` as the JSON key instead of `scope`.**
  - What it is: Use `"range": "week"` in the envelope (echoing
    the CLI flag name) instead of `"scope": "week"`.
  - Why rejected: `range` is a CLI flag for summary AND
    review (via `--week`/`--month`), but stats has no range
    — its scope is always "lifetime." The neutral term
    `scope` covers all three commands' concept of "what
    period does this digest cover." Future consumers asking
    `jq .scope` get a stable answer regardless of which of
    the three commands produced the JSON. `range` would
    require either (a) `stats` setting `"range": "lifetime"`
    (semantic stretch — nothing was ranged) or (b) `stats`
    omitting the key (different shape across commands —
    defeats the point).

- **Option E: `filters` as `null` when absent.**
  - What it is: When no filter flags were passed, render
    `"filters": null` instead of `"filters": {}`.
  - Why rejected: Symmetry across the populated and empty
    cases is friendlier to consumers — they always know
    `filters` is an object, no null-check needed. JSON
    purist style would split "absent" from "empty" via null
    vs `{}`, but in this domain the distinction is
    meaningless: there is no third state where `filters`
    is "not applicable." Always-an-object also matches
    DEC-011's empty-array discipline (`[]` not `null`).

- **Option F: Counts as arrays (`[{"type": "shipped",
  "count": 3}, ...]`) instead of objects.**
  - What it is: Render `counts_by_type` as a JSON array of
    `{type, count}` objects, preserving the markdown side's
    DESC-by-count ordering.
  - Why rejected: Loses the natural lookup ergonomics
    (`jq '.counts_by_type.shipped'` becomes
    `jq '.counts_by_type[] | select(.type == "shipped") | .count'`).
    The map-based shape sorts alphabetical-ASC by key
    (Go's json encoder, deterministic since 1.12), which
    is different from the markdown ordering, but JSON
    consumers re-sort if they care. The ergonomic win on
    lookups is worth the asymmetry. Documented as a known
    asymmetry in choice (3) and locked.

- **Option G (chosen): All six choices above, applied
  together.**
  - What it is: Single-object envelope + top-level flat keys
    + `scope`-not-`range` + `filters` always object +
    counts-as-objects-not-arrays + DEC-013-style markdown
    convention + indent=2 + rolling-window semantics.
  - Why selected: Each sub-choice has either a prior DEC it
    aligns with (DEC-013 for markdown convention, DEC-011
    for indent and empty-array discipline), a deliberately-
    deferred alternative captured in `backlog.md`, or a
    grounded ergonomic argument. The combined shape gives
    SPEC-018/019/020 a consistent paste-into-AI surface
    without re-litigating each choice.

## Consequences

- **Positive:** One shape, documented once. Consumers writing
  `jq .generated_at`, `jq .scope`, `jq .filters` get a stable
  answer across all three rule-based commands. SPEC-018's
  goldens (markdown + JSON) lock the envelope; SPEC-019 and
  SPEC-020 inherit and extend without re-litigating. The
  `internal/aggregate` package's data-layer types
  (`TypeCount`, `ProjectCount`, `EntryRef`,
  `ProjectHighlights`) compose cleanly with the JSON
  envelope and the markdown rendering.

- **Negative:** Two divergent shapes for `brag` JSON output
  now exist (DEC-011 array, DEC-014 envelope). Consumers
  writing generic `brag *` AI tooling have to know which
  command emits which shape. Mitigated by: (a) command
  groupings make this clean — `list`/`export` emit lists,
  `summary`/`review`/`stats` emit digests; (b) every
  command's shape is documented in `docs/api-contract.md`
  with a DEC cross-link; (c) a future ai-summary command
  in PROJ-002 can wrap the existing shapes without changing
  them.

- **Negative:** Adding a new rule-based command in PROJ-002
  (e.g., `brag ai-summary`) means another consumer of
  DEC-014. The shape's stability becomes a forward
  commitment. Acceptable: the envelope is thin enough that
  extension means adding new top-level keys, which is
  backward-compatible by design (consumers ignore unknown
  keys).

- **Neutral:** The counts-as-objects asymmetry between
  markdown (DESC by count) and JSON (alphabetical-ASC by
  key) is a documented quirk. Consumers can write a small
  `jq` re-sort recipe if they want DESC ordering. Renaming
  to arrays would lose lookup ergonomics; this is the right
  trade-off for the data shape.

- **Neutral:** The rolling-window semantics (vs calendar)
  trade-off is the simplest implementation that doesn't
  require day-of-week/month logic. Calendar-week support
  is a backlog candidate if a real workflow benefits.

## Validation

Right if:
- A user running `brag summary --range week --format json |
  jq .scope` gets `"week"`. A user running
  `brag review --week --format json | jq .scope` gets
  `"week"`. A user running `brag stats --format json |
  jq .scope` gets `"lifetime"`. The three commands' scope
  field is comparable.
- SPEC-018's load-bearing goldens
  (`TestToSummaryMarkdown_DEC014FullDocumentGolden` and
  `TestToSummaryJSON_DEC014ShapeGolden`) pass on the shared
  fixture. If either ever fails, this DEC has been violated
  and the question is "was the deviation deliberate?" before
  fixing the test.
- A consumer writing
  `if scope == "lifetime" { ... } else if scope == "week"
  { ... } else if scope == "month" { ... }` covers all valid
  scope values across the three rule-based commands.
- A consumer reading `filters` always sees a JSON object
  (possibly empty), never `null`. Consumers writing to it
  also serialize an object.

Revisit if:
- A real workflow needs calendar-week semantics (Mon–Sun
  or 1st-of-month). Then `--range` semantics split into a
  flag (rolling vs calendar) or a new spec adds calendar-
  range support.
- AI consumers concretely ask for the counts-as-arrays form
  (DESC-by-count preserved in JSON). Then promote
  Option F's backlog entry; the shape change is
  backward-incompatible and would need its own DEC.
- A future PROJ-002 command needs metadata that doesn't fit
  the envelope (e.g., per-section pagination tokens).
  Then either revise DEC-014 with a coordinated update
  across all consumers, or introduce a `meta` sub-object
  for that specific command (and document the extension
  pattern here).
- DEC-013's count-ordering rule changes (e.g., a future
  spec wants ASC-by-count for some workflow). DEC-014
  inherits, so it migrates in lockstep.

Confidence: 0.80. Each sub-choice is grounded; two softer
sub-choices keep the composite below 0.85:
- Choice (1) (envelope-not-array) is strong (0.90) — the
  alternative makes consumers lie about list-vs-digest
  shape.
- Choice (2) (top-level flat) is strong (0.90) — `jq`
  ergonomics vs nested-payload nesting is a clear win.
- Choice (3) (markdown convention reuse) is strong (0.90)
  — DEC-013 already proved out the convention.
- Choice (4) (empty-state values) is strong (0.85) —
  inherits DEC-011's `[]` discipline; the date-as-null and
  object-as-`{}` decisions are minor consistency calls.
- Choice (5) (indent=2) is strong (0.90) — matches
  DEC-011/013.
- Choice (6) (rolling window) is the softest (0.65) — a
  calendar-week reading of "week" is also defensible and
  some users will expect it. The rolling form is simpler
  to implement correctly across timezone boundaries and
  tested by SPEC-018 unit tests; if a user reports the
  expectation mismatch, this is the first revisit.

The composite (0.80) reflects choice 6 dragging the average;
the JSON envelope structural choices are stronger
individually.

## References

- Related specs:
  - SPEC-018 (emits this DEC; wires `brag summary --range
    week|month`; seeds `internal/aggregate`).
  - SPEC-019 (pending; consumes this DEC for `brag review
    --week|--month` envelope).
  - SPEC-020 (pending; consumes this DEC for `brag stats`
    envelope; extends `internal/aggregate` with `Streak`,
    `MostCommon`, `Span`).

- Related decisions:
  - DEC-011 (shared JSON output shape — naked array). DEC-014
    INTENTIONALLY DIVERGES because aggregations carry
    metadata that doesn't fit a per-entry row. Both DECs
    cross-reference each other so the divergence is visible
    to future contributors as a deliberate symmetry, not
    drift.
  - DEC-013 (markdown export shape). DEC-014 reuses DEC-013's
    provenance + summary block conventions for the markdown
    half: the `Generated:` line (renamed from `Exported:`
    because "Generated" reads better for digests), the
    `## Summary` heading, the `**By type**` / `**By
    project**` count blocks with DESC-by-count + alpha-ASC
    tiebreak + `(no project)` last.
  - DEC-004 (tags comma-joined TEXT). Indirectly relevant —
    summary's filters object can carry `"tag": "auth"`; no
    splitting at the I/O boundary.
  - DEC-006 (cobra framework) — `brag summary` is a new
    cobra subcommand following the same pattern as every
    other.
  - DEC-007 (required-flag validation in `RunE`) — `--range`
    on summary, `--week`/`--month` on review, `--format` on
    all three use `UserErrorf` for missing/empty/unknown
    values.

- Related constraints: `stdout-is-for-data-stderr-is-for-humans`
  (markdown/JSON bodies to stdout; error messages to stderr).

- Related backlog entries:
  - `--out <path>` for the three rule-based commands (deferred).
  - `--compact` / non-pretty JSON for the three rule-based
    commands (covered by the existing all-JSON-emitting-commands
    backlog entry).
  - Calendar-week / calendar-month semantics for `--range`
    (revisit if real workflow benefits).
  - Counts-as-arrays JSON form (Option F above) — capture
    if AI consumers ask for DESC-by-count preservation.

- Related docs:
  - `docs/api-contract.md` (`brag summary` section rewritten
    by SPEC-018; `brag review` and `brag stats` sections
    arrive in SPEC-019/020).
  - `docs/data-model.md` References list (gains a DEC-014
    row alongside DEC-011 / DEC-013).
