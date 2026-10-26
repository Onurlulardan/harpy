package compiler

import (
	"crypto/sha256"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/google/uuid"
)

func (statement AggregateStatement) AuthorizationContextSHA256() ([sha256.Size]byte, error) {
	if statement.Validate() != nil || statement.binding.scope != resolved.SnapshotScopeTenant {
		return [sha256.Size]byte{}, ErrContract
	}
	return statement.binding.authorization, nil
}

func (statement AggregateStatement) MandatorySHA256() ([sha256.Size]byte, error) {
	if statement.Validate() != nil {
		return [sha256.Size]byte{}, ErrContract
	}
	return statement.binding.mandatory, nil
}

func (statement AggregateStatement) QueryIdentitySHA256() ([sha256.Size]byte, error) {
	if statement.Validate() != nil {
		return [sha256.Size]byte{}, ErrContract
	}
	return statement.binding.identity, nil
}

func (statement AggregateStatement) ContinuationSHA256() ([sha256.Size]byte, error) {
	if statement.Validate() != nil {
		return [sha256.Size]byte{}, ErrContract
	}
	return statement.binding.continuation, nil
}

func (statement AggregateStatement) TenantID() (uuid.UUID, error) {
	if statement.Validate() != nil {
		return uuid.Nil, ErrContract
	}
	return statement.binding.scopeID, nil
}

func (statement AggregateStatement) ActorID() (uuid.UUID, error) {
	if statement.Validate() != nil || statement.binding.scope != resolved.SnapshotScopeTenant {
		return uuid.Nil, ErrContract
	}
	return statement.binding.subject, nil
}

func (statement AggregateStatement) Catalog() (api.CatalogReference, error) {
	if statement.Validate() != nil {
		return api.CatalogReference{}, ErrContract
	}
	return statement.binding.catalog, nil
}

func (statement AggregateStatement) Owner() (string, error) {
	if statement.Validate() != nil {
		return "", ErrContract
	}
	owner, err := statement.binding.catalog.Owner()
	if err != nil {
		return "", ErrContract
	}
	return owner, nil
}

func (statement AggregateStatement) ValidatePlan(
	plan resolved.AuthorizedPlan,
) error {
	binding, err := aggregatePlanBinding(plan)
	if statement.Validate() != nil || err != nil || binding != statement.binding {
		return ErrContract
	}
	return nil
}
