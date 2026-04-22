package storage

import (
	"context"
	"database/sql"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

// openRawDB opens a raw *sql.DB at path for direct introspection of
// sqlite_master, entries_fts, and schema_migrations. Registers a
// t.Cleanup that closes the connection.
func openRawDB(t *testing.T, path string) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("sql.Open(%q): %v", path, err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// TestFTS_SmokeCreateUnderPureGoDriver proves FTS5 is compiled into
// modernc.org/sqlite. If this ever fails, SPEC-011 is blocked until
// the driver situation is re-evaluated (see decision #6).
func TestFTS_SmokeCreateUnderPureGoDriver(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "smoke.sqlite")

	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if _, err := db.ExecContext(context.Background(),
		`CREATE VIRTUAL TABLE fts5_smoke USING fts5(content)`,
	); err != nil {
		t.Fatalf("CREATE VIRTUAL TABLE ... USING fts5: %v", err)
	}
}

// TestFTS_VirtualTableShape asserts the entries_fts virtual table
// exists with the expected indexed columns, external-content linkage
// to entries, and default (unicode61) tokenizer.
func TestFTS_VirtualTableShape(t *testing.T) {
	_, path := newTestStore(t)
	db := openRawDB(t, path)

	var name, ddl string
	err := db.QueryRow(
		`SELECT name, sql FROM sqlite_master
		 WHERE type = 'table' AND name = 'entries_fts'`,
	).Scan(&name, &ddl)
	if err == sql.ErrNoRows {
		t.Fatal("entries_fts not found in sqlite_master")
	}
	if err != nil {
		t.Fatalf("query sqlite_master: %v", err)
	}

	for _, col := range []string{"title", "description", "tags", "project", "impact"} {
		if !strings.Contains(ddl, col) {
			t.Errorf("entries_fts DDL missing column %q; ddl=%q", col, ddl)
		}
	}
	if !strings.Contains(ddl, "content='entries'") {
		t.Errorf("entries_fts DDL missing content='entries'; ddl=%q", ddl)
	}
	if !strings.Contains(ddl, "content_rowid='id'") {
		t.Errorf("entries_fts DDL missing content_rowid='id'; ddl=%q", ddl)
	}
	if strings.Contains(ddl, "tokenize") {
		t.Errorf("entries_fts DDL should use the default tokenizer; ddl=%q", ddl)
	}
}

// TestFTS_TriggersExistAfterMigration asserts the three FTS-sync
// triggers exist (and no others).
func TestFTS_TriggersExistAfterMigration(t *testing.T) {
	_, path := newTestStore(t)
	db := openRawDB(t, path)

	rows, err := db.Query(`SELECT name FROM sqlite_master WHERE type = 'trigger' ORDER BY name`)
	if err != nil {
		t.Fatalf("query triggers: %v", err)
	}
	defer rows.Close()

	var got []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			t.Fatalf("scan: %v", err)
		}
		got = append(got, n)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows.Err: %v", err)
	}

	want := []string{"entries_ad", "entries_ai", "entries_au"}
	sort.Strings(got)
	if len(got) != len(want) {
		t.Fatalf("triggers = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("triggers = %v, want %v", got, want)
		}
	}
}

// TestFTS_BothMigrationsTracked asserts both 0001 and 0002 are
// recorded in schema_migrations after Open.
func TestFTS_BothMigrationsTracked(t *testing.T) {
	_, path := newTestStore(t)
	db := openRawDB(t, path)

	rows, err := db.Query(`SELECT version FROM schema_migrations ORDER BY version`)
	if err != nil {
		t.Fatalf("query schema_migrations: %v", err)
	}
	defer rows.Close()

	var got []string
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			t.Fatalf("scan: %v", err)
		}
		got = append(got, v)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows.Err: %v", err)
	}

	want := []string{"0001_initial", "0002_add_fts"}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("schema_migrations = %v, want %v", got, want)
	}
}

