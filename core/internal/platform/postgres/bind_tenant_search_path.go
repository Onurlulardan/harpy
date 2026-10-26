package postgres

import (
	"context"
	"slices"
)

func (transaction *RuntimeTransaction) BindTenantSearchPath(
	ctx context.Context,
	record TenantRecord,
) error {
	if !transaction.usable() {
		return ErrTransactionClosed
	}
	searchPath, err := tenantSearchPath(record.schemaName)
	if err != nil {
		return err
	}
	var applied string
	if err := transaction.transaction.QueryRow(
		ctx, setTenantSearchPathSQL, searchPath,
	).Scan(&applied); err != nil {
		return err
	}
	if applied != searchPath {
		return ErrTenantSchemaInvalid
	}
	var current []string
	if err := transaction.transaction.QueryRow(ctx, verifyTenantSearchPathSQL).Scan(&current); err != nil {
		return err
	}
	if !slices.Equal(current, []string{"pg_catalog", record.schemaName}) {
		return ErrTenantSchemaInvalid
	}
	return nil
}
