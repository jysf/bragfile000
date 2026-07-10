---
# Maps to ContextCore epic-level conventions.
# A Stage is a coherent chunk of work within a Project.
# It has a spec backlog and ships as a unit when the backlog is done.

stage:
  id: STAGE-016
  status: active
  priority: medium
  target_complete: null

project:
  id: PROJ-005
repo:
  id: bragfile

created_at: 2026-07-10
shipped_at: null
---

# STAGE-016: v0.4.x polish

## What This Stage Is

Clears the small debts and read-surface gaps the story-surface wave
(PROJ-004) left behind, so the deeper agent-native work lands on a clean
base. Two real pieces — promoting the duplicated calendar-window upper
bound into storage (`ListFilter.Until`) and a lightweight `brag spark`
pulse — plus a handful of verified tier-1 micro-fixes. Nothing here is
"agent-native depth"; it is the substrate-tidying that legitimately opens
a wave whose own synthesis names *completeness* and *the read surface* as
preconditions.

## Why Now

PROJ-004 shipped with two explicit follow-ups captured as backlog: the
`created_at < end` upper-bound filter is now duplicated Go-side across
four commands (`impact`, `story`, `wrapped`, `coverage`) — well past the
rule-of-three "promote to storage" trigger DEC-030/DEC-032 flagged — and
a quick sparkline-only read (`brag spark`) was sketched but deferred. Both
are small, self-contained, and best done before deeper features build more
consumers on the current shape. The backlog scan also surfaced verified
micro-fixes (milestone `type` null diluting by-type analytics; cosmetic/
doc corrections) cheap enough to batch alongside.

## Success Criteria

- `ListFilter.Until` exists in storage (SQL `created_at < ?`, guarded by
  `!Until.IsZero()`), the three/four commands use it, and the duplicated
  Go-side filtering is gone — with existing goldens BYTE-IDENTICAL and CLI
  behavior tests still green.
- `brag spark` prints a sparklines-only pulse (Total + by-project) over a
  recent window, reusing `internal/spark` + `internal/aggregate`, in the
  DEC-014 envelope, markdown default + JSON raw-counts (DEC-031).
- Verified tier-1 micro-fixes land without back-migrating historical data.
- A v0.5.0 minor release cut ships the batch (new commands → minor).

## Scope

### In scope
- **SPEC-056 — `ListFilter.Until` storage promotion** (+ a storage-layer
  DEC, DEC-035): add `Until` to `storage.ListFilter`, refactor `impact`/
  `story`/`wrapped`/`coverage` off Go-side upper-bound filtering (coverage is
  the confirmed 4th consumer). Behavior-preserving; goldens are fixture-fed so
  stay byte-identical. (IDs are assigned at creation: this refactor's actual id
  is SPEC-056; the `brag spark` item's actual id is assigned when created next.)
- **`brag spark`**: sparklines-only pulse for a recent window,
  Total + by-project rows. Design must resolve the real forks (no `--week`
  in the calendar core; no sub-month bucketer; new multi-row render).
- **Tier-1 micro-fixes** (fold in as small specs as they fit one-spec-
  per-PR): milestone-write `type` null (R5), `project status` trailing
  empty column when `state_note` blank, WAL-safe backup doc note, and the
  documented `sprint:<id>` tag convention.

### Explicitly out of scope
- The deferred `stats` cadence sparkline — needs a new lifetime-cadence
  data slot + a DEC; not the visual-only change it looks like. Defer.
- Any new schema column (including `sprint` — stays a freeform tag).
- Anything requiring network or CGO.

## Spec Backlog

Format: `- [status] SPEC-ID (cycle) — one-line summary`

- [ ] SPEC-056 (design) — `ListFilter.Until` storage promotion (+ DEC-035);
      remove duplicated Go-side upper-bound filtering across the 4 consumers.
- [ ] `brag spark` (frame) — sparklines-only pulse (Total + by-project) over
      `--week|--month|--quarter`. (Id assigned at creation — next free SPEC-*.)
- [ ] (candidate) — milestone-write `type` (R5); `project status`
      trailing-column cosmetic; WAL-safe backup doc + `sprint:` tag
      convention note. Split per one-spec-per-PR at frame time.

**Count:** 0 shipped / 0 active / 2 pending (+ micro-fix candidates)

## Design Notes

- **`ListFilter.Until` is behavior-preserving.** Model the `Until` block
  on the existing `Since` block in `Store.List` (RFC3339 UTC, guarded by
  `!f.Until.IsZero()` so the current-period zero-`end` path stays a
  no-op). `impact`/`story` source the bound from `windowCutoff`'s `end`;
  `wrapped` sources `nextBoundary` from `parseWrappedPeriod` — different
  upstream helpers, same new field. Export goldens are FIXTURE-FED (they
  never touch `Store.List`) so they cannot change bytes; the guardrails
  are the CLI-level bounded-window tests. Needs a storage-layer DEC per
  DEC-032's revisit note.
- **`brag spark` has genuine design forks** to lock at design, not build:
  (1) the calendar core (`window.go`) has month/quarter/year/since but NO
  `week` — either extend it or adopt `review`'s rolling-7-day semantics
  (and reconcile that `--month` means *calendar month* here but *last 30
  days* in `review`); (2) `aggregate.Cadence`/`CoverageByMonth` are
  monthly-only — a sub-month (daily/weekly) bucketer is new; (3) "Total +
  by-project rows of sparklines" is a new render shape (bucket each
  project over the same shared label axis so rows align). Copy
  `internal/cli/coverage.go` as the structural template; reuse the shared
  `lookupSparkEnv` var (do not redeclare). JSON stays raw counts, no
  glyphs (DEC-031).
- **Micro-fixes leave history untouched** — e.g. milestone `type` fix does
  not back-migrate the 42 untyped historical rows.

## Dependencies

### Depends on
- PROJ-004 (STAGE-013) — shipped `internal/spark`, the calendar-window
  infra, and the DEC-014 envelope this stage reuses; spawned both the
  `Until` and `spark` follow-ups as backlog.

### Enables
- A clean substrate for the deeper agent-native stages (memory,
  provenance, benchmark) to build on without inheriting duplicated
  filtering or a missing quick-read surface.

## Stage-Level Reflection

*Filled in when status moves to shipped.*

- **Did we deliver the outcome in "What This Stage Is"?** <yes/no + notes>
- **How many specs did it actually take?** <number vs. plan>
- **What changed between starting and shipping?** <one sentence>
- **Lessons that should update AGENTS.md, templates, or constraints?**
  - <one-line updates>
- **Should any spec-level reflections be promoted to stage-level lessons?**
  - <one-line items>
