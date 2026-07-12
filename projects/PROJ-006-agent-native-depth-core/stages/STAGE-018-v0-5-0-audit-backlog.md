---
# Maps to ContextCore epic-level conventions.
# A Stage is a coherent chunk of work within a Project.
# It has a spec backlog and ships as a unit when the backlog is done.

stage:
  id: STAGE-018
  status: proposed                  # proposed | active | shipped | cancelled | on_hold
  priority: low
  target_complete: null

project:
  id: PROJ-006
repo:
  id: bragfile

created_at: 2026-07-12
shipped_at: null
---

# STAGE-018: v0.5.0 audit backlog (LOW / NITs)

## What This Stage Is

A parking lot for the small correctness/robustness nits surfaced by the v0.5.0
pre-release audit and never actioned. Individually none of them justifies a
stage; collectively they are a coherent "harden the edges" pass that can ship as
one unit (or be slotted in opportunistically alongside a deeper pillar stage).
**Proposed, not active — created to capture the backlog as framed work, not to
schedule it.** Pull it active only when there's slack or when one of these bites.

## Why Now

Not now, deliberately. These were triaged LOW at the v0.5.0 audit; the product
is a complete personal tool without them. This stage exists so the items live in
the work hierarchy (not just a prose list in the project brief) and can be
picked up cheaply whenever convenient — or folded into whichever deeper PROJ-006
stage touches the same code.

## Success Criteria

- Each item below is either fixed (with a regression test) or explicitly closed
  as WONTFIX with a one-line rationale.
- No behavior change visible to existing users beyond the corrected edge cases.
- Full gate set green.

## Scope

### In scope
The v0.5.0 audit LOW/NIT backlog (inherited verbatim from the project brief):

- `mcp_install` atomic write (temp + rename).
- MCP `list`/`search` negative-`limit` parity with the CLI.
- `search -foo` → clear cobra-flag error (not an FTS query).
- `brag project new` name cap; run the edit / `Store.Update` path through
  `internal/capture.Validate`.
- `brag spark` same-second exclusive-edge.
- backup-filename same-second collision.
- empty-`type` sentinel handling.
- export-md sort id-tiebreak.
- `MergeTags` position dup.
- double-wrapped db-path error.
- `$EDITOR`-with-spaces handling.

### Explicitly out of scope
- The deeper agent-native pillars (memory / signed provenance / capture
  completeness / benchmark) — separate PROJ-006 stages.
- `ParseSince` wall-clock impurity — **already handled**: STAGE-017 / SPEC-068
  folded in the `since.go` clock-seam fix (audit L4). Verify before re-listing;
  do not double-count.

## Spec Backlog

Ordered list of specs composing this stage. IDs assigned at creation.

Format: `- [status] SPEC-ID (cycle) — one-line summary`

- [ ] (not yet written) — one or more specs grouping the items above by area
      (e.g. a "MCP/CLI parity" spec, a "capture validation edges" spec, a
      "filesystem-write robustness" spec). Split at framing.

**Count:** 0 shipped / 0 active / 0 pending (unframed — stage is `proposed`)

## Design Notes

- Batch by touched package, not by audit order — several of these live in the
  same file and share a test. One spec per cluster keeps review cheap.
- Prefer folding an item into a deeper-pillar stage when that stage already
  edits the same code, rather than a standalone visit.

## Dependencies

### Depends on
- Nothing new — all items are against shipped v0.5.0 code.

### Enables
- A marginally more robust substrate for the deeper PROJ-006 stages that build
  on the same MCP / capture / storage paths.

## Stage-Level Reflection

*Filled in when status moves to shipped.*

- **Did we deliver the outcome in "What This Stage Is"?** <yes/no + notes>
- **How many specs did it actually take?** <number vs. plan>
- **What changed between starting and shipping?** <one sentence>
- **Lessons that should update AGENTS.md, templates, or constraints?**
  - <one-line updates>
