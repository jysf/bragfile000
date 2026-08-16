---
# Maps to ContextCore project.* semantic conventions.
# A project is a bounded wave of work against the repo (the app).

project:
  id: PROJ-009
  status: proposed                  # proposed | active | shipped | cancelled
  priority: low
  target_ship: null

repo:
  id: bragfile

created_at: 2026-08-16
shipped_at: null
---

# PROJ-009: Scale Baseline and the Harvested Harness

> **CANDIDATE — proposed, not opened.** Captured 2026-08-16 as the explicit home
> for work cut from PROJ-007 at framing, so the cut items have an address rather
> than living only in a table of deferrals. **Not committed work.** Priority is
> deliberately `low`: this is the least urgent of the three open candidates.

## What This Project Is

A **right-sized, time-boxed** answer to "does bragfile hold up at scale, and can
we prove it?" — plus the reusable measurement methodology harvested from
crustyimg rather than reinvented here.

The honest framing, inherited from PROJ-007's brief: **bragfile is a small local
SQLite CLI and raw performance is not its story and never will be.** The target
is *"measured, it holds"* — not *"blazing fast."* A project that drifts into
optimization has failed, not succeeded.

## Why Now

**Not now — that is the point of this file.** It exists so three cut items have
a home:

1. **The scale/perf + concurrency baseline harness**, cut from PROJ-007 at
   framing (2026-08-15) to keep that project short.
2. **The crustyimg methodology harvest**, which PROJ-007's brief listed as a
   dependency. It exists to serve the harness work, so it travels with it. The
   transferable part is *methodology* — harness structure, CI wiring, the
   measurement approach — **not** crustyimg's perf targets, which are
   image-processing numbers that do not transfer to a SQLite CLI.
3. Any perf question STAGE-022 surfaces while wiring lint and coverage.

**Open it when** one of these fires: a real corpus grows large enough that a
command feels slow; a concurrency bug appears in practice; or the portfolio
story genuinely needs a measured number rather than a described discipline.

Verified 2026-08-16: `func Benchmark` count in this repo is **0**.

## Success Criteria

*To be set at framing.* Likely themes:

- A small `testing.B` harness covering the operations that would actually
  degrade — FTS query time, startup, DB growth, and the memory slice's
  three-read `Gather` — at a corpus size well beyond today's (~10k entries
  against a live corpus of 368).
- A **concurrency stress check** that would have caught DEC-038's class of bug.
- Numbers **recorded with their conditions** (machine, corpus size, Go version)
  so a later run is comparable rather than merely newer.
- The methodology written down once, in a form the next repo can adopt.

## Scope

### In scope
- The benchmark/scale harness and its recorded baseline.
- A concurrency stress check.
- The crustyimg harvest: separate transferable methodology from
  crustyimg-specific perf targets, and adopt only the former.

### Explicitly out of scope
- **Optimization.** Measure first. If a number is bad, that is a finding and
  probably a separate spec — not licence to tune inside this project.
- Anything that would make performance bragfile's identity. It is not.
- Feature work of any kind.

## Stage Plan

*Proposed, not framed.*

- [ ] (not yet defined) — harvest the transferable methodology from crustyimg.
- [ ] (not yet defined) — the scale/benchmark harness and its recorded baseline.
- [ ] (not yet defined) — the concurrency stress check.

**Count:** 0 shipped / 0 active / 3 pending

## Dependencies

### Depends on
- **PROJ-007 / STAGE-022** — measure quality on a substrate whose lint and
  coverage are already wired and whose known-defect list is empty. Reading
  crustyimg's lint/coverage config as *reference* during STAGE-022 is explicitly
  **not** this harvest and needs no project.
- **crustyimg** — the source of the methodology. External to this repo.

### Enables
- A portfolio claim about scale that is measured rather than asserted.
- Confidence for any future work that would grow the corpus substantially — the
  git-import cold-start miner being the obvious one, since it would take a
  corpus from hundreds to thousands in a single command.

## Project-Level Reflection

*Filled in when status moves to shipped.*

- **Did we deliver the outcome in "What This Project Is"?** <yes/no + notes>
- **How many stages did it actually take?** <number, compare to plan>
- **What changed between starting and shipping?** <one or two sentences>
- **Lessons that should update AGENTS.md, templates, or constraints?**
  - <one-line updates>
- **What did we defer to the next project?**
  - <one-line items>
