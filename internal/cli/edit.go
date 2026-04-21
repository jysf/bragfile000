package cli

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/jysf/bragfile000/internal/config"
	"github.com/jysf/bragfile000/internal/editor"
	"github.com/jysf/bragfile000/internal/storage"
	"github.com/spf13/cobra"
)

// testEditFunc is an editor.EditFunc override used by tests to avoid
// spawning a real editor subprocess. Production leaves this nil;
// runEdit falls back to editor.Default when unset.
var testEditFunc editor.EditFunc

// NewEditCmd returns the `brag edit <id>` subcommand. This is the
// primary update mechanism for PROJ-001 (DEC-009 pins the buffer
// format; flag-based update is a deferred polish spec).
func NewEditCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "edit <id>",
		Short: "Edit a brag entry in $EDITOR",
		Long: `Open a brag entry in $EDITOR and save changes back to the database.

This is the update mechanism for brag entries — edit any field by rewriting
the buffer and saving. Save without modifications (or quit without saving)
to abort cleanly with no database write.

Examples:
  brag edit 42                # open entry 42 in $EDITOR
  EDITOR=code brag edit 42    # override editor for one invocation`,
		RunE: runEdit,
	}
}

func runEdit(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return UserErrorf("edit requires exactly one <id> argument")
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

	current, err := s.Get(id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return UserErrorf("no entry with id %d", id)
		}
		return fmt.Errorf("get entry: %w", err)
	}

	initial := editor.Render(editor.Fields{
		Title:       current.Title,
		Description: current.Description,
		Tags:        current.Tags,
		Project:     current.Project,
		Type:        current.Type,
		Impact:      current.Impact,
	})

	editFn := testEditFunc
	if editFn == nil {
		editFn = editor.Default
	}
	edited, changed, err := editor.Launch(initial, editFn)
	if err != nil {
		return fmt.Errorf("launch editor: %w", err)
	}
	if !changed {
		fmt.Fprintln(cmd.ErrOrStderr(), "No changes.")
		return nil
	}

	parsed, err := editor.Parse(edited)
	if err != nil {
		return UserErrorf("invalid buffer: %v", err)
	}

	if _, err := s.Update(id, storage.Entry{
		Title:       parsed.Title,
		Description: parsed.Description,
		Tags:        parsed.Tags,
		Project:     parsed.Project,
		Type:        parsed.Type,
		Impact:      parsed.Impact,
	}); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return UserErrorf("no entry with id %d", id)
		}
		return fmt.Errorf("update entry: %w", err)
	}
	fmt.Fprintln(cmd.ErrOrStderr(), "Updated.")
	return nil
}
