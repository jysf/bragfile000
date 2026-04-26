package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/jysf/bragfile000/internal/cli"
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

	if err := root.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "brag: %s\n", err.Error())
		if errors.Is(err, cli.ErrUser) {
			os.Exit(1)
		}
		os.Exit(2)
	}
}
