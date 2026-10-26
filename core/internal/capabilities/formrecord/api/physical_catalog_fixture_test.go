package recordkernel

import (
	"testing"

	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
	"github.com/google/uuid"
)

func physicalCatalogFixture(
	t *testing.T,
) (Catalog, []byte, [32]byte, []PhysicalMapping, [32]byte) {
	t.Helper()
	mappings := []PhysicalMapping{
		physicalObject(t, formID, "form_table", "record", "forms_record_table_v2"),
		physicalField(t, fieldID, "text", "text_nfc_v1"),
		physicalField(t, countFieldID, "integer", "integer_base10_v1"),
		physicalField(t, calcFieldID, "integer", "integer_base10_v1"),
	}
	mappingSHA, err := DigestMappings(mappings)
	if err != nil {
		t.Fatal(err)
	}
	input := fixtureCompileInput(t, DefinitionV2Contract)
	input.Binding.MappingCatalogSHA256 = mappingSHA
	catalog, err := CompileCatalog(input)
	if err != nil {
		t.Fatal(err)
	}
	physical, physicalSHA, err := CompilePhysicalCatalogDocument(catalog, mappingSHA, mappings)
	if err != nil {
		t.Fatal(err)
	}
	return catalog, physical, physicalSHA, mappings, mappingSHA
}

func physicalObject(
	t *testing.T, id string, kind string, storage string, profile string,
) PhysicalMapping {
	t.Helper()
	identifierKind := schemav2.IdentifierForm
	if kind == "relation_table" {
		identifierKind = schemav2.IdentifierRelation
	}
	name, err := schemav2.DeriveIdentifier(identifierKind, id)
	if err != nil {
		t.Fatal(err)
	}
	return PhysicalMapping{LogicalID: uuid.MustParse(id), Kind: kind,
		PhysicalName: name, StorageType: storage, Variant: "object",
		Generation: 1, Profile: profile}
}

func physicalField(
	t *testing.T, id string, storage string, profile string,
) PhysicalMapping {
	t.Helper()
	name, err := schemav2.DeriveFieldIdentifier(id, 1, schemav2.VariantValue)
	if err != nil {
		t.Fatal(err)
	}
	return PhysicalMapping{LogicalID: uuid.MustParse(id), Kind: "field_column",
		PhysicalName: name, StorageType: storage, Variant: "value",
		Generation: 1, Profile: profile}
}
