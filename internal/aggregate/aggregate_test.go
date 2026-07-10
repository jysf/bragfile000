package aggregate

import (
	"reflect"
	"testing"
	"time"
	_ "time/tzdata" // embeds the IANA DB in the test binary only, keeping LoadLocation hermetic in CI

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

// entryAt is a compact constructor for streak/span tests — only the
// fields these helpers actually read (CreatedAt, optionally ID/Title)
// are populated.
func entryAt(year, month, day, hour int) storage.Entry {
	return storage.Entry{
		CreatedAt: time.Date(year, time.Month(month), day, hour, 0, 0, 0, time.UTC),
	}
}

// TestStreak_CurrentAndLongest locks the DEC-022 streak semantics:
// current counts the consecutive local-day run ending on TODAY or
// YESTERDAY (alive-through-yesterday), 0 only after two empty days;
// longest is the longest consecutive local-day run; same-day entries
// dedupe to one day. now is injected (instant + zone); no time.Sleep.
// These subtests keep now in time.UTC, so local day == UTC day — the
// alive-through-yesterday axis is exercised here; the local-day and DST
// axes are exercised by the two dedicated tests below.
func TestStreak_CurrentAndLongest(t *testing.T) {
	now := time.Date(2026, 4, 25, 12, 0, 0, 0, time.UTC)

	// ● preservation: today has an entry → count today and walk back.
	t.Run("today_has_entries", func(t *testing.T) {
		entries := []storage.Entry{
			entryAt(2026, 4, 23, 10), entryAt(2026, 4, 24, 10), entryAt(2026, 4, 25, 10),
		}
		current, longest := Streak(entries, now)
		if current != 3 || longest != 3 {
			t.Errorf("today_has_entries: got (%d,%d), want (3,3)", current, longest)
		}
	})

	// ▲ new: today empty, yesterday (4/24) present → streak ALIVE (was 0).
	t.Run("today_empty_yesterday_alive", func(t *testing.T) {
		entries := []storage.Entry{
			entryAt(2026, 4, 22, 10), entryAt(2026, 4, 23, 10), entryAt(2026, 4, 24, 10),
		}
		current, longest := Streak(entries, now)
		if current != 3 || longest != 3 {
			t.Errorf("today_empty_yesterday_alive: got (%d,%d), want (3,3)", current, longest)
		}
	})

	// ▲ new (the BUG's canonical case): run ending yesterday, now=today,
	// current == run length (not 0).
	t.Run("run_ending_yesterday_now_today", func(t *testing.T) {
		entries := []storage.Entry{entryAt(2026, 4, 23, 10), entryAt(2026, 4, 24, 10)}
		current, longest := Streak(entries, now)
		if current != 2 || longest != 2 {
			t.Errorf("run_ending_yesterday_now_today: got (%d,%d), want (2,2)", current, longest)
		}
	})

	// ▲ new: a single entry dated yesterday still reads current 1.
	t.Run("single_entry_yesterday", func(t *testing.T) {
		entries := []storage.Entry{entryAt(2026, 4, 24, 10)}
		current, longest := Streak(entries, now)
		if current != 1 || longest != 1 {
			t.Errorf("single_entry_yesterday: got (%d,%d), want (1,1)", current, longest)
		}
	})

	// ● preservation: two empty days (4/24 AND 4/25) → current 0. The fix
	// grants one day of grace, not immortality.
	t.Run("streak_dead_after_two_empty_days", func(t *testing.T) {
		entries := []storage.Entry{
			entryAt(2026, 4, 21, 10), entryAt(2026, 4, 22, 10), entryAt(2026, 4, 23, 10),
		}
		current, longest := Streak(entries, now)
		if current != 0 || longest != 3 {
			t.Errorf("streak_dead_after_two_empty_days: got (%d,%d), want (0,3)", current, longest)
		}
	})

	// ▲ changed: today empty, yesterday 4/24 present → current 2 (was 0);
	// longest preserved at 5 (the 4/10–4/14 run).
	t.Run("gap_mid_corpus_longest", func(t *testing.T) {
		entries := []storage.Entry{
			entryAt(2026, 4, 10, 10), entryAt(2026, 4, 11, 10), entryAt(2026, 4, 12, 10),
			entryAt(2026, 4, 13, 10), entryAt(2026, 4, 14, 10),
			entryAt(2026, 4, 23, 10), entryAt(2026, 4, 24, 10),
		}
		current, longest := Streak(entries, now)
		if current != 2 || longest != 5 {
			t.Errorf("gap_mid_corpus_longest: got (%d,%d), want (2,5)", current, longest)
		}
	})

	// ● preservation.
	t.Run("single_entry_today", func(t *testing.T) {
		entries := []storage.Entry{entryAt(2026, 4, 25, 10)}
		current, longest := Streak(entries, now)
		if current != 1 || longest != 1 {
			t.Errorf("single_entry_today: got (%d,%d), want (1,1)", current, longest)
		}
	})

	// ● preservation: same-day multiple entries dedupe to one streak day.
	t.Run("multiple_entries_same_day", func(t *testing.T) {
		entries := []storage.Entry{
			entryAt(2026, 4, 25, 8), entryAt(2026, 4, 25, 12), entryAt(2026, 4, 25, 16),
		}
		current, longest := Streak(entries, now)
		if current != 1 || longest != 1 {
			t.Errorf("multiple_entries_same_day: got (%d,%d), want (1,1)", current, longest)
		}
	})

	// ● preservation.
	t.Run("empty_corpus", func(t *testing.T) {
		current, longest := Streak(nil, now)
		if current != 0 || longest != 0 {
			t.Errorf("empty_corpus: got (%d,%d), want (0,0)", current, longest)
		}
	})
}

// TestStreak_BucketsByLocalDay ▲ proves current-streak buckets by the
// user's LOCAL day, not UTC. Two entries whose UTC date is 2026-04-25
// (05:00Z and 16:00Z) fall on DIFFERENT local days in America/Los_Angeles
// (04-24 22:00 and 04-25 09:00). With now on the local evening of 04-25,
// the local streak is 2. The OLD code buckets both entries under UTC
// 04-25 (one day) while seeding its cursor from now's LOCAL day (04-25,
// since it reads now.Day() directly), so it returns (1,1) — this test
// fails on old (1,1 != 2,2), passes on new. Offsets verified against
// tzdata; old/new values confirmed against a reference impl at design.
func TestStreak_BucketsByLocalDay(t *testing.T) {
	la, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		t.Fatalf("load America/Los_Angeles: %v", err)
	}
	entries := []storage.Entry{
		entryAt(2026, 4, 25, 5),  // UTC 05:00Z == 2026-04-24 22:00 PDT
		entryAt(2026, 4, 25, 16), // UTC 16:00Z == 2026-04-25 09:00 PDT
	}
	now := time.Date(2026, 4, 25, 20, 0, 0, 0, la) // 2026-04-25 evening local
	current, longest := Streak(entries, now)
	if current != 2 || longest != 2 {
		t.Errorf("BucketsByLocalDay: got (%d,%d), want (2,2)", current, longest)
	}
}

