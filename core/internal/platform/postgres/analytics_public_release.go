package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (pool *Pool) ReleaseAnalyticsPublic(
	ctx context.Context, leaseID, tenantID, releaseID uuid.UUID, generation int64,
) error {
	if pool == nil || ctx == nil || ctx.Err() != nil || leaseID == uuid.Nil ||
		tenantID == uuid.Nil || releaseID == uuid.Nil || generation < 1 {
		return fmt.Errorf("analytics public release input is invalid")
	}
	database, err := pool.BeginRuntimeTransaction(ctx, TransactionCommand)
	if err != nil {
		return err
	}
	if _, err = database.transaction.Exec(ctx, releaseAnalyticsPublicSQL,
		leaseID, tenantID, releaseID, generation); err != nil {
		_ = database.Rollback()
		return err
	}
	return database.Commit(ctx)
}
