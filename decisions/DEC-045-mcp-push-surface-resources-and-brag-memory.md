---
# Maps to ContextCore insight.* semantic conventions.

insight:
  id: DEC-045
  type: decision
  confidence: 0.82                   # honest: the resource set, the custom
                                     # scheme, the markdown MIME, the shared
                                     # 2000 budget, and the soft-boost
                                     # resolution are strong and behaviorally
                                     # pre-flighted; the soft spots are the
                                     # unsuppressable `cacheScope: public`
                                     # announcement (a stance, not a fix) and
                                     # the omitted `Resource.Size` hint, which
                                     # is a bet about client behavior we cannot
                                     # observe from here — see Validation.
  audience:
    - developer
    - agent

agent:
  id: claude-opus-5
  session_id: null

# Decisions are repo-level, but it's useful to track which project
# caused them to be emitted.
project:
  id: PROJ-006
repo:
  id: bragfile

created_at: 2026-08-08
supersedes: null
superseded_by: null

tags:
  - mcp
  - resources
  - agent-native
  - memory
  - uri-scheme
  - caching
  - privacy
---

# DEC-045: The MCP push surface — three `brag://` resources, markdown bodies, and `brag_memory` as the fifth tool

## Decision

`brag mcp serve` gains a **push surface**: three MCP **resources** a client can
auto-load before the agent works, plus **`brag_memory`**, the parameterized pull
counterpart — the server's **fifth** tool.

```
brag://memory/recent            static resource   text/markdown
brag://memory/project/{name}    resource template text/markdown
brag://projects                 static resource   text/markdown
brag_memory {query, project, budget, format}      fifth tool
```

Nine coupled sub-decisions are locked. They are one decision because the
resource set is meaningless without the URI scheme it is addressed by, the
rendering it serves, and the budget it is pinned to — a resource, unlike a
tool, carries no arguments, so every choice a caller would otherwise make is
made here.

### 1. The resource set — both a fixed set and a template, three entries

Two **static** resources and one **template**. The static pair is what makes
the surface *push*: a client can attach `brag://memory/recent` with no
knowledge of the corpus at all. The template is what makes it *useful*: the
one dimension a caller reliably knows before it reads anything is which project
it is working in, and enumerating one static resource per registered project
would make the resource list grow without bound and go stale on every
`brag project ensure`.

`brag://projects` exists so the `{name}` slot can be filled with a **real**
name. Without it the template is a hole an agent fills by guessing, and a
guessed project name silently produces an unboosted slice (sub-decision 6) —
a failure with no error message. The projects resource is the template's
lookup table, not a separate feature.

### 2. URI scheme — a custom `brag://`, not `file://`

These are not files. Nothing on disk corresponds to `brag://memory/recent`;
the bytes are computed per read from a SQLite corpus. `file://` would be a
lie that the SDK actively cooperates with — its `readFileResource` path
resolves `file://` URIs against roots on the real filesystem — so borrowing the
scheme would put us one refactor away from something that tries to open a path.

Behaviorally confirmed (§12(b), see Context): `brag://…` survives both
`url.Parse` (which `AddResource` panics on) and `uritemplate.New` (which
`AddResourceTemplate` panics on), and a concrete `brag://memory/project/bragfile`
genuinely reaches the template handler.

**The `{name}` segment is percent-decoded by us, not by the SDK.** The handler
receives `req.Params.URI` verbatim — `brag://memory/project/my%20project` arrives
with the `%20` intact (observed) — so the handler calls `url.PathUnescape` on the
trailing segment before using it as a project name. RFC 6570 simple expansion does
not cross `/`, so `brag://memory/project/a/b` matches no template and the SDK
answers not-found before any handler runs (observed). Both are locked and tested.

### 3. MIME and rendering — `text/markdown` for all three; JSON stays on the CLI and the tool

A resource is **context for a model**, not a data pipe. The consumer is a
context window, and markdown is materially cheaper per token than the same
content as JSON — DEC-044 made the token cost of this exact body a first-class
number, and serving it as JSON would inflate it for nothing. Every structured
form stays reachable: `brag memory --format json` on the CLI, and
`brag_memory {"format":"json"}` on the tool (sub-decision 8).

