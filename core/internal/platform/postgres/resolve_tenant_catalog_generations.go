package postgres

import (
	"context"
	_ "embed"

	"github.com/google/uuid"
)

//go:embed adapters/postgres/sql/resolve_tenant_catalog_generations.sql
var resolveTenantCatalogGenerationsSQL string

func (pool *Pool) ResolveTenantCatalogGenerations(
	ctx context.Context, tenantID uuid.UUID,
) (int64, int64, error) {
	if pool == nil || ctx == nil || ctx.Err() != nil || tenantID == uuid.Nil {
		return 0, 0, ErrAuthorityMismatch
	}
	database, err := pool.BeginRuntimeTransaction(ctx, TransactionSnapshotQuery)
	if err != nil {
		return 0, 0, err
	}
	var groupGeneration, catalogGeneration int64
	err = database.transaction.QueryRow(ctx, resolveTenantCatalogGenerationsSQL, tenantID).
		Scan(&groupGeneration, &catalogGeneration)
	if err != nil || groupGeneration < 1 || catalogGeneration < 1 {
		_ = database.Rollback()
		return 0, 0, ErrAuthorityMismatch
	}
	if err = database.Commit(ctx); err != nil {
		return 0, 0, err
	}
	return groupGeneration, catalogGeneration, nil
}
