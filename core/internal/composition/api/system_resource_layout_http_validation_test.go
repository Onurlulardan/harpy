package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPRejectsUnknownResourceQueryAndBody(t *testing.T) {
	fixture := newLayoutHTTPFixture(t)
	unknown := layoutHTTPRequest(t, http.MethodGet, "")
	unknown.SetPathValue("resource", "accounts")
	unknownResponse := httptest.NewRecorder()
	fixture.handler.getRevision(unknownResponse, unknown)
	if unknownResponse.Code != http.StatusNotFound || fixture.app.getCalls != 0 {
		t.Fatalf("unknown resource=%d calls=%d",
			unknownResponse.Code, fixture.app.getCalls)
	}
	query := layoutHTTPRequest(t, http.MethodGet, "")
	query.URL.RawQuery = "extra=true"
	queryResponse := httptest.NewRecorder()
	fixture.handler.getRevision(queryResponse, query)
	if queryResponse.Code != http.StatusBadRequest || fixture.app.getCalls != 0 {
		t.Fatalf("unknown query=%d calls=%d", queryResponse.Code, fixture.app.getCalls)
	}
	body := layoutHTTPRequest(t, http.MethodPut,
		`{"definition":{"contract":"harpy/system-resource-layout/v2",`+
			`"surfaces":[],"unknown":true}}`)
	body.Header.Set("Content-Type", "application/json")
	bodyResponse := httptest.NewRecorder()
	fixture.handler.putRevision(bodyResponse, body)
	if bodyResponse.Code != http.StatusBadRequest || fixture.app.createCalls != 0 {
		t.Fatalf("unknown body=%d calls=%d", bodyResponse.Code, fixture.app.createCalls)
	}
}

func TestHTTPRejectsUUIDAndETagAmbiguity(t *testing.T) {
	fixture := newLayoutHTTPFixture(t)
	badID := layoutHTTPRequest(t, http.MethodGet, "")
	badID.SetPathValue("revision_id", "018f47f4-7b1b-4cc0-8000-000000000033")
	badIDResponse := httptest.NewRecorder()
	fixture.handler.getRevision(badIDResponse, badID)
	if badIDResponse.Code != http.StatusBadRequest || fixture.app.getCalls != 0 {
		t.Fatalf("UUID response=%d calls=%d", badIDResponse.Code, fixture.app.getCalls)
	}
	missing := layoutHTTPRequest(t, http.MethodDelete, "")
	missingResponse := httptest.NewRecorder()
	fixture.handler.deleteRevision(missingResponse, missing)
	if missingResponse.Code != http.StatusPreconditionRequired {
		t.Fatalf("missing If-Match=%d", missingResponse.Code)
	}
	duplicate := layoutHTTPRequest(t, http.MethodDelete, "")
	duplicate.Header.Add("If-Match", `"layout-row-revision-1"`)
	duplicate.Header.Add("If-Match", `"layout-row-revision-2"`)
	duplicateResponse := httptest.NewRecorder()
	fixture.handler.deleteRevision(duplicateResponse, duplicate)
	if duplicateResponse.Code != http.StatusBadRequest || fixture.app.deleteCalls != 0 {
		t.Fatalf("duplicate If-Match=%d calls=%d",
			duplicateResponse.Code, fixture.app.deleteCalls)
	}
}
