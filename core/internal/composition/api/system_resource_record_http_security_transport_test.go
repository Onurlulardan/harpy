package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	systemresources "github.com/Onurlulardan/harpy/core/internal/orchestration/systemresources"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
)

func newRecordHTTPSecurityRoutes(
	t *testing.T,
	app *recordHTTPSecurityApplication,
	access systemResourceRecordHTTPAccessVerifier,
) *httpedge.Routes {
	t.Helper()
	if access == nil {
		access = newRecordHTTPFixture(t).access
	}
	handler, err := newSystemResourceRecordHTTPHandler(systemResourceRecordHTTPDependencies{
		Access: access, Application: app, Report: func(error) {},
	})
	if err != nil {
		t.Fatal(err)
	}
	routes, err := httpedge.NewRoutes(func(error) {}, handler.Routes()...)
	if err != nil {
		t.Fatal(err)
	}
	return routes
}

func recordHTTPSecurityDo(
	routes *httpedge.Routes,
	method, path, body string,
	headers map[string]string,
) *httptest.ResponseRecorder {
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	request := httptest.NewRequest(method, path, reader)
	request.Header.Set("Authorization", "Bearer fixture")
	request.Header.Set("X-Tenant-ID", recordHTTPTestTenant)
	request.Header.Set("X-Tenant-Slug", "harpy")
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	for name, value := range headers {
		request.Header.Set(name, value)
	}
	response := httptest.NewRecorder()
	routes.ServeHTTP(response, request)
	return response
}

func assertRecordHTTPSecurityFlow(
	t *testing.T,
	app *recordHTTPSecurityApplication,
	resource systemresources.Resource,
	password string,
	responses []*httptest.ResponseRecorder,
) {
	t.Helper()
	if len(app.calls) != 6 {
		t.Fatalf("delegated calls=%v, want prepare/create/get/update/delete/list", app.calls)
	}
	for operation, seen := range app.calls {
		if seen != resource {
			t.Fatalf("%s delegated to %s, want %s", operation, seen, resource)
		}
	}
	wantStatus := []int{201, 200, 200, 200, 200}
	for index, response := range responses {
		if response.Code != wantStatus[index] {
			t.Fatalf("response[%d]=%d body=%s", index, response.Code, response.Body)
		}
	}
	for _, index := range []int{0, 1, 2} {
		if responses[index].Header().Get("ETag") != `"revision-11"` ||
			!strings.Contains(responses[index].Body.String(), recordHTTPTestID) ||
			!strings.Contains(responses[index].Body.String(), `"revision":11`) {
			t.Fatalf("response[%d] lacks identity contract: %s", index, responses[index].Body)
		}
	}
	if string(app.password) != password ||
		app.createKey != "security-create-"+string(resource) ||
		password != "" && (strings.Contains(responses[0].Body.String(), password) ||
			strings.Contains(responses[0].Body.String(), `"password"`)) {
		t.Fatalf("password contract failed: captured=%q body=%s", app.password, responses[0].Body)
	}
	if !strings.Contains(responses[4].Body.String(), recordHTTPTestID) ||
		!strings.Contains(responses[4].Body.String(), `"revision":11`) {
		t.Fatalf("list response lacks identity contract: %s", responses[4].Body)
	}
	createdID := app.fields[resource][0].ID
	if _, found := app.createValues[createdID]; !found {
		t.Fatalf("public name was not mapped to stable field ID %s", createdID.String())
	}
	if _, found := app.updateValues[createdID]; !found {
		t.Fatalf("updated public name was not mapped to stable field ID %s", createdID.String())
	}
	id, revision, _ := app.updateInput.Identity()
	key, _ := app.updateInput.IdempotencyKey()
	deleteID, deleteRevision, _ := app.deleteInput.Identity()
	deleteKey, _ := app.deleteInput.IdempotencyKey()
	if id.String() != recordHTTPTestID || revision != 9 ||
		key != "security-update-"+string(resource) ||
		deleteID != id || deleteRevision != 9 ||
		deleteKey != "security-delete-"+string(resource) {
		t.Fatal("If-Match or idempotency headers were not delegated")
	}
}

func assertRecordHTTPSecurityViewFailure(
	t *testing.T,
	app *recordHTTPSecurityApplication,
) {
	t.Helper()
	app.view.Revision = 0
	app.view.Data["revision"] = []byte(`0`)
	response := recordHTTPSecurityDo(
		newRecordHTTPSecurityRoutes(t, app, nil),
		http.MethodGet, "/api/users/"+recordHTTPTestID, "", nil,
	)
	if response.Code != http.StatusServiceUnavailable ||
		response.Header().Get("ETag") != "" {
		t.Fatalf("invalid view status=%d headers=%v", response.Code, response.Header())
	}
}
