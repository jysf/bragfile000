---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-025
  type: story                      # epic | story | task | bug | chore
  cycle: verify
  blocked: false
  priority: high
  complexity: L                    # S | M | L  (L accepted, not split — see Context)

project:
  id: PROJ-002
  stage: STAGE-006
repo:
  id: bragfile

agents:
  architect: claude-opus-4-8
  implementer: claude-opus-4-8     # usually same Claude, different session
  created_at: 2026-06-06

references:
  decisions: [DEC-015, DEC-004, DEC-002, DEC-010, DEC-011, DEC-013, DEC-014, DEC-005]
  constraints:
    - migrations-are-append-only
    - storage-tests-use-tempdir
    - no-sql-in-cli-layer
    - timestamps-in-utc-rfc3339
    - test-before-implementation
    - errors-wrap-with-context
    - no-cgo
    - one-spec-per-pr
  related_specs: [SPEC-011, SPEC-007, SPEC-026]
---

# SPEC-025: Normalize tag storage

## Context

This is the **first and load-bearing** spec of STAGE-006 (Tag
normalization), the foundation stage of PROJ-002 (Projects and tags).
When it ships, tags stop being a comma-joined string inside
`entries.tags` and become a normalized, polymorphic taxonomy that any
object type can share — laid down once so STAGE-007 (Projects) and
STAGE-008 (polish) build on it rather than migrating it twice.

It **emits DEC-015** at design (this spec), superseding DEC-004. DEC-004
stored tags as comma-joined `TEXT` for the MVP (confidence 0.65) and
named two revisit triggers — "tag rename becomes a user ask" and "a
second consumer appears." PROJ-002 trips both, so DEC-015 is DEC-004
firing **as designed**, not a reversal. See DEC-015 for the full
rationale and the paper sketch proving the model generalizes to a third
taggable type.

**Scope framing — single atomic in-place migration (complexity L,
accepted, not split).** STAGE-006 was reframed (working-tree decision
2026-06-06, confirmed by the spec author for this design) from a
3-spec expand→contract pair to a **2-spec** plan: SPEC-025 does the
whole migration in one atomic `0003_*` transaction (schema + ETL + FTS
re-point + drop the legacy column), and SPEC-026 adds the user-visible
taxonomy surface on top. The expand→contract split (a dual-written
`entries.tags` shadow carried across two PRs) was dropped as
over-engineered here: bragfile has effectively one user, v0.2.0 is
adopted only at project close, and the downgrade story is already a `cp`
backup of the SQLite file — so the inter-PR "keep `main` green" safety
the split bought is worthless. An import/export round-trip was also
considered and rejected (more error-prone than a tested transactional
migration, and it needs an import command that does not exist). The
atomic migration reads L; that is accepted because the single-user
premise removes the value a split would provide. The fallback, *only* if
the build genuinely cannot hold it together as one spec, is the
expand→contract split (back to 3 specs).

- Parent stage: `STAGE-006-tag-normalization.md` (this spec is the first
  of two; SPEC-026 follows).
- Project: `PROJ-002` brief (dev/prod DB isolation governs this stage —
  see Implementation Context).
- Supersedes the data shape from `DEC-004`; emits `DEC-015`.

## Goal

Replace the comma-joined `entries.tags` column with a normalized `tags`
+ polymorphic `taggings` model in one atomic forward-only `0003_*`
migration (lossless ETL, FTS re-pointed onto the join, the legacy column
dropped), and cut the Store's read/write/`list --tag` paths over to the
join — with **zero user-visible change**: every existing `brag list`,
`list --tag`, `search`, export, and digest output stays byte-identical.

## Inputs

