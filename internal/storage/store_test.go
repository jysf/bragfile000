package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
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

func TestOpen_DBFileMode(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "db.sqlite")

	s, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Errorf("DB file mode = %04o, want 0600", got)
	}
}

func TestOpen_DirMode(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "db.sqlite")

	s, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	info, err := os.Stat(filepath.Dir(path))
	if err != nil {
		t.Fatalf("stat dir: %v", err)
	}
	if got := info.Mode().Perm(); got != 0o700 {
		t.Errorf("parent dir mode = %04o, want 0700", got)
	}
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

	want := []string{"0001_initial", "0002_add_fts", "0003_normalize_tags", "0004_add_projects"}
	if len(versions) != len(want) || versions[0] != want[0] || versions[1] != want[1] || versions[2] != want[2] || versions[3] != want[3] {
		t.Fatalf("schema_migrations = %v, want %v", versions, want)
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
	// 0001_initial + 0002_add_fts + 0003_normalize_tags (SPEC-025) + 0004_add_projects (SPEC-027).
	if count != 4 {
		t.Fatalf("schema_migrations count = %d, want 4", count)
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

	// Empty tags via Add (no Tags field = no taggings row).
	addWithTags(t, s, "empty", "", "", "")

	// No-tags row via direct SQL INSERT omitting the tags column entirely
	// (a row with no taggings is the "no tags" case after 0003).
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	now := time.Now().UTC().Format(time.RFC3339)
	if _, err := db.Exec(
		`INSERT INTO entries (title, description, project, type, impact, created_at, updated_at)
		 VALUES (?, NULL, NULL, NULL, NULL, ?, ?)`,
		"nulltags", now, now,
	); err != nil {
		t.Fatalf("INSERT no-tags row: %v", err)
	}

	got, err := s.List(ListFilter{Tag: "auth"})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("len=%d, want 0 (empty and no-tags rows must not match) — titles=%v", len(got), titlesOf(got))
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

// TestList_FilterByAuthor ▲ SPEC-043 — an entry is "agent-authored" iff it
// carries at least one reserved provenance tag (agent:* or model:*, DEC-024);
// "human" is the complement. The classifier is prefix-anchored (agent:%,
// model:%), so a topic tag like "agentic" or "modeling" (no colon) does NOT
// count — mirrors the TagFilterNoFalsePositive guard.
func TestList_FilterByAuthor(t *testing.T) {
	s, _ := newTestStore(t)

	addWithTags(t, s, "human-plain", "perf", "", "")
	addWithTags(t, s, "human-none", "", "", "")
	addWithTags(t, s, "human-fp", "agentic,modeling", "", "") // no colon → not provenance
	addWithTags(t, s, "agent-only", "agent:claude-code", "", "")
	addWithTags(t, s, "agent-both", "perf,agent:claude-code,model:claude-opus-4-8", "", "")

	got, err := s.List(ListFilter{Author: "agent"})
	if err != nil {
		t.Fatalf("Author=agent: %v", err)
	}
	if len(got) != 2 || !containsTitle(got, "agent-only") || !containsTitle(got, "agent-both") {
		t.Errorf("Author=agent: want {agent-only,agent-both}; got %v", titlesOf(got))
	}

	got, err = s.List(ListFilter{Author: "human"})
	if err != nil {
		t.Fatalf("Author=human: %v", err)
	}
	if len(got) != 3 || !containsTitle(got, "human-plain") || !containsTitle(got, "human-none") || !containsTitle(got, "human-fp") {
		t.Errorf("Author=human: want {human-plain,human-none,human-fp}; got %v", titlesOf(got))
	}

	got, err = s.List(ListFilter{})
	if err != nil {
		t.Fatalf("Author unset: %v", err)
	}
	if len(got) != 5 {
		t.Errorf("Author unset: len=%d, want 5 (all) (titles=%v)", len(got), titlesOf(got))
	}

	// Composes with the tag filter (AND): only agent-both carries both perf and provenance.
	got, err = s.List(ListFilter{Author: "agent", Tag: "perf"})
	if err != nil {
		t.Fatalf("Author=agent+Tag=perf: %v", err)
	}
	if len(got) != 1 || got[0].Title != "agent-both" {
		t.Errorf("Author=agent+Tag=perf: want {agent-both}; got %v", titlesOf(got))
	}

	// Composes with limit.
	got, err = s.List(ListFilter{Author: "agent", Limit: 1})
	if err != nil {
		t.Fatalf("Author=agent+Limit=1: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("Author=agent+Limit=1: len=%d, want 1", len(got))
	}

	// Invalid author value is an error, not a silent all-pass.
	if _, err := s.List(ListFilter{Author: "bogus"}); err == nil {
		t.Error("Author=bogus: want error, got nil")
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

// --- SPEC-025: tag normalization ---

// apply0001Only opens a raw DB in a temp file, applies only 0001_initial.sql
// (plus its schema_migrations row), and returns the db and path. The caller
// uses this to seed a corpus into entries.tags before calling Open (which
// applies 0002 + 0003).
func apply0001Only(t *testing.T) (*sql.DB, string) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "db.sqlite")
	ctx := context.Background()

	rawDB, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	if _, err := rawDB.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		version TEXT PRIMARY KEY,
		applied_at TEXT NOT NULL
	)`); err != nil {
		t.Fatalf("create schema_migrations: %v", err)
	}
	body, err := fs.ReadFile(migrationsFS, "migrations/0001_initial.sql")
	if err != nil {
		t.Fatalf("read 0001_initial.sql: %v", err)
	}
	if _, err := rawDB.ExecContext(ctx, string(body)); err != nil {
		t.Fatalf("exec 0001_initial.sql: %v", err)
	}
	if _, err := rawDB.ExecContext(ctx,
		`INSERT INTO schema_migrations (version, applied_at) VALUES (?, ?)`,
		"0001_initial", time.Now().UTC().Format(time.RFC3339),
	); err != nil {
		t.Fatalf("record 0001_initial: %v", err)
	}
	return rawDB, path
}

// TestMigrate_ETL_Lossless seeds a representative corpus via raw INSERT
// into entries.tags (while only 0001 is applied), then calls Open to apply
// 0002+0003 and verifies the ETL is lossless: exact tags/taggings counts,
// correct projected Tags per entry, and no non-entry taggings. Also folds
// in the byte-stable search assertion (SPEC-025 §12(b) confirmed [1 2 3]).
func TestMigrate_ETL_Lossless(t *testing.T) {
	rawDB, path := apply0001Only(t)

	ctx := context.Background()
	now := time.Now().UTC().Format(time.RFC3339)
	corpus := []struct {
		id       int64 // set after insert
		tagsIn   string
		tagsWant string
	}{
		{tagsIn: "auth,perf", tagsWant: "auth,perf"},
		{tagsIn: "perf,auth", tagsWant: "perf,auth"},
		{tagsIn: " auth , auth ,perf", tagsWant: "auth,perf"},
		{tagsIn: "", tagsWant: ""},
		{tagsIn: "", tagsWant: ""}, // NULL via empty string sentinel
		{tagsIn: "solo", tagsWant: "solo"},
	}
	// Row 5 should be NULL in the DB; insert it specially.
	for i, row := range corpus {
		var tagsVal interface{} = row.tagsIn
		if i == 4 { // id=5: NULL
			tagsVal = nil
		}
		res, err := rawDB.ExecContext(ctx,
			`INSERT INTO entries (title, description, tags, project, type, impact, created_at, updated_at)
			 VALUES (?, NULL, ?, NULL, NULL, NULL, ?, ?)`,
			fmt.Sprintf("etl-%d", i+1), tagsVal, now, now,
		)
		if err != nil {
			t.Fatalf("seed row %d: %v", i+1, err)
		}
		id, _ := res.LastInsertId()
		corpus[i].id = id
	}
	rawDB.Close()

	// Open applies 0002 + 0003.
	s, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	db := openRawDB(t, path)

	// (a) 3 distinct tags: auth, perf, solo.
	var tagCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM tags").Scan(&tagCount); err != nil {
		t.Fatalf("count tags: %v", err)
	}
	if tagCount != 3 {
		t.Errorf("tags count = %d, want 3", tagCount)
	}
	rows, err := db.Query("SELECT name FROM tags ORDER BY name")
	if err != nil {
		t.Fatalf("query tags: %v", err)
	}
	var tagNames []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			t.Fatalf("scan tag: %v", err)
		}
		tagNames = append(tagNames, n)
	}
	rows.Close()
	wantNames := []string{"auth", "perf", "solo"}
	if len(tagNames) != len(wantNames) {
		t.Fatalf("tag names = %v, want %v", tagNames, wantNames)
	}
	for i, n := range wantNames {
		if tagNames[i] != n {
			t.Fatalf("tag names = %v, want %v", tagNames, wantNames)
		}
	}

	// (b) 7 taggings total.
	var taggingCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM taggings").Scan(&taggingCount); err != nil {
		t.Fatalf("count taggings: %v", err)
	}
	if taggingCount != 7 {
		t.Errorf("taggings count = %d, want 7", taggingCount)
	}

	// (c) projection round-trips correctly for each entry.
	for _, row := range corpus {
		got, err := s.Get(row.id)
		if err != nil {
			t.Fatalf("Get(%d): %v", row.id, err)
		}
		if got.Tags != row.tagsWant {
			t.Errorf("id=%d: Tags=%q, want %q", row.id, got.Tags, row.tagsWant)
		}
	}

	// (d) only 'entry' taggings exist.
	var nonEntry int
	if err := db.QueryRow(
		"SELECT COUNT(*) FROM taggings WHERE taggable_type <> 'entry'",
	).Scan(&nonEntry); err != nil {
		t.Fatalf("count non-entry taggings: %v", err)
	}
	if nonEntry != 0 {
		t.Errorf("non-entry taggings = %d, want 0", nonEntry)
	}

	// Byte-stable search: 'perf' should match entries 1, 2, 3 (the set {1,2,3}).
	results, err := s.Search("perf", 0)
	if err != nil {
		t.Fatalf("Search(perf): %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("Search(perf) len=%d, want 3", len(results))
	}
	var gotIDs []int64
	for _, e := range results {
		gotIDs = append(gotIDs, e.ID)
	}
	sort.Slice(gotIDs, func(i, j int) bool { return gotIDs[i] < gotIDs[j] })
	wantIDs := []int64{corpus[0].id, corpus[1].id, corpus[2].id}
	for i, id := range wantIDs {
		if gotIDs[i] != id {
			t.Errorf("Search(perf) sorted ids[%d]=%d, want %d (full: %v)", i, gotIDs[i], id, gotIDs)
		}
	}
}

// TestAdd_TagsWriteThroughJoin verifies Add writes tags into the join,
// not a column, and that Get reconstructs them in insertion order.
func TestAdd_TagsWriteThroughJoin(t *testing.T) {
	s, path := newTestStore(t)

	e, err := s.Add(Entry{Title: "join-write", Tags: "perf,auth"})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}

	// (a) Get returns Tags == "perf,auth" (insertion order preserved).
	got, err := s.Get(e.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Tags != "perf,auth" {
		t.Errorf("Tags = %q, want %q", got.Tags, "perf,auth")
	}

	// (b) two taggings exist for this entry, positions 0,1 → perf,auth.
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()

	rows, err := db.Query(
		`SELECT t.name, tg.position FROM taggings tg
		 JOIN tags t ON t.id = tg.tag_id
		 WHERE tg.taggable_type = 'entry' AND tg.taggable_id = ?
		 ORDER BY tg.position`,
		e.ID,
	)
	if err != nil {
		t.Fatalf("query taggings: %v", err)
	}
	defer rows.Close()

	type pair struct {
		name     string
		position int
	}
	var pairs []pair
	for rows.Next() {
		var p pair
		if err := rows.Scan(&p.name, &p.position); err != nil {
			t.Fatalf("scan: %v", err)
		}
		pairs = append(pairs, p)
	}
	if len(pairs) != 2 {
		t.Fatalf("taggings count = %d, want 2", len(pairs))
	}
	if pairs[0].name != "perf" || pairs[0].position != 0 {
		t.Errorf("tagging[0] = {%q, %d}, want {perf, 0}", pairs[0].name, pairs[0].position)
	}
	if pairs[1].name != "auth" || pairs[1].position != 1 {
		t.Errorf("tagging[1] = {%q, %d}, want {auth, 1}", pairs[1].name, pairs[1].position)
	}

	// (c) tags table has rows for perf and auth.
	for _, name := range []string{"perf", "auth"} {
		var id int64
		if err := db.QueryRow("SELECT id FROM tags WHERE name = ?", name).Scan(&id); err != nil {
			t.Errorf("tag %q missing: %v", name, err)
		}
	}
}

// TestAdd_TagsCanonicalizeTrimDedup verifies non-canonical input is
// trimmed and deduplicated on write.
func TestAdd_TagsCanonicalizeTrimDedup(t *testing.T) {
	s, path := newTestStore(t)

	e, err := s.Add(Entry{Title: "canon", Tags: " auth , auth ,perf"})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	got, err := s.Get(e.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Tags != "auth,perf" {
		t.Errorf("Tags = %q, want %q", got.Tags, "auth,perf")
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	var count int
	if err := db.QueryRow(
		"SELECT COUNT(*) FROM taggings WHERE taggable_type='entry' AND taggable_id=?", e.ID,
	).Scan(&count); err != nil {
		t.Fatalf("count taggings: %v", err)
	}
	if count != 2 {
		t.Errorf("taggings count = %d, want 2 (auth,perf after dedup)", count)
	}
}

// TestAdd_EmptyTagsNoTaggings verifies an entry with no tags produces
// zero taggings and returns Tags == "".
func TestAdd_EmptyTagsNoTaggings(t *testing.T) {
	s, path := newTestStore(t)

	e, err := s.Add(Entry{Title: "notags"})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	got, err := s.Get(e.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Tags != "" {
		t.Errorf("Tags = %q, want empty", got.Tags)
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	var count int
	if err := db.QueryRow(
		"SELECT COUNT(*) FROM taggings WHERE taggable_type='entry' AND taggable_id=?", e.ID,
	).Scan(&count); err != nil {
		t.Fatalf("count taggings: %v", err)
	}
	if count != 0 {
		t.Errorf("taggings count = %d, want 0", count)
	}
}

// TestUpdate_TagsReplacedThroughJoin verifies Update replaces taggings
// atomically: old membership gone, new membership written, orphan tag
// row is not referenced by any tagging.
func TestUpdate_TagsReplacedThroughJoin(t *testing.T) {
	s, path := newTestStore(t)

	inserted, err := s.Add(Entry{Title: "u", Tags: "a,b"})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	if _, err := s.Update(inserted.ID, Entry{Title: "u", Tags: "b,c"}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, err := s.Get(inserted.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Tags != "b,c" {
		t.Errorf("Tags = %q, want %q", got.Tags, "b,c")
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()

	// Taggings for entry = exactly {b, c}.
	rows, err := db.Query(
		`SELECT t.name FROM taggings tg JOIN tags t ON t.id = tg.tag_id
		 WHERE tg.taggable_type = 'entry' AND tg.taggable_id = ?
		 ORDER BY t.name`,
		inserted.ID,
	)
	if err != nil {
		t.Fatalf("query taggings: %v", err)
	}
	defer rows.Close()
	var names []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			t.Fatalf("scan: %v", err)
		}
		names = append(names, n)
	}
	if len(names) != 2 || names[0] != "b" || names[1] != "c" {
		t.Errorf("taggings names = %v, want [b c]", names)
	}

	// Orphaned tag "a" must not be referenced by any tagging.
	var aRefs int
	if err := db.QueryRow(
		`SELECT COUNT(*) FROM taggings tg JOIN tags t ON t.id = tg.tag_id WHERE t.name = 'a'`,
	).Scan(&aRefs); err != nil {
		t.Fatalf("count a refs: %v", err)
	}
	if aRefs != 0 {
		t.Errorf("orphan tag 'a' still referenced by %d tagging(s)", aRefs)
	}
}

// TestDelete_RemovesTaggings verifies Delete removes taggings for the
// entry and the entries_fts row.
func TestDelete_RemovesTaggings(t *testing.T) {
	s, path := newTestStore(t)

	e, err := s.Add(Entry{Title: "del", Tags: "x,y"})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	if err := s.Delete(e.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()

	var taggingCount int
	if err := db.QueryRow(
		"SELECT COUNT(*) FROM taggings WHERE taggable_type='entry' AND taggable_id=?", e.ID,
	).Scan(&taggingCount); err != nil {
		t.Fatalf("count taggings: %v", err)
	}
	if taggingCount != 0 {
		t.Errorf("taggings count = %d, want 0 after Delete", taggingCount)
	}

	var ftsCount int
	if err := db.QueryRow(
		"SELECT COUNT(*) FROM entries_fts WHERE rowid = ?", e.ID,
	).Scan(&ftsCount); err != nil {
		t.Fatalf("count entries_fts: %v", err)
	}
	if ftsCount != 0 {
		t.Errorf("entries_fts count = %d, want 0 after Delete", ftsCount)
	}
}

// TestList_TagFilterThroughJoin is an explicit join-path test for the
// tag filter — mirrors TestList_FilterByTag but names the join mechanism.
func TestList_TagFilterThroughJoin(t *testing.T) {
	s, _ := newTestStore(t)

	addWithTags(t, s, "ap", "auth,perf", "", "")
	addWithTags(t, s, "pb", "perf,backend", "", "")
	addWithTags(t, s, "a", "auth", "", "")

	cases := []struct {
		tag     string
		wantLen int
	}{
		{"auth", 2},
		{"perf", 2},
		{"backend", 1},
		{"nonesuch", 0},
	}
	for _, tc := range cases {
		got, err := s.List(ListFilter{Tag: tc.tag})
		if err != nil {
			t.Fatalf("List(Tag=%q): %v", tc.tag, err)
		}
		if len(got) != tc.wantLen {
			t.Errorf("List(Tag=%q) len=%d, want %d (titles=%v)",
				tc.tag, len(got), tc.wantLen, titlesOf(got))
		}
	}
}
