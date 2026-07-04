---
# Maps to ContextCore epic-level conventions.
# A Stage is a coherent chunk of work within a Project.
# It has a spec backlog and ships as a unit when the backlog is done.

stage:
  id: STAGE-009                     # stable, zero-padded, repo-global (never reused)
  status: active                    # activated 2026-07-03; SPEC-038 shipped, SPEC-039..041 pending
  priority: high                    # v0.3.0-cutting stage; the project's committed core
  target_complete: 2026-07-08       # activated 2026-07-03; ~3 working days (Jul 4 holiday weekend intervening)

project:
  id: PROJ-003                      # parent project
repo:
  id: bragfile

created_at: 2026-07-03
shipped_at: null
---

# STAGE-009: Agent-native spine + capture delight (v0.3.0 core)

## What This Stage Is

When this stage ships, **bragfile has an agent-native write spine and
capture is delightful and correct** — and it is released as v0.3.0.
Concretely: `brag mcp serve` runs a local stdio MCP server exposing
`brag_add` / `brag_list` / `brag_search` / `brag_stats` as thin wrappers
over the existing `Store`, so an MCP-client agent captures and recalls
brags via native tool calls with no shell and no network; the MCP
`brag_add` tool stamps the caller's agent + model as reserved-namespace
tags (`agent:<name>` / `model:<id>`) so agent-driven work is attributable
later, with zero schema change; `brag add` prints one celebratory
TTY-only stderr line on crossing a total/streak/per-project milestone
(silent under `--json`/pipes); the `brag stats` current-streak reads
correctly (alive through yesterday, bucketed by local day) so the streak
milestone fires on a right number; and all three agent surfaces — the MCP
server, the `/brag` slash-command, and a quiet session-end capture-nudge
hook — bundle into an installable Claude Code plugin that documents the
provenance convention. The stage closes with the v0.3.0 cut per §4. This
is the whole committed 3-day core of PROJ-003; the impact *read* surface
is a separate stretch stage (STAGE-010).

## Why Now

This is the first and only committed stage of PROJ-003, so "why now" is
"why this shape, in this order":