The two memory resources serve **exactly the bytes `export.ToMemoryMarkdown`
produces** — the same `memory.Result`, the same renderer, the same goldens
SPEC-073 pinned. There is no second rendering. This is DEC-024's CLI-parity
contract carried onto the resource surface: same options → same bytes.

`brag://projects` is **also** `text/markdown`, and the reasoning is the same one,
applied consistently rather than abandoned at the third resource. It does *not*
serve `brag project list`'s bytes in either form: the plain form is a TSV for
shell pipelines (the data pipe this sub-decision rejects) and the `--format json`
form is a 7-key-per-project record whose `id`, `created_at`, `updated_at`, and
`locations` are noise for the one job this resource has. It gets a small dedicated
renderer (`export.ToProjectsMarkdown`) over `Store.ProjectStatuses()`, with two
deliberate exclusions:

- **Locations are excluded.** They are local filesystem paths, irrelevant to
  choosing a name, and they do not belong in a body the SDK announces as
  publicly cacheable (sub-decision 5).
- **Archived projects are excluded** (`ProjectStatuses` already filters them).
  An archived name is precisely a name an agent should not be capturing into.

`state_note` is excluded too, for a weaker reason: unbounded length against a
resource whose whole value is being cheap. Revisit trigger in Validation.

### 4. Size policy — `Resource.Size` is NOT populated

`Resource.Size` is an `int64` the SDK documents as *"the size of the raw resource
content, in bytes … if known,"* used by hosts to estimate context-window usage. It
is `omitempty`, and it does round-trip to the client (observed: `size=778` on the
wire).

We leave it unset, for three reasons in order of weight:

1. **It is not known.** The contents are generated per read from a corpus that
   changes whenever the developer captures an entry. Any value written at
   registration time is a prediction about a future computation, and "if known"
   is the SDK's own guard against exactly this.
2. **Computing it truthfully costs a full slice per `resources/list`.** SPEC-073
   does compute an exact `EstimateTokens` per slice — but only *after* running
   `memory.Slice` over three bounded `Store` reads. A client may call
   `resources/list` on every reconnect; paying three DB reads × three resources
   for a number that is stale by the time of the read is a bad trade.
3. **The number a caller actually needs is a token budget, not a byte count** —
   different unit, and it is knowable in advance because it is *pinned*
   (sub-decision 7). It is stated in each resource's `Description` and reported
   exactly, per read, in the body's own `## Budget` section.

Rejected: populating `Size` with the static upper bound
`DefaultBudget × CharsPerToken` (= 8000 bytes). It is honest as a bound but
pessimistic by roughly an order of magnitude against a real slice, and a host
that budgets on it may decline to auto-load the resource — which defeats the
entire point of a push surface. A wrong hint is worse here than no hint.

### 5. Cache semantics — `ttlMs: 0` deliberately; `cacheScope: "public"` is announced against our will, and we say so

New at go-sdk v1.7.0: `ReadResourceResult` embeds `Cacheable{TTLMs, CacheScope}`,
neither `omitempty`, so every read response now carries both fields on the wire.

- **`TTLMs` passes through** the server intact (observed: a handler-set `60000`
  reached the client). We set **`0`**, which the SDK documents as *"the response
  SHOULD be considered immediately stale."* That is the correct hint and not
  merely the zero value: the corpus changes on every capture, and a cached memory
  slice is worse than no memory slice — it is stale context presented as current,
  which is the specific failure this whole stage exists to remove.
- **`CacheScope` is clobbered.** `Server.readResource` calls
  `setDefaultCacheableValues()`, which assigns `CacheScope = "public"`
  **unconditionally**; a handler setting `"private"` is overwritten (observed:
  handler set `private`, client saw `public`). A private cache scope is **not
  expressible** through this SDK.

So the server announces a body that is the developer's private work history as
*"any client or intermediary MAY cache and serve"* it. Our position, stated
rather than shrugged off:

**Do not work around it.** `brag mcp serve` is stdio-only, local, and
single-user (DEC-024): there is no intermediary between client and server and no
network hop, so the misannouncement has no reachable consequence in the shipped
transport. Hand-writing the JSON to defeat the SDK would buy nothing real and
cost us a fork of the protocol layer.

