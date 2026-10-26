package billingv1

import (
	"context"
	"net/http"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

type couponReferenceAccess struct {
	offering billingapi.Access
	campaign billingapi.Access
}

func boolPointer(value bool) *bool { return &value }

func (handler *Handler) couponReferences(
	access billingapi.Access,
) (couponReferenceAccess, error) {
	if !access.ValidFor(billingapi.PermissionPlatformBillingManage) ||
		access.Scope() != billingapi.AccessScopeGlobal {
		return couponReferenceAccess{}, billingapi.ErrForbidden
	}
	return couponReferenceAccess{offering: access, campaign: access}, nil
}

func (handler *Handler) resolveOfferingVersions(
	ctx context.Context,
	access billingapi.Access,
	legacyIDs []string,
) ([]string, error) {
	result := make([]string, len(legacyIDs))
	for index, id := range legacyIDs {
		canonicalID, err := handler.canonicalLegacyID(
			ctx, access, billingapi.LegacyIdentityOffering, id,
		)
		if err != nil {
			return nil, err
		}
		offering, err := handler.catalog.GetOffering(ctx, access, canonicalID)
		if err != nil {
			return nil, err
		}
		result[index] = offering.VersionID
	}
	return result, nil
}

func (handler *Handler) canonicalPromotionID(
	r *http.Request,
	access billingapi.Access,
	kind billingapi.LegacyIdentityKind,
	identifier string,
) (string, error) {
	return handler.canonicalLegacyID(r.Context(), access, kind, identifier)
}
