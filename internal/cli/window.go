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

// windowCutoff computes the inclusive lower bound, the exclusive upper
// bound, and the scope token for the selected window. Pure and
// deterministic given (window, sinceRaw, now, previous). CALENDAR periods
// (time.Date constructors), NEVER day subtraction — this is the
// correctness core of the calendar-vs-rolling divergence (DEC-028 choice
// 1). now is UTC.
//
// The `end` return is the load-bearing addition for `--previous` (DEC-032
// / SPEC-053):
//
//   - previous == false: the CURRENT period. start is the current-period
//     cutoff (math unchanged); end is the ZERO time.Time SENTINEL meaning
//     "open upper edge / up to now" (every stored created_at <= now, so the
//     lower bound alone bounds the window). Callers treat a zero end as "no
//     upper-bound filter", preserving the [cutoff, now] behavior
//     byte-for-byte. scope is today's token ("quarter" / "since:<raw>" …).
//   - previous == true: the LAST-COMPLETED period, a bounded
//     [prev-start, prev-end) window. start is the current-period start
//     shifted back one period via AddDate (never day subtraction — rolls
//     year boundaries: a January --month --previous lands in the prior
//     December of the prior year). end is the CURRENT period's start (the
//     exclusive upper boundary between the completed period and the
//     in-progress one). scope is "<window>:previous".
//
// `--previous` is undefined for --since (an explicit anchor, not a calendar
// period): windowCutoff returns a UserError, the helper-level guard backing
// up the CLI's flag-combo check (DEC-032 choice 4 / LD3).
//
// Lifted verbatim from impact.go at SPEC-049 (the third-caller threshold,
// SPEC-018) so `impact` and `story` share one calendar core; impact's
// existing tests stay green byte-for-byte on the current-period path.
func windowCutoff(window, sinceRaw string, now time.Time, previous bool) (start, end time.Time, scope string, err error) {
	switch window {
	case "quarter":
		qStartMonth := ((int(now.Month())-1)/3)*3 + 1 // 1, 4, 7, 10
		curStart := time.Date(now.Year(), time.Month(qStartMonth), 1, 0, 0, 0, 0, time.UTC)
		if previous {
			return curStart.AddDate(0, -3, 0), curStart, "quarter:previous", nil
		}
		return curStart, time.Time{}, "quarter", nil
	case "month":
		curStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		if previous {
			return curStart.AddDate(0, -1, 0), curStart, "month:previous", nil
		}
		return curStart, time.Time{}, "month", nil
	case "year":
		curStart := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
		if previous {
			return curStart.AddDate(-1, 0, 0), curStart, "year:previous", nil
		}
		return curStart, time.Time{}, "year", nil
	case "since":
		if previous {
			return time.Time{}, time.Time{}, "", UserErrorf("--previous cannot be combined with --since (--since is an explicit anchor, not a calendar period)")
		}
		start, err = ParseSince(sinceRaw)
		if err != nil {
			return time.Time{}, time.Time{}, "", UserErrorf("invalid --since value: %v", err)
		}
		return start, time.Time{}, "since:" + sinceRaw, nil
	default:
		return time.Time{}, time.Time{}, "", fmt.Errorf("windowCutoff: unhandled window %q", window)
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
