package queryhttp

import (
	"net/http"
	"testing"
)

func TestGet(t *testing.T) {
	if !routeExists(http.MethodGet, queryPath) {
		t.Fatalf("missing Analytics query route")
	}
}
