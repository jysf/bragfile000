# Using `brag` — tutorial

> **Scope:** what you can do with `brag` today. See
> [`docs/api-contract.md`](./api-contract.md) for the full command
> surface.

---

## 1. Check you're wired up

```bash
brag --version    # brag version dev
brag --help       # shows add, list, --db, --version
```

If `brag` isn't found, `~/go/bin` probably isn't on your `$PATH`:

```bash
echo 'export PATH="$HOME/go/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

Or re-install: `just install` (from the repo root).

---

## 2. Capture your first brag

Minimum — just a title:

```bash
brag add --title "shipped STAGE-001 end-to-end in a day"
# or with the shorthand:
brag add -t "shipped STAGE-001 end-to-end in a day"
```

It prints the inserted ID (just a number) on stdout. You can pipe it:

```bash
id=$(brag add -t "untangled the auth flake")
echo "captured as id=$id"
```

---

## 3. Capture with full metadata

All five optional fields:

```bash
brag add \
  --title "cut p99 login latency from 600ms to 120ms" \
  --description "replaced the join-on-every-request with a redis lookup; \
postmortem in doc-042" \
  --tags "auth,perf,backend" \
  --project "platform" \
  --type "shipped" \
  --impact "unblocked mobile v3 release"
```

Or with single-letter shorthands (`-t` title, `-d` description, `-T`
tags, `-p` project, `-k` type, `-i` impact):

```bash
brag add \
  -t "cut p99 login latency from 600ms to 120ms" \
  -d "replaced the join-on-every-request with a redis lookup; \
postmortem in doc-042" \
  -T "auth,perf,backend" \
  -p "platform" \
  -k "shipped" \
  -i "unblocked mobile v3 release"
```

Run `brag add --help` to see them listed alongside the long forms.

Notes on the fields:

- **`--title`** is required. Everything else is optional.
- **`--tags`** is a comma-joined string (e.g. `"auth,perf"`). Tags are
  stored in a normalized taxonomy — see "Tag taxonomy" below for
  `brag tags`, `brag tag rename`, and `brag tag merge`.
- **`--type`** is free-form text — pick whatever feels useful
  (`shipped`, `learned`, `mentored`, `unblocked`, …). No enforced
  enum.
- **`--project`** is whatever you call the initiative (`platform`,
  `onboarding-redesign`, client name, etc.).
- **`--impact`** is a short outcome statement (a metric, a quote, a
  business result). This is what you'll copy-paste into review
  self-assessments later — spend a few extra seconds on it.

### Capture in `$EDITOR` (multi-paragraph descriptions)

When you want to write a multi-paragraph description without fighting
shell quoting, run `brag add` with no flags:

```bash
brag add
# opens $EDITOR on a template that looks like:
#
#   Title:
#   Tags:
#   Project:
#   Type:
#   Impact:
#
#   <body / description goes here>
#
# fill in Title (required) and any other fields, save + quit → prints
# the inserted ID on stdout. Save unchanged (or quit without writing) →
# prints "Aborted." on stderr, exit 0, no row inserted.
```

The editor is resolved as `$EDITOR` → `$VISUAL` → `vi`, same as `brag
edit`. Setting any single entry-field flag (e.g. `-t`, `-d`, …) flips
back to flag mode and bypasses the editor; the persistent `--db` flag
does not (`brag add --db /tmp/work.db` still opens the editor).

### Capture from a script: `--json`

For programmatic capture — a Claude session-end hook, an import
script, piping from another tool — `brag add --json` reads a single
JSON object from stdin:

```bash
echo '{"title":"shipped FTS5 search"}' | brag add --json
# prints the inserted ID on stdout, same as flag mode

