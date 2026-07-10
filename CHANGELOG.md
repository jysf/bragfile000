# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.4.0] - 2026-07-07

The **story surface** release. bragfile grows from "capture and list" into a
read/story layer that turns a corpus of brags into calendar-windowed digests, an
audience-shaped narrative bundle, and a shareable year/quarter-in-review — plus
in-terminal cadence sparklines and a personal agent-assist measure. Everything is
**additive** and **local-first** (no model, no network, no secrets in the binary):
the narrative shaping is a **pure pipe** — bragfile owns the data and the shaping,
an LLM already in your workflow (an agent or a paste-in session) is the optional
upgrade to polished prose. **No schema change, no migration, no CLI breakage** —
every existing command behaves exactly as before.

### Added

- **`brag impact`** — a calendar-windowed, initiative-grouped impact digest and
  the fourth consumer of the DEC-014 rule-based digest envelope. Selects the
  entries carrying an `impact` statement over a required window
  (`--quarter|--month|--year|--since <date>`, mutually exclusive), groups them by
  project (= initiative), and renders each impact in full. `--format
  markdown|json` (default `markdown`); the `--project`/`--type`/`--tag` filters
  compose with the window. Local-first and deterministic — a report, not an LLM
  feature.
- **`brag story --audience me|manager|skip|exec`** — the narrative surface that
  answers "tell the story of my work, shaped for who's listening." Related brags
  become **beats in an arc**, not bullets in a list, at an altitude set by the
  audience: `me` (reflective, year-default), `manager` and `skip` (the middle of
  the gradient), and `exec` (high-altitude, quarter-default). `--audience` is
  required; the window flags reuse `impact`'s calendar machinery, or fall back to
  the audience profile's default window. The command emits a **shaped bundle**
  (`--format markdown|json`) that is useful standalone *and* pasteable into an
  LLM, plus an embedded **framing directive** (`--print-directive`) that tells the
  model how to voice it — **no model, no network** in bragfile itself; the LLM is
  the optional last mile. Audiences are **profiles-as-data** (bundled
  `profiles/*.yaml` + `directives/*.md`, user-overridable): adding one needs no
  code change.
- **`brag wrapped [year|quarter]`** — a shareable, celebratory year- or
  quarter-in-review digest ("your year in brags") and the fifth DEC-014 consumer.
  Curates a retrospective highlight reel over a named calendar period; quarterly
  is first-class (companies report by the quarter). `--format markdown|json`.
- **In-terminal cadence sparklines.** `brag wrapped`'s `## Cadence` section now
  renders a Unicode block-glyph sparkline (`▁▂▃▄▅▆▇█`) over its period counts —
  **local-first, zero new dependency, no network** (pure-Go block characters).
  Default-on in a terminal; escaped by `--no-spark` or `NO_COLOR`. JSON output
  stays raw (a sparkline is a visual rendering, not data — no glyphs enter any
  envelope).
- **`--previous`** — a last-completed-period window modifier for the
  calendar-windowed story commands (`impact`, `story`, `wrapped`). Shifts a window
  from the current period to the previous **completed** one — "last quarter" /
  "last month" / "last year" — as a bounded `[prev-start, prev-end)` window via
  calendar math (never day subtraction, so year boundaries roll correctly). The
  bare-command default is unchanged (still the current period); `--previous` is
  the uniform opt-in.
- **`brag coverage`** — a personal agent-assist measure and the sixth DEC-014
  consumer. Reports **provenance share** (agent- vs human-authored counts and
  share, bucketed by month) over the reserved `agent:`/`model:` provenance corpus,
  a **monthly agent-share trend** rendered as a sparkline, and a **self-reference
  density** measure (entries mentioning `brag`/`bragfile`). A rule-based read over
  existing data — **no schema change**; the classifier is single-sourced with
  `brag list --author` so the two never drift.

### Upgrading from v0.3.1

No manual steps and **no migration** — v0.4.0 is entirely read-side and adds no
schema changes and no breaking CLI changes (every v0.3.1 command behaves
identically). `brew upgrade jysf/bragfile/bragfile` moves a v0.3.1 install to
v0.4.0 in place; `brag --version` then reports `0.4.0`. On a first tap install,
the two one-time frictions still apply: on **Homebrew 6.0+**, run `brew trust
--cask jysf/bragfile/bragfile` once; on **macOS**, clear an unsigned binary's
Gatekeeper quarantine with `xattr -dr com.apple.quarantine` (see the README
install note). To surface the new commands inside Claude Code, reinstall the
plugin so it runs the v0.4.0 binary.

