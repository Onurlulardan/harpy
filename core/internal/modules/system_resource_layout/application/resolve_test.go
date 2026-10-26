package application

import (
	"context"
	"testing"

	presentation "github.com/Onurlulardan/harpy/core/internal/platform/fieldcatalog/presentation"
	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
)

func TestResolveBuildsDeterministicDefaultOnlyWhenPublicationIsAbsent(t *testing.T) {
	catalog := fixtureCatalogForIDs(t, fixtureFieldTwo, fixtureFieldOne)
	store := &fixtureStore{}
	store.resolve = func(
		context.Context, layoutapi.ResourceKind,
	) (PublishedPresentation, error) {
		return PublishedPresentation{}, layoutapi.ErrNotFound
	}
	application := fixturePresentationApplication(t, store)
	first, err := application.Resolve(context.Background(), layoutapi.ResourceUser, catalog)
	if err != nil {
		t.Fatal(err)
	}
	second, err := application.Resolve(context.Background(), layoutapi.ResourceUser, catalog)
	if err != nil {
		t.Fatal(err)
	}
	firstJSON, firstSHA, err := presentation.EncodeDocument(first, catalog)
	if err != nil {
		t.Fatal(err)
	}
	secondJSON, secondSHA, err := presentation.EncodeDocument(second, catalog)
	if err != nil || !sameBytes(firstJSON, secondJSON) || firstSHA != secondSHA {
		t.Fatalf("default presentation is not deterministic: %v", err)
	}
}

func TestResolveFailsClosedOnCatalogDrift(t *testing.T) {
	publishedCatalog := fixtureCatalogForIDs(t, fixtureFieldOne, fixtureFieldTwo)
	currentCatalog := fixtureCatalogForIDs(
		t, fixtureFieldOne, fixtureFieldTwo, fixtureFieldThree,
	)
	published := fixturePublishedPresentation(
		t, fixtureRevision(t, layoutapi.RevisionPublished, 1), publishedCatalog,
	)
	store := &fixtureStore{}
	store.resolve = func(
		context.Context, layoutapi.ResourceKind,
	) (PublishedPresentation, error) {
		return published, nil
	}
	_, err := fixturePresentationApplication(t, store).Resolve(
		context.Background(), layoutapi.ResourceUser, currentCatalog,
	)
	if err != layoutapi.ErrUnavailable {
		t.Fatalf("catalog drift error = %v", err)
	}
}

func TestResolveFailsClosedOnIncompleteStoredEvidence(t *testing.T) {
	catalog := fixtureCatalogForIDs(t, fixtureFieldOne, fixtureFieldTwo)
	published := fixturePublishedPresentation(
		t, fixtureRevision(t, layoutapi.RevisionPublished, 1), catalog,
	)
	published.EvidenceComplete = false
	store := &fixtureStore{}
	store.resolve = func(
		context.Context, layoutapi.ResourceKind,
	) (PublishedPresentation, error) {
		return published, nil
	}
	_, err := fixturePresentationApplication(t, store).Resolve(
		context.Background(), layoutapi.ResourceUser, catalog,
	)
	if err != layoutapi.ErrUnavailable {
		t.Fatalf("incomplete evidence error = %v", err)
	}
}