1. **The delight depends on the correctness.** Milestone notifications
   (SPEC-039) include a streak milestone, and the current-streak metric
   is a confirmed defect (reads 0 before the day's first entry). Shipping
   the notification on a broken streak would fire celebrations on wrong
   numbers. So SPEC-038 (streak fix) lands first and blocks SPEC-039.
2. **The small specs de-risk the variant flip.** Day 1's SPEC-038/039 are
   XS–S, well-bounded, and test-heavy — the safe place to shake out the
   `claude-plus-agents` variant (if the coordinator flips it) before the
   L-risk MCP spine.
3. **The MCP server is the headline and the highest risk.** SPEC-040
   introduces a new top-level Go dependency (a DEC), a new transport, and
   a subcommand-vs-binary question. Time-boxing it to Day 2 with a hand-
   rolled-stdio fallback keeps the stage shippable.
4. **The plugin makes the spine usable and is the natural closer.** Once
   the MCP server exists, SPEC-041 bundles it with the slash-command and
   the capture-nudge hook into one installable thing, documents the
   provenance convention, and the stage cuts v0.3.0.

No external blockers at the repo level. The one new dependency (a Go MCP
SDK) is introduced and justified inside SPEC-040. The Homebrew tap and
release secrets carry from PROJ-001/002.

## Success Criteria

Concrete and re-verifiable at stage close:

- **`brag mcp serve` runs a local stdio MCP server** exposing `brag_add`,
  `brag_list`, `brag_search`, `brag_stats` as thin wrappers over the
  existing `Store` (SQL stays in `internal/storage`); local-only, no
  network; the MCP protocol stream is never polluted by human-facing
  output (the stdout-is-data spine, enforced at the transport as strictly
  as the CLI enforces it — see Design Notes on the stdio split).
- **MCP `brag_add` stamps provenance.** A brag written via the MCP tool
  carries `agent:<name>` and `model:<id>` reserved tags; `brag list --tag
  model:<id>` filters and `brag tags` counts them; **no migration**.
- **Milestone line on `brag add`.** Crossing a total/streak/per-project
  threshold prints one celebratory line to **stderr, TTY-only**; `brag
  add --json` and non-TTY/piped invocations emit nothing extra (asserted
  with the §9 split-buffer `errBuf.Len()==0` shape).
- **Current-streak is correct.** `brag stats` keeps the streak alive
  through yesterday and buckets by local day; a run ending yesterday with
  `now`=today reports `Current` = run length, not 0. Storage timestamps
  stay UTC RFC3339 (the constraint is untouched); only the derived metric
  goes local.
- **Installable Claude Code plugin.** One plugin bundles `brag mcp serve`
  + the `/brag` slash-command + a quiet, skippable session-end/Stop
  capture-nudge hook; its manifest was pre-flighted against the current
  plugin loader at design (§12(b)); the shipped hook + slash-command +
  BRAG.md document the reserved `agent:`/`model:` convention.
- **v0.3.0 tagged and released per §4.** Optional `v0.3.0-rc1` smoke →
  `v0.3.0` per the dual-tag-on-same-commit rule; Homebrew formula bumped;
  clean `brew upgrade` from v0.2.x verified; the Gatekeeper xattr note +
  the Homebrew 6.0 `brew trust --cask` step are in the release pre-flight.
- **Docs match shipped reality.** `BRAG.md`, `docs/api-contract.md`,
  `docs/architecture.md`, `docs/tutorial.md`, and the CHANGELOG `[0.3.0]`
  describe the MCP surface, the plugin, the milestone behavior, and the
  provenance convention. Each doc-touching spec runs its premise-audit
  greps and enumerates hits under `## Outputs`.
- **No regressions.** All STAGE-001..008 success criteria still hold;
  `go test ./...`, `gofmt -l .`, `go vet ./...`, `CGO_ENABLED=0 go build
  ./...` clean through every spec.

## Scope

### In scope

- The streak-correctness fix in `internal/aggregate` (+ its test-gap
  case) and any local-day seam it needs.
- The milestone-notification path on `brag add` (threshold detection off
  the post-insert aggregate; TTY/stderr gating).
- The `brag mcp serve` subcommand + a local stdio MCP server + the four
  tool wrappers + provenance stamping in the MCP `brag_add`.
- The Claude Code plugin: manifest, bundling of the MCP server +
  slash-command + capture-nudge hook, and the provenance-convention docs
  in the shipped assets.
- The v0.3.0 release cut (CHANGELOG, RC/final tags, tap bump, upgrade
  verification) and the doc sweep for the new surface.

### Explicitly out of scope

- `brag impact` / the AI-pipe super-brag / Notion export (STAGE-010
  stretch). `brag_impact` is **not** a core MCP tool.
- First-class `agent`/`model` columns + the DEC-011 envelope extension —
  the "later, if earned" promotion. Core ships the tag convention only.
- Any network/cloud/multi-user MCP mode; WAL + busy-timeout concurrency
  hardening (noted, deferred unless multi-agent dogfooding forces it).
- macOS notarization (v0.2.1); goals; the wider stats/storytelling
  cluster; `brag project` ergonomics polish.
- A schema migration of any kind (the core is migration-free; if a spec
  design nonetheless needs one, that is a design-time surprise to raise,
  not a silent addition — and it re-arms the §12(a) migration-count-bump
  premise audit).

## Spec Backlog

Four specs, one stage, cut as v0.3.0. Sizing carries the same L-watch
discipline STAGE-007/008 used (peel rather than let a spec grow to L).

Format: `- [status] SPEC-ID (cycle) — one-line summary`

- [x] SPEC-038 (shipped 2026-07-03, PR #57) — **XS/S — Current-streak fix.** `Streak()`
      keeps the current streak alive through *yesterday* and buckets by
      the user's *local* day (storage stays UTC; only the derived metric
      goes local). **Design complete (2026-07-03):** spec on
      `feat/spec-038-streak-fix`; **DEC-022 emitted** (local-day derived
      metric — settled surfaced question (b): a small DEC, not just an
      inline note, because it reverses SPEC-020 §6's two *locked* streak
      decisions and records why localizing a derived metric does not relax
      the blocking `timestamps-in-utc-rfc3339`). Failing tests written
      (alive-through-yesterday, still-0-after-two-empty-days, local-day
      bucketing, DST spring-forward guard); the required run-ending-
      yesterday case added; premise audit enumerated the two inverting
      subtests + the doc/help hits; §12(a) expected values verified against
      a reference impl at design (no `time.Sleep`; injected `now`). **Blocks
      SPEC-039.** Awaiting build (fresh session).

- [ ] SPEC-039 (proposed) — **S — Milestone notifications on `brag
      add`.** TTY-only stderr celebratory line on crossing total
      (10/25/50/100/250/500/1000), streak (7/30/100-day), and per-project
      (10th/50th) thresholds, plus a quiet "first brag today/this week."
      Silent under `--json`/non-TTY. Reuses `internal/aggregate`; reads
      the corrected streak. **Depends on SPEC-038.** **Premise-audit
      triggers:** *§9 split-buffer* — the core assertion is that `brag add
      --json` (and any non-TTY path) leaves `errBuf` empty; write the
      separate `outBuf`/`errBuf` test with the `errBuf.Len()==0` check.
      *NOT-contains self-audit (§12)* — grep the spec's own literal
      milestone strings against the "silent under --json" assertion so a
      milestone token can't leak into a data path. *Additive/count-bump* —
      the threshold set is a fixed-shape collection; if any test asserts
      its membership/count, enumerate those sites.

- [ ] SPEC-040 (proposed) — **M/L (the headline) — `brag mcp serve` MCP
      server + provenance.** New `brag mcp serve` subcommand running a
      local stdio MCP server exposing `brag_add` / `brag_list` /
      `brag_search` / `brag_stats` as thin wrappers over the existing
      `Store`; SQL stays in storage. The MCP `brag_add` stamps the
      caller's `agent:<name>` / `model:<id>` reserved tags. **DEC expected
      at design** (the MCP transport + the new top-level Go dependency per
      `no-new-top-level-deps-without-decision`; subcommand-vs-separate-
      binary). Time-box the Go MCP SDK eval; fall back to a hand-rolled
      stdio loop if the SDK doesn't earn its place. **Premise-audit
      triggers:** *new dep* — the go.mod addition fires the warning-level
      constraint; the DEC is the gate. *stdout-is-data at a new
      transport* — the stdio MCP stream carries protocol frames; nothing
      human-facing may leak onto it (the §9 split-buffer rule generalized
      to the transport). *§12(b) design-time pre-flight* — stand up the
      chosen SDK/loop against a real MCP client (or the protocol's own
      conformance harness) at design and confirm the four tools
      round-trip before locking. *Additive* — new subcommand surface; grep
      command-count / help-text assertions.

- [ ] SPEC-041 (proposed) — **S/M — Claude Code plugin packaging + v0.3.0
      cut.** Bundle `brag mcp serve` + `examples/brag-slash-command.md` +
      a session-end/Stop capture-nudge hook (evolving
      `scripts/claude-code-post-session.sh`) into an installable Claude
      Code plugin; document the reserved `agent:`/`model:` provenance
      convention in the shipped hook/slash-command/BRAG.md (transcribe /
      refine the backlog draft literal — literal-artifact-as-spec).
      §12(b)-pre-flight the plugin manifest against the current plugin
      loader. Then cut v0.3.0 per §4 (CHANGELOG `[0.3.0]`; optional
      `v0.3.0-rc1` → `v0.3.0` dual-tag rule; tap bump; `brew trust` +
      Gatekeeper pre-flight; clean-upgrade verification). **Premise-audit
      triggers:** *§12(b)* — run the manifest through the real loader; the
      capture-nudge hook + slash-command are literal artifacts, diff
      against the embedded literals. *status-change (doc)* — the plugin
      changes the integration story from "copy these files" to "install
      this plugin"; grep `BRAG.md` / `docs/` / `README.md` for the SPEC-022
      asset mentions and enumerate each as update-or-stays. *release
      mechanics* — trust-but-verify the "pushed tag / bumped formula"
      claims via `gh release view` / the tap cask read.

**Count:** 1 shipped / 0 active / 3 pending

**Complexity check:** four specs, one L-risk (SPEC-040). The plugin +
release cut are bundled in SPEC-041 because they share the "make the
spine installable and shipped" seam; if SPEC-041 reads L at design (the
manifest pre-flight surprises, or the release cut wants its own runbook
spec as SPEC-037 did), peel the v0.3.0 cut into its own S spec rather
than letting SPEC-041 grow — mirroring the STAGE-007 SPEC-029→033 and
STAGE-008 doc/release-split discipline.

