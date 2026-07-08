---
# Maps to ContextCore task.* semantic conventions.
# RELEASE-CUT variant of spec.md — the stage's closing release action.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-054
  type: story                      # a release cut is a story-sized closing action
  cycle: ship
  blocked: false                   # all v0.4.0 feature specs are on main (see Context)
  priority: high
  complexity: S

project:
  id: PROJ-004
  stage: STAGE-013
repo:
  id: bragfile

agents:
  architect: claude-opus-4-8
  implementer: claude-opus-4-8     # same Claude, different session (claude-only variant)
  created_at: 2026-07-07

references:
  decisions: [DEC-028, DEC-029, DEC-030, DEC-031, DEC-032, DEC-033, DEC-021]
  constraints: [one-spec-per-pr, timestamps-in-utc-rfc3339]
  related_specs: [SPEC-047, SPEC-048, SPEC-049, SPEC-050, SPEC-051, SPEC-052, SPEC-053, SPEC-045]
---

# SPEC-054: v0.4.0 release cut

## Context

STAGE-013's feature backlog — the whole PROJ-004 read/story surface — is
built, dogfooded, and merged to `main`, and unreleased. This spec is the
stage's (and the project's) **closing release action**: it cuts and ships
**v0.4.0**, a **minor** release over v0.3.1, following
[AGENTS.md §4](../../../AGENTS.md) release mechanics.

The v0.4.0 surface (all on `main`):

- **SPEC-048** (`brag impact`) — the calendar-windowed, initiative-grouped
  impact digest; the deterministic data foundation the story surface reads
  (DEC-028).
- **SPEC-049 / SPEC-050** (`brag story --audience me|manager|skip|exec`) — the
  narrative shaping surface: the four-audience gradient, the coalesce-into-arcs
  bundle, and the embedded framing directives; a pure pipe (LLM optional),
  profiles-as-data (DEC-029).
- **SPEC-051** (`brag wrapped [year|quarter]`) — the shareable year/quarter-in-
  review digest, fifth DEC-014 consumer (DEC-030).
- **SPEC-052** (the sparklines/visual pass) — an in-terminal Unicode block-glyph
  sparkline over `wrapped`'s cadence, default-on with `--no-spark`/`NO_COLOR`
  escape, JSON raw (DEC-031).
- **SPEC-053** (`--previous`) — the last-completed-period window modifier across
  `impact`/`story`/`wrapped` (DEC-032, extends DEC-028).
- **SPEC-045** (`brag coverage`) — the P3 agent-assist provenance-share metric,
  sixth DEC-014 consumer (DEC-033).

Like the last three cuts this spec is peeled from the feature work at design
(the release tag is cut from `main` *after* the code PRs land, so it cannot
share a PR with the code it tags — `one-spec-per-pr`). It mirrors **SPEC-047**'s
v0.3.1 runbook (which mirrored SPEC-042's v0.3.0 cut, in turn SPEC-037's v0.2.0
precedent). It carries the **release runtime/operational pre-flight checklist**
(cross-project-retro R2; the release-cut template per AGENTS.md §4) so every §4
gotcha earned in prod (goreleaser dual-tag, macOS Gatekeeper, Homebrew 6.0+
brew-trust, prod-DB migration) is a **ticked design-time item**, not re-learned
in prod.

Existing tags are `v0.1.0`, `v0.2.0`, `v0.3.0`, `v0.3.1`; v0.4.0 is the next
**minor** — a minor (not a patch) because it adds six new user-facing surfaces
(`impact`, `story`, `wrapped`, sparklines, `--previous`, `coverage`). It is
additive and **migration-free** (all v0.4.0 work is read-side — no new file
under `internal/storage/migrations/`, confirmed at build).

## Goal

Cut, tag, and publish v0.4.0 to the Homebrew tap per AGENTS.md §4, verify a
clean `brew upgrade` from v0.3.1, run the §12(b) behavioral check on the built
plugin, and close STAGE-013 (and PROJ-004).

## Split of responsibilities (mechanical prep vs. irreversible cut)

This spec's **PR** does the mechanical, reversible, CI-gated prep only:
author the CHANGELOG `[0.4.0]` section, bump the plugin version pin, tick the
pre-flight, get CI green, merge to `main`. The **irreversible cut** (RC tag,
smoke, dual-tag delete, final `v0.4.0` tag, goreleaser publish, tap bump, brew
verify) is driven by the coordinator/human from `main` after this merges —
verbatim per the §4 Pattern 1 sequence below. **No tag is created in this PR.**

