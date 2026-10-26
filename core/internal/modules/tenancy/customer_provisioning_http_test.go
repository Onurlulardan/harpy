package tenancy

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	provisioning "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/customerprovisioning"
	"github.com/google/uuid"
)

type customerProvisioningHTTPFixture struct {
	operation provisioning.Operation
	err       error
	command   provisioning.Command
	reissue   provisioning.ReissueCommand
}

func (fixture *customerProvisioningHTTPFixture) Create(
	_ context.Context, command provisioning.Command,
) (provisioning.Operation, error) {
	fixture.command = command
	return fixture.operation, fixture.err
}

func (fixture *customerProvisioningHTTPFixture) Read(
	context.Context, provisioning.ReadQuery,
) (provisioning.Operation, error) {
	return fixture.operation, fixture.err
}

func (fixture *customerProvisioningHTTPFixture) ReissueInitialAdminSetup(
	_ context.Context, command provisioning.ReissueCommand,
) (provisioning.Operation, error) {
	fixture.reissue = command
	return fixture.operation, fixture.err
}

func TestCustomerProvisioningHTTPCreateReturnsStableLocation(t *testing.T) {
	t.Parallel()
	operation := customerProvisioningHTTPOperation()
	fixture := &customerProvisioningHTTPFixture{operation: operation}
	handler, err := NewCustomerProvisioningHTTP(fixture, func(error) {})
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, "/api/platform/tenants",
		strings.NewReader(`{"name":"Acme","slug":"acme","first_admin":{"display_name":"Ada","email":"ada@acme.test"}}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", "customer-acme")
	request.Header.Set("Authorization", "Bearer token")
	response := httptest.NewRecorder()
	handler.create(response, request)
	location := customerProvisioningOperationsPath + operation.OperationID.String()
	if response.Code != http.StatusAccepted || response.Header().Get("Location") != location {
		t.Fatalf("status=%d location=%q body=%s", response.Code,
			response.Header().Get("Location"), response.Body.String())
	}
	if fixture.command.IdempotencyKey != "customer-acme" ||
		fixture.command.Customer.FirstAdmin.Email != "ada@acme.test" ||
		len(fixture.command.Credential.AuthorizationValues) != 1 {
		t.Fatalf("command = %#v", fixture.command)
	}
	if strings.Contains(response.Body.String(), "ada@acme.test") ||
		strings.Contains(response.Body.String(), operation.RequestedBySubjectID.String()) {
		t.Fatal("accepted response leaked request or subject data")
	}
}

func TestCustomerProvisioningHTTPRejectsUnknownBodyAndMapsMode(t *testing.T) {
	t.Parallel()
	fixture := &customerProvisioningHTTPFixture{operation: customerProvisioningHTTPOperation()}
	handler, _ := NewCustomerProvisioningHTTP(fixture, func(error) {})
	unknown := httptest.NewRequest(http.MethodPost, "/api/platform/tenants",
		strings.NewReader(`{"name":"Acme","slug":"acme","first_admin":{"display_name":"Ada","email":"ada@acme.test"},"status":"ACTIVE"}`))
	unknown.Header.Set("Content-Type", "application/json")
	unknown.Header.Set("Idempotency-Key", "customer-acme")
	response := httptest.NewRecorder()
	handler.create(response, unknown)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("unknown field status=%d body=%s", response.Code, response.Body.String())
	}
	fixture.err = provisioning.ErrModeDenied
	denied := httptest.NewRequest(http.MethodPost, "/api/platform/tenants",
		strings.NewReader(`{"name":"Acme","slug":"acme","first_admin":{"display_name":"Ada","email":"ada@acme.test"}}`))
	denied.Header.Set("Content-Type", "application/json")
	denied.Header.Set("Idempotency-Key", "customer-acme")
	response = httptest.NewRecorder()
	handler.create(response, denied)
	if response.Code != http.StatusForbidden ||
		!strings.Contains(response.Body.String(), "DEPLOYMENT_MODE_DENIED") {
		t.Fatalf("mode status=%d body=%s", response.Code, response.Body.String())
	}
}

func customerProvisioningHTTPOperation() provisioning.Operation {
	now := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	return provisioning.Operation{
		OperationID:          uuid.MustParse("11111111-1111-7111-8111-111111111111"),
		TenantID:             uuid.MustParse("22222222-2222-7222-8222-222222222222"),
		RequestedBySubjectID: uuid.MustParse("33333333-3333-7333-8333-333333333333"),
		State:                tenantapi.OnboardingMigrating, SetupDeliveryState: "pending",
		CreatedAt: now, UpdatedAt: now,
	}
}

func TestCustomerProvisioningHTTPFailureDetails(t *testing.T) {
	t.Parallel()
	status, code, _ := customerProvisioningFailureDetails(errors.New("unknown"))
	if status != http.StatusServiceUnavailable || code != "TENANT_PROVISIONING_UNAVAILABLE" {
		t.Fatalf("status=%d code=%s", status, code)
	}
}

func TestCustomerProvisioningHTTPReissueIsBodylessAndTenantScoped(t *testing.T) {
	t.Parallel()
	operation := customerProvisioningHTTPOperation()
	fixture := &customerProvisioningHTTPFixture{operation: operation}
	handler, err := NewCustomerProvisioningHTTP(fixture, func(error) {})
	if err != nil {
		t.Fatal(err)
	}
	path := "/api/platform/tenants/" + operation.TenantID.String() +
		"/admin-setup/reissue"
	request := httptest.NewRequest(http.MethodPost, path, nil)
	request.SetPathValue("tenant_id", operation.TenantID.String())
	request.Header.Set("Idempotency-Key", "reissue-acme-admin")
	request.Header.Set("Authorization", "Bearer token")
	response := httptest.NewRecorder()
	handler.reissueInitialAdminSetup(response, request)
	if response.Code != http.StatusAccepted ||
		fixture.reissue.TenantID != operation.TenantID ||
		fixture.reissue.IdempotencyKey != "reissue-acme-admin" {
		t.Fatalf("status=%d command=%#v body=%s", response.Code,
			fixture.reissue, response.Body.String())
	}
	request = httptest.NewRequest(http.MethodPost, path, strings.NewReader(`{}`))
	request.SetPathValue("tenant_id", operation.TenantID.String())
	request.Header.Set("Idempotency-Key", "reissue-acme-admin")
	response = httptest.NewRecorder()
	handler.reissueInitialAdminSetup(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("non-empty body status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestCustomerProvisioningHTTPReadHasDistinctTimestamps(t *testing.T) {
	t.Parallel()
	operation := customerProvisioningHTTPOperation()
	operation.UpdatedAt = operation.CreatedAt.Add(time.Minute)
	handler, err := NewCustomerProvisioningHTTP(
		&customerProvisioningHTTPFixture{operation: operation}, func(error) {},
	)
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodGet,
		customerProvisioningOperationsPath+operation.OperationID.String(), nil)
	request.SetPathValue("operation_id", operation.OperationID.String())
	response := httptest.NewRecorder()
	handler.read(response, request)
	var body struct {
		Success bool `json:"success"`
		Data    struct {
			CreatedAt time.Time `json:"created_at"`
			UpdatedAt time.Time `json:"updated_at"`
		} `json:"data"`
	}
	if response.Code != http.StatusOK || json.Unmarshal(response.Body.Bytes(), &body) != nil ||
		!body.Success || !body.Data.CreatedAt.Equal(operation.CreatedAt) ||
		!body.Data.UpdatedAt.Equal(operation.UpdatedAt) {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}
