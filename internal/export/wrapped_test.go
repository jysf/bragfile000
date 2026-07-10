package export

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/jysf/bragfile000/internal/aggregate"
	"github.com/jysf/bragfile000/internal/storage"
)

// wrappedYearFixture: 7 entries across calendar 2026. Exercises every
// section — cadence spread (Jan/Feb/Apr/Jul/Nov), projects alpha(3)/beta(2)/
// gamma(1) + one (no project), 3 with-impact entries (alpha x2, beta x1),
// a 2-day local streak (Jul 4 + Jul 5) as the longest run, varied tags/types.
var wrappedYearFixture = []storage.Entry{
	{ID: 1, Title: "kickoff", Project: "alpha", Type: "shipped", Tags: "auth,api",
		Impact:    "cut p95 login latency 40%",
		CreatedAt: time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)},
	{ID: 2, Title: "docs pass", Project: "beta", Type: "shipped", Tags: "docs",
		CreatedAt: time.Date(2026, 2, 3, 10, 0, 0, 0, time.UTC)},
	{ID: 3, Title: "migration", Project: "alpha", Type: "shipped", Tags: "db,api",
		Impact:    "removed the nightly cron entirely",
		CreatedAt: time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC)},
	{ID: 4, Title: "reflection", Project: "", Type: "learned", Tags: "process",
		CreatedAt: time.Date(2026, 4, 22, 10, 0, 0, 0, time.UTC)},
	{ID: 5, Title: "launch", Project: "beta", Type: "shipped", Tags: "api",
		Impact:    "onboarding time down to 1 day",
		CreatedAt: time.Date(2026, 7, 4, 10, 0, 0, 0, time.UTC)},
	{ID: 6, Title: "hotfix", Project: "alpha", Type: "fixed", Tags: "auth",
		CreatedAt: time.Date(2026, 7, 5, 10, 0, 0, 0, time.UTC)},
	{ID: 7, Title: "retro notes", Project: "gamma", Type: "learned", Tags: "process,docs",
		CreatedAt: time.Date(2026, 11, 30, 10, 0, 0, 0, time.UTC)},
}
var wrappedYearNow = time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC)

// yearMonths is the 12 ordered "YYYY-MM" labels for calendar 2026.
var yearMonths = []string{
	"2026-01", "2026-02", "2026-03", "2026-04", "2026-05", "2026-06",
	"2026-07", "2026-08", "2026-09", "2026-10", "2026-11", "2026-12",
}

var q3Months = []string{"2026-07", "2026-08", "2026-09"}

