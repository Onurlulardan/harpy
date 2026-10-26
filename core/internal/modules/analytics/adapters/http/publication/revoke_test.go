package publicationhttp

import (
	"net/http"
	"testing"
)

func TestRevoke(t *testing.T) {
	if !routeExists(http.MethodPost, publicationPath+"/revocations") {
		t.Fatalf("missing Analytics publication route")
	}
}
