package queryhttp

import (
	"net/http"
	"testing"
)

func TestDelete(t *testing.T) {
	if !routeExists(http.MethodDelete, queryPath) {
		t.Fatalf("missing Analytics query route")
	}
}
