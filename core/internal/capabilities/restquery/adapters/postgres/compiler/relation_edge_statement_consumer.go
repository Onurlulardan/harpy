package compiler

import "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"

type RelationEdgeStatementConsumer interface {
	ConsumeRelationEdge(RelationEdgeStatement) error
}

func ConsumeRelationEdgeStatement(
	statement RelationEdgeStatement,
	plan resolved.AuthorizedPlan,
	graph IncludeGraph,
	node IncludeNode,
	consumer RelationEdgeStatementConsumer,
) error {
	if statement.ValidatePlan(plan, graph, node) != nil ||
		includeConsumerIsNil(consumer) {
		return ErrContract
	}
	return consumer.ConsumeRelationEdge(statement)
}
