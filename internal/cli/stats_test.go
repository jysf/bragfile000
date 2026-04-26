package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// newStatsTestRoot builds a fresh root command with the stats subcommand
// attached and isolates BRAGFILE_DB so the host env can't leak in.
// Mirrors newSummaryTestRoot / newReviewTestRoot.
func newStatsTestRoot(t *testing.T) (*cobra.Command, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	t.Setenv("BRAGFILE_DB", "")
	root := NewRootCmd("test")
	root.AddCommand(NewStatsCmd())
	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	root.SetOut(outBuf)
	root.SetErr(errBuf)
	return root, outBuf, errBuf
}

func runStatsCmd(t *testing.T, dbPath string, args ...string) (stdout, stderr string, runErr error) {
	t.Helper()
	root, outBuf, errBuf := newStatsTestRoot(t)
	full := append([]string{"--db", dbPath, "stats"}, args...)
	root.SetArgs(full)
	runErr = root.Execute()
	return outBuf.String(), errBuf.String(), runErr
}

// TestStatsCmd_BareDefaultsToMarkdown pairs locked decisions §1 (default
// markdown + scope hard-code) and §8 (--format default).
func TestStatsCmd_BareDefaultsToMarkdown(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	seedListEntry(t, dbPath, "alpha-fresh", "auth", "alpha", "shipped")
	seedListEntry(t, dbPath, "beta-fresh", "perf", "beta", "learned")

	out, errOut, err := runStatsCmd(t, dbPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errOut != "" {
		t.Fatalf("expected empty stderr, got %q", errOut)
	}
	if !strings.HasPrefix(out, "# Bragfile Stats\n\n") {
		t.Errorf("expected markdown default header, got prefix %q", firstChars(out, 60))
	}

	scopeFound := false
	filtersFound := false
	statsFound := false
	for _, ln := range strings.Split(out, "\n") {
		if ln == "Scope: lifetime" {
			scopeFound = true
		}
		if ln == "Filters: (none)" {
			filtersFound = true
		}
		if ln == "## Stats" {
			statsFound = true
		}
	}
	if !scopeFound {
		t.Errorf("expected line %q in output:\n%s", "Scope: lifetime", out)
	}
	if !filtersFound {
		t.Errorf("expected line %q in output:\n%s", "Filters: (none)", out)
	}
	if !statsFound {
		t.Errorf("expected line %q in output (corpus is non-empty)", "## Stats")
	}

	rfc := regexp.MustCompile(`^Generated: \d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`)
	if !lineMatches(out, rfc) {
		t.Errorf("expected a Generated: line matching RFC3339 regex in:\n%s", out)
	}
}

// TestStatsCmd_FormatJSONShape pairs locked decisions §1 (envelope), §2
// (top_tags array-of-objects), §5 (corpus_span sub-object), §8 (--format).
func TestStatsCmd_FormatJSONShape(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	seedListEntry(t, dbPath, "a", "auth", "alpha", "shipped")
	seedListEntry(t, dbPath, "b", "auth,security", "alpha", "shipped")
	seedListEntry(t, dbPath, "c", "perf", "beta", "learned")

	out, errOut, err := runStatsCmd(t, dbPath, "--format", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errOut != "" {
		t.Fatalf("expected empty stderr, got %q", errOut)
	}
	if !strings.HasSuffix(out, "\n") {
		t.Errorf("expected trailing newline (fmt.Fprintln), got %q", firstChars(out, 60))
	}

	body := strings.TrimRight(out, "\n")
	var m map[string]any
	if err := json.Unmarshal([]byte(body), &m); err != nil {
		t.Fatalf("parse json: %v\n%s", err, body)
	}
	wantKeys := []string{
		"generated_at", "scope", "filters", "total_count",
		"entries_per_week", "current_streak", "longest_streak",
		"top_tags", "top_projects", "corpus_span",
	}
	for _, k := range wantKeys {
		if _, ok := m[k]; !ok {
			t.Errorf("expected top-level key %q in output", k)
		}
	}
	if got := m["scope"]; got != "lifetime" {
		t.Errorf("expected scope=lifetime, got %v", got)
	}
	filters, ok := m["filters"].(map[string]any)
	if !ok {
		t.Fatalf("filters not an object: %T", m["filters"])
	}
	if len(filters) != 0 {
		t.Errorf("expected empty filters object, got %v", filters)
	}
	if v, ok := m["total_count"].(float64); !ok || int(v) != 3 {
		t.Errorf("total_count: got %v(%T), want 3", m["total_count"], m["total_count"])
	}

	tags, ok := m["top_tags"].([]any)
	if !ok {
		t.Fatalf("top_tags not an array: %T", m["top_tags"])
	}
	if len(tags) == 0 {
		t.Fatalf("top_tags empty; expected at least 1")
	}
	first, ok := tags[0].(map[string]any)
	if !ok {
		t.Fatalf("top_tags[0] not an object: %T", tags[0])
	}
	if first["tag"] != "auth" {
		t.Errorf("top_tags[0].tag: got %v, want auth", first["tag"])
	}
	if v, ok := first["count"].(float64); !ok || int(v) != 2 {
		t.Errorf("top_tags[0].count: got %v, want 2", first["count"])
	}

	projects, ok := m["top_projects"].([]any)
	if !ok {
		t.Fatalf("top_projects not an array: %T", m["top_projects"])
	}
	if len(projects) < 2 {
		t.Fatalf("top_projects: expected at least 2 entries, got %d", len(projects))
	}
	p0 := projects[0].(map[string]any)
	if p0["project"] != "alpha" {
		t.Errorf("top_projects[0].project: got %v, want alpha", p0["project"])
	}
	if v, ok := p0["count"].(float64); !ok || int(v) != 2 {
		t.Errorf("top_projects[0].count: got %v, want 2", p0["count"])
	}
	p1 := projects[1].(map[string]any)
	if p1["project"] != "beta" {
		t.Errorf("top_projects[1].project: got %v, want beta", p1["project"])
	}
	if v, ok := p1["count"].(float64); !ok || int(v) != 1 {
		t.Errorf("top_projects[1].count: got %v, want 1", p1["count"])
	}

	span, ok := m["corpus_span"].(map[string]any)
	if !ok {
		t.Fatalf("corpus_span not an object: %T", m["corpus_span"])
	}
	for _, k := range []string{"first_entry_date", "last_entry_date", "days"} {
		if _, ok := span[k]; !ok {
			t.Errorf("corpus_span missing key %q", k)
		}
	}
}

// TestStatsCmd_UnknownFormatIsUserError pairs locked decision §8 (DEC-007
// RunE validation; --format accepted set).
func TestStatsCmd_UnknownFormatIsUserError(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	out, _, err := runStatsCmd(t, dbPath, "--format", "yaml")
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
	for _, needle := range []string{"yaml", "markdown", "json"} {
		if !strings.Contains(msg, needle) {
			t.Errorf("expected error to contain %q, got %q", needle, msg)
		}
	}
}

// TestStatsCmd_HelpShowsFormatOnly pairs locked decisions §8 + §9 + the
// SPEC-019-earned Long-vs-help self-audit watch-pattern: positive
// substring assertions for --format/markdown/json; negative substring
// assertions for the eight undeclared flag tokens.
func TestStatsCmd_HelpShowsFormatOnly(t *testing.T) {
	root, outBuf, errBuf := newStatsTestRoot(t)
	root.SetArgs([]string{"stats", "--help"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errBuf.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", errBuf.String())
	}
	out := outBuf.String()
	for _, needle := range []string{"--format", "markdown", "json"} {
		if !strings.Contains(out, needle) {
			t.Errorf("expected help to contain %q, got:\n%s", needle, out)
		}
	}
	forbidden := []string{
		"--tag", "--project", "--type", "--out",
		"--range", "--since", "--week", "--month",
	}
	for _, bad := range forbidden {
		if strings.Contains(out, bad) {
			t.Errorf("help must NOT contain undeclared flag %q, got:\n%s", bad, out)
		}
	}
}

// TestStatsCmd_UndeclaredFlagsRejectedAsUnknown locks decision §9: the
// eight flag names below are GENUINELY UNDECLARED on stats, so cobra
// surfaces them as unknown flag errors (not ErrUser RunE rejections).
func TestStatsCmd_UndeclaredFlagsRejectedAsUnknown(t *testing.T) {
	cases := []struct {
		flag  string
		value string // empty for boolean-style flags
	}{
		{flag: "--tag", value: "x"},
		{flag: "--project", value: "x"},
		{flag: "--type", value: "x"},
		{flag: "--out", value: "x"},
		{flag: "--range", value: "week"},
		{flag: "--since", value: "x"},
		{flag: "--week", value: ""},
		{flag: "--month", value: ""},
	}
	for _, tc := range cases {
		t.Run(tc.flag, func(t *testing.T) {
			dbPath := filepath.Join(t.TempDir(), "test.db")
			args := []string{tc.flag}
			if tc.value != "" {
				args = append(args, tc.value)
			}
			_, _, err := runStatsCmd(t, dbPath, args...)
			if err == nil {
				t.Fatalf("expected error for %s, got nil", tc.flag)
			}
			if errors.Is(err, ErrUser) {
				t.Errorf("expected NOT errors.Is(err, ErrUser) for cobra unknown-flag; got %v", err)
			}
			msg := err.Error()
			if !strings.Contains(msg, "unknown flag") {
				t.Errorf("expected %q in error, got %q", "unknown flag", msg)
			}
			if !strings.Contains(msg, tc.flag) {
				t.Errorf("expected error to name %q, got %q", tc.flag, msg)
			}
		})
	}
}
