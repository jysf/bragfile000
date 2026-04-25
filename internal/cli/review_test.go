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

// newReviewTestRoot builds a fresh root command with the review
// subcommand attached and isolates BRAGFILE_DB so the host env can't
// leak in. Mirrors newSummaryTestRoot's shape.
func newReviewTestRoot(t *testing.T) (*cobra.Command, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	t.Setenv("BRAGFILE_DB", "")
	root := NewRootCmd("test")
	root.AddCommand(NewReviewCmd())
	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	root.SetOut(outBuf)
	root.SetErr(errBuf)
	return root, outBuf, errBuf
}

func runReviewCmd(t *testing.T, dbPath string, args ...string) (stdout, stderr string, runErr error) {
	t.Helper()
	root, outBuf, errBuf := newReviewTestRoot(t)
	full := append([]string{"--db", dbPath, "review"}, args...)
	root.SetArgs(full)
	runErr = root.Execute()
	return outBuf.String(), errBuf.String(), runErr
}

// TestReviewCmd_BareDefaultsToWeek pairs locked decisions §1 (default
// markdown + scope plumb-through), §4 (named flags), §5 (silent
// default to --week without stderr notice).
func TestReviewCmd_BareDefaultsToWeek(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	seedListEntry(t, dbPath, "alpha-fresh", "", "alpha", "shipped")
	seedListEntry(t, dbPath, "beta-fresh", "", "beta", "learned")

	out, errOut, err := runReviewCmd(t, dbPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errOut != "" {
		t.Fatalf("expected empty stderr (silent default), got %q", errOut)
	}
	if !strings.HasPrefix(out, "# Bragfile Review\n\n") {
		t.Errorf("expected markdown default header, got prefix %q", firstChars(out, 60))
	}

	scopeFound := false
	entriesFound := false
	reflectionFound := false
	for _, ln := range strings.Split(out, "\n") {
		if ln == "Scope: week" {
			scopeFound = true
		}
		if ln == "## Entries" {
			entriesFound = true
		}
		if ln == "## Reflection questions" {
			reflectionFound = true
		}
	}
	if !scopeFound {
		t.Errorf("expected line %q in output:\n%s", "Scope: week", out)
	}
	if !entriesFound {
		t.Errorf("expected line %q in output", "## Entries")
	}
	if !reflectionFound {
		t.Errorf("expected line %q in output", "## Reflection questions")
	}

	rfc := regexp.MustCompile(`^Generated: \d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`)
	if !lineMatches(out, rfc) {
		t.Errorf("expected a Generated: line matching RFC3339 regex in:\n%s", out)
	}
}

// TestReviewCmd_WeekAndMonthMutuallyExclusiveIsUserError pairs locked
// decision §4 (DEC-007 RunE validation, NOT MarkFlagsMutuallyExclusive).
func TestReviewCmd_WeekAndMonthMutuallyExclusiveIsUserError(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	out, _, err := runReviewCmd(t, dbPath, "--week", "--month")
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
	for _, needle := range []string{"--week", "--month", "mutually exclusive"} {
		if !strings.Contains(msg, needle) {
			t.Errorf("expected error to contain %q, got %q", needle, msg)
		}
	}
}

// TestReviewCmd_UnknownFormatIsUserError pairs locked decision §1
// (DEC-014 + DEC-007 — --format validation in RunE).
func TestReviewCmd_UnknownFormatIsUserError(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	out, _, err := runReviewCmd(t, dbPath, "--week", "--format", "yaml")
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

// TestReviewCmd_FormatJSON_MonthScopeAndEntriesGroupedShape pairs
// locked decisions §1, §2, §3, §4, §10.
func TestReviewCmd_FormatJSON_MonthScopeAndEntriesGroupedShape(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	seedListEntry(t, dbPath, "alpha-1", "", "alpha", "shipped")
	seedListEntry(t, dbPath, "alpha-2", "", "alpha", "learned")
	seedListEntry(t, dbPath, "beta-1", "", "beta", "shipped")

	out, errOut, err := runReviewCmd(t, dbPath, "--month", "--format", "json")
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
	wantKeys := []string{"generated_at", "scope", "filters", "entries_grouped", "reflection_questions"}
	for _, k := range wantKeys {
		if _, ok := m[k]; !ok {
			t.Errorf("expected top-level key %q in %v", k, m)
		}
	}
	if got := m["scope"]; got != "month" {
		t.Errorf("expected scope=month, got %v", got)
	}
	filters, ok := m["filters"].(map[string]any)
	if !ok {
		t.Fatalf("filters not an object: %T(%v)", m["filters"], m["filters"])
	}
	if len(filters) != 0 {
		t.Errorf("expected empty filters, got %v", filters)
	}
	groups, ok := m["entries_grouped"].([]any)
	if !ok {
		t.Fatalf("entries_grouped not an array: %T", m["entries_grouped"])
	}
	if len(groups) != 2 {
		t.Fatalf("expected 2 entries_grouped, got %d (%v)", len(groups), groups)
	}
	g0, _ := groups[0].(map[string]any)
	if g0["project"] != "alpha" {
		t.Errorf("expected groups[0].project=alpha, got %v", g0["project"])
	}
	g0Entries, _ := g0["entries"].([]any)
	if len(g0Entries) != 2 {
		t.Errorf("expected groups[0].entries len 2, got %d", len(g0Entries))
	}
	g1, _ := groups[1].(map[string]any)
	if g1["project"] != "beta" {
		t.Errorf("expected groups[1].project=beta, got %v", g1["project"])
	}
	g1Entries, _ := g1["entries"].([]any)
	if len(g1Entries) != 1 {
		t.Errorf("expected groups[1].entries len 1, got %d", len(g1Entries))
	}
	// Full DEC-011 9-key entry shape inside entries[0].
	e0, ok := g0Entries[0].(map[string]any)
	if !ok {
		t.Fatalf("expected entries[0] as object, got %T", g0Entries[0])
	}
	wantEntryKeys := []string{"id", "title", "description", "tags", "project", "type", "impact", "created_at", "updated_at"}
	for _, k := range wantEntryKeys {
		if _, ok := e0[k]; !ok {
			t.Errorf("expected DEC-011 entry key %q, got %v", k, e0)
		}
	}
	rq, ok := m["reflection_questions"].([]any)
	if !ok {
		t.Fatalf("reflection_questions not an array: %T", m["reflection_questions"])
	}
	if len(rq) != 3 {
		t.Errorf("expected reflection_questions len 3, got %d", len(rq))
	}
	for i, q := range rq {
		if _, ok := q.(string); !ok {
			t.Errorf("reflection_questions[%d] is not a string: %T", i, q)
		}
	}
}

// TestReviewCmd_HelpShowsWeekMonthAndFormat pairs locked decisions §4,
// §5, §8.
func TestReviewCmd_HelpShowsWeekMonthAndFormat(t *testing.T) {
	root, outBuf, errBuf := newReviewTestRoot(t)
	root.SetArgs([]string{"review", "--help"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errBuf.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", errBuf.String())
	}
	out := outBuf.String()
	for _, needle := range []string{"--week", "--month", "--format", "markdown", "json"} {
		if !strings.Contains(out, needle) {
			t.Errorf("expected help to contain %q, got:\n%s", needle, out)
		}
	}
	for _, banned := range []string{"--tag", "--project", "--type", "--out", "--since"} {
		if strings.Contains(out, banned) {
			t.Errorf("help must NOT advertise %q (filter/out flags not declared on review); got:\n%s", banned, out)
		}
	}
}

// TestReviewCmd_FilterAndOutFlagsRejectedAsUnknown locks decision §8 —
// filter and out flags are explicitly undeclared, so cobra surfaces
// them as unknown rather than the command rejecting them in RunE.
func TestReviewCmd_FilterAndOutFlagsRejectedAsUnknown(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	cases := []string{"--tag", "--project", "--type", "--out", "--since"}
	for _, flag := range cases {
		t.Run(strings.TrimPrefix(flag, "--"), func(t *testing.T) {
			_, _, err := runReviewCmd(t, dbPath, "--week", flag, "X")
			if err == nil {
				t.Fatalf("expected error for %s, got nil", flag)
			}
			if errors.Is(err, ErrUser) {
				t.Errorf("expected !errors.Is(err, ErrUser) for cobra unknown-flag path; got %v", err)
			}
			msg := err.Error()
			if !strings.Contains(msg, "unknown flag") {
				t.Errorf("expected %q in error, got %q", "unknown flag", msg)
			}
			if !strings.Contains(msg, flag) {
				t.Errorf("expected %q in error, got %q", flag, msg)
			}
		})
	}
}
