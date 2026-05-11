package cli

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

// NewCompletionCmd returns a subcommand that generates shell completion scripts
// using cobra's built-in generators. root must be the root brag command so the
// generated scripts include all registered subcommands, not just this one.
func NewCompletionCmd(root *cobra.Command) *cobra.Command {
	return &cobra.Command{
		Use:   "completion <shell>",
		Short: "Generate shell completion script",
		Long: `Generate a shell completion script for brag and print it to stdout.

Supported shells: zsh, bash, fish.

To load completions in your current shell session:

  zsh:
    source <(brag completion zsh)

  bash:
    source <(brag completion bash)

  fish:
    brag completion fish | source

To load completions permanently, add the sourcing line above to your
shell's startup file (~/.zshrc, ~/.bashrc, or ~/.config/fish/config.fish).`,
		ValidArgs: []string{"zsh", "bash", "fish"},
		Args:      cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return completionRun(root, cmd.OutOrStdout(), args[0])
		},
	}
}

func completionRun(root *cobra.Command, w io.Writer, shell string) error {
	switch shell {
	case "zsh":
		return root.GenZshCompletion(w)
	case "bash":
		return root.GenBashCompletion(w)
	case "fish":
		return root.GenFishCompletion(w, true)
	default:
		return fmt.Errorf("completion: unsupported shell %q (supported: zsh, bash, fish): %w",
			shell, ErrUser)
	}
}
