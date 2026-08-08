---
# Maps to ContextCore project.* semantic conventions.
# A project is a bounded wave of work against the repo (the app).

project:
  id: PROJ-006
  status: active
  # activity = the type of work currently active within the project. The coarse
  # `status` (active/shipped/archived) is what tooling keys on; `activity` is the
  # human-facing detail. Suggested vocabulary (extend as needed):
  #   requirements | design | build | test | blocked
  activity: design
  priority: high
  target_ship: null

repo:
  id: bragfile

roadmap:
  - pillar: Corpus-as-memory
    resume_when: "repeatedly asking 'what did I do on X', or re-deriving context"
  - pillar: Signed provenance
    resume_when: "acting on agent/model/cost data, or sharing the corpus"
  - pillar: Git-hash / evidence links
    resume_when: "wanting to tie a brag to the commit that proves it"
  - pillar: Benchmark
    resume_when: "reliably stamping cost/tokens and wanting to compare models"
  - pillar: Capture completeness
    resume_when: "the corpus feels sparse, or you notice unlogged work"
  - pillar: Narrative / story surface (v2)
    resume_when: "wanting a report that shows the arc (foundation -> payoff), not a flat win-list"

created_at: 2026-07-10
shipped_at: null
---

# PROJ-006: Agent-Native Depth — Memory, Trust & Completeness

> **Goal — make bragfile's corpus a working memory, not a write-only log.**
> bragfile already captures what you and your agents ship; PROJ-006 makes that
> corpus something the agents **read back, trust, and are measured by.** The
> insight: *the corpus's highest-frequency reader is the agent that produced it.*
> So the arc is write-only journal → **working memory** — history an agent
> consults *before* it acts (stops re-deriving context, stops repeating dead
> ends), with provenance it **can't fake** (so it can drive real decisions),
> complete enough to be worth reading, and measurable enough to answer "which
> model earned its keep." All local, all developer-owned — the one memory layer
> an agent has that no cloud console can see. The pillars are that arc as a
> causal chain: **consult → trust → complete → measure.**

> **Requirements-gathering phase (active).** Opened 2026-07-10 as the successor
> to PROJ-005 (which shipped the *opening* of agent-native depth as v0.5.0). The
> direction below is set by `docs/research/proj-005-idea-synthesis.md`; the
> **stages and specs are NOT framed yet — and that is deliberate, not overdue.**
> This project is in *requirements gathering*: we are actively deciding *which*
> pillar is worth building and *what shape* it should take before any spec is
> written. Treat everything under "Candidate scope" and "Candidate stage plan" as
> requirements input, not a commitment.
>
> **Why we gather before we frame (2026-07-11).** After v0.5.0/v0.5.1 the product
> is a complete personal tool; the deeper pillars here are bets on a future
> (agents reading history, provenance-driven decisions, benchmarking) that real
> usage should confirm before we build. Deliberate stance: *use it, and let the
> next spec come from a recurring annoyance* (as `brag spark` and `brag list
> --day` did). A pillar graduates from requirement → framed stage when its signal
> below actually shows up in practice — not on a schedule. Having no specs in
> flight is the *expected* state of this phase, not a sign the project is done.
>
> **Requirements signals we're gathering (which pillar, and the evidence that
> would pull it into framing):**
> - Corpus-as-memory — you/an agent repeatedly asking "what did I do on X" or
>   re-deriving project context (note: `--day` + `spark` may already scratch this).
> - Signed provenance — you *acting on* `agent:`/`model:`/`cost:` data, or sharing
>   the corpus with anyone.
> - **Git-hash / evidence links** — you wanting to tie a brag to the actual commit
>   that proves it (see the note in Candidate scope; a `commit:<hash>` freeform
>   tag works today).
> - Benchmark — you reliably stamping `cost:`/`tokens:` AND wanting to compare models.
> - Capture completeness — the corpus starts feeling sparse / you notice unlogged work.
>
> **Observations log (requirements captured as they surface — the working
> output of this phase):**
> - **Git-hash evidence links (2026-07-12).** Establishing the *pattern* before
>   framing a stage: a `commit:<hash>` freeform tag (works today, no code) ties a
>   brag to the commit that proves it — self-attesting provenance. Documented the
>   convention in [`BRAG.md`](../../BRAG.md) ("Evidence links: the `commit:` tag");
>   now dogfooding it in real project brags. Signal to watch: does the tag get
>   used naturally, and do we start wanting it *validated/typed*? If yes, that
>   graduates the capture-completeness pillar (evidence-links slice) into framing.
> - **Outward-impact capture landed in the templates (2026-07-16).** The spec
>   `Reflection (Ship)` gained a Q4 "what can a user do now that they couldn't?"
>   and the release template gained a one-line Cut record — so per-spec outcomes
>   are now captured at ship (before → after), the raw material the story surface
>   reads. This is the *enabler* for the new Narrative / story-surface (v2)
>   candidate below; its signal to watch: once a stage's specs carry real Q4
>   answers, do we want them assembled into a stage story / an arc in `wrapped`?
> - **Corpus-as-memory baseline is more built than first framed (2026-08-07).**
>   Code check: the MCP read *tools* already ship; the remaining work is MCP
>   resources + `Since/Until/Author` filter parity + a token-budgeted retrieval.
>   Re-scoped in Candidate scope #1 so framing starts from the true baseline.
> - **Corpus-as-memory GRADUATED requirement → framed stage (2026-08-07).** The
>   first deep pillar is now `STAGE-019` (active), with a three-spec backlog
>   (SPEC-072/073/074) and four DECs identified (DEC-042 filter parity + parser
>   home, DEC-043 ranking, DEC-044 token budget, DEC-045 resource shape). The
>   pull was the *cost* signal, not a new annoyance: once the true baseline was
>   verified, the remaining work turned out to be additive plumbing plus one
>   algorithm — no storage change, no migration, no new dependency. The project's
>   requirements phase stays open for the other four pillars; this is one pillar
>   graduating, not the phase closing.

