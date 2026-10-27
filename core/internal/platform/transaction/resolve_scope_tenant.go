package transaction

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/google/uuid"
)

func resolveScopeTenant(
	ctx context.Context,
	database *postgres.RuntimeTransaction,
	tenantID uuid.UUID,
	mode ScopeMode,
) (postgres.TenantRecord, error) {
	if mode == SnapshotQuery {
		return database.ResolveTenantSnapshot(ctx, tenantID)
	}
	return database.ResolveTenant(ctx, tenantID)
}
