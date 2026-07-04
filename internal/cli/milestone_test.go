package cli

import (
	"strconv"
	"testing"
)

// itoa builds expected numbers the same way the implementation formats them,
// so the expected copy strings stay in lockstep with the literals.
func itoa(n int) string { return strconv.Itoa(n) }

// TestMilestoneLine_TotalThresholds locks the total set + copy + that
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

// TestMilestoneLine_StreakThresholds locks the streak set and the
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

// TestMilestoneLine_PerProjectThresholds locks the per-project set, the
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

// TestMilestoneLine_Precedence locks total → streak → per-project when
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

// TestMilestoneLine_FirstBragNudges locks the quiet tier and that it sits
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

// TestMilestoneLine_NoTriggerEmpty: an ordinary add (nothing crosses, not
// first-of-period) yields no line.
func TestMilestoneLine_NoTriggerEmpty(t *testing.T) {
	if got := milestoneLine(milestoneInputs{total: 3, streakBefore: 1, streakAfter: 1}); got != "" {
		t.Errorf("ordinary add: got %q, want \"\"", got)
	}
}
