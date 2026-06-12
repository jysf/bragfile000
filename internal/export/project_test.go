package export

import (
	"strings"
	"testing"
	"time"

	"github.com/jysf/bragfile000/internal/storage"
)

func makeTestProject(name string, locs []string) storage.Project {
	now := time.Date(2026, 6, 9, 0, 0, 0, 0, time.UTC)
	return storage.Project{
		ID:        1,
		Name:      name,
		Status:    "active",
		StateNote: "",
		Locations: locs,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func TestToProjectsJSON_NakedArrayShape(t *testing.T) {
	p := makeTestProject("bragfile", []string{"/code/bragfile"})
	out, err := ToProjectsJSON([]storage.Project{p})
	if err != nil {
		t.Fatalf("ToProjectsJSON: %v", err)
	}
	s := string(out)

	// naked array
	trimmed := strings.TrimSpace(s)
	if !strings.HasPrefix(trimmed, "[") {
		t.Errorf("expected naked array, got %q", trimmed)
	}

	// 2-space indent — element opens on its own indented line
	if !strings.Contains(s, "\n  {") {
		t.Errorf("expected 2-space indent (\"\\n  {\") in output, got %q", s)
	}
	if !strings.Contains(s, "    \"id\"") {
		t.Errorf("expected 4-space indent for keys (\"    \\\"id\\\"\") in output, got %q", s)
	}

	// seven required keys present in order
	for _, key := range []string{"\"id\"", "\"name\"", "\"status\"", "\"state_note\"", "\"locations\"", "\"created_at\"", "\"updated_at\""} {
		if !strings.Contains(s, key) {
			t.Errorf("missing key %q in output: %q", key, s)
		}
	}

	// locations is a JSON array
	if !strings.Contains(s, "\"locations\": [") {
		t.Errorf("expected locations as JSON array, got %q", s)
	}
}

func TestToProjectsJSON_EmptyIsBracketsNotNull(t *testing.T) {
	// nil input
	outNil, err := ToProjectsJSON(nil)
	if err != nil {
		t.Fatalf("ToProjectsJSON(nil): %v", err)
	}
	if strings.TrimSpace(string(outNil)) != "[]" {
		t.Errorf("nil input: want \"[]\", got %q", string(outNil))
	}

	// empty slice input
	outEmpty, err := ToProjectsJSON([]storage.Project{})
	if err != nil {
		t.Fatalf("ToProjectsJSON(empty): %v", err)
	}
	if strings.TrimSpace(string(outEmpty)) != "[]" {
		t.Errorf("empty slice: want \"[]\", got %q", string(outEmpty))
	}
}

func TestToProjectJSON_SingleObjectAndEmptyLocations(t *testing.T) {
	// with locations
	p := makeTestProject("bragfile", []string{"/code/bragfile"})
	out, err := ToProjectJSON(p)
	if err != nil {
		t.Fatalf("ToProjectJSON: %v", err)
	}
	trimmed := strings.TrimSpace(string(out))
	if !strings.HasPrefix(trimmed, "{") {
		t.Errorf("expected single JSON object, got %q", trimmed)
	}

	// with empty locations — must render [] not null
	p2 := makeTestProject("empty", nil)
	out2, err := ToProjectJSON(p2)
	if err != nil {
		t.Fatalf("ToProjectJSON(empty locs): %v", err)
	}
	if !strings.Contains(string(out2), "\"locations\": []") {
		t.Errorf("expected \"locations\": [] for empty locs, got %q", string(out2))
	}
}

func makeTestProjectStatus(name string, bragCount int, stateNote string) storage.ProjectStatus {
	now := time.Date(2026, 6, 9, 0, 0, 0, 0, time.UTC)
	return storage.ProjectStatus{
		ID:        1,
		Name:      name,
		Status:    "active",
		StateNote: stateNote,
		BragCount: bragCount,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func TestToProjectStatusesJSON_NakedArrayShape(t *testing.T) {
	st := makeTestProjectStatus("bragfile", 7, "")
	out, err := ToProjectStatusesJSON([]storage.ProjectStatus{st})
	if err != nil {
		t.Fatalf("ToProjectStatusesJSON: %v", err)
	}
	s := string(out)

	trimmed := strings.TrimSpace(s)
	if !strings.HasPrefix(trimmed, "[") {
		t.Errorf("expected naked array, got %q", trimmed)
	}

	if !strings.Contains(s, "\n  {") {
		t.Errorf("expected 2-space indent (\"\\n  {\") in output, got %q", s)
	}
	if !strings.Contains(s, "    \"id\"") {
		t.Errorf("expected 4-space indent for keys (\"    \\\"id\\\"\") in output, got %q", s)
	}

	// seven required keys in order
	for _, key := range []string{"\"id\"", "\"name\"", "\"status\"", "\"state_note\"", "\"brag_count\"", "\"created_at\"", "\"updated_at\""} {
		if !strings.Contains(s, key) {
			t.Errorf("missing key %q in output: %q", key, s)
		}
	}

	// brag_count must be a JSON number (not a string)
	if !strings.Contains(s, "\"brag_count\": 7") {
		t.Errorf("expected brag_count as JSON number 7, got %q", s)
	}
}

func TestToProjectStatusesJSON_EmptyIsBracketsNotNull(t *testing.T) {
	outNil, err := ToProjectStatusesJSON(nil)
	if err != nil {
		t.Fatalf("ToProjectStatusesJSON(nil): %v", err)
	}
	if strings.TrimSpace(string(outNil)) != "[]" {
		t.Errorf("nil input: want \"[]\", got %q", string(outNil))
	}

	outEmpty, err := ToProjectStatusesJSON([]storage.ProjectStatus{})
	if err != nil {
		t.Fatalf("ToProjectStatusesJSON(empty): %v", err)
	}
	if strings.TrimSpace(string(outEmpty)) != "[]" {
		t.Errorf("empty slice: want \"[]\", got %q", string(outEmpty))
	}
}

func TestToProjectStatusesJSON_StateNoteNotTruncated(t *testing.T) {
	longNote := strings.Repeat("x", 80)
	st := makeTestProjectStatus("bragfile", 0, longNote)
	out, err := ToProjectStatusesJSON([]storage.ProjectStatus{st})
	if err != nil {
		t.Fatalf("ToProjectStatusesJSON: %v", err)
	}
	if !strings.Contains(string(out), longNote) {
		t.Errorf("expected full 80-char note in JSON output (no truncation), got %q", string(out))
	}
}
