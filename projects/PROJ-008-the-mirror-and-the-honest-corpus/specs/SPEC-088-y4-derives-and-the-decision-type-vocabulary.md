---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-088
  type: chore                      # epic | story | task | bug | chore
  cycle: frame                     # frame | design | build | verify | ship
                                   # NOT YET FRAMED. This file was created at
                                   # SPEC-087 ship to CLAIM the id, because
                                   # eleven prose references already pointed at
                                   # SPEC-088 while `scripts/_lib.sh:107-119`
                                   # derives next_id from FILENAMES — so the
                                   # next `just new-spec` would have handed 088
                                   # to unrelated work and silently redirected
                                   # every one of them.
  blocked: false
  priority: medium                 # not urgent: Y4's literal pin is correct
                                   # today and the type gap hard-fails loudly.
  complexity: S                    # PROVISIONAL — framing re-checks. Item 2
                                   # touches a user-facing table, which is the
                                   # half that could push this past S.

project:
  id: PROJ-008
  stage: STAGE-023
repo:
  id: bragfile

agents:
  architect: claude-opus-5
  implementer: claude-opus-5       # usually same Claude, different session
  created_at: 2026-09-07

references:
  decisions: []                    # DEC-050 is spoken for by SPEC-086.
  constraints:
    - one-spec-per-pr
  related_specs:
    - SPEC-087                     # routed both items here (LD6, V-F3);
                                   # its `inv_row` helper is the tool item 1
                                   # would reuse
    - SPEC-082                     # authored Z7, which is where item 2's
                                   # hard-fail arrived
    - SPEC-080                     # authored Y4's pin and the questions rows
---

# SPEC-088: Y4 derives and the decision-type vocabulary

## Context

> **Cycle: frame — not yet framed.** This file exists so the id is *claimed*
> rather than *reserved in prose*. Everything below is transcribed evidence
> routed here by SPEC-087; nothing here is a decision. Framing decides GO/NO-GO,
> the fork in item 2, and whether the two items belong in one spec at all.

Two items were measured by SPEC-087 and deliberately not decided by it. Both
read the same two files (`scripts/test-docs.sh`, `scripts/inventory.sh`), which
is why they were routed to one owner.

**Item 1 — `Y4` still caches two literals (SPEC-087 LD6).** `Y4` pins
`Questions tracked … | 21 |` and `… still open | 8 |` against
`guidance/questions.yaml` by `grep -F`. SPEC-087 made `Y3` derive and left `Y4`
alone, on measurement rather than preference: framing had scoped `Y4` in *"only
if Fork 2 makes it free"*, Fork 2 was rejected, and the equivalent oracle is a
YAML parse of a register whose entries are hand-written prose — a materially
different problem from counting files in a directory. Taking it inside SPEC-087
would have pushed that spec past **S**.

What SPEC-087 leaves behind for this one: `inv_row()` in `scripts/test-docs.sh`
already pulls a Value cell out of an emitted `scripts/inventory.sh` row by its
exact What-column label, and `Y3`/`Z7` demonstrate the shape — compare the
**emitted** number against an oracle in a *different mechanical class* from the
producer's own filter. The open question `Y4` inherits is what that independent
oracle is for a hand-written YAML register, and whether one exists that is not
just `inventory.sh`'s two `grep -cE` expressions copied.

**Item 2 — the decision template advertises five `insight.type` values; the
inventory tolerates two (SPEC-087 V-F3).** `decisions/_template.md` offers
`decision | analysis | recommendation | observation | reservation`.
`scripts/inventory.sh` has a row for `decision` and one for `reservation` only.
A `DEC-*.md` carrying any of the other three is counted by **neither** row,
vanishes from the page, and hard-fails `Z7`. Measured on `main` at `f4658b0`
with a `type: analysis` stub (simulation S-3):

```
FAIL: Z7: the inventory covers 49 of 50 decisions/DEC-*.md files (48 decision + 1 reservation).
```

**Dated deliberately: this predates SPEC-087.** It arrived with `Z7` at
SPEC-082 and fires identically on `main`, so it is not a SPEC-087 regression.
The hard fail is the *correct* behaviour — an uncounted decision is invisible
and the page under-reports. What is wrong is that the template is where an
author picks the value, and it currently invites three values that break the
harness. SPEC-087 verify added a warning line to `decisions/_template.md`
naming the consequence; that line is a warning, **not** the decision.

