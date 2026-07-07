---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-053
  type: story
  cycle: design
  blocked: false
  priority: medium
  complexity: S

project:
  id: PROJ-004
  stage: STAGE-013
repo:
  id: bragfile

agents:
  architect: claude-opus-4-8
  implementer: claude-opus-4-8     # usually same Claude, different session
  created_at: 2026-07-07

references:
  decisions: [DEC-032, DEC-028, DEC-030, DEC-014, DEC-008, DEC-007]
  constraints:
    - stdout-is-for-data-stderr-is-for-humans
    - no-sql-in-cli-layer
    - no-cgo
    - test-before-implementation
    - errors-wrap-with-context
  related_specs: [SPEC-048, SPEC-049, SPEC-051, SPEC-018]
---

# SPEC-053: `--previous` — the last-completed-period window modifier

## Context

`--previous` shifts a calendar window from the **current** period to the
**last-completed** one — the clean, additive modifier DEC-028 foresaw ("add
`--previous`, don't change the default") and that `brag wrapped`'s design
(DEC-030 choice 2) deferred to when it defaulted to the *current* calendar
year. It makes "last quarter" / "last month" / "last year" a first-class ask
across the calendar-windowed story surface (`impact`, `story`, `wrapped`),
reusing the shared window infrastructure in `internal/cli/window.go` and the
bounded-window discipline `wrapped` already ships.

- Parent stage: `STAGE-013` (Polish + v0.4.0 cut). Small, cross-cutting.
- Project: `PROJ-004` (the story surface).
- **The core semantic shift:** the current-period windows are `[start, now]`
  (open upper edge — every stored `created_at <= now`). A *completed* period
  has a real END, so `--previous` produces a **bounded** `[prev-start,
  prev-end)` window — exactly the shape `wrapped` already ships (DEC-030 choice
  3), where `prev-end` is the **current** period's start (the boundary between
  the completed period and the in-progress one). The upper bound is filtered in
  Go, SQL-free (`no-sql-in-cli-layer`).
- **Calendar math only.** The previous-period start is the current-period start
  shifted back one period via `time.Date` + `AddDate` (`AddDate(0,-3,0)` for a
  quarter, `AddDate(0,-1,0)` for a month, `AddDate(-1,0,0)` for a year) — never
  day subtraction. `AddDate` rolls year boundaries correctly (a January
  `--month --previous` lands in the prior December of the prior year).
- **New decision this spec emits: DEC-032** — "`--previous` = last-completed
  calendar period, bounded `[prev-start, prev-end)`, shared across the
  calendar-windowed commands." Confidence 0.82. DEC-032 extends DEC-028 (the
  window family it foresaw this modifier joining) and reuses DEC-030's bounded
  upper-edge pattern; it does not relitigate either.
- **Resolves** `wrapped-default-current-vs-last-completed` in
  `/guidance/questions.yaml`: the bare-command default stays CURRENT (DEC-030
  choice 2 holds); `--previous` is the uniform last-completed path across the
  whole family, keeping "which period" logic in one mechanism rather than split
  between a default and a flag. (The entry is already `status: resolved` /
  `resolution: current-calendar-year`; this spec is the `--previous` half it
  named, and closes the DEC-030 revisit trigger.)

## Goal

Add a `--previous` boolean modifier to `brag impact`, `brag story`, and `brag
wrapped` that shifts the selected calendar window to the **last-completed**
period, as a **bounded** `[prev-start, prev-end)` window. `--previous` composes
with each command's existing window selection, errors clearly (`UserError`) on
incoherent combinations, and is deterministic under the injected `nowFunc` /
`storyNowFunc` clock seam.

## Inputs

