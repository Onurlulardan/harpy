package postgres

import (
	"context"
	"database/sql"
	"fmt"
)

func InsertTenantSchemaIdentity(
	ctx context.Context,
	transaction *sql.Tx,
	target TenantMigrationTarget,
	origin string,
) error {
	if _, err := transaction.ExecContext(
		ctx, insertTenantSchemaIdentitySQL,
		target.TenantID, target.SchemaName, origin,
	); err != nil {
		return fmt.Errorf("insert tenant schema identity: %w", err)
	}
	return nil
}
