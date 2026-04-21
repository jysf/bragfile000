package cli

import (
	"fmt"
	"strconv"
	"time"
)

// ParseSince parses the --since flag value per DEC-008. It accepts:
//   - an ISO date YYYY-MM-DD, returned as midnight UTC on that day;
//   - a relative duration Nd / Nw / Nm (positive integer N, units
//     d=days, w=7 days, m=30 days — month is approximate), returned
//     as time.Now().UTC() - duration.
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
	return time.Now().UTC().Add(-d), nil
}
