package systemresource

import (
	"testing"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	"github.com/google/uuid"
)

func TestInstallationDisplayNamePublishesCreateValidationContract(t *testing.T) {
	fieldID := uuid.MustParse("e1000000-0000-4000-8000-0000000000aa")
	catalog, err := BuildInstallationFieldCatalog(
		uuid.MustParse("f1000000-0000-4000-8000-0000000000aa"),
		fieldID,
		publicresource.Users,
	)
	if err != nil {
		t.Fatal(err)
	}
	descriptor, found, err := catalog.Descriptor(publicresource.ID(fieldID))
	if err != nil || !found {
		t.Fatalf("display_name descriptor missing: found=%v err=%v", found, err)
	}
	required, requiredErr := descriptor.RequiredForCreate()
	maximumBytes, maximumErr := descriptor.MaximumInputBytes()
	if requiredErr != nil || !required || maximumErr != nil ||
		maximumBytes != 128 {
		t.Fatalf("display_name contract required=%v max=%d errors=%v/%v",
			required, maximumBytes, requiredErr, maximumErr)
	}
}

func TestInstallationOrganizationPresentationDefersOwnerCatalogField(t *testing.T) {
	catalog, err := BuildInstallationFieldCatalog(
		uuid.MustParse("f1000000-0000-4000-8000-0000000000ab"),
		uuid.MustParse("e1000000-0000-4000-8000-0000000000ab"),
		publicresource.Organizations,
	)
	if err != nil {
		t.Fatal(err)
	}
	presentationCatalog, err := catalog.Presentation()
	if err != nil {
		t.Fatal(err)
	}
	if _, found, err := presentationCatalog.Field(
		publicresource.OrganizationOwnerIDsFieldID,
	); err != nil || found {
		t.Fatalf("owner layout field found=%v err=%v", found, err)
	}
	_, _, querySHA, err := catalog.Reference()
	if err != nil || querySHA == [32]byte{} {
		t.Fatal("owner-aware query catalog evidence was lost")
	}
}
