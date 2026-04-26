// Package aggregate computes structured statistics over
// []storage.Entry. It is the data layer for the rule-based commands
// brag summary (SPEC-018), brag review (SPEC-019), and brag stats
// (SPEC-020). Rendering lives in internal/export.
package aggregate

import (
	"sort"
	"time"

	"github.com/jysf/bragfile000/internal/storage"
)

// NoProjectKey is the literal sentinel used in place of an empty-
// string Project. Locked by DEC-014; the markdown and JSON renderers
// both display this exact string.
const NoProjectKey = "(no project)"

// TypeCount is one row of ByType's result.
type TypeCount struct {
	Type  string
	Count int
}

// ProjectCount is one row of ByProject's result.
type ProjectCount struct {
	Project string
	Count   int
}

// EntryRef is the projection of a storage.Entry that highlights
// carries: ID + Title only. Description, tags, project, type,
// timestamps are intentionally elided per SPEC-018's "skim before
// pasting" goal.
type EntryRef struct {
	ID    int64
	Title string
}

// ProjectHighlights is one project's group of entries.
type ProjectHighlights struct {
	Project string
	Entries []EntryRef
}

// ByType returns entries grouped by Type, ordered DESC by count with
// alphabetical-ASC tiebreak. Empty input returns a non-nil empty
// slice (so JSON renders [] not null).
func ByType(entries []storage.Entry) []TypeCount {
	m := make(map[string]int)
	for _, e := range entries {
		m[e.Type]++
	}
	out := make([]TypeCount, 0, len(m))
	for t, c := range m {
		out = append(out, TypeCount{Type: t, Count: c})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].Type < out[j].Type
	})
	return out
}

// ByProject is identical in shape to ByType, except entries with
// empty-string Project are rendered under NoProjectKey and forced
// LAST regardless of count (matches DEC-013's (no project)-last
// convention; locked by DEC-014).
func ByProject(entries []storage.Entry) []ProjectCount {
	m := make(map[string]int)
	for _, e := range entries {
		key := e.Project
		if key == "" {
			key = NoProjectKey
		}
		m[key]++
	}
	out := make([]ProjectCount, 0, len(m))
	for p, c := range m {
		out = append(out, ProjectCount{Project: p, Count: c})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Project == NoProjectKey {
			return false
		}
		if out[j].Project == NoProjectKey {
			return true
		}
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].Project < out[j].Project
	})
	return out
}

// ProjectEntryGroup carries one project's full entries, used by brag
// review (SPEC-019). Mirrors ProjectHighlights's shape but retains the
// full storage.Entry instead of the EntryRef projection — JSON
// consumers (DEC-011 9-key per-entry shape) need descriptions and
// metadata that highlights elides.
type ProjectEntryGroup struct {
	Project string
	Entries []storage.Entry
}

// GroupEntriesByProject mirrors GroupForHighlights's grouping + sort
// logic exactly: alpha-ASC by project name with NoProjectKey last;
// chrono-ASC within group with ID as tiebreak. Differs only in
// carrying full storage.Entry (not EntryRef). Used by review's
// markdown path (renders id+title only at render time) and review's
// JSON path (serializes full DEC-011 shape).
func GroupEntriesByProject(entries []storage.Entry) []ProjectEntryGroup {
	buckets := make(map[string][]storage.Entry)
	for _, e := range entries {
		key := e.Project
		if key == "" {
			key = NoProjectKey
		}
		buckets[key] = append(buckets[key], e)
	}
	out := make([]ProjectEntryGroup, 0, len(buckets))
	for proj, group := range buckets {
		sorted := make([]storage.Entry, len(group))
		copy(sorted, group)
		sort.SliceStable(sorted, func(i, j int) bool {
			if !sorted[i].CreatedAt.Equal(sorted[j].CreatedAt) {
				return sorted[i].CreatedAt.Before(sorted[j].CreatedAt)
			}
			return sorted[i].ID < sorted[j].ID
		})
		out = append(out, ProjectEntryGroup{Project: proj, Entries: sorted})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Project == NoProjectKey {
			return false
		}
		if out[j].Project == NoProjectKey {
			return true
		}
		return out[i].Project < out[j].Project
	})
	return out
}

// NameCount is a generic top-N count: Name is the value (a tag string
// or a project string), Count is the occurrence count. Renderer-side
// callers wrap []NameCount into the per-spec semantic JSON shapes
// ({tag,count} / {project,count}). SPEC-020.
type NameCount struct {
	Name  string
	Count int
}

// CorpusSpan describes the lifetime span of a corpus: First / Last
// CreatedAt (UTC) and Days inclusive on both endpoints. Empty corpus
// → zero-value (all three fields zero). SPEC-020.
type CorpusSpan struct {
	First time.Time
	Last  time.Time
	Days  int
}

