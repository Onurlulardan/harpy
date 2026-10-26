package queryexecutionhttp

import (
	"net/http"
	"testing"
)

func TestExecuteRoute(t *testing.T) {
	routes := (&Handler{}).Routes()
	if len(routes) != 1 || routes[0].Method != http.MethodPost ||
		routes[0].Pattern != queryExecutionsPath || routes[0].Handler == nil {
		t.Fatal("missing canonical Analytics query batch execution route")
	}
}
