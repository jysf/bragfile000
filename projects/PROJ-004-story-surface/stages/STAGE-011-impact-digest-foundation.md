---
# Maps to ContextCore epic-level conventions.
# A Stage is a coherent chunk of work within a Project.
# It has a spec backlog and ships as a unit when the backlog is done.

stage:
  id: STAGE-011                     # stable, zero-padded, repo-global (never reused)
  status: proposed                  # proposed | active | shipped | cancelled | on_hold
  priority: high                    # the digest foundation the story surface reads
  target_complete: null             # optional: YYYY-MM-DD (ships v0.4.0)

project:
  id: PROJ-004                      # parent project
repo:
  id: bragfile

created_at: 2026-07-05
shipped_at: null
---

# STAGE-011: `brag impact` — the digest foundation

## What This Stage Is

When this stage ships, bragfile has the **rule-based, time-windowed,
initiative-grouped impact digest** the story surface reads: `brag impact`
(`--quarter|--month|--year|--since`) surfaces the `impact` fields grouped by
project/initiative over the DEC-014 envelope, reusing `internal/aggregate`.
This is the deterministic data foundation STAGE-012's `brag story` shapes per
audience. It ships as part of **v0.4.0**.

## Why Now

- The corpus + write spine (v0.3.0) exist and are dogfooded; the digest reads a
  real, provenance-tagged corpus.
- `brag impact` is retro action **P1** and the direct dependency of STAGE-012's
  narrative surface — build the deterministic foundation before the shaping.

## Success Criteria

- **`brag impact`** produces a rule-based impact digest — time-windowed,
  grouped by initiative/project, surfacing the `impact` fields — over the
  DEC-014 envelope, reusing `internal/aggregate`. (Its own spec, TBD.)
- No schema migration; local-first / no-network intact; ships as v0.4.0.

## Scope

### In scope
- `brag impact` — the rule-based, time-windowed, initiative-grouped digest.
  (Spec TBD.)

### Explicitly out of scope
- `brag story --audience …` and the audience taxonomy — STAGE-012.
- The cost/session/token capture seed — its own single-concern stage
  (STAGE-014, a v0.3.x patch), on a different release line.
- Any networked / multi-user MCP mode.

## Spec Backlog

Ordered list of specs composing this stage. Add specs as you identify
them. Update status as specs progress.

Format: `- [status] SPEC-ID (cycle) — one-line summary`

- [ ] (not yet written) — `brag impact` — the rule-based, time-windowed,
      initiative-grouped impact digest (DEC-014 envelope; the STAGE-012 basis).

**Count:** 0 shipped / 0 active / 1 pending

## Design Notes

- `brag impact` extends the DEC-014 rule-based output family and reuses
  `internal/aggregate` (`ByType`/`ByProject`/`Span`) — data + shaping only, no
  LLM in the binary (same pipe posture as `brag review`/`summary`).

## Dependencies

### Depends on
- **PROJ-003 (v0.3.0, shipped)** — `internal/aggregate` for the digest, and the
  provenance-tagged corpus it reads.
- **DEC-014** (rule-based output envelope) — `brag impact` extends the family.

### Enables
- **STAGE-012** — `brag story --audience` reads the impact digest.

## Stage-Level Reflection

*Filled in when status moves to shipped. Run Prompt 1c (Stage Ship) in
FIRST_SESSION_PROMPTS.md to draft this.*

- **Did we deliver the outcome in "What This Stage Is"?** <tbd>
- **How many specs did it actually take?** <tbd>
- **What changed between starting and shipping?** <tbd>
- **Lessons that should update AGENTS.md, templates, or constraints?** <tbd>
- **Should any spec-level reflections be promoted to stage-level lessons?** <tbd>
