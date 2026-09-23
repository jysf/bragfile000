package export

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jysf/bragfile000/internal/aggregate"
	"github.com/jysf/bragfile000/internal/storage"
)

// SummaryOptions controls the rule-based summary digest. Filters is
// the pre-formatted markdown line ("(none)" or an echoed flag string);
// FiltersJSON is the object the JSON envelope renders (empty map →
// "{}", populated → an object with alphabetically-sorted keys per
// Go's encoding/json map handling). Now is injected for deterministic
// goldens — mirrors MarkdownOptions.Now. DEC-014 locks the shape.
type SummaryOptions struct {
	Scope       string
	Filters     string
	FiltersJSON map[string]string
	Now         time.Time
}

// ToSummaryMarkdown renders entries as a rule-based digest per
// DEC-014. Returns bytes with the trailing "\n" stripped (matches
// ToJSON / ToMarkdown). On empty input, only the header + provenance
// block is emitted; the Summary and Highlights sections are omitted.
//
// The recorded failures (aggregate.SplitFailures, over EVERY in-window
// entry, because highlights list every entry) leave ## Highlights for
// ## What didn't work, in the same grouped shape (DEC-050 row 3). Each of
// the two sections is emitted only when it has an entry. The By type and
// By project counts still count both.
func ToSummaryMarkdown(entries []storage.Entry, opts SummaryOptions) ([]byte, error) {
	var buf bytes.Buffer
	fmt.Fprintln(&buf, "# Bragfile Summary")
	fmt.Fprintln(&buf)
	fmt.Fprintf(&buf, "Generated: %s\n", opts.Now.UTC().Format(time.RFC3339))
	fmt.Fprintf(&buf, "Scope: %s\n", opts.Scope)
	fmt.Fprintf(&buf, "Filters: %s\n", opts.Filters)

	if len(entries) == 0 {
		return trimTrailingNewline(buf.Bytes()), nil
	}

	fmt.Fprintln(&buf)
	fmt.Fprintln(&buf, "## Summary")
	fmt.Fprintln(&buf)
	fmt.Fprintln(&buf, "**By type**")
	for _, tc := range aggregate.ByType(entries) {
		fmt.Fprintf(&buf, "- %s: %d\n", tc.Type, tc.Count)
	}
	fmt.Fprintln(&buf)
	fmt.Fprintln(&buf, "**By project**")
	for _, pc := range aggregate.ByProject(entries) {
		fmt.Fprintf(&buf, "- %s: %d\n", pc.Project, pc.Count)
	}
	worked, failed := aggregate.SplitFailures(entries)
	if len(worked) > 0 {
		fmt.Fprintln(&buf)
		fmt.Fprintln(&buf, "## Highlights")
		writeHighlightGroups(&buf, worked)
	}
	if len(failed) > 0 {
		fmt.Fprintln(&buf)
		fmt.Fprintln(&buf, "## What didn't work")
		writeHighlightGroups(&buf, failed)
	}
	return trimTrailingNewline(buf.Bytes()), nil
}

// writeHighlightGroups renders entries grouped by project as `### <project>`
// blocks of `- <id>: <title>` lines. ## Highlights and ## What didn't work
// both render through it, so a failure reads in the same shape as the
// highlight it used to be.
func writeHighlightGroups(buf *bytes.Buffer, entries []storage.Entry) {
	for _, group := range aggregate.GroupForHighlights(entries) {
		fmt.Fprintln(buf)
		fmt.Fprintf(buf, "### %s\n", group.Project)
		fmt.Fprintln(buf)
		for _, ref := range group.Entries {
			fmt.Fprintf(buf, "- %d: %s\n", ref.ID, ref.Title)
		}
	}
}

// summaryEnvelope is the on-the-wire shape for ToSummaryJSON. Field
// order in this struct definition is the JSON key order DEC-014 locks
// (Go's encoding/json preserves struct-tag declaration order).
type summaryEnvelope struct {
	GeneratedAt     string            `json:"generated_at"`
	Scope           string            `json:"scope"`
	Filters         map[string]string `json:"filters"`
	CountsByType    map[string]int    `json:"counts_by_type"`
	CountsByProject map[string]int    `json:"counts_by_project"`
	Highlights      []highlightGroup  `json:"highlights"`
	// FailuresByProject is DEC-050 rule 5's key: the recorded failures,
	// grouped exactly as Highlights is. Always present, [] when empty.
	FailuresByProject []highlightGroup `json:"failures_by_project"`
}

type highlightGroup struct {
	Project string           `json:"project"`
	Entries []highlightEntry `json:"entries"`
}

type highlightEntry struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
}

// ToSummaryJSON renders the JSON envelope per DEC-014: single object,
// flat top-level keys (generated_at, scope, filters, counts_by_type,
// counts_by_project, highlights, failures_by_project), pretty-printed
// with 2-space indent. highlights carries the entries that are not
// recorded failures and failures_by_project the ones that are (DEC-050);
// both counts maps still count every in-window entry. Empty-state values
// per DEC-014 choice (4): counts maps render as {} and both arrays as
// [], never null.
func ToSummaryJSON(entries []storage.Entry, opts SummaryOptions) ([]byte, error) {
	env := summaryEnvelope{
		GeneratedAt:     opts.Now.UTC().Format(time.RFC3339),
		Scope:           opts.Scope,
		Filters:         opts.FiltersJSON,
		CountsByType:    map[string]int{},
		CountsByProject: map[string]int{},
	}
	if env.Filters == nil {
		env.Filters = map[string]string{}
	}
	for _, tc := range aggregate.ByType(entries) {
		env.CountsByType[tc.Type] = tc.Count
	}
	for _, pc := range aggregate.ByProject(entries) {
		env.CountsByProject[pc.Project] = pc.Count
	}
	worked, failed := aggregate.SplitFailures(entries)
	env.Highlights = highlightGroups(worked)
	env.FailuresByProject = highlightGroups(failed)
	return json.MarshalIndent(env, "", "  ")
}

// highlightGroups projects entries onto the {project, entries: [{id,
// title}]} group shape both summary arrays share. Non-nil for empty
// input, so either key renders as [] rather than null.
func highlightGroups(entries []storage.Entry) []highlightGroup {
	out := []highlightGroup{}
	for _, group := range aggregate.GroupForHighlights(entries) {
		hg := highlightGroup{
			Project: group.Project,
			Entries: make([]highlightEntry, 0, len(group.Entries)),
		}
		for _, ref := range group.Entries {
			hg.Entries = append(hg.Entries, highlightEntry{
				ID: ref.ID, Title: ref.Title,
			})
		}
		out = append(out, hg)
	}
	return out
}
