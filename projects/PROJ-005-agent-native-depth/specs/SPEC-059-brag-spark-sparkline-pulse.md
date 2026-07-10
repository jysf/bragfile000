---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-059
  type: story                      # epic | story | task | bug | chore
  cycle: verify
  blocked: false
  priority: medium
  complexity: M                    # S | M | L  (L means split it)
  # M, not S: two genuinely new surfaces — a new pure aggregate primitive
  # (RollingBuckets, the sub-month bucketer) AND a new multi-row render shape
  # (Total + by-project sparkline rows over a shared axis). Neither is a copy
  # of an existing renderer; both need paired tests.

project:
  id: PROJ-005
  stage: STAGE-016
repo:
  id: bragfile

agents:
  architect: claude-opus-4-8
  implementer: claude-opus-4-8     # usually same Claude, different session
  created_at: 2026-07-10

references:
  decisions: [DEC-037, DEC-014, DEC-031]
  constraints:
    - no-sql-in-cli-layer
    - stdout-is-for-data-stderr-is-for-humans
    - timestamps-in-utc-rfc3339
    - test-before-implementation
    - errors-wrap-with-context
  related_specs: [SPEC-045, SPEC-052, SPEC-019, SPEC-051, SPEC-056]
---

# SPEC-059: `brag spark` — sparkline pulse

## Context

STAGE-016 (v0.4.x polish) carried `brag spark` as a sketched-but-deferred
"sparklines-only read" — a quick visual of recent activity, distinct from the
full digests (`summary`/`review`/`wrapped`/`coverage`). It sits in the STAGE-016
backlog directly below the shipped `ListFilter.Until` promotion (SPEC-056).

The stage flagged three genuine design forks to resolve HERE, not at build
(STAGE-016 Design Notes): (1) the calendar-window core (`internal/cli/window.go`)
has month/quarter/year/since but **no `week`**; (2) `aggregate.Cadence` /
`CoverageByMonth` are **calendar-month-only** (hard-coded `"2006-01"` labels),
so a sub-month bucketer is new; (3) "Total + by-project rows of sparklines" is a
**new multi-row render shape**. DEC-037 resolves all three: rolling (not
calendar) windows, a new pure `aggregate.RollingBuckets`, and a Total + top-8
by-project row layout over one shared axis.

This is the sixth-and-simplest member of the `internal/spark` reuse family
(after `wrapped`'s cadence line and `coverage`'s trend line) and the seventh
DEC-014 envelope consumer.

## Goal

Add `brag spark [--week|--month|--quarter] [--project <name>] [--format <fmt>]
[--no-spark]`: a sparklines-only "pulse" over a ROLLING recent window — a
`Total` row plus up to the top-8 by-project rows, each a per-row min→max
sparkline over the window with its total count in parens — in the DEC-014
envelope (markdown default; JSON = raw per-bucket counts, no glyphs).

## Inputs

- **Files to read:**
  - `decisions/DEC-037-brag-spark-rolling-windows-and-multirow-pulse.md` — the
    six locked choices (rolling windows, no `--previous`, top-8 + `--project`
    selector, per-row scaling, the new bucketer, envelope/JSON rules).
  - `decisions/DEC-014-rule-based-output-shape.md` — the envelope shape.
  - `decisions/DEC-031-sparkline-primitive-normalization-and-placement.md` —
    `spark.Line` behavior + markdown-only/JSON-raw + `--no-spark`/`NO_COLOR`.
  - `internal/spark/spark.go` — `spark.Line([]int) string` (reuse verbatim).
  - `internal/aggregate/aggregate.go` — `ByProject`, `NoProjectKey`, `Cadence`
    (the calendar-month bucketer `RollingBuckets` complements — model its style).
  - `internal/cli/coverage.go` + `internal/export/coverage.go` — the CLOSEST
    structural template (DEC-014 consumer, spark-escape wiring, markdown+JSON).
  - `internal/cli/review.go` — the rolling `--week`/`--month` precedent.
  - `internal/cli/wrapped.go` — the shared `lookupSparkEnv` var (do NOT redeclare).
- **Related code paths:** `internal/cli/`, `internal/export/`,
  `internal/aggregate/`, `cmd/brag/main.go`.
