package cli

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jysf/bragfile000/internal/config"
	"github.com/jysf/bragfile000/internal/storage"
	"github.com/jysf/bragfile000/internal/story"
	"github.com/spf13/cobra"
)

// storyNowFunc is the injectable wall-clock seam (AGENTS.md §9), mirroring
// impact's nowFunc. runStory calls it once and threads the instant into
// both the window math and the bundle's Generated: line.
var storyNowFunc = func() time.Time { return time.Now().UTC() }

// NewStoryCmd returns the `brag story` subcommand (SPEC-049): the first
// narrative surface. It coalesces the in-window corpus into deterministic
// threads, assembles a throughline skeleton, shapes selection/altitude per
// a data-driven audience profile, and emits an arc-aware bundle (markdown
// default / JSON envelope) with a per-audience framing directive appended.
// No model, no network in the binary — the bundle is a complete artifact
// standalone; the LLM (already in the caller) is the optional upgrade.
func NewStoryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "story",
		Short: "Audience-shaped narrative bundle: brags coalesced into threads for an LLM to weave",
		Long: `Emit an audience-shaped narrative bundle: your brags coalesced into threads (initiatives, time-ordered, with impact beats marked) plus a throughline skeleton and a per-audience framing directive. bragfile shapes the data; an LLM (already in your session, or a paste-in) writes the prose. No model, no network in the binary — the bundle is a complete, readable artifact on its own; the LLM is an optional upgrade.

--audience is required and selects a shaping profile (selection + threading + altitude + framing directive):
  me     candid reflection — every thread, the messy middle and lessons kept, low altitude
  exec   impact-forward promotion — impact-bearing threads only, one headline arc, terse

Each audience carries a default window; an explicit window flag overrides it. Windows are CALENDAR periods (like brag impact), mutually exclusive:
  --quarter / --month / --year / --since D   (D: YYYY-MM-DD or Nd/Nw/Nm)

Audiences are extensible profiles, not a fixed list: drop a <name>.yaml in your story-profiles override directory to add or reshape one.

Output is markdown (default) or a JSON envelope (--format json). --theme <tag> adds a cross-project thread for that tag. --print-directive prints only the audience's framing directive.

Examples:
  brag story --audience me                                   # candid, this year (me's default window)
  brag story --audience exec --quarter                       # exec, this calendar quarter
  brag story --audience exec --year --format json            # arc-aware JSON envelope
  brag story --audience me --theme perf                      # add a cross-project perf arc
  brag story --audience exec --print-directive               # just the framing directive`,
		RunE: runStory,
	}
	cmd.Flags().String("audience", "", "shaping profile (required; one of: me, exec, or a user profile)")
	cmd.Flags().Bool("quarter", false, "window: the current calendar quarter (overrides the profile default)")
	cmd.Flags().Bool("month", false, "window: the current calendar month (overrides the profile default)")
	cmd.Flags().Bool("year", false, "window: the current calendar year (overrides the profile default)")
	cmd.Flags().String("since", "", "window: entries since a date (YYYY-MM-DD or Nd/Nw/Nm)")
	cmd.Flags().String("theme", "", "add a cross-project thread grouping entries with this tag")
	cmd.Flags().String("format", "markdown", "output format (one of: markdown, json)")
	cmd.Flags().Bool("print-directive", false, "print only the audience's framing directive and exit")
	cmd.Flags().String("project", "", "filter to entries with this project (exact match)")
	cmd.Flags().String("type", "", "filter to entries with this type (exact match)")
	cmd.Flags().String("tag", "", "filter to entries whose tags contain this token")
	return cmd
}

