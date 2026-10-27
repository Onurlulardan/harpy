package executor

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/compiler"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

type relationEdgeConsumer struct {
	statement compiler.RelationEdgeStatement
	consumed  bool
}

func (consumer *relationEdgeConsumer) ConsumeRelationEdge(
	statement compiler.RelationEdgeStatement,
) error {
	if consumer == nil || consumer.consumed || statement.Validate() != nil {
		return ErrContract
	}
	consumer.statement = statement
	consumer.consumed = true
	return nil
}

func consumeRelationEdge(
	compiled compiler.RelationEdgeStatement,
	plan resolved.AuthorizedPlan,
	graph compiler.IncludeGraph,
	node compiler.IncludeNode,
) (statement, error) {
	consumer := &relationEdgeConsumer{}
	if compiler.ConsumeRelationEdgeStatement(
		compiled, plan, graph, node, consumer,
	) != nil || !consumer.consumed {
		return statement{}, ErrContract
	}
	sql, sqlErr := consumer.statement.SQL()
	explain, explainErr := consumer.statement.ExplainSQL()
	arguments, argumentErr := consumer.statement.Arguments()
	class, classErr := consumer.statement.Class()
	columns, columnErr := consumer.statement.ColumnCount()
	if sqlErr != nil || explainErr != nil || argumentErr != nil ||
		classErr != nil || columnErr != nil || columns != 3 ||
		class != planner.StatementClassInclude {
		return statement{}, ErrContract
	}
	return statement{
		sql: sql, explainSQL: explain, arguments: arguments, class: class,
	}, nil
}