## Inputs

- **Files to read:** [AGENTS.md §4](../../../AGENTS.md) (release mechanics +
  the three lessons-earned addenda — dual-tag Pattern 1, Gatekeeper, brew-trust);
  [`SPEC-047`](done/SPEC-047-v0-3-1-release-cut.md) (the v0.3.1 runbook this
  mirrors); `CHANGELOG.md` (the `[Unreleased]` section + the compare-link refs);
  `.goreleaser.yaml` + `.github/workflows/release.yml` (the cut machinery);
  `README.md` §Install (Gatekeeper xattr + brew-trust notes);
  `plugin/.claude-plugin/plugin.json` (the version pin, `0.3.1` → bump to
  `0.4.0`); the DECs/specs of the surface the CHANGELOG describes
  (DEC-028..033 + SPEC-048/049/050/051/052/053/045).
- **External:** GitHub Releases; the `jysf/homebrew-bragfile` tap.
- **Related code paths:** none — this spec ships **no Go code**.

## Outputs

- **Files created:** `projects/PROJ-004-story-surface/specs/SPEC-054-...md`
  (this spec; archived to `done/` at ship).
- **Files modified:**
  - `CHANGELOG.md` — a dated `[0.4.0]` section (`### Added` grouping the six
    surfaces + an `### Upgrading from v0.3.1` block); compare-links repointed.
  - `plugin/.claude-plugin/plugin.json` — `version` `0.3.1` → `0.4.0`.
  - `projects/PROJ-004-story-surface/stages/STAGE-013-...md` — backlog marks
    SPEC-054 `[x]` shipped; count bumped. (Stage frontmatter `status` stays
    `active` — NOT flipped to `shipped` until the tag is live; the coordinator
    does that after v0.4.0 publishes. PROJ-004 frontmatter is likewise left for
    the coordinator.)
  - `jysf/homebrew-bragfile` (separate repo) — `bragfile` cask bumped to
    `0.4.0` with matching sha256s (goreleaser publishes this **at the cut**, not
    in this PR).
- **New exports:** none.
- **Database changes:** **none** — v0.4.0 is migration-free. All six feature
  specs are read-side (they add `internal/export`/`internal/story`/`internal/spark`
  renderers and CLI wiring over existing storage; no new migration file), so the
  DEC-021 auto-backup safety belt does **not** fire on a v0.3.1→v0.4.0 open.

## The CHANGELOG `[0.4.0]` literal (transcribed at build)

Literal-artifact-as-spec (§12). Build inserts a dated section immediately below
`## [Unreleased]` (which stays, empty), grouping the six surfaces under `###
Added`, plus an `### Upgrading from v0.3.1` block, and repoints the link refs at
the file's bottom. The ten-verb Group-O4 assertion already passes from the
`[0.1.0]` section and is unaffected. See the merged `CHANGELOG.md` `[0.4.0]`
section for the authored text; the link-reference block becomes:

```markdown
[Unreleased]: https://github.com/jysf/bragfile000/compare/v0.4.0...HEAD
[0.4.0]: https://github.com/jysf/bragfile000/compare/v0.3.1...v0.4.0
[0.3.1]: https://github.com/jysf/bragfile000/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/jysf/bragfile000/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/jysf/bragfile000/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/jysf/bragfile000/releases/tag/v0.1.0
```

## Acceptance Criteria

Each gate is independently checkable. Prep gates (this PR) are green before the
PR opens; cut gates describe the observable end state the coordinator produces
at the tag (mirrors SPEC-047).

- [x] **Prep gates green** (this PR, before it opens): `go test ./...` green;
      `gofmt -l .` empty; `go vet ./...` clean; `CGO_ENABLED=0 go build ./...`
      succeeds; `just test-docs` exits 0; `just test-hook` exits 0.
- [x] **CHANGELOG reconciled:** a dated `## [0.4.0] - 2026-07-07` section is
      present, `[Unreleased]` stays present-and-empty, and the compare-links are
      repointed (`test-docs.sh` Group O stays green).
- [x] **Plugin version pin matches the tag:** `plugin/.claude-plugin/plugin.json`
      `version` == `0.4.0` == the intended git tag.
