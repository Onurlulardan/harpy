package postgres

import "context"

func (transaction *RuntimeTransaction) BindActivationTenantSearchPath(
	ctx context.Context,
	schemaName string,
) error {
	return transaction.BindTenantSearchPath(ctx, TenantRecord{schemaName: schemaName})
}
