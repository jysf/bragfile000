---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-068
  type: story                      # epic | story | task | bug | chore
  cycle: verify
  blocked: false
  priority: medium
  complexity: S                    # S | M | L  (L means split it)

project:
  id: PROJ-006
  stage: STAGE-017
repo:
  id: bragfile

agents:
  architect: claude-opus-4-8
  implementer: claude-opus-4-8     # usually same Claude, different session
  created_at: 2026-07-10

references:
  decisions: [DEC-039, DEC-035, DEC-022]
  constraints:
    - no-sql-in-cli-layer
    - timestamps-in-utc-rfc3339
    - stdout-is-for-data-stderr-is-for-humans
    - errors-wrap-with-context
    - test-before-implementation
  related_specs: [SPEC-056, SPEC-038, SPEC-007]
---

# SPEC-068: brag list --day (local calendar-day window)

## Context

Why does this spec exist? What problem does it solve?

`brag list` today has only a lower bound (`--since`) and no clean way to scope to
a single day. "What did I do yesterday?" needs a `jq` upper bound (documented in
tutorial §9), and `--since`'s bare-date UTC-midnight anchoring skews for a
non-UTC user (in PDT, "today" via `--since` starts at 5pm the day before). This
spec adds ONE flag — `brag list --day <YYYY-MM-DD|today|yesterday>` — that scopes
the listing to exactly one **local** calendar day via `ListFilter.Since` +
`Until` (the exclusive upper bound already shipped in SPEC-056/DEC-035).

- Parent stage: `STAGE-017` (list time-window ergonomics), the ship-early
  read-surface win that opens `PROJ-006`. This is the stage's only spec.
- The one design fork (LOCAL vs UTC day) is settled in **DEC-039**.

## Goal

Add `brag list --day <YYYY-MM-DD|today|yesterday>`: scope the listing to exactly
that single **local** calendar day (`[local-midnight, next-local-midnight)`) by
setting both bounds on the existing `storage.ListFilter`. One flag, one concept —
mutually exclusive with `--since`; composes with every other filter.

## Inputs

- **Files to read:**
  - `internal/cli/list.go` — flag defs + how `--since` builds `ListFilter{Since}`.
  - `internal/cli/since.go` — `ParseSince` + the inline-clock L4 impurity.
  - `internal/storage/entry.go` — `ListFilter.Until` (exists, DEC-035).
  - `internal/storage/store.go` — the `Until` WHERE block (exclusive `<`).
- **Related code paths:** `internal/cli/impact.go` (`nowFunc`, the *UTC* clock
  seam — deliberately NOT reused; see DEC-039), `internal/storage/storagetest/`
  (`Backdate` test helper).

## Outputs

- **Files modified:**
  - `internal/cli/since.go` — add `var clock = time.Now` (injectable local seam);
    route the `--since` duration path through it (L4 fix); add `ParseDay`.
  - `internal/cli/list.go` — add `--day` flag, its window-setting + `--since`
    mutual-exclusion logic, and Long examples.
  - `internal/cli/list_test.go`, `internal/cli/since_test.go` — the failing
    tests below.
  - `docs/api-contract.md`, `docs/tutorial.md` — document `--day` alongside
    `--since`.
- **New exports:** `ParseDay(value string) (start, end time.Time, err error)` in
  package `cli`.
- **Database changes:** none. Pure query-time filtering over the existing
  `Since`/`Until` fields; no migration.

## Acceptance Criteria

- [ ] `brag list --day <YYYY-MM-DD>` returns exactly that local day's entries
      (`[day-start, next-day-start)`): a 23:30-local entry is IN, a
      00:30-next-day-local entry is OUT (local bounds + exclusive upper edge).
- [ ] `brag list --day today` / `--day yesterday` resolve against the local wall
      clock; an evening-PDT entry (yesterday LOCAL, today UTC) lands in `--day
      yesterday`, NOT `--day today`.
- [ ] `--day` keywords are case-insensitive; a bare `YYYY-MM-DD` is a local day.
- [ ] `--day <garbage>` → `UserError` naming the accepted forms (today,
      yesterday, YYYY-MM-DD); exit 1, empty stdout, message on stderr.
- [ ] `--day` + `--since` → `UserError` naming the conflict.
- [ ] `--day` composes with `--project`/`--type`/`--tag`/`--limit`/`-P`/`--format`.
- [ ] Plain `list` and existing `--since` behavior unchanged (green regressions).
- [ ] Help output advertises `--day`.

## Failing Tests

Written during **design**, BEFORE build. Made to pass during **build**.

- **`internal/cli/list_test.go`** (with a `stubDayClock` helper + fixed PDT zone)
  - `TestListCmd_DayFilter_LocalDayBoundsExclusiveUpperEdge` — `--day <date>`
    includes the 23:30-local entry, excludes the 00:30-next-day entry
    (DEC-039 §1/§2: local bounds + exclusive upper edge).
  - `TestListCmd_DayFilter_TodayYesterdayLocalSkew` — with a stubbed
    evening/morning-PDT clock, a "yesterday-local but today-UTC" entry lands in
    `--day yesterday`, NOT `--day today` (DEC-039 §1: the skew this fixes).
  - `TestListCmd_DayFilter_GarbageIsUserError` — `--day notaday` → `ErrUser`,
    empty stdout, message names `today`/`yesterday`/`YYYY-MM-DD` (DEC-039 parsing).
  - `TestListCmd_DayFilter_MutuallyExclusiveWithSince` — `--day today --since 7d`
    → `ErrUser` naming both flags (DEC-039 §4), empty stdout.
  - `TestListCmd_DayFilter_ComposesWithProject` — `--day <date> --project X`
    filters within the day (DEC-039 §4: filters compose).
  - `TestListCmd_HelpShowsFilters` — extended to require `--day` in help.
