---
# Maps to ContextCore epic-level conventions.
# A Stage is a coherent chunk of work within a Project.
# It has a spec backlog and ships as a unit when the backlog is done.

stage:
  id: STAGE-005                     # stable, zero-padded within the project
  status: proposed                  # proposed | active | shipped | cancelled | on_hold
  priority: medium                  # critical | high | medium | low
  target_complete: 2026-04-30       # ~3–5 days from framing; outside boundary 2026-05-04 per brief

project:
  id: PROJ-001                      # parent project
repo:
  id: bragfile

created_at: 2026-04-25
shipped_at: null
---

# STAGE-005: Distribution and cleanup

## What This Stage Is

Make the feature-complete `brag` CLI **distributable** and **introducable**:
when this stage ships, a stranger who lands on the repo via `brew install
jysf/bragfile/bragfile` (or the GitHub URL) can install the binary, read a
user-facing README that explains what `brag` does and how to use it, find
a checked-in JSON-Schema contract for the AI-integration story, copy a
working Claude Code session-end hook example, and wire shell-completion
into their zsh/bash/fish session — all without ever needing to know the
spec-driven development process the repo was built with. STAGE-001–004
shipped the *tool*; STAGE-005 ships the *project's public surface*. After
this stage ships, **PROJ-001 closes**.

## Why Now

Three reasons converge:

1. **STAGE-004 shipped 2026-04-25** with the digest trio (`brag summary`,
   `brag review`, `brag stats`) closing the feature surface for the
   personal-workflow use case. There is no further feature work the brief
   names; everything else has been triaged to backlog or to PROJ-002.
   Distribution is the only remaining work between "I use this daily" and
   "PROJ-001 is done."

2. **The README documents the dev process, not the tool.** External Claude
   review on 2026-04-24 surfaced what was already obvious in retrospect:
   `README.md` opens with the spec-driven workflow and the
   Repo→Project→Stage→Spec hierarchy, not with what `brag` does or how to
   capture an entry. Anyone landing on the repo via a homebrew install
   would be confused. The fix has to land *with* the homebrew install or
   the install ships strangers into a misframed doc.

3. **The user wants `brew install bragfile` for the learning value, not
   for adoption.** The motivation is: practice the distribution mechanics
   end-to-end (goreleaser + GitHub Actions + tap + CHANGELOG + completions)
   on a real personal project. No marketing push. This framing matters
   for scope discipline — we ship the mechanics correctly without
   over-investing in adoption-ready polish (analytics, install metrics,
   broad-platform support beyond darwin+linux × arm64+x86_64).

No external blockers. The pre-stage chores landed 2026-04-25:
`github.com/jysf/homebrew-bragfile` exists empty, AGENTS.md §10 has the
push-discipline rule codified, and `just archive-spec` rejects empty
`<answer>` placeholders. All 14 DECs apply forward unchanged; no new DECs
expected this stage (distribution mechanics are configuration, not
project-binding decisions).

## Success Criteria

- **`brew install jysf/bragfile/bragfile` works on macOS arm64.** From a
  clean shell on a Mac without `brag` already on `$PATH`, the tap-install
  flow produces a working `brag --version` and `brag add --title "x"` →
  `brag list` round-trip against `~/.bragfile/db.sqlite`. Same flow on
  darwin x86_64 if a test machine is available; linux arm64+x86_64 builds
  produced and uploaded by goreleaser as a nice-to-have but not gate.
- **README is user-facing.** Top-of-`README.md` answers "what is `brag`,
  why would I use it, how do I install it, how do I capture my first
  entry, where do entries live" within the first two scrolls — no mention
  of cycles, stages, or specs above the fold. Spec-driven dev-process
  content moved to `CONTRIBUTING.md` (or `docs/development.md` — SPEC-021
  picks). Internal links audited; nothing broken.
- **`docs/brag-entry.schema.json` exists and is referenced from BRAG.md.**
  Schema mirrors DEC-012's stdin contract (required `title`; optional
  `description, tags, project, type, impact`; server fields tolerated; no
  unknown keys). BRAG.md cross-reference points at it as the validation
  contract AI agents must produce against.
