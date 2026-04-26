package cli

import (
	"fmt"
	"time"

	"github.com/jysf/bragfile000/internal/config"
	"github.com/jysf/bragfile000/internal/export"
	"github.com/jysf/bragfile000/internal/storage"
	"github.com/spf13/cobra"
)

// NewStatsCmd returns the `brag stats` subcommand. SPEC-020 ships it as
// the third (and final) DEC-014 consumer in STAGE-004: six lifetime
// aggregations over the entire corpus, rendered as markdown (default)
// or the DEC-014 JSON envelope. No filter flags, no --range, no --out.
func NewStatsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Lifetime stats: six aggregations over the entire corpus",
		Long: `Print six lifetime aggregations over the entire corpus: total entries, entries/week (rolling average over the corpus span), current and longest streak (consecutive UTC days with entries), top-5 most-common tags, top-5 most-common projects, plus the corpus span (first entry, last entry, days). No LLM ships in the binary.

Output is markdown (default) or a single-object JSON envelope (--format json) per DEC-014. The JSON shape uses arrays of objects for top_tags and top_projects to preserve DESC-by-count ordering; corpus_span is a sub-object with date-or-null fields.

Stats covers the lifetime corpus only — no time window, no filters. Use brag summary for windowed digests; brag review for reflection over the last 7 or 30 days. Stdout only; redirect with > if you want a file.

Examples:
  brag stats                        # lifetime corpus, markdown
  brag stats --format json          # lifetime corpus, JSON envelope`,
		RunE: runStats,
	}
	cmd.Flags().String("format", "markdown", "output format (one of: markdown, json)")
	return cmd
}

func runStats(cmd *cobra.Command, _ []string) error {
	format, _ := cmd.Flags().GetString("format")
	if format != "markdown" && format != "json" {
		return UserErrorf("unknown --format value %q (accepted: markdown, json)", format)
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

	entries, err := s.List(storage.ListFilter{})
	if err != nil {
		return fmt.Errorf("list entries: %w", err)
	}

	// Single Now source: locked decision §10. The same value drives the
	// renderer's Generated: line AND aggregate.Streak's today reference,
	// avoiding a midnight-UTC race between two time.Now() calls.
	opts := export.StatsOptions{Now: time.Now().UTC()}

	var body []byte
	switch format {
	case "markdown":
		body, err = export.ToStatsMarkdown(entries, opts)
	case "json":
		body, err = export.ToStatsJSON(entries, opts)
	}
	if err != nil {
		return fmt.Errorf("render stats: %w", err)
	}

	fmt.Fprintln(cmd.OutOrStdout(), string(body))
	return nil
}
