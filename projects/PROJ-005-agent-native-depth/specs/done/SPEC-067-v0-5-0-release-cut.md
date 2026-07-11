---
# Maps to ContextCore task.* semantic conventions.
# RELEASE-CUT variant of spec.md — the stage's closing release action.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-067
  type: story                      # a release cut is a story-sized closing action
  cycle: ship
  blocked: false                   # all v0.5.0 feature specs are on main (see Context)
  priority: high
  complexity: S

project:
  id: PROJ-005
  stage: STAGE-016
repo:
  id: bragfile

agents:
  architect: claude-opus-4-8
  implementer: claude-opus-4-8     # same Claude, different session (claude-only variant)
  created_at: 2026-07-10

references:
  decisions: [DEC-034, DEC-035, DEC-036, DEC-037, DEC-038]
  constraints: [one-spec-per-pr]
  related_specs: [SPEC-055, SPEC-056, SPEC-057, SPEC-058, SPEC-059, SPEC-060, SPEC-061, SPEC-062, SPEC-063, SPEC-064, SPEC-065, SPEC-066]
---

# SPEC-067: v0.5.0 release cut

## Context

STAGE-016's backlog — PROJ-005's first two stages (STAGE-015 MCP-first-class +
STAGE-016 v0.4.x polish, plus the pre-release hardening pass they absorbed) — is
built, dogfooded, and merged to `main`, and unreleased. This spec is the stage's
**closing release action**: it cuts and ships **v0.5.0**, a **minor** release
over v0.4.0, following [AGENTS.md §4](../../../AGENTS.md) release mechanics.

STAGE-016 was marked `shipped` one step early — its own success criteria include
"a v0.5.0 minor release cut ships the batch," so the cut is its true closing
action. The stage has been **reopened** (`status: active`, `shipped_at: null`)
to host this spec; the orchestrator re-closes it once v0.5.0 publishes (this
mirrors how SPEC-054 closed STAGE-013 for the v0.4.0 cut).

The v0.5.0 surface (all on `main`):

