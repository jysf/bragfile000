---
# Maps to ContextCore insight.* semantic conventions.

insight:
  id: DEC-032
  type: decision
  confidence: 0.82
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

created_at: 2026-07-07
supersedes: null
superseded_by: null

tags:
  - cli
  - io-contract
  - time
  - digest
  - window
  - human-consumer
  - ai-consumer
---

# DEC-032: `--previous` — the last-completed-period window modifier, bounded and shared across the calendar-windowed commands

## Decision

`--previous` is a boolean window modifier that shifts a selected calendar
window from the **current** (in-progress) period to the **last-completed**
one, across `brag impact`, `brag story`, and `brag wrapped`. It is the clean,
additive modifier DEC-028 foresaw ("add `--previous`, don't change the
default") and that DEC-030 choice 2 deferred to. Seven choices are locked; the
DEC-014 envelope, DEC-028's calendar-period concept, and DEC-030's bounded
upper-edge pattern are all inherited, not relitigated.

1. **`--previous` = the last-completed calendar period, as a BOUNDED
   `[prev-start, prev-end)` window.** The current-period windows are
   `[start, now]` (open upper edge — every stored `created_at <= now`). A
   completed period has a real END, so `--previous` produces a bounded window
   whose exclusive upper bound `prev-end` is the **current** period's start
   (the boundary between the completed period and the in-progress one). This is
   the same bounded shape `wrapped` already ships (DEC-030 choice 3):
   `Store.List(ListFilter{Since: prevStart})` in SQL, `created_at < prevEnd`
   filtered in Go (`no-sql-in-cli-layer` intact — `ListFilter` has no `Until`).

2. **The previous-period start is the current-period start shifted back one
   period via `time.Date` + `AddDate` — never day subtraction.** `--quarter
   --previous`: `curStart.AddDate(0,-3,0)`. `--month --previous`:
   `curStart.AddDate(0,-1,0)`. `--year --previous`: `curStart.AddDate(-1,0,0)`.
   `AddDate` rolls year boundaries correctly (a January `--month --previous`
   lands in the prior December of the prior year; `--quarter --previous` in the
   prior Q4). This inherits DEC-028's calendar-math rule verbatim and reuses the
   exact current-period start computation as the anchor.

3. **Shared uniformly across `impact`/`story`/`wrapped`, but wired per each
   command's existing window mechanism.** `impact`/`story` route through the
   shared `windowCutoff`, which is EXTENDED (a `previous bool` param + a second
   `end time.Time` return: zero = open/up-to-now, non-zero = exclusive upper
   bound) — NOT forked into a parallel path. `wrapped` never routed through
   `windowCutoff` (it has `parseWrappedPeriod` for positional named periods), so
   its `--previous` is `now.Year()-1` fed into the same tested year path. One
   modifier, one meaning; two wiring points because the two window mechanisms
   already differ.

4. **On `impact`/`story`, `--previous` composes with the window SELECTION;
   `--previous` + `--since` is a `UserError`.** `--quarter --previous` shifts the
   explicit flag; on `story`, `--previous` alone shifts the audience profile's
   default window. `--since` is an explicit anchor, not a calendar period, so
   "the previous `--since`" is undefined — rejected as an incoherent combo,
   guarded at both the CLI flag level and the `windowCutoff` helper level.

5. **On `wrapped`, `--previous` is valid ONLY with no positional period; it
   shifts the annual default to the last-completed year.** `brag wrapped
   --previous` = `brag wrapped <last-year>` (e.g. 2025 in 2026), bounded.
   `brag wrapped 2026 --previous` / `brag wrapped 2026 Q3 --previous` is a
   `UserError`: the positional arg already names a bounded period, so shifting
   it is redundant and ambiguous (previous quarter? previous year?). The simplest
   coherent rule — `--previous` shifts the default only — keeps `--previous`
   meaning exactly one thing on `wrapped`.

6. **On `impact`, `--previous` with no window flag is a `UserError` (a modifier
   is not a window); on `story` it is valid (shifts the profile default).**
   `impact` requires exactly one window (DEC-028) — that existing rule fires,
   no new code. `story` has a profile default, so `--previous` shifts it.

7. **The `scope` provenance token echoes the shift: `<window>:previous` on
   impact/story; the concrete resolved year on wrapped.** `--quarter --previous`
   → `"quarter:previous"`; `--year --previous` → `"year:previous"`. `brag
   wrapped --previous` → `"2025"` (the concrete named year — wrapped resolves to
   a named period, so its existing scope already says which one; a `:previous`
   suffix would double-encode). This is a scope-STRING convention, not a
   DEC-014 envelope reshape.

## Context

`--previous` completes the calendar-window family the story surface has been
building since DEC-028. The dominant digest workflow is "help me write my
update *now*, mid-period" — hence current-period defaults everywhere (DEC-028
choice 1, DEC-030 choice 2). But the *secondary* workflow is real: retros and
reviews look back at the period that just finished ("last quarter,"
"last year"). DEC-028 named this modifier as the clean way to serve it without
changing any default, and DEC-030 explicitly deferred wrapped's
last-completed-year framing to it. This DEC delivers it.

