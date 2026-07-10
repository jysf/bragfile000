---
# Maps to ContextCore insight.* semantic conventions.

insight:
  id: DEC-034
  type: decision
  confidence: 0.80                    # honest: the merge algorithm and the
                                       # claude-code/cursor paths are high-
                                       # confidence (verified against official
                                       # docs + a design-time byte-exact
                                       # pre-flight); the soft spots are the
                                       # semantic-not-byte-for-byte rewrite of
                                       # the big managed ~/.claude.json file and
                                       # the cross-OS Claude Desktop path — see
                                       # Validation.
  audience:
    - developer
    - agent

agent:
  id: claude-opus-4-8
  session_id: null

# Decisions are repo-level, but it's useful to track which project
# caused them to be emitted.
project:
  id: PROJ-005
repo:
  id: bragfile

created_at: 2026-07-10
supersedes: null
superseded_by: null

tags:
  - mcp
  - install
  - config-merge
  - idempotency
  - agent-native
  - stdout-stderr-spine
---

# DEC-034: `brag mcp install` — idempotent, no-clobber client-config merge

## Decision

`brag mcp install [--client claude-code|claude-desktop|cursor]
[--scope user|project] [--dir PATH] [--dry-run]` writes or merges the MCP
server block `{"command":"brag","args":["mcp","serve"]}` under the
`mcpServers` object of the target client's config file, so `brag mcp serve`
(DEC-024) is registered as a server named `brag`. Four sub-decisions ride
with it:

1. **Supported clients + default.** Three clients are supported:
   `claude-code` (the **default**), `claude-desktop`, `cursor`. The default
   is stated explicitly in the flag definition (flag-default-explicitness,
   AGENTS.md §12). The registered server block is **exactly the shape the
   shipped plugin already uses** (`plugin/.mcp.json`,
   `plugin/.claude-plugin/plugin.json`): `{"command":"brag","args":["mcp",
   "serve"]}` — no `type` field (all three clients infer stdio from the
   presence of `command`).