brag list --format json | jq '.[0]' | brag add --json
# round-trips an entry (without jq del — server fields are
# tolerated-and-ignored)
```

Required: `title` (non-empty). Optional: `description`, `tags`,
`project`, `type`, `impact` — all free-form text. `tags` stays a
comma-joined string (per
[DEC-004](../decisions/DEC-004-tags-comma-joined-for-mvp.md)); array
form is rejected. Unknown keys are rejected with the offending key
named — catches `"titl"` typos before they become silently-missing
entries. Mutually exclusive with flag-mode field flags. Schema locked
by [DEC-012](../decisions/DEC-012-brag-add-json-stdin-schema.md).

---

## 4. Read them back

```bash
brag list
```

Prints **tab-separated** `<id>\t<created_at>\t<title>`, most recent
first:

```
4	2026-04-20T21:34:12Z	cut p99 login latency from 600ms to 120ms
3	2026-04-20T21:33:05Z	untangled the auth flake
2	2026-04-20T21:32:41Z	shipped STAGE-001 end-to-end in a day
```

Tab-separated output means you can pipe it straightforwardly:

```bash
brag list | head -5                        # 5 most recent
brag list | cut -f3                        # just the titles
brag list | awk -F'\t' '{print $1, $3}'    # id + title, space-separated
```

`list` shows only the one-line summary. Description, tags, project,
type, and impact are stored but not printed — use `brag show <id>`
(below) to dump a single entry in full.

### Filter flags

Narrow the list without piping through `grep`:

```bash
brag list --tag auth                            # entries tagged "auth"
brag list --project platform --since 7d         # last week, one project
brag list --type shipped --limit 5              # 5 most recent shipped
brag list --since 2026-01-01                    # since a specific date
```

- `--tag` matches a single tag as a comma-separated token — `--tag
  auth` matches `"auth"` and `"auth,perf"` but not `"authoring"`.
- `--project` and `--type` are exact, case-sensitive.
- `--since` accepts `YYYY-MM-DD` (midnight UTC) or a duration like
  `7d`, `2w`, `3m` (days / weeks / 30-day months).
- `--limit N` caps the row count.
- Multiple filters combine via AND.

### See project at scan time

Add `-P` (or `--show-project`) to slot a `project` column between
`created_at` and `title`:

```bash
brag list -P
# 4	2026-04-20T21:34:12Z	platform	cut p99 login latency from 600ms to 120ms
# 3	2026-04-20T21:33:05Z	-	untangled the auth flake
# 2	2026-04-20T21:32:41Z	platform	shipped STAGE-001 end-to-end in a day
```

Entries with no project render as `-` so the tab-separated shape
stays consistent. Plain `brag list` (without `-P`) keeps the
three-column output unchanged, so existing `cut -f3` scripts that
pull titles from plain output keep working — under `-P`, titles
shift to field 4 (`cut -f4`) and the project lands at field 3
(`cut -f3`).

Because the project column is just field 3, you can answer common
"which projects am I bragging about" questions with `awk`/`sort`/
`uniq` — no dedicated subcommand needed:

```bash
# Distinct project names across all entries
brag list -P | awk -F'\t' 'NF>2 {print $3}' | sort -u

# Same, but ranked by how many entries each project has
brag list -P | awk -F'\t' 'NF>2 {print $3}' | sort | uniq -c | sort -rn
```

Always pass `awk -F'\t'` — the default whitespace split only lands
on the project column by luck (it relies on `created_at` having no
space). For a format-proof version that survives column reordering,
parse the JSON instead:

```bash
brag list -P --format json \
  | jq -r '.[].project | select(. != "" and . != null)' \
  | sort -u
```

These count the free-text `project` field stored **on entries**,
which is distinct from the project **registry** (`brag project
list`) — see that command for the registered-projects view.

The same `--format json | jq | sort | uniq -c` shape answers the
type and tag questions too. `type` is a plain string; `tags` is a
comma-joined string, so split it before counting:

```bash
# Entries per type, most common first
brag list --format json \
  | jq -r '.[].type | select(. != "" and . != null)' \
  | sort | uniq -c | sort -rn

# Entries per tag (one entry can carry several tags)
brag list --format json \
  | jq -r '.[].tags | select(. != "" and . != null) | split(",")[]' \
  | sort | uniq -c | sort -rn
