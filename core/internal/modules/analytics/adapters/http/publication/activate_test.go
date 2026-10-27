package publicationhttp

import (
	"net/http"
	"testing"
)

func TestActivate(t *testing.T) {
	if !routeExists(http.MethodPost, publicationPath+"/activations") {
		t.Fatalf("missing Analytics publication route")
	}
}
