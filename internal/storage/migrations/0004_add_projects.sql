-- 0004_add_projects.sql — SPEC-027 (PROJ-002 / STAGE-007)
-- First-class projects entity (DEC-017). Adds the `projects` table and a
-- `project_locations` join supporting one-project-many-directories.
-- Forward-only (DEC-002); runs inside the migration runner's per-migration
-- transaction — do NOT add BEGIN/COMMIT or a schema_migrations insert here.
--
-- DEC-017 (soft string match): entries.project stays free text and is NOT
-- touched by this migration — no FK, no link column, no backfill. The
-- relationship is an opportunistic join on projects.name at query time.
-- Validated at design (§12(b)) against modernc.org/sqlite 1.51.0
-- (SQLite 3.53.1): tables create, name/path UNIQUE enforced, status and
-- state_note defaults applied, entries untouched.

-- The projects entity. status is a Store-validated enum (active | paused |
-- done | archived), not a DB CHECK — mirroring entries.type's free-text
-- column, so a future status value is an additive Store change, not a
-- table rebuild under forward-only migrations. state_note is the single
-- free-text state/next-action note rendered by brag project status (SPEC-030).
CREATE TABLE projects (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    name       TEXT NOT NULL UNIQUE,
    status     TEXT NOT NULL DEFAULT 'active',
    state_note TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

-- One project, many directories. path is globally UNIQUE so a cwd resolves
-- to at most one project (the guarantee SPEC-031's `here` resolver relies on).
CREATE TABLE project_locations (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL REFERENCES projects(id),
    path       TEXT NOT NULL UNIQUE
);

CREATE INDEX idx_project_locations_project ON project_locations(project_id);