## [0.3.1] - 2026-07-06

A small, additive **patch** that begins seeding per-work economics history. The
MCP `brag_add` tool now accepts optional `session` / `cost` / `tokens` inputs
and stamps them as reserved `session:` / `cost:` / `tokens:` tags, and the
plugin's capture-nudge hook forwards the Claude Code `session_id` so an
agent-captured entry carries a stable session join-key. No schema change, no CLI
change — cost/session history simply starts accruing now, ahead of the reporting
layer that will read it.

### Added

- **Optional cost / session / token capture on `brag_add` (MCP).** The MCP
  `brag_add` tool accepts three new **optional** inputs — `session`, `cost`,
  `tokens` — and stamps each as a reserved-namespace tag (`session:<id>`,
  `cost:<n>`, `tokens:<n>`) alongside the existing `agent:` / `model:`
  provenance. All three are optional: an omitted input stamps no tag, and
  bragfile never fabricates a value. `cost` must be a non-negative USD decimal
  and `tokens` a non-negative integer — a non-numeric or negative value is
  rejected as a tool error rather than silently stored. Reserved but **not**
  author-provenance: a `session:` / `cost:` / `tokens:`-only entry still
  classifies as `human` under `brag list --author` (DEC-027).
- **Session join-key forwarding in the capture-nudge hook.** The Claude Code
  plugin's session-end capture-nudge hook now surfaces the Claude Code
  `session_id` in its agent-facing context and instructs Claude to forward it as
  the `session` input on `brag_add`, so agent-captured entries carry a stable
  per-session join-key. The hook still never runs `brag` itself; its
  silent-degradation and once-per-session contracts are unchanged.

### Upgrading from v0.3.0

No manual steps and **no migration** — v0.3.1 adds no schema changes (the new
tags ride the existing taggings join) and no CLI changes (the capture is
MCP-path-only). `brew upgrade jysf/bragfile/bragfile` moves a v0.3.0 install to
v0.3.1 in place; `brag --version` then reports `0.3.1`. On a first tap install,
the two one-time frictions still apply: on **Homebrew 6.0+**, run `brew trust
--cask jysf/bragfile/bragfile` once; on **macOS**, clear an unsigned binary's
Gatekeeper quarantine with `xattr -dr com.apple.quarantine` (see the README
install note). To pick up the new capture behavior, reinstall the Claude Code
plugin so it runs the v0.3.1 binary.

## [0.3.0] - 2026-07-05

This release makes bragfile **agent-native**. A local MCP server lets an
agent capture and recall brags through native tool calls — no shell, no
network — and agent-written entries label themselves with reserved
`agent:`/`model:` provenance tags. The whole surface installs as a Claude
Code plugin. Capture also gets more delightful (milestone notifications),
and the current-streak metric now reads correctly.

### Added

- `brag mcp serve` — a local stdio MCP server exposing `brag_add`,
  `brag_list`, `brag_search`, and `brag_stats` as native tools over your
  existing database (local-only, no network), so an MCP-client agent
  captures and recalls brags without a shell.
- **Agent/model provenance.** The MCP `brag_add` tool stamps the calling
  agent and model as reserved `agent:<name>` / `model:<id>` tags, making
  agent-authored entries attributable — with no schema change.
- `brag list --author agent|human` — filter entries by provenance
  authorship: `agent` selects entries carrying an `agent:`/`model:` tag,
  `human` selects the rest (`brag list --author agent --format json | jq
  length` counts agent-authored entries).
- **Milestone notifications.** `brag add` prints one celebratory line to
  stderr when you cross a total, streak, or per-project milestone — TTY-only,
  and silent under `--json` and in pipes.
- **Claude Code plugin.** bragfile ships as an installable Claude Code plugin
  bundling `brag mcp serve`, a `/brag` slash-command, and a quiet session-end
  capture-nudge hook; the plugin documents the reserved provenance convention.

### Fixed

- **Current-streak is correct.** `brag stats` keeps the current streak alive
  through *yesterday* and buckets by your *local* day, so it reads correctly
  before the day's first entry (previously it read 0). Storage timestamps
  stay UTC RFC3339; only the derived metric is localized.

### Upgrading from v0.2.x

