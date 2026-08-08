package timewindow

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// pdt is a fixed UTC-7 zone standing in for a non-UTC local zone. Using a
// fixed zone (not the host's) keeps every day-boundary assertion independent
// of where the test runs.
var pdt = time.FixedZone("PDT", -7*3600)

// ref is the reference instant every relative-duration case measures from.
// Migrated from internal/cli/since_test.go, where the same cases had to
// compare against time.Now() within a 1-second tolerance; with `now` an
// explicit parameter the assertions become exact.
var ref = time.Date(2026, 7, 6, 8, 0, 0, 0, pdt)

func TestParseSince_ISODate(t *testing.T) {
	got, err := ParseSince("2026-01-01", ref)
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
	// The bare-date path is anchored to UTC midnight (DEC-008) and must not
	// consult `now` at all — a different now yields the same instant.
	other, err := ParseSince("2026-01-01", ref.AddDate(0, 0, 300))
	if err != nil {
		t.Fatalf("ParseSince: %v", err)
	}
	if !other.Equal(want) {
		t.Errorf("bare date consulted now: got %v, want %v", other, want)
	}
}

func TestParseSince_Days(t *testing.T) {
	got, err := ParseSince("7d", ref)
	if err != nil {
		t.Fatalf("ParseSince: %v", err)
	}
	if want := ref.Add(-7 * 24 * time.Hour).UTC(); !got.Equal(want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestParseSince_Weeks(t *testing.T) {
	got, err := ParseSince("2w", ref)
	if err != nil {
		t.Fatalf("ParseSince: %v", err)
	}
	if want := ref.Add(-14 * 24 * time.Hour).UTC(); !got.Equal(want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestParseSince_Months(t *testing.T) {
	got, err := ParseSince("3m", ref)
	if err != nil {
		t.Fatalf("ParseSince: %v", err)
	}
	// A "month" is an approximate 30 days by DEC-008, not a calendar month.
	if want := ref.Add(-90 * 24 * time.Hour).UTC(); !got.Equal(want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestParseSince_NormalizesToUTC(t *testing.T) {
	got, err := ParseSince("1d", ref)
	if err != nil {
		t.Fatalf("ParseSince: %v", err)
	}
	if loc := got.Location().String(); loc != "UTC" {
		t.Errorf("Location = %q, want UTC (a relative since is stored-comparable)", loc)
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
			if _, err := ParseSince(in, ref); err == nil {
				t.Errorf("ParseSince(%q): expected error, got nil", in)
			}
		})
	}
}

// TestParseSince_IsPureFunctionOfNow pins DEC-042 sub-decision 1: the clock
// seam belongs to the CALLER, so shifting `now` shifts the result by exactly
// the same delta. If the parser ever reached for time.Now() internally, the
// two results would not track.
func TestParseSince_IsPureFunctionOfNow(t *testing.T) {
	const shift = 36 * time.Hour
	a, err := ParseSince("7d", ref)
	if err != nil {
		t.Fatalf("ParseSince: %v", err)
	}
	b, err := ParseSince("7d", ref.Add(shift))
	if err != nil {
		t.Fatalf("ParseSince: %v", err)
	}
	if got := b.Sub(a); got != shift {
		t.Errorf("result delta = %v, want %v (parser is not a pure function of now)", got, shift)
	}
}

func TestParseDay_ExplicitDateIsLocalMidnightHalfOpen(t *testing.T) {
	// The zone a day resolves in comes from now.Location().
	now := time.Date(2026, 7, 5, 12, 0, 0, 0, pdt)

	start, end, err := ParseDay("2026-07-05", now)
	if err != nil {
		t.Fatalf("ParseDay: %v", err)
	}
	wantStart := time.Date(2026, 7, 5, 0, 0, 0, 0, pdt)
	wantEnd := time.Date(2026, 7, 6, 0, 0, 0, 0, pdt)
	if !start.Equal(wantStart) {
		t.Errorf("start: got %v, want %v", start, wantStart)
	}
	if !end.Equal(wantEnd) {
		t.Errorf("end: got %v, want %v", end, wantEnd)
	}
	if start.Location() != pdt || end.Location() != pdt {
		t.Errorf("bounds must carry now's location: start %v, end %v", start.Location(), end.Location())
	}
}

func TestParseDay_TodayYesterdayResolveAgainstNow(t *testing.T) {
	now := time.Date(2026, 7, 6, 8, 0, 0, 0, pdt)

	todayStart, _, err := ParseDay("today", now)
	if err != nil {
		t.Fatalf("ParseDay(today): %v", err)
	}
	if want := time.Date(2026, 7, 6, 0, 0, 0, 0, pdt); !todayStart.Equal(want) {
		t.Errorf("today start: got %v, want %v", todayStart, want)
	}

	yStart, yEnd, err := ParseDay("yesterday", now)
	if err != nil {
		t.Fatalf("ParseDay(yesterday): %v", err)
	}
	if want := time.Date(2026, 7, 5, 0, 0, 0, 0, pdt); !yStart.Equal(want) {
		t.Errorf("yesterday start: got %v, want %v", yStart, want)
	}
	if want := time.Date(2026, 7, 6, 0, 0, 0, 0, pdt); !yEnd.Equal(want) {
		t.Errorf("yesterday end: got %v, want %v", yEnd, want)
	}
}

func TestParseDay_KeywordsAreCaseInsensitive(t *testing.T) {
	now := time.Date(2026, 7, 6, 8, 0, 0, 0, pdt)

	lower, _, err := ParseDay("today", now)
	if err != nil {
		t.Fatalf("ParseDay(today): %v", err)
	}
	upper, _, err := ParseDay("TODAY", now)
	if err != nil {
		t.Fatalf("ParseDay(TODAY): %v", err)
	}
	if !lower.Equal(upper) {
		t.Errorf("case-insensitive keyword mismatch: %v vs %v", lower, upper)
	}
}

func TestParseDay_InvalidValue(t *testing.T) {
	now := time.Date(2026, 7, 6, 8, 0, 0, 0, pdt)

	for _, in := range []string{"", "notaday", "7d", "2026-13-01", " today ", "tomorrow", "07/05/2026"} {
		t.Run(in, func(t *testing.T) {
			_, _, err := ParseDay(in, now)
			if err == nil {
				t.Fatalf("ParseDay(%q): expected error, got nil", in)
			}
			// The message must name the accepted forms so a caller (human or
			// agent) can correct without reading the source.
			msg := err.Error()
			for _, form := range []string{"YYYY-MM-DD", "today", "yesterday"} {
				if !strings.Contains(msg, form) {
					t.Errorf("error %q does not name accepted form %q", msg, form)
				}
			}
		})
	}
}

// TestPackageReadsNoWallClock is the mechanical guard behind DEC-042's
// "internal/timewindow holds no clock": the shared package must never call
// time.Now(), because two callers (cli, mcpserver) own two different clock
// policies and a package-level seam would put them in conflict. Same idiom as
// internal/mcpserver/import_audit_test.go.
func TestPackageReadsNoWallClock(t *testing.T) {
	err := filepath.WalkDir(".", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if strings.Contains(string(b), "time.Now(") {
			t.Errorf("%s calls time.Now(); the clock seam belongs to the caller (DEC-042)", path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
