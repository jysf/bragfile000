---
# Maps to ContextCore insight.* semantic conventions.

insight:
  id: DEC-022
  type: decision
  confidence: 0.88                    # honest: semantics are well-grounded; residual
                                       # uncertainty is only the "local" source for a
                                       # future non-CLI caller (SPEC-040 MCP), flagged below
  audience:
    - developer
    - agent

agent:
  id: claude-opus-4-8
  session_id: null

# Decisions are repo-level, but it's useful to track which project
# caused them to be emitted.
project:
  id: PROJ-003
repo:
  id: bragfile

created_at: 2026-07-03
supersedes: null                       # revises SPEC-020 §6(a)/(b) inline locked
                                        # decisions, not a prior DEC — see Context
superseded_by: null

tags:
  - aggregate
  - stats
  - streak
  - timezone
  - derived-metric
  - dst
---

# DEC-022: current-streak is a local-day derived metric, alive through yesterday

## Decision

`aggregate.Streak` computes the **current** streak in the user's **local
calendar day** and keeps it **alive through yesterday**: the current
streak is the length of the consecutive run of local days with ≥1 entry
that ends on **today or yesterday** (0 only once *both* today and
yesterday are empty). The "local" timezone is the location carried by the
injected `now time.Time` (`now.Location()`) — no new parameter, no package
global. **Storage is unchanged:** timestamps remain UTC RFC3339 per the
blocking `timestamps-in-utc-rfc3339` constraint; only this *derived
metric* localizes, at read time, by converting each stored UTC instant
into `now.Location()` before bucketing. All day arithmetic (cursor
stepping, longest-run adjacency) uses calendar operations
(`AddDate`/date-label compare), never instant subtraction, so it is
correct across DST transitions.

## Context

