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

Dispatch rule: if `--json` is set, runs in json mode (reads stdin).
Else if any of `--title/-t`, `--description/-d`, `--tags/-T`,
`--project/-p`, `--type/-k`, `--impact/-i` is set, runs in flag mode.
Otherwise runs in editor mode. The persistent `--db` flag is a path
override, not an entry field, so `brag add --db /tmp/x.db` still opens
the editor. `--json` combined with any field flag exits 1 (user
error).

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

**STAGE-003 (JSON stdin form):**

```
echo '{"title":"shipped"}' | brag add --json
brag list --format json | jq '.[0]' | brag add --json
```

- `--json` reads a single JSON object from stdin and inserts it.
  Required: `title` (non-empty). Optional: `description`, `tags`,
  `project`, `type`, `impact` — all free-form text. Server-owned
  fields (`id`, `created_at`, `updated_at`) are tolerated-and-ignored
  if present, so `brag list --format json | jq '.[0]' | brag add
  --json` round-trips without `jq del`.
- Unknown keys are strict-rejected with the offending key named in
  the error (catches typos like `"titl"` before they become silently-
  missing entries).
- `tags` stays a comma-joined string per
  [DEC-004](../decisions/DEC-004-tags-comma-joined-for-mvp.md); array
  form (`["a","b"]`) is rejected with an error naming DEC-004.
- `--json` is mutually exclusive with field flags (`--title`,
  `--description`, `--tags`, `--project`, `--type`, `--impact`).
  Combining them exits 1.
- Stdout on success: the inserted ID, one line, no prefix (same as
  flag-mode contract).
- Schema lock:
  [DEC-012](../decisions/DEC-012-brag-add-json-stdin-schema.md).

### `brag list` — list entries

```
brag list [-P|--show-project] [--format json|tsv] [--tag T] [--project P] [--type T] [--since 2026-01-01] [--limit N]
```

- `--tag`, `--project`, `--type` filter on exact field value
  (tags filter uses substring against the comma-joined column in MVP).
- `--since` accepts `YYYY-MM-DD` or a duration like `7d`, `2w`, `3m`.
- `--limit` defaults to unlimited; useful for `brag list --limit 5`.
- Order: `created_at DESC`.
- Output (default / plain mode): one line per entry, tab-separated
  columns: `<id>\t<created_at>\t<title>`.
- With `-P` / `--show-project`: output becomes
  `<id>\t<created_at>\t<project>\t<title>` (four tab-separated
  columns). Empty project renders as `-`.
- With `--format json`: pretty-printed JSON array, one object per
  entry, 9 keys per object (`id, title, description, tags, project,
  type, impact, created_at, updated_at`). Shape locked by
  [DEC-011](../decisions/DEC-011-json-output-shape.md); same bytes
  as `brag export --format json` on the same rows.
- With `--format tsv`: header row
  (`id\ttitle\tdescription\ttags\tproject\ttype\timpact\tcreated_at\tupdated_at`)
  followed by one data row per entry. Field order matches `--format
  json`. Empty fields render as the empty string between tabs (no
  dash filler — that is plain-mode `-P` behavior only).
- Unknown `--format` values exit 1 (user error).
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
brag export --format json                         # stdout: JSON array
brag export --format json --out entries.json      # write to file
brag export --format json --project platform      # filter before exporting
brag export --format json --tag auth --since 30d

brag export --format markdown                     # stdout: grouped markdown
brag export --format markdown --flat              # stdout: flat chronological
brag export --format markdown --out report.md     # write to file
brag export --format markdown --project platform  # filter before exporting
```

- `--format` is required. Accepted values: `json`, `markdown`.
- `--out <path>` optional; defaults to stdout. If set, overwrites any
  existing file at the path without prompting.
- `--flat` optional; markdown-only. Skips the default group-by-project
  rendering in favor of a single `## Entries (chronological)` section.
  Rejected with a user error when combined with `--format json`.
- Accepts the same filter flags as `brag list` (`--tag`, `--project`,
  `--type`, `--since`, `--limit`). `ListFilter` is shared verbatim
  between the two commands.
- JSON output shape locked by
  [DEC-011](../decisions/DEC-011-json-output-shape.md); byte-identical
  to `brag list --format json` on the same rows.
- Markdown output shape locked by
  [DEC-013](../decisions/DEC-013-markdown-export-shape.md): level-1
  document heading, provenance block (`Exported:`, `Entries:`,
  `Filters:`), `## Summary` with `**By type**` / `**By project**`
  counts, then entries grouped under `## <project>` in alphabetical-ASC
  order (`(no project)` last) with within-group chronological-ASC
  ordering. `--flat` swaps grouping for a single
  `## Entries (chronological)` wrapper.
