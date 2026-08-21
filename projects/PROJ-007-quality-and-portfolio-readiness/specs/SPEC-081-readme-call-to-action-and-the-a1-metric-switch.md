---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-081
  type: chore                      # epic | story | task | bug | chore
  cycle: frame                     # frame | design | build | verify | ship
  blocked: false
  priority: medium
  complexity: S                    # S | M | L  (L means split it)

project:
  id: PROJ-007
  stage: STAGE-021
repo:
  id: bragfile

agents:
  architect: claude-opus-5
  implementer: claude-opus-5       # usually same Claude, different session
  created_at: 2026-08-19

references:
  decisions: []
  constraints: []
  related_specs:
    - SPEC-021                     # the README rewrite that created test-docs and set A1's band
    - SPEC-079                     # added the `## How this repo is built` section; re-pinned A1 to 260
    - SPEC-080                     # the stage's other legibility spec
---

# SPEC-081: the README call-to-action, and A1 measured in words

> **Cycle: frame.** This is a **go/no-go**, not a design. It records the
> measurement, states the two changes, and hands design the decisions. Per
> AGENTS.md §6, design happens in a **fresh session**.

## Context

STAGE-021's third and final spec, and the one that closes the stage.

Two changes that belong together because they touch the same file and the second
is what makes the first affordable.

Parent: `STAGE-021-make-the-discipline-legible`, spec 3 of 3.
Project: `PROJ-007`.

## Goal

An agent-using developer landing on the README sees the one command that makes
`brag` useful to their agent **before** they scroll, and the guard protecting the
README's length measures the thing it actually cares about.

## Change 1 — promote the call-to-action out of the Status blockquote

The CTA already exists. It is at `README.md:16-19`, **inside a blockquote whose
first word is "Status:"**:

> Working with an AI agent? `brag mcp install` wires brag into Claude Code,
> Cursor, or Claude Desktop as five typed tools plus an auto-loadable memory
> resource — see [Using brag from an AI agent (MCP)](#using-brag-from-an-ai-agent-mcp) below.

That is good copy in the wrong container. Readers skip status blocks, and the
full MCP section sits at **line 219 of 260** — near the floor.

**Why this is worth doing, and the honest limit of the argument.** The PROJ-005
synthesis calls corpus-as-agent-memory *"the single highest leverage/effort
item"* and says it reframes bragfile from capture-sink to agent
**infrastructure**; local-first/developer-owned is named as the one lane
incumbents structurally cannot follow. If that is the differentiator, it should
not be at line 219 behind a status header.

**But nobody has measured that it drives adoption.** bragfile has one human
author and has never been written about anywhere, so there is no signal to read.
This is a reasoned bet, not a validated one, and the spec should say so rather
than implying evidence it does not have.

**The detail work is small** — the copy exists; this is a move plus a heading,
not new writing. The full MCP section stays where it is; detail belongs low.

## Change 2 — A1 measures words, not lines

`scripts/test-docs.sh:117`:

```
assert_line_count_band "A1" "README.md" 100 260
```

Measured 2026-08-19: **README.md is 260 lines / 1,268 words.** A1's ceiling is
260. **Zero headroom, by design** — SPEC-079's LD5 re-pinned it tight rather than
widening it, on the reasoning that a guard widened whenever it fires is not a
guard. That reasoning was right, and it has a cost this spec now pays: **any**
net-positive edit to the README fails the harness, including Change 1.

**The deeper problem is the metric, not the bound.** A1 counts `wc -l` on a
**hard-wrapped** file, so it conflates three different events:

- adding content — the thing the guard exists to catch;
- **rewrapping a paragraph** — no content change at all;
- adding a code fence.

Only the first is what anyone cares about. Rewording one sentence so it wraps to
an extra line turns CI red having added nothing. A guard that fires on non-events
gets disarmed by habit — which is precisely the "bump the number until green"
failure LD5 refused.

**`wc -w` is reflow-invariant.** Rewording costs nothing, moving a section costs
nothing, and only genuinely new content counts. That makes the guard *more*
honest, not looser, and it is what makes Change 1 affordable without touching a
bound.

**On a reasonable README length:** 1,268 words is about a five-minute read — the
long side of healthy but not bloated. For a CLI with install, quickstart and
pointers, roughly **900–1,400 words** is where it still gets read; past ~1,500
people scan headings and leave. The README is not too long today. A1 has been
measuring the wrong thing.

## Decisions for the design session

1. **What band?** The number must be chosen and justified, not inherited.
   A ceiling at the current 1,268 repeats LD5's zero-headroom trade in a new
   unit and blocks Change 1 immediately; a ceiling with room invites the sprawl
   LD5 rejected. Note the two are not symmetric — a word budget tolerates
   headroom better than a line budget, because reflow can no longer consume it.
   State the floor too, and say what each end is protecting against.
2. **Add a sibling helper — do not repurpose the existing one.** Checked at
   framing: `assert_line_count_band` has **six callers** — `A1` (README),
   `C2` (CONTRIBUTING), `D2` (development.md), `J2` (the slash-command),
   `T2` (the agent doc) and `X7` (the practices page). Changing its behaviour
   would silently re-scope five unrelated guards. So the shape is a new
   `assert_word_count_band` beside it, with A1 the only caller switched.
   Design still decides: whether A1 **keeps its id** (the practices page and
   several commit messages reference "A1" by name), and whether any of the
   other five deserve the same treatment — they are all hard-wrapped prose
   files with the same reflow problem. Resist scope creep here; note it and
   move on unless one is trivially in the way.
3. **Where does the CTA go, and how loud?** Immediately after `## Install` is the
   obvious slot — the reader has just learned how to get it. Decide whether it is
   a named section or a lead-in paragraph, and what it must *not* duplicate from
   the section at line 219.
4. **The Status blockquote is pinned — work with that, not around it.**
   Checked at framing: `W3` greps README for the literal
   `^> \*\*Status:\*\* v<latest> shipped`, deriving `<latest>` from the
   CHANGELOG (`scripts/test-docs.sh:1280`). **That exact line and its blockquote
   prefix are load-bearing** — this is the guard SPEC-077 added after the version
   line rotted through two releases, so it is not a candidate for removal.
   Design decides what the blockquote keeps once the CTA leaves, subject to that
   line surviving verbatim. Confirm `W3` green after any edit to it.

## Out of scope

- Moving or rewriting the full MCP section at line 219. It stays.
- Any change to `docs/engineering-practices.md`'s guarded inventory block, or to
  `scripts/inventory.sh`. If a change moves a derived number, run `just
  inventory` and paste — the script is the source.
- Anything in STAGE-022 (lint, coverage, the `Entries:` envelope, and the four
  items routed there by SPEC-080).
- `internal/storage/window.go:12-13` (V2 from SPEC-080's re-verify — the window
  flags are shared by `coverage` too). Recorded there, unrouted. **A one-line
  prose fix; fold it in only if design judges it in scope, and say so.**

## Go / no-go

**GO.** Complexity **S**. The copy exists, the metric switch is a helper call and
a band, and both changes are confined to `README.md` and `scripts/test-docs.sh`.

**Why now:** it is the stage's last backlog item, and it is sequenced here
deliberately — SPEC-079 and SPEC-080 both had to land first because A1's zero
headroom meant any concurrent README edit would trip them.

**What would make this a no-go:** if Change 2 were framed as "raise the ceiling
so Change 1 fits." That is the bump-until-green move LD5 refused, and it would
make this spec the thing that disarmed a working guard. The justification has to
be that **`wc -l` measures the wrong event** — which stands on its own, whether
or not Change 1 exists.
