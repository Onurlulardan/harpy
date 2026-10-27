package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycle"
)

func (store *store) PasswordDeliveryTarget(ctx context.Context,
	scope credentialScope, query lifecycle.PasswordDeliveryTargetQuery,
) (lifecycle.DeliveryTarget, error) {
	if !credentialUsable(ctx, scope, credentialPasswordReset) ||
		(query.IdentifierKind != "email_v1" && query.IdentifierKind != "portal_v1") ||
		query.IdentifierSHA256 == [32]byte{} {
		return lifecycle.DeliveryTarget{}, lifecycle.ErrInvalid
	}
	bound, err := store.bindCredential(scope)
	if err != nil {
		return lifecycle.DeliveryTarget{}, err
	}
	return scanDeliveryTarget(bound.database.QueryRow(ctx, passwordDeliveryTargetSQL,
		query.IdentifierKind, query.IdentifierSHA256[:]))
}