- Unknown or missing `--format` values exit 1 (user error).

### `brag summary --range week|month` (STAGE-004)

```
brag summary --range week                          # last 7 UTC days, markdown
brag summary --range month --format json           # last 30 UTC days, JSON envelope
brag summary --range week --tag auth --project p   # compose filters
```

Rule-based digest of the rolling time window. Output is a markdown
document (default) or single-object JSON envelope (`--format json`)
carrying:

- **Provenance:** `Generated:` (RFC3339), `Scope:` (week|month),
  `Filters:` (echoed flags or `(none)`).
- **Summary block:** counts by type and by project (DESC by count,
  alphabetical-ASC tiebreak; `(no project)` last in the by-project
  list).
- **Highlights:** entry titles + IDs grouped by project,
  chronological-ASC within group; descriptions are intentionally
  elided for the "skim before pasting" goal.

Flags:
- `--range week|month` REQUIRED. `week` = last 7 UTC days from
  `time.Now()`; `month` = last 30 UTC days. Rolling window, NOT
  calendar week/month.
- `--format markdown|json` defaults to `markdown`. JSON is a
  single-object envelope (NOT an array — diverges from DEC-011's
  list shape because aggregations carry metadata). Shape locked by
  [DEC-014](../decisions/DEC-014-rule-based-output-shape.md).
- `--tag <token>`, `--project <name>`, `--type <name>` reuse `brag
  list`'s `ListFilter` semantics. No `--since`/`--limit`/`--out` on
  summary in MVP.
- Output goes to stdout. Redirect with `>` if you want a file.

Unknown or missing `--range` or `--format` values exit 1 (user error).

### `brag review --week | --month` (STAGE-004)

```
brag review                                  # last 7 UTC days, markdown (silent default)
brag review --week                           # explicit; same as bare
brag review --month --format json            # last 30 UTC days, JSON envelope
```

Reflection digest: recent entries grouped by project, followed by
three hard-coded reflection questions designed to be pasted into an
external AI session for guided self-review. No LLM ships in the
binary.

Document structure:

- **Provenance:** `Generated:` (RFC3339), `Scope:` (week|month),
  `Filters: (none)` (review never accepts filter flags; the value is
  constant).
- **Entries:** under `## Entries`, per-project `### <project>` groups
  with bulleted `- <id>: <title>` per entry. Group order alpha-ASC
  with `(no project)` last; within-group entries chrono-ASC. Markdown
  elides descriptions for compactness; the JSON form includes the
  full per-entry shape.
- **Reflection questions:** under `## Reflection questions`, three
  hard-coded numbered questions. The questions ALWAYS render — even
  when zero entries match — because the questions are the point of
  the command.

Flags:
- `--week` and `--month` are mutually exclusive boolean flags. Bare
  `brag review` silently defaults to `--week` (no stderr notice).
  Rolling-window semantics: `--week` = last 7 UTC days; `--month` =
  last 30 UTC days.
- `--format markdown|json` defaults to `markdown`. JSON is the
  single-object envelope locked by
  [DEC-014](../decisions/DEC-014-rule-based-output-shape.md), with
  top-level keys `generated_at`, `scope`, `filters`,
  `entries_grouped`, `reflection_questions`. Each item inside
  `entries_grouped[].entries` is the DEC-011 9-key per-entry shape
  ([DEC-011](../decisions/DEC-011-json-output-shape.md)).
- Filter flags `--tag` / `--project` / `--type` are NOT accepted on
  review — the digest is "the last 7/30 days, period." No `--out`
  flag; output goes to stdout (redirect with `>` for a file).

Unknown `--format` values, or `--week` and `--month` together, exit 1
(user error).

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
- `DEC-011` — shared JSON output shape for `brag list --format json` and `brag export --format json`
- `DEC-013` — markdown export shape for `brag export --format markdown` (+`--flat`)
- `DEC-012` — stdin-JSON schema for `brag add --json` (single object, title required, server-owned fields tolerated-and-ignored)
- `DEC-014` — rule-based output shape for `brag summary`, `brag review`, and `brag stats` (arriving later in STAGE-004): single-object JSON envelope with `generated_at` / `scope` / `filters` provenance + per-spec payload keys; markdown convention reuses DEC-013's provenance + summary-block style.
