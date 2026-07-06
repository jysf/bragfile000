---
# Maps to ContextCore epic-level conventions.
# A Stage is a coherent chunk of work within a Project.
# It has a spec backlog and ships as a unit when the backlog is done.

stage:
  id: STAGE-011                     # stable, zero-padded, repo-global (never reused)
  status: proposed                  # proposed | active | shipped | cancelled | on_hold
  priority: high                    # the digest foundation the story surface reads
  target_complete: null             # optional: YYYY-MM-DD

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
audience.

The stage also carries a small, decoupled **capture-side seed** (SPEC-046):
extend the MCP `brag_add` provenance path to record a reliable **session
join-key** now — plus optional caller-supplied cost/tokens — as reserved tags,
so cost/session history starts accruing *before* the economics layer
(PROJ-005) needs it. This is the same "stamp early or the corpus is empty in
hindsight" lesson provenance just taught; it ships as a **v0.3.x patch**,
ahead of and independent of the v0.4.0 digest work.

## Why Now

- The corpus + write spine (v0.3.0) exist and are dogfooded; the digest reads a
  real, provenance-tagged corpus.
- `brag impact` is retro action **P1** and the direct dependency of STAGE-012's
  narrative surface — build the deterministic foundation before the shaping.
- The cost/session seed (SPEC-046) is time-sensitive in the same way provenance
  was: history only accrues going forward, so the seed lands as early as a
  point release allows, not gated on the digest.

## Success Criteria

- **`brag impact`** produces a rule-based impact digest — time-windowed,
  grouped by initiative/project, surfacing the `impact` fields — over the
  DEC-014 envelope, reusing `internal/aggregate`. (Its own spec, TBD.)
- **SPEC-046 seed:** the MCP `brag_add` path records a `session:<id>` join-key
  (plus optional `cost:<n>` / `tokens:<n>`) as reserved tags, migration-free,
  with zero fabrication (all three inputs optional; empty → no tag). The
  capture-nudge hook surfaces `session_id` and instructs Claude to forward it.
- No schema migration; local-first / no-network intact; ships within the
  v0.3.x / v0.4.0 window.

## Scope

### In scope
- `brag impact` — the rule-based, time-windowed, initiative-grouped digest.
  (Spec TBD.)
- **SPEC-046** — the "seed early" cost/session/token capture on the MCP
  `brag_add` provenance path + the hook `session_id` surfacing. Ships as a
  v0.3.x patch.

### Explicitly out of scope
- `brag story --audience …` and the audience taxonomy — STAGE-012.
- First-class cost/tokens/session **columns** and exact-token reconciliation
  (join `session:` → usage logs) — PROJ-005 (DEC-027 accepts the stringly-typed
  tag as debt, per the DEC-004→DEC-015 precedent).
- Any networked / multi-user MCP mode.

## Spec Backlog

Ordered list of specs composing this stage. Add specs as you identify
them. Update status as specs progress.

Format: `- [status] SPEC-ID (cycle) — one-line summary`

- [ ] SPEC-046 (design) — seed cost/session/token capture on the MCP `brag_add`
      provenance path (reserved tags `session:`/`cost:`/`tokens:`, DEC-027); a
      v0.3.x patch.
- [ ] (not yet written) — `brag impact` — the rule-based, time-windowed,
      initiative-grouped impact digest (DEC-014 envelope; the STAGE-012 basis).

**Count:** 0 shipped / 0 active / 2 pending

## Design Notes

- SPEC-046 is a **seed**, not the economics feature: it captures a reliable
  session JOIN-KEY now (bragfile cannot self-count tokens; the stdio MCP
  transport exposes no session id — only `clientInfo.Name`), plus *optional*
  real cost/tokens only when a caller provides them. The tag→column promotion
  is accepted debt (DEC-027 extends DEC-024's reserved namespace).
- **Author classification is unaffected:** `session:`/`cost:`/`tokens:` are
  reserved but are **not** author-provenance tags; `store.go`'s
  `provenanceExistsClause` stays `agent:%`/`model:%`-only (see DEC-027).

## Dependencies

### Depends on
- **PROJ-003 (v0.3.0, shipped)** — the MCP `brag_add` provenance path
  (`internal/mcpserver/provenance.go`, `server.go`), DEC-024's reserved
  namespace, the capture-nudge hook, and `internal/aggregate` for the digest.
- **DEC-014** (rule-based output envelope) — `brag impact` extends the family.

### Enables
- **STAGE-012** — `brag story --audience` reads the impact digest.
- **PROJ-005+** — the cost/session history SPEC-046 seeds is the substrate for
  the economics / exec-ROI story and exact-token reconciliation.

## Stage-Level Reflection

*Filled in when status moves to shipped. Run Prompt 1c (Stage Ship) in
FIRST_SESSION_PROMPTS.md to draft this.*

- **Did we deliver the outcome in "What This Stage Is"?** <tbd>
- **How many specs did it actually take?** <tbd>
- **What changed between starting and shipping?** <tbd>
- **Lessons that should update AGENTS.md, templates, or constraints?** <tbd>
- **Should any spec-level reflections be promoted to stage-level lessons?** <tbd>
