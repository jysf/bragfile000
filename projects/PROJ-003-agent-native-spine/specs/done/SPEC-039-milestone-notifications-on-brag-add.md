---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-039
  type: story                      # epic | story | task | bug | chore
  cycle: ship
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
  decisions: [DEC-023, DEC-022, DEC-011, DEC-012]
  constraints: [stdout-is-for-data-stderr-is-for-humans, test-before-implementation, no-cgo, no-new-top-level-deps-without-decision, errors-wrap-with-context]
  related_specs: [SPEC-038, SPEC-020, SPEC-017, SPEC-019]
---

# SPEC-039: Milestone notifications on `brag add`

## Context

Capture should feel good. Today `brag add` is silent on success beyond the
inserted ID — a correct, pipeable contract, but a joyless one. This spec
adds a single **celebratory, TTY-only, stderr** line when an `add` crosses
a milestone: a lifetime total (10/25/50/100/250/500/1000), a current-streak
day count (7/30/100), or a per-project count (10th/50th), plus a quiet
"first brag today / this week" nudge. It fires on an action the user (or an
agent) already takes — no new command to remember — which is why the
2026-06-16 brainstorm tagged milestone notifications the best
effort-to-payoff pick in the backlog.

This is **SPEC-039**, the second spec of **STAGE-009** (PROJ-003's v0.3.0
core). It is **blocked by SPEC-038** (shipped 2026-07-03, PR #57): the
streak milestone must fire on a *correct* number, and SPEC-038 corrected
`aggregate.Streak` to the alive-through-yesterday / local-day semantics
(DEC-022). That corrected metric is on `main`; this spec consumes it.

The design tension the whole spec turns on is the repo's spine:
**`stdout-is-for-data-stderr-is-for-humans`** (blocking). The milestone is
human chatter — it must never touch stdout (which carries the ID for
`id=$(brag add ...)`) and must stay silent on the machine-facing paths
(`--json`, pipes, non-TTY), so scripted and agent-driven capture stays
byte-clean. STAGE-009's Design Notes name this line the CLI-side mirror of
the "stdout/stderr spine at a new transport" that SPEC-040's MCP server
faces; the two specs test the same spine at two surfaces.

## Goal

On a successful `brag add` in flag mode or editor mode, when stderr is a
terminal, print exactly one celebratory line to **stderr** if the add
crossed a total / streak / per-project milestone threshold (else a quiet
first-brag-today/this-week nudge); emit **nothing** under `--json`, on any
non-TTY/piped invocation, or when no threshold is crossed — and never touch
stdout. The milestone is best-effort: computing it must never fail the add.

## Inputs

- **Files to read:**
  - `internal/cli/add.go` — the command being hooked. `runAddFlags`
    (132–165) and `runAddEditor` (167–212) are the two human paths that
    gain a milestone call after the ID is printed; `runAdd` dispatch
    (111–130) shows why `--json` routes away before any milestone runs.
  - `internal/cli/add_json.go` — `runAddJSON` (99–124): the machine path
    that stays silent (never calls the milestone hook).
  - `internal/aggregate/aggregate.go` — `Streak(entries, now) (current,
    longest)` (166–228), corrected by SPEC-038/DEC-022; the milestone reads
    `current`. Note the local-day bucketing rides on `now.Location()`.
  - `internal/cli/stats.go` (59–64) — the single-`time.Now()`-source pattern
    and DEC-022 comment; the milestone clock follows the same "one local
    `time.Now()` reaches the metric" rule.
  - `internal/cli/add_test.go` (1–26, 231–247, 543–560) — the
    `newRootWithAdd` harness, the split-buffer `outBuf`/`errBuf` shape, and
    the `installAddEditFunc` editor seam.
  - `internal/storage/storagetest/storagetest.go` — `Backdate(dbPath, id,
    at)`: seeds past-dated rows out-of-band (Store.Add always stamps
    `time.Now()`); used by the end-to-end streak test.
  - `internal/storage/store.go` — `Add` (131–171), `List(ListFilter{})`
    (303); `internal/storage/entry.go` — `Entry`, `ListFilter`.
  - `decisions/DEC-023-milestone-notifications-copy-and-semantics.md` — the
    governing decision (emitted by this spec).
  - `decisions/DEC-022-...` — the corrected streak this consumes.
- **External APIs:** none.
- **Related code paths:** `internal/cli/`, `internal/aggregate/`.

## Outputs

- **Files created:**
  - `internal/cli/milestone.go` — the milestone module (below):
    - The pure decision function `milestoneLine(milestoneInputs) string`
      and its `milestoneInputs` struct (no Store, no clock, no TTY — the
      whole threshold/precedence/copy matrix is unit-testable here).
    - The threshold sets `totalThresholds`, `streakThresholds`,
      `projectThresholds` and the `crossed(before, after, thresholds)`
      helper.
    - The CLI glue `emitMilestone(cmd, s, inserted)` +
      `computeMilestoneInputs(s, inserted, now)` and the two injectable
      seams `addClock` (`= time.Now`) and `addStderrIsTTY`
      (`= defaultStderrIsTTY`, a stdlib `os.ModeCharDevice` probe of the
      real `os.Stderr`).
  - `internal/cli/milestone_test.go` — pure-function tests (no DB).
  - `internal/cli/add_milestone_test.go` — integration/gating tests
    (DB + injected seams).
  - `decisions/DEC-023-milestone-notifications-copy-and-semantics.md`
    (emitted at design).
- **Files modified:**
  - `internal/cli/add.go` — call `emitMilestone(cmd, s, inserted)` after the
    `fmt.Fprintln(cmd.OutOrStdout(), inserted.ID)` line in **both**
    `runAddFlags` and `runAddEditor`. Add a short, generic milestone note to
    the cobra `Long` (see Notes — must NOT contain the literal celebratory
    strings, per the NOT-contains self-audit). Add the `time`, `os`, and
    `internal/aggregate` imports as needed (via milestone.go; add.go itself
    only gains the `emitMilestone` call).
  - `internal/cli/add_test.go` — modify the shared `newRootWithAdd` helper
    to pin `addStderrIsTTY = func() bool { return false }` with `t.Cleanup`
    restoring `defaultStderrIsTTY` (hermetic default — see premise audit);
    add a `setStderrIsTTY(t, bool)` helper for the milestone tests to opt
    in. Add `TestAdd_HelpMentionsMilestone` (help-mentions test, mirrors
    `TestAdd_HelpMentionsAutoFill`).
  - `docs/api-contract.md` — the `brag add` section (esp. line 71 "stderr
    empty" for editor mode, and the implicit flag-mode "stderr empty on
    success") gains the milestone-on-stderr contract (see premise audit;
    the `id=$(brag add ...)` stdout contract is unaffected).
- **New exports:** none exported outside `package cli` (all
  lower-case/package-private: `milestoneLine`, `milestoneInputs`,
  `emitMilestone`, `computeMilestoneInputs`, `crossed`, the threshold
  slices, `addClock`, `addStderrIsTTY`, `defaultStderrIsTTY`).
- **Database changes:** none. **No migration.** No new go.mod dependency
  (see Locked decision 6: stdlib `os.ModeCharDevice`, not `go-isatty`).

### Premise audit (run at design, per §9 — enumerate, don't discover at build)

**Inversion/removal → planned doc-reference update + test-harness guard.**
The milestone line is *new* stderr output on a success path that the
contract currently documents as silent. Two premises invert:

| Hit | Verdict |
|---|---|
| `docs/api-contract.md:71` ("stderr empty" for editor-mode success) | **fix now** — now conditionally false: on a TTY, a milestone/nudge may print to stderr. Reword to the TTY-gated contract; keep the stdout ID contract intact. |
| `docs/api-contract.md` flag-mode block (45–48) — implicitly "stdout = ID, nothing else" | **augment** — add one line: on a TTY, `add` may print a milestone line to **stderr**; stdout stays the ID alone, so `id=$(brag add ...)` is unaffected; `--json`/piped/non-TTY runs stay silent. |
| Existing add success tests asserting `errBuf.Len()==0` (e.g. `TestAdd_SuccessPrintsIDToStdoutOnly`, `TestAdd_OutputIsPipeable`, all `newRootWithAdd` flag/editor callers) | **guard, not rewrite** — their premise ("stderr empty on success") is *preserved*, not inverted, because the milestone is TTY-gated. But whether `go test` runs under a terminal is environment-dependent, so `defaultStderrIsTTY()` could return `true` and fire a nudge during them. Pin `addStderrIsTTY=false` in the shared `newRootWithAdd` helper so every existing assertion stays hermetic regardless of runner. This is a harness change, not an assertion change. |

**Status-change (doc).** Grep run at design:
`grep -rin "stderr empty\|stderr is empty\|nothing.*stderr\|milestone" docs/ README.md BRAG.md internal/cli`.
Verdicts: `docs/api-contract.md:71` and the flag-mode block → fixed above;
`docs/api-contract.md:184-185` (`No changes.`/`Updated.` for `brag edit`) →
**stays** (different command, not touched); other `stderr` hits describe
`Aborted.`/`Created project...` messages on other commands → **stays**. No
`README.md`/`BRAG.md` hit claims add-is-silent-on-stderr. (Broader BRAG.md /
tutorial coverage of the milestone *feature* is the STAGE-009 doc-sweep's
job, not this spec — see Out of scope.)

**Additive/count-bump.** `totalThresholds`, `streakThresholds`,
`projectThresholds` are new fixed-shape collections. Grep at design:
`grep -rn "totalThreshold\|streakThreshold\|projectThreshold\|10, 25, 50\|milestone" internal/**/*_test.go`
→ **zero** existing hits (brand-new surface); no existing test asserts their
membership or a literal count, so there is nothing to enumerate as a
count-bump. Their membership is locked by the *new* pure-function tests
below, not by any pre-existing assertion.

**NOT-contains self-audit (§12).** The core assertion `TestAddMilestone_
SilentUnderJSON` proves `add --json` leaves `errBuf` empty *even with TTY
forced on*. Grep the spec's own literal celebratory tokens against the
load-bearing prose and the data paths at design:
`grep -rn "brags and counting\|day streak\|story taking shape\|First brag\|🎉\|🔥\|🎯\|✨" internal/cli/add.go internal/cli/add_json.go`
→ the tokens must appear **only** in `milestone.go` (the stderr path),
**never** in `runAddJSON`, in any `cmd.OutOrStdout()` write, or in the cobra
`Long` string. Expected hits at design: zero in the two files above (the
`Long` milestone note is generic — no celebratory literal); after build,
the tokens live only in `milestone.go`. This is enforced structurally
(`runAddJSON` never calls `emitMilestone`) and at runtime by the silent-
under-JSON test.

## Acceptance Criteria

- [ ] A successful flag-mode `add` that crosses a **total** threshold
      (10/25/50/100/250/500/1000) prints one line to **stderr** on a TTY;
      stdout is still the ID alone.
- [ ] Crossing a **streak** threshold (7/30/100) — current streak moving
      from below the threshold to at-or-above it as a result of this add —
      prints the streak line; the value comes from the SPEC-038-corrected
      `aggregate.Streak` (alive-through-yesterday, local day).
- [ ] Crossing a **per-project** threshold (10th/50th brag in a *non-empty*
      project) prints the per-project line naming the project; an entry with
      no project never earns a per-project milestone.
- [ ] "Crossing" is `before < T <= after`, **not** equality: a second add
      the same day (streak unchanged, or a total/project count that did not
      newly reach a threshold) prints nothing extra.
- [ ] When multiple thresholds cross on one add, exactly **one** line
      prints, chosen by the locked precedence total → streak → per-project →
      first-this-week → first-today.
- [ ] A first-of-period add that crosses no threshold prints the quiet
      nudge: "first brag this week" (which outranks and implies "first brag
      today").
- [ ] `add --json` prints **nothing** to stderr even when stderr is a TTY
      (`errBuf.Len()==0`); a non-TTY/piped flag-mode `add` prints nothing;
      an ordinary add (no crossing, not first-of-period) on a TTY prints
      nothing.
- [ ] The milestone never writes to stdout and never fails the add (a
      compute error is swallowed).
- [ ] No new go.mod dependency; no migration.
- [ ] `go test ./...`, `gofmt -l .`, `go vet ./...` clean;
      `CGO_ENABLED=0 go build ./...` clean.

## Failing Tests

Written during **design**, BEFORE build. Build makes these pass by adding
`internal/cli/milestone.go` and wiring `emitMilestone` into `add.go`. The
pure-function tests (`milestone_test.go`) lock the threshold/precedence/copy
matrix with **no DB and no clock**; the integration tests
(`add_milestone_test.go`) lock the gate + wiring using the date-independent
**total** threshold (plus one end-to-end **streak** test via `Backdate`).
**No `time.Sleep` anywhere** — the one date-dependent test uses the injected
clock / real-now-relative backdating (§9).

### `internal/cli/milestone_test.go` (pure decision function — no DB)

```go
package cli

import "testing"

// TestMilestoneLine_TotalThresholds ▲ locks the total set + copy + that
// crossing (not equality) drives the line. before = total-1 (each add is
// +1), so a total in the set fires; a non-member fires nothing.
func TestMilestoneLine_TotalThresholds(t *testing.T) {
	for _, total := range []int{10, 25, 50, 100, 250, 500, 1000} {
		in := milestoneInputs{total: total}
		got := milestoneLine(in)
		want := "🎉 " + itoa(total) + " brags and counting — nice work!"
		if got != want {
			t.Errorf("total=%d: got %q, want %q", total, got, want)
		}
	}
	for _, total := range []int{1, 9, 11, 24, 26, 99, 101, 999, 1001} {
		if got := milestoneLine(milestoneInputs{total: total}); got != "" {
			t.Errorf("non-threshold total=%d: got %q, want \"\"", total, got)
		}
	}
}

// TestMilestoneLine_StreakThresholds ▲ locks the streak set and the
// crossing-not-equality rule: fires only when streakBefore < T <= streakAfter.
func TestMilestoneLine_StreakThresholds(t *testing.T) {
	for _, s := range []int{7, 30, 100} {
		got := milestoneLine(milestoneInputs{streakBefore: s - 1, streakAfter: s})
		want := "🔥 " + itoa(s) + "-day streak! Keep it going."
		if got != want {
			t.Errorf("streak cross to %d: got %q, want %q", s, got, want)
		}
	}
	// same-day re-add: streak unchanged at a threshold value → no re-fire.
	if got := milestoneLine(milestoneInputs{streakBefore: 7, streakAfter: 7}); got != "" {
		t.Errorf("streak steady at 7: got %q, want \"\"", got)
	}
	// advanced but not onto a threshold → nothing.
	if got := milestoneLine(milestoneInputs{streakBefore: 7, streakAfter: 8}); got != "" {
		t.Errorf("streak 7->8: got %q, want \"\"", got)
	}
}

// TestMilestoneLine_PerProjectThresholds ▲ locks the per-project set, the
// project name in the copy, and that an empty project never earns one.
func TestMilestoneLine_PerProjectThresholds(t *testing.T) {
	// total=11 is a non-total-threshold, so the project tier is what fires
	// (10 and 50 are themselves total thresholds — hold total off the set
	// to isolate the per-project line).
	for _, c := range []int{10, 50} {
		in := milestoneInputs{total: 11, project: "platform", projectCount: c}
		got := milestoneLine(in)
		want := "🎯 " + itoa(c) + " brags on \"platform\" — a story taking shape."
		if got != want {
			t.Errorf("projectCount=%d: got %q, want %q", c, got, want)
		}
	}
	// empty project never earns a per-project milestone.
	if got := milestoneLine(milestoneInputs{total: 11, project: "", projectCount: 10}); got != "" {
		t.Errorf("empty project at count 10: got %q, want \"\"", got)
	}
	// non-threshold project count → nothing.
	if got := milestoneLine(milestoneInputs{total: 11, project: "platform", projectCount: 12}); got != "" {
		t.Errorf("projectCount=12: got %q, want \"\"", got)
	}
}

// TestMilestoneLine_Precedence ▲ locks total → streak → per-project when
// several cross on one add (exactly one line).
func TestMilestoneLine_Precedence(t *testing.T) {
	// total(10) + streak(7) + project(10) all cross → total wins.
	all := milestoneInputs{total: 10, streakBefore: 6, streakAfter: 7, project: "p", projectCount: 10}
	if got := milestoneLine(all); got != "🎉 10 brags and counting — nice work!" {
		t.Errorf("all-cross: got %q, want total line", got)
	}
	// streak(7) + project(10), total non-threshold → streak wins.
	sp := milestoneInputs{total: 11, streakBefore: 6, streakAfter: 7, project: "p", projectCount: 10}
	if got := milestoneLine(sp); got != "🔥 7-day streak! Keep it going." {
		t.Errorf("streak+project: got %q, want streak line", got)
	}
	// project only.
	p := milestoneInputs{total: 11, project: "p", projectCount: 10}
	if got := milestoneLine(p); got != "🎯 10 brags on \"p\" — a story taking shape." {
		t.Errorf("project-only: got %q, want project line", got)
	}
}

// TestMilestoneLine_FirstBragNudges ▲ locks the quiet tier and that it sits
// BELOW all thresholds; first-this-week outranks first-today.
func TestMilestoneLine_FirstBragNudges(t *testing.T) {
	if got := milestoneLine(milestoneInputs{total: 3, firstToday: true, firstThisWeek: true}); got != "✨ First brag this week." {
		t.Errorf("first week+today: got %q, want week line", got)
	}
	if got := milestoneLine(milestoneInputs{total: 3, firstToday: true, firstThisWeek: false}); got != "✨ First brag today." {
		t.Errorf("first today only: got %q, want today line", got)
	}
	// a crossed threshold outranks the nudge.
	if got := milestoneLine(milestoneInputs{total: 10, firstToday: true, firstThisWeek: true}); got != "🎉 10 brags and counting — nice work!" {
		t.Errorf("threshold beats nudge: got %q, want total line", got)
	}
}

// TestMilestoneLine_NoTriggerEmpty ● an ordinary add (nothing crosses, not
// first-of-period) yields no line.
func TestMilestoneLine_NoTriggerEmpty(t *testing.T) {
	if got := milestoneLine(milestoneInputs{total: 3, streakBefore: 1, streakAfter: 1}); got != "" {
		t.Errorf("ordinary add: got %q, want \"\"", got)
	}
}
```

> Note on the `itoa` helper: the tests reference a tiny local
> `func itoa(int) string` (wrap `strconv.Itoa`) so the expected strings are
> built the same way the implementation formats them; build adds it to the
> test file (or inlines `strconv.Itoa`). The expected literals were
> transcribed from Locked decision 5 (the copy artifact) — build diffs
> against that literal, verify diffs against both.

### `internal/cli/add_milestone_test.go` (integration + gating — DB + seams)

Uses `newRootWithAdd(t)` (which now pins `addStderrIsTTY=false`) and
`setStderrIsTTY(t, true)` to opt in. Seeds prior entries by opening the
store at `dbPath` and calling `storage.Store.Add` directly (no CLI, so no
milestone fires during setup).

```go
// TestAddMilestone_FiresOnTTY ▲: 9 prior entries, add the 10th on a TTY →
// the total line on stderr; stdout is still just the new ID.
func TestAddMilestone_FiresOnTTY(t *testing.T) {
	root, dbPath := newRootWithAdd(t)
	seedEntries(t, dbPath, 9, "")           // 9 rows, no project
	setStderrIsTTY(t, true)
	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"--db", dbPath, "add", "--title", "tenth"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(errBuf.String(), "🎉 10 brags and counting") {
		t.Errorf("expected total milestone on stderr, got %q", errBuf.String())
	}
	if _, err := strconv.ParseInt(strings.TrimSpace(outBuf.String()), 10, 64); err != nil {
		t.Errorf("stdout should be the ID alone, got %q", outBuf.String())
	}
}

// TestAddMilestone_SilentUnderJSON ▲ (the §9 split-buffer core + NOT-contains
// enforcement): a milestone WOULD cross (10th entry) but --json is silent
// even with TTY forced ON.
func TestAddMilestone_SilentUnderJSON(t *testing.T) {
	root, dbPath := newRootWithAdd(t)
	seedEntries(t, dbPath, 9, "")
	setStderrIsTTY(t, true)
	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetIn(strings.NewReader(`{"title":"tenth"}`))
	root.SetArgs([]string{"--db", dbPath, "add", "--json"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errBuf.Len() != 0 {
		t.Fatalf("expected stderr empty under --json, got %q", errBuf.String())
	}
	if _, err := strconv.ParseInt(strings.TrimSpace(outBuf.String()), 10, 64); err != nil {
		t.Errorf("stdout should be the ID alone, got %q", outBuf.String())
	}
}

// TestAddMilestone_SilentWhenNotTTY ▲: same crossing, TTY off → nothing.
func TestAddMilestone_SilentWhenNotTTY(t *testing.T) {
	root, dbPath := newRootWithAdd(t)
	seedEntries(t, dbPath, 9, "")
	setStderrIsTTY(t, false)                 // explicit; also the harness default
	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"--db", dbPath, "add", "--title", "tenth"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errBuf.Len() != 0 {
		t.Fatalf("expected stderr empty when not a TTY, got %q", errBuf.String())
	}
}

// TestAddMilestone_EditorModeFires ▲: editor path also emits (the other
// human path). 9 prior entries; editor writes a valid Title; TTY on.
func TestAddMilestone_EditorModeFires(t *testing.T) {
	installAddEditFunc(t, func(path string) error {
		return os.WriteFile(path, []byte("Title: tenth\n\n"), 0o600)
	})
	root, dbPath := newRootWithAdd(t)
	seedEntries(t, dbPath, 9, "")
	setStderrIsTTY(t, true)
	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"--db", dbPath, "add"})   // no field flags → editor mode
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(errBuf.String(), "🎉 10 brags and counting") {
		t.Errorf("expected total milestone via editor mode, got %q", errBuf.String())
	}
}

// TestAddMilestone_PerProjectFires ▲: 9 prior entries in project "platform",
// add a 10th in "platform" → per-project line. Pad with 13 no-project rows so
// the post-add TOTAL is 23 (13+9+1), NOT a total-threshold — isolating the
// per-project tier as the line we observe (10 and 50 are themselves total
// thresholds, so the padding must keep total off the set).
func TestAddMilestone_PerProjectFires(t *testing.T) {
	root, dbPath := newRootWithAdd(t)
	seedEntries(t, dbPath, 13, "")            // 13 no-project rows (total padding)
	seedEntries(t, dbPath, 9, "platform")     // 9 in "platform"; total 22, → 23 after add
	setStderrIsTTY(t, true)
	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"--db", dbPath, "add", "--title", "tenth-plat", "--project", "platform"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(errBuf.String(), `🎯 10 brags on "platform"`) {
		t.Errorf("expected per-project milestone, got %q", errBuf.String())
	}
}

// TestAddMilestone_StreakEndToEnd ▲ proves the CLI reads the SPEC-038
// corrected streak: 6 entries on the 6 LOCAL days ending YESTERDAY (streak
// alive-through-yesterday = 6), add today → crosses 7. Uses the DEFAULT
// addClock (real now) because Store.Add server-stamps the triggering
// entry's created_at to real now and the milestone clock must agree;
// Backdate seeds the six priors relative to that same now. No time.Sleep.
func TestAddMilestone_StreakEndToEnd(t *testing.T) {
	root, dbPath := newRootWithAdd(t)
	now := time.Now()
	ids := seedEntries(t, dbPath, 6, "")
	for i, id := range ids {              // i=0..5 → local days -6..-1 at local noon
		d := now.AddDate(0, 0, -6+i)
		at := time.Date(d.Year(), d.Month(), d.Day(), 12, 0, 0, 0, now.Location())
		if err := storagetest.Backdate(dbPath, id, at); err != nil {
			t.Fatalf("backdate id=%d: %v", id, err)
		}
	}
	setStderrIsTTY(t, true)
	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"--db", dbPath, "add", "--title", "today"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(errBuf.String(), "🔥 7-day streak!") {
		t.Errorf("expected 7-day streak milestone, got %q", errBuf.String())
	}
}

// TestAddMilestone_OrdinaryAddSilentOnTTY ●: 2 prior entries today, add a
// 3rd on a TTY → nothing (total 3 not a threshold; not first today/week;
// streak steady). Proves we don't celebrate every add.
func TestAddMilestone_OrdinaryAddSilentOnTTY(t *testing.T) {
	root, dbPath := newRootWithAdd(t)
	seedEntries(t, dbPath, 2, "")            // both stamped ~now (today)
	setStderrIsTTY(t, true)
	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"--db", dbPath, "add", "--title", "third"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errBuf.Len() != 0 {
		t.Fatalf("expected no milestone for an ordinary add, got %q", errBuf.String())
	}
}
```

**Test helpers to add** (in `add_test.go` or `add_milestone_test.go`):

```go
// setStderrIsTTY overrides the milestone TTY seam for one test and
// restores the real detector on cleanup.
func setStderrIsTTY(t *testing.T, v bool) {
	t.Helper()
	addStderrIsTTY = func() bool { return v }
	t.Cleanup(func() { addStderrIsTTY = defaultStderrIsTTY })
}

// seedEntries inserts n rows directly through the storage layer (no CLI,
// so no milestone fires during setup) and returns their IDs in order.
// project may be "".
func seedEntries(t *testing.T, dbPath string, n int, project string) []int64 {
	t.Helper()
	s, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("storage.Open: %v", err)
	}
	defer s.Close()
	ids := make([]int64, 0, n)
	for i := 0; i < n; i++ {
		e, err := s.Add(storage.Entry{Title: "seed", Project: project})
		if err != nil {
			t.Fatalf("seed Add: %v", err)
		}
		ids = append(ids, e.ID)
	}
	return ids
}
```

**Harness change to `newRootWithAdd`** (premise-audit guard): after building
`root`, pin the seam so existing success tests stay hermetic —

```go
func newRootWithAdd(t *testing.T) (*cobra.Command, string) {
	t.Helper()
	addStderrIsTTY = func() bool { return false }   // hermetic: milestone off unless opted in
	t.Cleanup(func() { addStderrIsTTY = defaultStderrIsTTY })
	root := NewRootCmd("test")
	root.AddCommand(NewAddCmd())
	dbPath := filepath.Join(t.TempDir(), "test.db")
	return root, dbPath
}
```

**Fail-first map (§9):** every ▲ test fails on current `main` for the
*expected* reason — `milestone.go` does not exist yet, so `milestoneLine` /
`emitMilestone` / `addStderrIsTTY` are undefined (a compile error in the
`cli` test binary). That is the correct fail-first signal here: the assertion
cannot run until the symbols exist. Build creates `milestone.go`, the pure
tests then exercise the assertions, and the integration tests exercise the
wiring. The ● tests (`NoTriggerEmpty`, `OrdinaryAddSilentOnTTY`) assert the
*absence* of a line and pass once the module exists and correctly returns
"". Confirm, per §9, that after adding the module the ▲ pure tests fail on a
**wrong string / wrong bool**, not on a stray symbol, before finalizing the
implementation.

## Implementation Context

*Read this section (and the files it points to) before starting the build
cycle. It is the equivalent of a handoff document, folded into the spec.*

### Decisions that apply

- `DEC-023` — **the governing decision** (emitted by this spec). The
  threshold sets, the crossing-not-equality rule, the precedence order, the
  TTY/stderr gate, the `--json`/non-TTY silence, the copy literals, the
  stdlib `os.ModeCharDevice` TTY probe (rejecting a `go-isatty` promotion),
  and the ISO-week definition of "this week." Read it in full before build.
- `DEC-022` — the corrected current-streak (local day, alive through
  yesterday) that the streak milestone reads via `aggregate.Streak`. The
  milestone passes a **local** `now` (`addClock() = time.Now()`), so the
  zone rides on `now.Location()` exactly as `brag stats` does.
- `DEC-011` — stored entry shape / UTC RFC3339 timestamps; the milestone
  reads `CreatedAt` (a UTC instant) and localizes only for the derived
  first-today/this-week checks, never writing anything.
- `DEC-012` — the `brag add --json` schema; the machine path this spec
  keeps byte-clean (never emits a milestone).

### Constraints that apply

- `stdout-is-for-data-stderr-is-for-humans` (**blocking**) — the spine this
  spec is built around: milestone → stderr only, ID → stdout only, silent on
  `--json`/non-TTY. Enforced by the split-buffer tests.
- `test-before-implementation` (blocking) — the tests above are written
  first; build makes them pass.
- `no-cgo` (blocking) — pure Go; `time`/`os`/`strconv` are stdlib; the TTY
  probe is `os.Stderr.Stat()` + `os.ModeCharDevice` (no CGO).
- `no-new-top-level-deps-without-decision` (warning) — **not fired**: the
  stdlib TTY probe means no go.mod change (DEC-023 records the `go-isatty`
  rejection so verify does not expect a dep/DEC).
- `errors-wrap-with-context` (warning) — the milestone compute path wraps
  its one error (`fmt.Errorf("milestone: list entries: %w", err)`), but that
  error is *swallowed* by `emitMilestone` (best-effort; a celebration must
  never fail an add), so it never reaches the CLI boundary. `no-sql-in-cli-
  layer` is honored — the milestone reads through `storage.Store.List`, no
  SQL in `internal/cli`.

### Prior related work

- `SPEC-038` (shipped 2026-07-03, PR #57) — corrected `aggregate.Streak`;
  **this spec's blocker**, now on `main`. The streak milestone is the reason
  the fix had to land first.
- `SPEC-020` (shipped) — introduced `Streak`/`MostCommon`/`Span` and the
  single-`time.Now()`-source pattern the milestone clock follows.
- `SPEC-017` (shipped) — `brag add --json`; the split-buffer round-trip test
  shapes and the machine-path contract this spec must not pollute.
- `SPEC-019`/`SPEC-020` — the canonical `outBuf`/`errBuf` split-buffer test
  shape (separate buffers, assert no cross-leakage).

### Out of scope (for this spec specifically)

- The MCP surface (SPEC-040) and the plugin/hook (SPEC-041). The milestone
  is CLI-only; the MCP `brag_add` tool does **not** emit a milestone (its
  stdout is the protocol stream — SPEC-040 handles that spine at its own
  transport).
- Broad docs for the milestone *feature* in `BRAG.md` / `docs/tutorial.md` /
  `docs/architecture.md` — the STAGE-009 doc-sweep owns those. This spec
  touches only `docs/api-contract.md` (the `brag add` I/O contract it
  directly changes) plus the add `--help` note.
- Persisting a "already celebrated" flag. The core is migration-free; a
  milestone can re-fire if the corpus is deleted back below a threshold and
  re-crosses it — accepted (rare, and genuinely *is* a re-crossing). First-
  class "celebrated-once" state would need schema and is explicitly not in
  the v0.3.0 core.
- A configurable / silenceable milestone (env var, `--quiet`, `NO_COLOR`,
  emoji-off plain mode). TTY-gating already silences every scripted/agent
  path; a user toggle is a future-if-asked item (noted in DEC-023).
- Any change to `stats`/`summary`/`review` or to `aggregate` itself — the
  milestone only *reads* `aggregate.Streak`.

## Notes for the Implementer

- **Module shape (guidance; the literals in Locked decisions 3–5 are
  binding).** `milestone.go` has three parts: (1) the pure
  `milestoneLine(milestoneInputs) string` + `crossed` + the three threshold
  slices; (2) the seams `addClock`/`addStderrIsTTY`/`defaultStderrIsTTY`;
  (3) the glue `emitMilestone(cmd, s, inserted)` → `computeMilestoneInputs`.
  Keep the pure function free of Store/clock/TTY so `milestone_test.go`
  needs none of them.
- **`computeMilestoneInputs` reference shape:**
  ```go
  func computeMilestoneInputs(s *storage.Store, inserted storage.Entry, now time.Time) (milestoneInputs, error) {
      all, err := s.List(storage.ListFilter{})
      if err != nil {
          return milestoneInputs{}, fmt.Errorf("milestone: list entries: %w", err)
      }
      loc := now.Location()
      today := now.In(loc).Format("2006-01-02")
      ny, nw := now.In(loc).ISOWeek()
      before := make([]storage.Entry, 0, len(all))
      projectCount := 0
      firstToday, firstThisWeek := true, true
      for _, e := range all {
          if e.ID != inserted.ID {
              before = append(before, e)
              d := e.CreatedAt.In(loc)
              if d.Format("2006-01-02") == today {
                  firstToday = false
              }
              if y, w := d.ISOWeek(); y == ny && w == nw {
                  firstThisWeek = false
              }
          }
          if inserted.Project != "" && e.Project == inserted.Project {
              projectCount++
          }
      }
      sb, _ := aggregate.Streak(before, now)
      sa, _ := aggregate.Streak(all, now)
      return milestoneInputs{
          total: len(all), project: inserted.Project, projectCount: projectCount,
          streakBefore: sb, streakAfter: sa, firstToday: firstToday, firstThisWeek: firstThisWeek,
      }, nil
  }
  ```
  `firstThisWeek ⟹ firstToday` (this week ⊇ today), so the precedence
  (week before today) always shows the "bigger" nudge — no contradictory
  pair is reachable.
- **`emitMilestone` gates then delegates:**
  ```go
  func emitMilestone(cmd *cobra.Command, s *storage.Store, inserted storage.Entry) {
      if !addStderrIsTTY() {
          return
      }
      in, err := computeMilestoneInputs(s, inserted, addClock())
      if err != nil {
          return // best-effort: never fail an add on a celebration
      }
      if line := milestoneLine(in); line != "" {
          fmt.Fprintln(cmd.ErrOrStderr(), line)
      }
  }
  ```
  Call it in `runAddFlags` and `runAddEditor` **after** the ID is printed
  and while `s` is still open (before the deferred `s.Close()`). Do **not**
  call it in `runAddJSON`.
- **`defaultStderrIsTTY` probes the real `os.Stderr`, not
  `cmd.ErrOrStderr()`** — the gate is about the process's real terminal;
  the *write* still goes to `cmd.ErrOrStderr()` so tests capture it in
  `errBuf`:
  ```go
  func defaultStderrIsTTY() bool {
      fi, err := os.Stderr.Stat()
      if err != nil {
          return false
      }
      return fi.Mode()&os.ModeCharDevice != 0
  }
  ```
- **`id=$(brag add ...)` is safe.** In command substitution only stdout is
  captured; stderr stays attached to the terminal, so a milestone on stderr
  neither pollutes the captured ID nor is suppressed — it shows and is
  correct. The TTY gate keys on stderr, which is exactly the stream the
  milestone uses.
- **The cobra `Long` note must stay generic** (NOT-contains self-audit): add
  something like — *"On a terminal, `brag add` prints a short milestone or
  first-brag note to stderr on success; scripted (`--json`), piped, and
  non-terminal runs stay silent, and stdout always carries just the entry
  ID."* — with **no** emoji and **no** celebratory literal, so the milestone
  tokens live only in `milestone.go`. Then `TestAdd_HelpMentionsMilestone`
  asserts a generic substring (e.g. `"milestone"`).
- **Do not add a `time.Sleep`.** The only date-dependent test uses the
  default clock with `Backdate`-seeded priors relative to real `now` (§9);
  everything else is date-independent (total threshold) or clock-free (pure
  function).
- **Fail-first check (build step):** `go test ./internal/cli/` will not
  compile until `milestone.go` exists (undefined symbols) — that is the
  expected first state. After adding the module, re-run and confirm the ▲
  pure tests fail on wrong *strings/bools* if the copy or precedence is off,
  before declaring the implementation done.

## Locked design decisions

Each behavior decision (1–6) has ≥1 paired test that fails without it (§9).

1. **The milestone is TTY-gated to stderr; `--json` and non-TTY are silent.**
   Gate on the real `os.Stderr` being a char device; write to
   `cmd.ErrOrStderr()`; `runAddJSON` never calls the hook. *Pair (▲):*
   `TestAddMilestone_FiresOnTTY`, `TestAddMilestone_SilentUnderJSON`
   (TTY on, still silent), `TestAddMilestone_SilentWhenNotTTY`.
   - **Rejected alternative:** gate on `cmd.ErrOrStderr()` being a terminal
     — it's a `bytes.Buffer` in tests and `os.Stderr` in prod; probing the
     real fd via an injectable seam is what makes the gate both correct in
     prod and deterministic in tests.

2. **"Crossing" is `before < T <= after`, not equality.** A second same-day
   add (streak steady, or a count that did not newly reach a threshold)
   stays silent. *Pair (▲):* `TestMilestoneLine_StreakThresholds`
   (steady-at-7 → "") and the total/project non-member cases.
   - **Rejected alternative:** fire on `post == T`. Rejected — it double-
     fires the streak milestone on every same-day re-add once the streak is
     at a threshold value.

3. **Threshold sets are fixed literals:** total
   `{10, 25, 50, 100, 250, 500, 1000}`, streak `{7, 30, 100}`, per-project
   `{10, 50}`. Per-project fires only for a **non-empty** project. *Pair
   (▲):* the three `TestMilestoneLine_*Thresholds` tests (members fire,
   non-members and empty-project don't).

4. **Precedence (exactly one line):** total → streak → per-project →
   first-this-week → first-today. *Pair (▲):* `TestMilestoneLine_Precedence`
   and `TestMilestoneLine_FirstBragNudges` (nudge sits below thresholds;
   week outranks today).
   - Confidence note: the *order among simultaneously-crossed thresholds* is
     a low-stakes judgment call (rare co-occurrence); locked here for
     determinism, flagged at DEC-023 confidence 0.8. The gating/crossing
     mechanics are high-confidence.

5. **Copy literals (the §12 literal artifact — transcribe verbatim):**
   | Tier | Format (Go, `%d`/`%q` as shown) |
   |---|---|
   | total | `🎉 %d brags and counting — nice work!` |
   | streak | `🔥 %d-day streak! Keep it going.` |
   | per-project | `🎯 %d brags on %q — a story taking shape.` |
   | first-this-week | `✨ First brag this week.` |
   | first-today | `✨ First brag today.` |
   The em dash is `—` (U+2014). `%q` on the project renders Go-quoted
   (`"platform"`). Each line is printed with `Fprintln` (one trailing
   newline). *Pair (▲):* every pure-function test asserts the exact string.

6. **TTY detection is stdlib `os.ModeCharDevice`, not a new dependency.**
   `os.Stderr.Stat().Mode()&os.ModeCharDevice != 0`. *Validated by absence*
   (no go.mod change; `no-new-top-level-deps-without-decision` does not
   fire) and by the gating tests exercising the injectable seam.
   - **Rejected alternative:** promote the already-*indirect* `mattn/go-isatty`
     to a direct dependency. Rejected in DEC-023 — the stdlib probe is
     sufficient for "is stderr a terminal," pure-Go, and avoids a DEC-gated
     dependency for a cosmetic feature.

**Doc-text change (prose, not behavior — no paired test):**
`docs/api-contract.md` `brag add` section drops the unqualified "stderr
empty" for the TTY-gated milestone contract (enumerated in the premise
audit); no test asserts that prose, so none is added.

---

## Build Completion

*Filled in at the end of the **build** cycle, before advancing to verify.*

- **Branch:** `feat/spec-039-milestone-notifications`
- **PR (if applicable):** [#59](https://github.com/jysf/bragfile000/pull/59)
  — carries design + build (see Deviations: the design PR was never merged
  separately; this repo's house style is one branch/PR per spec, as SPEC-038
  shipped via a single PR #57).
- **All acceptance criteria met?** yes — all boxes hold:
  total/streak/per-project crossings each fire their line on a TTY
  (`TestAddMilestone_FiresOnTTY`, `_PerProjectFires`, `_StreakEndToEnd`);
  streak reads the SPEC-038-corrected `aggregate.Streak` (the end-to-end test
  seeds 6 local days ending yesterday via `Backdate`, adds today → crosses 7);
  crossing-not-equality (`TestMilestoneLine_StreakThresholds` steady-at-7 →
  "", `_OrdinaryAddSilentOnTTY`); one line by locked precedence
  (`_Precedence`, `_FirstBragNudges`); silent under `--json` with TTY forced
  on (`_SilentUnderJSON`, `errBuf.Len()==0`), non-TTY (`_SilentWhenNotTTY`),
  and ordinary adds; never on stdout (ID-parses assertions); no new go.mod
  dep, no migration. Gates green: `go test ./...` (555), `gofmt -l .` empty,
  `go vet ./...` clean, `CGO_ENABLED=0 go test ./...` green; `scripts/
  test-docs.sh` ALL OK.
- **New decisions emitted:**
  - `DEC-023` — milestone notifications copy & semantics (emitted at design;
    no new DEC at build). No go.mod change, so
    `no-new-top-level-deps-without-decision` did not fire (stdlib
    `os.ModeCharDevice` probe, as locked).
- **Deviations from spec:**
  - **None on behavior.** Tests transcribed verbatim; `milestone.go` built to
    the locked shapes; the copy literals match DEC-023 / Locked decision 5
    exactly; `emitMilestone` wired into `runAddFlags` + `runAddEditor` (not
    `runAddJSON`); `newRootWithAdd` pinned `addStderrIsTTY=false`;
    `docs/api-contract.md` reworded per the premise audit.
  - **Two mechanical modernizations** (not spec-visible): the spec's
    `seedEntries` reference used `for i := 0; i < n; i++`; built as
    `for range n` (Go 1.26 range-over-int) to keep the tree lint-clean. No
    behavior change.
  - **Process note (not a code deviation):** the precondition "design PR #59
    is merged to main" was **false** at build start — #59 was still OPEN and
    `origin/main` lacked the spec/DEC-023. The feat branch was already
    correctly based on the current main tip (`59314c6`), matching this repo's
    one-branch-per-spec model (cf. SPEC-038 / PR #57). Build proceeded on the
    existing feat branch per the operative rules ("do all work on
    `feat/spec-039-milestone-notifications`… open/confirm the PR"); main was
    never touched. Flagged for the coordinator.
- **Follow-up work identified:**
  - None new. SPEC-040 (MCP) and SPEC-041 (plugin) remain the STAGE-009
    backlog; the milestone is CLI-only and does not touch them.

## Reflection (Ship)

*Appended during the **ship** cycle.*

1. **What would I do differently next time?**
   — Almost nothing. The pure-function-plus-thin-glue split (decision matrix
   in `milestoneLine` with no Store/clock/TTY; ~45-line glue layer tested for
   wiring only) made both build and verify fast and mechanical. The
   literal-artifact-as-spec pattern for the copy table (Locked decision 5)
   eliminated all copy-string ambiguity: verify simply diffs against the table.
   One observation for next time: the process discrepancy (design PR #59 was
   still open at build start, not merged as the spec assumed) was friction-free
   in practice because the one-branch-per-spec model meant build simply
   continued on the existing feat branch — but a single sentence in the spec's
   preconditions making the "open PR, not merged PR" state explicit would have
   prevented the need to flag it at all.

2. **Does any template, constraint, or decision need updating?**
   — No update earned now. DEC-023 is the durable home for the milestone
   copy/semantics/gating; the pure-function-plus-thin-glue shape is a WATCH
   item (N=1, per §9 same-outcome N=3 bar). One pattern worth watching:
   the premise-audit "stays" row discipline — enumerating why a doc/test hit
   is out of scope, not just that it stays — pre-empted verify ambiguity here
   the same way it would have in SPEC-038 (noted there as a watch item at N=1).
   Two cases now. Not yet N=3, but converging; keep watching.

3. **Is there a follow-up spec I should write now before I forget?**
   — No new spec needed. SPEC-040 (MCP server) is already in the STAGE-009
   backlog and is the natural next spec; it consumes the same stdout-is-data
   spine at a new transport that SPEC-039 exercised at the CLI. No new
   follow-up beyond what is already planned.

---

### Build-phase reflection (3 questions, short answers)

1. **What was unclear in the spec that slowed you down?**
   — Nothing in the spec itself. The only friction was the environment: the
   "design PR merged" precondition contradicted the actual git state (open
   #59, one-branch model). The spec's own reference shapes
   (`computeMilestoneInputs`, `emitMilestone`, `defaultStderrIsTTY`, the copy
   table) made the implementation a transcription, and the fail-first map
   correctly predicted the undefined-symbol compile failure as the expected
   first state.

2. **Was there a constraint or decision that should have been listed but wasn't?**
   — No. `references.constraints`
   (stdout-is-for-data-stderr-is-for-humans, test-before-implementation,
   no-cgo, no-new-top-level-deps-without-decision, errors-wrap-with-context)
   and DEC-023/022/011/012 covered everything the build touched. The
   `go-isatty`-vs-stdlib decision was pre-settled in DEC-023 Locked decision
   6, so the TTY probe was never an open question.

3. **If you did this task again, what would you do differently?**
   — Nothing structural. One observation worth a WATCH (N=1, not codified):
   the pure-function-plus-thin-glue split (decision matrix with no
   Store/clock/TTY, tested exhaustively without a DB; a ~45-line glue layer
   tested for wiring only) made both the design and the build cheap and is a
   reusable shape for any "compute-then-render-a-human-line" CLI feature.
   Parks alongside the existing testing-conventions habits; revisit if a
   second feature (e.g. an MCP-side or `stats`-side nudge) re-uses it.
