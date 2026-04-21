package storage

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jysf/bragfile000/internal/storage/storagetest"

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

// addWithTags is a convenience for filter tests: calls Add with only
// the fields the filter tests exercise populated.
func addWithTags(t *testing.T, s *Store, title, tags, project, typ string) Entry {
	t.Helper()
	e, err := s.Add(Entry{Title: title, Tags: tags, Project: project, Type: typ})
	if err != nil {
		t.Fatalf("Add(%q): %v", title, err)
	}
	return e
}

// mustBackdate forwards to storagetest.Backdate and t.Fatals on error.
// Same name and shape as the helper in cli's list_test.go so future
// readers don't have to mentally translate between packages.
func mustBackdate(t *testing.T, path string, id int64, at time.Time) {
	t.Helper()
	if err := storagetest.Backdate(path, id, at); err != nil {
		t.Fatalf("storagetest.Backdate: %v", err)
	}
}

func titlesOf(entries []Entry) []string {
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		out = append(out, e.Title)
	}
	return out
}

func containsTitle(entries []Entry, title string) bool {
	for _, e := range entries {
		if e.Title == title {
			return true
		}
	}
	return false
}

func TestList_FilterByTag(t *testing.T) {
	s, _ := newTestStore(t)

	addWithTags(t, s, "ap", "auth,perf", "", "")
	addWithTags(t, s, "pb", "perf,backend", "", "")
	addWithTags(t, s, "a", "auth", "", "")

	got, err := s.List(ListFilter{Tag: "auth"})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("Tag=auth: len=%d, want 2 (titles=%v)", len(got), titlesOf(got))
	}

	got, err = s.List(ListFilter{Tag: "perf"})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("Tag=perf: len=%d, want 2 (titles=%v)", len(got), titlesOf(got))
	}

	got, err = s.List(ListFilter{Tag: "backend"})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("Tag=backend: len=%d, want 1 (titles=%v)", len(got), titlesOf(got))
	}

	got, err = s.List(ListFilter{Tag: "nonesuch"})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if got == nil {
		t.Error("Tag=nonesuch: got nil slice, want non-nil empty")
	}
	if len(got) != 0 {
		t.Errorf("Tag=nonesuch: len=%d, want 0", len(got))
	}
}

func TestList_TagFilterNoFalsePositive(t *testing.T) {
	s, _ := newTestStore(t)

	addWithTags(t, s, "exact", "auth", "", "")
	addWithTags(t, s, "superstring", "authoring", "", "")

	got, err := s.List(ListFilter{Tag: "auth"})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len=%d, want 1 (titles=%v)", len(got), titlesOf(got))
	}
	if got[0].Title != "exact" {
		t.Errorf("matched %q, want %q", got[0].Title, "exact")
	}
}

func TestList_TagFilterNullAndEmpty(t *testing.T) {
	s, path := newTestStore(t)

	// Empty tags via Add (stored as "").
	addWithTags(t, s, "empty", "", "", "")

	// NULL tags via direct SQL INSERT (Store.Add always writes "").
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	now := time.Now().UTC().Format(time.RFC3339)
	if _, err := db.Exec(
		`INSERT INTO entries (title, description, tags, project, type, impact, created_at, updated_at)
		 VALUES (?, NULL, NULL, NULL, NULL, NULL, ?, ?)`,
		"nulltags", now, now,
	); err != nil {
		t.Fatalf("INSERT NULL tags: %v", err)
	}

	got, err := s.List(ListFilter{Tag: "auth"})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("len=%d, want 0 (null and empty tags must not match) — titles=%v", len(got), titlesOf(got))
	}
}

func TestList_FilterByProject(t *testing.T) {
	s, _ := newTestStore(t)

	addWithTags(t, s, "p1", "", "platform", "")
	addWithTags(t, s, "g", "", "growth", "")
	addWithTags(t, s, "p2", "", "platform", "")

	got, err := s.List(ListFilter{Project: "platform"})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("Project=platform: len=%d, want 2 (titles=%v)", len(got), titlesOf(got))
	}

	got, err = s.List(ListFilter{Project: "Platform"})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("Project=Platform (case-sensitive): len=%d, want 0", len(got))
	}
}

