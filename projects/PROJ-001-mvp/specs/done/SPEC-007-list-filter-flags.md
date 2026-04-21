---
task:
  id: SPEC-007
  type: story
  cycle: ship
  blocked: false
  priority: high
  complexity: M

project:
  id: PROJ-001
  stage: STAGE-002
repo:
  id: bragfile

agents:
  architect: claude-opus-4-7
  implementer: claude-opus-4-7
  created_at: 2026-04-20

references:
  decisions:
    - DEC-004  # tags comma-joined (revisited; stays)
    - DEC-005  # integer autoincrement IDs
    - DEC-006  # cobra framework
    - DEC-007  # RunE-validated required args (positional AND flags)
    - DEC-008  # --since date format (emitted during this spec's design)
  constraints:
    - no-sql-in-cli-layer
    - storage-tests-use-tempdir
    - stdout-is-for-data-stderr-is-for-humans
    - errors-wrap-with-context
    - timestamps-in-utc-rfc3339
    - migrations-are-append-only
    - test-before-implementation
    - one-spec-per-pr
  related_specs:
    - SPEC-002  # shipped; Store.List(ListFilter{}) stub + scan patterns
    - SPEC-004  # shipped; list command shape
    - SPEC-005  # shipped; assertion-specificity lesson
    - SPEC-006  # shipped; DEC-007 extension to positional-arg validation
---

# SPEC-007: `list` filter flags + Store filtering

## Context

Third spec in STAGE-002. `brag list` currently prints every entry
ever captured — fine at 5 rows, unwieldy past ~30. SPEC-002 shipped
`ListFilter` as a named empty struct specifically so this spec could
extend it without changing `Store.List`'s signature. This spec gives
users five filter flags (`--tag`, `--project`, `--type`, `--since`,
`--limit`), wires them through to a WHERE-clause-building `Store.List`,
and closes out DEC-004's open question on the tag-storage model by
applying the sentinel-comma trick (`"auth"` must match `"auth,perf"`
but NOT `"authoring"`).

It also introduces DEC-008 (`--since` date format) — opinionated but
conventional, written alongside this spec's design.

Parent stage: `STAGE-002-capture-and-retrieval.md`. Project: PROJ-001.

## Goal

Ship five filter flags on `brag list` — `--tag`, `--project`,
`--type`, `--since`, `--limit` — each combinable via logical AND,
with correct semantics (tag-token matching via sentinel-comma;
exact-match for project/type; date-or-duration for `--since`; strict
positive integer for `--limit`) — such that the resulting `Store.List`
query is a single correctly-parameterized SQL statement that
preserves the existing ordering (`created_at DESC, id DESC`) and the
existing empty-result contract.

## Inputs

- **Files to read:**
  - `docs/api-contract.md` — `brag list` filter-flag spec (formats,
    semantics, examples).
  - `docs/data-model.md` — `entries` columns and nullability; the
    "tags comma-joined" note; the indexes that already exist
    (`idx_entries_created_at`, `idx_entries_project`).
  - `AGENTS.md` §8 (conventions), §9 (testing: separate buffers,
    tie-break ordering, fail-first, assertion specificity).
  - `/decisions/DEC-004-tags-comma-joined-for-mvp.md` — the tag
    format; this spec validates the revisit criterion.
  - `/decisions/DEC-005-integer-autoincrement-ids.md`
  - `/decisions/DEC-006-cobra-cli-framework.md`
  - `/decisions/DEC-007-required-flag-validation-in-runE.md` — the
    RunE-validation pattern (flags AND positional args).
  - `/decisions/DEC-008-since-date-format.md` — format contract for
    `--since`.
  - `/guidance/constraints.yaml`
  - `/guidance/questions.yaml` — the `tags-storage-model` question
    gets answered during this spec.
  - `internal/storage/entry.go` — `ListFilter` is currently
    `struct{}`; this spec extends it.
  - `internal/storage/store.go` — `Store.List` current shape;
    extends its SQL to build WHERE + LIMIT from the filter.
  - `internal/storage/store_test.go` — existing list tests;
    regression-check them after the WHERE logic lands.
  - `internal/cli/list.go` + `internal/cli/list_test.go` — existing
    list command; gains 5 flags.
  - `internal/cli/errors.go` — `ErrUser` / `UserErrorf`.
- **External APIs:** none.
- **Related code paths:** `internal/cli/`, `internal/storage/`.

## Outputs

- **Files created:**
  - `internal/cli/since.go` — `ParseSince(s string) (time.Time,
    error)` helper implementing DEC-008's format.
  - `internal/cli/since_test.go` — pure-function tests for the
    parser (no DB, no cobra).
