---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-066
  type: bug                        # epic | story | task | bug | chore
  cycle: verify
  blocked: false
  priority: medium
  complexity: S                    # S | M | L  (L means split it)

project:
  id: PROJ-005
  stage: STAGE-016
repo:
  id: bragfile

agents:
  architect: claude-opus-4-8
  implementer: claude-opus-4-8     # usually same Claude, different session
  created_at: 2026-07-10

references:
  decisions:
    - DEC-024
  constraints:
    - errors-wrap-with-context
    - stdout-is-for-data-stderr-is-for-humans
  related_specs: []
---

# SPEC-066: mcp serve clean shutdown exit code

## Context

A pre-release audit flagged this as MEDIUM #4 (part of STAGE-016 polish
under PROJ-005). `internal/cli/mcp.go` (`runMCPServe`) returned
`srv.Run(cmd.Context(), &mcp.StdioTransport{})` verbatim.

On a NORMAL client shutdown — the client closes stdin with a request still
in flight — the go-sdk's `Server.Run` does not return nil. It returns the
internal `jsonrpc2.ErrServerClosing` sentinel wrapped with the underlying
`io.EOF`, so `err.Error()` is `"server is closing: EOF"`. `cmd/brag/main.go`
maps every non-`ErrUser` error to **exit 2** and prints
`brag: server is closing: EOF`. A supervising MCP client reasonably logs
that nonzero exit + line as a crash of the headline `brag mcp serve`
feature.

This is a shutdown race, not a real failure: a clean close with **no**
in-flight request already drains to `nil` (RC 0). Only when a request is in
flight does the closing sentinel surface.

## Goal

Treat a normal MCP transport shutdown as a clean exit (RC 0): if
`srv.Run` returns an error that is `io.EOF`, `context.Canceled`, or the
SDK's `ErrServerClosing`, exit 0. Propagate every other error unchanged so
a genuine serve failure still exits nonzero.

## Inputs

- **Files to read:** `internal/cli/mcp.go` — `runMCPServe`;
  `cmd/brag/main.go` — the exit-code mapping (non-`ErrUser` → exit 2).
- **SDK source (module cache):**
  `github.com/modelcontextprotocol/go-sdk@v1.6.1/mcp/server.go`
  (`Server.Run` → `ServerSession.Wait` → `Connection.wait`),
  `internal/jsonrpc2/conn.go` (`shuttingDown`, `wait`), and
  `internal/jsonrpc2/wire.go` (`ErrServerClosing = NewError(-32004, "server is closing")`,
  `WireError.Is` compares by `Code`). The public `jsonrpc` package
  re-exports the wire error type as `jsonrpc.Error = jsonrpc2.WireError`.

## Outputs

- **Files modified:** `internal/cli/mcp.go` — add `isCleanShutdown(err) bool`
  and a thin `serve(ctx, srv, transport)` wrapper; `runMCPServe` calls
  `serve` instead of returning `srv.Run` verbatim.
  `internal/cli/mcp_test.go` — new fail-first tests.
- **New exports:** none (helpers are package-private).
- **Database changes:** none.

## Acceptance Criteria

- [x] A client that closes the transport with a request in flight makes
      `serve` return `nil` (would map to RC 0), not the
      `"server is closing: EOF"` error.
- [x] `isCleanShutdown` returns true for `nil`, bare/wrapped `io.EOF`,
      `context.Canceled` (bare/wrapped), and the SDK's wrapped
      `ErrServerClosing` shape.
- [x] `isCleanShutdown` returns false for a generic error and for a
      wrapped `open store: ...` failure (real serve failures stay nonzero).
- [x] A propagated genuine failure is wrapped with context (`errors.Is`
      still works through it).
- [x] Full gate set passes.

## Failing Tests

