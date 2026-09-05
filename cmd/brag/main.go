// Package main is the brag CLI entrypoint. It resolves the build version
// (goreleaser's ldflags, or the embedded module version for a
// `go install ./cmd/brag@latest` build — see resolveVersion), wires every
// internal/cli command onto the root cobra command, and maps the returned
// error to an exit code: ErrUser and storage.ErrDevProdMigrate exit 1
// (user-actionable), everything else exits 2 (internal fault). It holds no
// business logic of its own — see internal/cli for the commands and
// internal/storage for persistence.
package main

import (
	"errors"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/jysf/bragfile000/internal/cli"
	"github.com/jysf/bragfile000/internal/storage"
)

// version is set to "dev" for local builds. goreleaser injects the real
// value via an ldflag (-X main.version=...) at release-build time. See
// .goreleaser.yaml. main.commit and main.date were removed at SPEC-082:
// goreleaser set them and nothing ever read them.
var version = "dev"

func main() {
	// Resolve before anything reads it: goreleaser's ldflags win, and a
	// `go install …@latest` build recovers its module version from the
	// embedded build info. See resolveVersion — a pseudo-version deliberately
	// does NOT count, because that would switch off the guard below.
	version = resolveVersion(version, debug.ReadBuildInfo)

	// Record the build version for the dev/prod-migration guard (DEC-026):
	// an unreleased build refuses to migrate the real ~/.bragfile.
	storage.SetBuildVersion(version)

	root := cli.NewRootCmd(version)
	root.AddCommand(cli.NewAddCmd())
	root.AddCommand(cli.NewLearnCmd())
	root.AddCommand(cli.NewListCmd())
	root.AddCommand(cli.NewShowCmd())
	root.AddCommand(cli.NewDeleteCmd())
	root.AddCommand(cli.NewEditCmd())
	root.AddCommand(cli.NewSearchCmd())
	root.AddCommand(cli.NewExportCmd())
	root.AddCommand(cli.NewSummaryCmd())
	root.AddCommand(cli.NewReviewCmd())
	root.AddCommand(cli.NewStatsCmd())
	root.AddCommand(cli.NewImpactCmd())
	root.AddCommand(cli.NewWrappedCmd())
	root.AddCommand(cli.NewCoverageCmd())
	root.AddCommand(cli.NewSparkCmd())
	root.AddCommand(cli.NewStoryCmd())
	root.AddCommand(cli.NewMemoryCmd())
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
