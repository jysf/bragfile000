# Data Model

The entire bragfile schema lives in a single SQLite database file,
default path `~/.bragfile/db.sqlite`. Two tables ship in v0.1: `entries`
(the user's brags) and `schema_migrations` (applied-version tracking).
STAGE-002 adds a third virtual table (`entries_fts`, SPEC-011) for
FTS5 search. STAGE-006 adds `tags` and `taggings` (SPEC-025). STAGE-007
adds `projects` and `project_locations` (SPEC-027).

## Entities

### Entity: `entries`

One row per brag-worthy moment the user captures.

| Field | Type | Nullable | Default | Description |
|---|---|---|---|---|
| `id` | INTEGER PRIMARY KEY AUTOINCREMENT | no | auto | Opaque stable ID; shown in CLI output and used as the argument to `show`, `edit`, `delete`. |
| `title` | TEXT NOT NULL | no | — | Short headline. Required. |
| `description` | TEXT | yes | NULL | Free-form markdown body. The "narrative" for retros/resumes. |
| `project` | TEXT | yes | NULL | Work project / client / initiative this brag belongs to. |
| `type` | TEXT | yes | NULL | Free-form category (`"shipped"`, `"learned"`, `"mentored"`, …). No enum yet. |
| `impact` | TEXT | yes | NULL | Free-form impact statement (metric, quote, outcome). |
| `created_at` | TEXT NOT NULL | no | — | RFC3339 UTC, written by the Go layer, not by SQLite. |
| `updated_at` | TEXT NOT NULL | no | — | RFC3339 UTC, set equal to `created_at` on insert, bumped on every `edit`. |

**Relationships:** tags are stored in the normalized `tags` / `taggings`
join (DEC-015, SPEC-025). The `Entry.Tags` field projected by the read
path is the comma-joined, position-ordered reconstruction via
`GROUP_CONCAT`; it is not a column on `entries`.

### Entity: `tags` (shipped in SPEC-025)

One row per unique tag name.

| Field | Type | Nullable | Default | Description |
|---|---|---|---|---|
| `id` | INTEGER PRIMARY KEY AUTOINCREMENT | no | auto | Opaque stable ID. |
| `name` | TEXT NOT NULL UNIQUE | no | — | Canonical tag name (whitespace-trimmed, deduped on write). |

### Entity: `taggings` (shipped in SPEC-025)

Polymorphic membership join: one row per (tag × taggable object) pair.
Currently only `taggable_type = 'entry'` is used.

| Field | Type | Nullable | Default | Description |
|---|---|---|---|---|
| `id` | INTEGER PRIMARY KEY AUTOINCREMENT | no | auto | Opaque stable ID. |
| `tag_id` | INTEGER NOT NULL REFERENCES tags(id) | no | — | Foreign key into `tags`. |
| `taggable_type` | TEXT NOT NULL | no | — | Object type discriminator (currently always `'entry'`). |
| `taggable_id` | INTEGER NOT NULL | no | — | ID of the tagged object in its source table. |
| `position` | INTEGER NOT NULL | no | — | 0-based order within an entry's tag list; drives `ORDER BY` in the `GROUP_CONCAT` projection. |

Unique constraint: `(taggable_type, taggable_id, tag_id)` — one tag per entry, once.

Tag rename mutates this join indirectly via the `tags_au` trigger (renames the `tags` row); tag merge mutates it directly via DELETE+INSERT (DEC-016 choice 3).

Indexes: `idx_taggings_tag (tag_id)`, `idx_taggings_taggable (taggable_type, taggable_id)`.

### Entity: `projects` (shipped in SPEC-027)

One row per registered first-class project (DEC-017).

| Field | Type | Nullable | Default | Description |
|---|---|---|---|---|
| `id` | INTEGER PRIMARY KEY AUTOINCREMENT | no | auto | Opaque stable ID (DEC-005). |
| `name` | TEXT NOT NULL UNIQUE | no | — | Project name. Globally unique. Matches `entries.project` via soft string equality (DEC-017). |
| `status` | TEXT NOT NULL | no | `'active'` | Store-validated enum: `active` \| `paused` \| `done` \| `archived`. No DB CHECK — mirrors `entries.type` free-text approach (DEC-017). |
| `state_note` | TEXT NOT NULL | no | `''` | Single free-text state/next-action note rendered by `brag project status` (SPEC-030). |
| `created_at` | TEXT NOT NULL | no | — | RFC3339 UTC, written by the Go layer. |
| `updated_at` | TEXT NOT NULL | no | — | RFC3339 UTC, set equal to `created_at` on insert, bumped on mutations (SPEC-029). |

**Relationship to `entries`:** DEC-017 soft string match — `entries.project` is free text joined to `projects.name` opportunistically at query time. No FK, no link column, no backfill.

### Entity: `project_locations` (shipped in SPEC-027)

One row per filesystem path registered to a project. One project may have
many directories; each path maps to at most one project (globally unique).

| Field | Type | Nullable | Default | Description |
|---|---|---|---|---|
| `id` | INTEGER PRIMARY KEY AUTOINCREMENT | no | auto | Opaque stable ID; drives insertion-order hydration of `Project.Locations`. |
| `project_id` | INTEGER NOT NULL REFERENCES projects(id) | no | — | Foreign key into `projects`. |
| `path` | TEXT NOT NULL UNIQUE | no | — | Filesystem path, stored verbatim (SPEC-031 owns normalization). Globally unique — the guarantee `brag project here` relies on. |

Index: `idx_project_locations_project (project_id)`.

### Entity: `schema_migrations`

Tracks which migration files have been applied.

| Field | Type | Nullable | Default | Description |
|---|---|---|---|---|
| `version` | TEXT PRIMARY KEY | no | — | Filename stem of the applied migration (e.g. `0001_initial`). |
| `applied_at` | TEXT NOT NULL | no | — | RFC3339 UTC timestamp of application. |

### Virtual: `entries_fts` (shipped in SPEC-011; re-pointed in SPEC-025)

Regular (own-content) FTS5 virtual table indexing `title`, `description`,
`tags`, `project`, `impact`. Stores its own copies of indexed values
(no `content=` / `content_rowid=` — those were removed in SPEC-025 when
the `entries.tags` column was dropped and the trigger topology was
replaced). Uses the default unicode61 tokenizer.

**Trigger topology (post-SPEC-025, `0003_normalize_tags.sql`):**

| Trigger | Event | Action |
|---|---|---|
| `entries_ai` | AFTER INSERT ON entries | Inserts the FTS row with the `GROUP_CONCAT` tag projection. |
| `entries_au` | AFTER UPDATE ON entries | Updates FTS non-tag fields (`title`, `description`, `project`, `impact`). |
| `entries_ad` | AFTER DELETE ON entries | Deletes the FTS row. |
| `taggings_ai` | AFTER INSERT ON taggings (entry only) | Re-computes and updates `tags` in the FTS row for the affected entry. |
| `taggings_ad` | AFTER DELETE ON taggings (entry only) | Same. |
| `tags_au` | AFTER UPDATE ON tags | Updates `tags` in every FTS row that carries the renamed tag. |

`Store.Add`, `Update`, and `Delete` do not issue any FTS writes directly —
all FTS maintenance flows through these triggers.

Backfill: the `0002_add_fts.sql` migration seeds FTS from the entries
table; the `0003_normalize_tags.sql` migration drops and recreates
`entries_fts` as a regular table, then re-seeds it from the normalized
join projection.

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

Shipped in STAGE-006 (SPEC-025):
- `tags(id, name UNIQUE)` and `taggings(id, tag_id, taggable_type, taggable_id, position)`.
- Indexes `idx_taggings_tag` and `idx_taggings_taggable`.
- FTS5 re-pointed to own-content; 3 new triggers (`taggings_ai`, `taggings_ad`, `tags_au`).

Shipped in STAGE-007 (SPEC-027):
- `projects(id, name UNIQUE, status, state_note, created_at, updated_at)` — first-class projects entity (DEC-017).
- `project_locations(id, project_id, path UNIQUE)` — one-project-many-directories join.
- Index `idx_project_locations_project (project_id)`.

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

These are noted so readers don't treat the current schema as the end state.
None of them land in PROJ-001.

- ~~**Projects normalization.**~~ Shipped as first-class `projects` + `project_locations` entity in STAGE-007 (SPEC-027). See DEC-017 for the `entries.project` relationship model.
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
- `DEC-004` — comma-joined tags (MVP; superseded by DEC-015 / SPEC-025)
- `DEC-015` — normalized tag storage (`tags` + `taggings` join, polymorphic; supersedes DEC-004)
- `DEC-016` — tag mutation semantics: `brag tags` in-use-only taxonomy, rename-errors-into-existing, merge via DELETE+INSERT de-dup, orphan tags invisible (no GC)
- `DEC-017` — `entries.project` ↔ `projects` relationship: soft string match (free text, opportunistic join on `projects.name`, zero backfill); `projects.status` Store-validated enum; single `state_note` free-text column
- `DEC-005` — INTEGER auto-increment primary keys (MVP)
- `DEC-011` — shared JSON output shape for `brag list --format json` and `brag export --format json`: 9-key naked array mirroring the `entries` column names in order (`id, title, description, tags, project, type, impact, created_at, updated_at`).
- `DEC-013` — markdown export shape for `brag export --format markdown`: level-1 document heading, provenance block, `**By type**` / `**By project**` summary, entries grouped under `## <project>` (alphabetical-ASC; `(no project)` last) with within-group chronological-ASC ordering; `--flat` swaps the grouping for a single `## Entries (chronological)` wrapper.
- `DEC-012` — stdin-JSON schema for `brag add --json`: user-owned fields only (title required; description, tags, project, type, impact optional free-form text); server-owned fields (id, created_at, updated_at) tolerated-and-ignored; unknown keys strict-rejected.
- `DEC-014` — rule-based output shape for `brag summary`, `brag review`, and `brag stats`: single-object JSON envelope with `generated_at` / `scope` / `filters` provenance plus per-spec payload keys at top level; markdown convention reuses DEC-013's `Generated:` / `Scope:` / `Filters:` provenance lines and `**By type**` / `**By project**` count style. Counts maps render alphabetical-ASC by key in JSON (Go's encoder) but DESC-by-count in markdown — documented asymmetry.
