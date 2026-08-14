---
# Maps to ContextCore epic-level conventions.
# A Stage is a coherent chunk of work within a Project.
# It has a spec backlog and ships as a unit when the backlog is done.

stage:
  id: STAGE-020                     # stable, zero-padded, repo-global (never reused)
  status: active                    # proposed | active | shipped | cancelled | on_hold
  priority: high                    # critical | high | medium | low
  target_complete: null

project:
  id: PROJ-006                      # parent project
repo:
  id: bragfile

created_at: 2026-08-13
shipped_at: null
---

# STAGE-020: git-hash evidence links

## What This Stage Is

A brag entry can currently *claim* anything. This stage makes a claim
**checkable**: an entry carries the commit that proves it, and any reader with
the repo can verify that commit exists, who wrote it, and when — without
trusting bragfile's own metadata at all.

That is a different kind of provenance from the `agent:`/`model:` tags already
stamped (DEC-024/DEC-027). Those are *self-reported*: bragfile says an agent
wrote this, and you either believe the record or you don't. A commit hash is
**self-attesting** — the evidence lives outside the corpus and can be checked
against it. The project brief calls this "arguably a stronger and cheaper win"
than signed provenance, because it buys verifiability without any crypto,
key management, or new trust root.

## Why Now

**The trigger fired explicitly.** The roadmap's `resume_when` for this pillar is
*"wanting to tie a brag to the commit that proves it"*, and that is the stated
ask (2026-08-13).

But the sharper reason is a measurement, and it inverts the obvious framing of
this work:

> **The convention already exists, costs nothing, and has never once been
> used.** `commit:<hash>` was documented in BRAG.md on 2026-07-12 —
> deliberately, as "establish the *pattern* before framing a stage" — and it
> works today with zero code, riding DEC-015's normalized tags. Across the
> **whole corpus there are zero `commit:`, `pr:` or `issue:` tags** (361 entries,
> measured 2026-08-13 — the count grows, the zero is the point).

So the question this stage exists to answer is **not** "how do we support
evidence links" — they are supported.

**Corrected at framing review (2026-08-13).** An earlier draft of this section
diagnosed the zero as "hand-typing a hash is friction nobody will pay." That was
wrong, and the correction matters because it changes what should be built.
Checked instead of assumed:

- the MCP `brag_add` tool description — which an agent reads on **every call** —
  mentions `commit:` **nowhere**;
- the `/brag` slash command mentions it nowhere;
- the capture-nudge Stop hook mentions it nowhere *in its prompt* — but
  `capture-nudge.sh:34` already runs `git rev-parse HEAD` as its "did anything
  ship" signal, **and then discards the hash**.

So the instruction lives only in BRAG.md, a document an agent reads if it
happens to, and never at the moment of capture. The experiment "does telling
agents work?" has **never been run**, and the answer was already sitting in the
hook's own variable. The zero measures a missing instruction, not a rejected
one.

That inverts the build order. The first move is to **deliver the instruction
where capture happens** — tool description, slash command, nudge text — and read
the adoption number. Stamping machinery is the fallback if that fails, not the
opening move, and it may not be needed at all.

## Success Criteria

- An entry captured through the MCP `brag_add` path carries the repo's commit
  **without anyone typing a hash**, on the same "stamped, not hand-typed"
  footing as `agent:`/`model:`.
- The stamped value is **verifiable**: given an entry, a reader can run a single
  documented command to confirm the commit exists in the repo and read its
  author/date/message.
- **A wrong or unverifiable hash is visible, not silent.** Whatever is stamped
  can be checked later and reported as unverifiable rather than quietly trusted.
- The hand-typed `commit:<hash>` convention keeps working unchanged — this stage
  adds an automatic path, it does not replace the manual one or invalidate any
  existing entry.
- Adoption is *measurable*: after this ships, the share of new entries carrying
  an evidence link is a number we can state, and the zero above is the baseline.
- No new dependency, no CGO, no network, no schema migration if avoidable.

## Scope

### In scope
- **Deciding what commit an entry should point at**, which is the real design
  question (see Design Notes) — HEAD at capture time is the obvious answer and
  probably the wrong one.
- **Automatic stamping** on the paths that already stamp provenance, so the
  evidence link costs the user nothing.
- **A verification surface**: something that answers "does this entry's evidence
  actually check out?" against the repo.
- Promoting `commit:`/`pr:`/`issue:` from convention to a **reserved, validated**
  prefix set, alongside the DEC-046 reserved-prefix machinery that already
  exists (`capture.ReservedTagPrefixes`, single-sourced and mutation-pinned).

### Explicitly out of scope
- **Signed / attestable provenance** (the separate pillar). This stage buys
  verifiability *without* crypto; if it succeeds, it partly reduces the pressure
  for signing, which is a finding worth having before that stage is framed.
