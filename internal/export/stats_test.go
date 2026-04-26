package export

import (
	"bytes"
	"encoding/json"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/jysf/bragfile000/internal/storage"
)

// statsFixture is the load-bearing SPEC-020 fixture. 8 entries spanning
// Apr 12 → Apr 25 (UTC), exercising:
//   - top-5 tag cap with alpha-ASC tiebreak (refactor drops at boundary)
//   - (no project) excluded from top_projects
//   - span 14 days inclusive
//   - streak: current=2 (Apr 24+25), longest=3 (Apr 12+13+14)
//   - entries_per_week: 8/(14/7)=4.00
var statsFixture = []storage.Entry{
	{ID: 1, Title: "alpha-old", Tags: "auth,security",
		Project: "alpha", Type: "shipped",
		CreatedAt: time.Date(2026, 4, 12, 10, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 4, 12, 10, 0, 0, 0, time.UTC)},
	{ID: 2, Title: "alpha-mid", Tags: "auth,backend",
		Project: "alpha", Type: "shipped",
		CreatedAt: time.Date(2026, 4, 13, 10, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 4, 13, 10, 0, 0, 0, time.UTC)},
	{ID: 3, Title: "alpha-new", Tags: "auth",
		Project: "alpha", Type: "learned",
		CreatedAt: time.Date(2026, 4, 14, 10, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 4, 14, 10, 0, 0, 0, time.UTC)},
	{ID: 4, Title: "beta-mid", Tags: "db",
		Project: "beta", Type: "shipped",
		CreatedAt: time.Date(2026, 4, 18, 10, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 4, 18, 10, 0, 0, 0, time.UTC)},
	{ID: 5, Title: "unbound-1", Tags: "perf",
		Project: "", Type: "shipped",
		CreatedAt: time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC)},
	{ID: 6, Title: "gamma-1", Tags: "refactor",
		Project: "gamma", Type: "fixed",
		CreatedAt: time.Date(2026, 4, 22, 10, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 4, 22, 10, 0, 0, 0, time.UTC)},
	{ID: 7, Title: "beta-late", Tags: "auth,db",
		Project: "beta", Type: "shipped",
		CreatedAt: time.Date(2026, 4, 24, 10, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 4, 24, 10, 0, 0, 0, time.UTC)},
	{ID: 8, Title: "alpha-last", Tags: "security",
		Project: "alpha", Type: "shipped",
		CreatedAt: time.Date(2026, 4, 25, 10, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 4, 25, 10, 0, 0, 0, time.UTC)},
}

var statsFixedNow = time.Date(2026, 4, 25, 12, 0, 0, 0, time.UTC)

// TestToStatsMarkdown_DEC014FullDocumentGolden — LOAD-BEARING (written
// FIRST per SPEC-014/015/018/019 ship lessons). Locks every DEC-014
// markdown choice on the shared statsFixture in one byte-exact assertion.
func TestToStatsMarkdown_DEC014FullDocumentGolden(t *testing.T) {
	opts := StatsOptions{Now: statsFixedNow}

	want := `# Bragfile Stats

Generated: 2026-04-25T12:00:00Z
Scope: lifetime
Filters: (none)

## Stats

**Activity**
- Total entries: 8
- Entries/week: 4.00

**Streaks**
- Current: 2 days
- Longest: 3 days

**Top tags**
- auth: 4
- db: 2
- security: 2
- backend: 1
- perf: 1

**Top projects**
- alpha: 4
- beta: 2
- gamma: 1

**Corpus span**
- First entry: 2026-04-12
- Last entry: 2026-04-25
- Days: 14`

	got, err := ToStatsMarkdown(statsFixture, opts)
	if err != nil {
		t.Fatalf("ToStatsMarkdown: %v", err)
	}
	if !bytes.Equal(got, []byte(want)) {
		t.Fatalf("DEC-014 stats markdown golden mismatch\nwant:\n%s\n\ngot:\n%s", want, string(got))
	}

	// Line-based heading-level checks per AGENTS.md §9 SPEC-015
	// substring-trap addendum.
	lines := strings.Split(string(got), "\n")
	if lines[0] != "# Bragfile Stats" {
		t.Errorf("expected first line %q, got %q", "# Bragfile Stats", lines[0])
	}
	foundStats := false
	for _, ln := range lines {
		if ln == "## Stats" {
			foundStats = true
		}
	}
	if !foundStats {
		t.Errorf("expected a standalone %q line", "## Stats")
	}
}

