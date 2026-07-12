package capture

import (
	"strings"
	"testing"
)

// TestValidate_CapsAcceptedAtBoundary ▲ each field is accepted exactly at its
// byte cap (off-by-one guard).
func TestValidate_CapsAcceptedAtBoundary(t *testing.T) {
	f := Fields{
		Title:       strings.Repeat("x", MaxTitle),
		Description: strings.Repeat("d", MaxDescription),
		Tags:        strings.Repeat("t", MaxTags),
		Project:     strings.Repeat("p", MaxProject),
		Type:        strings.Repeat("y", MaxType),
		Impact:      strings.Repeat("i", MaxImpact),
	}
	if err := Validate(f); err != nil {
		t.Fatalf("fields at cap should validate, got %v", err)
	}
}

// TestValidate_CapsRejectedOverBoundary ▲ each field one byte over its cap is
// rejected, and the message names the field + limit.
func TestValidate_CapsRejectedOverBoundary(t *testing.T) {
	cases := []struct {
		name   string
		f      Fields
		substr string
	}{
		{"title", Fields{Title: strings.Repeat("x", MaxTitle+1)}, `"title" exceeds 200-character limit`},
		{"description", Fields{Title: "ok", Description: strings.Repeat("d", MaxDescription+1)}, `"description" exceeds 100000-character limit`},
		{"tags", Fields{Title: "ok", Tags: strings.Repeat("t", MaxTags+1)}, `"tags" exceeds 64-character limit`},
		{"project", Fields{Title: "ok", Project: strings.Repeat("p", MaxProject+1)}, `"project" exceeds 64-character limit`},
		{"type", Fields{Title: "ok", Type: strings.Repeat("y", MaxType+1)}, `"type" exceeds 64-character limit`},
		{"impact", Fields{Title: "ok", Impact: strings.Repeat("i", MaxImpact+1)}, `"impact" exceeds 256-character limit`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := Validate(tc.f)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tc.substr) {
				t.Errorf("error %q missing %q", err.Error(), tc.substr)
			}
		})
	}
}

// TestValidate_ControlCharsInSingleLineFields ▲ NUL, tab, newline, and CR are
// each rejected in every single-line field.
func TestValidate_ControlCharsInSingleLineFields(t *testing.T) {
	ctrls := map[string]string{"nul": "a\x00b", "tab": "a\tb", "newline": "a\nb", "cr": "a\rb"}
	build := map[string]func(v string) Fields{
		"title":   func(v string) Fields { return Fields{Title: v} },
		"tags":    func(v string) Fields { return Fields{Title: "ok", Tags: v} },
		"project": func(v string) Fields { return Fields{Title: "ok", Project: v} },
		"type":    func(v string) Fields { return Fields{Title: "ok", Type: v} },
		"impact":  func(v string) Fields { return Fields{Title: "ok", Impact: v} },
	}
	for field, mk := range build {
		for cname, v := range ctrls {
			t.Run(field+"/"+cname, func(t *testing.T) {
				if err := Validate(mk(v)); err == nil {
					t.Errorf("control char %q in %s should be rejected", cname, field)
				}
			})
		}
	}
}

// TestValidate_DescriptionMultiline ▲ description allows tab/newline but
// rejects NUL.
func TestValidate_DescriptionMultiline(t *testing.T) {
	if err := Validate(Fields{Title: "ok", Description: "line1\n\tline2"}); err != nil {
		t.Errorf("newline+tab in description should be allowed, got %v", err)
	}
	if err := Validate(Fields{Title: "ok", Description: "bad\x00body"}); err == nil {
		t.Errorf("NUL in description should be rejected")
	}
}

// TestValidate_ReservedNumericTags ▲ cost:/tokens: tokens in freeform tags are
// validated; agent:/model:/session: are left alone.
func TestValidate_ReservedNumericTags(t *testing.T) {
	bad := []string{"cost:-9", "tokens:xyz", "cost:$5", "perf,cost:-1,shipped", "tokens:1.5"}
	for _, tags := range bad {
		if err := Validate(Fields{Title: "ok", Tags: tags}); err == nil {
			t.Errorf("tags %q should be rejected", tags)
		}
	}
	good := []string{"cost:12.50", "tokens:18000", "cost:0,tokens:0", "agent:x,model:y,session:z", "perf,cost:3.5,agent:manual"}
	for _, tags := range good {
		if err := Validate(Fields{Title: "ok", Tags: tags}); err != nil {
			t.Errorf("tags %q should be accepted, got %v", tags, err)
		}
	}
}

// TestNormalizeCost ▲ DEC-027 — non-negative USD decimal string; reject
// non-numeric / negative; trims; empty → ("", nil, no tag). Moved from
// internal/mcpserver/provenance_test.go when the normalizers moved here.
func TestNormalizeCost(t *testing.T) {
	ok := map[string]string{"0.42": "0.42", "12": "12", "  3.5 ": "3.5", "0": "0"}
	for in, want := range ok {
		got, err := NormalizeCost(in)
		if err != nil || got != want {
			t.Errorf("NormalizeCost(%q)=%q,%v want %q,nil", in, got, err, want)
		}
	}
	if got, err := NormalizeCost(""); err != nil || got != "" {
		t.Errorf("empty cost → (\"\",nil), got %q,%v", got, err)
	}
	for _, bad := range []string{"abc", "-1", "1.2.3", "$5", "1e3"} {
		if _, err := NormalizeCost(bad); err == nil {
			t.Errorf("NormalizeCost(%q) expected error", bad)
		}
	}
}

// TestNormalizeTokens ▲ DEC-027 — non-negative integer; reject non-integer /
// negative; empty → ("", nil, no tag). Moved from
// internal/mcpserver/provenance_test.go when the normalizers moved here.
func TestNormalizeTokens(t *testing.T) {
	ok := map[string]string{"18000": "18000", " 0 ": "0", "42": "42"}
	for in, want := range ok {
		got, err := NormalizeTokens(in)
		if err != nil || got != want {
			t.Errorf("NormalizeTokens(%q)=%q,%v want %q,nil", in, got, err, want)
		}
	}
	if got, err := NormalizeTokens(""); err != nil || got != "" {
		t.Errorf("empty tokens → (\"\",nil), got %q,%v", got, err)
	}
	for _, bad := range []string{"abc", "-5", "1.5", "1,000", "0x10"} {
		if _, err := NormalizeTokens(bad); err == nil {
			t.Errorf("NormalizeTokens(%q) expected error", bad)
		}
	}
}
