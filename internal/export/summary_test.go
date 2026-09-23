package export

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strconv"
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
  ],
  "failures_by_project": []
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
	wantKeys := []string{"generated_at", "scope", "filters", "counts_by_type", "counts_by_project", "highlights", "failures_by_project"}
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
  "highlights": [],
  "failures_by_project": []
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

// summaryFailureFixture is summaryFixture plus three rows typed "failed" —
// the literal DEC-049 persists, spelled out rather than taken from
// aggregate.FailureType so these tests also fail if the constant drifts from
// the stored value. id 6 is an alpha failure dated between alpha's two
// highlights, so it must leave their group. id 7 is a failure with NO impact,
// and it still lands in ## What didn't work: summary's highlights list every
// entry, so its partition covers every entry (SPEC-095 LD1), unlike impact's
// and wrapped's. delta holds only id 7, so it is a failures-only project and
// must not appear under ## Highlights. id 8 is a (no project) failure, so the
// section's (no project)-last rule is exercised too. 8 in window: 5
// highlights, 3 failures.
var summaryFailureFixture = append(append([]storage.Entry{}, summaryFixture...),
	storage.Entry{ID: 6, Title: "pool-dead-end",
		Project: "alpha", Type: "failed",
		Impact:    "cost two days and produced nothing reusable",
		CreatedAt: time.Date(2026, 4, 21, 12, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 4, 21, 12, 0, 0, 0, time.UTC)},
	storage.Entry{ID: 7, Title: "retry-noimpact",
		Project: "delta", Type: "failed",
		Impact:    "", // no impact, and still listed (LD1)
		CreatedAt: time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)},
	storage.Entry{ID: 8, Title: "unbound-dead-end",
		Type:      "failed", // (no project) group
		Impact:    "ruled out the vendor SDK",
		CreatedAt: time.Date(2026, 4, 23, 12, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 4, 23, 12, 0, 0, 0, time.UTC)},
)

var summaryFailureOpts = SummaryOptions{
	Scope:       "week",
	Filters:     "(none)",
	FiltersJSON: map[string]string{},
	Now:         summaryFixedNow,
}

