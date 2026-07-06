package export

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/jysf/bragfile000/internal/storage"
)

// impactFixture: 5 entries. 3 carry impact (shown), 2 don't (counted,
// not shown). Projects alpha (2 with impact), beta (1 with impact),
// gamma (1 in-window but NO impact → excluded from body/counts),
// (no project) (1 in-window, NO impact → excluded). Chrono-ASC within
// alpha (IDs 1 then 4) with non-monotonic id/time pairing so the
// ID-tiebreak path in GroupEntriesByProject is exercised.
var impactFixture = []storage.Entry{
	{ID: 1, Title: "alpha-old", Description: "d", Tags: "auth",
		Project: "alpha", Type: "shipped",
		Impact:    "cut p95 login latency 40%",
		CreatedAt: time.Date(2026, 7, 1, 10, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 7, 1, 10, 0, 0, 0, time.UTC)},
	{ID: 2, Title: "beta-mid",
		Project: "beta", Type: "learned",
		Impact:    "onboarding time down to 1 day",
		CreatedAt: time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC)},
	{ID: 3, Title: "gamma-noimpact",
		Project: "gamma", Type: "shipped",
		Impact:    "", // in-window, NO impact → excluded
		CreatedAt: time.Date(2026, 7, 3, 10, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 7, 3, 10, 0, 0, 0, time.UTC)},
	{ID: 4, Title: "alpha-new",
		Project: "alpha", Type: "shipped",
		Impact:    "removed the nightly cron entirely",
		CreatedAt: time.Date(2026, 7, 4, 10, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 7, 4, 10, 0, 0, 0, time.UTC)},
	{ID: 5, Title: "unbound-noimpact",
		Type:      "fixed", // (no project), NO impact → excluded
		Impact:    "",
		CreatedAt: time.Date(2026, 7, 5, 10, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 7, 5, 10, 0, 0, 0, time.UTC)},
}

var impactFixedNow = time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)

