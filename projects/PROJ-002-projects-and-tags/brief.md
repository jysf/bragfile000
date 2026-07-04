---
project:
  id: PROJ-002
  status: shipped
  priority: medium
  target_ship: 2026-06-26

repo:
  id: bragfile

created_at: 2026-06-05
shipped_at: 2026-06-19
---

# PROJ-002: Projects and tags — a first-class data model

## What This Project Is

Deepen bragfile's data model. Add `projects` as a first-class entity —
a named thing with a filesystem location, a status, and a state note —
so the tool can answer "what am I working on, and where does it live on
this machine?" And normalize the existing comma-joined-string `tags`
storage (DEC-004) into a polymorphic many-to-many model that both
entries and projects share. This is personal-capture deepening, not AI
assist: the wave makes the *structure* richer, not the *output* smarter.

The slug is `projects-and-tags` deliberately, not `v0.2.0` or `phase-2`.
PROJ-001 was named "MVP" generically because it had no surface to
describe. PROJ-002 has a precise one — two new/normalized object shapes
— and the slug names that surface so a future reader scanning
`/projects/` understands what the wave changed without opening the
brief. Version numbers and phase labels are metadata; the slug is the
description.

## Why Now

Four reasons converge.

1. **The feature surface is complete; the data model is not.** bragfile
   v0.1.0 ships the full brag-entry surface (capture, retrieve, filter,
   search, export, the digest trio), but `entries.project` is still a
   free-text string. There is no first-class concept of a project — no
   way to ask the tool what projects exist, where they live, or what
   state they're in.

2. **The user forgets where projects live.** A recurring real pain:
   losing track of which working directory holds which project on the
   machine. A `projects` entity carrying `location` + `status` + a
   state/next-action note closes that gap directly, and `brag project
   here` (cwd auto-detect) makes the lookup zero-friction.

3. **DEC-004 is accepted v0.1 debt that gets cheaper to repay now than
   later.** DEC-004 (confidence 0.65) stored tags as a comma-joined
   `TEXT` column for the MVP, explicitly flagging "revisit if tag rename
   becomes a user ask" and "the normalized-table option becomes more
   attractive" once a second consumer appears. PROJ-002 *is* that second
   consumer: we're adding a second taggable object type. Normalizing
   while we touch the tag surface anyway is cheaper than normalizing
   later — the refactor cost compounds with each additional CSV-tag
   table we'd otherwise have to migrate independently.

4. **The polymorphic shape is designed to generalize past two types.**
   The new tags model is sketched with a third taggable object in mind
   (probable candidate: `goals`). That third type is paper-sketched in
   the successor DEC to validate the abstraction generalizes — it is
   **not built** in PROJ-002.

**Operational note — dev/prod DB isolation during development.** PROJ-002
introduces a schema migration, so the development binary must never open
the user's production database in v0.2.x format while v0.2.x is still
in flight. The isolation story for the duration of this project:

- Dev binary built via `just install` → `~/go/bin/brag`; dev DB at
  `~/.bragfile-dev/db.sqlite` via `BRAGFILE_DB` env or `--db` flag.
- Production binary stays brew-installed at v0.1.0
  (`/opt/homebrew/bin/brag`) until a deliberate `brew upgrade`.
- Production DB at `~/.bragfile/db.sqlite` is never opened by a
  v0.2.x-format binary during development.
- At v0.2.0 ship: a documented manual backup of the production DB
  (`cp` recipe in CHANGELOG + tutorial), then the upgrade, then an
  optional migration-prompt safety belt if STAGE-008 ships it.

## Success Criteria

- **Projects are first-class.** `brag project new <name> --path <dir>`
  registers a project with a location, a status, and a state note;
  `brag project here` run from inside that directory surfaces the
  matching project.
- **Projects are scannable.** `brag project status` lists active
  projects sorted by recency, showing each one's state note and recent
  brag count.
