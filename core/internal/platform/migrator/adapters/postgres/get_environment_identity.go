package postgres

import (
	"context"
	_ "embed"
	"fmt"

	goosedb "github.com/pressly/goose/v3/database"
)

//go:embed sql/get_environment_identity.sql
var getEnvironmentIdentitySQL string

func GetEnvironmentIdentity(
	ctx context.Context,
	database goosedb.DBTxConn,
) (EnvironmentIdentityRecord, error) {
	var record EnvironmentIdentityRecord
	if err := database.QueryRowContext(ctx, getEnvironmentIdentitySQL).Scan(
		&record.EnvironmentID, &record.DatabaseOrigin, &record.DatabaseSystemID,
		&record.DatabaseOID, &record.BootstrapReleaseID,
		&record.BootstrapAuthoritySHA256,
	); err != nil {
		return EnvironmentIdentityRecord{}, fmt.Errorf("get environment identity: %w", err)
	}
	return record, nil
}
