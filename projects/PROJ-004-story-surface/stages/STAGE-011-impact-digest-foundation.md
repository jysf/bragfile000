---
# Maps to ContextCore epic-level conventions.
# A Stage is a coherent chunk of work within a Project.
# It has a spec backlog and ships as a unit when the backlog is done.

stage:
  id: STAGE-011
  status: shipped
  priority: high                    # the digest foundation the story surface reads
  target_complete: null             # optional: YYYY-MM-DD (ships v0.4.0)

project:
  id: PROJ-004                      # parent project
repo:
  id: bragfile

created_at: 2026-07-05
shipped_at: 2026-07-06             # backlog complete + merged to main; v0.4.0 RELEASE deferred to the STAGE-013 cut
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

- [x] SPEC-048 (shipped on 2026-07-06) — `brag impact` — the rule-based,
      calendar-windowed, initiative-grouped, impact-first digest (DEC-014
      envelope + DEC-028 window/shape; the STAGE-012 basis). On main; reaches
      users in the v0.4.0 release.

**Count:** 1 shipped / 0 active / 0 pending

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

*Filled at ship (2026-07-06). Backlog complete + merged to main; the v0.4.0
release that puts it in users' hands is deferred to the STAGE-013 cut.*

- **Did we deliver the outcome in "What This Stage Is"?** Yes. `brag impact`
  ships: a rule-based, calendar-windowed (`--quarter|--month|--year|--since`),
  initiative-grouped, **impact-first** digest over the DEC-014 envelope,
  reusing `internal/aggregate`. It's the deterministic foundation STAGE-012's
  `brag story` will shape per audience.
- **How many specs did it actually take?** One — SPEC-048 (+ DEC-028).
- **What changed between starting and shipping?** DEC-028 chose **calendar**
  windows (current period to date), a deliberate, tested divergence from
  DEC-014's rolling windows — validated by the user (companies report by
  calendar quarter). Sparklines were considered and **deferred** to a dedicated
  visual spec (impact ships text-pure). `--previous` (last-completed period)
  was left as clean future scope.
- **Lessons that should update AGENTS.md, templates, or constraints?** Window
  semantics are now **per-command** (DEC-028 calendar vs DEC-014 rolling), not
  a single global rule — worth remembering when `brag story` picks its reading
  window. N=1, below the §12 codification bar; no template change.
- **Should any spec-level reflections be promoted to stage-level lessons?**
  SPEC-048's "surface visual/UX scope forks at design altitude, decide them
  explicitly" — carry it into STAGE-012/013, which hold more shaping/UX
  decisions.
