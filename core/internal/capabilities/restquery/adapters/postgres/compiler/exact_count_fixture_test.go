package compiler

import (
	"crypto/sha256"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func fixtureExactCountScenario(
	testingT testing.TB,
) (recordInput, resolved.RecordQuery) {
	testingT.Helper()
	name := fixtureField(testingT,
		"11111111-2222-4333-8444-555555555555", "name", "fld_name",
		fixtureFoldedTextType(testingT, false), "fld_name_folded", "",
		[]api.Operator{api.OperatorContains}, false, true,
	)
	owner := fixtureField(testingT,
		"22222222-3333-4444-8555-666666666666", "owner", "fld_owner",
		fixtureUUIDType(testingT), "", "", []api.Operator{api.OperatorEq}, false, false,
	)
	stable := fixtureField(testingT,
		"33333333-4444-4555-8666-777777777777", "id", "fld_id",
		fixtureUUIDType(testingT), "", "", []api.Operator{}, true, false,
	)
	where := fixturePredicateFilter(
		testingT, name, api.OperatorContains,
		[]resolved.Value{fixtureScalar(testingT, api.LogicalTypeText, "term%_")},
	)
	search, err := resolved.NewSearchTerm(
		name, fixtureScalar(testingT, api.LogicalTypeText, "needle"),
		fixturePath(testingT, "/search"),
	)
	if err != nil {
		testingT.Fatal(err)
	}
	page, err := resolved.NewPage(2, nil)
	if err != nil {
		testingT.Fatal(err)
	}
	query, err := resolved.NewRecordQuery(
		[]resolved.Projection{fixtureProjection(testingT, name, "name", false)},
		&where, []resolved.Sort{fixtureSort(
			testingT, stable, api.DirectionAscending, true,
		)}, []resolved.SearchTerm{search}, page, api.CountModeExact,
		fixturePath(testingT, "/count"), true, []resolved.Include{},
	)
	if err != nil {
		testingT.Fatal(err)
	}
	mandatoryPredicate, err := resolved.NewMandatoryPredicate(
		owner, api.OperatorEq, []resolved.Value{fixtureScalar(
			testingT, api.LogicalTypeUUID,
			"44444444-5555-4666-8777-888888888888",
		)},
	)
	if err != nil {
		testingT.Fatal(err)
	}
	mandatory, err := resolved.NewMandatoryPredicateFilter(mandatoryPredicate)
	if err != nil {
		testingT.Fatal(err)
	}
	visibility, err := resolved.NewPredicateResourceVisibility(
		sha256.Sum256([]byte("requirement")), fixtureReference(testingT, "forms"),
		fixturePath(testingT, ""), mandatory,
	)
	if err != nil {
		testingT.Fatal(err)
	}
	return fixtureRecordInput(testingT, "forms", query, visibility), query
}

func fixtureExactCountStatement(testingT testing.TB) ExactCountStatement {
	testingT.Helper()
	input, _ := fixtureExactCountScenario(testingT)
	statement, err := compileExactCount(input)
	if err != nil {
		testingT.Fatal(err)
	}
	return statement
}
