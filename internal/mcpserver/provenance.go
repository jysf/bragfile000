// Package mcpserver builds the local stdio MCP server (`brag mcp serve`)
// exposing brag_add/brag_list/brag_search/brag_stats as thin typed tools
// over *storage.Store. See DEC-024 and SPEC-040.
package mcpserver

import (
	"fmt"
	"strconv"
	"strings"
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
// pre-validated, pre-normalized numeric strings (see normalizeCost /
// normalizeTokens) appended verbatim so a validated number is never re-mangled.
// The result is a comma-joined string Store.Add canonicalizes like any other
// tags input.
func stampProvenance(tags, agent, model, session, cost, tokens string) string {
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
	if sv := reservedTag("session", session); sv != "" {
		toks = append(toks, sv)
	}
	if cost != "" { // already validated + normalized by normalizeCost
		toks = append(toks, "cost:"+cost)
	}
	if tokens != "" { // already validated + normalized by normalizeTokens
		toks = append(toks, "tokens:"+tokens)
	}
	return strings.Join(toks, ",")
}

// normalizeCost validates a caller-reported cost as a non-negative USD decimal
// string (DEC-027). It trims whitespace; empty input yields ("", nil) so an
// omitted cost stamps no tag. A non-empty value must match a plain decimal
// shape ([0-9]+(\.[0-9]+)?) — rejecting negatives, scientific notation,
// currency symbols, and thousands separators — and is returned as the trimmed
// canonical string (never a re-formatted float, to avoid drift). Anything else
// is an error.
func normalizeCost(raw string) (string, error) {
	v := strings.TrimSpace(raw)
	if v == "" {
		return "", nil
	}
	if !isDecimal(v) {
		return "", fmt.Errorf("cost %q: must be a non-negative decimal (e.g. 0.42)", raw)
	}
	if _, err := strconv.ParseFloat(v, 64); err != nil {
		return "", fmt.Errorf("cost %q: %w", raw, err)
	}
	return v, nil
}

// normalizeTokens validates a caller-reported token count as a non-negative
// integer string (DEC-027). It trims whitespace; empty input yields ("", nil)
// so an omitted value stamps no tag. A non-empty value is parsed with
// strconv.ParseUint (rejecting negatives, decimals, non-digits, and radix
// prefixes) and returned as the trimmed string. Anything else is an error.
func normalizeTokens(raw string) (string, error) {
	v := strings.TrimSpace(raw)
	if v == "" {
		return "", nil
	}
	if _, err := strconv.ParseUint(v, 10, 64); err != nil {
		return "", fmt.Errorf("tokens %q: must be a non-negative integer (e.g. 18000)", raw)
	}
	return v, nil
}

// isDecimal reports whether s is a plain non-negative decimal: one or more
// digits, optionally followed by a single '.' and one or more digits. No sign,
// exponent, separators, or leading/trailing dot. strconv.ParseFloat accepts
// "1e3", "+1", "  1", ".5" etc., so this tighter guard runs first.
func isDecimal(s string) bool {
	if s == "" {
		return false
	}
	dot := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '.' {
			if dot || i == 0 || i == len(s)-1 {
				return false
			}
			dot = true
			continue
		}
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
