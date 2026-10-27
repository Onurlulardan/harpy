package postgres

import (
	"context"
	_ "embed"
	"fmt"

	goosedb "github.com/pressly/goose/v3/database"
)

//go:embed sql/create_environment_identity.sql
var createEnvironmentIdentitySQL string

func CreateEnvironmentIdentity(
	ctx context.Context,
	database goosedb.DBTxConn,
	record EnvironmentIdentityRecord,
) error {
	result, err := database.ExecContext(
		ctx, createEnvironmentIdentitySQL,
		record.EnvironmentID, record.DatabaseOrigin, record.DatabaseSystemID,
		record.DatabaseOID, record.BootstrapReleaseID,
		record.BootstrapAuthoritySHA256,
	)
	if err != nil {
		return fmt.Errorf("create environment identity: %w", err)
	}
	if affected, err := result.RowsAffected(); err != nil || affected > 1 {
		return fmt.Errorf("create environment identity returned an invalid row count")
	}
	return nil
}
