package storage

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strings"
	"time"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// migrationStatus ensures schema_migrations exists, then partitions the
// embedded migrations into those already applied and those still pending
// (pending returned in lexical apply order). It applies nothing — it is the
// read-only half that both applyMigrations and the Open-time backup safety
// belt (backup.go) build on.
func migrationStatus(ctx context.Context, db *sql.DB, src fs.FS) (applied, pending []string, err error) {
	if _, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
        version TEXT PRIMARY KEY,
        applied_at TEXT NOT NULL
    )`); err != nil {
		return nil, nil, fmt.Errorf("create schema_migrations: %w", err)
	}

	appliedSet, err := loadApplied(ctx, db)
	if err != nil {
		return nil, nil, err
	}

	entries, err := fs.ReadDir(src, ".")
	if err != nil {
		return nil, nil, fmt.Errorf("read migrations dir: %w", err)
	}
	var files []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		files = append(files, e.Name())
	}
	sort.Strings(files)

	for _, name := range files {
		version := strings.TrimSuffix(name, ".sql")
		if _, ok := appliedSet[version]; ok {
			applied = append(applied, version)
		} else {
			pending = append(pending, version)
		}
	}
	return applied, pending, nil
}

// applyMigrations reads *.sql files from src, diffs them against the
// schema_migrations table, and applies each missing migration inside its
// own transaction (alongside the tracking INSERT) in lexical order.
func applyMigrations(ctx context.Context, db *sql.DB, src fs.FS) error {
	_, pending, err := migrationStatus(ctx, db, src)
	if err != nil {
		return err
	}
	for _, version := range pending {
		name := version + ".sql"
		sqlBytes, err := fs.ReadFile(src, name)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}
		if err := runMigration(ctx, db, version, string(sqlBytes)); err != nil {
			return fmt.Errorf("apply migration %s: %w", name, err)
		}
	}
	return nil
}

func loadApplied(ctx context.Context, db *sql.DB) (map[string]struct{}, error) {
	rows, err := db.QueryContext(ctx, "SELECT version FROM schema_migrations")
	if err != nil {
		return nil, fmt.Errorf("load applied migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[string]struct{})
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return nil, fmt.Errorf("scan applied migration: %w", err)
		}
		applied[v] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate applied migrations: %w", err)
	}
	return applied, nil
}

func runMigration(ctx context.Context, db *sql.DB, version, body string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	if _, err := tx.ExecContext(ctx, body); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("exec: %w", err)
	}
	if _, err := tx.ExecContext(ctx,
		"INSERT INTO schema_migrations (version, applied_at) VALUES (?, ?)",
		version, time.Now().UTC().Format(time.RFC3339),
	); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("record version: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}
