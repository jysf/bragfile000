# Bragfile

`brag` is a local-first command-line tool that captures your
brag-worthy work moments — shipped features, fixed bugs, things you
learned, mentoring you delivered — and lets you retrieve them later
for retros, reviews, and resumes. Entries live in an embedded SQLite
database at `~/.bragfile/db.sqlite` on your machine. No cloud, no
sync, no account.

> **Status:** v0.2.0 shipped. Capture, retrieve, search, export, and
> weekly/monthly digests, plus first-class tags and projects with
> cwd-aware capture, are all available, and `brew install
> jysf/bragfile/bragfile` installs the binary on macOS.

## Install

Homebrew (recommended):

```bash
brew install jysf/bragfile/bragfile
brag --version
```

> **Homebrew 6.0+ note.** bragfile is distributed as a cask from a
> third-party tap, and newer Homebrew refuses to load such casks until
> you trust the tap once (you'll see *"Refusing to load cask … from
> untrusted tap"*). If that happens, run this one-time, per-machine step,
> then re-run the install:
> ```bash
> brew trust --cask jysf/bragfile/bragfile
> ```

From source (works today):

```bash
git clone https://github.com/jysf/bragfile000.git
cd bragfile000
just install                 # or: go install ./cmd/brag
brag --version               # confirm ~/go/bin is on $PATH
```

The Homebrew install pulls a prebuilt binary — no Go required.
Requires Go 1.26+ if you build from source instead.

**macOS Gatekeeper note.** The Homebrew-installed binary is not
Apple-signed/notarized (v0.2.0 ships unsigned to keep the personal
project free of Apple Developer Program dues). The first time you
run `brag` after `brew install`, macOS may say *"Apple could not
verify 'brag' is free of malware…"* and refuse to execute it.
Clear the quarantine attribute once:

```bash
sudo xattr -dr com.apple.quarantine /opt/homebrew/Caskroom/bragfile/
```

Then `brag --version` works normally. You only need to do this
once per `brew install` / `brew upgrade`. If you'd rather avoid
the warning entirely, build from source instead — locally-built
binaries don't trigger Gatekeeper.

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

For programmatic capture from a script or AI agent, pipe JSON to
`brag add --json` (see [`BRAG.md`](BRAG.md)):

```bash
echo '{"title":"…","project":"…"}' | brag add --json
```

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
playbook — the four tool schemas, the `project`-not-auto-filled gotcha,
provenance stamping, and the `--db` override — is in
[`docs/for-ai-agents.md`](docs/for-ai-agents.md).

## Where to go next

- [`docs/tutorial.md`](docs/tutorial.md) — the deep-dive
  walkthrough: every command, every flag, every gotcha.
- [`BRAG.md`](BRAG.md) — guide for AI coding agents that want to
  propose brag entries from work sessions.
- [`docs/for-ai-agents.md`](docs/for-ai-agents.md) — the MCP playbook for AI
  agents: register the server, the four tool schemas, and the gotchas.
- [`CONTRIBUTING.md`](CONTRIBUTING.md) — how this repo is built
  and how to contribute.
- [`docs/api-contract.md`](docs/api-contract.md) — full CLI
  reference.

## License

MIT. See [`LICENSE`](LICENSE).