// Test 1 — TestToImpactMarkdown_DEC014FullDocumentGolden (LOAD-BEARING).
// Byte-exact assertion of the full markdown document over impactFixture.
func TestToImpactMarkdown_DEC014FullDocumentGolden(t *testing.T) {
	opts := ImpactOptions{
		Scope:           "quarter",
		Filters:         "(none)",
		FiltersJSON:     nil,
		EntriesInWindow: 5,
		Now:             impactFixedNow,
	}
	got, err := ToImpactMarkdown(impactFixture, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := `# Bragfile Impact

Generated: 2026-07-06T12:00:00Z
Scope: quarter
Filters: (none)
Entries: 3/5 with impact

## Impact

### alpha

- 1: alpha-old
  cut p95 login latency 40%
- 4: alpha-new
  removed the nightly cron entirely

### beta

- 2: beta-mid
  onboarding time down to 1 day`
	if string(got) != want {
		t.Errorf("markdown golden mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

// Test 2 — TestToImpactJSON_DEC028ShapeGolden (LOAD-BEARING).
// Byte-exact JSON assertion on the same fixture/opts (FiltersJSON nil → {}).
func TestToImpactJSON_DEC028ShapeGolden(t *testing.T) {
	opts := ImpactOptions{
		Scope:           "quarter",
		Filters:         "(none)",
		FiltersJSON:     nil,
		EntriesInWindow: 5,
		Now:             impactFixedNow,
	}
	got, err := ToImpactJSON(impactFixture, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := `{
  "generated_at": "2026-07-06T12:00:00Z",
  "scope": "quarter",
  "filters": {},
  "entries_in_window": 5,
  "entries_with_impact": 3,
  "counts_by_project": {
    "alpha": 2,
    "beta": 1
  },
  "impact_by_project": [
    {
      "project": "alpha",
      "entries": [
        {
          "id": 1,
          "title": "alpha-old",
          "project": "alpha",
          "impact": "cut p95 login latency 40%"
        },
        {
          "id": 4,
          "title": "alpha-new",
          "project": "alpha",
          "impact": "removed the nightly cron entirely"
        }
      ]
    },
    {
      "project": "beta",
      "entries": [
        {
          "id": 2,
          "title": "beta-mid",
          "project": "beta",
          "impact": "onboarding time down to 1 day"
        }
      ]
    }
  ]
}`
	if string(got) != want {
		t.Errorf("json golden mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

// Test 3 — TestToImpact_EmptyWindowShape.
func TestToImpact_EmptyWindowShape(t *testing.T) {
	opts := ImpactOptions{
		Scope:           "quarter",
		Filters:         "(none)",
		FiltersJSON:     nil,
		EntriesInWindow: 0,
		Now:             impactFixedNow,
	}

	md, err := ToImpactMarkdown(nil, opts)
	if err != nil {
		t.Fatalf("markdown: unexpected error: %v", err)
	}
	wantMD := `# Bragfile Impact

Generated: 2026-07-06T12:00:00Z
Scope: quarter
Filters: (none)
Entries: 0/0 with impact`
	if string(md) != wantMD {
		t.Errorf("empty markdown mismatch:\n--- got ---\n%s\n--- want ---\n%s", md, wantMD)
	}
	if strings.Contains(string(md), "## Impact") {
		t.Errorf("empty markdown must not contain the ## Impact body, got:\n%s", md)
	}

	jsonBytes, err := ToImpactJSON(nil, opts)
	if err != nil {
		t.Fatalf("json: unexpected error: %v", err)
	}
	var env map[string]json.RawMessage
	if err := json.Unmarshal(jsonBytes, &env); err != nil {
		t.Fatalf("json unmarshal: %v", err)
	}
	if string(env["entries_in_window"]) != "0" {
		t.Errorf("entries_in_window: got %s, want 0", env["entries_in_window"])
	}
	if string(env["entries_with_impact"]) != "0" {
		t.Errorf("entries_with_impact: got %s, want 0", env["entries_with_impact"])
	}
	if string(env["counts_by_project"]) != "{}" {
		t.Errorf("counts_by_project: got %s, want {}", env["counts_by_project"])
	}
	if string(env["impact_by_project"]) != "[]" {
		t.Errorf("impact_by_project: got %s, want []", env["impact_by_project"])
	}
	if string(env["filters"]) != "{}" {
		t.Errorf("filters: got %s, want {}", env["filters"])
	}
}

// Test 4 — TestToImpact_InWindowButNoImpactExcluded.
func TestToImpact_InWindowButNoImpactExcluded(t *testing.T) {
	opts := ImpactOptions{
		Scope:           "quarter",
		Filters:         "(none)",
		FiltersJSON:     nil,
		EntriesInWindow: 5,
		Now:             impactFixedNow,
	}

	md, err := ToImpactMarkdown(impactFixture, opts)
	if err != nil {
		t.Fatalf("markdown: unexpected error: %v", err)
	}
	mds := string(md)
	if !strings.Contains(mds, "3/5 with impact") {
		t.Errorf("expected tally '3/5 with impact' in:\n%s", mds)
	}
	if strings.Contains(mds, "gamma-noimpact") {
		t.Errorf("markdown must NOT contain the impact-less entry 'gamma-noimpact':\n%s", mds)
	}
	if strings.Contains(mds, "unbound-noimpact") {
		t.Errorf("markdown must NOT contain the impact-less entry 'unbound-noimpact':\n%s", mds)
	}

	jsonBytes, err := ToImpactJSON(impactFixture, opts)
	if err != nil {
		t.Fatalf("json: unexpected error: %v", err)
	}
	var env struct {
		EntriesInWindow   int `json:"entries_in_window"`
		EntriesWithImpact int `json:"entries_with_impact"`
	}
	if err := json.Unmarshal(jsonBytes, &env); err != nil {
		t.Fatalf("json unmarshal: %v", err)
	}
	if env.EntriesInWindow != 5 {
		t.Errorf("entries_in_window: got %d, want 5", env.EntriesInWindow)
	}
	if env.EntriesWithImpact != 3 {
		t.Errorf("entries_with_impact: got %d, want 3", env.EntriesWithImpact)
	}
}

// Test 5 — TestToImpact_ImpactTextRenderedInFull.
func TestToImpact_ImpactTextRenderedInFull(t *testing.T) {
	const longImpact = "cut infra cost 42%: retired the legacy queue, migrated 3 services, saved $12k/mo"
	entries := []storage.Entry{
		{ID: 9, Title: "big-win", Project: "alpha", Type: "shipped",
			Impact:    longImpact,
			CreatedAt: time.Date(2026, 7, 1, 10, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2026, 7, 1, 10, 0, 0, 0, time.UTC)},
	}
	opts := ImpactOptions{Scope: "quarter", Filters: "(none)", EntriesInWindow: 1, Now: impactFixedNow}

	md, err := ToImpactMarkdown(entries, opts)
	if err != nil {
		t.Fatalf("markdown: unexpected error: %v", err)
	}
	if !strings.Contains(string(md), longImpact) {
		t.Errorf("markdown must render the impact text in full:\nwant substring: %q\ngot:\n%s", longImpact, md)
	}

	jsonBytes, err := ToImpactJSON(entries, opts)
	if err != nil {
		t.Fatalf("json: unexpected error: %v", err)
	}
	if !strings.Contains(string(jsonBytes), longImpact) {
		t.Errorf("json must render the impact text in full:\nwant substring: %q\ngot:\n%s", longImpact, jsonBytes)
	}
}

// Test 6 — TestToImpactMarkdown_FiltersEchoed.
func TestToImpactMarkdown_FiltersEchoed(t *testing.T) {
	alphaEntries := []storage.Entry{impactFixture[0], impactFixture[3]}
	opts := ImpactOptions{
		Scope:           "quarter",
		Filters:         "--project alpha",
		FiltersJSON:     map[string]string{"project": "alpha"},
		EntriesInWindow: 2,
		Now:             impactFixedNow,
	}

	md, err := ToImpactMarkdown(alphaEntries, opts)
	if err != nil {
		t.Fatalf("markdown: unexpected error: %v", err)
	}
	if !strings.Contains(string(md), "Filters: --project alpha") {
		t.Errorf("expected 'Filters: --project alpha' line in:\n%s", md)
	}

	jsonBytes, err := ToImpactJSON(alphaEntries, opts)
	if err != nil {
		t.Fatalf("json: unexpected error: %v", err)
	}
	var env struct {
		Filters map[string]string `json:"filters"`
	}
	if err := json.Unmarshal(jsonBytes, &env); err != nil {
		t.Fatalf("json unmarshal: %v", err)
	}
	if got := env.Filters["project"]; got != "alpha" {
		t.Errorf("filters.project: got %q, want %q", got, "alpha")
	}
	if len(env.Filters) != 1 {
		t.Errorf("filters: got %v, want exactly {project: alpha}", env.Filters)
	}
}
