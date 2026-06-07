package storage

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"testing/fstest"

	_ "modernc.org/sqlite"
)

// openScratchDB opens a fresh sqlite DB in a temp file for migration-
// runner tests (we don't use Open here, since Open embeds its own FS).
func openScratchDB(t *testing.T) *sql.DB {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "db.sqlite")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestMigrate_AppliesInOrder(t *testing.T) {
	src := fstest.MapFS{
		"0001_a.sql": &fstest.MapFile{Data: []byte(`CREATE TABLE a(x INTEGER);`)},
		"0002_b.sql": &fstest.MapFile{Data: []byte(`CREATE TABLE b(x INTEGER);`)},
	}
	db := openScratchDB(t)

	if err := applyMigrations(context.Background(), db, src); err != nil {
		t.Fatalf("applyMigrations: %v", err)
	}

	rows, err := db.Query("SELECT version FROM schema_migrations ORDER BY applied_at, version")
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

	wantVersions := []string{"0001_a", "0002_b"}
	if len(versions) != len(wantVersions) {
		t.Fatalf("versions = %v, want %v", versions, wantVersions)
	}
	for i := range wantVersions {
		if versions[i] != wantVersions[i] {
			t.Fatalf("versions = %v, want %v", versions, wantVersions)
		}
	}

	for _, tbl := range []string{"a", "b"} {
		var got string
		err := db.QueryRow(
			"SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?",
			tbl,
		).Scan(&got)
		if err != nil {
			t.Errorf("table %q missing: %v", tbl, err)
		}
	}
}

func TestMigrate_FailedMigrationRollsBack(t *testing.T) {
	src := fstest.MapFS{
		"0001_good.sql": &fstest.MapFile{Data: []byte(`CREATE TABLE good(x INTEGER);`)},
		"0002_bad.sql":  &fstest.MapFile{Data: []byte(`CREATE TABLE if not a valid syntax;;`)},
	}
	db := openScratchDB(t)

	if err := applyMigrations(context.Background(), db, src); err == nil {
		t.Fatal("applyMigrations: want non-nil error, got nil")
	}

	rows, err := db.Query("SELECT version FROM schema_migrations")
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
	if len(versions) != 1 || versions[0] != "0001_good" {
		t.Fatalf("schema_migrations = %v, want [0001_good]", versions)
	}

	// good table exists
	var got string
	if err := db.QueryRow(
		"SELECT name FROM sqlite_master WHERE type = 'table' AND name = 'good'",
	).Scan(&got); err != nil {
		t.Fatalf("good table missing: %v", err)
	}

	// bad table does not exist
	err = db.QueryRow(
		"SELECT name FROM sqlite_master WHERE type = 'table' AND name = 'bad'",
	).Scan(&got)
	if err != sql.ErrNoRows {
		t.Errorf("expected no bad table; got name=%q err=%v", got, err)
	}
}

func TestMigrate_Idempotent(t *testing.T) {
	src := fstest.MapFS{
		"0001_a.sql": &fstest.MapFile{Data: []byte(`CREATE TABLE a(x INTEGER);`)},
		"0002_b.sql": &fstest.MapFile{Data: []byte(`CREATE TABLE b(x INTEGER);`)},
	}
	db := openScratchDB(t)

	if err := applyMigrations(context.Background(), db, src); err != nil {
		t.Fatalf("first applyMigrations: %v", err)
	}
	if err := applyMigrations(context.Background(), db, src); err != nil {
		t.Fatalf("second applyMigrations: %v", err)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count); err != nil {
		t.Fatalf("count schema_migrations: %v", err)
	}
	if count != 2 {
		t.Fatalf("schema_migrations count = %d, want 2", count)
	}
}

// TestOpen_TagSchemaExists asserts the normalized tag tables and their
// indexes exist after Open applies all migrations (SPEC-025).
func TestOpen_TagSchemaExists(t *testing.T) {
	_, path := newTestStore(t)
	db := openRawDB(t, path)

	for _, tbl := range []string{"tags", "taggings"} {
		var got string
		err := db.QueryRow(
			"SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?", tbl,
		).Scan(&got)
		if err != nil {
			t.Errorf("table %q missing: %v", tbl, err)
		}
	}

	for _, idx := range []string{"idx_taggings_tag", "idx_taggings_taggable"} {
		var got string
		err := db.QueryRow(
			"SELECT name FROM sqlite_master WHERE type = 'index' AND name = ?", idx,
		).Scan(&got)
		if err != nil {
			t.Errorf("index %q missing: %v", idx, err)
		}
	}
}

// TestOpen_TagsColumnDropped asserts entries.tags no longer exists after
// the 0003 migration (SPEC-025).
func TestOpen_TagsColumnDropped(t *testing.T) {
	_, path := newTestStore(t)
	db := openRawDB(t, path)

	rows, err := db.Query("PRAGMA table_info(entries)")
	if err != nil {
		t.Fatalf("PRAGMA table_info: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name, typ string
		var notnull int
		var dflt interface{}
		var pk int
		if err := rows.Scan(&cid, &name, &typ, &notnull, &dflt, &pk); err != nil {
			t.Fatalf("scan PRAGMA row: %v", err)
		}
		if name == "tags" {
			t.Errorf("entries.tags column still present after 0003 migration")
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows.Err: %v", err)
	}
}
