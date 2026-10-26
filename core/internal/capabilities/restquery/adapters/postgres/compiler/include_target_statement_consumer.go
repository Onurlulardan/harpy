package compiler

import "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"

type IncludeTargetStatementConsumer interface {
	ConsumeIncludeTarget(IncludeTargetStatement) error
}

func ConsumeIncludeTargetStatement(
	statement IncludeTargetStatement,
	plan resolved.AuthorizedPlan,
	graph IncludeGraph,
	node IncludeNode,
	consumer IncludeTargetStatementConsumer,
) error {
	if statement.ValidatePlan(plan, graph, node) != nil ||
		includeConsumerIsNil(consumer) {
		return ErrContract
	}
	return consumer.ConsumeIncludeTarget(statement)
}
