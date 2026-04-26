package export

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/jysf/bragfile000/internal/aggregate"
	"github.com/jysf/bragfile000/internal/storage"
)

// StatsOptions controls the rule-based stats digest. Now is injected
// for deterministic Generated: lines AND for the streak today-reference
// (single source — the renderer passes opts.Now straight through to
// aggregate.Streak). NO Scope field — stats always renders
// "Scope: lifetime". NO Filters / FiltersJSON fields — stats accepts
// no filter flags; "Filters: (none)" / "filters": {} are hard-coded.
// SPEC-020.
type StatsOptions struct {
	Now time.Time
}

// ToStatsMarkdown renders stats as the DEC-014 markdown digest. Returns
// bytes with trailing "\n" stripped (matches the byte contract of the
// other renderers). Empty entries → header + provenance only (no
// ## Stats wrapper, no metric body) per DEC-014 part (4).
func ToStatsMarkdown(entries []storage.Entry, opts StatsOptions) ([]byte, error) {
	var buf bytes.Buffer
	fmt.Fprintln(&buf, "# Bragfile Stats")
	fmt.Fprintln(&buf)
	fmt.Fprintf(&buf, "Generated: %s\n", opts.Now.UTC().Format(time.RFC3339))
	fmt.Fprintln(&buf, "Scope: lifetime")
	fmt.Fprintln(&buf, "Filters: (none)")

	if len(entries) == 0 {
		return trimTrailingNewline(buf.Bytes()), nil
	}

	span := aggregate.Span(entries)
	current, longest := aggregate.Streak(entries, opts.Now)
	tags := aggregate.MostCommon(extractTags(entries), 5)
	projects := aggregate.MostCommon(extractProjects(entries), 5)
	epw := computeEntriesPerWeek(len(entries), span.Days)

	fmt.Fprintln(&buf)
	fmt.Fprintln(&buf, "## Stats")
	fmt.Fprintln(&buf)
	fmt.Fprintln(&buf, "**Activity**")
	fmt.Fprintf(&buf, "- Total entries: %d\n", len(entries))
	fmt.Fprintf(&buf, "- Entries/week: %.2f\n", epw)
	fmt.Fprintln(&buf)
	fmt.Fprintln(&buf, "**Streaks**")
	fmt.Fprintf(&buf, "- Current: %d days\n", current)
	fmt.Fprintf(&buf, "- Longest: %d days\n", longest)
	fmt.Fprintln(&buf)
	fmt.Fprintln(&buf, "**Top tags**")
	for _, nc := range tags {
		fmt.Fprintf(&buf, "- %s: %d\n", nc.Name, nc.Count)
	}
	fmt.Fprintln(&buf)
	fmt.Fprintln(&buf, "**Top projects**")
	for _, nc := range projects {
		fmt.Fprintf(&buf, "- %s: %d\n", nc.Name, nc.Count)
	}
	fmt.Fprintln(&buf)
	fmt.Fprintln(&buf, "**Corpus span**")
	fmt.Fprintf(&buf, "- First entry: %s\n", span.First.UTC().Format("2006-01-02"))
	fmt.Fprintf(&buf, "- Last entry: %s\n", span.Last.UTC().Format("2006-01-02"))
	fmt.Fprintf(&buf, "- Days: %d\n", span.Days)

	return trimTrailingNewline(buf.Bytes()), nil
}

// statsEnvelope is the on-the-wire shape for ToStatsJSON. Field order
// in this struct definition is the JSON key order DEC-014 locks (Go's
// encoding/json preserves struct-tag declaration order).
type statsEnvelope struct {
	GeneratedAt    string            `json:"generated_at"`
	Scope          string            `json:"scope"`
	Filters        map[string]string `json:"filters"`
	TotalCount     int               `json:"total_count"`
	EntriesPerWeek float64           `json:"entries_per_week"`
	CurrentStreak  int               `json:"current_streak"`
	LongestStreak  int               `json:"longest_streak"`
	TopTags        []tagCount        `json:"top_tags"`
	TopProjects    []projCount       `json:"top_projects"`
	CorpusSpan     corpusSpanRecord  `json:"corpus_span"`
}

type tagCount struct {
	Tag   string `json:"tag"`
	Count int    `json:"count"`
}

type projCount struct {
	Project string `json:"project"`
	Count   int    `json:"count"`
}

// corpusSpanRecord uses *string for the date fields so JSON marshals
// empty corpus as null and non-empty as "YYYY-MM-DD". Days stays int —
// 0 on empty is the valid value.
type corpusSpanRecord struct {
	FirstEntryDate *string `json:"first_entry_date"`
	LastEntryDate  *string `json:"last_entry_date"`
	Days           int     `json:"days"`
}

// ToStatsJSON renders the JSON envelope per DEC-014 with the SPEC-020
// per-spec payload keys at top level.
func ToStatsJSON(entries []storage.Entry, opts StatsOptions) ([]byte, error) {
	env := statsEnvelope{
		GeneratedAt: opts.Now.UTC().Format(time.RFC3339),
		Scope:       "lifetime",
		Filters:     map[string]string{},
		TotalCount:  len(entries),
		TopTags:     []tagCount{},
		TopProjects: []projCount{},
		CorpusSpan:  corpusSpanRecord{},
	}

	if len(entries) > 0 {
		span := aggregate.Span(entries)
		current, longest := aggregate.Streak(entries, opts.Now)
		env.EntriesPerWeek = computeEntriesPerWeek(len(entries), span.Days)
		env.CurrentStreak = current
		env.LongestStreak = longest

		for _, nc := range aggregate.MostCommon(extractTags(entries), 5) {
			env.TopTags = append(env.TopTags, tagCount{Tag: nc.Name, Count: nc.Count})
		}
		for _, nc := range aggregate.MostCommon(extractProjects(entries), 5) {
			env.TopProjects = append(env.TopProjects, projCount{Project: nc.Name, Count: nc.Count})
		}

		first := span.First.UTC().Format("2006-01-02")
		last := span.Last.UTC().Format("2006-01-02")
		env.CorpusSpan = corpusSpanRecord{
			FirstEntryDate: &first,
			LastEntryDate:  &last,
			Days:           span.Days,
		}
	}

	return json.MarshalIndent(env, "", "  ")
}

// computeEntriesPerWeek implements SPEC-020 locked decision §4: decimal-
// weeks formula; sub-1-week corpus → 0.0; otherwise 2-decimal round.
func computeEntriesPerWeek(total, spanDays int) float64 {
	weeks := float64(spanDays) / 7.0
	if weeks < 1.0 {
		return 0.0
	}
	return math.Round((float64(total)/weeks)*100) / 100
}

// extractTags splits each entry's Tags on "," (per DEC-004), trims each
// token, and drops empties. Returns a flat []string suitable for
// aggregate.MostCommon.
func extractTags(entries []storage.Entry) []string {
	out := make([]string, 0)
	for _, e := range entries {
		if e.Tags == "" {
			continue
		}
		for _, raw := range strings.Split(e.Tags, ",") {
			tag := strings.TrimSpace(raw)
			if tag == "" {
				continue
			}
			out = append(out, tag)
		}
	}
	return out
}

// extractProjects returns each entry's non-empty Project field. Empty
// Project is excluded from counting (locked decision §3 — (no project)
// excluded from top_projects).
func extractProjects(entries []storage.Entry) []string {
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.Project == "" {
			continue
		}
		out = append(out, e.Project)
	}
	return out
}
