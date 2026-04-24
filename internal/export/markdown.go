package export

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/jysf/bragfile000/internal/storage"
)

// MarkdownOptions controls ToMarkdown's output. Flat switches from
// grouped-by-project to a single chronological section. Filters is the
// pre-formatted value the CLI layer assembled from its flag state
// (e.g. "(none)" or "--project platform --since 7d"). Now is the
// timestamp rendered into the "Exported:" provenance line — injected
// (not time.Now()-called internally) so tests produce deterministic
// goldens.
type MarkdownOptions struct {
	Flat    bool
	Filters string
	Now     time.Time
}

// RenderEntry writes e as a markdown block to w. The title appears at
// headingLevel (e.g., 1 → "# ", 3 → "### "). The "Description"
// sub-heading appears one level below (headingLevel + 1).
//
// Optional metadata rows (tags, project, type, impact) are suppressed
// when empty; an entry with no description omits the description
// heading entirely. Lifted from internal/cli/show.go in SPEC-015 —
// pre-lift behavior for `brag show` is preserved exactly at
// headingLevel == 1.
func RenderEntry(w io.Writer, e storage.Entry, headingLevel int) {
	titlePrefix := strings.Repeat("#", headingLevel) + " "
	descPrefix := strings.Repeat("#", headingLevel+1) + " "

	fmt.Fprintf(w, "%s%s\n\n", titlePrefix, e.Title)
	fmt.Fprintln(w, "| field       | value |")
	fmt.Fprintln(w, "| ----------- | ----- |")
	fmt.Fprintf(w, "| id          | %d |\n", e.ID)
	fmt.Fprintf(w, "| created_at  | %s |\n", e.CreatedAt.UTC().Format(time.RFC3339))
	fmt.Fprintf(w, "| updated_at  | %s |\n", e.UpdatedAt.UTC().Format(time.RFC3339))
	if e.Tags != "" {
		fmt.Fprintf(w, "| tags        | %s |\n", e.Tags)
	}
	if e.Project != "" {
		fmt.Fprintf(w, "| project     | %s |\n", e.Project)
	}
	if e.Type != "" {
		fmt.Fprintf(w, "| type        | %s |\n", e.Type)
	}
	if e.Impact != "" {
		fmt.Fprintf(w, "| impact      | %s |\n", e.Impact)
	}
	if e.Description != "" {
		fmt.Fprintf(w, "\n%sDescription\n\n%s\n", descPrefix, e.Description)
	}
}

const noProjectKey = "(no project)"

// ToMarkdown renders entries as a review-ready markdown document per
// DEC-013. Returns bytes with the trailing "\n" stripped; the CLI
// layer appends one newline via fmt.Fprintln, matching ToJSON's byte
// contract. Empty entries slice returns header + provenance block
// only (no summary, no groups).
func ToMarkdown(entries []storage.Entry, opts MarkdownOptions) ([]byte, error) {
	var buf bytes.Buffer

	fmt.Fprintln(&buf, "# Bragfile Export")
	fmt.Fprintln(&buf)
	fmt.Fprintf(&buf, "Exported: %s\n", opts.Now.UTC().Format(time.RFC3339))
	fmt.Fprintf(&buf, "Entries: %d\n", len(entries))
	fmt.Fprintf(&buf, "Filters: %s\n", opts.Filters)

	if len(entries) == 0 {
		return trimTrailingNewline(buf.Bytes()), nil
	}

	fmt.Fprintln(&buf)
	fmt.Fprintln(&buf, "## Summary")
	fmt.Fprintln(&buf)
	writeSummaryByType(&buf, entries)
	fmt.Fprintln(&buf)
	writeSummaryByProject(&buf, entries)

	if opts.Flat {
		writeFlatSection(&buf, entries)
	} else {
		writeGroupedSections(&buf, entries)
	}
	return trimTrailingNewline(buf.Bytes()), nil
}

func trimTrailingNewline(b []byte) []byte {
	return bytes.TrimRight(b, "\n")
}

type countedKey struct {
	key   string
	count int
}

func sortedCountsDescByCountAscByKey(m map[string]int) []countedKey {
	out := make([]countedKey, 0, len(m))
	for k, v := range m {
		out = append(out, countedKey{key: k, count: v})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].count != out[j].count {
			return out[i].count > out[j].count
		}
		return out[i].key < out[j].key
	})
	return out
}

func forceNoProjectLast(sorted []countedKey) []countedKey {
	idx := -1
	for i, e := range sorted {
		if e.key == noProjectKey {
			idx = i
			break
		}
	}
	if idx < 0 {
		return sorted
	}
	np := sorted[idx]
	sorted = append(sorted[:idx], sorted[idx+1:]...)
	return append(sorted, np)
}

func writeSummaryByType(w io.Writer, entries []storage.Entry) {
	fmt.Fprintln(w, "**By type**")
	counts := map[string]int{}
	for _, e := range entries {
		counts[e.Type]++
	}
	for _, kv := range sortedCountsDescByCountAscByKey(counts) {
		fmt.Fprintf(w, "- %s: %d\n", kv.key, kv.count)
	}
}

func writeSummaryByProject(w io.Writer, entries []storage.Entry) {
	fmt.Fprintln(w, "**By project**")
	counts := map[string]int{}
	for _, e := range entries {
		key := e.Project
		if key == "" {
			key = noProjectKey
		}
		counts[key]++
	}
	sorted := forceNoProjectLast(sortedCountsDescByCountAscByKey(counts))
	for _, kv := range sorted {
		fmt.Fprintf(w, "- %s: %d\n", kv.key, kv.count)
	}
}

func writeGroupedSections(w io.Writer, entries []storage.Entry) {
	groups := map[string][]storage.Entry{}
	for _, e := range entries {
		key := e.Project
		if key == "" {
			key = noProjectKey
		}
		groups[key] = append(groups[key], e)
	}
	keys := make([]string, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i] == noProjectKey {
			return false
		}
		if keys[j] == noProjectKey {
			return true
		}
		return keys[i] < keys[j]
	})

	for _, k := range keys {
		group := groups[k]
		sort.SliceStable(group, func(i, j int) bool {
			return group[i].CreatedAt.Before(group[j].CreatedAt)
		})
		fmt.Fprintln(w)
		fmt.Fprintf(w, "## %s\n", k)
		fmt.Fprintln(w)
		for i, e := range group {
			RenderEntry(w, e, 3)
			if i != len(group)-1 {
				fmt.Fprint(w, "\n---\n\n")
			}
		}
	}
}

func writeFlatSection(w io.Writer, entries []storage.Entry) {
	sorted := make([]storage.Entry, len(entries))
	copy(sorted, entries)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].CreatedAt.Before(sorted[j].CreatedAt)
	})
	fmt.Fprintln(w)
	fmt.Fprintln(w, "## Entries (chronological)")
	fmt.Fprintln(w)
	for i, e := range sorted {
		RenderEntry(w, e, 3)
		if i != len(sorted)-1 {
			fmt.Fprint(w, "\n---\n\n")
		}
	}
}
