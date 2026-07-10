---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-056
  type: chore                      # epic | story | task | bug | chore
  cycle: ship
  blocked: false
  priority: medium
  complexity: S                    # S | M | L  (L means split it)

project:
  id: PROJ-005
  stage: STAGE-016
repo:
  id: bragfile

agents:
  architect: claude-opus-4-8
  implementer: claude-opus-4-8     # usually same Claude, different session
  created_at: 2026-07-10

references:
  decisions: [DEC-035, DEC-030, DEC-032]
  constraints:
    - no-sql-in-cli-layer
    - timestamps-in-utc-rfc3339
    - storage-tests-use-tempdir
    - errors-wrap-with-context
    - test-before-implementation
  related_specs: [SPEC-051, SPEC-053, SPEC-045, SPEC-007]
---

# SPEC-056: ListFilter Until storage promotion

## Context

Since `brag wrapped` shipped (DEC-030), the bounded calendar-window UPPER edge
has been filtered in Go, post-`Store.List`, because `storage.ListFilter` has only
a `Since` lower bound. DEC-030 recorded that tradeoff and a revisit trigger:
promote the Go filter to a `ListFilter.Until` field with a storage-layer DEC once
a second bounded-window consumer appears. DEC-032 (`--previous`) added the second
and third consumers (`impact`, `story`) and restated the same deferral for a
third. SPEC-045 (`brag coverage`) then added a FOURTH copy. The count is now four
identical `created_at < <bound>` Go loops — two past DEC-030's bar, one past
DEC-032's bar, well clear of rule-of-three.

This spec is the promotion those two DECs deferred to exactly this threshold
(DEC-035). It is a **behavior-preserving refactor**: add `Until` to
`storage.ListFilter` (`created_at < ?`, `!IsZero()`-guarded, RFC3339 UTC,
symmetric with `Since`), rewire the four consumers off their Go-side upper-bound
loop, and delete the duplicated filtering. It is the direct storage-layer
analogue of the DEC-004 → DEC-015 tag-model promotion.

- Parent stage: `STAGE-016` (v0.4.x polish) — this is its first backlog item
  ("promote the duplicated calendar-window upper bound into `ListFilter.Until`").
- Project: `PROJ-005` (agent-native depth) — substrate-tidying that opens the
  wave on a clean base, per STAGE-016 "Why Now".
- Prior decisions: DEC-030 and DEC-032 (revisit triggers), DEC-035 (this
  promotion), DEC-015 (the canonical prior promote-to-storage precedent).

## Goal

Add an exclusive `Until time.Time` upper-bound field to `storage.ListFilter`
(SQL `created_at < ?`, guarded by `!Until.IsZero()`) and refactor `brag impact`,
`brag story`, `brag wrapped`, and `brag coverage` to set it instead of filtering
the window's upper edge in Go — deleting all four duplicated Go filters, with
existing goldens byte-identical and existing CLI behavior tests still green.

## Inputs

- **Files to read:**
  - `internal/storage/entry.go` — the `ListFilter` struct (add `Until` here,
    modeled on `Since`).
  - `internal/storage/store.go` — `Store.List` WHERE-clause builder (add the
    `Until` block, modeled on the `!f.Since.IsZero()` block).
  - `internal/storage/store_test.go` — existing `TestList_FilterBySince`,
    `TestList_FilterCombined`, `TestList_AuthorComposesWithOtherFilters`
    (+ the `newTestStore`/`mustBackdate`/`addWithTags`/`containsTitle`/
    `titlesOf` helpers) to model the new tests on.
  - `internal/cli/impact.go`, `internal/cli/story.go`, `internal/cli/wrapped.go`,
    `internal/cli/coverage.go` — the four Go-side upper-bound filters to remove.
  - `internal/cli/window.go` — `windowCutoff` (source of the `end` bound for
    impact/story/coverage; zero on the current-period path).