// TestFTS_MigrationBackfillsExistingRows manually builds a DB with
// only 0001 applied plus three pre-existing rows, then runs Open
// (which applies 0002) and asserts the backfill populated
// entries_fts with those rows.
//
// The 0001 migration SQL is read from the package-private
// migrationsFS to avoid duplication drift (the alternative allowed by
// the spec is to inline-duplicate 0001's CREATE TABLE statements; we
// chose embed-read since the test is in-package).
func TestFTS_MigrationBackfillsExistingRows(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "db.sqlite")

	ctx := context.Background()

	// Step 1: open a raw DB and apply ONLY 0001_initial.sql plus a
	// matching schema_migrations row, so the runner on next Open sees
	// 0001 as already applied and only applies 0002.
	func() {
		rawDB, err := sql.Open("sqlite", path)
		if err != nil {
			t.Fatalf("sql.Open: %v", err)
		}
		defer rawDB.Close()

		if _, err := rawDB.ExecContext(ctx,
			`CREATE TABLE IF NOT EXISTS schema_migrations (
				version TEXT PRIMARY KEY,
				applied_at TEXT NOT NULL
			)`,
		); err != nil {
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

		// Seed three distinctive rows into entries.
		now := time.Now().UTC().Format(time.RFC3339)
		for _, row := range []struct {
			title, tags, project, impact string
		}{
			{"alpha-backfill", "auth", "platform", "x"},
			{"beta-backfill", "perf", "growth", "y"},
			{"gamma-backfill", "backend", "platform", "z"},
		} {
			if _, err := rawDB.ExecContext(ctx,
				`INSERT INTO entries (title, description, tags, project, type, impact, created_at, updated_at)
				 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
				row.title, "desc", row.tags, row.project, "shipped", row.impact, now, now,
			); err != nil {
				t.Fatalf("seed row %q: %v", row.title, err)
			}
		}
	}()

	// Step 2: Open — runner should apply 0002 and backfill entries_fts.
	s, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	// Step 3: assert entries_fts has three rows matching the seeded titles.
	db := openRawDB(t, path)

	rows, err := db.Query(`SELECT rowid, title FROM entries_fts ORDER BY rowid`)
	if err != nil {
		t.Fatalf("query entries_fts: %v", err)
	}
	defer rows.Close()

	type indexed struct {
		rowid int64
		title string
	}
	var got []indexed
	for rows.Next() {
		var e indexed
		if err := rows.Scan(&e.rowid, &e.title); err != nil {
			t.Fatalf("scan: %v", err)
		}
		got = append(got, e)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows.Err: %v", err)
	}

	wantTitles := []string{"alpha-backfill", "beta-backfill", "gamma-backfill"}
	if len(got) != len(wantTitles) {
		t.Fatalf("entries_fts rows = %+v, want %d rows with titles %v", got, len(wantTitles), wantTitles)
	}
	for i, e := range got {
		if e.title != wantTitles[i] {
			t.Errorf("row %d title = %q, want %q", i, e.title, wantTitles[i])
		}
	}

	// Step 4: schema_migrations now contains both versions.
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM schema_migrations`).Scan(&count); err != nil {
		t.Fatalf("count schema_migrations: %v", err)
	}
	if count != 2 {
		t.Fatalf("schema_migrations count = %d, want 2", count)
	}
}

// TestFTS_TriggerInsertAddsToIndex asserts that Store.Add causes the
// row to appear in entries_fts (AFTER INSERT trigger).
func TestFTS_TriggerInsertAddsToIndex(t *testing.T) {
	s, path := newTestStore(t)

	inserted, err := s.Add(Entry{Title: "insert-trigger-xyz"})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}

	db := openRawDB(t, path)

	var (
		rowid int64
		title string
	)
	err = db.QueryRow(
		`SELECT rowid, title FROM entries_fts WHERE rowid = ?`, inserted.ID,
	).Scan(&rowid, &title)
	if err == sql.ErrNoRows {
		t.Fatalf("entries_fts has no row for id=%d", inserted.ID)
	}
	if err != nil {
		t.Fatalf("query entries_fts: %v", err)
	}
	if rowid != inserted.ID {
		t.Errorf("rowid = %d, want %d", rowid, inserted.ID)
	}
	if title != "insert-trigger-xyz" {
		t.Errorf("title = %q, want %q", title, "insert-trigger-xyz")
	}
}

// TestFTS_TriggerUpdateReplacesIndexedRow asserts that Store.Update
// replaces the indexed content — the old title no longer matches,
// the new title does.
func TestFTS_TriggerUpdateReplacesIndexedRow(t *testing.T) {
	s, path := newTestStore(t)

	inserted, err := s.Add(Entry{Title: "old-trigger-phrase"})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}

	if _, err := s.Update(inserted.ID, Entry{Title: "new-trigger-phrase"}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	db := openRawDB(t, path)

	// FTS5's MATCH syntax parses '-' as a unary NOT operator, so the
	// hyphenated titles must be wrapped in double quotes to be treated
	// as phrase literals.
	oldExpr := `"old-trigger-phrase"`
	newExpr := `"new-trigger-phrase"`

	// Old phrase must NOT match.
	var oldRow int64
	err = db.QueryRow(
		`SELECT rowid FROM entries_fts WHERE entries_fts MATCH ?`, oldExpr,
	).Scan(&oldRow)
	if err != sql.ErrNoRows {
		t.Fatalf("MATCH %s returned rowid=%d err=%v, want ErrNoRows", oldExpr, oldRow, err)
	}

	// New phrase must match exactly one row with the correct id.
	var newRow int64
	err = db.QueryRow(
		`SELECT rowid FROM entries_fts WHERE entries_fts MATCH ?`, newExpr,
	).Scan(&newRow)
	if err != nil {
		t.Fatalf("MATCH %s: %v", newExpr, err)
	}
	if newRow != inserted.ID {
		t.Errorf("MATCH %s rowid = %d, want %d", newExpr, newRow, inserted.ID)
	}
}

// TestFTS_TriggerDeleteRemovesFromIndex asserts Store.Delete strips
// the row from entries_fts.
func TestFTS_TriggerDeleteRemovesFromIndex(t *testing.T) {
	s, path := newTestStore(t)

	inserted, err := s.Add(Entry{Title: "to-be-removed-from-fts"})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}

	db := openRawDB(t, path)

	// Pre-check: it's there.
	var pre int
	if err := db.QueryRow(
		`SELECT COUNT(*) FROM entries_fts WHERE rowid = ?`, inserted.ID,
	).Scan(&pre); err != nil {
		t.Fatalf("pre count: %v", err)
	}
	if pre != 1 {
		t.Fatalf("pre count = %d, want 1", pre)
	}

	if err := s.Delete(inserted.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	var post int
	if err := db.QueryRow(
		`SELECT COUNT(*) FROM entries_fts WHERE rowid = ?`, inserted.ID,
	).Scan(&post); err != nil {
		t.Fatalf("post count: %v", err)
	}
	if post != 0 {
		t.Errorf("post count = %d, want 0", post)
	}
}

// TestFTS_MatchQueryReturnsExpectedIds asserts FTS MATCH returns the
// correct entry id for distinctive title tokens.
func TestFTS_MatchQueryReturnsExpectedIds(t *testing.T) {
	s, path := newTestStore(t)

	z, err := s.Add(Entry{Title: "zebrafish"})
	if err != nil {
		t.Fatalf("Add zebrafish: %v", err)
	}
	p, err := s.Add(Entry{Title: "platypus"})
	if err != nil {
		t.Fatalf("Add platypus: %v", err)
	}
	q, err := s.Add(Entry{Title: "quokka"})
	if err != nil {
		t.Fatalf("Add quokka: %v", err)
	}

	db := openRawDB(t, path)

	cases := []struct {
		token  string
		wantID int64
	}{
		{"zebrafish", z.ID},
		{"platypus", p.ID},
		{"quokka", q.ID},
	}
	for _, tc := range cases {
		var gotID int64
		err := db.QueryRow(
			`SELECT rowid FROM entries_fts WHERE entries_fts MATCH ? ORDER BY rowid`,
			tc.token,
		).Scan(&gotID)
		if err != nil {
			t.Fatalf("MATCH %q: %v", tc.token, err)
		}
		if gotID != tc.wantID {
			t.Errorf("MATCH %q rowid = %d, want %d", tc.token, gotID, tc.wantID)
		}
	}
}

// TestFTS_UnicodeTokenizerSplitsOnPunctuation locks decision #3:
// the default unicode61 tokenizer treats commas as separators, so a
// row with tags="auth,perf,backend" matches a search for "auth" and
// "perf" as individual tokens.
func TestFTS_UnicodeTokenizerSplitsOnPunctuation(t *testing.T) {
	s, path := newTestStore(t)

	inserted, err := s.Add(Entry{
		Title: "tokenizer-row",
		Tags:  "auth,perf,backend",
	})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}

	db := openRawDB(t, path)

	for _, token := range []string{"auth", "perf", "backend"} {
		var got int64
		err := db.QueryRow(
			`SELECT rowid FROM entries_fts WHERE entries_fts MATCH ?`, token,
		).Scan(&got)
		if err != nil {
			t.Fatalf("MATCH %q: %v", token, err)
		}
		if got != inserted.ID {
			t.Errorf("MATCH %q rowid = %d, want %d", token, got, inserted.ID)
		}
	}

	var missing int64
	err = db.QueryRow(
		`SELECT rowid FROM entries_fts WHERE entries_fts MATCH ?`,
		"xxx_missing_tag",
	).Scan(&missing)
	if err != sql.ErrNoRows {
		t.Fatalf("MATCH xxx_missing_tag: rowid=%d err=%v, want ErrNoRows", missing, err)
	}
}

