package tenancy

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	lifecycle "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/customerlifecycle"
)

func TestCustomerLifecycleHTTPRoutesCoverCanonicalAndLegacySurface(t *testing.T) {
	handler, _ := newLifecycleHTTPFixture(t)
	routes := handler.Routes()
	if len(routes) != 16 {
		t.Fatalf("routes=%d want=16", len(routes))
	}
	want := map[string]bool{
		"GET /api/platform/tenants":                    false,
		"POST /api/platform/tenants/{tenant_id}/purge": false,
		"GET /api/admin/tenants":                       false,
		"DELETE /api/admin/tenants/{tenant_id}":        false,
		"POST /api/admin/switch-tenant":                false,
	}
	for _, route := range routes {
		identity := route.Method + " " + route.Pattern
		if _, found := want[identity]; found {
			want[identity] = true
		}
	}
	for identity, found := range want {
		if !found {
			t.Fatalf("missing route %s", identity)
		}
	}
}

func TestCustomerLifecycleHTTPListUsesBoundedCursorAndLegacyEnvelope(t *testing.T) {
	handler, fixture := newLifecycleHTTPFixture(t)
	request := httptest.NewRequest(http.MethodGet,
		"/api/admin/tenants?limit=25&after="+httpLifecycleTenantID.String(), nil)
	response := httptest.NewRecorder()
	handler.list(response, request)
	if response.Code != http.StatusOK || fixture.listQuery.Limit != 25 ||
		fixture.listQuery.After != httpLifecycleTenantID {
		t.Fatalf("status=%d query=%+v body=%s", response.Code, fixture.listQuery, response.Body.String())
	}
	var document map[string]any
	if json.Unmarshal(response.Body.Bytes(), &document) != nil || document["success"] != nil ||
		document["data"] == nil || document["total"].(float64) != 0 {
		t.Fatalf("legacy list body=%s", response.Body.String())
	}
}

func TestCustomerLifecycleHTTPUpdateRequiresRevisionAndPinsCommand(t *testing.T) {
	handler, fixture := newLifecycleHTTPFixture(t)
	request := lifecycleHTTPRequest(http.MethodPut, "/api/platform/tenants/"+
		httpLifecycleTenantID.String(), []byte(`{"name":"Renamed"}`))
	request.Header.Set("Idempotency-Key", "tenant-update-http-0001")
	response := httptest.NewRecorder()
	handler.update(response, request)
	if response.Code != 428 || fixture.lastCall != "" {
		t.Fatalf("missing revision status=%d call=%s", response.Code, fixture.lastCall)
	}
	request = lifecycleHTTPRequest(http.MethodPut, "/api/platform/tenants/"+
		httpLifecycleTenantID.String(), []byte(`{"name":"Renamed"}`))
	request.Header.Set("If-Match", `"7"`)
	request.Header.Set("Idempotency-Key", "tenant-update-http-0001")
	response = httptest.NewRecorder()
	handler.update(response, request)
	if response.Code != http.StatusOK || fixture.lastCall != "update" ||
		fixture.mutation.ExpectedRevision != 7 || fixture.mutation.Name != "Renamed" ||
		response.Header().Get("ETag") != `"2"` {
		t.Fatalf("status=%d command=%+v body=%s", response.Code, fixture.mutation, response.Body.String())
	}
}

func TestLegacyStatusUpdatePointsToExplicitAction(t *testing.T) {
	handler, fixture := newLifecycleHTTPFixture(t)
	request := lifecycleHTTPRequest(http.MethodPut, "/api/admin/tenants/"+
		httpLifecycleTenantID.String(), []byte(`{"status":"SUSPENDED"}`))
	request.Header.Set("If-Match", `"1"`)
	request.Header.Set("Idempotency-Key", "tenant-status-http-0001")
	response := httptest.NewRecorder()
	handler.legacyUpdate(response, request)
	if response.Code != 426 || fixture.lastCall != "" || response.Header().Get("Link") == "" {
		t.Fatalf("status=%d call=%s headers=%v", response.Code, fixture.lastCall, response.Header())
	}
}

func TestLegacyDeleteIsAsyncOffboardNotSynchronousDrop(t *testing.T) {
	handler, fixture := newLifecycleHTTPFixture(t)
	request := lifecycleHTTPRequest(http.MethodDelete, "/api/admin/tenants/"+
		httpLifecycleTenantID.String(), nil)
	request.Header.Set("If-Match", `"2"`)
	request.Header.Set("Idempotency-Key", "tenant-delete-http-0001")
	response := httptest.NewRecorder()
	handler.offboard(response, request)
	if response.Code != http.StatusAccepted || fixture.lastCall != "offboard" ||
		fixture.async.ExpectedRevision != 2 || response.Header().Get("Location") == "" {
		t.Fatalf("status=%d call=%s command=%+v body=%s", response.Code,
			fixture.lastCall, fixture.async, response.Body.String())
	}
}

func TestLegacySwitchTenantIsRetired(t *testing.T) {
	handler, fixture := newLifecycleHTTPFixture(t)
	request := httptest.NewRequest(http.MethodPost, "/api/admin/switch-tenant",
		bytes.NewReader([]byte(`{"tenant_id":"`+httpLifecycleTenantID.String()+`"}`)))
	response := httptest.NewRecorder()
	handler.legacySwitchTenant(response, request)
	if response.Code != http.StatusGone || fixture.lastCall != "" ||
		response.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("status=%d call=%s", response.Code, fixture.lastCall)
	}
}

func TestCustomerLifecycleHTTPRejectsTypedNilApplication(t *testing.T) {
	var application *lifecycleHTTPFixture
	if _, err := NewCustomerLifecycleHTTP(application, func(error) {}); err == nil {
		t.Fatal("typed nil application accepted")
	}
	status, _, _ := customerLifecycleFailureDetails(lifecycle.ErrRevisionConflict)
	if status != http.StatusPreconditionFailed {
		t.Fatalf("revision status=%d", status)
	}
}

func lifecycleHTTPRequest(method, target string, body []byte) *http.Request {
	var reader *bytes.Reader
	if body == nil {
		reader = bytes.NewReader(nil)
	} else {
		reader = bytes.NewReader(body)
	}
	request := httptest.NewRequest(method, target, reader)
	request.SetPathValue("tenant_id", httpLifecycleTenantID.String())
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	return request
}
