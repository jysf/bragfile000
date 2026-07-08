# PROJ-005 idea synthesis — recommendations for the orchestrator

*Coalesced 2026-07-07 from a bounded 3-lens internal fan-out (market/SaaS,
technical/integrations, out-of-the-box), using the rubric in
[`proj-005-idea-sourcing-prompts.md`](proj-005-idea-sourcing-prompts.md).
Consensus = a signal; a lone contrarian pick = a different signal. Decision-ready
input for the fresh PROJ-005 session — not a survey.*

## Headline (the reframe)

The strongest ideas across all three lenses are things bragfile can build
**single-user, local, now** — and doing them **makes the eventual team /
economics / SaaS story trustworthy and complete before we take on federation's
complexity.** Recommendation: **the next wave should be "agent-native depth"
(local, single-user), not team federation.** Federation/economics/SaaS build on
top of a corpus that is *complete*, *trustworthy*, and *consulted by agents* —
none of which is true yet.

The SaaS hunch, interrogated: **directionally real but mis-shaped.** A personal
brag doc has weak willingness-to-pay, and pointing it at team-reporting lands in
the crowded, developer-*disliked* dev-analytics market (Jellyfish/LinearB/
Swarmia) that auto-pulls from git and needs no self-report. The defensible
business is **"the join"** (federation + narrative + economics) sold to a
platform / eng-effectiveness / finance buyer — with the free local CLI as the
funnel, and **developer-owned / opt-in / anti-surveillance** (local-first as the
*moat*) as the one lane incumbents structurally can't follow.

## Consensus clusters (independently found by ≥2 lenses)

