package executor

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/compiler"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/recordpage"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/cursor"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/google/uuid"
)

type graphRuntimeNode struct {
	compiled     compiler.IncludeNode
	material     materialNode
	candidates   []uuid.UUID
	stageOrdinal int
	edgeOrdinal  int
}

type graphRuntime struct {
	plan     resolved.AuthorizedPlan
	graph    compiler.IncludeGraph
	compiler compiler.Compiler
	batch    consumedBatch
	codec    cursor.Codec
	policy   planner.Policy
	barrier  *graphStageBarrier
	nodes    []graphRuntimeNode
	page     recordpage.Page
	root     []materialRecord
	total    int64
	rootIDAt int
}

func newGraphRuntime(
	plan resolved.AuthorizedPlan,
	graph compiler.IncludeGraph,
	postgresCompiler compiler.Compiler,
	batch consumedBatch,
	codec cursor.Codec,
	policy planner.Policy,
	rootIDAt int,
) (*graphRuntime, error) {
	nodes, err := graph.Nodes()
	if err != nil || len(nodes) < 1 || len(nodes) > 3 || batch.Validate() != nil ||
		postgresCompiler.Validate() != nil ||
		codec.Validate() != nil || validatePolicy(policy) != nil || rootIDAt < 0 {
		return nil, ErrContract
	}
	runtime := &graphRuntime{
		plan: plan, graph: graph, compiler: postgresCompiler, batch: batch, codec: codec,
		policy: policy, nodes: make([]graphRuntimeNode, len(nodes)), rootIDAt: rootIDAt,
	}
	depthCounts := [3]int{}
	for index, node := range nodes {
		material, depth, parent, err := materialNodeFromCompiled(node, index)
		if err != nil {
			return nil, ErrContract
		}
		runtime.nodes[index] = graphRuntimeNode{
			compiled: node, material: material,
			stageOrdinal: depthCounts[depth], edgeOrdinal: -1,
			candidates: []uuid.UUID{},
		}
		depthCounts[depth]++
		_ = parent
	}
	edgeParents := runtime.assignEdgeOrdinals()
	expected := []int{1, 1, depthCounts[1]}
	if depthCounts[2] > 0 {
		if edgeParents < 1 {
			return nil, ErrContract
		}
		expected = append(expected, edgeParents, depthCounts[2])
	}
	runtime.barrier, err = newGraphStageBarrier(policy, expected)
	if err != nil {
		return nil, ErrContract
	}
	return runtime, nil
}

func materialNodeFromCompiled(
	node compiler.IncludeNode,
	index int,
) (materialNode, int, int, error) {
	relation, relationErr := node.RelationName()
	cardinality, cardinalityErr := node.Cardinality()
	limit, limitErr := node.Limit()
	depth, depthErr := node.Depth()
	ordinal, ordinalErr := node.Ordinal()
	parent, parentErr := node.ParentOrdinal()
	if relationErr != nil || cardinalityErr != nil || limitErr != nil ||
		depthErr != nil || ordinalErr != nil || parentErr != nil || ordinal != index {
		return materialNode{}, 0, 0, ErrContract
	}
	return materialNode{
		relation: relation, cardinality: cardinality, limit: limit,
		parent: parent, depth: depth, ordinal: ordinal,
		records:      []materialRecord{},
		associations: map[uuid.UUID]map[uuid.UUID]struct{}{},
	}, depth, parent, nil
}

func (runtime *graphRuntime) assignEdgeOrdinals() int {
	next := 0
	for index := range runtime.nodes {
		if runtime.nodes[index].material.depth != 1 ||
			len(runtime.children(index)) == 0 {
			continue
		}
		runtime.nodes[index].edgeOrdinal = next
		next++
	}
	return next
}

func (runtime *graphRuntime) children(parent int) []int {
	result := make([]int, 0, 3)
	for index, node := range runtime.nodes {
		if node.material.parent == parent {
			result = append(result, index)
		}
	}
	return result
}
