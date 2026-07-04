package mcpserver

import "testing"

// TestBuildMatch ▲ mirrors cli.buildFTS5Query / DEC-010: whitespace tokenize,
// phrase-quote each token, join; empty/quote input is an error.
func TestBuildMatch(t *testing.T) {
	ok := map[string]string{
		"auth":          `"auth"`,
		"cut latency":   `"cut" "latency"`,
		"auth-refactor": `"auth-refactor"`,
	}
	for in, want := range ok {
		got, err := buildMatch(in)
		if err != nil || got != want {
			t.Errorf("buildMatch(%q)=%q,%v want %q", in, got, err, want)
		}
	}
	for _, bad := range []string{"", "   ", `has"quote`} {
		if _, err := buildMatch(bad); err == nil {
			t.Errorf("buildMatch(%q) expected error", bad)
		}
	}
}
