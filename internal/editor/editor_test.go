package editor

import (
	"reflect"
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

func TestParse_DuplicateTitleHeaderIsError(t *testing.T) {
	buf := []byte("Title: a\nTitle: b\n\nbody\n")
	f, err := Parse(buf)
	if err == nil {
		t.Fatal("Parse on buffer with duplicate Title: expected error, got nil")
	}
	msg := strings.ToLower(err.Error())
	if !strings.Contains(msg, "title") {
		t.Errorf("error must mention %q; got %q", "title", err.Error())
	}
	if !strings.Contains(msg, "duplicate") {
		t.Errorf("error must mention %q; got %q", "duplicate", err.Error())
	}
	if f != (Fields{}) {
		t.Errorf("Fields must be the zero value on error, got %+v", f)
	}
}

func TestParse_DuplicateImpactHeaderIsError(t *testing.T) {
	buf := []byte("Title: x\nImpact: cut latency 50%\nImpact: shipped nothing\n\nbody\n")
	_, err := Parse(buf)
	if err == nil {
		t.Fatal("Parse on buffer with duplicate Impact: expected error, got nil")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "impact") {
		t.Errorf("error must mention %q; got %q", "impact", err.Error())
	}
}

func TestParse_DuplicateHeaderIsCaseInsensitive(t *testing.T) {
	// textproto canonicalizes Title: and title: to the same key, so the
	// two values collide and must be caught as a duplicate.
	buf := []byte("Title: a\ntitle: b\n\nbody\n")
	_, err := Parse(buf)
	if err == nil {
		t.Fatal("Parse on buffer with case-varied duplicate Title: expected error, got nil")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "title") {
		t.Errorf("error must mention %q; got %q", "title", err.Error())
	}
}

func TestParse_DuplicateUnknownHeaderStillIgnored(t *testing.T) {
	// Pins the duplicate-header guard's scope to the five canonical keys —
	// an unknown header repeated twice must not trip it, preserving the
	// existing "unknown headers silently ignored" contract.
	buf := []byte("Title: x\nX-Note: one\nX-Note: two\n\nbody\n")
	f, err := Parse(buf)
	if err != nil {
		t.Fatalf("Parse: %v (duplicate unknown headers should be ignored, not rejected)", err)
	}
	if f.Title != "x" {
		t.Errorf("Title = %q, want %q", f.Title, "x")
	}
}

// TestParse_DuplicateGuardCoversEveryHeaderRenderEmits gives the duplicate
// guard a floor of its own. canonicalHeaders is a hand-typed slice; the four
// tests above only exercise two of its five entries, so deleting "Tags",
// "Project" or "Type" from it leaves the whole repo's suite green — and the
// silent-drop bug this spec exists to fix comes back on that field. This
// derives the set to test from Render's ACTUAL output rather than re-typing
// the list, so a sixth field added to Render but forgotten in
// canonicalHeaders turns this red instead of shipping unguarded.
func TestParse_DuplicateGuardCoversEveryHeaderRenderEmits(t *testing.T) {
	full := Fields{
		Title:   "t",
		Tags:    "g",
		Project: "p",
		Type:    "y",
		Impact:  "i",
	}
	var names []string
	for _, ln := range strings.Split(string(Render(full)), "\n") {
		if ln == "" {
			break // end of the header block
		}
		k, _, ok := strings.Cut(ln, ": ")
		if !ok {
			t.Fatalf("Render emitted a non-header line %q in the header block", ln)
		}
		names = append(names, k)
	}

	// Non-vacuity floor: Fields is (headers + Description), so the header
	// count is derivable from the struct. Without this, a Render that stopped
	// emitting headers would make the loop below iterate zero times and pass.
	if want := reflect.TypeOf(Fields{}).NumField() - 1; len(names) != want {
		t.Fatalf("Render emitted %d headers %v, want %d (every Fields member except Description)", len(names), names, want)
	}

	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			buf := []byte("Title: t\n" + name + ": one\n" + name + ": two\n\nbody\n")
			if _, err := Parse(buf); err == nil {
				t.Fatalf("Parse tolerated a duplicate %q header — canonicalHeaders is missing it, so that field is still silently droppable", name)
			}
		})
	}
}

// TestTemplates_CannotEmitADuplicateHeader pins the other half of the
// duplicate guard's blast radius. Before this spec a repeated header in a
// shipped template was harmless (first-wins); now it hard-fails every
// invocation of the command that uses it, and nothing else in the suite
// notices: a second "Title:" line in EmptyTemplate, or a second "Impact:" in
// FailureTemplate, both leave `go test ./...` green while breaking
// `brag add` and `brag learn` outright.
func TestTemplates_CannotEmitADuplicateHeader(t *testing.T) {
	for name, tpl := range map[string][]byte{
		"EmptyTemplate":   EmptyTemplate(),
		"FailureTemplate": FailureTemplate(),
	} {
		t.Run(name, func(t *testing.T) {
			// The templates carry empty values, so Parse fails on the
			// required-Title rule. That is expected; what must never appear
			// is the duplicate-header rejection.
			_, err := Parse(tpl)
			if err != nil && strings.Contains(err.Error(), "duplicate") {
				t.Fatalf("%s emits a duplicate header — every %s-backed command fails on an ordinary edit: %v", name, name, err)
			}
			// And it must still be parseable once a Title is supplied.
			filled := []byte(strings.Replace(string(tpl), "Title: \n", "Title: x\n", 1))
			if _, err := Parse(filled); err != nil {
				t.Fatalf("%s with a filled-in Title must parse; got %v", name, err)
			}
		})
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
