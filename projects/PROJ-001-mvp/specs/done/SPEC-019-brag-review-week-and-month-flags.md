---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-019
  type: story                      # epic | story | task | bug | chore
  cycle: ship
  blocked: false
  priority: medium
  complexity: S                    # S — consumes DEC-014 verbatim, one new helper, one new command, one new render file. No DEC emission, no new package.

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
    - DEC-006   # cobra framework — new `brag review` subcommand
    - DEC-007   # required-flag validation in RunE — `--week`/`--month` mutex + `--format` use UserErrorf
    - DEC-011   # naked-array JSON shape — entries_grouped[].entries items reuse the 9-key DEC-011 entry shape
    - DEC-013   # markdown export shape — provenance-block convention DEC-014 inherited and SPEC-019 reuses
    - DEC-014   # CONSUMED VERBATIM — single-object envelope (JSON) + provenance/summary block (markdown) for digest commands
  constraints:
    - no-sql-in-cli-layer
    - stdout-is-for-data-stderr-is-for-humans
    - errors-wrap-with-context
    - test-before-implementation
    - one-spec-per-pr
  related_specs:
    - SPEC-007   # shipped; ListFilter struct + Store.List read path; review uses ListFilter{Since: cutoff} only
    - SPEC-014   # shipped; emitted DEC-011 + created internal/export package; SPEC-019 lifts toEntryRecord helper here
    - SPEC-015   # shipped; provenance/summary-block markdown precedent; substring-trap addendum applies to heading-level asserts
    - SPEC-018   # shipped; emitted DEC-014, seeded internal/aggregate, defined `rangeCutoff` helper SPEC-019 reuses verbatim, established the Rejected-alternatives-build-time discipline this spec applies a second time
    - SPEC-020   # pending; sibling stage spec (`brag stats`); will extend internal/aggregate further with Streak/MostCommon/Span
---

# SPEC-019: `brag review --week | --month` — rule-based reflection digest consuming DEC-014

## Context

Second of three specs in STAGE-004. SPEC-018 shipped 2026-04-25 with the
load-bearing pieces of the stage in place: DEC-014 (rule-based output
shape — single-object JSON envelope plus provenance/summary markdown
convention), the `internal/aggregate` package (`ByType`, `ByProject`,
`GroupForHighlights`, `rangeCutoff`), and the first DEC-014 consumer
(`brag summary`). SPEC-019 is the second consumer. SPEC-020 (`brag
stats`) is the third.

`brag review --week | --month` is the **reflection ritual** of the
stage. Where `brag summary` is the lightweight digest ("what happened
this period?") and `brag stats` is the lifetime panorama ("show me the
chart-of-myself"), `brag review` is the prompt-for-self-reflection:
recent entries grouped by project, followed by three hard-coded
reflection questions designed to be pasted into an external AI
session for guided self-review. No LLM ships in the binary; that's
PROJ-002's reason for existing.

The spec does **three** things in one pass — a meaningful reduction
from SPEC-018's four-things-in-one because the load-bearing
infrastructure already exists:

1. **Adds `brag review --week | --month [--format markdown|json]`**
   as the second DEC-014 consumer. `--week` and `--month` are named
   mutually-exclusive flags (asymmetric with `brag summary --range
   week|month` — justified in DEC-014 ref + Locked design decisions §3).
   Bare `brag review` defaults to `--week`. NO filter flags
   (`--tag`/`--project`/`--type` deliberately rejected per stage
   Design Notes "Filter flag reuse"). stdout only; no `--out`.

2. **Renders via `internal/export/review.go`** (new file, sibling to
   `summary.go` + `markdown.go` + `json.go`). Consumes DEC-014's
   envelope verbatim. Per-spec payload keys: `entries_grouped` (array
   of `{project, entries: [...]}` with full DEC-011 entry shape inside)
   + `reflection_questions` (array of three strings). Markdown
   elides descriptions for compactness (matches summary's `## Highlights`
   bullet shape: `- <id>: <title>`); JSON includes the full DEC-011
   entry shape because AI consumers may want descriptions.

3. **Extends `internal/aggregate` minimally** — adds one new helper
   `GroupEntriesByProject` returning `[]ProjectEntryGroup{Project,
   Entries []storage.Entry}`. Mirrors `GroupForHighlights`'s sort
   logic exactly (alpha-ASC by project, `(no project)` last;
   chrono-ASC + ID-tiebreak within group), differing only in carrying
   full `storage.Entry` instead of the `EntryRef{ID, Title}`
   projection. Used by both review's markdown path (renders ID +
   Title only) and review's JSON path (serializes full DEC-011
   shape). Inlining the grouping in `review.go` was rejected; see
   Locked design decisions → Rejected alternatives (build-time) §3.

DEC-014 is **consumed verbatim** — no DEC emission this spec, no
re-litigation of the six locked choices. SPEC-018's load-bearing
goldens (`TestToSummaryMarkdown_DEC014FullDocumentGolden` and
`TestToSummaryJSON_DEC014ShapeGolden`) prove the envelope works on
the summary side; SPEC-019's mirror goldens
(`TestToReviewMarkdown_DEC014FullDocumentGolden` and
`TestToReviewJSON_DEC014ShapeGolden`) prove the review side composes
the same envelope shape. If review's goldens ever fail in a way that
implies the envelope changed, that's a DEC-014 violation — fix the
code, not the test.

Parent stage:
[`STAGE-004-rule-based-polish-summary-review-stats.md`](../stages/STAGE-004-rule-based-polish-summary-review-stats.md)
— Spec Backlog → SPEC-019 entry (lines ~207–217); Design Notes →
"DEC-014" / "`internal/aggregate` package" / "Filter flag reuse" /
"Output destination" / "SPEC-019-specific (`brag review`)" (lines
~356–388). Project: PROJ-001 (MVP).

## Goal

Ship `brag review --week | --month [--format markdown|json]` as the
second DEC-014 consumer: render recent entries grouped by project
followed by three hard-coded reflection questions, consuming DEC-014's
envelope verbatim, with the JSON shape mirroring summary's `highlights`
array-of-objects ergonomics and the markdown elision-of-descriptions
matching summary's `## Highlights` bullet shape.

## Inputs

