---
# Maps to ContextCore epic-level conventions.
# A Stage is a coherent chunk of work within a Project.
# It has a spec backlog and ships as a unit when the backlog is done.

stage:
  id: STAGE-007                     # stable, zero-padded, repo-global (never reused)
  status: active                    # proposed | active | shipped | cancelled | on_hold
  priority: high                    # critical | high | medium | low
  target_complete: 2026-06-20       # largest stage of PROJ-002; leaves buffer before project ship 2026-06-26

project:
  id: PROJ-002                      # parent project
repo:
  id: bragfile

created_at: 2026-06-08
shipped_at: null
---

# STAGE-007: Projects (core)

## What This Stage Is

When this stage ships, **projects become a first-class entity** in
bragfile: a named thing with one or more filesystem locations, a status,
and a state/next-action note. The tool can finally answer the brief's
headline question — *"what am I working on, and where does it live on
this machine?"* `brag project new <name> --path <dir>` registers a
project; `brag project list` / `show` read it; `brag project edit` /
`archive` / `delete` mutate it; `brag project status` is the scannable
dashboard (active projects by recency, each with its state note and a
recent-brag count); `brag project here`, run from inside a registered
directory, surfaces the matching project; and `brag add`, run from inside
a registered project's directory, auto-fills `--project` from the cwd so
capture knows where it is. The polymorphic `taggings` shape laid down in
STAGE-006 is confirmed by this stage — projects become a *second
taggable type* with **no schema change** — though whether the
`brag project tag` command surface ships here or is deferred is a
flagged design question, not a foregone scope item (default: schema-ready
only). The single load-bearing design question the stage must resolve is
**how the existing free-text `entries.project` column relates to the new
`projects` entity** (DEC-017); everything else builds on that answer.

## Why Now

STAGE-006 (tag normalization) shipped 2026-06-07, six days early, and
its reflection carries directly into this stage. Three reasons converge:

1. **The substrate is live and unblocks projects with zero schema debt.**
   STAGE-006 made `taggings` polymorphic on purpose so that "STAGE-007
   adds `taggable_type='project'` rows with no further schema change" —
   that prediction is now testable, and projects are the second taggable
   type DEC-015 was designed around. Building projects now validates the
   abstraction the prior stage paid for.

2. **Projects are the brief's headline feature.** PROJ-002 is named
   `projects-and-tags`; tags shipped first only because every later spec
   depended on the normalized schema (the load-bearing sequencing
   constraint). With that foundation in place, projects are the next and
   largest chunk — the user-visible payoff that answers the recurring
   real pain named in the brief ("the user forgets where projects live").

3. **It is the only thing standing between here and STAGE-008 polish.**
   STAGE-008 (docs refresh, CHANGELOG `[0.2.0]`, the migration-prompt
   safety belt, v0.2.0 cut) assumes both tags *and* projects exist. This
   stage is the last feature work in PROJ-002.

No external blockers. All 14 PROJ-001 DECs plus DEC-015 / DEC-016 apply
forward unchanged. This is the **second** schema change in PROJ-002, so
the brief's dev/prod DB isolation story still governs: the dev binary
runs against `~/.bragfile-dev` and must never open the prod v0.1.0 DB in
v0.2.x format.

## Success Criteria

- **Projects are first-class.** `brag project new <name> --path <dir>`
  registers a project with a location, a status, and a state note;
  `brag project here`, run from inside that directory, surfaces the
  matching project.
- **Full CRUD.** `new` / `list` / `show` / `edit` / `archive` / `delete`
  all ship. `archive` is a non-destructive status flip; `delete` is
  destructive and its blast radius (locations, any `'project'` taggings,
  and the `entries.project` relationship per DEC-017) is defined, tested,
  and consciously chosen — not incidental.
- **Projects are scannable.** `brag project status` lists active projects
  sorted by recency, each showing its state/next-action note and a
  recent-brag count.
- **Capture knows where it is.** `brag add`, run from inside a registered
  project's directory with no explicit `--project`, auto-fills
  `--project` from the cwd.
- **The `entries.project` relationship is resolved and lossless.** DEC-017
  records how the free-text `entries.project` column relates to the new
  entity; existing entries are handled with **zero data loss** (no entry
  loses its project string), and existing `brag list --project foo` /
  the project-grouped digests (`summary` / `review` / `stats`) keep
  working with no user-visible regression.
- **The polymorphic schema generalizes as predicted.** `taggings` carries
  (or is provably able to carry) `taggable_type='project'` rows with no
  schema change; `brag tags` counts them polymorphically with no code
  change the moment they exist (DEC-015 / DEC-016 validated forward).
