package cli

import (
	"github.com/spf13/cobra"
)

// NewRootCmd creates the root cobra command for the brag CLI.
func NewRootCmd(version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "brag",
		Short:   "Capture and retrieve career accomplishments",
		Long:    "Bragfile — a local-first CLI for engineers to capture and retrieve career accomplishments for retros, reviews, and resumes.",
		Version: version,
		// Print help when invoked with no args and no subcommands.
		// This also makes the command "runnable" so cobra includes
		// Usage and Flags sections in help output.
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.PersistentFlags().String("db", "", "path to bragfile db (default ~/.bragfile/db.sqlite)")

	return cmd
}
