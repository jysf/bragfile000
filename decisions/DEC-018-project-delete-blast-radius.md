---
# Maps to ContextCore insight.* semantic conventions.

insight:
  id: DEC-018
  type: decision
  confidence: 0.85
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

created_at: 2026-06-09
supersedes: null
superseded_by: null

tags:
  - data-model
  - projects
  - deletion
  - blast-radius
---

# DEC-018: `brag project delete` blast radius

## Decision

Deleting a project (`brag project delete`, `Store.DeleteProject`,
SPEC-029) removes the project and exactly its dependent rows, **in one
transaction**, and **leaves `entries` untouched**. Concretely, the
delete touches three relations and no others:

1. **`entries` — UNTOUCHED.** Under DEC-017 (soft string match) an
   entry owns the free-text `project` string it was captured with. A
   project's deletion does **not** rewrite, null, or remove any entry.
   `brag list --project <name>` keeps returning those entries after the
   project is gone. The blast radius on entries is the simplest possible:
   **none**.
2. **`project_locations` — DELETED MANUALLY, in-transaction.** FK
   enforcement is **OFF** in bragfile (`PRAGMA foreign_keys=ON` is never
   set — verified at SPEC-027), so the `project_id INTEGER NOT NULL
   REFERENCES projects(id)` clause is **decorative**: no `ON DELETE`
   cascade fires. `DeleteProject` therefore issues an explicit `DELETE
   FROM project_locations WHERE project_id = ?` inside the same
   transaction. This is also what **frees the globally-`UNIQUE` path**
   for re-registration to another project — without it, the path row
   would dangle and permanently block reuse.
3. **`'project'` taggings — DELETED, in-transaction.** No command writes
   `taggable_type='project'` taggings yet (the project-tag surface is
   schema-ready only per DEC-015; deferred to STAGE-008/PROJ-003). The
   `DELETE FROM taggings WHERE taggable_type='project' AND taggable_id=?`
   is therefore a **no-op today**, but is laid down now so the eventual
   project-tag surface inherits correct cleanup with **no change to
   delete** — mirroring how `Store.Delete` already removes an entry's
   taggings before the entry row.

A companion contract rides with this DEC: the **archive-vs-delete
distinction**. `brag project archive` (`Store.ArchiveProject`) is a
**non-destructive, recoverable** status flip (`status='archived'`,
everything else preserved, restorable via `edit --status active`);
`brag project delete` is **irreversible**. The two are deliberately kept
distinct in help text, and delete is confirmation-guarded (y/N + `--yes`/
`-y`, matching `brag delete` for entries).

## Context

STAGE-007's Success Criteria flag delete's blast radius as something that
must be "defined, tested, and consciously chosen — not incidental." The
question has three independent edges, each settled above:

- **What happens to brag entries that reference the project?** DEC-017
  already answered the *relationship* (soft string match), which makes
  the *delete* answer fall out cleanly: nothing. A brag is a historical
  record; deleting a registered project must not rewrite what the user
  captured. (An FK or optional-link model would have forced a
  cascade/SET-NULL/orphan decision here — one more reason DEC-017 chose
  soft match.)
- **What happens to the project's locations?** They must go — but the
  schema's `REFERENCES` does not remove them, because FK enforcement is
  off. SPEC-027's ship reflection explicitly carried this forward as the
  load-bearing input for SPEC-029: `DeleteProject` must delete
  `project_locations` manually. Missing this would dangle rows and, worse,
  keep the unique path reserved forever.
- **What happens to any `'project'` taggings?** None exist yet, but the
  delete should define the answer rather than leave a latent orphan bug
  for whatever spec first writes them. Cleaning them in-tx now is one
  harmless statement that future-proofs the project-tag surface.

This is the strongest DEC candidate STAGE-007 flagged because the
contract is **durable and cross-spec**: SPEC-030's recent-brag count
(`entries.project = projects.name`) relies on entries surviving a
project's deletion; any future `brag project tag` spec relies on delete
already cleaning `'project'` taggings; and the path-reuse guarantee
SPEC-031's resolver implicitly assumes (one path → at most one project)
depends on delete actually freeing the path.