- **The git-import cold-start miner** — the other half of capture completeness.
  It shares the git seam but is a different capability (backfilling a corpus vs.
  attesting new entries).
- Any hosted/remote verification. Local-first, offline-checkable.
- Forge-specific integration (GitHub/GitLab APIs). `pr:`/`issue:` may be
  *recorded* as opaque refs; resolving them over a network is not this stage.

## Spec Backlog

Ordered list of specs composing this stage. Add specs as you identify
them. Update status as specs progress.

Format: `- [status] SPEC-ID (cycle) — one-line summary`

- [x] SPEC-078 (shipped on 2026-08-14) — **tell agents to record evidence links.** Instruction
      only: the `brag_add` tool description, `/brag`, the Stop-hook nudge text
      and the agent docs. No stamping, no schema, no validation. Runs the
      experiment the zero never tested, and makes adoption measurable against a
      recorded baseline (0 / 361 at 2026-08-13).
- [ ] (not yet written) — **gated on a measurement due 2026-09-14**, not on a
      feeling: re-run `brag tags | grep -cE '^(commit|pr|issue):'` then (~1 month
      and ~90 new entries at the observed ~22/week). Write this spec only if the
      number is still at or near zero. A gate with no date is a gate nobody
      pulls. If adoption moved, drop it and say so. The
      stamping path + reserved-prefix promotion, and the decision about WHICH
      commit that requires. Deliberately gated on evidence rather than
      scheduled.
- [ ] (not yet written) — the verification surface (`brag verify`-shaped).

**Count:** 1 shipped (SPEC-078) / 0 active / 2 pending

> **The stage stays `active`.** SPEC-078 meets one of the six Success Criteria
> substantively (adoption is measurable) and two vacuously by having changed
> nothing; the three about stamping and verification are deferred by LD1. The
> criteria are deliberately NOT rewritten around the instruction-first approach —
> that would retro-fit the goalposts to the spec that just shipped and erase the
> framing correction that explains why the expensive spec is second. They stay
> the standard this stage is still trying to meet, and the 2026-09-14
> measurement decides whether specs 2 and 3 get written or dropped.

> **Ordering note.** The expensive spec is second on purpose. If instruction
> moves adoption, most of the stamping work is unnecessary — and stamping
> carries a hazard instruction does not, because under squash-merge the hashes
> available at capture time are the ones that will not survive (SPEC-078 LD3).

## Design Notes

**The hard question is not the tag, it is which commit.** HEAD-at-capture-time
is the obvious implementation and it is wrong often enough to matter:

- You brag *before* committing (the work is done, the commit comes after) —
  HEAD points at the previous commit, which does not prove the claim.
- The work spans several commits, or a squash-merge that does not exist yet at
  capture time.
- You capture from a different repo than the one the work happened in — this
  corpus is cross-project by design (`crustyimg`, `animal-slots`, `bragfile`
  all in one DB).
- The capture happens in a dirty tree, so no commit represents that state.

A stamped-but-wrong hash is **worse than no hash**, because it looks like
evidence. Whatever this stage does must be able to say "unverifiable" out loud.

**Precedent to follow:** `agent:`/`model:` (DEC-024/027) are the shape that
worked — reserved prefix, stamped by the tool, never hand-typed, riding
DEC-015's tags with no schema change. `capture.ReservedTagPrefixes` is already
the single source both stamping and validation read (DEC-046 punch-list item 2,
mutation-pinned), so adding a prefix there is a solved problem rather than a new
mechanism.

**Precedent to avoid:** the reserved-tag forgery gap. A freeform `tags` string
can contain `commit:deadbeef` typed by hand, and validation cannot distinguish
it from a stamped one — DEC-046 documents this and accepts it. For *self-reported*
tags that is tolerable. For an **evidence** tag it is the whole point, so this
stage has to decide whether an unverified hash is acceptable, and the answer
should be "yes, but it must be checkable" rather than "no".

**Measure adoption, not just capability.** This stage's own premise is that a
capability nobody uses is not a win. Whatever ships should make the adoption
number answerable — the baseline is zero, against a corpus of 361.

## Dependencies

### Depends on
- Nothing blocking. `capture.ReservedTagPrefixes` (DEC-046) and the normalized
  tags model (DEC-015) both already exist and are the substrate.
- The BRAG.md `commit:` convention is the documented starting point, not a
  dependency.

### Enables
- **Signed provenance (the next pillar) can be scoped honestly** — if
  self-attesting evidence covers most of the need, signing has a smaller job.
  That is a finding worth having *before* framing it.
- The **benchmark** pillar, which needs provenance that can be trusted rather
  than merely recorded.
- The git-import miner, which shares the git seam this stage introduces.

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
