package iam

import (
	"net/http"
	"testing"
)

func TestIdentityHTTPUsesStrictBoundedJSON(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		configure  func(*http.Request)
		wantStatus int
		wantCode   string
	}{
		{
			"reserved member", `{"name":"Research","slug":"research","password_hash":true}`,
			func(_ *http.Request) {}, http.StatusBadRequest, "REQUEST_INVALID",
		},
		{
			"unsupported media", `{"name":"Research","slug":"research"}`,
			func(request *http.Request) { request.Header.Set("Content-Type", "text/plain") },
			http.StatusUnsupportedMediaType, "MEDIA_TYPE_UNSUPPORTED",
		},
		{
			"oversize", `{"name":"Research","slug":"research"}`,
			func(request *http.Request) { request.ContentLength = identityHTTPBodyLimit + 1 },
			http.StatusRequestEntityTooLarge, "REQUEST_BODY_TOO_LARGE",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := newIdentityHTTPTestFixture(t)
			request := fixture.request(http.MethodPost, "/api/organizations", test.body)
			request.Header.Set("Idempotency-Key", identityHTTPTestIdempotencyKey)
			test.configure(request)
			response := fixture.serve(request)
			if response.Code != test.wantStatus || len(fixture.app.calls) != 0 ||
				!responseHasIdentityErrorCode(response.Body.Bytes(), test.wantCode) {
				t.Fatalf("status=%d calls=%d body=%s", response.Code,
					len(fixture.app.calls), response.Body.String())
			}
		})
	}
}
