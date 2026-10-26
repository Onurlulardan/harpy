package lifecyclehttp

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	cataloghttp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/adapters/http"
)

func TestInstallRejectsMalformedNilAndNonCanonicalPackageIDs(t *testing.T) {
	runtime, err := cataloghttp.NewRuntime(
		new(installationAccess), func(error) {},
		func(*http.Request) string { return "018f0f8b-4c31-7d4e-8a2f-9e0b7c6d5a41" },
	)
	if err != nil {
		t.Fatal(err)
	}
	for _, packageID := range []string{
		"not-a-uuid", "00000000-0000-0000-0000-000000000000",
		"E3000000-0000-0000-0000-000000000001",
	} {
		t.Run(packageID, func(t *testing.T) {
			application := new(installationApplication)
			request := httptest.NewRequest(
				http.MethodPost, "/api/catalog/installations",
				strings.NewReader(`{"package_id":"`+packageID+`","version":"1.2.3"}`),
			)
			request.Header.Set("Content-Type", "application/json")
			request.Header.Set("Idempotency-Key", "catalog-test-key-0001")
			response := httptest.NewRecorder()
			New(runtime, application).InstallRoute().Handler.ServeHTTP(response, request)
			if response.Code != http.StatusBadRequest || application.called != "" {
				t.Fatalf("status=%d call=%s", response.Code, application.called)
			}
		})
	}
}
