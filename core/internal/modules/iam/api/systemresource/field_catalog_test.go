package systemresource

import (
	"crypto/sha256"
	"testing"

	fieldregistry "github.com/Onurlulardan/harpy/core/internal/platform/fieldcatalog/fieldregistry"
	formpresentation "github.com/Onurlulardan/harpy/core/internal/platform/fieldcatalog/presentation"
)

func TestFieldCatalogBindsIAMAndPresentationEvidence(t *testing.T) {
	catalog := testUserFieldCatalog(t)
	if catalog.Validate() != nil {
		t.Fatal("catalog validation failed")
	}
	presentation, err := catalog.Presentation()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewFieldCatalog(
		Roles, mustSystemResourceID(t, UsersCatalogResourceID),
		mustSystemResourceID(t, "018f0f4e-7b61-7c25-8000-000000000031"),
		sha256.Sum256([]byte("query-catalog")), presentation,
		[]FieldDescriptor{mustFieldDescriptor(t, UserIDFieldID, "id")},
	); err == nil {
		t.Fatal("cross-resource catalog ID accepted")
	}
}

func TestFieldDescriptorCapabilitiesDefaultClosedAndAffectExactEquality(t *testing.T) {
	fieldID := mustSystemResourceID(t, UserIDFieldID)
	legacy, err := NewFieldDescriptor(fieldID, "id")
	if err != nil {
		t.Fatal(err)
	}
	sortable, sortableErr := legacy.Sortable()
	searchable, searchableErr := legacy.Searchable()
	required, requiredErr := legacy.RequiredForCreate()
	maximumBytes, maximumErr := legacy.MaximumInputBytes()
	if sortableErr != nil || searchableErr != nil || requiredErr != nil ||
		maximumErr != nil || sortable || searchable || required || maximumBytes != 0 {
		t.Fatalf("legacy capabilities sortable=%v searchable=%v required=%v max=%d errors=%v/%v/%v/%v",
			sortable, searchable, required, maximumBytes,
			sortableErr, searchableErr, requiredErr, maximumErr)
	}
	capable, err := NewFieldDescriptorWithCapabilities(fieldID, "id", true, true)
	if err != nil {
		t.Fatal(err)
	}
	base := testUserFieldCatalogWithDescriptor(t, legacy)
	changed := testUserFieldCatalogWithDescriptor(t, capable)
	equal, equalErr := base.ExactEqual(changed)
	if equalErr != nil || equal {
		t.Fatalf("capability change equality=%v error=%v", equal, equalErr)
	}
	requiredDescriptor, err := NewFieldDescriptorWithContract(
		fieldID, "id", false, false, true, 0,
	)
	if err != nil {
		t.Fatal(err)
	}
	requiredChanged := testUserFieldCatalogWithDescriptor(t, requiredDescriptor)
	equal, equalErr = base.ExactEqual(requiredChanged)
	if equalErr != nil || equal {
		t.Fatalf("create-required change equality=%v error=%v", equal, equalErr)
	}
	maximumDescriptor, err := NewFieldDescriptorWithContract(
		fieldID, "id", false, false, false, 128,
	)
	if err != nil {
		t.Fatal(err)
	}
	maximumChanged := testUserFieldCatalogWithDescriptor(t, maximumDescriptor)
	equal, equalErr = base.ExactEqual(maximumChanged)
	if equalErr != nil || equal {
		t.Fatalf("maximum input bytes change equality=%v error=%v", equal, equalErr)
	}
	if _, err := (FieldDescriptor{}).Sortable(); err != ErrContract {
		t.Fatalf("zero descriptor sortable error=%v", err)
	}
}

func testUserFieldCatalog(t *testing.T) FieldCatalog {
	t.Helper()
	return testUserFieldCatalogWithDescriptor(
		t, mustFieldDescriptor(t, UserIDFieldID, "id"),
	)
}

func testUserFieldCatalogWithDescriptor(
	t *testing.T,
	descriptor FieldDescriptor,
) FieldCatalog {
	t.Helper()
	field, err := formpresentation.NewField(
		"a1000000-0000-4000-8000-000000000001",
		fieldregistry.TypeUUID, fieldregistry.ValueUUIDRFC4122LowerV1,
		formpresentation.FieldReadOnly,
	)
	if err != nil {
		t.Fatal(err)
	}
	presentation, err := formpresentation.NewFieldCatalog([]formpresentation.Field{field})
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := NewFieldCatalog(
		Users, mustSystemResourceID(t, UsersCatalogResourceID),
		mustSystemResourceID(t, "018f0f4e-7b61-7c25-8000-000000000031"),
		sha256.Sum256([]byte("query-catalog")), presentation,
		[]FieldDescriptor{descriptor},
	)
	if err != nil || catalog.Validate() != nil {
		t.Fatalf("catalog error=%v", err)
	}
	return catalog
}

func mustFieldDescriptor(t *testing.T, id, name string) FieldDescriptor {
	t.Helper()
	parsed, err := ParseID(id)
	if err != nil {
		t.Fatal(err)
	}
	descriptor, err := NewFieldDescriptor(parsed, name)
	if err != nil {
		t.Fatal(err)
	}
	return descriptor
}
