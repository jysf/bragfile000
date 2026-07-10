---
# Maps to ContextCore epic-level conventions.
# A Stage is a coherent chunk of work within a Project.
# It has a spec backlog and ships as a unit when the backlog is done.

stage:
  id: STAGE-014                     # stable, zero-padded, repo-global (never reused)
  status: shipped                   # proposed | active | shipped | cancelled | on_hold
  priority: high                    # time-sensitive: economics history only accrues forward
  target_complete: null             # optional: YYYY-MM-DD

project:
  id: PROJ-004                      # parent project
repo:
  id: bragfile

created_at: 2026-07-06
shipped_at: 2026-07-06
---

# STAGE-014: Seed cost / session / token capture (v0.3.x)

## What This Stage Is

When this stage ships, the MCP `brag_add` provenance path records a reliable
**session join-key** now — plus **optional** caller-supplied cost/tokens — as
reserved-namespace tags (`session:<id>`, `cost:<n>`, `tokens:<n>`), and the
capture-nudge hook surfaces the `session_id` so Claude forwards it on
`brag_add`. It extends the SPEC-040 / DEC-024 provenance path with the same
`stampProvenance` helper; migration-free (rides the DEC-015 taggings join); it
fabricates nothing (all three inputs optional — empty → no tag, exactly like
`agent`/`model` today); and it leaves `--author` classification untouched
(`session:`/`cost:`/`tokens:` are NOT author-provenance tags).

This is a **single-concern stage** deliberately split off from the digest work:
it **ships as a v0.3.x patch**, ahead of and independent of the v0.4.0
`brag impact` digest (STAGE-011) that will eventually *read* this data. A stage
is the atomic ship-unit, so the seed — on a different release line — gets its
own stage rather than riding STAGE-011's v0.4.0 backlog.

## Why Now

Economics history only accrues going forward. When provenance landed in v0.3.0,
every pre-v0.3.0 entry was permanently un-attributable because it was stamped
late (the corpus had **0** agent-authored history in hindsight). PROJ-005's
economics layer will want per-work cost/token/session data; if we wait until
then to start capturing it, the corpus is empty in hindsight exactly the same
way. Seeding now — cheaply, on a point release — is the whole value, so it
lands as early as a v0.3.x patch allows, not gated on the digest.

## Success Criteria

- The MCP `brag_add` path records `session:<id>` (plus optional `cost:<n>` /
  `tokens:<n>`) as reserved tags, migration-free, with zero fabrication (all
  optional; empty → no tag; bad numerics rejected as a tool error).
- `--author agent|human` classification is unchanged: a `session:`/`cost:`/
  `tokens:`-only entry classifies as `human` (`provenanceExistsClause` stays
  `agent:%`/`model:%`-only).
- The capture-nudge hook surfaces `session_id` and instructs Claude to forward
  it; silent-degradation, once-per-session, and never-runs-`brag` contracts
  intact.
- Ships as a **v0.3.x patch**; no schema migration; local-first / no-network
  intact.

## Scope

### In scope
- **SPEC-046** — the "seed early" cost/session/token capture on the MCP
  `brag_add` provenance path (reserved tags `session:`/`cost:`/`tokens:`,
  DEC-027) + the capture-nudge hook `session_id` surfacing. A v0.3.x patch.

### Explicitly out of scope
- `brag impact` and the audience story surface — STAGE-011 / STAGE-012
  (v0.4.0).
- First-class cost/tokens/session **columns** and exact-token reconciliation
  (join `session:` → usage logs) — PROJ-005 (DEC-027 accepts the stringly-typed
  tag as debt, per the DEC-004→DEC-015 precedent).
- CLI `--session`/`--cost`/`--tokens` flags — MCP-path-only (DEC-027 Option D
  rejected).
- Any networked / multi-user MCP mode; bragfile estimating tokens or cost.

## Spec Backlog

Ordered list of specs composing this stage. Add specs as you identify
them. Update status as specs progress.

