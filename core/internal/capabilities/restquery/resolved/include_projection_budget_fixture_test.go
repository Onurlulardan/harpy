package resolved

import (
	"fmt"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/platform/physicalid"
	"github.com/Onurlulardan/harpy/core/internal/platform/sqlidentifier"
	"github.com/google/uuid"
)

func resolvedBudgetInclude(t *testing.T, visible int) (api.CatalogReference, Include) {
	t.Helper()
	source := resolvedBudgetReference(t, "forms", 1, "8")
	target := resolvedBudgetReference(t, "crm", 2, "9")
	stable := resolvedBudgetField(t, 1)
	relationID := uuid.MustParse("65000000-0000-4000-8000-000000000001")
	tableName, _ := physicalid.Name("rel_", relationID)
	table, _ := sqlidentifier.MintDynamic(tableName)
	relationStorage, _ := api.NewRelationStorage(relationID.String(), table)
	relation, err := api.NewRelationDefinition(
		relationID.String(), "children", api.RelationCardinalityToMany,
		relationStorage, target, true, 100,
	)
	if err != nil {
		t.Fatal(err)
	}
	projections := make([]Projection, 0, visible+1)
	for index := 0; index < visible; index++ {
		field := resolvedBudgetField(t, index+2)
		key, _ := NewOutputKey(fmt.Sprintf("field_%02d", index))
		path, _ := NewRequestPath(fmt.Sprintf("/include/0/select/%d", index))
		projection, err := NewProjection(field, key, path, false)
		if err != nil {
			t.Fatal(err)
		}
		projections = append(projections, projection)
	}
	hiddenKey, _ := NewOutputKey("id")
	selectPath, _ := NewRequestPath("/include/0/select")
	hidden, _ := NewProjection(stable, hiddenKey, selectPath, true)
	projections = append(projections, hidden)
	sortPath, _ := NewRequestPath("/include/0/sort")
	sort, _ := NewSort(stable, api.DirectionAscending, sortPath, true)
	includePath, _ := NewRequestPath("/include/0")
	targetTable, _ := sqlidentifier.MintDynamic("frm_resolved_budget_target")
	targetStorage, _ := api.NewResourceStorage(targetTable)
	include, err := NewInclude(
		relation, target, targetStorage, relationStorage, stable, includePath,
		projections, nil, []Sort{sort}, 100, []Include{},
	)
	if err != nil {
		t.Fatal(err)
	}
	return source, include
}

func resolvedBudgetReference(
	t *testing.T,
	owner string,
	offset int,
	hashDigit string,
) api.CatalogReference {
	t.Helper()
	resource := fmt.Sprintf("66000000-0000-4000-8000-%012d", offset)
	version := fmt.Sprintf("67000000-0000-4000-8000-%012d", offset)
	hash, _ := api.NewCatalogHash(strings.Repeat(hashDigit, 64))
	reference, err := api.NewCatalogReference(owner, resource, version, hash)
	if err != nil {
		t.Fatal(err)
	}
	return reference
}

func resolvedBudgetField(t *testing.T, offset int) api.FieldDefinition {
	t.Helper()
	identifier := fmt.Sprintf("68000000-0000-4000-8000-%012d", offset)
	name := fmt.Sprintf("field_%02d", offset)
	if offset == 1 {
		name = "id"
	}
	fieldType, _ := api.NewUUIDFieldType()
	column, _ := sqlidentifier.MintDynamic(fmt.Sprintf("fld_resolved_budget_%02d", offset))
	storage, _ := api.NewFieldStorage(column, nil, nil)
	capabilities, _ := api.NewFieldCapabilities(
		fieldType, api.FieldCapabilitiesInput{
			Selectable: true, Sortable: true, CursorCanonicalBytes: 36,
			Filters: []api.FilterCapabilityInput{{Operator: api.OperatorEq, CountSafe: true}},
		},
	)
	field, err := api.NewFieldDefinition(
		identifier, name, fieldType, storage, capabilities,
	)
	if err != nil {
		t.Fatal(err)
	}
	return field
}
