---
# Maps to ContextCore insight.* semantic conventions.

insight:
  id: DEC-039                         # stable, never reused
  type: decision
  confidence: 0.88                    # honest: the LOCAL-day policy is
                                      # well-grounded (mirrors DEC-022's
                                      # shipped local-day streak and the
                                      # storage Since+Until primitive is
                                      # already tested); residual
                                      # uncertainty is only that "the user's
                                      # local day" is read off the host zone,
                                      # which a future non-CLI/multi-host
                                      # caller may need to source explicitly
                                      # (flagged in Validation). Above §14's
                                      # 0.7 line — no new open question filed.
  audience:
    - developer
    - agent

agent:
  id: claude-opus-4-8
  session_id: null

# Decisions are repo-level, but it's useful to track which project
# caused them to be emitted.
project:
  id: PROJ-006
repo:
  id: bragfile

created_at: 2026-07-11
supersedes: null
superseded_by: null

tags:
  - cli
  - list
  - time
  - window
  - timezone
  - local-day
  - clock-seam
---

# DEC-039: `brag list --day` scopes to a single LOCAL calendar day

## Decision

`brag list --day <value>` scopes the listing to exactly ONE **local
calendar day** — the half-open window `[local-midnight, next-local-midnight)`
— by setting BOTH bounds on the existing filter:
`storage.ListFilter{Since: dayStart, Until: nextDayStart}` (the `Until`
field shipped in SPEC-056/DEC-035; this is a CLI wire-up, not new
storage). `value` is one of:

- the keywords `today` / `yesterday` (case-insensitive), resolved
  against the local wall clock; or
- a bare `YYYY-MM-DD` date, taken as that local calendar day.

Any other value is a `UserError` naming the accepted forms. Four points
are locked:

1. **A "day" is a LOCAL calendar day.** `--day 2026-07-05` is that date
   from local midnight to the next local midnight, computed in
   `clock().Location()` (the host zone in production). This is what a
   human means by "today"/"yesterday" and is consistent with DEC-022's
   local-day streak. It **deliberately differs** from bare-date `--since`,
   which stays UTC-midnight (unchanged, for backward compatibility) —
   the two flags are mutually exclusive, so the divergence is never
   observable within one invocation.

2. **The boundary is computed local, compared UTC.** `dayStart` /
   `nextDayStart` are built with `time.Date(..., loc)` and
   `start.AddDate(0, 0, 1)` (calendar arithmetic, DST-correct — never a
   fixed `24h` `Add`). `Store.List` formats them `.UTC().Format(RFC3339)`
   for the `created_at >= ? AND created_at < ?` comparison. Storage stays
   UTC (the blocking `timestamps-in-utc-rfc3339` constraint is untouched);
   only the derived *boundary* localizes at read time — the same
   "derive-local, store-UTC" carve-out DEC-022 established.

3. **Boundaries and keywords resolve through an injectable clock.**
   `today`/`yesterday` and the local zone are read off a package-level
   `var clock = time.Now` (`internal/cli/since.go`), so both are
   deterministically testable across timezones with a fixed instant+zone
   stub. Routing the `--since` *duration* path through the same seam
   removes the inline `time.Now().UTC()` at `since.go:41` — the audit's
   L4 wall-clock impurity. `clock` is intentionally distinct from
   `impact.go`'s `nowFunc`, which returns UTC for the calendar-window
   commands: a "day" is local, a reporting quarter/month/year is
   UTC-anchored there, so the two seams carry different zones on purpose.

4. **`--day` sets the whole window, so it is mutually exclusive with
   `--since`.** Passing both → `UserError` naming the conflict. The
   `--project`/`--type`/`--tag`/`--limit`/`-P`/`--format`/`--author`
   filters still COMPOSE with `--day` (they AND through the same
   `ListFilter`). Output shape (plain TSV / `-P` / `--format json|tsv`)
   is unchanged; only the row set is day-bounded.

## Context

`brag list` had only a lower bound (`--since`) and no clean way to scope
to a single day: "yesterday only" needed a `jq` upper bound (documented
in tutorial §9), and `--since`'s bare-date UTC-midnight anchoring skews
for a non-UTC user (in PDT, "today" via `--since` starts at 5pm the day
before). STAGE-017 opens PROJ-006 with a small read-surface win: make
"what did I do on a given day" a one-liner.

The storage primitive already exists — `ListFilter.Until` (exclusive
upper bound, `created_at < ?`) shipped in DEC-035/SPEC-056 and is tested.
So the only real design fork is **what a "day" means**: local vs UTC.
This DEC settles that fork and folds in the L4 clock-purity fix that
makes keyword resolution testable.

A prior framing attempt sketched a heavier design — a general `--until`
primitive plus `--today`/`--yesterday` boolean flags. STAGE-017 pivoted
to a single `--day` flag; this DEC records why (see Alternatives B and C).

## Alternatives Considered

- **Option A: define a "day" as a UTC calendar day.**
  - What it is: `--day 2026-07-05` → `[2026-07-05T00:00Z, 2026-07-06T00:00Z)`,
    with `today`/`yesterday` off `time.Now().UTC()`. No local zone, no new
    clock seam needed.
  - Why rejected: simpler to implement, but it re-creates the exact skew
    that motivates the feature. For the Pacific user, an evening entry
    lands on what UTC calls the next day, so `--day today` would omit
    entries the user made "today" and `--day yesterday` would misfile
    them. This is the same defect DEC-022 fixed for the streak; shipping a
    day filter that reproduces it would be a regression in intuition. The
    only cost of LOCAL is the injectable-zone seam, which we want anyway
    (L4).

