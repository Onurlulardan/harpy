package queryhttp

import (
	"net/http"
	"testing"
)

func TestPublish(t *testing.T) {
	if !routeExists(http.MethodPost, queryPath+"/publications") {
		t.Fatalf("missing Analytics query route")
	}
}
