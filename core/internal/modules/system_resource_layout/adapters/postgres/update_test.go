package postgres

import (
	"context"
	"crypto/sha256"
	"testing"

	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
	layoutapp "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/application"
)

func TestUpdateMapsOptimisticConflict(t *testing.T) {
	store := fixtureLayoutStore(fixtureRow{err: fixtureStateError("40001")})
	digest := sha256.Sum256([]byte(adapterDefinition))
	_, err := store.Update(context.Background(), layoutapp.UpdateStoreInput{
		ID: adapterRevisionID, ExpectedRowRevision: 1,
		DefinitionJSON: adapterDefinition, DefinitionSHA256: digest, Now: adapterNow,
	})
	if err != layoutapi.ErrConflict {
		t.Fatalf("update error = %v, want conflict", err)
	}
}