// TestToStatsJSON_DEC014ShapeGolden — LOAD-BEARING (written SECOND).
// Locks every DEC-014 JSON choice on the shared fixture in one byte-
// exact assertion plus a parse-and-key-order check.
func TestToStatsJSON_DEC014ShapeGolden(t *testing.T) {
	opts := StatsOptions{Now: statsFixedNow}

	want := `{
  "generated_at": "2026-04-25T12:00:00Z",
  "scope": "lifetime",
  "filters": {},
  "total_count": 8,
  "entries_per_week": 4,
  "current_streak": 2,
  "longest_streak": 3,
  "top_tags": [
    {
      "tag": "auth",
      "count": 4
    },
    {
      "tag": "db",
      "count": 2
    },
    {
      "tag": "security",
      "count": 2
    },
    {
      "tag": "backend",
      "count": 1
    },
    {
      "tag": "perf",
      "count": 1
    }
  ],
  "top_projects": [
    {
      "project": "alpha",
      "count": 4
    },
    {
      "project": "beta",
      "count": 2
    },
    {
      "project": "gamma",
      "count": 1
    }
  ],
  "corpus_span": {
    "first_entry_date": "2026-04-12",
    "last_entry_date": "2026-04-25",
    "days": 14
  }
}`

	got, err := ToStatsJSON(statsFixture, opts)
	if err != nil {
		t.Fatalf("ToStatsJSON: %v", err)
	}
	if !bytes.Equal(got, []byte(want)) {
		t.Fatalf("DEC-014 stats JSON golden mismatch\nwant:\n%s\n\ngot:\n%s", want, string(got))
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
	wantKeys := []string{
		"generated_at", "scope", "filters", "total_count",
		"entries_per_week", "current_streak", "longest_streak",
		"top_tags", "top_projects", "corpus_span",
	}
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

	// ORDERING + CAP locks via a parsed walk.
	var m map[string]any
	if err := json.Unmarshal(got, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	tags, ok := m["top_tags"].([]any)
	if !ok {
		t.Fatalf("top_tags not an array: %T", m["top_tags"])
	}
	if len(tags) != 5 {
		t.Errorf("top_tags cap: got len %d, want 5", len(tags))
	}
	first, ok := tags[0].(map[string]any)
	if !ok {
		t.Fatalf("top_tags[0] not an object: %T", tags[0])
	}
	if first["tag"] != "auth" {
		t.Errorf("top_tags[0].tag: got %v, want auth", first["tag"])
	}
}

// TestToStats_EmptyCorpusShape locks DEC-014's empty-state contract on
// stats: markdown ends after the Filters: line; JSON full envelope with
// zero scalars, [] arrays, null date fields in corpus_span.
func TestToStats_EmptyCorpusShape(t *testing.T) {
	opts := StatsOptions{Now: statsFixedNow}

	t.Run("markdown", func(t *testing.T) {
		want := `# Bragfile Stats

Generated: 2026-04-25T12:00:00Z
Scope: lifetime
Filters: (none)`
		got, err := ToStatsMarkdown([]storage.Entry{}, opts)
		if err != nil {
			t.Fatalf("ToStatsMarkdown: %v", err)
		}
		if !bytes.Equal(got, []byte(want)) {
			t.Fatalf("empty-state markdown mismatch\nwant:\n%s\n\ngot:\n%s", want, string(got))
		}
		// Negative line-based assertions per SPEC-015 substring-trap addendum.
		forbidden := []string{
			"## Stats",
			"**Activity**",
			"**Streaks**",
			"**Top tags**",
			"**Top projects**",
			"**Corpus span**",
		}
		for _, ln := range strings.Split(string(got), "\n") {
			for _, bad := range forbidden {
				if ln == bad {
					t.Errorf("empty-state markdown contains forbidden line %q", bad)
				}
			}
		}
	})

	t.Run("json", func(t *testing.T) {
		want := `{
  "generated_at": "2026-04-25T12:00:00Z",
  "scope": "lifetime",
  "filters": {},
  "total_count": 0,
  "entries_per_week": 0,
  "current_streak": 0,
  "longest_streak": 0,
  "top_tags": [],
  "top_projects": [],
  "corpus_span": {
    "first_entry_date": null,
    "last_entry_date": null,
    "days": 0
  }
}`
		got, err := ToStatsJSON([]storage.Entry{}, opts)
		if err != nil {
			t.Fatalf("ToStatsJSON: %v", err)
		}
		if !bytes.Equal(got, []byte(want)) {
			t.Fatalf("empty-state JSON mismatch\nwant:\n%s\n\ngot:\n%s", want, string(got))
		}

		var m map[string]any
		if err := json.Unmarshal(got, &m); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if v, ok := m["total_count"].(float64); !ok || v != 0 {
			t.Errorf("total_count: got %v(%T), want 0", m["total_count"], m["total_count"])
		}
		tags, ok := m["top_tags"].([]any)
		if !ok {
			t.Fatalf("top_tags not an array: %T", m["top_tags"])
		}
		if len(tags) != 0 {
			t.Errorf("top_tags empty: got len %d, want 0", len(tags))
		}
		projects, ok := m["top_projects"].([]any)
		if !ok {
			t.Fatalf("top_projects not an array: %T", m["top_projects"])
		}
		if len(projects) != 0 {
			t.Errorf("top_projects empty: got len %d, want 0", len(projects))
		}
		span, ok := m["corpus_span"].(map[string]any)
		if !ok {
			t.Fatalf("corpus_span not an object: %T", m["corpus_span"])
		}
		if span["first_entry_date"] != nil {
			t.Errorf("first_entry_date: got %v, want nil", span["first_entry_date"])
		}
		if span["last_entry_date"] != nil {
			t.Errorf("last_entry_date: got %v, want nil", span["last_entry_date"])
		}
		if v, ok := span["days"].(float64); !ok || v != 0 {
			t.Errorf("days: got %v(%T), want 0", span["days"], span["days"])
		}
	})
}

// TestToStats_TopFiveCapEnforcedAtBoundary builds 6-element fixtures
// where all six tie at count=1 and asserts the renderer caps at 5 with
// alpha-ASC tiebreak determining which 5.
func TestToStats_TopFiveCapEnforcedAtBoundary(t *testing.T) {
	day := time.Date(2026, 4, 25, 10, 0, 0, 0, time.UTC)
	opts := StatsOptions{Now: statsFixedNow}

	t.Run("tags", func(t *testing.T) {
		entries := []storage.Entry{
			{ID: 1, Title: "z", Tags: "zebra", CreatedAt: day, UpdatedAt: day},
			{ID: 2, Title: "y", Tags: "yak", CreatedAt: day, UpdatedAt: day},
			{ID: 3, Title: "x", Tags: "x-ray", CreatedAt: day, UpdatedAt: day},
			{ID: 4, Title: "w", Tags: "wolf", CreatedAt: day, UpdatedAt: day},
			{ID: 5, Title: "v", Tags: "vulture", CreatedAt: day, UpdatedAt: day},
			{ID: 6, Title: "u", Tags: "umbrella", CreatedAt: day, UpdatedAt: day},
		}
		got, err := ToStatsJSON(entries, opts)
		if err != nil {
			t.Fatalf("ToStatsJSON: %v", err)
		}
		var m map[string]any
		if err := json.Unmarshal(got, &m); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		tags, ok := m["top_tags"].([]any)
		if !ok {
			t.Fatalf("top_tags not an array: %T", m["top_tags"])
		}
		if len(tags) != 5 {
			t.Fatalf("top_tags cap: got len %d, want 5", len(tags))
		}
		wantNames := []string{"umbrella", "vulture", "wolf", "x-ray", "yak"}
		for i, want := range wantNames {
			el := tags[i].(map[string]any)
			if el["tag"] != want {
				t.Errorf("tags[%d].tag: got %v, want %v", i, el["tag"], want)
			}
		}
		for _, t0 := range tags {
			el := t0.(map[string]any)
			if el["tag"] == "zebra" {
				t.Errorf("zebra should be excluded by alpha-ASC at the boundary")
			}
		}
	})

	t.Run("projects", func(t *testing.T) {
		entries := []storage.Entry{
			{ID: 1, Title: "z", Project: "zebra", CreatedAt: day, UpdatedAt: day},
			{ID: 2, Title: "y", Project: "yak", CreatedAt: day, UpdatedAt: day},
			{ID: 3, Title: "x", Project: "x-ray", CreatedAt: day, UpdatedAt: day},
			{ID: 4, Title: "w", Project: "wolf", CreatedAt: day, UpdatedAt: day},
			{ID: 5, Title: "v", Project: "vulture", CreatedAt: day, UpdatedAt: day},
			{ID: 6, Title: "u", Project: "umbrella", CreatedAt: day, UpdatedAt: day},
		}
		got, err := ToStatsJSON(entries, opts)
		if err != nil {
			t.Fatalf("ToStatsJSON: %v", err)
		}
		var m map[string]any
		if err := json.Unmarshal(got, &m); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		projects, ok := m["top_projects"].([]any)
		if !ok {
			t.Fatalf("top_projects not an array: %T", m["top_projects"])
		}
		if len(projects) != 5 {
			t.Fatalf("top_projects cap: got len %d, want 5", len(projects))
		}
		wantNames := []string{"umbrella", "vulture", "wolf", "x-ray", "yak"}
		for i, want := range wantNames {
			el := projects[i].(map[string]any)
			if el["project"] != want {
				t.Errorf("projects[%d].project: got %v, want %v", i, el["project"], want)
			}
		}
		for _, p := range projects {
			el := p.(map[string]any)
			if el["project"] == "zebra" {
				t.Errorf("zebra should be excluded by alpha-ASC at the boundary")
			}
		}
	})
}

// TestEntriesPerWeek_DecimalWeeksAndSubWeekZero locks the §4 formula:
// decimal-weeks divided into total, sub-1-week → 0.0, 2-decimal round.
func TestEntriesPerWeek_DecimalWeeksAndSubWeekZero(t *testing.T) {
	opts := StatsOptions{Now: statsFixedNow}

	parseEPW := func(t *testing.T, body []byte) float64 {
		t.Helper()
		var m map[string]any
		if err := json.Unmarshal(body, &m); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		v, ok := m["entries_per_week"].(float64)
		if !ok {
			t.Fatalf("entries_per_week: got %T(%v), want float64", m["entries_per_week"], m["entries_per_week"])
		}
		return v
	}

	t.Run("sub_week_zero", func(t *testing.T) {
		entries := []storage.Entry{
			{ID: 1, CreatedAt: time.Date(2026, 4, 25, 10, 0, 0, 0, time.UTC)},
			{ID: 2, CreatedAt: time.Date(2026, 4, 25, 11, 0, 0, 0, time.UTC)},
			{ID: 3, CreatedAt: time.Date(2026, 4, 26, 10, 0, 0, 0, time.UTC)},
		}
		got, err := ToStatsJSON(entries, opts)
		if err != nil {
			t.Fatalf("ToStatsJSON: %v", err)
		}
		v := parseEPW(t, got)
		if v != 0.0 {
			t.Errorf("sub_week_zero: got %v, want 0.0", v)
		}
	})

	t.Run("exactly_one_week", func(t *testing.T) {
		entries := []storage.Entry{
			{ID: 1, CreatedAt: time.Date(2026, 4, 19, 10, 0, 0, 0, time.UTC)},
			{ID: 2, CreatedAt: time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC)},
			{ID: 3, CreatedAt: time.Date(2026, 4, 21, 10, 0, 0, 0, time.UTC)},
			{ID: 4, CreatedAt: time.Date(2026, 4, 22, 10, 0, 0, 0, time.UTC)},
			{ID: 5, CreatedAt: time.Date(2026, 4, 23, 10, 0, 0, 0, time.UTC)},
			{ID: 6, CreatedAt: time.Date(2026, 4, 24, 10, 0, 0, 0, time.UTC)},
			{ID: 7, CreatedAt: time.Date(2026, 4, 25, 10, 0, 0, 0, time.UTC)},
		}
		got, err := ToStatsJSON(entries, opts)
		if err != nil {
			t.Fatalf("ToStatsJSON: %v", err)
		}
		v := parseEPW(t, got)
		if math.Abs(v-7.0) > 0.001 {
			t.Errorf("exactly_one_week: got %v, want ~7.0", v)
		}
	})

	t.Run("partial_weeks_two_decimals", func(t *testing.T) {
		entries := []storage.Entry{
			{ID: 1, CreatedAt: time.Date(2026, 4, 15, 10, 0, 0, 0, time.UTC)},
			{ID: 2, CreatedAt: time.Date(2026, 4, 17, 10, 0, 0, 0, time.UTC)},
			{ID: 3, CreatedAt: time.Date(2026, 4, 19, 10, 0, 0, 0, time.UTC)},
			{ID: 4, CreatedAt: time.Date(2026, 4, 23, 10, 0, 0, 0, time.UTC)},
			{ID: 5, CreatedAt: time.Date(2026, 4, 25, 10, 0, 0, 0, time.UTC)},
		}
		got, err := ToStatsJSON(entries, opts)
		if err != nil {
			t.Fatalf("ToStatsJSON: %v", err)
		}
		v := parseEPW(t, got)
		// span_days = 11, weeks = 11/7 ≈ 1.5714, value = 5/1.5714 ≈ 3.18
		if math.Abs(v-3.18) > 0.001 {
			t.Errorf("partial_weeks_two_decimals: got %v, want ~3.18", v)
		}
	})
}
