package compiler

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
)

func (statement AggregateStatement) SQL() (string, error) {
	if statement.Validate() != nil {
		return "", ErrContract
	}
	return statement.sql, nil
}

func (statement AggregateStatement) ExplainSQL() (string, error) {
	if statement.Validate() != nil {
		return "", ErrContract
	}
	return explainJSONPrefix + statement.sql, nil
}

func (statement AggregateStatement) Arguments() ([]any, error) {
	if statement.Validate() != nil {
		return nil, ErrContract
	}
	return cloneArguments(statement.arguments), nil
}

func (statement AggregateStatement) Groups() ([]AggregateGroupColumn, error) {
	if statement.Validate() != nil {
		return nil, ErrContract
	}
	return append([]AggregateGroupColumn{}, statement.groups...), nil
}

func (statement AggregateStatement) Measures() ([]AggregateMeasureColumn, error) {
	if statement.Validate() != nil {
		return nil, ErrContract
	}
	return append([]AggregateMeasureColumn{}, statement.measures...), nil
}

func (statement AggregateStatement) Limit() (int, error) {
	if statement.Validate() != nil {
		return 0, ErrContract
	}
	return statement.limit, nil
}

func (statement AggregateStatement) Class() (planner.StatementClass, error) {
	if statement.Validate() != nil {
		return 0, ErrContract
	}
	return statement.class, nil
}
