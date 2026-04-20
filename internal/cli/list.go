package cli

import (
	"fmt"
	"time"

	"github.com/jysf/bragfile000/internal/config"
	"github.com/jysf/bragfile000/internal/storage"
	"github.com/spf13/cobra"
)

// NewListCmd returns the `brag list` subcommand. It prints every entry
// to stdout, one per line, as `<id>\t<created_at>\t<title>` in the
// order returned by Store.List (created_at DESC, id DESC tie-break).
// No filter flags ship in STAGE-001.
func NewListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all brag entries",
		Long:  "List all brag entries in reverse-chronological order. Each line is tab-separated: <id>\\t<created_at>\\t<title>.",
		RunE:  runList,
	}
}

func runList(cmd *cobra.Command, _ []string) error {
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

	out := cmd.OutOrStdout()
	for _, e := range entries {
		fmt.Fprintf(out, "%d\t%s\t%s\n",
			e.ID,
			e.CreatedAt.UTC().Format(time.RFC3339),
			e.Title)
	}
	return nil
}