Four choices needed a decision the prior DECs left open:

- **Bounded vs open upper edge.** The current windows are `[cutoff, now]`. A
  completed period has a real end, so `--previous` must be bounded — otherwise
  a "last quarter" digest run mid-this-quarter would leak this quarter's
  entries. DEC-030 already built the bounded pattern for `wrapped`; `--previous`
  reuses it. The upper bound is the *current* period start (the completed
  period ends exactly where the in-progress one begins).
- **Extend the shared helper vs fork a path.** All calendar-anchor math lives
  in `windowCutoff` (DEC-028's core). Forking a `previousWindowCutoff` would
  duplicate the quarter/month/year start math and risk drift. Extending
  `windowCutoff` (a `previous` param + an upper-bound return) keeps it in one
  place — and reuses the exact current-period start as the shift anchor.
- **How `--previous` composes on `wrapped`'s positional surface.** `wrapped`
  names periods positionally, not via flags. Shifting a named positional period
  ("the previous 2026 Q3") is ambiguous and competes with the positional
  surface itself. The coherent rule: `--previous` shifts the *default* only; an
  explicit period + `--previous` is a `UserError`. `--previous` then means
  exactly one thing on `wrapped`: the last-completed default year.
- **The provenance token.** impact/story echo a *relative* window, so
  `quarter:previous` tells a reader (or `jq .scope`) it was the previous period.
  wrapped resolves to a *named* concrete period, so it reuses its `"2025"`
  scope — the scope already says which year. Encoding `:previous` there would
  double-encode and lose the concrete-year provenance.

DEC-028 and DEC-030 are extended, not relitigated: the envelope, empty-state
rules, calendar math, project=initiative grouping, and the bounded upper-edge
pattern are all inherited. Only the previous-period shift, its flag surface,
its incoherent-combo rules, and the scope-token convention are new.

## Alternatives Considered

- **Option A: A `--last-quarter`/`--last-month`/`--last-year` flag trio.**
  - Why rejected: triples the flag surface and does not compose with `story`'s
    profile default (there is no "`--last-<profile-default>`" flag to add). One
    orthogonal `--previous` that shifts whatever window is selected is smaller
    and composes across all three commands — the definition of a modifier.

- **Option B: `--previous` changes the default period instead of adding a
  modifier.**
  - Why rejected: DEC-028 and DEC-030 both deliberately chose current-period
    defaults and named *this* modifier as the last-completed path. Flipping a
    default relitigates two shipped decisions and splits "which period" logic
    between a default and a flag — exactly what DEC-030 choice 2's rationale
    warned against.

- **Option C: `[prev-start, now]` window (inherit impact's open upper edge).**
  - Why rejected: a "last-completed-period" digest that leaked current-period
    entries would be wrong by definition. The bounded upper edge (= current
    period start) is the only correct, testable semantics.

- **Option D: `brag wrapped 2026 Q3 --previous` → 2026 Q2 (shift an explicit
  positional period).**
  - Why rejected: introduces a second period-shifting mechanism competing with
    wrapped's positional surface, and forces a quarter-vs-year "which axis
    shifts" rule the positional surface already handles cleanly (name it:
    `2026 Q2`). Keeping `--previous` to "shift the default only" makes it mean
    one unambiguous thing on `wrapped`.

- **Option E: A `previous: true` sibling key in the JSON envelope.**
  - Why rejected: a DEC-014 envelope reshape for one modifier is heavier than a
    scope-string convention. `<window>:previous` in the existing `scope` field
    carries the same signal with zero envelope change — the "extend, never
    reshape" discipline every DEC-014 consumer has held.

- **Option F: A separate `previousWindowCutoff` helper.**
  - Why rejected: duplicates the quarter/month/year start math and risks the two
    copies drifting. Extending the one `windowCutoff` (per the reuse-the-shared-
    infra discipline) keeps the calendar anchor in one function.

- **Option G (chosen): a bounded `[prev-start, prev-end)` modifier, extending
  `windowCutoff`, shared across the three commands with per-command wiring,
  incoherent combos as `UserError`s, and a `<window>:previous` scope
  convention.**
  - Why selected: each new choice is grounded in the retro/review workflow and
    the reuse-not-fork discipline; everything inheritable from DEC-028/DEC-030 is
    inherited. `--previous` reads as the natural completion of the calendar-window
    family, with exactly the bounded-upper-edge divergence DEC-030 already
    established, now generalized to a relative period.

## Consequences

- **Positive:** "last quarter"/"last month"/"last year" become first-class
  across the whole story surface with one modifier and no envelope change.
  `jq .scope` distinguishes `quarter` from `quarter:previous`. The bounded-window
  machinery DEC-030 built is reused verbatim; no schema change, no new dep.
  Resolves DEC-030's revisit trigger (the last-completed variant is delivered as
  the uniform `--previous` mechanism, not a default flip).
