---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-014
  type: story                      # epic | story | task | bug | chore
  cycle: build
  blocked: false
  priority: high
  complexity: M                    # borderline; called out below. Single-spec per user direction 2026-04-23.

project:
  id: PROJ-001
  stage: STAGE-003
repo:
  id: bragfile

agents:
  architect: claude-opus-4-7
  implementer: claude-opus-4-7     # usually same Claude, different session
  created_at: 2026-04-23

references:
  decisions:
    - DEC-004   # tags comma-joined TEXT — JSON keeps tags as string, not array
    - DEC-006   # cobra framework — new `brag export` subcommand and `--format` flag declarations
    - DEC-007   # required-flag validation in RunE — `--format` missing/unknown returns UserErrorf, never MarkFlagRequired
    - DEC-011   # EMITTED HERE — shared JSON output shape for list/export/add-json round-trip
  constraints:
    - no-sql-in-cli-layer
    - stdout-is-for-data-stderr-is-for-humans
    - errors-wrap-with-context
    - test-before-implementation
    - one-spec-per-pr
  related_specs:
    - SPEC-004   # shipped; original `brag list` 3-column output — flag-off shape preserved under `--format` absent
    - SPEC-007   # shipped; list filter flags (ListFilter struct) — reused verbatim by `brag export`
    - SPEC-013   # shipped; `brag list -P` — test patterns for plain-mode byte-stability under new-flag addition are this spec's template
    - SPEC-015   # design; extends `brag export` with `--format markdown`; reads DEC-011 shape for nothing of its own, but reuses `internal/export` package
    - SPEC-017   # design; reads DEC-011 verbatim for `brag add --json` stdin shape (minus server fields)
---

# SPEC-014: JSON trio and shared shape DEC — `brag list --format json|tsv`, `brag export --format json`, DEC-011

## Context

Second active spec in STAGE-003 (SPEC-013 shipped 2026-04-23). STAGE-003
closes the I/O loop for AI agents and makes `brag export` real. This
spec is the anchor for the JSON-shape side of that loop: it emits
DEC-011 (shared JSON shape) and wires two consumers at once —
`brag list --format json|tsv` (machine-readable output on an existing
command) and `brag export --format json` (new `brag export` command
surface, first spec to ship it). SPEC-015 will extend `brag export`
with `--format markdown`; SPEC-017 will read DEC-011 verbatim for
`brag add --json`'s stdin shape minus server fields.

The load-bearing contribution beyond the two commands is **one shape,
one helper, one byte-identical assertion**: both consumers route
through a new `internal/export` package's `ToJSON()` helper, and a
single test locks that `brag list --format json` and
`brag export --format json` produce byte-identical output on the same
fixture. If that assertion ever fails, DEC-011 has been violated.

Parent stage: `STAGE-003-reports-and-ai-friendly-i-o.md` — Design
Notes → "Shared JSON shape (DEC-011 scope)" is the authoritative
lock for the six choices DEC-011 codifies; Design Notes → "Filter
flag reuse", "`--out <path>` semantics", and "Premise-audit hot
spots → SPEC-014" apply directly. Project: PROJ-001 (MVP).

## Goal

Ship (a) DEC-011 as a new decision file pinning the six JSON-shape
choices; (b) `--format json|tsv` on `brag list`, preserving plain-mode
byte-stability under `--format` absent; (c) `brag export --format json
[--out path]` as the first concrete `brag export` command surface,
reusing `ListFilter` for filter flags; (d) a new `internal/export`
package with `ToJSON(entries []storage.Entry) ([]byte, error)` used by
both consumers — locked so the shape-consistency assertion is a single
test, not a sprawl.

## Inputs

