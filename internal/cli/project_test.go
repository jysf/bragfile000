package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/jysf/bragfile000/internal/storage"
	"github.com/spf13/cobra"
)

func newProjectTestRoot(t *testing.T) (*cobra.Command, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	t.Setenv("BRAGFILE_DB", "")
	root := NewRootCmd("test")
	root.AddCommand(NewProjectCmd())
	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	root.SetOut(outBuf)
	root.SetErr(errBuf)
	return root, outBuf, errBuf
}

func runProjectCmd(t *testing.T, dbPath string, args ...string) (stdout, stderr string, runErr error) {
	t.Helper()
	root, outBuf, errBuf := newProjectTestRoot(t)
	full := append([]string{"--db", dbPath, "project"}, args...)
	root.SetArgs(full)
	runErr = root.Execute()
	return outBuf.String(), errBuf.String(), runErr
}

func TestProjectCmd_BarePrintsHelp(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	out, errOut, err := runProjectCmd(t, dbPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errOut != "" {
		t.Errorf("expected empty stderr, got %q", errOut)
	}
	if !strings.Contains(out, "Usage:") {
		t.Errorf("expected 'Usage:' in help output, got %q", out)
	}
	for _, sub := range []string{"new", "list", "show"} {
		if !strings.Contains(out, sub) {
			t.Errorf("expected subcommand %q in help output, got %q", sub, out)
		}
	}
}

func TestProjectNew_CreatesAndAttaches(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	out, errOut, err := runProjectCmd(t, dbPath, "new", "bragfile", "--path", "/tmp/x")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "" {
		t.Errorf("expected empty stdout on new, got %q", out)
	}
	if !strings.Contains(errOut, `Created project "bragfile".`) {
		t.Errorf("expected confirmation on stderr, got %q", errOut)
	}

	// verify via show
	showOut, showErr, showRunErr := runProjectCmd(t, dbPath, "show", "bragfile")
	if showRunErr != nil {
		t.Fatalf("show: unexpected error: %v", showRunErr)
	}
	if showErr != "" {
		t.Errorf("show: expected empty stderr, got %q", showErr)
	}
	if !strings.Contains(showOut, "Status: active") {
		t.Errorf("show: expected 'Status: active', got %q", showOut)
	}
	if !strings.Contains(showOut, "/tmp/x") {
		t.Errorf("show: expected /tmp/x in locations, got %q", showOut)
	}
}

func TestProjectNew_RequiresPath(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	_, _, err := runProjectCmd(t, dbPath, "new", "bragfile")
	if !errors.Is(err, ErrUser) {
		t.Fatalf("expected ErrUser, got %v", err)
	}

	// nothing created
	listOut, _, listErr := runProjectCmd(t, dbPath, "list")
	if listErr != nil {
		t.Fatalf("list: unexpected error: %v", listErr)
	}
	if strings.TrimSpace(listOut) != "" {
		t.Errorf("expected empty list after failed new, got %q", listOut)
	}
}

func TestProjectNew_EmptyNameErrUser(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	_, _, err := runProjectCmd(t, dbPath, "new", "", "--path", "/tmp/x")
	if !errors.Is(err, ErrUser) {
		t.Fatalf("expected ErrUser for empty name, got %v", err)
	}
}

func TestProjectNew_DuplicateNameErrUser(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	_, _, err := runProjectCmd(t, dbPath, "new", "bragfile", "--path", "/tmp/a")
	if err != nil {
		t.Fatalf("first new: unexpected error: %v", err)
	}

	_, _, err = runProjectCmd(t, dbPath, "new", "bragfile", "--path", "/tmp/b")
	if !errors.Is(err, ErrUser) {
		t.Fatalf("expected ErrUser on duplicate name, got %v", err)
	}

	// check that error message names the project
	_, _, err = runProjectCmd(t, dbPath, "new", "bragfile", "--path", "/tmp/b")
	if err == nil || !strings.Contains(err.Error(), "bragfile") {
		t.Errorf("expected error naming 'bragfile', got %v", err)
	}

	// exactly one row
	listOut, _, listErr := runProjectCmd(t, dbPath, "list")
	if listErr != nil {
		t.Fatalf("list: %v", listErr)
	}
	lines := strings.Split(strings.TrimRight(listOut, "\n"), "\n")
	if len(lines) != 1 || lines[0] == "" {
		t.Errorf("expected exactly 1 project, got %d lines: %q", len(lines), listOut)
	}
}

func TestProjectNew_PathAlreadyRegisteredErrUser_NoOrphan(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	_, _, err := runProjectCmd(t, dbPath, "new", "projA", "--path", "/p")
	if err != nil {
		t.Fatalf("new projA: %v", err)
	}

	_, _, err = runProjectCmd(t, dbPath, "new", "projB", "--path", "/p")
	if !errors.Is(err, ErrUser) {
		t.Fatalf("expected ErrUser on duplicate path, got %v", err)
	}
	if !strings.Contains(err.Error(), "/p") {
		t.Errorf("expected error mentioning path /p, got %v", err)
	}

	// projB must NOT have been created
	_, _, showErr := runProjectCmd(t, dbPath, "show", "projB")
	if !errors.Is(showErr, ErrUser) {
		t.Errorf("expected ErrUser on show projB (not created), got %v", showErr)
	}

	// exactly one project in list
	listOut, _, listErr := runProjectCmd(t, dbPath, "list")
	if listErr != nil {
		t.Fatalf("list: %v", listErr)
	}
	lines := strings.Split(strings.TrimRight(listOut, "\n"), "\n")
	if len(lines) != 1 || !strings.HasPrefix(lines[0], "projA") {
		t.Errorf("expected exactly projA in list, got %q", listOut)
	}
}

func TestProjectNew_StdoutStderrSeparation(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	out, errOut, err := runProjectCmd(t, dbPath, "new", "bragfile", "--path", "/tmp/x")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "" {
		t.Errorf("stdout must be empty on success new, got %q", out)
	}
	if !strings.Contains(errOut, "Created project") {
		t.Errorf("confirmation must be on stderr, got errOut=%q", errOut)
	}
}

func TestProjectList_PlainOrderingAndShape(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	for _, args := range [][]string{
		{"new", "p1", "--path", "/p1"},
		{"new", "p2", "--path", "/p2"},
		{"new", "p3", "--path", "/p3"},
	} {
		if _, _, err := runProjectCmd(t, dbPath, args...); err != nil {
			t.Fatalf("new %v: %v", args, err)
		}
	}

	out, errOut, err := runProjectCmd(t, dbPath, "list")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if errOut != "" {
		t.Errorf("expected empty stderr, got %q", errOut)
	}

	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d: %q", len(lines), out)
	}
	// newest-created first (id DESC): p3, p2, p1
	wantOrder := []string{"p3", "p2", "p1"}
	for i, want := range wantOrder {
		if !strings.HasPrefix(lines[i], want+"\t") {
			t.Errorf("line[%d] = %q, want prefix %q", i, lines[i], want+"\t")
		}
	}
	// check shape: name<TAB>status<TAB>path
	cols := strings.Split(lines[0], "\t")
	if len(cols) != 3 {
		t.Errorf("expected 3 tab-separated columns, got %d: %q", len(cols), lines[0])
	}
	if cols[1] != "active" {
		t.Errorf("col[1] (status) = %q, want %q", cols[1], "active")
	}
}

