package catalog

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	"github.com/Onurlulardan/harpy/core/internal/platform/sqlidentifier"
	"github.com/google/uuid"
)

func TestBuildPublishesFinalCoreCatalogs(t *testing.T) {
	for _, kind := range []profile.ResourceKind{
		profile.ResourceUser, profile.ResourceRole, profile.ResourceOrganization,
	} {
		t.Run(string(kind), func(t *testing.T) {
			published := buildFixture(t, kind, nil)
			set, err := published.CatalogSet()
			catalog, catalogErr := set.RootCatalog()
			resource, resourceErr := published.Resource()
			reference, referenceErr := catalog.Reference()
			owner, ownerErr := reference.Owner()
			resourceID, idErr := reference.ResourceID()
			stableID, stableErr := catalog.StableID()
			fields, fieldsErr := catalog.Fields()
			defaults, defaultsErr := catalog.DefaultSelect()
			if err != nil || catalogErr != nil || resourceErr != nil ||
				referenceErr != nil || ownerErr != nil || idErr != nil ||
				stableErr != nil || fieldsErr != nil || defaultsErr != nil {
				t.Fatalf("catalog access failed")
			}
			if owner != Owner || resourceID != resource.ID() || stableID == uuid.Nil ||
				len(fields) == 0 || len(defaults) == 0 {
				t.Fatalf("final catalog shape is incomplete")
			}
			assertStaticTable(t, catalog, resource.Table())
			assertStatusField(t, catalog, resource.StatusFieldID())
		})
	}
}

func TestBuildSealsCustomFieldToStableUUIDJSONBPath(t *testing.T) {
	fieldID := uuid.MustParse("e1000000-0000-4000-8000-000000000001")
	field, err := profile.NewField(profile.FieldInput{
		ID: fieldID, Resource: profile.ResourceUser, Key: "security_training_complete",
		Type: profile.LogicalBoolean, Cardinality: profile.CardinalitySingle,
		Required:       true,
		Classification: profile.ClassificationRestricted,
		Normalization:  profile.NormalizationNoneV1, State: profile.FieldActive,
		Validation: profile.ValidationInput{MaximumBytes: 64},
	})
	if err != nil {
		t.Fatal(err)
	}
	published := buildFixture(t, profile.ResourceUser, []profile.Field{field})
	set, _ := published.CatalogSet()
	catalog, _ := set.RootCatalog()
	definition, found, definitionErr := catalog.Field(fieldID)
	if definitionErr != nil || !found {
		t.Fatal("custom field missing")
	}
	storage, _ := definition.Storage()
	key, present, keyErr := storage.JSONBKey()
	name, nameErr := definition.PublicName()
	classification, classErr := published.Classification(fieldID)
	required, requiredErr := published.RequiredForCreate(fieldID)
	maximumBytes, maximumErr := published.MaximumInputBytes(fieldID)
	if keyErr != nil || !present || key != fieldID || nameErr != nil ||
		name != "security_training_complete" || classErr != nil ||
		classification != profile.ClassificationRestricted ||
		requiredErr != nil || !required || maximumErr != nil || maximumBytes != 64 {
		t.Fatal("custom field did not retain sealed identity and classification")
	}
	document, _ := sqlidentifier.MintStatic("profile_values")
	expected, _ := sqlidentifier.Fingerprint(document)
	column, _ := storage.Column()
	actual, _ := sqlidentifier.Fingerprint(column)
	if expected != actual {
		t.Fatal("custom field storage is not profile_values")
	}
}

func buildFixture(
	t *testing.T,
	kind profile.ResourceKind,
	fields []profile.Field,
) Published {
	t.Helper()
	if fields == nil {
		fields = []profile.Field{}
	}
	if kind == profile.ResourceUser {
		hasDisplayName := false
		for _, field := range fields {
			hasDisplayName = hasDisplayName || field.Key() == "display_name"
		}
		if !hasDisplayName {
			displayName, err := profile.NewField(profile.FieldInput{
				ID:       uuid.MustParse("e1000000-0000-4000-8000-0000000000ff"),
				Resource: profile.ResourceUser, Key: "display_name",
				Type: profile.LogicalString, Cardinality: profile.CardinalitySingle,
				Classification: profile.ClassificationPublic,
				Normalization:  profile.NormalizationNoneV1, State: profile.FieldActive,
			})
			if err != nil {
				t.Fatal(err)
			}
			fields = append(fields, displayName)
		}
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
	published, err := Build(queryCatalog)
	if err != nil {
		t.Fatal(err)
	}
	return published
}

func assertStaticTable(t *testing.T, catalog api.FieldCatalog, value string) {
	t.Helper()
	storage, _ := catalog.Storage()
	actualToken, _ := storage.Table()
	expectedToken, _ := sqlidentifier.MintStatic(value)
	actual, _ := sqlidentifier.Fingerprint(actualToken)
	expected, _ := sqlidentifier.Fingerprint(expectedToken)
	if actual != expected {
		t.Fatalf("table token mismatch")
	}
}

func assertStatusField(t *testing.T, catalog api.FieldCatalog, id uuid.UUID) {
	t.Helper()
	field, found, err := catalog.Field(id)
	fieldType, typeErr := field.Type()
	kind, kindErr := fieldType.Kind()
	if err != nil || !found || typeErr != nil || kindErr != nil ||
		kind != api.LogicalTypeEnum {
		t.Fatal("terminal status field is not a closed enum")
	}
}
