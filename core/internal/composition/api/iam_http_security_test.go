package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/config"
)

func TestIAMHTTPSecurityBuildsExactJWKSRoute(t *testing.T) {
	setAPIStartupEnvironment(t)
	settings, err := config.LoadAPI()
	if err != nil {
		t.Fatal(err)
	}
	security, err := buildIAMHTTPSecurity(settings)
	if err != nil {
		t.Fatal(err)
	}
	defer security.Close()
	if !security.Valid() {
		t.Fatal("IAM HTTP security is incomplete")
	}
	routes, err := buildPublicRoutes(security, func(error) {})
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	routes.ServeHTTP(
		response,
		httptest.NewRequest(http.MethodGet, "/.well-known/jwks.json", nil),
	)
	if response.Code != http.StatusOK ||
		response.Header().Get("Content-Type") != "application/jwk-set+json" ||
		response.Header().Get("ETag") == "" {
		t.Fatalf("unexpected JWKS response: %d %v", response.Code, response.Header())
	}
}

func TestBuildPublicRoutesRequiresSecurity(t *testing.T) {
	if _, err := buildPublicRoutes(nil, func(error) {}); err == nil {
		t.Fatal("nil IAM HTTP security was accepted")
	}
}
