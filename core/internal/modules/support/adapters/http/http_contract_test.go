package supporthttp

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRoutesExposeExactLegacyCompatibleSupportSurface(t *testing.T) {
	handler, _, _ := handlerFixture()
	routes := handler.Routes()
	if len(routes) != 17 {
		t.Fatalf("route count = %d, want 17", len(routes))
	}
	want := map[string]struct{}{
		"POST /api/contact": {}, "GET /api/public/booking-url": {},
		"GET /api/admin/contact-messages": {}, "GET /api/admin/contact-messages/stats": {},
		"GET /api/admin/contact-messages/{id}": {}, "PUT /api/admin/contact-messages/{id}": {},
		"POST /api/bug-reports": {}, "GET /api/bug-reports/mine": {},
		"GET /api/bug-reports/{id}": {}, "GET /api/bug-reports/admin": {},
		"GET /api/bug-reports/admin/stats": {}, "GET /api/bug-reports/admin/{id}": {},
		"PUT /api/bug-reports/admin/{id}": {}, "GET /api/admin/bug-reports": {},
		"GET /api/admin/bug-reports/stats": {}, "GET /api/admin/bug-reports/{id}": {},
		"PUT /api/admin/bug-reports/{id}": {},
	}
	for _, route := range routes {
		key := route.Method + " " + route.Pattern
		if _, found := want[key]; !found {
			t.Fatalf("unexpected route %s", key)
		}
		delete(want, key)
	}
	if len(want) != 0 {
		t.Fatalf("missing routes: %v", want)
	}
}

func TestPublicContactRequiresStrictBodyAndIdempotencyEvidence(t *testing.T) {
	handler, application, _ := handlerFixture()
	route := handler.Routes()[0]
	request := httptest.NewRequest(http.MethodPost, "/api/contact", strings.NewReader(
		`{"name":"Ada","email":"ada@example.com","message":"Need help","contact_type":"general","unknown":true}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", "contact-command-0001")
	response := httptest.NewRecorder()
	route.Handler.ServeHTTP(response, request)
	if response.Code != http.StatusUnprocessableEntity || application.contactCalls != 0 {
		t.Fatalf("unknown body status=%d calls=%d", response.Code, application.contactCalls)
	}

	request = httptest.NewRequest(http.MethodPost, "/api/contact", strings.NewReader(
		`{"name":"Ada","email":"ada@example.com","message":"Need help","contact_type":"general"}`))
	request.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	route.Handler.ServeHTTP(response, request)
	if response.Code != http.StatusUnprocessableEntity || application.contactCalls != 0 {
		t.Fatalf("missing key status=%d calls=%d", response.Code, application.contactCalls)
	}
}

func TestPublicContactSuccessIsNoStoreAndDoesNotLeakAdmissionData(t *testing.T) {
	handler, application, _ := handlerFixture()
	request := httptest.NewRequest(http.MethodPost, "/api/contact", strings.NewReader(
		`{"name":"Ada","email":"ada@example.com","message":"Need help","contact_type":"general"}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", "contact-command-0001")
	response := httptest.NewRecorder()
	handler.Routes()[0].Handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK || application.contactCalls != 1 ||
		response.Header().Get("Cache-Control") != "no-store" ||
		strings.Contains(response.Body.String(), "rate") {
		t.Fatalf("contact response status=%d calls=%d body=%s",
			response.Code, application.contactCalls, response.Body.String())
	}
}

func TestBookingConstructorRejectsMutableOrUnsafeConfiguration(t *testing.T) {
	app := &applicationFixture{}
	apps := Applications{Contact: app, Reporter: app, Tenant: app, Platform: app}
	_, err := NewHandler(Dependencies{Access: &accessFixture{}, Applications: apps,
		RateKeys: rateFixture{}, Booking: BookingConfig{URL: "http://example.com", Provider: "demo"},
		Report: func(error) {}, Correlation: func(*http.Request) string { return "correlation" }})
	if err == nil {
		t.Fatal("unsafe booking URL was accepted")
	}
}
