package export

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/jysf/bragfile000/internal/aggregate"
	"github.com/jysf/bragfile000/internal/spark"
	"github.com/jysf/bragfile000/internal/storage"
)

// CoverageOptions controls the rule-based coverage digest (SPEC-045), the
// sixth DEC-014 consumer. Scope echoes the DEC-028/DEC-032 window token
// ("year", "quarter:previous", "since:<raw>"). Filters is the pre-formatted
// markdown line ("(none)" or echoed flags); FiltersJSON is the object the
// JSON envelope renders (nil → {}). ScopeMonths is the ordered set of
// "YYYY-MM" labels in scope — the CLI derives it from the window so the
// monthly series is always fully present (zero-filled), even on an empty
// window. Now is injected for a deterministic Generated: line.
//
// The renderer receives the ALREADY-in-window slice (the CLI does the
// bounded-window filtering) plus ScopeMonths, exactly like wrapped. Spark,
// when true and rendering markdown, prints the agent-share sparkline line
// inside ## Monthly trend (DEC-031); JSON ignores it — a sparkline is a
// lossy visual of by_month[].share, not data (DEC-031 choice f).
type CoverageOptions struct {
	Scope       string
	Filters     string
	FiltersJSON map[string]string
	ScopeMonths []string
	Now         time.Time
	Spark       bool
}

// ToCoverageMarkdown renders the in-window entries as the coverage digest per
// DEC-014/DEC-033: header + provenance block, then ## Provenance share (agent
// vs human counts + %), ## Monthly trend (the agent-share sparkline + per-month
// lines), and ## Self-reference. Returns bytes with the trailing "\n" stripped
// (matches every other renderer). On an empty window only the header +
// provenance block (through "Entries: 0") is emitted; the body sections are
// omitted (DEC-014 part 4).
func ToCoverageMarkdown(entries []storage.Entry, opts CoverageOptions) ([]byte, error) {
	var buf bytes.Buffer
	fmt.Fprintln(&buf, "# Bragfile Coverage")
	fmt.Fprintln(&buf)
	fmt.Fprintf(&buf, "Generated: %s\n", opts.Now.UTC().Format(time.RFC3339))
	fmt.Fprintf(&buf, "Scope: %s\n", opts.Scope)
	fmt.Fprintf(&buf, "Filters: %s\n", opts.Filters)
	fmt.Fprintf(&buf, "Entries: %d\n", len(entries))

	if len(entries) == 0 {
		return trimTrailingNewline(buf.Bytes()), nil
	}

	// Provenance share (overall).
	agent, human := partitionProvenance(entries)
	total := len(entries)
	fmt.Fprintln(&buf)
	fmt.Fprintln(&buf, "## Provenance share")
	fmt.Fprintln(&buf)
	fmt.Fprintf(&buf, "- Agent-authored: %d (%.1f%%)\n", agent, pct(agent, total))
	fmt.Fprintf(&buf, "- Human-authored: %d (%.1f%%)\n", human, pct(human, total))

	// Monthly trend.
	buckets := aggregate.CoverageByMonth(entries, opts.ScopeMonths)
	fmt.Fprintln(&buf)
	fmt.Fprintln(&buf, "## Monthly trend")
	fmt.Fprintln(&buf)
	if opts.Spark {
		shareInts := make([]int, len(buckets))
		for i, b := range buckets {
			shareInts[i] = int(math.Round(b.Share * 100))
		}
		fmt.Fprintf(&buf, "Agent share: %s\n", spark.Line(shareInts))
		fmt.Fprintln(&buf)
	}
	for _, b := range buckets {
		fmt.Fprintf(&buf, "- %s: %d agent / %d human (%.0f%%)\n", b.Period, b.Agent, b.Human, b.Share*100)
	}

	// Self-reference.
	selfRef := aggregate.SelfReferenceCount(entries)
	fmt.Fprintln(&buf)
	fmt.Fprintln(&buf, "## Self-reference")
	fmt.Fprintln(&buf)
	fmt.Fprintf(&buf, "- Entries mentioning brag/bragfile: %d (%.1f%%)\n", selfRef, pct(selfRef, total))

	return trimTrailingNewline(buf.Bytes()), nil
}

// coverageEnvelope is the on-the-wire shape for ToCoverageJSON. Struct-tag
// declaration order is the JSON key order DEC-014/DEC-033 lock (encoding/json
// preserves it).
type coverageEnvelope struct {
	GeneratedAt   string                     `json:"generated_at"`
	Scope         string                     `json:"scope"`
	Filters       map[string]string          `json:"filters"`
	TotalEntries  int                        `json:"total_entries"`
	AgentEntries  int                        `json:"agent_entries"`
	HumanEntries  int                        `json:"human_entries"`
	AgentShare    float64                    `json:"agent_share"`
	ByMonth       []aggregate.CoverageBucket `json:"by_month"`
	SelfReference selfReferenceRecord        `json:"self_reference"`
}

type selfReferenceRecord struct {
	Count int     `json:"count"`
	Share float64 `json:"share"`
}

// ToCoverageJSON renders the DEC-014 envelope with DEC-033's payload keys.
// Every key is always emitted; on an empty window numbers are 0, filters {},
// and by_month is still the full zero-filled month series so the trend slot is
// present even when empty. JSON never contains glyphs — the sparkline is a
// markdown-only rendering (DEC-031 choice f).
func ToCoverageJSON(entries []storage.Entry, opts CoverageOptions) ([]byte, error) {
	agent, human := partitionProvenance(entries)
	total := len(entries)
	selfRef := aggregate.SelfReferenceCount(entries)

	env := coverageEnvelope{
		GeneratedAt:  opts.Now.UTC().Format(time.RFC3339),
		Scope:        opts.Scope,
		Filters:      opts.FiltersJSON,
		TotalEntries: total,
		AgentEntries: agent,
		HumanEntries: human,
		AgentShare:   aggregate.Share(agent, total),
		ByMonth:      aggregate.CoverageByMonth(entries, opts.ScopeMonths),
		SelfReference: selfReferenceRecord{
			Count: selfRef,
			Share: aggregate.Share(selfRef, total),
		},
	}
	if env.Filters == nil {
		env.Filters = map[string]string{}
	}

	return json.MarshalIndent(env, "", "  ")
}

// partitionProvenance counts agent- vs human-authored entries over the full
// slice via aggregate.IsAgentAuthored (SPEC-045 LD7: the totals are the source
// of truth, not a sum of the per-month buckets).
func partitionProvenance(entries []storage.Entry) (agent, human int) {
	for _, e := range entries {
		if aggregate.IsAgentAuthored(e) {
			agent++
		} else {
			human++
		}
	}
	return agent, human
}

// pct returns num/den as a 0-100 percentage (float), or 0 when den == 0. Used
// only for the markdown %.1f%% / %.0f%% display; the JSON shares use
// aggregate.Share (4-decimal fraction).
func pct(num, den int) float64 {
	if den == 0 {
		return 0
	}
	return float64(num) / float64(den) * 100
}
