---
stage:
  id: STAGE-002
  status: active
  priority: high
  target_complete: 2026-05-04

project:
  id: PROJ-001
repo:
  id: bragfile

created_at: 2026-04-20
shipped_at: null
---

# STAGE-002: Capture & Retrieval

## What This Stage Is

Turn `brag` from a minimum-viable write/read binary into a daily-use
tool. STAGE-001 shipped `add` and `list`; STAGE-002 fills in the CRUD
gaps (`show`, `edit`, `delete`), adds filter flags on `list`,
introduces full-text search via FTS5, and ships two ergonomic
capture paths: shorthand flags on `add` (for scripting / fast typing)
and an `$EDITOR`-launch form of `add` / `edit` (for narrative
entries with structured metadata). When this stage ships, the author
can manage an accumulating personal brag history entirely within
`brag` — no more `sqlite3 ~/.bragfile/db.sqlite` escape hatch.

## Why Now

Three reasons:

1. **STAGE-001 proved the shape.** `brag add --title "x"` + `brag
   list` works and rows accumulate correctly. The architecture held
   across four specs; the accumulated AGENTS.md lessons are stable.
   Expanding behavior now, on top of that base, is lower-risk than it
   would have been a week ago.
2. **Dogfooding surfaces missing capabilities quickly.** The author
   has already flagged `--title` / `--tags` / ... as "a lot to type",
   and `list` without filters gets unwieldy past ~20 rows. These are
   real usage signals, not speculative features.
3. **STAGE-003 (export/summary) needs richer retrieval.** Grouping a
   week's entries by project for `summary` requires `list` to support
   filters; formatting a Markdown export cleanly requires
   `show`-style per-entry rendering. STAGE-002's retrieval work is
   the foundation STAGE-003 reads from.

## Success Criteria

- `brag add -t "x"` works (shorthand flags available for every `add`
  field; letters are stable and documented in `brag add --help`).
- `brag add` with no args opens `$EDITOR` on a templated markdown
  buffer; saving writes an entry, saving an unchanged/empty buffer
  aborts cleanly.
- `brag list` supports `--tag`, `--project`, `--type`, `--since`, and
  `--limit` filters; tests cover each filter's WHERE-clause behavior
  including the `"auth"` vs `"authoring"` tag substring edge case.
- `brag show <id>` prints the full entry as markdown;
  `brag edit <id>` round-trips through `$EDITOR`;
  `brag delete <id>` removes a row (with confirmation unless
  `--yes`).
- `brag search "query"` returns matching entries via SQLite FTS5.
- Author has used `show`, `list --tag`, or `search` to retrieve a
  past entry at least once per working day for one week (the habit
  signal that retrieval works, not just writing).
- All STAGE-001 success criteria still hold; no regressions in
  existing tests. `go test ./...`, `gofmt -l .`, `go vet ./...`, and
  `CGO_ENABLED=0 go build ./...` all remain clean.

## Scope

### In scope

- **Ergonomic polish on `add`** (shorthand flags, richer help).
- **Editor-launch capture**: `brag add` with no args; `brag edit <id>`
  round-trip. Format for the editable buffer is a stdlib-parsable
  header+body (see Design Notes — leaning `net/textproto`-style
  header block to avoid a YAML dependency).
- **`list` filter flags**: `--tag`, `--project`, `--type`, `--since`,
  `--limit`. `ListFilter` struct gains fields; `Store.List` gains
  WHERE-clause + LIMIT logic. Tag filter handles the `",<tag>,"`
  sentinel-comma trick to avoid substring false-positives.
- **CRUD completion**: `Store.Get(id)`, `Store.Update(id, ...)`,
  `Store.Delete(id)` on the storage side; `show`, `edit`, `delete`
  commands on the CLI side.
- **Full-text search**: `0002_add_fts.sql` migration adds
  `entries_fts` virtual table plus INSERT/UPDATE/DELETE triggers on
  `entries`. `Store.Search(query)` method and `brag search` command.
- **Root help polish**: `brag --help` explicitly tells users to run
  `brag <command> --help` for per-subcommand flag details.

### Explicitly out of scope

- **Export commands** (`brag export --format markdown|sqlite`) —
  STAGE-003.
- **Summary command** (`brag summary --range week|month`) —
  STAGE-003.
- **Any LLM-backed feature** (narrative generation, resume-bullet
  rewriting). Entirely out of PROJ-001.
- **TUI / Bubble Tea frontend** — deferred indefinitely.
- **Schema normalization** (split tags into `tags` + `entry_tags`
  tables). DEC-004 pinned comma-joined for MVP with a revisit
  trigger in the `tags-storage-model` question. STAGE-002 revisits
  that question during SPEC-006 design; if comma-joined still wins,
  we stay with it.
- **Binary distribution** (goreleaser, homebrew tap) — STAGE-004.
- **CI setup** — STAGE-004.
- **Shareable / opaque IDs** (ULID column alongside INTEGER PK).
  DEC-005's revisit question (`shareable-ids`) stays parked until a
  sync or sharing use case appears.
