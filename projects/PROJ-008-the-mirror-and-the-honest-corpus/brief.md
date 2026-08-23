---
# Maps to ContextCore project.* semantic conventions.
# A project is a bounded wave of work against the repo (the app).

project:
  id: PROJ-008
  status: active
  priority: high
  target_ship: null

repo:
  id: bragfile

created_at: 2026-08-16
activated_at: 2026-08-23
shipped_at: null
---

# PROJ-008: The Mirror and the Honest Corpus

> **CANDIDATE — proposed, not opened.** Captured 2026-08-16 while PROJ-007 is
> the active project, so the direction is on record rather than living in a
> chat log. It is **not committed work**, and everything below is scope input,
> not a plan. Open it when PROJ-007's two stages land.

## What This Project Is

The corpus stops being a list you read and starts being **something that tells
you what it sees.** Three surfaces, one idea:

- **`brag learn`** — capture what *didn't* work. The corpus currently records
  only wins, which makes it a flattering and incomplete input to both a
  retrospective and an agent.
- **The mirror** — narrative intelligence. The app notices patterns across the
  corpus and says them out loud: *"your last six entries are all unblocking
  others — you're operating like a lead"*; *"this project reads like a
  turnaround"*; *"you undersold three of these — impact, but no numbers."*
- **Story-surface v2** — the output surface both of the above feed, showing the
  **arc** (foundation → payoff) rather than a flat win-list.

They are not three features. They are input, observation, and output of the same
loop, and each is weaker alone: a mirror over a wins-only corpus flatters, and a
story with nothing to rank has no basis for choosing what leads.

## Why Now

**Because the highest-leverage primitive is small and blocks all three.**

Measured 2026-08-15 across 368 entries: **294 (80%) carry an `impact`** — which
is genuinely good capture discipline — but only **52 (14%) carry a number with a
unit**, and 51 (14%) a before/after shape. So **242 impact statements are
prose.**

`brag impact` already reports *presence* in its header (`Entries: 178/204 with
impact`). It counts whether the field is filled, never whether it says anything
measurable. That single missing distinction — **quantified vs. prose** — is what
blocks all three surfaces:

- `brag wrapped`'s "Impact moments" section is byte-identical to `brag impact`'s
  body: **all 178 entries**, inside a digest its own help calls "shareable,
  celebratory." It cannot select a highlight reel because nothing ranks impact.
- The mirror's sharpest observation — *you undersold three of these* — is
  exactly this classifier.
- Story v2 has no basis for deciding what leads a thread.

One classifier, three surfaces. That is the ordering argument for this project.

**Secondary trigger, already fired:** the PROJ-005 synthesis parked `brag learn`
with the explicit promotion condition *"promote to (A) if the memory work
lands."* It landed in v0.6.0.

## Success Criteria

*To be set at framing.* Likely themes:

- Impact quality is **measurable**, not just presence — a stated rule for what
  counts as quantified, applied consistently wherever impact is surfaced.
- `brag wrapped`'s impact section is a **selection**, not the full ledger.
- The corpus can hold a recorded failure, and `brag memory` surfaces them, so an
  agent reading the history learns what *didn't* work as well as what did.
- At least one mirror observation that is **true, specific, and not obvious** —
  something the maintainer did not already know about their own corpus.
- No LLM in the binary. Rule-based composition; prose via the existing pipe.

## Scope

### In scope
1. **Impact quality** — the classifier, and every surface that should consume
   it (`impact`, `wrapped`, the density report).
2. **`brag learn`** — capturing failures/anti-brags as first-class corpus
   content, including how they participate in `memory` and the digests.
3. **The mirror** — rule-based pattern detection over the corpus, with an
   honest bar for what counts as an observation worth surfacing.
4. **Story-surface v2** — the arc-shaped narrative output.

### Explicitly out of scope
- **Any LLM inside the binary.** Rule-based core, prose via pipe — the same
  boundary held since PROJ-001.
- Anything requiring the network.
- Federation, sharing, multi-user, SaaS. The PROJ-005 synthesis is explicit
  that those wait until the corpus is trustworthy and complete.
- Signed provenance and capture completeness — separate parked pillars. Note
  **capture completeness has a fired trigger** (see Dependencies) and may
  deserve its own project ahead of this one.

## Stage Plan

*Proposed, not framed. Sequence/split at framing.*

- [ ] (not yet defined) — **impact quality**: the classifier and its consumers.
      Almost certainly first — it unblocks the other two.

> **USER PRIORITY, stated 2026-08-23 at activation: `brag learn` matters most.**
> Recorded here as a framing input, not as a settled ordering. It sits against
> the brief's own argument two sections up — *one classifier, three surfaces* —
> which puts impact quality first because the other pillars consume it. Framing
> owes an explicit call, with the rejected option written down. Two things
> weigh for `brag learn` beyond the preference: its promotion condition from the
> PROJ-005 synthesis (*"promote if the memory work lands"*) **already fired**
> when v0.6.0 shipped, and it is the only pillar that changes what the corpus
> can *hold* rather than how it is read — a wins-only corpus flatters every
> surface built on it, including the classifier. If framing still puts impact
> quality first, it must say why the flattery problem can wait.
- [ ] (not yet defined) — **`brag learn`**: the corpus learns to hold failures.
- [ ] (not yet defined) — **the mirror**: observations over the corpus.
- [ ] (not yet defined) — **story-surface v2**: the arc.

**Count:** 0 shipped / 0 active / 4 pending

## Dependencies

### Depends on
- **PROJ-007** landing, so quality work is not competing with feature work.
- Nothing technical. Every input ships already: `internal/aggregate`,
  `internal/memory`, `internal/story`, the DEC-014 envelope.

### Related, and possibly ahead of this in line
- **Capture completeness** (the staging inbox + git-import cold-start miner) is
  a parked PROJ-006 pillar whose `resume_when` — *"the corpus feels sparse, or
  you notice unlogged work"* — **fired on 2026-08-15**, when two releases, a
  project closure and a live prod escape were found unlogged for five days. It
  is also the strongest answer to new-user cold start, since a fresh install
  shows an empty database and every compelling surface needs a populated corpus.
  Decide at framing whether it belongs here, ahead of here, or in its own
  project.

### Enables
- The benchmark pillar (*which model earns its keep*), which needs a corpus
  whose impact is comparable, not just present.
- Any eventual sharing or team story — the synthesis is emphatic that those
  depend on the corpus being trustworthy and complete first.

## Project-Level Reflection

*Filled in when status moves to shipped.*

- **Did we deliver the outcome in "What This Project Is"?** <yes/no + notes>
- **How many stages did it actually take?** <number, compare to plan>
- **What changed between starting and shipping?** <one or two sentences>
- **Lessons that should update AGENTS.md, templates, or constraints?**
  - <one-line updates>
- **What did we defer to the next project?**
  - <one-line items>
