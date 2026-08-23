# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed

- **`brag memory`'s headline count is now `Candidates: <N>`, not
  `Entries: <N>`** ([DEC-048](decisions/DEC-048-provenance-count-names-what-it-counted.md)).
  The number was never the corpus size and never a cap: it is the deduped
  union of up to three 200-row reads, so it *grew* when you passed a flag —
  `200` bare and `243` with `--project` on the same 387-entry corpus, while
  `brag export --project` correctly *narrowed* 387 → 74. Same word, same
  flag, opposite directions. The five other commands that print `Entries:`
  are unchanged and correct; on those it means entries in scope.
  `brag://memory/recent` and `brag://memory/project/{name}` carry the new
  header too, since they are byte-identical to `brag memory`.
- **Breaking: `brag memory --format json` (and the `brag_memory` MCP tool
  with `format: "json"`) renames the `entries` key to `candidates`.** Same
  number, honest name — and it removes a collision with the `entries` key
  that means *an array of entry objects* on `brag impact`, `brag review`,
  `brag summary` and `brag wrapped`. Update any `jq .entries` to
  `jq .candidates`.

## [0.6.1] - 2026-08-13

A correctness-and-edges release. **No schema change, no migration, no new
dependency.** Seven fixes from the v0.5.0 audit backlog, plus `go install`
finally reporting its own version.

### Added

- **`go install github.com/jysf/bragfile000/cmd/brag@latest` is a documented,
  working install path.** It always worked — bragfile is a public Go module
  with a pure-Go SQLite driver and `CGO_ENABLED=0`, so no toolchain setup is
  needed — but it was undocumented and reported `brag version dev`, because
  `go install` does not run goreleaser's ldflags. The version is now recovered
  from the embedded build info. Deliberately narrow: only a clean `vX.Y.Z` tag
  is accepted, so a pseudo-version (an untagged build) or `(devel)` still
  reports `dev` and still trips the DEC-026 guard that stops an unreleased
  binary migrating your real database.

### Fixed

- **`brag mcp install` writes atomically** (temp file + rename, in the same
  directory). It rewrites a config file you did not author — a client's
  `.mcp.json` can list every other MCP server you have registered — and the
  previous plain write truncated before writing, so an interrupted run could
  leave it empty.
- **Every command's db-path error said `open store:` twice.** `storage.Open`
  already supplies that context and all 31 CLI call sites added it again.
- **`brag search -foo` (and any malformed flag) now exits 1, not 2.** Exit 2 is
  reserved for internal faults; a bad flag is user-actionable. The message was
  already correct — only the exit status was wrong, which mattered to scripts.
- **`brag export --format markdown` is deterministic for same-second entries.**
  `created_at` is second-resolution and the sort had no tie-break, so entries
  captured in the same second were ordered by whatever the database returned.
  Now ordered by `(created_at, id)` in both grouped and flat modes.
- **`brag spark` no longer drops an entry captured in the same second it
  runs** — reachable whenever a capture hook is followed immediately by a
  spark.
- **An untyped entry renders as `-` in export's "By type"** instead of a bare,
  blank-labelled `- : 3` row, matching what `brag memory` and `brag list`
  already do.
- **Pre-migration backup filenames carry sub-second precision.** Two migrating
  opens in the same second produced the same sidecar name; `VACUUM INTO`
  refuses an existing destination and DEC-021 turns a failed backup into an
  aborted open, so the second process could not open the database at all.

### Documentation

- The README status line and the tutorial's "shipped as of" line said **v0.5.1
  through two releases** (v0.5.2 and v0.6.0). Corrected, and now pinned by a
  test that derives the expected version from the CHANGELOG, so they cannot
  silently rot again.

## [0.6.0] - 2026-08-10

The **agent-native depth** release: your corpus becomes something an agent reads
*before* it works, instead of a write-only log.

> ### ⚠️ Upgrading from 0.5.1 or earlier — one-time manual step
>
> v0.5.2 moved distribution from a Homebrew **cask** to a **formula** on a new
> tap (DEC-040). `brew upgrade` does **not** cross that boundary: if you
> installed before v0.5.2 you are still on the old cask, `brew outdated` says
> nothing, and you will not receive this release. Check with `brag --version`,
> and if it prints `0.5.1` or lower:
>
> ```bash
> brew uninstall --cask bragfile
> brew untap jysf/bragfile
> brew install jysf/tap/bragfile
> ```
>
> Your database is untouched by this — it lives in `~/.bragfile`, not in the
> Cellar or Caskroom. Verify with `brag --version` (expect `0.6.0`) and
> `brag list`.

### Added

