---
# Maps to ContextCore insight.* semantic conventions.

insight:
  id: DEC-017
  type: decision
  confidence: 0.80
  audience:
    - developer
    - agent

agent:
  id: claude-opus-4-8
  session_id: null

# Decisions are repo-level, but it's useful to track which project
# caused them to be emitted.
project:
  id: PROJ-002
repo:
  id: bragfile

created_at: 2026-06-08
supersedes: null
superseded_by: null

tags:
  - data-model
  - projects
  - migrations
  - entries
---

# DEC-017: `entries.project` relates to `projects` by soft string match

## Decision

The existing free-text `entries.project` column **stays free text** and
is **not** linked to the new `projects` entity by a foreign key or a link
column. The relationship is an **opportunistic join on `projects.name`**
computed at query time. Concretely:

- `entries.project` remains `TEXT` (nullable), written and read exactly as
  it is today. `ListFilter.Project` keeps its `e.project = ?` equality.
- The `0004_add_projects.sql` migration **adds two tables and backfills
  nothing** — it never reads or rewrites a single `entries` row.
- Where a feature needs "brags belonging to this project" (e.g.
  `brag project status`'s recent-brag count, SPEC-030), it computes
  `entries.project = projects.name` — a join on the string, not a key.

Two schema-shape sub-decisions ride with this DEC (STAGE-007 design
question #3), laid down in the same `0004_*` migration ("schema down
once," mirroring STAGE-006):

- **`projects.status` is a four-value enum — `active` | `paused` | `done`
  | `archived` — defaulting to `active`, validated in the Store, NOT a DB
  `CHECK`.** `active` is the working default; `archived` is the
  non-destructive flip `brag project archive` sets (SPEC-029); `done`
  marks completion; `paused` is set-aside-may-return. Enforcing the enum
  in Go rather than a `CHECK` mirrors how `entries.type` is a free-text
  column, and — under forward-only migrations (DEC-002) — keeps "add a
  status value later" an additive Store change instead of a SQLite table
  rebuild (you cannot alter a `CHECK` in place).
- **The state/next-action note is a single free-text `state_note TEXT NOT
  NULL DEFAULT ''` column**, not a split `state` + `next_action` pair.
  One note the user writes freely ("Shipped tags; next: cut v0.2.0") is
  the minimal primitive; splitting is deferred until the dashboard
  (SPEC-030) proves it needs structure. This is the one sub-choice below
  0.8 — a question is filed (see Validation).

## Context

STAGE-007 makes `projects` a first-class entity. bragfile already has a
free-text `entries.project` column (DEC-005-era schema, `0001_initial`):
every brag optionally records the project string the user typed at
capture. The load-bearing question the whole stage hinges on is **how
that pre-existing string relates to the new registered-projects table** —
the answer determines whether the `0004_*` migration backfills anything,
what `brag project delete` does to entries that reference a project
(SPEC-029), and how the status dashboard counts a project's recent brags
(SPEC-030). STAGE-007's brief surfaced three candidate shapes and left
the choice to this spec (SPEC-027).

The choice is also the **SPEC-027 split-watch trigger**: the stage flagged
SPEC-027 as L-risk and held it to M on the condition that DEC-017 not
require a non-trivial `entries.project` backfill. Soft string match's
defining property is **zero backfill** — which is precisely what keeps the
foundation spec at M.

A second framing point: bragfile is a **capture-first, append-only**
tool. A brag is a historical record of a moment. The project string on an
old entry is *what the user wrote then* — arguably it should not be
retroactively rewritten when a project is later registered or renamed.
Soft match honors that; a hard FK fights it.

## Alternatives Considered

- **Option A: Hard FK — `entries.project_id REFERENCES projects(id)`.**
  - What it is: replace the free-text column with a foreign key; every
    entry's project must be a registered project.
  - Why rejected: forces a backfill decision for every existing entry
    whose string matches no registered project (create-on-migrate? a
    synthetic "unfiled" project? null?), each lossy or surprising. Breaks
    the capture-first ethos — you could no longer jot `--project
    whatever` without first registering it. Inverts the premise of every
    existing `ListFilter.Project` / `ByProject` / `GroupEntriesByProject`
    test (they read a string). And the backfill is exactly the non-trivial
    work that would push SPEC-027 to L. Highest cost, highest risk.

- **Option B: Free-text with optional link — keep the string, add a
  nullable `entries.project_id` that links when a name matches.**
  - What it is: backwards-compatible column addition; the link is
    populated where a registered project matches.
  - Why rejected: two sources of truth (`project` string + `project_id`)
    that must be kept consistent on every write, plus — to be *useful* —
    a backfill that links existing entries to matching projects at
    migration time. That backfill is the L-split-watch trigger; absorbing
    it would cross SPEC-027 into L (or peel a seventh spec). The
    rename-desync problem soft match has is *also* present here unless the
    link is maintained, so the extra column buys little. Real but
    middling cost for marginal benefit at this stage.

- **Option C (chosen): Soft string match — free text, opportunistic join
  on `projects.name`.**
  - What it is: `entries.project` unchanged; "entries of a project" is
    `entries.project = projects.name` at query time.
  - Why selected: **zero migration risk, zero backfill** (the property
    that holds SPEC-027 at M); registration is purely additive (you can
    register a project for brags you already filed, with no data touch);
    **cleanest delete semantics** — deleting a project leaves entries
    completely untouched (no cascade, no `SET NULL`, no orphan handling),
    they simply keep the string they always had; faithful to the
    capture-first, append-only ethos; preserves every existing
    `entries.project` test premise verbatim. The tradeoff is **project-
    rename desync**: renaming a registered project's name does not rewrite
    the strings on old entries. Weighed against optional-link, this is
    acceptable and arguably *correct* — a brag is a historical record, and
    a later rename should not rewrite history. If real usage shows the
    desync hurts, the optional-link column is a clean additive follow-up
    (it supersedes this DEC; no rework of what ships now).

## Consequences

- **Positive:** `0004_*` is a pure additive schema migration — no
  `entries` read/write, no lossy backfill decision, no L-split. Every
  existing project-touching test (`list --project`, the `ByProject` /
  `GroupEntriesByProject` digests) keeps passing unchanged. `brag project
  delete` (SPEC-029) has the simplest possible blast radius on entries:
  none. Registration is retroactive and additive.
- **Positive:** the polymorphic-tags forward guarantee (DEC-015) is
  untouched — projects are independently a second taggable type whenever
  STAGE-007 chooses to write `'project'` taggings.
- **Negative:** no referential integrity between `entries.project` and
  `projects.name`; the same human project can be spelled two ways across
  entries and they won't co-count. **Project rename desyncs old entries'
  strings** — the explicit, accepted tradeoff.
- **Negative:** "recent-brag count" (SPEC-030) is a string-equality join,
  so it's exact-match only (no case-folding, no fuzzy). Acceptable at
  personal scale; revisit with the dashboard if it bites.
- **Neutral:** the `status` enum and `state_note` columns ship in `0004_*`
  now even though SPEC-030 renders them — "schema down once."

## Validation

Right if:
- `0004_*` applies on a populated v0.2.x dev DB with **zero** change to any
  `entries` row, and every existing `entries.project` test passes unchanged
  (SPEC-027 lossless-and-unchanged test).
- SPEC-030's recent-brag count via `entries.project = projects.name`
  reads naturally and needs no schema change.
- `brag project delete` (SPEC-029) defines an entries blast radius of
  "none" without contortion.

Revisit if:
- Project rename desync proves painful in real use → add the optional
  `entries.project_id` link column (Option B) as an additive,
  superseding follow-up. **No data ships now that this would have to
  undo.**
- A consumer needs case-insensitive or fuzzy project matching at the join.

The composite confidence is **0.80**: the soft-match relationship and the
status enum are well-grounded (0.80–0.82); the **single free-text
`state_note` vs. a split `state` + `next_action` pair** sub-choice is
softer (~0.78) because the dashboard (SPEC-030) might want structure.
Per §14 and the spec's design instruction (sub-choice < 0.8), a question
is filed: `project-state-note-shape` in `/guidance/questions.yaml`,
to be resolved at SPEC-030 design before the dashboard renders the note.

## References

- Related specs: **SPEC-027** (emits this DEC; the `0004_*` migration +
  the Store read primitives), SPEC-028/029 (the CRUD CLI + mutations that
  build on the soft-match model), SPEC-030 (the dashboard whose recent-brag
  count is the `entries.project = projects.name` join), SPEC-031 (the
  `here` resolver relying on the `project_locations.path` global-uniqueness
  guarantee laid down here).
- Related decisions: DEC-002 (forward-only embedded migrations — the
  `0004_*` mechanism; the no-`CHECK` rationale for `status`), DEC-005
  (INTEGER autoincrement PKs — `projects.id`, `project_locations.id`),
  DEC-015 (polymorphic taggings — projects as a second taggable type,
  schema-ready only here), DEC-011/013/014 (output shapes — untouched;
  the project renderers in SPEC-028+ reuse the envelope family).
- Related questions: `project-state-note-shape` in
  `/guidance/questions.yaml`.
- Related docs: `docs/data-model.md` (the two new entity tables).
