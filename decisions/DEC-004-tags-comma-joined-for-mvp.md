---
insight:
  id: DEC-004
  type: decision
  confidence: 0.65
  audience:
    - developer
    - agent

agent:
  id: claude-opus-4-7
  session_id: null

project:
  id: PROJ-001
repo:
  id: bragfile

created_at: 2026-04-19
supersedes: null
superseded_by: null

tags:
  - data-model
  - mvp-tradeoff
---

# DEC-004: Store tags as a comma-joined TEXT column for MVP

## Decision

For PROJ-001, `entries.tags` is a single `TEXT` column holding a
comma-joined list (e.g., `"auth,perf,platform"`). Filtering by tag uses
SQL `LIKE '%tag%'` or an FTS5 match. No `tags` / `entry_tags` tables.

## Context

The spec has `tags` as a user-visible field with filtering (`brag list
--tag ...`) and FTS (`brag search`). The "right" relational shape is a
`tags(id, name)` table plus an `entry_tags(entry_id, tag_id)` join.
That shape also enables tag rename, tag autocomplete, and a `brag tags`
listing.

At MVP scale (one user, hundreds of rows, tags typed by hand) none of
those features have value yet, and the migration cost is low whenever
they do.

Confidence is 0.65 because we know the choice is suboptimal past a
certain scale — we are explicitly accepting that tradeoff. A companion
question is in `/guidance/questions.yaml` to force a revisit when
STAGE-002 implements `list --tag`.

## Alternatives Considered

- **Option A: Normalized `tags` + `entry_tags` tables**
  - What it is: Two tables, proper FK, `SELECT ... JOIN ...` queries.
  - Why rejected (for MVP): More schema, more code, features that
    would justify it (rename, listing, autocomplete) are all deferred.

- **Option B: JSON column**
  - What it is: `tags TEXT` holding a JSON array.
  - Why rejected: SQLite's JSON functions work but add ceremony; comma
    strings are more grep-friendly for the humans who will open the
    DB file during debugging.

- **Option C (chosen): Comma-joined TEXT**
  - What it is: One column, one string, filter with LIKE or FTS5.
  - Why selected: Smallest code path, smallest schema, trivially
    round-trippable to/from the editor-launch form
    (`tags: auth, perf`). Cheapest to migrate away from later
    (one-shot SQL script splitting the column).

## Consequences

- **Positive:** Minimal schema. Round-trip to the markdown editor form
  is obvious. Matches how the user types it.
- **Negative:** No tag rename, no tag listing, no autocomplete. LIKE
  filters have false-positive risk (tag `"auth"` matches `"authoring"`
  unless we LIKE against `",<tag>,"` with sentinel commas).
- **Negative:** Tag FTS depends on how we tokenize in the `entries_fts`
  virtual table (STAGE-002). Needs a deliberate choice then.

## Validation

Right if:
- Tag filter behaves correctly (tests must cover the `"auth"` vs
  `"authoring"` edge case).
- We don't discover a user need for tag rename or listing during
  PROJ-001.

- Validated during SPEC-007 (2026-04-20): sentinel-comma LIKE pattern
  (`',' || tags || ',' LIKE '%,<tag>,%'`) handles the `"auth"` vs
  `"authoring"` false-positive correctly; no normalization needed at
  MVP scale. Answers the `tags-storage-model` question.

Revisit if:
- Tag rename becomes a user ask.
- Tag count grows past ~100 unique values (unlikely at personal
  scale).
- FTS5 tokenization of comma-joined tags produces bad search behavior
  in STAGE-002 — at that point the normalized-table option becomes
  more attractive than the tokenizer tweak.

## References

- Related specs: SPEC-002 (initial migration), future STAGE-002 specs
  (`search`, `list --tag`).
- Related questions: `tags-storage-model` in `/guidance/questions.yaml`.
- Related docs: `./docs/data-model.md`.
