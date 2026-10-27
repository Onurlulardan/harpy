package api

import (
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/sqlidentifier"
	"github.com/google/uuid"
)

const (
	rootResourceID    = "10000000-0000-4000-8000-000000000001"
	rootVersionID     = "10000000-0000-4000-8000-000000000002"
	targetResourceID  = "20000000-0000-4000-8000-000000000001"
	targetVersionID   = "20000000-0000-4000-8000-000000000002"
	rootStableField   = "30000000-0000-4000-8000-000000000001"
	rootSearchField   = "30000000-0000-4000-8000-000000000002"
	targetStableField = "40000000-0000-4000-8000-000000000001"
	targetSearchField = "40000000-0000-4000-8000-000000000002"
	rootRelationID    = "50000000-0000-4000-8000-000000000001"
)

func mustCatalogReferenceFixture(
	t *testing.T, owner, resourceID, versionID, hashDigit string,
) CatalogReference {
	t.Helper()
	hash, err := NewCatalogHash(strings.Repeat(hashDigit, 64))
	if err != nil {
		t.Fatal(err)
	}
	reference, err := NewCatalogReference(owner, resourceID, versionID, hash)
	if err != nil {
		t.Fatal(err)
	}
	return reference
}

func mustUUIDCatalogField(
	t *testing.T, id, name, tokenName string,
	selectable, sortable bool, bound int,
) FieldDefinition {
	t.Helper()
	fieldType, _ := NewUUIDFieldType()
	token, _ := sqlidentifier.MintDynamic(tokenName)
	storage, _ := NewFieldStorage(token, nil, nil)
	capabilities, err := NewFieldCapabilities(fieldType, FieldCapabilitiesInput{
		Selectable: selectable, Sortable: sortable, CursorCanonicalBytes: bound,
		Filters:   []FilterCapabilityInput{{Operator: OperatorEq, CountSafe: true}},
		Groupable: true, Aggregates: []AggregateFunction{AggregateFunctionCountDistinct},
	})
	if err != nil {
		t.Fatal(err)
	}
	field, err := NewFieldDefinition(id, name, fieldType, storage, capabilities)
	if err != nil {
		t.Fatal(err)
	}
	return field
}

func mustSearchCatalogField(
	t *testing.T, id, name, columnName, foldedName string,
	selectable, searchable bool,
) FieldDefinition {
	t.Helper()
	fieldType := mustCasefoldTextFieldType(t)
	column, _ := sqlidentifier.MintDynamic(columnName)
	folded, _ := sqlidentifier.MintDynamic(foldedName)
	storage, _ := NewFieldStorage(column, &folded, nil)
	capabilities, err := NewFieldCapabilities(fieldType, FieldCapabilitiesInput{
		Selectable: selectable, Searchable: searchable,
		Filters: []FilterCapabilityInput{{Operator: OperatorContains, CountSafe: false}},
	})
	if err != nil {
		t.Fatal(err)
	}
	field, err := NewFieldDefinition(id, name, fieldType, storage, capabilities)
	if err != nil {
		t.Fatal(err)
	}
	return field
}

func catalogFixtureInput(
	t *testing.T,
	reference CatalogReference,
	prefix string,
	stableID string,
	searchID string,
	relations []RelationDefinition,
) FieldCatalogInput {
	t.Helper()
	table, _ := sqlidentifier.MintDynamic("frm_" + prefix)
	storage, _ := NewResourceStorage(table)
	capabilities, _ := NewCatalogCapabilities(true)
	stable := mustUUIDCatalogField(
		t, stableID, "id", "fld_"+prefix+"_id", true, true, 36,
	)
	search := mustSearchCatalogField(
		t, searchID, "name", "fld_"+prefix+"_name",
		"fld_"+prefix+"_name_folded", true, true,
	)
	return FieldCatalogInput{
		Reference: reference, Storage: storage, Capabilities: capabilities,
		Fields: []FieldDefinition{stable, search}, Relations: relations,
		DefaultSelect: []uuid.UUID{uuid.MustParse(stableID), uuid.MustParse(searchID)},
		StableID:      uuid.MustParse(stableID), SearchFields: []uuid.UUID{uuid.MustParse(searchID)},
	}
}

func mustFieldCatalogFixture(t *testing.T, input FieldCatalogInput) FieldCatalog {
	t.Helper()
	catalog, err := NewFieldCatalog(input)
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}