```

The tag count has a built-in shortcut — `brag tags` lists every tag
with its usage count directly. Reach for the `jq` form when you want
to combine it with a filter (e.g. tag counts within one project) or
reshape the output.

### Machine-readable output: `--format json|tsv`

When you want to pipe entries into `jq`, a spreadsheet, or another
tool, ask for `json` or `tsv`:

```bash
brag list --format json                     # pretty-printed JSON array
brag list --format tsv                      # tab-separated with a header row
brag export --format json --out b.json      # durable dump to a file
brag list --format json | jq '.[0]'         # first entry
brag list --format json | jq '.[] | .title' # just titles
```

Both JSON and TSV output include the same nine fields in the same
order as the `entries` table: `id`, `title`, `description`, `tags`,
`project`, `type`, `impact`, `created_at`, `updated_at`. Tags stay
a comma-joined string (per
[DEC-004](../decisions/DEC-004-tags-comma-joined-for-mvp.md));
timestamps are RFC3339. Empty fields render as `""` in JSON and as
the empty string between tabs in TSV (no dash filler — that is
plain-mode `-P` only).

The JSON shape is locked by
[DEC-011](../decisions/DEC-011-json-output-shape.md) and is
byte-identical between `brag list --format json` and `brag export
--format json` on the same rows, so piping one into `brag add --json`
round-trips an entry without shape transforms (see
[DEC-012](../decisions/DEC-012-brag-add-json-stdin-schema.md) for the
stdin schema).

### Review-ready export: `--format markdown`

For a document you can paste into a quarterly review, a retro writeup,
or a promo packet, export as markdown:

```bash
brag export --format markdown                     # grouped by project
brag export --format markdown --flat              # flat chronological
brag export --format markdown --out report.md     # write to file
brag export --format markdown --project platform  # filter first
brag export --format markdown --since 90d > q.md  # quarter to stdout
```

The default shape groups entries under `## <project>` headings in
alphabetical-ASC order (entries without a project render last under
`## (no project)`), with within-group ordering chronological-ASC so
you can read the period forward through time. An executive summary
at the top breaks down entries `**By type**` and `**By project**`,
and a provenance block records when the export was taken and what
filters applied:

```
# Bragfile Export

Exported: 2026-04-23T12:00:00Z
Entries: 4
Filters: --project platform --since 90d

## Summary

**By type**
- shipped: 3
- learned: 1

**By project**
- platform: 4

## platform

### cut p99 latency
...
```

`--flat` swaps the project grouping for a single
`## Entries (chronological)` wrapper — useful when you want a pure
timeline rather than project-axis buckets. The full shape is locked by
[DEC-013](../decisions/DEC-013-markdown-export-shape.md).

### Publish your brags to a website

To turn a slice of your brags into a post — say, everything you shipped
on one project — start from the same `export`, filtered to what you want
to publish:

```bash
brag export --format markdown --project bragfile --out bragfile.md
brag export --format markdown --since 90d --out last-quarter.md
brag export --format markdown --project bragfile --type shipped --out shipped.md
```

That markdown is valid and pasteable, but the default per-entry field
table reads like a debug dump. For clean blog prose — a heading, a date
line, the impact, and the description as body — reshape the JSON export
with `jq`:

```bash
brag export --format json --project bragfile --out brags.json
jq -r 'sort_by(.created_at) | .[] |
  "## \(.title)\n",
  "*\(.created_at[0:10])*\(if .tags != "" then " · `\(.tags)`" else "" end)\n",
  (if .impact != "" then "**Impact:** \(.impact)\n" else empty end),
  (if .description != "" then .description + "\n" else empty end)
' brags.json > site.md
```

Each entry becomes a `## <title>` your site's markdown renderer styles
natively, with the **impact** — the outcome, the part worth bragging
about — surfaced under it. Tweak the template freely: drop `tags`, add a
`"---\n"` line between entries for separators, or filter to `--type
shipped` for a milestones-only post. Pipe through `pandoc site.md -o
site.html` if your site wants HTML.

**Slicing by release or time window.** Brags carry a *product* name in
`project` (e.g. `bragfile`), not a release or milestone label — so to
post "just what shipped in v1," slice by date. `export` itself only
filters `--since` (on or after), so add a `jq` upper bound for a closed
window:

```bash
# everything before a cutoff date (e.g. one release era)
jq 'map(select(.created_at < "2026-06-01"))' brags.json > v1.json
# a closed window: on/after A and before B
jq 'map(select(.created_at >= "2026-01-01" and .created_at < "2026-04-01"))' \
  brags.json > q1.json
```

Then run the prose template above over the sliced file. Pick the cutoff
from a natural gap in your history (`brag list -P` shows dates) so a
release boundary lands cleanly between entries.

