package storage

import (
	"context"
	"database/sql"
	"path/filepath"
	"strings"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func TestBackup_CreatesSnapshotForExistingDBWithPending(t *testing.T) {
	rawDB, path := apply0001Only(t)

	ctx := context.Background()
	now := time.Now().UTC().Format(time.RFC3339)
	if _, err := rawDB.ExecContext(ctx,
		`INSERT INTO entries (title, created_at, updated_at) VALUES (?, ?, ?)`,
		"seed entry", now, now,
	); err != nil {
		t.Fatalf("seed entry: %v", err)
	}
	if err := rawDB.Close(); err != nil {
		t.Fatalf("rawDB.Close: %v", err)
	}

	frozen, _ := time.Parse(time.RFC3339, "2026-06-12T09:30:15Z")
	orig := clock
	clock = func() time.Time { return frozen }
	defer func() { clock = orig }()

	s, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer s.Close()

	matches, err := filepath.Glob(path + ".pre-*.backup")
	if err != nil {
		t.Fatalf("Glob: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("want 1 backup file, got %d: %v", len(matches), matches)
	}
	want := path + ".pre-0004_add_projects.20260612T093015Z.backup"
	if matches[0] != want {
		t.Errorf("backup name: got %q, want %q", matches[0], want)
	}

	sidecarDB := openRawDB(t, matches[0])

	var sidecarCount int
	if err := sidecarDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM entries`).Scan(&sidecarCount); err != nil {
		t.Fatalf("count entries in sidecar: %v", err)
	}
	if sidecarCount != 1 {
		t.Errorf("sidecar entries count: got %d, want 1", sidecarCount)
	}

	if objectExists(t, sidecarDB, "table", "projects") {
		t.Error("sidecar should NOT have projects table (pre-migration snapshot)")
	}

	liveDB := openRawDB(t, path)

	var schemaCount int
	if err := liveDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM schema_migrations`).Scan(&schemaCount); err != nil {
		t.Fatalf("count schema_migrations in live DB: %v", err)
	}
	if schemaCount != 4 {
		t.Errorf("live schema_migrations count: got %d, want 4", schemaCount)
	}

	if !objectExists(t, liveDB, "table", "projects") {
		t.Error("live DB should have projects table after migration")
	}
}

func TestBackup_NoSnapshotForFreshDB(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "db.sqlite")

	s, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer s.Close()

	matches, err := filepath.Glob(path + ".pre-*.backup")
	if err != nil {
		t.Fatalf("Glob: %v", err)
	}
	if len(matches) != 0 {
		t.Errorf("want 0 backup files for fresh DB, got %d: %v", len(matches), matches)
	}
}

func TestBackup_NoSnapshotWhenAlreadyAtHead(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "db.sqlite")

	// First open: applied==0 at start → no backup, applies all migrations.
	s, err := Open(path)
	if err != nil {
		t.Fatalf("Open (first): %v", err)
	}
	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	matches, err := filepath.Glob(path + ".pre-*.backup")
	if err != nil {
		t.Fatalf("Glob after first open: %v", err)
	}
	if len(matches) != 0 {
		t.Fatalf("want 0 backup files after first open, got %d: %v", len(matches), matches)
	}

	// Second open: applied==4, pending==0 → no backup.
	s2, err := Open(path)
	if err != nil {
		t.Fatalf("Open (second): %v", err)
	}
	defer s2.Close()

	matches, err = filepath.Glob(path + ".pre-*.backup")
	if err != nil {
		t.Fatalf("Glob after second open: %v", err)
	}
	if len(matches) != 0 {
		t.Errorf("want 0 backup files for at-head re-open, got %d: %v", len(matches), matches)
	}
}

