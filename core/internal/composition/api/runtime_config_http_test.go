package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/buildinfo"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
)

func TestRuntimeConfigHTTPReturnsPersistedOnPremiseSnapshot(t *testing.T) {
	handler := runtimeConfigFixture(t, runtimeConfigSnapshot{
		deploymentMode: config.DeploymentOnPrem, primaryTenantSlug: "harpy",
	})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/app/config", nil))
	if recorder.Code != http.StatusOK || recorder.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("status=%d headers=%v", recorder.Code, recorder.Header())
	}
	var response runtimeConfigEnvelope
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if !response.Success || response.Data.DeploymentMode != "onpremise" ||
		response.Data.DefaultTenantSlug == nil || *response.Data.DefaultTenantSlug != "harpy" ||
		response.Data.ServiceAvailability.Work {
		t.Fatalf("response=%+v", response)
	}
}

func TestRuntimeConfigHTTPPublishesOnlyComposedServices(t *testing.T) {
	handler := runtimeConfigFixture(t, runtimeConfigSnapshot{
		deploymentMode: config.DeploymentOnPrem, primaryTenantSlug: "harpy",
	})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/app/config", nil))
	var response runtimeConfigEnvelope
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Data.ServiceAvailability.Work {
		t.Fatalf("Work stays unpublished until the snapshot marks it composed: %+v",
			response.Data.ServiceAvailability)
	}
	composed := runtimeConfigFixture(t, runtimeConfigSnapshot{
		deploymentMode: config.DeploymentOnPrem, primaryTenantSlug: "harpy",
		workComposed: true,
	})
	composedRecorder := httptest.NewRecorder()
	composed.ServeHTTP(composedRecorder, httptest.NewRequest(http.MethodGet, "/api/app/config", nil))
	var composedResponse runtimeConfigEnvelope
	if err := json.Unmarshal(composedRecorder.Body.Bytes(), &composedResponse); err != nil {
		t.Fatal(err)
	}
	if !composedResponse.Data.ServiceAvailability.Work {
		t.Fatalf("composed services unpublished: %+v", composedResponse.Data.ServiceAvailability)
	}
}

func TestRuntimeConfigHTTPHidesSaaSControlTenantAndSupportsHead(t *testing.T) {
	handler := runtimeConfigFixture(t, runtimeConfigSnapshot{deploymentMode: config.DeploymentSaaS})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodHead, "/api/app/config", nil))
	if recorder.Code != http.StatusOK || recorder.Body.Len() != 0 ||
		recorder.Header().Get("Content-Length") == "" {
		t.Fatalf("status=%d body=%q headers=%v", recorder.Code, recorder.Body.String(), recorder.Header())
	}
}

func TestRuntimeConfigHTTPRejectsInvalidModeSlugCombinations(t *testing.T) {
	for _, snapshot := range []runtimeConfigSnapshot{
		{},
		{deploymentMode: config.DeploymentOnPrem},
		{deploymentMode: config.DeploymentSaaS, primaryTenantSlug: "control"},
	} {
		if _, err := newRuntimeConfigHandler(runtimeConfigInfo(t), snapshot, func(error) {}); err == nil {
			t.Fatalf("snapshot %+v was accepted", snapshot)
		}
	}
}

func TestRuntimeConfigRoutesPublishOnlyGetAndHead(t *testing.T) {
	routes, err := runtimeConfigRoutes(runtimeConfigFixture(t, runtimeConfigSnapshot{
		deploymentMode: config.DeploymentOnPrem, primaryTenantSlug: "harpy",
	}))
	if err != nil || len(routes) != 2 || routes[0].Method != http.MethodGet ||
		routes[1].Method != http.MethodHead {
		t.Fatalf("routes=%v err=%v", routes, err)
	}
}

func runtimeConfigFixture(t *testing.T, snapshot runtimeConfigSnapshot) *runtimeConfigHandler {
	t.Helper()
	handler, err := newRuntimeConfigHandler(runtimeConfigInfo(t), snapshot, func(error) {})
	if err != nil {
		t.Fatal(err)
	}
	return handler
}

func runtimeConfigInfo(t *testing.T) buildinfo.Info {
	t.Helper()
	info, err := buildinfo.ReadAPI(config.EnvironmentTest)
	if err != nil {
		t.Fatal(err)
	}
	return info
}