- **External APIs:** none.
- **Related code paths:** `internal/storage/`, `internal/cli/`.

## Outputs

- **Files modified:**
  - `internal/storage/entry.go` — add `Until time.Time` to `ListFilter` with a
    doc comment symmetric with `Since` (`entries.created_at < Until (RFC3339
    UTC, exclusive)`).
  - `internal/storage/store.go` — in `Store.List`, add after the `Since` block:
    ```go
    if !f.Until.IsZero() {
        conds = append(conds, "e.created_at < ?")
        args = append(args, f.Until.UTC().Format(time.RFC3339))
    }
    ```
  - `internal/cli/impact.go` — set `Until: end` on the `ListFilter` literal;
    delete the `if !end.IsZero() { … CreatedAt.Before(end) … }` loop (~125–139)
    and its comment block.
  - `internal/cli/story.go` — set `Until: end`; delete the same loop (~198–212).
  - `internal/cli/wrapped.go` — set `Until: nextBoundary` on the `ListFilter`
    literal; delete the always-applied `created_at < nextBoundary` loop
    (~217–227); collapse `all` → the `entries` result of `s.List` (the loop that
    built `entries` from `all` is gone).
  - `internal/cli/coverage.go` — set `Until: end`; delete the `if !end.IsZero()`
    filter loop (~121–134). **Keep** the `end`-based `upper`/`scopeMonths` label
    derivation below it (~136–145) untouched — `end` is still read there.
  - `internal/storage/store_test.go` — add `TestList_FilterByUntil` and
    `TestList_FilterBySinceAndUntil` (the failing tests).
- **New exports:** `storage.ListFilter.Until time.Time` (new field on an existing
  exported struct; no new function).
- **Database changes:** none — query-time filter only, no schema/migration.

### Premise audit (§9 additive + audit-grep cross-check)

This spec **adds a field** to a tracked fixed-shape struct (`ListFilter`), so the
§9 additive case applies: grep existing tests for literal-count / struct-shape
assertions on `ListFilter` that an added field would break. Greps run at design
and reconciled:

- `grep -rn "ListFilter{" internal/**/*_test.go` — every literal is **keyed**
  (`ListFilter{}`, `ListFilter{Tag: …}`, `ListFilter{Author: …}`, …). An added
  field defaults to its zero value in all of them, so none break at compile time
  and none change behavior. (~40 hits across `add_test.go`, `add_json_test.go`,
  `mcpserver/server_test.go`, `provenance_agreement_test.go`,
  `project_migration_test.go`; all keyed.)
- `grep -rn "NumField\|reflect.TypeOf(ListFilter" internal/` — **zero hits**. No
  test asserts the struct's field count or shape.
- `grep -rn "\.Until\|Until " internal/` — the only pre-existing hit is a
  `wrapped.go` *comment* ("ListFilter has no Until field"), which is deleted with
  its loop. No code references `Until` yet.

**Conclusion:** the additive change breaks no existing test (no planned deletions
or count-bumps). The behavior guardrails for the CLI refactor are existing tests
that must stay GREEN (not be rewritten) — enumerated under Acceptance Criteria.

## Acceptance Criteria

Testable outcomes. Cover happy path, error cases, edge cases.

- [ ] `storage.ListFilter` has an `Until time.Time` field documented as an
      exclusive RFC3339-UTC upper bound, symmetric with `Since`.
- [ ] `Store.List` applies `e.created_at < ?` when `!f.Until.IsZero()`, with the
      bound formatted `f.Until.UTC().Format(time.RFC3339)`; a zero `Until` adds
      no clause. (New: `TestList_FilterByUntil`.)
- [ ] The upper edge is EXCLUSIVE: an entry created exactly at `Until` is
      excluded; earlier entries are included. (New: `TestList_FilterByUntil`.)
- [ ] `Since` + `Until` compose into a half-open `[Since, Until)` window and AND
      with each other; boundary entries handled per each edge's inclusivity.
      (New: `TestList_FilterBySinceAndUntil`.)
