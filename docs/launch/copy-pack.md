# Launch copy pack

Ready-to-paste promotion copy for bragfile, organized by **facet** — the same
tool described through the lens each community cares about (agent / Go / local-first
/ career). The goal is **usage/adoption**, not revenue.

## Integrity rules (read before pasting)

- **Stay true to shipped reality.** This copy describes **v0.5.1**: automatic
  capture, provenance tags (`agent:`/`model:`/`cost:`/`session:`/`commit:`), a
  typed MCP write surface, FTS5 search, digests/story/wrapped/export, local
  SQLite, no account/server.
- **Do NOT claim the agent-memory read-back loop** (an agent reading its own
  history back as context). That is PROJ-006 and **not yet built**. Add the
  "…and reads it back as memory" line to every block *after* it ships — it is
  the line that turns "another logger" into "infrastructure," and it is worth
  holding the marquee Show HN until you can say it truthfully.
- **Post it yourself.** Show HN is for the creator. Do **not** recruit people to
  post or upvote — coordinated posting / vote solicitation is bannable on HN,
  Lobsters, and Reddit and backfires. Build standing and an audience so your own
  post lands; don't manufacture proxies.
- **Fill in the placeholders:** `<repo-url>`, `<tap>`.

## The five doorways

bragfile enters five distinct communities that won't hear about it from each
other. Match the facet to the venue every time:

| Facet | Lead with | Venues |
|---|---|---|
| AI-agent tool | agent writes it, provenance | MCP directories, r/ClaudeAI, Cursor/Claude Discords, AI newsletters |
| Go CLI | pure-Go/no-cgo SQLite, FTS5 | r/golang, awesome-go, Golang Weekly, Terminal Trove |
| Local-first | own your data, no server | r/selfhosted, local-first community, awesome-selfhosted |
| Brag doc / career | writes itself for reviews | LinkedIn, AlternativeTo, "brag document" SEO post |
| MCP primitive | the format / write surface | MCP registries, `bragfmt` spec |

---

## Show HN — title (pick one)

```
Show HN: Bragfile – a local work journal your AI coding agent writes to
Show HN: A provenance-stamped work log your AI agent fills in (Go, SQLite, no account)
```

## Show HN — first comment

```
I build with Claude Code and Cursor all day, and I kept losing track of what
the agents actually did — and what each session cost. So I made bragfile: a
small CLI that keeps a work journal, which your agent can write to over MCP.

Each entry can carry provenance in the tags — which agent, which model, token
count, cost, the commit hash — so months later you can tell agent-authored work
from your own, and see what a given model actually shipped.

Design choices, since they're the point:
- Local-first. It's a single SQLite file at ~/.bragfile you own. No account,
  no server, no telemetry. `brew install` and it works offline.
- Pure-Go SQLite (no cgo), embedded migrations, FTS5 full-text search.
- Capture is meant to be near-zero effort — the agent proposes entries; you
  approve. It also does weekly/monthly/quarterly digests and export.

It's honestly still young (v0.5.1) and single-user. The next thing I'm building
is letting an agent *read* its own history back as context so it stops
re-deriving the same project state every session — feedback on that direction
especially welcome.

Repo: <repo-url>   Install: brew install <tap>/bragfile
```

---

## MCP directory blurb

*mcp.so · Smithery · Glama · PulseMCP · awesome-mcp-servers · modelcontextprotocol/servers*

```
bragfile — a local-first work-journal MCP server. Agents log what they shipped
via brag_add (with brag_list / brag_search / brag_stats read tools), stamped
with agent/model/cost/session provenance. Backed by a single local SQLite file
the user owns — no account, no network. Go, no cgo.
```

## Go / CLI blurb

*r/golang · awesome-go · Terminal Trove*

```
bragfile — a personal work-journal CLI in pure Go. Pure-Go SQLite driver (no
cgo), embedded forward-only migrations, FTS5 full-text search, and a typed MCP
server so AI coding agents can write entries with provenance. Single local
SQLite file, brew-installable, no account or server. Digests, tags, cwd-aware
projects, JSON/markdown export.
```

## Local-first / self-hosted blurb

*r/selfhosted · local-first community · awesome-selfhosted*

```
bragfile — own your work history. A local CLI that logs what you (and your AI
agents) shipped into a single SQLite file on your machine. No account, no
server, no telemetry — it runs offline and the data never leaves your disk. AI
agents can write to it over MCP with provenance, so it fills itself without
becoming surveillance.
```

## Brag-doc / career blurb

*LinkedIn · AlternativeTo · the "brag document" SEO post*

```
The "brag document" — the running list of your wins for reviews, promo packets,
and resumes — except you don't have to remember to update it. bragfile is a
local CLI where your AI coding agents log what shipped as they work, then rolls
it up into weekly/monthly/quarterly digests and a shareable year-in-review. Your
data stays in a file you own.
```

## Reddit post — r/ClaudeAI (problem-first)

```
Title: My agents kept losing track of what they'd done, so I built a local work
log they write to

Anyone else lose the thread of what Claude Code actually shipped across a week
of sessions — and what it cost? I made a small local CLI (bragfile) that the
agent writes to over MCP: what it did, which model, token/cost, the commit. It's
just a SQLite file on my machine — no account, no server. Does weekly/quarterly
digests too.

Still single-user and early (v0.5.1). Curious whether others want this, and what
you'd want it to capture. Repo: <repo-url>
```

## README hero / one-liner

*Top of repo · X/Bluesky bio · anywhere you need one sentence*

```
A local work journal your AI coding agent writes to — provenance included, no
account, no server. brew install <tap>/bragfile
```

---

## Presence-building checklist (do this BEFORE the launch spike)

The launch platforms have gates and cultures; a cold account posting to silence
is a wasted first impression. Earn standing and an audience first.

- [ ] **Seed the evergreen directories now** (not launch moments): MCP dirs,
      awesome-go, awesome-mcp-servers, Terminal Trove, awesome-selfhosted. One
      PR each; they compound quietly and make a later launch look already-real.
- [ ] **Tag the GitHub repo** so topic-browse discovery works: `mcp`,
      `local-first`, `cli`, `ai`, `golang`, `claude`, `agent`.
- [ ] **Build in public** on X and/or Bluesky — short weekly "shipped this"
      posts. This is where the people who *organically* share your launch come
      from. Relationships, not recruitment.
- [ ] **Start a weeknotes thread** (dev.to / blog) so HN/Reddit have something
      to link to that isn't just the repo.
- [ ] **Earn standing** on HN / Lobsters / Reddit as a *participant* first —
      comment helpfully in agent/CLI/local-first threads. Lobsters needs an
      invite from an existing member; line that up early.
- [ ] **Get your first real users** in small Discords (Claude Code, MCP,
      local-first). They become your first advocates.

### Launch sequence (once the agent-memory read-back ships)

1. Directories + GitHub topics seeded (evergreen — already done above).
2. Show HN — post it **yourself**, once. This is your one clean first impression.
3. Same week: r/ClaudeAI, r/golang, r/selfhosted — each with its own facet opener.
4. Newsletters (Console.dev, Golang Weekly, Changelog News, an AI-dev letter).
5. Product Hunt / Lobsters as a second wave.
6. LinkedIn + the "brag document" SEO post for the career doorway.

**Hold steps 2–6 until the memory loop is real** — a spike into "another logger"
wastes the one first impression each venue gives you.