## Design Notes

Glue and cross-cutting direction. **No DEC is written at framing, and no
DEC number is pinned** (next free is DEC-022, but the first DEC is
assigned at the owning spec's design — SPEC-040's MCP-dep DEC is the
likely first). The questions below are **surfaced for spec-design time,
not decided here.**

### SURFACED design questions (resolve at spec design; do not decide now)

**(a) Go MCP SDK / dependency + transport + subcommand-vs-binary
(SPEC-040 design — a DEC).** Which Go MCP library (if any) backs the
server, versus a hand-rolled stdio JSON-RPC loop; the transport (stdio is
the local-first default — no network); and whether the server is a `brag
mcp serve` subcommand of the existing binary (recommended starting point
— one binary, one install, shared `Store` wiring) or a separate binary.
Load-bearing because it introduces the wave's one new top-level
dependency (`no-new-top-level-deps-without-decision`) and sets the
transport contract. Recommendation to weigh at design: subcommand + stdio
+ time-boxed SDK eval with a hand-rolled-loop fallback.

**(b) Streak local-day vs UTC semantics (SPEC-038 design — possibly a
small DEC).** The fix computes the streak in the user's *local* day while
storage timestamps stay UTC RFC3339 (`timestamps-in-utc-rfc3339` is
blocking and untouched — only the derived metric localizes). Confirm this
split is a clean "derived-metric" carve-out that doesn't need to relax the
constraint (it shouldn't), and whether the local-day boundary is worth a
small DEC or just a documented design decision + the injected-clock seam
(§9). Watch the DST edge.

**(c) Claude Code plugin manifest format (SPEC-041 design — §12(b)
pre-flight).** The plugin manifest's exact shape depends on the *current*
Claude Code plugin loader, which is external and moving. Treat the
manifest as a literal artifact and **run it through the real loader at
design** before locking (the §12(b) discipline that caught the goreleaser
key-rename and the cobra bash-marker drift). Do not transcribe a manifest
shape from memory.

**(d) The Stop/session-end capture-nudge hook UX (SPEC-041 design).** The
hook must be **quiet, skippable, and non-annoying** — it nudges, it does
not nag, and it never auto-posts (BRAG.md's approval loop is
non-negotiable). Open shape questions: does it fire on every session end
or only when the session plausibly shipped something; how a user
silences it (env var / config / uninstall); how it degrades on non-TTY.
The existing `scripts/claude-code-post-session.sh` is the starting point
(it already honours the approval loop and the stdout-is-data-by-example
split) — evolve it, don't replace the contract.

**(e) Agent/model provenance — reserved-tag convention vs first-class
columns (SPEC-040 emits, SPEC-041 documents).** Confirm the
reserved-namespace convention (`agent:<name>` / `model:<id>`, e.g.
`agent:claude-code`, `model:claude-opus-4-8`) emitted by the MCP
`brag_add` tool and documented in the shipped hook/slash-command/BRAG.md
— **vs.** deferring nothing and instead promoting straight to first-class
`agent`/`model` columns now. Recommendation (per the framework's
own "don't normalize until a second consumer / real query need"
philosophy, the same path DEC-004→DEC-015 took for tags): **reserved
tags now, first-class columns later if earned.** Two sub-questions to
settle at design:
  1. **Auto-stamp vs explicit params.** Does `brag_add` *auto-stamp*
     provenance from the caller's MCP context (stronger — can't be
     forgotten) or take `agent`/`model` as explicit tool params (simpler,
     but forgettable)? This depends on **what identity the chosen MCP
     transport actually exposes to the tool** — settle it alongside
     question (a), because the transport choice determines whether
     auto-stamp is even possible.
  2. **The reserved-namespace literal.** Transcribe/refine the backlog's
     draft literal (backlog.md "Reserved-namespace convention — draft
     literal") as the §12 literal artifact: lowercase, no spaces;
     `model:` uses the canonical AGENTS.md model ID; normally one
     `agent:` + one `model:` per entry; `agent`/`model` are RESERVED
     prefixes, never topic tags; auto-populated, rarely hand-typed.
  - **Note the promotion path + revisit trigger:** first-class columns
    follow the DEC-004→DEC-015 accepted-debt→normalize pattern (a DEC +
    migration extending the DEC-011 JSON envelope, 9→11 keys, versioned).
    Revisit trigger: provenance filtering/reporting becomes a real ask,
    OR `agent:`/`model:` tags visibly pollute the `brag tags` taxonomy.

### The stdout/stderr spine at a new transport (cross-cutting)

The MCP stdio server is a *new* place the stdout-is-data /
stderr-is-humans discipline must hold: MCP protocol frames own stdout,
and *nothing* human-facing may leak onto that stream (a stray log line
corrupts the protocol). SPEC-040 should carry the §9 split-buffer test
shape generalized to the transport. The milestone line (SPEC-039) is the
mirror case on the CLI side — it must stay off stdout and off the
`--json`/non-TTY paths. Both specs are testing the same spine at two
surfaces; keep the assertion shape consistent.

### Variant flip (coordinator decision — surfaced, not owned here)

PROJ-003 is the long-deferred `claude-plus-agents` variant test. The
recommendation (see brief Dependencies) is to flip `.variant` and wire
the architect/implementer/reviewer roles before SPEC-038 build, using
Day-1's small specs as the shakeout. This is a `spec-driven-template` /
coordinator call, **not a scope item this stage decides**; it brings the
trust-but-verify-agent-push-reports WATCH item (N=2) into live relevance
for every SPEC-040/041 "pushed/bumped" claim.

**RESOLVED — HOLD on claude-only (2026-07-03, coordinator, PROJ-003
Step-1).** The framing recommendation ("do it, ~30-min pre-work") rested
on the flip being a flag-flip-plus-wiring chore. Step-1 investigation
falsified that premise:

1. **The scaffold is gone.** `just init` did `cp -r variants/<chosen>/. .`
   then `rm -rf variants/` (justfile:41-42). There is no
   `variants/claude-plus-agents/` in the repo — no variant-specific
   AGENTS.md, no `FIRST_SESSION_PROMPTS`, no `/handoffs/` directory — to
   flip *to*.
2. **No role tooling exists.** No `.claude/agents/` definitions; no
   handoffs mechanism. `new-spec.sh`'s `if VARIANT = claude-plus-agents`
   branch (scripts/new-spec.sh:49) is a no-op — both arms select the same
   `spec.md` template.
3. **The instruction set is claude-only-specific.** AGENTS.md is titled
   "Claude-Only Variant", §13 is "Session Hygiene (claude-only specific)",
   and §2 states the `agents.architect`/`agents.implementer` frontmatter
   is *informational only* under this variant. Flipping the flag without
   reconciling these would make every spec's `agents.*` frontmatter imply
   separate-agent provenance that §2 says must be inferred from
   commit/session timestamps.

A genuine flip is therefore scaffold-reconstruction + three role-agent
definitions + an AGENTS.md/prompts reconciliation pass — real
`spec-driven-template` infrastructure work, not pre-build prep. A
flag-only "cosmetic flip" was rejected: it makes `.variant` assert a role
model the repo cannot operate.

**Consequences for STAGE-009:** the whole stage runs `claude-only`; the
§13 fresh-session-per-cycle discipline (design → build → verify in
separate sessions) remains the contamination guard. SPEC-038/039 lose
nothing as designs — they only lose their role as a *variant* shakeout,
which required an apparatus that does not yet exist. The
trust-but-verify-agent-push-reports WATCH item stays parked at N=2 (no
agent push-reports to verify while claude-only). **Future-flip
prerequisites** (gate any later flip on all three): (a) reconstruct or
re-author the `claude-plus-agents` scaffold incl. `/handoffs/`; (b) author
architect/implementer/reviewer agent definitions + make
`new-spec.sh`/templates actually branch on them; (c) reconcile AGENTS.md
§2/§13 and the doc titles for the two-role model. Tracked as a
separately-scoped chore, not owned by any STAGE-009 spec.

## Dependencies

### Depends on

- **STAGE-008 / PROJ-002 (shipped 2026-06-19).** Released v0.2.0;
  polymorphic `tags`/`taggings` (provenance rides these); first-class
  projects (the per-project milestone + MCP project axis);
  `internal/aggregate` (`Streak`, `MostCommon`, the counts milestones
  read); DEC-021 migration safety belt (guards any future migration —
  though the core adds none).
- **SPEC-022 Claude-integration assets.** `BRAG.md`,
  `examples/brag-slash-command.md`, `scripts/claude-code-post-session.sh`,
  `docs/brag-entry.schema.json` — the convention the plugin formalizes.
- **AGENTS.md §4/§9/§12/§13** as enumerated in the brief.
- **External: one new Go dependency likely (a Go MCP SDK), introduced +
  DEC'd inside SPEC-040.** The Homebrew tap and release secrets carry
  from PROJ-001/002.

### Enables

- **PROJ-003 close** (this is the committed stage; shipping it makes the
  project closeable, modulo the STAGE-010 stretch decision).
- **STAGE-010 (stretch) — the impact read surface**, which reads from the
  agent-native-written, provenance-tagged corpus this stage produces.
- **First-class provenance columns, if dogfooding earns the promotion.**

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
