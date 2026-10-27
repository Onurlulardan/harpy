package iam

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/systemresource"
)

func TestIdentityHTTPGETAndPOSTUseOneRESTQueryApplication(t *testing.T) {
	resources := []struct {
		name, path string
		resource   IdentityResource
	}{
		{"users", "/api/users", IdentityUsers},
		{"roles", "/api/roles", IdentityRoles},
		{"organizations", "/api/organizations", IdentityOrganizations},
	}
	transports := []struct{ method, suffix, body, transport string }{
		{http.MethodGet, "?page%5Bsize%5D=25", "", "GET"},
		{http.MethodPost, "/query", `{"select":[],"where":null,"sort":[],` +
			`"page":{"size":25,"after":null},"search":null,"count":"none",` +
			`"include":[],"aggregate":null}`, "POST"},
	}
	for _, resource := range resources {
		for _, transport := range transports {
			t.Run(resource.name+"/"+transport.transport, func(t *testing.T) {
				fixture := newIdentityHTTPTestFixture(t)
				count := int64(1)
				fixture.app.results["Query"] = IdentityQueryView{
					Items:   []json.RawMessage{json.RawMessage(`{"id":"70000000-0000-4000-8000-000000000001"}`)},
					HasMore: true, NextCursor: "opaque", NextTransport: "POST", ExactCount: &count,
				}
				response := fixture.serve(fixture.request(
					transport.method, resource.path+transport.suffix, transport.body,
				))
				if response.Code != http.StatusOK || len(fixture.app.calls) != 1 {
					t.Fatalf("status=%d calls=%d body=%s", response.Code,
						len(fixture.app.calls), response.Body.String())
				}
				input := fixture.app.calls[0].input.(identityHTTPTestQuery)
				if input.resource != resource.resource || input.request.Transport() != transport.transport {
					t.Fatalf("resource=%s transport=%s", input.resource, input.request.Transport())
				}
				var envelope identityQueryEnvelope
				if json.Unmarshal(response.Body.Bytes(), &envelope) != nil ||
					!envelope.Data.Pagination.HasMore || envelope.Data.Pagination.NextCursor != "opaque" ||
					envelope.Data.Pagination.ExactCount == nil || *envelope.Data.Pagination.ExactCount != 1 {
					t.Fatalf("query envelope=%s", response.Body.String())
				}
			})
		}
	}
}

func TestIdentityHTTPRESTQueryErrorsKeepPublishedApplicationCode(t *testing.T) {
	_, failure := queryapp.DecodePOST([]byte(`{"unknown":true}`))
	if failure == nil {
		t.Fatal("invalid query fixture did not produce a published failure")
	}
	fixture := newIdentityHTTPTestFixture(t)
	fixture.app.failures["Query"] = failure
	response := fixture.serve(fixture.request(http.MethodGet, "/api/users", ""))
	if response.Code != http.StatusUnprocessableEntity || len(fixture.app.calls) != 1 ||
		!responseHasIdentityErrorCode(response.Body.Bytes(), "QUERY_INVALID") {
		t.Fatalf("status=%d calls=%d body=%s", response.Code,
			len(fixture.app.calls), response.Body.String())
	}
}

func TestIdentityHTTPRejectsUnknownQueryBeforeApplication(t *testing.T) {
	fixture := newIdentityHTTPTestFixture(t)
	request := fixture.request(http.MethodPost, "/api/users/query", `{"unknown":true}`)
	response := fixture.serve(request)
	if response.Code != http.StatusUnprocessableEntity || len(fixture.app.calls) != 0 ||
		!responseHasIdentityErrorCode(response.Body.Bytes(), "QUERY_INVALID") {
		t.Fatalf("status=%d calls=%d body=%s", response.Code,
			len(fixture.app.calls), response.Body.String())
	}
}

func TestIdentityHTTPNeverReturnsQuerySecrets(t *testing.T) {
	fixture := newIdentityHTTPTestFixture(t)
	fixture.app.results["Query"] = IdentityQueryView{Items: []json.RawMessage{
		json.RawMessage(`{"id":"70000000-0000-4000-8000-000000000001","password_hash":"secret"}`),
	}}
	response := fixture.serve(fixture.request(http.MethodGet, "/api/users", ""))
	if response.Code != http.StatusServiceUnavailable || len(fixture.errors) != 1 ||
		!responseHasIdentityErrorCode(response.Body.Bytes(), "IAM_UNAVAILABLE") {
		t.Fatalf("status=%d errors=%d body=%s", response.Code,
			len(fixture.errors), response.Body.String())
	}
	if strings.Contains(response.Body.String(), "password_hash") ||
		strings.Contains(response.Body.String(), "secret") {
		t.Fatal("secret key escaped into query response")
	}
}
