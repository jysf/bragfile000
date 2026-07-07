package story

import (
	"encoding/json"
	"strings"
	"testing"
)

// meBundleOpts builds the StoryOptions the me golden asserts: me policy
// threading over storyFixture, the me directive spliced verbatim.
func meBundleOpts(t *testing.T) StoryOptions {
	t.Helper()
	threads := BuildThreads(storyFixture, meThreadOpts)
	return StoryOptions{
		Audience:        "me",
		Scope:           "year",
		Filters:         "(none)",
		FiltersJSON:     nil,
		EntriesInWindow: 6,
		Now:             storyFixedNow,
		Threads:         threads,
		Throughline:     BuildThroughline(threads),
		Directive:       mustDirective(t, "me.md"),
	}
}

func mustDirective(t *testing.T, basename string) string {
	t.Helper()
	b, err := directiveAsset(basename)
	if err != nil {
		t.Fatalf("directiveAsset(%q): %v", basename, err)
	}
	return string(b)
}

// mustDirectiveTrimmed returns the directive with its single trailing
// newline stripped — the markdown bundle appends the directive verbatim
// but the whole document is trailing-newline-trimmed (the directive is
// the last section), so the golden's spliced tail must match that.
func mustDirectiveTrimmed(t *testing.T, basename string) string {
	t.Helper()
	return strings.TrimRight(mustDirective(t, basename), "\n")
}

