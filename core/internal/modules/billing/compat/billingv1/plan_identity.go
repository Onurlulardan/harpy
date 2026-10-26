package billingv1

import (
	"context"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func (handler *Handler) resolvePlanIdentity(
	ctx context.Context,
	access billingapi.Access,
	rawID string,
) (billingapi.LegacyIdentityView, error) {
	identity, err := handler.identity.ResolveLegacyIdentity(
		ctx, billingapi.LegacyIdentityQuery{
			Access: access, Kind: billingapi.LegacyIdentityOffering, LegacyID: rawID,
		},
	)
	if err != nil {
		return billingapi.LegacyIdentityView{}, err
	}
	return validatePlanIdentity(identity, rawID)
}

func (handler *Handler) resolvePublicPlanIdentity(
	ctx context.Context,
	rawID string,
) (billingapi.LegacyIdentityView, error) {
	identity, err := handler.identity.ResolvePublicLegacyIdentity(
		ctx, billingapi.LegacyIdentityOffering, rawID,
	)
	if err != nil {
		return billingapi.LegacyIdentityView{}, err
	}
	return validatePlanIdentity(identity, rawID)
}

func validatePlanIdentity(
	identity billingapi.LegacyIdentityView,
	rawID string,
) (billingapi.LegacyIdentityView, error) {
	if identity.LegacyID == "" || identity.CanonicalID == "" ||
		identity.LegacyID != rawID && identity.CanonicalID != rawID {
		return billingapi.LegacyIdentityView{}, ErrSnapshotRequired
	}
	return identity, nil
}

func planWithLegacyID(
	offering billingapi.OfferingView,
	identity billingapi.LegacyIdentityView,
) (planWire, error) {
	if identity.CanonicalID != offering.ID || identity.LegacyID == "" {
		return planWire{}, ErrSnapshotRequired
	}
	offering.ID = identity.LegacyID
	return legacyPlan(offering), nil
}