- **Files modified:**
  - `internal/storage/entry.go` — `ListFilter` extended:
    ```go
    type ListFilter struct {
        Tag     string    // empty = no filter
        Project string
        Type    string
        Since   time.Time // zero = no filter
        Limit   int       // 0 = no limit
    }
    ```
  - `internal/storage/store.go` — `Store.List` builds WHERE clause
    and LIMIT conditionally from populated filter fields; preserves
    existing ORDER BY.
  - `internal/storage/store_test.go` — add `TestList_FilterBy*`
    tests; existing `TestList_*` (the two from SPEC-002/SPEC-004)
    stay green unchanged.
  - `internal/cli/list.go` — `NewListCmd` declares the five flags;
    `runList` populates `ListFilter` using `cmd.Flags().Changed(...)`
    to distinguish "flag not set" from "set to empty/zero"; `--since`
    parsed via `ParseSince`; `--limit` validated `> 0`; `Long`
    gains an Examples block.
  - `internal/cli/list_test.go` — add `TestListCmd_FilterBy*` and
    invalid-input tests; existing tests stay green.
  - `docs/tutorial.md` — update §4 "Read them back" with filter
    examples; strike `brag list --tag auth` from §9 "What's NOT
    there yet".
  - `guidance/questions.yaml` — mark `tags-storage-model` as
    `status: answered` with a note referencing this spec.
  - `decisions/DEC-004-tags-comma-joined-for-mvp.md` — add a
    Validation note: comma-joined confirmed adequate at MVP scale;
    sentinel-comma pattern handles the `"auth"` vs `"authoring"`
    edge case without requiring normalization.
- **New exports:**
  - `cli.ParseSince(s string) (time.Time, error)`
  - `storage.ListFilter` gains the 5 fields above (was empty struct).
- **Database changes:** none. No new migration. Existing indexes
  (`idx_entries_created_at`, `idx_entries_project`) already cover
  the common filter cases.

## Filter semantics

Locked in the spec so build / verify don't re-litigate:

| Flag | Type | Empty/zero | Match |
|---|---|---|---|
| `--tag` | string | no filter | sentinel-comma token match (see below) |
| `--project` | string | no filter | exact string equality on `entries.project` |
| `--type` | string | no filter | exact string equality on `entries.type` |
| `--since` | string (parsed) | no filter | `entries.created_at >= parsed-value RFC3339` |
| `--limit` | int | 0 = no limit | `LIMIT n` appended |

All populated filters combine via SQL **AND**. Empty/unset filters
contribute no WHERE clause. Ordering is unchanged:
`ORDER BY created_at DESC, id DESC`.

**Tag sentinel-comma pattern:**

```sql
WHERE ',' || tags || ',' LIKE '%,' || ? || ',%'
```

with the `?` bound to the tag value. This makes `--tag auth` match a
row whose `tags` column is `"auth"`, `"auth,perf"`, `"perf,auth"`,
or `"perf,auth,backend"`, but NOT `"authoring"` or `"authors"`. Rows
with NULL or empty `tags` never match any tag filter (SQL `NULL LIKE
anything` evaluates to NULL → treated as false). Required regression
test case: the spec's `TestList_TagFilterNoFalsePositive` covers
`"auth"` vs `"authoring"` and closes DEC-004's revisit criterion.

**Flag-set detection:**

`cmd.Flags().Changed("flagname")` distinguishes "flag not passed" from
"flag set to empty string / zero". The CLI layer only populates the
corresponding `ListFilter` field if `Changed` is true. This keeps
value-based `ListFilter` semantics simple — zero values mean "no
filter" — without losing the user's intent when they explicitly pass
an empty value. (Passing `--tag ""` is rejected as `ErrUser` — see
acceptance criteria.)

**No shorthand flags this spec.** `--tag`/`--project`/`--type`/
`--since`/`--limit` stay long-form only. SPEC-005 took `-t`, `-d`,
`-T`, `-p`, `-k`, `-i` for the `add` command's fields, and reusing
any of those letters for `list` would create cross-subcommand
semantic collision (`-t` means "title" on add; it would mean "tag"
on list — confusing). Can revisit in a dedicated polish spec if
real usage demands it.

## Acceptance Criteria

- [ ] `brag list` with no flags behaves identically to before
      (backwards-compatible). *[regression: existing
      TestListCmd_EmptyPrintsNothing and
      TestListCmd_PrintsReverseChronological still pass]*
- [ ] `brag list --tag auth` returns only rows whose tags contain
      `auth` as a comma-separated token. *[TestListCmd_FilterByTag]*
- [ ] `brag list --tag auth` does NOT match a row whose tags is
      `"authoring"` or `"authors"`. *[TestListCmd_TagFilterNoFalsePositive]*
- [ ] `brag list --project platform` returns only rows where
      `project` equals `"platform"` exactly (case-sensitive).
      *[TestListCmd_FilterByProject]*
- [ ] `brag list --type shipped` returns only rows where `type`
      equals `"shipped"` exactly. *[TestListCmd_FilterByType]*
- [ ] `brag list --since 2026-01-01` returns only rows with
      `created_at >= 2026-01-01T00:00:00Z`. *[TestListCmd_FilterBySince_ISODate]*
