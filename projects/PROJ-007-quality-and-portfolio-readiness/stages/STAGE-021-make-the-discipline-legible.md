---
# Maps to ContextCore epic-level conventions.
# A Stage is a coherent chunk of work within a Project.
# It has a spec backlog and ships as a unit when the backlog is done.

stage:
  id: STAGE-021
  status: active                    # proposed | active | shipped | cancelled | on_hold
  priority: high
  target_complete: null

project:
  id: PROJ-007
repo:
  id: bragfile

created_at: 2026-08-15
shipped_at: null
---

# STAGE-021: make the discipline legible

## What This Stage Is

A stranger who lands on this repository can find the engineering discipline that
already exists. **Not writing documentation — building an index to work that has
already been done.**

The gap is not quality, it is discoverability. Counted 2026-08-15, the repo
holds **45 decision records, 75 archived specs, a pre-distribution security
review, a three-project cross-repo retrospective, two blog posts, four
framework-feedback documents and four research syntheses** — behind a
**1,186-word README with no entry point to any of it.** Every artifact a
reviewer would find persuasive is present and unreachable.

This stage ships when the evidence is one click from the front door.

## Why Now

**Because it is the differentiated half, and PROJ-007's other stage is not.**
Lint and coverage are table stakes: every competent Go repo has them, and having
them proves only that the author knows the convention. What this repo can show
that almost none can is *mutation-checked tests, mechanical guards that replaced
remembered ones, decision records that log their own corrections, and reflections
that name what was wrong.* A hiring manager can get a coverage badge anywhere.

There is also a sequencing argument: **the docs stage tells you what the lint
stage should claim.** Writing the practices page forces an honest inventory of
what the test regime actually guarantees, and that inventory is what a coverage
number should be presented against. Doing it in the other order invites a
percentage with no story attached — and this repo has already been bitten four
times by a coverage sentence that sounded true and pinned nothing (SPEC-073).

## Success Criteria

- A reader arriving at the README can reach the DEC log, the spec archive, the
  byte-exact test regime, the security posture and the incident→hardening pattern
  **without knowing the repo's directory layout.**
- Every claim on the practices page is **backed by a countable artifact** — a
  path, a count, a named test — not an adjective. No "rigorous", no
  "comprehensive"; if it cannot be pointed at, it does not go on the page.
- `go doc` on the public surface reads as intentional rather than as leftover
  comments.
- `decisions/` has no unexplained numbering gap.
- `guidance/questions.yaml` describes **live uncertainty only** — every entry is
  either genuinely open with a stated resolve condition, or closed.
- The counts on the page are **derived, not typed** (or pinned by `test-docs`),
  so they cannot rot the way the README version line did through two releases.

## Scope

### In scope

1. **An engineering-practices entry point.** A README section plus a document it
   points at, covering: the DEC log and how decisions get corrected rather than
   silently replaced; the spec-driven cycle; byte-exact golden discipline;
   the `test-docs` assertion harness (W1–W4 replaced four previously-remembered
   checks with mechanical ones); the security posture; and the
   incident→hardening pattern (DEC-021 backup, DEC-038 concurrency, the v0.5.2
   tap-token saga, the cask→formula upgrade cliff).
2. **A godoc pass** over the public surface — `internal/` packages that are
   really the API of this codebase, and `cmd/brag`.
3. **`SECURITY.md`** — *(landed 2026-08-15, ahead of framing)*. A real policy
   pointing at the checked-in 2026-04-26 review.
4. **The `DEC-041` gap.** `decisions/` jumps 040 → 042. The reservation belongs
   to the deferred SPEC-070 (`brag project goto`) and its unwritten
   "multi-location primary policy" decision. The gap is explained **in the
   backlog**, which a browsing reader never sees. Land a reservation tombstone
   in `decisions/` that names the holder and links the backlog item, so the gap
   stops reading as a lost decision.
