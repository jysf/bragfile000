package story

import (
	"strings"
	"testing"
)

// Test P1 — TestLoadProfile_Manager (bundled load, LOAD-BEARING for the
// config). Asserts the manager profile loads from the embedded FS with the
// gradient-locked values: keep-all body (like me), month cadence, tactical
// directive.
func TestLoadProfile_Manager(t *testing.T) {
	withOverrideDir(t, t.TempDir()) // bundled-only
	p, err := LoadProfile("manager")
	if err != nil {
		t.Fatalf("LoadProfile(manager): %v", err)
	}
	if p.Name != "manager" || p.DefaultWindow != "month" {
		t.Errorf("got name=%q window=%q, want manager/month", p.Name, p.DefaultWindow)
	}
	// Keep-all body policy (same as me): nothing folded, nothing dropped.
	if p.ImpactThreadsOnly || p.DropImpactlessBeats || p.FoldSmallThreads {
		t.Errorf("manager keeps all threads/beats, got %+v", p)
	}
	if p.ThreadOrder != "initiative" || p.Directive != "manager.md" {
		t.Errorf("got order=%q directive=%q, want initiative/manager.md", p.ThreadOrder, p.Directive)
	}
	if p.Candor != "candid" {
		t.Errorf("manager candor: got %q, want candid", p.Candor)
	}
}

// Test P2 — TestLoadProfile_Skip (bundled load, LOAD-BEARING for the
// config). Asserts skip folds zero-impact threads (like exec) but KEEPS
// their non-impact beats (unlike exec), grouped by initiative.
func TestLoadProfile_Skip(t *testing.T) {
	withOverrideDir(t, t.TempDir())
	p, err := LoadProfile("skip")
	if err != nil {
		t.Fatalf("LoadProfile(skip): %v", err)
	}
	if p.Name != "skip" || p.DefaultWindow != "quarter" {
		t.Errorf("got name=%q window=%q, want skip/quarter", p.Name, p.DefaultWindow)
	}
	// Fold zero-impact threads (like exec) ...
	if !p.ImpactThreadsOnly || !p.FoldSmallThreads {
		t.Errorf("skip folds zero-impact threads, got %+v", p)
	}
	// ... but KEEP the non-impact beats inside surfaced threads (UNLIKE exec).
	if p.DropImpactlessBeats {
		t.Errorf("skip must KEEP non-impact beats (drop_impactless_beats=false), got true")
	}
	// Group by initiative, NOT exec's one-headline impact-desc.
	if p.ThreadOrder != "initiative" {
		t.Errorf("skip order: got %q, want initiative", p.ThreadOrder)
	}
	if p.Directive != "skip.md" || p.Candor != "promotional" {
		t.Errorf("got directive=%q candor=%q, want skip.md/promotional", p.Directive, p.Candor)
	}
}

// Test P3 — TestProfiles_FourWayGradientDivergence (LOAD-BEARING — the
// divergence assertion vs me/exec). Loads all four bundled profiles and
// asserts (a) no two are byte-identical as parsed structs, (b) the skip
// body over storyFixture differs from BOTH me's and exec's on the axes
// that place it strictly between them.
func TestProfiles_FourWayGradientDivergence(t *testing.T) {
	withOverrideDir(t, t.TempDir())
	names := []string{"me", "manager", "skip", "exec"}
	profs := map[string]Profile{}
	for _, n := range names {
		p, err := LoadProfile(n)
		if err != nil {
			t.Fatalf("LoadProfile(%s): %v", n, err)
		}
		profs[n] = p
	}

	// (a) All four distinct as parsed structs (no two identical).
	for i := 0; i < len(names); i++ {
		for j := i + 1; j < len(names); j++ {
			if profs[names[i]] == profs[names[j]] {
				t.Errorf("%s and %s parse to identical profiles: %+v",
					names[i], names[j], profs[names[i]])
			}
		}
	}

	// (b) Body divergence over the SAME corpus (storyFixture):
	threadsFor := func(n string) []Thread {
		return BuildThreads(storyFixture, ThreadOptionsFromProfile(profs[n], ""))
	}
	me := threadNames(threadsFor("me"))
	skip := threadNames(threadsFor("skip"))
	exec := threadNames(threadsFor("exec"))

	// me keeps all four threads (incl. (no project)); skip folds (no project)
	// but keeps the three impact-bearing initiatives in alpha-ASC order.
	wantMe := []string{"alpha", "beta", "gamma", "(no project)"}
	wantSkip := []string{"alpha", "beta", "gamma"} // (no project) folded, initiative order
	wantExec := []string{"beta", "alpha", "gamma"} // impact-desc headline order
	assertEqualSlice(t, "me threads", me, wantMe)
	assertEqualSlice(t, "skip threads", skip, wantSkip)
	assertEqualSlice(t, "exec threads", exec, wantExec)

	// The skip-vs-exec distinction: skip KEEPS alpha-messy (impact-less beat
	// in a surfaced thread); exec DROPS it. Find alpha in each.
	skipAlpha := findThread(threadsFor("skip"), "alpha")
	execAlpha := findThread(threadsFor("exec"), "alpha")
	if beatCount(skipAlpha) != 2 {
		t.Errorf("skip alpha should keep both beats (impact-less kept), got %d", beatCount(skipAlpha))
	}
	if beatCount(execAlpha) != 1 {
		t.Errorf("exec alpha should drop the impact-less beat, got %d", beatCount(execAlpha))
	}
}

// Test P4 — TestDirectives_ManagerSkip_ResolveAndVoice. Both new
// directives resolve from the embedded FS, are non-empty, pairwise-distinct
// from all four, and carry their voice's signature token.
func TestDirectives_ManagerSkip_ResolveAndVoice(t *testing.T) {
	mgr, err := directiveAsset("manager.md")
	if err != nil {
		t.Fatalf("directiveAsset(manager.md): %v", err)
	}
	skip, err := directiveAsset("skip.md")
	if err != nil {
		t.Fatalf("directiveAsset(skip.md): %v", err)
	}
	me := mustDirective(t, "me.md")
	exec := mustDirective(t, "exec.md")

	for _, d := range [][]byte{mgr, skip} {
		if len(d) == 0 {
			t.Fatal("manager/skip directives must be non-empty")
		}
	}
	// All four directives are pairwise distinct.
	all := map[string]string{"me": me, "exec": exec,
		"manager": string(mgr), "skip": string(skip)}
	seen := map[string]string{}
	for name, body := range all {
		if other, dup := seen[body]; dup {
			t.Errorf("%s and %s directives are byte-identical", name, other)
		}
		seen[body] = name
	}
	// On-voice substrings (the tactical vs outcomes-by-initiative split).
	if !strings.Contains(string(mgr), "blockers") {
		t.Errorf("manager directive should be tactical (mention blockers)")
	}
	if !strings.Contains(string(skip), "initiative") {
		t.Errorf("skip directive should frame outcomes by initiative")
	}
}

// --- P3 helpers (test-only) ---

// findThread returns the thread with the given name, or a zero Thread if
// absent (callers assert on beatCount, so a zero Thread reads as 0 beats).
func findThread(threads []Thread, name string) Thread {
	for _, t := range threads {
		if t.Thread == name {
			return t
		}
	}
	return Thread{}
}

// beatCount is the number of surfaced beats in a thread.
func beatCount(t Thread) int {
	return len(t.Beats)
}

// assertEqualSlice fails the test if got != want (order-sensitive).
func assertEqualSlice(t *testing.T, label string, got, want []string) {
	t.Helper()
	if !equalStrings(got, want) {
		t.Errorf("%s: got %v, want %v", label, got, want)
	}
}
