package compiler

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
)

func (statement IncludeTargetStatement) SQL() (string, error) {
	if statement.Validate() != nil {
		return "", ErrContract
	}
	return statement.sql, nil
}

func (statement IncludeTargetStatement) ExplainSQL() (string, error) {
	if statement.Validate() != nil {
		return "", ErrContract
	}
	return explainJSONPrefix + statement.sql, nil
}

func (statement IncludeTargetStatement) Arguments() ([]any, error) {
	if statement.Validate() != nil {
		return nil, ErrContract
	}
	return cloneArguments(statement.arguments), nil
}

func (statement IncludeTargetStatement) Columns() ([]ResultColumn, error) {
	if statement.Validate() != nil {
		return nil, ErrContract
	}
	return append([]ResultColumn{}, statement.columns...), nil
}

func (statement IncludeTargetStatement) StableIDColumnIndex() (int, error) {
	if statement.Validate() != nil {
		return 0, ErrContract
	}
	return statement.stableIndex, nil
}

func (statement IncludeTargetStatement) Class() (planner.StatementClass, error) {
	if statement.Validate() != nil {
		return 0, ErrContract
	}
	return statement.class, nil
}

func (statement IncludeTargetStatement) Catalog() (api.CatalogReference, error) {
	if statement.Validate() != nil {
		return api.CatalogReference{}, ErrContract
	}
	return statement.catalog, nil
}

func (statement IncludeTargetStatement) Owner() (string, error) {
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
