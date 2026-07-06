---
# Maps to ContextCore project.* semantic conventions.
# A project is a bounded wave of work against the repo (the app).

project:
  id: PROJ-004
  status: proposed                  # proposed | active | shipped | cancelled
  priority: high
  target_ship: null                 # v0.4.0 (single-user story surface)

repo:
  id: bragfile

created_at: 2026-07-06
shipped_at: null
---

# PROJ-004: The story surface — tell your own story (v0.4.0)

> **Framing draft (proposed).** The first, single-user layer of the
> "read & share" north-star ([`docs/roadmap/proj-004-read-and-share-scoping.md`](../../docs/roadmap/proj-004-read-and-share-scoping.md)).
> Team federation, the token/economics dimension, and org-weave are
> **explicitly a later project (PROJ-005+)** — this wave is single-user only.

## What This Project Is

Turn *your own* brag corpus into **audience-shaped stories.** v0.3.0 built the
agent-native *write* spine (agents capture what you did, with provenance);
PROJ-004 reads it back as **narrative** — shaped by *who the story is for*.

The centerpiece is a single insight: **purpose determines shaping.** You use
the tool differently when it's *for you* than when it's *for your company* —
and "for your company" isn't one thing either. What you tell your manager in a
1:1 is not what reaches your boss's-boss's-boss. So the same corpus produces
different artifacts along an **audience gradient**:

- **Reflect (for me)** — candid: what I did, what I learned, what I struggled
  with, patterns over time. For retros and growth. Complete, honest, includes
  the messy parts.
- **Promote (for my company)**, by altitude:
  - **Manager (1:1 / weekly)** — what shipped, blockers, what's next. Tactical, fairly complete.
  - **Skip-level / director** — outcomes grouped by initiative; less detail, more "so what."
  - **Executive** — business impact only, terse, promotional; "what shipped and why it mattered."

bragfile does the **rule-based shaping** — select, group, and attach a
per-audience **framing directive** (deterministic, local, no network). An
**LLM does the prose** — the bundle is emitted for synthesis, the model is
never baked into the binary (same pipe posture as `brag review`/`summary`,
which already say "paste into an AI session"). bragfile owns *data + shaping*;
the LLM owns *words*.

## Why Now

1. **The corpus and the write spine exist.** v0.3.0 shipped; there is a real,
   dogfooded, provenance-tagged corpus to read. Capture is only half the value.
2. **It's the original north-star, finally.** "Capture accomplishments for
   retros, reviews, and resumes" (AGENTS.md §1) has been deferred through
   PROJ-001/002/003. This is the payoff.
3. **It's the foundation the team layer reuses.** The single-user synthesis
   logic (select → group → frame → synthesize) is the same at wider scope;
   building it first de-risks PROJ-005's federation with **zero federation
   risk now** (no sharing, no transport, no identity problem).
4. **Additive, no migration.** It reads existing data — a clean v0.4.0 minor
   release, not a schema wave.

## Success Criteria

- **`brag impact`** produces a rule-based impact digest — time-windowed
  (`--quarter|--month|--year|--since`), grouped by initiative/project, with the
  `impact` fields surfaced. The deterministic data foundation the story reads.
- **`brag story --audience <me|manager|skip|exec>`** emits an audience-shaped
  bundle — demonstrably different **selection, grouping, framing, and length**
  per audience — ready for LLM synthesis. Same corpus → different artifacts.
- **The `me` story is candid** (learnings, patterns, struggles); the
  **`exec` story is impact-forward, selective, promotional.** The difference is
  visible and rule-driven, not just a tone hint.
- **Synthesis stays a pipe** — a documented bundle format + a bragfile-owned
  framing-directive asset (like BRAG.md); no model in the binary; local-first /
  no-network intact.
- **Ships as v0.4.0.** The full v0.2/v0.3 surface is unchanged; the core needs
  **no schema migration**.

## Scope

