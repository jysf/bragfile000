---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-038
  type: bug                        # epic | story | task | bug | chore
  cycle: design                    # frame | design | build | verify | ship
  blocked: false
  priority: high
  complexity: S                    # S | M | L  (L means split it)

project:
  id: PROJ-003
  stage: STAGE-009
repo:
  id: bragfile

agents:
  architect: claude-opus-4-8
  implementer: claude-opus-4-8     # claude-only variant: same model, separate session
  created_at: 2026-07-03

references:
  decisions: [DEC-022, DEC-011]
  constraints: [timestamps-in-utc-rfc3339, test-before-implementation, no-cgo, errors-wrap-with-context]
  related_specs: [SPEC-020, SPEC-039]
---

# SPEC-038: current-streak fix — local day + alive through yesterday

## Context

`brag stats` current-streak is a **confirmed defect** (backlog.md "BUG:
`brag stats` current-streak reads 0 until you log today", 2026-06-20): an
unbroken 14-day run showed `Current: 0` for the whole part of a day before
the day's first re-log. Root cause in `aggregate.Streak`
(`internal/aggregate/aggregate.go:172`): the cursor seeds at *today (UTC)*
and breaks immediately if today has no entry, and entries bucket by
`CreatedAt.UTC()` so an evening-Pacific entry lands on what UTC calls the
next day.

This spec is **SPEC-038**, the first spec of **STAGE-009** (PROJ-003's
committed v0.3.0 core) and its lead-off item. It **blocks SPEC-039**
(milestone notifications): the streak milestone must fire on a correct
number, so the correctness fix lands first.

The design carve-out — localize a *derived metric* while storage stays UTC
RFC3339 — is settled in **DEC-022** (emitted by this spec). It revises the
UTC-day + requires-today semantics that **SPEC-020 §6** locked inline.

## Goal

Make `aggregate.Streak`'s **current** streak count in the user's **local**
calendar day and stay **alive through yesterday** — reading an intact run's
true length all day, not 0-until-relog — while storage timestamps remain
UTC RFC3339 (only the derived metric localizes) and all day arithmetic
stays DST-correct.

## Inputs

- **Files to read:**
  - `internal/aggregate/aggregate.go` (`Streak`, lines 166–210) — the
    function to change; note the current UTC bucketing + today-seeded cursor
    + `Sub == 24*time.Hour` longest check.
  - `internal/aggregate/aggregate_test.go` (`TestStreak_CurrentAndLongest`,
    lines 294–368; `entryAt` helper, 285–292) — the test to rewrite.
  - `internal/export/stats.go` (lines 34, 43, 127) — passes `opts.Now`
    through to `Streak`; the `Generated:` line re-`.UTC()`s explicitly.
  - `internal/cli/stats.go` (lines 21, 59–62) — the caller that currently
    pre-strips the clock to `time.Now().UTC()` and the "single Now source"
    comment.
  - `decisions/DEC-022-streak-local-day-derived-metric.md` — the governing
    decision.
- **External APIs:** none.
- **Related code paths:** `internal/aggregate/`, `internal/export/`,
  `internal/cli/`.

## Outputs

- **Files modified:**
  - `internal/aggregate/aggregate.go` — rewrite `Streak` body: bucket by
    `now.Location()`; seed current-cursor at today, else yesterday; step via
    `AddDate`; compute longest-run adjacency via calendar `AddDate`
    (date-label compare), not `Sub == 24h`. Rewrite the doc comment
    (166–171) — it currently claims "UTC calendar days" and "today's UTC
    date has zero entries, current = 0", both now false.
  - `internal/aggregate/aggregate_test.go` — rewrite
    `TestStreak_CurrentAndLongest` (premise shift, see audit below); add
    `TestStreak_BucketsByLocalDay` and `TestStreak_CurrentStepsAcrossDSTBoundary`;
    add `import _ "time/tzdata"` (stdlib — embeds the IANA DB in the *test*
    binary only, keeping `LoadLocation` hermetic in CI; see Locked decisions).
  - `internal/cli/stats.go` — line 62: `time.Now().UTC()` → `time.Now()`
    so a real local zone reaches the metric (the "single Now source"
    decision holds — still one call; the `Generated:` line re-`.UTC()`s it).
    Line 21 `Long`: "consecutive UTC days with entries" → wording that
    matches the new local-day / alive-through-yesterday semantics.
  - `docs/api-contract.md` — line 363: "consecutive UTC days with entries"
    → corrected streak description.
- **New exports:** none (signature `Streak(entries, now) (current, longest)`
  is unchanged — the local zone rides on `now.Location()`, per DEC-022
  Option D; the `loc`-parameter alternative was rejected).
- **Database changes:** none. **No migration.** Storage stays UTC RFC3339;
  the `timestamps-in-utc-rfc3339` blocking constraint is untouched (DEC-022
  Context records why localizing a derived metric does not relax it).

### Premise audit (run at design, per §9 — enumerate, don't discover at build)

**Inversion/removal → planned test rewrite.** `TestStreak_CurrentAndLongest`
(`aggregate_test.go:294`) encodes the *old* premise (SPEC-020 §6b:
"today-without-entries breaks the streak immediately"). Two subtests
**invert/change** and are enumerated here as planned rewrites, not
build-time discoveries:

| Subtest (old) | Old expect | New expect | Why it changes |
|---|---|---|---|
| `today_zero_entries_yields_zero` (entries 4/22–4/24, now 4/25) | `(0,3)` | `(3,3)` | today empty but **yesterday 4/24 present** → streak alive; renamed `today_empty_yesterday_alive` |
| `gap_mid_corpus_longest` (…4/23,4/24, now 4/25) | `(0,5)` | `(2,5)` | today empty, yesterday 4/24 present → current = 2 (4/24,4/23), not 0 |

The other four existing subtests (`today_has_entries`, `single_entry`,
`multiple_entries_same_day`, `empty_corpus`) are **preserved** — they lock
behavior the fix intentionally keeps, and will *pass on the old code* (they
are regression guards, not fail-first cases; see Failing Tests note on which
subtests must fail first). `single_entry` is renamed `single_entry_today`
for symmetry with the new `single_entry_yesterday`.

**Status-change (doc/help) → planned doc-reference updates.** Grep run at
design: `grep -rin "consecutive.*day\|UTC day\|streak" internal/ docs/
README.md BRAG.md`. Each hit and its verdict:

| Hit | Verdict |
|---|---|
| `internal/aggregate/aggregate.go:166–171` (doc comment) | **fix now** — load-bearing, in the file being changed; claims "UTC calendar days" + "current = 0". |
| `internal/cli/stats.go:21` (`Long`: "consecutive UTC days with entries") | **fix now** — user-facing help, now factually wrong. No test asserts this substring (grep confirms), so no test churn. |
| `docs/api-contract.md:363` ("consecutive UTC days with entries") | **fix now** — factually wrong after the change. |
| `docs/tutorial.md:483` ("current and longest streak") | **stays** — generic, no UTC/semantics claim. |
| `docs/architecture.md:40,84`, `docs/blog/how-brag-was-built.md:123` | **stays** — merely list `Streak` as a helper name; no semantics claim. |
| `internal/cli/summary.go`, `review.go`, `docs/api-contract.md` "last 7/30 UTC days" | **stays** — these describe the `summary`/`review` *rolling window*, not the streak; out of scope. |

**Additive/count-bump.** None — the streak is not a counted fixed-shape
collection; no literal-count assertion depends on it.

**Existing stats goldens — audited, unaffected (no change).**
`internal/export/stats_test.go` (golden asserts `Current: 2`, `longest 3`)
and `internal/cli/stats_test.go` are **confirmed green** under the change:
the export fixture puts `now` in `time.UTC` with a *today* entry (4/25) —
the backward-compatible "today has entries" path — so local==UTC and the
alive-through-yesterday branch is not taken; the CLI test asserts JSON-key
presence, not streak *values*, and does not inject `now`. Enumerated so
build does not mistake their green state for a missed rewrite.

## Acceptance Criteria

- [ ] `Streak` current-streak counts the consecutive local-day run ending
      on **today or yesterday**; equals the run length when the last entry
      is yesterday and `now` is today (the bug's core case), not 0.
- [ ] Current-streak is **0** only when *both* today and yesterday
      (local) are empty (a genuinely broken streak still reads 0 — the fix
      grants one day of grace, not immortality).
- [ ] Entries bucket by **local** day (`now.Location()`); a fixture whose
      UTC date differs from its local date buckets by local date, changing
      the answer vs the old UTC bucketing.
- [ ] All day arithmetic is **DST-correct**: a current-streak run that
      steps across a spring-forward boundary counts every local day.
- [ ] `longest` is unchanged in contract (longest consecutive local-day
      run anywhere; same-day entries dedupe to one day) and computed via
      calendar adjacency, not instant subtraction.
- [ ] Storage untouched — no migration; `timestamps-in-utc-rfc3339` holds.
- [ ] `go test ./...`, `gofmt -l .`, `go vet ./...` clean;
      `CGO_ENABLED=0 go build ./...` clean.

## Failing Tests

Written during **design**, BEFORE build. Build makes these pass by
rewriting `Streak`. Expected values were computed against the fixtures at
design; the timezone offsets were verified against real `tzdata` (LA is
UTC-7/PDT in late April and after Mar 8, 2026; UTC-8/PST before) — do not
hand-retype them.

**Fail-first map (per §9 "confirm tests fail for the expected reason"):**
the subtests/tests marked ▲ assert *new* behavior and MUST fail on the
current code; those marked ● are *preservation* guards and will pass on the
current code (that is expected — they lock behavior the fix keeps).

- **`internal/aggregate/aggregate_test.go`** — replace
  `TestStreak_CurrentAndLongest` wholesale with the version below (reuses
  the existing `entryAt(year,month,day,hour)` UTC helper, lines 285–292):

  ```go
  // TestStreak_CurrentAndLongest locks the DEC-022 streak semantics:
  // current counts the consecutive local-day run ending on TODAY or
  // YESTERDAY (alive-through-yesterday), 0 only after two empty days;
  // longest is the longest consecutive local-day run; same-day entries
  // dedupe to one day. now is injected (instant + zone); no time.Sleep.
  // These subtests keep now in time.UTC, so local day == UTC day — the
  // alive-through-yesterday axis is exercised here; the local-day and DST
  // axes are exercised by the two dedicated tests below.
  func TestStreak_CurrentAndLongest(t *testing.T) {
  	now := time.Date(2026, 4, 25, 12, 0, 0, 0, time.UTC)

  	// ● preservation: today has an entry → count today and walk back.
  	t.Run("today_has_entries", func(t *testing.T) {
  		entries := []storage.Entry{
  			entryAt(2026, 4, 23, 10), entryAt(2026, 4, 24, 10), entryAt(2026, 4, 25, 10),
  		}
  		current, longest := Streak(entries, now)
  		if current != 3 || longest != 3 {
  			t.Errorf("today_has_entries: got (%d,%d), want (3,3)", current, longest)
  		}
  	})

  	// ▲ new: today empty, yesterday (4/24) present → streak ALIVE (was 0).
  	t.Run("today_empty_yesterday_alive", func(t *testing.T) {
  		entries := []storage.Entry{
  			entryAt(2026, 4, 22, 10), entryAt(2026, 4, 23, 10), entryAt(2026, 4, 24, 10),
  		}
  		current, longest := Streak(entries, now)
  		if current != 3 || longest != 3 {
  			t.Errorf("today_empty_yesterday_alive: got (%d,%d), want (3,3)", current, longest)
  		}
  	})

  	// ▲ new (the BUG's canonical case): run ending yesterday, now=today,
  	// current == run length (not 0).
  	t.Run("run_ending_yesterday_now_today", func(t *testing.T) {
  		entries := []storage.Entry{entryAt(2026, 4, 23, 10), entryAt(2026, 4, 24, 10)}
  		current, longest := Streak(entries, now)
  		if current != 2 || longest != 2 {
  			t.Errorf("run_ending_yesterday_now_today: got (%d,%d), want (2,2)", current, longest)
  		}
  	})

  	// ▲ new: a single entry dated yesterday still reads current 1.
  	t.Run("single_entry_yesterday", func(t *testing.T) {
  		entries := []storage.Entry{entryAt(2026, 4, 24, 10)}
  		current, longest := Streak(entries, now)
  		if current != 1 || longest != 1 {
  			t.Errorf("single_entry_yesterday: got (%d,%d), want (1,1)", current, longest)
  		}
  	})

  	// ● preservation: two empty days (4/24 AND 4/25) → current 0. The fix
  	// grants one day of grace, not immortality.
  	t.Run("streak_dead_after_two_empty_days", func(t *testing.T) {
  		entries := []storage.Entry{
  			entryAt(2026, 4, 21, 10), entryAt(2026, 4, 22, 10), entryAt(2026, 4, 23, 10),
  		}
  		current, longest := Streak(entries, now)
  		if current != 0 || longest != 3 {
  			t.Errorf("streak_dead_after_two_empty_days: got (%d,%d), want (0,3)", current, longest)
  		}
  	})

  	// ▲ changed: today empty, yesterday 4/24 present → current 2 (was 0);
  	// longest preserved at 5 (the 4/10–4/14 run).
  	t.Run("gap_mid_corpus_longest", func(t *testing.T) {
  		entries := []storage.Entry{
  			entryAt(2026, 4, 10, 10), entryAt(2026, 4, 11, 10), entryAt(2026, 4, 12, 10),
  			entryAt(2026, 4, 13, 10), entryAt(2026, 4, 14, 10),
  			entryAt(2026, 4, 23, 10), entryAt(2026, 4, 24, 10),
  		}
  		current, longest := Streak(entries, now)
  		if current != 2 || longest != 5 {
  			t.Errorf("gap_mid_corpus_longest: got (%d,%d), want (2,5)", current, longest)
  		}
  	})

  	// ● preservation.
  	t.Run("single_entry_today", func(t *testing.T) {
  		entries := []storage.Entry{entryAt(2026, 4, 25, 10)}
  		current, longest := Streak(entries, now)
  		if current != 1 || longest != 1 {
  			t.Errorf("single_entry_today: got (%d,%d), want (1,1)", current, longest)
  		}
  	})

  	// ● preservation: same-day multiple entries dedupe to one streak day.
  	t.Run("multiple_entries_same_day", func(t *testing.T) {
  		entries := []storage.Entry{
  			entryAt(2026, 4, 25, 8), entryAt(2026, 4, 25, 12), entryAt(2026, 4, 25, 16),
  		}
  		current, longest := Streak(entries, now)
  		if current != 1 || longest != 1 {
  			t.Errorf("multiple_entries_same_day: got (%d,%d), want (1,1)", current, longest)
  		}
  	})

  	// ● preservation.
  	t.Run("empty_corpus", func(t *testing.T) {
  		current, longest := Streak(nil, now)
  		if current != 0 || longest != 0 {
  			t.Errorf("empty_corpus: got (%d,%d), want (0,0)", current, longest)
  		}
  	})
  }
  ```

- **`internal/aggregate/aggregate_test.go`** — add (requires
  `import _ "time/tzdata"` in the import block):

  ```go
  // TestStreak_BucketsByLocalDay ▲ proves current-streak buckets by the
  // user's LOCAL day, not UTC. Two entries whose UTC date is 2026-04-25
  // (05:00Z and 16:00Z) fall on DIFFERENT local days in America/Los_Angeles
  // (04-24 22:00 and 04-25 09:00). With now on the local evening of 04-25,
  // the local streak is 2. The OLD code buckets both entries under UTC
  // 04-25 (one day) while seeding its cursor from now's LOCAL day (04-25,
  // since it reads now.Day() directly), so it returns (1,1) — this test
  // fails on old (1,1 != 2,2), passes on new. Offsets verified against
  // tzdata; old/new values confirmed against a reference impl at design.
  func TestStreak_BucketsByLocalDay(t *testing.T) {
  	la, err := time.LoadLocation("America/Los_Angeles")
  	if err != nil {
  		t.Fatalf("load America/Los_Angeles: %v", err)
  	}
  	entries := []storage.Entry{
  		entryAt(2026, 4, 25, 5),  // UTC 05:00Z == 2026-04-24 22:00 PDT
  		entryAt(2026, 4, 25, 16), // UTC 16:00Z == 2026-04-25 09:00 PDT
  	}
  	now := time.Date(2026, 4, 25, 20, 0, 0, 0, la) // 2026-04-25 evening local
  	current, longest := Streak(entries, now)
  	if current != 2 || longest != 2 {
  		t.Errorf("BucketsByLocalDay: got (%d,%d), want (2,2)", current, longest)
  	}
  }

  // TestStreak_CurrentStepsAcrossDSTBoundary ● GUARD (not fail-first):
  // locks that the NEW code's LOCAL cursor steps by CALENDAR day (AddDate),
  // not by 24h, across the 2026 US spring-forward (Mar 8, a 23-hour local
  // day). A run on local days Mar 7/8/9 with now on Mar 9 must count 3.
  // This fixture's UTC and local dates coincide, so the OLD code also
  // returns (3,3) — the test does NOT fail on current code (confirmed
  // against a reference impl). Its job is to pair with the calendar-
  // arithmetic decision: if the NEW code stepped with cursor.Add(-24h)
  // instead of AddDate, Mar 9 00:00 local minus 24h lands on Mar 7 23:00
  // (date Mar 7), SKIPPING Mar 8 → current would be 2, and this test would
  // fail. So it guards against a DST-unsafe reimplementation of the walk.
  func TestStreak_CurrentStepsAcrossDSTBoundary(t *testing.T) {
  	la, err := time.LoadLocation("America/Los_Angeles")
  	if err != nil {
  		t.Fatalf("load America/Los_Angeles: %v", err)
  	}
  	entries := []storage.Entry{
  		entryAt(2026, 3, 7, 20), // local noon PST (UTC-8) == 20:00Z, local 03-07
  		entryAt(2026, 3, 8, 19), // local noon PDT (UTC-7) == 19:00Z, local 03-08
  		entryAt(2026, 3, 9, 19), // local noon PDT == 19:00Z, local 03-09
  	}
  	now := time.Date(2026, 3, 9, 18, 0, 0, 0, la) // 2026-03-09 evening local
  	current, longest := Streak(entries, now)
  	if current != 3 || longest != 3 {
  		t.Errorf("CurrentStepsAcrossDSTBoundary: got (%d,%d), want (3,3)", current, longest)
  	}
  }
  ```

## Implementation Context

*Read this section (and the files it points to) before starting the build
cycle. It is the equivalent of a handoff document, folded into the spec.*

### Decisions that apply

- `DEC-022` — **the governing decision.** Current streak is a local-day
  derived metric, alive through yesterday; local zone rides on
  `now.Location()`; storage stays UTC; calendar arithmetic throughout for
  DST-safety. This spec emits it. Read it in full before build.
- `DEC-011` — stored entry shape / UTC RFC3339 timestamps; the streak
  reads `CreatedAt` (a UTC instant) and converts to local only for the
  derived value.

### Constraints that apply

- `timestamps-in-utc-rfc3339` (blocking) — **untouched.** Governs
  `internal/storage/**` writes; nothing here writes a timestamp. DEC-022
  records why a read-time local conversion of a derived metric does not
  relax it. Do NOT change any storage write path.
- `test-before-implementation` (blocking) — the tests above are written
  first; build makes them pass.
- `no-cgo` (blocking) — pure-Go only; `time`/`time/tzdata` are stdlib.
- `errors-wrap-with-context` (warning) — `Streak` returns no error; N/A to
  the body, but keep any touched CLI error paths wrapped.

### Prior related work

- `SPEC-020` (shipped) — introduced `Streak`/`MostCommon`/`Span` and locked
  the UTC-day + requires-today streak semantics (spec §6) this revises.
- `SPEC-039` (next) — milestone notifications; **blocked by this spec**; it
  reads the corrected current-streak for the streak milestone.

### Out of scope (for this spec specifically)

- Milestone notifications (SPEC-039).
- Any change to `summary`/`review` rolling-window ("last 7/30 UTC days")
  semantics — different metric, out of scope.
- First-class timezone config / a `--tz` flag / per-entry zone — the zone
  is the host zone via `now.Location()`; a configurable zone is a
  future-if-earned item (DEC-022 Validation).
- Promoting the "derive-local, store-UTC" rule into a new *constraint* —
  DEC-022 is the durable home for now; codify only if a third derived-metric
  consumer re-derives it (STAGE-010 would be the second).

## Notes for the Implementer

- **The rewrite shape (guidance, not locked line-by-line):** bucket
  `e.CreatedAt.In(now.Location())`; build `today := time.Date(now.Year(),
  now.Month(), now.Day(), 0,0,0,0, now.Location())`; seed `cursor := today`,
  and if today's date-label is absent, `cursor = cursor.AddDate(0,0,-1)`
  once (the yesterday grace); then walk back counting present days with
  `AddDate(0,0,-1)`. For `longest`, keep the sorted-date-label sweep but
  replace `curr.Sub(prev) == 24*time.Hour` with a calendar-adjacency check
  (e.g. parse each label and compare `prev.AddDate(0,0,1)`'s label to the
  next) so it is DST-immune. `now.Year()/Month()/Day()` already return
  **local** calendar fields when `now` carries a location — verified.
- **Why `time/tzdata` in the test:** `LoadLocation("America/Los_Angeles")`
  otherwise depends on the runner shipping the system zoneinfo. The blank
  import embeds the IANA DB into the *test* binary only (not the shipped
  `brag`), keeping CI hermetic. It is stdlib — **not** a go.mod dependency,
  so `no-new-top-level-deps-without-decision` does not fire and no DEC is
  needed for it. Note this in Build Completion so verify doesn't flag it.
- **The `time.Now().UTC()` → `time.Now()` caller change is load-bearing.**
  Without it the aggregate fix is inert in production (a UTC-located `now`
  buckets in UTC regardless of the new code). The "single Now source"
  decision (SPEC-020 §10) survives: still one `time.Now()` call, and the
  `Generated:` line at `stats.go` (markdown line 34 / the JSON renderer)
  re-`.UTC()`s it, so the timestamp display is unchanged.
- **Do not add a `time.Sleep` anywhere** — every day-boundary case uses the
  injected `now` (§9).
- **Fail-first check (build step):** run `go test ./internal/aggregate/` and
  confirm the ▲ subtests/tests fail on the *assertion* (wrong streak count),
  not on a compile error, before touching `Streak`. The ● preservation cases
  passing at that point is expected.

## Locked design decisions

**Behavior decisions (1–4) — each has ≥1 paired test that fails if the
decision is not implemented (per §9).** All expected values, and each
test's fail-first-vs-current-code status, were confirmed at design against
a reference implementation of both the old and new `Streak`.

1. **Current streak is alive through yesterday** (DEC-022). Current = the
   consecutive local-day run ending on today *or* yesterday. *Pair
   (fail-first ▲):* `today_empty_yesterday_alive`,
   `run_ending_yesterday_now_today`, `single_entry_yesterday`.
   - **Rejected alternative:** keep requires-today (SPEC-020 §6b). Rejected
     — it *is* the bug.
2. **Current streak is 0 only after two empty days.** The grace is exactly
   one day. *Pair (preservation ● — passes on old too):*
   `streak_dead_after_two_empty_days` (0 with a 3-run two days back).
3. **Bucket by local day, via `now.Location()`** (DEC-022 Option D). Signature
   `Streak(entries, now)` unchanged; the zone rides on the injected `now`.
   *Pair (fail-first ▲):* `TestStreak_BucketsByLocalDay` (old (1,1) vs
   new (2,2)).
   - **Rejected alternatives:** (a) explicit `loc` parameter — two sources
     of truth for the day boundary + signature/call-site/test churn for no
     gain; (b) read `time.Local` inside `Streak` — breaks the §9 injectable
     seam. Both rejected in DEC-022 Alternatives.
4. **All day arithmetic is calendar-based (`AddDate` + date-label compare),
   never instant subtraction** — DST-safe. Replaces the `Sub == 24*time.Hour`
   longest check and forbids `cursor.Add(-24h)` stepping. *Pair (guard ●):*
   `TestStreak_CurrentStepsAcrossDSTBoundary` — it does **not** fail on the
   current code (that fixture's UTC and local dates coincide, so old also
   returns (3,3)), but it **would fail** if the new walk stepped with
   `Add(-24h)` (→ (2,3) by skipping the 23-hour Mar-8 day). So it pairs with
   this decision as the "fails if implemented wrong" guard §9 asks for, and
   locks the shipped new-code DST behavior against future refactors.

**Constraints preserved (validated by the premise audit / by absence, not
by a dedicated failing test — no new behavior to assert):**

- **Storage stays UTC RFC3339; only the derived metric localizes**
  (DEC-022; blocking `timestamps-in-utc-rfc3339` untouched). No migration,
  no storage-write change. Validated by the "no DB changes" Output and the
  full suite continuing to store/read UTC unchanged. This is a *don't-touch*
  guarantee — a failing test that "storage isn't UTC" would be proving a
  negative; the audit + no-migration are the right validation.
- **Caller passes `time.Now()` (local), not `time.Now().UTC()`**
  (`internal/cli/stats.go:62`) so a real zone reaches the metric. This is
  mechanical wiring, not a design choice with alternatives (the aggregate
  can only see what `now` carries), so it carries no dedicated failing test;
  it is verified by inspection. Safety net: the existing
  `internal/export/stats_test.go` golden stays green, proving the
  `Generated:` line (which re-`.UTC()`s `now`) is unaffected by a local
  `now`, and the "single Now source" decision (SPEC-020 §10) is preserved
  (still one `time.Now()` call).

**Doc-text corrections (prose, not behavior — no paired test):** the
`aggregate.go` doc comment, `stats.go` `Long`, and `docs/api-contract.md:363`
drop "UTC calendar days"/"consecutive UTC days" for the local-day /
alive-through-yesterday wording. No test asserts these substrings
(grep-confirmed), so none is added; they are enumerated in the premise
audit as required updates.

---

## Build Completion

*Filled in at the end of the **build** cycle, before advancing to verify.*

- **Branch:** feat/spec-038-streak-fix
- **PR (if applicable):**
- **All acceptance criteria met?** <yes/no>
- **New decisions emitted:**
  - `DEC-022` — streak local-day derived metric (emitted at design)
- **Deviations from spec:**
  - [list]
- **Follow-up work identified:**
  - [any new specs for the stage's backlog]

### Build-phase reflection (3 questions, short answers)

1. **What was unclear in the spec that slowed you down?**
   — <answer>

2. **Was there a constraint or decision that should have been listed but wasn't?**
   — <answer>

3. **If you did this task again, what would you do differently?**
   — <answer>

---

## Reflection (Ship)

*Appended during the **ship** cycle.*

1. **What would I do differently next time?**
   — <answer>

2. **Does any template, constraint, or decision need updating?**
   — <answer>

3. **Is there a follow-up spec I should write now before I forget?**
   — <answer>
