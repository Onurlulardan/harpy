package executor

import (
	"errors"
	"testing"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func TestGraphLifecycleMaterializesCanonicalGraphAtomically(t *testing.T) {
	fixture := newGraphFakeFixture(t, api.CountModeExact)
	result, err := fixture.execute(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	page, err := result.Page()
	if err != nil {
		t.Fatal(err)
	}
	records, err := page.Records()
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		`{"label":"root-one","_included":{"children":[{"label":"alpha","_included":{"items":[{"label":"item-alpha","_included":{}}]}}],"owner":{"label":"owner-one","_included":{}}}}`,
		`{"label":"root-two","_included":{"children":[{"label":"middle","_included":{"items":[{"label":"item-middle","_included":{}}]}}],"owner":null}}`,
	}
	if len(records) != len(want) {
		t.Fatalf("record count = %d", len(records))
	}
	for index := range want {
		if string(records[index]) != want[index] {
			t.Fatalf("record[%d] = %s", index, records[index])
		}
	}
	total, present, err := result.ExactCount()
	if err != nil || !present || total != 7 {
		t.Fatalf("exact count = %d/%v, err = %v", total, present, err)
	}
	assertSiblingExplainsPrecedeData(t, fixture)
}

func TestGraphLifecycleStageRejectionReturnsNoPartialResult(t *testing.T) {
	fixture := newGraphFakeFixture(t, api.CountModeExact)
	ownerIndex := fixture.nodeIndex(t, "owner")
	statement := fixture.targets["owner"]
	fixture.drivers[ownerIndex+1].rowValues[statement.explainSQL] =
		explainDocument("999999999")
	result, err := fixture.execute(t.Context())
	if err == nil || result.Validate() == nil {
		t.Fatalf("result valid = %v, err = %v", result.Validate() == nil, err)
	}
	calls := fixture.log.snapshot()
	for _, relation := range []string{"children", "owner"} {
		if graphCallIndex(calls, "row", fixture.targets[relation]) < 0 ||
			graphCallIndex(calls, "query", fixture.targets[relation]) >= 0 {
			t.Fatalf("stage %s calls = %q", relation, calls)
		}
	}
}

func TestGraphLifecycleDataFailureStopsSiblingAndReturnsNoResult(t *testing.T) {
	fixture := newGraphFakeFixture(t, api.CountModeNone)
	childrenIndex := fixture.nodeIndex(t, "children")
	children := fixture.targets["children"]
	fixture.drivers[childrenIndex+1].queryErrors[children.sql] = errors.New("driver detail")
	result, err := fixture.execute(t.Context())
	if err == nil || result.Validate() == nil {
		t.Fatalf("result valid = %v, err = %v", result.Validate() == nil, err)
	}
	calls := fixture.log.snapshot()
	if graphCallIndex(calls, "query", children) < 0 ||
		graphCallIndex(calls, "query", fixture.targets["owner"]) >= 0 {
		t.Fatalf("failed stage calls = %q", calls)
	}
}

func TestGraphLifecycleRejectsTwoVisibleToOneTargets(t *testing.T) {
	fixture := newGraphFakeFixture(t, api.CountModeNone)
	ids := fixture.ids
	edge := fixture.edges["owner"]
	fixture.drivers[0].rows[edge.sql] = &fakeRows{values: [][]any{
		{ids.rootOne, ids.owner, int16(1)},
		{ids.rootOne, ids.owner2, int16(2)},
	}}
	ownerIndex := fixture.nodeIndex(t, "owner")
	fixture.prepareTarget(
		t, fixture.drivers[ownerIndex+1], ownerIndex,
		[]uuidValue{ids.owner, ids.owner2},
		[]graphTargetRow{{"owner-one", ids.owner}, {"owner-two", ids.owner2}},
	)
	result, err := fixture.execute(t.Context())
	var queryError api.QueryError
	if result.Validate() == nil || !errors.As(err, &queryError) ||
		queryError.Code() != restqueryv1.ErrorRelationCardinalityViolation {
		t.Fatalf("result=%v error=%v code=%s", result.Validate(), err, queryError.Code())
	}
}

func assertSiblingExplainsPrecedeData(t *testing.T, fixture *graphFakeFixture) {
	t.Helper()
	calls := fixture.log.snapshot()
	childrenExplain := graphCallIndex(calls, "row", fixture.targets["children"])
	ownerExplain := graphCallIndex(calls, "row", fixture.targets["owner"])
	childrenData := graphCallIndex(calls, "query", fixture.targets["children"])
	ownerData := graphCallIndex(calls, "query", fixture.targets["owner"])
	if childrenExplain < 0 || ownerExplain < 0 || childrenData < 0 || ownerData < 0 ||
		childrenExplain > childrenData || ownerExplain > childrenData ||
		childrenExplain > ownerData || ownerExplain > ownerData {
		t.Fatalf("sibling stage calls = %q", calls)
	}
}

func (fixture *graphFakeFixture) nodeIndex(t *testing.T, relation string) int {
	t.Helper()
	for index, node := range fixture.runtime.nodes {
		name, _ := node.compiled.RelationName()
		if name == relation {
			return index
		}
	}
	t.Fatalf("relation %s missing", relation)
	return -1
}
