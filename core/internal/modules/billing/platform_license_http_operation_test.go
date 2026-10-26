package billing

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	platformlicenseapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/platformlicense"
	"github.com/google/uuid"
)

func TestPlatformLicenseRefreshOperationPreservesNulls(t *testing.T) {
	handler, access, application, environment :=
		newPlatformLicenseHTTPFixture(t, "onpremise")
	operationID := platformLicenseHTTPV7(t)
	application.operation = platformlicenseapp.RefreshOperation{
		OperationID: operationID, State: "pending",
		CreatedAt: time.Date(2026, 7, 23, 12, 0, 0, 0, time.UTC),
	}
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.SetPathValue("operation_id", operationID.String())
	response := httptest.NewRecorder()
	handler.readRefreshOperation(response, request)
	var envelope struct {
		Success bool                                `json:"success"`
		Data    platformLicenseRefreshOperationView `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	environmentID, err := uuid.Parse(environment.EnvironmentID)
	if err != nil {
		t.Fatal(err)
	}
	if response.Code != http.StatusOK || !envelope.Success ||
		envelope.Data.SnapshotSequence != nil ||
		envelope.Data.SnapshotSHA256 != nil ||
		envelope.Data.FailureCode != nil || envelope.Data.CompletedAt != nil ||
		access.request.Capability != "platform.license.read" ||
		application.environmentID != environmentID ||
		application.operationID != operationID {
		t.Fatalf("code=%d body=%s request=%+v",
			response.Code, response.Body.String(), access.request)
	}
}
