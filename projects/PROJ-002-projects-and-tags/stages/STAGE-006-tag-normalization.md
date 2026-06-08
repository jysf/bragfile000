---
# Maps to ContextCore epic-level conventions.
# A Stage is a coherent chunk of work within a Project.
# It has a spec backlog and ships as a unit when the backlog is done.

stage:
  id: STAGE-006                     # stable, zero-padded, repo-global (never reused)
  status: proposed                  # proposed | active | shipped | cancelled | on_hold
  priority: high                    # critical | high | medium | low
  target_complete: 2026-06-13       # ~1 week; first + load-bearing stage of PROJ-002

project:
  id: PROJ-002                      # parent project
repo:
  id: bragfile

created_at: 2026-06-05
shipped_at: null
---

# STAGE-006: Tag normalization

## What This Stage Is

When this stage ships, tags stop being a comma-joined string hidden
inside `entries.tags` and become a **first-class, shared taxonomy**: a
normalized `tags` table plus a polymorphic `taggings` join that any
object type can reference, with the existing `entries.tags` CSV fully
migrated into it with zero data loss. The entire v0.1 tag surface —
`brag list --tag`, `brag search`, the markdown/JSON exports, and the
summary/review/stats digests — keeps working byte-for-byte against the
new backend (no user-visible change), and on top of it the user gains a
real taxonomy surface: `brag tags` lists every tag with counts, `brag
tag rename` updates a tag everywhere at once, and `brag tag merge` folds
one tag into another. The polymorphic shape is built now (only `entry`
rows exist until STAGE-007 makes projects taggable) so the model is laid
down once rather than migrated twice.

## Why Now

**This stage is not reorderable — it is the foundation every later
PROJ-002 spec stands on.** Three reasons converge:

1. **Every later spec depends on the normalized tag schema.** STAGE-007
   (Projects) makes projects a *second taggable type*; STAGE-008 (polish)
   refreshes the docs and the digests around the new model. Both assume
   `taggings` exists. Building Projects against the old CSV-tag shape
   would mean writing project-tagging twice — once against CSV, once
   against the join after the fact — and migrating the data twice. The
   brief's Stage Plan makes the sequencing constraint load-bearing and
   explicit; this stage honors it by shipping first.

2. **DEC-004 is accepted v0.1 debt that is cheapest to repay exactly
   now.** DEC-004 (confidence 0.65) stored tags as a comma-joined `TEXT`
   column for the MVP and explicitly named two revisit triggers: "tag
   rename becomes a user ask" and "a second consumer appears, at which
   point the normalized-table option becomes more attractive." PROJ-002
   trips **both** triggers — `brag tag rename` is in scope, and projects
   are the second taggable type. Normalizing while we are already in the
   tag code is cheaper than normalizing later; the cost compounds with
   each additional CSV-tag surface we would otherwise migrate
   independently.

3. **The taxonomy surface is what makes the migration visible.** The
   schema/ETL/refactor work is invisible by design (that's the success
   bar — no user-facing change). `brag tags` / `rename` / `merge` are the
   near-free byproducts that turn a silent internal refactor into an
   upgrade the user can see and use. Shipping them in the same stage that
   builds the schema keeps the "why did we do this?" answer concrete.

No external blockers. All 14 PROJ-001 DECs apply forward unchanged; the
dev/prod DB isolation story from the brief (dev binary → `~/.bragfile-dev`,
prod stays v0.1.0 on `~/.bragfile`) governs the whole stage because this
is where the schema first changes.

## Success Criteria

