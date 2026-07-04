package mcpserver

import "testing"

// TestReservedTag_Normalization ▲ locks the literal: lowercase, whitespace→'-',
// commas stripped; empty-after-normalization yields "".
func TestReservedTag_Normalization(t *testing.T) {
	cases := []struct{ prefix, in, want string }{
		{"agent", "claude-code", "agent:claude-code"},
		{"model", "claude-opus-4-8", "model:claude-opus-4-8"},
		{"agent", "Claude Code", "agent:claude-code"}, // space→'-', lowercased
		{"model", "  GPT 5  ", "model:gpt-5"},         // trim + space→'-'
		{"agent", "a,b", "agent:ab"},                  // comma stripped (DEC-004 model)
		{"agent", "", ""},                             // empty → no tag
		{"model", "   ", ""},                          // whitespace-only → no tag
	}
	for _, c := range cases {
		if got := reservedTag(c.prefix, c.in); got != c.want {
			t.Errorf("reservedTag(%q,%q)=%q want %q", c.prefix, c.in, got, c.want)
		}
	}
}

// TestStampProvenance ▲ locks append order (user tags, then agent:, then model:)
// and the omit cases.
func TestStampProvenance(t *testing.T) {
	if got := stampProvenance("perf", "claude-code", "claude-opus-4-8"); got != "perf,agent:claude-code,model:claude-opus-4-8" {
		t.Errorf("both: %q", got)
	}
	if got := stampProvenance("", "claude-code", ""); got != "agent:claude-code" {
		t.Errorf("agent-only: %q", got)
	}
	if got := stampProvenance("a,b", "", "claude-opus-4-8"); got != "a,b,model:claude-opus-4-8" {
		t.Errorf("model-only keeps user tags: %q", got)
	}
	if got := stampProvenance("perf", "", ""); got != "perf" {
		t.Errorf("no provenance → user tags unchanged: %q", got)
	}
	if got := stampProvenance("", "", ""); got != "" {
		t.Errorf("nothing → empty: %q", got)
	}
}
