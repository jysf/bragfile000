package story

import (
	"testing"
	"time"

	"github.com/jysf/bragfile000/internal/storage"
)

// storyFixture: 6 entries across the 2026 calendar year. Two initiatives
// (alpha, beta), a gamma initiative, plus one (no project) entry. Impact
// beats and impact-less beats are mixed so the me-vs-exec divergence is
// exercised: alpha has 2 beats (1 with impact, 1 without); beta has 2
// beats (both with impact); gamma has 1 (impact); (no project) has 1 beat
// (no impact); a 6th entry (gamma's perf-sweep) carries the theme tag
// `perf` and impact, for the --theme cross-cut (alpha-early also carries
// perf). Non-monotonic id/time pairing inside alpha exercises the ASC +
// ID-tiebreak path.
var storyFixture = []storage.Entry{
	{ID: 1, Title: "alpha-early", Project: "alpha", Type: "shipped",
		Tags:      "perf",
		Impact:    "cut p95 login latency 40%",
		CreatedAt: time.Date(2026, 2, 1, 10, 0, 0, 0, time.UTC)},
	{ID: 2, Title: "alpha-messy", Project: "alpha", Type: "learned",
		Impact:    "", // impact-less: KEPT by me, DROPPED by exec
		CreatedAt: time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC)},
	{ID: 3, Title: "beta-one", Project: "beta", Type: "shipped",
		Impact:    "onboarding time down to 1 day",
		CreatedAt: time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC)},
	{ID: 4, Title: "beta-two", Project: "beta", Type: "shipped",
		Impact:    "removed the nightly cron entirely",
		CreatedAt: time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC)},
	{ID: 5, Title: "loose-note", Type: "fixed",
		Impact:    "", // (no project), impact-less
		CreatedAt: time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)},
	{ID: 6, Title: "perf-sweep", Project: "gamma", Type: "shipped",
		Tags:      "perf",
		Impact:    "shaved 200ms off cold start",
		CreatedAt: time.Date(2026, 6, 15, 10, 0, 0, 0, time.UTC)},
}

var storyFixedNow = time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)

// meThreadOpts / execThreadOpts mirror the shipped me/exec profiles'
// threading policy, so the mechanism tests don't depend on the loader.
var meThreadOpts = ThreadOptions{Order: "initiative"}
var execThreadOpts = ThreadOptions{
	Order:               "impact-desc",
	ImpactThreadsOnly:   true,
	DropImpactlessBeats: true,
	FoldSmallThreads:    true,
}

func meThreadOptsWithTheme(theme string) ThreadOptions {
	o := meThreadOpts
	o.Theme = theme
	return o
}

// Test 5 — deterministic order + impact marking (me policy).
func TestBuildThreads_DeterministicOrderAndImpactMarking(t *testing.T) {
	threads := BuildThreads(storyFixture, meThreadOpts)

	wantOrder := []string{"alpha", "beta", "gamma", "(no project)"}
	if len(threads) != len(wantOrder) {
		t.Fatalf("thread count: got %d, want %d (%v)", len(threads), len(wantOrder), threadNames(threads))
	}
	for i, want := range wantOrder {
		if threads[i].Thread != want {
			t.Errorf("thread[%d]: got %q, want %q", i, threads[i].Thread, want)
		}
		if threads[i].Kind != KindInitiative {
			t.Errorf("thread[%d] kind: got %q, want %q", i, threads[i].Kind, KindInitiative)
		}
	}

	// alpha's beats: [id1, id2] (ASC + ID tiebreak).
	alpha := threads[0]
	if len(alpha.Beats) != 2 || alpha.Beats[0].ID != 1 || alpha.Beats[1].ID != 2 {
		t.Errorf("alpha beats: got %v, want [1 2]", beatIDs(alpha))
	}

	// IsImpactBeat == (Impact != ""): ids 1,3,4,6 true; ids 2,5 false.
	wantImpact := map[int64]bool{1: true, 2: false, 3: true, 4: true, 5: false, 6: true}
	for _, thr := range threads {
		for _, b := range thr.Beats {
			if b.IsImpactBeat != wantImpact[b.ID] {
				t.Errorf("beat %d IsImpactBeat: got %v, want %v", b.ID, b.IsImpactBeat, wantImpact[b.ID])
			}
		}
	}
}

