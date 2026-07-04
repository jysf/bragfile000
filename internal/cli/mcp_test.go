package cli

import (
	"bytes"
	"strings"
	"testing"
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
