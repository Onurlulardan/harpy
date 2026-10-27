package postgres

import (
	"context"
	"testing"

	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
)

func TestResolveMapsMissingPublication(t *testing.T) {
	store := fixtureLayoutStore(fixtureRow{err: fixtureStateError("P0002")})
	_, err := store.Resolve(context.Background(), layoutapi.ResourceUser)
	if err != layoutapi.ErrNotFound {
		t.Fatalf("resolve error = %v, want not found", err)
	}
}

func TestResolvePreservesIncompleteLegacyEvidenceForFailClosedApplication(t *testing.T) {
	store := fixtureLayoutStore(fixtureRow{document: fixturePublicationJSON(t, false)})
	published, err := store.Resolve(context.Background(), layoutapi.ResourceUser)
	if err != nil || published.EvidenceComplete || published.Generation != 1 {
		t.Fatalf("legacy publication = %+v: %v", published, err)
	}
}
