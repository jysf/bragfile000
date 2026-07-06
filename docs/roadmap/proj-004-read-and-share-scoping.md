# PROJ-004 (candidate) — "read & share" scoping note

> **Status: SCOPING / pre-brief — NOT framed.** A candidate-scope stash, not a
> project brief. PROJ-004's real brief is written at framing (a separate
> session). `PROJ-004` is the next free repo-global id (§2); the slug is
> provisional. Captured 2026-07-05; **updated 2026-07-06** to fold in the
> DuckDB federation spike ([`docs/research/duckdb-federation-spike.md`](../research/duckdb-federation-spike.md))
> and the storytelling thesis (see "The product is narrative").

## The product is narrative (the reframe)

**The warehouse is the substrate; the *product* is story.** PROJ-004 turns the
brag corpus into **audience-shaped narratives** along two axes:

- **Scope** — *your* story (single-user: "here's what I got done" — what a user
  should be doing/reflecting on) → *your team's* story (people **and** agents
  woven into one arc) → *multiple teams* → the org.
- **Audience** — self-reflection → peer → manager → **executive** ("what did we
  do, why, what did it cost, and what can we learn").

This is the original north-star of bragfile ("capture accomplishments for
retros, reviews, and resumes", AGENTS.md §1), deferred through
PROJ-001/002/003 — now extended into the agent-native, multi-actor era.
v0.3.0 built the agent-native *write* spine and made provenance measurable
(`brag list --author`); PROJ-004 is the *read & share* payoff.

### The two angles (the user's framing)

1. **Single-user — "telling your own story."** The traditional, high-value
   default: what *you* got done, shaped for self-reflection / review / résumé.
   No federation. Ships first.
2. **Team / multi-agent / multi-team — "telling the combined story."** Weave
   the work of people **and** agents (however the team is composed) into one
   narrative; then weave multiple teams. More than a status report — a
   *story* of how a team spent its **time and tokens**, what resulted, and
   what can be learned, told well enough to explain it up to executives.

### The token/economics dimension (new, differentiated)

"How a team spent its time **and tokens**" is a story no traditional status
report can tell — and it's the sharpest reason this matters in the agent era.
**bragfile does not capture cost today** (provenance is `agent:`/`model:` =
*who*, not `tokens:`/`cost:` = *how much*). So this needs a **data-source
decision** — the single most important new requirement for the exec/ROI story:
- **(a)** the MCP `brag_add` tool accepts a token/cost param the calling agent
  supplies (capture-time, self-reported); and/or
- **(b)** join brags to an external usage source (Claude Code session logs /
  usage API) by session + agent, post-hoc.

### Narrative synthesis is an LLM step, kept as a pipe

Turning structured brags into prose is the "super-brag" synthesis — an LLM
job. bragfile owns the **data + shaping** (emit a clean, provenance-and-cost-
stamped bundle scoped to self/team/audience); an **LLM owns the prose**. This
keeps local-first / no-network intact and matches how `brag review`/`summary`
are already "designed to be pasted into an AI session." Don't bake a model in.

## Layering → sequencing (de-risked, and it matches the two angles)

The single-user story is the **foundation the team story reuses** (same
synthesis logic, wider scope) — so angle 1 isn't a prerequisite to build
before angle 2, it's the first shippable layer that the rest wraps:

0. **Your story** (single-user read/story surface: `brag impact` → `brag story
   --audience` → `brag wrapped` / adapters). **v0.4.0** — additive read
   commands, no federation risk. *Highest value soonest.*
1. **Team federation + story** — the spike-validated warehouse (below) unions
   stamped exports; a team-narrative synthesizer weaves people + agents.
2. **Economics** — the time+token cost dimension → the ROI / executive story.
   Gated on the token data-source decision.
3. **Org weave** — multiple teams' stories woven together.

## What the DuckDB spike SETTLED (federation substrate)

The spike ([full report](../research/duckdb-federation-spike.md)) validated the
mechanics of layer 1. Treat these as decided:
- **Architecture: federated export → DuckDB, materialize-per-source, then
  union.** NOT live multi-attach — that returned *silently wrong* rollups
  (every alias resolved to the first-attached DB). NOT a shared remote DB
  (supersedes DEC-001). The warehouse is **derived & disposable** — rebuilt
  from each day's exports, a pure function of them (no CDC, no incremental
  merge). Daily latency is fine for retro/review cadence.
- **Identity = file-level stamp (Mechanism A), not per-entry `user:`/`host:`
  tags** — B lost synced entries to `NULL source` and pollutes the tag
  taxonomy. Identity is a property of the *export*, not the rows.
- **Two hard constraints on the export path:** (1) raw `~/.bragfile` DBs are
  **not** DuckDB-attachable — the FTS5 trigger DDL crashes DuckDB's SQLite
  parser; the export must strip `entries_fts` + its 6 triggers (or emit
  Parquet). (2) No global entry identity — `entries.id` is per-file
  autoincrement, so cross-machine dedup needs `(user, content-hash)` today, a
  capture-time UUID to be deterministic.
- **No concurrency hardening needed** (WAL/busy-timeout) — export-based
  aggregation has no shared write.

**The minimal bragfile-core change for federation** is therefore small: a
warehouse-friendly export (`brag export --warehouse|--format parquet`) + a
file-level identity stamp. The warehouse/queries themselves can live *outside*
the bragfile repo (derived tooling). That split lets us de-risk incrementally.

## Candidate features (rated by fit)

| Feature | Why it fits | Fit |
|---|---|---|
| `brag impact --quarter\|--month\|--year\|--since` (rule-based impact digest, retro **P1**) | Layer-0 read foundation; the warehouse reuses the same aggregation | **Core** |
| `brag story --audience self\|manager\|exec\|resume\|blog` + AI-pipe "super-brag" synthesis | The narrative product itself — emit a bundle, LLM weaves prose | **Core (headline)** |
| Team/org narrative synthesizer (weave people + agents + multiple teams) | Angle 2 — the combined story | **Core (layer 1+)** |
| Time+token economics (cost per brag/session; team ROI rollup) | The exec story; **needs a token data source** (MCP param vs external join) | **Core (layer 2) — new** |
| Federation: warehouse-friendly export + file-level identity stamp → DuckDB warehouse | Layer-1 substrate; **shape settled by the spike** | **Core** |
| `brag wrapped [year]`, activity sparkline, impact density, `brag achievements` | The shareable insight cluster | **Core** |
| Provenance-share / dogfooding query (retro **P3**, SPEC-045 draft) | Agent-vs-human share **from adoption forward** (not retroactive) | **Core (small)** |
| Output adapters (Notion; résumé / review / promo-packet / exec-deck) | The "share" endpoint — hand-off artifacts | **Core** |
| First-class `agent`/`model` (+ maybe `tokens`) columns | The DEC-004→015 "later, if earned" promotion; economics may force it | **Core-ish** (earned-by-data) |
| Capture-time entry UUID | Deterministic cross-machine dedup; its own DEC + migration | **Layer-1 fast-follow** |
| Goals as an object type | A *planning/forward* object — different shape | **Separate** (~PROJ-005) |
| macOS notarization; WAL concurrency | Ops track / not needed for export-based federation | **Out** |

## Open questions — the real substance of framing

The spike validated mechanics; **these are what PROJ-004 actually decides.**

1. **Share-scoping (the ethical spine — now heavier).** Which entries leave
   your machine — all / opt-in per entry / by project / by tag? You're
   potentially sharing career narrative **plus token/cost data** up to
   executives — sensitive on two axes. **Lean: default private, opt-in, share
   by project** (+ a per-entry override). This is *the* decision that defines
   the product.
2. **Transport + trust model.** Where stamped exports physically meet — a git
   repo of exports / shared dir / object store — and the consent model that
   rides with it. **Lean: a git repo of exports** (push what you choose;
   auditable; no server), object store as the org-scale upgrade.
3. **Token data source.** MCP capture-time param vs external usage-log join
   (see above). Load-bearing for the economics/exec layer.
4. **Identity: content-hash (MVP) vs capture-time UUID (durable).** Hash = zero
   schema change but splits on a post-sync `brag edit`; UUID = deterministic,
   its own DEC + migration.
5. **Cross-person taxonomy divergence.** `project`/`type`/`tags` are free-text
   per person — Alice's `perf` vs Bob's `performance` won't align, so org
   rollups fragment. The spike's single-corpus simulation couldn't surface
   this; the team story needs a reconciliation/mapping story.
6. **Where the narrative synthesizer lives** — a `brag` subcommand emitting a
   bundle for an external LLM, an in-repo prompt/asset (like BRAG.md), or a
   separate tool. Keep the model out of the binary either way.

## Settled going into framing (do not re-litigate)

- Federation = export→DuckDB, materialize-per-source, Mechanism A (spike).
- Warehouse derived & disposable; daily rebuild; local-first preserved (DEC-001).
- Version skew **parked** (assume everyone on latest release).
- Local day: **store UTC, shift in reporting** (DEC-022 semantics).
- SaaS / hosted multi-user: an explicit **non-goal now** (its own future release).
- Goals: **out** (separate project).

## Sources

- [`docs/research/duckdb-federation-spike.md`](../research/duckdb-federation-spike.md) — the federation-mechanics spike (verdict, blockers, identity/dedup, Mechanism A vs B, export requirements).
- [`projects/PROJ-003-agent-native-spine/brief.md`](../../projects/PROJ-003-agent-native-spine/brief.md) — Scope (out-of-scope list), Project-Level Reflection (deferred items).
- [`docs/reports/cross-project/2026-07-04-action-register.md`](../reports/cross-project/2026-07-04-action-register.md) — P1 (impact digest), P3 (dogfooding query).
- [`projects/PROJ-001-mvp/backlog.md`](../../projects/PROJ-001-mvp/backlog.md) — the impact-surfacing / storytelling cluster.
