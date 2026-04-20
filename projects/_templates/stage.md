---
# Maps to ContextCore epic-level conventions.
# A Stage is a coherent chunk of work within a Project.
# It has a spec backlog and ships as a unit when the backlog is done.

stage:
  id: STAGE-XXX                     # stable, zero-padded within the project
  status: proposed                  # proposed | active | shipped | cancelled | on_hold
  priority: medium                  # critical | high | medium | low
  target_complete: null             # optional: YYYY-MM-DD

project:
  id: PROJ-XXX                      # parent project
repo:
  id: <repo-id>

created_at: YYYY-MM-DD
shipped_at: null
---

# STAGE-XXX: <Short Title — the coherent outcome>

## What This Stage Is

One paragraph. What coherent outcome does this stage deliver when all
its specs ship? Describe as a capability, not as a list of files.

## Why Now

What makes this the right stage to work on now? Dependencies on other
stages, project priorities, blocking considerations.

## Success Criteria

Concrete outcomes that would mean this stage succeeded.

- ...
- ...

## Scope

### In scope
- ...

### Explicitly out of scope
- ...

## Spec Backlog

Ordered list of specs composing this stage. Add specs as you identify
them. Update status as specs progress.

Format: `- [status] SPEC-ID (cycle) — one-line summary`

- [ ] (not yet written) — <summary of a spec that will be needed>
- [ ] SPEC-NNN (design) — <summary>
- [x] SPEC-MMM (shipped on YYYY-MM-DD) — <summary>

**Count:** 0 shipped / 0 active / 0 pending

## Design Notes

Cross-cutting design decisions that span multiple specs within this
stage but aren't weighty enough for their own DEC file. Keep short.

Link to `DEC-*` entries for the weighty decisions. This section is
for glue: shared patterns, naming conventions, small tradeoffs.

## Dependencies

### Depends on
- STAGE-YYY (in this project or previous) — what it provides
- External: <third-party, vendor, approval>

### Enables
- STAGE-ZZZ — what becomes possible once this ships

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
