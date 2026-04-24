package export

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/jysf/bragfile000/internal/storage"
)

// Shared fixture used by the two golden tests and most others. 4 entries
// across 3 groups (alpha: 2, beta: 1, (no project): 1). Varied field-
// presence per SPEC-015's design notes. Timestamps chosen so chrono-ASC
// across all entries is 1 (T1), 3 (T3), 2 (T2), 4 (T4); within-alpha
// chrono-ASC is 1 (T1), 2 (T2).
var fixture = []storage.Entry{
	{
		ID: 1, Title: "alpha-old",
		Description: "old alpha", Tags: "auth",
		Project: "alpha", Type: "shipped", Impact: "did stuff",
		CreatedAt: time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC),
	},
	{
		ID: 2, Title: "alpha-new",
		Project: "alpha", Type: "shipped",
		CreatedAt: time.Date(2026, 4, 21, 10, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 4, 21, 10, 0, 0, 0, time.UTC),
	},
	{
		ID: 3, Title: "beta-only",
		Description: "beta desc",
		Project:     "beta", Type: "learned",
		CreatedAt: time.Date(2026, 4, 20, 11, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 4, 20, 11, 0, 0, 0, time.UTC),
	},
	{
		ID: 4, Title: "unbound",
		Type:      "shipped",
		CreatedAt: time.Date(2026, 4, 22, 10, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 4, 22, 10, 0, 0, 0, time.UTC),
	},
}

var fixedNow = time.Date(2026, 4, 23, 12, 0, 0, 0, time.UTC)

// TestToMarkdown_DEC013FullDocumentGolden locks all six DEC-013 choices
// against a literal grouped-mode document. If this fails, DEC-013 has
// been violated — fix the renderer, not the test. Load-bearing; write
// first.
func TestToMarkdown_DEC013FullDocumentGolden(t *testing.T) {
	opts := MarkdownOptions{Flat: false, Filters: "(none)", Now: fixedNow}

	want := `# Bragfile Export

Exported: 2026-04-23T12:00:00Z
Entries: 4
Filters: (none)

## Summary

**By type**
- shipped: 3
- learned: 1

**By project**
- alpha: 2
- beta: 1
- (no project): 1

## alpha

### alpha-old

| field       | value |
| ----------- | ----- |
| id          | 1 |
| created_at  | 2026-04-20T10:00:00Z |
| updated_at  | 2026-04-20T10:00:00Z |
| tags        | auth |
| project     | alpha |
| type        | shipped |
| impact      | did stuff |

#### Description

old alpha

---

### alpha-new

| field       | value |
| ----------- | ----- |
| id          | 2 |
| created_at  | 2026-04-21T10:00:00Z |
| updated_at  | 2026-04-21T10:00:00Z |
| project     | alpha |
| type        | shipped |

## beta

### beta-only

| field       | value |
| ----------- | ----- |
| id          | 3 |
| created_at  | 2026-04-20T11:00:00Z |
| updated_at  | 2026-04-20T11:00:00Z |
| project     | beta |
| type        | learned |

#### Description

beta desc

## (no project)

### unbound

| field       | value |
| ----------- | ----- |
| id          | 4 |
| created_at  | 2026-04-22T10:00:00Z |
| updated_at  | 2026-04-22T10:00:00Z |
| type        | shipped |`

	got, err := ToMarkdown(fixture, opts)
	if err != nil {
		t.Fatalf("ToMarkdown: %v", err)
	}
	if !bytes.Equal(got, []byte(want)) {
		t.Fatalf("DEC-013 shape drift:\nwant:\n%s\n\ngot:\n%s", want, got)
	}
	if n := len(got); n == 0 || got[n-1] != '|' {
		t.Fatalf("trailing byte: want %q, got %q (last byte)", "|", string(got[n-1]))
	}
}

