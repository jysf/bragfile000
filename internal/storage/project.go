package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
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

// RemoveLocation detaches path from the project identified by projectID. It is
// the single-path counterpart to AddLocation. Paths match VERBATIM against the
// stored value (storage is verbatim end to end; SPEC-031/DEC-019 own
// normalization at cwd-resolve time only). Errors:
//   - ErrLocationNotFound      if path is attached to no project
//   - ErrLocationOtherProject  if path is attached to a different project
//
// (DEC-020). Implemented over the same transactional engine as EditLocations
// so a single remove and a batch share one validated code path.
func (s *Store) RemoveLocation(projectID int64, path string) error {
	return s.EditLocations(projectID, []string{path}, nil)
}

// EditLocations applies a set of location removals and additions to the project
// identified by projectID in ONE transaction, all-or-nothing (DEC-020). Removes
// are applied before adds, so a path may be removed and re-added in the same
// call without a transient UNIQUE(path) collision. Any failure rolls the whole
// set back, leaving project_locations unchanged. Per-path rules:
//   - remove: path must be attached to projectID
//     (else ErrLocationNotFound / ErrLocationOtherProject)
//   - add:    path must be free after the removes
//     (else ErrLocationExists — the same global-uniqueness guard
//     AddLocation enforces; the in-tx COUNT also catches in-batch dups)
//
// Paths match verbatim. updated_at is NOT bumped: location editing is a
// structural change to project_locations, distinct from the scalar-field
// recency UpdateProject tracks (DEC-020).
func (s *Store) EditLocations(projectID int64, remove, add []string) error {
	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("edit locations: %w", err)
	}
	defer tx.Rollback()

	// Removes first — frees any UNIQUE path that a later add re-registers.
	for _, path := range remove {
		var ownerID int64
		err := tx.QueryRowContext(ctx,
			`SELECT project_id FROM project_locations WHERE path = ?`, path,
		).Scan(&ownerID)
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("edit locations: remove %q: %w", path, ErrLocationNotFound)
		}
		if err != nil {
			return fmt.Errorf("edit locations: remove %q: %w", path, err)
		}
		if ownerID != projectID {
			return fmt.Errorf("edit locations: remove %q: %w", path, ErrLocationOtherProject)
		}
		if _, err := tx.ExecContext(ctx,
			`DELETE FROM project_locations WHERE path = ?`, path,
		); err != nil {
			return fmt.Errorf("edit locations: remove %q: %w", path, err)
		}
	}

	// Adds — path must be free; the in-tx COUNT also backstops in-batch dups.
	for _, path := range add {
		var exists int
		if err := tx.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM project_locations WHERE path = ?`, path,
		).Scan(&exists); err != nil {
			return fmt.Errorf("edit locations: add %q: %w", path, err)
		}
		if exists > 0 {
			return fmt.Errorf("edit locations: add %q: %w", path, ErrLocationExists)
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO project_locations (project_id, path) VALUES (?, ?)`,
			projectID, path,
		); err != nil {
			return fmt.Errorf("edit locations: add %q: %w", path, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("edit locations: %w", err)
	}
	return nil
}

// validProjectStatuses is the DEC-017 status enum. Validated in the Store
// (not a DB CHECK) so adding a value later is an additive change under the
// forward-only migration regime (DEC-002), mirroring entries.type's
// free-text column.
var validProjectStatuses = map[string]bool{
	"active": true, "paused": true, "done": true, "archived": true,
}

