// Package spark renders a numeric series as a fixed-width Unicode
// block-glyph sparkline (▁▂▃▄▅▆▇█). It is pure (no I/O, no clock, no
// host state), stdlib-only (math), and dependency-free — the local-first
// visual primitive DEC-031 locks for brag wrapped's cadence line.
package spark

import "math"

// levels is the 8-rung block-glyph ladder, lowest to highest.
var levels = []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

// Line renders vals as a fixed-width Unicode block-glyph sparkline (one
// glyph per element, no resampling) using min→max linear normalization:
// level = round((v-min)/(max-min)*7) over ▁▂▃▄▅▆▇█ (DEC-031). An empty
// series renders ""; a flat series (max==min — includes all-zero and a
// single element) renders all ▁ (min→max has no variation to encode, so
// every element sits at the floor). Pure: no I/O, no clock, no host state.
func Line(vals []int) string {
	if len(vals) == 0 {
		return ""
	}
	min, max := vals[0], vals[0]
	for _, v := range vals {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	out := make([]rune, len(vals))
	if max == min {
		for i := range out {
			out[i] = levels[0]
		}
		return string(out)
	}
	span := float64(max - min)
	for i, v := range vals {
		lvl := int(math.Round(float64(v-min) / span * 7))
		out[i] = levels[lvl]
	}
	return string(out)
}
