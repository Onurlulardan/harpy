package executor

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/compiler"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

type includeTargetConsumer struct {
	statement compiler.IncludeTargetStatement
	consumed  bool
}

func (consumer *includeTargetConsumer) ConsumeIncludeTarget(
	statement compiler.IncludeTargetStatement,
) error {
	if consumer == nil || consumer.consumed || statement.Validate() != nil {
		return ErrContract
	}
	consumer.statement = statement
	consumer.consumed = true
	return nil
}

func consumeIncludeTarget(
	compiled compiler.IncludeTargetStatement,
	plan resolved.AuthorizedPlan,
	graph compiler.IncludeGraph,
	node compiler.IncludeNode,
) (statement, compiler.IncludeTargetStatement, error) {
	consumer := &includeTargetConsumer{}
	if compiler.ConsumeIncludeTargetStatement(
		compiled, plan, graph, node, consumer,
	) != nil || !consumer.consumed {
		return statement{}, compiler.IncludeTargetStatement{}, ErrContract
	}
	sql, sqlErr := consumer.statement.SQL()
	explain, explainErr := consumer.statement.ExplainSQL()
	arguments, argumentErr := consumer.statement.Arguments()
	class, classErr := consumer.statement.Class()
	columns, columnErr := consumer.statement.Columns()
	stableIndex, stableErr := consumer.statement.StableIDColumnIndex()
	if sqlErr != nil || explainErr != nil || argumentErr != nil ||
		classErr != nil || columnErr != nil || stableErr != nil ||
		class != planner.StatementClassInclude || stableIndex < 0 ||
		stableIndex >= len(columns) {
		return statement{}, compiler.IncludeTargetStatement{}, ErrContract
	}
	return statement{
		sql: sql, explainSQL: explain, arguments: arguments, class: class,
	}, consumer.statement, nil
}
