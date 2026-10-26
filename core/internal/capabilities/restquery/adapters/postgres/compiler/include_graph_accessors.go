package compiler

import "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"

func (graph IncludeGraph) Nodes() ([]IncludeNode, error) {
	if graph.Validate() != nil {
		return nil, ErrContract
	}
	return append([]IncludeNode{}, graph.nodes...), nil
}

func (graph IncludeGraph) ValidatePlan(plan resolved.AuthorizedPlan) error {
	if graph.Validate() != nil {
		return ErrContract
	}
	expected, err := buildIncludeGraph(plan)
	if err != nil || expected.stateSeal != graph.stateSeal {
		return ErrContract
	}
	return nil
}

func validateGraphNode(
	graph IncludeGraph, node IncludeNode,
) error {
	if graph.Validate() != nil || node.Validate() != nil ||
		node.ordinal < 0 || node.ordinal >= len(graph.nodes) ||
		graph.nodes[node.ordinal].stateSeal != node.stateSeal {
		return ErrContract
	}
	return nil
}
