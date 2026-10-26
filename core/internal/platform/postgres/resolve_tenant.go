package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (transaction *RuntimeTransaction) ResolveTenant(
	ctx context.Context,
	tenantID uuid.UUID,
) (TenantRecord, error) {
	if !transaction.usable() || tenantID == uuid.Nil {
		return TenantRecord{}, ErrTransactionClosed
	}
	var schemaName, slug, status string
	var notDeleted bool
	err := transaction.transaction.QueryRow(ctx, resolveTenantSQL, tenantID).Scan(
		&schemaName, &slug, &status, &notDeleted,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return TenantRecord{}, ErrTenantMissing
	}
	if err != nil {
		return TenantRecord{}, err
	}
	if !notDeleted || status != "ACTIVE" && status != "TRIAL" {
		return TenantRecord{}, ErrTenantInactive
	}
	if _, err := tenantSearchPath(schemaName); err != nil {
		return TenantRecord{}, err
	}
	return TenantRecord{schemaName: schemaName, slug: slug}, nil
}
