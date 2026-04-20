---
# Maps to ContextCore project.* semantic conventions.
# A project is a bounded wave of work against the repo (the app).

project:
  id: PROJ-XXX                      # stable, zero-padded, never reused
  status: proposed                  # proposed | active | shipped | cancelled
  priority: medium                  # critical | high | medium | low
  target_ship: null                 # optional: YYYY-MM-DD

repo:
  id: my-app                        # must match .repo-context.yaml

created_at: YYYY-MM-DD
shipped_at: null
---

# PROJ-XXX: <Short Title — the wave of work>

## What This Project Is

One paragraph. What wave of work is this? If someone asked "what are
you doing this quarter," this paragraph is the answer.

## Why Now

Why this wave, in this order, at this time. If the answer is "it
seemed like a good idea," the project isn't ready.

## Success Criteria

Concrete outcomes that would mean this project succeeded. Not a list
of features — a list of capabilities, metrics, or user-observable
changes.

- ...
- ...
- ...

## Scope

### In scope
- ...
- ...

### Explicitly out of scope
- ...
- ...

## Stage Plan

Ordered list of stages this project will produce. A project typically
has 2–5 stages. Update as work proceeds.

Format: `- [status] STAGE-ID — one-line summary`

- [ ] (not yet defined) — <one-line summary>
- [ ] STAGE-NNN (active) — <summary>
- [x] STAGE-MMM (shipped on YYYY-MM-DD) — <summary>

**Count:** 0 shipped / 0 active / 0 pending

## Dependencies

### Depends on
- External: <third-party API, vendor, approval>
- Previous projects: <PROJ-YYY shipped something this depends on>

### Enables
- Future projects: <what becomes possible after this ships>

## Project-Level Reflection

*Filled in when status moves to shipped.*

- **Did we deliver the outcome in "What This Project Is"?** <yes/no + notes>
- **How many stages did it actually take?** <number, compare to plan>
- **What changed between starting and shipping?** <one or two sentences>
- **Lessons that should update AGENTS.md, templates, or constraints?**
  - <one-line updates>
- **What did we defer to the next project?**
  - <one-line items>
