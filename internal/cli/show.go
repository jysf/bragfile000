package cli

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/jysf/bragfile000/internal/config"
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

	renderEntry(cmd.OutOrStdout(), entry)
	return nil
}

func renderEntry(w io.Writer, e storage.Entry) {
	fmt.Fprintf(w, "# %s\n\n", e.Title)
	fmt.Fprintln(w, "| field       | value |")
	fmt.Fprintln(w, "| ----------- | ----- |")
	fmt.Fprintf(w, "| id          | %d |\n", e.ID)
	fmt.Fprintf(w, "| created_at  | %s |\n", e.CreatedAt.UTC().Format(time.RFC3339))
	fmt.Fprintf(w, "| updated_at  | %s |\n", e.UpdatedAt.UTC().Format(time.RFC3339))
	if e.Tags != "" {
		fmt.Fprintf(w, "| tags        | %s |\n", e.Tags)
	}
	if e.Project != "" {
		fmt.Fprintf(w, "| project     | %s |\n", e.Project)
	}
	if e.Type != "" {
		fmt.Fprintf(w, "| type        | %s |\n", e.Type)
	}
	if e.Impact != "" {
		fmt.Fprintf(w, "| impact      | %s |\n", e.Impact)
	}
	if e.Description != "" {
		fmt.Fprintf(w, "\n## Description\n\n%s\n", e.Description)
	}
}
