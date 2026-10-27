package migrator

import (
	"context"
	"database/sql"
	"fmt"

	postgresstore "github.com/Onurlulardan/harpy/core/internal/platform/migrator/adapters/postgres"
)

func requireLifecycleReady(ctx context.Context, connection *sql.Conn) error {
	ready, err := postgresstore.MigrationRunInfrastructureReady(ctx, connection)
	if err != nil {
		return err
	}
	if !ready {
		return fmt.Errorf("migration run infrastructure is incomplete")
	}
	return nil
}
