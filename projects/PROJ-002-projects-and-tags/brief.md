---
project:
  id: PROJ-002
  status: active
  priority: medium
  target_ship: 2026-06-26

repo:
  id: bragfile

created_at: 2026-06-05
shipped_at: null
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
  table for the new `internal/projects` package); `CHANGELOG.md`
  `[0.2.0]` section.

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

- [ ] STAGE-006 (not yet framed) — **Tag normalization.** ~3–4 specs.
      Schema migration (`tags` + `taggings` tables; ETL from the
      `entries.tags` CSV); a new DEC superseding DEC-004; FTS5 trigger
      maintenance to keep `entries_fts.tags` synced; refactor of `brag
      list --tag`, `brag search`, the exports, and summary/review/stats
      to read through the join; the `brag tags` taxonomy view; `brag tag
      rename` / `brag tag merge`. The new DEC (likely DEC-015) emits at
      this stage's design time, not at framing.
- [ ] STAGE-007 (not yet framed) — **Projects (core).** ~5–6 specs.
      Schema (`projects` table + `project_locations` join for
      multi-directory support); CRUD (new/list/show/edit/archive/
      delete); status + state (the state-note model + the `brag project
      status` dashboard); `brag project here` cwd auto-detect; `brag
      add` integration (auto-fill `--project` from a registered
      project's cwd).
- [ ] STAGE-008 (not yet framed) — **Polish + integration.** ~2–3 specs.
      Tutorial addendum for the projects + tags workflows;
      `docs/api-contract.md` + `docs/architecture.md` refresh; CHANGELOG
      `[0.2.0]`; the optional migration-prompt safety belt (an S spec
      with a high safety-vs-cost ratio); any edge cases surfaced during
      the prior two stages. A possible DEC-016 lands here if the
      migration-prompt safety belt ships and encodes a project-binding
      choice.

**Count:** 0 shipped / 0 active / 3 pending

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

*Filled in when status moves to shipped (via Prompt 1e, in a fresh
session).*

- **Did we deliver the outcome in "What This Project Is"?** <yes/no + notes>
- **How many stages did it actually take?** <number, compare to plan>
- **What changed between starting and shipping?** <one or two sentences>
- **Lessons that should update AGENTS.md, templates, or constraints?**
  - <one-line updates>
- **What did we defer to the next project?**
  - <one-line items>
