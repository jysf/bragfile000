package cli

import (
	"fmt"
	"time"

	"github.com/jysf/bragfile000/internal/config"
	"github.com/jysf/bragfile000/internal/export"
	"github.com/jysf/bragfile000/internal/storage"
	"github.com/spf13/cobra"
)

// NewListCmd returns the `brag list` subcommand. It prints every entry
// matching the filter flags (or all entries when none are set) to
// stdout, one per line, as `<id>\t<created_at>\t<title>` in the order
// returned by Store.List (created_at DESC, id DESC tie-break).
func NewListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List brag entries",
		Long: `List brag entries in reverse-chronological order. Each line is tab-separated: <id>\t<created_at>\t<title>.

Examples:
  brag list                                       # all entries
  brag list --tag auth                            # entries tagged "auth"
  brag list --project platform --since 7d         # last week, one project
  brag list --type shipped --limit 5              # 5 most recent shipped
  brag list --since 2026-01-01                    # since a specific date
  brag list -P                                    # include project column
  brag list --format json                         # pretty-printed JSON array
  brag list --format tsv                          # tab-separated with header row`,
		RunE: runList,
	}
	cmd.Flags().String("tag", "", "filter to entries whose tags contain this token (comma-separated match)")
	cmd.Flags().String("project", "", "filter to entries with this project (exact match)")
	cmd.Flags().String("type", "", "filter to entries with this type (exact match)")
	cmd.Flags().String("since", "", "filter to entries on or after this point (YYYY-MM-DD or Nd/Nw/Nm)")
	cmd.Flags().Int("limit", 0, "cap the number of rows returned (must be > 0 when set)")
	cmd.Flags().BoolP("show-project", "P", false, "include project in output (adds column between created_at and title)")
	cmd.Flags().String("format", "", "output format (one of: json, tsv); default is plain tab-separated")
	return cmd
}

func runList(cmd *cobra.Command, _ []string) error {
	filter := storage.ListFilter{}

	if cmd.Flags().Changed("tag") {
		v, _ := cmd.Flags().GetString("tag")
		if v == "" {
			return UserErrorf("--tag must not be empty")
		}
		filter.Tag = v
	}
	if cmd.Flags().Changed("project") {
		v, _ := cmd.Flags().GetString("project")
		if v == "" {
			return UserErrorf("--project must not be empty")
		}
		filter.Project = v
	}
	if cmd.Flags().Changed("type") {
		v, _ := cmd.Flags().GetString("type")
		if v == "" {
			return UserErrorf("--type must not be empty")
		}
		filter.Type = v
	}
	if cmd.Flags().Changed("since") {
		raw, _ := cmd.Flags().GetString("since")
		t, err := ParseSince(raw)
		if err != nil {
			return UserErrorf("invalid --since %q: %v", raw, err)
		}
		filter.Since = t
	}
	if cmd.Flags().Changed("limit") {
		n, _ := cmd.Flags().GetInt("limit")
		if n <= 0 {
			return UserErrorf("--limit must be positive, got %d", n)
		}
		filter.Limit = n
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

	entries, err := s.List(filter)
	if err != nil {
		return fmt.Errorf("list entries: %w", err)
	}

	format, _ := cmd.Flags().GetString("format")
	showProject, _ := cmd.Flags().GetBool("show-project")
	out := cmd.OutOrStdout()

	switch format {
	case "":
		for _, e := range entries {
			if showProject {
				project := e.Project
				if project == "" {
					project = "-"
				}
				fmt.Fprintf(out, "%d\t%s\t%s\t%s\n",
					e.ID,
					e.CreatedAt.UTC().Format(time.RFC3339),
					project,
					e.Title)
			} else {
				fmt.Fprintf(out, "%d\t%s\t%s\n",
					e.ID,
					e.CreatedAt.UTC().Format(time.RFC3339),
					e.Title)
			}
		}
	case "json":
		b, err := export.ToJSON(entries)
		if err != nil {
			return fmt.Errorf("marshal json: %w", err)
		}
		fmt.Fprintln(out, string(b))
	case "tsv":
		fmt.Fprintln(out, export.TSVHeader)
		for _, e := range entries {
			fmt.Fprintln(out, export.ToTSVRow(e))
		}
	default:
		return UserErrorf("unknown --format value %q (accepted: json, tsv)", format)
	}
	return nil
}