- **The migration is lossless and invisible at the query surface.** The
  existing `entries.tags` CSV migrates fully into `tags` + `taggings`
  with zero data loss (every distinct trimmed token becomes a `tags`
  row; every entry's membership becomes `taggings` rows), and existing
  `brag list --tag foo` / `brag search` / export / digest output is
  byte-identical before and after on the same corpus.
- **Output shapes hold.** DEC-011's 9-key JSON array with `tags` as a
  comma-joined string, DEC-013's markdown export table, and DEC-014's
  stats "Top tags" block all render unchanged — the Store reconstructs
  `Entry.Tags` as a comma-joined projection so downstream consumers are
  untouched.
- **FTS search is preserved.** `brag search` over tags behaves
  identically (DEC-010), with `entries_fts.tags` kept synced against the
  join rather than against the (removed) `entries.tags` column.
- **Tags are a taxonomy, not per-row strings.** `brag tags` lists every
  tag across all taggable objects with counts; `brag tag rename <old>
  <new>` updates references globally; `brag tag merge <src> <dst>` folds
  one tag into another (de-duplicating memberships).
- **The model is polymorphic and forward-only.** `taggings` carries a
  `(taggable_type, taggable_id)` shape that STAGE-007 extends to
  `project` rows without a further schema change; the migration is
  forward-only (no down-migration), consistent with PROJ-001.
- **DEC-004 is formally superseded.** A new DEC (DEC-015) records the
  polymorphic schema and supersedes DEC-004; DEC-004's `superseded_by`
  is filled in. Emitted at SPEC-025 **design** time, not at framing.
- **No regressions.** `go test ./...`, `gofmt -l .`, `go vet ./...`, and
  `CGO_ENABLED=0 go build ./...` are clean through every spec; all
  STAGE-001..005 success criteria still hold.

## Scope

### In scope

- **`tags` + `taggings` schema.** A `tags(id, name UNIQUE)` table and a
  polymorphic `taggings(tag_id, taggable_type, taggable_id)` join,
  delivered as a new embedded migration (DEC-002 mechanism). Only
  `taggable_type = 'entry'` rows are created this stage.
- **ETL from the `entries.tags` CSV.** A data backfill inside the
  migration that splits each entry's comma-joined string into trimmed,
  de-duplicated tokens, upserts `tags` rows, and writes `taggings`
  memberships. Lossless and forward-only.
- **Store read/write cutover.** Writes (`Add`/`Update`) persist tags
  through the join; reads (`Get`/`List`/`Search`) reconstruct
  `Entry.Tags` as a comma-joined projection; `brag list --tag` filters
  via the join (replacing the DEC-004 sentinel-comma `LIKE`).
- **FTS5 re-sync against the join.** Move the `entries_fts.tags` sync off
  the `entries` row and onto `taggings`/`tags` mutations so search stays
  correct after a tag is added, renamed, or merged.
- **Tag taxonomy + mutation surface.** `brag tags` (taxonomy view with
  counts), `brag tag rename`, `brag tag merge` — the new user-visible
  payoff. (Confirmed in-scope at framing, 2026-06-06.)
- **DEC-015** superseding DEC-004 (emitted at SPEC-025 design).
- **Doc sweeps that fold into the originating spec** per the
  premise-audit rule: `docs/data-model.md` (the schema table, the FTS
  note, and the "Tags normalization" future-work section that this
  stage realizes), `docs/api-contract.md` (the new `tags`/`tag`
  commands), and any `docs/tutorial.md` tag mentions. Each spec
  enumerates its own greps under `## Outputs` and runs them at design.

### Explicitly out of scope

- **Projects as a taggable type.** The `taggings` schema is built
  polymorphic, but no `project` rows are written and no project command
  exists yet — that is STAGE-007. Validating the abstraction generalizes
  is a paper exercise in DEC-015, not code here.
- **Goals / any third taggable type.** Paper-sketched in DEC-015 to prove
  generality; no code (brief out-of-scope; PROJ-003 candidate).
- **Tag autocomplete.** A plausible future byproduct of the normalized
  schema, but not asked for and not in the brief's success criteria.
  Backlog if dogfooding surfaces a need.
- **Reverse-direction migration.** Forward-only, same as PROJ-001. The
  documented backup workflow (CHANGELOG + tutorial, landing in STAGE-008)
  is the downgrade story.
- **v0.2.0 release mechanics.** CHANGELOG `[0.2.0]`, the RC-tag cut, and
  the migration-prompt safety belt are STAGE-008.
- **Tag-faceted dashboards / cross-object tag views.** Enabled by this
  model but not built here.

## Spec Backlog

Two specs. The migration is a **single atomic in-place migration**
(scope decision, 2026-06-06): one `0003_*` migration creates the schema,
backfills the ETL, re-points FTS, and drops the legacy `entries.tags`
column in one forward-only transaction. The earlier expand→contract
split (a dual-written `entries.tags` shadow column carried across two
PRs) was dropped as over-engineered for this context — bragfile has
effectively one user, v0.2.0 only tags at project close, and the
downgrade story is already a `cp` backup of the SQLite file (brief). The
inter-PR "keep `main` green" safety the split was buying is therefore
worthless here. An import/export round-trip was also considered and
rejected: it would require building an import command that does not
exist and is more error-prone than a tested transactional migration. The
atomic migration reads complexity-**L**; that is accepted and justified,
**not split back**, because the single-user premise removes the value the
split provided. SPEC-026 adds the user-visible taxonomy surface on top.

Format: `- [status] SPEC-ID (cycle) — one-line summary`

- [x] SPEC-025 (shipped on 2026-06-07) — **L** *(accepted, not split)* —
      **Normalize tag storage.** One `0003_*` migration does it all in a
      single forward-only transaction: create `tags` + polymorphic
      `taggings` tables; lossless ETL from `entries.tags`; rewrite the
      `entries_fts.tags` sync off the `entries` row and onto
      `taggings`/`tags` mutations (dropping the old `entries`-row
      triggers); drop the legacy `entries.tags` column. Store write path
      splits/upserts into the join; read path reconstructs `Entry.Tags`
      via `GROUP_CONCAT` (output shapes byte-stable, DEC-011/013/014);
      `brag list --tag` filters through the join (replacing the DEC-004
      sentinel-comma `LIKE`); `brag search` byte-stable (DEC-010).
      **Emits DEC-015** (supersedes DEC-004) at design.
- [x] SPEC-026 (shipped on 2026-06-07) — **M** *(L considered, rejected)* —
      **Tag taxonomy + mutations.** `brag tags` (counted taxonomy view,
      sourced **through the `taggings` join** so it counts across all
      taggable types and omits orphans; count-DESC/name-ASC),
      `brag tag rename <old> <new>` (single `UPDATE`, rides `tags_au` →
      FTS; **errors** into an existing name, no auto-merge),
      `brag tag merge <src> <dst>` (de-dupes via **DELETE+INSERT** on
      `taggings` — fires the existing `taggings_ad`/`_ai`, so FTS stays
      correct with **no `taggings_au` / no `0004_*` migration**).
      Orphan-tag GC: **none** — orphans are invisible (the join) and
      reused by `INSERT OR IGNORE`, so cleanup is deferred. **Emits
      DEC-016** (mutation semantics) at design. Design done 2026-06-07;
      held for review before the build session.

**Count:** 2 shipped / 0 active / 0 pending — **stage complete, ready to close**

**Complexity check:** 1 × L (accepted) + 1 × M = 2 specs — under the
brief's ~3–4 estimate, but coherent: the L is a single load-bearing
migration that is cleaner atomic than split for a solo tool, and the
taxonomy work is genuinely separable. The fallback, only if SPEC-025's
design truly cannot hold together as one spec, is the expand→contract
split (back to 3); the default is one atomic migration.

## Design Notes

Glue and cross-cutting direction. The weighty decision (the polymorphic
schema) gets its own DEC-015 at SPEC-025 design — **not written at
framing** per the frame-cycle rule.

### The DEC-015 supersession (emit at SPEC-025 design, not now)

- **DEC-015 supersedes DEC-004.** It records the `tags` + polymorphic
  `taggings` schema, the ETL contract, the `Entry.Tags`-as-projection
  rule that preserves output shapes, and a paper sketch of a third
  taggable type to prove the abstraction generalizes (per brief reason
  4). At emit time, fill DEC-004's `superseded_by: DEC-015` and DEC-015's
  `supersedes: DEC-004`. Target confidence ≥ 0.7 (the normalization is
  well-understood and the brief commits to it); if any spec-level choice
  lands < 0.8, open a question in `/guidance/questions.yaml` per §14.
- **The supersession is DEC-004 firing as designed, not a reversal.**
  DEC-004's own Validation block names the revisit trigger "tag rename
  becomes a user ask" — and `brag tag rename` is a STAGE-006 success
  criterion. DEC-004 also predicted "the normalized-table option becomes
  more attractive once a second consumer appears"; projects are that
  consumer. So DEC-015 is the planned, conditional successor the original
  decision anticipated, not a correction of a wrong call. The DEC-015
  Context section should say this explicitly.

### Migration mechanics

- **Forward-only.** Same constraint as PROJ-001 (DEC-002: no down
  migrations). The downgrade story is the documented SQLite-file backup,
  spelled out in STAGE-008's CHANGELOG/tutorial work.
- **One atomic in-place migration (scope decision, 2026-06-06).** The
  earlier expand→contract split (a dual-written `entries.tags` shadow
  column carried across two PRs) was dropped as over-engineered: bragfile
  has effectively one user, v0.2.0 tags only at project close, and the
  backup story is already a `cp` of the SQLite file — so the inter-PR
  safety the split bought is worthless here. A hand-rolled import/export
  round-trip was also considered and rejected (more error-prone than a
  tested transactional migration, and it would require building an import
  command that does not exist). So `0003_*` does everything in one
  transaction: create `tags` + `taggings`, backfill the ETL, rewrite the
  FTS triggers, drop `entries.tags`. The spec is complexity-L and that is
  accepted, not split.
- **Output-shape stability is the mechanism that makes "invisible"
  true.** DEC-011 §3 locks `tags` as a comma-joined string in JSON;
  DEC-013 and DEC-014 render it similarly. The Store therefore keeps
  `Entry.Tags` as a comma-joined **projection** reconstructed from the
  join on read (`GROUP_CONCAT`, deterministic order — likely
  insertion/`name ASC`). Consumers (`internal/export`, the
  `extractTags` splitter in `stats.go`, the editor round-trip) stay
  unchanged. This is what lets "refactor every tag-touching command to
  read through the join" cost almost nothing at the call sites — the
  refactor lives inside the Store.
- **FTS topology moves, FTS shape does not — in the same migration.**
  `entries_fts.tags` stays a denormalized indexed column (DEC-010 search
  behavior unchanged). What changes is *what fires the sync*: today three
  triggers on `entries` copy `new.tags`; after normalization, tag
  membership changes happen via `taggings` INSERT/DELETE and tag renames
  via `tags` UPDATE, none of which touch the `entries` row — so the
  triggers must move onto those tables and recompute the affected entry's
  `GROUP_CONCAT`, and the old `entries`-row triggers must be dropped
  before `entries.tags` is dropped. This is the single highest-risk
  change in the stage; because the migration is atomic it lands inside
  SPEC-025 and demands that spec's most thorough test coverage (FTS
  add/rename/merge round-trips against a representative corpus).

