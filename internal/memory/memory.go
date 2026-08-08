// Package memory ranks a candidate pool of entries by DEC-043's blended
// reciprocal-rank fusion and trims the result to DEC-044's token budget —
// the corpus read back as working memory. The package is deliberately PURE:
// it reads no clock and no host state (the recency signal is ordinal, a
// rank rather than an age, so the ranking is time-invariant — DEC-043
// sub-decision 3), and it imports nothing outside the standard library plus
// internal/storage (for the Entry type).
//
// The token estimate this package computes sizes a single retrieval; it is
// never written to storage and never enters the caller-reported token-count
// tag DEC-027 reserves — see DEC-044's honesty clause.
package memory

import (
	"fmt"
	"sort"

	"github.com/jysf/bragfile000/internal/storage"
)

// Locked constants (DEC-043, DEC-044). Each is exported and golden-pinned:
// changing any one is a visible behavior change, not a silent one.
const (
	FusionK       = 60
	WeightRecency = 1.0
	WeightMatch   = 1.0
	WeightProject = 1.0
	PoolLimit     = 200
	DefaultBudget = 2000
	CharsPerToken = 4
)

// Options selects and bounds a memory slice. There is deliberately NO `now`
// field: the recency signal is ordinal (a rank, not an age), so the ranking
// is time-invariant (DEC-043 sub-decision 3). The renderer owns the wall
// clock, as in every other DEC-014 consumer.
type Options struct {
	// Query is echoed into the envelope only; ranking uses Matched.
	Query string
	// Project, when non-empty, contributes a soft boost term — never a filter
	// (DEC-043 sub-decision 4).
	Project string
	// Matched is the entry ids Store.Search returned, IN ITS RANK ORDER
	// (FTS5 `ORDER BY rank, id DESC`). Nil/empty means "no query". Ids not in
	// the candidate pool, and duplicates, are ignored. CALLER CONTRACT: the
	// order is the signal — passing an arbitrary id set silently degrades the
	// relevance term to whatever order you happened to build.
	Matched []int64
	// Budget is the token budget for the rendered entry lines. Must be > 0;
	// the caller rejects 0/negative before calling (DEC-044 sub-decision 1).
	Budget int
}

// Item is one selected entry plus everything the renderers need. Line is the
// DEC-044 rendering and Tokens is its estimated cost, so the estimate and the
// render cannot disagree.
type Item struct {
	Entry  storage.Entry
	Line   string
	Tokens int
	Score  float64
	Rank   int // 1-based position within the full fused ranking
}

// Result is the slice plus its accounting. Included+Skipped == Candidates.
type Result struct {
	Items           []Item // non-nil, in rank order
	Candidates      int    // deduped pool size
	Included        int
	Skipped         int
	EstimatedTokens int
	Budget          int
}

// Slice ranks entries by DEC-043's reciprocal-rank fusion over three ordinal
// lists — recency (re-derived from entries, not from input order), relevance
// (opts.Matched, in its given order), and project membership (opts.Project,
// recency-ordered) — then trims the ranked list to opts.Budget by DEC-044's
// skip-and-continue policy: an entry that does not fit is skipped and the
// fill continues, so a later cheaper entry can still be included.
//
// Slice is a pure function of its arguments. It dedupes the input by ID and
// is invariant to input order.
func Slice(entries []storage.Entry, opts Options) Result {
	pool := dedupeByID(entries)
	sort.SliceStable(pool, func(i, j int) bool {
		return recencyLess(pool[i], pool[j])
	})

	matchRank := buildMatchRank(pool, opts.Matched)
	projectRank := buildProjectRank(pool, opts.Project)

	type scored struct {
		entry storage.Entry
		score float64
	}
	scoredPool := make([]scored, len(pool))
	for i, e := range pool {
		recRank := i + 1
		s := WeightRecency / float64(FusionK+recRank)
		if r, ok := matchRank[e.ID]; ok {
			s += WeightMatch / float64(FusionK+r)
		}
		if r, ok := projectRank[e.ID]; ok {
			s += WeightProject / float64(FusionK+r)
		}
		scoredPool[i] = scored{entry: e, score: s}
	}

	// Stable sort over the already recency-ordered pool: equal scores fall
	// back to created_at DESC, id DESC with no separate comparator
	// (DEC-043 sub-decision 6).
	sort.SliceStable(scoredPool, func(i, j int) bool {
		return scoredPool[i].score > scoredPool[j].score
	})

	ranked := make([]Item, len(scoredPool))
	for i, sc := range scoredPool {
		line := RenderLine(sc.entry)
		ranked[i] = Item{
			Entry:  sc.entry,
			Line:   line,
			Tokens: EstimateTokens(line),
			Score:  sc.score,
			Rank:   i + 1,
		}
	}

	included := make([]Item, 0, len(ranked))
	remaining := opts.Budget
	estimated := 0
	skipped := 0
	for _, it := range ranked {
		if it.Tokens <= remaining {
			included = append(included, it)
			remaining -= it.Tokens
			estimated += it.Tokens
		} else {
			skipped++
		}
	}

	return Result{
		Items:           included,
		Candidates:      len(pool),
		Included:        len(included),
		Skipped:         skipped,
		EstimatedTokens: estimated,
		Budget:          opts.Budget,
	}
}

