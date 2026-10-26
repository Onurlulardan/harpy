package queryhttp

import (
	"net/http"
	"testing"
)

func TestCreate(t *testing.T) {
	if !routeExists(http.MethodPost, queriesPath) {
		t.Fatalf("missing Analytics query route")
	}
}
