package storage

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

// newTestStore opens a Store backed by a fresh file under t.TempDir()
// and registers a cleanup that closes it. It returns the store and the
// DB path so tests can open a raw *sql.DB for introspection.
func newTestStore(t *testing.T) (*Store, string) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "db.sqlite")
	s, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s, path
}

// objectExists returns whether the named object of the given kind
// ("table" or "index") exists in sqlite_master.
func objectExists(t *testing.T, db *sql.DB, kind, name string) bool {
	t.Helper()
	var got string
	err := db.QueryRow(
		"SELECT name FROM sqlite_master WHERE type = ? AND name = ?",
		kind, name,
	).Scan(&got)
	if err == sql.ErrNoRows {
		return false
	}
	if err != nil {
		t.Fatalf("objectExists(%s %s): %v", kind, name, err)
	}
	return got == name
}

func TestOpen_CreatesDBFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "db.sqlite")

	s, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
}

func TestOpen_CreatesParentDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "sub", "db.sqlite")

	s, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
}

func TestOpen_SchemaExists(t *testing.T) {
	_, path := newTestStore(t)

	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()

	wantTables := []string{"entries", "schema_migrations"}
	for _, name := range wantTables {
		if !objectExists(t, db, "table", name) {
			t.Errorf("missing table %q", name)
		}
	}

	wantIndexes := []string{"idx_entries_created_at", "idx_entries_project"}
	for _, name := range wantIndexes {
		if !objectExists(t, db, "index", name) {
			t.Errorf("missing index %q", name)
		}
	}
}

func TestOpen_MigrationsTracked(t *testing.T) {
	_, path := newTestStore(t)

	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT version FROM schema_migrations ORDER BY version")
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

	if len(versions) != 1 || versions[0] != "0001_initial" {
		t.Fatalf("schema_migrations = %v, want [0001_initial]", versions)
	}
}

func TestOpen_Idempotent(t *testing.T) {
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

	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count); err != nil {
		t.Fatalf("count schema_migrations: %v", err)
	}
	if count != 1 {
		t.Fatalf("schema_migrations count = %d, want 1", count)
	}

	// Schema still intact.
	for _, tbl := range []string{"entries", "schema_migrations"} {
		if !objectExists(t, db, "table", tbl) {
			t.Errorf("missing table %q after re-open", tbl)
		}
	}
}

func TestAdd_BasicInsert(t *testing.T) {
	s, _ := newTestStore(t)

	got, err := s.Add(Entry{Title: "x"})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	if got.ID <= 0 {
		t.Errorf("ID = %d, want > 0", got.ID)
	}
	if got.CreatedAt.IsZero() {
		t.Error("CreatedAt is zero")
	}
	if !got.UpdatedAt.Equal(got.CreatedAt) {
		t.Errorf("UpdatedAt (%v) != CreatedAt (%v)", got.UpdatedAt, got.CreatedAt)
	}
	if loc := got.CreatedAt.Location().String(); loc != "UTC" {
		t.Errorf("CreatedAt.Location = %q, want UTC", loc)
	}
}

func TestAdd_PersistsAllFields(t *testing.T) {
	s, _ := newTestStore(t)

	in := Entry{
		Title:       "shipped widget v1",
		Description: "did the thing, saved the day",
		Tags:        "auth,perf",
		Project:     "platform",
		Type:        "shipped",
		Impact:      "p99 -30%",
	}
	if _, err := s.Add(in); err != nil {
		t.Fatalf("Add: %v", err)
	}

	list, err := s.List(ListFilter{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("len(List) = %d, want 1", len(list))
	}
	got := list[0]

	if got.Title != in.Title ||
		got.Description != in.Description ||
		got.Tags != in.Tags ||
		got.Project != in.Project ||
		got.Type != in.Type ||
		got.Impact != in.Impact {
		t.Errorf("round-trip mismatch: got %+v, want %+v", got, in)
	}
}

func TestAdd_Duplicates(t *testing.T) {
	s, _ := newTestStore(t)

	a, err := s.Add(Entry{Title: "same"})
	if err != nil {
		t.Fatalf("Add a: %v", err)
	}
	b, err := s.Add(Entry{Title: "same"})
	if err != nil {
		t.Fatalf("Add b: %v", err)
	}
	if a.ID == b.ID {
		t.Errorf("duplicate IDs: both %d", a.ID)
	}

	list, err := s.List(ListFilter{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("len(List) = %d, want 2", len(list))
	}
}

func TestAdd_TimestampsAreRFC3339UTC(t *testing.T) {
	s, path := newTestStore(t)

	if _, err := s.Add(Entry{Title: "ts"}); err != nil {
		t.Fatalf("Add: %v", err)
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()

	var raw string
	if err := db.QueryRow("SELECT created_at FROM entries").Scan(&raw); err != nil {
		t.Fatalf("scan created_at: %v", err)
	}
	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		t.Fatalf("time.Parse(%q): %v", raw, err)
	}
	if loc := parsed.Location().String(); loc != "UTC" {
		t.Errorf("parsed.Location = %q, want UTC", loc)
	}
}

func TestList_EmptyReturnsEmpty(t *testing.T) {
	s, _ := newTestStore(t)

	got, err := s.List(ListFilter{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if got == nil {
		t.Fatal("List returned nil slice, want non-nil")
	}
	if len(got) != 0 {
		t.Fatalf("len(got) = %d, want 0", len(got))
	}
}

func TestList_ReverseChronological(t *testing.T) {
	s, _ := newTestStore(t)

	for _, title := range []string{"a", "b", "c"} {
		if _, err := s.Add(Entry{Title: title}); err != nil {
			t.Fatalf("Add %q: %v", title, err)
		}
		time.Sleep(10 * time.Millisecond)
	}

	got, err := s.List(ListFilter{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	var titles []string
	for _, e := range got {
		titles = append(titles, e.Title)
	}
	want := []string{"c", "b", "a"}
	if len(titles) != len(want) {
		t.Fatalf("titles = %v, want %v", titles, want)
	}
	for i := range want {
		if titles[i] != want[i] {
			t.Fatalf("titles = %v, want %v", titles, want)
		}
	}
}

func TestStore_CloseNoError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "db.sqlite")

	s, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}