- **Negative:** `windowCutoff`'s signature changes (a `previous` param + an
  upper-bound return), touching its `impact`/`story` callers. Mitigated: the
  zero-`end` sentinel preserves the current-period path byte-for-byte, and a
  regression test (`TestImpactCmd_NoPreviousUnchanged`) pins it.
- **Negative:** the upper-bound filter now runs in Go in a SECOND place
  (impact/story, joining wrapped). Mitigated: it is the same `created_at <
  end` one-liner; promoting it to a `ListFilter.Until` field remains a future
  call (DEC-030's own revisit note) if a third consumer appears — two Go-side
  filters is still below that bar.
- **Neutral:** `wrapped --previous` is annual-only by design (LD4/LD6); a
  completed *quarter* wrapped is named positionally (`brag wrapped 2026 Q2`).
  This keeps `--previous` unambiguous on `wrapped`.
- **Neutral:** the scope-token asymmetry (impact/story `:previous`, wrapped
  concrete year) is deliberate — relative vs named provenance — and documented
  in the `scope` echo per command.

## Validation

Right if:
- `brag impact --quarter --previous --format json | jq .scope` returns
  `"quarter:previous"`; `--year --previous` returns `"year:previous"`.
- `brag impact --quarter --previous` run mid-Q3-2026 includes an entry created
  on 2026-04-01 and 2026-06-30 (prev Q2) and EXCLUDES one created 2026-07-01
  (current Q3) — the upper bound is the current period start, not now. (The
  load-bearing bounded-previous assertion.)
- With `now` in January, `--month --previous` lands in the prior December of
  the prior year (`[2025-12-01, 2026-01-01)`) — `AddDate` rolls the boundary.
- `brag story --audience me --previous` (no window flag) shifts the `me`
  profile default (year) to the last-completed year, bounded; `scope` echoes
  `year:previous`.
- `brag wrapped --previous` covers the last-completed calendar year (2025 in
  2026), byte-identical to `brag wrapped 2025`; `scope` is `"2025"`.
- `brag wrapped 2026 --previous`, `brag impact --since 2026-01-01 --previous`,
  and `brag impact --previous` (no window) are each `UserError`s on stderr with
  a non-zero exit and empty stdout.

Revisit if:
- A repeat-count (`--previous 2` = two periods back) workflow appears (extend
  the boolean to an int with a small DEC; keep the bounded machinery).
- A third bounded-window consumer appears (then promote the Go upper-bound
  filter to `ListFilter.Until` with a storage-layer DEC — DEC-030's deferred
  trigger).
- A previous-*quarter* `wrapped` is wanted (it is namable positionally today;
  a `--previous`-composing-with-a-quarter rule would be a new spec).

Confidence: 0.82. The bounded-window reuse (choice 1), the calendar-math shift
(choice 2), and the extend-not-fork wiring (choice 3) are strong (0.85–0.9) —
they reuse proven DEC-028/DEC-030 machinery and the exact current-period anchor.
The composite is dragged by choice 5 (wrapped's explicit-period-+-`--previous`
= `UserError`): a reasonable user might expect `brag wrapped 2026 Q3 --previous`
to mean Q2, and the "shift the default only" rule is a coherence-over-power
call, not an obvious one. It is above §14's 0.7 threshold (no new open question
required), and the related `wrapped-default-current-vs-last-completed` entry in
`/guidance/questions.yaml` is RESOLVED by this DEC (current default holds;
`--previous` is the last-completed path). The scope-token asymmetry (choice 7)
is the second soft spot but low-risk (provenance-string only, no data change).

