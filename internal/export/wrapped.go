package export

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jysf/bragfile000/internal/aggregate"
	"github.com/jysf/bragfile000/internal/storage"
)

// WrappedOptions controls the rule-based wrapped digest (SPEC-051), the
// fifth DEC-014 consumer. Scope echoes the named-period token ("2026" or
// "2026-Q3"). Filters is the pre-formatted markdown line ("(none)" or
// echoed flags); FiltersJSON is the object the JSON envelope renders
// (nil → {}). ScopeMonths is the ordered set of "YYYY-MM" labels in
// scope (12 for a year, 3 for a quarter) — the CLI derives it from the
// period so the cadence series is always fully present, even on an empty
// period. Now is injected for a deterministic Generated: line.
//
// The renderer receives the ALREADY-in-period slice (the CLI does the
// bounded-window filtering, DEC-030 choice 3) plus ScopeMonths. Now here
// only feeds the Generated: line: wrapped surfaces LONGEST streak, which
// is period-scoped and independent of now (DEC-022 / DEC-030 note), so
// the streak number never couples to the clock seam.
type WrappedOptions struct {
	Scope       string
	Filters     string
	FiltersJSON map[string]string
	ScopeMonths []string
	Now         time.Time
}

// ToWrappedMarkdown renders the in-period entries as the celebratory
// wrapped digest per DEC-014/DEC-030: provenance, then the section arc
// Cadence → Top initiatives → Impact moments → Rhythm → Span. Returns
// bytes with the trailing "\n" stripped (matches every other renderer).
// On an empty period only the header + provenance block (through
// "Entries: 0") is emitted; the body sections are omitted (DEC-014 part
// 4).
func ToWrappedMarkdown(entries []storage.Entry, opts WrappedOptions) ([]byte, error) {
	var buf bytes.Buffer
	fmt.Fprintln(&buf, "# Bragfile Wrapped")
	fmt.Fprintln(&buf)
	fmt.Fprintf(&buf, "Generated: %s\n", opts.Now.UTC().Format(time.RFC3339))
	fmt.Fprintf(&buf, "Scope: %s\n", opts.Scope)
	fmt.Fprintf(&buf, "Filters: %s\n", opts.Filters)
	fmt.Fprintf(&buf, "Entries: %d\n", len(entries))

	if len(entries) == 0 {
		return trimTrailingNewline(buf.Bytes()), nil
	}

	// Cadence.
	series, busiest := aggregate.Cadence(entries, opts.ScopeMonths)
	fmt.Fprintln(&buf)
	fmt.Fprintln(&buf, "## Cadence")
	fmt.Fprintln(&buf)
	for _, b := range series {
		if b.Period == busiest {
			fmt.Fprintf(&buf, "Busiest month: %s (%d)\n", b.Period, b.Count)
			break
		}
	}
	fmt.Fprintln(&buf)
	for _, b := range series {
		fmt.Fprintf(&buf, "- %s: %d\n", b.Period, b.Count)
	}

	// Top initiatives (top-5 projects by count, excluding (no project)).
	initiatives := aggregate.MostCommon(extractProjects(entries), 5)
	fmt.Fprintln(&buf)
	fmt.Fprintln(&buf, "## Top initiatives")
	fmt.Fprintln(&buf)
	for _, nc := range initiatives {
		fmt.Fprintf(&buf, "- %s: %d\n", nc.Name, nc.Count)
	}

	// Impact moments (with-impact entries grouped by project, full text).
	fmt.Fprintln(&buf)
	fmt.Fprintln(&buf, "## Impact moments")
	for _, group := range aggregate.GroupEntriesByProject(aggregate.WithImpact(entries)) {
		fmt.Fprintln(&buf)
		fmt.Fprintf(&buf, "### %s\n", group.Project)
		fmt.Fprintln(&buf)
		for _, e := range group.Entries {
			fmt.Fprintf(&buf, "- %d: %s\n", e.ID, e.Title)
			fmt.Fprintf(&buf, "  %s\n", e.Impact)
		}
	}

	// Rhythm (longest streak, top-5 tags, top-3 types).
	_, longest := aggregate.Streak(entries, opts.Now)
	tags := aggregate.MostCommon(extractTags(entries), 5)
	types := aggregate.MostCommon(extractTypes(entries), 3)
	fmt.Fprintln(&buf)
	fmt.Fprintln(&buf, "## Rhythm")
	fmt.Fprintln(&buf)
	fmt.Fprintf(&buf, "Longest streak: %d days\n", longest)
	fmt.Fprintln(&buf)
	fmt.Fprintln(&buf, "**Top tags**")
	for _, nc := range tags {
		fmt.Fprintf(&buf, "- %s: %d\n", nc.Name, nc.Count)
	}
	fmt.Fprintln(&buf)
	fmt.Fprintln(&buf, "**Top types**")
	for _, nc := range types {
		fmt.Fprintf(&buf, "- %s: %d\n", nc.Name, nc.Count)
	}

	// Span (first/last entry date + active days).
	span := aggregate.Span(entries)
	fmt.Fprintln(&buf)
	fmt.Fprintln(&buf, "## Span")
	fmt.Fprintln(&buf)
	fmt.Fprintf(&buf, "- First entry: %s\n", span.First.UTC().Format("2006-01-02"))
	fmt.Fprintf(&buf, "- Last entry: %s\n", span.Last.UTC().Format("2006-01-02"))
	fmt.Fprintf(&buf, "- Active days: %d\n", span.Days)

	return trimTrailingNewline(buf.Bytes()), nil
}

