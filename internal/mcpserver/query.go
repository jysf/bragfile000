package mcpserver

import (
	"fmt"
	"strings"
)

// buildMatch converts a user-typed search argument into an FTS5
// MATCH-compatible string per DEC-010: tokenize on whitespace, phrase-quote
// each token, join with spaces (FTS5's implicit AND). Mirrors
// cli.buildFTS5Query (second consumer; extraction deferred — DEC-024).
// Empty, whitespace-only, or quote-containing input is an error.
func buildMatch(raw string) (string, error) {
	if strings.ContainsRune(raw, '"') {
		return "", fmt.Errorf("search query must not contain quotes")
	}
	tokens := strings.Fields(raw)
	if len(tokens) == 0 {
		return "", fmt.Errorf("search query must not be empty")
	}
	parts := make([]string, len(tokens))
	for i, tok := range tokens {
		parts[i] = `"` + tok + `"`
	}
	return strings.Join(parts, " "), nil
}