// Streak returns (current, longest) streak counts in UTC calendar days.
// current counts back from now's UTC date while each preceding date has
// >=1 entry; if now's UTC date has zero entries, current = 0 (NOT "the
// streak that ended yesterday"). longest is the longest consecutive
// UTC-date run anywhere in the corpus. Multiple entries on the same UTC
// date count as one streak day. Empty corpus → (0,0). SPEC-020 §6.
func Streak(entries []storage.Entry, now time.Time) (current, longest int) {
	if len(entries) == 0 {
		return 0, 0
	}
	dates := make(map[string]struct{}, len(entries))
	for _, e := range entries {
		dates[e.CreatedAt.UTC().Format("2006-01-02")] = struct{}{}
	}

	cursor := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	for {
		if _, ok := dates[cursor.Format("2006-01-02")]; !ok {
			break
		}
		current++
		cursor = cursor.AddDate(0, 0, -1)
	}

	keys := make([]string, 0, len(dates))
	for k := range dates {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	run := 1
	longest = 1
	for i := 1; i < len(keys); i++ {
		prev, _ := time.Parse("2006-01-02", keys[i-1])
		curr, _ := time.Parse("2006-01-02", keys[i])
		if curr.Sub(prev) == 24*time.Hour {
			run++
		} else {
			run = 1
		}
		if run > longest {
			longest = run
		}
	}
	return current, longest
}

// MostCommon returns up to n NameCount entries from values, ordered DESC
// by count with alpha-ASC tiebreak. Empty-string values are excluded
// from counting. Strict cap at n: when 6+ values tie at the boundary,
// alpha-ASC determines which n. Fewer than n distinct values returns
// however many exist (no padding). Empty input → non-nil empty slice.
// SPEC-020 §3.
func MostCommon(values []string, n int) []NameCount {
	counts := make(map[string]int)
	for _, v := range values {
		if v == "" {
			continue
		}
		counts[v]++
	}
	out := make([]NameCount, 0, len(counts))
	for name, count := range counts {
		out = append(out, NameCount{Name: name, Count: count})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].Name < out[j].Name
	})
	if n >= 0 && len(out) > n {
		out = out[:n]
	}
	return out
}

// Span returns the CorpusSpan for entries: earliest CreatedAt as First,
// latest as Last (both UTC), and Days inclusive on both endpoints
// computed from UTC-truncated calendar dates. Empty corpus → zero-value
// struct. SPEC-020 §7.
func Span(entries []storage.Entry) CorpusSpan {
	if len(entries) == 0 {
		return CorpusSpan{}
	}
	first := entries[0].CreatedAt.UTC()
	last := first
	for _, e := range entries[1:] {
		t := e.CreatedAt.UTC()
		if t.Before(first) {
			first = t
		}
		if t.After(last) {
			last = t
		}
	}
	firstDay := time.Date(first.Year(), first.Month(), first.Day(), 0, 0, 0, 0, time.UTC)
	lastDay := time.Date(last.Year(), last.Month(), last.Day(), 0, 0, 0, 0, time.UTC)
	days := int(lastDay.Sub(firstDay).Hours()/24) + 1
	return CorpusSpan{First: first, Last: last, Days: days}
}

// GroupForHighlights returns project groups in alpha-ASC order with
// NoProjectKey forced last; within each group, entries are sorted
// ASC by CreatedAt with ID as tie-break (AGENTS.md §9 SPEC-002
// monotonic-tiebreak rule).
func GroupForHighlights(entries []storage.Entry) []ProjectHighlights {
	buckets := make(map[string][]storage.Entry)
	for _, e := range entries {
		key := e.Project
		if key == "" {
			key = NoProjectKey
		}
		buckets[key] = append(buckets[key], e)
	}
	out := make([]ProjectHighlights, 0, len(buckets))
	for proj, group := range buckets {
		sorted := make([]storage.Entry, len(group))
		copy(sorted, group)
		sort.SliceStable(sorted, func(i, j int) bool {
			if !sorted[i].CreatedAt.Equal(sorted[j].CreatedAt) {
				return sorted[i].CreatedAt.Before(sorted[j].CreatedAt)
			}
			return sorted[i].ID < sorted[j].ID
		})
		refs := make([]EntryRef, 0, len(sorted))
		for _, e := range sorted {
			refs = append(refs, EntryRef{ID: e.ID, Title: e.Title})
		}
		out = append(out, ProjectHighlights{Project: proj, Entries: refs})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Project == NoProjectKey {
			return false
		}
		if out[j].Project == NoProjectKey {
			return true
		}
		return out[i].Project < out[j].Project
	})
	return out
}