- **`internal/cli/since_test.go`**
  - `TestParseDay_ExplicitDateIsLocalMidnightHalfOpen` — explicit date →
    `[local-midnight, next-local-midnight)` in `clock().Location()`.
  - `TestParseDay_TodayYesterdayResolveAgainstClock` — keywords resolve off the
    stubbed clock (DEC-039 §3, injectable clock).
  - `TestParseDay_KeywordsAreCaseInsensitive` — `TODAY` == `today`.
  - `TestParseDay_InvalidValue` — `""`, `notaday`, `7d`, `2026-13-01`, ` today `
    all error.
  - (Existing `TestParseSince_*` stay green — the `--since` duration path now
    routes through the `clock` seam but behaves identically.)

## Implementation Context

*Read this section (and the files it points to) before starting the build cycle.*

### Decisions that apply

- `DEC-039` — the day-boundary semantics (LOCAL calendar day via Since+Until;
  today/yesterday keywords; injectable clock; mutually exclusive with `--since`).
- `DEC-035` — `ListFilter.Until` is the shipped exclusive upper bound `--day`'s
  end rides on; this spec is a CLI wire-up, not new storage.
- `DEC-022` — the local-day streak precedent: the "derive-local, store-UTC"
  carve-out (storage stays UTC; only the derived boundary localizes) and the
  "location rides on the injected now" clock discipline.

### Constraints that apply

- `no-sql-in-cli-layer` — the window is expressed via `ListFilter`; no SQL in CLI.
- `timestamps-in-utc-rfc3339` — untouched: storage stays UTC; `Store.List`
  formats the (locally-computed) bounds `.UTC().Format(RFC3339)`. Compute local
  midnight, then let storage compare in UTC.
- `stdout-is-for-data-stderr-is-for-humans` — errors on stderr, empty stdout;
  CLI tests use separate outBuf/errBuf and assert no cross-leakage.
- `errors-wrap-with-context`, `test-before-implementation`.

### Prior related work

- `SPEC-056` / `DEC-035` (shipped) — `ListFilter.Until`.
- `SPEC-038` / `DEC-022` (shipped) — local-day streak.
- `SPEC-007` (shipped) — `ListFilter` + `Since`.

### Out of scope (for this spec specifically)

- A general `--until` flag / arbitrary bounded ranges — YAGNI; `--day` covers the
  need (DEC-039 alt B). The storage field is public if it is ever wanted.
- Exposing `--day` on the MCP `brag_list` tool or the calendar-window commands.
- Changing `--since`'s bare-date UTC-midnight semantics (backward-compat).

## Notes for the Implementer

- **Do NOT reuse `impact.go`'s `nowFunc`** — it returns `time.Now().UTC()`, so
  `now.Location()` is UTC and local-day math would silently become UTC-day. Add a
  DISTINCT `var clock = time.Now` in `since.go` (returns local). Both seams
  coexist in package `cli` on purpose; document each.
- Build day bounds with `time.Date(y, m, d, 0,0,0,0, loc)` and
  `start.AddDate(0, 0, 1)` — calendar arithmetic, DST-correct. Never
  `start.Add(24*time.Hour)`.
- `ParseDay` keyword match: `strings.ToLower(value)` (no trimming — ` today `
  should error, matching `--since`'s strict-whitespace behavior). Explicit date:
  `time.ParseInLocation("2006-01-02", value, loc)`.
- Mutual exclusion lives in `runList` (`cmd.Flags().Changed("day") &&
  Changed("since")`), returned as `UserErrorf`.
- No `test-docs.sh` assertion enumerates `list` flags (verified), so only the
  prose docs (api-contract, tutorial) need the `--day` addition.

---

## Build Completion

*Filled in at the end of the **build** cycle, before advancing to verify.*

- **Branch:** `feat/spec-068-list-time-windows`
- **PR (if applicable):** see PR opened at end of build.
- **All acceptance criteria met?** yes
- **New decisions emitted:**
  - `DEC-039` — `brag list --day` local-calendar-day window semantics.
- **Deviations from spec:**
  - None. Design and build ran in one session (fresh session for SPEC-068 per
    the task); the failing tests were written and confirmed fail-first (undefined
    `clock`/`ParseDay`/`--day`) before implementation.
- **Follow-up work identified:**
  - None required. A future `--until` (arbitrary ranges) or an MCP `brag_list`
    day filter would each be their own spec (both out of scope here).

### Build-phase reflection (3 questions, short answers)

1. **What was unclear in the spec that slowed you down?**
   — Nothing blocked, but the pre-existing `nowFunc` (UTC) in the same `cli`
   package forced a naming/semantics decision: a second, local-zoned `clock`
   seam. Resolved per DEC-039 §3 and the implementer note.

2. **Was there a constraint or decision that should have been listed but wasn't?**
   — No. DEC-035 (Until), DEC-022 (local-day + clock discipline), and the
   UTC-storage constraint carve-out covered every judgment call.

3. **If you did this task again, what would you do differently?**
   — Nothing material. Reusing the shipped `Until` primitive kept this to a CLI
   wire-up plus one clock seam; the local-vs-UTC fork was the only real decision
   and DEC-022 gave it a clear precedent.

---

## Reflection (Ship)

*Appended during the **ship** cycle. Outcome-focused reflection, distinct
from the process-focused build reflection above.*

1. **What would I do differently next time?**
   — <answer>

2. **Does any template, constraint, or decision need updating?**
   — <answer>

3. **Is there a follow-up spec I should write now before I forget?**
   — <answer>
