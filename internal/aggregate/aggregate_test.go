package aggregate

import (
	"reflect"
	"testing"
	"time"

	"github.com/jysf/bragfile000/internal/storage"
)

// TestByType_DESCByCountAlphaTiebreak locks the data-layer ordering
// rule on counts-by-type: DESC by count, alphabetical-ASC tiebreak.
// Pairs locked decision 2 (aggregate is the data-layer source) and
// reuses DEC-013's count-ordering convention.
func TestByType_DESCByCountAlphaTiebreak(t *testing.T) {
	t.Run("distinct_counts", func(t *testing.T) {
		input := []storage.Entry{
			{Type: "shipped"}, {Type: "shipped"}, {Type: "shipped"},
			{Type: "learned"}, {Type: "learned"},
			{Type: "fixed"},
		}
		want := []TypeCount{
			{Type: "shipped", Count: 3},
			{Type: "learned", Count: 2},
			{Type: "fixed", Count: 1},
		}
		got := ByType(input)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("ByType distinct_counts mismatch\nwant: %#v\ngot:  %#v", want, got)
		}
	})

	t.Run("alpha_tiebreak", func(t *testing.T) {
		input := []storage.Entry{
			{Type: "zebra"}, {Type: "zebra"},
			{Type: "alpha"}, {Type: "alpha"},
			{Type: "fixed"},
		}
		want := []TypeCount{
			{Type: "alpha", Count: 2},
			{Type: "zebra", Count: 2},
			{Type: "fixed", Count: 1},
		}
		got := ByType(input)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("ByType alpha_tiebreak mismatch\nwant: %#v\ngot:  %#v", want, got)
		}
	})
}

// TestByProject_NoProjectKeyForcedLast locks the (no project) sentinel
// and forced-last rule on counts-by-project — even when (no project)
// has the highest count it must still appear last. Pairs locked
// decision 3 ((no project) sentinel + forced-last rule).
func TestByProject_NoProjectKeyForcedLast(t *testing.T) {
	input := []storage.Entry{
		{Project: ""}, {Project: ""}, {Project: ""}, // 3 of (no project)
		{Project: "alpha"}, {Project: "alpha"}, // 2 of alpha
		{Project: "beta"}, // 1 of beta
	}
	want := []ProjectCount{
		{Project: "alpha", Count: 2},
		{Project: "beta", Count: 1},
		{Project: NoProjectKey, Count: 3},
	}
	got := ByProject(input)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ByProject mismatch\nwant: %#v\ngot:  %#v", want, got)
	}
}

// sharedFixture mirrors the spec's load-bearing fixture used by tests
// 5/6 in summary_test.go (the export package). Replicated here so the
// aggregate-layer test exercises the same shape.
var sharedFixture = []storage.Entry{
	{
		ID: 1, Title: "alpha-old",
		Description: "old alpha",
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
		Type:      "shipped",
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

// TestGroupForHighlights_ChronoASCWithNoProjectLast locks the
// highlights ordering: groups in alpha-ASC with NoProjectKey last;
// within each group entries chrono-ASC with ID as tie-break. Pairs
// locked decisions 2 and 3 plus the AGENTS.md §9 SPEC-002 monotonic-
// tiebreak rule.
func TestGroupForHighlights_ChronoASCWithNoProjectLast(t *testing.T) {
	t.Run("shared_fixture", func(t *testing.T) {
		want := []ProjectHighlights{
			{Project: "alpha", Entries: []EntryRef{
				{ID: 1, Title: "alpha-old"},
				{ID: 4, Title: "alpha-new"},
			}},
			{Project: "beta", Entries: []EntryRef{
				{ID: 2, Title: "beta-mid"},
			}},
			{Project: "gamma", Entries: []EntryRef{
				{ID: 5, Title: "gamma-only"},
			}},
			{Project: NoProjectKey, Entries: []EntryRef{
				{ID: 3, Title: "unbound-mid"},
			}},
		}
		got := GroupForHighlights(sharedFixture)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("GroupForHighlights shared_fixture mismatch\nwant: %#v\ngot:  %#v", want, got)
		}
	})

	t.Run("id_tiebreak_on_same_timestamp", func(t *testing.T) {
		ts := time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC)
		tieFixture := []storage.Entry{
			{ID: 99, Title: "later-id-same-time", Project: "alpha", CreatedAt: ts},
			{ID: 7, Title: "earlier-id-same-time", Project: "alpha", CreatedAt: ts},
		}
		want := []ProjectHighlights{
			{Project: "alpha", Entries: []EntryRef{
				{ID: 7, Title: "earlier-id-same-time"},
				{ID: 99, Title: "later-id-same-time"},
			}},
		}
		got := GroupForHighlights(tieFixture)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("GroupForHighlights id_tiebreak mismatch\nwant: %#v\ngot:  %#v", want, got)
		}
	})
}

