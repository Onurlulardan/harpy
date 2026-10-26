package billing

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func TestManualOnlyWebhookIngressIsPermanentlyClosed(t *testing.T) {
	body := []byte(`{"id":"evt-1"}`)
	result, err := NewManualOnlyWebhookHTTPIngress().ReceiveWebhook(
		context.Background(),
		billingapi.WebhookRequest{
			Provider: "manual",
			Headers:  map[string][]string{"X-Test": {"value"}},
			Body:     body,
		},
	)
	if !errors.Is(err, billingapi.ErrNotFound) {
		t.Fatalf("manual-only webhook error = %v, want not found", err)
	}
	if result != (billingapi.WebhookAcceptance{}) {
		t.Fatalf("manual-only webhook result = %#v", result)
	}
	if string(body) != `{"id":"evt-1"}` {
		t.Fatal("manual-only ingress unexpectedly retained or mutated request body")
	}
}

func TestManualOnlyBillingWebhookRouteNeverAcknowledgesProviderWork(t *testing.T) {
	handler := newHandlerFixture(t, nil, nil)
	handler.webhook = NewManualOnlyWebhookHTTPIngress()
	router := http.NewServeMux()
	route := handler.WebhookRoutes()[0]
	router.Handle(route.Method+" "+route.Pattern, route.Handler)

	body := &trackingWebhookBody{}
	request := httptest.NewRequest(
		http.MethodPost,
		"/webhooks/billing/manual",
		body,
	)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("manual-only webhook status = %d, want 404", response.Code)
	}
	if body.read {
		t.Fatal("manual-only webhook route read a body it cannot accept")
	}
}

type trackingWebhookBody struct{ read bool }

func (body *trackingWebhookBody) Read([]byte) (int, error) {
	body.read = true
	return 0, errors.New("body must not be read")
}