- **External APIs:** none. Local-first, zero new dependency.

## Outputs

- **Files created:**
  - `internal/cli/spark.go` — `NewSparkCmd() *cobra.Command` + `runSpark`.
  - `internal/export/spark.go` — `ToSparkMarkdown` / `ToSparkJSON` + `SparkOptions`.
  - (tests, this design) `internal/aggregate/rolling_test.go`,
    `internal/cli/spark_test.go`.
- **Files modified:**
  - `internal/aggregate/aggregate.go` — add `RollingBuckets` (the new bucketer).
  - `cmd/brag/main.go` — `root.AddCommand(cli.NewSparkCmd())`.
  - `docs/api-contract.md` — add a `brag spark` section (build-time; not a
    failing test — see Premise audit).
- **New exports:**
  - `aggregate.RollingBuckets(entries []storage.Entry, end time.Time, width
    time.Duration, n int) []int`
  - `cli.NewSparkCmd() *cobra.Command`
  - `export.ToSparkMarkdown(entries []storage.Entry, opts SparkOptions)
    ([]byte, error)` and `export.ToSparkJSON(...)`, plus `export.SparkOptions`.
- **Database changes:** none.

## Acceptance Criteria

- [ ] `brag spark` with no window flag defaults to `--month`; `scope` echoes
      `week`/`month`/`quarter`.
- [ ] `--week`/`--month`/`--quarter` are mutually exclusive → two set is a
      `UserError` (exit 1), stdout empty.
- [ ] Rolling axes per DEC-037: `--week` = 7 daily buckets over the last 7 days;
      `--month` = 4 weekly buckets over the last 28 days; `--quarter` = 13 weekly
      buckets over the last 91 days. All rows share the axis.
- [ ] `RollingBuckets` returns exactly `n` zero-filled counts; bucket `k` covers
      `[start+k*width, start+(k+1)*width)` (lower-inclusive, upper-exclusive);
      `start = end - width*n`; entries before `start` or at/after `end` excluded.
- [ ] Markdown: a `Total` row + up to top-8 by-project rows (desc count, alpha
      tiebreak, `(no project)` last); each row `<label> (<count>): <glyphs>`.
- [ ] `--project <name>` renders `Total` + that one project's row only (row
      selector, NOT a corpus filter); other project rows absent; a named project
      with zero in-window entries still renders (all-`▁` / zero counts).
- [ ] Markdown default-on renders glyphs; `--no-spark` OR a present `NO_COLOR`
      falls back to raw per-bucket integer counts (no glyph runes).
- [ ] `--format json` is valid JSON with the DEC-014 provenance keys + `total`
      (`{count, series}`) + `by_project` (`[{project, count, series}]`) + a
      `window` block; contains NO glyph runes (DEC-031 choice f).
- [ ] `--format` accepts only `markdown`/`json`; anything else is a `UserError`.
- [ ] Empty window → markdown header block only (through `Entries: 0`), no
      `## Pulse`; JSON emits the full zero-filled envelope.
- [ ] Success writes only to stdout; UserErrors keep stdout clean.

## Failing Tests

Written during **design**, BEFORE build. New Go tests will not COMPILE until the
new symbols (`aggregate.RollingBuckets`, `cli.NewSparkCmd`) exist — that is the
expected fail-first state (AGENTS.md §12 build step confirms the failure reason).

- **`internal/aggregate/rolling_test.go`** (hermetic, table-driven, no clock/I/O)
  - `TestRollingBuckets_DailyAndWeeklyAxes` — the three schemes (7 daily / 4
    weekly / 13 weekly), zero-fill, exactly `n` buckets. Pairs DEC-037 choice 1 + 5.
  - `TestRollingBuckets_BucketBoundaryAssignment` — lower-inclusive /
    upper-exclusive edges; entry `== end` excluded; entry `< start` excluded.
    Pairs DEC-037 choice 5's boundary rule.
  - `TestRollingBuckets_FeedsSparkLine` — **§12(b) hand-computed golden**: a
    known corpus buckets to `[1,0,2,3,1,0,2]`, and `spark.Line` renders it as
    `"▃▁▆█▃▁▆"` (computed by hand, confirmed against the real `spark.Line` at
    design). Ties the new bucketer to the primitive end-to-end.