- [x] **Pre-flight all ticked** (design; the checklist under Notes) with either
      concrete evidence or an explicit "verified by orchestrator at the cut".
- [x] **Migration-free confirmed at build:** no new file under
      `internal/storage/migrations/` since v0.3.1 (four files: `0001`..`0004`,
      unchanged). v0.4.0 is read-side only.
- [ ] **rc is a GitHub PRERELEASE with no tap change** *(cut)*: after
      `v0.4.0-rc1` CI completes, its GitHub release is *Pre-release* and the
      `jysf/homebrew-bragfile` tap's latest commit is **unchanged**.
- [ ] **RC smoke gate passes on a THROWAWAY DB** *(cut, never `~/.bragfile`)*:
      the rc binary reports `0.4.0-rc1`; `brag impact --quarter`, `brag story
      --audience exec`, `brag wrapped quarter`, and `brag coverage` each render a
      DEC-014 envelope in markdown and JSON; `wrapped`'s `## Cadence` shows a
      sparkline in a TTY and drops it under `--no-spark`/`NO_COLOR` with JSON
      unchanged; `--previous` shifts a bounded window; and — because v0.4.0 is
      migration-free — opening a seeded v0.3.1-schema throwaway DB fires **no**
      migration and writes **no** backup sidecar (DEC-021 trigger `applied>0 &&
      pending>0` is false).
- [ ] **Behavioral surfaces re-checked on the built artifact** *(cut, §12(b)
      refinement)*: against a clean install of the built plugin, `claude plugin
      details brag` shows the MCP server **registered** (not just `validate
      --strict` green), and the Stop hook fires once in a throwaway repo.
- [ ] **Final release has all 4 platform archives + checksums** *(cut)*: the
      `v0.4.0` GitHub release (NOT a prerelease) carries the four `darwin/linux
      × amd64/arm64` tarballs and `checksums.txt`.
- [ ] **Tap bumped to 0.4.0** *(cut)*: `jysf/homebrew-bragfile` has a new commit
      setting the `bragfile` cask to `0.4.0` with matching sha256s.
- [ ] **Brew binary reports 0.4.0** *(cut)*: after `brew update && brew upgrade
      jysf/bragfile/bragfile` (+ `brew trust --cask` first-time, + the Gatekeeper
      xattr clear), `brag --version` from the brew path reports `0.4.0`.
- [ ] **Prod DB opens with no migration** *(cut)*: opening the real
      `~/.bragfile` (already v0.3.1) with the released binary fires no migration
      and writes no new backup sidecar.
- [x] **No regressions:** the full Go suite, `gofmt`, `go vet`, and the
      `CGO_ENABLED=0` build stay green (verified at prep; this spec adds no code
      that could break them).

## Failing Tests

