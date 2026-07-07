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

// NewCoverageCmd returns the `brag coverage` subcommand (SPEC-045), the sixth
// DEC-014 consumer: a rule-based, calendar-windowed digest of the agent- vs
// human-authored provenance share over time (DEC-024/DEC-033). It reuses the
// shared calendar-window core (DEC-028) and the --previous modifier (DEC-032),
// classifies via aggregate.IsAgentAuthored (single-sourced with storage's SQL
// clause, SPEC-045 LD2), and renders a per-month agent-share sparkline
// (DEC-031).
func NewCoverageCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "coverage",
		Short: "Rule-based agent-vs-human provenance-share digest over a calendar period",
		Long: `Print a rule-based coverage digest: how much of your work over a calendar reporting period was agent-authored vs human-authored, and how that share is trending month by month. Provenance is read from the reserved agent:/model: tags the MCP write path stamps (DEC-024). No LLM.

Output is markdown (default) or a single-object JSON envelope (--format json) per DEC-014/DEC-033. Exactly one window is required and the windows are mutually exclusive:
  --quarter   the current calendar quarter (Jan-Mar / Apr-Jun / Jul-Sep / Oct-Dec), up to now
  --month     the current calendar month, up to now
  --year      the current calendar year, up to now
  --since D   entries on or after D (YYYY-MM-DD or Nd/Nw/Nm), up to now

--previous shifts the selected window to the last-completed period (bounded on both ends). It requires a window flag and is incompatible with --since. Windows are CALENDAR periods, not rolling.

The monthly trend is a per-month agent-share sparkline (markdown only; --no-spark or NO_COLOR suppresses it). Filter flags --tag/--project/--type compose with the window.

Examples:
  brag coverage --year                               # this calendar year, markdown
  brag coverage --quarter --previous                 # the whole previous quarter
  brag coverage --since 2026-01-01 --format json     # since a date, JSON envelope`,
		RunE: runCoverage,
	}
	cmd.Flags().Bool("quarter", false, "coverage for the current calendar quarter")
	cmd.Flags().Bool("month", false, "coverage for the current calendar month")
	cmd.Flags().Bool("year", false, "coverage for the current calendar year")
	cmd.Flags().String("since", "", "coverage since a date (YYYY-MM-DD or Nd/Nw/Nm)")
	cmd.Flags().Bool("previous", false, "shift the window to the last-completed period")
	cmd.Flags().String("format", "markdown", "output format (one of: markdown, json)")
	cmd.Flags().String("tag", "", "filter to entries whose tags contain this token")
	cmd.Flags().String("project", "", "filter to entries with this project (exact match)")
	cmd.Flags().String("type", "", "filter to entries with this type (exact match)")
	cmd.Flags().Bool("no-spark", false, "suppress the in-terminal agent-share sparkline")
	return cmd
}

func runCoverage(cmd *cobra.Command, _ []string) error {
	now := nowFunc()

	window, err := selectedWindow(cmd)
	if err != nil {
		return err
	}
	previous, _ := cmd.Flags().GetBool("previous")
	if previous && window == "since" {
		return UserErrorf("--previous cannot be combined with --since (--since is an explicit anchor, not a calendar period)")
	}
	sinceRaw, _ := cmd.Flags().GetString("since")
	cutoff, end, scope, err := windowCutoff(window, sinceRaw, now, previous)
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

	// Coverage reads all in-window rows once and classifies in Go (it needs
	// BOTH classes to compute a share; SPEC-045 LD7) — it does NOT set
	// ListFilter.Author.
	entries, err := s.List(filter)
	if err != nil {
		return fmt.Errorf("list entries: %w", err)
	}

	// Bounded-window upper edge for --previous (DEC-032 choice 1): a non-zero
	// end is the current-period start (the exclusive upper bound of the
	// last-completed period). The created_at < end filter runs here in Go so
	// no-sql-in-cli-layer stays intact. A zero end (the current-period path)
	// skips the filter, preserving [cutoff, now].
	if !end.IsZero() {
		bounded := entries[:0]
		for _, e := range entries {
			if e.CreatedAt.Before(end) {
				bounded = append(bounded, e)
			}
		}
		entries = bounded
	}

	// Derive the ordered "YYYY-MM" labels covering the window so the monthly
	// series is always fully present (zero-filled), even on an empty window.
	// The upper edge is `end` when non-zero (--previous), else `now`; for
	// --previous the last in-scope month is the month before `end`, so step
	// back a nanosecond off the exclusive boundary.
	upper := now
	if !end.IsZero() {
		upper = end.Add(-time.Nanosecond)
	}
	scopeMonths := monthLabelsBetween(cutoff, upper)

	// Sparkline is default-on in markdown, suppressed by --no-spark OR a
	// present NO_COLOR env var (no-color.org present-at-all semantics). Mirrors
	// wrapped's spark-escape posture (DEC-031).
	noSpark, _ := cmd.Flags().GetBool("no-spark")
	_, noColorSet := lookupSparkEnv("NO_COLOR")
	sparkOn := !noSpark && !noColorSet

	filtersMD, filtersJSON := echoFiltersForCoverage(cmd)
	opts := export.CoverageOptions{
		Scope:       scope,
		Filters:     filtersMD,
		FiltersJSON: filtersJSON,
		ScopeMonths: scopeMonths,
		Now:         now,
		Spark:       sparkOn,
	}

	var body []byte
	switch format {
	case "markdown":
		body, err = export.ToCoverageMarkdown(entries, opts)
	case "json":
		body, err = export.ToCoverageJSON(entries, opts)
	}
	if err != nil {
		return fmt.Errorf("render coverage: %w", err)
	}

	fmt.Fprintln(cmd.OutOrStdout(), string(body))
	return nil
}

// monthLabelsBetween returns the ordered "YYYY-MM" labels from start's month
// to upperInclusive's month, inclusive, stepping one calendar month at a time
// (time.Date / AddDate, never day subtraction). Coverage-local: wrapped's
// monthLabels takes an explicit count from a known year/quarter, whereas
// coverage derives labels from an arbitrary [cutoff, upper] window (since/
// previous make the span variable), so a shared lift is premature (the
// third-caller threshold, SPEC-018; SPEC-045 Rejected alternative). If upper
// precedes start (a degenerate window), returns an empty slice.
func monthLabelsBetween(start, upperInclusive time.Time) []string {
	cursor := time.Date(start.UTC().Year(), start.UTC().Month(), 1, 0, 0, 0, 0, time.UTC)
	endMonth := time.Date(upperInclusive.UTC().Year(), upperInclusive.UTC().Month(), 1, 0, 0, 0, 0, time.UTC)
	labels := []string{}
	for !cursor.After(endMonth) {
		labels = append(labels, cursor.Format("2006-01"))
		cursor = cursor.AddDate(0, 1, 0)
	}
	return labels
}

// echoFiltersForCoverage returns both the markdown filters line and the JSON
// filters object in one pass over coverage's three-flag set (tag, project,
// type). Empty result → "(none)" + empty map. Kept local, mirroring
// echoFiltersForImpact/echoFiltersForWrapped (the third-caller lift threshold
// applies to the shared-helper decision; DEC-028 recorded lifting it as a
// Rejected alternative).
func echoFiltersForCoverage(cmd *cobra.Command) (string, map[string]string) {
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