- **No regressions.** `go test ./...`, `gofmt -l .`, `go vet ./...`, and
  `CGO_ENABLED=0 go build ./...` are clean through every spec; all
  STAGE-001..006 success criteria still hold; output shapes
  (DEC-011/013/014) and search (DEC-010) stay byte-stable.

## Scope

### In scope

- **Projects schema.** A `projects` table (DEC-005 INTEGER autoincrement
  PK; `name`, `status`, a state/next-action note column, timestamps) and
  a `project_locations` join supporting **one project, many directories**.
  Delivered as a new embedded `0004_*` migration (DEC-002 mechanism,
  forward-only). The status enum and state-note column are laid down in
  this first migration even though the dashboard that renders them comes
  later — "lay the schema down once," mirroring STAGE-006.
- **The `entries.project` ↔ `projects` relationship (DEC-017).** The
  central design question (see Design Notes); resolved by the first spec,
  with whatever `0004_*` backfill/link the chosen model requires, lossless
  and forward-only.
- **Full project CRUD.** Store-layer typed methods + `brag project`
  command surface: `new` / `list` / `show` / `edit` / `archive` /
  `delete`, with plain and `--format json` output per the DEC-011/013/014
  output-shape family.
- **`brag project status` dashboard.** Active projects by recency, each
  with state note + recent-brag count (the "scannable" criterion).
- **`brag project here` cwd auto-detect.** Resolve the cwd against
  `project_locations` (resolution policy is a flagged design question:
  exact dir vs nearest-ancestor vs longest-prefix).
- **`brag add` `--project` auto-fill from cwd.** When `brag add` runs
  with no explicit `--project` inside a registered project location,
  auto-fill it. Reuses the `here` resolver and must agree with DEC-017's
  write-path semantics.
- **Per-spec doc updates that fold into the originating spec** per the
  premise-audit rule: `docs/data-model.md` (the new tables), targeted
  `docs/api-contract.md` entries for each new command. Each spec
  enumerates its own greps under `## Outputs` and runs them at design.

### Explicitly out of scope

- **The `brag project tag` command surface (flagged, default deferred).**
  The schema is *ready* for `taggable_type='project'` per DEC-015, and
  this stage may write `'project'` taggings if a spec needs them, but a
  dedicated tag-management command for projects is **not** a stage
  success criterion and is not in the brief's STAGE-007 line. Default:
  schema-ready only; revisit for STAGE-008 or PROJ-003. (Surfaced in
  Design Notes for the first spec to confirm.)
- **v0.2.0 release mechanics.** CHANGELOG `[0.2.0]`, the RC-tag cut, the
  migration-prompt safety belt — all STAGE-008.
- **Comprehensive doc sweep.** The full `docs/tutorial.md` projects+tags
  walkthrough and the `docs/architecture.md` diagram/responsibilities
  refresh are STAGE-008; only per-spec, premise-audit-driven doc updates
  fold in here.
- **Orphan-tag garbage collection.** Stays deferred (DEC-016 choice 4) —
  unchanged by this stage.
- **Goals / any third taggable type.** Paper-sketched in DEC-015; no code
  (PROJ-003 candidate).
- **Reverse-direction migration.** Forward-only, same as PROJ-001/006.
  The downgrade story is the documented SQLite-file backup (STAGE-008).
- **Tag autocomplete; per-type tag-count breakdown in `brag tags`.**
  STAGE-006-surfaced backlog candidates; stay backlogged (the per-type
  breakdown only becomes meaningful once projects are actually tagged).

## Spec Backlog

Six specs, deliberately kept S/M with the schema/migration foundation
split from the CLI surface (split preference, 2026-06-08). The one
L-risk spec is **SPEC-027** (schema + `0004_*` migration + the DEC-017
relationship decision + Store read primitives + the count-bump premise
audit); it is held to **M** by splitting the Store *mutation* methods
out into SPEC-029 and the location *resolver* out into SPEC-031. If
SPEC-027's design genuinely cannot hold together at M — most likely if
DEC-017 requires a non-trivial `entries.project` backfill — split that
backfill into its own spec rather than letting SPEC-027 grow to L.

Format: `- [status] SPEC-ID (cycle) — one-line summary`

