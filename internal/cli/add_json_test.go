package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jysf/bragfile000/internal/storage"
	"github.com/spf13/cobra"
)

// newRootWithAddAndList builds a root with both `add` and `list`
// subcommands attached, sharing one t.TempDir() DB path. Used by the
// round-trip test which needs to chain `brag list --format json` into
// `brag add --json` against the same DB.
func newRootWithAddAndList(t *testing.T) (*cobra.Command, string) {
	t.Helper()
	root := NewRootCmd("test")
	root.AddCommand(NewAddCmd())
	root.AddCommand(NewListCmd())
	dbPath := t.TempDir() + "/test.db"
	return root, dbPath
}

// seedDirect inserts e via storage.Open + Store.Add against dbPath and
// returns the hydrated row. Used to bypass the CLI surface when tests
// need a known row in the DB.
func seedDirect(t *testing.T, dbPath string, e storage.Entry) storage.Entry {
	t.Helper()
	s, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("storage.Open: %v", err)
	}
	defer s.Close()
	got, err := s.Add(e)
	if err != nil {
		t.Fatalf("Store.Add: %v", err)
	}
	return got
}

// listAll returns every row in dbPath via storage.List with a zero
// filter. Closes the store before returning.
func listAll(t *testing.T, dbPath string) []storage.Entry {
	t.Helper()
	s, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("storage.Open: %v", err)
	}
	defer s.Close()
	got, err := s.List(storage.ListFilter{})
	if err != nil {
		t.Fatalf("Store.List: %v", err)
	}
	return got
}

// TestAddCmd_JSON_RoundTripWithListJSON is the load-bearing test: it
// proves DEC-011 (output) and DEC-012 (input) agree byte-by-byte on
// the user-owned fields, with no `jq del` of server-owned fields. If
// it ever fails, one of the two DECs has drifted.
func TestAddCmd_JSON_RoundTripWithListJSON(t *testing.T) {
	root, dbPath := newRootWithAddAndList(t)

	source := seedDirect(t, dbPath, storage.Entry{
		Title:       "round-trip source",
		Description: "d",
		Tags:        "a,b",
		Project:     "p",
		Type:        "shipped",
		Impact:      "i",
	})

	// Step 1 — render the entry through `brag list --format json`.
	var listOut, listErr bytes.Buffer
	root.SetOut(&listOut)
	root.SetErr(&listErr)
	root.SetArgs([]string{"--db", dbPath, "list", "--format", "json"})
	if err := root.Execute(); err != nil {
		t.Fatalf("brag list --format json: %v", err)
	}
	if listErr.Len() != 0 {
		t.Fatalf("list stderr expected empty, got %q", listErr.String())
	}

	// Step 2 — extract element [0] as raw JSON bytes (jq '.[0]' equiv).
	var arr []json.RawMessage
	if err := json.Unmarshal(listOut.Bytes(), &arr); err != nil {
		t.Fatalf("parse list output: %v", err)
	}
	if len(arr) != 1 {
		t.Fatalf("expected 1 entry from list, got %d", len(arr))
	}
	jsonObj := arr[0]

	// Step 3 — pipe those bytes into `brag add --json` against same DB.
	root2 := NewRootCmd("test")
	root2.AddCommand(NewAddCmd())
	var outBuf, errBuf bytes.Buffer
	root2.SetOut(&outBuf)
	root2.SetErr(&errBuf)
	root2.SetIn(bytes.NewReader(jsonObj))
	root2.SetArgs([]string{"--db", dbPath, "add", "--json"})
	if err := root2.Execute(); err != nil {
		t.Fatalf("brag add --json: %v", err)
	}
	if errBuf.Len() != 0 {
		t.Fatalf("add stderr expected empty, got %q", errBuf.String())
	}
	if !regexp.MustCompile(`^\d+\n$`).MatchString(outBuf.String()) {
		t.Fatalf("add stdout expected to match %q, got %q", `^\d+\n$`, outBuf.String())
	}
	newID, err := strconv.ParseInt(strings.TrimSpace(outBuf.String()), 10, 64)
	if err != nil {
		t.Fatalf("parse new id: %v", err)
	}
	if newID == source.ID {
		t.Fatalf("expected new id != source id %d, got same", source.ID)
	}

	// Step 4 — verify both rows exist with matching user-owned fields.
	entries := listAll(t, dbPath)
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	var copyEntry storage.Entry
	for _, e := range entries {
		if e.ID == newID {
			copyEntry = e
			break
		}
	}
	if copyEntry.ID == 0 {
		t.Fatalf("did not find new entry id=%d in list", newID)
	}
	if copyEntry.Title != source.Title {
		t.Errorf("Title: got %q want %q", copyEntry.Title, source.Title)
	}
	if copyEntry.Description != source.Description {
		t.Errorf("Description: got %q want %q", copyEntry.Description, source.Description)
	}
	if copyEntry.Tags != source.Tags {
		t.Errorf("Tags: got %q want %q", copyEntry.Tags, source.Tags)
	}
	if copyEntry.Project != source.Project {
		t.Errorf("Project: got %q want %q", copyEntry.Project, source.Project)
	}
	if copyEntry.Type != source.Type {
		t.Errorf("Type: got %q want %q", copyEntry.Type, source.Type)
	}
	if copyEntry.Impact != source.Impact {
		t.Errorf("Impact: got %q want %q", copyEntry.Impact, source.Impact)
	}
	if copyEntry.ID == source.ID {
		t.Errorf("expected fresh ID, got source's: %d", copyEntry.ID)
	}
	// Server-field freshness (timestamps) is proven by the stronger
	// frozen-timestamp assertion in TestAddCmd_JSON_ServerFieldsToleratedAndIgnored
	// rather than by time-inequality here — see AGENTS.md §9 freshness-assertion
	// addendum (SPEC-017 ship lesson, 2026-04-24).
}

