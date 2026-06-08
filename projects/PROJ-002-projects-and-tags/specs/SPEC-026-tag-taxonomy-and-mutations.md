---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-026
  type: story                      # epic | story | task | bug | chore
  cycle: build                     # frame | design | build | verify | ship
  blocked: false
  priority: medium
  complexity: M                    # S | M | L  (M — L considered and rejected, see Context)

project:
  id: PROJ-002
  stage: STAGE-006
repo:
  id: bragfile

agents:
  architect: claude-opus-4-8
  implementer: claude-opus-4-8     # usually same Claude, different session
  created_at: 2026-06-07

references:
  decisions: [DEC-016, DEC-015, DEC-011, DEC-013, DEC-014, DEC-010, DEC-006, DEC-007, DEC-005]
  constraints:
    - no-sql-in-cli-layer
    - storage-tests-use-tempdir
    - errors-wrap-with-context
    - stdout-is-for-data-stderr-is-for-humans
    - test-before-implementation
    - one-spec-per-pr
    - no-cgo
    - no-new-top-level-deps-without-decision
  related_specs: [SPEC-025, SPEC-007, SPEC-020]
---

# SPEC-026: Tag taxonomy and mutations

## Context

This is the **second and final** spec of STAGE-006 (Tag normalization),
the foundation stage of PROJ-002. SPEC-025 (shipped, PR #36) laid down
the normalized `tags` + polymorphic `taggings` schema, the lossless ETL,
the Store read/write cutover, and the six-trigger FTS topology — but
deliberately shipped **no command surface**. SPEC-026 adds the
user-visible payoff that makes the migration worthwhile rather than a
silent internal refactor:

- `brag tags` — a taxonomy view: every in-use tag with its usage count.
- `brag tag rename <old> <new>` — rename a tag everywhere at once.
- `brag tag merge <src> <dst>` — fold one tag into another, de-duping.

It **emits DEC-016** (created at this design) recording the durable
behavior choices layered on DEC-015's schema: the `brag tags` shape +
sort, rename-into-existing semantics, the merge de-dup mechanism, and
the orphan-tag policy. DEC-016 does not supersede DEC-015 — it is the
behavior on top of that structure.

**Complexity — M (L considered and rejected).** The STAGE-006 framing
flagged a watch: if `brag tags` (read) + rename + merge + orphan GC
together read L at design, split the read command from the two
mutations. They do **not** read L, for three reasons: (1) there is **no
migration** — the load-bearing FTS gotcha is resolved entirely with
DELETE+INSERT against the *existing* SPEC-025 triggers, so none of the
migration/count-bump/§12(b)-migration cost applies; (2) orphan GC
collapses to *zero extra code* — `brag tags` counts through the join, so
orphans are simply invisible (DEC-016 choice 4); (3) rename is a
single-statement `UPDATE` and merge is one pre-flighted three-statement
transaction. The whole spec is three thin Store methods + three thin
cobra commands + one JSON helper + a DEC + a doc sweep. That is a
coherent **M**. The split fallback is recorded here but not taken.

- Parent stage: `STAGE-006-tag-normalization.md` (this is the second of
  two specs; SPEC-025 shipped first).
- Project: `PROJ-002` brief (dev/prod DB isolation governs this stage —
  see Implementation Context).
- Builds on `DEC-015` (the schema); emits `DEC-016` (the mutation
  semantics).

## Goal

Add the tag-taxonomy command surface on top of DEC-015's normalized
schema — `brag tags` (counted taxonomy view), `brag tag rename`, and
`brag tag merge` (de-duplicating) — implemented as new `Store` methods
behind thin CLI commands, with **no schema change and no migration**:
merge folds memberships via DELETE+INSERT so the existing SPEC-025 FTS
triggers keep `entries_fts` correct, and rename rides the existing
`tags_au` trigger.

## Inputs

- **Files to read:**
  - `internal/storage/store.go` — the SPEC-025 transactional patterns
    (`Add`/`Update`/`Delete`), `canonicalizeTags`, `insertTaggings`, and
    the `tagsProjection` fragment; new methods follow these shapes.
  - `internal/storage/entry.go` — `Entry` / `ListFilter` (unchanged); the
    new `TagCount` type lives near here or in `store.go`.
  - `internal/storage/errors.go` — `ErrNotFound`; add `ErrTagNotFound`
    and `ErrTagExists` sentinels.
  - `internal/storage/migrations/0003_normalize_tags.sql` — the live
    schema + the SIX triggers (`entries_ai/au/ad`, `taggings_ai/ad`,
    `tags_au`). **No `taggings_au` exists** — the merge gotcha. Read it.
  - `internal/cli/stats.go`, `internal/cli/list.go`, `internal/cli/delete.go`
    — command conventions: `RunE`, `getFlagString`, `config.ResolveDBPath`,
    `UserErrorf`, inline positional-arg validation (DEC-007), `--format`
    handling, stdout-vs-stderr discipline.
  - `internal/cli/errors.go` — `UserErrorf` / `ErrUser`.
  - `internal/export/json.go` — `ToJSON` (the DEC-011 naked-array
    marshal); the new `ToTagsJSON` helper sits beside it.
  - `internal/export/stats.go` — the `{tag, count}` JSON element shape
    (DEC-014 `top_tags`) the `brag tags` JSON mirrors; `extractTags`
    (unchanged).
  - `cmd/brag/main.go` — command registration (`root.AddCommand(...)`).
  - `decisions/DEC-016-*` (this spec's contract — read in full),
    `DEC-015`, `DEC-011`, `DEC-013`, `DEC-014`, `DEC-010`, `DEC-006`,
    `DEC-007`.
- **External APIs:** none. Plain `database/sql` against
  `modernc.org/sqlite`; no new dependency (would need a DEC under
  `no-new-top-level-deps-without-decision`).
- **Related code paths:** `internal/storage/` (the three new methods),
  `internal/cli/` (the three new commands), `internal/export/` (one JSON
  helper), `cmd/brag/main.go` (registration).

## Outputs

- **Files created:**
  - `decisions/DEC-016-tag-mutation-semantics.md` — **created at this
    design**. Records the four locked choices (see Implementation
    Context → Decisions).
  - `internal/cli/tags.go` — `NewTagsCmd()` (`brag tags`).
  - `internal/cli/tag.go` — `NewTagCmd()` (the `brag tag` parent) with
    `rename` and `merge` subcommands.
  - `internal/cli/tags_test.go`, `internal/cli/tag_test.go` — CLI tests.
  - `internal/storage/tags_test.go` — Store-method tests (kept separate
    from `store_test.go` for tidiness; `t.TempDir()` per constraint).
- **Files modified:**
  - `internal/storage/store.go` — add `TagCount` type and three methods:
    `TagCounts()`, `RenameTag(old, new)`, `MergeTags(src, dst)`.
  - `internal/storage/errors.go` — add `ErrTagNotFound`, `ErrTagExists`.
  - `internal/export/json.go` — add `ToTagsJSON([]storage.TagCount)`.
  - `cmd/brag/main.go` — register `NewTagsCmd()` and `NewTagCmd()`.
  - `docs/api-contract.md`, `docs/tutorial.md`, `docs/data-model.md`,
    `docs/architecture.md` — status-change doc sweep (enumerated below).
- **New exports:**
  - `storage.TagCount struct { Name string; Count int }`
  - `func (s *Store) TagCounts() ([]TagCount, error)`
  - `func (s *Store) RenameTag(oldName, newName string) error`
  - `func (s *Store) MergeTags(src, dst string) error`
  - `storage.ErrTagNotFound`, `storage.ErrTagExists`
  - `func cli.NewTagsCmd() *cobra.Command`,
    `func cli.NewTagCmd() *cobra.Command`
  - `func export.ToTagsJSON(tags []storage.TagCount) ([]byte, error)`
- **Database changes:** **NONE.** No migration. The FTS gotcha is
  resolved against the existing SPEC-025 triggers (DEC-016 choice 3).
  This is a load-bearing property of the spec, not an omission.

### Premise audit (`projects/_templates/premise-audit.md`), run at design

Greps were **run** at design and reconciled against the lists below.

```
- [x] Inversion/removal: greps run — NONE (purely additive commands)
- [x] Addition/count-bump: greps run — NONE (no migration; no count-coupled assertion)
- [x] Status-change: greps run, every doc hit listed as updates/stays
- [x] Cross-check: actual grep hits reconciled against the lists above
```

**1. Inversion / removal — NONE.** `brag tags` / `tag rename` / `tag
merge` are brand-new commands; no existing behavior is inverted, no flag
or column is removed. Greps run:
`grep -rn "func Test.*[Tt]ag" internal/...` surfaces only the SPEC-007/
SPEC-025 `list --tag` filter tests (`TestList_FilterByTag`,
`TestList_TagFilterNoFalsePositive`, `TestList_TagFilterNullAndEmpty`,
`TestListCmd_FilterByTag`, …) and the FTS tag tests — all unchanged in
premise (the `list --tag` join path SPEC-025 shipped is untouched). No
planned rewrites or deletions.

**2. Addition / count-bump — NONE.** No migration is added, so
`schema_migrations` is untouched (the SPEC-025 count of 3 stands). Greps
run: `grep -rn "Commands()\|len(.*Commands" internal/cli cmd` and
`cat internal/cli/root_test.go` — **no test enumerates or counts the
root subcommand set**; every CLI test builds its own root with only the
subcommands it needs (`newStatsTestRoot` pattern), so registering
`tags`/`tag` in `cmd/brag/main.go` couples to no assertion. No bumps.

**3. Status change — the new commands.** Grep run:
`grep -rn -i "tag" docs/ README.md`. Disposition of each status-bearing
hit:

- **Updates (this spec):**
  - `docs/api-contract.md`:
    - **Add** three new command sections — `### brag tags`,
      `### brag tag rename <old> <new>`, `### brag tag merge <src> <dst>`
      — placed after the `### brag stats` section, before
      `### brag completion`. (See ## Notes for the exact literal.)
    - References list (line ~444): **add** a `DEC-016` row; the existing
      `list --tag` line (~113–115, "exact tag-name membership via the
      normalized `taggings` join — DEC-015 / SPEC-025") **stays** (SPEC-025
      already corrected it).
  - `docs/tutorial.md`:
    - Lines ~80–82 ("No normalization, no tag registry. Future stages
      may add tag search and rename.") — the future-work claim is now
      **realized**. Update to point at `brag tags` / `tag rename` /
      `tag merge` (no longer "future"; tags are still a comma-joined
      *string at the input boundary*, which stays true — only the
      "no registry / no rename" half is struck).
    - **Add** a short "Tag taxonomy" subsection in §3 (near the
      `list --tag` material, ~line 175, or after `brag stats` ~line 405)
      showing `brag tags`, `brag tag rename`, `brag tag merge` with the
      one-line semantics (rename errors into an existing name; merge
      de-dups; counts are in-use only). House style: example block +
      bullet notes, mirroring the existing `list`/`stats` subsections.
  - `docs/data-model.md`:
    - References list (line ~167–168): **add** a `DEC-016` row
      (mutation semantics) beside the `DEC-015` row.
    - Under the `taggings` entity (~line 53) **add** a one-line note:
      tag rename/merge mutate this join per DEC-016 (rename via
      `tags_au`; merge via DELETE+INSERT firing `taggings_ad`/`_ai`).
      The schema tables themselves are **unchanged** (no migration).
  - `docs/architecture.md`:
    - `internal/storage` responsibility row (line ~74): the method list
      (`Add, List, Get, Update, Delete, Search`) **gains**
      `TagCounts, RenameTag, MergeTags`. (Minimal, parallel to SPEC-025's
      edit of the same row; the broader diagram/`internal/projects`
      refresh stays for STAGE-008 per the brief.)
    - References (line ~148–150): **add** a `DEC-016` line.
- **Stays (explicit, with reason):**
  - `README.md` line ~67 (`--tags auth,perf,backend`) — `add` usage
    example, behavior unchanged. Stays.
  - `docs/api-contract.md` lines ~96–98, `docs/tutorial.md` ~134–137,
    `docs/data-model.md` `DEC-012` line, `docs/brag-entry.schema.json` —
    the **`brag add` input contract** (tags as comma-joined string,
    DEC-004/DEC-012) is untouched; the taxonomy commands do not change
    how tags are entered. Stay.
  - `docs/architecture.md` Mermaid `Migrations` node (~line 47) — no new
    migration, so the `0001/0002/0003` list is unchanged. Stays.
  - All `docs/blog/**`, `docs/framework-feedback/**`, `docs/reports/**`,
    `docs/macos-notarization-checklist.md`, `docs/development.md` —
    historical/process/narrative; no shipped-behavior status claim about
    the tag command surface. Stay.

## Acceptance Criteria

- [ ] `brag tags` lists every **in-use** tag (count ≥ 1) as plain
  tab-separated `name\tcount` rows on stdout, ordered **count DESC, then
  name ASC**; stderr empty; exit 0. An orphan tag (a `tags` row with zero
  taggings) does **not** appear.
- [ ] `brag tags --format json` emits a **naked JSON array** of
  `{"tag": <name>, "count": <n>}` objects (2-space indent, DEC-011
  discipline), same order; an empty corpus emits `[]` (not `null`).
- [ ] Counts aggregate across **all** `taggable_type`s (only `'entry'`
  exists today); the query carries no per-type filter, so STAGE-007
  `'project'` taggings count automatically.
- [ ] `brag tag rename <old> <new>` renames the tag globally: every entry
  formerly tagged `<old>` reads `<new>` via `Store.Get`/`List`, and
  `brag search <new>` finds them while `brag search <old>` no longer
  matches them by tag (FTS re-synced via the existing `tags_au` trigger).
  stderr `Renamed.`; exit 0.
- [ ] `brag tag rename <old> <new>` where `<new>` already exists returns
  a **user error** (exit 1) naming `brag tag merge`; the DB is unchanged
  (both tags and all memberships intact).
- [ ] `brag tag rename` with a missing `<old>`, with `<old> == <new>`, or
  with the wrong number of args returns a user error (exit 1); DB
  unchanged.
- [ ] `brag tag merge <src> <dst>` folds `<src>`'s memberships into
  `<dst>`, **de-duplicating**: an object tagged both `<src>` and `<dst>`
  ends with exactly one `<dst>` tagging (UNIQUE respected); `<dst>`'s
  count rises by the previously-`src`-only memberships; the `<src>` tag
  row is deleted. stderr `Merged.`; exit 0.
- [ ] After merge, FTS is correct: `brag search <dst>` finds every
  formerly-`src` entry; `brag search <src>` matches none by tag. (Proven
  via DELETE+INSERT firing `taggings_ad`/`taggings_ai`, **not** a schema
  change.)
- [ ] `brag tag merge` with a missing `<src>` or `<dst>`, with
  `<src> == <dst>`, or wrong arg count returns a user error (exit 1); DB
  unchanged.
- [ ] No SQL is added under `internal/cli/` (commands call Store methods
  only); no new migration file exists; `go test ./...`, `gofmt -l .`,
  `go vet ./...`, and `CGO_ENABLED=0 go build ./...` are clean.

## Failing Tests

Written at **design**; build makes them pass. Paths + assertions below.
All are new (purely additive spec — no rewrites). Storage tests use
`t.TempDir()` via `newTestStore`; CLI tests use the `newXxxTestRoot` +
`seedListEntry` patterns.

- **`internal/storage/tags_test.go`** *(new file)*
  - `"TestTagCounts_SortedAcrossEntries"` — seed `auth,perf` /
    `perf` / `auth,backend`. `TagCounts()` returns exactly, in order,
    `[{auth,2},{perf,2},{backend,1}]` (count DESC; `auth`<`perf` ASC
    tiebreak at count 2).
  - `"TestTagCounts_ExcludesOrphans"` — `Add(Tags:"a,b")` then
    `Update(Tags:"b")` (drops `a`'s last membership). `TagCounts()`
    returns only `[{b,1}]` — `a` absent. Assert via raw SQL that the
    orphan row persists: `SELECT COUNT(*) FROM tags WHERE name='a'` == 1
    (orphans linger but are invisible — DEC-016 choice 4).
  - `"TestTagCounts_EmptyCorpus"` — no entries → `TagCounts()` returns a
    non-nil empty slice (len 0).
  - `"TestRenameTag_GlobalAndFTSReSync"` — seed two title-free entries
    tagged `auth` (and a third tagged `perf`). `RenameTag("auth",
    "authz")`. Assert: each formerly-`auth` `Get().Tags` now reads
    `authz`; `Search("authz")` returns both; `Search("auth")` returns
    zero (title-free fixtures so only the tag column can match).
  - `"TestRenameTag_IntoExistingErrors"` — seed `auth`, `perf`.
    `RenameTag("auth","perf")` → `errors.Is(err, ErrTagExists)`; assert
    no change (`auth` still exists, both memberships intact).
  - `"TestRenameTag_MissingOldErrors"` — `RenameTag("nope","x")` →
    `errors.Is(err, ErrTagNotFound)`.
  - `"TestMergeTags_FoldsDeDupsAndDropsSrc"` — seed e1 `auth,perf`, e2
    `perf`, e3 `auth` (title-free). `MergeTags("auth","perf")`. Assert:
    (a) `TagCounts()` == `[{perf,3}]` and `auth` is gone;
    (b) `Get(e1).Tags == "perf"` (de-duped — one `perf`, no `auth`);
    (c) raw `SELECT COUNT(*) FROM taggings tg JOIN tags t ON t.id=tg.tag_id
    WHERE tg.taggable_id=<e1> AND t.name='perf'` == 1;
    (d) raw `SELECT COUNT(*) FROM tags WHERE name='auth'` == 0;
    (e) `Search("perf")` → 3, `Search("auth")` → 0 (FTS correct).
  - `"TestMergeTags_MissingErrors"` — `MergeTags("nope","perf")` and
    `MergeTags("auth","nope")` each → `errors.Is(err, ErrTagNotFound)`.

- **`internal/cli/tags_test.go`** *(new file)*
  - `"TestTagsCmd_PlainSortedOutput"` — seed via `seedListEntry`;
    `brag tags` stdout is the `name\tcount` rows in count-DESC/name-ASC
    order; stderr empty; exit 0.
  - `"TestTagsCmd_JSON"` — `brag tags --format json` → valid JSON that
    unmarshals to `[]{tag,count}` in order; assert it is a naked array
    (`strings.HasPrefix(strings.TrimSpace(out), "[")`) and 2-space
    indented.
  - `"TestTagsCmd_EmptyCorpus"` — no entries → plain mode stdout empty
    (exit 0); `--format json` → `[]`.
  - `"TestTagsCmd_UnknownFormat"` — `brag tags --format xml` →
    `errors.Is(err, ErrUser)`.
  - `"TestTagsCmd_StdoutStderrSeparation"` — data on stdout only;
    `errBuf.Len() == 0` on success (the §9 separation discipline).
  - `"TestTagsCmd_HelpShowsExamples"` — `brag tags --help` contains
    `"Examples:"` and a distinctive `brag tags` example line (unique
    token per the SPEC-005 lesson).

- **`internal/cli/tag_test.go`** *(new file)*
  - `"TestTagCmd_Rename"` — seed `auth`; `brag tag rename auth authz`
    exits 0 with stderr `Renamed.`; `brag list --tag authz` then finds
    the entries and `brag list --tag auth` finds none.
  - `"TestTagCmd_RenameIntoExistingErrors"` — seed `auth`, `perf`;
    `brag tag rename auth perf` → `errors.Is(err, ErrUser)`, message
    mentions `merge`.
  - `"TestTagCmd_RenameMissingErrors"` — `brag tag rename nope x` →
    `ErrUser`, message names the missing tag.
  - `"TestTagCmd_RenameSameNameErrors"` — `brag tag rename auth auth` →
    `ErrUser`.
  - `"TestTagCmd_RenameArgCountErrors"` — `brag tag rename auth` (one arg)
    → `ErrUser`.
  - `"TestTagCmd_Merge"` — seed e1 `auth,perf`, e3 `auth`;
    `brag tag merge auth perf` exits 0 with stderr `Merged.`;
    `brag list --tag perf` returns both entries; `brag list --tag auth`
    returns none.
  - `"TestTagCmd_MergeDeDups"` — e1 tagged both `auth` and `perf`; after
    `brag tag merge auth perf`, `brag tags` shows `perf` with the
    de-duped count (e1 counted once), and `brag search auth` (via a
    `search` run, or assert through `tags`) shows `auth` gone.
  - `"TestTagCmd_MergeMissingErrors"` — `brag tag merge nope perf` and
    `brag tag merge auth nope` → `ErrUser` (each names the missing tag;
    the missing-`dst` message mentions `rename`).
  - `"TestTagCmd_MergeSameNameErrors"` — `brag tag merge auth auth` →
    `ErrUser`.
  - `"TestTagCmd_MergeArgCountErrors"` — wrong arg count → `ErrUser`.

> **Locked-decision ↔ test traceability (§9).** Each DEC-016 choice has a
> paired failing test: choice 1 (in-use-only + sort + JSON shape) →
> `TestTagCounts_SortedAcrossEntries` + `TestTagCounts_ExcludesOrphans` +
> `TestTagsCmd_PlainSortedOutput` + `TestTagsCmd_JSON`; choice 2
> (rename-errors-into-existing) → `TestRenameTag_IntoExistingErrors` +
> `TestTagCmd_RenameIntoExistingErrors`; choice 3 (merge DELETE+INSERT
> de-dup + FTS) → `TestMergeTags_FoldsDeDupsAndDropsSrc` +
> `TestTagCmd_MergeDeDups`; choice 4 (orphans invisible, no GC) →
> `TestTagCounts_ExcludesOrphans`.

## Implementation Context

*Read this section and the files it points to before build. This is the
whole handoff — the build session has only this spec.*

### Decisions that apply

- **`DEC-016`** (emitted by this spec) — the four locked choices: (1)
  `brag tags` is in-use-only, counted across all taggable types, sorted
  count-DESC/name-ASC, plain `name\tcount` or naked `[{tag,count}]` JSON;
  (2) rename **errors** into an existing name (no auto-merge); (3) merge
  is **DELETE+INSERT** de-dup (never `UPDATE taggings SET tag_id`), no
  migration; (4) orphans are invisible and **not GC'd**. Read it in full
  — it is the spec's primary contract.
- **`DEC-015`** — the `tags(id, name UNIQUE)` + polymorphic
  `taggings(tag_id, taggable_type, taggable_id, position)` schema and the
  six-trigger FTS topology these mutations operate on. Crucially: there
  is `taggings_ai` (AFTER INSERT) + `taggings_ad` (AFTER DELETE) +
  `tags_au` (AFTER UPDATE), but **no `taggings_au`** — the reason merge
  must be DELETE+INSERT.
- **`DEC-011`** — naked JSON array, 2-space indent, `[]`-not-`null` on
  empty. `brag tags --format json` follows it.
- **`DEC-013` / `DEC-014`** — count ordering (DESC by count, alpha-ASC
  tiebreak). `brag tags` default sort inherits it; the JSON `{tag,count}`
  element shape matches DEC-014's `top_tags`.
- **`DEC-010`** — `brag search` stays byte-stable; rename/merge keep FTS
  correct through the existing triggers, so search semantics don't change.
- **`DEC-006`** — cobra: each new command is a `*cobra.Command` built by
  a `NewXxxCmd()` constructor, registered in `cmd/brag/main.go`.
- **`DEC-007`** — required/positional validation lives inline in `RunE`
  via `UserErrorf` (cobra's built-in arg validators return unwrappable
  plain errors). Mirror `delete.go`/`show.go`.
- **`DEC-005`** — INTEGER autoincrement PKs; `tags.id`/`taggings.id`.

### Constraints that apply

- `no-sql-in-cli-layer` (blocking) — `tags.go`/`tag.go` import only
  `storage`/`config`/`export`/`cobra`; **never** `database/sql`. All SQL
  lives in the three new `Store` methods.
- `storage-tests-use-tempdir` (blocking) — all new storage tests use
  `t.TempDir()` (via `newTestStore`); never touch `~/.bragfile`.
- `errors-wrap-with-context` (warning) — wrap storage errors
  (`fmt.Errorf("rename tag: %w", err)` etc.); wrap tx begin/exec/commit.
- `stdout-is-for-data-stderr-is-for-humans` (blocking) — `brag tags`
  rows + JSON go to **stdout**; `Renamed.`/`Merged.` confirmations and
  all errors go to **stderr** (mirror `delete.go`'s `Deleted.`).
- `test-before-implementation` (blocking) — the Failing Tests above are
  the design deliverable.
- `one-spec-per-pr` (blocking) — the PR references SPEC-026 only.
- `no-cgo` / `no-new-top-level-deps-without-decision` — pure-Go path; no
  new dependency.

### Design-time pre-flight (§12(b)) — already done, results below

Per AGENTS.md §12(b), the merge and rename SQL were **run at design**
against the real driver (`modernc.org/sqlite` 1.51.0, SQLite 3.53.1) on
a representative corpus via a throwaway `_test.go` exercising the live
0003 schema + triggers (then deleted). Confirmed:

- **rename** `UPDATE tags SET name='authz' WHERE name='auth'` fires
  `tags_au`, re-syncing `entries_fts`: `Search("authz")` → the two
  entries, `Search("auth")` → 0. **rename-into-existing** (`UPDATE … SET
  name='perf'` when `perf` exists) raises
  `UNIQUE constraint failed: tags.name (2067)` — which is why the method
  pre-checks and returns `ErrTagExists` for a clean message rather than
  surfacing the raw constraint error.
- **merge** via the INSERT-where-absent → DELETE-src-taggings →
  DELETE-src-tag sequence (below) on a corpus where e1 was tagged both
  `src` and `dst`: e1 ends with exactly one `dst` tagging,
  `Get(e1).Tags == "dst"`, the de-duped count is correct (`perf` = 3),
  `entries_fts` shows only `dst`, the `src` tag row is gone, and
  `Search(dst)` → all, `Search(src)` → 0. No `taggings_au` needed.

Build does **not** need to re-derive this, but should re-run the
premise-audit greps (build-side cross-check, §9).

### Dev/prod DB isolation (PROJ-002 brief) — still mandatory this stage

The schema is already at v0.2.x (post-0003) from SPEC-025. While v0.2.x
is in flight:

- Build/run the dev binary against a **dev DB**:
  `BRAGFILE_DB=~/.bragfile-dev/db.sqlite` (or `--db`), via `just install`
  → `~/go/bin/brag`. **Never open the production `~/.bragfile/db.sqlite`**
  with a v0.2.x binary. (SPEC-026 adds no migration, but the binary still
  carries 0003, so the rule stands.)
- Production stays brew-installed at v0.1.0; the documented upgrade is
  STAGE-008.
- All tests use `t.TempDir()` regardless.

### Prior related work

- `SPEC-025` (shipped, PR #36) — the schema, triggers, Store cutover; it
  explicitly deferred the command surface **and orphan-tag GC** to this
  spec (see its Out-of-scope and Reflection Q3). Its `canonicalizeTags` /
  `insertTaggings` / transactional `Add`/`Update`/`Delete` are the
  patterns the new methods mirror.
- `SPEC-007` (shipped) — the `list --tag` exact-token semantics
  (`"auth"`≠`"authoring"`) the join already preserves; rename/merge don't
  touch it.
- `SPEC-020` (shipped) — `brag stats` `top_tags` `{tag,count}` JSON shape
  the `brag tags` JSON mirrors.

### Out of scope (for this spec specifically)

- **Projects as a taggable type** (`taggable_type='project'`) → STAGE-007.
  `brag tags` counts across all types already, so projects fold in with
  no change here; but no project rows are written this spec.
- **Any third taggable type** (`goals`, …) — paper-sketched in DEC-015
  only; no code.
- **A `taggings_au` trigger / any `0004_*` migration** — deliberately not
  added (DEC-016 choice 3 rejects it). If build thinks it needs one,
  STOP and raise a question — it is a scope and decision change.
- **Orphan-tag GC command / auto-sweep** — explicitly deferred
  (DEC-016 choice 4). Orphans are invisible and reused; no cleanup ships.
- **Per-taggable-type count breakdown in `brag tags`** — single total
  count only this spec; a per-type split is a STAGE-007-or-later additive
  extension.
- **`brag tags --format tsv` / `--limit` / `--sort` flags** — YAGNI;
  plain output is already tab-separated and pipeable, and `--format json`
  covers the machine path. Not added.
- **CHANGELOG `[0.2.0]`, RC cut, migration-prompt safety belt,
  architecture diagram refresh, `internal/projects`** → STAGE-008.

## Notes for the Implementer

### `storage` — the three methods (validated SQL literals)

Add to `internal/storage/errors.go`:

```go
// ErrTagNotFound is returned (wrapped) when a named tag does not exist
// (RenameTag old name, MergeTags src/dst). Callers map it to a
// user-facing "no tag named X" error.
var ErrTagNotFound = errors.New("tag not found")

// ErrTagExists is returned (wrapped) by RenameTag when the target name
// already names a tag. Callers map it to "use merge".
var ErrTagExists = errors.New("tag already exists")
```

Add the `TagCount` type and methods to `internal/storage/store.go`
(reuse the SPEC-025 tx pattern; wrap every begin/exec/commit). The SQL
below was pre-flighted — transcribe it.

```go
// TagCount is one row of the brag tags taxonomy view: a tag name and
// its total membership count across all taggable types (DEC-016).
type TagCount struct {
    Name  string
    Count int
}

// TagCounts returns every in-use tag (count >= 1) with its total
// taggings count across all taggable_types, ordered count DESC then
// name ASC (DEC-016 choice 1). Orphan tags (zero taggings) are omitted.
func (s *Store) TagCounts() ([]TagCount, error) {
    rows, err := s.db.QueryContext(context.Background(),
        `SELECT t.name, COUNT(tg.id) AS cnt
           FROM tags t
           JOIN taggings tg ON tg.tag_id = t.id
          GROUP BY t.id, t.name
          ORDER BY cnt DESC, t.name ASC`)
    if err != nil {
        return nil, fmt.Errorf("tag counts: %w", err)
    }
    defer rows.Close()
    out := make([]TagCount, 0)
    for rows.Next() {
        var tc TagCount
        if err := rows.Scan(&tc.Name, &tc.Count); err != nil {
            return nil, fmt.Errorf("tag counts: %w", err)
        }
        out = append(out, tc)
    }
    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("tag counts: %w", err)
    }
    return out, nil
}
```

`RenameTag` — pre-check target existence (clean `ErrTagExists` instead of
the raw UNIQUE error), then `UPDATE` (fires `tags_au` → FTS re-sync), and
use `RowsAffected == 0` to detect a missing `<old>`. Caller (`tag.go`)
guards `old == new` first, so this method assumes distinct names:

```go
func (s *Store) RenameTag(oldName, newName string) error {
    ctx := context.Background()
    tx, err := s.db.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("rename tag: %w", err)
    }
    defer tx.Rollback()

    var exists int
    if err := tx.QueryRowContext(ctx,
        `SELECT COUNT(*) FROM tags WHERE name = ?`, newName,
    ).Scan(&exists); err != nil {
        return fmt.Errorf("rename tag: %w", err)
    }
    if exists > 0 {
        return fmt.Errorf("rename tag %q -> %q: %w", oldName, newName, ErrTagExists)
    }

    res, err := tx.ExecContext(ctx,
        `UPDATE tags SET name = ? WHERE name = ?`, newName, oldName)
    if err != nil {
        return fmt.Errorf("rename tag: %w", err)
    }
    n, err := res.RowsAffected()
    if err != nil {
        return fmt.Errorf("rename tag: %w", err)
    }
    if n == 0 {
        return fmt.Errorf("rename tag %q: %w", oldName, ErrTagNotFound)
    }
    if err := tx.Commit(); err != nil {
        return fmt.Errorf("rename tag: %w", err)
    }
    return nil
}
```

`MergeTags` — resolve both ids (each missing → `ErrTagNotFound`), then
the **pre-flighted three-statement sequence**. Order is load-bearing:
INSERT dst-where-absent *before* deleting src taggings. Caller guards
`src == dst`:

```go
func (s *Store) MergeTags(src, dst string) error {
    ctx := context.Background()
    tx, err := s.db.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("merge tags: %w", err)
    }
    defer tx.Rollback()

    var srcID, dstID int64
    if err := tx.QueryRowContext(ctx,
        `SELECT id FROM tags WHERE name = ?`, src).Scan(&srcID); err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return fmt.Errorf("merge tags: source %q: %w", src, ErrTagNotFound)
        }
        return fmt.Errorf("merge tags: %w", err)
    }
    if err := tx.QueryRowContext(ctx,
        `SELECT id FROM tags WHERE name = ?`, dst).Scan(&dstID); err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return fmt.Errorf("merge tags: destination %q: %w", dst, ErrTagNotFound)
        }
        return fmt.Errorf("merge tags: %w", err)
    }

    // 1. Give every src-tagged object a dst tagging it doesn't already
    //    have (fires taggings_ai → FTS re-sync). NOT EXISTS skips the
    //    would-be UNIQUE duplicate for objects tagged both.
    if _, err := tx.ExecContext(ctx,
        `INSERT INTO taggings (tag_id, taggable_type, taggable_id, position)
         SELECT ?, s.taggable_type, s.taggable_id, s.position
           FROM taggings s
          WHERE s.tag_id = ?
            AND NOT EXISTS (SELECT 1 FROM taggings d
                             WHERE d.tag_id = ?
                               AND d.taggable_type = s.taggable_type
                               AND d.taggable_id  = s.taggable_id)`,
        dstID, srcID, dstID); err != nil {
        return fmt.Errorf("merge tags: graft dst: %w", err)
    }
    // 2. Drop all src taggings (fires taggings_ad → FTS re-sync, removing
    //    the src token from the projection where it still lingered).
    if _, err := tx.ExecContext(ctx,
        `DELETE FROM taggings WHERE tag_id = ?`, srcID); err != nil {
        return fmt.Errorf("merge tags: drop src taggings: %w", err)
    }
    // 3. Remove the now-unreferenced src tag row.
    if _, err := tx.ExecContext(ctx,
        `DELETE FROM tags WHERE id = ?`, srcID); err != nil {
        return fmt.Errorf("merge tags: drop src tag: %w", err)
    }

    if err := tx.Commit(); err != nil {
        return fmt.Errorf("merge tags: %w", err)
    }
    return nil
}
```

(`MergeTags` imports `database/sql` only for `sql.ErrNoRows`, which
`store.go` already imports.)

### `export` — the JSON helper

Add to `internal/export/json.go` (beside `ToJSON`), keeping the
naked-array + 2-space-indent DEC-011 discipline and the DEC-014
`{tag,count}` element shape:

```go
// tagCountJSON is the per-tag element of brag tags --format json. The
// key is "tag" (matching DEC-014 top_tags), not the SQL column "name".
type tagCountJSON struct {
    Tag   string `json:"tag"`
    Count int    `json:"count"`
}

// ToTagsJSON renders the tag taxonomy as a naked JSON array of
// {tag, count} objects (DEC-011 shape; DEC-016 choice 1). Empty input
// renders "[]", never "null".
func ToTagsJSON(tags []storage.TagCount) ([]byte, error) {
    out := make([]tagCountJSON, 0, len(tags))
    for _, tc := range tags {
        out = append(out, tagCountJSON{Tag: tc.Name, Count: tc.Count})
    }
    b, err := json.MarshalIndent(out, "", "  ")
    if err != nil {
        return nil, fmt.Errorf("marshal tags json: %w", err)
    }
    return b, nil
}
```

### `cli` — `tags.go` (brag tags)

Mirror `stats.go`'s `--format` handling and `list.go`'s open/resolve
pattern. Plain mode writes `name\tcount` rows to stdout; JSON mode writes
`export.ToTagsJSON`. Unknown `--format` → `UserErrorf`. Suggested `Long`
+ `Examples:` (include a unique `brag tags` token for the help test):

```
Use:   "tags"
Short: "List every tag with its usage count"
Long: `List every in-use tag with its total usage count across all entries, one per line as <name>\t<count>, sorted by count (descending) then name (ascending).

Output is plain tab-separated rows (default) or a JSON array of {tag, count} objects (--format json) per DEC-016. Tags with no remaining uses are omitted. Counts span all taggable objects (entries today; projects in a later release).

Examples:
  brag tags                         # name<TAB>count rows, most-used first
  brag tags --format json           # naked JSON array of {tag, count}`
```

Plain render:

```go
for _, tc := range tags {
    fmt.Fprintf(out, "%s\t%d\n", tc.Name, tc.Count)
}
```

### `cli` — `tag.go` (brag tag rename / merge)

`NewTagCmd()` returns a parent `brag tag` command (Use `"tag"`, Short
"Tag taxonomy operations (rename, merge)") with no `RunE` of its own —
attach two subcommands via `cmd.AddCommand(newTagRenameCmd(),
newTagMergeCmd())`. (A bare `brag tag` then prints help, like cobra's
default for a command with subcommands and no Run.) Each subcommand does
inline positional validation (DEC-007), opens the store, calls the Store
method, and maps sentinels to `UserErrorf`. Confirmations go to
**stderr** (`Renamed.` / `Merged.`), matching `delete.go`'s `Deleted.`.

Rename `RunE` shape:

```go
if len(args) != 2 {
    return UserErrorf("rename requires exactly <old> and <new> tag names")
}
oldName, newName := args[0], args[1]
if oldName == "" || newName == "" {
    return UserErrorf("tag names must not be empty")
}
if oldName == newName {
    return UserErrorf("old and new tag names are the same (%q)", oldName)
}
// ... open store ...
if err := s.RenameTag(oldName, newName); err != nil {
    switch {
    case errors.Is(err, storage.ErrTagNotFound):
        return UserErrorf("no tag named %q", oldName)
    case errors.Is(err, storage.ErrTagExists):
        return UserErrorf("tag %q already exists; use `brag tag merge %s %s` to combine them", newName, oldName, newName)
    }
    return fmt.Errorf("rename tag: %w", err)
}
fmt.Fprintln(cmd.ErrOrStderr(), "Renamed.")
return nil
```

Merge `RunE` shape (note the missing-`dst` message points at `rename`,
the symmetric inverse of rename pointing at `merge`):

```go
if len(args) != 2 {
    return UserErrorf("merge requires exactly <src> and <dst> tag names")
}
src, dst := args[0], args[1]
if src == "" || dst == "" {
    return UserErrorf("tag names must not be empty")
}
if src == dst {
    return UserErrorf("source and destination tags are the same (%q)", src)
}
// ... open store ...
if err := s.MergeTags(src, dst); err != nil {
    if errors.Is(err, storage.ErrTagNotFound) {
        // The wrapped message says which of src/dst was missing.
        // Re-derive a friendly hint: if dst is the missing one, suggest rename.
        return UserErrorf("%v (use `brag tag rename` if you meant to rename)", err)
    }
    return fmt.Errorf("merge tags: %w", err)
}
fmt.Fprintln(cmd.ErrOrStderr(), "Merged.")
return nil
```

> The merge error-mapping above is a *sketch*; the precise message
> wording is the implementer's call as long as the test assertions hold
> (missing tag named; `rename` mentioned when relevant). If you prefer to
> distinguish src-missing vs dst-missing with two sentinels for cleaner
> messages, that is acceptable — but it is an additive refinement, not a
> contract change, so keep it inside this spec (no new DEC).

### Registration

In `cmd/brag/main.go`, after the existing `root.AddCommand(...)` calls:

```go
root.AddCommand(cli.NewTagsCmd())
root.AddCommand(cli.NewTagCmd())
```

### Gotchas

- **Do NOT implement merge as `UPDATE taggings SET tag_id`.** There is no
  `taggings_au` trigger (SPEC-025), so FTS would silently desync, and an
  object tagged both src+dst would violate `UNIQUE(taggable_type,
  taggable_id, tag_id)`. DELETE+INSERT is the validated path (DEC-016
  choice 3). This is the single most important constraint in the spec.
- **`brag tags` counts through the join (INNER JOIN / GROUP BY), not the
  `tags` table directly.** That is what makes orphans invisible and GC
  unnecessary. Do not "helpfully" LEFT JOIN to show zero-count tags.
- **No SQL in `tag.go`/`tags.go`.** The `src==dst` / `old==new` / empty /
  arg-count checks are plain string checks (allowed, like `delete.go`'s
  id parse); everything touching the DB is a `Store` method call.
- **Confirmations on stderr, data on stdout.** `brag tags` rows/JSON →
  stdout; `Renamed.`/`Merged.` → stderr. A `brag tags` test should assert
  `errBuf.Len() == 0` (no leakage), per the §9 separation discipline.
- **Empty `brag tags`:** plain mode prints nothing (loop over an empty
  slice); JSON mode prints `[]` (the `make([]…, 0)` + `MarshalIndent`
  yields `[]`, never `null`).
- **Literal-artifact discipline (§12 / §12(b)):** the Store SQL and the
  JSON helper above are the validated literals — transcribe them; if the
  build must change the merge/rename SQL, re-run the design-time
  pre-flight before deviating.

---

## Build Completion

*Filled in at the end of the **build** cycle, before advancing to verify.*

- **Branch:**
- **PR (if applicable):**
- **All acceptance criteria met?** yes/no
- **New decisions emitted:**
  - `DEC-016` — Tag mutation semantics (emitted at design, not build)
- **Deviations from spec:**
  - [list]
- **Follow-up work identified:**
  - [any new specs for the stage's backlog]

### Build-phase reflection (3 questions, short answers)

Process-focused: how did the build go? What friction did the spec create?

1. **What was unclear in the spec that slowed you down?**
   — <answer>

2. **Was there a constraint or decision that should have been listed but wasn't?**
   — <answer>

3. **If you did this task again, what would you do differently?**
   — <answer>

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
