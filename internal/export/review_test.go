package export

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/jysf/bragfile000/internal/storage"
)

// reviewFixture is the load-bearing fixture shared across all four
// review renderer tests. 5 entries spanning 3 projects + (no project),
// chrono-ordering chosen to exercise within-alpha chrono-ASC + (no
// project) forced last + JSON description inclusion vs markdown
// description elision.
var reviewFixture = []storage.Entry{
	{
		ID: 1, Title: "alpha-old",
		Description: "did the auth refactor",
		Tags:        "auth", Project: "alpha", Type: "shipped",
		Impact:    "unblocked mobile",
		CreatedAt: time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC),
	},
	{
		ID: 2, Title: "beta-mid",
		Project: "beta", Type: "learned",
		CreatedAt: time.Date(2026, 4, 21, 10, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 4, 21, 10, 0, 0, 0, time.UTC),
	},
	{
		ID: 3, Title: "unbound-mid",
		Type:      "shipped", // (no project) group
		CreatedAt: time.Date(2026, 4, 22, 10, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 4, 22, 10, 0, 0, 0, time.UTC),
	},
	{
		ID: 4, Title: "alpha-new",
		Project: "alpha", Type: "shipped",
		CreatedAt: time.Date(2026, 4, 23, 10, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 4, 23, 10, 0, 0, 0, time.UTC),
	},
	{
		ID: 5, Title: "gamma-only",
		Project: "gamma", Type: "fixed",
		CreatedAt: time.Date(2026, 4, 24, 10, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 4, 24, 10, 0, 0, 0, time.UTC),
	},
}

var reviewFixedNow = time.Date(2026, 4, 25, 12, 0, 0, 0, time.UTC)

// TestToReviewMarkdown_DEC014FullDocumentGolden — LOAD-BEARING; written
// FIRST per SPEC-014/015/018 ship lessons. Locks every DEC-014 markdown
// choice on the shared fixture in one byte-exact assertion. Also locks
// the description-elision half of locked decision §3 (markdown elides;
// JSON includes).
func TestToReviewMarkdown_DEC014FullDocumentGolden(t *testing.T) {
	opts := ReviewOptions{
		Scope: "week",
		Now:   reviewFixedNow,
	}

	want := `# Bragfile Review

Generated: 2026-04-25T12:00:00Z
Scope: week
Filters: (none)

## Entries

### alpha

- 1: alpha-old
- 4: alpha-new

### beta

- 2: beta-mid

### gamma

- 5: gamma-only

### (no project)

- 3: unbound-mid

## Reflection questions

1. What pattern do you see in this period?
2. What did you underestimate?
3. What's missing here that should be?`

	got, err := ToReviewMarkdown(reviewFixture, opts)
	if err != nil {
		t.Fatalf("ToReviewMarkdown: %v", err)
	}
	if !bytes.Equal(got, []byte(want)) {
		t.Fatalf("DEC-014 review markdown golden mismatch\nwant:\n%s\n\ngot:\n%s", want, string(got))
	}

	// Line-based heading-level checks per SPEC-015 substring-trap addendum.
	lines := strings.Split(string(got), "\n")
	if lines[0] != "# Bragfile Review" {
		t.Errorf("expected first line %q, got %q", "# Bragfile Review", lines[0])
	}
	foundEntries := false
	foundReflection := false
	for _, ln := range lines {
		if ln == "## Entries" {
			foundEntries = true
		}
		if ln == "## Reflection questions" {
			foundReflection = true
		}
	}
	if !foundEntries {
		t.Errorf("expected a standalone %q line", "## Entries")
	}
	if !foundReflection {
		t.Errorf("expected a standalone %q line", "## Reflection questions")
	}

	// Elision lock: entry 1's Description must NOT appear in markdown bytes.
	if bytes.Contains(got, []byte("did the auth refactor")) {
		t.Errorf("markdown must elide entry descriptions; found %q in:\n%s",
			"did the auth refactor", string(got))
	}
}

