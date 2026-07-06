---
# Maps to ContextCore insight.* semantic conventions.

insight:
  id: DEC-027
  type: decision
  confidence: 0.82                   # honest: the namespace-extension and
                                     # author-classification isolation are
                                     # high-confidence (they ride DEC-024's
                                     # proven path and the store.go clause is
                                     # already prefix-anchored to agent:/model:
                                     # only); the residual soft spots are the
                                     # numeric-format choice for cost: (unit +
                                     # decimal shape) and the session-id
                                     # delivery relying on Claude to forward
                                     # the hook-surfaced value — see Validation.
  audience:
    - developer
    - agent

agent:
  id: claude-opus-4-8
  session_id: null

# Decisions are repo-level, but it's useful to track which project
# caused them to be emitted.
project:
  id: PROJ-004
repo:
  id: bragfile

created_at: 2026-07-05
supersedes: null
superseded_by: null

tags:
  - mcp
  - provenance
  - reserved-namespace
  - cost
  - tokens
  - session
  - seed-early
  - economics
---

# DEC-027: Seed cost / session / token capture as reserved-namespace tags

## Decision

Extend DEC-024's reserved-tag namespace on the MCP `brag_add` provenance path
with three new **optional** reserved tags — `session:<id>`, `cost:<n>`, and
`tokens:<n>` — stamped by the same `stampProvenance` path that already emits
`agent:`/`model:`. All three inputs are **optional** (empty → no tag, exactly
like `agent`/`model` today); bragfile **never fabricates** a value. The real,
reliable payload is the **`session:` join-key** (a stable per-session id the
caller forwards); `cost:` / `tokens:` carry real numbers **only when a caller
supplies them**. This is **migration-free** — the tags ride the DEC-015
taggings join unchanged — and ships as a **v0.3.x patch** so cost/session
history begins accruing before the economics layer (PROJ-005) needs it.

Two sub-decisions ride with it:

1. **`session:`/`cost:`/`tokens:` are reserved but are NOT author-provenance
   tags.** `store.go`'s `provenanceExistsClause` (the `--author agent|human`
   classifier) stays **`agent:%`/`model:%`-only** and must NOT enumerate the
   three new prefixes. An entry carrying only a `session:`/`cost:`/`tokens:`
   tag is **not** agent-authored.

2. **Value normalization for the numeric tags.** `cost:` is a **non-negative
   decimal string in USD** (the unit is fixed by convention, not encoded in the
   tag; up to a few fractional digits, e.g. `cost:0.42`, `cost:12`). `tokens:`
   is a **non-negative integer** (e.g. `tokens:18000`). `session:` is an opaque
   identifier normalized like any reserved tag. Non-numeric / negative
   `cost`/`tokens` inputs are **rejected** at the tool boundary (a tool error,
   like an over-length title), never silently coerced — so the corpus never
   accretes a `cost:abc` that the eventual column promotion would choke on.

## Context

PROJ-004's brief flags a time-sensitive seed under Dependencies→Enables:

> *Consider seeding early:* a minimal MCP `brag_add` `tokens:`/`cost:` capture
> (capture-time, self-reported) could slip into v0.3.x/v0.4.0 so cost history
> starts accruing *before* the economics layer exists — the same lesson
> provenance just taught (the corpus had **0** agent-authored history because
> we stamped late).

The lesson is concrete: history only accrues going forward. When provenance
landed in v0.3.0, every pre-v0.3.0 entry was permanently un-attributable
because it was stamped late. The economics story (PROJ-005) will want per-work
cost/token/session data; if we wait until PROJ-005 to start capturing it, the
corpus is empty in hindsight exactly as the provenance corpus was.

**What bragfile can and cannot capture (settled at framing):**

- bragfile **cannot self-count tokens or cost.** It is a local SQLite CLI/MCP
  server; it has no view of the model's usage accounting.
- The **stdio MCP transport exposes no session id** — the SDK's
  `Implementation` carries only `Name/Title/Version/WebsiteURL/Icons` (verified
  at DEC-024 design; this is exactly why `model:` had to be an explicit param).
  So a session id, like a model id, **cannot come from the transport** and must
  be a caller-supplied param.
- Therefore the honest payload now is a **reliable session JOIN-KEY** the
  caller forwards, plus **optional** real cost/tokens when the caller has them.
  Exact-token reconciliation — joining `session:<id>` against the model
  provider's usage logs to recover authoritative per-entry token/cost figures —
  is **explicitly PROJ-005**, so stringly-typed aggregation now is acceptable.

