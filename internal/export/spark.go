package export

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jysf/bragfile000/internal/aggregate"
	"github.com/jysf/bragfile000/internal/spark"
	"github.com/jysf/bragfile000/internal/storage"
)

// SparkOptions controls the rule-based sparkline "pulse" digest (SPEC-059), the
// seventh DEC-014 consumer. Scope echoes the rolling-window token
// ("week"/"month"/"quarter"). Now is the injected wall clock (the axis end);
// Width and Buckets define the rolling axis passed to aggregate.RollingBuckets.
// Project, when non-empty, is a ROW SELECTOR (DEC-037 choice 3): the by-project
// rows collapse to that one project (rendered even at zero count), while the
// Total row still spans the whole in-window corpus. Empty Project → the top-8
// by-project auto-selection. Spark, when true and rendering markdown, prints the
// glyph rows; when false it falls back to raw per-bucket counts. JSON never
// carries glyphs — each row carries a raw series int array (DEC-031 choice f).
type SparkOptions struct {
	Scope   string
	Now     time.Time
	Width   time.Duration
	Buckets int
	Project string
	Spark   bool
}

// sparkRow is one rendered row (Total or a project): its label, the raw
// per-bucket counts, and their sum (the parens magnitude).
type sparkRow struct {
	Label  string
	Series []int
	Count  int
}

// sparkRows builds the ordered rows for a pulse: the Total row (whole in-window
// corpus) followed by the by-project rows. When opts.Project is set, exactly
// that one project's row is returned (even if it has zero in-window entries);
// otherwise the top-8 projects by in-window volume (aggregate.ByProject order:
// desc count, alpha tiebreak, (no project) last).
func sparkRows(entries []storage.Entry, opts SparkOptions) (total sparkRow, projects []sparkRow) {
	totalSeries := aggregate.RollingBuckets(entries, opts.Now, opts.Width, opts.Buckets)
	total = sparkRow{Label: "Total", Series: totalSeries, Count: sumInts(totalSeries)}

	// Group entries by project key ((no project) sentinel for empty).
	byKey := make(map[string][]storage.Entry)
	for _, e := range entries {
		key := e.Project
		if key == "" {
			key = aggregate.NoProjectKey
		}
		byKey[key] = append(byKey[key], e)
	}

	makeRow := func(label string) sparkRow {
		series := aggregate.RollingBuckets(byKey[label], opts.Now, opts.Width, opts.Buckets)
		return sparkRow{Label: label, Series: series, Count: sumInts(series)}
	}

	if opts.Project != "" {
		// Row selector: the one named project, rendered even at zero count.
		// A named project maps to (no project) only via the empty string, which
		// the CLI already rejects, so use the label as-is.
		return total, []sparkRow{makeRow(opts.Project)}
	}

	// Top-8 by in-window volume, in aggregate.ByProject order.
	for _, pc := range aggregate.ByProject(entries) {
		if len(projects) >= 8 {
			break
		}
		projects = append(projects, makeRow(pc.Project))
	}
	return total, projects
}

// ToSparkMarkdown renders the in-window entries as the sparkline pulse per
// DEC-014/DEC-037: header + provenance block, then ## Pulse with a Total row and
// the by-project rows. Each row is "<label> (<count>): <glyphs>" (or raw
// space-joined counts when opts.Spark is false). On an empty window only the
// header block (through "Entries: 0") is emitted; the ## Pulse body is omitted
// (DEC-014 part 4). Trailing newline stripped (matches every other renderer).
func ToSparkMarkdown(entries []storage.Entry, opts SparkOptions) ([]byte, error) {
	var buf bytes.Buffer
	fmt.Fprintln(&buf, "# Bragfile Spark")
	fmt.Fprintln(&buf)
	fmt.Fprintf(&buf, "Generated: %s\n", opts.Now.UTC().Format(time.RFC3339))
	fmt.Fprintf(&buf, "Scope: %s\n", opts.Scope)
	fmt.Fprintf(&buf, "Filters: %s\n", sparkFiltersMarkdown(opts))
	fmt.Fprintf(&buf, "Entries: %d\n", len(entries))

	if len(entries) == 0 {
		return trimTrailingNewline(buf.Bytes()), nil
	}

	total, projects := sparkRows(entries, opts)

	fmt.Fprintln(&buf)
	fmt.Fprintln(&buf, "## Pulse")
	fmt.Fprintln(&buf)
	writeSparkRow(&buf, total, opts.Spark)
	for _, r := range projects {
		writeSparkRow(&buf, r, opts.Spark)
	}

	return trimTrailingNewline(buf.Bytes()), nil
}

