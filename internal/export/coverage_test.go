package export

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/jysf/bragfile000/internal/storage"
)

// coverageYearFixture: 10 entries across calendar 2026, modeling the
// post-v0.3.0 agent-adoption ramp — 0 agent-authored in H1, then agent
// entries appearing Jul/Sep/Nov/Dec. Exercises: agent: only (id 6),
// model: only (id 7), both (ids 4, 9), plain human (ids 1,2,3,5,10), the
// FALSE-POSITIVE guard (id 8: tags "agentic,modeling" — no colon → human),
// and self-reference (ids 1,4,9 mention "brag"). Totals: 4 agent / 6 human,
// agent_share 0.4; self-reference 3 (0.3).
var coverageYearFixture = []storage.Entry{
	{ID: 1, Title: "bragfile MVP retro", Description: "shipped the CLI", Tags: "process",
		CreatedAt: time.Date(2026, 2, 10, 10, 0, 0, 0, time.UTC)},
	{ID: 2, Title: "auth refactor", Description: "cleaned up login", Tags: "auth",
		CreatedAt: time.Date(2026, 3, 5, 10, 0, 0, 0, time.UTC)},
	{ID: 3, Title: "docs pass", Description: "rewrote the tutorial", Tags: "docs",
		CreatedAt: time.Date(2026, 5, 20, 10, 0, 0, 0, time.UTC)},
	{ID: 4, Title: "MCP server for brag", Description: "agent-native write spine",
		Tags:      "mcp,agent:claude-code,model:claude-opus-4-8",
		CreatedAt: time.Date(2026, 7, 4, 10, 0, 0, 0, time.UTC)},
	{ID: 5, Title: "hotfix streak bug", Description: "local-day streak", Tags: "fix",
		CreatedAt: time.Date(2026, 7, 5, 10, 0, 0, 0, time.UTC)},
	{ID: 6, Title: "impact digest", Description: "calendar windows", Tags: "agent:claude-code",
		CreatedAt: time.Date(2026, 9, 12, 10, 0, 0, 0, time.UTC)},
	{ID: 7, Title: "story surface", Description: "audience shaping", Tags: "model:claude-opus-4-8,narrative",
		CreatedAt: time.Date(2026, 11, 3, 10, 0, 0, 0, time.UTC)},
	{ID: 8, Title: "modeling notes", Description: "agentic patterns essay", Tags: "agentic,modeling",
		CreatedAt: time.Date(2026, 11, 20, 10, 0, 0, 0, time.UTC)},
	{ID: 9, Title: "wrapped + sparklines", Description: "shareable year in brags",
		Tags:      "agent:claude-code,model:claude-opus-4-8,visual",
		CreatedAt: time.Date(2026, 12, 15, 10, 0, 0, 0, time.UTC)},
	{ID: 10, Title: "release cut", Description: "v0.4.0 to homebrew", Tags: "release",
		CreatedAt: time.Date(2026, 12, 20, 10, 0, 0, 0, time.UTC)},
}
var coverageYearNow = time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC)
var coverageYearMonths = []string{ // the 12 ordered scope labels
	"2026-01", "2026-02", "2026-03", "2026-04", "2026-05", "2026-06",
	"2026-07", "2026-08", "2026-09", "2026-10", "2026-11", "2026-12"}

