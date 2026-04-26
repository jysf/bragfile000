---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-023
  type: chore                      # epic | story | task | bug | chore
                                   # chore (not story): no new `brag` CLI
                                   # surface ships in this spec; it's
                                   # build-and-release plumbing plus a
                                   # version-injection wiring change in
                                   # cmd/brag/main.go and a doc sweep
                                   # over previously-deferred references.
  cycle: verify
  blocked: false
  priority: medium
  complexity: M                    # S | M | L  (L means split it)
                                   # M (not L): four fixed-shape literal
                                   # artifacts (.goreleaser.yaml + two
                                   # GitHub Actions workflows + CHANGELOG.md
                                   # populated retroactively for v0.1.0)
                                   # plus a six-line const→var change in
                                   # cmd/brag/main.go for ldflags injection
                                   # plus a doc-sweep across six files
                                   # plus ~33 new shell asserts in
                                   # scripts/test-docs.sh. Stage flagged
                                   # "watch-for-split"; design rejected
                                   # the split (see § Bundle vs split
                                   # decision).

project:
  id: PROJ-001
  stage: STAGE-005
repo:
  id: bragfile

agents:
  architect: claude-opus-4-7
  implementer: claude-opus-4-7     # usually same Claude, different session
  created_at: 2026-04-26

references:
  decisions:
    - DEC-001                      # pure-Go SQLite driver — load-bearing
                                   # for the cross-compile matrix; the
                                   # `no-cgo` constraint already enforces
                                   # this and the goreleaser config sets
                                   # CGO_ENABLED=0 explicitly to make the
                                   # contract visible at build time.
  constraints:
    - no-cgo                       # goreleaser builds set CGO_ENABLED=0;
                                   # cross-compile target matrix relies
                                   # on this contract holding through
                                   # release time.
    - no-secrets-in-code           # HOMEBREW_TAP_GITHUB_TOKEN is a
                                   # repository secret; it is not
                                   # committed; the workflow YAML
                                   # references it via
                                   # ${{ secrets.HOMEBREW_TAP_GITHUB_TOKEN }}
                                   # only.
    - no-new-top-level-deps-without-decision
                                   # Stage notes: goreleaser + GitHub
                                   # Actions are tooling, not Go module
                                   # deps. Confirmed at design time:
                                   # this spec adds zero entries to
                                   # go.mod / go.sum. The constraint is
                                   # satisfied trivially.
    - one-spec-per-pr              # The PR for SPEC-023 references
                                   # exactly this spec; the doc sweep is
                                   # folded in per the premise-audit
                                   # rule (one spec ships its own
                                   # doc-shaped consequences).
    - test-before-implementation   # The `## Failing Tests` section
                                   # below adds ~33 shell asserts to
                                   # scripts/test-docs.sh BEFORE build
                                   # transcribes any artifact. Same
                                   # discipline as SPEC-021 / SPEC-022:
                                   # asserts are non-Go but the order
                                   # holds.
  related_specs:
    - SPEC-018                     # audit-grep cross-check addendum;
                                   # design has run + reconciled the
                                   # inherited SPEC-021 punch-list
                                   # greps and a SPEC-023-specific set
                                   # (see § Inherited SPEC-021 doc-sweep
                                   # audit + § SPEC-023 forward-reference
                                   # audit).
    - SPEC-019                     # NOT-contains self-audit pattern
                                   # (lighter surface here; few negative
                                   # asserts — see § NOT-contains
                                   # self-audit).
    - SPEC-020                     # negative-substring self-audit
                                   # codification reference.
    - SPEC-021                     # DIRECT precedent: created
                                   # scripts/test-docs.sh which this
                                   # spec extends; established the
                                   # cross-format §12 literal-artifact
                                   # pattern this spec applies to four
                                   # config artifacts; deferred the
                                   # `brew install jysf/bragfile/bragfile`
                                   # forward-reference and the
                                   # status-banner tense-flip into this
                                   # spec; deferred the inherited
                                   # punch-list of stale STAGE-NNN refs
                                   # into this spec.
    - SPEC-022                     # second confirming application of
                                   # §12 literal-artifact-as-spec
                                   # (cross-format: JSON + bash +
                                   # markdown + in-place markdown +
                                   # shell-script extension); SPEC-023
                                   # is the third confirming
                                   # application (now: YAML + workflow
                                   # YAML + CHANGELOG markdown +
                                   # six-line Go diff + shell-script
                                   # extension). Same scripts/test-docs.sh
                                   # extended further (groups L–P
                                   # appended after the SPEC-022 K-block).
    - SPEC-024                     # successor: shell completions; will
                                   # add a `brag completion` subcommand
                                   # but does NOT depend on SPEC-023's
                                   # distribution mechanics. Independent
                                   # specs sharing only the test-docs
                                   # harness.
---

# SPEC-023: distribution proper goreleaser gha changelog tap

## Bundle vs split decision (top-of-spec, decided 2026-04-26)

The stage's framing flagged this spec for a possible 2-way split into
SPEC-023a (`.goreleaser.yaml` + CI workflow) and SPEC-023b (release
workflow + tap-formula push + CHANGELOG). Design inspected the four
artifact bodies (drafted under § Notes for the Implementer below), the
~33-assert test-docs.sh extension, the six-file doc sweep, and the
end-to-end smoke-test plan, then **rejected the split**:

1. **Single ship narrative.** All four artifacts cohere under one
   sentence: *"`git tag v0.1.0 && git push origin v0.1.0` produces
   binaries on the GitHub release and a `bragfile.rb` formula on the
   tap repo, with a `[v0.1.0]` CHANGELOG entry attached."* That sentence
   does not split cleanly — every artifact is on the critical path of
   the smoke test. Splitting forces the SPEC-023a→SPEC-023b sequencing
   constraint where SPEC-023a's CI workflow has nothing meaningful to
   verify until SPEC-023b's release path proves end-to-end.

