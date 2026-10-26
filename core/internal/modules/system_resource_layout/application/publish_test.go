package application

import (
	"context"
	"testing"

	fieldregistry "github.com/Onurlulardan/harpy/core/internal/platform/fieldcatalog/fieldregistry"
	presentation "github.com/Onurlulardan/harpy/core/internal/platform/fieldcatalog/presentation"
	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
	"github.com/google/uuid"
)

func TestPublishPersistsExactCompiledPresentationEvidence(t *testing.T) {
	catalog := fixtureCatalogForIDs(t, fixtureFieldOne, fixtureFieldTwo)
	draft := fixtureRevision(t, layoutapi.RevisionDraft, 1)
	store := &fixtureStore{}
	store.get = func(context.Context, uuid.UUID) (layoutapi.Revision, error) {
		return draft, nil
	}
	var captured PublishStoreInput
	store.publish = func(
		_ context.Context,
		input PublishStoreInput,
	) (PublishedPresentation, error) {
		captured = input
		published := fixtureRevision(t, layoutapi.RevisionPublished, 1)
		return publishedFromInput(published, input), nil
	}
	result, err := fixturePresentationApplication(t, store).Publish(
		context.Background(), layoutapi.PublishPresentationInput{
			RevisionID: draft.ID, ExpectedRowRevision: 1, Catalog: catalog,
		},
	)
	if err != nil || result.State != layoutapi.RevisionPublished {
		t.Fatalf("publish result = %+v: %v", result, err)
	}
	if captured.CompilerVersion != presentationCompilerVersionV2 ||
		captured.CatalogContract != presentation.FieldCatalogContract ||
		captured.PresentationContract != presentation.ContractVersion ||
		captured.SourceDefinitionSHA256 != draft.DefinitionSHA256 ||
		captured.Now != fixtureNow {
		t.Fatalf("publish evidence = %+v", captured)
	}
}

func TestPublishRejectsIncompatibleWidgetBeforeStorage(t *testing.T) {
	profile, _ := presentation.DefaultValueProfile(fieldregistry.TypeBoolean)
	field, err := presentation.NewField(
		fixtureFieldOne.String(), fieldregistry.TypeBoolean,
		profile, presentation.FieldEditable,
	)
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := presentation.NewFieldCatalog([]presentation.Field{field})
	if err != nil {
		t.Fatal(err)
	}
	store := &fixtureStore{}
	store.get = func(context.Context, uuid.UUID) (layoutapi.Revision, error) {
		return fixtureRevision(t, layoutapi.RevisionDraft, 1), nil
	}
	called := false
	store.publish = func(
		context.Context, PublishStoreInput,
	) (PublishedPresentation, error) {
		called = true
		return PublishedPresentation{}, nil
	}
	_, err = fixturePresentationApplication(t, store).Publish(
		context.Background(), layoutapi.PublishPresentationInput{
			RevisionID: fixtureRevisionID.String(), ExpectedRowRevision: 1,
			Catalog: catalog,
		},
	)
	if err != layoutapi.ErrInvalid || called {
		t.Fatalf("incompatible publish error/called = %v/%v", err, called)
	}
}

func TestPublishReturnsExactAlreadyActiveReplay(t *testing.T) {
	catalog := fixtureCatalogForIDs(t, fixtureFieldOne, fixtureFieldTwo)
	published := fixtureRevision(t, layoutapi.RevisionPublished, 1)
	store := &fixtureStore{}
	store.get = func(context.Context, uuid.UUID) (layoutapi.Revision, error) {
		return published, nil
	}
	var calls int
	store.publish = func(
		_ context.Context, input PublishStoreInput,
	) (PublishedPresentation, error) {
		calls++
		return publishedFromInput(published, input), nil
	}
	result, err := fixturePresentationApplication(t, store).Publish(
		context.Background(), layoutapi.PublishPresentationInput{
			RevisionID: published.ID, ExpectedRowRevision: 1, Catalog: catalog,
		},
	)
	if err != nil || calls != 1 || result.ID != published.ID ||
		result.RowRevision != 2 {
		t.Fatalf("publication replay = %+v calls=%d: %v", result, calls, err)
	}
}

func publishedFromInput(
	revision layoutapi.Revision,
	input PublishStoreInput,
) PublishedPresentation {
	return PublishedPresentation{
		Revision: revision, Generation: 1,
		SourceDefinitionSHA256: input.SourceDefinitionSHA256,
		CompilerVersion:        input.CompilerVersion,
		CatalogContract:        input.CatalogContract,
		CatalogJSON:            append([]byte(nil), input.CatalogJSON...),
		CatalogSHA256:          input.CatalogSHA256,
		PresentationContract:   input.PresentationContract,
		PresentationJSON:       append([]byte(nil), input.PresentationJSON...),
		PresentationSHA256:     input.PresentationSHA256, EvidenceComplete: true,
	}
}