// TestToMarkdown_FlatGolden locks DEC-013 choice 5 (--flat wrapper) via
// a full-document byte compare.
func TestToMarkdown_FlatGolden(t *testing.T) {
	opts := MarkdownOptions{Flat: true, Filters: "(none)", Now: fixedNow}

	want := `# Bragfile Export

Exported: 2026-04-23T12:00:00Z
Entries: 4
Filters: (none)

## Summary

**By type**
- shipped: 3
- learned: 1

**By project**
- alpha: 2
- beta: 1
- (no project): 1

## Entries (chronological)

### alpha-old

| field       | value |
| ----------- | ----- |
| id          | 1 |
| created_at  | 2026-04-20T10:00:00Z |
| updated_at  | 2026-04-20T10:00:00Z |
| tags        | auth |
| project     | alpha |
| type        | shipped |
| impact      | did stuff |

#### Description

old alpha

---

### beta-only

| field       | value |
| ----------- | ----- |
| id          | 3 |
| created_at  | 2026-04-20T11:00:00Z |
| updated_at  | 2026-04-20T11:00:00Z |
| project     | beta |
| type        | learned |

#### Description

beta desc

---

### alpha-new

| field       | value |
| ----------- | ----- |
| id          | 2 |
| created_at  | 2026-04-21T10:00:00Z |
| updated_at  | 2026-04-21T10:00:00Z |
| project     | alpha |
| type        | shipped |

---

### unbound

| field       | value |
| ----------- | ----- |
| id          | 4 |
| created_at  | 2026-04-22T10:00:00Z |
| updated_at  | 2026-04-22T10:00:00Z |
| type        | shipped |`

	got, err := ToMarkdown(fixture, opts)
	if err != nil {
		t.Fatalf("ToMarkdown: %v", err)
	}
	if !bytes.Equal(got, []byte(want)) {
		t.Fatalf("flat-mode shape drift:\nwant:\n%s\n\ngot:\n%s", want, got)
	}
	if n := len(got); n == 0 || got[n-1] != '|' {
		t.Fatalf("trailing byte: want %q, got %q (last byte)", "|", string(got[n-1]))
	}
}

// TestRenderEntry_HeadingLevel1 locks that the lift preserved byte
// output for `brag show` (headingLevel == 1). Full-fields entry emits
// # title, ## Description, and all optional metadata rows.
func TestRenderEntry_HeadingLevel1(t *testing.T) {
	var buf bytes.Buffer
	RenderEntry(&buf, fixture[0], 1)
	got := buf.String()

	for _, want := range []string{
		"# alpha-old\n\n",
		"## Description\n\n",
		"| id          | 1 |",
		"| created_at  | 2026-04-20T10:00:00Z |",
		"| updated_at  | 2026-04-20T10:00:00Z |",
		"| tags        | auth |",
		"| project     | alpha |",
		"| type        | shipped |",
		"| impact      | did stuff |",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("want substring %q in output, got:\n%s", want, got)
		}
	}
	for _, unwanted := range []string{"### alpha-old", "#### Description"} {
		if strings.Contains(got, unwanted) {
			t.Errorf("did not expect %q in level-1 output, got:\n%s", unwanted, got)
		}
	}

	// Pre-lift byte snapshot. This is the exact output the now-removed
	// internal/cli/show.go:renderEntry emitted for this entry. The lift
	// is behavior-preserving iff these bytes match.
	wantSnapshot := `# alpha-old

| field       | value |
| ----------- | ----- |
| id          | 1 |
| created_at  | 2026-04-20T10:00:00Z |
| updated_at  | 2026-04-20T10:00:00Z |
| tags        | auth |
| project     | alpha |
| type        | shipped |
| impact      | did stuff |

## Description

old alpha
`
	if got != wantSnapshot {
		t.Fatalf("lift broke byte output:\nwant:\n%s\n\ngot:\n%s", wantSnapshot, got)
	}
}

