package cli

import (
	"fmt"
	"time"

	"github.com/jysf/bragfile000/internal/config"
	"github.com/jysf/bragfile000/internal/export"
	"github.com/jysf/bragfile000/internal/storage"
	"github.com/spf13/cobra"
)

// NewSparkCmd returns the `brag spark` subcommand (SPEC-059), the seventh
// DEC-014 consumer: a sparklines-only "pulse" of recent activity — a Total row
// plus up to the top-8 by-project rows, each a per-row min→max sparkline over a
// ROLLING recent window (DEC-037). Windows are rolling, not calendar
// (--week/--month/--quarter, mutually exclusive, default --month); --project is
// a row SELECTOR, not a corpus filter. Reuses the spark.Line primitive
// (DEC-031) and the shared lookupSparkEnv escape.
func NewSparkCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "spark",
		Short: "Sparklines-only pulse of recent activity (Total + by-project) over a rolling window",
		Long: `Print a sparklines-only pulse of your recent activity: a Total row plus a row per project, each a compact sparkline over a rolling recent window with the entry count in parentheses. Rule-based, deterministic, no LLM. A quick "what does my last stretch look like?" glance, not a full digest.

The window is ROLLING (it ends now), not a calendar period. Exactly one window may be given; they are mutually exclusive and default to --month:
  --week      the last 7 days, one bar per day (7 bars)
  --month     the last 28 days, one bar per week (4 bars, default)
  --quarter   the last 91 days, one bar per week (13 bars)

Every row is bucketed over the same axis, so bar position lines up across rows. Each row's sparkline is scaled to its own min-max shape; the count in parentheses carries the magnitude, so a steady project and a bursty one are both legible. By default the Total row plus the top 8 projects by entry count are shown; --project <name> narrows the by-project rows to that one project (the Total still spans everything).

Output is markdown (default) or a single-object JSON envelope (--format json) per DEC-014. The sparkline is markdown-only; JSON carries raw per-bucket counts (--format json | jq .total.series). --no-spark, or a present NO_COLOR env var, drops the glyphs and prints raw per-bucket counts instead.

Examples:
  brag spark                                  # last 28 days, weekly bars, markdown
  brag spark --week                           # last 7 days, one bar per day
  brag spark --quarter --project alpha        # 13 weekly bars: Total vs alpha
  brag spark --month --format json            # raw per-bucket counts, JSON envelope`,
		RunE: runSpark,
	}
	cmd.Flags().Bool("week", false, "pulse over the last 7 days (7 daily bars)")
	cmd.Flags().Bool("month", false, "pulse over the last 28 days (4 weekly bars; the default)")
	cmd.Flags().Bool("quarter", false, "pulse over the last 91 days (13 weekly bars)")
	cmd.Flags().String("project", "", "show only this project's row alongside Total (row selector, not a corpus filter)")
	cmd.Flags().String("format", "markdown", "output format (one of: markdown, json)")
	cmd.Flags().Bool("no-spark", false, "suppress glyphs; print raw per-bucket counts instead")
	return cmd
}

// sparkWindow maps a rolling-window token to its (bucket width, bucket count)
// axis per DEC-037: --week = 24h*7, --month = 7d*4, --quarter = 7d*13.
func sparkWindow(scope string) (width time.Duration, n int) {
	day := 24 * time.Hour
	switch scope {
	case "week":
		return day, 7
	case "quarter":
		return 7 * day, 13
	default: // "month"
		return 7 * day, 4
	}
}

func runSpark(cmd *cobra.Command, _ []string) error {
	// Sample the clock ONCE and truncate to the second so the query boundary
	// (storage stores/compares created_at at RFC3339 second precision) and the
	// bucket axis agree exactly — one axis feeds both the filter and the
	// bucketer below (SPEC-060).
	now := nowFunc().Truncate(time.Second)

	// Resolve the rolling window: exactly-one-or-none among week/month/quarter,
	// mutually exclusive, default month. (window.go's selectedWindow is for the
	// CALENDAR flag set — not reused here; DEC-037 rolling core.)
	scope, err := selectedSparkWindow(cmd)
	if err != nil {
		return err
	}
	width, n := sparkWindow(scope)

	format, _ := cmd.Flags().GetString("format")
	if format != "markdown" && format != "json" {
		return UserErrorf("unknown --format value %q (accepted: markdown, json)", format)
	}

	// --project is the ROW SELECTOR (DEC-037 choice 3), NOT a corpus filter: it
	// is threaded into SparkOptions, never onto ListFilter.
	var project string
	if cmd.Flags().Changed("project") {
		project, _ = cmd.Flags().GetString("project")
		if project == "" {
			return UserErrorf("--project must not be empty")
		}
	}

	// Query the in-window corpus once over the SAME half-open [start, now) axis
	// the bucketer uses. start = now - width*n is aggregate.RollingBuckets's
	// axis start; Until = now is its exclusive upper edge (the DEC-035 field, set
	// per that decision's "fifth bounded consumer" guidance). Bounding BOTH ends
	// keeps the markdown header count (len(entries)) and the ByProject top-8
	// selection aligned with the bucket sums, so an out-of-window row (clock
	// skew / import / multi-machine) can't inflate the header or take a top-8
	// slot (SPEC-060).
	start := now.Add(-width * time.Duration(n))
	filter := storage.ListFilter{Since: start, Until: now}

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

	// Sparkline is default-on in markdown, suppressed by --no-spark OR a present
	// NO_COLOR env var (no-color.org present-at-all semantics), via the SHARED
	// lookupSparkEnv (declared in wrapped.go) — DEC-031.
	noSpark, _ := cmd.Flags().GetBool("no-spark")
	_, noColorSet := lookupSparkEnv("NO_COLOR")
	sparkOn := !noSpark && !noColorSet

	opts := export.SparkOptions{
		Scope:   scope,
		Now:     now,
		Width:   width,
		Buckets: n,
		Project: project,
		Spark:   sparkOn,
	}

	var body []byte
	switch format {
	case "markdown":
		body, err = export.ToSparkMarkdown(entries, opts)
	case "json":
		body, err = export.ToSparkJSON(entries, opts)
	}
	if err != nil {
		return fmt.Errorf("render spark: %w", err)
	}

	fmt.Fprintln(cmd.OutOrStdout(), string(body))
	return nil
}

// selectedSparkWindow returns the rolling-window token among the mutually
// exclusive --week/--month/--quarter flags, defaulting to "month" when none is
// set. Two-plus set is a UserError. Local to spark (the rolling core, DEC-037);
// window.go's selectedWindow governs the calendar flag set instead.
func selectedSparkWindow(cmd *cobra.Command) (string, error) {
	var set []string
	for _, name := range []string{"week", "month", "quarter"} {
		if v, _ := cmd.Flags().GetBool(name); v {
			set = append(set, name)
		}
	}
	switch len(set) {
	case 0:
		return "month", nil
	case 1:
		return set[0], nil
	default:
		return "", UserErrorf("--week, --month, --quarter are mutually exclusive")
	}
}