// wrappedEnvelope is the on-the-wire shape for ToWrappedJSON. Struct-tag
// declaration order is the JSON key order DEC-014/DEC-030 lock
// (encoding/json preserves it).
type wrappedEnvelope struct {
	GeneratedAt    string               `json:"generated_at"`
	Scope          string               `json:"scope"`
	Filters        map[string]string    `json:"filters"`
	TotalEntries   int                  `json:"total_entries"`
	Cadence        cadenceRecord        `json:"cadence"`
	TopInitiatives []wrappedInitiative  `json:"top_initiatives"`
	ImpactMoments  []wrappedImpactGroup `json:"impact_moments"`
	LongestStreak  int                  `json:"longest_streak"`
	TopTags        []wrappedNameCount   `json:"top_tags"`
	TopTypes       []wrappedNameCount   `json:"top_types"`
	Span           wrappedSpanRecord    `json:"span"`
}

// cadenceRecord uses *string for BusiestMonth so an empty period renders
// null; the series is always fully present (zero-filled) so SPEC-052
// renders a flat sparkline, not a gap.
type cadenceRecord struct {
	BusiestMonth *string                   `json:"busiest_month"`
	Series       []aggregate.CadenceBucket `json:"series"`
}

type wrappedInitiative struct {
	Project string `json:"project"`
	Count   int    `json:"count"`
}

type wrappedImpactGroup struct {
	Project string         `json:"project"`
	Entries []wrappedEntry `json:"entries"`
}

// wrappedEntry is the same NARROW 4-key projection impact uses — enough
// to attribute an impact statement to an entry and its initiative.
type wrappedEntry struct {
	ID      int64  `json:"id"`
	Title   string `json:"title"`
	Project string `json:"project"`
	Impact  string `json:"impact"`
}

type wrappedNameCount struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// wrappedSpanRecord uses *string for the two date fields so an empty
// period renders null; active_days stays int (0 on empty is valid).
type wrappedSpanRecord struct {
	FirstEntryDate *string `json:"first_entry_date"`
	LastEntryDate  *string `json:"last_entry_date"`
	ActiveDays     int     `json:"active_days"`
}

// ToWrappedJSON renders the DEC-014 envelope with DEC-030's per-spec
// payload keys. Every key is always emitted; on an empty period arrays
// are [], objects {}, busiest_month/date fields null, numbers 0 (DEC-014
// part 4) — but cadence.series is still the full zero-filled month
// series so the sparkline slot is present even when empty.
func ToWrappedJSON(entries []storage.Entry, opts WrappedOptions) ([]byte, error) {
	series, busiest := aggregate.Cadence(entries, opts.ScopeMonths)

	env := wrappedEnvelope{
		GeneratedAt:    opts.Now.UTC().Format(time.RFC3339),
		Scope:          opts.Scope,
		Filters:        opts.FiltersJSON,
		TotalEntries:   len(entries),
		Cadence:        cadenceRecord{Series: series},
		TopInitiatives: []wrappedInitiative{},
		ImpactMoments:  []wrappedImpactGroup{},
		TopTags:        []wrappedNameCount{},
		TopTypes:       []wrappedNameCount{},
		Span:           wrappedSpanRecord{},
	}
	if env.Filters == nil {
		env.Filters = map[string]string{}
	}
	if busiest != "" {
		b := busiest
		env.Cadence.BusiestMonth = &b
	}

	if len(entries) > 0 {
		for _, nc := range aggregate.MostCommon(extractProjects(entries), 5) {
			env.TopInitiatives = append(env.TopInitiatives, wrappedInitiative{Project: nc.Name, Count: nc.Count})
		}
		for _, group := range aggregate.GroupEntriesByProject(aggregate.WithImpact(entries)) {
			g := wrappedImpactGroup{
				Project: group.Project,
				Entries: make([]wrappedEntry, 0, len(group.Entries)),
			}
			for _, e := range group.Entries {
				g.Entries = append(g.Entries, wrappedEntry{
					ID:      e.ID,
					Title:   e.Title,
					Project: group.Project,
					Impact:  e.Impact,
				})
			}
			env.ImpactMoments = append(env.ImpactMoments, g)
		}
		_, longest := aggregate.Streak(entries, opts.Now)
		env.LongestStreak = longest
		for _, nc := range aggregate.MostCommon(extractTags(entries), 5) {
			env.TopTags = append(env.TopTags, wrappedNameCount{Name: nc.Name, Count: nc.Count})
		}
		for _, nc := range aggregate.MostCommon(extractTypes(entries), 3) {
			env.TopTypes = append(env.TopTypes, wrappedNameCount{Name: nc.Name, Count: nc.Count})
		}
		span := aggregate.Span(entries)
		first := span.First.UTC().Format("2006-01-02")
		last := span.Last.UTC().Format("2006-01-02")
		env.Span = wrappedSpanRecord{
			FirstEntryDate: &first,
			LastEntryDate:  &last,
			ActiveDays:     span.Days,
		}
	}

	return json.MarshalIndent(env, "", "  ")
}

// extractTypes returns each entry's non-empty Type field, suitable for
// aggregate.MostCommon. Mirrors extractProjects (stats.go): empty Type
// is excluded from counting.
func extractTypes(entries []storage.Entry) []string {
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.Type == "" {
			continue
		}
		out = append(out, e.Type)
	}
	return out
}
