package postgres

import (
	"context"
	_ "embed"
	"fmt"

	goosedb "github.com/pressly/goose/v3/database"
)

//go:embed sql/migration_run_infrastructure_ready.sql
var migrationRunInfrastructureReadySQL string

func MigrationRunInfrastructureReady(
	ctx context.Context,
	database goosedb.DBTxConn,
) (bool, error) {
	var ready bool
	if err := database.QueryRowContext(ctx, migrationRunInfrastructureReadySQL).
		Scan(&ready); err != nil {
		return false, fmt.Errorf("inspect migration run infrastructure: %w", err)
	}
	return ready, nil
}
