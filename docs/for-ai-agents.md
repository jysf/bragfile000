# Using brag from an AI agent (MCP)

`brag` ships a local, stdio-only Model Context Protocol (MCP) server so an AI
coding agent can capture and retrieve brag entries as typed tool calls instead
of shelling out to the CLI. This page is the full agent playbook: how to
register the server, the exact tool contract, and the gotchas that bite agents.

Everything here is shipped behavior. For *when* to propose a brag and the
approval loop (propose, wait for approval, then capture), see
[`BRAG.md`](../BRAG.md). For the full CLI reference, see
[`api-contract.md`](./api-contract.md).

## 1. Register the server

One command registers `brag mcp serve` in your client's config, idempotently
and without clobbering any other MCP server already present:

```
brag mcp install [--client claude-code|claude-desktop|cursor] [--scope user|project] [--dir PATH] [--dry-run]
```

Defaults: `--client claude-code`, `--scope project`, `--dir` = the current
directory. Re-running is always safe — a byte-identical result is reported as
a no-op. `--dry-run` prints the exact JSON to **stdout** and the target path
to **stderr**, writing nothing to disk (a good way to see where a client's
config lives).

The resolved config file per client and scope:

| client         | scope   | config file |
|----------------|---------|-------------|
| claude-code    | project | `<dir>/.mcp.json` |
| claude-code    | user    | `~/.claude.json` |
| cursor         | project | `<dir>/.cursor/mcp.json` |
| cursor         | user    | `~/.cursor/mcp.json` |
| claude-desktop | user    | macOS: `~/Library/Application Support/Claude/claude_desktop_config.json` |

(`claude-desktop` has no project scope. On non-macOS systems run
`brag mcp install --client claude-desktop --scope user --dry-run` to print the
exact path, or see [`api-contract.md`](./api-contract.md).)

**Not using `install`?** Add this block to your client's `mcpServers` map by
hand:

```json
{"mcpServers":{"brag":{"command":"brag","args":["mcp","serve"]}}}
```

The `brag` binary must be on your `PATH` — `brew install jysf/bragfile/bragfile`
or `go install ./cmd/brag`. The client launches the server as `brag mcp serve`.

(If you use Claude Code, the same tools also ship as a plugin —
`claude plugin marketplace add jysf/bragfile000` then
`claude plugin install brag@bragfile`. See [`BRAG.md`](../BRAG.md).)

## 2. Reconnect after installing

MCP servers **connect at client startup**. If your client is already running
when you install, it will not see the `brag` tools until you restart or
**reconnect** the session. A running session cannot pick the server up
mid-flight.

## 3. The tools

The server advertises exactly four typed tools over stdio. All read and write
the same `~/.bragfile/db.sqlite` the CLI uses (see §6 to change that).

### `brag_add` — capture an entry

| param         | type    | required | notes |
|---------------|---------|----------|-------|
| `title`       | string  | **required** | non-empty; ≤200 characters |
| `description` | string  | optional | ≤100000 characters |
| `tags`        | string  | optional | comma-joined string (DEC-004), NOT an array; ≤64 characters |
| `project`     | string  | optional | ≤64 characters — **read §4** |
| `type`        | string  | optional | ≤64 characters, e.g. `shipped`, `fixed`, `learned` |
| `impact`      | string  | optional | ≤256 characters — **read §7** |
| `agent`       | string  | optional | provenance; stamped `agent:<name>` (see §5) |
| `model`       | string  | optional | provenance; stamped `model:<id>` |
| `session`     | string  | optional | provenance; stamped `session:<id>` |
| `cost`        | string  | optional | provenance; stamped `cost:<n>` |
| `tokens`      | string  | optional | provenance; stamped `tokens:<n>` |

Returns the created entry as a single JSON object with the nine standard keys:
`id`, `title`, `description`, `tags`, `project`, `type`, `impact`,
`created_at`, `updated_at`. A missing or empty `title` is a **tool error**,
never a silent insert. Unlike the CLI `brag add`, `brag_add` does **not** emit
a milestone line.

### `brag_list` — list entries

| param     | type    | required | notes |
|-----------|---------|----------|-------|
| `tag`     | string  | optional | exact-match filter |
| `project` | string  | optional | exact-match filter |
| `type`    | string  | optional | exact-match filter |
| `limit`   | integer | optional | `0` = unlimited |

