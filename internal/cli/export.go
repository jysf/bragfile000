package cli

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jysf/bragfile000/internal/config"
	"github.com/jysf/bragfile000/internal/export"
	"github.com/jysf/bragfile000/internal/storage"
	"github.com/spf13/cobra"
)

// NewExportCmd returns the `brag export` subcommand. STAGE-003 introduced
// it with --format json (SPEC-014); SPEC-015 added --format markdown and
// the markdown-only --flat modifier. Filter flags mirror `brag list`
// verbatim (SPEC-007's ListFilter).
func NewExportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export brag entries in a machine-readable or review-ready format",
		Long: `Export brag entries. --format is required (one of: json, markdown).

JSON output shape is locked by DEC-011: naked array of entry objects, 9 keys
in SQL-column order, tags as comma-joined string, timestamps as RFC3339.

Markdown output shape is locked by DEC-013: level-1 document heading, a
provenance block (Exported / Entries / Filters), an executive summary,
and entries grouped by project (alphabetical-ASC, (no project) last)
with within-group chronological-ASC ordering. --flat swaps grouping for
a single "## Entries (chronological)" wrapper.

Examples:
  brag export --format json                          # stdout: JSON array
  brag export --format json --out entries.json       # write to file (overwrites)
  brag export --format markdown                      # stdout: grouped markdown
  brag export --format markdown --flat               # stdout: flat chronological
  brag export --format markdown --out report.md      # write to file (overwrites)
  brag export --format markdown --project platform   # filter before exporting
  brag export --format json --tag auth --since 30d`,
		RunE: runExport,
	}
	cmd.Flags().String("format", "", "output format (required; one of: json, markdown)")
	cmd.Flags().String("out", "", "write output to this path instead of stdout (overwrites if present)")
	cmd.Flags().Bool("flat", false, "skip grouping in --format markdown (chrono-ASC single section)")
	cmd.Flags().String("tag", "", "filter to entries whose tags contain this token (comma-separated match)")
	cmd.Flags().String("project", "", "filter to entries with this project (exact match)")
	cmd.Flags().String("type", "", "filter to entries with this type (exact match)")
	cmd.Flags().String("since", "", "filter to entries on or after this point (YYYY-MM-DD or Nd/Nw/Nm)")
	cmd.Flags().Int("limit", 0, "cap the number of rows returned (must be > 0 when set)")
	return cmd
}

func runExport(cmd *cobra.Command, _ []string) error {
	format, _ := cmd.Flags().GetString("format")
	if format == "" {
		return UserErrorf("--format is required (accepted: json, markdown)")
	}
	if format != "json" && format != "markdown" {
		return UserErrorf("unknown --format value %q (accepted: json, markdown)", format)
	}

	flat, _ := cmd.Flags().GetBool("flat")
	if flat && format != "markdown" {
		return UserErrorf("--flat only applies to --format markdown")
	}

	filter := storage.ListFilter{}

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
	if cmd.Flags().Changed("since") {
		raw, _ := cmd.Flags().GetString("since")
		t, err := ParseSince(raw)
		if err != nil {
			return UserErrorf("invalid --since %q: %v", raw, err)
		}
		filter.Since = t
	}
	if cmd.Flags().Changed("limit") {
		n, _ := cmd.Flags().GetInt("limit")
		if n <= 0 {
			return UserErrorf("--limit must be positive, got %d", n)
		}
		filter.Limit = n
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

	var body []byte
	switch format {
	case "json":
		body, err = export.ToJSON(entries)
	case "markdown":
		body, err = export.ToMarkdown(entries, export.MarkdownOptions{
			Flat:    flat,
			Filters: echoFilters(cmd),
			Now:     time.Now().UTC(),
		})
	}
	if err != nil {
		return fmt.Errorf("marshal %s: %w", format, err)
	}

	outPath, _ := cmd.Flags().GetString("out")
	if outPath != "" {
		f, err := os.OpenFile(outPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
		if err != nil {
			return fmt.Errorf("write output: %w", err)
		}
		defer f.Close()
		if _, err := fmt.Fprintln(f, string(body)); err != nil {
			return fmt.Errorf("write output: %w", err)
		}
		return nil
	}

	fmt.Fprintln(cmd.OutOrStdout(), string(body))
	return nil
}

// echoFilters assembles the Filters: provenance line for markdown
// exports. Returns "(none)" when no filter flags were set; otherwise
// echoes each filter flag in flag-declaration order, space-separated,
// matching what a user would retype. Locked by DEC-013 choice 6.
func echoFilters(cmd *cobra.Command) string {
	var parts []string
	order := []string{"tag", "project", "type", "since", "limit"}
	for _, name := range order {
		if !cmd.Flags().Changed(name) {
			continue
		}
		if name == "limit" {
			n, _ := cmd.Flags().GetInt(name)
			parts = append(parts, fmt.Sprintf("--%s %d", name, n))
		} else {
			v, _ := cmd.Flags().GetString(name)
			parts = append(parts, fmt.Sprintf("--%s %s", name, v))
		}
	}
	if len(parts) == 0 {
		return "(none)"
	}
	return strings.Join(parts, " ")
}
