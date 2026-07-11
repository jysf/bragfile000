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

> **Reopened 2026-07-10 for the release cut.** This stage was marked `shipped`
> one step early — its own success criteria include "a v0.5.0 minor release cut
> ships the batch," so the cut (SPEC-067) is its closing action. Status is back
> to `active` (and `shipped_at: null`); the orchestrator re-closes it once
> v0.5.0 publishes (mirrors how SPEC-054 closed STAGE-013).

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

- [x] SPEC-056 (shipped on 2026-07-10) — `ListFilter.Until` storage promotion
      (+ DEC-035); removed duplicated Go-side upper-bound filtering across the 4
      consumers.
- [x] SPEC-059 (shipped on 2026-07-10) — `brag spark` sparklines-only pulse
      (Total + top-8 by-project) over rolling `--week|--month|--quarter`
      (default `--month`); new `aggregate.RollingBuckets` sub-month bucketer
      + DEC-037.
- [x] SPEC-060 (shipped on 2026-07-10) — fix `brag spark` upper-bound query:
      bound the corpus to the same `[start, now)` axis as the bucketer so the
      header count + top-8 exclude out-of-window entries (applies DEC-035).
- [x] SPEC-061 (shipped on 2026-07-10) — fix `brag project ensure` name cap to
      count bytes (not runes), restoring DEC-036 parity with the capture paths.
- [x] SPEC-062 (shipped on 2026-07-10) — fix SQLite concurrency: `busy_timeout`
      + `_txlock=immediate` + single connection so concurrent access WAITS
      instead of failing `database is locked` (+ DEC-038, WAL deferred).
- [x] SPEC-063 (shipped on 2026-07-10) — fix `brag tag rename`: canonicalize /
      reject the target (trim, reject comma/empty) so a rename can't silently
      corrupt membership on the next edit.
- [x] SPEC-064 (shipped on 2026-07-10) — harden capture input validation across
      all four ingress paths via one shared `internal/capture.Validate`
      (byte-caps + control-char rejection + reserved-numeric-tag checks).
- [x] SPEC-065 (shipped on 2026-07-10) — escape `|` as `\|` in markdown
      table cells so a pipe in a value keeps the row at two columns.
- [x] SPEC-066 (shipped on 2026-07-10) — treat a normal `brag mcp serve`
      shutdown (in-flight client close) as a clean exit (RC 0) instead of the
      RC-2 `server is closing: EOF` crash.
- [ ] SPEC-067 (verify) — v0.5.0 minor release cut: the stage's closing release
      action (CHANGELOG + version pin + pre-flight prep in one PR; the
      irreversible tag/publish follows from `main`).

**What shipped vs. what was cut (reconciled at stage ship, 2026-07-10):**
- SHIPPED — the two planned pieces (SPEC-056 `ListFilter.Until`, SPEC-059 `brag
  spark`) plus SPEC-060/061 (the batch's own MEDIUM fixes) and SPEC-062–066
  (a pre-release hardening pass: concurrency, tag-rename, capture-validation,
  markdown-escape, mcp-shutdown).
- SHIPPED as a doc chore — the `sprint:<id>` freeform-tag convention (sprint is
  just a tag; no schema, no reserved namespace).
- DROPPED — R5 milestone-write `type`: investigation found no auto-generated
  write path (the milestone notifier doesn't insert), and untyped entries are
  legitimately optional, so there was nothing to fix.
- DROPPED — WAL-safe backup doc note: `journal_mode=WAL` is not set and
  `backup.go` is already WAL-safe, so bare `cp` guidance is not unsafe — moot.
- DEFERRED — cosmetic `project status` trailing empty column when `state_note`
  is blank (not consumed by `standup`, which reads JSON).

**Count:** 9 shipped / 1 active (SPEC-067 release cut) / 0 pending

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

- **Did we deliver the outcome in "What This Stage Is"?** Yes. The two planned
  substrate-tidying pieces shipped (SPEC-056 `ListFilter.Until` storage
  promotion; SPEC-059 `brag spark`), and on top of them a full pre-release audit
  hardening pass landed six fixes across concurrency (SPEC-062), tag-rename
  canonicalization (SPEC-063), capture-input validation (SPEC-064),
  markdown-cell escaping (SPEC-065), and mcp-serve clean shutdown (SPEC-066),
  plus the batch's own two MEDIUMs (SPEC-060/061). The deeper agent-native work
  lands on a genuinely clean base.
- **How many specs did it actually take?** 9 shipped, vs. a plan that named 2
  real pieces + a loose "tier-1 micro-fix" bucket. The expansion was the audit.
- **What changed between starting and shipping?** The user asked for a bug check
  before the v0.5.0 release; that audit turned up pre-existing HIGH/MEDIUM issues
  (SQLite `database is locked`, silent tag-rename corruption, capture-validation
  drift) — several amplified by this batch's own MCP-first-class work — which
  became SPEC-062–066, and the planned R5 micro-fix was dropped when it proved
  to be a non-bug.
- **Lessons that should update AGENTS.md, templates, or constraints?**
  - Run the pre-release adversarial audit *before* cutting the release, not
    after — here it caught a HIGH concurrency bug the headline MCP server would
    have hit in normal use.
  - "Validate at the boundary, once" (SPEC-064's shared `internal/capture`
    package) is the pattern that ends per-path input-validation drift; worth
    reaching for whenever a new ingress is added.
- **Should any spec-level reflections be promoted to stage-level lessons?**
  - A shared cli-reachable canonicalizer for tag tokens (SPEC-063 follow-up)
    and running the edit path through `capture.Validate` (SPEC-064 follow-up)
    are the two open threads for a later polish/hardening spec.
