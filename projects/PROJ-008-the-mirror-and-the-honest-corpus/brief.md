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
5. **The read path's discoverability** — *candidate, added 2026-08-23.*
   `brag memory` is the corpus-as-working-memory call and it is well built:
   three bounded reads deduped into a candidate pool, ranked by RRF (DEC-043),
   fitted to a token budget, and exposed over MCP as `brag://memory/recent`
   and `brag://memory/project/{name}` — which an agent can attach **with no
   tool call at all** (DEC-045). `docs/for-ai-agents.md` §3–§4 documents all of
   it properly.

   **The problem is that none of that travels.** `BRAG.md` is the file designed
   to be dropped into any repo and read cold, and it is **write-only**: its
   *"Your role, in one sentence"* casts the agent purely as a capturer
   (*"propose a brag entry… only execute `brag add` after the user approves"*),
   and `brag memory` appears **twice**, both in passing inside the
   plugin-install section. The plugin reinforces the asymmetry — the only hook
   is a **`Stop`** hook (capture-nudge, session end). There is no session-start
   counterpart. So an agent working in another repo with `BRAG.md` dropped in
   is told to write to the corpus and never told to read it.

   This belongs to PROJ-008 rather than anywhere else because the mirror and
   `brag learn` both make the corpus more worth reading, and all three fail the
   same way if nothing tells an agent to read it. Measured 2026-08-23 on the
   live tree.

   **Shape of the fix (for framing, not settled):** a "read before you write"
   section in `BRAG.md` with the three real invocations, and possibly a
   `SessionStart` hook mirroring the existing `Stop` one. Note the second half
   is a behaviour change in the plugin and should be argued, not assumed —
   auto-loading context every session has a cost, and the MCP resource path may
   already be the better answer for agents that support it.

   **Same shape as a finding already filed upstream** in
   `docs/framework-feedback/process-feedback.md` §6: a capability exists, and
   the artifact that would cause it to be used does not mention it. That one
   was about the spec template's work-log hook; this one is about `BRAG.md`'s
   read path. Two instances now — worth noticing as a class.

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

*Framed 2026-09-05. The ordering below reverses the proposal above; the call
and its rejected option are recorded in `stages/STAGE-023-*.md` under
**The ordering call**.*

- [x] **STAGE-023 — `brag learn`: the corpus learns to hold failures.** ACTIVE.
      Framed 2026-09-05 with SPEC-085 at `frame`, GO, complexity M.
- [ ] **STAGE-024 — impact quality**: the classifier and its consumers.
      **No longer first.** Framing measured the classifier and found it is not
      the small primitive this brief assumed — candidate rules select between
      11 and 230 of 397 entries, 141 of the 230 digit-bearing impacts carry an
      *identifier* digit rather than a measurement, and 27 quantified impacts
      use spelled-out numerals invisible to any digit rule.
- [ ] **STAGE-025 — the mirror**: observations over the corpus.
- [ ] **STAGE-026 — story-surface v2**: the arc.

> **USER PRIORITY, stated 2026-08-23 at activation: `brag learn` matters most.**
> **Framing agreed, 2026-09-05 — but not merely on preference.** The deciding
> argument is that *the classifier's calibration data is the corpus itself*: an
> impact rule designed against 397 wins is tuned on a population that excludes
> the case it will most need to judge. The rejected option — this brief's own
> *one classifier, three surfaces* ordering — is written up in full in the stage
> file, including what the rejection costs (one stage of delay on `wrapped`'s
> highlight-reel problem, true since v0.4.0 and not degrading).

**Count:** 0 shipped / 1 active / 3 pending

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
