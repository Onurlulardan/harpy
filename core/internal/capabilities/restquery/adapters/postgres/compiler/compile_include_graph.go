package compiler

import "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"

func (compiler Compiler) CompileIncludeGraph(
	plan resolved.AuthorizedPlan,
) (IncludeGraph, error) {
	if compiler.Validate() != nil {
		return IncludeGraph{}, ErrContract
	}
	return buildIncludeGraph(plan)
}

func buildIncludeGraph(plan resolved.AuthorizedPlan) (IncludeGraph, error) {
	inputs, binding, err := collectIncludeNodeInputs(plan)
	if err != nil {
		return IncludeGraph{}, ErrContract
	}
	nodes := make([]IncludeNode, len(inputs))
	for index, input := range inputs {
		nodes[index], err = newIncludeNode(input, binding)
		if err != nil {
			return IncludeGraph{}, ErrContract
		}
	}
	return newIncludeGraph(nodes, binding)
}

func includeInputForPlan(
	plan resolved.AuthorizedPlan, graph IncludeGraph, node IncludeNode,
) (includeNodeInput, error) {
	if graph.ValidatePlan(plan) != nil || validateGraphNode(graph, node) != nil {
		return includeNodeInput{}, ErrContract
	}
	inputs, binding, err := collectIncludeNodeInputs(plan)
	if err != nil || binding != graph.binding || node.ordinal >= len(inputs) {
		return includeNodeInput{}, ErrContract
	}
	expected, err := newIncludeNode(inputs[node.ordinal], binding)
	if err != nil || expected.stateSeal != node.stateSeal {
		return includeNodeInput{}, ErrContract
	}
	return inputs[node.ordinal], nil
}
