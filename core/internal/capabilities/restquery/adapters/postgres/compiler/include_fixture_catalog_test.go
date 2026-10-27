package compiler

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

type includeCatalogFixture struct {
	set                  api.CatalogSet
	root, targetA        api.CatalogReference
	targetB, targetC     api.CatalogReference
	rootID, rootLabel    api.FieldDefinition
	aID, aLabel          api.FieldDefinition
	bID, bLabel          api.FieldDefinition
	cID, cLabel          api.FieldDefinition
	relationA, relationB api.RelationDefinition
	relationC            api.RelationDefinition
}

func newIncludeCatalogFixture(t *testing.T) includeCatalogFixture {
	t.Helper()
	fixture := includeCatalogFixture{
		root:    includeReference(t, "forms", includeRootResource, includeRootVersion, "1"),
		targetA: includeReference(t, "crm", includeAResource, includeAVersion, "2"),
		targetB: includeReference(t, "iam", includeBResource, includeBVersion, "3"),
		targetC: includeReference(t, "forms", includeCResource, includeCVersion, "4"),
	}
	fixture.rootID, fixture.rootLabel = includeFields(
		t, includeRootID, includeRootLabel, "root",
	)
	fixture.aID, fixture.aLabel = includeFields(t, includeAID, includeALabel, "a")
	fixture.bID, fixture.bLabel = includeFields(t, includeBID, includeBLabel, "b")
	fixture.cID, fixture.cLabel = includeFields(t, includeCID, includeCLabel, "c")
	fixture.relationC = includeRelation(
		t, includeRelationC, "grandchildren", fixture.targetC,
	)
	fixture.relationA = includeRelation(t, includeRelationA, "children_a", fixture.targetA)
	fixture.relationB = includeRelation(t, includeRelationB, "children_b", fixture.targetB)
	catalogs := []api.FieldCatalog{
		includeCatalog(t, fixture.root, "frm_include_root",
			[]api.FieldDefinition{fixture.rootID, fixture.rootLabel},
			[]api.RelationDefinition{fixture.relationA, fixture.relationB}, fixture.rootID),
		includeCatalog(t, fixture.targetA, "frm_include_a",
			[]api.FieldDefinition{fixture.aID, fixture.aLabel},
			[]api.RelationDefinition{fixture.relationC}, fixture.aID),
		includeCatalog(t, fixture.targetB, "frm_include_b",
			[]api.FieldDefinition{fixture.bID, fixture.bLabel}, nil, fixture.bID),
		includeCatalog(t, fixture.targetC, "frm_include_c",
			[]api.FieldDefinition{fixture.cID, fixture.cLabel}, nil, fixture.cID),
	}
	set, err := api.NewCatalogSet(fixture.root, catalogs)
	if err != nil {
		t.Fatal(err)
	}
	fixture.set = set
	return fixture
}

func includeFields(
	t *testing.T, id, label, suffix string,
) (api.FieldDefinition, api.FieldDefinition) {
	identifier := fixtureField(t, id, "id", "fld_"+suffix+"_id",
		fixtureUUIDType(t), "", "", []api.Operator{api.OperatorEq}, true, false)
	text := fixtureField(t, label, "label", "fld_"+suffix+"_label",
		fixtureFoldedTextType(t, false), "fld_"+suffix+"_label_folded", "",
		[]api.Operator{api.OperatorEq, api.OperatorNe}, true, false)
	return identifier, text
}