- **`internal/cli/spark_test.go`** (separate `outBuf`/`errBuf`, no cross-leakage)
  - `TestSparkCmd_DefaultWindowIsMonth` — bare `spark` → `Scope: month` + header.
  - `TestSparkCmd_WindowsMutuallyExclusive` — `--week --quarter` → `UserError`,
    empty stdout.
  - `TestSparkCmd_TotalAndByProjectRows` — `Total (`, `alpha (`, `beta (`,
    `(no project) (` all present.
  - `TestSparkCmd_TopEightCap` — 8 tied-at-2 projects kept, two count-1 projects
    dropped.
  - `TestSparkCmd_ProjectSelector` — `--project alpha` → `Total` + `alpha` row,
    `beta` row absent.
  - `TestSparkCmd_SparkDefaultOnAndSuppressed` — default has glyphs; `--no-spark`
    and `NO_COLOR` both remove all glyph runes.
  - `TestSparkCmd_JSONRawCountsNoGlyphs` — valid JSON, has `total`/`by_project`/
    `scope`, NO glyph runes.
  - `TestSparkCmd_UnknownFormat` — `--format yaml` → `UserError`, empty stdout.
  - `TestSparkCmd_EmptyCorpusHeaderOnly` — `Entries: 0`, no `## Pulse`, no glyphs.
  - `TestSparkCmd_StdoutStderrSeparation` — success stdout-only; UserError clean.
  - `TestSparkCmd_HelpShowsExamples` — help has `Examples:` and `brag spark`.

## Implementation Context

*Read this section (and the files it points to) before starting the build cycle.*

### Decisions that apply

