package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func newTagsTestRoot(t *testing.T) (*cobra.Command, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	t.Setenv("BRAGFILE_DB", "")
	root := NewRootCmd("test")
	root.AddCommand(NewTagsCmd())
	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	root.SetOut(outBuf)
	root.SetErr(errBuf)
	return root, outBuf, errBuf
}

func runTagsCmd(t *testing.T, dbPath string, args ...string) (stdout, stderr string, runErr error) {
	t.Helper()
	root, outBuf, errBuf := newTagsTestRoot(t)
	full := append([]string{"--db", dbPath, "tags"}, args...)
	root.SetArgs(full)
	runErr = root.Execute()
	return outBuf.String(), errBuf.String(), runErr
}

func TestTagsCmd_PlainSortedOutput(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	seedListEntry(t, dbPath, "e1", "auth,perf", "", "")
	seedListEntry(t, dbPath, "e2", "perf", "", "")
	seedListEntry(t, dbPath, "e3", "auth,backend", "", "")

	out, errOut, err := runTagsCmd(t, dbPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errOut != "" {
		t.Errorf("expected empty stderr, got %q", errOut)
	}

	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	want := []string{"auth\t2", "perf\t2", "backend\t1"}
	if len(lines) != len(want) {
		t.Fatalf("expected %d lines, got %d: %q", len(want), len(lines), out)
	}
	for i, w := range want {
		if lines[i] != w {
			t.Errorf("line[%d]: got %q, want %q", i, lines[i], w)
		}
	}
}

func TestTagsCmd_JSON(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	seedListEntry(t, dbPath, "e1", "auth,perf", "", "")
	seedListEntry(t, dbPath, "e2", "perf", "", "")

	out, errOut, err := runTagsCmd(t, dbPath, "--format", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errOut != "" {
		t.Errorf("expected empty stderr, got %q", errOut)
	}

	trimmed := strings.TrimSpace(out)
	if !strings.HasPrefix(trimmed, "[") {
		t.Errorf("expected naked array, got %q", trimmed)
	}

	var got []struct {
		Tag   string `json:"tag"`
		Count int    `json:"count"`
	}
	if err := json.Unmarshal([]byte(trimmed), &got); err != nil {
		t.Fatalf("JSON unmarshal: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(got))
	}
	if got[0].Tag != "perf" || got[0].Count != 2 {
		t.Errorf("got[0]: {%q,%d}, want {perf,2}", got[0].Tag, got[0].Count)
	}
	if got[1].Tag != "auth" || got[1].Count != 1 {
		t.Errorf("got[1]: {%q,%d}, want {auth,1}", got[1].Tag, got[1].Count)
	}

	// 2-space indent check.
	if !strings.Contains(out, "  \"tag\"") {
		t.Errorf("expected 2-space indent in JSON output, got %q", out)
	}
}

func TestTagsCmd_EmptyCorpus(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")

	// Plain mode: stdout empty, exit 0.
	out, errOut, err := runTagsCmd(t, dbPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "" {
		t.Errorf("plain empty corpus: expected empty stdout, got %q", out)
	}
	if errOut != "" {
		t.Errorf("expected empty stderr, got %q", errOut)
	}

	// JSON mode: emits [].
	out2, errOut2, err2 := runTagsCmd(t, dbPath, "--format", "json")
	if err2 != nil {
		t.Fatalf("unexpected error: %v", err2)
	}
	if strings.TrimSpace(out2) != "[]" {
		t.Errorf("json empty corpus: expected [], got %q", out2)
	}
	if errOut2 != "" {
		t.Errorf("expected empty stderr, got %q", errOut2)
	}
}

func TestTagsCmd_UnknownFormat(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	_, _, err := runTagsCmd(t, dbPath, "--format", "xml")
	if !errors.Is(err, ErrUser) {
		t.Fatalf("expected ErrUser for unknown format, got %v", err)
	}
}

func TestTagsCmd_StdoutStderrSeparation(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	seedListEntry(t, dbPath, "e1", "auth", "", "")

	_, errOut, err := runTagsCmd(t, dbPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errOut != "" {
		t.Errorf("stderr must be empty on success, got %q", errOut)
	}
}

func TestTagsCmd_HelpShowsExamples(t *testing.T) {
	root, outBuf, _ := newTagsTestRoot(t)
	root.SetArgs([]string{"tags", "--help"})
	_ = root.Execute()
	out := outBuf.String()
	if !strings.Contains(out, "Examples:") {
		t.Errorf("help output missing 'Examples:': %q", out)
	}
	if !strings.Contains(out, "brag tags --format json") {
		t.Errorf("help output missing distinctive 'brag tags --format json' example: %q", out)
	}
}