// dedupeByID returns entries with duplicate IDs collapsed to their first
// occurrence, preserving input order. The result feeds a full re-sort, so
// which duplicate copy is kept is immaterial as long as duplicates carry
// identical data for a given ID.
func dedupeByID(entries []storage.Entry) []storage.Entry {
	seen := make(map[int64]bool, len(entries))
	out := make([]storage.Entry, 0, len(entries))
	for _, e := range entries {
		if seen[e.ID] {
			continue
		}
		seen[e.ID] = true
		out = append(out, e)
	}
	return out
}

// recencyLess reports whether a sorts before b under created_at DESC, id
// DESC (Store.List's documented order), re-derived from the entries
// themselves so Slice depends on no input-order invariant.
func recencyLess(a, b storage.Entry) bool {
	if !a.CreatedAt.Equal(b.CreatedAt) {
		return a.CreatedAt.After(b.CreatedAt)
	}
	return a.ID > b.ID
}

// buildMatchRank assigns each id in matched its 1-based position among the
// FIRST occurrences of ids that are actually present in pool. Ids not in
// pool, and duplicate ids, are ignored (DEC-043 sub-decision 1).
func buildMatchRank(pool []storage.Entry, matched []int64) map[int64]int {
	inPool := make(map[int64]bool, len(pool))
	for _, e := range pool {
		inPool[e.ID] = true
	}
	ranks := make(map[int64]int)
	rank := 0
	for _, id := range matched {
		if !inPool[id] {
			continue
		}
		if _, dup := ranks[id]; dup {
			continue
		}
		rank++
		ranks[id] = rank
	}
	return ranks
}

// buildProjectRank assigns a 1-based rank, in pool order (already recency-
// ordered), to every entry whose Project equals project. Empty when project
// is empty or matches no entry (DEC-043 sub-decision 1 + 4).
func buildProjectRank(pool []storage.Entry, project string) map[int64]int {
	ranks := make(map[int64]int)
	if project == "" {
		return ranks
	}
	rank := 0
	for _, e := range pool {
		if e.Project != project {
			continue
		}
		rank++
		ranks[e.ID] = rank
	}
	return ranks
}

// RenderLine renders e as DEC-044 sub-decision 5's one-line entry shape:
//
//   - <id> <YYYY-MM-DD> [<project>/<type>] <title> — <impact>
//
// with "-" for an absent project or type, and the impact clause only when
// Impact is non-empty (no trailing em-dash artifact). The date is
// created_at's UTC date part only.
func RenderLine(e storage.Entry) string {
	project := e.Project
	if project == "" {
		project = "-"
	}
	typ := e.Type
	if typ == "" {
		typ = "-"
	}
	date := e.CreatedAt.UTC().Format("2006-01-02")
	line := fmt.Sprintf("- %d %s [%s/%s] %s", e.ID, date, project, typ, e.Title)
	if e.Impact != "" {
		line += " — " + e.Impact
	}
	return line
}

// EstimateTokens returns ceil(len(s)/4) over UTF-8 BYTES (not runes),
// DEC-044 sub-decision 2's documented chars-per-token heuristic. Byte
// counting degrades in the safe direction: it over-estimates the cost of
// multi-byte text rather than under-estimating it.
func EstimateTokens(s string) int {
	n := len(s)
	if n == 0 {
		return 0
	}
	return (n + CharsPerToken - 1) / CharsPerToken
}