- **Files to read:**
  - `internal/cli/window.go` — the shared calendar-window core. `windowCutoff`
    currently returns only a lower bound (`cutoff time.Time, scope string`); the
    period end is implicitly "now". `--previous` needs a bounded window, so this
    spec **extends `windowCutoff`** to optionally return an exclusive upper bound
    — see Implementation Context (LD1). `selectedWindow`, `windowFlagsSet`, and
    `windowFlagNames` are reused unchanged.
  - `internal/cli/impact.go` — `runImpact` wiring + the `nowFunc` seam.
  - `internal/cli/story.go` — `runStory` + `resolveWindow` (the profile-default
    composition path `--previous` must slot into) + `storyNowFunc`.
  - `internal/cli/wrapped.go` — `parseWrappedPeriod` (the bounded
    `[start, nextBoundary)` window + the `created_at < nextBoundary` Go filter
    `--previous` mirrors) and `runWrapped`.
  - `internal/export/impact.go` — the renderer receives an already-in-window
    slice; `--previous` changes only *which* slice, not the renderer.
  - `DEC-028` (impact's calendar windows — the family this modifier joins),
    `DEC-030` (wrapped's bounded upper edge — the pattern reused), `DEC-032`
    (this spec's decision).
- **Related code paths:** `internal/cli/` (window, impact, story, wrapped),
  `internal/storage/entry.go` (`ListFilter.Since`).

## Outputs

- **Files created:** none (no new files — `--previous` is a modifier layered on
  the existing window path).
- **Files modified:**
  - `internal/cli/window.go` — extend `windowCutoff` to return an exclusive
    upper boundary and accept a `previous bool`; add the previous-period
    boundary math (LD1). Update its callers' signatures.
  - `internal/cli/impact.go` — register `--previous`; thread it into
    `windowCutoff`; apply the Go upper-bound filter when bounded.
  - `internal/cli/story.go` — register `--previous`; thread it through
    `resolveWindow` (composes with an explicit window flag OR the profile
    default); apply the Go upper-bound filter when bounded.
  - `internal/cli/wrapped.go` — register `--previous`; when set with no
    positional period arg, shift the default period back one year (LD4).
  - `internal/cli/impact_test.go`, `story_test.go`, `wrapped_test.go`,
    `window.go`'s test coverage — the Failing Tests below.
  - `docs/api-contract.md`, `docs/tutorial.md`, `README.md`, `AGENTS.md` §11 —
    the docs sweep (build re-verifies exact lines per §12 audit-grep; the
    enumerated hits are in Implementation Context).
  - The STAGE-013 backlog line for SPEC-053 flips to a build state at build.
- **New exports:** none. `windowCutoff`'s signature changes (an internal
  helper, not exported); no public API surface added.
- **Database changes:** none. Read-only over existing schema
  (`Store.List(ListFilter{Since: prevStart})` + an in-CLI `created_at < prevEnd`
  upper-bound filter, exactly as `wrapped` already does).

## Acceptance Criteria

- [ ] `brag impact --quarter --previous` reports the **entire last-completed
      calendar quarter** as a bounded window `[prev-quarter-start,
      prev-quarter-end)`, where `prev-quarter-end` is the current quarter's
      start. Same for `--month --previous` and `--year --previous`.
- [ ] The window is **bounded on both ends**: an entry created the instant
      before `prev-start` is excluded (lower bound), and an entry created at or
      after `prev-end` — including one in the *current* in-progress period — is
      excluded (upper bound). This is the load-bearing divergence from the
      current-period `[cutoff, now]` window.
- [ ] The previous-period boundaries are computed with `time.Date` + `AddDate`
      (calendar shift), never day subtraction, and roll year boundaries
      correctly: with `now` in January, `--month --previous` lands in the prior
      December of the prior year; `--quarter --previous` lands in the prior Q4.
- [ ] `--previous` on `brag story` shifts BOTH an explicit window flag
      (`--quarter --previous`) AND the audience profile's default window
      (`--previous` alone → the profile default period, shifted back one) to the
      last-completed period, bounded.
- [ ] `brag wrapped --previous` (no positional period arg) covers the
      **last-completed calendar year** (e.g. 2025 when run in 2026), bounded —
      identical to `brag wrapped <last-year>`.
- [ ] `--previous` combined with an **explicit** `wrapped` period arg
      (`brag wrapped 2026 --previous` or `brag wrapped 2026 Q3 --previous`) is a
      `UserError` (LD4): the positional arg already names a bounded period, so
      shifting it is redundant/ambiguous. Only `brag wrapped --previous` (bare,
      shift-the-default) is valid.
- [ ] `--previous` combined with `--since` on `impact`/`story` is a `UserError`
      (LD3): `--since` is an explicit anchor, not a calendar period, so
      "the previous `--since`" is undefined.
- [ ] `--previous` on `impact` with **no** window flag is a `UserError` — the
      existing "exactly one window required" rule still fires first (`--previous`
      is a modifier, not a window). On `story`, `--previous` with no window flag
      is VALID (it shifts the profile default).
- [ ] The `scope` provenance token echoes the shift: `--quarter --previous` →
      `"quarter:previous"`, `--month --previous` → `"month:previous"`,
      `--year --previous` → `"year:previous"` (LD5). `brag wrapped --previous`
      echoes the concrete resolved year (e.g. `"2025"`) — it resolves to a named
      period, so it reuses the existing named-period scope, not a `:previous`
      suffix.
- [ ] `--previous` is deterministic given `(entries, now)` — the injected
      `nowFunc` / `storyNowFunc` seam makes the boundary math and any golden
      byte-stable.
- [ ] All existing impact/story/wrapped tests stay green (the change is additive
      — the current-period path is unchanged when `--previous` is absent).

## Failing Tests

Written during **design**, BEFORE build. Every boundary date below was computed
at design time against the REAL `time.Date`/`AddDate` calendar math (via a
scratch program inside the module tree, since removed) — they are faithful, not
hand-typed. The canonical frozen instant is:

```go
// spec53Now: a fixed instant mid-Q3-2026 (July), so every window has a
// well-defined, non-degenerate previous period within 2026 except --year
// (which lands in 2025). Chosen so the boundaries are legible.
var spec53Now = time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
```

Computed previous-period boundaries at `spec53Now` (faithful, `[start, end)`):

| flag                  | prev-start   | prev-end (exclusive) |
|-----------------------|--------------|----------------------|
| `--quarter --previous`| `2026-04-01` | `2026-07-01`         |
| `--month --previous`  | `2026-06-01` | `2026-07-01`         |
| `--year --previous`   | `2025-01-01` | `2026-01-01`         |

And at the year-boundary instant `time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)`:

| flag                  | prev-start   | prev-end (exclusive) |
|-----------------------|--------------|----------------------|
| `--quarter --previous`| `2025-10-01` | `2026-01-01`         |
| `--month --previous`  | `2025-12-01` | `2026-01-01`         |
| `--year --previous`   | `2025-01-01` | `2026-01-01`         |

### `internal/cli/window_test.go` (or the impact test file if window has none)

- **`TestWindowCutoff_PreviousBoundaries`** (LOAD-BEARING — the boundary-math
  core). Table-driven over `{window, now}` → `{wantStart, wantEnd, wantScope,
  wantBounded}`, asserting the extended `windowCutoff(window, sinceRaw, now,
  previous=true)` returns exactly the `[start, end)` pairs in both tables above
  and the `scope` tokens `"quarter:previous"` / `"month:previous"` /
  `"year:previous"`. Covers `spec53Now` AND the January year-boundary instant
  (the `AddDate` roll). Also asserts that with `previous=false` the returned
  window is unchanged from today's behavior (bounded=false, the current-period
  cutoff), so the extension is additive.

- **`TestWindowCutoff_PreviousSinceIsUserError`**. `windowCutoff("since",
  "2026-01-01", now, previous=true)` returns a `UserError` (assert via
  `errors.As`) — "the previous `--since`" is undefined (LD3). (The CLI layer
  also rejects the `--since --previous` flag combo before reaching here; this
  pins the core helper's own guard so the invariant holds even if a future
  caller forgets the flag-level check.)

### `internal/cli/impact_test.go`

(Harness reuses `seedImpactEntry` / `withNowFunc` / `runImpactCmd` unchanged.)

- **`TestImpactCmd_PreviousQuarterBounded`** (LOAD-BEARING — the bounded-window
  divergence). With `nowFunc` frozen at `spec53Now`, seed four entries: one at
  `2026-03-31` (in Q1, before prev-Q2), one at `2026-04-01` (prev-Q2 start, IN),
  one at `2026-06-30` (prev-Q2 end, IN), one at `2026-07-01` (current Q3 start,
  AFTER prev-end — excluded). `brag impact --quarter --previous` includes ONLY
  the two Q2 entries: `Entries: 2/2 with impact` (seed both Q2 entries with an
  impact string) and the Q1 + Q3 titles are absent. This proves the upper bound
  is the *current* period start, not `now` — the entry created "now-ish" in Q3
  is excluded. `scope` echoes `quarter:previous`.

- **`TestImpactCmd_PreviousMonthAndYear`**. Two sub-cases at `spec53Now`:
  - `--month --previous` → window `[2026-06-01, 2026-07-01)`; a `2026-06-15`
    entry is IN, a `2026-05-31` entry OUT, a `2026-07-02` entry OUT. Scope
    `month:previous`.
  - `--year --previous` → window `[2025-01-01, 2026-01-01)`; a `2025-07-01`
    entry IN, a `2024-12-31` entry OUT, a `2026-01-02` entry OUT. Scope
    `year:previous`.

- **`TestImpactCmd_PreviousYearBoundaryRoll`**. `nowFunc` frozen at
  `2026-01-15`; `--month --previous` → `[2025-12-01, 2026-01-01)`: a
  `2025-12-20` entry IN, a `2026-01-05` entry OUT. Proves `AddDate` rolls the
  month back across the year boundary (Dec of the prior year), not "month 0".

- **`TestImpactCmd_PreviousWithSinceIsUserError`**. `brag impact --since
  2026-01-01 --previous` → `UserError` (assert `errors.Is(err, ErrUser)`),
  stdout empty, message names `--previous` and `--since` (LD3).

- **`TestImpactCmd_PreviousWithoutWindowIsUserError`**. `brag impact --previous`
  (no `--quarter/--month/--year/--since`) → `UserError`: the existing
  "exactly one window required" rule fires (a modifier is not a window). stdout
  empty.

- **`TestImpactCmd_NoPreviousUnchanged`** (regression guard). `brag impact
  --quarter` (no `--previous`) at `spec53Now` still reports the CURRENT quarter
  up to now — a `2026-07-05` entry (current Q3) is IN, a `2026-06-30` entry
  (prev Q2) is OUT — and `scope` echoes plain `quarter`. Confirms the additive
  extension does not perturb the current-period path.

### `internal/cli/story_test.go`

(Harness reuses story's `withStoryNowFunc` / seed helpers unchanged.)

- **`TestStoryCmd_PreviousExplicitWindowBounded`**. `nowFunc` (story's seam)
  frozen at `spec53Now`; `brag story --audience exec --quarter --previous` →
  `scope` echoes `quarter:previous`, and a seeded `2026-05-01` entry (prev-Q2)
  appears in the bundle while a `2026-07-03` entry (current Q3) does not. Proves
  `--previous` shifts an explicit window flag AND applies the bounded upper edge
  on the story path.

- **`TestStoryCmd_PreviousShiftsProfileDefault`**. `brag story --audience me
  --previous` (no window flag). The `me` profile's default window is `year`
  (confirmed in `internal/story/profile.go`), so `--previous` → the
  last-completed year `[2025-01-01, 2026-01-01)`, bounded; `scope` echoes
  `year:previous`. A `2025-06-01` entry IN, a `2026-02-01` entry (current year)
  OUT. Proves `--previous` composes with the profile default, not just explicit
  flags.

- **`TestStoryCmd_PreviousWithSinceIsUserError`**. `brag story --audience me
  --since 2026-01-01 --previous` → `UserError`, stdout empty (LD3, story path).

### `internal/cli/wrapped_test.go`

(Harness reuses `seedWrappedEntry` / `withNowFunc` unchanged.)

- **`TestWrappedCmd_PreviousDefaultsToLastYear`** (LOAD-BEARING for wrapped).
  `nowFunc` frozen at `spec53Now` (2026); `brag wrapped --previous` (no
  positional arg) → `scope` echoes `2025` (the concrete resolved year, NOT a
  `:previous` suffix — LD5), covering the bounded window `[2025-01-01,
  2026-01-01)`. Seed a `2025-06-15` entry (IN), a `2024-12-31` entry (OUT), and
  a `2026-03-01` entry (current year, OUT): `Entries: 1`, only the 2025 title
  present. Identical result to `brag wrapped 2025` at the same `now`.

- **`TestWrappedCmd_PreviousWithExplicitPeriodIsUserError`** (LD4). Table:
  `["2026"] + --previous`, `["2026","Q3"] + --previous`. Each → `UserError`
  (assert `errors.As`), stdout empty, message explains `--previous` is only
  valid with no positional period (the arg already names a bounded period).

- **`TestWrappedCmd_PreviousQuarterlessStillYear`** (scope-consistency guard).
  `brag wrapped --previous` resolves to a YEAR (not a quarter) — `wrapped`'s
  bare default is annual (DEC-030 choice 2), so `--previous` shifts *that*
  (the year), giving `scope == "2025"`. Documents that `--previous` on `wrapped`
  is the annual-retrospective path; a previous-*quarter* wrapped is out of scope
  (a user names it positionally: `brag wrapped 2026 Q2`).

### Help / provenance tests

- **`TestImpactCmd_HelpShowsPrevious`** / **`TestStoryCmd_HelpShowsPrevious`** /
  **`TestWrappedCmd_HelpShowsPrevious`**. Each command's `--help` contains the
  literal `--previous` flag and a distinctive example line (e.g. `brag impact
  --quarter --previous`, `brag wrapped --previous`) — unique-token discipline
  (AGENTS.md §9). NOT-contains self-audit (§12): the locked `Long` strings below
  are grep-clean of any token a `--previous`-absent test asserts absent.

- **`TestImpactCmd_StdoutStderrSeparation_Previous`**. A successful `--previous`
  run writes only to stdout (`errBuf.Len()==0`); the `--since --previous`
  incoherent combo writes only to stderr (per
  `stdout-is-for-data-stderr-is-for-humans`).

## Implementation Context

*Read this section (and the files it points to) before starting the build
cycle.*

### Decisions that apply

- `DEC-032` (this spec) — locks `--previous` = the last-completed calendar
  period, a **bounded** `[prev-start, prev-end)` window where `prev-end` is the
  current period's start; shared uniformly across `impact`/`story`/`wrapped`;
  computed via `time.Date` + `AddDate` (never day subtraction); the scope token
  gains a `:previous` suffix on `impact`/`story` and resolves to the concrete
  named year on `wrapped`.
- `DEC-028` — the calendar-window family `--previous` joins. DEC-028 explicitly
  foresaw this modifier ("A `--previous` modifier is a clean future addition if
  the complete-period workflow materializes" / "add `--previous`, don't change
  the default"). `--previous` reuses `windowCutoff`'s current-period start math
  as its anchor, then shifts back one period. The four window flags stay
  mutually exclusive and `impact` still requires exactly one.
- `DEC-030` — the **bounded upper-edge** pattern (`[start, next-boundary)`,
  filtered `created_at < nextBoundary` in Go) `--previous` reuses verbatim.
  DEC-030 choice 2 (wrapped's bare default = current year) HOLDS; `--previous`
  is the last-completed layer it deferred to. This spec closes DEC-030's
  revisit trigger.
- `DEC-014` — the envelope is untouched; `--previous` changes only which slice
  the renderer receives and the `scope` provenance string.
- `DEC-008` — `ParseSince` is unaffected; `--since --previous` is rejected as an
  incoherent combo (LD3).
- `DEC-007` — required/invalid-flag validation via `UserErrorf` (the incoherent
  combos).

### Constraints that apply

- `no-sql-in-cli-layer` — the previous-period read is
  `Store.List(ListFilter{Since: prevStart})`; the `created_at < prevEnd`
  upper-bound filter is applied in Go in the CLI layer (exactly as `wrapped`
  already does — `ListFilter` has no `Until` field). No `database/sql` import
  under `internal/cli/`. Raw SQL for `created_at` rewriting stays confined to
  the test files (existing `seed*Entry` helpers).
- `stdout-is-for-data-stderr-is-for-humans` — digest body to stdout; the
  incoherent-combo `UserError`s to stderr. Tested by the separation test.
- `no-cgo` — no new deps; pure-Go stdlib (`time`) only.
- `test-before-implementation` — the Failing Tests above are written first.
- `errors-wrap-with-context` — storage/config errors wrapped; the incoherent
  combos use `UserErrorf`.

### Prior related work

- `SPEC-048` (shipped) — `brag impact`; DEC-028; `internal/cli/window.go`'s
  `windowCutoff`/`selectedWindow` core and the `nowFunc` seam. `--previous`
  extends `windowCutoff`.
- `SPEC-049` (shipped) — `brag story`; `resolveWindow` (profile-default vs
  explicit-flag composition) that `--previous` slots into, and `storyNowFunc`.
- `SPEC-051` (shipped) — `brag wrapped`; DEC-030; the bounded
  `[start, nextBoundary)` window + the `created_at < nextBoundary` Go filter
  `--previous` reuses, and `parseWrappedPeriod`.
- `SPEC-018` (shipped) — the DEC-014 envelope + the filter-echo provenance
  pattern the `scope` token extends.

### Out of scope (for this spec specifically)

- **A previous-*quarter* `wrapped`.** `brag wrapped --previous` shifts the
  annual default only (LD4/LD6); a completed quarter is named positionally
  (`brag wrapped 2026 Q2`). A `--previous` that also composed with an explicit
  positional period is a rejected combo (LD4).
- **`--previous N` (N periods back).** `--previous` is a boolean = one period
  back. A repeat-count is a future spec if the workflow appears.
- **`--since --previous`.** Rejected as incoherent (LD3) — `--since` is an
  explicit anchor, not a calendar period.
- **Changing any bare-command default.** `impact` still requires an explicit
  window; `wrapped`'s bare default stays the current year (DEC-030 choice 2).
  `--previous` is purely additive.
- **Any `ListFilter.Until` field.** The upper bound stays a one-line Go filter
  (the second bounded-window consumer does NOT yet justify promoting it — see
  DEC-030's own revisit note; two Go-side filters is still below that bar).
- Any LLM call-out; any network; any schema change.

## Notes for the Implementer

- **Extend `windowCutoff` (LD1).** New signature:
  `windowCutoff(window, sinceRaw string, now time.Time, previous bool) (start time.Time, end time.Time, scope string, err error)`.
  - When `previous == false`: `start` = today's current-period cutoff (math
    unchanged), `end` = the zero `time.Time` **sentinel meaning "open upper edge
    / up to now"**, `scope` = today's token (`"quarter"` / `"since:<raw>"` etc).
    Callers treat a zero `end` as "no upper-bound filter" — preserving the
    current `[cutoff, now]` behavior byte-for-byte.
  - When `previous == true`: compute the current-period start with the SAME
    `time.Date` math, then shift it back one period as the previous-period start,
    and set `end` to the current-period start (the exclusive upper boundary):
    - `quarter`: `curStart = time.Date(now.Year(), qStartMonth, 1, ...)`;
      `start = curStart.AddDate(0, -3, 0)`; `end = curStart`.
    - `month`: `curStart = time.Date(now.Year(), now.Month(), 1, ...)`;
      `start = curStart.AddDate(0, -1, 0)`; `end = curStart`.
    - `year`: `curStart = time.Date(now.Year(), 1, 1, ...)`;
      `start = curStart.AddDate(-1, 0, 0)`; `end = curStart`.
    - `scope` = `window + ":previous"`.
    - `since`: return `UserErrorf(...)` — "the previous `--since`" is undefined
      (LD3). This is the helper-level guard the CLI's flag check backs up.
  - Update all three callers (`runImpact`, `resolveWindow` in story, and — only
    if `wrapped` routes through `windowCutoff`, which it does NOT today — leave
    wrapped's own `parseWrappedPeriod` path). `impact`/`story` pass `previous`
    read from the flag; the current callers pass `false` where they don't set
    it. **NB:** wrapped does its own period parsing, so its `--previous` is
    handled in `parseWrappedPeriod` / `runWrapped`, NOT via `windowCutoff` (LD6).
- **The Go upper-bound filter (impact/story).** After
  `entries, _ := s.List(ListFilter{Since: start})`, when `end` is non-zero keep
  only `e.CreatedAt.Before(end)` — the same one-liner `wrapped` uses. When `end`
  is the zero sentinel, skip the filter (current-period behavior). Thread the
  bounded upper edge before the renderer sees the slice; `EntriesInWindow` /
  `len(entries)` then reflect the bounded set.
- **Flag registration.** Add `cmd.Flags().Bool("previous", false, "shift the
  window to the last-completed period")` to `NewImpactCmd`, `NewStoryCmd`,
  `NewWrappedCmd`. Read it via `cmd.Flags().GetBool("previous")`.
- **Incoherent-combo guards (CLI layer, before the read):**
  - impact/story: if `--previous` AND `--since` are both set → `UserErrorf(...)`
    naming both flags (LD3). Do this check where the window is resolved
    (`runImpact` after `selectedWindow`; story's `resolveWindow`).
  - wrapped: if `--previous` AND `len(args) > 0` → `UserErrorf(...)` (LD4).
    Check at the top of `runWrapped` (or inside `parseWrappedPeriod` given the
    args + a `previous` param).
  - impact `--previous` with no window flag: no new code — `selectedWindow`
    already errors "one of ... is required" (a modifier is not a window). The
    Failing Test just pins that this ordering holds.
- **wrapped `--previous` (LD6).** In `runWrapped`/`parseWrappedPeriod`: when
  `previous` is set and `args` is empty, resolve the year to `now.Year() - 1`
  and build the same bounded annual window `parseWrappedPeriod` builds for
  `brag wrapped <year>`. The `scope` is the concrete year string (`"2025"`),
  NOT `"year:previous"` — `wrapped` names a concrete period, so it reuses the
  named-period scope (LD5). Result is byte-identical to `brag wrapped 2025` at
  the same `now`.
- **Scope tokens (LD5).** impact/story: `"quarter:previous"` /
  `"month:previous"` / `"year:previous"` (the plain token + `:previous`).
  wrapped: the concrete resolved year (`"2025"`). Rationale: impact/story echo
  the *mechanism* (a relative window), so the reader sees it was the previous
  period; wrapped resolves to a *named* period and its existing scope already
  says which one.
- **Determinism.** Everything keys off the injected `nowFunc` / `storyNowFunc`.
  No `time.Now()` in new code. Goldens/assertions use the frozen `spec53Now`.
- **Docs sweep (AGENTS.md §9 / §12 audit-grep).** At build, grep `docs/
  README.md AGENTS.md` for the calendar-window flag enumerations
  (`--quarter`/`--month`/`--year`) in the `impact`/`story`/`wrapped` sections
  and add `--previous` where the window flags are documented; add a `--previous`
  line to each command's `docs/api-contract.md` flag table; add a `--previous`
  §11 glossary term to `AGENTS.md`. Enumerate the exact hits in `## Outputs` at
  build after running the grep (design lists the files; build re-verifies the
  exact lines per §12 audit-grep cross-check). The `wrapped-default-...`
  question entry in `guidance/questions.yaml` is already `resolved`; add a note
  under its `resolution_notes` that SPEC-053 delivered the `--previous` half (or
  leave as-is — it already names SPEC-053; build's call).

### Locked design decisions (build-time)

1. **LD1 — Extend the shared `windowCutoff`, do NOT fork a parallel previous-
   window path.** `windowCutoff` gains a `previous bool` param and a second
   `end time.Time` return (zero = open/up-to-now; non-zero = exclusive upper
   bound). The previous-period start is the current-period start shifted back
   one period via `AddDate`; the upper bound is the current-period start. This
   keeps ALL calendar-anchor math in one function (DEC-028's core), reusing the
   exact current-period start computation as the anchor. *Rejected:* a separate
   `previousWindowCutoff` helper — would duplicate the quarter/month/year start
   math and risk the two drifting; the orchestrator brief explicitly says "reuse
   the shared window infra, do NOT fork a parallel window path."
2. **LD2 — Bounded `[prev-start, prev-end)` window; upper bound = current-period
   start.** A completed period has a real end. Filter `Since: prevStart` in SQL,
   then `created_at < prevEnd` in Go (the DEC-030 pattern). The upper bound is
   the CURRENT period's start — the boundary between the completed period and
   the in-progress one — so an entry created "now" (in the current period) is
   correctly excluded. *Rejected:* `[prev-start, now]` (impact's open upper
   edge) — that would wrongly include current-period entries in a
   "last-completed-period" digest, defeating the whole point.
3. **LD3 — `--previous` is mutually exclusive with `--since`; a `UserError`.**
   `--since` is an explicit user-supplied anchor, not a calendar period, so
   "the previous `--since`" is undefined. Guarded at BOTH the CLI flag level
   (naming both flags) AND the `windowCutoff` helper level (defense in depth).
   *Rejected:* silently ignoring `--previous` when `--since` is set — violates
   the "error clearly on incoherent combos" brief and hides a user mistake.
4. **LD4 — On `wrapped`, `--previous` is valid ONLY with no positional period;
   `--previous` + an explicit period arg is a `UserError`.** The positional arg
   (`2026`, `2026 Q3`) already names a bounded period; "the previous 2026 Q3" is
   ambiguous (previous quarter? previous year?) and redundant (name it directly:
   `2026 Q2`). The simplest coherent rule: `--previous` shifts the *default*
   only. *Rejected alternatives:* (a) `brag wrapped 2026 Q3 --previous` → 2026 Q2
   — introduces a second period-shifting mechanism competing with the positional
   surface, and forces a quarter-vs-year "which axis shifts" rule the positional
   surface already handles cleanly; (b) `--previous` silently ignored when a
   period is named — hides a user mistake. The chosen rule keeps `--previous`
   meaning exactly one thing on `wrapped`: "the last-completed default year."
5. **LD5 — Scope token: `<window>:previous` on impact/story; the concrete year
   on wrapped.** impact/story echo a *relative* window, so the provenance says
   `quarter:previous` — a reader (or `jq .scope`) sees it was the previous
   period, distinct from the current `quarter`. wrapped resolves to a *named*
   concrete period, so it reuses the existing `"2025"` scope (no `:previous`
   suffix) — the scope already says which year. *Rejected:* a uniform
   `:previous` suffix everywhere — on wrapped it would double-encode ("2025 and
   also previous"), and lose the concrete-year provenance wrapped's scope exists
   to give; a `previous: true` sibling key in the envelope — a DEC-014 envelope
   reshape for one modifier, heavier than a scope-string convention.
6. **LD6 — wrapped's `--previous` is handled in its own period parser, not via
   `windowCutoff`.** `wrapped` never routed through `windowCutoff` (it has
   `parseWrappedPeriod` for its bounded positional periods); `--previous` there
   is `now.Year()-1` fed into the SAME `parseWrappedPeriod` year path, so the
   annual bounded window is built by the existing, tested code. impact/story
   (which DO use `windowCutoff`) get their `--previous` from the extended
   helper. *Rejected:* routing wrapped through `windowCutoff` to share the
   previous-year math — wrapped's window is positional/named and already bounded;
   forcing it through the flag-oriented `windowCutoff` would be a larger,
   riskier refactor than reusing `parseWrappedPeriod` with `now.Year()-1`.
7. **LD7 — `--previous` with no window flag: impact errors, story shifts the
   profile default.** On `impact`, a window is required (DEC-028), so `--previous`
   alone hits the existing "exactly one window required" `UserError` — no new
   code. On `story`, no window flag means "use the profile default," so
   `--previous` shifts that default period back one (composing in `resolveWindow`
   before the `windowCutoff` call, passing `previous=true`). *Rejected:* making
   `--previous` imply a default window on `impact` — would smuggle a default into
   a command DEC-028 deliberately gave none; keep impact's "explicit window
   required" invariant intact.

### Rejected alternatives (build-time)

- **A `--last-quarter` / `--last-month` / `--last-year` flag trio.** Rejected:
  triples the flag surface and doesn't compose with `story`'s profile default.
  One orthogonal `--previous` modifier that shifts whatever window is selected
  is smaller and composes across all three commands — the whole point of a
  modifier.
- **`--previous` changes the default instead of adding a modifier.** Rejected:
  DEC-028 and DEC-030 both explicitly chose current-period defaults and deferred
  last-completed to *this* modifier. Flipping a default now would relitigate two
  shipped decisions and split "which period" logic.
- **Promote the Go upper-bound filter to a `ListFilter.Until` field now that
  there are two bounded-window consumers.** Rejected here (noted as DEC-030's
  own revisit trigger): impact/story share ONE filter one-liner with wrapped's;
  the promotion is a storage-layer change with its own small DEC, out of scope
  for this additive modifier. Left as a future call if a THIRD consumer appears.
- **A `:prev` short suffix instead of `:previous`.** Rejected: `:previous`
  matches the flag name exactly (no abbreviation to memorize) and reads clearly
  in `jq .scope` output; the few extra bytes in a provenance string are free.

---

## Build Completion

*Filled in at the end of the **build** cycle, before advancing to verify.*

- **Branch:** feat/spec-053-previous
- **PR (if applicable):** (draft opened at design)
- **All acceptance criteria met?** (build fills in)
- **New decisions emitted:** DEC-032 (emitted at design).
- **Files changed:** (build enumerates after the docs-sweep audit-grep)
- **Docs sweep (§9 status-change / §12 audit-grep cross-check):** (build fills in)
- **Deviations from spec:** (build fills in)
- **Follow-up work identified:** (build fills in)

### Build-phase reflection (3 questions, short answers)

1. **What was unclear in the spec that slowed you down?** —
2. **Was there a constraint or decision that should have been listed but
   wasn't?** —
3. **If you did this task again, what would you do differently?** —
