package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
)

// mcpServerBlock is the fixed shape of an MCP server entry. Key order is
// deterministic (command then args) so the merged output is stable across
// runs, matching the block the shipped plugin uses (plugin/.mcp.json).
type mcpServerBlock struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

// canonicalBragBlock is the exact server block registered for `brag mcp serve`
// (DEC-024), reused verbatim from plugin/.mcp.json.
var canonicalBragBlock = mcpServerBlock{Command: "brag", Args: []string{"mcp", "serve"}}

// userHomeDir is an injectable seam for tests (AGENTS.md §9).
var userHomeDir = os.UserHomeDir

// mergeMCPConfig returns the full bytes to write for the target config file
// after ensuring mcpServers.<serverName> == block, preserving every other
// top-level key and every other server. Absent/empty existing input yields a
// file containing just the mcpServers.<serverName> block. Output is 2-space
// indented with a trailing newline. Preservation is SEMANTIC (values survive);
// encoding/json canonicalizes key order + whitespace on rewrite (DEC-034 #3).
func mergeMCPConfig(existing []byte, serverName string, block mcpServerBlock) ([]byte, error) {
	top := map[string]json.RawMessage{}
	if len(bytes.TrimSpace(existing)) > 0 {
		if err := json.Unmarshal(existing, &top); err != nil {
			return nil, fmt.Errorf("parse existing config: %w", err)
		}
	}
	servers := map[string]json.RawMessage{}
	if raw, ok := top["mcpServers"]; ok {
		if err := json.Unmarshal(raw, &servers); err != nil {
			return nil, fmt.Errorf("parse mcpServers: %w", err)
		}
	}
	blockBytes, err := json.Marshal(block)
	if err != nil {
		return nil, fmt.Errorf("marshal server block: %w", err)
	}
	servers[serverName] = blockBytes
	serversBytes, err := json.Marshal(servers)
	if err != nil {
		return nil, fmt.Errorf("marshal mcpServers: %w", err)
	}
	top["mcpServers"] = serversBytes
	out, err := json.MarshalIndent(top, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal config: %w", err)
	}
	return append(out, '\n'), nil
}

// resolveInstallTarget maps (client, scope, dir) → the config file path.
// dir defaults (in the caller) to getCwd() when empty; it is only consulted
// for project scope. Unsupported combos return UserErrorf(...).
func resolveInstallTarget(client, scope, dir string) (string, error) {
	switch client {
	case "claude-code":
		switch scope {
		case "project":
			return filepath.Abs(filepath.Join(dir, ".mcp.json"))
		case "user":
			home, err := userHomeDir()
			if err != nil {
				return "", fmt.Errorf("resolve home dir: %w", err)
			}
			return filepath.Join(home, ".claude.json"), nil
		default:
			return "", UserErrorf("unknown --scope %q (accepted: user, project)", scope)
		}
	case "cursor":
		switch scope {
		case "project":
			return filepath.Abs(filepath.Join(dir, ".cursor", "mcp.json"))
		case "user":
			home, err := userHomeDir()
			if err != nil {
				return "", fmt.Errorf("resolve home dir: %w", err)
			}
			return filepath.Join(home, ".cursor", "mcp.json"), nil
		default:
			return "", UserErrorf("unknown --scope %q (accepted: user, project)", scope)
		}
	case "claude-desktop":
		switch scope {
		case "project":
			return "", UserErrorf("claude-desktop has no project scope; use --scope user")
		case "user":
			home, err := userHomeDir()
			if err != nil {
				return "", fmt.Errorf("resolve home dir: %w", err)
			}
			switch runtime.GOOS {
			case "darwin":
				return filepath.Join(home, "Library", "Application Support", "Claude", "claude_desktop_config.json"), nil
			case "windows":
				appData := os.Getenv("APPDATA")
				if appData == "" {
					appData = filepath.Join(home, "AppData", "Roaming")
				}
				return filepath.Join(appData, "Claude", "claude_desktop_config.json"), nil
			default:
				return "", UserErrorf("claude-desktop config path is unknown on %s (supported: macOS, Windows)", runtime.GOOS)
			}
		default:
			return "", UserErrorf("unknown --scope %q (accepted: user, project)", scope)
		}
	default:
		return "", UserErrorf("unknown --client %q (accepted: claude-code, claude-desktop, cursor)", client)
	}
}

func newMCPInstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Register the brag MCP server in a client's config",
		Long: `Write or merge the brag MCP server block into a client's config file,
idempotently and without clobbering any other MCP server already present.

The registered block is exactly what 'brag mcp serve' expects:
{"command":"brag","args":["mcp","serve"]}. Running install repeatedly is
safe — a byte-identical result is detected and reported as a no-op. Any other
MCP server and any unrelated top-level key in the target file is preserved.

--dry-run prints the exact JSON that would be written to stdout and the
resolved target path to stderr, writing nothing to disk.

Note: MCP servers connect at client startup; restart or reconnect the client
session after installing.

Examples:
  brag mcp install                                  # claude-code, project scope (<cwd>/.mcp.json)
  brag mcp install --client cursor                  # cursor, project scope (<cwd>/.cursor/mcp.json)
  brag mcp install --client claude-desktop --scope user
  brag mcp install --dry-run                         # print the JSON + path, write nothing`,
		RunE: runMCPInstall,
	}
	cmd.Flags().String("client", "claude-code", "MCP client to configure (one of: claude-code, claude-desktop, cursor)")
	cmd.Flags().String("scope", "project", "config scope (one of: user, project)")
	cmd.Flags().String("dir", "", "project directory for project scope (default: current directory)")
	cmd.Flags().Bool("dry-run", false, "print the exact JSON + target path without writing")
	return cmd
}

func runMCPInstall(cmd *cobra.Command, _ []string) error {
	client := getFlagString(cmd, "client")
	scope := getFlagString(cmd, "scope")
	dir := getFlagString(cmd, "dir")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	if scope == "user" && cmd.Flags().Changed("dir") {
		return UserErrorf("--dir applies only to project scope")
	}

	if dir == "" {
		var err error
		dir, err = getCwd()
		if err != nil {
			return fmt.Errorf("resolve working directory: %w", err)
		}
	}

	target, err := resolveInstallTarget(client, scope, dir)
	if err != nil {
		return err
	}

	existing, err := os.ReadFile(target)
	if err != nil {
		if os.IsNotExist(err) {
			existing = nil
		} else {
			return fmt.Errorf("read existing config: %w", err)
		}
	}

	merged, err := mergeMCPConfig(existing, "brag", canonicalBragBlock)
	if err != nil {
		return err
	}

	if dryRun {
		fmt.Fprintf(cmd.ErrOrStderr(), "Would write to %s:\n", target)
		if _, err := cmd.OutOrStdout().Write(merged); err != nil {
			return fmt.Errorf("write dry-run output: %w", err)
		}
		return nil
	}

	if existing != nil && bytes.Equal(existing, merged) {
		fmt.Fprintf(cmd.ErrOrStderr(), "brag MCP server already registered in %s (no changes)\n", target)
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	if err := os.WriteFile(target, merged, 0o644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	fmt.Fprintf(cmd.ErrOrStderr(), "Registered brag MCP server in %s\n", target)
	return nil
}