// resolveWindow returns the (cutoff, scope) for story. An explicit window
// flag overrides the profile default; the window flags are mutually
// exclusive (reusing impact's selectedWindow). With no window flag set,
// the profile's DefaultWindow applies. selectedWindow's zero-flag
// UserError is avoided via windowFlagsSet — story supplies the default.
func resolveWindow(cmd *cobra.Command, profile story.Profile, now time.Time) (time.Time, string, error) {
	if windowFlagsSet(cmd) {
		window, err := selectedWindow(cmd)
		if err != nil {
			return time.Time{}, "", err
		}
		sinceRaw, _ := cmd.Flags().GetString("since")
		return windowCutoff(window, sinceRaw, now)
	}
	// No window flag → the profile default. A "since:<raw>" default carries
	// the raw after the colon; the plain tokens carry no sinceRaw.
	window := profile.DefaultWindow
	sinceRaw := ""
	if raw, ok := strings.CutPrefix(window, "since:"); ok {
		window, sinceRaw = "since", raw
	}
	return windowCutoff(window, sinceRaw, now)
}

func runStory(cmd *cobra.Command, _ []string) error {
	if !cmd.Flags().Changed("audience") {
		return UserErrorf("--audience is required (one of: me, exec, or a user profile)")
	}
	audience := getFlagString(cmd, "audience")
	if audience == "" {
		return UserErrorf("--audience must not be empty")
	}

	profile, err := story.LoadProfile(audience)
	if err != nil {
		if errors.Is(err, story.ErrProfileNotFound) {
			return UserErrorf("unknown audience %q (no bundled default and no override file)", audience)
		}
		return UserErrorf("load audience profile %q: %v", audience, err)
	}

	directive, err := story.ResolveDirective(profile)
	if err != nil {
		return fmt.Errorf("resolve framing directive: %w", err)
	}

	// --print-directive short-circuits: no window, no DB read, directive only.
	if changed, _ := cmd.Flags().GetBool("print-directive"); changed {
		fmt.Fprintln(cmd.OutOrStdout(), directive)
		return nil
	}

	format, _ := cmd.Flags().GetString("format")
	if format != "markdown" && format != "json" {
		return UserErrorf("unknown --format value %q (accepted: markdown, json)", format)
	}

	now := storyNowFunc()
	cutoff, scope, err := resolveWindow(cmd, profile, now)
	if err != nil {
		return err
	}

	filter := storage.ListFilter{Since: cutoff}
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

	theme := ""
	if cmd.Flags().Changed("theme") {
		theme, _ = cmd.Flags().GetString("theme")
		if theme == "" {
			return UserErrorf("--theme must not be empty")
		}
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

	threads := story.BuildThreads(entries, story.ThreadOptionsFromProfile(profile, theme))
	throughline := story.BuildThroughline(threads)

	filtersMD, filtersJSON := echoFiltersForStory(cmd)
	opts := story.StoryOptions{
		Audience:        audience,
		Scope:           scope,
		Filters:         filtersMD,
		FiltersJSON:     filtersJSON,
		EntriesInWindow: len(entries),
		Now:             now,
		Threads:         threads,
		Throughline:     throughline,
		Directive:       directive,
	}

	var body []byte
	switch format {
	case "markdown":
		body, err = story.ToStoryMarkdown(opts)
	case "json":
		body, err = story.ToStoryJSON(opts)
	}
	if err != nil {
		return fmt.Errorf("render story: %w", err)
	}

	fmt.Fprintln(cmd.OutOrStdout(), string(body))
	return nil
}

// echoFiltersForStory returns both the markdown filters line and the JSON
// filters object over story's three-flag set (tag, project, type). Empty →
// "(none)" + empty map. Mirrors echoFiltersForImpact (still below the
// third-caller lift threshold for the filter echo itself).
func echoFiltersForStory(cmd *cobra.Command) (string, map[string]string) {
	jsonObj := map[string]string{}
	var parts []string
	for _, name := range []string{"tag", "project", "type"} {
		if !cmd.Flags().Changed(name) {
			continue
		}
		v, _ := cmd.Flags().GetString(name)
		jsonObj[name] = v
		parts = append(parts, fmt.Sprintf("--%s %s", name, v))
	}
	if len(parts) == 0 {
		return "(none)", jsonObj
	}
	return strings.Join(parts, " "), jsonObj
}