// TestRenderEntry_HeadingLevel3 locks the heading-level parameterization
// used by `brag export --format markdown` (level 3 title, level 4
// description). Metadata table is identical across heading levels.
func TestRenderEntry_HeadingLevel3(t *testing.T) {
	var buf bytes.Buffer
	RenderEntry(&buf, fixture[0], 3)
	got := buf.String()

	for _, want := range []string{
		"### alpha-old\n\n",
		"#### Description\n\n",
		"| id          | 1 |",
		"| tags        | auth |",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("want substring %q in output, got:\n%s", want, got)
		}
	}
	// Use line-based matching so that "### alpha-old" doesn't count as
	// a substring hit for "# alpha-old" / "## alpha-old", and "####
	// Description" doesn't register as a "## Description" hit.
	for _, unwanted := range []string{"# alpha-old", "## alpha-old", "## Description"} {
		for _, ln := range strings.Split(got, "\n") {
			if ln == unwanted {
				t.Errorf("did not expect line %q in level-3 output, got:\n%s", unwanted, got)
				break
			}
		}
	}
}

// TestRenderEntry_OmitsEmptyMetadataAndDescription locks the
// empty-field suppression rules. Pairs the claim that the lift (and
// the heading-level arg) don't change the optional-row rules.
func TestRenderEntry_OmitsEmptyMetadataAndDescription(t *testing.T) {
	var buf bytes.Buffer
	RenderEntry(&buf, fixture[3], 3) // unbound: title + type only
	got := buf.String()

	for _, want := range []string{
		"### unbound\n\n",
		"| id          | 4 |",
		"| created_at  | 2026-04-22T10:00:00Z |",
		"| updated_at  | 2026-04-22T10:00:00Z |",
		"| type        | shipped |",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("want substring %q in output, got:\n%s", want, got)
		}
	}
	for _, unwanted := range []string{"| tags", "| project", "| impact", "#### Description", "## Description"} {
		if strings.Contains(got, unwanted) {
			t.Errorf("did not expect %q when corresponding field is empty, got:\n%s", unwanted, got)
		}
	}
}

// TestToMarkdown_GroupingOrderRules walks the level-2 headings to prove
// alphabetical-ASC group ordering with (no project) forced last, then
// checks within-group chrono-ASC and separator placement.
func TestToMarkdown_GroupingOrderRules(t *testing.T) {
	opts := MarkdownOptions{Flat: false, Filters: "(none)", Now: fixedNow}
	got, err := ToMarkdown(fixture, opts)
	if err != nil {
		t.Fatalf("ToMarkdown: %v", err)
	}

	lines := strings.Split(string(got), "\n")
	var h2 []string
	for _, ln := range lines {
		if strings.HasPrefix(ln, "## ") && !strings.HasPrefix(ln, "### ") {
			h2 = append(h2, ln)
		}
	}
	wantH2 := []string{"## Summary", "## alpha", "## beta", "## (no project)"}
	if len(h2) != len(wantH2) {
		t.Fatalf("level-2 headings count: want %d, got %d (%v)", len(wantH2), len(h2), h2)
	}
	for i := range wantH2 {
		if h2[i] != wantH2[i] {
			t.Errorf("level-2 heading %d: want %q, got %q", i, wantH2[i], h2[i])
		}
	}

	// Within-alpha chrono-ASC.
	alphaStart := strings.Index(string(got), "## alpha\n")
	betaStart := strings.Index(string(got), "## beta\n")
	if alphaStart < 0 || betaStart < 0 {
		t.Fatalf("could not locate alpha/beta group markers")
	}
	alphaSection := string(got[alphaStart:betaStart])
	var h3 []string
	for _, ln := range strings.Split(alphaSection, "\n") {
		if strings.HasPrefix(ln, "### ") {
			h3 = append(h3, ln)
		}
	}
	wantH3 := []string{"### alpha-old", "### alpha-new"}
	if len(h3) != len(wantH3) {
		t.Fatalf("alpha group entry count: want %d, got %d (%v)", len(wantH3), len(h3), h3)
	}
	for i := range wantH3 {
		if h3[i] != wantH3[i] {
			t.Errorf("alpha group entry %d: want %q, got %q", i, wantH3[i], h3[i])
		}
	}

	// Exactly one `---` within alpha, zero at the alpha→beta boundary.
	sepsInAlpha := strings.Count(alphaSection, "\n---\n")
	if sepsInAlpha != 1 {
		t.Errorf("within-alpha separators: want 1, got %d in section:\n%s", sepsInAlpha, alphaSection)
	}
	// After the last alpha entry, before `## beta`, there must be NO `---`.
	alphaNewStart := strings.Index(alphaSection, "### alpha-new")
	if alphaNewStart < 0 {
		t.Fatalf("could not locate `### alpha-new` marker in alpha section")
	}
	if strings.Contains(alphaSection[alphaNewStart:], "\n---\n") {
		t.Errorf("unexpected `---` separator after last alpha entry (cross-group separator):\n%s", alphaSection[alphaNewStart:])
	}
}