**Join-key delivery path (verified in code).** The plugin Stop hook
`plugin/hooks/capture-nudge.sh` already reads `session_id` from its stdin
payload and injects agent-facing `additionalContext`; it **never** calls
`brag_add` (Claude does, post-approval, per BRAG.md). So the hook's job is to
**surface** the `session_id` in its `additionalContext` and instruct Claude to
forward it as a `session:` param on `brag_add`. No stdio-transport plumbing is
invented (it does not exist); the id rides the same agent-facing nudge that
already carries the `agent:`/`model:` instruction.

This DEC **extends** DEC-024 (it does not supersede it): the SDK, the stdio
subcommand, the transport-purity rule, and the `agent:`/`model:` provenance all
stand; DEC-027 only widens the reserved namespace and the `stampProvenance`
inputs.

## Alternatives Considered

- **Option A: a numeric `cost` / `tokens` column now (schema migration).**
  - What it is: promote cost/tokens to first-class typed columns on `entries`
    with a forward migration, and a DEC-011 envelope extension.
  - Why rejected: the same accepted-debt→normalize path tags and provenance
    both took (DEC-004→DEC-015; DEC-024 Option C). The seed's whole value is
    *starting to capture now, cheaply*; a migration is the opposite of cheap and
    is not earned until there is a real query/reporting need. Reserved tags
    satisfy `brag list --tag session:<id>` and `brag tags` counting today with
    zero schema change. Promotion is the deferred "later, if earned" step
    (PROJ-005, where reconciliation lives). Rejecting a migration now is the
    load-bearing consistency call with the repo's own philosophy.

- **Option B: fabricate/estimate token counts in bragfile.**
  - What it is: have the MCP server estimate tokens from entry length, or stamp
    a placeholder cost.
  - Why rejected: bragfile has no authoritative view of usage; a fabricated
    number is worse than no number — it pollutes the very dataset PROJ-005 will
    trust, and it violates the honest-degradation posture DEC-024 set (the
    reason `model:` degrades to no-tag rather than guessing). All three inputs
    stay optional; empty → no tag.

- **Option C: derive the session id from the transport (auto-stamp).**
  - What it is: read the session id off the MCP session like `agent:` reads
    `clientInfo.Name`.
  - Why rejected: **impossible** — the stdio transport carries no session
    identity (verified at DEC-024 design; the SDK `Implementation` has no such
    field). This is the exact shape of DEC-024's Option D rejection for
    `model:`. The id must be an explicit caller-supplied param, surfaced to the
    caller by the hook.

- **Option D: expose the seed on the `brag add` CLI as well as MCP.**
  - What it is: add `--session` / `--cost` / `--tokens` flags to `brag add`.
  - Why rejected (recommendation; see Consequences): the seed's *provenance*
    story is agent-driven — the session id comes from the MCP client's session,
    surfaced by the hook to the agent, and the whole point is capturing the
    *agent-assisted* work's economics. A human typing `brag add` at a terminal
    has no session id and no per-entry token count to report. Adding three CLI
    flags widens the CLI surface (help text, tests, docs, the DEC-012
    `brag add --json` schema) for a path with no natural data source. Keep the
    seed **MCP-path-only**, matching the plan; the CLI `brag add --json` schema
    is unchanged. (Revisit if a real CLI use appears.)

- **Option E (chosen): three optional reserved tags on the MCP `brag_add`
  path, migration-free, MCP-only, with the session id surfaced by the hook.**
  - Why selected: it rides DEC-024's already-proven `stampProvenance` path with
    ~one-line-per-tag additions; captures the reliable join-key now and real
    cost/tokens when available; fabricates nothing; needs no migration and no
    transport plumbing; and keeps author classification untouched by staying
    out of the `agent:%`/`model:%` clause.

## Consequences

- **Positive:** cost/session history starts accruing *now*, on a v0.3.x patch,
  so PROJ-005's economics layer has a real dataset in hindsight instead of a
  late-stamp gap. The `session:` join-key is the reconciliation anchor for
  exact-token recovery later. Zero schema change; rides DEC-015. `brag list
  --tag session:<id>` groups an agent session's entries today.
- **Negative (accepted debt):** stringly-typed numerics. `cost:12.5` and
  `tokens:18000` are tags, not typed columns, so aggregation is string-parsed
  until the PROJ-005 column promotion (which carries a migration + a DEC-011
  envelope extension + the reconciliation join). This is the deliberate
  DEC-004→DEC-015 acceptance, not a surprise. The normalization rules (below)
  are chosen to keep the eventual promotion clean (reject non-numeric now, so no
  `cost:abc` ever lands).
