package billing

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func TestBillingHTTPRejectsOversizedBodiesBeforeApplicationPorts(t *testing.T) {
	tenant, catalogCalled := &tenantCommandStub{}, false
	catalog := limitCatalogStub(&catalogCalled)
	handler := newLimitHandler(t, tenant, catalog)
	t.Run("tenant command", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost,
			"/api/billing/v2/subscriptions/current/changes",
			strings.NewReader(strings.Repeat(" ", int(maximumCommandBodyBytes)+1)))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Idempotency-Key", "oversized-command")
		response := httptest.NewRecorder()
		handler.scheduleChange(response, request)
		if response.Code != http.StatusRequestEntityTooLarge || tenant.called {
			t.Fatalf("status=%d tenant_called=%t body=%s",
				response.Code, tenant.called, response.Body.String())
		}
	})
	t.Run("REST Query", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost,
			"/api/platform/billing/offerings/query",
			strings.NewReader(strings.Repeat(" ", int(maximumQueryBodyBytes)+1)))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		handler.queryOfferings(response, request)
		if response.Code != http.StatusRequestEntityTooLarge || catalogCalled {
			t.Fatalf("status=%d catalog_called=%t body=%s",
				response.Code, catalogCalled, response.Body.String())
		}
	})
}

func TestBillingHTTPRejectsPageAboveMaximumBeforeCatalogPort(t *testing.T) {
	called := false
	handler := newLimitHandler(t, tenantStub{}, limitCatalogStub(&called))
	requests := []*http.Request{
		httptest.NewRequest(http.MethodGet,
			"/api/platform/billing/offerings?page[size]=501", nil),
		httptest.NewRequest(http.MethodPost,
			"/api/platform/billing/offerings/query",
			strings.NewReader(`{"page":{"size":501,"after":null}}`)),
	}
	requests[1].Header.Set("Content-Type", "application/json")
	for _, request := range requests {
		response := httptest.NewRecorder()
		if request.Method == http.MethodGet {
			handler.listOfferings(response, request)
		} else {
			handler.queryOfferings(response, request)
		}
		if response.Code != http.StatusUnprocessableEntity || called {
			t.Fatalf("method=%s status=%d called=%t body=%s",
				request.Method, response.Code, called, response.Body.String())
		}
	}
}

func limitCatalogStub(called *bool) catalogStub {
	return catalogStub{listOfferings: func(
		context.Context, billingapi.Query,
	) (billingapi.QueryPage, error) {
		*called = true
		return billingapi.QueryPage{}, nil
	}, createOfferingVersion: func(
		context.Context, billingapi.CreateOfferingVersionCommand,
	) (billingapi.OfferingView, error) {
		return billingapi.OfferingView{}, nil
	}}
}

func newLimitHandler(
	t *testing.T,
	tenant billingapi.TenantOperations,
	catalog billingapi.PlatformCatalog,
) *Handler {
	t.Helper()
	handler, err := New(Dependencies{
		Access: accessVerifierStub{},
		Public: publicStub{active: func(context.Context) (*billingapi.CampaignView, error) {
			return nil, nil
		}},
		Tenant: tenant, Catalog: catalog, Reports: reportsStub{},
		Webhook: webhookStub{}, Report: func(error) {},
	})
	if err != nil {
		t.Fatal(err)
	}
	return handler
}
