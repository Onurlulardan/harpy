package queryhttp

import (
	"net/http"
	"testing"
)

func TestList(t *testing.T) {
	if !routeExists(http.MethodGet, queriesPath) {
		t.Fatalf("missing Analytics query route")
	}
}
