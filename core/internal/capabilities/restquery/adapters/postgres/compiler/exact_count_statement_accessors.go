package compiler

import "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"

func (statement ExactCountStatement) SQL() (string, error) {
	if statement.Validate() != nil {
		return "", ErrContract
	}
	return statement.sql, nil
}

func (statement ExactCountStatement) Arguments() ([]any, error) {
	if statement.Validate() != nil {
		return nil, ErrContract
	}
	return cloneArguments(statement.arguments), nil
}

func (statement ExactCountStatement) Class() (planner.StatementClass, error) {
	if statement.Validate() != nil {
		return 0, ErrContract
	}
	return statement.class, nil
}