- **SPEC-055** (`brag mcp install`) — one command to register the `brag mcp
  serve` MCP server into a client's config (`--client claude-code|claude-desktop
  |cursor`, `--scope user|project`, `--dir`, `--dry-run`), idempotent and
  non-clobbering (DEC-034). The MCP-first-class headline.
- **SPEC-056** (`ListFilter.Until` storage promotion) — the calendar-window
  upper bound moved into `storage.ListFilter.Until`, de-duplicating the Go-side
  `created_at < end` filter across `impact`/`story`/`wrapped`/`coverage`
  (DEC-035). Behavior-preserving; goldens byte-identical.
- **SPEC-057** (`brag project ensure`) — idempotent project registration
  (create-or-no-op), closing the unregistered-project gap (DEC-036).
- **SPEC-058** (`docs/for-ai-agents.md` + README MCP section) — the full MCP
  tool contract, the `project`-not-auto-filled gotcha, provenance stamping, and
  a how-to-log-a-win playbook.
- **SPEC-059** (`brag spark`) — a sparklines-only pulse (Total + top-8
  by-project) over a rolling `--week|--month|--quarter` window (default
  `--month`); markdown default, JSON raw counts, `--no-spark`/`NO_COLOR` escape
  (DEC-037).
- **SPEC-060..066** (the pre-release hardening pass) — `brag spark` upper-bound
  fix (060), `project ensure` byte-cap parity (061), SQLite concurrency
  `busy_timeout` + immediate-tx + single-conn (062, DEC-038), `tag rename`
  canonicalization (063), unified capture-input validation across all four
  ingress paths (064), markdown-cell `|` escaping (065), and `mcp serve` clean
  shutdown exit code (066).

Like the last four cuts this spec is peeled from the feature work at design (the
release tag is cut from `main` *after* the code PRs land, so it cannot share a
PR with the code it tags — `one-spec-per-pr`). It mirrors **SPEC-054**'s v0.4.0
runbook (in turn SPEC-047's v0.3.1 / SPEC-042's v0.3.0 / SPEC-037's v0.2.0
precedents). It carries the **release runtime/operational pre-flight checklist**
(cross-project-retro R2; the release-cut template per AGENTS.md §4) so every §4
gotcha earned in prod (goreleaser dual-tag, macOS Gatekeeper, Homebrew 6.0+
brew-trust, prod-DB migration) is a **ticked design-time item**, not re-learned
in prod.

Existing tags are `v0.1.0`, `v0.2.0`, `v0.3.0`, `v0.3.1`, `v0.4.0`; v0.5.0 is
the next **minor** — a minor (not a patch) because it adds new user-facing
commands (`brag mcp install`, `brag project ensure`, `brag spark`). It is
additive and **migration-free** (no new file under
`internal/storage/migrations/`, confirmed at build).

## Goal

Cut, tag, and publish v0.5.0 to the Homebrew tap per AGENTS.md §4, verify a
clean `brew upgrade` from v0.4.0, run the §12(b) behavioral check on the built
plugin, and close STAGE-016.

## Split of responsibilities (mechanical prep vs. irreversible cut)

This spec's **PR** does the mechanical, reversible, CI-gated prep only: author
the CHANGELOG `[0.5.0]` section, bump the plugin version pin, tick the
pre-flight, get CI green, merge to `main`. The **irreversible cut** (RC tag,
smoke, dual-tag delete, final `v0.5.0` tag, goreleaser publish, tap bump, brew
verify) is driven by the coordinator/human from `main` after this merges —
verbatim per the §4 Pattern 1 sequence below. **No tag is created in this PR.**

## Inputs

- **Files to read:** [AGENTS.md §4](../../../AGENTS.md) (release mechanics + the
  three lessons-earned addenda — dual-tag Pattern 1, Gatekeeper, brew-trust);
  [`SPEC-054`](../../PROJ-004-story-surface/specs/done/SPEC-054-v0-4-0-release-cut.md)
  (the v0.4.0 runbook this mirrors); `CHANGELOG.md` (the `[Unreleased]` section
  + the compare-link refs); `.goreleaser.yaml` + `.github/workflows/release.yml`
  (the cut machinery); `README.md` §Install (Gatekeeper xattr + brew-trust
  notes); `plugin/.claude-plugin/plugin.json` (the version pin, `0.4.0` → bump
  to `0.5.0`); the DECs/specs of the surface the CHANGELOG describes
  (DEC-034..038 + SPEC-055..066).
- **External:** GitHub Releases; the `jysf/homebrew-bragfile` tap.
- **Related code paths:** none — this spec ships **no Go code**.

## Outputs

- **Files created:** `projects/PROJ-005-agent-native-depth/specs/SPEC-067-...md`
  (this spec; archived to `done/` at ship).
- **Files modified:**
  - `CHANGELOG.md` — a dated `[0.5.0]` section (`### Added` / `### Changed` /
    `### Fixed` + an `### Upgrading from v0.4.0` block); compare-links repointed.
  - `plugin/.claude-plugin/plugin.json` — `version` `0.4.0` → `0.5.0`.
  - `projects/PROJ-005-agent-native-depth/stages/STAGE-016-...md` — reopened to
    `status: active` (`shipped_at: null`), a reopened note added, SPEC-067 on the
    backlog. (Stage frontmatter is re-closed by the coordinator after the tag is
    live; PROJ-005 frontmatter is likewise left for the coordinator.)
  - `jysf/homebrew-bragfile` (separate repo) — `bragfile` cask bumped to `0.5.0`
    with matching sha256s (goreleaser publishes this **at the cut**, not in this
    PR).
- **New exports:** none.
- **Database changes:** **none** — v0.5.0 is migration-free. All feature specs
  are read-side / config-side / hardening (they add CLI wiring, a docs file, and
  storage-layer safety over existing schema; no new migration file), so the
  DEC-021 auto-backup safety belt does **not** fire on a v0.4.0 open.

## The CHANGELOG `[0.5.0]` literal (transcribed at build)

