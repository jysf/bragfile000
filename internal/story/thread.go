package story

import (
	"sort"
	"strings"
	"time"

	"github.com/jysf/bragfile000/internal/aggregate"
	"github.com/jysf/bragfile000/internal/storage"
)

// Kind values for a Thread.
const (
	KindInitiative = "initiative"
	KindTheme      = "theme"
)

// Beat is one entry projected into a thread. IsImpactBeat mirrors
// aggregate.WithImpact's rule exactly: an entry is an impact beat iff its
// Impact field is non-empty.
type Beat struct {
	ID           int64
	Title        string
	Project      string
	Type         string
	Impact       string
	IsImpactBeat bool
	CreatedAt    time.Time
}

// Thread is a coalesced narrative unit: an initiative (the project axis)
// or a --theme cross-cut. Beats are time-ordered ASC with an ID tiebreak.
type Thread struct {
	Thread string // project name, NoProjectKey, or the theme tag
	Kind   string // KindInitiative | KindTheme
	Beats  []Beat
}

// ThreadOptions is the profile's threading/altitude policy, plus the
// optional theme cross-cut.
type ThreadOptions struct {
	Order               string // "initiative" | "impact-desc"
	ImpactThreadsOnly   bool
	DropImpactlessBeats bool
	FoldSmallThreads    bool
	Theme               string // "" or the --theme tag
}

// ThreadOptionsFromProfile lifts the threading policy off a Profile,
// attaching the (optional) theme cross-cut tag.
func ThreadOptionsFromProfile(p Profile, theme string) ThreadOptions {
	return ThreadOptions{
		Order:               p.ThreadOrder,
		ImpactThreadsOnly:   p.ImpactThreadsOnly,
		DropImpactlessBeats: p.DropImpactlessBeats,
		FoldSmallThreads:    p.FoldSmallThreads,
		Theme:               theme,
	}
}

func impactBeatCount(t Thread) int {
	n := 0
	for _, b := range t.Beats {
		if b.IsImpactBeat {
			n++
		}
	}
	return n
}

// BuildThreads coalesces the already-in-window entries into deterministic
// threads per the profile's policy (DEC-029 choice 1/3):
//  1. initiative threads via aggregate.GroupEntriesByProject (alpha-ASC,
//     (no project) last, beats ASC+ID-tiebreak already);
//  2. drop impact-less beats if DropImpactlessBeats;
//  3. fold threads with 0 impact beats if ImpactThreadsOnly/FoldSmallThreads;
//  4. reorder by impact-beat-count DESC (alpha-ASC tiebreak) if
//     Order == "impact-desc";
//  5. append ONE theme cross-cut thread (kind theme) after initiatives if
//     Theme != "" — never folded/dropped (it is an explicit opt-in).
func BuildThreads(entries []storage.Entry, opts ThreadOptions) []Thread {
	groups := aggregate.GroupEntriesByProject(entries)
	threads := make([]Thread, 0, len(groups))
	for _, g := range groups {
		beats := make([]Beat, 0, len(g.Entries))
		for _, e := range g.Entries {
			if opts.DropImpactlessBeats && e.Impact == "" {
				continue
			}
			beats = append(beats, entryToBeat(e, g.Project))
		}
		threads = append(threads, Thread{
			Thread: g.Project,
			Kind:   KindInitiative,
			Beats:  beats,
		})
	}

	if opts.ImpactThreadsOnly || opts.FoldSmallThreads {
		kept := threads[:0:0]
		for _, t := range threads {
			if impactBeatCount(t) == 0 {
				continue
			}
			kept = append(kept, t)
		}
		threads = kept
	}

	if opts.Order == "impact-desc" {
		// GroupEntriesByProject already gave alpha-ASC; a stable sort by
		// impact-beat-count DESC keeps that as the tiebreak.
		sort.SliceStable(threads, func(i, j int) bool {
			return impactBeatCount(threads[i]) > impactBeatCount(threads[j])
		})
	}

	if opts.Theme != "" {
		threads = append(threads, buildThemeThread(entries, opts.Theme))
	}

	return threads
}

// entryToBeat projects a storage.Entry into a Beat. project is passed in
// so (no project) entries carry aggregate.NoProjectKey (matching the
// group key), not the empty string.
func entryToBeat(e storage.Entry, project string) Beat {
	return Beat{
		ID:           e.ID,
		Title:        e.Title,
		Project:      project,
		Type:         e.Type,
		Impact:       e.Impact,
		IsImpactBeat: e.Impact != "",
		CreatedAt:    e.CreatedAt,
	}
}

// buildThemeThread groups every in-window entry carrying the theme tag
// into one cross-project thread, time-ordered ASC with an ID tiebreak.
// Tag membership is exact-token match over the comma-joined Tags field
// (matching ListFilter.Tag semantics, DEC-004/DEC-015).
func buildThemeThread(entries []storage.Entry, theme string) Thread {
	var matched []storage.Entry
	for _, e := range entries {
		if tagsContain(e.Tags, theme) {
			matched = append(matched, e)
		}
	}
	sort.SliceStable(matched, func(i, j int) bool {
		if !matched[i].CreatedAt.Equal(matched[j].CreatedAt) {
			return matched[i].CreatedAt.Before(matched[j].CreatedAt)
		}
		return matched[i].ID < matched[j].ID
	})
	beats := make([]Beat, 0, len(matched))
	for _, e := range matched {
		project := e.Project
		if project == "" {
			project = aggregate.NoProjectKey
		}
		beats = append(beats, entryToBeat(e, project))
	}
	return Thread{Thread: theme, Kind: KindTheme, Beats: beats}
}

func tagsContain(tags, token string) bool {
	for _, t := range strings.Split(tags, ",") {
		if strings.TrimSpace(t) == token {
			return true
		}
	}
	return false
}

// Arc is one throughline arc: a thread's beat/impact tally + span. The
// throughline is a deterministic SKELETON (DEC-029 choice 1) — the LLM,
// via the framing directive, weaves the actual narrative arc.
type Arc struct {
	Thread          string
	Kind            string
	BeatCount       int
	ImpactBeatCount int
	Span            Span
}

// Span is a thread's lifetime: first/last beat CreatedAt.
type Span struct {
	First time.Time
	Last  time.Time
}

// Throughline is the ordered skeleton of arcs (one per thread).
type Throughline struct {
	Arcs []Arc
}

// BuildThroughline produces one arc per thread carrying its beat count,
// impact-beat count, and span (first/last beat CreatedAt). A thread with
// no beats yields a zero-value span. Empty input → non-nil empty Arcs.
func BuildThroughline(threads []Thread) Throughline {
	arcs := make([]Arc, 0, len(threads))
	for _, t := range threads {
		arcs = append(arcs, Arc{
			Thread:          t.Thread,
			Kind:            t.Kind,
			BeatCount:       len(t.Beats),
			ImpactBeatCount: impactBeatCount(t),
			Span:            threadSpan(t),
		})
	}
	return Throughline{Arcs: arcs}
}

func threadSpan(t Thread) Span {
	if len(t.Beats) == 0 {
		return Span{}
	}
	first := t.Beats[0].CreatedAt
	last := first
	for _, b := range t.Beats[1:] {
		if b.CreatedAt.Before(first) {
			first = b.CreatedAt
		}
		if b.CreatedAt.After(last) {
			last = b.CreatedAt
		}
	}
	return Span{First: first, Last: last}
}
