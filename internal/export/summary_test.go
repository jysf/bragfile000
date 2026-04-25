package export

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/jysf/bragfile000/internal/storage"
)

// summaryFixture is the load-bearing fixture shared across the markdown
// + JSON goldens (tests #5, #6) and the empty-state + filters-echo
// tests (#7, #8). 5 entries spanning 3 projects + (no project), with
// chrono-ordering chosen to exercise within-alpha chrono-ASC (1 → 4,
// IDs and timestamps NOT monotonic together so ID tie-break is
// testable separately) and (no project) forced last regardless of
// count.
var summaryFixture = []storage.Entry{
	{
		ID: 1, Title: "alpha-old",
		Description: "old alpha", // NOT rendered in summary
		Tags:        "auth", Project: "alpha", Type: "shipped",
		Impact:    "did stuff",
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

var summaryFixedNow = time.Date(2026, 4, 25, 12, 0, 0, 0, time.UTC)

// TestToSummaryMarkdown_DEC014FullDocumentGolden — LOAD-BEARING; written
// FIRST per SPEC-014/015 ship lessons. Locks every DEC-014 markdown
// choice on the shared fixture in one byte-exact assertion.
func TestToSummaryMarkdown_DEC014FullDocumentGolden(t *testing.T) {
	opts := SummaryOptions{
		Scope:       "week",
		Filters:     "(none)",
		FiltersJSON: nil,
		Now:         summaryFixedNow,
	}

	want := `# Bragfile Summary

Generated: 2026-04-25T12:00:00Z
Scope: week
Filters: (none)

## Summary

**By type**
- shipped: 3
- fixed: 1
- learned: 1

**By project**
- alpha: 2
- beta: 1
- gamma: 1
- (no project): 1

## Highlights

### alpha

- 1: alpha-old
- 4: alpha-new

### beta

- 2: beta-mid

### gamma

- 5: gamma-only

### (no project)

- 3: unbound-mid`

	got, err := ToSummaryMarkdown(summaryFixture, opts)
	if err != nil {
		t.Fatalf("ToSummaryMarkdown: %v", err)
	}
	if !bytes.Equal(got, []byte(want)) {
		t.Fatalf("DEC-014 markdown golden mismatch\nwant:\n%s\n\ngot:\n%s", want, string(got))
	}

	// Line-based heading-level checks per AGENTS.md §9 SPEC-015
	// substring-trap addendum.
	lines := strings.Split(string(got), "\n")
	if lines[0] != "# Bragfile Summary" {
		t.Errorf("expected first line %q, got %q", "# Bragfile Summary", lines[0])
	}
	foundSummary := false
	foundHighlights := false
	for _, ln := range lines {
		if ln == "## Summary" {
			foundSummary = true
		}
		if ln == "## Highlights" {
			foundHighlights = true
		}
	}
	if !foundSummary {
		t.Errorf("expected a standalone %q line", "## Summary")
	}
	if !foundHighlights {
		t.Errorf("expected a standalone %q line", "## Highlights")
	}
}

// TestToSummaryJSON_DEC014ShapeGolden — LOAD-BEARING; written SECOND.
// Locks every DEC-014 JSON choice on the shared fixture in one byte-
// exact assertion plus a parse-and-key-order check.
func TestToSummaryJSON_DEC014ShapeGolden(t *testing.T) {
	opts := SummaryOptions{
		Scope:       "week",
		Filters:     "(none)",
		FiltersJSON: map[string]string{},
		Now:         summaryFixedNow,
	}

	want := `{
  "generated_at": "2026-04-25T12:00:00Z",
  "scope": "week",
  "filters": {},
  "counts_by_type": {
    "fixed": 1,
    "learned": 1,
    "shipped": 3
  },
  "counts_by_project": {
    "(no project)": 1,
    "alpha": 2,
    "beta": 1,
    "gamma": 1
  },
  "highlights": [
    {
      "project": "alpha",
      "entries": [
        {
          "id": 1,
          "title": "alpha-old"
        },
        {
          "id": 4,
          "title": "alpha-new"
        }
      ]
    },
    {
      "project": "beta",
      "entries": [
        {
          "id": 2,
          "title": "beta-mid"
        }
      ]
    },
    {
      "project": "gamma",
      "entries": [
        {
          "id": 5,
          "title": "gamma-only"
        }
      ]
    },
    {
      "project": "(no project)",
      "entries": [
        {
          "id": 3,
          "title": "unbound-mid"
        }
      ]
    }
  ]
}`

	got, err := ToSummaryJSON(summaryFixture, opts)
	if err != nil {
		t.Fatalf("ToSummaryJSON: %v", err)
	}
	if !bytes.Equal(got, []byte(want)) {
		t.Fatalf("DEC-014 JSON golden mismatch\nwant:\n%s\n\ngot:\n%s", want, string(got))
	}

	// Verify struct-tag declaration order on top-level keys via a
	// json.Decoder walk. DEC-014 rests on this key order.
	dec := json.NewDecoder(bytes.NewReader(got))
	tok, err := dec.Token()
	if err != nil {
		t.Fatalf("decoder.Token open: %v", err)
	}
	if d, ok := tok.(json.Delim); !ok || d != '{' {
		t.Fatalf("expected opening {, got %v", tok)
	}
	wantKeys := []string{"generated_at", "scope", "filters", "counts_by_type", "counts_by_project", "highlights"}
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
		// Skip the value (any depth) so we land on the next key.
		if err := skipValue(dec); err != nil {
			t.Fatalf("skip value for %q: %v", k, err)
		}
	}
}

// skipValue consumes one full JSON value from dec, recursing into
// objects and arrays so the next dec.Token returns the next sibling
// token. For scalars, the value token is already consumed by the
// initial dec.Token() call. For containers, key/value pairs (object)
// or values (array) are walked until the matching close delim is
// consumed.
func skipValue(dec *json.Decoder) error {
	tok, err := dec.Token()
	if err != nil {
		return err
	}
	d, ok := tok.(json.Delim)
	if !ok {
		return nil
	}
	switch d {
	case '{':
		for dec.More() {
			// Consume key (always a string token).
			if _, err := dec.Token(); err != nil {
				return err
			}
			// Recurse on value.
			if err := skipValue(dec); err != nil {
				return err
			}
		}
		// Consume closing }.
		if _, err := dec.Token(); err != nil {
			return err
		}
	case '[':
		for dec.More() {
			if err := skipValue(dec); err != nil {
				return err
			}
		}
		// Consume closing ].
		if _, err := dec.Token(); err != nil {
			return err
		}
	}
	return nil
}

// TestToSummary_EmptyEntriesEmitsProvenanceOnly locks DEC-014's empty-
// state rule on both renderers: provenance always renders; the summary
// + highlights sections are OMITTED for empty inputs.
func TestToSummary_EmptyEntriesEmitsProvenanceOnly(t *testing.T) {
	t.Run("markdown", func(t *testing.T) {
		opts := SummaryOptions{
			Scope:       "week",
			Filters:     "(none)",
			FiltersJSON: nil,
			Now:         summaryFixedNow,
		}
		want := `# Bragfile Summary

Generated: 2026-04-25T12:00:00Z
Scope: week
Filters: (none)`
		got, err := ToSummaryMarkdown([]storage.Entry{}, opts)
		if err != nil {
			t.Fatalf("ToSummaryMarkdown: %v", err)
		}
		if !bytes.Equal(got, []byte(want)) {
			t.Fatalf("empty-state markdown mismatch\nwant:\n%s\ngot:\n%s", want, string(got))
		}
		// Line-based no-section assertions.
		for _, ln := range strings.Split(string(got), "\n") {
			if ln == "## Summary" {
				t.Errorf("Summary heading present in empty output")
			}
			if ln == "## Highlights" {
				t.Errorf("Highlights heading present in empty output")
			}
		}
	})

	t.Run("json", func(t *testing.T) {
		opts := SummaryOptions{
			Scope:       "week",
			Filters:     "(none)",
			FiltersJSON: map[string]string{},
			Now:         summaryFixedNow,
		}
		want := `{
  "generated_at": "2026-04-25T12:00:00Z",
  "scope": "week",
  "filters": {},
  "counts_by_type": {},
  "counts_by_project": {},
  "highlights": []
}`
		got, err := ToSummaryJSON([]storage.Entry{}, opts)
		if err != nil {
			t.Fatalf("ToSummaryJSON: %v", err)
		}
		if !bytes.Equal(got, []byte(want)) {
			t.Fatalf("empty-state JSON mismatch\nwant:\n%s\ngot:\n%s", want, string(got))
		}
		var m map[string]any
		if err := json.Unmarshal(got, &m); err != nil {
			t.Fatalf("json.Unmarshal: %v", err)
		}
		hl, ok := m["highlights"].([]any)
		if !ok {
			t.Fatalf("expected highlights as []any, got %T(%v)", m["highlights"], m["highlights"])
		}
		if hl == nil {
			t.Fatal("highlights must be non-nil empty slice (renders as [], never null)")
		}
		if len(hl) != 0 {
			t.Fatalf("expected len 0, got %d", len(hl))
		}
	})
}

// TestToSummaryJSON_FiltersEchoShape locks the populated-and-empty
// shape of the filters object on the JSON envelope. Pairs locked
// decision 1 part (2): top-level flat keys + filters as object.
func TestToSummaryJSON_FiltersEchoShape(t *testing.T) {
	t.Run("none_via_nil_map", func(t *testing.T) {
		opts := SummaryOptions{
			Scope:       "week",
			FiltersJSON: nil, // nil → must still render as {}
			Now:         summaryFixedNow,
		}
		got, err := ToSummaryJSON([]storage.Entry{}, opts)
		if err != nil {
			t.Fatalf("ToSummaryJSON: %v", err)
		}
		if !bytes.Contains(got, []byte("\"filters\": {}")) {
			t.Errorf("expected filters as empty object, got:\n%s", string(got))
		}
	})

	t.Run("populated", func(t *testing.T) {
		opts := SummaryOptions{
			Scope:       "week",
			FiltersJSON: map[string]string{"project": "platform", "tag": "auth"},
			Now:         summaryFixedNow,
		}
		got, err := ToSummaryJSON([]storage.Entry{}, opts)
		if err != nil {
			t.Fatalf("ToSummaryJSON: %v", err)
		}
		if !bytes.Contains(got, []byte("\"filters\": {")) {
			t.Errorf("expected filters block to start, got:\n%s", string(got))
		}
		var parsed struct {
			Filters map[string]string `json:"filters"`
		}
		if err := json.Unmarshal(got, &parsed); err != nil {
			t.Fatalf("json.Unmarshal: %v", err)
		}
		if parsed.Filters["project"] != "platform" {
			t.Errorf("expected filters.project=platform, got %q", parsed.Filters["project"])
		}
		if parsed.Filters["tag"] != "auth" {
			t.Errorf("expected filters.tag=auth, got %q", parsed.Filters["tag"])
		}
		// Alphabetical-ASC key order for the filters block (Go's
		// encoding/json sort).
		filtersBlock := "  \"filters\": {\n    \"project\": \"platform\",\n    \"tag\": \"auth\"\n  },"
		if !bytes.Contains(got, []byte(filtersBlock)) {
			t.Errorf("expected filters block:\n%s\ngot:\n%s", filtersBlock, string(got))
		}
	})
}

// TestToSummaryMarkdown_FiltersLineFormat locks the markdown
// Filters: line shape on both none and echoed-flag inputs. Pairs
// locked decision 1 part (3) on the markdown side.
func TestToSummaryMarkdown_FiltersLineFormat(t *testing.T) {
	t.Run("none", func(t *testing.T) {
		opts := SummaryOptions{
			Scope:   "week",
			Filters: "(none)",
			Now:     summaryFixedNow,
		}
		got, err := ToSummaryMarkdown([]storage.Entry{}, opts)
		if err != nil {
			t.Fatalf("ToSummaryMarkdown: %v", err)
		}
		found := false
		for _, ln := range strings.Split(string(got), "\n") {
			if ln == "Filters: (none)" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected line %q in:\n%s", "Filters: (none)", string(got))
		}
	})

	t.Run("echoed_flags", func(t *testing.T) {
		opts := SummaryOptions{
			Scope:   "week",
			Filters: "--project platform --tag auth",
			Now:     summaryFixedNow,
		}
		got, err := ToSummaryMarkdown([]storage.Entry{}, opts)
		if err != nil {
			t.Fatalf("ToSummaryMarkdown: %v", err)
		}
		found := false
		for _, ln := range strings.Split(string(got), "\n") {
			if ln == "Filters: --project platform --tag auth" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected line %q in:\n%s", "Filters: --project platform --tag auth", string(got))
		}
	})
}
