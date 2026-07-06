---
project:
  id: PROJ-003
  status: shipped                   # committed core (STAGE-009) shipped as v0.3.0, 2026-07-05
  priority: high
  # activated 2026-07-03 (STAGE-009 selected as core; SPEC-038 shipped)
  target_ship: 2026-07-08           # ~3 working-day core; Jul 4 holiday weekend intervening

repo:
  id: bragfile

created_at: 2026-07-03
shipped_at: 2026-07-05
---

# PROJ-003: Agent-native spine + capture delight

## What This Project Is

Turn bragfile from a CLI you *run* into accomplishment memory that
**agents can write to** and that **surfaces on its own**. Today capture
is a human typing `brag add` (or an agent shelling out to it); retrieval
is a human running a command. PROJ-003's core wave adds the agent-native
write path — a local `brag mcp serve` MCP server exposing brag as typed
tools, plus an installable Claude Code plugin that bundles the MCP
server, the existing slash-command, and a session-end capture-nudge hook
— and makes capture *delightful and correct*: a celebratory milestone
line on `brag add`, riding on a corrected current-streak metric. The
north star the user named is **agent-native accomplishment memory**:
agents do the work *and* record why it mattered, and humans (and agents)
read the stories back out for reviews, promos, and identity. This wave
builds the write spine and the delight; the impact-story *read* surface
(`brag impact`, the AI-pipe super-brag) is named as a later stretch
stage, not the committed core.

The slug is `agent-native-spine`, not `v0.3.0` or `delight`. Following
the PROJ-002 precedent (`projects-and-tags` named the two object shapes
it added), this slug names the surface the wave changes: brag gains a
*spine* — a typed, agent-facing interface (MCP + plugin) that non-shell
and shell agents alike write through — where before it was CLI-only.

## Why Now

Four reasons converge.

1. **The substrate is ready and the record is real.** PROJ-002 shipped
   first-class projects + polymorphic `tags`/`taggings` and a released,
   installable v0.2.0. That substrate is exactly what makes this wave
   cheap: milestone counts/streaks already live in `internal/aggregate`;
   the MCP tools are thin wrappers over the existing `Store`; and
   agent/model provenance rides the normalized tags with **zero schema
   change**. There is now a real v0.2.0 corpus to dogfood against.

2. **Agent-driven capture is where storytelling quality is actually
   won.** The agent that did the work holds title, change, project (cwd),
   tags, and can articulate impact *while context is fresh* —
   capture-time impact beats recalled-weeks-later impact. The user
   already drafts brag content with Claude externally; the SPEC-022
   assets (BRAG.md, the slash-command, the post-session hook,
   `brag-entry.schema.json`) prove the pattern by convention. This wave
   makes it a first-class, installable, typed surface.

3. **Delight has the best effort-to-payoff ratio in the backlog, and it
   is passive.** Milestone notifications fire on an action the user (or
   an agent) already takes — no new command to remember. They were the
   "easy and great" pick of the 2026-06-16 brainstorm. But they are
   *incorrect* without the streak fix: `brag stats` current-streak reads
   0 for the whole part of a day before you re-log (confirmed defect,
   2026-06-20), so a streak-milestone would fire on a wrong number. The
   fix blocks the delight; both are in the core.

4. **This is the long-deferred `claude-plus-agents` variant test.** The
   repo has run `claude-only` for PROJ-001/002. PROJ-003's Day-1 specs
   (streak fix, milestone notifications — small, low-risk, well-bounded)
   are the safe shakeout for flipping `.variant → claude-plus-agents` and
   wiring separate architect/implementer/reviewer agents. This is a
   `spec-driven-template` coordinator decision to confirm at framing (see
   Dependencies), not a scope item this brief decides.

