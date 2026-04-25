---
# Maps to ContextCore epic-level conventions.
# A Stage is a coherent chunk of work within a Project.
# It has a spec backlog and ships as a unit when the backlog is done.

stage:
  id: STAGE-004                     # stable, zero-padded within the project
  status: proposed                  # proposed | active | shipped | cancelled | on_hold
  priority: medium                  # critical | high | medium | low
  target_complete: null             # optional: YYYY-MM-DD

project:
  id: PROJ-001                      # parent project
repo:
  id: bragfile

created_at: 2026-04-25
shipped_at: null
---

# STAGE-004: Rule-based polish (summary + review + stats)

## What This Stage Is

Add three rule-based aggregation commands that turn the accumulating
corpus into AI-pipeable reflection material without putting any LLM
inside the binary. `brag summary --range week|month` produces a
lightweight markdown digest (counts by type/project + grouped highlights,
no per-entry rendering) — distinct from `brag export --format markdown`
which is the full-document review-prep output. `brag review --week | --month`
prints recent entries grouped by project followed by three static
reflection questions, designed to be pasted into an external AI session
for guided self-review. `brag stats` prints six aggregations (entries/
week, total count, current + longest streak, most-common tags/projects,
corpus span) over the entire corpus. All three accept `--format
markdown|json` (markdown default), so the user's downstream AI workflow
is the same shape across the three commands. No LLM ships in the
binary; that's PROJ-002's reason for existing.

## Why Now

STAGE-003 shipped 2026-04-24 with the review-prep workflow complete:
capture → filter → paste-into-review-doc via `brag export --format
markdown`. The input side and the durable-document side are solved.
What's missing is the lighter-weight aggregation surface — "what
happened this week?", "let me reflect on the last seven days", "show me
the chart-of-myself" — that doesn't need a full export and is
specifically shaped for piping into Claude/GPT for a deeper reflection
pass. Cherry-picked 2026-04-24 from a 9-item provisional STAGE-004
through the user filter "will I actually use this?"; six items dropped
to backlog (emoji passes 1–4, `brag remind`, Claude session-end hook
moved to STAGE-005). The three survivors are the specs the user named
as personally load-bearing for sustained use.

No external blockers. All three specs layer cleanly on STAGE-003's
`Store.List(ListFilter{})`, the `internal/export` rendering
infrastructure, and DEC-011's JSON shape for per-entry payloads.

## Success Criteria

- **`brag summary --range week|month [filters]`** emits a markdown
  document with provenance block (echoing range + filters), a
  by-type/by-project counts block, and grouped-by-project highlights
  (entry titles + IDs only — no descriptions). `--format json` emits
  the same aggregation as a single object (not an array — this is one
  digest, not a list). Filter flags from `brag list` (`--tag`,
  `--project`, `--type`) compose with `--range`.

