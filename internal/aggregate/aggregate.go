// Package aggregate computes structured statistics over
// []storage.Entry. It is the data layer for the rule-based commands
// brag summary (SPEC-018), brag review (SPEC-019), and brag stats
// (SPEC-020). Rendering lives in internal/export.
package aggregate

import (
	"sort"

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
