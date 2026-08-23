package export

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/jysf/bragfile000/internal/memory"
)

// MemoryOptions controls the rule-based memory-slice digest (SPEC-073), the
// eighth DEC-014 consumer. Scope is always "lifetime" and is hard-coded
// rather than a field: DEC-014's Scope: is the TIME WINDOW a document covers,
// and memory applies none — a count bound (PoolLimit) is not a date bound.
// That is the property DEC-045 sub-decision 8 leans on when it withholds
// since/until/day from brag_memory. It is NOT a claim that every entry was
// ranked: the pool is capped, so entries older than the PoolLimit-th most
// recent are unreachable on a bare invocation (proven by
// TestMemoryCmd_ThreeReadsComposeThePool/bare-recency-read-is-capped).
// Filters is the pre-formatted markdown line ("(none)" or "--query X
// --project Y" in declared order); FiltersJSON is the object the JSON
// envelope renders (Go's map encoder sorts keys alphabetically — the
// documented DEC-014 markdown/JSON ordering asymmetry). Now is injected for a
// deterministic Generated: line.
type MemoryOptions struct {
	Filters     string
	FiltersJSON map[string]string
	Now         time.Time
}

// ToMemoryMarkdown renders a memory.Result as the DEC-014/DEC-043/DEC-044
// memory-slice digest: header + provenance block, then ## Slice (the ranked,
// budget-trimmed entry lines) and ## Budget (the accounting).
//
// The provenance count is Candidates: N (DEC-048) — how many DISTINCT entries
// were RANKED. It is not the corpus size and not the included count; ##
// Budget decomposes it (Included + Skipped == Candidates). It is a
// pool-composition artifact, not a cap: Gather runs up to three
// PoolLimit-capped reads and Slice dedupes their union, so the number GROWS
// when --query or --project adds a read. It is bounded by
// min(3*PoolLimit, corpus), never by PoolLimit alone. The five sibling
// DEC-014 exporters keep Entries:, which counts entries in scope and NARROWS
// under a filter; this number does the opposite, which is why it does not
// share the word.
//
// On an empty candidate pool only the header block (through "Candidates: 0")
// is emitted — no ## Slice, no ## Budget (DEC-014 part 4). A non-empty pool
// that includes zero entries still renders both sections. Returns bytes with
// the trailing "\n" stripped (matches every other renderer).
func ToMemoryMarkdown(result memory.Result, opts MemoryOptions) ([]byte, error) {
	var buf bytes.Buffer
	fmt.Fprintln(&buf, "# Bragfile Memory")
	fmt.Fprintln(&buf)
	fmt.Fprintf(&buf, "Generated: %s\n", opts.Now.UTC().Format(time.RFC3339))
	fmt.Fprintln(&buf, "Scope: lifetime")
	fmt.Fprintf(&buf, "Filters: %s\n", opts.Filters)
	fmt.Fprintf(&buf, "Candidates: %d\n", result.Candidates)

	if result.Candidates == 0 {
		return trimTrailingNewline(buf.Bytes()), nil
	}

	fmt.Fprintln(&buf)
	fmt.Fprintln(&buf, "## Slice")
	fmt.Fprintln(&buf)
	for _, item := range result.Items {
		fmt.Fprintln(&buf, item.Line)
	}

	fmt.Fprintln(&buf)
	fmt.Fprintln(&buf, "## Budget")
	fmt.Fprintln(&buf)
	fmt.Fprintf(&buf, "- Budget: %d tokens\n", result.Budget)
	fmt.Fprintf(&buf, "- Estimated: %d tokens\n", result.EstimatedTokens)
	fmt.Fprintf(&buf, "- Included: %d\n", result.Included)
	fmt.Fprintf(&buf, "- Skipped: %d\n", result.Skipped)

	return trimTrailingNewline(buf.Bytes()), nil
}

// memoryItemRecord is the deliberately NARROW 9-key per-item projection
// (DEC-011 choice narrowed per SPEC-073) — no description, no tags, no
// updated_at, plus rank/score/tokens. Struct-tag declaration order is the
// JSON key order DEC-014/DEC-043/DEC-044 lock.
type memoryItemRecord struct {
	ID        int64   `json:"id"`
	CreatedAt string  `json:"created_at"`
	Project   string  `json:"project"`
	Type      string  `json:"type"`
	Title     string  `json:"title"`
	Impact    string  `json:"impact"`
	Rank      int     `json:"rank"`
	Score     float64 `json:"score"`
	Tokens    int     `json:"tokens"`
}

// memoryEnvelope is the on-the-wire shape for ToMemoryJSON. Struct-tag order
// is the JSON key order DEC-014 locks (flat provenance first).
type memoryEnvelope struct {
	GeneratedAt     string             `json:"generated_at"`
	Scope           string             `json:"scope"`
	Filters         map[string]string  `json:"filters"`
	Candidates      int                `json:"candidates"`
	Budget          int                `json:"budget"`
	EstimatedTokens int                `json:"estimated_tokens"`
	Included        int                `json:"included"`
	Skipped         int                `json:"skipped"`
	Slice           []memoryItemRecord `json:"slice"`
}

// ToMemoryJSON renders the DEC-014 envelope with SPEC-073's payload keys.
// Every key is always emitted; on an empty candidate pool numbers are 0,
// filters {}, and slice [] (non-nil). score is rounded to 6 decimal places
// (DEC-043 sub-decision 6) so the goldens stay readable and the constants
// stay observable.
func ToMemoryJSON(result memory.Result, opts MemoryOptions) ([]byte, error) {
	slice := make([]memoryItemRecord, 0, len(result.Items))
	for _, item := range result.Items {
		slice = append(slice, memoryItemRecord{
			ID:        item.Entry.ID,
			CreatedAt: item.Entry.CreatedAt.UTC().Format(time.RFC3339),
			Project:   item.Entry.Project,
			Type:      item.Entry.Type,
			Title:     item.Entry.Title,
			Impact:    item.Entry.Impact,
			Rank:      item.Rank,
			Score:     math.Round(item.Score*1e6) / 1e6,
			Tokens:    item.Tokens,
		})
	}

	env := memoryEnvelope{
		GeneratedAt:     opts.Now.UTC().Format(time.RFC3339),
		Scope:           "lifetime",
		Filters:         opts.FiltersJSON,
		Candidates:      result.Candidates,
		Budget:          result.Budget,
		EstimatedTokens: result.EstimatedTokens,
		Included:        result.Included,
		Skipped:         result.Skipped,
		Slice:           slice,
	}
	if env.Filters == nil {
		env.Filters = map[string]string{}
	}

	return json.MarshalIndent(env, "", "  ")
}