func TestBackup_FailureAbortsOpenAndLeavesDBUnmigrated(t *testing.T) {
	rawDB, path := apply0001Only(t)

	ctx := context.Background()
	now := time.Now().UTC().Format(time.RFC3339)
	if _, err := rawDB.ExecContext(ctx,
		`INSERT INTO entries (title, created_at, updated_at) VALUES (?, ?, ?)`,
		"seed entry", now, now,
	); err != nil {
		t.Fatalf("seed entry: %v", err)
	}
	if err := rawDB.Close(); err != nil {
		t.Fatalf("rawDB.Close: %v", err)
	}

	frozen, _ := time.Parse(time.RFC3339, "2026-06-12T10:00:00Z")
	orig := clock
	clock = func() time.Time { return frozen }
	defer func() { clock = orig }()

	// Pre-create the exact target path as a real SQLite database so VACUUM INTO
	// fails with "output file already exists". An empty 0-byte file is
	// insufficient — modernc.org/sqlite v1.52.0 overwrites empty files silently;
	// a non-empty SQLite file triggers the expected error (noted in Build
	// Completion deviations).
	collisionPath := path + ".pre-0004_add_projects." + frozen.Format(backupTimeFormat) + ".backup"
	collisionDB, err := sql.Open("sqlite", collisionPath)
	if err != nil {
		t.Fatalf("create collision db: %v", err)
	}
	if _, err := collisionDB.ExecContext(ctx, `CREATE TABLE _placeholder (x INTEGER)`); err != nil {
		_ = collisionDB.Close()
		t.Fatalf("seed collision db: %v", err)
	}
	if err := collisionDB.Close(); err != nil {
		t.Fatalf("close collision db: %v", err)
	}

	s, err := Open(path)
	if s != nil {
		_ = s.Close()
		t.Error("want nil store on backup failure, got non-nil")
	}
	if err == nil {
		t.Fatal("want error on backup failure, got nil")
	}
	if !strings.Contains(err.Error(), "open store:") {
		t.Errorf("error should contain %q: %v", "open store:", err)
	}
	if !strings.Contains(err.Error(), "backup before migrations") {
		t.Errorf("error should contain %q: %v", "backup before migrations", err)
	}

	// DB must be un-migrated: only 0001 applied — abort happened before migrating.
	liveDB := openRawDB(t, path)
	var count int
	if err := liveDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM schema_migrations`).Scan(&count); err != nil {
		t.Fatalf("count schema_migrations: %v", err)
	}
	if count != 1 {
		t.Errorf("schema_migrations count: got %d, want 1 (abort should leave DB un-migrated)", count)
	}
}

func TestBackup_NoNewMigrationVersion(t *testing.T) {
	rawDB, path := apply0001Only(t)

	ctx := context.Background()
	now := time.Now().UTC().Format(time.RFC3339)
	if _, err := rawDB.ExecContext(ctx,
		`INSERT INTO entries (title, created_at, updated_at) VALUES (?, ?, ?)`,
		"seed entry", now, now,
	); err != nil {
		t.Fatalf("seed entry: %v", err)
	}
	if err := rawDB.Close(); err != nil {
		t.Fatalf("rawDB.Close: %v", err)
	}

	s, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer s.Close()

	liveDB := openRawDB(t, path)
	rows, err := liveDB.QueryContext(ctx, `SELECT version FROM schema_migrations ORDER BY version`)
	if err != nil {
		t.Fatalf("query schema_migrations: %v", err)
	}
	defer rows.Close()

	var versions []string
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			t.Fatalf("scan version: %v", err)
		}
		versions = append(versions, v)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate schema_migrations: %v", err)
	}

	want := []string{"0001_initial", "0002_add_fts", "0003_normalize_tags", "0004_add_projects"}
	if len(versions) != len(want) {
		t.Fatalf("schema_migrations count: got %d, want %d (%v)", len(versions), len(want), versions)
	}
	for i, v := range versions {
		if v != want[i] {
			t.Errorf("schema_migrations[%d]: got %q, want %q", i, v, want[i])
		}
	}
}
