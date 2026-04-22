package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/jysf/bragfile000/internal/config"
	"github.com/jysf/bragfile000/internal/storage"
	"github.com/spf13/cobra"
)

// NewSearchCmd returns the `brag search <query>` subcommand. It runs
// a full-text search over entries via FTS5 (see SPEC-011's
// entries_fts index) and prints matching rows to stdout in the same
// tab-separated shape as `brag list`. Query semantics are per DEC-010.
func NewSearchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search brag entries via FTS5",
		Long: `Search entries by content across title, description, tags, project, and impact.

Query semantics (DEC-010):
  - Tokens are whitespace-separated
  - Each token is treated as a literal string (no FTS5 operators)
  - Multiple tokens: AND semantics (entries must contain all words)

Examples:
  brag search "auth"                   # find entries mentioning auth
  brag search "cut latency"            # entries with both "cut" AND "latency"
  brag search "auth-refactor"          # literal match; hyphen is not NOT-operator
  brag search "redis" --limit 5        # top 5 matches`,
		RunE: runSearch,
	}
	cmd.Flags().Int("limit", 0, "cap result count (0 = unlimited)")
	return cmd
}

// buildFTS5Query converts a user-typed search argument into an FTS5
// MATCH-compatible string per DEC-010: tokenize on whitespace,
// phrase-quote each token, join with spaces (FTS5's implicit AND).
// Empty, whitespace-only, or quote-containing input is a user error.
func buildFTS5Query(raw string) (string, error) {
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

func runSearch(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return UserErrorf("search requires exactly one query argument")
	}

	fts5, err := buildFTS5Query(args[0])
	if err != nil {
		return UserErrorf("%v", err)
	}

	limit := 0
	if cmd.Flags().Changed("limit") {
		n, _ := cmd.Flags().GetInt("limit")
		if n < 0 {
			return UserErrorf("--limit must be zero or positive, got %d", n)
		}
		limit = n
	}

	dbFlag := getFlagString(cmd, "db")
	path, err := config.ResolveDBPath(dbFlag)
	if err != nil {
		return fmt.Errorf("resolve db path: %w", err)
	}

	s, err := storage.Open(path)
	if err != nil {
		return fmt.Errorf("open store: %w", err)
	}
	defer s.Close()

	entries, err := s.Search(fts5, limit)
	if err != nil {
		return fmt.Errorf("search: %w", err)
	}

	out := cmd.OutOrStdout()
	for _, e := range entries {
		fmt.Fprintf(out, "%d\t%s\t%s\n",
			e.ID,
			e.CreatedAt.UTC().Format(time.RFC3339),
			e.Title)
	}
	return nil
}