`brag stats` current-streak was a **confirmed defect** (backlog.md "BUG:
`brag stats` current-streak reads 0 until you log today", 2026-06-20): an
unbroken 14-day run reported `Current: 0` for the whole part of a day
before the day's first re-log. Two compounding root causes in
`aggregate.Streak` (`internal/aggregate/aggregate.go:172`):

1. **Requires-today.** The cursor seeds at *today* and breaks immediately
   if today has no entry — so every UTC-midnight zeroes an intact streak
   until you re-log. Standard streak semantics (GitHub, Duolingo) keep the
   streak alive through yesterday and only zero it after a *full* empty
   day.
2. **UTC-bucketed.** Entries bucket by `CreatedAt.UTC()`, so an
   evening-Pacific entry lands on what UTC calls the *next* day; "today"
   rolls over hours before the user's local day ends.

This reverses two choices that SPEC-020 locked **inline** (spec §6, not a
prior DEC), which is why a DEC is warranted rather than another inline
note — the reversal needs a durable, constraint-level home that outlives
the archived spec:

- SPEC-020 §6(a): "Streak boundaries are UTC calendar days … Time-zone
  handling deferred to backlog." → **now taken up:** local-day.
- SPEC-020 §6(b): "`now`'s UTC date with zero entries → `current_streak
  == 0` … today-without-entries breaks the streak immediately." →
  **reversed:** alive through yesterday.

The load-bearing question this DEC settles (STAGE-009 Design Notes,
surfaced question (b)): does localizing a derived metric require *relaxing*
the blocking `timestamps-in-utc-rfc3339` constraint? **No.** That
constraint governs what is *written to SQLite* (`internal/storage/**`).
The streak is computed in `internal/aggregate` over already-stored UTC
instants; converting a UTC instant to a local day for a display metric
writes nothing and stores nothing. Storage stays UTC; the constraint is
untouched. This DEC records that carve-out explicitly so a future reader
does not mistake the localization for a constraint violation, and so the
next calendar-derived metric (STAGE-010's `brag impact
--quarter|--month|--year`) inherits the same "derive-local, store-UTC"
rule instead of re-litigating it.

## Alternatives Considered

- **Option A: keep UTC-day, only fix requires-today.**
  - What it is: add alive-through-yesterday but keep UTC bucketing.
  - Why rejected: leaves root-cause #2. The user is Pacific; UTC rollover
    still mis-dates evening entries and would still surprise the streak
    (and the SPEC-039 streak milestone) near midnight. Half a fix.

- **Option B: add an explicit `loc *time.Location` parameter —
  `Streak(entries, now, loc)`.**
  - What it is: pass the timezone as a separate third argument.
  - Why rejected: `now` already carries a `Location()`; a separate `loc`
    creates two sources of truth for "when is the day boundary" that can
    disagree, and it churns the signature and every call site + both test
    files for zero added expressiveness. The injected `now` is already the
    §9 clock seam; the location rides it for free.

- **Option C: read `time.Local` / the OS zone directly inside `Streak`.**
  - What it is: `time.Now().Location()` or `time.Local` inside the pure
    function.
  - Why rejected: breaks the §9 injectable-seam discipline — the function
    would become non-deterministic and untestable across zones/DST without
    an env dance. The whole point of the injected `now` is that tests pin
    both the instant *and* the zone.

- **Option D (chosen): local day via `now.Location()`, alive through
  yesterday, calendar arithmetic throughout, storage stays UTC.**
  - What it is: bucket `e.CreatedAt.In(now.Location())`; seed the
    current-streak cursor at today, or yesterday if today is empty; step
    with `AddDate`; detect longest-run adjacency by comparing date labels
    via `AddDate`, never `Sub == 24h`.
  - Why selected: fixes both root causes; keeps the signature stable;
    keeps the injected-clock seam and makes the *zone* injectable through
    the same seam; DST-correct by construction; and localizes only the
    derived value, leaving the blocking UTC-storage constraint fully
    intact.

## Consequences

- **Positive:** `brag stats` current-streak now matches user intuition
  (an intact run reads its true length all day, not 0-until-relog).
  SPEC-039's streak milestone fires on a correct number — the reason
  SPEC-038 blocks SPEC-039. The "derive-local, store-UTC" carve-out is now
  a reusable, referenceable rule for STAGE-010's calendar-window metrics.
  DST-correctness is structural, not incidental.
- **Negative:** `internal/cli/stats.go` must stop pre-stripping the clock
  to UTC (`time.Now().UTC()` → `time.Now()`) so a real local zone reaches
  the metric; the "single `Now` source" decision (SPEC-020 §10) is
  preserved (still one `time.Now()` call; the `Generated:` line re-`.UTC()`s
  it explicitly). **current** and **longest** now bucket by *different*
  reference frames only in the sense that both use `now.Location()`, which
  for the CLI is the host zone — a machine whose `TZ`/system zone is wrong
  will bucket wrong (acceptable: the same host clock the user reads).
- **Neutral:** the existing `export`/`cli` stats goldens are unaffected
  because their fixtures put `now` in `time.UTC` with a *today* entry — the
  backward-compatible case (see spec premise audit). Only the two new
  timezone-specific aggregate tests exercise the local/DST behavior.

## Validation

Right if: the aggregate tests lock (i) alive-through-yesterday, (ii)
still-zero after two empty days, (iii) local-day bucketing that flips the
answer vs UTC, and (iv) a current-streak run that steps across a
spring-forward DST boundary — all with an injected `now` (zone + instant),
no `time.Sleep`. Revisit if: a non-CLI caller (SPEC-040's MCP `brag_stats`)
cannot supply a meaningful `now.Location()` — then the "location rides on
`now`" mechanism (not the local-day *policy*) may need an explicit zone
input; the policy in this DEC stands regardless. Also revisit if a
multi-host/synced-corpus future makes "the user's local day" ambiguous
(out of scope now; MCP is local-only per the brief).

## References

- Related specs: SPEC-038 (emits this DEC), SPEC-039 (consumes the
  corrected streak), SPEC-020 (locked the UTC-day + requires-today
  semantics this revises, §6)
- Related decisions: DEC-011 (entry shape / stored UTC timestamps)
- Related constraints: `timestamps-in-utc-rfc3339` (blocking — untouched;
  this DEC records why localizing a derived metric does not relax it)
- Discussions: backlog.md "BUG: `brag stats` current-streak reads 0 until
  you log today (confirmed)"; STAGE-009 Design Notes surfaced question (b)