## What This Project Is

The *depth* half of the agent-native wave PROJ-005 opened. Where PROJ-005 made
the MCP path first-class and cleaned the substrate, PROJ-006 makes the corpus
**trustworthy, complete, and consulted by the agents that write it** — still
strictly local-first and single-user. The synthesis' one-line thesis: the
corpus's highest-frequency reader is the agent that produced it, and the moment
provenance drives any decision it must be unforgeable.

## Why Now

v0.5.0 shipped the ergonomic front door (agents can connect + log correctly) and
a hardened substrate. That is the precondition for the deeper work: giving agents
the **read** role, making provenance **unforgeable**, and making the corpus
**complete** enough to read and (eventually) benchmark. Federation / economics /
SaaS still build on top of this and remain out of scope until the corpus is
trustworthy and complete.

## Success Criteria

*To be set at framing (next session).* Likely themes: an agent can read its own
ranked history over MCP before it works; a `verified` vs `claimed` provenance
tier exists with a local signing scheme + `brag verify`; the corpus can be
cold-started/completed (inbox + git-import + evidence links); and there is a
rule-based agent/model benchmark. All local-first, no network in the binary, no
CGO.

## Candidate scope (from the synthesis — to sequence at framing)

The four "act soon" pillars, ranked by the synthesis:

1. **Corpus-as-agent-memory** — read-side MCP an agent consults (ranked by
   project/recency/FTS) before working; the single highest leverage/effort item,
   likely the first stage. **Baseline already ships** (correction, verified in
   code 2026-08-07): the read *tools* exist — `brag_list` (recency-ordered),
   `brag_search` (bm25-ranked), `brag_stats` — so this is *not* "build read-side."
   The real remaining work is three smaller additions:
   - **(a) MCP resources** — today the corpus is exposed only as *tools* (pull:
     the agent must decide to call one). No **resources** exist (push: context the
     client auto-loads before the agent works). Memory = the resource surface.
   - **(b) Filter parity** — the MCP `list` input omits `Since`/`Until`/`Author`,
     which `storage.ListFilter` already supports; add them so an agent can ask for
     "recent" / "agent-authored" history over MCP.
   - **(c) Blended, token-budgeted retrieval** — recency (list) and relevance
     (search) are separate axes today, and `Limit` is a row count, not a token
     cost. The "memory slice" is a ranked blend (project + recency + match)
     trimmed to a token budget so it's cheap enough to auto-load every session.
2. **Signed / attestable provenance** — sign the `agent:/model:/cost:` block at
   capture with a local secret/keypair → `verified` vs `claimed` + `brag verify`.
   The trust primitive everything downstream rests on. **Subsumes the deferred
   audit finding #3** (reserved-tag forgery via freeform `tags`).
3. **Capture completeness** — a staging inbox (`status=proposed` + `brag
   promote`), a local git-import cold-start miner, and typed evidence links
   (commit/PR/issue refs as rows).
   - **Git-hash evidence links may be the highest-value / cheapest provenance
     slice** (idea, 2026-07-11). Tying a brag to the commit that proves it is a
     *self-attesting* form of provenance — anyone can check the commit exists,
     its author, and its date against the repo — so it's arguably a stronger and
     cheaper win than #2's signing (which makes *self-reported* tags unforgeable
     via crypto). A `commit:<hash>` freeform tag works TODAY with no code; the
     spec would promote it to a typed, validated evidence link (and pairs with
     the git-import miner). Strong candidate for an early PROJ-006 stage.
4. **Agent/model benchmark** — `brag benchmark --by model`: impact-per-1k-tokens-
   per-dollar over the provenance tags; pure aggregation, no model in the binary.
   Gated on #2 (trust) being real.

**Narrative / story surface (v2) — new candidate (2026-08-07, NOT from the
synthesis).** Make the story surface first-class to the *process* and to
*agents*. Four parts:
- **Stage-close story (the storytelling ladder).** Force an outcome story at
  stage close, assembled from each spec's Reflection-**Q4** answers. Ladder:
  spec = capability (Q4) → stage = paragraph → project/release = chapter.
