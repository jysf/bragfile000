package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// Project is a first-class registered project (DEC-017, SPEC-027).
// Locations is the ordered slice of filesystem paths attached via AddLocation.
type Project struct {
	ID        int64
	Name      string
	Status    string
	StateNote string
	Locations []string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// CreateProject inserts a new project and returns it hydrated with its
// generated ID and timestamps. Status defaults to "active" when empty.
// Returns an error wrapping ErrProjectExists when the name is already taken.
func (s *Store) CreateProject(p Project) (Project, error) {
	if p.Status == "" {
		p.Status = "active"
	}
	now := time.Now().UTC().Truncate(time.Second)
	ts := now.Format(time.RFC3339)

	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Project{}, fmt.Errorf("create project: %w", err)
	}
	defer tx.Rollback()

	var exists int
	if err := tx.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM projects WHERE name = ?`, p.Name,
	).Scan(&exists); err != nil {
		return Project{}, fmt.Errorf("create project: %w", err)
	}
	if exists > 0 {
		return Project{}, fmt.Errorf("create project %q: %w", p.Name, ErrProjectExists)
	}

	res, err := tx.ExecContext(ctx,
		`INSERT INTO projects (name, status, state_note, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?)`,
		p.Name, p.Status, p.StateNote, ts, ts,
	)
	if err != nil {
		return Project{}, fmt.Errorf("create project: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Project{}, fmt.Errorf("create project: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return Project{}, fmt.Errorf("create project: %w", err)
	}

	p.ID = id
	p.CreatedAt = now
	p.UpdatedAt = now
	p.Locations = []string{}
	return p, nil
}

// GetProject returns the project with the given id, with its Locations
// hydrated in insertion order. Returns an error wrapping ErrNotFound if
// no row matches.
func (s *Store) GetProject(id int64) (Project, error) {
	ctx := context.Background()

	var (
		p                          Project
		createdAtRaw, updatedAtRaw string
	)
	err := s.db.QueryRowContext(ctx,
		`SELECT id, name, status, state_note, created_at, updated_at
		 FROM projects WHERE id = ?`, id,
	).Scan(&p.ID, &p.Name, &p.Status, &p.StateNote, &createdAtRaw, &updatedAtRaw)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Project{}, fmt.Errorf("get project %d: %w", id, ErrNotFound)
		}
		return Project{}, fmt.Errorf("get project %d: %w", id, err)
	}

	created, err := time.Parse(time.RFC3339, createdAtRaw)
	if err != nil {
		return Project{}, fmt.Errorf("get project %d: parse created_at %q: %w", id, createdAtRaw, err)
	}
	updated, err := time.Parse(time.RFC3339, updatedAtRaw)
	if err != nil {
		return Project{}, fmt.Errorf("get project %d: parse updated_at %q: %w", id, updatedAtRaw, err)
	}
	p.CreatedAt = created.UTC()
	p.UpdatedAt = updated.UTC()

	locs, err := s.locationsForProject(ctx, id)
	if err != nil {
		return Project{}, fmt.Errorf("get project %d: %w", id, err)
	}
	p.Locations = locs
	return p, nil
}

// GetProjectByName returns the project with the given name (names are
// globally UNIQUE), with its Locations hydrated in insertion order.
// Returns an error wrapping ErrNotFound if no row matches. Mirrors
// GetProject; used by `brag project show <name|id>` (SPEC-028).
func (s *Store) GetProjectByName(name string) (Project, error) {
	ctx := context.Background()

	var (
		p                          Project
		createdAtRaw, updatedAtRaw string
	)
	err := s.db.QueryRowContext(ctx,
		`SELECT id, name, status, state_note, created_at, updated_at
		 FROM projects WHERE name = ?`, name,
	).Scan(&p.ID, &p.Name, &p.Status, &p.StateNote, &createdAtRaw, &updatedAtRaw)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Project{}, fmt.Errorf("get project %q: %w", name, ErrNotFound)
		}
		return Project{}, fmt.Errorf("get project %q: %w", name, err)
	}

	created, err := time.Parse(time.RFC3339, createdAtRaw)
	if err != nil {
		return Project{}, fmt.Errorf("get project %q: parse created_at %q: %w", name, createdAtRaw, err)
	}
	updated, err := time.Parse(time.RFC3339, updatedAtRaw)
	if err != nil {
		return Project{}, fmt.Errorf("get project %q: parse updated_at %q: %w", name, updatedAtRaw, err)
	}
	p.CreatedAt = created.UTC()
	p.UpdatedAt = updated.UTC()

	locs, err := s.locationsForProject(ctx, p.ID)
	if err != nil {
		return Project{}, fmt.Errorf("get project %q: %w", name, err)
	}
	p.Locations = locs
	return p, nil
}

// ListProjects returns all projects ordered updated_at DESC, id DESC,
// each hydrated with its Locations. Returns a non-nil empty slice when
// no projects exist.
func (s *Store) ListProjects() ([]Project, error) {
	ctx := context.Background()

	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name, status, state_note, created_at, updated_at
		 FROM projects ORDER BY updated_at DESC, id DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	defer rows.Close()

	out := make([]Project, 0)
	for rows.Next() {
		var (
			p                          Project
			createdAtRaw, updatedAtRaw string
		)
		if err := rows.Scan(&p.ID, &p.Name, &p.Status, &p.StateNote, &createdAtRaw, &updatedAtRaw); err != nil {
			return nil, fmt.Errorf("list projects: %w", err)
		}
		created, err := time.Parse(time.RFC3339, createdAtRaw)
		if err != nil {
			return nil, fmt.Errorf("list projects: parse created_at %q: %w", createdAtRaw, err)
		}
		updated, err := time.Parse(time.RFC3339, updatedAtRaw)
		if err != nil {
			return nil, fmt.Errorf("list projects: parse updated_at %q: %w", updatedAtRaw, err)
		}
		p.CreatedAt = created.UTC()
		p.UpdatedAt = updated.UTC()
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}

	for i := range out {
		locs, err := s.locationsForProject(ctx, out[i].ID)
		if err != nil {
			return nil, fmt.Errorf("list projects: hydrate locations for %d: %w", out[i].ID, err)
		}
		out[i].Locations = locs
	}
	return out, nil
}

// AddLocation attaches path to the project identified by projectID.
// Paths are stored verbatim (SPEC-031 owns normalization).
// Returns an error wrapping ErrLocationExists if path is already attached
// to any project (paths are globally unique — the guarantee SPEC-031 relies on).
func (s *Store) AddLocation(projectID int64, path string) error {
	ctx := context.Background()

	var exists int
	if err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM project_locations WHERE path = ?`, path,
	).Scan(&exists); err != nil {
		return fmt.Errorf("add location: %w", err)
	}
	if exists > 0 {
		return fmt.Errorf("add location %q: %w", path, ErrLocationExists)
	}

	if _, err := s.db.ExecContext(ctx,
		`INSERT INTO project_locations (project_id, path) VALUES (?, ?)`,
		projectID, path,
	); err != nil {
		return fmt.Errorf("add location: %w", err)
	}
	return nil
}

// locationsForProject returns paths for the given project ordered by
// project_locations.id (insertion order).
func (s *Store) locationsForProject(ctx context.Context, projectID int64) ([]string, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT path FROM project_locations WHERE project_id = ? ORDER BY id`, projectID,
	)
	if err != nil {
		return nil, fmt.Errorf("locations for project %d: %w", projectID, err)
	}
	defer rows.Close()

	locs := make([]string, 0)
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			return nil, fmt.Errorf("locations for project %d: scan: %w", projectID, err)
		}
		locs = append(locs, path)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("locations for project %d: %w", projectID, err)
	}
	return locs, nil
}
