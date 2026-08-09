package export

import (
	"strings"
	"testing"
	"time"

	"github.com/jysf/bragfile000/internal/memory"
	"github.com/jysf/bragfile000/internal/storage"
)

// SOURCE OF TRUTH: SPEC-073 Notes → "The fixture". FOUR packages transcribe
// these 8 rows — internal/memory, internal/export (here), internal/cli, and
// internal/storage. An edit here must be mirrored to all of them, especially
// internal/storage's TestSearch_SPEC073FixtureOrdersAuthQuery: that one would
// keep passing against its own divergent copy while silently no longer pinning
// the bm25 order the CLI actually queries.
func memoryFixtureEntries() []storage.Entry {
	mustTime := func(s string) time.Time {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			panic(err)
		}
		return t.UTC()
	}
	return []storage.Entry{
		{ID: 8, CreatedAt: mustTime("2026-08-07T09:15:00Z"), Project: "bragfile", Type: "shipped", Title: "MCP list filter parity", Impact: "agents can ask for a bounded window in one call"},
		{ID: 7, CreatedAt: mustTime("2026-08-05T17:40:00Z"), Project: "orbit", Type: "learned", Title: "Retry storms need jitter"},
		{ID: 6, CreatedAt: mustTime("2026-08-01T11:02:00Z"), Project: "bragfile", Type: "shipped", Title: "Sparkline pulse command"},
		{ID: 5, CreatedAt: mustTime("2026-07-20T08:00:00Z"), Title: "Read the auth spec"},
		{ID: 4, CreatedAt: mustTime("2026-06-11T14:25:00Z"), Project: "orbit", Type: "shipped", Title: "Auth refactor for the gateway", Impact: "cut p99 login latency from 600ms to 120ms"},
		{ID: 3, CreatedAt: mustTime("2026-05-02T19:30:00Z"), Project: "bragfile", Type: "learned", Title: "FTS5 triggers fight attach"},
		{ID: 2, CreatedAt: mustTime("2026-03-14T10:10:00Z"), Project: "orbit", Type: "shipped", Title: "Auth token rotation", Impact: "removed the last shared secret"},
		{ID: 1, CreatedAt: mustTime("2026-01-09T21:45:00Z"), Type: "learned", Title: "Auth is mostly caching"},
	}
}

var memoryFixtureNow = time.Date(2026, 8, 8, 12, 0, 0, 0, time.UTC)

const memoryGolden1 = `# Bragfile Memory

Generated: 2026-08-08T12:00:00Z
Scope: lifetime
Filters: --query auth --project orbit
Entries: 8

## Slice

- 4 2026-06-11 [orbit/shipped] Auth refactor for the gateway — cut p99 login latency from 600ms to 120ms
- 2 2026-03-14 [orbit/shipped] Auth token rotation — removed the last shared secret
- 7 2026-08-05 [orbit/learned] Retry storms need jitter
- 5 2026-07-20 [-/-] Read the auth spec
- 1 2026-01-09 [-/learned] Auth is mostly caching
- 8 2026-08-07 [bragfile/shipped] MCP list filter parity — agents can ask for a bounded window in one call
- 6 2026-08-01 [bragfile/shipped] Sparkline pulse command
- 3 2026-05-02 [bragfile/learned] FTS5 triggers fight attach

## Budget

- Budget: 2000 tokens
- Estimated: 143 tokens
- Included: 8
- Skipped: 0`

const memoryGolden2 = `# Bragfile Memory

Generated: 2026-08-08T12:00:00Z
Scope: lifetime
Filters: (none)
Entries: 8

## Slice

- 8 2026-08-07 [bragfile/shipped] MCP list filter parity — agents can ask for a bounded window in one call
- 7 2026-08-05 [orbit/learned] Retry storms need jitter
- 6 2026-08-01 [bragfile/shipped] Sparkline pulse command
- 5 2026-07-20 [-/-] Read the auth spec
- 4 2026-06-11 [orbit/shipped] Auth refactor for the gateway — cut p99 login latency from 600ms to 120ms
- 3 2026-05-02 [bragfile/learned] FTS5 triggers fight attach
- 2 2026-03-14 [orbit/shipped] Auth token rotation — removed the last shared secret
- 1 2026-01-09 [-/learned] Auth is mostly caching

## Budget

- Budget: 2000 tokens
- Estimated: 143 tokens
- Included: 8
- Skipped: 0`

