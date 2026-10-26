package dashboardhttp

import (
	"net/http"
	"testing"
)

func TestGrant(t *testing.T) {
	if !routeExists(http.MethodPost, dashboardPath+"/acl-grants") {
		t.Fatalf("missing Analytics dashboard route")
	}
}
