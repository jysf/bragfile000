# Action Register — Three-Project Retrospective

**Companion to:** [`2026-07-04-three-project-retrospective.md`](2026-07-04-three-project-retrospective.md)
**Status:** all recommendations accepted (coordinator, 2026-07-04). This register turns each finding into a self-contained, routable work item — every item states the change, the target file/section, a ready-to-apply artifact where feasible, acceptance criteria, size, dependencies, and how it routes through the spec-driven framework.
**Guardrail:** this document does **not** itself edit `AGENTS.md`, `constraints.yaml`, or templates — those are framework changes that must land through the normal design→build→verify→ship cycle (and, for codifications, the WATCH-bar meta-rule). This register is the actionable input to that cycle.

## Priority & sequencing at a glance

| ID | Title | Type | Size | Priority | Route | Depends on |
|---|---|---|---|---|---|---|
| **R1** | Codify *validate ≠ registration* (§12(b) refinement) | AGENTS.md codification | XS | **P0 — at-bar now** | STAGE-009 close / SPEC-042 ship reflection | — |
| **R2** | Release runtime/operational pre-flight checklist | template + §4 | S | **P0** | release-spec template; apply in SPEC-042 | — |
| **R3** | Dev-binary DB-isolation guardrail | DEC + spec | M | **P1** | new DEC + spec (STAGE-009 or 010) | DEC-021 (already shipped) |
| **P2** | Emit `agent:`/`model:` provenance from MCP path | spec | M | **P1 — highest leverage** | STAGE-009/010 spec | DEC-024, SPEC-040 (shipped) |
| **P1** | Productize impact digest (queries) | stage of specs | L | **P1** | STAGE-010 | streak correctness (DEC-022, shipped) |
| **R5** | Set `type` on milestone/auto writes | small spec / drive-by | S | **P2** | STAGE-009/010 spec | SPEC-039 (shipped) |
| **P3** | Dogfooding-coverage query | spec | S | **P2** | STAGE-010 | P2 (provenance emission) |
| **R4** | Extend §9 audit-grep to lettered harnesses | WATCH → codify | XS | **P2 — below bar (N=1)** | WATCH ledger; promote at N=2 | second instance |

Recommended order: **R1 + R2 now** (both at-bar / zero-dependency, both land naturally in the v0.3.0 cut) → **P2 + R3** (make the agent-native path measurable and safe) → **P1 + P3 + R5** (STAGE-010 read surface) → **R4** parks in WATCH until a second instance lands.

---

## R1 — Codify "validate ≠ registration" as a §12(b) refinement  `P0 · XS · at-bar`

**Finding addressed:** #2 — the highest-value process signal. SPEC-041's plugin registered **0 MCP servers** despite `claude plugin validate --strict` returning green, because the §12(b) pre-flight targeted the artifact's *shape validator*, not the loader's *registration surface*.
**Evidence:** [`SPEC-041`](../../../projects/PROJ-003-agent-native-spine/specs/done/SPEC-041-claude-code-plugin-packaging.md):921–934, 972–992; [`DEC-025`](../../../decisions/DEC-025-claude-code-plugin-packaging-and-capture-nudge.md) amendment.
**Why now:** the codification meta-rule (AGENTS.md §, lines 424–442) sets the **paired opposing-outcome bar at N=2**, and this rule already has its pair — SPEC-024 targeted the behavioral surface and caught the defect at design (POSITIVE); SPEC-041 targeted only the validator and the defect escaped to build (NEGATIVE). It clears the bar today; no third case needed.

**Target:** `AGENTS.md` §12 (adjacent to the existing §12(b) design-time pre-flight rule).
**Ready-to-apply artifact** (proposed rule text):

> **§12(b) refinement — target the behavioral surface, not the shape validator.** When a spec's literal makes a claim about *runtime behavior* — a component registers, a hook fires, a binary resolves on `PATH`, a server answers — the design-time pre-flight must run the literal through the tool surface that *exercises that behavior*, not merely the surface that *validates the artifact's shape*. Shape-validation and behavior-registration are different checks; neither substitutes for the other. Canonical pair: SPEC-024 ran cobra's actual `GenBashCompletion` (behavioral) and caught the `__start_brag` marker at design; SPEC-041 ran `claude plugin validate --strict` (shape-only) but not `claude plugin details` (registration), so a manifest that validated still registered zero MCP servers — the defect escaped to build. Earned N=2 paired-opposing (2026-07-04).

**Acceptance criteria:**
- [ ] Rule added to `AGENTS.md` §12 with the SPEC-024/SPEC-041 opposing pair cited.
- [ ] The `validate ≠ registration` WATCH item is removed from the STAGE-009 open-WATCH list (promoted, not carried).
- [ ] SPEC-042's release pre-flight references `claude plugin details` (not just `validate --strict`) as the plugin-registration check.