- **`brag memory` — the corpus as working memory** (SPEC-073 / DEC-043, DEC-044).
  A ranked, token-budgeted slice of your own history, cheap enough for an agent
  to load at the start of every session. Reading your history used to mean
  picking a single axis — `brag list` for recency or `brag search` for relevance
  — and getting back an unbounded row count. `brag memory` blends three signals
  (recency, `--query` match, `--project` membership) by **reciprocal-rank
  fusion**, so two incomparable orderings need no invented score normalization,
  and trims the result to a **token budget rather than a row count** (`--budget`,
  2000 by default) because a row count does not predict what a slice costs to
  read. Entries pack greedily in rank order; one that does not fit is skipped and
  the fill continues, so a single long entry cannot starve the rest. `--project`
  is a soft boost, not a filter. Rule-based and deterministic: no LLM, no
  network. Markdown (default) or `--format json` under the DEC-014 envelope.

- **MCP push surface: `brag://` resources + the `brag_memory` tool** (SPEC-074 /
  DEC-045). Every read of the corpus used to be **pull** — an agent saw your
  history only if it decided to call a tool. Three resources now let an MCP
  client **auto-load** the slice with no tool call and no configuration:
  `brag://memory/recent`, `brag://memory/project/{name}`, and `brag://projects`
  (every registered non-archived project, so an agent uses real names verbatim
  instead of inventing one). The memory resources serve byte-identical output to
  the corresponding `brag memory` invocation, minus the single trailing newline
  that command appends when printing. `brag_memory` is the parameterized
  pull counterpart for when an agent wants a specific budget, query, or project.

- **MCP `brag_list` time-window and provenance filters** (SPEC-072 / DEC-042).
  The tool now accepts `since`, `until`, `day`, and `author` alongside
  `tag`/`project`/`type`/`limit`, so a connected agent can ask for "the last
  week on this project" or "what agents logged" without pulling the whole
  corpus and filtering client-side. The grammar is identical to the CLI's —
  `since`/`until` take DEC-008's `YYYY-MM-DD`|`Nd`/`Nw`/`Nm`, `day` takes
  DEC-039's `YYYY-MM-DD`|`today`|`yesterday` local calendar day — because both
  surfaces now call the same parser (`internal/timewindow`), retiring the
  `cli↔mcpserver` import cycle that deferred `--since` at SPEC-040. `until` is
  an exclusive upper bound and is MCP-only by design; the CLI keeps `--day`.

### Changed

- **Capture field caps are derived from the corpus instead of inherited**
  (SPEC-075 / DEC-046). The caps `capture.Validate` enforces on every ingress
  path had no recorded rationale — DEC-012 introduced them for the `brag add
  --json` schema, and SPEC-064 then correctly generalised them to every path,
  with an unmeasured side effect: **74 of 285 entries with an `impact` (26%)
  were over cap**, so a large slice of your own history had become unwritable.
  Measured against a 359-entry snapshot and re-shaped per field:
  - `impact` **256 → 1024** — it was at 1.12× its own median (p50 228 against a
    256 cap), the only field without the headroom `project` and `type` had.
  - `title` **200 → 256** — deliberately *not* widened to fit its tail, which is
    a single malformed-capture episode rather than ordinary use.
  - the joined-string tags cap is replaced by **`MaxTagLen` 64 + `MaxTagCount`
    32**. The old cap was inverted from the abuse it should catch: every
    over-cap tag string held 7–9 individually short tags (longest 17 bytes),
    so it penalised using many legitimate tags rather than one absurd one.

  `project`, `type` and `description` are unchanged. No schema change, no
  migration.

### Fixed

- **`brag edit` now validates what it writes** (SPEC-075 / DEC-046). The edit
  path silently enforced no caps at all. It now runs `capture.ValidateChanged`,
  which validates a field **only when its value changed** — so the caps reach
  the edit path without making the already-over-cap corpus uneditable, and
  without a migration. Provenance tags stamped by the MCP server
  (`agent:`/`model:`/`session:`/`cost:`/`tokens:`) are excluded from the tag
  caps on this path, since they are not user-supplied text.

- **A negative `limit` is now a tool error on MCP `brag_list` / `brag_search`**
  instead of silently meaning "unlimited" (v0.5.0 pre-release audit item;
  `brag list --limit -1` already errored). `0` or omitted still means
  unlimited.

## [0.5.2] - 2026-07-30

### Changed

- **Distribution: Homebrew cask → binary formula, on the shared `jysf/homebrew-tap`**
  (DEC-040). Install is now `brew install jysf/tap/bragfile`. Because Homebrew
  formulae are not Gatekeeper-quarantined, **the macOS "Apple could not verify…"
  prompt and the `xattr -dr com.apple.quarantine` workaround are gone** — no code
  signing required. `--version` is unaffected (still the ldflags-stamped release
  binary). This reverses the undocumented v0.1.0 cask switch that introduced the
  signing friction. Note: `goreleaser` emits a `brews:` deprecation warning,
  accepted deliberately per DEC-040. The old per-project `homebrew-bragfile` tap
  is retired.

## [0.5.1] - 2026-07-11

A small ergonomics release. **No schema change, no migration.**

### Added

- **`brag list --day <YYYY-MM-DD|today|yesterday>`** — scope a listing to a single
  **local** calendar day (the half-open `[local-midnight, next-local-midnight)`
  window), built on the `ListFilter.Until` bound from v0.5.0. `today`/`yesterday`
  resolve against your local clock (so an entry logged at 9pm your time reads as
  that day, not the next UTC day); mutually exclusive with `--since`; composes
  with `--project`/`--type`/`--tag`/`--limit`/`-P`/`--format`. (DEC-039)

