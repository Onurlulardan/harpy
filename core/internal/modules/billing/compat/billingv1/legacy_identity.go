package billingv1

import (
	"context"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func (handler *Handler) canonicalLegacyID(
	ctx context.Context,
	access billingapi.Access,
	kind billingapi.LegacyIdentityKind,
	identifier string,
) (string, error) {
	identity, err := handler.identity.ResolveLegacyIdentity(
		ctx, billingapi.LegacyIdentityQuery{
			Access: access, Kind: kind, LegacyID: identifier,
		},
	)
	if err != nil {
		return "", err
	}
	return identity.CanonicalID, nil
}
