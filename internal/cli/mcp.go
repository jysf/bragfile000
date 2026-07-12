package cli

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/jysf/bragfile000/internal/config"
	"github.com/jysf/bragfile000/internal/mcpserver"
	"github.com/jysf/bragfile000/internal/storage"
	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
)

// NewMCPCmd returns the `brag mcp` parent command, with `serve` and `install`
// as its children.
func NewMCPCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Model Context Protocol server commands",
		Long: `Model Context Protocol (MCP) server commands.

See 'brag mcp serve --help' for the local stdio server that exposes
brag_add / brag_list / brag_search / brag_stats as typed tools.

See 'brag mcp install --help' to register that server in a client's config
(claude-code, claude-desktop, or cursor) idempotently.`,
	}
	cmd.AddCommand(newMCPServeCmd())
	cmd.AddCommand(newMCPInstallCmd())
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
	return serve(cmd.Context(), srv, &mcp.StdioTransport{})
}

// codeServerClosing is the JSON-RPC error code carried by the go-sdk's internal
// jsonrpc2.ErrServerClosing sentinel ("server is closing"). That sentinel is
// unexported by the SDK, but it is a *jsonrpc.Error (WireError) whose Is method
// compares by Code, so a value carrying the same code matches it via errors.Is.
const codeServerClosing = -32004

// errServerClosing mirrors the SDK's jsonrpc2.ErrServerClosing so a normal
// shutdown can be recognized through errors.Is even when the SDK wraps it (it
// arrives as `server is closing: EOF`, i.e. wrapped with the underlying EOF).
var errServerClosing = &jsonrpc.Error{Code: codeServerClosing}

// isCleanShutdown reports whether err represents a normal MCP transport
// shutdown rather than a genuine serve failure. When a client closes stdin —
// even with a request still in flight — the go-sdk's Server.Run returns a
// context.Canceled, a bare io.EOF, or the "server is closing" sentinel. None of
// these are crashes, so `brag mcp serve` must exit 0 rather than have a
// supervising client log a nonzero exit as a failure.
func isCleanShutdown(err error) bool {
	if err == nil {
		return true
	}
	return errors.Is(err, io.EOF) ||
		errors.Is(err, context.Canceled) ||
		errors.Is(err, errServerClosing)
}

// serve runs srv over t, folding a normal transport shutdown into a clean exit
// (nil) while propagating every genuine serve failure with context so it still
// maps to a nonzero exit code.
func serve(ctx context.Context, srv *mcp.Server, t mcp.Transport) error {
	if err := srv.Run(ctx, t); err != nil && !isCleanShutdown(err) {
		return fmt.Errorf("run mcp server: %w", err)
	}
	return nil
}
