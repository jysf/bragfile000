package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootCmd_VersionFlag(t *testing.T) {
	var buf bytes.Buffer
	cmd := NewRootCmd("test-v0")
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--version"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "test-v0") {
		t.Errorf("expected output to contain %q, got %q", "test-v0", buf.String())
	}
}

func TestRootCmd_HelpFlag(t *testing.T) {
	var buf bytes.Buffer
	cmd := NewRootCmd("test-v0")
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "--db") {
		t.Errorf("expected help to mention --db, got %q", out)
	}
	if !strings.Contains(out, "Usage:") {
		t.Errorf("expected help to contain 'Usage:', got %q", out)
	}
}

func TestRootCmd_NoArgs(t *testing.T) {
	var buf bytes.Buffer
	cmd := NewRootCmd("test-v0")
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "Usage:") {
		t.Errorf("expected output to contain 'Usage:', got %q", buf.String())
	}
}