2. **Doc sweep activates as a bundle.** The README's brew-install line
   (currently a forward-reference at `README.md:19` flanked by *"Homebrew
   (recommended once available)"* at `README.md:16` and *"Distribution
   via Homebrew is in progress"* at `README.md:11–12`) goes from
   forward-reference to live link only when *all four* artifacts ship.
   Same for the AGENTS.md §3 *"(arriving in STAGE-005)"* tense flip
   (`AGENTS.md:67`). Splitting the spec means one of the two halves
   touches docs whose claims are still partially false.

3. **Size lands at solid M, not L.** AC count is 35 (not the >50
   L-threshold). New test-docs asserts: 33 (not the >40 split-trigger
   the framing implied — *"~60-80 new test-docs asserts"* was the
   conservative ceiling; 33 is comfortably under). New artifacts: 4
   files (`.goreleaser.yaml`, two `.github/workflows/*.yml`,
   `CHANGELOG.md`); modified files: 6 (cmd/brag/main.go,
   README.md, AGENTS.md, docs/architecture.md, docs/tutorial.md,
   scripts/test-docs.sh). All artifacts are fixed-shape and decided at
   design time per §12; build transcribes verbatim.

4. **§10 push-discipline + bfa1474 archive-spec precondition both pass
   trivially in either shape.** Splitting offers no operational
   advantage on the rule-set the stage has already absorbed.

5. **Goreleaser config and the release workflow are tightly coupled.**
   The release workflow exists to invoke `goreleaser release --clean`;
   `.goreleaser.yaml`'s `brews:` block presupposes the workflow exposes
   `HOMEBREW_TAP_GITHUB_TOKEN`. Co-locating them in one spec eliminates
   the cross-spec coordination cost of getting both right. Same
   reasoning applies to CHANGELOG.md — the literal artifact embeds a
   `[v0.1.0]` section that goes stale the moment it ships separately
   from the tagged release that creates v0.1.0.

**Decision: BUNDLE as one M spec.** This is the load-bearing spec of
STAGE-005 (per stage notes line ~261); the bundling is deliberate and
the §12 literal-artifact pattern carries the verification cost.

## Context

`brag` is feature-complete as of STAGE-004 ship (2026-04-25). STAGE-005
opened with two cleanup workstreams (SPEC-021 README rewrite shipped
2026-04-25; SPEC-022 AI-integration distribution asset shipped
2026-04-26) and now owes the **distribution mechanics** that
PROJ-001's success criterion 4 names: *"`brew install bragfile`
installs a working binary on macOS via a public homebrew tap."*

Today the path from "code on `main`" to "user types `brew install
bragfile`" is **manual and unwired**:

- No CI gate. PRs can merge red. `gofmt -l .` is enforced by convention
  but not by automation.
- No release artifacts. `go install ./cmd/brag` works locally; nothing
  produces cross-compiled binaries.
- No homebrew formula. The tap repo at
  `github.com/jysf/homebrew-bragfile` was created empty as a STAGE-005
  pre-stage chore (2026-04-25) and has not yet received a `bragfile.rb`.
- No CHANGELOG. Release notes have been the commit log up to now.
- The README at `README.md:11-12` carries the status banner *"in active
  development. … Distribution via Homebrew is in progress."*; line 16
  hedges *"Homebrew (recommended once available)"*; line 19 prints a
  brew-install command that does not yet work.
- AGENTS.md §3 line 67 says distribution is *"(arriving in STAGE-005)"*;
  AGENTS.md §4 lines 97 + 106 still say *"(STAGE-004)"* in stale
  pre-reshuffle tense.
- `cmd/brag/main.go:13` declares `const Version = "dev"` with the
  comment *"goreleaser injects the real version via ldflags in
  STAGE-004."* Two problems: (a) `const` cannot be overridden by `-X`
  ldflags — must be `var`; (b) STAGE-004 is the wrong stage tag.
- `docs/architecture.md:103` says *"Distribution (STAGE-004) uses
  goreleaser…"* — stale stage tag (the goreleaser work was reshuffled
  to STAGE-005 between framing rounds).
- `docs/architecture.md:45` says *"`internal/export` (STAGE-003)
  Markdown-report and sqlite-file-copy exporters."* — stale: the
  `--format sqlite` exporter was deferred to backlog 2026-04-23 (the
  STAGE-003 detail in `brief.md` notes the reasoning); the exporter
  list today is markdown + JSON.
- `docs/tutorial.md:493` carries `| brew install bragfile | STAGE-005 |`
  in a *"What's NOT there yet"* table — a row that needs to leave
  the table the moment SPEC-023 ships.

This spec ships the missing automation and folds in the doc sweep that
the missing automation has been blocking. It does **not** ship any new
`brag` CLI surface (chore, not story); the Go-code change is a
six-line const→var rewiring of the version-injection point in
`cmd/brag/main.go` so `goreleaser`'s `-X main.version=...` ldflag has
something to bind to.

This spec is **STAGE-005's third workstream** (SPEC-021 shipped
2026-04-25; SPEC-022 shipped 2026-04-26; SPEC-023 ships next; SPEC-024
shell completions follows). Per the stage's spec backlog ordering,
SPEC-023 is independent of SPEC-024 and depends only on SPEC-021's
`scripts/test-docs.sh` harness existing (it does, post-2026-04-25,
extended to 63 asserts at SPEC-022 ship 2026-04-26).

Parents:
- Project: `projects/PROJ-001-mvp/` (PROJ-001 — MVP wave; PROJ-001
  closes when STAGE-005 ships).
- Stage: `projects/PROJ-001-mvp/stages/STAGE-005-distribution-and-cleanup.md`
  (the stage; SPEC-023 is its third spec). Stage notes lock the
  artifact set (`STAGE-005:Spec Backlog:SPEC-023`, ~lines 240–249) and
  the structural recommendations
  (`STAGE-005:Design Notes:SPEC-023-specific`, ~lines 407–435).
- Repo: `bragfile`.

Stage-level locked decisions that bind this spec:
- **No new DECs expected.** Distribution mechanics are configuration,
  not project-binding decisions. Confirmed at design: every choice
  here is either constraint-driven (no-CGO ⇒ `CGO_ENABLED=0`,
  no-new-top-level-deps ⇒ no go.mod additions), industry-standard
  (Keep-A-Changelog 1.1.0 format, semver tag triggers, goreleaser v2
  conventions, GHA `actions/setup-go@v5` + `actions/checkout@v4`
  pinning), or pre-locked at stage level (tap repo URL, secret name,
  cross-compile matrix dimensions, smoke-test approach). All other
  choices document inline as Locked Design Decisions per spec
  convention. (`STAGE-005:Design Notes:Cross-cutting`, line ~277.)
- **Trim heuristic applies cautiously.** STAGE-005's specs are
  format-distinct; SPEC-023's literal artifacts (YAML / workflow YAML /
  Keep-A-Changelog markdown / six-line Go diff) get **NO trim** per the
  §12 pattern (verbatim embed required). Prose around them follows
  SPEC-021/022 fuller-skeleton convention since SPEC-023 is the first
  STAGE-005 spec whose primary artifacts are CI/build configuration.
  (`STAGE-005:Design Notes:Cross-cutting`, line ~286.)
- **Premise audit applicability:** asymmetric here. Most of SPEC-023 is
  new-file work (no premise-audit triggers); the doc sweep is the
  load-bearing audit surface — see § Inherited SPEC-021 doc-sweep
  audit + § SPEC-023 forward-reference audit below. (`STAGE-005:Design
  Notes:Cross-cutting`, line ~300.)
- **§9 BSD-grep `--exclude-dir` rule applies forward.** SPEC-023's
  test-docs.sh extension uses `grep -F` for some asserts; the
  `--exclude-dir` set is decorative on macOS, the `case`-statement
  post-filter is the correctness boundary. (`AGENTS.md §9`, line ~215,
  codified at SPEC-021 ship.)
- **§12 literal-artifact-as-spec pattern applies maximally.** Four
  fixed-shape artifacts decidable at design time (goreleaser config,
  CI workflow, release workflow, retroactive CHANGELOG); all four
  embedded verbatim under Notes for the Implementer. The
  goreleaser-generated `bragfile.rb` formula is **not** a SPEC-023
  artifact — it lives in the tap repo, generated by the release
  workflow at run time, never committed to this repo. (`AGENTS.md §12`,
  line ~314, codified at SPEC-021 ship — third confirming
  cross-format application.)
- **§10 push-discipline rule applies at build merge time.** Codified at
  STAGE-005 framing (2026-04-25); SPEC-021 was the first proactive
  application (held cleanly); SPEC-022 was the second (held cleanly);
  SPEC-023 is the third — promotes the rule from "codified at framing"
  to "load-bearing across the stage" without rewording.
  (`AGENTS.md §10`, line ~242.)
- **`bfa1474` archive-spec precondition** rejects empty
  `<answer>` placeholders in Reflection (Ship). Belt-and-suspenders
  for the SPEC-019 reflection-orphan class. Validated as PASSING on
  SPEC-021 + SPEC-022 ships; SPEC-023 will be the third confirmation.

## Goal

Ship a complete `git tag v0.1.0 → brew install jysf/bragfile/bragfile`
release pipeline by adding `.goreleaser.yaml`, two GitHub Actions
workflows (`ci.yml` for PR/main gating, `release.yml` for tag-driven
goreleaser invocation), a Keep-A-Changelog 1.1.0 `CHANGELOG.md`
populated retroactively for v0.1.0, and a six-line `cmd/brag/main.go`
const→var rewiring for `-X main.version=...` ldflags injection — with
the inherited SPEC-021 doc-sweep folded in (stale STAGE-NNN refs
across `docs/architecture.md` and `docs/tutorial.md`, distribution-
status tense flips in README.md and AGENTS.md, the stale STAGE-004
comment in `cmd/brag/main.go:13`) and shape verified by an extension
to `scripts/test-docs.sh` (new groups L–P, 33 new shell asserts)
exposed via the existing `just test-docs` recipe.

## Inputs

**Files to read (build-cycle reading list):**

- `AGENTS.md` — especially §3 (Tech stack — line 67's distribution
  forward-reference flips post-ship; line 97's STAGE-004 release-
  commands header is stale; line 106's STAGE-004 macOS-note is stale),
  §4 (Commands — release commands already include `goreleaser release
  --clean` and `goreleaser build --snapshot --clean` per line 97–98,
  build keeps these unchanged), §6 (Cycle Model), §7 (Cross-Reference
  Rules), §8 (Coding Conventions — applies to the six-line `cmd/brag/main.go`
  diff: `errors-wrap-with-context`, `no dead code`), §9 (Testing
  Conventions — four premise-audit addenda + audit-grep cross-check +
  BSD-grep `--exclude-dir` warning all apply forward), §10 (Git/PR
  Conventions — push-discipline rule applies at build merge), §11
  (Domain Glossary — line 258's `tap` entry is correct; no new
  glossary entries needed), §12 (Cycle-Specific Rules — design-time
  decision rule, NOT-contains self-audit, literal-artifact-as-spec
  pattern). Source of truth for the constraint set the goreleaser
  config and workflow YAMLs must satisfy.

- `cmd/brag/main.go` — current state. 32 lines. Line 13 declares
  `const Version = "dev"` with comment `"goreleaser injects the real
  version via ldflags in STAGE-004"`. The const cannot be overridden
  by ldflags; SPEC-023 changes it to `var version = "dev"` (lowercase
  to match the conventional goreleaser injection target
  `-X main.version=...`) and updates the comment to STAGE-005. Wire-
  through is unchanged: `cli.NewRootCmd(version)` passes the value
  into cobra's `Version:` field.

- `internal/cli/root.go` — current state. The cobra root command's
  `Version: version` field renders as `brag version <value>` on
  `--version`. No change in this file; SPEC-023 only flips the
  upstream `cmd/brag/main.go` declaration.

- `internal/cli/root_test.go` — current state. `TestRootCmd_VersionFlag`
  (lines 9–26) constructs `NewRootCmd("test-v0")` and asserts stdout
  contains `"test-v0"`. **The test is unaffected by SPEC-023's
  const→var change** (it constructs the cobra command directly, never
  reads `main.version`). No test changes in this file.

- `LICENSE` — MIT, repo root. Goreleaser's archives block embeds it.
  No content change.

- `README.md` — current state, 146 lines, post-SPEC-021 rewrite. Three
  status-claim lines need updating once SPEC-023 ships (see §
  SPEC-023 forward-reference audit):
  - Line 10–12: blockquote *"in active development. Capture, retrieve,
    search, export, and weekly/monthly digests are shipped. Distribution
    via Homebrew is in progress."* → flip "is in progress" to past
    tense.
  - Line 16: *"Homebrew (recommended once available):"* → drop the
    parenthetical hedge.
  - Line 19: code block `brew install jysf/bragfile/bragfile` is the
    forward-reference activated by this spec. No code change; tense
    flip happens around it.
  - Line 32–33: *"Requires Go 1.26+ from source. The Homebrew install
    pulls a prebuilt binary — no Go required."* — accurate post-ship,
    no change.

- `BRAG.md` — repo-root file targeting AI agents. Currently 317 lines
  post-SPEC-022 insertion. **No content change in this spec.** The
  schema/hook/slash-command cross-references it carries are correct
  forward through SPEC-023. SPEC-023 verifies (assert P9 below) that
  the K-block test-docs asserts still pass.

- `docs/architecture.md` — current state, 117 lines. Two lines need
  updating (see § Inherited SPEC-021 doc-sweep audit):
  - Line 45: *"`internal/export` (STAGE-003) Markdown-report and
    sqlite-file-copy exporters."* → *"Markdown-report and JSON
    exporters."* (sqlite exporter was deferred to backlog 2026-04-23.)
  - Line 103–106: *"Distribution (STAGE-004) uses goreleaser to produce
    macOS (arm64, x86_64) and Linux (arm64, x86_64) binaries; a
    homebrew tap at `github.com/jysf/homebrew-bragfile` ships the
    macOS ones via `brew install bragfile`."* → flip the parenthetical
    `(STAGE-004)` to `(STAGE-005)` and the brew-install command to the
    fully-qualified `brew install jysf/bragfile/bragfile` form.

- `docs/api-contract.md` — current state, 417 lines. **No content
  change in this spec.** Inherited audit confirms all 13 STAGE-NNN
  hits are retrospective stage-tags accurate as written (see §
  Inherited SPEC-021 doc-sweep audit § api-contract.md). Line 402's
  CHANGELOG mention (*"flag names may change between `v0.x` releases
  with CHANGELOG notes"*) becomes literally true once `CHANGELOG.md`
  ships — the prose is accurate ahead of and after this spec.

- `docs/data-model.md` — current state, 149 lines. **No content
  change in this spec.** Inherited audit confirms all 5 STAGE-NNN
  hits are retrospective stage-tags accurate as written (see §
  Inherited SPEC-021 doc-sweep audit § data-model.md).

- `docs/tutorial.md` — current state, 507 lines. **One row to strike**
  at line 493 (the `| brew install bragfile | STAGE-005 |` row inside
  the `## 9. What's NOT there yet` table). The four other STAGE-NNN
  hits in this file (lines 32, 34, 157, 200) are **example brag
  titles** (the user is brag-logging "shipped STAGE-001…" as the
  literal title) and STAY — they are example content, not status
  claims. Section 9 collapses to a one-line note ("Everything in this
  tutorial is shipped today.") since the table now has zero rows.

- `BRAG.md` — see above; no content change.

- `scripts/test-docs.sh` — current state, 598 lines, 63 asserts post-
  SPEC-022 (40 from SPEC-021 in groups A–G + F4 self-pass meta + 23
  from SPEC-022 in groups H–K). **SPEC-023 extends this script** with
  five new groups L–P appended after group K (33 new asserts;
  total 96). Per the script's docstring (line 2–4) it is the single
  doc-content-assertion script that grows internally as later
  STAGE-005 specs add asserts. The harness primitives
  (`assert_file_exists`, `assert_line_count_band`,
  `assert_contains_literal`, `assert_not_contains_iregex`,
  `check_link_target`) are reused unchanged; SPEC-023 adds one new
  primitive: `assert_yaml_has_top_key` (see § Notes for the
  Implementer).

- `justfile` — current shape, includes `test-docs` recipe added by
  SPEC-021. **No change in this spec** — the existing `just test-docs`
  recipe runs the extended script unchanged.

- `go.mod` — current state, declares `go 1.26.2`. The CI workflow's
  `setup-go` step uses `go-version-file: go.mod` so the matrix
  always tracks whatever go.mod declares. No content change here;
  the no-new-top-level-deps constraint applies and is satisfied
  trivially (this spec adds zero deps).

- `projects/PROJ-001-mvp/brief.md` — project context. Success
  criterion 4 (line ~48): *"`brew install bragfile` installs a working
  binary on macOS (at minimum) via a public homebrew tap."* SPEC-023
  delivers the mechanics; PROJ-001-close after STAGE-005 ship verifies
  the criterion against a live install.

- `projects/PROJ-001-mvp/stages/STAGE-005-distribution-and-cleanup.md`
  — parent stage. Locks the structural recommendations and the
  cross-cutting rules (~lines 252–319) and the SPEC-023-specific
  Design Notes (~lines 407–435) — every choice in those notes is
  either adopted directly here or extended with one of the locked
  design decisions below.

- `projects/PROJ-001-mvp/specs/done/SPEC-021-readme-user-facing-rewrite-and-dev-process-migration.md`
  — direct precedent within STAGE-005. Codified §12 literal-artifact-
  as-spec. Created `scripts/test-docs.sh`. Deferred the punch-list
  this spec activates.

- `projects/PROJ-001-mvp/specs/done/SPEC-022-ai-integration-distribution-asset.md`
  — second precedent within STAGE-005. First post-codification
  application of §12 literal-artifact-as-spec across multiple formats
  (JSON + bash + markdown + in-place markdown + shell-script
  extension). Validated cross-format transcription discipline.

- `decisions/DEC-001-pure-go-sqlite-driver.md` — load-bearing for the
  cross-compile matrix. The `no-cgo` constraint already prevents any
  drift; SPEC-023 makes the contract visible at goreleaser-build time
  via `env: [CGO_ENABLED=0]`.

**External APIs / tools:**

- **GoReleaser v2** — invoked by the release workflow via
  `goreleaser/goreleaser-action@v6`. Docs:
  https://goreleaser.com/customization/ . Version 2 syntax (`version: 2`
  at top of `.goreleaser.yaml`) is locked.
- **GitHub Actions** — workflows in `.github/workflows/`. Pinned action
  versions: `actions/checkout@v4`, `actions/setup-go@v5`,
  `goreleaser/goreleaser-action@v6`. Pinning by major matches
  industry convention; bumping to a pinned-by-sha posture is
  out-of-scope for PROJ-001 (revisit at PROJ-002 framing if a supply-
  chain decision becomes load-bearing).
- **Homebrew tap repo** at `github.com/jysf/homebrew-bragfile` — pre-
  created 2026-04-25 as a STAGE-005 framing chore. Contains a
  one-paragraph README pointing at the main repo. SPEC-023's
  goreleaser config writes a `bragfile.rb` formula file there at
  release time.

**Related code paths:**

- `cmd/brag/main.go` — six-line const→var change for ldflags
  injection (only Go change in this spec).
- `scripts/test-docs.sh` — extension only; no rewrite.
- `.github/workflows/` — directory does not exist today; SPEC-023
  creates it.

## Outputs

### Files created

- **`.goreleaser.yaml`** (repo root, NEW) — goreleaser v2 config.
  Cross-compile matrix: `{darwin, linux} × {amd64, arm64}`. Archive
  format: `tar.gz` with embedded `LICENSE`, `README.md`, `CHANGELOG.md`.
  Checksum: `checksums.txt` with sha256. Built-in changelog generator
  grouped by conventional-commit type for release notes. `brews:`
  block points at `homebrew-bragfile` with `skip_upload: auto` so
  prerelease tags do not push to the tap. ldflags inject `version`,
  `commit`, `date` into `main`. Literal verbatim under § Notes for
  the Implementer.

- **`.github/workflows/ci.yml`** (NEW) — CI workflow. Triggers on
  `pull_request` (any) and `push` to `main`. Matrix: `{macos-latest,
  ubuntu-latest}`. Steps: checkout, `setup-go` reading `go.mod` for
  the version, `gofmt -l .` (fail if non-empty), `go vet ./...`,
  `go test ./...`. `permissions: contents: read` (least privilege).
  Literal verbatim under § Notes for the Implementer.

- **`.github/workflows/release.yml`** (NEW) — release workflow.
  Triggers on tag push matching `v*` (semver tags + prerelease tags
  both accepted; goreleaser distinguishes via `prerelease: auto`).
  Single-job `ubuntu-latest`. Steps: checkout with `fetch-depth: 0`
  (goreleaser needs full git history for changelog), `setup-go`
  reading `go.mod`, `goreleaser/goreleaser-action@v6` running
  `release --clean`. Env: `GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}`
  for the GitHub release upload + `HOMEBREW_TAP_GITHUB_TOKEN: ${{
  secrets.HOMEBREW_TAP_GITHUB_TOKEN }}` for the tap-formula push.
  `permissions: contents: write` (needs to create the GitHub
  release). Literal verbatim under § Notes for the Implementer.

- **`CHANGELOG.md`** (repo root, NEW) — Keep-A-Changelog 1.1.0
  format. Two sections at ship time: `## [Unreleased]` (empty placeholder
  for post-v0.1.0 commits) + `## [0.1.0] - YYYY-MM-DD` populated
  retroactively from PROJ-001's commit history (commits `7f802a2`
  forward — i.e. STAGE-001 onward — folded into Added /
  Decisions-of-record subsections). The date placeholder
  `YYYY-MM-DD` is filled in at build time when the v0.1.0 tag is
  cut (the rest of the file is verbatim from § Notes for the
  Implementer; only that one date string is build-time).

### Files modified

- **`cmd/brag/main.go`** — six-line diff (see § Notes for the
  Implementer for the literal):
  - Line 13: `const Version = "dev"` → `var version = "dev"`
    (lowercase `version` matches goreleaser's conventional `-X
    main.version=...` injection point).
  - Line 14 comment: `"injects the real version via ldflags in
    STAGE-004"` → `"injects the real version via ldflags in
    STAGE-005"`. (Two-line comment on lines 11–13; the substantive
    update is "STAGE-004" → "STAGE-005" plus const→var.)
  - Line 18: `cli.NewRootCmd(Version)` → `cli.NewRootCmd(version)`
    (rename ripple).
  - Two new `var` lines added alongside `version`: `commit = "none"`
    and `date = "unknown"` (goreleaser's conventional ldflag targets;
    unused at runtime today but available for a future
    `--version --verbose` enhancement without re-litigating the
    injection wiring).

- **`README.md`** — three tense flips (literal diffs in § Notes for
  the Implementer):
  - Lines 10–12 (status banner): drop *"Distribution via Homebrew is
    in progress."*; rewrite the blockquote to acknowledge the v0.1.0
    release.
  - Line 16: drop *"(recommended once available)"* parenthetical;
    `## Install` heading reads `### Homebrew` directly under it.
  - Line 32: replace forward-pointing wording — *"Requires Go 1.26+
    from source. The Homebrew install pulls a prebuilt binary — no
    Go required."* — with crisper present-tense wording.

- **`AGENTS.md`** — three line edits (literal diffs in § Notes for
  the Implementer):
  - Line 67 (§3 Tech Stack): *"`goreleaser` → GitHub Releases →
    homebrew tap at `github.com/jysf/homebrew-bragfile` (arriving in
    STAGE-005)."* → drop the parenthetical; the line is now a
    statement of fact.
  - Line 97 (§4 Commands): *"# --- release (STAGE-004) ---"* → *"#
    --- release (STAGE-005) ---"*.
  - Line 106 (§4 Commands): *"`brew install goreleaser` once we hit
    STAGE-004."* → *"`brew install goreleaser` before tagging a
    release."* (deletes the stage tag entirely; the surrounding macOS
    note is a contributor instruction, not a stage-tracking claim).

