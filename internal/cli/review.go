package cli

import (
	"fmt"
	"time"

	"github.com/jysf/bragfile000/internal/config"
	"github.com/jysf/bragfile000/internal/export"
	"github.com/jysf/bragfile000/internal/storage"
	"github.com/spf13/cobra"
)

// NewReviewCmd returns the `brag review` subcommand. SPEC-019 emits it
// as the second DEC-014 consumer: a reflection-prompt digest of recent
// entries grouped by project, followed by three hard-coded reflection
// questions designed to be pasted into an external AI session for
// guided self-review. No LLM ships in the binary.
func NewReviewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "review",
		Short: "Reflection digest: recent entries grouped by project + three reflection questions",
		Long: `Print a reflection digest of recent entries grouped by project, followed by three hard-coded reflection questions designed to be pasted into an external AI session for guided self-review. No LLM ships in the binary.

Output is markdown (default) or a single-object JSON envelope (--format json) per DEC-014. The JSON shape mirrors brag summary's envelope (entries_grouped is an array of {project, entries: [...]} objects).

--week and --month are mutually exclusive named flags. Bare 'brag review' silently defaults to --week. Rolling-window semantics: --week = last 7 UTC days; --month = last 30 UTC days.

Filter flags (tag, project, type) are NOT accepted on review — the digest is "the last 7/30 days, period." Stdout only; redirect with > if you want a file.

Examples:
  brag review                                  # last 7 UTC days, markdown (silent default)
  brag review --week                           # explicit; same as bare invocation
  brag review --month --format json            # last 30 UTC days, JSON envelope`,
		RunE: runReview,
	}
	cmd.Flags().Bool("week", false, "review last 7 UTC days (default if neither --week nor --month is set)")
	cmd.Flags().Bool("month", false, "review last 30 UTC days")
	cmd.Flags().String("format", "markdown", "output format (one of: markdown, json)")
	return cmd
}

func runReview(cmd *cobra.Command, _ []string) error {
	weekSet := cmd.Flags().Changed("week")
	monthSet := cmd.Flags().Changed("month")
	if weekSet && monthSet {
		return UserErrorf("--week and --month are mutually exclusive (use one or neither; neither defaults to --week)")
	}

	scope := "week"
	if monthSet {
		scope = "month"
	}
	cutoff, err := rangeCutoff(scope, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("compute cutoff: %w", err)
	}

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

	entries, err := s.List(storage.ListFilter{Since: cutoff})
	if err != nil {
		return fmt.Errorf("list entries: %w", err)
	}

	opts := export.ReviewOptions{
		Scope: scope,
		Now:   time.Now().UTC(),
	}

	var body []byte
	switch format {
	case "markdown":
		body, err = export.ToReviewMarkdown(entries, opts)
	case "json":
		body, err = export.ToReviewJSON(entries, opts)
	}
	if err != nil {
		return fmt.Errorf("render review: %w", err)
	}

	fmt.Fprintln(cmd.OutOrStdout(), string(body))
	return nil
}
