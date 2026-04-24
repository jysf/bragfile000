---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-015
  type: story                      # epic | story | task | bug | chore
  cycle: verify
  blocked: false
  priority: medium
  complexity: M                    # Stage framing said M; load-bearing golden pair + renderEntry lift + 14 tests keeps it honest at M.

project:
  id: PROJ-001
  stage: STAGE-003
repo:
  id: bragfile

agents:
  architect: claude-opus-4-7
  implementer: claude-opus-4-7
  created_at: 2026-04-23

references:
  decisions:
    - DEC-004   # tags comma-joined TEXT — render verbatim in metadata table
    - DEC-006   # cobra framework — --format markdown + --flat flag declarations
    - DEC-007   # required-flag validation in RunE — --flat + --format json rejection uses UserErrorf
    - DEC-011   # shared JSON shape; --out semantics locked there carry forward verbatim
    - DEC-013   # EMITTED HERE — markdown export shape (6 choices)
  constraints:
    - no-sql-in-cli-layer
    - stdout-is-for-data-stderr-is-for-humans
    - errors-wrap-with-context
    - test-before-implementation
    - one-spec-per-pr
  related_specs:
    - SPEC-006   # shipped; renderEntry's current home in internal/cli/show.go — lift target
    - SPEC-014   # shipped 2026-04-23; brag export command + internal/export package + --out semantics; SPEC-015 is purely additive on top
    - SPEC-017   # design; brag add --json + DEC-012 (stdin-JSON schema); unrelated to SPEC-015 beyond sharing the stage
---

# SPEC-015: `brag export --format markdown` and DEC-013 (markdown export shape)

## Context

Third (of three) active specs in STAGE-003 after SPEC-013 and SPEC-014
shipped on 2026-04-23. STAGE-003's value thesis is "close the loop
from accumulating corpus to something a human can paste into a
review doc and something an AI agent can read." SPEC-014 delivered
the AI-agent half via `--format json`. SPEC-015 delivers the human
half via `--format markdown`: a review-ready document with provenance,
executive summary, and entries grouped by project — the form the
user will paste into quarterly-review writeups, promo packets, and
retros.

The spec does three things in one pass:

1. **Emits DEC-013** — markdown export shape (six locked choices).
2. **Extends `brag export`** (created by SPEC-014) with
   `--format markdown` as the second accepted value. `--flat` joins as
   a markdown-only modifier; `--out` semantics carry over from
   SPEC-014 verbatim. Filter flags reuse `ListFilter`.
3. **Lifts `renderEntry`** from `internal/cli/show.go` into
   `internal/export/markdown.go`. The lift is a drive-by sanctioned by
   the stage Design Notes ("`renderEntry` lift"). `brag show`
   continues calling it (re-imported from the new location); the
   function gains a heading-level argument so markdown export can
   nest entry titles under per-project `## <project>` headings at
   level 3 while `brag show` keeps level 1.

Parent stage: `STAGE-003-reports-and-ai-friendly-i-o.md` — Design
Notes → "Markdown export shape (DEC-013 scope)" is the authoritative
lock for the six choices DEC-013 codifies; Design Notes →
"`renderEntry` lift", "Filter flag reuse", "`--out <path>` semantics",
and "Premise-audit hot spots → SPEC-015" apply directly. Project:
PROJ-001 (MVP).

## Goal

Ship (a) DEC-013 as a new decision file pinning the six markdown-
shape choices; (b) `brag export --format markdown [--flat] [--out
path] [filters]` as an extension of the SPEC-014 command,
byte-reproducible via a golden test on a fixed fixture; (c) the
`renderEntry` lift into `internal/export` with heading-level
parameterization, keeping `brag show`'s output byte-identical;
(d) a doc sweep making the markdown export documented, discoverable,
and correctly cross-referenced from tutorial, API contract, and
data-model docs.

## Inputs