func TestToWrappedMarkdown_DEC014FullDocumentGolden(t *testing.T) {
	opts := WrappedOptions{
		Scope:       "2026",
		Filters:     "(none)",
		FiltersJSON: nil,
		ScopeMonths: yearMonths,
		Now:         wrappedYearNow,
		Spark:       true,
	}
	got, err := ToWrappedMarkdown(wrappedYearFixture, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := `# Bragfile Wrapped

Generated: 2026-12-31T23:59:59Z
Scope: 2026
Filters: (none)
Entries: 7

## Cadence

Busiest month: 2026-04 (2)
Cadence: ▅▅▁█▁▁█▁▁▁▅▁

- 2026-01: 1
- 2026-02: 1
- 2026-03: 0
- 2026-04: 2
- 2026-05: 0
- 2026-06: 0
- 2026-07: 2
- 2026-08: 0
- 2026-09: 0
- 2026-10: 0
- 2026-11: 1
- 2026-12: 0

## Top initiatives

- alpha: 3
- beta: 2
- gamma: 1

## Impact moments

### alpha

- 1: kickoff
  cut p95 login latency 40%
- 3: migration
  removed the nightly cron entirely

### beta

- 5: launch
  onboarding time down to 1 day

## Rhythm

Longest streak: 2 days

**Top tags**
- api: 3
- auth: 2
- docs: 2
- process: 2
- db: 1

**Top types**
- shipped: 4
- learned: 2
- fixed: 1

## Span

- First entry: 2026-01-15
- Last entry: 2026-11-30
- Active days: 320`
	if string(got) != want {
		t.Errorf("markdown golden mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestToWrappedJSON_DEC030ShapeGolden(t *testing.T) {
	opts := WrappedOptions{
		Scope:       "2026",
		Filters:     "(none)",
		FiltersJSON: nil,
		ScopeMonths: yearMonths,
		Now:         wrappedYearNow,
	}
	got, err := ToWrappedJSON(wrappedYearFixture, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := `{
  "generated_at": "2026-12-31T23:59:59Z",
  "scope": "2026",
  "filters": {},
  "total_entries": 7,
  "cadence": {
    "busiest_month": "2026-04",
    "series": [
      {
        "period": "2026-01",
        "count": 1
      },
      {
        "period": "2026-02",
        "count": 1
      },
      {
        "period": "2026-03",
        "count": 0
      },
      {
        "period": "2026-04",
        "count": 2
      },
      {
        "period": "2026-05",
        "count": 0
      },
      {
        "period": "2026-06",
        "count": 0
      },
      {
        "period": "2026-07",
        "count": 2
      },
      {
        "period": "2026-08",
        "count": 0
      },
      {
        "period": "2026-09",
        "count": 0
      },
      {
        "period": "2026-10",
        "count": 0
      },
      {
        "period": "2026-11",
        "count": 1
      },
      {
        "period": "2026-12",
        "count": 0
      }
    ]
  },
  "top_initiatives": [
    {
      "project": "alpha",
      "count": 3
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
  "impact_moments": [
    {
      "project": "alpha",
      "entries": [
        {
          "id": 1,
          "title": "kickoff",
          "project": "alpha",
          "impact": "cut p95 login latency 40%"
        },
        {
          "id": 3,
          "title": "migration",
          "project": "alpha",
          "impact": "removed the nightly cron entirely"
        }
      ]
    },
    {
      "project": "beta",
      "entries": [
        {
          "id": 5,
          "title": "launch",
          "project": "beta",
          "impact": "onboarding time down to 1 day"
        }
      ]
    }
  ],
  "longest_streak": 2,
  "top_tags": [
    {
      "name": "api",
      "count": 3
    },
    {
      "name": "auth",
      "count": 2
    },
    {
      "name": "docs",
      "count": 2
    },
    {
      "name": "process",
      "count": 2
    },
    {
      "name": "db",
      "count": 1
    }
  ],
  "top_types": [
    {
      "name": "shipped",
      "count": 4
    },
    {
      "name": "learned",
      "count": 2
    },
    {
      "name": "fixed",
      "count": 1
    }
  ],
  "span": {
    "first_entry_date": "2026-01-15",
    "last_entry_date": "2026-11-30",
    "active_days": 320
  }
}`
	if string(got) != want {
		t.Errorf("json golden mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestToWrappedMarkdown_QuarterGolden(t *testing.T) {
	entries := []storage.Entry{wrappedYearFixture[4], wrappedYearFixture[5]}
	opts := WrappedOptions{
		Scope:       "2026-Q3",
		ScopeMonths: q3Months,
		Filters:     "(none)",
		Now:         time.Date(2026, 9, 30, 23, 59, 59, 0, time.UTC),
		Spark:       true,
	}
	got, err := ToWrappedMarkdown(entries, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := `# Bragfile Wrapped

Generated: 2026-09-30T23:59:59Z
Scope: 2026-Q3
Filters: (none)
Entries: 2

## Cadence

Busiest month: 2026-07 (2)
Cadence: █▁▁

- 2026-07: 2
- 2026-08: 0
- 2026-09: 0

## Top initiatives

- alpha: 1
- beta: 1

## Impact moments

### beta

- 5: launch
  onboarding time down to 1 day

## Rhythm

Longest streak: 2 days

**Top tags**
- api: 1
- auth: 1

**Top types**
- fixed: 1
- shipped: 1

## Span

- First entry: 2026-07-04
- Last entry: 2026-07-05
- Active days: 2`
	if string(got) != want {
		t.Errorf("quarter markdown golden mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

// TestToWrappedMarkdown_NoSparkOmitsGlyphLine (LOAD-BEARING — the escape
// path). With Spark: false the rendered markdown is byte-identical to the
// SPEC-051 (pre-sparkline) document: no Cadence: glyph line, everything
// else unchanged. Proves the Spark gate actually gates.
func TestToWrappedMarkdown_NoSparkOmitsGlyphLine(t *testing.T) {
	opts := WrappedOptions{
		Scope:       "2026",
		Filters:     "(none)",
		FiltersJSON: nil,
		ScopeMonths: yearMonths,
		Now:         wrappedYearNow,
		Spark:       false,
	}
	got, err := ToWrappedMarkdown(wrappedYearFixture, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := `# Bragfile Wrapped

Generated: 2026-12-31T23:59:59Z
Scope: 2026
Filters: (none)
Entries: 7

## Cadence

Busiest month: 2026-04 (2)

- 2026-01: 1
- 2026-02: 1
- 2026-03: 0
- 2026-04: 2
- 2026-05: 0
- 2026-06: 0
- 2026-07: 2
- 2026-08: 0
- 2026-09: 0
- 2026-10: 0
- 2026-11: 1
- 2026-12: 0

## Top initiatives

- alpha: 3
- beta: 2
- gamma: 1

## Impact moments

### alpha

- 1: kickoff
  cut p95 login latency 40%
- 3: migration
  removed the nightly cron entirely

### beta

- 5: launch
  onboarding time down to 1 day

## Rhythm

Longest streak: 2 days

**Top tags**
- api: 3
- auth: 2
- docs: 2
- process: 2
- db: 1

**Top types**
- shipped: 4
- learned: 2
- fixed: 1

## Span

- First entry: 2026-01-15
- Last entry: 2026-11-30
- Active days: 320`
	if string(got) != want {
		t.Errorf("no-spark markdown golden mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
	for _, line := range strings.Split(string(got), "\n") {
		if strings.HasPrefix(line, "Cadence: ") {
			t.Errorf("Spark: false must omit the glyph line, found: %q", line)
		}
	}
}

// TestToWrappedMarkdown_EmptyPeriodNoGlyphLine: over nil entries with
// Spark: true, the markdown is provenance-only (through Entries: 0) and
// contains neither ## Cadence nor a Cadence: glyph line (DEC-014 part-4
// body omission leaves no cadence section to decorate).
func TestToWrappedMarkdown_EmptyPeriodNoGlyphLine(t *testing.T) {
	opts := WrappedOptions{
		Scope:       "2026-Q3",
		ScopeMonths: q3Months,
		Filters:     "(none)",
		Now:         time.Date(2026, 9, 30, 23, 59, 59, 0, time.UTC),
		Spark:       true,
	}
	got, err := ToWrappedMarkdown(nil, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(string(got), "## Cadence") {
		t.Errorf("empty period must omit ## Cadence:\n%s", got)
	}
	if strings.Contains(string(got), "Cadence: ") {
		t.Errorf("empty period must have no Cadence: glyph line:\n%s", got)
	}
}

func TestToWrapped_EmptyPeriodShape(t *testing.T) {
	opts := WrappedOptions{
		Scope:       "2026-Q3",
		ScopeMonths: q3Months,
		Filters:     "(none)",
		Now:         time.Date(2026, 9, 30, 23, 59, 59, 0, time.UTC),
	}

	md, err := ToWrappedMarkdown(nil, opts)
	if err != nil {
		t.Fatalf("markdown: unexpected error: %v", err)
	}
	wantMD := `# Bragfile Wrapped

Generated: 2026-09-30T23:59:59Z
Scope: 2026-Q3
Filters: (none)
Entries: 0`
	if string(md) != wantMD {
		t.Errorf("empty markdown mismatch:\n--- got ---\n%s\n--- want ---\n%s", md, wantMD)
	}
	if strings.Contains(string(md), "## Cadence") {
		t.Errorf("empty markdown must omit body sections; found ## Cadence:\n%s", md)
	}

	jsonBytes, err := ToWrappedJSON(nil, opts)
	if err != nil {
		t.Fatalf("json: unexpected error: %v", err)
	}
	var env map[string]json.RawMessage
	if err := json.Unmarshal(jsonBytes, &env); err != nil {
		t.Fatalf("json unmarshal: %v\n%s", err, jsonBytes)
	}
	assertRaw(t, env, "total_entries", "0")
	assertRaw(t, env, "top_initiatives", "[]")
	assertRaw(t, env, "impact_moments", "[]")
	assertRaw(t, env, "longest_streak", "0")
	assertRaw(t, env, "top_tags", "[]")
	assertRaw(t, env, "top_types", "[]")
	assertRaw(t, env, "filters", "{}")

	var cadence struct {
		BusiestMonth *string                   `json:"busiest_month"`
		Series       []aggregate.CadenceBucket `json:"series"`
	}
	if err := json.Unmarshal(env["cadence"], &cadence); err != nil {
		t.Fatalf("cadence unmarshal: %v", err)
	}
	if cadence.BusiestMonth != nil {
		t.Errorf("empty period busiest_month: got %q, want null", *cadence.BusiestMonth)
	}
	if len(cadence.Series) != 3 {
		t.Fatalf("empty period series length: got %d, want 3", len(cadence.Series))
	}
	for _, b := range cadence.Series {
		if b.Count != 0 {
			t.Errorf("empty period series bucket %s: got count %d, want 0", b.Period, b.Count)
		}
	}

	var span struct {
		First      *string `json:"first_entry_date"`
		Last       *string `json:"last_entry_date"`
		ActiveDays int     `json:"active_days"`
	}
	if err := json.Unmarshal(env["span"], &span); err != nil {
		t.Fatalf("span unmarshal: %v", err)
	}
	if span.First != nil || span.Last != nil {
		t.Errorf("empty period span dates: got %v/%v, want null/null", span.First, span.Last)
	}
	if span.ActiveDays != 0 {
		t.Errorf("empty period active_days: got %d, want 0", span.ActiveDays)
	}
}

func TestToWrapped_ImpactTextRenderedInFull(t *testing.T) {
	longImpact := "reduced end-to-end onboarding from three weeks to a single afternoon by automating the account-provisioning handoff and cutting four manual approval steps"
	entries := []storage.Entry{
		{ID: 1, Title: "big win", Project: "alpha", Type: "shipped", Tags: "api",
			Impact:    longImpact,
			CreatedAt: time.Date(2026, 3, 10, 10, 0, 0, 0, time.UTC)},
	}
	opts := WrappedOptions{Scope: "2026", ScopeMonths: yearMonths, Filters: "(none)", Now: wrappedYearNow}

	md, err := ToWrappedMarkdown(entries, opts)
	if err != nil {
		t.Fatalf("markdown: %v", err)
	}
	if !strings.Contains(string(md), longImpact) {
		t.Errorf("full impact text must appear in markdown, never elided:\n%s", md)
	}

	jsonBytes, err := ToWrappedJSON(entries, opts)
	if err != nil {
		t.Fatalf("json: %v", err)
	}
	if !strings.Contains(string(jsonBytes), longImpact) {
		t.Errorf("full impact text must appear in json, never elided:\n%s", jsonBytes)
	}
}

func TestToWrapped_FiltersEchoed(t *testing.T) {
	entries := []storage.Entry{
		{ID: 1, Title: "x", Project: "alpha", Type: "shipped", Tags: "api",
			CreatedAt: time.Date(2026, 3, 10, 10, 0, 0, 0, time.UTC)},
	}
	opts := WrappedOptions{
		Scope:       "2026",
		ScopeMonths: yearMonths,
		Filters:     "--project alpha",
		FiltersJSON: map[string]string{"project": "alpha"},
		Now:         wrappedYearNow,
	}

	md, err := ToWrappedMarkdown(entries, opts)
	if err != nil {
		t.Fatalf("markdown: %v", err)
	}
	if !strings.Contains(string(md), "Filters: --project alpha") {
		t.Errorf("expected echoed filters line in markdown:\n%s", md)
	}

	jsonBytes, err := ToWrappedJSON(entries, opts)
	if err != nil {
		t.Fatalf("json: %v", err)
	}
	var env struct {
		Filters map[string]string `json:"filters"`
	}
	if err := json.Unmarshal(jsonBytes, &env); err != nil {
		t.Fatalf("json unmarshal: %v", err)
	}
	if len(env.Filters) != 1 || env.Filters["project"] != "alpha" {
		t.Errorf("filters json: got %v, want only {project: alpha}", env.Filters)
	}
}

func TestCadence_ZeroFilledAndBusiest(t *testing.T) {
	months := []string{"2026-07", "2026-08", "2026-09"}
	entries := []storage.Entry{
		{ID: 1, CreatedAt: time.Date(2026, 7, 4, 10, 0, 0, 0, time.UTC)},
		{ID: 2, CreatedAt: time.Date(2026, 7, 5, 10, 0, 0, 0, time.UTC)},
		{ID: 3, CreatedAt: time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC)},
	}
	series, busiest := aggregate.Cadence(entries, months)
	want := []aggregate.CadenceBucket{
		{Period: "2026-07", Count: 2},
		{Period: "2026-08", Count: 0},
		{Period: "2026-09", Count: 1},
	}
	if len(series) != len(want) {
		t.Fatalf("series length: got %d, want %d", len(series), len(want))
	}
	for i := range want {
		if series[i] != want[i] {
			t.Errorf("bucket %d: got %+v, want %+v", i, series[i], want[i])
		}
	}
	if busiest != "2026-07" {
		t.Errorf("busiest: got %q, want %q", busiest, "2026-07")
	}

	emptySeries, emptyBusiest := aggregate.Cadence(nil, months)
	if len(emptySeries) != 3 {
		t.Fatalf("empty series length: got %d, want 3", len(emptySeries))
	}
	for _, b := range emptySeries {
		if b.Count != 0 {
			t.Errorf("empty series bucket %s: got %d, want 0", b.Period, b.Count)
		}
	}
	if emptyBusiest != "" {
		t.Errorf("empty busiest: got %q, want \"\"", emptyBusiest)
	}
}

func assertRaw(t *testing.T, env map[string]json.RawMessage, key, want string) {
	t.Helper()
	got, ok := env[key]
	if !ok {
		t.Errorf("missing key %q", key)
		return
	}
	if strings.TrimSpace(string(got)) != want {
		t.Errorf("key %q: got %s, want %s", key, got, want)
	}
}
