package cli

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// clock is the injectable local-wall-clock seam (AGENTS.md §9). It
// returns the current instant carrying the LOCAL location (time.Local
// in production), so callers that need a local calendar day — ParseDay
// below — read the day boundary off `clock().Location()`. Tests
// substitute a fixed instant+zone to make keyword/day resolution
// deterministic across timezones (DEC-039; the audit's L4 fix — the
// --since duration path below no longer calls time.Now() inline).
//
// Distinct from impact.go's `nowFunc`, which deliberately returns UTC
// for the calendar-window commands; a "day" is a local concept, a
// reporting quarter/month/year is anchored in UTC there.
var clock = time.Now

// ParseSince parses the --since flag value per DEC-008. It accepts:
//   - an ISO date YYYY-MM-DD, returned as midnight UTC on that day;
//   - a relative duration Nd / Nw / Nm (positive integer N, units
//     d=days, w=7 days, m=30 days — month is approximate), returned
//     as clock() - duration, normalized to UTC.
//
// Any other input (empty, whitespace, negative N, unknown unit) is an
// error; the CLI layer surfaces it as UserError.
func ParseSince(s string) (time.Time, error) {
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
	return clock().Add(-d).UTC(), nil
}

// ParseDay parses the --day flag value (SPEC-068 / DEC-039) into the
// half-open [start, end) window bounding a single LOCAL calendar day.
// It accepts:
//   - the keywords "today" / "yesterday" (case-insensitive), resolved
//     against the local wall clock (clock());
//   - a bare ISO date YYYY-MM-DD, taken as that LOCAL calendar day.
//
// start is local midnight and end is the NEXT local midnight (via
// AddDate, so the boundary is DST-correct — never a fixed 24h Add). Both
// bounds carry the local location; storage normalizes them to UTC
// RFC3339 for the comparison (timestamps-in-utc-rfc3339 governs storage,
// not this derived boundary — the DEC-022 "derive-local, store-UTC"
// carve-out). Any other value is an error naming the accepted forms;
// the CLI layer surfaces it as UserError.
//
// A "day" is LOCAL here, deliberately unlike bare-date --since (which
// stays UTC-midnight for backward compatibility) — see DEC-039.
func ParseDay(value string) (start, end time.Time, err error) {
	loc := clock().Location()
	var y int
	var m time.Month
	var d int
	switch strings.ToLower(value) {
	case "today":
		y, m, d = clock().Date()
	case "yesterday":
		y, m, d = clock().AddDate(0, 0, -1).Date()
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
