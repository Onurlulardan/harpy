package billing

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

type tenantCommandStub struct {
	billingapi.TenantOperations
	called bool
}

func (stub *tenantCommandStub) ScheduleChange(
	context.Context,
	billingapi.ScheduleChangeCommand,
) (billingapi.SubscriptionView, error) {
	stub.called = true
	return billingapi.SubscriptionView{}, nil
}

func TestTenantCommandRejectsBodyTenantAuthorityBeforeApplication(t *testing.T) {
	tenant := &tenantCommandStub{}
	handler, err := New(Dependencies{
		Access: accessVerifierStub{}, Public: publicStub{active: func(
			context.Context,
		) (*billingapi.CampaignView, error) {
			return nil, nil
		}},
		Tenant: tenant,
		Catalog: catalogStub{createOfferingVersion: func(
			context.Context,
			billingapi.CreateOfferingVersionCommand,
		) (billingapi.OfferingView, error) {
			return billingapi.OfferingView{}, nil
		}},
		Reports: reportsStub{}, Webhook: webhookStub{}, Report: func(error) {},
	})
	if err != nil {
		t.Fatal(err)
	}
	document := `{
      "target_offering_version_id":"00000000-0000-4000-8000-000000000010",
      "target_price_version_id":"00000000-0000-4000-8000-000000000011",
      "expected_version":2,
      "tenant_id":"00000000-0000-4000-8000-000000000099"
    }`
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/billing/v2/subscriptions/current/changes",
		strings.NewReader(document),
	)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", "change-plan-tenant-authority")
	response := httptest.NewRecorder()
	handler.scheduleChange(response, request)
	if response.Code != http.StatusBadRequest || tenant.called {
		t.Fatalf("status=%d called=%t body=%s", response.Code, tenant.called, response.Body.String())
	}
}
