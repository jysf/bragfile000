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
// and the omit cases. Rewritten to the DEC-027 arity (session/cost/tokens added
// as trailing params; all empty here so behavior is unchanged from SPEC-040).
func TestStampProvenance(t *testing.T) {
	if got := stampProvenance("perf", "claude-code", "claude-opus-4-8", "", "", ""); got != "perf,agent:claude-code,model:claude-opus-4-8" {
		t.Errorf("both: %q", got)
	}
	if got := stampProvenance("", "claude-code", "", "", "", ""); got != "agent:claude-code" {
		t.Errorf("agent-only: %q", got)
	}
	if got := stampProvenance("a,b", "", "claude-opus-4-8", "", "", ""); got != "a,b,model:claude-opus-4-8" {
		t.Errorf("model-only keeps user tags: %q", got)
	}
	if got := stampProvenance("perf", "", "", "", "", ""); got != "perf" {
		t.Errorf("no provenance → user tags unchanged: %q", got)
	}
	if got := stampProvenance("", "", "", "", "", ""); got != "" {
		t.Errorf("nothing → empty: %q", got)
	}
}

// TestStampProvenance_SeedTags ▲ DEC-027 — session/cost/tokens are appended
// (in that order) after agent:/model:, each omittable; user tags preserved.
// NOTE: this asserts the NEW stampProvenance arity. The existing
// TestStampProvenance is rewritten to the new signature (see premise audit).
func TestStampProvenance_SeedTags(t *testing.T) {
	// signature: stampProvenance(tags, agent, model, session, cost, tokens string)
	if got := stampProvenance("perf", "claude-code", "claude-opus-4-8", "sess-abc", "0.42", "18000"); got !=
		"perf,agent:claude-code,model:claude-opus-4-8,session:sess-abc,cost:0.42,tokens:18000" {
		t.Errorf("all fields: %q", got)
	}
	if got := stampProvenance("", "", "", "sess-abc", "", ""); got != "session:sess-abc" {
		t.Errorf("session-only: %q", got)
	}
	if got := stampProvenance("perf", "claude-code", "", "", "", "18000"); got !=
		"perf,agent:claude-code,tokens:18000" {
		t.Errorf("skip empty middle fields: %q", got)
	}
	if got := stampProvenance("", "", "", "", "", ""); got != "" {
		t.Errorf("nothing → empty: %q", got)
	}
}

// TestNormalizeCost ▲ DEC-027 — non-negative USD decimal string; reject
// non-numeric / negative; trims; empty → ("", ok, no tag).
func TestNormalizeCost(t *testing.T) {
	ok := map[string]string{"0.42": "0.42", "12": "12", "  3.5 ": "3.5", "0": "0"}
	for in, want := range ok {
		got, err := normalizeCost(in)
		if err != nil || got != want {
			t.Errorf("normalizeCost(%q)=%q,%v want %q,nil", in, got, err, want)
		}
	}
	if got, err := normalizeCost(""); err != nil || got != "" {
		t.Errorf("empty cost → (\"\",nil), got %q,%v", got, err)
	}
	for _, bad := range []string{"abc", "-1", "1.2.3", "$5", "1e3"} {
		if _, err := normalizeCost(bad); err == nil {
			t.Errorf("normalizeCost(%q) expected error", bad)
		}
	}
}

// TestNormalizeTokens ▲ DEC-027 — non-negative integer; reject non-integer /
// negative; empty → ("", ok, no tag).
func TestNormalizeTokens(t *testing.T) {
	ok := map[string]string{"18000": "18000", " 0 ": "0", "42": "42"}
	for in, want := range ok {
		got, err := normalizeTokens(in)
		if err != nil || got != want {
			t.Errorf("normalizeTokens(%q)=%q,%v want %q,nil", in, got, err, want)
		}
	}
	if got, err := normalizeTokens(""); err != nil || got != "" {
		t.Errorf("empty tokens → (\"\",nil), got %q,%v", got, err)
	}
	for _, bad := range []string{"abc", "-5", "1.5", "1,000", "0x10"} {
		if _, err := normalizeTokens(bad); err == nil {
			t.Errorf("normalizeTokens(%q) expected error", bad)
		}
	}
}
