package cli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestMCP_ServeRegistered ▲ `brag mcp serve` is wired; help lists it.
func TestMCP_ServeRegistered(t *testing.T) {
	root := NewRootCmd("test")
	root.AddCommand(NewMCPCmd())
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetArgs([]string{"mcp", "--help"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "serve") {
		t.Errorf("`brag mcp --help` should list serve, got %q", out.String())
	}
}

// TestMCP_ServeHelpSaysLocalStdio ▲ the serve help states local/stdio (no network).
func TestMCP_ServeHelpSaysLocalStdio(t *testing.T) {
	root := NewRootCmd("test")
	root.AddCommand(NewMCPCmd())
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetArgs([]string{"mcp", "serve", "--help"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"stdio", "local"} {
		if !strings.Contains(strings.ToLower(out.String()), want) {
			t.Errorf("serve help missing %q: %q", want, out.String())
		}
	}
}

// TestIsCleanShutdown ▲ a normal MCP transport shutdown (bare/ wrapped io.EOF,
// a canceled context, or the SDK's "server is closing" sentinel wrapped with
// EOF) must be classified as clean; any other error must not. The
// server-is-closing case reproduces the exact shape go-sdk's Server.Run
// returns: fmt.Errorf("%w: %v", jsonrpc2.ErrServerClosing, io.EOF) — the
// sentinel wrapped with %w, EOF appended with %v (so it is NOT in the errors.Is
// chain), which is why it must be matched via the jsonrpc.Error code sentinel.
func TestIsCleanShutdown(t *testing.T) {
	serverClosing := fmt.Errorf("%w: %v", &jsonrpc.Error{Code: -32004, Message: "server is closing"}, io.EOF)
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, true},
		{"bare eof", io.EOF, true},
		{"wrapped eof", fmt.Errorf("read frame: %w", io.EOF), true},
		{"context canceled", context.Canceled, true},
		{"wrapped canceled", fmt.Errorf("session: %w", context.Canceled), true},
		{"sdk server closing", serverClosing, true},
		{"generic failure", errors.New("boom"), false},
		{"open store failure", fmt.Errorf("open store: %w", errors.New("disk full")), false},
	}
	for _, tc := range cases {
		if got := isCleanShutdown(tc.err); got != tc.want {
			t.Errorf("isCleanShutdown(%s) = %v, want %v", tc.name, got, tc.want)
		}
	}
}

// TestServe_InFlightClientCloseIsCleanExit ▲ end-to-end: a client that closes
// the transport with a tool request still in flight makes go-sdk's Server.Run
// return the wrapped "server is closing" sentinel. serve must fold that into a
// nil (exit 0) — a supervising client must not see `brag mcp serve` report a
// crash on ordinary shutdown. Deterministic: the test blocks the handler and
// waits for it to be entered before closing the read side.
func TestServe_InFlightClientCloseIsCleanExit(t *testing.T) {
	srv := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "1.0"}, nil)
	entered := make(chan struct{})
	release := make(chan struct{})
	var once sync.Once
	mcp.AddTool(srv, &mcp.Tool{Name: "block", Description: "blocks until released"},
		func(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, struct{}, error) {
			once.Do(func() { close(entered) })
			<-release
			return &mcp.CallToolResult{}, struct{}{}, nil
		})

	cr, cw := io.Pipe() // client -> server (server reads)
	sr, sw := io.Pipe() // server -> client (server writes)
	errCh := make(chan error, 1)
	go func() { errCh <- serve(context.Background(), srv, &mcp.IOTransport{Reader: cr, Writer: sw}) }()
	go func() { _, _ = io.Copy(io.Discard, sr) }() // drain responses so writes never block

	write := func(frame string) {
		t.Helper()
		if _, err := io.WriteString(cw, frame+"\n"); err != nil {
			t.Errorf("write frame: %v", err)
		}
	}
	write(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"c","version":"1"}}}`)
	write(`{"jsonrpc":"2.0","method":"notifications/initialized"}`)
	write(`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"block","arguments":{}}}`)

	select {
	case <-entered:
	case <-time.After(3 * time.Second):
		close(release)
		t.Fatal("tool handler never started; cannot exercise in-flight close")
	}
	cw.Close()     // client closes stdin while the request is in flight
	close(release) // handler completes and tries to respond over the closing conn

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("serve should treat an in-flight client close as a clean exit, got: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("serve did not return after the client closed the transport")
	}
}
