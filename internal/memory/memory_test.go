package memory

import (
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jysf/bragfile000/internal/storage"
)

// mustTime parses an RFC3339 UTC timestamp, panicking on a malformed literal
// (a test-fixture bug, never a runtime condition).
func mustTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t.UTC()
}

// fixtureEntries is the shared 8-entry fixture transcribed verbatim from the
// spec (SPEC-073 Notes → "The fixture"). Every rendering branch and every
// ranking branch is exercised: both-fields-present, no-project, no-type,
// no-project-and-no-type, impact-present, impact-absent, in-both-lists,
// in-match-only, in-project-only.
//
// FOUR packages transcribe these 8 rows — internal/memory (here),
// internal/export, internal/cli, and internal/storage. An edit here must be
// mirrored to all of them, especially internal/storage's
// TestSearch_SPEC073FixtureOrdersAuthQuery: that one would keep passing against
// its own divergent copy while silently no longer pinning the bm25 order the
// CLI actually queries.
func fixtureEntries() []storage.Entry {
	return []storage.Entry{
		{ID: 8, CreatedAt: mustTime("2026-08-07T09:15:00Z"), Project: "bragfile", Type: "shipped", Title: "MCP list filter parity", Impact: "agents can ask for a bounded window in one call"},
		{ID: 7, CreatedAt: mustTime("2026-08-05T17:40:00Z"), Project: "orbit", Type: "learned", Title: "Retry storms need jitter"},
		{ID: 6, CreatedAt: mustTime("2026-08-01T11:02:00Z"), Project: "bragfile", Type: "shipped", Title: "Sparkline pulse command"},
		{ID: 5, CreatedAt: mustTime("2026-07-20T08:00:00Z"), Title: "Read the auth spec"},
		{ID: 4, CreatedAt: mustTime("2026-06-11T14:25:00Z"), Project: "orbit", Type: "shipped", Title: "Auth refactor for the gateway", Impact: "cut p99 login latency from 600ms to 120ms"},
		{ID: 3, CreatedAt: mustTime("2026-05-02T19:30:00Z"), Project: "bragfile", Type: "learned", Title: "FTS5 triggers fight attach"},
		{ID: 2, CreatedAt: mustTime("2026-03-14T10:10:00Z"), Project: "orbit", Type: "shipped", Title: "Auth token rotation", Impact: "removed the last shared secret"},
		{ID: 1, CreatedAt: mustTime("2026-01-09T21:45:00Z"), Type: "learned", Title: "Auth is mostly caching"},
	}
}

func idsOf(r Result) []int64 {
	ids := make([]int64, len(r.Items))
	for i, it := range r.Items {
		ids[i] = it.Entry.ID
	}
	return ids
}

func scoreOf(t *testing.T, r Result, id int64) float64 {
	t.Helper()
	for _, it := range r.Items {
		if it.Entry.ID == id {
			return it.Score
		}
	}
	t.Fatalf("id %d not present in result", id)
	return 0
}

func indexOf(ids []int64, id int64) int {
	for i, v := range ids {
		if v == id {
			return i
		}
	}
	return -1
}

