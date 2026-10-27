package dashboardhttp

import (
	"net/http"
	"testing"
)

func TestUpdate(t *testing.T) {
	if !routeExists(http.MethodPatch, dashboardPath) {
		t.Fatalf("missing Analytics dashboard route")
	}
}
