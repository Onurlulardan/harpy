package iam

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/permissionrequest"
	"github.com/google/uuid"
)

type permissionRequestHTTPTestFixture struct {
	handler *PermissionRequestHTTP
	app     *permissionRequestHTTPTestApplication
	token   string
	tenant  uuid.UUID
}

func newPermissionRequestHTTPTestFixture(t *testing.T) *permissionRequestHTTPTestFixture {
	t.Helper()
	auth := newAuthHTTPTestFixture(t)
	app := &permissionRequestHTTPTestApplication{
		createResult: PermissionRequestRefView{
			ID: uuid.New().String(), State: "PENDING", Revision: 1,
			StepUpPurposeSHA256: strings.Repeat("a", 43),
		},
		pageResult: PermissionRequestPageView{Items: []PermissionRequestGroupView{}},
		result: PermissionRequestRefView{
			ID: uuid.New().String(), State: "APPROVED", Revision: 3,
			StepUpPurposeSHA256: strings.Repeat("b", 43),
		},
	}
	handler, err := newPermissionRequestHTTP(
		auth.security, app, func(error) {}, func() time.Time { return auth.now },
	)
	if err != nil {
		t.Fatal(err)
	}
	token, err := auth.security.accessSigner.Sign(auth.claims)
	if err != nil {
		t.Fatal(err)
	}
	return &permissionRequestHTTPTestFixture{
		handler: handler, app: app, token: token, tenant: auth.claims.Tenant(),
	}
}

func (fixture *permissionRequestHTTPTestFixture) request(
	method, path, body string,
) *http.Request {
	request := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	request.Header.Set("Authorization", "Bearer "+fixture.token)
	request.Header.Set(authTenantIDHeader, fixture.tenant.String())
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	return request
}

func TestPermissionRequestHTTPRequiresExactBearerAndTenant(t *testing.T) {
	fixture := newPermissionRequestHTTPTestFixture(t)
	valid := fixture.request(http.MethodGet, "/api/permission-requests", "")
	response := httptest.NewRecorder()
	fixture.handler.list(response, valid)
	if response.Code != http.StatusOK || fixture.app.listCalls != 1 {
		t.Fatalf("valid request = %d / calls=%d", response.Code, fixture.app.listCalls)
	}
	wrong := fixture.request(http.MethodGet, "/api/permission-requests", "")
	wrong.Header.Set(authTenantIDHeader, uuid.New().String())
	response = httptest.NewRecorder()
	fixture.handler.list(response, wrong)
	if response.Code != http.StatusUnauthorized || fixture.app.listCalls != 1 {
		t.Fatalf("tenant substitution = %d / calls=%d", response.Code, fixture.app.listCalls)
	}
	duplicate := fixture.request(http.MethodGet, "/api/permission-requests", "")
	duplicate.Header.Add("Authorization", "Bearer "+fixture.token)
	response = httptest.NewRecorder()
	fixture.handler.list(response, duplicate)
	if response.Code != http.StatusUnauthorized || fixture.app.listCalls != 1 {
		t.Fatalf("duplicate bearer = %d / calls=%d", response.Code, fixture.app.listCalls)
	}
}

func TestPermissionRequestCreateRequiresStrictJSONAndIdempotency(t *testing.T) {
	fixture := newPermissionRequestHTTPTestFixture(t)
	body := `{"mode":"SELF_SERVICE","expires_at":"2033-05-18T04:34:00Z","items":[{"target_kind":"USER","target_id":"10000000-0000-4000-8000-000000000001","permission_id":"60000000-0000-4000-8000-000000000001"}]}`
	request := fixture.request(http.MethodPost, "/api/permission-requests", body)
	response := httptest.NewRecorder()
	fixture.handler.create(response, request)
	if response.Code != http.StatusBadRequest || fixture.app.createCalls != 0 {
		t.Fatalf("missing idempotency = %d / calls=%d", response.Code, fixture.app.createCalls)
	}
	request = fixture.request(http.MethodPost, "/api/permission-requests", strings.TrimSuffix(body, "}")+`,"extra":true}`)
	request.Header.Set("Idempotency-Key", "permission-request-0001")
	response = httptest.NewRecorder()
	fixture.handler.create(response, request)
	if response.Code != http.StatusBadRequest || fixture.app.createCalls != 0 {
		t.Fatalf("unknown JSON = %d / calls=%d", response.Code, fixture.app.createCalls)
	}
	request = fixture.request(http.MethodPost, "/api/permission-requests", body)
	request.Header.Set("Idempotency-Key", "permission-request-0001")
	response = httptest.NewRecorder()
	fixture.handler.create(response, request)
	if response.Code != http.StatusCreated || fixture.app.createCalls != 1 ||
		fixture.app.createInput.Mode != permissionrequest.ModeSelfService {
		t.Fatalf("create = %d / calls=%d / %#v", response.Code, fixture.app.createCalls, fixture.app.createInput)
	}
}
