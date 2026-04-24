package cli

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/jysf/bragfile000/internal/config"
	"github.com/jysf/bragfile000/internal/export"
	"github.com/jysf/bragfile000/internal/storage"
	"github.com/spf13/cobra"
)

// NewShowCmd returns the `brag show <id>` subcommand. Prints the
// matched entry as markdown to stdout. Missing, non-numeric, or non-
// positive IDs surface as ErrUser (exit 1) via runShow; positional-arg
// validation lives inline here rather than in a cobra.Args validator
// because cobra's built-ins return unwrappable plain errors (DEC-007
// extension).
func NewShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <id>",
		Short: "Show a single brag entry as markdown",
		Long: `Show a single brag entry as markdown.

Examples:
  brag show 42              # print entry 42
  brag show 42 | glow       # render in a markdown viewer`,
		RunE: runShow,
	}
}

func runShow(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return UserErrorf("show requires exactly one <id> argument")
	}
	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return UserErrorf("invalid id %q: must be a positive integer", args[0])
	}
	if id <= 0 {
		return UserErrorf("invalid id %d: must be positive", id)
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

	entry, err := s.Get(id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return UserErrorf("no entry with id %d", id)
		}
		return fmt.Errorf("get entry: %w", err)
	}

	export.RenderEntry(cmd.OutOrStdout(), entry, 1)
	return nil
}