No manual steps and **no migration** — v0.3.0 adds no schema changes.
`brew upgrade jysf/bragfile/bragfile` moves a v0.2.x install to v0.3.0 in
place; `brag --version` then reports `0.3.0`. Two one-time frictions on
first tap install: on **Homebrew 6.0+**, run `brew trust --cask
jysf/bragfile/bragfile` once; on **macOS**, an unsigned binary may trigger a
Gatekeeper prompt — clear it with `xattr -dr com.apple.quarantine` (see the
README install note). To use the Claude Code plugin, `brag` must resolve on
your `PATH` (the plugin runs the Homebrew-installed binary).

## [0.2.0] - 2026-06-17

This release makes **tags** and **projects** first-class. Tags move from
a comma-joined string to a normalized, shared, rename/merge-able model;
projects become a managed entity with filesystem locations and cwd-aware
auto-fill. Schema migrations now snapshot your database before they run.

### Added

- `brag tags` — list every tag with its usage count.
- `brag tag rename <old> <new>` and `brag tag merge <src> <dst>` —
  first-class tag maintenance. `rename` re-labels a tag in place;
  `merge` folds one tag's entries into another and de-duplicates.
- `brag project` — manage named projects backed by filesystem paths,
  with subcommands `new`, `list`, `show`, `edit`, `archive`, `delete`,
  `status`, and `here`. `brag project here` reports the project owning
  the current directory; `brag project status` prints a per-project
  dashboard.
- `brag project edit` takes `--add-path` / `--remove-path` to attach or
  detach directories from a project.
- `brag add` now auto-fills `--project` from the current directory when
  the cwd sits under a registered project location (nearest-ancestor
  match). An explicit `--project` always wins.
- `brag completion <shell>` — generate tab-completion scripts for zsh,
  bash, and fish. Source into your shell rc for `brag <tab>` and flag
  completion.

### Changed

- **Tags are now first-class.** They are stored in a normalized
  `tags` + `taggings` model instead of a comma-joined string, so a tag
  is shared across entries and can be renamed or merged. Existing
  entries migrate automatically on first run; the `--tag` filter and
  every entry command behave the same for users.
- **Schema migrations back up your database first.** Applying a
  schema-bumping migration to an existing, non-empty database now writes
  a timestamped snapshot beside it (via SQLite `VACUUM INTO`, WAL-safe)
  before the migration runs — so an upgrade can never mutate an
  un-backed-up database. If the backup fails, the upgrade aborts rather
  than proceeding. Non-interactive: safe in `brag add --json` and other
  piped, non-TTY workflows.

### Upgrading from v0.1.x

No manual steps. `brew upgrade bragfile` (or any newer binary) migrates
your existing `~/.bragfile` database in place on the first command you
run — tags and entries carry forward losslessly. The migration writes a
timestamped `*.backup` snapshot beside your database first, so the
upgrade is recoverable. The upgrade is one-way: a v0.1.x binary cannot
read a v0.2.0 database afterward (restore the snapshot if you need to go
back). On macOS, an unsigned binary may trigger a Gatekeeper prompt on
first run — see the README's Gatekeeper note for the one-time `xattr`
clear.

### Decisions of record

The following architectural decisions are committed in this release.
Each decision file under `/decisions/` carries the full rationale.

