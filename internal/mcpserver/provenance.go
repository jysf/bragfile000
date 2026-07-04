// Package mcpserver builds the local stdio MCP server (`brag mcp serve`)
// exposing brag_add/brag_list/brag_search/brag_stats as thin typed tools
// over *storage.Store. See DEC-024 and SPEC-040.
package mcpserver

import "strings"

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

// stampProvenance appends agent:/model: reserved tags (in that order) to
// tags, after the caller's own tokens. Empty agent/model contribute no
// tag. The result is a comma-joined string Store.Add canonicalizes like
// any other tags input.
func stampProvenance(tags, agent, model string) string {
	toks := []string{}
	for _, t := range strings.Split(tags, ",") {
		if t = strings.TrimSpace(t); t != "" {
			toks = append(toks, t)
		}
	}
	if a := reservedTag("agent", agent); a != "" {
		toks = append(toks, a)
	}
	if m := reservedTag("model", model); m != "" {
		toks = append(toks, m)
	}
	return strings.Join(toks, ",")
}