// TestStreak_CurrentStepsAcrossDSTBoundary ● GUARD (not fail-first):
// locks that the NEW code's LOCAL cursor steps by CALENDAR day (AddDate),
// not by 24h, across the 2026 US spring-forward (Mar 8, a 23-hour local
// day). A run on local days Mar 7/8/9 with now on Mar 9 must count 3.
// This fixture's UTC and local dates coincide, so the OLD code also
// returns (3,3) — the test does NOT fail on current code (confirmed
// against a reference impl). Its job is to pair with the calendar-
// arithmetic decision: if the NEW code stepped with cursor.Add(-24h)
// instead of AddDate, Mar 9 00:00 local minus 24h lands on Mar 7 23:00
// (date Mar 7), SKIPPING Mar 8 → current would be 2, and this test would
// fail. So it guards against a DST-unsafe reimplementation of the walk.
func TestStreak_CurrentStepsAcrossDSTBoundary(t *testing.T) {
	la, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		t.Fatalf("load America/Los_Angeles: %v", err)
	}
	entries := []storage.Entry{
		entryAt(2026, 3, 7, 20), // local noon PST (UTC-8) == 20:00Z, local 03-07
		entryAt(2026, 3, 8, 19), // local noon PDT (UTC-7) == 19:00Z, local 03-08
		entryAt(2026, 3, 9, 19), // local noon PDT == 19:00Z, local 03-09
	}
	now := time.Date(2026, 3, 9, 18, 0, 0, 0, la) // 2026-03-09 evening local
	current, longest := Streak(entries, now)
	if current != 3 || longest != 3 {
		t.Errorf("CurrentStepsAcrossDSTBoundary: got (%d,%d), want (3,3)", current, longest)
	}
}

