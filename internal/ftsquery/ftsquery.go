// Package ftsquery holds the single DEC-010 search transform: the conversion
// from a user-typed search argument into an FTS5 MATCH expression.
//
// It exists because there were two of it. `cli.buildFTS5Query` and
// `mcpserver.buildMatch` were byte-identical implementations kept in agreement
// by nothing but care; DEC-024 recorded that duplication as a negative
// consequence at the moment it was created, with an explicit revisit trigger —
// *"a third consumer of the DEC-010 search transform appears → extract the
// shared query builder."* `brag memory` over MCP is that third consumer
// (SPEC-074 / DEC-045, LD8), so the trigger fires here.
//
// The package is pure and stdlib-only, and both `internal/cli` and
// `internal/mcpserver` import it — the same shape as `internal/timewindow`
// (SPEC-072 / DEC-042), which resolved the identical problem for the
// `--since`/`--day` parsers one spec earlier in this stage. Keeping the
// transform in one place is what lets `brag search`, `brag memory`,
// `brag_search` and `brag_memory` make the same promise about query syntax
// without a drift-guard test standing in for a shared function.
package ftsquery

import (
	"fmt"
	"strings"
)

// Build converts a user-typed search argument into an FTS5 MATCH-compatible
// string per DEC-010: tokenize on whitespace, phrase-quote each token, join
// with spaces (FTS5's implicit AND). Phrase-quoting is what makes a hyphenated
// token a literal rather than FTS5's NOT operator.
//
// Empty, whitespace-only, or quote-containing input is an error. Callers
// surface it as a user error: a UserError on the CLI, a tool error over MCP.
func Build(raw string) (string, error) {
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