- `DEC-037` — the six locked choices for `brag spark` (this spec's decision).
- `DEC-014` — single-object JSON envelope; markdown provenance block; empty-state
  rule (part 4: body omitted, header remains).
- `DEC-031` — `spark.Line` normalization + markdown-only + JSON-raw + the
  `--no-spark`/`NO_COLOR` escape via the shared `lookupSparkEnv` var.

### Constraints that apply

- `no-sql-in-cli-layer` — the bucketer is pure Go in `internal/aggregate`; the
  CLI queries only via `storage.ListFilter`. No SQL in `internal/cli`.
- `stdout-is-for-data-stderr-is-for-humans` — the pulse body to stdout;
  `UserError` to stderr with clean stdout.
- `timestamps-in-utc-rfc3339` — window math is instant-arithmetic on UTC
  timestamps (location-independent for `Before`/`Sub`).
- `test-before-implementation`, `errors-wrap-with-context`.

### Prior related work

- `SPEC-052`/`DEC-031` (shipped) — `internal/spark` primitive; the `lookupSparkEnv`
  escape pattern.
- `SPEC-045`/`DEC-033` (shipped) — `brag coverage`; the structural template
  (DEC-014 consumer + spark line + markdown/JSON split).
- `SPEC-019`/`DEC-014` (shipped) — `brag review`; the rolling `--week`/`--month`
  precedent (`rangeCutoff` uses `now - Nd`).
- `SPEC-051` (shipped) — `aggregate.Cadence`, the calendar-month bucketer.
- `SPEC-056`/`DEC-035` (shipped) — `ListFilter.Until` (available; left zero here).

### Out of scope (for this spec specifically)

- **`--previous`** (DEC-037 choice 2) — a pulse is "now"; deferred.
- **Calendar-aligned windows** — the rejected alternative (DEC-037 Option B).
- **Visual column padding** — glyphs align at the DATA level (shared axis →
  glyph *i* is the same bucket in every row); left-padding labels to a common
  width for pixel alignment is deferred (keeps goldens simple).
- **`stats` cadence sparkline** — needs a new lifetime-cadence slot (STAGE-016
  out-of-scope; DEC-031 defers it).
- **`--out <path>`** — stdout only, like every other digest.

## Notes for the Implementer

### The bucketer (`internal/aggregate/aggregate.go`)

Add, modeled on `Cadence`'s shape but sub-month and axis-driven:

```go
// RollingBuckets buckets entries onto a fixed rolling axis of n buckets each
// `width` wide whose LAST bucket ends (exclusive) at end. The axis start is
// end.Add(-width*n); bucket k (0-indexed) covers
// [start+k*width, start+(k+1)*width) — lower-inclusive, upper-exclusive.
// Returns exactly n zero-filled counts (spark/JSON-ready); entries before start
// or at/after end are excluded, so sum(result) == the in-axis subset size. Pure,
// stdlib-only, instant-arithmetic (location-independent) — the sub-month analog
// of Cadence, which is calendar-month-only. SPEC-059/DEC-037.
func RollingBuckets(entries []storage.Entry, end time.Time, width time.Duration, n int) []int {
	start := end.Add(-width * time.Duration(n))
	out := make([]int, n)
	for _, e := range entries {
		t := e.CreatedAt
		if t.Before(start) || !t.Before(end) {
			continue
		}
		k := int(t.Sub(start) / width)
		if k >= 0 && k < n {
			out[k]++
		}
	}
	return out
}
```

Window flag → `(width, n)` (all confirmed against a scratch of the above at
design; start offsets 7d / 28d / 91d):

| flag        | width | n  | span    | buckets |
|-------------|-------|----|---------|---------|
| `--week`    | 24h   | 7  | 7 days  | daily   |
| `--month`   | 7*24h | 4  | 28 days | weekly  |
| `--quarter` | 7*24h | 13 | 91 days | weekly  |

### The CLI (`internal/cli/spark.go`) — copy `coverage.go`'s spine

Literal cobra command (transcribe verbatim; NOT-contains self-audit run at
design — the `Long` contains zero block-glyph runes, so the JSON/`--no-spark`/
empty "no glyphs" assertions hold):

```go
cmd := &cobra.Command{
	Use:   "spark",
	Short: "Sparklines-only pulse of recent activity (Total + by-project) over a rolling window",
	Long: `Print a sparklines-only pulse of your recent activity: a Total row plus a row per project, each a compact sparkline over a rolling recent window with the entry count in parentheses. Rule-based, deterministic, no LLM. A quick "what does my last stretch look like?" glance, not a full digest.

The window is ROLLING (it ends now), not a calendar period. Exactly one window may be given; they are mutually exclusive and default to --month:
  --week      the last 7 days, one bar per day (7 bars)
  --month     the last 28 days, one bar per week (4 bars, default)
  --quarter   the last 91 days, one bar per week (13 bars)

Every row is bucketed over the same axis, so bar position lines up across rows. Each row's sparkline is scaled to its own min-max shape; the count in parentheses carries the magnitude, so a steady project and a bursty one are both legible. By default the Total row plus the top 8 projects by entry count are shown; --project <name> narrows the by-project rows to that one project (the Total still spans everything).

Output is markdown (default) or a single-object JSON envelope (--format json) per DEC-014. The sparkline is markdown-only; JSON carries raw per-bucket counts (--format json | jq .total.series). --no-spark, or a present NO_COLOR env var, drops the glyphs and prints raw per-bucket counts instead.

Examples:
  brag spark                                  # last 28 days, weekly bars, markdown
  brag spark --week                           # last 7 days, one bar per day
  brag spark --quarter --project alpha        # 13 weekly bars: Total vs alpha
  brag spark --month --format json            # raw per-bucket counts, JSON envelope`,
	RunE: runSpark,
}
cmd.Flags().Bool("week", false, "pulse over the last 7 days (7 daily bars)")
cmd.Flags().Bool("month", false, "pulse over the last 28 days (4 weekly bars; the default)")
cmd.Flags().Bool("quarter", false, "pulse over the last 91 days (13 weekly bars)")
cmd.Flags().String("project", "", "show only this project's row alongside Total (row selector, not a corpus filter)")
cmd.Flags().String("format", "markdown", "output format (one of: markdown, json)")
cmd.Flags().Bool("no-spark", false, "suppress glyphs; print raw per-bucket counts instead")
```

**Flag defaults (explicit, per §12 flag-default-explicitness):** `week`/`month`/
`quarter` default `false` (bare invocation → `--month`); `project` default `""`;
`format` default `"markdown"`; `no-spark` default `false`.

`runSpark` flow (mirror `runCoverage`):
1. `now := nowFunc()`.
2. Resolve the window: exactly-one-or-none among week/month/quarter → default
   `"month"`; two-plus → `UserErrorf("--week, --month, --quarter are mutually
   exclusive")`. (A small local helper like `review.go`'s check; `selectedWindow`
   in `window.go` is for the CALENDAR flag set — do not reuse it.)
