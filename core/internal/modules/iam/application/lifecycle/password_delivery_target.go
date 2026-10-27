package lifecycle

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (service *Service) PasswordDeliveryTarget(
	ctx context.Context,
	scope platformtx.CredentialTenantScope,
	query PasswordDeliveryTargetQuery,
) (DeliveryTarget, error) {
	if !service.Valid() || !credentialLookupUsable(
		ctx,
		scope,
		platformtx.CredentialPasswordReset,
	) || (query.IdentifierKind != "email_v1" &&
		query.IdentifierKind != "portal_v1") ||
		query.IdentifierSHA256 == [32]byte{} {
		return DeliveryTarget{}, ErrInvalid
	}
	return service.validDeliveryTarget(
		service.store.PasswordDeliveryTarget(ctx, scope, query),
	)
}
