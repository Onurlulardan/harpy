package publicationhttp

import (
	"net/http"
	"testing"
)

func TestSubmit(t *testing.T) {
	if !routeExists(http.MethodPost, publicationPath+"/submissions") {
		t.Fatalf("missing Analytics publication route")
	}
}
