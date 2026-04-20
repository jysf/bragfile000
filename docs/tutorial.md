# Using `brag` — tutorial

> **Scope:** what you can do with `brag` today (through the end of
> STAGE-001). `show`, `edit`, `delete`, `search`, `export`, and
> `summary` all arrive in later stages. See
> [`projects/PROJ-001-mvp/brief.md`](../projects/PROJ-001-mvp/brief.md)
> for the full plan.

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
- **`--tags`** is a comma-joined string today (e.g. `"auth,perf"`). No
  normalization, no tag registry. Future stages may add tag search
  and rename.
- **`--type`** is free-form text — pick whatever feels useful
  (`shipped`, `learned`, `mentored`, `unblocked`, …). No enforced
  enum.
- **`--project`** is whatever you call the initiative (`platform`,
  `onboarding-redesign`, client name, etc.).
- **`--impact`** is a short outcome statement (a metric, a quote, a
  business result). This is what you'll copy-paste into review
  self-assessments later — spend a few extra seconds on it.

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
type, and impact are stored but not printed. `brag show <id>` arrives
in STAGE-002 to dump a single entry in full.

---

## 5. Where the data lives

```bash
ls -la ~/.bragfile/db.sqlite
```

That's the default, and **every `brag` invocation from any directory
uses it** — the path is absolute and home-expanded, so it doesn't
matter whether you're in the bragfile repo or elsewhere.

- **Back up** by copying the file.
- **Move to a new machine** by copying the file.
- **Version-control your brags?** Just `cp ~/.bragfile/db.sqlite
  ~/some-private-repo/` and commit.

Peek at raw data (useful until `show` exists):

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

Until editor-launch ships in STAGE-002, a tiny wrapper saves typing:

```bash
# add to ~/.zshrc
bragit() { brag add --title "$*" --project "work"; }
```

Then:

```bash
bragit untangled a gnarly CORS bug
```

---

## 9. What's NOT there yet

So you don't ask the tool for things it can't do:

| Want | Status |
|---|---|
| `brag add` with no args (editor-launch on `$EDITOR`) | STAGE-002 |
| `brag show <id>` / `brag edit <id>` / `brag delete <id>` | STAGE-002 |
| `brag list --tag auth` (filter flags) | STAGE-002 |
| `brag search "query"` (FTS5 full-text search) | STAGE-002 |
| `brag export --format markdown` | STAGE-003 |
| `brag export --format sqlite` | STAGE-003 |
| `brag summary --range week\|month` | STAGE-003 |
| `brew install bragfile` | STAGE-004 |

For now, `sqlite3 ~/.bragfile/db.sqlite` is your escape hatch for
anything `list` doesn't surface.

---

## Further reading

- [`docs/api-contract.md`](./api-contract.md) — full CLI surface
  across all stages (what you'll get).
- [`docs/architecture.md`](./architecture.md) — how the pieces fit.
- [`docs/data-model.md`](./data-model.md) — schema today and planned.
- [`AGENTS.md`](../AGENTS.md) — conventions and daily commands for
  anyone working on `brag` itself.