- **Claude Code session-end hook artifacts ship.**
  `scripts/claude-code-post-session.sh` is a working shell script that
  consumes the JSON schema and proposes a `brag add --json` invocation;
  the `/brag` slash command template is a markdown file under `examples/`
  (or similar); a short install-instructions section in `BRAG.md` (or a
  `scripts/README.md`) tells a user where to copy each artifact in
  `~/.claude/`. **No new `brag` CLI surface** (artifacts not installers,
  per Q3 framing answer).
- **CI runs on every PR.** GitHub Actions workflow runs `go test ./...`,
  `gofmt -l .` (fail on diff), and `go vet ./...` against the latest
  stable Go on macOS-latest and ubuntu-latest. PR cannot merge red.
- **Release on tag publishes binaries + auto-pushes formula to the tap.**
  `git tag v0.1.0 && git push origin v0.1.0` triggers the release workflow,
  goreleaser cross-compiles (darwin+linux × arm64+x86_64), uploads
  archives to the GitHub release, generates release notes from
  conventional-commit subjects, and pushes a `bragfile.rb` formula to
  `github.com/jysf/homebrew-bragfile`. `CHANGELOG.md` populated with a
  `[v0.1.0]` section.
- **`brag completion zsh|bash|fish` writes a working completion script.**
  Tab-completion works for `brag <tab>`, `brag add --<tab>`, `brag list
  --<tab>` after sourcing the generated script in a shell rc.
- All STAGE-001/002/003/004 success criteria still hold. `go test ./...`,
  `gofmt -l .`, `go vet ./...`, `CGO_ENABLED=0 go build ./...` clean
  through every spec.

## Scope

### In scope

- **`README.md` user-facing rewrite** + spec-driven dev-process content
  migrated to `CONTRIBUTING.md` (or `docs/development.md` — SPEC-021
  picks). Internal-link audit; redirect-from-old-paths only if needed.
  (SPEC-021)
- **`docs/brag-entry.schema.json`** — JSON Schema (draft 2020-12 or
  similar) mirroring DEC-012's stdin shape. BRAG.md cross-reference added
  pointing at it as the validation contract. (SPEC-022)
- **`scripts/claude-code-post-session.sh`** — shell script consuming the
  schema; emits a candidate `brag add --json` payload at session end.
  Plus a `/brag` slash command template (markdown under `examples/` or
  `templates/`) and a short install-instructions section explaining how
  to copy each artifact into `~/.claude/`. **Examples + docs only — no
  new `brag` CLI surface.** (SPEC-022)
- **`.goreleaser.yaml`** — cross-compile config (darwin+linux ×
  arm64+x86_64), archive naming, release-notes hook, brews block pointing
  at the tap repo. (SPEC-023)
- **`.github/workflows/ci.yml`** — test + lint + vet on PR (matrix
  macOS-latest + ubuntu-latest). (SPEC-023)
- **`.github/workflows/release.yml`** — tag-triggered release;
  cross-compiles, uploads to GitHub release, pushes formula to tap.
  (SPEC-023)
- **`CHANGELOG.md`** — Keep-A-Changelog style, retroactively populated
  for v0.1.0 from PROJ-001's commit history. (SPEC-023)
- **`brag completion zsh|bash|fish`** — cobra's built-in completion API,
  exposed as a subcommand. Help text + smoke tests. (SPEC-024)
- **Doc sweeps folded into originating specs** per the premise-audit
  rule: `AGENTS.md §3` distribution mention may need an "arrived" tense
  change post-SPEC-023 ship; `docs/tutorial.md` install section gets a
  brew-install paragraph; `BRAG.md` adds the schema cross-reference.
- **Blog-post draft** as an in-repo artifact (`docs/blog-spec-driven-bragfile.md`
  or similar) — write-up of the spec-driven process applied to PROJ-001.
  Artifact, **not a spec**. Publication target TBD post-draft (HN /
  dev.to / Substack / personal site / GitHub Discussion); the in-repo
  copy is canonical regardless.

### Explicitly out of scope

- **Marketing push.** No HN-front-page coordination, no Reddit posting
  campaign, no Show-HN. The user wants `brew install bragfile` for the
  learning value of shipping the mechanics correctly. Adoption is not a
  success criterion.