### Premise-audit triggers (every migration spec must reconcile these)

- **Migration count-bump (the additive case).** Adding `0003_*` breaks
  the literal-count assertions in **two** test files (four assertion
  sites) that hard-code `schema_migrations count = 2` and the exact list
  `["0001_initial","0002_add_fts"]`:
  `internal/storage/store_test.go` and `internal/storage/fts_test.go`.
  SPEC-025 ground-truthed this at design (audit-grep cross-check, §9):
  `internal/storage/migrate_test.go`'s `count == 2` assertions run
  against in-test `fstest.MapFS` fixtures (`0001_a`/`0002_b`), **not** the
  real `migrationsFS`, so they stay untouched — a correction to this
  stage's original "≈five sites across three files" estimate. SPEC-025
  enumerates the count-bump (2 → 3) under `## Outputs`.
- **Status-change (the doc case).** `docs/data-model.md` documents tags
  as comma-joined (DEC-004), the FTS tokenizer note, the "No index on
  `tags`" line, and a "Tags normalization" future-work section that this
  stage *realizes* — all become stale. `docs/api-contract.md` gains the
  new commands. Run `grep -rn -i "tag" docs/ README.md` per the
  status-change rule and enumerate each hit as "updates" or "stays."
- **Inversion/removal (the test case).** Replacing the sentinel-comma
  `LIKE` with a join changes the premise of
  `TestList_FilterByTag` / `TestList_TagFilterNoFalsePositive` /
  `TestList_TagFilterNullAndEmpty` (`store_test.go`) and the FTS
  tags-column-shape assertion in `fts_test.go` — list these under
  `## Outputs` as planned rewrites, not build-time discoveries.

