package dashboardhttp

import (
	"net/http"
	"testing"
)

func TestDelete(t *testing.T) {
	if !routeExists(http.MethodDelete, dashboardPath) {
		t.Fatalf("missing Analytics dashboard route")
	}
}
