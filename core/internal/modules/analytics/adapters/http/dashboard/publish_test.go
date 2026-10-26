package dashboardhttp

import (
	"net/http"
	"testing"
)

func TestPublish(t *testing.T) {
	if !routeExists(http.MethodPost, dashboardPath+"/publications") {
		t.Fatalf("missing Analytics dashboard route")
	}
}
