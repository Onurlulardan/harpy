package billingv1

import (
	"context"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	"github.com/google/uuid"
)

type deleteAccessStub struct{}

func (deleteAccessStub) Verify(
	_ context.Context, request billingapi.AccessRequest,
) (billingapi.Access, error) {
	if request.Permission != billingapi.PermissionPlatformBillingManage {
		return billingapi.Access{}, billingapi.ErrForbidden
	}
	return billingapi.NewGlobalAccess(uuid.New(), 1, request.Permission)
}

type deleteIdentityStub struct {
	identifier string
	kind       billingapi.LegacyIdentityKind
	canonical  string
}

func (stub deleteIdentityStub) ResolveLegacyIdentity(
	_ context.Context, query billingapi.LegacyIdentityQuery,
) (billingapi.LegacyIdentityView, error) {
	if query.LegacyID != stub.identifier || query.Kind != stub.kind {
		return billingapi.LegacyIdentityView{}, billingapi.ErrNotFound
	}
	return billingapi.LegacyIdentityView{
		LegacyID: stub.identifier, CanonicalID: stub.canonical,
	}, nil
}

func (stub deleteIdentityStub) ResolvePublicLegacyIdentity(
	context.Context, billingapi.LegacyIdentityKind, string,
) (billingapi.LegacyIdentityView, error) {
	return billingapi.LegacyIdentityView{}, billingapi.ErrNotFound
}

type deleteCatalogStub struct {
	billingapi.PlatformCatalog
	discount    billingapi.DiscountView
	discountErr error
	campaign    billingapi.CampaignView
	campaignErr error
}

func (stub deleteCatalogStub) GetDiscount(
	context.Context, billingapi.Access, string,
) (billingapi.DiscountView, error) {
	return stub.discount, stub.discountErr
}

func (stub deleteCatalogStub) GetCampaign(
	context.Context, billingapi.Access, string,
) (billingapi.CampaignView, error) {
	return stub.campaign, stub.campaignErr
}

func (stub deleteCatalogStub) CreateDiscount(
	context.Context, billingapi.CreateDiscountCommand,
) (billingapi.DiscountView, error) {
	return stub.discount, stub.discountErr
}
