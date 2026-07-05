package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/jysf/bragfile000/internal/cli"
	"github.com/jysf/bragfile000/internal/storage"
)

// version is set to "dev" for local builds. goreleaser injects the
// real values via ldflags (-X main.version=... -X main.commit=...
// -X main.date=...) at release-build time. See .goreleaser.yaml.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// Record the build version for the dev/prod-migration guard (DEC-026):
	// an unreleased build refuses to migrate the real ~/.bragfile.
	storage.SetBuildVersion(version)

	root := cli.NewRootCmd(version)
	root.AddCommand(cli.NewAddCmd())
	root.AddCommand(cli.NewListCmd())
	root.AddCommand(cli.NewShowCmd())
	root.AddCommand(cli.NewDeleteCmd())
	root.AddCommand(cli.NewEditCmd())
	root.AddCommand(cli.NewSearchCmd())
	root.AddCommand(cli.NewExportCmd())
	root.AddCommand(cli.NewSummaryCmd())
	root.AddCommand(cli.NewReviewCmd())
	root.AddCommand(cli.NewStatsCmd())
	root.AddCommand(cli.NewTagsCmd())
	root.AddCommand(cli.NewTagCmd())
	root.AddCommand(cli.NewProjectCmd())
	root.AddCommand(cli.NewCompletionCmd(root))
	root.AddCommand(cli.NewMCPCmd())

	if err := root.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "brag: %s\n", err.Error())
		// ErrUser and the dev/prod-migration guard (DEC-026) are user-actionable
		// → exit 1; everything else is internal → exit 2.
		if errors.Is(err, cli.ErrUser) || errors.Is(err, storage.ErrDevProdMigrate) {
			os.Exit(1)
		}
		os.Exit(2)
	}
}