### Search your entries

`brag search "query"` runs a full-text search over every indexed
field (title, description, tags, project, impact):

```bash
brag search "latency"                    # rows mentioning "latency" anywhere
brag search "cut latency"                # rows with BOTH "cut" and "latency"
brag search "auth-refactor"              # literal match; hyphens are fine
brag search "redis" --limit 5            # cap to 5 most relevant
```

Output format matches `brag list` (tab-separated
`<id>\t<created_at>\t<title>`), so the same pipe tricks apply:

```bash
brag search "latency" | cut -f3          # just the titles
brag search "auth" | head -3             # top 3 most relevant
```

Multi-word queries combine with AND — a row with only `cut` or only
`latency` will not match `brag search "cut latency"`. Hyphens,
asterisks, and other FTS5 operators inside the query are treated as
literal characters (see
[DEC-010](../decisions/DEC-010-search-query-syntax.md)), so
`brag search "SPEC-011"` finds the entries you expect. Results are
ordered by relevance.

### Edit an entry

Fix a typo, flesh out a description, or revise metadata after the fact:

```bash
brag edit 42
# opens $EDITOR on a buffer that looks like:
#
#   Title: untangled the auth flake
#   Tags: auth
#
#   notes go here
#
# change any header or the body, save + quit → prints "Updated." on stderr.
# quit without saving (or save unchanged) → prints "No changes." on stderr.
```

`brag` resolves the editor via `$EDITOR` → `$VISUAL` → `vi`, so you can
override it per invocation:

```bash
EDITOR="code --wait" brag edit 42
```

Deleting the `Title:` header (or leaving it empty) fails the save with
an exit-1 user error and leaves the entry untouched.

### Delete an entry

Caught a typo and want to start over? `brag delete <id>`:

```bash
brag delete 42
# prints to stderr:
#   Delete entry 42 ("untnagled the auth flake")? [y/N]
# type y + enter to confirm; anything else aborts cleanly.

brag delete 42 -y    # skip the prompt (scripting / muscle memory)
```

Declining at the prompt exits 0 (no harm done). The delete is a hard
delete — there is no undo, no trash bin. The `.sqlite` file is your
backup.

### Weekly reflection: `brag review`

Run `brag review --week` (or just `brag review`) to see your last 7
days of entries grouped by project, followed by three reflection
questions designed to seed deeper self-review:

1. What pattern do you see in this period?
2. What did you underestimate?
3. What's missing here that should be?

Pipe the JSON form into your favorite LLM for guided reflection:

```bash
brag review --week --format json | claude "use the entries and questions above to reflect on my week"
```

Use `brag review --month` for a 30-day window. Filter flags are not
accepted — the digest is the unfiltered window. (`brag summary` is
the right command if you want filter composition.)

### Lifetime stats: `brag stats`

Run `brag stats` for the lifetime panorama: total entries,
entries-per-week rolling average, current and longest streak, top-5
most-common tags and projects, plus your corpus span:

```bash
brag stats
```

Or pipe the JSON form into your favorite LLM for a "what does my
year look like?" prompt:

```bash
brag stats --format json | claude "summarize my brag history"
```

Stats is corpus-wide — there are no filter or range flags. Use `brag
summary` for windowed digests, or `brag review` for reflection over
the last 7 or 30 days.

### Impact by initiative: `brag impact`

When you're writing a review or a quarterly update, `brag impact`
pulls the entries that carry an `impact` statement, grouped by
initiative (project), over a **calendar** reporting period — the way
your manager and skip-level think about time:

```bash
brag impact --quarter                   # this calendar quarter, by initiative
brag impact --month                     # this calendar month
brag impact --year --format json        # this calendar year, JSON envelope
brag impact --since 2026-01-01          # since a date (YYYY-MM-DD or Nd/Nw/Nm)
```

Exactly one window is required, and the windows are calendar periods
(not the rolling windows `brag summary`/`review` use) — "this quarter"
means the actual calendar quarter, up to today. Only entries with a
non-empty impact appear in the body; a `<shown>/<in-window> with
impact` tally keeps you honest about what was left out. Filter flags
`--tag`/`--project`/`--type` compose with the window. Pipe the JSON
form into an LLM to draft the narrative:

