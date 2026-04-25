---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-018
  type: story                      # epic | story | task | bug | chore
  cycle: ship
  blocked: false
  priority: medium
  complexity: M                    # M honestly: DEC emission + new package + new command + four-things-in-one. Stage Design Notes called this the load-bearing spec of STAGE-004.

project:
  id: PROJ-001
  stage: STAGE-004
repo:
  id: bragfile

agents:
  architect: claude-opus-4-7
  implementer: claude-opus-4-7     # usually same Claude, different session
  created_at: 2026-04-25

references:
  decisions:
    - DEC-004   # tags comma-joined TEXT — no normalization at the I/O boundary
    - DEC-006   # cobra framework — new `brag summary` subcommand
    - DEC-007   # required-flag validation in RunE — `--range` and `--format` use UserErrorf
    - DEC-011   # naked-array JSON shape — DEC-014 INTENTIONALLY DIVERGES (envelope, not array)
    - DEC-013   # markdown export shape — DEC-014 reuses the provenance/summary block convention
    - DEC-014   # EMITTED HERE — rule-based output shape (envelope JSON + provenance markdown) for summary/review/stats
  constraints:
    - no-sql-in-cli-layer
    - stdout-is-for-data-stderr-is-for-humans
    - errors-wrap-with-context
    - test-before-implementation
    - one-spec-per-pr
  related_specs:
    - SPEC-007   # shipped; ListFilter struct + --tag/--project/--type flags reused verbatim
    - SPEC-014   # shipped; structural mirror — emitted DEC-011 + created internal/export package + anchored shape for downstream
    - SPEC-015   # shipped; provenance/summary-block precedent (DEC-013) that DEC-014 reuses for the markdown half
    - SPEC-019   # pending; will reuse DEC-014 envelope + internal/aggregate.ByProject + add review-grouping helpers
    - SPEC-020   # pending; will reuse DEC-014 envelope + extend internal/aggregate with Streak/MostCommon/Span
---

# SPEC-018: `brag summary --range week|month` + DEC-014 (rule-based output shape) + seeds `internal/aggregate`

## Context

First (of three) specs in STAGE-004. The stage's value thesis is "turn
the accumulating corpus into AI-pipeable reflection material without
shipping an LLM in the binary." STAGE-003 closed the input/output loop
(capture → filter → durable JSON/markdown export); STAGE-004 adds the
rule-based aggregation surface that sits between "list everything" and
"reflect on a period." `brag summary --range week|month` is the
lightest-weight of the three: a digest of counts and grouped highlights
(titles + IDs only — no descriptions) for a rolling 7- or 30-day
window, designed to be skimmed before pasting into an external AI
session for deeper reflection.

The spec is the load-bearing one in STAGE-004 because it does **four
things in one pass**:

1. **Emits DEC-014** (rule-based output shape — single-object JSON
   envelope + markdown provenance/summary convention). DEC-014 locks
   the cross-cutting shape choices that SPEC-019 (`brag review`) and
   SPEC-020 (`brag stats`) inherit without re-litigating.
2. **Seeds `internal/aggregate`** — a new package separate from
   `internal/export`. Aggregation maps `[]storage.Entry → structured
   stats`; rendering maps `structured stats → bytes`. SPEC-018 lands
   `ByType`, `ByProject`, `GroupForHighlights`. SPEC-019 will reuse
   `ByProject` and add review-grouping helpers; SPEC-020 will extend
   with `Streak`, `MostCommon`, `Span`.
3. **Implements `brag summary`** — the digest command. `--range
   week|month` required; `--format markdown|json` honored (markdown
   default); filter flags `--tag --project --type` reuse `ListFilter`;
   stdout only.
