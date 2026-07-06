package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/jysf/bragfile000/internal/config"
	"github.com/jysf/bragfile000/internal/export"
	"github.com/jysf/bragfile000/internal/storage"
	"github.com/spf13/cobra"
)

// nowFunc is the injectable wall-clock seam (AGENTS.md §9). runImpact
// calls it once and threads the result into both windowCutoff and
// ImpactOptions.Now so the calendar math and the Generated: line share
// a single instant. Tests substitute a fixed instant (see impact_test.go).
var nowFunc = func() time.Time { return time.Now().UTC() }

// NewImpactCmd returns the `brag impact` subcommand (SPEC-048), the
// fourth DEC-014 consumer: a rule-based, calendar-windowed, initiative-
// grouped digest of entries carrying an impact statement.
func NewImpactCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "impact",
		Short: "Rule-based impact digest for a calendar reporting period",
		Long: `Print a rule-based impact digest for a calendar reporting period: the entries that carry an impact statement, grouped by initiative (project), with each impact rendered in full. No LLM.

Output is markdown (default) or a single-object JSON envelope (--format json) per DEC-014/DEC-028. Exactly one window is required and the windows are mutually exclusive:
  --quarter   the current calendar quarter (Jan-Mar / Apr-Jun / Jul-Sep / Oct-Dec), up to now
  --month     the current calendar month, up to now
  --year      the current calendar year, up to now
  --since D   entries on or after D (YYYY-MM-DD or Nd/Nw/Nm), up to now

Windows are CALENDAR periods, not rolling — this differs from brag summary on purpose (the story surface reports by quarter/month/year). Only entries with a non-empty impact appear in the body; the provenance line tallies how many in-window entries had one. Filter flags --tag/--project/--type compose with the window.

Examples:
  brag impact --quarter                              # this calendar quarter, markdown
  brag impact --year --format json                   # this calendar year, JSON envelope
  brag impact --since 2026-01-01 --project alpha     # since a date, one initiative`,
		RunE: runImpact,
	}
	cmd.Flags().Bool("quarter", false, "impact for the current calendar quarter")
	cmd.Flags().Bool("month", false, "impact for the current calendar month")
	cmd.Flags().Bool("year", false, "impact for the current calendar year")
	cmd.Flags().String("since", "", "impact since a date (YYYY-MM-DD or Nd/Nw/Nm)")
	cmd.Flags().String("format", "markdown", "output format (one of: markdown, json)")
	cmd.Flags().String("tag", "", "filter to entries whose tags contain this token")
	cmd.Flags().String("project", "", "filter to entries with this project (exact match)")
	cmd.Flags().String("type", "", "filter to entries with this type (exact match)")
	return cmd
}

// windowCutoff computes the inclusive lower bound and scope token for
// the selected window. Pure and deterministic given (window, sinceRaw,
// now). CALENDAR periods (time.Date constructors), NEVER day
// subtraction — this is the correctness core of the calendar-vs-rolling
// divergence (DEC-028 choice 1). now is UTC; the period end is always
// "now" (implicit — every stored created_at <= now, so the lower bound
// alone bounds the window).
func windowCutoff(window, sinceRaw string, now time.Time) (cutoff time.Time, scope string, err error) {
	switch window {
	case "quarter":
		qStartMonth := ((int(now.Month())-1)/3)*3 + 1 // 1, 4, 7, 10
		cutoff = time.Date(now.Year(), time.Month(qStartMonth), 1, 0, 0, 0, 0, time.UTC)
		return cutoff, "quarter", nil
	case "month":
		cutoff = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		return cutoff, "month", nil
	case "year":
		cutoff = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
		return cutoff, "year", nil
	case "since":
		cutoff, err = ParseSince(sinceRaw)
		if err != nil {
			return time.Time{}, "", UserErrorf("invalid --since value: %v", err)
		}
		return cutoff, "since:" + sinceRaw, nil
	default:
		return time.Time{}, "", fmt.Errorf("windowCutoff: unhandled window %q", window)
	}
}

// selectedWindow returns the single set window flag's canonical name,
// or a UserError if zero or two-plus are set (mutually exclusive +
// required, DEC-028 choice 1 / DEC-007). Cobra's
// MarkFlagsMutuallyExclusive handles pairs but not "exactly one
// required" cleanly across a bool+string mix and routes its error off
// the UserError→stderr path, so the check is explicit here.
func selectedWindow(cmd *cobra.Command) (string, error) {
	var set []string
	for _, name := range []string{"quarter", "month", "year", "since"} {
		if cmd.Flags().Changed(name) {
			set = append(set, name)
		}
	}
	switch len(set) {
	case 0:
		return "", UserErrorf("one of --quarter, --month, --year, --since is required")
	case 1:
		return set[0], nil
	default:
		return "", UserErrorf("--quarter, --month, --year, --since are mutually exclusive (got --%s)", strings.Join(set, ", --"))
	}
}

func runImpact(cmd *cobra.Command, _ []string) error {
	now := nowFunc()

	window, err := selectedWindow(cmd)
	if err != nil {
		return err
	}
	sinceRaw, _ := cmd.Flags().GetString("since")
	cutoff, scope, err := windowCutoff(window, sinceRaw, now)
	if err != nil {
		return err
	}

	format, _ := cmd.Flags().GetString("format")
	if format != "markdown" && format != "json" {
		return UserErrorf("unknown --format value %q (accepted: markdown, json)", format)
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

	filtersMD, filtersJSON := echoFiltersForImpact(cmd)
	opts := export.ImpactOptions{
		Scope:           scope,
		Filters:         filtersMD,
		FiltersJSON:     filtersJSON,
		EntriesInWindow: len(entries),
		Now:             now,
	}

	var body []byte
	switch format {
	case "markdown":
		body, err = export.ToImpactMarkdown(entries, opts)
	case "json":
		body, err = export.ToImpactJSON(entries, opts)
	}
	if err != nil {
		return fmt.Errorf("render impact: %w", err)
	}

	fmt.Fprintln(cmd.OutOrStdout(), string(body))
	return nil
}

// echoFiltersForImpact returns both the markdown filters line and the
// JSON filters object in one pass over impact's three-flag set (tag,
// project, type). Empty result → "(none)" + empty map. Kept local
// (mirrors echoFiltersForSummary): two callers with an identical
// three-flag set is still below the third-caller lift threshold
// SPEC-018 set; lifting a shared helper is a DEC-028 Rejected
// alternative.
func echoFiltersForImpact(cmd *cobra.Command) (string, map[string]string) {
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
