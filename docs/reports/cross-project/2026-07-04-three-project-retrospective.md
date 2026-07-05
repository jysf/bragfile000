# Bragfile — Three-Project Cross-Project Retrospective

**Report date:** 2026-07-04
**Scope:** PROJ-001 (MVP, shipped), PROJ-002 (Projects & Tags, shipped), PROJ-003 (Agent-Native Spine, in progress — STAGE-009: SPEC-038–041 shipped, SPEC-042 / v0.3.0 cut pending)
**Corpus:** 9 stages · 42 numbered specs (40 shipped, 1 deferred, 1 pending) · 25 DECs · 2 releases shipped (v0.1.0, v0.2.0) + 1 pending (v0.3.0) · 181-entry brag DB
**Method:** read-only fan-out over `projects/`, `stages/`, `specs/`, `decisions/`, `AGENTS.md`, `guidance/`, `CHANGELOG.md`, `scripts/`, `git`, and a read-only copy of `~/.bragfile/db.sqlite`. Numbers are directly sourced unless marked *(inferred)*.

---

## Executive summary

Three projects have driven `brag` from an empty repo to a distributed, agent-native CLI across ~11 weeks (2026-04-19 → 2026-07-04). The spec-driven framework held remarkably tight: **40 of 42 specs shipped, one deferred by choice (SPEC-016), one pending (SPEC-042); every project shipped its planned stage set with zero reordering and zero mid-stage cancellations; 25 DECs with zero post-lock drift and exactly one supersession — and that supersession (DEC-004→DEC-015) fired on its own pre-declared triggers.** This is a process operating as designed.

The single most valuable finding is in the **defect-escape distribution**. Product-logic and spec defects are caught early and cheaply — at design (premise audits, §12(b) pre-flight) or build (TDD). Every defect that actually *escaped* a cycle was **operational / runtime**, living in the gap between *"the artifact validates"* and *"the tool behaves correctly in its real runtime environment"*:

| Escaped defect | Project | Root gap |
|---|---|---|
| `brag stats` current-streak reads 0 (timezone/day boundary) | PROJ-003 | logic correct on UTC, wrong on local wall-clock → escaped to prod dogfooding |
| goreleaser dual-tag-on-same-commit (`422 already_exists`) | PROJ-001 | config valid, release runtime rejected it |
| macOS Gatekeeper quarantine on unsigned binary | PROJ-001 | build fine, first-run runtime blocked |
| Homebrew 6.0 `brew trust --cask` gate | PROJ-002 | tap valid, install runtime gated |
| dev binary migrated the **production** `~/.bragfile` | PROJ-002 | isolation discipline bypassed, no backup |
| plugin registered **0 MCP servers** despite `validate --strict` green | PROJ-003 | validation ≠ registration |

The framework's headline invention — the **§12(b) design-time pre-flight** ("run the literal through its target tool before declaring design done") — was born precisely to close this gap, and it works: it caught the SPEC-024 cobra bash-marker mismatch and the SPEC-036 `VACUUM INTO` behavior at *design*. But PROJ-003's MCP-registration bug proves the gap is **narrowed, not closed**: the pre-flight ran `claude plugin validate --strict` (a *shape* validator) but not `claude plugin details` (the *registration* surface), so a manifest that validated still failed to register at runtime. **Validate ≠ registration** is the newest, still-uncodified refinement of the same recurring lesson.

The brag corpus itself is a healthy, self-eating dogfood: **181 entries, 65.7% carry a non-empty impact, a 15-day peak streak, and 85% of entries self-reference the spec-driven workflow.** But the corpus surfaces the clearest STAGE-010 signal: **provenance tagging is essentially absent** (`ai`×4, `claude`×2; zero `agent:`/`model:` tags) even though DEC-024 reserved that exact namespace in SPEC-040 — because the MCP write path only just shipped and nothing writes through it yet.

**Top recommendations** (detailed in §11): (1) codify **validate≠registration** as a §12(b) refinement — the target tool for a runtime-registration claim is the *registration* surface, not the *validation* surface; (2) add a **runtime/operational pre-flight** class to release specs so dual-tag / Gatekeeper / brew-trust / prod-DB-isolation are checklist items, not earned-in-prod gotchas; (3) STAGE-010 should productize the impact digest **and** start emitting `agent:`/`model:` provenance from the MCP path so the dogfood corpus can measure its own agent-native adoption.