- **Pre-1.0 backward-compat promise** beyond what's already documented in
  `docs/api-contract.md`. v0.x while the contract is plastic; v1.0 is a
  separate decision after real usage at scale.
- **New `brag` CLI surface for installing artifacts** (e.g.
  `brag install-claude-hook`). Examples + docs only — copy-paste install
  by the user. Per Q3 framing answer 2026-04-25.
- **Tags-normalization migration** (e.g. moving from comma-joined string
  per DEC-004 to a proper `entry_tags` join table). External Claude
  review 2026-04-24 raised this as a v0.1-debt concern; user decision:
  accept the debt, revisit if/when a concrete pain materializes.
- **Soft-delete + edit-history.** Same shape as above — accepted v0.1
  debt; revisit on concrete pain.
- **Premise-audit sub-template extraction** flagged at SPEC-015 ship and
  re-flagged at STAGE-004 ship. Deferred to **PROJ-002 framing** (where
  feature work in the same shape returns) per Q5A framing answer
  2026-04-25; STAGE-005 specs are mostly new-file work and the
  extraction's value is concentrated in status-change work.
- **Pulling from `backlog.md`.** STAGE-005's scope is distribution-shape,
  not feature-shape; backlog items are feature-shaped and stay backlogged
  for PROJ-002 or later projects.
- **Linuxbrew, apt/yum packages, Windows support, Chocolatey.** Per brief
  scope: macOS-first; linux falls out of goreleaser if it's free but is
  not a success criterion.
- **Anything that introduces a new top-level Go dependency** without an
  accompanying DEC. (`no-new-top-level-deps-without-decision` constraint
  applies forward unchanged.)
- **LLM integration anywhere in the binary.** PROJ-002's reason for
  existing.

## Spec Backlog

Ordered by recommended build sequence. SPEC-021 ships first because it's
lowest risk, doesn't depend on other specs, and sets the user-facing
framing the later specs reference. SPEC-022 + SPEC-023 are independent
of each other and could run in parallel after SPEC-021 ships, but
sequencing them serially is cleaner. SPEC-024 ships last because it's
isolated and small.

- [x] SPEC-021 (shipped 2026-04-25, **M**) — **README user-facing
      rewrite + dev-process migration to `CONTRIBUTING.md` AND
      `docs/development.md`.** README rewritten as user-facing
      (install / capture / list / search / export / digests / where-
      data-lives / where-to-go-next / license). New CONTRIBUTING.md
      (thin GitHub-conventional pointer). New docs/development.md
      (spec-driven framework details). New scripts/test-docs.sh
      (40 grep-based shell asserts) exposed via new `just test-docs`
      recipe (NOT wired into `just test`). Shipped via PR #21
      (squash-merged `9abdeb6`). Clean cycle — 39/39 AC, 40/40
      scripted asserts, 8/8 byte-identity checks; both watch-patterns
      (§9 audit-grep cross-check + §12 NOT-contains self-audit)
      no-op'd at build (second consecutive spec); §10 push-discipline
      held on its first proactive application post-codification.
      Two AGENTS.md addenda earned + codified at this ship: §9
      BSD-grep `--exclude-dir` warning + §12 literal-artifact-as-spec
      pattern (three confirming cases SPEC-018/020/021). Inherits
      SPEC-023 punch list (deferred stale STAGE-NNN refs across
      docs/api-contract.md / architecture.md / data-model.md /
      tutorial.md:493).
