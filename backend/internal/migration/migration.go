package migration

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	iofs "github.com/golang-migrate/migrate/v4/source/iofs"
)

const LatestVersion uint = 20260928203000

//go:embed migrations/*.sql
var files embed.FS

type Status struct {
	Version uint
	Dirty   bool
}

// New creates a migrator backed by the embedded SQL migration files.
func New(db *sql.DB) (*migrate.Migrate, error) {
	sourceFS, err := fs.Sub(files, "migrations")
	if err != nil {
		return nil, fmt.Errorf("open embedded migrations: %w", err)
	}
	source, err := iofs.New(sourceFS, ".")
	if err != nil {
		return nil, fmt.Errorf("create migration source: %w", err)
	}
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("create postgres migration driver: %w", err)
	}
	instance, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
	if err != nil {
		return nil, fmt.Errorf("create migrator: %w", err)
	}
	return instance, nil
}

// Up applies every pending database migration.
func Up(db *sql.DB) error {
	m, err := New(db)
	if err != nil {
		return err
	}
	defer closeMigrator(m)
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate up: %w", err)
	}
	return nil
}

// Down reverts the requested number of database migrations.
func Down(db *sql.DB, steps int) error {
	m, err := New(db)
	if err != nil {
		return err
	}
	defer closeMigrator(m)
	if err := m.Steps(-steps); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate down: %w", err)
	}
	return nil
}

// Current returns the applied migration version and dirty state.
func Current(db *sql.DB) (Status, error) {
	var exists bool
	if err := db.QueryRow("SELECT to_regclass('schema_migrations') IS NOT NULL").Scan(&exists); err != nil {
		return Status{}, fmt.Errorf("find migration table: %w", err)
	}
	if !exists {
		return Status{}, nil
	}
	var status Status
	if err := db.QueryRow("SELECT version, dirty FROM schema_migrations LIMIT 1").Scan(&status.Version, &status.Dirty); errors.Is(err, sql.ErrNoRows) {
		return Status{}, nil
	} else if err != nil {
		return Status{}, fmt.Errorf("read migration version: %w", err)
	}
	return status, nil
}

// Wait blocks until the expected clean migration version is available.
func Wait(ctx context.Context, db *sql.DB, expected uint, interval time.Duration) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		status, err := Current(db)
		if err == nil {
			if status.Dirty {
				return fmt.Errorf("database migration %d is dirty", status.Version)
			}
			if status.Version == expected {
				return nil
			}
			if status.Version > expected {
				return fmt.Errorf("database schema %d is newer than expected %d", status.Version, expected)
			}
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("wait for database schema %d: %w", expected, ctx.Err())
		case <-ticker.C:
		}
	}
}

// closeMigrator releases migration source and database driver resources.
func closeMigrator(m *migrate.Migrate) {
	_, _ = m.Close()
}
