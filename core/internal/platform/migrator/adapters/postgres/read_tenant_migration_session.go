package postgres

import (
	"context"
	"database/sql"
	"fmt"
)

func ReadTenantMigrationSession(
	ctx context.Context,
	database *sql.DB,
) (TenantSessionProof, error) {
	var proof TenantSessionProof
	if err := database.QueryRowContext(ctx, readTenantMigrationSessionSQL).Scan(
		&proof.BackendPID, &proof.SearchPath, &proof.SchemaName,
	); err != nil {
		return TenantSessionProof{}, fmt.Errorf("read tenant migration session: %w", err)
	}
	if proof.BackendPID <= 0 || proof.SearchPath == "" || proof.SchemaName == "" {
		return TenantSessionProof{}, fmt.Errorf("tenant migration session is invalid")
	}
	return proof, nil
}
