---
# Maps to ContextCore insight.* semantic conventions.

insight:
  id: DEC-031                         # stable, never reused
  type: decision                     # decision | analysis | recommendation | observation
  confidence: 0.78                   # 0.0 - 1.0, honest assessment
  audience:                          # who needs to know?
    - developer
    - agent

agent:
  id: claude-opus-4-8
  session_id: null

# Decisions are repo-level, but it's useful to track which project
# caused them to be emitted.
project:
  id: PROJ-004                       # the project during which this was decided
repo:
  id: bragfile

created_at: 2026-07-06
supersedes: null
superseded_by: null

tags:
  - sparkline
  - visual
  - local-first
  - wrapped
  - normalization
---

# DEC-031: Sparkline primitive — normalization, placement, and default-on-with-escape

## Decision

Add a pure, zero-dependency `internal/spark` package whose sole exported
function `spark.Line([]int) string` maps a numeric series to a fixed-width
Unicode block-glyph string (`▁▂▃▄▅▆▇█`, one glyph per bucket, no resampling),
using **min→max linear normalization** with `level = round((v-min)/(max-min)*7)`
(Go `math.Round`, half-away-from-zero); a **flat** series (`max==min`, which
includes all-zero and single-point series) renders **all `▁`**, and an **empty**
series renders `""`. The sparkline is a **markdown-only** rendering, added to the
`brag wrapped` `## Cadence` section as a labeled `Cadence: <glyphs>` line over the
existing `cadence.series` counts; **JSON is unchanged** (raw counts only — a
sparkline is a visual, not data). It is **default-on in markdown** with an escape
hatch: a `--no-spark` flag OR a non-empty `NO_COLOR` environment variable
suppresses the glyph line (either one wins). `stats` and `impact` do **not** get
a sparkline in this spec.

## Context

STAGE-013's "dedicated visual pass," deferred on purpose from `brag impact`
(which shipped text-pure) and foreseen throughout SPEC-051, which left
`wrapped`'s `cadence.series` as a zero-filled, sparkline-ready `[]{period,count}`
slot. The user asked for this ("sparklines could be fun") and it should make the
digests feel alive. Hard constraints: **local-first, ZERO new dependency, no
network** (DEC-001, `no-cgo`, `no-new-top-level-deps-without-decision`) — Unicode
block characters only, computed in pure Go. An external-plotter pipe is an
explicitly LATER layer, not this spec (STAGE-013 design note). The open questions
this DEC settles are: (a) the primitive's normalization + edge-case behavior;
(b) where the primitive lives; (c) fixed-width vs resampled; (d) where it renders
and whether default-on or opt-in; (e) which digests get it now; (f) that JSON
stays raw.

## Alternatives Considered

### (a) Normalization rule

- **Option A1: min→max linear across 8 levels (chosen).**
  - What it is: `level = round((v-min)/(max-min)*7)`, indexing `▁▂▃▄▅▆▇█`.
    Verified byte-exact at design time (scratch program): a `0..7` ramp maps to
    the full `▁▂▃▄▅▆▇█`; the SPEC-051 year series `[1,1,0,2,0,0,2,0,0,0,1,0]`
    maps to `▅▅▁█▁▁█▁▁▁▅▁`; `[2,0,0]` maps to `█▁▁`.
  - Why selected: it is the standard sparkline scaling (holman/spark and most
    min-max implementations), maximizes visible contrast within any series (the
    max always hits `█`, the min always hits `▁`), and is trivially
    hand-checkable, which is exactly what the "goldens must be correct at design
    time" discipline (the SPEC-049 lesson) demands.

- **Option A2: absolute scale (0 → `▁`, fixed ceiling → `█`).**
  - What it is: normalize against a fixed maximum (e.g. a hard-coded 10 or the
    corpus-wide max) rather than the series max.
  - Why rejected: there is no meaningful absolute ceiling for "entries in a
    month," and a fixed ceiling flattens most real series into the bottom two
    glyphs (a typical month has 1–5 entries), defeating the "feel alive" goal.
    Min→max is self-scaling per series, which reads best for a per-period reel.

### (a′) Flat / all-zero / single-point / empty behavior

