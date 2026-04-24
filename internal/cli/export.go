package cli

import (
	"fmt"
	"os"

	"github.com/jysf/bragfile000/internal/config"
	"github.com/jysf/bragfile000/internal/export"
	"github.com/jysf/bragfile000/internal/storage"
	"github.com/spf13/cobra"
)

// NewExportCmd returns the `brag export` subcommand. STAGE-003 introduces
// it with --format json (required); SPEC-015 will add --format markdown.
// Filter flags mirror `brag list` verbatim (SPEC-007's ListFilter).
func NewExportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export brag entries in a machine-readable format",
		Long: `Export brag entries in a machine-readable format. --format is required.

JSON output shape is locked by DEC-011: naked array of entry objects, 9 keys
in SQL-column order, tags as comma-joined string, timestamps as RFC3339.

Examples:
  brag export --format json                          # stdout: JSON array
  brag export --format json --out entries.json       # write to file (overwrites)
  brag export --format json --project platform       # filter before exporting
  brag export --format json --tag auth --since 30d`,
		RunE: runExport,
	}
	cmd.Flags().String("format", "", "output format (required; one of: json)")
	cmd.Flags().String("out", "", "write output to this path instead of stdout (overwrites if present)")
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
		return UserErrorf("--format is required (accepted: json)")
	}
	if format != "json" {
		return UserErrorf("unknown --format value %q (accepted: json)", format)
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

	body, err := export.ToJSON(entries)
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}

	outPath, _ := cmd.Flags().GetString("out")
	if outPath != "" {
		f, err := os.OpenFile(outPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
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
