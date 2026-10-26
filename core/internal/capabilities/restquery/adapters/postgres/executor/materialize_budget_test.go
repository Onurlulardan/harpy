package executor

import (
	"errors"
	"testing"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func TestMaterializeGraphRejectsOneThousandAndFirstOccurrenceWithoutPage(t *testing.T) {
	rootIDs := make([]uuidValue, 11)
	for index := range rootIDs {
		rootIDs[index] = materialUUID(100 + index)
	}
	page, roots := materialRootPage(t, rootIDs)
	children := make([]materialRecord, 100)
	targetIDs := make([]uuidValue, 100)
	for index := range children {
		targetIDs[index] = materialUUID(1000 + index)
		children[index] = materialFixtureRecord(
			t, targetIDs[index], []byte(`{"value":1}`),
		)
	}
	associations := make(map[uuidValue]map[uuidValue]struct{}, len(rootIDs))
	for _, rootID := range rootIDs {
		associations[rootID] = materialTargets(targetIDs...)
	}
	graph := materialGraph{root: roots, nodes: []materialNode{{
		relation: "children", cardinality: api.RelationCardinalityToMany,
		limit: 100, parent: -1, depth: 1, ordinal: 0,
		records: children, associations: associations,
	}}}
	result, err := materializeGraph(page, graph)
	var queryError api.QueryError
	if result.Validate() == nil || !errors.As(err, &queryError) ||
		queryError.Code() != restqueryv1.ErrorQueryBudgetExceeded {
		t.Fatalf("result=%v error=%v code=%s", result.Validate(), err, queryError.Code())
	}
}
