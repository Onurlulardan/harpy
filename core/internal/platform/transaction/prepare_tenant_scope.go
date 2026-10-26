package transaction

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/requestattribution"
)

func prepareTenantScope(
	ctx context.Context,
	transaction *postgres.RuntimeTransaction,
	request ScopeRequest,
) (TenantScope, error) {
	tenantID := request.Actor.tenantID
	record, err := resolveScopeTenant(ctx, transaction, tenantID, request.Mode)
	if err != nil {
		return TenantScope{}, mapPreparationError(err)
	}
	if err := transaction.BindTenantSearchPath(ctx, record); err != nil {
		return TenantScope{}, mapPreparationError(err)
	}
	if err := transaction.VerifyTenantIdentity(ctx, tenantID, record); err != nil {
		return TenantScope{}, mapPreparationError(err)
	}
	operation := request.Operation
	if err := lockOperationAuthority(ctx, transaction, tenantID, operation); err != nil {
		return TenantScope{}, err
	}
	_ = requestattribution.Bind(ctx, request.Actor.actorID, tenantID)
	state := &scopeState{
		actorID: request.Actor.actorID, tenantID: tenantID, tenantSlug: record.Slug(),
		operation: operation, mode: request.Mode, transaction: transaction,
	}
	state.active.Store(true)
	return TenantScope{state: state}, nil
}
