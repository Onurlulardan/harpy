package application

import (
	"context"
	"testing"

	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
)

func TestUpdateCarriesOptimisticRevisionAndDigest(t *testing.T) {
	var captured UpdateStoreInput
	store := &fixtureStore{}
	store.update = func(_ context.Context, input UpdateStoreInput) (layoutapi.Revision, error) {
		captured = input
		revision := fixtureRevision(t, layoutapi.RevisionDraft, 1)
		revision.RowRevision = input.ExpectedRowRevision + 1
		return revision, nil
	}
	_, err := fixtureApplication(t, store).Update(context.Background(),
		layoutapi.UpdateInput{ID: fixtureRevisionID.String(), ExpectedRowRevision: 7,
			Definition: fixtureDefinition()})
	if err != nil || captured.ExpectedRowRevision != 7 ||
		captured.DefinitionJSON == "" || captured.DefinitionSHA256 == [32]byte{} {
		t.Fatalf("update input = %+v: %v", captured, err)
	}
}

func TestUpdateRejectsStoreRevisionDrift(t *testing.T) {
	store := &fixtureStore{}
	store.update = func(
		_ context.Context, _ UpdateStoreInput,
	) (layoutapi.Revision, error) {
		return fixtureRevision(t, layoutapi.RevisionDraft, 1), nil
	}
	_, err := fixtureApplication(t, store).Update(context.Background(),
		layoutapi.UpdateInput{ID: fixtureRevisionID.String(), ExpectedRowRevision: 7,
			Definition: fixtureDefinition()})
	if err != layoutapi.ErrUnavailable {
		t.Fatalf("update revision drift error = %v", err)
	}
}
