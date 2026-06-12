---
# Maps to ContextCore insight.* semantic conventions.

insight:
  id: DEC-020
  type: decision
  confidence: 0.82
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

created_at: 2026-06-11
supersedes: null
superseded_by: null

tags:
  - data-model
  - projects
  - locations
  - editing
---

# DEC-020: `brag project edit` location-editing semantics

## Decision

`brag project edit` gains `--add-path` and `--remove-path`, backed by a
new Store method `RemoveLocation` (the single-row counterpart to the
shipped `AddLocation`) and a transactional batch method `EditLocations`.
Five semantics are fixed here:

1. **Removing a path attached to no project is a user error** —
   `ErrLocationNotFound`, not a silent no-op. A `--remove-path` typo is
   caught and reported rather than swallowed, mirroring how every other
   "operate on a thing that must exist" path in the CLI surfaces a miss.

2. **Removing a path attached to a *different* project is refused** —
   `ErrLocationOtherProject`. Paths are globally `UNIQUE(path)` (SPEC-027),
   so a path belongs to at most one project; `edit <p> --remove-path <q>`
   where `<q>` belongs to another project never silently deletes that
   other project's location. (Re-homing a path is `remove` on the owner
   then `add` on the new project — two deliberate edits, not one
   accidental one.)

3. **Paths match verbatim against the stored value.** `AddLocation`
   stores paths verbatim and the `new`-command path pre-check compares
   verbatim; `RemoveLocation` matches verbatim too, so add and remove are
   perfectly symmetric — the path you remove is the exact string you
   added. `filepath.Clean`/normalization lives **only** at cwd-resolution
   time (`ProjectForPath`, DEC-019); the *storage* layer is verbatim end
   to end. A `--remove-path /a/b/` will not match a stored `/a/b`; that is
   the intended, predictable consequence of the verbatim contract (and is
   reported as `ErrLocationNotFound`, not a silent miss).

4. **A single `edit` invocation's location changes apply atomically** —
   one transaction, all-or-nothing, **removes applied before adds**. The
   removes-before-adds order lets a path be removed and re-added in the
   same call (e.g. to re-home it within the same project) without a
   transient `UNIQUE(path)` collision, and a mid-batch failure (a typo'd
   remove, an occupied add, an in-batch duplicate) rolls the entire set
   back, leaving `project_locations` exactly as it was. This is the
   manual single-row / single-batch counterpart to DEC-018's in-tx
   delete cascade.

5. **Location edits do not bump `projects.updated_at`.** `updated_at`
   tracks **scalar-field** recency (the SPEC-029 `UpdateProject` contract:
   name/status/state_note). Editing the `project_locations` set is a
   structural change to a *different* table and leaves `updated_at`
   untouched, so `EditLocations` never writes the `projects` row. Scalar
   edits (`--name`/`--status`/`--state-note`) still bump it as before.

A companion contract rides with this DEC: **scalar edits and location
edits in one `edit` call are applied sequentially (scalar first, then
locations), each atomic, but not jointly atomic.** Scalar-first means an
invalid `--status` or a duplicate `--name` aborts inside `UpdateProject`
*before* any location is touched (the cheap validation fails fast). The
only residual partial window is a valid scalar change followed by a
failing location change (e.g. a `--remove-path` typo): the scalar change
stands and the user re-runs the location edit. Joining the two into one
transaction would require unifying `UpdateProject` with the location
methods (which today each own their transaction) for a benign,
recoverable window — disproportionate to the value.

## Context

STAGE-007's `brag project edit` (SPEC-029) shipped **scalar-only**
(`--name`/`--status`/`--state-note`); location editing was the L-watch
that fired and was peeled into **SPEC-033** (this spec). The peel left
four cross-cutting semantics undecided — what "remove" means when the
path is missing or owned by another project, how paths are matched, and
whether a multi-path edit is atomic. These are durable and cross-spec:

- `RemoveLocation` is the inverse of `AddLocation`'s global-uniqueness
  guarantee (SPEC-027) and the manual single-row counterpart to DEC-018's
  delete-time `project_locations` cascade. The "one path → at most one
  project" invariant that SPEC-031's `here` resolver (DEC-019) relies on
  is preserved on the *remove* side by refusing cross-project deletes.