- **Negative:** `session:`/`cost:`/`tokens:` add to the `brag tags` taxonomy
  surface. Unlike `agent:`/`model:` (a small closed set), `session:` is
  high-cardinality — one tag per session. This is tolerable for a join-key
  (that is its purpose) but is a taxonomy-pollution watch item: if `brag tags`
  becomes dominated by `session:` rows, that is a promotion trigger (see
  Validation), mirroring DEC-024's `agent:`/`model:`-pollution trigger.
- **Neutral:** author classification is unchanged by construction —
  `provenanceExistsClause` stays `agent:%`/`model:%`-only, so a `session:`- or
  `cost:`-only entry classifies as `human` unless it also carries an
  `agent:`/`model:` tag (which agent-authored entries always will, since the
  same `brag_add` call stamps `agent:` by fallback). In practice an
  agent-captured entry gets `agent:` **and** `session:` together, so
  classification is correct without the clause needing to know about `session:`.
- **Neutral:** the hook change is agent-facing text only — it surfaces
  `session_id` (already parsed from stdin) into `additionalContext` and tells
  Claude to forward it. Silent-degradation and once-per-session contracts are
  untouched; the hook still never runs `brag add`.

## Validation

**Right if:** the `internal/mcpserver` tests hold — each of `session`/`cost`/
`tokens`, when supplied, is stamped as its reserved tag (append order after
`agent:`/`model:`); when omitted, no tag (parity with `agent`/`model` today);
non-numeric or negative `cost`/`tokens` is a tool error, not a silent insert;
and — the load-bearing regression guard — a `brag_add` with `session:`/`cost:`/
`tokens:` but no `agent`/`model` does **not** classify as `--author agent`
(`store.go`'s clause stays `agent:%`/`model:%`-only). The hook harness stays
green (the `additionalContext` still contains "brag"; silent-degradation and
once-per-session unchanged).

**Revisit if:** (a) a real reporting need for cost/tokens aggregation appears,
OR the `session:` tags visibly dominate the `brag tags` taxonomy → promote
cost/tokens (and possibly session) to first-class columns (Option A; carries a
migration + DEC-011 envelope extension + the PROJ-005 reconciliation join);
(b) exact-token reconciliation (join `session:<id>` → provider usage logs)
becomes a real ask → that is the PROJ-005 economics work this seed enables;
(c) a genuine CLI use for `--session`/`--cost`/`--tokens` appears → reconsider
Option D (MCP-only is the current call); (d) the USD-decimal / integer
normalization proves too narrow (e.g. a non-USD currency need) → widen the
format, ideally at the column-promotion boundary rather than by re-encoding the
tag.

## References

- Related specs: SPEC-046 (emits + implements this DEC — the seed capture on
  the MCP `brag_add` path + the hook `session_id` surfacing; PROJ-004 /
  STAGE-014, a v0.3.x patch), SPEC-040 (the provenance path this extends),
  SPEC-041 (the plugin + capture-nudge hook this edits), the future
  `brag impact` spec (STAGE-011) and PROJ-005 economics work (the
  reconciliation consumer)
- Related decisions: **DEC-024** (the reserved-namespace + `stampProvenance`
  path this **extends**, not supersedes — `agent:`/`model:` via explicit params
  with an `agent`/`clientInfo.Name` fallback; the stdio transport carries no
  session/model identity), DEC-015 (polymorphic tags normalization — the new
  tags ride the taggings join with no schema change), DEC-011 (JSON entry
  shape — `brag_add`'s output is unchanged; the new params are inputs only),
  DEC-012 (`brag add --json` schema — **unchanged**; the seed is MCP-only,
  Option D rejected), DEC-004→DEC-015 (the accepted-debt→normalize path this
  follows), DEC-025 (the Claude Code plugin + capture-nudge hook this edits)
- Related constraints: `stdout-is-for-data-stderr-is-for-humans` (blocking —
  the stdio transport purity DEC-024 generalized still holds; the tool errors
  for bad numerics are MCP `IsError` results, never raw stdout), `no-cgo`,
  `errors-wrap-with-context`, `test-before-implementation`; **no** migration so
  `migrations-are-append-only` is not engaged
- External: Model Context Protocol spec (stdio transport; `clientInfo` carries
  no session/model identity — the reason `session:`/`model:` are explicit
  params); `github.com/modelcontextprotocol/go-sdk` v1.6.1
- Discussions: PROJ-004 brief Dependencies→Enables ("Consider seeding early");
  MEMORY note `project_proj004_federation_spike` (economics/token dimension
  = PROJ-005; latest-only + stringly-typed-now settled)