Returns a JSON array of entry objects, byte-identical to
`brag list --format json` on the same rows. There is **no `--since` filter**
over MCP (deferred); filter by time on the CLI if you need it.

### `brag_search` — full-text search

| param   | type    | required | notes |
|---------|---------|----------|-------|
| `query` | string  | **required** | FTS query, whitespace-tokenized and AND-joined (DEC-010) |
| `limit` | integer | optional | `0` = unlimited |

Returns a JSON array of entry objects (same shape as `brag_list`).

### `brag_stats` — lifetime stats

Takes **no parameters**. Returns the lifetime stats envelope, byte-identical
to `brag stats --format json` for the same corpus.

## 4. Gotcha: `project` is not auto-filled over MCP

On the CLI, `brag add` auto-fills `project` from your current directory when
you omit it (nearest registered project location). The MCP `brag_add` tool
**does not auto-fill `project`** — the server has no meaningful working
directory relative to you, the calling agent. If you omit `project`, the entry
lands **project-less**, and downstream consumers that map entries to repos
won't see it.

**Always pass `project` explicitly.** To make the name map cleanly for those
consumers, register it once with the CLI (idempotent — safe before every
capture):

```
brag project ensure standup
brag project ensure standup --location ~/code/standup
```

`brag project ensure <name>` creates the project if absent and is a no-op if
it already exists. See [`api-contract.md`](./api-contract.md) for its full
contract. Capture stays free text — bragfile never silently auto-registers an
unknown `project` for you (DEC-036).

## 5. Provenance stamping

`brag_add` records *who* and *what* produced an entry as reserved-namespace
tags. Each is appended after your own `tags` and canonicalized like any tag
(lowercased, whitespace runs → `-`, commas stripped) with **no schema change**
— so they filter and count like any other tag:

- `agent:<name>` — from the `agent` param. Falls back to the MCP client's
  `clientInfo.Name` when you omit it. Pass it explicitly for a deterministic
  value.
- `model:<id>` — from the `model` param. Explicit-only; the transport carries
  no model identity, so there is no fallback.
- `session:<id>` — from the `session` param. An opaque, stable per-session id
  used as a join key. No fallback (forward the id your hook surfaces).
- `cost:<n>` — from the `cost` param. A non-negative USD decimal, e.g.
  `cost:0.42`. A non-numeric or negative value is a **tool error**, never a
  coerced tag. bragfile never estimates it.
- `tokens:<n>` — from the `tokens` param. A non-negative integer, e.g.
  `tokens:18000`. Same validation as `cost`. bragfile never estimates it.

Omit any param and no tag is stamped. `agent:`/`model:` are the
**author-provenance** tags — `brag list --author agent` (and `--author human`)
classify on them; `session:`/`cost:`/`tokens:` are **seed metadata**, not
author-provenance, so an entry carrying only those still classifies as
`human`. Query any of them with the normal filters, e.g.
`brag list --tag model:claude-opus-4-8`. See DEC-024 and DEC-027.

## 6. Choosing the database

The server reads and writes the same database the CLI resolves, in this order
(DEC-003):

1. the `--db` flag — run the server as `brag --db PATH mcp serve`;
2. otherwise the `BRAGFILE_DB` environment variable, if set;
3. otherwise the default `~/.bragfile/db.sqlite`.

To point an agent at a scratch database, either register the server with a
`--db` flag in its `args`, or set `BRAGFILE_DB` in the client's environment
before it launches the server.

## 7. Logging a win

To capture a win, call `brag_add` with at least:

- `title` — a specific, action-verb headline;
- `type` — usually `"shipped"` (or `fixed`, `learned`, `documented`, …);
- `impact` — the concrete outcome (see below); and
- `project` — the name (see §4).

**Frame `impact` as the outcome, not the output.** State a metric or a named outcome —
who is better off, and by how much — not the change you made.
"Reduced p99 from 600ms to 120ms" beats "made it faster"; "unblocked mobile v3
release" beats "refactored auth". A brag entry without a specific outcome is a
reminder, not an artifact.

See [`BRAG.md`](../BRAG.md) for the full composition guide, the field quality
bar, worked examples, and the approval loop — propose the entry, wait for the
user's approval, and only then capture.
