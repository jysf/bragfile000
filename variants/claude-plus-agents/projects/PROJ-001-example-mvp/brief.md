---
project:
  id: PROJ-001
  status: active
  priority: high
  target_ship: 2026-06-01

repo:
  id: my-app

created_at: 2026-04-19
shipped_at: null
---

# PROJ-001: MVP (example)

> **Note:** This is an example project included to show the format.
> Delete this entire `PROJ-001-example-mvp/` folder once you've created
> your own first project with `just new-project` (or manually).

## What This Project Is

Deliver the first usable version of the app — just enough to put in front
of real users and learn. Foundational infrastructure, a single core user
flow end-to-end, and basic observability. Nothing more.

## Why Now

This is the first project. Nothing else can ship until the app exists in
a usable form. The goal is to hit "usable" as fast as possible and then
learn from real usage what's worth building next.

## Success Criteria

- A real user can sign up, complete the primary flow, and come back the
  next day without hand-holding.
- The service runs in production under a real domain, with basic
  monitoring and a rollback path.
- Every error a user sees is logged with enough context to reproduce.

## Scope

### In scope
- Foundational infrastructure (STAGE-001)
- Auth + single primary user flow (STAGE-002)
- Production deploy + basic observability (STAGE-003)

### Explicitly out of scope
- Secondary user flows (deferred to PROJ-002)
- Admin/ops tooling beyond the minimum
- Any performance optimization that isn't actively blocking users
- Multi-region / HA anything

## Stage Plan

- [ ] STAGE-001 (active) — Foundational infrastructure (logger, errors, env, health)
- [ ] (not yet defined) — Auth + primary flow
- [ ] (not yet defined) — Deploy + observability

**Count:** 0 shipped / 1 active / 2 pending

## Dependencies

### Depends on
- None. This is the first project.

### Enables
- PROJ-002 (improvements) — once MVP is in real usage, we'll know what to improve.

## Project-Level Reflection

*To be filled in when this project ships.*

- **Did we deliver the outcome?** <not yet>
- **How many stages did it actually take?** <not yet>
- **What changed between starting and shipping?** <not yet>
- **Lessons that should update AGENTS.md, templates, or constraints?** <not yet>
- **What did we defer to the next project?** <not yet>
