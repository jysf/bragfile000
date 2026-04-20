CREATE TABLE entries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    description TEXT,
    tags TEXT,
    project TEXT,
    type TEXT,
    impact TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE INDEX idx_entries_created_at ON entries(created_at DESC);
CREATE INDEX idx_entries_project ON entries(project);