- The verbatim-matching choice is the storage-side half of the
  SPEC-031/DEC-019 division of labor ("storage is verbatim; the resolver
  normalizes"). Deciding it once, here, keeps add and remove symmetric
  and keeps any future location feature from re-litigating it.

Confidence is **0.82**: the other-project refusal (0.90) and the
not-attached error (0.85) are well-grounded; verbatim matching (0.85) and
the in-tx atomicity model (0.82) are the softer sub-choices (a future
spec could add a normalizing match, or unify scalar+location into one
transaction, as additive follow-ups). No sub-choice is below 0.70, so no
`/guidance/questions.yaml` entry is filed (§14).

## Alternatives Considered

- **Not-attached remove → silent no-op (instead of `ErrLocationNotFound`).**
  - Why rejected: a `--remove-path` typo would silently do nothing and
    report success, the worst outcome for a destructive-intent command.
    Surfacing the miss as a user error catches the typo at the cheapest
    moment and matches the rest of the CLI's "no such thing" handling.

- **Clean both sides before matching (instead of verbatim).**
  - Why rejected: it would make `remove` use a *different* matching basis
    than `add` and the `new`-command pre-check (both verbatim), an
    asymmetry that could let a remove match a subtly different stored
    string. Verbatim keeps add/remove symmetric and confines all
    normalization to the cwd resolver (DEC-019), where it belongs. If a
    normalizing remove is ever wanted, it is a clean additive change that
    supersedes this sub-choice with no data to undo.

- **Sequential, non-atomic location ops (instead of one transaction).**
  - Why rejected: a mid-batch failure (occupied add, typo'd remove) would
    leave a half-edited location set — exactly the surprise the atomic
    batch prevents. The transaction is one `BeginTx`/`Commit` around a
    short loop, mirroring `DeleteProject`'s shape; cheap and correct.

- **Adds before removes (instead of removes-before-adds).**
  - Why rejected: re-adding a path that is also being removed (re-homing
    within the same project) would hit a transient `UNIQUE(path)`
    collision because the old row still exists when the add runs.
    Removes-first frees the path first.

- **One transaction spanning scalar + location edits.**
  - Why rejected: `UpdateProject` and the location methods each own their
    transaction today; merging them into a single combined method
    duplicates `UpdateProject` to close a benign, recoverable window
    (valid scalar + failing location). Scalar-first ordering already
    makes the common "typo'd status alongside a path change" case abort
    before any location write.

- **Bump `updated_at` on location edits.**
  - Why rejected: `updated_at` is the scalar-field recency signal
    (SPEC-029); making `EditLocations` also write the `projects` row to
    move it conflates structural location changes with scalar-field edits
    and complicates the location-only path for no clear gain. If location
    activity should affect recency, that is an additive change later.

## Consequences

- **Positive:** `RemoveLocation` is the clean inverse of `AddLocation`;
  the global "one path → one project" invariant is preserved on remove
  (cross-project deletes refused), so SPEC-031's resolver guarantee still
  holds after edits. Atomic batches mean no half-edited location sets.
  A removed path is immediately free for re-registration (same as after
  `brag project delete`, DEC-018).
- **Positive:** verbatim matching keeps `add`/`remove`/`new` on one
  matching basis and confines normalization to DEC-019.
- **Negative:** a path stored with a trailing slash or `.` segment must
  be removed with the identical string; a normalized-input variant will
  report `ErrLocationNotFound`. Accepted: it is predictable, and the
  error names the miss rather than silently doing nothing.
- **Negative:** scalar + location edits in one call are not jointly
  atomic; a valid scalar edit can stand while a subsequent location edit
  fails. Accepted and documented (the scalar change is benign and the
  location edit is re-runnable).
- **Neutral:** location edits leave `updated_at` unmoved; a project's
  recency in `brag project list` reflects scalar edits only.

## Validation

Right if:
- `RemoveLocation` deletes an attached path and frees it for
  re-registration; removing a not-attached path returns
  `ErrLocationNotFound`; removing another project's path returns
  `ErrLocationOtherProject` and leaves that path attached (SPEC-033
  storage tests).
- A batch `EditLocations` with a mid-batch failure leaves
  `project_locations` byte-for-byte unchanged (rollback test).
- `edit <p> --remove-path /a --add-path /a` (same path) succeeds and `/a`
  ends attached once (removes-before-adds).

Revisit if:
- Real use shows verbatim matching is too strict (paths registered with
  inconsistent trailing slashes) → add a normalizing match as an additive
  superseding sub-choice.
- A consumer needs scalar + location edits to be jointly atomic → unify
  the scalar and location mutation paths into one transaction (its own
  decision; supersedes the companion contract here).

## References

- Related specs: **SPEC-033** (emits this DEC; `--add-path`/`--remove-path`
  + `RemoveLocation`/`EditLocations`), SPEC-029 (the scalar `edit` this
  extends; the L-watch peel that created SPEC-033), SPEC-027 (the
  `project_locations` schema + `UNIQUE(path)` + `AddLocation` this is the
  counterpart to), SPEC-031 (the `here` resolver whose "one path → one
  project" guarantee the cross-project-remove refusal preserves).
- Related decisions: **DEC-018** (delete blast radius — the in-tx
  `project_locations` cascade that `RemoveLocation` is the manual
  single-row counterpart to; both free the `UNIQUE` path for reuse),
  **DEC-017** (soft string match — entries are independent of project
  locations, so location edits never touch `entries`), **DEC-019**
  (cwd-resolution normalization — the read-time normalization this DEC's
  verbatim storage-side matching deliberately complements), DEC-002
  (forward-only migrations — why FK enforcement stays off, so a removed
  location row is gone for good), DEC-005 (autoincrement PKs), DEC-006 /
  DEC-007 (cobra command + inline `RunE` validation).
- Related docs: `docs/api-contract.md` (the `brag project edit` section).