### Premise-audit sub-template — EXTRACTED at this framing

The premise-audit sub-template extraction, deferred 3× (SPEC-015 ship,
STAGE-004 ship, STAGE-005 framing) with the standing instruction to
decide at STAGE-006 framing, is **extracted now** (decision 2026-06-06).
Rationale: STAGE-006 is the first stage where the trigger fires
concretely and repeatedly — SPEC-025's migration hits the count-bump
across two test files, carries a status-change doc sweep, AND carries
an inversion/removal (the `list --tag` join), and the value compounds
across STAGE-006/007/008 (the projects schema migration in STAGE-007
hits the same count-bump; ~9–11 specs across the project all sit in the
audit family). A fourth "wait
for more volume" deferral would be the exact over-deferral the §12
codification meta-rule cautions against, with no new information to gain.
The sub-template lives at `projects/_templates/premise-audit.md`; each
spec references it from `## Outputs` instead of re-deriving the three
cases + the cross-check. This is a templates artifact, not a DEC.

### WATCH-list items carried from PROJ-001 close

STAGE-006 specs are the natural next test cases; none is codified yet
(do not codify mid-stage without the documented trigger):

- **§12(a) — run embedded test assertions against embedded literals at
  design time.** N=2 within SPEC-023; awaits a third confirming case
  across a distinct spec. SPEC-025's `0003_*` migration SQL is an
  embedded literal whose assertions (row counts post-ETL) should be run
  at design — a likely third case.