---

## R2 — Add a runtime/operational pre-flight checklist to the release-spec template  `P0 · S`

**Finding addressed:** #1 and #5 — every production escape was operational/runtime, and each §4 gotcha was earned in prod then codified after the fact. Make them a **checklist that gates the design of every release spec**, so they are ticked, not re-learned.
**Evidence:** AGENTS.md §4 (dual-tag lines 119–124, Gatekeeper 126–131, brew-trust 133–134); [`SPEC-037`](../../../projects/PROJ-002-projects-and-tags/specs/done/SPEC-037-v0-2-0-release-cut.md); [`SPEC-042`](../../../projects/PROJ-003-agent-native-spine/specs/SPEC-042-v0-3-0-release-cut.md):57–81 (already gestures at this — make it structural).

**Target:** the release-spec template (the `SPEC-NNN-vX-Y-Z-release-cut` shape) under the project/spec templates, plus a back-reference from `AGENTS.md` §4.
**Ready-to-apply artifact** (checklist block for the template's "Notes for the Implementer"):

```markdown
### Release runtime/operational pre-flight (all must be ticked at design)
- [ ] Dual-tag-on-same-commit: RC tag + release deleted before the final tag is cut at the same commit (§4 Pattern 1).
- [ ] macOS Gatekeeper: `xattr -dr com.apple.quarantine <bin>` note present in README §Install.
- [ ] Homebrew 6.0+: `brew trust --cask <tap>/<cask>` documented in README and run once at the cut.
- [ ] Dev/prod DB isolation: the RC smoke test runs against a THROWAWAY DB, never ~/.bragfile; the SPEC-036 auto-backup path is observed to fire.
- [ ] Clean upgrade: `brew upgrade` from the prior minor verified; `brag --version` prints the new tag; no migration surprise.
- [ ] CHANGELOG: the `[x.y.z]` dated section is moved out of `[Unreleased]`; compare-links repointed.
- [ ] Plugin version pin (v0.3.0+): `plugin/.claude-plugin/plugin.json` `version` matches the tag.
- [ ] Behavioral surfaces re-checked on the built artifact (see §12(b) refinement): `claude plugin details` shows the MCP server registered; the Stop hook fires in a throwaway repo.
```

**Acceptance criteria:**
- [ ] Checklist added to the release-spec template.
- [ ] `AGENTS.md` §4 gains a one-line pointer: "release specs must include the runtime/operational pre-flight checklist (template)."
- [ ] SPEC-042 adopts the checklist verbatim.

---

## R3 — Dev-binary DB-isolation guardrail  `P1 · M`

**Finding addressed:** the production-`~/.bragfile` migration incident — an isolation-discipline failure with **no code guardrail**; it only motivated the DEC-021 auto-backup after the fact.
**Evidence:** [`STAGE-008`](../../../projects/PROJ-002-projects-and-tags/stages/STAGE-008-polish-and-v0-2-0-release.md):57–65; [`DEC-021`](../../../decisions/DEC-021-migration-auto-backup-durability-model.md).

**Proposed change:** when the binary is a **dev/unreleased build** (version string carries `-dev`/`-dirty` or the release ldflag is unset) **and** the resolved DB path is the real `~/.bragfile/db.sqlite`, refuse to *apply migrations* unless an explicit override is set (e.g. `BRAG_ALLOW_DEV_PROD_MIGRATE=1`). Read-only commands stay unaffected; the SPEC-036 auto-backup remains the belt behind this suspenders.
**Target:** `internal/storage` open/migrate path; a new **DEC** for the policy (confidence-rated per §14) + a spec.
**Acceptance criteria:**
- [ ] DEC filed: what counts as a "dev build", the override mechanism, and the read-only carve-out.
- [ ] Dev build + real DB + pending migration → `Open` returns a wrapped `ErrUser` explaining the override, applies no migration.
- [ ] Released build, or override set, or throwaway DB → unchanged behavior.
- [ ] Hermetic tests (`t.TempDir()`, injected version var) cover both branches; `storage-tests-use-tempdir` respected.

---

## P2 — Emit `agent:` / `model:` provenance from the MCP write path  `P1 · M · highest leverage`

**Finding addressed:** #4 — DEC-024 reserved the `agent:`/`model:` provenance namespace, but the corpus shows **zero adoption** because nothing writes through the MCP path with provenance yet. Until it does, PROJ-003's agent-native thesis is unmeasurable.
**Evidence:** [`DEC-024`](../../../decisions/DEC-024-mcp-server-sdk-transport-and-provenance.md); [`SPEC-040`](../../../projects/PROJ-003-agent-native-spine/specs/done/SPEC-040-brag-mcp-serve-mcp-server-and-provenance.md); corpus provenance counts (`ai`×4, `claude`×2, `agent:`×0, `model:`×0 — see report §B2).

**Proposed change:** the MCP `brag_add` tool stamps reserved provenance tags from the calling context (agent identity + model id) when available, per the DEC-024 convention, so agent-authored entries are self-labeling. Preserve CLI byte-parity for human writes (no provenance stamped there).
**Target:** the MCP server tool handler (`brag mcp serve` / `internal/...`); may need a small spec under STAGE-009 or STAGE-010.
**Acceptance criteria:**
- [ ] An entry added via the MCP `brag_add` tool carries `agent:<id>` and/or `model:<id>` reserved tags when the caller supplies them.
- [ ] Human CLI `brag add` is unchanged (no provenance injected); byte-parity tests still pass.
- [ ] Reserved-tag normalization (SPEC-040) covers the stamped tags.
- [ ] A follow-up corpus check can distinguish agent-authored from human-authored entries. (Feeds P3.)

---

## P1 — Productize the impact digest as first-class queries (STAGE-010)  `P1 · L`

**Finding addressed:** 65.7% of entries already carry substantive impact (~278 chars avg); the read surface is the natural v0.3.x+ payoff and STAGE-010's stated purpose.
**Evidence:** report §B1; STAGE-010 is framed but explicitly out of the v0.3.0 commitment ([`PROJ-003 brief`](../../../projects/PROJ-003-agent-native-spine/brief.md):205–223).

**Queries this retrospective validated as worth shipping** (each already computed here against the live corpus — the report is the working prototype):
1. **entries-over-time** (monthly/weekly sparkline) — report §B1.
2. **impact-density** (% with non-empty impact, avg length) — report §B1.
3. **by-project / by-type rollups** — report §B1/B3.
4. **streak** with **local-day** correctness (now correct post-DEC-022) — report §B1.
5. **tag-taxonomy top-N** — report §B3.

**Target:** STAGE-010 specs (one per query family, or a `brag digest` umbrella command). Route through normal design→ship.
**Acceptance criteria:**
- [ ] Each query above exists as a `brag` subcommand/flag with JSON + human output (per `stdout-is-for-data-stderr-is-for-humans`).
- [ ] Streak uses local-day semantics (DEC-022), not UTC.
- [ ] Golden tests lock output shapes (literal-artifact-as-spec pattern).

---

## P3 — Dogfooding-coverage query  `P2 · S`

**Finding addressed:** #4 — the corpus should be able to report on its own agent-native adoption as the MCP path starts being used.
**Depends on:** P2 (provenance must be emitted before it can be measured).

**Proposed change:** a query reporting self-reference density and **provenance share** (fraction of entries carrying `agent:`/`model:` tags) over time, so v0.3.0's agent-native adoption is visible.
**Acceptance criteria:**
- [ ] Query returns provenance share (agent-authored vs human-authored) and self-reference density, windowed by month.
- [ ] Baseline captured at report date (agent-authored ≈ 0%) so the trend post-v0.3.0 is measurable.

---

## R5 — Set `type` on milestone / auto-generated writes  `P2 · S`

**Finding addressed:** 42/181 entries (23%) are untyped, diluting by-type analytics.
**Evidence:** report §B1 (by-type breakdown).

**Proposed change:** the milestone-notification path (SPEC-039) and any auto-brag write set a `type` (e.g. `milestone`) rather than leaving it null. Schema already supports it.
**Target:** SPEC-039 milestone write path / any auto-write; small spec or disclosed drive-by.
**Acceptance criteria:**
- [ ] Milestone/auto writes set a non-empty `type`.
- [ ] Existing untyped historical entries are left untouched (no back-migration required).

---

## R4 — Extend §9 audit-grep to already-lettered harnesses  `P2 · XS · below bar`

**Finding addressed:** SPEC-041 locked test-group letter "K" without grepping `scripts/test-docs.sh` for existing letters (collision, renamed to "S" at build).
**Evidence:** [`SPEC-041`](../../../projects/PROJ-003-agent-native-spine/specs/done/SPEC-041-claude-code-plugin-packaging.md):940–968.
**Status:** **N=1 — below the same-outcome N=3 bar.** Do not codify yet. Add to the STAGE-009 WATCH ledger as "next-free-letter in an already-lettered harness (extends §9 audit-grep to non-Go artifact classes)"; promote when a second instance lands.
**Acceptance criteria:**
- [ ] WATCH item recorded with N=1 and its source spec.
- [ ] Promotion deferred per the meta-rule; revisited at the next stage close that touches a lettered harness.

---

## Tracking

Each `- [ ]` above is an acceptance criterion. When these are picked up, they should become real specs/DECs/WATCH entries in the primary repo through the normal cycle — this register is their source-of-truth brief. Machine-readable form of the recommendation set is in [`metrics.json`](metrics.json) under `recommendations`.
