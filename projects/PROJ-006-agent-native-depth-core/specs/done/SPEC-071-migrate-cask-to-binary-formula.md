---
# Maps to ContextCore task.* semantic conventions.
# DRAFT — floated 2026-07-16 alongside DEC-040. PROJ-006 stages are unframed;
# `stage` is a placeholder. This is a distribution-shape change, so it MUST run
# docs/distribution-decisions.md before merge (it already has, via DEC-040).

task:
  id: SPEC-071
  type: story
  cycle: ship
  blocked: false
  priority: medium
  complexity: S

project:
  id: PROJ-006
  stage: STAGE-XXX                 # TBD — distribution/ops
repo:
  id: bragfile

agents:
  architect: claude-opus-4-8
  implementer: claude-opus-4-8
  created_at: 2026-07-16

references:
  decisions: []                    # emits/consumes DEC-040 (binary formula over cask)
  constraints: [one-spec-per-pr]
  related_specs: []                # SPEC-023 (original distribution), SPEC-037 (v0.2.0 cut)
---

# SPEC-071: Migrate distribution from Homebrew cask to a binary formula

## Context

DEC-040 decides bragfile ships as a **binary Homebrew formula**, not a cask, to
avoid macOS code-signing entirely (formulae are not Gatekeeper-quarantined;
casks are). This reverses the undocumented cask switch (commit `1582572`) that
introduced the whole signing problem. This spec is the mechanical migration.

The change reuses the pre-built, ldflags-stamped goreleaser release binary, so
there is **no** version-stamping code change and **no** compile cost.

## Goal

Replace the `homebrew_casks:` block in `.goreleaser.yaml` with a binary `brews:`
formula on the **shared `jysf/homebrew-tap`** (install `jysf/tap/bragfile`),
update every live install doc to drop the Gatekeeper `xattr` workaround and the
old tap/command, and verify a clean-host `brew install` runs with no Gatekeeper
prompt.

## Inputs

- **Files to read:** `.goreleaser.yaml` (the `homebrew_casks:` block);
  `README.md` §Install (the `xattr` note + brew-trust note); `AGENTS.md` §4
  (Gatekeeper + brew-trust addenda); `docs/macos-notarization-checklist.md`
- **Related code paths:** the tap repo `github.com/jysf/homebrew-tap`

## Outputs

- **Files modified:**
  - `.goreleaser.yaml` — `homebrew_casks:` → `brews:` on `homebrew-tap` (block below)
  - `README.md` — new install (`jysf/tap/bragfile`); remove the `xattr` step and
    the `brew trust` step (both obsolete for a formula — precedent: `crustyimg`
    on the same tap installs with neither)
  - `CHANGELOG.md` — `[Unreleased]` entry for the cask→formula + tap move
  - `docs/macos-notarization-checklist.md` — superseded-by-DEC-040 banner (kept
    for the record; only relevant if a signed cask channel is ever revived)
  - Live install references retargeted to `jysf/tap/bragfile` / `homebrew-tap`:
    `AGENTS.md` (§3 Distribution, §4 addenda, glossary), `.repo-context.yaml`,
    `.github/workflows/release.yml`, `plugin/README.md`, `BRAG.md`,
    `docs/architecture.md`, `docs/for-ai-agents.md`, `docs/tutorial.md`
  - **NOT touched (historical record):** past `CHANGELOG` version entries, the
    closed `PROJ-001` backlog, and prior DECs (e.g. DEC-025) — they describe what
    was true at the time
- **Database changes:** none

## Acceptance Criteria

- [ ] `goreleaser check` passes; the `brews:` deprecation warning is present and
      **accepted** (documented in the CHANGELOG/PR per DEC-040 — a nudge, not a wall).
- [ ] On a clean Mac: `brew install jysf/tap/bragfile`, then `brag --version` →
      **no Gatekeeper prompt, no trust step**, prints the tagged version
      (matching the `crustyimg` formula precedent on the same tap).
- [ ] README no longer instructs `xattr` or `brew trust`; it asserts a clean
      formula install.
- [ ] No `homebrew_casks:` remains in `.goreleaser.yaml`; the tap is `homebrew-tap`.

## Failing Tests

Release-cut-style doc/harness assertions (there is no Go surface here):
- **`scripts/test-docs.sh`** — assert `.goreleaser.yaml` declares `brews:` (not
  `homebrew_casks:`) on `homebrew-tap` (L6/L7/L11); README install literal is
  `brew install jysf/tap/bragfile`; README §Install does **not** contain
  `xattr -dr com.apple.quarantine`.

## Implementation Context

### The exact `brews:` block to restore (from pre-`1582572` config)

```yaml
brews:
  - name: bragfile
    repository:
      owner: jysf
      name: homebrew-tap
      branch: main
      token: "{{ .Env.HOMEBREW_TAP_GITHUB_TOKEN }}"
    homepage: "https://github.com/jysf/bragfile000"
    description: "Local-first Go CLI to capture and retrieve career-worthy moments."
    license: "MIT"
    skip_upload: auto
    test: |
      system "#{bin}/brag", "--version"
    install: |
      bin.install "brag"
```

