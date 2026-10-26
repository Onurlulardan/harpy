package transaction

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

func prepareProviderScope(
	ctx context.Context,
	database *postgres.RuntimeTransaction,
	lookup ProviderLookup,
	target verifiedProviderTarget,
) (ProviderTenantScope, error) {
	record, err := database.ResolveTenant(ctx, target.tenantID)
	if err != nil {
		return ProviderTenantScope{}, mapPreparationError(err)
	}
	if err := database.BindTenantSearchPath(ctx, record); err != nil {
		return ProviderTenantScope{}, mapPreparationError(err)
	}
	if err := database.VerifyTenantIdentity(ctx, target.tenantID, record); err != nil {
		return ProviderTenantScope{}, mapPreparationError(err)
	}
	if err := lockOperationAuthority(ctx, database, target.tenantID, target.operation); err != nil {
		return ProviderTenantScope{}, err
	}
	state := &providerScopeState{lookup: lookup, target: target, transaction: database}
	state.active.Store(true)
	return ProviderTenantScope{state: state}, nil
}
