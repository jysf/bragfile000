package cli

import (
	"bufio"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/jysf/bragfile000/internal/config"
	"github.com/jysf/bragfile000/internal/storage"
	"github.com/spf13/cobra"
)

// NewDeleteCmd returns the `brag delete <id>` subcommand. Prompts for
// y/N confirmation on stdin unless --yes/-y is passed; on decline the
// command exits 0 cleanly (a deliberate user choice, not an error).
// Positional-arg validation lives inline in runDelete per DEC-007.
func NewDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a brag entry",
		Long: `Delete a brag entry. Prompts for confirmation unless --yes is passed.

Examples:
  brag delete 42              # prompts y/N on stdin
  brag delete 42 --yes        # skip prompt
  brag delete 42 -y           # same via shorthand`,
		RunE: runDelete,
	}
	cmd.Flags().BoolP("yes", "y", false, "skip confirmation prompt")
	return cmd
}

func runDelete(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return UserErrorf("delete requires exactly one <id> argument")
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

	yes, _ := cmd.Flags().GetBool("yes")
	if !yes {
		fmt.Fprintf(cmd.ErrOrStderr(),
			"Delete entry %d (%q)? [y/N] ", id, entry.Title)
		reader := bufio.NewReader(cmd.InOrStdin())
		line, _ := reader.ReadString('\n')
		line = strings.TrimRight(line, "\r\n")
		if line != "y" && line != "Y" {
			fmt.Fprintln(cmd.ErrOrStderr(), "Aborted.")
			return nil
		}
	}

	if err := s.Delete(id); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return UserErrorf("no entry with id %d", id)
		}
		return fmt.Errorf("delete entry: %w", err)
	}
	fmt.Fprintln(cmd.ErrOrStderr(), "Deleted.")
	return nil
}
