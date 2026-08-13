---
# Maps to ContextCore task.* semantic conventions.
# RELEASE-CUT variant of spec.md — use this template for a
# SPEC-NNN-vX-Y-Z-release-cut spec (the stage's closing release action).
# Everything in spec.md applies; the difference is the runtime/operational
# pre-flight checklist under ## Notes for the Implementer, which MUST be
# ticked at design (see AGENTS.md §4).

task:
  id: SPEC-077
  type: story                      # a release cut is a story-sized closing action
  cycle: design                    # frame | design | build | verify | ship
  blocked: false                   # everything it ships is already on main
  priority: high
  complexity: S                    # S | M | L  (L means split it)

project:
  id: PROJ-006
  stage: STAGE-018
repo:
  id: bragfile

agents:
  architect: claude-opus-5
  implementer: claude-opus-5       # usually same Claude, different session
  created_at: 2026-08-13

references:
  decisions: [DEC-021, DEC-026, DEC-040, DEC-046]
  constraints: [one-spec-per-pr]
  related_specs: [SPEC-069, SPEC-076]
---

# SPEC-077: v0.6.1 release cut

## Context

A patch release. v0.6.0 shipped the agent-native depth surface; this cuts the
correctness work that trailed it — STAGE-018's seven audit-nit fixes, each with
a regression test — plus `go install` finally reporting its own version, and the
doc version claims that had been stale through two releases.