// TestToMarkdown_SummaryCountsAndSorting locks the summary block
// ordering: DESC by count, alphabetical-ASC tiebreak, (no project)
// forced last in the by-project list regardless of count.
func TestToMarkdown_SummaryCountsAndSorting(t *testing.T) {
	opts := MarkdownOptions{Flat: false, Filters: "(none)", Now: fixedNow}
	got, err := ToMarkdown(fixture, opts)
	if err != nil {
		t.Fatalf("ToMarkdown: %v", err)
	}

	// By type block
	typeBlock := extractBetween(t, string(got), "**By type**\n", "\n\n**By project**")
	wantType := "- shipped: 3\n- learned: 1"
	if typeBlock != wantType {
		t.Errorf("by-type block:\nwant:\n%s\n\ngot:\n%s", wantType, typeBlock)
	}

	// By project block
	projBlock := extractBetween(t, string(got), "**By project**\n", "\n\n## alpha")
	wantProj := "- alpha: 2\n- beta: 1\n- (no project): 1"
	if projBlock != wantProj {
		t.Errorf("by-project block:\nwant:\n%s\n\ngot:\n%s", wantProj, projBlock)
	}

	// Tie-break fixture: two types with equal count.
	tieFixture := []storage.Entry{
		{ID: 1, Title: "a", Type: "shipped", CreatedAt: fixedNow, UpdatedAt: fixedNow},
		{ID: 2, Title: "b", Type: "shipped", CreatedAt: fixedNow, UpdatedAt: fixedNow},
		{ID: 3, Title: "c", Type: "learned", CreatedAt: fixedNow, UpdatedAt: fixedNow},
		{ID: 4, Title: "d", Type: "learned", CreatedAt: fixedNow, UpdatedAt: fixedNow},
		{ID: 5, Title: "e", Type: "fixed", CreatedAt: fixedNow, UpdatedAt: fixedNow},
	}
	out2, err := ToMarkdown(tieFixture, opts)
	if err != nil {
		t.Fatalf("ToMarkdown: %v", err)
	}
	typeBlock2 := extractBetween(t, string(out2), "**By type**\n", "\n\n**By project**")
	wantType2 := "- learned: 2\n- shipped: 2\n- fixed: 1"
	if typeBlock2 != wantType2 {
		t.Errorf("tie-break by-type block:\nwant:\n%s\n\ngot:\n%s", wantType2, typeBlock2)
	}

	// (no project) high-count fixture: 5 unbound + 3 beta + 2 alpha; even
	// with highest count, (no project) must render last.
	np := time.Date(2026, 4, 23, 0, 0, 0, 0, time.UTC)
	npFixture := []storage.Entry{
		{ID: 1, Title: "a1", Project: "alpha", CreatedAt: np, UpdatedAt: np},
		{ID: 2, Title: "a2", Project: "alpha", CreatedAt: np, UpdatedAt: np},
		{ID: 3, Title: "b1", Project: "beta", CreatedAt: np, UpdatedAt: np},
		{ID: 4, Title: "b2", Project: "beta", CreatedAt: np, UpdatedAt: np},
		{ID: 5, Title: "b3", Project: "beta", CreatedAt: np, UpdatedAt: np},
		{ID: 6, Title: "n1", CreatedAt: np, UpdatedAt: np},
		{ID: 7, Title: "n2", CreatedAt: np, UpdatedAt: np},
		{ID: 8, Title: "n3", CreatedAt: np, UpdatedAt: np},
		{ID: 9, Title: "n4", CreatedAt: np, UpdatedAt: np},
		{ID: 10, Title: "n5", CreatedAt: np, UpdatedAt: np},
	}
	out3, err := ToMarkdown(npFixture, opts)
	if err != nil {
		t.Fatalf("ToMarkdown: %v", err)
	}
	// The next section after (no project) summary is the first group:
	// by alphabetical-ASC that's "## alpha". Use "\n\n## alpha" as the
	// terminator.
	projBlock3 := extractBetween(t, string(out3), "**By project**\n", "\n\n## alpha")
	wantProj3 := "- beta: 3\n- alpha: 2\n- (no project): 5"
	if projBlock3 != wantProj3 {
		t.Errorf("no-project-last by-project block:\nwant:\n%s\n\ngot:\n%s", wantProj3, projBlock3)
	}
}

