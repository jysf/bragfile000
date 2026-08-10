---
# Maps to ContextCore task.* semantic conventions.
# RELEASE-CUT variant of spec.md — the stage's closing release action.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-076
  type: story                      # a release cut is a story-sized closing action
  cycle: design                    # frame | design | build | verify | ship
  blocked: false                   # SPEC-072/073/074/075 are all on main
  priority: high
  complexity: S

project:
  id: PROJ-006
  stage: STAGE-019
repo:
  id: bragfile

agents:
  architect: claude-opus-5
  implementer: claude-opus-5       # usually same Claude, different session
  created_at: 2026-08-10

references:
  decisions:
    - DEC-040                      # binary formula over cask — the migration this cut exposes
    - DEC-043                      # blended rank fusion (brag memory)
    - DEC-044                      # token budget + line shape
    - DEC-045                      # MCP push surface
    - DEC-046                      # capture field caps + the edit path
  constraints:
    - one-spec-per-pr
  related_specs:
    - SPEC-069                     # the v0.5.1 release-cut runbook this mirrors
    - SPEC-071                     # cask → formula migration (v0.5.2)
    - SPEC-072                     # MCP list filter parity
    - SPEC-073                     # brag memory
    - SPEC-074                     # MCP resources + brag_memory
    - SPEC-075                     # capture validation caps (the v0.6.0 gate)
---

# SPEC-076: v0.6.0 release cut

## Context

v0.6.0 is the **agent-native depth** release. It closes STAGE-019
(`corpus-as-agent-memory`, backlog complete: SPEC-072/073/074) and ships behind
STAGE-018's gate, SPEC-075 — the caps spec that had to land *first* because
DEC-044 derived the memory slice's entire budget model from caps that turned out
to be underived. Cutting v0.6.0 before SPEC-075 would have shipped that model
knowing it was wrong.

What ships: `brag memory` (the blended, token-budgeted slice), the MCP push
surface (three `brag://` resources + `brag_memory`), MCP `brag_list` time-window
and provenance filters, and the re-derived capture caps with `brag edit`
validating on write-of-changed-field.

Prior runbook: SPEC-069 (v0.5.1). Release mechanics: AGENTS.md §4.

**This cut is not routine.** Design-time pre-flight surfaced four defects, three
of them in the pre-flight machinery itself, and one that is a live, silent,
user-facing regression dating to v0.5.2. They are recorded in full under
*Design-time pre-flight findings* below because the whole value of the R2
checklist is that operational problems become ticked design-time items rather
than re-learned prod surprises — and this round the checklist was the thing that
had rotted.

## Goal

Cut, tag, and publish v0.6.0 per AGENTS.md §4; repair the three pre-flight
defects so the checklist is honest again; and fix the cask→formula upgrade
cliff so existing users actually receive this release.

## Inputs

- **Files to read:** `CHANGELOG.md`, `plugin/.claude-plugin/plugin.json`,
  `README.md` §Install, `.goreleaser.yaml`, `.github/workflows/release.yml`,
  `projects/_templates/spec-release-cut.md`, AGENTS.md §4.
- **External APIs:** GitHub Releases, the `jysf/homebrew-tap` Homebrew tap,
  the retired `jysf/homebrew-bragfile` tap.
- **Related code paths:** none — this cut changes no Go code.

## Outputs

- **Files created:** this spec.
- **Files modified:** `CHANGELOG.md` (dated `[0.6.0]` section; the missing
  `[0.5.2]` compare-link; `[Unreleased]` repointed),
  `plugin/.claude-plugin/plugin.json` (version pin), `README.md` (§Install
  upgrade note for pre-v0.5.2 cask installs),
  `projects/_templates/spec-release-cut.md` (pre-flight repairs),
  `projects/.../stages/STAGE-019-…md` (closed).
- **New exports:** —
- **Database changes:** none. Migration-free, as a release cut should be.

## Design-time pre-flight findings

Four defects, found by running the checklist rather than reading it.

### F1 — Two pre-flight items have been un-tickable since v0.5.2 *(fixed)*

The checklist required that README §Install contain an `xattr -dr
com.apple.quarantine` note and a `brew trust --cask` step. SPEC-071
**deliberately deleted both** when it migrated cask → formula: formulae are not
Gatekeeper-quarantined (DEC-040), which is the entire point of the migration.
AGENTS.md §4 was updated to say so; the template was not.

So from v0.5.2 until now, ticking those two items honestly was impossible —
satisfying them would mean re-adding documentation DEC-040 removed. Replaced
with a check on the *property* DEC-040 buys (clean-host install, no Gatekeeper
prompt, no `xattr`) rather than the presence of a workaround.

### F2 — The mandated package-manager-policy check was never in the list *(fixed)*

