package ddl

import (
	"strings"
	"testing"

	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
)

func TestRenderRelationCreateV2PinsManyEdgeSlots(t *testing.T) {
	mapping, arguments := relationFixtureV2(t, schemav2.CardinalityMany)
	statement, err := renderRelationCreateV2("tenant_schema", mapping, arguments)
	if err != nil {
		t.Fatal(err)
	}
	for _, contract := range []string{
		"edge_slot smallint NOT NULL", "CHECK (edge_slot BETWEEN 1 AND 100)",
		"UNIQUE (source_id, edge_slot)", "PRIMARY KEY (source_id, target_id)",
	} {
		if !strings.Contains(statement.Text, contract) {
			t.Fatalf("many relation is missing %q", contract)
		}
	}
}

func TestRenderRelationCreateV2PinsOneEdgeSlot(t *testing.T) {
	mapping, arguments := relationFixtureV2(t, schemav2.CardinalityOne)
	statement, err := renderRelationCreateV2("tenant_schema", mapping, arguments)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(statement.Text, "CHECK (edge_slot = 1)") ||
		!strings.Contains(statement.Text, "UNIQUE (source_id)") ||
		strings.Contains(statement.Text, "UNIQUE (source_id, edge_slot)") {
		t.Fatalf("one relation cardinality DDL differs: %s", statement.Text)
	}
}

func relationFixtureV2(
	t *testing.T,
	cardinality schemav2.Cardinality,
) (schemav2.IdentifierMapping, schemav2.RelationArguments) {
	t.Helper()
	ids := []string{
		"00000000-0000-4000-8000-000000000001",
		"00000000-0000-4000-8000-000000000002",
		"00000000-0000-4000-8000-000000000003",
		"00000000-0000-4000-8000-000000000004",
		"00000000-0000-4000-8000-000000000005",
		"00000000-0000-4000-8000-000000000006",
	}
	derive := func(kind schemav2.IdentifierKind, id string) schemav2.Identifier {
		name, err := schemav2.DeriveIdentifier(kind, id)
		if err != nil {
			t.Fatal(err)
		}
		return schemav2.Identifier{LogicalID: id, PhysicalName: name}
	}
	mapping := schemav2.IdentifierMapping{
		Form: derive(schemav2.IdentifierForm, ids[0]),
		RelatedForms: []schemav2.Identifier{
			derive(schemav2.IdentifierForm, ids[1]),
		},
		Fields: []schemav2.FieldIdentifier{},
		Relations: []schemav2.Identifier{
			derive(schemav2.IdentifierRelation, ids[2]),
		},
		Indexes: []schemav2.Identifier{},
		Constraints: []schemav2.Identifier{
			derive(schemav2.IdentifierConstraint, ids[3]),
			derive(schemav2.IdentifierConstraint, ids[4]),
		},
	}
	arguments := schemav2.RelationArguments{
		RelationID: ids[2], TargetFormID: ids[1], SourceConstraintID: ids[3],
		TargetConstraintID: ids[4], Cardinality: cardinality,
		DeletePolicy: schemav2.DeleteRestrict,
	}
	if cardinality == schemav2.CardinalityOne {
		mapping.Constraints = append(mapping.Constraints,
			derive(schemav2.IdentifierConstraint, ids[5]))
		arguments.CardinalityConstraintID = &ids[5]
	}
	return mapping, arguments
}
