package postgres

import (
	"context"

	"github.com/google/uuid"
)

func (pool *Pool) ResolveTenantWorkflowGenerations(
	ctx context.Context,
	tenantID uuid.UUID,
) (int64, int64, error) {
	if pool == nil || ctx == nil || ctx.Err() != nil || tenantID == uuid.Nil {
		return 0, 0, ErrAuthorityMismatch
	}
	database, err := pool.BeginRuntimeTransaction(ctx, TransactionSnapshotQuery)
	if err != nil {
		return 0, 0, err
	}
	var groupGeneration, workflowGeneration int64
	err = database.transaction.QueryRow(
		ctx, resolveTenantWorkflowGenerationsSQL, tenantID,
	).Scan(&groupGeneration, &workflowGeneration)
	if err != nil || groupGeneration < 1 || workflowGeneration < 1 {
		_ = database.Rollback()
		return 0, 0, ErrAuthorityMismatch
	}
	if err = database.Commit(ctx); err != nil {
		return 0, 0, err
	}
	return groupGeneration, workflowGeneration, nil
}
