package dashboardhttp

import (
	"net/http"
	"testing"
)

func TestSynchronousDashboardExportRoutesRemainUnpublished(t *testing.T) {
	for _, format := range []string{"csv", "excel", "pdf"} {
		pattern := dashboardPath + "/export/" + format
		if routeExists(http.MethodGet, pattern) {
			t.Fatalf("synchronous dashboard export route must remain unpublished: %s", pattern)
		}
	}
}
