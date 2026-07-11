package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jysf/bragfile000/internal/config"
	"github.com/jysf/bragfile000/internal/storage"
	"github.com/spf13/cobra"
)

// NewTagCmd returns the `brag tag` parent command with rename and merge
// subcommands. A bare `brag tag` prints help (cobra default for a
// command with subcommands and no RunE).
func NewTagCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tag",
		Short: "Tag taxonomy operations (rename, merge)",
	}
	cmd.AddCommand(newTagRenameCmd())
	cmd.AddCommand(newTagMergeCmd())
	return cmd
}

func newTagRenameCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rename <old> <new>",
		Short: "Rename a tag everywhere",
		Long: `Rename a tag globally: every entry formerly tagged <old> will read <new>.
FTS search is re-synced automatically. Returns an error if <new> already exists
(use 'brag tag merge' to combine two tags instead).

Examples:
  brag tag rename auth authz
  brag tag rename backend infra`,
		RunE: runTagRename,
	}
}

func runTagRename(cmd *cobra.Command, args []string) error {
	if len(args) != 2 {
		return UserErrorf("rename requires exactly <old> and <new> tag names")
	}
	oldName := args[0]
	// Canonicalize the rename target the same way every capture path
	// (add/edit → canonicalizeTags) does, so a renamed tag can never
	// round-trip into different or missing membership. A tag stored in a
	// non-canonical form (untrimmed, whitespace-only, or containing the
	// DEC-004 comma separator) is silently re-split or dropped on the
	// entry's next edit — corrupting membership. Reject or normalize here,
	// at the CLI input boundary, before touching storage.
	newName := strings.TrimSpace(args[1])
	if oldName == "" || newName == "" {
		return UserErrorf("tag names must not be empty")
	}
	if strings.Contains(newName, ",") {
		return UserErrorf("tag name %q must not contain a comma (%q is the tag separator)", newName, ",")
	}
	if oldName == newName {
		return UserErrorf("old and new tag names are the same (%q)", oldName)
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

	if err := s.RenameTag(oldName, newName); err != nil {
		switch {
		case errors.Is(err, storage.ErrTagNotFound):
			return UserErrorf("no tag named %q", oldName)
		case errors.Is(err, storage.ErrTagExists):
			return UserErrorf("tag %q already exists; use `brag tag merge %s %s` to combine them", newName, oldName, newName)
		}
		return fmt.Errorf("rename tag: %w", err)
	}
	fmt.Fprintln(cmd.ErrOrStderr(), "Renamed.")
	return nil
}

func newTagMergeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "merge <src> <dst>",
		Short: "Merge one tag into another, de-duplicating",
		Long: `Fold <src>'s memberships into <dst>, de-duplicating. An entry tagged both
<src> and <dst> ends with exactly one <dst> tagging. The <src> tag row is
deleted. FTS search is re-synced automatically. Both <src> and <dst> must exist
(use 'brag tag rename' to rename a tag that has no conflicts).

Examples:
  brag tag merge auth authz
  brag tag merge backend infra`,
		RunE: runTagMerge,
	}
}

func runTagMerge(cmd *cobra.Command, args []string) error {
	if len(args) != 2 {
		return UserErrorf("merge requires exactly <src> and <dst> tag names")
	}
	src, dst := args[0], args[1]
	if src == "" || dst == "" {
		return UserErrorf("tag names must not be empty")
	}
	if src == dst {
		return UserErrorf("source and destination tags are the same (%q)", src)
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

	if err := s.MergeTags(src, dst); err != nil {
		if errors.Is(err, storage.ErrTagNotFound) {
			return UserErrorf("%v (use `brag tag rename` if you meant to rename)", err)
		}
		return fmt.Errorf("merge tags: %w", err)
	}
	fmt.Fprintln(cmd.ErrOrStderr(), "Merged.")
	return nil
}
