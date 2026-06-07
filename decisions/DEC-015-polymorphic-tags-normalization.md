---
# Maps to ContextCore insight.* semantic conventions.

insight:
  id: DEC-015
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

created_at: 2026-06-06
supersedes: DEC-004
superseded_by: null

tags:
  - data-model
  - tags
  - migrations
  - polymorphic
---

# DEC-015: Normalize tags into a polymorphic `tags` + `taggings` model

## Decision

Tags become a normalized, first-class taxonomy shared across object
types. Two tables replace the comma-joined `entries.tags` column:

- **`tags(id INTEGER PK AUTOINCREMENT, name TEXT NOT NULL UNIQUE)`** —
  one row per distinct tag name. Names are stored trimmed.
- **`taggings(id INTEGER PK AUTOINCREMENT, tag_id INTEGER NOT NULL
  REFERENCES tags(id), taggable_type TEXT NOT NULL, taggable_id INTEGER
  NOT NULL, position INTEGER NOT NULL, UNIQUE(taggable_type,
  taggable_id, tag_id))`** — a polymorphic membership join. A tagging
  binds one tag to one *taggable* identified by the
  `(taggable_type, taggable_id)` pair. In STAGE-006 only
  `taggable_type = 'entry'` rows exist; STAGE-007 adds `'project'`
  rows with **no further schema change**.

Three contracts ride with the schema:

1. **ETL is lossless at the tag-set level.** Each entry's comma-joined
   `entries.tags` string splits into tokens, each token is `trim`-med,
   empty tokens are dropped, and **duplicates within one entry are
   collapsed to their first occurrence**. Every distinct trimmed token
   becomes a `tags` row (deduplicated globally via the `UNIQUE` name);
   every surviving `(entry, tag)` pair becomes a `taggings` row. The
   migration is **forward-only** (DEC-002): no down-migration; the
   downgrade path is the documented SQLite-file backup.

2. **`Entry.Tags` is a reconstructed projection, not stored state.**
   The Store keeps `Entry.Tags` as a comma-joined string for every read
   (`Get`/`List`/`Search`), reconstructed from the join via
   `GROUP_CONCAT(tags.name, ',' ORDER BY taggings.position)`. The
   `position` column preserves each entry's original tag order, so a
   **canonical** input (already trimmed, de-duplicated) round-trips
   byte-identically — `"perf,auth"` reads back `"perf,auth"`, not
   sorted to `"auth,perf"`. Non-canonical inputs are canonicalized
   (trimmed + de-duplicated) losslessly at the set level. This keeps
   DEC-011 (JSON `tags` as comma-joined string), DEC-013 (markdown
   export), and DEC-014 (stats "Top tags") byte-stable: those consumers
   read the projected string and never see the join.

3. **FTS indexes the projection, not the column.** `entries_fts.tags`
   stays a denormalized indexed column with the default unicode61
   tokenizer (DEC-010 search behavior unchanged), but it is fed the
   reconstructed projection and re-synced on `taggings`/`tags`
   mutations rather than on `entries`-row writes.

## Context

DEC-004 (confidence 0.65) stored tags as a comma-joined `TEXT` column
for the MVP. **This decision is DEC-004 firing exactly as designed — it
is not a reversal.** DEC-004's own Validation block named two revisit
triggers, and PROJ-002 trips **both**:

- *"Tag rename becomes a user ask."* `brag tag rename` is a STAGE-006
  success criterion (delivered in SPEC-026).
- *"A second consumer appears, at which point the normalized-table
  option becomes more attractive."* PROJ-002 makes **projects** a second
  taggable object type (STAGE-007).