// UpdateProject replaces the editable scalar fields (Name, Status,
// StateNote) on the project with id and bumps updated_at. Locations are
// NOT edited here (SPEC-033). Returns the hydrated project (id and
// created_at preserved). Errors:
//   - ErrInvalidStatus if Status is not a DEC-017 enum value
//   - ErrProjectExists if Name is already taken by a *different* project
//   - ErrNotFound if no row matches id
//
// This is the first Store method to advance updated_at past created_at,
// making ListProjects' recency ordering observable.
func (s *Store) UpdateProject(id int64, p Project) (Project, error) {
	if p.Status == "" {
		p.Status = "active"
	}
	if !validProjectStatuses[p.Status] {
		return Project{}, fmt.Errorf("update project %d: status %q: %w", id, p.Status, ErrInvalidStatus)
	}
	now := time.Now().UTC().Truncate(time.Second)

	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Project{}, fmt.Errorf("update project %d: %w", id, err)
	}
	defer tx.Rollback()

	// Name is UNIQUE; a rename collides only with a *different* project.
	// Self-exclude (id != ?) so renaming to the current name is a no-op.
	var exists int
	if err := tx.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM projects WHERE name = ? AND id != ?`, p.Name, id,
	).Scan(&exists); err != nil {
		return Project{}, fmt.Errorf("update project %d: %w", id, err)
	}
	if exists > 0 {
		return Project{}, fmt.Errorf("update project %d name %q: %w", id, p.Name, ErrProjectExists)
	}

	res, err := tx.ExecContext(ctx,
		`UPDATE projects SET name = ?, status = ?, state_note = ?, updated_at = ?
		 WHERE id = ?`,
		p.Name, p.Status, p.StateNote, now.Format(time.RFC3339), id,
	)
	if err != nil {
		return Project{}, fmt.Errorf("update project %d: %w", id, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return Project{}, fmt.Errorf("update project %d: %w", id, err)
	}
	if n == 0 {
		return Project{}, fmt.Errorf("update project %d: %w", id, ErrNotFound)
	}

	if err := tx.Commit(); err != nil {
		return Project{}, fmt.Errorf("update project %d: %w", id, err)
	}
	return s.GetProject(id)
}

// ArchiveProject sets the project's status to "archived" (the DEC-017
// non-destructive flip) and bumps updated_at. Name, state_note, and
// locations are preserved — archive is recoverable via
// UpdateProject(..., Status:"active"). Returns ErrNotFound if no row matches.
func (s *Store) ArchiveProject(id int64) error {
	now := time.Now().UTC().Truncate(time.Second)
	res, err := s.db.ExecContext(context.Background(),
		`UPDATE projects SET status = 'archived', updated_at = ? WHERE id = ?`,
		now.Format(time.RFC3339), id,
	)
	if err != nil {
		return fmt.Errorf("archive project %d: %w", id, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("archive project %d: %w", id, err)
	}
	if n == 0 {
		return fmt.Errorf("archive project %d: %w", id, ErrNotFound)
	}
	return nil
}

// DeleteProject permanently removes the project with id and its full blast
// radius (DEC-018), all in one transaction and in this order:
//  1. any 'project' taggings (none are written yet — schema-ready per
//     DEC-015 — but cleaned now so a future project-tag surface needs no
//     delete change)
//  2. the project's project_locations rows — FK enforcement is OFF in
//     bragfile, so the REFERENCES clause does NOT cascade; deleting these
//     manually is also what frees the globally-UNIQUE path for reuse
//  3. the projects row itself
//
// entries are deliberately UNTOUCHED (DEC-017 soft string match): an entry
// keeps the free-text project string it was captured with. Returns
// ErrNotFound if no row matches id.
func (s *Store) DeleteProject(id int64) error {
	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("delete project %d: %w", id, err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx,
		`DELETE FROM taggings WHERE taggable_type = 'project' AND taggable_id = ?`, id,
	); err != nil {
		return fmt.Errorf("delete project %d: remove taggings: %w", id, err)
	}
	if _, err := tx.ExecContext(ctx,
		`DELETE FROM project_locations WHERE project_id = ?`, id,
	); err != nil {
		return fmt.Errorf("delete project %d: remove locations: %w", id, err)
	}

	res, err := tx.ExecContext(ctx, `DELETE FROM projects WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete project %d: %w", id, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete project %d: %w", id, err)
	}
	if n == 0 {
		return fmt.Errorf("delete project %d: %w", id, ErrNotFound)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("delete project %d: %w", id, err)
	}
	return nil
}

