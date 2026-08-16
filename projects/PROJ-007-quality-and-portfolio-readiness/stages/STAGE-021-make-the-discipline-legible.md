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
5. **Open-questions hygiene.** 8 of 17 are open. Three date from **2026-04-19**
   and have not moved in four months (`shareable-ids`,
   `editor-template-format`, `summary-grouping-heuristics`) — at least one is
   likely answered-in-practice but never closed. For each open question: mark
   answered and link what settled it, delete it if it resolved informally, or
   restate its resolve condition so it is actionable. **Do not close
   `tag-ordering-projection`** — it is live and blocking (see STAGE-022).
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

- [ ] SPEC-079 (design) — **the practices entry point.** README section + the
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
- [ ] (not yet framed) — **the godoc pass**, plus the two legibility repairs
      that need no design (`DEC-041` tombstone, open-questions hygiene).

**Count:** 0 shipped / 1 active / 1 pending

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
