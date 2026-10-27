package postgres

import (
	"context"
	"crypto/sha256"
	"testing"

	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
	layoutapp "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/application"
)

func TestCreateDecodesStoredDraft(t *testing.T) {
	digest := sha256.Sum256([]byte(adapterDefinition))
	store := fixtureLayoutStore(fixtureRow{
		document: fixtureRevisionJSON(t, layoutapi.RevisionDraft),
	})
	revision, err := store.Create(context.Background(), layoutapp.CreateStoreInput{
		ID: adapterRevisionID, Resource: layoutapi.ResourceUser,
		DefinitionJSON: adapterDefinition, DefinitionSHA256: digest, Now: adapterNow,
	})
	if err != nil || revision.ID != adapterRevisionID.String() ||
		revision.State != layoutapi.RevisionDraft {
		t.Fatalf("create revision = %+v: %v", revision, err)
	}
}
