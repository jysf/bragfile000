package story

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jysf/bragfile000/internal/storage"
)

// failureStoryFixture holds three recorded failures (SPEC-094). Types are the
// literal "failed", never aggregate.FailureType, so a change to the persisted
// value fires these goldens rather than moving with them (SPEC-086 LD8).
//   - 2 (alpha) is a failure WITH an impact: still an impact beat (DEC-054).
//   - 4 (beta) is a failure with NO impact: labelled, no impact line.
//   - 5 (delta) is delta's ONLY impact beat, so a promotional profile must
//     fold delta — and still count 5 in its Omitted: line.
//   - 7 (gamma) is typed "Failed": a near-miss, not a failure (DEC-050 rule 1).
var failureStoryFixture = []storage.Entry{
	{ID: 1, Title: "shipped the cache", Project: "alpha", Type: "shipped",
		Impact:    "cut p95 latency 40%",
		CreatedAt: time.Date(2026, 2, 1, 10, 0, 0, 0, time.UTC)},
	{ID: 2, Title: "tried a worker pool", Project: "alpha", Type: "failed",
		Impact:    "cost two days",
		CreatedAt: time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC)},
	{ID: 3, Title: "onboarding guide", Project: "beta", Type: "shipped",
		Impact:    "onboarding down to 1 day",
		CreatedAt: time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC)},
	{ID: 4, Title: "abandoned the rewrite", Project: "beta", Type: "failed",
		CreatedAt: time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC)},
	{ID: 5, Title: "migrated to the wrong queue", Project: "delta", Type: "failed",
		Impact:    "lost a week",
		CreatedAt: time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)},
	{ID: 6, Title: "delta notes", Project: "delta", Type: "learned",
		CreatedAt: time.Date(2026, 6, 10, 10, 0, 0, 0, time.UTC)},
	{ID: 7, Title: "gamma near-miss", Project: "gamma", Type: "Failed",
		Impact:    "shipped anyway",
		CreatedAt: time.Date(2026, 6, 20, 10, 0, 0, 0, time.UTC)},
}

// candorBundleOpts is the CLI's pipeline over failureStoryFixture, minus the
// store: OmitFailures, then BuildThreads, then the renderer's options.
func candorBundleOpts(t *testing.T, audience string) StoryOptions {
	t.Helper()
	p, err := LoadProfile(audience)
	if err != nil {
		t.Fatalf("LoadProfile(%q): %v", audience, err)
	}
	shown, omitted := OmitFailures(failureStoryFixture, p)
	threads := BuildThreads(shown, ThreadOptionsFromProfile(p, ""))
	return StoryOptions{
		Audience:        audience,
		Scope:           "year",
		Filters:         "(none)",
		EntriesInWindow: len(failureStoryFixture),
		OmittedFailures: omitted,
		Now:             storyFixedNow,
		Threads:         threads,
		Throughline:     BuildThroughline(threads),
		Directive:       mustDirective(t, p.Directive),
	}
}

// TestOmitFailures_OnlyExactPromotionalOmits ▲ SPEC-094 LD1. Candor is an
// unvalidated string in a user profile, so the rule keys on exactly
// "promotional" and every other value keeps every entry. The positive case is
// in the same test: without it, a rule that never omits passes the rest.
func TestOmitFailures_OnlyExactPromotionalOmits(t *testing.T) {
	shown, omitted := OmitFailures(failureStoryFixture, Profile{Candor: "promotional"})
	if omitted != 3 {
		t.Errorf("promotional: omitted = %d, want 3 (ids 2, 4, 5)", omitted)
	}
	if got := entryIDs(shown); !equalInt64s(got, []int64{1, 3, 6, 7}) {
		t.Errorf("promotional: shown = %v, want [1 3 6 7] (7 is a near-miss, not a failure)", got)
	}

	for _, candor := range []string{"candid", "", "Promotional", "PROMOTIONAL", "promotional ", "promo", "promotion"} {
		shown, omitted := OmitFailures(failureStoryFixture, Profile{Candor: candor})
		if omitted != 0 {
			t.Errorf("candor %q: omitted = %d, want 0 — only the exact value omits", candor, omitted)
		}
		if got := entryIDs(shown); !equalInt64s(got, entryIDs(failureStoryFixture)) {
			t.Errorf("candor %q: shown = %v, want every entry", candor, got)
		}
	}
}

