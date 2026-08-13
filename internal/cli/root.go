package cli

import (
	"github.com/spf13/cobra"
)

// NewRootCmd creates the root cobra command for the brag CLI.
func NewRootCmd(version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "brag",
		Short: "Capture and retrieve career accomplishments",
		Long: `Bragfile — a local-first CLI for engineers to capture and retrieve career accomplishments for retros, reviews, and resumes.

Run 'brag <command> --help' for command-specific flags and usage.`,
		Version: version,
		// Print help when invoked with no args and no subcommands.
		// This also makes the command "runnable" so cobra includes
		// Usage and Flags sections in help output.
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
		// main.go owns error formatting and exit-code mapping; silence
		// cobra's default "Error: ..." line and usage-on-failure dump.
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.PersistentFlags().String("db", "", "path to bragfile db (default ~/.bragfile/db.sqlite)")

	// A malformed flag is a USER error, not an internal fault. Without this,
	// cobra's flag-parse errors reach main.go unwrapped, fail the
	// errors.Is(err, ErrUser) check, and exit 2 — the code reserved for
	// internal faults — so `brag search -foo` reported a bad flag with the
	// same exit status as a corrupt database. Wrapping here rather than in
	// each subcommand fixes every command at once, since flag parsing is
	// cobra's job and the classification is identical everywhere.
	// (STAGE-018 audit nit.)
	cmd.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		return UserErrorf("%s", err)
	})

	return cmd
}
