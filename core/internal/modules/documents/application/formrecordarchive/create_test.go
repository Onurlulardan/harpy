package formrecordarchive

import (
	"errors"
	"strings"
	"testing"

	formrecordarchiveapi "github.com/Onurlulardan/harpy/core/internal/modules/documents/api/formrecordarchive"
)

func TestCreateFailsClosedBeforeScopeWhenAccessRefreshIsDenied(t *testing.T) {
	fixture := newApplicationFixture(t)
	fixture.access.err = errors.New("denied")
	_, err := fixture.application.Create(t.Context(), fixture.apiAccess(t),
		fixtureCreateRequest(1))
	if !errors.Is(err, formrecordarchiveapi.ErrForbidden) ||
		fixture.scopes.calls != 0 || fixture.links.calls != 0 ||
		len(fixture.archives.createRequests) != 0 {
		t.Fatalf("denial was not fail-closed: err=%v scopes=%d links=%d archives=%d",
			err, fixture.scopes.calls, fixture.links.calls,
			len(fixture.archives.createRequests))
	}
}

func TestCreateRequiresExactLinksBeforeArchiveMutation(t *testing.T) {
	fixture := newApplicationFixture(t)
	fixture.links.err = formrecordarchiveapi.ErrReferenceMismatch
	_, err := fixture.application.Create(t.Context(), fixture.apiAccess(t),
		fixtureCreateRequest(2))
	if !errors.Is(err, formrecordarchiveapi.ErrReferenceMismatch) ||
		fixture.links.calls != 1 || len(fixture.archives.createRequests) != 0 {
		t.Fatalf("link substitution reached archive mutation: err=%v", err)
	}
}

func TestCreateNamespacesIdempotencyByFormRecord(t *testing.T) {
	fixture := newApplicationFixture(t)
	request := fixtureCreateRequest(3)
	first, err := fixture.application.Create(t.Context(), fixture.apiAccess(t), request)
	if err != nil {
		t.Fatal(err)
	}
	second, err := fixture.application.Create(t.Context(), fixture.apiAccess(t), request)
	if err != nil {
		t.Fatal(err)
	}
	request.RecordID = formrecordarchiveapi.ID(
		fixtureDomainID("018f0000-0000-7000-8000-000000000153"))
	if _, err = fixture.application.Create(t.Context(), fixture.apiAccess(t), request); err != nil {
		t.Fatal(err)
	}
	keys := fixture.archives.createRequests
	if len(keys) != 3 || keys[0].IdempotencyKey != keys[1].IdempotencyKey ||
		keys[0].IdempotencyKey == keys[2].IdempotencyKey ||
		!strings.HasPrefix(keys[0].IdempotencyKey, "forms-record-archive-") ||
		first.ArchiveID != second.ArchiveID {
		t.Fatalf("idempotency namespace is not deterministic: %#v", keys)
	}
}

func fixtureCreateRequest(seed byte) formrecordarchiveapi.CreateRequest {
	return formrecordarchiveapi.CreateRequest{
		DefinitionID: formrecordarchiveapi.ID(
			fixtureDomainID("018f0000-0000-7000-8000-000000000151")),
		RecordID: formrecordarchiveapi.ID(
			fixtureDomainID("018f0000-0000-7000-8000-000000000152")),
		ItemIDs: []formrecordarchiveapi.ID{formrecordarchiveapi.ID(
			fixtureDomainID("018f0000-0000-7000-8000-000000000154"))},
		IdempotencyKey: "record-archive-idempotency-" + string([]byte{'0' + seed}),
	}
}
