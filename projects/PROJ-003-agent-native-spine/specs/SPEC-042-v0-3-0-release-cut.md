---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-042
  type: story
  cycle: design                    # designed 2026-07-05; build = rehearsal, ship = the cut
  blocked: true                    # on the STAGE-009 feature specs merging to main (see Dependencies)
  priority: high
  complexity: S

project:
  id: PROJ-003
  stage: STAGE-009
repo:
  id: bragfile

agents:
  architect: claude-opus-4-8
  implementer: claude-opus-4-8     # same Claude, different session (claude-only variant)
  created_at: 2026-07-04

references:
  decisions: [DEC-024, DEC-025, DEC-021]
  constraints: [one-spec-per-pr, timestamps-in-utc-rfc3339]
  related_specs: [SPEC-041, SPEC-043, SPEC-037]
---

# SPEC-042: v0.3.0 release cut

## Context

STAGE-009's feature specs — SPEC-038 (streak fix), SPEC-039 (milestones),
SPEC-040 (`brag mcp serve` + provenance), SPEC-041 (Claude Code plugin), and
SPEC-043 (`brag list --author` provenance filter) — deliver the v0.3.0
surface. This spec **cuts and ships v0.3.0**, the stage's closing action,
following [AGENTS.md §4](../../../AGENTS.md) release mechanics. It was peeled
from SPEC-041 at design (the release tag is cut from `main` *after* the code
PRs land, so it cannot share a PR with the code it tags — `one-spec-per-pr`),
mirroring SPEC-037's v0.2.0 release-runbook precedent.

