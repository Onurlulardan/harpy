package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	systemresources "github.com/Onurlulardan/harpy/core/internal/orchestration/systemresources"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
)

func TestSystemResourceRecordHTTPRoutesAreExactCanonicalCRUDInventory(t *testing.T) {
	fixture := newRecordHTTPFixture(t)
	routes := fixture.handler.Routes()
	expected := identitySystemResourceCRUDRouteKeys()
	seen := make(map[string]struct{}, len(routes))
	for _, route := range routes {
		key := route.Method + "\x00" + route.Pattern
		if _, found := expected[key]; !found || route.Handler == nil {
			t.Fatalf("unexpected route: %+v", route)
		}
		if _, duplicate := seen[key]; duplicate {
			t.Fatalf("duplicate route: %s %s", route.Method, route.Pattern)
		}
		seen[key] = struct{}{}
	}
	if len(routes) != 15 || len(seen) != len(expected) {
		t.Fatalf("routes=%d unique=%d expected=%d", len(routes), len(seen), len(expected))
	}
}

func TestSystemResourceRecordHTTPDeleteRequiresAndForwardsExactHeaders(t *testing.T) {
	fixture := newRecordHTTPFixture(t)
	request := recordHTTPRequest(http.MethodDelete, "/api/users/"+recordHTTPTestID)
	request.Header.Set("Idempotency-Key", "delete-user-fixture-0001")
	request.Header.Set("If-Match", `"revision-9"`)
	response := httptest.NewRecorder()

	fixture.handler.delete(systemresources.Users).ServeHTTP(response, request)

	id, revision, err := fixture.app.deleted.Identity()
	key, keyErr := fixture.app.deleted.IdempotencyKey()
	if response.Code != http.StatusOK || fixture.app.deleteCalls != 1 ||
		err != nil || keyErr != nil || id.String() != recordHTTPTestID ||
		revision != 9 || key != "delete-user-fixture-0001" ||
		fixture.access.request.Operation != "iam.identity.user.delete" {
		t.Fatalf("status=%d calls=%d id=%s revision=%d key=%q access=%+v err=%v/%v",
			response.Code, fixture.app.deleteCalls, id.String(), revision, key,
			fixture.access.request, err, keyErr)
	}
	var envelope systemResourceRecordHTTPDeleteEnvelope
	if json.Unmarshal(response.Body.Bytes(), &envelope) != nil ||
		!envelope.Success || envelope.DeletedCount != 1 {
		t.Fatalf("unexpected response: %s", response.Body.String())
	}
}

func TestSystemResourceRecordHTTPUpdateRejectsMissingRevisionBeforeApplication(t *testing.T) {
	fixture := newRecordHTTPFixture(t)
	request := recordHTTPRequest(http.MethodPut, "/api/users/"+recordHTTPTestID)
	request.Header.Set("Idempotency-Key", "update-user-fixture-0001")
	response := httptest.NewRecorder()

	fixture.handler.update(systemresources.Users).ServeHTTP(response, request)

	if response.Code != http.StatusPreconditionRequired {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	var envelope systemResourceRecordHTTPErrorEnvelope
	if json.Unmarshal(response.Body.Bytes(), &envelope) != nil ||
		envelope.Success || envelope.Error.Code != "PRECONDITION_REQUIRED" {
		t.Fatalf("unexpected response: %s", response.Body.String())
	}
}

func TestMintSystemResourceRecordFieldValuesUsesPublicNamesAndEditableCatalog(t *testing.T) {
	emailID, _ := systemresources.ParseID(
		"9c65f0c0-1000-4000-8000-000000000002",
	)
	statusID, _ := systemresources.ParseID(
		"9c65f0c0-1000-4000-8000-000000000003",
	)
	document := map[string]json.RawMessage{
		"email":         json.RawMessage(`"admin@example.com"`),
		"email_display": json.RawMessage(`"Admin"`),
		"status":        json.RawMessage(`"active"`),
	}
	values, err := mintSystemResourceRecordFieldValues(
		document,
		map[string]systemresources.ID{"email": emailID, "status": statusID},
		map[systemresources.ID]bool{emailID: true, statusID: false},
	)
	raw, rawErr := values.Values()
	if err != nil || rawErr != nil || len(raw) != 1 ||
		string(raw[emailID]) != `"admin@example.com"` {
		t.Fatalf("values=%v err=%v/%v", raw, err, rawErr)
	}
	_, err = mintSystemResourceRecordFieldValues(
		map[string]json.RawMessage{"status": json.RawMessage(`"active"`)},
		map[string]systemresources.ID{"status": statusID},
		map[systemresources.ID]bool{statusID: false},
	)
	if !errors.Is(err, systemresources.ErrContract) {
		t.Fatalf("expected contract failure, got %v", err)
	}
}

func TestRebindIdentitySystemResourceRoutesRejectsCanonicalDuplicates(t *testing.T) {
	fixture := newRecordHTTPFixture(t)
	direct := make([]httpedge.Route, 0, 15)
	for _, route := range fixture.handler.Routes() {
		direct = append(direct, route)
	}
	canonical := fixture.handler.Routes()
	canonical[len(canonical)-1] = canonical[0]
	if _, err := rebindIdentitySystemResourceRoutes(direct, canonical); err == nil {
		t.Fatal("expected duplicate canonical route rejection")
	}
}
