---
project:
  id: PROJ-001
  status: active
  priority: high
  target_ship: 2026-05-03

repo:
  id: bragfile

created_at: 2026-04-19
shipped_at: null
---

# PROJ-001: MVP — capture, retrieve, export, ship

## What This Project Is

Deliver a usable, distributable MVP of Bragfile: a local-first Go CLI
(`brag`) backed by an embedded SQLite database that lets an engineer
capture a brag-worthy moment in under ten seconds and retrieve, filter,
and export those moments later. The wave of work ends when the author
has been using it daily at work for two weeks and it is installable via
`brew install bragfile`.

## Why Now

Brag-worthy moments evaporate between review cycles. The author
currently reconstructs them from commit logs under time pressure
immediately before self-reviews, and loses most of them. We want the
scaffolding (tool + habit) in place before the next review cycle so the
data exists when it is needed, rather than racing to rebuild it after
the fact. Scope is deliberately narrow — capture, retrieve, export,
distribute — so we can ship in roughly two weeks and start accumulating
real entries.

## Success Criteria

- A new entry can be captured in under 10 seconds from an open terminal
  (either `brag add "..."` one-shot or `brag add` → `$EDITOR` template).
- Past entries can be listed, filtered (tag / project / type / since),
  searched (full-text), shown, edited, and deleted from the CLI.
- `brag export` produces either (a) a readable Markdown report or
  (b) a portable copy of the SQLite file suitable for backup or another
  machine.
- `brew install bragfile` installs a working binary on macOS (at minimum)
  via a public homebrew tap.
- Author has logged ≥1 entry per working day for two consecutive weeks
  using the shipped binary.

## Scope

### In scope

- Core CRUD: `add`, `list`, `show`, `edit`, `delete` (STAGE-001, STAGE-002).
- Full-text search via SQLite FTS5 (STAGE-002).
- Editor-launch capture: `brag add` with no args opens `$EDITOR` against
  a templated markdown buffer; fields parsed on save (STAGE-002).
- Export: Markdown report and raw SQLite file copy (STAGE-003).
  JSON export added to STAGE-003 scope (2026-04-21) — useful for
  AI/programmatic consumers.
- Rule-based `summary --range=week|month` (group by tag/project, counts,
  rendered as markdown). No AI (STAGE-003).
- UX polish: emoji decoration on stderr feedback messages and the
  `brag show` / `brag list --pretty` output, with `NO_COLOR` +
  TTY-detection escape hatch following industry conventions
  (STAGE-003, added 2026-04-21).
- Distribution: goreleaser build, GitHub release, homebrew tap
  (STAGE-004).
- Data model designed so a future `brag ai-summary` command can read rows
  and POST to an LLM without schema changes. No AI code ships in this
  project.

### Explicitly out of scope

- LLM-backed summaries, narrative generation, resume-bullet rewriting
  — deferred to a future project (PROJ-00N — AI assist).
- Bubble Tea / interactive TUI — deferred. v0.1 is plain cobra CLI.
- Multi-device sync, cloud backup, encrypted-at-rest storage — deferred.
- Rich export targets (JSON, HTML, PDF, resume-bullet format). Markdown
  + sqlite-file only.
- Non-macOS homebrew (Linuxbrew), apt/yum packages, Windows support.
  macOS-first; Linux is a nice-to-have that falls out of goreleaser if
  it's free, but is not a success criterion.
- Auth, accounts, sharing. This is a single-user local-first tool.

## Stage Plan

- [x] STAGE-001 (shipped on 2026-04-20) — Foundations: repo skeleton,
      Cobra scaffold, SQLite schema + migrations, `add` + `list`.
      4 specs shipped (SPEC-001/002/003/004), 1 new DEC (DEC-007),
      4 lessons landed in AGENTS.md §9/§10.