3. Map window → `(width, n)` per the table; compute `end := now`.
4. `format` guard (markdown/json only) → `UserErrorf` on anything else.
5. `--project` (if set and empty string → `UserErrorf("--project must not be
   empty")`) is the ROW SELECTOR passed into `SparkOptions`, NOT a
   `ListFilter.Project`. Do not put it on the query.
6. Query: `storage.ListFilter{Since: now.Add(-width*n)}` — leave `Until` zero
   (DEC-037: a pure "last N days ending now"; the bucketer's `end` is the
   effective exclusive upper edge). No tag/type filters in v1.
7. `sparkOn := !noSpark && !noColorSet` using the SHARED `lookupSparkEnv`
   (declared in `wrapped.go`; do NOT redeclare).
8. Build `SparkOptions{Scope, Now: now, Width: width, Buckets: n, Project,
   Spark: sparkOn}` and dispatch to `ToSparkMarkdown` / `ToSparkJSON`.
9. `fmt.Fprintln(cmd.OutOrStdout(), string(body))`.

### The renderer (`internal/export/spark.go`) — mirror `coverage.go`

`SparkOptions` carries `Scope string`, `Now time.Time`, `Width time.Duration`,
`Buckets int`, `Project string` (empty = top-8 auto-select), `Spark bool`. The
renderer does the aggregation (as `coverage.go` calls `CoverageByMonth` inside):

Row selection:
- `total := aggregate.RollingBuckets(entries, opts.Now, opts.Width, opts.Buckets)`.
- Group entries by project key (empty → `aggregate.NoProjectKey`) into
  `map[string][]storage.Entry`.
- If `opts.Project != ""`: the single row is that project (its subset may be
  empty → all-zero buckets, still rendered).
- Else: order projects via `aggregate.ByProject(entries)` (desc count, alpha
  tiebreak, `NoProjectKey` last), take the first 8 names, bucket each.
- Each row's parens count = the SUM of that row's buckets (visual/number agree).

**Markdown** (`ToSparkMarkdown`), following DEC-014 + `coverage.go`:

```
# Bragfile Spark

Generated: 2026-07-10T12:00:00Z
Scope: month
Filters: (none)
Entries: 6

## Pulse

Total (6): █▂▁▅
alpha (3): █▁▁▅
beta (2): █▁▁▁
(no project) (1): ▁▁▁█
```

- `Filters:` — v1 has no tag/type filters; when `--project` is set echo
  `--project <name>`, else `(none)`. (Reuse an `echoFilters`-shaped local, or
  inline; `--project` is the only echoable flag.)
