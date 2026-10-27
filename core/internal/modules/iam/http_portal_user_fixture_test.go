package iam

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

const portalUserHTTPTestKey = "portal-user-command-0001"

type portalUserHTTPTestFixture struct {
	handler *PortalUserHTTP
	mux     *http.ServeMux
	app     *portalUserHTTPTestApplication
	auth    *authHTTPTestFixture
	token   string
}

func newPortalUserHTTPTestFixture(t *testing.T) *portalUserHTTPTestFixture {
	t.Helper()
	auth := newAuthHTTPTestFixture(t)
	view := PortalUserView{
		ID:        uuid.MustParse("018f47f4-7b1b-7000-8000-000000000010"),
		BindingID: uuid.MustParse("018f47f4-7b1b-7000-8000-000000000011"),
		SubjectID: uuid.MustParse("018f47f4-7b1b-7000-8000-000000000012"),
		Email:     "portal@example.com", Status: "PENDING", AccountType: "portal",
		SubjectType: "employee", Revision: 3, SetupRequired: true,
		CreatedAt: auth.now.Add(-time.Minute), UpdatedAt: auth.now,
	}
	application := &portalUserHTTPTestApplication{view: view}
	handler, err := auth.security.NewPortalUserHTTP(application, func(error) {})
	if err != nil {
		t.Fatal(err)
	}
	handler.clock = func() time.Time { return auth.now }
	token, err := auth.security.accessSigner.Sign(auth.claims)
	if err != nil {
		t.Fatal(err)
	}
	fixture := &portalUserHTTPTestFixture{
		handler: handler, mux: http.NewServeMux(), app: application, auth: auth, token: token,
	}
	for _, route := range handler.Routes() {
		fixture.mux.Handle(route.Method+" "+route.Pattern, route.Handler)
	}
	return fixture
}

func (fixture *portalUserHTTPTestFixture) request(
	method, path, body string,
) *http.Request {
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Authorization", "Bearer "+fixture.token)
	request.Header.Set(authTenantIDHeader, fixture.auth.claims.Tenant().String())
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	return request
}

func (fixture *portalUserHTTPTestFixture) serve(request *http.Request) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	fixture.mux.ServeHTTP(recorder, request)
	return recorder
}

func portalUserMutationHeaders(request *http.Request, revision int) {
	request.Header.Set("Idempotency-Key", portalUserHTTPTestKey)
	request.Header.Set("If-Match", `"revision-`+strconv.Itoa(revision)+`"`)
}
