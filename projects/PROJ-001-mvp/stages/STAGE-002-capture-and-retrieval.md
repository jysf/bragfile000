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

- [ ] SPEC-006 (not yet framed, **M**) — **`list` filter flags + Store
      filtering.** Populate `ListFilter` with `Tag`, `Project`, `Type`,
      `Since`, `Limit` fields. Extend `Store.List` to build WHERE
      clauses and apply `LIMIT`. Wire `--tag / --project / --type /
      --since / --limit` on the `list` command. `--since` accepts
      `YYYY-MM-DD` or durations (`7d`, `2w`, `3m`). Tag filter uses
      sentinel-comma (`",<tag>,"`) to avoid the `"auth"` vs
      `"authoring"` false-positive flagged in DEC-004. Revisits the
      `tags-storage-model` question in `/guidance/questions.yaml` —
      either keeps comma-joined (answer the question with a
      supersession note on DEC-004) or proposes normalization as a
      separate spec. Design session decides before build.

- [ ] SPEC-007 (not yet framed, **S**) — **`brag show <id>` +
      `Store.Get(id)`.** Thin cobra subcommand + storage method.
      Prints the entry as markdown: `# <title>`, a small metadata
      table, `## Description` body. Exit 1 if the ID does not exist
      (user error; returns via `ErrUser`).

- [ ] SPEC-008 (not yet framed, **S**) — **`brag delete <id>` +
      `Store.Delete(id)`.** Prompts for confirmation on stdin unless
      `--yes` is passed. Exit 1 if ID missing or the user declines.
      Storage method is a one-liner `DELETE FROM entries WHERE id =
      ?`. Hard delete; no soft-delete column.

- [ ] SPEC-009 (not yet framed, **M**) — **`internal/editor` package
      + `brag edit <id>` + `Store.Update(id, Entry)`.** Introduces the
      editable buffer format (leaning stdlib `net/textproto` header +
      markdown body — no YAML dep — decided during spec design).
      Round-trip: `edit <id>` reads the entry via `Store.Get`,
      renders to the editor format, spawns `$EDITOR`, parses on save,
      writes via `Store.Update` with an updated `updated_at`.
      Unsaved / unchanged buffer aborts cleanly. Likely emits a DEC
      for the chosen buffer format.

- [ ] SPEC-010 (not yet framed, **S**) — **`brag add` no-args editor
      launch.** Reuses `internal/editor` from SPEC-009. `brag add`
      with no flags opens `$EDITOR` on an empty template; save writes
      via `Store.Add`. If the buffer is unchanged/empty, the command
      aborts with exit 0 and no write.

- [ ] SPEC-011 (not yet framed, **M**) — **FTS5 virtual table +
      triggers.** New migration `0002_add_fts.sql` creates
      `entries_fts` mirroring `title, description, tags, project,
      impact`; adds AFTER INSERT / UPDATE / DELETE triggers on
      `entries` that keep `entries_fts` in sync. No CLI change in
      this spec — isolates the schema move. Tests: migration runs on
      a DB that has existing rows (backfill path); triggers keep FTS
      synchronized across CRUD.

- [ ] SPEC-012 (not yet framed, **S**) — **`brag search "query"` +
      `Store.Search(query)`.** Thin wrapper over FTS5's `MATCH`
      operator. Output format mirrors `list` (tab-separated). Same
      filter flags as `list` if trivial; otherwise they land in a
      later polish spec.

**Count:** 1 shipped / 0 active / 7 pending

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
