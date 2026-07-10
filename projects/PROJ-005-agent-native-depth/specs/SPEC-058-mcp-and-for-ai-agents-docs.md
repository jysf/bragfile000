---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-058
  type: story                      # epic | story | task | bug | chore
  cycle: design                    # frame | design | build | verify | ship
  blocked: false
  priority: medium
  complexity: S                    # S | M | L  (L means split it)

project:
  id: PROJ-005
  stage: STAGE-015
repo:
  id: bragfile

agents:
  architect: claude-opus-4-8
  implementer: claude-opus-4-8     # usually same Claude, different session
  created_at: 2026-07-10

references:
  decisions: [DEC-024, DEC-027, DEC-034, DEC-036]
  constraints: []                  # pure docs; no code constraint applies (see Implementation Context)
  related_specs: [SPEC-055, SPEC-057, SPEC-040, SPEC-041, SPEC-021, SPEC-022]
---

# SPEC-058: MCP and for-AI-agents docs

## Context

This is the **last** spec in STAGE-015 (MCP first-class for agents) and it
closes the stage. SPEC-040 shipped `brag mcp serve` (four typed tools),
SPEC-041 shipped the Claude Code plugin, SPEC-055 shipped `brag mcp install`
(DEC-034), and SPEC-057 shipped `brag project ensure` (DEC-036). The
*capability* to connect an agent over MCP and log correctly-attributed wins
is fully built — but there is no single, agent-facing document that lifts it
into a self-serve playbook. The stage's motivating observation (STAGE-015
"Why Now") was a live Claude Code agent falling back to `brag add -p standup`
on the CLI because nothing told it how to connect over MCP.

This spec marks the front door. It **lifts already-true, already-shipped
behavior** into an agent-friendly form so an AI agent can connect to
`brag mcp serve` and log wins correctly with **no source-diving**. It invents
nothing: every claim is transcribed from source and from `docs/api-contract.md`
(the CLI contract) and verified against the shipped code.

- Parent stage: `STAGE-015` — this is the third and final backlog item.
- Project: `PROJ-005` (agent-native depth).
- Prior work it documents: SPEC-040/DEC-024 (serve), SPEC-055/DEC-034
  (install), SPEC-057/DEC-036 (project ensure), DEC-027 (seed provenance).

## Goal

Ship an agent-facing documentation page (`docs/for-ai-agents.md`) plus a short
README section that together let an AI agent register the brag MCP server,
learn the exact four-tool contract, and log correctly-attributed,
correctly-projected wins — all from docs, without reading source.

## Inputs

- **Files to read:**
  - `internal/mcpserver/server.go` — the `addIn`/`listIn`/`searchIn`/`statsIn`
    structs + handlers = the source of truth for tool params, types,
    required/optional, and return shapes.
  - `internal/mcpserver/provenance.go` — how `agent:`/`model:`/`session:`/
    `cost:`/`tokens:` are normalized and stamped.
  - `internal/cli/mcp_install.go` — the shipped `brag mcp install` (flags,
    per-client×scope target-path resolution).
  - `internal/cli/mcp.go` — `brag mcp serve` reads the root `--db` persistent
    flag via `config.ResolveDBPath`.
  - `internal/config/config.go` — DB path resolution (DEC-003).
  - `docs/api-contract.md` §`brag mcp serve`, §`brag mcp install`,
    §`brag project ensure` — the transcription baseline.
  - `BRAG.md` — the existing composition guide (approval loop, impact-framing
    convention); the new page cross-links it and must not duplicate it.
- **Related code paths:** none modified — this spec is docs-only.

## Outputs

- **Files created:** `docs/for-ai-agents.md` — the full agent MCP playbook
  (registration, reconnect note, four tool schemas, the `project` gotcha,
  provenance stamping, `--db` override, impact-framing playbook). Literal
  content embedded verbatim under `## Notes for the Implementer`.
- **Files modified:**
  - `README.md` — adds a `## Using brag from an AI agent (MCP)` section that
    points to the new page and the one-command install, plus a
    `docs/for-ai-agents.md` bullet under `## Where to go next`. Literal
    content embedded under `## Notes for the Implementer`.
  - `scripts/test-docs.sh` — a new lettered assertion group (Group T) added
    during **this design cycle** (see `## Failing Tests`).
