---
insight:
  id: DEC-003
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
  id: bragfile

created_at: 2026-04-19
supersedes: null
superseded_by: null

tags:
  - config
  - cli
  - conventions
---

# DEC-003: Config resolution order — flag → env → default

## Decision

For any configurable value, `brag` resolves in this order:

1. CLI flag (highest precedence).
2. Environment variable.
3. Built-in default (lowest).

The first path currently subject to this: DB location.
- Flag: `--db <path>`
- Env: `BRAGFILE_DB`
- Default: `~/.bragfile/db.sqlite`

No config file in v0.1.

## Context

The CLI needs a predictable way to override runtime settings for
testing (every test uses `t.TempDir()` via `--db`), for alternate
deployments (someone wanting to sync via iCloud/Dropbox), and for
eventual other settings (editor command, default range for
`summary`, etc.).

## Alternatives Considered

- **Option A: Flag only**
  - Why rejected: Every shell invocation would need `--db`. Ergonomic
    tax on the 10-second-capture goal.

- **Option B: Env only**
  - Why rejected: Hard to override for a single invocation. Common
    source of surprise in test suites (env leaks between test runs).

- **Option C: Config file (`~/.bragfile/config.yaml`) + flag**
  - Why rejected: More code, more failure modes, no current user need.
    We can add a config file later without breaking the flag/env API
    (just insert it between env and default in the resolution order).

- **Option D (chosen): flag → env → default**
  - Why selected: The standard Unix convention. Zero-config default
    works out of the box. `--db` trivially overrides for tests.
    `BRAGFILE_DB` exists for users who want a permanent override in
    their shell rc file.

## Consequences

- **Positive:** Matches user intuition. Makes testing clean
  (`--db=$(mktemp ...)`). Easy to extend with a config file later.
- **Negative:** Values with no flag or env (currently none) would be
  surprising — we should add a flag+env for any new user-visible
  configurable.
- **Neutral:** Env var is namespaced `BRAGFILE_*`; adding more env-vars
  is a forward-compatible move.

## Validation

Right if:
- Tests use `--db` exclusively; env is never set in CI.
- Users never ask "why didn't my config take effect?".

Revisit if:
- The number of configurables grows past ~5. At that point a config
  file becomes worth the complexity.

## References

- Related specs: SPEC-001 (root command wiring)
- Related docs: `./docs/api-contract.md` — Global flags section
