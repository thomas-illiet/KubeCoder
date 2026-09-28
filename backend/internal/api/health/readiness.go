package health

import (
	"context"
	"database/sql"
	"errors"

	"github.com/thomas-illiet/KubeCoder/backend/internal/migration"
)

// SchemaReady returns a check that accepts only the current, clean migration version.
func SchemaReady(database interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}) func(context.Context) error {
	return func(ctx context.Context) error {
		var version uint
		var dirty bool
		err := database.QueryRowContext(ctx, "SELECT version, dirty FROM schema_migrations LIMIT 1").Scan(&version, &dirty)
		if err != nil {
			return err
		}
		if dirty || version != migration.LatestVersion {
			return errors.New("database schema is not ready")
		}
		return nil
	}
}