func equalIDs(a, b []int64) bool {
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

func round9(f float64) float64 {
	return math.Round(f*1e9) / 1e9
}

// TestSlice_DegeneratesToRecencyWithNoQueryOrProject pins DEC-043's "one
// algorithm, both cases": with no query and no project, Slice's output
// order is exactly created_at DESC, id DESC — identical to Store.List.
func TestSlice_DegeneratesToRecencyWithNoQueryOrProject(t *testing.T) {
	result := Slice(fixtureEntries(), Options{Budget: 2000})
	want := []int64{8, 7, 6, 5, 4, 3, 2, 1}
	if got := idsOf(result); !equalIDs(got, want) {
		t.Errorf("ids = %v, want %v", got, want)
	}
}

// TestSlice_BlendDiffersFromBothInputs pins the stage success criterion:
// with a query and a project, the output order differs from both plain
// recency and plain bm25 over the shared fixture.
func TestSlice_BlendDiffersFromBothInputs(t *testing.T) {
	result := Slice(fixtureEntries(), Options{
		Query:   "auth",
		Project: "orbit",
		Matched: []int64{5, 1, 2, 4},
		Budget:  2000,
	})
	want := []int64{4, 2, 7, 5, 1, 8, 6, 3}
	got := idsOf(result)
	if !equalIDs(got, want) {
		t.Fatalf("ids = %v, want %v", got, want)
	}

	plainRecency := []int64{8, 7, 6, 5, 4, 3, 2, 1}
	if equalIDs(got, plainRecency) {
		t.Errorf("blend matches plain recency: %v", got)
	}
	plainBM25 := []int64{5, 1, 2, 4}
	if len(got) >= len(plainBM25) && equalIDs(got[:len(plainBM25)], plainBM25) {
		t.Errorf("blend matches a prefix of plain bm25: %v", got)
	}
}

// TestSlice_ScoresMatchTheLockedConstants pins FusionK and the three unit
// weights: each item's Score equals the hand-computable sum (SPEC-073
// Notes → §12(b) pre-flight → A), to 9 decimal places.
func TestSlice_ScoresMatchTheLockedConstants(t *testing.T) {
	result := Slice(fixtureEntries(), Options{
		Query:   "auth",
		Project: "orbit",
		Matched: []int64{5, 1, 2, 4},
		Budget:  2000,
	})
	want := map[int64]float64{
		8: 0.016393443,
		7: 0.032522475,
		6: 0.015873016,
		5: 0.032018443,
		4: 0.047138648,
		3: 0.015151515,
		2: 0.046671405,
		1: 0.030834915,
	}
	for id, w := range want {
		got := round9(scoreOf(t, result, id))
		if got != w {
			t.Errorf("id %d: score = %.9f, want %.9f", id, got, w)
		}
	}
}

// TestSlice_AbsentFromAListContributesZero: an entry in no Matched and no
// project scores exactly 1/(60+rank_recency).
func TestSlice_AbsentFromAListContributesZero(t *testing.T) {
	result := Slice(fixtureEntries(), Options{
		Query:   "auth",
		Project: "orbit",
		Matched: []int64{5, 1, 2, 4},
		Budget:  2000,
	})
	// id 8 is absent from both the match list and the project (bragfile,
	// not orbit); its recency rank is 1 (the most recent entry).
	want := 1.0 / (60 + 1)
	if got := scoreOf(t, result, 8); math.Abs(got-want) > 1e-12 {
		t.Errorf("score(8) = %.12f, want %.12f", got, want)
	}
}

// TestSlice_MatchedIgnoresUnknownAndDuplicateIDs: Matched with an unknown id
// and a duplicate produces the same ranking as the deduped, in-pool subset.
func TestSlice_MatchedIgnoresUnknownAndDuplicateIDs(t *testing.T) {
	a := Slice(fixtureEntries(), Options{Matched: []int64{999, 5, 5, 1}, Budget: 2000})
	b := Slice(fixtureEntries(), Options{Matched: []int64{5, 1}, Budget: 2000})
	if !equalIDs(idsOf(a), idsOf(b)) {
		t.Errorf("ids = %v, want %v", idsOf(a), idsOf(b))
	}
}

// TestSlice_ProjectListIsRecencyOrdered: with Project "orbit" and no query,
// 7 (project rank 1) outranks 4 (project rank 2) outranks 2, and all three
// outrank every non-orbit entry.
func TestSlice_ProjectListIsRecencyOrdered(t *testing.T) {
	result := Slice(fixtureEntries(), Options{Project: "orbit", Budget: 2000})
	ids := idsOf(result)
	i7, i4, i2 := indexOf(ids, 7), indexOf(ids, 4), indexOf(ids, 2)
	if !(i7 < i4 && i4 < i2) {
		t.Fatalf("expected 7 < 4 < 2 in rank order, got positions 7=%d 4=%d 2=%d (%v)", i7, i4, i2, ids)
	}
	for _, other := range []int64{8, 6, 5, 3, 1} {
		io := indexOf(ids, other)
		if io < i2 {
			t.Errorf("non-orbit entry %d (pos %d) outranks orbit entry 2 (pos %d): %v", other, io, i2, ids)
		}
	}
}

// TestSlice_UnknownProjectAddsNoTerm: Project "nope" (no matching entries)
// yields the identical ordering to Project "".
func TestSlice_UnknownProjectAddsNoTerm(t *testing.T) {
	a := Slice(fixtureEntries(), Options{Project: "nope", Budget: 2000})
	b := Slice(fixtureEntries(), Options{Project: "", Budget: 2000})
	if !equalIDs(idsOf(a), idsOf(b)) {
		t.Errorf("ids = %v, want %v", idsOf(a), idsOf(b))
	}
}

// TestSlice_InputOrderInvariantAndDedupes: the fixture shuffled, with entry
// 8 appended a second time, yields identical Items and Candidates == 8.
func TestSlice_InputOrderInvariantAndDedupes(t *testing.T) {
	base := fixtureEntries()
	shuffled := []storage.Entry{base[3], base[7], base[1], base[5], base[0], base[6], base[4], base[2], base[0]}

	want := Slice(fixtureEntries(), Options{Budget: 2000})
	got := Slice(shuffled, Options{Budget: 2000})

	if got.Candidates != 8 {
		t.Errorf("Candidates = %d, want 8", got.Candidates)
	}
	if !equalIDs(idsOf(got), idsOf(want)) {
		t.Errorf("ids = %v, want %v", idsOf(got), idsOf(want))
	}
	for i := range want.Items {
		if got.Items[i].Line != want.Items[i].Line {
			t.Errorf("item %d: Line = %q, want %q", i, got.Items[i].Line, want.Items[i].Line)
		}
	}
}

// TestSlice_EqualScoresTieBreakByRecencyThenID: two entries with identical
// CreatedAt and no match/project terms order by id DESC.
func TestSlice_EqualScoresTieBreakByRecencyThenID(t *testing.T) {
	tie := mustTime("2026-06-01T00:00:00Z")
	entries := []storage.Entry{
		{ID: 10, CreatedAt: tie, Title: "a"},
		{ID: 20, CreatedAt: tie, Title: "b"},
	}
	result := Slice(entries, Options{Budget: 2000})
	want := []int64{20, 10}
	if got := idsOf(result); !equalIDs(got, want) {
		t.Errorf("ids = %v, want %v", got, want)
	}
}

// TestRenderLine_ShapeAndAbsentFields covers the four bracket forms, the
// UTC YYYY-MM-DD date, and the no-trailing-dash-on-empty-impact rule.
// Expected strings transcribed verbatim from SPEC-073 Notes → "Per-entry
// token costs".
func TestRenderLine_ShapeAndAbsentFields(t *testing.T) {
	cases := map[int64]string{
		8: "- 8 2026-08-07 [bragfile/shipped] MCP list filter parity — agents can ask for a bounded window in one call",
		7: "- 7 2026-08-05 [orbit/learned] Retry storms need jitter",
		6: "- 6 2026-08-01 [bragfile/shipped] Sparkline pulse command",
		5: "- 5 2026-07-20 [-/-] Read the auth spec",
		4: "- 4 2026-06-11 [orbit/shipped] Auth refactor for the gateway — cut p99 login latency from 600ms to 120ms",
		3: "- 3 2026-05-02 [bragfile/learned] FTS5 triggers fight attach",
		2: "- 2 2026-03-14 [orbit/shipped] Auth token rotation — removed the last shared secret",
		1: "- 1 2026-01-09 [-/learned] Auth is mostly caching",
	}
	byID := map[int64]storage.Entry{}
	for _, e := range fixtureEntries() {
		byID[e.ID] = e
	}
	for id, want := range cases {
		got := RenderLine(byID[id])
		if got != want {
			t.Errorf("id %d: RenderLine = %q, want %q", id, got, want)
		}
	}
}

// TestEstimateTokens_CeilOverBytes pins ceil(len(bytes)/4), including
// multi-byte-rune BYTE semantics (a 3-byte rune counts as 3, not 1).
func TestEstimateTokens_CeilOverBytes(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{"", 0},
		{"a", 1},
		{"abcd", 1},
		{"abcde", 2},
		{"世界", 2}, // 2 runes, 6 UTF-8 bytes: ceil(6/4)=2, not ceil(2/4)=1
	}
	for _, c := range cases {
		if got := EstimateTokens(c.in); got != c.want {
			t.Errorf("EstimateTokens(%q) = %d, want %d", c.in, got, c.want)
		}
	}
}

