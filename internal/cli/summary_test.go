package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

// newSummaryTestRoot builds a fresh root command with the summary
// subcommand attached and isolates BRAGFILE_DB so the host env can't
// leak in. The caller drives args via cmd.SetArgs.
func newSummaryTestRoot(t *testing.T) (*cobra.Command, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	t.Setenv("BRAGFILE_DB", "")
	root := NewRootCmd("test")
	root.AddCommand(NewSummaryCmd())
	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	root.SetOut(outBuf)
	root.SetErr(errBuf)
	return root, outBuf, errBuf
}

func runSummaryCmd(t *testing.T, dbPath string, args ...string) (stdout, stderr string, runErr error) {
	t.Helper()
	root, outBuf, errBuf := newSummaryTestRoot(t)
	full := append([]string{"--db", dbPath, "summary"}, args...)
	root.SetArgs(full)
	runErr = root.Execute()
	return outBuf.String(), errBuf.String(), runErr
}

// TestSummaryCmd_RangeRequiredIsUserError pairs locked decision 4
// (--range required, no default).
func TestSummaryCmd_RangeRequiredIsUserError(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	out, _, err := runSummaryCmd(t, dbPath)
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
	for _, needle := range []string{"--range", "week", "month"} {
		if !strings.Contains(msg, needle) {
			t.Errorf("expected error to contain %q, got %q", needle, msg)
		}
	}
}

// TestSummaryCmd_RangeUnknownValueIsUserError pairs locked decision 4
// (--range validation rejects unknown values).
func TestSummaryCmd_RangeUnknownValueIsUserError(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	out, _, err := runSummaryCmd(t, dbPath, "--range", "yearly")
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
	for _, needle := range []string{"yearly", "week", "month"} {
		if !strings.Contains(msg, needle) {
			t.Errorf("expected error to contain %q, got %q", needle, msg)
		}
	}
}

