package cli

import (
	"testing"
	"time"
)

func TestParseSince_ISODate(t *testing.T) {
	got, err := ParseSince("2026-01-01")
	if err != nil {
		t.Fatalf("ParseSince: %v", err)
	}
	want := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("got %v, want %v", got, want)
	}
	if loc := got.Location().String(); loc != "UTC" {
		t.Errorf("Location = %q, want UTC", loc)
	}
}

func TestParseSince_Days(t *testing.T) {
	got, err := ParseSince("7d")
	if err != nil {
		t.Fatalf("ParseSince: %v", err)
	}
	want := time.Now().UTC().Add(-7 * 24 * time.Hour)
	delta := got.Sub(want)
	if delta < 0 {
		delta = -delta
	}
	if delta >= time.Second {
		t.Errorf("got %v, want within 1s of %v (delta %v)", got, want, delta)
	}
}

func TestParseSince_Weeks(t *testing.T) {
	got, err := ParseSince("2w")
	if err != nil {
		t.Fatalf("ParseSince: %v", err)
	}
	want := time.Now().UTC().Add(-14 * 24 * time.Hour)
	delta := got.Sub(want)
	if delta < 0 {
		delta = -delta
	}
	if delta >= time.Second {
		t.Errorf("got %v, want within 1s of %v (delta %v)", got, want, delta)
	}
}

func TestParseSince_Months(t *testing.T) {
	got, err := ParseSince("3m")
	if err != nil {
		t.Fatalf("ParseSince: %v", err)
	}
	want := time.Now().UTC().Add(-90 * 24 * time.Hour)
	delta := got.Sub(want)
	if delta < 0 {
		delta = -delta
	}
	if delta >= time.Second {
		t.Errorf("got %v, want within 1s of %v (delta %v)", got, want, delta)
	}
}

func TestParseSince_InvalidFormat(t *testing.T) {
	cases := []string{
		"",
		"7",
		"d",
		"0d",
		"-3d",
		"abc",
		"7x",
		"2026-13-01",
		"  7d  ",
	}
	for _, in := range cases {
		t.Run(in, func(t *testing.T) {
			if _, err := ParseSince(in); err == nil {
				t.Errorf("ParseSince(%q): expected error, got nil", in)
			}
		})
	}
}

// withStubbedClock swaps the injectable local-wall-clock seam and restores it.
func withStubbedClock(t *testing.T, now time.Time) {
	t.Helper()
	prev := clock
	clock = func() time.Time { return now }
	t.Cleanup(func() { clock = prev })
}

func TestParseDay_ExplicitDateIsLocalMidnightHalfOpen(t *testing.T) {
	// The zone --day resolves in comes from clock().Location().
	loc := time.FixedZone("PDT", -7*3600)
	withStubbedClock(t, time.Date(2026, 7, 5, 12, 0, 0, 0, loc))

	start, end, err := ParseDay("2026-07-05")
	if err != nil {
		t.Fatalf("ParseDay: %v", err)
	}
	wantStart := time.Date(2026, 7, 5, 0, 0, 0, 0, loc)
	wantEnd := time.Date(2026, 7, 6, 0, 0, 0, 0, loc)
	if !start.Equal(wantStart) {
		t.Errorf("start: got %v, want %v", start, wantStart)
	}
	if !end.Equal(wantEnd) {
		t.Errorf("end: got %v, want %v", end, wantEnd)
	}
}

func TestParseDay_TodayYesterdayResolveAgainstClock(t *testing.T) {
	loc := time.FixedZone("PDT", -7*3600)
	withStubbedClock(t, time.Date(2026, 7, 6, 8, 0, 0, 0, loc))

	todayStart, _, err := ParseDay("today")
	if err != nil {
		t.Fatalf("ParseDay(today): %v", err)
	}
	if want := time.Date(2026, 7, 6, 0, 0, 0, 0, loc); !todayStart.Equal(want) {
		t.Errorf("today start: got %v, want %v", todayStart, want)
	}

	yStart, yEnd, err := ParseDay("yesterday")
	if err != nil {
		t.Fatalf("ParseDay(yesterday): %v", err)
	}
	if want := time.Date(2026, 7, 5, 0, 0, 0, 0, loc); !yStart.Equal(want) {
		t.Errorf("yesterday start: got %v, want %v", yStart, want)
	}
	if want := time.Date(2026, 7, 6, 0, 0, 0, 0, loc); !yEnd.Equal(want) {
		t.Errorf("yesterday end: got %v, want %v", yEnd, want)
	}
}

func TestParseDay_KeywordsAreCaseInsensitive(t *testing.T) {
	loc := time.FixedZone("PDT", -7*3600)
	withStubbedClock(t, time.Date(2026, 7, 6, 8, 0, 0, 0, loc))

	lower, _, err := ParseDay("today")
	if err != nil {
		t.Fatalf("ParseDay(today): %v", err)
	}
	upper, _, err := ParseDay("TODAY")
	if err != nil {
		t.Fatalf("ParseDay(TODAY): %v", err)
	}
	if !lower.Equal(upper) {
		t.Errorf("case-insensitive keyword mismatch: %v vs %v", lower, upper)
	}
}

func TestParseDay_InvalidValue(t *testing.T) {
	loc := time.FixedZone("PDT", -7*3600)
	withStubbedClock(t, time.Date(2026, 7, 6, 8, 0, 0, 0, loc))

	for _, in := range []string{"", "notaday", "7d", "2026-13-01", " today "} {
		t.Run(in, func(t *testing.T) {
			if _, _, err := ParseDay(in); err == nil {
				t.Errorf("ParseDay(%q): expected error, got nil", in)
			}
		})
	}
}
