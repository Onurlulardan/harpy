package executor

import (
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func TestMaterializeGraphUsesRequestOrderAndNestedFinalMembers(t *testing.T) {
	rootID, parentID, childID := materialUUID(1), materialUUID(2), materialUUID(3)
	missingID := materialUUID(4)
	page, roots := materialRootPage(t, []uuidValue{rootID})
	parent := materialFixtureRecord(t, parentID, []byte(`{"name":"parent"}`))
	child := materialFixtureRecord(t, childID, []byte(`{"name":"child"}`))
	graph := materialGraph{root: roots, nodes: []materialNode{
		{
			relation: "zeta", cardinality: api.RelationCardinalityToMany,
			limit: 1, parent: -1, depth: 1, ordinal: 0,
			records: []materialRecord{parent},
			associations: map[uuidValue]map[uuidValue]struct{}{
				rootID: materialTargets(missingID, parentID),
			},
		},
		{
			relation: "items", cardinality: api.RelationCardinalityToMany,
			limit: 5, parent: 0, depth: 2, ordinal: 1,
			records: []materialRecord{child},
			associations: map[uuidValue]map[uuidValue]struct{}{
				parentID: materialTargets(childID),
			},
		},
		{
			relation: "alpha", cardinality: api.RelationCardinalityToOne,
			limit: 1, parent: -1, depth: 1, ordinal: 2,
			records:      []materialRecord{},
			associations: map[uuidValue]map[uuidValue]struct{}{},
		},
	}}
	result, err := materializeGraph(page, graph)
	if err != nil {
		t.Fatal(err)
	}
	records, _ := result.Records()
	want := `{"name":"root-0","metadata":null,"_included":` +
		`{"zeta":[{"name":"parent","_included":{"items":` +
		`[{"name":"child","_included":{}}]}}],"alpha":null}}`
	if len(records) != 1 || string(records[0]) != want {
		t.Fatalf("record=%s", records[0])
	}
	if strings.Index(want, `"zeta"`) > strings.Index(want, `"alpha"`) {
		t.Fatal("relation request order was not preserved")
	}
}

func TestMaterialRecordSealRejectsHiddenMaterialTamper(t *testing.T) {
	record := materialFixtureRecord(t, materialUUID(10), []byte(`{"name":"safe"}`))
	record.order[0] = []byte(`"tampered"`)
	if record.validate() == nil {
		t.Fatal("hidden effective sort tamper was accepted")
	}
}