- [ ] `impact`, `story`, `wrapped`, `coverage` set `filter.Until` (from `end` /
      `nextBoundary`) and no longer filter the upper edge in Go:
      `grep -rn "CreatedAt.Before" internal/cli/` returns nothing.
- [ ] All existing bounded-window CLI tests stay GREEN, unchanged (behavior
      preservation) — see the guardrail list below.
- [ ] All export goldens stay BYTE-IDENTICAL (they are fixture-fed; they never
      touch `Store.List`, so they cannot move).
- [ ] `gofmt -l .` clean; `go vet ./...` clean; `go test ./...` green;
      `just test-docs` green.

### Existing CLI guardrail tests (must stay GREEN — do NOT rewrite)

These pin the bounded-window behavior the refactor preserves. If any turns red,
the refactor changed behavior and must be corrected — they are the regression
contract, not failing tests to author:

- `internal/cli/impact_test.go`: `TestImpactCmd_PreviousQuarterBounded`,
  `TestImpactCmd_PreviousMonthAndYear`, `TestImpactCmd_PreviousYearBoundaryRoll`,
  `TestImpactCmd_NoPreviousUnchanged` (the current-period byte-for-byte guard),
  `TestWindowCutoff_PreviousBoundaries`.
- `internal/cli/story_test.go`: `TestStoryCmd_PreviousExplicitWindowBounded`,
  `TestStoryCmd_PreviousShiftsProfileDefault`.
- `internal/cli/wrapped_test.go`: `TestWrappedCmd_BoundedWindow` (the
  always-bounded upper-edge guard), `TestWrappedCmd_PreviousDefaultsToLastYear`,
  `TestWrappedCmd_PreviousQuarterlessStillYear`.
- `internal/cli/coverage_test.go`: `TestCoverageCmd_CalendarWindowAndScope`.

## Failing Tests

Written during **design**, BEFORE build. The implementer's job in **build** is to
make these pass. Both live in `internal/storage/store_test.go`, use
`newTestStore(t)` (which uses `t.TempDir()`, honoring `storage-tests-use-tempdir`)
and the `mustBackdate` helper, modeled on `TestList_FilterBySince` /
`TestList_FilterCombined`.

- **`internal/storage/store_test.go`**
  - `"TestList_FilterByUntil"` — asserts `ListFilter{Until: t}` returns only
    entries with `created_at < t`. Backdate three entries (e.g. `-3d`, `-1d`,
    and exactly at the bound), set `Until` to the middle boundary, assert the
    older entries are returned and the at/after-bound entries are excluded. **Key
    edge assertion:** an entry whose `created_at` is EXACTLY `Until` is EXCLUDED
    (proves `<`, not `<=`). Also assert `ListFilter{}` (zero `Until`) returns all
    rows (the `!IsZero()` no-op guard).
  - `"TestList_FilterBySinceAndUntil"` — asserts a bounded `[Since, Until)`
    window: backdate entries below `Since`, inside `[Since, Until)`, exactly at
    `Since` (INCLUDED — `>=`), exactly at `Until` (EXCLUDED — `<`), and above
    `Until`; assert `List` returns exactly the in-window set and that `Since` and
    `Until` AND-compose. Model composition on
    `TestList_AuthorComposesWithOtherFilters` / `TestList_FilterCombined`.

Illustrative shape (build may adjust names/values; the assertions above are the
contract):

