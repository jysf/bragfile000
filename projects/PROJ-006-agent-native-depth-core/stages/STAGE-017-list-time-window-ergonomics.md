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

A small, ship-early read-surface win that opens PROJ-006: make "what did I do on
a given day" a one-liner. `brag list` today has only a lower bound (`--since`)
and no clean way to scope to a single day. This stage adds one flag —
`brag list --day <YYYY-MM-DD|today|yesterday>` — that returns exactly that
calendar day's entries, using the upper bound (`storage.ListFilter.Until`,
already shipped in v0.5.0/SPEC-056 but never surfaced on the CLI). Not
agent-native depth; a quick ergonomics improvement that delivers value before
the deeper stages.

## Why Now

A real usage gap: there's no clean way to see just today's or just yesterday's
entries. `--since` covers "today" roughly (UTC midnight) but "yesterday only"
needs a `jq` filter (no upper bound), and `--since`'s bare-date UTC-midnight
anchoring skews for a non-UTC user (PDT: "today" starts at 5pm the day before).
A single `--day` flag with `today`/`yesterday` keywords solves the actual need —
day-scoped retrieval — in one concept, and the storage upper bound already
exists, so it's cheap. Doing it first lets PROJ-006 deliver a shippable win
while its deeper stages (memory, provenance, completeness, benchmark) are framed.

## Success Criteria

- `brag list --day <YYYY-MM-DD>` returns exactly that calendar day's entries
  (`[day-start, next-day-start)`), using `ListFilter.Since` + `Until`.
- `brag list --day today` and `--day yesterday` work as keyword values.
- The day boundary is human-intuitive (LOCAL day) and documented (DEC-039).
- `--day` is mutually exclusive with `--since` (UserError on conflict).
- No regressions to existing `list` behavior/tests; full gate set green.

## Scope

### In scope
- SPEC-068: `brag list --day <YYYY-MM-DD|today|yesterday>` (one flag, bounded to
  a single local calendar day via Since+Until); DEC-039 (day-boundary
  semantics); fix the `since.go` wall-clock impurity (use an injectable clock
  seam so `--day today`/`yesterday` are deterministically testable) — audit L4.

### Explicitly out of scope
- A general `--until` flag / arbitrary bounded ranges — YAGNI; `--day` covers the
  stated need. Add `--until` later only if arbitrary windows are actually wanted.
- The deeper agent-native pillars (memory / signed provenance / capture
  completeness / benchmark) — later PROJ-006 stages, framed separately.
- Exposing windows on the calendar-window commands (impact/story/etc.) — they
  already own their windows.

## Spec Backlog

Format: `- [status] SPEC-ID (cycle) — one-line summary`

- [x] SPEC-068 (shipped on 2026-07-11) — `brag list --day
      <YYYY-MM-DD|today|yesterday>`: scope the listing to one local calendar day
      via Since+Until (+ DEC-039 day semantics; fixes since.go clock impurity /
      audit L4).
- [~] SPEC-069 (verify) — v0.5.1 release cut: the stage's closing release action
      (CHANGELOG `[0.5.1]` + plugin version pin + §4 pre-flight; the irreversible
      tag/publish is orchestrator-driven after this PR merges).

**Count:** 1 shipped / 1 active / 0 pending

## Design Notes

- **One flag, not three.** `--day <value>` sets BOTH window bounds internally
  (`Since = day-start`, `Until = next-day-start`, reusing the shipped
  `ListFilter.Until` from DEC-035) — so it's a CLI wire-up, not new storage.
  Chosen over an `--until` primitive + `--today`/`--yesterday` presets: the user
  need is day-scoped, and one concept is simpler than three (YAGNI on arbitrary
  ranges).
- **`value` accepts:** a bare `YYYY-MM-DD`, or the keywords `today` /
  `yesterday`. Lock the exact set + error message for an unparseable value.
- **The one real fork (DEC-039): what is a "day" — LOCAL vs UTC.** Lean: LOCAL
  day (what a human means by "today"/"yesterday," and consistent with DEC-022's
  local-day streak). `--day 2026-07-05` → that date's LOCAL midnight-to-midnight
  window. Document that this differs from bare-date `--since` (UTC midnight,
  unchanged for backward-compat). Weigh the all-UTC alternative in the DEC (it's
  simpler but re-creates the very skew that motivated this).
- Compute day boundaries through an injectable clock seam (not `time.Now()`
  directly) so keyword resolution is deterministically testable — folds in the
  audit L4 (`since.go` wall-clock impurity).
- **Mutual exclusion:** `--day` sets the whole window, so it conflicts with
  `--since` → UserError naming the conflict. (No `--until` exists to conflict
  with.) The `--project`/`--type`/`--tag`/`--limit` filters still compose.

## Dependencies

### Depends on
- PROJ-005 / SPEC-056 (shipped): `storage.ListFilter.Until` is the storage
  primitive `--day`'s upper bound rides on.

### Enables
- Faster daily standups/retros; a day vocabulary later agent-read surfaces reuse.

## Stage-Level Reflection

*Filled in when status moves to shipped.*

- **Did we deliver the outcome in "What This Stage Is"?** <yes/no + notes>
- **How many specs did it actually take?** <number vs. plan>
- **What changed between starting and shipping?** <one sentence>
- **Lessons that should update AGENTS.md, templates, or constraints?**
  - <one-line updates>
- **Should any spec-level reflections be promoted to stage-level lessons?**
  - <one-line items>