The design carries the **release runtime/operational pre-flight checklist**
(cross-project-retro R2, PR #64) so every §4 gotcha earned in prod
(goreleaser dual-tag, Gatekeeper, brew-trust, prod-DB migration) is ticked at
design, not re-learned in prod.

## Goal

Cut, tag, and publish v0.3.0 to the Homebrew tap per AGENTS.md §4, verify a
clean `brew upgrade` from v0.2.x, and close STAGE-009.

## Dependencies (blocked-on)

This spec is **blocked** until the v0.3.0 surface is all on `main`:

- **SPEC-043** (PR #66) — `brag list --author`; the last STAGE-009 feature.
- **The retro R1/R2 framework PR** (PR #65) — this spec's design adopts the R2
  checklist/template and the §12(b) refinement AC; land it first.
- SPEC-038/039/040/041 are already merged (PRs #57/#59/#61/#62).

The tag is cut from `main` only after all of the above merge.

## Inputs

- **Files to read:** [AGENTS.md §4](../../../AGENTS.md) (release mechanics +
  the three lessons-earned addenda — dual-tag, Gatekeeper, brew-trust);
  [`SPEC-037`](../../PROJ-002-projects-and-tags/specs/done/SPEC-037-v0-2-0-release-cut.md)
  (the v0.2.0 runbook this mirrors); `CHANGELOG.md` (the empty `[Unreleased]`
  section + the compare-link refs); `.goreleaser.yaml` +
  `.github/workflows/release.yml` (the cut machinery); `README.md` install
  section (Gatekeeper + brew-trust notes); `plugin/.claude-plugin/plugin.json`
  (the version pin, already `0.3.0`).
- **External:** GitHub Releases; the `jysf/homebrew-bragfile` tap.
- **Related code paths:** none — this spec ships **no Go code**.

## Outputs

- **Files modified:**
  - `CHANGELOG.md` — the empty `[Unreleased]` block becomes a dated `[0.3.0]`
    section (literal below); compare-links repointed.
  - `docs/tutorial.md` + `docs/architecture.md` — the plugin walkthroughs
    deferred here from SPEC-041's Outputs; any release-note mention.
  - `jysf/homebrew-bragfile` (separate repo) — `bragfile` cask bumped to
    `0.3.0` with matching sha256s (goreleaser publishes this).
- **New exports:** none.
- **Database changes:** **none** — v0.3.0 is migration-free (STAGE-009 core
  added no migration; the DEC-021 safety belt therefore does not fire on a
  v0.2.x→v0.3.0 open).

## The CHANGELOG `[0.3.0]` literal (transcribe verbatim at build)

Literal-artifact-as-spec (§12). Build replaces the empty `## [Unreleased]`
block with the following (fill the date with the actual cut date), and
repoints the link refs at the file's bottom. The ten-verb Group-O4 assertion
already passes from the `[0.1.0]`/`[0.2.0]` sections and is unaffected.

```markdown
## [Unreleased]

## [0.3.0] - 2026-07-DD

This release makes bragfile **agent-native**. A local MCP server lets an
agent capture and recall brags through native tool calls — no shell, no
network — and agent-written entries label themselves with reserved
`agent:`/`model:` provenance tags. The whole surface installs as a Claude
Code plugin. Capture also gets more delightful (milestone notifications),
and the current-streak metric now reads correctly.

### Added

- `brag mcp serve` — a local stdio MCP server exposing `brag_add`,
  `brag_list`, `brag_search`, and `brag_stats` as native tools over your
  existing database (local-only, no network), so an MCP-client agent
  captures and recalls brags without a shell.
- **Agent/model provenance.** The MCP `brag_add` tool stamps the calling
  agent and model as reserved `agent:<name>` / `model:<id>` tags, making
  agent-authored entries attributable — with no schema change.
- `brag list --author agent|human` — filter entries by provenance
  authorship: `agent` selects entries carrying an `agent:`/`model:` tag,
  `human` selects the rest (`brag list --author agent --format json | jq
  length` counts agent-authored entries).
- **Milestone notifications.** `brag add` prints one celebratory line to
  stderr when you cross a total, streak, or per-project milestone — TTY-only,
  and silent under `--json` and in pipes.
- **Claude Code plugin.** bragfile ships as an installable Claude Code plugin
  bundling `brag mcp serve`, a `/brag` slash-command, and a quiet session-end
  capture-nudge hook; the plugin documents the reserved provenance convention.

### Fixed

- **Current-streak is correct.** `brag stats` keeps the current streak alive
  through *yesterday* and buckets by your *local* day, so it reads correctly
  before the day's first entry (previously it read 0). Storage timestamps
  stay UTC RFC3339; only the derived metric is localized.

### Upgrading from v0.2.x

No manual steps and **no migration** — v0.3.0 adds no schema changes.
`brew upgrade jysf/bragfile/bragfile` moves a v0.2.x install to v0.3.0 in
place; `brag --version` then reports `0.3.0`. Two one-time frictions on
first tap install: on **Homebrew 6.0+**, run `brew trust --cask
jysf/bragfile/bragfile` once; on **macOS**, an unsigned binary may trigger a
Gatekeeper prompt — clear it with `xattr -dr com.apple.quarantine` (see the
README install note). To use the Claude Code plugin, `brag` must resolve on
your `PATH` (the plugin runs the Homebrew-installed binary).
```

Link-reference block at the bottom of `CHANGELOG.md` becomes:

```markdown
[Unreleased]: https://github.com/jysf/bragfile000/compare/v0.3.0...HEAD
[0.3.0]: https://github.com/jysf/bragfile000/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/jysf/bragfile000/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/jysf/bragfile000/releases/tag/v0.1.0
```

## Acceptance Criteria

Each gate is independently checkable by a verifier. Destructive steps are
coordinator/human-gated; the ACs describe the *observable end state* each
produces (mirrors SPEC-037).

- [ ] **Pre-flight all green** (P1–P6 below): `main` clean at the intended
      release SHA, `local == origin`; `CGO_ENABLED=0 go test ./...` green;
      `gofmt -l .` empty; `go vet ./...` clean; `bash scripts/test-docs.sh`
      exits 0; `just test-hook` exits 0; `goreleaser check` passes; `goreleaser
      build --snapshot --clean` succeeds and the host-arch binary reports a
      goreleaser-injected version (NOT `dev`).
- [ ] **CHANGELOG reconciled on the release SHA:** the empty `[Unreleased]`
      block has become the `[0.3.0]` literal above, its date == the actual cut
      date, and the compare-links are repointed (`test-docs.sh` Group O — O5
      `[Unreleased]:` ref present — stays green).
- [ ] **Plugin version pin matches the tag:** `plugin/.claude-plugin/plugin.json`
      `version` == `0.3.0` == the git tag (already `0.3.0`; re-confirmed at
      cut).
- [ ] **rc is a GitHub PRERELEASE with no tap change:** after `v0.3.0-rc1` CI
      completes, its GitHub release is marked *Pre-release* and the
      `jysf/homebrew-bragfile` tap's latest commit is **unchanged**.
- [ ] **RC smoke gate passes on a THROWAWAY DB** (never `~/.bragfile`): the rc
      binary reports `0.3.0-rc1`; `brag mcp serve` starts and answers a
      `tools/list` over stdio; `brag add` (milestone line on a TTY),
      `brag list --author agent|human`, `brag stats` (streak), and the plugin
      install (`claude plugin details brag` → `MCP servers (1) brag`) all
      behave; and — because v0.3.0 is migration-free — opening a seeded
      v0.2.x-schema throwaway DB fires **no** migration and writes **no**
      backup sidecar (DEC-021 trigger `applied>0 && pending>0` is false).
- [ ] **Behavioral surfaces re-checked on the built artifact** (per the §12(b)
      refinement, AGENTS.md §12): against a clean install of the built plugin,
      `claude plugin details brag` shows the MCP server **registered** (not
      just `validate --strict` green), and the Stop hook fires once in a
      throwaway repo after a commit.
- [ ] **Final release has all 4 platform archives + checksums:** the `v0.3.0`
      GitHub release (NOT a prerelease) carries the four `darwin/linux ×
      amd64/arm64` tarballs and `checksums.txt`.
- [ ] **Tap bumped to 0.3.0:** `jysf/homebrew-bragfile` has a new commit
      setting the `bragfile` cask to `0.3.0` with matching sha256s.
- [ ] **Brew binary reports 0.3.0:** after `brew update && brew upgrade
      jysf/bragfile/bragfile` (+ `brew trust --cask` first-time, + the
      Gatekeeper xattr clear), `brag --version` from the brew path reports
      `0.3.0`.
- [ ] **Prod DB opens with no migration:** opening the real `~/.bragfile`
      (already v0.2.x) with the released binary fires no migration and writes
      no new backup sidecar.
- [ ] **Docs match shipped reality:** `docs/tutorial.md` +
      `docs/architecture.md` carry the plugin walkthrough; the doc
      premise-audit greps run clean.
- [ ] **No regressions:** the full Go suite, `gofmt`, `go vet`, and the
      `CGO_ENABLED=0` build stay green (verified at pre-flight; this spec adds
      no code that could break them).

## Failing Tests

This spec ships **no Go code**, so there are no new `*_test.go` files — the
`test-before-implementation` constraint is satisfied the way SPEC-037/034/035
(also code-free) satisfied it: the enforcement mechanism is the **pre-flight
gate suite** (the existing green `go test ./...`, `scripts/test-docs.sh`
including Group O's CHANGELOG-shape assertions, `just test-hook`, and
`goreleaser check`) plus the **post-release verification checklist** (the ACs
above), not new Go assertions. "The test" for a release runbook is the
executable gate, not an assertion in Go. The regression guard is the
pre-existing suite, run unchanged at Pre-flight.

---

## The build cycle is a REHEARSAL (read before building)

The build cycle for SPEC-042 does **not** cut the release. It rehearses the
runbook against throwaway targets so the real cut (coordinator + human
go-aheads) is mechanical:

1. **Author the CHANGELOG `[0.3.0]`** from the literal above (fill the date),
   repoint the link refs, and run `bash scripts/test-docs.sh` — Group O must
   stay green.
2. **`goreleaser check`** — validate `.goreleaser.yaml` against the installed
   goreleaser v2 (§12(b) design-time pre-flight: run the embedded config
   through its tool; SPEC-023 earned this — deprecated keys surface here).
3. **`goreleaser build --snapshot --clean`** — local cross-compile of all four
   targets; confirm a host-arch binary reports a goreleaser-injected version
   (NOT `dev`).
4. **Smoke-sequence dry-run on throwaway tags/DBs** — run the RC SMOKE GATE
   commands (below) substituting the snapshot-built binary for the downloaded
   rc binary and a `/tmp` throwaway DB for `~/.bragfile`. Seed a v0.2.x
   throwaway DB and confirm **no** sidecar is written (migration-free upgrade).
   **No tag is pushed; `goreleaser release` is never invoked; no `gh release`
   call is made.**
5. **A written walkthrough** of the destructive sequence (§4 Pattern 1
   commands, verbatim) left in the spec for the coordinator to execute.

## RC smoke gate (throwaway DB — never ~/.bragfile)

```bash
export BRAGFILE_DB=$(mktemp -d)/smoke.db        # THROWAWAY — never ~/.bragfile
brag --version                                  # → 0.3.0-rc1 (or snapshot version in rehearsal)
brag add --title "smoke" --tags "perf"          # human write; milestone line on a TTY
brag list --author human                        # → the smoke entry
brag list --author agent                        # → empty (no provenance yet)
brag stats                                      # streak reads correctly (SPEC-038)
# MCP: start the server and confirm it answers tools/list over stdio.
brag mcp serve </dev/null &                     # or drive it from an in-editor MCP client
# Plugin registration (the §12(b) refinement check):
claude plugin marketplace add jysf/bragfile000
claude plugin install brag@bragfile
claude plugin details brag                      # → MCP servers (1) brag  (registered, not just valid)
# Migration-free upgrade: seed a v0.2.x-schema throwaway DB, open it, assert
# NO *.backup sidecar appears (applied>0 && pending>0 is false at v0.3.0).
```

## Destructive sequence (coordinator/human-gated — §4 Pattern 1)

The real cut, verbatim, executed only after the rehearsal is green and the
human go-aheads are given:

```bash
# 0. main is clean, at the release SHA, local == origin; CHANGELOG [0.3.0] merged.
# 1. Optional RC to exercise CI + the tap path:
git tag v0.3.0-rc1 && git push origin v0.3.0-rc1     # CI builds a PRERELEASE; skip_upload:auto holds the tap
# 2. Run the RC SMOKE GATE above against the downloaded rc binary. If good:
# 3. Dual-tag-on-same-commit — delete the RC tag + release BEFORE the final tag (§4 Pattern 1):
gh release delete v0.3.0-rc1 --yes --cleanup-tag
git tag v0.3.0 && git push origin v0.3.0             # CI cuts the real release + bumps the tap
# 4. Post-release verification: the AC checklist (archives+checksums, tap 0.3.0,
#    brew upgrade → 0.3.0, prod DB no-migration).
```

## Implementation Context

### Decisions that apply

- `DEC-024` — the MCP server + `agent:`/`model:` provenance convention this
  release ships and the CHANGELOG describes.
- `DEC-025` — the plugin layout + MCP-on-PATH + capture-nudge model; the
  §12(b) *registration* check (`claude plugin details`, not just `validate
  --strict`) is now a codified AGENTS.md §12 rule and a release AC.
- `DEC-021` — the migration auto-backup safety belt; relevant as the *negative*
  check here (v0.3.0 is migration-free, so it must NOT fire).

### Constraints that apply

- `one-spec-per-pr` — the release tag is cut from `main` after the feature PRs
  land; this spec is its own PR.
- `timestamps-in-utc-rfc3339` — untouched; SPEC-038 localized only the derived
  streak metric, and the CHANGELOG says so.

### Prior related work

- `SPEC-037` (shipped) — the v0.2.0 release-cut runbook this mirrors
  (pre-flight → RC prerelease → smoke → dual-tag → tap bump → brew verify).
- `SPEC-041` (shipped, PR #62) — the plugin this release includes; deferred
  its tutorial/architecture walkthroughs to this spec's doc sweep.
- `SPEC-043` (PR #66) — the `brag list --author` filter included in the cut.

### Out of scope (for this spec specifically)

- macOS notarization (still deferred; the Gatekeeper xattr note stands).
- Any code change to the v0.3.0 surface — if the rehearsal surfaces a bug,
  that is a new fix spec, not an edit here.
- STAGE-010 read-surface work (impact digest, dogfooding-coverage query).

## Notes for the Implementer

Gotchas, style preferences, reuse opportunities.

- The `[Unreleased]` section is currently **empty** — the v0.3.0 specs shipped
  without CHANGELOG entries, so the `[0.3.0]` section is authored whole from
  the literal above (not moved from accumulated Unreleased notes).
- `test-docs.sh` Group O asserts: `[Unreleased]:` and `[0.1.0]:` link refs
  present (O5), the ten verbs in backticks (O4 — already satisfied), the
  `## [0.1.0]` heading (O3). Keep all of these; add the `[0.3.0]:` ref.
- The plugin version pin is **already** `0.3.0` — no bump needed unless the
  cut slips the number; if it does, bump `plugin/.claude-plugin/plugin.json`
  to match the tag.

### Release runtime/operational pre-flight (all must be ticked at design)

Adopted from the release-cut spec template (`projects/_templates/spec-release-cut.md`,
per AGENTS.md §4). Concretized for the v0.3.0 `brag` cut.

- [ ] Dual-tag-on-same-commit: `v0.3.0-rc1` tag + release deleted before the
      final `v0.3.0` tag is cut at the same commit (§4 Pattern 1).
- [ ] macOS Gatekeeper: `xattr -dr com.apple.quarantine <bin>` note present in
      README §Install.
- [ ] Homebrew 6.0+: `brew trust --cask jysf/bragfile/bragfile` documented in
      README and run once at the cut.
- [ ] Dev/prod DB isolation: the RC smoke test runs against a THROWAWAY DB,
      never ~/.bragfile; the SPEC-036 auto-backup path is observed to fire.
- [ ] Clean upgrade: `brew upgrade jysf/bragfile/bragfile` from v0.2.x
      verified; `brag --version` prints `v0.3.0`; no migration surprise (the
      core is migration-free).
- [ ] CHANGELOG: the `[0.3.0]` dated section is moved out of `[Unreleased]`;
      compare-links repointed.
- [ ] Plugin version pin (v0.3.0+): `plugin/.claude-plugin/plugin.json`
      `version` matches the `v0.3.0` tag.
- [ ] Behavioral surfaces re-checked on the built artifact (per the §12(b)
      refinement): `claude plugin details` shows the MCP server registered;
      the Stop hook fires in a throwaway repo.

---

## Build Completion

*Filled in at the end of the **build** (rehearsal) cycle, before advancing to verify.*

- **Branch:**
- **PR (if applicable):**
- **All acceptance criteria met?** yes/no
- **New decisions emitted:**
  - `DEC-NNN` — <title> (if any)
- **Deviations from spec:**
  - [list]
- **Follow-up work identified:**
  - [any new specs for the stage's backlog]

### Build-phase reflection (3 questions, short answers)

1. **What was unclear in the spec that slowed you down?**
   — <answer>

2. **Was there a constraint or decision that should have been listed but wasn't?**
   — <answer>

3. **If you did this task again, what would you do differently?**
   — <answer>

---

## Reflection (Ship)

*Appended during the **ship** (cut) cycle.*

1. **What would I do differently next time?**
   — <answer>

2. **Does any template, constraint, or decision need updating?**
   — <answer>

3. **Is there a follow-up spec I should write now before I forget?**
   — <answer>
