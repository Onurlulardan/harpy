package formrecordarchive

import (
	"errors"
	"testing"
	"time"

	formrecordarchiveapi "github.com/Onurlulardan/harpy/core/internal/modules/documents/api/formrecordarchive"
	archiveapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/archive"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func TestGetRefreshesAccessAndReturnsActorOwnedArchive(t *testing.T) {
	fixture := newApplicationFixture(t)
	itemID := fixtureDomainID("018f0000-0000-7000-8000-000000000161")
	digest, valid := archiveapp.RequestDigest(archiveapp.CreateRequest{
		ItemIDs: []domain.ID{itemID}})
	if !valid {
		t.Fatal("archive digest fixture is invalid")
	}
	archiveID := fixtureDomainID("018f0000-0000-7000-8000-000000000162")
	readyAt := time.Date(2026, 8, 2, 10, 0, 0, 0, time.UTC)
	fixture.archives.getResult = archiveapp.Snapshot{ID: archiveID,
		TenantID: fixture.tenantID, ActorID: fixture.actorID,
		JobID: fixtureDomainID("018f0000-0000-7000-8000-000000000163"),
		State: domain.ArchiveReady, RequestSHA256: digest, EntryCount: 1,
		PlaintextBytes: 11, CiphertextBytes: 29, ReadyAt: readyAt,
		ExpiresAt: readyAt.Add(24 * time.Hour)}
	result, err := fixture.application.Get(t.Context(), fixture.apiAccess(t),
		formrecordarchiveapi.GetRequest{ArchiveID: formrecordarchiveapi.ID(archiveID),
			ItemIDs:       []formrecordarchiveapi.ID{formrecordarchiveapi.ID(itemID)},
			RequestSHA256: digest})
	if err != nil || result.State != formrecordarchiveapi.StateReady ||
		result.ArchiveID != formrecordarchiveapi.ID(archiveID) ||
		fixture.access.calls != 1 || fixture.archives.getCalls != 1 {
		t.Fatalf("status query failed: result=%#v err=%v", result, err)
	}
}

func TestGetRejectsDigestSubstitutionBeforeAuthorization(t *testing.T) {
	fixture := newApplicationFixture(t)
	archiveID := fixtureDomainID("018f0000-0000-7000-8000-000000000171")
	itemID := fixtureDomainID("018f0000-0000-7000-8000-000000000172")
	_, err := fixture.application.Get(t.Context(), fixture.apiAccess(t),
		formrecordarchiveapi.GetRequest{ArchiveID: formrecordarchiveapi.ID(archiveID),
			ItemIDs:       []formrecordarchiveapi.ID{formrecordarchiveapi.ID(itemID)},
			RequestSHA256: [32]byte{9}})
	if !errors.Is(err, formrecordarchiveapi.ErrInvalid) || fixture.access.calls != 0 ||
		fixture.scopes.calls != 0 || fixture.archives.getCalls != 0 {
		t.Fatalf("digest substitution was not rejected early: err=%v", err)
	}
}