- **`docs/architecture.md`** — two line edits (literal diffs in §
  Notes for the Implementer):
  - Line 45 (Components table): *"`internal/export` (STAGE-003)
    Markdown-report and sqlite-file-copy exporters."* → *"`internal/export`
    (STAGE-003) Markdown-report and JSON exporters."* The
    sqlite-file-copy exporter was deferred to backlog on 2026-04-23
    (see `brief.md` lines 44–47 + 79–85 for the deferral rationale);
    the line as written is wrong. JSON exporter shipped in SPEC-014.
  - Line 103: *"Distribution (STAGE-004) uses goreleaser…"* →
    *"Distribution (STAGE-005) uses goreleaser…"*.
  - Line 106 (same paragraph): *"`brew install bragfile`"* → *"`brew
    install jysf/bragfile/bragfile`"* (the fully-qualified form, since
    the tap is third-party — this matches README.md:19 and
    AGENTS.md:67).

- **`docs/tutorial.md`** — one section edit (literal diff in § Notes
  for the Implementer):
  - Lines 487–496 (§9 *"What's NOT there yet"*): strike the row `|
    brew install bragfile | STAGE-005 |` and the surrounding table,
    replacing the section body with a one-line note that everything
    in the tutorial is shipped today. Section heading stays.

- **`scripts/test-docs.sh`** — five new groups L–P appended after
  group K. 33 new asserts. One new primitive
  (`assert_yaml_has_top_key`). Literal verbatim under § Notes for
  the Implementer.

### Inherited SPEC-021 doc-sweep audit (RAN at design time, 2026-04-26)

Per AGENTS.md §9 audit-grep cross-check (codified at SPEC-018 ship):
the design-time greps RAN against the live repo and reconciled below.

**Grep:** `grep -n "STAGE-00[0-9]" docs/api-contract.md docs/architecture.md docs/data-model.md docs/tutorial.md`

**Hits and disposition:**

| File | Line | Hit | Disposition |
|---|---|---|---|
| `api-contract.md` | 33 | `**STAGE-001 (flags-only form):**` | STAYS — retrospective stage-tag, accurate as written |
| `api-contract.md` | 50 | `**STAGE-002 (editor-launch form):**` | STAYS — retrospective |
| `api-contract.md` | 80 | `**STAGE-003 (JSON stdin form):**` | STAYS — retrospective |
| `api-contract.md` | 134–135 | `STAGE-001 ships without filter flags … added in STAGE-002.` | STAYS — retrospective |
| `api-contract.md` | 137 | `### \`brag show <id>\` — show a single entry (STAGE-002)` | STAYS — retrospective header tag |
| `api-contract.md` | 146 | `### \`brag edit <id>\` — edit via $EDITOR (STAGE-002)` | STAYS — retrospective |
| `api-contract.md` | 173 | `### \`brag delete <id>\` — delete an entry (STAGE-002)` | STAYS — retrospective |
| `api-contract.md` | 186 | `### \`brag search "query"\` — full-text search (STAGE-002)` | STAYS — retrospective |
| `api-contract.md` | 215 | `### \`brag export\` — export entries (STAGE-003)` | STAYS — retrospective |
| `api-contract.md` | 251 | `### \`brag summary --range week\|month\` (STAGE-004)` | STAYS — retrospective |
| `api-contract.md` | 287 | `### \`brag review --week \| --month\` (STAGE-004)` | STAYS — retrospective |
| `api-contract.md` | 334 | `### \`brag stats\` (STAGE-004)` | STAYS — retrospective |
| `architecture.md` | 24 | Mermaid `shipped in STAGE-002 … export shipped in STAGE-003 … summary STAGE-004` | STAYS — retrospective component-history annotation |
| `architecture.md` | 44 | `\`internal/editor\` \| (STAGE-002) Launches \`$EDITOR\`…` | STAYS — retrospective |
| `architecture.md` | **45** | `\`internal/export\` \| (STAGE-003) Markdown-report and sqlite-file-copy exporters.` | **UPDATE** — drop "sqlite-file-copy", say "Markdown-report and JSON exporters" (sqlite-file-copy deferred to backlog 2026-04-23) |
| `architecture.md` | **103** | `Distribution (STAGE-004) uses goreleaser…` | **UPDATE** — STAGE-004 → STAGE-005 (stage was reshuffled between framing rounds) |
| `data-model.md` | 6 | `STAGE-002 adds a third virtual table (\`entries_fts\`, SPEC-011)` | STAYS — retrospective |
| `data-model.md` | 89 | `Planned for STAGE-001 (ship with the initial migration):` | STAYS — retrospective |
| `data-model.md` | 94 | `supports \`list --project=...\` filter (flags land in STAGE-002).` | STAYS — retrospective |
| `data-model.md` | 96 | `Shipped in STAGE-002 (SPEC-011):` | STAYS — retrospective |
| `data-model.md` | 111 | `\`brag edit\` (STAGE-002) opens \`$EDITOR\` on a round-tripped` | STAYS — retrospective |
| `tutorial.md` | 32 | `brag add --title "shipped STAGE-001 end-to-end in a day"` | STAYS — example brag title (not a status claim) |
| `tutorial.md` | 34 | `brag add -t "shipped STAGE-001 end-to-end in a day"` | STAYS — example brag title |
| `tutorial.md` | 157 | `2 2026-04-20T21:32:41Z shipped STAGE-001 end-to-end in a day` | STAYS — example list output |
| `tutorial.md` | 200 | `# 2 2026-04-20T21:32:41Z platform shipped STAGE-001 end-to-end in a day` | STAYS — example list output |
| `tutorial.md` | **493** | `\| \`brew install bragfile\` \| STAGE-005 \|` | **UPDATE** — strike the row + collapse §9 to a one-line note |

**Tally:** 27 hits, **3 updates** (architecture.md:45, architecture.md:103, tutorial.md:493 + section-9 collapse). Reconciles to the session-log claim "13 + 4 + 5 + 1 = 23 hits" with one delta: api-contract.md has 13 hits (matches), architecture.md has 4 hits (matches; updates touch 2 of 4 — line 45 + line 103 + an adjacent line 106 wording fix to the brew-install command), data-model.md has 5 hits (matches; 0 updates), tutorial.md has **5** hits not 1 (the 4 example-content hits at lines 32/34/157/200 were not in the session-log count; they are STAYS regardless and the tally adjusts upward without changing dispositions).

### SPEC-023 forward-reference audit (RAN at design time, 2026-04-26)

Per AGENTS.md §9 audit-grep cross-check: a second design-time grep
sweeps the distribution-shipping forward-references that ACTIVATE
when SPEC-023 ships (rather than the inherited STAGE-NNN sweep above).

**Grep:** `grep -n -E "brew install|homebrew|goreleaser|CHANGELOG" docs/api-contract.md docs/architecture.md docs/data-model.md docs/tutorial.md README.md AGENTS.md BRAG.md CONTRIBUTING.md docs/development.md cmd/brag/main.go`

**Hits and disposition:**

| File | Line | Hit | Disposition |
|---|---|---|---|
| `tutorial.md` | 493 | `\| \`brew install bragfile\` \| STAGE-005 \|` | **UPDATE** — already covered above |
| `architecture.md` | 59 | `**No CGO.** … goreleaser can` | STAYS — accurate principle statement |
| `architecture.md` | 103 | `Distribution (STAGE-004) uses goreleaser` | **UPDATE** — already covered above |
| `architecture.md` | 104–106 | `homebrew tap at github.com/jysf/homebrew-bragfile` + `brew install bragfile` | **UPDATE** — `brew install bragfile` → `brew install jysf/bragfile/bragfile` (fully-qualified form) |
| `api-contract.md` | 10 | `**Binary:** \`brag\` (homebrew formula: \`bragfile\`).` | STAYS — accurate |
| `api-contract.md` | 402 | `flag names may change between \`v0.x\` releases with CHANGELOG notes` | STAYS — becomes literally true once CHANGELOG.md ships |
| `README.md` | 11 | blockquote: `Distribution via Homebrew is in progress.` | **UPDATE** — past tense |
| `README.md` | 16 | `Homebrew (recommended once available):` | **UPDATE** — drop parenthetical |
| `README.md` | 19 | `brew install jysf/bragfile/bragfile` | STAYS — code unchanged; surrounding tense flips |
| `README.md` | 32 | `Requires Go 1.26+ from source. The Homebrew install pulls a prebuilt binary — no Go required.` | STAYS — accurate |
| `AGENTS.md` | 13 | `Bragfile (\`brag\` CLI, homebrew formula \`bragfile\`)` | STAYS — accurate |
| `AGENTS.md` | 16 | `… ship a usable, distributable MVP via \`brew install bragfile\` within ~2 weeks.` | STAYS — accurate retrospective |
| `AGENTS.md` | 67 | `goreleaser → GitHub Releases → homebrew tap at github.com/jysf/homebrew-bragfile (arriving in STAGE-005).` | **UPDATE** — drop "(arriving in STAGE-005)" |
| `AGENTS.md` | 97 | `# --- release (STAGE-004) ---` | **UPDATE** — STAGE-004 → STAGE-005 |
| `AGENTS.md` | 98 | `goreleaser release --clean` | STAYS — accurate command |
| `AGENTS.md` | 106 | `\`brew install goreleaser\` once we hit STAGE-004.` | **UPDATE** — drop stage tag, change wording |
| `AGENTS.md` | 258 | `tap — a homebrew tap repo (github.com/jysf/homebrew-bragfile) hosting the bragfile.rb formula. Created in STAGE-005.` | STAYS — accurate retrospective |
| `BRAG.md` | (none) | — | (no hits) |
| `CONTRIBUTING.md` | (none) | — | (no hits) |
| `docs/development.md` | (none) | — | (no hits) |
| `cmd/brag/main.go` | 12 | `// Version is set to "dev" for local builds. goreleaser injects the real // version via ldflags in STAGE-004.` | **UPDATE** — STAGE-004 → STAGE-005 (folded into the const→var diff) |

**Tally:** 21 hits, **8 updates** (tutorial.md:493 already counted in the inherited audit; architecture.md:103 already counted; architecture.md:104–106 brew-install-form fix is new in this audit; README.md:11 + 16 are new; AGENTS.md:67 + 97 + 106 are new; cmd/brag/main.go:12 is new). Net unique updates across both audits: **3 (inherited) + 6 (new) = 9 lines edited across 6 files** (cmd/brag/main.go, README.md, AGENTS.md, docs/architecture.md, docs/tutorial.md, plus the four NEW files).

### NOT-contains self-audit (per AGENTS.md §12, 2026-04-26)

The Failing Tests below contain a handful of NOT-contains assertions
(group P P3, P5, P6, P7, P8, P10). Per the §12 NOT-contains
self-audit rule: grep this spec's load-bearing prose for the
forbidden tokens to confirm they do not appear in any artifact this
spec embeds verbatim.

**Forbidden tokens (from group P assertions) and their load-bearing
locations:**

| Token | Asserted absent from | Load-bearing prose check |
|---|---|---|
| `arriving` (case-insensitive, in AGENTS.md context) | `AGENTS.md` | The literal §3 line 67 edit drops "(arriving in STAGE-005)"; the embedded literal under § Notes for the Implementer does NOT contain "arriving". ✓ |
| `(STAGE-004)` (literal, in AGENTS.md release-commands context) | `AGENTS.md` | The literal §4 line 97 edit changes `(STAGE-004)` → `(STAGE-005)`. The embedded literal does NOT contain `(STAGE-004)` after the edit. ✓ |
| `sqlite-file-copy` | `docs/architecture.md` | The literal line-45 edit changes "sqlite-file-copy exporters" → "JSON exporters". The embedded literal does NOT contain `sqlite-file-copy`. ✓ |
| `Distribution (STAGE-004)` | `docs/architecture.md` | The literal line-103 edit changes the parenthetical. The embedded literal does NOT contain `Distribution (STAGE-004)`. ✓ |
| `brew install bragfile` (under "What's NOT" heading section) | `docs/tutorial.md` | The literal §9 collapse drops the row entirely; the section body becomes a one-line note. The embedded literal does NOT contain "brew install bragfile" anywhere in the §9 body. ✓ |
| `STAGE-004` (within `cmd/brag/main.go`) | `cmd/brag/main.go` | The literal six-line diff includes a comment update from STAGE-004 → STAGE-005. The post-edit file does NOT contain `STAGE-004`. ✓ |
| `in active development` | `README.md` | The literal status-banner edit removes the "in active development" wording entirely. The post-edit blockquote does NOT contain that token. ✓ |

All seven forbidden tokens validated absent from the post-edit literals
embedded under § Notes for the Implementer. **Zero hits at design**;
zero false-positive risk at build per the §12 codified rule.

### Files NOT modified by this spec (deferred — explicitly enumerated)

- `docs/api-contract.md` — all 13 STAGE-NNN hits are retrospective and
  accurate; no content change. Line 402's CHANGELOG mention becomes
  literally true post-ship without prose change.
- `docs/data-model.md` — all 5 STAGE-NNN hits retrospective; no
  content change.
- `BRAG.md` — assertions K1–K4 from SPEC-022 still pass; no content
  change. Group P9 assertion (below) verifies.
- `CONTRIBUTING.md` — no relevant content; no change.
- `docs/development.md` — no relevant content; no change.
- `internal/cli/root.go`, `internal/cli/root_test.go` — `--version`
  rendering wires through unchanged; existing test passes against
  the const→var change without modification.
- `justfile` — `test-docs` recipe added by SPEC-021 runs the extended
  script unchanged.
- `go.mod`, `go.sum` — zero deps added (no-new-top-level-deps
  constraint satisfied trivially).
- `scripts/new-spec.sh`, `scripts/archive-spec.sh`, etc. — unchanged.
- `docs/CONTEXTCORE_ALIGNMENT.md` — unchanged (no distribution-related
  claims).
- `GETTING_STARTED.md`, `FIRST_SESSION_PROMPTS.md` — unchanged
  (contributor onboarding; no distribution claims).

### Manual GitHub UI steps (capture in ship checklist; NOT in this spec's literals)

These are NOT artifacts SPEC-023 ships; they are operations the
human performs during build/ship. The spec's ship checklist
(post-build, pre-archive) names them:

1. **Create the `HOMEBREW_TAP_GITHUB_TOKEN` repository secret.**
   Settings → Secrets and variables → Actions → New repository
   secret. Name: `HOMEBREW_TAP_GITHUB_TOKEN`. Value: a GitHub
   personal access token (classic or fine-grained) with `repo` scope
   (classic) or "Contents: read+write" + "Metadata: read" (fine-
   grained) on `github.com/jysf/homebrew-bragfile` only.
2. **Enable branch protection on `main`** requiring the `ci`
   workflow's status checks to pass before PRs merge. Settings →
   Branches → Branch protection rules → Add rule for `main` → Require
   status checks to pass before merging → search for `test
   (macos-latest)` and `test (ubuntu-latest)` after the workflow has
   run at least once on a PR.
