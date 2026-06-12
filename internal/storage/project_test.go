package storage

import (
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
	"time"
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

// backdateProject rewrites a project row's created_at and updated_at via a
// second sql.Open handle (the §9 no-sleep technique for making updated_at
// bump / reorder assertions deterministic without time.Sleep).
func backdateProject(t *testing.T, dbPath string, id int64, at time.Time) {
	t.Helper()
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("backdateProject sql.Open: %v", err)
	}
	defer db.Close()
	ts := at.UTC().Format(time.RFC3339)
	if _, err := db.Exec(
		`UPDATE projects SET created_at = ?, updated_at = ? WHERE id = ?`, ts, ts, id,
	); err != nil {
		t.Fatalf("backdateProject id=%d: %v", id, err)
	}
}

func TestUpdateProject_EditsScalarFieldsAndBumpsUpdatedAt(t *testing.T) {
	s, path := newTestStore(t)

	p, err := s.CreateProject(Project{Name: "bragfile", Status: "active", StateNote: "old note"})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	past := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	backdateProject(t, path, p.ID, past)

	updated, err := s.UpdateProject(p.ID, Project{Name: "bragfile", Status: "paused", StateNote: "x"})
	if err != nil {
		t.Fatalf("UpdateProject: %v", err)
	}
	if updated.Status != "paused" {
		t.Errorf("Status = %q, want %q", updated.Status, "paused")
	}
	if updated.StateNote != "x" {
		t.Errorf("StateNote = %q, want %q", updated.StateNote, "x")
	}
	if !updated.UpdatedAt.After(updated.CreatedAt) {
		t.Errorf("UpdatedAt %v should be After CreatedAt %v (the backdated past)", updated.UpdatedAt, updated.CreatedAt)
	}

	got, err := s.GetProject(p.ID)
	if err != nil {
		t.Fatalf("GetProject: %v", err)
	}
	if got.Status != "paused" {
		t.Errorf("GetProject Status = %q, want %q", got.Status, "paused")
	}
	if !got.UpdatedAt.After(got.CreatedAt) {
		t.Errorf("GetProject: UpdatedAt %v should be After CreatedAt %v", got.UpdatedAt, got.CreatedAt)
	}
}

func TestUpdateProject_BumpReordersListProjects(t *testing.T) {
	s, path := newTestStore(t)

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

	// Backdate to distinct descending instants: p1 most recent → first in list.
	backdateProject(t, path, p1.ID, time.Date(2020, 3, 1, 0, 0, 0, 0, time.UTC))
	backdateProject(t, path, p2.ID, time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC))
	backdateProject(t, path, p3.ID, time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC))

	// Pre-check: order is p1, p2, p3.
	before, err := s.ListProjects()
	if err != nil {
		t.Fatalf("ListProjects before: %v", err)
	}
	if before[0].ID != p1.ID || before[1].ID != p2.ID || before[2].ID != p3.ID {
		t.Fatalf("pre-update order = [%d,%d,%d], want [%d,%d,%d]",
			before[0].ID, before[1].ID, before[2].ID, p1.ID, p2.ID, p3.ID)
	}

	// Update the oldest (p3); its updated_at = now (2026) beats all backdated values.
	if _, err := s.UpdateProject(p3.ID, Project{Name: "third", Status: "active", StateNote: "bumped"}); err != nil {
		t.Fatalf("UpdateProject: %v", err)
	}

	after, err := s.ListProjects()
	if err != nil {
		t.Fatalf("ListProjects after: %v", err)
	}
	if after[0].ID != p3.ID {
		t.Errorf("after update: [0].ID = %d, want %d (p3, bumped to now)", after[0].ID, p3.ID)
	}
}

