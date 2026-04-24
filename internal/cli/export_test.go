package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jysf/bragfile000/internal/export"
	"github.com/jysf/bragfile000/internal/storage"
	"github.com/spf13/cobra"
)

// newExportTestRoot mirrors newListTestRoot: fresh root with the export
// subcommand attached and separate stdout/stderr buffers.
func newExportTestRoot(t *testing.T) (*cobra.Command, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	t.Setenv("BRAGFILE_DB", "")
	root := NewRootCmd("test")
	root.AddCommand(NewExportCmd())
	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	root.SetOut(outBuf)
	root.SetErr(errBuf)
	return root, outBuf, errBuf
}

func runExportCmd(t *testing.T, dbPath string, args ...string) (stdout, stderr string, runErr error) {
	t.Helper()
	root, outBuf, errBuf := newExportTestRoot(t)
	full := append([]string{"--db", dbPath, "export"}, args...)
	root.SetArgs(full)
	runErr = root.Execute()
	return outBuf.String(), errBuf.String(), runErr
}

// runListForCompare runs `brag list` via the list subcommand harness.
// Used by the byte-identical test.
func runListForCompare(t *testing.T, dbPath string, args ...string) string {
	t.Helper()
	t.Setenv("BRAGFILE_DB", "")
	root := NewRootCmd("test")
	root.AddCommand(NewListCmd())
	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	root.SetOut(outBuf)
	root.SetErr(errBuf)
	full := append([]string{"--db", dbPath, "list"}, args...)
	root.SetArgs(full)
	if err := root.Execute(); err != nil {
		t.Fatalf("brag list: %v", err)
	}
	if errBuf.Len() != 0 {
		t.Fatalf("brag list stderr: %q", errBuf.String())
	}
	return outBuf.String()
}

func TestExportCmd_FormatRequiredIsUserError(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	seedListEntry(t, dbPath, "solo", "", "", "")

	out, _, err := runExportCmd(t, dbPath)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrUser) {
		t.Errorf("expected errors.Is(err, ErrUser); got %v", err)
	}
	if out != "" {
		t.Errorf("expected empty stdout, got %q", out)
	}
	msg := err.Error()
	if !strings.Contains(msg, "--format") {
		t.Errorf("expected error to mention --format, got %q", msg)
	}
	if !strings.Contains(msg, "json") {
		t.Errorf("expected error to mention accepted value json, got %q", msg)
	}
}

func TestExportCmd_FormatUnknownValueIsUserError(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	seedListEntry(t, dbPath, "solo", "", "", "")

	out, _, err := runExportCmd(t, dbPath, "--format", "yaml")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrUser) {
		t.Errorf("expected errors.Is(err, ErrUser); got %v", err)
	}
	if out != "" {
		t.Errorf("expected empty stdout, got %q", out)
	}
	msg := err.Error()
	if !strings.Contains(msg, "yaml") {
		t.Errorf("expected error to mention offending value %q, got %q", "yaml", msg)
	}
	if !strings.Contains(msg, "json") {
		t.Errorf("expected error to mention accepted value json, got %q", msg)
	}
}

func TestExportCmd_FormatJSON_StdoutEmitsDEC011Shape(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	a := seedListEntry(t, dbPath, "first", "auth", "platform", "shipped")
	b := seedListEntry(t, dbPath, "second", "", "growth", "")

	out, errOut, err := runExportCmd(t, dbPath, "--format", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errOut != "" {
		t.Fatalf("expected empty stderr, got %q", errOut)
	}

	expected, err := export.ToJSON([]storage.Entry{b, a})
	if err != nil {
		t.Fatalf("export.ToJSON: %v", err)
	}
	want := string(expected) + "\n"
	if out != want {
		t.Fatalf("stdout mismatch\nwant:\n%s\ngot:\n%s", want, out)
	}

	var arr []map[string]any
	if err := json.Unmarshal([]byte(strings.TrimRight(out, "\n")), &arr); err != nil {
		t.Fatalf("parse json: %v", err)
	}
	if len(arr) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(arr))
	}
	wantKeys := []string{"id", "title", "description", "tags", "project", "type", "impact", "created_at", "updated_at"}
	if len(arr[0]) != len(wantKeys) {
		t.Errorf("entry 0 keys: got %d, want %d", len(arr[0]), len(wantKeys))
	}
	for _, k := range wantKeys {
		if _, ok := arr[0][k]; !ok {
			t.Errorf("entry 0: missing key %q", k)
		}
	}
}

