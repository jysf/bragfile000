---
# Maps to ContextCore project.* semantic conventions.
# A project is a bounded wave of work against the repo (the app).

project:
  id: PROJ-005
  status: shipped
  priority: high
  target_ship: null

repo:
  id: bragfile

created_at: 2026-07-10
shipped_at: 2026-07-10
---

# PROJ-005: Agent-Native Depth

## What This Project Is

The wave that turns bragfile from a capture-and-report tool into
**infrastructure that agents depend on** — making the corpus trustworthy,
complete, and *consulted* by the agents that write it, all while staying
strictly local and single-user. It opens by making the existing MCP
server first-class (so the AI-native path is real, not a fallback to the
CLI) and by clearing the small debts and read/capture-surface gaps that
the story-surface wave (PROJ-004) left behind. Federation, economics, and
any SaaS story are explicitly *later* — they build on top of a corpus
that is complete and trustworthy, neither of which is true yet.

See `docs/research/proj-005-idea-synthesis.md` for the full reframe
("agent-native depth before federation") this project executes against.

## Why Now

v0.4.0 shipped the story surface (read side for *humans*). The synthesis
fan-out found the highest-leverage next work is single-user, local, and
cheap: give **agents** the read role, make the MCP path ergonomic, and
tighten the capture/read substrate everything downstream rests on. A live
gap already bit us — a Claude Code agent that wanted to log work over MCP
had no documented way to connect, so it fell back to `brag add`. The
front door is unmarked; this wave marks it.

This project starts with the two lowest-risk, highest-fit pieces (MCP
ergonomics + accumulated polish) so the wave delivers value immediately
and the deeper trust/completeness work (signed provenance, corpus-as-
memory, capture completeness, benchmark) lands on a clean base.

## Success Criteria

- From a fresh MCP client, an agent can connect on the first try using
  only documented config, log a win that appears in `brag list`, and have
  `project` set correctly — no CLI fallback.
- The MCP tool contract (every param, type, return shape) and the
  agent-facing gotchas (no cwd project auto-fill; provenance stamping)
  are documented well enough that an agent needs no source-diving.
- The unregistered-project gap is closed: an entry's `project` can be
  ensured/registered so downstream consumers (`standup`) map entries→repos
  reliably.
- Accumulated PROJ-004 debt is cleared: the duplicated calendar-window
  upper-bound filter lives in storage (`ListFilter.Until`), and a
  lightweight `brag spark` pulse exists for a quick read.
- Everything ships local-first, no network in the binary, no new CGO.

## Scope

### In scope
- **STAGE-015 — MCP first-class for agents:** `brag mcp install`
  (idempotent client-config merge), closing the unregistered-project gap
  (`brag project ensure` / auto-register), and the MCP + "For AI agents"
  documentation.
- **STAGE-016 — v0.4.x polish:** `ListFilter.Until` storage promotion,
  `brag spark` sparkline pulse, and tier-1 micro-fixes (milestone `type`,
  cosmetic/doc corrections) surfaced by the post-v0.4.0 backlog scan.
- A v0.5.0 minor release cut (new commands → minor, not a patch).

### Explicitly out of scope
- Team federation, multi-user, any network transport, hosted tiers,
  external connectors inside the `brag` binary (permissible only as
  out-of-process adapters — the local-first line holds).
- Signed/attestable provenance, corpus-as-agent-memory (read-side MCP
  resources), capture completeness (inbox / git-import / evidence links),
  agent/model benchmark — these are the *deeper* agent-native stages that
  follow this opening; framed later in this project or a successor.
- A `sprint` schema field. Sprint is just a freeform tag
  (`--tag "sprint:<id>"`) today; at most a documented convention. No
  migration, no reserved namespace.

## Stage Plan

Format: `- [status] STAGE-ID — one-line summary`

- [x] STAGE-015 (shipped on 2026-07-10) — MCP first-class for agents
      (install command, unregistered-project gap, agent docs)
- [x] STAGE-016 (shipped on 2026-07-10) — v0.4.x polish (ListFilter.Until,
      brag spark) + a pre-release audit hardening pass (SPEC-060–066)

**Count:** 2 shipped / 0 active / 0 pending

## Dependencies

### Depends on
- Previous projects: PROJ-003 (agent-native spine — MCP server, plugin,
  provenance) and PROJ-004 (story surface — `internal/spark`,
  calendar-window infra, DEC-014 envelope) shipped the substrate this
  wave extends.

### Enables
- Future stages of this project (and any successor): signed provenance,
  corpus-as-agent-memory, capture completeness, and agent/model benchmark
  all rest on an ergonomic MCP path + a clean, complete substrate.

## Project-Level Reflection

*Shipped 2026-07-10 as v0.5.0.*

- **Did we deliver the outcome in "What This Project Is"?** Yes for the wave's
  **opening**, which is what this project was scoped to (all five Success
  Criteria met). The MCP path is first-class (agents self-register + log
  correctly-mapped wins with a documented contract), the unregistered-project
  gap is closed, and PROJ-004's debt is cleared (`ListFilter.Until` + `brag
  spark`). The **deeper** agent-native depth named in the vision paragraph
  (corpus-as-memory, signed provenance, capture completeness, benchmark) was
  always "explicitly out of scope / a successor" — it moves to PROJ-006.
- **How many stages did it actually take?** 2 (STAGE-015 MCP first-class,
  STAGE-016 v0.4.x polish) — matched the plan. STAGE-016 grew a pre-release
  hardening pass (6 fixes) beyond its two planned pieces.
- **What changed between starting and shipping?** The scope held (the opening),
  but a **user-requested pre-release bug audit** turned STAGE-016 into a
  hardening wave — a 4-agent whole-codebase sweep caught two pre-existing HIGH
  bugs the batch's own MCP-first-class work *amplified* (SQLite concurrent-access
  failures on the MCP path; silent `tag rename` corruption), fixed before the
  cut. That audit materially de-risked the release.
- **Lessons that should update AGENTS.md, templates, or constraints?** (candidates,
  not yet codified — hold to the §12 N=2/N=3 bar)
  - Cut the release BEFORE closing its stage — STAGE-016 was marked shipped a
    step early and had to reopen for SPEC-067 (the cut is the stage's closing
    action, per SPEC-054/STAGE-013).
  - A pre-release adversarial audit of the *whole* codebase (not just the diff)
    earns its keep when a release amplifies pre-existing surfaces — worth a
    release-cut pre-flight line.
  - Orchestration: writing sub-agents must run in isolated worktrees, and never
    `git add -A` with live worktrees / uncommitted WIP present (two accidental-
    capture incidents this project).
- **What did we defer to the next project (PROJ-006)?**
  - Corpus-as-agent-memory (read-side MCP resources) — the synthesis's #1.
  - Signed / attestable provenance — subsumes the deferred tag-forgery audit gap.
  - Capture completeness (inbox → git-import → evidence-links); agent/model
    benchmark.
  - The audit LOW/NIT backlog (mcp_install atomic-write, negative-`limit` MCP
    parity, `search -foo`, `brag project new` uncapped, the `spark` same-second
    edge, and the rest — see the STAGE-016 record).
