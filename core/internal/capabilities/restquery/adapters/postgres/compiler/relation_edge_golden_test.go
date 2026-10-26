package compiler

import (
	"slices"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
	"github.com/Onurlulardan/harpy/core/internal/platform/physicalid"
	"github.com/google/uuid"
)

var includeSourceIDs = []uuid.UUID{
	uuid.MustParse("70000000-0000-4000-8000-000000000003"),
	uuid.MustParse("70000000-0000-4000-8000-000000000001"),
	uuid.MustParse("70000000-0000-4000-8000-000000000002"),
	uuid.MustParse("70000000-0000-4000-8000-000000000001"),
}

func TestCompileRelationEdgeGolden(t *testing.T) {
	fixture := newIncludePlanFixture(t)
	compiler := includeCompiler(t)
	graph, _ := compiler.CompileIncludeGraph(fixture.plan)
	nodes, _ := graph.Nodes()
	input := append([]uuid.UUID{}, includeSourceIDs...)
	statement, err := compiler.CompileRelationEdge(fixture.plan, graph, nodes[0], input)
	if err != nil {
		t.Fatal(err)
	}
	table, _ := physicalid.Name("rel_", uuid.MustParse(includeRelationA))
	wantSQL := `SELECT "source_id", "target_id", "edge_slot" FROM "` + table +
		`" WHERE "source_id" = ANY(CAST($1 AS pg_catalog.uuid[])) ` +
		`ORDER BY "source_id" ASC, "edge_slot" ASC, "target_id" ASC`
	sql, sqlErr := statement.SQL()
	explain, explainErr := statement.ExplainSQL()
	arguments, argumentsErr := statement.Arguments()
	class, classErr := statement.Class()
	owner, ownerErr := statement.Owner()
	columns, columnsErr := statement.ColumnCount()
	wantIDs := []uuid.UUID{includeSourceIDs[1], includeSourceIDs[2], includeSourceIDs[0]}
	if sqlErr != nil || explainErr != nil || argumentsErr != nil || classErr != nil ||
		ownerErr != nil || columnsErr != nil || sql != wantSQL ||
		explain != explainJSONPrefix+wantSQL || len(arguments) != 1 ||
		!slices.Equal(arguments[0].([]uuid.UUID), wantIDs) ||
		class != planner.StatementClassInclude || owner != "forms" || columns != 3 ||
		statement.ValidatePlan(fixture.plan, graph, nodes[0]) != nil {
		t.Fatal("relation-edge statement contract mismatch")
	}
	input[0] = uuid.Nil
	arguments[0].([]uuid.UUID)[0] = uuid.Nil
	again, _ := statement.Arguments()
	if !slices.Equal(again[0].([]uuid.UUID), wantIDs) || statement.Validate() != nil {
		t.Fatal("relation-edge identifier ownership escaped")
	}
}

func FuzzRelationEdgeUUIDNormalization(f *testing.F) {
	f.Add([]byte{3, 1, 2, 1})
	f.Fuzz(func(t *testing.T, input []byte) {
		if len(input) == 0 || len(input) > 64 {
			return
		}
		values := make([]uuid.UUID, len(input))
		for index, value := range input {
			values[index] = uuid.NewSHA1(uuid.NameSpaceOID, []byte{value})
		}
		normalized, err := normalizeIncludeUUIDs(values, maximumIncludeSourceIDs)
		if err != nil || len(normalized) < 1 {
			t.Fatal("valid UUID set was rejected")
		}
		for index := 1; index < len(normalized); index++ {
			if slices.Compare(normalized[index-1][:], normalized[index][:]) >= 0 {
				t.Fatal("UUIDs are not canonical and deduplicated")
			}
		}
	})
}
