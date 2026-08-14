---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-078
  type: story                      # epic | story | task | bug | chore
  cycle: ship
  blocked: false
  priority: high
  complexity: S                    # S | M | L  (L means split it)

project:
  id: PROJ-006
  stage: STAGE-020
repo:
  id: bragfile

agents:
  architect: claude-opus-5
  implementer: claude-opus-5       # usually same Claude, different session
  created_at: 2026-08-13

references:
  decisions:
    - DEC-015                      # normalized tags — the substrate, unchanged
    - DEC-024                      # agent:/model: stamped, not hand-typed — the precedent
    - DEC-025                      # the capture-nudge Stop hook
    - DEC-046                      # reserved prefixes, single-sourced + mutation-pinned
  constraints:
    - one-spec-per-pr
  related_specs:
    - SPEC-041                     # the capture-nudge hook
---

# SPEC-078: tell agents to record evidence links

## Context

`commit:<hash>` has worked since 2026-07-12. It needs no code — it rides
DEC-015's normalized tags, so `brag list --tag commit:abc123` already filters
and `brag tags` already counts it. BRAG.md documents it, with a worked example
and the argument for why it is worth the habit.

**It has never been used. Zero `commit:`, `pr:` or `issue:` tags across the
whole corpus** (361 entries when this was written, 363 when re-measured at
build the same day — the corpus grows continuously, which is why the ZERO is
the figure that matters and the count is only context).

STAGE-020's first framing read that zero as friction — hand-typing a hash is
work nobody will do. That was wrong, and checking rather than assuming produced
a much cheaper spec:

| surface an agent reads at capture time | mentions evidence links? |
|---|---|
| MCP `brag_add` tool `Description` (served once via `tools/list`, then present in context for **every** call) | **no** — it is one sentence: *"Capture a new brag entry. Requires a non-empty title."* |
| `/brag` slash command (`plugin/commands/brag.md`) | **no** |
| capture-nudge Stop hook prompt (`plugin/hooks/capture-nudge.sh`) | **no** |
| BRAG.md | yes — a document an agent reads *if it happens to* |

The instruction exists in exactly one place, and it is not a place anyone stands
at the moment of capture. **The experiment "does telling agents work?" has never
been run.**

And the answer was already in hand: `capture-nudge.sh:34` runs
`git rev-parse HEAD` as its "did anything ship" signal, and line 52 compares it
against a stored `BASELINE` — so at nudge time the hook knows **every commit
that landed during the session**, and discards all of it.

The precedent that settles the approach is inside this repo: `agent:` and
`model:` are *also* just tags on the same normalized model, and they carry
**70 and 56 uses**. Not because they are special — because `brag_add` stamps
them and nobody types them. The gap between 70 and zero is about who writes the
tag, not about the tag.

## Goal

Put the instruction where capture actually happens, and make the resulting
adoption measurable — so the next decision (stamp automatically, or leave it to
instruction) is made on a number instead of an intuition.

## Inputs

- **Files to read:** `BRAG.md` (§Evidence links, including the ref-preference
  guidance added 2026-08-13), `internal/mcpserver/server.go` (the tool
  descriptions), `plugin/commands/brag.md`, `plugin/hooks/capture-nudge.sh`,
  `docs/for-ai-agents.md`.
- **Related code paths:** `internal/mcpserver/`, `plugin/`.
- **External APIs:** none.

## Locked design decisions

### LD1 — Instruction only. No stamping, no schema, no validation.

This spec adds **no automatic tagging**. It changes tool descriptions, the
slash command, the hook's nudge text and the agent docs. Nothing else.

The reason is that stamping is the *expensive* answer to a question we have not
yet asked, and it carries a hazard instruction does not (LD3). If instruction
moves adoption off zero, most of the stamping work is unnecessary. If it does
not, we will know that having spent a doc change rather than a subsystem.

### LD2 — Prefer `pr:`, and say why, at every surface.