4. **Replaces the api-contract.md placeholder** at lines 251–261 (the
   brief's earliest sketch left a stub there) with the real, shipped
   shape.

DEC-014 is the structural mirror of SPEC-014's DEC-011: same "one
shape, one helper, one byte-identical golden" pattern, but the shape
itself **intentionally diverges from DEC-011**. Where DEC-011 is a
naked JSON array (because `list`/`export` emit entry lists), DEC-014
is a single-object envelope (because aggregations carry metadata —
`generated_at`, `scope`, `filters` — that does not fit a per-entry
row). Both DECs document this divergence explicitly so future
contributors see it as a deliberate symmetry, not a drift.

Parent stage:
[`STAGE-004-rule-based-polish-summary-review-stats.md`](../stages/STAGE-004-rule-based-polish-summary-review-stats.md) —
Spec Backlog → SPEC-018 entry, Design Notes → "DEC-014" /
"`internal/aggregate` package" / "Filter flag reuse" / "Output
destination" / "Premise audit → SPEC-018" / "SPEC-018-specific". Project:
PROJ-001 (MVP).

## Goal

Ship (a) DEC-014 as a new decision file pinning the six rule-based
shape choices (single-object envelope, top-level payload keys,
provenance/summary markdown convention, empty-state values,
indent=2 JSON, range = rolling-window semantics); (b) the new
`internal/aggregate` package with `ByType`, `ByProject`,
`GroupForHighlights` plus their result types; (c) `brag summary
--range week|month [filters] [--format markdown|json]` as the first
DEC-014 consumer; (d) the api-contract.md placeholder REPLACED with
the shipped shape (status-change premise audit) and tutorial/README
mentions corrected.

## Inputs

- **Files to read:**
  - `/AGENTS.md` — §6 cycle rules; §7 spec anatomy; §8 DEC emission +
    honest confidence; §9 premise-audit family with the SPEC-015
    substring-trap addendum and the SPEC-017 freshness-assertion
    addendum (SPEC-018 is ADDITION + STATUS-CHANGE — both heuristics
    apply); §12 CLI test harness rules.
  - `/projects/PROJ-001-mvp/brief.md` — §"Stage Plan" STAGE-004 entry
    (provisional sketch) and §"Detail on individual ideas → summary"
    if present; treat the stage-plan entry plus the stage doc as the
    authoritative scope, not the brief sketch.
  - `/projects/PROJ-001-mvp/stages/STAGE-004-rule-based-polish-summary-review-stats.md`
    — THE authoritative scope for this spec. Spec Backlog → SPEC-018
    (lines ~190–201) — locked scope; Design Notes → "DEC-014",
    "`internal/aggregate` package", "Filter flag reuse", "Output
    destination", "Premise audit → SPEC-018 hot spot",
    "SPEC-018-specific (`brag summary`)".
  - `/projects/PROJ-001-mvp/backlog.md` — NOT for scope; for
    awareness of out-of-scope siblings (`brag remind`, emoji passes,
    `brag add --at`, `--group-by` for export markdown which is a
    different command, `--out` deferral, `--compact` JSON
    deferral).
  - `/docs/api-contract.md` — lines 251–261 carry a placeholder `brag
    summary` section from the earliest brief sketch. SPEC-018
    REPLACES that section. Other sections (`brag list`, `brag
    export`) stay untouched.
  - `/docs/tutorial.md` — line 3 (Scope blurb mentions `brag
    summary`); line ~453 (the "What's NOT there yet" table row);
    both update from "arrives in a later stage" to "shipped in
    STAGE-004."
  - `/docs/data-model.md` — gets a DEC-014 cross-reference in the
    References list at the bottom.
  - `/README.md` — line ~61 Scope blurb mentions `brag summary`
    arrives in a later stage; updates to mention `brag summary
    --range week|month` as shipped, with `brag review` / `brag
    stats` still forward-referenced.
  - `/guidance/constraints.yaml` — full constraint list.
  - `/decisions/DEC-004-tags-comma-joined-for-mvp.md` — relevant if
    summary aggregates by tag in the future; SPEC-018 does not
    aggregate by tag, but DEC-014's shape is tag-agnostic.
  - `/decisions/DEC-006-cobra-cli-framework.md`
  - `/decisions/DEC-007-required-flag-validation-in-runE.md` —
    applies to `--range` and `--format` validation in `RunE`.
  - `/decisions/DEC-011-json-output-shape.md` — naked-array shape;
    DEC-014 INTENTIONALLY DIVERGES (envelope, not array). Read for
    contrast and to see the divergence call out.
  - `/decisions/DEC-013-markdown-export-shape.md` — provenance block
    + summary block precedent. DEC-014 reuses these conventions for
    the markdown half.
  - `/decisions/DEC-014-rule-based-output-shape.md` — emitted by
    THIS spec; the six-choice lock.
  - `/projects/PROJ-001-mvp/specs/done/SPEC-014-json-trio-and-shared-shape-dec.md`
    — DIRECT structural precedent. SPEC-014 emitted DEC-011 +
    created `internal/export` + anchored shape for SPEC-017.
    SPEC-018 mirrors that structure: emits DEC-014 + creates
    `internal/aggregate` + anchors shape for SPEC-019/020.
  - `/projects/PROJ-001-mvp/specs/done/SPEC-015-brag-export-markdown-and-shape-dec.md`
    — markdown provenance/summary block precedent that DEC-014's
    markdown half reuses. Read for the `Generated:` / `Scope:` /
    `Filters:` line shape (re-named from `Exported:` / `Entries:`
    in SPEC-015's DEC-013, but the shape pattern is the same).
  - `/projects/PROJ-001-mvp/specs/done/SPEC-007-list-filter-flags.md`
    — `ListFilter` struct + `--tag`/`--project`/`--type`/`--since`
    flags reused verbatim. Validation patterns
    (`cmd.Flags().Changed(...)`, empty-string user errors) carry
    forward.
  - `/internal/cli/list.go` — existing command + filter-flag
    plumbing (lines 44–82); `runSummary` populates `ListFilter`
    identically.
  - `/internal/cli/export.go` — existing `runExport` with
    `--format` validation pattern + `echoFilters` helper; both
    are direct templates for `runSummary`. `echoFilters` may be
    reused; see Notes for the Implementer.
  - `/internal/cli/errors.go` — `ErrUser` sentinel + `UserErrorf`
    helper (used for `--range`, `--format` validation).
  - `/internal/cli/since.go` — `ParseSince` (used to compute
    --range cutoff via `time.Now().AddDate(0, 0, -7)` etc., NOT
    via ParseSince — see Notes for the Implementer).
  - `/internal/storage/store.go`, `/internal/storage/entry.go` —
    `Entry` struct + `Store.List(ListFilter)` (read-only); aggregate
    package operates on the returned slice.
  - `/internal/export/json.go`, `/internal/export/markdown.go` —
    siblings of the new `summary.go` file. Read for byte-contract
    conventions (`MarshalIndent` indent=2; trim-trailing-newline
    helper; `fmt.Fprintln` writer wrapping; `MarkdownOptions`-style
    options struct with injectable `Now`).
  - `/cmd/brag/main.go` — gains one
    `root.AddCommand(cli.NewSummaryCmd())` line.
- **External APIs:** none. stdlib `encoding/json`, `time`, `sort`,
  `strings`, `bytes` cover the needs. No new Go module dependencies.
- **Related code paths:** new `internal/aggregate/`; new
  `internal/cli/summary.go`; new `internal/export/summary.go`;
  `cmd/brag/main.go`; `decisions/`; `docs/`; `README.md`.

## Outputs

- **Files created:**
  - `/decisions/DEC-014-rule-based-output-shape.md` — emitted in
    design alongside this spec. Six locked choices with honest
    confidence (0.80); alternatives (naked-array reuse-of-DEC-011,
    per-spec independent shapes, nested-`payload`-key, range-as-key
    instead of scope, filters-as-null-when-empty); consequences;
    revisit criteria; cross-refs to SPEC-018/019/020 and
    DEC-011/013/004/006/007.
  - `/internal/aggregate/aggregate.go` — new package, new file.
    Exports:
    - `type TypeCount struct { Type string; Count int }`
    - `type ProjectCount struct { Project string; Count int }`
    - `type EntryRef struct { ID int64; Title string }`
    - `type ProjectHighlights struct { Project string; Entries []EntryRef }`
    - `const NoProjectKey = "(no project)"`
    - `func ByType(entries []storage.Entry) []TypeCount` — entries
      grouped by `Type`, returned DESC by count with alphabetical-ASC
      tiebreak. Empty input → empty (non-nil) slice. Empty-string
      `Type` is included in the result with key `""` if any entry has
      it (callers decide whether to filter).
    - `func ByProject(entries []storage.Entry) []ProjectCount` —
      same shape; empty-string `Project` is rendered under
      `NoProjectKey` and **forced last** regardless of count
      (matches DEC-013's "(no project) last" convention).
    - `func GroupForHighlights(entries []storage.Entry) []ProjectHighlights`
      — entries grouped by `Project`, each group's entries sorted
      ASC by `CreatedAt` then ASC by `ID` as tie-break; groups
      themselves ordered alphabetically-ASC by project name with
      `NoProjectKey` forced last. Each `EntryRef` carries `ID`
      and `Title` only — no description, no metadata (the
      "skim before pasting" goal).
  - `/internal/aggregate/aggregate_test.go` — new file. Four
    tests against literal `[]storage.Entry` slices (no DB, no
    cobra) — see Failing Tests.
  - `/internal/export/summary.go` — new file in the existing
    `internal/export` package (created by SPEC-014, extended by
    SPEC-015). Exports:
    - `type SummaryOptions struct { Scope string; Filters string; FiltersJSON map[string]string; Now time.Time }`
      — `Scope` is the `--range` value (`"week"` or `"month"`);
      `Filters` is the pre-formatted markdown line (`"(none)"` or
      `"--project platform --tag auth"`); `FiltersJSON` is the
      object the JSON envelope renders (`map[string]string{}` when
      none, `map[string]string{"project": "platform"}` when set);
      `Now` is injected for deterministic `Generated:` lines.
    - `func ToSummaryMarkdown(entries []storage.Entry, opts SummaryOptions) ([]byte, error)`
      — renders the markdown digest per DEC-014. Returns bytes with
      trailing `\n` stripped (matches `ToJSON` / `ToMarkdown` byte
      contract). Empty entries slice returns header + provenance
      block only (no summary block, no highlights), per DEC-014
      empty-state rule.
    - `func ToSummaryJSON(entries []storage.Entry, opts SummaryOptions) ([]byte, error)`
      — renders the JSON envelope per DEC-014. Single object, top-
      level keys: `generated_at`, `scope`, `filters` (object —
      empty `{}` when no filters), `counts_by_type` (object,
      key=type, value=count), `counts_by_project` (object,
      key=project name, value=count, `(no project)` rendered
      under that literal key), `highlights` (array of `{project,
      entries: [{id, title}]}` objects, group order alphabetical-
      ASC with `(no project)` last, within-group entries
      chrono-ASC). Pretty-printed with 2-space indent.
  - `/internal/export/summary_test.go` — new file. Five tests
    against fixed `[]storage.Entry` fixtures + explicit
    `SummaryOptions`. Includes the load-bearing markdown golden
    AND the load-bearing JSON golden. See Failing Tests.
  - `/internal/cli/summary.go` — new file. Exports
    `func NewSummaryCmd() *cobra.Command` plus unexported
    `runSummary`. Declares `--range` (required, RunE-validated per
    DEC-007; accepted: `week`, `month`), `--format` (default
    `markdown`; RunE-validated; accepted: `markdown`, `json`),
    filter flags `--tag`, `--project`, `--type` (no `--since`,
    no `--limit` — `--range` covers the time window; `--limit`
    on a digest doesn't compose meaningfully). Computes the
    `--range` cutoff as `time.Now().UTC().AddDate(0, 0, -7)` for
    week or `-30` for month, populates `filter.Since`, calls
    `Store.List(filter)`, renders via `ToSummaryMarkdown` /
    `ToSummaryJSON` to stdout.
  - `/internal/cli/summary_test.go` — new file. Five tests
    using `t.TempDir()` for DB paths, seeding entries via the
    package-local `seedListEntry` helper (already in
    `internal/cli/list_test.go`, same package — direct reuse).
- **Files modified:**
  - `/cmd/brag/main.go` — one added line:
    `root.AddCommand(cli.NewSummaryCmd())` after the existing
    seven `AddCommand` calls.
  - `/docs/api-contract.md` — lines 251–261 (the placeholder
    `### brag summary --range week|month (STAGE-003)` section)
    REPLACED with the shipped shape:
    - synopsis updated to STAGE-004; full flag list (`--range`,
      `--format`, `--tag`, `--project`, `--type`);
    - prose describes provenance + counts-by-type/project +
      grouped highlights (titles + IDs only);
    - cross-link to DEC-014;
    - mention markdown is default, JSON envelope is single object
      (NOT array — divergence from DEC-011 noted);
    - rolling-window semantics (week = last 7 UTC days, month =
      last 30 UTC days, NOT calendar week/month) documented
      explicitly per DEC-014.
    The end-of-file References list (line ~284–290) gains a
    `DEC-014` row.
  - `/docs/tutorial.md` — (a) line 3 Scope blurb: `brag summary
    arrives in a later stage` → `brag review` / `brag stats`
    arrive in later STAGE-004 specs (or similar — keep it brief;
    `brag summary` is now shipped). (b) Line ~453 "What's NOT
    there yet" table: STRIKE the `| brag summary --range
    week|month | STAGE-003 |` row. (c) Optional but
    recommended: §4 "Read them back" gains a small
    `### Weekly digest: brag summary` subsection showing the
    paste-into-AI workflow. (Author judgment call — see Notes.)
  - `/docs/data-model.md` — add DEC-014 to the References list at
    the bottom alongside DEC-011 and DEC-013. No schema change.
  - `/README.md` — line ~61 Scope blurb: `brag summary arrives in
    a later STAGE-003 spec` → `brag summary --range week|month`
    is shipped (STAGE-004); `brag review` and `brag stats`
    arrive in later STAGE-004 specs.
  - `/AGENTS.md` — §11 Domain Glossary multi-edit pass:
    (a) line 250 `summary` entry STAGE-003 → STAGE-004;
    (b) line 251 `tap` entry STAGE-004 → STAGE-005 (caught
    by status-change premise audit; the cherry-pick on
    2026-04-24 moved tap to STAGE-005 but glossary line
    was not updated); (c) ADD `aggregate` entry
    (alphabetical placement); (d) ADD `digest` entry;
    (e) ADD `review` entry; (f) ADD `stats` entry. Full
    content under Notes for the Implementer → AGENTS.md
    §11 Domain Glossary updates.
- **New exports:**
  - `aggregate.TypeCount`, `aggregate.ProjectCount`,
    `aggregate.EntryRef`, `aggregate.ProjectHighlights`,
    `aggregate.NoProjectKey`, `aggregate.ByType`,
    `aggregate.ByProject`, `aggregate.GroupForHighlights`.
  - `export.SummaryOptions`, `export.ToSummaryMarkdown`,
    `export.ToSummaryJSON`.
  - `cli.NewSummaryCmd`.
- **Database changes:** none. Pure read path; uses existing
  `Store.List(ListFilter)` from SPEC-007. No migration.

## Acceptance Criteria

Every criterion is testable. Paired failing test name in italics
where applicable; combinations are noted as such. SPEC-018 has
**15 failing tests** across **3 new files**, with the load-bearing
goldens written FIRST per SPEC-014 + SPEC-015 ship lessons.

- [ ] DEC-014 exists at
      `/decisions/DEC-014-rule-based-output-shape.md` with the six
      locked choices, rejected alternatives (naked-array reuse,
      per-spec independent shapes, nested-`payload`-key, range-as-
      key instead of scope, filters-as-null), honest confidence
      (0.80), and references to SPEC-018/019/020 and
      DEC-011/013/004/006/007. *[manual: `ls decisions/DEC-014*`
      returns the file; grep for "0.80" and "single-object
      envelope" in it.]*
- [ ] `aggregate.ByType(fixture)` returns a slice ordered DESC by
      count with alphabetical-ASC tiebreak. Empty input returns
      a non-nil empty slice (`len == 0`, not `nil`).
      *TestByType_DESCByCountAlphaTiebreak*
- [ ] `aggregate.ByProject(fixture)` returns a slice ordered DESC
      by count with alphabetical-ASC tiebreak; entries with
      empty-string `Project` are rendered under `NoProjectKey`
      and forced LAST regardless of count.
      *TestByProject_NoProjectKeyForcedLast*
- [ ] `aggregate.GroupForHighlights(fixture)` returns
      `[]ProjectHighlights` with groups ordered alphabetical-ASC
      (`(no project)` last); within each group, entries are
      sorted ASC by `CreatedAt` (with `ID` as tie-break — locks
      determinism per AGENTS.md §9 SPEC-002 lesson, monotonic
      column for distinctness). Each `EntryRef` carries only
      `ID` and `Title` — no description.
      *TestGroupForHighlights_ChronoASCWithNoProjectLast*
- [ ] `aggregate` empty-input contract: `ByType(nil)`,
      `ByProject(nil)`, `GroupForHighlights(nil)` each return a
      non-nil empty slice (`len == 0`, distinct from `nil`) so
      the JSON renderer can marshal them as `[]` not `null`.
      *TestAggregate_EmptyInputReturnsNonNilEmptySlice*
- [ ] `export.ToSummaryMarkdown(fixture, opts{Scope: "week",
      Filters: "(none)", Now: <fixed>})` emits a byte-identical
      markdown document locking all DEC-014 markdown choices on
      the fixture. *(LOAD-BEARING — write FIRST per SPEC-014/15
      ship lessons.)* *TestToSummaryMarkdown_DEC014FullDocumentGolden*
- [ ] `export.ToSummaryJSON(fixture, opts{Scope: "week",
      FiltersJSON: map[string]string{}, Now: <fixed>})` emits a
      byte-identical JSON envelope locking all DEC-014 JSON
      choices on the same fixture: `generated_at`, `scope` =
      `"week"`, `filters` = `{}`, `counts_by_type` /
      `counts_by_project` as objects, `highlights` as array of
      group objects with chrono-ASC entries. *(LOAD-BEARING —
      write SECOND.)* *TestToSummaryJSON_DEC014ShapeGolden*
- [ ] `ToSummaryMarkdown` on empty entries emits `# Bragfile
      Summary` heading + provenance block (`Generated`, `Scope`,
      `Filters`) ending with the Filters line. NO `## Summary`
      section, NO highlights section. Mirrors DEC-013's
      empty-state precedent.
      *TestToSummary_EmptyEntriesEmitsProvenanceOnly* (markdown
      and JSON variants in subtests).
- [ ] `ToSummaryJSON` empty-state: `counts_by_type` is `{}`,
      `counts_by_project` is `{}`, `highlights` is `[]`,
      `generated_at` and `scope` are still populated, `filters`
      is `{}` when no filters were echoed. JSON renders as
      `[]` for empty arrays NOT `null` (matches DEC-011's
      empty-array discipline). Covered by the empty-state
      subtest above.
- [ ] `ToSummaryJSON` with non-empty `FiltersJSON` echoes filters
      as a JSON object: `{"project": "platform", "tag":
      "auth"}` (key order: alphabetical, per Go's
      `encoding/json` map sorting; deterministic across runs).
      *TestToSummaryJSON_FiltersEchoShape* (subtests `none` →
      `{}`, `populated` → object).
- [ ] `brag summary` (no `--range`) exits 1 (user error) with a
      message naming `--range` and the accepted values (`week`,
      `month`). Per DEC-007 — `UserErrorf` in `RunE`, no
      `MarkFlagRequired`. *TestSummaryCmd_RangeRequiredIsUserError*
- [ ] `brag summary --range yearly` (unknown value) exits 1
      (user error) with a message naming the unknown value and
      the accepted list. *TestSummaryCmd_RangeUnknownValueIsUserError*
- [ ] `brag summary --range week --format json` writes
      `ToSummaryJSON(entries, opts)` bytes + trailing newline to
      stdout. Filter flags compose: `--tag X --project Y --type
      Z` apply on top of `--range`. `(no project)` entries are
      preserved through the pipeline. Time window is rolling 7
      days from `time.Now().UTC()` (entries older than 7d are
      filtered out by `filter.Since`).
      *TestSummaryCmd_FormatJSON_RangeWeekAndFiltersCompose*
- [ ] `brag summary --range week` (no `--format`) defaults to
      markdown (output starts with `# Bragfile Summary`,
      contains `## Summary`, contains `## Highlights`).
      `--range week` and `--range month` produce envelopes
      whose `Scope:` line differs while document structure
      (heading set + section order) stays identical.
      *TestSummaryCmd_ScopeFieldAndMarkdownDefault*
- [ ] The rolling-window arithmetic itself is locked at the
      pure-helper level, NOT via DB-level backdating.
      `rangeCutoff("week", now)` returns `now - 7 days`;
      `rangeCutoff("month", now)` returns `now - 30 days`;
      empty and unknown values return `UserErrorf` with
      messages naming the offending input and the accepted
      list. The `Store.SetCreatedAtForTesting` alternative
      that would have allowed end-to-end DB backdating is
      explicitly rejected — see Locked design decisions →
      Rejected alternatives (build-time) §1.
      *TestRangeCutoff_WeekMonthArithmeticAndErrors*
- [ ] `brag summary --help` output contains `--range` AND
      `week` AND `month` AND `--format` AND `markdown` AND
      `json` (each as distinctive needles per AGENTS.md §9
      assertion-specificity rule). Errors empty.
      *TestSummaryCmd_HelpShowsRangeAndFormat*
- [ ] `brag --help` lists `summary` as a subcommand (cobra auto-
      registers it via `cmd/brag/main.go` AddCommand).
      *[manual: `go build ./cmd/brag && ./brag --help`
      shows `summary` in the command list; `./brag summary
      --help` shows the synopsis.]*
- [ ] `gofmt -l .` empty; `go vet ./...` clean; `CGO_ENABLED=0
      go build ./...` succeeds; `go test ./...` and `just test`
      green.
- [ ] Doc sweep: `docs/api-contract.md` lines 251–261 REPLACED
      (no longer says `(STAGE-003)` or "the brief's earliest
      sketch"); `docs/tutorial.md` line 3 + line ~453 updated;
      `docs/data-model.md` References list gains DEC-014;
      `README.md` line ~61 Scope blurb updated; `AGENTS.md`
      §11 Domain Glossary updates applied (summary entry's
      stage fix, tap entry's stale STAGE-004 → STAGE-005
      fix caught by status-change premise audit, plus four
      new entries: `aggregate`, `digest`, `review`, `stats`).
      *[manual greps listed under Premise audit below.]*

## Locked design decisions

Reproduced here so build / verify don't re-litigate. Each is paired
with at least one failing test below per AGENTS.md §9 SPEC-009
ship lesson.

1. **DEC-014 six choices (emitted in `/decisions/DEC-014-*`).**
   (1) JSON is a single object envelope, NOT a naked array
   (DIVERGES from DEC-011); (2) top-level flat keys
   `generated_at` (RFC3339), `scope` (string), `filters`
   (object), plus per-spec payload keys at top level (no nested
   `payload` wrapper); (3) markdown convention `# <Doc Title>`
   level-1 heading + provenance block (`Generated:` /
   `Scope:` / `Filters:`) + summary block (where applicable)
   reusing DEC-013's `**By type**` / `**By project**`
   convention; (4) empty-state values: numeric → 0, arrays →
   `[]` (non-nil), date fields → null in JSON / `-` in
   markdown, objects → `{}`; provenance always renders, summary
   + payload sections OMITTED on empty entries; (5) JSON pretty-
   printed indent=2 (matches DEC-011); (6) `--range`
   semantics are ROLLING window (week = last 7 UTC days from
   `time.Now()`, month = last 30 UTC days), NOT calendar
   week/month — explicit in DEC-014 and api-contract.md so
   future debate is anchored. *Pair: load-bearing
   `TestToSummaryMarkdown_DEC014FullDocumentGolden` +
   `TestToSummaryJSON_DEC014ShapeGolden` cover (1)–(5);
   `TestRangeCutoff_WeekMonthArithmeticAndErrors` covers
   (6) at the unit level; `TestSummaryCmd_ScopeFieldAndMarkdownDefault`
   covers (6) at the CLI plumbing level.*
2. **`internal/aggregate` is the SINGLE data-layer source.**
   Aggregation maps `[]storage.Entry → structured stats`;
   `internal/export` (`summary.go`) maps `structured stats →
   bytes`. The seam is locked. SPEC-019 / SPEC-020 add to
   `internal/aggregate` without touching the rendering layer's
   internals. *Pair: `TestByType_DESCByCountAlphaTiebreak` +
   `TestByProject_NoProjectKeyForcedLast` +
   `TestGroupForHighlights_ChronoASCWithNoProjectLast` exercise
   the data-layer functions directly; the goldens prove the
   render layer composes them correctly.*
3. **`(no project)` is the literal key for empty-string
   `Project`.** Locked at `aggregate.NoProjectKey`. Both the
   markdown rendering (`## (no project)` heading; bulleted
   `- (no project): N` row in summary block) and the JSON
   rendering (`"(no project)": N` key in `counts_by_project`;
   `"project": "(no project)"` value in highlights group) use
   this single sentinel. Inherits DEC-013's "no project last"
   convention. *Pair:
   `TestByProject_NoProjectKeyForcedLast` +
   `TestGroupForHighlights_ChronoASCWithNoProjectLast` +
   `TestToSummaryMarkdown_DEC014FullDocumentGolden` (asserts
   `## (no project)` last in the section ordering).*
4. **`--range week|month` REQUIRED, no default.** Missing
   `--range` → `UserErrorf` (DEC-007 pattern); never
   `MarkFlagRequired`. Accepted values: `week`, `month`.
   Symmetric with `brag export --format` (also required).
   *Pair: `TestSummaryCmd_RangeRequiredIsUserError` +
   `TestSummaryCmd_RangeUnknownValueIsUserError`.*
5. **`--format markdown|json` default markdown, validated in
   `RunE`.** Unknown values → `UserErrorf` naming the offending
   value and accepted list. *Pair:
   `TestSummaryCmd_ScopeFieldAndMarkdownDefault` (default
   branch) + `TestSummaryCmd_FormatJSON_RangeWeekAndFiltersCompose`
   (json branch) + `TestSummaryCmd_HelpShowsRangeAndFormat`
   (help advertises both).*
6. **Filter flags `--tag --project --type` reuse `ListFilter`
   verbatim.** No `--since`, no `--limit` on summary —
   `--range` covers the window; `--limit` on a digest is
   semantically odd. Validation is identical to `runList` /
   `runExport` (`cmd.Flags().Changed(...)`, empty strings →
   user error). *Pair:
   `TestSummaryCmd_FormatJSON_RangeWeekAndFiltersCompose`
   asserts a `--tag` + `--project` + `--type` combination
   filters the result.*
7. **stdout only; no `--out` flag in MVP.** Per stage Design
   Notes "Output destination" — these are pipe-friendly digests
   that users redirect with `>` if they want a file. Backlog if
   a user asks. Distinct from `brag export --out` which is the
   durable-document case. *Pair: no test specifically exercises
   the absence of `--out`; the `TestSummaryCmd_FormatJSON_*`
   tests assert `errBuf.Len() == 0` and outputs land in
   `outBuf` — implicitly verifying stdout is the only sink.*
8. **`Now` is INJECTED via `SummaryOptions.Now`, not called
   inside the renderer.** Mirrors DEC-013's `MarkdownOptions.Now`
   pattern. Tests use a fixed timestamp so goldens are
   deterministic. The CLI layer passes `time.Now().UTC()`.
   *Pair: every renderer test passes `Now: fixedNow`; the CLI
   tests assert on the substring shape and use a regex strip
   for the `Generated:` line (or accept `time.Now()`-window
   non-determinism via prefix/contains assertions instead of
   byte-exact compare).*
9. **api-contract.md placeholder REPLACED, not patched.** Per
   §9 status-change premise audit. Lines 251–261 currently
   advertise `(STAGE-003)` and a hand-waved shape; SPEC-018
   replaces with the shipped synopsis + full flag list +
   DEC-014 cross-link + rolling-window semantics. The grep for
   `brag summary` will return zero hits in api-contract.md
   AFTER the rewrite that point to the placeholder. *Pair:
   manual grep under Premise audit + the doc-sweep acceptance
   criterion.*

**Out of scope (by design — backlog entries exist or are explicitly
deferred):**

- `--out <path>` flag on summary. Backlog deferral noted in stage
  Design Notes ("Output destination").
- `--since` / arbitrary date ranges on summary. `--range
  week|month` only for MVP per stage scope.
- `--group-by <field>` on summary (axis other than project).
  Distinct from `brag export --format markdown --group-by`
  backlog entry — different command. Backlog if a real workflow
  emerges.
- `--limit` on summary. Aggregation-of-aggregations doesn't
  compose with row-cap meaningfully.
- `--compact` / non-pretty JSON for summary (or any of the three
  STAGE-004 commands). Inherits DEC-011's pretty-default; same
  backlog entry covers them.
- Time-zone configuration for `--range` cutoff. UTC-only for MVP
  (matches storage's `time.Now().UTC()`); revisit if a user
  notices a window break across timezone changes.
- `brag review` / `brag stats` — those are SPEC-019/SPEC-020.
  SPEC-018 only seeds the `internal/aggregate` package they
  consume.
- LLM piping / AI integration — PROJ-002 territory.
- Adding `--since` to ListFilter validation in summary —
  `--since` is intentionally NOT exposed by `brag summary`
  (ListFilter has it, but the flag isn't declared on the cobra
  command).
- Calendar-week / calendar-month semantics. Rolling-window only;
  if a user requests calendar boundaries (Mon–Sun, 1st–end),
  it's a follow-up spec.

**Rejected alternatives (build-time):**

These are choices the build agent might consider, with the
prescribed path locked here so the call doesn't off-load to
build-time and slip into Deviations later. Per SPEC-017 ship
reflection: "either-is-fine off-loads to build" is the
anti-pattern these locks deliberately avoid.

1. **`Store.SetCreatedAtForTesting(id int64, ts time.Time)`
   public method on `internal/storage/store.go` — REJECTED.**
   The path would let `internal/cli/summary_test.go` backdate
   one entry's `created_at` to a value between 7 and 30 days
   ago, then assert end-to-end that `--range week` excludes
   it while `--range month` includes it. This would prove the
   date-cutoff arithmetic via DB-level evidence rather than
   pure-function evidence.

   *Why rejected:*
   - **Test-only public API in production code is the anti-
     pattern this spec deliberately avoids.** The "ForTesting"
     suffix is a doc-comment guarantee, not a structural one
     — the compiler can't enforce it. A future refactor that
     drops the suffix or that calls the method from a non-
     test context can erode the boundary silently. The
     storage surface stays cleaner if it doesn't ship a
     test-only method at all.
   - **The arithmetic IS a pure function, and pure functions
     deserve pure tests.** Going through DB seeding to prove
     `now - 7 days` works correctly is end-to-end testing for
     testing's sake. Test #14 (`TestRangeCutoff_*`) covers the
     arithmetic deterministically with `now` injected as a
     parameter; that's the right layer.
   - **The CLI plumbing assertion is just as well covered
     without backdating.** Test #13
     (`TestSummaryCmd_ScopeFieldAndMarkdownDefault`) seeds
     two fresh entries and asserts the `Scope:` line differs
     between `--range week` and `--range month` invocations
     — proving the flag plumbs through. Combined with Test
     #14's arithmetic lock, the two tests cover what
     end-to-end backdating would have covered, with cleaner
     test-architecture and no production-code surface
     impact.

   If a future spec genuinely needs backdated entries (e.g.,
   SPEC-020's streak/span semantics), that spec earns its
   own decision on how to provide the test fixture. DEC-014
   / SPEC-018 do not pre-commit to the storage-layer surface
   either way.

2. **Inline `--range` parsing in `runSummary` instead of
   extracting `rangeCutoff` — REJECTED.** The parsing logic
   (week / month / unknown / empty) is small enough to live
   inline at the top of `runSummary`. Rejected because:
   - Extracting the helper makes Test #14's pure-function
     assertion straightforward; inlining would force the
     arithmetic test to either go through the cobra command
     (unnecessarily heavy) or to duplicate the `now -
     7d/30d` logic in the test (redundant and silently
     drifts on refactor).
   - The locked-decisions-need-tests discipline (AGENTS.md
     §9 SPEC-009 lesson) applies most cleanly to a named
     unit. `rangeCutoff` is the named unit; the inline
     alternative obscures the contract.

3. **Reusing `export.go`'s `echoFilters` from `runSummary`
   — REJECTED.** Already documented in Notes for the
   Implementer with the "caller has to know it's a superset"
   coupling argument. Captured here for verify visibility:
   the prescribed path is an inline `echoFiltersForSummary`
   helper in `internal/cli/summary.go` that returns both
   markdown line and JSON object in one pass, iterating
   summary's three-flag set explicitly.

## Premise audit (AGENTS.md §9 — addition + status-change)

SPEC-018 is an **addition** case (new command, new package, new
DEC, new exports) AND a **status-change** case (the api-contract.md
placeholder + tutorial.md "What's NOT there yet" row + tutorial
Scope blurb + README Scope blurb all advertise `brag summary` as
deferred behavior that SPEC-018 supersedes). Both AGENTS.md §9
heuristics apply.

**Addition heuristics** (SPEC-011 ship lesson — grep tracked
collections for count coupling):

- Root command list: `cmd/brag/main.go` has seven `AddCommand`
  calls today (verified 2026-04-25). SPEC-018 makes it eight. Run
  `grep -rn 'AddCommand\|root.Commands()\|cmd.Commands()'
  internal/cli/*.go cmd/ 2>/dev/null` and audit each hit:
  - `cmd/brag/main.go`: the seven existing calls. Adding an
    eighth doesn't break any test.
  - `internal/cli/list_test.go:627` (per SPEC-014 audit): iterates
    by name, not count. Safe.
  - No test asserts `len(root.Commands()) == 7` or similar.
  Verify before build by re-running the grep.
- DEC collection: SPEC-018 adds DEC-014. SPEC-014 added DEC-011,
  SPEC-015 added DEC-013, SPEC-017 added DEC-012. No test
  asserts on the DEC count or directory size.
  `guidance/constraints.yaml` and `AGENTS.md §15` reference the
  directory generically.
- `internal/export` package exports: SPEC-018 adds
  `SummaryOptions`, `ToSummaryMarkdown`, `ToSummaryJSON`. No
  test asserts on the package's export list.
- New `internal/aggregate` package: brand-new; nothing existed
  to break. Verify with `find internal/aggregate -type f`
  returns nothing before this spec lands.
- `--format` accepted values: distinct flag per command. The
  list/export tests asserting on `(accepted: json, tsv)` /
  `(accepted: json, markdown)` are unaffected by summary's
  separate `(accepted: markdown, json)` list.
- Help-command subcommand counts: per stage Design Notes, the
  per-spec premise-audit hot spot for SPEC-018 mentions
  "Help-command tests that count subcommands need a +1." Run
  `grep -rn 'NumCommand\|len.*Commands()' internal/cli/`. If
  no hits, no update needed; if any hit asserts a literal
  count, bump it. Verified 2026-04-25 against the working tree:
  no such literal-count assertion exists in `internal/cli/`.
- `internal/cli/list_test.go` `seedListEntry` helper: SPEC-018
  reuses it. Helper signature stays identical; reuse is
  source-level, no test surface affected.

**Status-change heuristics** (SPEC-012 ship lesson — grep feature
name across docs):

Explicit grep commands for the build session to run, with expected
doc-level actions in parens:

```
grep -rn 'brag summary' docs/ README.md AGENTS.md
  # → docs/api-contract.md lines 251–261 (REPLACE the entire
  #   placeholder section with the shipped shape: synopsis,
  #   --range/--format/filter flags, rolling-window semantics,
  #   DEC-014 cross-link).
  # → docs/api-contract.md References list line ~284–290 (ADD
  #   a DEC-014 row).
  # → docs/tutorial.md line 3 ("Scope:" blurb mentioning
  #   brag summary arrives later) — UPDATE to mention summary
  #   shipped, review/stats arrive in later STAGE-004 specs.
  # → docs/tutorial.md line ~453 ("| brag summary --range
  #   week|month | STAGE-003 |") — STRIKE the row entirely.
  # → README.md line ~61 ("brag summary arrives in a later
  #   STAGE-003 spec") — UPDATE to mention summary shipped,
  #   review/stats deferred.
  # → AGENTS.md §11 Domain Glossary mentions "summary" as
  #   STAGE-003 ("rule-based aggregation grouped by
  #   project/type over a time range. STAGE-003.") — UPDATE
  #   STAGE-003 to STAGE-004 in the entry.

grep -rn 'STAGE-003.*summary\|summary.*STAGE-003' docs/ README.md AGENTS.md
  # → catches the cross-references that the brief.md /
  #   stage docs no longer use; ensures every "summary
  #   arrives in STAGE-003" stale claim is found.
  # Expected hits: docs/api-contract.md lines ~251 + 256;
  #   docs/tutorial.md line ~453; README.md line ~61;
  #   AGENTS.md §11 glossary entry. All map to actions
  #   above.

grep -rn 'tap.*STAGE-004\|STAGE-004.*tap\|Created in STAGE-004' AGENTS.md
  # → AGENTS.md §11 line 251 ("**tap** — ... Created in
  #   STAGE-004.") — second status-change hot spot.
  #   The tap moved to STAGE-005 during the 2026-04-24
  #   cherry-pick; the glossary line was not updated then.
  #   SPEC-018's broader §11 sweep catches this.
  #   UPDATE to "Created in STAGE-005."

grep -rn 'aggregate\|digest\|brag review\|brag stats' docs/ README.md AGENTS.md
  # → AGENTS.md §11 currently has no entries for these terms.
  #   ADD four glossary entries: aggregate (the
  #   internal/aggregate package); digest (collective name
  #   for the rule-based commands); review (SPEC-019
  #   command); stats (SPEC-020 command). Full content in
  #   Notes for the Implementer → AGENTS.md §11 Domain
  #   Glossary updates.
  # → docs/, README.md: no hits expected (these are repo-
  #   level concepts; the glossary is the right home).
  # → If unexpectedly hits a doc, audit whether the doc
  #   should be updated in lockstep.

grep -rn 'placeholder\|earliest sketch' docs/api-contract.md
  # → confirms zero hits AFTER rewrite (sanity check the
  #   replacement doesn't accidentally narrate itself).

grep -rn 'internal/aggregate' .
  # → BEFORE this spec: zero hits (package doesn't exist).
  # → AFTER build: hits in
  #     internal/aggregate/aggregate.go (definitions),
  #     internal/aggregate/aggregate_test.go (tests),
  #     internal/export/summary.go (importer),
  #     internal/cli/summary.go (importer).
  # No other importers expected.

grep -rn 'NewSummaryCmd\|summary.go' internal/ cmd/
  # → BEFORE this spec: zero hits.
  # → AFTER build: definition in internal/cli/summary.go,
  #   registration in cmd/brag/main.go, tests in
  #   internal/cli/summary_test.go,
  #   internal/export/summary.go,
  #   internal/export/summary_test.go.
```

**Existing test audit** (addition-case doesn't add tracked-count
coupling; verify nothing breaks):

- `internal/cli/list_test.go` — assertions on `--format`
  needles `json`, `tsv` are list-specific; summary's
  `--format` doesn't affect them. `seedListEntry` helper is
  reused; signature stable.
- `internal/cli/export_test.go` — assertions on `--format`
  needles `json`, `markdown` are export-specific. Unaffected.
- `internal/cli/show_test.go`, `internal/cli/add_test.go`,
  `internal/cli/edit_test.go`, `internal/cli/delete_test.go`,
  `internal/cli/search_test.go`, `internal/cli/add_json_test.go`
  — no overlap with summary surface.
- `internal/export/json_test.go`,
  `internal/export/markdown_test.go` — sibling files in the
  package; new `summary_test.go` lands alongside, no
  cross-coupling.
- `internal/storage/*_test.go` — read-only path; summary uses
  existing `Store.List`. Unaffected.

**Symmetric action from `## Outputs`:** every grep hit above maps
to a concrete file modification in Outputs (api-contract.md,
tutorial.md, README.md, data-model.md, AGENTS.md §11). No
discoveries expected at build time.

## Failing Tests

Written now, during **design**. Fifteen tests total across **3
new files**. All follow AGENTS.md §9: separate `outBuf` /
`errBuf` with no-cross-leakage asserts; fail-first run before
implementation; assertion specificity on help/error substrings;
every locked decision paired with at least one failing test; line-
based equality (not `strings.Contains`) for any heading-level
assertion (SPEC-015 substring-trap addendum); ID-based (not
timestamp-based) distinctness for any freshness or ordering
tie-break (SPEC-017 freshness-assertion addendum).

Goldens reuse a single fixture so all renderer choices anchor to
the same canonical entries. Aggregate tests use a similar but
distinct fixture (kept smaller and focused on ordering rules).

### Shared renderer fixture (used by tests 5, 6, 7, 9)

```go
// 5 entries spanning 3 projects + (no project), with chrono
// ordering chosen to exercise:
//   * within-alpha chrono-ASC: 1 (T1), 4 (T4)  [IDs are NOT
//     monotonic with timestamps so ID tie-break is testable]
//   * (no project) forced last regardless of count
//   * one type with multiple counts, one with single count
//
// Fixture timestamps are within the 7-day rolling window from
// fixedNow so every entry passes the --range week filter when
// the CLI test injects fixedNow as the cutoff anchor.
var fixture = []storage.Entry{
    {
        ID: 1, Title: "alpha-old",
        Description: "old alpha",  // NOT rendered in summary
        Tags: "auth", Project: "alpha", Type: "shipped",
        Impact: "did stuff",
        CreatedAt: time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC),
        UpdatedAt: time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC),
    },
    {
        ID: 2, Title: "beta-mid",
        Project: "beta", Type: "learned",
        CreatedAt: time.Date(2026, 4, 21, 10, 0, 0, 0, time.UTC),
        UpdatedAt: time.Date(2026, 4, 21, 10, 0, 0, 0, time.UTC),
    },
    {
        ID: 3, Title: "unbound-mid",
        Type: "shipped",  // no project; (no project) group
        CreatedAt: time.Date(2026, 4, 22, 10, 0, 0, 0, time.UTC),
        UpdatedAt: time.Date(2026, 4, 22, 10, 0, 0, 0, time.UTC),
    },
    {
        ID: 4, Title: "alpha-new",
        Project: "alpha", Type: "shipped",
        CreatedAt: time.Date(2026, 4, 23, 10, 0, 0, 0, time.UTC),
        UpdatedAt: time.Date(2026, 4, 23, 10, 0, 0, 0, time.UTC),
    },
    {
        ID: 5, Title: "gamma-only",
        Project: "gamma", Type: "fixed",
        CreatedAt: time.Date(2026, 4, 24, 10, 0, 0, 0, time.UTC),
        UpdatedAt: time.Date(2026, 4, 24, 10, 0, 0, 0, time.UTC),
    },
}

var fixedNow = time.Date(2026, 4, 25, 12, 0, 0, 0, time.UTC)
```

Type counts on this fixture: `shipped: 3` (entries 1, 3, 4),
`learned: 1` (entry 2), `fixed: 1` (entry 5).

Project counts: `alpha: 2` (entries 1, 4), `beta: 1` (entry 2),
`gamma: 1` (entry 5), `(no project): 1` (entry 3).

Highlights ordering (alpha-ASC by project name, `(no project)`
last; chrono-ASC within group):
- alpha → [1: alpha-old (T1), 4: alpha-new (T4)]
- beta → [2: beta-mid (T2)]
- gamma → [5: gamma-only (T5)]
- (no project) → [3: unbound-mid (T3)]

### `internal/aggregate/aggregate_test.go` (new file — 4 tests)

Tests against literal `[]storage.Entry` slices. No cobra, no DB.
Pure stdlib + the new `internal/aggregate` package.

#### Test 1 — `TestByType_DESCByCountAlphaTiebreak`

Build a fixture with deliberate ties:
```go
input := []storage.Entry{
    {Type: "shipped"}, {Type: "shipped"}, {Type: "shipped"},
    {Type: "learned"}, {Type: "learned"},
    {Type: "fixed"},
}
```

Call `got := aggregate.ByType(input)`. Expected (DESC by count;
alpha-ASC tiebreak between `learned` and `shipped` doesn't apply
because they differ; `fixed` last):

```go
want := []aggregate.TypeCount{
    {Type: "shipped", Count: 3},
    {Type: "learned", Count: 2},
    {Type: "fixed",   Count: 1},
}
```

Then build a second fixture with a tie:
```go
tieInput := []storage.Entry{
    {Type: "zebra"}, {Type: "zebra"},
    {Type: "alpha"}, {Type: "alpha"},
    {Type: "fixed"},
}
```
Expected: `alpha: 2` before `zebra: 2` (alphabetical-ASC tiebreak),
then `fixed: 1`.

Assertions: `reflect.DeepEqual(got, want)` for each. Pairs
locked decision 2 (aggregate is the data-layer source) and
reuses DEC-013's count-ordering rule.

#### Test 2 — `TestByProject_NoProjectKeyForcedLast`

Build a fixture where `(no project)` has the HIGHEST count (the
adversarial case):
```go
input := []storage.Entry{
    {Project: ""}, {Project: ""}, {Project: ""},     // 3 of (no project)
    {Project: "alpha"}, {Project: "alpha"},           // 2 of alpha
    {Project: "beta"},                                // 1 of beta
}
```

Call `got := aggregate.ByProject(input)`. Expected: alpha first
(highest among real projects), beta second, `(no project)` LAST
despite having the highest count:

```go
want := []aggregate.ProjectCount{
    {Project: "alpha", Count: 2},
    {Project: "beta",  Count: 1},
    {Project: aggregate.NoProjectKey, Count: 3},
}
```

Assertion: `reflect.DeepEqual(got, want)`. This is the load-
bearing assertion for locked decision 3 (`(no project)` sentinel
+ forced-last rule). Pairs DEC-013's "(no project) last
regardless of count" precedent.

#### Test 3 — `TestGroupForHighlights_ChronoASCWithNoProjectLast`

Use the **shared fixture** from above. Call `got :=
aggregate.GroupForHighlights(fixture)`. Expected:

```go
want := []aggregate.ProjectHighlights{
    {Project: "alpha", Entries: []aggregate.EntryRef{
        {ID: 1, Title: "alpha-old"},   // T1 (2026-04-20)
        {ID: 4, Title: "alpha-new"},   // T4 (2026-04-23)
    }},
    {Project: "beta", Entries: []aggregate.EntryRef{
        {ID: 2, Title: "beta-mid"},
    }},
    {Project: "gamma", Entries: []aggregate.EntryRef{
        {ID: 5, Title: "gamma-only"},
    }},
    {Project: aggregate.NoProjectKey, Entries: []aggregate.EntryRef{
        {ID: 3, Title: "unbound-mid"},
    }},
}
```

Assertions:
1. `reflect.DeepEqual(got, want)` — full structural equality.
2. Walk `got[0].Entries` and verify
   `got[0].Entries[0].CreatedAt.Before(got[0].Entries[1].CreatedAt)`
   — wait, `EntryRef` carries only `ID` and `Title` (NOT
   `CreatedAt`). The chrono-ASC rule is enforced INSIDE the
   aggregate function; the test verifies via the `ID` ordering
   alone (1 before 4 within alpha proves the sort happened
   because we constructed timestamps so that `T1 < T4`).

   For an explicit chrono-ASC-with-ID-tiebreak test, add a
   second fixture where two entries in the same project share
   `CreatedAt` but have different IDs:

   ```go
   tieFixture := []storage.Entry{
       {ID: 99, Title: "later-id-same-time", Project: "alpha",
        CreatedAt: time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC)},
       {ID: 7,  Title: "earlier-id-same-time", Project: "alpha",
        CreatedAt: time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC)},
   }
   ```

   Expected: `[{ID: 7, ...}, {ID: 99, ...}]` (lower ID first
   when timestamps tie). Asserts the AGENTS.md §9 SPEC-002
   monotonic-tiebreak rule on the in-memory sort.

3. Verify `EntryRef` carries no `CreatedAt`, no `Description`
   field — by structure (the test imports `aggregate.EntryRef`
   and constructs literals with only `ID` and `Title`; if the
   struct grows new fields, this test breaks at compile time
   and that's an explicit reminder to revisit).

Pairs locked decision 2 (data layer correctness) and decision
3 (`(no project)` last) and the AGENTS.md §9 ID-tiebreak
discipline.

#### Test 4 — `TestAggregate_EmptyInputReturnsNonNilEmptySlice`

Call each of `ByType(nil)`, `ByProject(nil)`,
`GroupForHighlights(nil)` AND each of `ByType([]storage.Entry{})`,
`ByProject([]storage.Entry{})`,
`GroupForHighlights([]storage.Entry{})`. For each:

1. Assert `got != nil` (non-nil slice — important so JSON
   marshaling renders `[]` not `null`).
2. Assert `len(got) == 0`.

Pairs locked decision 1 part (4) (empty-state arrays are `[]`
in JSON, never `null`) on the data-layer side. The renderer-
side empty-state test (#7) verifies the corresponding JSON
output literally renders `[]`.

### `internal/export/summary_test.go` (new file — 5 tests)

Tests against the shared fixture + explicit `SummaryOptions`. No
cobra, no DB.

#### Test 5 — `TestToSummaryMarkdown_DEC014FullDocumentGolden` (LOAD-BEARING — write FIRST)

Build `opts := SummaryOptions{Scope: "week", Filters: "(none)",
FiltersJSON: nil, Now: fixedNow}` (`FiltersJSON` is unused on the
markdown path; either nil or empty map works). Call `got, err :=
ToSummaryMarkdown(fixture, opts)`. Assert `err == nil` and
`bytes.Equal(got, []byte(want))` where `want` is the literal
string below. If this fails, DEC-014 has been violated — fix
code, not test.

Expected output (byte-exact, no trailing newline — the CLI layer
adds one via `fmt.Fprintln`):

```
# Bragfile Summary

Generated: 2026-04-25T12:00:00Z
Scope: week
Filters: (none)

## Summary

**By type**
- shipped: 3
- fixed: 1
- learned: 1

**By project**
- alpha: 2
- beta: 1
- gamma: 1
- (no project): 1

## Highlights

### alpha

- 1: alpha-old
- 4: alpha-new

### beta

- 2: beta-mid

### gamma

- 5: gamma-only

### (no project)

- 3: unbound-mid
```

Notes on the layout:
- `## Summary` is the heading for the summary block (counts).
- `## Highlights` is a single wrapper heading; each project is a
  level-3 heading underneath it. Reusing DEC-013's `## <project>`
  level wouldn't compose cleanly with the level-2 `Summary`
  heading. (DEC-014 documents this; it's a deliberate divergence
  from DEC-013's heading-level choice because summary's structure
  is provenance + summary + highlights, all at the same
  document-level depth; DEC-013 was just provenance + summary +
  groups-as-second-level.)
- Counts ordering DESC by count + alpha-ASC tiebreak per
  DEC-013's precedent: `shipped: 3` > `fixed: 1` (alpha) =
  `learned: 1`. The tiebreak puts `fixed` before `learned`.
- `(no project)` always last in `**By project**` and in the
  highlights group ordering.
- Within-group bullets are `<id>: <title>` per the stage Design
  Notes (titles + IDs only, no descriptions).
- Within-group ordering is chrono-ASC per locked decision 3 +
  the highlights test #3.

Assertions:
1. `bytes.Equal(got, []byte(want))` — single exact-byte compare.
2. On failure, print both `want` and `got` for diffability (copy
   the SPEC-014 / SPEC-015 helper pattern).
3. Line-based assertion on heading levels per AGENTS.md §9
   substring-trap addendum (SPEC-015 lesson): split `got` into
   lines and assert `lines[0] == "# Bragfile Summary"` and that
   `"## Summary"` appears as a standalone line (`ln == "## Summary"`)
   not just as a substring of `"### Summary"` somewhere.

#### Test 6 — `TestToSummaryJSON_DEC014ShapeGolden` (LOAD-BEARING — write SECOND)

Build `opts := SummaryOptions{Scope: "week", Filters: "(none)",
FiltersJSON: map[string]string{}, Now: fixedNow}`. Call `got, err :=
ToSummaryJSON(fixture, opts)`. Assert `err == nil` and
`bytes.Equal(got, []byte(want))` where `want` is the literal JSON
below.

Expected output (byte-exact, no trailing newline):

```json
{
  "generated_at": "2026-04-25T12:00:00Z",
  "scope": "week",
  "filters": {},
  "counts_by_type": {
    "fixed": 1,
    "learned": 1,
    "shipped": 3
  },
  "counts_by_project": {
    "(no project)": 1,
    "alpha": 2,
    "beta": 1,
    "gamma": 1
  },
  "highlights": [
    {
      "project": "alpha",
      "entries": [
        {
          "id": 1,
          "title": "alpha-old"
        },
        {
          "id": 4,
          "title": "alpha-new"
        }
      ]
    },
    {
      "project": "beta",
      "entries": [
        {
          "id": 2,
          "title": "beta-mid"
        }
      ]
    },
    {
      "project": "gamma",
      "entries": [
        {
          "id": 5,
          "title": "gamma-only"
        }
      ]
    },
    {
      "project": "(no project)",
      "entries": [
        {
          "id": 3,
          "title": "unbound-mid"
        }
      ]
    }
  ]
}
```

Notes on the JSON shape:
- `counts_by_type` and `counts_by_project` are JSON objects (not
  arrays). Go's `encoding/json` sorts object keys alphabetical-
  ASC when marshaling a `map[string]int` — that's why
  `"(no project)": 1` lands first in `counts_by_project`
  (parenthesis sorts before letters in ASCII), `"alpha": 2`
  next, etc. This is DETERMINISTIC across runs (`encoding/json`
  has emitted sorted map keys since Go 1.12) but does NOT match
  the markdown side's DESC-by-count ordering. JSON consumers
  re-sort if they care; markdown is the human view. This
  asymmetry is locked in DEC-014 and called out explicitly so
  build/verify don't revisit it.
- `highlights` is a JSON array of `{project, entries}` objects
  in DEC-014-locked order (alpha-ASC by project name, `(no
  project)` last). Within `entries`, chrono-ASC enforced by the
  aggregate layer.
- `filters: {}` is an empty object (the `FiltersJSON` map was
  empty); when populated, e.g. `{"project": "platform"}` per
  test #9. Always object-shaped, never null — locked in
  decision 1 part (4).
- `id` renders as JSON number; `title` as string.
- Indent=2 per DEC-014 choice (5).

Assertions:
1. `bytes.Equal(got, []byte(want))` — single exact-byte compare.
2. Parse `got` via `json.Decoder` and verify object's top-level
   keys appear in struct-tag declaration order (`generated_at`,
   `scope`, `filters`, `counts_by_type`, `counts_by_project`,
   `highlights`) — Go preserves struct-tag order in
   `MarshalIndent`. This is the load-bearing key-order
   assertion that DEC-014 rests on.

#### Test 7 — `TestToSummary_EmptyEntriesEmitsProvenanceOnly`

Subtests `markdown` and `json`:

- `markdown`: Render `ToSummaryMarkdown([]storage.Entry{},
  SummaryOptions{Scope: "week", Filters: "(none)", Now:
  fixedNow})`. Expected bytes (byte-exact):

  ```
  # Bragfile Summary

  Generated: 2026-04-25T12:00:00Z
  Scope: week
  Filters: (none)
  ```

  **Last byte is `)`** (closing paren of `(none)`, trailing
  newline stripped). No `## Summary` heading. No `## Highlights`
  heading. Document ends after the `Filters:` line.

  Assertions:
  1. `bytes.Equal(got, []byte(want))`.
  2. Line-split + `for _, ln := range lines { if ln ==
     "## Summary" { t.Fatal("Summary heading present in empty
     output") } }` — line-based, not `strings.Contains`, per
     SPEC-015 substring-trap addendum.

- `json`: Render `ToSummaryJSON([]storage.Entry{},
  SummaryOptions{Scope: "week", FiltersJSON: map[string]string{},
  Now: fixedNow})`. Expected bytes:

  ```json
  {
    "generated_at": "2026-04-25T12:00:00Z",
    "scope": "week",
    "filters": {},
    "counts_by_type": {},
    "counts_by_project": {},
    "highlights": []
  }
  ```

  Assertions:
  1. `bytes.Equal(got, []byte(want))`.
  2. Parse via `json.Unmarshal` into `map[string]any` and
     assert `m["highlights"]` is a `[]any` with `len == 0`
     (NOT `nil`, NOT a JSON `null`). Locks the empty-state
     "non-nil empty slice" rule from decision 1 part (4).

Pairs locked decision 1 part (4) (empty-state shape) on both
output sides.

#### Test 8 — `TestToSummaryJSON_FiltersEchoShape`

Subtests `none` and `populated`:

- `none`: opts.FiltersJSON is `map[string]string{}`. Already
  covered by test #6's `"filters": {}`. Rather than duplicating,
  this subtest renders with `FiltersJSON: nil` and asserts the
  output ALSO contains `"filters": {}` (nil map → `{}` not
  `null`). Locks the nil-map handling.

- `populated`: opts.FiltersJSON is
  `map[string]string{"project": "platform", "tag": "auth"}`.
  Render and parse the result. Assertions:
  1. The output contains the substring `"filters": {`.
  2. Parse via `json.Unmarshal` into `struct {Filters
     map[string]string}` and assert
     `got.Filters["project"] == "platform"` and
     `got.Filters["tag"] == "auth"`.
  3. The literal byte-substring of the filters block (`"filters":
     {\n    "project": "platform",\n    "tag": "auth"\n  }`)
     is asserted via line-walking, locking key-order =
     alphabetical-ASC (Go's map sort: `project` after `tag`?
     no, alphabetical: `project` (p) before `tag` (t)). The
     concrete expected bytes are:

     ```
       "filters": {
         "project": "platform",
         "tag": "auth"
       },
     ```

     Asserted via `bytes.Contains(got, []byte(filtersBlock))`.

Pairs locked decision 1 part (2) (top-level flat keys, filters
as object) on the populated case.

#### Test 9 — `TestToSummaryMarkdown_FiltersLineFormat`

Subtests `none` and `echoed_flags`:

- `none`: render with `Filters: "(none)"`. Assert output line
  `Filters: (none)` appears (line-based equality, not
  substring).
- `echoed_flags`: render with `Filters: "--project platform
  --tag auth"`. Assert output contains the line `Filters:
  --project platform --tag auth` (line-based equality). The
  renderer treats `Filters` as opaque — CLI layer assembles.

Pairs locked decision 1 part (3) (provenance markdown
convention) on the markdown side.

(That's 5 tests in `summary_test.go`: 5, 6, 7, 8, 9.)

### `internal/cli/summary_test.go` (new file — 5 tests)

Reuse `seedListEntry` + `runListCmd` patterns from
`internal/cli/list_test.go` (same package, so direct reuse). Use
`t.TempDir()` for DB paths; never touch `~/.bragfile`.

#### Test 10 — `TestSummaryCmd_RangeRequiredIsUserError`

Run `brag summary` with no flags. Assertions:
1. `err != nil`; `errors.Is(err, ErrUser) == true`.
2. `outBuf.String() == ""`; `errBuf.Len() == 0` (cobra-RunE
   error → no leakage to outBuf or errBuf inside the test
   harness; cobra writes via `cmd.SetErr` which we capture but
   the `RunE` returning the error doesn't write to it; main.go
   handles it).
3. The error message (`err.Error()`) contains `--range` AND
   `week` AND `month` (each as distinctive needles).

Pairs locked decision 4 (range required, no default).

#### Test 11 — `TestSummaryCmd_RangeUnknownValueIsUserError`

Run `brag summary --range yearly`. Assertions:
1. `err != nil`; `errors.Is(err, ErrUser) == true`.
2. `outBuf.String() == ""`.
3. Error message contains `yearly` AND `week` AND `month`.

Pairs locked decision 4 (range validation rejects unknown
values).

#### Test 12 — `TestSummaryCmd_FormatJSON_RangeWeekAndFiltersCompose`

Seed 4 entries via `seedListEntry`:
- `{Title: "in-window-platform-auth", Project: "platform",
  Tags: "auth", Type: "shipped"}` — within 7d, all filters
  match.
- `{Title: "in-window-other-project", Project: "growth",
  Tags: "auth", Type: "shipped"}` — within 7d, project differs.
- `{Title: "in-window-other-tag", Project: "platform",
  Tags: "perf", Type: "shipped"}` — within 7d, tag differs.
- `{Title: "in-window-other-type", Project: "platform",
  Tags: "auth", Type: "learned"}` — within 7d, type differs.

(All within 7d so the range filter doesn't accidentally exclude
them; the test isolates filter composition.)

Run `brag summary --range week --format json --tag auth
--project platform --type shipped`.

Assertions:
1. `err == nil`; `errBuf.Len() == 0`.
2. Parse stdout as JSON. Assert top-level keys are exactly
   (any-order) `generated_at`, `scope`, `filters`,
   `counts_by_type`, `counts_by_project`, `highlights`.
3. Assert `scope == "week"`.
4. Assert `filters` is an object containing exactly `{"tag":
   "auth", "project": "platform", "type": "shipped"}`.
5. Assert `counts_by_type` is `{"shipped": 1}` (only one
   entry passed all filters).
6. Assert `counts_by_project` is `{"platform": 1}`.
7. Assert `highlights` is an array of one group object:
   `[{"project": "platform", "entries": [{"id": <ID>,
   "title": "in-window-platform-auth"}]}]`.
8. Assert stdout ends with `\n` (cli's `fmt.Fprintln` newline).

Pairs locked decisions 5, 6 (filter flag composition + JSON
format on summary).

#### Test 13 — `TestSummaryCmd_ScopeFieldAndMarkdownDefault`

CLI plumbing test. The build does NOT backdate `created_at`
to discriminate `--range week` vs `--range month` end-to-end
— backdating would require either a public
`Store.SetCreatedAtForTesting` method (rejected per the
Locked design decisions → Rejected alternatives subsection)
or raw SQL from `internal/cli/` test code (violates
`no-sql-in-cli-layer`). The arithmetic itself is locked by
Test #14 below; this test only verifies the CLI plumbs the
`--range` flag through to the renderer's `Scope:` line and
that markdown is the default format.

Seed 2 fresh entries via `seedListEntry`:
- `{Title: "alpha-fresh", Project: "alpha", Type: "shipped"}`
- `{Title: "beta-fresh", Project: "beta", Type: "learned"}`

Both are created with `time.Now().UTC()` (within both
windows). Run `brag summary --range week`. Assertions:
1. `err == nil`; `errBuf.Len() == 0`.
2. Output starts with `# Bragfile Summary\n\n` (markdown is
   the default — proves locked decision 5).
3. Output contains line-based-equality `Scope: week` (`ln ==
   "Scope: week"`, not substring — substring-trap addendum).
4. Output contains line-based-equality `## Summary`, and the
   line `## Highlights`. Both via line-walk, not
   `strings.Contains`.

Run `brag summary --range month` against the same DB.
Assertions:
1. `err == nil`; output's `Scope:` line is line-based-equal
   to `Scope: month`.
2. Document structure (`## Summary` heading line,
   `## Highlights` heading line) is identical to the
   `--range week` invocation — proves the envelope shape
   stays consistent across `--range` values; `scope` is the
   only differentiator at this layer.

Pairs locked decisions 1 part (6) (CLI plumbs `--range`
through to the `Scope:` line) and 5 (markdown default).

#### Test 14 — `TestRangeCutoff_WeekMonthArithmeticAndErrors`

Pure-function unit test on the cutoff helper. Lives in
`internal/cli/summary_test.go` alongside the other CLI tests
but does not use cobra, the store, or any test fixture —
direct calls into the helper with a fixed `now` parameter.

The CLI's range arithmetic is extracted into `func
rangeCutoff(rangeFlag string, now time.Time) (time.Time,
error)` in `internal/cli/summary.go` (see Notes for the
Implementer for the sketch). Assertions:

1. `rangeCutoff("week", fixedNow)` returns
   `fixedNow.AddDate(0, 0, -7)` exactly (`got.Equal(want) ==
   true`); `err == nil`.
2. `rangeCutoff("month", fixedNow)` returns
   `fixedNow.AddDate(0, 0, -30)` exactly; `err == nil`.
3. `rangeCutoff("yearly", fixedNow)` returns a zero
   `time.Time` and a non-nil error; `errors.Is(err, ErrUser)
   == true`; error message contains `yearly` AND `week` AND
   `month` (each as a distinctive needle per AGENTS.md §9).
4. `rangeCutoff("", fixedNow)` returns a zero `time.Time`
   and a non-nil user error; message contains `--range` AND
   `week` AND `month`.

Pairs locked decision 1 part (6) (rolling-window arithmetic
locked at the unit layer; deterministic across timezone
boundaries because `now` is supplied as input) AND the
"prescribe pure helper" choice in Locked design decisions →
Rejected alternatives (build-time) §1.

(Note: Tests #10 and #11 separately exercise the same error
paths via end-to-end CLI invocation. The duplication is
deliberate: Tests #10/#11 prove the CLI surfaces these
errors; Test #14 proves the helper is the single source of
those messages, so a future refactor that, say, moves the
validation into a flag-parsing hook can be caught by Test
#14 alone. AGENTS.md §9 locked-decisions-need-tests
discipline reads this as "the contract is the helper; the
CLI is one consumer of it.")

#### Test 15 — `TestSummaryCmd_HelpShowsRangeAndFormat`

Run `brag summary --help` with separate buffers. Assertions:
1. `err == nil`; `errBuf.Len() == 0`.
2. `outBuf.String()` contains EACH of these distinctive needles
   (assertion-specificity per AGENTS.md §9 SPEC-005 lesson —
   help text contents, not generic words like `summary` that
   cobra auto-renders): `--range`, `week`, `month`, `--format`,
   `markdown`, `json`, `--tag`, `--project`, `--type`.

Pairs locked decisions 4, 5, 6 on the help side.

### Test count summary

**15 tests across 3 new files** (load-bearing goldens written
FIRST):

- `internal/aggregate/aggregate_test.go` (new) — 4 tests
  (#1–#4).
- `internal/export/summary_test.go` (new) — 5 tests (#5–#9;
  #5 is the markdown golden written FIRST, #6 is the JSON
  golden written SECOND).
- `internal/cli/summary_test.go` (new) — 6 tests (#10–#15;
  #14 is `TestRangeCutoff_*`, the pure-helper unit test
  added when the build-time rejected alternative
  `Store.SetCreatedAtForTesting` was rejected — see Locked
  design decisions → Rejected alternatives (build-time)
  §1).

Plus regression locks (existing tests continue passing without
modification):
- All `internal/cli/list_test.go` tests — `seedListEntry`
  helper reused; signature stable.
- All `internal/cli/export_test.go` tests — disjoint surface.
- All `internal/export/json_test.go` and
  `internal/export/markdown_test.go` tests — sibling files,
  no cross-coupling.
- All `internal/storage/*_test.go` tests — read-only path.

## Implementation Context

*Read this section and the files it points to before starting the
build cycle. It is the equivalent of a handoff document, folded
into the spec since there is no separate receiving agent.*

### Decisions that apply

- **`DEC-014`** (emitted in this spec) — six locked rule-based-
  output choices. `internal/export/summary.go` implements them.
  Every deviation from the goldens in
  `TestToSummaryMarkdown_DEC014FullDocumentGolden` /
  `TestToSummaryJSON_DEC014ShapeGolden` is a DEC-014 violation —
  stop and raise a question before "fixing" the test.
- **`DEC-013`** — markdown export shape. DEC-014 reuses DEC-013's
  provenance + summary-block conventions for the markdown half;
  the renderer's count-ordering rule (DESC by count, alpha-ASC
  tiebreak, `(no project)` last) is INHERITED, not redefined.
- **`DEC-011`** — JSON output shape (naked array). DEC-014
  EXPLICITLY DIVERGES from DEC-011's array-shape because
  aggregations carry metadata that doesn't fit a per-entry row.
  Both DECs cross-reference each other.
- **`DEC-004`** — tags comma-joined TEXT. SPEC-018 doesn't
  aggregate by tag (filter only), so the rendering doesn't see
  tag normalization. If a future SPEC-019/020 aggregates by
  tag, DEC-004 still holds: split on comma at the aggregate
  layer, render verbatim.
- **`DEC-006`** — cobra framework. New `brag summary` subcommand
  follows the same pattern as every other command:
  `NewSummaryCmd() *cobra.Command` constructor in
  `internal/cli/summary.go`; unexported `runSummary` RunE
  handler; flags declared in the constructor.
- **`DEC-007`** — required-flag validation in `RunE`. `--range`
  is required; `--format` is required-with-default; both go
  through `UserErrorf` on missing/empty/unknown values. No
  `MarkFlagRequired`.

### Constraints that apply

For `internal/cli/**`, `internal/export/**`, `internal/aggregate/**`,
`docs/**`, `README.md`, `cmd/brag/main.go`, `decisions/**`:

- `no-sql-in-cli-layer` — blocking. `summary.go` (cli) and the
  new `internal/aggregate` package must not import
  `database/sql` or any SQL driver. They work with
  `[]storage.Entry` returned by `Store.List`. `internal/export/
  summary.go` is also SQL-free (siblings of json.go +
  markdown.go are already SQL-free).
- `stdout-is-for-data-stderr-is-for-humans` — blocking.
  Markdown/JSON bodies go through stdout. Errors go through
  stderr via main.go's `brag: <message>\n` wrapper. Every
  happy-path test asserts `errBuf.Len() == 0`.
- `errors-wrap-with-context` — warning. Storage / rendering
  errors from `runSummary` wrap with `fmt.Errorf("...: %w",
  err)`.
- `test-before-implementation` — blocking. Write all 15 tests
  first (load-bearing goldens FIRST: #5 markdown, then #6
  JSON), run `go test ./internal/aggregate ./internal/export
  ./internal/cli` (all three packages), confirm every new test
  fails for the expected reason (missing package, missing
  symbol, missing flag, missing command — NOT a compilation
  error unrelated to the spec), THEN implement.
- `one-spec-per-pr` — blocking. Branch
  `feat/spec-018-brag-summary-aggregate-and-shape-dec`. Diff
  touches only the files in Outputs.
- `no-new-top-level-deps-without-decision` — not triggered;
  stdlib only.
- `timestamps-in-utc-rfc3339` — applies indirectly: rendering
  uses `time.UTC().Format(time.RFC3339)`; the storage layer
  already enforces this for `created_at`/`updated_at`.

### AGENTS.md lessons that apply

- **§9 separate `outBuf` / `errBuf`** (SPEC-001) — every cli
  test in `summary_test.go` asserts both buffers.
- **§9 fail-first** (SPEC-003) — confirm each of the 15 tests
  fails for the expected reason before implementation. The
  load-bearing goldens (#5, #6) are the most informative
  fail-first signals: they fail with "package
  `internal/export` has no symbol `SummaryOptions`" or similar
  — distinctive errors that signal the test wires up
  correctly.
- **§9 assertion specificity** (SPEC-005) — help and error
  tests assert on distinctive needles (`--range`, `week`,
  `month`, `--format`), not generic words like `summary` or
  `format` that cobra auto-renders.
- **§9 substring-trap addendum** (SPEC-015) — every markdown
  heading-level assertion in tests (#5, #7-markdown, #13) is
  line-based (`ln == "## Summary"`), NOT
  `strings.Contains(out, "## Summary")`. Critical for the
  empty-state test (#7-markdown): asserting "no `## Summary`
  heading appears" must be line-based to avoid a false-pass
  from `### Summary` somewhere or a substring match in some
  other heading.
- **§9 freshness/distinctness addendum** (SPEC-017) — the
  highlights chrono-ASC + ID-tiebreak rule (#3 second fixture)
  uses ID inequality (`ID: 7` vs `ID: 99` with same
  `CreatedAt`), NOT timestamp inequality, for the tie-break
  assertion. AUTOINCREMENT guarantees ID monotonicity;
  timestamp ties (RFC3339 second-precision) can occur in
  storage and would cause the sort to flake without a
  monotonic tiebreak.
- **§9 locked-decisions-need-tests** (SPEC-009) — nine locked
  decisions above plus three rejected-alternatives
  (build-time) entries; each locked decision paired with at
  least one of the 15 failing tests (or with the manual
  greps for the api-contract.md replacement). The new pure-
  helper test (#14, `TestRangeCutoff_*`) is the
  locked-decision pairing for the rejected alternative
  "extract arithmetic into pure helper" — without it,
  decision 1 part (6) would only have CLI-level coverage,
  which slips the substance to integration testing.
- **§9 premise audit — addition case** (SPEC-011) — grepped
  above; no tracked-collection count coupling that breaks.
  AddCommand list bumps from 7 to 8 with no test asserting on
  count.
- **§9 premise audit — status-change case** (SPEC-012) —
  grepped above; doc-level actions enumerated under Outputs
  (api-contract.md, tutorial.md, README.md, data-model.md,
  AGENTS.md §11), NOT discovered at build time. The
  api-contract.md placeholder + the tutorial "What's NOT
  there yet" row are the status-change hot spots.
- **§9 SPEC-014 / SPEC-015 ship lessons (load-bearing tests
  written first)** — the markdown golden (#5) is the FIRST
  test written in `summary_test.go`; the JSON golden (#6)
  follows. SPEC-014's reflection note ("write the byte-
  identical cross-path test FIRST") generalizes here: the
  shape-locking goldens are the most informative tests and
  deserve top billing in the file's reading order.

### Prior related work

- **SPEC-014** (shipped 2026-04-23) — direct structural
  precedent. Emitted DEC-011, created `internal/export`
  package, anchored the JSON shape for SPEC-017 to consume.
  SPEC-018 mirrors that structure: emit DEC-014, create
  `internal/aggregate` package, anchor the envelope shape for
  SPEC-019/020 to consume. SPEC-014's load-bearing
  byte-identical-cross-path test (`brag list --format json` ==
  `brag export --format json`) doesn't have a direct
  equivalent here because summary is a single command, not a
  pair, but the spirit (one shape, one helper, one golden) is
  preserved via tests #5 and #6.
- **SPEC-015** (shipped 2026-04-24) — markdown export shape +
  provenance/summary block convention. SPEC-018 reuses the
  `Generated:` (renamed from DEC-013's `Exported:`) /
  `Scope:` (new — DEC-013 didn't have one because export is
  always "lifetime") / `Filters:` lines. The injected `Now`
  time pattern (`MarkdownOptions.Now`) carries forward as
  `SummaryOptions.Now` verbatim. The `(no project)` last
  rule + DESC-by-count summary-block ordering both inherit.
- **SPEC-017** (shipped 2026-04-24) — `brag add --json`
  freshness/distinctness lesson: assert ID inequality, not
  timestamp inequality, for any ordering tie-break. Applies
  to test #3's same-timestamp-different-ID assertion.
- **SPEC-007** (shipped 2026-04-20) — `ListFilter` struct.
  `runSummary` populates it identically to `runList` /
  `runExport` (copy the `cmd.Flags().Changed(...)` block).
  No `--since` flag — `--range` computes the cutoff and sets
  `filter.Since`.

### Out of scope (for this spec specifically)

If any of these feels necessary during build, write a new spec
rather than expanding this one.

- **`brag review` / `brag stats`** — SPEC-019/020. Don't
  pre-build their helpers; let them land additively on
  `internal/aggregate`.
- **`--out <path>` flag on summary** — backlog deferral noted
  in stage Design Notes.
- **`--since` / arbitrary date ranges on summary** —
  `--range week|month` only.
- **`--group-by <field>` on summary** — different from
  `brag export --format markdown --group-by` backlog item;
  not in scope.
- **`--limit` on summary** — aggregation-of-aggregations
  doesn't compose with row-cap.
- **`--compact` / non-pretty JSON on summary** — inherits
  DEC-011's pretty-default; same backlog covers all
  JSON-emitting commands.
- **Calendar-week / calendar-month semantics** — rolling-
  window only; revisit if a user requests Mon–Sun or 1st–end.
- **Time-zone configuration for the cutoff** — UTC-only for
  MVP.
- **Aggregating by tag in summary** — counts are by-type
  and by-project only. Tags are filter input, not aggregation
  axis, in MVP.
- **`renderEntry` reuse in summary** — summary explicitly
  renders only `<id>: <title>` per entry (no metadata table,
  no description). The `RenderEntry` helper from
  `internal/export/markdown.go` is NOT used by summary.
- **Changes to `storage.Entry`, `storage.Store`,
  `storage.ListFilter`** — read-only; no storage-layer
  modifications. The rejected alternative (a public
  `Store.SetCreatedAtForTesting` method to backdate entries
  for the `--range` test) is documented under Notes for the
  Implementer; the prescribed path keeps the storage surface
  untouched and locks the date arithmetic via a pure helper
  unit test.
- **Changes to existing tests** — append-only; no edits to
  existing `_test.go` files (other than possibly adding a
  small reusable helper in `internal/cli/list_test.go` if
  `seedListEntry`'s signature needs widening — UNLIKELY).

## Notes for the Implementer

Gotchas, style preferences, reuse opportunities. Read after the
Implementation Context; these are the "how" details.

- **`internal/aggregate/aggregate.go` layout.** Small file.
  Single package, single source file. Sketch:

  ```go
  // Package aggregate computes structured statistics over
  // []storage.Entry. It is the data layer for the rule-based
  // commands brag summary (SPEC-018), brag review (SPEC-019),
  // and brag stats (SPEC-020). Rendering lives in
  // internal/export.
  package aggregate

  import (
      "sort"

      "github.com/jysf/bragfile000/internal/storage"
  )

  const NoProjectKey = "(no project)"

  type TypeCount struct {
      Type  string
      Count int
  }

  type ProjectCount struct {
      Project string
      Count   int
  }

  type EntryRef struct {
      ID    int64
      Title string
  }

  type ProjectHighlights struct {
      Project string
      Entries []EntryRef
  }

  // ByType returns entries grouped by Type, ordered DESC by
  // count with alphabetical-ASC tiebreak. Empty input returns
  // a non-nil empty slice (so JSON renders [] not null).
  func ByType(entries []storage.Entry) []TypeCount { ... }

  // ByProject is identical in shape to ByType, except entries
  // with empty-string Project are rendered under NoProjectKey
  // and forced LAST regardless of count (matches DEC-013's
  // (no project)-last convention).
  func ByProject(entries []storage.Entry) []ProjectCount { ... }

  // GroupForHighlights returns project groups in alpha-ASC
  // order with NoProjectKey forced last; within each group,
  // entries are sorted ASC by CreatedAt with ID as tie-break
  // (AGENTS.md §9 SPEC-002 monotonic-tiebreak rule).
  // EntryRef carries ID + Title only — descriptions are
  // intentionally elided per the stage's "skim before pasting"
  // goal.
  func GroupForHighlights(entries []storage.Entry) []ProjectHighlights { ... }
  ```

  Sort sketches:

  ```go
  // ByType: count up, then sort.
  func ByType(entries []storage.Entry) []TypeCount {
      m := make(map[string]int)
      for _, e := range entries {
          m[e.Type]++
      }
      out := make([]TypeCount, 0, len(m))
      for t, c := range m {
          out = append(out, TypeCount{Type: t, Count: c})
      }
      sort.Slice(out, func(i, j int) bool {
          if out[i].Count != out[j].Count {
              return out[i].Count > out[j].Count // DESC
          }
          return out[i].Type < out[j].Type // alpha-ASC
      })
      return out
  }
  ```

  ```go
  // ByProject: count up, sort, then re-position NoProjectKey.
  func ByProject(entries []storage.Entry) []ProjectCount {
      m := make(map[string]int)
      for _, e := range entries {
          key := e.Project
          if key == "" {
              key = NoProjectKey
          }
          m[key]++
      }
      out := make([]ProjectCount, 0, len(m))
      for p, c := range m {
          out = append(out, ProjectCount{Project: p, Count: c})
      }
      sort.Slice(out, func(i, j int) bool {
          // Force NoProjectKey last regardless of count.
          if out[i].Project == NoProjectKey {
              return false
          }
          if out[j].Project == NoProjectKey {
              return true
          }
          if out[i].Count != out[j].Count {
              return out[i].Count > out[j].Count
          }
          return out[i].Project < out[j].Project
      })
      return out
  }
  ```

  ```go
  // GroupForHighlights: bucket, sort within bucket, sort
  // buckets, force NoProjectKey-bucket last.
  func GroupForHighlights(entries []storage.Entry) []ProjectHighlights {
      buckets := make(map[string][]storage.Entry)
      for _, e := range entries {
          key := e.Project
          if key == "" {
              key = NoProjectKey
          }
          buckets[key] = append(buckets[key], e)
      }
      out := make([]ProjectHighlights, 0, len(buckets))
      for proj, group := range buckets {
          // chrono-ASC + ID tie-break.
          sorted := make([]storage.Entry, len(group))
          copy(sorted, group)
          sort.SliceStable(sorted, func(i, j int) bool {
              if !sorted[i].CreatedAt.Equal(sorted[j].CreatedAt) {
                  return sorted[i].CreatedAt.Before(sorted[j].CreatedAt)
              }
              return sorted[i].ID < sorted[j].ID
          })
          refs := make([]EntryRef, 0, len(sorted))
          for _, e := range sorted {
              refs = append(refs, EntryRef{ID: e.ID, Title: e.Title})
          }
          out = append(out, ProjectHighlights{Project: proj, Entries: refs})
      }
      sort.Slice(out, func(i, j int) bool {
          if out[i].Project == NoProjectKey {
              return false
          }
          if out[j].Project == NoProjectKey {
              return true
          }
          return out[i].Project < out[j].Project
      })
      return out
  }
  ```

- **`internal/export/summary.go` layout.** Sibling to
  `markdown.go` and `json.go`. Imports `internal/aggregate`.
  Sketch:

  ```go
  package export

  import (
      "bytes"
      "encoding/json"
      "fmt"
      "sort"
      "time"

      "github.com/jysf/bragfile000/internal/aggregate"
      "github.com/jysf/bragfile000/internal/storage"
  )

  // SummaryOptions controls the rule-based summary digest.
  // Filters is the pre-formatted markdown line ("(none)" or
  // an echoed flag string); FiltersJSON is the object that the
  // JSON envelope renders (empty map → "{}", populated → an
  // object with alphabetically-sorted keys per Go's encoding/
  // json map handling). Now is injected for deterministic
  // goldens — mirrors MarkdownOptions.Now.
  type SummaryOptions struct {
      Scope       string             // "week" | "month"
      Filters     string             // markdown line value
      FiltersJSON map[string]string  // JSON envelope value
      Now         time.Time
  }

  func ToSummaryMarkdown(entries []storage.Entry, opts SummaryOptions) ([]byte, error) {
      var buf bytes.Buffer
      fmt.Fprintln(&buf, "# Bragfile Summary")
      fmt.Fprintln(&buf)
      fmt.Fprintf(&buf, "Generated: %s\n", opts.Now.UTC().Format(time.RFC3339))
      fmt.Fprintf(&buf, "Scope: %s\n", opts.Scope)
      fmt.Fprintf(&buf, "Filters: %s\n", opts.Filters)
      if len(entries) == 0 {
          return trimTrailingNewline(buf.Bytes()), nil
      }
      fmt.Fprintln(&buf)
      fmt.Fprintln(&buf, "## Summary")
      fmt.Fprintln(&buf)
      fmt.Fprintln(&buf, "**By type**")
      for _, tc := range aggregate.ByType(entries) {
          fmt.Fprintf(&buf, "- %s: %d\n", tc.Type, tc.Count)
      }
      fmt.Fprintln(&buf)
      fmt.Fprintln(&buf, "**By project**")
      for _, pc := range aggregate.ByProject(entries) {
          fmt.Fprintf(&buf, "- %s: %d\n", pc.Project, pc.Count)
      }
      fmt.Fprintln(&buf)
      fmt.Fprintln(&buf, "## Highlights")
      for _, group := range aggregate.GroupForHighlights(entries) {
          fmt.Fprintln(&buf)
          fmt.Fprintf(&buf, "### %s\n", group.Project)
          fmt.Fprintln(&buf)
          for _, ref := range group.Entries {
              fmt.Fprintf(&buf, "- %d: %s\n", ref.ID, ref.Title)
          }
      }
      return trimTrailingNewline(buf.Bytes()), nil
  }
  ```

  Reuse `trimTrailingNewline` from `markdown.go` (already
  package-local; if not exported, that's fine — same package).

  JSON renderer sketch:

  ```go
  type summaryEnvelope struct {
      GeneratedAt     string                          `json:"generated_at"`
      Scope           string                          `json:"scope"`
      Filters         map[string]string               `json:"filters"`
      CountsByType    map[string]int                  `json:"counts_by_type"`
      CountsByProject map[string]int                  `json:"counts_by_project"`
      Highlights      []highlightGroup                `json:"highlights"`
  }

  type highlightGroup struct {
      Project string             `json:"project"`
      Entries []highlightEntry   `json:"entries"`
  }

  type highlightEntry struct {
      ID    int64  `json:"id"`
      Title string `json:"title"`
  }

  func ToSummaryJSON(entries []storage.Entry, opts SummaryOptions) ([]byte, error) {
      env := summaryEnvelope{
          GeneratedAt:     opts.Now.UTC().Format(time.RFC3339),
          Scope:           opts.Scope,
          Filters:         opts.FiltersJSON,
          CountsByType:    map[string]int{},
          CountsByProject: map[string]int{},
          Highlights:      []highlightGroup{},
      }
      if env.Filters == nil {
          env.Filters = map[string]string{}
      }
      for _, tc := range aggregate.ByType(entries) {
          env.CountsByType[tc.Type] = tc.Count
      }
      for _, pc := range aggregate.ByProject(entries) {
          env.CountsByProject[pc.Project] = pc.Count
      }
      for _, group := range aggregate.GroupForHighlights(entries) {
          hg := highlightGroup{
              Project: group.Project,
              Entries: make([]highlightEntry, 0, len(group.Entries)),
          }
          for _, ref := range group.Entries {
              hg.Entries = append(hg.Entries, highlightEntry{
                  ID: ref.ID, Title: ref.Title,
              })
          }
          env.Highlights = append(env.Highlights, hg)
      }
      return json.MarshalIndent(env, "", "  ")
  }
  ```

  Note: `env.CountsByType` and `env.CountsByProject` are
  `map[string]int` — Go's `encoding/json` sorts map keys
  alphabetically when marshaling, which is deterministic but
  different from the markdown side's DESC-by-count ordering.
  This is the documented asymmetry per locked decision 1.
  `env.Highlights` is a slice (NOT a map) so its ordering is
  preserved verbatim from `aggregate.GroupForHighlights`.
  `env.Filters` defaults to `{}` when the input is nil — the
  conditional `if env.Filters == nil` is the load-bearing
  empty-state guard.

- **`internal/cli/summary.go` structure.** Mirror `export.go`.
  `NewSummaryCmd` constructor declares flags in a clear order
  (`--range` first, `--format` second, then `--tag` /
  `--project` / `--type` reusing list/export wording).
  `runSummary` validates `--range`, validates `--format`,
  computes the cutoff, builds `ListFilter`, opens the store,
  calls `Store.List`, renders, writes.

  Flag declaration sketch:

  ```go
  cmd.Flags().String("range", "", "time range (required; one of: week, month)")
  cmd.Flags().String("format", "markdown", "output format (one of: markdown, json)")
  cmd.Flags().String("tag", "", "filter to entries whose tags contain this token (comma-separated match)")
  cmd.Flags().String("project", "", "filter to entries with this project (exact match)")
  cmd.Flags().String("type", "", "filter to entries with this type (exact match)")
  ```

  Validation in `runSummary`:

  ```go
  rangeFlag, _ := cmd.Flags().GetString("range")
  if rangeFlag == "" {
      return UserErrorf("--range is required (accepted: week, month)")
  }
  var sinceCutoff time.Time
  switch rangeFlag {
  case "week":
      sinceCutoff = time.Now().UTC().AddDate(0, 0, -7)
  case "month":
      sinceCutoff = time.Now().UTC().AddDate(0, 0, -30)
  default:
      return UserErrorf("unknown --range value %q (accepted: week, month)", rangeFlag)
  }

  format, _ := cmd.Flags().GetString("format")
  if format != "markdown" && format != "json" {
      return UserErrorf("unknown --format value %q (accepted: markdown, json)", format)
  }
  ```

  Use `sinceCutoff` to populate `filter.Since` (do NOT call
  `ParseSince` — that's for user input like `7d`, `30d`; here
  we have a deterministic cutoff already).

  Filter assembly: copy the `cmd.Flags().Changed("tag")` etc.
  blocks from `runList` / `runExport` verbatim. Do NOT
  extract a shared helper; SPEC-019 will add a fourth caller
  and at that point a refactor is justified — three identical
  copies is the cheaper code path today (matches SPEC-014's
  Notes-for-the-Implementer guidance).

  Filters echo — PRESCRIBED PATH: inline a private helper
  `echoFiltersForSummary(cmd *cobra.Command) (markdownLine
  string, jsonObj map[string]string)` in
  `internal/cli/summary.go`. Returns BOTH representations in
  one pass. Iterates a fixed flag list `["tag", "project",
  "type"]` (summary's filter set, NOT export's broader set);
  for each `cmd.Flags().Changed(name)`, appends `--<name>
  <value>` to the markdown parts AND sets `jsonObj[name] =
  value`. Empty result → markdown line is `"(none)"`, JSON
  object is `map[string]string{}`. Sketch:

  ```go
  func echoFiltersForSummary(cmd *cobra.Command) (string, map[string]string) {
      jsonObj := map[string]string{}
      var parts []string
      for _, name := range []string{"tag", "project", "type"} {
          if !cmd.Flags().Changed(name) {
              continue
          }
          v, _ := cmd.Flags().GetString(name)
          jsonObj[name] = v
          parts = append(parts, fmt.Sprintf("--%s %s", name, v))
      }
      if len(parts) == 0 {
          return "(none)", jsonObj
      }
      return strings.Join(parts, " "), jsonObj
  }
  ```

  Render dispatch:

  ```go
  filtersMD, filtersJSON := echoFiltersForSummary(cmd)
  opts := export.SummaryOptions{
      Scope:       rangeFlag,
      Filters:     filtersMD,
      FiltersJSON: filtersJSON,
      Now:         time.Now().UTC(),
  }

  var body []byte
  switch format {
  case "markdown":
      body, err = export.ToSummaryMarkdown(entries, opts)
  case "json":
      body, err = export.ToSummaryJSON(entries, opts)
  }
  if err != nil {
      return fmt.Errorf("render summary: %w", err)
  }
  fmt.Fprintln(cmd.OutOrStdout(), string(body))
  return nil
  ```

  **Rejected alternative — reuse `export.go`'s `echoFilters`.**
  That helper iterates `["tag", "project", "type", "since",
  "limit"]` (export's broader set). Calling it from `summary`
  works in practice (`cmd.Flags().Changed("since")` returns
  false for an undeclared flag), but the caller has to KNOW
  that the helper happens to be a superset of summary's flags
  — coupling that future refactors can break silently. Inline
  is cheaper: clearer per-command filter set, no
  cross-command read-the-source-to-trust dance. The
  three-callers threshold for shared helpers is not met
  (export and summary have divergent filter sets); revisit
  if SPEC-019 / SPEC-020 add review/stats with the same
  three-flag set, at which point lifting to a shared helper
  is the right move.

- **`--range` cutoff is rolling, not calendar.** Locked in
  DEC-014 choice (6). Compute as `time.Now().UTC().AddDate(0,
  0, -7)` for week, `-30` for month. Document this in
  `runSummary`'s comment block and in the api-contract.md
  rewrite. Tests #12 and #13 verify the two ranges produce
  different filtered sets.

- **`Now` injection in tests vs production.** The renderer
  takes `Now` via `SummaryOptions`. The CLI passes
  `time.Now().UTC()`. Tests pass a fixed `time.Date(...)`. For
  the CLI tests (#12, #13) that DON'T inject a clock, the
  `Generated:` line is non-deterministic. Strategies:
    - Strip the `Generated:` line via regex before comparing.
    - Or assert the line MATCHES `^Generated: \d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`
      (line-based regex check) and skip byte-exact compare.
    - Or assert presence of the literal substring `Generated:`
      and skip the rest.
  Recommendation: use the "line matches RFC3339 regex" approach
  for both #12 and #13; gives the strongest assertion that
  isn't time-of-day-dependent.

- **Test #13: `--range` arithmetic + CLI plumbing — PRESCRIBED
  PATH (no backdating).** The build does NOT backdate any
  entry's `created_at`. Instead:

  1. **Extract the range arithmetic into a pure helper.** Add
     to `internal/cli/summary.go`:

     ```go
     // rangeCutoff returns the inclusive lower bound for the
     // --range filter. Pure function: deterministic given
     // (rangeFlag, now); takes `now` as input so callers can
     // inject a fixed time in tests.
     //
     // "week"  → now - 7 days
     // "month" → now - 30 days
     // anything else → UserErrorf naming the accepted values.
     // Empty string → UserErrorf naming the flag.
     func rangeCutoff(rangeFlag string, now time.Time) (time.Time, error) {
         switch rangeFlag {
         case "":
             return time.Time{}, UserErrorf("--range is required (accepted: week, month)")
         case "week":
             return now.AddDate(0, 0, -7), nil
         case "month":
             return now.AddDate(0, 0, -30), nil
         default:
             return time.Time{}, UserErrorf("unknown --range value %q (accepted: week, month)", rangeFlag)
         }
     }
     ```

     `runSummary` calls `rangeCutoff(rangeFlag, time.Now().UTC())`
     and uses the returned cutoff as `filter.Since`. The two
     `UserErrorf` paths there are the same paths tests #10 and
     #11 already exercise via end-to-end CLI invocation; the
     helper unifies the source of those messages.

  2. **Test the helper directly** under Test #13 subtest
     `range_cutoff_arithmetic` — pure function call with
     `fixedNow`, byte-exact `Equal` assertion on the returned
     time, error-path coverage for empty + unknown.

  3. **Test the CLI plumbing without backdating** under Test #13
     subtest `cli_scope_and_markdown_default` — seed two fresh
     entries (both within both windows), invoke `--range week`
     and `--range month`, assert the `Scope:` line differs
     while the rest of the document structure stays consistent.
     This proves the flag plumbs through to the renderer; the
     date arithmetic is already locked by the helper test.

  **Rejected alternative — `Store.SetCreatedAtForTesting`.** A
  public storage method that lets tests rewrite an entry's
  `created_at` would let the test exercise the date cutoff
  end-to-end (verify that an entry between 7 and 30 days old
  appears under `--range month` but not `--range week`).
  Rejected because: (a) it ships test-only surface in
  production code (the doc comment "intended only for tests"
  is a soft guarantee that the next refactor erodes); (b) the
  arithmetic is genuinely a pure function and deserves a pure
  test — going through DB seeding to prove it is end-to-end
  testing for testing's sake; (c) the CLI plumbing assertion
  is just as well covered by the `Scope:` line check on fresh
  entries, with no date math involved. The pure-helper path
  is cheaper to write, faster to run, and free of the
  test-only-surface-in-production smell. If a future spec
  needs backdated entries (e.g., for stats' streak / span
  semantics in SPEC-020), that spec earns its own decision
  on the helper question — DEC-014 / SPEC-018 do not
  pre-commit to the storage-layer surface.

- **fail-first run.** Before implementing, run:

  ```bash
  go test ./internal/aggregate ./internal/export ./internal/cli -run \
      "TestByType|TestByProject|TestGroupForHighlights|TestAggregate|TestToSummary|TestSummaryCmd|TestRangeCutoff"
  ```

  Expected: every one of the 15 new tests fails for the
  expected reason (undefined package `internal/aggregate`,
  undefined `SummaryOptions`, unknown command `summary`,
  etc.). If any passes unexpectedly, investigate.

- **Doc-sweep order.** Execute in this order to keep diffs
  reviewable:
  1. api-contract.md lines 251–261 REPLACE (status-change
     hot spot).
  2. api-contract.md References list ADD DEC-014.
  3. tutorial.md line 3 Scope blurb UPDATE.
  4. tutorial.md line ~453 "What's NOT there yet" table
     STRIKE the summary row.
  5. tutorial.md (optional) §4 ADD `### Weekly digest:
     brag summary` subsection.
  6. README.md line ~61 Scope blurb UPDATE.
  7. data-model.md References list ADD DEC-014.
  8. AGENTS.md §11 Domain Glossary — multi-edit pass:
     (a) UPDATE summary entry (line 250): STAGE-003 →
     STAGE-004;
     (b) UPDATE tap entry (line 251): STAGE-004 → STAGE-005
     (the cherry-pick on 2026-04-24 moved tap to STAGE-005;
     glossary was not updated then; SPEC-018's status-change
     premise audit catches this);
     (c) ADD aggregate entry (alphabetical placement);
     (d) ADD digest entry (alphabetical placement);
     (e) ADD review entry (alphabetical placement);
     (f) ADD stats entry (alphabetical placement).
     See the AGENTS.md §11 Domain Glossary updates section
     above for full content.

- **api-contract.md replacement template.** Suggested
  structure:

  ```markdown
  ### `brag summary --range week|month` (STAGE-004)

  ```
  brag summary --range week                          # last 7 UTC days, markdown
  brag summary --range month --format json           # last 30 UTC days, JSON envelope
  brag summary --range week --tag auth --project p   # compose filters
  ```

  Rule-based digest of the rolling time window. Output is a
  markdown document (default) or single-object JSON envelope
  (`--format json`) carrying:
  - **Provenance:** `Generated:` (RFC3339), `Scope:` (week|
    month), `Filters:` (echoed flags or `(none)`).
  - **Summary block:** counts by type and by project (DESC by
    count, alphabetical-ASC tiebreak; `(no project)` last in
    the by-project list).
  - **Highlights:** entry titles + IDs grouped by project,
    chronological-ASC within group; descriptions are
    intentionally elided for the "skim before pasting" goal.

  Flags:
  - `--range week|month` REQUIRED. `week` = last 7 UTC days
    from `time.Now()`; `month` = last 30 UTC days. Rolling
    window, NOT calendar week/month.
  - `--format markdown|json` defaults to `markdown`. JSON is
    a single-object envelope (NOT an array — diverges from
    DEC-011's list shape because aggregations carry
    metadata). Shape locked by [DEC-014](../decisions/DEC-014-rule-based-output-shape.md).
  - `--tag <token>`, `--project <name>`, `--type <name>`
    reuse `brag list`'s `ListFilter` semantics. No
    `--since`/`--limit`/`--out` on summary in MVP.
  - Output goes to stdout. Redirect with `>` if you want a
    file.

  Unknown `--range` or `--format` values exit 1 (user error).
  ```

- **AGENTS.md §11 Domain Glossary updates.** Two stale-stage
  fixes plus four new entries.

  **Fix line 250 (summary):** stale STAGE-003 → STAGE-004.

  Current:
  ```
  - **summary** — a rule-based (non-LLM) aggregation of entries grouped
    by project/type over a time range. STAGE-003.
  ```

  Replace with:
  ```
  - **summary** — a rule-based (non-LLM) aggregation of entries grouped
    by project/type over a rolling 7- or 30-day time window
    (`brag summary --range week|month`). STAGE-004.
  ```

  **Fix line 251 (tap):** stale STAGE-004 → STAGE-005. The tap
  was originally framed for STAGE-004 but moved to STAGE-005
  during the 2026-04-24 cherry-pick; the glossary line was
  not updated at the time. SPEC-018's status-change premise
  audit catches this.

  Current:
  ```
  - **tap** — a homebrew tap repo (`github.com/jysf/homebrew-bragfile`) hosting the `bragfile.rb` formula. Created in STAGE-004.
  ```

  Replace with:
  ```
  - **tap** — a homebrew tap repo (`github.com/jysf/homebrew-bragfile`) hosting the `bragfile.rb` formula. Created in STAGE-005.
  ```

  **Add four new entries** (placement: insert in alphabetical
  order — `aggregate` near the top alongside `brag`/`capture`,
  `digest` between `capture` and `export`, `review` between
  `migration` and `summary`, `stats` between `summary` and
  `tap`):

  ```
  - **aggregate** — the Go package `internal/aggregate/`
    (introduced in STAGE-004 by SPEC-018). Pure data layer
    that maps `[]storage.Entry → structured stats`
    (`ByType`, `ByProject`, `GroupForHighlights`, plus
    SPEC-019's grouping helpers and SPEC-020's `Streak` /
    `MostCommon` / `Span`). Rendering is the responsibility
    of `internal/export`; aggregate stays SQL-free and
    dependency-free.
  ```

  ```
  - **digest** — collective name for the rule-based commands
    (`brag summary`, `brag review`, `brag stats`) that emit a
    single-document aggregation per DEC-014's envelope
    shape. Distinct from `brag export`, which emits a
    multi-entry document. STAGE-004.
  ```

  ```
  - **review** — `brag review --week | --month`: prints
    recent entries grouped by project followed by three
    hard-coded reflection questions. Designed to be pasted
    into an external AI session for guided self-reflection.
    STAGE-004 (SPEC-019).
  ```

  ```
  - **stats** — `brag stats`: six lifetime aggregations
    (total entries, entries/week rolling average, current
    streak, longest streak, top-5 most-common tags, top-5
    most-common projects, corpus span). STAGE-004
    (SPEC-020).
  ```

  Note: the `review` and `stats` entries are added by
  SPEC-018 even though those commands ship in SPEC-019 /
  SPEC-020 — the glossary terms are STAGE-004-level concepts
  that DEC-014 references and that SPEC-018's spec body
  uses (out-of-scope guardrails, premise audit). Adding
  them here makes the glossary self-consistent across the
  three STAGE-004 specs from the moment SPEC-018 lands;
  SPEC-019 / SPEC-020 then don't need to backfill the
  glossary at ship time.

- **No helper extraction beyond what's listed.** Specifically:
  do NOT lift `echoFilters` from `export.go` to a shared file.
  `summary.go` inlines its own `echoFiltersForSummary` per the
  prescribed path above (returns both markdown line and JSON
  object in one pass; iterates summary's three-flag set, NOT
  export's broader five-flag set). Three callers with the
  same filter set is the threshold for a shared helper —
  we're at two callers with divergent filter sets today.
  Wait for SPEC-019 / SPEC-020 to either confirm or reset
  the threshold.

- **Branch:** `feat/spec-018-brag-summary-aggregate-and-shape-dec`.
  IMPORTANT: the design cycle does NOT create this branch.
  Design-cycle commits land on `main`; the build session
  creates the feature branch as its first step. (Three clean
  data points so far on this rhythm — SPEC-014/017 +
  STAGE-004 framing — vs. SPEC-015's single slip. Keep it
  three-and-going.)

---

## Build Completion

*Filled in at the end of the **build** cycle, before advancing to verify.*

- **Branch:** `feat/spec-018-brag-summary-aggregate-package-and-shape-dec`
- **PR (if applicable):** opened post-build (URL in commit footer / final assistant message)
- **All acceptance criteria met?** yes — 15/15 new tests green; existing suite unchanged; `gofmt -l .` empty; `go vet ./...` clean; `CGO_ENABLED=0 go build ./...` succeeds; `go test ./...` green; both load-bearing goldens (markdown + JSON) byte-match DEC-014.
- **New decisions emitted:**
  - none from build (DEC-014 was design-time, emitted by the design cycle alongside this spec; build did not introduce any new build-time DEC).
- **Deviations from spec:**
  - **None to the prescribed code path.** All locked decisions implemented as specified; both rejected build-time alternatives stayed rejected (no `Store.SetCreatedAtForTesting`, no `echoFilters` reuse).
  - **Two doc-sweep audit misses surfaced during build (NOT corrected — out-of-spec scope, flagged for verify):**
    - `AGENTS.md` line 67 (§3 Tech Stack → Distribution) still reads "homebrew tap … (arriving in STAGE-004)" — symmetric to the §11 line 251 fix the spec did enumerate. The spec's `tap.*STAGE-004` premise-audit grep would have hit both, but only §11 was prescribed under Outputs. Left untouched per "if the spec doesn't say it, don't do it."
    - `docs/architecture.md` line 24 still reads "export / summary STAGE-003" inside the mermaid diagram. The spec's `STAGE-003.*summary` audit grep was scoped to `docs/ README.md AGENTS.md` and would have hit this; the Outputs-side enumeration listed only `docs/api-contract.md`, `docs/tutorial.md`, `docs/data-model.md`, `README.md`, `AGENTS.md`. Left untouched for the same reason.
  - **One observed simplification vs. the spec's predicted importer set:** the spec's `internal/aggregate` audit predicted importers in `internal/export/summary.go` AND `internal/cli/summary.go`. The actual build only needs the export-side importer — the CLI talks to renderers (`export.ToSummaryMarkdown`, `export.ToSummaryJSON`) and never names `aggregate` itself. Cleaner seam than predicted; no behavior change.
- **Follow-up work identified:**
  - **Two-line follow-up doc fix (verify or a small chore PR):** sync the AGENTS.md §3 line 67 tap reference and the `docs/architecture.md` mermaid summary node to match the §11 / api-contract.md / tutorial.md / README.md updates from this spec. Either is below DEC threshold; both are pure status-text edits.
  - SPEC-019 / SPEC-020 already enumerated in the stage backlog; nothing new added by this build.

### Build-phase reflection (3 questions, short answers)

Process-focused: how did the build go? What friction did the spec create?

1. **What was unclear in the spec that slowed you down?**
   — Nothing material. The spec's literal code sketches for the aggregate sort logic, the renderer envelope, and the `rangeCutoff` helper meant build was almost transcription — including the load-bearing goldens, which made fail-first the most informative signal. The only minor friction was the spec's Notes-for-the-Implementer suggesting the test plumbing for the JSON key-order assertion in pseudo-form; I implemented a small `skipValue` helper in `summary_test.go` to walk the decoder properly, which the spec didn't fully sketch.

2. **Was there a constraint or decision that should have been listed but wasn't?**
   — The status-change premise audit (§9 SPEC-012 lesson) was correctly applied to the four enumerated docs but missed two adjacent hits the audit greps would have caught: `AGENTS.md` line 67 (§3 distribution mention of `tap STAGE-004`) and `docs/architecture.md` line 24 (mermaid label `summary STAGE-003`). Both are in files the audit greps targeted, but neither was enumerated in Outputs as a planned edit. The audit's "expected hits" list correctly named the §11 glossary entry as the tap hot spot but did not predict the §3 line 67 hit; similarly the `STAGE-003.*summary` audit's "expected hits" enumerated 5 files but missed `docs/architecture.md`. Both are spec defects (audit incomplete), not build defects. Build flagged them rather than silently expanding scope.

3. **If you did this task again, what would you do differently?**
   — Run the audit greps myself before the doc sweep — not just trust the spec's enumeration. The spec's audit blocks include the actual grep commands; running them in build would have surfaced both audit misses immediately and let me ask "is this in scope or out of scope?" up front instead of mid-sweep. The lesson generalizes: when a spec lists "expected hits" for a premise-audit grep, the build should re-run the grep and reconcile the actual hits against the enumeration, treating any delta as a question for the spec author rather than a unilateral expansion.

---

## Reflection (Ship)

*Appended during the **ship** cycle. Outcome-focused reflection, distinct
from the process-focused build reflection above.*

1. **What would I do differently next time?**
   — Run the audit greps during design, not just write them down. The spec enumerated two greps with expected-hit lists (`tap.*STAGE-004` scoped to AGENTS.md; `STAGE-003.*summary` scoped to docs/ + README.md + AGENTS.md) but did not actually execute them. Executing the first against the whole repo would have caught `AGENTS.md:67`; executing the second would have caught `docs/architecture.md:24`. Both would have landed in `## Outputs`, neither would have surfaced as a build-flagged miss. The discipline mirrors SPEC-012's ship lesson one level up: SPEC-012 said "grep when status changes"; SPEC-018 says "actually run the greps you enumerate." Spec size (2487 lines) was justified — investment paid back as zero build-time drift on locked decisions and three rejected build-time alternatives held firm. Size was not the problem; trusting the enumeration without verifying it was.

2. **Does any template, constraint, or decision need updating?**
   — Yes, two refinements at the three-data-point threshold that historically codified prior §9 lessons. Both land in this ship cycle as a separate `chore(AGENTS)` commit on this branch.
   (a) **§9 audit-grep cross-check, both sides.** One bullet, two clauses. *Design-side:* when you write a premise-audit grep, run it and reconcile actual hits against Outputs — enumeration without execution is aspirational. *Build-side:* before doc-sweep, re-run the spec's audit greps and treat any delta from Outputs as a question for the spec author, not a unilateral expansion of scope. SPEC-010/011/012 established the design-side premise-audit family (inversion / addition / status-change); SPEC-018 verify (2026-04-25) surfaced the symmetric build-side mirror that closes the loop.
   (b) **§12 (During design): decide at design time when decidable.** SPEC-018 was the first spec to enumerate a "Rejected alternatives (build-time)" subsection within Locked design decisions, and the build held all three (`Store.SetCreatedAtForTesting`, inline `--range` parsing, `echoFilters` reuse). With SPEC-007 verify (test-helper either-A-or-B that violated `no-sql-in-cli-layer` under one branch) and SPEC-017 ship reflection ("either-is-fine off-loads to build" pattern named) as the prior data points, that is three confirming cases. Codify: "When a 'multiple paths' choice is decidable at design time — meaning one path violates a blocking constraint, encodes a structural anti-pattern, or duplicates a known lesson — lock the prescribed path AND list the rejected alternatives explicitly. 'Either-is-fine' is the off-load that 90% of the time slips into Deviations later."

3. **Is there a follow-up spec I should write now before I forget?**
   — No. Build and verify both said no. SPEC-019 (`brag review`) and SPEC-020 (`brag stats`) are framed in the stage backlog and inherit DEC-014 + `internal/aggregate` directly — no preconditions SPEC-018 left undone. The two §9/§12 refinements above are AGENTS.md chores, not specs, landing in this ship cycle as `chore(AGENTS)`. The cosmetic doc-sync (`AGENTS.md:67`/`:68` tap+CI distribution, `docs/architecture.md:24` mermaid label, `docs/tutorial.md:453` `brew install bragfile` row — all stale STAGE-004 references for assets that moved to STAGE-005, plus the architecture mermaid's STAGE-003 grouping that conflated export with summary) was folded into PR #18 as commit `baca793` (`chore(SPEC-018): doc-sync 4 stale stage references caught at verify`) — fixing the three verify-flagged misses plus an adjacent fourth (line 68's CI mention, same shape as line 67) caught while the file was open.
