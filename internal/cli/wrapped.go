package cli

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/jysf/bragfile000/internal/config"
	"github.com/jysf/bragfile000/internal/export"
	"github.com/jysf/bragfile000/internal/storage"
	"github.com/spf13/cobra"
)

// yearArgPattern matches a 4-digit year token. Plausibility bounds
// ([2000,2999]) are checked separately after the shape match (DEC-030
// choice 1 / LD3).
var yearArgPattern = regexp.MustCompile(`^\d{4}$`)

// lookupSparkEnv reads the NO_COLOR opt-out (no-color.org). Package var
// so tests inject it deterministically without touching the real env
// (AGENTS.md §9 os-state-via-var; DEC-023 LD6 precedent).
var lookupSparkEnv = os.LookupEnv

// NewWrappedCmd returns the `brag wrapped` subcommand (SPEC-051), the
// fifth DEC-014 consumer: a shareable, celebratory year- or quarter-in-
// review digest over a named calendar period. It curates the existing
// internal/aggregate toolbox into a retrospective highlight reel.
func NewWrappedCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wrapped [year] [Qn]",
		Short: "Shareable year- or quarter-in-review digest for a named calendar period",
		Long: `Print a shareable, celebratory year- or quarter-in-review digest for a named calendar period: your year in brags. Rule-based, deterministic, no LLM.

The period is named positionally. With no argument it covers the current calendar year; a 4-digit year covers that whole calendar year; a year plus a quarter token covers that calendar quarter (case-insensitive):
  brag wrapped            the current calendar year
  brag wrapped 2026       calendar year 2026
  brag wrapped 2026 Q3    calendar quarter Q3 2026 (Jul-Sep)

The window is bounded on both ends: a named period covers only entries created within it, so a completed year or quarter does not spill past its end. This differs from brag impact, which reports the current period up to now.

Output is markdown (default) or a single-object JSON envelope (--format json). The digest renders these sections: Cadence (busiest month + per-month counts), Top initiatives, Impact moments, Rhythm (longest streak, top tags, top types), and Span. Filter flags --tag/--project/--type compose with the period.

Examples:
  brag wrapped                                # the current calendar year, markdown
  brag wrapped 2026 --format json            # calendar year 2026, JSON envelope
  brag wrapped 2026 Q3 --project alpha       # one quarter, one initiative`,
		RunE: runWrapped,
	}
	cmd.Flags().String("format", "markdown", "output format (one of: markdown, json)")
	cmd.Flags().String("tag", "", "filter to entries whose tags contain this token")
	cmd.Flags().String("project", "", "filter to entries with this project (exact match)")
	cmd.Flags().String("type", "", "filter to entries with this type (exact match)")
	cmd.Flags().Bool("no-spark", false, "suppress the in-terminal cadence sparkline")
	return cmd
}

// parseWrappedPeriod resolves the positional period args into the scope
// token, the ordered "YYYY-MM" month labels in scope, and the bounded
// window [start, nextBoundary) — the upper edge is EXCLUSIVE so the CLI
// filters created_at < nextBoundary without a last-second off-by-one
// (DEC-030 choice 3). All period math uses time.Date calendar
// constructors, never day subtraction (inheriting DEC-028's rule).
//
//   - 0 args → current calendar year (now.Year()).
//   - 1 arg  → a 4-digit year in [2000,2999]; anything else is a UserError.
//   - 2 args → year + quarter token Q<1..4> (case-insensitive).
//   - 3+ args → UserError (extra tokens).
func parseWrappedPeriod(args []string, now time.Time) (scope string, months []string, start, nextBoundary time.Time, err error) {
	switch len(args) {
	case 0:
		year := now.Year()
		start = time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
		nextBoundary = start.AddDate(1, 0, 0)
		return fmt.Sprintf("%04d", year), monthLabels(year, 1, 12), start, nextBoundary, nil
	case 1:
		year, perr := parseYearArg(args[0])
		if perr != nil {
			return "", nil, time.Time{}, time.Time{}, perr
		}
		start = time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
		nextBoundary = start.AddDate(1, 0, 0)
		return fmt.Sprintf("%04d", year), monthLabels(year, 1, 12), start, nextBoundary, nil
	case 2:
		year, perr := parseYearArg(args[0])
		if perr != nil {
			return "", nil, time.Time{}, time.Time{}, perr
		}
		quarter, perr := parseQuarterArg(args[1])
		if perr != nil {
			return "", nil, time.Time{}, time.Time{}, perr
		}
		startMonth := (quarter-1)*3 + 1
		start = time.Date(year, time.Month(startMonth), 1, 0, 0, 0, 0, time.UTC)
		nextBoundary = start.AddDate(0, 3, 0)
		scope = fmt.Sprintf("%04d-Q%d", year, quarter)
		return scope, monthLabels(year, startMonth, 3), start, nextBoundary, nil
	default:
		return "", nil, time.Time{}, time.Time{}, UserErrorf("too many arguments for wrapped: expected [<year>] [Q<n>], got %d", len(args))
	}
}