The ref-preference guidance already in BRAG.md is the content to propagate, in
its established order: `pr:<n>` first, then `commit:<hash on main>`, then
several `commit:` tags, never a range, and **record nothing when in doubt**.

This is not stylistic. Verified on this repo: merging PR #151 replaced branch
commit `4d6d24c` with `29b6e88` on `main`; `4d6d24c` is unreachable from any
branch, its ref was deleted, and it does not exist in a fresh clone. A `commit:`
tag naming it is an evidence link that verifies for nobody.

### LD3 — The hook must NOT suggest the session's own commits.

Tempting and wrong, and this is the spec's sharpest finding.

The hook fires at session end. Under this repo's squash-merge workflow, the
commits available at session end are **branch** commits that are about to be
rewritten by the squash and then orphaned by `--delete-branch`. So the moment
the hook knows the most about what shipped is precisely the moment its hashes
are least likely to survive.

Surfacing `BASELINE..HEAD` in the nudge would therefore manufacture exactly the
failure BRAG.md warns against — a hash that reads as evidence and resolves for
nobody — and it would do it *automatically*, at scale, which is worse than a
human occasionally getting it wrong by hand.

The hook may **remind** ("if this work has a PR, record `pr:<n>`"). It must not
propose a hash it happens to know.

### LD4 — Adoption is the deliverable, so the baseline goes in the record.

Zero at 2026-08-13, against a corpus of 361→363 the same day. The count is
incidental and moves; the zero is the claim. The spec is not "done" because the
words shipped;
it is answerable when the share of *new* entries carrying an evidence link can
be stated against that baseline. Verification re-measures rather than assuming.

## Outputs

- **Files modified:** `internal/mcpserver/server.go` (`brag_add` description),
  `plugin/commands/brag.md`, `plugin/hooks/capture-nudge.sh` (nudge text only),
  `docs/for-ai-agents.md`, `scripts/test-docs.sh` (W5/W6) and
  `scripts/test-capture-nudge.sh` (H8/H9a/H9b).
- **Files modified, untracked:** `.claude/skills/brag-capture/SKILL.md` — the
  FIFTH capture-time surface, found at verify. `.claude/` is gitignored, so this
  cannot ship as a tracked file; it is local-only and must be applied wherever
  the measurement will be taken, or LD4's number is uninterpretable.
- **Files created:** none.
- **New exports:** none.
- **Database changes:** none.

## Acceptance Criteria

- [x] `brag_add`'s MCP `Description` names the evidence-link convention and the
      `pr:`-first preference — the one string present in the agent's context for
      every call.
- [x] `/brag` and the Stop-hook nudge text both mention it.
- [x] `docs/for-ai-agents.md` carries the ref-preference order, matching BRAG.md
      rather than paraphrasing it.
- [x] **The hook proposes no hash** (LD3) — verified by reading its emitted
      text, not just its source.
- [x] Nothing about existing entries changes; the hand-typed convention keeps
      working identically.
- [x] The adoption baseline (0 evidence tags at 2026-08-13) is recorded in this spec's
      Build Completion, with the measuring command, so verification re-runs it.
- [x] Full gate set green.

## Failing Tests

- **`scripts/test-docs.sh`**
  - `"brag_add's tool description mentions evidence links"` — asserts the
    literal `pr:` guidance appears in the description string in
    `internal/mcpserver/server.go`. Fails today: the description is one
    sentence about a non-empty title.
  - `"the agent-facing docs carry the ref-preference order"` — asserts
    `docs/for-ai-agents.md` names `pr:` ahead of `commit:`. Fails today.
- **`scripts/test-capture-nudge.sh`**
  - `"the nudge mentions evidence links"` — fails today.
  - `"the nudge does NOT contain a commit hash"` (LD3) — asserts the emitted
    text carries no 7+ hex-char token. Passes today for the wrong reason (the
    nudge says nothing at all), so it must be written to fail against a
    deliberately hash-injected variant before it is trusted.