DEC-004 explicitly anticipated this successor ("cheapest to migrate away
from later — one-shot SQL script splitting the column"), so DEC-015 is
the planned, conditional successor, not a correction of a wrong call.
The brief (reason 3) frames the repayment as cheapest *now*, while we
are already in the tag code, because each additional CSV-tag surface we
would otherwise migrate independently compounds the cost.

DEC-004 also predicted the negative it accepted: "no tag rename, no tag
listing, no autocomplete." Normalization unlocks all three.

## Generality — paper sketch of a third taggable type (no code)

The brief (reason 4) requires proving the abstraction generalizes past
two types. The probable third type is `goals` (PROJ-003 candidate, **not
built**). Under this schema it requires **zero schema change**:

```
-- hypothetical, NOT in PROJ-002:
INSERT INTO goals(name, ...) VALUES ('ship v0.3', ...);          -- id 7
INSERT INTO tags(name) VALUES ('q3') ON CONFLICT DO NOTHING;     -- id 12
INSERT INTO taggings(tag_id, taggable_type, taggable_id, position)
VALUES (12, 'goal', 7, 0);
```

`brag tags` would then count a tag's uses across `entry`, `project`,
and `goal` taggings with a single `GROUP BY tag_id` — no per-type
column, no per-type table. The `(taggable_type, taggable_id)` pair is
the whole generality mechanism: adding a type is adding a string value,
not a migration. This validates the polymorphic shape over a
`entry_tags(entry_id, tag_id)` + future `project_tags(...)` pair, which
would need a new join table (and new FTS wiring) per type.

## Alternatives Considered

- **Option A: Keep DEC-004 (comma-joined TEXT).**
  - Why rejected: both DEC-004 revisit triggers are tripped. Tag rename
    against a CSV column is an `UPDATE ... LIKE`-and-rewrite across every
    row with false-positive risk; a second taggable type means a second
    CSV column to split later. The debt only gets more expensive.

- **Option B: Type-specific join tables (`entry_tags`, later
  `project_tags`).**
  - What it is: the classic `tags(id,name)` + `entry_tags(entry_id,
    tag_id)` pair from `docs/data-model.md`'s "Future schema shapes."
  - Why rejected: does not generalize. Each new taggable type needs its
    own join table, its own FTS sync, and its own `brag tags`
    aggregation branch. The brief explicitly wants one model that a
    third type joins without re-litigation.

- **Option C: Tags-as-array in JSON, JSON column in SQLite.**
  - Why rejected: rejected at DEC-004 (Option B there) and again at
    DEC-011 (tags stay a string at the I/O boundary). A JSON column is
    no more queryable for cross-object taxonomy than CSV and breaks the
    byte-stable string contract.

- **Option D (chosen): Polymorphic `tags` + `taggings(taggable_type,
  taggable_id, position)`.**
  - Why selected: one model for all taggable types; `brag tags` /
    `rename` / `merge` become natural; the `position` column preserves
    byte-stable `Entry.Tags`; output shapes (DEC-011/013/014) and search
    (DEC-010) are untouched because the Store hands consumers the same
    projected string they always saw.

### Sub-choice softened: tag ordering within an entry

The one sub-choice below 0.8 is **how to order a reconstructed
`Entry.Tags` string**: preserve insertion order via the `position`
column (chosen) vs. sort `name ASC` (simpler, no column). Insertion
order is byte-faithful for *all* canonical inputs, not just
already-sorted ones, which is why it wins for a "lossless and invisible"
migration. But the entire existing test/production corpus that round-
trips through the Store is already alphabetically ordered, so `name ASC`
would also pass every test today — meaning `position` is insurance whose
value is not yet observable. Confidence on this sub-choice ≈ 0.75; a
question is filed (`tag-ordering-projection` in
`/guidance/questions.yaml`) to revisit dropping `position` if no
consumer ever depends on per-entry order. The composite decision is 0.80
(the schema, ETL, and projection-as-contract are all strong and
pre-flighted against the real driver; only the ordering sub-choice is
soft).

## Consequences

- **Positive:** Tag rename/merge/taxonomy become single-table
  operations (SPEC-026). A third taggable type is free. Output shapes
  and search are byte-stable because the Store projects the same string.
  The migration was validated end-to-end at design (§12(b)) against
  modernc.org/sqlite 1.51.0 / SQLite 3.53.1.
- **Negative:** Reads gain a correlated `GROUP_CONCAT` subquery; writes
  (`Add`/`Update`/`Delete`) become multi-statement and must run in a
  transaction (entry row + taggings). FTS sync moves from three
  `entries`-row triggers to a six-trigger topology (`entries` ×3,
  `taggings` ×2, `tags` ×1) — the highest-risk surface, isolated and
  test-heavy in SPEC-025.
- **Negative:** Orphan `tags` rows (a name with zero taggings after an
  `Update` drops its last use) are not garbage-collected in STAGE-006.
  Harmless for reads; cleanup is a SPEC-026 concern.
- **Neutral:** `entries.tags` is dropped in the same `0003_*` migration
  (single atomic in-place migration — see SPEC-025). DEC-011's
  field-names-match-SQL note now has one fewer backing column, but the
  JSON `tags` key is unchanged.

## Validation

Right if:
- Every distinct trimmed token in the pre-migration corpus becomes a
  `tags` row and every membership a `taggings` row (lossless-ETL test on
  a representative corpus, SPEC-025).
- `brag list --tag`, `brag search`, the exports, and the digests produce
  byte-identical output before and after on the same corpus.
- A canonical `Entry.Tags` string round-trips byte-identically through
  `Add` → `Get`/`List`.
- STAGE-007 adds `taggable_type='project'` rows with no schema change.

Revisit if:
- The `position` column proves unused by any consumer (then `name ASC`
  and drop the column — see the filed question).
- Cross-object tag counts (`brag tags`) need an index shape the two
  `taggings` indexes don't cover at real scale (unlikely at personal
  scale).

## References

- Supersedes: **DEC-004** (comma-joined tags for MVP) — `superseded_by:
  DEC-015` filled in there.
- Related decisions: DEC-002 (embedded forward-only migrations —
  mechanism), DEC-010 (search syntax — kept byte-stable), DEC-011 /
  DEC-013 / DEC-014 (output shapes — kept byte-stable via the
  projection), DEC-005 (INTEGER autoincrement PKs — `tags.id`,
  `taggings.id`).
- Related specs: SPEC-025 (emits this DEC; the atomic `0003_*`
  migration + Store cutover), SPEC-026 (`brag tags` / `tag rename` /
  `tag merge` on top).
- Related questions: `tag-ordering-projection` in
  `/guidance/questions.yaml`.
- Related docs: `docs/data-model.md` (schema + FTS), `docs/api-contract.md`
  (`list --tag`).
