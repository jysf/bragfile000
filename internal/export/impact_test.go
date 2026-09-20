package export

import (
	"encoding/json"
	"reflect"
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
  ],
  "failures_by_project": []
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
	if string(env["failures_by_project"]) != "[]" {
		t.Errorf("failures_by_project: got %s, want []", env["failures_by_project"])
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

// impactFailureFixture is impactFixture plus three rows typed "failed" — the
// literal DEC-049 persists, spelled out rather than taken from
// aggregate.FailureType so these tests also fail if the constant drifts from
// the stored value. id 6 is an alpha failure WITH an impact, dated between
// alpha's two wins, so it must leave their group; id 7 is a failure with NO
// impact, so it appears nowhere (impact-first, DEC-028 choice 3); id 8 puts
// gamma in the body only through a failure, so gamma is a failures-only
// project. 8 in window, 5 with impact: 3 in ## Impact, 2 in ## What didn't
// work.
var impactFailureFixture = append(append([]storage.Entry{}, impactFixture...),
	storage.Entry{ID: 6, Title: "pool-dead-end",
		Project: "alpha", Type: "failed",
		Impact:    "cost two days and produced nothing reusable",
		CreatedAt: time.Date(2026, 7, 2, 11, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 7, 2, 11, 0, 0, 0, time.UTC)},
	storage.Entry{ID: 7, Title: "retry-noimpact",
		Project: "delta", Type: "failed",
		Impact:    "", // a failure with no impact → counted, not shown
		CreatedAt: time.Date(2026, 7, 3, 11, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 7, 3, 11, 0, 0, 0, time.UTC)},
	storage.Entry{ID: 8, Title: "vendor-sdk-dead-end",
		Project: "gamma", Type: "failed",
		Impact:    "ruled out the vendor SDK",
		CreatedAt: time.Date(2026, 7, 5, 11, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 7, 5, 11, 0, 0, 0, time.UTC)},
)

var impactFailureOpts = ImpactOptions{
	Scope:           "quarter",
	Filters:         "(none)",
	EntriesInWindow: 8,
	Now:             impactFixedNow,
}

// TestToImpactMarkdown_FailureSectionGolden (LOAD-BEARING, SPEC-086 LD2/LD3).
// A recorded failure with an impact leaves ## Impact and renders under
// ## What didn't work, in the same per-entry shape; the Entries: tally is
// unchanged in meaning — 5 is the with-impact subset, which both sections show.
func TestToImpactMarkdown_FailureSectionGolden(t *testing.T) {
	got, err := ToImpactMarkdown(impactFailureFixture, impactFailureOpts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := `# Bragfile Impact

Generated: 2026-07-06T12:00:00Z
Scope: quarter
Filters: (none)
Entries: 5/8 with impact

## Impact

### alpha

- 1: alpha-old
  cut p95 login latency 40%
- 4: alpha-new
  removed the nightly cron entirely

### beta

- 2: beta-mid
  onboarding time down to 1 day

## What didn't work

### alpha

- 6: pool-dead-end
  cost two days and produced nothing reusable

### gamma

- 8: vendor-sdk-dead-end
  ruled out the vendor SDK`
	if string(got) != want {
		t.Errorf("markdown golden mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

// TestToImpactJSON_FailureSectionGolden (LOAD-BEARING, SPEC-086 LD4). The
// failures leave impact_by_project and arrive in failures_by_project, the
// same group shape and 4-key projection; counts_by_project still counts both.
func TestToImpactJSON_FailureSectionGolden(t *testing.T) {
	got, err := ToImpactJSON(impactFailureFixture, impactFailureOpts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := `{
  "generated_at": "2026-07-06T12:00:00Z",
  "scope": "quarter",
  "filters": {},
  "entries_in_window": 8,
  "entries_with_impact": 5,
  "counts_by_project": {
    "alpha": 3,
    "beta": 1,
    "gamma": 1
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
  ],
  "failures_by_project": [
    {
      "project": "alpha",
      "entries": [
        {
          "id": 6,
          "title": "pool-dead-end",
          "project": "alpha",
          "impact": "cost two days and produced nothing reusable"
        }
      ]
    },
    {
      "project": "gamma",
      "entries": [
        {
          "id": 8,
          "title": "vendor-sdk-dead-end",
          "project": "gamma",
          "impact": "ruled out the vendor SDK"
        }
      ]
    }
  ]
}`
	if string(got) != want {
		t.Errorf("json golden mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

// TestToImpactJSON_CountsByProjectSpansBothSections pins SPEC-086 LD3, the
// DEC-048 obligation: no existing count changes what it counts. DEC-028
// defines counts_by_project over the with-impact subset, so it keeps summing
// to entries_with_impact, and each project's count is its rows across BOTH
// sections. Deriving the map from the narrowed impact_by_project loop — the
// obvious shortcut — makes alpha 2 and drops gamma, and fails here by name.
func TestToImpactJSON_CountsByProjectSpansBothSections(t *testing.T) {
	jsonBytes, err := ToImpactJSON(impactFailureFixture, impactFailureOpts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var env struct {
		EntriesWithImpact int            `json:"entries_with_impact"`
		CountsByProject   map[string]int `json:"counts_by_project"`
		ImpactByProject   []struct {
			Project string            `json:"project"`
			Entries []json.RawMessage `json:"entries"`
		} `json:"impact_by_project"`
		FailuresByProject []struct {
			Project string            `json:"project"`
			Entries []json.RawMessage `json:"entries"`
		} `json:"failures_by_project"`
	}
	if err := json.Unmarshal(jsonBytes, &env); err != nil {
		t.Fatalf("json unmarshal: %v", err)
	}
	rows := map[string]int{}
	for _, g := range env.ImpactByProject {
		rows[g.Project] += len(g.Entries)
	}
	for _, g := range env.FailuresByProject {
		rows[g.Project] += len(g.Entries)
	}
	sum := 0
	for p, n := range env.CountsByProject {
		sum += n
		if rows[p] != n {
			t.Errorf("counts_by_project[%q] = %d, but the two sections hold %d rows for it", p, n, rows[p])
		}
	}
	if len(rows) != len(env.CountsByProject) {
		t.Errorf("counts_by_project has %d projects, the two sections have %d", len(env.CountsByProject), len(rows))
	}
	if sum != env.EntriesWithImpact || sum != 5 {
		t.Errorf("counts_by_project sums to %d; entries_with_impact is %d; want both 5", sum, env.EntriesWithImpact)
	}
}

// TestToImpactMarkdown_SectionsRenderOnlyWhenNonEmpty pins SPEC-086 LD5 (Fork
// D) on impact: each section heading appears only when it has an entry. A
// clean window grows no "## What didn't work"; a failures-only window grows no
// bare "## Impact". Line equality, not substring (AGENTS.md §9, SPEC-015).
func TestToImpactMarkdown_SectionsRenderOnlyWhenNonEmpty(t *testing.T) {
	headings := func(md []byte) []string {
		var out []string
		for _, ln := range strings.Split(string(md), "\n") {
			if strings.HasPrefix(ln, "## ") {
				out = append(out, ln)
			}
		}
		return out
	}
	failuresOnly := []storage.Entry{impactFailureFixture[5], impactFailureFixture[7]}
	cases := []struct {
		name    string
		entries []storage.Entry
		want    []string
	}{
		{"no failures", impactFixture, []string{"## Impact"}},
		{"failures only", failuresOnly, []string{"## What didn't work"}},
		{"both", impactFailureFixture, []string{"## Impact", "## What didn't work"}},
	}
	for _, c := range cases {
		opts := impactFailureOpts
		opts.EntriesInWindow = len(c.entries)
		md, err := ToImpactMarkdown(c.entries, opts)
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", c.name, err)
		}
		if got := headings(md); !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s: ## headings = %q, want %q\n%s", c.name, got, c.want, md)
		}
	}
}
