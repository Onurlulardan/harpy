package transaction

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

func prepareCredentialScope(
	ctx context.Context,
	database *postgres.RuntimeTransaction,
	target VerifiedCredentialTarget,
) (CredentialTenantScope, error) {
	record, err := database.ResolveTenant(ctx, target.tenantID)
	if err != nil {
		return CredentialTenantScope{}, mapPreparationError(err)
	}
	if err := database.BindTenantSearchPath(ctx, record); err != nil {
		return CredentialTenantScope{}, mapPreparationError(err)
	}
	if err := database.VerifyTenantIdentity(ctx, target.tenantID, record); err != nil {
		return CredentialTenantScope{}, mapPreparationError(err)
	}
	if err := lockOperationAuthority(ctx, database, target.tenantID, target.operation); err != nil {
		return CredentialTenantScope{}, err
	}
	state := &credentialScopeState{target: target, transaction: database}
	state.active.Store(true)
	return CredentialTenantScope{state: state}, nil
}