- **`brag wrapped` impact-arc.** `wrapped` today does counts + cadence and does
  **not** read `.Impact`; compose the existing `aggregate.WithImpact` digest,
  window by month/quarter, and surface a **foundation → payoff arc** (not a flat
  win-list) via reference-thread hints + an LLM framing directive, cross-window
  via `--previous`.
- **`arc:foundation` / `arc:payoff` reserved tag.** LLM-classified by default
  (impact language + position tags + reference threads); the explicit tag is
  authoritative where present; untagged = neutral connective work. Rides the
  reserved-tag model, zero schema change.
- **MCP exposure** of story/wrapped: deterministic bundle from the binary, prose
  from the LLM (DEC-029).
*Depends on* the outward-impact capture (Reflection-Q4/Q5 + release Cut record,
landed in the templates 2026-07-16) being populated, and shares the MCP read
surface with #1. *Why it matters:* foundation work gets credit **through the
payoff it enabled** — a progression story only the corpus can tell (it has the
reference threads linking cause to effect).

Promising-but-spike-first (from the synthesis §B): bragfmt interchange /
`import --format jsonl`; outcome reconciliation ("did the brag hold up");
anti-surveillance share shape; `brag reconcile` as divergence audit.

### Also inherited: the v0.5.0 audit backlog (LOW/NITs, slot in opportunistically)
`mcp_install` atomic write (temp+rename); MCP `list/search` negative-`limit`
parity; `search -foo` cobra-flag error; `brag project new` name cap (+ run the
edit/`Store.Update` path through `internal/capture.Validate`); the `brag spark`
same-second exclusive-edge; `ParseSince` wall-clock impurity; backup-filename
same-second collision; empty-`type` sentinel; export-md sort id-tiebreak;
`MergeTags` position dup; double-wrapped db-path error; `$EDITOR`-with-spaces.

## Stage plan

**Shipped (a tactical read-ergonomics opener, not one of the deeper pillars):**
- [x] STAGE-017 (shipped 2026-07-11) — list time-window ergonomics: `brag list
      --day` (v0.5.1). A quick win that opened PROJ-006; not agent-native depth.

**Proposed (parked, not scheduled):**
- [ ] STAGE-018 (proposed) — v0.5.0 audit backlog (LOW/NITs): a framed parking
      lot for the small correctness nits, to slot in opportunistically. Not a
      deeper pillar; created so the backlog lives in the hierarchy, not just here.

**Active (the first deep pillar, framed 2026-08-07):**
- [ ] STAGE-019 (active) — **corpus as agent memory**: MCP `brag_list` filter
      parity (`since`/`until`/`day`/`author`, retiring the `cli↔mcpserver` cycle),
      a deterministic blended + token-budgeted memory slice (`internal/memory` +
      `brag memory`), and the MCP **resources** push surface. Backlog:
      SPEC-072/073/074; DECs: DEC-042 (parser home + MCP time vocabulary),
      DEC-043 (ranking), DEC-044 (token budget), DEC-045 (resource shape).

**Candidate deeper pillars (NOT framed — for a later session's discussion):**
- [ ] (STAGE-020?) — signed / attestable provenance (+ closes the tag-forgery gap)
- [ ] (STAGE-021?) — capture completeness (inbox / git-import / evidence links) —
      includes promoting the `commit:`/`pr:`/`issue:` evidence-link convention
      (now documented in BRAG.md + being dogfooded) to typed, validated links.
- [ ] (STAGE-022?) — agent/model benchmark
- [ ] (STAGE-023?) — narrative / story surface v2 (stage-close story + `wrapped`
      impact-arc + `arc:` tag + MCP exposure); depends on Q4 capture being
      populated + shares #1's MCP read surface
- (sequence, split, and de-scope at framing; IDs assigned at creation)

**Count:** 1 shipped / 1 proposed (parked) / 1 active / (remaining deeper pillars unframed)

## Dependencies

### Depends on
- PROJ-005 (shipped v0.5.0): the ergonomic MCP path (`brag mcp install`, the tool
  contract + docs), the unregistered-project primitive (`brag project ensure`),
  the hardened concurrency/validation substrate, and `internal/capture.Validate`
  / `internal/spark` / `aggregate.RollingBuckets` to build on.

### Enables
- The eventual team / economics / SaaS story — which the synthesis says must wait
  until the corpus is trustworthy (signed) and complete, both of which are this
  project's job.

## Project-Level Reflection

*Filled in when status moves to shipped.*

- **Did we deliver the outcome in "What This Project Is"?** <yes/no + notes>
- **How many stages did it actually take?** <number, compare to plan>
- **What changed between starting and shipping?** <one or two sentences>
- **Lessons that should update AGENTS.md, templates, or constraints?**
  - <one-line updates>
- **What did we defer to the next project?**
  - <one-line items>