func TestAddCmd_JSON_ValidInputInsertsEntryAndEmitsID(t *testing.T) {
	root, dbPath := newRootWithAdd(t)
	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetIn(strings.NewReader(`{"title":"j1","description":"body","tags":"a,b","project":"p","type":"shipped","impact":"i"}`))
	root.SetArgs([]string{"--db", dbPath, "add", "--json"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errBuf.Len() != 0 {
		t.Fatalf("expected stderr empty, got %q", errBuf.String())
	}
	if !regexp.MustCompile(`^\d+\n$`).MatchString(outBuf.String()) {
		t.Fatalf("expected stdout to match %q, got %q", `^\d+\n$`, outBuf.String())
	}

	entries := listAll(t, dbPath)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	e := entries[0]
	if e.Title != "j1" {
		t.Errorf("Title: got %q want %q", e.Title, "j1")
	}
	if e.Description != "body" {
		t.Errorf("Description: got %q want %q", e.Description, "body")
	}
	if e.Tags != "a,b" {
		t.Errorf("Tags: got %q want %q", e.Tags, "a,b")
	}
	if e.Project != "p" {
		t.Errorf("Project: got %q want %q", e.Project, "p")
	}
	if e.Type != "shipped" {
		t.Errorf("Type: got %q want %q", e.Type, "shipped")
	}
	if e.Impact != "i" {
		t.Errorf("Impact: got %q want %q", e.Impact, "i")
	}
}

func TestAddCmd_JSON_MissingTitleIsUserError(t *testing.T) {
	root, dbPath := newRootWithAdd(t)
	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetIn(strings.NewReader(`{"description":"orphan"}`))
	root.SetArgs([]string{"--db", dbPath, "add", "--json"})

	err := root.Execute()
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, ErrUser) {
		t.Fatalf("expected errors.Is(err, ErrUser); got %v", err)
	}
	if outBuf.Len() != 0 {
		t.Errorf("expected stdout empty, got %q", outBuf.String())
	}
	if got := len(listAll(t, dbPath)); got != 0 {
		t.Errorf("expected 0 entries, got %d", got)
	}
}

func TestAddCmd_JSON_EmptyTitleIsUserError(t *testing.T) {
	cases := []struct {
		name  string
		stdin string
	}{
		{name: "empty", stdin: `{"title":"","description":"d"}`},
		{name: "whitespace", stdin: `{"title":"   "}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root, dbPath := newRootWithAdd(t)
			var outBuf, errBuf bytes.Buffer
			root.SetOut(&outBuf)
			root.SetErr(&errBuf)
			root.SetIn(strings.NewReader(tc.stdin))
			root.SetArgs([]string{"--db", dbPath, "add", "--json"})

			err := root.Execute()
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !errors.Is(err, ErrUser) {
				t.Fatalf("expected errors.Is(err, ErrUser); got %v", err)
			}
			if outBuf.Len() != 0 {
				t.Errorf("expected stdout empty, got %q", outBuf.String())
			}
			if got := len(listAll(t, dbPath)); got != 0 {
				t.Errorf("expected 0 entries, got %d", got)
			}
		})
	}
}

func TestAddCmd_JSON_UnknownFieldNamedInError(t *testing.T) {
	root, dbPath := newRootWithAdd(t)
	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetIn(strings.NewReader(`{"title":"x","titl":"typo"}`))
	root.SetArgs([]string{"--db", dbPath, "add", "--json"})

	err := root.Execute()
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, ErrUser) {
		t.Fatalf("expected errors.Is(err, ErrUser); got %v", err)
	}
	if !strings.Contains(err.Error(), `unknown field "titl"`) {
		t.Errorf("expected error to contain %q, got %q", `unknown field "titl"`, err.Error())
	}
	if outBuf.Len() != 0 {
		t.Errorf("expected stdout empty, got %q", outBuf.String())
	}
	if got := len(listAll(t, dbPath)); got != 0 {
		t.Errorf("expected 0 entries, got %d", got)
	}
}

func TestAddCmd_JSON_ServerFieldsToleratedAndIgnored(t *testing.T) {
	root, dbPath := newRootWithAdd(t)
	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetIn(strings.NewReader(`{"id":999,"title":"j2","created_at":"2001-01-01T00:00:00Z","updated_at":"2001-01-01T00:00:00Z"}`))
	root.SetArgs([]string{"--db", dbPath, "add", "--json"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	entries := listAll(t, dbPath)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	e := entries[0]
	if e.ID == 999 {
		t.Errorf("expected fresh ID (not 999), got %d", e.ID)
	}
	if e.Title != "j2" {
		t.Errorf("Title: got %q want %q", e.Title, "j2")
	}
	frozen := time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC)
	if e.CreatedAt.Equal(frozen) {
		t.Errorf("CreatedAt: expected fresh, got user's frozen value %s", e.CreatedAt)
	}
	if e.UpdatedAt.Equal(frozen) {
		t.Errorf("UpdatedAt: expected fresh, got user's frozen value %s", e.UpdatedAt)
	}
}

func TestAddCmd_JSON_TagsAsArrayRejectedWithDEC004Reference(t *testing.T) {
	root, dbPath := newRootWithAdd(t)
	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetIn(strings.NewReader(`{"title":"x","tags":["a","b"]}`))
	root.SetArgs([]string{"--db", dbPath, "add", "--json"})

	err := root.Execute()
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, ErrUser) {
		t.Fatalf("expected errors.Is(err, ErrUser); got %v", err)
	}
	if !strings.Contains(err.Error(), "tags must be a comma-joined string") {
		t.Errorf("expected error to contain %q, got %q", "tags must be a comma-joined string", err.Error())
	}
	if outBuf.Len() != 0 {
		t.Errorf("expected stdout empty, got %q", outBuf.String())
	}
	if got := len(listAll(t, dbPath)); got != 0 {
		t.Errorf("expected 0 entries, got %d", got)
	}
}

func TestAddCmd_JSON_ArrayInputIsUserError(t *testing.T) {
	root, dbPath := newRootWithAdd(t)
	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetIn(strings.NewReader(`[{"title":"x"}]`))
	root.SetArgs([]string{"--db", dbPath, "add", "--json"})

	err := root.Execute()
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, ErrUser) {
		t.Fatalf("expected errors.Is(err, ErrUser); got %v", err)
	}
	if outBuf.Len() != 0 {
		t.Errorf("expected stdout empty, got %q", outBuf.String())
	}
	if got := len(listAll(t, dbPath)); got != 0 {
		t.Errorf("expected 0 entries, got %d", got)
	}
}

func TestAddCmd_JSON_MutuallyExclusiveWithFieldFlags(t *testing.T) {
	cases := []struct {
		name      string
		args      []string
		substring string
	}{
		{
			name:      "title",
			args:      []string{"add", "--json", "--title", "x"},
			substring: "--json cannot be combined with --title",
		},
		{
			name:      "description",
			args:      []string{"add", "--json", "--description", "d"},
			substring: "--json cannot be combined with --description",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root, dbPath := newRootWithAdd(t)
			var outBuf, errBuf bytes.Buffer
			root.SetOut(&outBuf)
			root.SetErr(&errBuf)
			// Empty stdin: if the parser ran, this would surface as
			// "invalid JSON" rather than the mutual-exclusion message,
			// so the substring assertion doubles as proof that stdin
			// was NOT consumed on the error path.
			root.SetIn(strings.NewReader(""))
			args := append([]string{"--db", dbPath}, tc.args...)
			root.SetArgs(args)

			err := root.Execute()
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !errors.Is(err, ErrUser) {
				t.Fatalf("expected errors.Is(err, ErrUser); got %v", err)
			}
			if !strings.Contains(err.Error(), tc.substring) {
				t.Errorf("expected error to contain %q, got %q", tc.substring, err.Error())
			}
			if outBuf.Len() != 0 {
				t.Errorf("expected stdout empty, got %q", outBuf.String())
			}
			if got := len(listAll(t, dbPath)); got != 0 {
				t.Errorf("expected 0 entries, got %d", got)
			}
		})
	}
}

func TestAddCmd_JSON_InvalidJSONSyntaxIsUserError(t *testing.T) {
	root, dbPath := newRootWithAdd(t)
	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetIn(strings.NewReader(`{"title":`))
	root.SetArgs([]string{"--db", dbPath, "add", "--json"})

	err := root.Execute()
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, ErrUser) {
		t.Fatalf("expected errors.Is(err, ErrUser); got %v", err)
	}
	if outBuf.Len() != 0 {
		t.Errorf("expected stdout empty, got %q", outBuf.String())
	}
	if got := len(listAll(t, dbPath)); got != 0 {
		t.Errorf("expected 0 entries, got %d", got)
	}
}

// TestAddCmd_JSON_AloneDispatchesToJSONMode covers decision 8 (dispatch
// priority json > flag > editor): `--json` with no field flags would
// otherwise route to editor mode under SPEC-010. testEditFunc is left
// nil so a misroute would fail loudly (editor.Default would try to
// spawn $EDITOR).
func TestAddCmd_JSON_AloneDispatchesToJSONMode(t *testing.T) {
	root, dbPath := newRootWithAdd(t)
	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetIn(strings.NewReader(`{"title":"json-mode-only"}`))
	root.SetArgs([]string{"--db", dbPath, "add", "--json"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errBuf.Len() != 0 {
		t.Fatalf("expected stderr empty, got %q", errBuf.String())
	}
	if !regexp.MustCompile(`^\d+\n$`).MatchString(outBuf.String()) {
		t.Fatalf("expected stdout to match %q, got %q", `^\d+\n$`, outBuf.String())
	}
	entries := listAll(t, dbPath)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Title != "json-mode-only" {
		t.Errorf("Title: got %q want %q", entries[0].Title, "json-mode-only")
	}
}

func TestAddCmd_JSON_HelpShowsJSONFlag(t *testing.T) {
	root, _ := newRootWithAdd(t)
	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"add", "--help"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errBuf.Len() != 0 {
		t.Fatalf("expected stderr empty, got %q", errBuf.String())
	}
	out := outBuf.String()
	if !strings.Contains(out, "--json") {
		t.Errorf("expected help to contain %q, got %q", "--json", out)
	}
	if !strings.Contains(out, "read a single JSON entry from stdin") {
		t.Errorf("expected help to contain %q, got %q", "read a single JSON entry from stdin", out)
	}
}
