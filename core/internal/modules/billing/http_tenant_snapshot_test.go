package billing

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

type currentSubscriptionTenantStub struct {
	billingapi.TenantOperations
	view billingapi.SubscriptionView
}

func (stub currentSubscriptionTenantStub) GetCurrentSubscription(
	context.Context,
	billingapi.SubscriptionAccess,
) (billingapi.SubscriptionView, error) {
	return stub.view, nil
}

func TestTenantHTTPReturnsHistoricalPrivateOfferingSnapshot(t *testing.T) {
	snapshot := currentSubscription(t, testV7(t)).View
	handler, err := New(Dependencies{
		Access: accessVerifierStub{},
		Public: publicStub{active: func(context.Context) (*billingapi.CampaignView, error) {
			return nil, nil
		}},
		Tenant: currentSubscriptionTenantStub{view: snapshot},
		Catalog: catalogStub{createOfferingVersion: func(
			context.Context, billingapi.CreateOfferingVersionCommand,
		) (billingapi.OfferingView, error) {
			return billingapi.OfferingView{}, nil
		}},
		Reports: reportsStub{}, Webhook: webhookStub{}, Report: func(error) {},
	})
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodGet, "/api/billing/v2/subscriptions/current", nil)
	response := httptest.NewRecorder()
	handler.getCurrentSubscription(response, request)
	var envelope struct {
		Success bool                        `json:"success"`
		Data    billingapi.SubscriptionView `json:"data"`
	}
	if decodeErr := json.Unmarshal(response.Body.Bytes(), &envelope); decodeErr != nil {
		t.Fatal(decodeErr)
	}
	if response.Code != http.StatusOK || !envelope.Success ||
		envelope.Data.Offering.State != "retired" || envelope.Data.Offering.Public ||
		envelope.Data.Offering.VersionID != envelope.Data.OfferingVersionID {
		t.Fatalf("status=%d subscription=%+v", response.Code, envelope.Data)
	}
}