// TestToReviewJSON_DEC014ShapeGolden — LOAD-BEARING; written SECOND.
// Locks every DEC-014 JSON choice on the shared fixture: single-object
// envelope, top-level flat keys in struct-tag order, filters always
// {}, entries_grouped array-of-objects with full DEC-011 9-key entry
// shape inside, reflection_questions array of three strings.
func TestToReviewJSON_DEC014ShapeGolden(t *testing.T) {
	opts := ReviewOptions{
		Scope: "week",
		Now:   reviewFixedNow,
	}

	want := `{
  "generated_at": "2026-04-25T12:00:00Z",
  "scope": "week",
  "filters": {},
  "entries_grouped": [
    {
      "project": "alpha",
      "entries": [
        {
          "id": 1,
          "title": "alpha-old",
          "description": "did the auth refactor",
          "tags": "auth",
          "project": "alpha",
          "type": "shipped",
          "impact": "unblocked mobile",
          "created_at": "2026-04-20T10:00:00Z",
          "updated_at": "2026-04-20T10:00:00Z"
        },
        {
          "id": 4,
          "title": "alpha-new",
          "description": "",
          "tags": "",
          "project": "alpha",
          "type": "shipped",
          "impact": "",
          "created_at": "2026-04-23T10:00:00Z",
          "updated_at": "2026-04-23T10:00:00Z"
        }
      ]
    },
    {
      "project": "beta",
      "entries": [
        {
          "id": 2,
          "title": "beta-mid",
          "description": "",
          "tags": "",
          "project": "beta",
          "type": "learned",
          "impact": "",
          "created_at": "2026-04-21T10:00:00Z",
          "updated_at": "2026-04-21T10:00:00Z"
        }
      ]
    },
    {
      "project": "gamma",
      "entries": [
        {
          "id": 5,
          "title": "gamma-only",
          "description": "",
          "tags": "",
          "project": "gamma",
          "type": "fixed",
          "impact": "",
          "created_at": "2026-04-24T10:00:00Z",
          "updated_at": "2026-04-24T10:00:00Z"
        }
      ]
    },
    {
      "project": "(no project)",
      "entries": [
        {
          "id": 3,
          "title": "unbound-mid",
          "description": "",
          "tags": "",
          "project": "",
          "type": "shipped",
          "impact": "",
          "created_at": "2026-04-22T10:00:00Z",
          "updated_at": "2026-04-22T10:00:00Z"
        }
      ]
    }
  ],
  "reflection_questions": [
    "What pattern do you see in this period?",
    "What did you underestimate?",
    "What's missing here that should be?"
  ]
}`

	got, err := ToReviewJSON(reviewFixture, opts)
	if err != nil {
		t.Fatalf("ToReviewJSON: %v", err)
	}
	if !bytes.Equal(got, []byte(want)) {
		t.Fatalf("DEC-014 review JSON golden mismatch\nwant:\n%s\n\ngot:\n%s", want, string(got))
	}

	// Verify struct-tag declaration order on top-level keys.
	dec := json.NewDecoder(bytes.NewReader(got))
	tok, err := dec.Token()
	if err != nil {
		t.Fatalf("decoder.Token open: %v", err)
	}
	if d, ok := tok.(json.Delim); !ok || d != '{' {
		t.Fatalf("expected opening {, got %v", tok)
	}
	wantKeys := []string{"generated_at", "scope", "filters", "entries_grouped", "reflection_questions"}
	for _, k := range wantKeys {
		tok, err := dec.Token()
		if err != nil {
			t.Fatalf("decoder.Token key %q: %v", k, err)
		}
		gotKey, ok := tok.(string)
		if !ok {
			t.Fatalf("expected string key %q, got %T(%v)", k, tok, tok)
		}
		if gotKey != k {
			t.Fatalf("expected key %q, got %q", k, gotKey)
		}
		if err := skipValue(dec); err != nil {
			t.Fatalf("skip value for %q: %v", k, err)
		}
	}

	// Inclusion lock: entry 1's Description must appear in the JSON body.
	var m map[string]any
	if err := json.Unmarshal(got, &m); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	groups, ok := m["entries_grouped"].([]any)
	if !ok || len(groups) == 0 {
		t.Fatalf("expected entries_grouped non-empty, got %T(%v)", m["entries_grouped"], m["entries_grouped"])
	}
	g0, ok := groups[0].(map[string]any)
	if !ok {
		t.Fatalf("expected entries_grouped[0] as object, got %T", groups[0])
	}
	gEntries, ok := g0["entries"].([]any)
	if !ok || len(gEntries) == 0 {
		t.Fatalf("expected entries_grouped[0].entries non-empty, got %T(%v)", g0["entries"], g0["entries"])
	}
	e0, ok := gEntries[0].(map[string]any)
	if !ok {
		t.Fatalf("expected first entry as object, got %T", gEntries[0])
	}
	if e0["description"] != "did the auth refactor" {
		t.Errorf("JSON must include full DEC-011 entry shape with descriptions; got description=%v", e0["description"])
	}
}

