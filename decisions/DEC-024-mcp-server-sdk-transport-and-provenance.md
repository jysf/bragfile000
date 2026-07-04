---
# Maps to ContextCore insight.* semantic conventions.

insight:
  id: DEC-024
  type: decision
  confidence: 0.85                    # honest: the SDK/transport/subcommand
                                       # choices are high-confidence (a clean
                                       # §12(b) pre-flight retired the fallback);
                                       # the residual soft spots are the
                                       # provenance auto-stamp fallback and the
                                       # dependency weight — see Validation.
  audience:
    - developer
    - agent

agent:
  id: claude-opus-4-8
  session_id: null

# Decisions are repo-level, but it's useful to track which project
# caused them to be emitted.
project:
  id: PROJ-003
repo:
  id: bragfile

created_at: 2026-07-04
supersedes: null
superseded_by: null

tags:
  - mcp
  - transport
  - dependency
  - provenance
  - stdout-stderr-spine
  - agent-native
---

# DEC-024: `brag mcp serve` — official Go MCP SDK, stdio subcommand, provenance via reserved tags

## Decision

`brag mcp serve` is a **subcommand of the existing `brag` binary** (a parent
`mcp` cobra command with a `serve` child) that runs a **local stdio MCP
server** built on the **official Go MCP SDK
`github.com/modelcontextprotocol/go-sdk` (v1.6.1)** — the wave's one new
top-level Go dependency, gated by this DEC per
`no-new-top-level-deps-without-decision`. The server exposes four typed tools
— `brag_add`, `brag_list`, `brag_search`, `brag_stats` — as thin wrappers over
the existing `*storage.Store`; SQL stays in `internal/storage`. There is **no
network transport** (stdio only) and **no schema migration**.

Two sub-decisions ride with it:

1. **Transport/stdout purity.** The stdio transport owns `os.Stdout` for MCP
   protocol frames; nothing human-facing may touch that stream. This is the
   `stdout-is-for-data-stderr-is-for-humans` blocking constraint generalized to
   a new transport (§9 split-buffer). The SDK's default logger is
   `slog.DiscardHandler` (verified at design), so the SDK does not fight this;
   the server code and Store must likewise never write to `os.Stdout`.

2. **Provenance is captured as reserved-namespace tags, via explicit tool
   params with a transport-derived `agent` fallback.** The MCP `brag_add` tool
   accepts optional `agent` and `model` params and stamps them as reserved tags
   `agent:<name>` / `model:<id>` (lowercase, whitespace→`-`, commas stripped),
   appended to the caller's `tags` and canonicalized by the existing Store tags
   path (DEC-015) — **zero schema change**. `agent` falls back to the MCP
   client's `clientInfo.Name` when the param is omitted (auto-stamp when
   possible); `model` is **explicit-param only**, because the MCP transport does
   not carry a model identity (verified at design — see Context). First-class
   `agent`/`model` columns are explicitly deferred ("later, if earned").

The **hand-rolled stdio JSON-RPC fallback is retired**, not held: the
time-boxed SDK eval (§12(b)) was clean on every axis (below), so re-implementing
framing + schema + validation by hand would be owned code for no benefit.

## Context

STAGE-009's surfaced design question (a) named three coupled choices — SDK vs
hand-rolled loop, the transport, and subcommand-vs-separate-binary — and
question (e) asked whether provenance is auto-stamped from caller context or
taken as explicit params, *"depending on what identity the chosen transport
actually exposes."* Both were to be **settled at design with a DEC**, and the
transport contract was to be **pre-flighted against a real MCP client before
locking** (§12(b), the discipline that caught the goreleaser key-rename at
SPEC-023 and the cobra bash-marker at SPEC-024).

A design-time pre-flight stood the **official SDK v1.6.1** up with the four
tools as typed tools over `mcp.NewInMemoryTransports()`, driven by a real
`mcp.Client`. Findings (all green):

- **Round-trip works.** `tools/list` returns exactly the four tools with
  SDK-inferred 2020-12 JSON schemas (from `jsonschema` struct tags:
  `required:["title"]`, `additionalProperties:false`); all four `tools/call`
  round-trip.