This spec ships **no Go code**, so there are no new `*_test.go` files — the
`test-before-implementation` constraint is satisfied the way SPEC-047/042/037
(also code-free) satisfied it: the enforcement mechanism is the **prep gate
suite** (the existing green `go test ./...`, `scripts/test-docs.sh` including
Group O's CHANGELOG-shape assertions, and `just test-hook`) plus the
**post-release verification checklist** (the ACs above), not new Go assertions.
"The test" for a release runbook is the executable gate, not a Go assertion. The
regression guard is the pre-existing suite, run unchanged at prep.

---

## The build cycle is CHANGELOG + version-bump prep (read before building)

The build cycle for SPEC-054 does **not** cut the release. It authors the
reversible artifacts so the real cut (coordinator + human go-aheads) is
mechanical:

1. **Author the CHANGELOG `[0.4.0]`** (date `2026-07-07`) — the six surfaces
   under `### Added` + the `### Upgrading from v0.3.1` block — repoint the link
   refs, and run `just test-docs` — Group O must stay green.
2. **Bump `plugin/.claude-plugin/plugin.json`** `version` to `0.4.0`.
3. **Confirm migration-free** — no new file under `internal/storage/migrations/`.
4. **Run the six prep gates** (below) and confirm all green.
5. **Update the STAGE-013 backlog** — mark SPEC-054 `[x]` shipped, bump the
   count. Leave the stage frontmatter `status` untouched (the coordinator flips
   it after the tag publishes); leave PROJ-004 frontmatter untouched too.

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
brag --version                                  # → 0.4.0-rc1
# Seed a few entries, then exercise the six new surfaces (markdown + json):
brag impact --quarter --format json | jq .      # DEC-014 envelope
brag story --audience exec                      # arc bundle; --print-directive prints the framing directive
brag story --audience manager --format json | jq .
brag wrapped quarter                            # ## Cadence shows a sparkline in a TTY
NO_COLOR=1 brag wrapped quarter                 # sparkline dropped; --no-spark does the same; JSON unchanged
brag wrapped quarter --previous                 # bounded [prev-start, prev-end) window
brag coverage --format json | jq .              # agent/human share + agent-share sparkline (markdown) + self-ref density
# Plugin registration (the §12(b) refinement check):
claude plugin marketplace add jysf/bragfile000
claude plugin install brag@bragfile
claude plugin details brag                      # → MCP servers (1) brag  (registered, not just valid)
# Migration-free upgrade: seed a v0.3.1-schema throwaway DB, open it, assert
# NO *.backup sidecar appears (applied>0 && pending>0 is false at v0.4.0).
```

## Destructive sequence (coordinator/human-gated — §4 Pattern 1)

The real cut, verbatim, executed only after this PR merges and the human
go-aheads are given:

```bash
# 0. main is clean, at the release SHA (this PR's merge commit), local == origin;
#    CHANGELOG [0.4.0] + plugin version 0.4.0 merged.
# 1. Optional RC to exercise CI + the tap path:
git tag v0.4.0-rc1 && git push origin v0.4.0-rc1     # CI builds a PRERELEASE; skip_upload:auto holds the tap
# 2. Run the RC SMOKE GATE above against the downloaded rc binary. If good:
# 3. Dual-tag-on-same-commit — delete the RC tag + release BEFORE the final tag (§4 Pattern 1):
gh release delete v0.4.0-rc1 --yes --cleanup-tag
git tag v0.4.0 && git push origin v0.4.0             # CI cuts the real release + bumps the tap
# 4. Post-release verification: the AC checklist (4 archives + checksums, tap 0.4.0,
#    brew upgrade → 0.4.0, prod DB no-migration).
# 5. Flip STAGE-013 frontmatter status → shipped + shipped_at once v0.4.0 is live;
#    close PROJ-004.
```

## Implementation Context

*Read this section (and the files it points to) before starting the build cycle.*

### Decisions that apply

- `DEC-028` — the impact digest window + shape (the calendar-window machinery
  `story`/`wrapped`/`--previous` all reuse); the CHANGELOG describes its surface.
- `DEC-029` — story audience shaping-profiles + thread definition (the
  four-audience gradient, arc coalescing, framing directives, profiles-as-data).
- `DEC-030` — wrapped period selection + section taxonomy (the fifth digest).
- `DEC-031` — the sparkline primitive: normalization + placement +
  default-on-with-escape (JSON raw).
- `DEC-032` — `--previous` = last-completed calendar period, bounded
  `[prev-start, prev-end)`, shared across the family (extends DEC-028).
- `DEC-033` — coverage metric definition + surface; classifier unification with
  `brag list --author`.
- `DEC-021` — the migration auto-backup safety belt; relevant here as the
  *negative* check (v0.4.0 is migration-free, so it must NOT fire).

### Constraints that apply

- `one-spec-per-pr` — the release tag is cut from `main` after the feature PRs
  landed; this spec is its own PR, separate from the six feature PRs.
- `timestamps-in-utc-rfc3339` — untouched; the story surface reads existing
  timestamps and derives calendar windows in Go, changing no storage semantics.

### Prior related work

- `SPEC-047` (shipped) — the v0.3.1 release-cut runbook this mirrors (pre-flight
  → RC prerelease → smoke → dual-tag → tap bump → brew verify); its ship
  reflection confirmed the release-cut template fit-for-purpose for a patch.
- `SPEC-048/049/050/051/052/053/045` (all shipped, on `main`) — the v0.4.0
  surface this release publishes.

### Out of scope (for this spec specifically)

- macOS notarization (still deferred; the Gatekeeper xattr note stands).
- Any code change to the v0.4.0 surface — if the cut surfaces a bug, that is a
  new fix spec, not an edit here.
- Team/multi-user federation + token-economics reconciliation (PROJ-005), and
  the `ListFilter.Until` storage promotion SPEC-053 surfaced (backlog, not
  blocking v0.4.0).
- An external-plotter pipe beyond the in-terminal sparkline (deferred).

## Notes for the Implementer

Gotchas, style preferences, reuse opportunities.

- The `[Unreleased]` section is **empty** at prep — the six feature specs shipped
  without CHANGELOG entries, so the `[0.4.0]` section is authored whole (not moved
  from accumulated Unreleased notes). `[Unreleased]` stays as a present-but-empty
  heading (Group O5 asserts the `[Unreleased]:` ref present).
- `test-docs.sh` Group O asserts: `[Unreleased]:` and `[0.1.0]:` link refs
  present (O5), the ten verbs in backticks (O4 — already satisfied by the
  `[0.1.0]` section), the `## [0.1.0]` heading (O3), `keepachangelog.com` (O2).
  Keep all of these; add the `[0.4.0]:` ref and repoint `[Unreleased]:` to
  `v0.4.0...HEAD`.
- The plugin version pin is `0.3.1` today — bump it to `0.4.0` to match the tag
  (the pre-flight requires the pin == the tag).
- **Minor, not patch:** v0.4.0 adds six new user-facing surfaces, so the
  CHANGELOG is `### Added`-only (no `### Fixed`/`### Changed`) — every existing
  command behaves identically.

### Release runtime/operational pre-flight (all must be ticked at design)

Adopted from the release-cut spec template (`projects/_templates/spec-release-cut.md`,
per AGENTS.md §4). Concretized for the v0.4.0 `brag` cut. Items verifiable now
carry concrete evidence; items that can only be checked AT the tag/RC are marked
"verified by orchestrator at the cut" (honest — not faked).

- [x] Dual-tag-on-same-commit: RC tag + release deleted before the final tag is
      cut at the same commit (§4 Pattern 1). — **Documented** in the Destructive
      sequence above (verbatim §4 Pattern 1 commands: `gh release delete
      v0.4.0-rc1 --yes --cleanup-tag` then `git tag v0.4.0 && git push`).
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
      smoke gate above sets `BRAGFILE_DB=$(mktemp -d)/smoke.db`. For v0.4.0 the
      DEC-021 backup path is asserted **not** to fire (migration-free); *the
      throwaway-DB smoke is run by orchestrator at the cut.*
- [x] Clean upgrade: `brew upgrade` from the prior minor verified; `brag
      --version` prints the new tag; no migration surprise. — **Verified by
      orchestrator at the cut** (`brew upgrade jysf/bragfile/bragfile` from
      v0.3.1 → `brag --version` == `0.4.0`; migration-free per DEC-021 negative
      check). Design evidence that it *will* be clean: **confirmed at build** that
      no migration file was added under `internal/storage/migrations/` (four
      files `0001`..`0004`, unchanged since v0.3.0).
- [x] CHANGELOG: the `[0.4.0]` dated section is authored; compare-links
      repointed. — **Done in this PR** (the six surfaces under `### Added` + the
      `### Upgrading from v0.3.1` block; Group O stays green — verified at prep).
- [x] Plugin version pin (v0.3.0+): `plugin/.claude-plugin/plugin.json`
      `version` matches the tag. — **Done in this PR** (bumped `0.3.1` →
      `0.4.0`).
- [x] Behavioral surfaces re-checked on the built artifact (§12(b) refinement):
      `claude plugin details` shows the MCP server registered; the Stop hook
      fires in a throwaway repo. — **Verified by orchestrator at the cut** (needs
      the built plugin + a clean Claude install; not run here to avoid mutating
      the user's Claude config, per SPEC-047/042's precedent).

---

## Build Completion

*Filled in at the end of the **build** (prep) cycle, before advancing to verify.*

- **Branch:** `feat/spec-054-v0-4-0-release-cut`
- **PR (if applicable):** this PR — authors the CHANGELOG `[0.4.0]` + bumps the
  plugin version pin to `0.4.0` + updates the STAGE-013 backlog. The **tag cut
  itself is NOT in this PR** (it is the coordinator/human-gated ship step, run
  from `main` after this merges).
- **All acceptance criteria met?** **Prep ACs green; cut ACs deferred to the
  orchestrator's tag.** Verified at prep (2026-07-07):
  - `go test ./...` green; `gofmt -l .` empty; `go vet ./...` clean;
    `CGO_ENABLED=0 go build ./...` succeeds; `just test-docs` exits 0;
    `just test-hook` exits 0.
  - CHANGELOG `[0.4.0]` authored (date `2026-07-07`) + link-refs repointed;
    `test-docs.sh` Group O green.
  - `plugin/.claude-plugin/plugin.json` `version` == `0.4.0`.
  - Migration-free confirmed: `internal/storage/migrations/` holds four files
    (`0001`..`0004`), unchanged since v0.3.0.
  - **Deferred to the cut** (need the real tag / a plugin install / a throwaway
    DB): the RC-is-prerelease + no-tap-change gate, the RC binary smoke (the six
    surfaces + sparkline escape + `--previous` window + migration-free open), the
    final release's 4 archives + checksums, the tap bump to `0.4.0`, `brew
    upgrade → 0.4.0`, and the §12(b) `claude plugin details brag` registration +
    Stop-hook check on the built plugin.
- **New decisions emitted:** none.
- **Deviations from spec:** none. CHANGELOG dated `2026-07-07` (the intended cut
  date); if the cut slips, the orchestrator bumps the date to match.
- **Follow-up work identified:**
  - The **ship (cut)** step — coordinator/human-gated `git tag v0.4.0-rc1` →
    smoke → delete RC → `git tag v0.4.0` (§4 Pattern 1), then the tap bump +
    brew verify, then flip STAGE-013 frontmatter `status` → shipped and close
    PROJ-004. Closes STAGE-013 and the project.
  - The `ListFilter.Until` storage promotion SPEC-053 surfaced (a third
    `created_at < end` consumer) — backlog, not blocking v0.4.0.

### Build-phase reflection (3 questions, short answers)

1. **What was unclear in the spec that slowed you down?**
   — Nothing. The SPEC-047 precedent made the mechanical/irreversible split
   obvious, and the pre-flight was a checklist. The only real authoring work was
   an accurate, additive `### Added`-only CHANGELOG describing six surfaces — the
   DECs (028..033) gave clean per-feature framing.

2. **Was there a constraint or decision that should have been listed but wasn't?**
   — No. `one-spec-per-pr` (separate PR from the six feature PRs) and the DEC-021
   negative check (migration-free open) were both listed and both held; the
   migration-free property was confirmed at build (no new file under
   `internal/storage/migrations/`).

3. **If you did this task again, what would you do differently?**
   — Nothing structural. A minor cut is the same runbook as a patch with a
   longer, grouped `### Added` — the fourth consecutive use of the R2 pre-flight
   kept the prep mechanical.

---

## Reflection (Ship)

*Appended during the **ship** cycle. Outcome-focused, distinct from the
process-focused build reflection above. This spec's ship = the mechanical prep
merged to main; the irreversible tag/publish is the orchestrator's.*

1. **What would I do differently next time?**
   — Little. The v0.4.0 cut is the fourth consecutive use of the R2 release-cut
   pre-flight, and the first *minor* since SPEC-042 (v0.3.0). It kept the prep
   mechanical: every §4 gotcha (dual-tag, Gatekeeper, brew-trust, prod-DB) was a
   ticked item with either README evidence or an explicit "verified by
   orchestrator at the cut", so the reversible PR and the irreversible cut stayed
   cleanly separated. As with v0.3.1, this session hard-stops before the tag, so
   several pre-flight items are deferred-to-cut rather than ticked-with-evidence —
   the deferred set is enumerated explicitly (in Build Completion) rather than
   left implied, because the checklist only earns its keep if the orchestrator
   actually runs those items.

2. **Does any template, constraint, or decision need updating?**
   — No — but the observation SPEC-047 logged now has N=2: the `spec-release-cut.md`
   pre-flight has no built-in slot to distinguish "ticked with evidence now" from
   "deferred to the cut", and this spec again had to draw that line in prose. A
   template that split the two into explicit columns would make the honest split
   first-class. Worth a small template refinement at the next quiet moment; still
   not blocking.

3. **Is there a follow-up spec I should write now before I forget?**
   — No new spec for PROJ-004 — this cut closes the project. The live threads are
   PROJ-005 (team federation + token-economics reconciliation) and the
   `ListFilter.Until` storage promotion SPEC-053 surfaced; both are already
   recorded (the latter on the backlog) and neither is a STAGE-013 follow-up.
   STAGE-013 and PROJ-004 close with this cut.
