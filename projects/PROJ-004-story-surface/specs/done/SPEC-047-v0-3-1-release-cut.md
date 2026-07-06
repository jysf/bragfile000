---
# Maps to ContextCore task.* semantic conventions.
# RELEASE-CUT variant of spec.md — the stage's closing release action.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-047
  type: story                      # a release cut is a story-sized closing action
  cycle: ship
  blocked: false                   # SPEC-046 (the only STAGE-014 feature) is on main (f8fc7dd, PR #79)
  priority: high
  complexity: S

project:
  id: PROJ-004
  stage: STAGE-014
repo:
  id: bragfile

agents:
  architect: claude-opus-4-8
  implementer: claude-opus-4-8     # same Claude, different session (claude-only variant)
  created_at: 2026-07-06

references:
  decisions: [DEC-027, DEC-024, DEC-021]
  constraints: [one-spec-per-pr, timestamps-in-utc-rfc3339]
  related_specs: [SPEC-046, SPEC-042]
---

# SPEC-047: v0.3.1 release cut

## Context

STAGE-014's only feature — **SPEC-046** (PR #79, merged to `main` at `f8fc7dd`
on 2026-07-06) — seeds cost/session/token capture on the MCP `brag_add`
provenance path. This spec is the stage's **closing release action**: it cuts
and ships **v0.3.1**, a patch over v0.3.0, following
[AGENTS.md §4](../../../AGENTS.md) release mechanics.

It was peeled from SPEC-046 at design (the release tag is cut from `main`
*after* the code PR lands, so it cannot share a PR with the code it tags —
`one-spec-per-pr`), mirroring **SPEC-042**'s v0.3.0 release-cut runbook, which
in turn mirrored SPEC-037's v0.2.0 precedent.

The design carries the **release runtime/operational pre-flight checklist**
(cross-project-retro R2; the release-cut template per AGENTS.md §4) so every §4
gotcha earned in prod (goreleaser dual-tag, macOS Gatekeeper, Homebrew 6.0+
brew-trust, prod-DB migration) is a **ticked design-time item**, not re-learned
in prod. SPEC-042's first real use validated the template as fit-for-purpose.

Existing tags are `v0.1.0`, `v0.2.0`, `v0.3.0`; v0.3.1 is the next **patch**
(SPEC-046 is additive and migration-free — no minor bump warranted).

## Goal

Cut, tag, and publish v0.3.1 to the Homebrew tap per AGENTS.md §4, verify a
clean `brew upgrade` from v0.3.0, and close STAGE-014.

## Split of responsibilities (mechanical prep vs. irreversible cut)

This spec's **PR** does the mechanical, reversible, CI-gated prep only:
author the CHANGELOG `[0.3.1]` section, bump the plugin version pin, tick the
pre-flight, get CI green, merge to `main`. The **irreversible cut** (RC tag,
smoke, dual-tag delete, final `v0.3.1` tag, goreleaser publish, tap bump, brew
verify) is driven by the coordinator/human from `main` after this merges —
verbatim per the §4 Pattern 1 sequence below. No tag is created in this PR.

## Inputs

- **Files to read:** [AGENTS.md §4](../../../AGENTS.md) (release mechanics +
  the three lessons-earned addenda — dual-tag Pattern 1, Gatekeeper, brew-trust);
  [`SPEC-042`](../../PROJ-003-agent-native-spine/specs/done/SPEC-042-v0-3-0-release-cut.md)
  (the v0.3.0 runbook this mirrors); `CHANGELOG.md` (the `[Unreleased]` section
  + the compare-link refs); `.goreleaser.yaml` + `.github/workflows/release.yml`
  (the cut machinery); `README.md` §Install (Gatekeeper xattr + brew-trust
  notes); `plugin/.claude-plugin/plugin.json` (the version pin, `0.3.0` → bump
  to `0.3.1`); `decisions/DEC-027-*.md` + `SPEC-046` (the surface the CHANGELOG
  describes).
- **External:** GitHub Releases; the `jysf/homebrew-bragfile` tap.
- **Related code paths:** none — this spec ships **no Go code**.

## Outputs

- **Files created:** `projects/PROJ-004-story-surface/specs/SPEC-047-...md`
  (this spec; archived to `done/` at ship).
