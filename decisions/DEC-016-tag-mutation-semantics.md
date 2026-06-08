---
# Maps to ContextCore insight.* semantic conventions.

insight:
  id: DEC-016
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

created_at: 2026-06-07
supersedes: null
superseded_by: null

tags:
  - cli
  - tags
  - mutations
  - data-model
  - io-contract
---

# DEC-016: Tag mutation semantics — rename errors-into-existing, merge de-dups via DELETE+INSERT, orphans are invisible (no GC)

## Decision

The tag-taxonomy mutation surface built on DEC-015's `tags` + `taggings`
schema (delivered by SPEC-026: `brag tags`, `brag tag rename`, `brag tag
merge`) is governed by four locked choices:

1. **`brag tags` lists only in-use tags, counted across all taggable
   types.** The taxonomy view is sourced through the `taggings` join
   (`tags JOIN taggings … GROUP BY tag_id`), so a tag's count is its
   total membership across every `taggable_type` (only `'entry'` exists
   today; `'project'` joins for free in STAGE-007). A tag with zero
   memberships (an *orphan*) does **not** appear. Default sort is
   **count DESC, then name ASC** — the established DEC-013/DEC-014
   count-ordering rule. Output is plain tab-separated `name\tcount` rows
   (human/pipe default) or a naked JSON array of `{ "tag": …, "count": … }`
   objects under `--format json` (DEC-011 naked-array discipline; the
   `{tag,count}` element shape matches DEC-014's `top_tags`).

2. **`rename <old> <new>` errors when `<new>` already exists — it does
   not auto-merge.** Rename is a single `UPDATE tags SET name=?` (which
   fires the SPEC-025 `tags_au` trigger and re-syncs FTS automatically).
   When `<new>` already names a tag, the `tags.name` UNIQUE constraint
   collides; rename returns a user error directing the caller to `brag
   tag merge`, rather than silently folding memberships and destroying a
   tag. Rename requires `<old>` to exist and `<old> != <new>`.

3. **`merge <src> <dst>` folds `src` into `dst` by DELETE+INSERT on
   `taggings` (never `UPDATE taggings SET tag_id`), de-duplicating, then
   deletes the `src` tag row.** Both `src` and `dst` must already exist
   (else a user error — "use rename to create it"); `src != dst`. The
   mechanism is forced by the SPEC-025 trigger topology: there is **no
   `taggings_au` trigger**, and `UPDATE taggings SET tag_id=dst WHERE
   tag_id=src` would both (a) violate `UNIQUE(taggable_type,
   taggable_id, tag_id)` for any object tagged *both* `src` and `dst`,
   and (b) silently desync `entries_fts` (no update trigger fires).
   Instead, inside one transaction: INSERT `dst` taggings for every
   `src` membership that lacks a `dst` membership (`taggings_ai` fires →
   FTS re-syncs), DELETE all `src` taggings (`taggings_ad` fires → FTS
   re-syncs), DELETE the `src` tag row. The result is exactly one `dst`
   tagging per affected object, FTS correct, **no schema change / no
   `0004_*` migration**.

4. **Orphan tags are left in place; there is no GC command and no
   auto-sweep in SPEC-026.** A `tags` row can reach zero memberships via
   `Store.Update` replacing an entry's tags (SPEC-025 deferred this).
   Because choice (1) makes orphans invisible at every read surface and
   `INSERT OR IGNORE INTO tags(name)` transparently *reuses* an existing
   orphan row on the next write, an orphan costs nothing observable.
   `merge` removes `src` entirely (not an orphan); `rename` creates no
   orphan. A dedicated `brag tag gc` / autocomplete-driven cleanup is
   deferred to a future spec if a consumer ever reads the `tags` table
   directly.

## Context

STAGE-006 normalizes tag storage. SPEC-025 (shipped) laid down the
schema, the ETL, the read/write cutover, and the six-trigger FTS
topology — but deliberately shipped **no command surface** and left
three behavioral questions open for SPEC-026, the stage's second and
final spec:

- What does the taxonomy view (`brag tags`) show, and how is it sorted?
- What happens when `rename`'s target name already exists?
- How is `merge` implemented so it de-dups *and* keeps FTS correct given
  there is no `taggings_au` trigger?
- Who, if anyone, cleans up orphan `tags` rows?

These are durable, user-visible contract choices (an upgrade-visible
payoff of the normalization, per the STAGE-006 "Why Now") that outlive
SPEC-026, so they warrant a DEC rather than living only in spec prose.
The schema itself is DEC-015; this DEC records the *behavior* layered on
top. All four choices were pre-flighted at SPEC-026 design (§12(b))
against `modernc.org/sqlite` 1.51.0 / SQLite 3.53.1 on a representative
corpus: rename auto-re-syncs FTS and rename-into-existing raises
`UNIQUE constraint failed: tags.name (2067)`; merge de-dups an
object tagged both `src` and `dst` to a single `dst` tagging with FTS
showing only `dst` and the `src` tag row removed.

## Alternatives Considered

- **rename auto-merges into the existing target (rejected, choice 2).**
  - What it is: when `<new>` exists, fold `<old>`'s memberships into it
    and drop `<old>` — i.e. make rename a superset of merge.
  - Why rejected: principle of least surprise. A command named *rename*
    that silently deletes a second tag and re-points its memberships is
    a data-losing side effect the user did not ask for. Keeping rename
    and merge orthogonal (rename = "this tag was misnamed"; merge =
    "these two tags are the same concept") makes each predictable, keeps
    rename a pure single-statement `UPDATE`, and the error message hands
    the user the right tool. Symmetric with merge requiring `dst` to
    already exist (choice 3).

- **merge via `UPDATE taggings SET tag_id = dst WHERE tag_id = src`
  (rejected, choice 3).**
  - What it is: the obvious one-statement re-point.
  - Why rejected: load-bearing FTS gotcha. SPEC-025 created
    `taggings_ai` (AFTER INSERT) and `taggings_ad` (AFTER DELETE) but
    **no `taggings_au`** — so an UPDATE fires nothing and `entries_fts`
    silently desyncs. Worse, an object tagged both `src` and `dst`
    violates `UNIQUE(taggable_type, taggable_id, tag_id)`. DELETE+INSERT
    fires the existing add/delete triggers and the `NOT EXISTS` guard
    skips the would-be duplicate, so it is correct with zero schema
    change.

- **Add a `taggings_au` trigger in a new `0004_*` migration, then merge
  via UPDATE (rejected).**
  - What it is: close the trigger gap and use the natural UPDATE.
  - Why rejected: a migration (and the §9 count-bump it triggers across
    `store_test.go` / `fts_test.go`, plus a §12(b) migration pre-flight)
    is real cost to buy nothing DELETE+INSERT doesn't already give. The
    UNIQUE-violation case would *still* need special handling (an
    object tagged both src+dst can't be UPDATE-ed onto dst). DELETE+INSERT
    is strictly simpler and was pre-flighted green. Reserve a `0004_*`
    for a change that genuinely needs new schema (STAGE-007 projects).

- **`brag tags` lists every `tags` row including orphans (rejected,
  choices 1 + 4).**
  - What it is: source from the `tags` table via LEFT JOIN so a
    zero-membership tag shows with count 0.
  - Why rejected: an orphan is not part of the *live* taxonomy — nothing
    is tagged with it — so surfacing it is noise, and it would force a GC
    command to keep the view tidy. Counting through the (inner) join
    makes orphans invisible, which in turn makes GC unnecessary in
    SPEC-026. "Every tag" is read as "every tag in use."

- **Eager orphan GC on every `Update`/`merge` (rejected, choice 4).**
  - What it is: after any operation that could drop a tag's last
    membership, `DELETE FROM tags WHERE id NOT IN (SELECT tag_id FROM
    taggings)`.
  - Why rejected: gold-plating. Orphans are invisible (choice 1) and
    reused by `INSERT OR IGNORE` on the next write, so they are
    functionally harmless. Eager GC adds a write-path sweep and a
    delete-while-iterating hazard for no observable benefit. Deferred
    until a consumer reads `tags` directly (e.g. autocomplete).

- **Chosen: all four together** — in-use-only counted taxonomy (DESC
  count / ASC name), rename-errors-into-existing, merge-by-DELETE+INSERT,
  no orphan GC. Each is the least-surprising, smallest-correct option,
  and together they make the mutation surface a clean, schema-stable
  byproduct of DEC-015.

## Consequences

- **Positive:** The taxonomy + mutation surface ships with **no schema
  change** — entirely Store methods + thin CLI commands. FTS stays
  correct through the existing SPEC-025 triggers (rename via `tags_au`,
  merge via `taggings_ai`/`taggings_ad`). rename and merge are
  orthogonal and predictable. `brag tags` counts generalize to
  `project` taggings in STAGE-007 with no change. Orphans need no
  attention.
- **Negative:** merge is multi-statement (INSERT-where-absent → DELETE
  taggings → DELETE tag) and must run in a transaction; a reviewer must
  understand *why* it is not a one-line UPDATE (the FTS/UNIQUE gotcha).
  Orphan `tags` rows accumulate over a long edit history — bounded,
  invisible, reused, but not zero on disk.
- **Neutral:** rename-into-existing and merge-missing-`dst` both return
  user errors that name the other command; this is a deliberate
  symmetry, not a dead end. The `brag tags` JSON element key is `tag`
  (matching DEC-014 `top_tags`), not `name` (the SQL column) — a small,
  documented divergence chosen for cross-command consistency.

## Validation

Right if:
- After `brag tag rename a b`, `brag search b` finds every entry
  formerly tagged `a` and `brag search a` finds none by tag; renaming
  into an existing name returns a user error pointing at `merge`.
- After `brag tag merge src dst`, an entry tagged both ends with exactly
  one `dst` tagging, `dst`'s count rises by the de-duplicated membership
  set, the `src` tag row is gone, and FTS reflects `dst` not `src`.
- `brag tags` shows counts DESC (name-ASC tiebreak), omits orphans, and
  its JSON form is a naked `[{tag,count}]` array.
- No `0004_*` migration is added by SPEC-026.

Revisit if:
- A consumer (autocomplete, a tag picker) reads the `tags` table
  directly and needs orphans gone → add a GC spec (and possibly a
  `tags_ad`/orphan-sweep).
- Users report that rename-into-existing *should* merge → reconsider
  choice 2 (it is the one most likely to attract a "just do what I
  mean" request; confidence on it is ~0.8, not 1.0).
- A second taggable type lands (STAGE-007) and `brag tags` wants a
  per-type count breakdown → extend the view (additive; the single
  total-count column stays valid).

## References

- Related decisions:
  - **DEC-015** (the `tags` + polymorphic `taggings` schema, the ETL,
    the `Entry.Tags` projection, and the six-trigger FTS topology this
    DEC's mutations operate on). DEC-016 is the behavior layered on
    DEC-015's structure; it does not supersede it.
  - DEC-011 (naked JSON array + 2-space indent + `[]`-not-null) — the
    `brag tags --format json` shape inherits it.
  - DEC-013 / DEC-014 (count ordering: DESC by count, alpha-ASC
    tiebreak) — `brag tags` default sort inherits it; the JSON
    `{tag,count}` element shape matches DEC-014's `top_tags`.
  - DEC-002 (forward-only embedded migrations) — relevant by its
    *absence* here: merge deliberately needs no new migration.
  - DEC-005 (INTEGER autoincrement PKs) — `tags.id` / `taggings.id`.
- Related specs:
  - SPEC-025 (shipped) — built the schema + triggers; explicitly
    deferred the command surface and orphan GC to SPEC-026.
  - SPEC-026 (emits this DEC) — `brag tags` / `tag rename` / `tag
    merge`.
- Related questions: `tag-ordering-projection` (DEC-015) — merge reuses
  `src`'s `position` for inserted `dst` taggings; order-only, harmless,
  consistent with that open question.
- Related docs: `docs/api-contract.md` (the three new command sections),
  `docs/tutorial.md` (the tag-taxonomy workflow), `docs/data-model.md`
  (the read path counts through the join).
