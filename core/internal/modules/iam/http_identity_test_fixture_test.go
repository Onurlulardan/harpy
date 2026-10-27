package iam

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

const identityHTTPTestIdempotencyKey = "identity-http-test-key-0001"

type identityHTTPTestFixture struct {
	handler *IdentityHTTP
	mux     *http.ServeMux
	app     *identityHTTPTestApplication
	auth    *authHTTPTestFixture
	token   string
	errors  []error
}

func newIdentityHTTPTestFixture(t *testing.T) *identityHTTPTestFixture {
	t.Helper()
	auth := newAuthHTTPTestFixture(t)
	application := newIdentityHTTPTestApplication()
	fixture := &identityHTTPTestFixture{app: application, auth: auth}
	access, err := auth.security.accessSigner.Sign(auth.claims)
	if err != nil {
		t.Fatal(err)
	}
	fixture.token = access
	handler, err := newIdentityHTTP(
		auth.security, application,
		func(err error) { fixture.errors = append(fixture.errors, err) },
		func() time.Time { return auth.now },
	)
	if err != nil {
		t.Fatal(err)
	}
	fixture.handler = handler
	fixture.mux = http.NewServeMux()
	for _, route := range handler.Routes() {
		fixture.mux.Handle(route.Method+" "+route.Pattern, route.Handler)
	}
	return fixture
}

func (fixture *identityHTTPTestFixture) request(
	method string, path string, body string,
) *http.Request {
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Authorization", "Bearer "+fixture.token)
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	return request
}

func (fixture *identityHTTPTestFixture) serve(
	request *http.Request,
) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	fixture.mux.ServeHTTP(recorder, request)
	return recorder
}

func identityHTTPTestUser(id uuid.UUID, now time.Time) IdentityUserView {
	return IdentityUserView{
		ID: id, Email: "user@example.com", FirstName: "Ada", LastName: "Lovelace",
		Status: "ACTIVE", RoleIDs: []uuid.UUID{}, OrganizationIDs: []uuid.UUID{},
		Revision: 2, CreatedAt: now.Add(-time.Hour), UpdatedAt: now,
	}
}

func identityHTTPTestRole(
	id uuid.UUID, tenant uuid.UUID, now time.Time,
) IdentityRoleView {
	return IdentityRoleView{
		ID: id, Name: "Operator", ScopeKind: "TENANT", ScopeID: tenant,
		Status: "ACTIVE", Revision: 2,
		CreatedAt: now.Add(-time.Hour), UpdatedAt: now,
	}
}

func identityHTTPTestOrganization(
	id uuid.UUID, now time.Time,
) IdentityOrganizationView {
	return IdentityOrganizationView{
		ID: id, Name: "Research", Slug: "research", Status: "ACTIVE", Revision: 2,
		OwnerIDs:  []uuid.UUID{},
		CreatedAt: now.Add(-time.Hour), UpdatedAt: now,
	}
}

func identityHTTPMutationHeaders(request *http.Request) {
	request.Header.Set("Idempotency-Key", identityHTTPTestIdempotencyKey)
	request.Header.Set("If-Match", `"revision-1"`)
}
