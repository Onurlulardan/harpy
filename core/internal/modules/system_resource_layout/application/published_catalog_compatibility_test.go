package application

import (
	"crypto/sha256"
	"testing"

	fieldregistry "github.com/Onurlulardan/harpy/core/internal/platform/fieldcatalog/fieldregistry"
	presentation "github.com/Onurlulardan/harpy/core/internal/platform/fieldcatalog/presentation"
	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
)

func TestPublishedVerificationAllowsGenerationIndependentOwnerExpansion(
	t *testing.T,
) {
	revision, published, current := organizationOwnerExpansionFixture(t)
	for _, generation := range []int64{1, 2, 97} {
		published.Generation = generation
		if _, err := verifyPublishedPresentation(
			published, revision.Resource, current,
		); err != nil {
			t.Fatalf("generation %d predecessor rejected: %v", generation, err)
		}
	}
}

func TestPublishedVerificationRejectsCatalogDriftBeyondExactOwnerExpansion(
	t *testing.T,
) {
	revision, published, current := organizationOwnerExpansionFixture(t)
	t.Run("tampered catalog", func(t *testing.T) {
		tampered := published
		tampered.CatalogJSON = append(
			append([]byte(nil), tampered.CatalogJSON...), ' ',
		)
		tampered.CatalogSHA256 = sha256.Sum256(tampered.CatalogJSON)
		if _, err := verifyPublishedPresentation(
			tampered, revision.Resource, current,
		); err == nil {
			t.Fatal("tampered predecessor catalog was accepted")
		}
	})
	t.Run("wrong owner contract", func(t *testing.T) {
		fields, _ := fixtureCatalogForIDs(
			t, fixtureFieldOne, fixtureFieldTwo,
		).Fields()
		wrongOwner, err := presentation.NewField(
			organizationOwnerIDs, fieldregistry.TypeText,
			fieldregistry.ValueTextNFCV1, presentation.FieldEditable,
		)
		if err != nil {
			t.Fatal(err)
		}
		wrongCurrent, err := presentation.NewFieldCatalog(append(fields, wrongOwner))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := verifyPublishedPresentation(
			published, revision.Resource, wrongCurrent,
		); err == nil {
			t.Fatal("wrong owner contract was accepted")
		}
	})
	t.Run("another missing field", func(t *testing.T) {
		fields, _ := fixtureCatalogForIDs(
			t, fixtureFieldOne, fixtureFieldTwo, fixtureFieldThree,
		).Fields()
		owner := organizationOwnerFieldFixture(t)
		expanded, err := presentation.NewFieldCatalog(append(fields, owner))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := verifyPublishedPresentation(
			published, revision.Resource, expanded,
		); err == nil {
			t.Fatal("generic catalog subset was accepted")
		}
	})
	t.Run("other resource", func(t *testing.T) {
		if _, err := verifyPublishedPresentation(
			published, layoutapi.ResourceRole, current,
		); err == nil {
			t.Fatal("non-organization expansion was accepted")
		}
	})
}

func organizationOwnerExpansionFixture(
	t *testing.T,
) (layoutapi.Revision, PublishedPresentation, presentation.FieldCatalog) {
	t.Helper()
	legacy := fixtureCatalogForIDs(t, fixtureFieldOne, fixtureFieldTwo)
	revision := fixtureRevision(t, layoutapi.RevisionPublished, 1)
	revision.Resource = layoutapi.ResourceOrganization
	published := fixturePublishedPresentation(t, revision, legacy)
	fields, err := legacy.Fields()
	if err != nil {
		t.Fatal(err)
	}
	current, err := presentation.NewFieldCatalog(
		append(fields, organizationOwnerFieldFixture(t)),
	)
	if err != nil {
		t.Fatal(err)
	}
	return revision, published, current
}

func organizationOwnerFieldFixture(t *testing.T) presentation.Field {
	t.Helper()
	field, err := presentation.NewField(
		organizationOwnerIDs, fieldregistry.TypeUUIDList,
		fieldregistry.ValueUUIDListV1, presentation.FieldReadOnly,
	)
	if err != nil {
		t.Fatal(err)
	}
	return field
}