---

## Lens A — The framework / process

### A1. Sizing accuracy: planned vs shipped

**Stages per project — every planned stage shipped, none reordered, none cancelled.**

| Project | Stages planned | Stages shipped | Notes |
|---|---|---|---|
| PROJ-001 | 5 (STAGE-001–005) | 5 | shipped in framing order ([`brief.md`](../../../projects/PROJ-001-mvp/brief.md):551–556) |
| PROJ-002 | 3 (STAGE-006–008) | 3 | "no reordering and no stage cancelled" ([`brief.md`](../../../projects/PROJ-002-projects-and-tags/brief.md):319–320) |
| PROJ-003 | 1 committed (STAGE-009) + 1 stretch (STAGE-010) | STAGE-009 in progress | STAGE-010 explicitly *not* part of the v0.3.0 commitment ([`brief.md`](../../../projects/PROJ-003-agent-native-spine/brief.md):205–223) |

**Specs per stage** (shipped counts; `scripts/specs-by-stage.sh` cross-check):

```
STAGE-001  4  ▃    STAGE-006  2  ▁
STAGE-002  8  █    STAGE-007  7  ▇
STAGE-003  4  ▃    STAGE-008  4  ▃
STAGE-004  3  ▂    STAGE-009  4  ▃  (+1 pending SPEC-042)
STAGE-005  4  ▃
                   spec count per stage: ▃█▃▂▃▁▇▃▃
```

**Estimate vs actual** where the stage brief stated a plan (PROJ-002, [`brief.md`](../../../projects/PROJ-002-projects-and-tags/brief.md):174–212):

| Stage | Estimated specs | Shipped | Delta | Driver |
|---|---|---|---|---|
| STAGE-006 | ~3–4 | 2 | **−1 to −2** | atomic single-migration rescope kept SPEC-025 at one L instead of splitting |
| STAGE-007 | ~5–6 | 7 | **+1** | SPEC-029 L-watch fired → peeled location-editing into SPEC-033 |
| STAGE-008 | ~2–3 | 4 | **+1 to +2** | incident promoted migration-safety belt (SPEC-036) in-scope; release cut split from doc/CHANGELOG work |

Sizing bias is small and **self-correcting via the peel discipline** rather than via scope creep: under-estimates get a logged peel, over-estimates get a logged rescope. No stage silently ballooned.

### A2. Peel events

| Peel | Project | Decided at | Rationale (quoted) |
|---|---|---|---|
| **SPEC-029 → SPEC-033** | PROJ-002 | design | "Location editing is peeled into its own spec (SPEC-033) and `edit` here is scalar-only… The peel is **flagged, not silently absorbed**." ([`SPEC-029`](../../../projects/PROJ-002-projects-and-tags/specs/done/SPEC-029-brag-project-edit-archive-delete.md):76–88) |
| **SPEC-041 → SPEC-042** | PROJ-003 | design | "the peel WAS taken … breadth (packaging + a folded-in MCP regression test + a release runbook) plus the structural merge boundary (the release tag is cut from `main` after the plugin PR lands, so bundling would break `one-spec-per-pr`) read L." ([`STAGE-009`](../../../projects/PROJ-003-agent-native-spine/stages/STAGE-009-mcp-plugin-and-capture-delight.md):261–270) |

Both peels were **decided at design, not discovered at build** — the mature form of the pattern. SPEC-041's spec is explicit: "decided explicitly at design, not discovered at build" ([`SPEC-041`](../../../projects/PROJ-003-agent-native-spine/specs/done/SPEC-041-claude-code-plugin-packaging.md):49–54). The release-cut-as-its-own-spec shape (SPEC-037, SPEC-042) is now a repeated precedent.

*One deferral, not a peel:* SPEC-016 (`brag export --format sqlite`) deferred 2026-04-23 because `cp ~/.bragfile/db.sqlite backup.db` already covered the use case; slot reserved, SPEC-017 not renumbered ([`STAGE-003`](../../../projects/PROJ-001-mvp/stages/STAGE-003-reports-and-ai-friendly-i-o.md):239–245).

### A3. DEC lifecycle

