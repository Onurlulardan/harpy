package iam

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSystemTransferSubmitExportUsesExactPermission(t *testing.T) {
	handler, access, application := newSystemTransferHTTPFixture(t)
	request := httptest.NewRequest(http.MethodPost, systemExportCollection,
		strings.NewReader(`{"format":"csv","query":{"select":["id"]}}`))
	request.SetPathValue("resource", "users")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", "iam-transfer-test-key-0001")
	response := httptest.NewRecorder()
	handler.submitSystemExport(response, request)
	if response.Code != http.StatusAccepted ||
		access.request.Permission != "iam.users.export" ||
		access.request.Operation != SystemExportSubmitOperation ||
		application.exportInput.IdempotencyKey != "iam-transfer-test-key-0001" {
		t.Fatalf("status=%d access=%+v input=%+v body=%s", response.Code,
			access.request, application.exportInput, response.Body.String())
	}
}

func TestSystemTransferSourceRequiresIdempotencyAndClosedResource(t *testing.T) {
	handler, _, application := newSystemTransferHTTPFixture(t)
	for _, resource := range []string{"accounts", "users"} {
		request := httptest.NewRequest(http.MethodPost, systemImportSources,
			strings.NewReader("email\na@b.test\n"))
		request.SetPathValue("resource", resource)
		request.Header.Set("Content-Type", "text/csv; charset=utf-8")
		response := httptest.NewRecorder()
		handler.createSystemImportSource(response, request)
		if response.Code != http.StatusBadRequest {
			t.Fatalf("resource=%s status=%d", resource, response.Code)
		}
	}
	if application.sourceInput.Body != nil {
		t.Fatal("invalid source reached application")
	}
}
