---
# Maps to ContextCore insight.* semantic conventions.

insight:
  id: DEC-035                         # stable, never reused
  type: decision
  confidence: 0.9                     # honest: this is a behavior-preserving
                                      # promotion of an already-shipped, tested
                                      # Go filter into the storage layer, exactly
                                      # symmetric with the existing Since block,
                                      # firing a revisit trigger two prior DECs
                                      # named explicitly. Nothing novel is
                                      # decided; the only residual risk is the
                                      # mechanical refactor of four callers,
                                      # each covered by an existing green test.
  audience:
    - developer
    - agent

agent:
  id: claude-opus-4-8
  session_id: null

# Decisions are repo-level, but it's useful to track which project
# caused them to be emitted.
project:
  id: PROJ-005
repo:
  id: bragfile

created_at: 2026-07-10
supersedes: null
superseded_by: null

tags:
  - storage
  - query
  - time
  - window
  - refactor
  - developer
  - ai-consumer
---

# DEC-035: Promote the calendar-window upper bound into `storage.ListFilter.Until`

## Decision

Add an `Until time.Time` field to `storage.ListFilter` and apply it in
`Store.List` as an exclusive `e.created_at < ?` WHERE clause, guarded by
`!f.Until.IsZero()`, with the bound formatted `f.Until.UTC().Format(time.RFC3339)`
— exactly symmetric with the existing `Since` (inclusive `>=`) lower bound. The
four CLI commands that currently filter the calendar-window upper edge in Go
(`impact`, `story`, `wrapped`, `coverage`) drop their post-`List` filter loop and
set `filter.Until` instead. This moves the last piece of window filtering into
`internal/storage`, honoring `no-sql-in-cli-layer`, and eliminates four
copies of the same one-liner.

Four points are locked:

1. **`Until` is an exclusive upper bound (`created_at < ?`), `Since` stays
   inclusive (`created_at >= ?`).** Together they express the half-open
   `[Since, Until)` calendar window every bounded consumer already assumes: the
   completed period ends exactly where the next one begins, so an entry created
   at the boundary belongs to the NEXT period and must be excluded. This matches
   the Go filter it replaces verbatim — `e.CreatedAt.Before(end)` is `< end`.

2. **The `!f.Until.IsZero()` guard makes a zero `Until` a no-op.** A zero
   `time.Time` contributes no clause, exactly as `Since`'s `!IsZero()` guard
   does. This is load-bearing: `impact`/`story`/`coverage` pass `windowCutoff`'s
   `end`, which is the ZERO sentinel on the current-period (non-`--previous`)
   path. A zero `Until` there yields byte-for-byte the current `[cutoff, now]`
   behavior — no new upper filter — so no golden and no current-period test can
   move.

3. **RFC3339 UTC, written from Go.** `f.Until.UTC().Format(time.RFC3339)`,
   identical to the `Since` block, honoring `timestamps-in-utc-rfc3339`. The
   stored `created_at` strings are RFC3339 UTC, so the comparison is a
   lexicographic string comparison that agrees with chronological order.

4. **The four consumers wire their existing upper bound into the new field, no
   new bound is computed.** `impact`/`story`/`coverage` set `Until: end` (from
   `windowCutoff`); `wrapped` sets `Until: nextBoundary` (from
   `parseWrappedPeriod`, always non-zero — a named period is always bounded).
   Same field, two upstream helpers, mirroring how `--previous` already wires
   one modifier through two window mechanisms (DEC-032 choice 3).

## Context

The bounded-window upper edge has been filtered in Go, post-`Store.List`, since
`brag wrapped` shipped (DEC-030). DEC-030 explicitly recorded the tradeoff and a
revisit trigger:

> The upper-bound filter runs in Go (the CLI layer) because `ListFilter` has no
> `Until`. […] adding an `Until` to `ListFilter` was considered heavier than
> warranted for one caller (revisit if a second bounded-window consumer appears).
> […] Revisit if: A second bounded-window consumer appears (then promote the Go
> upper-bound filter into a `ListFilter.Until` field with a small DEC).