// TestAggregate_EmptyInputReturnsNonNilEmptySlice locks the empty-
// input contract on all three aggregate functions: each must return
// a non-nil empty slice so the JSON renderer marshals it as [] not
// null. Pairs locked decision 1 part (4) (empty-state arrays are []
// in JSON, never null).
func TestAggregate_EmptyInputReturnsNonNilEmptySlice(t *testing.T) {
	cases := []struct {
		name   string
		nilIn  func() any
		zeroIn func() any
	}{
		{
			name:   "ByType",
			nilIn:  func() any { return ByType(nil) },
			zeroIn: func() any { return ByType([]storage.Entry{}) },
		},
		{
			name:   "ByProject",
			nilIn:  func() any { return ByProject(nil) },
			zeroIn: func() any { return ByProject([]storage.Entry{}) },
		},
		{
			name:   "GroupForHighlights",
			nilIn:  func() any { return GroupForHighlights(nil) },
			zeroIn: func() any { return GroupForHighlights([]storage.Entry{}) },
		},
	}
	for _, tc := range cases {
		t.Run(tc.name+"_nil_input", func(t *testing.T) {
			got := tc.nilIn()
			v := reflect.ValueOf(got)
			if v.Kind() != reflect.Slice {
				t.Fatalf("%s nil-input: expected slice, got %T", tc.name, got)
			}
			if v.IsNil() {
				t.Fatalf("%s nil-input: expected non-nil empty slice, got nil", tc.name)
			}
			if v.Len() != 0 {
				t.Fatalf("%s nil-input: expected len 0, got %d", tc.name, v.Len())
			}
		})
		t.Run(tc.name+"_empty_slice_input", func(t *testing.T) {
			got := tc.zeroIn()
			v := reflect.ValueOf(got)
			if v.Kind() != reflect.Slice {
				t.Fatalf("%s empty-input: expected slice, got %T", tc.name, got)
			}
			if v.IsNil() {
				t.Fatalf("%s empty-input: expected non-nil empty slice, got nil", tc.name)
			}
			if v.Len() != 0 {
				t.Fatalf("%s empty-input: expected len 0, got %d", tc.name, v.Len())
			}
		})
	}
}

// TestGroupEntriesByProject_OrderingAndIDTiebreak locks the SPEC-019
// helper: alpha-ASC group order with NoProjectKey forced last; chrono-
// ASC within group with ID as tie-break (AGENTS.md §9 SPEC-002 rule).
// Each Entries slice carries the FULL storage.Entry, not the EntryRef
// projection.
func TestGroupEntriesByProject_OrderingAndIDTiebreak(t *testing.T) {
	t.Run("shared_fixture", func(t *testing.T) {
		want := []ProjectEntryGroup{
			{Project: "alpha", Entries: []storage.Entry{
				sharedFixture[0], // ID 1
				sharedFixture[3], // ID 4
			}},
			{Project: "beta", Entries: []storage.Entry{
				sharedFixture[1], // ID 2
			}},
			{Project: "gamma", Entries: []storage.Entry{
				sharedFixture[4], // ID 5
			}},
			{Project: NoProjectKey, Entries: []storage.Entry{
				sharedFixture[2], // ID 3
			}},
		}
		got := GroupEntriesByProject(sharedFixture)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("GroupEntriesByProject shared_fixture mismatch\nwant: %#v\ngot:  %#v", want, got)
		}
	})

	t.Run("id_tiebreak", func(t *testing.T) {
		ts := time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC)
		tieFixture := []storage.Entry{
			{ID: 99, Title: "later-id-same-time", Project: "alpha", CreatedAt: ts},
			{ID: 7, Title: "earlier-id-same-time", Project: "alpha", CreatedAt: ts},
		}
		want := []ProjectEntryGroup{
			{Project: "alpha", Entries: []storage.Entry{
				{ID: 7, Title: "earlier-id-same-time", Project: "alpha", CreatedAt: ts},
				{ID: 99, Title: "later-id-same-time", Project: "alpha", CreatedAt: ts},
			}},
		}
		got := GroupEntriesByProject(tieFixture)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("GroupEntriesByProject id_tiebreak mismatch\nwant: %#v\ngot:  %#v", want, got)
		}
	})
}

// TestGroupEntriesByProject_EmptyInputReturnsNonNilEmptySlice locks the
// empty-state contract on the SPEC-019 helper: nil or empty input must
// return a non-nil empty slice so JSON marshaling renders [] not null.
func TestGroupEntriesByProject_EmptyInputReturnsNonNilEmptySlice(t *testing.T) {
	t.Run("nil_input", func(t *testing.T) {
		got := GroupEntriesByProject(nil)
		if got == nil {
			t.Fatalf("expected non-nil empty slice, got nil")
		}
		if len(got) != 0 {
			t.Fatalf("expected len 0, got %d", len(got))
		}
	})
	t.Run("empty_slice_input", func(t *testing.T) {
		got := GroupEntriesByProject([]storage.Entry{})
		if got == nil {
			t.Fatalf("expected non-nil empty slice, got nil")
		}
		if len(got) != 0 {
			t.Fatalf("expected len 0, got %d", len(got))
		}
	})
}