**Operational note — DB state is already v0.2.x.** Unlike PROJ-002, there
is no dev/prod isolation dance this wave: production `~/.bragfile` is
already at the v0.2.0 schema (the DEC-021 safety belt guards any future
migration), bare `brag` is the released v0.2.0 binary, and PROJ-003's
core adds **no schema migration** (provenance rides existing tags; MCP is
a read/write wrapper over the current Store). If the "later, if earned"
promotion of provenance to first-class columns is ever taken up, *that*
carries a migration — but it is explicitly out of the core.

## Success Criteria

Concrete, user-observable, re-verifiable at project close:

- **Agents can write brags through a typed local interface.** `brag mcp
  serve` runs a local stdio MCP server exposing `brag_add`, `brag_list`,
  `brag_search`, and `brag_stats` as thin wrappers over the existing
  `Store`. A Claude Code (or other MCP-client) agent can capture and
  recall brags via native tool calls — no shell required — against the
  same `~/.bragfile/db.sqlite`, with no network boundary.
- **The interface honours the same contracts as the CLI.** MCP tool
  output respects the stdout-is-data spine (the MCP protocol stream is
  never polluted by human-facing chatter); `brag_add` enforces the same
  required-`title` / server-owned-field rules as `brag add --json`; SQL
  stays inside `internal/storage`.
- **Agent + model provenance is captured on agent-driven brags.** A brag
  written through the MCP `brag_add` tool carries the caller's agent and
  model as reserved-namespace tags (`agent:<name>` / `model:<id>`), so
  multi-agent work is attributable in hindsight — `brag list --tag
  model:claude-opus-4-8` filters and `brag tags` counts them, with zero
  schema change.
- **Capture is delightful.** Crossing a total/streak/per-project
  threshold on `brag add` prints one celebratory line — **TTY-only, to
  stderr**, silent under `--json`/non-TTY/pipes so scripted and
  agent-driven capture stays byte-clean.
- **The current-streak metric is correct.** `brag stats` keeps the
  streak alive through *yesterday* and buckets by the user's *local*
  day; an intact multi-day run no longer reports `Current: 0` before the
  day's first entry.
- **bragfile ships as an installable Claude Code plugin.** A single
  plugin bundles `brag mcp serve`, the `/brag` slash-command, and a
  quiet, skippable session-end/Stop capture-nudge hook; installing it
  wires all three into a Claude Code session. The shipped hook,
  slash-command, and BRAG.md document the reserved `agent:`/`model:`
  provenance convention.
- **v0.3.0 ships and upgrades cleanly.** v0.3.0 reaches the public
  Homebrew tap and `brew upgrade jysf/bragfile/bragfile` cleanly moves a
  v0.2.x install forward, following the §4 release mechanics (the RC dual-
  tag rule, the Gatekeeper xattr note, and the Homebrew 6.0 `brew trust
  --cask` pre-flight).
- **No regressions.** All PROJ-001/002 success criteria still hold; the
  full v0.2 feature surface works unchanged on the v0.3.0 binary;
  `go test ./...`, `gofmt -l .`, `go vet ./...` clean.

## Scope

### In scope (the 3-day core — STAGE-009, v0.3.0)

- **Streak correctness fix.** `Streak()` (currently
  `internal/aggregate/aggregate.go`) keeps the current streak alive
  through yesterday and buckets by local day rather than UTC-today. A
  test case exercises the no-entry-yet-today path. Blocks milestone
  notifications.
- **Milestone notifications on `brag add`.** TTY-only stderr celebratory
  line on crossing total (10/25/50/100/250/500/1000), streak
  (7/30/100-day), and per-project (10th/50th) thresholds, plus a quiet
  "first brag today/this week." Silent under `--json`/non-TTY. Reuses
  `internal/aggregate`.
- **`brag mcp serve` — local stdio MCP server.** A new `brag mcp serve`
  subcommand running a local stdio MCP server exposing `brag_add` /
  `brag_list` / `brag_search` / `brag_stats` as thin wrappers over the
  existing `Store`. Local-only, no network. The MCP `brag_add` tool
  stamps agent/model provenance (see below).
