---
task:
  id: SPEC-001
  type: task
  cycle: design
  blocked: false
  priority: high
  complexity: S

project:
  id: PROJ-001
  stage: STAGE-001
repo:
  id: my-app

handoff:
  from_agent: claude-opus-4-7
  to_agent: kilo-code
  created_at: 2026-04-19

references:
  decisions: [DEC-001]
  constraints:
    - no-secrets-in-code
    - test-before-implementation
    - no-new-top-level-deps-without-decision
  related_specs: []
---

# SPEC-001: Implement the project logger module

> **Note:** Example spec. Delete once you have your own.

## Context

This spec is the first task under `STAGE-001` (foundational infra) in
`PROJ-001` (MVP). `DEC-001` established we'll use a single logger
wrapper around pino. This spec implements that wrapper — effectively
every subsequent module will import it.

## Goal

Create `src/lib/log.ts` exporting a `log` object with `info`, `warn`,
`error`, `debug` methods, wrapping pino with JSON output and a redaction
list for known-sensitive field names.

## Inputs

- **Files to read:**
  - `DEC-001` — the rationale and design constraints.
  - `AGENTS.md` — coding conventions.
  - `STAGE-001` — design notes on module structure.
- **External APIs:** None.
- **Related code paths:** None (first module).

## Outputs

- **Files created:**
  - `src/lib/log.ts`
  - `src/lib/log.test.ts`
- **Files modified:**
  - `package.json` — add `pino`.
- **New exports:**
  - `log.info(msg: string, meta?: Record<string, unknown>): void`
  - `log.warn(msg: string, meta?: Record<string, unknown>): void`
  - `log.error(msg: string, meta?: Record<string, unknown> | Error): void`
  - `log.debug(msg: string, meta?: Record<string, unknown>): void`
- **Database changes:** None.

## Acceptance Criteria

- [ ] Four methods exist with signatures above.
- [ ] Output is valid JSON on each line.
- [ ] `log.error` accepts an Error and serializes name, message, stack.
- [ ] Fields named `password`, `token`, `apiKey`, `authorization`,
      `cookie` are redacted to `"[REDACTED]"` (case-insensitive).
- [ ] Default level is `info`; `LOG_LEVEL=debug` enables debug.
- [ ] No `console.*` calls in committed source.

## Failing Tests

Written during design in `src/lib/log.test.ts`.

- **`src/lib/log.test.ts`**
  - `"emits JSON on info"` — asserts `log.info('hi')` produces JSON
    with `level: "info"` and `msg: "hi"`.
  - `"serializes Error in log.error"` — asserts passing an Error
    yields JSON with `name`, `message`, `stack`.
  - `"redacts sensitive fields"` — asserts `log.info('x', { password: 'p',
    token: 't', safe: 'ok' })` produces `password: "[REDACTED]"`,
    `token: "[REDACTED]"`, `safe: "ok"`.
  - `"respects LOG_LEVEL env"` — asserts debug output appears only
    with `LOG_LEVEL=debug`.

## Non-Goals

- No log shipping / transport.
- No per-module log level overrides.
- No file output or rotation.
- No pretty-printing for dev.

## Notes for the Implementer

- Pino's built-in `redact` option handles redaction; use it.
- Keep the module under 50 lines. This is glue, not a framework.

---

## Reflection

*To be appended during ship.*

1. **What would I do differently next time?** — <not yet shipped>
2. **Does any template, constraint, or decision need updating?** — <not yet shipped>
3. **Is there a follow-up spec to write now?** — <not yet shipped>