- **Capture knows where it is.** `brag add` run from inside a registered
  project's directory auto-fills `--project` from the cwd.
- **Tags are a taxonomy, not just per-row strings.** `brag tags` lists
  every tag across entries *and* projects with counts; `brag tag rename
  <old> <new>` updates references globally; `brag tag merge` folds one
  tag into another.
- **The migration is lossless and invisible at the query surface.** The
  existing `entries.tags` CSV column migrates fully to the polymorphic
  schema with zero data loss, and existing `brag list --tag foo` /
  `brag search` queries keep working with no user-visible change.
- **v0.2.0 ships and upgrades cleanly.** v0.2.0 reaches the public
  Homebrew tap and `brew upgrade jysf/bragfile/bragfile` cleanly moves a
  v0.1.x install forward.
- **v0.1.x users are not disrupted mid-development, and the upgrade is a
  deliberate, well-messaged action.** Existing v0.1.x users can keep
  running v0.1.x against `~/.bragfile/db.sqlite` while v0.2.x is in
  development; the move to v0.2.x schema is a documented, opt-in action
  with a spelled-out backup workflow.
- **No regressions.** All PROJ-001 success criteria still hold; the full
  v0.1 feature surface works unchanged on the v0.2.0 binary.

## Scope

### In scope

- **Projects entity.** Schema (`projects` table + a `project_locations`
  join for multi-directory support); CRUD commands
  (new/list/show/edit/archive/delete); a status enum and a
  state/next-action note model; `brag project here` cwd auto-detect;
  `brag add` integration that auto-fills `--project` from the cwd of a
  registered project.
- **Tags polymorphic normalization (Path B).** Schema migration adding
  `tags` + `taggings` tables with ETL from the `entries.tags` CSV;
  FTS5 trigger maintenance so `entries_fts.tags` stays synced against
  the new join; refactor of every tag-touching command (`brag list
  --tag`, `brag search`, the exports, and the summary/review/stats
  digests) to read through the join. A new DEC supersedes DEC-004.
- **Tag taxonomy surface.** `brag tags` taxonomy view; `brag tag rename`
  and `brag tag merge`. These are near-free byproducts of the
  normalization and are what make the upgrade visibly worthwhile to the
  user rather than a silent internal refactor.
- **v0.2.0 release mechanics.** CHANGELOG `[0.2.0]` entry; the RC-tag
  pattern from AGENTS.md §4 (optional `v0.2.0-rc1` smoke-test under the
  dual-tag-on-same-commit rule); an optional migration-prompt safety
  belt as a STAGE-008 candidate.
- **Doc sweep.** `docs/tutorial.md` (new projects + tags sections);
  `docs/api-contract.md` (new commands + the new tags surface);
  `docs/architecture.md` (refresh the diagram + the responsibilities
  table for the project surface — which lives in
  `internal/storage/project.go` + `internal/cli/project.go`, **not** a
  separate `internal/projects` package; see the project-close drift
  reconciliation); `CHANGELOG.md` `[0.2.0]` section.

### Explicitly out of scope

- **Goals as a shipped object type.** Paper-sketched in the new
  polymorphic-tags DEC to prove the abstraction generalizes past two
  types, but **no code**. Deferred to PROJ-003 if dogfooding surfaces
  real signal during v0.1.0 use. (See `backlog.md` "goals / levels".)
- **Journal / blog as object types.** Ruled out at framing — wrong shape
  for a personal-capture tool.
- **AI assist / LLM integration.** Originally PROJ-002's raison d'être
  per PROJ-001's brief, **explicitly dropped at this framing.** Grounds:
  (a) the user already drafts brag content with Claude externally, so
  the in-tool value is marginal; (b) embedding AI makes the binary
  heavyweight (vendor SDKs, network deps, API-key plumbing) against
  bragfile's local-first ethos. Deferred indefinitely; the
  `brag summarize` backlog entry stays put.
