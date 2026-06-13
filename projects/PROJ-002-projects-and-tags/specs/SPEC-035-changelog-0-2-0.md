---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-035
  type: story                      # epic | story | task | bug | chore
  cycle: design                    # frame | design | build | verify | ship
  blocked: false
  priority: medium
  complexity: S                    # S | M | L  (L means split it)

project:
  id: PROJ-002
  stage: STAGE-008
repo:
  id: bragfile

agents:
  architect: claude-opus-4-7
  implementer: claude-opus-4-7     # usually same Claude, different session
  created_at: 2026-06-12

references:
  decisions: [DEC-015, DEC-016, DEC-017, DEC-018, DEC-019, DEC-020, DEC-021]
  constraints: []
  related_specs: [SPEC-034, SPEC-036, SPEC-037]
---

# SPEC-035: CHANGELOG [0.2.0]

## Context

This is the CHANGELOG spec in STAGE-008's backlog (the final stage of
PROJ-002, the v0.2.0 release). STAGE-006 (tags normalization, shipped
2026-06-07) and STAGE-007 (projects, shipped 2026-06-12) completed the
entire v0.2.0 feature surface; SPEC-036 added the migration auto-backup
safety belt. Nothing is left to build for v0.2.0 — only to record it.

This spec writes the user-facing `## [0.2.0]` release-notes entry and
reconciles three stale items the `[0.1.0]` release left behind: the
`YYYY-MM-DD` placeholder date, the lingering `[Unreleased]` `brag
completion` line, and the compare-links at the bottom. The actual git
tag and release are cut by **SPEC-037**, which references this entry;
this spec only writes the changelog.

This is a **literal-artifact-as-spec** deliverable (AGENTS.md §12): the
CHANGELOG.md edit *is* the artifact. The full post-edit file is embedded
verbatim under `## Implementation Context` for the implementer to
transcribe, and verify diffs against it. There is no Go code and no Go
test — the only automated gate is `scripts/test-docs.sh` Group O
(O1–O5), which must stay green.

## Goal

Write the `## [0.2.0] - 2026-06-12` entry in `CHANGELOG.md` (Keep-a-
Changelog format, terse and user-facing) covering tags normalization,
the `brag project` surface, the migration auto-backup safety belt, and
the folded-in `brag completion` line; and reconcile the `[0.1.0]`
placeholder date, the stale `[Unreleased]` section, and the bottom
compare-links.

## Inputs

- **Files to read:** `CHANGELOG.md` — the file to edit (current state
  embedded below).
- **Files to read:** `scripts/test-docs.sh` (Group O, lines 712–755) —
  the only automated assertions over CHANGELOG.md; the edit must keep
  them green.
- **Decisions of record:** `/decisions/DEC-015..021` — the v0.2.0
  architectural decisions the entry lists (titles transcribed below;
  no need to re-read full files for the changelog).
- **Stage context:** `projects/PROJ-002-projects-and-tags/stages/STAGE-008-polish-and-v0-2-0-release.md`
  — Design Note 4 (CHANGELOG reconciliations) and Success Criteria.

## Outputs

- **Files modified:** `CHANGELOG.md` — add the `## [0.2.0]` section;
  empty the `## [Unreleased]` section (fold `brag completion` into
  `[0.2.0]`); set the `[0.1.0]` date to `2026-05-10`; add the `[0.2.0]`
  compare-link and repoint `[Unreleased]` to `v0.2.0...HEAD`.
- **Files created:** none.
- **New exports:** none.
- **Database changes:** none.

### Premise audit (§9 status-change + inversion/count-bump)

Docs-only spec — **no code, no migration, no Go test touched.** The
required greps were run at design and reconciled here:

- **Does any Go test assert CHANGELOG content?** `grep -rln -i
  "changelog" internal/ cmd/` → **no hits.** The only CHANGELOG gate is
  the shell test `scripts/test-docs.sh` Group O (O1–O5). The edit is
  designed to satisfy all five (see Acceptance Criteria O1–O5).
- **Inversion/count-bump:** NONE. No collection whose count is asserted
  is touched (no migration, no DEC list in a Go test, no
  `schema_migrations` change).