3. **Verify the tap repo's GitHub Pages is NOT enabled.** GoReleaser
   pushes the formula to `main`; consumers fetch from `raw.githubusercontent.com`
   via `brew tap`. Pages is not in the path.
4. **Smoke test pre-flight.** Tag `v0.1.0-rc1`, push, verify the
   release workflow runs end-to-end and goreleaser correctly
   identifies `0.1.0-rc1` as a prerelease (`brews.skip_upload: auto`
   skips the tap-formula push). Delete the prerelease GitHub release
   + the `v0.1.0-rc1` tag after verification. Then tag v0.1.0 proper.

## Acceptance Criteria

Testable outcomes. Cover happy path + error cases + edge cases.

**Artifacts created:**
- [ ] AC-1: `.goreleaser.yaml` exists at repo root and `goreleaser
      check` exits 0.
- [ ] AC-2: `.github/workflows/ci.yml` exists; YAML parses; declares
      a `test` job with a `{macos-latest, ubuntu-latest}` matrix.
- [ ] AC-3: `.github/workflows/release.yml` exists; YAML parses;
      triggers on tag push matching `v*`.
- [ ] AC-4: `CHANGELOG.md` exists at repo root and conforms to the
      Keep-A-Changelog 1.1.0 reference structure (link to the spec at
      the top of the file; `## [Unreleased]` and `## [0.1.0] -
      YYYY-MM-DD` sections; `[Unreleased]` and `[0.1.0]` link
      references at the bottom).

**Goreleaser config — content:**
- [ ] AC-5: `.goreleaser.yaml` opens with `version: 2` (goreleaser v2
      schema lock).
- [ ] AC-6: `.goreleaser.yaml` declares `CGO_ENABLED=0` in the
      `builds[].env` block (no-cgo constraint visible at build time).
- [ ] AC-7: `.goreleaser.yaml` declares both `darwin` and `linux` in
      `builds[].goos` and both `amd64` and `arm64` in
      `builds[].goarch`.
- [ ] AC-8: `.goreleaser.yaml` injects `version`, `commit`, `date`
      via `-X main.<name>=` ldflags.
- [ ] AC-9: `.goreleaser.yaml` declares `archives[].format: tar.gz`
      and embeds `LICENSE`, `README.md`, `CHANGELOG.md` in
      `archives[].files`.
- [ ] AC-10: `.goreleaser.yaml` declares `checksum.algorithm: sha256`
      and `checksum.name_template: checksums.txt`.
- [ ] AC-11: `.goreleaser.yaml` declares a `brews:` block with
      `repository.owner: jysf`, `repository.name: homebrew-bragfile`,
      `skip_upload: auto`, `license: MIT`, and a `test:` snippet
      invoking `brag --version`.
- [ ] AC-12: `.goreleaser.yaml`'s `changelog:` block uses
      `use: github` with conventional-commit-grouped sort.

**CI workflow — content:**
- [ ] AC-13: `.github/workflows/ci.yml` triggers on `pull_request`
      AND `push` to `main`.
- [ ] AC-14: `.github/workflows/ci.yml` runs `gofmt -l .` AND fails
      if the output is non-empty.
- [ ] AC-15: `.github/workflows/ci.yml` runs `go vet ./...`.
- [ ] AC-16: `.github/workflows/ci.yml` runs `go test ./...`.
- [ ] AC-17: `.github/workflows/ci.yml` uses `actions/setup-go@v5`
      with `go-version-file: go.mod` (no hard-pinned version, so
      latest-stable-go updates flow through go.mod bumps).
- [ ] AC-18: `.github/workflows/ci.yml` declares
      `permissions: contents: read` (least privilege).

**Release workflow — content:**
- [ ] AC-19: `.github/workflows/release.yml` triggers on tag push
      pattern `v*`.
- [ ] AC-20: `.github/workflows/release.yml` uses
      `goreleaser/goreleaser-action@v6` with `args: release --clean`.
- [ ] AC-21: `.github/workflows/release.yml` exposes
      `HOMEBREW_TAP_GITHUB_TOKEN` to the goreleaser step via
      `env:`/`secrets`.
- [ ] AC-22: `.github/workflows/release.yml` uses checkout with
      `fetch-depth: 0` (goreleaser needs full git history).
- [ ] AC-23: `.github/workflows/release.yml` declares
      `permissions: contents: write`.

**CHANGELOG — content:**
- [ ] AC-24: `CHANGELOG.md` lists `brag add`, `brag list`, `brag
      show`, `brag edit`, `brag delete`, `brag search`, `brag
      export`, `brag summary`, `brag review`, `brag stats` under the
      `### Added` subsection of `[0.1.0]`.
- [ ] AC-25: `CHANGELOG.md` lists DEC-001 through DEC-014 under a
      `### Decisions of record` subsection of `[0.1.0]`.
- [ ] AC-26: `CHANGELOG.md` ends with two link reference definitions
      `[Unreleased]: …compare/v0.1.0...HEAD` and `[0.1.0]:
      …releases/tag/v0.1.0`.

**`cmd/brag/main.go` change:**
- [ ] AC-27: `cmd/brag/main.go` declares `var version = "dev"` (NOT
      `const Version`); the existing `internal/cli` test
      `TestRootCmd_VersionFlag` still passes (it constructs
      `NewRootCmd("test-v0")` directly).
- [ ] AC-28: `cmd/brag/main.go` declares `var commit = "none"` and
      `var date = "unknown"` alongside `version` (placeholders for
      the goreleaser ldflag targets; unused at runtime today).
- [ ] AC-29: `cmd/brag/main.go` no longer contains the literal token
      `STAGE-004` anywhere.

**Doc sweep:**
- [ ] AC-30: `README.md` line 11–12 status banner does NOT contain
      "in progress" or "in active development"; reflects the v0.1.0
      release.
- [ ] AC-31: `AGENTS.md` line ~67 does NOT contain "arriving in
      STAGE-005".
- [ ] AC-32: `AGENTS.md` no longer contains the literal token
      `(STAGE-004)` anywhere.
- [ ] AC-33: `docs/architecture.md` line ~45 does NOT contain
      "sqlite-file-copy" and DOES contain "JSON".
- [ ] AC-34: `docs/architecture.md` no longer contains the literal
      token `Distribution (STAGE-004)`.
- [ ] AC-35: `docs/tutorial.md` §9 *"What's NOT there yet"* body
      does NOT contain `brew install`.

**Test harness:**
- [ ] AC-36: `just test-docs` exits 0 against the post-build state
      (96 asserts pass: 40 SPEC-021 groups A–G + F4 + 23 SPEC-022
      groups H–K + 33 SPEC-023 groups L–P).

**End-to-end (smoke; performed at ship-checklist time, not in CI):**
- [ ] AC-37: `goreleaser build --snapshot --clean` exits 0 locally on
      the user's macOS arm64 box and produces archives for all four
      `{darwin, linux} × {amd64, arm64}` targets.
- [ ] AC-38: A pushed `v0.1.0-rc1` tag triggers `release.yml`; the
      workflow exits 0; goreleaser publishes a prerelease GitHub
      release with all four archives + `checksums.txt`; goreleaser
      does NOT push a formula to the tap repo (`skip_upload: auto`
      detected the prerelease).
- [ ] AC-39: A pushed `v0.1.0` tag triggers `release.yml`; the
      workflow exits 0; goreleaser publishes a stable GitHub release
      with all four archives + `checksums.txt`; goreleaser pushes
      `Formula/bragfile.rb` to `github.com/jysf/homebrew-bragfile@main`.
- [ ] AC-40: From a clean shell on macOS arm64 without `brag` on
      `$PATH`, `brew install jysf/bragfile/bragfile` exits 0; `brag
      --version` prints `brag version 0.1.0` (the ldflag-injected
      string, not "dev"); `brag add --title "x" && brag list` round-
      trips against `~/.bragfile/db.sqlite`.

(AC-37 through AC-40 are the smoke tests; they ARE acceptance
criteria but are performed against the live release workflow at
ship time, not asserted by the test-docs.sh harness. They go on the
ship-checklist sub-list under § Notes for the Implementer.)

## Failing Tests

Written during **design**, BEFORE build. The implementer's job in
**build** is to make these pass by transcribing the literal artifacts
under § Notes for the Implementer verbatim.

All 33 new asserts append to `scripts/test-docs.sh` after group K,
following the SPEC-022 cross-format precedent (groups H–K added 23
asserts in 161 lines on top of SPEC-021's 437-line baseline). Group
identifiers L (goreleaser config), M (CI workflow), N (release
workflow), O (CHANGELOG), P (doc sweep + tense flips). The harness
primitives (`assert_file_exists`, `assert_line_count_band`,
`assert_contains_literal`, `assert_not_contains_iregex`,
`check_link_target`) are reused unchanged; SPEC-023 adds one new
primitive (`assert_yaml_has_top_key`) for goreleaser-block
structural checks.

- **`scripts/test-docs.sh` (extension; verbatim under § Notes for
  the Implementer)**

  - **Group L — `.goreleaser.yaml` shape (10 asserts).**
    - `"L1"` — file exists at `.goreleaser.yaml`. asserts:
      `assert_file_exists "L1" ".goreleaser.yaml"`.
    - `"L2"` — opens with `version: 2`. asserts:
      `head -n 5 of file matches '^version: 2$'`.
    - `"L3"` — declares `CGO_ENABLED=0`. asserts:
      `assert_contains_literal "L3" ".goreleaser.yaml" "CGO_ENABLED=0"`.
    - `"L4"` — declares both `darwin` and `linux` goos values.
      asserts: two `assert_contains_literal` checks (or one combined).
    - `"L5"` — declares both `amd64` and `arm64` goarch values.
      asserts: two `assert_contains_literal` checks.
    - `"L6"` — declares a top-level `brews:` block. asserts:
      `assert_yaml_has_top_key "L6" ".goreleaser.yaml" "brews"`.
    - `"L7"` — brews block points at `homebrew-bragfile`. asserts:
      `assert_contains_literal "L7" ".goreleaser.yaml" "name: homebrew-bragfile"`.
    - `"L8"` — brews block has `skip_upload: auto`. asserts:
      `assert_contains_literal "L8" ".goreleaser.yaml" "skip_upload: auto"`.
    - `"L9"` — declares `-X main.version=` ldflag. asserts:
      `assert_contains_literal "L9" ".goreleaser.yaml" "-X main.version="`.
    - `"L10"` — archive format is `tar.gz`. asserts:
      `assert_contains_literal "L10" ".goreleaser.yaml" "format: tar.gz"`.

  - **Group M — `.github/workflows/ci.yml` shape (8 asserts).**
    - `"M1"` — file exists. asserts:
      `assert_file_exists "M1" ".github/workflows/ci.yml"`.
    - `"M2"` — triggers on `pull_request`. asserts:
      `assert_contains_literal "M2" ".github/workflows/ci.yml" "pull_request:"`.
    - `"M3"` — triggers on push to `main`. asserts: shell grep
      checks both `push:` and a following `branches:\n      - main`
      block (use a single `grep -A` with a follow-on assertion).
    - `"M4"` — matrix includes `macos-latest`. asserts:
      `assert_contains_literal "M4" ".github/workflows/ci.yml" "macos-latest"`.
    - `"M5"` — matrix includes `ubuntu-latest`. asserts:
      `assert_contains_literal "M5" ".github/workflows/ci.yml" "ubuntu-latest"`.
    - `"M6"` — runs `go test ./...`. asserts:
      `assert_contains_literal "M6" ".github/workflows/ci.yml" "go test ./..."`.
    - `"M7"` — runs `gofmt -l .`. asserts:
      `assert_contains_literal "M7" ".github/workflows/ci.yml" "gofmt -l ."`.
    - `"M8"` — runs `go vet ./...`. asserts:
      `assert_contains_literal "M8" ".github/workflows/ci.yml" "go vet ./..."`.

  - **Group N — `.github/workflows/release.yml` shape (5 asserts).**
    - `"N1"` — file exists. asserts:
      `assert_file_exists "N1" ".github/workflows/release.yml"`.
    - `"N2"` — triggers on tag push pattern `v*`. asserts:
      shell-grep for the literal `tags:\n      - 'v*'` shape.
    - `"N3"` — uses `goreleaser/goreleaser-action@v6`. asserts:
      `assert_contains_literal "N3" ".github/workflows/release.yml" "goreleaser/goreleaser-action@v6"`.
    - `"N4"` — passes `HOMEBREW_TAP_GITHUB_TOKEN` env. asserts:
      `assert_contains_literal "N4" ".github/workflows/release.yml" "HOMEBREW_TAP_GITHUB_TOKEN"`.
    - `"N5"` — checkout uses `fetch-depth: 0`. asserts:
      `assert_contains_literal "N5" ".github/workflows/release.yml" "fetch-depth: 0"`.

  - **Group O — `CHANGELOG.md` shape (5 asserts).**
    - `"O1"` — file exists. asserts:
      `assert_file_exists "O1" "CHANGELOG.md"`.
    - `"O2"` — references Keep-A-Changelog. asserts:
      `assert_contains_literal "O2" "CHANGELOG.md" "keepachangelog.com"`.
    - `"O3"` — has `## [0.1.0]` heading. asserts: line-equality check
      (shell-grep with `^## \[0\.1\.0\]`).
    - `"O4"` — lists each shipped command verb under Added. asserts:
      ten `assert_contains_literal` checks for `\`brag add\``,
      `\`brag list\``, `\`brag show\``, `\`brag edit\``, `\`brag
      delete\``, `\`brag search\``, `\`brag export\``, `\`brag
      summary\``, `\`brag review\``, `\`brag stats\``. (Implemented
      as a single shell loop over a list; counts as one named
      assertion.)
    - `"O5"` — has `[Unreleased]` and `[0.1.0]` link reference
      definitions at file end. asserts: two `assert_contains_literal`
      checks for `[Unreleased]:` and `[0.1.0]:` (counts as one named
      assertion).

  - **Group P — Doc sweep + tense flips (10 asserts).**
    - `"P1"` — README.md status banner does NOT contain "in
      progress". asserts:
      `assert_not_contains_iregex "P1" "README.md" "in progress"`.
    - `"P2"` — README.md status banner does NOT contain "in active
      development". asserts:
      `assert_not_contains_iregex "P2" "README.md" "in active development"`.
    - `"P3"` — AGENTS.md does NOT contain "arriving in STAGE". asserts:
      `assert_not_contains_iregex "P3" "AGENTS.md" "arriving in STAGE"`.
    - `"P4"` — AGENTS.md does NOT contain literal `(STAGE-004)`.
      asserts: `assert_not_contains_iregex "P4" "AGENTS.md" "\\(STAGE-004\\)"`.
    - `"P5"` — docs/architecture.md does NOT contain "sqlite-file-
      copy". asserts:
      `assert_not_contains_iregex "P5" "docs/architecture.md" "sqlite-file-copy"`.
    - `"P6"` — docs/architecture.md does NOT contain "Distribution
      (STAGE-004)". asserts:
      `assert_not_contains_iregex "P6" "docs/architecture.md" "Distribution \\(STAGE-004\\)"`.
    - `"P7"` — docs/tutorial.md §9 body (lines after `## 9. What's
      NOT there yet`) does NOT contain `brew install`. asserts: a
      sectioned `awk`/`sed` slice + `not_contains` check (helper
      `assert_section_not_contains` defined alongside; counts as one
      named assertion).
    - `"P8"` — cmd/brag/main.go does NOT contain the literal token
      `STAGE-004`. asserts:
      `assert_not_contains_iregex "P8" "cmd/brag/main.go" "STAGE-004"`.
    - `"P9"` — BRAG.md still references the SPEC-022 schema/hook/
      slash-command artifacts (regression check on the K-block
      claims). asserts: three `assert_contains_literal` checks for
      `docs/brag-entry.schema.json`,
      `scripts/claude-code-post-session.sh`,
      `examples/brag-slash-command.md` (counts as one named
      assertion).
    - `"P10"` — README.md does NOT contain "(recommended once
      available)" hedge. asserts:
      `assert_not_contains_iregex "P10" "README.md" "recommended once available"`.

  - **F4 self-pass meta (existing; stays as the last assertion).**
    The current scripts/test-docs.sh structure prints `OK: F4` after
    all named-group asserts pass; SPEC-023's extension preserves
    this — F4 still fires last.