- [ ] `brag list --since 7d` returns only rows with `created_at`
      within the last 7 × 24 hours. *[TestListCmd_FilterBySince_Days]*
- [ ] `brag list --since 2w` and `brag list --since 3m` behave
      analogously (2w = 14 days; 3m = 90 days — DEC-008).
      *[TestListCmd_FilterBySince_Weeks,
      TestListCmd_FilterBySince_Months]*
- [ ] `brag list --since notadate` and other malformed values return
      `ErrUser` (`main.go` maps to exit 1).
      *[TestListCmd_InvalidSinceIsUserError]*
- [ ] `brag list --limit 2` caps the result count at 2.
      *[TestListCmd_FilterByLimit]*
- [ ] `brag list --limit 0` and `brag list --limit -5` return
      `ErrUser` (limit must be positive when passed).
      *[TestListCmd_InvalidLimitIsUserError]*
- [ ] `brag list --tag "" --project ""` (explicit empty strings)
      returns `ErrUser` — an explicitly-set empty filter is a user
      mistake, not "show everything". *[TestListCmd_EmptyFilterValueIsUserError]*
      (**Rationale:** Cobra's `Changed` returns true for `--tag ""`.
      If we silently ignored empty values we'd mask typos; rejecting
      is defensive.)
- [ ] Multiple filters combine via AND: `brag list --project platform
      --since 7d --limit 5` returns at most 5 rows that are both
      project=platform AND from the last 7 days.
      *[TestListCmd_CombinedFilters]*
- [ ] Ordering is preserved: filtered results still come back in
      `created_at DESC, id DESC` order — the tie-break guarantee
      from AGENTS.md §9 holds.
      *[TestListCmd_FilterPreservesOrder]*
- [ ] `brag list --help` shows all five filter flags with
      descriptions and at least one example line.
      *[TestListCmd_HelpShowsFilters]*
- [ ] `ParseSince` tests pass in isolation: valid ISO date, valid
      `Nd`/`Nw`/`Nm`, various invalid inputs (empty, unknown unit,
      zero/negative N, non-numeric N, whitespace).
      *[TestParseSince_* — see Failing Tests]*
- [ ] Existing SPEC-001/002/003/004/005/006 tests remain green. No
      existing test is modified. *[manual: go test ./...]*
- [ ] `gofmt -l .` empty, `go vet ./...` clean, `CGO_ENABLED=0 go
      build ./...` succeeds, `go test ./...` green.
- [ ] `guidance/questions.yaml` — `tags-storage-model` moved to
      `status: answered` with a reference to this SPEC and a note
      that comma-joined + sentinel-comma is sufficient at MVP
      scale.
- [ ] `decisions/DEC-004-tags-comma-joined-for-mvp.md` — Validation
      section updated to reference SPEC-007 as evidence.

## Failing Tests

Written now. Every happy-path CLI test uses separate `outBuf` /
`errBuf` (§9). Every output-shape assertion targets distinctive
content (§9, SPEC-005 lesson). Fail-first run before implementation
(§9, SPEC-003 lesson). Storage tests use `t.TempDir()` (§9, `storage-
tests-use-tempdir` constraint).

### `internal/cli/since_test.go` (new file, pure-function tests)

Imports: `testing`, `time`, package under test.

- **`TestParseSince_ISODate`** — `ParseSince("2026-01-01")` returns
  `time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)` exactly, no error.
- **`TestParseSince_Days`** — `ParseSince("7d")` returns a `time.Time`
  within 1 second of `time.Now().UTC().Add(-7 * 24 * time.Hour)`.
  (Use a tolerance check: `|got - want| < time.Second`.)
- **`TestParseSince_Weeks`** — `ParseSince("2w")` returns within 1s
  of `time.Now().UTC().Add(-14 * 24 * time.Hour)`.
- **`TestParseSince_Months`** — `ParseSince("3m")` returns within 1s
  of `time.Now().UTC().Add(-90 * 24 * time.Hour)` (30-day
  approximation per DEC-008).
- **`TestParseSince_InvalidFormat`** — table-driven test over:
  `""`, `"7"`, `"d"`, `"0d"`, `"-3d"`, `"abc"`, `"7x"`, `"2026-13-01"`
  (bad month), `"  7d  "` (whitespace — design decision: reject, not
  trim). Each must return a non-nil error.

### `internal/storage/store_test.go` (new tests, existing stay green)

Reuse the existing `newTestStore` helper.

- **`TestList_FilterByTag`** — insert three entries with tags
  `"auth,perf"`, `"perf,backend"`, `"auth"`. `List(ListFilter{Tag:
  "auth"})` returns 2 entries; `Tag: "perf"` returns 2; `Tag:
  "backend"` returns 1; `Tag: "nonesuch"` returns empty non-nil
  slice.
- **`TestList_TagFilterNoFalsePositive`** — insert entries with tags
  `"auth"` and `"authoring"`. `List(ListFilter{Tag: "auth"})` returns
  exactly one entry (the one with `"auth"`), not both.