// ProjectStatus is one row of the `brag project status` dashboard: a
// non-archived project plus its brag count — the number of entries whose
// free-text project string equals the project name (the DEC-017
// soft-string-match join, entries.project = projects.name). Locations are
// intentionally omitted: the dashboard is about activity, not where a
// project lives (that is `brag project show`).
type ProjectStatus struct {
	ID        int64
	Name      string
	Status    string
	StateNote string
	BragCount int
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ProjectStatuses returns every non-archived project (status != 'archived')
// ordered updated_at DESC, id DESC, each with its total brag count via the
// DEC-017 soft-string-match join (entries.project = projects.name). A
// project with no matching entries has BragCount 0 (LEFT JOIN + COUNT(e.id)).
// Returns a non-nil empty slice when no non-archived projects exist.
func (s *Store) ProjectStatuses() ([]ProjectStatus, error) {
	ctx := context.Background()
	rows, err := s.db.QueryContext(ctx,
		`SELECT p.id, p.name, p.status, p.state_note, p.created_at, p.updated_at,
		        COUNT(e.id) AS brag_count
		   FROM projects p
		   LEFT JOIN entries e ON e.project = p.name
		  WHERE p.status != 'archived'
		  GROUP BY p.id
		  ORDER BY p.updated_at DESC, p.id DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("project statuses: %w", err)
	}
	defer rows.Close()

	out := make([]ProjectStatus, 0)
	for rows.Next() {
		var (
			st                         ProjectStatus
			createdAtRaw, updatedAtRaw string
		)
		if err := rows.Scan(&st.ID, &st.Name, &st.Status, &st.StateNote,
			&createdAtRaw, &updatedAtRaw, &st.BragCount); err != nil {
			return nil, fmt.Errorf("project statuses: %w", err)
		}
		created, err := time.Parse(time.RFC3339, createdAtRaw)
		if err != nil {
			return nil, fmt.Errorf("project statuses: parse created_at %q: %w", createdAtRaw, err)
		}
		updated, err := time.Parse(time.RFC3339, updatedAtRaw)
		if err != nil {
			return nil, fmt.Errorf("project statuses: parse updated_at %q: %w", updatedAtRaw, err)
		}
		st.CreatedAt = created.UTC()
		st.UpdatedAt = updated.UTC()
		out = append(out, st)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("project statuses: %w", err)
	}
	return out, nil
}

// ProjectForPath resolves cwd against all registered project_locations
// using nearest-ancestor (longest-prefix) matching (DEC-019). Both cwd
// and each stored path are cleaned with filepath.Clean before comparison.
// Returns nil, nil — not ErrNotFound — when no location is an ancestor
// of cwd; callers distinguish "no project here" from a real error by the
// nil check. Locations is not hydrated on the returned Project.
func (s *Store) ProjectForPath(cwd string) (*Project, error) {
	cwd = filepath.Clean(cwd)
	sep := string(filepath.Separator)

	ctx := context.Background()
	rows, err := s.db.QueryContext(ctx,
		`SELECT pl.path, p.id, p.name, p.status, p.state_note,
		        p.created_at, p.updated_at
		   FROM project_locations pl
		   JOIN projects p ON p.id = pl.project_id`,
	)
	if err != nil {
		return nil, fmt.Errorf("project for path %q: %w", cwd, err)
	}
	defer rows.Close()

	var (
		bestPath    string
		bestProject *Project
	)

	for rows.Next() {
		var (
			locPath                    string
			p                          Project
			createdAtRaw, updatedAtRaw string
		)
		if err := rows.Scan(&locPath, &p.ID, &p.Name, &p.Status, &p.StateNote,
			&createdAtRaw, &updatedAtRaw); err != nil {
			return nil, fmt.Errorf("project for path %q: %w", cwd, err)
		}
		cleanLoc := filepath.Clean(locPath)

		// Nearest-ancestor check: cwd must equal cleanLoc (exact match) or
		// start with cleanLoc + separator (cleanLoc is a parent directory).
		// The separator suffix prevents /home/user/work matching /home/user/worker.
		if cwd != cleanLoc && !strings.HasPrefix(cwd, cleanLoc+sep) {
			continue
		}

		// Longest prefix wins (most-specific registered ancestor, DEC-019).
		if bestProject == nil || len(cleanLoc) > len(bestPath) {
			created, err := time.Parse(time.RFC3339, createdAtRaw)
			if err != nil {
				return nil, fmt.Errorf("project for path %q: parse created_at %q: %w",
					cwd, createdAtRaw, err)
			}
			updated, err := time.Parse(time.RFC3339, updatedAtRaw)
			if err != nil {
				return nil, fmt.Errorf("project for path %q: parse updated_at %q: %w",
					cwd, updatedAtRaw, err)
			}
			p.CreatedAt = created.UTC()
			p.UpdatedAt = updated.UTC()
			// Locations intentionally nil — callers that need the full
			// location list call GetProject(p.ID).
			bestPath = cleanLoc
			cp := p
			bestProject = &cp
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("project for path %q: %w", cwd, err)
	}
	return bestProject, nil
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
