package dashboardhttp

import (
	"net/http"
	"testing"
)

func TestGet(t *testing.T) {
	if !routeExists(http.MethodGet, dashboardPath) {
		t.Fatalf("missing Analytics dashboard route")
	}
}
