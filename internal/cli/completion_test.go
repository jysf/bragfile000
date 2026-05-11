package cli

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func newCompletionTestRoot(t *testing.T) (*cobra.Command, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	root := NewRootCmd("test")
	root.AddCommand(NewCompletionCmd(root))
	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	root.SetOut(outBuf)
	root.SetErr(errBuf)
	return root, outBuf, errBuf
}

// TestCompletionCmd_Zsh pairs locked decision §2 (root parameter) and
// verifies the §12(b) zsh marker.
func TestCompletionCmd_Zsh(t *testing.T) {
	root, outBuf, errBuf := newCompletionTestRoot(t)
	root.SetArgs([]string{"completion", "zsh"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errBuf.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", errBuf.String())
	}
	if !strings.Contains(outBuf.String(), "#compdef brag") {
		t.Errorf("zsh completion missing '#compdef brag'; got prefix %q", firstChars(outBuf.String(), 80))
	}
}

// TestCompletionCmd_Bash pairs locked decision §2 and verifies the §12(b) bash
// marker (__start_brag, NOT _brag_completion — design-time verified).
func TestCompletionCmd_Bash(t *testing.T) {
	root, outBuf, errBuf := newCompletionTestRoot(t)
	root.SetArgs([]string{"completion", "bash"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errBuf.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", errBuf.String())
	}
	if !strings.Contains(outBuf.String(), "__start_brag") {
		t.Errorf("bash completion missing '__start_brag'; got %d bytes", outBuf.Len())
	}
}

// TestCompletionCmd_Fish pairs locked decision §2 and verifies the §12(b) fish
// marker.
func TestCompletionCmd_Fish(t *testing.T) {
	root, outBuf, errBuf := newCompletionTestRoot(t)
	root.SetArgs([]string{"completion", "fish"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errBuf.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", errBuf.String())
	}
	if !strings.Contains(outBuf.String(), "complete -c brag") {
		t.Errorf("fish completion missing 'complete -c brag'; got %d bytes", outBuf.Len())
	}
}

// TestCompletionCmd_UnsupportedShell pairs locked decision §1 (powershell
// skipped) and §3 (stdout empty on error).
func TestCompletionCmd_UnsupportedShell(t *testing.T) {
	root, outBuf, _ := newCompletionTestRoot(t)
	root.SetArgs([]string{"completion", "powershell"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for unsupported shell, got nil")
	}
	if !errors.Is(err, ErrUser) {
		t.Errorf("expected ErrUser for unsupported shell, got %v", err)
	}
	if outBuf.Len() != 0 {
		t.Errorf("expected empty stdout on error, got %q", outBuf.String())
	}
	if !strings.Contains(err.Error(), "powershell") {
		t.Errorf("error should name the unsupported shell, got %q", err.Error())
	}
}

// TestCompletionCmd_NoArgs pairs locked decision §2 (ExactArgs enforcement).
func TestCompletionCmd_NoArgs(t *testing.T) {
	root, outBuf, _ := newCompletionTestRoot(t)
	root.SetArgs([]string{"completion"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when no shell arg given, got nil")
	}
	if outBuf.Len() != 0 {
		t.Errorf("expected empty stdout on arg error, got %q", outBuf.String())
	}
}

// TestCompletionCmd_HelpShowsSourcingInstructions pairs locked decision §3
// (Long string contains per-shell sourcing pattern).
func TestCompletionCmd_HelpShowsSourcingInstructions(t *testing.T) {
	root, outBuf, errBuf := newCompletionTestRoot(t)
	root.SetArgs([]string{"completion", "--help"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errBuf.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", errBuf.String())
	}
	out := outBuf.String()
	for _, needle := range []string{
		"source <(brag completion zsh)",
		"source <(brag completion bash)",
		"brag completion fish | source",
	} {
		if !strings.Contains(out, needle) {
			t.Errorf("help text missing sourcing instruction %q", needle)
		}
	}
}
