---
# Maps to ContextCore epic-level conventions.
# A Stage is a coherent chunk of work within a Project.
# It has a spec backlog and ships as a unit when the backlog is done.

stage:
  id: STAGE-019
  status: active                    # proposed | active | shipped | cancelled | on_hold
  priority: high
  target_complete: null

project:
  id: PROJ-006
repo:
  id: bragfile

created_at: 2026-08-07
shipped_at: null
---

# STAGE-019: corpus as agent memory

## What This Stage Is

The first *deep* PROJ-006 pillar: turn the corpus from something an agent can
query **if it thinks to** into working memory it consults **before it acts**.
When this stage ships, a connected MCP client can auto-load a ranked,
token-budgeted slice of the developer's own history at session start — without
the agent choosing to call anything — and, when it does choose to ask, it can
express the same time/authorship windows the CLI already offers.

The read layer is largely built and this stage does **not** rebuild it. Verified
baseline (code-checked 2026-08-07): `brag_list` (recency-ordered), `brag_search`
(bm25-ranked) and `brag_stats` already ship as MCP *tools*, and
`storage.ListFilter` already carries `Since`/`Until`/`Author`. Three additions
remain, and they are this stage:

1. **Filter parity** — the MCP `brag_list` input omits `since`/`until`/`day`/
   `author`, so an agent cannot ask for "recent" or "agent-authored" history over
   MCP even though storage supports both.
2. **The memory slice** — recency (`List`) and relevance (`Search`) are separate
   axes today, and `Limit` is a row count, not a token cost. The slice is a
   deterministic *blended* ranking (project + recency + match) trimmed to a
   **token budget**, cheap enough to auto-load every session.
3. **The resources surface** — the corpus is exposed only as tools today (pull:
   the agent must decide to call one). MCP **resources** are the push side:
   context a client can attach before the agent works. This is the actual
   "memory" in corpus-as-memory; without it the other two are just better tools.

## Why Now