- [x] STAGE-002 (shipped on 2026-04-22) — Capture & retrieval: `add`
      shorthand flags, `list` filter flags, `show`/`edit`/`delete`,
      editor-launch for `add` and `edit`, FTS5 `search`. All 8 specs
      shipped (SPEC-005–012). 4 new DECs (DEC-007/008/009/010),
      5 new AGENTS.md §9/§12 lessons including the full premise-
      audit family. Target was 2026-05-04; shipped 12 days ahead.
- [ ] STAGE-003 (not yet framed) — **Reports + AI-friendly I/O +
      project visibility (ship-blocking scope only).** 3 must-haves:
      project visible in `brag list` output, export trio
      (markdown/sqlite/JSON), machine-readable I/O
      (`brag add --json` + `brag list --format`). No nice-to-have
      bucket — polish items moved to STAGE-004 on 2026-04-22
      reshuffle to keep STAGE-003 tight. `BRAG.md` shipped as a
      chore before framing (done 2026-04-22).
- [ ] STAGE-004 (not yet framed, **provisional**) — **UX polish
      pass.** 9 items grouped into: retrospection & summary
      (`brag summary`), emoji decoration (4 passes — stderr
      feedback, show/list type icons, NO_COLOR + TTY detection),
      AI-agent integrations (Claude session-end hook), habit &
      stats (`brag remind`, `brag stats`), retrospection command
      (`brag review --week`). **Escape hatch preserved**: if
      STAGE-003 ships with enough utility, STAGE-004 may dissolve
      entirely — unreleased items drop to backlog and the project
      jumps directly to STAGE-005 distribution. Three ex-STAGE-004
      items were dropped to backlog on 2026-04-22 reshuffle
      (`--exclude-tag` redaction, git-context auto-capture,
      `--link/--refs`).
- [ ] STAGE-005 (not yet framed) — **Distribution.** goreleaser
      cross-compile, GitHub Actions CI + release workflow,
      homebrew tap at `github.com/jysf/homebrew-bragfile`,
      CHANGELOG + release notes discipline, README polish +
      GIF/demo video.

**Count:** 2 shipped / 0 active / 3 pending

## Dependencies

### Depends on

- None. This is the first project in this repo.
- External-ish: a public GitHub repo at
  `github.com/jysf/bragfile000` (exists) and a second repo for the
  homebrew tap (`github.com/jysf/homebrew-bragfile`, will be created in
  STAGE-004).

### Enables

- Future project: AI-assisted summaries and narrative generation (reads
  rows written by this project).
- Future project: TUI / Bubble Tea polish layer over the same data.
- Future project: sync or export-to-cloud if that ever becomes desired.

## STAGE-003 pre-framing notes

*Seed notes for the STAGE-003 framing session (Prompt 1c). This
section gets consumed into the stage file's Design Notes + Spec
Backlog when STAGE-003 is framed, and trimmed from the brief at
that point. Keep content here accurate; it's the handoff.*

### Post-triage scope (2026-04-22) — authoritative priority

Triaged from a full-ideas-dump via external Claude evaluation.
User filters applied: personal utility first; "nice Go app"
craftsmanship second; adoption pressure explicitly not a driver.
Non-STAGE-003 items land in `backlog.md` or are sketched below
under "STAGE-004 sketch". Detail on individual ideas follows
below this block — the scope list is the authoritative slice.

**STAGE-003 ship-blocking scope (3 items + 1 trivial pre-spec):**

Reshuffled 2026-04-22 (second pass): the "nice-to-have" bucket
moved entirely into STAGE-004 to keep STAGE-003 tight. STAGE-003
is now purely core functionality — must-haves only — with emoji,
summary, and Claude-hook work grouped into a proper polish stage.

0. **`BRAG.md` at repo root.** Trivial (S). **Shipped as a chore
   on 2026-04-22** (commit before STAGE-003 framing); counted as
   pre-spec work rather than a STAGE-003 spec. Delivers the
   agent-integration guide at a canonical location. Content lives
   in `/BRAG.md`; former `docs/agent-brag-guide.md` deleted.