func TestProjectList_JSON(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	if _, _, err := runProjectCmd(t, dbPath, "new", "bragfile", "--path", "/code/bragfile"); err != nil {
		t.Fatalf("new: %v", err)
	}

	out, errOut, err := runProjectCmd(t, dbPath, "list", "--format", "json")
	if err != nil {
		t.Fatalf("list --format json: %v", err)
	}
	if errOut != "" {
		t.Errorf("expected empty stderr, got %q", errOut)
	}

	trimmed := strings.TrimSpace(out)
	if !strings.HasPrefix(trimmed, "[") {
		t.Errorf("expected naked array, got %q", trimmed)
	}

	var got []struct {
		ID        int64    `json:"id"`
		Name      string   `json:"name"`
		Status    string   `json:"status"`
		StateNote string   `json:"state_note"`
		Locations []string `json:"locations"`
		CreatedAt string   `json:"created_at"`
		UpdatedAt string   `json:"updated_at"`
	}
	if err := json.Unmarshal([]byte(trimmed), &got); err != nil {
		t.Fatalf("JSON unmarshal: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 element, got %d", len(got))
	}
	if got[0].Name != "bragfile" {
		t.Errorf("Name = %q, want bragfile", got[0].Name)
	}
	if len(got[0].Locations) != 1 || got[0].Locations[0] != "/code/bragfile" {
		t.Errorf("Locations = %v, want [\"/code/bragfile\"]", got[0].Locations)
	}

	// 2-space indent check
	if !strings.Contains(out, "    \"id\"") {
		t.Errorf("expected 4-space indent for keys in JSON, got %q", out)
	}
}

