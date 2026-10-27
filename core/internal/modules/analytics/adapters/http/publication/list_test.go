package publicationhttp

import (
	"net/http"
	"testing"
)

func TestList(t *testing.T) {
	if !routeExists(http.MethodGet, publicationsPath) {
		t.Fatalf("missing Analytics publication route")
	}
}
