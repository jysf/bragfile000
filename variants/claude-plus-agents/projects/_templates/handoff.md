---
# Maps to ContextCore handoff.* semantic conventions.

handoff:
  id: HANDOFF-XXX
  from_agent: claude-opus-4-7
  to_agent: kilo-code              # or factory-droid, adal, etc.
  from_role: architect
  to_role: implementer
  created_at: YYYY-MM-DD
  status: pending                  # pending | accepted | completed | rejected

task:
  spec_id: SPEC-XXX

project:
  id: PROJ-XXX
  stage: STAGE-XXX
repo:
  id: my-app
---

# HANDOFF-XXX: <Task Title — same as the spec's title>

## Delegation Summary

One sentence: `<from_agent>` (acting as `<from_role>`) hands `SPEC-XXX`
to `<to_agent>` (acting as `<to_role>`) for implementation during the
**build** cycle.

## Context the Receiving Agent Needs

The receiving agent MUST read these before starting work. Keep the
list tight, but don't omit anything necessary.

### Primary

- **Project brief:** `./projects/PROJ-XXX-<slug>/brief.md`
- **Stage:** `./projects/PROJ-XXX-<slug>/stages/STAGE-XXX-<slug>.md`
- **Spec:** `./projects/PROJ-XXX-<slug>/specs/SPEC-XXX-<slug>.md`

### Decisions that apply

- `DEC-NNN` — <one-line summary of why this matters here>
- `DEC-MMM` — <one-line summary>

### Constraints that apply

Check `./guidance/constraints.yaml` for full text. These constraints
apply to the paths touched by this task:

- `constraint-id-1` — <one-line summary>
- `constraint-id-2` — <one-line summary>

### Prior related work

- `HANDOFF-YYY` — <one-line summary, if relevant>
- `PR #NNN` — <link, if relevant>

## Expected Deliverables

- Code changes implementing SPEC-XXX's Acceptance Criteria.
- All failing tests in SPEC-XXX now passing.
- Any new tests required to cover edge cases.
- A PR against `main` from branch `feat/spec-XXX-<slug>`.
- PR description referencing: this handoff ID, the spec ID, the stage
  ID, the project ID, all referenced `DEC-*`, and any new `DEC-*`
  created during implementation.

## Out of Scope

Explicit list of what this handoff does NOT include. If the implementer
thinks any of these need to happen, they should create a new spec in
the stage's backlog, not expand this handoff.

- ...

## Return Criteria

The implementer signals completion by:
1. Filling in the `## Completion` section (including reflection).
2. Updating `handoff.status` → `completed`.
3. Updating the spec's `task.cycle` → `verify` (or use `just advance-cycle SPEC-XXX verify`).
4. Opening a PR.

If the implementer cannot complete the task:
1. Fill in the `## Completion` section with what was done and what blocked.
2. Update `handoff.status` → `rejected`.
3. Set the spec's `task.blocked: true` and add a question to
   `/guidance/questions.yaml`.

---

## Completion

*Filled in by the receiving agent when the handoff is complete. The
three reflection questions below are part of completion, not optional.*

### Execution notes

- **PR:** [link]
- **Completed at:** YYYY-MM-DD
- **All acceptance criteria met?** yes/no (if no, explain)

### Drift and new artifacts

- **New decisions emitted:**
  - `DEC-NNN` — <title> (if any)
- **Deviations from spec:**
  - [list]
- **Follow-up work identified:**
  - [any new specs that should be added to the stage's backlog]

### Implementer reflection (3 questions, short answers)

1. **What was unclear in the spec or handoff that slowed you down?**
   — <answer>

2. **Was there a constraint or decision that should have been listed but wasn't?**
   — <answer>

3. **If you did this task again, what would you do differently?**
   — <answer>