- **Soft delete, undo, trash bin**. `brag delete <id>` is a hard
  delete. If a user asks for recovery, we open a follow-up spec.

## Spec Backlog

Ordered by recommended build sequence (earlier specs unblock later
ones). Complexity mix: 5 × S, 3 × M, no L. Stage size is 8 specs —
at the upper bound of the 3–8 guideline but coherent; no item splits
cleanly to a different stage.

- [x] SPEC-005 (shipped on 2026-04-20, **S**) — **`brag add`
      ergonomic polish.** Shipped single-letter shorthands for all
      six `add` flags (`-t -d -T -p -k -i`), Examples block in `Long`,
      root help pointer to `brag <command> --help`. Approved clean,
      no punch list. Earned a §9 lesson on assertion specificity
      (distinctive tokens, not generic substrings).

- [x] SPEC-006 (shipped on 2026-04-20, **S**) — **`brag show <id>`
      + `Store.Get(id)`.** Cobra subcommand, markdown output (title
      heading + metadata table + optional `## Description`;
      rows/sections omitted when empty). Introduced `storage.
      ErrNotFound` sentinel. Extended DEC-007 to positional-arg
      validation (no `cobra.ExactArgs`). Approved clean, no punch
      list. DEC-007's References section amended to document the
      extension.

- [x] SPEC-007 (shipped on 2026-04-20, **M**) — **`list` filter
      flags + Store filtering.** Shipped `--tag / --project / --type
      / --since / --limit` with AND-combined WHERE clauses and
      tie-break ordering preserved. `--tag` uses sentinel-comma
      pattern — `"auth"` doesn't match `"authoring"`. Introduced
      `cli.ParseSince` + DEC-008 (YYYY-MM-DD + Nd/Nw/Nm). Answered
      DEC-004's `tags-storage-model` question (stays comma-joined).
      One verify punch-list loop on a `no-sql-in-cli-layer`
      violation in a test helper — fixed by introducing
      `internal/storage/storagetest` sub-package. Earned the
      "During design" rule in AGENTS.md §12.

- [x] SPEC-008 (shipped on 2026-04-20, **S**) — **`brag delete
      <id>` + `Store.Delete(id)`.** Shipped y/N confirmation
      prompt + `--yes`/`-y` bypass, hard delete, strict
      stdout-empty stream discipline. Amended api-contract.md:
      decline → exit 0 (deliberate user choice), not exit 1.
      Approved with a yellow-flag note on template field semantics
      (addressed in ship via AGENTS.md §2 note). No new DECs.

- [x] SPEC-009 (shipped on 2026-04-21, **M**) — **`brag edit <id>`
      + `internal/editor` package + `Store.Update`.** THE update
      mechanism for PROJ-001 (flag-based `brag update` deferred).
      Introduced DEC-009 (editor buffer format: `net/textproto`
      header + markdown body; no YAML dep). 34 tests green after
      punch-list iteration added the git `:cq` semantics test.
      Earned a §9 rule: every locked design decision needs a paired
      failing test.