// TestToSummaryMarkdown_FailureSectionGolden ▲ SPEC-095 LD2/LD3
// (LOAD-BEARING). A recorded failure leaves ## Highlights and renders under
// ## What didn't work, in the same grouped `- <id>: <title>` shape. By type
// and By project are unchanged in meaning: they count all eight entries,
// both sections, and `failed: 3` stays where it always was.
func TestToSummaryMarkdown_FailureSectionGolden(t *testing.T) {
	got, err := ToSummaryMarkdown(summaryFailureFixture, summaryFailureOpts)
	if err != nil {
		t.Fatalf("ToSummaryMarkdown: %v", err)
	}
	want := `# Bragfile Summary

Generated: 2026-04-25T12:00:00Z
Scope: week
Filters: (none)

## Summary

**By type**
- failed: 3
- shipped: 3
- fixed: 1
- learned: 1

**By project**
- alpha: 3
- beta: 1
- delta: 1
- gamma: 1
- (no project): 2

## Highlights

### alpha

- 1: alpha-old
- 4: alpha-new

### beta

- 2: beta-mid

### gamma

- 5: gamma-only

### (no project)

- 3: unbound-mid

## What didn't work

### alpha

- 6: pool-dead-end

### delta

- 7: retry-noimpact

### (no project)

- 8: unbound-dead-end`
	if string(got) != want {
		t.Errorf("markdown golden mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

// TestToSummaryJSON_FailureSectionGolden ▲ SPEC-095 LD3/LD4 (LOAD-BEARING).
// The failures leave highlights and arrive in failures_by_project, in the
// same {project, entries: [{id, title}]} shape; both counts maps still count
// every entry.
func TestToSummaryJSON_FailureSectionGolden(t *testing.T) {
	got, err := ToSummaryJSON(summaryFailureFixture, summaryFailureOpts)
	if err != nil {
		t.Fatalf("ToSummaryJSON: %v", err)
	}
	want := `{
  "generated_at": "2026-04-25T12:00:00Z",
  "scope": "week",
  "filters": {},
  "counts_by_type": {
    "failed": 3,
    "fixed": 1,
    "learned": 1,
    "shipped": 3
  },
  "counts_by_project": {
    "(no project)": 2,
    "alpha": 3,
    "beta": 1,
    "delta": 1,
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
  ],
  "failures_by_project": [
    {
      "project": "alpha",
      "entries": [
        {
          "id": 6,
          "title": "pool-dead-end"
        }
      ]
    },
    {
      "project": "delta",
      "entries": [
        {
          "id": 7,
          "title": "retry-noimpact"
        }
      ]
    },
    {
      "project": "(no project)",
      "entries": [
        {
          "id": 8,
          "title": "unbound-dead-end"
        }
      ]
    }
  ]
}`
	if string(got) != want {
		t.Errorf("json golden mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

// TestToSummary_PartitionCoversEveryInWindowEntry ▲ SPEC-095 LD1/LD3 — DEC-050
// rules 2 and 3 on summary. Every in-window entry, with or without an impact,
// is listed exactly once across the two sections: a failure in ## What didn't
// work, anything else in ## Highlights. The impact-less failure (id 7) is the
// row impact and wrapped would not list; here it must be listed, or it is
// counted in By type and shown nowhere. Both counts maps still sum to every
// entry. Checked in both formats.
func TestToSummary_PartitionCoversEveryInWindowEntry(t *testing.T) {
	wantHighlights := "1,2,3,4,5"
	wantFailures := "6,7,8"

	md, err := ToSummaryMarkdown(summaryFailureFixture, summaryFailureOpts)
	if err != nil {
		t.Fatalf("ToSummaryMarkdown: %v", err)
	}
	// ids listed under a `## <heading>`, in ascending order, by line.
	sectionIDs := func(heading string) string {
		var ids []int
		in := false
		for _, ln := range strings.Split(string(md), "\n") {
			if strings.HasPrefix(ln, "## ") {
				in = ln == "## "+heading
				continue
			}
			var id int
			if in && strings.HasPrefix(ln, "- ") {
				if _, err := fmt.Sscanf(ln, "- %d:", &id); err == nil {
					ids = append(ids, id)
				}
			}
		}
		sort.Ints(ids)
		var out []string
		for _, id := range ids {
			out = append(out, strconv.Itoa(id))
		}
		return strings.Join(out, ",")
	}
	if got := sectionIDs("Highlights"); got != wantHighlights {
		t.Errorf("## Highlights ids = %q, want %q\n%s", got, wantHighlights, md)
	}
	if got := sectionIDs("What didn't work"); got != wantFailures {
		t.Errorf("## What didn't work ids = %q, want %q\n%s", got, wantFailures, md)
	}

	raw, err := ToSummaryJSON(summaryFailureFixture, summaryFailureOpts)
	if err != nil {
		t.Fatalf("ToSummaryJSON: %v", err)
	}
	var env struct {
		CountsByType      map[string]int   `json:"counts_by_type"`
		CountsByProject   map[string]int   `json:"counts_by_project"`
		Highlights        []highlightGroup `json:"highlights"`
		FailuresByProject []highlightGroup `json:"failures_by_project"`
	}
	if err := json.Unmarshal(raw, &env); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	groupIDs := func(groups []highlightGroup) string {
		var ids []int
		for _, g := range groups {
			for _, e := range g.Entries {
				ids = append(ids, int(e.ID))
			}
		}
		sort.Ints(ids)
		var out []string
		for _, id := range ids {
			out = append(out, strconv.Itoa(id))
		}
		return strings.Join(out, ",")
	}
	if got := groupIDs(env.Highlights); got != wantHighlights {
		t.Errorf("highlights ids = %q, want %q", got, wantHighlights)
	}
	if got := groupIDs(env.FailuresByProject); got != wantFailures {
		t.Errorf("failures_by_project ids = %q, want %q", got, wantFailures)
	}
	sum := func(m map[string]int) int {
		n := 0
		for _, v := range m {
			n += v
		}
		return n
	}
	if got := sum(env.CountsByType); got != len(summaryFailureFixture) {
		t.Errorf("counts_by_type sums to %d, want %d (both sections)", got, len(summaryFailureFixture))
	}
	if got := sum(env.CountsByProject); got != len(summaryFailureFixture) {
		t.Errorf("counts_by_project sums to %d, want %d (both sections)", got, len(summaryFailureFixture))
	}
}

// TestToSummary_SectionsRenderOnlyWhenNonEmpty ▲ SPEC-095 LD2 — DEC-050 rule 4
// on summary, and the failures-only window. Each `##` body section renders
// only when it has an entry, so a window holding only failures has no bare
// `## Highlights` heading (SPEC-086 pinned the same for `## Impact`). JSON
// always carries both keys: the failures-only window renders `"highlights":
// []` (DEC-014 part 4), and the clean window `"failures_by_project": []`.
func TestToSummary_SectionsRenderOnlyWhenNonEmpty(t *testing.T) {
	headings := func(md []byte) []string {
		var out []string
		for _, ln := range strings.Split(string(md), "\n") {
			if strings.HasPrefix(ln, "## ") {
				out = append(out, ln)
			}
		}
		return out
	}
	failuresOnly := []storage.Entry{summaryFailureFixture[5], summaryFailureFixture[6]}
	cases := []struct {
		name         string
		entries      []storage.Entry
		want         []string
		wantHL       int
		wantFailures int
	}{
		{"no failures", summaryFixture, []string{"## Summary", "## Highlights"}, 4, 0},
		{"failures only", failuresOnly, []string{"## Summary", "## What didn't work"}, 0, 2},
		{"both", summaryFailureFixture, []string{"## Summary", "## Highlights", "## What didn't work"}, 4, 3},
	}
	for _, c := range cases {
		md, err := ToSummaryMarkdown(c.entries, summaryFailureOpts)
		if err != nil {
			t.Fatalf("%s: ToSummaryMarkdown: %v", c.name, err)
		}
		if got := headings(md); !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s: ## headings = %q, want %q\n%s", c.name, got, c.want, md)
		}

		raw, err := ToSummaryJSON(c.entries, summaryFailureOpts)
		if err != nil {
			t.Fatalf("%s: ToSummaryJSON: %v", c.name, err)
		}
		var env map[string]json.RawMessage
		if err := json.Unmarshal(raw, &env); err != nil {
			t.Fatalf("%s: json.Unmarshal: %v", c.name, err)
		}
		for key, wantLen := range map[string]int{"highlights": c.wantHL, "failures_by_project": c.wantFailures} {
			v, ok := env[key]
			if !ok {
				t.Errorf("%s: JSON is missing %q:\n%s", c.name, key, raw)
				continue
			}
			var groups []highlightGroup
			if err := json.Unmarshal(v, &groups); err != nil || groups == nil {
				t.Errorf("%s: %q = %s, want an array (never null)", c.name, key, v)
				continue
			}
			if len(groups) != wantLen {
				t.Errorf("%s: %q has %d groups, want %d:\n%s", c.name, key, len(groups), wantLen, raw)
			}
		}
	}
}