- **Option B: a general `--until <date>` primitive (arbitrary bounded
  ranges), with `--day` layered on top or dropped.**
  - What it is: expose the shipped `ListFilter.Until` directly as
    `--until`, letting users express any `[--since, --until)` window;
    "one day" becomes `--since D --until D+1`.
  - Why rejected: YAGNI. The stated user need is day-scoped retrieval, and
    `--until` makes the common case ("just today") a two-flag,
    off-by-one-prone incantation while asking the user to know the window
    is half-open. `--day` is one concept for the actual need. `--until`
    can be added later, with its own DEC, if arbitrary windows are ever
    genuinely wanted — nothing here forecloses it (the storage field is
    already public).

- **Option C: `--today` / `--yesterday` boolean flags plus a separate
  date flag.**
  - What it is: three flags (`--today`, `--yesterday`, and a bare-date
    window flag) instead of one `--day <value>`.
  - Why rejected: three flags for one concept, with their own mutual-
    exclusion matrix, versus a single `--day` whose value is a small
    closed set. One flag is simpler to document, test, and reason about.

- **Option D (chosen): one `--day <YYYY-MM-DD|today|yesterday>` flag,
  LOCAL calendar day, Since+Until, injectable clock, mutually exclusive
  with `--since`.**
  - Why selected: it delivers the exact need (day-scoped retrieval) in one
    concept, reuses the shipped `Until` storage primitive with zero schema
    change, gives the human-intuitive local-day answer consistent with
    DEC-022, and fixes the L4 clock impurity as a side effect. Minimal
    surface, maximal alignment with existing precedent.

## Consequences

- **Positive:** "what did I do today / yesterday / on date X" is a single
  flag; the tutorial's `jq`-upper-bound workaround is no longer the only
  path to a closed day window. The local-day answer matches user
  intuition and the shipped streak semantics. The `--since` duration path
  is now deterministically testable (L4 closed), and the CLI has a
  local-wall-clock seam future day-oriented surfaces can reuse.
- **Negative:** the `cli` package now carries two clock seams (`clock`
  local, `nowFunc` UTC). Mitigated by documenting each at its declaration
  and in this DEC — they mean genuinely different things. "The user's
  local day" is read off the host zone, so a machine with a wrong
  `TZ`/system zone buckets wrong (acceptable: it is the same host clock
  the user reads, exactly as DEC-022 accepted for the streak).
- **Neutral:** no schema change, no migration — `--day` is pure query-time
  filtering over the existing `Since`/`Until` fields. `--day` and `--since`
  diverge on bare-date semantics (local vs UTC midnight), but mutual
  exclusion means no single invocation exposes the difference.

## Validation

Right if:
- `--day <date>` returns exactly that local day's entries: a 23:30-local
  entry is IN, a 00:30-next-day-local entry is OUT (local bounds +
  exclusive upper edge).
- `--day today` / `--day yesterday` with a stubbed evening-PDT clock place
  a "yesterday-local but today-UTC" entry in `--day yesterday`, NOT
  `--day today` (the skew this fixes).
- `--day <garbage>` and `--day --since ...` are both `UserError` (exit 1,
  empty stdout, message on stderr); the other filters compose with `--day`.
- Plain `list` and existing `--since` behavior are unchanged (green
  regression tests).

Revisit if:
- A non-CLI or multi-host caller (e.g. an MCP `brag_list` day filter, or a
  synced corpus) cannot meaningfully source "the user's local day" from
  the host zone — then the *zone source* (not the local-day policy) needs
  an explicit input, exactly as DEC-022 flagged for its own `now`.
- Arbitrary bounded ranges become a real need — add `--until` then, with
  its own DEC; `--day` stays the ergonomic shortcut.

Confidence: 0.88. The LOCAL-day policy mirrors the shipped DEC-022 streak
and rides an already-tested storage primitive, so little here is novel;
held below higher only because "local = host zone" is an assumption a
future non-CLI caller may need to revisit. Above §14's 0.7 line, so no new
open question is filed.

## References

- Related specs:
  - SPEC-068 (emits this DEC; adds `brag list --day`, the `ParseDay`
    helper, the `clock` seam / L4 fix).
  - SPEC-056 (shipped) — DEC-035; added `ListFilter.Until`, the exclusive
    upper bound `--day` rides on.
  - SPEC-038 (shipped) — DEC-022; the local-day streak precedent this
    reuses.
  - SPEC-007 (shipped) — introduced `ListFilter` + the `Since` lower bound.
- Related decisions:
  - DEC-035 (`ListFilter.Until`) — the storage primitive; `--day` sets its
    `Until` to the next local midnight.
  - DEC-022 (local-day streak) — the "derive-local, store-UTC" carve-out
    and the "location rides on the injected now" clock discipline this
    follows.
  - DEC-008 (`--since` parsing) — the bare-date UTC-midnight semantics
    `--day` deliberately diverges from (and which stay unchanged).
- Related constraints: `timestamps-in-utc-rfc3339` (blocking — untouched;
  storage stays UTC, only the derived boundary localizes), `no-sql-in-cli-layer`
  (the window is expressed via `ListFilter`, no SQL in the CLI),
  `stdout-is-for-data-stderr-is-for-humans`, `errors-wrap-with-context`,
  `test-before-implementation`.
- Discussions: STAGE-017 Design Notes (the one-flag pivot; the LOCAL-vs-UTC
  fork this DEC settles).
