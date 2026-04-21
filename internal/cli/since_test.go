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