func TestExportCmd_FormatJSON_OutPathWritesFile(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	a := seedListEntry(t, dbPath, "first", "", "", "")
	b := seedListEntry(t, dbPath, "second", "", "", "")

	outPath := filepath.Join(t.TempDir(), "export.json")
	if err := os.WriteFile(outPath, []byte("PRE-EXISTING CONTENT\n"), 0o644); err != nil {
		t.Fatalf("pre-seed file: %v", err)
	}

	stdout, stderr, err := runExportCmd(t, dbPath, "--format", "json", "--out", outPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stdout != "" {
		t.Errorf("expected empty stdout under --out, got %q", stdout)
	}
	if stderr != "" {
		t.Errorf("expected empty stderr, got %q", stderr)
	}

	got, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	expected, err := export.ToJSON([]storage.Entry{b, a})
	if err != nil {
		t.Fatalf("export.ToJSON: %v", err)
	}
	want := string(expected) + "\n"
	if string(got) != want {
		t.Fatalf("file content mismatch\nwant:\n%s\ngot:\n%s", want, string(got))
	}
	if strings.Contains(string(got), "PRE-EXISTING CONTENT") {
		t.Errorf("expected overwrite; sentinel still present in file")
	}
}

func TestExportCmd_FormatJSON_FiltersApply(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	seedListEntry(t, dbPath, "platform-hit", "", "platform", "")
	seedListEntry(t, dbPath, "growth-miss", "", "growth", "")
	seedListEntry(t, dbPath, "no-project", "", "", "")

	out, errOut, err := runExportCmd(t, dbPath, "--format", "json", "--project", "platform")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errOut != "" {
		t.Fatalf("expected empty stderr, got %q", errOut)
	}

	var arr []map[string]any
	if err := json.Unmarshal([]byte(strings.TrimRight(out, "\n")), &arr); err != nil {
		t.Fatalf("parse json: %v\n%s", err, out)
	}
	if len(arr) != 1 {
		t.Fatalf("expected 1 entry after --project platform filter, got %d", len(arr))
	}
	if arr[0]["project"] != "platform" {
		t.Errorf("want project=platform, got %v", arr[0]["project"])
	}
}

// TestExportCmd_FormatJSON_ByteIdenticalToListJSON is the load-bearing
// cross-path assertion for DEC-011: both commands must route through
// internal/export.ToJSON and produce byte-identical stdout.
func TestExportCmd_FormatJSON_ByteIdenticalToListJSON(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	seedListEntryFull(t, dbPath, storage.Entry{
		Title:       "full",
		Description: "desc-full",
		Tags:        "t1,t2",
		Project:     "platform",
		Type:        "shipped",
		Impact:      "imp-full",
	})
	seedListEntryFull(t, dbPath, storage.Entry{Title: "bare"})
	seedListEntryFull(t, dbPath, storage.Entry{
		Title:   "middle",
		Tags:    "t1",
		Project: "growth",
	})

	listOut := runListForCompare(t, dbPath, "--format", "json")
	exportOut, errOut, err := runExportCmd(t, dbPath, "--format", "json")
	if err != nil {
		t.Fatalf("export: unexpected error: %v", err)
	}
	if errOut != "" {
		t.Fatalf("export stderr: %q", errOut)
	}
	if listOut != exportOut {
		t.Fatalf("DEC-011 drift: list and export JSON differ\nlist:\n%s\n\nexport:\n%s", listOut, exportOut)
	}
}

func TestExportCmd_HelpShowsFormat(t *testing.T) {
	root, outBuf, errBuf := newExportTestRoot(t)
	root.SetArgs([]string{"export", "--help"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errBuf.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", errBuf.String())
	}
	out := outBuf.String()
	for _, needle := range []string{"--format", "json"} {
		if !strings.Contains(out, needle) {
			t.Errorf("expected help to contain %q, got:\n%s", needle, out)
		}
	}
}
