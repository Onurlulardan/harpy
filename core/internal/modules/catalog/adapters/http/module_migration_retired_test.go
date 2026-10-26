package cataloghttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type retirementAccessFixture struct{}

func (retirementAccessFixture) Verify(context.Context, AccessRequest) (Access, error) {
	return Access{}, ErrForbidden
}

func TestModuleMigrationRetirementRoutesAreRejectOnly(t *testing.T) {
	t.Parallel()
	runtime, err := NewRuntime(retirementAccessFixture{}, func(error) {},
		func(*http.Request) string { return "00000000-0000-0000-0000-000000000001" })
	if err != nil {
		t.Fatal(err)
	}
	routes := ModuleMigrationRetirementRoutes(runtime)
	if len(routes) != 4 {
		t.Fatalf("routes = %d", len(routes))
	}
	for _, route := range routes {
		request := httptest.NewRequest(route.Method, route.Pattern, nil)
		response := httptest.NewRecorder()
		route.Handler.ServeHTTP(response, request)
		if response.Code != http.StatusGone ||
			!strings.Contains(response.Body.String(), "STANDALONE_MODULE_MIGRATION_RETIRED") ||
			!strings.Contains(response.Header().Get("Link"), "/api/catalog/installations/") {
			t.Fatalf("route=%s status=%d body=%s", route.Pattern,
				response.Code, response.Body.String())
		}
	}
}
