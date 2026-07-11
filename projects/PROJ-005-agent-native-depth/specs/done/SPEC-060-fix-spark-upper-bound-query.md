---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-060
  type: story                      # epic | story | task | bug | chore
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
  architect: claude-opus-4-7
  implementer: claude-opus-4-7     # usually same Claude, different session
  created_at: 2026-07-10

references:
  decisions: [DEC-035, DEC-037]
  constraints: [no-sql-in-cli-layer, timestamps-in-utc-rfc3339, stdout-is-for-data-stderr-is-for-humans]
  related_specs: [SPEC-056, SPEC-059]
---

# SPEC-060: fix spark upper-bound query

## Context

`brag spark` (SPEC-059/DEC-037) queries its in-window corpus with
`storage.ListFilter{Since: now - width*n}` and **no upper bound**. A
pre-release adversarial sweep found that only the bucket sums
(`aggregate.RollingBuckets`) honour the `[start, now)` axis — the raw
`entries` slice feeds two other consumers that do NOT:

- the markdown header `Entries: %d` = `len(entries)`
  (`internal/export/spark.go`), and
- the project ranking/top-8 selection `aggregate.ByProject(entries)`
  (same file).

So an entry with `created_at >= now` (clock skew, a second machine, or an
imported row) inflates the header count and can occupy a phantom all-`▁`
top-8 slot, evicting a real in-window project. The design comment claiming
"the bucketer's `end=now` is the effective upper edge" is true only for the
sums. This is exactly the "fifth bounded consumer" DEC-035 anticipated when
it promoted `ListFilter.Until` to the storage layer. Part of `STAGE-016`
(polish) under `PROJ-005`.

## Goal

Bound the `brag spark` query to the same half-open `[start, now)` axis the
bucketer uses, so the header count, the top-8 by-project selection, and the
bucket sums all describe exactly the in-window corpus and out-of-window
entries are excluded everywhere.

## Inputs

- **Files to read:** `internal/cli/spark.go` (`runSpark`) — the unbounded
  query; `internal/aggregate/rolling.go` — confirm the bucket axis is
  `start = end - width*n`, `[start, end)`; `internal/storage/entry.go` —
  `ListFilter.Until` (exclusive `created_at < Until`).
- **Related code paths:** `internal/export/spark.go` (the two unbounded
  consumers — read-only; no change needed once the query is bounded).

## Outputs

- **Files modified:**
  - `internal/cli/spark.go` — sample `now` once and truncate to the second;
    compute `start` once; set BOTH `Since: start` and `Until: now` on the
    filter; pass the same `now` into `SparkOptions.Now` so query and buckets
    share one axis.
  - `internal/cli/spark_test.go` — add the failing-first regression test.
- **New exports:** none.
- **Database changes:** none.

## Acceptance Criteria

- [x] With one in-window entry and one future-dated entry, the markdown
  header shows `Entries: 1` (not 2).
- [x] The future entry's project does NOT appear as a by-project row.
- [x] The JSON `total.count` and `sum(total.series)` both equal the header
  count (1).
- [x] `now` is sampled once, truncated to the second, and feeds both the
  filter bounds and the bucketer (single axis).
- [x] No SQL in the CLI layer (query via `ListFilter`); timestamps UTC
  RFC3339; existing spark behaviour unchanged (all prior spark tests green).

## Failing Tests

Written during **design**, BEFORE build.

- **`internal/cli/spark_test.go`**
  - `TestSparkCmd_ExcludesOutOfWindowEntries/markdown-header-and-rows` —
    asserts header `Entries: 1`, the in-window project row present, the
    future project row absent. Fails pre-fix (header shows 2, phantom row
    present).
  - `TestSparkCmd_ExcludesOutOfWindowEntries/json-total-sum-matches-count` —
    asserts `total.count == sum(total.series) == 1` and no `phantom` entry
    in `by_project`.

## Implementation Context

