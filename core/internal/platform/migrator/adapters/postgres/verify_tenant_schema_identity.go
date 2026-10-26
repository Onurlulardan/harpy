package postgres

import (
	"context"
	"database/sql"
	"fmt"
)

func VerifyTenantSchemaIdentity(
	ctx context.Context,
	transaction *sql.Tx,
	target TenantMigrationTarget,
	origin string,
) error {
	var exact bool
	if err := transaction.QueryRowContext(
		ctx, verifyTenantSchemaIdentitySQL,
		target.TenantID, target.SchemaName, origin,
	).Scan(&exact); err != nil {
		return fmt.Errorf("verify tenant schema identity: %w", err)
	}
	if !exact {
		return fmt.Errorf("tenant schema identity is not exact")
	}
	return nil
}
