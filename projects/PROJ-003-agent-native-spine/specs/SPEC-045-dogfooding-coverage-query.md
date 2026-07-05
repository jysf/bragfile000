---
# Maps to ContextCore task.* semantic conventions.
# DRAFT SCAFFOLD (retro P3) — frame stub, not yet through design.

task:
  id: SPEC-045
  type: story
  cycle: frame                     # DRAFT — proposed from cross-project-retro P3; needs a design session
  blocked: false                   # P2 (SPEC-043 --author) shipped, so the classifier exists
  priority: medium                 # retro P2-priority; the read/measure half of the agent-native thesis
  complexity: M

project:
  id: PROJ-003
  stage: STAGE-010                 # the impact/read-surface stage (stretch; not yet activated)
repo:
  id: bragfile

agents:
  architect: claude-opus-4-8
  implementer: claude-opus-4-8
  created_at: 2026-07-05

references:
  decisions: [DEC-024, DEC-014]
  constraints: [no-sql-in-cli-layer, stdout-is-for-data-stderr-is-for-humans]
  related_specs: [SPEC-043, SPEC-020]
---

# SPEC-045: dogfooding-coverage query

> **DRAFT SCAFFOLD (retro P3).** Proposed from
> [`2026-07-04-action-register.md`](../../../docs/reports/cross-project/2026-07-04-action-register.md)
> item P3. Not yet designed — this stub captures the scope. **Depends on
> SPEC-043** (the `agent:`/`model:` classifier), which shipped in v0.3.0.

## Context

v0.3.0 made the corpus *able* to distinguish agent- from human-authored
entries (`brag list --author`, SPEC-043). The agent-native thesis is now
**measurable but not yet measured over time**. At the v0.3.0 cut the baseline
is **0% agent-authored** (189 human / 0 agent) — captured here so the
post-v0.3.0 trend is visible as the MCP write path gets used.

## Goal

A query that reports **provenance share** (agent-authored vs human-authored)
and **self-reference density**, windowed by month, so v0.3.0's agent-native
adoption shows up as a trend rather than a single number.

## Proposed approach (to lock at design)

- **Classifier reuse.** Reuse SPEC-043's reserved-tag predicate
  (`agent:`/`model:` present) — factor it into `internal/aggregate` as a pure
  `IsAgentAuthored(Entry) bool` if not already, so storage's SQL classifier
  and this Go-level bucketer share one definition (closes the SPEC-043
  drift-coupling WATCH).
- **Aggregation.** `internal/aggregate` groups `[]Entry` into monthly buckets
  → `{month, agent, human, share}` + a self-reference-density measure
  (entries mentioning `brag`/`bragfile`). SQL stays in storage; aggregate is
  SQL-free (as today).
- **Surface.** A read command/flag with JSON + human output per
  `stdout-is-for-data` — e.g. `brag stats --provenance` or a dedicated
  `brag coverage`. Decide at design; keep the DEC-014 envelope shape.

## Acceptance Criteria (sketch)

- [ ] Returns provenance share (agent vs human) and self-reference density,
      windowed by month.
- [ ] Baseline at report date is captured/derivable (agent ≈ 0% at v0.3.0).
- [ ] JSON + human output; golden tests lock the shapes (literal-artifact).
- [ ] The classifier is single-sourced with SPEC-043 (no second definition
      of "what counts as agent-authored").

## Notes for the Implementer

- This is the natural STAGE-010 opener alongside the impact digest (retro P1)
  — activating STAGE-010 is a prerequisite (the stage file does not exist yet).
- Do **not** re-implement the `agent:`/`model:` membership test with a third
  literal; converge storage's SQL `LIKE` and this bucketer on one predicate.
