---
# Maps to ContextCore project.* semantic conventions.
# A project is a bounded wave of work against the repo (the app).

project:
  id: PROJ-007
  status: shipped
  priority: medium
  target_ship: null

repo:
  id: bragfile

created_at: 2026-08-07
activated_at: 2026-08-15
shipped_at: 2026-08-23
---

# PROJ-007: Quality, Performance & Portfolio Readiness

> **OPENED 2026-08-15**, after PROJ-006 closed as a scope reduction. Captured as
> a candidate 2026-08-07; the scope below has been narrowed at framing, not
> adopted wholesale — see **Stage Plan** for what was cut and where it went.
>
> **Deliberately kept short.** PROJ-007 is a two-stage quality-and-legibility
> pass, not a wave. The feature work it might have absorbed lives in
> **PROJ-008** (the mirror, `brag learn`, story-surface v2); the scale/perf
> harness and the crustyimg methodology harvest live in **PROJ-009**. The point
> of the short scope is room for what comes up.

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

*Framed 2026-08-15. Two stages, docs first — the candidate ordering was
deliberately inverted.*

- [x] **STAGE-021 — make the discipline legible** *(shipped 2026-08-20)*.
      Three specs against a plan of two — SPEC-081 was created by SPEC-079's own
      LD5, which pinned the README guard tight rather than widening it. All six
      success criteria met and checkable: 15/15 packages carry a rendered
      `go doc` comment, `decisions/` has no unexplained gap, questions are 6 open
      of 18 with every one carrying a resolve condition, and every current-state
      number on the practices page is derived by a script and diffed by a test.
      Honest limit recorded in the reflection: **the guards reach numbers, not
      claims** — eleven verify findings across three specs, and the ones that
      mattered were prose contradicted by the path it cited, invisible to `X3`
      and `X5` both.

      *Delivered:* the engineering-practices entry point, a godoc pass, and three
      legibility repairs (`SECURITY.md`, the `DEC-041` reservation tombstone,
      open-questions hygiene) — plus the README call-to-action and the A1 metric
      switch that SPEC-079's own guard made necessary.
