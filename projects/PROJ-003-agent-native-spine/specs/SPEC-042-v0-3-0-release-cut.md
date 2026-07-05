---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-042
  type: story                      # epic | story | task | bug | chore
  cycle: frame                     # STUB — peeled from SPEC-041; awaiting its own design session
  blocked: true                    # blocked on SPEC-041 merging to main (the tag is cut from main after)
  priority: high
  complexity: S                    # S | M | L  (L means split it)

project:
  id: PROJ-003
  stage: STAGE-009
repo:
  id: bragfile

agents:
  architect: claude-opus-4-8
  implementer: claude-opus-4-8     # same Claude, different session (claude-only variant)
  created_at: 2026-07-04

references:
  decisions: []
  constraints: [one-spec-per-pr]
  related_specs: [SPEC-041, SPEC-037]
---

# SPEC-042: v0.3.0 release cut

> **STUB — peeled from SPEC-041 at design (2026-07-04).** This spec exists
> because SPEC-041's L-watch decided to PEEL the v0.3.0 release cut into its
> own spec (see SPEC-041 "L-watch outcome"), mirroring SPEC-037's
> release-runbook precedent and the STAGE-007/008 doc/release-split
> discipline. It is **not yet designed** — a later design session fills the
> sections below. Recorded here so the scope is not lost and the STAGE-009
> backlog reflects the true 5-spec shape.

## Context

STAGE-009's four build specs (SPEC-038 streak fix, SPEC-039 milestones,
SPEC-040 MCP server, SPEC-041 plugin packaging) deliver the v0.3.0 feature
surface. This spec **cuts and ships v0.3.0** — the stage's closing action —
following AGENTS.md §4 release mechanics. It is **blocked on SPEC-041 merging
to `main`**: the v0.3.0 tag is cut from `main` *after* the plugin lands, which
is the structural reason SPEC-041 peeled this out (a release tag cannot live
in the same PR that adds the code it tags — `one-spec-per-pr`).

## Goal

Cut, tag, and publish v0.3.0 to the Homebrew tap per AGENTS.md §4, and verify
a clean `brew upgrade` from v0.2.x, closing STAGE-009.

## Scope (to design)

The §4 release mechanics, all of which apply (trust-but-verify each claim at
build/ship, per the coordinator's release-mechanics premise-audit trigger —
`gh release view` / the tap cask read):

- **CHANGELOG `[0.3.0]`** — move the `[Unreleased]` block to a dated `[0.3.0]`
  section covering SPEC-038 (streak fix), SPEC-039 (milestones), SPEC-040
  (`brag mcp serve` + provenance), SPEC-041 (Claude Code plugin).
- **RC → final dual-tag** — the optional `v0.3.0-rc1` smoke, then `v0.3.0`,
  under the **dual-tag-on-same-commit** rule (§4: default to pattern 1 —
  delete the RC tag + release before cutting the final tag at the same commit;
  recovery commands documented in §4).
- **Homebrew tap bump** — bump `bragfile.rb` in `github.com/jysf/homebrew-bragfile`.
- **Release pre-flight** — the **Homebrew 6.0 `brew trust --cask
  jysf/bragfile/bragfile`** step + the **Gatekeeper `xattr -dr
  com.apple.quarantine`** note, both carried into the pre-flight (§4). Add the
  "check the package manager's current install/trust policy" line.
- **Clean-upgrade verification** — `brew upgrade jysf/bragfile/bragfile` moves
  a v0.2.x install to v0.3.0 with no migration surprise (the core is
  migration-free); `brag --version` reports v0.3.0; the full v0.2 surface plus
  the new `brag mcp serve` work on the upgraded binary.
