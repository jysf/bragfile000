# CLI Contract

`brag` has no HTTP/RPC API. Its external contract is its argv surface
and its stdout/stderr behavior. This doc is the frozen spec of that
contract across PROJ-001 and PROJ-002: what each command takes, what it prints, and
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

**STAGE-007 (cwd `--project` auto-fill):**

When `--project` is not supplied, `brag add` resolves the current working
directory against registered project locations (nearest-ancestor match,
DEC-019) and auto-fills the entry's project from the matching project's
name. This applies to all three input modes:

- flag mode — fires only when `--project`/`-p` is not passed at all;
  passing `-p` (even `-p ""`) is an explicit choice and is recorded
  verbatim.
- JSON mode — fires only when the stdin object has no non-empty
  `"project"` field.
- editor mode — fires only when the buffer's `Project:` header is empty.

Auto-fill is silent (no stderr) and best-effort: if the cwd cannot be
resolved or matches no registered project, the entry is saved with an
empty project, exactly as before. An explicit project always wins over
the cwd. See `brag project here` (SPEC-031) for the shared resolver.

### `brag list` — list entries

```
brag list [-P|--show-project] [--format json|tsv] [--tag T] [--project P] [--type T] [--since 2026-01-01] [--limit N]
```

- `--tag`, `--project`, `--type` filter on exact field value
  (tags filter uses exact tag-name membership via the normalized
  `taggings` join — DEC-015 / SPEC-025).
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

### `brag stats` (STAGE-004)

```
brag stats                        # lifetime corpus, markdown
brag stats --format json          # lifetime corpus, JSON envelope
```

Lifetime panorama: six aggregations over the entire corpus — total
entries, entries/week (rolling average over the corpus span), current
and longest streak (consecutive local days with entries; the current
streak stays alive through yesterday), top-5
most-common tags, top-5 most-common projects, plus the corpus span
(first entry, last entry, days). No LLM ships in the binary.

Document structure:

- **Provenance:** `Generated:` (RFC3339), `Scope: lifetime`
  (hard-coded — stats has no time window),
  `Filters: (none)` (hard-coded — stats accepts no filter flags).
- **Stats body** (markdown only; omitted on empty corpus per
  [DEC-014](../decisions/DEC-014-rule-based-output-shape.md)): under
  `## Stats`, five bold sub-headers (`**Activity**`, `**Streaks**`,
  `**Top tags**`, `**Top projects**`, `**Corpus span**`) with bulleted
  metric content under each.

Flags:

- `--format markdown|json` defaults to `markdown`. JSON is the
  single-object envelope locked by
  [DEC-014](../decisions/DEC-014-rule-based-output-shape.md). Top-level
  keys: `generated_at`, `scope` (always `"lifetime"`), `filters`
  (always `{}`), `total_count`, `entries_per_week` (decimal weeks
  rolling average rounded to 2 decimals; sub-1-week corpus → `0`),
  `current_streak`, `longest_streak`, `top_tags`, `top_projects`,
  `corpus_span`.
- `top_tags` and `top_projects` are arrays of `{tag, count}` /
  `{project, count}` objects in DESC-by-count order with alpha-ASC
  tiebreak. Strict cap at 5 (alpha-ASC determines which 5 when the
  boundary count ties); array shape (rather than map keyed by name)
  preserves the DESC-by-count ordering, which a map encoding would
  lose under `encoding/json`'s alpha-sort.
- `corpus_span` is a sub-object with `first_entry_date`,
  `last_entry_date` (`"YYYY-MM-DD"` UTC, or `null` on empty corpus),
  and `days` (int, inclusive on both endpoints; single-day corpus
  → `1`).
- Lifetime corpus only — no `--range`, no `--week`/`--month`. Use
  `brag summary --range week|month` for windowed digests; `brag
  review --week|--month` for reflection over recent windows.
- Filter flags `--tag` / `--project` / `--type` are NOT accepted on
  stats — the value of stats is the unfiltered lifetime view. No
  `--out` flag; output goes to stdout (redirect with `>` for a file).

