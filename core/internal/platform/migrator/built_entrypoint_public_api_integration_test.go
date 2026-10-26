//go:build integration

package migrator

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/requestmeta"
)

type emptyAPIProbe struct {
	method string
	path   string
	head   bool
}

func assertBuiltEntrypointEmptyAPI(t *testing.T, address string) {
	t.Helper()
	client := &http.Client{
		Timeout: 2 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	waitBuiltEntrypointPublic(t, client, address)
	probes := []emptyAPIProbe{
		{method: http.MethodGet, path: "/missing"},
		{method: http.MethodPost, path: "/missing"},
		{method: http.MethodHead, path: "/missing", head: true},
		{method: http.MethodGet, path: "/missing/"},
		{method: http.MethodGet, path: "/a/../missing"},
		{method: http.MethodGet, path: "//missing"},
	}
	for index, probe := range probes {
		correlationID := "built-empty-api-" + string(rune('a'+index))
		response := executeEmptyAPIProbe(t, client, address, probe, correlationID)
		assertEmptyAPIResponse(t, response, probe.head, correlationID)
	}
}

func waitBuiltEntrypointPublic(t *testing.T, client *http.Client, address string) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for {
		request, err := http.NewRequest(http.MethodGet, "http://"+address+"/startup", nil)
		if err != nil {
			t.Fatal(err)
		}
		request.Header.Set(requestmeta.CorrelationHeader, "built-api-startup")
		response, requestErr := client.Do(request)
		if requestErr == nil {
			_ = response.Body.Close()
			if response.StatusCode == http.StatusNotFound {
				return
			}
		}
		if time.Now().After(deadline) {
			t.Fatalf("built core-api public listener did not become ready: %v", requestErr)
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func executeEmptyAPIProbe(
	t *testing.T,
	client *http.Client,
	address string,
	probe emptyAPIProbe,
	correlationID string,
) *http.Response {
	t.Helper()
	var body io.Reader
	if probe.method == http.MethodPost {
		body = strings.NewReader(`{"ignored":true}`)
	}
	request, err := http.NewRequest(probe.method, "http://"+address+probe.path, body)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set(requestmeta.CorrelationHeader, correlationID)
	request.Header.Set("Origin", "https://cross-origin.invalid")
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("built core-api %s %s: %v", probe.method, probe.path, err)
	}
	return response
}

func assertEmptyAPIResponse(
	t *testing.T,
	response *http.Response,
	head bool,
	correlationID string,
) {
	t.Helper()
	defer response.Body.Close()
	document, err := io.ReadAll(io.LimitReader(response.Body, 4097))
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusNotFound || response.Header.Get("Location") != "" ||
		response.Header.Get("Content-Type") != "application/json; charset=utf-8" ||
		response.Header.Get(requestmeta.CorrelationHeader) != correlationID {
		t.Fatalf("unexpected empty API response: %d %#v", response.StatusCode, response.Header)
	}
	for name := range response.Header {
		if strings.HasPrefix(strings.ToLower(name), "access-control-") {
			t.Fatalf("empty API emitted CORS header %s", name)
		}
	}
	if head {
		if len(document) != 0 || response.ContentLength <= 0 {
			t.Fatalf("unexpected empty API HEAD body: length=%d content-length=%d", len(document), response.ContentLength)
		}
		return
	}
	var envelope runtimev1.ErrorEnvelope
	if len(document) > 4096 || json.Unmarshal(document, &envelope) != nil ||
		envelope.Validate() != nil || envelope.Success ||
		envelope.Error.Code != "ROUTE_NOT_FOUND" ||
		envelope.Error.Message != "Route not found" ||
		envelope.Error.CorrelationID != correlationID || len(envelope.Error.Issues) != 0 {
		t.Fatalf("unexpected empty API envelope: %s", document)
	}
}