func TestUpdateProject_RenameDoesNotRewriteEntries(t *testing.T) {
	s, _ := newTestStore(t)

	if _, err := s.Add(Entry{Title: "did a thing", Project: "old"}); err != nil {
		t.Fatalf("Add entry: %v", err)
	}
	proj, err := s.CreateProject(Project{Name: "old"})
	if err != nil {
		t.Fatalf("CreateProject old: %v", err)
	}

	if _, err := s.UpdateProject(proj.ID, Project{Name: "new", Status: "active"}); err != nil {
		t.Fatalf("UpdateProject rename old→new: %v", err)
	}

	oldEntries, err := s.List(ListFilter{Project: "old"})
	if err != nil {
		t.Fatalf("List project=old: %v", err)
	}
	if len(oldEntries) != 1 {
		t.Errorf("List(old) = %d entries, want 1 (entry keeps its captured string)", len(oldEntries))
	}

	newEntries, err := s.List(ListFilter{Project: "new"})
	if err != nil {
		t.Fatalf("List project=new: %v", err)
	}
	if len(newEntries) != 0 {
		t.Errorf("List(new) = %d entries, want 0 (no rewrite)", len(newEntries))
	}
}

func TestUpdateProject_DuplicateNameErrProjectExists(t *testing.T) {
	s, _ := newTestStore(t)

	_, err := s.CreateProject(Project{Name: "a"})
	if err != nil {
		t.Fatalf("CreateProject a: %v", err)
	}
	b, err := s.CreateProject(Project{Name: "b"})
	if err != nil {
		t.Fatalf("CreateProject b: %v", err)
	}

	_, err = s.UpdateProject(b.ID, Project{Name: "a", Status: "active"})
	if !errors.Is(err, ErrProjectExists) {
		t.Fatalf("err = %v, want ErrProjectExists", err)
	}

	got, err := s.GetProject(b.ID)
	if err != nil {
		t.Fatalf("GetProject b: %v", err)
	}
	if got.Name != "b" {
		t.Errorf("Name = %q, want %q (unchanged after duplicate-name attempt)", got.Name, "b")
	}
}

func TestUpdateProject_SameNameSelfRenameAllowed(t *testing.T) {
	s, _ := newTestStore(t)

	a, err := s.CreateProject(Project{Name: "a", StateNote: "n"})
	if err != nil {
		t.Fatalf("CreateProject a: %v", err)
	}

	got, err := s.UpdateProject(a.ID, Project{Name: "a", Status: "active", StateNote: "n2"})
	if err != nil {
		t.Fatalf("UpdateProject self-rename: %v", err)
	}
	if got.StateNote != "n2" {
		t.Errorf("StateNote = %q, want %q", got.StateNote, "n2")
	}
}

func TestUpdateProject_InvalidStatusErrInvalidStatus(t *testing.T) {
	s, _ := newTestStore(t)

	p, err := s.CreateProject(Project{Name: "proj"})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	_, err = s.UpdateProject(p.ID, Project{Name: "proj", Status: "bogus"})
	if !errors.Is(err, ErrInvalidStatus) {
		t.Fatalf("err = %v, want ErrInvalidStatus", err)
	}

	got, err := s.GetProject(p.ID)
	if err != nil {
		t.Fatalf("GetProject: %v", err)
	}
	if got.Status != "active" {
		t.Errorf("Status = %q, want %q (unchanged)", got.Status, "active")
	}
}

