package iam

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/uuid"
)

func TestIdentityCollectionDeleteRequiresBoundedRevisionMap(t *testing.T) {
	fixture := newIdentityHTTPTestFixture(t)
	first := uuid.MustParse("83000000-0000-4000-8000-000000000002")
	second := uuid.MustParse("83000000-0000-4000-8000-000000000001")
	fixture.app.results["DeleteUsers"] = 2
	body := `{"ids":["` + first.String() + `","` + second.String() +
		`"],"expected_revisions":{"` + first.String() + `":4,"` +
		second.String() + `":3}}`
	request := fixture.request(http.MethodDelete, "/api/users", body)
	request.Header.Set("Idempotency-Key", identityHTTPTestIdempotencyKey)
	response := fixture.serve(request)
	if response.Code != http.StatusOK || len(fixture.app.calls) != 1 {
		t.Fatalf("status=%d calls=%d body=%s", response.Code,
			len(fixture.app.calls), response.Body.String())
	}
	input, ok := fixture.app.calls[0].input.(IdentityBulkDelete)
	if !ok || len(input.Items) != 2 || input.Items[0].ID != second ||
		input.Items[0].ExpectedRevision != 3 || input.Items[1].ID != first {
		t.Fatalf("bulk input=%#v", fixture.app.calls[0].input)
	}
	var envelope identityDeleteEnvelope
	if json.Unmarshal(response.Body.Bytes(), &envelope) != nil ||
		envelope.DeletedCount != 2 {
		t.Fatalf("delete response=%s", response.Body.String())
	}
}

func TestIdentityCollectionDeleteRejectsUnversionedIDs(t *testing.T) {
	fixture := newIdentityHTTPTestFixture(t)
	id := uuid.MustParse("83000000-0000-4000-8000-000000000001")
	request := fixture.request(
		http.MethodDelete, "/api/roles", `{"ids":["`+id.String()+`"]}`,
	)
	request.Header.Set("Idempotency-Key", identityHTTPTestIdempotencyKey)
	response := fixture.serve(request)
	if response.Code != http.StatusBadRequest || len(fixture.app.calls) != 0 {
		t.Fatalf("status=%d calls=%d", response.Code, len(fixture.app.calls))
	}
}
