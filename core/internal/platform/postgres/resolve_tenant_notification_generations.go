package postgres

import (
	"context"
	_ "embed"

	"github.com/google/uuid"
)

//go:embed adapters/postgres/sql/resolve_tenant_notification_generations.sql
var resolveTenantNotificationGenerationsSQL string

func (pool *Pool) ResolveTenantNotificationGenerations(
	ctx context.Context, tenantID uuid.UUID,
) (int64, int64, error) {
	if pool == nil || ctx == nil || ctx.Err() != nil || tenantID == uuid.Nil {
		return 0, 0, ErrAuthorityMismatch
	}
	database, err := pool.BeginRuntimeTransaction(ctx, TransactionSnapshotQuery)
	if err != nil {
		return 0, 0, err
	}
	var groupGeneration, notificationGeneration int64
	err = database.transaction.QueryRow(ctx, resolveTenantNotificationGenerationsSQL,
		tenantID).Scan(&groupGeneration, &notificationGeneration)
	if err != nil || groupGeneration < 1 || notificationGeneration < 1 {
		_ = database.Rollback()
		return 0, 0, ErrAuthorityMismatch
	}
	if err = database.Commit(ctx); err != nil {
		return 0, 0, err
	}
	return groupGeneration, notificationGeneration, nil
}