- **`TestList_TagFilterNullAndEmpty`** — insert entries with tags
  `""` and (via direct SQL if needed) `NULL`. `List(ListFilter{Tag:
  "auth"})` returns none. (NULL `LIKE` anything → NULL → false;
  sentinel-comma of empty string is `",,"` which doesn't contain
  `",auth,"`.)
- **`TestList_FilterByProject`** — insert entries with projects
  `"platform"`, `"growth"`, `"platform"`. Filter by
  `"platform"` returns 2; by `"Platform"` (capitalized) returns 0.
- **`TestList_FilterByType`** — analogous to project; exact match.
- **`TestList_FilterBySince`** — insert three entries with
  `created_at` spaced one day apart (use direct `UPDATE` on the DB
  after `Add` to backdate; or have `Add` accept a time override —
  direct UPDATE is simpler and self-contained in the test).
  `List(ListFilter{Since: twoDaysAgo})` returns only entries whose
  `created_at >= twoDaysAgo`.
- **`TestList_FilterByLimit`** — insert 5 entries. `List(
  ListFilter{Limit: 2})` returns exactly 2, in correct order.
- **`TestList_FilterCombined`** — insert entries with various
  tag/project/date permutations. Filter with three fields populated;
  assert result is the AND of all predicates.
- **`TestList_FilterPreservesOrder`** — insert three entries with
  tags `"x"` in rapid succession (same second). `List(ListFilter{Tag:
  "x"})` returns them in `id DESC` order (tie-break), matching
  `Store.List`'s unfiltered ordering contract.

### `internal/cli/list_test.go` (new tests, existing stay green)

Reuse the existing `newRootWithList` helper (or equivalent). Every
test uses separate `outBuf` / `errBuf` with a no-cross-leakage assert.

- **`TestListCmd_FilterByTag`** — insert two entries with disjoint
  tags via `Store.Add`, close store, run `list --tag <one of them>`,
  assert only the matching entry's title appears in stdout and the
  other does not.
- **`TestListCmd_TagFilterNoFalsePositive`** — `"auth"` vs
  `"authoring"`, same shape.
- **`TestListCmd_FilterByProject`**, **`TestListCmd_FilterByType`**,
  **`TestListCmd_FilterBySince_ISODate`**,
  **`TestListCmd_FilterBySince_Days`**,
  **`TestListCmd_FilterBySince_Weeks`**,
  **`TestListCmd_FilterBySince_Months`**,
  **`TestListCmd_FilterByLimit`**: each follows the same pattern —
  insert representative data, filter, assert inclusion/exclusion by
  distinctive title substring.
- **`TestListCmd_CombinedFilters`** — two or three filters together;
  assert the AND semantics.
- **`TestListCmd_InvalidSinceIsUserError`** — `list --since
  notadate`. Assert `errors.Is(err, ErrUser)`, `outBuf.Len() == 0`.
- **`TestListCmd_InvalidLimitIsUserError`** — table over
  `"--limit 0"`, `"--limit -5"`. Each returns `ErrUser`.
- **`TestListCmd_EmptyFilterValueIsUserError`** — `list --tag ""`.
  Assert `errors.Is(err, ErrUser)`.
- **`TestListCmd_HelpShowsFilters`** — `list --help`. Assert
  `outBuf` contains each of `"--tag"`, `"--project"`, `"--type"`,
  `"--since"`, `"--limit"`, AND the distinctive literal `"Examples:"`
  (SPEC-005 lesson).

Notes for the implementer on testing patterns:

- Fail-first: write all the new tests, run `go test ./...`, verify
  each fails for the expected reason (missing field on `ListFilter`,
  undefined `ParseSince`, missing flag on `list` command, etc.). If
  any test unexpectedly passes, tighten the assertion (§9, SPEC-005
  lesson).