**25 DECs, cleanly partitioned by project scope:**

```
PROJ-001  DEC-001–014  14  █
PROJ-002  DEC-015–021   7  ▃
PROJ-003  DEC-022–025   4  ▁
```

The 14→7→4 taper mirrors the work: MVP front-loads decisions, deepening needs fewer, the native-spine fewest.

**Supersessions / deprecations: exactly one, and it fired as designed.** DEC-004 (comma-joined tags, confidence 0.65) → DEC-015 (polymorphic tags, 0.80): "DEC-004's two named revisit triggers — 'tag rename becomes a user ask' and 'a second consumer appears' — were both tripped by PROJ-002 … This is DEC-004 firing as designed, not a reversal." ([`DEC-004`](../../../decisions/DEC-004-tags-comma-joined-for-mvp.md):30–35). No other DEC is superseded or deprecated.

**§14 confidence distribution** (from each DEC's `confidence:` frontmatter):

| Confidence | Count | Bar |
|---|---|---|
| 0.65 | 1 | █ (DEC-004, superseded) |
| 0.70 | 1 | █ (DEC-005, open question filed) |
| 0.80 | 8 | ████████ |
| 0.82 | 3 | ███ |
| 0.85 | 7 | ███████ |
| 0.86 | 1 | █ |
| 0.88 | 1 | █ |
| 0.90 | 2 | ██ |
| 0.95 | 1 | █ (DEC-003) |

Mean **0.823**, range **[0.65, 0.95]**, **17/25 (68%) at ≥0.80**, none at 1.0 (forced candor per §14), only two below 0.75 — both flagged with revisit triggers or open questions.

**Drift between design-lock and ship: none** (target: none). No DEC carries a post-lock revision date or amendment body; DEC-004's supersession is recorded in *both* headers at creation time, not retrofitted. *(One nuance: DEC-025 gained an amendment during SPEC-041's build — the `.mcp.json` registration fact — but that is a build-cycle refinement within the same spec, not post-ship drift; see A4.)*

### A4. Defect-escape analysis *(highest-value)*

Classifying every real defect by the cycle that **caught** it. Product/spec-logic defects are caught early; the defects that *escape* are operational/runtime.

**Escape-stage distribution** (notable defects; counts are of distinct logged defects, *(inferred)* grouping):

```
design-caught   ▓▓▓▓            4   premise-audit / §12(b) pre-flight prevented before build
build-caught    ▓▓▓▓▓▓▓▓        8   TDD / live-check surfaced during build
verify-caught   ▓▓▓▓            4   punch-list at verify
ship-caught     ▓               1   stage/release integrity check
escaped-to-prod ▓▓▓▓▓▓          6   surfaced only in real runtime / dogfooding
```

**The pattern:** design/build/verify catch **spec-logic** defects; **every prod escape is operational/runtime.**

| Defect | Project / spec | Caught at | Class |
|---|---|---|---|
| Cobra `MarkFlagRequired` returns unwrappable error vs `ErrUser` | SPEC-003 | build (→ DEC-007) | logic |
| Premise inverted → delete `TestAdd_MissingTitleIsUserError` | SPEC-010 | **design** (premise audit) | logic |
| `VACUUM INTO` rejects existing dest / needs unique name | SPEC-036 | **design** (§12(b) pre-flight, dep-pinned) | runtime, caught early |
| Cobra bash marker `__start_brag` ≠ assumed `_brag_completion()` | SPEC-024 | **design** (§12(b) pre-flight) | runtime, caught early |
| "either is acceptable" spec-prose defect | SPEC-007 | verify | spec-logic |
| Integer-id resolution untested on mutation cmds | SPEC-029 | verify | coverage |
| goreleaser deprecated `brews:`→`homebrew_casks:` (D3) | SPEC-023 | verify (pre-tracked) | runtime |
| SPEC-019 reflection placeholders never pushed pre-merge | SPEC-019 | **ship** (stage integrity) | process |
| **`brag stats` current-streak = 0 (UTC vs local day)** | SPEC-038 fix | **escaped → prod dogfooding** | runtime |
| **goreleaser dual-tag `422 already_exists`** | SPEC-023 | **escaped → prod ship cycle** | runtime |
| **macOS Gatekeeper quarantine** | SPEC-023 | **escaped → prod first-run** | runtime |
| **Homebrew 6.0 `brew trust --cask`** | SPEC-037 | **escaped → prod install** | runtime |
| **dev binary migrated production `~/.bragfile`** | STAGE-008 | **escaped → prod** (motivated DEC-021) | operational |
| **plugin registered 0 MCP servers despite `validate --strict` green** | SPEC-041 | build live-check (post-build punch-list) | runtime |