- **Status-change — does any other doc reference the changelog
  version?** `grep -rn -iE "changelog|0\.2\.0|\[unreleased\]" docs/
  README.md` hits, each dispositioned:
  - `docs/macos-notarization-checklist.md:132,148` — **stays.** Release
    mechanics (`v0.2.0-rc1` → `v0.2.0` tagging), SPEC-037's domain, not
    a changelog claim.
  - `docs/api-contract.md:670` — **stays.** Generic stability note ("flag
    names may change between `v0.x` releases with CHANGELOG version") —
    references the changelog as a concept, no specific version.
  - `docs/api-contract.md:571`, `docs/tutorial.md:502` — **stays.**
    Illustrative `--state-note "...next: cut v0.2.0"` example text, not a
    version claim.
  - `docs/tutorial.md:664` — **stays.** "Everything in this tutorial is
    shipped in v0.2.0." Already set by SPEC-034's doc sweep; consistent
    with this entry.
  - `docs/blog/README.md:51` — **stays (out of scope).** "`CHANGELOG.md`
    — the public-facing v0.1.0 history." A blog-asset narrative line, not
    gated by `test-docs.sh`; describes the changelog's origin rather than
    making a version claim this spec must update. Flagged as a possible
    PROJ-002-close follow-up, not a blocker here.

  Conclusion: writing `[0.2.0]` requires **no companion doc edit.**

## Acceptance Criteria

Concrete and grep-checkable against the post-edit `CHANGELOG.md`.

- [ ] **AC1 — `[0.2.0]` heading present and dated.** A line equals
      `## [0.2.0] - 2026-06-12` (line-based equality, not substring).
- [ ] **AC2 — tags surface documented.** The `[0.2.0]` section names
      `` `brag tags` ``, `` `brag tag rename` `` (or `brag tag rename
      <old> <new>`), and `` `brag tag merge` `` under Added, and states
      tags are now first-class / normalized under Changed.
- [ ] **AC3 — projects surface documented.** The `[0.2.0]` section names
      `` `brag project` `` with the subcommands `new`, `list`, `show`,
      `edit`, `archive`, `delete`, `status`, `here`; the `--add-path` /
      `--remove-path` flags; and `brag add` cwd `--project` auto-fill.
- [ ] **AC4 — safety belt documented.** The `[0.2.0]` section describes
      the pre-migration auto-backup (timestamped DB snapshot before a
      schema migration; abort-on-failure; non-interactive).
- [ ] **AC5 — completion folded in.** The `brag completion` line appears
      under `[0.2.0]` Added and **no longer** under `[Unreleased]`.
- [ ] **AC6 — DECs of record listed.** A `### Decisions of record`
      subsection under `[0.2.0]` lists DEC-015, DEC-016, DEC-017, DEC-018,
      DEC-019, DEC-020, and DEC-021 (matching the `[0.1.0]` convention,
      which enumerates `DEC-xxx —` lines).
- [ ] **AC7 — `[0.1.0]` date reconciled.** The line reads
      `## [0.1.0] - 2026-05-10` (the `YYYY-MM-DD` placeholder is gone);
      the heading still matches `^## \[0\.1\.0\]` (keeps O3 green).
- [ ] **AC8 — `[Unreleased]` emptied, not deleted.** The `## [Unreleased]`
      heading remains (Keep-a-Changelog convention) with no entries under
      it; the `[Unreleased]:` link reference still exists (keeps O5 green).
- [ ] **AC9 — compare-links reconciled.** Bottom link references read:
      `[Unreleased]: …/compare/v0.2.0...HEAD`,
      `[0.2.0]: …/compare/v0.1.0...v0.2.0`, and the existing
      `[0.1.0]: …/releases/tag/v0.1.0` is unchanged. All use the
      `jysf/bragfile000` slug (see Slug decision below).
- [ ] **AC10 — `test-docs.sh` Group O green.** Running
      `bash scripts/test-docs.sh` reports O1–O5 pass (file exists;
      `keepachangelog.com` present; `^## \[0\.1\.0\]` heading present; all
      ten `[0.1.0]` command verbs still in backticks; `[Unreleased]:` and
      `[0.1.0]:` link refs present).

### Slug decision (AC9 rationale)

The compare-links keep the **`jysf/bragfile000`** slug — they are left
as-is. `git remote -v` confirms `origin` is
`https://github.com/jysf/bragfile000.git`; compare-links are clickable
GitHub URLs that must resolve to the **actual source repo**, which is
`bragfile000`. The names `jysf/bragfile/bragfile` (Homebrew install) and
`jysf/homebrew-bragfile` (tap) are the *distribution/product* identity,
not the source-repo path, so they do not belong in these URLs. The
existing `[0.1.0]` links are already correct for the real remote; the
only changes are adding `[0.2.0]` and repointing `[Unreleased]`.

### Date decision (AC1)

The entry is **dated `2026-06-12`** (today). Keep-a-Changelog dates an
entry when its release is cut; SPEC-037 cuts `v0.2.0` next and
same-day, so a date is correct rather than leaving it undated. If
SPEC-037 slips to a later calendar day, that spec owns bumping this
date — noted under Notes for the Implementer.

## Failing Tests

This is a docs spec — there are **no Go failing tests** (no code under
test). The verification surface is:

- **`scripts/test-docs.sh`** (Group O, lines 712–755)
  - `O1` — `CHANGELOG.md` exists. (Already green; edit must not delete it.)
  - `O2` — contains `keepachangelog.com`. (Header preserved.)
  - `O3` — a line matches `^## \[0\.1\.0\]`. Satisfied by
    `## [0.1.0] - 2026-05-10`.
  - `O4` — each of the ten verbs `brag add|list|show|edit|delete|search|
    export|summary|review|stats` appears in backticks. All remain in the
    `[0.1.0]` section, untouched.
  - `O5` — `[Unreleased]:` and `[0.1.0]:` link refs both present.
    Satisfied: `[Unreleased]:` is repointed (still present) and `[0.1.0]:`
    is unchanged.
  - These five must stay GREEN after the edit; they are the spec's
    regression gate. (No new shell assertion is added — `[0.2.0]`-specific
    coverage rides on the grep-checkable ACs above.)

## Implementation Context

*Read this section before starting build. The CHANGELOG edit is a
literal artifact — transcribe the post-edit file verbatim. Build adds
nothing creative; verify diffs the working tree against the embedded
literal.*

### The edit, as three reconciliations + one new section

For reviewer clarity, the change is four moves against the current file:

1. **Empty `[Unreleased]`.** Remove its `### Added` + `brag completion`
   body; keep the `## [Unreleased]` heading with nothing under it.
2. **Insert `## [0.2.0] - 2026-06-12`** between `[Unreleased]` and
   `[0.1.0]` (the new section, below).
3. **`[0.1.0]` date:** `YYYY-MM-DD` → `2026-05-10` (one-line change to the
   `## [0.1.0]` heading; the rest of the `[0.1.0]` body is untouched).
4. **Compare-links:** repoint `[Unreleased]` to `v0.2.0...HEAD`, add a
   `[0.2.0]` line, keep `[0.1.0]`.

### Post-edit CHANGELOG.md — embedded literal (transcribe verbatim)

The full file after the edit. The `[0.1.0]` Added + Decisions-of-record
body (lines between its heading and the link refs) is **unchanged from
the current file** and is shown collapsed as `‹[0.1.0] Added +
Decisions-of-record body — unchanged›`; do not retype or reflow it,
leave those existing lines exactly as they are. Only the `[0.1.0]`
**heading line** changes (date).

````markdown
# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.2.0] - 2026-06-12

This release makes **tags** and **projects** first-class. Tags move from
a comma-joined string to a normalized, shared, rename/merge-able model;
projects become a managed entity with filesystem locations and cwd-aware
auto-fill. Schema migrations now snapshot your database before they run.

### Added

- `brag tags` — list every tag with its usage count.
- `brag tag rename <old> <new>` and `brag tag merge <src> <dst>` —
  first-class tag maintenance. `rename` re-labels a tag in place;
  `merge` folds one tag's entries into another and de-duplicates.
- `brag project` — manage named projects backed by filesystem paths,
  with subcommands `new`, `list`, `show`, `edit`, `archive`, `delete`,
  `status`, and `here`. `brag project here` reports the project owning
  the current directory; `brag project status` prints a per-project
  dashboard.
- `brag project edit` takes `--add-path` / `--remove-path` to attach or
  detach directories from a project.
- `brag add` now auto-fills `--project` from the current directory when
  the cwd sits under a registered project location (nearest-ancestor
  match). An explicit `--project` always wins.
- `brag completion <shell>` — generate tab-completion scripts for zsh,
  bash, and fish. Source into your shell rc for `brag <tab>` and flag
  completion.

### Changed

- **Tags are now first-class.** They are stored in a normalized
  `tags` + `taggings` model instead of a comma-joined string, so a tag
  is shared across entries and can be renamed or merged. Existing
  entries migrate automatically on first run; the `--tag` filter and
  every entry command behave the same for users.
- **Schema migrations back up your database first.** Applying a
  schema-bumping migration to an existing, non-empty database now writes
  a timestamped snapshot beside it (via SQLite `VACUUM INTO`, WAL-safe)
  before the migration runs — so an upgrade can never mutate an
  un-backed-up database. If the backup fails, the upgrade aborts rather
  than proceeding. Non-interactive: safe in `brag add --json` and other
  piped, non-TTY workflows.

### Decisions of record

The following architectural decisions are committed in this release.
Each decision file under `/decisions/` carries the full rationale.

- DEC-015 — normalize tags into a polymorphic `tags` + `taggings`
  model (supersedes DEC-004's comma-joined string).
- DEC-016 — tag mutation semantics: `rename` errors into an existing
  tag, `merge` de-dups via DELETE+INSERT, orphaned tags are invisible
  (no garbage collection).
- DEC-017 — `entries.project` relates to `projects` by soft string
  match (project stays free text on the entry; no hard foreign key).
- DEC-018 — `brag project delete` blast radius: what a delete removes
  and what it leaves behind.
- DEC-019 — `brag project here` resolves the cwd by nearest-ancestor
  (longest-prefix) matching.
- DEC-020 — `brag project edit` location-editing semantics
  (`--add-path` / `--remove-path`).
- DEC-021 — migration auto-backup durability model: trigger on
  pending-migration-meets-non-empty-DB, snapshot via `VACUUM INTO`,
  abort `storage.Open` if the backup fails.

## [0.1.0] - 2026-05-10

‹[0.1.0] Added + Decisions-of-record body — unchanged›

[Unreleased]: https://github.com/jysf/bragfile000/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/jysf/bragfile000/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/jysf/bragfile000/releases/tag/v0.1.0
````

### The two changed `[0.1.0]`-region lines (exact)

To make the "unchanged body" boundary unambiguous, the only edits inside
the existing `[0.1.0]` region are:

- **Heading:** `## [0.1.0] - YYYY-MM-DD` → `## [0.1.0] - 2026-05-10`
- **Link refs block** (the last two lines of the file) is replaced by the
  three-line block shown above (`[Unreleased]` repointed, `[0.2.0]`
  added, `[0.1.0]` unchanged).

### Decisions that apply

- `DEC-015` — tags normalization; the headline "Changed" item.
- `DEC-016` — tag mutation semantics; backs `brag tag rename` / `merge`.
- `DEC-017`..`DEC-020` — the projects entity, delete blast radius, `here`
  resolution, and location-editing semantics; back the `brag project`
  Added block.
- `DEC-021` — migration auto-backup durability model; backs the safety-
  belt "Changed" item. SPEC-036 shipped the implementation.

### Constraints that apply

None of `/guidance/constraints.yaml` touches `CHANGELOG.md` (all are
scoped to `internal/**`, `cmd/**`, `go.mod`, or migration paths). No
blocking constraint applies to this docs-only edit. `test-before-
implementation` is satisfied vacuously — the regression gate is the
pre-existing `test-docs.sh` Group O, not a new failing test.

### Prior related work

- `SPEC-034` (shipped 2026-06-12) — comprehensive doc sweep
  (tutorial/architecture/api-contract); set `docs/tutorial.md:664`
  "shipped in v0.2.0", which this entry is consistent with.
- `SPEC-036` (shipped 2026-06-12) — the migration auto-backup safety
  belt this entry documents; emitted DEC-021.
- `SPEC-037` (pending) — cuts the actual `v0.2.0` tag/release that
  references this entry. Owns the date if the release slips past
  2026-06-12.

### Out of scope (for this spec specifically)

- The actual git tag and GitHub release (`v0.2.0-rc1` → `v0.2.0`),
  goreleaser, and the Homebrew formula bump — all **SPEC-037**.
- Any code change. This spec touches `CHANGELOG.md` only.
- `docs/blog/README.md:51`'s "v0.1.0 history" phrasing — a blog-asset
  follow-up, not gated and not a blocker (see Premise audit).
- Adding a new `test-docs.sh` assertion for `[0.2.0]`-specific content —
  the grep-checkable ACs cover it; Group O stays the regression gate.

## Notes for the Implementer

- **Transcribe, don't reflow.** The `[0.2.0]` block above is the locked
  artifact. Leave the `[0.1.0]` Added + Decisions-of-record body byte-for-
  byte as it already is in the file; only its heading-line date changes.
- **Keep `## [Unreleased]` as an empty section.** Keep-a-Changelog wants a
  standing Unreleased section, and Group O's O5 requires the
  `[Unreleased]:` link reference to survive — so do not delete the
  heading or the link, just clear the body.
- **Slug is intentional.** Do not "correct" `bragfile000` to `bragfile`;
  it matches the real `origin` remote and is required for the compare-
  links to resolve. (Rationale under Slug decision.)
- **Date ownership.** If SPEC-037 cuts the release on a different calendar
  day, that spec bumps `## [0.2.0] - 2026-06-12` to the cut date; this
  spec writes today's date as the best current value.
- **Verify quickly:** `bash scripts/test-docs.sh` and confirm Group O
  (O1–O5) all pass; then eyeball the diff against the embedded literal.

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
