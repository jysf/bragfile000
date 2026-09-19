---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-094
  type: story                      # epic | story | task | bug | chore
  cycle: frame                     # frame | design | build | verify | ship
                                   # Created at SPEC-086 design (2026-09-18) to
                                   # CLAIM the id in the same edit as the STAGE-023
                                   # line that routes work here. It is not framed:
                                   # it carries the evidence and the posture
                                   # DEC-050 already decided, and owes its own
                                   # framing pass.
  blocked: false
  priority: high                   # every failure written before this ships is
                                   # listed as a win on two more surfaces
  complexity: M                    # provisional — framing owns it. Two surfaces,
                                   # and `story` carries a decision of its own.

project:
  id: PROJ-008
  stage: STAGE-023
repo:
  id: bragfile

agents:
  architect: claude-opus-5
  implementer: claude-opus-5       # usually same Claude, different session
  created_at: 2026-09-18

references:
  decisions:
    - DEC-050                      # rows 3 and 4 are this spec's posture; it
                                   # implements them, it does not re-decide them
    - DEC-029                      # profiles are data — why story's mechanics
                                   # are a decision, not a line of code
    - DEC-014                      # the envelope, incl. part 4's empty-state rule
    - DEC-048                      # a count names what it counted
  constraints:
    - one-spec-per-pr
  related_specs:
    - SPEC-086                     # split this out; implements DEC-050 rows 1-2
---

# SPEC-094: summary and story stop listing a failure as a win

## Context

> **Cycle: frame.** This file exists so the id is claimed by a file rather
> than reserved in prose (SPEC-087 LD6). It was created at SPEC-086's design
> in the same edit as the STAGE-023 backlog line that points here. Nothing
> below is designed.
>
> Measured 2026-09-18 against a frozen copy of the live corpus: 606 entries,
> 4 `type: failed` (ids 420, 433, 465, 473), all four with an impact. Binary
> built from `main` at `3201f50`.

This is the other half of SPEC-086's Fork B. SPEC-086 was split along the
defect shape that framing measured:

- **`impact` and `wrapped` could not represent a failure as one,** in either
  format. DEC-028's 4-key projection carries no `type`. SPEC-086 fixes those
  two surfaces.
- **`summary` and `story` already carry the data and drop it only in
  markdown.** That is this spec.

**The posture is decided.** DEC-050 rows 3 and 4 state it, and this spec
implements it:

| DEC-050 row | Surface | Posture |
|---|---|---|
| 3 | `brag summary` | A section, pulled out of `## Highlights`: `## What didn't work`, rendered only when it has an entry. JSON key `failures_by_project`, because highlights group by project. `By type` is already honest and stays. |
| 4 | `brag story` | Labelled on every profile, and never dropped. A promotional profile may not remove failures from the bundle. |

## What was measured

**`summary`** (`brag summary --range month`):

- `## Summary → By type` prints `- failed: 4`, and JSON `counts_by_type`
  carries `"failed": 4`. That part is honest.
- `## Highlights` lists all four as `- <id>: <title>`, with no type. The JSON
  highlight entry is `["id","title"]`, which is 2 keys, narrower than
  `impact`'s 4.

**`story`** (`--year`, all four bundled profiles):

- **Every profile** renders each failure as `- ★ <id>: <title>`. Measured
  with `grep -c -E '(473|465|433|420): '`: `exec` 4, `skip` 4, `manager` 4,
  `me` 4.
- The markdown carries no type. The JSON does:
  `.threads[].beats[] | select(.type=="failed")` returns `[420, 433, 465,
  473]` under `--audience exec`.
- `exec` and `skip` are `candor: promotional`, and their directives tell a
  model to promote the beats. `manager` and `me` are `candid`.

## The decision this spec still owes

DEC-050 row 4 leaves `story` two questions, because both turn `Candor` from
metadata into a body rule. `internal/story/profile.go:24` documents `Candor`
as *"metadata surfaced to the LLM, not a body rule,"* and DEC-029 choice 2
makes profiles data rather than code.

1. **How a promotional profile presents a labelled failure.** It can stay
   inline in its thread, or move to a separate block in the bundle.
   Dropping it is ruled out by DEC-050.
2. **Whether a failure still counts as an impact beat.** `IsImpactBeat` is
   computed inline at `internal/story/thread.go:135` as `e.Impact != ""`.
   It does not call `aggregate.WithImpact`, despite what DEC-029's text
   says. The answer moves `exec`'s `impact_threads_only` selection, the
   throughline's impact-beat counts, and the JSON `is_impact_beat` field.
   That last one is a count-bearing field, so DEC-048 applies.

If framing finds row 4 cannot be honoured without dropping failures from a
promotional profile, that is DEC-050's revisit trigger T3, and it goes back
to the user.

## Reuse from SPEC-086

- `aggregate.FailureType`, `aggregate.IsFailure` and `aggregate.SplitFailures`
  exist. Use them, and do not restate the predicate.
- The heading literal is `## What didn't work`, and it renders only when
  non-empty (DEC-050 rules 4 and 5).
- The standing traps still apply. A NOT-contains assertion passes on a corpus
  with no failure in it, so pair it with a positive. A green suite is not
  evidence: framing's M-1 survived the whole suite.

## Out of scope

- `impact` and `wrapped`. SPEC-086 covers them.
- `--type` negation, which is a separate `bug` entry on STAGE-023.
- The impact-quality classifier (STAGE-024).
