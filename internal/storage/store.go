package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// Store is the typed wrapper around *sql.DB for the bragfile database.
// All persistence flows through a Store; no other package imports a
// SQL driver.
type Store struct {
	db *sql.DB
}

// Open opens the SQLite database at path, creating the parent
// directory if needed, and applies any pending embedded migrations
// before returning.
func Open(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("open store: %w", err)
	}

	// sql.Open is lazy — the driver creates the file 0644 on first query.
	// Pre-create at 0600 so personal brag data is never world-readable.
	if f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0o600); err == nil {
		_ = f.Close()
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open store: %w", err)
	}

	sub, err := fs.Sub(migrationsFS, "migrations")
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("open store: %w", err)
	}

	// Safety belt: snapshot an existing DB before any forward-only,
	// irreversible migration (DEC-002) mutates it. No-op for a brand-new or
	// already-current DB. A failed snapshot aborts Open rather than migrate
	// an un-backed-up DB (DEC-021).
	ctx := context.Background()
	if err := backupBeforeMigrations(ctx, db, path, sub); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("open store: %w", err)
	}
	if err := applyMigrations(ctx, db, sub); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("open store: %w", err)
	}

	return &Store{db: db}, nil
}

// Close closes the underlying *sql.DB.
func (s *Store) Close() error {
	if err := s.db.Close(); err != nil {
		return fmt.Errorf("close store: %w", err)
	}
	return nil
}

// canonicalizeTags splits a comma-joined tag string, trims whitespace,
// drops empty tokens, and deduplicates keeping first occurrence.
// Returns nil when there are no tokens.
func canonicalizeTags(raw string) []string {
	parts := strings.Split(raw, ",")
	seen := make(map[string]bool)
	var tokens []string
	for _, p := range parts {
		t := strings.TrimSpace(p)
		if t == "" || seen[t] {
			continue
		}
		seen[t] = true
		tokens = append(tokens, t)
	}
	return tokens
}

