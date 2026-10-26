package postgres

import (
	"context"

	"github.com/google/uuid"
)

func (transaction *RuntimeTransaction) VerifyTenantIdentity(
	ctx context.Context,
	tenantID uuid.UUID,
	record TenantRecord,
) error {
	if !transaction.usable() {
		return ErrTransactionClosed
	}
	var actualTenantID, schemaName, origin string
	err := transaction.transaction.QueryRow(ctx, verifyTenantIdentitySQL).Scan(
		&actualTenantID, &schemaName, &origin,
	)
	if err != nil || actualTenantID != tenantID.String() || schemaName != record.schemaName ||
		origin != "new" && origin != "v1_adopted" {
		return ErrTenantIdentityMismatch
	}
	return nil
}
