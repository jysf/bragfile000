---
# Maps to ContextCore project.* semantic conventions.
# A project is a bounded wave of work against the repo (the app).

project:
  id: PROJ-006
  status: active
  priority: high
  target_ship: null

repo:
  id: bragfile

created_at: 2026-07-10
shipped_at: null
---

# PROJ-006: Agent-Native Depth — Memory, Trust & Completeness

> **Shell / direction-only.** Opened 2026-07-10 as the successor to PROJ-005
> (which shipped the *opening* of agent-native depth as v0.5.0). The direction
> below is set by `docs/research/proj-005-idea-synthesis.md`; the **stages and
> specs are NOT framed yet** — that is the job of a dedicated next session. Treat
> everything under "Candidate scope" and "Candidate stage plan" as the
> discussion starting point, not a commitment.

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

## Candidate stage plan (NOT framed — for next session's discussion)

- [ ] (STAGE-017?) — corpus-as-agent-memory (read-side MCP resources)
- [ ] (STAGE-018?) — signed / attestable provenance (+ closes the tag-forgery gap)
- [ ] (STAGE-019?) — capture completeness (inbox / git-import / evidence links)
- [ ] (STAGE-020?) — agent/model benchmark
- (sequence, split, and de-scope at framing; IDs assigned at creation)

**Count:** 0 shipped / 0 active / 0 pending (unframed)

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