## Implementation Context

### Decisions that apply

- `DEC-024` — `agent:`/`model:` stamped, not hand-typed. The precedent this
  spec is *deliberately not yet following*, and the fallback if instruction
  fails.
- `DEC-046` — `capture.ReservedTagPrefixes` is single-sourced and
  mutation-pinned, so promoting `commit:`/`pr:` to reserved prefixes later is a
  solved problem. Not done here (LD1).
- `DEC-025` / `SPEC-041` — the Stop hook's contract. This spec touches its
  **text**, never its firing logic.

### Out of scope

- Automatic stamping, reserved-prefix promotion, validation of hash shape, and
  any verification surface (`brag verify`) — all STAGE-020's later specs.
- The git-import miner. Forge APIs. Signing.

## Notes for the Implementer

The `brag_add` description is the highest-leverage string in this spec: it is
read on every tool call by the agent that writes ~70% of this corpus. It is also
the most constrained — it must stay a *description*, not become a manual. One
added clause naming `pr:` and pointing at BRAG.md is the whole job.

Resist the hook change growing. It knows `BASELINE` and `HEAD`; LD3 exists
because that knowledge is a trap, not an asset, under squash-merge.

## Build Completion

**Four surfaces changed, no code behaviour touched.**

| surface | change |
|---|---|
| `internal/mcpserver/server.go` | `brag_add`'s `Description` — the string an agent reads on **every** call — now names the evidence-link convention and the `pr:`-first order |
| `plugin/commands/brag.md` | `/brag` carries the same guidance, next to the existing `agent:`/`model:` instruction |
| `plugin/hooks/capture-nudge.sh` | nudge text only; the firing logic is untouched |
| `docs/for-ai-agents.md` | new §6 subsection with the full preference order and the squash rationale |

### Adoption baseline (LD4)

Measured at build, with the command verification must re-run:

```bash
brag tags | grep -cE '^(commit|pr|issue):'      # → 0
```

**0 evidence tags**, corpus 363 entries, 2026-08-13. For contrast on the same
corpus and the same tag mechanism: `agent:claude-code` **70**,
`model:claude-opus-4-8` **56** — stamped, never typed. That contrast is the
spec's whole premise, and the number to beat is any value above zero.

### The LD3 test needed its own mutation check, and the first one was wrong

H9 asserts the nudge proposes no commit hash. It passed before any change was
made — trivially, because the nudge said nothing — so the spec required proving
it could fail before trusting it.

The **first** mutation was invalid: injecting `$HEAD` into the nudge string
changed nothing, because that text lives inside a single-quoted `jq` program
where `$HEAD` is literal. H9 stayed green and briefly looked toothless when in
fact the mutation had not happened. Re-done the way the hook actually
interpolates — `--arg head "$HEAD"` plus `\(\$head)`, matching how
`session:` is already threaded — H9 failed correctly:

```
FAIL: H9: nudge must not propose a commit hash (LD3); found: ff626df772f3be…
```

Worth recording because it is a new variant of this stage's recurring lesson: a
mutation that does not actually mutate produces the same green as a test with no
teeth, and the two are indistinguishable unless you check that the mutant
behaves differently.

### Deviations from the design

**One, and it was not recorded until verify caught it.** The Notes set an
explicit budget for the `brag_add` description: *"it must stay a description,
not become a manual. One added clause naming `pr:` and pointing at BRAG.md is
the whole job."* The build shipped **three added sentences, 421 chars / 71
words** — about 6× the budget, and 4× the longest sibling description
(`brag_memory`, 106 chars). It also carried repo-specific squash mechanics to
every MCP client on earth.

Trimmed at verify to **235 chars / 39 words**: the preference order and the
BRAG.md pointer, with the mechanics left to BRAG.md where they are already
conditioned correctly. Still ~2× `brag_memory`, which is defensible for the one
tool that writes, but no longer a manual.