- **Chosen: flat (`max==min`, incl. all-zero and `len==1`) → all `▁`; empty →
  `""`.**
  - Why selected: with min→max scaling there is no variation to encode, so every
    bucket lands at the floor of the normalized range — `▁` is the honest,
    deterministic answer and matches holman/spark. Empty must be `""` because
    there is nothing to draw (the caller decides whether to print the label).
- **Rejected: flat-nonzero → mid or high glyph (e.g. `▄`/`█`).**
  - What it is: render a flat non-zero series at a mid/high level to signal
    "consistent activity."
  - Why rejected: it requires a second, magnitude-aware rule that min→max
    scaling deliberately does not have (there is no reference magnitude), and it
    would make a flat series render differently from an all-zero series with no
    principled boundary between them. `▁` for both is simpler and defensible.
    Documented as the honest tradeoff: a flat non-zero month-series looks like a
    flat zero one — acceptable, because a flat cadence is rare over a real
    year/quarter and the adjacent per-month counts disambiguate it.

### (b) Where the primitive lives

- **Option B1: a new `internal/spark` package (chosen).**
  - What it is: `spark.Line([]int) string` in its own tiny SQL-free,
    dependency-free package.
  - Why selected: it is a pure rendering primitive with no export-envelope or
    aggregate coupling; a dedicated package keeps it reusable by `wrapped` now
    and `stats`/`impact` later without importing the whole `export` surface, and
    gives it a clean, exhaustively unit-testable home (mirrors how
    `internal/aggregate` isolates pure data logic).
- **Option B2: an unexported helper inside `internal/export`.**
  - Why rejected: a future `stats` cadence sparkline (foreseen below) would have
    to either duplicate it or export it out of `export`; putting the primitive in
    its own package now avoids that refactor — the same reasoning SPEC-051 LD8
    used to put `Cadence` in `aggregate` rather than `export`.

### (c) Fixed-width vs resampled

- **Option C1: fixed width — one glyph per bucket, no resampling (chosen).**
  - Why selected: the wrapped cadence is 12 buckets (year) or 3 (quarter) —
    small enough to render one glyph each with zero ambiguity, and one-glyph-
    per-bucket keeps the sparkline aligned conceptually with the per-month list
    printed directly below it. Byte-exact and trivially hand-checkable.
- **Option C2: target-width resampling (bucket down to N glyphs).**
  - Why rejected: unnecessary for ≤12-point series and it introduces a
    resampling rule (averaging/decimation) that is a real design surface of its
    own — defer until a consumer actually has a long series (a future daily
    cadence could earn it; not now).

### (d) Placement + default-on vs opt-in

- **Option D1: default-on in markdown, `--no-spark` OR `NO_COLOR` escape
  (chosen).**
  - What it is: the glyph line renders by default inside `## Cadence` as
    `Cadence: <glyphs>`; passing `--no-spark` or setting a non-empty `NO_COLOR`
    suppresses it (either wins).
  - Why selected: Unicode block glyphs are broadly supported in modern
    terminals, and default-on is the "fun" the user asked for. `NO_COLOR` is the
    community-standard opt-out signal for terminal decoration
    (no-color.org), and honoring it costs nothing; `--no-spark` gives an
    explicit per-invocation escape for plain/legacy terminals or when piping the
    markdown somewhere glyph-hostile. The repo already has a stdlib-only,
    dependency-free precedent for this posture: the milestone line
    (`internal/cli/milestone.go`, DEC-023 LD6) auto-suppresses off a TTY via
    `os.ModeCharDevice`. We deliberately do NOT gate the sparkline on TTY,
    though (see D3).
- **Option D2: opt-in behind `--spark` (default off).**
  - Why rejected: hides the feature the user explicitly wanted; a default-off
    visual is a visual nobody sees.
- **Option D3: auto-suppress off a TTY (like the milestone line).**
  - What it is: only print the sparkline when stdout is a char device.
  - Why rejected for the sparkline: the wrapped digest's whole purpose is to be
    **captured and shared** (piped to a file, pasted into a doc/PR) — precisely
    the non-TTY paths where a TTY gate would wrongly strip the visual. The
    milestone line is ephemeral stderr chatter (correct to hide when not
    interactive); a shareable digest is the opposite. `NO_COLOR`/`--no-spark`
    give the user explicit control without a TTY heuristic that fights the
    command's intent.

