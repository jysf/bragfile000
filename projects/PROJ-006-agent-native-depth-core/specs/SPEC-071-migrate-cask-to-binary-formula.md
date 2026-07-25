---
# Maps to ContextCore task.* semantic conventions.
# DRAFT — floated 2026-07-16 alongside DEC-040. PROJ-006 stages are unframed;
# `stage` is a placeholder. This is a distribution-shape change, so it MUST run
# docs/distribution-decisions.md before merge (it already has, via DEC-040).

task:
  id: SPEC-071
  type: story
  cycle: design
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
formula, update the docs to drop the Gatekeeper `xattr` workaround, and verify a
clean-host `brew install` runs with no Gatekeeper prompt.

## Inputs

- **Files to read:** `.goreleaser.yaml` (the `homebrew_casks:` block);
  `README.md` §Install (the `xattr` note + brew-trust note); `AGENTS.md` §4
  (Gatekeeper + brew-trust addenda); `docs/macos-notarization-checklist.md`
- **Related code paths:** the tap repo `github.com/jysf/homebrew-bragfile`

## Outputs

- **Files modified:**
  - `.goreleaser.yaml` — `homebrew_casks:` → `brews:` (block below)
  - `README.md` — remove the `xattr -dr com.apple.quarantine` install step
    (obsolete: formulae aren't quarantined); **keep** the one-time
    `brew trust` step (tap-trust gate is unchanged)
  - `CHANGELOG.md` — note the distribution change under the next version
  - `docs/macos-notarization-checklist.md` — mark superseded-by-DEC-040 (kept for
    the record; only relevant if a signed cask channel is ever revived)
- **Database changes:** none

## Acceptance Criteria

- [ ] `goreleaser check` passes; the `brews:` deprecation warning is present and
      **accepted** (documented in the CHANGELOG/PR per DEC-040 — a nudge, not a wall).
- [ ] On a clean Mac: `brew trust --cask jysf/bragfile/bragfile` (unchanged),
      then `brew install jysf/bragfile/bragfile`, then `brag --version` → **no
      Gatekeeper prompt**, prints the tagged version.
- [ ] README no longer instructs `xattr`; still documents `brew trust`.
- [ ] No `homebrew_casks:` remains in `.goreleaser.yaml`.

## Failing Tests

Release-cut-style doc/harness assertions (there is no Go surface here):
- **`scripts/test-docs.sh`** — assert README §Install contains `brew trust` and
  does **not** contain `xattr -dr com.apple.quarantine`.

## Implementation Context

### The exact `brews:` block to restore (from pre-`1582572` config)

```yaml
brews:
  - name: bragfile
    repository:
      owner: jysf
      name: homebrew-bragfile
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

---

## Build Completion

*Filled at the end of the build cycle, before verify.*

- **Branch:**
- **PR (if applicable):**
- **All acceptance criteria met?** yes/no
- **New decisions emitted:** DEC-040 (if not already committed)
- **Deviations from spec:**

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

4. **What can a user do now that they couldn't before?** — one sentence,
   before → after; quote the confirming number if one exists, name the outcome
   if not. Write `none` if this spec has no user-visible outcome — that is a
   real, greppable result, not a blank. This is the line a brag's `impact` field
   is transcribed from, and both halves are already written above (## Context is
   the before, ## Goal is the after): confirm the prediction, don't reconstruct
   it from memory.