// Test 1 — TestToStoryMarkdown_MeProfile_FullDocumentGolden (LOAD-BEARING).
func TestToStoryMarkdown_MeProfile_FullDocumentGolden(t *testing.T) {
	got, err := ToStoryMarkdown(meBundleOpts(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := `# Bragfile Story

Generated: 2026-07-06T12:00:00Z
Scope: year
Audience: me
Filters: (none)
Threads: 4
Beats: 6/6

## Threads

### alpha

- ★ 1: alpha-early
  cut p95 login latency 40%
- · 2: alpha-messy

### beta

- ★ 3: beta-one
  onboarding time down to 1 day
- ★ 4: beta-two
  removed the nightly cron entirely

### gamma

- ★ 6: perf-sweep
  shaved 200ms off cold start

### (no project)

- · 5: loose-note

## Throughline (skeleton)

- alpha [initiative]: 2 beats, 1 with impact (2026-02-01 → 2026-04-01)
- beta [initiative]: 2 beats, 2 with impact (2026-03-01 → 2026-05-01)
- gamma [initiative]: 1 beat, 1 with impact (2026-06-15 → 2026-06-15)
- (no project) [initiative]: 1 beat, 0 with impact (2026-06-01 → 2026-06-01)

## Framing directive

` + mustDirectiveTrimmed(t, "me.md")
	if string(got) != want {
		t.Errorf("me markdown golden mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

// Test 2 — TestToStoryMarkdown_ExecProfile_FullDocumentGolden (LOAD-BEARING).
func TestToStoryMarkdown_ExecProfile_FullDocumentGolden(t *testing.T) {
	threads := BuildThreads(storyFixture, execThreadOpts)
	opts := StoryOptions{
		Audience:        "exec",
		Scope:           "year",
		Filters:         "(none)",
		FiltersJSON:     nil,
		EntriesInWindow: 6,
		Now:             storyFixedNow,
		Threads:         threads,
		Throughline:     BuildThroughline(threads),
		Directive:       mustDirective(t, "exec.md"),
	}
	got, err := ToStoryMarkdown(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := `# Bragfile Story

Generated: 2026-07-06T12:00:00Z
Scope: year
Audience: exec
Filters: (none)
Threads: 3
Beats: 4/6

## Threads

### beta

- ★ 3: beta-one
  onboarding time down to 1 day
- ★ 4: beta-two
  removed the nightly cron entirely

### alpha

- ★ 1: alpha-early
  cut p95 login latency 40%

### gamma

- ★ 6: perf-sweep
  shaved 200ms off cold start

## Throughline (skeleton)

- beta [initiative]: 2 beats, 2 with impact (2026-03-01 → 2026-05-01)
- alpha [initiative]: 1 beat, 1 with impact (2026-02-01 → 2026-02-01)
- gamma [initiative]: 1 beat, 1 with impact (2026-06-15 → 2026-06-15)

## Framing directive

` + mustDirectiveTrimmed(t, "exec.md")
	if string(got) != want {
		t.Errorf("exec markdown golden mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

// Test 3 — TestToStoryJSON_MeProfile_ShapeGolden (LOAD-BEARING).
func TestToStoryJSON_MeProfile_ShapeGolden(t *testing.T) {
	got, err := ToStoryJSON(meBundleOpts(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	dir, _ := json.Marshal(mustDirective(t, "me.md"))
	want := `{
  "generated_at": "2026-07-06T12:00:00Z",
  "scope": "year",
  "audience": "me",
  "filters": {},
  "threads": [
    {
      "thread": "alpha",
      "kind": "initiative",
      "span": {
        "first": "2026-02-01T10:00:00Z",
        "last": "2026-04-01T10:00:00Z"
      },
      "beats": [
        {
          "id": 1,
          "title": "alpha-early",
          "project": "alpha",
          "type": "shipped",
          "impact": "cut p95 login latency 40%",
          "is_impact_beat": true,
          "created_at": "2026-02-01T10:00:00Z"
        },
        {
          "id": 2,
          "title": "alpha-messy",
          "project": "alpha",
          "type": "learned",
          "impact": "",
          "is_impact_beat": false,
          "created_at": "2026-04-01T10:00:00Z"
        }
      ]
    },
    {
      "thread": "beta",
      "kind": "initiative",
      "span": {
        "first": "2026-03-01T10:00:00Z",
        "last": "2026-05-01T10:00:00Z"
      },
      "beats": [
        {
          "id": 3,
          "title": "beta-one",
          "project": "beta",
          "type": "shipped",
          "impact": "onboarding time down to 1 day",
          "is_impact_beat": true,
          "created_at": "2026-03-01T10:00:00Z"
        },
        {
          "id": 4,
          "title": "beta-two",
          "project": "beta",
          "type": "shipped",
          "impact": "removed the nightly cron entirely",
          "is_impact_beat": true,
          "created_at": "2026-05-01T10:00:00Z"
        }
      ]
    },
    {
      "thread": "gamma",
      "kind": "initiative",
      "span": {
        "first": "2026-06-15T10:00:00Z",
        "last": "2026-06-15T10:00:00Z"
      },
      "beats": [
        {
          "id": 6,
          "title": "perf-sweep",
          "project": "gamma",
          "type": "shipped",
          "impact": "shaved 200ms off cold start",
          "is_impact_beat": true,
          "created_at": "2026-06-15T10:00:00Z"
        }
      ]
    },
    {
      "thread": "(no project)",
      "kind": "initiative",
      "span": {
        "first": "2026-06-01T10:00:00Z",
        "last": "2026-06-01T10:00:00Z"
      },
      "beats": [
        {
          "id": 5,
          "title": "loose-note",
          "project": "(no project)",
          "type": "fixed",
          "impact": "",
          "is_impact_beat": false,
          "created_at": "2026-06-01T10:00:00Z"
        }
      ]
    }
  ],
  "throughline": {
    "arcs": [
      {
        "thread": "alpha",
        "kind": "initiative",
        "beat_count": 2,
        "impact_beat_count": 1,
        "span": {
          "first": "2026-02-01T10:00:00Z",
          "last": "2026-04-01T10:00:00Z"
        }
      },
      {
        "thread": "beta",
        "kind": "initiative",
        "beat_count": 2,
        "impact_beat_count": 2,
        "span": {
          "first": "2026-03-01T10:00:00Z",
          "last": "2026-05-01T10:00:00Z"
        }
      },
      {
        "thread": "gamma",
        "kind": "initiative",
        "beat_count": 1,
        "impact_beat_count": 1,
        "span": {
          "first": "2026-06-15T10:00:00Z",
          "last": "2026-06-15T10:00:00Z"
        }
      },
      {
        "thread": "(no project)",
        "kind": "initiative",
        "beat_count": 1,
        "impact_beat_count": 0,
        "span": {
          "first": "2026-06-01T10:00:00Z",
          "last": "2026-06-01T10:00:00Z"
        }
      }
    ]
  },
  "framing_directive": ` + string(dir) + `
}`
	if string(got) != want {
		t.Errorf("me JSON golden mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

// Test 4 — TestToStory_EmptyWindow.
func TestToStory_EmptyWindow(t *testing.T) {
	threads := BuildThreads(nil, meThreadOpts)
	opts := StoryOptions{
		Audience:        "me",
		Scope:           "year",
		Filters:         "(none)",
		FiltersJSON:     nil,
		EntriesInWindow: 0,
		Now:             storyFixedNow,
		Threads:         threads,
		Throughline:     BuildThroughline(threads),
		Directive:       mustDirective(t, "me.md"),
	}

	md, err := ToStoryMarkdown(opts)
	if err != nil {
		t.Fatalf("markdown: %v", err)
	}
	wantMD := `# Bragfile Story

Generated: 2026-07-06T12:00:00Z
Scope: year
Audience: me
Filters: (none)
Threads: 0
Beats: 0/0

## Framing directive

` + mustDirectiveTrimmed(t, "me.md")
	if string(md) != wantMD {
		t.Errorf("empty-window markdown mismatch:\n--- got ---\n%s\n--- want ---\n%s", md, wantMD)
	}

	jsonBody, err := ToStoryJSON(opts)
	if err != nil {
		t.Fatalf("json: %v", err)
	}
	var env struct {
		Threads     []json.RawMessage `json:"threads"`
		Throughline struct {
			Arcs []json.RawMessage `json:"arcs"`
		} `json:"throughline"`
		FramingDirective string            `json:"framing_directive"`
		Filters          map[string]string `json:"filters"`
	}
	if err := json.Unmarshal(jsonBody, &env); err != nil {
		t.Fatalf("json unmarshal: %v\n%s", err, jsonBody)
	}
	if env.Threads == nil || len(env.Threads) != 0 {
		t.Errorf("threads: got %v, want []", env.Threads)
	}
	if env.Throughline.Arcs == nil || len(env.Throughline.Arcs) != 0 {
		t.Errorf("throughline.arcs: got %v, want []", env.Throughline.Arcs)
	}
	if env.FramingDirective != mustDirective(t, "me.md") {
		t.Errorf("framing_directive should be the me directive text")
	}
	if env.Filters == nil || len(env.Filters) != 0 {
		t.Errorf("filters: got %v, want {}", env.Filters)
	}
	// Empty-window JSON must contain [] arrays, not null.
	if !strings.Contains(string(jsonBody), `"threads": []`) {
		t.Errorf("expected empty threads array literal:\n%s", jsonBody)
	}
}
