package postgres

import (
	"context"

	"github.com/google/uuid"
)

func (pool *Pool) ResolveTenantWorkGenerations(
	ctx context.Context, tenantID uuid.UUID,
) (int64, int64, error) {
	if pool == nil || ctx == nil || ctx.Err() != nil || tenantID == uuid.Nil {
		return 0, 0, ErrAuthorityMismatch
	}
	database, err := pool.BeginRuntimeTransaction(ctx, TransactionSnapshotQuery)
	if err != nil {
		return 0, 0, err
	}
	var groupGeneration, workGeneration int64
	err = database.transaction.QueryRow(ctx, resolveTenantWorkGenerationsSQL,
		tenantID).Scan(&groupGeneration, &workGeneration)
	if err != nil || groupGeneration < 1 || workGeneration < 1 {
		_ = database.Rollback()
		return 0, 0, ErrAuthorityMismatch
	}
	if err = database.Commit(ctx); err != nil {
		return 0, 0, err
	}
	return groupGeneration, workGeneration, nil
}