func TestList_FilterByType(t *testing.T) {
	s, _ := newTestStore(t)

	addWithTags(t, s, "a", "", "", "shipped")
	addWithTags(t, s, "b", "", "", "learned")
	addWithTags(t, s, "c", "", "", "shipped")

	got, err := s.List(ListFilter{Type: "shipped"})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("Type=shipped: len=%d, want 2 (titles=%v)", len(got), titlesOf(got))
	}

	got, err = s.List(ListFilter{Type: "Shipped"})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("Type=Shipped (case-sensitive): len=%d, want 0", len(got))
	}
}

func TestList_FilterBySince(t *testing.T) {
	s, path := newTestStore(t)

	now := time.Now().UTC()
	a := addWithTags(t, s, "old", "", "", "")
	b := addWithTags(t, s, "recent", "", "", "")
	c := addWithTags(t, s, "newest", "", "", "")

	mustBackdate(t, path, a.ID, now.Add(-3*24*time.Hour))
	mustBackdate(t, path, b.ID, now.Add(-1*24*time.Hour))
	mustBackdate(t, path, c.ID, now)

	twoDaysAgo := now.Add(-2 * 24 * time.Hour)
	got, err := s.List(ListFilter{Since: twoDaysAgo})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len=%d, want 2 (titles=%v)", len(got), titlesOf(got))
	}
	if !containsTitle(got, "recent") || !containsTitle(got, "newest") {
		t.Errorf("want {recent,newest}; got %v", titlesOf(got))
	}
	if containsTitle(got, "old") {
		t.Errorf("should not include backdated %q; got %v", "old", titlesOf(got))
	}
}

func TestList_FilterByLimit(t *testing.T) {
	s, _ := newTestStore(t)

	for _, title := range []string{"a", "b", "c", "d", "e"} {
		addWithTags(t, s, title, "", "", "")
		time.Sleep(1 * time.Millisecond)
	}

	got, err := s.List(ListFilter{Limit: 2})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len=%d, want 2 (titles=%v)", len(got), titlesOf(got))
	}
	// ORDER BY id DESC — newest IDs first; last-inserted "e" has highest id.
	if got[0].Title != "e" || got[1].Title != "d" {
		t.Errorf("titles=%v, want [e d]", titlesOf(got))
	}
}