// parseYearArg validates a 4-digit, plausibility-bounded year (LD3).
func parseYearArg(arg string) (int, error) {
	if !yearArgPattern.MatchString(arg) {
		return 0, UserErrorf("invalid year %q (expected a 4-digit year, e.g. 2026)", arg)
	}
	year, _ := time.Parse("2006", arg) // shape already validated by the regexp
	y := year.Year()
	if y < 2000 || y > 2999 {
		return 0, UserErrorf("year %q out of range (expected 2000-2999)", arg)
	}
	return y, nil
}

// parseQuarterArg validates a case-insensitive Q<1..4> token (LD3).
func parseQuarterArg(arg string) (int, error) {
	up := strings.ToUpper(arg)
	if len(up) != 2 || up[0] != 'Q' {
		return 0, UserErrorf("invalid quarter %q (expected Q1-Q4)", arg)
	}
	switch up[1] {
	case '1', '2', '3', '4':
		return int(up[1] - '0'), nil
	default:
		return 0, UserErrorf("invalid quarter %q (expected Q1-Q4)", arg)
	}
}

// monthLabels returns count consecutive "YYYY-MM" labels starting at
// (year, startMonth), rolling the year forward via time.Date so a
// quarter never crosses a year boundary here (quarters are 3 months
// inside one year) but the helper stays general.
func monthLabels(year, startMonth, count int) []string {
	labels := make([]string, 0, count)
	cursor := time.Date(year, time.Month(startMonth), 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < count; i++ {
		labels = append(labels, cursor.Format("2006-01"))
		cursor = cursor.AddDate(0, 1, 0)
	}
	return labels
}

func runWrapped(cmd *cobra.Command, args []string) error {
	now := nowFunc()

	scope, months, start, nextBoundary, err := parseWrappedPeriod(args, now)
	if err != nil {
		return err
	}

	format, _ := cmd.Flags().GetString("format")
	if format != "markdown" && format != "json" {
		return UserErrorf("unknown --format value %q (accepted: markdown, json)", format)
	}

	filter := storage.ListFilter{Since: start}
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

	all, err := s.List(filter)
	if err != nil {
		return fmt.Errorf("list entries: %w", err)
	}

	// Bounded-window upper edge (DEC-030 choice 3): ListFilter.Since is the
	// SQL lower bound; the exclusive created_at < nextBoundary upper filter
	// runs here in Go so no-sql-in-cli-layer stays intact (ListFilter has no
	// Until field). The period end is NOT "now" — this is the load-bearing
	// divergence from impact's [cutoff, now].
	entries := make([]storage.Entry, 0, len(all))
	for _, e := range all {
		if e.CreatedAt.Before(nextBoundary) {
			entries = append(entries, e)
		}
	}

	// Sparkline is default-on in markdown, suppressed by --no-spark OR a
	// present NO_COLOR env var. NO_COLOR uses "present-at-all" semantics
	// (no-color.org): set to ANY value (including empty) opts out. Either
	// signal alone suffices (DEC-031 LD7/LD8).
	noSpark, _ := cmd.Flags().GetBool("no-spark")
	_, noColorSet := lookupSparkEnv("NO_COLOR")
	sparkOn := !noSpark && !noColorSet

	filtersMD, filtersJSON := echoFiltersForWrapped(cmd)
	opts := export.WrappedOptions{
		Scope:       scope,
		Filters:     filtersMD,
		FiltersJSON: filtersJSON,
		ScopeMonths: months,
		Now:         now,
		Spark:       sparkOn,
	}

	var body []byte
	switch format {
	case "markdown":
		body, err = export.ToWrappedMarkdown(entries, opts)
	case "json":
		body, err = export.ToWrappedJSON(entries, opts)
	}
	if err != nil {
		return fmt.Errorf("render wrapped: %w", err)
	}

	fmt.Fprintln(cmd.OutOrStdout(), string(body))
	return nil
}

// echoFiltersForWrapped returns both the markdown filters line and the
// JSON filters object in one pass over wrapped's three-flag set (tag,
// project, type). Empty result → "(none)" + empty map. Kept local,
// mirroring echoFiltersForImpact — the third-caller lift threshold
// SPEC-018 set applies to the shared-helper decision, and DEC-028
// recorded lifting this as a Rejected alternative.
func echoFiltersForWrapped(cmd *cobra.Command) (string, map[string]string) {
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