- **Files to read:**
  - `/AGENTS.md` — §6 cycle rules; §7 spec anatomy; §8 DEC emission
    (note: SPEC-019 does NOT emit a DEC; consumed-verbatim DEC-014);
    §9 premise-audit family — SPEC-019 is **addition + status-change**;
    apply the §9 audit-grep cross-check addendum (SPEC-018-earned —
    "design enumerates → design verifies its enumeration → build
    re-verifies and questions deltas") at design time; §11 Domain
    Glossary already has `review` entry from SPEC-018 (line 253 —
    glossary entry is descriptive, does NOT quote question wording;
    SPEC-019 augments with the locked question wording so future
    readers don't track it down through specs); §12 Cycle-Specific
    Rules including the "decide at design time when decidable"
    discipline SPEC-018 first applied proactively.
  - `/projects/PROJ-001-mvp/brief.md` — STAGE-004 stage-plan + sketch.
  - `/projects/PROJ-001-mvp/stages/STAGE-004-rule-based-polish-summary-review-stats.md`
    — THE authoritative scope for this spec. Spec Backlog → SPEC-019
    entry (lines ~207–217); Design Notes → "DEC-014" (lines ~250–261)
    consumed verbatim; "`internal/aggregate` package" (lines ~263–274)
    — extension is allowed and locked here; "Filter flag reuse" (lines
    ~276–282) — `brag review` does NOT accept filter flags in MVP
    (locked NO); "Output destination" (lines ~284–289) — stdout only,
    no `--out` (locked NO); "SPEC-019-specific (`brag review`)"
    (lines ~356–388).
  - `/projects/PROJ-001-mvp/session-log.md` — recent state.
  - `/projects/PROJ-001-mvp/backlog.md` — NOT for scope; for awareness
    of out-of-scope siblings (`brag remind`, emoji passes, `brag add
    --at`, `--out` deferral, configurable reflection questions if a
    user ever asks).
  - `/decisions/DEC-014-rule-based-output-shape.md` — THE shape
    contract this spec consumes verbatim. Choices (1)–(6) all apply;
    (1) envelope-not-array, (2) top-level flat keys, (3) markdown
    convention, (4) empty-state values, (5) indent=2 JSON,
    (6) rolling-window semantics for `--week`/`--month` (review's
    flags map to the same `rangeCutoff` helper SPEC-018 created).
  - `/decisions/DEC-013-markdown-export-shape.md` — provenance/
    summary block convention DEC-014 inherits from; review's markdown
    follows the same provenance shape (`Generated:` / `Scope:` /
    `Filters:` lines).
  - `/decisions/DEC-011-json-output-shape.md` — naked-array shape
    for `brag list` / `brag export`. Review's `entries_grouped[].entries`
    items use the same 9-key per-entry shape DEC-011 locks (`id`,
    `title`, `description`, `tags`, `project`, `type`, `impact`,
    `created_at`, `updated_at`); the wrapper that holds them is the
    DEC-014 envelope, not DEC-011's naked array. The two DECs
    compose without conflict — DEC-011 governs the per-row shape;
    DEC-014 governs the per-document envelope.
  - `/decisions/DEC-006-cobra-cli-framework.md` — new `brag review`
    is a cobra subcommand following the same pattern.
  - `/decisions/DEC-007-required-flag-validation-in-runE.md` — applies
    to mutual-exclusion validation between `--week` and `--month` and
    to `--format` validation; both go through `UserErrorf` in
    `RunE`, never `MarkFlagsMutuallyExclusive` (DEC-007 disposition).
  - `/guidance/constraints.yaml` — full constraint list.
  - `/projects/PROJ-001-mvp/specs/done/SPEC-018-brag-summary-aggregate-package-and-shape-dec.md`
    — DIRECT precedent. SPEC-018 created `internal/aggregate` +
    `internal/export/summary.go`; SPEC-019 follows the same shape:
    `internal/export/review.go` (sibling to `summary.go`), reuses
    `internal/aggregate.GroupForHighlights` (markdown side via the
    new `GroupEntriesByProject` helper) + the `rangeCutoff` helper.
    Read its Notes for the Implementer for code-sketch patterns and
    its Locked-design-decisions section for the Rejected
    alternatives (build-time) discipline this spec re-applies.
  - `/projects/PROJ-001-mvp/specs/done/SPEC-017-brag-add-json-stdin-and-schema-dec.md`
    — for the SPEC-017 ship reflection on "decide at design time
    when decidable." SPEC-019 follows SPEC-018's precedent of a
    Rejected-alternatives-build-time subsection; four entries below.
  - `/projects/PROJ-001-mvp/specs/done/SPEC-014-json-trio-and-shared-shape-dec.md`
    — created `internal/export/json.go` and `entryRecord` (currently
    package-private). SPEC-019 extracts `toEntryRecord(storage.Entry)
    entryRecord` as a sibling helper in `json.go` so review's JSON
    path can serialize per-DEC-011 entry shape without duplicating
    the field map. Lifting `entryRecord` to exported is rejected
    (Rejected alternatives §4) — package-private is sufficient.
  - `/internal/aggregate/aggregate.go` — exists post-SPEC-018; read
    to understand the `GroupForHighlights` signature + sort logic
    SPEC-019's `GroupEntriesByProject` mirrors. Note `NoProjectKey`
    constant + `ProjectHighlights{Project, Entries []EntryRef}`
    pattern.
  - `/internal/aggregate/aggregate_test.go` — sibling tests file;
    SPEC-019's new helper gets two tests added at the bottom of
    the existing file (NOT a separate file, since it's the same
    package and the existing file is the convention).
  - `/internal/export/summary.go` — sibling renderer; review.go
    follows the same shape (`SummaryOptions`-style options struct
    with injectable `Now`; markdown returns bytes with trailing
    `\n` stripped; JSON via `MarshalIndent` indent=2).
  - `/internal/export/json.go` — `entryRecord` lives here; SPEC-019
    extracts `toEntryRecord(storage.Entry) entryRecord` package-
    private helper.
  - `/internal/export/markdown.go` — sibling; precedent for the
    `noProjectKey` constant + provenance line format (`Exported:` /
    `Entries:` / `Filters:`). Review's provenance is
    `Generated:` / `Scope:` / `Filters:` per DEC-014.
  - `/internal/cli/summary.go` — direct precedent for the CLI
    command structure: `--format` validation in `RunE`, `UserErrorf`
    patterns, the `runSummary` flow (resolve DB → open store →
    `Store.List(ListFilter{Since: cutoff})` → render → write to
    stdout). Review's CLI structure is symmetric, with three
    differences: (a) named `--week`/`--month` mutex flags instead
    of `--range week|month`; (b) NO filter flags; (c) bare invocation
    silently defaults to `--week` (no notice on stderr).
  - `/internal/cli/summary_test.go` — direct precedent for CLI test
    harness (`newSummaryTestRoot`, `runSummaryCmd` helpers; line-
    based equality for heading-level asserts; regex strip for
    `Generated:` line; separate `outBuf` / `errBuf` with no-
    cross-leakage asserts). Review's test file follows the same
    shape.
  - `/internal/cli/list_test.go` — `seedListEntry` helper at line
    ~258, same package, direct reuse.
  - `/internal/cli/errors.go` — `ErrUser` sentinel + `UserErrorf`
    helper.
  - `/cmd/brag/main.go` — gains one `root.AddCommand(cli.NewReviewCmd())`
    line after the existing eight `AddCommand` calls.
- **External APIs:** none. stdlib `encoding/json`, `time`, `bytes`,
  `fmt` cover the needs. No new Go module dependencies. Per
  `no-new-top-level-deps-without-decision`, any proposed dep needs
  its own DEC.
- **Related code paths:** `internal/aggregate/` (small extension);
  `internal/export/` (new `review.go` file + `entryRecord` helper
  extraction in `json.go`); `internal/cli/` (new `review.go` file);
  `cmd/brag/main.go`; `docs/`; `README.md`; `AGENTS.md`.

## Outputs

- **Files created:**
  - `/internal/export/review.go` — new file in the existing
    `internal/export` package. Exports:
    - `type ReviewOptions struct { Scope string; Now time.Time }`
      — `Scope` is `"week"` or `"month"` (echoed into `Scope:`
      provenance line + JSON `scope` key); `Now` is injected for
      deterministic `Generated:` lines (mirrors
      `MarkdownOptions.Now` + `SummaryOptions.Now`). NO `Filters`
      / `FiltersJSON` fields — review does not accept filter flags.
      The provenance line for filters is hard-coded `(none)` per
      DEC-014's filters-is-an-object discipline (markdown side
      renders `Filters: (none)` for symmetry with summary; JSON
      side renders `"filters": {}`).
    - `func ToReviewMarkdown(entries []storage.Entry, opts ReviewOptions) ([]byte, error)`
      — renders the markdown digest per DEC-014. Returns bytes with
      trailing `\n` stripped (matches `ToJSON` / `ToMarkdown` /
      `ToSummaryMarkdown` byte contract). Empty entries slice still
      renders the reflection-questions block (the distinguishing
      contract: questions are the *point* of the command — they
      always appear, even when zero entries match). Document
      structure: `# Bragfile Review` heading + provenance block +
      `## Entries` (when non-empty: per-project groups underneath)
      + `## Reflection questions` (always renders).
    - `func ToReviewJSON(entries []storage.Entry, opts ReviewOptions) ([]byte, error)`
      — renders the JSON envelope per DEC-014. Single object, top-
      level keys: `generated_at`, `scope`, `filters` (always `{}`),
      `entries_grouped` (array of `{project, entries: [<full
      DEC-011 entry>, ...]}` objects, group order alpha-ASC with
      `(no project)` last, within-group entries chrono-ASC),
      `reflection_questions` (array of three strings). Pretty-
      printed with 2-space indent.
    - Unexported package-level `var reflectionQuestions = []string{...}`
      — the three locked questions, verbatim per Locked design
      decisions §4. Lives in `review.go` (single file, no
      `review_questions.go` carve-out — three strings don't justify
      a separate file). The same `var` is consumed by both renderers.
  - `/internal/export/review_test.go` — new file. Six tests against a
    fixed `[]storage.Entry` fixture + explicit `ReviewOptions`. Two
    load-bearing goldens (markdown + JSON). See Failing Tests.
  - `/internal/cli/review.go` — new file. Exports
    `func NewReviewCmd() *cobra.Command` plus unexported
    `runReview`. Declares `--week` (boolean flag), `--month`
    (boolean flag), `--format` (default `markdown`; RunE-validated;
    accepted: `markdown`, `json`). NO filter flags (`--tag`,
    `--project`, `--type` deliberately not declared per stage
    Design Notes "Filter flag reuse"). Validates mutual exclusion
    in `RunE` via `cmd.Flags().Changed("week")` +
    `cmd.Flags().Changed("month")` (both true → `UserErrorf`
    naming both flags). Bare invocation (neither flag changed)
    silently defaults to `--week` (no stderr notice — locked
    decision §5). Computes the cutoff via the existing
    `rangeCutoff(scope, time.Now().UTC())` helper from
    `summary.go`. Calls `Store.List(ListFilter{Since: cutoff})`,
    renders via `ToReviewMarkdown` / `ToReviewJSON` to stdout.
  - `/internal/cli/review_test.go` — new file. Five tests using
    `t.TempDir()` for DB paths, seeding entries via the package-
    local `seedListEntry` helper. See Failing Tests.
- **Files modified:**
  - `/internal/aggregate/aggregate.go` — adds:
    - `type ProjectEntryGroup struct { Project string; Entries []storage.Entry }`
    - `func GroupEntriesByProject(entries []storage.Entry) []ProjectEntryGroup`
      — mirrors `GroupForHighlights`'s sort logic exactly (groups
      alpha-ASC, `(no project)` last; within-group chrono-ASC with
      ID tiebreak). Returns full `storage.Entry` slices instead of
      the `EntryRef{ID, Title}` projection. Empty input returns
      a non-nil empty slice (matches the empty-state contract from
      DEC-014 part (4) + the existing helpers).
  - `/internal/aggregate/aggregate_test.go` — adds two tests at the
    end of the file (same package convention; not a separate file):
    - `TestGroupEntriesByProject_OrderingAndIDTiebreak`
    - `TestGroupEntriesByProject_EmptyInputReturnsNonNilEmptySlice`
  - `/internal/export/json.go` — extracts `toEntryRecord(e
    storage.Entry) entryRecord` package-private helper from the
    existing inline literal in `ToJSON` (lines 41–51 of the current
    file). The body of `ToJSON` becomes:

    ```go
    records := make([]entryRecord, 0, len(entries))
    for _, e := range entries {
        records = append(records, toEntryRecord(e))
    }
    ```

    The new helper is consumed by `internal/export/review.go`'s JSON
    path. NO behavior change to `ToJSON`. NO export change (still
    package-private). Existing `internal/export/json_test.go`
    goldens stay byte-identical.
  - `/cmd/brag/main.go` — one added line:
    `root.AddCommand(cli.NewReviewCmd())` after the existing eight
    `AddCommand` calls (line 25 in the post-SPEC-018 file).
  - `/docs/api-contract.md` — adds a new `### brag review --week | --month`
    section after the `### brag summary --range week|month` section
    (line ~261 today, after SPEC-018's REPLACE-the-placeholder
    rewrite). Section content (drafted in detail under Notes for
    the Implementer):
    - synopsis with `--week` / `--month` (mutually exclusive; bare
      defaults to `--week`); `--format markdown|json` (default
      markdown);
    - prose describes recent-entries-grouped-by-project + three
      hard-coded reflection questions appended;
    - cross-link to DEC-014 + DEC-011 (envelope shape + per-entry
      shape inside `entries_grouped`);
    - mention markdown elides descriptions / JSON includes them
      (the documented asymmetry);
    - rolling-window semantics (week = last 7 UTC days; month = last
      30 UTC days) inheriting DEC-014 choice (6);
    - filter flags NOT accepted (locked NO per stage Design Notes;
      stated explicitly so future contributors don't try to add them);
    - stdout only (no `--out` per stage Design Notes; locked NO).
    The end-of-file References list (line ~284–290 today) is
    unchanged: DEC-014 was already added by SPEC-018 and already
    references "and `brag review` / `brag stats` arriving later
    in STAGE-004" — SPEC-019 ships review, but the line stays
    accurate (review now shipped, stats still pending; the line
    can be updated by SPEC-020 when stats ships, or this spec can
    tighten the wording to "and `brag stats` arriving later in
    STAGE-004"). SPEC-019 takes the latter path — see Notes for
    the Implementer.
  - `/docs/tutorial.md` — (a) line 3–4 Scope blurb: `\`brag review\`
    and \`brag stats\` arrive in later STAGE-004 specs` →
    `\`brag stats\` arrives in a later STAGE-004 spec` (review
    now shipped). (b) Adds an optional `### Weekly reflection: brag
    review` subsection under §4 "Read them back" showing the AI-
    paste pattern (entries-grouped-by-project view + the three
    reflection questions, with a one-liner on piping the JSON form
    into Claude/GPT). Author-judgment call on subsection placement;
    keep the prose short (5–10 lines). Mirrors SPEC-018's
    optional-but-recommended `### Weekly digest: brag summary`
    addition.
  - `/docs/data-model.md` — line 149 update: the existing DEC-014
    row mentions "and `brag review` / `brag stats` in later STAGE-004
    specs". Update to "`brag review` (this spec) and `brag stats`
    (later STAGE-004 spec)" — review is now an active consumer of
    DEC-014, not a forward reference. No schema change.
  - `/README.md` — line ~63 update: the existing wording mentions
    `\`brag review\` and \`brag stats\`` as deferred. Update to
    mention `brag review --week|--month` shipped (STAGE-004), with
    `brag stats` still forward-referenced.
  - `/AGENTS.md` — §11 Domain Glossary updates:
    - **`review` entry (line 253)** — currently descriptive, does
      NOT quote question wording. AUGMENT with the three locked
      questions verbatim so future contributors don't have to track
      them through specs. New text (replacing the existing entry):
      > **review** — `brag review --week | --month`: prints recent
      > entries grouped by project followed by three hard-coded
      > reflection questions ("What pattern do you see in this
      > period?", "What did you underestimate?", "What's missing
      > here that should be?"). Markdown elides per-entry
      > descriptions for compactness; JSON includes the full
      > DEC-011 entry shape. Designed to be pasted into an external
      > AI session for guided self-reflection. STAGE-004 (SPEC-019).
    - **`aggregate` entry (line 246)** — currently mentions
      "SPEC-019's grouping helpers and SPEC-020's `Streak` /
      `MostCommon` / `Span`". Update the SPEC-019 part to name the
      actual helper added: "SPEC-019's `GroupEntriesByProject`
      helper". Keeps the glossary aligned with what shipped.
- **New exports:**
  - `aggregate.ProjectEntryGroup`, `aggregate.GroupEntriesByProject`.
  - `export.ReviewOptions`, `export.ToReviewMarkdown`,
    `export.ToReviewJSON`.
  - `cli.NewReviewCmd`.
- **Database changes:** none. Pure read path; uses existing
  `Store.List(ListFilter)` from SPEC-007. No migration.

## Acceptance Criteria

Every criterion is testable. Paired failing test name in italics
where applicable. SPEC-019 has **12 failing tests** across **2 new
files** plus **1 existing file** (`internal/aggregate/aggregate_test.go`
gains 2 tests at the end). Load-bearing goldens written FIRST per
SPEC-014 / SPEC-015 / SPEC-018 ship lessons.

- [ ] `aggregate.GroupEntriesByProject(fixture)` returns
      `[]ProjectEntryGroup` with groups ordered alpha-ASC by
      project name (`(no project)` last regardless of count);
      within each group, entries are sorted ASC by `CreatedAt`
      with `ID` as tie-break (locks AGENTS.md §9 SPEC-002
      monotonic-tiebreak rule). Each `Entries` slice carries
      full `storage.Entry` (NOT `EntryRef`).
      *TestGroupEntriesByProject_OrderingAndIDTiebreak*
- [ ] `aggregate.GroupEntriesByProject(nil)` and
      `aggregate.GroupEntriesByProject([]storage.Entry{})` each
      return a non-nil empty slice (`len == 0`, distinct from
      `nil`) so the JSON renderer can marshal as `[]` not `null`.
      Matches the empty-state discipline shipped by SPEC-018's
      `TestAggregate_EmptyInputReturnsNonNilEmptySlice` for the
      three existing helpers.
      *TestGroupEntriesByProject_EmptyInputReturnsNonNilEmptySlice*
- [ ] `export.ToReviewMarkdown(fixture, opts{Scope: "week", Now:
      <fixed>})` emits a byte-identical markdown document locking
      DEC-014 markdown choices on the fixture: `# Bragfile Review`
      heading + provenance block (`Generated:` / `Scope: week` /
      `Filters: (none)`) + `## Entries` wrapper + per-project
      `### <project>` groups with bulleted `- <id>: <title>` per
      entry (descriptions ELIDED) + `## Reflection questions` with
      three numbered questions. *(LOAD-BEARING — write FIRST per
      SPEC-014/15/18 ship lessons.)*
      *TestToReviewMarkdown_DEC014FullDocumentGolden*
- [ ] `export.ToReviewJSON(fixture, opts{Scope: "week", Now:
      <fixed>})` emits a byte-identical JSON envelope locking
      DEC-014 JSON choices on the same fixture: `generated_at`
      (RFC3339), `scope: "week"`, `filters: {}`,
      `entries_grouped` (array-of-objects with full DEC-011
      9-key entry shape inside, NOT `EntryRef` projection),
      `reflection_questions` (array of three strings, exact
      verbatim wording). Pretty-printed indent=2. *(LOAD-BEARING
      — write SECOND.)*
      *TestToReviewJSON_DEC014ShapeGolden*
- [ ] `ToReviewMarkdown` AND `ToReviewJSON` on empty entries still
      emit the reflection-questions block. Document structure on
      empty-markdown: `# Bragfile Review` heading + provenance
      block + `## Reflection questions` + three numbered questions
      — NO `## Entries` wrapper, NO per-project groups. Document
      structure on empty-JSON: full envelope with
      `entries_grouped: []` and `reflection_questions: [<three
      strings>]`. The distinguishing contract — questions are the
      *point* of the command, so they always render. Diverges
      from SPEC-018's `TestToSummary_EmptyEntriesEmitsProvenanceOnly`
      because review's payload has a non-entry-derived part
      (questions) that doesn't elide.
      *TestToReview_EmptyEntriesStillEmitsReflectionQuestions*
      (subtests `markdown` and `json`).
- [ ] The three reflection questions appear with exact verbatim
      wording in BOTH markdown (numbered list under
      `## Reflection questions`) and JSON (string array
      `reflection_questions`):
      1. `What pattern do you see in this period?`
      2. `What did you underestimate?`
      3. `What's missing here that should be?`
      Wording locked from STAGE-004 Design Notes lines 367–369
      (NOT the looser Success Criteria paraphrase nor the AGENTS.md
      glossary line which is descriptive-only).
      *TestToReview_ReflectionQuestionsExactWording* (subtests
      `markdown` and `json`).
- [ ] `brag review` (no flags) silently defaults to `--week` —
      stdout starts with `# Bragfile Review` (markdown default);
      a line `Scope: week` appears; stderr is empty (no
      "defaulting to --week" notice). *TestReviewCmd_BareDefaultsToWeek*
- [ ] `brag review --week --month` exits 1 (user error) with a
      message naming both `--week` AND `--month` AND describing
      mutual exclusion. Per DEC-007: validated in `RunE` via
      `UserErrorf`, NOT via `MarkFlagsMutuallyExclusive`.
      *TestReviewCmd_WeekAndMonthMutuallyExclusiveIsUserError*
- [ ] `brag review --week --format yaml` (unknown format) exits
      1 (user error) with a message naming `yaml` AND the
      accepted set (`markdown`, `json`).
      *TestReviewCmd_UnknownFormatIsUserError*
- [ ] `brag review --month --format json` writes the JSON
      envelope per DEC-014 to stdout: top-level keys
      `generated_at`, `scope` = `"month"`, `filters` = `{}`,
      `entries_grouped` (array; structure per DEC-014 +
      SPEC-019 payload key), `reflection_questions` (array of
      three strings). Trailing newline from `fmt.Fprintln`;
      stderr empty.
      *TestReviewCmd_FormatJSON_MonthScopeAndEntriesGroupedShape*
- [ ] `brag review --help` output contains `--week`, `--month`,
      `--format`, `markdown`, `json` (each as distinctive needles
      per AGENTS.md §9 assertion-specificity rule). The help text
      DOES NOT advertise `--tag`, `--project`, `--type`, or
      `--out` (filter and out flags are not declared on this
      command). Stderr empty. *TestReviewCmd_HelpShowsWeekMonthAndFormat*
- [ ] `brag review --tag X` exits with cobra's "unknown flag"
      error (NOT a `RunE` user-error path). Locks the "filter
      flags not accepted" decision per stage Design Notes "Filter
      flag reuse" — `--tag` is genuinely undeclared, not declared-
      and-rejected. Same applies to `--project`, `--type`, `--out`,
      and `--since`. *TestReviewCmd_FilterAndOutFlagsRejectedAsUnknown*
      (subtests for each rejected flag name).
- [ ] `brag --help` lists `review` as a subcommand (cobra auto-
      registers it via `cmd/brag/main.go` AddCommand).
      *[manual: `go build ./cmd/brag && ./brag --help` shows
      `review` in the command list; `./brag review --help`
      shows the synopsis.]*
- [ ] `gofmt -l .` empty; `go vet ./...` clean; `CGO_ENABLED=0
      go build ./...` succeeds; `go test ./...` and `just test`
      green. Existing SPEC-018 goldens
      (`TestToSummaryMarkdown_DEC014FullDocumentGolden`,
      `TestToSummaryJSON_DEC014ShapeGolden`) AND existing
      `TestToJSON_*` goldens stay byte-identical (proves the
      `toEntryRecord` extraction in `json.go` was a no-op
      refactor).
- [ ] Doc sweep: `docs/api-contract.md` gains a `### brag review
      --week | --month` section after the summary section;
      `docs/api-contract.md` line ~315 updates "and `brag review`
      / `brag stats` arriving later in STAGE-004" → "and `brag
      stats` arriving later in STAGE-004"; `docs/tutorial.md`
      line 3–4 Scope blurb updated; `docs/data-model.md` line 149
      updated; `README.md` line ~63 updated; `AGENTS.md` §11
      Domain Glossary `review` entry augmented with question
      wording, `aggregate` entry updated to name the new helper.
      *[manual greps listed under Premise audit below.]*

## Locked design decisions

Reproduced here so build / verify don't re-litigate. Each is paired
with at least one failing test below per AGENTS.md §9 SPEC-009 ship
lesson + the SPEC-018 §12 "decide at design time when decidable"
discipline.

1. **DEC-014 consumed verbatim — no DEC emission, no re-litigation.**
   All six locked choices apply: (1) JSON single-object envelope,
   (2) top-level flat keys (`generated_at`, `scope`, `filters` plus
   per-spec payload keys `entries_grouped`, `reflection_questions`
   at top level — no nested `payload` wrapper), (3) markdown
   provenance-block convention (`# <Doc Title>` heading + `Generated:`
   / `Scope:` / `Filters:` lines), (4) empty-state values (numeric
   → 0; arrays → `[]`; objects → `{}`; date fields → null in JSON
   / `-` in markdown), (5) JSON pretty-printed indent=2,
   (6) `--week`/`--month` rolling-window semantics (week = last 7
   UTC days; month = last 30 UTC days). The `scope` field echoes
   `"week"` or `"month"` (matches summary's `scope` echoes for
   cross-command jq ergonomics). *Pair: load-bearing
   `TestToReviewMarkdown_DEC014FullDocumentGolden` +
   `TestToReviewJSON_DEC014ShapeGolden` cover (1)–(5);
   `TestReviewCmd_BareDefaultsToWeek` +
   `TestReviewCmd_FormatJSON_MonthScopeAndEntriesGroupedShape`
   cover (6) at the CLI plumbing level (the SPEC-018 unit test
   `TestRangeCutoff_WeekMonthArithmeticAndErrors` already covers
   the helper arithmetic — review reuses the same helper, no
   duplicate unit test needed).*

2. **`entries_grouped` JSON shape: array-of-objects matching
   summary's `highlights` shape, NOT a project-keyed map.** Each
   element is `{project: "<name>", entries: [<full DEC-011 9-key
   entry>, ...]}`. Group order: alpha-ASC by project name with
   `(no project)` last. Within-group entries: chrono-ASC with
   ID-tiebreak. Symmetric with summary's
   `highlights: [{project, entries: [{id, title}, ...]}]` —
   review carries the FULL entry shape inside `entries`, summary
   carries the `EntryRef{ID, Title}` projection. Same wrapper,
   different per-row shape; both consumers of DEC-014. *Pair:
   `TestToReviewJSON_DEC014ShapeGolden` byte-locks the array-
   of-objects shape; `TestReviewCmd_FormatJSON_MonthScopeAndEntriesGroupedShape`
   exercises the CLI path.*

3. **Markdown shape mirrors summary's `## Highlights`: `## Entries`
   wrapper around per-project `### <project>` groups with bulleted
   `- <id>: <title>` per entry; descriptions ELIDED.** Document
   structure is provenance-then-entries-then-questions. The
   `## Entries` wrapper exists to keep document depth uniform with
   summary's `## Summary` + `## Highlights` two-section layout
   (review's `## Entries` parallels `## Highlights`); leaves room
   for a future `## Summary` block if review ever grows one. JSON
   includes full descriptions (per DEC-011 9-key shape inside
   `entries_grouped[].entries`); markdown elides them. The
   asymmetry is documented and locked — fast-scan view in markdown
   gets pasted into AI; programmatic AI consumers ingest the JSON
   with full text. *Pair:
   `TestToReviewMarkdown_DEC014FullDocumentGolden` byte-locks the
   markdown elision; `TestToReviewJSON_DEC014ShapeGolden` byte-
   locks the JSON inclusion.*

4. **`--week` and `--month` are named boolean flags, mutually
   exclusive, validated in `RunE` via `UserErrorf` (DEC-007),
   NEVER via `MarkFlagsMutuallyExclusive`.** Bare `brag review`
   (neither flag changed) silently defaults to `--week` — no
   stderr notice (locked decision §5). Asymmetric with `brag
   summary --range week|month` — review's cadence is invoked as
   a verb (named reflection ritual: "do my weekly review"), not
   parameterized aggregation. The asymmetry is intentional and
   surface-level only; both commands route through the SAME
   `rangeCutoff(scope, now)` helper from `summary.go` for the
   actual arithmetic. *Pair: `TestReviewCmd_BareDefaultsToWeek`
   covers the silent default; `TestReviewCmd_WeekAndMonthMutuallyExclusiveIsUserError`
   covers the mutex validation; the helper itself is already
   tested by SPEC-018's `TestRangeCutoff_WeekMonthArithmeticAndErrors`.*

5. **Bare `brag review` defaults silently to `--week` — no
   "defaulting to --week" stderr notice.** Help text mentions the
   default behavior in the Long description. The silent path
   keeps stdout pure (DEC-007 + `stdout-is-for-data-stderr-is-for-
   humans`), avoids surprising scripted users, and is reversible:
   if a user later complains "I expected an error, not a default"
   the deprecation-style notice can be added in a follow-up spec
   without breaking the silent-path behavior. *Pair:
   `TestReviewCmd_BareDefaultsToWeek` asserts `errBuf.Len() == 0`
   alongside the stdout shape.*

6. **Three reflection questions HARD-CODED, verbatim from STAGE-004
   Design Notes lines 367–369:**
   1. `What pattern do you see in this period?`
   2. `What did you underestimate?`
   3. `What's missing here that should be?`

   Lives as an unexported package-level
   `var reflectionQuestions = []string{...}` in
   `internal/export/review.go`. NO separate `review_questions.go`
   carve-out (three strings don't justify a file). Both renderers
   consume the same `var`. NOT configurable in MVP — see Rejected
   alternatives (build-time) §2. *Pair:
   `TestToReview_ReflectionQuestionsExactWording` (subtests
   markdown + json) byte-locks the exact wording.*

7. **`reflection_questions` in JSON is an array of three strings,
   NOT an array of objects with metadata.** Simplest shape that
   matches the markdown rendering one-to-one; AI consumers can
   extract trivially via `jq '.reflection_questions[]'`. An
   object-form like `[{question: "...", id: "pattern"}]` would
   be a forward-incompatible bet on metadata that doesn't yet
   exist; YAGNI. *Pair: covered by
   `TestToReviewJSON_DEC014ShapeGolden` byte-lock + the
   wording-exact subtests.*

8. **Filter flags `--tag`, `--project`, `--type` and `--out` are
   NOT declared on `brag review`.** Per stage Design Notes
   "Filter flag reuse" — review is "the last 7/30 days, period."
   Adding filter composition would multiply the spec scope without
   clear ergonomic win and would couple review to summary's
   filter handling. `--out` is also not declared — review is a
   pipe-friendly digest (redirect with `>` if you want a file)
   per stage Design Notes "Output destination". Cobra's auto-
   parser surfaces `unknown flag: --tag` errors for these names,
   which is the desired behavior — explicitly undeclared, not
   declared-and-rejected. *Pair:
   `TestReviewCmd_FilterAndOutFlagsRejectedAsUnknown` exercises
   each undeclared flag and asserts the error path.*

9. **`internal/aggregate` extension: ONE new helper
   (`GroupEntriesByProject`).** Mirrors `GroupForHighlights`'s
   sort + grouping logic verbatim, differing only in carrying full
   `storage.Entry` instead of the `EntryRef{ID, Title}`
   projection. Used by both review's markdown path (renders ID +
   Title only at render time) and review's JSON path (serializes
   full DEC-011 shape). NO refactor of `GroupForHighlights` into
   `GroupEntriesByProject + projection` — that's a sound future
   refactor (rejected alternatives §3 captures the path) but not
   SPEC-019's scope; out-of-scope per stage Design Notes "reuses
   what's there without extending" guidance, with the single-helper
   addition as the minimum delta. *Pair:
   `TestGroupEntriesByProject_OrderingAndIDTiebreak` +
   `TestGroupEntriesByProject_EmptyInputReturnsNonNilEmptySlice`.*

10. **`internal/export/json.go` extracts `toEntryRecord(storage.Entry)
    entryRecord` package-private helper.** Both `ToJSON` (existing
    DEC-011 consumer) and `ToReviewJSON` (new SPEC-019 consumer)
    serialize the same 9-key per-entry shape; sharing the field-
    map keeps drift impossible. Helper stays package-private
    (sufficient — both files in `internal/export`). NO behavior
    change to `ToJSON`; existing goldens
    (`internal/export/json_test.go`) stay byte-identical. *Pair:
    `TestToReviewJSON_DEC014ShapeGolden` exercises the helper on
    the JSON path; the build acceptance criterion (existing
    SPEC-014 goldens stay byte-identical) covers the no-op
    refactor side.*

**Out of scope (by design — backlog entries exist or are explicitly
deferred):**

- `--out <path>` flag on review. Backlog deferral noted in stage
  Design Notes "Output destination". Same backlog entry covers
  summary, review, and stats.
- `--since` / arbitrary date ranges on review. `--week`/`--month`
  only for MVP per stage scope.
- Filter flags (`--tag`/`--project`/`--type`) on review. Stage
  Design Notes "Filter flag reuse" locks NO; backlog if a real
  workflow emerges (likely never — the whole point of review is
  the unfiltered window).
- Configurable reflection questions (config file, flag, or env
  var). Hard-coded for MVP per stage Design Notes; backlog with
  revisit trigger "user wants to swap one out."
- A fourth+ reflection question. Three is the locked count; expanding
  the count would be a backlog item with revisit trigger "user
  request after using review for several weeks."
- `--compact` / non-pretty JSON for review (or any of the three
  STAGE-004 commands). Inherits DEC-011's pretty-default; same
  backlog entry covers them all.
- Time-zone configuration for `--week`/`--month` cutoff. UTC-only
  for MVP (matches storage's `time.Now().UTC()`); revisit if a
  user notices a window break across timezone changes.
- LLM piping / AI integration — PROJ-002 territory.
- `brag stats` (SPEC-020) — sibling spec, parallel scope.
- Calendar-week / calendar-month semantics. Rolling-window only;
  inherits DEC-014 choice (6).
- Any change to summary's existing behavior. SPEC-019 modifies
  `internal/export/json.go` to extract a package-private helper
  but `ToJSON`'s output bytes are unchanged.

**Rejected alternatives (build-time):**

These are choices the build agent might consider, with the
prescribed path locked here so the call doesn't off-load to
build-time and slip into Deviations later. Per SPEC-018 ship
reflection: "either-is-fine off-loads to build" is the
anti-pattern these locks deliberately avoid. SPEC-018 was the
first spec to apply this discipline proactively; SPEC-019 is the
second.

1. **Map-keyed `entries_grouped` JSON shape — REJECTED.** The
   path would render `entries_grouped` as a JSON object keyed by
   project name (`{"alpha": [...], "(no project)": [...]}`)
   instead of an array of `{project, entries}` objects.

   *Why rejected:*
   - **Loses iteration order.** Go's `encoding/json` sorts
     `map[string][]T` keys alphabetical-ASC when marshaling,
     which would put `(no project)` FIRST (parenthesis sorts
     before letters in ASCII) — directly contradicting DEC-014
     + DEC-013's "(no project) last" convention. The
     array-of-objects shape preserves the locked ordering.
   - **Breaks symmetry with summary's `highlights`.** DEC-014's
     value thesis is one stage-level shape across the three
     digest commands. summary uses
     `highlights: [{project, entries: [{id, title}]}]`; review
     using `entries_grouped: {"alpha": [...]}` would force AI
     consumers writing cross-command tooling to case-analyze
     the wrapper shape per command. The cost of array-of-objects
     (slightly verbose to lookup-by-name) is small; the payoff
     (consistent jq recipes across summary + review) is the
     stage's value thesis.
   - **Asymmetric jq recipes.** With map-keyed: `jq
     '.entries_grouped["alpha"][]'`. With array-of-objects:
     `jq '.entries_grouped[] | select(.project == "alpha") |
     .entries[]'`. The latter is identical to summary's
     `jq '.highlights[] | select(.project == "alpha") |
     .entries[]'`. Same recipe, different command — the win.

2. **Configurable reflection questions (config file or flag) —
   REJECTED.** The path would let users override the three
   locked questions via `~/.bragfile/review-questions.txt`,
   `--questions <path>` flag, or `BRAGFILE_REVIEW_QUESTIONS`
   env var.

   *Why rejected:*
   - **No current ask.** Hard-coded for MVP per stage Design
     Notes; the user has named the three questions as the
     personally-load-bearing wording.
   - **Spec scope creep.** Configurability adds: a path-resolution
     concern (file vs flag vs env precedence), a parse error
     surface, an empty/malformed-file user-error path, a
     "default to hard-coded if no override" fallback. Each is
     small; together they triple the spec scope without
     ergonomic win.
   - **Backlog with revisit trigger.** Captured in `backlog.md`
     with the trigger "user wants to swap one out." If the user
     ever asks, it's a focused follow-up spec — the hard-coded
     baseline is the right starting shape.

3. **Inline grouping in `internal/export/review.go` (no
   `aggregate.GroupEntriesByProject` helper) — REJECTED.** The
   path would do the project-bucket-and-sort logic directly in
   the review renderer: a `map[string][]storage.Entry`
   accumulation, the same alpha-ASC sort with `(no project)`
   last, the same chrono-ASC + ID-tiebreak within group.

   *Why rejected:*
   - **Duplicates `GroupForHighlights`'s sort logic.** Two
     places where the same alpha-ASC + no-project-last + chrono-
     ASC + ID-tiebreak rules live. Future-bug surface where one
     place changes and the other drifts (e.g., a future spec
     decides `(no project)` should sort first; both files must
     update or the renderers diverge).
   - **The locked-decisions-need-tests discipline applies most
     cleanly to a named unit.** `GroupEntriesByProject` is the
     named unit; the inline alternative obscures the contract
     and forces the test to either go through the cobra+render+
     parse layer (heavy and noisy) or to exercise an unexported
     helper (awkward).
   - **Reasonable scale.** One small helper is the right size
     for a SPEC-019-scope addition. The aggregate package's
     contract — pure data layer, structured output, no
     rendering — is honored cleanly by adding the helper.

4. **Lifting `entryRecord` to exported / package-public —
   REJECTED.** The path would change `entryRecord`'s name to
   `EntryRecord` (exported) so future packages could consume
   the DEC-011 9-key projection without going through
   `internal/export`.

   *Why rejected:*
   - **Package-private is sufficient.** Both files that consume
     `entryRecord` (`json.go` and the new `review.go`) live in
     `internal/export`. No third consumer exists today.
   - **YAGNI on SPEC-020.** SPEC-020 (`brag stats`) emits
     aggregations (counts, streaks, span) — NOT entry lists.
     Pre-extracting for SPEC-020 is speculative; if SPEC-020
     ever needs it, that spec earns the export decision.
   - **Stays-in-package keeps DEC-011's surface contained.** The
     DEC-011 per-entry shape is a serialization concern. Exporting
     `EntryRecord` would invite consumers outside the renderer
     (e.g., the CLI layer, or an MCP handler in PROJ-002) to
     marshal their own records, which would split the source
     of truth.

## Premise audit (AGENTS.md §9 — addition + status-change, with
audit-grep cross-check applied at design)

SPEC-019 is an **addition** case (new command, new aggregate helper,
new export functions, new exports list) AND a **status-change** case
(stage-doc + AGENTS.md + tutorial.md + README.md + data-model.md +
api-contract.md all forward-reference `brag review` as deferred
behavior; SPEC-019 supersedes those references). Both AGENTS.md §9
heuristics apply.

This spec is the **first proactive validation case for the SPEC-018-
earned audit-grep cross-check addendum** ("design enumerates →
design verifies its enumeration → build re-verifies and questions
deltas"). Each grep below was actually executed at design time
against the current working tree; the enumerated hits below match
the actual `rg` output as of 2026-04-25.

**Addition heuristics** (SPEC-011 ship lesson — grep tracked
collections for count coupling):

- Root command list: `cmd/brag/main.go` has eight `AddCommand`
  calls today (verified 2026-04-25, post-SPEC-018). SPEC-019 makes
  it nine. Grep:

  ```
  grep -rn 'AddCommand\|root.Commands()\|cmd.Commands()' \
    internal/cli/*.go cmd/ 2>/dev/null
  ```

  Audit each hit:
  - `cmd/brag/main.go:17–24`: the eight existing calls. Adding a
    ninth doesn't break any test.
  - `internal/cli/list_test.go`: iterates by name, not count. Safe.
  - No test asserts `len(root.Commands()) == 8` or similar.

- DEC collection: SPEC-019 adds NO new DEC. Verified 2026-04-25:
  the grep `find decisions -name 'DEC-*' | wc -l` returns 14 today
  (DEC-001 through DEC-014); SPEC-019 leaves it at 14.

- `internal/aggregate` exports: SPEC-019 adds two
  (`ProjectEntryGroup` type + `GroupEntriesByProject` function).
  No test asserts on the package's exported-symbol count.

- `internal/export` exports: SPEC-019 adds three (`ReviewOptions`,
  `ToReviewMarkdown`, `ToReviewJSON`). No test asserts on the
  package's exported-symbol count.

- `--format` accepted values: distinct flag per command. The
  list/export/summary tests asserting on `(accepted: json, tsv)`
  / `(accepted: json, markdown)` / `(accepted: markdown, json)`
  are unaffected by review's separate `(accepted: markdown, json)`
  list (same accepted set as summary, but the tests are command-
  scoped — verified by reading `internal/cli/summary_test.go`'s
  `TestSummaryCmd_HelpShowsRangeAndFormat` which only asserts on
  summary's help output).

- Help-command subcommand counts: per stage Design Notes, the
  per-spec premise-audit hot spot for SPEC-019 mentions "help-
  command subcommand counts +1." Grep:

  ```
  grep -rn 'NumCommand\|len.*Commands()' internal/cli/
  ```

  Verified 2026-04-25 against the working tree: no literal-count
  assertions exist. Safe.

- `internal/cli/list_test.go` `seedListEntry` helper: SPEC-019
  reuses it (same package, source-level reuse). Helper signature
  stable.

- `entryRecord` shape: SPEC-019 extracts `toEntryRecord(storage.Entry)
  entryRecord` from the inline literal in `ToJSON`. Behavior
  preserved. Existing `internal/export/json_test.go` goldens
  (`TestToJSON_*`) stay byte-identical — verified by inspection
  of the helper's body (same 9-field assignment, same RFC3339
  format, same UTC normalization).

**Status-change heuristics** (SPEC-012 ship lesson — grep feature
name across docs) — explicit grep commands AND actual hits as of
2026-04-25:

```
rg "brag review" docs/ README.md AGENTS.md
```

Actual hits (verified 2026-04-25):

- `README.md:63` — "`\`brag review\` and \`brag
  stats\` arrive in later STAGE-004 specs" (Scope blurb).
  → UPDATE: review now shipped, stats still forward-referenced.
- `AGENTS.md:249` — `digest` glossary entry mentions all three
  rule-based commands. Already accurate; no change needed.
- `AGENTS.md:253` — `review` glossary entry. Currently
  descriptive ("prints recent entries grouped by project followed
  by three hard-coded reflection questions"); does NOT quote
  question wording. AUGMENT with the locked verbatim question
  wording so future readers don't track it down through specs;
  no other change.
- `docs/data-model.md:149` — DEC-014 row mentions "and `brag
  review` / `brag stats` in later STAGE-004 specs". UPDATE: review
  now shipped — "`brag review` (this spec) and `brag stats` (later
  STAGE-004 spec)" or equivalent wording.
- `docs/tutorial.md:3` — Scope blurb opens with "`brag review` and"
  starting the deferral list. UPDATE Scope blurb so review is
  removed from the deferred list and stats is the only deferred
  STAGE-004 command.
- `docs/api-contract.md:315` — DEC-014 References row mentions
  "and `brag review` / `brag stats` arriving later in STAGE-004".
  UPDATE: "and `brag stats` arriving later in STAGE-004".

**Reflection-question wording cross-check (the §9 audit-grep
cross-check addendum applied — SPEC-018-earned, SPEC-019 first
validation case):**

```
rg "What pattern do you see"
rg "What did you underestimate"
rg "What's missing.*should be"
```

Actual hits (verified 2026-04-25):

- All wording hits land in the stage doc itself
  (`projects/PROJ-001-mvp/stages/STAGE-004-rule-based-polish-summary-review-stats.md`),
  with two distinct wording variants:
  - Stage Success Criteria (line ~73): paraphrased shorter form —
    `"What pattern do you see?"`, `"What did you underestimate?"`,
    `"What's missing that should be here?"`. Trailing question
    marks; no "in this period" qualifier on Q1; "should be here"
    word-order on Q3.
  - Stage Design Notes (lines 367–369): authoritative locked form —
    `What pattern do you see in this period?`, `What did you
    underestimate?`, `What's missing here that should be?`.
- AGENTS.md line 253 (the `review` glossary entry) is descriptive
  only; does NOT quote question wording. (Initial premise
  hypothesized a glossary-vs-stage mismatch; the audit-grep
  cross-check refuted that — the actual mismatch is internal to
  the stage doc.)
- No hits in `docs/`, `README.md`, or `internal/`. The wording
  exists in only one source file today.

**Symmetric action from `## Outputs` for the wording cross-check:**

- The locked wording is the Design Notes lines 367–369 form (per
  user direction in clarifying questions Q3). The stage doc's
  Success Criteria paraphrase is a known internal inconsistency
  but is NOT modified by SPEC-019 (the stage doc is parent-scope
  authoritative and not edited by child specs; the inconsistency
  is a stage-level lesson for ship-time stage reflection).
- AGENTS.md `review` glossary entry (line 253) is augmented with
  the locked verbatim wording so the wording propagates from
  spec → glossary, leaving a single authoritative quote outside
  the stage doc. Captured under `## Outputs` files-modified
  list.
- All test asserts for question wording use the Design Notes
  verbatim form. Locked decision §6 enumerates the three strings.

**Doc-sweep symmetric action enumeration** (from the rg hits
above — every hit maps to a planned `## Outputs` modification):

| Grep hit | Action under `## Outputs` |
| --- | --- |
| `README.md:63` | Update Scope blurb |
| `AGENTS.md:249` (digest entry) | No change (already accurate) |
| `AGENTS.md:253` (review entry) | Augment with locked wording + helper-name update on aggregate entry |
| `docs/data-model.md:149` | Update DEC-014 row wording |
| `docs/tutorial.md:3` | Update Scope blurb + add §4 weekly-reflection subsection (optional but recommended) |
| `docs/api-contract.md:315` | Update DEC-014 References row + add new `### brag review` section |

No discoveries expected at build time. Build re-runs all the rg
commands above and questions any delta from this enumerated list
(per the audit-grep cross-check both-sides discipline).

**Existing-test audit** (addition-case doesn't add tracked-count
coupling; verify nothing breaks):

- `internal/cli/summary_test.go` — assertions on summary's
  `--format` needles `markdown`, `json` are summary-scoped.
  Unaffected by review's separate flag set.
- `internal/cli/list_test.go`, `export_test.go`, `show_test.go`,
  `add_test.go`, `add_json_test.go`, `edit_test.go`,
  `delete_test.go`, `search_test.go` — no overlap with review
  surface.
- `internal/export/summary_test.go`, `markdown_test.go`,
  `json_test.go` — sibling files in the package; new
  `review_test.go` lands alongside, no cross-coupling.
- `internal/aggregate/aggregate_test.go` — gains two tests at the
  end of the file; existing four tests
  (`TestByType_DESCByCountAlphaTiebreak`,
  `TestByProject_NoProjectKeyForcedLast`,
  `TestGroupForHighlights_ChronoASCWithNoProjectLast`,
  `TestAggregate_EmptyInputReturnsNonNilEmptySlice`) stay
  byte-identical.
- `internal/storage/*_test.go` — read-only path; review uses
  existing `Store.List`. Unaffected.

## Failing Tests

Written now, during **design**. Twelve tests total across **2 new
files** plus **2 tests appended to `internal/aggregate/aggregate_test.go`**.
All follow AGENTS.md §9: separate `outBuf` / `errBuf` with no-cross-
leakage asserts; fail-first run before implementation; assertion
specificity on help/error substrings; every locked decision paired
with at least one failing test; line-based equality (not
`strings.Contains`) for any heading-level assertion (SPEC-015
substring-trap addendum); ID-based (not timestamp-based) distinctness
for any freshness or ordering tie-break (SPEC-017 freshness-assertion
addendum).

Goldens reuse a single fixture so all renderer choices anchor to the
same canonical entries. Aggregate tests use the shared fixture for
the ordering case and a tie-break-only fixture for the
ID-tiebreak case.

### Shared renderer fixture (used by tests 3, 4, 5, 6)

```go
// 5 entries spanning 3 projects + (no project), with chrono ordering
// chosen to exercise:
//   * within-alpha chrono-ASC: 1 (T1), 4 (T4) [IDs are NOT monotonic
//     with timestamps so ID tie-break is testable in the aggregate
//     helper test]
//   * (no project) forced last regardless of count
//   * one entry with description (entry 1) to exercise markdown elision
//     (description NOT in markdown bytes) AND JSON inclusion
//     (description IN JSON bytes) — the locked asymmetry from
//     decision §3.
var fixture = []storage.Entry{
    {
        ID: 1, Title: "alpha-old",
        Description: "did the auth refactor",  // NOT rendered in
                                                 // review markdown;
                                                 // RENDERED in review
                                                 // JSON.
        Tags: "auth", Project: "alpha", Type: "shipped",
        Impact: "unblocked mobile",
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

Group ordering on this fixture (alpha-ASC by project, `(no project)`
last; chrono-ASC within group):
- alpha → [1: alpha-old (T1), 4: alpha-new (T4)]
- beta → [2: beta-mid (T2)]
- gamma → [5: gamma-only (T5)]
- (no project) → [3: unbound-mid (T3)]

### `internal/aggregate/aggregate_test.go` (existing file — 2 tests appended)

The new helper tests append after the existing four tests in the
same file, sharing imports and any package-local helpers.

#### Test 1 — `TestGroupEntriesByProject_OrderingAndIDTiebreak`

Two subtests covering ordering on the shared fixture and the
ID-tiebreak case.

**Subtest `shared_fixture`:** call `got :=
GroupEntriesByProject(sharedFixture)` (the same `sharedFixture`
SPEC-018 introduced in `aggregate_test.go`). Expected:

```go
want := []ProjectEntryGroup{
    {Project: "alpha", Entries: []storage.Entry{
        {ID: 1, Title: "alpha-old", /* ... full fixture entry 1 */},
        {ID: 4, Title: "alpha-new", /* ... full fixture entry 4 */},
    }},
    {Project: "beta", Entries: []storage.Entry{
        {ID: 2, Title: "beta-mid", /* ... full fixture entry 2 */},
    }},
    {Project: "gamma", Entries: []storage.Entry{
        {ID: 5, Title: "gamma-only", /* ... full fixture entry 5 */},
    }},
    {Project: NoProjectKey, Entries: []storage.Entry{
        {ID: 3, Title: "unbound-mid", /* ... full fixture entry 3 */},
    }},
}
```

Assert `reflect.DeepEqual(got, want)`. Asserts: alpha-ASC group
order; `NoProjectKey` last; chrono-ASC within alpha (entry 1 before
entry 4 because T1 < T4); FULL `storage.Entry` carried (not the
`EntryRef{ID, Title}` projection — verified by `Description` /
`Tags` / `Impact` fields surviving in the comparison).

**Subtest `id_tiebreak`:** build a fixture where two alpha entries
share `CreatedAt` but have different IDs:

```go
tieFixture := []storage.Entry{
    {ID: 99, Title: "later-id-same-time", Project: "alpha",
     CreatedAt: time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC)},
    {ID: 7,  Title: "earlier-id-same-time", Project: "alpha",
     CreatedAt: time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC)},
}
```

Call `got := GroupEntriesByProject(tieFixture)`. Expected:

```go
want := []ProjectEntryGroup{
    {Project: "alpha", Entries: []storage.Entry{
        {ID: 7, Title: "earlier-id-same-time", /* ... */},
        {ID: 99, Title: "later-id-same-time", /* ... */},
    }},
}
```

Assert `reflect.DeepEqual(got, want)`. Lower ID wins when timestamps
tie — locks AGENTS.md §9 SPEC-002 monotonic-tiebreak rule (RFC3339
is second-precision; timestamps can tie within the same second; ID
is the deterministic tie-breaker).

Pairs locked decision §9 (single-helper extension; sort logic
mirrors `GroupForHighlights`).

#### Test 2 — `TestGroupEntriesByProject_EmptyInputReturnsNonNilEmptySlice`

Call `GroupEntriesByProject(nil)` AND
`GroupEntriesByProject([]storage.Entry{})`. For each:

1. Assert `got != nil` (non-nil slice — important so JSON marshaling
   renders `[]` not `null`).
2. Assert `len(got) == 0`.

Pairs locked decision §1 part (4) (empty-state arrays are `[]` in
JSON, never `null`) on the new helper. The renderer-side empty-
state test (#5) verifies the corresponding JSON output literally
renders `[]` for `entries_grouped`.

### `internal/export/review_test.go` (new file — 4 tests)

Tests against the shared fixture + explicit `ReviewOptions`. No
cobra, no DB.

#### Test 3 — `TestToReviewMarkdown_DEC014FullDocumentGolden` (LOAD-BEARING — write FIRST)

Build `opts := ReviewOptions{Scope: "week", Now: fixedNow}`. Call
`got, err := ToReviewMarkdown(fixture, opts)`. Assert `err == nil`
and `bytes.Equal(got, []byte(want))` where `want` is the literal
string below. If this fails, DEC-014 has been violated — fix code,
not test.

Expected output (byte-exact, no trailing newline — the CLI layer
adds one via `fmt.Fprintln`):

```
# Bragfile Review

Generated: 2026-04-25T12:00:00Z
Scope: week
Filters: (none)

## Entries

### alpha

- 1: alpha-old
- 4: alpha-new

### beta

- 2: beta-mid

### gamma

- 5: gamma-only

### (no project)

- 3: unbound-mid

## Reflection questions

1. What pattern do you see in this period?
2. What did you underestimate?
3. What's missing here that should be?
```

Notes on the layout:
- `# Bragfile Review` is the document title (parallels `# Bragfile
  Summary`).
- Provenance block reuses DEC-014's three lines (`Generated:`,
  `Scope:`, `Filters:`). `Filters: (none)` is hard-coded — review
  doesn't accept filter flags so the value is constant.
- `## Entries` is the wrapper around per-project groups. Parallels
  summary's `## Highlights`. (The user-confirmed `## Entries`
  wording — see SPEC-019 Q2 design clarification.)
- `### <project>` per group (level-3 — parallels summary's
  `### <project>` under `## Highlights`).
- Bulleted `- <id>: <title>` per entry (titles + IDs only; descriptions
  ELIDED — locked decision §3).
- `(no project)` last regardless of count (DEC-013 + DEC-014 +
  `aggregate.NoProjectKey` constant).
- `## Reflection questions` section — heading at level 2, parallels
  `## Entries`. Three questions as a numbered list (`1.`, `2.`,
  `3.`).
- Verbatim wording of the three questions per locked decision §6.

Assertions:
1. `bytes.Equal(got, []byte(want))` — single exact-byte compare.
2. On failure, print both `want` and `got` for diffability (copy
   the SPEC-014 / SPEC-015 / SPEC-018 helper pattern).
3. Line-based assertion on heading levels per AGENTS.md §9
   substring-trap addendum (SPEC-015 lesson): split `got` into
   lines and assert `lines[0] == "# Bragfile Review"` and that
   `"## Entries"` and `"## Reflection questions"` each appear as
   standalone lines, not as substrings of deeper headings.
4. ELISION lock: assert `!bytes.Contains(got, []byte("did the auth
   refactor"))` — entry 1's `Description` field MUST NOT appear in
   the markdown bytes. This is the load-bearing assertion for the
   "markdown elides descriptions" half of locked decision §3.

Pairs locked decisions §1, §3, §6.

#### Test 4 — `TestToReviewJSON_DEC014ShapeGolden` (LOAD-BEARING — write SECOND)

Build `opts := ReviewOptions{Scope: "week", Now: fixedNow}`. Call
`got, err := ToReviewJSON(fixture, opts)`. Assert `err == nil` and
`bytes.Equal(got, []byte(want))` where `want` is the literal JSON
below.

Expected output (byte-exact, no trailing newline):

```json
{
  "generated_at": "2026-04-25T12:00:00Z",
  "scope": "week",
  "filters": {},
  "entries_grouped": [
    {
      "project": "alpha",
      "entries": [
        {
          "id": 1,
          "title": "alpha-old",
          "description": "did the auth refactor",
          "tags": "auth",
          "project": "alpha",
          "type": "shipped",
          "impact": "unblocked mobile",
          "created_at": "2026-04-20T10:00:00Z",
          "updated_at": "2026-04-20T10:00:00Z"
        },
        {
          "id": 4,
          "title": "alpha-new",
          "description": "",
          "tags": "",
          "project": "alpha",
          "type": "shipped",
          "impact": "",
          "created_at": "2026-04-23T10:00:00Z",
          "updated_at": "2026-04-23T10:00:00Z"
        }
      ]
    },
    {
      "project": "beta",
      "entries": [
        {
          "id": 2,
          "title": "beta-mid",
          "description": "",
          "tags": "",
          "project": "beta",
          "type": "learned",
          "impact": "",
          "created_at": "2026-04-21T10:00:00Z",
          "updated_at": "2026-04-21T10:00:00Z"
        }
      ]
    },
    {
      "project": "gamma",
      "entries": [
        {
          "id": 5,
          "title": "gamma-only",
          "description": "",
          "tags": "",
          "project": "gamma",
          "type": "fixed",
          "impact": "",
          "created_at": "2026-04-24T10:00:00Z",
          "updated_at": "2026-04-24T10:00:00Z"
        }
      ]
    },
    {
      "project": "(no project)",
      "entries": [
        {
          "id": 3,
          "title": "unbound-mid",
          "description": "",
          "tags": "",
          "project": "",
          "type": "shipped",
          "impact": "",
          "created_at": "2026-04-22T10:00:00Z",
          "updated_at": "2026-04-22T10:00:00Z"
        }
      ]
    }
  ],
  "reflection_questions": [
    "What pattern do you see in this period?",
    "What did you underestimate?",
    "What's missing here that should be?"
  ]
}
```

Notes on the JSON shape:
- Top-level keys appear in struct-tag declaration order
  (`generated_at`, `scope`, `filters`, `entries_grouped`,
  `reflection_questions`) — Go's `encoding/json` preserves struct-
  tag order in `MarshalIndent`. Locked decision §1 part (2).
- `filters: {}` is an empty object (review never accepts filter
  flags; the value is constant). Locked decision §1 part (4).
- `entries_grouped` is a JSON array of `{project, entries}` objects
  in DEC-014-locked group order (alpha-ASC by project name with
  `(no project)` last). Locked decision §2.
- Each object inside `entries[...]` is the FULL DEC-011 9-key shape:
  `id`, `title`, `description`, `tags`, `project`, `type`, `impact`,
  `created_at`, `updated_at`. Locked decision §3 (JSON includes
  descriptions). Note the inner `project` field on entry 3 (the
  `(no project)` case): the inner field is `""` (the storage layer
  value), while the outer `project` key on the wrapping group is
  `"(no project)"` (the rendered sentinel). Both are correct;
  the rendering layer projects the empty-string to the sentinel
  for human display, but the per-entry data preserves the storage
  truth.
- `reflection_questions` is a JSON array of three strings.
  Verbatim wording per locked decision §6.
- Indent=2 per DEC-014 choice (5).

Assertions:
1. `bytes.Equal(got, []byte(want))` — single exact-byte compare.
2. Parse `got` via `json.Decoder` and verify object's top-level
   keys appear in struct-tag declaration order (`generated_at`,
   `scope`, `filters`, `entries_grouped`, `reflection_questions`).
   This is the load-bearing key-order assertion that DEC-014 part
   (2) rests on.
3. INCLUSION lock: parse `got` and assert
   `m["entries_grouped"][0]["entries"][0]["description"] == "did
   the auth refactor"`. This is the load-bearing assertion for the
   "JSON includes descriptions" half of locked decision §3 and the
   `toEntryRecord` helper extraction in `json.go` (decision §10).

Pairs locked decisions §1, §2, §3, §6, §7, §10.

#### Test 5 — `TestToReview_EmptyEntriesStillEmitsReflectionQuestions`

Subtests `markdown` and `json`. The distinguishing contract: review's
`## Reflection questions` block ALWAYS renders, even on empty
entries — questions are the *point* of the command, not a payload
derived from entries.

**Subtest `markdown`:** render `ToReviewMarkdown([]storage.Entry{},
ReviewOptions{Scope: "week", Now: fixedNow})`. Expected bytes
(byte-exact):

```
# Bragfile Review

Generated: 2026-04-25T12:00:00Z
Scope: week
Filters: (none)

## Reflection questions

1. What pattern do you see in this period?
2. What did you underestimate?
3. What's missing here that should be?
```

NO `## Entries` heading. NO per-project group headings. Document
ends after the third question line (trailing newline stripped by
the byte-contract).

Assertions:
1. `bytes.Equal(got, []byte(want))`.
2. Line-split + walk: assert NO line equals `"## Entries"`
   (line-based, not `strings.Contains`, per SPEC-015 substring-
   trap addendum).
3. Line-split + walk: assert ONE line equals `"## Reflection
   questions"` — questions appear even with zero entries.

**Subtest `json`:** render `ToReviewJSON([]storage.Entry{},
ReviewOptions{Scope: "week", Now: fixedNow})`. Expected bytes
(byte-exact):

```json
{
  "generated_at": "2026-04-25T12:00:00Z",
  "scope": "week",
  "filters": {},
  "entries_grouped": [],
  "reflection_questions": [
    "What pattern do you see in this period?",
    "What did you underestimate?",
    "What's missing here that should be?"
  ]
}
```

Assertions:
1. `bytes.Equal(got, []byte(want))`.
2. Parse via `json.Unmarshal` into `map[string]any` and:
   - assert `m["entries_grouped"]` is a `[]any` with `len == 0`
     (NOT nil, NOT `null`) — locks empty-state non-nil rule.
   - assert `m["reflection_questions"].([]any)` has `len == 3` —
     questions render even when entries are empty.

Pairs locked decisions §1 part (4), §3, §6.

This test diverges from SPEC-018's `TestToSummary_EmptyEntriesEmitsProvenanceOnly`
because review's payload has a non-entry-derived part (questions)
that doesn't elide; summary's payload was entirely entry-derived
so the entire summary + highlights block elides on empty.

#### Test 6 — `TestToReview_ReflectionQuestionsExactWording`

Subtests `markdown` and `json`. Locks the EXACT verbatim wording of
the three reflection questions per stage Design Notes lines 367–369
(NOT the looser Success Criteria paraphrase).

**Subtest `markdown`:** render `ToReviewMarkdown(fixture, opts)`
with shared opts. Line-walk the output and assert these three
lines appear in this exact order:

```
1. What pattern do you see in this period?
2. What did you underestimate?
3. What's missing here that should be?
```

Use `lines[i] == "1. What pattern do you see in this period?"`
form (line-based equality), not `strings.Contains` (SPEC-015
substring-trap discipline).

**Subtest `json`:** render `ToReviewJSON(fixture, opts)`. Parse
via `json.Unmarshal` into a struct with `ReflectionQuestions
[]string \`json:"reflection_questions"\``. Assert
`reflect.DeepEqual(parsed.ReflectionQuestions, []string{
    "What pattern do you see in this period?",
    "What did you underestimate?",
    "What's missing here that should be?",
})`.

Pairs locked decision §6.

### `internal/cli/review_test.go` (new file — 6 tests)

Reuse `seedListEntry` patterns from `internal/cli/list_test.go`
(same package, so direct reuse). Use `t.TempDir()` for DB paths;
never touch `~/.bragfile`.

#### Test 7 — `TestReviewCmd_BareDefaultsToWeek`

Run `brag review` (no flags). Seed two fresh entries first via
`seedListEntry` so the entries-section renders (the markdown body
exists; we check the scope-line and headers).

Assertions:
1. `err == nil`; `errBuf.Len() == 0` (silent default — locked
   decision §5).
2. `outBuf.String()` starts with `"# Bragfile Review\n\n"`
   (markdown default — locked decision §1; reuses summary's
   default-format pattern).
3. Line-walk `outBuf.String()`:
   - one line equals `"Scope: week"` (silent default to `--week`).
   - one line equals `"## Entries"` (entries section present
     because we seeded entries).
   - one line equals `"## Reflection questions"` (always present).
4. `Generated:` line matches the RFC3339 regex
   (`^Generated: \d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`) — reuse
   the `lineMatches` + `regexp.MustCompile` helpers from
   `summary_test.go` (same package, source-level reuse).

Pairs locked decisions §1 (default markdown + scope plumb-through),
§4 (named flags), §5 (silent default).

#### Test 8 — `TestReviewCmd_WeekAndMonthMutuallyExclusiveIsUserError`

Run `brag review --week --month`. Assertions:
1. `err != nil`; `errors.Is(err, ErrUser) == true`.
2. `outBuf.String() == ""` (no leakage).
3. The error message (`err.Error()`) contains `--week` AND
   `--month` (each as distinctive needles per AGENTS.md §9
   assertion-specificity rule). The wording also conveys
   mutual exclusion (e.g., "cannot use --week and --month
   together" or "mutually exclusive"); assert one such
   needle (`mutually exclusive` or `together` — pick one;
   the implementer chooses the wording, the test asserts
   what was chosen).

   *Implementer note:* lock the wording in this test's expected
   needle to `mutually exclusive` to avoid wording drift later.
   The implementation message MUST contain the literal phrase
   `mutually exclusive`.

Pairs locked decision §4 (DEC-007 RunE validation, NOT
`MarkFlagsMutuallyExclusive`).

#### Test 9 — `TestReviewCmd_UnknownFormatIsUserError`

Run `brag review --week --format yaml`. Assertions:
1. `err != nil`; `errors.Is(err, ErrUser) == true`.
2. `outBuf.String() == ""`.
3. Error message contains `yaml` AND `markdown` AND `json`.

Pairs locked decision §1 (DEC-014 + DEC-007 — `--format`
validation in RunE).

#### Test 10 — `TestReviewCmd_FormatJSON_MonthScopeAndEntriesGroupedShape`

Seed three fresh entries via `seedListEntry`:
- `{Title: "alpha-1", Project: "alpha", Type: "shipped"}`
- `{Title: "alpha-2", Project: "alpha", Type: "learned"}`
- `{Title: "beta-1", Project: "beta", Type: "shipped"}`

Run `brag review --month --format json`. Assertions:
1. `err == nil`; `errBuf.Len() == 0`.
2. `outBuf.String()` ends with `"\n"` (trailing newline from
   `fmt.Fprintln`).
3. Body (`strings.TrimRight(out, "\n")`) parses as JSON via
   `json.Unmarshal` into `map[string]any`.
4. Top-level keys: `generated_at`, `scope`, `filters`,
   `entries_grouped`, `reflection_questions`. Assert each
   key present.
5. `m["scope"] == "month"` (locked decision §1 part 6 + §4 —
   month flag plumbs through).
6. `m["filters"]` is a `map[string]any` with `len == 0` (always
   `{}` — locked decision §8).
7. `m["entries_grouped"]` is a `[]any` with `len == 2` (alpha
   group + beta group; deterministic alpha-ASC order).
8. `groups[0]["project"] == "alpha"`; `groups[0]["entries"]`
   has `len == 2`.
9. `groups[1]["project"] == "beta"`; `groups[1]["entries"]`
   has `len == 1`.
10. The entries inside `groups[0]["entries"]` are full DEC-011
    9-key objects: each has `id`, `title`, `description`, `tags`,
    `project`, `type`, `impact`, `created_at`, `updated_at`. Walk
    the keys and assert presence (locked decision §3 — JSON
    includes full entry shape).
11. `m["reflection_questions"].([]any)` has `len == 3`; each
    element is a string (don't assert exact wording here — test
    #6 covers wording).

Pairs locked decisions §1, §2, §3, §4, §10.

#### Test 11 — `TestReviewCmd_HelpShowsWeekMonthAndFormat`

Run `brag review --help`. Assertions:
1. `err == nil`; `errBuf.Len() == 0`.
2. `outBuf.String()` contains needles `--week`, `--month`,
   `--format`, `markdown`, `json` (each as distinctive needles
   per AGENTS.md §9 assertion-specificity rule).
3. `outBuf.String()` DOES NOT contain `--tag`, `--project`,
   `--type`, `--out`, `--since` — these flags are NOT declared
   on review. Use line-based assertion (don't accept generic
   substring "tag" in prose; check the `--tag` literal absence
   line-by-line). Locks decision §8 at the help-surface level.

Pairs locked decisions §4, §5 (help mentions default), §8 (no
filter flags advertised).

#### Test 12 — `TestReviewCmd_FilterAndOutFlagsRejectedAsUnknown`

Subtests for each undeclared flag: `--tag`, `--project`, `--type`,
`--out`, `--since`. For each:

- Run `brag review --week --<flag> X`.
- Assert `err != nil`.
- Assert `errors.Is(err, ErrUser) == false` — cobra's
  `unknown flag` error is NOT an `ErrUser` (it's cobra's own
  error type wrapped through). Assert by negation rather than
  positive identity, since the exact error type cobra returns
  is implementation detail.
- Assert `err.Error()` contains the literal `unknown flag` AND
  the offending flag name (e.g., `--tag`).

Locks decision §8 — filter and out flags are explicitly
undeclared, so cobra's parser surfaces them as unknown rather
than the command rejecting them in `RunE`.

## Implementation Context

### Decisions that apply

- `DEC-014` — rule-based output shape (envelope JSON + markdown
  provenance) consumed verbatim. SPEC-019 is the second consumer;
  goldens prove conformance. NO new DEC; if you think you need
  one, STOP and ask.
- `DEC-013` — markdown export shape; provenance + summary block
  conventions DEC-014 inherits. Review's `## Entries` section
  parallels summary's `## Highlights`; both inherit DEC-013's
  `(no project)` last + DESC-by-count + alpha-ASC tiebreak rules
  through DEC-014.
- `DEC-011` — naked-array per-entry shape. Each item inside
  `entries_grouped[].entries` is a DEC-011 9-key entry record.
  The wrapper is the DEC-014 envelope, not the DEC-011 array;
  the two compose without conflict (envelope contains arrays of
  arrays of records).
- `DEC-007` — required-flag validation in `RunE`. Applies to
  `--week`/`--month` mutex AND `--format` validation. NEVER use
  `MarkFlagRequired` or `MarkFlagsMutuallyExclusive`; always
  `UserErrorf` in `RunE`.
- `DEC-006` — cobra framework. New `brag review` is a cobra
  subcommand following the pattern of every other.

### Constraints that apply

These constraints apply to the paths touched by this task (see
`/guidance/constraints.yaml` for full text):

- `no-sql-in-cli-layer` — the new CLI file `internal/cli/review.go`
  MUST NOT import `database/sql`. Storage access goes through
  `storage.Open(...).List(...)`. Mirrors summary's structure.
- `stdout-is-for-data-stderr-is-for-humans` — markdown/JSON bodies
  to stdout via `cmd.OutOrStdout()`. Errors via the cobra return
  path; `main.go` writes them to stderr. Tests assert
  `errBuf.Len() == 0` on success paths.
- `errors-wrap-with-context` — `fmt.Errorf("...: %w", err)` for
  wrapped errors. `UserErrorf` for user-error paths.
- `test-before-implementation` — write the twelve failing tests
  first, run `go test ./...`, confirm the expected failure modes
  (compile errors for missing types/funcs, assertion failures
  for the goldens), THEN implement.
- `one-spec-per-pr` — one feature branch + one PR for SPEC-019.

### Prior related work

- `SPEC-018` (shipped 2026-04-25) — direct precedent. SPEC-018
  emitted DEC-014, seeded `internal/aggregate`, created
  `internal/export/summary.go`, and established the Rejected-
  alternatives-build-time discipline this spec re-applies.
- `SPEC-014` (shipped) — emitted DEC-011, created the
  `internal/export` package, defined `entryRecord` SPEC-019
  extracts the helper from. Structural precedent for the
  goldens-locked DEC pattern.
- `SPEC-015` (shipped) — markdown export with DEC-013 provenance/
  summary conventions DEC-014 inherits. SPEC-015's substring-
  trap addendum applies to this spec's heading-level asserts.
- `SPEC-017` (shipped) — earned the §12 "decide at design time
  when decidable" lesson SPEC-018 first applied proactively;
  SPEC-019 is the second proactive application.
- `SPEC-007` (shipped) — `ListFilter{Since: ...}` is the only
  filter field review uses; `--tag`/`--project`/`--type` are NOT
  declared on review per stage Design Notes.

### Out of scope (for this spec specifically)

Explicit list of what this spec does NOT include. If any of these
feel necessary during build, create a new spec rather than
expanding this one.

- DEC emission. DEC-014 is consumed verbatim. Any "I think we
  need a DEC for X" thought during build is a STOP-and-ask.
- Refactor of `GroupForHighlights` into a parameterized helper.
  Sound future refactor; out-of-scope for this spec. Capture as
  a backlog candidate IF a third consumer with the same grouping
  shape but different per-row projection appears (third caller
  is the threshold per the established rule).
- Export of `entryRecord` (`EntryRecord`). Package-private is
  sufficient.
- Any new aggregate helper beyond `GroupEntriesByProject`.
- Any change to `summary.go` or `summary_test.go` semantics.
- Any change to `ToJSON` byte output. The `toEntryRecord`
  extraction is a no-op refactor; existing
  `internal/export/json_test.go` goldens MUST stay byte-
  identical.
- Modification of the parent stage doc (the wording-mismatch
  between Success Criteria and Design Notes is a stage-level
  lesson; not edited by child specs).
- Any expansion of `--format` accepted values beyond `markdown`
  and `json` (e.g., `tsv`, `csv`, `yaml`). Same set as summary.
- Any addition of `--out` / `--tag` / `--project` / `--type` /
  `--since`. Locked NO.

## Notes for the Implementer

Gotchas, style preferences, reuse opportunities.

### Aggregate helper

`internal/aggregate/aggregate.go` gains:

```go
// ProjectEntryGroup carries one project's full entries, used by
// brag review (SPEC-019). Mirrors ProjectHighlights's shape but
// retains the full storage.Entry instead of the EntryRef
// projection — JSON consumers (DEC-011 9-key per-entry shape)
// need descriptions and metadata that highlights elides.
type ProjectEntryGroup struct {
    Project string
    Entries []storage.Entry
}

// GroupEntriesByProject mirrors GroupForHighlights's grouping +
// sort logic exactly: alpha-ASC by project name with NoProjectKey
// last; chrono-ASC within group with ID as tiebreak. Differs only
// in carrying full storage.Entry (not EntryRef). Used by review's
// markdown path (renders id+title only at render time) and
// review's JSON path (serializes full DEC-011 shape).
func GroupEntriesByProject(entries []storage.Entry) []ProjectEntryGroup {
    buckets := make(map[string][]storage.Entry)
    for _, e := range entries {
        key := e.Project
        if key == "" {
            key = NoProjectKey
        }
        buckets[key] = append(buckets[key], e)
    }
    out := make([]ProjectEntryGroup, 0, len(buckets))
    for proj, group := range buckets {
        sorted := make([]storage.Entry, len(group))
        copy(sorted, group)
        sort.SliceStable(sorted, func(i, j int) bool {
            if !sorted[i].CreatedAt.Equal(sorted[j].CreatedAt) {
                return sorted[i].CreatedAt.Before(sorted[j].CreatedAt)
            }
            return sorted[i].ID < sorted[j].ID
        })
        out = append(out, ProjectEntryGroup{Project: proj, Entries: sorted})
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

Note the structural symmetry to `GroupForHighlights` — only the
return type's `Entries` field differs. A future spec with a third
consumer of the same grouping logic may motivate a parameterized
refactor; for SPEC-019, the duplication is acceptable (small,
mechanical, mirror-image of the existing function).

### `toEntryRecord` extraction in `json.go`

Before:

```go
func ToJSON(entries []storage.Entry) ([]byte, error) {
    if len(entries) == 0 {
        return []byte("[]"), nil
    }
    records := make([]entryRecord, 0, len(entries))
    for _, e := range entries {
        records = append(records, entryRecord{
            ID:          e.ID,
            Title:       e.Title,
            // ... 9 fields ...
        })
    }
    return json.MarshalIndent(records, "", "  ")
}
```

After:

```go
// toEntryRecord projects a storage.Entry to the DEC-011 9-key
// serialization shape. Both ToJSON (list/export consumer) and
// internal/export/review.go's ToReviewJSON (review consumer) use
// this helper — sharing the field map keeps drift impossible.
// Stays package-private; consumers outside internal/export should
// not depend on the shape of a serialization detail.
func toEntryRecord(e storage.Entry) entryRecord {
    return entryRecord{
        ID:          e.ID,
        Title:       e.Title,
        Description: e.Description,
        Tags:        e.Tags,
        Project:     e.Project,
        Type:        e.Type,
        Impact:      e.Impact,
        CreatedAt:   e.CreatedAt.UTC().Format(time.RFC3339),
        UpdatedAt:   e.UpdatedAt.UTC().Format(time.RFC3339),
    }
}

func ToJSON(entries []storage.Entry) ([]byte, error) {
    if len(entries) == 0 {
        return []byte("[]"), nil
    }
    records := make([]entryRecord, 0, len(entries))
    for _, e := range entries {
        records = append(records, toEntryRecord(e))
    }
    return json.MarshalIndent(records, "", "  ")
}
```

NO behavior change. Run existing `internal/export/json_test.go`
goldens (`TestToJSON_*`) after the extraction; they should all
pass byte-identical.

### `internal/export/review.go` skeleton

```go
package export

import (
    "bytes"
    "encoding/json"
    "fmt"
    "time"

    "github.com/jysf/bragfile000/internal/aggregate"
    "github.com/jysf/bragfile000/internal/storage"
)

// reflectionQuestions are the three hard-coded prompts SPEC-019
// locks. Wording verbatim from STAGE-004 Design Notes lines
// 367–369. Configurability is backlogged with revisit trigger
// "user wants to swap one out."
var reflectionQuestions = []string{
    "What pattern do you see in this period?",
    "What did you underestimate?",
    "What's missing here that should be?",
}

// ReviewOptions controls ToReviewMarkdown / ToReviewJSON. Scope
// is "week" or "month" (echoed into provenance + envelope). Now
// is injected for deterministic Generated: lines (mirrors
// MarkdownOptions.Now + SummaryOptions.Now). No Filters field —
// review does not accept filter flags; "(none)" is hard-coded
// in the markdown provenance line and {} is hard-coded in the
// JSON envelope.
type ReviewOptions struct {
    Scope string
    Now   time.Time
}

// ToReviewMarkdown renders the DEC-014 markdown digest for
// brag review. Returns bytes with trailing "\n" stripped (matches
// the byte contract of the other renderers). The Reflection
// questions block ALWAYS renders, even on empty entries — the
// questions are the point of the command.
func ToReviewMarkdown(entries []storage.Entry, opts ReviewOptions) ([]byte, error) {
    var buf bytes.Buffer
    fmt.Fprintln(&buf, "# Bragfile Review")
    fmt.Fprintln(&buf)
    fmt.Fprintf(&buf, "Generated: %s\n", opts.Now.UTC().Format(time.RFC3339))
    fmt.Fprintf(&buf, "Scope: %s\n", opts.Scope)
    fmt.Fprintln(&buf, "Filters: (none)")

    if len(entries) > 0 {
        fmt.Fprintln(&buf)
        fmt.Fprintln(&buf, "## Entries")
        for _, group := range aggregate.GroupEntriesByProject(entries) {
            fmt.Fprintln(&buf)
            fmt.Fprintf(&buf, "### %s\n", group.Project)
            fmt.Fprintln(&buf)
            for _, e := range group.Entries {
                fmt.Fprintf(&buf, "- %d: %s\n", e.ID, e.Title)
            }
        }
    }

    fmt.Fprintln(&buf)
    fmt.Fprintln(&buf, "## Reflection questions")
    fmt.Fprintln(&buf)
    for i, q := range reflectionQuestions {
        fmt.Fprintf(&buf, "%d. %s\n", i+1, q)
    }
    return trimTrailingNewline(buf.Bytes()), nil
}

// reviewEnvelope is the on-the-wire JSON shape for ToReviewJSON.
// Field order locks DEC-014's top-level key order
// (generated_at, scope, filters, entries_grouped,
// reflection_questions) via struct-tag declaration order.
type reviewEnvelope struct {
    GeneratedAt         string                `json:"generated_at"`
    Scope               string                `json:"scope"`
    Filters             map[string]string     `json:"filters"`
    EntriesGrouped      []reviewProjectGroup  `json:"entries_grouped"`
    ReflectionQuestions []string              `json:"reflection_questions"`
}

type reviewProjectGroup struct {
    Project string        `json:"project"`
    Entries []entryRecord `json:"entries"`
}

// ToReviewJSON renders the DEC-014 envelope for brag review. Per-
// entry shape inside entries_grouped[].entries is the DEC-011
// 9-key shape (via the toEntryRecord helper).
func ToReviewJSON(entries []storage.Entry, opts ReviewOptions) ([]byte, error) {
    env := reviewEnvelope{
        GeneratedAt:         opts.Now.UTC().Format(time.RFC3339),
        Scope:               opts.Scope,
        Filters:             map[string]string{},
        EntriesGrouped:      []reviewProjectGroup{},
        ReflectionQuestions: append([]string{}, reflectionQuestions...),
    }
    for _, group := range aggregate.GroupEntriesByProject(entries) {
        rg := reviewProjectGroup{
            Project: group.Project,
            Entries: make([]entryRecord, 0, len(group.Entries)),
        }
        for _, e := range group.Entries {
            rg.Entries = append(rg.Entries, toEntryRecord(e))
        }
        env.EntriesGrouped = append(env.EntriesGrouped, rg)
    }
    return json.MarshalIndent(env, "", "  ")
}
```

Notes:
- The `reflectionQuestions` slice is COPIED into the envelope
  (`append([]string{}, ...)` rather than direct assignment) so a
  consumer mutating `env.ReflectionQuestions` cannot mutate the
  package-level `var`. Defensive but cheap; the alternative
  (return-the-package-level-slice) is unsafe across goroutines
  and is the kind of tiny correctness lock that future refactors
  appreciate.
- `entries_grouped` is initialized to `[]reviewProjectGroup{}`
  (non-nil empty) so `MarshalIndent` renders `[]` not `null` on
  empty input. Locks DEC-014 part (4).
- The wrapping `## Reflection questions` heading uses the lower-
  case "questions" suffix (per the locked markdown shape under
  decision §3 + Test #3 expected). Don't capitalize "Questions"
  — the test asserts byte-exact.

### `internal/cli/review.go` skeleton

```go
package cli

import (
    "fmt"
    "time"

    "github.com/jysf/bragfile000/internal/config"
    "github.com/jysf/bragfile000/internal/export"
    "github.com/jysf/bragfile000/internal/storage"
    "github.com/spf13/cobra"
)

// NewReviewCmd returns the `brag review` subcommand. SPEC-019 emits
// it as the second DEC-014 consumer: a reflection-prompt digest
// of recent entries grouped by project, followed by three hard-
// coded reflection questions. No LLM.
func NewReviewCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "review",
        Short: "Reflection digest: recent entries grouped by project + three reflection questions",
        Long: `Print a reflection digest of recent entries grouped by project, followed by three hard-coded reflection questions designed to be pasted into an external AI session for guided self-review. No LLM ships in the binary.

Output is markdown (default) or a single-object JSON envelope (--format json) per DEC-014. The JSON shape mirrors brag summary's envelope (entries_grouped is an array of {project, entries: [...]} objects).

--week and --month are mutually exclusive named flags. Bare 'brag review' silently defaults to --week. Rolling-window semantics: --week = last 7 UTC days; --month = last 30 UTC days.

Filter flags (--tag, --project, --type) are NOT accepted on review — the digest is "the last 7/30 days, period." Stdout only; no --out flag (redirect with > if you want a file).

Examples:
  brag review                                  # last 7 UTC days, markdown (silent default)
  brag review --week                           # explicit; same as bare invocation
  brag review --month --format json            # last 30 UTC days, JSON envelope`,
        RunE: runReview,
    }
    cmd.Flags().Bool("week", false, "review last 7 UTC days (default if neither --week nor --month is set)")
    cmd.Flags().Bool("month", false, "review last 30 UTC days")
    cmd.Flags().String("format", "markdown", "output format (one of: markdown, json)")
    return cmd
}

func runReview(cmd *cobra.Command, _ []string) error {
    weekSet := cmd.Flags().Changed("week")
    monthSet := cmd.Flags().Changed("month")
    if weekSet && monthSet {
        return UserErrorf("--week and --month are mutually exclusive (use one or neither; neither defaults to --week)")
    }

    scope := "week"
    if monthSet {
        scope = "month"
    }
    cutoff, err := rangeCutoff(scope, time.Now().UTC())
    if err != nil {
        // rangeCutoff only errors on empty/unknown scope; both
        // "week" and "month" are valid here. Defensive:
        return fmt.Errorf("compute cutoff: %w", err)
    }

    format, _ := cmd.Flags().GetString("format")
    if format != "markdown" && format != "json" {
        return UserErrorf("unknown --format value %q (accepted: markdown, json)", format)
    }

    dbFlag := getFlagString(cmd, "db")
    path, err := config.ResolveDBPath(dbFlag)
    if err != nil {
        return fmt.Errorf("resolve db path: %w", err)
    }

    s, err := storage.Open(path)
    if err != nil {
        return fmt.Errorf("open store: %w", err)
    }
    defer s.Close()

    entries, err := s.List(storage.ListFilter{Since: cutoff})
    if err != nil {
        return fmt.Errorf("list entries: %w", err)
    }

    opts := export.ReviewOptions{
        Scope: scope,
        Now:   time.Now().UTC(),
    }

    var body []byte
    switch format {
    case "markdown":
        body, err = export.ToReviewMarkdown(entries, opts)
    case "json":
        body, err = export.ToReviewJSON(entries, opts)
    }
    if err != nil {
        return fmt.Errorf("render review: %w", err)
    }

    fmt.Fprintln(cmd.OutOrStdout(), string(body))
    return nil
}
```

Notes:
- `rangeCutoff(scope, now)` is the SAME helper SPEC-018 added in
  `internal/cli/summary.go`. Same package (`cli`); direct reuse.
  Both `"week"` and `"month"` are valid scopes; the error path is
  defensive only (the wrapping `fmt.Errorf` makes the unreachable
  case explicit instead of falling through to a nil cutoff).
- The mutex check uses `cmd.Flags().Changed(...)` rather than the
  flag values themselves, because `--week=false --month` should
  also exit user-error (the user specified --week explicitly,
  even if to false; same shape as "specified twice" for the
  validation purpose).

  *Edge case considered:* `brag review --week=false` is a degenerate
  invocation (the user explicitly opted out of week without opting
  into month). With the `Changed()` check, this combined with
  `--month` exits user-error — the right behavior. With
  `--week=false` alone, `weekSet == true` but `monthSet == false`;
  `scope` falls through to default `"week"` via the if-else
  (since `monthSet == false`). That's slightly counterintuitive
  (the user said --week=false but got week behavior anyway), but
  the alternative — error on `--week=false` alone — would
  surprise scripted callers who pass `--week=$WEEK_FLAG` where
  `$WEEK_FLAG` resolves to `false` to mean "default behavior."
  Locked: explicit `--week=false` alone defaults to week (silent),
  same as bare invocation; only `--week --month` together (both
  Changed) exits user-error. Tests #7 (bare default) and #8 (mutex)
  cover the locked behavior; the `--week=false`-only case is not
  separately tested (low value; behavior follows naturally from
  the code).

### `cmd/brag/main.go` update

One line added after the existing eight `AddCommand` calls:

```go
root.AddCommand(cli.NewReviewCmd())
```

The full block becomes:

```go
root.AddCommand(cli.NewAddCmd())
root.AddCommand(cli.NewListCmd())
root.AddCommand(cli.NewShowCmd())
root.AddCommand(cli.NewDeleteCmd())
root.AddCommand(cli.NewEditCmd())
root.AddCommand(cli.NewSearchCmd())
root.AddCommand(cli.NewExportCmd())
root.AddCommand(cli.NewSummaryCmd())
root.AddCommand(cli.NewReviewCmd())
```

### Test file harness

`internal/cli/review_test.go` follows `summary_test.go`'s pattern
exactly — `newReviewTestRoot(t)` builds a fresh root with the
review subcommand attached, isolates `BRAGFILE_DB`, returns
`*cobra.Command` + `outBuf` + `errBuf`. `runReviewCmd(t, dbPath,
args...)` builds the args slice with `--db <dbPath> review` prefix.

Reuse `firstChars` and `lineMatches` helpers from `summary_test.go`
(same package, source-level reuse).

### Doc updates checklist

Run these greps before AND after the doc sweep to confirm the
audit-grep cross-check both-sides discipline:

```
rg "brag review" docs/ README.md AGENTS.md
rg "What pattern do you see"
rg "STAGE-004.*review|review.*STAGE-004" docs/ README.md AGENTS.md
rg "internal/aggregate" .
rg "GroupEntriesByProject\|ProjectEntryGroup" .
```

Expected post-sweep state:
- `rg "brag review" docs/ README.md AGENTS.md` returns hits in
  api-contract.md (new section), tutorial.md (Scope blurb +
  optional weekly-reflection section), README.md (updated Scope
  blurb), AGENTS.md (digest entry + augmented review entry),
  data-model.md (DEC-014 row).
- `rg "What pattern do you see"` returns hits in stage doc
  (unchanged), AGENTS.md (NEW — augmented glossary entry quotes
  the locked wording), and `internal/export/review.go` (NEW —
  the `var reflectionQuestions` declaration).
- `rg "STAGE-004.*review"` returns hits referring to review's
  shipped status, not deferred status (except the stage doc
  itself, which uses the consistent "review is in this stage's
  scope" wording).

If any post-sweep grep returns an unexpected hit, audit it before
declaring the sweep done.

### Optional tutorial subsection (recommended)

Under `docs/tutorial.md` §4 "Read them back", add (after the
SPEC-018 `### Weekly digest: brag summary` subsection):

```markdown
### Weekly reflection: brag review

Run `brag review --week` (or just `brag review`) to see your last
7 days of entries grouped by project, followed by three reflection
questions designed to seed deeper self-review:

1. What pattern do you see in this period?
2. What did you underestimate?
3. What's missing here that should be?

Pipe the JSON form into your favorite LLM for guided reflection:

    brag review --week --format json | claude "use the entries
    and questions above to reflect on my week"

Use `brag review --month` for a 30-day window. Filter flags are
not accepted — the digest is the unfiltered window. (`brag summary`
is the right command if you want filter composition.)
```

Keep the prose short — the tutorial value is the *paste-into-AI*
pattern, which is the same shape as the SPEC-018 summary
subsection.

---

## Build Completion

*Filled in at the end of the **build** cycle, before advancing to verify.*

- **Branch:** `feat/spec-019-brag-review-week-and-month-flags`
- **PR (if applicable):** *opened after build complete*
- **All acceptance criteria met?** yes
- **New decisions emitted:** none — DEC-014 consumed verbatim per
  Locked design decisions §1.
- **Deviations from spec:**
  - The spec's `internal/cli/review.go` Long-description sketch
    referenced filter flags by their `--`-prefixed names ("Filter
    flags (--tag, --project, --type) are NOT accepted on review …
    no --out flag"). That wording would have failed the spec's own
    acceptance criterion *TestReviewCmd_HelpShowsWeekMonthAndFormat*
    ("help text DOES NOT advertise `--tag`, `--project`, `--type`,
    or `--out`") since `cobra` includes the Long block in `--help`
    output. Adjusted the prose to drop the `--` prefixes
    ("Filter flags (tag, project, type) are NOT accepted on
    review … redirect with > if you want a file") so the help
    surface no longer literally advertises the rejected flag
    names. Behavior, semantics, and accepted-flag set unchanged.
- **Follow-up work identified:** none. SPEC-020 (`brag stats`) is
  the natural sibling and was already on the stage backlog before
  this build; no fresh items surfaced.

### Build-phase reflection (3 questions, short answers)

Process-focused: how did the build go? What friction did the spec create?

1. **What was unclear in the spec that slowed you down?**
   — One small contradiction: the `Long`-description sketch in
   *Notes for the Implementer* literally contained `--tag`,
   `--out`, etc. while the matching acceptance criterion +
   failing test asserted `--help` output must NOT contain those
   strings. The sketch's intent was "state explicitly that these
   flags are unsupported"; the test's intent was "the help
   surface advertises only the supported flag names." Both are
   defensible — the test wins because it's the locking
   contract — and a one-line note in the sketch ("strip `--`
   prefixes from the Long description so the rejection-prose
   doesn't trip the help-surface assertion") would have removed
   the friction entirely.

2. **Was there a constraint or decision that should have been listed but wasn't?**
   — No. Constraints, decisions, locked rejected-alternatives,
   and the audit-grep enumeration were exhaustive. The design-
   side audit-grep cross-check landed every doc-sweep target
   correctly; build re-running the same greps surfaced zero
   deltas, exactly as the addendum prescribes.

3. **If you did this task again, what would you do differently?**
   — Run `cobra`'s `--help` rendering against the planned
   `Long` description during design (not just spec-prose review),
   to catch the kind of help-surface-contradicts-prose tension
   above before the failing-tests-first run hits it. Otherwise
   the build flowed exactly as the SPEC-018 precedent suggested —
   goldens-first, single sketch-per-file, no scope creep, no DEC
   emission, all four locked rejected alternatives held.

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
