package resolver

import (
	"fmt"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/platform/sqlidentifier"
	"github.com/google/uuid"
)

const (
	budgetRootResource   = "61000000-0000-4000-8000-000000000001"
	budgetRootVersion    = "61000000-0000-4000-8000-000000000002"
	budgetTargetResource = "61000000-0000-4000-8000-000000000003"
	budgetTargetVersion  = "61000000-0000-4000-8000-000000000004"
	budgetRelation       = "61000000-0000-4000-8000-000000000005"
	budgetRootID         = "62000000-0000-4000-8000-000000000001"
	budgetRootName       = "62000000-0000-4000-8000-000000000002"
	budgetTargetID       = "62000000-0000-4000-8000-000000000003"
)

func projectionBudgetCatalogs(
	t *testing.T,
	defaultTargetVisible int,
) api.CatalogSet {
	t.Helper()
	rootReference := fixtureReference(
		t, "forms", budgetRootResource, budgetRootVersion, "5",
	)
	targetReference := fixtureReference(
		t, "crm", budgetTargetResource, budgetTargetVersion, "6",
	)
	relation, err := api.NewRelationDefinition(
		budgetRelation, "children", api.RelationCardinalityToMany,
		fixtureRelationStorage(t, budgetRelation), targetReference, true, 100,
	)
	if err != nil {
		t.Fatal(err)
	}
	rootFields := []api.FieldDefinition{
		fixtureUUIDField(t, budgetRootID, "id", "fld_budget_root_id"),
		fixtureTextField(t, budgetRootName, "name", "budget_root_name"),
	}
	root := fixtureCatalog(
		t, rootReference, "frm_budget_root", rootFields,
		[]api.RelationDefinition{relation}, budgetRootID, budgetRootName, true,
	)
	target := projectionBudgetTarget(t, targetReference, defaultTargetVisible)
	set, err := api.NewCatalogSet(
		rootReference, []api.FieldCatalog{root, target},
	)
	if err != nil {
		t.Fatal(err)
	}
	return set
}

func projectionBudgetTarget(
	t *testing.T,
	reference api.CatalogReference,
	defaultVisible int,
) api.FieldCatalog {
	t.Helper()
	fields := make([]api.FieldDefinition, 51)
	fields[0] = fixtureUUIDField(t, budgetTargetID, "id", "fld_budget_target_id")
	for index := 1; index < len(fields); index++ {
		fields[index] = fixtureTextField(
			t, budgetFieldID(index), budgetFieldName(index),
			fmt.Sprintf("budget_target_%02d", index),
		)
	}
	defaults := make([]uuid.UUID, defaultVisible)
	for index := range defaults {
		defaults[index], _ = fields[index].ID()
	}
	table, _ := sqlidentifier.MintDynamic("frm_budget_target")
	storage, _ := api.NewResourceStorage(table)
	capabilities, _ := api.NewCatalogCapabilities(true)
	catalog, err := api.NewFieldCatalog(api.FieldCatalogInput{
		Reference: reference, Storage: storage, Capabilities: capabilities,
		Fields: fields, Relations: []api.RelationDefinition{},
		DefaultSelect: defaults, StableID: uuid.MustParse(budgetTargetID),
		SearchFields: []uuid.UUID{},
	})
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}

func budgetFieldID(index int) string {
	return fmt.Sprintf("62000000-0000-4000-8000-%012d", index+3)
}

func budgetFieldName(index int) string {
	return fmt.Sprintf("field_%02d", index)
}
