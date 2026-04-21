# CLI Contract

`brag` has no HTTP/RPC API. Its external contract is its argv surface
and its stdout/stderr behavior. This doc is the frozen spec of that
contract across PROJ-001: what each command takes, what it prints, and
what it exits with.

## Overview

- **Binary:** `brag` (homebrew formula: `bragfile`).
- **Auth:** none.
- **Versioning:** `brag --version`. Semver. v0.x while the contract is
  still plastic.
- **Exit codes:**
  - `0` — success.
  - `1` — user error (missing arg, no such entry, invalid flag value).
  - `2` — internal error (DB open failed, I/O error, migration failed).

## Global flags (all commands)

| Flag | Env | Default | Purpose |
|---|---|---|---|
| `--db <path>` | `BRAGFILE_DB` | `~/.bragfile/db.sqlite` | DB file location. Parent dir is auto-created. |
| `--version` | — | — | Print version and exit 0. |
| `--help`, `-h` | — | — | Print help and exit 0. |

Precedence: `--db` > `BRAGFILE_DB` > default (DEC-003).

## Commands

### `brag add` — capture a new entry

**STAGE-001 (flags-only form):**

```
brag add --title "shipped the auth refactor" \
         [--description "cut login latency p99 from 600ms to 120ms"] \
         [--tags "auth,perf"] \
         [--project "platform"] \
         [--type "shipped"] \
         [--impact "unblocked mobile v3 release"]
```

- `--title` is required. Everything else is optional.
- Stdout on success: the inserted entry's ID, one line, no prefix (e.g.
  `42`). This keeps piping trivial (`id=$(brag add --title ...)`).
- Exit 0 on success; 1 if `--title` is missing or empty; 2 on storage
  error.

**STAGE-002 (editor-launch form):**

```
brag add            # no args → opens $EDITOR on a template buffer
```

- Template is a markdown file with YAML front-matter holding the
  optional fields (tags/project/type/impact) and a `## Description`
  body section. On save+exit, `brag` parses the buffer, validates
  required fields, and inserts. If the user saves an empty buffer
  (template unchanged), the command aborts with exit 0 and no write.

### `brag list` — list entries

```
brag list [--tag T] [--project P] [--type T] [--since 2026-01-01] [--limit N]
```

- `--tag`, `--project`, `--type` filter on exact field value
  (tags filter uses substring against the comma-joined column in MVP).
- `--since` accepts `YYYY-MM-DD` or a duration like `7d`, `2w`, `3m`.
- `--limit` defaults to unlimited; useful for `brag list --limit 5`.
- Order: `created_at DESC`.
- Output: one line per entry, tab-separated columns:
  `<id>\t<created_at>\t<title>`.
- STAGE-001 ships without filter flags — plumbing exists, flags are
  added in STAGE-002.

### `brag show <id>` — show a single entry (STAGE-002)

```
brag show 42
```

Prints the full entry as markdown (title as `# `, metadata as a small
table, description as body). Exit 1 if the ID does not exist.

### `brag edit <id>` — edit via $EDITOR (STAGE-002)

```
brag edit 42
```

The primary mechanism for updating entries in PROJ-001. Flag-based
update (e.g. `brag update 42 -t "new"`) is a deferred polish spec;
for now, every change goes through the editor round-trip.

`brag` resolves the editor as `$EDITOR` → `$VISUAL` → `vi`, writes
the entry to a temp `.md` file as a `net/textproto`-style header block
plus markdown body (DEC-009), and spawns the editor against it. On
save, it re-parses the buffer, preserves `id` and `created_at`,
bumps `updated_at`, and writes the new field values via
`Store.Update`.

- Saving byte-identical content (SHA-256 comparison) aborts cleanly:
  stderr prints `No changes.`, exit 0, DB untouched.
- Saving a successful edit prints `Updated.` to stderr, exit 0.
- If the saved buffer is missing or has an empty `Title:` header, the
  command exits 1 (user error); the DB is unchanged.
- If the editor exits non-zero (e.g. `:cq` in vim) the command exits
  2 (internal error); the DB is unchanged.
- Missing/non-numeric/non-positive `<id>` or a no-longer-existent
  entry exit 1 (user error).

### `brag delete <id>` — delete an entry (STAGE-002)

```
brag delete 42
brag delete 42 --yes
brag delete 42 -y
```

Prompts for confirmation on stdin unless `--yes` (`-y`) is passed.
Exit 1 if the ID does not exist, the arg is invalid, or missing.
Exit 0 (no error) if the user declines the confirmation prompt — a
deliberate choice, not an error.

### `brag search "query"` — full-text search (STAGE-002)

```
brag search "auth refactor"
brag search "latency" --limit 10
```

FTS5 query against `entries_fts`. Same output shape as `list`. Match
ranking: default FTS5 relevance.

### `brag export` — export entries (STAGE-003)

```
brag export --format markdown                 # stdout: all entries as a markdown report
brag export --format markdown --out report.md # write to file
brag export --format sqlite --out backup.db   # raw DB file copy
```

- `--format` is required. Values: `markdown`, `sqlite`.
- `--out <path>` optional; defaults to stdout for markdown, required
  for sqlite.
- `sqlite` export is a `VACUUM INTO` copy (portable, no WAL-dependent
  state).

### `brag summary --range week|month` (STAGE-003)

```
brag summary --range week
brag summary --range month --project platform
```

Rule-based aggregation. Output: a markdown block grouped by `project`,
then `type`, with counts and entry titles. No LLM. Optional
`--project` filter narrows the set first.

## Error output

All human-readable errors go to stderr with the prefix `brag: `. Example:

```
brag: no entry with id 42 (did you mean `brag list`?)
```

Machine-parseable output is stdout only; stderr is for humans.

## Stability guarantees

- **Pre-1.0:** command names and exit codes are stable within a minor
  version; flag names may change between `v0.x` releases with CHANGELOG
  notes.
- **Post-1.0 (future):** no breaking changes to command names, flag
  names, or exit codes without a major bump.

## References

- Architecture: [./architecture.md](./architecture.md)
- Data model: [./data-model.md](./data-model.md)
- `DEC-003` — config resolution order
- `DEC-009` — editor buffer format (`brag edit <id>`)
