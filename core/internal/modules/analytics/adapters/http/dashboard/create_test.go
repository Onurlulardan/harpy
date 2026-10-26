package dashboardhttp

import (
	"net/http"
	"testing"
)

func TestCreate(t *testing.T) {
	if !routeExists(http.MethodPost, dashboardsPath) {
		t.Fatalf("missing Analytics dashboard route")
	}
}
