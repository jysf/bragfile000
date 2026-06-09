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
