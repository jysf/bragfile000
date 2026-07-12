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
  activity: requirements
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

created_at: 2026-07-10
shipped_at: null
---

# PROJ-006: Agent-Native Depth — Memory, Trust & Completeness

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
> - _(none yet — 2026-07-11)_

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

1. **Corpus-as-agent-memory** — read-side MCP *resources/tools* an agent
   consults (ranked by project/recency/FTS) before working. Nearly free (wraps
   the shipped query layer); the single highest leverage/effort item. Likely the
   first stage.
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

**Candidate deeper pillars (NOT framed — for next session's discussion):**
- [ ] (STAGE-018?) — corpus-as-agent-memory (read-side MCP resources) — the
      synthesis' #1; likely the first deep stage.
- [ ] (STAGE-019?) — signed / attestable provenance (+ closes the tag-forgery gap)
- [ ] (STAGE-020?) — capture completeness (inbox / git-import / evidence links)
- [ ] (STAGE-021?) — agent/model benchmark
- (sequence, split, and de-scope at framing; IDs assigned at creation)

**Count:** 1 shipped / 0 active / (deeper pillars unframed)

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
