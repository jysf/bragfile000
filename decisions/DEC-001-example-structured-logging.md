---
insight:
  id: DEC-001
  type: decision
  confidence: 0.95
  audience:
    - developer
    - agent

agent:
  id: claude-opus-4-7
  session_id: null

project:
  id: PROJ-001
repo:
  id: my-app

created_at: 2026-04-19
supersedes: null
superseded_by: null

tags:
  - example
  - infrastructure
  - logging
---

# DEC-001: Use structured logging via a single project logger module

> **Note:** This is an example decision. Delete this file once you've
> created your own.

## Decision

All application logging goes through `src/lib/log.ts`, which wraps a
single logging library (pino) and emits JSON-structured logs.

## Context

We need consistent logging across the codebase for three reasons:
1. Structured logs are cheaper to query and filter in production.
2. A single logger makes it trivial to later ship logs to Loki or any
   OTLP-compatible backend.
3. Multiple logging libraries in one repo is a known source of bugs
   and duplicate log lines.

This decision was made during initial architecture of PROJ-001, before
any code was written.

## Alternatives Considered

- **Option A: `console.log` everywhere**
  - What it is: Just use the built-in console.
  - Why rejected: No structure, no levels, no redaction. Impossible
    to filter in noisy environments.

- **Option B: Let each module pick its own logger**
  - What it is: No standard — teams/contributors choose.
  - Why rejected: Guarantees inconsistency. Every contributor picks
    differently; in six months we have 3 logging libraries.

- **Option C (chosen): One wrapper module, one underlying library**
  - What it is: `src/lib/log.ts` exports `log.info`, `log.warn`,
    `log.error`, `log.debug`. Wraps pino. All application code
    imports from here.
  - Why selected: One import path, one place to configure, one place
    to swap the underlying library if needed. JSON output makes future
    shipping to Loki or Datadog a configuration change, not a refactor.

## Consequences

- **Positive:** Single point of control for log format, levels,
  destinations. Easy future migration to any aggregator. Enforceable
  via lint rule and the `use-project-logger` constraint.
- **Negative:** One more abstraction layer. Contributors unfamiliar
  with the convention may reach for `console.log` out of habit
  (mitigated by lint rule and the constraint).
- **Neutral:** Pino-specific features are accessible but only through
  the wrapper.

## Validation

This decision is right if:
- We never end up with a second logging library in the repo.
- Log queries in production are tractable (structured fields work).

Revisit if:
- Pino stops being maintained or develops a major security issue.
- We need per-module log configuration the wrapper can't express.

## References

- Related specs: SPEC-001
- Related decisions: (first one)
- External docs: https://getpino.io
- Related constraint: `use-project-logger` in `/guidance/constraints.yaml`
