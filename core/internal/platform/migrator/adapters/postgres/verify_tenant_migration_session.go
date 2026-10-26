package postgres

import (
	"context"
	"fmt"

	goosedb "github.com/pressly/goose/v3/database"
)

func VerifyTenantMigrationSession(
	ctx context.Context,
	database goosedb.DBTxConn,
	proof TenantSessionProof,
	target TenantMigrationTarget,
) error {
	var backend, searchPath, schema, lock bool
	if err := database.QueryRowContext(
		ctx, verifyTenantMigrationSessionSQL,
		proof.BackendPID, proof.SearchPath, proof.SchemaName, target.TenantID,
	).Scan(&backend, &searchPath, &schema, &lock); err != nil {
		return fmt.Errorf("verify tenant migration session: %w", err)
	}
	if !backend || !searchPath || !schema || !lock || proof.SchemaName != target.SchemaName {
		return fmt.Errorf("tenant migration session continuity failed")
	}
	return nil
}
