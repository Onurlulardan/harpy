package documentshttp

import (
	"context"
	"fmt"
	"testing"
)

func TestRoutesAreClosedAndUnambiguous(t *testing.T) {
	handler := newTestHandler(t, applicationStub(func(context.Context, Request) (Response, error) {
		return Response{Status: 200, Data: map[string]bool{"ok": true}}, nil
	}))
	routes := handler.Routes()
	if len(routes) != 32 {
		t.Fatalf("route count = %d, want 32", len(routes))
	}
	seen := make(map[string]struct{}, len(routes))
	for _, route := range routes {
		key := route.Method + " " + route.Pattern
		if _, duplicate := seen[key]; duplicate {
			t.Fatalf("duplicate route %s", key)
		}
		seen[key] = struct{}{}
	}
	for _, required := range []string{
		"POST /api/document-uploads",
		"PUT /api/document-uploads/{upload_id}/content",
		"POST /api/document-archives",
		"GET /api/document-archives/{archive_id}/content",
		"GET /api/documents/{document_id}/versions/{version_id}/content",
	} {
		if _, found := seen[required]; !found {
			t.Fatal(fmt.Sprintf("required route missing: %s", required))
		}
	}
}
