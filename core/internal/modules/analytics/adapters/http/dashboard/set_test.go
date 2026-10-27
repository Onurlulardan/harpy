package dashboardhttp

import (
	"net/http"
	"testing"
)

func TestSet(t *testing.T) {
	if !routeExists(http.MethodPut, dashboardPath+"/defaults/{surface}") {
		t.Fatalf("missing Analytics dashboard route")
	}
}