**Three anchor cases, in the framework's own words:**

1. **Streak bug — escaped to production.** "an unbroken 14-day run showed `Current: 0` … Root cause in `aggregate.Streak`: the cursor seeds at *today (UTC)* … an evening-Pacific entry lands on what UTC calls the next day." ([`SPEC-038`](../../../projects/PROJ-003-agent-native-spine/specs/done/SPEC-038-current-streak-fix-local-day-alive-through-yesterday.md):35–43). Discovered 2026-06-20 in dogfooding, fixed by SPEC-038 / DEC-022. The correctness bug was invisible to a UTC-based test suite — it needed a real user in a real timezone.

2. **MCP registration — build live-check caught what design pre-flight missed.** "design's §12(b) pre-flight ran the manifest through `claude plugin validate --strict` (the loader's *shape* validator) but not `claude plugin details` (the loader's component-*registration* surface) … a literal that validated strict still failed to register at runtime, and no AC-mapped test caught it because every AC-mapped test asserted JSON shape, not the loader's runtime behavior." ([`SPEC-041`](../../../projects/PROJ-003-agent-native-spine/specs/done/SPEC-041-claude-code-plugin-packaging.md):921–934). Fix: add `plugin/.mcp.json`; DEC-025 amended. **This is the single most important process signal in the whole corpus** — the pre-flight discipline was applied, but against the wrong surface.

3. **SPEC-041 ran build→build-fix→verify→ship.** The build-fix was not a code drift — it was a missing structural artifact (`plugin/.mcp.json`) discovered only when the *installed* plugin was checked against `claude plugin details`. All gates (569 tests, `test-docs`, `test-hook`, `validate --strict`) re-ran green post-fix ([`SPEC-041`](../../../projects/PROJ-003-agent-native-spine/specs/done/SPEC-041-claude-code-plugin-packaging.md):972–992).

