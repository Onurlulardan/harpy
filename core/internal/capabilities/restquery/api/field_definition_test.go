package api

import (
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/sqlidentifier"
)

func TestFieldDefinitionCanonicalRoundTrip(t *testing.T) {
	field := mustUUIDCatalogField(
		t, rootStableField, "record_id", "fld_record_id", true, true, 36,
	)
	id, idErr := field.ID()
	name, nameErr := field.PublicName()
	fieldType, typeErr := field.Type()
	storage, storageErr := field.Storage()
	capabilities, capabilityErr := field.Capabilities()
	kind, kindErr := fieldType.Kind()
	selectable, selectableErr := capabilities.Selectable()
	if idErr != nil || nameErr != nil || typeErr != nil || storageErr != nil ||
		capabilityErr != nil || kindErr != nil || selectableErr != nil {
		t.Fatal("valid field definition accessor failed")
	}
	if id.String() != rootStableField || name != "record_id" ||
		kind != LogicalTypeUUID || !selectable || storage.Validate() != nil {
		t.Fatal("field definition did not round trip")
	}
}

func TestFieldDefinitionRejectsInvalidIdentityOrComposition(t *testing.T) {
	valid := mustUUIDCatalogField(
		t, rootStableField, "record_id", "fld_record_id", true, true, 36,
	)
	invalidIDs := []string{"", strings.ToUpper("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"), "not-a-uuid"}
	invalidNames := []string{"", "_included", "e\u0301", "bad\x00name"}
	for _, id := range invalidIDs {
		if _, err := NewFieldDefinition(
			id, valid.publicName, valid.fieldType, valid.storage, valid.capabilities,
		); err == nil {
			t.Fatalf("invalid field ID accepted: %q", id)
		}
	}
	for _, name := range invalidNames {
		if _, err := NewFieldDefinition(
			rootStableField, name, valid.fieldType, valid.storage, valid.capabilities,
		); err == nil {
			t.Fatalf("invalid field name accepted: %q", name)
		}
	}
	wrongType := mustCasefoldTextFieldType(t)
	if _, err := NewFieldDefinition(
		rootStableField, "record_id", wrongType, valid.storage, valid.capabilities,
	); err == nil {
		t.Fatal("capabilities were accepted with a different field type")
	}
	if _, err := NewFieldDefinition(
		rootStableField, "record_id", valid.fieldType, FieldStorage{}, valid.capabilities,
	); err == nil {
		t.Fatal("zero field storage accepted")
	}
}

func TestFieldDefinitionEnforcesTextCompanionProfile(t *testing.T) {
	column, _ := sqlidentifier.MintDynamic("fld_title")
	folded, _ := sqlidentifier.MintDynamic("fld_title_folded")
	accent, _ := sqlidentifier.MintDynamic("fld_title_accent")
	profile, _ := NewTextProfile(
		TextCasefoldUnicodeDefaultV1, TextAccentInsensitiveV1,
	)
	fieldType, _ := NewTextFieldType(profile)
	capabilities, _ := NewFieldCapabilities(fieldType, FieldCapabilitiesInput{Searchable: true})
	for _, companions := range []struct {
		folded *sqlidentifier.Token
		accent *sqlidentifier.Token
		valid  bool
	}{
		{nil, nil, false}, {&folded, nil, false}, {&folded, &accent, true},
	} {
		storage, storageErr := NewFieldStorage(column, companions.folded, companions.accent)
		_, err := NewFieldDefinition(
			rootSearchField, "title", fieldType, storage, capabilities,
		)
		if storageErr != nil || (err == nil) != companions.valid {
			t.Fatalf("companion profile valid=%v, want %v", err == nil, companions.valid)
		}
	}
}
