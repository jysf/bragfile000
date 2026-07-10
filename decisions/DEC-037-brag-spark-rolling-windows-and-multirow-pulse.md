---
# Maps to ContextCore insight.* semantic conventions.

insight:
  id: DEC-037                         # stable, never reused
  type: decision                     # decision | analysis | recommendation | observation
  confidence: 0.72                   # 0.0 - 1.0, honest assessment
  audience:                          # who needs to know?
    - developer
    - agent

agent:
  id: claude-opus-4-8
  session_id: null

# Decisions are repo-level, but it's useful to track which project
# caused them to be emitted.
project:
  id: PROJ-005                       # the project during which this was decided
repo:
  id: bragfile

created_at: 2026-07-10
supersedes: null
superseded_by: null

tags:
  - sparkline
  - cli
  - rolling-window
  - aggregation
  - pulse
  - local-first
---

# DEC-037: `brag spark` — rolling (not calendar) windows, sub-month bucketer, and the multi-row pulse

## Decision

`brag spark` is a sparklines-ONLY "pulse" over a **rolling** recent window (end
= now), NOT a calendar window. Six choices are locked:

1. **Rolling windows, three fixed schemes, default `--month`.** `--week |
   --month | --quarter` are mutually exclusive; none set → `--month`.
   Each maps to a fixed (bucket-width, bucket-count) axis ending at `now`:
   - `--week`  → last **7 days**, **7 daily** buckets (width 24h, n 7).
   - `--month` → last **28 days**, **4 weekly** buckets (width 7d, n 4).
   - `--quarter` → last **91 days**, **13 weekly** buckets (width 7d, n 13).
   The axis start is `now - width*n`; every row buckets over this SAME shared
   axis so glyph position *i* is the same time bucket in every row.

2. **No `--previous` in v1.** A pulse is inherently "now"; a previous-window
   variant is deferred as a clean follow-up.

3. **Rows = a `Total` row + up to the top-8 projects by in-window entry volume**
   (descending, alpha-ASC tiebreak, `(no project)` last — the existing
   `aggregate.ByProject` ordering). `--project <name>` replaces the top-8
   auto-selection with exactly that one project's row (still alongside `Total`),
   and it is a **row selector, not a corpus filter**: `Total` always spans the
   whole in-window corpus, so `Total` vs the one project reads as "this project
   against everything." An explicitly-named `--project` row renders even at zero
   count (a definitive "you did nothing on X"); the top-8 auto-selection
   includes only projects with a non-zero in-window count.

4. **Per-row min→max scaling (via `spark.Line`), magnitude shown in parens.**
   Each row's sparkline is normalized to ITS OWN min→max shape, and the row's
   total in-window count is printed in parentheses for magnitude — so a
   flat-but-busy project and a spiky-but-sparse one are both legible. The parens
   count is the SUM of that row's bucket counts (visual and number always agree).

5. **A NEW pure sub-month bucketer in `internal/aggregate`.** `Cadence` /
   `CoverageByMonth` are calendar-month-only (hard-coded `"2006-01"` labels);
   the pulse needs daily/weekly buckets on a fixed rolling axis. Signature:
   `RollingBuckets(entries []storage.Entry, end time.Time, width time.Duration,
   n int) []int` — returns exactly `n` zero-filled counts; bucket *k* covers
   `[start+k*width, start+(k+1)*width)` (lower-inclusive, upper-exclusive) where
   `start = end.Add(-width*n)`; entries before `start` or at/after `end` are
   excluded. Pure, stdlib-only, instant-arithmetic (location-independent).

6. **DEC-014 envelope; markdown default, JSON = raw counts (DEC-031 choice f).**
   `scope` echoes the window (`week`/`month`/`quarter`). Markdown renders the
   glyph rows by default; `--no-spark` or a present `NO_COLOR` (shared
   `lookupSparkEnv` var) FALLS BACK to per-bucket integer counts (a
   sparkline-only command that suppressed its glyphs still emits its data).
   JSON never contains glyphs — each row carries a raw `series` int array.
   Empty window → markdown header block only (DEC-014 part 4); JSON always emits
   the full zero-filled envelope.

## Context

STAGE-016 (v0.4.x polish) carried `brag spark` as a sketched-but-deferred
"sparklines-only read." The batch brief hypothesized CALENDAR windows with
`--previous`, but explicitly noted "design may refine." Framing refined it to
keep v1 tight. Three genuine forks the stage flagged had to be resolved at
design, not build (STAGE-016 Design Notes): (1) the calendar core (`window.go`)
has month/quarter/year/since but no `week`; (2) `aggregate.Cadence` /
`CoverageByMonth` are monthly-only; (3) "Total + by-project rows of sparklines"
is a new multi-row render shape.

Hard constraints inherited: local-first, zero new dependency, no network/CGO
(DEC-001); SQL only in `internal/storage` (`no-sql-in-cli-layer`); UTC RFC3339
timestamps (`timestamps-in-utc-rfc3339`); the DEC-014 envelope and DEC-031
sparkline-JSON rule.

## Alternatives Considered

- **Option A (chosen): rolling windows + a new rolling bucketer.**
  - What it is: choices 1–6 above.
  - Why selected: a "pulse" means *recent activity up to now*, which is
    inherently rolling; it matches the shipped `brag review` rolling precedent
    (`--week` = last 7 days) and, crucially, **avoids ever defining "calendar
    week"** (ISO week? Monday- vs Sunday-start?) — a real design surface with no
    obviously-right answer for a solo tool. Fixed daily/weekly buckets give
    clean, equal-width glyph columns (7 / 4 / 13) that a min→max sparkline reads
    well. Because storage is UTC and the axis is instant-arithmetic, the buckets
    are DST-immune with no calendar-day logic at all.