- [x] SPEC-022 (shipped 2026-04-26, **M**) — **AI-integration
      distribution asset.** Three new artifacts shipped via the
      §12 literal-artifact-as-spec pattern (third application,
      first cross-format): `docs/brag-entry.schema.json` (50
      lines, JSON Schema draft 2020-12 mirroring DEC-012);
      `scripts/claude-code-post-session.sh` (67 lines, pure-stdin
      bash hook with chmod +x); `examples/brag-slash-command.md`
      (14 lines, tight slash-command template). Plus BRAG.md
      insertion (50 lines, "## JSON contract for programmatic
      capture" section) and scripts/test-docs.sh extension (161
      lines, 23 new asserts in groups H/I/J/K). Shipped via
      PR #22 (squash-merged `079bb89`). Clean cycle — 30/30 AC,
      63/63 test-docs (40 SPEC-021 + 23 new), all 10 rejected-
      alternatives held, no DEC. Cross-format literal-artifact
      pattern validated (JSON + bash + markdown + in-place
      markdown + shell extension all clean on first transcription
      with zero drift). Both watch-patterns no-op'd at build for
      the second consecutive spec; §10 push-discipline held on
      its second proactive application; bfa1474 archive-spec
      precondition exercised second time on a real ship and
      passed. One Q3 cosmetic observation (BRAG.md HR-separator
      asymmetry) WATCHED at N=1, not codified. STAGE-005 at
      2/4.
- [ ] SPEC-023 (designed 2026-04-26, build pending, **M**) —
      **Distribution proper: goreleaser + GitHub Actions + CHANGELOG
      + tap-formula auto-publish.** `.goreleaser.yaml` (cross-compile
      darwin+linux × arm64+amd64), `.github/workflows/ci.yml` (test +
      gofmt + vet matrix on macos-latest + ubuntu-latest),
      `.github/workflows/release.yml` (tag-`v*` → goreleaser release
      → brews-push to pre-existing `github.com/jysf/homebrew-bragfile`
      with `skip_upload: auto` for prereleases), `CHANGELOG.md`
      populated retroactively for v0.1.0 in Keep-A-Changelog 1.1.0
      shape (literal-artifact-as-spec, fourth confirming application).
      Plus six-line const→var rewiring of `cmd/brag/main.go` for
      ldflags injection, plus six-file doc sweep folding in the
      inherited SPEC-021 punch-list (architecture.md:45 sqlite-file-
      copy fix + architecture.md:103 STAGE-004→STAGE-005 +
      tutorial.md:493 §9 collapse) and SPEC-023's own forward-
      reference flips (README.md status banner + AGENTS.md §3 line 67
      + §4 lines 97/106 + cmd/brag/main.go:12 STAGE-004→STAGE-005),
      verified by 33 new test-docs.sh asserts in groups L–P (running
      total: 96 asserts after build). Bundle-vs-split decision
      explicit at top of spec — KEEP BUNDLED (single ship narrative;
      doc sweep activates as a unit; size lands at solid M not L; no
      operational gain from splitting). Zero new DECs expected.
      End-to-end smoke (AC-37/38/39/40): local `goreleaser build
      --snapshot --clean` + RC-tag pre-flight + `v0.1.0` tag-cut +
      `brew install jysf/bragfile/bragfile` from clean shell.
- [ ] SPEC-024 (not yet designed, **S**) — **Shell completions.**
      `brag completion zsh|bash|fish` subcommand wrapping cobra's
      built-in completion API. Smoke tests on at least zsh and bash;
      help text; tutorial.md install-section addendum showing how to
      source the script.

**Count:** 2 shipped / 0 active / 2 pending

**Complexity check:** 1 × S/M + 2 × M + 1 × S; no L-complexity entries
(none expected per framing prompt). Total ~4 specs, within the 3–5
healthy-stage band. SPEC-023 is the load-bearing one — if any spec slips
to L it's that one (goreleaser + two GHA workflows + CHANGELOG + tap
integration is the heaviest workstream). Watch at SPEC-023 design;
willing to split into "goreleaser + CI" + "release + tap + CHANGELOG"
two-spec form if scope inspection at design time argues for it.

## Design Notes

Cross-cutting patterns and per-spec direction. AGENTS.md §9 lessons all
apply unchanged (buffer split, tie-break, assertion specificity,
locked-decisions-need-tests, premise-audit family, audit-grep
cross-check, NOT-contains self-audit). §10 push-discipline rule codified
2026-04-25 as a pre-stage chore. `just archive-spec` precondition for
empty `<answer>` placeholders codified same day.

### Cross-cutting

- **No new DECs expected.** Distribution mechanics are configuration
  choices, not project-binding decisions. If a spec at design time finds
  a load-bearing choice that future projects in this repo will inherit
  (e.g. "every release publishes to homebrew tap X by convention"), then
  emit a DEC; otherwise document inline as a Locked Design Decision per
  spec convention. Confidence-target check: a choice at <0.8 inside a
  spec gets a question in `/guidance/questions.yaml` per AGENTS.md §14,
  same as any other stage.