```bash
brag impact --quarter --format json | claude "draft my quarterly impact summary"
```

### Tell your story: `brag story`

Where `brag impact` gives you the grouped data, `brag story` shapes it
into a **narrative arc for a specific audience**. It coalesces your brags
into **threads** (initiatives, time-ordered, with impact beats marked
`★`), assembles a **throughline skeleton**, and appends a per-audience
**framing directive** — a complete artifact you can paste into an LLM, or
read as-is. No model ships in the binary; the LLM (already in your
session) writes the prose:

```bash
brag story --audience me                          # candid reflection, this year
brag story --audience manager --month             # tactical 1:1 update, this month
brag story --audience exec --quarter              # impact-forward, this quarter
brag story --audience exec --year --format json   # arc-aware JSON envelope
brag story --audience me --theme perf             # add a cross-project perf arc
brag story --audience exec --print-directive      # just the framing directive
```

`--audience` is required. The same corpus tells a **different story** per
audience, rule-driven not just toned. There are four built-ins along a
gradient: `me` keeps every thread and the messy middle (impact-less beats
included, low altitude); `manager` also keeps everything but on a tighter
monthly cadence with a tactical (shipped / blockers / next) voice; `skip`
surfaces only impact-bearing initiatives yet keeps their supporting beats,
grouped by initiative for the "so what"; `exec` surfaces only impact-bearing
threads, drops impact-less beats, and leads with the highest-impact arc.
Each audience carries a default window (`me` → year, `manager` → month,
`skip` → quarter, `exec` → quarter) that an explicit window flag overrides.
Audiences are extensible profiles, not a fixed list — drop a `<name>.yaml`
in your story-profiles directory to add one. Pipe it into an LLM to finish:

```bash
brag story --audience exec --quarter | claude "weave these threads into one headline arc"
```

### Tag taxonomy

See every tag you've used, with usage counts:

```bash
brag tags                         # name<TAB>count rows, most-used first
brag tags --format json           # naked JSON array of {tag, count}
```

Only tags with at least one remaining entry appear. Sort order is count
(descending) then name (ascending).

Rename a tag everywhere at once — all entries formerly tagged `auth` will
read `authz`, and FTS search re-syncs automatically:

```bash
brag tag rename auth authz
```

- If `authz` already exists, the command errors and directs you to `merge`.

Fold one tag into another, de-duplicating. An entry tagged both `auth` and
`perf` ends up with exactly one `perf` tagging:

```bash
brag tag merge auth perf
```

- Both tags must already exist (use `rename` if the destination tag is new).
- The `auth` tag row is deleted; `perf`'s count rises by the previously
  `auth`-only memberships.

### Projects

A **project** is a first-class, named workspace you register once and
then attach to brags automatically. Registering a project's directory
lets `brag add` fill in `--project` for you whenever you work inside it.

Register a project and point it at a directory:

```bash
brag project new bragfile --path ~/code/bragfile
# stderr: Created project "bragfile".
```

`--path` is required and stored verbatim; a path already registered to
another project is rejected.

Ask which project the current directory belongs to:

```bash
cd ~/code/bragfile/internal/storage
brag project here
# bragfile	active	-
# name<TAB>status<TAB>state-note. Nearest-ancestor match: any
# subdirectory of a registered path resolves, not just the exact root.
```

Outside any registered project, `brag project here` prints `not inside
any registered project` to stderr and exits 1.

#### Auto-fill `--project` from your working directory

Once a directory is registered, `brag add` fills in `--project` for you
whenever you don't pass one — in flag, editor, and `--json` modes alike:

```bash
cd ~/code/bragfile
brag add -t "shipped the projects walkthrough"
# the entry's project is auto-set to "bragfile" — no -p needed

brag add -t "a cross-cutting note" -p platform
# an explicit -p always wins; auto-fill never overrides it
```

Auto-fill is silent and best-effort: outside any registered project, or
if the directory can't be resolved, the entry is just saved with no
project, exactly as before.

#### Review your projects