- [x] **STAGE-022 — quality measured and enforced, known defects closed**
      *(shipped 2026-08-23)*. **Three specs against a plan of two**, the second
      consecutive stage to do so — and in both cases the third was generated by
      the second *refusing to widen its own scope*: SPEC-081 by SPEC-079's LD5,
      SPEC-083 by SPEC-082's LD7. That is the loop working, not estimation
      drift. All four success criteria met and checkable: nine chosen linters
      gate CI at **0 issues**; coverage is **83.5%** against an enforced
      **80.0%** floor; the coverage claim was mutation-checked in both
      directions; and `brag memory` no longer calls its candidate pool by the
      word five sibling exporters use for entries-in-scope.

      *Delivered:* `SPEC-082` (lint + coverage gating CI, `.golangci.yml`,
      `scripts/coverage.sh`), `SPEC-083` (`DEC-047` — the **first amendment to
      `guidance/constraints.yaml` in the repo's history**), and `SPEC-084`
      (`DEC-048`, `Entries:` → `Candidates:`, resolving a collision between four
      earlier decision records).

      **One premise the stage overturned:** *"What This Stage Is"* promised a
      coverage **badge**. SPEC-082 Fork 3 decided against one, argued from this
      repo's own derived-not-typed rule — a badge is a *cached* number served by
      a third party. Recorded as overturned rather than quietly dropped.

      Honest limit, confirming STAGE-021's rather than resolving it: **a guard
      that is green is not evidence its claim is true.** `AA3` is NOT-contains
      only, `Y1` asserts a package comment *exists* rather than what it says,
      and `X3` absorbed two mispredicted rows without complaint. Every one
      behaved correctly; none would have caught the thing it sat beside. golangci-lint gating CI, coverage with an honest floor, and
      **one** remaining correctness item — the `Entries:` envelope
      inconsistency. *(Narrowed 2026-08-18: `MergeTags` position density and
      `$EDITOR` quoting were measured, found to have no current victim, and
      deferred to `PROJ-001/backlog.md` with evidence and real triggers.)*

      **It also inherits six items routed from STAGE-021**, each with evidence:
      the missing `internal/cli` audit test for the `no-sql-in-cli-layer`
      constraint — a `severity: blocking` rule with **no test for the package
      its own glob names**, with `depguard` named as the mechanism and the five
      colliding test files listed; the deferred totality assertion for the
      decision/reservation counts; three stale comments (`store.go`'s `Store`
      type comment, false since 2026-04-20; `list_test.go:275`; V1's
      package-vs-layer noun); and two from SPEC-081 (A1's "smallest deep-dive
      doc" over-reach, and A11's comment describing an A1 that no longer
      measures lines).

**Why docs first, against this brief's own ROI ranking.** Lint and coverage are
table stakes — having them proves the author knows the convention. The
differentiated asset is the discipline that is invisible from outside:
mutation-checked tests, mechanical guards that replaced remembered ones,
decision records that log their own corrections, reflections that name what was
wrong. There is also a sequencing reason: the practices page forces an honest
inventory of what the tests actually pin, and that inventory is what a coverage
number should be presented against. The reverse order invites a percentage with
no story — the exact failure SPEC-073 hit four times.

**Cut from the candidate scope, and where it went:**

| Cut | Where |
|---|---|
| Scale/perf + concurrency baseline harness | **PROJ-009** |
| The crustyimg methodology harvest | **PROJ-009** — it exists to serve the harness work, so it travels with it. Reading crustyimg's lint/coverage config as *reference* while framing STAGE-022 is not the harvest and needs no project. |
| Story-surface v2 | **PROJ-008**, with the mirror and `brag learn` |
| `CODE_OF_CONDUCT.md` | Declined. Boilerplate for a single-maintainer project, and adding it is the cargo-culting this repo's story is that it does not do. |
| `govulncheck` CI step | Left in `PROJ-001/backlog.md`. Its 2026-04-26 deferral (redundant with Dependabot; marginal noise reduction) was re-weighed under the portfolio lens on 2026-08-15 and **still holds** — visibility does not justify a second scanner when the first one already fires on every advisory. Revisit conditions unchanged. |

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

- **Did we deliver the outcome in "What This Project Is"?** **Yes — as an
  explicit scope reduction, not a completion.** Two of the three themes framing
  named are delivered: the discipline is discoverable to an outside reader
  (STAGE-021), and the standard signals are measured and *gating* rather than
  merely available (STAGE-022). The third — a scale/perf and concurrency
  baseline at ~10k entries — was **cut at framing on 2026-08-15 and routed to
  PROJ-009**, before any work began. Closing on two of three is therefore the
  plan executing, not the plan slipping. This is the second consecutive project
  to close this way (PROJ-006, 2026-08-15), and the pattern is deliberate:
  narrow at framing, ship what was narrowed to, and name where the rest went.

- **How many stages did it actually take?** **Two, exactly as planned** —
  though each ran **three specs against a plan of two**, so six specs against a
  plan of four. Both overruns came from the same mechanism, and it is worth
  naming because it looks like drift and is not: in each stage the third spec
  was **generated by the second refusing to widen its own scope.** SPEC-079's
  LD5 pinned the README guard tight rather than loosening it to stay green,
  which created SPEC-081. SPEC-082 mechanised a blocking constraint and refused
  to amend the rule it was mechanising, which created SPEC-083. A process that
  produces follow-on specs instead of silently absorbing them is the one you
  want; the estimate is what should adjust.

- **What changed between starting and shipping?** The project opened intending
  to *make existing quality legible* and discovered, twice, that the things it
  was making legible **were not true yet** — a blocking constraint with no
  automated guard for the package its own path glob named, and a document
  header that had meant the wrong thing across four decision records. Writing
  the claims down honestly is what surfaced them; the legibility pass became a
  correctness pass on the way.

- **Lessons that should update AGENTS.md, templates, or constraints?**
  - **Promoted at the STAGE-022 close:** the mutation protocol at **N=2**
    (confirm by content hash, restore from a `/tmp` backup — §12), and §9
    half-(b)'s third clause at **N=4** (*don't reason about which derived rows
    move; regenerate and diff*).
  - **Promoted at the STAGE-021 close:** *when a literal caches derived output,
    the derivation outranks the cache* (§9, N=3).
  - **Landed in the template rather than AGENTS.md:** the stage template gained
    a **stage-scoped** *"what can a user do now"* question, and both it and the
    spec-level question now say *capture it before closing the cycle*. This
    project is where the gap showed: a spec answered the question well and no
    entry was created for two days, until a hand-written scratch note prompted
    one.
  - **Held deliberately, not promoted:** `Y3` and `X3` pin the same numbers by
    different means, and `Y3` has now cached a value the shipping spec moves for
    **four consecutive specs**. The fix is a change to `scripts/test-docs.sh`'s
    assertion design — spec work, not a one-line rule — so it is routed rather
    than codified.

- **What did we defer to the next project?**
  - **PROJ-009** — the scale/perf and concurrency baseline harness at ~10k
    entries, and the crustyimg methodology harvest that serves it. Cut at
    framing, not abandoned.
  - **PROJ-008** — story-surface v2, the mirror, and `brag learn`. Also the
    natural home for **dynamic shell completion** (`--project` / `--tag` /
    `--type`): `brag completion` ships and works, but the only registration is
    `ValidArgs` on the `completion` command itself, so no flag *values*
    complete.
  - **Routed, unowned by a project, each with evidence in STAGE-022's Design
    Notes:** the unguarded `"four"` numeral in `root.go:13` / `.golangci.yml:71`
    (the fix is to drop the numeral; `DEC-047`'s T1 is *not* its guard);
    `store.go:11`'s package-vs-layer noun, which `DEC-047` made mechanically
    decidable; and the `goreleaser` `brews` deprecation **together with** the
    Homebrew formula not installing shell completions — same file, one spec.
  - **Declined outright, recorded:** `CODE_OF_CONDUCT.md` (boilerplate for a
    single-maintainer project) and a `govulncheck` CI step (redundant with
    Dependabot; re-weighed under the portfolio lens on 2026-08-15 and the
    deferral still held).