// TestToReview_EmptyEntriesStillEmitsReflectionQuestions locks the
// distinguishing contract: review's reflection-questions block ALWAYS
// renders, even on empty entries. Diverges from SPEC-018's
// TestToSummary_EmptyEntriesEmitsProvenanceOnly because review's
// payload has a non-entry-derived part (questions) that doesn't elide.
func TestToReview_EmptyEntriesStillEmitsReflectionQuestions(t *testing.T) {
	opts := ReviewOptions{
		Scope: "week",
		Now:   reviewFixedNow,
	}

	t.Run("markdown", func(t *testing.T) {
		want := `# Bragfile Review

Generated: 2026-04-25T12:00:00Z
Scope: week
Filters: (none)

## Reflection questions

1. What pattern do you see in this period?
2. What did you underestimate?
3. What's missing here that should be?`
		got, err := ToReviewMarkdown([]storage.Entry{}, opts)
		if err != nil {
			t.Fatalf("ToReviewMarkdown: %v", err)
		}
		if !bytes.Equal(got, []byte(want)) {
			t.Fatalf("empty-state review markdown mismatch\nwant:\n%s\ngot:\n%s", want, string(got))
		}
		// Line-based no-Entries-section assertion + Reflection-present.
		foundEntries := false
		foundReflection := false
		for _, ln := range strings.Split(string(got), "\n") {
			if ln == "## Entries" {
				foundEntries = true
			}
			if ln == "## Reflection questions" {
				foundReflection = true
			}
		}
		if foundEntries {
			t.Errorf("Entries heading must NOT appear when entries is empty")
		}
		if !foundReflection {
			t.Errorf("Reflection questions heading must appear even when entries is empty")
		}
	})

	t.Run("json", func(t *testing.T) {
		want := `{
  "generated_at": "2026-04-25T12:00:00Z",
  "scope": "week",
  "filters": {},
  "entries_grouped": [],
  "reflection_questions": [
    "What pattern do you see in this period?",
    "What did you underestimate?",
    "What's missing here that should be?"
  ]
}`
		got, err := ToReviewJSON([]storage.Entry{}, opts)
		if err != nil {
			t.Fatalf("ToReviewJSON: %v", err)
		}
		if !bytes.Equal(got, []byte(want)) {
			t.Fatalf("empty-state review JSON mismatch\nwant:\n%s\ngot:\n%s", want, string(got))
		}
		var m map[string]any
		if err := json.Unmarshal(got, &m); err != nil {
			t.Fatalf("json.Unmarshal: %v", err)
		}
		eg, ok := m["entries_grouped"].([]any)
		if !ok {
			t.Fatalf("expected entries_grouped as []any, got %T(%v)", m["entries_grouped"], m["entries_grouped"])
		}
		if eg == nil {
			t.Fatal("entries_grouped must be non-nil empty slice (renders as [], never null)")
		}
		if len(eg) != 0 {
			t.Errorf("expected entries_grouped len 0, got %d", len(eg))
		}
		rq, ok := m["reflection_questions"].([]any)
		if !ok {
			t.Fatalf("expected reflection_questions as []any, got %T", m["reflection_questions"])
		}
		if len(rq) != 3 {
			t.Errorf("expected reflection_questions len 3, got %d", len(rq))
		}
	})
}

// TestToReview_ReflectionQuestionsExactWording locks the verbatim
// wording of the three reflection questions per stage Design Notes
// lines 367–369 (NOT the looser Success Criteria paraphrase).
func TestToReview_ReflectionQuestionsExactWording(t *testing.T) {
	opts := ReviewOptions{
		Scope: "week",
		Now:   reviewFixedNow,
	}
	wantWords := []string{
		"What pattern do you see in this period?",
		"What did you underestimate?",
		"What's missing here that should be?",
	}

	t.Run("markdown", func(t *testing.T) {
		got, err := ToReviewMarkdown(reviewFixture, opts)
		if err != nil {
			t.Fatalf("ToReviewMarkdown: %v", err)
		}
		lines := strings.Split(string(got), "\n")
		wantLines := []string{
			"1. What pattern do you see in this period?",
			"2. What did you underestimate?",
			"3. What's missing here that should be?",
		}
		// Find the start of the numbered list and assert the three
		// consecutive lines match verbatim.
		startIdx := -1
		for i, ln := range lines {
			if ln == wantLines[0] {
				startIdx = i
				break
			}
		}
		if startIdx == -1 {
			t.Fatalf("could not find reflection question 1 line %q in:\n%s", wantLines[0], string(got))
		}
		for i, want := range wantLines {
			if startIdx+i >= len(lines) {
				t.Fatalf("reached EOF; expected line %q at index %d", want, startIdx+i)
			}
			if lines[startIdx+i] != want {
				t.Errorf("line[%d]=%q, want %q", startIdx+i, lines[startIdx+i], want)
			}
		}
	})

	t.Run("json", func(t *testing.T) {
		got, err := ToReviewJSON(reviewFixture, opts)
		if err != nil {
			t.Fatalf("ToReviewJSON: %v", err)
		}
		var parsed struct {
			ReflectionQuestions []string `json:"reflection_questions"`
		}
		if err := json.Unmarshal(got, &parsed); err != nil {
			t.Fatalf("json.Unmarshal: %v", err)
		}
		if !reflect.DeepEqual(parsed.ReflectionQuestions, wantWords) {
			t.Errorf("reflection_questions wording mismatch\nwant: %#v\ngot:  %#v",
				wantWords, parsed.ReflectionQuestions)
		}
	})
}