### Fixed

- `--since <duration>` and the new `--day` keywords now resolve the current time
  through an injectable clock seam instead of reading the wall clock inline — an
  internal testability fix with no behavior change to `--since`.

## [0.5.0] - 2026-07-10

The **agent-native depth (opening)** release. bragfile makes its MCP path
first-class for AI agents and hardens the substrate underneath. Local-first as
ever — no network in the binary, no CGO, and **no schema change or migration**.
Capture input validation is now enforced consistently across every ingress path
(a title/field that the CLI accepts is exactly what `--json`/MCP accept).

### Added

- **`brag mcp install [--client claude-code|claude-desktop|cursor] [--scope
  user|project] [--dir PATH] [--dry-run]`** — one command to register the `brag
  mcp serve` MCP server into a client's config. Idempotent and never clobbers
  other servers already in the file (DEC-034).
- **`brag project ensure <name> [--location PATH]`** — idempotent project
  registration (create-or-no-op), closing the unregistered-project gap so
  entries map cleanly for downstream consumers (DEC-036).
- **`brag spark [--week|--month|--quarter] [--project <name>]`** — a
  sparklines-only "pulse" (a Total row + top-8 by-project) over a rolling recent
  window; markdown default, `--format json` for raw counts, `--no-spark`/
  `NO_COLOR` to suppress glyphs (DEC-037).
- **`docs/for-ai-agents.md`** + a README "Using brag from an AI agent (MCP)"
  section — the full MCP tool contract (schemas for `brag_add`/`brag_list`/
  `brag_search`/`brag_stats`), the `project`-not-auto-filled gotcha, provenance
  stamping, and a how-to-log-a-win playbook.
- The `sprint:<id>` freeform-tag convention, documented in the tutorial (sprint
  is just a tag — no schema field).
- `just lifetime-report` — a dated whole-repo lifetime-report prompt generator
  (workflow tooling).

### Changed

- **Concurrency:** the SQLite database now opens with `busy_timeout` + immediate
  transactions + a single connection, so concurrent access — e.g. `brag mcp
  serve` running while a shell `brag add` or a hook fires — waits and succeeds
  instead of failing with `database is locked` (DEC-038).
- **Capture validation is unified** across all ingress paths (flags, `$EDITOR`,
  `--json`, MCP) through one shared validator: the same field byte-caps and the
  same rejection of embedded control characters everywhere. (A flag/editor-
  captured entry can no longer exceed limits that `--json`/MCP would reject.)
- Internal: the calendar-window upper bound moved into
  `storage.ListFilter.Until`, de-duplicating Go-side filtering across
  `impact`/`story`/`wrapped`/`coverage` (DEC-035).

### Fixed

- `brag tag rename` now canonicalizes/rejects its target, so a comma or blank
  name can no longer silently corrupt tag membership on a later edit.
- `brag mcp serve` exits cleanly (status 0) on a normal client shutdown instead
  of reporting a spurious `server is closing` error.
- Field values containing `|` no longer break markdown tables in `brag show` /
  `brag export --format markdown` (the `|` is escaped).
- `brag spark` no longer counts out-of-window (future-dated) entries in its
  header or top-8 selection.
- Invalid `cost:`/`tokens:` reserved tags supplied via the freeform tags field
  are now rejected, matching the dedicated params.

### Upgrading from v0.4.0

No manual steps and **no migration** — v0.5.0 adds no schema changes. `brew
upgrade jysf/bragfile/bragfile` moves a v0.4.0 install to v0.5.0 in place; `brag
--version` then reports `0.5.0`. On a first tap install, the two one-time
frictions still apply: on **Homebrew 6.0+**, run `brew trust --cask
jysf/bragfile/bragfile` once; on **macOS**, clear an unsigned binary's Gatekeeper
quarantine with `xattr -dr com.apple.quarantine` (see the README install note).
To register the MCP server and slash-command inside your AI client, reinstall
the plugin (or run `brag mcp install`) so it runs the v0.5.0 binary.

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

[Unreleased]: https://github.com/jysf/bragfile000/compare/v0.6.1...HEAD
[0.6.1]: https://github.com/jysf/bragfile000/compare/v0.6.0...v0.6.1
[0.6.0]: https://github.com/jysf/bragfile000/compare/v0.5.2...v0.6.0
[0.5.2]: https://github.com/jysf/bragfile000/compare/v0.5.1...v0.5.2
[0.5.1]: https://github.com/jysf/bragfile000/compare/v0.5.0...v0.5.1
[0.5.0]: https://github.com/jysf/bragfile000/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/jysf/bragfile000/compare/v0.3.1...v0.4.0
[0.3.1]: https://github.com/jysf/bragfile000/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/jysf/bragfile000/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/jysf/bragfile000/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/jysf/bragfile000/releases/tag/v0.1.0