- **Files modified:**
  - `CHANGELOG.md` — a dated `[0.3.1]` section (literal below); compare-links
    repointed.
  - `plugin/.claude-plugin/plugin.json` — `version` `0.3.0` → `0.3.1`.
  - `projects/PROJ-004-story-surface/stages/STAGE-014-...md` — backlog marks
    SPEC-047 `[x]` shipped; count bumped. (Stage frontmatter `status` stays
    `proposed`/`active` — NOT flipped to `shipped` until the tag is live; the
    coordinator does that after v0.3.1 publishes.)
  - `jysf/homebrew-bragfile` (separate repo) — `bragfile` cask bumped to
    `0.3.1` with matching sha256s (goreleaser publishes this **at the cut**, not
    in this PR).
- **New exports:** none.
- **Database changes:** **none** — v0.3.1 is migration-free. SPEC-046 added no
  migration (the seed rides the DEC-015 taggings join unchanged), so the DEC-021
  auto-backup safety belt does **not** fire on a v0.3.0→v0.3.1 open.

## The CHANGELOG `[0.3.1]` literal (transcribe verbatim at build)

Literal-artifact-as-spec (§12). Build inserts the following dated section
immediately below `## [Unreleased]` (which stays, empty), and repoints the link
refs at the file's bottom. The ten-verb Group-O4 assertion already passes from
the `[0.1.0]`..`[0.3.0]` sections and is unaffected.

```markdown
## [Unreleased]

## [0.3.1] - 2026-07-06

A small, additive **patch** that begins seeding per-work economics history. The
MCP `brag_add` tool now accepts optional `session` / `cost` / `tokens` inputs
and stamps them as reserved `session:` / `cost:` / `tokens:` tags, and the
plugin's capture-nudge hook forwards the Claude Code `session_id` so an
agent-captured entry carries a stable session join-key. No schema change, no CLI
change — cost/session history simply starts accruing now, ahead of the reporting
layer that will read it.

### Added

- **Optional cost / session / token capture on `brag_add` (MCP).** The MCP
  `brag_add` tool accepts three new **optional** inputs — `session`, `cost`,
  `tokens` — and stamps each as a reserved-namespace tag (`session:<id>`,
  `cost:<n>`, `tokens:<n>`) alongside the existing `agent:` / `model:`
  provenance. All three are optional: an omitted input stamps no tag, and
  bragfile never fabricates a value. `cost` must be a non-negative USD decimal
  and `tokens` a non-negative integer — a non-numeric or negative value is
  rejected as a tool error rather than silently stored. Reserved but **not**
  author-provenance: a `session:` / `cost:` / `tokens:`-only entry still
  classifies as `human` under `brag list --author` (DEC-027).
- **Session join-key forwarding in the capture-nudge hook.** The Claude Code
  plugin's session-end capture-nudge hook now surfaces the Claude Code
  `session_id` in its agent-facing context and instructs Claude to forward it as
  the `session` input on `brag_add`, so agent-captured entries carry a stable
  per-session join-key. The hook still never runs `brag` itself; its
  silent-degradation and once-per-session contracts are unchanged.

### Upgrading from v0.3.0

No manual steps and **no migration** — v0.3.1 adds no schema changes (the new
tags ride the existing taggings join) and no CLI changes (the capture is
MCP-path-only). `brew upgrade jysf/bragfile/bragfile` moves a v0.3.0 install to
v0.3.1 in place; `brag --version` then reports `0.3.1`. On a first tap install,
the two one-time frictions still apply: on **Homebrew 6.0+**, run `brew trust
--cask jysf/bragfile/bragfile` once; on **macOS**, clear an unsigned binary's
Gatekeeper quarantine with `xattr -dr com.apple.quarantine` (see the README
install note). To pick up the new capture behavior, reinstall the Claude Code
plugin so it runs the v0.3.1 binary.
```

Link-reference block at the bottom of `CHANGELOG.md` becomes:

```markdown
[Unreleased]: https://github.com/jysf/bragfile000/compare/v0.3.1...HEAD
[0.3.1]: https://github.com/jysf/bragfile000/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/jysf/bragfile000/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/jysf/bragfile000/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/jysf/bragfile000/releases/tag/v0.1.0
```

## Acceptance Criteria

Each gate is independently checkable. Prep gates (this PR) are green before the
PR opens; cut gates describe the observable end state the coordinator produces
at the tag (mirrors SPEC-042).