### (e) Which digests get it now

- **Chosen: `wrapped` cadence only.**
  - Why: `wrapped` already has the prepared, zero-filled `cadence.series` slot
    (SPEC-051) — rendering a sparkline over it reshapes nothing.
- **Rejected (deferred): `stats`.**
  - Why deferred: `stats` exposes **no** monthly cadence series today (its
    payload is `entries_per_week`, `current/longest_streak`, top tags/projects,
    corpus span — all scalars/rankings). Giving `stats` a sparkline requires
    designing a NEW lifetime-cadence data slot (new JSON keys + a
    span→months derivation), which is a separate data-shape decision, not a
    visual one. Folding it in here would reshape the `stats` envelope and bloat
    this spec past S/M. Foreseen as a clean follow-up: once a lifetime-cadence
    slot exists in `stats`, the SAME `spark.Line` primitive renders it.
- **Rejected (deferred): `impact`.**
  - Why deferred: `impact` shipped text-pure by design and has no series slot
    (its payload is counts-by-project + impact-by-project). No cadence series to
    sparkline. If `impact` ever gains a cadence slot it reuses the primitive.

### (f) JSON

- **Chosen: JSON unchanged — raw counts only.**
  - A sparkline is a lossy visual rendering of `cadence.series[].count`; putting
    glyphs in JSON would duplicate data in a non-machine-readable form and couple
    the wire format to a terminal-rendering concern. Consumers that want a
    sparkline render one from the raw counts. This preserves the SPEC-051
    envelope byte-for-byte (no golden churn on the JSON side).

## Consequences

- **Positive:** the headline shareable digest gains a compact visual with zero
  new dependency and zero JSON churn; the primitive is a reusable, pure,
  exhaustively-testable package (`internal/spark`) ready for `stats`/`impact`/a
  future daily cadence; `NO_COLOR` support lands a community-standard escape the
  whole CLI can later reuse.
- **Negative:** a flat non-zero cadence renders identically to a flat zero one
  (min→max has no magnitude reference) — accepted, rare, and disambiguated by the
  adjacent per-month counts. Default-on means a glyph-hostile terminal sees
  mojibake unless the user knows `--no-spark`/`NO_COLOR` — mitigated by the escape
  and by block glyphs being near-universally supported.
- **Neutral:** the markdown `## Cadence` section grows one line; the year golden's
  Cadence block gains `Cadence: ▅▅▁█▁▁█▁▁▁▅▁`, the quarter's `Cadence: █▁▁`. The
  SPEC-051 JSON goldens are untouched.

## Validation

- Right if: the wrapped digest reads more alive at a glance, the primitive is
  reused by a later `stats`/daily-cadence spec without change, and no user files a
  "garbage characters in my digest" issue that `NO_COLOR`/`--no-spark` didn't
  already answer.
- Revisit if: a consumer needs a long series (→ resampling, Option C2 earned), or
  the flat-nonzero-looks-empty tradeoff bites in practice (→ a magnitude-aware
  variant), or an external-plotter pipe supersedes in-terminal glyphs for the
  shareable path.
- Confidence 0.78: the primitive + normalization + placement are locked and
  byte-verified; the residual uncertainty is the default-on ergonomics (glyph
  support across the user's real terminals) and the flat-nonzero convention — both
  reversible and escape-hatched. Per §14 (< 0.8), a question is logged in
  `guidance/questions.yaml`.

## References

- Related specs: SPEC-052 (this decision's spec), SPEC-051 (`wrapped`, prepared
  the `cadence.series` slot), SPEC-020 (`stats`), SPEC-048 (`impact`).
- Related decisions: DEC-030 (wrapped period + cadence.series slot), DEC-014
  (digest envelope), DEC-001 (pure-Go / no-CGO / local-first), DEC-023 (milestone
  TTY-detection precedent, LD6 — stdlib-only, no go-isatty).
- Constraints: `no-cgo`, `no-new-top-level-deps-without-decision`,
  `stdout-is-for-data-stderr-is-for-humans`, `test-before-implementation`.
- External: no-color.org (the `NO_COLOR` convention); holman/spark (min→max block
  sparkline prior art).
