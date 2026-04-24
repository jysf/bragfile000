---
# Maps to ContextCore epic-level conventions.
# A Stage is a coherent chunk of work within a Project.
# It has a spec backlog and ships as a unit when the backlog is done.

stage:
  id: STAGE-003                     # stable, zero-padded within the project
  status: proposed                  # proposed | active | shipped | cancelled | on_hold
  priority: high
  target_complete: null

project:
  id: PROJ-001                      # parent project
repo:
  id: bragfile

created_at: 2026-04-22
shipped_at: null
---

# STAGE-003: Reports and AI-friendly I/O

## What This Stage Is

Close the value loop: turn the accumulating corpus of captured entries
into something a human can paste into a review doc and something an AI
agent can read and write through a stable, documented interface. When
this stage ships, `brag export --format markdown` produces a
review-ready document (provenance + executive summary + grouped by
project) that replaces the `brag list | while read | brag show` shell
workaround users currently type by hand; `brag export --format json` /
`brag list --format json|tsv` emit structured output on a shape shared
with `brag add --json`'s stdin contract, so a downstream agent can
round-trip entries through the binary without touching SQL.
(SQLite-file export was scoped out to backlog on 2026-04-23 — `cp
~/.bragfile/db.sqlite backup.db` already covers the portable-backup
use case with zero new code; revisit if VACUUM INTO's defragmentation
or cross-process consistency ever becomes a real need.) Plus `brag list -P` surfaces the
`project` field inline so daily scanning answers "what have I been
working on" at a glance.

## Why Now

STAGE-001 shipped foundations (2026-04-20). STAGE-002 shipped capture
+ retrieval + FTS5 search (2026-04-22, 12 days ahead of target). The
author now has real entries accumulating in `~/.bragfile/db.sqlite`
— the input side is solved. STAGE-003 extracts value from that input
along three axes that are all ship-blocking for any useful review-prep
workflow:

1. **Export trio** closes the "I have entries, now what?" gap that
   today requires ad-hoc `sqlite3` queries or a hand-typed shell loop
   over `brag show`. Quarterly reviews, promo packets, and resume
   updates need a durable document, not a pipeline.
2. **Machine-readable I/O** closes the AI-integration loop. The data
   model was designed (per the brief) so agents can read rows and
   POST to an LLM without schema changes; `--format json` on both
   sides is the stable contract that unblocks every downstream
   integration, including STAGE-004's Claude session-end hook and
   any future `brag ai-summary`.