1. **Project visible in `brag list` output.** User flagged
   elevated to must-have ("I want to see quickly what projects
   I have been working on"). Implementation is a design call —
   see "Project column" detail below for the four options
   (default 4-col / `--format` flag / `--pretty` bundle /
   `--columns` list). Framer picks.
2. **Report export trio:** `brag export --format markdown`,
   `--format sqlite`, `--format json`. The original STAGE-003
   backbone, user-flagged pushed up in priority. JSON shape
   detailed below. Can be one M spec or split as 3 × S specs —
   framer decides.
3. **Machine-readable input/output.** `brag add --json` (stdin
   JSON → `Store.Add`) plus `brag list --format json|tsv`.
   Pairs with the export trio to close the I/O loop. Unblocks
   every downstream AI integration. Can be one M spec or split
   as 2 × S specs.

No nice-to-have bucket for STAGE-003. If the framing session
wants additional scope, take it from STAGE-004's list rather than
adding anew.

### STAGE-004 sketch — polish pass (9 items; provisional)

Reshuffled 2026-04-22 (second pass): the ex-STAGE-003 nice-to-
haves (summary, emoji 4 passes, Claude session-end hook) moved
here because they're polish work, not core functionality. Stage
grows from 3 items to 9. Still provisional — if STAGE-003 ships
and real usage shows STAGE-004 items aren't missed, the stage
can dissolve (items drop to `backlog.md`) and the project jumps
to STAGE-005 distribution.

**Retrospection & summary (1 M):**

- **`brag summary --range week|month`** — rule-based aggregation
  grouped by project/type, rendered as markdown. Was original
  STAGE-003 backbone; demoted to polish on 2026-04-22 because
  `brag export --format markdown` (STAGE-003 must-have) already
  gives a decent quarterly doc — summary becomes the
  lighter-weight sibling.

**Emoji decoration (4 passes):**

Four-pass plan carried forward from the earlier pre-framing
notes; detail section below remains the canonical reference.
Framer can pick how many passes to include (Pass 1 is trivial
and should almost certainly ship; Passes 2–4 are opt-in polish).

- Pass 1 — stderr feedback prefixes (`✏️`/`🗑`/`—`/`✗`) on
  edit/delete/confirmation messages (S, trivial).
- Pass 2 — type-based emoji in `brag show` markdown table (S).
- Pass 3 — type-based emoji in `brag list --pretty` (S, bundles
  naturally with STAGE-003 #1 if that picks the `--pretty` path).
- Pass 4 — `NO_COLOR` env var + TTY auto-detection (S).

**AI-agent integrations (1 S):**

- **Claude Code session-end hook example.** Ships
  `scripts/claude-code-post-session.sh` + a `/brag` slash
  command shape. Pure shell + prompt; requires STAGE-003's
  `brag add --json` landed. User explicitly tagged as low
  priority in the 2026-04-22 triage; include only if trivial
  to bolt on after the JSON input spec.

**Habit & stats (2 S):**

- **`brag remind`** — nudge command that prints a message if no
  entries in last N days. Habit-reinforcement; tiny command.
- **`brag stats`** — entries/week, longest streak, most common
  tags/projects. Aggregations over existing schema.

**Retrospection command (1 S):**

- **`brag review --week`** — print entries grouped by project +
  three static reflection questions ("what pattern do you see?",
  "what did you underestimate?", "what's one thing missing?").
  Human-in-the-loop reflection; distinct from `summary`
  aggregation.

**Explicitly NOT in STAGE-004 — deferred to `backlog.md` on
2026-04-22 reshuffle:**

- `brag export --exclude-tag <tag>` redaction filter
- Git-context auto-capture on `brag add`
- `--link` / `--refs` multi-valued field

### STAGE-005 sketch — Distribution

Persistent since project start; no triage needed. For reference:
goreleaser config (cross-compile darwin+linux arm64+x86_64),
GitHub Actions CI (test + lint on PR) and release (tag →
binaries → release notes), homebrew tap at
`github.com/jysf/homebrew-bragfile`, CHANGELOG discipline, README
one-liner polish + GIF/demo of the Claude-session → brag-entry
flow (moved from backlog during 2026-04-22 triage).

---

### Detail on individual ideas (reference material — framer reads as needed)


### Emoji decoration (accumulated 2026-04-21)

**Four-pass plan, each pass a separate candidate spec.** The
framer can decide how many to include in STAGE-003 vs. defer.

**Pass 1 — stderr feedback (S, low risk, recommended for STAGE-003).**
Prefix emoji on the four existing stderr feedback lines produced
by `brag edit` and `brag delete`:
- `Updated.` → `✏️ Updated.`
- `Deleted.` → `🗑 Deleted.`
- `No changes.` → `— No changes.`
- `Aborted.` → `✗ Aborted.`
- `Delete entry N ("title")? [y/N]` → `⚠️ Delete entry N (...)? [y/N]`

Zero risk to stdout scripting (stderr-for-humans constraint
already protects this). Existing tests assert on substrings
(`"Updated."`, `"Deleted."`) which remain intact — the emoji is
prefix-only. ~5 line code changes. No TTY detection needed for
Pass 1; stderr is always human-facing.

**Pass 2 — type-based emoji in `brag show` (S).**
`brag show <id>` produces markdown; emoji in markdown renders
fine everywhere. Decorate the `type` row of the metadata table:
- `shipped` → 🚀
- `fixed` → 🔧
- `learned` → 🎓
- `documented` → 📝
- `mentored` → 🤝
- `unblocked` → 🔓
- `reviewed` → 👀
- (unknown types) → `•` or no icon

Example:
```
| type        | 🚀 shipped                    |
```
Still valid markdown, still pipeable, still parseable. Copy-paste
into review docs preserves the emoji.

**Pass 3 — `brag list --pretty` (S or S+).**
A new `--pretty` flag that opts into emoji + formatting. Default
`brag list` stays plain for scripting. `--pretty` prefixes each
row with the type's emoji (Pass 2's mapping). Tab-separated
columns stay stable.

Example:
```
🚀  42  2026-04-21  shipped the SPEC-011 FTS5 layer
🔧  41  2026-04-20  fixed the stale .gitignore rule
🎓  40  2026-04-19  learned SQLite's external-content FTS5 pattern
```

**Pass 4 — TTY auto-detection + `NO_COLOR` (S).**
Industry conventions for colored / decorated CLI output:
- If stdout is a terminal (via
  `golang.org/x/term.IsTerminal(int(os.Stdout.Fd()))`), pretty mode
  on by default for `list` and `show`.
- If stdout is piped (redirected to file, `grep`, `cut`), plain
  mode automatically — prevents polluting pipes with emoji that
  downstream tools can't parse.
- Respect `NO_COLOR` env var (see https://no-color.org) — any
  non-empty value forces plain mode. Industry standard; also what
  accessibility-tool users set.
- Optional `BRAG_PLAIN=1` env var as brag-specific override.

Implementation: ~10 lines in a small helper; used by `list`,
`show`, and any future decorated output.

### Emoji caveats worth naming in the spec(s)

- **Cross-platform rendering.** Modern terminals (GNOME Terminal,
  Konsole, Alacritty, Kitty, WezTerm, iTerm2, Windows Terminal,
  st) + any modern Linux/macOS distro with Noto Color Emoji
  handle these fine. Bare Linux TTY, legacy cmd.exe on Windows 10,
  and some enterprise-locked-down terminals don't. `NO_COLOR`
  + `--plain` is the escape hatch for those 2–5% of users.
- **Column width.** Emoji are typically 2 cells wide but some
  terminals compute this as 1, causing subtle misalignment in
  markdown tables. Cosmetic only.
- **Variation selectors + ZWJ sequences.** Avoid complex emoji
  like 👨‍💻 (combined codepoints); prefer single-codepoint
  emoji (🚀, 🔧, 📝) which render consistently.
- **Screen readers.** `"🚀 shipped"` reads as "rocket shipped."
  Slightly noisy; screen-reader users typically set `NO_COLOR=1`
  in their shell rc already.

### JSON export (accumulated 2026-04-21)

**Sibling of markdown / sqlite exports. Primary motivation: AI/
programmatic consumers.** Example flow: a user's downstream AI
agent reads `brag export --format json --since 90d` to produce a
quarterly summary or resume bullet. Also useful for piping into
`jq`, backup tooling, or cross-tool integration.

Shape: an array of entry objects, one per entry. Field names
match the SQL column names (`id`, `title`, `description`, `tags`,
`project`, `type`, `impact`, `created_at`, `updated_at`). RFC3339
timestamps as strings (matches storage layer). `tags` stays as a
comma-joined string (matches DEC-004 — don't normalize to array
here unless the design session has a reason).

Example output:
```json
[
  {
    "id": 11,
    "title": "Shipped FTS5 full-text search index layer...",
    "description": "Shipped SPEC-011 — new 0002_add_fts.sql migration...",
    "tags": "sqlite,fts5,migrations,search",
    "project": "bragfile",
    "type": "shipped",
    "impact": "Prepared the data layer for brag search...",
    "created_at": "2026-04-22T06:30:00Z",
    "updated_at": "2026-04-22T06:30:00Z"
  },
  ...
]
```

Same filter flags as `brag list` (so
`brag export --format json --project bragfile --since 7d` works).
Output to stdout by default; `--out file.json` to write to a file.

Use stdlib `encoding/json` — no new dep needed. Pretty-printed by
default (indent=2) for human readability; `--compact` flag to
disable indentation for pipe consumers if anyone asks.

### `brag list` display options — show project column (accumulated 2026-04-22)

**Motivation.** Today `brag list` prints 3 columns tab-separated:
`<id>\t<created_at>\t<title>`. The `project` field is populated
per-entry but invisible at scan time — users can't see "what
projects have I been working on lately?" without `brag show <id>`
or filtering by a specific `--project`. Adding project to the
list output makes daily scanning more useful.

**Design options for the framer to pick between:**

1. **Part of `--pretty` mode** (leans composable with emoji pass 3).
   `brag list` stays 3-column plain for backwards-compat with
   scripts. `brag list --pretty` adds emoji + project column
   (e.g., `🚀  12  2026-04-22  [bragfile]  shipped FTS5 search`).
   One new flag, one coherent pretty-mode feature bundle.
2. **Dedicated `--show-project` / `-P` flag.** Opts in just the
   project column without emoji. Clean separation from the emoji
   concern; combines freely with `--pretty`.
3. **`--columns <list>` flag.** User picks which columns to show,
   e.g., `--columns id,created_at,project,title`. Most flexible;
   most complex. Probably overkill for MVP.
4. **Change default** to always include project. Breaks existing
   scripts that parse 3 tab-separated columns. Rejected — breaking
   stable contract for a UX nicety isn't worth it.

**Author's lean: option 1 (bundle into `--pretty`)** — simplest
composition with the emoji work, keeps plain mode byte-stable.
Framer can decide otherwise if dogfooding reveals that project-
in-plain-output is a strong standalone ask.

**Visibility of empty project.** When an entry has no `project`
set, the column should render as `-` or `(none)` rather than an
empty field, so the tab-separated shape stays consistent.

### Triage history — ideas-dump resolved 2026-04-22

Full ideas-dump (external Claude evaluation) was triaged on
2026-04-22 into the Post-triage scope block above, the
STAGE-004 sketch above, and `backlog.md` (new file). Nothing
pending. Items in `backlog.md` have revisit triggers and can be
promoted into future stages or projects when the trigger fires.

---

## Project-Level Reflection

*Filled in when status moves to shipped.*

- **Did we deliver the outcome in "What This Project Is"?** <not yet>
- **How many stages did it actually take?** <not yet>
- **What changed between starting and shipping?** <not yet>
- **Lessons that should update AGENTS.md, templates, or constraints?**
  - <not yet>
- **What did we defer to the next project?**
  - <not yet>
