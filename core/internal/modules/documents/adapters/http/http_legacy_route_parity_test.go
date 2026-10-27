package documentshttp

import (
	"context"
	"testing"
)

func TestRoutesContainExactLegacyGatewaySurface(t *testing.T) {
	handler := newTestHandler(t, applicationStub(func(
		context.Context, Request,
	) (Response, error) {
		return Response{Status: 200}, nil
	}))
	actual := make(map[string]bool, len(handler.Routes()))
	for _, route := range handler.Routes() {
		actual[route.Method+" "+route.Pattern] = true
	}
	legacy := []string{
		"GET /api/folders",
		"POST /api/folders",
		"GET /api/folders/{folder_id}",
		"PUT /api/folders/{folder_id}",
		"DELETE /api/folders/{folder_id}",
		"POST /api/folders/{folder_id}/move",
		"GET /api/folders/{folder_id}/contents",
		"GET /api/folders/{folder_id}/download",
		"GET /api/documents",
		"POST /api/documents",
		"POST /api/documents/batch",
		"GET /api/documents/{document_id}",
		"PUT /api/documents/{document_id}",
		"DELETE /api/documents/{document_id}",
		"POST /api/documents/{document_id}/move",
		"POST /api/documents/{document_id}/copy",
		"GET /api/documents/{document_id}/download",
		"GET /api/documents/{document_id}/versions",
		"POST /api/documents/{document_id}/versions",
		"GET /api/documents/{document_id}/versions/latest",
	}
	for _, route := range legacy {
		if !actual[route] {
			t.Fatalf("legacy Gateway route missing: %s", route)
		}
	}
}
