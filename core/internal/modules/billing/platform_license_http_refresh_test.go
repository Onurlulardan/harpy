package billing

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	platformlicenseapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/platformlicense"
	"github.com/google/uuid"
)

func TestPlatformLicenseRefreshIsDurableAndIdempotent(t *testing.T) {
	handler, access, application, environment :=
		newPlatformLicenseHTTPFixture(t, "onpremise")
	operationID := platformLicenseHTTPV7(t)
	application.refresh = platformlicenseapp.RefreshAccepted{
		OperationID: operationID,
		CreatedAt:   time.Date(2026, 7, 23, 12, 0, 0, 0, time.UTC),
	}
	request := httptest.NewRequest(http.MethodPost, "/", nil)
	request.Header.Set("Idempotency-Key", "license-refresh-0001")
	response := httptest.NewRecorder()
	handler.createRefresh(response, request)
	var envelope struct {
		Success bool                               `json:"success"`
		Data    platformLicenseRefreshAcceptedView `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if response.Code != http.StatusAccepted || !envelope.Success ||
		envelope.Data.State != "accepted" ||
		response.Header().Get("Location") !=
			"/api/platform/license/refresh-operations/"+operationID.String() {
		t.Fatalf("code=%d headers=%v body=%s",
			response.Code, response.Header(), response.Body.String())
	}
	command := application.refreshCommand
	environmentID, err := uuid.Parse(environment.EnvironmentID)
	if err != nil {
		t.Fatal(err)
	}
	if access.request.Capability != "platform.license.refresh" ||
		command.EnvironmentID != environmentID ||
		command.IdempotencyKey != "license-refresh-0001" ||
		len(command.RequestSHA256) != 64 {
		t.Fatalf("request=%+v command=%+v", access.request, command)
	}
}

func TestPlatformLicenseRefreshRejectsInvalidKey(t *testing.T) {
	handler, _, application, _ := newPlatformLicenseHTTPFixture(t, "onpremise")
	application.err = platformlicenseapp.ErrInvalid
	request := httptest.NewRequest(http.MethodPost, "/", nil)
	request.Header.Set("Idempotency-Key", "short")
	response := httptest.NewRecorder()
	handler.createRefresh(response, request)
	if response.Code != http.StatusBadRequest ||
		!strings.Contains(response.Body.String(), "REQUEST_INVALID") {
		t.Fatalf("code=%d body=%s", response.Code, response.Body.String())
	}
}

func TestPlatformLicenseRefreshRejectsDuplicateKey(t *testing.T) {
	handler, _, application, _ := newPlatformLicenseHTTPFixture(t, "onpremise")
	request := httptest.NewRequest(http.MethodPost, "/", nil)
	request.Header.Add("Idempotency-Key", "license-refresh-0001")
	request.Header.Add("Idempotency-Key", "license-refresh-0002")
	response := httptest.NewRecorder()
	handler.createRefresh(response, request)
	if response.Code != http.StatusBadRequest ||
		application.refreshCommand.IdempotencyKey != "" {
		t.Fatalf("code=%d command=%+v body=%s",
			response.Code, application.refreshCommand, response.Body.String())
	}
}