- **New exports:** none.
- **Database changes:** none.
- **Premise audit (additive-collection / status-change / doc-consistency):**
  - `grep -rn "for-ai-agents" . --include='*.md' --include='*.sh'` → **zero
    hits** outside this spec (run 2026-07-10). The doc name is unclaimed, so
    no existing link-integrity or reference assertion is affected by creating
    it.
  - `scripts/test-docs.sh` currently greps for **none** of `brag_add`/
    `brag_list`/`brag_search`/`brag_stats` (run 2026-07-10) — Group T is a
    genuinely new set of assertions, colliding with no existing test name.
  - README's existing `A1` line-count band is `100..250`. README is currently
    187 lines; the added section (≈16 lines) keeps it well inside the band.
    `A5` (all 7 verbs in fenced blocks) is untouched — the new section adds
    no fenced block that removes a verb.
  - The existing `E1` internal-link assertion runs over README.md — once build
    adds the `docs/for-ai-agents.md` link AND the file, `E1` stays green; both
    land in the same build cycle.

## Acceptance Criteria

- [ ] `docs/for-ai-agents.md` exists and is a coherent agent playbook.
- [ ] The page documents `brag mcp install` with the resolved config path per
      client/scope AND the manual `mcpServers` JSON snippet for anyone not
      using `install`.
- [ ] The page states the client-startup-reconnect note.
- [ ] The page gives the full param contract (name, type, required/optional)
      and return shape for all four tools (`brag_add`, `brag_list`,
      `brag_search`, `brag_stats`), matching `internal/mcpserver/server.go`.
- [ ] The page states the `project`-not-auto-filled-over-MCP gotcha and
      cross-links `brag project ensure <name>`.
- [ ] The page documents provenance stamping (`agent:`/`model:`/`session:`/
      `cost:`/`tokens:`) matching `internal/mcpserver/provenance.go` +
      DEC-024/DEC-027.
- [ ] The page documents the `--db` → `BRAGFILE_DB` → default resolution
      order (DEC-003).
- [ ] The page carries a "log a win" playbook with the impact-framing
      convention (outcome, not output), consistent with `BRAG.md`.
- [ ] `README.md` has an agent/MCP section linking `docs/for-ai-agents.md`.
- [ ] Every claim is lifted from shipped source — no invented behavior.
      (Any temptation to decide behavior is recorded under Open Questions,
      NOT resolved in the doc.)

## Failing Tests

Written during **design**, BEFORE build. These are doc-content assertions
added to `scripts/test-docs.sh` as **Group T** (T is the next unused letter
after S). They FAIL today (the page does not exist; README lacks the link)
and PASS once build transcribes the embedded literals. All are grep-based and
follow the existing lettered-group shape; per AGENTS.md §9, correctness rests
on `assert_contains_literal` / band checks, never on `--exclude-dir`.

Design-time note (§12(a) expected-value literals): every literal asserted
below is a byte-for-byte substring of the embedded doc/README text in
`## Notes for the Implementer` — reconciled at design, so build's transcribe
step and these assertions cannot disagree.

- **`scripts/test-docs.sh` — Group T (for-ai-agents docs, SPEC-058)**
  - `T1` — asserts `docs/for-ai-agents.md` exists (`assert_file_exists`).
  - `T2` — asserts `docs/for-ai-agents.md` line count is in band `120..500`
    (`assert_line_count_band`) — a full playbook, not a stub.
  - `T3` — asserts the page contains **all four** tool names, iterating
    internally over `brag_add`, `brag_list`, `brag_search`, `brag_stats`
    (single named assertion, `o4`-style loop). Fails listing any missing name.
  - `T4` — asserts the page contains the registration command literal
    `brag mcp install` (`assert_contains_literal`).
  - `T5` — asserts the page contains the manual JSON snippet literal
    `{"mcpServers":{"brag":{"command":"brag","args":["mcp","serve"]}}}`.
  - `T6` — asserts the page contains the reconnect note: literals
    `connect at client startup` AND `reconnect` (both required).
  - `T7` — asserts the page contains the gotcha phrase literal
    `does not auto-fill` AND the literal `project` (gotcha names the field).
  - `T8` — asserts the page cross-links the fix: literal
    `brag project ensure`.
  - `T9` — asserts the page documents provenance: iterates over the five
    reserved-namespace literals `agent:<name>`, `model:<id>`, `session:<id>`,
    `cost:<n>`, `tokens:<n>` (all required; fails listing any missing).
  - `T10` — asserts the page documents DB resolution: iterates over the three
    literals `--db`, `BRAGFILE_DB`, `~/.bragfile/db.sqlite` (all required).
  - `T11` — asserts the page contains the impact-framing convention literal
    `a metric or a named outcome` (distinctive phrase, lifted from `BRAG.md`).
  - `T12` — asserts the page contains the resolved per-client config paths:
    iterates over the four literals `.mcp.json`, `~/.claude.json`,
    `.cursor/mcp.json`, `claude_desktop_config.json` (all required).
  - `T13` — asserts `README.md` links the page: literal `docs/for-ai-agents.md`
    (`assert_contains_literal`).
  - `T14` — asserts `README.md` has an agent/MCP section heading: a line
    matching `^## .*[Aa]gent` (line-regex, avoids substring trap per §9).

