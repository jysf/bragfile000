---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-037
  type: chore                      # release mechanics, not feature code
  cycle: build                     # frame | design | build | verify | ship
  blocked: false
  priority: high                   # release-cutting spec; closes STAGE-008 + gates PROJ-002 close
  complexity: S                    # runbook + pre-flight + verification checklist; no code

project:
  id: PROJ-002
  stage: STAGE-008
repo:
  id: bragfile

agents:
  architect: claude-opus-4-8
  implementer: claude-opus-4-8     # usually same Claude, different session
  created_at: 2026-06-12

references:
  decisions: [DEC-001, DEC-021]    # cited of record (no NEW DEC); see Implementation Context
  constraints: [no-cgo, no-secrets-in-code, migrations-are-append-only, one-spec-per-pr]
  related_specs: [SPEC-023, SPEC-024, SPEC-034, SPEC-035, SPEC-036]
---

# SPEC-037: v0.2.0 release cut

## Context

This is the **final spec of STAGE-008 and of PROJ-002**. STAGE-006 (tags) and
STAGE-007 (projects) completed the v0.2.0 feature surface; SPEC-034 swept the
docs, SPEC-035 wrote the CHANGELOG `[0.2.0]`, and SPEC-036 shipped the
migration auto-backup safety belt (DEC-021). Nothing is left to build. This
spec turns "feature-complete on `main`" into "released and safe to upgrade
into" by cutting `v0.2.0-rc1 → v0.2.0` through the existing goreleaser →
GitHub Releases → Homebrew-tap pipeline, per **AGENTS.md §4 Pattern 1**
(delete the RC tag + release, then cut the final tag at the same commit — the
dual-tag-on-same-commit gotcha avoidance earned at the v0.1.0 cut).