**But it is a live constraint on the transport question, not a curiosity.** This
DEC therefore adds a clause to DEC-024's revisit trigger (b): a networked or
multi-user MCP mode must confront `cacheScope: public` on resource reads
alongside the WAL/busy-timeout concurrency question — at that point the
announcement stops being vacuous and starts being wrong. A test pins
`ttlMs == 0` and pins `cacheScope == "public"` **as observed, not as chosen**, so
that the day the SDK stops clobbering is a visible red test rather than a silent
change in what we tell the world about someone's work history.

### 6. `brag://memory/project/{name}` is a SOFT boost — the URI over-promises, and the correction goes where consumers actually look

A URI reads as an **address**, and `brag://memory/project/orbit` reads like a
scope. It is not one. DEC-043 sub-decision 4 locks `--project` as *"a SOFT boost,
never a hard filter"*, and the shipped flag help says so in as many words. So a
client reading this resource gets entries from other projects, ranked lower but
present.

**The soft boost is kept.** Three corrections ride with it, each on a path a
consumer necessarily takes:

1. **`Name`/`Title`** name the operation, not a scope — `Title: "Memory slice,
   boosted toward <project>"`. This is what renders in a client's resource picker.
2. **`Description`** states it in the CLI's own words — *"a soft boost, not a
   filter"* — and the SDK documents `Description` as *"a hint to the model … to
   improve the LLM's understanding of available resources"*, which makes it the
   designed-for place to correct an address's implication, read by the same
   audience that reads the URI.
3. **The body is self-describing.** The DEC-014 envelope's first five lines print
   `Scope: lifetime` and `Filters: --project orbit`. You cannot consume this
   resource without reading that.

The residual misreading window is: a consumer that sees the URI, skips the
description, and does not read the body it just loaded. That is not a reachable
failure mode.

Two live alternatives, both rejected:

- **Make the resource a hard filter.** Rejected on three independent grounds.
  (a) It would diverge from `brag memory --project` for the same concept,
  breaking the "same options → same bytes" parity SPEC-073 established and
  DEC-024 contracted. (b) Worse, it would diverge from **`brag_memory`'s own
  `project` param** on the same server — two surfaces of one concept disagreeing,
  which no amount of documentation repairs. (c) Most importantly, DEC-043's
  reason bites hardest exactly here: `entries.project` is a soft string match
  (DEC-017) and is frequently empty, so a hard filter *silently drops* history —
  and the resource surface is the zero-config auto-load path, whose caller is the
  one least equipped to notice that anything went missing. A hard filter is the
  wrong semantics precisely where it is most dangerous.
- **Rename so the URI stops over-promising** (`brag://memory/for-project/{name}`
  or similar). Rejected as the weaker fix, not the wrong one: `for-project` is
  barely less scope-flavored, every genuinely unambiguous alternative
  (`…/project-boosted/…`) is an awkward permanent tax on the thing humans read in
  a picker and agents type, and it buys nothing the three corrections above do
  not already buy. **This is the pre-authorized first move** if the misreading
  turns out to be real (Validation) — the rename, not the hard filter.

**Two error semantics fall out of soft-boost and are locked here** because they
are the surface where a hard-filter reflex would show up:

- **An unknown project name is NOT an error.** `brag://memory/project/nosuchthing`
  returns a slice — the project term boosts nothing and the score degenerates to
  plain recency, exactly as `brag memory --project nosuchthing` does. A hard
  filter would have returned empty, which is the failure the soft boost exists to
  avoid.
- **An empty name IS an error.** `brag://memory/project/` matches the template and
  reaches the handler with an empty segment (observed — the SDK does not reject
  it). It is answered with `mcp.ResourceNotFoundError(uri)`: the URI promises a
  project scope and names none, which is the same condition `brag memory
  --project ""` rejects as a `UserError`.

### 7. The resource budget is `memory.DefaultBudget` (2000) — the same constant, deliberately

A static URI carries no `--budget`, so each resource is **pinned** to one number,
and because this is the auto-loaded-every-session path, that number *is* the whole
cost story for the surface.