- **Option B (rejected for v1): calendar windows (reuse the `windowCutoff`
  family) + `--previous`.**
  - What it is: reuse the impact/coverage/wrapped calendar core (DEC-028) so
    `--month` = the calendar month, `--quarter` = the calendar quarter, plus the
    DEC-032 `--previous` modifier.
  - Why rejected: it forces a "calendar week" definition for the finest bucket
    (the whole point of a pulse's daily/weekly resolution), which ISO-vs-locale
    ambiguity makes a genuine sub-decision; the shipped calendar core has no
    weekly bucketer either, so it would need the same new aggregate helper
    anyway; and a partial current calendar month gives ragged, unequal buckets
    (a 3-day-old month draws almost nothing). `--previous` adds a second window
    axis to a command whose entire value is "right now." Deferred, not deleted:
    if a real workflow wants "my last completed month, calendar-aligned," that is
    a follow-up that can reuse `windowCutoff` + `--previous` without disturbing
    this rolling core.

- **Option C (rejected): `--project` filters the corpus (like every other
  command).**
  - What it is: `--project alpha` narrows the whole query to alpha, so `Total`
    == the alpha row.
  - Why rejected: it produces two identical rows (Total and the one project) —
    no information. spark's shape is *inherently multi-row-by-project*, so the
    useful semantic is a row selector: keep `Total` over the whole corpus and
    render the one project beside it. Documented as a deliberate divergence from
    the corpus-filtering `--project` on coverage/wrapped/impact.

- **Option D (rejected): suppress → show nothing, or keep glyphs in JSON.**
  - Why rejected: a sparkline-only command whose glyphs are suppressed must
    still emit *something*; falling back to per-bucket integer counts in markdown
    is the honest "the sparkline is a visual over these numbers." Putting glyphs
    in JSON is barred by DEC-031 choice f (a sparkline is a lossy visual, not
    data); each JSON row carries the raw `series` instead.

## Consequences

- **Positive:** a fast, dependency-free "what's my recent shape" read that
  reuses `internal/spark` + the DEC-014 envelope verbatim; one small, pure,
  exhaustively-testable new aggregate primitive (`RollingBuckets`) that a future
  daily/weekly cadence anywhere can reuse; no calendar-week definition to
  litigate; DST-immune by construction (UTC instants).

- **Negative:** `--month` now means THREE different spans across the CLI —
  rolling 28 days here, rolling 30 days in `brag review`, and the calendar month
  in impact/coverage/wrapped. This is a real ergonomic wart (see the logged
  question). Mitigated by: the `scope` echo + the `Generated:` line make the
  window unambiguous in output, and 28 days is required to get four clean weekly
  buckets. `--project` also behaves differently here (selector, not filter) than
  elsewhere — documented, and arguably the only sensible reading for a multi-row
  command.

- **Neutral:** a flat non-zero row renders all `▁` (min→max has no magnitude
  reference, DEC-031's accepted tradeoff); the parens count disambiguates it
  here even better than wrapped's adjacent list does. The top-8 cap is a
  round, adjustable number, not a deep choice.

## Validation

- Right if: the pulse reads useful at a glance for the solo corpus; the
  `RollingBuckets` primitive is reused by a later cadence spec without change;
  no user is misled by the rolling-`--month` span once the `scope`/`Generated`
  lines are read.
- Revisit if: the cross-command `--month` drift bites in practice (→ rename
  spark's flags, e.g. `--7d/--28d/--91d`, or unify semantics); a real workflow
  needs calendar-aligned or `--previous` pulses (→ the deferred Option B
  follow-up); a row needs a longer axis than 13 glyphs (→ resampling, DEC-031
  Option C2).
- Confidence 0.72: the primitive, the rolling axis math, and the envelope are
  locked and byte-verified at design (a hand-computed sparkline golden run
  through the real `spark.Line`, and the bucket-boundary assignment run through
  a scratch of `RollingBuckets`). The residual uncertainty is the cross-command
  `--month` semantic drift (choice 1) and the `--project`-as-selector divergence
  (choice 3) — both reversible, both ergonomics bets rather than correctness
  ones. Per §14 (≥0.7, so not required) a question is nonetheless filed for the
  `--month` drift because it is exactly the kind of cross-surface inconsistency
  weekly-review should watch.

## References

- Related specs: SPEC-059 (this decision's spec); SPEC-052/DEC-031 (`spark.Line`
  primitive + sparkline JSON/markdown rule); SPEC-045 (`brag coverage`, the
  closest structural template — DEC-014 consumer + `lookupSparkEnv` wiring);
  SPEC-019 (`brag review`, the rolling `--week`/`--month` precedent); SPEC-051
  (`aggregate.Cadence`, the calendar-month bucketer `RollingBuckets` complements);
  SPEC-056/DEC-035 (`ListFilter.Until`, available but left zero for a pure
  "last N days ending now").
- Related decisions: DEC-014 (rule-based envelope), DEC-031 (sparkline
  normalization + markdown-only + JSON-raw), DEC-028 (calendar-window family,
  the rejected Option B), DEC-032 (`--previous`, deferred here), DEC-001
  (local-first / no-CGO), DEC-023 LD6 (`lookupSparkEnv` injectable-var precedent).
- Constraints: `no-sql-in-cli-layer`, `stdout-is-for-data-stderr-is-for-humans`,
  `timestamps-in-utc-rfc3339`, `test-before-implementation`, `errors-wrap-with-context`.
- External: no-color.org (`NO_COLOR`); holman/spark (min→max block sparkline).
