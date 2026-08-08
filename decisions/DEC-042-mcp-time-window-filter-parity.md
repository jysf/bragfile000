---
# Maps to ContextCore insight.* semantic conventions.

insight:
  id: DEC-042
  type: decision
  confidence: 0.88                   # honest: the structural fork (shared pure
                                     # package vs duplicate vs RFC3339-only) is
                                     # high-confidence and well-evidenced by
                                     # DEC-024's own recorded regret; the
                                     # residual soft spot is the deliberate
                                     # MCP-first `until` asymmetry, which is
                                     # cheap to reverse — see Validation.
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

created_at: 2026-08-07
supersedes: null
superseded_by: null

tags:
  - mcp
  - agent-native
  - time-window
  - package-boundaries
  - parity
---

# DEC-042: MCP time-window filter parity — one shared pure parser, the CLI's grammar verbatim, and an MCP-first `until`

## Decision

`brag_list` gains the four filters `storage.ListFilter` already supports but the
MCP boundary never exposed — `since`, `until`, `day`, `author` — under three
coupled sub-decisions:

1. **Parser home.** `ParseSince` and `ParseDay` move out of `internal/cli` into
   a new pure, stdlib-only package **`internal/timewindow`**, and become
   **pure functions taking an explicit `now`**:

   ```go
   func ParseSince(s string, now time.Time) (time.Time, error)
   func ParseDay(value string, now time.Time) (start, end time.Time, err error)
   ```

   The wall-clock seam stays with each **caller** (`cli.clock`,
   `mcpserver.nowFunc`), never inside the shared package. `internal/timewindow`
   holds no package state.

2. **The MCP time vocabulary is the CLI grammar, verbatim.** `since` and `until`
   accept DEC-008's grammar (`YYYY-MM-DD` | `Nd`/`Nw`/`Nm`); `day` accepts
   DEC-039's (`YYYY-MM-DD` | `today` | `yesterday`, a LOCAL calendar day).
   No RFC3339, no epoch, no MCP-only form. `day` is mutually exclusive with both
   `since` and `until` (it sets the whole window); `author` accepts exactly
   `agent` | `human`.

3. **`until` is MCP-first, deliberately.** `ListFilter.Until` (DEC-035) is
   exposed over MCP but stays **off** the CLI — STAGE-017 dropped `--until` as
   YAGNI because `--day` covers the human need in one concept.

**Rider (folded-in audit item).** `brag_list` and `brag_search` reject
`limit < 0` as a tool error, matching `brag list --limit`'s rejection instead of
silently treating a negative as "unlimited". This is the v0.5.0 audit's
"MCP `list`/`search` negative-`limit` parity" item, folded in here per
STAGE-018's own Design Note ("prefer folding an item into a stage that already
edits the same code").

## Context

SPEC-040 deferred `--since` at the MCP boundary and left the reason in the code
(`internal/mcpserver/server.go:169`): *"ParseSince lives in package cli and
importing it would risk a cli↔mcpserver cycle."* The cycle is real and
one-directional — `internal/cli/mcp.go` imports `internal/mcpserver` to run
`brag mcp serve`, so `mcpserver` can never import `cli`.

The obvious workaround has already been tried in this repo and already been
regretted in writing. DEC-024 lists, under **Consequences → Negative**, that the
DEC-010 search tokenization is duplicated in `internal/mcpserver` because
`buildFTS5Query` is unexported in `internal/cli`, and carries the explicit
revisit trigger *"a third consumer of the DEC-010 search transform appears →
extract the shared query builder."* Repeating that move for the time parser
would be the second instance, on a surface with strictly more edge cases: DST
transitions, and DEC-039's deliberate split between a LOCAL `--day` and a
UTC-midnight bare-date `--since`.

STAGE-019's premise is that an agent consults history **before** it acts. The two
windows that question actually needs are "recent" and "agent-authored." Both
already exist in `storage.ListFilter`; neither is reachable over MCP. Nothing
about this requires new storage — only that the boundary stop dropping fields.

