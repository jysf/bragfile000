package cli

import (
	"fmt"
	"strings"

	"github.com/jysf/bragfile000/internal/config"
	"github.com/jysf/bragfile000/internal/editor"
	"github.com/jysf/bragfile000/internal/storage"
	"github.com/spf13/cobra"
)

// addFieldFlags is the closed list of entry-field flag names that
// trigger flag-mode dispatch. The root persistent flag --db is
// deliberately NOT in this list (it's a path override, not an entry
// field), so `brag add --db /tmp/x.db` still opens the editor.
var addFieldFlags = []string{"title", "description", "tags", "project", "type", "impact"}

// NewAddCmd returns the `brag add` subcommand. Two modes:
//   - flag mode: any of the six entry-field flags set; --title is
//     required (DEC-007).
//   - editor mode: no entry-field flag set; opens $EDITOR on a hint
//     template (DEC-009 buffer format), parses + persists on save.
func NewAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new brag entry",
		Long: `Add a new brag entry, either via flags or by opening $EDITOR.

Flag mode (any of -t, -d, -T, -p, -k, -i set): inserts directly from flag
values. --title is required.

Editor mode (no entry-field flags set): opens $EDITOR on a template buffer.
Save a valid entry to insert it; save unchanged to abort cleanly.

Examples:
  brag add                                          # editor mode
  brag add -t "shipped the auth refactor"
  brag add -t "cut p99 latency" -T "auth,perf" -p "platform" \
           -i "unblocked mobile v3 release"
  brag add --title "..." --description "..." --tags "..." \
           --project "..." --type "..." --impact "..."

Short forms: -t title, -d description, -T tags, -p project,
-k type, -i impact.`,
		RunE: runAdd,
	}
	cmd.Flags().StringP("title", "t", "", "short headline (required in flag mode)")
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

// runAdd dispatches to flag mode or editor mode based on whether any
// of the six entry-field flags was set on the command line.
func runAdd(cmd *cobra.Command, args []string) error {
	for _, name := range addFieldFlags {
		if cmd.Flags().Changed(name) {
			return runAddFlags(cmd, args)
		}
	}
	return runAddEditor(cmd)
}

func runAddFlags(cmd *cobra.Command, _ []string) error {
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

func runAddEditor(cmd *cobra.Command) error {
	editFn := testEditFunc
	if editFn == nil {
		editFn = editor.Default
	}
	edited, changed, err := editor.Launch(editor.EmptyTemplate(), editFn)
	if err != nil {
		return fmt.Errorf("launch editor: %w", err)
	}
	if !changed {
		fmt.Fprintln(cmd.ErrOrStderr(), "Aborted.")
		return nil
	}

	parsed, err := editor.Parse(edited)
	if err != nil {
		return UserErrorf("invalid buffer: %v", err)
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

	inserted, err := s.Add(storage.Entry{
		Title:       parsed.Title,
		Description: parsed.Description,
		Tags:        parsed.Tags,
		Project:     parsed.Project,
		Type:        parsed.Type,
		Impact:      parsed.Impact,
	})
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
