package editor

import (
	"strings"
	"testing"
)

func TestRender_MinimalEntry(t *testing.T) {
	got := Render(Fields{Title: "x"})
	want := "Title: x\n\n"
	if string(got) != want {
		t.Fatalf("Render = %q, want %q", string(got), want)
	}
}

func TestRender_FullEntry(t *testing.T) {
	f := Fields{
		Title:       "shipped auth refactor",
		Tags:        "auth,perf",
		Project:     "platform",
		Type:        "shipped",
		Impact:      "unblocked mobile v3",
		Description: "Replaced the join-on-every-request with a redis lookup.\n",
	}
	got := string(Render(f))

	titleIdx := strings.Index(got, "Title: shipped auth refactor\n")
	if titleIdx != 0 {
		t.Fatalf("Title header must be the first line; got start=%q", got)
	}
	tagsIdx := strings.Index(got, "Tags:")
	if tagsIdx < 0 || tagsIdx < titleIdx {
		t.Errorf("Tags header must appear after Title; got tagsIdx=%d", tagsIdx)
	}
	// blank line separator between headers and body.
	blankIdx := strings.Index(got, "\n\n")
	if blankIdx < 0 {
		t.Fatalf("expected blank line separator in output, got %q", got)
	}
	bodyIdx := strings.Index(got, "Replaced the join-on-every-request")
	if bodyIdx < 0 {
		t.Fatalf("expected description body in output, got %q", got)
	}
	if bodyIdx < blankIdx {
		t.Errorf("body must come after blank-line separator; blankIdx=%d bodyIdx=%d", blankIdx, bodyIdx)
	}
}

func TestRender_OmitsEmptyHeaders(t *testing.T) {
	got := string(Render(Fields{Title: "x", Description: "body"}))
	for _, header := range []string{"Tags:", "Project:", "Type:", "Impact:"} {
		if strings.Contains(got, header) {
			t.Errorf("expected output to NOT contain %q (empty field should be omitted), got %q", header, got)
		}
	}
}

func TestParse_HappyPath(t *testing.T) {
	buf := []byte(
		"Title: shipped auth refactor\n" +
			"Tags: auth,perf\n" +
			"Project: platform\n" +
			"Type: shipped\n" +
			"Impact: unblocked mobile v3\n" +
			"\n" +
			"Replaced the join-on-every-request with a redis lookup.\n",
	)
	f, err := Parse(buf)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if f.Title != "shipped auth refactor" {
		t.Errorf("Title = %q, want %q", f.Title, "shipped auth refactor")
	}
	if f.Tags != "auth,perf" {
		t.Errorf("Tags = %q, want %q", f.Tags, "auth,perf")
	}
	if f.Project != "platform" {
		t.Errorf("Project = %q", f.Project)
	}
	if f.Type != "shipped" {
		t.Errorf("Type = %q", f.Type)
	}
	if f.Impact != "unblocked mobile v3" {
		t.Errorf("Impact = %q", f.Impact)
	}
	if !strings.Contains(f.Description, "Replaced the join-on-every-request") {
		t.Errorf("Description = %q; expected to contain body text", f.Description)
	}
}

func TestParse_CaseInsensitiveHeaders(t *testing.T) {
	// net/textproto's ReadMIMEHeader canonicalizes header keys, so
	// TAGS: foo, Tags: foo, and tags: foo all land under the same map
	// entry.
	buf := []byte("Title: x\nTAGS: foo\n\n")
	f, err := Parse(buf)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if f.Tags != "foo" {
		t.Errorf("Tags = %q, want %q (case-insensitive header read)", f.Tags, "foo")
	}
}

func TestParse_MissingTitle(t *testing.T) {
	buf := []byte("Tags: foo\n\nbody\n")
	_, err := Parse(buf)
	if err == nil {
		t.Fatal("Parse on buffer with no Title: expected error, got nil")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "title") {
		t.Errorf("error must mention 'title'; got %q", err.Error())
	}
}

func TestParse_EmptyTitle(t *testing.T) {
	buf := []byte("Title:\n\nbody\n")
	_, err := Parse(buf)
	if err == nil {
		t.Fatal("Parse on buffer with empty Title: expected error, got nil")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "title") {
		t.Errorf("error must mention 'title'; got %q", err.Error())
	}
}

func TestParse_UnknownHeadersIgnored(t *testing.T) {
	buf := []byte("Title: x\nMood: tired\nTags: a\n\nbody\n")
	f, err := Parse(buf)
	if err != nil {
		t.Fatalf("Parse: %v (unknown headers should be ignored silently)", err)
	}
	if f.Title != "x" {
		t.Errorf("Title = %q, want %q", f.Title, "x")
	}
	if f.Tags != "a" {
		t.Errorf("Tags = %q, want %q", f.Tags, "a")
	}
}

func TestParse_MultilineDescription(t *testing.T) {
	body := "First paragraph.\n\nSecond paragraph with blank line above.\n\nThird.\n"
	buf := []byte("Title: x\n\n" + body)
	f, err := Parse(buf)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if f.Description != body {
		t.Errorf("Description mismatch.\n got: %q\nwant: %q", f.Description, body)
	}
}

func TestEmptyTemplate_ContainsAllHeaders(t *testing.T) {
	tpl := string(EmptyTemplate())
	wanted := []string{"Title: \n", "Tags: \n", "Project: \n", "Type: \n", "Impact: \n"}
	prev := -1
	for _, s := range wanted {
		idx := strings.Index(tpl, s)
		if idx < 0 {
			t.Errorf("EmptyTemplate missing header substring %q; got %q", s, tpl)
			continue
		}
		if idx <= prev {
			t.Errorf("header %q must appear after the previous one (idx=%d, prev=%d) in %q", s, idx, prev, tpl)
		}
		prev = idx
	}
}

func TestEmptyTemplate_EndsWithBlankLine(t *testing.T) {
	tpl := string(EmptyTemplate())
	if !strings.HasSuffix(tpl, "Impact: \n\n") {
		t.Errorf("EmptyTemplate must end with %q (header block + blank line); got %q", "Impact: \\n\\n", tpl)
	}
}

func TestEmptyTemplate_ParsesToMissingTitleError(t *testing.T) {
	_, err := Parse(EmptyTemplate())
	if err == nil {
		t.Fatalf("Parse(EmptyTemplate()): expected error, got nil")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "title") {
		t.Errorf("error must mention 'title'; got %q", err.Error())
	}
}

func TestRoundTrip_AllFields(t *testing.T) {
	f := Fields{
		Title:       "shipped auth refactor",
		Tags:        "auth,perf",
		Project:     "platform",
		Type:        "shipped",
		Impact:      "unblocked mobile v3",
		Description: "Replaced the join-on-every-request with a redis lookup.\n",
	}
	got, err := Parse(Render(f))
	if err != nil {
		t.Fatalf("Parse(Render): %v", err)
	}
	if got != f {
		t.Errorf("round-trip mismatch:\n got %+v\nwant %+v", got, f)
	}
}