**Reading:** the process is *strong* at spec-logic correctness (design premise audits + TDD + verify punch-lists form a dense net) and *structurally weaker* at runtime/operational behavior, because those defects only exist once the artifact meets its real host (a shell, a release runner, a package manager, a plugin loader, a user's timezone). The §12(b) pre-flight is the correct instrument; it just needs to always target the *behavioral* surface, not the *validation* surface.

### A5. WATCH → codification pipeline

The WATCH ledger is the framework's learning loop: a lesson observed in a spec reflection enters the stage WATCH list with an N-count, is carried forward across stage closes, and is promoted into AGENTS.md when it clears the bar.

**The promotion meta-rule** (AGENTS.md §, lines 424–442, quoted): *"paired opposing-outcome cases earn codification at N=2; same-outcome confirming cases still need N=3 … each outcome independently constrains the rule's shape: the negative case proves 'skipping the pre-flight has a cost,' the positive case proves 'doing the pre-flight prevents the cost.'"* Earned at PROJ-001 close (2026-05-17).

**Adherence — every codification we traced obeyed the bar:**

| Codified rule | AGENTS § | Evidence | Bar met |
|---|---|---|---|
| Literal-artifact-as-spec | §12 | SPEC-018/020/021 | N=3 same-outcome ✓ |
| Push-before-merge | §10 | SPEC-013/018/019 | N=3 same-outcome ✓ |
| NOT-contains prose audit | §12 | SPEC-019/020 | N=2 same-outcome → promoted at SPEC-020 ship |
| **§12(b) design-time pre-flight** | §12(b) | SPEC-023 NEGATIVE + SPEC-024 POSITIVE | **N=2 paired-opposing ✓** (canonical) |
| §12(b) test-assertion pre-flight | §12(b) | SPEC-023 L2/O4 + SPEC-025 | N=3 same-outcome ✓ |
| Flag-default explicitness | §12 | SPEC-026/028/029 | N=3 same-outcome ✓ |
| Injectable os-var seam (`getCwd`,`clock`) | §9 | SPEC-031/032/036 | N=3 same-outcome ✓ |

**Stages-to-codify** (latency from first observation to promotion): premise-audit family took 3 specs (SPEC-010→012) within one stage; §12(b) took 2 specs across the STAGE-005 boundary; injectable-os-var took 3 specs spanning STAGE-007→008. **Codification is deliberately lagged** — the framework refuses to codify on first sight, which is why only one supersession and zero rule reversals exist.

**Currently open WATCH items** (below bar, carried into STAGE-009 / PROJ-003):

| WATCH item | N | Bar | Status |
|---|---|---|---|
| **validate ≠ registration** (plugin MCP claim) | 1 | needs N=2 / paired | flagged in DEC-025 amendment, **not codified** |
| premise-audit "stays" rows for scope-excluded residuals | 1 | N=2 | SPEC-038 |
| pure-function-plus-thin-glue shape | 1 | reusable | SPEC-039 |
| test-group-letter collision (extend §9 audit-grep) | 1 | N=2 | SPEC-041 |
| grade-by-intent for doc ACs | 1 | N=2 | SPEC-034 (carried from PROJ-002) |
| design-time pre-flight names exact dep version | 1 | N=2 | SPEC-036 (`modernc.org/sqlite` changed behavior across a minor) |
| trust-but-verify agent-push-reports | 2 | N=3 | parked (claude-only; no agent roles wired) |

### A6. Cycle-discipline incidents

- **Loose output on main:** none logged across all three projects; every spec shipped via PR + squash-merge + `--delete-branch`.
- **Stop-gate overruns:** two. (1) **STAGE-005** ran +6 days past its framing-time boundary, driven by a user-requested pre-distribution security review (280-line report, 7 hardening fixes) plus the Phase-2 manual ship cycle ([`STAGE-005`](../../../projects/PROJ-001-mvp/stages/STAGE-005-distribution-and-cleanup.md):588–607). (2) **SPEC-038** shipped clean but is noted (project memory) as having *shipped past its design gate* — the streak fix's semantics reversal (DEC-022) was settled during the build, not held at the design stop-gate.
- **Fresh-session adherence:** the §13 working-tree-preservation reflex is observed working — a parallel session's uncommitted `docs/framework-feedback/process-feedback.md` was preserved untouched across STAGE-007 and STAGE-008 closes ([`STAGE-008`](../../../projects/PROJ-002-projects-and-tags/stages/STAGE-008-polish-and-v0-2-0-release.md):551–554).
- **Trust-but-verify catches:** the coordinator's "check every 'pushed' claim via `git ls-remote`/PR-state" reflex caught SPEC-019's orphaned reflection commit at stage ship and recovered it via `git show`; the same reflex sits behind the DEC-025-in-truncated-stat scare and the stale-plugin-cache observation in PROJ-003. Held at N=2, parked because the claude-only variant wires no agent roles.

### A7. Premise-audit family ROI

The §9 premise-audit family (extract every test whose *premise* a change invalidates) is the framework's highest-ROI defect catcher. Its four cases and the defects they caught:

| Case | Origin | What it catches | Confirmed catch |
|---|---|---|---|
| **Inversion / removal** | SPEC-010 | tests asserting behavior you're *removing* | deleted `TestAdd_MissingTitleIsUserError` at design |
| **Count-bump / additive** | SPEC-011 | literal count assertions when adding a migration/DEC/constraint | `TestOpen_MigrationsTracked` count-of-1 (a *miss* that earned the corollary) |
| **Status-change** | SPEC-012 | doc references when a feature changes status | tutorial "later stage" line; SPEC-018 missed two adjacent hits (earned audit-grep) |
| **Audit-grep cross-check** | SPEC-018 | design must *run* the greps; build re-verifies deltas | codified at N=3 (SPEC-018/019/020) |

The family is self-improving: two of the four cases were born from a *miss* (SPEC-011 count-of-1, SPEC-018's two adjacent hits), which is why the audit-grep cross-check now demands the greps be executed, not merely described. STAGE-009 proposes a fifth extension — "next-free-letter in an already-lettered harness" — after SPEC-041 locked test-group "K" without grepping `test-docs.sh` for collisions.

### A8. Constraint activity

`guidance/constraints.yaml` carries **11 constraints (9 blocking, 2 warning)**, all repo-global and prophylactic. None carries a "fired" marker in any spec/stage reflection — they are enforced continuously rather than tripped as failure modes. The load-bearing ones trace directly to DECs:

| Constraint | Sev | Anchors to |
|---|---|---|
| `no-cgo` | blocking | DEC-001 pure-Go sqlite; the entire cross-compile/goreleaser distribution story |
| `no-sql-in-cli-layer` | blocking | architecture principle — CLI is a thin shell over `internal/storage` |
| `migrations-are-append-only` | blocking | DEC-002 embedded migrations |
| `timestamps-in-utc-rfc3339` | blocking | *(note: this is the very rule whose UTC-vs-local tension produced the streak bug — the storage layer is correctly UTC; the **derived streak metric** needed local-day logic, which DEC-022 supplied)* |
| `test-before-implementation` | blocking | TDD spine of the whole framework |
| `one-spec-per-pr` | blocking | drove the SPEC-041→042 peel (tag cut from `main` breaks the rule if bundled) |
| `no-new-top-level-deps-without-decision` | warning | forces a DEC before any go.mod addition |

The `timestamps-in-utc-rfc3339` / streak interaction is worth calling out: the constraint is correct and the storage layer obeyed it; the defect lived one layer up, in a *derived* metric that conflated storage-time with display-day. No constraint change is warranted — DEC-022 is the right fix.

### A9. Test growth

Go test count over the window (stage/spec-close anchors):

```
STAGE-006  401  ▁
STAGE-007  531  ▆   (+130, seven specs)
STAGE-008  536  ▇   (+5, SPEC-036 safety belt; doc/release specs add no Go tests)
SPEC-038   541  ▇   (+5)
SPEC-039   555  ▇   (+14)
SPEC-041   569  █   (+14 incl. MCP regression guard)
                 ▁▆▇▇▇█
```
*(SPEC-040 sits between 541 and 569, ~536–540; not separately pinned in its reflection — inferred.)* PROJ-001's early stages report per-spec counts rather than a running total, so the pre-401 baseline isn't a single number; the STAGE-005 distribution phase instead grew a **shell-assertion harness** (`scripts/test-docs.sh`) from 40 → 63 → 96 → 116 cumulative assertions ([`STAGE-005`](../../../projects/PROJ-001-mvp/stages/STAGE-005-distribution-and-cleanup.md):599–602). Two bespoke harnesses now guard non-Go surfaces: `test-docs.sh` (doc/manifest assertions, group-lettered) and `test-capture-nudge.sh` (H1–H7, the session-end hook, with a PATH-shadowing `brag` stub that proves the hook *never* invokes `brag`).

### A10. Release mechanics

| Release | Tag date | DECs of record | Gotcha earned → codified |
|---|---|---|---|
| **v0.1.0** | 2026-05-10 *(stable cut/smoke 2026-05-11)* | DEC-001–014 | **dual-tag-on-same-commit** (`422 already_exists`; recovery = delete RC tag/release first) → AGENTS.md §4; **macOS Gatekeeper** (xattr workaround; notarization deferred) → AGENTS.md §4 |
| **v0.2.0** | 2026-06-17 *(stage ship 2026-06-19)* | DEC-015–021 | **Homebrew 6.0 `brew trust --cask`** (tap-level gate, not notarization) → AGENTS.md §4 |
| **v0.3.0** | pending (SPEC-042) | DEC-022–025 | *anticipated:* CHANGELOG `[0.3.0]`, RC→final dual-tag, tap bump, brew-trust + Gatekeeper pre-flight, clean `v0.2.x→v0.3.0` upgrade (migration-free core), plugin `version` pin |

Every release has contributed exactly the operational gotchas predicted by A4: each was **earned in the real runtime and codified into §4 afterward**. SPEC-042's anticipated-gotchas list ([`SPEC-042`](../../../projects/PROJ-003-agent-native-spine/specs/SPEC-042-v0-3-0-release-cut.md):57–81) is the first release spec that carries *all* prior §4 lessons forward as pre-flight items — evidence the codification loop is closing on release mechanics specifically.

---

## Lens B — The brag product & corpus (read-only DB)

*Source: read-only copy of `~/.bragfile/db.sqlite` (181 live entries; recent ids reach 189, so ~8 lifetime deletions). This is the author's **global** brag DB across all their projects, not just bragfile — which is itself the strongest possible dogfooding signal.*

### B1. Corpus shape

- **181 entries · 5 projects · 214 tags · 581 taggings**, spanning 2026-04-20 → 2026-07-05 (UTC), ~76 days.
- **Entries over time (monthly):** `2026-04: 37 ▃ · 2026-05: 6 ▁ · 2026-06: 121 █ · 2026-07: 17 ▂` → sparkline `▃▁█▂`. May was near-dormant (MVP distribution grind + a 34-day gap); June is the surge (PROJ-002 + heavy multi-project work).
- **Weekly (last 9 weeks):** `32, 6, 1, 4, 7, 28, 55, 31, 17` → `▅▂▁▁▂▅█▅▃`, peaking the week of 2026-06-21 (55 entries).
- **By type:** `shipped 121 (67%) · (untyped) 42 (23%) · milestone 5 · learned 5 · release 2 · planning 2 · …` — a delivery-log character. The 42 untyped entries are mostly early/informal captures *(inferred)*.
- **Streak** *(UTC-derived — local-day per DEC-022 may differ slightly):* 36 distinct active days / 76 (47% coverage); **longest run 15 consecutive days** (2026-06-07→21); **current run 3 days** ending 2026-07-05.
- **Impact density:** **119/181 = 65.7%** carry a non-empty impact, averaging **~278 chars** — substantive, not placeholder. This is a strong base for the STAGE-010 impact digest.

### B2. Dogfooding signal

- **Does the repo brag about itself? Heavily.** The `bragfile` project alone has **43 entries**; **85% of all entries (154/181)** self-reference the spec-driven workflow (mentions of `brag`/`SPEC-`/`MCP`/`plugin`/`milestone`/`streak`). Sample titles: *"shipped SPEC-005: can now type brag add -t"*, *"Shipped FTS5 full-text search index layer with automatic sync triggers"*, *"Shipped the patch lane — the #1 dogfood recommendation."* The corpus is a live, self-eating record of its own construction.
- **Provenance tagging: essentially absent — the clearest gap.** Across 214 tags, provenance-namespace usage is `ai`×4, `claude`×2, `ai-integration`×2, `ai-agents`×2, `ai-tooling`×1, `ai-safety`×1 — and **zero `agent:`/`model:` tags**, even though **DEC-024 (SPEC-040) reserved exactly that namespace** for MCP-written entries. The reason is timing, not neglect: the `brag mcp serve` write path only shipped in STAGE-009, so nothing has written through it with provenance yet. **The corpus cannot yet measure its own agent-native adoption** — which is precisely what STAGE-010 exists to enable.

### B3. Tag taxonomy (top of 214)

| Tag | Uses | | Tag | Uses |
|---|---|---|---|---|
| cli | 50 | | stage-complete | 9 |
| spec-driven | 30 | | architecture | 9 |
| rust | 28 | | process | 8 |
| go | 16 | | image | 8 |
| crustyimg | 14 | | typescript | 8 |
| react | 11 | | engine | 8 |
| ui | 11 | | sqlite | 7 |
| bragfile | 9 | | security / polish / frontend | 6 each |
| framework | 9 | | milestone / projects | 5 each |

The taxonomy is tech-and-workflow shaped (`cli`, `spec-driven`, `rust`, `go`, `framework`, `process`, `stage-complete`) — evidence the corpus is used as an engineering journal, not a status board. `spec-driven`×30 and `stage-complete`×9 are effectively meta-tags the workflow emits about itself.

---

## Findings, ranked

1. **The escape-stage distribution is the core diagnostic: spec-logic defects are caught at design/build; every prod escape is operational/runtime.** The process net is dense on correctness, sparse on runtime-behavior. (§A4)
2. **Validate ≠ registration** — SPEC-041's MCP bug proves the §12(b) pre-flight must target the *behavioral* surface (`claude plugin details`), not the *validation* surface (`validate --strict`). Newest, still-uncodified lesson. (§A4, §A5)
3. **The framework's discipline metrics are near-ideal:** 40/42 specs shipped, all planned stages shipped in order, 25 DECs with one designed supersession and zero drift, mean confidence 0.823 with forced candor. (§A1–A3)
4. **Provenance tagging is absent** despite DEC-024 reserving the namespace — the dogfood corpus can't yet see its own agent-native adoption. (§B2)
5. **Release mechanics are a solved-by-accretion problem:** each of dual-tag / Gatekeeper / brew-trust was earned in prod then codified into §4; v0.3.0's spec is the first to carry them all forward as pre-flight. (§A10)

---

## Recommendations

### To AGENTS.md / templates / constraints

- **R1 — Codify "validate ≠ registration" as a §12(b) refinement.** State the rule generally: *for a claim about runtime behavior (a component registers, a hook fires, a binary is on PATH), the pre-flight target tool is the surface that exercises the behavior, not the surface that validates the artifact's shape.* This has a natural opposing-outcome pair already: SPEC-024 (targeted the behavioral surface → caught at design) vs SPEC-041 (targeted only the validator → escaped to build). That's the **N=2 paired-opposing bar** — it can be codified now rather than waiting for a third same-outcome case.
- **R2 — Add a "runtime/operational pre-flight" section to the release-spec template.** Turn the §4 gotchas (dual-tag recovery, Gatekeeper xattr, brew-trust, dev/prod DB isolation + mandatory backup) from prose lessons into an explicit checklist every release spec (SPEC-042 onward) must tick. SPEC-042 already gestures at this; make it structural.
- **R3 — Promote a `dev-binary-must-not-touch-~/.bragfile` guardrail.** The production-DB migration incident (the motive for DEC-021) was an *isolation-discipline* failure with no code guardrail. Consider a startup check or env-gated refusal when a dev/unreleased build points at the real DB, plus the belt-and-suspenders auto-backup already shipped in SPEC-036.
- **R4 — Extend the §9 audit-grep case list** with the "next-free-letter in an already-lettered harness" rule (SPEC-041's group-K collision) once a second instance lands.
- **R5 — Add a `type` to the milestone/streak write paths** so the 23% untyped-entry share shrinks going forward (schema already supports it; this is a capture-ergonomics nudge).

### What STAGE-010 should prioritize (from the corpus)

- **P1 — Productize the impact digest on the 65.7% of entries that already carry impact.** The data is there and substantive (~278 chars avg). The analyses in this report that STAGE-010 should ship as first-class `brag` queries: **entries-over-time sparkline** (B1), **impact-density report** (B1), **by-project / by-type rollups** (B1), **streak with local-day correctness** (now correct post-DEC-022), and **tag-taxonomy top-N** (B3).
- **P2 — Start emitting `agent:` / `model:` provenance from the MCP path immediately.** DEC-024 reserved the namespace; nothing populates it. Until the MCP write path stamps provenance, the agent-native thesis of PROJ-003 is unmeasurable. This is the highest-leverage corpus change — it turns the dogfood into an instrument.
- **P3 — Add a "dogfooding coverage" query** (self-reference density, provenance share) so the corpus can report on its own agent-native adoption as v0.3.0 lands and the MCP path starts being used.

---

## Action register

All recommendations were accepted (coordinator, 2026-07-04) and are turned into self-contained, routable work items — with ready-to-apply artifacts, acceptance criteria, sizes, dependencies, and framework routing — in **[`2026-07-04-action-register.md`](2026-07-04-action-register.md)**. Headline sequencing: **R1** (codify *validate ≠ registration*, at-bar now) and **R2** (release runtime pre-flight checklist) land in the v0.3.0 cut; **P2** (emit `agent:`/`model:` provenance from the MCP path) is the highest-leverage change; **P1/P3** are STAGE-010's read surface.

## Appendix — machine-readable metrics

See [`metrics.json`](metrics.json) (all figures in this report, keyed), [`specs-by-stage.csv`](specs-by-stage.csv), and [`decs.csv`](decs.csv) alongside this file.

**Source-fidelity note:** every quantitative claim above is directly sourced from the cited file / spec / DEC / git ref or the read-only DB copy, except items explicitly marked *(inferred)* (escape-stage groupings, SPEC-040's exact test count, the untyped-entry interpretation, and the local-vs-UTC streak caveat).