// Test 1 — TestToCoverageMarkdown_DEC014FullDocumentGolden (LOAD-BEARING).
func TestToCoverageMarkdown_DEC014FullDocumentGolden(t *testing.T) {
	opts := CoverageOptions{
		Scope:       "2026",
		Filters:     "(none)",
		FiltersJSON: nil,
		ScopeMonths: coverageYearMonths,
		Now:         coverageYearNow,
		Spark:       true,
	}
	want := `# Bragfile Coverage

Generated: 2026-12-31T23:59:59Z
Scope: 2026
Filters: (none)
Entries: 10

## Provenance share

- Agent-authored: 4 (40.0%)
- Human-authored: 6 (60.0%)

## Monthly trend

Agent share: ▁▁▁▁▁▁▅▁█▁▅▅

- 2026-01: 0 agent / 0 human (0%)
- 2026-02: 0 agent / 1 human (0%)
- 2026-03: 0 agent / 1 human (0%)
- 2026-04: 0 agent / 0 human (0%)
- 2026-05: 0 agent / 1 human (0%)
- 2026-06: 0 agent / 0 human (0%)
- 2026-07: 1 agent / 1 human (50%)
- 2026-08: 0 agent / 0 human (0%)
- 2026-09: 1 agent / 0 human (100%)
- 2026-10: 0 agent / 0 human (0%)
- 2026-11: 1 agent / 1 human (50%)
- 2026-12: 1 agent / 1 human (50%)

## Self-reference

- Entries mentioning brag/bragfile: 3 (30.0%)`

	got, err := ToCoverageMarkdown(coverageYearFixture, opts)
	if err != nil {
		t.Fatalf("ToCoverageMarkdown: %v", err)
	}
	if string(got) != want {
		t.Errorf("markdown golden mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

// Test 2 — TestToCoverageJSON_DEC033ShapeGolden (LOAD-BEARING).
func TestToCoverageJSON_DEC033ShapeGolden(t *testing.T) {
	opts := CoverageOptions{
		Scope:       "2026",
		Filters:     "(none)",
		FiltersJSON: nil,
		ScopeMonths: coverageYearMonths,
		Now:         coverageYearNow,
		Spark:       true, // ignored by JSON
	}
	want := `{
  "generated_at": "2026-12-31T23:59:59Z",
  "scope": "2026",
  "filters": {},
  "total_entries": 10,
  "agent_entries": 4,
  "human_entries": 6,
  "agent_share": 0.4,
  "by_month": [
    {
      "period": "2026-01",
      "agent": 0,
      "human": 0,
      "share": 0
    },
    {
      "period": "2026-02",
      "agent": 0,
      "human": 1,
      "share": 0
    },
    {
      "period": "2026-03",
      "agent": 0,
      "human": 1,
      "share": 0
    },
    {
      "period": "2026-04",
      "agent": 0,
      "human": 0,
      "share": 0
    },
    {
      "period": "2026-05",
      "agent": 0,
      "human": 1,
      "share": 0
    },
    {
      "period": "2026-06",
      "agent": 0,
      "human": 0,
      "share": 0
    },
    {
      "period": "2026-07",
      "agent": 1,
      "human": 1,
      "share": 0.5
    },
    {
      "period": "2026-08",
      "agent": 0,
      "human": 0,
      "share": 0
    },
    {
      "period": "2026-09",
      "agent": 1,
      "human": 0,
      "share": 1
    },
    {
      "period": "2026-10",
      "agent": 0,
      "human": 0,
      "share": 0
    },
    {
      "period": "2026-11",
      "agent": 1,
      "human": 1,
      "share": 0.5
    },
    {
      "period": "2026-12",
      "agent": 1,
      "human": 1,
      "share": 0.5
    }
  ],
  "self_reference": {
    "count": 3,
    "share": 0.3
  }
}`

	got, err := ToCoverageJSON(coverageYearFixture, opts)
	if err != nil {
		t.Fatalf("ToCoverageJSON: %v", err)
	}
	if string(got) != want {
		t.Errorf("JSON golden mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

// Test 3 — TestToCoverage_EmptyWindowShape.
func TestToCoverage_EmptyWindowShape(t *testing.T) {
	opts := CoverageOptions{
		Scope:       "2026-Q3",
		Filters:     "(none)",
		ScopeMonths: []string{"2026-07", "2026-08", "2026-09"},
		Now:         coverageYearNow,
		Spark:       true,
	}

	md, err := ToCoverageMarkdown(nil, opts)
	if err != nil {
		t.Fatalf("ToCoverageMarkdown: %v", err)
	}
	mds := string(md)
	if !strings.Contains(mds, "Entries: 0") {
		t.Errorf("empty markdown missing 'Entries: 0':\n%s", mds)
	}
	for _, section := range []string{"## Provenance share", "## Monthly trend", "## Self-reference"} {
		if strings.Contains(mds, section) {
			t.Errorf("empty markdown must omit %q:\n%s", section, mds)
		}
	}

	jsonBytes, err := ToCoverageJSON(nil, opts)
	if err != nil {
		t.Fatalf("ToCoverageJSON: %v", err)
	}
	var env struct {
		TotalEntries int     `json:"total_entries"`
		AgentEntries int     `json:"agent_entries"`
		HumanEntries int     `json:"human_entries"`
		AgentShare   float64 `json:"agent_share"`
		ByMonth      []struct {
			Period string  `json:"period"`
			Agent  int     `json:"agent"`
			Human  int     `json:"human"`
			Share  float64 `json:"share"`
		} `json:"by_month"`
		SelfReference struct {
			Count int     `json:"count"`
			Share float64 `json:"share"`
		} `json:"self_reference"`
		Filters map[string]string `json:"filters"`
	}
	if err := json.Unmarshal(jsonBytes, &env); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if env.TotalEntries != 0 || env.AgentEntries != 0 || env.HumanEntries != 0 || env.AgentShare != 0 {
		t.Errorf("empty JSON totals not all 0: %+v", env)
	}
	if len(env.ByMonth) != 3 {
		t.Fatalf("empty JSON by_month len = %d, want 3", len(env.ByMonth))
	}
	for _, b := range env.ByMonth {
		if b.Agent != 0 || b.Human != 0 || b.Share != 0 {
			t.Errorf("empty JSON bucket %q not zero: %+v", b.Period, b)
		}
	}
	if env.SelfReference.Count != 0 || env.SelfReference.Share != 0 {
		t.Errorf("empty JSON self_reference not zero: %+v", env.SelfReference)
	}
	if env.Filters == nil || len(env.Filters) != 0 {
		t.Errorf("empty JSON filters must be {} (non-nil, empty): %v", env.Filters)
	}
}

// Test 4 — TestToCoverage_SparklineMarkdownOnlyAndEscaped.
func TestToCoverage_SparklineMarkdownOnlyAndEscaped(t *testing.T) {
	base := CoverageOptions{
		Scope:       "2026",
		Filters:     "(none)",
		ScopeMonths: coverageYearMonths,
		Now:         coverageYearNow,
	}
	const glyphs = "▁▂▃▄▅▆▇█"

	// Spark:true → the Agent share: line is present, all runes after the label
	// are block glyphs.
	on := base
	on.Spark = true
	md, err := ToCoverageMarkdown(coverageYearFixture, on)
	if err != nil {
		t.Fatalf("markdown spark-on: %v", err)
	}
	mds := string(md)
	const label = "Agent share: "
	idx := strings.Index(mds, label)
	if idx < 0 {
		t.Fatalf("expected %q line with Spark:true:\n%s", label, mds)
	}
	rest := mds[idx+len(label):]
	if nl := strings.IndexByte(rest, '\n'); nl >= 0 {
		rest = rest[:nl]
	}
	if rest == "" {
		t.Fatal("Agent share: line has no glyphs")
	}
	for _, r := range rest {
		if !strings.ContainsRune(glyphs, r) {
			t.Errorf("non-glyph rune %q in sparkline %q", r, rest)
		}
	}

	// Spark:false → no Agent share: line.
	off := base
	off.Spark = false
	mdOff, err := ToCoverageMarkdown(coverageYearFixture, off)
	if err != nil {
		t.Fatalf("markdown spark-off: %v", err)
	}
	if strings.Contains(string(mdOff), "Agent share:") {
		t.Errorf("Spark:false must omit 'Agent share:':\n%s", mdOff)
	}

	// JSON never contains glyphs nor a sparkline key, regardless of Spark.
	for _, spark := range []bool{true, false} {
		o := base
		o.Spark = spark
		jb, err := ToCoverageJSON(coverageYearFixture, o)
		if err != nil {
			t.Fatalf("json spark=%v: %v", spark, err)
		}
		js := string(jb)
		for _, g := range glyphs {
			if strings.ContainsRune(js, g) {
				t.Errorf("JSON (spark=%v) must be glyph-free, found %q:\n%s", spark, g, js)
			}
		}
		if strings.Contains(js, "sparkline") {
			t.Errorf("JSON (spark=%v) must not contain a sparkline key", spark)
		}
	}
}

// Test 5 — TestToCoverage_FiltersEchoed.
func TestToCoverage_FiltersEchoed(t *testing.T) {
	opts := CoverageOptions{
		Scope:       "2026",
		Filters:     "--project alpha",
		FiltersJSON: map[string]string{"project": "alpha"},
		ScopeMonths: coverageYearMonths,
		Now:         coverageYearNow,
		Spark:       true,
	}
	md, err := ToCoverageMarkdown(coverageYearFixture, opts)
	if err != nil {
		t.Fatalf("ToCoverageMarkdown: %v", err)
	}
	if !strings.Contains(string(md), "Filters: --project alpha") {
		t.Errorf("markdown missing echoed filters line:\n%s", md)
	}

	jb, err := ToCoverageJSON(coverageYearFixture, opts)
	if err != nil {
		t.Fatalf("ToCoverageJSON: %v", err)
	}
	var env struct {
		Filters map[string]string `json:"filters"`
	}
	if err := json.Unmarshal(jb, &env); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(env.Filters) != 1 || env.Filters["project"] != "alpha" {
		t.Errorf("filters object = %v, want {project:alpha}", env.Filters)
	}
}