func TestProjectList_EmptyJSONIsBrackets(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")

	// JSON mode: emits []
	out, _, err := runProjectCmd(t, dbPath, "list", "--format", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(out) != "[]" {
		t.Errorf("expected [], got %q", out)
	}

	// plain mode: empty stdout
	out2, _, err2 := runProjectCmd(t, dbPath, "list")
	if err2 != nil {
		t.Fatalf("unexpected error: %v", err2)
	}
	if strings.TrimSpace(out2) != "" {
		t.Errorf("expected empty plain output, got %q", out2)
	}
}

func TestProjectList_UnknownFormatErrUser(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	_, _, err := runProjectCmd(t, dbPath, "list", "--format", "xml")
	if !errors.Is(err, ErrUser) {
		t.Fatalf("expected ErrUser for unknown format, got %v", err)
	}
}

func TestProjectShow_ByName(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	if _, _, err := runProjectCmd(t, dbPath, "new", "bragfile", "--path", "/tmp/x"); err != nil {
		t.Fatalf("new: %v", err)
	}

	out, errOut, err := runProjectCmd(t, dbPath, "show", "bragfile")
	if err != nil {
		t.Fatalf("show: %v", err)
	}
	if errOut != "" {
		t.Errorf("expected empty stderr, got %q", errOut)
	}
	if !strings.Contains(out, "Name: bragfile") {
		t.Errorf("missing 'Name: bragfile', got %q", out)
	}
	if !strings.Contains(out, "Status: active") {
		t.Errorf("missing 'Status: active', got %q", out)
	}
	if !strings.Contains(out, "State note:") {
		t.Errorf("missing 'State note:', got %q", out)
	}
	if !strings.Contains(out, "Locations:") {
		t.Errorf("missing 'Locations:', got %q", out)
	}
	if !strings.Contains(out, "  /tmp/x") {
		t.Errorf("missing location '  /tmp/x', got %q", out)
	}
}

func TestProjectShow_ById(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	if _, _, err := runProjectCmd(t, dbPath, "new", "bragfile", "--path", "/tmp/x"); err != nil {
		t.Fatalf("new: %v", err)
	}

	s, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("Open store: %v", err)
	}
	defer s.Close()
	projects, err := s.ListProjects()
	if err != nil {
		t.Fatalf("ListProjects: %v", err)
	}
	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(projects))
	}
	idStr := strconv.FormatInt(projects[0].ID, 10)

	out, _, err := runProjectCmd(t, dbPath, "show", idStr)
	if err != nil {
		t.Fatalf("show by id %q: %v", idStr, err)
	}
	if !strings.Contains(out, "Name: bragfile") {
		t.Errorf("expected 'Name: bragfile', got %q", out)
	}
}