- [x] SPEC-027 (shipped on 2026-06-08) — **M** *(L-risk; split-watch did
      NOT fire — soft match needs no backfill)* — **Projects schema + `0004_*` migration +
      DEC-017.** `projects` (id PK, name, status enum, state_note,
      timestamps) + `project_locations` (project_id FK, path UNIQUE)
      tables; forward-only `0004_add_projects.sql`. **DEC-017 emitted:**
      `entries.project` ↔ `projects` is **soft string match** (free text,
      join on `projects.name`, zero backfill) — the choice that holds the
      spec at M; status enum `active|paused|done|archived` (default
      `active`, Store-validated, no DB CHECK) + single free-text
      `state_note` ride in DEC-017 (state-note shape filed as
      `project-state-note-shape`, < 0.8). Store read primitives:
      `CreateProject` / `GetProject` / `ListProjects` / `AddLocation`.
      Premise audit run at design: count-bump 3→4 (four sites confirmed:
      `store_test.go:172,206-208`; `fts_test.go:149,266-270`;
      `migrate_test.go` MapFS stays); inversion → zero rewrites
      (soft match preserves every `entries.project` premise); §12(b)
      migration pre-flighted against modernc.org/sqlite. Mutations →
      SPEC-029, `here` resolver → SPEC-031.
- [x] SPEC-028 (shipped on 2026-06-08) — **M** — **`brag project new` / `list` /
      `show`.** Read+create CLI on the SPEC-027 primitives; `brag project`
      parent (no RunE) mirroring SPEC-026's `brag tag`; plain and
      `--format json` output (DEC-011 array for `list`, single object for
      `show`; locations as a JSON array). Design locked four CLI
      decisions (**no DEC**, all ≥0.8): `--path` required (LD1);
      `show <name|id>` resolves name-first then id-fallback + a small
      `GetProjectByName` Store helper (LD2); `new` path pre-check via
      `ListProjects` prevents an orphan project on a path conflict (LD3);
      `--format` default `""` stated explicitly per the STAGE-006
      flag-default WATCH item (LD4). Premise audit run at design:
      inversion NONE, count-bump NONE (no migration), status-change →
      `docs/api-contract.md` only (tutorial/architecture deferred to
      STAGE-008). Registration gap noted (no test enumerates the root
      set; build confirms `brag project --help` in the real binary).
- [ ] SPEC-029 (not yet written) — **M** — **`brag project edit` /
      `archive` / `delete`.** Mutation CLI + Store `UpdateProject` /
      `ArchiveProject` (status flip) / `DeleteProject` (destructive;
      blast radius defined against DEC-017 + locations + any `'project'`
      taggings).
- [ ] SPEC-030 (not yet written) — **M** — **status + state-note model +
      `brag project status` dashboard.** The "scannable" criterion:
      active projects by recency, state note, recent-brag count (the
      count depends on the DEC-017 linkage).
- [ ] SPEC-031 (not yet written) — **S/M** — **`brag project here`
      cwd auto-detect.** Resolve cwd against `project_locations`
      (resolution policy decided at design: exact / nearest-ancestor /
      longest-prefix). Provides the resolver SPEC-032 reuses.
- [ ] SPEC-032 (not yet written) — **M** — **`brag add` `--project`
      auto-fill from cwd.** When no explicit `--project` and cwd is
      inside a registered location, auto-fill it via the SPEC-031
      resolver; write-path must agree with DEC-017.

**Count:** 2 shipped / 0 active / 4 pending

**Complexity check:** 6 specs, all S/M by construction (no L). The split
preference traded STAGE-006's one-atomic-L approach for a foundation spec
(SPEC-027) plus four CLI specs plus the resolver — coherent because the
CLI surfaces are genuinely separable and the schema is the only shared
dependency. Six sits at the top of the brief's ~5–6 estimate, not over
it. The single watch item is SPEC-027: if the `entries.project` backfill
DEC-017 mandates is non-trivial, peel it into a seventh spec rather than
let SPEC-027 cross into L.

## Design Notes

Glue and cross-cutting direction. The weighty decision (the
`entries.project` ↔ `projects` relationship) gets its own **DEC-017** at
SPEC-027 design — **not written at framing** per the frame-cycle rule.
The five load-bearing design questions below are **surfaced for
spec-design time, not decided here.** The first specs resolve them.

### SURFACED design questions (resolve at spec design; do not decide now)

