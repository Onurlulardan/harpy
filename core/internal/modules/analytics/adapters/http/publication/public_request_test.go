package publicationhttp

import (
	"net/http"
	"testing"
)

func TestPublicGETRequiresCanonicalContiguousParameters(t *testing.T) {
	for _, raw := range []string{
		"?param[flags][0]=true&param[flags][2]=false",
		"?param[state]=a+plus", "?param[state]=one&param[state]=two", "?unknown=value",
	} {
		request := opaqueRequest(http.MethodGet, "queries", raw)
		if _, err := publicGETParameters(request); err == nil {
			t.Fatalf("raw query accepted: %s", raw)
		}
	}
	request := opaqueRequest(http.MethodGet, "queries", "?param[flags][0]=true&param[flags][1]=false")
	if document, err := publicGETParameters(request); err != nil || len(document) == 0 {
		t.Fatalf("canonical query document=%s error=%v", document, err)
	}
}