// TestSlice_EstimateMatchesRenderedBody pins "the estimate and the render
// cannot disagree": Result.EstimatedTokens equals the sum of
// EstimateTokens(RenderLine(entry)) over the included set, and each
// Item.Tokens equals EstimateTokens(Item.Line).
func TestSlice_EstimateMatchesRenderedBody(t *testing.T) {
	result := Slice(fixtureEntries(), Options{Budget: 2000})
	sum := 0
	for _, it := range result.Items {
		if it.Tokens != EstimateTokens(it.Line) {
			t.Errorf("id %d: Tokens = %d, want %d", it.Entry.ID, it.Tokens, EstimateTokens(it.Line))
		}
		if it.Line != RenderLine(it.Entry) {
			t.Errorf("id %d: Line = %q, want %q", it.Entry.ID, it.Line, RenderLine(it.Entry))
		}
		sum += EstimateTokens(RenderLine(it.Entry))
	}
	if result.EstimatedTokens != sum {
		t.Errorf("EstimatedTokens = %d, want %d", result.EstimatedTokens, sum)
	}
}

// TestSlice_CountsPartitionTheCandidatePool: Included + Skipped ==
// Candidates across three budgets.
func TestSlice_CountsPartitionTheCandidatePool(t *testing.T) {
	for _, budget := range []int{2000, 40, 1} {
		result := Slice(fixtureEntries(), Options{Budget: budget})
		if result.Included+result.Skipped != result.Candidates {
			t.Errorf("budget %d: Included(%d)+Skipped(%d) != Candidates(%d)",
				budget, result.Included, result.Skipped, result.Candidates)
		}
	}
}