- **Agent/model provenance via reserved tag namespaces.** A documented
  convention of reserved tags `agent:<name>` and `model:<id>` (e.g.
  `agent:claude-code`, `model:claude-opus-4-8`), emitted by the MCP
  `brag_add` tool and documented across the shipped plugin assets. Rides
  the polymorphic tags from STAGE-006 — **no schema change**.
- **Claude Code plugin packaging.** Bundle `brag mcp serve` + the
  existing `examples/brag-slash-command.md` + a session-end/Stop
  capture-nudge hook (evolving `scripts/claude-code-post-session.sh`)
  into an installable Claude Code plugin, with a manifest pre-flighted
  against the current plugin loader at design time. Document the
  reserved-namespace provenance convention in the shipped
  hook/slash-command/BRAG.md.
- **v0.3.0 release mechanics.** CHANGELOG `[0.3.0]`; the RC-tag pattern
  from §4 (optional `v0.3.0-rc1` smoke-test under the dual-tag-on-same-
  commit rule); the Gatekeeper xattr note + the Homebrew 6.0 `brew trust
  --cask` step carried into the release pre-flight.
- **Doc sweep for the new surface.** `BRAG.md`, `docs/api-contract.md`,
  `docs/architecture.md`, `docs/tutorial.md` gain the `brag mcp serve`
  command, the plugin install path, the milestone behavior, and the
  provenance convention. Each doc spec runs its premise-audit greps.

### Explicitly out of scope

- **The impact-story read surface (deferred to STAGE-010, stretch).**
  `brag impact --quarter|--month|--year`, the AI-pipe "super-brag"
  quarterly synthesis, and the Notion export adapter. Named in the Stage
  Plan as a *later stage*, explicitly OUT of the 3-day core.
  `brag_impact` is intentionally **not** a core MCP tool — it depends on
  the impact digest that STAGE-010 would build.
- **Promoting provenance to first-class `agent`/`model` columns.** The
  core ships the reserved-tag *convention* only. First-class columns (a
  DEC + migration extending the DEC-011 JSON envelope) are the "later, if
  earned" step, on the same accepted-debt→normalize path tags themselves
  took (DEC-004→DEC-015). Revisit trigger: provenance filtering/reporting
  becomes a real ask, OR `agent:`/`model:` tags visibly pollute the `brag
  tags` taxonomy.
- **Multi-user / cloud sync / any network boundary.** MCP is local stdio
  only. Networked or multi-user capture is a separate concern with its
  own DEC-at-need; the WAL + busy-timeout concurrency question (several
  agents writing at once) is noted but not solved here unless real
  multi-agent dogfooding forces it.
- **macOS code signing + notarization.** Tracked separately as v0.2.1
  "macOS distribution hardening" (external Apple lead time; the user will
  pay the $99 fee). Notarization removes the Gatekeeper prompt but **not**
  the `brew trust` step — distinct frictions. Not coupled to this wave.
- **Goals as a shipped object type; the wider stats/storytelling
  cluster** (`brag wrapped`, `brag achievements`, `brag story`, the
  "so-what ladder", impact density, sparklines). Menu options for later
  stages/projects, informed by v0.3.0 dogfooding. Not built here.
- **`brag project` ergonomics polish** (project→entries shortcut,
  symlink cwd matching, cosmetic `project status` column). Dogfooding
  backlog; not this wave unless a spec has spare room.

## Stage Plan

Two stages. **STAGE-009 is the committed 3-day core and ships as
v0.3.0.** STAGE-010 is an explicitly-stretch follow-on (the impact read
surface) that is framed only if v0.3.0 dogfooding earns it — it is *not*
part of the v0.3.0 commitment and is named here only so the read-surface
work has a home.

Format: `- [status] STAGE-ID — one-line summary`