- **Reverse-direction migrations.** PROJ-002 migrations are forward-only,
  same as PROJ-001's. Users wanting a downgrade path back up their
  SQLite file before upgrading; CHANGELOG + tutorial spell this out.
- **Soft-delete / edit-history.** Accepted v0.1 debt; not blocking
  PROJ-002.
- **macOS notarization.** Separately tracked at
  `docs/macos-notarization-checklist.md`; trigger conditions unchanged.
- **TUI / web interface / sync / cloud / multi-user.** Ruled out at
  PROJ-001 framing; still ruled out.
- **Most `backlog.md` entries that aren't Projects- or tags-adjacent.**
  Backlog stays backlogged unless a stage framing pulls it in.

## Stage Plan

Three stages. The sequencing constraint is load-bearing: **STAGE-006
ships first and is not reorderable**, because every later spec — every
Projects command, every digest, every export — depends on the
normalized tag schema. Building Projects against the old CSV-tag shape
would mean migrating it twice.

- [x] STAGE-006 (shipped 2026-06-07) — **Tag normalization.** ~3–4 specs
      estimated; **2 shipped** (SPEC-025 L, SPEC-026 M — atomic-migration
      rescope held the L as one spec). Schema migration (`tags` +
      `taggings` tables; ETL from the `entries.tags` CSV); DEC-015
      superseded DEC-004; FTS5 trigger maintenance to keep
      `entries_fts.tags` synced; refactor of `brag list --tag`, `brag
      search`, the exports, and summary/review/stats to read through the
      join; the `brag tags` taxonomy view; `brag tag rename` / `brag tag
      merge`. DEC-015 + DEC-016 emitted at design time.
