package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPGetAndListReturnExactRevisionViews(t *testing.T) {
	fixture := newLayoutHTTPFixture(t)
	get := layoutHTTPRequest(t, http.MethodGet, "")
	getResponse := httptest.NewRecorder()
	fixture.handler.getRevision(getResponse, get)
	if getResponse.Code != http.StatusOK || fixture.app.getCalls != 1 ||
		getResponse.Header().Get("ETag") != `"layout-row-revision-2"` ||
		!strings.Contains(getResponse.Body.String(), `"contract":"harpy/system-resource-layout/v2"`) {
		t.Fatalf("get=%d headers=%v body=%s",
			getResponse.Code, getResponse.Header(), getResponse.Body.String())
	}
	list := layoutHTTPRequest(t, http.MethodGet, "")
	list.SetPathValue("revision_id", "")
	listResponse := httptest.NewRecorder()
	fixture.handler.listRevisions(listResponse, list)
	if listResponse.Code != http.StatusOK || fixture.app.listCalls != 1 ||
		!strings.Contains(listResponse.Body.String(), `"items":[`) {
		t.Fatalf("list=%d body=%s", listResponse.Code, listResponse.Body.String())
	}
	assertLayoutHTTPAccess(t, fixture, layoutHTTPActionRead,
		"system_resource_layout.list")
}

func TestHTTPPresentationResolveUsesTypedIAMAdmission(t *testing.T) {
	fixture := newLayoutHTTPFixture(t)
	request := layoutHTTPRequest(t, http.MethodGet, "")
	request.SetPathValue("revision_id", "")
	response := httptest.NewRecorder()
	fixture.handler.getPresentation(response, request)
	if response.Code != http.StatusOK || fixture.app.resolveCalls != 1 ||
		!strings.HasPrefix(response.Header().Get("ETag"), `"presentation-`) ||
		!strings.Contains(response.Body.String(),
			`"registry_contract":"forms-presentation-registry.v1"`) {
		t.Fatalf("resolve=%d headers=%v body=%s",
			response.Code, response.Header(), response.Body.String())
	}
	assertLayoutHTTPAccess(t, fixture, layoutHTTPActionRead,
		"system_resource_layout.resolve")
}
