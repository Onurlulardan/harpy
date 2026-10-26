package analytics

import (
	"testing"

	compatibilityhttp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/adapters/http/compatibility"
	dashboardhttp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/adapters/http/dashboard"
	dashboardexporthttp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/adapters/http/dashboardexport"
	publicationhttp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/adapters/http/publication"
	queryhttp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/adapters/http/query"
	queryexecutionhttp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/adapters/http/queryexecution"
)

func TestHTTPRouteCountMatchesCanonicalAdapterInventory(t *testing.T) {
	routes := append([]HTTPRoute{}, (&queryhttp.Handler{}).Routes()...)
	routes = append(routes, (&queryexecutionhttp.Handler{}).Routes()...)
	routes = append(routes, (&dashboardhttp.Handler{}).Routes()...)
	routes = append(routes, (&dashboardexporthttp.Handler{}).Routes()...)
	routes = append(routes, (&publicationhttp.Handler{}).Routes()...)
	routes = append(routes, (&compatibilityhttp.Handler{}).Routes()...)
	if err := validateAnalyticsRoutes(routes); err != nil {
		t.Fatalf("Analytics HTTP route manifest is invalid: %v", err)
	}
	handler := &HTTPHandler{routes: routes}
	if !handler.Ready() || len(handler.Routes()) != analyticsHTTPRouteCount {
		t.Fatal("Analytics HTTP handler rejected the canonical route manifest")
	}
	handler.routes[1] = handler.routes[0]
	if handler.Ready() || handler.Routes() != nil {
		t.Fatal("Analytics HTTP handler accepted a duplicate method and path")
	}
}