3. **Project visibility in `brag list`** was flagged by the user
   mid-STAGE-002 ("I want to see quickly what projects I have been
   working on"). It's a scan-time nicety that's become load-bearing
   now that corpus size makes visual scanning useful.

No external blockers. All work layers cleanly on STAGE-002's
`Store.List(ListFilter{})`, `Store.Get`, `Store.Add`, and the existing
`renderEntry` helper in `internal/cli/show.go`.

## Success Criteria

- **Project column**: `brag list -P` (or `--show-project`) renders a
  four-column tab-separated output `<id>\t<created_at>\t<project>\t<title>`
  with empty `project` rendering as `-`. Plain `brag list` without the
  flag is byte-identical to STAGE-002 output (no regressions for
  existing scripts).
- **Markdown export**: `brag export --format markdown [filters]`
  produces a single document with (a) a provenance header naming
  the export timestamp, entry count, and echoed filter flags; (b) an
  executive summary block (counts by type, counts by project); (c)
  entries grouped by project, each group's entries rendered
  chronologically ascending using the same shape as `brag show`
  (title heading + metadata table + description), separated by `---`.
  Writes to stdout by default; `--out report.md` writes to a file.
  `--flat` flag produces un-grouped chronological output.
- **JSON I/O shape is unified**: `brag export --format json [filters]`
  and `brag list --format json [filters]` emit a pretty-printed JSON
  array with identical per-entry structure (field names = SQL columns,
  tags as comma-joined string per DEC-004, RFC3339 timestamps).
  `brag list --format tsv` emits a tab-separated variant of the same
  field set with a header row.
- **Stdin-JSON round-trip**: `brag list --format json | jq '.[0]' |
  brag add --json` inserts a new entry whose user-owned fields match
  the source entry. Unknown fields are strict-rejected; server-owned
  fields (`id`, `created_at`, `updated_at`) are ignored if present so
  round-trip works without `jq del`.
- All STAGE-001 + STAGE-002 success criteria still hold. `go test
  ./...`, `gofmt -l .`, `go vet ./...`, `CGO_ENABLED=0 go build
  ./...` remain clean.

## Scope

### In scope

- **`brag list -P` / `--show-project`** — new flag adding project
  column to list output; empty project renders as `-`; flag off
  preserves existing three-column byte-for-byte.
- **`brag export --format markdown|json`** — new command.
  Both formats reuse `ListFilter`. Markdown groups by project by
  default (`--flat` escape). `--format sqlite` deferred to backlog
  on 2026-04-23 (see Explicitly out of scope below).
- **`brag list --format json|tsv`** — augments existing `brag list`
  with machine-readable output formats. Plain (default) stays
  byte-stable.
- **`brag add --json`** — read stdin as a single JSON object,
  validate against the shape-minus-server-fields schema, strict-reject
  unknown keys, insert via `Store.Add`, print inserted ID on stdout
  (same contract as flag-mode `add`).
- **Candidate DECs**:
  - `DEC-011` — shared JSON shape for `export --format json` +
    `list --format json` + the output side of `add --json` round-trips.
    Born in SPEC-014.
  - `DEC-012` — stdin-JSON schema for `brag add --json` (accepted
    keys, strict-reject rule, server-field tolerance). Born in
    SPEC-017.
  - `DEC-013` — markdown export shape (provenance + executive
    summary + default grouping + separator conventions). Born in
    SPEC-015.
- **Lift `renderEntry`** from `internal/cli/show.go` into a shared
  internal package (either `internal/export` or `internal/render`)
  so `brag show` and `brag export --format markdown` stay in sync.
  Drive-by inside SPEC-015 build, disclosed in deviations.
- **Doc sweeps** folded into their originating spec per the premise-
  audit rule: SPEC-013 updates `api-contract.md` + `tutorial.md` for
  the `-P` flag; the export specs update `api-contract.md` +
  `data-model.md` for each new format; SPEC-017 updates the `brag
  add` section of `api-contract.md` + `tutorial.md`. Stale bits
  from the STAGE-002 ship reflection (`README.md:52`, `docs/data-
  model.md:67` listing search/show/edit/delete as "future work")
  fold into SPEC-013 or SPEC-015, whichever lands first.

### Explicitly out of scope

Deferred to backlog (see `backlog.md` for full entries with revisit
triggers):

- **NDJSON / array-batch stdin** for `brag add --json`. Single object
  only for MVP. Revisit if a bulk-import workflow appears.
- **Lenient-accept mode** (`--json-lenient`). Strict-reject is the
  default and only mode; add `--lenient` only if strict proves
  annoying in practice.
- **JSON envelope** (`{generated_at, count, filters, entries: [...]}`).
  Naked array ships; wrap behind `--envelope` if a consumer asks.
- **`--compact` / non-pretty JSON**. Pretty-printed (indent=2) only.
- **`brag export --format sqlite` (full-DB `VACUUM INTO`).** Moved
  to backlog on 2026-04-23 (post-SPEC-013 scope-tightening). `cp
  ~/.bragfile/db.sqlite backup.db` already handles the
  portable-backup use case the brief named; `VACUUM INTO`'s
  marginal wins (defragmentation, WAL-flushed consistency) aren't
  urgent for a personal-CLI workflow. Revisit if real demand
  emerges.
- **Filtered SQLite export** (fresh-DB + INSERT-SELECT path).
  Deferred alongside the full-DB variant above; same revisit
  trigger shape.
- **TOC in markdown export**. Headings are scannable enough for the
  MVP use cases.
- **`--group-by type` or `--group-by <field>`**. Group-by project is
  the default; `--flat` is the only escape. Multi-axis grouping is a
  later polish.
- **`--template <path>`** custom markdown templates.
- **Resume-bullet, HTML, PDF formats** — explicitly out per brief
  "Explicitly out of scope" section.

Deferred to STAGE-004 (polish pass):

- **`--pretty` mode** on `brag list` (emoji + bundled project column).
  STAGE-003 ships standalone `-P` per the 2026-04-22 brief reshuffle;
  STAGE-004 `--pretty` can default `-P` on inside its emoji bundle.
- **`brag summary --range week|month`**.
- **Emoji decoration passes 1–4** (stderr feedback, show/list icons,
  NO_COLOR + TTY detection).
- **Claude Code session-end hook example**. Depends on SPEC-017's
  `brag add --json` landing; will ship as STAGE-004 polish.
- **`brag remind` / `brag stats` / `brag review --week`**.

Deferred to STAGE-005 / never:

- **Distribution** (goreleaser, CI, homebrew tap) — STAGE-005.
- **LLM-backed features** — out of PROJ-001 entirely; belongs to
  PROJ-002 "AI assist" whenever that project opens.
- **Multi-device sync, cloud backup, auth** — out of PROJ-001.

## Spec Backlog

Ordered by recommended build sequence; some specs are independent and
can run in parallel.

- [x] SPEC-013 (shipped 2026-04-23, **S**) — **`brag list
      --show-project / -P`.** New flag adds a fourth tab-separated
      column (`<id>\t<created_at>\t<project>\t<title>`); empty project
      renders as `-`; plain `brag list` byte-identical to STAGE-002.
      Updates `api-contract.md` list section. Shipped via PR #13
      (squash-merged `7f802a2`). Clean build+verify cycle — no DEC
      emitted, no deviations, no follow-ups.

- [x] SPEC-014 (shipped 2026-04-23, **M**) — **`brag list
      --format json|tsv` + `brag export --format json` + DEC-011
      (shared JSON shape).** DEC-011 locks six choices (naked array,
      SQL-column field order, comma-joined tags per DEC-004, RFC3339
      timestamps, empty-string-not-omit, indent=2). Wires two
      consumers through shared `internal/export.ToJSON`:
      `list --format json|tsv` (default plain unchanged) and
      `export --format json` with filter flags reusing `ListFilter`.
      Shipped via PR #14 (squash-merged `9c52ad1`). Clean
      build+verify cycle — no build-time DECs, no deviations, no
      follow-ups. New `internal/export` package created to host
      `ToJSON` + `TSVHeader` + `ToTSVRow` (sanctioned per spec's
      recommendation; SPEC-015's markdown renderEntry lift will be
      purely additive here). Sized M honestly at design time (stage
      framed it as S; 14 tests, new package, new command, new DEC
      pushed it to M).

- [ ] SPEC-015 (design, **M**) — **`brag export --format markdown` +
      DEC-013 (markdown export shape).** Largest spec in the stage.
      Lifts `renderEntry` into a shared helper (drive-by, disclosed),
      adds provenance header, executive summary block, default
      grouping by project (groups ordered alphabetically, entries
      within a group chrono-ASC), `--flat` escape, `---` separator
      between entries, `--out file.md` writer. Filter-aware via
      `ListFilter` reuse. Sized M because five distinct rendering
      components compose into one output; if design-session discovers
      the grouping + summary work is heavier than expected, split
      grouping into a follow-up S rather than stretching to L.

- [~] SPEC-016 — **Deferred to backlog on 2026-04-23** (post-
      SPEC-013 scope-tightening). Slot number preserved for
      traceability; SPEC-017 is not renumbered. Original scope:
      `brag export --format sqlite` via `VACUUM INTO`. Deferred
      because `cp ~/.bragfile/db.sqlite backup.db` already covers
      the portable-backup use case. Full entry in `backlog.md` →
      "`brag export --format sqlite` (full-DB VACUUM INTO)".

- [ ] SPEC-017 (design, **S**) — **`brag add --json` + DEC-012
      (stdin-JSON schema).** Stdin as a single JSON object; DEC-012
      pins the accepted shape (user-owned fields only; strict-reject
      unknown keys; server-owned `id`/`created_at`/`updated_at`
      tolerated-and-ignored so round-trip works). Reuses `Store.Add`.
      Depends on DEC-011 having landed (shape consistency between
      what comes out of `list --format json` and what goes into
      `add --json` minus server fields).

**Count:** 2 shipped / 0 active / 2 pending / 1 deferred

**Complexity check:** 1 × S remaining (SPEC-017), 1 × M
remaining (SPEC-015). DEC-011 shipped in SPEC-014 — unblocks
both pending specs. Build sequence: SPEC-015 || SPEC-017 (both
read DEC-011; can run parallel fresh-session if context
allows).

## Design Notes

Cross-cutting patterns that span multiple specs in this stage. All
AGENTS.md §9 lessons (buffer split, tie-break, assertion specificity,
locked-decisions-need-tests, premise-audit trio) apply unchanged.

- **Shared JSON shape (DEC-011 scope).** One DEC binds
  `list --format json`, `export --format json`, and the round-trip
  shape-minus-server-fields contract for `add --json`. Field names
  match SQL columns exactly (`id, title, description, tags, project,
  type, impact, created_at, updated_at`). Tags stay a comma-joined
  string per DEC-004 — don't re-normalize at the I/O boundary.
  Timestamps are RFC3339 strings (matches storage). Empty-string
  fields stay as `""`, not omitted — keeps AI-consumer schemas
  stable. Pretty-printed (indent=2) by default; no `--compact` for
  MVP. No top-level envelope (naked array) so `jq '.[]'` and the
  round-trip with `add --json` both stay simple.

- **Stdin-JSON schema (DEC-012 scope).** Accepts the same shape
  `list --format json` emits, minus server-owned fields. Required:
  `title` non-empty. Optional: `description, tags, project, type,
  impact`. `id`, `created_at`, `updated_at` are tolerated-and-ignored
  if present (so `brag list --format json | jq '.[0]' | brag add
  --json` works without `jq del(.id, .created_at, .updated_at)`).
  Unknown keys are strict-rejected with the offending key named in
  the error — catches typos like `titl` or `descripton` before they
  become silently-missing entries. Single object only; batch / NDJSON
  / array input deferred to backlog. Output (stdout on success) is
  the inserted ID, matching flag-mode `add` contract.

- **Markdown export shape (DEC-013 scope).** Top of document:
  `# Bragfile Export` heading, then a provenance block with
  `Exported: <RFC3339 timestamp>`, `Entries: <N>`, and a `Filters:`
  line echoing the flags that were passed (or `(none)`). Then an
  executive summary block: counts by type (`shipped: 18, learned:
  11, fixed: 8, ...`) and counts by project. Then entries grouped by
  project, sections headed `## <project name>` (alphabetical;
  entries without a project go under `## (no project)`, rendered
  last). Within each group: entries chronologically ascending
  (ASC, not DESC — review docs read forward through time), each
  entry rendered via the lifted `renderEntry` helper (title as `###
  <title>` inside groups, metadata table, optional `#### Description`
  body), separated by `---`. `--flat` flag skips the grouping and
  renders entries chrono-ASC under a single section. `--out file.md`
  writes to a file instead of stdout.

- **`renderEntry` lift.** The helper in `internal/cli/show.go:70`
  becomes the canonical per-entry markdown renderer. SPEC-015 moves
  it to a shared internal package (`internal/export` is the natural
  home given the architecture.md note about STAGE-003's
  `internal/export`). `brag show` continues to use it (re-imported);
  `brag export --format markdown` calls it per-entry inside the
  grouping loop. Keeps both commands' rendering in sync — any future
  metadata field lands in one place. Heading level is the one
  contextual difference: `brag show` uses `#` for title; markdown
  export uses `###` so entries nest under the `##`-project section.
  The helper takes a heading-level argument.

- **Project column rendering (SPEC-013).** Standalone `-P` /
  `--show-project` flag; default off to preserve the STAGE-002
  three-column contract for scripts. Empty project renders as `-`
  (dash), not empty string, so the tab-separated shape stays
  consistent and downstream `cut -f3` never collapses. `--pretty`
  is deliberately NOT introduced in STAGE-003 — it arrives in
  STAGE-004 as the emoji+project bundle and can default `-P` on
  inside its bundle.

- **SQLite export mechanism — DEFERRED 2026-04-23.** Originally
  SPEC-016; the `VACUUM INTO <path>` design is preserved in
  `backlog.md` under the full-DB entry for when/if the work is
  pulled back in. Left here as a historical marker so the stage
  file reads straight for future maintainers.

- **Filter flag reuse.** `brag export --format markdown|json` accepts
  the same `--tag / --project / --type / --since / --limit` flags as
  `brag list` (SPEC-007). `ListFilter` is the shared input struct.
  No new filter logic is written in this stage — all of it exists
  in `Store.List` already.

- **`--out <path>` semantics.** Markdown + JSON default to stdout
  when `--out` absent. If `--out` points to an existing file:
  overwrite (match `goreleaser` / `jq -o` conventions; no prompt).
  Directory or unwritable path: exit 2 (internal error, not user
  error — disk state issue).

- **Premise audit discipline (AGENTS.md §9 three cases) applies to
  every spec in this stage.** Specific hot spots to audit in design:
  - **SPEC-013 (status-change)**: grep `grep -rn 'brag list' docs/
    README.md` and update every documented example that now has a
    `-P` option; grep for "3 tab-separated columns" / "three columns"
    in docs.
  - **SPEC-014 (addition)**: `list --format json` is new behavior;
    audit `TestList_*` family for implicit plain-format assumptions
    that need a mode-explicit rewrite. `export --format json` is
    also a new command surface — audit help-command tests that
    assert on command counts.
  - **SPEC-015 (addition)**: `brag export` appears in
    `api-contract.md` as a STAGE-003 stub with `--format markdown`
    already documented; audit that doc section for drift between the
    stub and the shipped behavior. Lifting `renderEntry` changes its
    import path — audit every test that imports from
    `internal/cli`.
  - ~~**SPEC-016**~~: deferred 2026-04-23 — no premise-audit
    scope; the work moved to `backlog.md`.
  - **SPEC-017 (addition + status-change)**: adds `--json` flag to
    `brag add`; the dispatch rule in the existing `add` command
    (flag mode vs editor mode, established by SPEC-010's DEC) needs
    a new branch for `--json` mode. Audit `runAdd` dispatcher for
    the dispatch rule change.

- **DEC-007 carries forward.** All new flag validation (empty
  `--format`, invalid `--format` value, bad JSON on stdin) goes
  through `UserErrorf` in `RunE`, never `MarkFlagRequired`.

- **CLI test harness.** Separate `outBuf` / `errBuf` per §9;
  `id DESC` tie-break on any ordering test; `fail-first` run before
  implementation; every locked decision paired with a failing test
  per SPEC-009 ship lesson.

## Dependencies

### Depends on

- **STAGE-002 (shipped 2026-04-22)** — provides `Store.List(ListFilter)`
  with all filter fields, `Store.Get(id)`, `Store.Add(entry)`,
  `Store.Update`, the existing `renderEntry` helper in
  `internal/cli/show.go:70`, editor-launch dispatch pattern in
  `runAdd` (which SPEC-017's `--json` mode extends as a third branch),
  and the schema as currently shipped (`entries` + `schema_migrations`
  + `entries_fts`).
- **DEC-001 through DEC-010** — all apply forward unchanged.
- **External:** none. stdlib `encoding/json` covers JSON. No new Go
  module dependencies expected; if a spec proposes one it needs its
  own DEC per `no-new-top-level-deps-without-decision`.

### Enables

- **STAGE-004 (polish pass)**. `--pretty` bundles the STAGE-003
  `-P` column with emoji. Claude session-end hook depends on
  SPEC-017's `brag add --json`. `brag summary` can reuse the
  executive-summary rendering helper from SPEC-015.
- **STAGE-005 (distribution)**. Shipping the machine-readable I/O
  loop in PROJ-001 means the `brew install bragfile` release is a
  full MVP, not a "CLI that can't be piped" release. Agents can
  integrate on day one of the tap going public.
- **PROJ-002 (AI assist), when opened**. DEC-011's JSON shape is
  the stable contract that the future `brag ai-summary` will POST
  to an LLM. DEC-012's stdin shape is how that project will write
  suggested entries back.

## Stage-Level Reflection

*Filled in when status moves to shipped. Run Prompt 1c (Stage Ship) in
FIRST_SESSION_PROMPTS.md to draft this.*

- **Did we deliver the outcome in "What This Stage Is"?** <yes/no + notes>
- **How many specs did it actually take?** <number vs. plan>
- **What changed between starting and shipping?** <one sentence>
- **Lessons that should update AGENTS.md, templates, or constraints?**
  - <one-line updates>
- **Should any spec-level reflections be promoted to stage-level lessons?**
  - <one-line items>