```go
func TestList_FilterByUntil(t *testing.T) {
    s, path := newTestStore(t)
    now := time.Now().UTC()

    a := addWithTags(t, s, "old", "", "", "")      // -3d, before bound
    b := addWithTags(t, s, "recent", "", "", "")   // -1d, before bound
    c := addWithTags(t, s, "at-bound", "", "", "")  // exactly at bound → excluded
    d := addWithTags(t, s, "after", "", "", "")     // now, after bound → excluded

    bound := now.Add(-12 * time.Hour)
    mustBackdate(t, path, a.ID, now.Add(-3*24*time.Hour))
    mustBackdate(t, path, b.ID, now.Add(-1*24*time.Hour))
    mustBackdate(t, path, c.ID, bound) // created_at == Until → excluded (<, not <=)
    mustBackdate(t, path, d.ID, now)

    got, err := s.List(ListFilter{Until: bound})
    if err != nil {
        t.Fatalf("List: %v", err)
    }
    if len(got) != 2 || !containsTitle(got, "old") || !containsTitle(got, "recent") {
        t.Fatalf("want {old,recent}; got %v", titlesOf(got))
    }
    if containsTitle(got, "at-bound") {
        t.Errorf("entry at exactly Until must be EXCLUDED (< not <=); got %v", titlesOf(got))
    }
    // zero Until is a no-op: returns all.
    all, _ := s.List(ListFilter{})
    if len(all) != 4 {
        t.Errorf("zero Until should be a no-op; len=%d want 4", len(all))
    }
}
```

> **NOTE (expected at design):** these two tests reference `ListFilter.Until`,
> which does not exist yet, so `internal/storage` will not compile until build
> adds the field. That is the intended fail-first state (§9 / `test-before-
> implementation`); build confirms they fail for the authored reason (missing
> field, then the assertion) before implementing.

**No failing tests are authored for the CLI refactor itself.** It is
behavior-preserving, so its regression contract is the existing green CLI
guardrail tests enumerated above — duplicating them as new "failing" tests would
be redundant and they would not fail-first.

## Implementation Context

*Read this section (and the files it points to) before starting the build cycle.*

### Decisions that apply

- `DEC-035` — this promotion: add exclusive `!IsZero()`-guarded RFC3339-UTC
  `Until`, symmetric with `Since`; rewire the four consumers; zero `Until` is a
  no-op. The four locked points ARE the design here.
- `DEC-030` — `brag wrapped`'s bounded window; recorded the Go-filter tradeoff
  and the "promote to `ListFilter.Until`" revisit trigger this spec fires.
- `DEC-032` — `--previous`; `windowCutoff`'s `end` return is the bound
  impact/story/coverage feed into `Until` (zero on the current-period path,
  which is why the `!IsZero()` guard is load-bearing for byte-identical current
  behavior).

### Constraints that apply

- `no-sql-in-cli-layer` — the WHOLE POINT: the upper-bound filter moves INTO
  `internal/storage`. Do not import `database/sql` in `internal/cli`; the CLI
  only sets a struct field.
- `timestamps-in-utc-rfc3339` — the `Until` bound is written
  `f.Until.UTC().Format(time.RFC3339)`, identical to the `Since` block.
- `storage-tests-use-tempdir` — the new storage tests use `newTestStore(t)`
  (`t.TempDir()`).
- `errors-wrap-with-context` — no new error path is added (the `Until` block
  builds a clause; existing `list entries: %w` wrapping stays).
- `test-before-implementation` — the two storage tests are authored here and made
  to pass in build.

### Prior related work

- `SPEC-051` (shipped) — `brag wrapped`; the first `created_at < nextBoundary` Go
  filter this promotion replaces.
- `SPEC-053` (shipped) — `--previous`; added the impact/story filters and named
  this exact `ListFilter.Until` promotion as the deferred third-consumer
  follow-up in its ship reflection.
- `SPEC-045` (shipped) — `brag coverage`; the fourth Go filter.
- `SPEC-007` (shipped) — introduced `ListFilter` and the `Since` block this field
  mirrors.
- The DEC-004 → DEC-015 tag-model promotion — the canonical prior
  "deferred-then-executed-at-threshold" storage change; DEC-035 is its analogue.

### Out of scope (for this spec specifically)

- Any schema change, migration, or index on `created_at`.
- Any new bounded-window CONSUMER (e.g. `brag spark` — that is the separate next
  spec; it will simply set `Until` when it lands).