### In scope
- `brag impact` — the rule-based, time-windowed impact digest (retro **P1**).
- `brag story --audience …` — the audience-shaped bundle + the framing-directive
  assets. **The headline.**
- The **audience taxonomy + shaping rules** (me / manager / skip / exec →
  selection filter, grouping, framing directive, target length). Its own DEC.
- The **bundle format** an LLM consumes, and the framing-directive assets.
- Optionally `brag wrapped [year]` (shareable year-in-review) and folding in the
  drafted **SPEC-045** (P3 provenance-share) as a personal "how much of my work
  was agent-assisted" metric.

### Explicitly out of scope (→ PROJ-005+ / later)
- **Team / multi-user federation + the DuckDB warehouse** (spike-validated, but
  a separate wave — the whole "share" half).
- **The token / economics dimension** (needs a data source; the exec-ROI story).
  *But see Dependencies — a tiny "start capturing cost now" MCP change may be
  worth slipping in early so history accrues, like provenance did.*
- **Cross-person taxonomy reconciliation**, sharing/transport/consent.
- **Baking an LLM into the binary** — synthesis stays a pipe.
- Goals as an object type; macOS notarization.

## Stage Plan

Ordered; a project typically has 2–5 stages. IDs are the next-free repo-global
(§2): STAGE-011+, SPEC-046+, DEC-027+.

- [ ] STAGE-011 (proposed) — **`brag impact` — the digest foundation.** The
      rule-based, time-windowed, initiative-grouped impact aggregation the story
      reads. Reuses `internal/aggregate`; DEC-014 envelope.
- [ ] STAGE-012 (proposed) — **`brag story --audience` — the narrative surface.**
      The audience taxonomy + shaping rules (a DEC), the LLM-pipe bundle format,
      the framing-directive assets. The headline.
- [ ] STAGE-013 (proposed, maybe) — **Polish + v0.4.0 cut.** `brag wrapped`,
      the P3 personal metric, output adapters if earned, and the v0.4.0 release
      cut (per the `spec-release-cut` template + §4).

**Count:** 0 shipped / 0 active / 3 proposed

## Dependencies

### Depends on
- **PROJ-003 (v0.3.0, shipped).** The corpus + provenance; `internal/aggregate`
  (`ByType`/`ByProject`/`Streak`/`Span` — the digest's building blocks);
  `brag list --author` (the P3 basis); `brag review`/`summary` as precursors.
- **DEC-014** (rule-based output envelope) — `brag impact` extends the family.
- **A framing-directive asset convention** (like BRAG.md / the plugin assets).

### Enables
- **PROJ-005+ (team federation + story + economics).** The single-user
  synthesis logic (select → group → frame → LLM) is reused at team/org scope
  over the spike-validated warehouse; the exec-ROI story adds the token
  dimension on top.
- **Consider seeding early:** a minimal MCP `brag_add` `tokens:`/`cost:` capture
  (capture-time, self-reported) could slip into v0.3.x/v0.4.0 so cost history
  starts accruing *before* the economics layer exists — the same lesson
  provenance just taught (the corpus had **0** agent-authored history because we
  stamped late). Small, decoupled, high-leverage for PROJ-005.

## Project-Level Reflection

*Filled in when status moves to shipped.*

- **Did we deliver the outcome in "What This Project Is"?** <tbd>
- **How many stages did it actually take?** <tbd>
- **What changed between starting and shipping?** <tbd>
- **Lessons that should update AGENTS.md, templates, or constraints?** <tbd>
- **What did we defer to the next project?** <tbd>

### Numbering (at framing)

Highest consumed: PROJ-003, STAGE-010 (referenced/never-activated), SPEC-045
(P3 draft), DEC-026. Next free (repo-global monotonic, §2): **PROJ-004**,
**STAGE-011**, **SPEC-046**, **DEC-027** — the first DEC (likely the audience
taxonomy) is assigned at emission, not pinned here.