const memoryGolden3 = `# Bragfile Memory

Generated: 2026-08-08T12:00:00Z
Scope: lifetime
Filters: (none)
Entries: 8

## Slice

- 8 2026-08-07 [bragfile/shipped] MCP list filter parity — agents can ask for a bounded window in one call
- 5 2026-07-20 [-/-] Read the auth spec

## Budget

- Budget: 40 tokens
- Estimated: 37 tokens
- Included: 2
- Skipped: 6`

const memoryGolden4 = `# Bragfile Memory

Generated: 2026-08-08T12:00:00Z
Scope: lifetime
Filters: (none)
Entries: 0`

const memoryJSONGolden = `{
  "generated_at": "2026-08-08T12:00:00Z",
  "scope": "lifetime",
  "filters": {
    "project": "orbit",
    "query": "auth"
  },
  "entries": 8,
  "budget": 2000,
  "estimated_tokens": 143,
  "included": 8,
  "skipped": 0,
  "slice": [
    {
      "id": 4,
      "created_at": "2026-06-11T14:25:00Z",
      "project": "orbit",
      "type": "shipped",
      "title": "Auth refactor for the gateway",
      "impact": "cut p99 login latency from 600ms to 120ms",
      "rank": 1,
      "score": 0.047139,
      "tokens": 27
    },
    {
      "id": 2,
      "created_at": "2026-03-14T10:10:00Z",
      "project": "orbit",
      "type": "shipped",
      "title": "Auth token rotation",
      "impact": "removed the last shared secret",
      "rank": 2,
      "score": 0.046671,
      "tokens": 22
    },
    {
      "id": 7,
      "created_at": "2026-08-05T17:40:00Z",
      "project": "orbit",
      "type": "learned",
      "title": "Retry storms need jitter",
      "impact": "",
      "rank": 3,
      "score": 0.032522,
      "tokens": 14
    },
    {
      "id": 5,
      "created_at": "2026-07-20T08:00:00Z",
      "project": "",
      "type": "",
      "title": "Read the auth spec",
      "impact": "",
      "rank": 4,
      "score": 0.032018,
      "tokens": 10
    },
    {
      "id": 1,
      "created_at": "2026-01-09T21:45:00Z",
      "project": "",
      "type": "learned",
      "title": "Auth is mostly caching",
      "impact": "",
      "rank": 5,
      "score": 0.030835,
      "tokens": 13
    },
    {
      "id": 8,
      "created_at": "2026-08-07T09:15:00Z",
      "project": "bragfile",
      "type": "shipped",
      "title": "MCP list filter parity",
      "impact": "agents can ask for a bounded window in one call",
      "rank": 6,
      "score": 0.016393,
      "tokens": 27
    },
    {
      "id": 6,
      "created_at": "2026-08-01T11:02:00Z",
      "project": "bragfile",
      "type": "shipped",
      "title": "Sparkline pulse command",
      "impact": "",
      "rank": 7,
      "score": 0.015873,
      "tokens": 15
    },
    {
      "id": 3,
      "created_at": "2026-05-02T19:30:00Z",
      "project": "bragfile",
      "type": "learned",
      "title": "FTS5 triggers fight attach",
      "impact": "",
      "rank": 8,
      "score": 0.015152,
      "tokens": 15
    }
  ]
}`

func blendedResult() memory.Result {
	return memory.Slice(memoryFixtureEntries(), memory.Options{
		Query:   "auth",
		Project: "orbit",
		Matched: []int64{5, 1, 2, 4},
		Budget:  2000,
	})
}