It is `memory.DefaultBudget` — **the same exported constant the CLI defaults to**,
not a second number that happens to equal it. DEC-044 derived 2000 from the
auto-load cost story *specifically* (the ~20k-token session-opening ritual this
stage replaces; 2000 is ~10% of it, ≈110 entries). That derivation was always
about this surface; the CLI inherited it. Minting a resource-specific default
would give one claim two numbers and would break the dogfooding property SPEC-073
established — that `brag memory` is a faithful, byte-exact preview of what the
resource serves.

**Accepted cost, named:** a client that attaches all three resources pays
≈2000 + ≈2000 + a small projects list before the agent does anything. Attachment
is per-resource and is the client's/user's choice, and any caller wanting a
different number has `brag_memory {"budget": n}`. Revisit trigger in Validation.

### 8. `brag_memory` — the fifth tool, mirroring `brag memory` exactly

Params: **`query`, `project`, `budget`, `format`.** Nothing else.

- **`budget` is a `*int`, not an `int`.** A non-pointer `int` cannot distinguish
  an omitted field from an explicit `0`, and DEC-044 makes `0` a **`UserError`**
  rather than "unlimited". The SDK's schema inference handles `*int` cleanly
  (observed: `"type":["null","integer"]`; omitted → `nil`, explicit `0` → a
  pointer to `0`), so the CLI's semantics survive the transport verbatim: omitted
  → `DefaultBudget`, `0` or negative → tool error. This is deliberately *unlike*
  `brag_list`'s `limit int`, where `0 = unlimited` is the same on both sides.
- **`format` defaults to `"markdown"`,** matching the CLI. `brag_memory` is the
  first tool with an output-format param — the other four are JSON-only — and the
  asymmetry is the point: sub-decision 3 says markdown is materially cheaper as
  model context, and denying the *pull* path the cheap form while handing it to
  the *push* path would be incoherent. An unknown value is a tool error, mirroring
  the CLI's `UserError`. DEC-024's parity contract generalizes rather than breaks:
  `brag_memory` is byte-identical to `brag memory --format <the same value>`.
- **No `since` / `until` / `day`.** SPEC-073 put time windows out of scope for the
  memory slice with a reason worth preserving verbatim: the blend already prefers
  recent, hard windows are `brag list`'s job (and MCP has them there since
  SPEC-072/DEC-042), and a window would make the envelope's `Scope: lifetime` line
  a lie. Adding them to the tool would reintroduce that contradiction on the one
  surface where nobody can see the flag help explaining it.

### 9. Resources exercise DEC-043's degenerate path, by design

A resource URI carries at most a project name, never a query. So on the resource
surface `Options.Matched` is always empty and the score reduces to
`Wr/(k + rank_recency)` — plus `Wp/(k + rank_project)` for the template. **The
relevance term is never exercised by a resource.**

This is the intended consequence of DEC-043's "one algorithm, both cases," and it
is stated rather than merely true because it has a testing implication: a bm25
regression cannot be caught by resource tests, only by `brag memory --query` and
`brag_memory {"query": …}`. The relevance term's coverage lives entirely on the
pull surfaces.

Rejected: a query-carrying template (`brag://memory/query/{q}`). A query is a
per-ask parameter, which is the definition of the pull half; putting it in a URI
would mean a resource list that cannot enumerate its own useful members. That is
what `brag_memory` is for, and the push/pull split is the stage's whole shape.

## Context

STAGE-019's third and last spec (SPEC-074). SPEC-072 gave the corpus its filters;
SPEC-073 gave it a blended, token-budgeted slice (`internal/memory.Slice` +
`export.ToMemoryMarkdown`/`ToMemoryJSON` + `brag memory`). Both are **pull**: the
agent must decide to call something. This DEC is the **push** half, and it is the
actual "memory" in corpus-as-memory — without it the other two are better tools,
not memory.

