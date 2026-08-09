---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-074
  type: story                      # epic | story | task | bug | chore
  cycle: verify
  blocked: false
  priority: high
  complexity: M                    # S | M | L  (L means split it)

project:
  id: PROJ-006
  stage: STAGE-019
repo:
  id: bragfile

agents:
  architect: claude-opus-5
  implementer: claude-opus-5       # usually same Claude, different session
  created_at: 2026-08-08

references:
  decisions:
    - DEC-045                      # emitted by this spec — the push surface
    - DEC-024                      # the MCP server contract this extends (and amends)
    - DEC-043                      # the ranking the resources serve
    - DEC-044                      # the budget + line shape; DefaultBudget is the resource budget
    - DEC-014                      # the envelope the memory bodies render
    - DEC-017                      # entries.project is a soft string match
    - DEC-042                      # the MCP time vocabulary this tool does NOT extend
    - DEC-010                      # the search transform internal/ftsquery extracts
    - DEC-011                      # the JSON shape brag_memory --format json narrows
    - DEC-034                      # brag mcp install — how a client reaches these resources
  constraints:
    - no-sql-in-cli-layer
    - test-before-implementation
    - errors-wrap-with-context
    - no-new-top-level-deps-without-decision
    - no-cgo
    - stdout-is-for-data-stderr-is-for-humans
    - one-spec-per-pr
  related_specs:
    - SPEC-073                     # shipped — memory.Slice + export.ToMemory* + brag memory
    - SPEC-072                     # shipped — MCP filter parity + internal/timewindow precedent
    - SPEC-040                     # shipped — brag mcp serve + the four tools
    - SPEC-041                     # shipped — the plugin; "validated ≠ registered" lesson
    - SPEC-058                     # shipped — docs/for-ai-agents.md, the tool playbook
---

# SPEC-074: The MCP push surface — resources + `brag_memory`

## Context

The last spec in STAGE-019. When it ships, corpus-as-agent-memory is complete
and the stage closes.

SPEC-072 gave the corpus its filters. SPEC-073 gave it a blended,
token-budgeted slice (`internal/memory.Slice` + `export.ToMemoryMarkdown` /
`ToMemoryJSON` + `brag memory`). Both are **pull**: the agent must decide to
call something. That is the exact behavior STAGE-019 exists to replace — a
corpus an agent queries *if it thinks to* is not memory.

This spec is the **push** half. MCP **resources** are the protocol's only
auto-attachable primitive: context a client can put in front of a model
*before* the agent works, with no tool call and no configuration. Three
resources ship —

```
brag://memory/recent            the default slice          (static)
brag://memory/project/{name}    project-boosted slice      (template)
brag://projects                 the registered-project list (static)
```

— plus **`brag_memory`**, the parameterized pull counterpart and the server's
**fifth** tool. `brag://projects` is not a fourth feature: it is the template's
lookup table, so an agent fills `{name}` with a real project name instead of
inventing one (a guessed name produces a silently unboosted slice — a failure
with no error message).

**DEC-045** is emitted with this spec and carries the weight: the resource set,
the URI scheme, the MIME/rendering, the size policy, the cache semantics, the
soft-boost resolution, the pinned budget, the degenerate-ranking note, and the
tool's parameter set. Read it before this file's Failing Tests; it is not
re-litigated here.

