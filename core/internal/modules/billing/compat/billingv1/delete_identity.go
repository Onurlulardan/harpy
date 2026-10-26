package billingv1

import (
	"errors"
	"net/http"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func (handler *Handler) resolveDeleteIdentity(
	r *http.Request,
	access billingapi.Access,
	kind billingapi.LegacyIdentityKind,
	identifier string,
) (string, bool, error) {
	canonicalID, err := handler.canonicalPromotionID(
		r, access, kind, identifier,
	)
	if err == nil {
		return canonicalID, true, nil
	}
	if !errors.Is(err, billingapi.ErrNotFound) {
		return "", false, err
	}
	wrongKind, err := handler.existsAsOtherCatalogKind(
		r, access, kind, identifier,
	)
	if err != nil {
		return "", false, err
	}
	if wrongKind {
		return "", false, ErrSnapshotRequired
	}
	return "", false, nil
}

func (handler *Handler) existsAsOtherCatalogKind(
	r *http.Request,
	access billingapi.Access,
	expected billingapi.LegacyIdentityKind,
	identifier string,
) (bool, error) {
	for _, candidate := range catalogIdentityKinds() {
		if candidate == expected {
			continue
		}
		_, err := handler.canonicalLegacyID(
			r.Context(), access, candidate, identifier,
		)
		if err == nil {
			return true, nil
		}
		if !errors.Is(err, billingapi.ErrNotFound) {
			return false, err
		}
	}
	return false, nil
}

func catalogIdentityKinds() []billingapi.LegacyIdentityKind {
	return []billingapi.LegacyIdentityKind{
		billingapi.LegacyIdentityOffering,
		billingapi.LegacyIdentityDiscount,
		billingapi.LegacyIdentityCampaign,
	}
}