- **Trim heuristic — STAGE-005 specs are heterogeneous, apply with care.**
  STAGE-004 ship reflection codified (N=1, soft): "first spec in any new
  stage keeps a fuller Notes-for-the-Implementer skeleton; specs 2+ can
  compress to signatures + invariants only when the first spec's shape
  is the construction precedent." STAGE-005's workstreams do **not**
  share a uniform shape — SPEC-021 is prose-heavy, SPEC-022 is
  shell+JSON+markdown, SPEC-023 is YAML+CI, SPEC-024 is Go-with-cobra.
  Recommendation: SPEC-021 keeps a fuller skeleton; SPEC-022/023/024
  apply trim only if their respective design sessions find construction
  precedent inside the same spec or in a shipped neighbor. **Default to
  fuller skeleton** if uncertain — the trim heuristic earned its keep on
  same-shape sibling specs (SPEC-018→020), not across format-distinct
  workstreams.

- **Premise audit applicability is asymmetric across specs.** STAGE-005
  is mostly new-file work (new YAML, new JSON, new shell script, new
  CHANGELOG) where the audit-family rules — inversion-removal,
  count-bump, status-change — don't have natural triggers. **The
  exception is SPEC-021** (README rewrite) which IS a status-change at
  scale: every doc that references the old README structure or
  development-process placement needs a sweep. SPEC-021 design must
  enumerate `grep -rn` patterns explicitly under `## Outputs` AND run
  them at design time per the SPEC-018 audit-grep cross-check rule.
  SPEC-022/023/024 will likely have minimal audit scope — design
  sessions should still walk the rule but should expect "no hits, no
  outputs" outcomes more often than not.

- **Doc sweeps fold into originating specs** per the premise-audit rule.
  Stage-level expected hot spots: `AGENTS.md §3` distribution mention
  changes from "(arriving in STAGE-005)" to "shipped" tense post-SPEC-023;
  `docs/tutorial.md` install section gets a brew-install paragraph
  post-SPEC-023; `BRAG.md` cross-reference to the schema lands in
  SPEC-022; the README itself is fully rewritten in SPEC-021.

- **Blog post is an artifact, not a spec.** Lives at
  `docs/blog-spec-driven-bragfile.md` (or similar — exact filename a
  drafting decision) as the canonical source of truth. Versioned with
  the project and reachable from `git log` if it informs future
  framework refinements (likely). Publication target — HN, dev.to,
  Substack, personal site, GitHub Discussion — is **independent and
  TBD**, decided post-draft. Whichever spec the user chooses to draft
  the post within (or as a separate non-spec session) leaves the
  publication-target decision as a TODO in the artifact.

- **Tap repo is pre-existing.** `github.com/jysf/homebrew-bragfile`
  was created 2026-04-25 as a pre-stage chore (gh repo create + tap-repo
  README pointing at the main repo). SPEC-023's PR is repo-local
  goreleaser+GHA code only; no manual tap-repo bootstrap inside the
  spec. The first tagged release auto-publishes the formula via
  goreleaser's `brews:` block.

- **CI test harness lessons all carry forward.** Per AGENTS.md §9:
  separate `outBuf` / `errBuf` per command test; assert no
  cross-leakage; line-based equality for any markdown heading-level
  assertion; ID-based freshness checks (not timestamp-based);
  fail-first run before implementation; every locked decision paired
  with a failing test. SPEC-024 (`brag completion`) is the only spec
  with new CLI surface this stage; the harness rules apply directly.
  SPEC-021/022/023 mostly produce non-Go artifacts whose verification
  may be golden-file or integration-style rather than unit-test —
  per-spec design-session call.

### SPEC-021-specific (`README` rewrite + dev-process migration)