**Two extractions ride along**, both firing the same argument (see LD8/LD9 and
DEC-045's structural Consequence). This spec is the third consumer of the
DEC-010 search transform, which is DEC-024's own revisit trigger (c); and the
resources plus the tool must reproduce `internal/cli/memory.go`'s three-read
pool composition *exactly*, because the parity property this surface claims
("the resource **is** the markdown") is only true if one composition and one
tokenizer run on both sides — not two that happen to agree today.

**This spec starts from go-sdk v1.7.0**, not from the original v1.6.1
pre-flight. PR #131 re-ran the resources pre-flight behaviorally and recorded
three deltas; a further targeted pre-flight was run at this spec's design for
the claims that go beyond it. Both are transcribed under
*Notes → §12(b) design-time pre-flight*.

## Goal

Ship the MCP push surface: three `brag://` resources serving byte-identical
`brag memory` markdown, plus the `brag_memory` tool — over one shared pool
composition (`memory.Gather`) and one shared query transform
(`internal/ftsquery`), with registration proven against a real `mcp.Client`.

## Inputs

- **Files to read (in this order):**
  - `decisions/DEC-045-mcp-push-surface-resources-and-brag-memory.md` — **read
    first.** Nine locked sub-decisions. Do not re-derive; every "why" this spec
    omits is there.
  - `decisions/DEC-024-mcp-server-sdk-transport-and-provenance.md` — the server
    contract being extended: CLI byte-parity, stdio purity, the SDK gate, and
    revisit triggers (b) and (c) which this spec fires. **Its "four typed tools"
    line is a historical record — do not rewrite it** (see premise audit).
  - `decisions/DEC-044-memory-slice-token-budget-and-line-shape.md` — read the
    **Consequences**, specifically the last `Neutral` bullet: `ToMemoryMarkdown`
    / `ToMemoryJSON` take a precomputed `memory.Result`, unlike the seven prior
    DEC-014 renderers. That note was written for this spec. **Do not reach for
    the entries+options renderer template.**
  - `decisions/DEC-043-memory-slice-blended-rank-fusion.md` — sub-decision 4
    (the soft boost, the source of DEC-045 sub-decision 6) and sub-decision 5
    (the three-read pool, the thing `memory.Gather` extracts).
  - `decisions/DEC-014-rule-based-output-shape.md` — the envelope. The resource
    bodies are its markdown form verbatim.
  - `internal/mcpserver/` — **all of it.** `server.go` (registration + the four
    handlers + `textResult`), `query.go` (`buildMatch`, whose comment names this
    spec), `provenance.go`, `server_test.go` (the harness: `newTestServer`,
    `callJSON`, `seedViaStore`, `setNowFunc` — and the SPEC-072 enumeration
    comment above `TestServer_ToolsListed`, written for this spec to consume),
    `transport_test.go`, `import_audit_test.go`.
  - `internal/cli/memory.go` — how the three reads compose the pool. The
    resources and the tool must do the same; LD9 makes "the same" mean *the same
    function*.
  - `internal/memory/memory.go` — `Slice`, `Options`, `Result`, the locked
    constants (`DefaultBudget`, `PoolLimit`), and the package doc's purity claim
    that LD9 must not break.
  - `internal/export/memory.go` — `MemoryOptions`, `ToMemoryMarkdown`,
    `ToMemoryJSON`. The resource body is `ToMemoryMarkdown`'s output, unmodified.
  - `internal/export/project.go` — where `ToProjectsMarkdown` lands, next to the
    existing project serializers.
  - `internal/timewindow/` — the extraction precedent for LD8 (same stage, same
    shape: a small pure package both `cli` and `mcpserver` import).
  - `internal/cli/search.go` — `buildFTS5Query`, the other half of the LD8
    extraction.
  - `internal/storage/project.go` — `ProjectStatuses()` (non-archived, with
    `BragCount`, ordered `updated_at DESC, id DESC`) and the `ProjectStatus`
    struct.
  - `internal/storage/storagetest/storagetest.go` — `Backdate(dbPath, id, at)`,
    the only sanctioned way for a non-storage test to seed past-dated rows.
- **External APIs:** the go-sdk **v1.7.0** resource surface —
  `Server.AddResource`, `Server.AddResourceTemplate`, `ResourceHandler`,
  `ReadResourceResult`/`ResourceContents`, `mcp.ResourceNotFoundError`,
  `jsonrpc.CodeInvalidParams`. No network. No new module.
- **Related code paths:** `internal/aggregate/` (untouched), `plugin/`
  (README only — `.mcp.json` needs no change; resources are advertised by the
  server, not the manifest).

## Outputs

> **⚠ LD8 and LD9 ALREADY LANDED — do not build them again.** Both extractions
> shipped ahead of this spec as standalone refactors (PR **#135**, plus **#136**
> completing two missed items), for the reason recorded in the stage: bundled,
> a reviewer would have to judge *"did this refactor preserve behaviour?"* and
> *"is this new surface correct?"* in one diff. Every entry below marked
> **✅ LANDED** is already on `main` — verify it rather than write it. The
> remaining entries are this spec's actual work.
>
> What landed: `internal/ftsquery` (`Build`), `internal/memory/pool.go`
> (`Source`, `GatherOptions`, `Pool`, `Gather`, `ErrQuery`), both original
> helpers deleted, all call sites retargeted, and `internal/cli/memory.go`
> reduced to a single `memory.Gather` call. Mutation-checked at two layers:
> reversing the bm25 order inside `Gather` reddens both the new
> `TestGather_QueryAddsMatchedInSearchOrder` and SPEC-073's
> `TestMemoryCmd_JSONScoresReflectFTS5MatchOrder`; dropping the project read
> reddens both `TestGather_ProjectAddsScopedRead` and
> `TestMemoryCmd_ThreeReadsComposeThePool`.
>
> **One deviation from LD9 as specified:** `Gather` returns a distinct
> `*memory.ErrQuery` for a malformed query, rather than a bare error the caller
> string-matches. LD9 required the caller to surface it as a USER error on both
> surfaces; a typed error is how that becomes checkable (`errors.As`) instead of
> conventional. `internal/cli/memory.go` already does this.

- **Files created:**
  - `internal/ftsquery/ftsquery.go` — **✅ LANDED.** the extracted DEC-010 transform. Exports
    `Build(raw string) (string, error)`. Stdlib only (`fmt`, `strings`); imports
    nothing internal.
  - `internal/ftsquery/ftsquery_test.go` — **✅ LANDED.** the union of the two deleted test
    sets (see planned deletions).
  - `internal/memory/pool.go` — **✅ LANDED.** the extracted three-read pool composition.
    Exports `Source`, `GatherOptions`, `Pool`, `Gather`.
  - `internal/memory/pool_test.go` — **✅ LANDED.** read-count/wiring tests against a fake
    `Source`, plus the "unknown project is not an error" case.
  - `internal/mcpserver/resources.go` — `addResources(srv, s)`, the three
    handlers, the URI constants, and the `{name}` extraction.
  - `internal/mcpserver/resources_test.go` — the §12(b) registration harness and
    the resource behavior tests.
  - `internal/mcpserver/memory.go` — `memoryIn`, `handleMemory`, and the shared
    `renderMemory` helper both the tool and the resource handlers call.
  - `internal/mcpserver/memory_test.go` — the tool's tests.
  - `decisions/DEC-045-mcp-push-surface-resources-and-brag-memory.md` —
    **already written at design** (emitted with this spec).
- **Files modified:**
  - `internal/mcpserver/server.go` — `New` registers `brag_memory` (after
    `brag_stats`) and calls `addResources`; the `New` doc comment goes
    **four → five** and gains a resources clause; `handleSearch` calls
    `ftsquery.Build`.
  - `internal/mcpserver/provenance.go` — the package doc comment's tool list
    (line 2) gains `brag_memory`.
  - `internal/mcpserver/server_test.go` — `TestServer_ToolsListed`'s `want`
    gains `"brag_memory"` in **third** position (`brag_add` < `brag_list` <
    `brag_memory` < `brag_search` < `brag_stats`); the enumeration comment
    above it is rewritten to name the *five* tools and to point at the resources;
    line 243's `buildMatch("latency")` becomes `ftsquery.Build("latency")`.
  - `internal/mcpserver/transport_test.go` — "drive all four tools" → five, and
    the stdout-purity round-trip drives `brag_memory` **and** a `resources/read`
    (the new transport path that could leak bytes).
  - `internal/cli/search.go` — **✅ LANDED.** `buildFTS5Query` **deleted**; `runSearch` calls
    `ftsquery.Build`.
  - `internal/cli/search_test.go` — **✅ LANDED.** the six `buildFTS5Query` unit tests
    (lines ~16–72) **deleted** (moved to `internal/ftsquery`).
  - `internal/cli/memory.go` — **✅ LANDED.** the inline three-read composition replaced by one
    `memory.Gather` call; the `buildFTS5Query` call goes with it.
  - `internal/cli/memory_test.go` — **✅ LANDED.** line ~476's comment names `buildFTS5Query`;
    retarget to `ftsquery.Build`. **The test itself is unchanged** (it asserts
    CLI behavior, not the helper's name).
  - `internal/cli/mcp.go` — the `Long` (line ~26) naming the four tools.
  - `internal/export/project.go` — `ToProjectsMarkdown([]storage.ProjectStatus)
    ([]byte, error)`.
  - `internal/export/project_test.go` — the byte-exact `ToProjectsMarkdown`
    goldens (populated + empty).
  - `docs/api-contract.md` — the `brag mcp serve` section (line ~1232): four →
    five tools, a `brag_memory` bullet, and a new **Resources** subsection; plus
    one DEC-index row for DEC-045 (after the DEC-044 row, line ~1383).
  - `docs/for-ai-agents.md` — §3 intro (line 65) four → five; a new
    `### brag_memory` subsection after `### brag_stats`; and a new
    **`## 4. The resources`** section before the current §4, renumbering
    §4→§5 … §7→§8.
  - `docs/architecture.md` — the `internal/mcpserver` row (line 86): tool list
    + a resources clause; the `internal/memory` row (line 85): `Gather` and the
    `internal/ftsquery` import; **one new row** for `internal/ftsquery`.
  - `docs/tutorial.md` — the `brag mcp serve` paragraph (lines ~1034–1035): five
    tools + one sentence on resources.
  - `README.md` — line 17's "four typed tools" → five.
  - `BRAG.md` — line 214's plugin tool list.
  - `plugin/README.md` — lines 6–7's tool list.
  - `docs/launch/copy-pack.md` — line 80's MCP directory blurb read-tool list.
  - `scripts/test-docs.sh` — group **T3**'s loop gains `brag_memory`; a new
    group **V** for the resources docs.
  - `AGENTS.md` §11 — the `memory` entry's `buildFTS5Query` reference →
    `ftsquery.Build`, plus a resources clause; and a new **`MCP resources`**
    glossary entry.
  - `projects/PROJ-006-agent-native-depth-core/stages/STAGE-019-corpus-as-agent-memory.md`
    — backlog status line for SPEC-074.
- **Planned deletions** (inversion case — the premise of these tests is that the
  transform lives in a command package; LD8 inverts it):
  - `internal/mcpserver/query.go` and `internal/mcpserver/query_test.go` —
    deleted outright. `TestBuildMatch`'s three OK cases and three error cases
    move to `internal/ftsquery/ftsquery_test.go`.
  - `internal/cli/search_test.go`'s `buildFTS5Query` tests — six functions,
    deleted; their cases (`latency`, `cut latency`, `auth-refactor`, `""`,
    whitespace-only, quote-containing) move to the same place. The union is
    **six distinct inputs**; the two suites overlap on all six, so the merged
    table has six rows and loses no coverage.
- **New exports:** `ftsquery.Build`; `memory.Source`, `memory.GatherOptions`,
  `memory.Pool`, `memory.Gather`; `export.ToProjectsMarkdown`.
- **Removed exports:** none (both deleted helpers were unexported).
- **Database changes:** **none.** No migration, no schema change, no new storage
  method, no new SQL. `ProjectStatuses` already exists.
- **`go.mod`:** untouched. `github.com/modelcontextprotocol/go-sdk/jsonrpc` is a
  package of the module DEC-024 already gated — not a new dependency.

### Premise audit — run at design, hits reconciled against the list above

Per `projects/_templates/premise-audit.md`. Every grep below was **executed**
against the repo at design; the "actual hits" column is real output. All greps
exclude `.claude/worktrees/` (a live git worktree checkout of the
`chore/spec-074-resources-preflight-sdk-1-7-0` branch — `git worktree list`
confirms; its copies of these files are not live sites) and `projects/` +
`decisions/` where noted.

| case | grep | actual hits | disposition |
|---|---|---|---|
| **additive** (MCP tool count) | `grep -rniE "four (typed )?tools\|brag_add.*brag_list.*brag_search.*brag_stats" --include='*.go' --include='*.md' --include='*.sh' --include='*.json' .` | `internal/cli/mcp.go:26`; `internal/mcpserver/server.go:28,29`; `internal/mcpserver/provenance.go:2`; `internal/mcpserver/server_test.go:90,102`; `internal/mcpserver/transport_test.go:21`; `docs/architecture.md:86`; `docs/for-ai-agents.md:65`; `docs/api-contract.md:1232`; `docs/tutorial.md:1034,1035`; `docs/launch/copy-pack.md:80`; `README.md:17`; `BRAG.md:214`; `plugin/README.md:6,7`; `scripts/test-docs.sh:1023,1028`; `decisions/DEC-024:50,51,94,103,163,189` | **14 live sites, all enumerated in Outputs above.** The cross-check surfaced **four the stage brief did not list** — `internal/cli/mcp.go:26`, `internal/mcpserver/provenance.go:2`, `README.md:17`, `docs/launch/copy-pack.md:80` — all added. `decisions/DEC-024` is a **historical record: NOT rewritten** (DEC-045 is the amendment and says so in its Consequences). |
| **additive** (test-docs T3) | `grep -n "T3" -A 20 scripts/test-docs.sh` | `:1023` comment "names all four MCP tools"; `:1028` `for tool in brag_add brag_list brag_search brag_stats` | Planned: the loop gains `brag_memory`, the comment says five. This is the group that **fails** if `docs/for-ai-agents.md` is not updated — it is the mechanical guard on the doc sweep. |
| **additive** (DEC-014 consumer ordinal) | `grep -rn "DEC-014 consumer" docs/ AGENTS.md README.md BRAG.md internal/` | `docs/api-contract.md:444,522,612,692,867,1379,1380,1382,1383`; `AGENTS.md:297,300,301`; `internal/cli/{summary,review,stats,coverage,wrapped,spark,memory}.go`; `internal/memory/memory.go:36` | **No ordinal changes.** `brag_memory` renders the **eighth** consumer's envelope over a different transport; it is not a ninth. Stated explicitly in DEC-045's Consequences so a future reader does not "fix" the chain. Only the DEC-index list gains a row (for DEC-045, which is not a DEC-014 consumer row). |
| **additive** (docs package table) | `grep -n "internal/memory\|internal/timewindow\|internal/spark" docs/architecture.md` | `docs/architecture.md:85` (`internal/memory` row), `:40` (mermaid diagram); the table also lacks rows for `spark`/`timewindow`/`story`/`capture` | Planned: **one new row** for `internal/ftsquery` + edits to the `internal/memory` and `internal/mcpserver` rows. **Stays:** the mermaid diagram at `:40` (depicts the STAGE-004 digest path) and the table's pre-existing omissions (a separate chore — do **not** widen; SPEC-073 dispositioned this identically). |
| **status change** | `grep -rniE "brag://\|resources/list\|resources/read\|MCP resource" --include='*.go' --include='*.md' --include='*.sh' --include='*.json' .` (excl. `.claude/`, `projects/`, `decisions/`) | `docs/research/proj-005-idea-synthesis.md:44,46,50` only | Nothing in the shipped docs claims resources exist or do not exist, so there is **no "not yet shipped" row to strike**. The research doc is a dated artifact — **stays**. Purely additive; every resources mention is new prose. |
| **inversion / removal** | `grep -rn "buildFTS5Query\|buildMatch" --include='*.go' internal/` | `internal/cli/search.go:39,43,63`; `internal/cli/search_test.go:16,19,30,41,52,62,69`; `internal/cli/memory.go:97`; `internal/cli/memory_test.go:476`; `internal/mcpserver/query.go:*`; `internal/mcpserver/query_test.go:5,14,16,20,21`; `internal/mcpserver/server.go:281`; `internal/mcpserver/server_test.go:243` | LD8 **inverts the premise** of every test that calls the helper by its package-private name. Planned deletions are enumerated under Outputs: `query.go` + `query_test.go` deleted, `search_test.go`'s six tests deleted, both merged into `internal/ftsquery/ftsquery_test.go`. Call sites retargeted, comments at `cli/memory_test.go:476` and `AGENTS.md:301` retargeted. **No test is silently lost** — the merged table's six rows are listed in Failing Tests. |
| **inversion** (purity claim) | `grep -n "internal/memory" docs/architecture.md AGENTS.md` | `docs/architecture.md:85` — *"No SQL, no clock, no dependency outside `internal/storage`"*; `AGENTS.md:301` | LD9 adds `internal/ftsquery` to `internal/memory`'s import set, and `pool.go` calls `Store` methods through an interface. **Both claims become stale.** Planned rewrites listed under Outputs; the *mechanical* guards (`TestPackageReadsNoWallClock`, `TestPackageEmitsNoReservedTagNamespace`) are **unaffected** — `pool.go` reads no clock and quotes no reserved prefix — and a Failing Test asserts they still pass over the enlarged package. |
| **NOT-contains** | see the self-audit table in *Notes → NOT-contains self-audit* | zero hits in load-bearing prose | Three absence assertions (`brag_memory` has no `since`/`until`/`day`; the resource `Description`s must not say "filter" of the project boost except in the negating phrase), each grep'd against the locked literals at design. |

```
Premise audit (projects/_templates/premise-audit.md), run at design:
- [x] Inversion/removal: greps run — LD8 inverts two test suites; deletions enumerated
- [x] Addition/count-bump: greps run — 14 live tool-count sites + test-docs T3 loop
- [x] Status-change: greps run — one dated-artifact hit, dispositioned "stays"
- [x] Cross-check: actual grep hits reconciled against ## Outputs — 4 sites the
      stage brief omitted were surfaced by the grep and ADDED to Outputs
```

## Locked design decisions

DEC-045 carries the reasoning; this list is the implementable form. Each has at
least one Failing Test that fails without it (AGENTS.md §9 traceability — the
mapping is in *Notes → Decision↔test traceability*).

**LD1 — Three resources, exact URIs, all `text/markdown`.**
`brag://memory/recent` (static), `brag://memory/project/{name}` (template),
`brag://projects` (static). `MIMEType: "text/markdown"` on all three
registrations and on every `ResourceContents` the handlers return.

**LD2 — `Resource.Size` is left at its zero value on all three.** It is
`omitempty`, so `size` is absent from the wire.

**LD3 — `TTLMs: 0` explicitly on every read result; `CacheScope` is not set.**
The server clobbers `CacheScope` to `"public"` regardless (observed). A test
pins both — `ttlMs == 0` as *chosen*, `cacheScope == "public"` as *observed*.

**LD4 — The project template is a SOFT boost.** `Options.Project` is set;
nothing is filtered. Three corrections ride with it:
`Title: "Memory slice, boosted toward a project"` on the template;
`Description` containing the exact phrase **`a soft boost, not a filter`**
(byte-identical to `internal/cli/memory.go`'s shipped flag help, so drift
between the two is a test failure); and the body's own `Filters:` line.
An **unknown** project name returns a slice; an **empty** name returns
`mcp.ResourceNotFoundError(uri)`.

**LD5 — Every resource is pinned to `memory.DefaultBudget`.** The constant, not
a literal `2000`.

**LD6 — `brag_memory` params are exactly `query`, `project`, `budget`,
`format`.** `Budget` is a **`*int`**: `nil` → `memory.DefaultBudget`, `0` or
negative → tool error (mirroring DEC-044's `UserError`). `Format` defaults to
`"markdown"`; `"json"` is accepted; anything else is a tool error. **No
`since`/`until`/`day`.**

**LD7 — The `{name}` segment is percent-decoded** with `url.PathUnescape` before
use. A decode failure is `mcp.ResourceNotFoundError(uri)`.

**LD8 — `internal/ftsquery`.** The DEC-010 transform moves to a new pure package
exporting `Build(raw string) (string, error)`. `cli.buildFTS5Query` and
`mcpserver.buildMatch` are **deleted**; three call sites retarget
(`cli/search.go`, `cli/memory.go` — via LD9 — and `mcpserver/server.go`).
Precedent: `internal/timewindow` (SPEC-072/DEC-042), same stage.

**LD9 — `memory.Gather`.** The three-read pool composition moves out of
`internal/cli/memory.go` into `internal/memory/pool.go`:

```go
// Source is the read surface Gather needs. *storage.Store satisfies it.
type Source interface {
    List(f storage.ListFilter) ([]storage.Entry, error)
    Search(query string, limit int) ([]storage.Entry, error)
}

type GatherOptions struct {
    Query   string // raw user query; "" skips the relevance read
    Project string // "" skips the project read
}

type Pool struct {
    Entries []storage.Entry
    Matched []int64
}

// Gather performs DEC-043 sub-decision 5's one-bounded-read-per-list
// composition and returns the union pool plus the bm25 id order.
// A malformed query is returned as an error and must be surfaced as a
// USER error by the caller (a UserError on the CLI, a tool error over MCP).
func Gather(src Source, opts GatherOptions) (Pool, error)
```

`internal/memory` gains one internal import (`internal/ftsquery`) and a file
that performs I/O through an interface. `Slice` stays pure and both mechanical
guards stay green.

**LD10 — `export.ToProjectsMarkdown`** over `Store.ProjectStatuses()`.
Locations, archived projects, and `state_note` are excluded (DEC-045
sub-decision 3). Literal body locked in *Notes*.

### Rejected alternatives (build-time)

Enumerated so they are not rediscovered as "either is fine":

1. **One static resource per registered project** instead of a template — the
   resource list grows without bound and goes stale on every
   `brag project ensure`.
2. **Duplicating the pool composition in `mcpserver`** with a parity test —
   this is the mistake `buildMatch` already made, at a larger scale, in the same
   PR that fixes it.
3. **Serving `brag project list`'s bytes** (TSV or `--format json`) for
   `brag://projects` — the TSV is a shell pipe and the JSON is 7 keys of which
   four are noise for name selection. DEC-045 sub-decision 3.
4. **Hand-writing the read-result JSON** to force `cacheScope: "private"` — a
   fork of the protocol layer to change a field with no reachable consequence
   over stdio. DEC-045 sub-decision 5.
5. **`Budget int` with `0` → default** — indistinguishable from an omitted
   field, and it would give `0` opposite meanings on the CLI (`UserError`) and
   over MCP (default) for the same word.
6. **A `brag://memory/query/{q}` template** — a query is a per-ask parameter;
   that is what the tool is for.
7. **Making the project resource a hard filter** — DEC-045 sub-decision 6
   rejects it on three independent grounds. If build feels the pull, re-read
   DEC-017.
8. **Populating `Resource.Size` with the `DefaultBudget × CharsPerToken` bound**
   — honest but ~10× pessimistic; a host budgeting on it may decline the
   auto-load. DEC-045 sub-decision 4.

## Acceptance Criteria

- [ ] `resources/list` advertises **exactly** `brag://memory/recent` and
      `brag://projects`; `resources/templates/list` advertises **exactly**
      `brag://memory/project/{name}`.
- [ ] All three carry `mimeType: "text/markdown"`, a non-empty `Description`,
      and **no `size`** on the wire.
- [ ] `resources/read` round-trips all three against a real `mcp.Client` over
      `mcp.NewInMemoryTransports()`, and the template **matches a concrete URI**
      — registration, not shape (§12(b); SPEC-041's lesson).
- [ ] `brag://memory/recent`'s body is **byte-identical** to
      `export.ToMemoryMarkdown` over `memory.Slice(memory.Gather(store, {}),
      Options{Budget: memory.DefaultBudget})` on the same store.
- [ ] `brag://memory/project/orbit`'s body is byte-identical to the same
      computation with `Project: "orbit"` — and **contains entries whose project
      is not `orbit`** (the soft boost, proven positively).
- [ ] `brag://memory/project/nosuchthing` returns a **non-empty slice**, not an
      error and not an empty body.
- [ ] `brag://memory/project/` (empty name) is an error whose unwrapped
      `*jsonrpc.Error` code is `jsonrpc.CodeInvalidParams` (**-32602**).
- [ ] `brag://memory/project/my%20project` reaches the handler and boosts the
      project literally named `my project`.
- [ ] `brag://memory/project/a/b` is not found (RFC 6570 `{name}` does not cross
      `/`), and `brag://unregistered` is not found — both `-32602`.
- [ ] Every read result carries `ttlMs == 0` and `cacheScope == "public"`.
- [ ] `brag://projects` lists non-archived projects with status and brag count,
      **omits locations**, and its empty state names `brag project ensure`.
- [ ] `tools/list` returns **exactly five** tools:
      `brag_add`, `brag_list`, `brag_memory`, `brag_search`, `brag_stats`.
- [ ] `brag_memory` with no arguments is byte-identical to `brag memory` at the
      default budget; with `{"format":"json"}` byte-identical to
      `brag memory --format json`.
- [ ] `brag_memory {"budget": 0}` and `{"budget": -1}` are tool errors;
      `{"budget": 500}` is honoured; an omitted `budget` uses
      `memory.DefaultBudget`.
- [ ] `brag_memory {"format":"xml"}` and a malformed `{"query":"has\"quote"}`
      are tool errors.
- [ ] `brag_memory`'s input schema exposes **no** `since`, `until`, or `day`
      property.
- [ ] `ftsquery.Build` produces byte-identical output to the two deleted
      helpers on all six merged inputs; no `buildFTS5Query`/`buildMatch` symbol
      remains in the tree.
- [ ] `memory.Gather` performs exactly one `List` when neither query nor project
      is given, and exactly three reads when both are.
- [ ] `internal/cli/memory.go` and `internal/mcpserver` both compose the pool via
      `memory.Gather` — no second composition exists.
- [ ] `internal/memory`'s two mechanical guards still pass over the enlarged
      package (`TestPackageReadsNoWallClock`,
      `TestPackageEmitsNoReservedTagNamespace`).
- [ ] A full five-tool + `resources/read` round-trip writes **nothing** to
      `os.Stdout`.
- [ ] `internal/mcpserver` imports no `database/sql` (existing
      `import_audit_test.go`, which walks test files too — use
      `internal/storage/storagetest` if a test needs raw SQL).
- [ ] `go.mod` is unchanged. No CGO, no network, no migration.
- [ ] Full gate set green: `gofmt -l .` empty, `go vet ./...`, `go test ./...`,
      `just test-docs`.

## Failing Tests

Written during **design**, BEFORE build. Every locked decision above has at
least one test here that fails without it.

> **Coverage-claim discipline (SPEC-073 ship reflection).** A sentence saying
> "this test pins X and Y" is aspirational until each clause is
> mutation-checked. The *Notes → Mutation-check table* lists, for every coverage
> claim below, the exact mutation and the test that must redden. **Run each
> mutation at build and confirm the expected test fails, before writing the
> claim into `## Build Completion`.**

### `internal/ftsquery/ftsquery_test.go` (new — absorbs two deleted suites)

- `"TestBuild_Table"` — a six-row table, the **union** of the deleted
  `cli.buildFTS5Query` and `mcpserver.buildMatch` suites:

  | input | want | want error |
  |---|---|---|
  | `latency` | `"latency"` | no |
  | `cut latency` | `"cut" "latency"` | no |
  | `auth-refactor` | `"auth-refactor"` | no |
  | `""` | — | yes |
  | `"   \t  "` | — | yes |
  | `with "quote"` | — | yes |

  (Expected values transcribed from the two deleted suites, not retyped from
  the algorithm — §12(a) expected-value-literals.)
- `"TestPackageImportsNothingInternal"` — walks the package's non-test `.go`
  files and asserts none contains `github.com/jysf/bragfile000/internal/`.
  **Pins LD8's "pure leaf package"**: if `ftsquery` ever imports back into a
  command package the cycle it was created to break returns.

### `internal/memory/pool_test.go` (new)

Uses a `fakeSource` recording each call (no DB — `internal/memory` must stay
driver-free).

- `"TestGather_NoQueryNoProject_OneRead"` — `GatherOptions{}` → exactly **one**
  `List` call, with `ListFilter{Limit: memory.PoolLimit}`; zero `Search` calls;
  `Pool.Matched` is empty.
- `"TestGather_QueryAndProject_ThreeReads"` — `{Query:"auth", Project:"orbit"}` →
  `List{Limit:PoolLimit}`, `Search(` + "`\"auth\"`" + `, PoolLimit)`, and
  `List{Project:"orbit", Limit:PoolLimit}` — **three** reads, in that order.
  Asserts the `Search` argument is `ftsquery.Build`'s output, not the raw query.
- `"TestGather_MatchedCarriesSearchOrder"` — a `Search` returning ids
  `[5 1 2 4]` yields `Pool.Matched == []int64{5,1,2,4}` **in that order**
  (DEC-043's caller contract: the order *is* the signal).
- `"TestGather_UnionIncludesAllThreeReads"` — entries returned only by the
  project read, and only by the search read, are both present in `Pool.Entries`.
- `"TestGather_MalformedQueryIsAnError"` — `{Query: "has\"quote"}` returns an
  error and performs **zero** reads (validation precedes I/O).
- `"TestGather_UnknownProjectIsNotAnError"` — a project read returning zero rows
  yields a non-error `Pool` whose `Entries` still carry the recency read's rows.
  **Pins the soft boost at the composition layer** (LD4), one level below the
  resource.
- `"TestPackageReadsNoWallClock"` / `"TestPackageEmitsNoReservedTagNamespace"` —
  **existing**, in `memory_test.go`; they walk non-test files and must stay
  green now that `pool.go` exists. Not modified; listed because LD9 puts them at
  risk.

### `internal/export/project_test.go` (modified)

- `"TestToProjectsMarkdown_Golden"` — three `ProjectStatus` rows → **byte-exact**
  match against *Notes → GOLDEN 3*. Asserts by full-body equality, so a stray
  field (a location, a `state_note`) is a diff.
- `"TestToProjectsMarkdown_Empty"` — `nil` input → **byte-exact** match against
  *Notes → GOLDEN 4*, which names `brag project ensure`.

### `internal/mcpserver/resources_test.go` (new — the §12(b) harness)

All use `newTestServer` (the existing harness). Seeding uses `seedViaStore` +
`storagetest.Backdate` where dates matter.

**Registration (the SPEC-041 bar: registration, not shape).**

- `"TestResources_Listed"` — `cs.ListResources` returns **exactly**
  `["brag://memory/recent", "brag://projects"]` (sorted). Asserts each carries
  `MIMEType == "text/markdown"`, a non-empty `Description`, and
  **`Size == 0`** (LD2).
- `"TestResourceTemplates_Listed"` — `cs.ListResourceTemplates` returns
  **exactly** `["brag://memory/project/{name}"]`, `MIMEType ==
  "text/markdown"`, and a `Description` containing the literal
  `a soft boost, not a filter` (LD4 correction 2).
- `"TestResourceTemplate_MatchesConcreteURI"` — reading
  `brag://memory/project/bragfile` succeeds and the returned
  `Contents[0].URI` is the **concrete** URI. This is the assertion PR #131's
  re-run existed to make and the one a shape validator cannot make.
- `"TestResources_ReadRoundTripsAllThree"` — each of the three URIs reads
  successfully with exactly one `ResourceContents` whose `MIMEType` is
  `text/markdown` and whose `Text` is non-empty.

**Body parity (LD1, LD5).**

- `"TestResourceRecent_IsByteIdenticalToMemoryMarkdown"` — seeds a store, then
  computes the expected body **in the test** as
  `export.ToMemoryMarkdown(memory.Slice(pool, memory.Options{Budget: memory.DefaultBudget}), export.MemoryOptions{Filters: "(none)", FiltersJSON: map[string]string{}, Now: <frozen>})`
  and asserts equality with the resource body. The options are **hard-coded in
  the test**, not read back from the handler, so a wrong budget or a stray
  filter is a diff. `nowFunc` is frozen via `setNowFunc` so the `Generated:`
  line matches.
- `"TestResourceRecent_UsesDefaultBudget"` — the body's `## Budget` section
  contains `- Budget: 2000 tokens`, and the assertion is written against
  `memory.DefaultBudget` (formatted), **not** a bare literal — so raising the
  constant moves the test with it rather than breaking it (LD5).

**The soft boost, proven positively (LD4).**

- `"TestResourceProject_IncludesOtherProjects"` — seeds entries across `orbit`
  and `bragfile`; reads `brag://memory/project/orbit`; asserts the body contains
  **at least one `[bragfile/…]` line** and that the first slice line is an
  `[orbit/…]` one. This is the test that fails if someone converts the boost to
  a filter — the *only* one that fails for that reason in a way a passing suite
  would otherwise hide.
- `"TestResourceProject_UnknownNameStillReturnsASlice"` — reads
  `brag://memory/project/zzz-nonexistent`; asserts **no error** and a non-empty
  `## Slice` section. The filter-conversion twin: under a hard filter this
  returns empty.
- `"TestResourceProject_EchoesTheProjectInFilters"` — the body's `Filters:` line
  is exactly `--project orbit` (LD4 correction 3 — the body is
  self-describing).
- `"TestResourceProject_EmptyNameIsInvalidParams"` — reads
  `brag://memory/project/`; asserts the error unwraps via `errors.As` to a
  `*jsonrpc.Error` with `Code == jsonrpc.CodeInvalidParams`. **Must not
  reference the deprecated `mcp.CodeResourceNotFound`.**
- `"TestResourceProject_PercentDecodesTheName"` — seeds a project literally
  named `my project`; reads `brag://memory/project/my%20project`; asserts the
  `Filters:` line is `--project my project` (LD7).
- `"TestResourceProject_SlashInNameDoesNotMatch"` — reading
  `brag://memory/project/a/b` errors with `-32602` (the template does not cross
  `/`).
- `"TestResource_UnregisteredURIIsInvalidParams"` — `brag://nope` errors with
  `-32602`.

**Cache semantics (LD3).**

- `"TestResources_CacheHints"` — for each of the three URIs, asserts
  `result.TTLMs == 0` **and** `result.CacheScope == "public"`. The comment above
  it states plainly: *`ttlMs` is ours and chosen; `cacheScope` is the SDK's
  and observed — `Server.readResource` clobbers it unconditionally, so this
  clause reddening means the SDK changed, and DEC-045 sub-decision 5 should then
  set `"private"`.*

**The projects resource (LD10).**

- `"TestResourceProjects_ListsNamesStatusAndCount"` — seeds two projects and
  entries; asserts the body contains both names, both statuses, and the brag
  counts.
- `"TestResourceProjects_OmitsLocations"` — registers a project with a location
  `/tmp/somewhere-distinctive`; asserts the body does **not** contain that
  string (DEC-045 sub-decision 3's privacy exclusion).
- `"TestResourceProjects_ExcludesArchived"` — archives a project; asserts its
  name is absent.
- `"TestResourceProjects_EmptyState"` — a store with no projects; asserts the
  body contains `brag project ensure`.

### `internal/mcpserver/memory_test.go` (new — the fifth tool)

- `"TestMemoryTool_ByteIdenticalToCLIMarkdown"` — `brag_memory` with no
  arguments vs. the same test-side computation as
  `TestResourceRecent_IsByteIdenticalToMemoryMarkdown`. Pins DEC-024's parity
  contract at the new tool.
- `"TestMemoryTool_JSONFormat"` — `{"format":"json"}` returns bytes that
  `json.Unmarshal` into an object carrying `scope: "lifetime"`, `budget: 2000`,
  and a `slice` array; and `{"format":"markdown"}` returns the markdown body.
- `"TestMemoryTool_DefaultFormatIsMarkdown"` — with `format` omitted the result
  starts with `# Bragfile Memory` and is **not** valid JSON (LD6).
- `"TestMemoryTool_JSONAndMarkdownSelectTheSameSlice"` — the ids in the JSON
  `slice` equal the ids parsed off the markdown body's slice lines, in order.
  The MCP twin of SPEC-073's `TestMemoryCmd_JSONAndMarkdownSelectTheSameSlice`.
- `"TestMemoryTool_BudgetOmittedUsesDefault"` — omitted `budget` → body reports
  `- Budget: 2000 tokens` (asserted against `memory.DefaultBudget`).
- `"TestMemoryTool_BudgetZeroIsAnError"` / `"TestMemoryTool_BudgetNegativeIsAnError"`
  — `{"budget":0}` and `{"budget":-1}` return `IsError == true`. **This is the
  pair that fails if `Budget` is an `int` rather than a `*int`** — under a plain
  `int`, `{"budget":0}` is indistinguishable from omission and would succeed
  (LD6).
- `"TestMemoryTool_BudgetHonoured"` — `{"budget":40}` → `- Budget: 40 tokens`
  and `Skipped` > 0 over the seeded corpus.
- `"TestMemoryTool_ProjectIsASoftBoost"` — `{"project":"orbit"}` returns entries
  from other projects too (the tool-side twin of the resource assertion).
- `"TestMemoryTool_UnknownFormatIsAnError"` — `{"format":"xml"}` → `IsError`.
- `"TestMemoryTool_MalformedQueryIsAnError"` — `{"query":"has\"quote"}` →
  `IsError`.
- `"TestMemoryTool_SchemaHasNoTimeWindowParams"` — reads `brag_memory`'s
  `InputSchema` off `tools/list` and asserts its `properties` keys are
  **exactly** `{budget, format, project, query}` — so `since`, `until`, and
  `day` are absent by construction rather than by three separate NOT-contains
  checks (LD6, and SPEC-073's scope note preserved).

### `internal/mcpserver/server_test.go` (modified)

- `"TestServer_ToolsListed"` — `want` becomes
  `[]string{"brag_add", "brag_list", "brag_memory", "brag_search", "brag_stats"}`.
  The enumeration comment above it is rewritten: five tools, and a pointer to
  `resources_test.go` for the resource set.

### `internal/mcpserver/transport_test.go` (modified)

- `"TestServer_StdoutCarriesNoStrayBytes"` — the round-trip drives **five**
  tools and additionally performs a `resources/read` on each of the three URIs.
  The resource path is a new code path into the same stdio transport; without
  this it is unguarded.

### `scripts/test-docs.sh` — group **T3** (modified) and new group **V**

- **T3** — the loop gains `brag_memory`, so `docs/for-ai-agents.md` failing to
  document the fifth tool is a red build.
- **V1** — `docs/for-ai-agents.md` contains all three literal URIs
  (`brag://memory/recent`, `brag://memory/project/{name}`, `brag://projects`).
- **V2** — `docs/api-contract.md` contains `brag://memory/recent`.
- **V3** — `docs/for-ai-agents.md` contains the literal
  `a soft boost, not a filter` — the doc-level guard on LD4's correction, so the
  docs cannot describe the project resource as a scope.
- **V4** — `docs/architecture.md` contains `internal/ftsquery`.

## Implementation Context

*Read this section (and the files it points to) before starting the build
cycle.*

### Decisions that apply

- `DEC-045` — **emitted by this spec.** Nine sub-decisions; the source of every
  LD above. Read before the Failing Tests.
- `DEC-024` — the server contract: CLI byte-parity, stdio purity, the SDK gate.
  This spec fires its revisit trigger **(c)** (third consumer of the DEC-010
  transform → LD8) and extends its trigger **(b)** (DEC-045 sub-decision 5 adds
  `cacheScope` to what a networked transport must confront). Its "four typed
  tools" line stays as written.
- `DEC-043` — sub-decision 4 (soft boost → LD4), sub-decision 5 (three-read
  pool → LD9), sub-decision 3 (no clock; resources exercise the degenerate path
  → DEC-045 sub-decision 9).
- `DEC-044` — **read its Consequences.** `ToMemoryMarkdown`/`ToMemoryJSON` take
  a precomputed `memory.Result`, unlike the seven prior DEC-014 renderers; the
  last `Neutral` bullet was written for this spec. `DefaultBudget` is the pinned
  resource budget (LD5).
- `DEC-014` — the envelope the memory bodies render. Unchanged.
- `DEC-017` — `entries.project` is a soft string match. The reason a hard filter
  silently drops history (rejected alternative 7).
- `DEC-042` — the MCP time vocabulary. `brag_memory` deliberately does **not**
  extend it; also the `internal/timewindow` extraction precedent LD8 copies.
- `DEC-010` — the search transform LD8 extracts. Output shape unchanged.
- `DEC-011` — the JSON entry shape `brag_memory --format json` narrows (via
  `ToMemoryJSON`'s 9-key per-item projection).
- `DEC-034` — `brag mcp install`. Unchanged; it is how a client reaches these
  resources.

### Constraints that apply

- `no-sql-in-cli-layer` (blocking) — extended to `internal/mcpserver` by test
  convention. `import_audit_test.go` **walks test files too**; use
  `internal/storage/storagetest` if a test needs raw SQL. `internal/memory` gets
  no SQL either — `Gather` goes through the `Source` interface.
- `test-before-implementation` (blocking) — the tests above are written first.
  Run `go test ./...` once after writing them and confirm each fails for the
  *asserted* reason, not a compile error.
- `errors-wrap-with-context` (blocking) — handler errors wrap with the tool or
  URI name. Note the one exception: `mcp.ResourceNotFoundError(uri)` is returned
  **unwrapped**, because wrapping it would hide the `*jsonrpc.Error` from the
  client's `errors.As` and lose the `-32602` code.
- `no-new-top-level-deps-without-decision` (warning) — `go.mod` untouched.
  `.../go-sdk/jsonrpc` is a package of an already-gated module.
- `no-cgo`, `stdout-is-for-data-stderr-is-for-humans` (blocking) — the stdio
  stream carries protocol frames only; the resource path is newly guarded by
  `transport_test.go`.
- `one-spec-per-pr` (blocking) — **and check the whole branch, not just the
  delta.** SPEC-073's ship reflection: an unrelated commit rode its branch
  through three verify passes because every review was range-scoped. Run
  `git log main..HEAD` before handing off to verify.

### Prior related work

- `SPEC-073` (shipped, PR #132/#134) — `memory.Slice`, `export.ToMemory*`,
  `brag memory`. Its goldens cover this spec's resource bodies for free. Read
  its Reflection: the two rules it earned are applied here (the mutation-check
  table, and `git log main..HEAD`).
- `SPEC-072` (shipped) — `internal/timewindow`, the extraction template LD8
  copies verbatim in shape.
- `SPEC-040` (shipped) — the server and the four tools.
- `SPEC-041` (shipped) — the plugin. Its lesson is this spec's §12(b) bar: a
  manifest that *validated* still registered **zero** MCP servers, so the
  resource tests assert registration through a real client, never shape.
- `SPEC-058` (shipped) — `docs/for-ai-agents.md` and the `test-docs.sh` T-group
  this spec extends.
- `PR #124` (merged) — the go-sdk `1.6.1 → 1.7.0` bump.
- `PR #131` (merged) — the §12(b) re-run whose three deltas this spec starts
  from.

### Out of scope (for this spec specifically)

- **MCP prompts, sampling, subscriptions, `resources/updated` notifications.**
  Resources are read on demand. `ttlMs: 0` is the honest stand-in for
  change-notification (DEC-045 revisit (f)).
- **A query-carrying resource template.** Rejected alternative 6.
- **Time-window params on `brag_memory`.** LD6; `brag_list` already has them.
- **Writing memory back** (`brag_remember`, agent-authored summaries). The stage
  makes the corpus *readable* as memory; it adds no second write path.
- **Populating `Resource.Size`.** LD2; revisit trigger (b) in DEC-045.
- **Back-filling `docs/architecture.md`'s missing package rows** for
  `spark`/`timewindow`/`story`/`capture`, and the missing AGENTS.md §11 entries
  for `spark`/`story`. Pre-existing gaps; a docs chore, not this spec. (Same
  disposition SPEC-073 made.)
- **Rewriting DEC-024's "four tools" line.** It is a historical record.
- **`state_note` in `brag://projects`.** DEC-045 revisit (g).

## Notes for the Implementer

### §12(b) design-time pre-flight — what was run, and what it caught

Two pre-flights feed this spec. **Do not re-run either unless `go.mod` moves.**

**A. PR #131's re-run (recorded in STAGE-019's Design Notes).** Behavioral, at
v1.7.0, over `mcp.NewInMemoryTransports()` driven by a real `mcp.Client`. Three
API shapes held (`AddResource`, `AddResourceTemplate`,
`ReadResourceResult.Contents []*ResourceContents{URI, MIMEType, Text}`). Three
deltas this spec starts from: `ReadResourceResult` embeds `Cacheable` +
elicitation fields + a custom `MarshalJSON`; `CacheScope` is clobbered to
`"public"`; `CodeResourceNotFound` moved `-32002 → -32602` and is deprecated.

**B. This spec's targeted pre-flight (design, 2026-08-08, go-sdk v1.7.0).** Run
in a throwaway package (`preflight074/`, deleted after) for the claims that go
**beyond** A — a second static resource, `Resource.Size`, the error the *client*
actually sees, template edge cases, and `*int` schema inference. Verbatim
observations:

```
[capabilities]  resources: &{ListChanged:true Subscribe:false}

[resources/list]
  uri=brag://memory/recent   name=memory-recent  mime=text/markdown size=778
  uri=brag://projects        name=projects       mime=text/markdown size=0
  wire: {"resultType":"complete","_meta":{...},"ttlMs":0,"cacheScope":"public",
         "resources":[{...,"size":778,...},{... no "size" key ...}]}

[resources/templates/list]
  tmpl=brag://memory/project/{name}  name=memory-project  mime=text/markdown

[resources/read brag://memory/recent]        ttlMs=0     scope="public"
[resources/read brag://projects]             ttlMs=60000 scope="public"   <- handler set scope="private", TTL=60000
[read brag://memory/project/bragfile]        raw="bragfile"     unescape="bragfile"
[read brag://memory/project/PROJ-006]        raw="PROJ-006"     unescape="PROJ-006"
[read brag://memory/project/my%20project]    raw="my%20project" unescape="my project"
[read brag://memory/project/a/b]             ERR  Resource not found
[read brag://memory/project/]                raw=""             unescape=""       <- REACHES THE HANDLER
[read brag://nomime] handler left URI+MIMEType empty -> uri="brag://nomime" mime="text/markdown"

[error shape]
  err  = calling "resources/read": Resource not found
  type = *fmt.wrapError                       <- NOT a *jsonrpc.Error
  json.Marshal(err) = {}
  errors.As -> *jsonrpc.Error code=-32602 msg="Resource not found" data={"uri":"brag://unregistered"}
  jsonrpc.CodeInvalidParams = -32602 ; mcp.CodeResourceNotFound = -32602 (deprecated)

[precedence] a static resource whose URI also matches a template WINS (lookupResourceHandler tries resources first)

[*int schema] "budget":{"type":["null","integer"]}
  {}              -> Budget == nil
  {"budget":0}    -> Budget == *0     <- distinguishable from omitted
  {"budget":500}  -> Budget == *500
  {"budget":-1}   -> Budget == *-1
```

**Five findings changed the design:**

1. **The client-side error is a `*fmt.wrapError`, not a `*jsonrpc.Error`.** A
   type assertion fails and `json.Marshal` yields `{}`. Every error-path test
   **must** use `errors.As`. Had this not been checked, the obvious test
   (`err.(*jsonrpc.Error)`) would have compiled and failed at build.
2. **`brag://memory/project/` reaches the handler** with an empty segment — the
   SDK does not reject it. Without this, "empty name" would have been an
   unhandled case producing an unboosted slice under a URI promising a scope.
   LD4 makes it `-32602`.
3. **Percent-encoding is not decoded by the SDK.** `req.Params.URI` carries
   `%20` verbatim → LD7.
4. **`Resource.Size` is `omitempty` and does round-trip.** Leaving it zero means
   the key is simply absent from the wire, which is what makes LD2 clean rather
   than a `"size":0` that reads like a claim.
5. **`*int` schema inference works** (`["null","integer"]`), so LD6 can preserve
   DEC-044's `0`-is-an-error semantics across the transport instead of
   diverging.

**Not needed, confirmed:** the handler may leave `ResourceContents.URI` and
`.MIMEType` empty — `readResource` back-fills both from the request URI and the
registered resource. We set them explicitly anyway, so the contents are correct
independent of that convenience.

### The `Resource` / `ResourceTemplate` literals (transcribe verbatim)

```go
const (
    uriMemoryRecent  = "brag://memory/recent"
    uriProjects      = "brag://projects"
    tmplMemoryProject = "brag://memory/project/{name}"
    uriMemoryProjectPrefix = "brag://memory/project/"
)

&mcp.Resource{
    Name:        "memory-recent",
    Title:       "Recent memory slice",
    Description: "A ranked, token-budgeted slice of the developer's own work history — the corpus read back as working memory. Blended by recency (reciprocal-rank fusion, DEC-043) and trimmed to an estimated 2000-token budget. Load this before you start work: it is what the developer already knows, so you do not re-derive it. Markdown; the same bytes as `brag memory`.",
    MIMEType:    "text/markdown",
    URI:         uriMemoryRecent,
}

&mcp.ResourceTemplate{
    Name:        "memory-project",
    Title:       "Memory slice, boosted toward a project",
    Description: "The same memory slice, with entries in {name} ranked higher. {name} is a soft boost, not a filter: entries from other projects still appear, ranked lower, because a decision made elsewhere is often the one that matters here. Use `brag://projects` to get real project names. Markdown; the same bytes as `brag memory --project <name>`.",
    MIMEType:    "text/markdown",
    URITemplate: tmplMemoryProject,
}

&mcp.Resource{
    Name:        "projects",
    Title:       "Registered projects",
    Description: "Every registered, non-archived project with its status and brag count. Use these names verbatim for the `project` field on brag_add and brag_memory, and for the {name} slot of brag://memory/project/{name} — bragfile never auto-registers a name you invent.",
    MIMEType:    "text/markdown",
    URI:         uriProjects,
}
```

**`Size` is deliberately absent from all three** (LD2). Do not add it.

### GOLDEN 3 — `export.ToProjectsMarkdown` (populated)

Input: three `ProjectStatus` rows in `ProjectStatuses()` order
(`updated_at DESC, id DESC`):
`{Name:"bragfile", Status:"active", BragCount:137}`,
`{Name:"orbit", Status:"paused", BragCount:42}`,
`{Name:"atlas", Status:"active", BragCount:0}`.

```
# Bragfile Projects

Registered projects, most-recently-updated first. Use these names verbatim.

- bragfile — active, 137 brags
- orbit — paused, 42 brags
- atlas — active, 0 brags
```

Note `0 brags` — plural, no special-casing. Returned with the trailing newline
stripped, matching every other renderer.

### GOLDEN 4 — `export.ToProjectsMarkdown` (empty)

```
# Bragfile Projects

No projects registered yet. Register one with `brag project ensure <name>`.
```

### Decision↔test traceability (AGENTS.md §9)

| decision | test that fails without it |
|---|---|
| LD1 (set, URIs, MIME) | `TestResources_Listed`, `TestResourceTemplates_Listed`, `TestResources_ReadRoundTripsAllThree` |
| LD2 (no `Size`) | `TestResources_Listed` (`Size == 0`) |
| LD3 (cache hints) | `TestResources_CacheHints` |
| LD4 (soft boost + errors) | `TestResourceProject_IncludesOtherProjects`, `TestResourceProject_UnknownNameStillReturnsASlice`, `TestResourceProject_EmptyNameIsInvalidParams`, `TestResourceTemplates_Listed` (description phrase), `TestGather_UnknownProjectIsNotAnError` |
| LD5 (`DefaultBudget`) | `TestResourceRecent_UsesDefaultBudget`, `TestMemoryTool_BudgetOmittedUsesDefault` |
| LD6 (tool params) | `TestMemoryTool_BudgetZeroIsAnError`, `TestMemoryTool_DefaultFormatIsMarkdown`, `TestMemoryTool_SchemaHasNoTimeWindowParams`, `TestMemoryTool_UnknownFormatIsAnError` |
| LD7 (percent-decode) | `TestResourceProject_PercentDecodesTheName` |
| LD8 (`ftsquery`) | `TestBuild_Table`, `TestPackageImportsNothingInternal` |
| LD9 (`Gather`) | the five `TestGather_*` tests |
| LD10 (`ToProjectsMarkdown`) | `TestToProjectsMarkdown_Golden`, `TestToProjectsMarkdown_Empty`, `TestResourceProjects_OmitsLocations`, `TestResourceProjects_ExcludesArchived` |

### Mutation-check table — run these at build before writing any coverage claim

SPEC-073's ship reflection: *"a prose claim about what a test pins is
aspirational until each clause is mutation-checked."* Break each mutation,
confirm the named test reddens, revert.

| # | mutation | must redden |
|---|---|---|
| 1 | In the template handler, set `Options.Project` and also drop non-matching entries from the pool (make it a filter) | `TestResourceProject_IncludesOtherProjects` **and** `TestResourceProject_UnknownNameStillReturnsASlice` — **both**, independently |
| 2 | Pass a literal `2000` instead of `memory.DefaultBudget`, then change the constant to `2001` | `TestResourceRecent_UsesDefaultBudget`, `TestMemoryTool_BudgetOmittedUsesDefault` |
| 3 | Change `memoryIn.Budget` from `*int` to `int` with `0 → DefaultBudget` | `TestMemoryTool_BudgetZeroIsAnError` |
| 4 | Return the empty-name case as a slice instead of `ResourceNotFoundError` | `TestResourceProject_EmptyNameIsInvalidParams` |
| 5 | Skip `url.PathUnescape` on the `{name}` segment | `TestResourceProject_PercentDecodesTheName` |
| 6 | Set `Size: 8000` on the two static resources | `TestResources_Listed` |
| 7 | Set `TTLMs: 60000` | `TestResources_CacheHints` |
| 8 | Have the resource handler build its own `MemoryOptions` with `Filters: ""` | `TestResourceRecent_IsByteIdenticalToMemoryMarkdown` |
| 9 | Include locations in `ToProjectsMarkdown` | `TestToProjectsMarkdown_Golden`, `TestResourceProjects_OmitsLocations` |
| 10 | Swap `ProjectStatuses()` for `ListProjects()` | `TestResourceProjects_ExcludesArchived` |
| 11 | In `Gather`, drop the project read | `TestGather_QueryAndProject_ThreeReads`, `TestGather_UnionIncludesAllThreeReads` |
| 12 | In `Gather`, pass the raw query to `Search` instead of `ftsquery.Build`'s output | `TestGather_QueryAndProject_ThreeReads` |
| 13 | Add a `since` field to `memoryIn` | `TestMemoryTool_SchemaHasNoTimeWindowParams` |
| 14 | Register `brag_memory` but not the resources | `TestResources_Listed`, `TestResourceTemplates_Listed` |
| 15 | Wrap `mcp.ResourceNotFoundError(uri)` with `fmt.Errorf("read %s: %w", …)` | *(expected: nothing reddens — `errors.As` still unwraps.* **Confirm this, and if nothing reddens, leave the code unwrapped anyway and note it: the constraint exception in Implementation Context is then a style choice, not a correctness one.)* |

### NOT-contains self-audit (run at design; zero hits in load-bearing prose)

| assertion | grep target | result |
|---|---|---|
| `brag_memory` schema has no `since`/`until`/`day` | the locked `memoryIn` struct above | no such fields — the schema test asserts the exact key set, which is stronger than three absence checks |
| the resource `Description`s must not describe the project boost as a filter | the three `Description` literals above | `filter` appears **once**, inside `"a soft boost, not a filter"` — the negating phrase. No other occurrence. |
| V3's doc guard phrase | `internal/cli/memory.go:46`'s flag help | `"boost entries in this project (a soft boost, not a filter)"` — the phrase is byte-identical to the one in the template `Description`, so drift between CLI help and MCP description is a test failure |

### Ordering and other gotchas

- **Sorted tool position.** `brag_memory` sorts **third** in
  `TestServer_ToolsListed`'s `want`: `brag_add` < `brag_list` < `brag_memory` <
  `brag_search` < `brag_stats`. Registration order in `New` is separate — append
  it after `brag_stats` so the source reads chronologically.
- **`Filters` for `brag://memory/recent`** is `"(none)"` with an empty (non-nil)
  `map[string]string{}` — reuse `cli.echoFiltersForMemory`'s output shape. That
  function is package-private in `cli`; **do not export it and do not import
  `cli`** (the one-way dependency). The MCP side builds the same two values
  inline — it is four lines, and unlike the pool composition and the tokenizer
  there is no drift risk worth a package: the values are asserted byte-exactly
  by the parity tests. *(Named explicitly so build does not extract a third
  package by pattern-matching on LD8/LD9.)*
- **`nowFunc`.** The resource and tool handlers read `mcpserver.nowFunc()` for
  the envelope's `Generated:` line, exactly as `handleStats` does. Tests freeze
  it with the existing `setNowFunc` helper.
- **Errors from `Gather`.** A malformed query is a *user* error: on the CLI
  `runMemory` turns it into `UserErrorf`; over MCP the handler returns it as a
  tool error / `-32602`. `Gather` itself just returns the error — it does not
  know which surface it is on.
- **`brag mcp install` and `plugin/.mcp.json` need no change.** Resources are
  advertised by the running server at `initialize`, not declared in a manifest.
  Only `plugin/README.md`'s prose changes.
- **Complexity honesty.** This is M, but it is a *wide* M: four new files, two
  extractions, and a 14-site doc sweep. If build finds it running long, the
  clean split is `internal/ftsquery` + `internal/memory/pool.go` as their own
  PR — they have no user-visible surface. Note that splitting **re-sequences**
  the stage (the extraction PR ships first, SPEC-074 second); it is not a
  descope, and the stage does not close until both land.

---

## Build Completion

*Filled in at the end of the **build** cycle, before advancing to verify.*

- **Branch:** `feat/spec-074-mcp-resources`
- **PR (if applicable):** (opened at hand-off; see PR description)
- **All acceptance criteria met?** Yes — all bullets under `## Acceptance
  Criteria` verified: exact resource/template sets, MIME/Description/no-`size`
  on all three, real-client round-trips (§12(b)), byte-identical resource
  bodies (recent and project-boosted), the soft-boost positive proof (contains
  other-project entries) and its two error semantics (unknown name → slice,
  empty name → `-32602`), percent-decoding, `a/b` and unregistered-URI both
  `-32602`, `ttlMs`/`cacheScope` cache hints, the projects resource (names,
  status, count, no locations, empty-state message), five-tool `tools/list`,
  `brag_memory` byte-parity in both formats, budget `*int` semantics
  (0/negative error, honoured, default), format validation, no
  since/until/day in the schema, `ftsquery.Build`/`memory.Gather` (already
  landed, re-verified here), single pool composition on both surfaces, the
  two `internal/memory` mechanical guards, zero stray stdout bytes across a
  five-tool + `resources/read` round-trip, `no-sql-in-cli-layer` extension,
  unchanged `go.mod`, and the full six-gate set green.
- **New decisions emitted:**
  - `DEC-045` — The MCP push surface: three `brag://` resources, markdown
    bodies, and `brag_memory` as the fifth tool (written at design)
- **Deviations from spec:**
  - `internal/cli/mcp.go` has **two** `Long` strings that enumerate the tool
    names — `NewMCPCmd`'s (the spec's named line ~26) and
    `newMCPServeCmd`'s (a second one, a few lines below, whose tool list wraps
    across two source lines). The design-time premise-audit grep only caught
    the first: its single-line regex doesn't match text split across a line
    break. Re-running the same grep fresh at build (per AGENTS.md's
    audit-grep cross-check discipline) surfaced no *new* file, but did
    surface that the second `Long` in the *same* file the Outputs entry
    already named was about to go stale next to a freshly-updated one.
    Updated both — leaving one four-tool string beside a five-tool string in
    the same file would have been a self-contradiction the grep's line-break
    blind spot would otherwise have shipped silently.
  - LD9's already-landed `internal/memory/pool_test.go` (PR #135/#136) uses
    test names (`TestGather_ProjectAddsScopedRead`,
    `TestGather_QueryAddsMatchedInSearchOrder`,
    `TestGather_UnknownProjectIsNotAnError`, …) that differ from the
    illustrative names in this spec's `Decision↔test traceability` table
    (`TestGather_QueryAndProject_ThreeReads`, etc.) — same coverage, different
    names, because that extraction shipped ahead of this spec's design being
    finalized. Verified against the actual names during the mutation-check
    pass (rows 11–12); not renamed, per "LD8/LD9 already landed — verify,
    don't write."
- **Follow-up work identified:** None beyond DEC-045's own revisit triggers
  (a–g), which already cover the live open questions (the soft-boost rename
  fallback, `Resource.Size` population, multi-resource attachment cost, a
  networked-transport `cacheScope` reckoning, `state_note` on
  `brag://projects`).

### Build-phase reflection (3 questions, short answers)

1. **What was unclear in the spec that slowed you down?**
   — Nothing substantive. DEC-045's nine sub-decisions plus the §12(b)
   pre-flight transcript answered essentially every API question before a
   line of code was written — the one small gap was the second `cli/mcp.go`
   `Long` string not being named explicitly (see Deviations above), and
   that cost a grep re-run, not real time.

2. **Was there a constraint or decision that should have been listed but wasn't?**
   — No. `DEC-045/044/043/024/017/042/010/011/034` together covered the
   ranking, the budget, the envelope, the soft-boost semantics, the time
   vocabulary boundary, the search transform, and the JSON shape with no
   gaps found during implementation.

3. **If you did this task again, what would you do differently?**
   — Nothing structural. The one process step that earned its keep was
   re-running the design-time premise-audit grep fresh at build (AGENTS.md's
   audit-grep cross-check) — it is what caught the second `mcp.go` `Long`
   string before it shipped stale, which is exactly the discipline's stated
   purpose.

---

## Reflection (Ship)

*Appended during the **ship** cycle.*

1. **What would I do differently next time?**
   — <answer>

2. **Does any template, constraint, or decision need updating?**
   — <answer>

3. **Is there a follow-up spec I should write now before I forget?**
   — <answer>

4. **What can a user do now that they couldn't before?**
   — <answer>
