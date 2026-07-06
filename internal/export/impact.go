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
// digest per DEC-014/DEC-028. The renderer receives the already-in-
// window slice; it selects the with-impact subset (aggregate.WithImpact),
// groups it by project (aggregate.GroupEntriesByProject), and renders
// each shown entry's impact text in full. Returns bytes with the
// trailing "\n" stripped (matches ToSummaryMarkdown). On zero with-
// impact entries, only the header + provenance block is emitted; the
// ## Impact body is omitted.
func ToImpactMarkdown(entries []storage.Entry, opts ImpactOptions) ([]byte, error) {
	withImpact := aggregate.WithImpact(entries)

	var buf bytes.Buffer
	fmt.Fprintln(&buf, "# Bragfile Impact")
	fmt.Fprintln(&buf)
	fmt.Fprintf(&buf, "Generated: %s\n", opts.Now.UTC().Format(time.RFC3339))
	fmt.Fprintf(&buf, "Scope: %s\n", opts.Scope)
	fmt.Fprintf(&buf, "Filters: %s\n", opts.Filters)
	fmt.Fprintf(&buf, "Entries: %d/%d with impact\n", len(withImpact), opts.EntriesInWindow)

	if len(withImpact) == 0 {
		return trimTrailingNewline(buf.Bytes()), nil
	}

	fmt.Fprintln(&buf)
	fmt.Fprintln(&buf, "## Impact")
	for _, group := range aggregate.GroupEntriesByProject(withImpact) {
		fmt.Fprintln(&buf)
		fmt.Fprintf(&buf, "### %s\n", group.Project)
		fmt.Fprintln(&buf)
		for _, e := range group.Entries {
			fmt.Fprintf(&buf, "- %d: %s\n", e.ID, e.Title)
			fmt.Fprintf(&buf, "  %s\n", e.Impact)
		}
	}
	return trimTrailingNewline(buf.Bytes()), nil
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
// payload keys: generated_at, scope, filters, entries_in_window,
// entries_with_impact, counts_by_project (map over the with-impact
// subset), impact_by_project (array of grouped 4-key projections).
// 2-space indent. Empty-state per DEC-014 choice (4): counts {},
// impact_by_project [], filters {}, never null.
func ToImpactJSON(entries []storage.Entry, opts ImpactOptions) ([]byte, error) {
	withImpact := aggregate.WithImpact(entries)

	env := impactEnvelope{
		GeneratedAt:       opts.Now.UTC().Format(time.RFC3339),
		Scope:             opts.Scope,
		Filters:           opts.FiltersJSON,
		EntriesInWindow:   opts.EntriesInWindow,
		EntriesWithImpact: len(withImpact),
		CountsByProject:   map[string]int{},
		ImpactByProject:   []impactProjectGroup{},
	}
	if env.Filters == nil {
		env.Filters = map[string]string{}
	}

	for _, group := range aggregate.GroupEntriesByProject(withImpact) {
		env.CountsByProject[group.Project] = len(group.Entries)
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
		env.ImpactByProject = append(env.ImpactByProject, g)
	}
	return json.MarshalIndent(env, "", "  ")
}
