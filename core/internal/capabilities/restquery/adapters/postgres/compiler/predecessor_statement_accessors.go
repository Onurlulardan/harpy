package compiler

import (
	"crypto/sha256"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/google/uuid"
)

func (statement PredecessorStatement) SQL() (string, error) {
	if statement.Validate() != nil || statement.binding.scope != resolved.SnapshotScopeTenant {
		return "", ErrContract
	}
	return statement.sql, nil
}

func (statement PredecessorStatement) ExplainSQL() (string, error) {
	if statement.Validate() != nil {
		return "", ErrContract
	}
	return explainJSONPrefix + statement.sql, nil
}

func (statement PredecessorStatement) Arguments() ([]any, error) {
	if statement.Validate() != nil {
		return nil, ErrContract
	}
	return cloneArguments(statement.arguments), nil
}

func (statement PredecessorStatement) Columns() ([]ResultColumn, error) {
	if statement.Validate() != nil {
		return nil, ErrContract
	}
	return append([]ResultColumn{}, statement.columns...), nil
}

func (statement PredecessorStatement) PredecessorIndex() (int64, error) {
	if statement.Validate() != nil {
		return 0, ErrContract
	}
	return statement.predecessorIndex, nil
}

func (statement PredecessorStatement) Class() (planner.StatementClass, error) {
	if statement.Validate() != nil {
		return 0, ErrContract
	}
	return statement.class, nil
}

func (statement PredecessorStatement) TenantID() (uuid.UUID, error) {
	if statement.Validate() != nil {
		return uuid.Nil, ErrContract
	}
	return statement.binding.scopeID, nil
}

func (statement PredecessorStatement) ActorID() (uuid.UUID, error) {
	if statement.Validate() != nil || statement.binding.scope != resolved.SnapshotScopeTenant {
		return uuid.Nil, ErrContract
	}
	return statement.binding.subject, nil
}

func (statement PredecessorStatement) Catalog() (api.CatalogReference, error) {
	if statement.Validate() != nil {
		return api.CatalogReference{}, ErrContract
	}
	return statement.binding.catalog, nil
}

func (statement PredecessorStatement) Owner() (string, error) {
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

func (statement PredecessorStatement) ContinuationSHA256() ([sha256.Size]byte, error) {
	if statement.Validate() != nil {
		return [sha256.Size]byte{}, ErrContract
	}
	return statement.binding.continuation, nil
}