func TestList_FilterCombined(t *testing.T) {
	s, path := newTestStore(t)

	now := time.Now().UTC()
	// Hit: project=platform, tag=auth, recent.
	hit := addWithTags(t, s, "hit", "auth,perf", "platform", "shipped")
	// Miss: wrong project.
	missP := addWithTags(t, s, "missP", "auth", "growth", "shipped")
	// Miss: wrong tag.
	missT := addWithTags(t, s, "missT", "perf", "platform", "shipped")
	// Miss: too old.
	old := addWithTags(t, s, "old", "auth", "platform", "shipped")

	mustBackdate(t, path, hit.ID, now)
	mustBackdate(t, path, missP.ID, now)
	mustBackdate(t, path, missT.ID, now)
	mustBackdate(t, path, old.ID, now.Add(-10*24*time.Hour))

	got, err := s.List(ListFilter{
		Tag:     "auth",
		Project: "platform",
		Since:   now.Add(-7 * 24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len=%d, want 1 (titles=%v)", len(got), titlesOf(got))
	}
	if got[0].Title != "hit" {
		t.Errorf("got %q, want %q", got[0].Title, "hit")
	}
}

func TestList_FilterPreservesOrder(t *testing.T) {
	s, _ := newTestStore(t)

	// Insert three same-tag rows in rapid succession so created_at may
	// collide at second-precision; id DESC tie-break must carry.
	a := addWithTags(t, s, "first", "x", "", "")
	b := addWithTags(t, s, "second", "x", "", "")
	c := addWithTags(t, s, "third", "x", "", "")

	got, err := s.List(ListFilter{Tag: "x"})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("len=%d, want 3 (titles=%v)", len(got), titlesOf(got))
	}
	// Expect id DESC: c, b, a.
	wantIDs := []int64{c.ID, b.ID, a.ID}
	for i, e := range got {
		if e.ID != wantIDs[i] {
			t.Errorf("pos %d: id=%d, want %d (titles=%v)", i, e.ID, wantIDs[i], titlesOf(got))
		}
	}
}

func TestDelete_RemovesRow(t *testing.T) {
	s, _ := newTestStore(t)

	inserted, err := s.Add(Entry{Title: "to be deleted"})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	if err := s.Delete(inserted.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := s.Get(inserted.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get after Delete: err = %v, want ErrNotFound", err)
	}
}

func TestDelete_NotFoundReturnsErrNotFound(t *testing.T) {
	s, _ := newTestStore(t)

	err := s.Delete(999)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("Delete(999): err = %v, want ErrNotFound", err)
	}
}

func TestUpdate_ReplacesUserEditableFields(t *testing.T) {
	s, _ := newTestStore(t)

	inserted, err := s.Add(Entry{Title: "orig", Description: "orig body", Tags: "a"})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}

	if _, err := s.Update(inserted.ID, Entry{
		Title:       "new",
		Description: "new body",
		Tags:        "x,y",
		Project:     "p",
		Type:        "t",
		Impact:      "i",
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, err := s.Get(inserted.ID)
	if err != nil {
		t.Fatalf("Get after Update: %v", err)
	}
	if got.Title != "new" || got.Description != "new body" ||
		got.Tags != "x,y" || got.Project != "p" ||
		got.Type != "t" || got.Impact != "i" {
		t.Errorf("Update did not replace fields; got %+v", got)
	}
}

func TestUpdate_PreservesID(t *testing.T) {
	s, _ := newTestStore(t)

	inserted, err := s.Add(Entry{Title: "orig"})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	returned, err := s.Update(inserted.ID, Entry{Title: "new"})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if returned.ID != inserted.ID {
		t.Errorf("ID = %d, want %d", returned.ID, inserted.ID)
	}
}

func TestUpdate_PreservesCreatedAt(t *testing.T) {
	s, _ := newTestStore(t)

	inserted, err := s.Add(Entry{Title: "orig"})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	origCreated := inserted.CreatedAt
	time.Sleep(1100 * time.Millisecond)

	returned, err := s.Update(inserted.ID, Entry{Title: "new"})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if !returned.CreatedAt.Equal(origCreated) {
		t.Errorf("CreatedAt = %v, want %v (must be preserved)", returned.CreatedAt, origCreated)
	}
}

func TestUpdate_BumpsUpdatedAt(t *testing.T) {
	s, _ := newTestStore(t)

	inserted, err := s.Add(Entry{Title: "orig"})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	origUpdated := inserted.UpdatedAt
	time.Sleep(1100 * time.Millisecond)

	returned, err := s.Update(inserted.ID, Entry{Title: "new"})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if !returned.UpdatedAt.After(origUpdated) {
		t.Errorf("UpdatedAt = %v, want strictly after %v", returned.UpdatedAt, origUpdated)
	}
	if loc := returned.UpdatedAt.Location().String(); loc != "UTC" {
		t.Errorf("UpdatedAt.Location = %q, want UTC", loc)
	}
}

func TestUpdate_NotFoundReturnsErrNotFound(t *testing.T) {
	s, _ := newTestStore(t)

	_, err := s.Update(999, Entry{Title: "x"})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("Update(999): err = %v, want errors.Is(err, ErrNotFound)", err)
	}
}

func TestUpdate_ReturnsHydratedEntry(t *testing.T) {
	s, _ := newTestStore(t)

	inserted, err := s.Add(Entry{Title: "orig"})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	time.Sleep(1100 * time.Millisecond)

	returned, err := s.Update(inserted.ID, Entry{Title: "new", Tags: "x"})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if returned.ID == 0 {
		t.Errorf("returned.ID is zero; expected hydrated Entry")
	}
	if returned.Title != "new" || returned.Tags != "x" {
		t.Errorf("returned = %+v; expected input values reflected", returned)
	}
	if !returned.UpdatedAt.After(inserted.UpdatedAt) {
		t.Errorf("returned.UpdatedAt = %v, want strictly after %v", returned.UpdatedAt, inserted.UpdatedAt)
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