// TestMostCommon_TopNCapAlphaTiebreakAndEmpty locks the top-N counter
// contract per SPEC-020 locked decision §3 (strict cap + alpha tiebreak;
// empty-string exclusion; non-nil empty on empty input).
func TestMostCommon_TopNCapAlphaTiebreakAndEmpty(t *testing.T) {
	t.Run("cap_at_n", func(t *testing.T) {
		input := []string{"a", "a", "a", "b", "b", "c", "c", "d", "e"}
		want := []NameCount{{Name: "a", Count: 3}, {Name: "b", Count: 2}, {Name: "c", Count: 2}}
		got := MostCommon(input, 3)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("cap_at_n: got %#v, want %#v", got, want)
		}
	})

	t.Run("boundary_tie_alpha_resolves", func(t *testing.T) {
		input := []string{"zebra", "yak", "x-ray", "wolf", "vulture", "umbrella"}
		want := []NameCount{
			{Name: "umbrella", Count: 1},
			{Name: "vulture", Count: 1},
			{Name: "wolf", Count: 1},
			{Name: "x-ray", Count: 1},
			{Name: "yak", Count: 1},
		}
		got := MostCommon(input, 5)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("boundary_tie_alpha_resolves: got %#v, want %#v", got, want)
		}
	})

	t.Run("fewer_than_n", func(t *testing.T) {
		input := []string{"a", "a", "b"}
		want := []NameCount{{Name: "a", Count: 2}, {Name: "b", Count: 1}}
		got := MostCommon(input, 5)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("fewer_than_n: got %#v, want %#v", got, want)
		}
	})

	t.Run("empty_strings_excluded", func(t *testing.T) {
		input := []string{"a", "", "a", "", "b"}
		want := []NameCount{{Name: "a", Count: 2}, {Name: "b", Count: 1}}
		got := MostCommon(input, 5)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("empty_strings_excluded: got %#v, want %#v", got, want)
		}
	})

	t.Run("empty_input_nonnil_slice", func(t *testing.T) {
		got := MostCommon(nil, 5)
		if got == nil {
			t.Errorf("empty_input_nonnil_slice (nil): expected non-nil slice")
		}
		if len(got) != 0 {
			t.Errorf("empty_input_nonnil_slice (nil): expected len 0, got %d", len(got))
		}
		got = MostCommon([]string{}, 5)
		if got == nil {
			t.Errorf("empty_input_nonnil_slice (empty): expected non-nil slice")
		}
		if len(got) != 0 {
			t.Errorf("empty_input_nonnil_slice (empty): expected len 0, got %d", len(got))
		}
	})
}

// TestSpan_FirstLastAndInclusiveDays locks the SPEC-020 corpus span
// contract per locked decision §7 (inclusive on both ends; same-day
// multiple entries → 1; empty → zero-value).
func TestSpan_FirstLastAndInclusiveDays(t *testing.T) {
	t.Run("multi_day", func(t *testing.T) {
		entries := []storage.Entry{
			entryAt(2026, 4, 12, 10),
			entryAt(2026, 4, 18, 10),
			entryAt(2026, 4, 25, 10),
		}
		got := Span(entries)
		wantFirst := time.Date(2026, 4, 12, 10, 0, 0, 0, time.UTC)
		wantLast := time.Date(2026, 4, 25, 10, 0, 0, 0, time.UTC)
		if !got.First.Equal(wantFirst) {
			t.Errorf("multi_day First: got %v, want %v", got.First, wantFirst)
		}
		if !got.Last.Equal(wantLast) {
			t.Errorf("multi_day Last: got %v, want %v", got.Last, wantLast)
		}
		if got.Days != 14 {
			t.Errorf("multi_day Days: got %d, want 14", got.Days)
		}
	})

	t.Run("single_day", func(t *testing.T) {
		entries := []storage.Entry{entryAt(2026, 4, 25, 10)}
		got := Span(entries)
		if got.Days != 1 {
			t.Errorf("single_day Days: got %d, want 1", got.Days)
		}
	})

	t.Run("same_day_multiple_entries", func(t *testing.T) {
		entries := []storage.Entry{
			entryAt(2026, 4, 25, 8),
			entryAt(2026, 4, 25, 12),
			entryAt(2026, 4, 25, 16),
		}
		got := Span(entries)
		if got.Days != 1 {
			t.Errorf("same_day_multiple_entries Days: got %d, want 1", got.Days)
		}
	})

	t.Run("empty_corpus", func(t *testing.T) {
		got := Span(nil)
		if !got.First.IsZero() {
			t.Errorf("empty_corpus First: expected zero, got %v", got.First)
		}
		if !got.Last.IsZero() {
			t.Errorf("empty_corpus Last: expected zero, got %v", got.Last)
		}
		if got.Days != 0 {
			t.Errorf("empty_corpus Days: got %d, want 0", got.Days)
		}
	})
}

