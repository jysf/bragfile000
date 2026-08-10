package capture

import (
	"strings"
	"testing"
)

// TestValidate_CapsAcceptedAtBoundary ▲ each field is accepted exactly at its
// byte cap (off-by-one guard). Tags exercises the per-tag boundary: a single
// tag at exactly MaxTagLen bytes (count-cap tests live separately below).
func TestValidate_CapsAcceptedAtBoundary(t *testing.T) {
	f := Fields{
		Title:       strings.Repeat("x", MaxTitle),
		Description: strings.Repeat("d", MaxDescription),
		Tags:        strings.Repeat("t", MaxTagLen),
		Project:     strings.Repeat("p", MaxProject),
		Type:        strings.Repeat("y", MaxType),
		Impact:      strings.Repeat("i", MaxImpact),
	}
	if err := Validate(f); err != nil {
		t.Fatalf("fields at cap should validate, got %v", err)
	}
}

// TestValidate_CapsRejectedOverBoundary ▲ each field one byte over its cap is
// rejected, and the message names the field + limit. The tags row is
// deliberately absent here: LD3 replaced the joined-string cap with a
// per-tag + count model, covered by the dedicated tag tests below (a
// different shape, not a different number — SPEC-075 Outputs).
func TestValidate_CapsRejectedOverBoundary(t *testing.T) {
	cases := []struct {
		name   string
		f      Fields
		substr string
	}{
		{"title", Fields{Title: strings.Repeat("x", MaxTitle+1)}, `"title" exceeds 256-character limit`},
		{"description", Fields{Title: "ok", Description: strings.Repeat("d", MaxDescription+1)}, `"description" exceeds 100000-character limit`},
		{"project", Fields{Title: "ok", Project: strings.Repeat("p", MaxProject+1)}, `"project" exceeds 64-character limit`},
		{"type", Fields{Title: "ok", Type: strings.Repeat("y", MaxType+1)}, `"type" exceeds 64-character limit`},
		{"impact", Fields{Title: "ok", Impact: strings.Repeat("i", MaxImpact+1)}, `"impact" exceeds 1024-character limit`},
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

// TestValidate_UnchangedCapsHoldTheLine ▲ pins description/project/type at
// their EXISTING caps so the uniform-multiplier temptation (widening
// everything because two fields moved) shows up as a red test rather than a
// silent drift. LD3's "not touched" clause.
func TestValidate_UnchangedCapsHoldTheLine(t *testing.T) {
	cases := []struct {
		name string
		mk   func(v string) Fields
		max  int
	}{
		{"description", func(v string) Fields { return Fields{Title: "ok", Description: v} }, MaxDescription},
		{"project", func(v string) Fields { return Fields{Title: "ok", Project: v} }, MaxProject},
		{"type", func(v string) Fields { return Fields{Title: "ok", Type: v} }, MaxType},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := Validate(tc.mk(strings.Repeat("x", tc.max))); err != nil {
				t.Errorf("%s at %d should validate, got %v", tc.name, tc.max, err)
			}
			if err := Validate(tc.mk(strings.Repeat("x", tc.max+1))); err == nil {
				t.Errorf("%s at %d should be rejected", tc.name, tc.max+1)
			}
		})
	}
	if MaxProject != 64 || MaxType != 64 || MaxDescription != 100000 {
		t.Errorf("MaxProject/MaxType/MaxDescription must stay 64/64/100000, got %d/%d/%d", MaxProject, MaxType, MaxDescription)
	}
}

// TestValidate_SingleTagOverLimitNamesTheTag ▲ LD3: a single tag over
// MaxTagLen bytes is rejected, and the error names the offending TAG, not
// the joined string.
func TestValidate_SingleTagOverLimitNamesTheTag(t *testing.T) {
	longTag := strings.Repeat("z", MaxTagLen+1)
	err := Validate(Fields{Title: "ok", Tags: "ok," + longTag})
	if err == nil {
		t.Fatalf("expected error for a tag over %d bytes", MaxTagLen)
	}
	if !strings.Contains(err.Error(), longTag) {
		t.Errorf("error should name the offending tag %q, got %v", longTag, err)
	}
}

// TestValidate_TagCountLimit ▲ LD3: 32 tags pass, 33 are rejected with a
// count-shaped error.
func TestValidate_TagCountLimit(t *testing.T) {
	tags32 := strings.TrimSuffix(strings.Repeat("x,", MaxTagCount), ",")
	if err := Validate(Fields{Title: "ok", Tags: tags32}); err != nil {
		t.Errorf("%d tags should validate, got %v", MaxTagCount, err)
	}

	tags33 := strings.TrimSuffix(strings.Repeat("x,", MaxTagCount+1), ",")
	err := Validate(Fields{Title: "ok", Tags: tags33})
	if err == nil {
		t.Fatalf("expected error for %d tags", MaxTagCount+1)
	}
	if !strings.Contains(err.Error(), "32") {
		t.Errorf("error should be count-shaped (mention the %d limit), got %v", MaxTagCount, err)
	}
}

// TestValidate_NineTagJoinedStringNowPasses ▲ LD3 — the reported regression.
// The exact corpus string from entry 44 (89 bytes, 9 tags, longest tag 16
// bytes) used to be rejected by the old joined-string MaxTags=64 cap. It
// must now pass: this is what "the cap fires on count, never on length" was
// supposed to mean all along.
func TestValidate_NineTagJoinedStringNowPasses(t *testing.T) {
	tags := "rust,cli,spec-driven,rspeed,reqwest,async-trait,public-api,measurement-core,latency-probe"
	if len(tags) != 89 {
		t.Fatalf("test fixture drift: expected 89 bytes, got %d", len(tags))
	}
	if err := Validate(Fields{Title: "ok", Tags: tags}); err != nil {
		t.Errorf("9 short tags joined to 89 bytes should validate, got %v", err)
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