5. **Open-questions hygiene.** **8 of 18 are open.** (It was 8 of 17; `tag-ordering-projection` closed and SPEC-079 opened `dec-amendment-heading-convention`, so the count returned to 8 against a larger total. This line has now been wrong twice in three days, which is the argument for the hygiene pass deriving it rather than restating it.) Three
   date from **2026-04-19** and have not moved in four months
   (`shareable-ids`, `editor-template-format`, `summary-grouping-heuristics`) —
   at least one is likely answered-in-practice but never closed. For each open
   question: mark answered and link what settled it, delete it if it resolved
   informally, or restate its resolve condition so it is actionable.

   *`tag-ordering-projection` was closed ahead of this spec (2026-08-18) and is
   the worked example of what this item is for.* It had been open since
   2026-06-06 carrying its own falsifiable trigger — drop the `position` column
   if no production entry is observed with unsorted tags. Nobody had run it.
   Measured: **260 of 278 tagged entries are not in name-ASC order**, so the
   trigger fails and `position` stays. Resolved by observation, no DEC — the
   pattern this hygiene pass should look for is exactly that: **a question whose
   own stated condition has already been decided by reality, waiting only for
   someone to check.**
6. **The `pr:` cross-repo collision note in `BRAG.md`** — *(landed 2026-08-15,
   ahead of framing)*.

### Explicitly out of scope

- Lint, coverage, and any CI change — STAGE-022.
- Any feature work, any new command, any change to what the binary does.
- A `CODE_OF_CONDUCT.md`. Deliberate: it is boilerplate for a single-maintainer
  project, and adding it is exactly the cargo-culting this repo's story is that
  it does not do. A sharp reviewer reads it as padding.
- Rewriting the tutorial or the blog posts. They exist; this stage links them.

## Spec Backlog

Ordered list of specs composing this stage. IDs assigned at creation.

Format: `- [status] SPEC-ID (cycle) — one-line summary`

- [x] SPEC-079 (shipped on 2026-08-18) — **the practices entry point.** README section + the
      document it points at; every claim backed by a countable artifact; counts
      derived or pinned rather than typed. Framed 2026-08-16, **GO at complexity
      M** — the counting rule is a real design problem, not a formality. Framing
      caught two claims that do not survive contact with the code (there are no
      "golden files"; the corrections claim has no counting rule) and three
      already-stale counts in this project's own brief.
      **Designed 2026-08-16.** All four forks settled with rejected alternatives
      (LD1 counting rule = one derived script + one whole-table diff; LD2 split
      point = the README carries routing and *zero* numbers, because guard
      coverage is the split; LD3 honest-close = its own named section,
      second-to-last, foreshadowed from the intro, by citation; LD4 = do NOT
      create a `## Amendment` convention — cite, and carry a derived count).
      Plus LD5: `test-docs` A1's band was already at its ceiling (README exactly
      250), so it is **re-pinned tight to 260**, not widened. Design re-measured
      every count: framing's **164 doc assertions is wrong and unstable** (it
      double-counts `S3` and is environment-dependent); the reproducible figure
      is 163 distinct ids, 171 after this spec. Also stale: "four W-series
      guards" (there are six) and the brief's "zero deprecated" (DEC-004 is
      superseded by DEC-015 — which strengthens the story rather than damaging
      it).
      **Built 2026-08-18** on `build/spec-079-practices-entry-point`. All six
      literals transcribed; `just test-docs` green at **171 distinct assertion
      ids** (was 163, delta exactly `X1`–`X8`), `just test` / `gofmt -l .` /
      `go vet ./...` unaffected. `X3` earned its keep on day one: the design
      snapshot recorded **7 projects**, but PR #165 (`PROJ-008`/`PROJ-009`
      scaffolds) merged ~2h before the design PR, so the true count on `main`
      is **9**. Transcribing literal ① verbatim made `X3` the *only* failing
      assertion, naming exactly one drifted row; the fix was the guard's own
      mechanical remedy (`just inventory`, paste). The counting guard caught a
      stale count before a human read the page — which is the whole thesis of
      this spec, demonstrated at its own expense.
      **Verified 2026-08-18 — ⚠ PUNCH LIST (3 items), back to build.** The
      mechanical half is clean: all six literals byte-faithful, `X3` confirmed
      passing for the right reason (both sides non-empty and byte-identical,
      and it goes red when the *repo* drifts, not just the page), the id delta
      exactly `X1`–`X8` in both directions, and all eight Group X assertions
      plus the re-pinned `A1` mutation-tested with the mutant proven real. The
      `Projects` deviation was the correct resolution and is honestly recorded.
      What did not survive is the page's own standard, in three prose claims no
      guard reaches: it says the `TestMemoryCmd_EndToEndMarkdownGolden` comment
      lists **the budget** as unpinned when the comment lists the declared
      Matched order and the golden provably *does* pin the budget
      (mutation-confirmed); it routes the first three "does not measure" gaps to
      STAGE-022 when STAGE-022 lists benchmarks under **Explicitly out of
      scope** and never mentions `test-docs` at all; and it says every archived
      spec ends with a build-phase reflection when **six of 75 lack the standard
      `### Build-phase reflection` heading — of which five carry none at all,
      SPEC-046 having written one under `### Honest reflection`.** Each is a
      claim backed by a path that contradicts it — the exact failure the page
      exists to prevent. Fixed in both the page and literal ①; the delta was
      re-verified (✅ 2026-08-18) before ship.
      *(This sentence itself first read "six of 75 carry none", which the
      re-verify caught: the grep is syntactic, the claim was semantic, and they
      diverge on exactly one spec. Recorded rather than quietly corrected —
      it is the same defect class, in the stage file that owns it.)*
