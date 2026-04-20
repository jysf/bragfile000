---
stage:
  id: STAGE-001
  status: active
  priority: high
  target_complete: 2026-05-03

project:
  id: PROJ-001
repo:
  id: my-app

created_at: 2026-04-19
shipped_at: null
---

# STAGE-001: Foundational infrastructure

> **Note:** This is an example stage. Delete once you've created your own.

## What This Stage Is

The bare-minimum infrastructure the app needs before any user-facing
logic is safe to build: a logger, typed error classes, environment-variable
loading, and a basic health-check endpoint. When this stage ships, every
subsequent stage can import and rely on these foundations rather than
re-deciding them spec-by-spec.

## Why Now

Everything in PROJ-001 depends on it. Without it, every subsequent spec
would reinvent logging, error handling, and env loading, and we'd get
drift immediately.

## Success Criteria

- A new engineer (or agent) can clone the repo, run `just dev` (or the
  app's dev command), and see a running service with a `/health`
  endpoint returning 200.
- All subsequent specs import logging and error handling from modules
  created here, not rolled inline.
- Local dev and production run the same code — only config differs.

## Scope

### In scope
- Logger module and conventions
- Typed error classes + error-handling middleware
- Env-var loader with validation
- Health-check endpoint
- CI pipeline (lint + typecheck + test on every PR)

### Explicitly out of scope
- Authentication (STAGE-002)
- Database / persistence (STAGE-002)
- Any user-facing functionality
- Metrics / traces beyond basic logging (STAGE-003)

## Spec Backlog

- [ ] SPEC-001 (design) — Logger module (pino wrapper with redaction)
- [ ] (not yet written) — Typed error classes and error-handling middleware
- [ ] (not yet written) — Env-var loader with schema validation
- [ ] (not yet written) — Health check endpoint
- [ ] (not yet written) — CI pipeline (lint, typecheck, test)

**Count:** 0 shipped / 1 active / 4 pending

## Design Notes

- All foundational modules live under `src/lib/` or `src/config/`.
- Each module has zero runtime dependencies on the others where possible
  (logger depends on nothing; errors depend on nothing; env depends on
  Zod; health depends on all three).
- Naming: foundational modules use short noun-shaped names (`log`,
  `env`, `errors`, `health`). Business modules elsewhere may be
  verb-shaped.

See `DEC-001` for logger-specific rationale.

## Dependencies

### Depends on
- None. This is the foundational stage.

### Enables
- Every subsequent stage.

## Stage-Level Reflection

*To be filled in when this stage ships.*

- **Did we deliver the outcome?** <not yet>
- **How many specs did it actually take?** <not yet>
- **What changed?** <not yet>
- **Lessons?** <not yet>