- **Above-the-fold of README answers user questions, not contributor
  questions.** First scroll: name + tagline + status badge (if any) +
  one-paragraph description. Second scroll: install (brew + source). 
  Third scroll: first-entry capture + reading-back + where data lives.
  Fourth scroll: links to tutorial, BRAG.md, full command reference. Move
  the spec-driven workflow content (cycles, hierarchy, four-habits
  discipline) to `CONTRIBUTING.md` (recommended over
  `docs/development.md` — `CONTRIBUTING.md` is the GitHub-conventional
  filename and surfaces in PR-create UI).
- **Premise audit is the load-bearing concern.** `grep -rn "spec-driven"
  README.md docs/ AGENTS.md` and `grep -rn "Repo.*Project.*Stage.*Spec"
  README.md docs/ AGENTS.md` and `grep -rn "frame.*design.*build.*verify
  .*ship" README.md docs/` to find every reference to the dev-process
  framing. Each hit gets enumerated under `## Outputs` as either "stays
  here" or "moves to CONTRIBUTING.md". Run the greps at design time per
  audit-grep cross-check.
- **`GETTING_STARTED.md`** at repo root is a contributor onboarding doc
  (it walks through your first PROJECT, not your first brag entry). It
  may need a sibling rewrite or a clarifying preamble — design call.
- **`FIRST_SESSION_PROMPTS.md`** is purely for contributors / agents and
  stays as-is.

### SPEC-022-specific (JSON schema + Claude hook + slash command)

- **JSON Schema source of truth.** Use draft 2020-12 (currently the
  most-supported modern draft); `$id` pointing at the
  `github.com/jysf/bragfile000/blob/main/docs/brag-entry.schema.json` URL.
  Mirror DEC-012 exactly: `title` required + non-empty; `description,
  tags, project, type, impact` optional strings; `id, created_at,
  updated_at` permitted-but-ignored (`additionalProperties: false`
  rejects unknowns to match the strict-reject rule).
- **`brag add --json`'s actual implementation does not validate against
  this schema.** The schema is documentation of the contract for
  external consumers (AI agents, integrations); the binary's stdin
  parser is the authoritative validator. **Failing test idea:** a
  small Go test reads a fixture entry, passes it through `Store.Add`'s
  JSON path, AND validates it against the checked-in schema (using
  `gojsonschema` or similar) — proves the schema and the binary's
  parser agree at one moment in time. Keeps the two from drifting
  silently. Adding `gojsonschema` is a new top-level dep needing a
  DEC; alternative is a Python or Node validation step in CI (no Go
  dep change). Design call.
- **Hook script artifacts:** `scripts/claude-code-post-session.sh` is a
  bash script readable by a developer (well-commented, no obscure
  jq incantations), reads transcript context, emits a candidate
  `brag add --json` payload to stdout for the user to review. The
  `/brag` slash-command template is a markdown file (likely under
  `examples/` per convention) with the prompt that triggers Claude to
  draft a brag entry from the current session. `BRAG.md` (or a new
  `scripts/README.md`) explains where each artifact goes in
  `~/.claude/`. **Confirm Claude Code's hook config + slash-command
  paths at design time** — `~/.claude/settings.json` for hooks,
  `~/.claude/commands/<name>.md` for slash commands as of late 2025;
  re-verify against current docs.
- **Cross-reference from BRAG.md** to the schema — natural fit in the
  "Fields" section or a new "JSON contract" section.

### SPEC-023-specific (goreleaser + GHA + CHANGELOG + tap auto-publish)

- **Goreleaser config** is the heaviest piece. Cross-compile matrix:
  darwin × {arm64, amd64}, linux × {arm64, amd64}. Archive format: tar.gz
  with embedded LICENSE + README. Checksum file. Use goreleaser's
  built-in changelog generator (group by conventional-commit type) for
  release notes; CHANGELOG.md gets a v0.1.0 section retroactively from
  PROJ-001's commit history at SPEC-023 design or build time (decision
  call).
- **`brews:` block** points at `github.com/jysf/homebrew-bragfile`. Needs
  a GitHub token with `repo` scope on the tap repo — use a personal
  access token stored as `HOMEBREW_TAP_GITHUB_TOKEN` repository secret;
  the release workflow exposes it to goreleaser.