// Test 6 — exec policy folds impact-less threads + drops impact-less beats.
func TestBuildThreads_ExecPolicyFoldsAndDrops(t *testing.T) {
	threads := BuildThreads(storyFixture, execThreadOpts)

	wantOrder := []string{"beta", "alpha", "gamma"}
	if got := threadNames(threads); !equalStrings(got, wantOrder) {
		t.Fatalf("exec thread order: got %v, want %v", got, wantOrder)
	}

	// alpha keeps only its impact beat (id1); alpha-messy (id2) dropped.
	var alpha Thread
	for _, thr := range threads {
		if thr.Thread == "alpha" {
			alpha = thr
		}
	}
	if len(alpha.Beats) != 1 || alpha.Beats[0].ID != 1 {
		t.Errorf("exec alpha beats: got %v, want [1]", beatIDs(alpha))
	}
}

// Test 7 — --theme adds exactly one cross-project cross-cut after the
// initiative threads.
func TestBuildThreads_ThemeCrossCut(t *testing.T) {
	threads := BuildThreads(storyFixture, meThreadOptsWithTheme("perf"))

	// The four initiative threads are unchanged; one theme thread appended.
	if len(threads) != 5 {
		t.Fatalf("with theme: got %d threads, want 5 (%v)", len(threads), threadNames(threads))
	}
	theme := threads[len(threads)-1]
	if theme.Kind != KindTheme || theme.Thread != "perf" {
		t.Fatalf("theme thread: got {%q %q}, want {perf theme}", theme.Thread, theme.Kind)
	}
	if got := beatIDs(theme); len(got) != 2 || got[0] != 1 || got[1] != 6 {
		t.Errorf("theme beats: got %v, want [1 6] (time-ordered)", got)
	}
	// Initiative threads unchanged.
	if got := threadNames(threads[:4]); !equalStrings(got, []string{"alpha", "beta", "gamma", "(no project)"}) {
		t.Errorf("initiative threads changed by theme: %v", got)
	}
}

// Test 8 — throughline skeleton counts + span.
func TestBuildThroughline_SkeletonCountsAndSpan(t *testing.T) {
	threads := BuildThreads(storyFixture, meThreadOpts)
	tl := BuildThroughline(threads)
	if len(tl.Arcs) != len(threads) {
		t.Fatalf("arc count: got %d, want %d", len(tl.Arcs), len(threads))
	}
	alpha := tl.Arcs[0]
	if alpha.Thread != "alpha" || alpha.BeatCount != 2 || alpha.ImpactBeatCount != 1 {
		t.Errorf("alpha arc: got {%q b=%d ib=%d}, want {alpha 2 1}", alpha.Thread, alpha.BeatCount, alpha.ImpactBeatCount)
	}
	wantFirst := time.Date(2026, 2, 1, 10, 0, 0, 0, time.UTC)
	wantLast := time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC)
	if !alpha.Span.First.Equal(wantFirst) || !alpha.Span.Last.Equal(wantLast) {
		t.Errorf("alpha span: got {%v %v}, want {%v %v}", alpha.Span.First, alpha.Span.Last, wantFirst, wantLast)
	}

	// Empty threads → non-nil empty Arcs.
	empty := BuildThroughline(nil)
	if empty.Arcs == nil {
		t.Error("empty throughline Arcs must be non-nil")
	}
	if len(empty.Arcs) != 0 {
		t.Errorf("empty throughline: got %d arcs, want 0", len(empty.Arcs))
	}
}

// --- helpers ---

func threadNames(threads []Thread) []string {
	out := make([]string, len(threads))
	for i, t := range threads {
		out[i] = t.Thread
	}
	return out
}

func beatIDs(t Thread) []int64 {
	out := make([]int64, len(t.Beats))
	for i, b := range t.Beats {
		out[i] = b.ID
	}
	return out
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
