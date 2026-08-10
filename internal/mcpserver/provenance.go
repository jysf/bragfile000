// Package mcpserver builds the local stdio MCP server (`brag mcp serve`)
// exposing brag_add/brag_list/brag_memory/brag_search/brag_stats as thin
// typed tools over *storage.Store, plus the brag:// push surface (three
// resources serving a ranked, token-budgeted memory slice with no tool
// call). See DEC-024, SPEC-040, DEC-045, and SPEC-074.
package mcpserver

import (
	"fmt"
	"strings"

	"github.com/jysf/bragfile000/internal/capture"
)

// reservedTag normalizes value into a reserved-namespace tag
// "<prefix>:<value>": lowercased, whitespace runs collapsed to a single
// '-', commas stripped (a comma would split the tag, DEC-004). Returns ""
// when value is empty or whitespace-only.
func reservedTag(prefix, value string) string {
	v := strings.ToLower(strings.TrimSpace(value))
	v = strings.Join(strings.Fields(v), "-")
	v = strings.ReplaceAll(v, ",", "")
	if v == "" {
		return ""
	}
	return prefix + ":" + v
}

// stampProvenance appends the reserved provenance tags (in a fixed order:
// agent:, model:, session:, cost:, tokens:) to tags, after the caller's own
// tokens. Empty inputs contribute no tag (DEC-024/DEC-027). session reuses the
// reservedTag normalization (opaque id); cost/tokens are the caller's
// pre-validated, pre-normalized numeric strings (see capture.NormalizeCost /
// capture.NormalizeTokens) appended verbatim so a validated number is never re-mangled.
// The result is a comma-joined string Store.Add canonicalizes like any other
// tags input.
//
// The prefix set stamped is driven by capture.ReservedTagPrefixes — the
// SAME list isReservedTag/validateTagsChanged strip against on the edit
// path (DEC-046 punch-list item 2) — rather than an independently
// maintained set of literals. A prefix in that list with no case below
// panics loudly instead of silently stamping nothing (or, before this fix,
// letting a genuinely new stamped prefix escape stripping because the two
// lists were free to drift apart).
func stampProvenance(tags, agent, model, session, cost, tokens string) string {
	toks := []string{}
	for _, t := range strings.Split(tags, ",") {
		if t = strings.TrimSpace(t); t != "" {
			toks = append(toks, t)
		}
	}
	for _, prefix := range capture.ReservedTagPrefixes {
		var tok string
		switch prefix {
		case "agent:":
			tok = reservedTag("agent", agent)
		case "model:":
			tok = reservedTag("model", model)
		case "session:":
			tok = reservedTag("session", session)
		case "cost:":
			if cost != "" { // already validated + normalized by NormalizeCost
				tok = "cost:" + cost
			}
		case "tokens:":
			if tokens != "" { // already validated + normalized by NormalizeTokens
				tok = "tokens:" + tokens
			}
		default:
			panic(fmt.Sprintf("stampProvenance: reserved prefix %q has no stamping case — add one above", prefix))
		}
		if tok != "" {
			toks = append(toks, tok)
		}
	}
	return strings.Join(toks, ",")
}
