---
insight:
  id: DEC-007
  type: decision
  confidence: 0.80
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

created_at: 2026-04-20
supersedes: null
superseded_by: null

tags:
  - cli
  - error-handling
  - conventions
---

# DEC-007: Required-flag validation lives in `RunE`, not `MarkFlagRequired`

## Decision

Commands in `internal/cli/` do NOT use `cobra.Command.MarkFlagRequired`
to enforce required flags. Instead, each `RunE` validates its own
required flags up front and returns a `cli.UserErrorf(...)` on
violation. For string flags, the canonical check is
`strings.TrimSpace(v) == ""`.

## Context

SPEC-003 established the `cli.ErrUser` sentinel and a contract that
`main.go` maps `errors.Is(err, cli.ErrUser)` to exit code 1 and
anything else to exit code 2. Acceptance test
`TestAdd_MissingTitleIsUserError` requires `errors.Is(err, ErrUser)`
to be true when `--title` is not set.

Cobra's `MarkFlagRequired` causes `validateRequiredFlags` to fail
before `RunE` runs, returning a plain
`errors.New("required flag(s) \"title\" not set")`. There is no
supported hook to wrap that error, so `errors.Is(err, ErrUser)` returns
false and the test fails — and more importantly `main.go` would exit 2
("internal error") for a plain user mistake.

The `TrimSpace` check in `RunE` already handles all three failure
shapes the spec enumerates (missing flag, empty string, whitespace-
only), because the flag's default value is `""`. That makes
`MarkFlagRequired` both redundant and actively harmful in this layout.

## Alternatives Considered

- **Option A: Keep `MarkFlagRequired`, accept the exit-2 mismatch.**
  - Why rejected: violates the api-contract.md exit-code table (missing
    required flag is a user error, exit 1). Also inconsistent with the
    empty/whitespace branches which correctly exit 1.

- **Option B: Keep `MarkFlagRequired`, wrap via a custom
  `FlagErrorFunc` or post-process error in `main.go`.**
  - Why rejected: `FlagErrorFunc` only fires on argv parse errors, not
    required-flag validation. Post-processing by string-matching
    cobra's error text in `main.go` would couple us to cobra internals
    and break silently on upgrade.

- **Option C (chosen): Drop `MarkFlagRequired`; validate in `RunE`.**
  - Why selected: one code path, one error shape, one exit code. The
    validation is three lines. Keeps `main.go` dumb. Extends cleanly
    to commands with multiple required flags (one `if` per flag, each
    returning `UserErrorf`).

## Consequences

- **Positive:** Missing, empty, and whitespace-only values all produce
  the same `ErrUser`-wrapped error and exit code. `brag` error output
  stays consistent (`brag: <message>` on stderr, exit 1). Easy to
  extend across future subcommands.
- **Negative:** Cobra's auto-generated help text no longer shows
  `(required)` next to `--title`. We compensate in the flag
  description (`"short headline (required)"`). If we ever want the
  asterisk in help, we can add both `MarkFlagRequired` and the
  `RunE` check — cobra's validation runs first and would mask our
  wrapped error, but we could invert that by intercepting early. Not
  worth doing in PROJ-001.
- **Neutral:** This is a project-wide convention for PROJ-001. Later
  subcommands (`list`, `show`, `edit`, `delete`, `search`, `export`,
  `summary`) inherit it.

## Validation

Right if:
- Every "user entered bad input" path in the CLI exits 1, not 2.
- `errors.Is(err, cli.ErrUser)` is the only mechanism `main.go` needs
  to distinguish user from internal errors.

Revisit if:
- Cobra gains a first-class hook for wrapping required-flag errors.
- We introduce a different error-classification scheme (multiple
  sentinels, typed errors) that has a better way to propagate exit
  codes.

## References

- Related specs: SPEC-003 (introduces `ErrUser`, exit-code mapping)
- Related constraints: `stdout-is-for-data-stderr-is-for-humans`
- Related docs: `./docs/api-contract.md` (exit-code table)
