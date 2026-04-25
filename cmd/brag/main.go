package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/jysf/bragfile000/internal/cli"
)

// Version is set to "dev" for local builds. goreleaser injects the real
// version via ldflags in STAGE-004.
const Version = "dev"

func main() {
	root := cli.NewRootCmd(Version)
	root.AddCommand(cli.NewAddCmd())
	root.AddCommand(cli.NewListCmd())
	root.AddCommand(cli.NewShowCmd())
	root.AddCommand(cli.NewDeleteCmd())
	root.AddCommand(cli.NewEditCmd())
	root.AddCommand(cli.NewSearchCmd())
	root.AddCommand(cli.NewExportCmd())
	root.AddCommand(cli.NewSummaryCmd())
	root.AddCommand(cli.NewReviewCmd())

	if err := root.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "brag: %s\n", err.Error())
		if errors.Is(err, cli.ErrUser) {
			os.Exit(1)
		}
		os.Exit(2)
	}
}