## Alternatives Considered

- **Rely on the `REFERENCES` clause to cascade locations.**
  - Why rejected: **factually wrong** for bragfile. FK enforcement is
    OFF, so SQLite never cascades — the `REFERENCES` is documentation,
    not behavior. Tested at SPEC-027. Relying on it would silently leave
    `project_locations` rows behind and block path reuse.

- **Cascade / `SET NULL` onto `entries` (hard-link model).**
  - Why rejected: rewrites captured history, violating DEC-017's
    capture-first ethos. It also only makes sense under an FK/optional-
    link relationship, which DEC-017 rejected. Deleting a project should
    not mutate brags the user already filed.

- **Defer the `'project'`-taggings delete until the project-tag surface
  exists.**
  - Why rejected: a latent orphan-row bug parked in a future spec for no
    saving. The cleanup is a single in-tx statement that is harmless
    while no `'project'` taggings exist and correct the moment they do.
    Cheaper to lay down once, now, than to remember to add it later.

- **Make `archive` a flavor of `delete` (soft delete behind a flag).**
  - Why rejected: STAGE-007 wants archive and delete "clearly distinct."
    A plain `archived` status the user can flip back with `edit --status
    active` is the simplest recoverable primitive and keeps `delete`
    unambiguously irreversible.

## Consequences

- **Positive:** `DeleteProject` is three `DELETE`s in one transaction
  with no `entries` write — the simplest correct blast radius. Path reuse
  works immediately after delete (the location row is gone). The future
  project-tag surface needs no delete change. Archive and delete are
  unambiguous to the user.
- **Positive:** every existing `entries`/`entries.project` test premise
  is preserved (delete touches no `entries` row), consistent with
  DEC-017's inversion-free guarantee.
- **Negative:** a deleted project's brags become "orphaned by string" —
  they still carry the project name but no registered project backs it.
  This is the accepted DEC-017 tradeoff (a brag is a historical record),
  not new debt: re-registering a project of the same name re-associates
  them by the soft-match join with zero data work.
- **Neutral:** the `'project'`-taggings `DELETE` executes on every
  project delete even though it affects zero rows until the tag surface
  ships — negligible cost, and it removes a future foot-gun.

## Validation

Right if:
- `DeleteProject` removes the project + its `project_locations` rows + any
  `'project'` taggings in one tx, and a `SELECT COUNT(*)` confirms the
  location/tagging rows are gone (SPEC-029 storage tests).
- After deleting a project, the path it held can be re-registered to a new
  project (no `ErrLocationExists`) — the CLI-observable proof the manual
  cascade fired.
- After deleting a project, a brag entry captured with that project string
  still exists and `brag list --project <name>` still returns it (entries
  blast radius = none).

Revisit if:
- FK enforcement is ever turned ON (`PRAGMA foreign_keys=ON`) — then the
  manual `project_locations` delete becomes redundant with a real cascade;
  that would be its own decision (and a migration/runtime-pragma change).
- A project-tag surface ships and wants different delete semantics for
  `'project'` taggings (e.g. preserve-and-reassign) — supersede this DEC.

## References

- Related specs: **SPEC-029** (emits this DEC; `UpdateProject` /
  `ArchiveProject` / `DeleteProject` + the mutation CLI), SPEC-027 (the
  schema + the FK-off forward note this realizes), SPEC-028 (the
  `new`/`list`/`show` surface this completes), SPEC-030 (recent-brag count
  relies on entries surviving delete), SPEC-031 (path-reuse / one-path-one-
  project guarantee), SPEC-033 (location editing — `RemoveLocation` —
  peeled from SPEC-029).
- Related decisions: **DEC-017** (soft string match — the reason entries
  are untouched; the status enum `archive` flips within), DEC-015
  (polymorphic taggings — projects as a second taggable type, schema-ready
  only; the `'project'`-taggings cleanup future-proofs it), DEC-002
  (forward-only migrations — why FK enforcement stays off rather than a
  table rebuild), DEC-005 (autoincrement PKs).
- Related docs: `docs/api-contract.md` (the `brag project edit/archive/
  delete` sections).
