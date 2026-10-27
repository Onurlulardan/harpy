package dashboardexporthttp

import (
	"net/http"
	"strings"
	"testing"
)

func TestRoutesPublishOnlyCanonicalAsyncDashboardExports(t *testing.T) {
	t.Parallel()
	routes := (&Handler{}).Routes()
	want := []struct{ method, pattern string }{
		{http.MethodPost, "/api/analytics/dashboards/{dashboard_id}/exports"},
		{http.MethodGet, "/api/analytics/dashboards/{dashboard_id}/exports/{export_id}"},
		{http.MethodDelete, "/api/analytics/dashboards/{dashboard_id}/exports/{export_id}"},
		{http.MethodGet, "/api/analytics/dashboards/{dashboard_id}/exports/{export_id}/content"},
	}
	if len(routes) != len(want) {
		t.Fatalf("route count = %d, want %d", len(routes), len(want))
	}
	for index, route := range routes {
		if route.Method != want[index].method || route.Pattern != want[index].pattern {
			t.Fatalf("route[%d] = %s %s, want %s %s", index, route.Method,
				route.Pattern, want[index].method, want[index].pattern)
		}
		if strings.Contains(route.Pattern, "/export/") {
			t.Fatalf("legacy synchronous export route was published: %s", route.Pattern)
		}
	}
}
