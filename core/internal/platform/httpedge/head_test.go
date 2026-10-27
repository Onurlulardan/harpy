package httpedge

import (
	"net/http"
	"testing"
)

func TestUnmatchedHEADHasNoBody(t *testing.T) {
	admission := &Admission{}
	admission.Open()
	routes, err := NewRoutes(func(error) {})
	if err != nil {
		t.Fatal(err)
	}
	handler := newTestHandler(t, admission, routes, nil)
	recorder := requestEdge(handler, http.MethodHead, "/missing")
	if recorder.Code != http.StatusNotFound || recorder.Body.Len() != 0 || recorder.Header().Get("Content-Length") == "" {
		t.Fatalf("unexpected HEAD response: %d %#v %q", recorder.Code, recorder.Header(), recorder.Body.String())
	}
}