- [ ] **Prep gates green** (this PR, before it opens): `go test ./...` green;
      `gofmt -l .` empty; `go vet ./...` clean; `CGO_ENABLED=0 go build ./...`
      succeeds; `just test-docs` exits 0; `just test-hook` exits 0.
- [ ] **CHANGELOG reconciled:** a dated `## [0.3.1] - 2026-07-06` section is
      present (the literal above), `[Unreleased]` stays present-and-empty, and
      the compare-links are repointed (`test-docs.sh` Group O stays green).
- [ ] **Plugin version pin matches the tag:** `plugin/.claude-plugin/plugin.json`
      `version` == `0.3.1` == the intended git tag.
- [ ] **Pre-flight all ticked** (design; the checklist under Notes) with either
      concrete evidence or an explicit "verified by orchestrator at the cut".
- [ ] **rc is a GitHub PRERELEASE with no tap change** *(cut)*: after
      `v0.3.1-rc1` CI completes, its GitHub release is *Pre-release* and the
      `jysf/homebrew-bragfile` tap's latest commit is **unchanged**.
- [ ] **RC smoke gate passes on a THROWAWAY DB** *(cut, never `~/.bragfile`)*:
      the rc binary reports `0.3.1-rc1`; `brag mcp serve` answers a `tools/list`
      over stdio; a `brag_add` with `session`/`cost`/`tokens` stamps the three
      reserved tags (bad numerics rejected as a tool error); a
      `session:`/`cost:`/`tokens:`-only entry classifies as `--author human`;
      and — because v0.3.1 is migration-free — opening a seeded v0.3.0-schema
      throwaway DB fires **no** migration and writes **no** backup sidecar
      (DEC-021 trigger `applied>0 && pending>0` is false).
- [ ] **Behavioral surfaces re-checked on the built artifact** *(cut, §12(b)
      refinement)*: against a clean install of the built plugin, `claude plugin
      details brag` shows the MCP server **registered** (not just `validate
      --strict` green), and the Stop hook fires once in a throwaway repo.
- [ ] **Final release has all 4 platform archives + checksums** *(cut)*: the
      `v0.3.1` GitHub release (NOT a prerelease) carries the four `darwin/linux
      × amd64/arm64` tarballs and `checksums.txt`.
- [ ] **Tap bumped to 0.3.1** *(cut)*: `jysf/homebrew-bragfile` has a new commit
      setting the `bragfile` cask to `0.3.1` with matching sha256s.
- [ ] **Brew binary reports 0.3.1** *(cut)*: after `brew update && brew upgrade
      jysf/bragfile/bragfile` (+ `brew trust --cask` first-time, + the Gatekeeper
      xattr clear), `brag --version` from the brew path reports `0.3.1`.
- [ ] **Prod DB opens with no migration** *(cut)*: opening the real
      `~/.bragfile` (already v0.3.0) with the released binary fires no migration
      and writes no new backup sidecar.
- [ ] **No regressions:** the full Go suite, `gofmt`, `go vet`, and the
      `CGO_ENABLED=0` build stay green (verified at prep; this spec adds no code
      that could break them).

## Failing Tests

