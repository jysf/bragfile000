---
# Maps to ContextCore epic-level conventions.
# A Stage is a coherent chunk of work within a Project.
# It has a spec backlog and ships as a unit when the backlog is done.

stage:
  id: STAGE-028
  status: proposed
  priority: medium
  target_complete: null

project:
  id: PROJ-008
repo:
  id: bragfile

created_at: 2026-11-05
shipped_at: null
---

# STAGE-028: App review hardening

## What This Stage Is

The findings of the 2026-10-03 app code review
([`docs/reports/code-review-2026-10-03.md`](../../../docs/reports/code-review-2026-10-03.md))
that are worth acting on, triaged into specs. The review found no critical or
high-severity defects: this stage closes the one confirmed data-correctness bug
(M1), tightens file permissions and error handling in storage, bounds the agent
write surface, and puts the gates the review ran by hand (`-race`,
`govulncheck`, pinned toolchain) into CI.

It is a hardening stage, not a feature stage. It adds no user-facing capability.

## Why Now

- Triage on 2026-11-05 re-checked M1, M2, L17 and L18 against `main` at
  `233e0b4`: all four are still open.
- The CI and permission items are cheap, and the CI items protect the v0.7.0
  release cut.
- M5 and L13 sit on the MCP write/read surface that STAGE-027 (`brag_edit`) is
  about to extend. Bounding it first means the new tool inherits the bounds.
- M6 (request `ctx`) touches the same MCP handlers as STAGE-027 and should ride
  with it, not here.

## Success Criteria

- A pre-existing world-readable DB file is tightened to 0600 on open, and the
  migration backup sidecar is created 0600 (M2, L3), each pinned by a test.
- A project literally named `(no project)` no longer merges with unassigned
  entries in any aggregate, spark or story surface (M1).
- MCP provenance strings (`agent`/`model`/`session`) and unbounded
  `brag_list`/`brag_search` calls are capped (M5, L13).
- CI runs `go test -race` and `govulncheck`, and the Go and GoReleaser versions
  are pinned (L17, L18).
- `just test`, `just test-docs` and `just lint` stay green throughout.

## Scope

### In scope

Triage keys are the review's finding IDs.

| Spec (to frame) | Findings | Notes |
|---|---|---|
| **A. storage and ingress hardening** | M2, L1, L3, L7, L10, L11, L14, L15 | One PR, no design needed. `storage.Open` chmods an existing DB and surfaces the pre-create error; backup sidecar 0600; `export --out` returns the close error; `LoadProfile` basename check; `capture-nudge.sh` allowlists `[A-Za-z0-9_-]` for `session_id`; `--since Nd` range check; `memory.Slice` enforces its documented budget. |
| **B. `(no project)` sentinel collision** | M1 | Needs a design call: key by flag or struct and render the sentinel only at output, or reject the name at ingress. Check against DEC-017 (`entries.project` is free text) before choosing. |
| **C. bound the MCP surface** | M5, L13 | Length-cap provenance strings; default and maximum `limit` on `brag_list` and `brag_search`. Likely a DEC for the numbers. |
| **D. CI gates** | L17, L18 | Add `-race` and `govulncheck` steps; replace `check-latest: true` and `~> v2` with reviewed pins. Pairs naturally with the dependency-bump flow. |
| **E. test debt** | M3, M4, L19, L20 | Direct tests for the spark renderers; replace sleep-based storage tests with the existing `Backdate` helper; cover `runMCPServe`; fuzz the three boundary parsers. Filler work. |

### Explicitly out of scope

- **M6** (thread request `ctx` through storage): moves to STAGE-027, which
  edits the same MCP handlers. A cancelled `brag_add` committing is a
  duplicate-risk worth designing once, with `brag_edit` in view.
- **L12** (label stored text as untrusted in LLM-facing `story` and `memory`
  output): a design question about output framing, not a hardening fix. Park
  as a roadmap candidate, likely under STAGE-026.
- L2, L4, L5, L6, L8, L9, L16: self-inflicted, currently unreachable, or races
  already guarded by UNIQUE constraints. Backlog; revisit only if one fires in
  practice.
- L21, N1 to N3: refactors and nits, to do when convenient.
- Release signing or provenance attestation (an observation in the review, not
  a defect).

## Spec Backlog

- [ ] (not yet written) — A. storage and ingress hardening (M2, L1, L3, L7, L10, L11, L14, L15)
- [ ] (not yet written) — B. `(no project)` sentinel collision (M1)
- [ ] (not yet written) — C. bound the MCP surface (M5, L13)
- [ ] (not yet written) — D. CI gates: `-race`, `govulncheck`, pinned toolchain (L17, L18)
- [ ] (not yet written) — E. test debt (M3, M4, L19, L20)

**Count:** 0 shipped / 0 active / 5 pending

Suggested order: D first (cheap, protects the release), then A, then B and C,
with E as filler. Each spec is framed in its own session (AGENTS.md §6); none
has been scaffolded yet, so IDs are not reserved.

## Design Notes

- The review's severities are post-verification, but most findings are
  "code-read" confidence. Each spec's frame cycle should reproduce its finding
  with a failing test before changing code. M2 and M1 are the two where the
  reproduction is easy and the review already re-verified the mechanics.
- Mind the traps recorded in project memory: `next_id` is filename-only, so use
  `just new-spec` rather than picking an ID in prose.

## Dependencies

### Depends on
- Nothing blocking. Sequence D before the v0.7.0 release cut if possible.

### Enables
- STAGE-027 inherits bounded MCP inputs (spec C) and the CI gates (spec D).

## Stage-Level Reflection

*Filled in when status moves to shipped.*
