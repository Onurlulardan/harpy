package compiler

import (
	"slices"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
	"github.com/google/uuid"
)

var includeCandidateIDs = []uuid.UUID{
	uuid.MustParse("71000000-0000-4000-8000-000000000003"),
	uuid.MustParse("71000000-0000-4000-8000-000000000001"),
	uuid.MustParse("71000000-0000-4000-8000-000000000002"),
	uuid.MustParse("71000000-0000-4000-8000-000000000001"),
}

func TestCompileIncludeTargetGolden(t *testing.T) {
	fixture := newIncludePlanFixture(t)
	compiler := includeCompiler(t)
	graph, _ := compiler.CompileIncludeGraph(fixture.plan)
	nodes, _ := graph.Nodes()
	input := append([]uuid.UUID{}, includeCandidateIDs...)
	statement, err := compiler.CompileIncludeTarget(fixture.plan, graph, nodes[0], input)
	if err != nil {
		t.Fatal(err)
	}
	wantSQL := `SELECT "fld_a_id", "fld_a_label" FROM "frm_include_a" ` +
		`WHERE "fld_a_id" = ANY(CAST($1 AS pg_catalog.uuid[])) AND ` +
		`("fld_a_label" COLLATE pg_catalog.pg_unicode_fast <> $2) AND ` +
		`("fld_a_label" COLLATE pg_catalog.pg_unicode_fast = $3) ` +
		`ORDER BY "fld_a_label" COLLATE pg_catalog.pg_unicode_fast ASC NULLS LAST, ` +
		`"fld_a_id" ASC NULLS LAST`
	sql, sqlErr := statement.SQL()
	explain, explainErr := statement.ExplainSQL()
	arguments, argumentsErr := statement.Arguments()
	columns, columnsErr := statement.Columns()
	stableIndex, stableErr := statement.StableIDColumnIndex()
	class, classErr := statement.Class()
	owner, ownerErr := statement.Owner()
	wantIDs := []uuid.UUID{includeCandidateIDs[1], includeCandidateIDs[2], includeCandidateIDs[0]}
	if sqlErr != nil || explainErr != nil || argumentsErr != nil || columnsErr != nil ||
		stableErr != nil || classErr != nil || ownerErr != nil || sql != wantSQL ||
		explain != explainJSONPrefix+wantSQL || len(arguments) != 3 ||
		!slices.Equal(arguments[0].([]uuid.UUID), wantIDs) ||
		arguments[1] != "blocked" || arguments[2] != includeInjectionValue ||
		len(columns) != 2 || stableIndex != 0 || class != planner.StatementClassInclude ||
		owner != "crm" || strings.Contains(sql, includeInjectionValue) ||
		strings.Contains(sql, " LIMIT ") ||
		statement.ValidatePlan(fixture.plan, graph, nodes[0]) != nil {
		t.Fatal("include-target statement contract mismatch")
	}
	assertIncludeTargetColumns(t, columns, fixture)
	input[0] = uuid.Nil
	arguments[0].([]uuid.UUID)[0] = uuid.Nil
	columns[0].outputKey = "mutated"
	againArgs, _ := statement.Arguments()
	againColumns, _ := statement.Columns()
	if !slices.Equal(againArgs[0].([]uuid.UUID), wantIDs) ||
		againColumns[0].outputKey != "id" || statement.Validate() != nil {
		t.Fatal("include-target ownership escaped")
	}
}

func assertIncludeTargetColumns(
	t *testing.T, columns []ResultColumn, fixture includePlanFixture,
) {
	t.Helper()
	id, _ := columns[0].FieldID()
	idKey, idVisible, idKeyErr := columns[0].OutputKey()
	idOrdinal, idSorted, idSortErr := columns[0].SortOrdinal()
	label, _ := columns[1].FieldID()
	labelKey, labelVisible, labelKeyErr := columns[1].OutputKey()
	labelOrdinal, labelSorted, labelSortErr := columns[1].SortOrdinal()
	wantID, _ := fixture.catalogs.aID.ID()
	wantLabel, _ := fixture.catalogs.aLabel.ID()
	if idKeyErr != nil || idSortErr != nil || labelKeyErr != nil || labelSortErr != nil ||
		id != wantID || idKey != "id" || !idVisible || idOrdinal != 1 || !idSorted ||
		label != wantLabel || labelKey != "" || labelVisible ||
		labelOrdinal != 0 || !labelSorted {
		t.Fatal("include-target result columns mismatch")
	}
}