Format: `- [status] SPEC-ID (cycle) — one-line summary`

- [x] SPEC-046 (shipped on 2026-07-06) — seed cost/session/token capture on the
      MCP `brag_add` provenance path (reserved tags `session:`/`cost:`/`tokens:`,
      DEC-027); a v0.3.x patch.
- [x] SPEC-047 (shipped on 2026-07-06) — **v0.3.1 release cut** (the stage's
      closing action): authored CHANGELOG `[0.3.1]`, bumped the plugin version
      pin, ticked the operational pre-flight. The mechanical prep is merged to
      `main`; the irreversible tag + publish + `brew upgrade` verify is driven by
      the coordinator from `main`. Closes STAGE-014.

**Count:** 2 shipped / 0 active / 0 pending

## Design Notes

- SPEC-046 is a **seed**, not the economics feature: it captures a reliable
  session JOIN-KEY now (bragfile cannot self-count tokens; the stdio MCP
  transport exposes no session id — only `clientInfo.Name`), plus *optional*
  real cost/tokens only when a caller provides them. The tag→column promotion
  is accepted debt (DEC-027 extends DEC-024's reserved namespace).
- **Author classification is unaffected:** `session:`/`cost:`/`tokens:` are
  reserved but are **not** author-provenance tags; `store.go`'s
  `provenanceExistsClause` stays `agent:%`/`model:%`-only (see DEC-027).
- **Stage-number vs. ship-order:** STAGE-014 is a higher number than the
  v0.4.0 stages it ships *before* — honest, because it was split out after the
  original STAGE-011/012/013 were framed, and §2 stage IDs track creation
  order, not ship order (cf. STAGE-010, created-but-never-activated).

## Dependencies

### Depends on
- **PROJ-003 (v0.3.0, shipped)** — the MCP `brag_add` provenance path
  (`internal/mcpserver/provenance.go`, `server.go`), DEC-024's reserved
  namespace, and the capture-nudge hook (DEC-025).

### Enables
- **PROJ-005+** — the cost/session history this stage seeds is the substrate
  for the economics / exec-ROI story and exact-token reconciliation.

## Stage-Level Reflection

*Filled in when status moves to shipped (v0.3.1, 2026-07-06).*

- **Did we deliver the outcome in "What This Stage Is"?** Yes. v0.3.1 ships the
  seed: the MCP `brag_add` path accepts optional `session`/`cost`/`tokens`,
  stamped as reserved `session:`/`cost:`/`tokens:` tags (DEC-027), and the
  capture-nudge hook forwards the Claude Code `session_id`. Cost/session history
  now accrues going forward — the seed-early goal that motivated ordering this
  ahead of the v0.4.0 story work.
- **How many specs did it actually take?** Two: SPEC-046 (the capture feature)
  and SPEC-047 (the v0.3.1 release cut, the stage's closing action).
- **What changed between starting and shipping?** The stage itself was created
  mid-flight — SPEC-046 was first framed under STAGE-011 (the v0.4.0 `brag
  impact` stage); orchestrator review split it into this dedicated STAGE-014
  because a stage is the atomic ship-unit and cannot straddle two release lines
  (v0.3.x vs v0.4.0). The technical design held unchanged: reserved-tags-now,
  MCP-only, migration-free.
- **Lessons that should update AGENTS.md, templates, or constraints?** One
  candidate, held below the codification bar: "a stage is the atomic ship-unit
  and cannot straddle release lines; project *lineage* may span releases, a
  stage may not." Under the §12 meta-rule this is N=1 — a watch item, not yet
  codified. The Gatekeeper `xattr` quarantine friction recurred at this cut
  exactly as §4 documents (already captured; no change needed).
- **Should any spec-level reflections be promoted to stage-level lessons?**
  SPEC-046's placement lesson (state the ship-unit invariant in the design
  handoff charter up front) IS the stage-level lesson above; no further
  promotion needed.