func TestUpdateProject_NotFound(t *testing.T) {
	s, _ := newTestStore(t)

	_, err := s.UpdateProject(99999, Project{Name: "fresh-unused", Status: "active"})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestArchiveProject_FlipsStatusNonDestructive(t *testing.T) {
	s, path := newTestStore(t)

	p, err := s.CreateProject(Project{Name: "bragfile", StateNote: "keep"})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	if err := s.AddLocation(p.ID, "/a"); err != nil {
		t.Fatalf("AddLocation /a: %v", err)
	}
	if err := s.AddLocation(p.ID, "/b"); err != nil {
		t.Fatalf("AddLocation /b: %v", err)
	}

	past := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	backdateProject(t, path, p.ID, past)

	if err := s.ArchiveProject(p.ID); err != nil {
		t.Fatalf("ArchiveProject: %v", err)
	}

	got, err := s.GetProject(p.ID)
	if err != nil {
		t.Fatalf("GetProject: %v", err)
	}
	if got.Status != "archived" {
		t.Errorf("Status = %q, want %q", got.Status, "archived")
	}
	if got.StateNote != "keep" {
		t.Errorf("StateNote = %q, want %q (preserved)", got.StateNote, "keep")
	}
	if len(got.Locations) != 2 {
		t.Errorf("Locations = %v, want 2 (preserved)", got.Locations)
	}
	if !got.UpdatedAt.After(got.CreatedAt) {
		t.Errorf("UpdatedAt %v should be After CreatedAt %v (the bump)", got.UpdatedAt, got.CreatedAt)
	}
}

func TestArchiveProject_RecoverableViaUpdate(t *testing.T) {
	s, _ := newTestStore(t)

	p, err := s.CreateProject(Project{Name: "bragfile"})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	if err := s.ArchiveProject(p.ID); err != nil {
		t.Fatalf("ArchiveProject: %v", err)
	}

	if _, err := s.UpdateProject(p.ID, Project{Name: "bragfile", Status: "active"}); err != nil {
		t.Fatalf("UpdateProject back to active: %v", err)
	}

	got, err := s.GetProject(p.ID)
	if err != nil {
		t.Fatalf("GetProject: %v", err)
	}
	if got.Status != "active" {
		t.Errorf("Status = %q, want %q (recovered)", got.Status, "active")
	}
}

func TestArchiveProject_NotFound(t *testing.T) {
	s, _ := newTestStore(t)

	err := s.ArchiveProject(99999)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestDeleteProject_RemovesProjectAndLocations(t *testing.T) {
	s, path := newTestStore(t)

	p, err := s.CreateProject(Project{Name: "bragfile"})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	if err := s.AddLocation(p.ID, "/a"); err != nil {
		t.Fatalf("AddLocation /a: %v", err)
	}
	if err := s.AddLocation(p.ID, "/b"); err != nil {
		t.Fatalf("AddLocation /b: %v", err)
	}

	if err := s.DeleteProject(p.ID); err != nil {
		t.Fatalf("DeleteProject: %v", err)
	}

	_, err = s.GetProject(p.ID)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("GetProject err = %v, want ErrNotFound", err)
	}

	projects, err := s.ListProjects()
	if err != nil {
		t.Fatalf("ListProjects: %v", err)
	}
	if len(projects) != 0 {
		t.Errorf("ListProjects = %d, want 0", len(projects))
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM project_locations WHERE project_id = ?`, p.ID).Scan(&count); err != nil {
		t.Fatalf("count project_locations: %v", err)
	}
	if count != 0 {
		t.Errorf("project_locations count = %d, want 0 (manual in-tx delete)", count)
	}
}

func TestDeleteProject_FreesPathForReuse(t *testing.T) {
	s, _ := newTestStore(t)

	a, err := s.CreateProject(Project{Name: "a"})
	if err != nil {
		t.Fatalf("CreateProject a: %v", err)
	}
	if err := s.AddLocation(a.ID, "/p"); err != nil {
		t.Fatalf("AddLocation /p to a: %v", err)
	}

	if err := s.DeleteProject(a.ID); err != nil {
		t.Fatalf("DeleteProject a: %v", err)
	}

	b, err := s.CreateProject(Project{Name: "b"})
	if err != nil {
		t.Fatalf("CreateProject b: %v", err)
	}
	if err := s.AddLocation(b.ID, "/p"); err != nil {
		t.Errorf("AddLocation /p to b: %v (path should be free for reuse)", err)
	}
}

func TestDeleteProject_LeavesEntriesUntouched(t *testing.T) {
	s, _ := newTestStore(t)

	if _, err := s.Add(Entry{Title: "did a thing", Project: "bragfile"}); err != nil {
		t.Fatalf("Add entry: %v", err)
	}
	proj, err := s.CreateProject(Project{Name: "bragfile"})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	if err := s.DeleteProject(proj.ID); err != nil {
		t.Fatalf("DeleteProject: %v", err)
	}

	entries, err := s.List(ListFilter{Project: "bragfile"})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("List(project=bragfile) = %d entries, want 1 (untouched)", len(entries))
	}
}

func TestDeleteProject_RemovesProjectTaggings(t *testing.T) {
	s, path := newTestStore(t)

	proj, err := s.CreateProject(Project{Name: "bragfile"})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	// Insert a 'project' tagging via a second raw sql handle (nothing in the
	// CLI writes these yet; this is the forward-proof cleanup test per DEC-018).
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	res, err := db.Exec(`INSERT INTO tags(name) VALUES ('test-tag')`)
	if err != nil {
		db.Close()
		t.Fatalf("INSERT tags: %v", err)
	}
	tagID, _ := res.LastInsertId()
	if _, err := db.Exec(
		`INSERT INTO taggings(tag_id, taggable_type, taggable_id, position) VALUES (?, 'project', ?, 0)`,
		tagID, proj.ID,
	); err != nil {
		db.Close()
		t.Fatalf("INSERT taggings: %v", err)
	}
	db.Close()

	if err := s.DeleteProject(proj.ID); err != nil {
		t.Fatalf("DeleteProject: %v", err)
	}

	db2, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("sql.Open (verify): %v", err)
	}
	defer db2.Close()
	var count int
	if err := db2.QueryRow(
		`SELECT COUNT(*) FROM taggings WHERE taggable_type='project' AND taggable_id=?`, proj.ID,
	).Scan(&count); err != nil {
		t.Fatalf("count taggings: %v", err)
	}
	if count != 0 {
		t.Errorf("taggings count = %d, want 0 (in-tx cleanup)", count)
	}
}

func TestDeleteProject_NotFound(t *testing.T) {
	s, _ := newTestStore(t)

	err := s.DeleteProject(99999)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestProjectStatuses_CountsAndExcludesArchived(t *testing.T) {
	s, _ := newTestStore(t)

	bf, err := s.CreateProject(Project{Name: "bragfile"})
	if err != nil {
		t.Fatalf("CreateProject bragfile: %v", err)
	}
	pl, err := s.CreateProject(Project{Name: "platform", Status: "paused"})
	if err != nil {
		t.Fatalf("CreateProject platform: %v", err)
	}
	sp, err := s.CreateProject(Project{Name: "sideproj", Status: "done"})
	if err != nil {
		t.Fatalf("CreateProject sideproj: %v", err)
	}
	old, err := s.CreateProject(Project{Name: "oldthing"})
	if err != nil {
		t.Fatalf("CreateProject oldthing: %v", err)
	}
	if err := s.ArchiveProject(old.ID); err != nil {
		t.Fatalf("ArchiveProject oldthing: %v", err)
	}

	for range 2 {
		if _, err := s.Add(Entry{Title: "t", Project: "bragfile"}); err != nil {
			t.Fatalf("Add bragfile entry: %v", err)
		}
	}
	if _, err := s.Add(Entry{Title: "t", Project: "platform"}); err != nil {
		t.Fatalf("Add platform entry: %v", err)
	}
	if _, err := s.Add(Entry{Title: "t", Project: "oldthing"}); err != nil {
		t.Fatalf("Add oldthing entry: %v", err)
	}

	statuses, err := s.ProjectStatuses()
	if err != nil {
		t.Fatalf("ProjectStatuses: %v", err)
	}
	if len(statuses) != 3 {
		t.Fatalf("len = %d, want 3 (archived oldthing excluded)", len(statuses))
	}

	for _, st := range statuses {
		if st.Status == "archived" {
			t.Errorf("archived project %q should be excluded", st.Name)
		}
	}

	counts := map[string]int{}
	for _, st := range statuses {
		counts[st.Name] = st.BragCount
	}
	if counts["bragfile"] != 2 {
		t.Errorf("bragfile BragCount = %d, want 2", counts["bragfile"])
	}
	if counts["platform"] != 1 {
		t.Errorf("platform BragCount = %d, want 1", counts["platform"])
	}
	if counts["sideproj"] != 0 {
		t.Errorf("sideproj BragCount = %d, want 0", counts["sideproj"])
	}
	_ = bf
	_ = pl
	_ = sp
}

func TestProjectStatuses_SoftJoinIgnoresUnregisteredAndNull(t *testing.T) {
	s, _ := newTestStore(t)

	bf, err := s.CreateProject(Project{Name: "bragfile"})
	if err != nil {
		t.Fatalf("CreateProject bragfile: %v", err)
	}

	// exact match
	if _, err := s.Add(Entry{Title: "t", Project: "bragfile"}); err != nil {
		t.Fatalf("Add bragfile: %v", err)
	}
	// unregistered project string
	if _, err := s.Add(Entry{Title: "t", Project: "bragfile-old"}); err != nil {
		t.Fatalf("Add bragfile-old: %v", err)
	}
	// empty project string
	if _, err := s.Add(Entry{Title: "t", Project: ""}); err != nil {
		t.Fatalf("Add empty project: %v", err)
	}
	// no Project field set (zero-value = empty string)
	if _, err := s.Add(Entry{Title: "t"}); err != nil {
		t.Fatalf("Add no project: %v", err)
	}

	statuses, err := s.ProjectStatuses()
	if err != nil {
		t.Fatalf("ProjectStatuses: %v", err)
	}
	if len(statuses) != 1 {
		t.Fatalf("len = %d, want 1", len(statuses))
	}
	if statuses[0].Name != "bragfile" {
		t.Fatalf("Name = %q, want bragfile", statuses[0].Name)
	}
	if statuses[0].BragCount != 1 {
		t.Errorf("BragCount = %d, want 1 (exact-match only)", statuses[0].BragCount)
	}
	_ = bf
}

func TestProjectStatuses_OrderedByUpdatedAtThenIDDesc(t *testing.T) {
	s, path := newTestStore(t)

	p1, err := s.CreateProject(Project{Name: "p1"})
	if err != nil {
		t.Fatalf("CreateProject p1: %v", err)
	}
	p2, err := s.CreateProject(Project{Name: "p2"})
	if err != nil {
		t.Fatalf("CreateProject p2: %v", err)
	}
	p3, err := s.CreateProject(Project{Name: "p3"})
	if err != nil {
		t.Fatalf("CreateProject p3: %v", err)
	}

	// Sub-case A: all same updated_at → id DESC tie-break: p3, p2, p1
	statuses, err := s.ProjectStatuses()
	if err != nil {
		t.Fatalf("ProjectStatuses (sub-case A): %v", err)
	}
	if len(statuses) != 3 {
		t.Fatalf("len = %d, want 3", len(statuses))
	}
	if statuses[0].ID != p3.ID {
		t.Errorf("sub-case A: [0].ID = %d, want %d (p3)", statuses[0].ID, p3.ID)
	}
	if statuses[1].ID != p2.ID {
		t.Errorf("sub-case A: [1].ID = %d, want %d (p2)", statuses[1].ID, p2.ID)
	}
	if statuses[2].ID != p1.ID {
		t.Errorf("sub-case A: [2].ID = %d, want %d (p1)", statuses[2].ID, p1.ID)
	}

	// Sub-case B: bump p1 via UpdateProject → p1 floats to top
	past := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	backdateProject(t, path, p1.ID, past)
	backdateProject(t, path, p2.ID, past)
	backdateProject(t, path, p3.ID, past)

	if _, err := s.UpdateProject(p1.ID, Project{Name: "p1", Status: "active"}); err != nil {
		t.Fatalf("UpdateProject p1: %v", err)
	}

	statuses2, err := s.ProjectStatuses()
	if err != nil {
		t.Fatalf("ProjectStatuses (sub-case B): %v", err)
	}
	if len(statuses2) != 3 {
		t.Fatalf("len = %d, want 3", len(statuses2))
	}
	if statuses2[0].ID != p1.ID {
		t.Errorf("sub-case B: [0].ID = %d, want %d (p1, bumped to now)", statuses2[0].ID, p1.ID)
	}
}

func TestProjectStatuses_EmptyReturnsNonNilSlice(t *testing.T) {
	s, _ := newTestStore(t)

	statuses, err := s.ProjectStatuses()
	if err != nil {
		t.Fatalf("ProjectStatuses: %v", err)
	}
	if statuses == nil {
		t.Error("expected non-nil slice, got nil")
	}
	if len(statuses) != 0 {
		t.Errorf("len = %d, want 0", len(statuses))
	}
}

func TestProjectForPath_ExactMatch(t *testing.T) {
	s, _ := newTestStore(t)

	created, err := s.CreateProject(Project{Name: "bragfile"})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	if err := s.AddLocation(created.ID, "/home/user/work"); err != nil {
		t.Fatalf("AddLocation: %v", err)
	}

	got, err := s.ProjectForPath("/home/user/work")
	if err != nil {
		t.Fatalf("ProjectForPath: %v", err)
	}
	if got == nil {
		t.Fatal("ProjectForPath: got nil, want non-nil")
	}
	if got.Name != "bragfile" {
		t.Errorf("Name = %q, want %q", got.Name, "bragfile")
	}
}

func TestProjectForPath_SubdirectoryMatch(t *testing.T) {
	s, _ := newTestStore(t)

	created, err := s.CreateProject(Project{Name: "bragfile"})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	if err := s.AddLocation(created.ID, "/home/user/work"); err != nil {
		t.Fatalf("AddLocation: %v", err)
	}

	got, err := s.ProjectForPath("/home/user/work/internal/cli")
	if err != nil {
		t.Fatalf("ProjectForPath: %v", err)
	}
	if got == nil {
		t.Fatal("ProjectForPath: got nil, want non-nil")
	}
	if got.Name != "bragfile" {
		t.Errorf("Name = %q, want %q", got.Name, "bragfile")
	}
}

func TestProjectForPath_NoMatch(t *testing.T) {
	s, _ := newTestStore(t)

	got, err := s.ProjectForPath("/some/random/path")
	if err != nil {
		t.Fatalf("ProjectForPath: unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("ProjectForPath: got %+v, want nil (no locations registered)", got)
	}
}

func TestProjectForPath_LongestPrefixWins(t *testing.T) {
	s, _ := newTestStore(t)

	pA, err := s.CreateProject(Project{Name: "A"})
	if err != nil {
		t.Fatalf("CreateProject A: %v", err)
	}
	if err := s.AddLocation(pA.ID, "/a/b"); err != nil {
		t.Fatalf("AddLocation /a/b: %v", err)
	}

	pB, err := s.CreateProject(Project{Name: "B"})
	if err != nil {
		t.Fatalf("CreateProject B: %v", err)
	}
	if err := s.AddLocation(pB.ID, "/a/b/sub"); err != nil {
		t.Fatalf("AddLocation /a/b/sub: %v", err)
	}

	got, err := s.ProjectForPath("/a/b/sub/deep")
	if err != nil {
		t.Fatalf("ProjectForPath: %v", err)
	}
	if got == nil {
		t.Fatal("ProjectForPath: got nil, want project B")
	}
	if got.Name != "B" {
		t.Errorf("Name = %q, want %q (longest prefix /a/b/sub wins over /a/b)", got.Name, "B")
	}
}

func TestProjectForPath_SeparatorGuard(t *testing.T) {
	s, _ := newTestStore(t)

	created, err := s.CreateProject(Project{Name: "work"})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	if err := s.AddLocation(created.ID, "/home/user/work"); err != nil {
		t.Fatalf("AddLocation: %v", err)
	}

	got, err := s.ProjectForPath("/home/user/worker")
	if err != nil {
		t.Fatalf("ProjectForPath: %v", err)
	}
	if got != nil {
		t.Errorf("ProjectForPath: got %+v, want nil (separator guard prevents /work matching /worker)", got)
	}
}

func TestProjectForPath_NormalizesCleanPath(t *testing.T) {
	s, _ := newTestStore(t)

	created, err := s.CreateProject(Project{Name: "bragfile"})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	if err := s.AddLocation(created.ID, "/home/user/work"); err != nil {
		t.Fatalf("AddLocation: %v", err)
	}

	got, err := s.ProjectForPath("/home/user/work/./src")
	if err != nil {
		t.Fatalf("ProjectForPath: %v", err)
	}
	if got == nil {
		t.Fatal("ProjectForPath: got nil, want non-nil (filepath.Clean normalizes ./)")
	}
	if got.Name != "bragfile" {
		t.Errorf("Name = %q, want %q", got.Name, "bragfile")
	}
}

func TestProjectForPath_MultipleProjectsOneMatch(t *testing.T) {
	s, _ := newTestStore(t)

	pA, err := s.CreateProject(Project{Name: "A"})
	if err != nil {
		t.Fatalf("CreateProject A: %v", err)
	}
	if err := s.AddLocation(pA.ID, "/alpha/repo"); err != nil {
		t.Fatalf("AddLocation /alpha/repo: %v", err)
	}

	pB, err := s.CreateProject(Project{Name: "B"})
	if err != nil {
		t.Fatalf("CreateProject B: %v", err)
	}
	if err := s.AddLocation(pB.ID, "/beta/repo"); err != nil {
		t.Fatalf("AddLocation /beta/repo: %v", err)
	}

	got, err := s.ProjectForPath("/alpha/repo/cmd")
	if err != nil {
		t.Fatalf("ProjectForPath: %v", err)
	}
	if got == nil {
		t.Fatal("ProjectForPath: got nil, want project A")
	}
	if got.Name != "A" {
		t.Errorf("Name = %q, want %q (only A matches; B's /beta/repo is not an ancestor)", got.Name, "A")
	}
}

func TestRemoveLocation_RemovesAttachedPath(t *testing.T) {
	s, _ := newTestStore(t)

	p, err := s.CreateProject(Project{Name: "p"})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	if err := s.AddLocation(p.ID, "/a"); err != nil {
		t.Fatalf("AddLocation: %v", err)
	}
	if err := s.RemoveLocation(p.ID, "/a"); err != nil {
		t.Fatalf("RemoveLocation: %v", err)
	}
	got, err := s.GetProject(p.ID)
	if err != nil {
		t.Fatalf("GetProject: %v", err)
	}
	if len(got.Locations) != 0 {
		t.Errorf("Locations = %v, want empty", got.Locations)
	}
}

func TestRemoveLocation_NotAttachedErrLocationNotFound(t *testing.T) {
	s, _ := newTestStore(t)

	p, err := s.CreateProject(Project{Name: "p"})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	err = s.RemoveLocation(p.ID, "/nope")
	if !errors.Is(err, ErrLocationNotFound) {
		t.Fatalf("err = %v, want ErrLocationNotFound", err)
	}
}

func TestRemoveLocation_OtherProjectErrLocationOtherProject(t *testing.T) {
	s, _ := newTestStore(t)

	p1, err := s.CreateProject(Project{Name: "p1"})
	if err != nil {
		t.Fatalf("CreateProject p1: %v", err)
	}
	if err := s.AddLocation(p1.ID, "/a"); err != nil {
		t.Fatalf("AddLocation: %v", err)
	}
	p2, err := s.CreateProject(Project{Name: "p2"})
	if err != nil {
		t.Fatalf("CreateProject p2: %v", err)
	}
	err = s.RemoveLocation(p2.ID, "/a")
	if !errors.Is(err, ErrLocationOtherProject) {
		t.Fatalf("err = %v, want ErrLocationOtherProject", err)
	}
	got, err := s.GetProject(p1.ID)
	if err != nil {
		t.Fatalf("GetProject p1: %v", err)
	}
	if len(got.Locations) != 1 || got.Locations[0] != "/a" {
		t.Errorf("p1 Locations = %v, want [\"/a\"] (other project's location untouched)", got.Locations)
	}
}

func TestRemoveLocation_VerbatimMatch(t *testing.T) {
	s, _ := newTestStore(t)

	p, err := s.CreateProject(Project{Name: "p"})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	if err := s.AddLocation(p.ID, "/a/b"); err != nil {
		t.Fatalf("AddLocation: %v", err)
	}
	// trailing slash differs from stored "/a/b" — verbatim, not filepath.Clean'd
	err = s.RemoveLocation(p.ID, "/a/b/")
	if !errors.Is(err, ErrLocationNotFound) {
		t.Fatalf("trailing slash: err = %v, want ErrLocationNotFound", err)
	}
	// exact stored string succeeds
	if err := s.RemoveLocation(p.ID, "/a/b"); err != nil {
		t.Fatalf("RemoveLocation exact: %v", err)
	}
}

func TestEditLocations_RepeatableAdds(t *testing.T) {
	s, _ := newTestStore(t)

	p, err := s.CreateProject(Project{Name: "p"})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	if err := s.EditLocations(p.ID, nil, []string{"/a", "/b"}); err != nil {
		t.Fatalf("EditLocations: %v", err)
	}
	got, err := s.GetProject(p.ID)
	if err != nil {
		t.Fatalf("GetProject: %v", err)
	}
	if len(got.Locations) != 2 || got.Locations[0] != "/a" || got.Locations[1] != "/b" {
		t.Errorf("Locations = %v, want [\"/a\",\"/b\"]", got.Locations)
	}
}

func TestEditLocations_RemovesBeforeAdds(t *testing.T) {
	s, _ := newTestStore(t)

	p, err := s.CreateProject(Project{Name: "p"})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	if err := s.AddLocation(p.ID, "/a"); err != nil {
		t.Fatalf("AddLocation: %v", err)
	}
	// remove and re-add /a in same call — no transient UNIQUE collision
	if err := s.EditLocations(p.ID, []string{"/a"}, []string{"/a"}); err != nil {
		t.Fatalf("EditLocations: %v", err)
	}
	got, err := s.GetProject(p.ID)
	if err != nil {
		t.Fatalf("GetProject: %v", err)
	}
	if len(got.Locations) != 1 || got.Locations[0] != "/a" {
		t.Errorf("Locations = %v, want [\"/a\"]", got.Locations)
	}
}

func TestEditLocations_AddDuplicateErrLocationExists(t *testing.T) {
	s, _ := newTestStore(t)

	p1, err := s.CreateProject(Project{Name: "p1"})
	if err != nil {
		t.Fatalf("CreateProject p1: %v", err)
	}
	if err := s.AddLocation(p1.ID, "/a"); err != nil {
		t.Fatalf("AddLocation: %v", err)
	}
	p2, err := s.CreateProject(Project{Name: "p2"})
	if err != nil {
		t.Fatalf("CreateProject p2: %v", err)
	}
	err = s.EditLocations(p2.ID, nil, []string{"/a"})
	if !errors.Is(err, ErrLocationExists) {
		t.Fatalf("err = %v, want ErrLocationExists", err)
	}
}

func TestEditLocations_RollsBackOnOccupiedAdd(t *testing.T) {
	s, path := newTestStore(t)

	owner, err := s.CreateProject(Project{Name: "owner"})
	if err != nil {
		t.Fatalf("CreateProject owner: %v", err)
	}
	if err := s.AddLocation(owner.ID, "/occupied"); err != nil {
		t.Fatalf("AddLocation /occupied: %v", err)
	}
	p, err := s.CreateProject(Project{Name: "p"})
	if err != nil {
		t.Fatalf("CreateProject p: %v", err)
	}
	err = s.EditLocations(p.ID, nil, []string{"/free", "/occupied"})
	if !errors.Is(err, ErrLocationExists) {
		t.Fatalf("err = %v, want ErrLocationExists", err)
	}
	got, err := s.GetProject(p.ID)
	if err != nil {
		t.Fatalf("GetProject p: %v", err)
	}
	if len(got.Locations) != 0 {
		t.Errorf("Locations = %v, want empty (/free was rolled back)", got.Locations)
	}
	// confirm via a second sql.Open handle that /free was not committed
	db2, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("sql.Open second handle: %v", err)
	}
	defer db2.Close()
	var count int
	if err := db2.QueryRow(`SELECT COUNT(*) FROM project_locations WHERE path = '/free'`).Scan(&count); err != nil {
		t.Fatalf("COUNT query: %v", err)
	}
	if count != 0 {
		t.Errorf("COUNT(/free) = %d, want 0 (rollback not persisted)", count)
	}
}

func TestEditLocations_RollsBackOnNotAttachedRemove(t *testing.T) {
	s, _ := newTestStore(t)

	p, err := s.CreateProject(Project{Name: "p"})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	if err := s.AddLocation(p.ID, "/a"); err != nil {
		t.Fatalf("AddLocation: %v", err)
	}
	err = s.EditLocations(p.ID, []string{"/a", "/typo"}, nil)
	if !errors.Is(err, ErrLocationNotFound) {
		t.Fatalf("err = %v, want ErrLocationNotFound", err)
	}
	got, err := s.GetProject(p.ID)
	if err != nil {
		t.Fatalf("GetProject: %v", err)
	}
	if len(got.Locations) != 1 || got.Locations[0] != "/a" {
		t.Errorf("Locations = %v, want [\"/a\"] (first remove rolled back)", got.Locations)
	}
}

func TestEditLocations_DoesNotBumpUpdatedAt(t *testing.T) {
	s, path := newTestStore(t)

	p, err := s.CreateProject(Project{Name: "p"})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	backdated := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	backdateProject(t, path, p.ID, backdated)

	if err := s.EditLocations(p.ID, nil, []string{"/a"}); err != nil {
		t.Fatalf("EditLocations: %v", err)
	}
	got, err := s.GetProject(p.ID)
	if err != nil {
		t.Fatalf("GetProject: %v", err)
	}
	if !got.UpdatedAt.Equal(backdated) {
		t.Errorf("UpdatedAt = %v, want %v (location edit must not bump updated_at)", got.UpdatedAt, backdated)
	}
}
