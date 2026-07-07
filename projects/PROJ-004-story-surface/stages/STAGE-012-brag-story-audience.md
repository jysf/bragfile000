---
# Maps to ContextCore epic-level conventions.
# A Stage is a coherent chunk of work within a Project.
# It has a spec backlog and ships as a unit when the backlog is done.

stage:
  id: STAGE-012
  status: active
  priority: high                    # the narrative headline of v0.4.0
  target_complete: null             # optional: YYYY-MM-DD (ships v0.4.0)

project:
  id: PROJ-004                      # parent project
repo:
  id: bragfile

created_at: 2026-07-06
shipped_at: null
---

# STAGE-012: `brag story --audience` — the narrative surface

## What This Stage Is

The **headline** of v0.4.0. `brag story --audience <me|manager|exec>` reads the
corpus and emits an **audience-shaped bundle** that *coalesces a set of brags
into narrative arcs* — a throughline with a beginning, middle, and "so what" —
not merely a filtered list. It ships as **v0.4.0**.

The essence, and what separates a **story** from `brag impact`'s **report**:
related brags become **beats in one arc**, not bullets in a list.

**Synthesis is a PURE PIPE (settled, DEC to come):** bragfile owns *data +
shaping*, the LLM owns *words*. bragfile emits a shaped bundle + a
framing-directive asset; the LLM that is **already in the caller** (an agent
like Claude Code, or a paste-in session) writes the prose. **No model, no
network, no secrets in the binary** — same posture as `brag review`/`summary`,
preserving DEC-001. The LLM is **optional**: the shaped bundle is useful
standalone (readable / pasteable / scriptable); the LLM is the upgrade to
polished narrative.

## Why Now

- **The deterministic foundation exists.** `brag impact` (STAGE-011) shipped the
  windowed, initiative-grouped, impact-first digest; `brag story` builds on its
  grouped data, adding narrative threading + the framing directive.
- **It's the original north-star, finally** (AGENTS.md §1: retros, reviews,
  resumes). The write spine (v0.3.0) + the read digest (STAGE-011) are the
  scaffolding; this is the payoff.

## Success Criteria

- **`brag story --audience <me|manager|exec>`** emits an audience-shaped bundle
  that **coalesces the corpus into arcs** — demonstrably different *selection,
  threading, and altitude* per audience (exec = one headline arc, impact-forward;
  `me` = every thread, candid, including the messy middle + lessons).
- A **framing-directive asset** ships alongside the bundle (like BRAG.md / the
  plugin assets) — the LLM's instructions for weaving the arc.
- **LLM is optional:** the bundle is a complete, useful artifact on its own; no
  model in the binary; local-first / no-network intact.
- Same corpus → different stories, and the difference is **rule-driven**
  (selection + threading + altitude), not just a tone hint.

## Scope

### In scope
- `brag story --audience …` — the command, the audience-shaped bundle, the
  threading/coalescing, and the framing-directive assets.
- The **audience taxonomy + shaping rules** — an **extensible shaping-profile**
  mechanism (each audience = data/config: selection filter + threading/grouping +
  framing directive + altitude/length), shipping defaults (me / manager / exec),
  **NOT a hard-coded enum**. Its own DEC.
- The **bundle format** an LLM consumes + the framing-directive assets.

### Explicitly out of scope (→ later)
- Baking an LLM into the binary / any CLI→LLM call-out (settled: pipe only).
- Team / multi-user federation (PROJ-005).
- The sparklines/visual pass and `brag wrapped` (STAGE-013).
- Cross-person taxonomy reconciliation, sharing/transport.

## Spec Backlog

Ordered list of specs composing this stage. Add specs as identified.

Format: `- [status] SPEC-ID (cycle) — one-line summary`

- [ ] SPEC-049 (design) — `brag story --audience` — the extensible
      shaping-profile mechanism, the coalesce-into-arcs bundle, and the
      framing-directive assets (emits DEC-029, the audience taxonomy + the
      thread-definition choice). The headline spec. **Assessed L → split
      taken:** SPEC-049 = the mechanism + the gradient ENDPOINTS
      (`me`/`exec`); SPEC-050 = `manager`/`skip` as config-only additions
      (the extensibility proof). Split needs orchestrator sign-off.
- [ ] SPEC-050 (pending, planned in SPEC-049's split) — `manager` (and
      optionally `skip`) audience profiles + directive assets, shipped as
      bundled defaults with ZERO Go change — the proof that DEC-029's
      profiles-as-data mechanism is extensible. Plus any polish
      (per-profile fold thresholds, doc/tutorial pass).

**Count:** 0 shipped / 0 active / 2 pending

## Design Notes

- **Posture (settled): pure pipe, LLM optional.** bragfile assembles the set +
  deterministic threads + time-ordering + impact beats + a throughline
  *skeleton*; the framing directive + the caller's LLM *weave* the arc. Rejected
  a CLI→LLM call-out (a one-way door onto keys/network/cost/flaky tests).
- **Audience sets *how many arcs* and *at what altitude*** — not just the filter
  and tone. Exec = one headline arc; `me` = every thread with the messy middle.
- **Audiences are extensible profiles, not a locked enum** ("same in theory,
  diverge in practice") — data/config: selection + threading + framing directive
  + altitude. Ship me/manager/exec defaults.
- **Reuses `brag impact` (STAGE-011) grouped data** as the raw material.
- **RESOLVED (DEC-029, SPEC-049 design) — what defines a "thread"/arc**:
  bragfile assembles **deterministic threads** = initiative (the `project`
  axis, reusing `GroupEntriesByProject`), time-ordered, with **impact beats**
  marked (`WithImpact`), plus an **opt-in `--theme` cross-project cross-cut**,
  and a **throughline SKELETON** (ordered thread refs + span + beat/impact-beat
  counts). The **framing directive + the caller's LLM find the throughline**
  and weave the arc. Rejected: LLM-inferred threading in the binary (bakes in a
  model / breaks standalone use); theme or time-progression as the *primary*
  axis (folded in as a cross-cut + within-thread ordering instead). Audiences
  are **data-driven profiles** (bundled `embed.FS` defaults + user override),
  not a Go enum. Confidence 0.72 (two sub-choices — profile-file format and the
  two-audience slice — filed as questions).

## Dependencies

### Depends on
- **STAGE-011 (`brag impact`, shipped to main)** — the grouped/impact data the
  story coalesces.
- **DEC-014** (rule-based output envelope) + the corpus + provenance (v0.3.0).
- **A framing-directive asset convention** (like BRAG.md / the plugin assets).

### Enables
- **STAGE-013** — polish + the v0.4.0 cut (`brag wrapped`, the visual pass).
- **PROJ-005+** — the single-user shaping logic (select → thread → frame → LLM)
  is reused at team/org scope over the spike-validated warehouse.

## Stage-Level Reflection

*Filled in when status moves to shipped.*

- **Did we deliver the outcome in "What This Stage Is"?** <tbd>
- **How many specs did it actually take?** <tbd>
- **What changed between starting and shipping?** <tbd>
- **Lessons that should update AGENTS.md, templates, or constraints?** <tbd>
- **Should any spec-level reflections be promoted to stage-level lessons?** <tbd>