- **Files to read:**
  - `/AGENTS.md` — §6 cycle rules; §7 spec anatomy; §8 DEC emission +
    honest confidence; §9 premise-audit family (SPEC-014 is
    ADDITION + STATUS-CHANGE); §12 CLI test harness rules.
  - `/projects/PROJ-001-mvp/brief.md` — "Detail on individual ideas →
    JSON export shape" (authoritative shape choices).
  - `/projects/PROJ-001-mvp/stages/STAGE-003-reports-and-ai-friendly-i-o.md`
    — Design Notes → "Shared JSON shape (DEC-011 scope)", "Filter
    flag reuse", "`--out <path>` semantics", "Premise-audit hot
    spots → SPEC-014".
  - `/projects/PROJ-001-mvp/backlog.md` — NOT for scope; for
    awareness of deferred siblings (JSON envelope, `--compact`,
    NDJSON stdin, lenient-accept, SQLite export). Do NOT pull from
    here for SPEC-014.
  - `/docs/api-contract.md` — current `brag list` section (lines
    78–96) and STALE `brag export` section (lines 175–187) that
    advertises `markdown, sqlite`; SPEC-014 rewrites that section.
  - `/docs/data-model.md` — entries schema (the 9 field names DEC-011
    locks); gains a DEC-011 cross-reference.
  - `/docs/tutorial.md` — §4 "Read them back" (gets `--format json|tsv`
    subsection); §9 "What's NOT there yet" table (strike
    `brag export --format sqlite`; `brag export --format markdown`
    stays — arrives in SPEC-015); §2/line 3 Scope blurb (rewrite
    "`export` and `summary` arrive in later stages" now that
    export is partially shipping).
  - `/guidance/constraints.yaml` — full constraint list.
  - `/decisions/DEC-004-tags-comma-joined-for-mvp.md` — directly
    constrains JSON: tags render as comma-joined string, NOT array.
  - `/decisions/DEC-006-cobra-cli-framework.md`
  - `/decisions/DEC-007-required-flag-validation-in-runE.md` — applies
    to `--format` validation on both commands.
  - `/decisions/DEC-011-json-output-shape.md` — emitted by THIS spec;
    the six-choice lock.
  - `/projects/PROJ-001-mvp/specs/done/SPEC-004-brag-list-command.md`
    — original 3-column list output that plain mode preserves.
  - `/projects/PROJ-001-mvp/specs/done/SPEC-007-list-filter-flags.md`
    — `ListFilter` struct; test patterns for `--tag`/`--project`/
    `--type`/`--since`/`--limit`.
  - `/projects/PROJ-001-mvp/specs/done/SPEC-013-brag-list-show-project-column.md`
    — "plain mode byte-stable under new-flag addition" is the
    template for this spec's "plain mode byte-stable under
    `--format` absence" test.
  - `/internal/cli/list.go` — existing command; gains `--format`.
  - `/internal/cli/list_test.go` — existing tests; extends.
  - `/internal/cli/root.go`, `/internal/cli/errors.go` — `ErrUser`
    sentinel + `UserErrorf` helper; `runExport` reuses.
  - `/internal/cli/show.go` — `renderEntry` helper (NOT lifted in
    SPEC-014; lift is SPEC-015's drive-by).
  - `/internal/storage/entry.go` — `Entry` struct field names that
    the JSON keys must match: `ID/Title/Description/Tags/Project/
    Type/Impact/CreatedAt/UpdatedAt`.
  - `/cmd/brag/main.go` — gains one `root.AddCommand(cli.NewExportCmd())`
    line.
  - `/README.md` — line 56–57 Scope blurb + line 105 `internal/`
    purpose list entry; both mention `export`.
- **External APIs:** none. stdlib `encoding/json` only.
- **Related code paths:** `internal/cli/`, new `internal/export/`,
  `docs/`, `README.md`, `cmd/brag/main.go`, `decisions/`.

## Outputs

- **Files created:**
  - `/decisions/DEC-011-json-output-shape.md` — emitted in design
    alongside this spec. Six locked choices with honest confidence
    (0.85); alternatives (envelope, tags-as-array, per-format,
    omitempty, compact-default); consequences; revisit criteria;
    cross-refs to SPEC-014/017 and DEC-004/006/007.
  - `/internal/export/json.go` — new package. Exports
    `func ToJSON(entries []storage.Entry) ([]byte, error)` that
    marshals the slice per DEC-011: ordered struct-backed per-entry
    record with 9 fields (id, title, description, tags, project,
    type, impact, created_at, updated_at), timestamps as RFC3339
    strings, `json.MarshalIndent(v, "", "  ")` for 2-space
    indentation, empty array returns `[]`. Consumed by `list.go` and
    `export.go`.
  - `/internal/export/json_test.go` — package-local tests against
    `ToJSON` using a fixed `[]storage.Entry` fixture.
  - `/internal/cli/export.go` — new file. Exports
    `func NewExportCmd() *cobra.Command` plus unexported `runExport`.
    Declares `--format` (required, RunE-validated per DEC-007),
    `--out` (optional path; overwrite existing; directory/unwritable
    returns internal error), filter flags identical to `brag list`
    (`--tag`/`--project`/`--type`/`--since`/`--limit`). Builds
    `storage.ListFilter` the same way `runList` does, calls
    `Store.List(filter)`, marshals via `export.ToJSON`, writes to
    stdout or `--out` path.
  - `/internal/cli/export_test.go` — new file. Eight new tests (see
    Failing Tests) including the load-bearing shape-consistency
    assertion between `list --format json` and `export --format json`.
- **Files modified:**
  - `/internal/cli/list.go` — add `--format` string flag (default
    `""`; RunE-validated; accepted values: `json`, `tsv`). In
    `runList`, after computing entries: if `--format json`, write
    `export.ToJSON(entries)` to `cmd.OutOrStdout()`; if
    `--format tsv`, write header row + per-entry TSV lines; if
    `--format` absent, preserve existing plain / `-P` branches
    byte-for-byte. Long / Examples block gains two new lines
    (`brag list --format json`, `brag list --format tsv`).
  - `/internal/cli/list_test.go` — append seven new tests (one of
    them is a plain-mode byte-stability regression lock relying on
    SPEC-013's existing `TestListCmd_PlainOutputByteIdenticalToSTAGE002`
    remaining green; SPEC-014 adds the `--format`-absent explicit
    test separately to pair directly with the "plain stays stable
    under --format absent" decision).
  - `/cmd/brag/main.go` — one added line:
    `root.AddCommand(cli.NewExportCmd())` alongside the six
    existing `AddCommand` calls.
  - `/docs/api-contract.md` — (a) `brag list` section: synopsis
    gains `[--format json|tsv]`; new sub-bullet describing `--format
    json` → DEC-011 shape and `--format tsv` → header + same field
    order; cross-link `/decisions/DEC-011-json-output-shape.md`.
    (b) `brag export` section (lines 175–187) REWRITTEN to describe
    only what ships in SPEC-014: `--format json` required,
    `--format markdown` mentioned as "arrives in SPEC-015" one-liner,
    filter-flag reuse, `--out` semantics. Stale `sqlite` mention
    removed (deferred 2026-04-23).
  - `/docs/data-model.md` — new reference bullet at end pointing
    `DEC-011` → JSON output shape; no schema change.
  - `/docs/tutorial.md` — (a) §4 "Read them back" gains a new
    `### Machine-readable output: --format json|tsv` subsection
    between "Filter flags" and "See project at scan time" (or
    after the latter — either order ok; the spec suggests after
    `-P` so plain-mode discussion stays contiguous). Keeps
    examples short; does not document DEC-011's six choices in
    prose (the DEC is the authoritative doc; tutorial links).
    (b) §9 "What's NOT there yet" table: strike
    `brag export --format sqlite` (deferred to backlog). Keep
    `brag export --format markdown` (arrives in SPEC-015). Keep
    `brag summary`. (c) §2 Scope blurb line 3: "`export` and
    `summary` arrive in later stages" → "`brag export --format
    markdown` and `brag summary` arrive in later stages" (or
    similar; key change is `export` is no longer uniformly
    unavailable).
  - `/README.md` — line 56–57 Scope blurb: mention that
    `brag list --format json|tsv` and `brag export --format json`
    ship in STAGE-003; retain "summary arrives in STAGE-003" /
    polish-stage forward references as still-accurate. Line 105
    `internal/` row already mentions `export` as a future package;
    update to past tense or drop the "(STAGE-003)" qualifier in
    AGENTS.md §5 directory structure if needed (optional — not
    load-bearing).
- **New exports:**
  - `export.ToJSON(entries []storage.Entry) ([]byte, error)` (new
    package `internal/export`).
  - `cli.NewExportCmd() *cobra.Command` (new file
    `internal/cli/export.go`).
- **Database changes:** none. Pure read path; uses existing
  `Store.List(ListFilter)` from SPEC-007. No migration.

## Acceptance Criteria

Every criterion is testable. Paired failing test name in italics;
three are covered by combinations of tests and are noted as such.

- [ ] DEC-011 exists at `/decisions/DEC-011-json-output-shape.md`
      with the six locked choices, rejected alternatives (envelope,
      tags-as-array, per-format, omitempty, compact-default), honest
      confidence (0.85), and references to SPEC-014/017 and
      DEC-004/006/007. *[manual: `ls decisions/DEC-011*` returns the
      file; grep for "0.85" and "naked array" in it.]*
- [ ] `export.ToJSON([]storage.Entry{...})` emits a pretty-printed
      (2-space indent) naked JSON array with 9 keys per entry in SQL-
      column order, tags as comma-joined string, timestamps as RFC3339,
      empty fields as `""`. All six DEC-011 choices verifiable from
      the golden output. *TestToJSON_DEC011ShapeGolden*
- [ ] `export.ToJSON([]storage.Entry{})` (empty input) emits exactly
      `[]\n` (pretty-printed empty array with trailing newline is OK;
      see Implementation Context for the exact expected bytes).
      *TestToJSON_EmptyInputEmitsEmptyArray*
- [ ] `brag list --format json` writes `export.ToJSON(entries)` to
      stdout (byte-identical to the helper's output) after applying
      any filter flags. Errors on unknown format values. Plain-mode
      under `--format` absent remains byte-identical to STAGE-002 +
      SPEC-013 (no regression). *TestListCmd_FormatJSON_EmitsExportJSON*
      + *TestListCmd_FormatJSON_FiltersApply* +
      *TestListCmd_FormatUnknownValueIsUserError* + the existing
      SPEC-013 `TestListCmd_PlainOutputByteIdenticalToSTAGE002` stays
      green unchanged.
- [ ] `brag list --format tsv` emits a header row
      (`id\ttitle\tdescription\ttags\tproject\ttype\timpact\t
      created_at\tupdated_at\n`) as the first line, followed by
      entry rows in the same 9-field order; each entry row has
      exactly 8 tab characters and 9 fields; empty fields render as
      the empty string between tabs.
      *TestListCmd_FormatTSV_HeaderAndDataShape*
- [ ] `brag list --help` output includes the string `--format` and
      the accepted values `json` and `tsv` (via the flag's usage
      string). *TestListCmd_FormatHelpShowsAcceptedValues*
- [ ] `brag export` (no `--format`) exits 1 (user error) with a
      message naming the flag and the accepted value(s). DEC-007
      applies — no `MarkFlagRequired`; `UserErrorf` in `RunE`.
      *TestExportCmd_FormatRequiredIsUserError*
- [ ] `brag export --format yaml` (unknown value) exits 1 (user error)
      with a message naming the unknown value and the accepted list.
      *TestExportCmd_FormatUnknownValueIsUserError*
- [ ] `brag export --format json` writes `export.ToJSON(entries)` to
      stdout. Shape-consistency with `brag list --format json` is
      locked by byte-identical comparison on the same fixture (the
      load-bearing test). *TestExportCmd_FormatJSON_StdoutEmitsDEC011Shape*
      + *TestExportCmd_FormatJSON_ByteIdenticalToListJSON*
- [ ] `brag export --format json --out <path>` writes the same bytes
      to `<path>` (not stdout). Existing file at `<path>` is
      overwritten without prompt (per stage Design Notes). Stdout is
      empty in this mode.
      *TestExportCmd_FormatJSON_OutPathWritesFile*
- [ ] `brag export --format json --tag X --project Y --type Z
      --since D --limit N` applies the filters (same `ListFilter`
      fields as `brag list`); filtered-out rows do not appear in
      output; rendering is identical for the rows that remain.
      *TestExportCmd_FormatJSON_FiltersApply*
- [ ] `brag export --help` shows `--format` with its accepted
      value(s); help text references the `json` shape pointer to
      DEC-011 (or at minimum includes the `--format` flag). Command
      is registered on `cmd/brag/main.go` root (manually verified by
      running `brag --help` / `brag export --help`).
      *TestExportCmd_HelpShowsFormat* + *[manual: `go build
      ./cmd/brag && ./brag export --help` shows the subcommand.]*
- [ ] `gofmt -l .` empty; `go vet ./...` clean; `CGO_ENABLED=0 go
      build ./...` succeeds; `go test ./...` and `just test` green.
- [ ] Doc sweep: `docs/api-contract.md`, `docs/tutorial.md`,
      `docs/data-model.md`, and `README.md` all updated per Outputs.
      The stale `--format sqlite` mention is removed from both
      `docs/api-contract.md` and `docs/tutorial.md` §9. *[manual
      greps listed under Premise audit below.]*

## Locked design decisions

Reproduced here so build / verify don't re-litigate. Each is paired
with at least one failing test below per AGENTS.md §9 (SPEC-009 ship
lesson).

1. **DEC-011 six choices (emitted in `/decisions/DEC-011-*`):** (1)
   naked array, (2) 9 keys in SQL-column order matching
   `entries` exactly, (3) tags as comma-joined string per DEC-004,
   (4) RFC3339 timestamp strings, (5) empty fields as `""`,
   (6) pretty-print indent=2. All six verifiable from one golden
   fixture. *Pair: `TestToJSON_DEC011ShapeGolden`.*
2. **`internal/export` package with `ToJSON` helper is the single
   shape source.** Both `brag list --format json` and `brag export
   --format json` route through it. Shape-consistency is locked by
   byte-identical comparison on the same fixture. *Pair:
   `TestExportCmd_FormatJSON_ByteIdenticalToListJSON` (the
   load-bearing cross-path assertion).*
3. **`brag list` plain mode is byte-stable under `--format` absent.**
   Adding the `--format` flag must not change any byte of plain
   `brag list` or `brag list -P` output. *Pair: existing
   SPEC-013 `TestListCmd_PlainOutputByteIdenticalToSTAGE002` stays
   green unchanged, plus the explicit new
   `TestListCmd_FormatJSON_EmitsExportJSON` asserts `--format json`
   routes through the helper without corrupting the default branch.*
4. **TSV has a header row (deliberate divergence from plain list).**
   Stage Design Notes lock this. Header row is the first line of
   output, field names in the same order as JSON keys:
   `id\ttitle\tdescription\ttags\tproject\ttype\timpact\t
   created_at\tupdated_at\n`. *Pair:
   `TestListCmd_FormatTSV_HeaderAndDataShape`.*
5. **`brag export --format` is required.** Missing flag → `UserErrorf`
   (DEC-007 pattern); never `MarkFlagRequired`. Help text lists the
   accepted values. Shipped values in SPEC-014: just `json`.
   SPEC-015 extends with `markdown`. *Pair:
   `TestExportCmd_FormatRequiredIsUserError`.*
6. **Unknown `--format` value is a user error.** Both `brag list` and
   `brag export` return `UserErrorf` naming the accepted values on
   an unrecognized value. *Pair:
   `TestListCmd_FormatUnknownValueIsUserError` (list) +
   `TestExportCmd_FormatUnknownValueIsUserError` (export).*
7. **Filter flags reuse `ListFilter` verbatim on `brag export`.**
   No new filter logic; same flag names and semantics as SPEC-007.
   *Pair: `TestExportCmd_FormatJSON_FiltersApply`.*
8. **`brag export --out <path>` writes to file and overwrites without
   prompt.** Stage Design Notes lock overwrite-no-prompt semantics
   (match `jq -o` / `goreleaser` conventions). Stdout stays empty
   in this mode. *Pair: `TestExportCmd_FormatJSON_OutPathWritesFile`.*
9. **api-contract.md `brag export` section REWRITTEN, not patched.**
   Per Q5 during design: SPEC-014 is the first spec to actually ship
   the `brag export` command; the existing stub advertising
   `markdown, sqlite` is stale. Rewrite to describe only what ships
   in SPEC-014 (`--format json`) with a one-sentence forward
   reference to SPEC-015's `markdown`. Stale `sqlite` mention
   removed entirely. *Pair: manual grep under Premise audit.*

**Out of scope (by design — backlog entries exist for each):**

- `--envelope` / `--wrap` flag on `--format json`.
- `--compact` / non-pretty `--format json`.
- Tags as a JSON array (would require DEC-004 migration).
- Per-field include/exclude flags on JSON output.
- NDJSON output (line-delimited JSON).
- `brag export --format sqlite` (deferred to backlog 2026-04-23; the
  stale api-contract.md mention is removed as part of this spec's
  doc sweep).
- `brag export --format markdown` (arrives in SPEC-015; forward-
  referenced in api-contract.md / tutorial.md only).
- `-P` composition with `--format`. `--format json|tsv` always
  includes the `project` field (it's in the 9-key shape), so `-P`
  is structurally redundant under `--format` and is silently
  allowed (no error, no effect). Not tested; documented under Notes
  for the Implementer.

## Premise audit (AGENTS.md §9 — addition + status-change)

SPEC-014 is an **addition** case (new command, new flag, new package,
new DEC) with **status-change** flavor (existing doc stubs advertise
deferred behavior that SPEC-014 supersedes). Both AGENTS.md §9
heuristics apply:

**Addition heuristics** (SPEC-011 ship lesson — grep tracked
collections for count coupling):

- Root command list: `cmd/brag/main.go` has six `AddCommand` calls;
  SPEC-014 makes it seven. Run `grep -rn 'AddCommand\|Commands()'
  internal/cli/*.go cmd/ 2>/dev/null` — verified 2026-04-23: no
  test asserts `len(root.Commands()) == 6` or similar. The only
  `root.Commands()` hit is `list_test.go:627`, which iterates to
  find the `list` subcommand by name, not by count. Safe to add.
- DEC collection: `/decisions/DEC-001..DEC-010` currently; SPEC-014
  adds `DEC-011`. No test or doc asserts the count; `guidance/
  constraints.yaml` and `AGENTS.md §15` reference the directory, not
  a count.
- `ListFilter` fields: unchanged. SPEC-014 reuses SPEC-007's
  `ListFilter` verbatim.

**Status-change heuristics** (SPEC-012 ship lesson — grep feature
name across docs):

Explicit grep commands for the build session to run, with expected
doc-level actions in parens:

```
grep -rn 'brag export' docs/ README.md AGENTS.md
  # → docs/api-contract.md lines 175–187 (REWRITE: only --format json
  #   now; strike sqlite; forward-ref markdown to SPEC-015)
  # → docs/tutorial.md §9 "What's NOT there yet" table
  #   (STRIKE `brag export --format sqlite`; KEEP `--format markdown`)
  # → docs/tutorial.md §2 Scope blurb line 3 (REWRITE: export
  #   partially ships now)
  # → README.md line 56–57 Scope blurb (UPDATE mention)
  # → AGENTS.md §5 directory structure mentions `internal/export/
  #   (STAGE-003)` — status is now "active", not "future"; optional
  #   update; not load-bearing for acceptance criteria.

grep -rn 'format sqlite\|--format sqlite' docs/ README.md
  # → docs/api-contract.md line 184 (STRIKE entirely; in the rewritten
  #   section, sqlite is not mentioned)
  # → docs/tutorial.md line ~352 (STRIKE the row from the "What's NOT
  #   there yet" table)
  # Backlog.md keeps its deferral entry intact — that's the one place
  # sqlite continues to be documented.

grep -rn 'brag list' docs/ README.md
  # → docs/api-contract.md §brag list lines 78–96 (SYNOPSIS: add
  #   `[--format json|tsv]`; OUTPUT: add sub-bullets for json/tsv
  #   shapes; keep default/-P bullets)
  # → docs/tutorial.md §4 lines 123–186 (ADD new `### Machine-readable
  #   output: --format json|tsv` subsection; keep existing content)
  # → README.md line 52 (current text: "list (with filters...; add -P
  #   to include the project)" — add mention of `--format json|tsv`
  #   inline or adjacent)

grep -rn '\bexport\b' docs/data-model.md
  # → add a reference bullet at the end for DEC-011 (JSON output
  #   shape); no schema change.
```

**Existing test audit** (addition-case doesn't add tracked-count
coupling; no existing tests break, but verify):

- `internal/cli/list_test.go` `TestListCmd_HelpShowsFilters` asserts
  the help contains `--tag`, `--project`, `--type`, `--since`,
  `--limit`, `Examples:`. It's a lower-bound substring check, so
  adding `--format` won't break it. No modification needed.
- `TestListCmd_TabSeparatedFormat` asserts exactly 2 tabs on the
  default-mode output line. Default mode is preserved under
  `--format` absent, so this stays green.
- `TestListCmd_PlainOutputByteIdenticalToSTAGE002` (SPEC-013) is
  the strongest plain-mode regression lock. Stays green unchanged.
- `cmd/brag/main.go` has no tests asserting the command list
  length or names.

**Symmetric action from `## Outputs`:** every grep hit above maps to
a concrete file modification in Outputs. No discoveries expected at
build time.

## Failing Tests

Written now, during **design**. Fourteen tests total. All follow
AGENTS.md §9: separate `outBuf` / `errBuf` with no-cross-leakage
asserts; fail-first run before implementation; assertion-specificity
on help substrings; every locked decision paired with at least one
failing test.

### `internal/export/json_test.go` (new file — 2 tests)

Tests against a fixed `[]storage.Entry` fixture. No cobra, no DB.
Pure stdlib + the `internal/export` package.

- **`TestToJSON_DEC011ShapeGolden`** — construct a two-entry fixture:
  - Entry 1: fully populated. `ID: 1`, `Title: "shipped FTS5"`,
    `Description: "migration 0002 landed"`, `Tags: "sqlite,fts5"`,
    `Project: "bragfile"`, `Type: "shipped"`,
    `Impact: "unblocked search"`, `CreatedAt: time.Date(2026, 4,
    22, 6, 30, 0, 0, time.UTC)`, `UpdatedAt` same.
  - Entry 2: minimal. `ID: 2`, `Title: "note"`, everything else
    zero-value (empty strings, zero time for `CreatedAt/UpdatedAt`
    — render as `"0001-01-01T00:00:00Z"` which is acceptable for
    this test's purpose because the cli layer always populates real
    RFC3339 timestamps; this test is about the helper's shape, not
    realism).

  Expected output (golden literal, byte-exact):

  ```json
  [
    {
      "id": 1,
      "title": "shipped FTS5",
      "description": "migration 0002 landed",
      "tags": "sqlite,fts5",
      "project": "bragfile",
      "type": "shipped",
      "impact": "unblocked search",
      "created_at": "2026-04-22T06:30:00Z",
      "updated_at": "2026-04-22T06:30:00Z"
    },
    {
      "id": 2,
      "title": "note",
      "description": "",
      "tags": "",
      "project": "",
      "type": "",
      "impact": "",
      "created_at": "0001-01-01T00:00:00Z",
      "updated_at": "0001-01-01T00:00:00Z"
    }
  ]
  ```

  (No trailing newline from `MarshalIndent`; the cli layer adds one
  when writing to stdout — see `TestListCmd_FormatJSON_EmitsExportJSON`
  for the trailing-newline contract.)

  Assertions — all six DEC-011 choices must be verifiable from this
  golden:
  1. Top level is a JSON array (first byte `[`, last byte `]`).
  2. Each object has exactly 9 keys in the order `id, title,
     description, tags, project, type, impact, created_at,
     updated_at` (`json.MarshalIndent` preserves struct-tag order;
     pair with a secondary parse-via-`json.Decoder` check that
     iterates in declaration order).
  3. `tags` is a JSON string (`"sqlite,fts5"`), not an array.
  4. `created_at` / `updated_at` parse as RFC3339.
  5. Entry 2's empty fields serialize as `""`, not omitted.
  6. Indentation is 2 spaces; lines begin with `    "<key>":` for
     object fields (4-space indent: 2 for the array, 2 for the
     object).

  Implementation note: the test body does one `bytes.Equal(got,
  []byte(wantGolden))` against the literal string above. If that
  passes, all six choices are locked.

- **`TestToJSON_EmptyInputEmitsEmptyArray`** — `ToJSON(nil)` and
  `ToJSON([]storage.Entry{})` both return exactly `[]` (2 bytes, no
  trailing newline — matching `MarshalIndent`'s behavior on an empty
  slice). Error is nil. Catches accidental `null` (which is what
  `json.Marshal` emits on a nil slice without care) — DEC-011 says
  naked array, so empty must be `[]`, not `null`.

### `internal/cli/list_test.go` (5 new tests appended)

Reuse existing `newListTestRoot(t)`, `seedListEntry(t, ...)`,
`runListCmd(t, ...)` helpers.

- **`TestListCmd_FormatJSON_EmitsExportJSON`** — seed two entries
  via `seedListEntry`. Run `brag list --format json`. Expected
  output: the bytes returned by
  `export.ToJSON([]storage.Entry{e2, e1})` (reverse-insertion order,
  matching Store.List's `created_at DESC, id DESC` ordering), with
  a trailing newline appended by the cli writer. Assert
  `stdout == string(expectedBytes) + "\n"` byte-exact;
  `stderr == ""`; `err == nil`. This test pairs decision 2 (`internal/
  export` is the shape source) and decision 3 (plain mode byte-stable
  under `--format` absent — the test's fail-first run asserts that
  without `--format`, output stays 3-column plain).

  Note on trailing-newline contract: see Notes for the Implementer
  §"JSON trailing newline" — the cli-layer writes
  `fmt.Fprintln(out, string(b))` so the output has exactly one
  trailing newline. The fixture-level helper (`ToJSON`) returns
  newline-free bytes (matching `MarshalIndent` behavior). This test
  is the single authoritative pairing of those two facts.

- **`TestListCmd_FormatJSON_FiltersApply`** — seed three entries
  varying project: one `platform`, one `growth`, one `""`. Run
  `brag list --format json --project platform`. Parse stdout as
  JSON; assert the array has exactly one element; assert that
  element's `project` field equals `"platform"` and its `title`
  matches the seeded title. Assert `stderr == ""`. Pairs decision
  7 on the list side (filter reuse) — the list command's filter
  plumbing is SPEC-007's; this test proves `--format json` doesn't
  bypass it.

- **`TestListCmd_FormatTSV_HeaderAndDataShape`** — seed two entries:
  - Entry A: `Title: "full"`, `Description: "desc-full"`, `Tags:
    "t1,t2"`, `Project: "platform"`, `Type: "shipped"`, `Impact:
    "imp-full"`.
  - Entry B: minimal, `Title: "bare"`, all other fields empty.

  Run `brag list --format tsv`. Assert:
  1. Output splits into exactly 3 lines (1 header + 2 data rows),
     each ending with `\n`.
  2. Line 0 (header) is **byte-exact** equal to the literal:
     `id\ttitle\tdescription\ttags\tproject\ttype\timpact\t
     created_at\tupdated_at` (9 fields, 8 tabs, no newline in the
     comparison string since we split on `\n`).
  3. Each of lines 1 and 2 has exactly 8 tab characters
     (`strings.Count(line, "\t") == 8`) and 9 fields after
     `strings.Split(line, "\t")`.
  4. For line 1 (entry B, rendered first due to reverse-insertion
     order): fields `[1]..[6]` are `"bare", "", "", "", "", ""`
     respectively; `fields[3]` (tags) is exactly `""` (empty string
     between tabs); `fields[4]` (project) is exactly `""` (NOT `-`
     — decision 5 paired with this assertion: TSV empty ≠ plain's
     dash-filler behavior under `-P`).
  5. For line 2 (entry A): `fields[1..6]` = `"full", "desc-full",
     "t1,t2", "platform", "shipped", "imp-full"`.
  6. `stderr == ""`.

  This test pairs decision 4 (TSV header row + field order) and
  covers the per-row shape in one go.

- **`TestListCmd_FormatUnknownValueIsUserError`** — run
  `brag list --format yaml`. Assert `err != nil`;
  `errors.Is(err, ErrUser)`; `stdout == ""`; the error message
  (`err.Error()`) contains both `yaml` (the offending value) and at
  least one of the accepted values (`json` or `tsv`). Pairs
  decision 6 on the list side.

- **`TestListCmd_FormatHelpShowsAcceptedValues`** — run `list --help`
  with separate buffers. Assert `outBuf.String()` contains
  `--format` AND both `json` AND `tsv` (the latter two via the
  flag's one-line usage string, per AGENTS.md §9 assertion-
  specificity — not generic words). Assert `errBuf.Len() == 0`.

### `internal/cli/export_test.go` (new file — 7 tests)

New file; establishes export-command test harness parallel to
`list_test.go`. Uses `t.TempDir()` for DB paths, seeds entries via a
local `seedExportEntry` or reuses `seedListEntry` if the helper
accessibility allows (same package `cli`, so it does — reuse directly).

- **`TestExportCmd_FormatRequiredIsUserError`** — run `brag export`
  against a DB with seeded rows, no `--format`. Assert `err != nil`;
  `errors.Is(err, ErrUser)`; `stdout == ""`; error message mentions
  `--format` and `json`. Pairs decision 5.

- **`TestExportCmd_FormatUnknownValueIsUserError`** — run
  `brag export --format yaml`. Same assertions as the list version
  (UserErrorf, stdout empty, message names `yaml` and `json`). Pairs
  decision 6 on the export side.

- **`TestExportCmd_FormatJSON_StdoutEmitsDEC011Shape`** — seed two
  entries. Run `brag export --format json`. Expected: bytes equal to
  `export.ToJSON(entries)` + trailing newline. Parse stdout as JSON;
  assert array length is 2; assert first element's 9 keys match
  DEC-011 order. Assert `stderr == ""`. Primary happy-path for
  export JSON.

- **`TestExportCmd_FormatJSON_OutPathWritesFile`** — (a) Seed two
  entries. Create a file at `filepath.Join(t.TempDir(), "export.json")`
  pre-filled with sentinel bytes (`"PRE-EXISTING CONTENT\n"`) to
  verify overwrite semantics. (b) Run `brag export --format json
  --out <path>`. (c) Assert `stdout == ""` (nothing on stdout when
  `--out` is set); `stderr == ""`; `err == nil`. (d) Read the file;
  assert its contents equal `string(export.ToJSON(entries)) + "\n"`
  (cli adds trailing newline when writing to file too — consistent
  with stdout behavior). (e) Assert sentinel is gone (file truly
  overwritten). Pairs decision 8.

- **`TestExportCmd_FormatJSON_FiltersApply`** — seed three entries
  across projects: `platform`, `growth`, `""`. Run `brag export
  --format json --project platform`. Parse stdout as JSON; assert
  length 1; assert the one entry's `project` is `"platform"`. Pairs
  decision 7 on the export side.

- **`TestExportCmd_FormatJSON_ByteIdenticalToListJSON`** — THE
  load-bearing cross-path assertion. Seed three entries (varied
  fields to exercise empty-string and non-empty cases). Run
  `brag list --format json` → capture `stdoutList`. Run
  `brag export --format json` → capture `stdoutExport`. Assert
  `stdoutList == stdoutExport` byte-for-byte. Assert both stderrs
  empty. If this fails, DEC-011 or the `internal/export.ToJSON`
  routing has drifted. Pairs decisions 1 (DEC-011 shape) and 2
  (single-source-of-truth helper).

- **`TestExportCmd_HelpShowsFormat`** — construct a root command
  with `NewExportCmd` attached; run `export --help` with separate
  buffers. Assert `outBuf.String()` contains `--format` AND `json`
  (accepted value via flag's one-line usage string). Assert
  `errBuf.Len() == 0`. Covers decision 5 on the help side. Also
  indirectly verifies `NewExportCmd` returns a usable `*cobra.Command`
  (registration on `cmd/brag/main.go` root is a manual verification
  in acceptance criteria).

### Test count summary

14 failing tests across 3 files:

- `internal/export/json_test.go` — 2 tests.
- `internal/cli/list_test.go` — 5 tests (appended).
- `internal/cli/export_test.go` — 7 tests (new file).

Plus one existing test's continued green (`SPEC-013`'s
`TestListCmd_PlainOutputByteIdenticalToSTAGE002`) stands as the
plain-mode regression lock.

## Implementation Context

*Read this section and the files it points to before starting the
build cycle. It is the equivalent of a handoff document, folded into
the spec since there is no separate receiving agent.*

### Decisions that apply

- **`DEC-011`** (emitted in this spec) — six locked JSON shape
  choices. `internal/export/json.go` implements them. Every
  deviation from the golden in
  `TestToJSON_DEC011ShapeGolden` is a DEC-011 violation — stop and
  raise a question before "fixing" the test.
- **`DEC-004`** — tags are comma-joined TEXT in storage and remain
  comma-joined string in JSON. Do NOT split on comma at the I/O
  boundary. `storage.Entry.Tags` goes through verbatim.
- **`DEC-006`** — cobra framework. New `brag export` subcommand
  follows the same pattern as every other command: `NewExportCmd()
  *cobra.Command` constructor in `internal/cli/export.go`;
  unexported `runExport` RunE handler; flags declared on the command
  in the constructor.
- **`DEC-007`** — required-flag validation in `RunE`. `--format` is
  required on `brag export`; missing/empty/unknown value → return
  `cli.UserErrorf(...)`. Do NOT use `MarkFlagRequired`. On
  `brag list`, `--format` is NOT required (default = plain), but
  unknown values still go through `UserErrorf`.

### Constraints that apply

For `internal/cli/**`, `internal/export/**`, `docs/**`,
`README.md`, `cmd/brag/main.go`, `decisions/**`:

- `no-sql-in-cli-layer` — blocking. `export.go` and the new
  `internal/export` package must not import `database/sql` or any
  SQL driver. They work with `[]storage.Entry` returned by
  `Store.List`.
- `stdout-is-for-data-stderr-is-for-humans` — blocking. JSON/TSV
  bodies and the `brag export --out` silent-success path go through
  stdout (nothing when `--out` set). Human-readable errors go
  through `main.go`'s `fmt.Fprintf(os.Stderr, "brag: %s\n", ...)`
  wrapper. Every happy-path test asserts `errBuf.Len() == 0`.
- `errors-wrap-with-context` — warning. Any storage/io errors from
  `export.go` wrap with `fmt.Errorf("...: %w", err)`.
- `test-before-implementation` — blocking. Write all 14 tests
  first, run `go test ./internal/export ./internal/cli` (both
  packages), confirm every new test fails for the expected reason
  (missing symbol, missing flag, missing command — NOT a compilation
  error unrelated to the spec), THEN implement.
- `one-spec-per-pr` — blocking. Branch
  `feat/spec-014-json-trio-and-shape-dec`. Diff touches only the
  files in Outputs.
- `no-new-top-level-deps-without-decision` — not triggered;
  `encoding/json` is stdlib.

### AGENTS.md lessons that apply

- **§9 separate `outBuf` / `errBuf`** (SPEC-001) — every cli test
  in both files asserts both buffers.
- **§9 fail-first** (SPEC-003) — confirm each of the 14 tests fails
  for the expected reason before implementation.
- **§9 assertion specificity** (SPEC-005) — help tests assert on
  distinctive needles (`--format`, `json`, `tsv`), not generic
  words.
- **§9 locked-decisions-need-tests** (SPEC-009) — nine locked
  decisions above; each paired with at least one of the 14 tests
  (plus one pairing with a manual grep for the api-contract.md
  rewrite).
- **§9 premise audit — addition case** (SPEC-011) — grepped above;
  no tracked-collection count coupling; safe to add.
- **§9 premise audit — status-change case** (SPEC-012) — grepped
  above; doc-level actions enumerated under Outputs, not discovered
  at build time. The stale `--format sqlite` mention is the
  status-change hotspot.

### Prior related work

- **SPEC-004** (shipped 2026-04-20) — original `brag list` 3-column
  contract. SPEC-014 preserves it byte-for-byte under `--format`
  absent.
- **SPEC-007** (shipped 2026-04-20) — `ListFilter` struct with
  `Tag/Project/Type/Since/Limit`. `runExport` populates `ListFilter`
  the same way `runList` does; copy the `cmd.Flags().Changed(...)`
  + `ParseSince` dispatch block verbatim.
- **SPEC-013** (shipped 2026-04-23) — `brag list -P`. Its
  `TestListCmd_PlainOutputByteIdenticalToSTAGE002` is the
  regression lock for plain mode; SPEC-014's addition must not
  disturb it. `-P` remains functional under `--format` absent and
  is silently no-op under `--format json|tsv` (the 9-key shape
  already includes `project`).
- **SPEC-011 / SPEC-012** (shipped 2026-04-22) — FTS5 + `brag
  search`. Unrelated to SPEC-014 directly; only noted to confirm
  that `brag search` does NOT get `--format` in this spec (a
  separate spec if it ever does).

### Out of scope (for this spec specifically)

If any of these feels necessary during build, write a new spec
rather than expanding this one.

- **`--envelope` / `--wrap` flag** on JSON output. Backlog: "JSON
  output envelope".
- **`--compact` / non-pretty JSON**. Backlog: "`--compact` /
  non-pretty JSON output".
- **Tags as JSON array**. Would require DEC-004 migration. Backlog:
  the rejected alternative lives inside DEC-011 itself.
- **Per-field include/exclude flags**. Not in any backlog entry;
  no ask.
- **NDJSON output**. Backlog: "NDJSON / array-batch stdin for
  `brag add --json`" (the input side); the output side has no
  backlog entry — raise one if a need emerges.
- **`brag export --format sqlite`**. Deferred 2026-04-23. Backlog:
  "`brag export --format sqlite` (full-DB VACUUM INTO)" + "Filtered
  SQLite export".
- **`brag export --format markdown`**. Arrives in SPEC-015. SPEC-014
  only forward-references it in api-contract.md / tutorial.md.
- **Lifting `renderEntry` into `internal/export`**. SPEC-015's
  drive-by. SPEC-014 creates the package but does NOT lift
  `renderEntry`; keep it in `internal/cli/show.go` for now.
- **`brag add --json`**. SPEC-017. Reads DEC-011 verbatim for its
  stdin shape.
- **`-P` behavior under `--format`**. Documented as silently no-op;
  not tested. If a user reports confusion, raise a follow-up spec.
- **Changes to `storage.Entry`, `storage.Store`, or `storage.ListFilter`**.
  This spec is pure CLI + new helper package; no storage-layer
  modifications.
- **Changes to existing tests in `list_test.go`** (beyond appending
  new ones). No modifications; the addition must compose
  non-destructively.

## Notes for the Implementer

Gotchas, style preferences, reuse opportunities. Read after the
Implementation Context; these are the "how" details.

- **`internal/export/json.go` layout.** Small file — a single
  package-level `type entryRecord struct { … }` with `json` struct
  tags in the locked order, a `toRecord(e storage.Entry) entryRecord`
  mapper, and `func ToJSON(entries []storage.Entry) ([]byte, error)`
  that builds a `[]entryRecord` and returns
  `json.MarshalIndent(records, "", "  ")`. Returning
  `json.MarshalIndent`'s output directly (no custom trailing newline
  handling) is what the helper-level test asserts. An explicit
  non-nil empty slice (`records := []entryRecord{}` or `if len ==
  0 { return []byte("[]"), nil }`) is required to avoid `null`
  output on nil input — `TestToJSON_EmptyInputEmitsEmptyArray`
  locks this.

  Suggested sketch:

  ```go
  package export

  import (
      "encoding/json"
      "time"

      "github.com/jysf/bragfile000/internal/storage"
  )

  type entryRecord struct {
      ID          int64  `json:"id"`
      Title       string `json:"title"`
      Description string `json:"description"`
      Tags        string `json:"tags"`
      Project     string `json:"project"`
      Type        string `json:"type"`
      Impact      string `json:"impact"`
      CreatedAt   string `json:"created_at"`
      UpdatedAt   string `json:"updated_at"`
  }

  func ToJSON(entries []storage.Entry) ([]byte, error) {
      records := make([]entryRecord, 0, len(entries))
      for _, e := range entries {
          records = append(records, entryRecord{
              ID:          e.ID,
              Title:       e.Title,
              Description: e.Description,
              Tags:        e.Tags,
              Project:     e.Project,
              Type:        e.Type,
              Impact:      e.Impact,
              CreatedAt:   e.CreatedAt.UTC().Format(time.RFC3339),
              UpdatedAt:   e.UpdatedAt.UTC().Format(time.RFC3339),
          })
      }
      return json.MarshalIndent(records, "", "  ")
  }
  ```

  `make([]entryRecord, 0, 0)` is the nil-safe empty slice: it
  marshals as `[]`, not `null`. Verified by the second test.

- **JSON trailing newline.** Important and easy to get wrong. The
  helper returns `MarshalIndent`'s raw bytes (no trailing newline).
  The cli layer writes them and appends exactly one newline:
  `fmt.Fprintln(out, string(b))`. This means stdout has exactly one
  trailing newline (matches every other CLI output in `brag`).
  The file write via `--out` uses the same
  `string(b) + "\n"` pattern — do NOT omit the newline just because
  a file viewer shows it trailing-newline-free; Unix tooling
  (`wc -l`, `cat`, `tail`) all expect trailing newlines. Both
  `TestListCmd_FormatJSON_EmitsExportJSON` and
  `TestExportCmd_FormatJSON_OutPathWritesFile` assert the trailing
  newline explicitly.

- **`internal/cli/list.go` `--format` flag declaration.** Add after
  the existing six flags, before the `return cmd`:

  ```go
  cmd.Flags().String("format", "", "output format (one of: json, tsv); default is plain tab-separated")
  ```

  The usage string substring `"one of: json, tsv"` is load-bearing:
  `TestListCmd_FormatHelpShowsAcceptedValues` asserts on the
  needles `"json"` and `"tsv"` inside help. Pick any phrasing but
  make sure both literals appear. If you change the exact wording,
  update this spec first, then code.

- **`runList` dispatch for `--format`.** After `entries, err :=
  s.List(filter)` and before the current plain/`-P` branching, read
  the flag and branch on its value. Plain/`-P` branches live inside
  the `case ""` branch (empty = no flag = plain) so they remain
  byte-identical to today's behavior.

  Sketch:

  ```go
  format, _ := cmd.Flags().GetString("format")
  showProject, _ := cmd.Flags().GetBool("show-project")
  out := cmd.OutOrStdout()

  switch format {
  case "":
      // existing plain / -P loop, unchanged
      for _, e := range entries { ... }
  case "json":
      b, err := export.ToJSON(entries)
      if err != nil {
          return fmt.Errorf("marshal json: %w", err)
      }
      fmt.Fprintln(out, string(b))
  case "tsv":
      fmt.Fprintln(out, tsvHeader)
      for _, e := range entries {
          fmt.Fprintln(out, tsvRow(e))
      }
  default:
      return UserErrorf("unknown --format value %q (accepted: json, tsv)", format)
  }
  ```

  The `tsvHeader` / `tsvRow` helpers can live in `list.go` as
  unexported package-local funcs, OR in `internal/export/tsv.go`
  alongside JSON. Recommendation: keep them in `internal/export`
  for consistency (one package owns output formats), exporting
  `export.TSVHeader` (string const) and `export.ToTSVRow(e
  storage.Entry) string`. This means the tsv-header test can
  import the constant directly, strengthening the test —
  `fields := strings.Split(export.TSVHeader, "\t")` proves the
  constant really does have 9 fields with 8 tabs. But if the build
  agent judges the export package gets too busy, inlining in
  `list.go` is acceptable — either option passes the tests
  verbatim. Pick one and document under Deviations if different
  from this recommendation.

- **TSV row format.** Use `\t` as the separator, not any escaping.
  If a field contains a literal tab, the output breaks column
  alignment — same accepted trade-off as SPEC-004's plain-list
  treatment of tabs in titles and SPEC-013's `project` tab
  handling. Document this as a MVP limitation in any code comment
  if you write one; don't escape.

  Example `ToTSVRow`:

  ```go
  func ToTSVRow(e storage.Entry) string {
      return strings.Join([]string{
          strconv.FormatInt(e.ID, 10),
          e.Title,
          e.Description,
          e.Tags,
          e.Project,
          e.Type,
          e.Impact,
          e.CreatedAt.UTC().Format(time.RFC3339),
          e.UpdatedAt.UTC().Format(time.RFC3339),
      }, "\t")
  }
  ```

  Header constant:

  ```go
  const TSVHeader = "id\ttitle\tdescription\ttags\tproject\ttype\timpact\tcreated_at\tupdated_at"
  ```

- **`internal/cli/export.go` structure.** Mirror `list.go`: a
  `NewExportCmd` constructor declares flags in order (`--format`
  first, then `--out`, then the five filter flags in the same
  order as `list`). `runExport` validates `--format`, builds
  `ListFilter` identically to `runList` (copy the block; don't try
  to share — the shared copy is DRY but the spec says "no helper
  extraction" for SPEC-014 since SPEC-015 will naturally want to
  extract). Opens the store. Calls `Store.List(filter)`. Marshals
  via `export.ToJSON`. Writes to stdout or `--out` path.

  `--out` validation: if the path's parent directory doesn't exist
  or isn't writable, return `fmt.Errorf("write output: %w", err)`
  (NOT `UserErrorf` — disk state is an internal error per the
  stage Design Notes "`--out <path>` semantics" bullet). If the
  path points to an existing file, open with `os.O_WRONLY |
  os.O_CREATE | os.O_TRUNC` to overwrite; no prompt.

- **`ListFilter` population.** Copy the block from `runList`
  (SPEC-007). The five flags: `--tag`, `--project`, `--type`,
  `--since`, `--limit`. Validation is identical — empty strings
  are user errors; invalid `--since` goes through `ParseSince`
  which returns a wrappable error; non-positive `--limit` is a
  user error. The same `cmd.Flags().Changed(...)` pattern
  distinguishes "not set" from "set to empty".

- **`cmd/brag/main.go` update.** One-line addition, alphabetically
  between `NewEditCmd` and `NewListCmd` (or at the end after
  `NewSearchCmd` — author preference; existing order is
  insertion-order, not alphabetical). Suggested location: after
  `NewSearchCmd()` to keep STAGE-003 additions visually grouped
  at the bottom.

  ```go
  root.AddCommand(cli.NewExportCmd())
  ```

- **Doc updates (execute in this order).** Each is enumerated as a
  concrete file modification under Outputs; here's the narrative
  for the two non-obvious ones:

  1. `docs/api-contract.md` `brag export` section — full rewrite.
     Drop the existing stub (lines 175–187) and replace with a
     section that documents only what SPEC-014 ships plus the
     SPEC-015 forward reference. Pattern:

     ```
     ### `brag export` — export entries (STAGE-003)

     ```
     brag export --format json                         # stdout: JSON array
     brag export --format json --out entries.json      # write to file
     brag export --format json --project platform      # filter before exporting
     ```

     - `--format` is required. Accepted values: `json`. `markdown`
       arrives in SPEC-015.
     - `--out <path>` optional; defaults to stdout. Overwrites an
       existing file without prompt.
     - Accepts the same filter flags as `brag list` (`--tag`,
       `--project`, `--type`, `--since`, `--limit`). `ListFilter`
       is shared.
     - JSON shape locked by
       [DEC-011](../decisions/DEC-011-json-output-shape.md).
     ```

  2. `docs/tutorial.md` §4 — add a new subsection. Suggested
     placement: after "See project at scan time" (the `-P` block)
     so plain-mode and single-flag extensions stay grouped, then
     the machine-readable block leads into search/edit/delete.
     Content:

     ```
     ### Machine-readable output: `--format json|tsv`

     ```bash
     brag list --format json                # pretty-printed JSON array
     brag list --format tsv                 # tab-separated with a header row
     brag export --format json --out b.json # durable dump
     ```

     JSON and TSV output share the same 9 fields in the same
     order as the `entries` table: `id, title, description,
     tags, project, type, impact, created_at, updated_at`.
     Tags stay a comma-joined string (per DEC-004);
     timestamps are RFC3339. Pipe into `jq` for ad-hoc queries.

     Shape is locked by
     [DEC-011](../decisions/DEC-011-json-output-shape.md).
     ```

- **fail-first run.** Before implementing, run:

  ```bash
  go test ./internal/export ./internal/cli -run \
      "TestToJSON|TestListCmd_Format|TestExportCmd"
  ```

  Expected: every one of the 14 new tests fails for the expected
  reason (undefined package `internal/export`, undefined
  `NewExportCmd`, unknown flag `--format`). If any passes
  unexpectedly, investigate — a passing-before-implementation test
  is either a too-weak assertion or a pre-existing symbol
  accidentally overlapping (SPEC-003 ship lesson).

- **No helper extraction beyond the `internal/export` package.**
  Specifically: do NOT DRY the `ListFilter` population block
  between `runList` and `runExport` by extracting a helper.
  SPEC-015 will naturally want that helper; duplication is the
  cheaper path today, and the explicit copy reads as "two commands
  with identical filter semantics" at diff review time.

- **Branch:** `feat/spec-014-json-trio-and-shape-dec`.

---

## Build Completion

*Filled in at the end of the **build** cycle, before advancing to verify.*

- **Branch:**
- **PR (if applicable):**
- **All acceptance criteria met?** yes/no
- **New decisions emitted:**
  - `DEC-NNN` — <title> (if any)
- **Deviations from spec:**
  - [list]
- **Follow-up work identified:**
  - [any new specs for the stage's backlog]

### Build-phase reflection (3 questions, short answers)

Process-focused: how did the build go? What friction did the spec create?

1. **What was unclear in the spec that slowed you down?**
   — <answer>

2. **Was there a constraint or decision that should have been listed but wasn't?**
   — <answer>

3. **If you did this task again, what would you do differently?**
   — <answer>

---

## Reflection (Ship)

*Appended during the **ship** cycle. Outcome-focused reflection, distinct
from the process-focused build reflection above.*

1. **What would I do differently next time?**
   — <answer>

2. **Does any template, constraint, or decision need updating?**
   — <answer>

3. **Is there a follow-up spec I should write now before I forget?**
   — <answer>
