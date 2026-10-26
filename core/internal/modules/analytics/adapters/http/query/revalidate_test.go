package queryhttp

import (
	"net/http"
	"testing"
)

func TestRevalidate(t *testing.T) {
	if !routeExists(http.MethodPost, queryPath+"/revalidations") {
		t.Fatalf("missing Analytics query route")
	}
}