Literal-artifact-as-spec (§12). Build inserts a dated section immediately below
`## [Unreleased]` (which stays, empty), grouping the surface under `### Added` /
`### Changed` / `### Fixed`, plus an `### Upgrading from v0.4.0` block, and
repoints the link refs at the file's bottom. The ten-verb Group-O4 assertion
already passes from the `[0.1.0]` section and is unaffected. See the merged
`CHANGELOG.md` `[0.5.0]` section for the authored text; the link-reference block
becomes:

```markdown
[Unreleased]: https://github.com/jysf/bragfile000/compare/v0.5.0...HEAD
[0.5.0]: https://github.com/jysf/bragfile000/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/jysf/bragfile000/compare/v0.3.1...v0.4.0
...
```

## Acceptance Criteria

Each gate is independently checkable. Prep gates (this PR) are green before the
PR opens; cut gates describe the observable end state the coordinator produces
at the tag (mirrors SPEC-054).

- [x] **Prep gates green** (this PR, before it opens): `go test ./...` green;
      `gofmt -l .` empty; `go vet ./...` clean; `CGO_ENABLED=0 go build ./...`
      succeeds; `just test-docs` exits 0; `just test-hook` exits 0.
- [x] **CHANGELOG reconciled:** a dated `## [0.5.0] - 2026-07-10` section is
      present, `[Unreleased]` stays present-and-empty, and the compare-links are
      repointed (`test-docs.sh` Group O stays green).
- [x] **Plugin version pin matches the tag:** `plugin/.claude-plugin/plugin.json`
      `version` == `0.5.0` == the intended git tag.
- [x] **Pre-flight all ticked** (design; the checklist under Notes) with either
      concrete evidence or an explicit "verified by orchestrator at the cut".
- [x] **Migration-free confirmed at build:** no new file under
      `internal/storage/migrations/` since v0.4.0 (four files: `0001`..`0004`,
      unchanged). v0.5.0 is additive / read-side / hardening only.
- [ ] **rc is a GitHub PRERELEASE with no tap change** *(cut)*: after
      `v0.5.0-rc1` CI completes, its GitHub release is *Pre-release* and the
      `jysf/homebrew-bragfile` tap's latest commit is **unchanged**.
- [ ] **RC smoke gate passes on a THROWAWAY DB** *(cut, never `~/.bragfile`)*:
      the rc binary reports `0.5.0-rc1`; `brag mcp install --dry-run`, `brag
      project ensure <name>`, and `brag spark --month` each behave as specified
      (idempotent install merge, create-or-no-op, sparkline pulse markdown + JSON
      raw counts); the concurrency fix is observable (a second `brag add` while
      `brag mcp serve` holds the DB waits and succeeds, not `database is
      locked`); and — because v0.5.0 is migration-free — opening a seeded
      v0.4.0-schema throwaway DB fires **no** migration and writes **no** backup
      sidecar (DEC-021 trigger `applied>0 && pending>0` is false).
- [ ] **Behavioral surfaces re-checked on the built artifact** *(cut, §12(b)
      refinement)*: against a clean install of the built plugin, `claude plugin
      details brag` shows the MCP server **registered** (not just `validate
      --strict` green), and the Stop hook fires once in a throwaway repo.
- [ ] **Final release has all 4 platform archives + checksums** *(cut)*: the
      `v0.5.0` GitHub release (NOT a prerelease) carries the four `darwin/linux ×
      amd64/arm64` tarballs and `checksums.txt`.
- [ ] **Tap bumped to 0.5.0** *(cut)*: `jysf/homebrew-bragfile` has a new commit
      setting the `bragfile` cask to `0.5.0` with matching sha256s.
- [ ] **Brew binary reports 0.5.0** *(cut)*: after `brew update && brew upgrade
      jysf/bragfile/bragfile` (+ `brew trust --cask` first-time, + the Gatekeeper
      xattr clear), `brag --version` from the brew path reports `0.5.0`.