- For `--since`-backed tests, backdate with the canonical helper
  `storagetest.Backdate(dbPath, id, at)` from
  `internal/storage/storagetest/`. That sub-package owns the direct
  SQL UPDATE so `internal/cli/**` test files can stay free of
  `database/sql` (the `no-sql-in-cli-layer` constraint is blocking
  and applies to test files too — there is no `_test.go` exclusion
  in the constraint's path glob). Storage tests call the same helper
  for consistency; an earlier draft of this section said "either is
  acceptable" between a CLI-local helper and a storage-side helper,
  but a CLI-local helper imports `database/sql` and violates the
  constraint, so it is NOT acceptable. Always use `storagetest`.
- For `TestList_TagFilterNullAndEmpty`, setting NULL requires a
  direct SQL INSERT (the `Entry` struct has a non-pointer string
  field; `Add` always stores `""`, never NULL). That's fine — the
  test is specifically exercising the NULL-tolerance of the
  sentinel-comma SQL.

## Implementation Context

*Read before starting build. Self-contained handoff.*

### Decisions that apply

- `DEC-004` (revisited) — Tags stay comma-joined TEXT. This spec
  applies the sentinel-comma pattern that resolves the
  `tags-storage-model` question marked open when DEC-004 was
  written. **Update DEC-004's Validation section in this same spec's
  PR** — add a bullet noting SPEC-007 validated the format. Do NOT
  emit a superseding DEC; the original decision stands.
- `DEC-005` — IDs are `INTEGER AUTOINCREMENT`. The tie-break in
  ordering still applies to filtered result sets.
- `DEC-006` — Cobra framework. Use `cmd.Flags().String(...)` and
  `cmd.Flags().Int(...)` with `cmd.Flags().Changed(name)` in
  `runList` to detect "flag was actually passed".
- `DEC-007` — Validation in `RunE`. Applies to `--since` parse
  failures, `--limit <= 0`, and empty-string filter values. Do NOT
  use cobra validators; return `UserErrorf(...)` directly.
- `DEC-008` — `--since` format. `ParseSince` is the canonical
  implementation; build session reads the DEC for any edge-case
  questions rather than inventing.

### Constraints that apply

For `internal/cli/**`, `internal/storage/**`, `docs/tutorial.md`,
`guidance/questions.yaml`, `decisions/DEC-004-*.md`:

- `no-sql-in-cli-layer` — blocking. `since.go` and `list.go` import
  only `config` + `storage` (+ stdlib). WHERE-clause construction
  lives in `store.go`.
- `storage-tests-use-tempdir` — blocking. All new storage tests use
  `t.TempDir()`.
- `stdout-is-for-data-stderr-is-for-humans` — blocking. Filtered
  rows go to stdout; every happy-path test asserts `errBuf.Len() ==
  0`.
- `errors-wrap-with-context` — warning. Wrap every returned storage
  error: `fmt.Errorf("list entries: %w", err)`. Wrap parse errors in
  CLI: `UserErrorf("invalid --since: %w", err)` is acceptable even
  though the wrapped error comes from `ParseSince`.
- `timestamps-in-utc-rfc3339` — blocking. `--since` parsed to UTC
  `time.Time`; passed to SQL as `.UTC().Format(time.RFC3339)`.
- `migrations-are-append-only` — blocking. **This spec does NOT add
  a migration.** Existing schema is sufficient. If anyone in build
  thinks a new index is needed, that's a separate spec; STAGE-001's
  `idx_entries_created_at` and `idx_entries_project` already cover
  the common cases.
- `test-before-implementation` — blocking.
- `one-spec-per-pr` — blocking. Branch
  `feat/spec-007-list-filter-flags`.

### AGENTS.md lessons that apply

- §9 separate `outBuf` / `errBuf` in CLI tests (SPEC-001).
- §9 monotonic tie-break in ordering tests (SPEC-002) — explicitly
  exercised by `TestList_FilterPreservesOrder`.
- §9 fail-first test run before implementation (SPEC-003 ship).
- §9 assertion specificity (SPEC-005 ship) — help-test asserts on
  `"Examples:"`, not a generic `"brag list"` that Usage already
  carries.
- §10 `/`-anchored gitignore — not touched by this spec.

### Prior related work

- **SPEC-002** (shipped). `ListFilter struct{}` was defined
  specifically to absorb future fields without breaking callers.
  `Store.List(ListFilter{})` semantics (return all, reverse-chrono)
  become the "no filter" baseline this spec preserves.
- **SPEC-004** (shipped). `internal/cli/list.go` establishes the
  list-command shape. This spec adds flags and extends `runList`;
  does not replace the command.
- **SPEC-005** (shipped). Shorthand discipline — we deliberately do
  NOT add shorthands to filter flags to avoid cross-subcommand
  letter collision with `add`'s `-t`/`-d`/`-T`/`-p`/`-k`/`-i`.
  Explicit decision, not an oversight.
- **SPEC-006** (shipped). `storage.ErrNotFound` sentinel exists but
  is NOT applicable here — `List` on an empty filtered result set
  is not an error, it's an empty slice.

### Out of scope (for this spec specifically)

If any of these feels necessary during build, write a new spec.

- **Schema normalization** (split `tags` into `tags` / `entry_tags`
  tables). DEC-004 revisited and kept; this spec is the explicit
  "stays comma-joined" decision.
- **New indexes.** `idx_entries_created_at` and `idx_entries_project`
  already exist. No index on `type` or `tags` — at MVP scale (O(100)
  rows) the scan cost is negligible and indexes aren't earning
  their upkeep.
- **Case-insensitive project/type matching.** Exact match only for
  MVP. Users can type their own values consistently. Case-insensitive
  is a future polish if someone asks.
- **Filter on `--tag tag1,tag2`** (AND multiple tags) or
  **`--tag tag1 --tag tag2`** (repeated flag → OR). Single-tag
  filter only in this spec. A future spec can extend `ListFilter.Tag`
  to `[]string` if real usage demands it.
- **`--until` flag** (reverse of `--since`). Not asked for; easy to
  add later without breaking existing behavior.
- **Filter flags on `show`, `delete`, `search`.** Those subcommands
  either take a single ID (show/delete) or use FTS5's MATCH
  operator (search). Not in this spec's scope.
- **Shorthand flags** (`-t` for tag, etc.). Deliberately deferred —
  see Filter Semantics section above.
- **`brag list --reverse`** (chronological order). Not asked for.

## Notes for the Implementer

- **`ListFilter` field order.** In `entry.go`, keep the struct
  field order Tag / Project / Type / Since / Limit so it matches
  the flag declaration order and the acceptance-criteria table.
  Visual symmetry helps future readers.

- **WHERE-clause builder in `Store.List`.** Keep it simple — build
  `[]string` conditions and `[]interface{}` args, then join with
  `" AND "`:
  ```go
  var conds []string
  var args []interface{}
  if f.Tag != "" {
      conds = append(conds, "',' || tags || ',' LIKE ?")
      args = append(args, "%,"+f.Tag+",%")
  }
  if f.Project != "" {
      conds = append(conds, "project = ?")
      args = append(args, f.Project)
  }
  // ... Type, Since ...
  q := "SELECT ... FROM entries"
  if len(conds) > 0 {
      q += " WHERE " + strings.Join(conds, " AND ")
  }
  q += " ORDER BY created_at DESC, id DESC"
  if f.Limit > 0 {
      q += " LIMIT ?"
      args = append(args, f.Limit)
  }
  rows, err := s.db.Query(q, args...)
  ```
  Keep the existing scan/hydrate loop unchanged.

- **Tag bind parameter.** Build the `%,tag,%` string in Go, bind as
  a plain parameter. Do NOT concatenate the tag value into the SQL
  string — that's a SQL injection risk even for a single-user local
  tool (bad habits propagate).

- **`--since` parameter.** Format the parsed `time.Time` as
  `.UTC().Format(time.RFC3339)` before binding. RFC3339 strings
  compare correctly via lexical ordering when both sides are UTC.

- **`cmd.Flags().Changed(name)`.** This is the mechanism for
  distinguishing "flag not passed" from "flag set to empty/zero":
  ```go
  f := storage.ListFilter{}
  if cmd.Flags().Changed("tag") {
      tag, _ := cmd.Flags().GetString("tag")
      if tag == "" {
          return UserErrorf("--tag must not be empty")
      }
      f.Tag = tag
  }
  // ... analogous for project, type
  if cmd.Flags().Changed("since") {
      raw, _ := cmd.Flags().GetString("since")
      t, err := ParseSince(raw)
      if err != nil {
          return UserErrorf("invalid --since %q: %v", raw, err)
      }
      f.Since = t
  }
  if cmd.Flags().Changed("limit") {
      n, _ := cmd.Flags().GetInt("limit")
      if n <= 0 {
          return UserErrorf("--limit must be positive, got %d", n)
      }
      f.Limit = n
  }
  ```

- **`ParseSince` shape.** Pure function. No state, no I/O. ~25 lines:
  ```go
  func ParseSince(s string) (time.Time, error) {
      if t, err := time.Parse("2006-01-02", s); err == nil {
          return t.UTC(), nil
      }
      if len(s) < 2 {
          return time.Time{}, fmt.Errorf("expected YYYY-MM-DD or Nd/Nw/Nm, got %q", s)
      }
      unit := s[len(s)-1]
      nstr := s[:len(s)-1]
      n, err := strconv.Atoi(nstr)
      if err != nil || n <= 0 {
          return time.Time{}, fmt.Errorf("expected positive integer before %c, got %q", unit, nstr)
      }
      var d time.Duration
      switch unit {
      case 'd':
          d = time.Duration(n) * 24 * time.Hour
      case 'w':
          d = time.Duration(n) * 7 * 24 * time.Hour
      case 'm':
          d = time.Duration(n) * 30 * 24 * time.Hour
      default:
          return time.Time{}, fmt.Errorf("unknown unit %q in %q (use d/w/m)", unit, s)
      }
      return time.Now().UTC().Add(-d), nil
  }
  ```

- **Backdating in storage tests.** `Add` sets `created_at` to
  `time.Now().UTC()`. To test `--since` filtering, tests need rows
  with `created_at` in the past. Simplest approach: open a second
  `*sql.DB` in the test against the same temp path and run an
  `UPDATE entries SET created_at = ? WHERE id = ?` statement with
  the backdated RFC3339 string. Self-contained; no Store API
  extension needed.

- **Command `Long` with Examples.** Follow SPEC-005's pattern:
  ```go
  Long: `List brag entries in reverse-chronological order.

  Examples:
    brag list                                       # all entries
    brag list --tag auth                            # entries tagged "auth"
    brag list --project platform --since 7d         # last week, one project
    brag list --type shipped --limit 5              # 5 most recent shipped
    brag list --since 2026-01-01                    # since a specific date`,
  ```

- **DEC-004 Validation note.** Append a bullet like:
  ```
  - Validated during SPEC-007 (2026-04-20): sentinel-comma LIKE
    pattern (`',' || tags || ',' LIKE '%,<tag>,%'`) handles the
    `"auth"` vs `"authoring"` false-positive correctly; no
    normalization needed at MVP scale.
  ```

- **`guidance/questions.yaml` update.** Change
  `tags-storage-model`'s `status: open` to `status: answered`. Add
  a note referencing SPEC-007 and DEC-004's updated Validation
  section. Keep the question in the file as a record; don't delete.

- **Tutorial update.** §4 gets filter examples (copied from the
  command's Examples block works fine). §9's "What's NOT there yet"
  table strikes the `brag list --tag auth` row.

- **No `init()` functions.**

---

## Build Completion

*Filled in at the end of the **build** cycle, before advancing to verify.*

- **Branch:** `feat/spec-007-list-filter-flags`
- **PR (if applicable):** opened after `just advance-cycle`.
- **All acceptance criteria met?** yes. `go test ./...`, `gofmt -l .`,
  `go vet ./...`, `CGO_ENABLED=0 go build ./...` all clean. All 25 new
  tests green alongside every prior SPEC-001…006 test.
- **New decisions emitted:**
  - None. DEC-008 was written during design; no non-trivial build-
    time choice warranted a new DEC.
- **Deviations from spec:**
  - Renamed the CLI test helper to `seedListEntry` to avoid a
    package-level collision with `seedEntry` in SPEC-006's
    `show_test.go` (same `cli` package). The spec suggested
    "`seed…`/`newRootWithList`"; this is a mechanical rename, not a
    semantic change.
  - Added a small `backdateCLI` helper inside `list_test.go` (mirrors
    the storage package's `backdateCreatedAt`). The spec's
    implementer notes mentioned the second-`*sql.DB` technique; I
    inlined the three-line helper rather than exporting anything
    from `Store`.
  - `brag list`'s `Short` string changed from "List all brag entries"
    to "List brag entries" — one-word edit, reads cleaner now that
    the command takes filters. Noted because it's a user-visible
    string change.
- **Follow-up work identified:**
  - No new specs required for STAGE-002. Backlog unchanged.

### Build-phase reflection (3 questions, short answers)

1. **What was unclear in the spec that slowed you down?**
   — Nothing substantive. The `Implementation Context` spelled out
   the WHERE-clause builder, the `Changed()` pattern, and the
   backdating approach; each mapped 1:1 to a block of code. The
   only blip was the helper-name collision with `show_test.go`'s
   `seedEntry`, which the spec couldn't have known about without
   asking me to grep — a two-minute fix.

2. **Was there a constraint or decision that should have been listed but wasn't?**
   — No new one. DEC-008 arriving alongside the spec meant the
   `--since` contract was locked; DEC-007's extension to positional
   args didn't apply here but the flag-validation half did. Every
   blocking constraint I hit (stdout-vs-stderr, timestamps-UTC-RFC3339,
   no-SQL-in-CLI) was already listed.

3. **If you did this task again, what would you do differently?**
   — Run `grep -rn seedEntry internal/cli` before naming a new test
   helper in the same package. It's not a tools-problem — the spec
   even listed `show_test.go` as a prior-art file — I just didn't
   cross-check helper names before writing the file. Cheap lesson.
   Also: the `_ = dbPath` line in `newListTestRoot` is dead (the
   spec's helper kept it from a prior iteration). I left it alone to
   keep the diff minimal but a future spec could drop it.

### Punch-list iteration (2026-04-20)

Verify returned a punch list. Two items, both addressed in this
iteration:

1. **Constraint violation: `database/sql` in `internal/cli/`.** The
   prior build added a `backdateCLI` helper to
   `internal/cli/list_test.go` that called `sql.Open("sqlite", …)`
   directly. The `no-sql-in-cli-layer` constraint (severity:
   blocking, paths: `["internal/cli/**"]`) has no `_test.go`
   exclusion, so the test file violated it. The "Notes for the
   Implementer" sentence saying the second-`*sql.DB` technique was
   "acceptable" was a spec defect that misled the build into the
   simpler, non-conforming path; a spec cannot override a blocking
   constraint.

   Resolution:
   - New sub-package `internal/storage/storagetest/` exporting
     `func Backdate(dbPath string, id int64, at time.Time) error`.
     Lives under `internal/storage/`, so importing it does not pull
     `database/sql` into the CLI test file's import set.
   - `internal/cli/list_test.go` no longer imports `database/sql`;
     `backdateCLI` is replaced by a four-line `mustBackdate` wrapper
     that forwards to `storagetest.Backdate` and `t.Fatal`s on
     error. Verified with `grep -rn database/sql internal/cli/`
     (no matches).
   - `internal/storage/store_test.go` had its own `backdateCreatedAt`
     helper duplicating the same UPDATE; renamed to `mustBackdate`
     and pointed at `storagetest.Backdate` for a single canonical
     implementation.
   - Spec's "Notes for the Implementer" rewritten to prescribe
     `storagetest` and explicitly call out that the prior "either
     is acceptable" wording was a defect.

2. **Mailed-in Q1 of the build reflection.** The original Q1 said
   nothing was unclear, but the deviations enumerated in the same
   block — `seedListEntry` rename, `backdateCLI` helper, `Short`
   tweak, the dead `_ = dbPath` — were themselves friction signals.
   Rewriting Q1 honestly:

   > **Q1 (rewritten).** Three things were friction worth naming.
   > First and largest: the spec's "Either is acceptable" sentence
   > about backdating made the path of least resistance — a
   > CLI-package helper that directly opens `*sql.DB` — feel
   > sanctioned. It wasn't; it violated `no-sql-in-cli-layer` (a
   > blocking constraint with no `_test.go` exclusion). I should
   > have noticed the contradiction during build and stopped to
   > raise a spec defect instead of taking the easy path. Lesson:
   > when a spec says "either A or B is acceptable" and one option
   > would clearly cross a blocking constraint, the spec is wrong
   > and you flag it before writing code. Second: the
   > `seedListEntry` rename (from the spec-suggested `seedEntry`)
   > happened because SPEC-006's `show_test.go` already defined
   > `seedEntry` in the same `cli` package. I caught it via
   > compiler error rather than by reading prior-art tests, even
   > though the spec listed `show_test.go` explicitly. Same-package
   > helper-name collision is now plausible enough across SPEC-002
   > / SPEC-006 / SPEC-007 that a naming convention (e.g. always
   > `seed<Cmd>Entry` / always `mustBackdate`) is worth a future
   > housekeeping spec. Third: I left `_ = dbPath` in
   > `newListTestRoot` as dead code rather than removing both the
   > line and the unused parameter. This punch-list iteration
   > deletes both — `dbPath` was vestigial from an earlier draft
   > that set `BRAGFILE_DB` to it, and Go doesn't require unused
   > parameters to be silenced.

   The other two reflection answers (Q2 on missing
   constraints/decisions; Q3 on what to do differently) stand as
   originally written, with the addition that "grep for the
   helper name in the same package before defining a new one"
   from Q3 also generalizes to "read the prior spec's test files
   completely before writing test scaffolding".

**Files changed in this iteration:**

- `internal/storage/storagetest/storagetest.go` (new) — canonical
  `Backdate` helper.
- `internal/cli/list_test.go` — drop `database/sql` import; drop
  `backdateCLI`; replace with `mustBackdate` wrapping
  `storagetest.Backdate`; drop `_ = dbPath` and the unused
  `dbPath` parameter from `newListTestRoot`; drop two unused
  `dbPath` declarations in the help-only tests.
- `internal/storage/store_test.go` — drop `backdateCreatedAt`;
  replace with `mustBackdate` wrapping `storagetest.Backdate`;
  add storagetest import.
- This spec — Notes for the Implementer rewritten; this
  Punch-list iteration block appended.

**Verification:** `gofmt -l .` empty, `go vet ./...` clean,
`CGO_ENABLED=0 go build ./...` clean, `go test ./...` green
(all prior tests + the SPEC-007 additions still pass).

---

## Reflection (Ship)

*Appended 2026-04-20 during the **ship** cycle. Outcome-focused,
distinct from the process-focused build reflection above.*

1. **What would I do differently next time?**
   Test every implementation option in a spec's Notes for the
   Implementer against the blocking constraints, not just the
   happy-path approach. SPEC-007's Notes said "Either is acceptable"
   for the backdate helper — a CLI-local `*sql.DB` open, OR a
   storage-package helper. The first option violates
   `no-sql-in-cli-layer` (blocking), so it was never actually an
   option. I offered choice where there was only one valid path.
   Verify correctly caught this; the punch-list iteration introduced
   the proper `internal/storage/storagetest` sub-package and fixed
   the spec's Notes inline. Next design session: when tempted to
   offer multiple implementation paths in a spec, mentally run each
   against the constraints list before writing "either is
   acceptable."

2. **Does any template, constraint, or decision need updating?**
   Yes — two things, both applied in this ship commit.
   - Add a new "During design" subsection to `AGENTS.md` §12 with
     one rule: *spec prose cannot relax a blocking constraint; when
     a spec's Notes offer multiple implementation approaches, each
     must pass the constraint list independently*. Promotes the
     SPEC-007 punch-list lesson from an anecdote to an explicit
     design-phase discipline.
   - Minor cosmetic cleanup to `DEC-004`'s Validation section (stray
     blank line / continuation bullet flagged by verify as a nit).

3. **Is there a follow-up spec I should write now before I forget?**
   No. SPEC-008 (`brag delete`) is next pending in STAGE-002 and
   unblocked. SPEC-009 through SPEC-012 (editor package, add-with-
   editor, FTS5, search) remain queued. The AGENTS.md §12 "During
   design" discipline applies prospectively to every remaining
   spec.