- **Files to read:**
  - `internal/storage/store.go` — `Add`/`Get`/`Update`/`Delete`/`List`/
    `Search`; the read scans and the `List` tag-filter `LIKE` to replace.
  - `internal/storage/entry.go` — `Entry` (the `Tags` string field stays)
    and `ListFilter` (the `Tag` field stays; only its mechanism changes).
  - `internal/storage/migrate.go` — the runner: each `.sql` file is run
    as **one multi-statement `ExecContext` inside one transaction**
    (`runMigration`). The new migration must be a single such file.
  - `internal/storage/migrations/0001_initial.sql`,
    `0002_add_fts.sql` — the shape `0003_*` builds on; the FTS table +
    triggers `0003` replaces.
  - `internal/storage/store_test.go`, `fts_test.go`, `migrate_test.go`,
    `get_test.go` — the tests touched (see ## Outputs premise audit).
  - `internal/export/stats.go` — `extractTags` (consumes `Entry.Tags`;
    must stay byte-stable, untouched).
  - `internal/cli/list.go` — sets `ListFilter.Tag`; **must not change**
    (`no-sql-in-cli-layer`).
  - `decisions/DEC-015-*` (the schema/ETL/projection contract this spec
    implements), `DEC-002`, `DEC-010`, `DEC-011`, `DEC-013`, `DEC-014`.
- **External APIs:** none. The join + `GROUP_CONCAT` are plain
  `database/sql` against `modernc.org/sqlite` (no new dep — would need a
  DEC under `no-new-top-level-deps-without-decision`).
- **Related code paths:** `internal/storage/` (all changes live here;
  the CLI, export, aggregate, and editor layers are untouched).

## Outputs

- **Files created:**
  - `internal/storage/migrations/0003_normalize_tags.sql` — the atomic
    migration (full literal in ## Notes for the Implementer; validated
    at design, see §12(b) note below).
  - `decisions/DEC-015-polymorphic-tags-normalization.md` — **created at
    this design** (supersedes DEC-004; `DEC-004.superseded_by: DEC-015`
    filled in).
- **Files modified:**
  - `internal/storage/store.go` — `Add`/`Update`/`Delete` become
    transactional and write through the join (split → upsert `tags` →
    insert `taggings`); `Get`/`List`/`Search` reconstruct `Entry.Tags`
    via the projection subquery; `List` tag filter becomes an `EXISTS`
    against `taggings`.
  - `internal/storage/entry.go` — doc comments on `Entry.Tags` /
    `ListFilter.Tag` updated to reference DEC-015 (the fields' types are
    unchanged).
  - `docs/data-model.md`, `docs/api-contract.md`, `docs/architecture.md`
    — status-change doc sweep (enumerated below).
  - The test files enumerated in the premise audit below.
- **New exports:** none. `Entry`, `ListFilter`, and every `Store` method
  signature are unchanged — the normalization is entirely internal.
- **Database changes:** new `tags` + `taggings` tables (+ two indexes);
  `entries_fts` recreated as a regular (own-content) FTS5 table; FTS
  trigger topology rewritten (`entries` ×3, `taggings` ×2, `tags` ×1);
  `entries.tags` column **dropped**. Forward-only (DEC-002).

### Premise audit (`projects/_templates/premise-audit.md`), run at design

Greps were **run** at design and reconciled against the lists below.

```
- [x] Inversion/removal: greps run, invalidated tests listed as planned rewrites
- [x] Addition/count-bump: greps run, literal-count assertions listed as planned bumps
- [x] Status-change: greps run, every doc hit listed as updates/stays
- [x] Cross-check: actual grep hits reconciled against the lists above
```

**1. Addition / count-bump** — adding `0003_*` to `schema_migrations`.
Greps run: `grep -rn "schema_migrations" internal/**/*_test.go` and
`grep -rn "0002_add_fts" internal`. Coupled assertions (planned bumps,
`2 → 3` / extend the exact list to add `"0003_normalize_tags"`):

- `internal/storage/store_test.go`
  - `TestOpen_MigrationsTracked` — `want := []string{"0001_initial",
    "0002_add_fts"}` → append `"0003_normalize_tags"` (and the
    `versions[2]` check).
  - `TestOpen_Idempotent` — `count != 2` / "want 2" → `3`.
- `internal/storage/fts_test.go`
  - `TestFTS_BothMigrationsTracked` — `want := []string{"0001_initial",
    "0002_add_fts"}` → append `"0003_normalize_tags"`.
  - `TestFTS_MigrationBackfillsExistingRows` — final `count != 2` /
    "want 2" → `3`.

**Cross-check delta (record it, per §9 audit-grep both-sides rule):**
the STAGE-006 notes estimate "~five assertion sites across three test
files" and name `migrate_test.go` among them. Running the grep shows
**`migrate_test.go` is NOT coupled to the real migration set** — its
`count == 2` / exact-list assertions
(`TestMigrate_AppliesInOrder`/`_Idempotent`/`_FailedMigrationRollsBack`)
run against in-test `fstest.MapFS` fixtures (`0001_a`/`0002_b`), not the
embedded `migrationsFS`, so they **stay** untouched. Actual coupling =
**4 assertion sites across 2 files** (store_test.go, fts_test.go). Build
should re-run these greps (build-side cross-check) and treat any further
delta as a question, not a silent change.

**2. Inversion / removal** — (a) the `list --tag` sentinel-comma `LIKE`
becomes a join `EXISTS`; (b) `entries.tags` is dropped; (c) the FTS
external-content table + 3 entries-row triggers are replaced. Greps run:
`grep -rn "func Test.*[Tt]ag" internal/...`, and
`grep -rn "INTO entries" internal --include=*_test.go | grep -i tags`.

- **Genuine planned rewrites** (premise invalidated — do at build, not
  as Deviations):
  - `internal/storage/store_test.go` `TestList_TagFilterNullAndEmpty` —
    inserts a NULL-tags row via **raw SQL naming the `tags` column**
    (`INSERT INTO entries (title, description, tags, ...)`). After
    `0003` drops the column this statement fails to compile. **Rewrite**
    the raw INSERT to omit `tags` (a row with no `taggings` *is* the
    "no tags" case); the assertion (`Tag:"auth"` → 0 rows, covering both
    the empty-via-`Add` row and the no-tags raw row) is **preserved**.
  - `internal/storage/fts_test.go` `TestFTS_VirtualTableShape` — asserts
    the DDL contains `content='entries'` and `content_rowid='id'` and no
    `tokenize`. Under the regular (own-content) FTS5 table those
    external-content markers are gone. **Rewrite** to assert the new
    shape: the five columns (`title, description, tags, project,
    impact`) present, default tokenizer (no `tokenize=` clause), and
    that it is no longer external-content.
  - `internal/storage/fts_test.go` `TestFTS_TriggersExistAfterMigration`
    — asserts **exactly** `[entries_ad, entries_ai, entries_au]`.
    **Rewrite** to the new six-trigger set:
    `[entries_ad, entries_ai, entries_au, taggings_ad, taggings_ai,
    tags_au]`.
- **Behavioral tests that STAY GREEN** (premise *preserved* — the
  mechanism changes but the observable contract does not; re-verify at
  build, update only stale mechanism comments):
  - `store_test.go`: `TestList_FilterByTag`, `TestList_TagFilterNoFalsePositive`,
    `TestList_FilterCombined`, `TestList_FilterPreservesOrder`,
    `TestAdd_PersistsAllFields`, `TestUpdate_ReplacesUserEditableFields`,
    `TestUpdate_ReturnsHydratedEntry`.
  - `get_test.go`: `TestGet_RoundTripsAllFields`,
    `TestGet_PartiallyEmptyFieldsHydrateAsEmptyStrings`.
  - `fts_test.go`: `TestFTS_TriggerInsertAddsToIndex`,
    `TestFTS_TriggerUpdateReplacesIndexedRow`,
    `TestFTS_TriggerDeleteRemovesFromIndex`,
    `TestFTS_MatchQueryReturnsExpectedIds`,
    `TestFTS_UnicodeTokenizerSplitsOnPunctuation`,
    `TestFTS_MigrationBackfillsExistingRows` (besides its count-bump;
    its pre-`Open` raw INSERT names `tags` but runs against the post-`0001`
    schema where the column still exists, so it **stays**).
  - `cli/list_test.go`: `TestListCmd_FilterByTag`,
    `TestListCmd_TagFilterNoFalsePositive`.
  - (All `internal/export/*` and `internal/aggregate/*` tag tests
    operate on in-memory `[]storage.Entry` literals, never round-trip
    through the DB, and are **byte-untouched** by construction.)

**3. Status change** — the migration removes a column, realizes the
"Tags normalization" future-work, and changes the FTS topology. Grep
run: `grep -rn -i "tag" docs/ README.md`. Disposition of each
status-bearing hit:

- **Updates (this spec):**
  - `docs/data-model.md`:
    - `entries` table — remove the `tags` column row (line ~20); note it
      is superseded by the `tags`/`taggings` tables (DEC-015).
    - "Relationships: none ... Tag normalization ... deferred option"
      (line ~27) → now realized; document the polymorphic relationship.
    - `entries_fts` section (lines ~40–65) → regular FTS5 + the new
      six-trigger topology; the "comma-joined tags tokenize one-per-token"
      note becomes "the projected `tags` string tokenizes one-per-token."
    - "No index on `tags` ... migration to a `tags`/`entry_tags` join
      pair is the answer" (lines ~101–103) → realized as
      `tags`/`taggings` (polymorphic, not `entry_tags`).
    - "Future schema shapes → Tags normalization" (lines ~126–129) →
      strike / mark realized by DEC-015 (note polymorphic `taggings`
      shape, not `entry_tags`).
    - References (line ~144) → add DEC-015; note DEC-004 superseded.
    - **Add** a new `tags` + `taggings` entity section.
  - `docs/api-contract.md`:
    - `brag list` `--tag` note (lines ~113–114): "tags filter uses
      substring against the comma-joined column in MVP" is now stale →
      "filters on exact tag-token membership via the normalized
      `taggings` join (DEC-015); behavior is unchanged from the MVP
      sentinel-comma match." (No new commands here — `brag tags`/`tag
      rename`/`tag merge` are **SPEC-026**.)
  - `docs/architecture.md` (storage-behavior hits only; the broader
    diagram + `internal/projects` refresh stays for STAGE-008 per brief):
    - Mermaid `Migrations` node (lines ~46–47) lists `0001`/`0002` files
      → add `0003_normalize_tags.sql`.
    - `internal/storage` responsibility row (line ~74): the FTS sync
      description references the old `entries`-row triggers → update to
      "syncs `entries_fts` against `taggings`/`tags` mutations
      (DEC-015)."
    - References (line ~148): DEC-004 line → add DEC-015 / note
      superseded.
- **Stays (explicit, with reason):**
  - `docs/tutorial.md` — every tag mention describes the **user-facing**
    `add`/`list --tag`/`search`/export workflow, which is byte-stable.
    Lines ~80–81 ("No normalization, no tag registry. Future stages may
    add tag search...") become partly realized, but the *visible*
    workflow is unchanged and the taxonomy commands that warrant a
    tutorial update ship in **SPEC-026**. Stays for SPEC-025.
  - `README.md` line ~67 (`--tags auth,perf,backend`) — usage example,
    behavior unchanged. Stays.
  - `docs/brag-entry.schema.json` — the `brag add --json` input contract
    (tags as comma-joined string, DEC-012/DEC-004) is **unchanged**;
    DEC-011/DEC-012 still take/emit the string. Stays.
  - `docs/blog/**`, `docs/framework-feedback/**`, `docs/reports/**`,
    `docs/CONTEXTCORE_ALIGNMENT.md`, `docs/macos-notarization-checklist.md`,
    `docs/development.md` — historical/process/narrative; no shipped-
    behavior status claim about tag storage. Stay.

## Acceptance Criteria

- [ ] A new `internal/storage/migrations/0003_normalize_tags.sql` is
  applied automatically on `storage.Open`; `schema_migrations` lists
  exactly `0001_initial`, `0002_add_fts`, `0003_normalize_tags` and a
  re-`Open` is a no-op (count stays 3).
- [ ] After migration, `tags(id, name UNIQUE)` and `taggings(id, tag_id,
  taggable_type, taggable_id, position, UNIQUE(taggable_type,
  taggable_id, tag_id))` exist with indexes `idx_taggings_tag` and
  `idx_taggings_taggable`; only `taggable_type = 'entry'` rows exist.
- [ ] The ETL is **lossless** on a representative corpus: every distinct
  trimmed non-empty token across all entries becomes one `tags` row;
  every `(entry, distinct-trimmed-tag)` pair becomes one `taggings` row;
  whitespace is trimmed and within-entry duplicates collapse to first
  occurrence.
- [ ] `entries.tags` no longer exists after migration (`PRAGMA
  table_info(entries)` has no `tags`).
- [ ] `Entry.Tags` round-trips byte-identically for canonical inputs:
  `Add(Entry{Tags:"perf,auth"})` then `Get`/`List` returns
  `Tags == "perf,auth"` (insertion order preserved, **not** sorted);
  non-canonical `" auth , auth ,perf"` canonicalizes to `"auth,perf"`;
  empty/no-tags returns `""`.
- [ ] `Store.List(ListFilter{Tag:"auth"})` returns exactly the entries
  tagged `auth` (exact token; `"authoring"` excluded; empty/no-tags
  excluded), ordered `created_at DESC, id DESC` as before.
- [ ] `Store.Search` results are byte-identical to pre-migration on the
  same corpus (DEC-010); a tag added/renamed via `taggings`/`tags`
  mutation is reflected in subsequent `search`.
- [ ] `Add`/`Update`/`Delete` are transactional: a failure leaves
  neither a half-written entry nor orphaned/missing taggings; `Delete`
  removes the entry's taggings and its `entries_fts` row.
- [ ] `internal/cli/list.go` is unchanged (no SQL added to the CLI
  layer); `internal/export`/`internal/aggregate`/`internal/editor`
  output is byte-identical.
- [ ] `go test ./...`, `gofmt -l .`, `go vet ./...`, and
  `CGO_ENABLED=0 go build ./...` are clean.

## Failing Tests

Written at **design**; build makes them pass. Paths + assertions below.
New tests are additive; the rewrites/bumps are the premise-audit items
above (do those as planned, not as discoveries).

- **`internal/storage/migrate_test.go`** *(new tests; existing MapFS
  tests untouched)*
  - `"TestOpen_TagSchemaExists"` — after `Open`, asserts tables `tags`
    and `taggings` exist and indexes `idx_taggings_tag`,
    `idx_taggings_taggable` exist (via `sqlite_master`).
  - `"TestOpen_TagsColumnDropped"` — `PRAGMA table_info(entries)` has no
    row named `tags`.

- **`internal/storage/store_test.go`** *(new)*
  - `"TestMigrate_ETL_Lossless"` — the load-bearing lossless-ETL test.
    Manually apply only `0001` to a temp DB (mirror
    `TestFTS_MigrationBackfillsExistingRows`'s setup), seed a
    **representative corpus** via raw INSERT into the still-present
    `entries.tags` column:
    | id | tags input | expected projection |
    |----|------------|---------------------|
    | 1  | `auth,perf` | `auth,perf` |
    | 2  | `perf,auth` | `perf,auth` |
    | 3  | ` auth , auth ,perf` | `auth,perf` |
    | 4  | `` (empty) | `` |
    | 5  | NULL | `` |
    | 6  | `solo` | `solo` |
    Then `Open` (applies `0002` + `0003`). Assert:
    (a) `SELECT COUNT(*) FROM tags == 3` and the names are exactly
    `{auth, perf, solo}`;
    (b) `SELECT COUNT(*) FROM taggings == 7`;
    (c) for each id, `Store.Get(id).Tags` equals the expected projection
    (byte-for-byte, proving order preservation + trim + dedup + empty);
    (d) `SELECT COUNT(*) FROM taggings WHERE taggable_type <> 'entry' == 0`.
  - `"TestAdd_TagsWriteThroughJoin"` — `Add(Entry{Tags:"perf,auth"})`;
    assert (a) `Get` returns `Tags == "perf,auth"`; (b) two `taggings`
    rows exist for the entry with positions 0,1 mapping perf,auth; (c)
    `tags` has rows `perf`,`auth`.
  - `"TestAdd_TagsCanonicalizeTrimDedup"` —
    `Add(Entry{Tags:" auth , auth ,perf"})`; `Get().Tags == "auth,perf"`
    and exactly two taggings for the entry.
  - `"TestAdd_EmptyTagsNoTaggings"` — `Add(Entry{Title:"x"})` (no tags);
    `Get().Tags == ""` and zero taggings for the entry.
  - `"TestUpdate_TagsReplacedThroughJoin"` — `Add(Tags:"a,b")` then
    `Update(Tags:"b,c")`; `Get().Tags == "b,c"`, taggings for the entry
    are exactly `{b,c}` (old `a` membership gone), and tag `a` may remain
    orphaned in `tags` (assert it is NOT referenced by any taggings — GC
    is out of scope).
  - `"TestDelete_RemovesTaggings"` — `Add(Tags:"x,y")` then `Delete`;
    zero taggings for that entry and the `entries_fts` row is gone.
  - `"TestList_TagFilterThroughJoin"` — seed `auth,perf` / `perf,backend`
    / `auth`; `List(Tag:"auth")` → 2, `Tag:"perf"` → 2, `Tag:"backend"`
    → 1, `Tag:"nonesuch"` → 0 (mirrors `TestList_FilterByTag` but is
    explicit about the join path).
  - `TestList_TagFilterNullAndEmpty` — **rewrite** (premise audit): drop
    the raw `INSERT ... (tags) ... NULL`; insert a no-tags row via
    `Store.Add(Entry{Title:"nulltags"})` (or raw INSERT omitting the
    `tags` column). Same assertion: `Tag:"auth"` → 0.

- **`internal/storage/fts_test.go`**
  - `TestFTS_VirtualTableShape` — **rewrite** (regular FTS5 shape:
    columns present; no `content='entries'`/`content_rowid='id'`; no
    `tokenize=`).
  - `TestFTS_TriggersExistAfterMigration` — **rewrite** to the six-name
    set `[entries_ad, entries_ai, entries_au, tags_au, taggings_ad,
    taggings_ai]` (sorted).
  - `"TestFTS_ReSyncsOnTaggingMembership"` *(new)* — `Add(Title:"row")`
    (no tags), then add a tag via the Store write path (e.g. `Update` to
    `Tags:"newtag"`); `Store.Search("newtag")` returns the entry (proves
    `taggings_ai` synced FTS).
  - `"TestFTS_ReSyncsOnTagRename"` *(new)* — `Add(Tags:"auth,perf")`;
    rename the `auth` tag via raw `UPDATE tags SET name='authz' WHERE
    name='auth'` (SPEC-026 ships the command; here we drive the trigger
    directly); `Search("authz")` finds the entry and `Search` for the
    old token no longer finds it **by tag** (title-free fixture).
  - `"TestSearch_ByteStableAcrossMigration"` *(new, or fold into ETL
    test)* — build the representative corpus, capture `Search` results
    for a tag-only token pre-migration intent vs the post-migration
    `Store.Search` on the same logical corpus; assert identical id
    ordering for `perf` (the §12(b) pre-flight confirmed `[1 2 3]`).

- **`internal/storage/store_test.go` / `fts_test.go`** — the four
  count-bump assertions updated as enumerated in the premise audit.

> **Locked-decision ↔ test traceability (§9).** DEC-015's load-bearing
> sub-decisions each have a paired failing test: ETL losslessness →
> `TestMigrate_ETL_Lossless`; projection-as-byte-stable-string +
> insertion order → the projection column of that test +
> `TestAdd_TagsWriteThroughJoin`; trim/dedup → `TestAdd_TagsCanonicalizeTrimDedup`;
> polymorphic-only-`entry` → ETL test (d); FTS re-topology → the two
> `TestFTS_ReSyncs*` tests + the trigger-list rewrite; column drop →
> `TestOpen_TagsColumnDropped`; `list --tag` join → `TestList_TagFilterThroughJoin`.

## Implementation Context

*Read this section and the files it points to before build. This is the
whole handoff — the build session has only this spec.*

### Decisions that apply

- **`DEC-015`** (emitted by this spec) — the polymorphic `tags` +
  `taggings(taggable_type, taggable_id, position)` schema, the lossless
  ETL contract, and the **`Entry.Tags`-as-projection** rule
  (`GROUP_CONCAT(name ORDER BY position)`, insertion order preserved).
  This is the spec's primary contract — read it in full.
- **`DEC-004`** (superseded by DEC-015) — the comma-joined MVP shape you
  are migrating away from; its `superseded_by` is already set. Read it to
  understand the sentinel-comma `LIKE` (`',' || tags || ',' LIKE
  '%,<tag>,%'`) you are replacing and the `"auth"`-vs-`"authoring"` edge
  case that the join must keep handling correctly.
- **`DEC-002`** — embedded forward-only migrations; **no down-migration**.
  Each `.sql` runs as one multi-statement `ExecContext` inside one
  transaction (`migrate.go runMigration`). Do **not** put `BEGIN`/`COMMIT`
  in the file. Once merged, the file is immutable
  (`migrations-are-append-only`); fix forward only.
- **`DEC-010`** — `brag search` query syntax must stay byte-stable.
  `entries_fts` keeps the same five columns and the **default unicode61
  tokenizer** (no `tokenize=` clause), so MATCH/rank semantics are
  identical. The only change is *what fires the sync* (now
  `taggings`/`tags`, plus the non-tag `entries` columns).
- **`DEC-011`** — JSON output: `tags` is a comma-joined **string** key.
  Untouched because the Store hands `export.ToJSON` the projected string.
- **`DEC-013`** — markdown export shape; consumes `Entry.Tags` string.
  Untouched.
- **`DEC-014`** — stats/summary/review shapes; `stats.go extractTags`
  splits `Entry.Tags` on `,`. Untouched — keep the projection a
  comma-joined string.
- **`DEC-005`** — INTEGER AUTOINCREMENT PKs; `tags.id`/`taggings.id`
  follow it; `id DESC` stays the deterministic tie-break.

### Constraints that apply

- `migrations-are-append-only` (blocking) — `0003_*` is a brand-new
  file; never edit `0001`/`0002`.
- `storage-tests-use-tempdir` (blocking) — all new storage tests use
  `t.TempDir()`; never touch `~/.bragfile`.
- `no-sql-in-cli-layer` (blocking) — the join lives entirely in
  `internal/storage`; `internal/cli/list.go` only sets `ListFilter.Tag`.
- `timestamps-in-utc-rfc3339` (blocking) — unchanged; entries timestamps
  stay Go-written RFC3339 UTC. (Tags carry no timestamps.)
- `test-before-implementation` (blocking) — the Failing Tests above are
  the design deliverable; build makes them pass.
- `errors-wrap-with-context` (warning) — wrap new storage errors
  (`fmt.Errorf("add entry: %w", err)` etc.); the transactional paths
  must wrap begin/exec/commit failures.
- `no-cgo` (blocking) — pure-Go path only; no new dependency
  (`no-new-top-level-deps-without-decision`).
- `one-spec-per-pr` (blocking) — this PR references SPEC-025 only.

### Design-time pre-flight (§12(b)) — already done, results below

Per AGENTS.md §12(b) and the STAGE-006 WATCH-list (§12(a) third
confirming case — **note, do not codify**), the embedded migration SQL,
the ETL row counts, and the read/filter SQL fragments were **run at
design** against the real driver (`modernc.org/sqlite v1.51.0`, SQLite
**3.53.1**), including the runner-faithful path (one multi-statement
`ExecContext` inside one `BeginTx` transaction). Confirmed:

- `GROUP_CONCAT(name, ',' ORDER BY position)` (in-aggregate `ORDER BY`)
  is supported (SQLite ≥ 3.44) and preserves insertion order.
- The full `0003` script runs as a single transactional multi-statement
  exec — including `DROP TABLE entries_fts`, `ALTER TABLE entries DROP
  COLUMN tags`, and all trigger creation — with no statement-splitter
  issues.
- On the representative corpus: **3 tags / 7 taggings**, projection
  exactly as the ETL-test table above, `entries.tags` dropped, `Search`
  byte-stable (`MATCH perf → [1 2 3]`), and membership-add / rename /
  delete all propagate to `entries_fts`.

Build does **not** need to re-derive this, but should re-run the
premise-audit greps (build-side cross-check, §9).

### Dev/prod DB isolation (PROJ-002 brief) — mandatory this stage

This is where the schema first changes. While v0.2.x is in flight:

- Build/run the dev binary against a **dev DB**:
  `BRAGFILE_DB=~/.bragfile-dev/db.sqlite` (or `--db`), via `just install`
  → `~/go/bin/brag`. **Never open the production
  `~/.bragfile/db.sqlite`** with a v0.2.x-format (post-`0003`) binary —
  the migration is forward-only and would irreversibly rewrite it.
- Production stays brew-installed at v0.1.0. The deliberate, documented
  upgrade (backup `cp` + `brew upgrade`) is STAGE-008, not this spec.
- All tests use `t.TempDir()` regardless — they never touch either real
  DB.

### Out of scope (for this spec specifically)

- **`brag tags` / `brag tag rename` / `brag tag merge`** → **SPEC-026**.
  The schema + the `tags_au` rename trigger are built here so the
  topology is complete and correct, but no command, no CLI surface, and
  no rename/merge *flow* ships in SPEC-025.
- **Orphan-tag garbage collection.** `Update` may leave a `tags` row with
  zero `taggings`. Harmless for reads; cleanup is SPEC-026's concern.
- **Projects as a taggable type** (writing `taggable_type='project'`
  rows) → STAGE-007. The schema is polymorphic now; no project rows are
  written here.
- **Any third taggable type** (`goals`, etc.) — paper-sketched in
  DEC-015 only; no code.
- **`docs/architecture.md` diagram refresh beyond the storage-behavior
  hits** and the `internal/projects` package → STAGE-008.
- **CHANGELOG `[0.2.0]`, RC-tag cut, migration-prompt safety belt** →
  STAGE-008.
- **Reverse/down migration** — forward-only (DEC-002); the `cp`-backup
  downgrade story is STAGE-008 docs.

## Notes for the Implementer

### The migration literal — transcribe verbatim

Create `internal/storage/migrations/0003_normalize_tags.sql` with
**exactly** this content (validated at design through the runner-faithful
path; do not paraphrase the SQL):

```sql
-- 0003_normalize_tags.sql — SPEC-025 (PROJ-002 / STAGE-006)
-- Single atomic in-place tag normalization (DEC-015, supersedes DEC-004).
-- Creates the polymorphic tags taxonomy, backfills it losslessly from the
-- comma-joined entries.tags column, re-points FTS sync onto the join, and
-- drops the legacy entries.tags column — all forward-only (DEC-002).
-- Runs inside the migration runner's per-migration transaction; do NOT
-- add BEGIN/COMMIT here. Validated at design against modernc.org/sqlite
-- 1.51.0 (SQLite 3.53.1): 3 tags / 7 taggings on the representative
-- corpus, search byte-stable, entries.tags dropped, FTS re-sync correct.

-- 1. Normalized taxonomy + polymorphic membership join.
CREATE TABLE tags (
    id   INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE taggings (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    tag_id        INTEGER NOT NULL REFERENCES tags(id),
    taggable_type TEXT    NOT NULL,
    taggable_id   INTEGER NOT NULL,
    position      INTEGER NOT NULL,
    UNIQUE (taggable_type, taggable_id, tag_id)
);

CREATE INDEX idx_taggings_tag      ON taggings(tag_id);
CREATE INDEX idx_taggings_taggable ON taggings(taggable_type, taggable_id);

-- 2. Lossless ETL from entries.tags: split on ',', trim each token, drop
--    empties, de-duplicate within an entry keeping first-occurrence
--    position. Upsert tag names, then write 'entry' memberships.
WITH RECURSIVE split(id, tok, rest, pos) AS (
    SELECT id, '', tags || ',', 0
      FROM entries
     WHERE tags IS NOT NULL AND tags <> ''
    UNION ALL
    SELECT id,
           substr(rest, 1, instr(rest, ',') - 1),
           substr(rest, instr(rest, ',') + 1),
           pos + 1
      FROM split
     WHERE rest <> ''
),
tokens AS (
    SELECT id AS entry_id, trim(tok) AS name, pos
      FROM split
     WHERE trim(tok) <> ''
)
INSERT OR IGNORE INTO tags(name)
SELECT DISTINCT name FROM tokens;

WITH RECURSIVE split(id, tok, rest, pos) AS (
    SELECT id, '', tags || ',', 0
      FROM entries
     WHERE tags IS NOT NULL AND tags <> ''
    UNION ALL
    SELECT id,
           substr(rest, 1, instr(rest, ',') - 1),
           substr(rest, instr(rest, ',') + 1),
           pos + 1
      FROM split
     WHERE rest <> ''
),
tokens AS (
    SELECT id AS entry_id, trim(tok) AS name, pos
      FROM split
     WHERE trim(tok) <> ''
),
firsts AS (
    SELECT entry_id, name, MIN(pos) AS position
      FROM tokens
     GROUP BY entry_id, name
)
INSERT OR IGNORE INTO taggings(tag_id, taggable_type, taggable_id, position)
SELECT t.id, 'entry', f.entry_id, f.position
  FROM firsts f
  JOIN tags t ON t.name = f.name;

-- 3. Re-point FTS. The old entries-row triggers copy new.tags/old.tags,
--    which is about to disappear; drop them, swap entries_fts from
--    external-content to a regular (own-content) FTS5 table, and backfill
--    its tags column from the join projection. Search shape (columns +
--    default unicode61 tokenizer) is unchanged, so DEC-010 holds.
DROP TRIGGER entries_ai;
DROP TRIGGER entries_au;
DROP TRIGGER entries_ad;
DROP TABLE entries_fts;

CREATE VIRTUAL TABLE entries_fts USING fts5(
    title, description, tags, project, impact
);

INSERT INTO entries_fts(rowid, title, description, tags, project, impact)
SELECT e.id, e.title, e.description,
       COALESCE((SELECT GROUP_CONCAT(t.name, ',' ORDER BY tg.position)
                   FROM taggings tg
                   JOIN tags t ON t.id = tg.tag_id
                  WHERE tg.taggable_type = 'entry' AND tg.taggable_id = e.id), ''),
       e.project, e.impact
  FROM entries e;

-- 4. Drop the legacy shadow column (SQLite 3.35+ ALTER ... DROP COLUMN;
--    safe now that no trigger or index references it).
ALTER TABLE entries DROP COLUMN tags;

-- 5. New trigger topology. entries triggers maintain the non-tag columns
--    (and seed an empty tags cell on insert); taggings/tags triggers
--    recompute the affected entry's tags projection.
CREATE TRIGGER entries_ai AFTER INSERT ON entries BEGIN
    INSERT INTO entries_fts(rowid, title, description, tags, project, impact)
    VALUES (new.id, new.title, new.description,
            COALESCE((SELECT GROUP_CONCAT(t.name, ',' ORDER BY tg.position)
                        FROM taggings tg JOIN tags t ON t.id = tg.tag_id
                       WHERE tg.taggable_type = 'entry' AND tg.taggable_id = new.id), ''),
            new.project, new.impact);
END;

CREATE TRIGGER entries_au AFTER UPDATE ON entries BEGIN
    UPDATE entries_fts
       SET title = new.title, description = new.description,
           project = new.project, impact = new.impact
     WHERE rowid = new.id;
END;

CREATE TRIGGER entries_ad AFTER DELETE ON entries BEGIN
    DELETE FROM entries_fts WHERE rowid = old.id;
END;

CREATE TRIGGER taggings_ai AFTER INSERT ON taggings
WHEN new.taggable_type = 'entry' BEGIN
    UPDATE entries_fts
       SET tags = COALESCE((SELECT GROUP_CONCAT(t.name, ',' ORDER BY tg.position)
                              FROM taggings tg JOIN tags t ON t.id = tg.tag_id
                             WHERE tg.taggable_type = 'entry'
                               AND tg.taggable_id = new.taggable_id), '')
     WHERE rowid = new.taggable_id;
END;

CREATE TRIGGER taggings_ad AFTER DELETE ON taggings
WHEN old.taggable_type = 'entry' BEGIN
    UPDATE entries_fts
       SET tags = COALESCE((SELECT GROUP_CONCAT(t.name, ',' ORDER BY tg.position)
                              FROM taggings tg JOIN tags t ON t.id = tg.tag_id
                             WHERE tg.taggable_type = 'entry'
                               AND tg.taggable_id = old.taggable_id), '')
     WHERE rowid = old.taggable_id;
END;

CREATE TRIGGER tags_au AFTER UPDATE ON tags BEGIN
    UPDATE entries_fts
       SET tags = COALESCE((SELECT GROUP_CONCAT(t.name, ',' ORDER BY tg.position)
                              FROM taggings tg JOIN tags t ON t.id = tg.tag_id
                             WHERE tg.taggable_type = 'entry'
                               AND tg.taggable_id = entries_fts.rowid), '')
     WHERE rowid IN (SELECT taggable_id FROM taggings
                      WHERE taggable_type = 'entry' AND tag_id = new.id);
END;
```

### Store read path — the projection fragment (reuse in Get/List/Search)

Replace the `tags` value in the `SELECT` lists of `Get`, `List`, and
`Search` with this correlated scalar subquery (it needs no `GROUP BY`,
so the existing query shapes — including `Search`'s `ORDER BY rank, e.id
DESC` — are otherwise untouched). Validated at design.

```sql
COALESCE((
    SELECT GROUP_CONCAT(t.name, ',' ORDER BY tg.position)
      FROM taggings tg
      JOIN tags t ON t.id = tg.tag_id
     WHERE tg.taggable_type = 'entry' AND tg.taggable_id = e.id
), '') AS tags
```

The scan stays the same shape (tags is never NULL now — `COALESCE` →
`''`); keep scanning into the existing `tags` variable. The full read
column list becomes:
`e.id, e.title, e.description, <projection> AS tags, e.project, e.type, e.impact, e.created_at, e.updated_at`.
(`type` is **not** indexed by FTS and is not a tag — leave it as the
plain `entries.type` column.)

### Store `list --tag` filter — replace the LIKE with EXISTS

In `Store.List`, swap the `DEC-004` sentinel-comma clause:

```go
// OLD (remove):
conds = append(conds, "',' || tags || ',' LIKE ?")
args  = append(args, "%,"+f.Tag+",%")
// NEW:
conds = append(conds, `EXISTS (SELECT 1 FROM taggings tg
    JOIN tags t ON t.id = tg.tag_id
   WHERE tg.taggable_type = 'entry' AND tg.taggable_id = e.id
     AND t.name = ?)`)
args = append(args, f.Tag)
```

Note this requires the `FROM entries` to be aliased `e` (the projection
subquery already references `e.id`); alias the `List` base query
`FROM entries e` and qualify the other filter columns (`e.project`,
`e.type`, `e.created_at`). Exact `t.name = ?` is what preserves the
`"auth"`-not-`"authoring"` guarantee DEC-004's sentinel commas gave.

### Store write path — transactional dual nothing (single source now)

`entries.tags` is **gone**, so there is no shadow to write — the join is
the single source of truth. Make `Add`/`Update`/`Delete` transactional:

- **Add:** in a tx — `INSERT INTO entries (title, description, project,
  type, impact, created_at, updated_at)` (no `tags` column); take
  `LastInsertId`; canonicalize `e.Tags` (split on `,`, `TrimSpace` each,
  drop empties, dedup keeping first occurrence); for each token in order
  `INSERT OR IGNORE INTO tags(name)`, look up `tag_id`, then `INSERT INTO
  taggings(tag_id, 'entry', id, position)` with `position` = its index.
  Commit. Return `e` with `e.Tags` set to the canonical
  `strings.Join(tokens, ",")` so the returned struct matches what reads
  will project. (A Go-side splitter mirroring `stats.go extractTags`'s
  trim/drop-empty, plus dedup, is the clean shape — but tags is written
  to the **join**, not a column.)
- **Update:** in a tx — `UPDATE entries SET title=…, description=…,
  project=…, type=…, impact=…, updated_at=…` (no `tags`); `DELETE FROM
  taggings WHERE taggable_type='entry' AND taggable_id=id`; re-insert the
  canonical tokens exactly as in Add. The `entries_au` trigger refreshes
  the FTS non-tag columns; the taggings delete/insert refresh FTS tags.
  Commit; return `Get(id)`.
- **Delete:** in a tx — `DELETE FROM taggings WHERE taggable_type='entry'
  AND taggable_id=id`; `DELETE FROM entries WHERE id=?`; check
  `RowsAffected` on the entries delete for `ErrNotFound`. Commit. (The
  `entries_ad` trigger removes the FTS row; deleting taggings first keeps
  the FTS tag-resync harmless.)

Keep all the existing error wrapping; wrap tx begin/exec/commit failures.

### Gotchas

- **Do not add `BEGIN`/`COMMIT` to the migration file** — the runner
  already wraps it (`migrate.go runMigration`). It is one
  `tx.ExecContext` of the whole file.
- **`entries_fts` is now a regular (own-content) FTS5 table.** That means
  ordinary `INSERT … (rowid, …)`, `UPDATE … WHERE rowid=?`, and `DELETE …
  WHERE rowid=?` all work in triggers — no external-content `'delete'`
  command. Search queries keep joining `entries e ON e.id =
  entries_fts.rowid`.
- **`migrate_test.go` MapFS tests are not coupled** to the real
  migration set — don't "fix" their `count == 2` (it refers to their own
  2-file fixtures). Only the four real-FS assertions (store_test.go,
  fts_test.go) bump.
- **Order matters in the migration:** create tags/taggings → ETL (reads
  `entries.tags`, still present) → drop old triggers → drop+recreate FTS
  → backfill FTS from the join → drop the column → create new triggers.
  The column must still exist during the ETL and the FTS backfill, and be
  dropped before the new triggers (which never reference it) are created.
- **`position` base differs by write path — harmless, but choose
  consciously.** The ETL's recursive-split CTE starts its empty sentinel
  at `pos=0`, so the first real token lands at `position=1` (1-based),
  whereas the `Add`/`Update` Go path writes `position` = the 0-based
  token index. `position` is only ever an `ORDER BY` key, so projections
  are byte-identical either way and no test asserts an absolute ETL
  position — but do not write a test that compares absolute positions
  *across* the two paths, and prefer making the ETL 0-based (e.g. start
  the recursion so the first token is `position=0`) if a uniform base is
  cheap, so a migrated entry and a freshly-added one are indistinguishable.
- **Literal-artifact discipline (§12 / §12(b)):** the migration SQL and
  the read/filter fragments above are the validated literals — build
  transcribes; verify diffs against them. If the build needs to change
  the SQL, re-run the design-time pre-flight before deviating.

---

## Build Completion

*Filled in at the end of the **build** cycle, before advancing to verify.*

- **Branch:** `feat/spec-025-normalize-tag-storage`
- **PR (if applicable):** TBD (opened after verify)
- **All acceptance criteria met?** yes
- **New decisions emitted:**
  - none beyond DEC-015, which was emitted at design
- **Deviations from spec:**
  - Trigger sorted-order in `TestFTS_TriggersExistAfterMigration`: spec listed `tags_au` before `taggings_*` but lexicographic order is `taggings_*` < `tags_*` (at byte 3, 'g' < 's'). Fixed the test `want` slice — not a deviation in the migration itself, only a test fixture correction.
- **Follow-up work identified:**
  - none

### Build-phase reflection (3 questions, short answers)

Process-focused: how did the build go? What friction did the spec create?

1. **What was unclear in the spec that slowed you down?**
   — The trigger-name sorted order in §11 was listed incorrectly (`tags_au, taggings_ad, taggings_ai`). This created a test failure that needed investigation before the fix was trivial. Everything else in the spec was precise enough to implement directly.

2. **Was there a constraint or decision that should have been listed but wasn't?**
   — The position-base gotcha (ETL is 1-based because the sentinel row starts at 0; Go write path is 0-based) was documented in the spec header but not echoed in the taggings INSERT helper section. A brief note there would have made the `insertTaggings` implementation self-contained to read.

3. **If you did this task again, what would you do differently?**
   — Run the migration pre-flight against the actual test corpus _before_ writing the trigger-exists test, so the expected sorted list is derived from observation rather than typed by hand. The spec's §12(b) pre-flight is exactly the right tool; the trigger test should cite its output directly.

---

## Reflection (Ship)

*Appended during the **ship** cycle. Outcome-focused reflection, distinct
from the process-focused build reflection above.*

1. **What would I do differently next time?**
   — <answer>

2. **Does any template, constraint, or decision need updating?**
   — <answer>

3. **Is there a follow-up spec I should write now before I forget?**
   — <answer>