**The pre-flight, and the version it had to be re-run against.** PR #124 bumped
the go-sdk `1.6.1 → 1.7.0`, invalidating the resources pre-flight recorded at
STAGE-019 framing. PR #131 re-ran it behaviorally (a throwaway server driven by a
real `mcp.Client` over `mcp.NewInMemoryTransports()`) and recorded three deltas
this DEC starts from rather than inherits stale: `ReadResourceResult` gained
`Cacheable` + elicitation fields + a custom `MarshalJSON`; `CacheScope` is
server-clobbered to `"public"`; and `CodeResourceNotFound` moved `-32002 →
jsonrpc.CodeInvalidParams` (**-32602**) per SEP-2164 and is now deprecated. Three
API shapes held unchanged (`AddResource`, `AddResourceTemplate`,
`ReadResourceResult.Contents []*ResourceContents{URI, MIMEType, Text}`), so the
handler contract is untouched.

SPEC-074 design ran a **further targeted pre-flight** at v1.7.0 for the claims
this DEC adds beyond that re-run — a second static resource, `Resource.Size`
round-tripping, the error the *client* actually sees, template edge cases, and
`*int` schema inference. Its observations are transcribed in SPEC-074's §12(b)
section and are the source of every "(observed)" in this file. Two of them
changed the design:

- **The not-found error does not arrive as a `*jsonrpc.Error`.** At the client it
  is a `*fmt.wrapError` (`calling "resources/read": Resource not found`); a plain
  type assertion fails and `json.Marshal` yields `{}`. `errors.As` recovers it —
  `Code == jsonrpc.CodeInvalidParams` (-32602), `Message == "Resource not found"`,
  `Data == {"uri":"…"}`. Any error-path test must use `errors.As`, and must not
  reference the deprecated `mcp.CodeResourceNotFound` constant.
- **A static resource shadows a template whose pattern would also match it.**
  `Server.lookupResourceHandler` tries registered resources first. Nothing in this
  DEC's set collides, but the precedence is now known rather than assumed.

**The unnamed fork.** Sub-decision 6 — whether a URI that reads as a scope may
denote a boost — was not in STAGE-019's Design Notes. It surfaced at SPEC-074
design when the resource set met DEC-043 sub-decision 4, and it is the only fork
here where the easiest implementation (whatever `--project` already does) and the
most defensible answer had to be checked against each other rather than assumed
equal.

## Alternatives Considered

Sub-decision-level alternatives are argued inline above (they are local to their
fork). Three whole-surface alternatives were considered and rejected:

- **Option A: tools only — no resources at all; ship `brag_memory` and stop.**
  - What it is: expose the memory slice as a fifth tool and let agents call it.
  - Why rejected: it is the status quo with one more tool, and it fails the
    stage's headline success criterion — *"an MCP client can auto-load history
    without any tool call."* A tool is pull: the agent must decide to call it,
    which is precisely the "consults it if it thinks to" behavior corpus-as-memory
    exists to replace. Resources are the only MCP mechanism that puts context in
    front of a model before it acts.

- **Option B: resources only — no `brag_memory` tool.**
  - What it is: the three resources, and callers who want a different budget or a
    query use the CLI.
  - Why rejected: it strands every parameterized ask. A resource URI cannot carry
    a query, and the relevance half of DEC-043 would then have no MCP surface at
    all — an agent could read history but never ask *"what do I already know about
    this?"* over the protocol. It also forces a shell-out for callers that have no
    shell.

- **Option C: MCP prompts instead of resources.**
  - What it is: expose the slice as an MCP *prompt* the client can insert.
  - Why rejected: prompts are user-invoked templates, so this reintroduces a
    deliberate act at a different layer while adding a third concept. STAGE-019
    put prompts, sampling, and subscriptions explicitly out of scope. Resources
    are the auto-attachable primitive; this is the job they exist for.

- **Option D (chosen): both static resources and a template, plus the tool.**
  - Why selected: it is the only shape that satisfies both halves of the stage —
    zero-config push for the session opener, parameterized pull for the targeted
    ask — over one ranking, one renderer, and one budget constant, with no second
    rendering and no new storage read.

## Consequences

- **Positive:** the stage's headline criterion is met — a connected client can
  auto-load a ranked, budget-trimmed slice of the developer's own history with no
  tool call and no configuration.
- **Positive:** the resource body *is* `export.ToMemoryMarkdown`'s output, so
  CLI↔MCP parity is by construction, not by a parity test. SPEC-073's goldens
  cover the resource surface for free, and DEC-044's "the resource *is* the
  markdown" consequence lands as written.
