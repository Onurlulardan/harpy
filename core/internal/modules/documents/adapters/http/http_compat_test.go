package documentshttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLegacyByteFlowsRecordHitAndRequireUpgrade(t *testing.T) {
	for _, fixture := range []struct {
		name, method, path, operation, resource string
	}{
		{"folder archive", http.MethodGet,
			"/api/folders/" + testDocumentID + "/download",
			operationLegacyArchive, testDocumentID},
		{"document upload", http.MethodPost, "/api/documents",
			operationLegacyUpload, ""},
		{"version upload", http.MethodPost,
			"/api/documents/" + testDocumentID + "/versions",
			operationLegacyUpload, testDocumentID},
	} {
		t.Run(fixture.name, func(t *testing.T) {
			called := false
			handler := newTestHandler(t, applicationStub(func(_ context.Context,
				request Request,
			) (Response, error) {
				called = request.Operation == fixture.operation &&
					request.ResourceID == fixture.resource
				return Response{Status: http.StatusNoContent}, nil
			}))
			request := httptest.NewRequest(fixture.method, fixture.path, nil)
			response := httptest.NewRecorder()
			routeMux(handler.Routes()).ServeHTTP(response, request)
			if !called || response.Code != http.StatusUpgradeRequired ||
				!strings.Contains(response.Body.String(),
					"CLIENT_UPGRADE_REQUIRED") {
				t.Fatalf("legacy response = called:%v status:%d body:%s",
					called, response.Code, response.Body.String())
			}
		})
	}
}