// TestStatsAggregate_EmptyInputContract consolidates the empty-input
// contract for the three SPEC-020 helpers per locked decision §11.
func TestStatsAggregate_EmptyInputContract(t *testing.T) {
	now := time.Date(2026, 4, 25, 12, 0, 0, 0, time.UTC)

	current, longest := Streak(nil, now)
	if current != 0 || longest != 0 {
		t.Errorf("Streak(nil): got (%d,%d), want (0,0)", current, longest)
	}

	mc := MostCommon(nil, 5)
	if mc == nil {
		t.Errorf("MostCommon(nil): expected non-nil empty slice, got nil")
	}
	if len(mc) != 0 {
		t.Errorf("MostCommon(nil): expected len 0, got %d", len(mc))
	}

	sp := Span(nil)
	if !sp.First.IsZero() || !sp.Last.IsZero() || sp.Days != 0 {
		t.Errorf("Span(nil): got %+v, want zero-value", sp)
	}
}

// TestWithImpact_FiltersNonEmptyImpactPreservingOrder pairs SPEC-048
// locked decision 4 (impact-first): WithImpact returns exactly the
// entries with non-empty Impact, in input order, dropping the empties.
func TestWithImpact_FiltersNonEmptyImpactPreservingOrder(t *testing.T) {
	in := []storage.Entry{
		{ID: 1, Title: "a", Project: "alpha", Impact: "cut latency"},
		{ID: 2, Title: "b", Project: "beta", Impact: "onboarding 1 day"},
		{ID: 3, Title: "c", Project: "gamma", Impact: ""},
		{ID: 4, Title: "d", Project: "alpha", Impact: "removed cron"},
		{ID: 5, Title: "e", Impact: ""},
	}
	got := WithImpact(in)
	wantIDs := []int64{1, 2, 4}
	if len(got) != len(wantIDs) {
		t.Fatalf("WithImpact: got %d entries, want %d (%v)", len(got), len(wantIDs), wantIDs)
	}
	for i, e := range got {
		if e.ID != wantIDs[i] {
			t.Errorf("WithImpact[%d].ID: got %d, want %d (order must be preserved)", i, e.ID, wantIDs[i])
		}
		if e.Impact == "" {
			t.Errorf("WithImpact[%d]: empty-impact entry leaked into result", i)
		}
	}
}

// TestWithImpact_EmptyInputAndAllEmptyImpact pairs SPEC-048 locked
// decision 4: both an empty input and an all-empty-impact input return
// a non-nil empty slice so JSON callers never see null.
func TestWithImpact_EmptyInputAndAllEmptyImpact(t *testing.T) {
	got := WithImpact(nil)
	if got == nil {
		t.Errorf("WithImpact(nil): expected non-nil empty slice, got nil")
	}
	if len(got) != 0 {
		t.Errorf("WithImpact(nil): expected len 0, got %d", len(got))
	}

	allEmpty := []storage.Entry{
		{ID: 1, Title: "a", Impact: ""},
		{ID: 2, Title: "b", Impact: ""},
	}
	got2 := WithImpact(allEmpty)
	if got2 == nil {
		t.Errorf("WithImpact(all-empty): expected non-nil empty slice, got nil")
	}
	if len(got2) != 0 {
		t.Errorf("WithImpact(all-empty): expected len 0, got %d", len(got2))
	}
}

// --- SPEC-045: provenance coverage helpers ------------------------------

