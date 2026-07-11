---
# Maps to ContextCore task.* semantic conventions.
# RELEASE-CUT variant of spec.md — the stage's closing release action.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-069
  type: story                      # a release cut is a story-sized closing action
  cycle: verify
  blocked: false                   # SPEC-068 (the only v0.5.1 feature) is on main
  priority: high
  complexity: S

project:
  id: PROJ-006
  stage: STAGE-017
repo:
  id: bragfile

agents:
  architect: claude-opus-4-8
  implementer: claude-opus-4-8     # same Claude, different session (claude-only variant)
  created_at: 2026-07-11

references:
  decisions: [DEC-039]
  constraints: [one-spec-per-pr]
  related_specs: [SPEC-068]
---

# SPEC-069: v0.5.1 release cut

## Context

STAGE-017's single feature spec — **SPEC-068** (`brag list --day
<YYYY-MM-DD|today|yesterday>`, DEC-039), plus the small internal clock-seam fix
that rode with it — is built, dogfooded, and merged to `main` (#115), and
unreleased. This spec is the stage's **closing release action**: it cuts and
ships **v0.5.1**, a **patch** release over v0.5.0, following
[AGENTS.md §4](../../../AGENTS.md) release mechanics.

The v0.5.1 surface (all on `main`):

- **SPEC-068** (`brag list --day`) — one flag that scopes a listing to a single
  **local** calendar day (the half-open `[local-midnight, next-local-midnight)`
  window), built on the `ListFilter.Until` bound shipped in v0.5.0/SPEC-056. The
  `today`/`yesterday` keywords resolve against the local wall clock; the flag is
  mutually exclusive with `--since` and composes with every other filter
  (DEC-039).
- The internal **clock-seam fix** that rode with SPEC-068 — `--since`'s duration
  path and the new `--day` keywords now resolve the current time through an
  injectable `clock` seam in `internal/cli/since.go` instead of reading the wall
  clock inline. A testability fix (audit L4); **no behavior change** to `--since`.

Like the last five cuts this spec is peeled from the feature work at design (the
release tag is cut from `main` *after* the code PR lands, so it cannot share a PR
with the code it tags — `one-spec-per-pr`). It mirrors **SPEC-067**'s v0.5.0
runbook (in turn SPEC-054's v0.4.0 / SPEC-047's v0.3.1 precedents). It carries
the **release runtime/operational pre-flight checklist** (cross-project-retro R2;
the release-cut template per AGENTS.md §4) so every §4 gotcha earned in prod
(goreleaser dual-tag, macOS Gatekeeper, Homebrew 6.0+ brew-trust, prod-DB
migration) is a **ticked design-time item**, not re-learned in prod.

Existing tags are `v0.1.0`, `v0.2.0`, `v0.3.0`, `v0.3.1`, `v0.4.0`, `v0.5.0`;
v0.5.1 is the next **patch** — a patch (not a minor) because it adds no new
top-level command: `--day` is a new *flag* on the existing `brag list`, and the
clock seam is internal. It is additive and **migration-free** (no new file under
`internal/storage/migrations/`, confirmed at build — four files `0001`..`0004`,
unchanged).

## Goal

Cut, tag, and publish v0.5.1 to the Homebrew tap per AGENTS.md §4, verify a clean
`brew upgrade` from v0.5.0, run the §12(b) behavioral check on the built plugin,
and close STAGE-017.

## Split of responsibilities (mechanical prep vs. irreversible cut)

This spec's **PR** does the mechanical, reversible, CI-gated prep only: author
the CHANGELOG `[0.5.1]` section, bump the plugin version pin, tick the pre-flight,
get CI green, merge to `main`. The **irreversible cut** (RC tag, smoke, dual-tag
delete, final `v0.5.1` tag, goreleaser publish, tap bump, brew verify) is driven
by the orchestrator/human from `main` after this merges — verbatim per the §4
Pattern 1 sequence below. **No tag is created in this PR.**

## Inputs

- **Files to read:** [AGENTS.md §4](../../../AGENTS.md) (release mechanics + the
  three lessons-earned addenda — dual-tag Pattern 1, Gatekeeper, brew-trust);
  [`SPEC-067`](../../PROJ-005-agent-native-depth/specs/done/SPEC-067-v0-5-0-release-cut.md)
  (the v0.5.0 runbook this mirrors); `CHANGELOG.md` (the `[Unreleased]` section +
  the compare-link refs); `.goreleaser.yaml` + `.github/workflows/release.yml`
  (the cut machinery); `README.md` §Install (Gatekeeper xattr + brew-trust notes);
  `plugin/.claude-plugin/plugin.json` (the version pin, `0.5.0` → bump to
  `0.5.1`); the DEC/spec of the surface the CHANGELOG describes (DEC-039 +
  SPEC-068).
- **External:** GitHub Releases; the `jysf/homebrew-bragfile` tap.
- **Related code paths:** none — this spec ships **no Go code**.

## Outputs

- **Files created:**
  `projects/PROJ-006-agent-native-depth-core/specs/SPEC-069-v0-5-1-release-cut.md`
  (this spec; archived to `done/` at ship).
- **Files modified:**
  - `CHANGELOG.md` — a dated `[0.5.1]` section (`### Added` / `### Fixed`);
    compare-links repointed.
  - `plugin/.claude-plugin/plugin.json` — `version` `0.5.0` → `0.5.1`.
  - `projects/PROJ-006-agent-native-depth-core/stages/STAGE-017-...md` — SPEC-068
    marked shipped, SPEC-069 (the release cut) added as the stage's closing
    action, the Count updated. (STAGE-017 stays `status: active`; the orchestrator
    closes it after v0.5.1 publishes.)
  - `jysf/homebrew-bragfile` (separate repo) — `bragfile` cask bumped to `0.5.1`
    with matching sha256s (goreleaser publishes this **at the cut**, not in this
    PR).