- **Validation is automatic.** `brag_add` with no `title` returns
  `IsError=true` *before* the handler runs (schema-validated by the SDK).
- **CLI byte-parity is achievable by construction.** A tool with `Out=any`
  returning an explicit `TextContent` block carrying the exact
  `brag <cmd> --format json` bytes round-trips byte-identically to the client
  (`PARITY OK`). So the four tools mirror the existing DEC-011 (list/search) and
  DEC-014 (stats) contracts by reusing `internal/export` — no new output shape.
- **Provenance identity, settled (question e).** The handler reads
  `req.Session.InitializeParams().ClientInfo.Name` — the client *application*
  name (e.g. `claude-code`) — so `agent:` **can** be auto-filled from the
  transport. But the SDK's `Implementation` struct carries only
  `Name/Title/Version/WebsiteURL/Icons` — **no model** — so `model:` **cannot**
  come from the transport and must be an explicit param. This is why the design
  is *explicit params + `agent` fallback*, not pure auto-stamp: the model
  simply is not in the protocol.
- **stdout stays clean.** Default logger discards; no stray bytes on stdout/
  stderr during the full round-trip.
- **Pure Go.** `CGO_ENABLED=0 go build` of the server is clean; no `import "C"`
  in the SDK tree (`no-cgo` held).

## Alternatives Considered

- **Option A: hand-rolled stdio JSON-RPC loop (no SDK).**
  - What it is: implement the MCP `initialize`/`tools/list`/`tools/call`
    framing, JSON-Schema emission, and input validation by hand.
  - Why rejected: the SDK pre-flight was clean, so hand-rolling buys nothing and
    costs ~200–400 lines of owned protocol code plus re-implemented schema
    inference and validation — surface that drifts from the evolving MCP spec.
    Retained only as the fallback the pre-flight was designed to trigger; it did
    not trigger.

- **Option B: a separate `brag-mcp` binary.**
  - What it is: ship the server as its own executable, released alongside `brag`.
  - Why rejected: a second binary means a second install path, a second
    goreleaser artifact + Homebrew wiring, and duplicated `Store`/config
    resolution. A subcommand reuses one binary, one install, one
    `config.ResolveDBPath`, and one `storage.Open` — the recommendation the
    stage carried, confirmed cheap.

- **Option C: promote provenance to first-class `agent`/`model` columns now.**
  - What it is: extend the DEC-011 envelope (9→11 keys) + a migration.
  - Why rejected: the STAGE-009 core is migration-free, and this is the same
    accepted-debt→normalize path tags took (DEC-004→DEC-015): don't normalize
    until a real query need or taxonomy pollution appears. Reserved tags satisfy
    `brag list --tag model:<id>` and `brag tags` counting today with zero schema
    change. Promotion is the "later, if earned" step with an explicit revisit
    trigger (below).

- **Option D: pure auto-stamp provenance from the MCP session (no params).**
  - What it is: derive both `agent` and `model` from the transport, so they
    "can't be forgotten."
  - Why rejected: impossible for `model` — the MCP transport does not expose a
    model identity (verified). Auto-stamp-only would silently drop the model on
    every entry. Explicit params with an `agent` fallback captures both when
    available and degrades honestly when not.

- **Option E (chosen): official SDK + `brag mcp serve` subcommand + stdio +
  reserved-tag provenance via explicit params with an `agent` fallback.**
  - Why selected: the pre-flight retired every risk; one binary; CLI-parity by
    reusing `internal/export`; provenance rides DEC-015 with no migration; and
    the whole thing is pure-Go and stdout-clean.

## Consequences

- **Positive:** non-shell agents get a typed local write/read surface with no
  network boundary; the four tools honour the exact CLI JSON contracts
  (DEC-011/DEC-014) by construction; provenance is queryable *today*
  (`brag list --tag model:<id>`, `brag tags`) with zero schema change; the SDK
  gives schema inference, input validation, and MCP-spec conformance for free.
