package storage

import (
	"errors"
	"path/filepath"
	"testing"
)

func TestGetProjectByName_RoundTrip(t *testing.T) {
	s, _ := newTestStore(t)

	created, err := s.CreateProject(Project{Name: "bragfile", StateNote: "n"})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	if err := s.AddLocation(created.ID, "/a"); err != nil {
		t.Fatalf("AddLocation: %v", err)
	}

	got, err := s.GetProjectByName("bragfile")
	if err != nil {
		t.Fatalf("GetProjectByName: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("ID = %d, want %d", got.ID, created.ID)
	}
	if got.Name != "bragfile" {
		t.Errorf("Name = %q, want %q", got.Name, "bragfile")
	}
	if got.Status != "active" {
		t.Errorf("Status = %q, want %q", got.Status, "active")
	}
	if got.StateNote != "n" {
		t.Errorf("StateNote = %q, want %q", got.StateNote, "n")
	}
	if len(got.Locations) != 1 || got.Locations[0] != "/a" {
		t.Errorf("Locations = %v, want [\"/a\"]", got.Locations)
	}
}

func TestGetProjectByName_NotFound(t *testing.T) {
	s, _ := newTestStore(t)

	_, err := s.GetProjectByName("nope")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestCreateProject_RoundTrip(t *testing.T) {
	s, _ := newTestStore(t)

	p, err := s.CreateProject(Project{
		Name:      "bragfile",
		Status:    "active",
		StateNote: "next: cut v0.2.0",
	})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	if p.ID <= 0 {
		t.Fatalf("ID = %d, want > 0", p.ID)
	}
	if p.CreatedAt.IsZero() {
		t.Fatal("CreatedAt is zero")
	}
	if !p.CreatedAt.Equal(p.UpdatedAt) {
		t.Errorf("CreatedAt %v != UpdatedAt %v on fresh insert", p.CreatedAt, p.UpdatedAt)
	}

	got, err := s.GetProject(p.ID)
	if err != nil {
		t.Fatalf("GetProject: %v", err)
	}
	if got.Name != "bragfile" {
		t.Errorf("Name = %q, want %q", got.Name, "bragfile")
	}
	if got.Status != "active" {
		t.Errorf("Status = %q, want %q", got.Status, "active")
	}
	if got.StateNote != "next: cut v0.2.0" {
		t.Errorf("StateNote = %q, want %q", got.StateNote, "next: cut v0.2.0")
	}
}

func TestCreateProject_StatusDefaultsActive(t *testing.T) {
	s, _ := newTestStore(t)

	p, err := s.CreateProject(Project{Name: "empty-status"})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	got, err := s.GetProject(p.ID)
	if err != nil {
		t.Fatalf("GetProject: %v", err)
	}
	if got.Status != "active" {
		t.Errorf("Status = %q, want %q", got.Status, "active")
	}
	if got.StateNote != "" {
		t.Errorf("StateNote = %q, want empty", got.StateNote)
	}
}

func TestCreateProject_DuplicateNameErrProjectExists(t *testing.T) {
	s, _ := newTestStore(t)

	_, err := s.CreateProject(Project{Name: "dup"})
	if err != nil {
		t.Fatalf("first CreateProject: %v", err)
	}

	_, err = s.CreateProject(Project{Name: "dup"})
	if !errors.Is(err, ErrProjectExists) {
		t.Fatalf("err = %v, want ErrProjectExists", err)
	}

	projects, err := s.ListProjects()
	if err != nil {
		t.Fatalf("ListProjects: %v", err)
	}
	if len(projects) != 1 {
		t.Errorf("ListProjects len = %d, want 1 (no partial insert)", len(projects))
	}
}

func TestGetProject_NotFound(t *testing.T) {
	s, _ := newTestStore(t)

	_, err := s.GetProject(99999)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestAddLocation_RoundTripOrderedByID(t *testing.T) {
	s, _ := newTestStore(t)

	p, err := s.CreateProject(Project{Name: "myproj"})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	if err := s.AddLocation(p.ID, "/a"); err != nil {
		t.Fatalf("AddLocation /a: %v", err)
	}
	if err := s.AddLocation(p.ID, "/b"); err != nil {
		t.Fatalf("AddLocation /b: %v", err)
	}

	got, err := s.GetProject(p.ID)
	if err != nil {
		t.Fatalf("GetProject: %v", err)
	}
	want := []string{"/a", "/b"}
	if len(got.Locations) != len(want) {
		t.Fatalf("Locations = %v, want %v", got.Locations, want)
	}
	for i, loc := range got.Locations {
		if loc != want[i] {
			t.Errorf("Locations[%d] = %q, want %q", i, loc, want[i])
		}
	}
}

func TestAddLocation_DuplicatePathErrLocationExists_SameProject(t *testing.T) {
	s, _ := newTestStore(t)

	p, err := s.CreateProject(Project{Name: "proj1"})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	if err := s.AddLocation(p.ID, "/a"); err != nil {
		t.Fatalf("first AddLocation: %v", err)
	}

	err = s.AddLocation(p.ID, "/a")
	if !errors.Is(err, ErrLocationExists) {
		t.Fatalf("err = %v, want ErrLocationExists", err)
	}
}

func TestAddLocation_DuplicatePathErrLocationExists_DifferentProject(t *testing.T) {
	s, _ := newTestStore(t)

	p1, err := s.CreateProject(Project{Name: "proj-one"})
	if err != nil {
		t.Fatalf("CreateProject proj-one: %v", err)
	}
	p2, err := s.CreateProject(Project{Name: "proj-two"})
	if err != nil {
		t.Fatalf("CreateProject proj-two: %v", err)
	}

	if err := s.AddLocation(p1.ID, "/a"); err != nil {
		t.Fatalf("AddLocation to proj-one: %v", err)
	}

	err = s.AddLocation(p2.ID, "/a")
	if !errors.Is(err, ErrLocationExists) {
		t.Fatalf("err = %v, want ErrLocationExists (global uniqueness)", err)
	}
}

func TestListProjects_OrderedByUpdatedAtThenIDDesc(t *testing.T) {
	s, _ := newTestStore(t)

	// Create three projects in a single second — same updated_at — so the
	// id DESC tie-break is the only ordering signal (§9 no-sleep rule).
	p1, err := s.CreateProject(Project{Name: "first"})
	if err != nil {
		t.Fatalf("CreateProject first: %v", err)
	}
	p2, err := s.CreateProject(Project{Name: "second"})
	if err != nil {
		t.Fatalf("CreateProject second: %v", err)
	}
	p3, err := s.CreateProject(Project{Name: "third"})
	if err != nil {
		t.Fatalf("CreateProject third: %v", err)
	}

	projects, err := s.ListProjects()
	if err != nil {
		t.Fatalf("ListProjects: %v", err)
	}
	if len(projects) != 3 {
		t.Fatalf("len = %d, want 3", len(projects))
	}
	// Newest id (p3) first, then p2, then p1.
	if projects[0].ID != p3.ID {
		t.Errorf("[0].ID = %d, want %d (p3)", projects[0].ID, p3.ID)
	}
	if projects[1].ID != p2.ID {
		t.Errorf("[1].ID = %d, want %d (p2)", projects[1].ID, p2.ID)
	}
	if projects[2].ID != p1.ID {
		t.Errorf("[2].ID = %d, want %d (p1)", projects[2].ID, p1.ID)
	}
}

func TestListProjects_HydratesLocations(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "db.sqlite")
	s, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	p, err := s.CreateProject(Project{Name: "located"})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	if err := s.AddLocation(p.ID, "/home/user/located"); err != nil {
		t.Fatalf("AddLocation 1: %v", err)
	}
	if err := s.AddLocation(p.ID, "/work/located"); err != nil {
		t.Fatalf("AddLocation 2: %v", err)
	}

	projects, err := s.ListProjects()
	if err != nil {
		t.Fatalf("ListProjects: %v", err)
	}
	if len(projects) != 1 {
		t.Fatalf("len = %d, want 1", len(projects))
	}
	if len(projects[0].Locations) != 2 {
		t.Fatalf("Locations = %v, want 2 entries", projects[0].Locations)
	}
	if projects[0].Locations[0] != "/home/user/located" {
		t.Errorf("Locations[0] = %q, want %q", projects[0].Locations[0], "/home/user/located")
	}
	if projects[0].Locations[1] != "/work/located" {
		t.Errorf("Locations[1] = %q, want %q", projects[0].Locations[1], "/work/located")
	}
}
