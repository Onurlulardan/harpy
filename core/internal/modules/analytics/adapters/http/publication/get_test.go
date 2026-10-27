package publicationhttp

import (
	"net/http"
	"testing"
)

func TestGet(t *testing.T) {
	if !routeExists(http.MethodGet, publicationPath) {
		t.Fatalf("missing Analytics publication route")
	}
}
