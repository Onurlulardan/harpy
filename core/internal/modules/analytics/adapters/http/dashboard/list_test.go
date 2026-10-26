package dashboardhttp

import (
	"net/http"
	"testing"
)

func TestList(t *testing.T) {
	if !routeExists(http.MethodGet, dashboardsPath) {
		t.Fatalf("missing Analytics dashboard route")
	}
}
