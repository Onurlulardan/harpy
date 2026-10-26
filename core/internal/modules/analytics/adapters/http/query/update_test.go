package queryhttp

import (
	"net/http"
	"testing"
)

func TestUpdate(t *testing.T) {
	if !routeExists(http.MethodPatch, queryPath) {
		t.Fatalf("missing Analytics query route")
	}
}