The related free improvement was taken as the trade: the `tags` **field**
description now carries a pointer (*"evidence links go here — see the tool
description"*), which is cheaper than description length and closer to the
point of use.

LD1 itself held — nothing stamped, no schema moved, no validation added.

### Build-phase reflection (3 questions, short answers)

1. **What surprised you?** How much of the work was already written. BRAG.md had
   the full rationale; the job was moving it to where it is read.

   *(Corrected at verify: this answer originally said the `brag_add` description
   "went from 9 words to a sentence". It went from 9 words to **71 words across
   5 sentences** — understating its own change by roughly 6×, in the one
   sentence describing the spec's highest-leverage edit. Exactly the defect class
   this stage keeps producing, and exactly why it was over the Notes' budget
   without anyone noticing. Now 39 words after the verify trim.)*

2. **What was harder than expected?** Proving H9 could fail. Two attempts, and
   the first failure mode was invisible: a mutation that silently did nothing.

3. **What would you tell the next implementer?** Do not add stamping until the
   baseline is re-measured. The whole design rests on the claim that instruction
   was never delivered — if adoption moves, the expensive spec is unnecessary,
   and if it does not, that is a *measured* result rather than an assumption.

## Reflection (Ship)

1. **What would I do differently next time?**
   — Write acceptance criteria that an assertion can actually hold. AC #3 said
   the agent docs must carry the ref-preference order *"matching BRAG.md rather
   than paraphrasing it"* — and nothing tested the second half, so the build
   paraphrased on all four surfaces in four different wordings. That is how the
   unconditional squash claim reached the two files that ship to other people's
   repos. **An AC phrased as "matching X rather than paraphrasing it" needs an
   assertion that compares the two texts, or it is decoration.** W6 also passed
   on the wrong evidence: its first `pr:` hit is an incidental `--tag pr:151`
   filtering example, not the preference list, so it would pass on a doc saying
   *"never use `pr:`"*.

2. **Does any template, constraint, or decision need updating?**
   — No template change, but one lesson generalises beyond this spec and belongs
   in the stage record: **a test can be green, have teeth, and still pin the
   wrong proposition.** H9 asserted "the emitted text contains no hex-ish token",
   which is not LD3's property. It had teeth (an injected hash failed it) and was
   still false in production, because a real session id is a UUID whose segments
   are hex — it stayed green only because the fixture used `session-1`. The
   mutation check proved the regex catches an *injected* hash; nobody checked
   what it already caught in the text that was there. Fixed by asserting the
   precise property (`$HEAD`/`$BASELINE` absent) with the generic sweep demoted
   to a backstop, and by making the fixture UUID-shaped so the case cannot hide.

3. **Is there a follow-up spec I should write now before I forget?**
   — No new spec. But the gate on this stage's second spec needs a **date**, not
   a condition: "only if adoption stays near zero" with no re-measure window is a
   gate nobody pulls. Recorded in STAGE-020's backlog as **2026-09-14** (~1 month,
   ~90 new entries at the observed ~22/week).

4. **What can a user do now that they couldn't before?** — one sentence,
   before → after; quote the confirming number if one exists, name the outcome
   if not. Write `none` if this spec has no user-visible outcome — that is a
   real, greppable result, not a blank. This is the line a brag's `impact` field
   is transcribed from, and both halves are already written above (## Context is
   the before, ## Goal is the after): confirm the prediction, don't reconstruct
   it from memory.
   — Before, an agent capturing a brag was never told that evidence links exist:
   the convention worked but appeared at **none** of the five surfaces an agent
   reads at capture time, and **0 of 363** entries carried one. Now the
   `brag_add` tool description, `/brag`, the session-end nudge, the agent docs
   and the capture checklist all say to record `pr:<n>` (or a default-branch
   `commit:`) — so a claim can be checked by anyone with the repo, and whether
   instruction alone is enough becomes a measurement due 2026-09-14 rather than
   an assumption.
