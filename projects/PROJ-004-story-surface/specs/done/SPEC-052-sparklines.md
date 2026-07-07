---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-052
  type: feature
  cycle: ship
  blocked: false
  priority: high
  complexity: M

project:
  id: PROJ-004
  stage: STAGE-013
repo:
  id: bragfile

agents:
  architect: claude-opus-4-8
  implementer: claude-opus-4-8     # usually same Claude, different session
  created_at: 2026-07-06

references:
  decisions: [DEC-031, DEC-030, DEC-014, DEC-001, DEC-023, DEC-007]
  constraints:
    - no-cgo
    - no-new-top-level-deps-without-decision
    - stdout-is-for-data-stderr-is-for-humans
    - no-sql-in-cli-layer
    - test-before-implementation
    - errors-wrap-with-context
  related_specs: [SPEC-051, SPEC-020, SPEC-048]
---

# SPEC-052: The sparklines / visual pass — an in-terminal Unicode sparkline over `wrapped` cadence

## Context

STAGE-013's **dedicated visual pass**, deferred on purpose from `brag impact`
(which shipped text-pure) and foreseen throughout SPEC-051, which left
`wrapped`'s `cadence.series` as a zero-filled, sparkline-ready
`[]{period,count}` slot for exactly this. The user asked for it ("sparklines
could be fun"); it should make the shareable digest feel alive.

This spec adds ONE small, pure primitive — an in-terminal Unicode block-glyph
sparkline (`▁▂▃▄▅▆▇█`) — and wires it into `brag wrapped`'s markdown `## Cadence`
section over the counts that already print there. **Local-first, ZERO new
dependency, no network** (DEC-001, `no-cgo`): Unicode block characters computed
in pure Go. An external-plotter pipe is an explicitly LATER layer, NOT this spec
(STAGE-013 design note).

- Parent stage: `STAGE-013` (Polish + v0.4.0 cut). The second spec in its backlog.
- Project: `PROJ-004` (the story surface). The visual polish on the shareable
  `wrapped` digest.
- New decision this spec emits: **DEC-031** (sparkline primitive: normalization +
  placement + default-on-with-escape). Confidence 0.78 (< 0.8 → a question is
  logged in `guidance/questions.yaml`; see Implementation Context).
- **JSON stays raw.** A sparkline is a lossy visual rendering of
  `cadence.series[].count`, not data — no glyphs enter any JSON envelope. The
  SPEC-051 wrapped JSON goldens are untouched by this spec.

## Goal

Ship a pure `internal/spark` package (`spark.Line([]int) string`) that maps a
numeric series to a fixed-width Unicode block-glyph string via min→max
normalization, and render it as a single labeled `Cadence: <glyphs>` line in
`brag wrapped`'s markdown `## Cadence` section — **default-on**, suppressible via
`--no-spark` or a non-empty `NO_COLOR` env var. `stats` and `impact` are NOT
touched by this spec (see Out of scope + DEC-031).

## Inputs

- **Files to read:**
  - `internal/export/wrapped.go` — `WrappedOptions`, `ToWrappedMarkdown` (the
    `## Cadence` block this spec adds a line to), `ToWrappedJSON` (UNCHANGED). The
    renderer already calls `aggregate.Cadence(entries, opts.ScopeMonths)` and gets
    `series []aggregate.CadenceBucket` — the counts the sparkline reads.
  - `internal/aggregate/aggregate.go` — `CadenceBucket{Period, Count}` and
    `Cadence` (the sparkline-ready slot SPEC-051 LD8 put here). The primitive
    reads `series[i].Count`.
  - `internal/cli/wrapped.go` — `NewWrappedCmd` (flag set), `runWrapped` (where
    `--no-spark`/`NO_COLOR` resolve and `opts.Spark` is set).
  - `internal/cli/milestone.go` — the `defaultStderrIsTTY` / package-`var` seam
    precedent (DEC-023 LD6): stdlib-only host-state probing behind an injectable
    package var, no `go-isatty` dep. The `NO_COLOR` env seam mirrors this shape.
  - `DEC-031` (this spec's decision), `DEC-030` (wrapped envelope + cadence
    slot), `DEC-014` (digest envelope), `DEC-001` (pure-Go / local-first).
- **Related code paths:** `internal/spark/` (new), `internal/export/wrapped.go`,
  `internal/cli/wrapped.go`.

## Outputs

- **Files created:**
  - `internal/spark/spark.go` — the `spark.Line([]int) string` primitive + the
    `levels` glyph table.
  - `internal/spark/spark_test.go` — the normalization table + edge-case tests
    below.
- **Files modified:**
  - `internal/export/wrapped.go` — add `Spark bool` to `WrappedOptions`; in
    `ToWrappedMarkdown`, when `opts.Spark` is true (and there are entries), print
    the `Cadence: <spark.Line(counts)>` line inside `## Cadence`. `ToWrappedJSON`
    is UNCHANGED (raw counts only).
  - `internal/export/wrapped_test.go` — the SPEC-051 markdown goldens
    (`TestToWrappedMarkdown_DEC014FullDocumentGolden`,
    `TestToWrappedMarkdown_QuarterGolden`) gain the `Cadence: <glyphs>` line
    (they run with `Spark: true`); a new `TestToWrappedMarkdown_NoSparkOmitsGlyphLine`
    asserts `Spark: false` reproduces the SPEC-051 (pre-sparkline) Cadence block.
    The JSON golden and empty-period tests are UNCHANGED. (See § Premise audit.)
  - `internal/cli/wrapped.go` — add the `--no-spark` bool flag; add the
    `lookupSparkEnv` package-`var` env seam; in `runWrapped`, compute
    `spark := !noSpark && !noColorSet()` and set `opts.Spark = spark`.
  - `internal/cli/wrapped_test.go` — `TestWrappedCmd_SparkDefaultOn`,
    `TestWrappedCmd_NoSparkFlagSuppresses`, `TestWrappedCmd_NoColorEnvSuppresses`.
  - `docs/api-contract.md` — note the `--no-spark` flag + `NO_COLOR` behavior in
    the `brag wrapped` section.
  - `docs/tutorial.md` — mention the cadence sparkline in the `wrapped` subsection.
  - `README.md` — add `--no-spark` / `NO_COLOR` if the `wrapped` flags are
    enumerated (build re-verifies via the audit-grep, see Implementation Context).
  - `AGENTS.md` §11 — add a `sparkline` glossary term.
  - `guidance/questions.yaml` — log the DEC-031 < 0.8 question (§14).
  - The STAGE-013 backlog line for SPEC-052 flips to a build state at build.
- **New exports:**
  - `spark.Line(vals []int) string`
  - `export.WrappedOptions.Spark bool` (new field on the existing struct)
- **Database changes:** none. Pure rendering over existing data.

## Acceptance Criteria

- [ ] `spark.Line` maps a numeric series to a Unicode block-glyph string,
      **one glyph per element** (fixed width, no resampling), using min→max
      normalization `level = round((v-min)/(max-min)*7)` over `▁▂▃▄▅▆▇█`.
- [ ] Edge cases: empty series → `""`; a flat series (`max==min`, INCLUDING
      all-zero and a single element) → all `▁`; a ramp `0..7` → `▁▂▃▄▅▆▇█`.
- [ ] The primitive is pure (no I/O, no clock, no host state), lives in
      `internal/spark`, and adds NO dependency (`go.mod`/`go.sum` unchanged;
      `CGO_ENABLED=0 go build ./...` still clean).
- [ ] `brag wrapped` (markdown, default) renders a `Cadence: <glyphs>` line inside
      `## Cadence`, between the `Busiest month:` line and the per-month list, over
      the same zero-filled `cadence.series` counts.
- [ ] For the SPEC-051 year fixture the glyph line is exactly
      `Cadence: ▅▅▁█▁▁█▁▁▁▅▁`; for the quarter fixture it is exactly
      `Cadence: █▁▁`. (Byte-exact, computed at design time against the real
      primitive — see § Failing Tests.)
- [ ] `--no-spark` OR a non-empty `NO_COLOR` env var suppresses the glyph line;
      the rest of the digest (including the per-month counts) is byte-identical to
      the SPEC-051 output. Either signal alone suffices.
- [ ] The **JSON** envelope is UNCHANGED — no glyphs, raw `cadence.series` counts
      only; the SPEC-051 JSON golden still passes byte-for-byte.
- [ ] The empty-period digest is unaffected: no `## Cadence` body, hence no glyph
      line (the body-omission from SPEC-051/DEC-014 part 4 still holds).
- [ ] `stats` and `impact` output are unchanged by this spec (no sparkline added
      to either — DEC-031).

## Failing Tests

Written during **design**, BEFORE build. Every glyph golden below was computed at
design time against the REAL primitive (a scratch program running the exact
`spark.Line` algorithm — since removed): they are faithful, not hand-typed. The
year series `[1,1,0,2,0,0,2,0,0,0,1,0]` → `▅▅▁█▁▁█▁▁▁▅▁`; the quarter series
`[2,0,0]` → `█▁▁`; `0..7` → `▁▂▃▄▅▆▇█`. (Reproduce: `level =
int(math.Round(float64(v-min)/float64(max-min)*7))`, `max==min → ▁` for every
element, empty → `""`.)

### `internal/spark/spark_test.go`

- **`TestLine_NormalizationTable`** (LOAD-BEARING — the primitive's contract).
  Table-driven, each row `{name, in []int, want string}`, asserting
  `spark.Line(in) == want`:

  | name | in | want |
  |---|---|---|
  | `empty` | `[]int{}` | `""` |
  | `nil` | `nil` | `""` |
  | `single` | `[]int{5}` | `"▁"` |
  | `flat-zero` | `[]int{0,0,0}` | `"▁▁▁"` |
  | `flat-nonzero` | `[]int{3,3,3}` | `"▁▁▁"` |
  | `ramp-0-7` | `[]int{0,1,2,3,4,5,6,7}` | `"▁▂▃▄▅▆▇█"` |
  | `two-point-min-max` | `[]int{0,1}` | `"▁█"` |
  | `wrapped-year` | `[]int{1,1,0,2,0,0,2,0,0,0,1,0}` | `"▅▅▁█▁▁█▁▁▁▅▁"` |
  | `wrapped-quarter` | `[]int{2,0,0}` | `"█▁▁"` |
  | `classic` | `[]int{0,2,5,4,7,8,3,1}` | `"▁▃▅▅▇█▄▂"` |

  Each row is independently verifiable by hand:
  - `ramp-0-7`: `min=0,max=7`, `level=round(v/7*7)=v`, so glyph index == value →
    the full ladder `▁▂▃▄▅▆▇█`.
  - `two-point-min-max`: `min=0,max=1`, `0→▁`, `1→round(7)=7→█`.
  - `wrapped-year`: `min=0,max=2`; `0→▁`, `1→round(3.5)=4→▅`, `2→7→█`.
  - `wrapped-quarter`: `min=0,max=2`; `2→█`, `0→▁`, `0→▁`.
  - `classic`: `min=0,max=8`; `round(v/8*7)` → `0,2,4,4,6,7,3,1` glyph indices →
    `▁▃▅▅▇█▄▂`.
  - `flat-*` / `single`: `max==min` branch → every element `▁`.
  - `empty`/`nil`: `len==0` branch → `""`.

- **`TestLine_LengthMatchesInput`**. For a handful of non-empty inputs, assert
  `len([]rune(spark.Line(in))) == len(in)` (fixed-width, one glyph per element —
  guards against an off-by-one or accidental resampling).

- **`TestLine_OnlyBlockGlyphsOrEmpty`**. For every non-empty input tested, assert
  each rune of the output is one of `▁▂▃▄▅▆▇█` (no stray ASCII, no space) — pins
  the glyph table.

### `internal/export/wrapped_test.go` (UPDATED goldens + one new test)

The two markdown goldens from SPEC-051 now run with `Spark: true` and gain the
glyph line. **Only the `## Cadence` block changes** — the insertion point is a new
line between `Busiest month:` and the blank line preceding the per-month list.

- **`TestToWrappedMarkdown_DEC014FullDocumentGolden`** (UPDATED). Same fixture
  (`wrappedYearFixture`) and opts as SPEC-051 plus `Spark: true`. The `## Cadence`
  section of `want` becomes:

```
## Cadence

Busiest month: 2026-04 (2)
Cadence: ▅▅▁█▁▁█▁▁▁▅▁

- 2026-01: 1
- 2026-02: 1
- 2026-03: 0
- 2026-04: 2
- 2026-05: 0
- 2026-06: 0
- 2026-07: 2
- 2026-08: 0
- 2026-09: 0
- 2026-10: 0
- 2026-11: 1
- 2026-12: 0
```

  (Every other section — Top initiatives, Impact moments, Rhythm, Span — is
  byte-identical to SPEC-051. The trailing `\n` is still stripped.)

- **`TestToWrappedMarkdown_QuarterGolden`** (UPDATED). Same quarter fixture/opts as
  SPEC-051 plus `Spark: true`. The `## Cadence` section of `want` becomes:

```
## Cadence

Busiest month: 2026-07 (2)
Cadence: █▁▁

- 2026-07: 2
- 2026-08: 0
- 2026-09: 0
```

- **`TestToWrappedMarkdown_NoSparkOmitsGlyphLine`** (LOAD-BEARING — the escape
  path). Same `wrappedYearFixture`/opts but `Spark: false`. Assert the rendered
  markdown is byte-identical to the SPEC-051 (pre-sparkline) document — i.e. the
  `## Cadence` block is exactly `Busiest month: 2026-04 (2)\n\n- 2026-01: 1\n...`
  with NO `Cadence:` glyph line, AND assert `!strings.Contains(md, "Cadence: ▅")`
  (or, more precisely, that no line begins with `"Cadence: "`). This proves the
  `Spark` gate actually gates.

- **`TestToWrappedJSON_DEC030ShapeGolden`** (UNCHANGED — restated here as a
  premise anchor). The JSON byte golden from SPEC-051 must still pass verbatim:
  the sparkline never enters JSON. If this test needs ANY edit, the JSON-stays-raw
  invariant was violated — treat that as a spec defect, not a golden update.

- **`TestToWrappedMarkdown_EmptyPeriodNoGlyphLine`** (NEW, small). Over `nil`
  entries with `Spark: true`, assert the markdown is provenance-only (through
  `Entries: 0`) and contains neither `## Cadence` nor a `Cadence: ` glyph line
  (the DEC-014 part-4 body omission means there is no cadence section to decorate).

### `internal/cli/wrapped_test.go`

(Harness mirrors SPEC-051's: `newWrappedTestRoot`, `runWrappedCmd`,
`seedWrappedEntry`, `withNowFunc`. Add a `withLookupSparkEnv` helper that swaps
the `lookupSparkEnv` package var and restores it — same shape as `withNowFunc`,
per AGENTS.md §9 os-state-through-a-package-var.)

- **`TestWrappedCmd_SparkDefaultOn`**. With `NO_COLOR` unset (via
  `withLookupSparkEnv` returning `("", false)`) and no `--no-spark`, a seeded
  corpus's `brag wrapped <year> --format markdown` output contains a line
  matching `^Cadence: [▁▂▃▄▅▆▇█]+$`. Asserts default-on.

- **`TestWrappedCmd_NoSparkFlagSuppresses`**. Same seeded corpus, `brag wrapped
  <year> --no-spark`. Output contains NO line beginning `Cadence: ` and still
  contains the per-month `- YYYY-MM: N` lines (only the glyph line is gone).

- **`TestWrappedCmd_NoColorEnvSuppresses`**. Same seeded corpus, `--no-spark` NOT
  passed, but `withLookupSparkEnv` returns `("1", true)` for `NO_COLOR`. Output
  contains NO `Cadence: ` glyph line. Proves the env escape works independently of
  the flag.

- **`TestWrappedCmd_JSONHasNoGlyphs`**. `brag wrapped <year> --format json` (spark
  defaulted on) — assert the raw JSON bytes contain none of the runes
  `▁▂▃▄▅▆▇█` (JSON stays raw regardless of the spark toggle).

- **`TestWrappedCmd_NoSparkHelpListed`**. `brag wrapped --help` contains the
  `--no-spark` flag (cobra auto-renders the flag; assert the flag name appears —
  unique-token discipline, AGENTS.md §9, `--no-spark` is unique).

## Implementation Context

*Read this section (and the files it points to) before starting the build cycle.*

### Decisions that apply

- `DEC-031` (this spec) — locks the primitive (min→max, `round`, `▁..█`,
  fixed-width one-glyph-per-bucket), the flat/all-zero/single→`▁` and empty→`""`
  edge cases, the `internal/spark` home, the markdown-only `Cadence: <glyphs>`
  placement inside `## Cadence`, default-on with `--no-spark`/`NO_COLOR` escape,
  JSON-stays-raw, and `wrapped`-only scope (stats/impact deferred).
- `DEC-030` — the wrapped envelope + the zero-filled `cadence.series` slot the
  sparkline reads. This spec renders over that slot without reshaping it.
- `DEC-014` — the digest envelope; the empty-state body omission (part 4) is why
  an empty period has no cadence section and hence no glyph line.
- `DEC-001` — pure-Go / local-first / no-CGO; the reason the primitive is Unicode
  block glyphs computed in Go, not an external plotter.
- `DEC-023` (LD6) — the milestone TTY probe is the precedent for reading host
  state (here `NO_COLOR`) through a stdlib-only, injectable package `var`, no
  `go-isatty`/env dependency.
- `DEC-007` — `--no-spark` is a well-formed bool flag; no new user-error surface
  (an unset/false flag is not an error).

### Constraints that apply

- `no-cgo` + `no-new-top-level-deps-without-decision` — `internal/spark` uses only
  `math` (stdlib). No `go.mod`/`go.sum` change. Build cycle MUST confirm
  `git diff main -- go.mod go.sum` is empty and `CGO_ENABLED=0 go build ./...` is
  clean (the DEC-031 acceptance line).
- `stdout-is-for-data-stderr-is-for-humans` — the sparkline is part of the digest
  BODY, which goes to stdout (it is the shareable artifact — see DEC-031 D3 on why
  it is NOT TTY-gated). No stderr involvement.
- `no-sql-in-cli-layer` — unchanged; this spec adds no SQL. The `NO_COLOR` env read
  is `os.LookupEnv` behind the `lookupSparkEnv` package var (not `database/sql`).
- `test-before-implementation` — the Failing Tests above are written first.
- `errors-wrap-with-context` — no new error paths (`--no-spark` false is not an
  error); nothing to wrap beyond the existing `runWrapped` chain.

### Prior related work

- `SPEC-051` (shipped) — `brag wrapped`; created `internal/export/wrapped.go`,
  `WrappedOptions`, the `## Cadence` markdown block, and (via LD8) the
  `aggregate.Cadence` + `CadenceBucket` sparkline-ready slot this spec reads.
  The closest sibling — its `wrapped_test.go` byte-golden shape and CLI test
  harness (`nowFunc`/`seedWrappedEntry`/`withNowFunc`) are copied and extended.
- `SPEC-020` (`stats`) / `SPEC-048` (`impact`) — the two digests DEC-031
  deliberately does NOT touch (no cadence series slot exists in either today).

### Out of scope (for this spec specifically)

- **`stats` and `impact` sparklines.** `stats` has no monthly cadence slot (its
  payload is `entries_per_week`/streak/top-tags/projects/span — scalars and
  rankings); giving it a sparkline requires a NEW lifetime-cadence data slot,
  which is a data-shape decision, not this visual one. `impact` shipped text-pure
  with no series slot. Both are DEC-031-documented deferrals; the SAME primitive
  renders them once a slot exists. **Foreseen follow-up: a `stats` lifetime-cadence
  slot + its sparkline.**
- **An external-plotter pipe** — the explicit LATER layer (STAGE-013 note).
- **Resampling / target-width sparklines** — fixed-width only; a long-series
  consumer (e.g. a future daily cadence) can earn resampling in its own spec
  (DEC-031 Option C2).
- **TTY-gating the sparkline** — deliberately rejected (DEC-031 D3): a shareable
  digest is captured on non-TTY paths; `NO_COLOR`/`--no-spark` are the escape.
- **Color / ANSI** — block glyphs are monochrome; no ANSI escapes (which is also
  why `NO_COLOR` is a fitting, if slightly loose, opt-out signal — it is the
  community-standard "no terminal decoration" flag).
- Any schema change, any network, any LLM.

### Premise audit (AGENTS.md §9 / §12 — additive + inversion cases)

This spec CHANGES two shipped byte-goldens and ADDS a field to a shipped struct.
Walk each per the §12 audit-grep family:

1. **Inversion of a golden's premise (SPEC-051 markdown goldens).**
   `TestToWrappedMarkdown_DEC014FullDocumentGolden` and
   `TestToWrappedMarkdown_QuarterGolden` assert the pre-sparkline `## Cadence`
   block. This spec inserts the `Cadence: <glyphs>` line, so those two `want`
   blocks MUST be updated (enumerated above under Outputs, not discovered at
   build). Every other section of both goldens stays byte-identical.
2. **The JSON golden's premise is PRESERVED, not inverted.**
   `TestToWrappedJSON_DEC030ShapeGolden` must NOT change — the JSON-stays-raw
   invariant. Restated as `TestToWrappedJSON_DEC030ShapeGolden` (unchanged) so
   build treats any needed edit to it as a defect signal.
3. **Additive field on `WrappedOptions` (`Spark bool`).** Grep for
   `WrappedOptions{` across `internal/` at build (`grep -rn "WrappedOptions{"
   internal/`) and confirm every construction site compiles with the new field
   (Go zero-values it to `false` if omitted — so `runWrapped` is the only site
   that must set it explicitly; test sites that want the glyph line set
   `Spark: true`). Enumerate the actual hits at build per the §12 cross-check.
4. **Docs status-change (§9).** `brag wrapped` gains a flag + a visible output
   line. At build, run `grep -rn "brag wrapped\|wrapped" docs/ README.md AGENTS.md`
   and enumerate the hits that describe wrapped's flags/output; add `--no-spark` /
   the cadence sparkline where the command surface is documented. Do NOT touch
   historical DEC-provenance lines (the SPEC-051 build left the DEC-014 *inventory*
   sentences alone for the same reason).

### NOT-contains self-audit (AGENTS.md §12)

`TestToWrappedMarkdown_NoSparkOmitsGlyphLine` and `TestWrappedCmd_NoSparkFlagSuppresses`
assert the output does NOT contain a `Cadence: ` glyph line. The load-bearing prose
that reaches the binary is the renderer's `fmt.Fprintf` format string and the cobra
`Long`/help text. The renderer prints `Cadence: <glyphs>` ONLY under
`if opts.Spark`. The cobra `Long` string must NOT contain a literal `Cadence: ▁…`
example (it may say "a cadence sparkline" in prose — that does not match
`^Cadence: [▁-█]`). Build: grep the locked `Long` string for `Cadence: ` before
locking help text.

## Notes for the Implementer

- **The primitive (`internal/spark/spark.go`):**

```go
package spark

import "math"

// levels is the 8-rung block-glyph ladder, lowest to highest.
var levels = []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

// Line renders vals as a fixed-width Unicode block-glyph sparkline (one
// glyph per element, no resampling) using min→max linear normalization:
// level = round((v-min)/(max-min)*7) over ▁▂▃▄▅▆▇█ (DEC-031). An empty
// series renders ""; a flat series (max==min — includes all-zero and a
// single element) renders all ▁ (min→max has no variation to encode, so
// every element sits at the floor). Pure: no I/O, no clock, no host state.
func Line(vals []int) string {
	if len(vals) == 0 {
		return ""
	}
	min, max := vals[0], vals[0]
	for _, v := range vals {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	out := make([]rune, len(vals))
	if max == min {
		for i := range out {
			out[i] = levels[0]
		}
		return string(out)
	}
	span := float64(max - min)
	for i, v := range vals {
		lvl := int(math.Round(float64(v-min) / span * 7))
		out[i] = levels[lvl]
	}
	return string(out)
}
```

  This is the literal artifact (AGENTS.md §12 literal-artifact-as-spec) — build
  transcribes it verbatim; the `TestLine_NormalizationTable` goldens were computed
  against exactly this code at design time.

- **Renderer wiring (`internal/export/wrapped.go`):**
  - Add `Spark bool` to `WrappedOptions` (documented: "when true and markdown,
    render the cadence sparkline line; JSON ignores it — a sparkline is a visual,
    not data").
  - In `ToWrappedMarkdown`, inside the `## Cadence` block, AFTER the
    `Busiest month:` line and BEFORE the `fmt.Fprintln(&buf)` that precedes the
    per-month loop, add:
    ```go
    if opts.Spark {
        counts := make([]int, len(series))
        for i, b := range series {
            counts[i] = b.Count
        }
        fmt.Fprintf(&buf, "Cadence: %s\n", spark.Line(counts))
    }
    ```
    (`series` is the value already returned by `aggregate.Cadence(entries,
    opts.ScopeMonths)` at the top of the Cadence block.) Import
    `github.com/jysf/bragfile000/internal/spark`.
  - `ToWrappedJSON` is UNCHANGED — it does not read `opts.Spark`.
- **CLI wiring (`internal/cli/wrapped.go`):**
  - Add the flag in `NewWrappedCmd`:
    `cmd.Flags().Bool("no-spark", false, "suppress the in-terminal cadence sparkline")`.
  - Add the env seam near the top of the file (AGENTS.md §9 os-state-via-var):
    ```go
    // lookupSparkEnv reads the NO_COLOR opt-out (no-color.org). Package var
    // so tests inject it deterministically without touching the real env.
    var lookupSparkEnv = os.LookupEnv
    ```
    (add `"os"` to the imports).
  - In `runWrapped`, after resolving `format`, compute:
    ```go
    noSpark, _ := cmd.Flags().GetBool("no-spark")
    noColor, ok := lookupSparkEnv("NO_COLOR")
    // NO_COLOR is "set to any value, even empty" per the convention; the CLI
    // treats a set-but-empty value as opt-out too. Either signal suppresses.
    sparkOn := !noSpark && !(ok && noColor == "" || noColor != "" && ok)
    ```
    Prefer the simpler, equivalent form: `sparkOn := !noSpark && !noColorSet()`
    where `noColorSet()` is `_, ok := lookupSparkEnv("NO_COLOR"); return ok` — the
    `NO_COLOR` convention is "opt out if the variable is present at all, regardless
    of value." Set `opts.Spark = sparkOn`. (Lock the "present-at-all" reading in a
    one-line comment; `TestWrappedCmd_NoColorEnvSuppresses` seeds `("1", true)`,
    and default-on seeds `("", false)`, so both a value and unset are covered; if
    you want to also prove set-but-empty suppresses, add a `("", true)` row.)
  - `opts` construction gains `Spark: sparkOn`.
- **The two updated markdown goldens** differ from SPEC-051 by exactly one inserted
  line each (`Cadence: ▅▅▁█▁▁█▁▁▁▅▁` / `Cadence: █▁▁`). Do NOT edit any other line.
- **Docs sweep** (§9): add the `--no-spark`/`NO_COLOR` behavior to the
  `brag wrapped` section of `docs/api-contract.md`, a one-line mention of the
  cadence sparkline to `docs/tutorial.md`'s wrapped subsection, `--no-spark` to
  `README.md` if wrapped's flags are listed, and a `sparkline` §11 glossary term.
  Enumerate the actual grep hits in `## Outputs` at build (§12 audit-grep
  cross-check).

### Locked design decisions (build-time)

1. **LD1 — min→max linear normalization, `round`, `▁▂▃▄▅▆▇█`.** `level =
   round((v-min)/(max-min)*7)` (Go `math.Round`, half-away-from-zero). Byte-verified
   at design (ramp `0..7 → ▁▂▃▄▅▆▇█`; year `→ ▅▅▁█▁▁█▁▁▁▅▁`; quarter `→ █▁▁`).
   *Rejected:* absolute scale against a fixed ceiling — no meaningful ceiling for
   "entries per month," and it flattens typical small series into the bottom glyphs
   (DEC-031 A2).
2. **LD2 — flat (`max==min`, incl. all-zero and single element) → all `▁`;
   empty → `""`.** min→max has no variation to encode, so the floor `▁` is the
   honest deterministic answer; empty draws nothing. *Rejected:* flat-nonzero → a
   mid/high glyph — needs a magnitude-aware second rule min→max deliberately lacks
   (DEC-031 A′). Accepted tradeoff: a flat non-zero cadence looks like a flat zero
   one; rare, and the adjacent per-month counts disambiguate.
3. **LD3 — fixed width, one glyph per bucket, no resampling.** ≤12-point series;
   one glyph per bucket aligns with the per-month list below it. *Rejected:*
   target-width resampling — unnecessary here and a design surface of its own;
   defer to a long-series consumer (DEC-031 C2).
4. **LD4 — the primitive lives in `internal/spark`.** Pure, SQL-free,
   dependency-free; reusable by a future `stats`/daily cadence without importing
   `export`. *Rejected:* an unexported `export` helper — a future reuse would force
   a duplicate or an out-of-`export` refactor (same reasoning as SPEC-051 LD8 for
   `Cadence`).
5. **LD5 — sparkline is markdown-only; JSON stays raw.** Glyphs never enter any
   JSON envelope; the SPEC-051 JSON golden is untouched. *Rejected:* a
   `cadence.sparkline` JSON key — duplicates data in a non-machine-readable form
   and couples the wire format to a terminal concern (DEC-031 f).
6. **LD6 — placement: a labeled `Cadence: <glyphs>` line inside `## Cadence`,
   between `Busiest month:` and the per-month list.** Self-describing in a shared
   digest; adjacent to the counts it visualizes. *Rejected:* an unlabeled bare
   glyph line (ambiguous when pasted out of context); putting it in a separate
   section (over-weights a one-line visual).
7. **LD7 — default-on in markdown; `--no-spark` OR non-empty/ present `NO_COLOR`
   suppresses (either wins).** Default-on is the requested "fun"; `NO_COLOR`
   (no-color.org) is the community-standard opt-out and honoring it is free;
   `--no-spark` is the explicit per-invocation escape. *Rejected:* opt-in behind
   `--spark` (a default-off visual is unseen); TTY-gating like the milestone line
   (DEC-031 D3 — a shareable digest is captured on non-TTY paths, so a TTY gate
   fights the command's intent).
8. **LD8 — `NO_COLOR` "present-at-all" semantics, read through the
   `lookupSparkEnv` package var.** Opt out if `NO_COLOR` is set to ANY value
   (including empty), per the convention. The env read is behind an injectable
   package var (AGENTS.md §9 os-state-via-var; DEC-023 LD6 precedent) so tests are
   deterministic and hermetic. *Rejected:* reading `os.LookupEnv` directly (leaks
   host env into tests); requiring a non-empty value (departs from the convention).
9. **LD9 — `wrapped` only; `stats`/`impact` deferred.** `wrapped` has the prepared
   `cadence.series` slot; `stats`/`impact` have no series slot (adding one is a
   data-shape decision, not this visual one). *Rejected:* folding a `stats`
   lifetime-cadence slot into this spec — reshapes the `stats` envelope and bloats
   past S/M (DEC-031 e). Foreseen as a clean follow-up.

### Rejected alternatives (build-time)

- **Rendering the sparkline in the empty-period digest.** Rejected: DEC-014 part 4
  omits the whole body (including `## Cadence`) when there are no entries, so there
  is no cadence section to decorate. `TestToWrappedMarkdown_EmptyPeriodNoGlyphLine`
  pins this.
- **A `Spark` field default of `true`.** Rejected: Go zero-values `false`, so every
  existing `WrappedOptions{...}` construction that omits `Spark` (all the SPEC-051
  tests except the ones this spec updates) keeps its current no-glyph behavior for
  free — the explicit `Spark: true` at the two updated goldens + `runWrapped` is
  the only opt-in surface, which keeps the additive-field blast radius minimal.
- **Adding the sparkline to the JSON as a convenience.** Covered in LD5.
- **A `go-isatty`/`fatih/color`-style dependency for terminal detection.**
  Rejected: violates `no-new-top-level-deps-without-decision` and `no-cgo`-adjacent
  minimalism; the milestone line already proved stdlib-only host probing (DEC-023
  LD6), and here we do not even probe the TTY (LD7).

---

## Build Completion

*Filled in at the end of the **build** cycle, before advancing to verify.*

- **Branch:** feat/spec-052-sparklines
- **PR (if applicable):** #90 (ready for review, not merged).
- **All acceptance criteria met?** Yes — all eight AC boxes. The primitive
  is min→max/`round`/`▁▂▃▄▅▆▇█`, one glyph per element; edge cases
  (empty→`""`, flat/all-zero/single→all `▁`, ramp `0..7`→`▁▂▃▄▅▆▇█`) pass;
  the year fixture renders `Cadence: ▅▅▁█▁▁█▁▁▁▅▁` and the quarter `Cadence:
  █▁▁` byte-exact; `--no-spark`/`NO_COLOR` each suppress independently; the
  JSON golden is byte-unchanged; the empty period has no cadence section
  hence no glyph line; `stats`/`impact` untouched. `go.mod`/`go.sum` are
  byte-identical to main and `CGO_ENABLED=0 go build ./...` is clean.
- **New decisions emitted:** DEC-031 (emitted at design).
- **Files changed:**
  - Created `internal/spark/spark.go` (the `Line` primitive + `levels`
    table, verbatim from the spec's literal artifact) and
    `internal/spark/spark_test.go` (the normalization table + length +
    glyph-table tests, 13 subtests).
  - `internal/export/wrapped.go` — `Spark bool` on `WrappedOptions`; the
    `Cadence: <glyphs>` line inside `## Cadence` under `if opts.Spark`;
    `internal/spark` import. `ToWrappedJSON` unchanged.
  - `internal/export/wrapped_test.go` — the two SPEC-051 markdown goldens
    now run with `Spark: true` + the inserted glyph line; new
    `TestToWrappedMarkdown_NoSparkOmitsGlyphLine` (byte-identical to the
    pre-sparkline document) and `TestToWrappedMarkdown_EmptyPeriodNoGlyphLine`.
    The JSON golden is untouched.
  - `internal/cli/wrapped.go` — `--no-spark` flag; `lookupSparkEnv`
    package-var env seam (`os.LookupEnv`); `sparkOn := !noSpark &&
    !noColorSet` set into `opts.Spark`.
  - `internal/cli/wrapped_test.go` — `withLookupSparkEnv` helper +
    `TestWrappedCmd_SparkDefaultOn` / `_NoSparkFlagSuppresses` /
    `_NoColorEnvSuppresses` / `_NoColorEnvEmptyValueSuppresses` /
    `_JSONHasNoGlyphs` / `_NoSparkHelpListed`.
  - `docs/api-contract.md`, `docs/tutorial.md`, `AGENTS.md` §11
    (`sparkline` term + refreshed `wrapped` term), `guidance/questions.yaml`
    (question resolved).
- **Docs sweep (§9 status-change / §12 audit-grep cross-check):**
  `grep "WrappedOptions{"` → only `runWrapped` sets `Spark` in production;
  every test site zero-values it except the four that explicitly opt in.
  `grep "wrapped" docs/ README.md AGENTS.md` → api-contract's `## Cadence`
  description + flags list, tutorial's wrapped subsection (its stale
  "text-first, ready for a later visual pass" tail replaced), and AGENTS
  §11 all updated. README lists `wrapped` only as a single illustrative
  example line (flags not enumerated there, matching every other command),
  so `--no-spark` was intentionally not added — no flag inventory to keep in
  sync. The `Long` string was grepped for `Cadence: ` (no glyph-literal
  example) before locking the NOT-contains tests.
- **Deviations from spec:** None material. Two minor, in-spirit choices:
  (1) added `TestWrappedCmd_NoColorEnvEmptyValueSuppresses` (the spec's
  optional `("", true)` set-but-empty row) to fully pin the present-at-all
  `NO_COLOR` reading; (2) used the spec's preferred simpler
  `sparkOn := !noSpark && !noColorSet` form (the `_, ok := lookupSparkEnv`
  present-at-all reading) rather than the longer intermediate expression.
- **Follow-up work identified:** The DEC-031 deferral stands — a `stats`
  lifetime-cadence data slot + its sparkline is the foreseen clean
  follow-up (the same primitive renders it once a slot exists). No new
  follow-up surfaced during build.

### Build-phase reflection (3 questions, short answers)

1. **What was unclear in the spec that slowed you down?** Nothing slowed the
   build — the spec was unusually complete. The literal `spark.Line`
   artifact, the byte-exact normalization table, and the two enumerated
   golden diffs meant the primitive matched every golden first-try; the
   golden-faithfulness step confirmed rather than corrected.
2. **Was there a constraint or decision that should have been listed but
   wasn't?** No. The README "flags enumerated?" conditional was the one
   judgment call left to build, and the spec framed it correctly as
   conditional — README carries only an example line, so nothing to add.
3. **If you did this task again, what would you do differently?** Nothing
   structural. Implementing + running the `spark` package in isolation
   FIRST (before any renderer wiring) was the right order — it de-risked the
   goldens before they were embedded in the wrapped tests.

---

## Verify

*Fresh, independent verify session per AGENTS.md §12/§13 — re-derived from the
spec + DEC-031 + constraints, not the build's self-report.*

**Verdict: ✅ APPROVED.**

- **Six gates (independently re-run):** `go test ./...` PASS (700 tests, 11
  packages, exit 0); `gofmt -l .` empty (exit 0); `go vet ./...` clean (exit 0);
  `CGO_ENABLED=0 go build ./...` clean (exit 0); `just test-docs` ALL OK (exit
  0); `just test-hook` ALL OK (exit 0).
- **Acceptance criteria:** all eight met. Primitive is min→max/`math.Round`/
  `▁▂▃▄▅▆▇█` one-glyph-per-element (`internal/spark/spark.go:18-44`); edge cases
  empty/nil→`""`, flat/all-zero/single→all `▁`, ramp `0..7`→`▁▂▃▄▅▆▇█`; pure,
  `math`-only, `go.mod`/`go.sum` byte-unchanged vs main. The `Cadence: <glyphs>`
  line renders inside `## Cadence` between `Busiest month:` and the per-month
  list (`internal/export/wrapped.go:71-77`); year fixture → `Cadence: ▅▅▁█▁▁█▁▁▁▅▁`,
  quarter → `Cadence: █▁▁` (byte-exact goldens). `--no-spark` OR present
  `NO_COLOR` each suppress independently (`internal/cli/wrapped.go:214-216`).
  JSON unchanged; empty period has no cadence section hence no glyph line;
  `stats`/`impact` untouched.
- **Independent normalization re-derivation:** a from-scratch transcription of
  the DEC-031 algorithm (NOT importing the repo's `spark`) reproduced every
  golden byte-for-byte, including the load-bearing year/quarter/ramp rows;
  output length == input length; only block glyphs or empty. No mis-normalized
  golden (SPEC-049 lesson clean).
- **JSON invariant:** `TestToWrappedJSON_DEC030ShapeGolden` is byte-identical to
  main (function-body diff empty); zero glyphs in the runtime JSON envelope; no
  glyph literal in `wrapped.go`'s JSON path. The two markdown goldens each gained
  exactly one inserted `Cadence:` line — no other line touched, no glyph line
  removed.
- **Live default-on + suppression:** built binary, seeded corpus. Default
  `brag wrapped 2026` → `Cadence: ▁▁▁▁▁▁█▁▁▁▁▁`; `--no-spark` → no glyph line,
  per-month list intact; `NO_COLOR=1` → no glyph line; `NO_COLOR=` (empty but
  present) → no glyph line (present-at-all semantics); piping default to a file
  → glyph line survives (NOT TTY-gated); `--format json` → zero glyphs.
- **No new dependency / no smuggled decision:** `internal/spark` imports only
  stdlib `math`; `go.mod`/`go.sum` byte-unchanged; DEC-031 covers the decision,
  no DEC-032; the `guidance/questions.yaml` sparkline question is `status:
  resolved` linking DEC-031. No `database/sql` in `internal/spark` or the wrapped
  renderer.

Advancing to **ship** (merge orchestrated separately; PR #90 stays open).

## Reflection (Ship)

*Appended at ship (2026-07-07). Merged to main; no release — ships in the v0.4.0
cut (SPEC-054).*

1. **What would I do differently next time?** — Nothing notable. Deferring the
   sparkline out of `impact`/`wrapped` into a dedicated pass (with `wrapped`
   pre-shaping a sparkline-ready `cadence.series`) paid off: this spec was a
   clean, additive, one-inserted-line-per-golden change with a self-contained
   pure primitive. Pre-shaping the data slot in the earlier spec is the reusable
   move.
2. **Does any template, constraint, or decision need updating?** — No. The
   `internal/spark` primitive is pure/stdlib-only and the visual stayed
   markdown-only with JSON untouched — no constraint tension. The
   default-on-with-`NO_COLOR`-escape pattern reused DEC-023's injectable-env
   precedent cleanly.
3. **Is there a follow-up spec I should write now before I forget?** — No new
   ones. The foreseen `stats` cadence-sparkline (needs a new data-shape decision,
   not a visual one) is documented in DEC-031 as a clean future follow-up, not
   forgotten. STAGE-013's remaining backlog (`--previous`, P3 metric, the v0.4.0
   cut) is intact.