- **Release doc sweep** — `docs/tutorial.md` + `docs/architecture.md` plugin
  walkthroughs (deferred here from SPEC-041's Outputs) and any release-note
  tutorial mention. Run the doc premise-audit greps.

## Notes

- **Related:** SPEC-041 (the plugin this release includes; must merge first),
  SPEC-037 (the release-runbook precedent), AGENTS.md §4 (the authoritative
  release mechanics + the three lessons-earned addenda), DEC-025 (plugin
  packaging), DEC-024 (MCP + provenance).
- **Plugin version pin:** `plugin/.claude-plugin/plugin.json` carries
  `version: "0.3.0"`; if this cut slips the number, bump the manifest to match
  (the `claude plugin tag` flow validates plugin.json ↔ marketplace agreement).
- Design this in a fresh session once SPEC-041 has merged.

## Inputs

- **Files to read:** `path/to/file.ext` — why
- **External APIs:** <name, docs link, auth>
- **Related code paths:** `src/some/module/`

## Outputs

- **Files created:** `path/to/new.ext` — purpose
- **Files modified:** `path/to/existing.ext` — what changes
- **New exports:** <names and signatures>
- **Database changes:** <migrations>

## Acceptance Criteria

Testable outcomes. Cover happy path, error cases, edge cases.

- [ ] Criterion 1 (testable)
- [ ] Criterion 2 (testable)

## Failing Tests

Written during **design**, BEFORE build. The implementer's job in
**build** is to make these pass.

- **`path/to/test.file`**
  - `"test description 1"` — asserts: ...

## Implementation Context

*Read this section (and the files it points to) before starting
the build cycle. It is the equivalent of a handoff document, folded
into the spec since there is no separate receiving agent.*

### Decisions that apply

- `DEC-NNN` — <one-line summary of why this matters here>
- `DEC-MMM` — <one-line summary>

### Constraints that apply

These constraints apply to the paths touched by this task (see
`/guidance/constraints.yaml` for full text):

- `constraint-id-1` — <one-line summary>
- `constraint-id-2` — <one-line summary>

### Prior related work

- `SPEC-YYY` (shipped) — <one-line summary, if relevant>
- `PR #NNN` — <link, if relevant>

### Out of scope (for this spec specifically)

Explicit list of what this spec does NOT include. If any of these feel
necessary during build, create a new spec rather than expanding this one.

- ...

## Notes for the Implementer

Gotchas, style preferences, reuse opportunities.

### Release runtime/operational pre-flight (all must be ticked at design)

Adopted from the release-cut spec template (`projects/_templates/spec-release-cut.md`,
per AGENTS.md §4). Concretized for the v0.3.0 `brag` cut.

- [ ] Dual-tag-on-same-commit: `v0.3.0-rc1` tag + release deleted before the
      final `v0.3.0` tag is cut at the same commit (§4 Pattern 1).
- [ ] macOS Gatekeeper: `xattr -dr com.apple.quarantine <bin>` note present in
      README §Install.
- [ ] Homebrew 6.0+: `brew trust --cask jysf/bragfile/bragfile` documented in
      README and run once at the cut.
- [ ] Dev/prod DB isolation: the RC smoke test runs against a THROWAWAY DB,
      never ~/.bragfile; the SPEC-036 auto-backup path is observed to fire.
- [ ] Clean upgrade: `brew upgrade jysf/bragfile/bragfile` from v0.2.x
      verified; `brag --version` prints `v0.3.0`; no migration surprise (the
      core is migration-free).
- [ ] CHANGELOG: the `[0.3.0]` dated section is moved out of `[Unreleased]`;
      compare-links repointed.
- [ ] Plugin version pin (v0.3.0+): `plugin/.claude-plugin/plugin.json`
      `version` matches the `v0.3.0` tag.
- [ ] Behavioral surfaces re-checked on the built artifact (per the §12(b)
      refinement): `claude plugin details` shows the MCP server registered;
      the Stop hook fires in a throwaway repo.

---

## Build Completion

*Filled in at the end of the **build** cycle, before advancing to verify.*

- **Branch:**
- **PR (if applicable):**
- **All acceptance criteria met?** yes/no
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
