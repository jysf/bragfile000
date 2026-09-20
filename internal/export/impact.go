package export

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jysf/bragfile000/internal/aggregate"
	"github.com/jysf/bragfile000/internal/storage"
)

// ImpactOptions controls the rule-based impact digest (SPEC-048), the
// fourth DEC-014 consumer. Scope echoes the window token
// ("quarter"|"month"|"year"|"since:<raw>"). Filters is the
// pre-formatted markdown line ("(none)" or echoed flags); FiltersJSON
// is the object the JSON envelope renders (nil → {}). EntriesInWindow
// is the raw in-window count (the CLI does the windowing and passes it
// so the renderer can print the <shown>/<in-window> tally without
// re-deriving it). Now is injected for deterministic goldens.
// DEC-014 locks the envelope; DEC-028 locks the per-spec payload.
type ImpactOptions struct {
	Scope           string
	Filters         string
	FiltersJSON     map[string]string
	EntriesInWindow int
	Now             time.Time
}

// ToImpactMarkdown renders the in-window entries as an impact-first
// digest per DEC-014/DEC-028/DEC-050. The renderer receives the already-in-
// window slice; it selects the with-impact subset (aggregate.WithImpact),
// splits the recorded failures out of it (aggregate.SplitFailures), and
// renders each half grouped by project with its impact text in full: the
// rest under ## Impact, the failures under ## What didn't work. Each section
// is emitted only when it has an entry (DEC-050), and the Entries: tally
// still counts both — it is the with-impact subset the body shows. Returns
// bytes with the trailing "\n" stripped (matches ToSummaryMarkdown). On zero
// with-impact entries, only the header + provenance block is emitted.
func ToImpactMarkdown(entries []storage.Entry, opts ImpactOptions) ([]byte, error) {
	withImpact := aggregate.WithImpact(entries)
	worked, failed := aggregate.SplitFailures(withImpact)

	var buf bytes.Buffer
	fmt.Fprintln(&buf, "# Bragfile Impact")
	fmt.Fprintln(&buf)
	fmt.Fprintf(&buf, "Generated: %s\n", opts.Now.UTC().Format(time.RFC3339))
	fmt.Fprintf(&buf, "Scope: %s\n", opts.Scope)
	fmt.Fprintf(&buf, "Filters: %s\n", opts.Filters)
	fmt.Fprintf(&buf, "Entries: %d/%d with impact\n", len(withImpact), opts.EntriesInWindow)

	if len(worked) > 0 {
		fmt.Fprintln(&buf)
		fmt.Fprintln(&buf, "## Impact")
		writeImpactGroups(&buf, worked)
	}
	if len(failed) > 0 {
		fmt.Fprintln(&buf)
		fmt.Fprintln(&buf, "## What didn't work")
		writeImpactGroups(&buf, failed)
	}
	return trimTrailingNewline(buf.Bytes()), nil
}

// writeImpactGroups renders entries grouped by project as `### <project>`
// blocks of `- <id>: <title>` plus an indented `  <impact>` line — the
// per-entry shape DEC-028 choice 4 locks. brag impact and brag wrapped both
// render their two impact-bearing sections through it, so an entry reads
// byte-identically on either surface.
func writeImpactGroups(buf *bytes.Buffer, entries []storage.Entry) {
	for _, group := range aggregate.GroupEntriesByProject(entries) {
		fmt.Fprintln(buf)
		fmt.Fprintf(buf, "### %s\n", group.Project)
		fmt.Fprintln(buf)
		for _, e := range group.Entries {
			fmt.Fprintf(buf, "- %d: %s\n", e.ID, e.Title)
			fmt.Fprintf(buf, "  %s\n", e.Impact)
		}
	}
}

// impactEnvelope is the on-the-wire shape for ToImpactJSON. Field order
// is the JSON key order DEC-014/DEC-028 lock (encoding/json preserves
// struct-tag declaration order).
type impactEnvelope struct {
	GeneratedAt       string               `json:"generated_at"`
	Scope             string               `json:"scope"`
	Filters           map[string]string    `json:"filters"`
	EntriesInWindow   int                  `json:"entries_in_window"`
	EntriesWithImpact int                  `json:"entries_with_impact"`
	CountsByProject   map[string]int       `json:"counts_by_project"`
	ImpactByProject   []impactProjectGroup `json:"impact_by_project"`
	FailuresByProject []impactProjectGroup `json:"failures_by_project"`
}

type impactProjectGroup struct {
	Project string        `json:"project"`
	Entries []impactEntry `json:"entries"`
}

// impactEntry is the deliberately NARROW 4-key projection (DEC-028
// choice 4) — not DEC-011's 9-key shape. The narrative pipe
// (STAGE-012) needs only enough to attribute an impact statement to an
// entry and its initiative.
type impactEntry struct {
	ID      int64  `json:"id"`
	Title   string `json:"title"`
	Project string `json:"project"`
	Impact  string `json:"impact"`
}

// ToImpactJSON renders the DEC-014 envelope with DEC-028's per-spec
// payload keys plus DEC-050's: generated_at, scope, filters,
// entries_in_window, entries_with_impact, counts_by_project (map over the
// whole with-impact subset — failures included, so it still sums to
// entries_with_impact), impact_by_project (the with-impact entries that are
// not failures) and failures_by_project (the ones that are), each an array of
// grouped 4-key projections. 2-space indent. Empty-state per DEC-014 choice
// (4): counts {}, both arrays [], filters {}, never null.
func ToImpactJSON(entries []storage.Entry, opts ImpactOptions) ([]byte, error) {
	withImpact := aggregate.WithImpact(entries)
	worked, failed := aggregate.SplitFailures(withImpact)

	env := impactEnvelope{
		GeneratedAt:       opts.Now.UTC().Format(time.RFC3339),
		Scope:             opts.Scope,
		Filters:           opts.FiltersJSON,
		EntriesInWindow:   opts.EntriesInWindow,
		EntriesWithImpact: len(withImpact),
		CountsByProject:   map[string]int{},
		ImpactByProject:   impactGroups(worked),
		FailuresByProject: impactGroups(failed),
	}
	if env.Filters == nil {
		env.Filters = map[string]string{}
	}
	// Counted over withImpact, not over worked: DEC-028 defines this map over
	// the with-impact subset, and narrowing it to one section would change
	// what an existing key counts without renaming it (DEC-048).
	for _, group := range aggregate.GroupEntriesByProject(withImpact) {
		env.CountsByProject[group.Project] = len(group.Entries)
	}
	return json.MarshalIndent(env, "", "  ")
}

// impactGroups projects entries into project groups of the NARROW 4-key
// entry shape. Non-nil on empty input, so an empty section renders [].
func impactGroups(entries []storage.Entry) []impactProjectGroup {
	out := make([]impactProjectGroup, 0)
	for _, group := range aggregate.GroupEntriesByProject(entries) {
		g := impactProjectGroup{
			Project: group.Project,
			Entries: make([]impactEntry, 0, len(group.Entries)),
		}
		for _, e := range group.Entries {
			g.Entries = append(g.Entries, impactEntry{
				ID:      e.ID,
				Title:   e.Title,
				Project: group.Project,
				Impact:  e.Impact,
			})
		}
		out = append(out, g)
	}
	return out
}
