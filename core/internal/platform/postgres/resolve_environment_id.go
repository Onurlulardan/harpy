package postgres

import (
	"context"

	"github.com/google/uuid"
)

func (pool *Pool) ResolveEnvironmentID(ctx context.Context) (uuid.UUID, error) {
	if pool == nil || ctx == nil || ctx.Err() != nil {
		return uuid.Nil, ErrAuthorityMismatch
	}
	database, err := pool.BeginRuntimeTransaction(ctx, TransactionSnapshotQuery)
	if err != nil {
		return uuid.Nil, err
	}
	var environmentID uuid.UUID
	err = database.transaction.QueryRow(ctx, resolveEnvironmentIDSQL).Scan(&environmentID)
	if err != nil || environmentID == uuid.Nil {
		_ = database.Rollback()
		return uuid.Nil, ErrAuthorityMismatch
	}
	if err := database.Commit(ctx); err != nil {
		return uuid.Nil, err
	}
	return environmentID, nil
}