- **`internal/cli/mcp_test.go`**
  - `TestIsCleanShutdown` — table: `nil`, `io.EOF`, wrapped `io.EOF`,
    `context.Canceled`, wrapped `context.Canceled`, the SDK
    `fmt.Errorf("%w: %v", &jsonrpc.Error{Code:-32004}, io.EOF)` shape → all
    true; generic error and wrapped `open store` failure → false.
  - `TestServe_InFlightClientCloseIsCleanExit` — end-to-end over an
    `mcp.IOTransport` pipe with a deliberately blocked tool handler:
    handshake, call the blocking tool, wait until the handler is entered,
    then close the read side (in-flight EOF). Asserts `serve` returns nil.
    Deterministically fails against the verbatim `srv.Run` (returns
    `server is closing: EOF`); passes with the fix.

## Implementation Context

### Decisions that apply

- `DEC-024` — MCP server SDK/transport choice (go-sdk, local stdio
  `StdioTransport`). This fix hardens that transport's shutdown contract at
  the CLI boundary; it does not revisit the SDK/transport decision.

### Constraints that apply

- `errors-wrap-with-context` — the clean-shutdown sentinels must be
  recognized via `errors.Is` even when the SDK wraps them; a propagated
  genuine failure is wrapped `fmt.Errorf("run mcp server: %w", err)` so it
  stays legible and `errors.Is`-transparent.
- `stdout-is-for-data-stderr-is-for-humans` — unchanged: `mcp serve` still
  writes only protocol frames to stdout; nothing human-facing is added
  there.

### Out of scope (for this spec specifically)

- Changing `cmd/brag/main.go`'s exit-code mapping (the fix returns nil so
  the existing mapping already yields RC 0).
- Any change to `internal/mcpserver` behavior or the tool set.

## Notes for the Implementer

The `ErrServerClosing` sentinel lives in the SDK's **internal**
`jsonrpc2` package and cannot be imported. It is a `*jsonrpc2.WireError`
(code `-32004`), and the public `jsonrpc` package re-exports that type as
`jsonrpc.Error`. `WireError.Is` compares by `Code`, so a local
`&jsonrpc.Error{Code: -32004}` sentinel matches the wrapped SDK error via
`errors.Is`. Note the SDK appends the underlying `io.EOF` with `%v` (not
`%w`), so `errors.Is(err, io.EOF)` is **false** for that shape — the code
match is what recognizes it.

---

## Build Completion

*Filled in at the end of the **build** cycle, before advancing to verify.*

- **Branch:** `fix/spec-066-mcp-clean-shutdown`
- **PR (if applicable):** see PR opened against `main`
- **All acceptance criteria met?** yes
- **New decisions emitted:**
  - none — the fix operationalizes DEC-024's transport choice; no new
    decision needed.
- **Deviations from spec:**
  - none.
- **Follow-up work identified:**
  - none.

### Build-phase reflection (3 questions, short answers)

1. **What was unclear in the spec that slowed you down?**
   — The scaffold pointed at `ErrServerClosing` as "the SDK sentinel", but
   it turned out to live in an internal package (not importable). Reading
   the SDK source resolved it: match the public `jsonrpc.Error` by code.

2. **Was there a constraint or decision that should have been listed but wasn't?**
   — No. `errors-wrap-with-context` and DEC-024 covered it. I confirmed the
   exact returned shape empirically (a throwaway harness over `IOTransport`)
   before writing the matcher rather than guessing.

3. **If you did this task again, what would you do differently?**
   — Nothing material. Extracting a testable `serve()` seam up front let the
   integration test drive the real SDK shutdown deterministically (block the
   handler, wait for it to enter, then EOF) instead of racing.

---

## Reflection (Ship)

*Appended during the **ship** cycle. Outcome-focused reflection, distinct
from the process-focused build reflection above.*

1. **What would I do differently next time?**
   — <answer>

2. **Does any template, constraint, or decision need updating?**
   — <answer>

3. **Is there a follow-up spec I should write now before I forget?**
   — <answer>