- Changing any window SEMANTICS: no new inclusive-upper variant, no change to
  `windowCutoff`/`parseWrappedPeriod` math, no change to the `scope` tokens.
- Any change to the export renderers or their fixtures/goldens.
- An inclusive `Until` (`<=`) — explicitly rejected in DEC-035 (Option C).

## Notes for the Implementer

- **Model `Until` on `Since`, verbatim.** In `Store.List`, place the `Until`
  block immediately AFTER the `!f.Since.IsZero()` block (~store.go:347–350).
  Order among AND-ed conditions does not matter to correctness, but keeping the
  two range bounds adjacent reads best. The `time` import is already present.
- **`entry.go` doc comment:** mirror the `Since` line, e.g.
  `Until   time.Time // entries.created_at < Until (RFC3339 UTC, exclusive upper bound)`.
- **The four consumers — exactly what to change:**
  - `impact.go`: `filter := storage.ListFilter{Since: cutoff}` →
    `storage.ListFilter{Since: cutoff, Until: end}`; delete the `if !end.IsZero()
    { … }` loop and its comment (~125–139). `end` comes from `windowCutoff`.
  - `story.go`: same — add `Until: end` to the `filter` literal (~150); delete
    the loop (~198–212). `end` comes from `resolveWindow` → `windowCutoff`.
  - `wrapped.go`: `filter := storage.ListFilter{Since: start}` →
    `{Since: start, Until: nextBoundary}` (~177); replace `all, err := s.List(
    filter)` + the `entries := make(...)` loop (~212–227) with
    `entries, err := s.List(filter)`. `nextBoundary` is ALWAYS non-zero (a named
    period is always bounded), so the guard fires every time — behavior
    unchanged.
  - `coverage.go`: add `Until: end` to `filter` (~78); delete the `if
    !end.IsZero()` filter loop (~121–134). **Do NOT touch** the `upper`/
    `scopeMonths` block (~136–145) — it legitimately still reads `end` for label
    derivation; only the FILTER loop goes.
- **After the refactor, `grep -rn "CreatedAt.Before" internal/cli/` must return
  nothing** — the cheap completeness check that all four copies are gone.
- **Why goldens can't move:** the export tests call the renderers
  (`ToImpactMarkdown`, wrapped/story/coverage renderers) with hand-built
  `[]storage.Entry` fixtures, never through `Store.List`. Moving the upper filter
  into `Store.List` is upstream of the fixtures, so the rendered bytes are
  identical. The CLI-level `Previous*`/`Bounded`/`NoPreviousUnchanged` tests
  (which DO run through `Store.List`) are the ones that prove the bounded slice is
  unchanged — keep them green, unchanged.
- **Fail-first:** after authoring the two storage tests and before implementing,
  run `go test ./internal/storage/...` and confirm they fail on the missing
  `Until` field / the exclusive-edge assertion (§12 build rule), not a stray
  error.

---

## Build Completion

*Filled in at the end of the **build** cycle, before advancing to verify.*

- **Branch:** feat/spec-056-listfilter-until
- **PR (if applicable):** https://github.com/jysf/bragfile000/pull/97
- **All acceptance criteria met?** yes
  - `ListFilter.Until time.Time` added with a doc comment symmetric with `Since`
    (exclusive RFC3339-UTC upper bound).
  - `Store.List` applies `e.created_at < ?` (bound `f.Until.UTC().Format(time.RFC3339)`)
    when `!f.Until.IsZero()`; a zero `Until` adds no clause. Verified by the new
    `TestList_FilterByUntil` (exclusive edge: entry at exactly `Until` excluded;
    zero-`Until` no-op returns all rows).
  - Half-open `[Since, Until)` composition + AND verified by `TestList_FilterBySinceAndUntil`.
  - Four consumers (`impact`, `story`, `wrapped`, `coverage`) now set `filter.Until`
    (`end` / `nextBoundary`) and their Go upper-bound loops are deleted;
    `grep -rn "CreatedAt.Before" internal/cli/` returns nothing.
  - All enumerated bounded-window CLI guardrail tests stay GREEN, unchanged (26
    tests incl. subtests). Export goldens byte-identical (`go test ./...` green).
  - `gofmt -l .` clean; `go vet ./...` clean; `go test ./...` green (764 passed);
    `just test-docs` green; `just test-hook` green.
