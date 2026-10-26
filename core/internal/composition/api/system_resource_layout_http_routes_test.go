package api

import (
	"net/http"
	"reflect"
	"testing"
)

func TestHTTPRoutesAreCanonicalAndExact(t *testing.T) {
	routes := new(systemResourceLayoutHTTPHandler).Routes()
	seen := make([]string, len(routes))
	for index, route := range routes {
		seen[index] = route.Method + " " + route.Pattern
	}
	want := []string{
		http.MethodGet + " " + layoutHTTPCollectionPath,
		http.MethodPut + " " + layoutHTTPRevisionPath,
		http.MethodGet + " " + layoutHTTPRevisionPath,
		http.MethodPatch + " " + layoutHTTPRevisionPath,
		http.MethodDelete + " " + layoutHTTPRevisionPath,
		http.MethodPut + " " + layoutHTTPPublicationPath,
		http.MethodGet + " " + layoutHTTPPresentationPath,
	}
	if !reflect.DeepEqual(seen, want) {
		t.Fatalf("routes = %v", seen)
	}
}