- **Trust-but-verify agent push reports.** N=2 within SPEC-023; a
  2-second `git ls-remote origin <branch>` after any "pushed" claim.
  Coordinator-discipline reflex, mechanically adjacent to §10/§13.
- **§13 fresh-session working-tree preservation.** N=1 from SPEC-024
  build; note for cross-reference, not promotable yet.

## Dependencies

### Depends on

- **PROJ-001 (shipped 2026-05-17).** All 14 DECs apply forward; DEC-004
  is superseded by DEC-015 at SPEC-025 design (not a removal — the
  decision firing as designed). Mechanism DECs that govern this stage
  directly: DEC-002 (embedded forward-only migrations), DEC-010 (search
  syntax — must stay byte-stable), DEC-011 / DEC-013 / DEC-014 (output
  shapes — must stay byte-stable). AGENTS.md conventions all apply: §2 ID
  numbering (this stage's specs are SPEC-025..026; the new DEC is
  DEC-015), §9 audit family, §10 push-discipline, §12 literal-artifact +
  §12(b) design-time pre-flight + codification meta-rule, §13
  fresh-session rule.
- **PROJ-002 brief (dev/prod DB isolation).** This is the stage where the
  schema first changes, so the dev binary must run against
  `~/.bragfile-dev` and never open the prod v0.1.0 DB in v0.2.x format.
- **External: none new.** No new top-level Go dependency is anticipated
  (the join + `GROUP_CONCAT` are plain `database/sql`); any dep would
  need its own DEC per the `no-new-top-level-deps-without-decision`
  constraint.

### Enables

- **STAGE-007 (Projects, core).** Makes `taggings` carry `project` rows
  with no further schema change; project CRUD and `brag project` build on
  a tag model that already exists.
- **STAGE-008 (polish + integration).** The doc refresh, CHANGELOG
  `[0.2.0]`, and migration-prompt safety belt all assume the normalized
  schema is in place.
- **PROJ-003 (if framed).** The polymorphic abstraction validated here is
  what a third taggable type (`goals`) would build on without
  re-litigating the tag schema.

## Stage-Level Reflection

*Filled in when status moves to shipped. Run Prompt 1c (Stage Ship) in
FIRST_SESSION_PROMPTS.md to draft this.*

- **Did we deliver the outcome in "What This Stage Is"?** <yes/no + notes>
- **How many specs did it actually take?** <number vs. plan>
- **What changed between starting and shipping?** <one sentence>
- **Lessons that should update AGENTS.md, templates, or constraints?**
  - <one-line updates>
- **Should any spec-level reflections be promoted to stage-level lessons?**
  - <one-line items>
