-- External-content FTS5 index over entries. rowid maps to entries.id
-- via content_rowid='id'. Indexed fields are the human-readable text
-- columns: title, description, tags, project, impact. Numeric/
-- timestamp columns (id, created_at, updated_at) are not indexed.
-- Default unicode61 tokenizer splits on punctuation, which makes the
-- comma-joined tags column (DEC-004) tokenize one tag per token.
CREATE VIRTUAL TABLE entries_fts USING fts5(
    title, description, tags, project, impact,
    content='entries',
    content_rowid='id'
);

-- Backfill existing rows. On a fresh DB this SELECT returns zero rows
-- and the INSERT is a no-op; on an upgraded DB it populates the index
-- from whatever is already in `entries`. Runs inside the migration
-- runner's per-migration transaction (DEC-002).
INSERT INTO entries_fts(rowid, title, description, tags, project, impact)
SELECT id, title, description, tags, project, impact FROM entries;

-- Keep the index in sync on future writes. FTS5 uses
--   INSERT INTO entries_fts(entries_fts, rowid, ...) VALUES('delete', ...)
-- to remove a row from the index (SQLite FTS5 docs §4.4.3).
CREATE TRIGGER entries_ai AFTER INSERT ON entries BEGIN
    INSERT INTO entries_fts(rowid, title, description, tags, project, impact)
    VALUES (new.id, new.title, new.description, new.tags, new.project, new.impact);
END;

CREATE TRIGGER entries_ad AFTER DELETE ON entries BEGIN
    INSERT INTO entries_fts(entries_fts, rowid, title, description, tags, project, impact)
    VALUES ('delete', old.id, old.title, old.description, old.tags, old.project, old.impact);
END;

CREATE TRIGGER entries_au AFTER UPDATE ON entries BEGIN
    INSERT INTO entries_fts(entries_fts, rowid, title, description, tags, project, impact)
    VALUES ('delete', old.id, old.title, old.description, old.tags, old.project, old.impact);
    INSERT INTO entries_fts(rowid, title, description, tags, project, impact)
    VALUES (new.id, new.title, new.description, new.tags, new.project, new.impact);
END;
