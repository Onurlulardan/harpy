package executor

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/platform/sqlidentifier"
	"github.com/google/uuid"
)

type graphCatalogFixture struct {
	set              api.CatalogSet
	root, targetA    api.CatalogReference
	targetB, targetC api.CatalogReference
}

func newGraphCatalogFixture(t *testing.T) graphCatalogFixture {
	t.Helper()
	fixture := graphCatalogFixture{
		root:    graphReference(t, "forms", graphRootResource, graphRootVersion, "1"),
		targetA: graphReference(t, "crm", graphAResource, graphAVersion, "2"),
		targetB: graphReference(t, "iam", graphBResource, graphBVersion, "3"),
		targetC: graphReference(t, "forms", graphCResource, graphCVersion, "4"),
	}
	rootFields := graphFields(t, graphRootID, graphRootLabel, "graph_root")
	aFields := graphFields(t, graphAID, graphALabel, "graph_a")
	bFields := graphFields(t, graphBID, graphBLabel, "graph_b")
	cFields := graphFields(t, graphCID, graphCLabel, "graph_c")
	relationC := graphRelation(
		t, graphRelationC, "items", api.RelationCardinalityToMany, fixture.targetC,
	)
	relationA := graphRelation(
		t, graphRelationA, "children", api.RelationCardinalityToMany, fixture.targetA,
	)
	relationB := graphRelation(
		t, graphRelationB, "owner", api.RelationCardinalityToOne, fixture.targetB,
	)
	catalogs := []api.FieldCatalog{
		graphCatalog(t, fixture.root, "frm_graph_root", rootFields,
			[]api.RelationDefinition{relationA, relationB}),
		graphCatalog(t, fixture.targetA, "frm_graph_a", aFields,
			[]api.RelationDefinition{relationC}),
		graphCatalog(t, fixture.targetB, "frm_graph_b", bFields, nil),
		graphCatalog(t, fixture.targetC, "frm_graph_c", cFields, nil),
	}
	set, err := api.NewCatalogSet(fixture.root, catalogs)
	if err != nil {
		t.Fatal(err)
	}
	fixture.set = set
	return fixture
}

func graphFields(
	t *testing.T,
	id string,
	label string,
	suffix string,
) []api.FieldDefinition {
	t.Helper()
	idType, _ := api.NewUUIDFieldType()
	profile, _ := api.NewTextProfile(
		api.TextCasefoldUnicodeDefaultV1, api.TextAccentDisabled,
	)
	textType, _ := api.NewTextFieldType(profile)
	return []api.FieldDefinition{
		fixtureField(t, id, "id", suffix+"_id", idType,
			api.FieldCapabilitiesInput{
				Selectable: true, Sortable: true, CursorCanonicalBytes: 36,
				Filters: []api.FilterCapabilityInput{{
					Operator: api.OperatorEq, CountSafe: true,
				}},
			}),
		fixtureField(t, label, "label", suffix+"_label", textType,
			api.FieldCapabilitiesInput{
				Selectable: true, Sortable: true, CursorCanonicalBytes: 256,
				Filters: []api.FilterCapabilityInput{{
					Operator: api.OperatorEq, CountSafe: true,
				}},
			}),
	}
}

func graphCatalog(
	t *testing.T,
	reference api.CatalogReference,
	table string,
	fields []api.FieldDefinition,
	relations []api.RelationDefinition,
) api.FieldCatalog {
	t.Helper()
	token, _ := sqlidentifier.MintDynamic(table)
	storage, _ := api.NewResourceStorage(token)
	capabilities, _ := api.NewCatalogCapabilities(true)
	stable, _ := fields[0].ID()
	defaults := []uuid.UUID{stable}
	label, _ := fields[1].ID()
	defaults = append(defaults, label)
	catalog, err := api.NewFieldCatalog(api.FieldCatalogInput{
		Reference: reference, Storage: storage, Capabilities: capabilities,
		Fields: fields, Relations: relations, DefaultSelect: defaults,
		StableID: stable, SearchFields: []uuid.UUID{},
	})
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}
