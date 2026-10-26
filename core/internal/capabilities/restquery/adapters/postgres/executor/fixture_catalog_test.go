package executor

import (
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/platform/sqlidentifier"
	"github.com/google/uuid"
)

const (
	fixtureResource = "10000000-0000-4000-8000-000000000001"
	fixtureVersion  = "10000000-0000-4000-8000-000000000002"
	fixtureID       = "20000000-0000-4000-8000-000000000001"
	fixtureName     = "20000000-0000-4000-8000-000000000002"
	fixtureJSON     = "20000000-0000-4000-8000-000000000003"
)

func fixtureCatalogSet(t *testing.T, owner string) api.CatalogSet {
	t.Helper()
	hash, _ := api.NewCatalogHash(strings.Repeat("a", 64))
	reference, err := api.NewCatalogReference(
		owner, fixtureResource, fixtureVersion, hash,
	)
	if err != nil {
		t.Fatal(err)
	}
	idType, _ := api.NewUUIDFieldType()
	profile, _ := api.NewTextProfile(
		api.TextCasefoldUnicodeDefaultV1, api.TextAccentDisabled,
	)
	nameType, _ := api.NewTextFieldType(profile)
	jsonType, _ := api.NewJSONFieldType()
	fields := []api.FieldDefinition{
		fixtureField(t, fixtureID, "id", "id", idType,
			api.FieldCapabilitiesInput{
				Selectable: true, Sortable: true, CursorCanonicalBytes: 36,
				Filters: []api.FilterCapabilityInput{{
					Operator: api.OperatorEq, CountSafe: true,
				}},
			}),
		fixtureField(t, fixtureName, "name", "name", nameType,
			api.FieldCapabilitiesInput{Selectable: true}),
		fixtureField(t, fixtureJSON, "metadata", "metadata", jsonType,
			api.FieldCapabilitiesInput{Selectable: true}),
	}
	table, _ := sqlidentifier.MintDynamic("frm_executor_fixture")
	storage, _ := api.NewResourceStorage(table)
	capabilities, _ := api.NewCatalogCapabilities(true)
	catalog, err := api.NewFieldCatalog(api.FieldCatalogInput{
		Reference: reference, Storage: storage, Capabilities: capabilities,
		Fields: fields, DefaultSelect: []uuid.UUID{
			uuid.MustParse(fixtureID), uuid.MustParse(fixtureName),
		},
		StableID: uuid.MustParse(fixtureID),
	})
	if err != nil {
		t.Fatal(err)
	}
	set, err := api.NewCatalogSet(reference, []api.FieldCatalog{catalog})
	if err != nil {
		t.Fatal(err)
	}
	return set
}

func fixtureField(
	t *testing.T,
	id string,
	name string,
	columnName string,
	fieldType api.FieldType,
	input api.FieldCapabilitiesInput,
) api.FieldDefinition {
	t.Helper()
	column, _ := sqlidentifier.MintDynamic("fld_" + columnName)
	var folded *sqlidentifier.Token
	kind, _ := fieldType.Kind()
	if kind == api.LogicalTypeText {
		token, _ := sqlidentifier.MintDynamic("fld_" + columnName + "_folded")
		folded = &token
	}
	storage, err := api.NewFieldStorage(column, folded, nil)
	if err != nil {
		t.Fatal(err)
	}
	capabilities, err := api.NewFieldCapabilities(fieldType, input)
	if err != nil {
		t.Fatal(err)
	}
	field, err := api.NewFieldDefinition(id, name, fieldType, storage, capabilities)
	if err != nil {
		t.Fatal(err)
	}
	return field
}
