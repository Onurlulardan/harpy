package publicationhttp

import (
	"net/http"
	"testing"
)

func TestApprove(t *testing.T) {
	if !routeExists(http.MethodPost, publicationPath+"/approvals") {
		t.Fatalf("missing Analytics publication route")
	}
}
