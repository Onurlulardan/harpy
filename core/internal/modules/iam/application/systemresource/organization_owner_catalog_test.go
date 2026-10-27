package systemresource

import (
	"testing"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	formpresentation "github.com/Onurlulardan/harpy/core/internal/platform/fieldcatalog/presentation"
)

func TestOrganizationOwnerFieldIsEditableAndRequiredAtSystemResourceBoundary(
	t *testing.T,
) {
	published := publishedFixture(t, profile.ResourceOrganization, nil)
	visibility, err := NewCatalogVisibility(true, true, true)
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := BuildFieldCatalog(published, visibility)
	if err != nil {
		t.Fatal(err)
	}
	presentation, err := catalog.Presentation()
	if err != nil {
		t.Fatal(err)
	}
	field, found, err := presentation.Field(
		publicresource.OrganizationOwnerIDsFieldID,
	)
	if err != nil || !found {
		t.Fatal("owner_ids presentation field missing")
	}
	mode, err := field.Mode()
	if err != nil || mode != formpresentation.FieldEditable {
		t.Fatal("owner_ids presentation field is not editable")
	}
	descriptor, found, err := catalog.Descriptor(
		mustSystemResourcePublicID(
			t, publicresource.OrganizationOwnerIDsFieldID,
		),
	)
	if err != nil || !found {
		t.Fatal("owner_ids descriptor missing")
	}
	name, err := descriptor.PublicName()
	if err != nil || name != "owner_ids" {
		t.Fatal("owner_ids descriptor name drifted")
	}
	required, requiredErr := descriptor.RequiredForCreate()
	maximumBytes, maximumErr := descriptor.MaximumInputBytes()
	if requiredErr != nil || maximumErr != nil || !required || maximumBytes != 4096 {
		t.Fatalf("owner_ids input contract required=%v max=%d", required, maximumBytes)
	}
}

func mustSystemResourcePublicID(
	t *testing.T,
	value string,
) publicresource.ID {
	t.Helper()
	id, err := publicresource.ParseID(value)
	if err != nil {
		t.Fatal(err)
	}
	return id
}