- **New exports:** none.
- **Database changes:** **none** — v0.5.1 is migration-free. SPEC-068 is pure
  query-time filtering over the existing `Since`/`Until` fields, so the DEC-021
  auto-backup safety belt does **not** fire on a v0.5.0 open.

## The CHANGELOG `[0.5.1]` literal (transcribed at build)

Literal-artifact-as-spec (§12). Build inserts a dated section immediately below
`## [Unreleased]` (which stays, empty), grouping the surface under `### Added` /
`### Fixed`, and repoints the link refs at the file's bottom. The ten-verb
Group-O4 assertion already passes from the `[0.1.0]` section and is unaffected.
See the merged `CHANGELOG.md` `[0.5.1]` section for the authored text; the
link-reference block becomes:

```markdown
[Unreleased]: https://github.com/jysf/bragfile000/compare/v0.5.1...HEAD
[0.5.1]: https://github.com/jysf/bragfile000/compare/v0.5.0...v0.5.1
[0.5.0]: https://github.com/jysf/bragfile000/compare/v0.4.0...v0.5.0
...
```

## Acceptance Criteria

Each gate is independently checkable. Prep gates (this PR) are green before the
PR opens; cut gates describe the observable end state the orchestrator produces
at the tag (mirrors SPEC-067).

- [x] **Prep gates green** (this PR, before it opens): `go test ./...` green;
      `gofmt -l .` empty; `go vet ./...` clean; `CGO_ENABLED=0 go build ./...`
      succeeds; `just test-docs` exits 0; `just test-hook` exits 0.
- [x] **CHANGELOG reconciled:** a dated `## [0.5.1] - 2026-07-11` section is
      present, `[Unreleased]` stays present-and-empty, and the compare-links are
      repointed (`test-docs.sh` Group O stays green).
- [x] **Plugin version pin matches the tag:** `plugin/.claude-plugin/plugin.json`
      `version` == `0.5.1` == the intended git tag.
- [x] **Pre-flight all ticked** (design; the checklist under Notes) with either
      concrete evidence or an explicit "verified by orchestrator at the cut".
