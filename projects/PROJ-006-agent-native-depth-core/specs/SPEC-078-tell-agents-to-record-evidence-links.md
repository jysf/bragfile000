---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-078
  type: story                      # epic | story | task | bug | chore
  cycle: design                    # frame | design | build | verify | ship
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
whole corpus** (361 entries, measured 2026-08-13).

STAGE-020's first framing read that zero as friction — hand-typing a hash is
work nobody will do. That was wrong, and checking rather than assuming produced
a much cheaper spec:

| surface an agent reads at capture time | mentions evidence links? |
|---|---|
| MCP `brag_add` tool `Description` (read on **every** call) | **no** — it is one sentence: *"Capture a new brag entry. Requires a non-empty title."* |
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

Zero of 361 at 2026-08-13. The spec is not "done" because the words shipped;
it is answerable when the share of *new* entries carrying an evidence link can
be stated against that baseline. Verification re-measures rather than assuming.

## Outputs

- **Files modified:** `internal/mcpserver/server.go` (`brag_add` description),
  `plugin/commands/brag.md`, `plugin/hooks/capture-nudge.sh` (nudge text only),
  `docs/for-ai-agents.md`, and `scripts/test-docs.sh` for the assertions below.
- **Files created:** none.
- **New exports:** none.
- **Database changes:** none.

## Acceptance Criteria

- [ ] `brag_add`'s MCP `Description` names the evidence-link convention and the
      `pr:`-first preference, in the one string an agent reads on every call.
- [ ] `/brag` and the Stop-hook nudge text both mention it.
- [ ] `docs/for-ai-agents.md` carries the ref-preference order, matching BRAG.md
      rather than paraphrasing it.
- [ ] **The hook proposes no hash** (LD3) — verified by reading its emitted
      text, not just its source.
- [ ] Nothing about existing entries changes; the hand-typed convention keeps
      working identically.
- [ ] The adoption baseline (0 / 361 at 2026-08-13) is recorded in this spec's
      Build Completion, with the measuring command, so verification re-runs it.
- [ ] Full gate set green.

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

*Filled in during the **build** cycle. Must record the adoption baseline and the
exact command used to measure it (LD4).*

## Reflection (Ship)

*Appended during the **ship** cycle.*