- **Negative:** a new top-level dependency (`modelcontextprotocol/go-sdk`) with
  a transitive footprint of ~6 indirect modules (`google/jsonschema-go`,
  `segmentio/asm`, `segmentio/encoding`, `yosida95/uritemplate/v3`,
  `golang.org/x/oauth2`, `golang.org/x/sys`). Notably the SDK bundles its
  HTTP/SSE/streamable transports (hence `x/oauth2` + segmentio) even though
  `brag` uses **stdio only** — dead weight in the dependency graph, though it
  cross-compiles pure-Go and adds no runtime cost to the stdio path.
  Dependencies are forever (the constraint's rationale); this DEC is the
  deliberate acceptance.
- **Negative:** the DEC-010 search tokenization is duplicated in
  `internal/mcpserver` (the existing `buildFTS5Query` is unexported in
  `internal/cli`); this is now a *second* consumer of that transform, which by
  the repo's own DEC-004→DEC-015 philosophy argues for extraction into a shared
  helper — deferred to keep this spec at M, with a parity test guarding drift.
- **Neutral:** the MCP `brag_add` deliberately does **not** emit the SPEC-039
  milestone line (its stdout is the protocol stream, a different spine at this
  transport). `brag mcp serve` is additive — no existing command or test asserts
  a subcommand count that it bumps.

## Validation

Right if: the `internal/mcpserver` in-memory-transport tests (the §12(b)
conformance harness, green at design) hold — four tools listed, each round-trips,
`brag_add` requires `title`, list/search/stats are byte-identical to the CLI
`--format json` on the same rows, provenance tags are stamped and filterable,
and `agent` auto-fills from `clientInfo.Name`; the stdout-purity test proves no
handler/Store path writes to `os.Stdout`; and `go.mod` gains exactly the one
direct dependency this DEC gates.

Revisit if: (a) provenance filtering/reporting becomes a real ask, OR
`agent:`/`model:` tags visibly pollute the `brag tags` taxonomy → promote to
first-class columns (Option C, carries a migration + DEC-011 envelope
extension); (b) a networked/multi-user MCP mode is genuinely needed → a new DEC
picks a transport and confronts the WAL + busy-timeout concurrency question the
core defers; (c) a third consumer of the DEC-010 search transform appears →
extract the shared query builder; (d) the SDK's dependency weight or API churn
becomes painful enough to justify the retired hand-rolled loop (Option A).

## References

- Related specs: SPEC-040 (emits + implements this DEC — the `brag mcp serve`
  server + provenance), SPEC-041 (packages the server into the Claude Code
  plugin and documents the reserved-namespace convention in the shipped assets),
  SPEC-039 (the CLI-side milestone — the mirror stdout/stderr spine this DEC
  faces at a new transport)
- Related decisions: DEC-015 (polymorphic tags normalization — provenance rides
  the taggings join), DEC-011 (JSON entry shape — `brag_list`/`brag_search`/
  `brag_add` output parity), DEC-012 (`brag add --json` schema — the `brag_add`
  required-title / server-owned-field contract this mirrors), DEC-014
  (rule-based output envelope — `brag_stats` output parity), DEC-010 (search
  query syntax — the tokenization `brag_search` reuses), DEC-006 (cobra —
  `mcp`/`serve` are cobra subcommands), DEC-004→DEC-015 (the accepted-debt→
  normalize path provenance follows)
- Related constraints: `no-new-top-level-deps-without-decision` (warning — this
  DEC is the gate), `stdout-is-for-data-stderr-is-for-humans` (blocking —
  generalized to the stdio transport), `no-sql-in-cli-layer` (blocking — the
  server wraps `Store`, no SQL in `internal/cli` or `internal/mcpserver`),
  `no-cgo` (blocking — the SDK is pure Go), `errors-wrap-with-context`
- External: `github.com/modelcontextprotocol/go-sdk` v1.6.1 (the official MCP
  Go SDK); Model Context Protocol spec (stdio transport; `tools/list`,
  `tools/call`, `initialize`/`clientInfo`)
- Discussions: STAGE-009 Design Notes surfaced questions (a) + (e) and "The
  stdout/stderr spine at a new transport"; PROJ-003 brief Dependencies ("one new
  Go dependency likely at SPEC-040 design")