- DEC-015 — normalize tags into a polymorphic `tags` + `taggings`
  model (supersedes DEC-004's comma-joined string).
- DEC-016 — tag mutation semantics: `rename` errors into an existing
  tag, `merge` de-dups via DELETE+INSERT, orphaned tags are invisible
  (no garbage collection).
- DEC-017 — `entries.project` relates to `projects` by soft string
  match (project stays free text on the entry; no hard foreign key).
- DEC-018 — `brag project delete` blast radius: what a delete removes
  and what it leaves behind.
- DEC-019 — `brag project here` resolves the cwd by nearest-ancestor
  (longest-prefix) matching.
- DEC-020 — `brag project edit` location-editing semantics
  (`--add-path` / `--remove-path`).
- DEC-021 — migration auto-backup durability model: trigger on
  pending-migration-meets-non-empty-DB, snapshot via `VACUUM INTO`,
  abort `storage.Open` if the backup fails.

## [0.1.0] - 2026-05-10

Initial public release of `brag`, a local-first Go CLI for capturing
and retrieving career-worthy moments. Entries live in an embedded
SQLite database at `~/.bragfile/db.sqlite`. No cloud, no sync, no
account.

### Added

- `brag add` — capture an entry via flags (`-t/--title`, `-d`, `-T`,
  `-p`, `-k`, `-i`) or via `$EDITOR` against a templated markdown
  buffer.
- `brag add --json` — programmatic capture from stdin, validated
  against the DEC-012 single-object schema (title required;
  optional user-owned fields; server-owned fields tolerated and
  ignored; unknown keys strict-rejected).
- `brag list` — list entries newest-first, with `--project`,
  `--tag`, `--type`, `--since` filters and `--show-project / -P`
  for an extra project column. `--format json|tsv` for
  machine-readable output.
- `brag show` — display a single entry by ID with full metadata.
- `brag edit <id>` — round-trip an entry through `$EDITOR`.
- `brag delete` — delete an entry by ID with `[y/N]` confirmation.
- `brag search <query>` — SQLite FTS5 full-text search across
  title, description, tags, project, and impact.
- `brag export` — bulk export with `--format markdown|json` and the
  same filter flags as `list`. `--out file` to write to disk.
- `brag summary --range week|month` — rule-based aggregation
  grouped by project and type, rendered as markdown or JSON
  (DEC-014 envelope).
- `brag review --week|--month` — entries grouped by project plus
  three reflection questions, designed to be piped into an
  external AI session.
- `brag stats` — six lifetime metrics: total entries, weekly
  rolling average, current streak, longest streak, top-5 tags,
  top-5 projects, corpus span.
- `docs/brag-entry.schema.json` — JSON Schema (draft 2020-12)
  mirroring the `brag add --json` stdin contract for AI-agent
  validation.
- `scripts/claude-code-post-session.sh` + `examples/brag-slash-command.md`
  — reference Claude Code session-end hook and slash-command
  template demonstrating the round-trip.
- macOS (arm64, amd64) and Linux (arm64, amd64) binaries via
  goreleaser.
- Homebrew tap at `github.com/jysf/homebrew-bragfile` —
  `brew install jysf/bragfile/bragfile`.

### Decisions of record

The following architectural decisions are committed in this release.
Each decision file under `/decisions/` carries the full rationale.

- DEC-001 — pure-Go SQLite driver (`modernc.org/sqlite`); no CGO.
- DEC-002 — embedded migrations via `embed.FS`, applied on
  `storage.Open`.
- DEC-003 — config resolution order: `--db` flag → `BRAGFILE_DB`
  env → `~/.bragfile/db.sqlite` default.
- DEC-004 — tags stored as a comma-joined string for MVP.
- DEC-005 — integer `AUTOINCREMENT` primary keys.
- DEC-006 — `spf13/cobra` as the CLI framework.
- DEC-007 — required-flag validation in `RunE` (cobra's
  `MarkFlagRequired` reports errors via stderr + non-zero exit;
  the project owns user-error rendering uniformly).
- DEC-008 — `--since` accepts date (`2026-04-19`) or duration
  (`7d`, `30d`).
- DEC-009 — editor buffer format for `brag add` / `brag edit`
  (markdown front-matter on top, free body below).
- DEC-010 — `brag search` query syntax (auto-tokenize whitespace;
  treat hyphens / dots as literal; phrase-quote multi-word
  fragments).
- DEC-011 — JSON output shape for `brag list --format json` and
  `brag export --format json`: naked array of nine-key entry
  objects; field names match SQL columns verbatim.
- DEC-012 — `brag add --json` stdin schema: single object, title
  required, server-owned fields tolerated-and-ignored, unknown
  keys strict-rejected.
- DEC-013 — markdown export shape for `brag export --format
  markdown` (+ `--flat`): per-entry markdown blocks under
  per-project headings; `--flat` flattens.
- DEC-014 — rule-based output envelope for `brag summary` /
  `brag review` / `brag stats`: single-object JSON envelope with
  `generated_at` / `scope` / `filters` provenance + per-spec
  payload keys; markdown convention reuses DEC-013's provenance
  + summary-block style.

[Unreleased]: https://github.com/jysf/bragfile000/compare/v0.4.0...HEAD
[0.4.0]: https://github.com/jysf/bragfile000/compare/v0.3.1...v0.4.0
[0.3.1]: https://github.com/jysf/bragfile000/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/jysf/bragfile000/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/jysf/bragfile000/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/jysf/bragfile000/releases/tag/v0.1.0
