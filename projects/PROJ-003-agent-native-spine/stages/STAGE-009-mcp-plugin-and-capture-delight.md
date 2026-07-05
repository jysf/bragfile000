---
# Maps to ContextCore epic-level conventions.
# A Stage is a coherent chunk of work within a Project.
# It has a spec backlog and ships as a unit when the backlog is done.

stage:
  id: STAGE-009                     # stable, zero-padded, repo-global (never reused)
  status: active                    # activated 2026-07-03; SPEC-038/039/040/041 shipped; SPEC-042 stub (peeled from 041)
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

- [x] SPEC-039 (shipped 2026-07-04, PR #59) — **S — Milestone notifications
      on `brag add`.** TTY-only stderr celebratory line on crossing total
      (10/25/50/100/250/500/1000), streak (7/30/100-day), and per-project
      (10th/50th) thresholds, plus a quiet "first brag today/this week."
      Silent under `--json`/non-TTY. Pure decision function
      `milestoneLine(milestoneInputs)` + thin glue layer; injectable seams
      (`addClock`, `addStderrIsTTY`); DEC-023 emitted; stdlib
      `os.ModeCharDevice` TTY probe (no new dep).

- [x] SPEC-040 (shipped 2026-07-04) — **M (headline) — `brag
      mcp serve` MCP server + provenance.** New `brag mcp serve` subcommand
      running a local stdio MCP server exposing `brag_add` / `brag_list` /
      `brag_search` / `brag_stats` as thin wrappers over the existing `Store`;
      SQL stays in storage. The MCP `brag_add` stamps the caller's
      `agent:<name>` / `model:<id>` reserved tags. **Design complete:** spec on
      `feat/spec-040-...` (build to open it); **DEC-024 emitted** (official Go
      MCP SDK `modelcontextprotocol/go-sdk` v1.6.1 + `brag mcp serve` subcommand
      + stdio + provenance-via-reserved-tags with explicit params and an
      `agent`/`clientInfo.Name` fallback; confidence 0.85). **§12(b) pre-flight
      RUN at design (all green):** four typed tools round-trip over
      `mcp.NewInMemoryTransports` against a real `mcp.Client`; `tools/list`
      returns exactly the four with SDK-inferred schemas; `brag_add` w/o title →
      `IsError=true`; `Out=any`+`TextContent` gives CLI byte-parity (reuse
      `internal/export`); SDK default logger discards (stdout-clean); pure-Go
      (`CGO_ENABLED=0`). **Surfaced Q(a) resolved:** subcommand+stdio+SDK (not a
      separate binary; hand-rolled loop RETIRED by the clean eval). **Surfaced
      Q(e) resolved:** the MCP transport exposes `clientInfo.Name` (→ `agent:`
      auto-fill) but **no model** (→ `model:` explicit-param only); provenance is
      reserved tags riding DEC-015 with zero schema change. Failing tests written
      (in-memory round-trip = the conformance harness; provenance/tokenizer pure
      tests; the stdout-purity test = §9 split-buffer generalized to the
      transport; cli wiring). **Premise audit enumerated + run:** *new dep* fires
      `no-new-top-level-deps-without-decision` → DEC-024 is the gate; *stdout-at-
      transport* → the purity test; *additive* → grep found **zero** command-count
      / help-list assertions and A5 covers 7 verbs (not `mcp`) → nothing to bump.
      **L-watch:** sized M/L, reads M at design — **no peel** (the pre-flight
      retired both L-drivers; provenance is ~30 validated lines); scope trimmed at
      the edges instead (deferred `--since`, cwd auto-fill, shared-query-builder
      extraction). Doc touch scoped to `docs/api-contract.md` (broad MCP docs are
      SPEC-041's sweep). **Blocks SPEC-041.** Awaiting build (fresh session).

- [x] SPEC-041 (shipped 2026-07-04, PR #62) — **M — Claude Code
      plugin packaging.** Bundle `brag mcp serve` + the `/brag:brag`
      slash-command + a `Stop` capture-nudge hook (evolving
      `scripts/claude-code-post-session.sh`) into an installable Claude Code
      plugin (`plugin/` + repo-root `.claude-plugin/marketplace.json`);
      document the reserved `agent:`/`model:` provenance convention in the
      shipped hook/slash-command/`plugin/README.md`/BRAG.md; fold in one
      MCP-surface regression-guard test (`brag_add` return-value byte-parity
      with `export.ToJSON`). **Design complete:** spec on
      `feat/spec-041-plugin-and-release`; **DEC-025 emitted** (plugin layout +
      MCP-on-PATH + capture-nudge delivery model; confidence 0.8).
      **§12(b) pre-flight RUN at design (all green):** the plugin *and*
      marketplace manifests were built and run through the **real loader**
      (`claude plugin validate --strict`, Claude Code 2.1.201) — **caught a
      real strict-mode drift** (marketplace `description` required for
      `--strict`; fixed in the embedded literal). The `capture-nudge.sh` hook
      was exercised against crafted `Stop` payloads across all fire/silence
      paths + a PATH-shadowing `brag` stub proving it **never invokes `brag`**
      (approval loop held). **Surfaced Q(c) resolved:** manifest pre-flighted,
      not memory-transcribed. **Surfaced Q(d) resolved:** the nudge fires on
      `Stop` (every turn) but **only once per session, only after a commit
      lands** (the "plausibly shipped" answer), delivered as agent-facing
      `additionalContext` — a §12(b) *contract discovery* that the Stop-hook
      surface has **no TTY** (superseding the stage note's "TTY-only" framing).
      **Premise audit run:** *status-change (doc)* grep enumerated every
      SPEC-022 asset mention as update-or-stays (loose assets STAY; BRAG.md +
      README + `plugin/README.md` gain the plugin path). **L-watch: PEEL
      TAKEN** — the v0.3.0 release cut peeled into **SPEC-042** (release tag is
      cut from `main` *after* this PR merges → `one-spec-per-pr`; plus the
      folded-in regression test + the release runbook's distinct kind-of-work,
      per SPEC-037). **Blocks SPEC-042.** **Build+punch-list+verify:** build
      shipped all 9 ACs green; a post-build, pre-verify coordinator-confirmed
      punch-list pass root-caused a real bug — the plugin installed but
      registered **0 MCP servers**, because Claude Code's loader reads MCP
      registration from a separate `plugin/.mcp.json` file, not the inline
      `mcpServers` key in `plugin.json` that `claude plugin validate --strict`
      was checking (a validate-≠-registration gap, flagged WATCH in DEC-025's
      §12(b) note at N=1, not yet codified). Fixed by adding
      `plugin/.mcp.json`; confirmed `MCP servers (1) brag` via a clean
      before/after scratch-marketplace install; DEC-025 amended; a
      regression guard (group-S S12/S12-jq) added. Verify then
      **✅ APPROVED, no punch list**. SPEC-042 unblocked.

- [ ] SPEC-042 (designed 2026-07-05 — build=rehearsal, ship=the cut) —
      **S — v0.3.0 release cut.** Cut/tag/publish v0.3.0 per §4: CHANGELOG
      `[0.3.0]` (authored whole — `[Unreleased]` was empty); optional
      `v0.3.0-rc1` → `v0.3.0` dual-tag rule; Homebrew tap bump; `brew trust
      --cask` + Gatekeeper xattr in the release pre-flight; clean `brew
      upgrade` from v0.2.x verification; the deferred `docs/tutorial.md` +
      `docs/architecture.md` plugin walkthroughs. **Design complete** — the
      §4 runbook, the CHANGELOG `[0.3.0]` literal, observable-end-state ACs
      (incl. the §12(b)-refinement `claude plugin details` registration
      check), the rehearsal framing, and the coordinator/human-gated §4
      Pattern-1 destructive sequence. Adopts the release runtime/operational
      pre-flight checklist (retro R2, template
      `projects/_templates/spec-release-cut.md`). **Blocked on** the STAGE-009
      feature specs on `main` (all merged as of PR #66). Mirrors SPEC-037's
      release-runbook precedent. Awaiting build (rehearsal, fresh session).

- [x] SPEC-043 (merged 2026-07-05, PR #66 — verify ✅ APPROVED — from retro
      P2) — **S — `brag list --author agent|human` provenance filter.** The
      **read half** of the agent-native spine: distinguishes agent-authored
      entries (carrying a reserved `agent:`/`model:` tag, DEC-024) from
      human-authored ones, making PROJ-003's agent-native thesis measurable.
      Added because cross-project-retro P2 ("emit provenance from the MCP
      path") was found already shipped by SPEC-040 (PR #61) — the live gap was
      the query, not the write. Thin filter on `storage.List`
      (`ListFilter.Author` + prefix-anchored `LIKE 'agent:%'/'model:%'`
      `EXISTS`), a validated `--author` cobra flag, migration-free; human
      `brag add` byte-parity kept. **No new DEC** (rides DEC-024 + DEC-015).
      A cross-package round-trip test pins the `mcpserver` stamp literal to the
      `storage` classifier. Unblocks retro P3 (dogfooding-coverage query,
      STAGE-010).

**Count:** 5 shipped (SPEC-038/039/040/041/043) / 1 designed (SPEC-042 → build=rehearsal → cut)

**Complexity check:** **five** specs after the SPEC-041→042 peel; SPEC-040 was
the one L-risk (resized to M after a clean §12(b) pre-flight retired the
SDK/transport risk). The plugin + release cut were bundled in the original
SPEC-041; **at SPEC-041 design the peel WAS taken** — the manifest pre-flight
retired the manifest's L-risk, but the breadth (packaging + a folded-in MCP
regression test + a release runbook) plus the structural merge boundary (the
release tag is cut from `main` after the plugin PR lands, so bundling would
break `one-spec-per-pr`) read L. The v0.3.0 cut is now SPEC-042 — mirroring the
STAGE-007 SPEC-029→033 and STAGE-008 doc/release-split discipline, and
SPEC-037's release-runbook precedent.

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