// TestOmitFailures_RunsBeforeTheFold ▲ SPEC-094 LD2. delta's only impact beat
// is a failure. Threaded as-is, exec keeps delta (the control); omitted first,
// delta folds under impact_threads_only and its failure is still counted.
func TestOmitFailures_RunsBeforeTheFold(t *testing.T) {
	if got := threadNames(BuildThreads(failureStoryFixture, execThreadOpts)); !containsString(got, "delta") {
		t.Fatalf("control: exec over the unfiltered fixture should keep delta, got %v", got)
	}
	shown, omitted := OmitFailures(failureStoryFixture, Profile{Candor: CandorPromotional})
	got := threadNames(BuildThreads(shown, execThreadOpts))
	if containsString(got, "delta") {
		t.Errorf("exec after OmitFailures still carries delta, whose only impact beat was a failure: %v", got)
	}
	if omitted != 3 {
		t.Errorf("omitted = %d, want 3: delta's failure folds with its thread and is still counted", omitted)
	}
}

// TestBuildThreads_AFailureWithImpactIsStillAnImpactBeat ▲ SPEC-094 LD5
// (framing's Question 2). is_impact_beat keeps meaning "non-empty impact", so
// no count changes what it counts (DEC-048); IsFailure is what renders it apart.
func TestBuildThreads_AFailureWithImpactIsStillAnImpactBeat(t *testing.T) {
	threads := BuildThreads(failureStoryFixture, meThreadOpts)
	type flags struct{ impact, failure bool }
	want := map[int64]flags{
		1: {true, false}, 2: {true, true}, 3: {true, false}, 4: {false, true},
		5: {true, true}, 6: {false, false}, 7: {true, false},
	}
	seen := 0
	for _, thr := range threads {
		for _, b := range thr.Beats {
			seen++
			if got := (flags{b.IsImpactBeat, b.IsFailure}); got != want[b.ID] {
				t.Errorf("beat %d: {IsImpactBeat IsFailure} = %v, want %v", b.ID, got, want[b.ID])
			}
		}
	}
	if seen != 7 {
		t.Fatalf("me kept %d beats, want all 7", seen)
	}
	for _, a := range BuildThroughline(threads).Arcs {
		if a.Thread == "delta" && a.ImpactBeatCount != 1 {
			t.Errorf("delta impact_beat_count = %d, want 1: its failure carries an impact", a.ImpactBeatCount)
		}
	}
}

// TestStoryPackage_CallsTheSharedPredicates ▲ SPEC-094 LD6. The impact and
// failure rules live in internal/aggregate; a copy restated here is one a
// change there leaves behind with every behavioural test green, which is what
// thread.go:135 and :87 were until SPEC-094. Scans non-test sources only.
func TestStoryPackage_CallsTheSharedPredicates(t *testing.T) {
	forbidden := []string{`Impact != ""`, `Impact == ""`, `"failed"`}
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	scanned := 0
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		scanned++
		for _, needle := range forbidden {
			if strings.Contains(string(b), needle) {
				t.Errorf("%s restates a predicate (%s); call aggregate.HasImpact / aggregate.IsFailure instead", f, needle)
			}
		}
	}
	if scanned < 4 {
		t.Fatalf("scanned %d source files, want at least 4 (bundle, embed, profile, thread)", scanned)
	}
}

