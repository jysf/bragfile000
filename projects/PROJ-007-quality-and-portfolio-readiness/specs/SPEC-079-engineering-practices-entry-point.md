---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-079
  type: story                      # epic | story | task | bug | chore
  cycle: frame                     # frame | design | build | verify | ship
  blocked: false
  priority: high
  complexity: M                    # S | M | L  (L means split it)

project:
  id: PROJ-007
  stage: STAGE-021
repo:
  id: bragfile

agents:
  architect: claude-opus-5
  implementer: claude-opus-5       # usually same Claude, different session
  created_at: 2026-08-16

references:
  decisions: []
  constraints: []
  related_specs:
    - SPEC-021                     # README user-facing rewrite — the last time this file was reshaped
    - SPEC-077                     # pinned the README version line after it rotted through two releases
---

# SPEC-079: engineering practices entry point

> **Cycle: frame.** This document is a **go/no-go**, not a design. It establishes
> the problem, the verified inventory, and the shape of the answer, and it hands
> the design session a set of decisions to make. Per AGENTS.md §6, design happens
> in a **fresh session**.

## Context

STAGE-021's premise is that bragfile's engineering discipline is **strong and
unreachable**. This spec is the stage's primary deliverable: the entry point that
makes it reachable.

The work is an **index, not an essay.** Every artifact a reviewer would find
persuasive already exists in the repository. What does not exist is any path to
it from the front door.

Parent: `STAGE-021-make-the-discipline-legible`, spec 1 of 2.
Project: `PROJ-007`.

## Goal

A first-time reader arriving at the README can reach — and verify — the decision
log, the spec-driven cycle, the test regime, the security posture and the
incident→hardening pattern, **without knowing the repository's directory
layout.** Every claim on the page points at something countable.

## The verified inventory (measured 2026-08-16)

Design should re-measure rather than trust these; they are recorded so the design
session starts from facts rather than adjectives.

| Artifact | Count | Location |
|---|---|---|
| Decision records | **45** | `decisions/DEC-*.md` |
| Archived specs | **75** | `projects/*/specs/done/` |
| Test files / source files | **78 / 69** | `internal/`, `cmd/` |
| `test-docs` assertions | **164** | `just test-docs` |
| Security review | 1, dated 2026-04-26 | `docs/reports/security/` |
| Cross-project retrospective | 1, three repos | `docs/reports/cross-project/` |
| Blog posts | 2 | `docs/blog/` |
| Research syntheses | 4 | `docs/research/` |
| Framework-feedback docs | 4 | `docs/framework-feedback/` |
| README | **1,186 words**, no entry point to any of the above | `README.md` |

## The finding that shapes this spec

**Three of the brief's own counts are already stale**, in the very document that
motivates a project about legibility:

- "the DEC log (**39 decisions**, zero deprecated)" — it is now 45.
- "**66 test files to 60 source files**" — it is now 78 / 69.
- "a **116-assert** doc-test harness" — it is now 164.

None was wrong when written. All three rotted, which is exactly how the README's
version line rotted through two releases until SPEC-077 pinned it to a derived
value. **A practices page full of hand-typed counts is that defect with a larger
blast radius**, because this page exists to be trusted by someone who cannot
check it easily.

So the counting rule is not a nice-to-have — it is the spec's central design
problem.

### Two claims that do not survive contact with the code

Design must not restate these as-is. Both were caught while framing:

1. **"Golden files."** There are none. `find . -name '*.golden'` returns nothing
   and there are no `testdata/` directories. The discipline is real but its
   mechanism is different: **byte-exact expected output embedded as Go string
   literals in the test files.** A page claiming "golden files" would be
   false and trivially falsifiable by the exact reader it is written for.
2. **"Decision records that log their own corrections."** True in substance —
   DEC-044 carries a correction box, DEC-043's first repair was itself wrong on a
   second axis — but **not currently countable**. Only **1** DEC has an explicit
   `## Amendment` section; a keyword grep matches all 45 because ordinary prose
   uses those words. There is no counting rule, so there is no number. Design
   must either define one (a convention DECs adopt going forward) or make the
   claim by *citation* — naming the two or three specific records — rather than
   by count.

## Outputs (shape only — design decides the specifics)

- **Files created:** one practices document under `docs/`.
- **Files modified:** `README.md` — a section pointing into it.
- **Possibly modified:** the `test-docs` harness, if counts are pinned there.

## Acceptance Criteria (frame-level; design refines into testable form)

- [ ] From the README, a reader reaches the DEC log, the spec archive, the test
      regime, the security posture and the incident→hardening pattern.
- [ ] Every claim on the page is backed by a path, a count, or a named test.
      **No adjective stands alone** — no "rigorous", no "comprehensive".
- [ ] Every number on the page is **derived at read time or pinned by
      `test-docs`.** A hand-typed count is a defect, not a shortcut.
- [ ] The page describes the test regime **as it actually is** (embedded
      byte-exact literals), not as "golden files".
- [ ] The correction claim is either countable by a stated rule or made by
      citation.

## Decisions for the design session

These are the real forks. Framing deliberately does not settle them.

1. **How are counts kept true?** Three candidates, and they are not equivalent:
   *(a)* a `test-docs` assertion per count — consistent with the four W-series
   guards, but each count becomes a test to maintain; *(b)* a `just` recipe that
   prints the inventory, with the page pointing at the command rather than
   quoting numbers — no rot possible because no number is stored; *(c)* generate
   the page's inventory table from the repo. **(b) looks strongest** — it makes
   the number unrotting by construction rather than by guard — but it changes the
   page from "here are the numbers" to "here is how to get them," which is a real
   readability tradeoff for the hiring-manager audience.
2. **One document or a README section?** The stage says one document plus a
   section. Design should confirm the split point: what the README carries
   inline versus what lives behind the link.
3. **How much of the honest-close material goes on the page?** PROJ-006 closed as
   an explicit scope reduction; DEC-043's first correction was wrong again. This
   is the most differentiated material in the repo and also the most easily
   misread. Design decides how prominently it sits.
4. **Does the page name the counting rule for corrections**, thereby creating a
   convention future DECs must follow? That is a small process change riding in
   on a docs spec — it should be a conscious call.

## Out of scope (for this spec specifically)

- The **godoc pass** and the two legibility repairs (`DEC-041` tombstone,
  open-questions hygiene) — STAGE-021's second spec.
- Lint, coverage, badges, CI — STAGE-022.
- Rewriting the tutorial or the blog posts. This spec **links** them.
- Any change to what the binary does.
- Fixing the brief's three stale counts. Tempting, but it is a different file
  with a different purpose; note them and move on, or the spec grows a tail.

## Go / no-go

**GO.** Complexity **M**, not S: the writing is small, and the counting rule is a
genuine design problem with three non-equivalent answers and a readability
tradeoff. Sizing it S is what would produce a page of hand-typed numbers that
rot by December.

**Why now:** it is the first spec of the active stage, it blocks nothing, and it
establishes what STAGE-022's coverage number is allowed to claim.

**What would make this a no-go:** if the intent were new prose about the project
rather than an index to existing artifacts. That is a different, larger spec and
should be framed as one.
