package billing

import (
	"context"
	"testing"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	"github.com/google/uuid"
)

type accessVerifierStub struct{}

func (accessVerifierStub) Verify(
	_ context.Context,
	request AccessRequest,
) (billingapi.Access, error) {
	principal := uuid.MustParse("00000000-0000-4000-8000-000000000001")
	switch request.Permission {
	case billingapi.PermissionBillingRead, billingapi.PermissionBillingManage,
		billingapi.PermissionBillingUsageRead:
		return billingapi.NewTenantAccess(
			principal,
			uuid.MustParse("00000000-0000-4000-8000-000000000002"),
			7,
			request.Permission,
		)
	default:
		return billingapi.NewGlobalAccess(principal, 7, request.Permission)
	}
}

type publicStub struct {
	billingapi.PublicQueries
	active func(context.Context) (*billingapi.CampaignView, error)
}

func (stub publicStub) GetActiveCampaign(ctx context.Context) (*billingapi.CampaignView, error) {
	return stub.active(ctx)
}

type catalogStub struct {
	billingapi.PlatformCatalog
	listOfferings func(
		context.Context,
		billingapi.Query,
	) (billingapi.QueryPage, error)
	createOfferingVersion func(
		context.Context,
		billingapi.CreateOfferingVersionCommand,
	) (billingapi.OfferingView, error)
}

func (stub catalogStub) ListOfferings(
	ctx context.Context,
	query billingapi.Query,
) (billingapi.QueryPage, error) {
	return stub.listOfferings(ctx, query)
}

func (stub catalogStub) CreateOfferingVersion(
	ctx context.Context,
	command billingapi.CreateOfferingVersionCommand,
) (billingapi.OfferingView, error) {
	return stub.createOfferingVersion(ctx, command)
}

type tenantStub struct{ billingapi.TenantOperations }
type reportsStub struct{ billingapi.PlatformReports }
type webhookStub struct{ billingapi.WebhookIngress }

func newHandlerFixture(
	t *testing.T,
	public billingapi.PublicQueries,
	catalog billingapi.PlatformCatalog,
) *Handler {
	t.Helper()
	if public == nil {
		public = publicStub{active: func(context.Context) (*billingapi.CampaignView, error) {
			return nil, nil
		}}
	}
	if catalog == nil {
		catalog = catalogStub{listOfferings: func(
			context.Context,
			billingapi.Query,
		) (billingapi.QueryPage, error) {
			return billingapi.QueryPage{}, nil
		}, createOfferingVersion: func(
			context.Context,
			billingapi.CreateOfferingVersionCommand,
		) (billingapi.OfferingView, error) {
			return billingapi.OfferingView{}, nil
		}}
	}
	handler, err := New(Dependencies{
		Access: accessVerifierStub{}, Public: public, Tenant: tenantStub{},
		Catalog: catalog, Reports: reportsStub{}, Webhook: webhookStub{},
		Report: func(error) {},
	})
	if err != nil {
		t.Fatal(err)
	}
	return handler
}