This is the synthesis' ranked #1 (`docs/research/proj-005-idea-synthesis.md`)
and the head of the project's causal chain — **consult → trust → complete →
measure**. It is first because it is the only pillar that pays off with *today's*
corpus: signed provenance (#2) is worth building once provenance drives a
decision, completeness (#3) once the corpus feels sparse, benchmark (#4) once
cost/token stamping is reliable. Consulting history needs none of those.

It is also the cheapest of the four now that the baseline is understood. The
brief's original framing ("build read-side MCP") was wrong by a wide margin: the
read tools shipped in v0.3.0/v0.5.0. What is left is additive plumbing plus one
genuinely new algorithm — no storage rewrite, no migration, no new dependency.

The requirements signal the project was waiting for has also arrived in the
weaker, honest form: the session-opening ritual in this repo is still "read the
brief, read the last stage, re-derive where we are." That is re-derived context a
corpus should be answering.

## Success Criteria

- An MCP client that connects to `brag mcp serve` can **auto-load** history
  without any tool call: `resources/list` advertises the corpus, and
  `resources/read` returns a ranked, budget-trimmed slice.
- The slice is **blended**, not one-axis: given a project and/or a query, its
  ordering demonstrably differs from both plain recency and plain bm25, and the
  blend is deterministic (byte-exact goldens, no wall-clock in the core).
- The slice is bounded by an **estimated token budget**, not a row count, and the
  envelope reports what was included, what was dropped, and the estimate — so a
  caller can calibrate rather than guess.
- `brag_list` accepts `since` / `until` / `day` / `author`, matching what
  `storage.ListFilter` supports, with the CLI's mutual-exclusion semantics.
- The `cli↔mcpserver` import cycle that deferred `--since` at SPEC-040 is
  **resolved structurally** (shared parser), not worked around per-caller.
- No new top-level dependency; no CGO; no migration; no network. Full gate set
  green (`gofmt -l .`, `go vet ./...`, `go test ./...`, `just test-docs`, CodeQL).

## Scope

### In scope

- **Filter parity on `brag_list`** (`since`, `until`, `day`, `author`) and the
  structural fix that unblocks it: extracting the time-window parser out of
  `internal/cli` into a package both `cli` and `mcpserver` can import.
- **A deterministic memory-slice core** (new `internal/memory` package): pure
  `[]storage.Entry + options → ranked, budget-trimmed slice`. No SQL, no clock
  reads inside the ranking (the "now" used for recency decay is injected).
- **A `brag memory` CLI surface** rendering that slice under the DEC-014
  envelope (markdown + `--format json`) — the human dogfooding path, and the
  byte-exact golden contract the MCP surfaces mirror.
- **MCP resources**: a default slice, a project-scoped slice via a resource
  template, and the registered-projects list (so an agent uses real project
  names instead of inventing them).
- **A `brag_memory` MCP tool** — the parameterized pull counterpart to the
  zero-config push resources.
- Doc sweep for the new surfaces (`README.md`, `docs/api-contract.md`,
  `docs/tutorial.md`, `BRAG.md` where relevant, `AGENTS.md` glossary) and the
  `scripts/test-docs.sh` assertions that go with them.

### Explicitly out of scope

- **Any storage rewrite.** The stage reuses `Store.List` and `Store.Search`
  as-is. No new SQL, no schema change, no migration, no FTS re-tokenization.
- **Embeddings / vector search / any semantic ranker.** bm25 + recency + project
  is the blend; a model in the binary is out of the question (local-first,
  no-network, no new deps).
- **A real tokenizer.** The budget uses a documented deterministic *estimate*;
  bragfile does not ship or call a tokenizer. See DEC-044's honesty clause on
  why this estimate must never be confused with DEC-027's caller-reported
  `tokens:` provenance.
- **MCP prompts, sampling, subscriptions, `resources/updated` notifications.**
  Resources are read on demand; change-notification is a later ask if any.
- **Writing memory back** (agent-authored summaries, session notes, a
  `brag_remember` tool). This stage makes the corpus *readable as memory*; it
  does not add a second write path.
- **Signed provenance, capture completeness, benchmark, story-surface v2** —
  the other PROJ-006 pillars, framed separately.
- **`brag list --until` on the CLI.** STAGE-017 deliberately dropped it as YAGNI
  and `--day` covers the human need; `until` is exposed MCP-first (DEC-042
  records the asymmetry and the trigger that would close it).

## Spec Backlog

Format: `- [status] SPEC-ID (cycle) — one-line summary`

- [x] SPEC-072 (shipped on 2026-08-07) — MCP `brag_list` filter parity (`since`/`until`/`day`/
      `author`) + extract the time-window parser into `internal/timewindow`,
      retiring the `cli↔mcpserver` cycle that deferred `--since` at SPEC-040.
      **DEC-042 emitted** (parser home + the MCP time vocabulary + the `until`
      asymmetry; folds in STAGE-018's negative-`limit` parity item).
      Complexity M.
- [x] SPEC-073 (shipped on 2026-08-08) — the memory slice: `internal/memory` deterministic
      blended ranking + token-budget trim, surfaced as `brag memory` (markdown +
      JSON, DEC-014 envelope) with byte-exact goldens. The **EIGHTH** DEC-014
      consumer. **DEC-043 emitted** (reciprocal-rank fusion over three ORDINAL
      lists — recency, relevance, project — because bm25 rank and recency have no
      shared unit; `k=60` and unit weights locked and golden-pinned; `--project`
      is a soft boost, not a filter; ordinal recency means the ranking reads no
      clock at all). **DEC-044 emitted** (a positive-only `--budget` token count
      defaulting to 2000, a documented `ceil(bytes/4)` estimate scoped away from
      DEC-027's caller-reported token-count tag by a mechanical guard, greedy
      skip-and-continue enforcement, and the one-line entry rendering that *is*
      the unit of cost). Complexity M.
- [x] SPEC-074 (shipped on 2026-08-09) — the MCP push surface: resources
      (`brag://memory/recent`, `brag://memory/project/{name}`, `brag://projects`)
      plus the `brag_memory` tool — the server's **FIFTH** tool. **DEC-045
      emitted** (nine locked sub-decisions: the resource set, the custom
      `brag://` scheme, `text/markdown` on all three, `Resource.Size` left
      unpopulated, the `ttlMs: 0` / server-clobbered-`cacheScope` stance, the
      project template as a SOFT boost with its three corrections, the resource
      budget pinned to `memory.DefaultBudget`, the degenerate-ranking note, and
      `brag_memory`'s parameter set). **Absorbs two extractions:**
      `internal/ftsquery` (DEC-024's revisit trigger (c) — third consumer of the
      DEC-010 transform) and `memory.Gather` (the three-read pool composition
      both surfaces must share). Complexity M — a *wide* M; the pre-authorized
      split is the two extractions as their own PR, which re-sequences the stage
      rather than descoping it. Branch `feat/spec-074-mcp-resources`.
      **Gate: CLEARED.** PR #124 bumped the go-sdk `1.6.1 → 1.7.0`, invalidating
      the resources-API pre-flight; PR #131 (merged 2026-08-08, commit
      `f325013`) re-ran it and recorded the three deltas in the Design Notes
      below. `go.mod` has not moved since, so DEC-045 was locked against
      v1.7.0 as required.

**Count:** 3 shipped / 0 active / 0 pending — **backlog complete**

> **The stage is NOT complete until SPEC-074 ships.** All three specs are now
> scaffolded, so `just archive-spec` can see them — but do not run the Stage
> Ship prompt until SPEC-074's own archive step reports the backlog complete.

## Design Notes

Cross-cutting glue. The weighty forks each get a DEC (identified below with
their real alternatives, so design does not rediscover them).

**Sequencing.** 072 → 073 → 074. 072 is independent and unblocks nothing, but it
is the cheapest and it removes a structural blocker (the import cycle) that 074
would otherwise trip over. 074 depends on 073: a resource with nothing good to
serve is not a memory surface.

**One stage, not two.** Parity and resources are plumbing; the ranking is the
one real algorithm. They ship as one stage because the outcome is single —
"the corpus is working memory" — and none of the three is independently
worth a client reconnect. Each is still its own spec and its own PR
(`one-spec-per-pr`).

**The import cycle, resolved structurally (SPEC-072 / DEC-042).** `internal/cli`
imports `internal/mcpserver` (`brag mcp serve`), so `mcpserver` can never import
`cli` — which is why SPEC-040 deferred `--since` at the MCP boundary. Do **not**
duplicate `ParseSince` into `mcpserver`: DEC-024 already records the cost of
exactly that move (the DEC-010 tokenization is duplicated there and listed as a
negative consequence). Lean: move `ParseSince`/`ParseDay` (and their `clock`
seam + tests) into a small pure package both import, and update the three `cli`
call sites (`list.go`, `window.go`, `export.go`). The alternative — accept
RFC3339 only at the MCP boundary — is cheaper but makes the agent compute date
arithmetic that the binary already does correctly, including DEC-039's local-day
boundary. Weigh both in DEC-042; the shared-package path is the lean.

**The blend (SPEC-073 / DEC-043).** The hard part is that the two orderings are
incomparable: `List` returns `created_at DESC` and `Search` returns bm25 `rank`
(negative, unbounded, corpus-relative). Candidate approaches to weigh:
(a) **reciprocal-rank fusion** — `Σ 1/(k + rank_i)` over the recency list and
the relevance list, with a project term; needs no score normalization, is
deterministic, and degenerates cleanly to plain recency when no query is given
(one algorithm, both cases); (b) **weighted linear** over normalized components —
requires inventing a bm25 normalization, which is the fragile part;
(c) **lexicographic tiers** (matched∧same-project > matched > same-project >
recent, recency within tier) — most explainable, least tunable; (d) recency-only
within an FTS pre-filter — not really a blend. Lean (a) with an explicit project
term. Whatever wins, **lock the constants in the DEC and pin them with goldens**;
an unpinned weight is a silent behavior change later.

**The budget (SPEC-073 / DEC-044).** Four sub-choices, all design-decidable:
- *Expression* — a token count (`--budget N`), not rows. Lock the default.
- *Estimation* — deterministic and dependency-free (a documented chars-per-token
  heuristic). **Honesty clause:** DEC-027 states bragfile never estimates tokens;
  that rule is about caller-reported provenance (`tokens:` tags), and this
  estimate must be named, scoped, and documented so the two can never be
  conflated — it sizes a retrieval, it is never stamped on an entry.
- *Enforcement* — greedy fill in rank order, and the fork is **stop at the first
  overflow** (monotone, simple) vs **skip the oversize entry and continue**
  (better packing; prevents one long entry at rank 1 from starving the slice).
  Lean skip-and-continue, with the counts reported.
- *Per-entry rendering* — the slice's line shape is part of its token cost, so
  the rendering is locked with the budget, not left to the surface.
- The envelope reports `included` / `skipped` / `estimated_tokens` so a caller
  can calibrate. Budget accounting must be a pure function of the rendered
  bytes — the estimate and the render cannot disagree.

**The resource surface (SPEC-074 / DEC-045).** SDK API confirmed present at
v1.6.1 and **re-confirmed against v1.7.0** (the version in `go.mod` since #124;
see the delta note below) — `Server.AddResource(*Resource, ResourceHandler)`,
`Server.AddResourceTemplate(*ResourceTemplate, ResourceHandler)`, contents as
`ReadResourceResult.Contents []*ResourceContents{URI, MIMEType, Text}`. All
three signatures are byte-identical across the bump, as are the `Resource` and
`ResourceTemplate` structs (so `Resource.Size` is still available for the size
fork). Forks: the resource **set** (fixed vs template vs both — lean both);
the **URI scheme** (`brag://…` custom scheme vs `file://` — lean custom, this is
not a file); the **MIME/rendering** (markdown vs JSON — lean `text/markdown` for
the memory slices, because a resource is context for a model, not a data pipe,
and markdown is materially cheaper per token; JSON stays available via the CLI
and the tool); and **size policy** (`Resource.Size` is a client hint for context
budgeting — decide whether to populate it, given the content is generated per
read). Per §12(b), stand the resources up against a real in-memory `mcp.Client`
at design and confirm `resources/list` + `resources/read` round-trip *and* that
the template actually matches — shape validation is not registration.

**go-sdk v1.7.0 delta (re-run of the §12(b) pre-flight, 2026-08-08).** The
v1.6.1 finding above was re-verified behaviorally, not just by shape: a
throwaway server registering one static resource (`brag://memory/recent`) plus
one template (`brag://memory/project/{name}`) over `mcp.NewInMemoryTransports()`
driven by a real `mcp.Client` confirmed that `resources/list` advertises the
static resource, `resources/templates/list` advertises the template,
`resources/read` round-trips both, and the template genuinely matches a concrete
URI (`brag://memory/project/bragfile` reached the template handler carrying the
concrete URI). The custom `brag://` scheme survives both `url.Parse` (which
`AddResource` panics on) and `uritemplate.New` (which `AddResourceTemplate`
panics on), so the lean on a custom scheme still holds. **Three things did
change, and DEC-045 must start from them, not from the v1.6.1 shape:**
- *`ReadResourceResult` grew fields.* It now embeds `Cacheable`, adding the
  JSON fields `ttlMs` (int) and `cacheScope` (string) — **neither tagged
  `omitempty`** — and carries `InputRequests` / `RequestState` / `NeedsInput()`
  for multi-round-trip elicitation, plus a custom `MarshalJSON`. A read response
  now serializes as `{"_meta":…,"ttlMs":0,"cacheScope":"public","contents":[…],
  "resultType":"complete"}`, not `{"contents":[…]}`. `Contents` itself is
  unchanged, so the *handler* contract is untouched — but any golden pinned on
  the resources/read **wire** shape must include these fields.
- *`CacheScope` is server-clobbered; `TTLMs` is not.* `Server.readResource`
  calls `setDefaultCacheableValues()`, which assigns `CacheScope = "public"`
  unconditionally — a handler setting `"private"` is overwritten (observed:
  handler set `private`, client saw `public`). `TTLMs` passes through intact.
  This is a real input to the size/caching fork: a per-read TTL hint is
  expressible, a private cache scope is not.
- *`CodeResourceNotFound` changed value and is deprecated.* It went from
  `const = -32002` to `var int64 = jsonrpc.CodeInvalidParams` (**-32602**) per
  SEP-2164, and the symbol is now marked deprecated ("use
  `jsonrpc.CodeInvalidParams` directly; will be removed"). Pre-1.7.0 behavior is
  restorable only via `MCPGODEBUG=customresnotfounderrcode=1`. Any SPEC-074
  error-path test for an unregistered URI must assert `-32602` /
  `jsonrpc.CodeInvalidParams` and must not reference the deprecated constant.

> **Settled at SPEC-074 design (2026-08-08), and one fork these notes did not
> name.** All four forks above resolved as leaned — both a fixed set and a
> template, the custom `brag://` scheme, `text/markdown` on all three, and
> `Resource.Size` **not** populated (it is `omitempty`, the content is generated
> per read, and computing it truthfully costs a full slice per
> `resources/list`). Two additions:
>
> - **Cache semantics**, a fork created by the v1.7.0 delta below and absent
>   from the original notes: `ttlMs: 0` is set deliberately (the corpus changes
>   on every capture, so a cached slice is stale context presented as current),
>   and the server-clobbered `cacheScope: "public"` is documented as a stance
>   rather than worked around — vacuous over stdio, a hard blocker on any
>   networked transport, and now wired into DEC-024's revisit trigger (b).
> - **The fork nobody had named: a "project" resource that is not a scope.**
>   `brag://memory/project/{name}` reads as an address but denotes DEC-043
>   sub-decision 4's *soft boost*, so it returns entries from other projects.
>   Resolved as **keep the boost**, with three corrections on the consumption
>   path (`Title`/`Description` naming the operation, the `Description` carrying
>   the CLI flag help's exact phrase `a soft boost, not a filter`, and the
>   body's own `Filters:` line). A **hard filter was rejected** on three
>   independent grounds — it would diverge from the CLI, diverge from
>   `brag_memory`'s own `project` param on the same server, and (DEC-017) drop
>   history silently on the one surface whose caller is least equipped to notice.
>   A **rename** is the pre-authorized first move if the misread proves real.

**Determinism.** Everything in `internal/memory` is a pure function of
`(entries, options, now)`. `now` is injected, never read inside the package
(AGENTS.md §9 os-level-seam habit). This is what makes byte-exact goldens
possible for a ranking that is otherwise time-dependent.

> **Settled stronger at SPEC-073 design (2026-08-08).** There is no `now` at all.
> DEC-043 makes the recency signal **ordinal** (a rank, not an age), so the
> ranking is time-INVARIANT rather than time-dependent-but-injected — a strictly
> stronger determinism property than this note anticipated. `memory.Options`
> therefore carries no `Now` field (the renderer owns the wall clock, as in every
> other DEC-014 consumer), and `TestPackageReadsNoWallClock` enforces the absence
> mechanically. Ordinal recency also self-normalizes to the developer's capture
> cadence and is what lets the fusion stay principled — mixing a continuous decay
> with an ordinal bm25 rank would reintroduce the normalization problem RRF exists
> to avoid.

**Reuse, don't re-query.** The slice composes `Store.List` and `Store.Search`
results in Go. No new storage method unless a spec proves one is needed — and if
one is, it is a DEC, not a build-time discovery.

## Dependencies

### Depends on
- PROJ-003 / SPEC-040 + DEC-024 (shipped): the MCP server, the SDK choice, the
  stdio transport, and the CLI-byte-parity tool contract this extends.
- PROJ-005 / SPEC-056 + DEC-035 (shipped): `ListFilter.Until` in storage.
- PROJ-004 / SPEC-045 + DEC-033 (shipped): the `author`/provenance classifier
  (`ListFilter.Author` + `aggregate.IsAgentAuthored`) that filter parity exposes.
- STAGE-017 / SPEC-068 + DEC-039 (shipped): `ParseDay` and the local-day
  boundary the shared parser carries over to MCP.
- DEC-014 (shipped): the rule-based output envelope `brag memory` renders into.

### Enables
- **Signed provenance (next pillar).** Once an agent *reads* history back, the
  question "can I trust what it says about who did this" stops being theoretical
  — this stage is what makes #2 worth building.
- **Story surface v2.** The memory slice and the story surface share this MCP
  read surface; the resource shape locked here is the one `story`/`wrapped`
  would ride.
- Cheaper session starts generally: less re-derived context per session, which
  is the project brief's stated goal.

## Stage-Level Reflection

*Filled in when status moves to shipped. Run Prompt 1c (Stage Ship) in
FIRST_SESSION_PROMPTS.md to draft this.*

- **Did we deliver the outcome in "What This Stage Is"?** <yes/no + notes>
- **How many specs did it actually take?** <number vs. plan>
- **What changed between starting and shipping?** <one sentence>
- **Lessons that should update AGENTS.md, templates, or constraints?**
  - <one-line updates>
- **Should any spec-level reflections be promoted to stage-level lessons?**
  - <one-line items>
