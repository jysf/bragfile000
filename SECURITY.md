# Security Policy

`bragfile` is a local-first CLI. It stores your entries in a SQLite database at
`~/.bragfile/db.sqlite` on your own machine, and the binary makes **no network
calls of any kind** — no telemetry, no sync, no update check, no LLM. That is a
design constraint (DEC-001), not a default, and it is enforced by review rather
than assumed.

## Supported versions

Only the **latest released version** receives security fixes. Check yours with
`brag --version` and see the [CHANGELOG](CHANGELOG.md) for what is current.

> **If you installed before v0.5.2, you are probably not on the latest version
> and will not be told so.** Distribution moved from a Homebrew cask to a
> formula, and `brew upgrade` does not cross that boundary — `brew outdated`
> stays silent. See [README §Install](README.md#install) for the one-time
> migration.

## Reporting a vulnerability

Report privately through GitHub's
[security advisory form](https://github.com/jysf/bragfile000/security/advisories/new)
rather than opening a public issue.

Please include what you did, what happened, and what you expected — a
reproduction is worth more than a severity rating. This is a personal project
maintained by one person: expect an acknowledgement within about a week, and
expect a fix to ship as a normal release rather than an out-of-band patch unless
the issue is actively exploitable.

There is no bug bounty.

## What is in scope

The threat model is a single-user tool handling the user's own data, so the
interesting surfaces are the ones where *someone else's* bytes reach it:

- **Ingress parsing** — `brag add --json` (stdin), the MCP server's tool
  arguments, and any field that survives into storage. All ingress paths run
  the same `capture.Validate` caps (DEC-046).
- **The MCP server** (`brag mcp serve`) — it speaks to an agent over stdio and
  writes to your real database. `brag mcp install` also rewrites a client
  config file you did not author, which is why that write is atomic.
- **Anything that runs a subprocess or touches a path you did not type** —
  `$EDITOR` launching, the database path, backup sidecars.
- **Distribution integrity** — the release archives, checksums, and the
  Homebrew formula.

## What is out of scope

- **Local attackers who already have your user account.** The database is a
  plain SQLite file with your own permissions; anyone who can read your home
  directory can read it. Encrypting it would not change that, and bragfile does
  not pretend otherwise.
- **Content you deliberately put in your own entries.** `brag` stores text you
  wrote; it does not sanitise your prose.
- **The LLM you pipe output into.** `brag story`/`brag review` emit text for you
  to paste or pipe somewhere else. What that destination does with it is
  outside this tool.

## Prior review

A full pre-distribution security review was run on 2026-04-26 and is checked
into this repository:
[`docs/reports/security/2026-04-26-pre-distribution-security-review.md`](docs/reports/security/2026-04-26-pre-distribution-security-review.md).
Its findings were triaged, fixed, or explicitly deferred with reasons — the
deferrals are recorded in `projects/PROJ-001-mvp/backlog.md` rather than dropped.