*Read this section (and the files it points to) before starting
the build cycle. It is the equivalent of a handoff document, folded
into the spec since there is no separate receiving agent.*

### Decisions that apply

- `DEC-035` — `ListFilter.Until` (exclusive `created_at < Until`) lives in the
  storage layer precisely so bounded consumers just set the field; spark is
  the anticipated "fifth bounded consumer".
- `DEC-037` — brag spark's rolling window is half-open `[start, now)`; the
  query axis must match the bucket axis, not merely overlap it.

### Constraints that apply

These constraints apply to the paths touched by this task (see
`/guidance/constraints.yaml` for full text):

- `no-sql-in-cli-layer` — the fix stays in `runSpark`, bounding via
  `storage.ListFilter`; no SQL enters the CLI.
- `timestamps-in-utc-rfc3339` — `now` is UTC, truncated to the second so the
  boundary matches storage's RFC3339 second-precision comparison.
- `stdout-is-for-data-stderr-is-for-humans` — unchanged; the pulse stays on
  stdout.

### Prior related work

- `SPEC-056` (shipped) — promoted `ListFilter.Until` to storage (DEC-035).
- `SPEC-059` (shipped) — brag spark itself (the code being fixed).

### Out of scope (for this spec specifically)

- No change to `internal/export/spark.go`: once the corpus is bounded, both
  `len(entries)` and `aggregate.ByProject` are correct by construction.
- No new flags, no calendar-window behaviour, no visual/sparkline changes.

## Notes for the Implementer

Sample `now` once, truncate to the second, and thread the SAME instant into
both the filter (`Since`/`Until`) and `SparkOptions.Now` — a single axis is
the whole point. Confirm `RollingBuckets` computes `start = end - width*n`
identical to the query `Since`.

---

## Build Completion

*Filled in at the end of the **build** cycle, before advancing to verify.*

- **Branch:** `fix/spec-060-spark-until-bound`
- **PR (if applicable):** see PR opened against `main`.
- **All acceptance criteria met?** yes
- **New decisions emitted:**
  - none — this applies DEC-035 (the anticipated bounded consumer) rather
    than introducing new policy.
- **Deviations from spec:**
  - none.
- **Follow-up work identified:**
  - none. The two export-layer consumers (`len(entries)`, `ByProject`) are
    now correct because the corpus itself is bounded; no export change was
    needed.

### Build-phase reflection (3 questions, short answers)

1. **What was unclear in the spec that slowed you down?**
   — Nothing; the fix was fully specified. The one thing worth verifying
   was that `SparkOptions.Now` already flowed into `RollingBuckets`, so
   passing the single truncated `now` there closed the axis with no export
   change.

2. **Was there a constraint or decision that should have been listed but wasn't?**
   — No. DEC-035 (Until) and DEC-037 (spark) plus the three constraints
   covered it. Truncating `now` to the second matters because storage
   compares `created_at` at RFC3339 second precision — worth an inline note.

3. **If you did this task again, what would you do differently?**
   — Nothing material; write the fail-first test, confirm the two authored
   failure reasons, then bound both ends of the filter.

---

## Reflection (Ship)

*Appended during the **ship** cycle. Outcome-focused reflection, distinct
from the process-focused build reflection above.*

1. **What would I do differently next time?**
   — Nothing material — a small, well-scoped fix. Threading one truncated
   `now` into both the `[start, now)` filter and the bucketer (a single axis)
   is what made the header count, top-8, and bucket sums agree; it landed
   clean with no export-layer change.

2. **Does any template, constraint, or decision need updating?**
   — No. This applies DEC-035 (`ListFilter.Until`) and DEC-037 (spark's
   half-open window) rather than introducing policy; nothing to update.

3. **Is there a follow-up spec I should write now before I forget?**
   — None. DEC-035 named this the anticipated "fifth bounded consumer" and no
   further unbounded spark consumers remain.
