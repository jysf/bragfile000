# Bragfile

`brag` is a local-first command-line tool that captures your
brag-worthy work moments — shipped features, fixed bugs, things you
learned, mentoring you delivered — and lets you retrieve them later
for retros, reviews, and resumes. Entries live in an embedded SQLite
database at `~/.bragfile/db.sqlite` on your machine. No cloud, no
sync, no account.

> **Status:** v0.5.1 shipped. Capture, retrieve, search, export,
> weekly/monthly/quarterly digests, tags, and cwd-aware projects are all
> here — and a local MCP server now lets AI coding agents capture and read
> entries as typed tool calls. `brew install jysf/tap/bragfile` installs
> the binary on macOS.
>
> Working with an AI agent? `brag mcp install` wires brag into Claude Code,
> Cursor, or Claude Desktop as five typed tools plus an auto-loadable memory
> resource — see
> [Using brag from an AI agent (MCP)](#using-brag-from-an-ai-agent-mcp) below.

## Install

Homebrew (recommended):

```bash
brew install jysf/tap/bragfile
brag --version
```

> bragfile is distributed as a Homebrew **formula** (DEC-040), not a cask — so
> installing needs no code-signing, no Gatekeeper "Apple could not verify…"
> prompt, no `xattr`, and no `brew trust` step, even though the binary is
> unsigned. (Being a formula is what avoids Apple Developer Program dues without
> the signing friction a cask would require; the `crustyimg` formula on the same
> tap installs the same clean way.)

### Upgrading from 0.5.1 or earlier

The formula above replaced a Homebrew **cask** on a different tap in v0.5.2.
`brew upgrade` does not cross that boundary, and Homebrew gives you no signal
that it hasn't: `brew outdated` stays silent while you sit on the old version.

Check what you actually have:

```bash
brag --version
```

If it prints `0.5.1` or lower, migrate once:

```bash
brew uninstall --cask bragfile
brew untap jysf/bragfile
brew install jysf/tap/bragfile
```

Your data is not affected — the database lives in `~/.bragfile`, not in the
Cellar or Caskroom. Confirm with `brag --version` and `brag list`.

From source (works today):

```bash
git clone https://github.com/jysf/bragfile000.git
cd bragfile000
just install                 # or: go install ./cmd/brag
brag --version               # confirm ~/go/bin is on $PATH
```

The Homebrew install pulls a prebuilt binary — no Go required.
Requires Go 1.26+ if you build from source instead.

- **Claude Code plugin:** `claude plugin marketplace add jysf/bragfile000`
  then `claude plugin install brag@bragfile` — see `plugin/README.md`.
- **MCP server (any client):** `brag mcp install` registers the `brag mcp
  serve` server in a client's config idempotently (`--client
  claude-code|claude-desktop|cursor`, `--scope project|user`, `--dry-run` to
  preview) — see [`docs/api-contract.md`](docs/api-contract.md).

## Capture an entry

The fastest path — one flag:

```bash
brag add --title "shipped FTS5 search end-to-end"
# prints the new entry's ID on stdout, e.g. "12"
```

With full metadata:

```bash
brag add \
  --title "cut p99 login latency from 600ms to 120ms" \
  --project platform \
  --type shipped \
  --tags auth,perf,backend \
  --impact "unblocked mobile v3 release"
```

For longer narrative entries, `brag add` with no flags opens
`$EDITOR` against a templated buffer:

```bash
brag add        # → editor opens; fill in the fields, save, quit
```

For programmatic capture from a script or AI agent, pipe a single JSON object
to `brag add --json`. Only `title` is required, and stdout is just the new
entry's ID, so it composes:

```bash
echo '{"title":"shipped the auth refactor"}' | brag add --json
```

All fields, with a heredoc so you do not fight shell quoting:

```bash
cat <<'EOF' | brag add --json
{
  "title": "Cut p99 latency on the auth path",
  "description": "Replaced the per-request JWKS fetch with a 5-minute in-process cache.",
  "project": "platform",
  "type": "ship",
  "tags": "auth,perf,backend",
  "impact": "p99 1.8s -> 240ms; unblocked the mobile v3 release"
}
EOF
```

Note `tags` is a **comma-joined string**, not an array (`["auth","perf"]` is
rejected, naming [DEC-004](decisions/DEC-004-tags-comma-joined-for-mvp.md)).
Unknown keys are rejected with the offending key named, so a typo like
`"titl"` fails loudly instead of silently dropping the field.

`brag list --format json` emits the same shape and `id`/`created_at`/
`updated_at` are ignored on input, so entries round-trip between databases
with no transform:

```bash
brag list --format json | jq -c '.[0]' | brag add --json --db /path/to/other.sqlite
```

Full contract in [`BRAG.md`](BRAG.md); the schema is checked in at
[`docs/brag-entry.schema.json`](docs/brag-entry.schema.json).

## Read entries back

List them, newest first:

```bash
brag list                                  # all entries
brag list --project platform --since 30d   # filter by project + window
brag list -P                               # add a project column
brag list --format json                    # machine-readable
```

Search across every field via SQLite FTS5:

```bash
brag search "latency"
brag search "auth-refactor"     # hyphens are literal, not operators
```

Show the full record for a single entry, edit it, or delete it:

```bash
brag show 12
brag edit 12
brag delete 12
```

## Export for reviews

Markdown report grouped by project (paste into a quarterly review
or promo packet):

```bash
brag export --format markdown --since 90d > q-review.md
```

JSON dump (for AI piping or backup):

```bash
brag export --format json --since 90d > q-review.json
```

To publish a slice of brags to a website (filter, then reshape into
clean blog prose with `jq`), see the tutorial's
[Publish your brags to a website](docs/tutorial.md) section.

## Weekly and monthly digests

Rule-based aggregations of recent entries — no LLM, no network.
Pipe the JSON into your favourite AI session for guided
reflection.

```bash
brag summary --range week               # 7-day digest, grouped
brag summary --range month --format json
brag review --week                      # entries + reflection prompts
brag stats                              # lifetime metrics
brag impact --quarter                   # this quarter's impact, by initiative
brag wrapped 2026                       # shareable year-in-review; also: brag wrapped 2026 Q3
brag coverage --year                    # agent-vs-human provenance share + monthly trend
brag story --audience exec --quarter    # audience-shaped narrative bundle for an LLM
brag spark                              # sparkline pulse of recent activity (Total + per-project)
brag memory --query auth --project orbit # ranked, token-budgeted slice of your history
```

## Where the data lives

```
~/.bragfile/db.sqlite
```

Back up by copying the file. Move to a new machine by copying the
file. Override the path with the `--db` flag or the `BRAGFILE_DB`
environment variable.

## Using brag from an AI agent (MCP)

`brag` ships a local, stdio-only MCP server so AI coding agents can capture and
retrieve entries as typed tool calls. Register it in one command:

```bash
brag mcp install                 # claude-code, project scope (writes ./.mcp.json)
```

Then reconnect your client (MCP servers connect at startup). The full agent
playbook — the five tool schemas, the three auto-loadable `brag://` resources,
the `project`-not-auto-filled gotcha, provenance stamping, and the `--db`
override — is in
[`docs/for-ai-agents.md`](docs/for-ai-agents.md).

## Where to go next

- [`docs/tutorial.md`](docs/tutorial.md) — the deep-dive
  walkthrough: every command, every flag, every gotcha.
- [`BRAG.md`](BRAG.md) — guide for AI coding agents that want to
  propose brag entries from work sessions.
- [`docs/for-ai-agents.md`](docs/for-ai-agents.md) — the MCP playbook for AI
  agents: register the server, the five tool schemas, the `brag://` resources,
  and the gotchas.
- [`CONTRIBUTING.md`](CONTRIBUTING.md) — how this repo is built
  and how to contribute.
- [`docs/api-contract.md`](docs/api-contract.md) — full CLI
  reference.

## License

MIT. See [`LICENSE`](LICENSE).