**Why this was not filed in `guidance/questions.yaml`.** Filing a question moves
`Questions tracked …` 21 → 22 and `… still open` 8 → 9 — two inventory rows —
which forces a `Y4` re-pin and a regeneration of `docs/engineering-practices.md`.
The routing cost is not a reason to drop a finding, but it is a real reason to
prefer the home that costs nothing. It is also a live demonstration of item 1:
the register is expensive to add to precisely because a literal pins its size.

## Goal

*To be written at framing.* The shape it must take: item 1 is a mechanical
question (find the independent oracle, or record that none exists and say so);
item 2 is a **fork** with a user-facing table on one side of it —

- **(a)** teach `scripts/inventory.sh` three more rows, or
- **(b)** narrow `decisions/_template.md`'s vocabulary to the two the harness
  counts, or
- **(c)** something else framing measures.

Neither branch is chosen here. Framing must also decide whether the two items
are one spec or two — they share files, not a problem.

## Inputs

- **Files to read:** `path/to/file.ext` — why
- **External APIs:** <name, docs link, auth>
- **Related code paths:** `src/some/module/`

## Outputs

- **Files created:** `path/to/new.ext` — purpose
- **Files modified:** `path/to/existing.ext` — what changes
- **New exports:** <names and signatures>
- **Database changes:** <migrations>

## Acceptance Criteria

Testable outcomes. Cover happy path, error cases, edge cases.

- [ ] Criterion 1 (testable)
- [ ] Criterion 2 (testable)

## Failing Tests

Written during **design**, BEFORE build. The implementer's job in
**build** is to make these pass.

- **`path/to/test.file`**
  - `"test description 1"` — asserts: ...

## Implementation Context

*Read this section (and the files it points to) before starting
the build cycle. It is the equivalent of a handoff document, folded
into the spec since there is no separate receiving agent.*

### Decisions that apply

- `DEC-NNN` — <one-line summary of why this matters here>
- `DEC-MMM` — <one-line summary>

### Constraints that apply

These constraints apply to the paths touched by this task (see
`/guidance/constraints.yaml` for full text):

- `constraint-id-1` — <one-line summary>
- `constraint-id-2` — <one-line summary>

### Prior related work

- `SPEC-YYY` (shipped) — <one-line summary, if relevant>
- `PR #NNN` — <link, if relevant>

### Out of scope (for this spec specifically)

Explicit list of what this spec does NOT include. If any of these feel
necessary during build, create a new spec rather than expanding this one.

- ...

## Notes for the Implementer

Gotchas, style preferences, reuse opportunities.

---

## Build Completion

*Filled in at the end of the **build** cycle, before advancing to verify.*

- **Branch:**
- **PR (if applicable):**
- **All acceptance criteria met?** yes/no
- **New decisions emitted:**
  - `DEC-NNN` — <title> (if any)
- **Deviations from spec:**
  - [list]
- **Follow-up work identified:**
  - [any new specs for the stage's backlog]

### Build-phase reflection (3 questions, short answers)

Process-focused: how did the build go? What friction did the spec create?

1. **What was unclear in the spec that slowed you down?**
   — <answer>

2. **Was there a constraint or decision that should have been listed but wasn't?**
   — <answer>

3. **If you did this task again, what would you do differently?**
   — <answer>

---

## Reflection (Ship)

*Appended during the **ship** cycle. Outcome-focused reflection, distinct
from the process-focused build reflection above.*

1. **What would I do differently next time?**
   — <answer>

2. **Does any template, constraint, or decision need updating?**
   — <answer>

3. **Is there a follow-up spec I should write now before I forget?**
   — <answer>

4. **What can a user do now that they couldn't before?** — one sentence,
   before → after; quote the confirming number if one exists, name the outcome
   if not. Write `none` if this spec has no user-visible outcome — that is a
   real, greppable result, not a blank. This is the line a brag's `impact` field
   is transcribed from, and both halves are already written above (## Context is
   the before, ## Goal is the after): confirm the prediction, don't reconstruct
   it from memory.
   **If this answer is not `none`, capture it before closing the cycle** — the
   sentence is the deliverable, and an uncaptured one decays into a
   reconstruction. Evidence ref: under `one-spec-per-pr` this spec has exactly
   one PR by construction, so tag `pr:<n>` rather than a commit hash (a
   squash-merge destroys the branch commit you were looking at).
   — <answer>
