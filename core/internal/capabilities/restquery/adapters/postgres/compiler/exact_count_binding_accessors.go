package compiler

import (
	"crypto/sha256"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/google/uuid"
)

func (statement ExactCountStatement) TenantID() (uuid.UUID, error) {
	if statement.Validate() != nil || statement.binding.scope != resolved.SnapshotScopeTenant {
		return uuid.Nil, ErrContract
	}
	return statement.binding.scopeID, nil
}

func (statement ExactCountStatement) ActorID() (uuid.UUID, error) {
	if statement.Validate() != nil || statement.binding.scope != resolved.SnapshotScopeTenant {
		return uuid.Nil, ErrContract
	}
	return statement.binding.subject, nil
}

func (statement ExactCountStatement) AuthorizationContextSHA256() (
	[sha256.Size]byte, error,
) {
	if statement.Validate() != nil {
		return [sha256.Size]byte{}, ErrContract
	}
	return statement.binding.authorization, nil
}

func (statement ExactCountStatement) MandatorySHA256() ([sha256.Size]byte, error) {
	if statement.Validate() != nil {
		return [sha256.Size]byte{}, ErrContract
	}
	return statement.binding.mandatory, nil
}

func (statement ExactCountStatement) Catalog() (api.CatalogReference, error) {
	if statement.Validate() != nil {
		return api.CatalogReference{}, ErrContract
	}
	return statement.binding.catalog, nil
}

func (statement ExactCountStatement) Owner() (string, error) {
	if statement.Validate() != nil {
		return "", ErrContract
	}
	owner, err := statement.binding.catalog.Owner()
	if err != nil {
		return "", ErrContract
	}
	return owner, nil
}

func (statement ExactCountStatement) QueryIdentitySHA256() (
	[sha256.Size]byte, error,
) {
	if statement.Validate() != nil {
		return [sha256.Size]byte{}, ErrContract
	}
	return statement.binding.identity, nil
}

func (statement ExactCountStatement) ContinuationSHA256() (
	[sha256.Size]byte, error,
) {
	if statement.Validate() != nil {
		return [sha256.Size]byte{}, ErrContract
	}
	return statement.binding.continuation, nil
}

func (statement ExactCountStatement) ValidatePlan(
	plan resolved.AuthorizedPlan,
) error {
	binding, count, err := recordPlanBinding(plan)
	if statement.Validate() != nil || err != nil ||
		count != api.CountModeExact || binding != statement.binding {
		return ErrContract
	}
	return nil
}