## Implementation Context

*Read this section (and the files it points to) before starting
the build cycle. It is the equivalent of a handoff document, folded
into the spec since there is no separate receiving agent.*

This is a **literal-artifact-as-spec** deliverable (AGENTS.md §12): the doc
page and README section are fixed-shape markdown decidable at design time.
Build **transcribes the literals verbatim** from `## Notes for the
Implementer`; verify **diffs** the shipped files against those literals. Do
NOT paraphrase — the Group T assertions are keyed to exact substrings of the
embedded text.

### Decisions that apply

- `DEC-024` — `brag mcp serve`: official Go MCP SDK, stdio-only, four tools,
  transport purity, and provenance via reserved `agent:`/`model:` tags with
  the `agent`→`clientInfo.Name` fallback. The tool contract + provenance the
  page documents.
- `DEC-027` — seed `session:`/`cost:`/`tokens:` reserved tags on `brag_add`
  (migration-free; MCP-path-only; validated at the handler boundary — bad
  `cost`/`tokens` is a tool error, never a coerced tag). The seed-provenance
  content in §5.
- `DEC-034` — `brag mcp install`: idempotent no-clobber merge, per-client×scope
  target-path table, dry-run(JSON→stdout)/human(→stderr). The registration
  content + path table in §1.
- `DEC-036` — `brag project ensure`: idempotent register-by-name; `brag_add`
  stays free-text (no silent auto-register). The cross-link in §4.
- `DEC-003` (informational; not in front-matter references, no behavior of
  this spec turns on it beyond transcription) — DB path resolution order:
  flag → `BRAGFILE_DB` → default `~/.bragfile/db.sqlite`. The §6 content.

### Constraints that apply

This spec ships **markdown only** — no Go, no shell logic beyond grep
assertions. None of the blocking constraints in `/guidance/constraints.yaml`
govern documentation content:

- `no-sql-in-cli-layer`, `no-cgo`, `errors-wrap-with-context`,
  `storage-tests-use-tempdir`, etc. — code-path constraints, N/A to docs.
- `stdout-is-for-data-stderr-is-for-humans` — the page *describes* this
  behavior (dry-run JSON on stdout, human lines on stderr) accurately, but the
  spec introduces no code that could violate it.
- `one-spec-per-pr` — honored: one branch, one PR.

`references.constraints` is therefore intentionally empty.

### Prior related work

- `SPEC-040` (shipped) — `brag mcp serve` + the four tools.
- `SPEC-041` (shipped) — Claude Code plugin (`plugin/.mcp.json`, hook,
  slash-command). The page's "not using install?" path and the plugin path
  are complementary; the page links the plugin route lightly and does not
  re-document it.
- `SPEC-055` (shipped) — `brag mcp install` (DEC-034).
- `SPEC-057` (shipped) — `brag project ensure` (DEC-036).
- `SPEC-021` (shipped) — README user-facing rewrite + `scripts/test-docs.sh`
  harness introduction; the lettered-group assertion pattern reused here.
- `SPEC-022` (shipped) — `BRAG.md` + JSON schema + hook/slash assets; the
  composition guide the new page defers to.

### Out of scope (for this spec specifically)

- Any change to code behavior. If documenting surfaces a bug or a gap, file a
  new spec — do not fix it here.
- Re-documenting the CLI `brag add`/`--json` composition flow already covered
  by `BRAG.md` — the page cross-links it, it does not duplicate it.
- Re-documenting the plugin packaging (SPEC-041) beyond a one-line pointer.
- A `--since` MCP filter (deferred at SPEC-040; the page states its absence,
  it does not propose adding it).
- Any DEC. This spec lifts shipped behavior; it makes no architectural or
  behavioral decision. **DEC-037 is reserved for another spec — do not create
  or claim it.** If build finds itself needing to decide behavior, that means
  content is being invented: stop and raise it (see Open Questions).

### Open Questions

