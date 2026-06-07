-- 0003_normalize_tags.sql — SPEC-025 (PROJ-002 / STAGE-006)
-- Single atomic in-place tag normalization (DEC-015, supersedes DEC-004).
-- Creates the polymorphic tags taxonomy, backfills it losslessly from the
-- comma-joined entries.tags column, re-points FTS sync onto the join, and
-- drops the legacy entries.tags column — all forward-only (DEC-002).
-- Runs inside the migration runner's per-migration transaction; do NOT
-- add BEGIN/COMMIT here. Validated at design against modernc.org/sqlite
-- 1.51.0 (SQLite 3.53.1): 3 tags / 7 taggings on the representative
-- corpus, search byte-stable, entries.tags dropped, FTS re-sync correct.

-- 1. Normalized taxonomy + polymorphic membership join.
CREATE TABLE tags (
    id   INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE taggings (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    tag_id        INTEGER NOT NULL REFERENCES tags(id),
    taggable_type TEXT    NOT NULL,
    taggable_id   INTEGER NOT NULL,
    position      INTEGER NOT NULL,
    UNIQUE (taggable_type, taggable_id, tag_id)
);

CREATE INDEX idx_taggings_tag      ON taggings(tag_id);
CREATE INDEX idx_taggings_taggable ON taggings(taggable_type, taggable_id);

-- 2. Lossless ETL from entries.tags: split on ',', trim each token, drop
--    empties, de-duplicate within an entry keeping first-occurrence
--    position. Upsert tag names, then write 'entry' memberships.
WITH RECURSIVE split(id, tok, rest, pos) AS (
    SELECT id, '', tags || ',', 0
      FROM entries
     WHERE tags IS NOT NULL AND tags <> ''
    UNION ALL
    SELECT id,
           substr(rest, 1, instr(rest, ',') - 1),
           substr(rest, instr(rest, ',') + 1),
           pos + 1
      FROM split
     WHERE rest <> ''
),
tokens AS (
    SELECT id AS entry_id, trim(tok) AS name, pos
      FROM split
     WHERE trim(tok) <> ''
)
INSERT OR IGNORE INTO tags(name)
SELECT DISTINCT name FROM tokens;

WITH RECURSIVE split(id, tok, rest, pos) AS (
    SELECT id, '', tags || ',', 0
      FROM entries
     WHERE tags IS NOT NULL AND tags <> ''
    UNION ALL
    SELECT id,
           substr(rest, 1, instr(rest, ',') - 1),
           substr(rest, instr(rest, ',') + 1),
           pos + 1
      FROM split
     WHERE rest <> ''
),
tokens AS (
    SELECT id AS entry_id, trim(tok) AS name, pos
      FROM split
     WHERE trim(tok) <> ''
),
firsts AS (
    SELECT entry_id, name, MIN(pos) AS position
      FROM tokens
     GROUP BY entry_id, name
)
INSERT OR IGNORE INTO taggings(tag_id, taggable_type, taggable_id, position)
SELECT t.id, 'entry', f.entry_id, f.position
  FROM firsts f
  JOIN tags t ON t.name = f.name;

-- 3. Re-point FTS. The old entries-row triggers copy new.tags/old.tags,
--    which is about to disappear; drop them, swap entries_fts from
--    external-content to a regular (own-content) FTS5 table, and backfill
--    its tags column from the join projection. Search shape (columns +
--    default unicode61 tokenizer) is unchanged, so DEC-010 holds.
DROP TRIGGER entries_ai;
DROP TRIGGER entries_au;
DROP TRIGGER entries_ad;
DROP TABLE entries_fts;

CREATE VIRTUAL TABLE entries_fts USING fts5(
    title, description, tags, project, impact
);

INSERT INTO entries_fts(rowid, title, description, tags, project, impact)
SELECT e.id, e.title, e.description,
       COALESCE((SELECT GROUP_CONCAT(t.name, ',' ORDER BY tg.position)
                   FROM taggings tg
                   JOIN tags t ON t.id = tg.tag_id
                  WHERE tg.taggable_type = 'entry' AND tg.taggable_id = e.id), ''),
       e.project, e.impact
  FROM entries e;

-- 4. Drop the legacy shadow column (SQLite 3.35+ ALTER ... DROP COLUMN;
--    safe now that no trigger or index references it).
ALTER TABLE entries DROP COLUMN tags;

-- 5. New trigger topology. entries triggers maintain the non-tag columns
--    (and seed an empty tags cell on insert); taggings/tags triggers
--    recompute the affected entry's tags projection.
CREATE TRIGGER entries_ai AFTER INSERT ON entries BEGIN
    INSERT INTO entries_fts(rowid, title, description, tags, project, impact)
    VALUES (new.id, new.title, new.description,
            COALESCE((SELECT GROUP_CONCAT(t.name, ',' ORDER BY tg.position)
                        FROM taggings tg JOIN tags t ON t.id = tg.tag_id
                       WHERE tg.taggable_type = 'entry' AND tg.taggable_id = new.id), ''),
            new.project, new.impact);
END;

CREATE TRIGGER entries_au AFTER UPDATE ON entries BEGIN
    UPDATE entries_fts
       SET title = new.title, description = new.description,
           project = new.project, impact = new.impact
     WHERE rowid = new.id;
END;

CREATE TRIGGER entries_ad AFTER DELETE ON entries BEGIN
    DELETE FROM entries_fts WHERE rowid = old.id;
END;

CREATE TRIGGER taggings_ai AFTER INSERT ON taggings
WHEN new.taggable_type = 'entry' BEGIN
    UPDATE entries_fts
       SET tags = COALESCE((SELECT GROUP_CONCAT(t.name, ',' ORDER BY tg.position)
                              FROM taggings tg JOIN tags t ON t.id = tg.tag_id
                             WHERE tg.taggable_type = 'entry'
                               AND tg.taggable_id = new.taggable_id), '')
     WHERE rowid = new.taggable_id;
END;

CREATE TRIGGER taggings_ad AFTER DELETE ON taggings
WHEN old.taggable_type = 'entry' BEGIN
    UPDATE entries_fts
       SET tags = COALESCE((SELECT GROUP_CONCAT(t.name, ',' ORDER BY tg.position)
                              FROM taggings tg JOIN tags t ON t.id = tg.tag_id
                             WHERE tg.taggable_type = 'entry'
                               AND tg.taggable_id = old.taggable_id), '')
     WHERE rowid = old.taggable_id;
END;

CREATE TRIGGER tags_au AFTER UPDATE ON tags BEGIN
    UPDATE entries_fts
       SET tags = COALESCE((SELECT GROUP_CONCAT(t.name, ',' ORDER BY tg.position)
                              FROM taggings tg JOIN tags t ON t.id = tg.tag_id
                             WHERE tg.taggable_type = 'entry'
                               AND tg.taggable_id = entries_fts.rowid), '')
     WHERE rowid IN (SELECT taggable_id FROM taggings
                      WHERE taggable_type = 'entry' AND tag_id = new.id);
END;
