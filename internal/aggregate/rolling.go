package aggregate

import (
	"time"

	"github.com/jysf/bragfile000/internal/storage"
)

// RollingBuckets buckets entries onto a fixed rolling axis of n buckets each
// `width` wide whose LAST bucket ends (exclusive) at end. The axis start is
// end.Add(-width*n); bucket k (0-indexed) covers
// [start+k*width, start+(k+1)*width) — lower-inclusive, upper-exclusive.
// Returns exactly n zero-filled counts (spark/JSON-ready); entries before start
// or at/after end are excluded, so sum(result) == the in-axis subset size. Pure,
// stdlib-only, instant-arithmetic (location-independent) — the sub-month analog
// of Cadence, which is calendar-month-only. SPEC-059/DEC-037.
func RollingBuckets(entries []storage.Entry, end time.Time, width time.Duration, n int) []int {
	start := end.Add(-width * time.Duration(n))
	out := make([]int, n)
	for _, e := range entries {
		t := e.CreatedAt
		if t.Before(start) || !t.Before(end) {
			continue
		}
		k := int(t.Sub(start) / width)
		if k >= 0 && k < n {
			out[k]++
		}
	}
	return out
}