- **Positive:** `brag://projects` closes the guessed-project-name hole. The
  template's `{name}` slot now has an enumerable source, and the failure it
  prevents (a guessed name producing a silently unboosted slice) has no error
  message to catch it downstream.
- **Positive:** the push surface pays the price of DEC-043's cap and DEC-044's
  budget *once, in advance*, on the path where those numbers matter most. A
  caller cannot accidentally auto-load an unbounded slice.
- **Negative (accepted):** the server announces private work history as
  `cacheScope: "public"` and we cannot suppress it (sub-decision 5). Vacuous over
  stdio; a hard blocker on any networked transport, and now wired to DEC-024's
  revisit trigger (b) so it cannot be forgotten there.
- **Negative (accepted):** `Resource.Size` is absent, so hosts that budget context
  from the resource list have no hint from us (sub-decision 4). The bet is that no
  hint beats an order-of-magnitude-pessimistic one; we cannot observe host
  behavior from here, which is the largest single reason this DEC is 0.82 and not
  higher.
- **Negative (accepted):** three attachable resources at ~2000 tokens each
  compound in a client that attaches them all (sub-decision 7).
- **Negative (accepted):** the project URI reads as a scope and denotes a boost
  (sub-decision 6). Corrected in three places on the consumption path; the
  residual is a picker-level misread by a human who does not open the resource.