func extractBetween(t *testing.T, s, start, end string) string {
	t.Helper()
	i := strings.Index(s, start)
	if i < 0 {
		t.Fatalf("start marker %q not found in:\n%s", start, s)
	}
	rest := s[i+len(start):]
	j := strings.Index(rest, end)
	if j < 0 {
		t.Fatalf("end marker %q not found after %q in:\n%s", end, start, s)
	}
	return rest[:j]
}

// TestToMarkdown_FiltersLineFormat locks the passthrough of opts.Filters
// verbatim into the Filters: provenance line.
func TestToMarkdown_FiltersLineFormat(t *testing.T) {
	t.Run("none", func(t *testing.T) {
		got, err := ToMarkdown(fixture, MarkdownOptions{Flat: false, Filters: "(none)", Now: fixedNow})
		if err != nil {
			t.Fatalf("ToMarkdown: %v", err)
		}
		if !strings.Contains(string(got), "Filters: (none)\n") {
			t.Errorf("want line %q, got:\n%s", "Filters: (none)\n", got)
		}
	})
	t.Run("echoed_flags", func(t *testing.T) {
		got, err := ToMarkdown(fixture, MarkdownOptions{Flat: false, Filters: "--project platform --since 7d", Now: fixedNow})
		if err != nil {
			t.Fatalf("ToMarkdown: %v", err)
		}
		if !strings.Contains(string(got), "Filters: --project platform --since 7d\n") {
			t.Errorf("want line %q, got:\n%s", "Filters: --project platform --since 7d\n", got)
		}
	})
}

// TestToMarkdown_EmptyEntriesEmitsHeaderAndZeroCount locks the empty-
// input early-return: header + provenance only; no summary, no groups.
func TestToMarkdown_EmptyEntriesEmitsHeaderAndZeroCount(t *testing.T) {
	got, err := ToMarkdown([]storage.Entry{}, MarkdownOptions{Flat: false, Filters: "(none)", Now: fixedNow})
	if err != nil {
		t.Fatalf("ToMarkdown: %v", err)
	}

	want := `# Bragfile Export

Exported: 2026-04-23T12:00:00Z
Entries: 0
Filters: (none)`
	if !bytes.Equal(got, []byte(want)) {
		t.Fatalf("empty output mismatch:\nwant:\n%s\n\ngot:\n%s", want, got)
	}
	if strings.Contains(string(got), "## Summary") {
		t.Errorf("expected no summary block on empty input, got:\n%s", got)
	}
	// No `## <anything>` at all on empty input.
	if strings.Contains(string(got), "\n## ") {
		t.Errorf("expected no level-2 headings on empty input, got:\n%s", got)
	}
	if n := len(got); n == 0 || got[n-1] != ')' {
		t.Fatalf("trailing byte: want %q, got %q (last byte)", ")", string(got[n-1]))
	}
}
