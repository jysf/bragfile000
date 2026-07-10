---
# Maps to ContextCore epic-level conventions.
# A Stage is a coherent chunk of work within a Project.
# It has a spec backlog and ships as a unit when the backlog is done.

stage:
  id: STAGE-015
  status: active
  priority: high
  target_complete: null

project:
  id: PROJ-005
repo:
  id: bragfile

created_at: 2026-07-10
shipped_at: null
---

# STAGE-015: MCP first-class for agents

## What This Stage Is

Makes `brag mcp serve` a first-class, self-serve path for AI agents
instead of an undocumented one that agents fall back off of. When this
stage ships, an agent (or its human) can register the brag MCP server
with a common client in one command, knows the exact tool contract from
docs, and lands wins with a correctly-set `project` — closing the gap
that today sends agents to `brag add` on the CLI.

## Why Now

The MCP server already exposes `brag_add`/`brag_list`/`brag_search`/
`brag_stats` over local stdio, but there is no documented, ergonomic way
to register it — so agents can't discover the path and fall back to the
CLI (observed live: a Claude Code agent logging for the `standup` project
used `brag add -p standup …` because nothing told it how to connect). The
capability exists; only the front door is missing. Marking that door is
the highest-leverage, lowest-risk opener for the agent-native wave, and it
unblocks every deeper agent-native feature (memory, provenance, benchmark)
that assumes agents actually use the MCP path.

## Success Criteria

- `brag mcp install` writes/merges the correct client config
  idempotently, never clobbering other MCP servers already present, with a
  `--dry-run` that prints the exact JSON + target path.
- From a fresh MCP client using only the documented config: connect on
  the first try, `brag_add` creates an entry visible in `brag list`, and
  `project` is set correctly.
- The unregistered-project gap is closed — an entry's `project` can be
  ensured/registered so `standup` and other consumers map entries→repos
  reliably.
- Docs give an agent the full tool contract and the gotchas (no cwd
  project auto-fill over MCP; provenance stamping; `--db` override) with
  no source-diving.

## Scope

### In scope
- `brag mcp install [--client claude-code|claude-desktop|cursor]
  [--scope user|project] [--dir PATH] [--dry-run]` — idempotent config
  merge (Claude Code `.mcp.json` / user-scope equivalent; Claude Desktop
  `claude_desktop_config.json` → `mcpServers`).
- Closing the unregistered-project gap: `brag project ensure <name>
  [--location PATH]` (or an auto-register/warn on `brag_add` for unknown
  projects) — decision to lock at spec design.
- MCP + "For AI agents" documentation: README section + a docs page —
  copy-paste registration snippet per client; the client-startup-reconnect
  note; full tool schemas (every param, type, required/optional, return
  shape); the `project`-not-auto-filled-over-MCP gotcha; provenance
  stamping (`agent:`/`model:`/`session:`/`cost:`/`tokens:`); the `--db`
  override; how to log a win + the impact-framing convention.

### Explicitly out of scope
- Corpus-as-agent-memory (read-side MCP resources), signed provenance,
  capture completeness, benchmark — later agent-native stages.
- Any network transport for the MCP server (stdio only, DEC-024).
- Auto-editing a *running* client's live session — docs note that MCP
  servers connect at client startup and a session must reconnect.

## Spec Backlog

Format: `- [status] SPEC-ID (cycle) — one-line summary`

- [x] SPEC-055 (shipped on 2026-07-10) — `brag mcp install`: idempotent
      client-config merge (DEC-034), `--dry-run`, never clobbers other
      `mcpServers`.
- [x] SPEC-057 (shipped on 2026-07-10) — close the unregistered-project
      gap: `brag project ensure <name> [--location PATH]`, an idempotent
      create-or-no-op storage primitive (`EnsureProject`/`EnsureLocation`)
      + documented the two soft-link facts (`project list` locations
      authoritative-but-incomplete; a project may have multiple
      locations). `brag_add` stays free-text — no silent auto-register
      (DEC-036). Full agent-facing MCP docs are SPEC-058.
- [ ] SPEC-058 (design) — MCP + "For AI agents" docs (README + docs page,
      full tool schemas, gotchas, impact-framing convention). Comes last,
      once install + project-ensure shapes are final.

**Count:** 2 shipped / 0 active / 1 pending

*(Numbering note: the "unregistered gap" work landed as SPEC-057 and the
docs work as SPEC-058 — the earlier backlog's SPEC-056 label was consumed
by the unrelated `ListFilter.Until` promotion. IDs are repo-global and
monotonic, §2.)*

## Design Notes

- **Idempotency + no-clobber is the load-bearing property** of `mcp
  install`: read the target file if present, merge only the `brag` key
  under `mcpServers`, preserve every other server and unrelated keys,
  write back. `--dry-run` prints the exact JSON that *would* be written +
  the resolved target path. This is a literal-artifact-shaped spec (the
  emitted JSON snippet is fixed); embed the exact snippet per client and
  diff at verify (AGENTS.md §12 literal-artifact-as-spec).
- **Config path resolution is the external-reality surface** — the
  correct target path per client/scope/OS must be validated at design
  time (§12(b) pre-flight): confirm the actual file names/locations
  (Claude Code `.mcp.json` project vs user scope; Claude Desktop
  `claude_desktop_config.json`) rather than assuming.
- **Unregistered-project gap:** entries store `project` as free text with
  no referential check, and MCP `brag_add` does NOT auto-fill `project`
  from cwd (unlike the CLI). `brag project list` reads the `projects`
  table (authoritative but incomplete); a project may have multiple
  on-disk `locations`. Whatever SPEC-056 lands must document both facts
  for downstream mappers (`standup`).
- The MCP tool contract already exists in `docs/api-contract.md`
  (§`brag mcp serve`); SPEC-057 lifts it into an agent-facing form, it
  does not invent it.

## Dependencies

### Depends on
- PROJ-003 (STAGE-009) — shipped the MCP server, plugin packaging, and
  reserved-tag provenance this stage makes ergonomic.

### Enables
- Every deeper agent-native stage (memory, signed provenance, benchmark)
  that assumes agents actually connect over MCP.
- Downstream integrators (`standup`) that map entries→repos.

## Stage-Level Reflection

*Filled in when status moves to shipped.*

- **Did we deliver the outcome in "What This Stage Is"?** <yes/no + notes>
- **How many specs did it actually take?** <number vs. plan>
- **What changed between starting and shipping?** <one sentence>
- **Lessons that should update AGENTS.md, templates, or constraints?**
  - <one-line updates>
- **Should any spec-level reflections be promoted to stage-level lessons?**
  - <one-line items>