- [x] STAGE-009 (shipped 2026-07-05, v0.3.0) — **Agent-native spine +
      capture delight.** Delivered as **6 specs**: streak fix (SPEC-038) →
      milestone notifications (SPEC-039) → `brag mcp serve` + provenance
      (SPEC-040) → Claude Code plugin packaging (SPEC-041) → `brag list
      --author` provenance read half (SPEC-043, added from retro P2) →
      v0.3.0 cut (SPEC-042). Plus a coordinator-directed post-core hardening
      spec, SPEC-044 (retro R3 dev/prod-migration guardrail, DEC-026), which
      shipped after the cut (loosely tagged STAGE-010; STAGE-010 was never
      activated as a formal stage).
- [ ] STAGE-010 (stretch — **NOT pursued**) — **Impact story surface.**
      `brag impact --quarter|--month|--year` + the AI-pipe super-brag + a
      Notion export adapter. Never activated. The read-surface ideas — plus
      the drafted SPEC-045 (P3 dogfooding-coverage query) — **carry forward
      to a future project**, decided after the v0.3.x dogfooding + a DuckDB
      federation spike (see Project-Level Reflection). Note: the single-user
      impact read surface is additive features → a **v0.4.0**, not a v0.3.1
      patch.

**Count:** 1 committed stage shipped (STAGE-009, 6 specs) / STAGE-010 stretch not pursued

## Dependencies

### Depends on

- **PROJ-002 (shipped 2026-06-19).** The polymorphic `tags`/`taggings`
  model is what lets provenance ride as reserved tags with no schema
  change; first-class projects give the MCP tools and milestones their
  project axis; `internal/aggregate` already computes the counts/streaks
  milestones read. All DEC-001..021 apply forward unchanged.
- **The SPEC-022 Claude-integration assets.** `BRAG.md`,
  `examples/brag-slash-command.md`, `scripts/claude-code-post-session.sh`,
  and `docs/brag-entry.schema.json` are the convention-level proof the
  plugin packages and formalizes. The plugin bundles/evolves these rather
  than inventing a new surface.
- **AGENTS.md conventions.** §2 repo-global monotonic IDs; §9 testing
  (the injectable-os-var seam for the streak's clock/local-day and the
  MCP server's os-state; the stdout-is-data / split-buffer rules, which
  the MCP transport must honour as strictly as the CLI); §12 +
  §12(b) literal-artifact-as-spec + design-time pre-flight (the plugin
  manifest and the reserved-namespace literal both pre-flight against
  their target tools); §13 fresh-session discipline. The §4 release
  addenda (dual-tag recovery, Gatekeeper xattr, Homebrew 6.0 `brew
  trust`) apply directly to the v0.3.0 cut.
- **External: none new at the repo level; one new Go dependency likely at
  SPEC-040 design.** The MCP server will probably pull a Go MCP SDK — a
  new top-level dep that triggers `no-new-top-level-deps-without-decision`
  and gets its DEC at SPEC-040 design (the eval is time-boxed; fallback is
  a hand-rolled stdio loop). The Homebrew tap and release secrets carry
  from PROJ-001/002.
- **Coordinator decision to confirm at framing (not decided here): flip
  `.variant → claude-plus-agents`.** PROJ-003 is the long-deferred
  variant test; Day-1's small specs (SPEC-038/039) are the safe shakeout.
  Recommendation: **do it**, and wire the architect/implementer/reviewer
  agent roles before SPEC-038 build — but this is a `spec-driven-template`
  / coordinator call, carrying the trust-but-verify-agent-push-reports
  WATCH item (N=2) into live relevance, not a scope item this brief owns.

### Enables

- **The impact story surface (STAGE-010 / later).** A released
  agent-native write spine + a real provenance-tagged corpus is the
  substrate the impact digest, super-brag, and storytelling features read
  from.
