package storage

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func TestOpen_ProjectsTablesExist(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "db.sqlite")

	s, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	db := openRawDB(t, path)

	for _, tbl := range []string{"projects", "project_locations"} {
		if !objectExists(t, db, "table", tbl) {
			t.Errorf("table %q does not exist after Open", tbl)
		}
	}

	// Raw insert honoring the schema succeeds; status/state_note get defaults.
	_, err = db.Exec(
		`INSERT INTO projects (name, created_at, updated_at) VALUES (?, ?, ?)`,
		"raw-test", "2026-06-08T00:00:00Z", "2026-06-08T00:00:00Z",
	)
	if err != nil {
		t.Fatalf("raw INSERT into projects: %v", err)
	}

	var status, stateNote string
	if err := db.QueryRow(`SELECT status, state_note FROM projects WHERE name = ?`, "raw-test").
		Scan(&status, &stateNote); err != nil {
		t.Fatalf("SELECT status, state_note: %v", err)
	}
	if status != "active" {
		t.Errorf("status = %q, want %q", status, "active")
	}
	if stateNote != "" {
		t.Errorf("state_note = %q, want empty", stateNote)
	}

	// Duplicate name must fail.
	_, err = db.Exec(
		`INSERT INTO projects (name, created_at, updated_at) VALUES (?, ?, ?)`,
		"raw-test", "2026-06-08T00:00:00Z", "2026-06-08T00:00:00Z",
	)
	if err == nil {
		t.Error("duplicate name INSERT should have failed")
	}

	// Insert a location row.
	var projectID int64
	if err := db.QueryRow(`SELECT id FROM projects WHERE name = ?`, "raw-test").Scan(&projectID); err != nil {
		t.Fatalf("get projectID: %v", err)
	}
	_, err = db.Exec(
		`INSERT INTO project_locations (project_id, path) VALUES (?, ?)`,
		projectID, "/tmp/raw-test",
	)
	if err != nil {
		t.Fatalf("raw INSERT into project_locations: %v", err)
	}

	// Duplicate path must fail.
	_, err = db.Exec(
		`INSERT INTO project_locations (project_id, path) VALUES (?, ?)`,
		projectID, "/tmp/raw-test",
	)
	if err == nil {
		t.Error("duplicate path INSERT should have failed")
	}
}

func TestOpen_MigrationsTracked_Includes0004(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "db.sqlite")

	s, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()

	rows, err := db.Query(`SELECT version FROM schema_migrations ORDER BY version`)
	if err != nil {
		t.Fatalf("query schema_migrations: %v", err)
	}
	defer rows.Close()

	var versions []string
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			t.Fatalf("scan: %v", err)
		}
		versions = append(versions, v)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows.Err: %v", err)
	}

	want := []string{"0001_initial", "0002_add_fts", "0003_normalize_tags", "0004_add_projects"}
	if len(versions) != len(want) {
		t.Fatalf("schema_migrations = %v, want %v", versions, want)
	}
	for i, w := range want {
		if versions[i] != w {
			t.Errorf("versions[%d] = %q, want %q", i, versions[i], w)
		}
	}
}

func TestOpen_0004Idempotent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "db.sqlite")

	s1, err := Open(path)
	if err != nil {
		t.Fatalf("first Open: %v", err)
	}
	if err := s1.Close(); err != nil {
		t.Fatalf("first Close: %v", err)
	}

	s2, err := Open(path)
	if err != nil {
		t.Fatalf("second Open: %v", err)
	}
	t.Cleanup(func() { _ = s2.Close() })

	db := openRawDB(t, path)
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM schema_migrations`).Scan(&count); err != nil {
		t.Fatalf("count schema_migrations: %v", err)
	}
	if count != 4 {
		t.Fatalf("schema_migrations count = %d, want 4", count)
	}
}

func TestMigration0004_DoesNotTouchEntries(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "db.sqlite")

	s, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	_, err = s.Add(Entry{Title: "lossless-check", Project: "platform"})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	// Re-open: 0004 is already applied, so it's a no-op.
	s2, err := Open(path)
	if err != nil {
		t.Fatalf("re-Open: %v", err)
	}
	t.Cleanup(func() { _ = s2.Close() })

	entries, err := s2.List(ListFilter{Project: "platform"})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("len = %d, want 1", len(entries))
	}
	if entries[0].Project != "platform" {
		t.Errorf("entry.Project = %q, want %q", entries[0].Project, "platform")
	}
}
