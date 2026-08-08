---
# Maps to ContextCore project.* semantic conventions.
# A project is a bounded wave of work against the repo (the app).

project:
  id: PROJ-007
  status: proposed                  # proposed | active | shipped | cancelled
  priority: medium
  target_ship: null

repo:
  id: bragfile

created_at: 2026-08-07
shipped_at: null
---

# PROJ-007: Quality, Performance & Portfolio Readiness

> **CANDIDATE — proposed, not opened.** Captured 2026-08-07 while PROJ-006 is the
> active project. This exists so the idea is on record; it is not committed work.
> Open it after PROJ-006's depth pillars land (or when portfolio-readiness itself
> becomes the priority). Treat everything below as scope input, not a plan.

## What This Project Is

A cross-cutting **quality + presentation** pass, framed by one fact: **bragfile
is a portfolio project.** Where PROJ-001–006 built and deepened features, this
project makes the engineering quality *measured, enforced, and legible to an
outside reader* (a hiring manager, a collaborator, future-you).

**The portfolio reframe (why this differs from crustyimg's quality work).**
crustyimg's quality axis is **performance** — benchmarking is central because the
product *is* image-processing speed. bragfile is a small local SQLite CLI; raw
perf is not its story and never will be. bragfile's portfolio value is
**engineering judgment made legible**: the DEC log (39 decisions, zero
deprecated), the spec-driven process, byte-exact golden discipline, security-first
distribution, and honest incident→hardening handling (DEC-021 backup, DEC-038
concurrency, the v0.5.2 tap-token saga). Most repos show none of that. So this
project is **measure + surface + close the visible gaps** — not remediation.

## Why Now

Baseline quality is already strong but **invisible** to a stranger, and a few
standard signals a reviewer greps for are missing (verified in code 2026-08-07):
- **No golangci-lint** — CI runs only `gofmt` + `go vet` (AGENTS.md §3 even notes
  lint is "welcome but not required yet").
- **No coverage measurement** — no `-cover`, no badge (despite 66 test files to
  60 source files, byte-exact goldens, and a 116-assert doc-test harness).
- **Zero benchmarks** — `func Benchmark` count = 0; no scale/perf baseline.
- **No surfaced godoc / API docs** and no "engineering practices" entry point.

## Success Criteria

*To be set at framing.* Likely themes:
- golangci-lint configured and **gating in CI**; coverage measured with a
  reported number/badge and an honest floor.
- A right-sized **scale/perf baseline** (FTS query time, startup, DB growth,
  concurrency stress at ~10k entries) — measured and documented, not optimized to
  death.
- The existing discipline **made discoverable** to an outside reader (a README
  "engineering practices" section pointing at the DEC log, golden-test discipline,
  security posture, incident→hardening pattern).

## Scope

### In scope
1. **Close the standard visible signals (highest ROI):** golangci-lint + config +
   CI gate; coverage measurement + badge; a godoc/API-doc pass.
2. **Surface the discipline that already exists:** an "engineering practices" doc
   / README section; make the DEC log, golden-test regime, and security posture
   legible to a first-time reader.
3. **A right-sized system-quality pass (lower ROI — time-box it):** a small
   `testing.B` benchmark/scale harness + a concurrency stress check; the bragfile
   analog of crustyimg's benchmarking, but a modest story — "measured, it holds,"
   not "blazing fast."

### Explicitly out of scope
- Heavy performance optimization / deep benchmarking beyond a baseline — that's
  crustyimg's game, not bragfile's. Do NOT let this project turn into a perf
  project.
- Any feature work (that's PROJ-006's depth pillars).

## Stage Plan

*Proposed, not framed. Sequence/split at framing.*
- [ ] (STAGE-A?) — lint + coverage in CI (golangci-lint config, `-cover`, badge).
- [ ] (STAGE-B?) — discipline-surfacing docs (engineering-practices section +
      godoc pass).
- [ ] (STAGE-C?) — scale/perf + concurrency baseline harness (time-boxed).

## Dependencies

### Depends on
- **Harvest crustyimg's methodology first.** crustyimg already did significant
  benchmarking + code analysis; that is a *reusable methodology* (harness
  structure, CI wiring, coverage/lint/static-analysis approach). Route it through
  the framework harvest so this project adopts it instead of re-inventing — see
  the crustyimg harvest handoff (2026-08-07). Separate the transferable parts
  (harness/CI/coverage/lint discipline) from the crustyimg-specific perf targets
  (image-processing numbers won't transfer).
- A stable substrate — ideally PROJ-006's depth pillars mostly landed, so quality
  is measured on the finished shape.

### Enables
- A **portfolio-grade** bragfile: the repo demonstrates senior-level engineering
  judgment legibly, and the quality pass itself produces strong outward-impact /
  `wrapped`-arc material (metrics, before/after, retrospective).

## Project-Level Reflection

*Filled in when status moves to shipped.*

- **Did we deliver the outcome in "What This Project Is"?** <yes/no + notes>
- **How many stages did it actually take?** <number, compare to plan>
- **What changed between starting and shipping?** <one or two sentences>
- **Lessons that should update AGENTS.md, templates, or constraints?**
  - <one-line updates>
- **What did we defer to the next project?**
  - <one-line items>