One more thing the move fixes. The parser is impure by construction today:
`ParseSince("7d")` reads `cli.clock()` from inside. STAGE-017/SPEC-068 fixed the
*testability* half (the v0.5.0 audit's L4 item) by routing that read through an
injectable package var. Promoting `now` to a parameter finishes the job — a
shared package with no global state cannot become a place where two callers with
two different clock policies fight over one seam.

## Alternatives Considered

- **Option A: accept RFC3339 (or epoch) at the MCP boundary; no shared code.**
  - What it is: `brag_list` takes `since: "2026-08-01T00:00:00Z"`, parsed inline
    in `mcpserver` with `time.Parse`. No package move, no cycle, ~5 lines.
  - Why rejected: it relocates date arithmetic to the party least able to do it
    correctly and most able to do it confidently-wrong. "Last week" becomes the
    caller's subtraction; "today" becomes the caller's guess at the user's
    timezone — which is precisely the skew DEC-039 exists to prevent (for a PDT
    user, "today" in UTC starts at 17:00 the previous local day). It also breaks
    parity as a *contract*, not just as ergonomics: two surfaces over one store
    would accept two different time grammars, and every doc from then on has to
    say which one it means. Cheapest option; worst contract.

- **Option B: duplicate `ParseSince`/`ParseDay` into `internal/mcpserver`.**
  - What it is: copy the ~60 lines across, add a cross-package drift-guard test
    (the shape `TestProvenanceClassifier_GoPredicateMatchesSQLClause` uses).
  - Why rejected: DEC-024 already ran this experiment on the DEC-010 tokenizer
    and wrote the result down as a cost to be paid back later. A drift guard only
    catches divergence *after* someone edits one copy — it does not prevent the
    edit, and it does not help the reader who finds two `ParseDay`s and has to
    work out which is authoritative. Doing it a second time, on the harder
    surface, converts a noted debt into a pattern.

- **Option C: move the parser but keep an internal clock
  (`timewindow.Clock` as an exported package var).**
  - What it is: a straight file move; both callers share one mutable global seam.
  - Why rejected: an exported mutable global that tests in two packages both
    mutate is a race in waiting — `go test ./...` runs *packages* in parallel and
    only serializes tests within a package, so `cli`'s and `mcpserver`'s
    substitutions would overlap. The two callers also legitimately want their own
    policy: `cli.clock` is local-wall-clock for `--day`, and
    `mcpserver.nowFunc` is documented as deliberately *not* `.UTC()`'d so the
    streak buckets by local day (DEC-022). An explicit `now` parameter gives both
    callers what they need and gives the shared package no state to race on.

- **Option D: expose `until` on the CLI too (`brag list --until`), for symmetric
  parity.**
  - What it is: add the flag STAGE-017 declined, so both surfaces match exactly.
  - Why rejected: STAGE-017 dropped `--until` **on evidence**, not by oversight —
    the user need was day-scoped, `--day` covers it in one concept, and that
    stage's reflection names "prefer the one-flag/one-concept shape over a general
    primitive when the actual need is narrow" as its lesson. Re-opening a settled
    call to satisfy a symmetry argument, with no new user need, is exactly the
    move that lesson warns against. The asymmetry here is directional and
    principled: an agent *composes* bounded windows programmatically — that is
    what `until` is for — while a human says "yesterday."

- **Option E (chosen): shared pure `internal/timewindow` + the CLI grammar
  verbatim + MCP-first `until`.**
  - Why selected: it removes the structural blocker instead of routing around it,
    makes grammar parity true *by construction* (it is literally the same
    function, not two functions kept in agreement), leaves the shared package
    with no global state, and costs one mechanical package move with zero
    behavior change to any existing surface.

## Consequences

- **Positive:** `brag_list` becomes usable as memory rather than as a dump —
  "the last 7 days on this project, agent-authored" is one call instead of an
  unbounded list the caller post-filters. This is the minimum viable version of
  STAGE-019's premise and it lands before any of the harder work.
- **Positive:** grammar parity is structural. There is no drift to guard against
  because there is nothing to drift from.
- **Positive:** the migrated tests get *simpler* — a pure `ParseDay(value, now)`
  needs no clock substitution at all, so the seam-swapping in
  `internal/cli/since_test.go` disappears rather than moving.
- **Positive:** retires the SPEC-040 deferral note and clears a blocker
  STAGE-019's later specs (and any future non-`cli` consumer) would hit again.
- **Negative:** the move touches four `cli` files (`since.go` removed,
  `list.go` / `window.go` / `export.go` call sites, plus a new home for the
  `clock` var) and relocates a test file. Real churn in files unrelated to the
  feature. Mitigated by there being **no behavior change**: every existing `cli`
  test must pass unchanged, which is the acceptance bar.
- **Negative:** `until` on MCP and not on the CLI is an asymmetry a reader can
  trip over. Mitigated by documenting it as deliberate in `docs/api-contract.md`
  and `docs/for-ai-agents.md` (not merely omitting it), and by the pre-authorized
  reversal below.
- **Neutral:** `DEC-028` and two `docs/api-contract.md` lines refer to
  `cli.ParseSince` by its old home. Those are historical records of *which
  grammar was reused*, not of where the code lives; they are not rewritten.
  This DEC is the current pointer.
- **Neutral:** no schema change, no migration, no new dependency, no new storage
  method. Every field being exposed already exists on `ListFilter`.

## Validation

**Right if:** the `internal/mcpserver` in-memory-transport tests prove each of
`since` / `until` / `day` / `author` narrows the returned set exactly as the
equivalent `brag list` invocation does over the same rows; `day`+`since` and
`day`+`until` are tool errors naming the conflict; an invalid `author` value and
a negative `limit` are tool errors rather than silent no-ops; `tools/list`
advertises the new input properties; and **every existing `internal/cli` test
passes unchanged** after the package move (the no-behavior-change bar).

**Revisit if:**
- (a) a human asks for a bounded `list` window on the CLI → add `--until`; the
  asymmetry closes and no new DEC is needed, **this one pre-authorizes it**;
- (b) a third consumer of the DEC-010 search transform appears → extract
  `buildFTS5Query` the same way (DEC-024's revisit trigger (c), now with a
  precedent to copy rather than a judgment call to make);
- (c) MCP callers turn out to want an absolute-instant form after all → add
  RFC3339 as an **additional** accepted form inside `internal/timewindow`, so
  both surfaces gain it together. Never add a form to one surface only — that is
  the failure mode this DEC exists to prevent.

## References

- Related specs: SPEC-072 (emits and implements this DEC), SPEC-040 (shipped —
  created the deferral this retires), SPEC-068 (shipped — `--day` / `ParseDay`
  and the L4 clock-seam fix this builds on), SPEC-073 / SPEC-074 (the rest of
  STAGE-019, which consume the unblocked boundary)
- Related decisions: DEC-024 (MCP SDK/transport/provenance — the deferral's
  origin and the recorded cost of Option B), DEC-008 (`--since` grammar, reused
  verbatim), DEC-039 (local calendar day — the `--day` boundary this carries to
  MCP), DEC-035 (`ListFilter.Until` storage promotion — the field `until`
  exposes), DEC-033 (provenance classifier — what `author` filters on), DEC-022
  (local-day derived metrics — why `mcpserver.nowFunc` is deliberately not UTC),
  DEC-011 (JSON entry shape — unchanged by this)
- Related constraints: `no-sql-in-cli-layer` (blocking — `internal/timewindow`
  is pure stdlib and imports no storage), `no-new-top-level-deps-without-decision`
  (none added), `errors-wrap-with-context`, `test-before-implementation`
- Discussions: STAGE-019 Design Notes ("The import cycle, resolved
  structurally"); STAGE-018 Design Notes (the fold-in rule the `limit` rider
  follows)