DEC-032 (`--previous`) added the second and third consumers (`impact`, `story`)
and restated the same deferral:

> the upper-bound filter now runs in Go in a SECOND place […] promoting it to a
> `ListFilter.Until` field remains a future call (DEC-030's own revisit note) if
> a third consumer appears — two Go-side filters is still below that bar.
> […] Revisit if: A third bounded-window consumer appears (then promote the Go
> upper-bound filter to `ListFilter.Until` with a storage-layer DEC — DEC-030's
> deferred trigger).

SPEC-045 (`brag coverage`) then added a FOURTH copy. The count is now
unambiguous: four consumers (`impact`, `story`, `wrapped`, `coverage`) each carry
the same `created_at < <bound>` Go loop. That is two past DEC-030's
"second-consumer" bar and one past DEC-032's "third-consumer" bar — well clear of
rule-of-three. The trigger both prior DECs named has fired; this DEC is that
promotion. It is the direct storage-layer analogue of DEC-004 → DEC-015 (the
comma-joined tag column promoted to a normalized model once a second consumer
appeared): a deliberately-deferred shape change, executed when the named
threshold is crossed, not a reversal.

The promotion is behavior-preserving. The upper bound moves from a Go loop into
one SQL clause that is structurally identical to the `Since` clause beside it;
the `!IsZero()` guard preserves the current-period path exactly. The export
goldens are fixture-fed (the export renderers are called with hand-built entry
slices, never through `Store.List`), so they are structurally incapable of
changing; the guardrails for the CLI-level bounded-window behavior are the
existing `Previous*` / `Bounded` / `NoPreviousUnchanged` command tests, which
must stay green.

## Alternatives Considered

- **Option A: leave the four Go filters as they are.**
  - Why rejected: it is a fourth copy of the same filter in the CLI layer, past
    every rule-of-three bar the prior DECs set, and it keeps window filtering
    split across two layers (lower bound in SQL, upper bound in Go). Every new
    bounded consumer re-copies the loop. The prior DECs deferred precisely
    *until this point*, not indefinitely.

- **Option B: a `Between(since, until)` method or a dedicated bounded-window
  method on `Store`.**
  - Why rejected: `ListFilter` is the established, composable filter surface
    (`Tag`/`Project`/`Type`/`Since`/`Limit`/`Author` all AND-compose through the
    one `List`). A parallel method fractures that surface and cannot compose with
    the other filters the consumers already set. One more field on the existing
    struct is the smaller, symmetric change.

- **Option C: make `Until` inclusive (`created_at <= ?`).**
  - Why rejected: the consumers' bound is the NEXT period's start (an exclusive
    boundary — an entry at the boundary belongs to the next period). The Go
    filter being replaced is `CreatedAt.Before(end)`, i.e. strictly `<`. An
    inclusive `Until` would change behavior at the boundary and break the
    bounded-window semantics DEC-030/DEC-032 tested.

- **Option D (chosen): an exclusive `Until` field, `!IsZero()`-guarded,
  RFC3339-UTC, symmetric with `Since`; four consumers rewired.**
  - Why selected: it is the minimal, symmetric extension of the proven `Since`
    block; the guard preserves the current-period path byte-for-byte; it clears
    the CLI layer of the last window-filtering logic (`no-sql-in-cli-layer`); and
    it is exactly the promotion two prior DECs deferred to this threshold.

## Consequences

- **Positive:** window filtering is now entirely in `internal/storage`. Four
  copies of the Go upper-bound loop are deleted; a fifth bounded consumer (e.g.
  the sketched `brag spark`) sets one field instead of re-copying a loop.
  `no-sql-in-cli-layer` is honored for the upper bound, not just the lower.
  Resolves DEC-030's and DEC-032's shared revisit trigger.
- **Positive:** the storage layer now owns and tests the exclusive-upper-edge
  semantics directly (`TestList_FilterByUntil`), rather than the semantics living
  only in per-command CLI tests.
- **Neutral:** `ListFilter` grows from six fields to seven. All existing literals
  are keyed (`ListFilter{Since: ...}`), so the added field defaults to its zero
  value everywhere and no caller or test breaks. No test asserts the struct's
  field count (premise-audit grep in SPEC-056 confirms zero hits).
- **Neutral:** no schema change, no migration — `Until` is a query-time filter
  over an existing indexed-by-nothing column, exactly like `Since`.
- **Negative:** the four command refactors are mechanical but must be done
  together for the "no duplicated Go filter remains" outcome to hold. Mitigated:
  each command's bounded-window behavior is pinned by an existing green test that
  must stay green through the refactor (enumerated in SPEC-056's premise audit).

## Validation

Right if:
- `Store.List(ListFilter{Until: t})` returns only entries with
  `created_at < t`; an entry created exactly at `t` is EXCLUDED (exclusive edge).
- `Store.List(ListFilter{Since: lo, Until: hi})` returns exactly the entries in
  the half-open `[lo, hi)` window and composes with the other filters via AND.
- `Store.List(ListFilter{})` and `Store.List(ListFilter{Since: lo})` (zero
  `Until`) behave byte-for-byte as before — the `!IsZero()` guard emits no upper
  clause.
- After the four-consumer refactor, `grep -rn "CreatedAt.Before" internal/cli/`
  returns nothing, and every `Previous*` / `Bounded` / `NoPreviousUnchanged`
  command test and every export golden stays green/byte-identical.

Revisit if:
- A window needs an inclusive upper bound or an open-ended "before now but not
  after some future point" semantics (add a sibling field or an explicit
  inclusive variant with a test; do not overload `Until`).
- `created_at` range scans become a performance concern at a scale this
  personal-CLI does not have today (add an index with its own DEC).

Confidence: 0.9. The decision is a behavior-preserving promotion that fires a
revisit trigger two prior DECs named explicitly, mirrors the shipped `Since`
block exactly, and decides nothing genuinely novel — hence high confidence. It is
held below 1.0 only because it rewires four callers in one change; each is
covered by an existing green test, but the "all four, no copy left behind"
outcome is a refactor whose completeness is the one thing to verify. Above §14's
0.8 line, so no new open question is required.

## References

- Related specs:
  - SPEC-056 (emits this DEC; adds `ListFilter.Until`, refactors the four
    consumers, adds the storage tests).
  - SPEC-051 (shipped) — `brag wrapped`; DEC-030; the first Go upper-bound
    filter (`created_at < nextBoundary`) this promotion replaces.
  - SPEC-053 (shipped) — `--previous`; DEC-032; added the impact/story Go
    filters and named the `ListFilter.Until` promotion as the deferred
    third-consumer follow-up.
  - SPEC-045 (shipped) — `brag coverage`; the fourth Go filter.
  - SPEC-007 (shipped) — introduced `ListFilter` and the `Since` lower bound
    this field is symmetric with.
- Related decisions:
  - DEC-030 (`brag wrapped` bounded window) — recorded the Go-filter tradeoff
    and the "promote to `ListFilter.Until` when a second consumer appears"
    revisit trigger this DEC fires.
  - DEC-032 (`--previous`) — restated the same deferral for the third consumer;
    this DEC CLOSES that trigger.
  - DEC-028 (`brag impact` calendar windows) — `windowCutoff`'s `end` return
    (the bound `impact`/`story`/`coverage` feed into `Until`).
  - DEC-015 (tags normalization) — the canonical prior "deliberately-deferred
    shape change, executed when the named threshold is crossed" precedent
    (DEC-004 → DEC-015); this DEC is its storage-query analogue.
- Related constraints: `no-sql-in-cli-layer` (the whole point — the upper-bound
  filter moves into `internal/storage`), `timestamps-in-utc-rfc3339` (the `Until`
  bound formats as RFC3339 UTC, matching `Since`), `storage-tests-use-tempdir`
  (the new storage tests), `test-before-implementation`, `errors-wrap-with-context`.
- Related docs:
  - `docs/data-model.md` (the `ListFilter` shape, if it enumerates fields).
  - `projects/PROJ-005-agent-native-depth/stages/STAGE-016-v0-4-x-polish.md`.
