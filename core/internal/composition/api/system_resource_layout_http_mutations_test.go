package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPPutRevisionUsesClientIdentityAndCanonicalHeaders(t *testing.T) {
	fixture := newLayoutHTTPFixture(t)
	request := layoutHTTPRequest(t, http.MethodPut, layoutHTTPDefinitionBody(t))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	fixture.handler.putRevision(response, request)
	if response.Code != http.StatusOK || fixture.app.createCalls != 1 ||
		fixture.app.lastID != layoutHTTPFixtureID ||
		response.Header().Get("ETag") != `"layout-row-revision-2"` ||
		response.Header().Get("Location") != layoutHTTPRevisionLocation(
			"users", layoutHTTPFixtureID,
		) || response.Header().Get("X-Correlation-ID") != "layout-request-1" {
		t.Fatalf("put response=%d headers=%v app=%+v body=%s",
			response.Code, response.Header(), fixture.app, response.Body.String())
	}
	assertLayoutHTTPAccess(t, fixture, layoutHTTPActionUpdate,
		"system_resource_layout.create")
}

func TestHTTPPatchAndDeleteRequireExactRowETag(t *testing.T) {
	fixture := newLayoutHTTPFixture(t)
	patch := layoutHTTPRequest(t, http.MethodPatch, layoutHTTPDefinitionBody(t))
	patch.Header.Set("Content-Type", "application/json")
	patch.Header.Set("If-Match", `"layout-row-revision-1"`)
	patchResponse := httptest.NewRecorder()
	fixture.handler.patchRevision(patchResponse, patch)
	if patchResponse.Code != http.StatusOK || fixture.app.updateCalls != 1 ||
		fixture.app.lastExpected != 1 {
		t.Fatalf("patch=%d app=%+v", patchResponse.Code, fixture.app)
	}
	deleted := layoutHTTPRequest(t, http.MethodDelete, "")
	deleted.Header.Set("If-Match", `"layout-row-revision-2"`)
	deleteResponse := httptest.NewRecorder()
	fixture.handler.deleteRevision(deleteResponse, deleted)
	if deleteResponse.Code != http.StatusOK || fixture.app.deleteCalls != 1 ||
		fixture.app.lastExpected != 2 || !strings.Contains(
		deleteResponse.Body.String(), `"deleted":true`,
	) {
		t.Fatalf("delete=%d app=%+v body=%s",
			deleteResponse.Code, fixture.app, deleteResponse.Body.String())
	}
}

func TestHTTPPublicationPUTUsesOriginalPreconditionAndLocation(t *testing.T) {
	fixture := newLayoutHTTPFixture(t)
	request := layoutHTTPRequest(t, http.MethodPut, "")
	request.Header.Set("If-Match", `"layout-row-revision-1"`)
	response := httptest.NewRecorder()
	fixture.handler.putPublication(response, request)
	if response.Code != http.StatusOK || fixture.app.publishCalls != 1 ||
		fixture.app.lastExpected != 1 || response.Header().Get("Location") !=
		layoutHTTPPublicationLocation("users", layoutHTTPFixtureID) {
		t.Fatalf("publication=%d app=%+v headers=%v",
			response.Code, fixture.app, response.Header())
	}
	assertLayoutHTTPAccess(t, fixture, layoutHTTPActionUpdate,
		"system_resource_layout.publish")
}

func assertLayoutHTTPAccess(
	t *testing.T, fixture layoutHTTPFixture, action, operation string,
) {
	t.Helper()
	request := fixture.access.request
	if request.Resource != "users" || request.Action != action ||
		request.Operation != operation || len(request.AuthorizationValues) != 1 ||
		len(request.TenantIDValues) != 1 {
		t.Fatalf("access request = %+v", request)
	}
}
