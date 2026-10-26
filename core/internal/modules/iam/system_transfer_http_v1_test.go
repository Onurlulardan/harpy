package iam

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestSystemTransferV1AliasesReturnUpgradeRequired(t *testing.T) {
	handler, access, application := newSystemTransferHTTPFixture(t)
	routes := handler.Routes()[9:]
	if len(routes) != 30 {
		t.Fatalf("V1 route count=%d want=30", len(routes))
	}
	secret := "v1-iam-transfer-secret"
	seen := make(map[string]bool, len(routes))
	contract := systemTransferV1OpenAPI(t)
	for _, route := range routes {
		key := route.Method + " " + route.Pattern
		if seen[key] {
			t.Fatalf("duplicate V1 route %s", key)
		}
		seen[key] = true
		request := httptest.NewRequest(route.Method, route.Pattern+"?token="+secret,
			strings.NewReader(`{"password":"`+secret+`"}`))
		response := httptest.NewRecorder()
		route.Handler.ServeHTTP(response, request)
		operation := contract.Paths[route.Pattern][strings.ToLower(route.Method)]
		if response.Code != http.StatusUpgradeRequired ||
			response.Header().Get("Cache-Control") != "no-store" ||
			response.Header().Get("Deprecation") != "true" ||
			response.Header().Get("Link") != "<"+operation.Successor+
				">; rel=\"successor-version\"" ||
			!bytes.Contains(response.Body.Bytes(), []byte("CLIENT_UPGRADE_REQUIRED")) ||
			bytes.Contains(response.Body.Bytes(), []byte(secret)) {
			t.Fatalf("retired %s response=%d headers=%v body=%s", key,
				response.Code, response.Header(), response.Body.String())
		}
	}
	if access.request.Operation != "" || application.exportInput.Resource != "" ||
		application.sourceInput.Resource != "" {
		t.Fatal("V1 alias reached IAM authorization or application execution")
	}
}

func systemTransferV1OpenAPI(t *testing.T) systemTransferV1Contract {
	t.Helper()
	document, err := os.ReadFile(
		"../../../../contracts/http/iam-system-transfer-v1-compat.paths.json")
	if err != nil {
		t.Fatal(err)
	}
	var contract systemTransferV1Contract
	if json.Unmarshal(document, &contract) != nil {
		t.Fatal("invalid IAM V1 transfer compatibility fragment")
	}
	return contract
}

func TestSystemTransferV1AliasesMatchLegacyGatewaySurface(t *testing.T) {
	wantSuffixes := []string{"GET /export", "POST /import/analyze",
		"POST /import/execute", "GET /import/jobs", "GET /import/jobs/{job_id}",
		"GET /import/jobs/{job_id}/rows", "POST /import/jobs/{job_id}/resume",
		"POST /import/jobs/{job_id}/cancel", "GET /import/jobs/{job_id}/report",
		"POST /import/sessions/{session_id}/cancel"}
	seen := map[string]bool{}
	for _, route := range new(SystemTransferHTTP).Routes()[9:] {
		seen[route.Method+" "+route.Pattern] = true
	}
	for _, resource := range []string{"users", "roles", "organizations"} {
		for _, suffix := range wantSuffixes {
			parts := strings.SplitN(suffix, " ", 2)
			key := parts[0] + " /api/" + resource + parts[1]
			if !seen[key] {
				t.Fatalf("missing legacy alias %s", key)
			}
		}
	}
}
