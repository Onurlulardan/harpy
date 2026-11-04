package goose

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/platform/foundation"
	"github.com/jackc/pgx/v5"
)

func requireGooseSchema(schema string) error {
	if schema == foundation.PlatformSchema || foundation.AllowedSchema(schema) && schema != "public" {
		return nil
	}
	return fmt.Errorf("goose schema %s is not a platform or tenant schema", schema)
}

func gooseSearchPath(schema string) string {
	return pgx.Identifier{schema}.Sanitize() + ", pg_catalog"
}

func applySearchPath(ctx context.Context, database *sql.DB, schema string) error {
	if err := requireGooseSchema(schema); err != nil {
		return err
	}
	_, err := database.ExecContext(
		ctx, "SELECT pg_catalog.set_config('search_path', $1, false)", gooseSearchPath(schema),
	)
	if err != nil {
		return fmt.Errorf("set goose search_path: %w", err)
	}
	return nil
}
