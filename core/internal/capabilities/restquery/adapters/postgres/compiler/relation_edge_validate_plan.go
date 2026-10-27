package compiler

import "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"

func (statement RelationEdgeStatement) ValidatePlan(
	plan resolved.AuthorizedPlan, graph IncludeGraph, node IncludeNode,
) error {
	if statement.Validate() != nil ||
		statement.binding.validatePlan(plan, graph, node) != nil ||
		statement.catalog != node.sourceCatalog {
		return ErrContract
	}
	return nil
}
