package dashboardhttp

import (
	"net/http"
	"testing"
)

func TestReadDefault(t *testing.T) {
	if !routeExists(http.MethodGet, defaultsPath+"/{surface}") {
		t.Fatalf("missing Analytics default dashboard route")
	}
}
