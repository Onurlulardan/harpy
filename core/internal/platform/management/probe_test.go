package management

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/health"
)

func TestLivezAndVersionExposeFixedContracts(t *testing.T) {
	fixture := newHandlerFixture(t)
	livez := requestProbe(t, fixture.handler, http.MethodGet, "/livez")
	if livez.Code != http.StatusOK || livez.Body.String() != `{"status":"alive","service":"core-api"}` {
		t.Fatalf("unexpected livez response: %d %s", livez.Code, livez.Body.String())
	}
	version := requestProbe(t, fixture.handler, http.MethodGet, "/version")
	if version.Code != http.StatusOK {
		t.Fatalf("unexpected version status: %d", version.Code)
	}
	var payload runtimev1.Version
	if err := json.Unmarshal(version.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Service != runtimev1.ServiceCoreAPI || payload.ContractVersion != 1 {
		t.Fatalf("unexpected version payload: %#v", payload)
	}
}

func TestReadyzFollowsMonitorState(t *testing.T) {
	fixture := newHandlerFixture(t)
	starting := requestProbe(t, fixture.handler, http.MethodGet, "/readyz")
	assertReadyz(t, starting, http.StatusServiceUnavailable, runtimev1.ReadinessReasonStarting)
	if starting.Header().Get("Retry-After") != "1" {
		t.Fatal("not-ready response is missing retry hint")
	}
	fixture.monitor.Update(health.States{
		Database: true, DatabaseRole: true,
		DeploymentPolicy: true, SchemaContract: true,
	})
	fixture.monitor.MarkRunning()
	ready := requestProbe(t, fixture.handler, http.MethodGet, "/readyz")
	assertReadyz(t, ready, http.StatusOK, runtimev1.ReadinessReasonNone)
	fixture.monitor.BeginDrain()
	draining := requestProbe(t, fixture.handler, http.MethodGet, "/readyz")
	assertReadyz(t, draining, http.StatusServiceUnavailable, runtimev1.ReadinessReasonDraining)
}

func TestProbeHEADMatchesGETHeadersWithoutBody(t *testing.T) {
	fixture := newHandlerFixture(t)
	get := requestProbe(t, fixture.handler, http.MethodGet, "/livez")
	head := requestProbe(t, fixture.handler, http.MethodHead, "/livez")
	if head.Code != get.Code || head.Body.Len() != 0 ||
		head.Header().Get("Content-Length") != get.Header().Get("Content-Length") {
		t.Fatalf("HEAD does not match GET: get=%#v head=%#v", get.Result(), head.Result())
	}
}

func assertReadyz(t *testing.T, recorder *httptest.ResponseRecorder, status int, reason runtimev1.ReadinessReason) {
	t.Helper()
	if recorder.Code != status {
		t.Fatalf("expected status %d, got %d", status, recorder.Code)
	}
	var payload runtimev1.Readyz
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Reason != reason || payload.Service != runtimev1.ServiceCoreAPI || len(payload.Checks) != 4 {
		t.Fatalf("unexpected readiness payload: %#v", payload)
	}
}

func requestProbe(t *testing.T, handler http.Handler, method, path string) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(method, path, nil)
	request.Header.Set("X-Correlation-ID", "probe-1")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	return recorder
}
