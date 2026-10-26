package compiler

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func apiPlanZero() resolved.AuthorizedPlan { return resolved.AuthorizedPlan{} }

func fixtureBoundStatement(
	testingT testing.TB, owner, outputKey string,
) RecordStatement {
	testingT.Helper()
	field := fixtureField(testingT,
		"66666666-6666-4666-8666-666666666666", "display", "fld_display",
		fixtureFoldedTextType(testingT, false), "fld_display_folded", "",
		[]api.Operator{}, false, false,
	)
	stable := fixtureField(testingT,
		"77777777-7777-4777-8777-777777777777", "id", "fld_id",
		fixtureUUIDType(testingT), "", "", []api.Operator{}, true, false,
	)
	query := fixtureRecordQuery(testingT,
		[]resolved.Projection{fixtureProjection(testingT, field, outputKey, false)}, nil,
		[]resolved.Sort{fixtureSort(
			testingT, stable, api.DirectionAscending, true,
		)}, nil, 1,
	)
	statement, err := compileRecord(fixtureRecordInput(
		testingT, owner, query, fixtureAllVisibilityOwner(testingT, owner),
	))
	if err != nil {
		testingT.Fatal(err)
	}
	return statement
}