- **Provenance as first-class data, if earned.** The reserved-tag
  convention is the accepted-debt phase; real multi-agent dogfooding
  reveals which provenance queries matter before any schema commitment.
- **Non-shell agent surfaces.** The MCP spine reaches agents that can't
  run a shell (sandboxed/desktop/hosted) — the narrow-but-real value the
  CLI can't serve.

## Project-Level Reflection

*Shipped 2026-07-05 — committed core (STAGE-009) released as v0.3.0.*

- **Did we deliver the outcome in "What This Project Is"?** **Yes.** bragfile
  is now agent-native: `brag mcp serve` exposes brag as typed MCP tools, the
  Claude Code plugin bundles the MCP server + `/brag` + a capture-nudge hook,
  agent-written entries self-label with `agent:`/`model:` provenance, and
  `brag list --author` reads that provenance back — with a corrected streak
  and milestone notifications for delight. Released and verified as v0.3.0
  (tap bumped, `brew upgrade` clean, prod migration-free).
- **How many stages did it actually take?** **1 of 2 framed.** STAGE-009 (the
  committed core) shipped as 6 specs (vs a 4-spec framing — the SPEC-041→042
  release peel + the retro-driven SPEC-043). STAGE-010 (stretch) was never
  activated. One post-core hardening spec (SPEC-044 / R3) shipped outside a
  formal stage.
- **What changed between starting and shipping?** The three-project
  cross-project retrospective landed mid-project (2026-07-04) and reshaped the
  tail: it added SPEC-043 (its P2 "emit provenance" was already shipped by
  SPEC-040 — the real gap was the read query) and drove the R1/R2/R3
  process+hardening work, all of which landed in this project's window.
- **Lessons that should update AGENTS.md, templates, or constraints?**
  - **Already codified and then validated by this project's own v0.3.0 cut:**
    R1 §12(b) "validate ≠ registration" refinement (AGENTS.md §12) and R2 the
    release-cut template + pre-flight checklist (`spec-release-cut.md`).
  - **New, uncodified (WATCH, N=1):** don't `--delete-branch` when merging the
    *base* of an open stacked PR — GitHub closes (not retargets) the child and
    it can't be reopened. Retarget the child to `main` first. (Cost PR #67 →
    rebuilt as #68 during the v0.3.0 work.)
  - **Retro R5 closed as won't-do:** "set `type` on milestone/auto writes" has
    no target — bragfile has *no* auto-write path (all `Store.Add` callers are
    deliberate captures; milestones only notify).
- **What did we defer to the next project?**
  - **The single-user impact read surface** (old STAGE-010: `brag impact`, the
    super-brag synthesis, Notion export) + the drafted **SPEC-045 (P3
    dogfooding-coverage query)** and **P1 impact digest**. This is a **v0.4.0**
    feature wave (additive read commands), best re-chartered as its own
    project.
  - **Multi-user / federation** — explicitly out-of-scope here (see Scope).
    The user's "pull many people's/agents'/machines' brags into one place"
    goal will be validated by a **DuckDB federation spike** (separate session);
    the PROJ-004 shape (federated export → warehouse, preserving local-first;
    *not* a shared remote DB, which would supersede DEC-001) is decided after
    the spike returns.
  - First-class `agent`/`model` provenance columns (still the "later, if
    earned" step); macOS notarization (v0.2.1 track); the wider
    stats/storytelling cluster.

### Numbering (at framing)

Highest consumed at framing: PROJ-002, STAGE-008, SPEC-037, DEC-021.
Next free (repo-global monotonic, §2): **PROJ-003**, **STAGE-009**,
**SPEC-038** (STAGE-009's specs are SPEC-038..041), **DEC-022** — the
first DEC is *assigned at emission, not pinned here* (SPEC-040's MCP-dep
DEC is the likely first, but earlier specs may surprise us; per the
standing "don't pin a not-yet-emitted DEC number" rule that bit earlier
briefs twice).
