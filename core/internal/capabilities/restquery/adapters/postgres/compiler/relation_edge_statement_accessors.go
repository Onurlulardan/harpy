package compiler

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
)

func (statement RelationEdgeStatement) SQL() (string, error) {
	if statement.Validate() != nil {
		return "", ErrContract
	}
	return statement.sql, nil
}

func (statement RelationEdgeStatement) ExplainSQL() (string, error) {
	if statement.Validate() != nil {
		return "", ErrContract
	}
	return explainJSONPrefix + statement.sql, nil
}

func (statement RelationEdgeStatement) Arguments() ([]any, error) {
	if statement.Validate() != nil {
		return nil, ErrContract
	}
	return cloneArguments(statement.arguments), nil
}

func (statement RelationEdgeStatement) Class() (planner.StatementClass, error) {
	if statement.Validate() != nil {
		return 0, ErrContract
	}
	return statement.class, nil
}

func (statement RelationEdgeStatement) Catalog() (api.CatalogReference, error) {
	if statement.Validate() != nil {
		return api.CatalogReference{}, ErrContract
	}
	return statement.catalog, nil
}

func (statement RelationEdgeStatement) Owner() (string, error) {
	catalog, err := statement.Catalog()
	if err != nil {
		return "", ErrContract
	}
	owner, err := catalog.Owner()
	if err != nil {
		return "", ErrContract
	}
	return owner, nil
}

func (statement RelationEdgeStatement) ColumnCount() (int, error) {
	if statement.Validate() != nil {
		return 0, ErrContract
	}
	return 3, nil
}
