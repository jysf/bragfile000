---
insight:
  id: DEC-008
  type: decision
  confidence: 0.80
  audience:
    - developer
    - agent

agent:
  id: claude-opus-4-7
  session_id: null

project:
  id: PROJ-001
repo:
  id: bragfile

created_at: 2026-04-20
supersedes: null
superseded_by: null

tags:
  - cli
  - time
  - parsing
---

# DEC-008: `--since` accepts `YYYY-MM-DD` and `Nd`/`Nw`/`Nm` durations

## Decision

The `--since` flag on `brag list` (and any future list-like command —
`search`, `summary`) accepts exactly two input formats:

1. **ISO date** `YYYY-MM-DD` — parsed as midnight UTC on that day.
2. **Relative duration** `Nd`, `Nw`, `Nm` where `N` is a positive
   integer. Units are:
   - `d` = days (N × 24h)
   - `w` = weeks (N × 7 × 24h)
   - `m` = months (N × 30 × 24h, **approximate** — calendar months
     aren't a fixed length)

   The result is `time.Now().UTC() - duration` as a `time.Time`.

Any other input (including empty, `N` missing, unknown unit, negative
`N`, whitespace, etc.) returns a `UserError` from the CLI layer.
Storage layer sees only parsed `time.Time` values.

## Context

`brag list --since` needs to answer "show me entries from the last
week" with an ergonomic argument format. Two reasonable user mental
models exist: "from a specific date" and "from N time-units ago". The
CLI should serve both without becoming either a mini-date-math-DSL or
a dependency on a heavyweight date library.

## Alternatives Considered

- **Option A: ISO-8601 durations (`P7D`, `PT24H`)**
  - What it is: The formal ISO duration syntax.
  - Why rejected: Users don't type `P7D`; they type `7d`. Every
    CLI that's ever used ISO durations has a FAQ about it.

- **Option B: Named aliases only (`yesterday`, `today`, `last-week`,
  `last-month`)**
  - What it is: Small vocabulary of human words.
  - Why rejected: Vocabulary creep — someone will want
    `the-day-before-yesterday`, `two-weeks-ago`, localized aliases.
    Date arithmetic is the cleaner primitive; aliases can layer on
    top later if demand justifies it.

- **Option C: Go's `time.ParseDuration` syntax (`168h`, `10080m`)**
  - What it is: Reuse the stdlib parser.
  - Why rejected: `time.ParseDuration` supports only ns/µs/ms/s/m/h
    — no days, no weeks, no months. Hours are too fine-grained for
    brag retrospection; "give me 168 hours" is unnatural. Also
    `m` means minutes, conflicting with our preferred "months"
    reading.

- **Option D (chosen): ISO date + `Nd`/`Nw`/`Nm` durations**
  - What it is: Two formats, both short, both obvious, both common
    in prior art (`git log --since`, `kubectl logs --since`, many
    others use a close variant).
  - Why selected: Covers the two natural user mental models. Small
    parser (<30 lines, pure function, unit-testable in isolation).
    Uses unit letters that don't collide with Go's
    `time.ParseDuration` (which has no `d`/`w` and would interpret
    `Nm` as minutes — we own the whole parser here so there's no
    confusion in our code).

## Consequences

- **Positive:** Parse logic is small, self-contained, covered by
  pure-function unit tests. Users type natural input (`7d`, `2w`).
  Future list-like commands inherit the parser by calling the same
  helper in `internal/cli`.
- **Negative:** Month length is approximated as 30 days. A user asking
  for `--since 1m` on 2026-03-15 will get results since 2026-02-13,
  not 2026-02-15. Documented, acceptable at MVP scale (no legal or
  billing implications in a personal brag tool).
- **Neutral:** The parser lives in the CLI layer (`internal/cli/
  since.go`). Storage only sees `time.Time`, keeping the storage
  layer agnostic to input-format conventions.

## Validation

This decision is right if:
- Users don't ask for other `--since` formats within PROJ-001
  (or STAGE-003's `summary`).
- No test relies on exact calendar-month math via `--since Nm`.

Revisit if:
- Calendar-correct month math becomes required (e.g., "first of the
  month" semantics). At that point consider pulling in a date
  library or expanding the parser to handle named boundaries
  (`this-week`, `last-month-start`) instead of extending the
  approximate-duration math.
- A second CLI command has an orthogonal time-filter need that
  doesn't map cleanly to "since a point in time" (unlikely).

## References

- Related specs: SPEC-007 (introduces `--since` on `brag list`;
  the parser lives in `internal/cli/since.go` per that spec).
- Related decisions: DEC-007 (validation failures return
  `UserErrorf` from the CLI layer — applies to `--since` parse
  errors).
- Related constraints: `stdout-is-for-data-stderr-is-for-humans`
  (parse errors go to stderr via `main.go`).
- External precedent: `git log --since`, `kubectl logs --since`,
  and similar CLIs accept comparable formats.