// --- SPEC-012: Store.Search ----------------------------------------

// TestSearch_OrdersByRelevanceThenIdDesc seeds three rows that match
// a single query token exactly once in the same indexed field, so the
// FTS5 rank is identical across rows. The `id DESC` tie-break must
// then yield highest-id first.
func TestSearch_OrdersByRelevanceThenIdDesc(t *testing.T) {
	s, _ := newTestStore(t)

	// Three otherwise-identical entries, each with the token "rankword"
	// once in title. rank will be the same across all three, forcing
	// the id-desc tie-break to determine ordering.
	a, err := s.Add(Entry{Title: "rankword"})
	if err != nil {
		t.Fatalf("Add a: %v", err)
	}
	b, err := s.Add(Entry{Title: "rankword"})
	if err != nil {
		t.Fatalf("Add b: %v", err)
	}
	c, err := s.Add(Entry{Title: "rankword"})
	if err != nil {
		t.Fatalf("Add c: %v", err)
	}

	got, err := s.Search(`"rankword"`, 0)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("len(got) = %d, want 3", len(got))
	}
	wantIDs := []int64{c.ID, b.ID, a.ID}
	for i, e := range got {
		if e.ID != wantIDs[i] {
			t.Errorf("position %d: id=%d, want %d (full order: got=%v want=%v)",
				i, e.ID, wantIDs[i],
				[]int64{got[0].ID, got[1].ID, got[2].ID}, wantIDs)
		}
	}
}

