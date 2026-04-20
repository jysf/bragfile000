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

agents:
  architect: claude-opus-4-7
  implementer: claude-opus-4-7
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

First task under `STAGE-001` (foundational infra) in `PROJ-001` (MVP).
`DEC-001` established we'll use a single logger wrapper around pino.
This spec implements that wrapper.

## Goal

Create `src/lib/log.ts` exporting a `log` object with `info`, `warn`,
`error`, `debug` methods, wrapping pino with JSON output and redaction
for sensitive field names.

## Inputs

- **Files to read:**
  - `DEC-001`
  - `AGENTS.md`
  - `STAGE-001`
- **External APIs:** None.
- **Related code paths:** None.

## Outputs

- **Files created:** `src/lib/log.ts`, `src/lib/log.test.ts`
- **Files modified:** `package.json` (add `pino`)
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
- [ ] `password`, `token`, `apiKey`, `authorization`, `cookie` redacted
      to `"[REDACTED]"` (case-insensitive).
- [ ] Default level `info`; `LOG_LEVEL=debug` enables debug.
- [ ] No `console.*` calls in committed source.

## Failing Tests

- **`src/lib/log.test.ts`**
  - `"emits JSON on info"` — `log.info('hi')` → JSON with `level: "info"`, `msg: "hi"`.
  - `"serializes Error in log.error"` — Error → JSON with `name`, `message`, `stack`.
  - `"redacts sensitive fields"` — `log.info('x', { password: 'p', token: 't', safe: 'ok' })`
    → `password: "[REDACTED]"`, `token: "[REDACTED]"`, `safe: "ok"`.
  - `"respects LOG_LEVEL env"` — debug output only with `LOG_LEVEL=debug`.

## Implementation Context

*Read before starting build.*

### Decisions that apply

- `DEC-001` — Logger choice (pino) and single-wrapper pattern. Don't
  deviate without a superseding decision.

### Constraints that apply

Full text in `/guidance/constraints.yaml`:

- `no-secrets-in-code` — use pino's `redact` option.
- `test-before-implementation` — tests above are already written.
- `no-new-top-level-deps-without-decision` — `pino` justified by `DEC-001`.
  Any other dependency needs a new DEC.

### Prior related work

None — first spec in PROJ-001.

### Out of scope

- No log shipping / transport.
- No log rotation.
- No per-module overrides.

## Notes for the Implementer

- Pino's built-in `redact` option handles redaction. Use it.
- Keep the module under 50 lines. Glue code, not a framework.

---

## Build Completion

*To be filled in at end of build.*

- **Branch:**
- **PR (if applicable):**
- **All acceptance criteria met?** <not yet built>
- **New decisions emitted:**
- **Deviations from spec:**
- **Follow-up work identified:**

### Build-phase reflection

1. **What was unclear that slowed you down?** —
2. **Constraint or decision that should have been listed but wasn't?** —
3. **If you did this task again, what would you do differently?** —

---

## Reflection (Ship)

*To be appended during ship.*

1. **What would I do differently next time?** — <not yet shipped>
2. **Does any template, constraint, or decision need updating?** — <not yet shipped>
3. **Is there a follow-up spec to write now?** — <not yet shipped>