Nothing here changes behaviour a user relies on, which is what makes it a patch
rather than a minor. Everything it ships is already on `main` (#150-#153).

Prior runbook: SPEC-076 (v0.6.0). Release mechanics: AGENTS.md §4.

## Goal

Cut, tag and publish v0.6.1 per AGENTS.md §4; verify the formula→formula upgrade
path, which is the first ordinary upgrade since the cask→formula migration; and
confirm `go install …@latest` now reports a real version.

## Inputs

- **Files to read:** `CHANGELOG.md`, `plugin/.claude-plugin/plugin.json`,
  `.goreleaser.yaml`, `.github/workflows/release.yml`, AGENTS.md §4.
- **External APIs:** GitHub Releases, the `jysf/homebrew-tap` tap, the Go module
  proxy (for the `go install` check).
- **Related code paths:** none — this cut changes no Go code.

## Outputs

- **Files created:** this spec.
- **Files modified:** none at cut time. The CHANGELOG `[0.6.1]` section, the
  compare-links, the plugin pin and the doc version claims all landed WITH the
  work (#153) rather than in a release-day sweep — which is the ordering
  W1/W2/W3 now enforce.
- **Database changes:** none. Migration-free.

## Acceptance Criteria

- [ ] `v0.6.1` tagged on `main`, release workflow green, four platform tarballs
      plus checksums published.
- [ ] `jysf/homebrew-tap`'s `Formula/bragfile.rb` reports `version "0.6.1"`.
- [ ] `brew upgrade bragfile` moves a 0.6.0 install to 0.6.1 — the first
      **formula→formula** upgrade since the cask migration, and the path
      SPEC-076's F4 could not exercise.
- [ ] `brag --version` prints `0.6.1`; `brag list` opens the existing database
      with no migration and no backup sidecar.
- [ ] `go install github.com/jysf/bragfile000/cmd/brag@latest` prints `0.6.1`,
      not `dev`. **Only checkable after publish** — the proxy serves the newest
      tag, so before the cut it necessarily still returns v0.6.0, which predates
      the fix.

## Failing Tests

Doc/harness assertions, all green on `main`, and every one of them added because
an earlier cut got the corresponding item wrong by hand:

- **`scripts/test-docs.sh`**
  - **W1** — plugin pin equals the latest dated CHANGELOG section (`0.6.1`).
  - **W2** — every dated CHANGELOG heading has a compare-link.
  - **W3** — README `**Status:**` and the tutorial's "shipped as of" name the
    latest release. This one **fired for real during preparation**: bumping the
    CHANGELOG to `0.6.1` immediately failed W3 and named both files.
  - **W4** — no stage with `status: shipped` still carries reflection
    placeholders (STAGE-018 and STAGE-019 both closed this cycle).
- **`cmd/brag/version_test.go`** — `resolveVersion` accepts only a clean
  `vX.Y.Z`; a pseudo-version or `(devel)` stays `dev`, so the DEC-026 guard
  still fires.

## Implementation Context

### Decisions that apply

- `DEC-026` — the dev/prod migration guard the `go install` change feeds; the
  reason its version pattern is deliberately strict.
- `DEC-040` — binary formula over cask. Unchanged this cut (verified).
- `DEC-021` — a failed backup aborts the open; why the backup-filename nit was
  fixed by widening the timestamp rather than dodging the collision.
- `DEC-046` — the caps and reserved-prefix machinery v0.6.0 shipped.

### Constraints that apply

- `one-spec-per-pr` — the tag is cut from `main` after this spec's PR lands.

### Prior related work

- `SPEC-076` (shipped) — the v0.6.0 cut. Its four findings produced most of the
  pre-flight items this cut is the first to actually run against.
- `SPEC-069` (shipped) — the v0.5.1 runbook.

### Out of scope

- The retired-tap `tap_migrations.json` + archive (runbook in
  `retired-tap-migration-prompt.md`). Not a release blocker; measured exposure
  is effectively zero.
- STAGE-020's evidence-links work.
- The two carried-forward audit nits (`MergeTags`, `$EDITOR`-with-spaces) —
  both need decisions, not patches.

## Notes for the Implementer

Gotchas, style preferences, reuse opportunities.

### Release runtime/operational pre-flight (all must be ticked at design)

First cut to run against the corrected checklist — SPEC-076 rewrote two items,
added two, and v0.6.1 added a third (W3). Evidence gathered before ticking.

- [x] **Dual-tag-on-same-commit** — N/A with reason: no RC tag is being cut, so
      §4 Pattern 1's failure mode cannot arise.
- [x] **macOS clean-host install** — the DEC-040 property holds. Distribution is
      still `brews:` (formula, not cask) and README §Install still states no
      Gatekeeper prompt, no `xattr`, no `brew trust`. Re-confirmed post-publish
      by the acceptance criteria.
- [x] **Package manager's current install/trust policy re-checked** — and it
      MOVED. Homebrew is now **6.0.17**; it was 6.0.15 at the v0.6.0 cut two
      days ago. Re-ran the install against the live tap: no trust gate, no
      `Refusing to load…`, and the third-party-tap warning that fired at v0.6.0
      (naming `anomalyco/tap` and `dopplerhq/doppler`, never `jysf/tap`) does
      not appear at all now. This is exactly the item's premise — their policy
      moves without any change on our side — so it was re-run rather than
      inherited from the last cut's answer.
- [x] **Distribution mechanism unchanged since the last cut** — verified, not
      assumed: `git diff v0.6.0..HEAD -- .goreleaser.yaml
      .github/workflows/release.yml` is **empty**. No distribution-decision
      checklist run required.
- [ ] **Dev/prod DB isolation** — smoke test against a THROWAWAY db via `--db`,
      never `~/.bragfile`. *(Build-cycle.)* Note the SPEC-036 auto-backup
      sub-clause will again be **N/A with reason**: v0.6.1 adds no migration, so
      there is nothing to back up and the path correctly will not fire. Recorded
      rather than ticked, per SPEC-076's precedent.
- [ ] **Clean upgrade** — `brew upgrade bragfile` from 0.6.0. *(Build-cycle.)*
      This is the **first ordinary formula→formula upgrade since the cask
      migration** — the path SPEC-076's F4 identified but could not exercise,
      because the machine had to cross the cask boundary manually first. It is
      the most informative check in this cut.
- [x] **Previous packaging SHAPE, not just previous version** — N/A with reason.
      v0.6.0 → v0.6.1 is formula → formula; no shape change, no tap move, no
      rename. The cross-shape path this item exists for was the v0.5.2 → v0.6.0
      transition and is behind us.
- [x] **CHANGELOG** — `[0.6.1]` dated section written, `[Unreleased]` empty,
      compare-links repointed. Landed in #153 with the work; W1/W2 pin it.
- [x] **Plugin version pin** — `0.6.1`, matching the tag. W1 pins it.
- [x] **Docs state the version being shipped, and describe what it added** —
      W3 green. It fired for real during preparation, naming both stale files
      the moment the CHANGELOG was bumped, which is the check earning its place
      one release after those claims had rotted through two.
- [ ] **Behavioral surfaces re-checked on the built artifact** — `claude plugin
      details` / the Stop hook in a throwaway repo, checked against the BUILT
      binary per §12(b). *(Build-cycle.)*

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
   — <answer>
