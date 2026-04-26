---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-021
  type: chore                      # epic | story | task | bug | chore
  cycle: design                    # frame | design | build | verify | ship
  blocked: false
  priority: medium
  complexity: M                    # S | M | L  (L means split it)
                                   # M (not S): doc restructure across
                                   # 3 new/rewritten files + multiple
                                   # cross-link audits + a new shell
                                   # test harness. Heavier than S, well
                                   # under L's "split it" threshold.

project:
  id: PROJ-001
  stage: STAGE-005
repo:
  id: bragfile

agents:
  architect: claude-opus-4-7
  implementer: claude-opus-4-7     # usually same Claude, different session
  created_at: 2026-04-25

references:
  decisions: []                    # No DECs referenced; no DECs emitted.
  constraints: []                  # No code-layer constraints touch
                                   # this spec; doc-restructure is
                                   # outside `no-sql-in-cli-layer` /
                                   # `storage-tests-use-tempdir` /
                                   # `stdout-is-for-data-stderr-is-for-humans`
                                   # / `test-before-implementation` —
                                   # those govern Go code. The Failing
                                   # Tests below ARE written before the
                                   # build cycle in spirit of TDD, but
                                   # the artefact is shell + markdown
                                   # not Go.
  related_specs:
    - SPEC-018                     # audit-grep cross-check addendum
    - SPEC-019                     # NOT-contains self-audit pattern
    - SPEC-020                     # negative-substring self-audit codified
    - SPEC-022                     # successor: schema + Claude hook
                                   #   (will inherit GIF/screenshot work)
    - SPEC-023                     # successor: goreleaser + CI + tap
                                   #   (will inherit deferred
                                   #   architecture.md / api-contract.md
                                   #   / data-model.md stale stage refs;
                                   #   will activate brew-install &
                                   #   CHANGELOG forward-references; will
                                   #   land badges if any)
---

# SPEC-021: readme user-facing rewrite and dev-process migration

## Context

`README.md` today opens with the spec-driven development *process* —
the Repo→Project→Stage→Spec hierarchy, the Frame→Design→Build→Verify→Ship
cycle, the "four habits" of session hygiene. None of that answers the
questions a stranger lands on the repo with after `brew install
jysf/bragfile/bragfile` (or after clicking through from a GitHub link):
**what is `brag`, why would I use it, how do I install it, how do I
capture my first entry, where do my entries live?** External Claude
review on 2026-04-24 surfaced this misframing; the user agreed; STAGE-005
opened with the rewrite as its first workstream.

This spec is **STAGE-005's first workstream** (per the stage's spec
backlog ordering — lowest risk, no dependencies, sets the user-facing
framing the later specs reference). It does two things:

1. **Rewrite `README.md` as user-facing.** The new README answers
   "what does `brag` do, install, capture, list/search, export,
   summary, review, stats, where data lives, where to go next" within
   ~150–200 lines. No mention of cycles, stages, specs, contributor
   onboarding, or session hygiene above the fold.
2. **Migrate the spec-driven dev-process content** out of `README.md`
   to **two new files**: a thin `CONTRIBUTING.md` at repo root (GitHub
   convention; surfaces in PR-create UI) pointing at a deeper
   `docs/development.md` (full framework details, four habits, daily
   commands, glossary cross-link to `AGENTS.md §11`).

The brief explicitly calls out a workflow demo shape — "install,
capture, list/search, export, summary, review, stats" — which is
exactly the user-facing surface the README needs to demonstrate. STAGE-005's
later specs (SPEC-022 schema + Claude hook, SPEC-023 goreleaser + CI +
tap, SPEC-024 shell completions) will each do their own README/docs
sweeps when their features ship; SPEC-021's job is to set the shape
those sweeps fold into.

Parents:
- Project: `projects/PROJ-001-mvp/` (PROJ-001 — MVP wave; closes when
  STAGE-005 ships).
- Stage: `projects/PROJ-001-mvp/stages/STAGE-005-distribution-and-cleanup.md`
  (the stage; SPEC-021 is its first spec). Stage notes lock the "above
  the fold" structure (`STAGE-005:Design Notes:SPEC-021-specific`,
  ~lines 323–346) and flag SPEC-021 as the highest premise-audit hot
  spot in the stage.
- Repo: `bragfile`.

Stage-level locked decisions that bind this spec:
- **No new DECs expected.** Doc structure is configuration, not a
  project-binding decision. (`STAGE-005:Design Notes:Cross-cutting`,
  line ~252.)
- **Trim heuristic does NOT apply.** STAGE-005's workstreams are
  format-distinct; SPEC-021 is the first spec in the stage so there is
  no in-stage construction precedent to trim against. Default to fuller
  skeleton. (`STAGE-005:Design Notes:Cross-cutting`, line ~261.)
- **Premise audit IS load-bearing here.** SPEC-021 is the only
  STAGE-005 spec that's a status-change at scale; the audit-family
  rules apply directly. (`STAGE-005:Design Notes:Cross-cutting`, line
  ~275.)
- **`CONTRIBUTING.md` recommended over `docs/development.md` alone**
  because GitHub-conventional. SPEC-021 picks the split between them.
  (`STAGE-005:Design Notes:SPEC-021-specific`, line ~330.)

External Claude review (2026-04-24) seeded this spec; user concurred
2026-04-24; STAGE-005 framed 2026-04-25; SPEC-021 scaffolded
2026-04-25.

## Goal

Rewrite `README.md` as a ~150–200-line user-facing introduction to the
`brag` tool (workflow demo: install → capture → list/search → export →
summary/review/stats → where data lives → where to go next), and
migrate the spec-driven dev-process content to a new thin
`CONTRIBUTING.md` (GitHub-conventional, points deeper) and
`docs/development.md` (full framework details), with internal-link
audits across all repo-root and `docs/` markdown verified by a new
`scripts/test-docs.sh` harness exposed via `just test-docs`.

## Inputs

**Files to read (build-cycle reading list):**
- `AGENTS.md` — especially §6 (Cycle Model), §7 (Cross-Reference
  Rules), §9 (Testing Conventions, particularly the four
  premise-audit addenda), §10 (Git/PR Conventions, particularly the
  newly-codified push-discipline rule), §11 (Domain Glossary, source
  of truth for any framework term referenced), §12 (Cycle-Specific
  Rules, particularly the design-time decision rule and the
  NOT-contains self-audit). Source of truth for everything migrating
  out of README.
- `README.md` — current state. The 117-line file documents the
  development *process* rather than the tool. Lines that need
  migrating: 10 (spec-driven workflow blurb), 12–20 (Hierarchy
  diagram including the Frame→Design→Build→Verify→Ship cycle on line
  19), 22–35 (Getting started section — currently points at
  `GETTING_STARTED.md` and lists `just status` / `just new-spec` /
  `just advance-cycle` / `just archive-spec` / `just weekly-review`),
  37–46 (Key discipline / four habits), 87–96 (Daily commands for
  working on `brag` itself), and the contributor rows of the
  Where-things-live table (98–112).
  Lines to keep + restructure: 1–9 (title/tagline/status), 48–64 (The
  app itself paragraph), 65–79 (Install locally), 80–86 (Tutorial
  pointer), 114–117 (License).
- `BRAG.md` — repo-root file targeting AI agents. Stays untouched
  by this spec (SPEC-022 owns the schema cross-reference). README
  links to it as the "AI integration" pointer.
- `docs/tutorial.md` — current user-facing tutorial. ~510 lines;
  the README is the 30-second elevator pitch and tutorial.md is the
  deeper dive. README must reference tutorial.md but not duplicate it.
- `docs/api-contract.md`, `docs/architecture.md`, `docs/data-model.md`
  — current docs. Read to confirm none of them house dev-process
  content that's also in README. (Confirmed at design time: they
  describe the tool's surface and code-path lineage, not the
  spec-driven framework. They DO contain `STAGE-NNN` tense
  references that are stale; those are deferred to SPEC-023's
  doc-sweep — see Premise audit § "Known stale stage refs deferred
  to later STAGE-005 specs".)
