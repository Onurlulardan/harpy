package dashboardhttp

import (
	"net/http"
	"testing"
)

func TestRevoke(t *testing.T) {
	if !routeExists(http.MethodPost, dashboardPath+"/acl-revocations") {
		t.Fatalf("missing Analytics dashboard route")
	}
}
