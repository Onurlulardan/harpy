package publicationhttp

import (
	"net/http"
	"testing"
)

func TestCreate(t *testing.T) {
	if !routeExists(http.MethodPost, publicationsPath) {
		t.Fatalf("missing Analytics publication route")
	}
}
