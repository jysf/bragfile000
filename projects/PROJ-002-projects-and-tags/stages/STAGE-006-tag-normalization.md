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

Three specs. The migration is framed as an **expand→contract pair**
(confirmed at framing, 2026-06-06) so `main` is never left in a
half-migrated state between PRs and the load-bearing migration does not
become a single complexity-L spec. SPEC-025 expands (adds the new
schema, backfills, dual-writes, moves reads); SPEC-026 contracts (moves
FTS onto the join, drops the shadow column); SPEC-027 adds the
user-visible taxonomy surface on top.

Format: `- [status] SPEC-ID (cycle) — one-line summary`

- [ ] SPEC-025 (design) — **M** — **Normalize tag storage (expand).**
      New `0003_*` migration: `tags` + polymorphic `taggings` tables +
      lossless ETL from `entries.tags`. Store write path splits/upserts
      into the join and dual-writes the legacy `entries.tags` column (so
      FTS keeps working untouched this spec); read path reconstructs
      `Entry.Tags` via `GROUP_CONCAT`; `brag list --tag` filters through
      the join. **Emits DEC-015** (supersedes DEC-004) at design. *Watch
      for L:* if dual-write + read cutover + `list --tag` balloon at
      design, peel the `list --tag` join migration into its own spec.
- [ ] SPEC-026 (not yet written) — **M** — **FTS re-sync + contract.**
      Rewrite the `entries_fts.tags` trigger topology so it tracks
      `taggings`/`tags` mutations instead of `entries`-row writes; drop
      the legacy `entries.tags` shadow column in a `0004_*` migration.
      `brag search` byte-stable (DEC-010). This is the trickiest
      correctness surface and is isolated deliberately.
- [ ] SPEC-027 (not yet written) — **M** — **Tag taxonomy + mutations.**
      `brag tags` (taxonomy view with counts, sourced from the `tags`
      table), `brag tag rename <old> <new>`, `brag tag merge <src> <dst>`
      (de-dupes memberships; rename-into-existing semantics decided at
      design). Stats/digest top-tags may read counts directly from `tags`
      where cleaner.

**Count:** 0 shipped / 0 active / 3 pending

**Complexity check:** 3 × M, no L — within the brief's ~3–4 estimate and
the 3–8 healthy band, so **no rescope**. The expand→contract split is
what keeps the migration off the L line; if SPEC-025 still reads L at
design, its `list --tag` join cutover is the clean peel-off (→ 4 specs,
still healthy).

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
- **FTS topology moves, FTS shape does not.** `entries_fts.tags` stays a
  denormalized indexed column (DEC-010 search behavior unchanged). What
  changes is *what fires the sync*: today three triggers on `entries`
  copy `new.tags`; after normalization, tag membership changes happen via
  `taggings` INSERT/DELETE and tag renames via `tags` UPDATE, none of
  which touch the `entries` row — so the triggers must move onto those
  tables and recompute the affected entry's `GROUP_CONCAT`. SPEC-026
  owns this; it is the single highest-risk change in the stage, which is
  why it is its own spec rather than bundled into SPEC-025.
- **Expand→contract keeps main green.** SPEC-025 retains `entries.tags`
  as a dual-written shadow so the existing FTS triggers keep working
  unchanged while reads/`list --tag` move to the join. SPEC-026 then
  re-points FTS at the join and drops the shadow. Neither PR leaves
  `main` with broken search or a half-migrated read path.

### Premise-audit triggers (every migration spec must reconcile these)

- **Migration count-bump (the additive case).** Adding `0003_*` (and
  `0004_*` in SPEC-026) breaks the literal-count assertions in **five**
  test files that hard-code `schema_migrations count = 2` and the exact
  list `["0001_initial","0002_add_fts"]`:
  `internal/storage/store_test.go`, `internal/storage/fts_test.go`,
  `internal/storage/migrate_test.go`. Each migration spec enumerates the
  count-bump under `## Outputs` and runs `grep -rn "schema_migrations"
  internal/**/*_test.go` plus `grep -rn "0002_add_fts" internal` at
  design (audit-grep cross-check, §9).
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
concretely and repeatedly — both migration specs hit the count-bump
across five test files, both carry a status-change doc sweep, and
SPEC-025 carries an inversion/removal — and the value compounds across
STAGE-006/007/008 (10–13 specs all in the audit family). A fourth "wait
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
  across a distinct spec. The `0003_*`/`0004_*` migration SQL is an
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
  numbering (this stage's specs are SPEC-025..027; the new DEC is
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