// TestToStoryMarkdown_CandidLabelsFailuresGolden ▲ SPEC-094 LD3/LD5. A candid
// profile lists every failure where it falls, as "✗ <id> (failed)", with its
// impact line when it has one. ★ never marks a failure. There is no Omitted:
// line and the directive is the asset, unchanged.
func TestToStoryMarkdown_CandidLabelsFailuresGolden(t *testing.T) {
	withOverrideDir(t, t.TempDir())
	got, err := ToStoryMarkdown(candorBundleOpts(t, "me"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := `# Bragfile Story

Generated: 2026-07-06T12:00:00Z
Scope: year
Audience: me
Filters: (none)
Threads: 4
Beats: 7/7

## Threads

### alpha

- ★ 1: shipped the cache
  cut p95 latency 40%
- ✗ 2 (failed): tried a worker pool
  cost two days

### beta

- ★ 3: onboarding guide
  onboarding down to 1 day
- ✗ 4 (failed): abandoned the rewrite

### delta

- ✗ 5 (failed): migrated to the wrong queue
  lost a week
- · 6: delta notes

### gamma

- ★ 7: gamma near-miss
  shipped anyway

## Throughline (skeleton)

- alpha [initiative]: 2 beats, 2 with impact (2026-02-01 → 2026-03-01)
- beta [initiative]: 2 beats, 1 with impact (2026-04-01 → 2026-05-01)
- delta [initiative]: 2 beats, 1 with impact (2026-06-01 → 2026-06-10)
- gamma [initiative]: 1 beat, 1 with impact (2026-06-20 → 2026-06-20)

## Framing directive

` + mustDirectiveTrimmed(t, "me.md")
	if string(got) != want {
		t.Errorf("candid markdown golden mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

// TestToStoryMarkdown_PromotionalOmitsWithNoteAndClauseGolden ▲ SPEC-094
// LD1/LD2/LD3/LD4. exec omits the three failures, says so under Beats:, folds
// delta, and ends its directive with the fixed clause. The near-miss (7) stays.
func TestToStoryMarkdown_PromotionalOmitsWithNoteAndClauseGolden(t *testing.T) {
	withOverrideDir(t, t.TempDir())
	got, err := ToStoryMarkdown(candorBundleOpts(t, "exec"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := `# Bragfile Story

Generated: 2026-07-06T12:00:00Z
Scope: year
Audience: exec
Filters: (none)
Threads: 3
Beats: 3/7
Omitted: 3 recorded failures, not listed for this audience (brag list --type failed)

## Threads

### alpha

- ★ 1: shipped the cache
  cut p95 latency 40%

### beta

- ★ 3: onboarding guide
  onboarding down to 1 day

### gamma

- ★ 7: gamma near-miss
  shipped anyway

## Throughline (skeleton)

- alpha [initiative]: 1 beat, 1 with impact (2026-02-01 → 2026-02-01)
- beta [initiative]: 1 beat, 1 with impact (2026-04-01 → 2026-04-01)
- gamma [initiative]: 1 beat, 1 with impact (2026-06-20 → 2026-06-20)

## Framing directive

` + mustDirectiveTrimmed(t, "exec.md") + `

This bundle omits 3 recorded failures for this audience. End with one line that says so; do not drop it.`
	if string(got) != want {
		t.Errorf("promotional markdown golden mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

// TestToStoryJSON_PromotionalCountsWhatItOmitted ▲ SPEC-094 LD3/LD4. The JSON
// carries the count as omitted_failure_count, directly after filters; no beat
// is a failure; and framing_directive carries the same clause the markdown
// does. The candid half is the pair: the key is present at 0, the failures
// are in the beats, and the directive is the asset byte for byte.
func TestToStoryJSON_PromotionalCountsWhatItOmitted(t *testing.T) {
	withOverrideDir(t, t.TempDir())
	type envelope struct {
		OmittedFailureCount *int `json:"omitted_failure_count"`
		Threads             []struct {
			Beats []struct {
				ID   int64  `json:"id"`
				Type string `json:"type"`
			} `json:"beats"`
		} `json:"threads"`
		FramingDirective string `json:"framing_directive"`
	}
	decode := func(audience string) (envelope, string) {
		body, err := ToStoryJSON(candorBundleOpts(t, audience))
		if err != nil {
			t.Fatalf("%s: %v", audience, err)
		}
		var env envelope
		if err := json.Unmarshal(body, &env); err != nil {
			t.Fatalf("%s: unmarshal: %v\n%s", audience, err, body)
		}
		return env, string(body)
	}
	failedBeats := func(env envelope) []int64 {
		var out []int64
		for _, th := range env.Threads {
			for _, b := range th.Beats {
				if b.Type == "failed" {
					out = append(out, b.ID)
				}
			}
		}
		return out
	}

	execEnv, execBody := decode("exec")
	if execEnv.OmittedFailureCount == nil || *execEnv.OmittedFailureCount != 3 {
		t.Errorf("exec omitted_failure_count = %v, want 3", execEnv.OmittedFailureCount)
	}
	if got := failedBeats(execEnv); len(got) != 0 {
		t.Errorf("exec JSON still carries failure beats %v", got)
	}
	clause := "This bundle omits 3 recorded failures for this audience. End with one line that says so; do not drop it.\n"
	if !strings.HasSuffix(execEnv.FramingDirective, "\n\n"+clause) {
		t.Errorf("exec framing_directive should end with the clause as its own paragraph, got:\n%q", execEnv.FramingDirective)
	}
	if !strings.Contains(execBody, "\"filters\": {},\n  \"omitted_failure_count\": 3,\n  \"threads\": [") {
		t.Errorf("omitted_failure_count must sit between filters and threads:\n%s", execBody)
	}

	meEnv, _ := decode("me")
	if meEnv.OmittedFailureCount == nil || *meEnv.OmittedFailureCount != 0 {
		t.Errorf("me omitted_failure_count = %v, want 0 and present", meEnv.OmittedFailureCount)
	}
	if got := failedBeats(meEnv); !equalInt64s(got, []int64{2, 4, 5}) {
		t.Errorf("me JSON failure beats = %v, want [2 4 5]", got)
	}
	if meEnv.FramingDirective != mustDirective(t, "me.md") {
		t.Errorf("me framing_directive must be the asset unchanged, got:\n%q", meEnv.FramingDirective)
	}
}

// TestToStory_EveryBeatOmittedStillSaysWhy ▲ SPEC-094 LD4 (the empty state,
// DEC-014 part 4). When every in-window entry is an omitted failure — `brag
// story --audience exec --type failed` — the threads are empty and the note
// is the only thing that says why. The zero-omission run of the same empty
// shape is the pair: no Omitted: line, and the directive is the asset.
func TestToStory_EveryBeatOmittedStillSaysWhy(t *testing.T) {
	p := Profile{Candor: CandorPromotional}
	onlyFailures := []storage.Entry{failureStoryFixture[1], failureStoryFixture[4]}
	shown, omitted := OmitFailures(onlyFailures, p)
	threads := BuildThreads(shown, execThreadOpts)
	opts := StoryOptions{
		Audience:        "exec",
		Scope:           "year",
		Filters:         "--type failed",
		FiltersJSON:     map[string]string{"type": "failed"},
		EntriesInWindow: len(onlyFailures),
		OmittedFailures: omitted,
		Now:             storyFixedNow,
		Threads:         threads,
		Throughline:     BuildThroughline(threads),
		Directive:       mustDirective(t, "exec.md"),
	}
	md, err := ToStoryMarkdown(opts)
	if err != nil {
		t.Fatalf("markdown: %v", err)
	}
	wantHead := "# Bragfile Story\n\nGenerated: 2026-07-06T12:00:00Z\nScope: year\nAudience: exec\nFilters: --type failed\nThreads: 0\nBeats: 0/2\nOmitted: 2 recorded failures, not listed for this audience (brag list --type failed)\n\n## Framing directive\n"
	if !strings.HasPrefix(string(md), wantHead) {
		t.Errorf("empty-state head mismatch:\n--- got ---\n%s\n--- want prefix ---\n%s", md, wantHead)
	}
	if !strings.HasSuffix(string(md), "\n\nThis bundle omits 2 recorded failures for this audience. End with one line that says so; do not drop it.") {
		t.Errorf("empty-state bundle must still end with the clause:\n%s", md)
	}
	body, err := ToStoryJSON(opts)
	if err != nil {
		t.Fatalf("json: %v", err)
	}
	if !strings.Contains(string(body), `"omitted_failure_count": 2,`) || !strings.Contains(string(body), `"threads": [],`) {
		t.Errorf("empty-state JSON should carry the count and an empty threads array:\n%s", body)
	}

	opts.OmittedFailures = 0
	md, err = ToStoryMarkdown(opts)
	if err != nil {
		t.Fatalf("markdown (no omission): %v", err)
	}
	if strings.Contains(string(md), "\nOmitted: ") {
		t.Errorf("zero omitted must render no Omitted: line:\n%s", md)
	}
	if !strings.HasSuffix(string(md), mustDirectiveTrimmed(t, "exec.md")) {
		t.Errorf("zero omitted must leave the directive as the asset:\n%s", md)
	}
}

// TestFramingDirective_ClauseAloneWhenDirectiveEmpty ▲ SPEC-094 LD4. A
// promotional user profile may carry no directive at all; the clause is then
// the whole directive, so the instruction still reaches the model. Also pins
// the singular form on both the note and the clause.
func TestFramingDirective_ClauseAloneWhenDirectiveEmpty(t *testing.T) {
	opts := StoryOptions{
		Audience:        "mine",
		Scope:           "year",
		Filters:         "(none)",
		EntriesInWindow: 1,
		OmittedFailures: 1,
		Now:             storyFixedNow,
		Threads:         []Thread{},
		Throughline:     BuildThroughline(nil),
		Directive:       "",
	}
	md, err := ToStoryMarkdown(opts)
	if err != nil {
		t.Fatalf("markdown: %v", err)
	}
	want := "# Bragfile Story\n\nGenerated: 2026-07-06T12:00:00Z\nScope: year\nAudience: mine\nFilters: (none)\nThreads: 0\nBeats: 0/1\nOmitted: 1 recorded failure, not listed for this audience (brag list --type failed)\n\n## Framing directive\n\nThis bundle omits 1 recorded failure for this audience. End with one line that says so; do not drop it."
	if string(md) != want {
		t.Errorf("clause-only directive mismatch:\n--- got ---\n%s\n--- want ---\n%s", md, want)
	}
	body, err := ToStoryJSON(opts)
	if err != nil {
		t.Fatalf("json: %v", err)
	}
	var env struct {
		FramingDirective string `json:"framing_directive"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if env.FramingDirective != "This bundle omits 1 recorded failure for this audience. End with one line that says so; do not drop it.\n" {
		t.Errorf("framing_directive = %q, want the clause alone", env.FramingDirective)
	}
}

// TestLoadProfile_OnlyExactPromotionalOmitsFailures ▲ SPEC-094 LD1. The
// bundled data decides which built-ins omit (exec, skip) and which label (me,
// manager); a user profile omits only on the exact value, and a promotional
// user profile's own directive file still gets the clause appended.
func TestLoadProfile_OnlyExactPromotionalOmitsFailures(t *testing.T) {
	withOverrideDir(t, t.TempDir())
	for name, want := range map[string]bool{"exec": true, "skip": true, "me": false, "manager": false} {
		p, err := LoadProfile(name)
		if err != nil {
			t.Fatalf("LoadProfile(%q): %v", name, err)
		}
		if p.OmitsFailures() != want {
			t.Errorf("bundled %s: OmitsFailures = %v, want %v (candor %q)", name, p.OmitsFailures(), want, p.Candor)
		}
	}

	dir := t.TempDir()
	withOverrideDir(t, dir)
	directivePath := filepath.Join(dir, "board.md")
	if err := os.WriteFile(directivePath, []byte("Pitch this to the board.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	profiles := map[string]string{
		"board":  "candor: promotional\ndirective: " + directivePath + "\n",
		"typo":   "candor: Promotional\n",
		"promo":  "candor: promo\n",
		"silent": "thread_order: initiative\n",
	}
	for name, body := range profiles {
		if err := os.WriteFile(filepath.Join(dir, name+".yaml"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for name, want := range map[string]bool{"board": true, "typo": false, "promo": false, "silent": false} {
		p, err := LoadProfile(name)
		if err != nil {
			t.Fatalf("LoadProfile(%q): %v", name, err)
		}
		if p.OmitsFailures() != want {
			t.Errorf("override %s (candor %q): OmitsFailures = %v, want %v", name, p.Candor, p.OmitsFailures(), want)
		}
	}

	board, _ := LoadProfile("board")
	directive, err := ResolveDirective(board)
	if err != nil {
		t.Fatalf("ResolveDirective: %v", err)
	}
	md, err := ToStoryMarkdown(StoryOptions{
		Audience: "board", Scope: "year", Filters: "(none)", EntriesInWindow: 2,
		OmittedFailures: 2, Now: storyFixedNow, Directive: directive,
	})
	if err != nil {
		t.Fatalf("markdown: %v", err)
	}
	if !strings.HasSuffix(string(md), "## Framing directive\n\nPitch this to the board.\n\nThis bundle omits 2 recorded failures for this audience. End with one line that says so; do not drop it.") {
		t.Errorf("a user directive file must get the clause appended:\n%s", md)
	}
}

// --- helpers ---

func entryIDs(entries []storage.Entry) []int64 {
	out := make([]int64, len(entries))
	for i, e := range entries {
		out[i] = e.ID
	}
	return out
}

func equalInt64s(a, b []int64) bool {
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

func containsString(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