// writeSparkRow writes one "<label> (<count>): <cells>" line. cells is the
// glyph sparkline when spark is true, else the space-joined raw counts (so a
// suppressed sparkline still emits its data — DEC-037 choice 6).
func writeSparkRow(buf *bytes.Buffer, r sparkRow, spk bool) {
	var cells string
	if spk {
		cells = spark.Line(r.Series)
	} else {
		cells = joinInts(r.Series)
	}
	fmt.Fprintf(buf, "%s (%d): %s\n", r.Label, r.Count, cells)
}

// sparkSeriesRecord is one JSON row: project label, sum count, and the raw
// per-bucket series (no glyphs — DEC-031 choice f).
type sparkSeriesRecord struct {
	Project string `json:"project"`
	Count   int    `json:"count"`
	Series  []int  `json:"series"`
}

// sparkTotalRecord is the Total row in JSON (no project label).
type sparkTotalRecord struct {
	Count  int   `json:"count"`
	Series []int `json:"series"`
}

// sparkWindowRecord describes the rolling axis in JSON.
type sparkWindowRecord struct {
	Buckets         int    `json:"buckets"`
	BucketWidthDays int    `json:"bucket_width_days"`
	Start           string `json:"start"`
	End             string `json:"end"`
}

// sparkEnvelope is the on-the-wire shape for ToSparkJSON. Struct-tag order is
// the JSON key order DEC-014 locks (flat provenance first).
type sparkEnvelope struct {
	GeneratedAt string              `json:"generated_at"`
	Scope       string              `json:"scope"`
	Filters     map[string]string   `json:"filters"`
	Window      sparkWindowRecord   `json:"window"`
	Total       sparkTotalRecord    `json:"total"`
	ByProject   []sparkSeriesRecord `json:"by_project"`
}

// ToSparkJSON renders the DEC-014 envelope with SPEC-059's payload keys. Every
// key is always emitted; on an empty window total is the full zero-filled
// series, by_project is [] (non-nil), and filters is {}. JSON never contains
// glyphs — each row carries a raw series int array (DEC-031 choice f).
func ToSparkJSON(entries []storage.Entry, opts SparkOptions) ([]byte, error) {
	total, projects := sparkRows(entries, opts)

	byProject := make([]sparkSeriesRecord, 0, len(projects))
	for _, r := range projects {
		byProject = append(byProject, sparkSeriesRecord{
			Project: r.Label,
			Count:   r.Count,
			Series:  r.Series,
		})
	}

	start := opts.Now.Add(-opts.Width * time.Duration(opts.Buckets))
	env := sparkEnvelope{
		GeneratedAt: opts.Now.UTC().Format(time.RFC3339),
		Scope:       opts.Scope,
		Filters:     sparkFiltersJSON(opts),
		Window: sparkWindowRecord{
			Buckets:         opts.Buckets,
			BucketWidthDays: int(opts.Width / (24 * time.Hour)),
			Start:           start.UTC().Format(time.RFC3339),
			End:             opts.Now.UTC().Format(time.RFC3339),
		},
		Total: sparkTotalRecord{
			Count:  total.Count,
			Series: total.Series,
		},
		ByProject: byProject,
	}

	return json.MarshalIndent(env, "", "  ")
}

// sparkFiltersMarkdown echoes the markdown Filters: line. v1 has no tag/type
// filters; --project is the only echoable flag (a row selector, echoed for
// provenance), else "(none)".
func sparkFiltersMarkdown(opts SparkOptions) string {
	if opts.Project != "" {
		return fmt.Sprintf("--project %s", opts.Project)
	}
	return "(none)"
}

// sparkFiltersJSON mirrors sparkFiltersMarkdown as the DEC-014 filters object
// (always non-nil; {} when none).
func sparkFiltersJSON(opts SparkOptions) map[string]string {
	obj := map[string]string{}
	if opts.Project != "" {
		obj["project"] = opts.Project
	}
	return obj
}

// sumInts returns the sum of a slice of ints.
func sumInts(xs []int) int {
	total := 0
	for _, x := range xs {
		total += x
	}
	return total
}

// joinInts renders a series as space-joined decimals (the --no-spark fallback).
func joinInts(xs []int) string {
	parts := make([]string, len(xs))
	for i, x := range xs {
		parts[i] = strconv.Itoa(x)
	}
	return strings.Join(parts, " ")
}