// coverageAggFixture mirrors the export package's coverageYearFixture (kept
// local because the two live in different packages). 4 agent / 6 human;
// self-reference ids 1,4,9 (3). See SPEC-045 Failing Tests.
var coverageAggFixture = []storage.Entry{
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

var coverageAggMonths = []string{
	"2026-01", "2026-02", "2026-03", "2026-04", "2026-05", "2026-06",
	"2026-07", "2026-08", "2026-09", "2026-10", "2026-11", "2026-12"}

// Test 6 — TestIsAgentAuthored_ClassifiesReservedNamespace: the Go predicate
// matches the SQL LIKE 'agent:%'/'model:%' anchoring (SPEC-045 LD2).
func TestIsAgentAuthored_ClassifiesReservedNamespace(t *testing.T) {
	cases := []struct {
		tags string
		want bool
	}{
		{"agent:x", true},
		{"model:x", true},
		{"a,agent:x,b", true},       // mid-list
		{"agent:x,model:y", true},   // both
		{"agentic,modeling", false}, // no colon → false-positive guard
		{"", false},                 // no tags
		{"auth,api", false},         // plain human
		{"agent:anything", true},    // prefix-anchored regardless of suffix
	}
	for _, c := range cases {
		got := IsAgentAuthored(storage.Entry{Tags: c.tags})
		if got != c.want {
			t.Errorf("IsAgentAuthored(Tags=%q) = %v, want %v", c.tags, got, c.want)
		}
	}
}

// Test 7 — TestCoverageByMonth_BucketsAndShareZeroFilled: 12 zero-filled
// buckets in month order; counts + 4-decimal share per bucket (SPEC-045 LD7).
func TestCoverageByMonth_BucketsAndShareZeroFilled(t *testing.T) {
	got := CoverageByMonth(coverageAggFixture, coverageAggMonths)
	want := []CoverageBucket{
		{Period: "2026-01", Agent: 0, Human: 0, Share: 0},
		{Period: "2026-02", Agent: 0, Human: 1, Share: 0},
		{Period: "2026-03", Agent: 0, Human: 1, Share: 0},
		{Period: "2026-04", Agent: 0, Human: 0, Share: 0},
		{Period: "2026-05", Agent: 0, Human: 1, Share: 0},
		{Period: "2026-06", Agent: 0, Human: 0, Share: 0},
		{Period: "2026-07", Agent: 1, Human: 1, Share: 0.5},
		{Period: "2026-08", Agent: 0, Human: 0, Share: 0},
		{Period: "2026-09", Agent: 1, Human: 0, Share: 1},
		{Period: "2026-10", Agent: 0, Human: 0, Share: 0},
		{Period: "2026-11", Agent: 1, Human: 1, Share: 0.5},
		{Period: "2026-12", Agent: 1, Human: 1, Share: 0.5},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("CoverageByMonth mismatch:\n got=%+v\nwant=%+v", got, want)
	}

	// Empty input over the same months → 12 {0,0,0} buckets.
	empty := CoverageByMonth(nil, coverageAggMonths)
	if len(empty) != 12 {
		t.Fatalf("empty input: got %d buckets, want 12", len(empty))
	}
	for _, b := range empty {
		if b.Agent != 0 || b.Human != 0 || b.Share != 0 {
			t.Errorf("empty input: bucket %q not zero: %+v", b.Period, b)
		}
	}
}

// Test 8 — TestSelfReferenceCount_SubstringCaseInsensitive (SPEC-045 LD5).
func TestSelfReferenceCount_SubstringCaseInsensitive(t *testing.T) {
	if got := SelfReferenceCount(coverageAggFixture); got != 3 {
		t.Errorf("SelfReferenceCount(fixture) = %d, want 3 (ids 1,4,9)", got)
	}

	cases := []struct {
		name    string
		entries []storage.Entry
		want    int
	}{
		{"title-only", []storage.Entry{{Title: "brag digest", Description: "x"}}, 1},
		{"description-only", []storage.Entry{{Title: "x", Description: "the brag tool"}}, 1},
		{"mixed-case", []storage.Entry{{Title: "BragFile release", Description: "y"}}, 1},
		{"non-match", []storage.Entry{{Title: "auth refactor", Description: "login"}}, 0},
		{"empty", nil, 0},
	}
	for _, c := range cases {
		if got := SelfReferenceCount(c.entries); got != c.want {
			t.Errorf("%s: SelfReferenceCount = %d, want %d", c.name, got, c.want)
		}
	}
}
