package postgres

import "context"

func (pool *Pool) ResolveGlobalBillingGenerations(
	ctx context.Context,
) (int64, int64, error) {
	if pool == nil || ctx == nil || ctx.Err() != nil {
		return 0, 0, ErrAuthorityMismatch
	}
	database, err := pool.BeginRuntimeTransaction(ctx, TransactionSnapshotQuery)
	if err != nil {
		return 0, 0, err
	}
	var groupGeneration, billingGeneration int64
	err = database.transaction.QueryRow(
		ctx, resolveGlobalBillingGenerationsSQL,
	).Scan(&groupGeneration, &billingGeneration)
	if err != nil || groupGeneration < 1 || billingGeneration < 1 {
		_ = database.Rollback()
		return 0, 0, ErrAuthorityMismatch
	}
	if err = database.Commit(ctx); err != nil {
		return 0, 0, err
	}
	return groupGeneration, billingGeneration, nil
}
