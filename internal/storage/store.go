package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
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
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("open store: %w", err)
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
	if err := applyMigrations(context.Background(), db, sub); err != nil {
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

// Add inserts e and returns it hydrated with the generated ID and
// CreatedAt/UpdatedAt timestamps (both set to now, UTC, RFC3339).
func (s *Store) Add(e Entry) (Entry, error) {
	now := time.Now().UTC().Truncate(time.Second)
	ts := now.Format(time.RFC3339)

	res, err := s.db.ExecContext(context.Background(),
		`INSERT INTO entries (
            title, description, tags, project, type, impact, created_at, updated_at
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		e.Title, e.Description, e.Tags, e.Project, e.Type, e.Impact, ts, ts,
	)
	if err != nil {
		return Entry{}, fmt.Errorf("add entry: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Entry{}, fmt.Errorf("add entry: %w", err)
	}

	e.ID = id
	e.CreatedAt = now
	e.UpdatedAt = now
	return e, nil
}

// Get returns the entry with the given id. Returns an error wrapping
// ErrNotFound if no row matches.
func (s *Store) Get(id int64) (Entry, error) {
	var (
		e                                       Entry
		description, tags, project, typ, impact sql.NullString
		createdAtRaw, updatedAtRaw              string
	)
	row := s.db.QueryRowContext(context.Background(),
		`SELECT id, title, description, tags, project, type, impact, created_at, updated_at
         FROM entries
         WHERE id = ?`,
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
	e.Tags = tags.String
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

// List returns all entries in created_at DESC order. The ListFilter
// fields are ignored for MVP (they land in STAGE-002).
func (s *Store) List(_ ListFilter) ([]Entry, error) {
	rows, err := s.db.QueryContext(context.Background(),
		`SELECT id, title, description, tags, project, type, impact, created_at, updated_at
         FROM entries
         ORDER BY created_at DESC, id DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("list entries: %w", err)
	}
	defer rows.Close()

	out := make([]Entry, 0)
	for rows.Next() {
		var (
			e                                       Entry
			description, tags, project, typ, impact sql.NullString
			createdAtRaw, updatedAtRaw              string
		)
		if err := rows.Scan(
			&e.ID, &e.Title, &description, &tags, &project, &typ, &impact,
			&createdAtRaw, &updatedAtRaw,
		); err != nil {
			return nil, fmt.Errorf("list entries: %w", err)
		}
		e.Description = description.String
		e.Tags = tags.String
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