- **CI workflow** runs on PR + push-to-main: `go test ./...`, `gofmt -l
  .` (fail on diff), `go vet ./...`. Matrix: latest stable Go × {macos-latest,
  ubuntu-latest}. Branch-protection update on `main` to require this
  workflow as a green-check before merge — that's a manual GitHub UI
  step, not in the workflow YAML; SPEC-023 ship checklist captures it.
- **Release workflow** triggers on tag push matching `v*.*.*`. Runs
  goreleaser; goreleaser handles cross-compile + GitHub release upload +
  tap formula push.
- **CHANGELOG.md** style: Keep-A-Changelog or a simpler conventional-commit
  digest — design call. Retroactive v0.1.0 section pulls from PROJ-001's
  commit history (`git log --oneline 902ed71` or similar — i.e. since the
  STAGE-004 close).
- **Smoke-test plan for ship checklist:** local `goreleaser build
  --snapshot --clean` produces archives without uploading; one
  end-to-end `git tag v0.1.0-rc1 && git push origin v0.1.0-rc1` against
  a pre-release tag verifies the workflow before cutting v0.1.0 proper.

### SPEC-024-specific (`brag completion zsh|bash|fish`)

- **Cobra has built-in completion** generators (`Command.GenZshCompletion`
  / `GenBashCompletion` / `GenFishCompletion`); the work is wiring them
  to a `completion` subcommand with one positional arg
  (`zsh|bash|fish|powershell` — possibly skip powershell unless trivial).
- **Smoke-test pattern:** generate the script to a buffer, parse first
  ~100 lines, assert presence of expected shell-specific markers (e.g.
  `_brag_completion()` for bash, `#compdef brag` for zsh). Goldens are
  brittle across cobra version bumps; presence-of-marker tests are
  robust enough.
- **Tutorial-section addendum** (post-ship doc sweep) shows how to source
  the generated script in zsh/bash rc files.

## Dependencies

### Depends on

- **STAGE-001 through STAGE-004 (all shipped)** — the entire feature
  surface. README rewrite cites the digest commands; goreleaser packages
  the binary that includes them; tutorial.md install section updates
  reflect them.
- **DEC-001 through DEC-014** — all apply forward unchanged. No new DECs
  expected this stage.
- **External: `github.com/jysf/homebrew-bragfile` tap repo.** Created
  2026-04-25 as a pre-stage chore. SPEC-023's `brews:` block targets it.
- **External: GitHub Actions secret `HOMEBREW_TAP_GITHUB_TOKEN`.** Will
  be created during SPEC-023 build (manual UI step, captured in ship
  checklist) — a personal access token with `repo` scope on the tap.
- **External: `brew` installed on the user's macOS** for end-to-end
  verification of the `brew install jysf/bragfile/bragfile` success
  criterion.
- **External (tooling): `goreleaser` available locally.** `brew install
  goreleaser` per AGENTS.md §3 macOS note. CI installs via the
  `goreleaser/goreleaser-action` GitHub Action.

### Enables

- **PROJ-001 closes** when STAGE-005 ships. This is the last stage of
  the project per `brief.md`'s Stage Plan.
- **PROJ-002 — AI assist (when opened).** A distributable bragfile binary
  is what AI agents will integrate against; the JSON schema artifact is
  the contract those agents validate against; the Claude hook example
  is the in-the-wild reference implementation they pattern off.
- **Future TUI / sync / cloud-backup projects.** All layer on a
  distributable binary, not source builds. Shipping the homebrew tap
  flow once means future projects can publish to the same tap without
  re-litigating the mechanics.
- **External adoption (if it ever becomes desired).** Not a goal of
  PROJ-001 per brief, but the mechanics shipping correctly leaves the
  door open without further work.

## Stage-Level Reflection

*Filled in when status moves to shipped. Run Prompt 1d (Stage Ship) in
FIRST_SESSION_PROMPTS.md to draft this in a fresh session.*

- **Did we deliver the outcome in "What This Stage Is"?** <yes/no + notes>
- **How many specs did it actually take?** <number vs. plan>
- **What changed between starting and shipping?** <one sentence>
- **Lessons that should update AGENTS.md, templates, or constraints?**
  - <one-line updates>
- **Should any spec-level reflections be promoted to stage-level lessons?**
  - <one-line items>
