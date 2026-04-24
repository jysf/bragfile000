# Data Model

The entire bragfile schema lives in a single SQLite database file,
default path `~/.bragfile/db.sqlite`. Two tables ship in v0.1: `entries`
(the user's brags) and `schema_migrations` (applied-version tracking).
STAGE-002 adds a third virtual table (`entries_fts`, SPEC-011) for
FTS5 search.

## Entities

### Entity: `entries`

One row per brag-worthy moment the user captures.

| Field | Type | Nullable | Default | Description |
|---|---|---|---|---|
| `id` | INTEGER PRIMARY KEY AUTOINCREMENT | no | auto | Opaque stable ID; shown in CLI output and used as the argument to `show`, `edit`, `delete`. |
| `title` | TEXT NOT NULL | no | — | Short headline. Required. |
| `description` | TEXT | yes | NULL | Free-form markdown body. The "narrative" for retros/resumes. |
| `tags` | TEXT | yes | NULL | Comma-joined list for MVP (DEC-004). Empty string and NULL both mean "no tags". |
| `project` | TEXT | yes | NULL | Work project / client / initiative this brag belongs to. |
| `type` | TEXT | yes | NULL | Free-form category (`"shipped"`, `"learned"`, `"mentored"`, …). No enum yet. |
| `impact` | TEXT | yes | NULL | Free-form impact statement (metric, quote, outcome). |
| `created_at` | TEXT NOT NULL | no | — | RFC3339 UTC, written by the Go layer, not by SQLite. |
| `updated_at` | TEXT NOT NULL | no | — | RFC3339 UTC, set equal to `created_at` on insert, bumped on every `edit`. |

**Relationships:** none. This is a single-table schema for MVP. Tag
normalization into its own table is a deferred option (see
"Schema Evolution" below).

### Entity: `schema_migrations`

Tracks which migration files have been applied.

| Field | Type | Nullable | Default | Description |
|---|---|---|---|---|
| `version` | TEXT PRIMARY KEY | no | — | Filename stem of the applied migration (e.g. `0001_initial`). |
| `applied_at` | TEXT NOT NULL | no | — | RFC3339 UTC timestamp of application. |

### Virtual: `entries_fts` (shipped in SPEC-011)

External-content FTS5 virtual table indexing `title`, `description`,
`tags`, `project`, `impact` from `entries`. Does not store the content
itself — only an inverted index referencing `entries` by `id` via
`content_rowid='id'`. Kept in sync by AFTER INSERT / UPDATE / DELETE
triggers on `entries` (`entries_ai`, `entries_au`, `entries_ad`), so
`Store.Add`, `Update`, and `Delete` stay unchanged. Uses the default
unicode61 tokenizer, which splits on punctuation — comma-joined tags
(DEC-004) tokenize one-tag-per-token for free.

Backfill: the `0002_add_fts.sql` migration issues an
`INSERT INTO entries_fts ... SELECT ... FROM entries` inside its
transaction, so users upgrading with an existing DB get their rows
indexed on the next `storage.Open`. On a fresh DB the SELECT returns
zero rows and the backfill is a no-op.

Shipped in `internal/storage/migrations/0002_add_fts.sql`:

```sql
CREATE VIRTUAL TABLE entries_fts USING fts5(
    title, description, tags, project, impact,
    content='entries',
    content_rowid='id'
);
```

Consumed by `brag search` (SPEC-012, shipped 2026-04-22) via the
`Store.Search(query, limit)` method. Query syntax is locked by
DEC-010 (tokenize + phrase-quote + AND-join); see the DEC for the
auto-quoting rationale and the rejected alternatives.

## Schema Evolution

- **Migrations are numbered `.sql` files** in
  `internal/storage/migrations/`, embedded into the binary via
  `embed.FS` (DEC-002).
- **Filenames are `NNNN_short_name.sql`** (`0001_initial.sql`,
  `0002_add_fts.sql`, …). Applied in lexical order on `storage.Open`.
- **Each migration runs in a single transaction** together with the
  `INSERT INTO schema_migrations`. If any statement fails the whole
  thing rolls back.
- **No down-migrations.** If a change needs reversal, write a new
  forward migration.
- **Backward compatibility:** once shipped, a migration file is never
  edited. Errors only get fixed by a follow-up migration.

## Indexes

Planned for STAGE-001 (ship with the initial migration):

- `CREATE INDEX idx_entries_created_at ON entries(created_at DESC);`
  — `brag list` default ordering.
- `CREATE INDEX idx_entries_project ON entries(project);`
  — supports `list --project=...` filter (flags land in STAGE-002).

Shipped in STAGE-002 (SPEC-011):
- `entries_fts` FTS5 virtual table (see above).
- AFTER INSERT / UPDATE / DELETE triggers on `entries`
  (`entries_ai`, `entries_au`, `entries_ad`) maintaining `entries_fts`.

No index on `tags` — the comma-joined format doesn't benefit from one.
If tag filtering becomes a hot path, the migration to a
`tags`/`entry_tags` join pair is the answer (see below), not an index.

## Data Lifecycle

- **Create.** `brag add` inserts a single row. `created_at` and
  `updated_at` are both set to `time.Now().UTC().Format(time.RFC3339)`
  in Go.
- **Read.** `brag list`, `show`, `search` are pure reads.
- **Update.** `brag edit` (STAGE-002) opens `$EDITOR` on a round-tripped
  markdown view, reparses on save, `UPDATE`s the row, bumps
  `updated_at`.
- **Delete.** `brag delete <id>` issues a `DELETE`. No soft-delete
  column and no audit trail in v0.1 — the SQLite file is the backup.
- **Archive / retention.** None. All rows live forever. The user can
  `brag export` their full DB as a Markdown report or as a raw SQLite
  file copy.
- **Backup.** Copy the file. That is the supported mechanism.

## Future schema shapes (deferred, not in PROJ-001)

These are noted so readers don't treat the MVP schema as the end state.
None of them land in PROJ-001.

- **Tags normalization.** Add `tags(id, name)` and `entry_tags(entry_id,
  tag_id)`. Justified if tag filtering gets slow, users want tag
  rename, or tag auto-complete. Reason for deferring: MVP users filter
  tags via LIKE across a few hundred rows; that's fast enough.
- **Projects normalization.** Same shape, same triggers. Deferred for
  the same reason.
- **Type enum / taxonomy.** If we ever want structured types, they
  become their own table. Free-form text is fine for now.
- **Soft delete.** `deleted_at` column + read-path filter. Only worth
  it if users ask for it.
- **Attachments.** Links to external files (PRs, screenshots). A
  `attachments(entry_id, kind, url)` table is the likely shape.

## References

- Architecture: [./architecture.md](./architecture.md)
- CLI surface: [./api-contract.md](./api-contract.md)
- `DEC-002` — embedded migrations
- `DEC-004` — comma-joined tags (MVP)
- `DEC-005` — INTEGER auto-increment primary keys (MVP)
- `DEC-011` — shared JSON output shape for `brag list --format json` and `brag export --format json`: 9-key naked array mirroring the `entries` column names in order (`id, title, description, tags, project, type, impact, created_at, updated_at`).
- `DEC-013` — markdown export shape for `brag export --format markdown`: level-1 document heading, provenance block, `**By type**` / `**By project**` summary, entries grouped under `## <project>` (alphabetical-ASC; `(no project)` last) with within-group chronological-ASC ordering; `--flat` swaps the grouping for a single `## Entries (chronological)` wrapper.
- `DEC-012` — stdin-JSON schema for `brag add --json`: user-owned fields only (title required; description, tags, project, type, impact optional free-form text); server-owned fields (id, created_at, updated_at) tolerated-and-ignored; unknown keys strict-rejected.
