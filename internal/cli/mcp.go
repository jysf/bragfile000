package cli

import (
	"fmt"

	"github.com/jysf/bragfile000/internal/config"
	"github.com/jysf/bragfile000/internal/mcpserver"
	"github.com/jysf/bragfile000/internal/storage"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
)

// NewMCPCmd returns the `brag mcp` parent command, with `serve` as its only
// child.
func NewMCPCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Model Context Protocol server commands",
		Long: `Model Context Protocol (MCP) server commands.

See 'brag mcp serve --help' for the local stdio server that exposes
brag_add / brag_list / brag_search / brag_stats as typed tools.`,
	}
	cmd.AddCommand(newMCPServeCmd())
	return cmd
}

func newMCPServeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Run a local stdio MCP server",
		Long: `Run a local stdio Model Context Protocol (MCP) server exposing brag_add,
brag_list, brag_search, and brag_stats as typed tools over the same
~/.bragfile/db.sqlite your shell commands use.

This is a local, stdio-only server: no network transport, no separate
install. Process stdout carries only MCP protocol frames — nothing
human-facing is ever printed there; the server's own diagnostics (if any)
go to stderr.

brag_add stamps caller-provided agent/model params as reserved-namespace
tags (agent:<name> / model:<id>); it does not emit a milestone line and
does not auto-fill --project from the server's cwd.

Examples:
  brag mcp serve                    # run until the client closes stdin`,
		RunE: runMCPServe,
	}
}

func runMCPServe(cmd *cobra.Command, _ []string) error {
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

	srv := mcpserver.New(s)
	return srv.Run(cmd.Context(), &mcp.StdioTransport{})
}
