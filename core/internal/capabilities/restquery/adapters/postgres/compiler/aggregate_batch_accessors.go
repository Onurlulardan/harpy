package compiler

import (
	"crypto/sha256"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/google/uuid"
)

func (batch AggregateBatch) TenantID() (uuid.UUID, error) {
	if batch.Validate() != nil || batch.binding.scope != resolved.SnapshotScopeTenant {
		return uuid.Nil, ErrContract
	}
	return batch.binding.scopeID, nil
}

func (batch AggregateBatch) ActorID() (uuid.UUID, error) {
	if batch.Validate() != nil || batch.binding.scope != resolved.SnapshotScopeTenant {
		return uuid.Nil, ErrContract
	}
	return batch.binding.subject, nil
}

func (batch AggregateBatch) AuthorizationContextSHA256() ([sha256.Size]byte, error) {
	if batch.Validate() != nil {
		return [sha256.Size]byte{}, ErrContract
	}
	return batch.binding.authorization, nil
}

func (batch AggregateBatch) MandatorySHA256() ([sha256.Size]byte, error) {
	if batch.Validate() != nil {
		return [sha256.Size]byte{}, ErrContract
	}
	return batch.binding.mandatory, nil
}

func (batch AggregateBatch) Catalog() (api.CatalogReference, error) {
	if batch.Validate() != nil {
		return api.CatalogReference{}, ErrContract
	}
	return batch.binding.catalog, nil
}

func (batch AggregateBatch) QueryIdentitySHA256() ([sha256.Size]byte, error) {
	if batch.Validate() != nil {
		return [sha256.Size]byte{}, ErrContract
	}
	return batch.binding.identity, nil
}

func (batch AggregateBatch) ContinuationSHA256() ([sha256.Size]byte, error) {
	if batch.Validate() != nil {
		return [sha256.Size]byte{}, ErrContract
	}
	return batch.binding.continuation, nil
}

func (batch AggregateBatch) ValidatePlan(plan resolved.AuthorizedPlan) error {
	binding, err := aggregatePlanBinding(plan)
	if batch.Validate() != nil || err != nil || binding != batch.binding {
		return ErrContract
	}
	return nil
}