(The cask switch dropped `license:`/`test:` — casks don't support them — and
added `binaries: [brag]`. Reverting restores all three.)

### Decisions that apply

- **DEC-040** — binary formula over cask; the full rationale + alternatives.

### Out of scope

- `go install` docs (a fine secondary path; separate change if wanted).
- Retiring the tap or the `brew trust` step (tap-trust is a separate gate).
- Any signing/notarization work (DEC-040 removes it from the path).

## Notes for the Implementer

- The `brews:` deprecation warning is expected and correct to accept — that is
  the whole point of DEC-040. Do not "fix" it by returning to a cask.
- Fallback (not this spec): if a future goreleaser removes `brews:`, hand-maintain
  a binary formula (`url` + `sha256` + `bin.install "brag"`) in the tap — still
  unquarantined, still no signing.
- **Tap-trust — resolved by precedent (DEC-040):** `crustyimg`, a binary formula
  on this same `jysf/homebrew-tap`, installs with no trust step and no Gatekeeper
  prompt. So a formula from this tap needs neither; docs assert a clean
  `brew install` with no `brew trust`. bragfile's own clean-host cut is final
  confirmation, but this is precedented, not open.

---

## Build Completion

*Filled at the end of the build cycle, before verify.*

- **Branch:** `distribution/cask-to-formula-shared-tap` → squash-merged as `a5da3e5`.
- **PR (if applicable):** #121 (merged).
- **All acceptance criteria met?** Phase 1 (config + docs) **yes** — `test-docs`
  green, `goreleaser check` validated. Phase 2 (clean-host install) **NO** — see
  Cut record.
- **New decisions emitted:** DEC-040 (committed).
- **Deviations from spec:** none in Phase 1.
- **Cut record (HONEST — the publish did NOT fully land):** v0.5.2 was tagged at
  `a5da3e5`; goreleaser built and published the **GitHub release + binary
  tarballs**, but the formula push to `jysf/homebrew-tap` **403'd** —
  `PUT .../homebrew-tap/contents/bragfile.rb: 403 Resource not accessible by
  personal access token`. So `Formula/bragfile.rb` never reached the tap and
  `brew install jysf/tap/bragfile` fails ("no formula"). **Root cause:** the
  `HOMEBREW_TAP_GITHUB_TOKEN` PAT has write to the old `homebrew-bragfile` tap,
  not the shared `homebrew-tap`. Secondary: goreleaser targeted the tap **root**,
  not `Formula/` (crustyimg's layout). **Fixes:** (1) grant the PAT
  Contents:read/write on `jysf/homebrew-tap` [owner action]; (2) `directory:
  Formula` added to the `brews:` block (branch `fix/tap-formula-publish`);
  (3) re-cut. **Status: RESOLVED & VERIFIED (2026-08-07).** After pointing the
  secret at a PAT with `push` on `jysf/homebrew-tap` and re-cutting, the release
  run went green — `Formula/bragfile.rb` (v0.5.2) landed on the tap beside
  `crustyimg.rb`, and `brew install jysf/tap/bragfile` installs clean on a real
  host: no Gatekeeper "Apple could not verify" prompt, no `brew trust` step,
  `brag --version` = 0.5.2. It took **two** failed cuts (the 403 recurred
  because the token wasn't fixed before re-cutting) — the N=2 that promotes
  "check the tap token" from a pre-flight note to a hard gate.

---

## Reflection (Ship)

*Appended during the **ship** cycle. Outcome-focused reflection, distinct
from the process-focused build reflection above.*

1. **What would I do differently next time?**
   — Make the pre-flight token check a **hard gate**, not a note. SPEC-071's own
   pre-flight named it ("confirm HOMEBREW_TAP_GITHUB_TOKEN is valid — the one
   thing that silently breaks the tap push") and it still wasn't checked before
   the cut, so a tap switch shipped a published-but-broken v0.5.2. A tap/channel
   change must verify token write-access to the *new* tap before tagging.

2. **Does any template, constraint, or decision need updating?**
   — Yes: the release pre-flight should assert token write-access **to the
   specific tap being pushed to** (not just "token exists"), especially when the
   tap changes. Feed this to the framework harvest (distribution lessons).

3. **Is there a follow-up spec I should write now before I forget?**
   — No new spec; the fix is `fix/tap-formula-publish` (directory: Formula) plus
   the owner's PAT-scope fix plus a re-cut. Tracked in the Cut record above.

4. **What can a user do now that they couldn't before?**
   — Install bragfile with a single clean `brew install jysf/tap/bragfile` —
   **no macOS "Apple could not verify" prompt, no `xattr` quarantine step, and no
   `brew trust`** — where before (the v0.1.0 cask) first run was blocked by
   Gatekeeper until the user ran a `sudo xattr` command. Verified end-to-end on a
   real host at v0.5.2.
