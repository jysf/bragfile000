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

// NewSummaryCmd returns the `brag summary` subcommand. SPEC-018 emits
// it as the first DEC-014 consumer: a rule-based digest of entries in
// a rolling 7- or 30-day window, rendered as markdown (default) or
// the DEC-014 JSON envelope.
func NewSummaryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "summary",
		Short: "Rule-based digest of entries in a rolling time window",
		Long: `Print a rule-based digest of entries in a rolling time window: counts by type and project plus grouped highlights (titles + IDs only). No LLM.

Output is markdown (default) or a single-object JSON envelope (--format json) per DEC-014. The JSON shape is intentionally an envelope (not the naked array DEC-011 uses for list/export), because aggregations carry per-document metadata.

--range is required: week = last 7 UTC days from time.Now(); month = last 30 UTC days. Rolling, NOT calendar. Filter flags --tag/--project/--type compose with the range.

Examples:
  brag summary --range week                          # last 7 UTC days, markdown
  brag summary --range month --format json           # last 30 UTC days, JSON envelope
  brag summary --range week --tag auth --project p   # compose filters`,
		RunE: runSummary,
	}
	cmd.Flags().String("range", "", "time range (required; one of: week, month)")
	cmd.Flags().String("format", "markdown", "output format (one of: markdown, json)")
	cmd.Flags().String("tag", "", "filter to entries whose tags contain this token (comma-separated match)")
	cmd.Flags().String("project", "", "filter to entries with this project (exact match)")
	cmd.Flags().String("type", "", "filter to entries with this type (exact match)")
	return cmd
}

// rangeCutoff returns the inclusive lower bound for the --range
// filter. Pure function: deterministic given (rangeFlag, now); takes
// `now` as input so callers can inject a fixed time in tests.
//
//	"week"  → now - 7 days
//	"month" → now - 30 days
//	""      → UserErrorf naming the flag.
//	other   → UserErrorf naming the offending value + accepted set.
//
// Locked at the unit layer per the spec's Rejected alternatives
// (build-time) §1: keeps the storage surface free of test-only
// methods.
func rangeCutoff(rangeFlag string, now time.Time) (time.Time, error) {
	switch rangeFlag {
	case "":
		return time.Time{}, UserErrorf("--range is required (accepted: week, month)")
	case "week":
		return now.AddDate(0, 0, -7), nil
	case "month":
		return now.AddDate(0, 0, -30), nil
	default:
		return time.Time{}, UserErrorf("unknown --range value %q (accepted: week, month)", rangeFlag)
	}
}

func runSummary(cmd *cobra.Command, _ []string) error {
	rangeFlag, _ := cmd.Flags().GetString("range")
	cutoff, err := rangeCutoff(rangeFlag, time.Now().UTC())
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

	filtersMD, filtersJSON := echoFiltersForSummary(cmd)
	opts := export.SummaryOptions{
		Scope:       rangeFlag,
		Filters:     filtersMD,
		FiltersJSON: filtersJSON,
		Now:         time.Now().UTC(),
	}

	var body []byte
	switch format {
	case "markdown":
		body, err = export.ToSummaryMarkdown(entries, opts)
	case "json":
		body, err = export.ToSummaryJSON(entries, opts)
	}
	if err != nil {
		return fmt.Errorf("render summary: %w", err)
	}

	fmt.Fprintln(cmd.OutOrStdout(), string(body))
	return nil
}

// echoFiltersForSummary returns both the markdown filters line and
// the JSON filters object in one pass over summary's three-flag set
// (tag, project, type). Empty result → "(none)" + empty map.
//
// Inlined here, not lifted from export.go's echoFilters: that helper
// iterates a five-flag superset (tag/project/type/since/limit) and
// reusing it would couple summary to export's flag set. Three callers
// with the same flag set is the threshold for shared code; we're at
// two callers with divergent sets today.
func echoFiltersForSummary(cmd *cobra.Command) (string, map[string]string) {
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