- **Files to read:**
  - `/AGENTS.md` — §6 cycle rules; §7 spec anatomy; §8 DEC emission +
    honest confidence; §9 premise-audit family (SPEC-015 is ADDITION
    + STATUS-CHANGE with a renderEntry-lift import-path twist); §12
    CLI test harness rules.
  - `/projects/PROJ-001-mvp/brief.md` — "Detail on individual ideas →
    JSON export" already satisfied by SPEC-014; markdown export is
    implied by the project's "export → review doc" thesis but not
    separately spec'd in the brief.
  - `/projects/PROJ-001-mvp/stages/STAGE-003-reports-and-ai-friendly-i-o.md`
    — Design Notes → "Markdown export shape (DEC-013 scope)",
    "`renderEntry` lift", "Filter flag reuse", "`--out <path>`
    semantics", "Premise-audit hot spots → SPEC-015".
  - `/projects/PROJ-001-mvp/backlog.md` — NOT for scope; for awareness
    of deferred siblings (TOC block, `--group-by`, `--template`). Do
    NOT pull from here for SPEC-015.
  - `/docs/api-contract.md` — `brag export` section (post-SPEC-014)
    forward-references `markdown` as "arrives in SPEC-015"; SPEC-015
    removes the forward reference and documents the shipped shape.
  - `/docs/tutorial.md` — §4 "Machine-readable output" (post-SPEC-014)
    gains a markdown-export example; §9 "What's NOT there yet" strikes
    the markdown row; §2 Scope blurb updates since markdown is no
    longer deferred.
  - `/docs/data-model.md` — gets a DEC-013 cross-reference alongside
    the existing DEC-011 reference.
  - `/guidance/constraints.yaml` — full constraint list.
  - `/decisions/DEC-004-tags-comma-joined-for-mvp.md` — tags render
    verbatim in the metadata table.
  - `/decisions/DEC-006-cobra-cli-framework.md`
  - `/decisions/DEC-007-required-flag-validation-in-runE.md` — applies
    to `--flat` + `--format json` mutual-exclusion.
  - `/decisions/DEC-011-json-output-shape.md` — sibling export format;
    `--out` semantics locked there carry forward.
  - `/decisions/DEC-013-markdown-export-shape.md` — emitted by THIS
    spec; the six-choice lock.
  - `/projects/PROJ-001-mvp/specs/done/SPEC-006-brag-show-command.md`
    — `renderEntry`'s current home + tests (which stay green after
    the lift).
  - `/projects/PROJ-001-mvp/specs/done/SPEC-014-json-trio-and-shared-shape-dec.md`
    — shipped 2026-04-23; `internal/export` package, `brag export`
    command, `--out` semantics, golden-test pattern.
  - `/internal/cli/show.go` — current `renderEntry` definition (lines
    70–92) + single caller at `runShow` line 66.
  - `/internal/cli/show_test.go` — existing tests; stay green
    unchanged after the lift (they assert on output substrings, not
    on the function's import path).
  - `/internal/cli/export.go` — current `NewExportCmd` + `runExport`;
    gains a `--format markdown` branch and a `--flat` flag.
  - `/internal/cli/export_test.go` — existing 7 tests from SPEC-014;
    extended with 5 new markdown-path tests.
  - `/internal/export/json.go` — existing package; new file
    `markdown.go` lands alongside.
  - `/internal/storage/entry.go` — `Entry` struct field names for the
    metadata table (`ID/Title/Description/Tags/Project/Type/Impact/
    CreatedAt/UpdatedAt`).
- **External APIs:** none. stdlib only (`bytes`, `fmt`, `io`, `sort`,
  `strings`, `time`).
- **Related code paths:** `internal/cli/`, `internal/export/`,
  `docs/`, `README.md`, `decisions/`.

## Outputs

- **Files created:**
  - `/decisions/DEC-013-markdown-export-shape.md` — emitted in design
    alongside this spec. Six locked choices with honest confidence
    (0.82); alternatives (ASC-vs-DESC within group, TOC block,
    `--group-by`, custom templates, flat-only default); consequences;
    revisit criteria; cross-refs to SPEC-015/014/006 and
    DEC-004/006/007/011.
  - `/internal/export/markdown.go` — new file in the existing
    `internal/export` package (created by SPEC-014). Exports:
    - `func RenderEntry(w io.Writer, e storage.Entry, headingLevel int)` —
      the lifted helper. `headingLevel` applies to the title
      (`strings.Repeat("#", headingLevel) + " "` prefix); the
      description heading auto-derives as `headingLevel + 1`. Writes
      to `w`; no return value (matches original `renderEntry`
      signature apart from the new int arg).
    - `type MarkdownOptions struct { Flat bool; Filters string; Now
      time.Time }` — exported struct passed by the CLI layer to
      `ToMarkdown`. `Flat` selects the flat-vs-grouped branch.
      `Filters` is the pre-formatted value of the provenance
      `Filters:` line (`"(none)"` when no filters set; the CLI layer
      builds this string from `cmd.Flags().Changed(...)`). `Now` is
      the timestamp that renders into the `Exported:` provenance
      line — injected (not `time.Now()`-called internally) so tests
      produce deterministic goldens.
    - `func ToMarkdown(entries []storage.Entry, opts MarkdownOptions)
      ([]byte, error)` — top-level renderer. Returns bytes with the
      trailing `\n` stripped (matches SPEC-014's `ToJSON` contract;
      the CLI layer writes them with `fmt.Fprintln` which adds one
      newline back). Empty entries slice returns header + provenance
      block only (no summary, no groups). See Notes for the
      Implementer for layout sketches.
  - `/internal/export/markdown_test.go` — new file. Nine tests (see
    Failing Tests) including the load-bearing
    `TestToMarkdown_DEC013FullDocumentGolden`.
- **Files modified:**
  - `/internal/cli/show.go` — `runShow` call site updated from
    `renderEntry(cmd.OutOrStdout(), entry)` to
    `export.RenderEntry(cmd.OutOrStdout(), entry, 1)`. The local
    `renderEntry` function (lines 70–92) is removed; its body moves
    verbatim into `internal/export/markdown.go`'s `RenderEntry` with
    the heading-level parameterization applied. Import added:
    `"github.com/jysf/bragfile000/internal/export"`.
  - `/internal/cli/export.go` — `runExport` gains:
    - `--format` accepted values expand: `json` OR `markdown`. Unknown
      values still `UserErrorf`. Missing `--format` error message
      updated to `(accepted: json, markdown)`.
    - New `--flat` bool flag declared alongside `--out`.
    - New validation: if `--flat` set AND `--format != "markdown"` →
      `UserErrorf("--flat only applies to --format markdown")`.
    - Switch in the post-list section: `case "json"` routes through
      `export.ToJSON` (existing); `case "markdown"` routes through
      `export.ToMarkdown` with a `MarkdownOptions` struct built from
      `--flat`, the assembled `Filters:` string, and
      `time.Now().UTC()`.
    - `--format` flag's help text updated:
      `"output format (required; one of: json, markdown)"`.
    - Long / Examples block gains a `brag export --format markdown`
      line and a `--flat` example.
  - `/internal/cli/export_test.go` — five new tests appended (see
    Failing Tests). Order them after SPEC-014's tests with a comment
    block naming SPEC-015.
  - `/docs/api-contract.md` — `brag export` section (post-SPEC-014)
    updated: `--format` accepted values become `json, markdown`; the
    `markdown arrives in SPEC-015` forward reference (line 196) is
    replaced with a description of the shipped shape. New bullets
    mention `--flat`, provenance, executive summary, grouping default.
    Cross-link to DEC-013 added.
  - `/docs/tutorial.md` — (a) §4 "Machine-readable output: `--format
    json|tsv`" section gains a sibling subsection
    `### Review-ready export: --format markdown` (or similar) with a
    short example; (b) §9 "What's NOT there yet" table: strike
    `brag export --format markdown` row; (c) §2 Scope blurb updates
    (now that both JSON and markdown export ship, "export" as a whole
    is in scope — only `brag summary` remains deferred).
  - `/docs/data-model.md` — add DEC-013 alongside DEC-011 in the
    References list at the bottom.
  - `/README.md` — scope blurb (line ~58) updated: removes "`brag
    export --format markdown` arrives in a later STAGE-003 spec" and
    mentions markdown as a shipped format alongside JSON.
- **New exports:**
  - `export.RenderEntry(w io.Writer, e storage.Entry, headingLevel
    int)` — lifted from `internal/cli`.
  - `export.ToMarkdown(entries []storage.Entry, opts MarkdownOptions)
    ([]byte, error)`.
  - `export.MarkdownOptions` struct.
- **Database changes:** none. Pure read path; uses existing
  `Store.List(ListFilter)` from SPEC-007. No migration.

## Acceptance Criteria

Every criterion is testable. Paired failing test name in italics.

- [ ] DEC-013 exists at `/decisions/DEC-013-markdown-export-shape.md`
      with the six locked choices, rejected alternatives (DESC-within-
      group, TOC, `--group-by`, custom templates, flat-only default),
      honest confidence (0.82), and references to SPEC-015/014/006
      and DEC-004/006/007/011. *[manual: `ls decisions/DEC-013*`; grep
      for "0.82" and "grouped entries" in it.]*
- [ ] `export.ToMarkdown(fixture, opts{Flat: false, Filters: "(none)",
      Now: <fixed>})` emits a byte-identical document locking all six
      DEC-013 choices from the fixture. *TestToMarkdown_DEC013FullDocumentGolden*
- [ ] `export.ToMarkdown(fixture, opts{Flat: true, Filters: "(none)",
      Now: <fixed>})` emits a byte-identical document for the flat
      variant: no `## <project>` groups, single `## Entries
      (chronological)` wrapper, entries chrono-ASC across all projects,
      `---` separators between every pair. *TestToMarkdown_FlatGolden*
- [ ] `export.RenderEntry(w, entry, 1)` emits the same bytes
      `internal/cli/show.go:renderEntry` did before the lift on the
      same full-fields entry: `# <title>` title, full metadata table
      with all optional rows, `## Description` heading. *TestRenderEntry_HeadingLevel1*
- [ ] `export.RenderEntry(w, entry, 3)` shifts headings: `### <title>`
      title, `#### Description` heading. Metadata table is unchanged
      across heading levels. *TestRenderEntry_HeadingLevel3*
- [ ] `export.RenderEntry` on a minimal entry (title + required
      server fields only, no optional metadata, no description) omits
      the optional metadata rows (`| tags |`, `| project |`, `| type
      |`, `| impact |`) and omits the description heading entirely.
      Behavior matches pre-lift `renderEntry`. *TestRenderEntry_OmitsEmptyMetadataAndDescription*
- [ ] `export.ToMarkdown` grouping rules: project groups in
      alphabetical-ASC order; `(no project)` renders LAST regardless of
      count; within each group entries sort by `created_at` ASC (not
      DESC); `---` separator appears between every pair of consecutive
      entries WITHIN a group and not between groups. *TestToMarkdown_GroupingOrderRules*
- [ ] `export.ToMarkdown` summary block rules: `**By type**` and
      `**By project**` each sort DESC by count with alphabetical-ASC
      tie-break; `(no project)` is forced last in the by-project list
      regardless of count. *TestToMarkdown_SummaryCountsAndSorting*
- [ ] `export.ToMarkdown` provenance `Filters:` line renders verbatim
      from `opts.Filters`: `"(none)"` when no filters, echoed flag
      string (e.g. `--project platform --since 7d`) when filters were
      passed. *TestToMarkdown_FiltersLineFormat* (subtests: `none`,
      `echoed_flags`).
- [ ] `export.ToMarkdown` on empty entries slice emits header +
      provenance block with `Entries: 0` + `Filters: <value>`; NO
      summary section, NO groups section. Document ends after the
      `Filters:` line. *TestToMarkdown_EmptyEntriesEmitsHeaderAndZeroCount*
- [ ] `brag export --format markdown` writes ToMarkdown bytes +
      trailing newline to stdout (byte-identical to the helper's
      output + `\n`). Filter flags apply. *TestExportCmd_FormatMarkdown_StdoutEmitsDEC013Shape*
- [ ] `brag export --format markdown --out <path>` writes the same
      bytes (plus trailing newline) to `<path>`, overwriting without
      prompt. Stdout stays empty. Mirrors SPEC-014's JSON `--out`
      semantics. *TestExportCmd_FormatMarkdown_OutPathWritesFile*
- [ ] `brag export --format markdown --project <X>` applies the
      filter before rendering; filtered-out entries do not appear in
      any group or in the summary counts. *TestExportCmd_FormatMarkdown_FiltersApply*
- [ ] `brag export --format json --flat` returns `UserErrorf`; error
      message contains `--flat` and `--format markdown`. Stdout empty.
      *TestExportCmd_FlatWithJSONIsUserError*
- [ ] `brag export --help` help text contains `--format` AND
      `markdown` (in addition to the existing `json`) AND `--flat`.
      *TestExportCmd_HelpShowsFormatMarkdownAndFlat*
- [ ] `brag show` regression: all existing `show_test.go` tests stay
      green unchanged after the `renderEntry` lift. *[manual: `go
      test ./internal/cli/ -run TestShowCmd` passes.]*
- [ ] SPEC-014 regression: all existing `export_test.go` tests stay
      green after the `--format markdown` addition (the error-message
      substring `json` stays present in the `(accepted: json,
      markdown)` message; the help test's `--format` + `json` needles
      remain). *[manual: `go test ./internal/cli/ -run TestExportCmd`
      passes.]*
- [ ] `gofmt -l .` empty; `go vet ./...` clean; `CGO_ENABLED=0 go
      build ./...` succeeds; `go test ./...` and `just test` green.
- [ ] Doc sweep: `docs/api-contract.md`, `docs/tutorial.md`,
      `docs/data-model.md`, and `README.md` all updated per Outputs.
      The forward-reference "markdown arrives in SPEC-015" is gone.
      *[manual greps listed under Premise audit below.]*

## Locked design decisions

Nine locked decisions. Each paired with at least one failing test
per AGENTS.md §9 (SPEC-009 ship lesson). Each is derived from the
stage Design Notes or DEC-013; build / verify do NOT re-litigate.

1. **DEC-013 six choices (emitted in `/decisions/DEC-013-*`):**
   document structure with provenance + summary + grouped entries;
   group-by-project default with alphabetical-ASC ordering and
   `(no project)` last; entries chrono-ASC within group; `###` title
   via lifted `RenderEntry`, `---` within-group separators; `--flat`
   escape with `## Entries (chronological)` wrapper at level 2;
   `--out` writer semantics carry over from SPEC-014. All six
   verifiable from the grouped-mode and flat-mode golden tests.
   *Pair: `TestToMarkdown_DEC013FullDocumentGolden` +
   `TestToMarkdown_FlatGolden`.*
2. **`internal/export/markdown.go` is the single source of markdown
   rendering.** Both `brag export --format markdown` and `brag show`
   (post-lift) call into it. *Pair:
   `TestToMarkdown_DEC013FullDocumentGolden` proves export-path
   integrity; `TestRenderEntry_HeadingLevel1` proves show-path
   integrity; existing `TestShowCmd_*` tests stay green after the
   lift (manual verification + the regression criteria above).*
3. **`renderEntry` lift is additive and behavior-preserving.**
   `brag show`'s output is byte-identical before and after the lift.
   The only change is the function's import path (from
   `internal/cli` to `internal/export`) and the added heading-level
   argument. `runShow` passes `1` to preserve pre-lift heading
   level. *Pair: `TestRenderEntry_HeadingLevel1` +
   `TestRenderEntry_OmitsEmptyMetadataAndDescription` + existing
   `show_test.go` tests staying green.*
