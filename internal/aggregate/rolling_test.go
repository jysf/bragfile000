package aggregate

import (
	"reflect"
	"testing"
	"time"

	"github.com/jysf/bragfile000/internal/spark"
	"github.com/jysf/bragfile000/internal/storage"
)

// entriesAt builds a []storage.Entry whose only meaningful field is
// CreatedAt, one entry per supplied instant. Hermetic; no clock, no I/O.
func entriesAt(times ...time.Time) []storage.Entry {
	out := make([]storage.Entry, 0, len(times))
	for _, t := range times {
		out = append(out, storage.Entry{CreatedAt: t})
	}
	return out
}

// TestRollingBuckets_DailyAndWeeklyAxes pins the three window schemes SPEC-059
// / DEC-037 lock: --week = 7 daily buckets, --month = 4 weekly buckets,
// --quarter = 13 weekly buckets. Axis start = end - width*n; every bucket is
// zero-filled and exactly n buckets are returned. Table-driven, hermetic.
func TestRollingBuckets_DailyAndWeeklyAxes(t *testing.T) {
	end := time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC)
	day := 24 * time.Hour
	week := 7 * day

	cases := []struct {
		name    string
		width   time.Duration
		n       int
		entries []storage.Entry
		want    []int
	}{
		{
			// --week: 7 daily buckets over [end-7d, end). start = 2026-07-03T12:00Z.
			// Placed: bucket0 x1, bucket2 x2, bucket6 x1; the rest zero-filled.
			name:  "week_daily_zerofilled",
			width: day,
			n:     7,
			entries: entriesAt(
				end.Add(-week),                        // bucket 0 (== start, lower-inclusive)
				end.Add(-week).Add(2*day),             // bucket 2
				end.Add(-week).Add(2*day+5*time.Hour), // bucket 2
				end.Add(-week).Add(6*day),             // bucket 6
			),
			want: []int{1, 0, 2, 0, 0, 0, 1},
		},
		{
			// --month: 4 weekly buckets over [end-28d, end).
			name:  "month_weekly",
			width: week,
			n:     4,
			entries: entriesAt(
				end.Add(-4*week),            // bucket 0
				end.Add(-4*week).Add(3*day), // bucket 0
				end.Add(-2*week),            // bucket 2
				end.Add(-time.Nanosecond),   // bucket 3 (just inside end)
			),
			want: []int{2, 0, 1, 1},
		},
		{
			// --quarter: 13 weekly buckets over [end-91d, end). Smoke: one entry
			// in the first and one in the last bucket, all else zero.
			name:  "quarter_weekly_endpoints",
			width: week,
			n:     13,
			entries: entriesAt(
				end.Add(-13*week),         // bucket 0
				end.Add(-time.Nanosecond), // bucket 12
			),
			want: []int{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		},
		{
			// Empty input → exactly n zero-filled buckets (never nil).
			name:    "empty_zerofilled",
			width:   day,
			n:       7,
			entries: nil,
			want:    []int{0, 0, 0, 0, 0, 0, 0},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := RollingBuckets(c.entries, end, c.width, c.n)
			if len(got) != c.n {
				t.Fatalf("want %d buckets, got %d (%v)", c.n, len(got), got)
			}
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("RollingBuckets = %v, want %v", got, c.want)
			}
		})
	}
}

// TestRollingBuckets_BucketBoundaryAssignment pins the exclusive/inclusive
// boundary rule DEC-037 chose: bucket k covers [start+k*width, start+(k+1)*width)
// — lower-INCLUSIVE, upper-EXCLUSIVE. An entry exactly at end is excluded (the
// axis upper edge is exclusive); an entry before start is excluded.
func TestRollingBuckets_BucketBoundaryAssignment(t *testing.T) {
	end := time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC)
	day := 24 * time.Hour
	start := end.Add(-7 * day) // 2026-07-03T12:00Z

	entries := entriesAt(
		start,                                // -> bucket 0 (lower-inclusive)
		start.Add(day),                       // -> bucket 1 (edge belongs to the later bucket)
		start.Add(day).Add(-time.Nanosecond), // -> bucket 0 (just before the edge)
		start.Add(6*day),                     // -> bucket 6
		end.Add(-time.Nanosecond),            // -> bucket 6 (just inside end)
		start.Add(-time.Nanosecond),          // excluded: before start
		end,                                  // excluded: == end (upper-exclusive)
	)

	want := []int{2, 1, 0, 0, 0, 0, 2} // sum 5
	got := RollingBuckets(entries, end, day, 7)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("boundary assignment: got %v, want %v", got, want)
	}
}

// TestRollingBuckets_FeedsSparkLine is the §12(b) design-time golden: a known
// daily corpus buckets to [1,0,2,3,1,0,2], which spark.Line renders as the
// hand-computed glyph string. Computed by hand and confirmed at design against
// the real spark.Line, so the golden is real, not aspirational. Ties the new
// bucketer to the sparkline primitive end-to-end.
func TestRollingBuckets_FeedsSparkLine(t *testing.T) {
	end := time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC)
	day := 24 * time.Hour
	start := end.Add(-7 * day)

	// Place entries to produce daily counts [1,0,2,3,1,0,2].
	var times []time.Time
	add := func(bucket, count int) {
		for i := 0; i < count; i++ {
			times = append(times, start.Add(time.Duration(bucket)*day).Add(time.Duration(i)*time.Hour))
		}
	}
	add(0, 1)
	add(2, 2)
	add(3, 3)
	add(4, 1)
	add(6, 2)

	counts := RollingBuckets(entriesAt(times...), end, day, 7)
	wantCounts := []int{1, 0, 2, 3, 1, 0, 2}
	if !reflect.DeepEqual(counts, wantCounts) {
		t.Fatalf("counts = %v, want %v", counts, wantCounts)
	}

	const wantGlyphs = "▃▁▆█▃▁▆" // hand-computed, verified via spark.Line at design
	if got := spark.Line(counts); got != wantGlyphs {
		t.Errorf("spark.Line(%v) = %q, want %q", counts, got, wantGlyphs)
	}
}
