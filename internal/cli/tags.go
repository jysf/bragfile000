package cli

import (
	"fmt"

	"github.com/jysf/bragfile000/internal/config"
	"github.com/jysf/bragfile000/internal/export"
	"github.com/jysf/bragfile000/internal/storage"
	"github.com/spf13/cobra"
)

// NewTagsCmd returns the `brag tags` subcommand. Lists every in-use tag
// with its usage count. Data goes to stdout; errors to stderr.
func NewTagsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tags",
		Short: "List every tag with its usage count",
		Long: `List every in-use tag with its total usage count across all entries, one per line as <name>` + "\t" + `<count>, sorted by count (descending) then name (ascending).

Output is plain tab-separated rows (default) or a JSON array of {tag, count} objects (--format json) per DEC-016. Tags with no remaining uses are omitted. Counts span all taggable objects (entries today; projects in a later release).

Examples:
  brag tags                         # name<TAB>count rows, most-used first
  brag tags --format json           # naked JSON array of {tag, count}`,
		RunE: runTags,
	}
	cmd.Flags().String("format", "", "output format (one of: json); default is plain tab-separated")
	return cmd
}

func runTags(cmd *cobra.Command, _ []string) error {
	format, _ := cmd.Flags().GetString("format")
	if format != "" && format != "json" {
		return UserErrorf("unknown --format value %q (accepted: json)", format)
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

	tags, err := s.TagCounts()
	if err != nil {
		return fmt.Errorf("tag counts: %w", err)
	}

	out := cmd.OutOrStdout()
	if format == "json" {
		body, err := export.ToTagsJSON(tags)
		if err != nil {
			return fmt.Errorf("render tags json: %w", err)
		}
		fmt.Fprintln(out, string(body))
		return nil
	}

	for _, tc := range tags {
		fmt.Fprintf(out, "%s\t%d\n", tc.Name, tc.Count)
	}
	return nil
}