AGENTS.md §4 explicitly requires "add a *check the package manager's current
install/trust policy* line to the release pre-flight, since this gate appeared
between the v0.1.0 and v0.2.0 cuts **with no change on our side**." That line
was never added. Added now, phrased to make the point that "we changed nothing"
is not evidence the install path still works.

### F3 — Two pre-flight items were silently missed at the v0.5.2 cut *(fixed here)*

- `plugin/.claude-plugin/plugin.json` is pinned to **`0.5.1`** although v0.5.2
  shipped. The item "plugin version pin matches the tag" was not ticked
  truthfully.
- `CHANGELOG.md` has **no `[0.5.2]` compare-link at all**, and `[Unreleased]`
  still points at `v0.5.1...HEAD`. The item "compare-links repointed" was
  likewise missed.

Both are cosmetic in isolation. Together they say the v0.5.2 cut — the one that
changed the distribution mechanism — ran its checklist loosely, which is the
same cut that produced F1 and F4.

### F4 — **The cask→formula migration has a silent upgrade cliff** *(the real one)*

`brew upgrade` does not move a pre-v0.5.2 user to v0.5.2 or later, and nothing
tells them so. Verified on this machine:

```
/opt/homebrew/bin/brag -> /opt/homebrew/Caskroom/bragfile/0.5.1/brag
brag version 0.5.1
brew info jysf/tap/bragfile   →  stable 0.5.2 … Not installed
brew outdated                 →  (silent — nothing about brag)
brag memory --help            →  ABSENT
```

The mechanism: v0.5.2 moved distribution to a **formula** on the new
`jysf/homebrew-tap`. The retired `jysf/homebrew-bragfile` tap is still tapped
and still contains a goreleaser-generated **cask pinned at `0.5.1`**, which
Homebrew considers current. So an existing cask user sits at 0.5.1 with:

- no upgrade signal (`brew outdated` is empty),
- no migration instruction anywhere — README and CHANGELOG say only that the
  old tap "is retired", which tells a user nothing about what to *do*,
- and, after this cut, none of v0.6.0's headline surface: `brag memory` is
  simply not in their binary.

This is the fourth consecutive prod escape in the operational/runtime class the
cross-project retro identified, and it is the sharpest: the release side looks
completely healthy — the tap has 0.5.2, CI is green, goreleaser succeeded — while
the install side is frozen. Nothing in the current pre-flight would catch it,
because "clean upgrade: `brew upgrade` from the prior minor verified" was read
as *formula → formula*, and the failing path is *cask → formula*, which only
exists once.

**Resolution for this cut** (deliberately the cheap half): document the one-time
migration prominently in README §Install and in the `[0.6.0]` CHANGELOG entry,
and add a pre-flight item that names the *previous packaging shape*, not just
the previous version. **Not** attempted here: republishing a terminal cask to
the retired tap to auto-point users at the formula. That touches the
distribution mechanism and is therefore a decision requiring a downsides pass
under AGENTS.md §4's tripwire rule — it is not something to slip into a release
cut, which is exactly the mistake (`1582572`) that created this whole situation.
Raised as a follow-up instead.

## Acceptance Criteria

- [ ] `CHANGELOG.md` has a dated `## [0.6.0] - 2026-08-10` section containing
      SPEC-072/073/074/075; `[Unreleased]` is empty of them.
- [ ] `CHANGELOG.md` compare-links include `[0.6.0]` and the previously missing
      `[0.5.2]`, and `[Unreleased]` points at `v0.6.0...HEAD`.
- [ ] `plugin/.claude-plugin/plugin.json` `version` == `0.6.0` == the tag.
- [ ] README §Install carries a "Upgrading from 0.5.1 or earlier" note with the
      exact uninstall-cask / install-formula commands.
- [ ] All pre-flight items in the corrected checklist are ticked truthfully, or
      explicitly recorded as N/A with a reason.
- [ ] `just test`, `just test-docs`, `just test-hook`, `go vet`, `gofmt -l .`
      all green on the tagged commit.
- [ ] After publish: `brew install jysf/tap/bragfile` on a clean host yields
      `brag version 0.6.0` with `brag memory` present, no Gatekeeper prompt and
      no `xattr` step.

## Failing Tests

Doc/harness assertions, per the release-cut convention. Both are the mechanical
form of an F3 item that was previously a hand-ticked checkbox.

- **`scripts/test-docs.sh`**
  - **W1** — the `plugin.json` version pin equals the **latest** dated
    `## [x.y.z]` section in the CHANGELOG.
  - **W2** — every dated `## [x.y.z]` heading has a matching `[x.y.z]:`
    compare-link.

