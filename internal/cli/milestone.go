package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/jysf/bragfile000/internal/aggregate"
	"github.com/jysf/bragfile000/internal/storage"
	"github.com/spf13/cobra"
)

// Milestone thresholds — fixed-shape collections (DEC-023). Crossing any of
// these on an add earns one celebratory stderr line.
var (
	totalThresholds   = []int{10, 25, 50, 100, 250, 500, 1000}
	streakThresholds  = []int{7, 30, 100}
	projectThresholds = []int{10, 50}
)

// addClock is the clock the milestone reads. Package-level so tests inject a
// fixed instant/zone; the zone rides on now.Location() (DEC-022), matching
// how brag stats sources a single local time.Now().
var addClock = time.Now

// addStderrIsTTY reports whether the milestone may print — i.e. whether the
// process's real stderr is a terminal. Package-level seam so tests force it
// on/off deterministically (no real terminal in tests).
var addStderrIsTTY = defaultStderrIsTTY

// defaultStderrIsTTY probes the real os.Stderr (not cmd.ErrOrStderr(), which
// is a buffer in tests) for a char device — stdlib only, no go-isatty
// dependency (DEC-023 Locked decision 6).
func defaultStderrIsTTY() bool {
	fi, err := os.Stderr.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

// milestoneInputs carries the already-computed metrics the decision needs.
// Keeping the decision pure (no Store, clock, or TTY) makes the
// threshold/precedence/copy matrix exhaustively unit-testable.
type milestoneInputs struct {
	total         int    // lifetime total AFTER this add
	project       string // the added entry's project ("" = no project)
	projectCount  int    // count in that project AFTER this add (0 if project == "")
	streakBefore  int    // current streak BEFORE this add
	streakAfter   int    // current streak AFTER this add
	firstToday    bool   // this add is the first entry today (local)
	firstThisWeek bool   // this add is the first entry this ISO week (local)
}

// milestoneLine returns the single celebratory line (no trailing newline) to
// print to stderr, or "" if this add earns none. Precedence (exactly one line
// ever): total → streak → per-project → first-this-week → first-today.
// "Crossing" is before < T <= after, so a same-day re-add never re-fires
// (DEC-023).
func milestoneLine(in milestoneInputs) string {
	if t, ok := crossed(in.total-1, in.total, totalThresholds); ok {
		return fmt.Sprintf("🎉 %d brags and counting — nice work!", t)
	}
	if t, ok := crossed(in.streakBefore, in.streakAfter, streakThresholds); ok {
		return fmt.Sprintf("🔥 %d-day streak! Keep it going.", t)
	}
	if in.project != "" {
		if t, ok := crossed(in.projectCount-1, in.projectCount, projectThresholds); ok {
			return fmt.Sprintf("🎯 %d brags on %q — a story taking shape.", t, in.project)
		}
	}
	if in.firstThisWeek {
		return "✨ First brag this week."
	}
	if in.firstToday {
		return "✨ First brag today."
	}
	return ""
}

// crossed reports whether some threshold T satisfies before < T <= after,
// returning the largest such T. With single-step increments (each add bumps a
// metric by at most 1) at most one threshold is crossed.
func crossed(before, after int, thresholds []int) (int, bool) {
	hit, ok := 0, false
	for _, t := range thresholds {
		if before < t && t <= after {
			hit, ok = t, true
		}
	}
	return hit, ok
}

// emitMilestone prints a milestone/nudge line to stderr after a successful
// flag- or editor-mode add, when stderr is a terminal. Best-effort: any error
// computing the milestone is swallowed (a celebration must never fail an
// add). Never called from --json mode (the machine path stays byte-clean).
func emitMilestone(cmd *cobra.Command, s *storage.Store, inserted storage.Entry) {
	if !addStderrIsTTY() {
		return
	}
	in, err := computeMilestoneInputs(s, inserted, addClock())
	if err != nil {
		return
	}
	if line := milestoneLine(in); line != "" {
		fmt.Fprintln(cmd.ErrOrStderr(), line)
	}
}

// computeMilestoneInputs derives the milestone metrics from the post-insert
// corpus. "Before" is the corpus minus the just-inserted row (by ID); the
// first-today / first-this-week checks localize each stored UTC instant into
// now.Location() (DEC-022 derive-local, store-UTC — nothing is written).
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
		total:         len(all),
		project:       inserted.Project,
		projectCount:  projectCount,
		streakBefore:  sb,
		streakAfter:   sa,
		firstToday:    firstToday,
		firstThisWeek: firstThisWeek,
	}, nil
}
