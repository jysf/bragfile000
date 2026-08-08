// Package timewindow parses the user-facing time-window vocabulary shared by
// every bragfile surface — the `--since` grammar (DEC-008) and the `--day`
// calendar-day grammar (DEC-039) — into `storage.ListFilter` bounds.
//
// The package is deliberately PURE: every function takes the current instant
// as an explicit `now` argument and the package holds no state, no clock, and
// no imports outside the standard library (DEC-042). The wall-clock seam
// belongs to each CALLER, because callers legitimately disagree about what
// "now" means — `internal/cli` reads a LOCAL clock so a "day" is the user's
// day, while `internal/cli/impact.go` anchors calendar reporting windows in
// UTC. A package-level seam here would force those policies into one
// variable that tests in two packages would race on; a parameter does not.
//
// This is also what makes MCP/CLI grammar parity structural rather than
// maintained: `brag list --since 7d` and the `brag_list` tool's
// `{"since":"7d"}` are the same function call, not two implementations kept in
// agreement (DEC-042; the alternative — duplicating the parser into
// internal/mcpserver — is the cost DEC-024 already recorded for the DEC-010
// tokenizer).
package timewindow

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ParseSince parses a `--since`-style value per DEC-008, relative to now. It
// accepts:
//   - an ISO date YYYY-MM-DD, returned as midnight UTC on that day (the bare
//     date is UTC-anchored and ignores now entirely, deliberately unlike
//     ParseDay — see DEC-039);
//   - a relative duration Nd / Nw / Nm (positive integer N, units d=days,
//     w=7 days, m=30 days — month is approximate), returned as now minus that
//     duration, normalized to UTC.
//
// Any other input (empty, whitespace-padded, negative or zero N, unknown
// unit) is an error; callers surface it as a UserError (CLI) or a tool error
// (MCP).
func ParseSince(s string, now time.Time) (time.Time, error) {
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t.UTC(), nil
	}
	if len(s) < 2 {
		return time.Time{}, fmt.Errorf("expected YYYY-MM-DD or Nd/Nw/Nm, got %q", s)
	}
	unit := s[len(s)-1]
	nstr := s[:len(s)-1]
	n, err := strconv.Atoi(nstr)
	if err != nil || n <= 0 {
		return time.Time{}, fmt.Errorf("expected positive integer before %c, got %q", unit, nstr)
	}
	var d time.Duration
	switch unit {
	case 'd':
		d = time.Duration(n) * 24 * time.Hour
	case 'w':
		d = time.Duration(n) * 7 * 24 * time.Hour
	case 'm':
		d = time.Duration(n) * 30 * 24 * time.Hour
	default:
		return time.Time{}, fmt.Errorf("unknown unit %q in %q (use d/w/m)", string(unit), s)
	}
	return now.Add(-d).UTC(), nil
}

// ParseDay parses a `--day`-style value (DEC-039) into the half-open
// [start, end) window bounding a single LOCAL calendar day. It accepts:
//   - the keywords "today" / "yesterday" (case-insensitive), resolved against
//     now;
//   - a bare ISO date YYYY-MM-DD, taken as that LOCAL calendar day.
//
// "Local" is now.Location(): the caller's clock carries the zone, so this
// function needs no notion of the host's timezone. start is local midnight
// and end is the NEXT local midnight (via AddDate, so the boundary is
// DST-correct — never a fixed 24h Add). Both bounds carry now's location;
// storage normalizes them to UTC RFC3339 for the comparison
// (timestamps-in-utc-rfc3339 governs storage, not this derived boundary — the
// DEC-022 "derive-local, store-UTC" carve-out).
//
// A "day" is LOCAL here, deliberately unlike bare-date ParseSince (which stays
// UTC-midnight for backward compatibility) — see DEC-039.
func ParseDay(value string, now time.Time) (start, end time.Time, err error) {
	loc := now.Location()
	var y int
	var m time.Month
	var d int
	switch strings.ToLower(value) {
	case "today":
		y, m, d = now.Date()
	case "yesterday":
		y, m, d = now.AddDate(0, 0, -1).Date()
	default:
		t, perr := time.ParseInLocation("2006-01-02", value, loc)
		if perr != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("expected YYYY-MM-DD, \"today\", or \"yesterday\", got %q", value)
		}
		y, m, d = t.Date()
	}
	start = time.Date(y, m, d, 0, 0, 0, 0, loc)
	end = start.AddDate(0, 0, 1)
	return start, end, nil
}