2. **Target path per client × scope, resolved flag → default (DEC-003
   discipline), validated at design time (§12(b)).** The default scope is
   **`project`** (the checked-in, shareable `.mcp.json`, which is the
   agent-native "register brag in this repo" path). Resolved paths (verified
   against official docs, 2026-07-10):

   | client         | scope   | path |
   |----------------|---------|------|
   | claude-code    | project | `<dir>/.mcp.json` |
   | claude-code    | user    | `~/.claude.json` (top-level `mcpServers`) |
   | cursor         | project | `<dir>/.cursor/mcp.json` |
   | cursor         | user    | `~/.cursor/mcp.json` |
   | claude-desktop | user    | macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`; Windows: `%APPDATA%\Claude\claude_desktop_config.json` |

   `<dir>` is the `--dir` value, defaulting to the current working directory
   (via the injectable `getCwd` seam). `--dir` is only meaningful for
   `project` scope. **Unsupported combinations are `UserError`s, never silent
   fallbacks:** `claude-desktop --scope project` (Desktop is a single-user
   GUI app with no project config), `--dir` with `--scope user`, an unknown
   `--client`, an unknown `--scope`, and `claude-desktop` on an OS whose
   Application-Support location is unknown (anything other than darwin /
   windows).

3. **The merge algorithm — idempotent, no-clobber, semantic-preserving.**
   Read the target bytes if the file exists → parse the top level into
   `map[string]json.RawMessage` (unknown top-level keys survive as raw) →
   parse `mcpServers` into `map[string]json.RawMessage` (every other server
   survives as raw) → **set only the `brag` key** to the canonical block →
   re-marshal with `json.MarshalIndent(v, "", "  ")` + a trailing newline.
   An absent/empty file produces a file containing just the `mcpServers.brag`
   block. A present-but-different `brag` block is **overwritten** (the brag
   key is fully replaced, not deep-merged). Because the block is rebuilt from
   a fixed Go struct (`command` then `args`) and the whole file is marshaled
   deterministically, a second run produces **byte-identical** output — the
   command detects this and reports a no-op (exit 0, no write).

   **Preservation is semantic, not byte-for-byte.** `encoding/json` sorts map
   keys and normalizes whitespace, so unrelated top-level keys and other
   servers are all preserved with their values intact but the file's key
   order and formatting are canonicalized on rewrite. This is the accepted
   trade (a formatting-preserving JSON editor would need a new dependency,
   which `no-new-top-level-deps-without-decision` discourages for a
   config-writer); the load-bearing guarantee is that **no other server and
   no unrelated key is lost or altered in value.**

4. **Output contract (`stdout-is-for-data-stderr-is-for-humans`).**
   `--dry-run` writes **nothing**; it prints the exact JSON that *would* be
   written to **stdout** (the machine-consumable artifact) and a
   `Would write to <path>:` annotation to **stderr**. A real write prints a
   human `Registered brag MCP server in <path>` (or, on a no-op,
   `brag MCP server already registered in <path> (no changes)`) to
   **stderr**, leaving **stdout empty** (the side effect is the file; there
   is no scriptable datum to emit). Exit 0 on success and on the idempotent
   no-op; the `UserError`s in sub-decision 2 exit 1 with stdout empty.

The command needs **no storage and no SQL** — it imports neither
`internal/storage` nor `database/sql`, so `no-sql-in-cli-layer` is
structurally satisfied.

## Context

STAGE-015 ("MCP first-class for agents") observed a live Claude Code agent
fall back to `brag add -p standup …` on the CLI because nothing told it how
to connect the already-shipped MCP server (DEC-024/SPEC-040). The capability
exists; only the front door is missing. `brag mcp install` is that door: one
command that registers the server in a client's config.

The load-bearing property is **idempotency + no-clobber** — a user (or agent)
must be able to run `install` repeatedly and against a config that already
contains other MCP servers, without ever losing or corrupting them. That
makes the merge scheme the decision worth recording, and makes the emitted
JSON a fixed-shape artifact suited to the literal-artifact-as-spec discipline
(AGENTS.md §12): the exact bytes were pre-flighted at design through a scratch
Go program running the real merge (`json.Valid` + byte-equality across two
runs), so the spec embeds known-good literals rather than described shapes.

The config paths are the external-reality surface (§12(b)): each was
confirmed against the client's official documentation at design
(2026-07-10) — Claude Code's project `.mcp.json` and user `~/.claude.json`
(both keyed `mcpServers`), Claude Desktop's
`claude_desktop_config.json` under the OS Application-Support dir, and
Cursor's `.cursor/mcp.json` (project) / `~/.cursor/mcp.json` (user).

## Alternatives Considered

- **Merge scheme: error when a `brag` block already exists (require manual
  edit).** Rejected: it defeats idempotency and re-runnability. `install`
  must be safe to run after the args ever change (e.g. a future `mcp serve`
  flag) and after re-installation; erroring forces the user into the exact
  hand-editing the command exists to avoid.

- **Merge scheme: deep-merge into the existing `brag` block (keep its extra
  keys).** Rejected: `brag` owns its own server block entirely. Deep-merging
  would let a stale `args` entry (or a removed argument) linger; a full
  replace guarantees the block always matches the current canonical shape.
  Other servers are a different owner and ARE preserved — the no-clobber
  guarantee is about *other* keys, not brag's own.

- **Formatting-preserving JSON edit (sjson-style / a JSON-with-comments
  editor) to keep the target file byte-stable outside the brag block.**
  Rejected for v1: it needs a new top-level dependency
  (`no-new-top-level-deps-without-decision`) and hand-rolling a
  format-preserving editor is far more code than the semantic-preserving
  `encoding/json` round-trip. The trade — canonicalized key order/indent on
  rewrite — is acceptable because the values survive intact. Revisit if the
  `~/.claude.json` churn (below) proves painful.

- **Default scope = `user`.** Rejected: for `claude-code` that would write to
  the large, actively-managed `~/.claude.json` on the common path, maximizing
  the churn from sub-decision 3's canonicalization. Project `.mcp.json` is
  the smaller, checked-in, shareable artifact and the better default for the
  "register brag in this repo so agents here can connect" story.

- **Silent fallback for unsupported client×scope combos** (e.g. treat
  `claude-desktop --scope project` as `user`). Rejected: hides a user
  mistake. Every incoherent combo is a `UserError` naming the problem — the
  repo's established "error clearly on incoherent combos" ethos (cf. DEC-032
  `--previous`+`--since`).

- **A separate `brag-install-mcp` script or a plugin-only path.** Rejected:
  the plugin (SPEC-041) already registers the server for Claude Code *plugin*
  users, but the many agents/users not on the plugin need a first-class,
  client-agnostic command in the one `brag` binary — one binary, one install,
  reusing the shape the plugin already encodes.

## Consequences

- **Positive:** an agent or human registers `brag` in any of the three common
  clients with one idempotent command; re-running is always safe; existing
  MCP servers are never clobbered; `--dry-run` gives a copy-pasteable,
  scriptable JSON artifact on stdout and the resolved path on stderr; the
  command carries no storage/SQL surface, so it is trivially hermetic to
  test.
- **Negative:** on rewrite, the target file's unrelated top-level keys and
  other servers are re-serialized with sorted keys and normalized indent —
  semantically identical but not byte-identical to the user's prior
  formatting. For the small `.mcp.json` / `.cursor/mcp.json` files this is
  negligible; for the large managed `~/.claude.json` (`claude-code --scope
  user`) it can produce a sizable diff. This is why the **default scope is
  `project`**, keeping the churn off the common path.
- **Negative:** Claude Desktop's config path is OS-specific and the app is
  effectively macOS/Windows only; other OSes are a `UserError` rather than a
  guess. Linux Claude Desktop, if it matters later, is an additive path
  entry.
- **Neutral:** `install` is additive — no existing command, test, or
  subcommand count changes except `brag mcp`'s child count (from 1 → 2), and
  `mcp_test.go` asserts help *contains* `serve`, not an exact child list.

## Validation

Right if: the pure `mergeMCPConfig` tests hold (absent file → the canonical
block; a pre-existing OTHER server + unrelated top-level key both survive with
values intact; a second merge is byte-identical; a stale `brag` block is
overwritten; malformed input errors with context); the CLI tests hold
(`install` writes the byte-exact literal, is a no-op on re-run while
preserving a pre-seeded unrelated server, `--dry-run` writes nothing while
printing JSON to stdout + path to stderr, and every unsupported combo is an
`errors.Is(err, ErrUser)` with empty stdout); and the path-resolution table
matches the DEC's table under a stubbed home dir.

Revisit if: (a) the `~/.claude.json` canonicalization churn is reported as
painful → adopt a format-preserving JSON editor (with its own dep DEC); (b) a
client changes its config path or key (the paths were verified 2026-07-10 and
are version-sensitive) → update the resolver + this table; (c) Linux Claude
Desktop or a fourth client (Windsurf, VS Code, Zed …) is requested → add a
path-table row (additive); (d) a client requires an explicit `"type":"stdio"`
field → add it to the canonical block (all three infer it today).

## References

- Related specs: SPEC-055 (emits + implements this DEC — the `brag mcp
  install` command), SPEC-040/SPEC-041 (the `brag mcp serve` server + the
  plugin artifacts whose `{command,args}` block this reuses), SPEC-056
  (closes the unregistered-project gap this install path exposes), SPEC-057
  (the agent-facing docs that will cite the per-client snippets)
- Related decisions: DEC-024 (the MCP server + stdio transport this registers;
  the `stdout-is-for-data` spine at a new surface), DEC-003 (flag → env →
  default resolution discipline mirrored for the target path), DEC-006 (cobra
  subcommands — `mcp install` is a child of `mcp`)
- Related constraints: `stdout-is-for-data-stderr-is-for-humans` (blocking —
  dry-run JSON to stdout, human messages to stderr), `no-sql-in-cli-layer`
  (blocking — install imports no storage/SQL), `errors-wrap-with-context`
  (parse/IO errors wrapped), `no-secrets-in-code`, `no-cgo`,
  `no-new-top-level-deps-without-decision` (the merge uses stdlib
  `encoding/json` only — no new dep)
- External (verified 2026-07-10): Claude Code MCP config
  (`.mcp.json` project scope; `~/.claude.json` user scope, top-level
  `mcpServers`); Claude Desktop `claude_desktop_config.json` under
  `~/Library/Application Support/Claude/` (macOS) / `%APPDATA%\Claude\`
  (Windows); Cursor `.cursor/mcp.json` (project) / `~/.cursor/mcp.json`
  (user), all keyed `mcpServers`
