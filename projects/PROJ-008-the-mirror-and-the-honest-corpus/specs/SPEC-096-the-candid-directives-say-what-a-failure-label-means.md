---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-096
  type: story                      # epic | story | task | bug | chore
  cycle: frame                     # frame | design | build | verify | ship
                                   # Created at SPEC-094 ship (2026-09-23) to
                                   # CLAIM the id in the same edit as the
                                   # STAGE-023 line that routes work here, on the
                                   # maintainer's ruling of the same day. It is
                                   # not framed: it carries SPEC-094 verify's
                                   # measurement and owes its own framing pass.
  blocked: false
  priority: medium                 # does NOT gate v0.7.0; the bundle is already
                                   # honest, and this is about the model's prose
  complexity: S                    # provisional, unframed: two asset lines plus
                                   # a 52-run re-measurement

project:
  id: PROJ-008
  stage: STAGE-023
repo:
  id: bragfile

agents:
  architect: claude-opus-5-5
  implementer: claude-opus-5-5     # usually same Claude, different session
  created_at: 2026-09-23

references:
  decisions:
    - DEC-054                      # T4 fired in part; its ship amendment
                                   # records the measurement this spec starts
                                   # from
    - DEC-050                      # the invariant: a failure is never rendered
                                   # as a win. Holds in the bundle, not always
                                   # in the prose
    - DEC-029                      # directives are data, so this is an asset
                                   # edit, not a code change
  constraints:
    - one-spec-per-pr
  related_specs:
    - SPEC-094                     # shipped the `✗ <id> (failed)` label; its
                                   # LD14 kept the directive assets out of scope
---

# SPEC-096: the candid directives say what a failure label means

## Context

> **Cycle: frame.** This file exists so the id is claimed by a file rather
> than reserved in prose (SPEC-087 LD6). It was created at SPEC-094's ship,
> in the same edit as the STAGE-023 backlog line that points here. Nothing
> below is designed or framed.

SPEC-094 made the two candid `brag story` profiles, `me` and `manager`, label
a recorded failure as `- ✗ <id> (failed): <title>` instead of the `- ★` a win
gets. The bundle is honest. **Neither `internal/story/directives/me.md` nor
`manager.md` tells the model what `✗` means.** They explain `·` (*"no
recorded impact"*) and say nothing about the new marker. SPEC-094's LD14 kept
both assets out of its scope on purpose: the directives are data (DEC-029),
and whether they needed a line was DEC-054 T4's to measure, not a guess to
build.

SPEC-094's verify measured it (V-F2), and on 2026-09-23 the maintainer ruled
that the result is recorded in DEC-054 and the change gets its own id.

## What verify measured

**Method** (SPEC-094 *Verification → V-F2*, in full there). Bundles were
built by the binaries on a frozen copy of the corpus (621 entries, 4
`failed`):
`story --audience <me|manager> --since 2026-09-01 --project bragfile`, which
is 9 beats with 3 failures (433, 465, 473). Each was built on the SPEC-094
binary (`✗`) and on the pre-build binary (`★`, the counterfactual), with a
no-failure control over `--project bragfile-site`. Each run was
`claude -p --model <haiku|sonnet> --tools "" --no-session-persistence --setting-sources ""`
with the tutorial's prompt, *"weave these threads into one headline arc"*.
Five runs per model per bundle and three per model per control bundle made
**52 runs**. Two graders were calibrated on the controls, and both erred, so
every non-absent judgment was read by hand.

| Cell | Model | Framed only as a failure | Credited as a win | Absent |
|---|---|---:|---:|---:|
| `manager`, `✗` | Haiku | 15 | **0** | 0 |
| `manager`, `✗` | Sonnet | 11 | **4** (all #473) | 0 |
| `manager`, `★` | Haiku | 5 | **9** | 1 |
| `manager`, `★` | Sonnet | 2 | **13** | 0 |
| `me`, `✗` | Haiku | 12 | **2** (both #473) | 1 |
| `me`, `✗` | Sonnet | 15 | **0** | 0 |
| `me`, `★` | Haiku | 13 | **0** | 2 |
| `me`, `★` | Sonnet | 14 | **1** (#473) | 0 |

- **On `manager` the label works.** Failures credited as wins fell from
  **22 of 30** mentions to **4 of 30**.
- **On `me` the label is inert.** The directive already asks for *"the messy
  middle"*, and the entries' titles carry the candour. It was 1 of 30 before
  and 2 of 30 after.
- **All six residual cases are #473**, whose recorded impact narrates its own
  fix (*"The guard now derives its keys from the renderer's own output"*).
  Sonnet on `manager` listed it as shipped in 4 of 5 runs, once in the same
  answer that labels 433 and 465 *"(failed)"*.
- **An inverse error, outside DEC-050:** two `✗` runs on `me` called #472, a
  win, a failure. The `★` runs were not audited for it.

So **DEC-050's invariant holds in the bundle and not always in the model's
prose**, when a failure's impact reads like a win. The binary cannot enforce
it downstream. The candidate lever is the directive.

## What this spec would change

One line in `me.md` and one in `manager.md`, telling the model what `✗`
means, in the same place and register as the existing `·` explanation.
Nothing in Go changes.

What framing owes, not answered here:

1. **Whether `me.md` needs the line at all.** On `me` the label is inert, and
   the residue there is 2 of 30. The line may be only for `manager`, or for
   both for symmetry. That is framing's call, on the numbers.
2. **Whether a directive line can move #473.** Its own impact text is the
   cause. A directive might not outweigh an impact that describes a fix, and
   then the remedy is on the data side (how `brag learn` prompts for an
   impact), which is a different spec.
3. **The prompt.** V-F2 used the tutorial's `exec`-shaped *"one headline
   arc"*. A manager-shaped prompt was not measured.

## Constraints this spec carries

- **It edits shipped profile assets**, `internal/story/directives/me.md` and
  `manager.md`, which SPEC-094's LD14 kept out of scope. The `story` goldens
  read `me.md` through `mustDirective` (`internal/story/bundle_test.go`,
  `candor_test.go`), so they follow the asset rather than pin it. Framing
  should check whether anything does pin its text. It is a user-visible
  change to what `brag story --print-directive` prints.
- **It requires re-running the 52-run measurement**, with SPEC-094 V-F2's
  method and controls, on the wording it locks. DEC-054 part 2's rule, that
  its wording is locked by measurement and not by taste, applies to this
  line too. A result that does not beat 4 of 30 on `manager` is a NO-GO for
  the line, not a softer criterion.
- **Window.** `--since 2026-09-01` holds the three failures permanently. The
  frozen-copy method keeps the live corpus unwritten.

## Out of scope

- The `promotional` path (`exec`, `skip`): DEC-054's clause already covers
  it, and T1 is its trigger.
- `summary`: SPEC-095.
- Any change to the `✗` label or to Go.

## Does not gate v0.7.0

**It does not gate v0.7.0.** The maintainer's ruling of 2026-09-23. The
bundle already satisfies DEC-050, and STAGE-023's Success Criterion 4 is about
what the digests render. This spec is about what a model writes from them.