// TestSearch_ReturnsHydratedEntries asserts Search populates every
// field on the Entry, not just id/title/created_at.
func TestSearch_ReturnsHydratedEntries(t *testing.T) {
	s, _ := newTestStore(t)

	inserted, err := s.Add(Entry{
		Title:       "uniquetitle",
		Description: "full description text",
		Tags:        "auth,perf",
		Project:     "platform",
		Type:        "shipped",
		Impact:      "notable impact",
	})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}

	got, err := s.Search(`"uniquetitle"`, 0)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len(got) = %d, want 1", len(got))
	}
	e := got[0]
	if e.ID != inserted.ID {
		t.Errorf("ID = %d, want %d", e.ID, inserted.ID)
	}
	if e.Title != "uniquetitle" {
		t.Errorf("Title = %q", e.Title)
	}
	if e.Description != "full description text" {
		t.Errorf("Description = %q", e.Description)
	}
	if e.Tags != "auth,perf" {
		t.Errorf("Tags = %q", e.Tags)
	}
	if e.Project != "platform" {
		t.Errorf("Project = %q", e.Project)
	}
	if e.Type != "shipped" {
		t.Errorf("Type = %q", e.Type)
	}
	if e.Impact != "notable impact" {
		t.Errorf("Impact = %q", e.Impact)
	}
	if e.CreatedAt.IsZero() {
		t.Errorf("CreatedAt is zero")
	}
	if e.UpdatedAt.IsZero() {
		t.Errorf("UpdatedAt is zero")
	}
}

// TestSearch_ZeroLimitMeansUnlimited asserts limit<=0 applies no LIMIT
// clause, matching Store.List's zero/negative-as-unset convention.
func TestSearch_ZeroLimitMeansUnlimited(t *testing.T) {
	s, _ := newTestStore(t)

	for i := 0; i < 7; i++ {
		if _, err := s.Add(Entry{Title: "limitword row"}); err != nil {
			t.Fatalf("Add %d: %v", i, err)
		}
	}

	zero, err := s.Search(`"limitword"`, 0)
	if err != nil {
		t.Fatalf("Search(0): %v", err)
	}
	if len(zero) != 7 {
		t.Errorf("Search(.., 0) len = %d, want 7", len(zero))
	}

	neg, err := s.Search(`"limitword"`, -1)
	if err != nil {
		t.Fatalf("Search(-1): %v", err)
	}
	if len(neg) != 7 {
		t.Errorf("Search(.., -1) len = %d, want 7", len(neg))
	}
}
