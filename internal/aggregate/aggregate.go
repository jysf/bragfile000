// Package aggregate computes structured statistics over
// []storage.Entry. It is the data layer for the rule-based commands
// brag summary (SPEC-018), brag review (SPEC-019), and brag stats
// (SPEC-020). Rendering lives in internal/export.
package aggregate

import (
	"math"
	"sort"
	"strings"
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

// Streak returns (current, longest) streak counts in the user's LOCAL
// calendar day, where "local" is the zone carried by the injected now
// (now.Location()) — DEC-022. Entries are bucketed by converting each
// stored UTC instant into now.Location() before taking its date; storage
// itself stays UTC RFC3339 (only this derived metric localizes).
//
// current is the length of the consecutive local-day run that ends on
// TODAY or YESTERDAY: the streak stays alive through yesterday and is 0
// only once BOTH today and yesterday are empty (one day of grace, not
// immortality). longest is the longest consecutive local-day run anywhere
// in the corpus. Multiple entries on the same local date count as one
// streak day. Empty corpus → (0,0).
//
// All day arithmetic uses calendar operations (AddDate + date-label
// compare), never instant subtraction, so it is correct across DST
// transitions. DEC-022; supersedes the UTC-day/requires-today semantics
// SPEC-020 §6 locked.
func Streak(entries []storage.Entry, now time.Time) (current, longest int) {
	if len(entries) == 0 {
		return 0, 0
	}
	loc := now.Location()
	dates := make(map[string]struct{}, len(entries))
	for _, e := range entries {
		dates[e.CreatedAt.In(loc).Format("2006-01-02")] = struct{}{}
	}

	// Seed the cursor at today's local date; if today is empty, step back
	// one calendar day (the alive-through-yesterday grace) before walking.
	cursor := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	if _, ok := dates[cursor.Format("2006-01-02")]; !ok {
		cursor = cursor.AddDate(0, 0, -1)
	}
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
		// Calendar adjacency: is keys[i] the day after keys[i-1]? Compare
		// date labels via AddDate, never Sub == 24h (DST-immune).
		if prev.AddDate(0, 0, 1).Format("2006-01-02") == keys[i] {
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

// WithImpact returns the subset of entries whose Impact field is
// non-empty, preserving input order. Used by brag impact (SPEC-048):
// the impact digest is impact-first — impact-less entries are counted
// in provenance but excluded from the grouped body. Empty input or an
// all-empty-impact input returns a non-nil empty slice (JSON callers
// never see null). Order is preserved deliberately: grouping
// (GroupEntriesByProject) does the sorting, not this filter.
func WithImpact(entries []storage.Entry) []storage.Entry {
	out := make([]storage.Entry, 0, len(entries))
	for _, e := range entries {
		if e.Impact != "" {
			out = append(out, e)
		}
	}
	return out
}

// CadenceBucket is one month's entry count in a cadence series: Period
// is the "YYYY-MM" label, Count the number of entries whose created_at
// falls in that month. SPEC-051. SPEC-052 renders series[].Count as a
// sparkline, so this shape must not change without a paired test.
// The json tags are the on-the-wire shape the wrapped envelope (and
// SPEC-052) consume: series[].period / series[].count. They live on the
// aggregate struct because the struct itself is the sparkline-ready slot
// (LD8) — the export renderer embeds it directly rather than reprojecting.
type CadenceBucket struct {
	Period string `json:"period"`
	Count  int    `json:"count"`
}

// Cadence buckets entries by UTC calendar month, then emits one
// CadenceBucket per label in months order (zero-filled for months with
// no entries), plus the busiest-month label. months is the ordered set
// of "YYYY-MM" labels in scope (12 for a year, 3 for a quarter); the CLI
// derives it from the period so the series is always fully present, even
// on an empty period (every bucket zero, busiest ""). This is the
// sparkline-ready data slot (SPEC-052 reads series[].Count); it lives in
// aggregate — SQL-free, pure — so SPEC-052 and any future stats cadence
// reuse it (DEC-030 choice 5, LD8).
//
// The busiest month is the first label (in months order) whose count is
// the maximum; ties break toward the earlier month. An all-zero series
// (empty period) returns "" for busiest so the caller renders null.
func Cadence(entries []storage.Entry, months []string) (series []CadenceBucket, busiest string) {
	counts := make(map[string]int, len(months))
	for _, e := range entries {
		counts[e.CreatedAt.UTC().Format("2006-01")]++
	}
	series = make([]CadenceBucket, 0, len(months))
	maxCount := 0
	for _, label := range months {
		c := counts[label]
		series = append(series, CadenceBucket{Period: label, Count: c})
		if c > maxCount {
			maxCount = c
			busiest = label
		}
	}
	return series, busiest
}

// IsAgentAuthored reports whether e carries a reserved provenance tag
// (agent:<name> or model:<id>, DEC-024) — the SINGLE Go-side definition of
// "agent-authored", kept in agreement with storage's provenanceExistsClause
// SQL predicate by TestProvenanceClassifier_GoPredicateMatchesSQLClause. It
// splits Entry.Tags (the comma-joined projection of the same taggings join
// the SQL clause queries) and prefix-matches each token, mirroring the
// LIKE 'agent:%' / 'model:%' anchoring: a topic tag like "agentic" or
// "modeling" (no colon) is NOT provenance. This is the classifier SPEC-043's
// --author filter reads in SQL; brag coverage reads it in Go so it can count
// BOTH classes from one query (SPEC-045 LD2/LD7).
func IsAgentAuthored(e storage.Entry) bool {
	for _, raw := range strings.Split(e.Tags, ",") {
		t := strings.TrimSpace(raw)
		if strings.HasPrefix(t, "agent:") || strings.HasPrefix(t, "model:") {
			return true
		}
	}
	return false
}

// CoverageBucket is one month's provenance split: Period is the "YYYY-MM"
// label, Agent/Human the classified counts, Share = Agent/(Agent+Human)
// rounded to 4 decimals (0 when the month is empty). SPEC-045. The json tags
// are the on-the-wire shape the coverage envelope embeds directly (LD8 of
// SPEC-051): by_month[].period / .agent / .human / .share.
type CoverageBucket struct {
	Period string  `json:"period"`
	Agent  int     `json:"agent"`
	Human  int     `json:"human"`
	Share  float64 `json:"share"`
}

// CoverageByMonth buckets entries by UTC calendar month, classifies each via
// IsAgentAuthored, and emits one CoverageBucket per label in months order
// (zero-filled). months is the ordered "YYYY-MM" set the CLI derives from the
// window (12 for a year, 3 for a quarter, N for --since) so the series is
// always fully present, even on an empty window. Mirrors Cadence (SPEC-051).
func CoverageByMonth(entries []storage.Entry, months []string) []CoverageBucket {
	agent := map[string]int{}
	human := map[string]int{}
	for _, e := range entries {
		lbl := e.CreatedAt.UTC().Format("2006-01")
		if IsAgentAuthored(e) {
			agent[lbl]++
		} else {
			human[lbl]++
		}
	}
	out := make([]CoverageBucket, 0, len(months))
	for _, lbl := range months {
		a, h := agent[lbl], human[lbl]
		out = append(out, CoverageBucket{Period: lbl, Agent: a, Human: h, Share: shareRound(a, a+h)})
	}
	return out
}

// SelfReferenceCount returns how many entries mention "brag" (case-insensitive)
// in Title or Description — a proxy for dogfooding density (the corpus talking
// about the tool itself). Substring match: "brag" subsumes "bragfile" (SPEC-045
// LD5).
func SelfReferenceCount(entries []storage.Entry) int {
	n := 0
	for _, e := range entries {
		hay := strings.ToLower(e.Title + " " + e.Description)
		if strings.Contains(hay, "brag") {
			n++
		}
	}
	return n
}

// shareRound returns num/den rounded to 4 decimals (half-away-from-zero via
// math.Round), or 0 when den == 0. Used for per-month and overall shares so
// the JSON number is stable and the goldens are byte-exact. There is ONE
// rounding definition (SPEC-045 Notes for the Implementer); Share is its
// exported alias for the export renderer's overall-share computation.
func shareRound(num, den int) float64 {
	if den == 0 {
		return 0
	}
	return math.Round(float64(num)/float64(den)*10000) / 10000
}

// Share is the exported alias of shareRound so the export renderer rounds the
// overall agent_share / self_reference.share identically to the per-month
// CoverageBucket.Share — one rounding definition, no drift (SPEC-045).
func Share(num, den int) float64 { return shareRound(num, den) }

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