- [ ] (not yet framed) — **the godoc pass**, plus the two legibility repairs
      that need no design (`DEC-041` tombstone, open-questions hygiene).

**Count:** 1 shipped / 0 active / 1 pending

> **The stage is NOT complete.** `just archive-spec` printed *"All specs for
> STAGE-021 are shipped"* on this ship — it counts written specs, not backlog
> items, so an unframed entry is invisible to it. The godoc / `DEC-041` /
> questions-hygiene spec below is still pending. Do not run the Stage Ship
> prompt on the strength of that message.

## Design Notes

- **Point at artifacts, never adjectives.** The persuasive version of this page
  is "45 decisions, 4 of which record their own correction — here they are,"
  not "we take decisions seriously." Every sentence should survive the question
  *what would I click to check that?*
- **Derive the counts.** The README version line was wrong through two releases
  (SPEC-077) until a test derived it from the CHANGELOG. A practices page full
  of hand-typed counts is the same defect waiting to happen — either compute
  them or pin them in `test-docs`.
- **The honest-close material is an asset, not an embarrassment.** PROJ-006
  closed as an explicit scope reduction; DEC-043's first correction was itself
  wrong on a second axis. Most repos cannot show a decision log that records
  being wrong. Surface it deliberately rather than quietly.
- Prefer one document plus a README section over a docs site. A site is a
  distribution project; this is an index.

## Dependencies

### Depends on
- Nothing. Every input already exists in the repo — that is the premise of the
  stage.

### Enables
- **STAGE-022** — the practices page is what a coverage number gets presented
  *against*, so writing it first tells the lint stage what it is allowed to
  claim.
- The PROJ-007 outcome generally: quality that is legible to an outside reader.

## Stage-Level Reflection

*Filled in when status moves to shipped. Run Prompt 1c (Stage Ship) in
FIRST_SESSION_PROMPTS.md to draft this.*

- **Did we deliver the outcome in "What This Stage Is"?** <yes/no + notes>
- **How many specs did it actually take?** <number vs. plan>
- **What changed between starting and shipping?** <one sentence>
- **Lessons that should update AGENTS.md, templates, or constraints?**
  - <one-line updates>
- **Should any spec-level reflections be promoted to stage-level lessons?**
  - <one-line items>