// insertTaggings writes canonical tags into the tags + taggings tables
// for the given entry inside the provided transaction. Position is
// 0-based (index in the canonical token slice).
func insertTaggings(ctx context.Context, tx *sql.Tx, entryID int64, tokens []string) error {
	for i, name := range tokens {
		if _, err := tx.ExecContext(ctx,
			`INSERT OR IGNORE INTO tags(name) VALUES (?)`, name,
		); err != nil {
			return fmt.Errorf("insert tag %q: %w", name, err)
		}
		var tagID int64
		if err := tx.QueryRowContext(ctx,
			`SELECT id FROM tags WHERE name = ?`, name,
		).Scan(&tagID); err != nil {
			return fmt.Errorf("lookup tag %q: %w", name, err)
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO taggings(tag_id, taggable_type, taggable_id, position)
			 VALUES (?, 'entry', ?, ?)`,
			tagID, entryID, i,
		); err != nil {
			return fmt.Errorf("insert tagging %q: %w", name, err)
		}
	}
	return nil
}

// tagsProjection is the correlated scalar subquery fragment that reconstructs
// Entry.Tags from the taggings join (validated at design, §12(b)).
const tagsProjection = `COALESCE((
    SELECT GROUP_CONCAT(t.name, ',' ORDER BY tg.position)
      FROM taggings tg
      JOIN tags t ON t.id = tg.tag_id
     WHERE tg.taggable_type = 'entry' AND tg.taggable_id = e.id
), '') AS tags`

// Add inserts e and returns it hydrated with the generated ID and
// CreatedAt/UpdatedAt timestamps (both set to now, UTC, RFC3339).
// Tags are canonicalized (trim, dedup) and stored via the join.
func (s *Store) Add(e Entry) (Entry, error) {
	now := time.Now().UTC().Truncate(time.Second)
	ts := now.Format(time.RFC3339)

	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Entry{}, fmt.Errorf("add entry: %w", err)
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx,
		`INSERT INTO entries (title, description, project, type, impact, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		e.Title, e.Description, e.Project, e.Type, e.Impact, ts, ts,
	)
	if err != nil {
		return Entry{}, fmt.Errorf("add entry: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Entry{}, fmt.Errorf("add entry: %w", err)
	}

	tokens := canonicalizeTags(e.Tags)
	if err := insertTaggings(ctx, tx, id, tokens); err != nil {
		return Entry{}, fmt.Errorf("add entry: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return Entry{}, fmt.Errorf("add entry: %w", err)
	}

	e.ID = id
	e.CreatedAt = now
	e.UpdatedAt = now
	e.Tags = strings.Join(tokens, ",")
	return e, nil
}

// Get returns the entry with the given id. Returns an error wrapping
// ErrNotFound if no row matches.
func (s *Store) Get(id int64) (Entry, error) {
	var (
		e                                 Entry
		description, project, typ, impact sql.NullString
		tags                              string
		createdAtRaw, updatedAtRaw        string
	)
	row := s.db.QueryRowContext(context.Background(),
		`SELECT e.id, e.title, e.description, `+tagsProjection+`,
		        e.project, e.type, e.impact, e.created_at, e.updated_at
		 FROM entries e
		 WHERE e.id = ?`,
		id,
	)
	if err := row.Scan(
		&e.ID, &e.Title, &description, &tags, &project, &typ, &impact,
		&createdAtRaw, &updatedAtRaw,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Entry{}, fmt.Errorf("get entry %d: %w", id, ErrNotFound)
		}
		return Entry{}, fmt.Errorf("get entry %d: %w", id, err)
	}
	e.Description = description.String
	e.Tags = tags
	e.Project = project.String
	e.Type = typ.String
	e.Impact = impact.String

	created, err := time.Parse(time.RFC3339, createdAtRaw)
	if err != nil {
		return Entry{}, fmt.Errorf("get entry %d: parse created_at %q: %w", id, createdAtRaw, err)
	}
	updated, err := time.Parse(time.RFC3339, updatedAtRaw)
	if err != nil {
		return Entry{}, fmt.Errorf("get entry %d: parse updated_at %q: %w", id, updatedAtRaw, err)
	}
	e.CreatedAt = created.UTC()
	e.UpdatedAt = updated.UTC()

	return e, nil
}

// Update replaces every user-editable field on the row with id and
// bumps updated_at. Tags are replaced atomically via the join.
// Returns the hydrated Entry (via a follow-up Get); id and created_at
// are preserved. Returns an error wrapping ErrNotFound if no row matches.
func (s *Store) Update(id int64, e Entry) (Entry, error) {
	now := time.Now().UTC().Truncate(time.Second)
	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Entry{}, fmt.Errorf("update entry %d: %w", id, err)
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx,
		`UPDATE entries
		 SET title = ?, description = ?, project = ?, type = ?, impact = ?, updated_at = ?
		 WHERE id = ?`,
		e.Title, e.Description, e.Project, e.Type, e.Impact,
		now.Format(time.RFC3339), id,
	)
	if err != nil {
		return Entry{}, fmt.Errorf("update entry %d: %w", id, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return Entry{}, fmt.Errorf("update entry %d: %w", id, err)
	}
	if n == 0 {
		return Entry{}, fmt.Errorf("update entry %d: %w", id, ErrNotFound)
	}

	if _, err := tx.ExecContext(ctx,
		`DELETE FROM taggings WHERE taggable_type = 'entry' AND taggable_id = ?`, id,
	); err != nil {
		return Entry{}, fmt.Errorf("update entry %d: remove taggings: %w", id, err)
	}

	tokens := canonicalizeTags(e.Tags)
	if err := insertTaggings(ctx, tx, id, tokens); err != nil {
		return Entry{}, fmt.Errorf("update entry %d: %w", id, err)
	}

	if err := tx.Commit(); err != nil {
		return Entry{}, fmt.Errorf("update entry %d: %w", id, err)
	}

	return s.Get(id)
}

// Delete removes the entry with the given id and its taggings.
// Returns an error wrapping ErrNotFound if no row matches.
func (s *Store) Delete(id int64) error {
	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("delete entry %d: %w", id, err)
	}
	defer tx.Rollback()

	// Delete taggings first; entries_ad trigger removes the FTS row.
	if _, err := tx.ExecContext(ctx,
		`DELETE FROM taggings WHERE taggable_type = 'entry' AND taggable_id = ?`, id,
	); err != nil {
		return fmt.Errorf("delete entry %d: remove taggings: %w", id, err)
	}

	res, err := tx.ExecContext(ctx, `DELETE FROM entries WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete entry %d: %w", id, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete entry %d: %w", id, err)
	}
	if n == 0 {
		return fmt.Errorf("delete entry %d: %w", id, ErrNotFound)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("delete entry %d: %w", id, err)
	}
	return nil
}

// List returns entries matching the populated fields of f, combined
// via AND, ordered created_at DESC with id DESC as the tie-break.
// A zero-value ListFilter returns every row.
func (s *Store) List(f ListFilter) ([]Entry, error) {
	var conds []string
	var args []any

	if f.Tag != "" {
		// Exact tag-token membership via the normalized taggings join (DEC-015).
		conds = append(conds, `EXISTS (SELECT 1 FROM taggings tg
		    JOIN tags t ON t.id = tg.tag_id
		   WHERE tg.taggable_type = 'entry' AND tg.taggable_id = e.id
		     AND t.name = ?)`)
		args = append(args, f.Tag)
	}
	if f.Project != "" {
		conds = append(conds, "e.project = ?")
		args = append(args, f.Project)
	}
	if f.Type != "" {
		conds = append(conds, "e.type = ?")
		args = append(args, f.Type)
	}
	if !f.Since.IsZero() {
		conds = append(conds, "e.created_at >= ?")
		args = append(args, f.Since.UTC().Format(time.RFC3339))
	}

	q := `SELECT e.id, e.title, e.description, ` + tagsProjection + `,
		         e.project, e.type, e.impact, e.created_at, e.updated_at
		  FROM entries e`
	if len(conds) > 0 {
		q += " WHERE " + strings.Join(conds, " AND ")
	}
	q += " ORDER BY e.created_at DESC, e.id DESC"
	if f.Limit > 0 {
		q += " LIMIT ?"
		args = append(args, f.Limit)
	}

	rows, err := s.db.QueryContext(context.Background(), q, args...)
	if err != nil {
		return nil, fmt.Errorf("list entries: %w", err)
	}
	defer rows.Close()

	out := make([]Entry, 0)
	for rows.Next() {
		var (
			e                                 Entry
			description, project, typ, impact sql.NullString
			tags                              string
			createdAtRaw, updatedAtRaw        string
		)
		if err := rows.Scan(
			&e.ID, &e.Title, &description, &tags, &project, &typ, &impact,
			&createdAtRaw, &updatedAtRaw,
		); err != nil {
			return nil, fmt.Errorf("list entries: %w", err)
		}
		e.Description = description.String
		e.Tags = tags
		e.Project = project.String
		e.Type = typ.String
		e.Impact = impact.String

		created, err := time.Parse(time.RFC3339, createdAtRaw)
		if err != nil {
			return nil, fmt.Errorf("list entries: parse created_at %q: %w", createdAtRaw, err)
		}
		updated, err := time.Parse(time.RFC3339, updatedAtRaw)
		if err != nil {
			return nil, fmt.Errorf("list entries: parse updated_at %q: %w", updatedAtRaw, err)
		}
		e.CreatedAt = created.UTC()
		e.UpdatedAt = updated.UTC()

		out = append(out, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list entries: %w", err)
	}
	return out, nil
}

// TagCount is one row of the brag tags taxonomy view: a tag name and
// its total membership count across all taggable types (DEC-016).
type TagCount struct {
	Name  string
	Count int
}

// TagCounts returns every in-use tag (count >= 1) with its total
// taggings count across all taggable_types, ordered count DESC then
// name ASC (DEC-016 choice 1). Orphan tags (zero taggings) are omitted.
func (s *Store) TagCounts() ([]TagCount, error) {
	rows, err := s.db.QueryContext(context.Background(),
		`SELECT t.name, COUNT(tg.id) AS cnt
		   FROM tags t
		   JOIN taggings tg ON tg.tag_id = t.id
		  GROUP BY t.id, t.name
		  ORDER BY cnt DESC, t.name ASC`)
	if err != nil {
		return nil, fmt.Errorf("tag counts: %w", err)
	}
	defer rows.Close()
	out := make([]TagCount, 0)
	for rows.Next() {
		var tc TagCount
		if err := rows.Scan(&tc.Name, &tc.Count); err != nil {
			return nil, fmt.Errorf("tag counts: %w", err)
		}
		out = append(out, tc)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("tag counts: %w", err)
	}
	return out, nil
}

// RenameTag renames the tag with oldName to newName everywhere. Fires
// the tags_au trigger which re-syncs entries_fts automatically (DEC-016).
// Returns ErrTagExists if newName already exists; ErrTagNotFound if
// oldName does not exist. The caller guards oldName == newName.
func (s *Store) RenameTag(oldName, newName string) error {
	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("rename tag: %w", err)
	}
	defer tx.Rollback()

	var exists int
	if err := tx.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM tags WHERE name = ?`, newName,
	).Scan(&exists); err != nil {
		return fmt.Errorf("rename tag: %w", err)
	}
	if exists > 0 {
		return fmt.Errorf("rename tag %q -> %q: %w", oldName, newName, ErrTagExists)
	}

	res, err := tx.ExecContext(ctx,
		`UPDATE tags SET name = ? WHERE name = ?`, newName, oldName)
	if err != nil {
		return fmt.Errorf("rename tag: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rename tag: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("rename tag %q: %w", oldName, ErrTagNotFound)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("rename tag: %w", err)
	}
	return nil
}

// MergeTags folds src's taggings into dst via DELETE+INSERT (DEC-016
// choice 3), de-duplicating, then deletes the src tag row. Both src
// and dst must exist; caller guards src == dst.
func (s *Store) MergeTags(src, dst string) error {
	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("merge tags: %w", err)
	}
	defer tx.Rollback()

	var srcID, dstID int64
	if err := tx.QueryRowContext(ctx,
		`SELECT id FROM tags WHERE name = ?`, src).Scan(&srcID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("merge tags: source %q: %w", src, ErrTagNotFound)
		}
		return fmt.Errorf("merge tags: %w", err)
	}
	if err := tx.QueryRowContext(ctx,
		`SELECT id FROM tags WHERE name = ?`, dst).Scan(&dstID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("merge tags: destination %q: %w", dst, ErrTagNotFound)
		}
		return fmt.Errorf("merge tags: %w", err)
	}

	// 1. Give every src-tagged object a dst tagging it doesn't already
	//    have (fires taggings_ai → FTS re-sync). NOT EXISTS skips the
	//    would-be UNIQUE duplicate for objects tagged both.
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO taggings (tag_id, taggable_type, taggable_id, position)
		 SELECT ?, s.taggable_type, s.taggable_id, s.position
		   FROM taggings s
		  WHERE s.tag_id = ?
		    AND NOT EXISTS (SELECT 1 FROM taggings d
		                     WHERE d.tag_id = ?
		                       AND d.taggable_type = s.taggable_type
		                       AND d.taggable_id  = s.taggable_id)`,
		dstID, srcID, dstID); err != nil {
		return fmt.Errorf("merge tags: graft dst: %w", err)
	}
	// 2. Drop all src taggings (fires taggings_ad → FTS re-sync, removing
	//    the src token from the projection where it still lingered).
	if _, err := tx.ExecContext(ctx,
		`DELETE FROM taggings WHERE tag_id = ?`, srcID); err != nil {
		return fmt.Errorf("merge tags: drop src taggings: %w", err)
	}
	// 3. Remove the now-unreferenced src tag row.
	if _, err := tx.ExecContext(ctx,
		`DELETE FROM tags WHERE id = ?`, srcID); err != nil {
		return fmt.Errorf("merge tags: drop src tag: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("merge tags: %w", err)
	}
	return nil
}

// Search returns entries whose entries_fts row matches the given FTS5
// query, ordered by FTS5 rank ascending (most relevant first) with
// entries.id DESC as the tie-break. The caller (CLI layer) is
// responsible for validating and transforming user input into a valid
// MATCH expression per DEC-010; Store.Search passes query to MATCH
// verbatim and assumes it is non-empty.
//
// limit <= 0 means no LIMIT is applied (matches Store.List's zero-is-
// no-limit convention).
func (s *Store) Search(query string, limit int) ([]Entry, error) {
	q := `SELECT e.id, e.title, e.description, ` + tagsProjection + `,
		         e.project, e.type, e.impact, e.created_at, e.updated_at
		  FROM entries_fts
		  JOIN entries e ON e.id = entries_fts.rowid
		  WHERE entries_fts MATCH ?
		  ORDER BY rank, e.id DESC`
	args := []any{query}
	if limit > 0 {
		q += " LIMIT ?"
		args = append(args, limit)
	}

	rows, err := s.db.QueryContext(context.Background(), q, args...)
	if err != nil {
		return nil, fmt.Errorf("search entries: %w", err)
	}
	defer rows.Close()

	out := make([]Entry, 0)
	for rows.Next() {
		var (
			e                                 Entry
			description, project, typ, impact sql.NullString
			tags                              string
			createdAtRaw, updatedAtRaw        string
		)
		if err := rows.Scan(
			&e.ID, &e.Title, &description, &tags, &project, &typ, &impact,
			&createdAtRaw, &updatedAtRaw,
		); err != nil {
			return nil, fmt.Errorf("search entries: %w", err)
		}
		e.Description = description.String
		e.Tags = tags
		e.Project = project.String
		e.Type = typ.String
		e.Impact = impact.String

		created, err := time.Parse(time.RFC3339, createdAtRaw)
		if err != nil {
			return nil, fmt.Errorf("search entries: parse created_at %q: %w", createdAtRaw, err)
		}
		updated, err := time.Parse(time.RFC3339, updatedAtRaw)
		if err != nil {
			return nil, fmt.Errorf("search entries: parse updated_at %q: %w", updatedAtRaw, err)
		}
		e.CreatedAt = created.UTC()
		e.UpdatedAt = updated.UTC()

		out = append(out, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("search entries: %w", err)
	}
	return out, nil
}
