package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jysf/bragfile000/internal/config"
	"github.com/jysf/bragfile000/internal/export"
	"github.com/jysf/bragfile000/internal/memory"
	"github.com/jysf/bragfile000/internal/storage"
	"github.com/spf13/cobra"
)

// NewMemoryCmd returns the `brag memory` subcommand (SPEC-073), the eighth
// DEC-014 consumer: a ranked, token-budgeted slice of the corpus — the
// corpus read back as working memory, blended by reciprocal-rank fusion
// (DEC-043) and trimmed to a token budget (DEC-044).
func NewMemoryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "memory",
		Short: "Ranked, token-budgeted slice of your history — the corpus as working memory",
		Long: `Print a ranked, token-budgeted slice of your own history — the corpus read back as working memory, cheap enough for an agent to load at the start of every session. Rule-based, deterministic, no LLM, no network.

Ranking blends three signals by reciprocal-rank fusion (DEC-043): how recent an entry is, how well it matches --query, and whether it belongs to --project. With neither flag the blend degenerates to plain recency, so a bare 'brag memory' is the zero-config session opener. --project is a soft boost, not a filter: it raises that project's entries without hiding anyone else's, so a decision you made elsewhere can still surface. Reach for 'brag list --project' when you want a hard filter.

The slice is bounded by an estimated TOKEN budget, not a row count, because a row count does not predict what a slice costs to read. Entries are packed greedily in rank order; one that does not fit is skipped and the fill continues, so a single long entry cannot starve the rest. The Budget section reports what was included, what was skipped, and the estimate, so your next --budget value is a calibration rather than a guess. The budget covers the entry lines; the surrounding envelope is a small fixed overhead on top.

The estimate is a documented character-count heuristic that sizes THIS retrieval. It is not a tokenizer, it is never stored, and it is never stamped on an entry (DEC-044).

Each entry renders as one line:
  - <id> <YYYY-MM-DD> [<project>/<type>] <title> — <impact>
with - for an absent project or type, and the impact clause only when there is one. The id is the handle: run 'brag show <id>' for the full record. The budget is always measured against this markdown line, in both formats, so --format json returns the same entries in the same order.

Output is markdown (default) or a single-object JSON envelope (--format json) per DEC-014.

Examples:
  brag memory                                   # the session opener: recent history, 2000 tokens
  brag memory --project bragfile                # boost one project without hiding the rest
  brag memory --query "auth rate limit"         # what do I already know about this?
  brag memory --query auth --project orbit      # both signals at once
  brag memory --budget 500                      # a tighter slice
  brag memory --format json | jq .slice[].id    # just the ids, for a follow-up call`,
		RunE: runMemory,
	}
	cmd.Flags().String("query", "", "boost entries matching this full-text query (FTS5, DEC-010 semantics)")
	cmd.Flags().String("project", "", "boost entries in this project (a soft boost, not a filter)")
	cmd.Flags().Int("budget", memory.DefaultBudget, "token budget for the entry lines (must be > 0)")
	cmd.Flags().String("format", "markdown", "output format (one of: markdown, json)")
	return cmd
}

func runMemory(cmd *cobra.Command, _ []string) error {
	now := nowFunc()

	budget, _ := cmd.Flags().GetInt("budget")
	if budget <= 0 {
		return UserErrorf("--budget must be greater than zero, got %d (an unbounded slice is `brag list`)", budget)
	}

	format, _ := cmd.Flags().GetString("format")
	if format != "markdown" && format != "json" {
		return UserErrorf("unknown --format value %q (accepted: markdown, json)", format)
	}

	query, _ := cmd.Flags().GetString("query")

	var project string
	if cmd.Flags().Changed("project") {
		project, _ = cmd.Flags().GetString("project")
		if project == "" {
			return UserErrorf("--project must not be empty")
		}
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

	// One bounded read per ranked list (DEC-043 sub-decision 5), shared with
	// the MCP surfaces so both run the SAME composition (SPEC-074 LD9).
	pool, gerr := memory.Gather(s, memory.GatherOptions{Query: query, Project: project})
	if gerr != nil {
		var qe *memory.ErrQuery
		if errors.As(gerr, &qe) {
			return UserErrorf("%v", qe)
		}
		return gerr
	}

	result := memory.Slice(pool.Entries, memory.Options{
		Query:   query,
		Project: project,
		Matched: pool.Matched,
		Budget:  budget,
	})

	filtersMD, filtersJSON := echoFiltersForMemory(query, project)
	opts := export.MemoryOptions{
		Filters:     filtersMD,
		FiltersJSON: filtersJSON,
		Now:         now,
	}

	var body []byte
	switch format {
	case "markdown":
		body, err = export.ToMemoryMarkdown(result, opts)
	case "json":
		body, err = export.ToMemoryJSON(result, opts)
	}
	if err != nil {
		return fmt.Errorf("render memory: %w", err)
	}

	fmt.Fprintln(cmd.OutOrStdout(), string(body))
	return nil
}

// echoFiltersForMemory returns both the markdown filters line (declared
// order: --query then --project) and the JSON filters object (Go's
// map[string]string encoder sorts alphabetically: project, then query — the
// documented DEC-014 markdown/JSON ordering asymmetry). Empty -> "(none)" +
// empty map.
func echoFiltersForMemory(query, project string) (string, map[string]string) {
	jsonObj := map[string]string{}
	var parts []string
	if query != "" {
		jsonObj["query"] = query
		parts = append(parts, fmt.Sprintf("--query %s", query))
	}
	if project != "" {
		jsonObj["project"] = project
		parts = append(parts, fmt.Sprintf("--project %s", project))
	}
	if len(parts) == 0 {
		return "(none)", jsonObj
	}
	return strings.Join(parts, " "), jsonObj
}