This spec ships **no Go code**, so there are no new `*_test.go` files — the
`test-before-implementation` constraint is satisfied the way SPEC-042/037/034/035
(also code-free) satisfied it: the enforcement mechanism is the **prep gate
suite** (the existing green `go test ./...`, `scripts/test-docs.sh` including
Group O's CHANGELOG-shape assertions, and `just test-hook`) plus the
**post-release verification checklist** (the ACs above), not new Go assertions.
"The test" for a release runbook is the executable gate, not a Go assertion. The
regression guard is the pre-existing suite, run unchanged at prep.

---

## The build cycle is CHANGELOG + version-bump prep (read before building)

The build cycle for SPEC-047 does **not** cut the release. It authors the
reversible artifacts so the real cut (coordinator + human go-aheads) is
mechanical:

1. **Author the CHANGELOG `[0.3.1]`** from the literal above (date `2026-07-06`),
   repoint the link refs, and run `just test-docs` — Group O must stay green.
2. **Bump `plugin/.claude-plugin/plugin.json`** `version` to `0.3.1`.
3. **Run the six prep gates** (below) and confirm all green.
4. **Update the STAGE-014 backlog** — mark SPEC-047 `[x]` shipped, bump the
   count. Leave the stage frontmatter `status` untouched (the coordinator flips
   it after the tag publishes).

No tag is pushed; `goreleaser release` is never invoked; no `gh release` call
and no `brew` command is made in this cycle. Those are the coordinator's, below.

## Prep gates (this PR)

```bash
go test ./...
gofmt -l .                    # must be empty
go vet ./...
CGO_ENABLED=0 go build ./...
just test-docs
just test-hook
```

## RC smoke gate (throwaway DB — never ~/.bragfile)

```bash
export BRAGFILE_DB=$(mktemp -d)/smoke.db        # THROWAWAY — never ~/.bragfile
brag --version                                  # → 0.3.1-rc1
# MCP: start the server and confirm it answers tools/list over stdio.
brag mcp serve </dev/null &                     # or drive it from an in-editor MCP client
# brag_add with the new inputs → session:/cost:/tokens: reserved tags stamped;
#   bad numerics (cost:abc / negative) rejected as a tool error, not stored.
# A session:/cost:/tokens:-only entry classifies as --author human:
brag list --author agent                        # → does NOT include the seed-only entry
brag list --author human                        # → includes it
# Plugin registration (the §12(b) refinement check):
claude plugin marketplace add jysf/bragfile000
claude plugin install brag@bragfile
claude plugin details brag                      # → MCP servers (1) brag  (registered, not just valid)
# Migration-free upgrade: seed a v0.3.0-schema throwaway DB, open it, assert
# NO *.backup sidecar appears (applied>0 && pending>0 is false at v0.3.1).
```

## Destructive sequence (coordinator/human-gated — §4 Pattern 1)

The real cut, verbatim, executed only after this PR merges and the human
go-aheads are given:

```bash
# 0. main is clean, at the release SHA (this PR's merge commit), local == origin;
#    CHANGELOG [0.3.1] + plugin version 0.3.1 merged.
# 1. Optional RC to exercise CI + the tap path:
git tag v0.3.1-rc1 && git push origin v0.3.1-rc1     # CI builds a PRERELEASE; skip_upload:auto holds the tap
# 2. Run the RC SMOKE GATE above against the downloaded rc binary. If good:
# 3. Dual-tag-on-same-commit — delete the RC tag + release BEFORE the final tag (§4 Pattern 1):
gh release delete v0.3.1-rc1 --yes --cleanup-tag
git tag v0.3.1 && git push origin v0.3.1             # CI cuts the real release + bumps the tap
# 4. Post-release verification: the AC checklist (4 archives + checksums, tap 0.3.1,
#    brew upgrade → 0.3.1, prod DB no-migration).
# 5. Flip STAGE-014 frontmatter status → shipped + shipped_at once v0.3.1 is live.
```

## Implementation Context

*Read this section (and the files it points to) before starting the build cycle.*

### Decisions that apply

- `DEC-027` — the reserved `session:`/`cost:`/`tokens:` namespace this release
  ships and the CHANGELOG describes; author classification stays
  `agent:%`/`model:%`-only (a seed-only entry is `human`).
- `DEC-024` — the reserved-namespace + `stampProvenance` path DEC-027 extends;
  the `agent:`/`model:` provenance already shipped in v0.3.0.
- `DEC-021` — the migration auto-backup safety belt; relevant here as the
  *negative* check (v0.3.1 is migration-free, so it must NOT fire).

### Constraints that apply

- `one-spec-per-pr` — the release tag is cut from `main` after the SPEC-046 PR
  landed; this spec is its own PR, separate from SPEC-046 (PR #79).
- `timestamps-in-utc-rfc3339` — untouched; SPEC-046 stamps opaque/numeric tags,
  no timestamp semantics changed.

### Prior related work

- `SPEC-042` (shipped) — the v0.3.0 release-cut runbook this mirrors (pre-flight
  → RC prerelease → smoke → dual-tag → tap bump → brew verify); its ship
  reflection confirmed the release-cut template fit-for-purpose.
- `SPEC-046` (shipped, PR #79, `f8fc7dd`) — the seed capture this release
  publishes.

### Out of scope (for this spec specifically)

- macOS notarization (still deferred; the Gatekeeper xattr note stands).
- Any code change to the v0.3.1 surface — if the cut surfaces a bug, that is a
  new fix spec, not an edit here.
- First-class cost/tokens/session columns + exact-token reconciliation
  (PROJ-005; DEC-027 accepts the stringly-typed tag as debt).
- v0.4.0 `brag impact` digest work (STAGE-011).

## Notes for the Implementer

Gotchas, style preferences, reuse opportunities.

- The `[Unreleased]` section is currently **empty** — SPEC-046 shipped without a
  CHANGELOG entry, so the `[0.3.1]` section is authored whole from the literal
  above (not moved from accumulated Unreleased notes). `[Unreleased]` stays as a
  present-but-empty heading (Group O5 asserts the `[Unreleased]:` ref present).
- `test-docs.sh` Group O asserts: `[Unreleased]:` and `[0.1.0]:` link refs
  present (O5), the ten verbs in backticks (O4 — already satisfied by prior
  sections), the `## [0.1.0]` heading (O3). Keep all of these; add the `[0.3.1]:`
  ref and repoint `[Unreleased]:` to `v0.3.1...HEAD`.
- The plugin version pin is `0.3.0` today — bump it to `0.3.1` to match the tag
  (the pre-flight requires the pin == the tag).

### Release runtime/operational pre-flight (all must be ticked at design)

Adopted from the release-cut spec template (`projects/_templates/spec-release-cut.md`,
per AGENTS.md §4). Concretized for the v0.3.1 `brag` cut. Items verifiable now
carry concrete evidence; items that can only be checked AT the tag/RC are marked
"verified by orchestrator at the cut" (honest — not faked).

- [x] Dual-tag-on-same-commit: RC tag + release deleted before the final tag is
      cut at the same commit (§4 Pattern 1). — **Documented** in the Destructive
      sequence above (verbatim §4 Pattern 1 commands: `gh release delete
      v0.3.1-rc1 --yes --cleanup-tag` then `git tag v0.3.1 && git push`).
      *Executed by orchestrator at the cut.*
- [x] macOS Gatekeeper: `xattr -dr com.apple.quarantine <bin>` note present in
      README §Install. — **Evidence:** `README.md:45` ("macOS Gatekeeper note")
      + the `sudo xattr -dr com.apple.quarantine /opt/homebrew/Caskroom/bragfile/`
      command at `README.md:53`.
- [x] Homebrew 6.0+: `brew trust --cask <tap>/<cask>` documented in README and
      run once at the cut. — **Evidence:** `README.md:30` (`brew trust --cask
      jysf/bragfile/bragfile`, under the "Homebrew 6.0+ note"). *The one-time
      `brew trust` + `brew upgrade` run is verified by orchestrator at the cut.*
- [x] Dev/prod DB isolation: the RC smoke test runs against a THROWAWAY DB,
      never ~/.bragfile; the auto-backup path is observed. — **Evidence:** the RC
      smoke gate above sets `BRAGFILE_DB=$(mktemp -d)/smoke.db`. For v0.3.1 the
      DEC-021 backup path is asserted **not** to fire (migration-free); *the
      throwaway-DB smoke is run by orchestrator at the cut.*
- [x] Clean upgrade: `brew upgrade` from the prior minor verified; `brag
      --version` prints the new tag; no migration surprise. — **Verified by
      orchestrator at the cut** (`brew upgrade jysf/bragfile/bragfile` from
      v0.3.0 → `brag --version` == `0.3.1`; migration-free per DEC-021 negative
      check). Design evidence that it *will* be clean: SPEC-046 added no
      migration file under `internal/storage/migrations/`.
- [x] CHANGELOG: the `[0.3.1]` dated section is authored (out of `[Unreleased]`);
      compare-links repointed. — **Done in this PR** (the literal above; Group O
      stays green — verified at prep).
- [x] Plugin version pin (v0.3.0+): `plugin/.claude-plugin/plugin.json`
      `version` matches the tag. — **Done in this PR** (bumped `0.3.0` →
      `0.3.1`).
- [x] Behavioral surfaces re-checked on the built artifact (§12(b) refinement):
      `claude plugin details` shows the MCP server registered; the Stop hook
      fires in a throwaway repo. — **Verified by orchestrator at the cut** (needs
      the built plugin + a clean Claude install; not run here to avoid mutating
      the user's Claude config, per SPEC-042's precedent).

---

## Build Completion

*Filled in at the end of the **build** (prep) cycle, before advancing to verify.*

- **Branch:** `feat/spec-047-v0-3-1-release-cut`
- **PR (if applicable):** this PR — authors the CHANGELOG `[0.3.1]` + bumps the
  plugin version pin to `0.3.1` + updates the STAGE-014 backlog. The **tag cut
  itself is NOT in this PR** (it is the coordinator/human-gated ship step, run
  from `main` after this merges).
- **All acceptance criteria met?** **Prep ACs green; cut ACs deferred to the
  orchestrator's tag.** Verified at prep (2026-07-06):
  - `go test ./...` green; `gofmt -l .` empty; `go vet ./...` clean;
    `CGO_ENABLED=0 go build ./...` succeeds; `just test-docs` exits 0;
    `just test-hook` exits 0.
  - CHANGELOG `[0.3.1]` authored (date `2026-07-06`) + link-refs repointed;
    `test-docs.sh` Group O green.
  - `plugin/.claude-plugin/plugin.json` `version` == `0.3.1`.
  - **Deferred to the cut** (need the real tag / a plugin install / a throwaway
    DB): the RC-is-prerelease + no-tap-change gate, the RC binary smoke
    (`session`/`cost`/`tokens` stamping + author-classification + migration-free
    open), the final release's 4 archives + checksums, the tap bump to `0.3.1`,
    `brew upgrade → 0.3.1`, and the §12(b) `claude plugin details brag`
    registration + Stop-hook check on the built plugin.
- **New decisions emitted:** none.
- **Deviations from spec:** none. CHANGELOG dated `2026-07-06` (the intended cut
  date); if the cut slips, the orchestrator bumps the date to match.
- **Follow-up work identified:**
  - The **ship (cut)** step — coordinator/human-gated `git tag v0.3.1-rc1` →
    smoke → delete RC → `git tag v0.3.1` (§4 Pattern 1), then the tap bump +
    brew verify, then flip STAGE-014 frontmatter `status` → shipped. Closes
    STAGE-014.

### Build-phase reflection (3 questions, short answers)

1. **What was unclear in the spec that slowed you down?**
   — Nothing. The embedded CHANGELOG literal transcribed verbatim and the prep
   gates were a checklist. The one judgement call — leaving the stage frontmatter
   `status` unflipped until the tag is live — was already spelled out in Outputs.

2. **Was there a constraint or decision that should have been listed but wasn't?**
   — No. `one-spec-per-pr` (separate PR from SPEC-046) and the DEC-021 negative
   check (migration-free open) were both listed and both held.

3. **If you did this task again, what would you do differently?**
   — Nothing structural. The SPEC-042 precedent made the mechanical/irreversible
   split obvious; the only real thinking was writing an accurate, additive
   CHANGELOG for a patch (no `### Fixed`, no `### Changed`), which the SPEC-046 /
   DEC-027 surface dictated cleanly.

---

## Reflection (Ship)

*Appended during the **ship** cycle. Outcome-focused, distinct from the
process-focused build reflection above. This spec's ship = the mechanical prep
merged to main; the irreversible tag/publish is the orchestrator's.*

1. **What would I do differently next time?**
   — Little. The v0.3.1 cut is the third consecutive use of the R2 release-cut
   pre-flight, and it kept the prep mechanical: every §4 gotcha (dual-tag,
   Gatekeeper, brew-trust, prod-DB) was a ticked item with either README evidence
   or an explicit "verified by orchestrator at the cut", so the reversible PR and
   the irreversible cut stayed cleanly separated. One honest note: because this
   session hard-stops before the tag, several pre-flight items are deferred-to-cut
   rather than ticked-with-evidence — the checklist earns its keep only if the
   orchestrator actually runs them, so the deferred set is enumerated explicitly
   rather than left implied.

2. **Does any template, constraint, or decision need updating?**
   — No. The `spec-release-cut.md` template held for a patch as well as it did for
   the v0.3.0 minor (SPEC-042). Worth noting for a future template refinement: the
   pre-flight checklist has no built-in slot to distinguish "ticked with evidence
   now" from "deferred to the cut", and this spec had to add that distinction in
   prose — a template that split the two columns explicitly would make the honest
   split first-class. Not urgent (N=1); watch at the next release cut.

3. **Is there a follow-up spec I should write now before I forget?**
   — No new spec. The open thread is PROJ-005 (economics: promote the
   stringly-typed `cost:`/`tokens:`/`session:` tags to typed columns + the
   exact-token reconciliation join), which DEC-027 already records as deferred
   debt — not a STAGE-014 follow-up. STAGE-014 closes with this cut.
