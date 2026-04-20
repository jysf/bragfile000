---
stage:
  id: STAGE-001
  status: active
  priority: high
  target_complete: 2026-04-26

project:
  id: PROJ-001
repo:
  id: bragfile

created_at: 2026-04-19
shipped_at: null
---

# STAGE-001: Foundations

## What This Stage Is

Stand up the skeleton of the `brag` CLI with enough machinery that
an entry can be written to and read from local persistent storage.
When this stage ships, running `brag add --title "..."` inserts a row
into `~/.bragfile/db.sqlite` (auto-created), and `brag list` prints it
back out. Everything after this stage is additive capability on top of
the same storage layer.

## Why Now

This is the first stage of the first project. Nothing else in PROJ-001
can be built without: (a) a binary that compiles, (b) a place to put
entries, and (c) proof both directions of the core loop (write → read)
work end-to-end. The minute this ships, we can dogfood capture on any
other ticket while STAGE-002 builds out the richer commands.

## Success Criteria

- `go build ./cmd/brag` produces a working binary on macOS arm64.
- `brag --version` and `brag --help` both work; root help lists `add`
  and `list` as available commands.
- `brag add --title "first brag"` inserts a row and prints the row's
  ID. Running the command twice with the same title creates two
  distinct entries (no implicit dedup).
- `brag list` prints all entries in reverse chronological order,
  one per line, including ID, created_at, and title.
- `~/.bragfile/db.sqlite` is auto-created on first run with correct
  schema; re-running `add` does not attempt to re-apply migrations.
- Unit tests cover the storage layer (open, migrate, add, list) using
  `t.TempDir()`; commands have at least one happy-path test each.
- `just test` (or `go test ./...`) passes green.

## Scope

### In scope
- Go module init at `github.com/jysf/bragfile000`.
- `cmd/brag/main.go` + root `cobra.Command`, subcommands for `add` and
  `list` only.
- `internal/storage` package with:
  - `Open(path string) (*Store, error)` that creates parent dir and
    applies embedded migrations.
  - One migration file: initial `entries` table plus
    `schema_migrations` tracking table.
  - `Add(entry)` and `List(filter)` methods (filter may be a zero
    value for now — actual filter flags land in a later stage).
- Sensible default DB path (`~/.bragfile/db.sqlite`) with
  `BRAGFILE_DB` env var override; `--db` CLI flag at root.
- Minimal output formatter for `list` (plain text, one row per line).
- Unit tests for storage (temp dir) and thin command tests.
- Update `AGENTS.md` Section 4 (commands) so later stages have real
  `build/test/lint` invocations.

### Explicitly out of scope
- `search`, `show`, `edit`, `delete`, `export`, `summary` commands —
  all in later stages.
- Filter flags on `list` (`--tag`, `--project`, etc.) — the *plumbing*
  of a filter struct exists; the *flags* land in STAGE-002.
- `$EDITOR` launch path — STAGE-002.
- FTS5 virtual table — STAGE-002.
- Any TUI or styled output.
- goreleaser, homebrew tap, cross-platform release — STAGE-004.
- Structured logging, observability, metrics.
- `entries.tags` storage format decisions beyond "a comma-joined
  string for MVP" — we'll revisit if search needs it in STAGE-002.

## Spec Backlog

- [ ] SPEC-001 (build) — Go module + Cobra scaffold (S). Module init,
      root command, `cmd/brag/main.go`, package layout, `--version`,
      `--help`, root `--db` flag wiring, `internal/config` path
      resolver per DEC-003.
- [ ] SPEC-002 (design) — SQLite storage + migrations (M). Pure-Go
      sqlite driver, `internal/storage/store.go`, embedded migration
      SQL, `entries` + `schema_migrations` tables, `Open`, `Add`,
      `List(filter)` stub, full test coverage on temp dir.
- [ ] SPEC-003 (design) — `brag add` command (S). Cobra subcommand,
      required `--title`, optional `--description --tags --project
      --type --impact`, writes via storage, prints inserted ID.
- [ ] SPEC-004 (design) — `brag list` command (S). Cobra subcommand,
      no filter flags yet, reads all rows reverse-chronological,
      plain-text output, test asserts output shape.

**Count:** 0 shipped / 0 active / 4 pending

**Complexity check:** no L specs. SPEC-002 is the only M; the others
are S. Stage size is healthy (4 specs, within the 3–8 guideline).

## Design Notes

All load-bearing decisions for this stage are now formalized as
`DEC-*` files (emitted during Prompt 2a — repo/project design). This
section points to them and adds a few stage-local details not worth
their own DEC.

- **SQLite driver:** `modernc.org/sqlite` (pure Go, no CGO).
  See `DEC-001`.
- **Migrations:** embedded numbered `.sql` files under
  `internal/storage/migrations/`, applied in lexical order by
  `storage.Open`, tracked in `schema_migrations`. See `DEC-002`.
- **Config resolution:** CLI flag `--db` → env `BRAGFILE_DB` →
  default `~/.bragfile/db.sqlite`. Parent dir auto-created.
  See `DEC-003`.
- **Tags field:** single `TEXT` column, comma-joined for MVP.
  See `DEC-004` (MVP tradeoff, revisit in STAGE-002).
- **IDs:** `INTEGER PRIMARY KEY AUTOINCREMENT`. See `DEC-005`.
- **Error handling:** wrap with `fmt.Errorf("<op>: %w", err)`. No
  custom error types this stage. See `errors-wrap-with-context`
  constraint.
- **Timestamps:** RFC3339 UTC `TEXT` column, written from Go (not
  SQLite `CURRENT_TIMESTAMP`). See `timestamps-in-utc-rfc3339`
  constraint.
- **Testing pattern:** storage tests take a `*testing.T` and use
  `t.TempDir()` for DB path (enforced by
  `storage-tests-use-tempdir` constraint). Command tests construct
  a `*cobra.Command` with an in-memory `bytes.Buffer` for stdout
  and assert on that buffer.

## Dependencies

### Depends on

- No prior stages. First stage of first project.
- External: none needed beyond standard Go toolchain (1.26.x),
  `cobra`, `modernc.org/sqlite`.

### Enables

- STAGE-002 (capture & retrieval) — needs the storage layer and
  `add`/`list` as the base it layers `show`/`edit`/`delete`/`search`
  onto.
- STAGE-003 (export & summary) — needs entries to export.
- STAGE-004 (distribution) — needs a buildable binary to ship.

## Stage-Level Reflection

*Filled in when status moves to shipped.*

- **Did we deliver the outcome in "What This Stage Is"?** <not yet>
- **How many specs did it actually take?** <not yet>
- **What changed between starting and shipping?** <not yet>
- **Lessons that should update AGENTS.md, templates, or constraints?**
  - <not yet>
- **Should any spec-level reflections be promoted to stage-level lessons?**
  - <not yet>