- [ ] **Prod DB opens with no migration** *(cut)*: opening the real `~/.bragfile`
      (already v0.4.0) with the released binary fires no migration and writes no
      new backup sidecar.
- [x] **No regressions:** the full Go suite, `gofmt`, `go vet`, and the
      `CGO_ENABLED=0` build stay green (verified at prep; this spec adds no code
      that could break them).

## Failing Tests

This spec ships **no Go code**, so there are no new `*_test.go` files — the
`test-before-implementation` constraint is satisfied the way SPEC-054/047/042/037
(also code-free) satisfied it: the enforcement mechanism is the **prep gate
suite** (the existing green `go test ./...`, `scripts/test-docs.sh` including
Group O's CHANGELOG-shape assertions, and `just test-hook`) plus the
**post-release verification checklist** (the ACs above), not new Go assertions.
"The test" for a release runbook is the executable gate, not a Go assertion. The
regression guard is the pre-existing suite, run unchanged at prep.

---

## The build cycle is CHANGELOG + version-bump prep (read before building)

The build cycle for SPEC-067 does **not** cut the release. It authors the
reversible artifacts so the real cut (coordinator + human go-aheads) is
mechanical:

1. **Author the CHANGELOG `[0.5.0]`** (date `2026-07-10`) — the new commands
   under `### Added`, the concurrency/validation/`Until` work under `###
   Changed`, the fixes under `### Fixed`, + the `### Upgrading from v0.4.0`
   block — repoint the link refs, and run `just test-docs` — Group O must stay
   green.
2. **Bump `plugin/.claude-plugin/plugin.json`** `version` to `0.5.0`.
3. **Confirm migration-free** — no new file under `internal/storage/migrations/`.
4. **Run the six prep gates** (below) and confirm all green.
5. **Reopen STAGE-016** — `status: active`, `shipped_at: null`, a reopened note,
   SPEC-067 on the backlog. Leave the re-close (and PROJ-005 frontmatter) to the
   coordinator after the tag publishes.

No tag is pushed; `goreleaser release` is never invoked; no `gh release` call and
no `brew` command is made in this cycle. Those are the coordinator's, below.

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
brag --version                                  # → 0.5.0-rc1
# Exercise the new commands:
brag mcp install --client claude-code --scope project --dir "$(mktemp -d)" --dry-run
brag project ensure demo                        # create-or-no-op; run twice → idempotent
brag spark --month                              # Total + top-8 by-project sparkline pulse
brag spark --month --format json | jq .         # raw per-bucket counts, no glyphs
NO_COLOR=1 brag spark --month                   # glyphs dropped; raw counts
# Concurrency (DEC-038): a write while the server holds the DB WAITS, not "locked":
brag mcp serve & sleep 1; brag add --title "concurrent" ; kill %1
# Migration-free upgrade: seed a v0.4.0-schema throwaway DB, open it, assert
# NO *.backup sidecar appears (applied>0 && pending>0 is false at v0.5.0).
```

## Destructive sequence (coordinator/human-gated — §4 Pattern 1)

The real cut, verbatim, executed only after this PR merges and the human
go-aheads are given:

```bash
# 0. main is clean, at the release SHA (this PR's merge commit), local == origin;
#    CHANGELOG [0.5.0] + plugin version 0.5.0 merged.
# 1. Optional RC to exercise CI + the tap path:
git tag v0.5.0-rc1 && git push origin v0.5.0-rc1     # CI builds a PRERELEASE; skip_upload:auto holds the tap
# 2. Run the RC SMOKE GATE above against the downloaded rc binary. If good:
# 3. Dual-tag-on-same-commit — delete the RC tag + release BEFORE the final tag (§4 Pattern 1):
gh release delete v0.5.0-rc1 --yes --cleanup-tag
git tag v0.5.0 && git push origin v0.5.0             # CI cuts the real release + bumps the tap
# 4. Post-release verification: the AC checklist (4 archives + checksums, tap 0.5.0,
#    brew upgrade → 0.5.0, prod DB no-migration).
# 5. Flip STAGE-016 frontmatter status → shipped + shipped_at once v0.5.0 is live.
```

## Implementation Context

*Read this section (and the files it points to) before starting the build cycle.*

### Decisions that apply

- `DEC-034` — the `brag mcp install` config-merge scheme (idempotent,
  non-clobbering, per-client target resolution); the CHANGELOG describes its
  surface.
- `DEC-035` — the `ListFilter.Until` storage promotion (SQL `created_at < ?`
  guarded by `!Until.IsZero()`), de-duplicating the four Go-side consumers.
- `DEC-036` — `brag project ensure` idempotent upsert (create-or-no-op) +
  byte-counted name cap parity with the capture paths.
- `DEC-037` — the `brag spark` rolling windows + multi-row pulse render
  (`--week|--month|--quarter`, Total + top-8 by-project, JSON raw counts).
- `DEC-038` — SQLite concurrency: `busy_timeout` + `_txlock=immediate` + single
  connection so concurrent access waits instead of failing `database is locked`
  (WAL deferred).
- `DEC-021` — the migration auto-backup safety belt; relevant here as the
  *negative* check (v0.5.0 is migration-free, so it must NOT fire).

### Constraints that apply

- `one-spec-per-pr` — the release tag is cut from `main` after the feature PRs
  landed; this spec is its own PR, separate from the twelve feature/fix PRs.

### Prior related work

- `SPEC-054` (shipped) — the v0.4.0 release-cut runbook this mirrors (pre-flight
  → RC prerelease → smoke → dual-tag → tap bump → brew verify); it closed
  STAGE-013 / PROJ-004 the same way this closes STAGE-016.
- `SPEC-055..066` (all shipped, on `main`) — the v0.5.0 surface this release
  publishes.

### Out of scope (for this spec specifically)

- macOS notarization (still deferred; the Gatekeeper xattr note stands).
- Any code change to the v0.5.0 surface — if the cut surfaces a bug, that is a
  new fix spec, not an edit here.
- The deeper PROJ-005 agent-native stages (memory, signed provenance,
  capture-completeness, model-benchmark) — later stages, not this cut.

## Notes for the Implementer

Gotchas, style preferences, reuse opportunities.

- The `[Unreleased]` section is **empty** at prep — the feature specs shipped
  without CHANGELOG entries, so the `[0.5.0]` section is authored whole (not
  moved from accumulated Unreleased notes). `[Unreleased]` stays as a
  present-but-empty heading (Group O5 asserts the `[Unreleased]:` ref present).
- `test-docs.sh` Group O asserts: `[Unreleased]:` and `[0.1.0]:` link refs
  present (O5), the ten verbs in backticks (O4 — already satisfied by the
  `[0.1.0]` section), the `## [0.1.0]` heading (O3), `keepachangelog.com` (O2).
  Keep all of these; add the `[0.5.0]:` ref and repoint `[Unreleased]:` to
  `v0.5.0...HEAD`.
- The plugin version pin is `0.4.0` today — bump it to `0.5.0` to match the tag
  (the pre-flight requires the pin == the tag). `plugin/.mcp.json` carries **no**
  version field, so it needs no bump.
- **Minor, not patch:** v0.5.0 adds new user-facing commands (`brag mcp
  install`, `brag project ensure`, `brag spark`), so it is a minor over v0.4.0.
  Unlike v0.4.0 (Added-only), this cut also carries `### Changed` (concurrency,
  unified capture validation, the `Until` internal refactor) and `### Fixed`
  (tag-rename, mcp-shutdown, markdown-pipe, spark window, reserved-tag reject).

### Release runtime/operational pre-flight (all must be ticked at design)

Adopted from the release-cut spec template (`projects/_templates/spec-release-cut.md`,
per AGENTS.md §4). Concretized for the v0.5.0 `brag` cut. Items verifiable now
carry concrete evidence; items that can only be checked AT the tag/RC are marked
"verified by orchestrator at the cut" (honest — not faked).

- [x] Dual-tag-on-same-commit: RC tag + release deleted before the final tag is
      cut at the same commit (§4 Pattern 1). — **Documented** in the Destructive
      sequence above (verbatim §4 Pattern 1 commands: `gh release delete
      v0.5.0-rc1 --yes --cleanup-tag` then `git tag v0.5.0 && git push`).
      *Executed by orchestrator at the cut.*
- [x] macOS Gatekeeper: `xattr -dr com.apple.quarantine <bin>` note present in
      README §Install. — **Evidence:** the README "macOS Gatekeeper note" +
      `xattr -dr com.apple.quarantine` command in the install section (unchanged
      since v0.4.0; verified present at prep). *The one-time clear is run by
      orchestrator at the cut.*
- [x] Homebrew 6.0+: `brew trust --cask <tap>/<cask>` documented in README and
      run once at the cut. — **Evidence:** `brew trust --cask
      jysf/bragfile/bragfile` documented in the README install section (under the
      "Homebrew 6.0+ note"). *The one-time `brew trust` + `brew upgrade` run is
      verified by orchestrator at the cut.*
- [x] Dev/prod DB isolation: the RC smoke test runs against a THROWAWAY DB, never
      ~/.bragfile; the auto-backup path is observed. — **Evidence:** the RC smoke
      gate above sets `BRAGFILE_DB=$(mktemp -d)/smoke.db`. For v0.5.0 the DEC-021
      backup path is asserted **not** to fire (migration-free); *the throwaway-DB
      smoke is run by orchestrator at the cut.*
- [x] Clean upgrade: `brew upgrade` from the prior minor verified; `brag
      --version` prints the new tag; no migration surprise. — **Verified by
      orchestrator at the cut** (`brew upgrade jysf/bragfile/bragfile` from v0.4.0
      → `brag --version` == `0.5.0`; migration-free per DEC-021 negative check).
      Design evidence that it *will* be clean: **confirmed at build** that no
      migration file was added under `internal/storage/migrations/` (four files
      `0001`..`0004`, unchanged since v0.2.0).
- [x] Prod-DB migration = **N/A this release**: v0.5.0 adds no migration, so the
      DEC-021 backup safety belt does not fire on a v0.4.0→v0.5.0 open. —
      **Confirmed at build** (no new file under `internal/storage/migrations/`).
- [x] CHANGELOG: the `[0.5.0]` dated section is authored; compare-links
      repointed. — **Done in this PR** (the surface under `### Added`/`###
      Changed`/`### Fixed` + the `### Upgrading from v0.4.0` block; Group O stays
      green — verified at prep).
- [x] Plugin version pin (v0.3.0+): `plugin/.claude-plugin/plugin.json` `version`
      matches the tag. — **Done in this PR** (bumped `0.4.0` → `0.5.0`).
- [x] Behavioral surfaces re-checked on the built artifact (§12(b) refinement):
      `claude plugin details` shows the MCP server registered; the Stop hook
      fires in a throwaway repo. — **Verified by orchestrator at the cut** (needs
      the built plugin + a clean Claude install; not run here to avoid mutating
      the user's Claude config, per SPEC-054/047's precedent).

---

## Build Completion

*Filled in at the end of the **build** (prep) cycle, before advancing to verify.*

- **Branch:** `feat/spec-067-v0-5-0-release-cut`
- **PR (if applicable):** this PR — authors the CHANGELOG `[0.5.0]` + bumps the
  plugin version pin to `0.5.0` + reopens STAGE-016 to host the cut. The **tag
  cut itself is NOT in this PR** (it is the coordinator/human-gated ship step,
  run from `main` after this merges). **No git tag is created in this PR.**
- **All acceptance criteria met?** **Prep ACs green; cut ACs deferred to the
  orchestrator's tag.** Verified at prep (2026-07-10):
  - `go test ./...` green; `gofmt -l .` empty; `go vet ./...` clean;
    `CGO_ENABLED=0 go build ./...` succeeds; `just test-docs` exits 0;
    `just test-hook` exits 0.
  - CHANGELOG `[0.5.0]` authored (date `2026-07-10`) + link-refs repointed;
    `test-docs.sh` Group O green.
  - `plugin/.claude-plugin/plugin.json` `version` == `0.5.0`.
  - Migration-free confirmed: `internal/storage/migrations/` holds four files
    (`0001`..`0004`), unchanged since v0.2.0.
  - **Deferred to the cut** (need the real tag / a plugin install / a throwaway
    DB): the RC-is-prerelease + no-tap-change gate, the RC binary smoke (the new
    commands + concurrency-waits + migration-free open), the final release's 4
    archives + checksums, the tap bump to `0.5.0`, `brew upgrade → 0.5.0`, and
    the §12(b) `claude plugin details brag` registration + Stop-hook check on the
    built plugin.
- **New decisions emitted:** none.
- **Deviations from spec:** CHANGELOG dated `2026-07-10` (the intended cut date);
  if the cut slips, the orchestrator bumps the date to match. An `### Upgrading
  from v0.4.0` block was included for consistency with every prior release entry
  (the transcribed literal covered Added/Changed/Fixed; the upgrade note follows
  house style and the pre-flight's clean-upgrade item).
- **Follow-up work identified:**
  - The **ship (cut)** step — coordinator/human-gated `git tag v0.5.0-rc1` →
    smoke → delete RC → `git tag v0.5.0` (§4 Pattern 1), then the tap bump + brew
    verify, then flip STAGE-016 frontmatter `status` → shipped. Closes STAGE-016.

### Build-phase reflection (3 questions, short answers)

1. **What was unclear in the spec that slowed you down?**
   — Nothing. The SPEC-054 precedent made the mechanical/irreversible split
   obvious, and the pre-flight was a checklist. The one authoring nuance vs.
   v0.4.0 was that v0.5.0 is not Added-only — it also carries `### Changed` and
   `### Fixed` for the hardening pass, so the CHANGELOG had three groups.

2. **Was there a constraint or decision that should have been listed but wasn't?**
   — No. `one-spec-per-pr` and the DEC-021 negative check (migration-free open)
   were both listed and both held; the migration-free property was confirmed at
   build (no new file under `internal/storage/migrations/`).

3. **If you did this task again, what would you do differently?**
   — Nothing structural. The fifth consecutive use of the R2 release-cut
   pre-flight kept the prep mechanical; the only added work over a patch cut is a
   longer, three-group CHANGELOG.

---

## Reflection (Ship)

*Appended during the **ship** cycle. Outcome-focused, distinct from the
process-focused build reflection above. This spec's ship = the mechanical prep
merged to main; the irreversible tag/publish is the orchestrator's.*

1. **What would I do differently next time?**
   — Cut the release cut BEFORE closing its stage. STAGE-016 was marked
   shipped one step early (during the fix-spec archival), so this spec had to
   reopen it — the stage's own success criteria include the v0.5.0 cut, so the
   cut is its closing action (as SPEC-054 was for STAGE-013). Sequence: ship
   the last feature/fix specs → cut the release → then close the stage.

2. **Does any template, constraint, or decision need updating?**
   — No. The §4 Pattern-1 pre-flight held exactly: the RC (`v0.5.0-rc1`)
   validated the goreleaser workflow, the RC release+tag were deleted before
   tagging `v0.5.0` at the same commit (no dual-tag 422), and the macOS
   Gatekeeper quarantine on the fresh cask binary needed the documented
   `xattr -dr com.apple.quarantine` — every §4 lesson-earned item fired as a
   ticked step, none as a surprise. brew-trust did not re-prompt (tap already
   trusted since v0.2.0).

3. **Is there a follow-up spec I should write now before I forget?**
   — Not for this release. One minor observation for the backlog: `brag spark`
   excludes an entry created in the same truncated-second as the spark run
   (the SPEC-060 exclusive `[start, now)` upper edge) — self-resolves the next
   second, consistent header/bucket behavior, low-priority polish only. The
   deferred audit LOW/NITs and the full provenance-signing work belong to the
   deeper PROJ-005 stages, which stay ahead (PROJ-005 remains active).
