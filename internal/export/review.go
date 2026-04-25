package export

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jysf/bragfile000/internal/aggregate"
	"github.com/jysf/bragfile000/internal/storage"
)

// reflectionQuestions are the three hard-coded prompts SPEC-019 locks.
// Wording verbatim from STAGE-004 Design Notes lines 367–369.
// Configurability is backlogged with revisit trigger "user wants to
// swap one out."
var reflectionQuestions = []string{
	"What pattern do you see in this period?",
	"What did you underestimate?",
	"What's missing here that should be?",
}

// ReviewOptions controls ToReviewMarkdown / ToReviewJSON. Scope is
// "week" or "month" (echoed into provenance + envelope). Now is
// injected for deterministic Generated: lines (mirrors
// MarkdownOptions.Now + SummaryOptions.Now). No Filters field —
// review does not accept filter flags; "(none)" is hard-coded in the
// markdown provenance line and {} is hard-coded in the JSON envelope.
type ReviewOptions struct {
	Scope string
	Now   time.Time
}

// ToReviewMarkdown renders the DEC-014 markdown digest for brag
// review. Returns bytes with trailing "\n" stripped (matches the byte
// contract of the other renderers). The Reflection questions block
// ALWAYS renders, even on empty entries — the questions are the point
// of the command.
func ToReviewMarkdown(entries []storage.Entry, opts ReviewOptions) ([]byte, error) {
	var buf bytes.Buffer
	fmt.Fprintln(&buf, "# Bragfile Review")
	fmt.Fprintln(&buf)
	fmt.Fprintf(&buf, "Generated: %s\n", opts.Now.UTC().Format(time.RFC3339))
	fmt.Fprintf(&buf, "Scope: %s\n", opts.Scope)
	fmt.Fprintln(&buf, "Filters: (none)")

	if len(entries) > 0 {
		fmt.Fprintln(&buf)
		fmt.Fprintln(&buf, "## Entries")
		for _, group := range aggregate.GroupEntriesByProject(entries) {
			fmt.Fprintln(&buf)
			fmt.Fprintf(&buf, "### %s\n", group.Project)
			fmt.Fprintln(&buf)
			for _, e := range group.Entries {
				fmt.Fprintf(&buf, "- %d: %s\n", e.ID, e.Title)
			}
		}
	}

	fmt.Fprintln(&buf)
	fmt.Fprintln(&buf, "## Reflection questions")
	fmt.Fprintln(&buf)
	for i, q := range reflectionQuestions {
		fmt.Fprintf(&buf, "%d. %s\n", i+1, q)
	}
	return trimTrailingNewline(buf.Bytes()), nil
}

// reviewEnvelope is the on-the-wire JSON shape for ToReviewJSON. Field
// order locks DEC-014's top-level key order (generated_at, scope,
// filters, entries_grouped, reflection_questions) via struct-tag
// declaration order.
type reviewEnvelope struct {
	GeneratedAt         string               `json:"generated_at"`
	Scope               string               `json:"scope"`
	Filters             map[string]string    `json:"filters"`
	EntriesGrouped      []reviewProjectGroup `json:"entries_grouped"`
	ReflectionQuestions []string             `json:"reflection_questions"`
}

type reviewProjectGroup struct {
	Project string        `json:"project"`
	Entries []entryRecord `json:"entries"`
}

// ToReviewJSON renders the DEC-014 envelope for brag review. Per-entry
// shape inside entries_grouped[].entries is the DEC-011 9-key shape
// (via the toEntryRecord helper).
func ToReviewJSON(entries []storage.Entry, opts ReviewOptions) ([]byte, error) {
	env := reviewEnvelope{
		GeneratedAt:         opts.Now.UTC().Format(time.RFC3339),
		Scope:               opts.Scope,
		Filters:             map[string]string{},
		EntriesGrouped:      []reviewProjectGroup{},
		ReflectionQuestions: append([]string{}, reflectionQuestions...),
	}
	for _, group := range aggregate.GroupEntriesByProject(entries) {
		rg := reviewProjectGroup{
			Project: group.Project,
			Entries: make([]entryRecord, 0, len(group.Entries)),
		}
		for _, e := range group.Entries {
			rg.Entries = append(rg.Entries, toEntryRecord(e))
		}
		env.EntriesGrouped = append(env.EntriesGrouped, rg)
	}
	return json.MarshalIndent(env, "", "  ")
}
