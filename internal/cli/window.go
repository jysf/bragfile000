package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// windowFlagNames is the canonical, ordered set of calendar-window flags
// shared by `brag impact` and `brag story`.
var windowFlagNames = []string{"quarter", "month", "year", "since"}

// windowCutoff computes the inclusive lower bound and scope token for the
// selected window. Pure and deterministic given (window, sinceRaw, now).
// CALENDAR periods (time.Date constructors), NEVER day subtraction — this
// is the correctness core of the calendar-vs-rolling divergence (DEC-028
// choice 1). now is UTC; the period end is always "now" (implicit — every
// stored created_at <= now, so the lower bound alone bounds the window).
//
// Lifted verbatim from impact.go at SPEC-049 (the third-caller threshold,
// SPEC-018) so `impact` and `story` share one calendar core; impact's
// existing tests stay green byte-for-byte.
func windowCutoff(window, sinceRaw string, now time.Time) (cutoff time.Time, scope string, err error) {
	switch window {
	case "quarter":
		qStartMonth := ((int(now.Month())-1)/3)*3 + 1 // 1, 4, 7, 10
		cutoff = time.Date(now.Year(), time.Month(qStartMonth), 1, 0, 0, 0, 0, time.UTC)
		return cutoff, "quarter", nil
	case "month":
		cutoff = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		return cutoff, "month", nil
	case "year":
		cutoff = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
		return cutoff, "year", nil
	case "since":
		cutoff, err = ParseSince(sinceRaw)
		if err != nil {
			return time.Time{}, "", UserErrorf("invalid --since value: %v", err)
		}
		return cutoff, "since:" + sinceRaw, nil
	default:
		return time.Time{}, "", fmt.Errorf("windowCutoff: unhandled window %q", window)
	}
}

// selectedWindow returns the single set window flag's canonical name, or a
// UserError if zero or two-plus are set (mutually exclusive + required,
// DEC-028 choice 1 / DEC-007). Cobra's MarkFlagsMutuallyExclusive handles
// pairs but not "exactly one required" cleanly across a bool+string mix
// and routes its error off the UserError→stderr path, so the check is
// explicit here.
//
// The zero-flag UserError is REQUIRED by `brag impact` (a window is
// mandatory there). `brag story` does NOT call this when no window flag is
// set — it uses windowFlagsSet to gate the call and supplies the profile
// default instead (see story.go). This keeps selectedWindow unchanged.
func selectedWindow(cmd *cobra.Command) (string, error) {
	var set []string
	for _, name := range windowFlagNames {
		if cmd.Flags().Changed(name) {
			set = append(set, name)
		}
	}
	switch len(set) {
	case 0:
		return "", UserErrorf("one of --quarter, --month, --year, --since is required")
	case 1:
		return set[0], nil
	default:
		return "", UserErrorf("--quarter, --month, --year, --since are mutually exclusive (got --%s)", strings.Join(set, ", --"))
	}
}

// windowFlagsSet reports whether the user set any calendar-window flag. It
// lets `brag story` decide between "explicit window" (call selectedWindow,
// which enforces exactly-one) and "no window → use the profile default"
// without changing selectedWindow's zero-flag error that `impact` relies
// on.
func windowFlagsSet(cmd *cobra.Command) bool {
	for _, name := range windowFlagNames {
		if cmd.Flags().Changed(name) {
			return true
		}
	}
	return false
}
