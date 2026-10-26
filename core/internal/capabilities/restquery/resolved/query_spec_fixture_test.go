package resolved

import (
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/platform/sqlidentifier"
	"github.com/google/uuid"
)

const (
	resolvedResourceID = "70000000-0000-4000-8000-000000000001"
	resolvedVersionID  = "70000000-0000-4000-8000-000000000002"
	resolvedFieldID    = "70000000-0000-4000-8000-000000000003"
)

func querySpecFixture(t *testing.T) QuerySpec {
	t.Helper()
	hash, err := api.NewCatalogHash(strings.Repeat("7", 64))
	if err != nil {
		t.Fatal(err)
	}
	reference, err := api.NewCatalogReference(
		"forms", resolvedResourceID, resolvedVersionID, hash,
	)
	if err != nil {
		t.Fatal(err)
	}
	field := resolvedStableField(t)
	table, _ := sqlidentifier.MintDynamic("frm_resolved_records")
	resourceStorage, _ := api.NewResourceStorage(table)
	catalogCapabilities, _ := api.NewCatalogCapabilities(true)
	fieldID := uuid.MustParse(resolvedFieldID)
	catalog, err := api.NewFieldCatalog(api.FieldCatalogInput{
		Reference: reference, Storage: resourceStorage,
		Capabilities: catalogCapabilities,
		Fields:       []api.FieldDefinition{field}, Relations: []api.RelationDefinition{},
		DefaultSelect: []uuid.UUID{fieldID}, StableID: fieldID,
		SearchFields: []uuid.UUID{},
	})
	if err != nil {
		t.Fatal(err)
	}
	set, err := api.NewCatalogSet(reference, []api.FieldCatalog{catalog})
	if err != nil {
		t.Fatal(err)
	}
	selectPath, _ := NewRequestPath("/select/0")
	sortPath, _ := NewRequestPath("/sort")
	countPath, _ := NewRequestPath("/count")
	rootPath, _ := NewRequestPath("")
	key, _ := NewOutputKey("id")
	projection, _ := NewProjection(field, key, selectPath, false)
	sort, _ := NewSort(field, api.DirectionAscending, sortPath, true)
	page, _ := NewPage(10, nil)
	record, err := NewRecordQuery(
		[]Projection{projection}, nil, []Sort{sort}, []SearchTerm{}, page,
		api.CountModeNone, countPath, true, []Include{},
	)
	if err != nil {
		t.Fatal(err)
	}
	spec, err := NewRecordQuerySpec(
		set, reference, resourceStorage, field, rootPath, record,
	)
	if err != nil {
		t.Fatal(err)
	}
	return spec
}

func resolvedStableField(t *testing.T) api.FieldDefinition {
	t.Helper()
	fieldType, _ := api.NewUUIDFieldType()
	column, _ := sqlidentifier.MintDynamic("fld_resolved_id")
	storage, _ := api.NewFieldStorage(column, nil, nil)
	capabilities, err := api.NewFieldCapabilities(
		fieldType, api.FieldCapabilitiesInput{
			Selectable: true, Sortable: true, CursorCanonicalBytes: 36,
			Filters: []api.FilterCapabilityInput{
				{Operator: api.OperatorEq, CountSafe: true},
			},
			Groupable:  true,
			Aggregates: []api.AggregateFunction{api.AggregateFunctionCountDistinct},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	field, err := api.NewFieldDefinition(
		resolvedFieldID, "id", fieldType, storage, capabilities,
	)
	if err != nil {
		t.Fatal(err)
	}
	return field
}