Unknown `--format` values exit 1 (user error). Undeclared flags
(`--tag`, `--project`, `--type`, `--out`, `--range`, `--since`,
`--week`, `--month`) surface as cobra `unknown flag` errors.

### `brag tags` — tag taxonomy view (STAGE-006)

```
brag tags                         # name<TAB>count rows, most-used first
brag tags --format json           # naked JSON array of {tag, count}
```

Lists every in-use tag with its total usage count across all entries, one per
line as `<name>\t<count>`, sorted by count (descending) then name (ascending).
Tags with no remaining memberships (orphans) are omitted — only in-use tags appear.
Counts span all taggable object types (only `'entry'` today; `'project'` rows
fold in automatically with STAGE-007 (shipped), with no change here).

Flags:

- `--format json` — emits a naked JSON array of `{"tag": <name>, "count": <n>}`
  objects (DEC-011 naked-array discipline; 2-space indent; `[]` on empty corpus,
  never `null`). The `{tag,count}` element shape matches DEC-014's `top_tags`.
- Default (no `--format`) — plain tab-separated `name\tcount` rows on stdout.

Unknown `--format` values exit 1 (user error). stdout carries data; stderr is
empty on success.

### `brag tag rename <old> <new>` — rename a tag everywhere (STAGE-006)

```
brag tag rename auth authz
```

Renames the tag `<old>` to `<new>` globally: every entry formerly tagged `<old>`
reads `<new>` after the operation. FTS search re-syncs automatically via the
existing `tags_au` trigger (DEC-016). No migration.

- Exits 0 on success; stderr: `Renamed.`
- Exits 1 (user error) if `<new>` already exists — directs the caller to
  `brag tag merge` to combine two tags.
- Exits 1 (user error) if `<old>` does not exist, if `<old> == <new>`, or if
  the wrong number of arguments is given.
- DB is unchanged on any error path (the operation is a single transaction).

### `brag tag merge <src> <dst>` — merge one tag into another (STAGE-006)

```
brag tag merge auth authz
```

Folds `<src>`'s memberships into `<dst>`, de-duplicating: an entry tagged both
`<src>` and `<dst>` ends with exactly one `<dst>` tagging. `<dst>`'s count rises
by the previously-`src`-only memberships. The `<src>` tag row is deleted.
Implemented via DELETE+INSERT on `taggings` (DEC-016 choice 3) so the existing
`taggings_ai`/`taggings_ad` triggers keep `entries_fts` correct. No migration.

