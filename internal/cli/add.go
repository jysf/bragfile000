package cli

import (
	"fmt"
	"strings"

	"github.com/jysf/bragfile000/internal/config"
	"github.com/jysf/bragfile000/internal/storage"
	"github.com/spf13/cobra"
)

// NewAddCmd returns the `brag add` subcommand. Requires --title; every
// other field is optional and persisted verbatim (tag normalization is
// deferred per DEC-004).
func NewAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new brag entry",
		Long: `Add a new brag entry. Title is required; other fields are optional.

Examples:
  brag add -t "shipped the auth refactor"
  brag add -t "cut p99 latency" -T "auth,perf" -p "platform" \
           -i "unblocked mobile v3 release"
  brag add --title "..." --description "..." --tags "..." \
           --project "..." --type "..." --impact "..."

Short forms: -t title, -d description, -T tags, -p project,
-k type, -i impact.`,
		RunE: runAdd,
	}
	cmd.Flags().StringP("title", "t", "", "short headline (required)")
	cmd.Flags().StringP("description", "d", "", "free-form body")
	cmd.Flags().StringP("tags", "T", "", "comma-joined tag list (e.g. \"auth,perf\")")
	cmd.Flags().StringP("project", "p", "", "project / initiative this brag belongs to")
	cmd.Flags().StringP("type", "k", "", "free-form category (shipped, learned, mentored, ...)")
	cmd.Flags().StringP("impact", "i", "", "impact statement (metric, quote, outcome)")
	// MarkFlagRequired is intentionally omitted: cobra's required-flag
	// validation returns a plain error that cannot carry our ErrUser
	// sentinel, and the RunE TrimSpace check below already covers
	// missing, empty, and whitespace-only --title. See DEC-007.
	return cmd
}

func runAdd(cmd *cobra.Command, _ []string) error {
	title := getFlagString(cmd, "title")
	if strings.TrimSpace(title) == "" {
		return UserErrorf("--title is required and must not be empty")
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

	entry := storage.Entry{
		Title:       title,
		Description: getFlagString(cmd, "description"),
		Tags:        getFlagString(cmd, "tags"),
		Project:     getFlagString(cmd, "project"),
		Type:        getFlagString(cmd, "type"),
		Impact:      getFlagString(cmd, "impact"),
	}
	inserted, err := s.Add(entry)
	if err != nil {
		return fmt.Errorf("add entry: %w", err)
	}

	fmt.Fprintln(cmd.OutOrStdout(), inserted.ID)
	return nil
}

func getFlagString(cmd *cobra.Command, name string) string {
	v, _ := cmd.Flags().GetString(name)
	return v
}
