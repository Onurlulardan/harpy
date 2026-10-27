package iam

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type profileHTTPTestFixture struct {
	handler *ProfileHTTP
	mux     *http.ServeMux
	app     *profileHTTPTestApplication
	auth    *authHTTPTestFixture
	token   string
	errors  []error
}

func newProfileHTTPTestFixture(t *testing.T) *profileHTTPTestFixture {
	t.Helper()
	auth := newAuthHTTPTestFixture(t)
	application := newProfileHTTPTestApplication()
	fixture := &profileHTTPTestFixture{app: application, auth: auth}
	access, err := auth.security.accessSigner.Sign(auth.claims)
	if err != nil {
		t.Fatal(err)
	}
	fixture.token = access
	handler, err := newProfileHTTP(
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

func (fixture *profileHTTPTestFixture) request(
	method, path, body string,
) *http.Request {
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Authorization", "Bearer "+fixture.token)
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	return request
}

func (fixture *profileHTTPTestFixture) serve(
	request *http.Request,
) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	fixture.mux.ServeHTTP(recorder, request)
	return recorder
}

func profileHTTPTestSelfView() ProfileSelfView {
	return ProfileSelfView{
		Document: []byte(`{"id":"51000000-0000-4000-8000-000000000001",` +
			`"email":"operator@example.com","status":"ACTIVE"}`),
		Revision: 2,
	}
}
