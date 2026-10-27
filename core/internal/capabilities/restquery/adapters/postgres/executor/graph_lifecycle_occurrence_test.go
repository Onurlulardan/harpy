package executor

import (
	"errors"
	"fmt"
	"testing"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func TestGraphLifecycleRejectsOneThousandAndFirstOccurrenceWithoutResult(t *testing.T) {
	fixture := newGraphFakeRuntimeFixture(t, api.CountModeNone, 500)
	prepareOccurrenceOverflowData(t, fixture)
	result, err := fixture.execute(t.Context())
	var queryError api.QueryError
	if result.Validate() == nil || !errors.As(err, &queryError) ||
		queryError.Code() != restqueryv1.ErrorQueryBudgetExceeded {
		t.Fatalf("result=%v error=%v code=%s", result.Validate(), err, queryError.Code())
	}
}

func prepareOccurrenceOverflowData(t *testing.T, fixture *graphFakeFixture) {
	t.Helper()
	ids := fixture.ids
	roots := make([]uuidValue, 500)
	rootRows := make([][]any, len(roots))
	rootColumns, _ := fixture.runtime.batch.record.Columns()
	for index := range roots {
		roots[index] = materialUUID(10_000 + index)
		rootRows[index] = graphMaterialRow(
			t, rootColumns, fmt.Sprintf("root-%03d", index), roots[index],
		)
	}
	rootStatements, _ := statementsFor(fixture.runtime.batch)
	fixture.drivers[0].rows[rootStatements[0].sql] = &fakeRows{values: rootRows}
	fixture.explain(fixture.drivers[0], rootStatements...)
	for index := range fixture.runtime.nodes {
		relation, _ := fixture.runtime.nodes[index].compiled.RelationName()
		sources, edges, candidates, targetRows := overflowRelationData(
			relation, roots, ids,
		)
		fixture.prepareEdge(
			t, fixture.edgeDriver(index), index, sources, edges,
		)
		fixture.prepareTarget(
			t, fixture.drivers[index+1], index, candidates, targetRows,
		)
	}
}

func overflowRelationData(
	relation string,
	roots []uuidValue,
	ids graphLifecycleIDs,
) ([]uuidValue, [][]any, []uuidValue, []graphTargetRow) {
	switch relation {
	case "children":
		return roots, repeatedEdges(roots, ids.alpha),
			[]uuidValue{ids.alpha}, []graphTargetRow{{"alpha", ids.alpha}}
	case "owner":
		return roots, repeatedEdges(roots, ids.owner),
			[]uuidValue{ids.owner}, []graphTargetRow{{"owner-one", ids.owner}}
	case "items":
		return []uuidValue{ids.alpha},
			[][]any{{ids.alpha, ids.itemAlpha, int16(1)}},
			[]uuidValue{ids.itemAlpha},
			[]graphTargetRow{{"item-alpha", ids.itemAlpha}}
	default:
		return nil, nil, nil, nil
	}
}

func repeatedEdges(sources []uuidValue, target uuidValue) [][]any {
	rows := make([][]any, len(sources))
	for index, source := range sources {
		rows[index] = []any{source, target, int16(1)}
	}
	return rows
}