**Total new asserts:** 10 (L) + 8 (M) + 5 (N) + 5 (O) + 10 (P) = 38
named assertion runs. (Some named assertions iterate over multiple
literals internally — e.g. O4 iterates 10 commands as one named
assertion; counted once. The test-docs.sh script will emit ~50 OK:
lines for these 38 named groups, similar to SPEC-022's expansion
ratio.)

**Total test-docs.sh size after build:** 63 (current) + ~33 named
assertion runs ≈ **~96 OK lines** (or ~100+ counting the multi-literal
expansions). Reconciles to the framing's "60–80 new asserts" upper
bound (38 named × ~1.3 expansion ≈ 50 OK lines for SPEC-023 alone).

## Implementation Context

*Read this section (and the files it points to) before starting
the build cycle. It is the equivalent of a handoff document, folded
into the spec since there is no separate receiving agent.*

### Decisions that apply

- **`DEC-001` — pure-Go SQLite driver (`modernc.org/sqlite`).** The
  whole reason cross-compile works without a CGO toolchain. The
  `.goreleaser.yaml` `builds[].env: [CGO_ENABLED=0]` line is the
  visible expression of this contract at release-build time. If this
  decision is ever reconsidered (e.g. switching to a CGO-based driver
  for performance), the goreleaser config must change in lockstep —
  worth noting so future maintainers see the link.
- No other DEC binds this spec directly. DEC-002 (embedded
  migrations), DEC-003 (config resolution), DEC-004 (tags
  comma-joined), DEC-005 (integer auto-increment IDs), DEC-006
  through DEC-014 — all apply forward unchanged. None gate the
  distribution mechanics.

### Constraints that apply

These constraints govern the paths SPEC-023 touches (see
`/guidance/constraints.yaml` for full text):

