# brag — Claude Code plugin

Bundles three ways to capture accomplishments without leaving a Claude Code
session:

- **MCP server** (`brag mcp serve`) — `brag_add` / `brag_list` /
  `brag_search` / `brag_stats` as typed tools over your `~/.bragfile/db.sqlite`.
- **`/brag:brag` slash-command** — draft a brag from the current session for
  your approval.
- **Capture-nudge hook** — after a commit lands in a session, quietly nudges
  the agent to propose a brag (it never posts one for you).

## Prerequisite

The plugin's MCP server runs the `brag` binary from your `PATH`. Install it
first:

    brew trust --cask jysf/bragfile/bragfile   # one-time (Homebrew 6.0+)
    brew install jysf/bragfile/bragfile

Verify: `brag --version`.

## Install

    claude plugin marketplace add jysf/bragfile000
    claude plugin install brag@bragfile

Then restart Claude Code. `claude plugin details brag` shows the loaded
components.

## Provenance convention

Agent-driven brags carry reserved-namespace tags so multi-agent work is
attributable later: `agent:<name>` (e.g. `agent:claude-code`) and
`model:<id>` (e.g. `model:claude-opus-4-8`) — lowercase, no spaces. The MCP
`brag_add` tool stamps these; query them with
`brag list --tag model:claude-opus-4-8` and `brag tags`.

## Silence the nudge

    export BRAG_CAPTURE_NUDGE=off

## Manual (non-plugin) path

Prefer to wire things by hand? Copy `examples/brag-slash-command.md` to
`~/.claude/commands/brag.md` for a bare `/brag`, and see BRAG.md for the
`scripts/claude-code-post-session.sh` pipe helper.