- `GETTING_STARTED.md` — ~190-line walkthrough for a contributor's
  first project in a fresh spec-driven repo. Stays at repo root
  unchanged (locked rejection #3).
- `FIRST_SESSION_PROMPTS.md` — phase prompts. Stays at repo root
  unchanged (locked rejection #3).
- `projects/PROJ-001-mvp/brief.md` — project context. STAGE-005
  scope notes for SPEC-021 (line ~118 + ~197 + ~278). Confirms the
  "install, capture, list/search, export, summary, review, stats"
  workflow demo shape.
- `projects/PROJ-001-mvp/stages/STAGE-005-distribution-and-cleanup.md`
  — parent stage. Locks the structural recommendations and the
  audit-grep enumeration (~lines 323–346 for SPEC-021-specific notes).
- `projects/PROJ-001-mvp/specs/done/SPEC-018-brag-summary-aggregate-package-and-shape-dec.md`
  — origin of the §9 audit-grep cross-check addendum. SPEC-021 is
  the most premise-audit-heavy doc-restructure spec in PROJ-001 and
  the cross-check earns its keep here.
- `projects/PROJ-001-mvp/specs/done/SPEC-019-brag-review-week-and-month-flags.md`
  — first run of the NOT-contains self-audit pattern (build-time
  self-catch on its own Long string).
- `projects/PROJ-001-mvp/specs/done/SPEC-020-brag-stats-six-lifetime-metrics.md`
  — design-time pre-emption of the NOT-contains pattern; eight
  forbidden tokens grep'd against locked Long with zero hits at
  design and zero false-positives at build. SPEC-021 applies the
  same self-audit (see § NOT-contains self-audit below).
- `scripts/archive-spec.sh` — newly extended (commit `bfa1474`,
  2026-04-25) to reject empty `<answer>` placeholders in Reflection
  (Ship). At ship time, SPEC-021's Reflection (Ship) section MUST be
  filled with real answers or `just archive-spec SPEC-021` will
  fail. Surfaced to the build session in Notes for the Implementer.
- `justfile` — current shape. SPEC-021 adds one new recipe
  (`test-docs`) without changing existing recipes; the `test` recipe
  stays Go-only. (Q4(a) answer 2026-04-25.)
- `scripts/` directory — flat shape (`status.sh`, `new-spec.sh`,
  `archive-spec.sh`, etc.). SPEC-021's new shell test goes at
  `scripts/test-docs.sh` (flat, matching established convention).
  No `scripts/test/` subdirectory. (Q4(b) answer 2026-04-25.)

**External APIs:** None.

**Related code paths:** None — this spec touches docs and a new shell
script only. No Go code changes.

## Outputs

### Files created

- **`CONTRIBUTING.md`** (repo root, NEW) — thin GitHub-conventional
  contributor entry-point (~50–70 lines). Contains:
  - One-paragraph welcome (this is a personal-tool project; the user
    is the primary contributor; PRs welcomed but not actively sought).
  - Development setup (`just install`, `just test`).
  - Brief description of the spec-driven workflow (one paragraph)
    pointing at `docs/development.md` for full details.
  - PR conventions (branch naming, commit style) cross-linking
    `AGENTS.md §10`.
  - License attribution.
  Surfaced by GitHub in the PR-create UI; that's why it lives at repo
  root rather than under `docs/`.

- **`docs/development.md`** (NEW) — full spec-driven framework
  details for contributors (~80–120 lines). Contains:
  - The Repo→Project→Stage→Spec hierarchy (the diagram migrated from
    `README.md:12–20`).
  - The Frame→Design→Build→Verify→Ship cycle (migrated from
    `README.md:19`).
  - The four habits of session hygiene (migrated from
    `README.md:37–46`).
  - Daily commands for working on `brag` itself (migrated from
    `README.md:87–96`).
  - Pointers to `AGENTS.md` (full conventions),
    `GETTING_STARTED.md` (first-project walkthrough),
    `FIRST_SESSION_PROMPTS.md` (phase prompts).
  - Cross-link to `AGENTS.md §11 Domain Glossary` (NOT a duplicated
    glossary — chain: README → CONTRIBUTING → development.md →
    AGENTS.md §11). (Q3 answer 2026-04-25.)

- **`scripts/test-docs.sh`** (NEW) — POSIX-compatible shell script
  that runs the doc-content assertions (positive-contains and
  not-contains greps) against the rewritten README. Not wired into
  `just test`; exposed only via `just test-docs`. Single script that
  grows internally as later STAGE-005 specs add doc-content asserts.
  (Q4(a)+(b) answers 2026-04-25.)

### Files modified

- **`README.md`** — fully rewritten as a ~150–200-line user-facing
  introduction. Workflow demo: install → capture → list/search →
  export → summary/review/stats → where data lives → where to go
  next. No mention of cycles, stages (in dev-process tense; the
  stage-tense reference at current line 79 — "Homebrew install
  arrives in STAGE-005" — gets replaced with a forward-reference to
  brew install that is structurally complete and functionally
  pending until SPEC-023 ships), specs, or session hygiene. See
  Notes for the Implementer § "README sketch" for the literal
  shape.

- **`justfile`** — one new recipe added under the daily-commands
  section:
  ```
  # Run doc-content assertions (separate from `just test`, which is
  # Go-only)
  test-docs:
      @./scripts/test-docs.sh
  ```
  No other recipes change.

### Files NOT modified by this spec (deferred)

- `BRAG.md` — schema cross-reference is SPEC-022's scope.
- `docs/tutorial.md`, `docs/api-contract.md`, `docs/architecture.md`,
  `docs/data-model.md` — STAGE-NNN tense references are stale (see
  Premise audit) but describe code-path lineage, not user-facing
  status claims that the README rewrite invalidates. Deferred to
  SPEC-023's doc-sweep with explicit punch-list enumeration so that
  sweep is mechanical.
- `AGENTS.md` — source of truth for all framework content; no
  changes. The new `docs/development.md` cross-links into it.
- `GETTING_STARTED.md`, `FIRST_SESSION_PROMPTS.md` — locked
  rejection #3 (stay at repo root unchanged). Inbound references
  to README from these files (FIRST_SESSION_PROMPTS.md:63,
  FIRST_SESSION_PROMPTS.md:227, GETTING_STARTED.md:40,
  GETTING_STARTED.md:177) reference the file's existence as a
  framework artifact, not its content; they don't need updating.
- `docs/CONTEXTCORE_ALIGNMENT.md` — its line 31 mention of the
  Frame→Design→Build→Verify→Ship cycle is framework-alignment
  content, the right home for it; stays.
- `LICENSE` — untouched.
- `.gitignore`, `go.mod`, `go.sum` — untouched.
- `cmd/`, `internal/` — Go code untouched.

### New exports / database changes

None.

## Acceptance Criteria

Testable outcomes. Cover positive structure (what the new shape MUST
contain), negative structure (what the new shape MUST NOT contain after
migration), and link integrity (no broken internal references).

### A. README shape (positive)

- [ ] **A1.** New `README.md` exists and is between 100 and 250 lines
      inclusive (target band 150–200; ±50 for tonal latitude).
- [ ] **A2.** New `README.md` opens with a `# Bragfile` (or
      equivalent H1 — the project's display name) heading on line 1
      or 2.
- [ ] **A3.** The first non-heading paragraph of `README.md` (the
      "above-the-fold" tagline) describes the tool in user terms:
      mentions "brag" (the binary) and at least one of {capture,
      retrieve, accomplishments, retros, reviews, resumes}. Does
      NOT contain {`spec-driven`, `architect`, `implementer`,
      `reviewer`, `cycle`, `hierarchy`} (case-insensitive).
- [ ] **A4.** New `README.md` contains an H2 install section
      (heading matching `^## .*[Ii]nstall`) with both:
      - a `brew install jysf/bragfile/bragfile` line
        (forward-reference, structurally complete; SPEC-023
        activates), AND
      - a `go install ./cmd/brag` (or `git clone … && just install`)
        line (working today).
- [ ] **A5.** New `README.md` demonstrates the workflow in this
      order, each command appearing as a fenced shell example with at
      least one concrete invocation:
      1. `brag add` (capture)
      2. `brag list` (read back)
      3. `brag search` (full-text)
      4. `brag export --format markdown` (review-ready dump)
      5. `brag summary --range week` or `--range month` (rule-based
         digest)
      6. `brag review --week` or `--month` (reflection)
      7. `brag stats` (lifetime metrics)
      Order may vary slightly (e.g., `list` and `search` adjacent),
      but all seven appear.
- [ ] **A6.** New `README.md` contains a "where data lives"
      reference to `~/.bragfile/db.sqlite` (the default DB path).
- [ ] **A7.** New `README.md` contains a deeper-dive pointer to
      `docs/tutorial.md` (link literal: `docs/tutorial.md` or
      `[…](docs/tutorial.md)`).
- [ ] **A8.** New `README.md` contains an AI-integration pointer to
      `BRAG.md` (link literal: `BRAG.md` or `[…](BRAG.md)`).
- [ ] **A9.** New `README.md` contains a contributor pointer to
      `CONTRIBUTING.md` (link literal: `CONTRIBUTING.md` or
      `[…](CONTRIBUTING.md)`).
- [ ] **A10.** New `README.md` ends with a license section (heading
      matching `^## [Ll]icense` or equivalent) referencing `LICENSE`
      and naming `MIT`.

### B. README shape (negative — migrated content absent)

- [ ] **B1.** New `README.md` does NOT contain the literal token
      `spec-driven` (case-insensitive). [The dev-process narrative
      lives in `docs/development.md`.]
- [ ] **B2.** New `README.md` does NOT contain the cycle phrase
      `Frame → Design → Build → Verify → Ship` nor lowercase
      `frame → design → build → verify → ship` (either arrow form,
      either case). [Cycle model lives in `docs/development.md` and
      `AGENTS.md §6`.]
- [ ] **B3.** New `README.md` does NOT contain the phrase
      `four habits` (case-insensitive). [Session hygiene lives in
      `docs/development.md` and `AGENTS.md §13`.]
- [ ] **B4.** New `README.md` does NOT contain the phrase
      `context contamination` (case-insensitive). [Same.]
- [ ] **B5.** New `README.md` does NOT contain any of these
      contributor-shaped just-recipe references:
      `just new-spec`, `just advance-cycle`, `just archive-spec`,
      `just weekly-review`, `just new-stage`. [Daily commands for
      working on `brag` itself live in `docs/development.md`.]
- [ ] **B6.** New `README.md` does NOT contain the phrase
      `Claude plays every role` (case-insensitive). [Variant
      framing lives in `AGENTS.md §1` and is contributor-only.]
- [ ] **B7.** New `README.md` does NOT contain a top-level
      table-of-contents block (no heading matching
      `^## .*[Tt]able of [Cc]ontents` and no contiguous block of 4+
      lines each starting with `- [` near the top of the file).
      [Locked rejection #5 — GitHub renders auto-TOC; manually
      maintained TOCs drift.]

### C. CONTRIBUTING.md shape

- [ ] **C1.** `CONTRIBUTING.md` exists at repo root.
- [ ] **C2.** `CONTRIBUTING.md` is between 30 and 120 lines inclusive
      (target ~50–70; thin pointer file, not a treatise).
- [ ] **C3.** `CONTRIBUTING.md` mentions `docs/development.md` (the
      deeper-dive pointer) at least once.
- [ ] **C4.** `CONTRIBUTING.md` mentions `AGENTS.md` (the source of
      truth) at least once.
- [ ] **C5.** `CONTRIBUTING.md` contains development setup commands
      (at minimum: `just install` AND `just test`).

### D. docs/development.md shape

- [ ] **D1.** `docs/development.md` exists.
- [ ] **D2.** `docs/development.md` is between 50 and 200 lines
      inclusive (target ~80–120).
- [ ] **D3.** `docs/development.md` contains the
      Repo→Project→Stage→Spec hierarchy (matches the substring
      `Repo` AND `Project` AND `Stage` AND `Spec` somewhere in the
      file, in any rendering — code block tree or prose).
- [ ] **D4.** `docs/development.md` contains the cycle phrase
      `Frame → Design → Build → Verify → Ship` (preserves the
      Unicode arrow form for searchability with the AGENTS.md
      source).
- [ ] **D5.** `docs/development.md` mentions `AGENTS.md` (or links
      to it) at least once. The "full conventions" pointer.
- [ ] **D6.** `docs/development.md` mentions `GETTING_STARTED.md`
      (or links to it) at least once. Pointer to the
      first-project walkthrough.
- [ ] **D7.** `docs/development.md` mentions
      `FIRST_SESSION_PROMPTS.md` (or links to it) at least once.
      Pointer to the phase prompts.
- [ ] **D8.** `docs/development.md` cross-links to
      `AGENTS.md§11` or `AGENTS.md#11` or text near `AGENTS.md`
      that names "glossary" (the chain README → CONTRIBUTING →
      development.md → AGENTS.md §11; glossary is NOT duplicated
      here per Q3 answer).

### E. Link integrity

- [ ] **E1.** Every relative-path markdown link `](./foo)` and
      `](foo)` and `](docs/foo)` and `](../foo)` in the new
      `README.md`, `CONTRIBUTING.md`, and `docs/development.md`
      resolves to an existing file in the repo (link-resolution
      check; absolute URLs `https://…` are exempt). No broken
      internal links introduced by this spec.
- [ ] **E2.** No file in the repo references `docs/development.md`
      as a target before this spec lands EXCEPT as introduced by
      this spec (i.e., this spec's own outputs and the new README /
      CONTRIBUTING.md). Sanity check that no in-repo doc was already
      pointing at the new file before it existed.
- [ ] **E3.** `CONTRIBUTING.md` has not been replaced or co-opted
      by an earlier convention (sanity check: no
      `git log --all --diff-filter=D -- CONTRIBUTING.md` history of
      a deletion). It's a brand-new file.

### F. Just-recipe wiring

- [ ] **F1.** `justfile` defines a `test-docs` recipe that
      executes `./scripts/test-docs.sh`.
- [ ] **F2.** `just test` is unchanged in behavior — it still runs
      `go test ./...` only. (`just test-docs` is separate. Q4(a)
      lock.)
- [ ] **F3.** `scripts/test-docs.sh` is executable
      (`chmod +x scripts/test-docs.sh`) and has a POSIX shebang
      (`#!/usr/bin/env sh` or `#!/bin/sh` or `#!/usr/bin/env bash`).
- [ ] **F4.** Running `just test-docs` against the rewritten
      README + new CONTRIBUTING.md + new docs/development.md exits
      0. Running it before the rewrite (against the current
      README) would exit 1; the script's purpose is to gate on the
      new shape, not the old.

### G. The harness's own behavior

- [ ] **G1.** `scripts/test-docs.sh` prints `OK: <check name>`
      (or equivalent) on each passing assertion AND
      `FAIL: <check name>: <reason>` on each failing one.
- [ ] **G2.** `scripts/test-docs.sh` exits 0 iff all assertions
      pass; exits 1 if any fail. (Standard shell test harness
      contract.)
- [ ] **G3.** `scripts/test-docs.sh` runs from any working
      directory — paths inside it are relative to the script's own
      location (`SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)`),
      mirroring the established `scripts/_lib.sh` /
      `scripts/status.sh` pattern. Sanity:
      `cd /tmp && /full/path/to/scripts/test-docs.sh` works.

### H. No-op for unrelated surfaces

- [ ] **H1.** `go test ./...` still passes on the post-spec tree.
      (Sanity — this spec touches no Go.)
- [ ] **H2.** `gofmt -l .` produces no output. (Sanity.)
- [ ] **H3.** `BRAG.md` is byte-identical to its pre-spec state.
- [ ] **H4.** `AGENTS.md` is byte-identical to its pre-spec state.
- [ ] **H5.** `GETTING_STARTED.md` is byte-identical to its
      pre-spec state.
- [ ] **H6.** `FIRST_SESSION_PROMPTS.md` is byte-identical to its
      pre-spec state.
- [ ] **H7.** `docs/tutorial.md` is byte-identical to its pre-spec
      state. (Tutorial doc-sweeps for the digest trio shipped in
      STAGE-004 already; no further sweep here.)
- [ ] **H8.** `docs/api-contract.md`, `docs/architecture.md`,
      `docs/data-model.md` are byte-identical to their pre-spec
      states. (Stale STAGE-NNN refs in these files are deferred to
      SPEC-023's doc-sweep — see Premise audit.)

**Total:** 39 acceptance criteria across 8 groupings.

## Failing Tests

Written during **design**, BEFORE build. The implementer's job in
**build** is to make these pass.

Doc-restructure work has a different shape than prior PROJ-001 specs
(which added Go code). Most acceptance criteria are content-shape
assertions on markdown files, verifiable via grep / structure checks /
link validation. There are **zero new Go test functions** in this spec;
the test harness is a shell script (`scripts/test-docs.sh`) plus a
small set of byte-comparison and link-resolution checks built into the
same script.

The script implements one assertion per acceptance criterion in groups
A–G; group H is checked by manual `git diff` (criterion-level
verification — see Notes for the Implementer § "Verify-cycle checklist
for group H").

### `scripts/test-docs.sh` (NEW)

POSIX-shell test harness. Runs every assertion below, prints
`OK: <name>` per pass and `FAIL: <name>: <detail>` per fail, exits 0
iff all pass.

The script's structure: a series of `assert_*` helper functions
defined at the top, then a sequential list of calls implementing each
named assertion. See Notes for the Implementer § "test-docs.sh
sketch" for the literal scaffold.

#### Group A — README shape (positive)

- **A1 — README line count band**
  - Assertion: `wc -l < README.md` returns an integer N where
    `100 <= N <= 250`.
  - Failure mode: outside band → `FAIL: A1: README has N lines
    (expected 100–250)`.

- **A2 — README opens with H1**
  - Assertion: line 1 or line 2 of README matches `^# `.
  - Failure mode: no H1 in first 2 lines.

- **A3 — Above-the-fold is user-facing**
  - Assertion: the first non-heading paragraph of README (lines
    2–~10 typically) contains `brag` (case-insensitive) AND at
    least one of {`capture`, `retrieve`, `accomplishment`, `retro`,
    `review`, `resume`} (case-insensitive). AND it does NOT
    contain any of {`spec-driven`, `architect`, `implementer`,
    `reviewer`, `cycle`, `hierarchy`} (case-insensitive). Use
    `head -n 12 README.md` as the scope.
  - Failure mode: above-the-fold prose fails either inclusion or
    exclusion.

- **A4 — Install section with both paths**
  - Assertion: README contains a heading matching `^## .*[Ii]nstall`
    AND contains the literal `brew install jysf/bragfile/bragfile`
    AND contains either `go install ./cmd/brag` OR `just install`.
  - Failure mode: install heading absent or either install path
    absent.

- **A5 — Workflow-demo command coverage**
  - Assertion: README contains all seven commands as fenced shell
    examples (each appears at least once, prefixed by `brag` and
    inside a fenced block):
    1. `brag add`
    2. `brag list`
    3. `brag search`
    4. `brag export`
    5. `brag summary`
    6. `brag review`
    7. `brag stats`
  - Implementation: extract fenced code blocks (`awk '/^```/{f=!f;next} f'`
    or similar), then grep for each invocation.
  - Failure mode: any of the seven absent.

- **A6 — Where-data-lives reference**
  - Assertion: README contains the literal string
    `~/.bragfile/db.sqlite`.
  - Failure mode: absent.

- **A7 — Tutorial pointer**
  - Assertion: README contains a markdown link or path reference
    to `docs/tutorial.md`. `grep -F 'docs/tutorial.md' README.md`
    returns at least one hit.
  - Failure mode: absent.

- **A8 — BRAG.md pointer**
  - Assertion: README contains a reference to `BRAG.md`.
    `grep -F 'BRAG.md' README.md` returns at least one hit.
  - Failure mode: absent.

- **A9 — CONTRIBUTING.md pointer**
  - Assertion: README contains a reference to `CONTRIBUTING.md`.
    `grep -F 'CONTRIBUTING.md' README.md` returns at least one hit.
  - Failure mode: absent.

- **A10 — License section**
  - Assertion: README contains a heading matching
    `^## [Ll]icense` AND the literal `MIT`.
  - Failure mode: heading absent or `MIT` absent.

#### Group B — README shape (negative — load-bearing)

- **B1 — No `spec-driven` token**
  - Assertion: `grep -i 'spec-driven' README.md` returns no hits.
  - Failure mode: any hit.

- **B2 — No cycle phrase**
  - Assertion: `grep -i 'frame.*design.*build.*verify.*ship' README.md`
    returns no hits AND `grep -i 'frame → design' README.md` returns
    no hits AND `grep -i 'frame -> design' README.md` returns no
    hits. (Three forms covered: spaced regex, Unicode arrow, ASCII
    arrow.)
  - Failure mode: any hit.

- **B3 — No `four habits` phrase**
  - Assertion: `grep -i 'four habits' README.md` returns no hits.
  - Failure mode: any hit.

- **B4 — No `context contamination` phrase**
  - Assertion: `grep -i 'context contamination' README.md` returns
    no hits.
  - Failure mode: any hit.

- **B5 — No contributor-shaped just-recipe refs**
  - Assertion: `grep -E 'just (new-spec|advance-cycle|archive-spec|weekly-review|new-stage)' README.md`
    returns no hits.
  - Failure mode: any hit.

- **B6 — No `Claude plays every role` phrase**
  - Assertion: `grep -i 'claude plays every role' README.md`
    returns no hits.
  - Failure mode: any hit.

- **B7 — No top-level table of contents**
  - Assertion (heading shape): `grep -i '^## .*table of contents' README.md`
    returns no hits. AND
  - Assertion (block shape): the file does NOT contain 4+ contiguous
    lines each starting with `- [` (a TOC block) within the first
    50 lines. (Implement as `head -n 50 README.md | awk '...'`
    counter; trip if the counter ever reaches 4.)
  - Failure mode: either form present.

#### Group C — CONTRIBUTING.md shape

- **C1 — File exists**
  - Assertion: `test -f CONTRIBUTING.md`.
- **C2 — Line count band**
  - Assertion: `wc -l < CONTRIBUTING.md` returns N with
    `30 <= N <= 120`.
- **C3 — Mentions `docs/development.md`**
  - Assertion: `grep -F 'docs/development.md' CONTRIBUTING.md`
    returns ≥1 hit.
- **C4 — Mentions `AGENTS.md`**
  - Assertion: `grep -F 'AGENTS.md' CONTRIBUTING.md` returns ≥1 hit.
- **C5 — Setup commands**
  - Assertion: CONTRIBUTING.md contains `just install` AND
    `just test`.

#### Group D — docs/development.md shape

- **D1 — File exists**
  - Assertion: `test -f docs/development.md`.
- **D2 — Line count band**
  - Assertion: `wc -l < docs/development.md` returns N with
    `50 <= N <= 200`.
- **D3 — Hierarchy present**
  - Assertion: file contains all four substrings (any case): `Repo`,
    `Project`, `Stage`, `Spec`.
- **D4 — Cycle phrase present**
  - Assertion: file contains `Frame → Design → Build → Verify → Ship`
    (Unicode arrow form, exact substring; preserves searchability
    with `AGENTS.md:166`).
- **D5–D7 — Pointers present**
  - Assertion (D5): `grep -F 'AGENTS.md' docs/development.md` ≥1 hit.
  - Assertion (D6): `grep -F 'GETTING_STARTED.md' docs/development.md`
    ≥1 hit.
  - Assertion (D7): `grep -F 'FIRST_SESSION_PROMPTS.md' docs/development.md`
    ≥1 hit.
- **D8 — Glossary cross-link**
  - Assertion: file contains the substring `AGENTS.md` AND the
    substring `glossary` (case-insensitive) within ±5 lines of
    each other. (Lightweight proxy for "links to AGENTS.md §11
    glossary" without depending on a specific link form.)
    Implement: extract line numbers of `AGENTS.md` mentions and
    of `glossary` mentions; assert min absolute difference ≤ 5.

#### Group E — Link integrity

- **E1 — Internal links resolve**
  - Assertion: extract every markdown link of the form
    `]( <path-not-starting-with-http> )` from `README.md`,
    `CONTRIBUTING.md`, and `docs/development.md`. For each,
    resolve relative to the source file's directory; assert
    `test -e` on the resolved path. Skip any anchor (`#…`) and
    any absolute URL (`http://…`, `https://…`, `mailto:…`).
  - Implementation: use `grep -oE '\]\([^)]+\)'` to extract;
    strip `](` and `)`; split on `#`; if not URL or anchor,
    resolve and check.
  - Failure mode: any unresolved link → `FAIL: E1: <source>:<link> → <resolved>`.

- **E2 — `docs/development.md` is brand-new**
  - Assertion: ` ! grep -rn -F 'docs/development.md' . --include='*.md' --exclude-dir=projects --exclude-dir=node_modules --exclude-dir=.git` matches only the new files this spec creates (CONTRIBUTING.md, README.md, docs/development.md itself). I.e., no pre-existing doc was already pointing at it.
  - Tolerable hits: README.md, CONTRIBUTING.md, docs/development.md (self), this spec file (`projects/PROJ-001-mvp/specs/SPEC-021…md`). Exclude `projects/` from the grep to keep planning docs out of scope.
  - Failure mode: any other file references it.

- **E3 — `CONTRIBUTING.md` is brand-new**
  - Assertion: `git log --all --diff-filter=D -- CONTRIBUTING.md`
    returns no commits. (No prior deletion of a file with the same
    name, which would suggest a different convention earlier in
    the project.)
  - Failure mode: prior deletion exists; investigate before
    overwriting.

#### Group F — Just-recipe wiring

- **F1 — `test-docs` recipe defined**
  - Assertion: `just --summary` (or `just --list`) output contains
    `test-docs`. Alternatively: `grep '^test-docs:' justfile`
    returns ≥1 hit.
- **F2 — `just test` unchanged**
  - Assertion: `git diff --no-color HEAD -- justfile` does NOT
    show changes to the existing `test:` recipe (i.e., the lines
    after `^test:` and before the next blank line are unchanged
    from the pre-spec state). Implementation: extract the `test:`
    recipe block from both pre-spec and post-spec justfile and
    compare verbatim.
- **F3 — `scripts/test-docs.sh` is executable + POSIX-headed**
  - Assertion: `test -x scripts/test-docs.sh` AND `head -n 1
    scripts/test-docs.sh` matches `^#!/(usr/bin/env (sh|bash))|bin/sh`.
- **F4 — `just test-docs` exits 0 on the post-spec tree**
  - Assertion: this is meta — the test harness's own pass run
    proves it. Implementation: a final `echo "OK: F4: harness
    self-pass"` line at the end of `test-docs.sh` after all other
    assertions emit OK.

#### Group G — Harness ergonomics

- **G1 — Per-check OK/FAIL output**
  - Assertion: every `assert_*` helper prints exactly one line
    starting with `OK: ` or `FAIL: ` per call. Verified by code
    review at verify cycle (no automated check).
- **G2 — Exit code contract**
  - Assertion: harness sets a `FAIL_COUNT` counter, increments per
    fail, exits `1` if `FAIL_COUNT > 0` else `0`.
- **G3 — Works from any cwd**
  - Assertion: `cd /tmp && /absolute/path/scripts/test-docs.sh`
    exits 0. Implement by relying on `SCRIPT_DIR` resolution at the
    top of the script (mirrors `scripts/status.sh`'s pattern).

#### Group H — Verify-cycle checklist (NOT in script; manual)

These are no-op-on-unrelated-surfaces sanity checks. They live in the
spec's verify-cycle review, not in `test-docs.sh`, because comparing
byte-identity across a doc-restructure spec is more efficient as a
`git diff` review than as scripted asserts.

- **H1 — `go test ./...` passes** (spec touches no Go).
- **H2 — `gofmt -l .` empty** (spec touches no Go).
- **H3 — `BRAG.md` byte-identical**: `git diff HEAD -- BRAG.md` empty.
- **H4 — `AGENTS.md` byte-identical**: `git diff HEAD -- AGENTS.md` empty.
- **H5 — `GETTING_STARTED.md` byte-identical**.
- **H6 — `FIRST_SESSION_PROMPTS.md` byte-identical**.
- **H7 — `docs/tutorial.md` byte-identical**.
- **H8 — `docs/{api-contract,architecture,data-model}.md` byte-identical**.

### Test count summary

- **Group A:** 10 assertions.
- **Group B:** 7 assertions.
- **Group C:** 5 assertions.
- **Group D:** 8 assertions.
- **Group E:** 3 assertions.
- **Group F:** 4 assertions.
- **Group G:** 3 assertions (G1 is review-only).
- **Group H:** 8 manual verify-cycle checks (NOT in script).

**Scripted total: 40 assertions across `scripts/test-docs.sh`. Manual
verify-cycle total: 8 byte-identity checks. Go test count: 0.**

The 1:1 mapping between Acceptance Criteria (39) and scripted
assertions (40) holds with one extra (B7's two-form check counts as
one criterion, two assertions inside the script).

## Implementation Context

*Read this section (and the files it points to) before starting the
build cycle. It is the equivalent of a handoff document, folded into
the spec since there is no separate receiving agent.*

### Decisions that apply

- **None new.** No DEC is emitted by this spec. Doc structure is
  configuration, not project-binding (per `STAGE-005:Design Notes`,
  line ~252).
- **DEC-001 through DEC-014** all apply to the broader project and
  inform what the README *describes* (e.g., DEC-003 config
  resolution shapes the "where data lives" section's
  `BRAGFILE_DB`/`--db` precedence story; DEC-004 tags-comma-joined
  shapes the capture examples; DEC-011 JSON shape and DEC-012 stdin
  schema shape the AI-integration pointer; DEC-014 envelope shapes
  the digest-trio examples), but no DEC is *referenced* in the
  rewritten README — the README is user-facing and DEC-* live below
  the surface in `docs/api-contract.md` and `decisions/`.

### Constraints that apply

The blocking constraint surface for SPEC-021 is essentially empty
because the spec touches no Go code:

- `no-sql-in-cli-layer` — N/A (no Go code).
- `storage-tests-use-tempdir` — N/A.
- `stdout-is-for-data-stderr-is-for-humans` — N/A; `scripts/test-docs.sh`
  is a developer test harness, not a CLI under the contract. (Its
  output is for humans; that's fine — the constraint applies to the
  shipped `brag` binary's output channels, not to test scripts.)
- `test-before-implementation` — applies *in spirit*: the assertions
  in `scripts/test-docs.sh` are written before the rewritten markdown
  exists. Concretely, the build-cycle order is: (1) write
  `scripts/test-docs.sh` with all 40 assertions; (2) run it against
  the *current* tree → expect a large pile of FAILs (most of group A
  fails because the current README doesn't have the install heading
  in the prescribed form, etc.); (3) write
  `docs/development.md` and `CONTRIBUTING.md`; (4) rewrite
  `README.md`; (5) re-run → expect all OK. (See Notes for the
  Implementer § "Build-cycle order" for the full sequence.)
- `one-spec-per-pr` — applies as always.
- `no-new-top-level-deps-without-decision` — N/A (spec adds no Go
  deps).

### Prior related work

- **SPEC-018** (shipped 2026-04-25) — origin of the audit-grep
  cross-check addendum (`AGENTS.md §9`, line ~225). SPEC-021 is the
  most premise-audit-heavy doc-restructure spec in PROJ-001 and the
  cross-check earns its keep here. **The build session must re-run
  the design-time greps below and reconcile actual hits against this
  spec's enumerated `## Outputs`.** Treat any delta as a question for
  the spec author (raise in Build Completion reflection), not as a
  unilateral expansion of scope.
- **SPEC-019** (shipped 2026-04-25) — first run of the
  NOT-contains self-audit pattern (build-time self-catch). SPEC-021
  has 7 NOT-contains assertions in group B; the self-audit check is
  performed below in § NOT-contains self-audit.
- **SPEC-020** (shipped 2026-04-25) — design-time pre-emption of
  the NOT-contains pattern. SPEC-021 mirrors the discipline: grep
  the spec's own load-bearing prose (the README sketch in Notes for
  the Implementer) for the seven forbidden tokens at design time.
  Result captured below.
- **SPEC-018 ship reflection** — codified the "routine STAGE-00N
  grep at stage start" carry-forward. SPEC-021's premise audit
  closes the loop on this carry-forward by enumerating the deferred
  STAGE-NNN tense refs in `docs/architecture.md`,
  `docs/api-contract.md`, `docs/data-model.md` as a SPEC-023 punch
  list (so SPEC-023's design doesn't re-grep).

### Out of scope (for this spec specifically)

Explicit list of what this spec does NOT include. If any of these feel
necessary during build, create a new spec rather than expanding this
one.

- **`BRAG.md` schema cross-reference.** Owned by SPEC-022.
- **GIF/screenshot/demo asset.** Defer. SPEC-022 ships the Claude
  hook artifact whose flow the GIF would demonstrate; SPEC-022
  (or a follow-up) owns the asset. SPEC-021's README has no GIF.
  No `![demo](demo.gif)` placeholder either — a placeholder for an
  asset that doesn't exist yet is a broken-link risk; activate when
  the asset actually arrives.
- **README badges** (build status, test coverage, version, license,
  etc.). Locked rejection #2. Depend on CI being live (SPEC-023);
  SPEC-023's doc-sweep activates them or post-MVP.
- **`docs/` subdirectory restructuring.** Locked rejection #1.
  `docs/development.md` is one new flat file; not
  `docs/dev/{architecture,process}.md`.
- **Move of `GETTING_STARTED.md` / `FIRST_SESSION_PROMPTS.md` to
  `docs/`.** Locked rejection #3. They stay at repo root.
- **README content for features that ship later in STAGE-005**
  (Claude hook artifact location, JSON schema cross-link, shell
  completions, CHANGELOG content). Locked rejection #4.
  Forward-references for `brew install jysf/bragfile/bragfile` and
  `CHANGELOG.md` are structurally complete and functionally
  pending — those are the only two forward-references.
- **README-level table of contents.** Locked rejection #5. GitHub
  auto-TOC handles it.
- **Stale STAGE-NNN references in `docs/architecture.md`,
  `docs/api-contract.md`, `docs/data-model.md`.** Deferred to
  SPEC-023's doc-sweep. Punch list below in Premise audit.
- **AGENTS.md updates.** None. Nothing migrating *into* AGENTS.md;
  AGENTS.md is already the source of truth for everything migrating
  out of README. The new `docs/development.md` cross-links into
  AGENTS.md, not the other way around.
- **Wiring `just test-docs` into `just test`.** Q4(a) lock — defer
  the "should it gate CI" question to SPEC-023.
- **Wiring `scripts/test-docs.sh` into a CI workflow.** Q4(a) lock
  — CI doesn't exist yet (SPEC-023 brings it).
- **Marketing copy / SEO / multi-language / branding / logo.** Out
  of scope guardrails per task brief.
- **LLM piping anywhere in the binary.** PROJ-002.
- **Tags-normalization migration / soft-delete / edit-history** —
  STAGE-005 brief out-of-scope.

### Premise audit

The load-bearing section. SPEC-021 is the only STAGE-005 spec that's a
status-change at scale; the audit-family rules (inversion-removal,
count-bump, status-change) all apply. The audit-grep cross-check
(§9 SPEC-018 addendum) was run at design time; results below
reconciled against the enumerated `## Outputs`.

#### Audit-grep enumeration (run at design time, 2026-04-25)

**Grep 1 — `spec-driven` token across user-facing surface:**
```
grep -i 'spec-driven' README.md AGENTS.md BRAG.md GETTING_STARTED.md \
    FIRST_SESSION_PROMPTS.md
grep -ri 'spec-driven' docs/
```
Hits at design time:
- `README.md:10` — "This repo uses a spec-driven workflow…"
  → **MIGRATES** to `docs/development.md`.
- (No hits in `AGENTS.md`, `BRAG.md`, `GETTING_STARTED.md`,
  `FIRST_SESSION_PROMPTS.md`, or `docs/`.)

**Grep 2 — Repo→Project→Stage→Spec hierarchy phrase:**
```
grep -n 'Hierarchy\|hierarchy' README.md AGENTS.md
grep -n 'Project.*Stage.*Spec' README.md AGENTS.md
```
Hits at design time:
- `README.md:12` — "## Hierarchy"
  → **MIGRATES** (the diagram on lines 12–20).
- `AGENTS.md:22` — "## 2. Work Hierarchy"
  → **STAYS** (source of truth).

**Grep 3 — Cycle phrase across the load-bearing surface:**
```
grep -ri 'frame.*design.*build.*verify.*ship\|frame → design\|frame->design' \
    README.md AGENTS.md docs/ BRAG.md GETTING_STARTED.md
```
Hits at design time:
- `README.md:19` — "Cycle (Frame → Design → Build → Verify → Ship)"
  → **MIGRATES** to `docs/development.md`.
- `AGENTS.md:166` — "frame → design → build → verify → ship"
  → **STAYS** (source of truth, AGENTS.md §6).
- `docs/CONTEXTCORE_ALIGNMENT.md:31` — "through Frame → Design →
  Build → Verify → Ship. ContextCore's…"
  → **STAYS** (framework-alignment doc; that file's home is the
  cycle).

**Grep 4 — Contributor-shaped tokens that need to leave the README:**
```
grep -n 'context contamination\|four habits\|new session\|architect.*implementer.*reviewer' README.md
```
Hits at design time:
- `README.md:10` — "Claude plays every role (architect, implementer,
  reviewer)" → **MIGRATES**.
- `README.md:39` — "Because Claude plays every role, context
  contamination is the biggest risk. Four habits keep it at bay:"
  → **MIGRATES**.
- `README.md:41` — "**New Claude session per cycle**…"
  → **MIGRATES** (part of the four-habits block).

**Grep 5 — STAGE-NNN tense references in the user-facing surface:**
```
grep -rn 'STAGE-00[0-9]' README.md BRAG.md docs/tutorial.md \
    docs/api-contract.md docs/architecture.md docs/data-model.md
```
Hits at design time, by file:

*`README.md`* (load-bearing for THIS spec):
- `:31` — `just new-spec "title" STAGE-001` (example arg)
  → **MIGRATES** out (contributor-shaped just-recipe).
- `:63` — "are shipped (STAGE-004; markdown default…)"
  → **REMOVED** in user-facing rewrite (digest commands are
  shipped reality; user-facing prose names them by behaviour, not
  stage tense).
- `:79` — "Homebrew install (`brew install bragfile`) arrives in
  STAGE-005."
  → **REPLACED** with dual-path forward-reference (`brew install
  jysf/bragfile/bragfile` written now; structurally complete,
  functionally pending until SPEC-023's first tag).
- `:111` — "| `cmd/brag/` | CLI entrypoint (added during STAGE-001) |"
  → **REMOVED** (the user-facing where-things-live shape no longer
  carries stage attribution; the contributor table that does is
  migrated to `docs/development.md`).

*`BRAG.md`*: zero STAGE-NNN hits. ✓ No action.

*`docs/tutorial.md`* (5 hits at lines 32, 34, 157, 200, 493):
- Lines 32/34/157/200 are example brag titles ("shipped STAGE-001
  end-to-end in a day") — these are *example data*, not status
  claims. ✓ No action.
- Line 493: "`brew install bragfile` | STAGE-005 |" in §9 "What's
  NOT there yet" table. → **DEFERRED** to SPEC-023's doc-sweep.
  When SPEC-023 ships brew install, it strikes this row.
  ENUMERATED in SPEC-023 punch list below.

*`docs/api-contract.md`* (13 hits at lines 33, 50, 80, 134–135, 137,
146, 173, 186, 215, 251, 287, 334):
- All are "(STAGE-NNN)" tense markers describing when each command
  shipped. → **DEFERRED** to SPEC-023's doc-sweep — these are
  code-path lineage references that read fine while the project
  is in motion; SPEC-023's tense sweep can rephrase to past
  tense or remove. ENUMERATED in SPEC-023 punch list below.

*`docs/architecture.md`* (4 hits at lines 24, 44, 45, 103):
- Same shape — code-path lineage. → **DEFERRED** to SPEC-023.

*`docs/data-model.md`* (5 hits at lines 6, 89, 94, 96, 111):
- Same shape — code-path lineage. → **DEFERRED** to SPEC-023.

**Grep 6 — Inbound references TO `README.md`:**
```
grep -rn -i 'README' AGENTS.md BRAG.md GETTING_STARTED.md \
    FIRST_SESSION_PROMPTS.md docs/
```
Hits at design time:
- `AGENTS.md:116` — directory tree path entry ("`├── README.md`").
  → **STAYS** (path reference, no content claim).
- `AGENTS.md:224, :225` — narrative mentions inside the
  premise-audit addenda (talk about how READMEs are typical
  cross-reference targets). → **STAYS** (commentary, no content
  claim about THIS README).
- `BRAG.md` — zero hits. ✓
- `GETTING_STARTED.md:40, :177` — references the README in the
  context of template scaffolding for a *new project's frame*.
  → **STAYS** (about creating a README for a new project, not
  about this README's content).
- `FIRST_SESSION_PROMPTS.md:63, :227` — `/README.md` in a list of
  files to read at session start. → **STAYS** (existence reference).
- `docs/`: zero hits.

**No inbound references to README depend on its current dev-process
content.** ✓ The rewrite is content-only; no inbound-link rewriting
is needed.

**Grep 7 — Inbound references TO `CONTRIBUTING.md` (must be empty
pre-spec):**
```
grep -rn -F 'CONTRIBUTING.md' . --include='*.md' --exclude-dir=projects \
    --exclude-dir=node_modules --exclude-dir=.git
```
Hits at design time: zero. ✓ No prior reference to a non-existent
file. After this spec, hits exist in `README.md` and `CONTRIBUTING.md`
itself.

**Grep 8 — Inbound references TO `docs/development.md` (must be
empty pre-spec):**
```
grep -rn -F 'docs/development.md' . --include='*.md' --exclude-dir=projects \
    --exclude-dir=node_modules --exclude-dir=.git
```
Hits at design time: zero. ✓ After this spec, hits exist in
`README.md`, `CONTRIBUTING.md`, and `docs/development.md` (self).
Excluding `projects/` keeps planning docs (which DO mention
`docs/development.md` as a STAGE-005 outcome) out of the integrity
check.

#### Premise-audit family applicability

- **Inversion/removal** (planned test deletion under
  `## Outputs`): N/A. No existing test's premise is invalidated by
  this spec; the only test artifact is the new
  `scripts/test-docs.sh`. No test deletions planned.
- **Addition** (count-bump on tracked collections): N/A. No
  migrations, DECs, constraints, or fixed-shape collections are
  modified. (The `decisions/` directory count is unchanged at 14;
  `internal/storage/migrations/` count is unchanged; `constraints.yaml`
  is untouched.)
- **Status change** (planned doc references update): **APPLIES at
  scale.** This is the load-bearing case for SPEC-021. The "feature
  status" being changed is "the README's content category" — from
  dev-process to user-facing. Every doc that referenced the README
  was checked above (Grep 6); no inbound link to README depended on
  its old dev-process content. Every doc that mirrored the migrated
  content was checked above (Greps 1–4); each migrated phrase is
  catalogued in `## Outputs`.
- **Audit-grep cross-check (both sides)**: Design-side run above;
  enumeration matches `## Outputs`. Build-side: re-run all 8 greps
  before locking the rewrite; treat any new delta as a question for
  the spec author (raise in Build Completion reflection), not a
  unilateral expansion of scope.

#### Known stale stage refs deferred to later STAGE-005 specs

Closes the SPEC-020 ship-reflection carry-forward ("routine STAGE-00N
grep at stage start should be routine"). These hits are deliberately
NOT touched by SPEC-021 because they describe code-path lineage
(historical, accurate regardless of current state) rather than
user-facing status claims invalidated by the README rewrite. Punch
list inherited by **SPEC-023**'s doc-sweep:

**Deferred to SPEC-023 doc-sweep:**

| File | Line | Current text | Proposed action at SPEC-023 |
|---|---|---|---|
| `docs/tutorial.md` | 493 | `\| \`brew install bragfile\` \| STAGE-005 \|` | Strike row entirely (feature shipped). |
| `docs/api-contract.md` | 33 | `**STAGE-001 (flags-only form):**` | Past tense or strip stage. |
| `docs/api-contract.md` | 50 | `**STAGE-002 (editor-launch form):**` | Past tense or strip stage. |
| `docs/api-contract.md` | 80 | `**STAGE-003 (JSON stdin form):**` | Past tense or strip stage. |
| `docs/api-contract.md` | 134–135 | "STAGE-001 ships without filter flags … added in STAGE-002." | Strip / past tense. |
| `docs/api-contract.md` | 137 | `### \`brag show <id>\` — show a single entry (STAGE-002)` | Strip stage suffix. |
| `docs/api-contract.md` | 146 | `### \`brag edit <id>\` — edit via $EDITOR (STAGE-002)` | Strip stage suffix. |
| `docs/api-contract.md` | 173 | `### \`brag delete <id>\` — delete an entry (STAGE-002)` | Strip stage suffix. |
| `docs/api-contract.md` | 186 | `### \`brag search "query"\` — full-text search (STAGE-002)` | Strip stage suffix. |
| `docs/api-contract.md` | 215 | `### \`brag export\` — export entries (STAGE-003)` | Strip stage suffix. |
| `docs/api-contract.md` | 251 | `### \`brag summary --range week\|month\` (STAGE-004)` | Strip stage suffix. |
| `docs/api-contract.md` | 287 | `### \`brag review --week \| --month\` (STAGE-004)` | Strip stage suffix. |
| `docs/api-contract.md` | 334 | `### \`brag stats\` (STAGE-004)` | Strip stage suffix. |
| `docs/architecture.md` | 24 | Mermaid label "shipped in STAGE-002 / STAGE-003 / summary STAGE-004" | Strip stage tense from Mermaid label; tool surface is shipped. |
| `docs/architecture.md` | 44 | `\| \`internal/editor\` \| (STAGE-002) Launches \`$EDITOR\`…` | Strip leading `(STAGE-NNN)`. |
| `docs/architecture.md` | 45 | `\| \`internal/export\` \| (STAGE-003) Markdown-report and sqlite-file-copy exporters.` | Strip leading `(STAGE-NNN)` AND fix the now-stale "sqlite-file-copy" claim — `--format sqlite` was scoped out to backlog. |
| `docs/architecture.md` | 103 | `Distribution (STAGE-004) uses goreleaser…` | Update to STAGE-005 actual; or rephrase past-tense. |
| `docs/data-model.md` | 6 | `STAGE-002 adds a third virtual table…` | Past tense. |
| `docs/data-model.md` | 89 | `Planned for STAGE-001 (ship with the initial migration):` | Past tense / strip. |
| `docs/data-model.md` | 94 | `(flags land in STAGE-002).` | Past tense / strip. |
| `docs/data-model.md` | 96 | `Shipped in STAGE-002 (SPEC-011):` | Already past-tense — leave or strip stage. |
| `docs/data-model.md` | 111 | `\`brag edit\` (STAGE-002) opens \`$EDITOR\`…` | Strip `(STAGE-NNN)`. |

**Deferred to SPEC-023 (release-tense activation):**
- `docs/tutorial.md:493` — strike the row from the "What's NOT
  there yet" table when brew install ships.
- `README.md` — switch the brew-install forward-reference from
  "structurally-pending" framing to "you can install via brew" once
  SPEC-023's first tag fires.

This punch list closes the SPEC-020 carry-forward — SPEC-023 inherits
a mechanical doc-sweep instead of re-grepping at build time.

#### NOT-contains self-audit (SPEC-019/SPEC-020 pattern)

Group B contains 7 NOT-contains assertions. Per the SPEC-020 codified
discipline, **grep the spec's own load-bearing prose** (the README
sketch in Notes for the Implementer below) for each forbidden token.
The load-bearing prose is what reaches the binary — for THIS spec,
that's the literal markdown block under "README sketch" in Notes for
the Implementer. The rest of this spec (Acceptance Criteria, Failing
Tests, this audit section itself) is *commentary* and may freely
contain forbidden tokens — those don't render anywhere.

**Forbidden tokens (from group B):**
1. `spec-driven` (B1)
2. `Frame → Design → Build → Verify → Ship` and ASCII variants (B2)
3. `four habits` (B3)
4. `context contamination` (B4)
5. `just new-spec`, `just advance-cycle`, `just archive-spec`,
   `just weekly-review`, `just new-stage` (B5)
6. `Claude plays every role` (B6)
7. Top-level TOC / TOC heading (B7)

**Self-audit at design time:** the README sketch under Notes for the
Implementer was reviewed for each of the 7 tokens. Results:
- Token 1 (`spec-driven`): zero hits in sketch. ✓
- Token 2 (cycle phrase, any form): zero hits. ✓
- Token 3 (`four habits`): zero hits. ✓
- Token 4 (`context contamination`): zero hits. ✓
- Token 5 (just-recipe contributor refs): zero hits. ✓
- Token 6 (`Claude plays every role`): zero hits. ✓
- Token 7 (TOC): the sketch has no `## Table of Contents` heading
  and no contiguous `- [...]` block in the first 50 lines. ✓

**Build-time re-audit:** before locking the rewritten `README.md`,
the build session **must** re-run these greps against the actual
file (not the sketch — the sketch is illustrative; the build session
may rephrase). The seven NOT-contains assertions in
`scripts/test-docs.sh` are the mechanized form of this audit.

## Notes for the Implementer

Gotchas, style preferences, reuse opportunities, and literal sketches.
Default to fuller skeleton (first STAGE-005 spec; trim heuristic does
NOT apply — see § stage Design Notes).

### Build-cycle order

1. **Start a fresh Claude session.** Do not continue from the design
   session. Read the spec's Implementation Context, AGENTS.md
   (especially §6, §9, §10, §12), and the parent stage file.
2. **Branch.** `git checkout -b chore/spec-021-readme-rewrite`. The
   `chore/` prefix is appropriate per AGENTS.md §10 because this
   spec's primary delivery is doc restructure — there's no
   feature/bug shape to it. (If you prefer `feat/spec-021-…` for
   consistency with stage convention, that's also acceptable; both
   pass `one-spec-per-pr`.)
3. **Re-run the design-time greps.** All 8 in § Premise audit
   above. Reconcile any delta against `## Outputs`. Raise
   discrepancies via Build Completion reflection rather than
   silently expanding scope.
4. **Write `scripts/test-docs.sh` first** (TDD-in-spirit). All 40
   assertions across groups A–G. Make the script executable
   (`chmod +x`). Do NOT write the new markdown yet.
5. **Run `just test-docs` against the current tree.** Expect the
   majority of group A to FAIL (current README doesn't have the
   prescribed install heading, lacks the workflow demo for all
   seven commands in the prescribed shape, etc.) and groups B/C/D
   to FAIL (current README contains migrated tokens; CONTRIBUTING.md
   and docs/development.md don't exist yet). **Confirm the FAILs
   are for the *expected* reason** (the assertion you wrote, not a
   stray shell error or `set -u` undefined-variable). Per AGENTS.md
   §9 fail-first discipline.
6. **Write `docs/development.md`.** Use the literal sketch below.
   Cross-link AGENTS.md §11 glossary (chain: README →
   CONTRIBUTING → development.md → AGENTS.md §11). DO NOT
   duplicate glossary content here.
7. **Write `CONTRIBUTING.md`.** Thin file (~50–70 lines). Use the
   literal sketch below. Point at `docs/development.md` and
   `AGENTS.md`.
8. **Rewrite `README.md`.** Use the literal sketch below. Aim for
   150–200 lines. Each command gets one 1–3-line example, then
   point to `docs/tutorial.md` for depth.
9. **Add `test-docs` recipe to `justfile`.** Single recipe under
   the daily-commands section. Do NOT modify the existing `test`
   recipe.
10. **Run `just test-docs` again.** Expect all 40 assertions to
    OK. Run `just test` separately to confirm `go test ./...`
    still passes (group H sanity).
11. **Run the manual group H verify-cycle checks** (eight
    byte-identity diffs). Confirm `BRAG.md`, `AGENTS.md`,
    `GETTING_STARTED.md`, `FIRST_SESSION_PROMPTS.md`,
    `docs/tutorial.md`, `docs/api-contract.md`,
    `docs/architecture.md`, `docs/data-model.md` are
    byte-identical to pre-spec via `git diff`.
12. **Fill in `## Build Completion`** including all three
    reflection answers (real, not placeholder — `archive-spec.sh`
    rejects empty `<answer>` placeholders post-`bfa1474`, and the
    same discipline applies to the build reflection).
13. **`just advance-cycle SPEC-021 verify`.**
14. **Open PR.** PR description includes `Project: PROJ-001`,
    `Stage: STAGE-005`, `Spec: SPEC-021`, "no DECs referenced or
    emitted", "constraints checked: N/A (no Go code)".

### Push-discipline reminder (§10 newly-codified rule)

**Critical at the merge step.** The push-discipline rule was just
codified at STAGE-005 framing (commit `76630fc`, 2026-04-25):
> "Any commit added to a feat branch just before
> `gh pr merge --squash --delete-branch` MUST be pushed to
> `origin/<feat-branch>` before the merge."

Three confirming cases led to the rule (SPEC-013, SPEC-018, SPEC-019).
SPEC-021 is the first spec post-codification; it's the test of whether
the rule has actually internalised. Concrete heuristic: after any
`git commit` on the feat branch in the same shell session as the PR
merge, run `git push origin HEAD` BEFORE `gh pr merge`. Especially
relevant for this spec because doc-tweaks-after-review-feedback are
likely (verify cycle may surface tone fixes that get committed
just before merge).

### README sketch

Literal markdown for the rewritten README. ~170 lines targeted. The
build session may rephrase tone but must preserve the structural
beats so that all of group A's assertions pass. **This sketch is
load-bearing prose** — the seven NOT-contains tokens MUST NOT appear
inside the triple-backticked block below. (Self-audited at design
time; zero hits.)

```markdown
# Bragfile

`brag` is a local-first command-line tool that captures your
brag-worthy work moments — shipped features, fixed bugs, things you
learned, mentoring you delivered — and lets you retrieve them later
for retros, reviews, and resumes. Entries live in an embedded SQLite
database at `~/.bragfile/db.sqlite` on your machine. No cloud, no
sync, no account.

> **Status:** in active development. Capture, retrieve, search,
> export, and weekly/monthly digests are shipped. Distribution via
> Homebrew is in progress.

## Install

Homebrew (recommended once available):

```bash
brew install jysf/bragfile/bragfile
brag --version
```

From source (works today):

```bash
git clone https://github.com/jysf/bragfile000.git
cd bragfile000
just install                 # or: go install ./cmd/brag
brag --version               # confirm ~/go/bin is on $PATH
```

Requires Go 1.26+ from source. The Homebrew install pulls a
prebuilt binary — no Go required.

## Capture an entry

The fastest path — one flag:

```bash
brag add --title "shipped FTS5 search end-to-end"
# prints the new entry's ID on stdout, e.g. "12"
```

With full metadata:

```bash
brag add \
  --title "cut p99 login latency from 600ms to 120ms" \
  --project platform \
  --type shipped \
  --tags auth,perf,backend \
  --impact "unblocked mobile v3 release"
```

For longer narrative entries, `brag add` with no flags opens
`$EDITOR` against a templated buffer:

```bash
brag add        # → editor opens; fill in the fields, save, quit
```

For programmatic capture from a script or AI agent, pipe JSON to
`brag add --json` (see [`BRAG.md`](BRAG.md)):

```bash
echo '{"title":"…","project":"…"}' | brag add --json
```

## Read entries back

List them, newest first:

```bash
brag list                                  # all entries
brag list --project platform --since 30d   # filter by project + window
brag list -P                               # add a project column
brag list --format json                    # machine-readable
```

Search across every field via SQLite FTS5:

```bash
brag search "latency"
brag search "auth-refactor"     # hyphens are literal, not operators
```

Show the full record for a single entry, edit it, or delete it:

```bash
brag show 12
brag edit 12
brag delete 12
```

## Export for reviews

Markdown report grouped by project (paste into a quarterly review
or promo packet):

```bash
brag export --format markdown --since 90d > q-review.md
```

JSON dump (for AI piping or backup):

```bash
brag export --format json --since 90d > q-review.json
```

## Weekly and monthly digests

Rule-based aggregations of recent entries — no LLM, no network.
Pipe the JSON into your favourite AI session for guided
reflection.

```bash
brag summary --range week               # 7-day digest, grouped
brag summary --range month --format json
brag review --week                      # entries + reflection prompts
brag stats                              # lifetime metrics
```

## Where the data lives

```
~/.bragfile/db.sqlite
```

Back up by copying the file. Move to a new machine by copying the
file. Override the path with the `--db` flag or the `BRAGFILE_DB`
environment variable.

## Where to go next

- [`docs/tutorial.md`](docs/tutorial.md) — the deep-dive
  walkthrough: every command, every flag, every gotcha.
- [`BRAG.md`](BRAG.md) — guide for AI coding agents that want to
  propose brag entries from work sessions.
- [`CONTRIBUTING.md`](CONTRIBUTING.md) — how this repo is built
  and how to contribute.
- [`docs/api-contract.md`](docs/api-contract.md) — full CLI
  reference.

## License

MIT. See [`LICENSE`](LICENSE).
```

(End of README sketch.)

**Notes on the sketch:**
- `# Bragfile` H1 satisfies A2.
- The first paragraph mentions `brag`, `capture`, `retrieve`,
  `accomplishment` (covering retros/reviews/resumes), satisfying
  A3's positive shape. It contains no spec-driven / architect /
  implementer / reviewer / cycle / hierarchy tokens, satisfying A3's
  negative shape.
- `## Install` heading + both `brew install jysf/bragfile/bragfile`
  AND `go install ./cmd/brag` lines satisfy A4.
- Each of the seven workflow commands (`brag add`, `brag list`,
  `brag search`, `brag export`, `brag summary`, `brag review`,
  `brag stats`) appears in fenced shell blocks, satisfying A5.
- `~/.bragfile/db.sqlite` appears literally, satisfying A6.
- `docs/tutorial.md`, `BRAG.md`, `CONTRIBUTING.md` all link literals,
  satisfying A7/A8/A9.
- `## License` heading + `MIT`, satisfying A10.

### CONTRIBUTING.md sketch

Literal markdown. ~55 lines targeted.

```markdown
# Contributing to Bragfile

Bragfile is a personal-tool project — built primarily for the
author's own daily use. PRs are welcome but not actively recruited.
If you've found a bug, opened an issue, or want to suggest a small
improvement, you're in the right place.

## Development setup

Requires Go 1.26+ and `just` (optional but recommended).

```bash
git clone https://github.com/jysf/bragfile000.git
cd bragfile000
just install              # or: go install ./cmd/brag
just test                 # run the Go test suite
brag --version            # confirm install
```

## How this repo is built

This project uses a structured workflow where Claude (the AI assistant)
plays each role across separate sessions: writing specifications,
implementing them, and reviewing the result. The development process
is documented in [`docs/development.md`](docs/development.md);
[`AGENTS.md`](AGENTS.md) is the full conventions document.

If you're proposing a change, the simplest path is to:

1. Open an issue describing what you'd like to change and why.
2. Wait for confirmation that the direction makes sense.
3. Open a PR against `main`.

## Pull request conventions

- One change per PR.
- Branch naming: `feat/<slug>` for features, `fix/<slug>` for
  fixes, `chore/<slug>` for tooling/docs.
- Commit messages: short conventional-style subject
  (e.g. `feat(storage): add Entry type`); body optional.
- See [`AGENTS.md` §10](AGENTS.md) for full git/PR conventions.

## Tests

```bash
just test                 # Go test suite
just test-docs            # documentation-content assertions
gofmt -l .                # formatting check (must be empty)
go vet ./...              # static checks
```

## License

By contributing, you agree that your contributions will be
licensed under the project's [MIT License](LICENSE).
```

(End of CONTRIBUTING.md sketch.)

### docs/development.md sketch

Literal markdown. ~95 lines targeted.

```markdown
# Development — the spec-driven workflow

How this project is built. Read this if you're contributing,
modifying the workflow, or curious why the repo has the structure
it does.

## TL;DR

Each piece of work is a **spec** — a single implementable task —
that moves through five **cycles**:

```
Frame → Design → Build → Verify → Ship
```

A fresh Claude session is started for each cycle. The spec file
itself is the handoff between sessions, so any session can pick
up where the previous left off without "remembering" prior
context.

## Hierarchy

Work organises into four levels:

```
Repo (this app)
 └─ Project (a wave of work — e.g. "MVP", "v2 redesign")
     └─ Stage (a coherent chunk within a project — 2–5 per project)
         └─ Spec (an individual implementable task)
              └─ Cycle (Frame → Design → Build → Verify → Ship)
```

The repo persists across all projects. A project is a bounded wave.
A stage is an epic-sized chunk within a project. A spec is one task.

## Session hygiene

Because one Claude assistant plays multiple roles, four habits keep
work coherent across sessions:

1. **Start a fresh session per cycle.** Especially design → build
   and build → verify.
2. **The spec file is the source of truth between sessions.** Don't
   rely on "as I said earlier" — the next session won't remember.
3. **Run a weekly review.** Without a second agent pushing back,
   drift compounds silently.
4. **Honest confidence values on decisions** — see
   [`AGENTS.md` §14](../AGENTS.md).

## Daily commands

```bash
just status                          # active project, stage, specs by cycle
just new-spec "title" STAGE-001      # scaffold a new spec
just advance-cycle SPEC-001 verify   # advance a spec's cycle
just archive-spec SPEC-001           # move a shipped spec to done/
just weekly-review                   # print the weekly review prompt
just specs-by-stage                  # group all specs by their stage
```

`just --list` shows every recipe.

## Where the conventions live

- [`AGENTS.md`](../AGENTS.md) — full conventions for working in
  this repo. The source of truth. See particularly:
  - **§6** — Cycle Model.
  - **§8/§9** — Coding and Testing Conventions.
  - **§10** — Git/PR Conventions.
  - **§11** — Domain Glossary (what we mean by "aggregate",
    "Store", "tap", and so on).
  - **§13** — Session Hygiene (the four habits above, expanded).
- [`GETTING_STARTED.md`](../GETTING_STARTED.md) — first-project
  walkthrough; if you're forking the framework into a new repo,
  start there.
- [`FIRST_SESSION_PROMPTS.md`](../FIRST_SESSION_PROMPTS.md) — the
  copy-paste prompts that drive each cycle.
- [`docs/CONTEXTCORE_ALIGNMENT.md`](./CONTEXTCORE_ALIGNMENT.md) —
  how this workflow maps to ContextCore's task taxonomy.

## Where to find what

| Looking for… | Look in |
|---|---|
| Architecture overview | [`docs/architecture.md`](./architecture.md) |
| Data model and schema | [`docs/data-model.md`](./data-model.md) |
| Full CLI reference | [`docs/api-contract.md`](./api-contract.md) |
| User tutorial | [`docs/tutorial.md`](./tutorial.md) |
| Decision log | [`decisions/`](../decisions/) |
| Repo-level rules | [`guidance/constraints.yaml`](../guidance/constraints.yaml) |
| Active project brief | [`projects/PROJ-001-mvp/brief.md`](../projects/PROJ-001-mvp/brief.md) |
```

(End of docs/development.md sketch.)

### test-docs.sh sketch

Literal scaffold for `scripts/test-docs.sh`. The implementer fills
in the per-assertion bodies; the structure shown here gives the
exit-code contract, the SCRIPT_DIR pattern, and the OK/FAIL output
shape.

```sh
#!/usr/bin/env sh
# scripts/test-docs.sh — documentation-content assertions for
# the bragfile repo. Exits 0 iff all assertions pass.
#
# Run via `just test-docs`. Not wired into `just test` (Go-only).
set -eu

SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
REPO_ROOT=$(cd "$SCRIPT_DIR/.." && pwd)
cd "$REPO_ROOT"

FAIL_COUNT=0

ok()   { printf 'OK:   %s\n' "$1"; }
fail() { printf 'FAIL: %s: %s\n' "$1" "$2"; FAIL_COUNT=$((FAIL_COUNT + 1)); }

# ----- helpers -----
# assert_contains <name> <file> <pattern>      (extended grep, fixed-string-safe)
# assert_not_contains <name> <file> <pattern>
# assert_line_count_band <name> <file> <min> <max>
# assert_file_exists <name> <path>
# (definitions go here)

# ===== Group A — README shape (positive) =====
# A1 — README line count band
# A2 — README opens with H1
# A3 — Above-the-fold is user-facing
# A4 — Install section with both paths
# A5 — Workflow-demo command coverage (all 7 brag verbs in fenced blocks)
# A6 — ~/.bragfile/db.sqlite reference
# A7–A9 — Tutorial / BRAG / CONTRIBUTING pointers
# A10 — License section

# ===== Group B — README shape (negative) =====
# B1–B6 — Forbidden tokens absent
# B7 — No top-level TOC

# ===== Group C — CONTRIBUTING.md =====
# C1–C5

# ===== Group D — docs/development.md =====
# D1–D8

# ===== Group E — Link integrity =====
# E1 — Internal links resolve in README, CONTRIBUTING, development.md
# E2 — docs/development.md not referenced before this spec (modulo this spec's own outputs)
# E3 — CONTRIBUTING.md is brand-new (no prior deletion)

# ===== Group F — Just-recipe wiring =====
# F1 — test-docs recipe defined
# F2 — `test:` recipe unchanged (recipe-block diff against pre-spec)
# F3 — scripts/test-docs.sh executable + POSIX-headed
# F4 — harness self-pass meta

# ===== Group G — Harness ergonomics =====
# G1 — per-check OK/FAIL output (review-only)
# G2 — exit-code contract (built-in, see below)
# G3 — works from any cwd (covered by SCRIPT_DIR pattern)

# ----- finalise -----
if [ "$FAIL_COUNT" -gt 0 ]; then
    printf '\nFAILED: %d assertion(s) failed.\n' "$FAIL_COUNT" >&2
    exit 1
fi
printf '\nALL OK: documentation-content assertions passed.\n'
exit 0
```

(End of test-docs.sh sketch.)

### Style preferences and gotchas

- **Match existing markdown tone.** Look at `BRAG.md` and
  `docs/tutorial.md` — both are in the same voice (technical, terse,
  example-driven, no marketing). Match that voice in the new files.
- **Do NOT introduce emojis** in the README/CONTRIBUTING/development.md
  files (the user's global preferences disallow gratuitous emojis,
  and the existing repo style is emoji-free in markdown apart from
  the stderr feedback codified for `brag edit` / `brag delete`).
- **Internal links use relative paths**, not absolute URLs.
  `docs/tutorial.md` from README is `docs/tutorial.md`, not
  `https://github.com/jysf/bragfile000/blob/main/docs/tutorial.md`.
- **Preserve `Frame → Design → Build → Verify → Ship` as the exact
  Unicode-arrow form** in `docs/development.md`. Group D4 asserts
  on this literal substring; matching keeps grep-discoverability with
  `AGENTS.md:166` and `docs/CONTEXTCORE_ALIGNMENT.md:31`.
- **Don't duplicate the AGENTS.md §11 glossary into
  `docs/development.md`.** Cross-link only. Q3 lock.
- **The `brag review --week` example using `claude` as a CLI
  pipe target** in `docs/tutorial.md:373` is illustrative, not a
  hard requirement; the README sketch above does NOT use that
  pattern — it just says "pipe the JSON into your favourite AI
  session." Keep README's examples generic.
- **`go install ./cmd/brag` vs `just install`.** Either works; the
  README sketch shows `just install` as the recommended path with
  `go install` as the alternative. A4 asserts on either form.
- **The brew-install line in the README is a forward-reference.**
  After SPEC-023 ships, the line becomes a working instruction
  rather than aspirational. The README's prose around it
  ("Homebrew (recommended once available)") is what marks it as
  pending; that prose updates in SPEC-023's doc-sweep.
- **Don't mention the homebrew tap repo URL
  (`github.com/jysf/homebrew-bragfile`) in README** — that's
  internal mechanics for SPEC-023; users only need `brew install
  jysf/bragfile/bragfile`.
- **`set -eu` in `scripts/test-docs.sh`.** Mirrors the established
  pattern in `scripts/_lib.sh` and friends. Without `set -u`, an
  uninitialised variable silently passes assertions; without
  `set -e`, a failing pipeline early in the script can silently
  skew counts. (`set -eu` is the right shape; `set -euo pipefail`
  is also fine if you want pipefail safety, but POSIX `sh` doesn't
  guarantee `pipefail` — use `bash` shebang if you want it.)
- **Counting `wc -l`.** `wc -l < README.md` returns just the
  number; `wc -l README.md` includes the filename. Use the input
  redirection form to avoid having to strip the filename.
- **Group F2 (recipe-block diff).** The simplest implementation:
  capture the pre-spec `test:` recipe block (lines from `^test:` up
  to the next blank line) at design time and store it in the spec
  for build to compare against. Pre-spec block is reproduced below
  for explicit reference:
  ```
  test:
      @go test ./...
  ```
  (Two lines: header + body. The body is `    @go test ./...` — four
  spaces of indent, the `@` quiet prefix, then the command.)
  Build compares against this pre-spec capture; F2 passes if the
  post-spec block matches verbatim.

### Verify-cycle checklist for group H

Group H lives outside the script as `git diff` checks. Verify-cycle
review runs:

```bash
git diff --stat HEAD~ -- BRAG.md AGENTS.md GETTING_STARTED.md \
    FIRST_SESSION_PROMPTS.md docs/tutorial.md docs/api-contract.md \
    docs/architecture.md docs/data-model.md
```

Expected: empty output (no changes to those eight files). Any line
returned indicates a violation of group H — investigate before
approving.

### Reuse opportunities

- **`scripts/_lib.sh`** — existing shared shell helpers (used by
  `status.sh`, `new-spec.sh`, etc.). Read it before writing
  `test-docs.sh`; if it has a `print_section` or color-print
  helper that fits the OK/FAIL output, reuse it. If not, don't
  invent one — `printf` is fine.
- **The `SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)` pattern**
  appears in multiple existing scripts. Use the same form.

### What NOT to do

- Don't add a CHANGELOG entry. CHANGELOG.md doesn't exist yet
  (SPEC-023's territory); creating it here violates locked
  rejection #4.
- Don't add CI integration. SPEC-023 brings CI. Q4(a) lock.
- Don't pre-write SPEC-022's BRAG.md schema cross-reference. That's
  SPEC-022's scope.
- Don't add a GIF or image placeholder. Defer to SPEC-022.
- Don't restructure `docs/`. Locked rejection #1.
- Don't move `GETTING_STARTED.md` or `FIRST_SESSION_PROMPTS.md`.
  Locked rejection #3.
- Don't add badges. Locked rejection #2.
- Don't add a TOC to README. Locked rejection #5.
- Don't refactor any of the eight files in group H — those are
  byte-identity invariants for this spec.

### Rejected alternatives (build-time)

Per SPEC-018+ discipline, alternatives explicitly rejected at design
time so they don't slip into Deviations later:

1. **Single new file (CONTRIBUTING.md OR docs/development.md, not
   both).** Rejected: GitHub-conventional CONTRIBUTING.md needs to
   be thin (PR conventions etc.) without bloating into framework
   internals. Two-file split (Q1 lock, option iii) gives the right
   shape.
2. **Move dev-process content entirely into AGENTS.md instead of
   a new docs/development.md.** Rejected: AGENTS.md is the source of
   truth and contains the FULL detail (§6 cycle model + §13 session
   hygiene + §11 glossary). docs/development.md is a digestible
   intermediate-tier doc that points at AGENTS.md. Removing it
   would force CONTRIBUTING.md to either duplicate AGENTS.md content
   (DRY violation) or punt readers straight into a 391-line
   conventions doc (UX cliff). Two-file split bridges this.
3. **`scripts/test/test-docs.sh` (subdirectory).** Rejected: Q4(b)
   lock — flat `scripts/` matches established convention; no
   payoff to a subdirectory for a single new script.
4. **Inlining the test assertions as `## Acceptance Criteria`
   checklist items run by hand.** Rejected: 40 assertions is too
   many for reliable manual eyeballing per verify cycle. The
   shell harness is a one-time write that makes every future
   STAGE-005 doc-tweak verifiable in a single `just test-docs`
   command.
5. **Wiring `just test-docs` into `just test`.** Rejected: Q4(a)
   lock — would change the shape of `just test` (Go-only) without
   payoff before CI lands. SPEC-023 can revisit when CI exists.
6. **Including the brew-install instructions only after SPEC-023
   ships (omit from this rewrite).** Rejected: Q1 option (b) lock —
   structurally-pending forward-reference is preferable to a hole
   in the install section. The README is a snapshot of what works
   today plus structurally-pending sections for what will work
   after STAGE-005's later specs.
7. **Including a CHANGELOG.md cross-reference in the README.**
   Considered then rejected: CHANGELOG.md doesn't exist until
   SPEC-023; a forward-reference to a non-existent file is a
   broken-link risk that group E1 would flag. The README
   intentionally has no CHANGELOG pointer; SPEC-023 adds one as
   part of its doc-sweep when CHANGELOG.md materialises.

---

## Build Completion

*Filled in at the end of the **build** cycle, before advancing to
verify.*

- **Branch:** <fill in>
- **PR (if applicable):** <fill in>
- **All acceptance criteria met?** <yes/no>
- **New decisions emitted:** none expected; confirm none added.
  - <if any: DEC-NNN — title>
- **Deviations from spec:**
  - <list, or "none">
- **Follow-up work identified:**
  - <any new specs for the stage's backlog, or "none">

### Build-phase reflection (3 questions, short answers)

Process-focused: how did the build go? What friction did the spec
create?

1. **What was unclear in the spec that slowed you down?**
   — <answer>

2. **Was there a constraint or decision that should have been listed
   but wasn't?**
   — <answer>

3. **If you did this task again, what would you do differently?**
   — <answer>

---

## Reflection (Ship)

*Appended during the **ship** cycle. Outcome-focused reflection,
distinct from the process-focused build reflection above. NOTE:
`scripts/archive-spec.sh` rejects empty `<answer>` placeholders
(commit `bfa1474`, 2026-04-25) — these MUST be filled with real
answers at ship time or `just archive-spec SPEC-021` will fail.*

1. **What would I do differently next time?**
   — <answer>

2. **Does any template, constraint, or decision need updating?**
   — <answer>

3. **Is there a follow-up spec I should write now before I forget?**
   — <answer>