4. **Heading-level parameterization: title at `headingLevel`;
   description at `headingLevel + 1`.** `brag show` uses 1 (title
   `#`, description `##`). `brag export --format markdown` uses 3
   (title `###`, description `####`) in both grouped and flat modes
   — consistency beats per-mode variance. *Pair:
   `TestRenderEntry_HeadingLevel1` +
   `TestRenderEntry_HeadingLevel3`.*
5. **Summary block ordering: DESC by count, alphabetical-ASC
   tie-break, `(no project)` last in by-project block regardless of
   count.** When entries empty, skip the summary block AND the
   groups section entirely (document ends after the provenance
   block). *Pair: `TestToMarkdown_SummaryCountsAndSorting` +
   `TestToMarkdown_EmptyEntriesEmitsHeaderAndZeroCount`.*
6. **Provenance `Filters:` line format.** `(none)` when no filter
   flags were passed; echoed long-form flags in
   flag-declaration order (`--tag`, `--project`, `--type`,
   `--since`, `--limit`) when filters are present — matches what
   a user would retype from the shell. The CLI layer builds this
   string from `cmd.Flags().Changed(...)`; `ToMarkdown` treats it
   as opaque. *Pair: `TestToMarkdown_FiltersLineFormat` (subtests
   `none` and `echoed_flags`).*
7. **`--flat` is markdown-only.** `brag export --format json --flat`
   returns `UserErrorf` naming both `--flat` and `--format
   markdown`. `--flat` without any `--format` hits the existing
   `--format required` error from SPEC-014 first (no ambiguity).
   *Pair: `TestExportCmd_FlatWithJSONIsUserError`.*
8. **`--format markdown` reuses SPEC-014's filter plumbing and
   `--out` writer.** No new filter logic. `ListFilter` is populated
   identically to the JSON path. `--out <path>` writes the markdown
   body plus trailing newline; existing file overwritten without
   prompt. *Pair: `TestExportCmd_FormatMarkdown_FiltersApply` +
   `TestExportCmd_FormatMarkdown_OutPathWritesFile`.*
9. **`ToMarkdown` output byte contract matches `ToJSON`.** Helper
   returns bytes with trailing `\n` stripped; CLI layer writes them
   via `fmt.Fprintln` which appends exactly one `\n`. Both stdout
   and `--out` file write paths append that single trailing newline
   — file viewers and Unix tools (`wc -l`, `cat`, `tail`) see a
   normally-terminated text file. *Pair:
   `TestExportCmd_FormatMarkdown_StdoutEmitsDEC013Shape` +
   `TestExportCmd_FormatMarkdown_OutPathWritesFile` both assert the
   trailing newline explicitly.*

**Out of scope (by design — backlog entries exist for each):**

- Table of contents block (`--toc`).
- `--group-by type|tag|month` multi-axis grouping.
- `--template <path>` custom markdown rendering.
- HTML / PDF / resume-bullet output formats.
- `brag export --format sqlite` (still deferred per STAGE-003
  2026-04-23 reshuffle).
- DESC-within-group ordering (locked to ASC here; if a real
  workflow benefits from DESC, split into a flag via a new spec).
- Automatic filter-summarization (e.g. "last 30 days" rather than
  `--since 30d` in the Filters line). Echo the literal flags.

## Premise audit (AGENTS.md §9 — addition + status-change)

SPEC-015 is an **addition** case (new `--format` accepted value, new
`--flat` flag, new file in existing `internal/export` package, new
DEC, new exported function) with **status-change** flavor (existing
doc stubs advertise markdown as "arrives in SPEC-015" and a tutorial
row lists markdown as "NOT there yet" — both supersede). Addition
heuristics and status-change heuristics both apply.

**Addition heuristics** (SPEC-011 ship lesson — grep tracked
collections for count coupling):

- DEC collection: SPEC-014 added DEC-011. SPEC-015 adds DEC-013
  (skipping DEC-012 because SPEC-017 emits that). No test asserts on
  the DEC count. Safe.
- `--format` accepted values: SPEC-014 shipped `json` as the only
  accepted value. SPEC-015 adds `markdown`. Existing SPEC-014 tests:
  - `TestExportCmd_FormatRequiredIsUserError` asserts the error
    message contains `json` substring — stays true under the new
    `(accepted: json, markdown)` message. No update needed.
  - `TestExportCmd_FormatUnknownValueIsUserError` uses `"yaml"` as
    the unknown value and asserts the error contains `yaml` and
    `json` — both substrings still present. No update needed.
  - `TestExportCmd_HelpShowsFormat` asserts help contains `--format`
    and `json` — both still present. No update needed. (SPEC-015
    adds a NEW test, `TestExportCmd_HelpShowsFormatMarkdownAndFlat`,
    to lock the additional advertised value and flag.)
- `internal/export` package: SPEC-014 shipped `json.go` with exports
  `ToJSON`, `TSVHeader`, `ToTSVRow`. SPEC-015 adds `markdown.go` to
  the same package with exports `RenderEntry`, `ToMarkdown`,
  `MarkdownOptions`. No test asserts on the package's export list.
  Safe.
- `runAdd` / `runList` / `runExport` dispatch: SPEC-015 adds a branch
  to `runExport`'s `--format` switch. SPEC-017 (separately) will add
  a `--json` dispatch branch to `runAdd`. No test counts dispatch
  branches. Safe.

**Status-change heuristics** (SPEC-012 ship lesson — grep feature
name across docs):

Explicit grep commands for the build session to run, with expected
doc-level actions in parens:

