package compiler

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/google/uuid"
)

func TestCompileIncludeGraphSealsBreadthFirstShape(t *testing.T) {
	fixture := newIncludePlanFixture(t)
	graph, err := includeCompiler(t).CompileIncludeGraph(fixture.plan)
	if err != nil {
		t.Fatal(err)
	}
	nodes, err := graph.Nodes()
	if err != nil || len(nodes) != 3 || graph.ValidatePlan(fixture.plan) != nil {
		t.Fatal("include graph did not round trip")
	}
	expected := []struct {
		name, source, owner, path string
		depth, parent, sibling    int
	}{
		{"children_a", "forms", "crm", "/include/0", 1, -1, 0},
		{"children_b", "forms", "iam", "/include/1", 1, -1, 1},
		{"grandchildren", "crm", "forms", "/include/0/include/0", 2, 0, 0},
	}
	for index, want := range expected {
		node := nodes[index]
		name, nameErr := node.RelationName()
		source, sourceErr := node.SourceOwner()
		owner, ownerErr := node.Owner()
		path, pathErr := node.Path()
		pathValue, valueErr := path.Value()
		depth, depthErr := node.Depth()
		ordinal, ordinalErr := node.Ordinal()
		parent, parentErr := node.ParentOrdinal()
		sibling, siblingErr := node.SiblingOrdinal()
		cardinality, cardinalityErr := node.Cardinality()
		limit, limitErr := node.Limit()
		if nameErr != nil || sourceErr != nil || ownerErr != nil || pathErr != nil ||
			valueErr != nil || depthErr != nil || ordinalErr != nil || parentErr != nil ||
			siblingErr != nil || cardinalityErr != nil || limitErr != nil ||
			name != want.name || source != want.source || owner != want.owner ||
			pathValue != want.path || depth != want.depth || ordinal != index ||
			parent != want.parent || sibling != want.sibling ||
			cardinality != api.RelationCardinalityToMany || limit != 2 {
			t.Fatalf("node %d shape mismatch", index)
		}
	}
}

func TestIncludeGraphOwnsNodesAndRejectsTransplant(t *testing.T) {
	fixture := newIncludePlanFixture(t)
	graph, _ := includeCompiler(t).CompileIncludeGraph(fixture.plan)
	nodes, _ := graph.Nodes()
	nodes[0].relationName = "mutated"
	again, _ := graph.Nodes()
	if name, _ := again[0].RelationName(); name != "children_a" || graph.Validate() != nil {
		t.Fatal("graph node ownership escaped")
	}
	other := newIncludePlanFixtureForTenant(
		t, uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaab"),
	)
	if !errors.Is(graph.ValidatePlan(other.plan), ErrContract) {
		t.Fatal("cross-plan graph transplant was accepted")
	}
	otherGraph, _ := includeCompiler(t).CompileIncludeGraph(other.plan)
	otherNodes, _ := otherGraph.Nodes()
	if !errors.Is(validateGraphNode(graph, otherNodes[0]), ErrContract) {
		t.Fatal("cross-graph node transplant was accepted")
	}
}

func TestIncludeGraphTamperAndSerializationFailClosed(t *testing.T) {
	fixture := newIncludePlanFixture(t)
	graph, _ := includeCompiler(t).CompileIncludeGraph(fixture.plan)
	nodes, _ := graph.Nodes()
	tampered := graph
	tampered.stateSeal[0] ^= 0xff
	if !errors.Is(tampered.Validate(), ErrContract) {
		t.Fatal("graph seal tamper was accepted")
	}
	node := nodes[0]
	node.depth++
	if !errors.Is(node.Validate(), ErrContract) {
		t.Fatal("node shape tamper was accepted")
	}
	for _, value := range []any{graph, nodes[0]} {
		if fmt.Sprint(value) != compilerRedaction {
			t.Fatal("include capability formatting leaked")
		}
		if document, err := json.Marshal(value); err == nil || document != nil {
			t.Fatal("include capability JSON did not fail closed")
		}
	}
}

func TestIncludeGraphRejectsBudgetsAndOrdinalCollision(t *testing.T) {
	fixture := newIncludePlanFixture(t)
	graph, _ := includeCompiler(t).CompileIncludeGraph(fixture.plan)
	nodes, _ := graph.Nodes()
	if _, err := newIncludeGraph(nil, graph.binding); !errors.Is(err, ErrContract) {
		t.Fatal("empty include graph was accepted")
	}
	if _, err := newIncludeGraph(
		[]IncludeNode{nodes[0], nodes[1], nodes[2], nodes[2]}, graph.binding,
	); !errors.Is(err, ErrContract) {
		t.Fatal("graph above three-edge/four-node budget was accepted")
	}
	if _, err := newIncludeGraph(
		[]IncludeNode{nodes[0], nodes[0]}, graph.binding,
	); !errors.Is(err, ErrContract) {
		t.Fatal("colliding include ordinal was accepted")
	}
}