- **New decisions emitted:**
  - `DEC-035` — ListFilter.Until storage promotion (emitted at design)
- **Deviations from spec:**
  - None. The two storage tests were already authored at design; build confirmed
    fail-first (unknown field `Until`) then made them pass. All four consumer
    refactors matched the spec's per-file change list exactly. In `wrapped.go` the
    result variable was renamed `all` → `entries` and the `entries := make(...)`
    filter loop deleted, as prescribed.
- **Follow-up work identified:**
  - None new. `brag spark` (the sketched fifth bounded consumer) will simply set
    `Until` when it lands — no re-copied Go loop — which is the intended payoff of
    this promotion.

### Build-phase reflection (3 questions, short answers)

Process-focused: how did the build go? What friction did the spec create?

1. **What was unclear in the spec that slowed you down?**
   — Nothing. The spec was fully mechanical: the four locked DEC-035 points, the
   exact per-file change list (with line ranges), the pre-authored failing tests,
   and the "model on `Since` verbatim" instruction left no design decision for
   build. The one thing worth double-checking — that coverage.go's `end`-based
   `upper`/`scopeMonths` block must NOT be touched — was called out explicitly, so
   I only deleted the filter loop.

2. **Was there a constraint or decision that should have been listed but wasn't?**
   — No. `no-sql-in-cli-layer`, `timestamps-in-utc-rfc3339`, `storage-tests-use-tempdir`,
   and `test-before-implementation` all applied and were listed. The refactor
   only removes Go from the CLI and adds one SQL clause in `internal/storage`, so
   the constraint set was exactly right.

3. **If you did this task again, what would you do differently?**
   — Nothing material. This is close to the ideal shape of a behavior-preserving
   refactor spec: the regression contract is a named set of existing green tests
   plus two new storage tests for the new field, so build is transcribe-and-verify.
   The `grep "CreatedAt.Before"` completeness check made "all four copies gone"
   trivially confirmable.

---

## Reflection (Ship)

*Appended during the **ship** cycle. Outcome-focused reflection, distinct
from the process-focused build reflection above.*

1. **What would I do differently next time?**
   — Nothing about the change itself — this was the near-ideal shape of a
   behavior-preserving refactor: a locked DEC, exact per-file line ranges,
   two pre-authored edge-pinning tests, an existing green-test regression
   contract, and a `grep -rn "CreatedAt.Before"` completeness check. The one
   process note is orchestration-level, not spec-level: the branch was
   stacked on SPEC-055's for race-free repo-global ID numbering and then
   rebased onto main to become a clean one-spec PR — pre-assigning IDs (as
   was done for the DEC files) or serializing scaffold creation avoids that
   rebase step next time.

2. **Does any template, constraint, or decision need updating?**
   — No. DEC-035 records the promotion and CLOSES the revisit trigger that
   DEC-030 and DEC-032 both deferred to exactly this threshold (the 4th
   Go-side consumer, well past rule-of-three) — the storage-query analogue of
   the DEC-004 → DEC-015 tag-model promotion. `no-sql-in-cli-layer` is now
   honored for the window's upper bound, not just its lower bound. No
   constraint or template text needs to change.

3. **Is there a follow-up spec I should write now before I forget?**
   — No new one. The next bounded-window consumer, `brag spark` (STAGE-016's
   other spec), simply sets `filter.Until` when it lands — it inherits this
   field instead of re-copying the Go loop, which is precisely the payoff.
