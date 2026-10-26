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

func TestPlatformLicenseRoutesAreExact(t *testing.T) {
	handler, _, _, _ := newPlatformLicenseHTTPFixture(t, "onpremise")
	want := map[string]string{
		http.MethodGet + " /api/platform/license":                                   "read",
		http.MethodPost + " /api/platform/license/refresh":                          "refresh",
		http.MethodGet + " /api/platform/license/refresh-operations/{operation_id}": "operation",
	}
	for _, route := range handler.Routes() {
		delete(want, route.Method+" "+route.Pattern)
	}
	if len(handler.Routes()) != 3 || len(want) != 0 {
		t.Fatalf("routes=%v missing=%v", handler.Routes(), want)
	}
}

func TestPlatformLicenseReadUsesOnPremiseCapability(t *testing.T) {
	handler, access, application, environment :=
		newPlatformLicenseHTTPFixture(t, "onpremise")
	environmentID, err := uuid.Parse(environment.EnvironmentID)
	if err != nil {
		t.Fatal(err)
	}
	application.license = platformlicenseapp.License{
		LicenseID: platformLicenseHTTPV7(t), EnvironmentID: environmentID,
		SnapshotSequence: 9,
		SnapshotSHA256:   "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		State:            "valid", ReleaseChannel: "stable",
		Features: map[string]bool{"workflow": true},
		Limits: map[string]platformlicenseapp.Limit{
			"users": {Value: 100},
		},
		DatabaseNow: time.Date(2026, 7, 23, 12, 0, 0, 0, time.UTC),
	}
	response := httptest.NewRecorder()
	handler.readLicense(response, httptest.NewRequest(http.MethodGet, "/", nil))
	var envelope struct {
		Success bool                `json:"success"`
		Data    platformLicenseView `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if response.Code != http.StatusOK || !envelope.Success ||
		envelope.Data.SnapshotSequence != 9 ||
		access.request.Capability != "platform.license.read" ||
		access.request.ExpectedMode != "onpremise" {
		t.Fatalf("code=%d body=%s request=%+v",
			response.Code, response.Body.String(), access.request)
	}
}

func TestPlatformLicenseSaaSGuardRunsBeforeIAM(t *testing.T) {
	handler, access, _, _ := newPlatformLicenseHTTPFixture(t, "saas")
	response := httptest.NewRecorder()
	handler.readLicense(response, httptest.NewRequest(http.MethodGet, "/", nil))
	if response.Code != http.StatusForbidden || access.calls != 0 {
		t.Fatalf("code=%d IAM calls=%d body=%s",
			response.Code, access.calls, response.Body.String())
	}
}
