package executor

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/compiler"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
)

type statement struct {
	sql        string
	explainSQL string
	arguments  []any
	class      planner.StatementClass
}

func statementsFor(batch consumedBatch) ([]statement, error) {
	if batch.Validate() != nil {
		return nil, ErrContract
	}
	record, err := recordStatement(batch.record)
	if err != nil {
		return nil, err
	}
	result := []statement{record}
	if batch.hasExactCount {
		count, err := countStatement(batch.exactCount)
		if err != nil {
			return nil, err
		}
		result = append(result, count)
	}
	return result, nil
}

func recordStatement(value compiler.RecordStatement) (statement, error) {
	sql, sqlErr := value.SQL()
	explain, explainErr := value.ExplainSQL()
	arguments, argumentsErr := value.Arguments()
	class, classErr := value.Class()
	if sqlErr != nil || explainErr != nil || argumentsErr != nil || classErr != nil ||
		class != planner.StatementClassRecord {
		return statement{}, ErrContract
	}
	return statement{sql: sql, explainSQL: explain, arguments: arguments, class: class}, nil
}

func countStatement(value compiler.ExactCountStatement) (statement, error) {
	sql, sqlErr := value.SQL()
	explain, explainErr := value.ExplainSQL()
	arguments, argumentsErr := value.Arguments()
	class, classErr := value.Class()
	if sqlErr != nil || explainErr != nil || argumentsErr != nil || classErr != nil ||
		class != planner.StatementClassExactCount {
		return statement{}, ErrContract
	}
	return statement{sql: sql, explainSQL: explain, arguments: arguments, class: class}, nil
}