// TestSummaryCmd_FormatJSON_RangeWeekAndFiltersCompose pairs locked
// decisions 5 (json format on summary) and 6 (filter-flag composition).
func TestSummaryCmd_FormatJSON_RangeWeekAndFiltersCompose(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	hit := seedListEntry(t, dbPath, "in-window-platform-auth", "auth", "platform", "shipped")
	seedListEntry(t, dbPath, "in-window-other-project", "auth", "growth", "shipped")
	seedListEntry(t, dbPath, "in-window-other-tag", "perf", "platform", "shipped")
	seedListEntry(t, dbPath, "in-window-other-type", "auth", "platform", "learned")

	out, errOut, err := runSummaryCmd(t, dbPath,
		"--range", "week",
		"--format", "json",
		"--tag", "auth",
		"--project", "platform",
		"--type", "shipped",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errOut != "" {
		t.Fatalf("expected empty stderr, got %q", errOut)
	}
	if !strings.HasSuffix(out, "\n") {
		t.Errorf("expected trailing newline (fmt.Fprintln), got %q", out)
	}

	body := strings.TrimRight(out, "\n")
	var m map[string]any
	if err := json.Unmarshal([]byte(body), &m); err != nil {
		t.Fatalf("parse json: %v\n%s", err, body)
	}
	wantKeys := []string{"generated_at", "scope", "filters", "counts_by_type", "counts_by_project", "highlights"}
	for _, k := range wantKeys {
		if _, ok := m[k]; !ok {
			t.Errorf("expected top-level key %q in %v", k, m)
		}
	}
	if got := m["scope"]; got != "week" {
		t.Errorf("expected scope=week, got %v", got)
	}
	filters, ok := m["filters"].(map[string]any)
	if !ok {
		t.Fatalf("filters not an object: %T(%v)", m["filters"], m["filters"])
	}
	wantFilters := map[string]string{"tag": "auth", "project": "platform", "type": "shipped"}
	if len(filters) != len(wantFilters) {
		t.Errorf("filters len: want %d, got %d (%v)", len(wantFilters), len(filters), filters)
	}
	for k, v := range wantFilters {
		if filters[k] != v {
			t.Errorf("filters[%q]=%v, want %v", k, filters[k], v)
		}
	}
	cType, ok := m["counts_by_type"].(map[string]any)
	if !ok {
		t.Fatalf("counts_by_type not an object: %T", m["counts_by_type"])
	}
	if v, ok := cType["shipped"].(float64); !ok || int(v) != 1 {
		t.Errorf("counts_by_type[shipped]=want 1, got %v (full: %v)", cType["shipped"], cType)
	}
	if len(cType) != 1 {
		t.Errorf("expected exactly one type bucket, got %v", cType)
	}
	cProj, ok := m["counts_by_project"].(map[string]any)
	if !ok {
		t.Fatalf("counts_by_project not an object: %T", m["counts_by_project"])
	}
	if v, ok := cProj["platform"].(float64); !ok || int(v) != 1 {
		t.Errorf("counts_by_project[platform]=want 1, got %v (full: %v)", cProj["platform"], cProj)
	}
	if len(cProj) != 1 {
		t.Errorf("expected exactly one project bucket, got %v", cProj)
	}
	hl, ok := m["highlights"].([]any)
	if !ok {
		t.Fatalf("highlights not an array: %T", m["highlights"])
	}
	if len(hl) != 1 {
		t.Fatalf("expected 1 highlight group, got %d (%v)", len(hl), hl)
	}
	g, ok := hl[0].(map[string]any)
	if !ok {
		t.Fatalf("highlight group not an object: %T", hl[0])
	}
	if g["project"] != "platform" {
		t.Errorf("expected highlight group project=platform, got %v", g["project"])
	}
	entries, ok := g["entries"].([]any)
	if !ok {
		t.Fatalf("highlight entries not an array: %T", g["entries"])
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry in highlight, got %d", len(entries))
	}
	e, ok := entries[0].(map[string]any)
	if !ok {
		t.Fatalf("highlight entry not an object: %T", entries[0])
	}
	if e["title"] != "in-window-platform-auth" {
		t.Errorf("expected entry title=in-window-platform-auth, got %v", e["title"])
	}
	if id, ok := e["id"].(float64); !ok || int64(id) != hit.ID {
		t.Errorf("expected entry id=%d, got %v", hit.ID, e["id"])
	}
}

// TestSummaryCmd_ScopeFieldAndMarkdownDefault pairs locked decisions
// 1(6) (Scope: line plumbed through from --range) and 5 (markdown
// default). Does NOT backdate entries — backdating-rejected per the
// spec's Rejected alternatives (build-time) §1.
func TestSummaryCmd_ScopeFieldAndMarkdownDefault(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	seedListEntry(t, dbPath, "alpha-fresh", "", "alpha", "shipped")
	seedListEntry(t, dbPath, "beta-fresh", "", "beta", "learned")

	weekOut, weekErr, err := runSummaryCmd(t, dbPath, "--range", "week")
	if err != nil {
		t.Fatalf("week: unexpected error: %v", err)
	}
	if weekErr != "" {
		t.Fatalf("week: expected empty stderr, got %q", weekErr)
	}
	if !strings.HasPrefix(weekOut, "# Bragfile Summary\n\n") {
		t.Errorf("expected markdown default header, got prefix %q", firstChars(weekOut, 60))
	}

	weekScope := false
	weekSummary := false
	weekHighlights := false
	for _, ln := range strings.Split(weekOut, "\n") {
		if ln == "Scope: week" {
			weekScope = true
		}
		if ln == "## Summary" {
			weekSummary = true
		}
		if ln == "## Highlights" {
			weekHighlights = true
		}
	}
	if !weekScope {
		t.Errorf("expected line %q in week output:\n%s", "Scope: week", weekOut)
	}
	if !weekSummary {
		t.Errorf("expected line %q in week output", "## Summary")
	}
	if !weekHighlights {
		t.Errorf("expected line %q in week output", "## Highlights")
	}

	monthOut, monthErr, err := runSummaryCmd(t, dbPath, "--range", "month")
	if err != nil {
		t.Fatalf("month: unexpected error: %v", err)
	}
	if monthErr != "" {
		t.Fatalf("month: expected empty stderr, got %q", monthErr)
	}
	monthScope := false
	monthSummary := false
	monthHighlights := false
	for _, ln := range strings.Split(monthOut, "\n") {
		if ln == "Scope: month" {
			monthScope = true
		}
		if ln == "## Summary" {
			monthSummary = true
		}
		if ln == "## Highlights" {
			monthHighlights = true
		}
	}
	if !monthScope {
		t.Errorf("expected line %q in month output:\n%s", "Scope: month", monthOut)
	}
	if !monthSummary {
		t.Errorf("expected line %q in month output", "## Summary")
	}
	if !monthHighlights {
		t.Errorf("expected line %q in month output", "## Highlights")
	}

	// Generated: line is non-deterministic in CLI tests; assert on
	// the RFC3339 shape per AGENTS.md §9 substring-trap addendum.
	rfc := regexp.MustCompile(`^Generated: \d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`)
	if !lineMatches(weekOut, rfc) {
		t.Errorf("week: expected a Generated: line matching RFC3339 regex in:\n%s", weekOut)
	}
	if !lineMatches(monthOut, rfc) {
		t.Errorf("month: expected a Generated: line matching RFC3339 regex in:\n%s", monthOut)
	}
}

// TestRangeCutoff_WeekMonthArithmeticAndErrors is the pure-helper
// unit test covering the cutoff arithmetic locked at the unit
// layer. Pairs locked decision 1(6) and the Rejected alternatives
// (build-time) §1 prescription (extract pure helper instead of
// adding Store.SetCreatedAtForTesting).
func TestRangeCutoff_WeekMonthArithmeticAndErrors(t *testing.T) {
	now := time.Date(2026, 4, 25, 12, 0, 0, 0, time.UTC)

	t.Run("week", func(t *testing.T) {
		got, err := rangeCutoff("week", now)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := now.AddDate(0, 0, -7)
		if !got.Equal(want) {
			t.Errorf("week: got %v, want %v", got, want)
		}
	})

	t.Run("month", func(t *testing.T) {
		got, err := rangeCutoff("month", now)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := now.AddDate(0, 0, -30)
		if !got.Equal(want) {
			t.Errorf("month: got %v, want %v", got, want)
		}
	})

	t.Run("unknown", func(t *testing.T) {
		got, err := rangeCutoff("yearly", now)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, ErrUser) {
			t.Errorf("expected errors.Is(err, ErrUser); got %v", err)
		}
		if !got.IsZero() {
			t.Errorf("expected zero time on error, got %v", got)
		}
		msg := err.Error()
		for _, needle := range []string{"yearly", "week", "month"} {
			if !strings.Contains(msg, needle) {
				t.Errorf("expected error to contain %q, got %q", needle, msg)
			}
		}
	})

	t.Run("empty", func(t *testing.T) {
		got, err := rangeCutoff("", now)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, ErrUser) {
			t.Errorf("expected errors.Is(err, ErrUser); got %v", err)
		}
		if !got.IsZero() {
			t.Errorf("expected zero time on error, got %v", got)
		}
		msg := err.Error()
		for _, needle := range []string{"--range", "week", "month"} {
			if !strings.Contains(msg, needle) {
				t.Errorf("expected error to contain %q, got %q", needle, msg)
			}
		}
	})
}

// TestSummaryCmd_HelpShowsRangeAndFormat pairs locked decisions 4, 5,
// 6 on the help side. Distinctive needles per AGENTS.md §9
// assertion-specificity rule.
func TestSummaryCmd_HelpShowsRangeAndFormat(t *testing.T) {
	root, outBuf, errBuf := newSummaryTestRoot(t)
	root.SetArgs([]string{"summary", "--help"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errBuf.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", errBuf.String())
	}
	out := outBuf.String()
	for _, needle := range []string{
		"--range", "week", "month",
		"--format", "markdown", "json",
		"--tag", "--project", "--type",
	} {
		if !strings.Contains(out, needle) {
			t.Errorf("expected help to contain %q, got:\n%s", needle, out)
		}
	}
}

// firstChars returns up to n characters of s — used for terse error
// messages without dumping the entire output.
func firstChars(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

// lineMatches reports whether any newline-separated line of s matches
// the regex.
func lineMatches(s string, re *regexp.Regexp) bool {
	for _, ln := range strings.Split(s, "\n") {
		if re.MatchString(ln) {
			return true
		}
	}
	return false
}
