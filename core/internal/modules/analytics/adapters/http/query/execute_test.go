package queryhttp

import (
	"net/http"
	"testing"
)

func TestExecute(t *testing.T) {
	if !routeExists(http.MethodPost, queryPath+"/executions") {
		t.Fatalf("missing Analytics query route")
	}
}
