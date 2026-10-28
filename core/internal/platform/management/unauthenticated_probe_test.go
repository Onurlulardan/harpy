package management

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/buildinfo"
	"github.com/Onurlulardan/harpy/core/internal/platform/health"
)

func TestManagementProbesSucceedWithoutAuth(t *testing.T) {
	fixture := newHandlerFixture(t)
	fixture.monitor.Update(health.States{
		Database: true, DatabaseRole: true,
		DeploymentPolicy: true, SchemaContract: true,
	})
	fixture.monitor.MarkRunning()
	for _, path := range []string{"/livez", "/readyz", "/version"} {
		recorder := getWithoutAuth(t, fixture.handler, path)
		if recorder.Code != http.StatusOK {
			t.Fatalf("%s status = %d", path, recorder.Code)
		}
		if recorder.Header().Get("WWW-Authenticate") != "" {
			t.Fatalf("%s challenged for authentication", path)
		}
		assertProbeOmitsSecrets(t, recorder.Body.Bytes())
		switch path {
		case "/livez":
			assertLivezPayload(t, recorder.Body.Bytes())
		case "/readyz":
			assertReadyzPayload(t, recorder.Body.Bytes())
		case "/version":
			assertVersionPayload(t, recorder.Body.Bytes())
		}
	}
}

func getWithoutAuth(t *testing.T, handler http.Handler, path string) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(http.MethodGet, path, nil)
	if request.Header.Get("Authorization") != "" || request.Header.Get("Cookie") != "" {
		t.Fatal("probe request carried auth material")
	}
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	return recorder
}

func assertLivezPayload(t *testing.T, body []byte) {
	t.Helper()
	assertExactKeys(t, body, "service", "status")
	var payload runtimev1.Livez
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Status != runtimev1.LivenessStatusAlive || payload.Service != runtimev1.ServiceCoreAPI {
		t.Fatal("livez payload does not match the liveness contract")
	}
}

func assertReadyzPayload(t *testing.T, body []byte) {
	t.Helper()
	object := assertExactKeys(t, body, "checks", "reason", "service", "status")
	var payload runtimev1.Readyz
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Status != runtimev1.ReadinessStatusReady ||
		payload.Service != runtimev1.ServiceCoreAPI ||
		payload.Reason != runtimev1.ReadinessReasonNone {
		t.Fatal("readyz payload does not match the ready contract")
	}
	var checks []map[string]json.RawMessage
	if err := json.Unmarshal(object["checks"], &checks); err != nil {
		t.Fatal(err)
	}
	names := []runtimev1.ReadinessCheckName{
		runtimev1.ReadinessCheckDatabase,
		runtimev1.ReadinessCheckDatabaseRole,
		runtimev1.ReadinessCheckDeploymentPolicy,
		runtimev1.ReadinessCheckSchemaContract,
	}
	if len(checks) != len(names) {
		t.Fatal("readyz check count mismatch")
	}
	for index, name := range names {
		if _, ok := checks[index]["name"]; !ok || len(checks[index]) != 2 {
			t.Fatal("readyz check fields are not name and status")
		}
		if _, ok := checks[index]["status"]; !ok {
			t.Fatal("readyz check fields are not name and status")
		}
		if payload.Checks[index].Name != name ||
			payload.Checks[index].Status != runtimev1.ReadinessStatusReady {
			t.Fatal("readyz check does not match the ready contract")
		}
	}
}

func assertVersionPayload(t *testing.T, body []byte) {
	t.Helper()
	assertExactKeys(t, body,
		"build_time", "commit", "contract_version", "environment",
		"go_version", "service", "version",
	)
	info, err := buildinfo.Read(runtimev1.ServiceCoreAPI, runtimev1.EnvironmentDevelopment)
	if err != nil {
		t.Fatal(err)
	}
	var payload runtimev1.Version
	if err = json.Unmarshal(body, &payload); err != nil {
		t.Fatal(err)
	}
	if payload != info.Version() {
		t.Fatal("version payload does not match buildinfo.Version")
	}
}

func assertExactKeys(t *testing.T, body []byte, want ...string) map[string]json.RawMessage {
	t.Helper()
	var object map[string]json.RawMessage
	if err := json.Unmarshal(body, &object); err != nil {
		t.Fatal(err)
	}
	if len(object) != len(want) {
		t.Fatalf("json key count = %d", len(object))
	}
	for _, key := range want {
		if _, ok := object[key]; !ok {
			t.Fatalf("missing json key %s", key)
		}
	}
	return object
}

func assertProbeOmitsSecrets(t *testing.T, body []byte) {
	t.Helper()
	folded := strings.ToLower(string(body))
	for _, needle := range []string{
		"password",
		"token",
		"dsn",
		"postgres://",
		"postgresql://",
		"private_key",
		"begin private",
		"begin rsa",
		"begin openssh",
		"api_key",
		"secret_key",
	} {
		if strings.Contains(folded, needle) {
			t.Fatalf("management probe body contains %q", needle)
		}
	}
}