```
grep -rn 'format markdown\|--format markdown\|markdown export' docs/ README.md AGENTS.md
  # → docs/api-contract.md line ~196 ("`markdown` arrives in SPEC-015.")
  #   → REPLACE the forward-reference with a concrete description:
  #     accepted value `markdown`, `--flat` modifier, shape per DEC-013.
  # → docs/tutorial.md §9 line ~379 ("| `brag export --format markdown` | STAGE-003 |")
  #   → STRIKE the row entirely (markdown is no longer deferred).
  # → README.md scope blurb line ~58 ("`brag export --format markdown` and `brag summary` arrive in later STAGE-003 specs.")
  #   → UPDATE to mention markdown as shipped; `brag summary` stays deferred.
  # → AGENTS.md §5 directory structure may mention `internal/export/` with
  #   a future annotation; optional update, not load-bearing.

grep -rn 'SPEC-015' docs/ README.md
  # → should be 1 hit (docs/api-contract.md forward-reference from SPEC-014).
  #   Remove that forward-reference as part of the rewrite above.

grep -rn '\bbrag show\b' docs/ README.md
  # → Zero behavior change — `brag show` output is byte-identical after
  #   the lift. Docs describing its output stay accurate. No updates needed.
  #   This grep's purpose is to confirm no doc says "renderEntry lives in
  #   internal/cli" (none do) and to audit whether any doc needs updating
  #   if the internal package structure is documented anywhere. Expected
  #   result: no action needed; just verifies.

grep -rn 'renderEntry\|RenderEntry' .
  # → Before lift: hits in internal/cli/show.go definition + call site only.
  #   Build session: after lifting, verify grep returns:
  #     - internal/export/markdown.go (definition as RenderEntry)
  #     - internal/cli/show.go (call site, updated to export.RenderEntry)
  #     - tests referencing RenderEntry
  #   No other files in internal/cli/ reference renderEntry today
  #   (verified 2026-04-23 during SPEC-015 design).
```

**Existing test audit** (addition-case doesn't add tracked-count
coupling; no existing tests break, but verify):

- `internal/cli/show_test.go` — asserts on output substrings, not on
  the function's import path. Stays green after the lift.
- `internal/cli/export_test.go` — SPEC-014's tests assert on `json`
  substrings that remain present under new `(accepted: json,
  markdown)` messaging. Stay green.
- `internal/export/json_test.go` — untouched. Stays green.

**Symmetric action from `## Outputs`:** every grep hit above maps to
a concrete file modification in Outputs. No doc discoveries expected
at build time.

## Failing Tests

Written now, during **design**. Fourteen tests total across two
files. All follow AGENTS.md §9: separate `outBuf`/`errBuf` per CLI
test with no-cross-leakage asserts; fail-first run before
implementation; assertion-specificity on help/error substrings;
every locked decision paired with at least one test. Goldens reuse
a single fixture so grouping + ordering + summary rules all anchor
to the same canonical entries.

### `internal/export/markdown_test.go` (new file — 9 tests)

Tests against a fixed `[]storage.Entry` fixture plus explicit
`MarkdownOptions`. No cobra, no DB.

**Shared fixture** (used by the two golden tests and most others):

```go
// Fixture: 4 entries across 3 groups (alpha: 2, beta: 1, (no project): 1).
// Varied field-presence: entry 1 has all fields; entry 2 is minimal (no
// description, no tags, no impact); entry 3 has description only; entry 4
// has no project (→ "(no project)" group) and is minimal.
// Timestamps are chosen so:
//   * chrono-ASC across all entries: 1 (T1), 3 (T3), 2 (T2), 4 (T4)
//   * within-alpha chrono-ASC: 1 (T1), 2 (T2)
var fixture = []storage.Entry{
    {
        ID: 1, Title: "alpha-old",
        Description: "old alpha", Tags: "auth",
        Project: "alpha", Type: "shipped", Impact: "did stuff",
        CreatedAt: time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC),
        UpdatedAt: time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC),
    },
    {
        ID: 2, Title: "alpha-new",
        Project: "alpha", Type: "shipped",
        CreatedAt: time.Date(2026, 4, 21, 10, 0, 0, 0, time.UTC),
        UpdatedAt: time.Date(2026, 4, 21, 10, 0, 0, 0, time.UTC),
    },
    {
        ID: 3, Title: "beta-only",
        Description: "beta desc",
        Project: "beta", Type: "learned",
        CreatedAt: time.Date(2026, 4, 20, 11, 0, 0, 0, time.UTC),
        UpdatedAt: time.Date(2026, 4, 20, 11, 0, 0, 0, time.UTC),
    },
    {
        ID: 4, Title: "unbound",
        Type: "shipped",
        CreatedAt: time.Date(2026, 4, 22, 10, 0, 0, 0, time.UTC),
        UpdatedAt: time.Date(2026, 4, 22, 10, 0, 0, 0, time.UTC),
    },
}

var fixedNow = time.Date(2026, 4, 23, 12, 0, 0, 0, time.UTC)
```

Type counts on this fixture: `shipped: 3, learned: 1` (entries 1, 2,
4 are shipped; entry 3 is learned).
Project counts: `alpha: 2, beta: 1, (no project): 1` — alpha first
(count-DESC), beta next (alphabetical-ASC over `(no project)` by
name, but `(no project)` is forced last anyway).

#### Test 1 — `TestToMarkdown_DEC013FullDocumentGolden` (LOAD-BEARING, write FIRST)

Build `opts := MarkdownOptions{Flat: false, Filters: "(none)", Now:
fixedNow}`. Call `got, err := ToMarkdown(fixture, opts)`. Assert
`err == nil` and `bytes.Equal(got, []byte(want))` where `want` is
the literal string below. If this fails, DEC-013 has been violated
— fix code, not test.

Expected output (grouped mode, byte-exact):

```
# Bragfile Export

Exported: 2026-04-23T12:00:00Z
Entries: 4
Filters: (none)

## Summary

**By type**
- shipped: 3
- learned: 1

**By project**
- alpha: 2
- beta: 1
- (no project): 1

## alpha

### alpha-old

| field       | value |
| ----------- | ----- |
| id          | 1 |
| created_at  | 2026-04-20T10:00:00Z |
| updated_at  | 2026-04-20T10:00:00Z |
| tags        | auth |
| project     | alpha |
| type        | shipped |
| impact      | did stuff |

#### Description

old alpha

---

### alpha-new

| field       | value |
| ----------- | ----- |
| id          | 2 |
| created_at  | 2026-04-21T10:00:00Z |
| updated_at  | 2026-04-21T10:00:00Z |
| project     | alpha |
| type        | shipped |

## beta

### beta-only

| field       | value |
| ----------- | ----- |
| id          | 3 |
| created_at  | 2026-04-20T11:00:00Z |
| updated_at  | 2026-04-20T11:00:00Z |
| project     | beta |
| type        | learned |

#### Description

beta desc

## (no project)

### unbound

| field       | value |
| ----------- | ----- |
| id          | 4 |
| created_at  | 2026-04-22T10:00:00Z |
| updated_at  | 2026-04-22T10:00:00Z |
| type        | shipped |
```

**Last byte is `|`** (the closing pipe of the `unbound` entry's
final table row). ToMarkdown strips the trailing `\n`; the CLI
layer adds it back via `fmt.Fprintln`. Test asserts on the
stripped form.

