//go:build integration

package migrator

import (
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
)

func waitBuiltEntrypointReady(
	t *testing.T,
	address string,
	service runtimev1.Service,
	done <-chan error,
	finished *bool,
	logFile *os.File,
) {
	t.Helper()
	client := &http.Client{Timeout: 500 * time.Millisecond}
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		select {
		case err := <-done:
			*finished = true
			t.Fatalf("built %s exited before readiness: %v: %s",
				service, err, readBuiltEntrypointLog(logFile))
		default:
		}
		if builtEntrypointProbe(client, address, "/livez") == http.StatusOK &&
			builtEntrypointProbe(client, address, "/readyz") == http.StatusOK {
			assertBuiltEntrypointVersion(t, client, address, service)
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("built %s entrypoint did not become ready: %s",
		service, readBuiltEntrypointLog(logFile))
}

func builtEntrypointProbe(client *http.Client, address, path string) int {
	response, err := client.Get("http://" + address + path)
	if err != nil {
		return 0
	}
	defer response.Body.Close()
	return response.StatusCode
}

func assertBuiltEntrypointVersion(
	t *testing.T,
	client *http.Client,
	address string,
	service runtimev1.Service,
) {
	t.Helper()
	response, err := client.Get("http://" + address + "/version")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	var version runtimev1.Version
	if response.StatusCode != http.StatusOK || json.NewDecoder(response.Body).Decode(&version) != nil {
		t.Fatalf("built %s entrypoint returned invalid version", service)
	}
	if version.Service != service || version.Version != "2.0.0-entrypoint-gate" ||
		version.Environment != runtimev1.EnvironmentTest {
		t.Fatalf("unexpected built %s version: %#v", service, version)
	}
}
