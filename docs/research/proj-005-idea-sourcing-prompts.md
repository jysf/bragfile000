# PROJ-005 idea-sourcing prompts

Two reusable prompts for sourcing and coalescing ideas about where bragfile
goes next. **Prompt 1** is fan-out (send to several *different* models — the
diversity comes from the models); **Prompt 2** coalesces their outputs into
decision-ready recommendations for the orchestrator. Saved 2026-07-07 after the
v0.4.0 ship. A first internal fan-out + synthesis lives alongside this file.

---

## Prompt 1 — Idea sourcing (send to several different agents)

```
You are a product/technical strategist doing divergent idea generation for an open-source tool called bragfile. Read the brief, then propose novel directions that are NOT already on its roadmap.

BRIEF — what bragfile is:
bragfile is a local-first CLI (Go + SQLite, zero network) for capturing career-worthy accomplishments ("brags") and reading them back as reports and audience-shaped narratives. It's built for the agent era: AI coding agents capture work autonomously via an MCP server + a Claude Code plugin (a Stop hook nudges capture after commits), stamping provenance (agent:/model: tags); humans capture too via `brag add`.

Current state (v0.4.0, just shipped):
- WRITE: `brag add` (CLI + $EDITOR), `brag mcp serve` (agents capture over MCP), a Claude Code plugin, provenance tags, and a seed of reserved session:/cost:/tokens: tags for future economics.
- READ: `brag list/search/show`, digests (`summary`/`review`/`stats`), and the v0.4.0 story surface — `brag impact` (impact digest), `brag story --audience me|manager|skip|exec` (coalesces brags into audience-shaped narrative ARCS; a pure LLM PIPE — the tool emits a shaped bundle + a framing directive, an external LLM writes the prose; NO model in the binary), `brag wrapped` (year/quarter review), in-terminal sparklines, and `brag coverage` (agent-vs-human provenance share).

Non-negotiable philosophy:
- Local-first, no network, NO model baked into the binary (LLM synthesis is always a pipe).
- Agent-native (agents are first-class authors).
- Spec-driven rigor (frame→design→build→verify→ship, with decision records).
- Single-user first; team/org is a later wave.

ALREADY ON THE ROADMAP — do not just re-suggest these; go beyond them:
- Team federation (export→DuckDB warehouse, materialize-per-source, never a shared remote DB), token/cost economics (the ROI story, gated on reconciling the session: key against provider usage logs), org-weave, cross-person taxonomy reconciliation, sharing/transport/consent, global entry-identity + dedup, hosted/SaaS.
- Smaller known items: output adapters (resume/LinkedIn/review-doc), goals-as-an-object-type, minor storage refactors.
(If you have access to the repo, also read AGENTS.md, docs/roadmap/proj-004-read-and-share-scoping.md, and docs/research/duckdb-federation-spike.md.)

YOUR TASK:
Propose 8–12 novel, high-leverage ideas for where bragfile could go next — across features, use cases, integrations, technical directions, distribution, positioning/market, DX, or risks worth pre-empting — that are NOT already on the roadmap. Bring your own distinct lens; favor non-obvious, high-leverage ideas over safe increments.

For EACH idea, give exactly:
- Name (short)
- What it is (2–3 sentences)
- The user & the job it serves (who is this for, what are they trying to do)
- Fit or tension with the philosophy (local-first / agent-native / pipe / single-user-first)
- Rough effort (S/M/L) + key dependencies or unknowns
- Category (write-side | read-side | integrations | economics | team | distribution | DX | positioning | risk)
- One line: "why this over the obvious alternative"

End with: your top 3 by conviction, and one idea you think everyone else will miss. Flag any idea that would require relaxing a non-negotiable, and say so explicitly.
```

---

## Prompt 2 — Coalesce the results into recommendations (run once, on all outputs)

```
You are synthesizing several independent research agents' idea lists for the tool "bragfile" into a decision-ready set of recommendations for its ORCHESTRATOR — who drives work through a spec-driven frame→design→build→verify→ship cycle and needs recommendations, not a survey.

CONTEXT: [paste the same BRIEF from Prompt 1 so you know the philosophy + what's already on the roadmap]

INPUTS: the raw idea lists from N agents are below.
[PASTE each agent's output, labeled Agent A / B / C ...]

DO THIS:
1. DEDUPE & CLUSTER — merge overlapping ideas across agents into distinct concepts. For each cluster, note which agents proposed it (multi-agent consensus is a signal; a lone contrarian pick is a different signal — keep both).
2. SCORE each cluster 1–5 on: strategic fit (with the read-&-share north-star + the agent-native thesis), user value, differentiation (vs traditional brag-docs / status-report / dev-analytics tools), effort (score 5 = cheapest), and philosophy-fit (local-first / pipe / single-user-first). List dependencies and the biggest unknown.
3. RANK into four buckets:
   - (A) ACT SOON — at-bar to frame as a project/stage now.
   - (B) PROMISING — worth a spike or more shaping before committing.
   - (C) PARK — interesting but low-leverage or premature; say what would change that.
   - (D) REJECT — conflicts with a non-negotiable; name which one.
4. For every (A) and (B) item, write a FRAMING-READY capsule: the problem, the proposed shape, why-now, the single biggest risk/unknown, and a rough project→stage→spec decomposition the orchestrator could frame directly.
5. CROSS-CUTTING: name any themes that span ideas, anything that should change the existing PROJ-005 plan (team federation / economics / identity-dedup), and the ONE highest-leverage thing to do next — with your reasoning, and whether it's a consensus or contrarian call.

Be decisive and opinionated. Prefer a short ranked list of strong bets over a long flat list. Where agents disagreed, say who was right and why.
```