- Both `<src>` and `<dst>` must exist (use `brag tag rename` to rename a tag
  that doesn't conflict with any existing name).
- Exits 0 on success; stderr: `Merged.`
- Exits 1 (user error) if either `<src>` or `<dst>` does not exist, if
  `<src> == <dst>`, or if the wrong number of arguments is given.
- DB is unchanged on any error path (the operation is a single transaction).

### `brag project new <name> --path <dir>` — register a project (STAGE-007)

```
brag project new bragfile --path ~/code/bragfile
```

Registers a new project named `<name>` with one initial filesystem location
`<dir>`. The project starts with status `active` and an empty state note
(use `brag project edit` to change them). `--path` is
required and stored verbatim (path normalization is `brag project here`'s
concern, STAGE-007).

- Exits 0 on success; stderr: `Created project "<name>".` (stdout empty).
- Exit 1 (user error) if `<name>` is empty, `--path` is missing/empty, the
  name already exists, or the path is already registered to another project
  (in which case nothing is created — the path is checked first).

### `brag project list` — list projects (STAGE-007)

```
brag project list                 # name<TAB>status<TAB>locations
brag project list --format json   # naked JSON array of project objects
```

Lists every registered project, most-recently-updated first
(`updated_at DESC, id DESC`), one per line as `<name>\t<status>\t<locations>`
(locations comma-joined; `-` when none).

- `--format json` — naked JSON array of project objects (DEC-011; 2-space
  indent; `[]` on empty, never `null`). Object keys: `id, name, status,
  state_note, locations, created_at, updated_at` (locations a JSON array of
  strings; timestamps RFC3339).
- Default (no `--format`) — plain tab-separated rows on stdout.
- Unknown `--format` exits 1 (user error). stdout carries data; stderr empty.

### `brag project show <name|id>` — show one project (STAGE-007)

```
brag project show bragfile
brag project show 3 --format json
```

Shows one project's name, status, state note, and locations. The argument is
resolved as a **name first**; if no project has that name and the argument is
a positive integer, it is resolved as a project **id**. (No recent-brag count
— that is `brag project status` (below).)

- Plain output is a labeled block (`Name:`, `Status:`, `State note:`,
  `Locations:`).
- `--format json` — a single JSON object (not an array) with the same element
  shape as `brag project list`.
- Exit 1 (user error) if no project matches the name or id, or on unknown
  `--format`.

### `brag project status` — active-project dashboard (STAGE-007)

```
brag project status                 # name<TAB>status<TAB>count<TAB>state note
brag project status --format json   # naked JSON array of status objects
```

Shows every **non-archived** project (status `active`, `paused`, or `done`),
most-recently-updated first (`updated_at DESC, id DESC`), as a scannable
dashboard. Each row carries the project name, status, a **brag count** (the
number of entries whose `project` string equals the project name — the DEC-017
soft string match, counted over all time), and the state note.

- Plain output: tab-separated `<name>\t<status>\t<brag_count>\t<state_note>`
  rows on stdout (a long state note is truncated; an empty note prints empty).
- `--format json` — naked JSON array of status objects (DEC-011; 2-space
  indent; `[]` on empty, never `null`). Object keys: `id, name, status,
  state_note, brag_count, created_at, updated_at` (`brag_count` a number;
  `state_note` carried in full, never truncated; timestamps RFC3339).
- Default (no `--format`) — plain rows. Unknown `--format` exits 1 (user
  error). stdout carries data; stderr empty.

### `brag project here` — show the project for the current directory (STAGE-007)

```
brag project here
brag project here --format json
```

Resolves `os.Getwd()` against registered project locations using
nearest-ancestor (longest-prefix) matching (DEC-019): you may be anywhere
inside a registered location's directory tree, not just at the exact root.
When multiple registered paths are ancestors of the cwd, the most specific
(longest) path wins.

- Plain output (default): a single tab-separated line
  `<name>\t<status>\t<state_note>` on stdout (`-` when state note is
  empty); stderr empty; exit 0.
- `--format json` — a single JSON object with the full project shape
  (same as `brag project show --format json`; `locations` hydrated).
- Not inside any registered project → stderr:
  `not inside any registered project`, exit 1, stdout empty.
- Unknown `--format` exits 1 (user error). No positional arguments;
  reads `os.Getwd()` only.

### `brag project edit <name|id>` — edit a project's fields (STAGE-007)

```
brag project edit bragfile --status paused
brag project edit bragfile --state-note "shipped tags; next: cut v0.2.0"
brag project edit bragfile --name brag-cli
brag project edit bragfile --add-path ~/code/bragfile
brag project edit bragfile --remove-path /srv/old-location
```

Edits a project's fields. The argument resolves as a **name first**, then as
a positive-integer **id**. Pass at least one of `--name`, `--status`,
`--state-note`, `--add-path`, or `--remove-path`; unspecified fields are
unchanged.

- `--name` — rename (must be unique). Renaming does **not** rewrite the project
  string on existing brag entries (DEC-017); they keep their captured string.
- `--status` — one of `active`, `paused`, `done`, `archived` (validated).
- `--state-note` — the free-text state/next-action note.
- `--add-path` / `--remove-path` (both repeatable) — attach/detach filesystem
  locations. Paths match **verbatim** against what was registered.
  `--remove-path` exits 1 if the path is not registered to this project or is
  registered to a **different** project; `--add-path` exits 1 if the path is
  already registered. All location changes in one invocation apply
  **atomically** (removes before adds); a failure leaves locations unchanged.
  Location edits do **not** change `updated_at` (DEC-020).
- A scalar edit bumps `updated_at` (so the project rises in `brag project list`
  recency order); a location-only edit does not.
- Exits 0 on success; stderr: `Edited project "<name>".` (stdout empty).
- Exit 1 (user error) if no flag is given, the project is not found, the new
  name is already taken, `--status` is outside the enum, or a location operation
  is rejected.

### `brag project archive <name|id>` — archive a project (STAGE-007)

```
brag project archive bragfile
```

Sets a project's status to `archived` — a **non-destructive, recoverable**
flip. The project, its state note, and its locations are all preserved. Restore
it with `brag project edit <name|id> --status active`. This is **not** delete.

- Exits 0 on success; stderr: `Archived project "<name>".` (stdout empty).
- Exit 1 (user error) if the project is not found.

### `brag project delete <name|id>` — permanently delete a project (STAGE-007)

```
brag project delete bragfile        # prompts y/N on stdin
brag project delete bragfile --yes  # skip the prompt
```

Permanently removes a project and its `project_locations` rows. **Irreversible**
(distinct from `archive`). Prompts for `y/N` confirmation on stderr unless
`--yes`/`-y` is passed; a non-`y` answer prints `Aborted.` and exits 0 without
deleting. Existing brag entries are **not** touched — an entry keeps its
project string (DEC-017), so `brag list --project <name>` still finds those
entries afterward (blast radius on entries: none — DEC-018).

- `--yes`, `-y` — skip the confirmation prompt.
- Exits 0 on success; stderr: `Deleted project "<name>".` (stdout empty).
- Exit 1 (user error) if the project is not found.

### `brag completion <shell>` — generate shell completion script (STAGE-005)

```
brag completion zsh|bash|fish
```

Writes a shell completion script to stdout. Supported shells: `zsh`, `bash`,
`fish`.

To source in the current session:

- **zsh:** `source <(brag completion zsh)`
- **bash:** `source <(brag completion bash)`
- **fish:** `brag completion fish | source`

For permanent setup, add the sourcing line to the shell's startup file
(`~/.zshrc`, `~/.bashrc`, or `~/.config/fish/config.fish`).

- Stdout on success: the completion script (pipe to `source` or redirect to a
  file).
- Stderr on success: empty.
- Exit 0 on success; 1 if an unsupported shell name is given (user error); cobra
  arg-count error if the shell arg is omitted.
- No `--db` / `BRAGFILE_DB` dependency — completion generation is stateless.
- PowerShell is not supported (bragfile distributes for macOS + Linux only).

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
- `DEC-014` — rule-based output shape for `brag summary`, `brag review`, and `brag stats`: single-object JSON envelope with `generated_at` / `scope` / `filters` provenance + per-spec payload keys; markdown convention reuses DEC-013's provenance + summary-block style.
- `DEC-016` — tag mutation semantics: `brag tags` in-use-only taxonomy (count-DESC/name-ASC; `{tag,count}` JSON shape), rename-errors-into-existing, merge via DELETE+INSERT, orphan tags invisible (no GC).
- `DEC-017` — `entries.project` ↔ `projects` relationship (soft string match) + `projects.status` enum + single `state_note`; the data `brag project show`/`list` render.
- `DEC-018` — `brag project delete` blast radius: entries untouched (soft match), project_locations deleted manually in-tx (FK off → no cascade), `'project'` taggings cleaned in-tx; archive is the recoverable status flip, delete is irreversible.
- `DEC-019` — `brag project here` / `brag add` cwd auto-fill: nearest-ancestor (longest-prefix) match against registered project locations
- `DEC-020` — `brag project edit` location editing: `RemoveLocation`/`EditLocations`; remove-not-attached and remove-other-project are user errors; verbatim path matching; one invocation's location changes are atomic (removes before adds); location edits don't bump `updated_at`.