It also reconciles the dev/prod DB story this session disrupted (STAGE-008
Design Note #3): the user's production `~/.bragfile` is already on the v0.2.x
schema, and bare `brag` currently resolves to a dev-built v0.2.x binary at
`~/go/bin/brag`. The release is the moment to retire that dev binary and
return to the brew-installed `v0.2.0` (the standing "switch back" intent).

> **This spec is a RUNBOOK, not code.** It produces a precise, verifiable
> sequence of commands with explicit gates. **The destructive steps — pushing
> tags, deleting the RC release, bumping the Homebrew tap, running
> `brew upgrade`, and uninstalling the dev binary — are executed by the
> COORDINATOR with per-step human go-ahead, NOT by any build agent.** The
> build cycle is a **REHEARSAL** (see "The build cycle is a rehearsal" below):
> `goreleaser check` + `goreleaser build --snapshot` + a dry-run of the smoke
> sequence on **throwaway tags/DBs only**. No tag is pushed and
> `goreleaser release` is never run during build.

## Goal

Cut and verify the public `v0.2.0` release of `brag` through the existing
pipeline, following §4 Pattern 1, and reconcile the local dev/prod state back
to the released brew binary — gated so each step is human-approved and each
outcome is independently checkable.

## Inputs

- **Files to read:**
  - `AGENTS.md` §4 (release ops, in full — the goreleaser commands, the
    dual-tag-on-same-commit recovery, the macOS Gatekeeper xattr note),
    §2/§9/§12/§13.
  - `.goreleaser.yaml` — the pipeline being driven (ldflags version wiring,
    `homebrew_casks.skip_upload: auto`, `release.prerelease: auto`).
  - `.github/workflows/release.yml` — tag-push (`v*`) triggered; runs
    `goreleaser release --clean` with `HOMEBREW_TAP_GITHUB_TOKEN`.
  - `CHANGELOG.md` — the `[0.2.0]` section (dated 2026-06-12) and compare links.
  - `cmd/brag/main.go` + `internal/cli/root.go` — the `main.version` ldflags
    target and how `brag --version` renders it.
  - `projects/PROJ-002-projects-and-tags/stages/STAGE-008-...md` — Success
    Criteria (the release criteria) and Design Notes (esp. #3, dev/prod story).
- **Related code paths:** `internal/storage/migrate.go` (DEC-021 safety belt,
  for the no-migration-on-prod verification); `justfile` (`install` /
  `uninstall` targets, the switch-back lever).
- **External (coordinator-side, not in repo):** the GitHub repo
  `jysf/bragfile000` (releases, tags, Actions, secrets) and the Homebrew tap
  `jysf/homebrew-bragfile`.

## Outputs

This spec ships **no Go code and no migration**. Its repo-level outputs are
the small release-prep status reconciliations that must land on the release
SHA *before* tagging (see Pre-flight P5/P6), plus the runbook record itself:

- **Files modified (release-prep, on the release SHA, by the coordinator):**
  - `README.md` — status line `:10` flips `v0.1.0 shipped` → `v0.2.0`, and the
    feature blurb + the `:36` "v0.1.0 ships unsigned" note generalize to
    v0.2.0 (status-change premise audit, below). This is a release-gated
    status claim; SPEC-034's sweep deliberately left it (flagged README as an
    out-of-scope follow-up), so it lands here.
  - `CHANGELOG.md` — **only if** the actual cut date differs from 2026-06-12:
    bump the `[0.2.0]` date to the real cut date (SPEC-035 assigned the
    date-bump ownership to SPEC-037).
- **No new exports. No database changes. No new migration**
  (`schema_migrations` stays at 4; DEC-021 already shipped in SPEC-036).
- **New `DEC-*`:** none. DEC-001 (no-cgo) and DEC-021 (safety belt) are cited
  of record only.

## Acceptance Criteria

Each gate is independently checkable by a verifier. The destructive steps are
coordinator-executed; the ACs below describe the *observable end state* each
produces.

- [ ] **Pre-flight all green** (P1–P7 below): `main` clean and at the intended
      release SHA, `local == origin`; `CGO_ENABLED=0 go test ./...` green;
      `gofmt -l .` empty; `go vet ./...` clean; `bash scripts/test-docs.sh`
      exits 0; `goreleaser check` passes; `goreleaser build --snapshot --clean`
      succeeds and the built binary reports a goreleaser-injected version (NOT
      `dev`), proving the `-X main.version` ldflags wiring is live.
- [ ] **Release-prep status reconciled on the release SHA:** README no longer
      claims `v0.1.0 shipped`; CHANGELOG `[0.2.0]` date == actual cut date.
- [ ] **rc is a GitHub PRERELEASE with no tap change:** after `v0.2.0-rc1` CI
      completes, the GitHub release for `v0.2.0-rc1` is marked *Pre-release*,
      and the `jysf/homebrew-bragfile` tap's latest commit is **unchanged**
      (skip_upload:auto held).
- [ ] **RC smoke gate passes** on a THROWAWAY DB (never `~/.bragfile`): the rc
      binary reports `0.2.0-rc1`; `project new/here/status`, `add` cwd
      auto-fill, and a tag op all behave; and the DEC-021 safety-belt path
      fires (a timestamped sidecar appears) when the rc binary opens a seeded
      v0.1.x-schema throwaway DB — which simultaneously verifies the clean
      v0.1.x → v0.2.0 DB upgrade (resolving Design Note #3).
- [ ] **Final release has all 4 platform archives + checksums:** the
      `v0.2.0` GitHub release (NOT a prerelease) carries
      `bragfile_0.2.0_darwin_amd64.tar.gz`, `..._darwin_arm64.tar.gz`,
      `..._linux_amd64.tar.gz`, `..._linux_arm64.tar.gz`, and `checksums.txt`.
- [ ] **Tap bumped to 0.2.0:** `jysf/homebrew-bragfile` has a new commit
      updating the `bragfile` cask to version `0.2.0` with matching sha256s.
- [ ] **Brew binary reports 0.2.0:** after `brew update && brew upgrade
      bragfile` (+ the Gatekeeper xattr clear), `brag --version` from the brew
      install path reports `0.2.0`.
- [ ] **Dev binary retired:** after `just uninstall`, `which brag` resolves to
      the brew path (not `~/go/bin/brag`) and bare `brag --version` → `0.2.0`.
- [ ] **Prod DB opens with no migration:** opening the real `~/.bragfile`
      (already v0.2.x, applied==4, pending==0) with the released binary fires
      no migration and writes no new backup sidecar (DEC-021 trigger
      `applied>0 && pending>0` is false).
- [ ] **No regressions:** the full Go suite, `gofmt`, `go vet`, and the
      `CGO_ENABLED=0` build stay green (verified at pre-flight; this spec adds
      no code that could break them).

## Failing Tests

This spec ships **no Go code**, so there are no new Go failing tests — the
`test-before-implementation` constraint is satisfied the way SPEC-034/035
(docs/CHANGELOG, also code-free) satisfied it: the enforcement mechanism is
the **pre-flight gate suite** (the existing green `go test ./...` — including
SPEC-036's 5 safety-belt tests that already prove the DEC-021 path —
`scripts/test-docs.sh`, and `goreleaser check`) plus the **post-release
verification checklist**, not new `*_test.go` files. Each AC above maps to a
concrete command in the runbook sections that follow; "the test" for a release
runbook is the executable gate, not an assertion in Go.

The regression guard is the pre-existing suite, run unchanged at Pre-flight P2.

---

## The build cycle is a REHEARSAL (read before building)

The build cycle for SPEC-037 does **not** cut the release. It rehearses the
runbook against throwaway targets so the real cut (coordinator + human
go-aheads) is mechanical:

1. **`goreleaser check`** — validate `.goreleaser.yaml` against the installed
   goreleaser v2 (design-time pre-flight per §12(b): run the embedded literal
   through its tool).
2. **`goreleaser build --snapshot --clean`** — local cross-compile of all four
   targets; confirm a host-arch binary reports a goreleaser-injected version
   (NOT `dev`).
3. **Smoke-sequence dry-run on throwaway tags/DBs** — run the RC SMOKE GATE
   commands (below) substituting the **snapshot-built binary** for the
   downloaded rc binary, and a `/tmp` throwaway DB for `~/.bragfile`. Seed a
   v0.1.x throwaway DB and confirm the safety belt fires + the upgrade is
   clean. **No tag is pushed; `goreleaser release` is never invoked; no `gh
   release` call is made.**
4. **A written walkthrough** of the destructive sequence (§4 Pattern 1
   commands, verbatim), left in the spec for the coordinator to execute
   step-by-step.

If any rehearsal step fails, fix the runbook (or flag a blocker) before the
spec advances — never the real cut.

---

## Pre-flight checklist (ALL must pass before any tag is pushed)

Run from a clean checkout of the intended release SHA on `main`.

```bash
# P0. Capture the intended release SHA (everything below tags THIS commit).
git rev-parse HEAD
git rev-parse --abbrev-ref HEAD            # → main

# P1. main clean, at the release SHA, local == origin.
git status --porcelain                     # → empty (see §13 note: ignore the
                                           #   pre-existing process-feedback.md
                                           #   change from a parallel session)
git fetch origin
git rev-parse HEAD origin/main             # → identical SHAs

# P2. Go suite green; format + vet clean; doc assertions green.
CGO_ENABLED=0 go test ./...                # → ok (incl. SPEC-036 safety-belt tests)
test -z "$(gofmt -l .)" && echo "gofmt clean"
go vet ./...
bash scripts/test-docs.sh                  # → exits 0

# P3. goreleaser config valid (design-time pre-flight, §12(b)).
goreleaser check                           # → "config is valid"

# P4. Local cross-compile smoke + ldflags version wiring proof.
goreleaser build --snapshot --clean
ls dist/*/brag
# Run the host-arch binary; expect a goreleaser version (e.g. 0.2.x-SNAPSHOT-…),
# NOT "dev" — this proves -X main.version reaches `brag --version`.
"dist/brag_$(go env GOOS)_$(go env GOARCH)"*/brag --version
```

```bash
# P5. CHANGELOG DATE RECONCILIATION (SPEC-035 handed date ownership here).
#     [0.2.0] is dated 2026-06-12. If the cut happens on a LATER day, bump it.
grep -n '## \[0.2.0\]' CHANGELOG.md        # check the date on this line
date +%F                                   # actual cut date
#  If they differ: edit CHANGELOG.md line "## [0.2.0] - <date>" to today's date,
#  commit on a chore branch, open a one-spec PR (SPEC-037), merge to main, and
#  RE-CAPTURE the release SHA (P0) — the tag must point at the corrected commit.

# P6. README status string flip (status-change premise audit).
grep -n -E 'v0\.1\.0 shipped|v0\.1\.0 ships unsigned' README.md
#  Edit README.md:10 status line  v0.1.0 → v0.2.0 (and the feature blurb), and
#  generalize the :36 "v0.1.0 ships unsigned" note to v0.2.0. Same chore PR /
#  merge / re-capture-SHA flow as P5. (One PR may carry both P5 + P6 under
#  SPEC-037 — one-spec-per-pr is satisfied; they are the same release-prep.)
```

```bash
# P7. Release secret present in CI  — COORDINATOR-ONLY (cannot verify from repo).
gh secret list --repo jysf/bragfile000     # → HOMEBREW_TAP_GITHUB_TOKEN present
#  (GITHUB_TOKEN is provided automatically; release.yml already grants
#   `permissions: contents: write`, which IS verifiable from the repo.)
```

After P5/P6 edits land and the SHA is re-captured, **re-run P1–P4** so the
tag is cut on a green, reconciled commit.

---

## The rc → release sequence (§4 Pattern 1) — COORDINATOR-EXECUTED, per-step go-ahead

### Step 1 — cut the RC (produces a GitHub PRERELEASE, no tap update)

```bash
git tag v0.2.0-rc1                         # at the reconciled release SHA
git push origin v0.2.0-rc1                 # triggers release.yml → goreleaser
```

CI runs `goreleaser release --clean`. Because `release.prerelease: auto`, the
`-rc1` suffix marks the GitHub release **Pre-release**; because
`homebrew_casks.skip_upload: auto`, the cask push is **skipped** on the
prerelease. **This is the rc→release gate built into the config** — confirmed
against the goreleaser v2 semantics: `prerelease: auto` keys off the
prerelease segment of the semver tag, and `skip_upload: "auto"` skips
formula/cask publishing whenever the version is a prerelease or snapshot. The
final `v0.2.0` tag (no suffix) is what flips both: full release + tap bump.

Verify the gate held:

```bash
gh run watch --repo jysf/bragfile000                       # wait for green
gh release view v0.2.0-rc1 --repo jysf/bragfile000 --json isPrerelease,assets
#   → isPrerelease: true ; assets include the 4 tar.gz + checksums.txt
# Tap MUST be unchanged (skip_upload held):
gh api repos/jysf/homebrew-bragfile/commits --jq '.[0].commit.message'
#   → still the v0.1.x-era commit; NO 0.2.0 bump yet
```

### Step 2 — RC SMOKE GATE (throwaway DB only; NEVER `~/.bragfile`)

```bash
# Download the rc archive for THIS platform and extract.
mkdir -p /tmp/brag-rc && \
gh release download v0.2.0-rc1 --repo jysf/bragfile000 \
  --pattern "*_$(go env GOOS)_$(go env GOARCH).tar.gz" --dir /tmp/brag-rc
tar -xzf /tmp/brag-rc/*.tar.gz -C /tmp/brag-rc
/tmp/brag-rc/brag --version                # → 0.2.0-rc1

# Exercise the v0.2.0 surface against a throwaway DB (DEC-003: BRAGFILE_DB wins
# over the ~/.bragfile default; the smoke NEVER touches prod).
export BRAGFILE_DB=/tmp/brag-smoke/db.sqlite
mkdir -p /tmp/brag-smoke /tmp/brag-smoke-proj
/tmp/brag-rc/brag project new smoke --path /tmp/brag-smoke-proj
( cd /tmp/brag-smoke-proj && /tmp/brag-rc/brag project here )   # → resolves "smoke"
/tmp/brag-rc/brag project status
( cd /tmp/brag-smoke-proj && /tmp/brag-rc/brag add --title "smoke entry" --tag demo )
/tmp/brag-rc/brag list --show-project       # → entry auto-filled project=smoke
/tmp/brag-rc/brag tags                       # → "demo" with count 1
unset BRAGFILE_DB
```

```bash
# CONFIRM THE DEC-021 SAFETY-BELT PATH (and the clean v0.1.x→v0.2.0 upgrade in
# one shot). Build a real v0.1.0 binary in a throwaway worktree, seed a
# v0.1.x-schema DB, then open it with the rc binary.
git worktree add /tmp/brag-v010 v0.1.0
( cd /tmp/brag-v010 && CGO_ENABLED=0 go build -o /tmp/brag010 ./cmd/brag )
export BRAGFILE_DB=/tmp/brag-upgrade/db.sqlite
mkdir -p /tmp/brag-upgrade
/tmp/brag010 add --title "v0.1.x seed"      # creates a v0.1.x-schema DB
/tmp/brag-rc/brag list                       # opening with rc → migrates 0003/0004
ls /tmp/brag-upgrade/*.backup                # → a timestamped sidecar EXISTS
                                             #   (applied>0 && pending>0 fired)
/tmp/brag-rc/brag list                       # → the seed entry survived the upgrade
unset BRAGFILE_DB
git worktree remove /tmp/brag-v010 && rm -f /tmp/brag010
```

**Gate:** if any smoke step misbehaves, STOP — do not cut the final tag; fix
forward and re-cut a fresh RC (`v0.2.0-rc2`).

### Step 3 — cut the final tag (§4 Pattern 1, verbatim)

Only after the RC smoke gate passes and the human gives the go-ahead:

```bash
gh release delete v0.2.0-rc1 --yes --cleanup-tag    # removes the GH release + remote rc tag
git tag -d v0.2.0-rc1                                # remove the stale LOCAL rc tag too,
                                                     #   so a later local goreleaser run
                                                     #   can't latch onto it
git tag v0.2.0                                        # at the SAME commit as the rc
git push origin v0.2.0                                # CI → full release + tap bump
```

This is the dual-tag-on-same-commit avoidance: with the rc release+tag gone,
goreleaser sees `v0.2.0` as the only tag on the commit and builds it as a full
release. (The alternative — leaving both tags on one commit — is exactly the
`422 already_exists` failure §4 documents.)

---

## Post-release verification (COORDINATOR-EXECUTED)

```bash
# V1. Full release present with all 4 archives + checksums (NOT a prerelease).
gh run watch --repo jysf/bragfile000
gh release view v0.2.0 --repo jysf/bragfile000 --json isPrerelease,assets \
  --jq '{prerelease: .isPrerelease, files: [.assets[].name]}'
#   → prerelease:false ; files include all four:
#     bragfile_0.2.0_darwin_amd64.tar.gz, _darwin_arm64.tar.gz,
#     _linux_amd64.tar.gz, _linux_arm64.tar.gz, checksums.txt

# V2. Homebrew tap bumped to 0.2.0 (skip_upload no longer holds on a stable tag).
gh api repos/jysf/homebrew-bragfile/commits --jq '.[0].commit.message'   # → 0.2.0 bump
gh api repos/jysf/homebrew-bragfile/contents/Casks/bragfile.rb \
  --jq '.content' | base64 -d | grep -E 'version|sha256'                 # → 0.2.0 + sums

# V3. Brew install reports 0.2.0 (+ Gatekeeper clear per §4).
brew update
brew upgrade bragfile        # or: brew reinstall bragfile
sudo xattr -dr com.apple.quarantine /opt/homebrew/Caskroom/bragfile/     # §4 macOS note
"$(brew --prefix)/bin/brag" --version        # → 0.2.0  (the brew-installed binary)
```

---

## Dev/prod reconciliation — the "switch back" (COORDINATOR-EXECUTED)

Standing intent (recalled): return bare `brag` to the brew-installed `v0.2.0`
and confirm the already-v0.2.x prod DB needs no migration. Do this **after**
V3 so a working brew binary exists before the dev one is removed.

```bash
# R1. Retire the dev binary so bare `brag` resolves to brew.
just uninstall                               # rm ~/go/bin/brag (justfile target)
hash -r                                      # drop the shell's cached path to brag
which brag                                   # → $(brew --prefix)/bin/brag, NOT ~/go/bin/brag
brag --version                               # → 0.2.0

# R2. Prod ~/.bragfile opens with NO migration (it is already v0.2.x: applied==4).
before=$(ls -1 ~/.bragfile/*.backup 2>/dev/null | wc -l)
brag list                                    # released v0.2.0 against the REAL prod DB
after=$(ls -1 ~/.bragfile/*.backup 2>/dev/null | wc -l)
test "$before" -eq "$after" && echo "no new sidecar → no migration fired"
#   DEC-021 trigger is `applied>0 && pending>0`; with pending==0 it stays silent.
```

**Safety backup sidecar from this session**
(`~/.bragfile/db.sqlite.v0.2.x-schema.*.backup`): **recommendation — keep it
through the release and remove it after R2 passes** (it is a one-time,
harmless insurance artifact; removal is optional and can also be left
indefinitely). It is unrelated to the DEC-021 auto-backups and need not be
pruned by the release.

---

## Premise audit (§9)

- **No code change; no hardcoded version constant.** `brag --version` is wired
  entirely through ldflags: `cmd/brag/main.go:14-18` defaults `version="dev"`,
  `internal/cli/root.go:15` sets cobra `Version: version`, `.goreleaser.yaml:26`
  injects `-X main.version={{ .Version }}`. Grep for `0.1.0` / `v0.1` / `0.2.0`
  across `*.go`, `go.mod`, `*.yaml`, `justfile` (excluding tests/CHANGELOG)
  returns **only two doc/example strings** — `internal/cli/project.go:383` and
  `guidance/questions.yaml:154`, both free-text `--state-note "…cut v0.2.0"`
  examples that do NOT reach the binary's version surface. **Nothing in source
  needs a version edit.**
- **No migration, no schema change.** SPEC-036 already shipped DEC-021;
  `schema_migrations` stays at 4. This spec adds no `0005_*`
  (`migrations-are-append-only` is not exercised).
- **Status-change (the only repo edits this spec gates):** `README.md:10`
  (`v0.1.0 shipped`) and `:36` (`v0.1.0 ships unsigned`) are release-gated
  status claims SPEC-034 deliberately left as a follow-up; they flip to v0.2.0
  on the release SHA (Pre-flight P6). `CHANGELOG.md` `[0.2.0]` date flips only
  if the cut slips past 2026-06-12 (P5; SPEC-035 assigned this here). Both are
  reconciliations, not new decisions.

## Implementation Context

*Read this section (and the files it points to) before starting the build
(REHEARSAL) cycle.*

### Decisions that apply

- `DEC-001` — pure-Go SQLite (`modernc.org/sqlite`), no CGO. This is why
  `.goreleaser.yaml` sets `CGO_ENABLED=0` and can cross-compile all four
  targets without a C toolchain; the snapshot build (P4) exercises exactly
  this. Cited of record; not re-decided here.
- `DEC-021` — migration auto-backup durability model (SPEC-036): trigger
  `applied>0 && pending>0`, snapshot via `VACUUM INTO`, abort Open on backup
  failure. Drives both the RC-smoke safety-belt check (fires on the seeded
  v0.1.x DB) and the prod no-migration check (silent at applied==4/pending==0).

### Constraints that apply

These constraints apply to the paths touched by this task (see
`/guidance/constraints.yaml` for full text):

- `no-cgo` (blocking) — the release pipeline depends on it; P3/P4 are the
  proof it still holds at release time.
- `no-secrets-in-code` (blocking) — `HOMEBREW_TAP_GITHUB_TOKEN` lives only in
  CI secrets, referenced via `{{ .Env.HOMEBREW_TAP_GITHUB_TOKEN }}`; the
  runbook never echoes or embeds it (P7 lists it, does not print its value).
- `migrations-are-append-only` (blocking) — confirmed not exercised (no new
  migration).
- `one-spec-per-pr` (blocking) — the P5/P6 release-prep edits go in a single
  PR referencing SPEC-037.

### Prior related work

- `SPEC-023` (shipped) — set up `.goreleaser.yaml` (`homebrew_casks:` /
  `formats:` keys) and earned the §12(b) "run the literal through
  `goreleaser check` at design" lesson. P3 honors it.
- `SPEC-024` (shipped) — the v0.1.0 cut that earned §4's dual-tag-on-same-commit
  recovery and the macOS Gatekeeper xattr note. This spec is the second
  application of that runbook.
- `SPEC-034` / `SPEC-035` / `SPEC-036` (shipped) — the doc sweep, the
  CHANGELOG `[0.2.0]`, and the DEC-021 safety belt this release ships.
- **Trust-but-verify agent push reports** (STAGE-008 WATCH, from SPEC-023):
  every "pushed the tag / bumped the formula" claim in this runbook is checked
  via `gh release view` / `gh api .../commits` / `git ls-remote`, never taken
  on faith.

### Out of scope (for this spec specifically)

- **macOS code signing + notarization.** Backlogged
  (`docs/macos-notarization-checklist.md`); the xattr workaround in V3 stays
  the v0.2.0 story. Trigger conditions unchanged.
- **The PROJ-002 close itself** (Prompt 1e) — runs *after* this spec ships:
  the project reflection, the WATCH-list final disposition, and the brief
  progress-marker reconciliation are project-close work, not release work.
- **Any new feature surface, automated daily backup (launchd), or
  `brag project tag`** — all out per STAGE-008 Scope.

### Coordinator-only flags (cannot be verified from inside the repo)

Surface these to the human before executing the destructive sequence:

1. **`HOMEBREW_TAP_GITHUB_TOKEN` present and valid** in `jysf/bragfile000`
   Actions secrets, with **push rights to `jysf/homebrew-bragfile`** (P7;
   `gh secret list` shows presence, not validity — a bad/expired token fails
   only at the `v0.2.0` tap-push step in CI).
2. **Branch/tag push permissions** — who may push `v*` tags to `origin`; main
   branch protection must still allow the SPEC-037 release-prep PR to merge.
3. **Tap default branch is `main`** (`.goreleaser.yaml` pins
   `branch: main`) — confirm the tap repo hasn't been renamed/retargeted since
   v0.1.0.

   *(Verifiable from the repo and already satisfied: `release.yml` grants
   `permissions: contents: write`, so the default `GITHUB_TOKEN` can publish
   the release.)*

## Notes for the Implementer

- **The build cycle never cuts the release** — see "The build cycle is a
  REHEARSAL." Your deliverable is a validated runbook (goreleaser check passes,
  snapshot build version-wired, smoke sequence proven on throwaway
  tags/DBs/worktrees), plus the spec's destructive-sequence commands left
  verbatim for the coordinator. Push no tag; run no `goreleaser release`; call
  no `gh release`.
- **Throwaway DB discipline (§ the incident that motivated SPEC-036):** every
  smoke command sets `BRAGFILE_DB` to a `/tmp` path (DEC-003 makes the env var
  win over the `~/.bragfile` default). Never run a smoke against `~/.bragfile`.
  Use `git worktree` for the v0.1.0 build so your checkout is untouched.
- **goreleaser snapshot version string** is not `0.2.0` — it is a snapshot
  template (e.g. `0.2.x-SNAPSHOT-<shortsha>`). The P4 assertion is "**not
  `dev`**", proving ldflags injection; do not assert the exact snapshot value.
- **Re-capture the SHA after P5/P6.** If README/CHANGELOG edits land, the tag
  must point at the *post-edit* commit; re-run P1–P4 before tagging.
- **Do not extract a `docs/RELEASING.md`.** AGENTS.md §4 is already the durable,
  reusable home for release ops (the dual-tag recovery + Gatekeeper note live
  there and were proven across two cuts); this spec is the v0.2.0 *record*.
  Adding a parallel RELEASING.md on the last spec of the project would
  duplicate §4 and risk drift. Recommendation: if a *third* release reuses this
  runbook verbatim, extract then — not now.

## §13 working-tree note

A parallel session has an uncommitted change to
`docs/framework-feedback/process-feedback.md`. **Do not commit, revert, or
stage it.** The Pre-flight P1 `git status --porcelain` will show it; treat that
single path as expected noise, not a dirty-tree blocker for the release.

---

## Build Completion

*Filled in at the end of the **build (REHEARSAL)** cycle, before advancing to
verify.*

- **Branch:** none — rehearsal produces no committable file changes
- **PR (if applicable):** none (no branch opened; no code changed; the cycle-advance
  commit below is the only repo change)
- **All acceptance criteria met?** YES (rehearsal scope): P1–P4 all green;
  snapshot version-wired (`0.1.0-SNAPSHOT-c0fd45c`, not `dev`); v0.2.0 surface
  smoke passed on throwaway DB; DEC-021 safety belt fired on v0.1.x-seeded DB
  (backup sidecar `db.sqlite.pre-0004_add_projects.20260613T061752Z.backup`
  appeared; seed entry survived). Cut/post-release/switch-back ACs are
  coordinator-time.
- **New decisions emitted:** none
- **Deviations from spec:**
  - Runbook Step 2 (RC SMOKE GATE) uses `--tag demo` for `brag add`; the actual
    flag is `--tags` (plural, per `brag add --help`). Corrected during rehearsal.
    The runbook text should read `--tags demo` before the real cut.
  - Snapshot version string is `0.1.0-SNAPSHOT-c0fd45c` (not `0.2.x-SNAPSHOT-…`)
    because `v0.2.0` has not been tagged yet. Expected per spec ("P4 assertion is
    'not dev'"); no fix needed.
  - `brag project here` returns `smoke	active	-` (three-column output including a
    `-` path column), not just `smoke`. Minor cosmetic difference from the runbook's
    prose "→ resolves 'smoke'"; behavior is correct.
- **Follow-up work identified:**
  - Fix the runbook: `--tag demo` → `--tags demo` in the RC SMOKE GATE block
    (coordinator edits before the real cut, or a quick chore commit on main).

### Build-phase reflection (3 questions, short answers)

Process-focused: how did the rehearsal go? What friction did the runbook create?

1. **What was unclear in the spec that slowed you down?**
   — The runbook used `--tag` (singular) in the smoke step; the CLI flag is
   `--tags` (plural). One failed run was needed to discover this. The spec
   is otherwise extremely precise — every command ran correctly once the flag
   was corrected.

2. **Was there a constraint or decision that should have been listed but wasn't?**
   — None. DEC-001 (no-cgo) and DEC-021 (safety belt) are the right citations.
   The `one-spec-per-pr` constraint correctly applies to the P5/P6 release-prep
   PR. All four constraints listed are exercised.

3. **If you did this task again, what would you do differently?**
   — Run `brag add --help` at the start of the smoke step to confirm exact flag
   names before transcribing runbook commands. Alternatively, the spec could note
   that `-T` / `--tags` is the tag flag for `brag add` in the smoke step to
   prevent the same trip for the real cut.

---

## Reflection (Ship)

*Appended during the **ship** cycle. Outcome-focused reflection, distinct
from the process-focused build reflection above.*

1. **What would I do differently next time?**
   — <answer>

2. **Does any template, constraint, or decision need updating?**
   — <answer>

3. **Is there a follow-up spec I should write now before I forget?**
   — <answer>
