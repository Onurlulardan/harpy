package executor

import (
	"errors"
	"testing"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func TestMaterializeToOneRejectsTwoVisibleTargets(t *testing.T) {
	rootID, oneID, twoID := materialUUID(20), materialUUID(21), materialUUID(22)
	page, roots := materialRootPage(t, []uuidValue{rootID})
	graph := materialGraph{root: roots, nodes: []materialNode{{
		relation: "owner", cardinality: api.RelationCardinalityToOne,
		limit: 1, parent: -1, depth: 1, ordinal: 0,
		records: []materialRecord{
			materialFixtureRecord(t, oneID, []byte(`{"name":"one"}`)),
			materialFixtureRecord(t, twoID, []byte(`{"name":"two"}`)),
		},
		associations: map[uuidValue]map[uuidValue]struct{}{
			rootID: materialTargets(oneID, twoID),
		},
	}}}
	result, err := materializeGraph(page, graph)
	var queryError api.QueryError
	if result.Validate() == nil || !errors.As(err, &queryError) ||
		queryError.Code() != restqueryv1.ErrorRelationCardinalityViolation {
		t.Fatalf("result=%v error=%v code=%s", result.Validate(), err, queryError.Code())
	}
}

func TestMaterializeToOneUsesNullWhenTargetIsInvisible(t *testing.T) {
	rootID := materialUUID(30)
	page, roots := materialRootPage(t, []uuidValue{rootID})
	graph := materialGraph{root: roots, nodes: []materialNode{{
		relation: "owner", cardinality: api.RelationCardinalityToOne,
		limit: 1, parent: -1, depth: 1, ordinal: 0,
		records: []materialRecord{},
		associations: map[uuidValue]map[uuidValue]struct{}{
			rootID: materialTargets(materialUUID(31)),
		},
	}}}
	result, err := materializeGraph(page, graph)
	records, recordsErr := result.Records()
	if err != nil || recordsErr != nil ||
		string(records[0]) != `{"name":"root-0","metadata":null,"_included":{"owner":null}}` {
		t.Fatalf("records=%q errors=%v/%v", records, err, recordsErr)
	}
}