func blendedOptions() MemoryOptions {
	return MemoryOptions{
		Filters:     "--query auth --project orbit",
		FiltersJSON: map[string]string{"query": "auth", "project": "orbit"},
		Now:         memoryFixtureNow,
	}
}

// TestToMemoryMarkdown_BlendedGolden is the load-bearing contract test:
// byte-exact equality against Golden 1.
func TestToMemoryMarkdown_BlendedGolden(t *testing.T) {
	got, err := ToMemoryMarkdown(blendedResult(), blendedOptions())
	if err != nil {
		t.Fatalf("ToMemoryMarkdown: %v", err)
	}
	if string(got) != memoryGolden1 {
		t.Errorf("markdown mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, memoryGolden1)
	}
	if len(got) != 778 {
		t.Errorf("byte length = %d, want 778", len(got))
	}
}

// TestToMemoryMarkdown_RecencyGolden: byte-exact against Golden 2 (no query,
// no project).
func TestToMemoryMarkdown_RecencyGolden(t *testing.T) {
	result := memory.Slice(memoryFixtureEntries(), memory.Options{Budget: 2000})
	opts := MemoryOptions{Filters: "(none)", FiltersJSON: map[string]string{}, Now: memoryFixtureNow}
	got, err := ToMemoryMarkdown(result, opts)
	if err != nil {
		t.Fatalf("ToMemoryMarkdown: %v", err)
	}
	if string(got) != memoryGolden2 {
		t.Errorf("markdown mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, memoryGolden2)
	}
}

// TestToMemoryMarkdown_TightBudgetGolden: byte-exact against Golden 3
// (--budget 40), proving skip-and-continue at the rendering layer.
func TestToMemoryMarkdown_TightBudgetGolden(t *testing.T) {
	result := memory.Slice(memoryFixtureEntries(), memory.Options{Budget: 40})
	opts := MemoryOptions{Filters: "(none)", FiltersJSON: map[string]string{}, Now: memoryFixtureNow}
	got, err := ToMemoryMarkdown(result, opts)
	if err != nil {
		t.Fatalf("ToMemoryMarkdown: %v", err)
	}
	if string(got) != memoryGolden3 {
		t.Errorf("markdown mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, memoryGolden3)
	}
}

// TestToMemoryMarkdown_EmptyCorpusGolden: byte-exact against Golden 4 — the
// document ends after Entries: 0, no ## Slice, no ## Budget (DEC-014 part 4).
func TestToMemoryMarkdown_EmptyCorpusGolden(t *testing.T) {
	result := memory.Slice(nil, memory.Options{Budget: 2000})
	opts := MemoryOptions{Filters: "(none)", FiltersJSON: map[string]string{}, Now: memoryFixtureNow}
	got, err := ToMemoryMarkdown(result, opts)
	if err != nil {
		t.Fatalf("ToMemoryMarkdown: %v", err)
	}
	if string(got) != memoryGolden4 {
		t.Errorf("markdown mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, memoryGolden4)
	}
}

// TestToMemoryMarkdown_ZeroIncludedStillRendersBothSections: Budget 1 over
// the non-empty fixture: ## Slice present with no bullets, ## Budget present
// reporting Included: 0 / Skipped: 8. Line-by-line equality on the headings,
// not strings.Contains (AGENTS.md §9 heading addendum).
func TestToMemoryMarkdown_ZeroIncludedStillRendersBothSections(t *testing.T) {
	result := memory.Slice(memoryFixtureEntries(), memory.Options{Budget: 1})
	opts := MemoryOptions{Filters: "(none)", FiltersJSON: map[string]string{}, Now: memoryFixtureNow}
	got, err := ToMemoryMarkdown(result, opts)
	if err != nil {
		t.Fatalf("ToMemoryMarkdown: %v", err)
	}
	var hasSlice, hasBudget, hasIncludedZero bool
	for _, ln := range strings.Split(string(got), "\n") {
		if ln == "## Slice" {
			hasSlice = true
		}
		if ln == "## Budget" {
			hasBudget = true
		}
		if ln == "- Included: 0" {
			hasIncludedZero = true
		}
	}
	if !hasSlice {
		t.Errorf("missing '## Slice' line:\n%s", got)
	}
	if !hasBudget {
		t.Errorf("missing '## Budget' line:\n%s", got)
	}
	if !hasIncludedZero {
		t.Errorf("missing '- Included: 0' line:\n%s", got)
	}
	if strings.Contains(string(got), "- 8 2026") {
		t.Errorf("expected no entry bullets, got:\n%s", got)
	}
}

// TestToMemoryJSON_BlendedGolden: byte-exact against the JSON golden,
// including score rounded to 6 dp and the flat DEC-014 key order.
func TestToMemoryJSON_BlendedGolden(t *testing.T) {
	got, err := ToMemoryJSON(blendedResult(), blendedOptions())
	if err != nil {
		t.Fatalf("ToMemoryJSON: %v", err)
	}
	if string(got) != memoryJSONGolden {
		t.Errorf("JSON mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, memoryJSONGolden)
	}
}

// TestToMemoryJSON_EmptyCorpusEmitsEveryKey: entries/budget/estimated_tokens/
// included/skipped are 0, slice is [] (non-nil, never null), filters is {}
// (DEC-014 part 4).
func TestToMemoryJSON_EmptyCorpusEmitsEveryKey(t *testing.T) {
	got, err := ToMemoryJSON(memory.Result{}, MemoryOptions{Filters: "(none)", Now: memoryFixtureNow})
	if err != nil {
		t.Fatalf("ToMemoryJSON: %v", err)
	}
	s := string(got)
	for _, want := range []string{
		`"entries": 0`,
		`"budget": 0`,
		`"estimated_tokens": 0`,
		`"included": 0`,
		`"skipped": 0`,
		`"slice": []`,
		`"filters": {}`,
	} {
		if !strings.Contains(s, want) {
			t.Errorf("missing %q in:\n%s", want, s)
		}
	}
	if strings.Contains(s, `"slice": null`) {
		t.Errorf("slice is null, want []: %s", s)
	}
}

// TestToMemoryJSON_FiltersEchoQueryAndProject: filters is
// {"project":"orbit","query":"auth"} (alphabetical, per Go's encoder) while
// the markdown line is --query auth --project orbit (declared order) — the
// documented DEC-014 markdown/JSON ordering asymmetry.
func TestToMemoryJSON_FiltersEchoQueryAndProject(t *testing.T) {
	got, err := ToMemoryJSON(blendedResult(), blendedOptions())
	if err != nil {
		t.Fatalf("ToMemoryJSON: %v", err)
	}
	want := "\"filters\": {\n    \"project\": \"orbit\",\n    \"query\": \"auth\"\n  },"
	if !strings.Contains(string(got), want) {
		t.Errorf("expected alphabetical filters object:\n%s\nin:\n%s", want, got)
	}
	md, err := ToMemoryMarkdown(blendedResult(), blendedOptions())
	if err != nil {
		t.Fatalf("ToMemoryMarkdown: %v", err)
	}
	if !strings.Contains(string(md), "Filters: --query auth --project orbit") {
		t.Errorf("expected declared-order markdown filters line, got:\n%s", md)
	}
}

// TestToMemoryMarkdown_TrailingNewlineStripped: the returned bytes do not
// end in "\n" (matches every sibling renderer).
func TestToMemoryMarkdown_TrailingNewlineStripped(t *testing.T) {
	got, err := ToMemoryMarkdown(blendedResult(), blendedOptions())
	if err != nil {
		t.Fatalf("ToMemoryMarkdown: %v", err)
	}
	if strings.HasSuffix(string(got), "\n") {
		t.Errorf("markdown ends with a newline: %q", got)
	}
}
