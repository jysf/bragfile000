package story

import (
	"encoding/json"
	"strings"
	"testing"
)

// Test E1 — TestToStory_EmptyDirectiveOmitsSection (closes the AC-8 gap
// SPEC-049 verify flagged). Targets bundle.go's `if opts.Directive != ""`
// omission branch on a NON-empty corpus (distinct from Test 4's empty-
// *corpus*, directive-renders case): a non-empty body with an empty
// *directive* must omit the ## Framing directive section (markdown) and
// render framing_directive as the empty string (JSON).
func TestToStory_EmptyDirectiveOmitsSection(t *testing.T) {
	threads := BuildThreads(storyFixture, meThreadOpts)
	opts := StoryOptions{
		Audience:        "me",
		Scope:           "year",
		Filters:         "(none)",
		FiltersJSON:     nil,
		EntriesInWindow: 6,
		Now:             storyFixedNow,
		Threads:         threads,
		Throughline:     BuildThroughline(threads),
		Directive:       "", // <- the omission trigger
	}

	md, err := ToStoryMarkdown(opts)
	if err != nil {
		t.Fatalf("markdown: %v", err)
	}
	// The ## Framing directive section is OMITTED entirely.
	if strings.Contains(string(md), "## Framing directive") {
		t.Errorf("empty directive must omit the ## Framing directive section:\n%s", md)
	}
	// But the rest renders: header, provenance, threads, throughline.
	for _, want := range []string{
		"# Bragfile Story", "Threads: 4", "Beats: 6/6",
		"## Threads", "### alpha", "## Throughline (skeleton)",
	} {
		if !strings.Contains(string(md), want) {
			t.Errorf("empty-directive bundle missing %q:\n%s", want, md)
		}
	}
	// The document must not end with a dangling blank "## " block or a
	// trailing directive artifact — the throughline is the final section.
	if !strings.HasSuffix(strings.TrimRight(string(md), "\n"),
		"(no project) [initiative]: 1 beat, 0 with impact (2026-06-01 → 2026-06-01)") {
		t.Errorf("empty-directive bundle should end at the throughline, got:\n%s", md)
	}

	// JSON: framing_directive is the empty string (not omitted key, not null).
	jsonBody, err := ToStoryJSON(opts)
	if err != nil {
		t.Fatalf("json: %v", err)
	}
	var env struct {
		FramingDirective *string `json:"framing_directive"`
	}
	if err := json.Unmarshal(jsonBody, &env); err != nil {
		t.Fatalf("unmarshal: %v\n%s", err, jsonBody)
	}
	if env.FramingDirective == nil {
		t.Fatalf("framing_directive key must be present (as \"\"), got null/absent:\n%s", jsonBody)
	}
	if *env.FramingDirective != "" {
		t.Errorf("framing_directive: got %q, want empty string", *env.FramingDirective)
	}
	// The literal empty-string form is present (not "null").
	if !strings.Contains(string(jsonBody), `"framing_directive": ""`) {
		t.Errorf("expected framing_directive empty-string literal:\n%s", jsonBody)
	}
}
