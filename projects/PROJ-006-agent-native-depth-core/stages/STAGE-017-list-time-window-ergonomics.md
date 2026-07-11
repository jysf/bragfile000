---
# Maps to ContextCore epic-level conventions.
# A Stage is a coherent chunk of work within a Project.
# It has a spec backlog and ships as a unit when the backlog is done.

stage:
  id: STAGE-017
  status: active
  priority: medium
  target_complete: null

project:
  id: PROJ-006
repo:
  id: bragfile

created_at: 2026-07-10
shipped_at: null
---

# STAGE-017: list time-window ergonomics

## What This Stage Is

A small, ship-early read-surface win that opens PROJ-006: make "what did I do
today / yesterday / in this window" a one-liner. `brag list` today has only a
lower bound (`--since`); this stage exposes the upper bound (`ListFilter.Until`,
already shipped in v0.5.0/SPEC-056 but never surfaced on the CLI) as
`--until`, and adds `--today`/`--yesterday` convenience presets — settling the
day-boundary (local vs UTC) question along the way. Not agent-native depth; a
quick ergonomics improvement that delivers value before the deeper stages.

## Why Now

A real usage gap: there's no clean way to see just today's or just yesterday's
entries. `--since` covers "today" roughly (UTC midnight), but "yesterday only"
needs a `jq` filter because `list` has no upper bound, and `--since`'s bare-date
UTC-midnight anchoring skews for a non-UTC user (PDT: "today" starts at 5pm the
day before). The fix is cheap — the storage upper bound already exists — and
high-value for daily standups/retros. Doing it first also lets PROJ-006 deliver
a shippable win while its deeper stages (memory, provenance, completeness,
benchmark) are still being framed.

## Success Criteria

- `brag list --until <date>` filters `created_at < until` (exclusive), reusing
  `storage.ListFilter.Until`; composes with `--since` for a bounded window.
- `brag list --today` and `--yesterday` return exactly that day's entries with a
  human-intuitive day boundary; mutually exclusive with each other and with
  `--since`/`--until` (UserError on conflict).
- The day-boundary semantics (local vs UTC) are decided and documented (DEC-039).
- No regressions to existing `list` behavior/tests; full gate set green.

## Scope

### In scope
- SPEC-068: `brag list --until` + `--today`/`--yesterday`; DEC-039 (day-boundary
  semantics); fix the `since.go` wall-clock impurity (use the injectable clock
  seam so the presets are testable) — the audit's L4.

### Explicitly out of scope
- The deeper agent-native pillars (memory / signed provenance / capture
  completeness / benchmark) — later PROJ-006 stages, framed separately.
- Exposing `--until`/presets on the calendar-window commands (impact/story/etc.)
  — they already own their windows.

## Spec Backlog

Format: `- [status] SPEC-ID (cycle) — one-line summary`

- [ ] SPEC-068 (design) — `brag list --until <date>` + `--today`/`--yesterday`
      presets (+ DEC-039 day-boundary semantics; fix since.go clock impurity).

**Count:** 0 shipped / 0 active / 1 pending

## Design Notes

- The upper bound already exists (`storage.ListFilter.Until`, DEC-035) — this is
  a CLI wire-up, not new storage. `--until` mirrors `--since` (bare date = a
  boundary; exclusive `created_at < until`).
- **The one real fork (DEC-039): day boundaries for `--today`/`--yesterday` —
  LOCAL vs UTC.** Lean: presets use LOCAL day (what a human means by "today," and
  consistent with DEC-022's local-day streak), while bare-date `--since`/`--until`
  stay UTC (backward-compat). Document the distinction; weigh the all-UTC
  alternative in the DEC.
- Compute "now"/day boundaries through the injectable clock seam (not
  `time.Now()` directly) so presets are deterministically testable — folds in the
  audit L4 (`since.go` wall-clock impurity).
- Mutual exclusion: `--today`/`--yesterday` are presets (they set the window), so
  they conflict with each other and with `--since`/`--until`; `--since`+`--until`
  compose.

## Dependencies

### Depends on
- PROJ-005 / SPEC-056 (shipped): `storage.ListFilter.Until` is the storage
  primitive this exposes on the CLI.

### Enables
- Faster daily standups/retros; a cleaner window vocabulary that later agent-read
  surfaces can reuse.

## Stage-Level Reflection

*Filled in when status moves to shipped.*

- **Did we deliver the outcome in "What This Stage Is"?** <yes/no + notes>
- **How many specs did it actually take?** <number vs. plan>
- **What changed between starting and shipping?** <one sentence>
- **Lessons that should update AGENTS.md, templates, or constraints?**
  - <one-line updates>
- **Should any spec-level reflections be promoted to stage-level lessons?**
  - <one-line items>
