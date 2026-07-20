---
# Maps to ContextCore task.* semantic conventions.
# RELEASE-CUT variant of spec.md — use this template for a
# SPEC-NNN-vX-Y-Z-release-cut spec (the stage's closing release action).
# Everything in spec.md applies; the difference is the runtime/operational
# pre-flight checklist under ## Notes for the Implementer, which MUST be
# ticked at design (see AGENTS.md §4).

task:
  id: SPEC-XXX
  type: story                      # a release cut is a story-sized closing action
  cycle: design                    # frame | design | build | verify | ship
  blocked: false                   # usually blocked on the feature specs merging to main
  priority: high
  complexity: S                    # S | M | L  (L means split it)

project:
  id: PROJ-XXX
  stage: STAGE-XXX
repo:
  id: <repo-id>

agents:
  architect: claude-opus-4-8
  implementer: claude-opus-4-8     # usually same Claude, different session
  created_at: YYYY-MM-DD

references:
  decisions: []
  constraints: [one-spec-per-pr]
  related_specs: []
---

# SPEC-XXX: vX.Y.Z release cut

## Context

Why does this release cut exist now? Link to:
- The parent `STAGE-XXX` whose backlog this cut closes
- The feature specs whose surface ships in this release
- AGENTS.md §4 (the authoritative release mechanics + the lessons-earned
  addenda)
- The prior release-cut spec (the runbook precedent)

## Goal

1–2 sentences. Cut, tag, and publish vX.Y.Z per AGENTS.md §4, and verify a
clean upgrade from the prior minor.

## Inputs

- **Files to read:** `CHANGELOG.md`, the homebrew tap formula/cask — why
- **External APIs:** GitHub Releases, the Homebrew tap
- **Related code paths:** `.goreleaser.yaml`, `.github/workflows/release.yml`

## Outputs

- **Files created:** —
- **Files modified:** `CHANGELOG.md` (dated section), tap formula/cask, any
  release-note docs
- **New exports:** —
- **Database changes:** none (a release cut should be migration-free; if it
  is not, that is a design-time surprise to raise)

## Acceptance Criteria

Testable outcomes. Cover the tag, the publish, and a clean upgrade.

- [ ] Criterion 1 (testable)
- [ ] Criterion 2 (testable)

## Failing Tests

Written during **design**, BEFORE build. For a release cut these are often
doc/harness assertions (CHANGELOG shape, compare-links, tap version).

- **`path/to/test.file`**
  - `"test description 1"` — asserts: ...

## Implementation Context

*Read this section (and the files it points to) before starting
the build cycle.*

### Decisions that apply

- `DEC-NNN` — <one-line summary>

### Constraints that apply

- `one-spec-per-pr` — the release tag is cut from `main` after the feature
  PRs land; it cannot share a PR with the code it tags.

### Prior related work

- `SPEC-YYY` (shipped) — the prior release-cut runbook precedent

### Out of scope (for this spec specifically)

- ...

## Notes for the Implementer

Gotchas, style preferences, reuse opportunities.

### Release runtime/operational pre-flight (all must be ticked at design)

- [ ] Dual-tag-on-same-commit: RC tag + release deleted before the final tag
      is cut at the same commit (§4 Pattern 1).
- [ ] macOS Gatekeeper: `xattr -dr com.apple.quarantine <bin>` note present in
      README §Install.
- [ ] Homebrew 6.0+: `brew trust --cask <tap>/<cask>` documented in README and
      run once at the cut.
- [ ] Dev/prod DB isolation: the RC smoke test runs against a THROWAWAY DB,
      never ~/.bragfile; the SPEC-036 auto-backup path is observed to fire.
- [ ] Clean upgrade: `brew upgrade` from the prior minor verified;
      `brag --version` prints the new tag; no migration surprise.
- [ ] CHANGELOG: the `[x.y.z]` dated section is moved out of `[Unreleased]`;
      compare-links repointed.
- [ ] Plugin version pin (v0.3.0+): `plugin/.claude-plugin/plugin.json`
      `version` matches the tag.
- [ ] Behavioral surfaces re-checked on the built artifact (per the §12(b)
      refinement): `claude plugin details` shows the MCP server registered;
      the Stop hook fires in a throwaway repo.

---

## Build Completion

*Filled in at the end of the **build** cycle, before advancing to verify.*

- **Branch:**
- **PR (if applicable):**
- **All acceptance criteria met?** yes/no
- **Cut record (one line):** the user-facing outcome now live, plus the clean-
  upgrade check that confirms the publish landed — the one line a review or brag
  would quote. E.g. *"v0.4.0→v0.5.0 `brew upgrade` clean; MCP-install now
  first-class; prod DB opened, 189 entries intact."* If the release is pure infra
  with no user-facing change, say so — a real, greppable outcome, not a blank.
  Filled once the tag is published.
- **New decisions emitted:**
  - `DEC-NNN` — <title> (if any)
- **Deviations from spec:**
  - [list]
- **Follow-up work identified:**
  - [any new specs for the stage's backlog]

### Build-phase reflection (3 questions, short answers)

Process-focused: how did the build go? What friction did the spec create?

1. **What was unclear in the spec that slowed you down?**
   — <answer>

2. **Was there a constraint or decision that should have been listed but wasn't?**
   — <answer>

3. **If you did this task again, what would you do differently?**
   — <answer>

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
   if not. Write `none` if this release has no user-visible change — that is a
   real, greppable result, not a blank. Pairs with the Cut record above (the
   confirmed publish); this is the line a brag's `impact` field is transcribed
   from.
