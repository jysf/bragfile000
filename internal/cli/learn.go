package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jysf/bragfile000/internal/capture"
	"github.com/jysf/bragfile000/internal/config"
	"github.com/jysf/bragfile000/internal/editor"
	"github.com/jysf/bragfile000/internal/storage"
)

// FailureType is the reserved entries.type value marking work that did not
// work (DEC-049). It is the ONE type value bragfile pins: `brag add --type`
// stays free-form, and `brag learn` exists so this value cannot fragment the
// way `shipped`/`ship` and `fixed`/`bugfix` already have in the live corpus.
const FailureType = "failed"

// learnFieldFlags are the entry-field flags whose presence routes `brag
// learn` to flag mode. "type" is deliberately ABSENT: the value is pinned,
// not chosen (DEC-049).
var learnFieldFlags = []string{"title", "description", "tags", "project", "impact"}

// NewLearnCmd builds `brag learn` — the capture verb for work that did not
// work. It is `brag add` with entries.type pinned to FailureType and the
// milestone nudge suppressed; see runLearn for why the nudge is dropped.
//
// Two modes, mirroring add's (DEC-007 / DEC-009) minus JSON mode:
//   - flag mode: any of the five entry-field flags set; --title is required.
//   - editor mode: no entry-field flag set; opens $EDITOR on a template whose
//     header block omits Type, because Type is not the user's to set here.
func NewLearnCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "learn",
		Short: "Record work that did not work, as a first-class entry",
		Long: `Record work that did not work — a dead end, an abandoned approach, a wrong
call — as a first-class entry rather than contorting it into a win.

Every entry brag learn writes is typed "failed". That is the point of the
verb: the value is pinned, so these stay greppable instead of fragmenting
across near-duplicate spellings the way free-form types do. There is no type
flag here. Use brag add if you want to choose the type yourself.

Flag mode (any of -t, -d, -T, -p, -i set): inserts directly from flag values.
--title is required.

Editor mode (no entry-field flags set): opens $EDITOR on a template buffer.
Save a valid entry to insert it; save unchanged to abort cleanly. The buffer
has no Type header — brag learn sets it for you.

Read them back the same way you read anything else:
  brag list --type failed
  brag list --type failed --since 90d

Examples:
  brag learn                                        # editor mode
  brag learn -t "shared-worker pool did not cut cold starts"
  brag learn -t "tried a bloom filter on the tag join" \
             -i "cost two days and produced nothing reusable"
  brag learn --title "..." --description "..." --tags "..." \
             --project "..." --impact "..."

Short forms: -t title, -d description, -T tags, -p project, -i impact.`,
		RunE: runLearn,
	}
	cmd.Flags().StringP("title", "t", "", "short headline (required in flag mode)")
	cmd.Flags().StringP("description", "d", "", "free-form body — what you tried, and why it did not work")
	cmd.Flags().StringP("tags", "T", "", "comma-joined tag list (e.g. \"auth,perf\")")
	cmd.Flags().StringP("project", "p", "", "project / initiative this belongs to")
	cmd.Flags().StringP("impact", "i", "", "what it cost, or what it ruled out")
	return cmd
}

// runLearn dispatches to flag mode or editor mode. There is no JSON mode:
// `brag add --json` with "type":"failed" already covers programmatic capture,
// and a second JSON ingress would be a second place to police the pinned
// value (DEC-049).
func runLearn(cmd *cobra.Command, args []string) error {
	for _, name := range learnFieldFlags {
		if cmd.Flags().Changed(name) {
			return runLearnFlags(cmd, args)
		}
	}
	return runLearnEditor(cmd)
}

func runLearnFlags(cmd *cobra.Command, _ []string) error {
	title := getFlagString(cmd, "title")
	if strings.TrimSpace(title) == "" {
		return UserErrorf("--title is required and must not be empty")
	}
	return insertLearned(cmd, capture.Fields{
		Title:       title,
		Description: getFlagString(cmd, "description"),
		Tags:        getFlagString(cmd, "tags"),
		Project:     getFlagString(cmd, "project"),
		Type:        FailureType,
		Impact:      getFlagString(cmd, "impact"),
	}, cmd.Flags().Changed("project"))
}

func runLearnEditor(cmd *cobra.Command) error {
	editFn := testEditFunc
	if editFn == nil {
		editFn = editor.Default
	}
	edited, changed, err := editor.Launch(editor.FailureTemplate(), editFn)
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
	// Type is overwritten, not read. The template omits the header, but a
	// user who adds one back does not get to redirect the pinned value.
	return insertLearned(cmd, capture.Fields{
		Title:       parsed.Title,
		Description: parsed.Description,
		Tags:        parsed.Tags,
		Project:     parsed.Project,
		Type:        FailureType,
		Impact:      parsed.Impact,
	}, parsed.Project != "")
}

// insertLearned validates, opens the store, auto-fills the project and
// inserts. It deliberately does NOT call emitMilestone: every milestone line
// is a congratulation ("🎉 N brags and counting — nice work!", "🔥 N-day
// streak!"), and firing one at the moment a user records a failure is the
// flattery this verb exists to remove. Silence on stderr, the id on stdout.
func insertLearned(cmd *cobra.Command, f capture.Fields, projectSet bool) error {
	if err := capture.Validate(f); err != nil {
		return UserErrorf("%v", err)
	}
	dbFlag := getFlagString(cmd, "db")
	path, err := config.ResolveDBPath(dbFlag)
	if err != nil {
		return fmt.Errorf("resolve db path: %w", err)
	}
	s, err := storage.Open(path)
	if err != nil {
		return err
	}
	defer s.Close()

	inserted, err := s.Add(storage.Entry{
		Title:       f.Title,
		Description: f.Description,
		Tags:        f.Tags,
		Project:     autoFillProject(s, f.Project, projectSet),
		Type:        f.Type,
		Impact:      f.Impact,
	})
	if err != nil {
		return fmt.Errorf("add entry: %w", err)
	}
	fmt.Fprintln(cmd.OutOrStdout(), inserted.ID)
	return nil
}