- [x] **Migration-free confirmed at build:** no new file under
      `internal/storage/migrations/` since v0.5.0 (four files: `0001`..`0004`,
      unchanged). v0.5.1 is a read-side flag + an internal clock-seam fix.
- [ ] **rc is a GitHub PRERELEASE with no tap change** *(cut)*: after
      `v0.5.1-rc1` CI completes, its GitHub release is *Pre-release* and the
      `jysf/homebrew-bragfile` tap's latest commit is **unchanged**.
- [ ] **RC smoke gate passes on a THROWAWAY DB** *(cut, never `~/.bragfile`)*:
      the rc binary reports `0.5.1-rc1`; `brag list --day today`, `--day
      yesterday`, and `--day <YYYY-MM-DD>` each scope to the correct local day
      (the 23:30-local entry IN, the 00:30-next-day entry OUT); `--day today
      --since 7d` errors as a `UserError` naming both flags; and — because v0.5.1
      is migration-free — opening a seeded v0.5.0-schema throwaway DB fires **no**
      migration and writes **no** backup sidecar (DEC-021 trigger `applied>0 &&
      pending>0` is false).
- [ ] **Behavioral surfaces re-checked on the built artifact** *(cut, §12(b)
      refinement)*: against a clean install of the built plugin, `claude plugin
      details brag` shows the MCP server **registered** (not just `validate
      --strict` green), and the Stop hook fires once in a throwaway repo.
- [ ] **Final release has all 4 platform archives + checksums** *(cut)*: the
      `v0.5.1` GitHub release (NOT a prerelease) carries the four `darwin/linux ×
      amd64/arm64` tarballs and `checksums.txt`.
- [ ] **Tap bumped to 0.5.1** *(cut)*: `jysf/homebrew-bragfile` has a new commit
      setting the `bragfile` cask to `0.5.1` with matching sha256s.
- [ ] **Brew binary reports 0.5.1** *(cut)*: after `brew update && brew upgrade
      jysf/bragfile/bragfile` (+ the Gatekeeper xattr clear if prompted),
      `brag --version` from the brew path reports `0.5.1`.
- [ ] **Prod DB opens with no migration** *(cut)*: opening the real `~/.bragfile`
      (already v0.5.0) with the released binary fires no migration and writes no
      new backup sidecar.
- [x] **No regressions:** the full Go suite, `gofmt`, `go vet`, and the
      `CGO_ENABLED=0` build stay green (verified at prep; this spec adds no code
      that could break them).

## Failing Tests

