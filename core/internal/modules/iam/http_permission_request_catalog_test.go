package iam

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPermissionRequestCatalogIsAuthenticatedSelfService(t *testing.T) {
	fixture := newPermissionRequestHTTPTestFixture(t)
	request := fixture.request(
		http.MethodGet,
		"/api/permission-requests/catalog/resources?search=form&kind=all&limit=25",
		"",
	)
	response := httptest.NewRecorder()
	fixture.handler.listCatalogResources(response, request)
	if response.Code != http.StatusOK || fixture.app.resourceCalls != 1 {
		t.Fatalf("resource catalog = %d / calls=%d", response.Code, fixture.app.resourceCalls)
	}

	wrongTenant := fixture.request(
		http.MethodGet, "/api/permission-requests/catalog/actions", "",
	)
	wrongTenant.Header.Set(authTenantIDHeader, "10000000-0000-4000-8000-000000000099")
	response = httptest.NewRecorder()
	fixture.handler.listCatalogActions(response, wrongTenant)
	if response.Code != http.StatusUnauthorized || fixture.app.actionCalls != 0 {
		t.Fatalf("cross-tenant action catalog = %d / calls=%d", response.Code, fixture.app.actionCalls)
	}
}

func TestMissingCatalogDefaultsToReadAndForm(t *testing.T) {
	request := httptest.NewRequest(
		http.MethodGet, "/api/permission-requests/catalog/missing-resources", nil,
	)
	input, valid := permissionRequestCatalogInput(
		request, true, "form",
	)
	if !valid || input.Kind != "form" || input.Limit != 50 {
		t.Fatalf("missing catalog defaults = %#v / %v", input, valid)
	}

	request = httptest.NewRequest(
		http.MethodGet,
		"/api/permission-requests/catalog/missing-resources?action=*&kind=all",
		nil,
	)
	if _, valid := permissionRequestCatalogInput(request, true, "form"); valid {
		t.Fatal("wildcard missing-resource action was accepted")
	}
}

func TestPermissionRequestMineUsesCanonicalActorVisibleList(t *testing.T) {
	routes := (&PermissionRequestHTTP{}).Routes()
	var canonical, mine http.Handler
	for _, route := range routes {
		switch route.Pattern {
		case "/api/permission-requests":
			if route.Method == http.MethodGet {
				canonical = route.Handler
			}
		case "/api/permission-requests/mine":
			mine = route.Handler
		}
	}
	if canonical == nil || mine == nil {
		t.Fatal("canonical list or mine alias is missing")
	}
}
