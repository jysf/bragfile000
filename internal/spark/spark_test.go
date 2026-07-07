package spark

import "testing"

// TestLine_NormalizationTable is the primitive's load-bearing contract:
// every glyph golden here was computed at design time against the real
// Line algorithm (DEC-031 min→max, math.Round, ▁▂▃▄▅▆▇█).
func TestLine_NormalizationTable(t *testing.T) {
	cases := []struct {
		name string
		in   []int
		want string
	}{
		{"empty", []int{}, ""},
		{"nil", nil, ""},
		{"single", []int{5}, "▁"},
		{"flat-zero", []int{0, 0, 0}, "▁▁▁"},
		{"flat-nonzero", []int{3, 3, 3}, "▁▁▁"},
		{"ramp-0-7", []int{0, 1, 2, 3, 4, 5, 6, 7}, "▁▂▃▄▅▆▇█"},
		{"two-point-min-max", []int{0, 1}, "▁█"},
		{"wrapped-year", []int{1, 1, 0, 2, 0, 0, 2, 0, 0, 0, 1, 0}, "▅▅▁█▁▁█▁▁▁▅▁"},
		{"wrapped-quarter", []int{2, 0, 0}, "█▁▁"},
		{"classic", []int{0, 2, 5, 4, 7, 8, 3, 1}, "▁▃▅▅▇█▄▂"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Line(tc.in)
			if got != tc.want {
				t.Errorf("Line(%v) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

// TestLine_LengthMatchesInput guards against an off-by-one or accidental
// resampling: exactly one glyph per element for any non-empty input.
func TestLine_LengthMatchesInput(t *testing.T) {
	cases := [][]int{
		{5},
		{0, 0, 0},
		{0, 1, 2, 3, 4, 5, 6, 7},
		{1, 1, 0, 2, 0, 0, 2, 0, 0, 0, 1, 0},
		{9, 3, 3, 7},
	}
	for _, in := range cases {
		if got := len([]rune(Line(in))); got != len(in) {
			t.Errorf("len([]rune(Line(%v))) = %d, want %d", in, got, len(in))
		}
	}
}

// TestLine_OnlyBlockGlyphsOrEmpty pins the glyph table: every rune of any
// non-empty output is one of ▁▂▃▄▅▆▇█ — no stray ASCII, no space.
func TestLine_OnlyBlockGlyphsOrEmpty(t *testing.T) {
	allowed := map[rune]bool{
		'▁': true, '▂': true, '▃': true, '▄': true,
		'▅': true, '▆': true, '▇': true, '█': true,
	}
	cases := [][]int{
		{5},
		{0, 0, 0},
		{3, 3, 3},
		{0, 1, 2, 3, 4, 5, 6, 7},
		{0, 1},
		{1, 1, 0, 2, 0, 0, 2, 0, 0, 0, 1, 0},
		{2, 0, 0},
		{0, 2, 5, 4, 7, 8, 3, 1},
	}
	for _, in := range cases {
		for _, r := range Line(in) {
			if !allowed[r] {
				t.Errorf("Line(%v) produced disallowed rune %q", in, r)
			}
		}
	}
}
