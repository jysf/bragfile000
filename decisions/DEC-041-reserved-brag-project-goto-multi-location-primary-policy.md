---
insight:
  id: DEC-041
  type: reservation
  audience:
    - developer
    - agent

agent:
  id: claude-opus-5
  session_id: null

project:
  id: PROJ-007
repo:
  id: bragfile

created_at: 2026-08-19
supersedes: null
superseded_by: null

tags:
  - reservation
  - backlog
---

# DEC-041: reserved — `brag project goto` / multi-location primary policy

## This is not a decision

`DEC-041` is reserved, not written. It exists only so `decisions/` has no
unexplained gap between [`DEC-040`](DEC-040-distribution-binary-formula-over-cask.md)
and [`DEC-042`](DEC-042-mcp-time-window-filter-parity.md) — there is no choice
recorded here to read, weigh, or cite.

The number was set aside on 2026-07-16 for `brag project goto` (SPEC-070), a
navigation-ergonomics feature that was drafted, then deferred — not built, not
rejected — on 2026-08-13.

## What DEC-041 would decide, if `brag project goto` is ever built

A project can have many locations (`project_locations`, one-to-many —
[`DEC-017`](DEC-017-entries-project-relationship.md),
[`DEC-019`](DEC-019-project-here-resolution-policy.md),
[`DEC-020`](DEC-020-project-location-editing-semantics.md)), so a `goto`
command has to pick one as "the" directory to jump to. The deferred draft
assumed **first-registered = primary** — a real choice, not an obvious one,
which is why it needed a decision record at all rather than a one-line
default.

## Where the rest of the design lives

The full navigation design — why a subprocess can't `cd` its parent shell,
the two-part `goto` (resolver) + `shell-init` (shell-function wrapper) shape,
and the trigger for revisiting it — is preserved in
[`../projects/PROJ-001-mvp/backlog.md`](../projects/PROJ-001-mvp/backlog.md),
under "`brag project goto` + `brag shell-init` — jump to a project's
directory."

## Disposition

- **If `brag project goto` is ever built**, it takes this number: the
  decision gets written into *this* file, replacing this notice, rather than
  claiming a new one.
- **If the backlog item is ever deleted outright** instead, this file is
  updated to say so — so the gap stays explained rather than becoming a
  second mystery.
- **Do not reuse `DEC-041` for anything else.** The backlog entry says so
  explicitly, and reusing a reserved number would leave a decision record and
  its own reservation notice referring to two different things.