**Mutation-checked, both directions.** Restoring the exact v0.5.2 state makes
each fail for its own reason and only its own reason: pinning `0.5.1` fails W1
(`pins 0.5.1 but the latest released CHANGELOG section is 0.6.0`) while W2 stays
green; deleting the `[0.6.0]` compare-link fails W2 (`version heading(s) with no
compare-link: 0.6.0`) while W1 stays green.

> **W1's first draft was toothless and is worth recording.** It asserted the pin
> had *a* dated section — but at the v0.5.2 cut the pin sat on `0.5.1`, which had
> a perfectly good dated section, so the weak form would have passed and caught
> exactly nothing. The defect is the pin falling **behind** the newest release,
> so the newest release is what it has to be compared against. This is the same
> failure this stage hit in SPEC-075's punch-list #2 — an assertion sized on the
> wrong axis, green for a reason unrelated to the property it claims to pin —
> caught here only because the mutation was run instead of assumed.

## Implementation Context

### Decisions that apply

- `DEC-040` — binary formula over cask. The decision is sound; F4 is a gap in
  its *migration*, not in the decision.
- `DEC-043`/`DEC-044`/`DEC-045`/`DEC-046` — the surface this release ships.

### Constraints that apply

- `one-spec-per-pr` — the tag is cut from `main` after this spec's PR lands; it
  cannot share a PR with the content it tags.

### Prior related work

- `SPEC-069` (shipped) — the v0.5.1 release-cut runbook this mirrors.
- `SPEC-071` (shipped) — the cask → formula migration whose follow-through F1
  and F4 both trace to.

### Out of scope (for this spec specifically)

- Republishing a terminal cask to the retired tap (see F4 — a distribution
  decision, not a release chore).
- The ~9 STAGE-018 audit nits, which trail into v0.6.1 per that stage's scope
  note.
- SPEC-070 (`brag project goto`), drafted but never built.

## Notes for the Implementer

### Release runtime/operational pre-flight (all must be ticked at design)

- [x] **Dual-tag-on-same-commit** — N/A by construction: no RC tag is being
      cut for v0.6.0, so §4 Pattern 1's failure mode cannot arise. Recorded as
      N/A-with-reason rather than ticked.
- [x] **macOS clean-host install** — the DEC-040 property holds: distribution
      is `brews:` in `.goreleaser.yaml` (formula, not cask), README §Install
      states no Gatekeeper prompt / no `xattr` / no `brew trust`, and the
      sibling `crustyimg` formula on the same tap is the live precedent.
      Re-confirmed post-publish by the clean-host acceptance criterion.
- [x] **Package manager's current install/trust policy re-checked** — local
      Homebrew is **6.0.15**, i.e. ≥6.0, the version that introduced the
      third-party-tap trust gate. `brew info jysf/tap/bragfile` resolves and
      reports `stable 0.5.2` with no trust prompt, confirming the formula path
      on this tap still clears the gate.
- [x] **Distribution mechanism unchanged since the last cut** — `.goreleaser.yaml`
      `brews:` block, the `jysf/homebrew-tap` target and the install path are
      byte-identical to v0.5.2. No distribution-decision checklist run is
      required for this cut.
- [ ] **Dev/prod DB isolation** — the smoke test must run against a THROWAWAY
      DB via `--db`, never `~/.bragfile`; observe the SPEC-036 auto-backup path
      fire. *(Build-cycle item — cannot be ticked at design.)*
- [ ] **Clean upgrade** — verify BOTH paths, per F4: formula→formula
      (`brew upgrade`), and the one-time **cask→formula** migration that F4
      shows is silent today. `brag --version` prints `0.6.0`; no migration
      surprise. *(Build-cycle item.)*
- [x] **CHANGELOG** — the `[0.6.0]` dated section and the compare-link repoint
      are carried by this spec's Outputs, including the `[0.5.2]` link missed
      at that cut (F3).
- [x] **Plugin version pin** — carried by this spec's Outputs: `0.5.1` → `0.6.0`,
      correcting the F3 miss in the same change.
- [ ] **Behavioral surfaces re-checked on the built artifact** — `claude plugin
      details` shows the MCP server registered; the Stop hook fires in a
      throwaway repo. Per the §12(b) refinement this is re-run on the *built
      artifact*, not inferred from source. *(Build-cycle item.)*

### New pre-flight item this cut earns

Add to the template once v0.6.0 ships:

- [ ] **Previous packaging shape, not just previous version.** "Clean upgrade
      from the prior minor" silently means formula→formula. When the prior
      release changed packaging *shape* (cask→formula, tap move, rename), the
      upgrade path that must be verified is the **cross-shape** one, and it
      exists exactly once — the cut after the migration. Missing it strands
      every pre-migration user with no signal on either side (F4).

---

## Build Completion

*Filled in during the **build** cycle.*

## Reflection (Ship)

*Appended during the **ship** cycle.*