This spec ships **no Go code**, so there are no new `*_test.go` files — the
`test-before-implementation` constraint is satisfied the way SPEC-067/054/047
(also code-free) satisfied it: the enforcement mechanism is the **prep gate
suite** (the existing green `go test ./...`, `scripts/test-docs.sh` including
Group O's CHANGELOG-shape assertions, and `just test-hook`) plus the
**post-release verification checklist** (the ACs above), not new Go assertions.
"The test" for a release runbook is the executable gate, not a Go assertion. The
regression guard is the pre-existing suite, run unchanged at prep.

---

## The build cycle is CHANGELOG + version-bump prep (read before building)

The build cycle for SPEC-069 does **not** cut the release. It authors the
reversible artifacts so the real cut (orchestrator + human go-aheads) is
mechanical:

1. **Author the CHANGELOG `[0.5.1]`** (date `2026-07-11`) — `brag list --day`
   under `### Added`, the internal clock-seam fix under `### Fixed` — repoint the
   link refs, and run `just test-docs` — Group O must stay green.
2. **Bump `plugin/.claude-plugin/plugin.json`** `version` to `0.5.1`.
3. **Confirm migration-free** — no new file under `internal/storage/migrations/`.
4. **Run the six prep gates** (below) and confirm all green.
5. **Update STAGE-017 backlog** — SPEC-068 `[x] (shipped on 2026-07-11)`, add
   SPEC-069 as the stage's closing action, update the Count. Leave the stage
   `status: active` (and the PROJ-006 frontmatter) to the orchestrator after the
   tag publishes.

No tag is pushed; `goreleaser release` is never invoked; no `gh release` call and
no `brew` command is made in this cycle. Those are the orchestrator's, below.

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
brag --version                                  # → 0.5.1-rc1
# Exercise the new flag:
brag add --title "late entry"                   # seed something today
brag list --day today                           # scopes to today's LOCAL day
brag list --day yesterday                        # scopes to yesterday's LOCAL day
brag list --day 2026-07-05                       # explicit local-day window
brag list --day today --since 7d                 # → UserError naming both flags, empty stdout
brag list --day notaday                          # → UserError naming today/yesterday/YYYY-MM-DD
# Migration-free upgrade: seed a v0.5.0-schema throwaway DB, open it, assert
# NO *.backup sidecar appears (applied>0 && pending>0 is false at v0.5.1).
```

## Destructive sequence (orchestrator/human-gated — §4 Pattern 1)

The real cut, verbatim, executed only after this PR merges and the human
go-aheads are given:

```bash
# 0. main is clean, at the release SHA (this PR's merge commit), local == origin;
#    CHANGELOG [0.5.1] + plugin version 0.5.1 merged.
# 1. Optional RC to exercise CI + the tap path:
git tag v0.5.1-rc1 && git push origin v0.5.1-rc1     # CI builds a PRERELEASE; skip_upload:auto holds the tap
# 2. Run the RC SMOKE GATE above against the downloaded rc binary. If good:
# 3. Dual-tag-on-same-commit — delete the RC tag + release BEFORE the final tag (§4 Pattern 1):
gh release delete v0.5.1-rc1 --yes --cleanup-tag
git tag v0.5.1 && git push origin v0.5.1             # CI cuts the real release + bumps the tap
# 4. Post-release verification: the AC checklist (4 archives + checksums, tap 0.5.1,
#    brew upgrade → 0.5.1, prod DB no-migration).
# 5. Flip STAGE-017 frontmatter status → shipped + shipped_at once v0.5.1 is live.
```

## Implementation Context

*Read this section (and the files it points to) before starting the build cycle.*

### Decisions that apply

- `DEC-039` — the `brag list --day` local-calendar-day window semantics (LOCAL
  day via Since+Until; today/yesterday keywords resolved off an injectable clock;
  mutually exclusive with `--since`); the CHANGELOG describes its surface. The DEC
  carries a revisit note: "local" here is the single host's wall-clock zone — a
  future non-CLI/multi-host caller may need to source the zone explicitly.
- `DEC-021` — the migration auto-backup safety belt; relevant here as the
  *negative* check (v0.5.1 is migration-free, so it must NOT fire).

### Constraints that apply

- `one-spec-per-pr` — the release tag is cut from `main` after the feature PR
  landed; this spec is its own PR, separate from the SPEC-068 feature PR (#115).

### Prior related work

- `SPEC-067` (shipped) — the v0.5.0 release-cut runbook this mirrors (pre-flight →
  RC prerelease → smoke → dual-tag → tap bump → brew verify); it closed STAGE-016
  the same way this closes STAGE-017.
- `SPEC-068` (shipped, on `main`) — the `brag list --day` surface this release
  publishes.

### Out of scope (for this spec specifically)

- macOS notarization (still deferred; the Gatekeeper xattr note stands).
- Any code change to the v0.5.1 surface — if the cut surfaces a bug, that is a
  new fix spec, not an edit here.
- The deeper PROJ-006 agent-native stages (memory, signed provenance,
  capture-completeness, model-benchmark) — later stages, not this cut.

## Notes for the Implementer

Gotchas, style preferences, reuse opportunities.

- The `[Unreleased]` section is **empty** at prep — SPEC-068 shipped without a
  CHANGELOG entry, so the `[0.5.1]` section is authored whole. `[Unreleased]`
  stays as a present-but-empty heading (Group O5 asserts the `[Unreleased]:` ref
  present).
- `test-docs.sh` Group O asserts: `[Unreleased]:` and `[0.1.0]:` link refs present
  (O5), the ten verbs in backticks (O4 — already satisfied by the `[0.1.0]`
  section), the `## [0.1.0]` heading (O3), `keepachangelog.com` (O2). Keep all of
  these; add the `[0.5.1]:` ref and repoint `[Unreleased]:` to `v0.5.1...HEAD`.
- The plugin version pin is `0.5.0` today — bump it to `0.5.1` to match the tag
  (the pre-flight requires the pin == the tag). `plugin/.mcp.json` carries **no**
  version field, so it needs no bump. Other `0.5.0` hits in the tree are
  historical CHANGELOG/docs references — leave those.
- **Patch, not minor:** v0.5.1 adds no new top-level command — `--day` is a flag
  on the existing `brag list`, and the clock seam is internal. So it is a patch
  over v0.5.0. Unlike v0.5.0 (Added/Changed/Fixed), this cut carries only
  `### Added` (the `--day` flag) and `### Fixed` (the internal clock seam).

### Release runtime/operational pre-flight (all must be ticked at design)

Adopted from the release-cut spec template (`projects/_templates/spec-release-cut.md`,
per AGENTS.md §4). Concretized for the v0.5.1 `brag` cut. Items verifiable now
carry concrete evidence; items that can only be checked AT the tag/RC are marked
"verified by orchestrator at the cut" (honest — not faked).

- [x] Dual-tag-on-same-commit: RC tag + release deleted before the final tag is
      cut at the same commit (§4 Pattern 1). — **Documented** in the Destructive
      sequence above (verbatim §4 Pattern 1 commands: `gh release delete
      v0.5.1-rc1 --yes --cleanup-tag` then `git tag v0.5.1 && git push`).
      *Executed by orchestrator at the cut.*
- [x] macOS Gatekeeper: `xattr -dr com.apple.quarantine <bin>` note present in
      README §Install. — **Evidence:** the README "macOS Gatekeeper note" +
      `xattr -dr com.apple.quarantine` command in the install section (unchanged
      since v0.5.0; verified present at prep). *The one-time clear is run by
      orchestrator at the cut.*
- [x] Homebrew 6.0+: `brew trust --cask <tap>/<cask>` documented in README and
      run once at the cut. — **Evidence:** `brew trust --cask
      jysf/bragfile/bragfile` documented in the README install section (under the
      "Homebrew 6.0+ note"). *Already-trusted since v0.2.0, so no re-prompt is
      expected on this upgrade; verified by orchestrator at the cut.*
- [x] Dev/prod DB isolation: the RC smoke test runs against a THROWAWAY DB, never
      ~/.bragfile; the auto-backup path is observed. — **Evidence:** the RC smoke
      gate above sets `BRAGFILE_DB=$(mktemp -d)/smoke.db`. For v0.5.1 the DEC-021
      backup path is asserted **not** to fire (migration-free); *the throwaway-DB
      smoke is run by orchestrator at the cut.*
- [x] Clean upgrade: `brew upgrade` from the prior release verified; `brag
      --version` prints the new tag; no migration surprise. — **Verified by
      orchestrator at the cut** (`brew upgrade jysf/bragfile/bragfile` from v0.5.0
      → `brag --version` == `0.5.1`; migration-free per DEC-021 negative check).
      Design evidence that it *will* be clean: **confirmed at build** that no
      migration file was added under `internal/storage/migrations/` (four files
      `0001`..`0004`, unchanged since v0.2.0).
- [x] Prod-DB migration = **N/A this release**: v0.5.1 adds no migration, so the
      DEC-021 backup safety belt does not fire on a v0.5.0→v0.5.1 open. —
      **Confirmed at build** (no new file under `internal/storage/migrations/`).
- [x] CHANGELOG: the `[0.5.1]` dated section is authored; compare-links repointed.
      — **Done in this PR** (the surface under `### Added`/`### Fixed`; Group O
      stays green — verified at prep).
- [x] Plugin version pin (v0.3.0+): `plugin/.claude-plugin/plugin.json` `version`
      matches the tag. — **Done in this PR** (bumped `0.5.0` → `0.5.1`).
- [x] Behavioral surfaces re-checked on the built artifact (§12(b) refinement):
      `claude plugin details` shows the MCP server registered; the Stop hook fires
      in a throwaway repo. — **Verified by orchestrator at the cut** (needs the
      built plugin + a clean Claude install; not run here to avoid mutating the
      user's Claude config, per SPEC-067/054's precedent).

---

## Build Completion

*Filled in at the end of the **build** (prep) cycle, before advancing to verify.*

- **Branch:** `feat/spec-069-v0-5-1-release-cut`
- **PR (if applicable):** this PR — authors the CHANGELOG `[0.5.1]` + bumps the
  plugin version pin to `0.5.1` + records SPEC-069 on the STAGE-017 backlog. The
  **tag cut itself is NOT in this PR** (it is the orchestrator/human-gated ship
  step, run from `main` after this merges). **No git tag is created in this PR.**
- **All acceptance criteria met?** **Prep ACs green; cut ACs deferred to the
  orchestrator's tag.** Verified at prep (2026-07-11):
  - `go test ./...` green; `gofmt -l .` empty; `go vet ./...` clean;
    `CGO_ENABLED=0 go build ./...` succeeds; `just test-docs` exits 0;
    `just test-hook` exits 0.
  - CHANGELOG `[0.5.1]` authored (date `2026-07-11`) + link-refs repointed;
    `test-docs.sh` Group O green.
  - `plugin/.claude-plugin/plugin.json` `version` == `0.5.1`.
  - Migration-free confirmed: `internal/storage/migrations/` holds four files
    (`0001`..`0004`), unchanged since v0.2.0.
  - **Deferred to the cut** (need the real tag / a plugin install / a throwaway
    DB): the RC-is-prerelease + no-tap-change gate, the RC binary smoke (the
    `--day` windows + mutual-exclusion error + migration-free open), the final
    release's 4 archives + checksums, the tap bump to `0.5.1`, `brew upgrade →
    0.5.1`, and the §12(b) `claude plugin details brag` registration + Stop-hook
    check on the built plugin.
- **New decisions emitted:** none.
- **Deviations from spec:** CHANGELOG dated `2026-07-11` (the intended cut date);
  if the cut slips, the orchestrator bumps the date to match.
- **Follow-up work identified:**
  - The **ship (cut)** step — orchestrator/human-gated `git tag v0.5.1-rc1` →
    smoke → delete RC → `git tag v0.5.1` (§4 Pattern 1), then the tap bump + brew
    verify, then flip STAGE-017 frontmatter `status` → shipped. Closes STAGE-017.

### Build-phase reflection (3 questions, short answers)

1. **What was unclear in the spec that slowed you down?**
   — Nothing. The SPEC-067 precedent made the mechanical/irreversible split
   obvious, and the pre-flight was a checklist. The only authoring nuance was that
   v0.5.1 is a small two-group CHANGELOG (`### Added` for the `--day` flag,
   `### Fixed` for the internal clock seam) — a patch, not a minor.

2. **Was there a constraint or decision that should have been listed but wasn't?**
   — No. `one-spec-per-pr` and the DEC-021 negative check (migration-free open)
   were both listed and both held; the migration-free property was confirmed at
   build (no new file under `internal/storage/migrations/`).

3. **If you did this task again, what would you do differently?**
   — Nothing structural. The sixth consecutive use of the R2 release-cut pre-flight
   kept the prep mechanical; a single-feature patch is the smallest cut yet.

---

## Reflection (Ship)

*Appended during the **ship** cycle. Outcome-focused, distinct from the
process-focused build reflection above. This spec's ship = the mechanical prep
merged to main; the irreversible tag/publish is the orchestrator's.*

1. **What would I do differently next time?**
   — <answer>

2. **Does any template, constraint, or decision need updating?**
   — <answer>

3. **Is there a follow-up spec I should write now before I forget?**
   — <answer>