## References

- Related specs:
  - SPEC-053 (emits this DEC; adds `--previous` to `impact`/`story`/`wrapped`;
    extends `windowCutoff`; reuses `parseWrappedPeriod` for wrapped).
  - SPEC-048 (shipped) — `brag impact`; DEC-028; `windowCutoff`/`selectedWindow`
    and the `nowFunc` seam `--previous` extends.
  - SPEC-049 (shipped) — `brag story`; `resolveWindow` (profile-default vs
    explicit-flag composition) `--previous` slots into.
  - SPEC-051 (shipped) — `brag wrapped`; DEC-030; the bounded
    `[start, nextBoundary)` window + Go upper-bound filter `--previous` reuses.
- Related decisions:
  - DEC-028 (`brag impact` calendar windows) — FORESAW this modifier ("add
    `--previous`, don't change the default"); its current-period start math is
    the shift anchor; its four-flag mutual exclusion + required-window rule hold.
  - DEC-030 (`brag wrapped` bounded window) — the `[start, next-boundary)`
    bounded upper-edge pattern reused verbatim; DEC-030 choice 2 (bare default
    = current year) HOLDS; this DEC delivers the deferred last-completed layer
    and CLOSES DEC-030's revisit trigger.
  - DEC-014 (rule-based envelope) — untouched; `--previous` changes only the
    in-window slice and the `scope` string.
  - DEC-008 (`--since` format) — `--since --previous` rejected as incoherent.
  - DEC-007 (required/invalid-flag validation via `UserErrorf`) — the incoherent
    combos.
- Related constraints: `stdout-is-for-data-stderr-is-for-humans`,
  `no-sql-in-cli-layer` (the upper-bound filter runs in Go, SQL-free),
  `test-before-implementation`.
- Related questions:
  - `wrapped-default-current-vs-last-completed` (`/guidance/questions.yaml`) —
    RESOLVED: current-calendar-year default holds; `--previous` is the uniform
    last-completed path this DEC establishes.
- Related docs:
  - `docs/api-contract.md` (gains `--previous` in the impact/story/wrapped flag
    tables — SPEC-053).
  - `projects/PROJ-004-story-surface/stages/STAGE-013-polish-and-v0-4-0-cut.md`.