- **`brag review --week`** and **`brag review --month`** print recent
  entries grouped by project (full titles + IDs, descriptions
  optional/elided for compactness) followed by three hard-coded
  reflection questions ("What pattern do you see?", "What did you
  underestimate?", "What's missing that should be here?"). `--week`
  and `--month` are mutually exclusive; bare `brag review` defaults to
  `--week`. `--format markdown|json` flag honored on both.

- **`brag stats [--format markdown|json]`** prints six aggregations
  over the lifetime corpus: total entries, entries/week (rolling
  average), current streak (consecutive days with ≥1 entry up to
  today), longest streak ever, most-common tags (top 5), most-common
  projects (top 5), corpus span (first → last entry date + days
  elapsed). Markdown default is human-readable; JSON is a single
  fixed-shape object.

- **DEC-014 (summary + review + stats output shape)** locks the
  cross-cutting markdown shape (provenance block convention, summary
  block convention, JSON envelope convention for the non-list outputs)
  consumed by all three specs. Born in SPEC-018; SPEC-019 and SPEC-020
  reference it.

- **`internal/aggregate` package** exists with at least
  `ByType([]Entry)`, `ByProject([]Entry)`, `Streak([]Entry, refDate)`,
  and `MostCommon(field string, n int)` operating on
  `[]storage.Entry`. Aggregation is separated from rendering — render
  layer is `internal/export`, aggregate layer is the new package, CLI
  composes the two.

- All STAGE-001/002/003 success criteria still hold. `go test ./...`,
  `gofmt -l .`, `go vet ./...`, `CGO_ENABLED=0 go build ./...` clean.

## Scope

### In scope

- **`brag summary --range week|month [filters] [--format markdown|json]`** —
  new command. Default `--format markdown`. Filter flags reuse
  `ListFilter`. JSON output is a single object (digest envelope), not
  an array.
- **`brag review --week | --month [--format markdown|json]`** — new
  command. `--week` and `--month` mutually exclusive; bare invocation
  defaults to `--week`. Three hard-coded reflection questions appended
  after grouped entries.
- **`brag stats [--format markdown|json]`** — new command, lifetime
  corpus only (no `--range` in MVP). Six aggregations (see Success
  Criteria).
- **DEC-014** — markdown + JSON shape locked across the three specs:
  provenance block convention, summary-block convention, JSON envelope
  for non-list payloads (single-object digest with `generated_at`,
  `range`/`scope`, `filters`, plus per-spec content keys). Born in
  SPEC-018; consumed by SPEC-019 and SPEC-020.
- **`internal/aggregate` package** — new package alongside
  `internal/export`. Pure functions over `[]storage.Entry`: returns
  structured stats; no rendering. Seeded by SPEC-018; extended by
  SPEC-019 (groups for review) and SPEC-020 (streaks, span).
  Aggregation is the data layer; `internal/export` (or a new
  spec-local renderer) is the bytes layer.
- **Doc sweeps**: `api-contract.md` already has a placeholder `brag
  summary --range week|month` section at lines 251–261 — premise-audit
  hot spot for SPEC-018, the placeholder gets replaced with the full
  shape; `tutorial.md` gets a "weekly reflection workflow" section
  showing the AI-paste pattern; `README.md` (whatever shape it's in
  pre-STAGE-005-rewrite) gets command list updates.

### Explicitly out of scope

Deferred to backlog (full entries with revisit triggers in
`backlog.md`):

- **LLM integration anywhere in the binary.** Period. Outputs are
  rule-based; AI happens at the user's layer (manual paste).
  PROJ-002's reason for existing.
- **Emoji decoration passes 1–4** (stderr feedback, type icons in
  show/list, `--pretty` mode, NO_COLOR + TTY detection). Backlog with
  revisit trigger "user picks a palette and shape they actually want."
- **`brag remind`** — habit-enforcement nudge command. Backlog with
  revisit trigger "first week with zero entries."
- **`brag add --at <date>`** — backdating flag. Backlog with revisit
  trigger "first time the user catches themselves wanting to log a
  brag from a previous day." Was on the table for STAGE-004 framing;
  user-deferred 2026-04-24 to keep the stage at 3 specs.
- **`brag review` reflection questions configurable.** Hard-coded for
  MVP per the prompt's recommendation. Revisit if user wants to swap
  one out.
- **`brag review --range <arbitrary>` or `brag stats --range <X>`.**
  Both stay scope-tight: review is week-or-month; stats is lifetime.
  Promote individually if real demand emerges.
- **`brag summary --group-by type|tag`.** Mirrors DEC-013's deferred
  `--group-by` for markdown export. Same revisit-trigger shape.
- **Compact JSON / non-pretty JSON** for the three new commands.
  Inherits DEC-011's pretty-default; same backlog entry covers all
  JSON-emitting commands.
- **Time-zone configuration for streak calculations.** UTC date
  boundaries for MVP (matches storage's `time.Now().UTC()`); revisit
  if a user explicitly notices a streak break across timezone changes.

Deferred to STAGE-005 (distribution + cleanup):

- **Claude Code session-end hook example** — moved out of STAGE-004
  on 2026-04-24 cherry-pick; depends on SPEC-017's `brag add --json`
  (already shipped) but ships as a STAGE-005 distribution asset, not
  STAGE-004 polish.
- **README rewrite** (current README documents the spec-driven
  process, not the user-facing tool).
- **`docs/brag-entry.schema.json`** mirroring DEC-012.
- **goreleaser + GitHub Actions release + homebrew tap + shell
  completions + CHANGELOG.**

Out of PROJ-001 entirely:

- Any LLM-backed feature (PROJ-002 — AI assist).
- Multi-device sync, cloud backup, auth (out of PROJ-001 per brief).

## Spec Backlog

Ordered by recommended build sequence. SPEC-018 lands first because it
seeds `internal/aggregate` and emits DEC-014; SPEC-019 and SPEC-020
both reuse those and could run in parallel after SPEC-018 ships, but
sequencing them is fine too.

- [ ] SPEC-018 (pending, **M**) — **`brag summary --range
      week|month` + DEC-014 (rule-based-output shape) + seeds
      `internal/aggregate`.** Largest spec in the stage. Lights up the
      "what happened this week/month?" digest with provenance +
      counts-by-type + counts-by-project + grouped highlights
      (titles + IDs only, no descriptions). DEC-014 locks the
      cross-cutting markdown shape (provenance/summary block reuse
      convention) and the JSON envelope for the three rule-based
      digests (single-object, `generated_at` + `scope` + `filters` +
      payload keys). Seeds `internal/aggregate` with `ByType`,
      `ByProject`, helpers used by SPEC-019/020. Filter flags reuse
      `ListFilter` per STAGE-003 pattern.

- [ ] SPEC-019 (pending, **S**) — **`brag review --week | --month`.**
      Static reflection workflow. Prints recent entries grouped by
      project (titles + IDs; descriptions elided for compactness in
      markdown, included in JSON), followed by three hard-coded
      reflection questions ("What pattern do you see?", "What did you
      underestimate?", "What's missing that should be here?").
      `--week` / `--month` mutually exclusive; bare `brag review`
      defaults to `--week`. `--format markdown|json` honored. Reuses
      `internal/aggregate.ByProject` and DEC-014's shape conventions.
      JSON shape: single-object envelope per DEC-014, payload includes
      `entries_grouped` + `reflection_questions` array.

- [ ] SPEC-020 (pending, **S**) — **`brag stats`.** Six lifetime
      aggregations: total count, entries/week (rolling average),
      current streak, longest streak, most-common tags (top 5),
      most-common projects (top 5), corpus span. `--format
      markdown|json` honored; markdown default is a small human-
      readable block, JSON is a single-object envelope per DEC-014.
      Extends `internal/aggregate` with `Streak`, `MostCommon`,
      `Span` helpers.

**Count:** 0 shipped / 0 active / 3 pending

**Complexity check:** 1×M + 2×S, total ~3 specs. Within the
"3–4 specs" healthy-stage band. SPEC-018 is the load-bearing one
(seeds aggregate package, emits DEC-014, defines digest shape that
019/020 inherit) — sized M honestly. No L-complexity entries; no
recommended split.

## Design Notes

Cross-cutting design decisions and per-spec direction. AGENTS.md §9
lessons all apply unchanged (buffer split, tie-break, assertion
specificity, locked-decisions-need-tests, premise-audit family).

### Cross-cutting

- **`--format markdown|json` on all three commands, markdown default.**
  Keeps the AI-consumer story symmetric across the stage and matches
  the existing `brag list --format json|tsv` and `brag export --format
  json|markdown` flag shape from STAGE-003. Unknown `--format` values
  exit 1 (user error) via `UserErrorf`, per DEC-007.

- **DEC-014 — shape lock for the three rule-based outputs.** Born in
  SPEC-018; consumed by SPEC-019 and SPEC-020 without re-litigation.
  Locks: (a) markdown convention — `# <Command Title>` heading,
  provenance block (`Generated: <RFC3339>`, scope/range, filters
  echoed), summary block where applicable, then payload; (b) JSON
  envelope for non-list digests — a single object (NOT an array, this
  is one digest), with stable top-level keys `generated_at`,
  `scope`/`range`, `filters`, plus per-spec payload keys. Distinct
  from DEC-011's naked-array shape because these commands emit
  digests, not lists. Confidence target ~0.80–0.85 (similar to DEC-013;
  shape choices have grounded rationale, but per-spec payload-key
  naming has multiple defensible answers).

- **`internal/aggregate` package — new, separate from
  `internal/export`.** Aggregation is a distinct concern from
  rendering: aggregation maps `[]storage.Entry → structured stats`,
  rendering maps `structured stats → bytes`. Putting them together
  would couple the two and make per-spec testing harder. SPEC-018
  seeds the package with `ByType`, `ByProject`, and any group/count
  helpers `summary` needs. SPEC-019 reuses `ByProject` and adds any
  helpers needed for the review-grouping shape. SPEC-020 extends with
  `Streak`, `MostCommon`, `Span`. Rendering for the three commands
  lives in `internal/export` (existing) or per-spec CLI files —
  framer's call at SPEC-018 design time, but the aggregate/render
  seam is locked.

- **Filter flag reuse.** `brag summary` accepts the same filter flags
  as `brag list` (`--tag`, `--project`, `--type`) on top of `--range`.
  `brag review` and `brag stats` do NOT accept arbitrary filters in
  MVP — review is "the last 7/30 days", stats is "lifetime"; adding
  filters would multiply the spec scope without clear ergonomic win.
  `ListFilter` is the shared input struct where used; no new filter
  logic is written in this stage.

- **Output destination.** Markdown + JSON go to stdout per the
  `stdout-is-for-data-stderr-is-for-humans` constraint. No `--out
  <path>` for the three new commands in MVP — these are pipe-friendly
  digests that users redirect with `>` if they want a file. (Distinct
  from `brag export --out report.md` which is the durable-document
  case.) Backlog if a user asks.

- **Premise audit (AGENTS.md §9 three-case family).** Per-spec hot
  spots:
  - **SPEC-018 (status-change + addition).** `api-contract.md` lines
    251–261 already have a placeholder `brag summary --range
    week|month` section from the brief's earliest sketch. SPEC-018
    REPLACES that placeholder with the full shape — flag this in
    `## Outputs` as a planned doc update, not a build-time discovery.
    Also: `tutorial.md` may reference the planned `summary` from
    earlier sketches; grep `grep -rn "brag summary" docs/ README.md`
    and audit each hit. Help-command tests that count subcommands
    need a +1.
  - **SPEC-019 (addition).** New command surface; help-command
    subcommand counts +1; `tutorial.md` gains a "weekly reflection"
    section.
  - **SPEC-020 (addition).** Same as SPEC-019 — new command, +1
    subcommand count, tutorial section.

- **CLI test harness.** Per AGENTS.md §9: separate `outBuf` / `errBuf`
  per command test; assert no cross-leakage; use line-based equality
  for any markdown heading-level assertion (substring trap from
  SPEC-015's §9 addendum); for any test that asserts on streak or
  date-window logic, fix the "today" reference via a test-injectable
  clock (real `time.Now()` is non-deterministic across midnight).

- **Locked-decisions-need-tests.** Every numbered choice in DEC-014
  must have at least one paired failing test in SPEC-018's `## Failing
  Tests`. SPEC-019 and SPEC-020 inherit DEC-014; their failing tests
  exercise the relevant subset. AGENTS.md §9 lesson from SPEC-009 ship
  reflection.

- **`internal/aggregate` test ergonomics.** Aggregation tests should
  use plain literal `storage.Entry` slices as input, NOT round-trip
  through `Store` — keeps tests fast and decouples aggregation
  correctness from storage correctness. Streak / span tests use a
  test-injectable "today" clock (e.g., a `func() time.Time` field on
  the function or a `refDate time.Time` parameter) so date-boundary
  cases are deterministic.

### SPEC-018-specific (`brag summary`)

- **Range semantics.** `--range week` = entries from the last 7 days
  inclusive of today (UTC). `--range month` = last 30 days. Not
  calendar-week or calendar-month — rolling window. Document
  explicitly in DEC-014 and `api-contract.md`.

- **Highlights vs. full entries.** Summary's distinguishing feature
  vs. `brag export --format markdown` is that it does NOT render full
  entries (no descriptions, no metadata tables). Just `<id>: <title>`
  per entry, grouped under `## <project>` like DEC-013. If the user
  wants full entries they run `brag export --format markdown --since
  7d`. Keeping the surface lean is the point — this is the "skim
  before pasting to AI" view.

- **JSON shape (one of DEC-014's locked choices).** Single object,
  not an array. Top-level keys: `generated_at`, `range` (`"week"` or
  `"month"`), `filters` (echoed flags as object, `(none)` represented
  as empty object `{}` or null — DEC-014 picks), `counts` (object with
  `by_type` and `by_project` sub-objects, key=field-value, value=count),
  `entries_grouped` (object with project name as key, array of
  `{id, title}` objects as value; `(no project)` key for unset).

- **Bare-entries case.** Empty result set: provenance block + an
  explicit `Entries: 0` line, summary and groups omitted. Mirrors
  DEC-013's empty-set treatment.

### SPEC-019-specific (`brag review`)

- **`--week` and `--month` as named flags, mutually exclusive.** Bare
  `brag review` defaults to `--week` (matches the brief's primary
  cadence). `brag review --week --month` exits 1 (user error,
  `UserErrorf`). Asymmetric with `brag summary --range week|month` —
  justified because review's cadence is invoked as a verb (named
  reflection ritual), not parameterized aggregation.

- **The three reflection questions are HARD-CODED for MVP.**
  Verbatim:
  1. What pattern do you see in this period?
  2. What did you underestimate?
  3. What's missing here that should be?

  Configurability (e.g. `~/.bragfile/review-questions.txt`) is
  backlogged with revisit trigger "user wants to swap one out." Test:
  a failing test asserts these three exact strings appear in
  markdown output, and a paired test asserts them as an array in JSON
  output.

- **Description elision.** Markdown output shows `<id>: <title>` per
  entry under the project group (one line per entry); descriptions
  are elided for compactness — the goal is a fast-scan view that gets
  pasted into AI for deeper questions. JSON output INCLUDES
  descriptions (since AI consumers may want them). DEC-014 documents
  this asymmetry explicitly.

- **JSON shape.** Single object per DEC-014: `generated_at`, `scope`
  (`"week"` or `"month"`), `entries_grouped` (project → array of
  full entry objects per DEC-011 shape), `reflection_questions`
  (array of three strings).

### SPEC-020-specific (`brag stats`)

- **Six metrics, lifetime corpus.** Total count, entries/week (total
  entries ÷ corpus-span-in-weeks, or the explicit "0 if span < 1
  week"), current streak (consecutive UTC days with ≥1 entry counting
  back from today; breaks on first day with zero entries), longest
  streak (over all time), top-5 most-common tags (split comma-joined
  per DEC-004, count occurrences across all entries, descending count
  with alpha-ASC tiebreak), top-5 most-common projects (same shape;
  `(no project)` excluded from the top-5 — it's not a project), corpus
  span (`first_entry_date`, `last_entry_date`, `days`).

- **Empty corpus.** Total count 0; everything else either 0, empty
  array, or null. DEC-014 clarifies: numeric metrics are 0; arrays
  are `[]`; date fields are `null` in JSON / `-` in markdown.

- **Streak edge cases.** Today with zero entries → current streak is
  0 (not "the streak that ended yesterday — was N"); the spec
  documents this explicitly. Streak boundaries are UTC days (matches
  storage's `time.Now().UTC()`). Time-zone handling deferred to
  backlog.

- **Time-of-computation determinism.** Test-injectable `now()` for
  current-streak and entries/week tests. Without this, tests flake
  across midnight UTC.

## Dependencies

### Depends on

- **STAGE-003 (shipped 2026-04-24)** — provides `Store.List(ListFilter)`
  with all filter fields, `internal/export` package as the model for
  where rendering helpers live, DEC-011 as the per-entry payload
  shape that JSON outputs reference, DEC-013 as the precedent for
  shape-locked output (markdown export shape).
- **DEC-001 through DEC-013** — all apply forward unchanged.
- **External:** none. stdlib `encoding/json` + `time` cover the
  needs. No new Go module dependencies. Per
  `no-new-top-level-deps-without-decision`, any proposed dep needs
  its own DEC.

### Enables

- **STAGE-005 (distribution + cleanup).** README rewrite can
  showcase the three new commands as the AI-pipe story. Claude
  session-end hook (in STAGE-005) can recommend `brag review --week`
  in its prompt template.
- **PROJ-002 (AI assist), when opened.** DEC-014's JSON envelope is
  the input shape PROJ-002's `brag ai-summary` (or whatever the LLM-
  backed sibling is named) consumes; `internal/aggregate` is the
  shared computation layer the LLM-backed commands wrap rather than
  reimplement.

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
