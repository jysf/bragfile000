package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootCmd_VersionFlag(t *testing.T) {
	var outBuf, errBuf bytes.Buffer
	cmd := NewRootCmd("test-v0")
	cmd.SetOut(&outBuf)
	cmd.SetErr(&errBuf)
	cmd.SetArgs([]string{"--version"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(outBuf.String(), "test-v0") {
		t.Errorf("expected stdout to contain %q, got %q", "test-v0", outBuf.String())
	}
	if errBuf.Len() != 0 {
		t.Errorf("expected stderr to be empty, got %q", errBuf.String())
	}
}

func TestRootCmd_HelpFlag(t *testing.T) {
	var outBuf, errBuf bytes.Buffer
	cmd := NewRootCmd("test-v0")
	cmd.SetOut(&outBuf)
	cmd.SetErr(&errBuf)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := outBuf.String()
	if !strings.Contains(out, "--db") {
		t.Errorf("expected help to mention --db, got %q", out)
	}
	if !strings.Contains(out, "Usage:") {
		t.Errorf("expected help to contain 'Usage:', got %q", out)
	}
	if errBuf.Len() != 0 {
		t.Errorf("expected stderr to be empty, got %q", errBuf.String())
	}
}

func TestRootCmd_NoArgs(t *testing.T) {
	var outBuf, errBuf bytes.Buffer
	cmd := NewRootCmd("test-v0")
	cmd.SetOut(&outBuf)
	cmd.SetErr(&errBuf)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(outBuf.String(), "Usage:") {
		t.Errorf("expected stdout to contain 'Usage:', got %q", outBuf.String())
	}
	if errBuf.Len() != 0 {
		t.Errorf("expected stderr to be empty, got %q", errBuf.String())
	}
}
