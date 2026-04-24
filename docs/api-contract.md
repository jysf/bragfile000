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
brag add            # no entry-field flags → opens $EDITOR on a template buffer
```

Dispatch rule: if any of `--title/-t`, `--description/-d`, `--tags/-T`,
`--project/-p`, `--type/-k`, `--impact/-i` is set, `brag add` runs in
flag mode (above). Otherwise it runs in editor mode. The persistent
`--db` flag is a path override, not an entry field — `brag add --db
/tmp/x.db` still opens the editor.

`brag` resolves the editor as `$EDITOR` → `$VISUAL` → `vi`, writes a
template (a `net/textproto`-style header block plus blank-line
separator plus empty body — see DEC-009 for the format) to a temp `.md`
file, and spawns the editor against it. On save:

- Successful save with a valid `Title:` header → row inserted; the
  inserted ID is printed to stdout (mirrors flag-mode contract so
  `id=$(brag add)` works); stderr empty; exit 0.
- Saving byte-identical content (SHA-256 comparison — i.e. the
  template was not modified) aborts cleanly: stderr prints
  `Aborted.`, exit 0, DB untouched.
- Saving a buffer that fails parse (missing or empty `Title:` header)
  exits 1 (user error); the DB is unchanged.
- Editor exec failure (e.g. `:cq` in vim with a modified buffer)
  exits 2 (internal error); the DB is unchanged.

### `brag list` — list entries

```
brag list [-P|--show-project] [--tag T] [--project P] [--type T] [--since 2026-01-01] [--limit N]
```

- `--tag`, `--project`, `--type` filter on exact field value
  (tags filter uses substring against the comma-joined column in MVP).
- `--since` accepts `YYYY-MM-DD` or a duration like `7d`, `2w`, `3m`.
- `--limit` defaults to unlimited; useful for `brag list --limit 5`.
- Order: `created_at DESC`.
- Output: one line per entry, tab-separated columns:
  `<id>\t<created_at>\t<title>`.
- With `-P` / `--show-project`: output becomes
  `<id>\t<created_at>\t<project>\t<title>` (four tab-separated
  columns). Empty project renders as `-`.
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
brag search "auth-refactor"
brag search "latency" --limit 10
```

FTS5 query against `entries_fts`. Query semantics are locked by
[DEC-010](../decisions/DEC-010-search-query-syntax.md): the CLI
tokenizes the argument on whitespace, phrase-quotes each token, and
joins with spaces so multi-word queries get AND semantics and
hyphens / other FTS5 operators inside a token are treated as
literal text. Power-user FTS5 operators are not exposed.

- Takes a single positional query argument. Zero or multiple args
  exit 1.
- Empty / whitespace-only / quote-containing queries exit 1.
- Same output shape as `list`: tab-separated
  `<id>\t<created_at>\t<title>` to stdout, newline-terminated.
- Order: FTS5 `rank` ascending (most relevant first), with `id DESC`
  as the tie-break for determinism when ranks are equal (DEC-005).
- `--limit N` caps the result count; `0` (the default) means
  unlimited. Negative values exit 1.
- Zero results is not an error: stdout empty, stderr empty, exit 0.
- Exit codes: `0` on success OR zero results; `1` on user-facing
  input problems (empty/quote/arg count/bad `--limit`); `2` on
  storage failure.

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
- `DEC-010` — `brag search` query syntax (auto-tokenize + phrase-quote)