func TestProjectShow_NameFirstResolution(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")

	// Create a project named "1" — its id is 1
	if _, _, err := runProjectCmd(t, dbPath, "new", "1", "--path", "/one"); err != nil {
		t.Fatalf("new '1': %v", err)
	}
	// Create another project with id=2 (or whatever comes next)
	if _, _, err := runProjectCmd(t, dbPath, "new", "other", "--path", "/other"); err != nil {
		t.Fatalf("new 'other': %v", err)
	}

	// show 1 must return the project *named* "1", not project id 1 (same here, but
	// the key check is that the name path is taken first)
	out, _, err := runProjectCmd(t, dbPath, "show", "1")
	if err != nil {
		t.Fatalf("show '1': %v", err)
	}
	// The project named "1" has path /one
	if !strings.Contains(out, "/one") {
		t.Errorf("show '1' should return project named '1' (path /one), got %q", out)
	}
}

func TestProjectShow_JSONSingleObject(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	if _, _, err := runProjectCmd(t, dbPath, "new", "bragfile", "--path", "/code/bragfile"); err != nil {
		t.Fatalf("new: %v", err)
	}

	out, errOut, err := runProjectCmd(t, dbPath, "show", "bragfile", "--format", "json")
	if err != nil {
		t.Fatalf("show --format json: %v", err)
	}
	if errOut != "" {
		t.Errorf("expected empty stderr, got %q", errOut)
	}

	trimmed := strings.TrimSpace(out)
	if !strings.HasPrefix(trimmed, "{") {
		t.Errorf("expected single JSON object, got %q", trimmed)
	}

	var got struct {
		Locations []string `json:"locations"`
	}
	if err := json.Unmarshal([]byte(trimmed), &got); err != nil {
		t.Fatalf("JSON unmarshal: %v", err)
	}
	if len(got.Locations) != 1 || got.Locations[0] != "/code/bragfile" {
		t.Errorf("Locations = %v, want [\"/code/bragfile\"]", got.Locations)
	}
}

func TestProjectShow_NotFoundErrUser(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")

	_, _, err := runProjectCmd(t, dbPath, "show", "nonexistent")
	if !errors.Is(err, ErrUser) {
		t.Fatalf("expected ErrUser for nonexistent name, got %v", err)
	}

	_, _, err = runProjectCmd(t, dbPath, "show", "99999")
	if !errors.Is(err, ErrUser) {
		t.Fatalf("expected ErrUser for numeric nonexistent id, got %v", err)
	}
}

func TestProjectShow_UnknownFormatErrUser(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	if _, _, err := runProjectCmd(t, dbPath, "new", "bragfile", "--path", "/tmp/x"); err != nil {
		t.Fatalf("new: %v", err)
	}

	_, _, err := runProjectCmd(t, dbPath, "show", "bragfile", "--format", "xml")
	if !errors.Is(err, ErrUser) {
		t.Fatalf("expected ErrUser for unknown format, got %v", err)
	}
}

func TestProjectNew_HelpShowsExamples(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")

	// new --help
	newOut, _, _ := runProjectCmd(t, dbPath, "new", "--help")
	if !strings.Contains(newOut, "Examples:") {
		t.Errorf("new --help missing 'Examples:', got %q", newOut)
	}
	if !strings.Contains(newOut, "brag project new") {
		t.Errorf("new --help missing distinctive 'brag project new' token, got %q", newOut)
	}

	// list --help
	listOut, _, _ := runProjectCmd(t, dbPath, "list", "--help")
	if !strings.Contains(listOut, "Examples:") {
		t.Errorf("list --help missing 'Examples:', got %q", listOut)
	}
	if !strings.Contains(listOut, "brag project list --format json") {
		t.Errorf("list --help missing 'brag project list --format json', got %q", listOut)
	}
}
