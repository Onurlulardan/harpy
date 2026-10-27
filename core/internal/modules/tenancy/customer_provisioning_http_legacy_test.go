package tenancy

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	provisioning "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/customerprovisioning"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
)

func TestCustomerProvisioningHTTPLegacyAdminUsesExactCommand(t *testing.T) {
	t.Parallel()
	operation := customerProvisioningHTTPOperation()
	fixture := &customerProvisioningHTTPFixture{operation: operation}
	handler, err := NewCustomerProvisioningHTTP(fixture, func(error) {})
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, "/api/admin/tenants",
		strings.NewReader(`{"name":"Acme","slug":"acme","first_admin":{"display_name":"Ada","email":"ada@acme.test"}}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", "legacy-admin-acme")
	request.Header.Set("Authorization", "Bearer token")
	response := httptest.NewRecorder()
	serveCustomerProvisioningRoutes(t, handler, response, request)
	if response.Code != http.StatusAccepted ||
		fixture.command.Caller != provisioning.CallerHTTP ||
		fixture.command.IdempotencyKey != "legacy-admin-acme" {
		t.Fatalf("status=%d command=%#v body=%s", response.Code,
			fixture.command, response.Body.String())
	}
}

func TestCustomerProvisioningHTTPLegacyAnonymousRegistrationIsRetired(t *testing.T) {
	t.Parallel()
	handler, _ := NewCustomerProvisioningHTTP(
		&customerProvisioningHTTPFixture{}, func(error) {},
	)
	request := httptest.NewRequest(http.MethodPost, "/api/auth/register-tenant", nil)
	response := httptest.NewRecorder()
	serveCustomerProvisioningRoutes(t, handler, response, request)
	if response.Code != http.StatusGone ||
		response.Header().Get("Link") == "" ||
		!strings.Contains(response.Body.String(),
			"LEGACY_ANONYMOUS_TENANT_REGISTRATION_RETIRED") {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestCustomerProvisioningHTTPLegacyAdminMapsOnPremiseDenial(t *testing.T) {
	t.Parallel()
	fixture := &customerProvisioningHTTPFixture{err: provisioning.ErrModeDenied}
	handler, _ := NewCustomerProvisioningHTTP(fixture, func(error) {})
	request := httptest.NewRequest(http.MethodPost, "/api/admin/tenants",
		strings.NewReader(`{"name":"Acme","slug":"acme","first_admin":{"display_name":"Ada","email":"ada@acme.test"}}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", "legacy-admin-acme")
	response := httptest.NewRecorder()
	serveCustomerProvisioningRoutes(t, handler, response, request)
	if response.Code != http.StatusForbidden ||
		!strings.Contains(response.Body.String(), "DEPLOYMENT_MODE_DENIED") {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}

func serveCustomerProvisioningRoutes(
	t *testing.T,
	handler *CustomerProvisioningHTTP,
	response *httptest.ResponseRecorder,
	request *http.Request,
) {
	t.Helper()
	routes, err := httpedge.NewRoutes(func(error) {}, handler.Routes()...)
	if err != nil {
		t.Fatal(err)
	}
	routes.ServeHTTP(response, request)
}