1. **THE central one — how does free-text `entries.project` relate to the
   new `projects` entity? (DEC-017, SPEC-027.)** Three candidate shapes,
   none chosen at framing:
   - **Hard FK** — `entries.project_id REFERENCES projects(id)`; every
     entry's project must be a registered project. Cleanest queries;
     forces a backfill decision for existing entries whose string matches
     no registered project (create-on-migrate? leave null? a synthetic
     "unfiled" project?).
   - **Soft string match** — keep `entries.project` as free text; join to
     `projects.name` opportunistically. Zero migration risk; no
     referential integrity; rename/merge of a project name silently
     desyncs entries.
   - **Free-text with optional link** — keep the string, add a nullable
     `project_id` that links when a match exists. Backwards-compatible;
     two sources of truth to keep consistent.
   The answer determines **whether `0004_*` backfills/links existing
   entries**, what `brag project delete` does to entries that reference
   it, and how `brag project status`'s recent-brag count is computed.
   This is the spec-027 split-watch trigger (see Spec Backlog).

2. **`project_locations` multi-directory model + `here` resolution.**
   Schema is clear (one project → many directory rows). The open
   question is how `brag project here` / `brag add` resolve a cwd:
   exact-directory match only, **nearest-ancestor** (cwd is *inside* a
   registered dir), or **longest-prefix** when multiple registered dirs
   are ancestors of the cwd. Pick one consciously; the resolution policy
   is the whole UX of "capture knows where it is." (SPEC-031, reused by
   SPEC-032.)

3. **The status enum values + the state/next-action note model.** What
   are the allowed `status` values (e.g. `active` / `paused` /
   `archived` / `done`?) and is the state note a single free-text column,
   a separate "next action" field, or both? The dashboard
   (`brag project status`) renders whatever this decides. Lay the columns
   down in `0004_*` (SPEC-027) even though SPEC-030 renders them.