- **Neutral:** the tool count moves **four → five**, which is asserted in prose in
  ~14 live places (enumerated in SPEC-074's premise audit). DEC-024 says "four" and
  is a **historical record — it is not rewritten**; this DEC is the amendment, and
  DEC-024's tool-set line should be read through it.
- **Neutral:** `brag_memory` is not a ninth DEC-014 consumer. It renders the
  eighth's envelope over a different transport; the consumer ordinal chain in
  `docs/api-contract.md` and AGENTS.md §11 is unchanged.
- **Neutral (structural, recorded here rather than as its own DEC):** SPEC-074
  fires DEC-024's revisit trigger (c) — a third consumer of the DEC-010 search
  transform — and extracts `cli.buildFTS5Query` / `mcpserver.buildMatch` into
  `internal/ftsquery`, following `internal/timewindow`'s precedent from this same
  stage. The identical argument applies to the three-read pool composition in
  `internal/cli/memory.go`, which the resources and the tool must reproduce
  exactly: it moves to `internal/memory/pool.go` as `memory.Gather`. Both are
  mechanical moves with a precedent and one obvious home, so they are locked as
  spec-level design decisions rather than minted as a DEC — but they are named
  here because the parity property this DEC claims ("the resource *is* the
  markdown") depends on both surfaces running *one* composition and *one*
  tokenizer, not two that agree today.

## Validation

**Right if:**
- A real client (Claude Code) connecting to `brag mcp serve` lists all three
  resources, reads each without a tool call, and the agent's session-opening
  behavior changes — it stops re-deriving "where were we" from the repo.
- `resources/list`, `resources/templates/list`, and `resources/read` round-trip
  against a real in-memory `mcp.Client`, and the template genuinely matches a
  concrete URI — **registration, not shape** (§12(b); SPEC-041's lesson that a
  manifest which validated still registered zero servers).
- The two memory resources' bodies are **byte-identical** to
  `brag memory` / `brag memory --project X` at the same budget — pinned by a test
  that compares the resource read against `export.ToMemoryMarkdown` on the same
  store, not by two goldens that agree.
- `brag://memory/project/nosuchthing` returns a slice (not empty, not an error)
  and `brag://memory/project/` returns `-32602` — the two assertions that would
  both flip if someone "fixed" the soft boost into a filter.
- `ttlMs == 0` and `cacheScope == "public"` on every read, the latter pinned as
  observed-not-chosen.
- The tool set is exactly five and `brag_memory` is byte-identical to
  `brag memory` at the same options in both formats.

**Revisit if:**
- **(a) A user is surprised by cross-project entries** in a project resource. First
  move is the **rename** (sub-decision 6's `brag://memory/for-project/{name}`), not
  a hard filter — the hard filter stays rejected for DEC-017's reason regardless of
  how the surprise is reported.
- **(b) A host declines to auto-load** for lack of a `Resource.Size` hint. Then
  populate it with the `DefaultBudget × CharsPerToken` upper bound and document it
  as a bound in the `Description`. This is the one sub-decision most likely to be
  wrong, and it is a one-field change.
- **(c) Multi-resource attachment proves expensive** in practice (a client
  attaching all three at once). First move: lower the *resource* budget only, which
  at that point becomes a second constant and needs the sub-decision-7 argument
  re-made rather than assumed.
- **(d) A networked or multi-user MCP mode is proposed.** Then `cacheScope:
  "public"` becomes load-bearing and must be confronted alongside DEC-024's
  existing WAL/busy-timeout question. Recorded as an addition to DEC-024's revisit
  trigger (b).
- **(e) The SDK stops clobbering `CacheScope`,** which the pinned test will surface
  as a red build. Then set `"private"` and delete this apology.
- **(f) Resource *subscriptions* are asked for** (`resources/updated` on capture).
  Out of scope by stage decree; `ttlMs: 0` is the honest stand-in until then, and
  the ask is the trigger.
- **(g) `state_note` turns out to be what agents actually need** from
  `brag://projects`. Then it is a length-bounded addition to
  `export.ToProjectsMarkdown`, not a second resource.

## References

- Related specs: SPEC-074 (emits + implements this DEC — the resources, the
  `brag_memory` tool, and the two extractions), SPEC-073 (shipped — the
  `memory.Slice` + `export.ToMemory*` body this serves), SPEC-072 (shipped — MCP
  filter parity + the `internal/timewindow` extraction precedent), SPEC-040
  (shipped — the MCP server this extends), SPEC-041 (shipped — the plugin
  packaging whose "validated ≠ registered" lesson shapes the §12(b) bar)
- Related decisions: DEC-024 (the MCP server contract — SDK, stdio purity,
  CLI byte-parity, the four-tool line this DEC amends, and revisit triggers (b)
  and (c) which this DEC fires and extends), DEC-043 (the ranking; sub-decision 4's
  soft boost is the source of this DEC's sub-decision 6, and sub-decision 3's
  degenerate path is its sub-decision 9), DEC-044 (the budget + the line shape;
  `DefaultBudget` is this DEC's pinned resource budget, and its "the resource *is*
  the markdown" consequence is realized here), DEC-014 (the envelope the memory
  bodies render), DEC-017 (`entries.project` is a soft string match — the reason a
  hard filter drops history), DEC-042 (the MCP time vocabulary this tool
  deliberately does not extend; also the `internal/timewindow` extraction
  precedent), DEC-010 (the search transform `internal/ftsquery` extracts),
  DEC-011 (the JSON shape `brag_memory --format json` narrows), DEC-034
  (`brag mcp install` — the registration path a client uses to reach these
  resources)
- Related constraints: `no-sql-in-cli-layer` (blocking — extended to
  `internal/mcpserver` by test convention), `no-new-top-level-deps-without-decision`
  (`go.mod` is untouched; `jsonrpc` is a package of the SDK module already gated by
  DEC-024), `no-cgo`, `stdout-is-for-data-stderr-is-for-humans` (the stdio frame
  stream), `test-before-implementation`, `errors-wrap-with-context`,
  `one-spec-per-pr`
- External: `github.com/modelcontextprotocol/go-sdk` v1.7.0 —
  `Server.AddResource`, `Server.AddResourceTemplate`, `ResourceHandler`,
  `Resource.Size`, `Cacheable{TTLMs, CacheScope}`,
  `Server.setDefaultCacheableValues`, `mcp.ResourceNotFoundError`,
  `jsonrpc.CodeInvalidParams`; MCP spec `resources/list`,
  `resources/templates/list`, `resources/read`; SEP-2164 (the
  `-32002 → -32602` move); RFC 6570 (URI templates — simple expansion does not
  cross `/`)
- Discussions: STAGE-019 Design Notes ("The resource surface (SPEC-074 /
  DEC-045)" + the "go-sdk v1.7.0 delta" subsection); PR #124 (the SDK bump);
  PR #131 (the §12(b) re-run whose three deltas this DEC starts from)
