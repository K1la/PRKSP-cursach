package migration

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Migrator struct {
	db  *pgxpool.Pool
	dir string
}

func NewMigrator(db *pgxpool.Pool, dir string) Migrator {
	return Migrator{db: db, dir: dir}
}

func (m Migrator) Up(ctx context.Context) error {
	if _, err := m.db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`); err != nil {
		return err
	}

	entries, err := os.ReadDir(m.dir)
	if err != nil {
		return err
	}

	files := make([]string, 0)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".up.sql") {
			continue
		}
		files = append(files, entry.Name())
	}
	sort.Strings(files)

	for _, file := range files {
		applied, err := m.isApplied(ctx, file)
		if err != nil {
			return err
		}
		if applied {
			continue
		}
		if err := m.apply(ctx, file); err != nil {
			return err
		}
	}

	return nil
}

func (m Migrator) isApplied(ctx context.Context, version string) (bool, error) {
	var exists bool
	err := m.db.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = $1)`, version).Scan(&exists)
	return exists, err
}

func (m Migrator) apply(ctx context.Context, version string) error {
	sqlBytes, err := os.ReadFile(filepath.Join(m.dir, version))
	if err != nil {
		return err
	}

	tx, err := m.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, string(sqlBytes)); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, version); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