```bash
brag project status
# name<TAB>status<TAB>brag-count<TAB>state-note, most-recent first.
# Archived projects are hidden. brag-count is how many entries carry
# that project name (the DEC-017 soft string match).

brag project list
# name<TAB>status<TAB>locations (comma-joined; "-" when none)

brag project show bragfile          # labeled block: Name / Status / State note / Locations
brag project show bragfile --format json
```

`status`, `list`, `show`, and `here` all accept `--format json` for
scripting.

#### Edit a project

```bash
brag project edit bragfile --status paused
brag project edit bragfile --state-note "shipped tags; next: cut v0.2.0"
brag project edit bragfile --name brag-cli
brag project edit bragfile --add-path ~/code/bragfile-fork
brag project edit bragfile --remove-path /srv/old-location
```

- `--status` is one of `active`, `paused`, `done`, `archived`.
- `--add-path` / `--remove-path` are repeatable and apply atomically
  (removes before adds); paths match verbatim against what was
  registered.
- Renaming a project does **not** rewrite the project string on existing
  brags — they keep what they were captured with.

#### Archive vs. delete

```bash
brag project archive bragfile
# status → "archived"; recoverable. Restore it with:
brag project edit bragfile --status active

brag project delete bragfile        # prompts y/N on stdin
brag project delete bragfile --yes  # skip the prompt
```

`archive` is a reversible status flip that hides the project from `brag
project status` but preserves everything. `delete` is **irreversible** —
it removes the project and its registered locations. **Neither touches
your brag entries:** an entry keeps its project string, so `brag list
--project bragfile` still finds those entries afterward.

---

## 5. Where the data lives

```bash
ls -la ~/.bragfile/db.sqlite
```

That's the default, and **every `brag` invocation from any directory
uses it** — the path is absolute and home-expanded, so it doesn't matter
whether you're in the bragfile repo or elsewhere.

### Back up your brags

The database is a single SQLite file, so a backup is a copy of that file
— but take the copy with SQLite's own backup command, not a bare `cp`,
so the snapshot is always transaction-consistent:

```bash
# preferred — a consistent single-file snapshot via the sqlite3 CLI:
sqlite3 ~/.bragfile/db.sqlite ".backup '$HOME/brag-backup.db'"

# equivalent, also consistent:
sqlite3 ~/.bragfile/db.sqlite "VACUUM INTO '$HOME/brag-backup.db'"
```

The result is a portable `.db` you can copy to another machine, commit
to a private repo, or stash anywhere:

```bash
cp ~/brag-backup.db ~/some-private-repo/   # the snapshot is safe to plain-copy
```

> Why not `cp ~/.bragfile/db.sqlite` directly? `brag` doesn't enable WAL
> mode, so a bare `cp` of an idle database is *currently* safe — but
> `.backup` / `VACUUM INTO` stay correct even if that changes or a write
> is in flight. Prefer them. (You can also `brag export --format json
> --out backup.json` for a tool-portable dump — see §4.)

### Automatic backup before an upgrade migrates your DB

When you move to a newer `brag` — e.g. `brew upgrade bragfile` — there is
**no manual migration step**: the first command you run opens your
existing database, applies any new schema migrations automatically, and
your entries carry forward (a v0.1.x database upgrades to the v0.2.0
schema in place, losslessly). Before it applies anything, it **snapshots
the database first** — automatically, before touching it. The snapshot
lands next to your DB as a timestamped sidecar:

```
~/.bragfile/db.sqlite.pre-0004_add_projects.20260612T093015Z.backup
```

- It fires **only** when an existing database has pending migrations to
  apply — a brand-new DB and an already-up-to-date DB are never copied.
- The copy is a consistent `VACUUM INTO` snapshot of the
  **pre-migration** state; open it with `sqlite3` to inspect or recover.
- If the snapshot can't be written, `brag` **aborts** rather than migrate
  an un-backed-up database (exit 2) — nothing is changed.
- It's silent and non-interactive, so it never breaks `brag add --json`
  or other scripted pipelines.
- Snapshots are **kept, not pruned** — delete old `*.backup` sidecars
  yourself when you no longer need them.

Peek at raw data:

```bash
sqlite3 ~/.bragfile/db.sqlite "select * from entries order by id desc limit 3"
```

---

## 6. Override the DB (rare, but useful)

For scratch testing or keeping separate DBs:

```bash
brag add --db /tmp/work.db --title "only for work stuff"
brag list --db /tmp/work.db

# or semi-permanently in your shell:
export BRAGFILE_DB=/tmp/work.db
brag list   # uses /tmp/work.db
unset BRAGFILE_DB
```

Precedence: `--db` flag > `BRAGFILE_DB` env > default
`~/.bragfile/db.sqlite`. See
[DEC-003](../decisions/DEC-003-config-resolution-order.md).

---

## 7. Exit codes (useful for scripts)

- `0` — success
- `1` — user error (missing/empty `--title`, bad arg)
- `2` — internal error (DB path unusable, disk error, etc.)

```bash
brag add --title "" && echo ok || echo "exited $?"   # prints "exited 1"
```

---

## 8. A daily habit that actually works

The tool is optimized for 10-second capture. Build the muscle:

```bash
# at the end of each work day, 30 seconds:
brag add --title "today's best thing" \
         --project "work" \
         --impact "why it mattered"
```

For 10-second flag-mode capture, a tiny wrapper saves typing:

```bash
# add to ~/.zshrc
bragit() { brag add --title "$*" --project "work"; }
```

Then:

```bash
bragit untangled a gnarly CORS bug
```

For longer narrative entries, `brag add` (no args) opens `$EDITOR` —
see §3 above.

---

## 9. Power-user escape hatch

Everything in this tutorial is shipped in v0.2.0. For corner cases
`brag list` doesn't surface, `sqlite3 ~/.bragfile/db.sqlite` is your
escape hatch.

---

## 10. Shell completions

`brag completion` generates tab-completion scripts for zsh, bash, and fish.

**zsh** — add to `~/.zshrc`:

```zsh
source <(brag completion zsh)
```

**bash** — add to `~/.bashrc`:

```bash
source <(brag completion bash)
```

**fish** — add to `~/.config/fish/config.fish`:

```fish
brag completion fish | source
```

After sourcing, `brag <tab>` and `brag add --<tab>` show available commands
and flags. Run `brag completion --help` for details.

---

## 11. Agent-native: MCP server, plugin, and provenance

As of v0.3.0, `brag` is usable by an AI agent through native tool calls,
not just by you at the shell.

### `brag mcp serve` — a local MCP server

```bash
brag mcp serve
```

This runs a local [Model Context Protocol](https://modelcontextprotocol.io)
server over stdio (no network) that exposes four tools —
`brag_add`, `brag_list`, `brag_search`, `brag_stats` — as thin wrappers over
your existing database. An MCP-client agent can capture and recall brags
without spawning a shell. The protocol stream owns stdout; nothing
human-facing is ever written there.

### The Claude Code plugin

The MCP server, a `/brag` slash-command, and a quiet session-end
capture-nudge hook are bundled as an installable Claude Code plugin:

```bash
claude plugin marketplace add jysf/bragfile000
claude plugin install brag@bragfile
```

The plugin runs the `brag` binary from your `PATH`, so install it first
(`brew install jysf/bragfile/bragfile`). The capture-nudge hook only
suggests a brag after a session that plausibly shipped something, never
posts on its own, and can be silenced with `BRAG_CAPTURE_NUDGE=off`. See
[`BRAG.md`](../BRAG.md) and [`plugin/README.md`](../plugin/README.md) for
the full walkthrough.

### Provenance: who wrote which brag

When an agent captures a brag through the MCP `brag_add` tool, the entry is
stamped with reserved `agent:<name>` and `model:<id>` tags (for example
`agent:claude-code`, `model:claude-opus-4-8`). These are a convention over
the normal tag system — no schema change — so you can tell agent-authored
entries from your own:

```bash
brag list --author agent                    # entries an agent captured
brag list --author human                    # entries you captured
brag list --author agent --format json | jq length
```

Your own `brag add` never stamps provenance — only the MCP path does.

---

## Further reading

- [`docs/api-contract.md`](./api-contract.md) — full CLI surface
  across all stages (what you'll get).
- [`docs/architecture.md`](./architecture.md) — how the pieces fit.
- [`docs/data-model.md`](./data-model.md) — schema today and planned.
- [`AGENTS.md`](../AGENTS.md) — conventions and daily commands for
  anyone working on `brag` itself.