Assertions:
1. `bytes.Equal(got, []byte(want))` — single exact-byte compare.
2. On failure, print both `want` and `got` to aid diffing (copy from
   SPEC-014's `TestToJSON_DEC011ShapeGolden` error formatting).

#### Test 2 — `TestToMarkdown_FlatGolden`

Build `opts := MarkdownOptions{Flat: true, Filters: "(none)", Now:
fixedNow}`. Call `got, err := ToMarkdown(fixture, opts)`. Assert
`bytes.Equal(got, []byte(want))` with the flat-mode golden below.

Expected output (flat mode, chrono-ASC across all entries, single
`## Entries (chronological)` wrapper):

```
# Bragfile Export

Exported: 2026-04-23T12:00:00Z
Entries: 4
Filters: (none)

## Summary

**By type**
- shipped: 3
- learned: 1

**By project**
- alpha: 2
- beta: 1
- (no project): 1

## Entries (chronological)

### alpha-old

| field       | value |
| ----------- | ----- |
| id          | 1 |
| created_at  | 2026-04-20T10:00:00Z |
| updated_at  | 2026-04-20T10:00:00Z |
| tags        | auth |
| project     | alpha |
| type        | shipped |
| impact      | did stuff |

#### Description

old alpha

---

### beta-only

| field       | value |
| ----------- | ----- |
| id          | 3 |
| created_at  | 2026-04-20T11:00:00Z |
| updated_at  | 2026-04-20T11:00:00Z |
| project     | beta |
| type        | learned |

#### Description

beta desc

---

### alpha-new

| field       | value |
| ----------- | ----- |
| id          | 2 |
| created_at  | 2026-04-21T10:00:00Z |
| updated_at  | 2026-04-21T10:00:00Z |
| project     | alpha |
| type        | shipped |

---

### unbound

| field       | value |
| ----------- | ----- |
| id          | 4 |
| created_at  | 2026-04-22T10:00:00Z |
| updated_at  | 2026-04-22T10:00:00Z |
| type        | shipped |
```

**Last byte is `|`** (same trailing-newline contract).

#### Test 3 — `TestRenderEntry_HeadingLevel1`

Call `RenderEntry(w, fixture[0], 1)` (the full-fields entry). Assert
the output contains `# alpha-old\n\n` (title at level 1) and `##
Description\n\n` (description heading at level 2). Assert it does NOT
contain `### alpha-old` or `#### Description` (wrong heading levels).

Assert the metadata table contains every optional row (`| tags`, `|
project`, `| type`, `| impact`) in the same order the current
(pre-lift) `renderEntry` emits. Use `strings.Contains` for each
expected line; full-byte match is covered by the golden tests above.

Also: assert that the output matches a pre-lift snapshot. Build the
snapshot by calling the CURRENT `renderEntry` (in `show.go`) on the
fixture — this guarantees the lift preserves byte output exactly.
(Build phase: this assertion means the first thing you do is capture
`renderEntry(buf, fixture[0])` via a temporary test helper before
deleting the function, then assert
`RenderEntry(&newBuf, fixture[0], 1).String() == oldSnapshot`.)

#### Test 4 — `TestRenderEntry_HeadingLevel3`

Same fixture entry. Call `RenderEntry(w, fixture[0], 3)`. Assert the
output contains `### alpha-old\n\n` and `#### Description\n\n`.
Assert it does NOT contain `# alpha-old` (level-1 title) or `## alpha-old` or `## Description`. Metadata table is identical to
heading-level-1 output (heading shifts don't affect the table).

#### Test 5 — `TestRenderEntry_OmitsEmptyMetadataAndDescription`

Call `RenderEntry(w, fixture[3], 3)` (the `unbound` entry: title +
type, no tags/project/impact/description). Assert output:
- contains `### unbound\n\n` (title).
- contains `| id          | 4 |`, `| created_at  |`, `| updated_at
  |`, `| type        | shipped |` (the four rows that always render:
  id/created/updated + one optional row that's set).
- does NOT contain `| tags`, `| project`, `| impact` (empty optional
  rows suppressed).
- does NOT contain `#### Description` or `## Description` (no
  description heading when description is empty).

Pairs decision 3 (lift is behavior-preserving) and decision 4
(heading-level parameterization doesn't affect the optional-row
suppression rule).

#### Test 6 — `TestToMarkdown_GroupingOrderRules`

Reuse the shared fixture. Render with `Flat: false`. Parse the
output into lines; walk them:

1. Find the lines starting with `## ` (level-2 headings). Expected
   sequence (ignoring `## Summary`):
   `["## Summary", "## alpha", "## beta", "## (no project)"]`.
   Assert this exact sequence (proves alphabetical ordering AND
   `(no project)` last). Write a helper that extracts all level-2
   headings in order.
2. Within the `## alpha` section (between `## alpha` and `## beta`),
   find lines starting with `### ` (level-3 headings). Expected:
   `["### alpha-old", "### alpha-new"]` — chrono-ASC within group.
3. Count `---` separators between `### alpha-old` and `### alpha-new`:
   exactly 1 (within-group separator).
4. Count `---` separators between the last alpha entry and `## beta`:
   exactly 0 (no cross-group separator).

Pairs decisions 1 (DEC-013 choices 2, 3, 4) via focused assertions
even if the full golden test above isn't the one debugging.

#### Test 7 — `TestToMarkdown_SummaryCountsAndSorting`

Reuse the shared fixture. Render with `Flat: false`. Extract the
lines between `**By type**` and `**By project**`; assert the bulleted
list is exactly:
```
- shipped: 3
- learned: 1
```

Extract lines between `**By project**` and `## alpha`; assert:
```
- alpha: 2
- beta: 1
- (no project): 1
```

Then construct a second fixture that exercises the tie-break: two
types with equal count, e.g., `shipped: 2, learned: 2, fixed: 1`.
Expected bulleted order DESC by count, alphabetical-ASC tiebreak:
`learned: 2`, `shipped: 2`, `fixed: 1`.

Also: a third fixture where `(no project)` has the HIGHEST count
among project groups. Expected bulleted `By project` list forces
`(no project)` last despite highest count:
```
- beta: 3
- alpha: 2
- (no project): 5
```

Pairs decision 5 (summary ordering rules).

#### Test 8 — `TestToMarkdown_FiltersLineFormat`

Subtests `none` and `echoed_flags`:

- `none`: Render `ToMarkdown(fixture, MarkdownOptions{Flat: false,
  Filters: "(none)", Now: fixedNow})`. Assert output contains
  `Filters: (none)\n` (stable substring with trailing newline, no
  leading whitespace).
- `echoed_flags`: Render with `Filters: "--project platform --since
  7d"`. Assert output contains `Filters: --project platform --since
  7d\n`. `ToMarkdown` treats `Filters` as opaque — CLI layer is
  responsible for assembly. This test locks the passthrough.

Pairs decision 6 (Filters line format) on the helper side. Decision
6's CLI-layer assembly is covered by the integration test
`TestExportCmd_FormatMarkdown_FiltersApply`.

#### Test 9 — `TestToMarkdown_EmptyEntriesEmitsHeaderAndZeroCount`

Render `ToMarkdown([]storage.Entry{}, MarkdownOptions{Flat: false,
Filters: "(none)", Now: fixedNow})`. Expected bytes (byte-exact):

```
# Bragfile Export

Exported: 2026-04-23T12:00:00Z
Entries: 0
Filters: (none)
```

**Last byte is `)`** (closing paren of `(none)`, trailing newline
stripped). No `## Summary` heading. No groups. Document ends after
the Filters line.

Assertions:
1. `bytes.Equal(got, []byte(want))` — byte-exact.
2. `!strings.Contains(got, "## Summary")` — double-lock the absence.
3. `!strings.Contains(got, "## ")` in the groups sense — no group
   headings.

### `internal/cli/export_test.go` (5 new tests appended)

Add after SPEC-014's tests with a separator comment block naming
SPEC-015. Reuse `newExportTestRoot`, `runExportCmd`, `seedListEntry`
from the existing file.

#### Test 10 — `TestExportCmd_FormatMarkdown_StdoutEmitsDEC013Shape`

Seed the DB with two entries (e.g. `{Title: "a", Project: "p"}`,
`{Title: "b", Project: "p"}`). Run `brag export --format markdown`.
Assertions:

1. `err == nil`; `stderr == ""`.
2. Stdout starts with `# Bragfile Export\n\n` and contains
   `Entries: 2`, `## p` (the one project group), `### a`, `### b`.
3. Stdout ends with `\n` (cli's Fprintln trailing newline).
4. Stdout, with trailing `\n` stripped, equals what
   `export.ToMarkdown(entries, opts)` returns when called directly
   against the same entries and `opts` built from the same filters
   and `time.Now()`.

To handle the `time.Now()` timing flakiness for assertion 4: use a
regex to strip the `Exported: <RFC3339>` line from both sides before
comparing (the `Exported:` line is the only time-dependent content;
everything else is deterministic from the entry fixture).

#### Test 11 — `TestExportCmd_FormatMarkdown_OutPathWritesFile`

Mirror SPEC-014's `TestExportCmd_FormatJSON_OutPathWritesFile`:

1. Seed two entries.
2. Pre-create `outPath := filepath.Join(t.TempDir(), "export.md")`
   with sentinel bytes `"PRE-EXISTING CONTENT\n"`.
3. Run `brag export --format markdown --out <outPath>`.
4. Assert `stdout == ""`, `stderr == ""`, `err == nil`.
5. Read the file. Assert its last byte is `\n` (cli's Fprintln
   behavior on file writes matches stdout).
6. Assert the file does NOT contain `"PRE-EXISTING CONTENT"`
   (overwritten).
7. Assert the file starts with `# Bragfile Export\n` and contains
   both entry titles.

#### Test 12 — `TestExportCmd_FormatMarkdown_FiltersApply`

Seed three entries across projects: `platform`, `growth`, `""`. Run
`brag export --format markdown --project platform`. Assertions:

1. `err == nil`; `stderr == ""`.
2. Stdout contains the `platform`-only entry title and not the other
   two.
3. Stdout's `## Summary` block reflects the filtered count: `Entries:
   1`, `**By project**\n- platform: 1` (and no other project lines).
4. Stdout's `Filters:` provenance line contains `--project platform`
   (echoed verbatim).

#### Test 13 — `TestExportCmd_FlatWithJSONIsUserError`

Seed one entry (so the happy path would otherwise succeed). Run
`brag export --format json --flat`. Assertions:

1. `err != nil`; `errors.Is(err, ErrUser) == true`.
2. `stdout == ""`; stderr empty (cli error goes through main.go,
   which isn't in the test harness — stderrBuf is what cobra wrote
   to, which is nothing for a RunE-returned error).
3. Error message contains `--flat` AND `--format markdown` (pairs
   decision 7's "error names both").

Pairs decision 7 directly.

#### Test 14 — `TestExportCmd_HelpShowsFormatMarkdownAndFlat`

Run `brag export --help`. Assertions:

1. `err == nil`; `errBuf.Len() == 0`.
2. Help output contains `--format` AND `json` AND `markdown` AND
   `--flat`. `markdown` proves the advertised value; `--flat` proves
   the new flag appears in help.

Pairs decision 8 (help visibility of new surface).

### Test count summary

**14 tests across 2 files:**

- `internal/export/markdown_test.go` (new) — 9 tests (including the
  load-bearing `TestToMarkdown_DEC013FullDocumentGolden` first).
- `internal/cli/export_test.go` (extended) — 5 new tests appended
  after SPEC-014's block.

Plus regression locks (existing tests continue to pass without
modification):
- All `internal/cli/show_test.go` tests — after `renderEntry` lift,
  output bytes preserved.
- All SPEC-014 `internal/cli/export_test.go` tests — `json`
  substring assertions survive the new `(accepted: json, markdown)`
  messaging.
- All `internal/export/json_test.go` tests — untouched.

## Implementation Context

*Read this section and the files it points to before starting the
build cycle. It is the equivalent of a handoff document, folded into
the spec since there is no separate receiving agent.*

### Decisions that apply

- **`DEC-013`** (emitted in this spec) — six locked markdown-shape
  choices. Every deviation from the golden in
  `TestToMarkdown_DEC013FullDocumentGolden` /
  `TestToMarkdown_FlatGolden` is a DEC-013 violation — stop and raise
  a question before "fixing" the test.
- **`DEC-011`** (shipped in SPEC-014) — sibling JSON shape. Its
  `--out` semantics (overwrite without prompt; directory/unwritable
  path returns internal error; stdout when absent) carry forward
  verbatim. The byte-contract pattern (helper strips trailing `\n`;
  CLI appends one via Fprintln) is copied exactly.
- **`DEC-004`** — tags comma-joined TEXT. The metadata table
  renders tags verbatim (`| tags        | auth,perf |`). Do NOT
  split into an array or rewrite. Storage-to-render is identity.
- **`DEC-006`** — cobra framework. `--format markdown` is a new
  accepted value on the existing `--format` flag; `--flat` is a new
  bool flag declared via `cmd.Flags().Bool("flat", false, "...")`.
  No new command.
- **`DEC-007`** — required-flag validation in `RunE`. `--flat` +
  `--format json` mutual-exclusion check uses `UserErrorf`; never
  `MarkFlagExclusive` or similar cobra helpers (they return
  unwrappable plain errors).

### Constraints that apply

For `internal/cli/**`, `internal/export/**`, `docs/**`, `README.md`,
`decisions/**`:

- `no-sql-in-cli-layer` — blocking. `export.go` continues to work
  with `[]storage.Entry` returned by `Store.List`; `markdown.go`
  never imports `database/sql`.
- `stdout-is-for-data-stderr-is-for-humans` — blocking. Markdown
  body goes to stdout (or `--out` file); human-facing errors stay on
  stderr via main.go's wrapper. Every happy-path test asserts
  `errBuf.Len() == 0`.
- `errors-wrap-with-context` — warning. Any io errors from the
  `--out` writer wrap with `fmt.Errorf("write output: %w", err)`.
  ToMarkdown itself can't error in practice (pure byte assembly over
  a slice) but its signature returns `error` for consistency with
  `ToJSON` and future-proofing.
- `test-before-implementation` — blocking. Write all 14 tests
  first. Run `go test ./internal/export ./internal/cli` and confirm
  every new test fails for the expected reason: undefined symbol
  (`ToMarkdown`, `RenderEntry`, `MarkdownOptions`) OR unknown flag
  (`--flat`) OR unknown `--format` value (`markdown`). NO
  compilation errors unrelated to the spec. If fail-first reports an
  unexpectedly-passing test, investigate (SPEC-005 lesson —
  assertion too weak or substring unintentionally present).
- `one-spec-per-pr` — blocking. Branch
  `feat/spec-015-brag-export-markdown-and-shape-dec`. Diff touches
  only the files in Outputs.
- `no-new-top-level-deps-without-decision` — not triggered;
  markdown rendering uses only stdlib (`bytes`, `fmt`, `io`,
  `sort`, `strings`, `time`).

### AGENTS.md lessons that apply

- **§9 separate `outBuf` / `errBuf`** (SPEC-001) — every cli test
  asserts both buffers.
- **§9 fail-first** (SPEC-003) — confirm each of the 14 tests fails
  for the expected reason before implementation.
- **§9 assertion specificity** (SPEC-005) — help/error tests assert
  on distinctive needles (`markdown`, `--flat`, `--format markdown`).
- **§9 locked-decisions-need-tests** (SPEC-009) — nine locked
  decisions above; each paired.
- **§9 premise audit — addition case** (SPEC-011) — grepped above;
  no tracked-collection count coupling; safe to add.
- **§9 premise audit — status-change case** (SPEC-012) — grepped
  above; doc-level actions enumerated under Outputs and Premise
  audit. The "markdown arrives in SPEC-015" forward-reference is the
  status-change hotspot.
- **SPEC-014 ship reflection**: load-bearing assertion deserves top
  billing — this spec writes `TestToMarkdown_DEC013FullDocumentGolden`
  first in `markdown_test.go`. The implementation order is
  helper-first (`markdown.go`) before command-level tests
  (`export_test.go`).

### Prior related work

- **SPEC-006** (shipped 2026-04-20) — original `brag show` and the
  current `renderEntry`. This spec lifts that helper, preserving
  byte-output on the `brag show` path.
- **SPEC-014** (shipped 2026-04-23) — `brag export` command,
  `internal/export` package, `--out` semantics, golden-test pattern,
  `ToJSON` / `TSVHeader` / `ToTSVRow`. SPEC-015 is purely additive
  on top: new file in the same package, new accepted value on an
  existing flag, new flag alongside existing ones, filter plumbing
  verbatim.
- **SPEC-007** (shipped 2026-04-20) — `ListFilter` struct. SPEC-015
  reuses verbatim via SPEC-014's filter-copy in `runExport`.

### Out of scope (for this spec specifically)

- Table of contents (`--toc`) — backlog entry exists.
- `--group-by <field>` — backlog entry exists.
- `--template <path>` — backlog entry exists.
- HTML / PDF / resume-bullet formats — project brief excludes.
- DESC-within-group ordering — DEC-013 locks ASC.
- Auto-humanizing the Filters echo (`--since 7d` → `"last 7 days"`)
  — echo literal flags.
- `brag summary` — STAGE-004 polish-pass item.
- `brag export --format sqlite` — deferred to backlog 2026-04-23.
- Changes to `storage.Entry`, `storage.Store`, or `storage.ListFilter` —
  pure CLI + export-package work.
- Any modification to SPEC-006's `show_test.go` assertions — the lift
  must be fully behavior-preserving for `brag show`.

## Notes for the Implementer

Gotchas, style preferences, reuse opportunities. Read after the
Implementation Context.

### `internal/export/markdown.go` layout

Build the file in three layers:

**Layer 1 — `RenderEntry`** (lift + heading-level arg):

```go
// RenderEntry writes e as a markdown block to w. The title appears at
// `headingLevel` (e.g., 1 → "# ", 3 → "### "). The "Description"
// sub-heading appears one level below (headingLevel + 1).
// Optional metadata rows (tags, project, type, impact) are suppressed
// when empty; an entry with no description omits the description
// heading entirely. Pre-lift behavior for `brag show` is preserved
// exactly at headingLevel == 1.
func RenderEntry(w io.Writer, e storage.Entry, headingLevel int) {
    titlePrefix := strings.Repeat("#", headingLevel) + " "
    descPrefix  := strings.Repeat("#", headingLevel+1) + " "

    fmt.Fprintf(w, "%s%s\n\n", titlePrefix, e.Title)
    fmt.Fprintln(w, "| field       | value |")
    fmt.Fprintln(w, "| ----------- | ----- |")
    fmt.Fprintf(w, "| id          | %d |\n", e.ID)
    fmt.Fprintf(w, "| created_at  | %s |\n", e.CreatedAt.UTC().Format(time.RFC3339))
    fmt.Fprintf(w, "| updated_at  | %s |\n", e.UpdatedAt.UTC().Format(time.RFC3339))
    if e.Tags != "" {
        fmt.Fprintf(w, "| tags        | %s |\n", e.Tags)
    }
    if e.Project != "" {
        fmt.Fprintf(w, "| project     | %s |\n", e.Project)
    }
    if e.Type != "" {
        fmt.Fprintf(w, "| type        | %s |\n", e.Type)
    }
    if e.Impact != "" {
        fmt.Fprintf(w, "| impact      | %s |\n", e.Impact)
    }
    if e.Description != "" {
        fmt.Fprintf(w, "\n%sDescription\n\n%s\n", descPrefix, e.Description)
    }
}
```

This is the body of the original `internal/cli/show.go:renderEntry`
with `titlePrefix` and `descPrefix` substituted. Tests 3/4/5 lock
behavior.

**Layer 2 — `MarkdownOptions`**:

```go
type MarkdownOptions struct {
    Flat    bool
    Filters string    // pre-formatted: "(none)" or echoed flags like "--project platform --since 7d"
    Now     time.Time // injected; renders into the "Exported:" line
}
```

Exported so the CLI layer can populate from its flag state. The
`Now` field is how tests produce deterministic goldens (no
`time.Now()` call inside the helper).

**Layer 3 — `ToMarkdown`**:

```go
func ToMarkdown(entries []storage.Entry, opts MarkdownOptions) ([]byte, error) {
    var buf bytes.Buffer
    // Document header
    fmt.Fprintln(&buf, "# Bragfile Export")
    fmt.Fprintln(&buf)
    // Provenance block
    fmt.Fprintf(&buf, "Exported: %s\n", opts.Now.UTC().Format(time.RFC3339))
    fmt.Fprintf(&buf, "Entries: %d\n", len(entries))
    fmt.Fprintf(&buf, "Filters: %s\n", opts.Filters)

    if len(entries) == 0 {
        return trimTrailingNewline(buf.Bytes()), nil
    }

    // Summary block
    fmt.Fprintln(&buf)
    fmt.Fprintln(&buf, "## Summary")
    fmt.Fprintln(&buf)
    writeSummaryByType(&buf, entries)
    fmt.Fprintln(&buf)
    writeSummaryByProject(&buf, entries)

    // Entries
    if opts.Flat {
        writeFlatSection(&buf, entries)
    } else {
        writeGroupedSections(&buf, entries)
    }
    return trimTrailingNewline(buf.Bytes()), nil
}

// trimTrailingNewline is the byte-contract sibling of ToJSON's
// json.MarshalIndent (which returns bytes without trailing newline).
// The CLI layer calls fmt.Fprintln which appends one newline.
func trimTrailingNewline(b []byte) []byte {
    return bytes.TrimRight(b, "\n")
}
```

Helper functions `writeSummaryByType`, `writeSummaryByProject`,
`writeFlatSection`, `writeGroupedSections` are unexported — keep the
package surface small (three exports: `RenderEntry`, `ToMarkdown`,
`MarkdownOptions`).

### Summary-block ordering

`writeSummaryByType`: build `map[string]int` of type counts; sort
keys by `-count` primary, alphabetical-ASC secondary; emit bulleted
list.

`writeSummaryByProject`: build `map[string]int` of project counts,
where the empty-string project is replaced with the literal key
`"(no project)"`. Sort the result DESC by count with alphabetical-
ASC tiebreak, THEN swap `(no project)` to the end of the slice if
present. Emit bulleted list.

Example helper shape:

```go
type kv struct{ key string; count int }

func sortedCountsDescByCountAscByKey(m map[string]int) []kv {
    out := make([]kv, 0, len(m))
    for k, v := range m { out = append(out, kv{k, v}) }
    sort.Slice(out, func(i, j int) bool {
        if out[i].count != out[j].count {
            return out[i].count > out[j].count
        }
        return out[i].key < out[j].key
    })
    return out
}

func forceNoProjectLast(sorted []kv) []kv {
    const marker = "(no project)"
    idx := -1
    for i, e := range sorted { if e.key == marker { idx = i; break } }
    if idx < 0 { return sorted }
    np := sorted[idx]
    sorted = append(sorted[:idx], sorted[idx+1:]...)
    return append(sorted, np)
}
```

### Grouped vs flat rendering

Grouped (`writeGroupedSections`):

1. Partition `entries` into `map[string][]storage.Entry` keyed by
   project (empty project maps to `"(no project)"`).
2. Sort project keys alphabetical-ASC, force `(no project)` last.
3. For each project group:
   - Emit blank line + `## <project>\n\n`.
   - Sort group entries by `CreatedAt` ASC.
   - For each entry in the group, emit:
     - `RenderEntry(&buf, e, 3)` — writes title + metadata + optional
       description with `\n` terminator.
     - If NOT the last entry in this group, emit `\n---\n\n`
       (blank-line-before-separator-and-after).
4. Between groups: no separator; the `## <project>` of the next
   group IS the transition.

Flat (`writeFlatSection`):

1. Sort all `entries` by `CreatedAt` ASC.
2. Emit blank line + `## Entries (chronological)\n\n`.
3. For each entry:
   - `RenderEntry(&buf, e, 3)`.
   - If NOT the last entry, emit `\n---\n\n`.

Note the subtle spacing: `RenderEntry` ends its output with `\n`
(from the last `fmt.Fprintf(..., "...\n", ...)` call). Then the
separator block writes `\n---\n\n`. This produces exactly one blank
line before and one blank line after the `---`, which is what the
golden shows. Between a group's last entry and the next `## <project>`,
the caller emits `\n## <next>\n\n` — one blank line between them.

### CLI-layer changes (`internal/cli/export.go`)

1. Declare the new flag in `NewExportCmd`:
   ```go
   cmd.Flags().Bool("flat", false, "skip grouping in --format markdown (chrono-ASC single section)")
   ```

2. Update the `--format` flag's help string:
   ```go
   cmd.Flags().String("format", "", "output format (required; one of: json, markdown)")
   ```

3. Update the Long/Examples block in the cobra command to mention
   `--format markdown` and `--flat`:
   ```
   brag export --format markdown                       # stdout: grouped markdown
   brag export --format markdown --flat                # stdout: flat chronological
   brag export --format markdown --out report.md       # write to file
   brag export --format markdown --project platform    # filter before exporting
   ```

4. In `runExport`, update the format validation and dispatch:
   ```go
   format, _ := cmd.Flags().GetString("format")
   if format == "" {
       return UserErrorf("--format is required (accepted: json, markdown)")
   }
   if format != "json" && format != "markdown" {
       return UserErrorf("unknown --format value %q (accepted: json, markdown)", format)
   }

   flat, _ := cmd.Flags().GetBool("flat")
   if flat && format != "markdown" {
       return UserErrorf("--flat only applies to --format markdown")
   }
   ```

5. After `entries, err := s.List(filter)`, branch on format:
   ```go
   var body []byte
   switch format {
   case "json":
       body, err = export.ToJSON(entries)
   case "markdown":
       body, err = export.ToMarkdown(entries, export.MarkdownOptions{
           Flat:    flat,
           Filters: echoFilters(cmd),
           Now:     time.Now().UTC(),
       })
   }
   if err != nil {
       return fmt.Errorf("marshal %s: %w", format, err)
   }
   ```

6. `echoFilters` is a new unexported helper in `export.go` (NOT in
   `internal/export` — it's a cli-layer concern because it reads
   `cmd.Flags().Changed(...)`). Signature:
   ```go
   func echoFilters(cmd *cobra.Command) string {
       var parts []string
       order := []string{"tag", "project", "type", "since", "limit"}
       for _, name := range order {
           if !cmd.Flags().Changed(name) { continue }
           if name == "limit" {
               n, _ := cmd.Flags().GetInt(name)
               parts = append(parts, fmt.Sprintf("--%s %d", name, n))
           } else {
               v, _ := cmd.Flags().GetString(name)
               parts = append(parts, fmt.Sprintf("--%s %s", name, v))
           }
       }
       if len(parts) == 0 {
           return "(none)"
       }
       return strings.Join(parts, " ")
   }
   ```

### CLI-layer changes (`internal/cli/show.go`)

One line change plus an import. Replace:
```go
renderEntry(cmd.OutOrStdout(), entry)
```
with:
```go
export.RenderEntry(cmd.OutOrStdout(), entry, 1)
```

Add import:
```go
"github.com/jysf/bragfile000/internal/export"
```

Delete the local `renderEntry` function (lines 70–92) — its body
moves to `internal/export/markdown.go:RenderEntry` verbatim apart
from the heading-level prefix substitution.

Remove any now-unused imports from `show.go` (`io` if only used by
`renderEntry`; `time` might still be used elsewhere — check after
the edit).

### Build order (strict)

1. Create `feat/spec-015-brag-export-markdown-and-shape-dec` branch
   (already done in design).
2. Write the 9 tests in `internal/export/markdown_test.go`. The
   load-bearing golden goes FIRST. Run `go test
   ./internal/export/` and confirm every new test fails with an
   undefined-symbol error (`ToMarkdown`, `RenderEntry`,
   `MarkdownOptions`).
3. Implement `internal/export/markdown.go`:
   - `RenderEntry` first (lifted with heading-level arg).
   - `MarkdownOptions` struct.
   - `ToMarkdown` + unexported helpers.
4. Run `go test ./internal/export/` — all 9 tests should pass.
5. Capture a pre-lift snapshot of `internal/cli/show.go:renderEntry`
   output on a full-fields entry (for Test 3's byte-preservation
   assertion). Easiest: write a small table-driven test using the
   current (unlifted) function, save the bytes as a const, then
   delete the current function.
6. Lift: update `internal/cli/show.go` — delete local `renderEntry`,
   call `export.RenderEntry(..., 1)`. Run `go test
   ./internal/cli/ -run TestShowCmd` — must stay green.
7. Write the 5 new tests in `internal/cli/export_test.go` (appended
   after SPEC-014's block with a comment separator). Run `go test
   ./internal/cli/ -run TestExportCmd` — new tests fail (flag not
   declared or unknown format value).
8. Implement cobra changes in `internal/cli/export.go`: declare
   `--flat`, update format validation, add markdown branch in the
   switch, add `echoFilters` helper, update Long/Examples and help
   text. Run the export tests — all 12 green.
9. Run the full test suite: `go test ./...` green; `gofmt -l .`
   empty; `go vet ./...` clean; `CGO_ENABLED=0 go build ./...`
   succeeds.
10. Doc sweep: `docs/api-contract.md`, `docs/tutorial.md`,
    `docs/data-model.md`, `README.md` per Outputs.
11. Fill in Build Completion + three reflection answers. Advance
    cycle to verify. Open PR.

### fail-first runs

Expected fail-first output when running `go test ./internal/export
./internal/cli -run "TestToMarkdown|TestRenderEntry|TestExportCmd_FormatMarkdown|TestExportCmd_Flat|TestExportCmd_HelpShowsFormatMarkdown"`:

- `markdown_test.go`: "undefined: ToMarkdown", "undefined:
  RenderEntry", "undefined: MarkdownOptions". Compilation failure —
  that is the expected fail-first signal.
- `export_test.go` new tests: "unknown flag: --flat" OR "unknown
  --format value \"markdown\"" depending on which is reached first.

If any new test passes on fail-first, investigate (SPEC-005 lesson).

### Heading level convention

- `brag show`: `headingLevel = 1` (title `# <title>`, description
  `## Description`).
- `brag export --format markdown` (grouped): `headingLevel = 3`
  (title `### <title>` nested under `## <project>`, description
  `#### Description`).
- `brag export --format markdown --flat`: `headingLevel = 3` (title
  `### <title>` nested under `## Entries (chronological)`,
  description `#### Description`). Same level as grouped mode —
  consistency beats per-mode variance.

### Pitfalls to avoid

- **Don't re-sort within the CLI layer.** `Store.List(filter)`
  already returns `created_at DESC, id DESC`. `ToMarkdown`
  re-sorts: ASC-within-group for grouped, ASC-across-all for flat.
  Pass the unsorted (DESC) slice into `ToMarkdown`; the helper does
  the re-sort. Keeps the CLI layer simple and puts sort logic next
  to the rendering that depends on it.
- **Don't forget the `(no project)` tie-break in the summary.** The
  sort helper handles alphabetical-ASC tie-break, but `(no project)`
  must be force-last AFTER sorting. Tested in
  `TestToMarkdown_SummaryCountsAndSorting`.
- **Trailing-newline contract.** `ToMarkdown` returns bytes with
  trailing `\n` stripped. CLI's `fmt.Fprintln(out, string(body))`
  adds one back. File writes via `--out` use the same pattern.
  Tested in `TestExportCmd_FormatMarkdown_StdoutEmitsDEC013Shape`
  and `TestExportCmd_FormatMarkdown_OutPathWritesFile`.
- **Time injection.** `ToMarkdown` takes `opts.Now`. CLI passes
  `time.Now().UTC()`. Tests pass a fixed time. Do NOT call
  `time.Now()` inside the helper — it breaks test determinism.
- **Heading-level bounds.** `headingLevel = 0` would render `" <title>"`
  with no hash prefix; `headingLevel = 7` would render `####### <title>`
  which is outside common markdown heading support (usually level 6
  max). This spec only calls with 1 or 3, so don't add runtime
  validation — premature. If a future caller passes junk, tests
  break; that's the contract.
- **`show_test.go` must not be touched.** The lift is
  behavior-preserving. If any show test fails after the lift,
  something's wrong with the lift, not with the tests. Stop and
  debug.

### Deviations-to-flag-during-build

If any of the following happen during build, flag under Deviations
in Build Completion before advancing to verify:

- You considered inlining `echoFilters` into `runExport` vs
  extracting it to a helper — spec recommends helper. Either is
  fine; note which you chose.
- You discovered a doc file this spec didn't enumerate. Update the
  file in the same commit and note it under Deviations.
- The `--flat` help text or error message needs wording tweaks to
  satisfy a failing test assertion — OK, but document the final
  wording in Deviations so future grep audits know what to expect.

### Branch name

`feat/spec-015-brag-export-markdown-and-shape-dec` (already created
during design). Commit message prefix: `feat(SPEC-015): ...`.

---

## Build Completion

*Filled in at the end of the **build** cycle, before advancing to verify.*

- **Branch:** `feat/spec-015-brag-export-markdown-and-shape-dec`
- **PR (if applicable):** pending (opened at end of build)
- **All acceptance criteria met?** yes
- **New decisions emitted:**
  - None. DEC-013 was design-time and pre-existed this build cycle.
- **Deviations from spec:**
  - **TDD order slip (self-corrected):** in Step 2 of the build plan I
    initially wrote the full `ToMarkdown` implementation alongside the
    `RenderEntry` lift rather than stopping at the lift. On realizing
    this bypassed the fail-first signal the spec asked for, I reverted
    `markdown.go` to only `RenderEntry` + `MarkdownOptions` (struct
    only), wrote all 14 tests, ran them, and confirmed the expected
    failure modes (`undefined: ToMarkdown`, `unknown flag: --flat`,
    `unknown --format value "markdown"`, help text missing `markdown`
    and `--flat`) before re-adding `ToMarkdown` and the CLI changes.
    Fail-first discipline was ultimately honored; the slip is noted so
    a future auditor reading the git log sees only the good end-state.
  - **Assertion shape for `TestRenderEntry_HeadingLevel3`:** the spec's
    sketch said "Assert it does NOT contain `# alpha-old` or `##
    alpha-old`"; a naive `strings.Contains` fires on "### alpha-old"
    which contains "# alpha-old" and "## alpha-old" as substrings. I
    switched to line-based equality (`ln == "# alpha-old"`) so the
    assertion is faithful to the intent. Same pattern applied to
    "## Description" vs "#### Description".
  - **Assertion shape for `TestToMarkdown_GroupingOrderRules`:** the
    spec's sketch checked "after the last alpha entry, before `##
    beta`, NO `---`". I implemented this as "`---` must not appear
    after `### alpha-new` in the alpha section" (more direct than
    LastIndex + ordering math).
  - `echoFilters` extracted to a helper (vs inlined) per the spec's
    recommendation.
- **Follow-up work identified:**
  - None beyond the already-enumerated backlog entries (`--toc`,
    `--group-by`, `--template`, DESC-within-group, sqlite export).

### Build-phase reflection (3 questions, short answers)

Process-focused: how did the build go? What friction did the spec create?

1. **What was unclear in the spec that slowed you down?**
   — Minor: the "assert NOT contains `# alpha-old`" pattern in the
   level-3 heading test is a substring-trap under `strings.Contains` —
   any "### alpha-old" line satisfies that substring. The spec should
   specify line-based assertions for heading-level tests so
   implementers don't rediscover the trap. Apart from that, the
   literal goldens, the lift sequencing, and the `echoFilters` sketch
   made the build mostly mechanical.

2. **Was there a constraint or decision that should have been listed but wasn't?**
   — The implicit "no blank line between `## <project>` section and
   its first `### <title>` that isn't covered by `RenderEntry`'s own
   title-prefix newlines" is subtle; I only caught it by tracing the
   byte sequence against the golden. The Notes for the Implementer's
   "subtle spacing" paragraph flagged it, so this is more "read that
   paragraph twice" than a missing constraint. Net: no gap.

3. **If you did this task again, what would you do differently?**
   — Stop at Step 2 exactly — lift `renderEntry`, add the single
   heading-level-1 test, run it, stop. Do not keep typing into
   `markdown.go` "while I'm there" because the remaining struct +
   `ToMarkdown` skeleton belongs to Step 4, not Step 2. Writing tests
   against an already-implemented function weakens the fail-first
   signal even when the test content is correct.

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
