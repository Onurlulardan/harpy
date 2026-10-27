package audithttp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/audit/application"
)

type permissionVerifier struct{ requests []AccessRequest }

func (verifier *permissionVerifier) Verify(
	_ context.Context, request AccessRequest,
) (Access, error) {
	verifier.requests = append(verifier.requests, request)
	return NewAccess(httpActor, httpTenant, 2, 3, 4,
		request.Permission, request.Operation)
}

func TestRoutesExposeOnlyFiveAuthenticatedReadSurfaces(t *testing.T) {
	handler := fixtureHandler(t, &applicationFixture{entry: fixtureEntry("public")})
	routes := handler.Routes()
	want := []string{
		"GET /api/changelog",
		"GET /api/changelog/stats",
		"GET /api/changelog/{id}",
		"GET /api/changelog/history/{entity_type}/{entity_id}",
		"GET /api/changelog/forms/{form_slug}/records/{record_id}/history",
	}
	if len(routes) != len(want) {
		t.Fatalf("route count = %d", len(routes))
	}
	for index, route := range routes {
		if got := route.Method + " " + route.Pattern; got != want[index] {
			t.Fatalf("route[%d] = %q", index, got)
		}
		if strings.Contains(route.Pattern, "ingest") {
			t.Fatal("arbitrary audit ingestion was exposed over HTTP")
		}
	}
}

func TestEachRouteRequestsExactReadPermissionAndOperation(t *testing.T) {
	verifier := &permissionVerifier{}
	handler, err := NewHandler(Dependencies{Access: verifier,
		Application: &applicationFixture{entry: fixtureEntry("public")},
		Report:      func(error) {}, Correlation: func(*http.Request) string { return "corr-safe" }})
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	for _, route := range handler.Routes() {
		mux.Handle(route.Method+" "+route.Pattern, route.Handler)
	}
	targets := []struct{ path, operation string }{
		{"/api/changelog", application.OperationList},
		{"/api/changelog/stats", application.OperationStats},
		{"/api/changelog/019c0000-0000-7000-8000-000000000203", application.OperationGet},
		{"/api/changelog/history/form_record/record-1", application.OperationEntityHistory},
		{"/api/changelog/forms/orders/records/record-1/history",
			application.OperationFormRecordHistory},
	}
	for _, target := range targets {
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, target.path, nil))
		if response.Code != http.StatusOK {
			t.Fatalf("%s status=%d body=%s", target.path, response.Code, response.Body.String())
		}
	}
	if len(verifier.requests) != len(targets) {
		t.Fatalf("permission requests=%d", len(verifier.requests))
	}
	for index, request := range verifier.requests {
		if request.Permission != application.PermissionRead ||
			request.Operation != targets[index].operation {
			t.Fatalf("request[%d]=%s/%s", index, request.Permission, request.Operation)
		}
	}
}

func TestListRejectsDynamicSortAndUnknownQueryBeforeStore(t *testing.T) {
	app := &applicationFixture{entry: fixtureEntry("public")}
	handler := fixtureHandler(t, app)
	for _, target := range []string{
		"/api/changelog?sort_direction=ASC",
		"/api/changelog?sort_by=entity_type",
		"/api/changelog?limit=101",
		"/api/changelog?unknown=value",
	} {
		response := httptest.NewRecorder()
		handler.list(response, httptest.NewRequest(http.MethodGet, target, nil))
		if response.Code != http.StatusUnprocessableEntity {
			t.Fatalf("%s status = %d", target, response.Code)
		}
	}
	if app.lists != 0 {
		t.Fatalf("invalid query reached application %d times", app.lists)
	}
}

func TestSecretHistoryNeverReturnsValuesOrDigests(t *testing.T) {
	app := &applicationFixture{entry: fixtureEntry("secret")}
	handler := fixtureHandler(t, app)
	request := httptest.NewRequest(http.MethodGet,
		"/api/changelog/forms/orders/records/record-1/history?page=1&limit=50", nil)
	request.SetPathValue("form_slug", "orders")
	request.SetPathValue("record_id", "record-1")
	response := httptest.NewRecorder()
	handler.formRecordHistory(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	body := response.Body.String()
	for _, forbidden := range []string{"old", "new", "before_sha256", "after_sha256"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("secret audit response leaked %q: %s", forbidden, body)
		}
	}
	var document map[string]any
	if json.Unmarshal(response.Body.Bytes(), &document) != nil || document["success"] != true {
		t.Fatal("response was not a valid success envelope")
	}
}