- `Entries:` — `len(entries)` (the whole in-window corpus, matching Total's sum).
- When `opts.Spark` is false, replace `<glyphs>` with the space-joined raw counts
  (e.g. `Total (6): 3 1 0 2`). Trailing newline stripped via the existing
  `trimTrailingNewline` helper.
- Empty window (`len(entries) == 0`): emit through `Entries: 0` and return (no
  `## Pulse`) — DEC-014 part 4.

**JSON** (`ToSparkJSON`) — struct with json tags in this key order (DEC-014 flat
provenance first; `series` is raw counts, no glyphs — DEC-031 choice f):

```json
{
  "generated_at": "2026-07-10T12:00:00Z",
  "scope": "month",
  "filters": {},
  "window": { "buckets": 4, "bucket_width_days": 7, "start": "2026-06-12T12:00:00Z", "end": "2026-07-10T12:00:00Z" },
  "total": { "count": 6, "series": [3, 1, 0, 2] },
  "by_project": [
    { "project": "alpha", "count": 3, "series": [2, 1, 0, 0] },
    { "project": "(no project)", "count": 1, "series": [0, 0, 0, 1] }
  ]
}
```

- `bucket_width_days`: `1` for `--week`, `7` for `--month`/`--quarter`.
- `filters` always an object (`{}` when none) — DEC-014.
- Empty window: `total.count` 0, `total.series` full zero-filled, `by_project`
  `[]` (non-nil). JSON is always fully emitted (mirrors `coverage`).
- `json.MarshalIndent(env, "", "  ")` (DEC-014 choice 5).

### Register

Add `root.AddCommand(cli.NewSparkCmd())` in `cmd/brag/main.go` (alongside the
other `New*Cmd()` calls).

### Premise audit (§9) — ran at design

- **New subcommand → any test asserting the exact root subcommand set / help?**
  `grep` for `Available Commands` / `len(...Commands())` / a command enumeration
  in `internal/cli/*_test.go` and `cmd/brag/*_test.go`: **no hits**. `root_test.go`
  asserts only `--version` / `--help` / a `<command>` placeholder — none couples
  to the command count. `scripts/test-docs.sh` A5 checks a FIXED 7-verb list
  (add/list/search/export/summary/review/stats) in README fenced blocks — it does
  NOT enumerate all commands (coverage/wrapped/impact are already absent from it),
  so adding `spark` breaks nothing there. **No planned test deletions/updates.**
- **Doc references (status-change case):** a new command should be documented.
  `docs/api-contract.md` gains a `brag spark` section at build; this is a doc
  addition, not a failing test, and no existing doc has a count-coupled "there
  are N commands" assertion (grep of `docs/`/`README.md` for `brag wrapped`/`brag
  coverage` shows prose sections, not counts). Listed under Outputs.
- **Additive-collection case:** `spark` adds no migration, DEC-count-asserted
  collection, or constraint. `NewSparkCmd` is a new `New*Cmd` in `main.go` — not
  count-asserted anywhere.

### Gotchas

- Do NOT redeclare `lookupSparkEnv` — it lives in `wrapped.go` (package `cli`).
- `--project` here is a ROW SELECTOR, not a corpus filter (DEC-037 choice 3) —
  deliberately different from `coverage`/`wrapped`/`impact`. Do not set it on
  `ListFilter`.
- `--month` here = rolling 28 days — DIFFERENT from `review`'s rolling 30 days
  and from the calendar month in impact/coverage/wrapped. This is intentional
  (DEC-037; a logged question watches it). Do not "fix" it to 30 or to a calendar
  month.
- Parens count and glyphs must agree: use `sum(series)`, not `len(subset)`.

---

## Build Completion

*Filled in at the end of the **build** cycle, before advancing to verify.*

- **Branch:** `feat/spec-059-brag-spark`
- **PR (if applicable):** see PR link in the branch's PR
- **All acceptance criteria met?** yes
- **New decisions emitted:**
  - none (DEC-037 already covered every choice)
- **Deviations from spec:**
  - `RollingBuckets` lives in a new file `internal/aggregate/rolling.go` (the
    test lives in `rolling_test.go`) rather than being appended to
    `aggregate.go` — the Outputs section listed `aggregate.go`, but a
    dedicated file keeps the new primitive with its test and matches the
    `rolling_test.go` sibling. Signature and behavior are byte-identical to
    the spec's transcribed body.
  - No other deviations; the literal cobra `Long`, flag defaults, render
    shape, and JSON key order were transcribed as specified.
- **Follow-up work identified:**
  - none new. The already-deferred `--previous`, calendar-aligned windows,
    and column padding remain the documented follow-ups (DEC-037).

### Build-phase reflection (3 questions, short answers)

Process-focused: how did the build go? What friction did the spec create?

1. **What was unclear in the spec that slowed you down?**
   — Nothing material. The spec was unusually complete: the bucketer body,
   the cobra `Long`, the flag table, the markdown/JSON shapes, and the
   §12(b) golden were all transcribable. The one micro-decision was where
   `RollingBuckets` should live (a new file vs appended); both were
   defensible, so it cost only a moment.

2. **Was there a constraint or decision that should have been listed but wasn't?**
   — No. `no-sql-in-cli-layer`, the stdout/stderr split, UTC RFC3339, and
   the `lookupSparkEnv` reuse were all called out, and each mapped cleanly
   to the authored tests.

3. **If you did this task again, what would you do differently?**
   — Nothing significant. Reading the failing tests before writing the
   renderer paid off (the JSON key names and the `Total (` / `alpha (`
   row-label format are pinned by the tests, so matching them first avoided
   any churn).

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