- **`no-cgo`** (blocking, **/*.go + go.mod + go.sum) — `.goreleaser.yaml`
  declares `CGO_ENABLED=0` in the builds env. If a future PR adds a
  CGO-only dependency, the goreleaser cross-compile breaks; the CI
  workflow's `go test ./...` on `ubuntu-latest` would catch most
  drift, but the explicit env declaration in goreleaser config
  surfaces the contract at the build stage where it matters.
- **`no-secrets-in-code`** (blocking, **) — `HOMEBREW_TAP_GITHUB_TOKEN`
  is a repository secret referenced as
  `${{ secrets.HOMEBREW_TAP_GITHUB_TOKEN }}` in
  `.github/workflows/release.yml`. The literal token never appears in
  any committed file. The constraint is satisfied trivially.
- **`no-new-top-level-deps-without-decision`** (warning, go.mod +
  go.sum) — SPEC-023 adds zero entries to go.mod or go.sum.
  Goreleaser and GitHub Actions are tooling, not Go module deps. The
  constraint is satisfied trivially.
- **`one-spec-per-pr`** (blocking) — SPEC-023's PR references exactly
  this spec. The doc sweep is folded in per the premise-audit rule
  (one spec ships its own doc-shaped consequences); not a separate
  bundled change.
- **`test-before-implementation`** (blocking) — the `## Failing
  Tests` section above defines the 33 new shell asserts BEFORE
  build transcribes the literal artifacts. Same TDD discipline as
  SPEC-021 + SPEC-022; the artefact is shell-script-extension-plus-
  YAML/markdown, not Go, but the order holds.
- **`stdout-is-for-data-stderr-is-for-humans`** (blocking,
  internal/cli/** + cmd/**) — applies to the `cmd/brag/main.go`
  six-line diff trivially (no I/O changes; just the version-
  injection wiring).
- **`errors-wrap-with-context`** (warning, internal/** + cmd/**) —
  applies to the `cmd/brag/main.go` diff trivially (no error-handling
  changes).

### Prior related work

- **`SPEC-021` (shipped 2026-04-25)** — DIRECT precedent. Created
  `scripts/test-docs.sh` (40 asserts in groups A–G plus F4 self-pass
  meta) under the §12 literal-artifact-as-spec pattern. Codified the
  pattern at ship. SPEC-023 is the third confirming application
  (cross-format YAML + workflow-YAML + Keep-A-Changelog markdown +
  six-line Go diff + shell-script extension); all literal artifacts
  embedded verbatim under § Notes for the Implementer.
- **`SPEC-022` (shipped 2026-04-26)** — second precedent. First
  cross-format application of the §12 pattern (JSON + bash +
  markdown + in-place markdown insertion + shell-script extension).
  Validated cross-format transcription discipline. SPEC-023 inherits
  the harness primitives unchanged + adds one new primitive
  (`assert_yaml_has_top_key`).
- **PR #21** (squash-merged `9abdeb6`, 2026-04-25) — SPEC-021 build
  PR.
- **PR #22** (squash-merged `079bb89`, 2026-04-26) — SPEC-022 build
  PR.

### Out of scope (for this spec specifically)

Explicit list of what this spec does NOT include. If any of these
feel necessary during build, create a new spec rather than expanding
this one.

- **Shell completions.** SPEC-024 ships `brag completion zsh|bash|fish`.
- **Branch protection on `main`.** Manual GitHub UI step; ship
  checklist captures.
- **Linux brew tap testing.** Goreleaser produces Linux archives but
  the tap formula targets macOS via `brew install`. Linux users go
  via direct archive download; not a SPEC-023 verification target.
- **Windows / Chocolatey / apt / yum / aur distribution.** Per brief:
  macOS-first; Linux falls out of goreleaser if free.
- **Pre-1.0 backward-compat promise.** Out per brief + per
  `docs/api-contract.md:399–405` stability section. No change.
- **Version bumping flow** (post-v0.1.0). The CHANGELOG's `[Unreleased]`
  section is the placeholder; subsequent releases follow Keep-A-
  Changelog convention. Not in this spec's literal artifacts beyond
  the v0.1.0 retroactive section.
- **Tags-normalization, soft-delete, edit-history migrations** — out
  per stage scope (debt accepted; revisit on concrete pain).
- **Goreleaser's snapshot / scoop / nix integrations.** Not needed
  for the homebrew-tap-only success criterion.
- **GitHub Releases auto-generated release notes** beyond goreleaser's
  built-in conventional-commit grouping. The CHANGELOG.md retroactive
  v0.1.0 section is the human-curated source of truth; goreleaser's
  release notes are derived (and overridable via the `release.body`
  template if the future needs it).
- **Pre-merge `goreleaser check` in CI.** Tempting (catches malformed
  config before tag time) but adds a tooling install step to ci.yml;
  the human ran `goreleaser check` locally at design-time-of-build
  per the smoke checklist (AC-37). Revisit if a malformed config
  ever ships.
- **A repo-level `.tool-versions` or `.go-version` file.** `go.mod`'s
  `go 1.26.2` line is sufficient and is the source of truth
  `setup-go` reads. Adding another version-pinning file is
  duplication.
- **Documentation of the manual GitHub UI steps in a long-lived doc
  file.** They live in this spec's § Notes for the Implementer
  (ship-checklist sub-list); after ship they are recorded in the
  spec's archived form. If the same UI steps recur for PROJ-002 or
  later, promote to `docs/development.md`.

### Locked design decisions

Numbered. Each maps to at least one Failing Test (per the SPEC-009
rule: "every locked decision needs a paired failing test").

1. **Bundle, not split** (Failing Test: AC-1 through AC-39 all in
   one spec; reasoning above under § Bundle vs split decision).
2. **Goreleaser v2 schema** — `version: 2` at top of
   `.goreleaser.yaml`. v1 is end-of-lifed. **Why:** v2 is the
   current stable major; v1 is end-of-lifed. (Failing Test: L2.)
3. **Cross-compile matrix:** `{darwin, linux} × {amd64, arm64}` —
   four binaries per release. Stage-locked. **Why:** matches the
   brief's macOS-first + Linux-as-nice-to-have framing without
   over-investing in Windows / aarch64-Windows / 386-anything.
   (Failing Tests: L4, L5.)
4. **CGO_ENABLED=0 explicit in builds env** — even though the binary
   is pure Go and would build CGO-free anyway, declaring the
   constraint at goreleaser-build time surfaces the no-CGO contract
   visibly. **Why:** if a CGO-pulling dep ever sneaks in, the build
   fails loudly at release rather than silently dropping back to
   per-arch native builds. (Failing Test: L3.)
5. **Archive format tar.gz with embedded `LICENSE`, `README.md`,
   `CHANGELOG.md`** — Homebrew + manual-tarball consumers expect tar.gz.
   Embedding LICENSE satisfies the formula's `license` field implicitly
   and makes the archive self-contained. README.md gives source-fetch
   users a quick on-ramp; CHANGELOG.md gives them release-notes
   without going to GitHub. **Why:** a tarball that can be inspected
   without re-cloning the repo. (Failing Tests: L9 + L10.)
6. **Checksum: sha256 to `checksums.txt`** — goreleaser default;
   homebrew formula uses the sha256 sums for download verification.
   **Why:** strongest widely-supported hash; matches every
   downstream consumer's expectation. (Failing Test: AC-10.)
7. **ldflags inject `version`, `commit`, `date`** — three
   conventional goreleaser injection targets. Today only `version`
   is consumed (`brag --version` output); `commit` + `date` are
   declared as `var` placeholders for a future `--version --verbose`
   without re-litigating the wiring. **Why:** zero-cost forward
   compat; reflexively defaulted by goreleaser convention. (Failing
   Tests: L9 + AC-27 + AC-28.)
8. **`var version = "dev"` (lowercase) in `cmd/brag/main.go`** —
   not `const Version`. **Why:** ldflags `-X` overrides require
   `var`, not `const`; lowercase `version` matches the conventional
   `-X main.version=...` injection target without per-binary path
   gymnastics. (Failing Test: AC-27.)
9. **Brews block: `repository.owner: jysf`, `repository.name:
   homebrew-bragfile`, `skip_upload: auto`, `license: MIT`** —
   stage-locked tap repo + industry-standard skip-on-prerelease +
   matches `LICENSE` file. **Why:** prereleases (`v0.1.0-rc1`)
   should NOT push a formula to the tap (rolling out to users
   prematurely). `skip_upload: auto` is goreleaser's
   prerelease-detection switch. (Failing Tests: L7 + L8 + L11.)
10. **Brews `test:` snippet runs `brag --version`** — minimum viable
    formula self-test; every Homebrew formula's `test do` block
    exercises one trivial happy path. **Why:** if the binary is
    fundamentally broken (missing libc, wrong arch), `--version`
    fails fast, and `brew test bragfile` surfaces it without
    requiring a working SQLite path. (Failing Test: AC-11.)
11. **Goreleaser changelog: `use: github`, conventional-commit
    grouped** — release notes derive from commits between tags,
    sorted ascending, grouped by `feat|fix|build|ci|chore|docs|other`.
    **Why:** the project's commit messages already follow
    conventional-commit shape (`feat(SPEC-XXX): …`,
    `chore(STAGE-XXX): …`); goreleaser's group-by-regex feature
    consumes that shape directly. Hand-curated CHANGELOG.md is the
    human-readable source of truth; goreleaser release notes are
    the auto-derived per-tag delta. (Failing Test: AC-12.)
12. **CI workflow triggers: `pull_request` (any branch) +
    `push: branches: main`** — the conventional GHA shape.
    `pull_request` covers all PR pushes regardless of source branch;
    `push: main` covers post-merge state-of-the-trunk verification.
    **Why:** PR-time check is what blocks bad merges;
    main-push-after-merge is the pessimistic fallback if a force-push
    or admin-bypass somehow lands an unverified commit. (Failing
    Tests: M2 + M3.)
13. **CI matrix: `{macos-latest, ubuntu-latest}`** — stage-locked.
    `windows-latest` is out per brief. **Why:** macOS-latest
    represents the user's primary dev machine + the brew-install
    target; ubuntu-latest catches Linux-specific drift the user
    won't see locally. (Failing Tests: M4 + M5.)
14. **CI Go version: `go-version-file: go.mod`** — not a hardcoded
    string. **Why:** when go.mod's go-version line bumps (e.g.
    1.26.2 → 1.26.3 → 1.27.0), CI follows automatically without
    a workflow YAML edit. Aligns with the user's
    `latest-stable-runtimes` memory rule. (Failing Test: AC-17.)
15. **CI permissions: `contents: read`** — least privilege.
    **Why:** the CI workflow only needs to read source; it does not
    write to the repo, comment on PRs, push tags, or create
    releases. Default `permissions: write-all` is wider than
    needed. (Failing Test: AC-18.)
16. **Release workflow trigger: tag push pattern `v*`** — broad
    pattern that matches `v0.1.0`, `v0.1.0-rc1`, `v1.0.0-beta`,
    `v0.1.1`, etc. Goreleaser's `prerelease: auto` + `skip_upload:
    auto` distinguish stable from prerelease at the goreleaser
    layer. **Why:** simpler than two separate triggers (one for
    `v[0-9]+.[0-9]+.[0-9]+`, one for `v*-*`) — goreleaser owns the
    prerelease semantics; the workflow trigger is the broad
    upstream filter. (Failing Test: N2.)
17. **Release workflow checkout `fetch-depth: 0`** — full git history
    needed for goreleaser's changelog generator. **Why:** `fetch-depth:
    1` (default) only checks out the tagged commit; goreleaser's
    `use: github` changelog needs the commit DAG between the previous
    tag and the current tag to compute the delta. (Failing Test:
    N5.)
18. **Release workflow permissions: `contents: write`** — needed
    to create the GitHub release. **Why:** least-privilege override
    of the default — the release workflow needs write to publish a
    release, but it doesn't need pull-requests:write or issues:write
    or actions:write. (Failing Test: AC-23.)
19. **Secret: `HOMEBREW_TAP_GITHUB_TOKEN`** — stage-locked name.
    **Why:** `GITHUB_TOKEN` (the workflow's auto-generated default
    token) is scoped to the running repo; it cannot push to a
    different repo (`homebrew-bragfile`). A PAT with `repo` scope on
    the tap is the minimal escalation needed. The name
    `HOMEBREW_TAP_GITHUB_TOKEN` matches goreleaser's documented
    convention so that the secret value is consumed without per-
    secret remapping in the workflow yaml. (Failing Test: N4.)
20. **`actions/checkout@v4`, `actions/setup-go@v5`,
    `goreleaser/goreleaser-action@v6`** — major-version pins.
    **Why:** matches industry convention. SHA-pinning is stricter
    (revisit if supply-chain becomes load-bearing in PROJ-002) but
    over-investing for a personal-tool project. (Failing Tests: AC-17
    + AC-20 + N3.)
21. **CHANGELOG style: Keep-A-Changelog 1.1.0** — explicit reference
    to `keepachangelog.com/en/1.1.0/` in the file's preamble.
    **Why:** widely-recognised; the format is self-documenting at
    the top of the file; consumer tooling (e.g. semantic-release-
    style automation, future) can parse the structure. The
    alternative — a simpler conventional-commit digest — was
    considered and rejected: the digest gives goreleaser something
    to autogenerate but does not give a human consumer something to
    *read* without browsing GitHub. CHANGELOG.md is the
    human-readable source of truth; goreleaser's autogenerated
    release notes are the per-tag delta. (Failing Test: O2.)
22. **CHANGELOG.md v0.1.0 section populated at design time
    (literal-artifact-as-spec)** — the v0.1.0 section is committed
    in this PR with `YYYY-MM-DD` placeholder for the actual ship
    date; build replaces the placeholder with the real date when
    cutting v0.1.0. **Why:** §12 pattern says fixed-shape artifacts
    embed verbatim at design; the v0.1.0 section is one such
    artifact (curated from PROJ-001's commit history). The
    `YYYY-MM-DD → 2026-MM-DD` substitution is the only build-time
    change. (Failing Tests: O3 + O4.)
23. **CHANGELOG.md `### Decisions of record` subsection lists DEC-001
    through DEC-014** — explicit roll-up of every architectural
    commitment that landed in PROJ-001. **Why:** the user-facing
    consumer of the CHANGELOG (a downstream maintainer pinning
    `bragfile@v0.1.0`) wants a single page that names the
    architectural choices binding the version. Pointing them at
    `/decisions/` requires a clone; the CHANGELOG inlines the
    list. (Failing Test: AC-25.)
24. **Comment update in `cmd/brag/main.go:11–13`** — STAGE-004 →
    STAGE-005 in the version-injection comment. **Why:** stale
    references rot; the inherited audit caught this; folding the
    comment-fix into the const→var diff means one PR, one diff,
    not two. (Failing Tests: AC-29 + P8.)
25. **`var commit = "none"` and `var date = "unknown"` declared but
    unused at runtime today** — placeholders for goreleaser's
    conventional ldflag targets. **Why:** zero-cost forward compat
    for `--version --verbose`; if SPEC-023 declared only `version`
    and a future spec wants commit + date, that future spec would
    have to re-litigate the wiring. Better to ship all three in
    SPEC-023 and accept two unused vars today. (Failing Test:
    AC-28.)
26. **Doc-sweep applies the audit-grep cross-check (§9, codified at
    SPEC-018 ship)** — the design-time greps RAN against the live
    repo, hits enumerated under § Inherited SPEC-021 doc-sweep audit
    and § SPEC-023 forward-reference audit. Build re-runs both greps
    before transcribing the doc-sweep edits and treats deltas as
    questions. (Failing Tests: P3, P4, P5, P6, P7, P8, P10 — group
    P collectively.)
27. **NO new `brag` CLI surface** — chore, not story. The
    `cmd/brag/main.go` change is wiring, not feature. Per stage-
    framing Q3 answer 2026-04-25. **Why:** scope discipline; the
    brief's success criterion 4 says "`brew install bragfile`
    installs a working binary", not "ships a new install-management
    subcommand". (Implicit acceptance criterion: AC-list contains
    no "`brag <new-verb>`" commands.)

### Rejected alternatives (build-time)

Per AGENTS.md §12 design-time decision rule (codified at SPEC-018
ship): when a "multiple paths" choice is decidable at design time,
lock the prescribed path AND list the rejected alternatives
explicitly. This section enumerates the alternatives a build session
might re-litigate. Build holds the line.

1. **REJECTED: Two-spec split (SPEC-023a goreleaser+CI vs SPEC-023b
   release+tap+CHANGELOG).** Reasoning under § Bundle vs split
   decision. Build does not split.

2. **REJECTED: SHA-pinned action versions** (e.g.
   `actions/checkout@a5ac7e51b41094c92402da3b24376905380afc29`).
   Major-version pins (`@v4` / `@v5` / `@v6`) are simpler and match
   industry convention for personal-tool projects. SHA-pinning is
   stricter on supply-chain risk; revisit at PROJ-002 framing if
   that surface becomes load-bearing. Build holds major-version
   pins.

3. **REJECTED: Custom GitHub release body via
   `.goreleaser.yaml.release.body`** — overrides goreleaser's
   default group-by-conventional-commit changelog with a hand-curated
   string per release. Out of scope for v0.1.0; the auto-generated
   release notes plus the curated `CHANGELOG.md` are the source-of-
   truth pair. Build does not add a `release.body` template.

4. **REJECTED: Adding `goreleaser check` step to ci.yml** — would
   catch malformed config before tag-time, but adds a tooling install
   step to ci.yml. The human runs `goreleaser check` + `goreleaser
   build --snapshot --clean` locally per the smoke checklist (AC-37).
   Revisit if a malformed config ever ships through to a tag.
   Build does not add the step.

5. **REJECTED: Committing a `bragfile.rb` formula scaffold to the
   tap repo at SPEC-023 PR time.** GoReleaser generates the formula
   on the first stable tag push; pre-committing a scaffold creates
   drift the moment goreleaser overwrites it. Build does not touch
   the tap repo from the SPEC-023 PR. (The tap repo itself was
   created as a STAGE-005 framing chore 2026-04-25 with a
   one-paragraph README; that's its only content until v0.1.0
   ships.)

6. **REJECTED: A separate `.go-version` or `.tool-versions` file**
   alongside `go.mod`'s `go 1.26.2` line. Duplication. Build relies
   on `setup-go: go-version-file: go.mod` to read the existing
   source of truth.

7. **REJECTED: Default `permissions: write-all` on workflows.** Both
   workflows declare narrow `permissions:` blocks (`contents: read`
   for ci.yml; `contents: write` for release.yml). Build does not
   widen.

8. **REJECTED: A `Makefile` shim alongside the existing `justfile`**
   for users who prefer `make`. The brief locks `just` as the
   command runner; goreleaser + GHA do not require `make` in the
   build path. Build does not add a Makefile.

9. **REJECTED: `gofmt -d .` (diff form) instead of `gofmt -l .`
   (list form) in ci.yml.** `-l .` is the AGENTS.md §3 + §9
   convention; `-d` is more verbose but redundant given that the CI
   logs already include the failure context. Build holds `-l .`.

10. **REJECTED: Promoting the manual GitHub UI steps to a
    `docs/development.md` "Release-engineering" section.** The
    steps are one-time per repo and live in this spec's archived
    form post-ship. Promoting them now is premature; revisit at
    PROJ-002 framing if a similar release-engineering shape recurs.

11. **REJECTED: Pre-creating a `v0.1.0-rc1` tag in the SPEC-023 PR
    body for branch-based smoke testing.** Tags are immutable; the
    smoke is performed AFTER PR merge to `main`, not on the feat
    branch. Build holds the order: merge SPEC-023 → tag rc1 → verify
    → tag v0.1.0.

12. **REJECTED: Including a goreleaser `nfpms:` (deb / rpm) block.**
    Out per brief: macOS-first; Linux falls out of goreleaser if
    free; deb/rpm packaging is not "free" — it's an additional
    surface that requires per-distro signing and repository hosting
    not in scope. Build does not add `nfpms`.

13. **REJECTED: A `## [Unreleased]` section in CHANGELOG.md
    populated with placeholder commits.** The `## [Unreleased]`
    section is empty at v0.1.0 ship time (it's a forward-looking
    placeholder for post-v0.1.0 commits). Build does not pre-
    populate.

### Build-side audit-grep cross-check (mandatory, per §9)

Before transcribing the doc-sweep edits, build re-runs the design-
time greps and reconciles against the tables under § Inherited
SPEC-021 doc-sweep audit + § SPEC-023 forward-reference audit. If
the live repo has drifted (e.g. someone landed a chore commit between
design and build that touched one of the audit-affected files),
build STOPS and asks the spec author rather than expanding scope
unilaterally.

The two greps to re-run:

```bash
grep -n "STAGE-00[0-9]" docs/api-contract.md docs/architecture.md docs/data-model.md docs/tutorial.md
grep -n -E "brew install|homebrew|goreleaser|CHANGELOG" docs/api-contract.md docs/architecture.md docs/data-model.md docs/tutorial.md README.md AGENTS.md BRAG.md CONTRIBUTING.md docs/development.md cmd/brag/main.go
```

Expected outputs (modulo line-number drift if a chore commit landed
between design and build): the two tables under § Inherited and §
SPEC-023 above. Any new hit not in those tables is a question for
the spec author.

### Confidence-target check (§14, codified at AGENTS.md line ~382)

Honest confidence values on this spec's locked design decisions.
Per §14: any choice <0.8 becomes a question in
`/guidance/questions.yaml`. **Result: zero choices below 0.8;
zero new questions.yaml entries needed.**

| Decision | Confidence | Rationale |
|---|---|---|
| Bundle, not split | 0.95 | Reasoning explicit; size lands at solid M; smoke-test plan demands bundling |
| Goreleaser v2 | 1.00 | v1 is end-of-lifed |
| Cross-compile matrix | 1.00 | Stage-locked |
| CGO_ENABLED=0 explicit | 0.95 | Belt-and-suspenders for the no-cgo constraint |
| tar.gz + LICENSE/README/CHANGELOG embed | 0.95 | Industry standard |
| sha256 checksum | 1.00 | Goreleaser default |
| ldflags inject version+commit+date | 0.90 | All three placeholders is forward-compat; runtime cost is two unused vars |
| `var version` (lowercase) | 1.00 | ldflags require var, not const; lowercase matches goreleaser convention |
| Brews block: skip_upload: auto + license: MIT + test: brag --version | 0.95 | Stage-locked; matches LICENSE file; minimum viable formula test |
| Goreleaser changelog use: github + grouped | 0.85 | Project's commit messages follow conventional-commit shape; auto-derivation is safe |
| CI triggers: pull_request + push:main | 1.00 | Conventional GHA shape |
| CI matrix: {macos, ubuntu}-latest | 1.00 | Stage-locked |
| go-version-file: go.mod | 0.95 | Aligns with latest-stable-runtimes user memory rule |
| CI permissions: contents: read | 1.00 | Least-privilege convention |
| Release trigger: v* (broad) | 0.85 | Goreleaser handles prerelease semantics; broader trigger simpler than two-trigger setup |
| Release fetch-depth: 0 | 1.00 | Goreleaser changelog needs full DAG |
| Release permissions: contents: write | 1.00 | Minimum needed to publish releases |
| HOMEBREW_TAP_GITHUB_TOKEN secret name | 0.95 | Goreleaser-documented convention |
| Action major-version pins | 0.85 | Matches industry convention; SHA-pinning is over-investment for personal project |
| CHANGELOG: Keep-A-Changelog 1.1.0 | 0.85 | Most-recognised format; tooling-friendly |
| CHANGELOG v0.1.0 populated at design | 0.90 | §12 literal-artifact-as-spec applies |
| Decisions-of-record subsection | 0.85 | User-facing roll-up convenience |
| const→var rewiring + STAGE-004→STAGE-005 comment | 1.00 | Mechanical; required for ldflags |
| `commit` + `date` placeholders unused at runtime | 0.85 | Future-compat; small cost today |
| Audit-grep cross-check at build | 1.00 | Codified §9 rule |

All confidences ≥ 0.85. No `/guidance/questions.yaml` entries.

## Notes for the Implementer

Gotchas, style preferences, reuse opportunities. **All four primary
artifacts are LITERAL — transcribe verbatim per §12 literal-artifact-
as-spec. Build verifies via `git diff` against this section's blocks
before commit.**

### `.goreleaser.yaml` (LITERAL — transcribe verbatim)

```yaml
# .goreleaser.yaml — see https://goreleaser.com/customization/
# Goreleaser v2 schema. Cross-compiles darwin+linux × amd64+arm64.
# CGO_ENABLED=0 is the visible expression of the no-cgo constraint
# (DEC-001 — pure-Go modernc.org/sqlite driver).

version: 2

before:
  hooks:
    - go mod tidy

builds:
  - id: brag
    main: ./cmd/brag
    binary: brag
    env:
      - CGO_ENABLED=0
    goos:
      - darwin
      - linux
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w
      - -X main.version={{ .Version }}
      - -X main.commit={{ .Commit }}
      - -X main.date={{ .Date }}

archives:
  - id: default
    name_template: >-
      {{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}
    format: tar.gz
    files:
      - README.md
      - LICENSE
      - CHANGELOG.md

checksum:
  name_template: "checksums.txt"
  algorithm: sha256

changelog:
  use: github
  sort: asc
  groups:
    - title: Features
      regexp: '^.*?feat(\(.+?\))??!?:.+$'
      order: 0
    - title: Bug fixes
      regexp: '^.*?fix(\(.+?\))??!?:.+$'
      order: 1
    - title: Build / CI
      regexp: '^.*?(build|ci|chore)(\(.+?\))??!?:.+$'
      order: 2
    - title: Documentation
      regexp: '^.*?docs(\(.+?\))??!?:.+$'
      order: 3
    - title: Other
      order: 999

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

release:
  github:
    owner: jysf
    name: bragfile000
  prerelease: auto
  draft: false
```

### `.github/workflows/ci.yml` (LITERAL — transcribe verbatim)

```yaml
# .github/workflows/ci.yml — PR + main-push gating
# Runs gofmt, go vet, go test on macOS-latest + ubuntu-latest with
# the Go version declared in go.mod.

name: ci

on:
  pull_request:
  push:
    branches:
      - main

permissions:
  contents: read

jobs:
  test:
    strategy:
      fail-fast: false
      matrix:
        os: [macos-latest, ubuntu-latest]
    runs-on: ${{ matrix.os }}
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
          check-latest: true
          cache: true

      - name: Verify gofmt
        run: |
          unformatted=$(gofmt -l .)
          if [ -n "$unformatted" ]; then
            printf 'gofmt -l reports unformatted files:\n%s\n' "$unformatted" >&2
            exit 1
          fi

      - name: Run go vet ./...
        run: go vet ./...

      - name: Run go test ./...
        run: go test ./...
```

### `.github/workflows/release.yml` (LITERAL — transcribe verbatim)

```yaml
# .github/workflows/release.yml — tag-push triggered release
# Cross-compiles via goreleaser, publishes the GitHub release, pushes
# the bragfile.rb formula to github.com/jysf/homebrew-bragfile on
# stable tags (skipped automatically on prerelease tags via
# brews.skip_upload: auto in .goreleaser.yaml).

name: release

on:
  push:
    tags:
      - 'v*'

permissions:
  contents: write

jobs:
  goreleaser:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
          check-latest: true
          cache: true

      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v6
        with:
          distribution: goreleaser
          version: '~> v2'
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          HOMEBREW_TAP_GITHUB_TOKEN: ${{ secrets.HOMEBREW_TAP_GITHUB_TOKEN }}
```

### `CHANGELOG.md` (LITERAL — transcribe verbatim; build replaces `YYYY-MM-DD` with the actual v0.1.0 ship date at tag-cut time)

```markdown
# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - YYYY-MM-DD

Initial public release of `brag`, a local-first Go CLI for capturing
and retrieving career-worthy moments. Entries live in an embedded
SQLite database at `~/.bragfile/db.sqlite`. No cloud, no sync, no
account.

### Added

- `brag add` — capture an entry via flags (`-t/--title`, `-d`, `-T`,
  `-p`, `-k`, `-i`) or via `$EDITOR` against a templated markdown
  buffer.
- `brag add --json` — programmatic capture from stdin, validated
  against the DEC-012 single-object schema (title required;
  optional user-owned fields; server-owned fields tolerated and
  ignored; unknown keys strict-rejected).
- `brag list` — list entries newest-first, with `--project`,
  `--tag`, `--type`, `--since` filters and `--show-project / -P`
  for an extra project column. `--format json|tsv` for
  machine-readable output.
- `brag show <id>` — display a single entry with full metadata.
- `brag edit <id>` — round-trip an entry through `$EDITOR`.
- `brag delete <id>` — delete with `[y/N]` confirmation.
- `brag search <query>` — SQLite FTS5 full-text search across
  title, description, tags, project, and impact.
- `brag export --format markdown|json` — bulk export with the same
  filter flags as `list`. `--out file` to write to disk.
- `brag summary --range week|month` — rule-based aggregation
  grouped by project and type, rendered as markdown or JSON
  (DEC-014 envelope).
- `brag review --week|--month` — entries grouped by project plus
  three reflection questions, designed to be piped into an
  external AI session.
- `brag stats` — six lifetime metrics: total entries, weekly
  rolling average, current streak, longest streak, top-5 tags,
  top-5 projects, corpus span.
- `docs/brag-entry.schema.json` — JSON Schema (draft 2020-12)
  mirroring the `brag add --json` stdin contract for AI-agent
  validation.
- `scripts/claude-code-post-session.sh` + `examples/brag-slash-command.md`
  — reference Claude Code session-end hook and slash-command
  template demonstrating the round-trip.
- macOS (arm64, amd64) and Linux (arm64, amd64) binaries via
  goreleaser.
- Homebrew tap at `github.com/jysf/homebrew-bragfile` —
  `brew install jysf/bragfile/bragfile`.

### Decisions of record

The following architectural decisions are committed in this release.
Each decision file under `/decisions/` carries the full rationale.

- DEC-001 — pure-Go SQLite driver (`modernc.org/sqlite`); no CGO.
- DEC-002 — embedded migrations via `embed.FS`, applied on
  `storage.Open`.
- DEC-003 — config resolution order: `--db` flag → `BRAGFILE_DB`
  env → `~/.bragfile/db.sqlite` default.
- DEC-004 — tags stored as a comma-joined string for MVP.
- DEC-005 — integer `AUTOINCREMENT` primary keys.
- DEC-006 — `spf13/cobra` as the CLI framework.
- DEC-007 — required-flag validation in `RunE` (cobra's
  `MarkFlagRequired` reports errors via stderr + non-zero exit;
  the project owns user-error rendering uniformly).
- DEC-008 — `--since` accepts date (`2026-04-19`) or duration
  (`7d`, `30d`).
- DEC-009 — editor buffer format for `brag add` / `brag edit`
  (markdown front-matter on top, free body below).
- DEC-010 — `brag search` query syntax (auto-tokenize whitespace;
  treat hyphens / dots as literal; phrase-quote multi-word
  fragments).
- DEC-011 — JSON output shape for `brag list --format json` and
  `brag export --format json`: naked array of nine-key entry
  objects; field names match SQL columns verbatim.
- DEC-012 — `brag add --json` stdin schema: single object, title
  required, server-owned fields tolerated-and-ignored, unknown
  keys strict-rejected.
- DEC-013 — markdown export shape for `brag export --format
  markdown` (+ `--flat`): per-entry markdown blocks under
  per-project headings; `--flat` flattens.
- DEC-014 — rule-based output envelope for `brag summary` /
  `brag review` / `brag stats`: single-object JSON envelope with
  `generated_at` / `scope` / `filters` provenance + per-spec
  payload keys; markdown convention reuses DEC-013's provenance
  + summary-block style.

[Unreleased]: https://github.com/jysf/bragfile000/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/jysf/bragfile000/releases/tag/v0.1.0
```

### `cmd/brag/main.go` (LITERAL diff)

Before (current state):

```go
package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/jysf/bragfile000/internal/cli"
)

// Version is set to "dev" for local builds. goreleaser injects the real
// version via ldflags in STAGE-004.
const Version = "dev"

func main() {
	root := cli.NewRootCmd(Version)
	root.AddCommand(cli.NewAddCmd())
	root.AddCommand(cli.NewListCmd())
	// ... rest unchanged
}
```

After (post-build):

```go
package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/jysf/bragfile000/internal/cli"
)

// version is set to "dev" for local builds. goreleaser injects the
// real values via ldflags (-X main.version=... -X main.commit=...
// -X main.date=...) at release-build time. See .goreleaser.yaml.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	root := cli.NewRootCmd(version)
	root.AddCommand(cli.NewAddCmd())
	root.AddCommand(cli.NewListCmd())
	// ... rest unchanged
}
```

Note the `_ = commit; _ = date` blank-identifier acknowledgement is
NOT added — Go does not flag package-level `var` declarations as
unused (only `var` declarations inside a function body). The two
unused package-level vars compile clean and are reachable for a
future `--version --verbose` enhancement without re-litigating the
wiring.

### `README.md` (LITERAL diffs — three edits)

**Edit 1: status banner (lines 10–12)**

Before:

```markdown
> **Status:** in active development. Capture, retrieve, search,
> export, and weekly/monthly digests are shipped. Distribution via
> Homebrew is in progress.
```

After:

```markdown
> **Status:** v0.1.0 shipped. Capture, retrieve, search, export, and
> weekly/monthly digests are all available, and `brew install
> jysf/bragfile/bragfile` installs the binary on macOS.
```

**Edit 2: Install heading (line 16)**

Before:

```markdown
Homebrew (recommended once available):
```

After:

```markdown
Homebrew (recommended):
```

**Edit 3: From-source clarification (lines 32–33)**

Before:

```markdown
Requires Go 1.26+ from source. The Homebrew install pulls a
prebuilt binary — no Go required.
```

After:

```markdown
The Homebrew install pulls a prebuilt binary — no Go required.
Requires Go 1.26+ if you build from source instead.
```

(All other README.md content stays byte-identical.)

### `AGENTS.md` (LITERAL diffs — three edits)

**Edit 1: §3 Tech Stack, line 67**

Before:

```markdown
- **Distribution:** `goreleaser` → GitHub Releases → homebrew tap at `github.com/jysf/homebrew-bragfile` (arriving in STAGE-005).
```

After:

```markdown
- **Distribution:** `goreleaser` → GitHub Releases → homebrew tap at `github.com/jysf/homebrew-bragfile`. Install via `brew install jysf/bragfile/bragfile`.
```

**Edit 2: §4 Commands, line 97**

Before:

```markdown
# --- release (STAGE-004) ---
```

After:

```markdown
# --- release (STAGE-005) ---
```

**Edit 3: §4 Commands, line 106**

Before:

```markdown
macOS note: `brew upgrade go` to move to the latest Go. `brew install goreleaser` once we hit STAGE-004.
```

After:

```markdown
macOS note: `brew upgrade go` to move to the latest Go. `brew install goreleaser` before tagging a release locally.
```

(All other AGENTS.md content stays byte-identical. Line 258's `tap`
glossary entry is correct as-is and STAYS.)

### `docs/architecture.md` (LITERAL diffs — three edits)

**Edit 1: line 45 (Components table, `internal/export` row)**

Before:

```markdown
| `internal/export` | (STAGE-003) Markdown-report and sqlite-file-copy exporters. |
```

After:

```markdown
| `internal/export` | (STAGE-003) Markdown-report and JSON exporters (`brag export --format markdown\|json`). |
```

**Edit 2: lines 102–106 (Deployment Topology paragraph)**

Before:

```markdown
There is none. `brag` is a single static binary that runs on the user's
machine. Distribution (STAGE-004) uses goreleaser to produce macOS
(arm64, x86_64) and Linux (arm64, x86_64) binaries; a homebrew tap at
`github.com/jysf/homebrew-bragfile` ships the macOS ones via
`brew install bragfile`.
```

After:

```markdown
There is none. `brag` is a single static binary that runs on the user's
machine. Distribution (STAGE-005) uses goreleaser to produce macOS
(arm64, amd64) and Linux (arm64, amd64) binaries; a homebrew tap at
`github.com/jysf/homebrew-bragfile` ships the macOS ones via
`brew install jysf/bragfile/bragfile`.
```

(Note: `x86_64` → `amd64` for consistency with the goreleaser config's
`goarch` values. Both label the same architecture; goreleaser uses
`amd64` per its conventions. Visible in the produced archive
filenames.)

(All other architecture.md content stays byte-identical.)

### `docs/tutorial.md` (LITERAL diff — one section edit)

**Edit 1: §9 *"What's NOT there yet"* (lines 487–496)**

Before:

```markdown
---

## 9. What's NOT there yet

So you don't ask the tool for things it can't do:

| Want | Status |
|---|---|
| `brew install bragfile` | STAGE-005 |

For now, `sqlite3 ~/.bragfile/db.sqlite` is your escape hatch for
anything `list` doesn't surface.

---
```

After:

```markdown
---

## 9. Power-user escape hatch

Everything in this tutorial is shipped in v0.1.0. For corner cases
`brag list` doesn't surface, `sqlite3 ~/.bragfile/db.sqlite` is your
escape hatch.

---
```

(Section heading renamed from "What's NOT there yet" to "Power-user
escape hatch" to reflect the new content; the closing horizontal rule
+ the `## Further reading` section that follows are unchanged. All
other tutorial.md content stays byte-identical.)

### `scripts/test-docs.sh` extension (LITERAL — transcribe verbatim, append after group K, before the `# ===== finalise =====` block at the existing line 587)

```bash
# ===== Group L — .goreleaser.yaml shape =====

GORELEASER="${REPO_ROOT}/.goreleaser.yaml"

# L1 — config file exists
assert_file_exists "L1" "$GORELEASER"

# L2 — opens with `version: 2`
if [ ! -f "$GORELEASER" ]; then
    fail "L2" "$GORELEASER does not exist"
elif head -n 5 "$GORELEASER" | grep -E -q '^version:[[:space:]]+2[[:space:]]*$'; then
    ok "L2"
else
    fail "L2" "$GORELEASER does not declare 'version: 2' in first 5 lines"
fi

# L3 — declares CGO_ENABLED=0
assert_contains_literal "L3" "$GORELEASER" "CGO_ENABLED=0"

# L4 — declares both darwin and linux goos values
assert_contains_literal "L4a" "$GORELEASER" "- darwin"
assert_contains_literal "L4b" "$GORELEASER" "- linux"

# L5 — declares both amd64 and arm64 goarch values
assert_contains_literal "L5a" "$GORELEASER" "- amd64"
assert_contains_literal "L5b" "$GORELEASER" "- arm64"

# L6 — declares a top-level `brews:` block
if [ ! -f "$GORELEASER" ]; then
    fail "L6" "$GORELEASER does not exist"
elif grep -E -q '^brews:[[:space:]]*$' "$GORELEASER"; then
    ok "L6"
else
    fail "L6" "$GORELEASER does not declare a top-level 'brews:' block"
fi

# L7 — brews block points at homebrew-bragfile
assert_contains_literal "L7" "$GORELEASER" "name: homebrew-bragfile"

# L8 — brews block has `skip_upload: auto`
assert_contains_literal "L8" "$GORELEASER" "skip_upload: auto"

# L9 — declares `-X main.version=` ldflag
assert_contains_literal "L9" "$GORELEASER" "-X main.version="

# L10 — archive format is `tar.gz`
assert_contains_literal "L10" "$GORELEASER" "format: tar.gz"

# L11 — brews block declares license: MIT
assert_contains_literal "L11" "$GORELEASER" 'license: "MIT"'

# ===== Group M — .github/workflows/ci.yml shape =====

CI_WORKFLOW=".github/workflows/ci.yml"

# M1 — file exists
assert_file_exists "M1" "$CI_WORKFLOW"

# M2 — triggers on pull_request
assert_contains_literal "M2" "$CI_WORKFLOW" "pull_request:"

# M3 — triggers on push to main (push: stanza followed by branches: + main)
if [ ! -f "$CI_WORKFLOW" ]; then
    fail "M3" "$CI_WORKFLOW does not exist"
elif awk '
        /^on:/ { in_on=1; next }
        in_on && /^[a-zA-Z]/ { in_on=0 }
        in_on && /push:/ { in_push=1; next }
        in_push && /branches:/ { in_branches=1; next }
        in_branches && /- main/ { print "ok"; exit }
    ' "$CI_WORKFLOW" | grep -q ok; then
    ok "M3"
else
    fail "M3" "$CI_WORKFLOW does not trigger on push to main"
fi

# M4 — matrix includes macos-latest
assert_contains_literal "M4" "$CI_WORKFLOW" "macos-latest"

# M5 — matrix includes ubuntu-latest
assert_contains_literal "M5" "$CI_WORKFLOW" "ubuntu-latest"

# M6 — runs `go test ./...`
assert_contains_literal "M6" "$CI_WORKFLOW" "go test ./..."

# M7 — runs `gofmt -l .`
assert_contains_literal "M7" "$CI_WORKFLOW" "gofmt -l ."

# M8 — runs `go vet ./...`
assert_contains_literal "M8" "$CI_WORKFLOW" "go vet ./..."

# M9 — uses actions/setup-go@v5
assert_contains_literal "M9" "$CI_WORKFLOW" "actions/setup-go@v5"

# ===== Group N — .github/workflows/release.yml shape =====

RELEASE_WORKFLOW=".github/workflows/release.yml"

# N1 — file exists
assert_file_exists "N1" "$RELEASE_WORKFLOW"

# N2 — triggers on tag push pattern v*
if [ ! -f "$RELEASE_WORKFLOW" ]; then
    fail "N2" "$RELEASE_WORKFLOW does not exist"
elif grep -E -q "^[[:space:]]+- 'v\\*'" "$RELEASE_WORKFLOW"; then
    ok "N2"
else
    fail "N2" "$RELEASE_WORKFLOW does not trigger on tag pattern 'v*'"
fi

# N3 — uses goreleaser/goreleaser-action@v6
assert_contains_literal "N3" "$RELEASE_WORKFLOW" "goreleaser/goreleaser-action@v6"

# N4 — passes HOMEBREW_TAP_GITHUB_TOKEN env
assert_contains_literal "N4" "$RELEASE_WORKFLOW" "HOMEBREW_TAP_GITHUB_TOKEN"

# N5 — checkout uses fetch-depth: 0
assert_contains_literal "N5" "$RELEASE_WORKFLOW" "fetch-depth: 0"

# ===== Group O — CHANGELOG.md shape =====

# O1 — file exists
assert_file_exists "O1" "CHANGELOG.md"

# O2 — references Keep-A-Changelog
assert_contains_literal "O2" "CHANGELOG.md" "keepachangelog.com"

# O3 — has `## [0.1.0]` heading (line-based equality avoids substring trap)
if [ ! -f CHANGELOG.md ]; then
    fail "O3" "CHANGELOG.md does not exist"
elif grep -E -q '^## \[0\.1\.0\]' CHANGELOG.md; then
    ok "O3"
else
    fail "O3" "CHANGELOG.md missing '## [0.1.0]' heading"
fi

# O4 — lists each shipped command verb under Added (single named
# assertion that iterates internally over the ten verbs)
o4_failed=""
for verb in "brag add" "brag list" "brag show" "brag edit" "brag delete" \
            "brag search" "brag export" "brag summary" "brag review" "brag stats"; do
    if ! grep -F -q -- "\`${verb}\`" CHANGELOG.md; then
        o4_failed="${o4_failed} ${verb}"
    fi
done
if [ -z "$o4_failed" ]; then
    ok "O4"
else
    fail "O4" "CHANGELOG.md missing command refs:$o4_failed"
fi

# O5 — has [Unreleased] and [0.1.0] link reference definitions
o5_failed=""
for ref in "[Unreleased]:" "[0.1.0]:"; do
    if ! grep -F -q -- "$ref" CHANGELOG.md; then
        o5_failed="${o5_failed} ${ref}"
    fi
done
if [ -z "$o5_failed" ]; then
    ok "O5"
else
    fail "O5" "CHANGELOG.md missing link refs:$o5_failed"
fi

# ===== Group P — Doc sweep + tense flips =====

# P1 — README.md status banner does NOT contain "in progress"
assert_not_contains_iregex "P1" "README.md" "in progress"

# P2 — README.md does NOT contain "in active development"
assert_not_contains_iregex "P2" "README.md" "in active development"

# P3 — AGENTS.md does NOT contain "arriving in STAGE"
assert_not_contains_iregex "P3" "AGENTS.md" "arriving in STAGE"

# P4 — AGENTS.md does NOT contain literal `(STAGE-004)`
assert_not_contains_iregex "P4" "AGENTS.md" "\\(STAGE-004\\)"

# P5 — docs/architecture.md does NOT contain "sqlite-file-copy"
assert_not_contains_iregex "P5" "docs/architecture.md" "sqlite-file-copy"

# P6 — docs/architecture.md does NOT contain "Distribution (STAGE-004)"
assert_not_contains_iregex "P6" "docs/architecture.md" "Distribution \\(STAGE-004\\)"

# P7 — docs/tutorial.md §9 body does NOT contain `brew install`
# (sectioned slice from the §9 heading to the next ## or --- divider)
if [ ! -f docs/tutorial.md ]; then
    fail "P7" "docs/tutorial.md does not exist"
else
    section_body=$(awk '
        /^## 9\./ { in_section=1; next }
        in_section && /^## / { in_section=0 }
        in_section && /^---[[:space:]]*$/ { in_section=0 }
        in_section { print }
    ' docs/tutorial.md)
    if printf '%s' "$section_body" | grep -F -q "brew install"; then
        fail "P7" "docs/tutorial.md §9 body contains 'brew install'"
    else
        ok "P7"
    fi
fi

# P8 — cmd/brag/main.go does NOT contain literal STAGE-004
assert_not_contains_iregex "P8" "cmd/brag/main.go" "STAGE-004"

# P9 — BRAG.md still references SPEC-022 artifacts (regression check)
p9_failed=""
for art in "docs/brag-entry.schema.json" \
           "scripts/claude-code-post-session.sh" \
           "examples/brag-slash-command.md"; do
    if ! grep -F -q -- "$art" BRAG.md; then
        p9_failed="${p9_failed} ${art}"
    fi
done
if [ -z "$p9_failed" ]; then
    ok "P9"
else
    fail "P9" "BRAG.md missing SPEC-022 artifact refs:$p9_failed"
fi

# P10 — README.md does NOT contain "(recommended once available)"
assert_not_contains_iregex "P10" "README.md" "recommended once available"

# ===== finalise =====
```

(The existing `# ===== finalise =====` block from line 587 onward —
the failure-count-check + `ok "F4"` self-pass meta + the
`ALL OK: …` printf — stays UNCHANGED. SPEC-023's extension inserts
groups L/M/N/O/P between the existing `K4` block and the existing
`finalise` block.)

### Build-side audit-grep cross-check (mandatory, per §9)

Re-run the design-time greps before transcribing the doc-sweep edits:

```bash
grep -n "STAGE-00[0-9]" docs/api-contract.md docs/architecture.md docs/data-model.md docs/tutorial.md
grep -n -E "brew install|homebrew|goreleaser|CHANGELOG" docs/api-contract.md docs/architecture.md docs/data-model.md docs/tutorial.md README.md AGENTS.md BRAG.md CONTRIBUTING.md docs/development.md cmd/brag/main.go
```

Reconcile against the tables under § Inherited SPEC-021 doc-sweep
audit + § SPEC-023 forward-reference audit. If the output differs
materially (a chore commit landed between design and build that
touched one of the audit-affected files), STOP and ask the spec
author rather than expanding scope unilaterally.

### Smoke-test plan (ship-checklist sub-list, post-PR-merge)

Performed by the human at ship time. AC-37 through AC-40 verified
against the live release pipeline:

1. **Local snapshot smoke (AC-37).** From the user's macOS arm64 box:
   ```bash
   goreleaser build --snapshot --clean
   ls dist/
   ```
   Expect four `brag_<version>_<os>_<arch>` directories (or the
   collapsed name template per the goreleaser config) plus a `dist/`
   metadata directory. Each binary should run on its target arch.
   `dist/brag_*_darwin_arm64/brag --version` should print
   `brag version <snapshot-version>` (e.g. `0.0.0-next`).

2. **Pre-merge: create the GitHub repository secret.** Settings →
   Secrets and variables → Actions → New repository secret. Name:
   `HOMEBREW_TAP_GITHUB_TOKEN`. Value: a fine-grained PAT with
   "Contents: read+write" + "Metadata: read" on
   `github.com/jysf/homebrew-bragfile` only.

3. **Merge the SPEC-023 PR to `main`.** Via squash merge per the
   project's existing convention (`gh pr merge --squash
   --delete-branch`). Confirm the `ci` workflow goes green on
   `main` post-merge.

4. **RC-tag pre-flight (AC-38).** From `main`:
   ```bash
   git checkout main && git pull
   git tag v0.1.0-rc1
   git push origin v0.1.0-rc1
   ```
   Watch the `release` workflow run in the Actions tab. Expect
   green. Verify on the GitHub Releases page: a new prerelease
   `v0.1.0-rc1` with all four archives + `checksums.txt`. Verify
   on `github.com/jysf/homebrew-bragfile/blob/main/Formula/bragfile.rb`:
   the file does NOT exist (skip_upload: auto detected the
   prerelease).

5. **Cut v0.1.0 (AC-39).** Tag-and-push:
   ```bash
   git tag v0.1.0
   git push origin v0.1.0
   ```
   Watch the `release` workflow. Expect green. Verify on the
   GitHub Releases page: a new stable release `v0.1.0` with all
   four archives + `checksums.txt`. Verify on
   `github.com/jysf/homebrew-bragfile/blob/main/Formula/bragfile.rb`:
   the formula file exists with the v0.1.0 sha256 + URLs.

6. **End-to-end install smoke (AC-40).** From a clean shell on
   macOS arm64 without `brag` on `$PATH`:
   ```bash
   brew install jysf/bragfile/bragfile
   brag --version    # → "brag version 0.1.0"
   brag add --title "v0.1.0 brew-install smoke from clean shell"
   brag list         # row visible
   ```
   First successful install closes AC-40.

7. **Post-RC cleanup.** Delete the prerelease `v0.1.0-rc1` GitHub
   release + delete the `v0.1.0-rc1` git tag remotely:
   ```bash
   git push origin --delete v0.1.0-rc1
   gh release delete v0.1.0-rc1
   ```
   (Local `git tag -d v0.1.0-rc1` is housekeeping; remote
   deletion is the meaningful cleanup.)

8. **Branch protection on `main`.** Settings → Branches → Branch
   protection rules → Add rule for `main` → Require status checks
   to pass before merging → require `test (macos-latest)` and
   `test (ubuntu-latest)` (visible after the workflow runs at
   least once). Out of SPEC-023's literal-artifact scope but
   captured in the ship checklist for completeness.

### `bfa1474` archive-spec precondition reminder

`scripts/archive-spec.sh` (extended at commit `bfa1474`, 2026-04-25)
fail-fasts at `just archive-spec SPEC-023` if the Reflection (Ship)
section below has unfilled `<answer>` placeholders. Surface to the
ship session: the three Reflection (Ship) questions MUST receive
real answers before archive runs.

### `§10` push-discipline reminder

After any local commit on the SPEC-023 feat branch in the same shell
session as the PR merge, run `git push origin HEAD` before
`gh pr merge --squash --delete-branch`. Three precedents: SPEC-013
slip, SPEC-018 chore-bundle slip, SPEC-019 reflection-orphan slip.
Two clean confirmations: SPEC-021 ship + SPEC-022 ship. SPEC-023 is
the third confirmation that promotes the rule from "codified at
framing" to "load-bearing across the stage."

---

## Build Completion

*Filled in at the end of the **build** cycle, before advancing to verify.*

- **Branch:** `feat/spec-023-distribution-proper`
- **PR (if applicable):** not yet opened
- **All acceptance criteria met?** yes (AC-1 through AC-36; AC-37–AC-40 are smoke-tests performed at ship time)
- **New decisions emitted:** NONE (per stage notes)
- **Deviations from spec:**
  1. **L2 / `.goreleaser.yaml` ordering** — spec's `.goreleaser.yaml` literal places four comment lines before `version: 2` (putting it on line 6), but test L2 asserts `version: 2` appears within the first 5 lines (`head -n 5`). These two spec literals contradict each other. Resolution: moved `version: 2` to line 1 (standard goreleaser convention; consistent with the assertion's stated intent "opens with `version: 2`"). The comments follow on lines 2–5. No functional change.
  2. **O4 / CHANGELOG three command entries** — spec's CHANGELOG literal formats three commands as `` `brag show <id>` ``, `` `brag delete <id>` ``, `` `brag export --format markdown|json` ``, but test O4 checks for the plain verb forms `` `brag show` ``, `` `brag delete` ``, `` `brag export` ``. These are not substrings of the argument forms. Resolution: changed the three bullets to use the plain verb in backticks (consistent with how `` `brag add` ``, `` `brag list` ``, `` `brag stats` `` are formatted) and moved arguments/flags into the description text. No functional change to meaning.
  3. **AC-1 / `.goreleaser.yaml` deprecated keys (post-verify punch-list)** — `goreleaser check` exited 2 against the shipped config because goreleaser 2.15.4 deprecated two keys the spec's verbatim literal still used: (a) `archives[].format: tar.gz` deprecated in favour of `archives[].formats: [tar.gz]`; (b) `brews:` deprecated in favour of `homebrew_casks:`. The `brews:` → `homebrew_casks:` migration is not a pure key rename: `homebrew_casks` does not accept `install:`, `test:`, or `license:` sub-fields. Resolution: replaced `format:` with `formats: [tar.gz]`; replaced `brews:` block with `homebrew_casks:` block using `binaries: [brag]` (replaces the `install: bin.install "brag"` stanza) and dropped the unsupported `license:`, `test:` sub-fields; `repository:`, `homepage:`, `description:`, and `skip_upload:` carry over unchanged. Group L assertions L6/L10/L11 in `scripts/test-docs.sh` updated in lockstep to match the migrated reality (`brews:` → `homebrew_casks:`; `format: tar.gz` → `formats: [tar.gz]`; `license: "MIT"` replaced with `- brag` binaries-symlink check). `goreleaser check` now exits 0; `bash scripts/test-docs.sh` now reports 96/96; snapshot build still produces all four darwin+linux × amd64+arm64 targets. Locally verified: `goreleaser release --snapshot --clean --skip=publish` produces a well-formed `dist/homebrew/Casks/bragfile.rb` with `binary "brag"` directive and correct URL templating per OS/arch.
- **Follow-up work identified:** none beyond the existing SPEC-024 (shell completions)

### Build-phase reflection (3 questions, short answers)

Process-focused: how did the build go? What friction did the spec create?

1. **What was unclear in the spec that slowed you down?**
   — Nothing genuinely unclear. The two literal-vs-literal conflicts (L2 and O4) were caught quickly by the test harness on first run. The spec's structure and §12 pattern made the work mechanical and fast. The mandatory audit-grep cross-check was the most deliberate step and reconciled cleanly.

2. **Was there a constraint or decision that should have been listed but wasn't?**
   — No missing constraints. The note about `version: 2` placement would be worth a single-sentence clarification in the spec's goreleaser literal header (e.g. "version: 2 must be within the first 5 lines to satisfy assertion L2"), but it's a minor improvement.

3. **If you did this task again, what would you do differently?**
   — Run `just test-docs` after transcribing the test-docs.sh extension but BEFORE transcribing the artifact files, to confirm the extension itself is well-formed. In this build the extension ran correctly, but isolating that failure mode would be useful. Also: for CHANGELOG literals, standardize bullet format to verb-only backtick (no arguments) across all commands — the mixed format (verb-only for some, verb+args for others) is what triggered O4.

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