4. **Are projects taggable IN STAGE-007 — command surface, or
   schema-only?** The schema is ready (DEC-015) with no change needed.
   Default at framing: **schema-ready only**, no `brag project tag`
   command (it's not a stage success criterion). The first spec touching
   project mutation should confirm this default or pull the tagging
   surface in explicitly. If `'project'` taggings *are* written, see
   design question 6.

5. **`0004_*` migration — premise audit + §12(a) expected-literals.**
   The new migration trips the additive count-bump case at the **same
   `schema_migrations` sites STAGE-006 exercised, now 3→4** (grounded at
   framing, see below). Per §12(a), the test's *expected-value literals*
   (the sorted `want` list and the `count == N`) must be computed/run
   against the real migration at design, not typed by hand.

6. **Position base when writing `'project'` taggings (carry-forward).**
   Confirmed at framing: the `0003` ETL numbers `taggings.position`
   **1-based** (recursive-split anchor at `pos=0` carries an empty token;
   the first real token lands at `pos=1`), while `insertTaggings`
   (Add/Update) is **0-based** (`for i := range tokens`). Migrated entry
   rows and freshly-added entry rows therefore *already* differ in
   position base. If STAGE-007 writes any `'project'` taggings, pick a
   base consciously so migrated and freshly-added rows are
   indistinguishable — and decide whether to also normalize the existing
   entry-row inconsistency (likely out of scope; flag, don't silently
   inherit). Carry-forward from STAGE-006.

### Premise-audit triggers (every migration/command spec must reconcile)

Reference `projects/_templates/premise-audit.md`; each spec runs its own
greps at design and enumerates hits under `## Outputs`.

- **Count-bump (additive) — `0004_*`, grounded at framing.** Adding the
  fourth migration breaks the literal-count assertions in **two test
  files, four sites**, all running against the real `migrationsFS`:
  - `internal/storage/store_test.go:172` (`want := []string{"0001_initial",
    "0002_add_fts", "0003_normalize_tags"}`) and `:206-208`
    (`count … want 3`).
  - `internal/storage/fts_test.go:149` (same `want` list) and `:266-270`
    (`count … want 3`).
  `internal/storage/migrate_test.go` runs against in-test `fstest.MapFS`
  fixtures (`count == 2`), **not** the real FS, so it stays untouched —
  the same correction STAGE-006 verified, restated here so SPEC-027 does
  not re-estimate it. §12(a): the `want` lists are lexically ordered, so
  `"0004_<slug>"` appends last — confirm the literal slug at design.
- **Status-change (doc).** The new `brag project*` commands and the new
  tables change shipping status. Run `grep -rn -i "project" docs/
  README.md` and enumerate each hit as "updates" or "stays." The
  comprehensive tutorial/architecture refresh is STAGE-008; only the
  per-spec `data-model.md` / `api-contract.md` updates fold in here.
- **Inversion/removal.** If DEC-017 changes `entries.project` semantics
  (e.g. free-text → linked), the premise of existing tests around
  `ListFilter.Project` (`e.project = ?` in `store.go`), `brag add -p`,
  and the project-grouped digests may be invalidated. SPEC-027 must
  enumerate these as planned rewrites, not build-time discoveries.

### Carry-forwards from STAGE-006 close (WATCH-list)

STAGE-007 specs are the natural next test cases; none is codified yet (do
not codify mid-stage without the documented trigger):

- **Trust-but-verify agent push reports** — N=2 (SPEC-023); applied as a
  coordinator reflex throughout STAGE-006 (every "pushed" claim checked
  via `git ls-remote` / PR state). Carry forward.
- **§13 fresh-session working-tree preservation** — N=1 (SPEC-024). Carry
  forward.
- **Contamination-heuristic exception for literal-artifact builds** —
  N=1 (SPEC-026). Discriminator to preserve: "'nothing was unclear' is
  the expected honest output of a literal-artifact-as-spec build;
  distinguish honest-frictionless from mailed-in by whether the
  reflection surfaces ANY specific observation."
- **Flag-default explicitness in literal-artifact CLI specs** — N=1,
  zero-cost (SPEC-026 `--format ""` default). STAGE-007's CRUD CLI specs
  (each a cobra `Long` + flags) are the natural N=2 test case: state each
  embedded flag's default, not just its accepted values. Fold a one-liner
  into the literal-artifact-as-spec guidance if a second case lands.

### Bookkeeping (numbering)

- Next free DEC id at framing is **DEC-017**, which SPEC-027 consumes for
  the `entries.project` relationship. The brief's STAGE-008 line
  previously pinned a concrete id for its hypothetical migration-prompt
  DEC, which went stale twice (DEC-016→017 at STAGE-006 close, then
  017→018 here). **Fixed robustly at this frame commit:** the brief no
  longer pins a number for a not-yet-emitted DEC — its id is assigned at
  design time. (STAGE-007 may itself emit more than one DEC; the
  resolution-policy and status-enum questions could each warrant one.)
- STAGE-007 specs are **SPEC-027..032** (SPEC-026 was the last shipped).

### Already-true, no work needed

- **`brag tags` already counts `'project'` taggings polymorphically.**
  Its `TagCounts()` query (`store.go`) groups across all `taggable_type`
  with no per-type filter, so the moment any `'project'` tagging exists
  it is counted automatically — no change required in this stage.

## Dependencies

### Depends on

- **STAGE-006 (shipped 2026-06-07).** The polymorphic `tags` + `taggings`
  schema is live; projects become a second taggable type with no schema
  change (DEC-015). DEC-016 (tag mutation semantics) applies unchanged.
- **PROJ-001 (shipped 2026-05-17).** All 14 DECs apply forward. Mechanism
  DECs governing this stage directly: DEC-002 (embedded forward-only
  migrations — the `0004_*` mechanism), DEC-005 (INTEGER autoincrement
  PKs — `projects.id`, `project_locations.id`), DEC-006 / DEC-007 (cobra
  framework + required-flag validation in `RunE` — the `brag project`
  command surface), DEC-011 / DEC-013 / DEC-014 (output shapes — the
  project list/show/status renderers reuse the JSON/markdown envelope
  family), DEC-010 (search — must stay byte-stable). AGENTS.md
  conventions all apply: §2 ID numbering, §9 audit family, §10
  push-discipline, §12 literal-artifact + §12(a)/(b) design-time
  pre-flight + codification meta-rule, §13 fresh-session rule.
- **PROJ-002 brief (dev/prod DB isolation).** This is the **second**
  schema change in PROJ-002, so the dev binary must run against
  `~/.bragfile-dev` and never open the prod v0.1.0 DB in v0.2.x format.
- **External: none new.** `projects` + `project_locations` are plain
  `database/sql`; cwd resolution is stdlib (`os.Getwd`, `path/filepath`).
  Any new top-level dependency would need its own DEC per the
  `no-new-top-level-deps-without-decision` constraint.

### Enables

- **STAGE-008 (polish + integration).** The full tutorial/api-contract/
  architecture doc sweep, CHANGELOG `[0.2.0]`, the migration-prompt
  safety belt (possible **DEC-018**), and the v0.2.0 cut all assume both
  tags and projects exist.
- **PROJ-003 (if framed).** A third taggable type (`goals`) joins the
  validated polymorphic model with no schema re-litigation; richer
  cross-object views (project-scoped digests, tag-faceted dashboards)
  layer on first-class projects.

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
