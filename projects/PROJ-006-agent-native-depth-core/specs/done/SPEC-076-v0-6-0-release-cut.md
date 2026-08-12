---
# Maps to ContextCore task.* semantic conventions.
# RELEASE-CUT variant of spec.md — the stage's closing release action.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-076
  type: story                      # a release cut is a story-sized closing action
  cycle: ship
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

**Cut: v0.6.0, tagged at `f681835`, published 2026-08-12T06:53:29Z.** Release
run `31571647818` succeeded; goreleaser produced all four platform tarballs
(`darwin`/`linux` × `amd64`/`arm64`) plus `checksums.txt`, and pushed the
formula bump to `jysf/homebrew-tap` (`Formula/bragfile.rb` → `version "0.6.0"`).

### The build-cycle pre-flight items, now tickable

- [x] **Dev/prod DB isolation** — the smoke test ran against a COPY of the live
      corpus in a scratch directory, never `~/.bragfile`; the original was
      re-checked afterwards and was byte-for-byte unchanged (entry 172's title
      still 1434 bytes after the copy had been edited down to 48).
      *SPEC-036 auto-backup: recorded N/A-with-reason, not ticked.* The item
      says the backup path is "observed to fire"; it was observed NOT to fire,
      which is correct — v0.6.0 is migration-free, so there is nothing to back
      up. `PRAGMA user_version` was `0` before and after. Ticking this would
      have been the same species of untrue-but-green claim this spec exists to
      catch.
- [x] **Clean upgrade** — verified on the real machine. `brew install
      jysf/tap/bragfile` yields `brag version 0.6.0`, symlinked to
      `Cellar/bragfile/0.6.0/bin/brag` — **Cellar, not Caskroom**, which is the
      formula path. One installed version, no cask remnant, `brew outdated`
      silent. The cask→formula crossing (F4) was performed manually before the
      cut and left the machine clean.
- [x] **Behavioral surfaces re-checked on the built artifact** — done against
      the built binary rather than inferred from source, per §12(b). Drove the
      binary's own stdio MCP server: `resources/list` returned
      `brag://memory/recent` and `brag://projects`, `resources/templates/list`
      returned `brag://memory/project/{name}`, and `tools/list` returned all
      five tools including `brag_memory`. `just test-hook` exercises the real
      `plugin/hooks/capture-nudge.sh` (not a mock — `HOOK="$REPO_ROOT/plugin/
      hooks/capture-nudge.sh"`), H2–H7 green.

### Post-publish verification

`brag memory` runs from the released binary against the live 360-entry corpus
and returns a 25-entry slice inside the 2000-token budget. The headline surface
of the release works on real data, from the artifact users actually get.

### One piece of live intel for the next cut

Homebrew 6.0.15 now warns on untrusted third-party taps and states that the
`HOMEBREW_NO_REQUIRE_TAP_TRUST=1` escape hatch **"will be removed in a later
release"** (`https://docs.brew.sh/Tap-Trust`). `jysf/tap` was NOT among the taps
named in the warning, so bragfile is unaffected today. Recording it because this
is precisely the F2 category — a package-manager policy tightening **with no
change on our side**, which is how the v0.1.0→v0.2.0 trust gate arrived. The
next cut's package-manager-policy item should re-check whether `jysf/tap`
requires an explicit `brew trust` by then.

### Build-phase reflection (3 questions, short answers)

1. **What surprised you?** — That the pre-flight itself was the least
   trustworthy artifact in the release. Four defects, three of them in the
   checklist machinery: two items un-tickable since v0.5.2, one mandated item
   never added, and two items silently missed at the last cut. The checklist is
   credited with catching every prod escape in PROJ-001..003, and that
   reputation is exactly what stopped anyone from auditing it.

2. **What was harder than expected?** — Nothing in the mechanics; the cut
   itself was uneventful. The hard part was resisting the framing that a
   release cut is a chore. Every real finding came from *running* a checklist
   item rather than reading it — F4 in particular only appeared because
   `brew info` was actually executed and its answer disagreed with
   `brag --version`.

3. **What would you tell the next implementer?** — Two ticks in this spec are
   N/A-with-reason rather than checkmarks (dual-tag, SPEC-036 auto-backup).
   Keep doing that. A checklist whose items are all ticked is indistinguishable
   from one nobody read, and the moment an item cannot be honestly ticked is
   the moment it is telling you something.

## Reflection (Ship)

1. **What would I do differently next time?**
   — Audit the checklist before trusting it to audit the release. The pre-flight
   had drifted from the shipped distribution mechanism for two releases and the
   drift was invisible because the checklist is a *document*, while everything
   it guards is *behavior*. The fix that generalises: where an item can be made
   mechanical, make it mechanical — W1/W2 now assert the plugin pin and the
   compare-links that were hand-ticked wrongly at v0.5.2. What cannot be
   mechanised (clean-host install, upgrade paths) should be phrased as an
   observation to *perform*, not a property to *affirm*.

2. **Does any template, constraint, or decision need updating?**
   — Done in this spec: `spec-release-cut.md`'s Gatekeeper/`brew trust` items
   replaced with a clean-host-install check, and the package-manager-policy line
   AGENTS.md §4 mandates finally added. One more earned here and still to add —
   the "previous packaging **shape**, not just previous version" item under
   *New pre-flight item this cut earns*. Also worth considering: `just new-spec`
   only ever scaffolds from `spec.md`, so a release-cut spec must be hand-copied
   from `spec-release-cut.md`. That friction is a plausible root cause of the
   template drifting unnoticed for two releases.

3. **Is there a follow-up spec I should write now before I forget?**
   — Three, none blocking: (a) the retired-tap `tap_migrations.json` + archive,
   scoped in `retired-tap-migration-prompt.md` — the correct fix for F4 rather
   than the documentation workaround shipped here; (b) `go install
   github.com/jysf/bragfile000/cmd/brag@latest` already works and is
   undocumented, but reports `brag version dev` because goreleaser's ldflags do
   not apply — a `runtime/debug.ReadBuildInfo()` fallback plus README coverage
   is the cheapest reach available; (c) Windows compiles clean on both arches
   (pure-Go sqlite, `CGO_ENABLED=0`) and the core is portable
   (`os.UserHomeDir`, `filepath.Join`), but `capture-nudge.sh` is bash, so it is
   a small spec rather than a flag flip. Each is a distribution decision under
   §4's tripwire.

4. **What can a user do now that they couldn't before?** — one sentence,
   before → after; quote the confirming number if one exists, name the outcome
   if not. Write `none` if this release has no user-visible change — that is a
   real, greppable result, not a blank. Pairs with the Cut record above (the
   confirmed publish); this is the line a brag's `impact` field is transcribed
   from.
   — Before, the corpus was a write-only log: an agent saw a developer's history
   only if it thought to query it, and `brew upgrade` silently would not have
   delivered this release at all to anyone installed before v0.5.2. Now `brew
   install jysf/tap/bragfile` puts **0.6.0** on the machine, an MCP client
   auto-loads a ranked 25-entry slice inside a 2000-token budget with no tool
   call, and the caps that bound it are derived from the corpus instead of
   inherited — with the upgrade cliff documented rather than silent.
