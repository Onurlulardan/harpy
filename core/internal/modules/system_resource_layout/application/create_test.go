package application

import (
	"context"
	"testing"

	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
)

func TestCreateCanonicalizesDraftBeforeStorage(t *testing.T) {
	var captured CreateStoreInput
	store := &fixtureStore{}
	store.create = func(_ context.Context, input CreateStoreInput) (layoutapi.Revision, error) {
		captured = input
		return fixtureRevision(t, layoutapi.RevisionDraft, 1), nil
	}
	revision, err := fixtureApplication(t, store).Create(context.Background(),
		layoutapi.CreateInput{Resource: layoutapi.ResourceUser,
			Definition: fixtureDefinition()})
	if err != nil || revision.State != layoutapi.RevisionDraft {
		t.Fatalf("create draft = %v: %v", revision.State, err)
	}
	if captured.ID != fixtureRevisionID || captured.Resource != layoutapi.ResourceUser ||
		captured.Now != fixtureNow || captured.DefinitionJSON == "" ||
		captured.DefinitionSHA256 == [32]byte{} {
		t.Fatalf("captured create input is incomplete: %+v", captured)
	}
}

func TestCreateRejectsStoreIdentityDrift(t *testing.T) {
	store := &fixtureStore{}
	store.create = func(
		_ context.Context, _ CreateStoreInput,
	) (layoutapi.Revision, error) {
		revision := fixtureRevision(t, layoutapi.RevisionDraft, 1)
		revision.ID = fixtureFieldOne.String()
		return revision, nil
	}
	_, err := fixtureApplication(t, store).Create(context.Background(),
		layoutapi.CreateInput{Resource: layoutapi.ResourceUser,
			Definition: fixtureDefinition()})
	if err != layoutapi.ErrUnavailable {
		t.Fatalf("create identity drift error = %v", err)
	}
}

func TestCreateUsesClientSuppliedUUIDv7AndAcceptsExactReplay(t *testing.T) {
	var captured CreateStoreInput
	store := &fixtureStore{}
	store.create = func(
		_ context.Context, input CreateStoreInput,
	) (layoutapi.Revision, error) {
		captured = input
		return fixtureRevision(t, layoutapi.RevisionPublished, 1), nil
	}
	revision, err := fixtureApplication(t, store).Create(
		context.Background(), layoutapi.CreateInput{
			ID: fixtureRevisionID.String(), Resource: layoutapi.ResourceUser,
			Definition: fixtureDefinition(),
		},
	)
	if err != nil || captured.ID != fixtureRevisionID ||
		revision.State != layoutapi.RevisionPublished || revision.RowRevision != 2 {
		t.Fatalf("create replay = %+v captured=%+v: %v", revision, captured, err)
	}
	_, err = fixtureApplication(t, store).Create(
		context.Background(), layoutapi.CreateInput{
			ID:       "11111111-1111-4111-8111-111111111111",
			Resource: layoutapi.ResourceUser, Definition: fixtureDefinition(),
		},
	)
	if err != layoutapi.ErrInvalid {
		t.Fatalf("non-v7 client identity error = %v", err)
	}
}