- [x] SPEC-010 (shipped on 2026-04-21, **S**) — **`brag add` no-args
      editor launch.** Shipped runAdd dispatcher + runAddEditor +
      `editor.EmptyTemplate()`. Flag mode byte-identical. 10
      locked decisions / 10 paired tests. Deleted
      `TestAdd_MissingTitleIsUserError` (premise inverted by
      decision #1) — earned the inverse of the §9
      locked-decisions-need-tests rule: removed behavior ↔
      planned test deletion, not a build-time discovery.

- [x] SPEC-011 (shipped on 2026-04-22, **M**) — **FTS5 virtual
      table + triggers.** Shipped `0002_add_fts.sql` with
      external-content FTS5, backfill-in-transaction, and three
      sync triggers. 10 tests green; 7 locked decisions honored.
      One punch-list-adjacent deviation honestly disclosed: the
      premise audit missed `TestOpen_MigrationsTracked`'s literal
      count-of-1 assertion. Earned the additive-invalidation
      corollary to the §9 premise-audit rule (SPEC-010's rule
      handled inversion/removal; SPEC-011's extends it to
      addition). Second deviation flagged as SPEC-012 design
      input: FTS5's `-` operator is binary NOT, so hyphenated
      user queries need phrase-quoting — `brag search` design
      session must decide auto-quote vs raw syntax.

- [ ] SPEC-012 (not yet framed, **S**) — **`brag search "query"` +
      `Store.Search(query)`.** Thin wrapper over FTS5's `MATCH`
      operator. Output format mirrors `list` (tab-separated). Same
      filter flags as `list` if trivial; otherwise they land in a
      later polish spec. **Open design question flagged by
      SPEC-011**: FTS5's `-` operator is binary NOT, so a user
      query like `"auth-refactor"` parses as `auth NOT refactor`
      and returns nothing. SPEC-012 design session must decide
      whether to auto-quote user input (`MATCH '"auth-refactor"'`),
      expose raw FTS5 syntax (power but surprise), or a hybrid.
      Recommend auto-quote for MVP; document the trade-off in the
      spec's Notes and add a DEC if it's non-obvious.

**Count:** 7 shipped / 0 active / 1 pending

**Complexity check:** 5 × S, 3 × M, 0 × L. Stage is at the upper
bound of the 3–8 spec guideline. No split recommended — each spec
maps to a distinct user-visible capability or a single architectural
addition (editor package, FTS5 schema). Build sequence matters:
SPEC-011 must land before SPEC-012; SPEC-009 must land before
SPEC-010; SPEC-007 is useful (not strictly required) before SPEC-009
for the Get round-trip.

## Design Notes

Cross-cutting patterns for specs in this stage. All SPEC-001 / -002 /
-003 / -004 lessons already in AGENTS.md (§9 buffer split, §9 tie-
break, §10 gitignore anchor, §9 fail-first test run) apply
unchanged.

- **Editor-launch buffer format.** Lean toward stdlib `net/textproto`
  header block + blank line + markdown body. Example:
  ```
  Title: cut p99 login latency
  Tags: auth,perf
  Project: platform
  Type: shipped
  Impact: unblocked mobile v3 release

  Replaced the join-on-every-request with a redis lookup.
  ```
  Chosen for: pure stdlib (no YAML dep, honors `no-new-top-level-
  deps-without-decision`), grep-friendly, matches how users type.
  The alternative (`gopkg.in/yaml.v3`) would need a DEC justifying
  the dep. Final call during SPEC-009 design.

- **Editor binary resolution.** `$EDITOR` → `$VISUAL` → `vi` as a
  fallback chain (matches `git`'s convention). One small helper in
  `internal/editor`.

- **`ListFilter` extension.** Keep the struct's zero value meaning
  "no filter" so existing `Store.List(ListFilter{})` callers (from
  STAGE-001) continue to work. New fields added as typed pointers
  OR with an explicit `IsZero`-style method. Decided during SPEC-006
  design.

- **Tag filter false-positive fix.** Use sentinel-comma pattern:
  rows match if `',' || tags || ',' LIKE '%,tag,%'`. Tests must
  include the `"auth"` vs `"authoring"` case. This is the DEC-004
  "revisit during STAGE-002" loop arriving.

- **FTS5 trigger shape.** AFTER INSERT / UPDATE / DELETE triggers on
  `entries` that exec `INSERT INTO entries_fts(...) VALUES(...)`
  (insert), `DELETE FROM entries_fts WHERE rowid = ?` (delete), and
  a delete+insert pair (update). Standard SQLite pattern; document
  in `0002_add_fts.sql`.

- **CLI test harness.** Every new subcommand test uses separate
  `outBuf` / `errBuf` per AGENTS.md §9. Every time-ordering test uses
  the `id DESC` tie-break per AGENTS.md §9. Every build session
  runs `go test ./...` against the failing tests once before writing
  implementation (SPEC-003 ship lesson, validated in SPEC-004).

- **DEC-007 carries forward.** Any new subcommand with required or
  validated flags does validation in `RunE` via `UserErrorf`, never
  `MarkFlagRequired`. This is already the pattern; spec authors
  should not reach for `MarkFlagRequired` without amending DEC-007.

- **Drive-by changes.** SPEC-003 narrowed `.gitignore` as a
  disclosed drive-by (AGENTS.md §10 lesson earned). STAGE-002 specs
  should avoid drive-bys; if one is unavoidable (load-bearing
  scope-adjacent), disclose in Build Completion deviations.

## Dependencies

### Depends on

- **STAGE-001 (shipped 2026-04-20)** — provides `cmd/brag/main.go`,
  `internal/cli/root.go`, `internal/cli/add.go`,
  `internal/cli/list.go`, `internal/cli/errors.go`,
  `internal/config`, `internal/storage` with `Store.Open`, `Add`,
  `List(ListFilter{})`, `entries` table, `0001_initial.sql`
  migration, and `schema_migrations` tracking.
- **DEC-001 / -002 / -003 / -004 / -005 / -006 / -007** — all apply
  forward. No decisions in STAGE-001 have been superseded.
- **External:** none beyond the existing Go toolchain + cobra +
  modernc.org/sqlite. SPEC-009's format choice may introduce one dep
  (`yaml.v3`) — leaning no, per Design Notes.

### Enables

- **STAGE-003 (export/summary).** `Store.Get` powers markdown per-
  entry rendering; `ListFilter` lets `summary` aggregate by
  project/type/range; `Store.Search` could feed a targeted export.
- **STAGE-004 (distribution).** The CLI contract stabilizes once
  STAGE-002 ships — no more "shipped only two commands" caveat —
  making a 1.0-ish release via homebrew defensible.

## Stage-Level Reflection

*Filled in when status moves to shipped. Run Prompt 1d.*

- **Did we deliver the outcome in "What This Stage Is"?** <not yet>
- **How many specs did it actually take?** <not yet>
- **What changed between starting and shipping?** <not yet>
- **Lessons that should update AGENTS.md, templates, or constraints?**
  - <not yet>
- **Should any spec-level reflections be promoted to stage-level lessons?**
  - <not yet>