| Cluster | Lenses | Read |
|---|---|---|
| **Signed / attestable provenance** | market + out-of-box + technical (all 3) | `agent:/model:/cost:` are self-reported, trivially spoofable. Sign at capture (local HMAC/keypair — no server). The trust primitive everything downstream rests on. |
| **Corpus-as-agent-memory** (read-side MCP) | out-of-box + technical (both top-3'd it) | Agents *read* their own history as durable context. Reframes bragfile from capture-sink to agent **infrastructure**. Nearly free — the query layer already ships. |
| **Open interchange standard** (bragfmt / "provenance-stamped work" format) | all 3 | Own the *format*, not the CLI. Input interop (any tool writes conformant brags) > more output adapters. |
| **Anti-surveillance / developer-owned share** | market + out-of-box | Aggregate-only / k-anonymized rollups + a consent model. The moat and the only durable SaaS lane. |
| **Corpus completeness (fill-itself)** | market + technical | A sparse, self-reported corpus is fatal to every read/share/economics surface. Needs a staging **inbox**, a **git-import** cold-start, and **evidence links**. |
| **Outcome reconciliation** ("did the brag hold up") | out-of-box + technical | Reverted/decayed/rework signal → the credible counterweight to a rosy ROI story. |
| **Agent/model benchmark** ("which tool earns its keep") | technical + market | provenance + cost + *outcome* co-located locally = a data product no console or dev-analytics tool has. |

## Ranked recommendations

### (A) ACT SOON — at-bar to frame now (local, single-user, high leverage/effort)

1. **Corpus-as-agent-memory — read-side MCP resources.** *Problem:* agents
   re-derive project context every session and repeat dead ends. *Shape:* expose
   the corpus as MCP resources/tools an agent reads (ranked by project/recency/
   FTS) before it works. *Why now:* nearly free (wrap existing `List`/`Search`),
   pure on every non-negotiable, and it turns bragfile into infrastructure agents
   *depend on*. *Risk:* keeping the returned slice token-cheap enough to actually
   get loaded. *Decomp:* 1 stage, ~2 specs (the MCP resource shape + a relevance
   ranking). **The single highest leverage/effort item in the set.**

2. **Signed / attestable provenance.** *Problem:* the moment `cost:`/`model:`/
   `agent:` drive a routing, purchasing, hiring, or credit decision, self-report
   becomes worth gaming. *Shape:* sign the provenance block at capture with a
   local secret/keypair the hook/MCP holds → a `verified` vs `claimed` tier, plus
   `brag verify`. *Why now:* build the trust tier *before* the incentive to spoof
   exists — it's load-bearing for economics, sharing, hiring, and benchmark.
   *Risk:* key management (where agent keys live, rotation) + a canonical
   serialization to hash. *Decomp:* a DEC on the signing scheme + 1–2 specs.

3. **Capture completeness: inbox → git-import → evidence-links.** *Problem:*
   cold-start (empty corpus) and self-report decay make the corpus too sparse to
   read or sell. *Shape:* a staging **inbox** (`status=proposed`, `brag promote`)
   so detection can be greedy without polluting the curated corpus; a local
   **git-import** miner that clusters your existing history into candidate brags;
   typed **evidence links** (commit/PR/issue refs as rows). *Why now:* it's the
   precondition for every downstream surface, and the inbox is the substrate that
   makes all future fill-itself integrations *safe*. *Risk:* commit→brag
   heuristic quality; keeping the inbox from becoming noise. *Decomp:* a stage
   (inbox migration + import + evidence side-table), ~3 specs.

4. **Agent/model benchmark — `brag benchmark --by model`.** *Problem:* "is
   Opus-tier worth it? which agent do I route to?" *Shape:* a rule-based digest —
   impact-bearing brags per 1k tokens per dollar, per model/agent — over the
   reserved tags, in the DEC-014 envelope. *Why now:* the most differentiated
   data product the substrate uniquely enables; pure aggregation, no model in the
   binary. *Risk:* depends on `cost:`/`tokens:` being populated *and* trustworthy
   → pairs with #2 (attestation). *Decomp:* 1 spec on top of `IsAgentAuthored`.

### (B) PROMISING — spike or shape before committing

- **Open interchange schema (bragfmt).** Do the *cheap* half now — document the
  entry+provenance shape + the `brag_add` contract as a versioned JSONL format
  and add `import --format jsonl` (input interop). Defer the "own the category /
  standards-body" ambition (adoption is a distribution problem, high variance).
- **Outcome reconciliation / rework signal.** Spike the signal quality first —
  attributing "decay" vs a legitimate refactor is genuinely hard (false
  positives). Depends on evidence-links (A#3).
- **Anti-surveillance share + developer-owned positioning.** This is the
  *consent shape* of the team wave — fold it into PROJ-005 federation scoping as
  a first-class constraint (aggregate-only / k-anon / opt-in), not a bolt-on.
  k-anonymity is hard on small teams — needs design.
- **Economics buyers + the human-time axis.** Agency **billing receipts** and
  **R&D-tax / compliance substantiation** are the strongest "who pays and why"
  (a *finance* buyer with a real dollar incentive); **time accounting** adds the
  human-hours half of "time *and* tokens." All gated on the economics substrate
  + attestation landing first — park the packaging, keep the buyers in view.
- **`brag reconcile` as audit.** The novel slice beyond the roadmap's
  session↔usage join: framing reconciliation as *divergence detection* (claimed
  vs actual cost) — a spoof tripwire, not just backfill. Pairs with #2/#4.
- **Non-engineer capture.** Decouple capture from the commit-centric model to
  reach agent-using PMs/designers/researchers/lawyers/founders (largest untapped
  market). Spike the distribution question — it likely needs a non-CLI front
  door, a real tension with the CLI identity.

### (C) PARK — interesting, low-leverage-now or premature

- **Anti-brags / `brag learn`** (capture failures) — cheap (S) and a structural
  anti-bragflation move + agent-memory synergy; promote to (A) if the memory
  work lands. **Longitudinal self-model / `brag trajectory`** (decade-scale
  personal work-graph — the durable single-user moat). **Bragflation lint /
  `brag audit`** (integrity guard; do it when metrics start being shared/gamed).
  **Job-seeker pack** (JD-tailored prep — incremental on the shipped pipe).
  **Public build-in-public profile** (organic acquisition, but hosted →
  relaxes local-first). **Attribution ledger** (mostly emerges from signing +
  evidence-links + coverage).

### (D) NON-NEGOTIABLE WATCH (constrain, don't necessarily reject)

- **Connectors (Linear/Jira/CI), calendar import, hosted tiers, public
  profile** — permissible **only** as *external adapter processes* / *opt-in
  hosted layers atop* a local core; **never network inside the `brag` binary**
  (same boundary as the LLM pipe). Hold the line explicitly or they break
  local-first.
- **Peer / agent co-signing (web-of-trust)** — relaxes single-user-first; a
  genuinely novel primitive (an agent as *witness*, not author) but scope-creepy;
  gate on signing (A#2) and treat as team-wave.

## Cross-cutting themes

1. **Trust is the through-line.** Signing → benchmark → economics → sharing all
   rest on provenance being unforgeable. It's currently a spoofable tag. This is
   the most-repeated finding across lenses.
2. **Completeness is the precondition.** Every read/share/economics surface dies
   on a sparse corpus. The inbox + git-import unblock is high-leverage plumbing.
3. **The reframe: agents as READERS, not just authors.** The corpus's
   highest-frequency reader is the agent that produced it. The roadmap never
   gives agents the read role — memory does.
4. **The SaaS shape:** business = "the join" (federation+narrative+economics)
   sold to a platform/finance buyer; personal doc = funnel; **developer-owned /
   anti-surveillance = the only defensible lane.**

## What this changes about the PROJ-005 plan

- **Re-sequence.** Do **agent-native depth (local, single-user)** — memory,
  signed provenance, benchmark, capture-completeness — as the next wave, *before*
  team federation. It's cheaper, purer, defensible, and it's what makes the
  eventual team/economics/SaaS story trustworthy and complete.
- **When federation does come:** consent / aggregate-only / attestation are
  **first-class** (the anti-surveillance moat), not afterthoughts. Identity/dedup
  (the spike's "genuinely hard part") remains the first federation domino.
- **Reframe economics:** from "what did it cost" (ROI) to **comparative
  benchmark + audit** (divergence detection) — more credible to a skeptical
  buyer, and more differentiated, than a spend dashboard.

## The single highest-leverage next thing

**Corpus-as-agent-memory (A#1), paired with signed provenance (A#2).** Cheapest,
purest, reframes the product from reporting-output into agent infrastructure, and
lays the trust foundation everything downstream needs — all single-user, local,
zero federation risk. A consensus-adjacent call: the two lateral/technical lenses
independently top-3'd memory; all three top-3'd signing.