None block build. Flagged for the orchestrator, since each is a place where
writing the doc risked *inventing* rather than *lifting* — resolved by lifting
only:

1. **Windows `claude-desktop` path.** `internal/cli/mcp_install.go` resolves a
   Windows path (`%APPDATA%\Claude\claude_desktop_config.json`) and
   `api-contract.md` lists it. The bragfile audience is macOS-first (Homebrew
   cask, Gatekeeper notes). The page documents the **macOS** claude-desktop
   path in prose (matching the repo's macOS-first posture) and points to
   `brag mcp install --dry-run` / `docs/api-contract.md` for the exact path on
   other OSes, rather than transcribing the full OS matrix. This is an
   editorial scoping choice, not a behavior claim — no DEC. (Group T asserts
   only the macOS filename `claude_desktop_config.json`.)
2. **`cursor` provenance fallback.** `agent` falls back to the MCP client's
   `clientInfo.Name`; what a non-Claude client (cursor) reports there is
   client-defined, not bragfile-defined. The page states the fallback
   mechanism (lift) and advises passing `agent` explicitly for determinism —
   it does not claim a specific cursor value (which would be invention).

## Notes for the Implementer

Transcribe the two literals below **verbatim**. `docs/for-ai-agents.md` is a
new file; the README changes are (a) a new `## Using brag from an AI agent
(MCP)` section and (b) a new bullet in the existing `## Where to go next`
list. Preferred placement of the new README section: immediately **before**
`## Where to go next` (so it follows `## Where the data lives`). Do not
reorder or alter other README sections.

Style: match the surrounding docs (kebab-case filename, `##`/`###` headings,
fenced code blocks, relative links). The page lives in `docs/`, so links to
root files use `../` (e.g. `../BRAG.md`) and to sibling docs use `./`
(e.g. `./api-contract.md`).

---

### LITERAL 1 — `docs/for-ai-agents.md` (new file, transcribe verbatim)

```markdown
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

**Frame `impact` as the outcome, not the output.** State a metric or a named
outcome — who is better off, and by how much — not the change you made.
"Reduced p99 from 600ms to 120ms" beats "made it faster"; "unblocked mobile v3
release" beats "refactored auth". A brag entry without a specific outcome is a
reminder, not an artifact.

See [`BRAG.md`](../BRAG.md) for the full composition guide, the field quality
bar, worked examples, and the approval loop — propose the entry, wait for the
user's approval, and only then capture.
```

---

### LITERAL 2a — `README.md` new section (insert immediately before `## Where to go next`)

```markdown
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

```

### LITERAL 2b — `README.md` new bullet (add to the existing `## Where to go next` list)

Add this bullet after the `BRAG.md` bullet in the `## Where to go next` list:

```markdown
- [`docs/for-ai-agents.md`](docs/for-ai-agents.md) — the MCP playbook for AI
  agents: register the server, the four tool schemas, and the gotchas.
```

---

### LITERAL 3 — `scripts/test-docs.sh` Group T (add before the `# ===== finalise =====` block)

Transcribe this group verbatim. It follows the existing lettered-group shape
(`assert_file_exists` / `assert_line_count_band` / `assert_contains_literal`
helpers already defined at the top of the script; the `o4`-style internal loop
for multi-literal checks).

```bash
# ===== Group T — for-ai-agents docs (SPEC-058) =====

AGENT_DOC="docs/for-ai-agents.md"

# T1 — page exists
assert_file_exists "T1" "$AGENT_DOC"

# T2 — page is a full playbook, not a stub
assert_line_count_band "T2" "$AGENT_DOC" 120 500

# T3 — names all four MCP tools
if [ ! -f "$AGENT_DOC" ]; then
    fail "T3" "$AGENT_DOC does not exist"
else
    t3_missing=""
    for tool in brag_add brag_list brag_search brag_stats; do
        if ! grep -F -q -- "$tool" "$AGENT_DOC"; then
            t3_missing="$t3_missing $tool"
        fi
    done
    if [ -z "$t3_missing" ]; then
        ok "T3"
    else
        fail "T3" "$AGENT_DOC missing tool names:$t3_missing"
    fi
fi

# T4 — documents the registration command
assert_contains_literal "T4" "$AGENT_DOC" "brag mcp install"

# T5 — gives the manual mcpServers JSON snippet
assert_contains_literal "T5" "$AGENT_DOC" '{"mcpServers":{"brag":{"command":"brag","args":["mcp","serve"]}}}'

# T6 — states the client-startup-reconnect note
if [ ! -f "$AGENT_DOC" ]; then
    fail "T6" "$AGENT_DOC does not exist"
else
    has_startup=no; has_reconnect=no
    if grep -F -q -- "connect at client startup" "$AGENT_DOC"; then has_startup=yes; fi
    if grep -F -q -- "reconnect" "$AGENT_DOC"; then has_reconnect=yes; fi
    if [ "$has_startup" = yes ] && [ "$has_reconnect" = yes ]; then
        ok "T6"
    else
        fail "T6" "reconnect note (startup=$has_startup reconnect=$has_reconnect)"
    fi
fi

# T7 — states the project-not-auto-filled gotcha (names the field)
if [ ! -f "$AGENT_DOC" ]; then
    fail "T7" "$AGENT_DOC does not exist"
else
    has_phrase=no; has_field=no
    if grep -F -q -- "does not auto-fill" "$AGENT_DOC"; then has_phrase=yes; fi
    if grep -F -q -- "project" "$AGENT_DOC"; then has_field=yes; fi
    if [ "$has_phrase" = yes ] && [ "$has_field" = yes ]; then
        ok "T7"
    else
        fail "T7" "gotcha (phrase=$has_phrase field=$has_field)"
    fi
fi

# T8 — cross-links the fix
assert_contains_literal "T8" "$AGENT_DOC" "brag project ensure"

# T9 — documents provenance stamping (all five reserved namespaces)
if [ ! -f "$AGENT_DOC" ]; then
    fail "T9" "$AGENT_DOC does not exist"
else
    t9_missing=""
    for tok in "agent:<name>" "model:<id>" "session:<id>" "cost:<n>" "tokens:<n>"; do
        if ! grep -F -q -- "$tok" "$AGENT_DOC"; then
            t9_missing="$t9_missing $tok"
        fi
    done
    if [ -z "$t9_missing" ]; then
        ok "T9"
    else
        fail "T9" "$AGENT_DOC missing provenance tokens:$t9_missing"
    fi
fi

# T10 — documents the DB resolution order
if [ ! -f "$AGENT_DOC" ]; then
    fail "T10" "$AGENT_DOC does not exist"
else
    t10_missing=""
    for tok in "--db" "BRAGFILE_DB" "~/.bragfile/db.sqlite"; do
        if ! grep -F -q -- "$tok" "$AGENT_DOC"; then
            t10_missing="$t10_missing $tok"
        fi
    done
    if [ -z "$t10_missing" ]; then
        ok "T10"
    else
        fail "T10" "$AGENT_DOC missing db-resolution tokens:$t10_missing"
    fi
fi

# T11 — carries the impact-framing convention (distinctive phrase)
assert_contains_literal "T11" "$AGENT_DOC" "a metric or a named outcome"

# T12 — gives the resolved per-client config paths
if [ ! -f "$AGENT_DOC" ]; then
    fail "T12" "$AGENT_DOC does not exist"
else
    t12_missing=""
    for tok in ".mcp.json" "~/.claude.json" ".cursor/mcp.json" "claude_desktop_config.json"; do
        if ! grep -F -q -- "$tok" "$AGENT_DOC"; then
            t12_missing="$t12_missing $tok"
        fi
    done
    if [ -z "$t12_missing" ]; then
        ok "T12"
    else
        fail "T12" "$AGENT_DOC missing config-path tokens:$t12_missing"
    fi
fi

# T13 — README links the page
assert_contains_literal "T13" "README.md" "docs/for-ai-agents.md"

# T14 — README has an agent/MCP section heading (line-regex avoids substring trap)
if [ ! -f README.md ]; then
    fail "T14" "README.md does not exist"
elif grep -E -q '^## .*[Aa]gent' README.md; then
    ok "T14"
else
    fail "T14" "README.md missing an agent/MCP '## ' section heading"
fi
```

---

## Build Completion

*Filled in at the end of the **build** cycle, before advancing to verify.*

- **Branch:**
- **PR (if applicable):**
- **All acceptance criteria met?** yes/no
- **New decisions emitted:**
  - `DEC-NNN` — <title> (if any)
- **Deviations from spec:**
  - [list]
- **Follow-up work identified:**
  - [any new specs for the stage's backlog]

### Build-phase reflection (3 questions, short answers)

Process-focused: how did the build go? What friction did the spec create?

1. **What was unclear in the spec that slowed you down?**
   — <answer>

2. **Was there a constraint or decision that should have been listed but wasn't?**
   — <answer>

3. **If you did this task again, what would you do differently?**
   — <answer>

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