- [x] STAGE-007 (shipped 2026-06-12) — **Projects (core).** ~5–6 specs
      estimated; **7 shipped** (SPEC-027..033 — one conscious peel:
      SPEC-029's L-watch fired and split location editing into SPEC-033).
      Schema (`projects` table + `project_locations` join for
      multi-directory support); CRUD (new/list/show/edit/archive/
      delete); status + state (the state-note model + the `brag project
      status` dashboard); `brag project here` cwd auto-detect; `brag
      add` integration (auto-fill `--project` from a registered
      project's cwd). DEC-017..020 emitted at design time. **Projects
      live in `internal/storage/project.go` + `internal/cli/project.go`,
      not a separate `internal/projects` package** (Scope drift,
      reconciled at project close).
- [x] STAGE-008 (shipped 2026-06-19) — **Polish and v0.2.0 release.**
      ~2–3 specs estimated; **4 shipped** (SPEC-034 doc sweep M, SPEC-035
      CHANGELOG S, SPEC-036 migration safety belt S/M, SPEC-037 v0.2.0
      release cut S). Comprehensive doc sweep
      (tutorial/api-contract/architecture + WAL-safe backup recipe);
      CHANGELOG `[0.2.0]`; the migration auto-backup safety belt
      (**promoted from "optional" to in-scope** on the strength of this
      session's prod-DB incident; **DEC-021** emitted); and the v0.2.0
      release cut. **Stage-name drift reconciled:** this plan line called
      STAGE-008 "Polish + integration"; the stage file is titled "polish
      and v0.2.0 release" (the release is the defining deliverable) —
      that title governs.

**Count:** 3 shipped / 0 active / 0 pending — **PROJ-002 complete**

## Dependencies

### Depends on

- **PROJ-001 (shipped 2026-05-17).** All 14 DECs apply forward
  unchanged; PROJ-002 will likely *supersede* DEC-004 via the new
  polymorphic-tags DEC at STAGE-006 design time, but no pre-emptive
  supersession review is needed at framing. AGENTS.md conventions all
  apply: §9 audit family, §10 push-discipline, §12 literal-artifact-as-
  spec + §12(b) design-time pre-flight + the codification meta-rule,
  §13 fresh-session rule. The §4 release addenda earned at v0.1.0 cut
  (dual-tag-on-same-commit recovery + macOS Gatekeeper xattr workaround)
  apply directly to the v0.2.0 cut.
- **WATCH-list framework rules carried from PROJ-001 close.** PROJ-002's
  specs are the natural next test cases for each; none is codified yet:
  - §12 sub-rule (a) — "run embedded test assertions against embedded
    literals at design time" (N=2 within SPEC-023; awaits a third
    confirming case across a distinct spec).
  - Trust-but-verify agent push reports (N=2 within SPEC-023;
    coordinator-discipline reflex, mechanically adjacent to §10/§13).
  - §13 fresh-session working-tree state preservation (N=1 from the
    SPEC-024 build cycle).
- **Premise-audit sub-template extraction** — flagged at SPEC-015 ship,
  re-flagged at STAGE-004 / STAGE-005, carried to PROJ-002 framing.
  **Decision deferred to STAGE-006 framing:** this is the first project
  where same-shape feature work returns at volume (Projects CRUD,
  digest refactors), so the extraction's value concentrates here — but
  whether to extract before the first PROJ-002 spec or hold once more is
  a stage-framing call, not a project-framing one.
- **External: none new.** The Homebrew tap
  (`github.com/jysf/homebrew-bragfile`) exists; repo secrets
  (`HOMEBREW_TAP_GITHUB_TOKEN`, GitHub Advanced Security) are all
  configured from PROJ-001.

### Enables

- **PROJ-003 (if framed).** The polymorphic-tags abstraction validated
  here is what a third taggable object type (`goals` the probable
  candidate) would build on without re-litigating the tag schema.
- **Richer cross-object views.** Once both entries and projects are
  taggable and projects are first-class, future work (a project-scoped
  digest, a tag-faceted dashboard) layers on this model rather than
  re-deriving it.
- **AI assist, whenever it returns.** A structured projects + tags model
  is a better substrate for any future external AI consumer than the
  free-text `project` field and CSV tags were.

## Project-Level Reflection

*Drafted via Prompt 1e (Project Close) in a fresh session on 2026-06-19;
all success criteria independently re-verified at close.*

### Success Criteria check

The brief enumerated eight success bullets. All eight shipped and were
re-verified at close (not assumed):

- **Projects are first-class.** ✅ `brag project new <name> --path <dir>`
  registers a project; `brag project here` surfaces the matching project
  from inside a registered directory (STAGE-007, SPEC-027/028/031).
- **Projects are scannable.** ✅ `brag project status` lists active
  projects by recency with state note + recent-brag count (SPEC-030).
- **Capture knows where it is.** ✅ `brag add` auto-fills `--project`
  from a registered project's cwd (SPEC-032).
- **Tags are a taxonomy, not just per-row strings.** ✅ `brag tags`
  counts across entries *and* projects; `brag tag rename` / `brag tag
  merge` mutate globally (STAGE-006, SPEC-025/026). The polymorphic
  prediction was confirmed empirically at the STAGE-007 close (a
  hand-written `'project'` tagging was counted by `brag tags` with zero
  code change).
- **The migration is lossless and invisible at the query surface.** ✅
  `entries.tags` CSV migrated to `tags` + `taggings` with zero data loss;
  `brag list --tag` / `brag search` byte-stable before/after (SPEC-025;
  DEC-017 soft match keeps `entries.project` lossless too).
- **v0.2.0 ships and upgrades cleanly.** ✅ Re-verified at this close:
  `gh release view v0.2.0` → `{"isPrerelease":false,"tagName":"v0.2.0"}`
  (full release), and `jysf/homebrew-bragfile`'s `Casks/bragfile.rb` is at
  `version "0.2.0"` with four matching sha256s. The clean v0.1.x → v0.2.0
  upgrade was proven during the SPEC-037 RC smoke gate against a seeded
  throwaway v0.1.x DB (which also fired the DEC-021 safety belt).
- **v0.1.x users are not disrupted mid-development; the upgrade is
  deliberate and well-messaged.** ✅ The dev/prod isolation discipline held
  for most of the project; the one incident (a v0.2.x dev binary migrating
  the prod DB) is exactly what produced the DEC-021 safety belt, turning a
  gap into a shipped guard. The CHANGELOG + tutorial spell out the backup
  workflow (WAL-safe `.backup` recipe).
- **No regressions.** ✅ Re-verified at close: `CGO_ENABLED=0 go test
  ./...` → 536 tests across 8 packages pass; `gofmt -l .` empty; `go vet
  ./...` clean; `bash scripts/test-docs.sh` exits 0.

### Reflection questions

- **Did we deliver the outcome in "What This Project Is"?** Yes. bragfile's
  data model is deepened exactly as framed: projects are a first-class
  entity (name + filesystem locations + status + state note, answering
  "what am I working on and where does it live?"), and the comma-joined
  `entries.tags` CSV is normalized into a polymorphic many-to-many
  `tags`/`taggings` model that entries and projects share. It is
  structure-deepening, not AI-assist — the AI scope the PROJ-001 brief had
  pencilled in for "PROJ-002" was deliberately dropped at this project's
  framing (local-first ethos; the user drafts brag content with Claude
  externally already). v0.2.0 is released, installable, and safe to upgrade
  into.
- **How many stages did it actually take?** **3** (STAGE-006/007/008),
  matching the brief's plan with **no reordering and no stage cancelled**.
  The load-bearing sequencing constraint (tags first, because every later
  spec depends on the normalized schema) held. Spec counts ran slightly
  above the per-stage estimates but for conscious, logged reasons:
  STAGE-006 **2** (vs ~3–4; atomic-migration rescope), STAGE-007 **7** (vs
  ~5–6; one L-watch peel into SPEC-033), STAGE-008 **4** (vs ~2–3; the
  incident-promoted safety belt + a separated release cut). **13 specs
  total (SPEC-025..037), 7 DECs (DEC-015..021)** — DEC-015 superseding
  DEC-004 (the planned conditional successor DEC-004 itself anticipated),
  the rest net-new; **zero DEC deprecations, zero supersession reversals.**
  All three stages shipped ahead of their `target_complete`, and the
  project shipped a week ahead of `target_ship` 2026-06-26.
- **What changed between starting and shipping?** Two material changes,
  both conscious and logged at the moment they happened: (1) the
  migration-safety belt moved from the brief's "optional, high
  safety-vs-cost" status to an in-scope STAGE-008 success criterion after a
  real prod-DB migration incident this session (the gap became the motive
  for DEC-021); (2) macOS notarization, which the user decided to pursue
  after hitting the install friction first-hand at the v0.2.0 cut, was
  scoped *out* of v0.2.0 and *into* a separate v0.2.1 effort. Otherwise the
  framing held intact — every DEC held verbatim from design lock to ship.
- **Lessons that should update AGENTS.md, templates, or constraints?**
  - **Codified at the STAGE-008 close (this stage, this project):**
    - **AGENTS.md §9 — injectable-os-var seam** (`os.Getwd`/`os.Getenv`/
      clock through a package `var`). N=3 same-outcome (SPEC-031/032/036).
    - **AGENTS.md §4 — Homebrew 6.0+ third-party tap trust** (`brew trust
      --cask`; tap-level, not avoided by notarization or a cask→formula
      switch). A concrete operational gotcha, codified at first occurrence
      like the existing §4 dual-tag and Gatekeeper notes.
  - **Earlier in the project:** §12(a) — "design-time pre-flight also covers
    the test's own expected-value literals" (codified at STAGE-006 close,
    N=3); flag-default-explicitness in literal-artifact CLI specs (codified
    at STAGE-007 close, N=3, AGENTS.md §12).
  - **Meta-signal:** every codification this project landed at the **stage
    close where its bar (N=3 same-outcome, or first-occurrence for a
    concrete operational fact) was reached** — the per-stage "WATCH-list
    update → carry-forward" ledger worked exactly as designed across all
    three stages, with no item force-codified early and no item silently
    dropped. No new durable framework artifact was needed.
  - **Bookkeeping (history correction, not rewrite):** SPEC-027's archived
    ship reflection refers to "the codified §12 contamination-exception,"
    but that heuristic was **never codified** — it has sat on WATCH since
    the STAGE-006 close (N=1) and is at N=2 now (SPEC-026 origin + SPEC-027
    confirming), still uncodified. This was flagged at the STAGE-007 close
    and is re-confirmed here; per §13 the archived spec is **annotated, not
    rewritten** — the loose phrasing stands in the archive with this
    project-close correction as its authoritative gloss.
- **What did we defer to the next project?**
  - **The "Impact & Fun" cluster → PROJ-003 ("delight + impact").** A
    brainstorm cluster captured 2026-06-16 (backlog.md): milestone
    notifications on `brag add` (TTY-only, stderr); an impact-axis digest
    (`brag impact --quarter|--month|--year`) plus an AI-pipe quarterly
    "super-brag"; and a stats reframe rewarding impact over volume. The
    through-lines are *passive surfacing* (features fire on actions the user
    already takes) and *impact over volume*. The polymorphic `taggings` +
    first-class projects substrate this project shipped is exactly what
    makes most of these cheap. Decide scope at PROJ-003 framing, informed by
    real v0.2.0 dogfooding. **Not scaffolded.**
  - **macOS code signing + notarization → v0.2.1 "macOS distribution
    hardening."** Now in scope (the user will pay the $99 Apple Developer
    fee) but on external Apple lead time, so a standalone patch-release
    effort, not coupled to PROJ-003. Note: notarization removes the
    Gatekeeper prompt but **not** the `brew trust` step — distinct
    frictions (backlog.md).
  - **The adjacent-data research prompt** — a self-contained, web-enabled
    research prompt (committed earlier this project) feeding future
    data-model direction; carried as PROJ-003 input.
  - **Carried WATCH-list items (uncodified, → PROJ-003 framing):** the
    contamination-heuristic exception (N=2, paired-opposing candidate);
    grade-by-intent for doc ACs (N=1); design-time library pre-flight should
    name the exact dep version (N=1); cobra `SilenceErrors`+`Fprintln`
    example (N=2); "AC-says-all-N → test each" (N=1); trust-but-verify push
    reports (N=2, variant-specific); §13 working-tree preservation (N=2).
    None force-codified at the project close; they carry to PROJ-003 as the
    PROJ-001 WATCH items carried into PROJ-002.

### Numbering after close

Next free IDs are **DEC-022** and **SPEC-038** (highest consumed:
DEC-021, SPEC-037). No project/stage IDs are reserved — PROJ-003's first
stage would be **STAGE-009**, continuing the repo-global monotonic
sequence (§2).

### PROJ-003 tee-up (one paragraph; nothing scaffolded)

If framed, **PROJ-003 is the "delight + impact" wave**: turn the now-rich
record (first-class projects + polymorphic tags) from a log into output
that is *fun, interesting, and impactful* — passive milestone
notifications on capture, an impact-first quarterly digest plus an
AI-pipe "super-brag," and a stats reframe that rewards captured impact
over raw volume. The north star the user named is **agent-native
accomplishment memory**: agents do the work *and* record why it mattered,
and humans (and agents) read stories back out for reviews, promos, and
identity. The PROJ-002 substrate makes most of it cheap; a third taggable
type (`goals`) and richer cross-object views also layer on the validated
polymorphic model without re-litigating the schema. Frame in a separate
Prompt 1a cycle, informed by real v0.2.0 dogfooding — **not scaffolded
here.**
