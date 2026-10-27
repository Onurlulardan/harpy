package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	systemresources "github.com/Onurlulardan/harpy/core/internal/orchestration/systemresources"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
)

func TestSystemResourceRecordHTTPSecurityAllResourcesDelegateCanonicalCRUD(t *testing.T) {
	for _, current := range []struct {
		resource systemresources.Resource
		slug     string
		password string
	}{
		{systemresources.Users, "users", "create-only-secret"},
		{systemresources.Roles, "roles", ""},
		{systemresources.Organizations, "organizations", ""},
	} {
		t.Run(current.slug, func(t *testing.T) {
			app := newRecordHTTPSecurityApplication(t)
			routes := newRecordHTTPSecurityRoutes(t, app, nil)
			body := `{"name":"visible"}`
			if current.password != "" {
				body = `{"name":"visible","password":"` + current.password + `"}`
			}
			create := recordHTTPSecurityDo(routes, http.MethodPost, "/api/"+current.slug,
				body, map[string]string{"Idempotency-Key": "security-create-" + current.slug})
			get := recordHTTPSecurityDo(routes, http.MethodGet,
				"/api/"+current.slug+"/"+recordHTTPTestID, "", nil)
			headers := map[string]string{
				"Idempotency-Key": "security-update-" + current.slug,
				"If-Match":        `"revision-9"`,
			}
			update := recordHTTPSecurityDo(routes, http.MethodPut,
				"/api/"+current.slug+"/"+recordHTTPTestID, `{"name":"changed"}`, headers)
			headers["Idempotency-Key"] = "security-delete-" + current.slug
			deleted := recordHTTPSecurityDo(routes, http.MethodDelete,
				"/api/"+current.slug+"/"+recordHTTPTestID, "", headers)
			list := recordHTTPSecurityDo(
				routes, http.MethodGet, "/api/"+current.slug, "", nil,
			)
			assertRecordHTTPSecurityFlow(
				t, app, current.resource, current.password,
				[]*httptest.ResponseRecorder{create, get, update, deleted, list},
			)
		})
	}
}

func TestSystemResourceRecordHTTPSecurityRejectsInvalidInputs(t *testing.T) {
	tests := []struct {
		name, method, path, body string
		headers                  map[string]string
		status                   int
		code                     string
	}{
		{"duplicate", http.MethodPost, "/api/users",
			`{"name":"one","name":"two","password":"secret"}`,
			map[string]string{"Idempotency-Key": "security-duplicate-users"},
			400, "REQUEST_INVALID"},
		{"unknown", http.MethodPost, "/api/users",
			`{"name":"one","unknown":true,"password":"secret"}`,
			map[string]string{"Idempotency-Key": "security-unknown-users"},
			422, "SYSTEM_RESOURCE_INPUT_INVALID"},
		{"password-update", http.MethodPut, "/api/users/" + recordHTTPTestID,
			`{"name":"one","password":"secret"}`, map[string]string{
				"Idempotency-Key": "security-password-update", "If-Match": `"revision-9"`,
			}, 422, "SYSTEM_RESOURCE_INPUT_INVALID"},
		{"role-password", http.MethodPost, "/api/roles",
			`{"name":"one","password":"secret"}`,
			map[string]string{"Idempotency-Key": "security-role-password"},
			422, "SYSTEM_RESOURCE_INPUT_INVALID"},
		{"organization-password", http.MethodPost, "/api/organizations",
			`{"name":"one","password":"secret"}`,
			map[string]string{"Idempotency-Key": "security-organization-password"},
			422, "SYSTEM_RESOURCE_INPUT_INVALID"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			app := newRecordHTTPSecurityApplication(t)
			response := recordHTTPSecurityDo(
				newRecordHTTPSecurityRoutes(t, app, nil),
				test.method, test.path, test.body, test.headers,
			)
			if response.Code != test.status ||
				!strings.Contains(response.Body.String(), `"code":"`+test.code+`"`) ||
				len(app.calls) != 0 {
				t.Fatalf("status=%d calls=%v body=%s", response.Code, app.calls, response.Body)
			}
		})
	}
}

func TestReboundSystemResourceGETRejectsLegacyV1QueryGrammar(t *testing.T) {
	app := newRecordHTTPSecurityApplication(t)
	handler, err := newSystemResourceRecordHTTPHandler(
		systemResourceRecordHTTPDependencies{
			Access:      newRecordHTTPFixture(t).access,
			Application: app, Report: func(error) {},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	canonical := handler.Routes()
	directCalls := 0
	direct := make([]httpedge.Route, len(canonical))
	for index, route := range canonical {
		direct[index] = route
		direct[index].Handler = http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			directCalls++
		})
	}
	preserved, err := rebindIdentitySystemResourceRoutes(direct, canonical)
	if err != nil {
		t.Fatal(err)
	}
	routes, err := httpedge.NewRoutes(
		func(error) {}, append(preserved, canonical...)...,
	)
	if err != nil {
		t.Fatal(err)
	}
	expected := []systemResourceRecordHTTPIssue{
		{Path: "/limit", Code: "MEMBER_UNKNOWN"},
		{Path: "/page", Code: "VALUE_NOT_CANONICAL"},
		{Path: "/sort/field", Code: "INDEX_INVALID"},
		{Path: "/sort/order", Code: "INDEX_INVALID"},
	}
	for _, resource := range []string{"users", "roles", "organizations"} {
		response := recordHTTPSecurityDo(
			routes, http.MethodGet,
			"/api/"+resource+"?page=1&limit=100&sort%5Bfield%5D=name&sort%5Border%5D=asc",
			"", nil,
		)
		var envelope systemResourceRecordHTTPErrorEnvelope
		if json.Unmarshal(response.Body.Bytes(), &envelope) != nil ||
			response.Code != http.StatusUnprocessableEntity ||
			envelope.Error.Code != "QUERY_INVALID" ||
			len(envelope.Error.Issues) != len(expected) {
			t.Fatalf("%s status=%d body=%s", resource, response.Code, response.Body)
		}
		for index, issue := range envelope.Error.Issues {
			if issue != expected[index] {
				t.Fatalf("%s issue[%d]=%+v, want %+v",
					resource, index, issue, expected[index])
			}
		}
	}
	if directCalls != 0 || len(app.calls) != 0 {
		t.Fatalf("legacy query reached displaced or application handler: direct=%d app=%v",
			directCalls, app.calls)
	}
}

func TestSystemResourceRecordHTTPSecurityMapsAdmissionAndInvalidViewFailures(t *testing.T) {
	for _, test := range []struct {
		err    error
		status int
		code   string
	}{
		{systemresources.ErrLayoutAccessAuthentication, 401, "AUTHENTICATION_FAILED"},
		{systemresources.ErrLayoutAccessForbidden, 403, "AUTHORIZATION_DENIED"},
		{systemresources.ErrLayoutAccessUnavailable, 503, "IAM_AUTHORITY_UNAVAILABLE"},
	} {
		app := newRecordHTTPSecurityApplication(t)
		routes := newRecordHTTPSecurityRoutes(
			t, app, recordHTTPSecurityAccessError{err: test.err},
		)
		response := recordHTTPSecurityDo(routes, http.MethodGet, "/api/users", "", nil)
		if response.Code != test.status ||
			!strings.Contains(response.Body.String(), `"code":"`+test.code+`"`) ||
			len(app.calls) != 0 {
			t.Fatalf("admission status=%d calls=%v body=%s", response.Code, app.calls, response.Body)
		}
	}
	assertRecordHTTPSecurityViewFailure(t, newRecordHTTPSecurityApplication(t))
}
