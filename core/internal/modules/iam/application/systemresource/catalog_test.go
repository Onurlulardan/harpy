package systemresource

import (
	"testing"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/catalog"
	formpresentation "github.com/Onurlulardan/harpy/core/internal/platform/fieldcatalog/presentation"
	"github.com/google/uuid"
)

func TestBuildFieldCatalogFiltersClassificationAndSealsModes(t *testing.T) {
	customID := uuid.MustParse("e1000000-0000-4000-8000-000000000001")
	custom, err := profile.NewField(profile.FieldInput{
		ID: customID, Resource: profile.ResourceUser, Key: "private_note",
		Type: profile.LogicalString, Cardinality: profile.CardinalitySingle,
		Classification: profile.ClassificationRestricted,
		Normalization:  profile.NormalizationNoneV1, State: profile.FieldActive,
	})
	if err != nil {
		t.Fatal(err)
	}
	published := publishedFixture(t, profile.ResourceUser, []profile.Field{custom})
	readOnly, _ := NewCatalogVisibility(true, false, false)
	visible, err := BuildFieldCatalog(published, readOnly)
	if err != nil {
		t.Fatal(err)
	}
	presentation, _ := visible.Presentation()
	if _, found, err := presentation.Field(customID.String()); err != nil || found {
		t.Fatal("restricted custom field leaked into read-only catalog")
	}
	idField, found, _ := presentation.Field(publicresource.UserIDFieldID)
	mode, _ := idField.Mode()
	if !found || mode != formpresentation.FieldReadOnly {
		t.Fatal("user ID is not sealed read-only")
	}
	if descriptorVisible(t, visible, customID, "private_note") {
		t.Fatal("restricted custom field name leaked into read-only catalog")
	}
	all, _ := NewCatalogVisibility(true, true, true)
	visible, err = BuildFieldCatalog(published, all)
	presentation, _ = visible.Presentation()
	customField, found, fieldErr := presentation.Field(customID.String())
	customMode, modeErr := customField.Mode()
	if err != nil || fieldErr != nil || modeErr != nil || !found ||
		customMode != formpresentation.FieldEditable {
		t.Fatal("authorized custom field was not editable")
	}
	if !descriptorVisible(t, visible, customID, "private_note") {
		t.Fatal("authorized custom field descriptor is missing")
	}
}

func TestBuildFieldCatalogCarriesQueryCapabilitiesForEveryIdentityResource(
	t *testing.T,
) {
	tests := []struct {
		kind           profile.ResourceKind
		searchableName string
	}{
		{profile.ResourceUser, "display_name"},
		{profile.ResourceRole, "name"},
		{profile.ResourceOrganization, "name"},
	}
	for _, test := range tests {
		t.Run(string(test.kind), func(t *testing.T) {
			published := publishedFixture(t, test.kind, nil)
			all, _ := NewCatalogVisibility(true, true, true)
			fieldCatalog, err := BuildFieldCatalog(published, all)
			if err != nil {
				t.Fatal(err)
			}
			assertDescriptorCapabilities(t, fieldCatalog, "id", true, false)
			assertDescriptorCapabilities(
				t, fieldCatalog, test.searchableName, false, true,
			)
			assertDescriptorCapabilities(t, fieldCatalog, "status", false, false)
			if test.kind == profile.ResourceUser {
				assertDescriptorCreateContract(
					t, fieldCatalog, "display_name", true, 128,
				)
			}
		})
	}
}

func assertDescriptorCreateContract(
	t *testing.T,
	fieldCatalog publicresource.FieldCatalog,
	name string,
	want bool,
	wantMaximumBytes int,
) {
	t.Helper()
	descriptors, err := fieldCatalog.Descriptors()
	if err != nil {
		t.Fatal(err)
	}
	for _, descriptor := range descriptors {
		publicName, _ := descriptor.PublicName()
		if publicName != name {
			continue
		}
		required, requiredErr := descriptor.RequiredForCreate()
		maximumBytes, maximumErr := descriptor.MaximumInputBytes()
		if requiredErr != nil || required != want || maximumErr != nil ||
			maximumBytes != wantMaximumBytes {
			t.Fatalf("%s create contract required=%v max=%d errors=%v/%v",
				name, required, maximumBytes, requiredErr, maximumErr)
		}
		return
	}
	t.Fatalf("descriptor %q is missing", name)
}

func assertDescriptorCapabilities(
	t *testing.T,
	fieldCatalog publicresource.FieldCatalog,
	name string,
	wantSortable bool,
	wantSearchable bool,
) {
	t.Helper()
	descriptors, err := fieldCatalog.Descriptors()
	if err != nil {
		t.Fatal(err)
	}
	for _, descriptor := range descriptors {
		publicName, nameErr := descriptor.PublicName()
		if nameErr != nil || publicName != name {
			continue
		}
		sortable, sortableErr := descriptor.Sortable()
		searchable, searchableErr := descriptor.Searchable()
		if sortableErr != nil || searchableErr != nil ||
			sortable != wantSortable || searchable != wantSearchable {
			t.Fatalf("%s capabilities=%v/%v errors=%v/%v",
				name, sortable, searchable, sortableErr, searchableErr)
		}
		return
	}
	t.Fatalf("descriptor %q is missing", name)
}

func descriptorVisible(
	t *testing.T,
	catalog publicresource.FieldCatalog,
	id uuid.UUID,
	wantName string,
) bool {
	t.Helper()
	descriptor, found, err := catalog.Descriptor(publicresource.ID(id))
	if err != nil || !found {
		return false
	}
	name, err := descriptor.PublicName()
	return err == nil && name == wantName
}

func publishedFixture(
	t *testing.T, kind profile.ResourceKind, fields []profile.Field,
) catalog.Published {
	t.Helper()
	if fields == nil {
		fields = []profile.Field{}
	}
	if kind == profile.ResourceUser {
		displayName, err := profile.NewField(profile.FieldInput{
			ID:       uuid.MustParse("e1000000-0000-4000-8000-0000000000ff"),
			Resource: profile.ResourceUser, Key: "display_name",
			Type: profile.LogicalString, Cardinality: profile.CardinalitySingle,
			Required:       true,
			Classification: profile.ClassificationPublic,
			Normalization:  profile.NormalizationNoneV1, State: profile.FieldActive,
			Validation: profile.ValidationInput{MaximumBytes: 128},
		})
		if err != nil {
			t.Fatal(err)
		}
		fields = append(fields, displayName)
	}
	schema, err := profile.NewPublishedSchema(profile.SchemaInput{
		ID:       uuid.MustParse("f1000000-0000-4000-8000-000000000001"),
		Resource: kind, Version: 1, State: profile.SchemaActive, Fields: fields,
	})
	if err != nil {
		t.Fatal(err)
	}
	queryCatalog, err := profile.NewQueryCatalog(schema)
	if err != nil {
		t.Fatal(err)
	}
	published, err := catalog.Build(queryCatalog)
	if err != nil {
		t.Fatal(err)
	}
	return published
}
