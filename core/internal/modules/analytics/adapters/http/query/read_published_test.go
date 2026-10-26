package queryhttp

import (
	"net/http"
	"testing"
)

func TestReadPublishedRoute(t *testing.T) {
	if !routeExists(
		http.MethodGet,
		queryPath+"/versions/latest-published",
	) {
		t.Fatal("missing latest published Analytics query route")
	}
}