// TestSlice_SkipsOversizeAndContinues pins DEC-044's skip-and-continue
// against stop-at-first-overflow (which would give [8]/1/27).
func TestSlice_SkipsOversizeAndContinues(t *testing.T) {
	result := Slice(fixtureEntries(), Options{Budget: 40})
	want := []int64{8, 5}
	if got := idsOf(result); !equalIDs(got, want) {
		t.Fatalf("ids = %v, want %v", got, want)
	}
	if result.Included != 2 {
		t.Errorf("Included = %d, want 2", result.Included)
	}
	if result.Skipped != 6 {
		t.Errorf("Skipped = %d, want 6", result.Skipped)
	}
	if result.EstimatedTokens != 37 {
		t.Errorf("EstimatedTokens = %d, want 37", result.EstimatedTokens)
	}
}

// TestSlice_BudgetTooSmallForAnyEntryIncludesNothing: Budget 1 yields
// Included 0, Skipped 8, EstimatedTokens 0, and a non-nil empty Items.
func TestSlice_BudgetTooSmallForAnyEntryIncludesNothing(t *testing.T) {
	result := Slice(fixtureEntries(), Options{Budget: 1})
	if result.Included != 0 {
		t.Errorf("Included = %d, want 0", result.Included)
	}
	if result.Skipped != 8 {
		t.Errorf("Skipped = %d, want 8", result.Skipped)
	}
	if result.EstimatedTokens != 0 {
		t.Errorf("EstimatedTokens = %d, want 0", result.EstimatedTokens)
	}
	if result.Items == nil {
		t.Error("Items is nil, want non-nil empty slice")
	}
	if len(result.Items) != 0 {
		t.Errorf("len(Items) = %d, want 0", len(result.Items))
	}
}

// TestSlice_EmptyPool: Slice(nil, opts) returns Candidates == 0, Included ==
// 0, Skipped == 0, Items non-nil and empty.
func TestSlice_EmptyPool(t *testing.T) {
	result := Slice(nil, Options{Budget: 2000})
	if result.Candidates != 0 {
		t.Errorf("Candidates = %d, want 0", result.Candidates)
	}
	if result.Included != 0 {
		t.Errorf("Included = %d, want 0", result.Included)
	}
	if result.Skipped != 0 {
		t.Errorf("Skipped = %d, want 0", result.Skipped)
	}
	if result.Items == nil {
		t.Error("Items is nil, want non-nil empty slice")
	}
	if len(result.Items) != 0 {
		t.Errorf("len(Items) = %d, want 0", len(result.Items))
	}
}

// TestPackageReadsNoWallClock is the mechanical guard behind DEC-043
// sub-decision 3: the ranking core reads no clock at all. Copied from
// internal/timewindow/timewindow_test.go's idiom.
func TestPackageReadsNoWallClock(t *testing.T) {
	err := filepath.WalkDir(".", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if strings.Contains(string(b), "time.Now(") {
			t.Errorf("%s calls time.Now(); the recency signal is ordinal, not clock-read (DEC-043 sub-decision 3)", path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// TestPackageEmitsNoReservedTagNamespace is the mechanical guard behind
// DEC-044's honesty clause: this package sizes a RETRIEVAL, and nothing it
// computes may ever be stamped onto an entry or enter the reserved tag
// namespace DEC-027 governs. Matching on the QUOTED prefix (rather than the
// bare token) is deliberate: it catches "this code is building a reserved tag"
// while staying immune to Go field names like `Tokens:` in a struct literal.
// Same idiom as internal/timewindow's TestPackageReadsNoWallClock.
func TestPackageEmitsNoReservedTagNamespace(t *testing.T) {
	forbidden := []string{`"agent:"`, `"model:"`, `"session:"`, `"cost:"`, `"tokens:"`}
	err := filepath.WalkDir(".", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, f := range forbidden {
			if strings.Contains(string(b), f) {
				t.Errorf("%s contains %s; this package never stamps a reserved tag (DEC-044 honesty clause)", path, f)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
