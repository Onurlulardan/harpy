package executor

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/compiler"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
)

func aggregateStatement(
	value compiler.AggregateStatement,
) (statement, error) {
	sql, sqlErr := value.SQL()
	explain, explainErr := value.ExplainSQL()
	arguments, argumentsErr := value.Arguments()
	class, classErr := value.Class()
	if sqlErr != nil || explainErr != nil || argumentsErr != nil || classErr != nil ||
		class != planner.StatementClassAggregate {
		return statement{}, ErrContract
	}
	return statement{
		sql: sql, explainSQL: explain, arguments: arguments, class: class,
	}, nil
}
