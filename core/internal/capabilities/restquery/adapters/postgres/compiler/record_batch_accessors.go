package compiler

import (
	"crypto/sha256"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/google/uuid"
)

func (batch RecordBatch) CountMode() (api.CountMode, error) {
	if batch.Validate() != nil || batch.binding.scope != resolved.SnapshotScopeTenant {
		return 0, ErrContract
	}
	return batch.countMode, nil
}

func (batch RecordBatch) TenantID() (uuid.UUID, error) {
	if batch.Validate() != nil {
		return uuid.Nil, ErrContract
	}
	return batch.binding.scopeID, nil
}

func (batch RecordBatch) ActorID() (uuid.UUID, error) {
	if batch.Validate() != nil || batch.binding.scope != resolved.SnapshotScopeTenant {
		return uuid.Nil, ErrContract
	}
	return batch.binding.subject, nil
}

func (batch RecordBatch) AuthorizationContextSHA256() ([sha256.Size]byte, error) {
	if batch.Validate() != nil {
		return [sha256.Size]byte{}, ErrContract
	}
	return batch.binding.authorization, nil
}

func (batch RecordBatch) MandatorySHA256() ([sha256.Size]byte, error) {
	if batch.Validate() != nil {
		return [sha256.Size]byte{}, ErrContract
	}
	return batch.binding.mandatory, nil
}

func (batch RecordBatch) Catalog() (api.CatalogReference, error) {
	if batch.Validate() != nil {
		return api.CatalogReference{}, ErrContract
	}
	return batch.binding.catalog, nil
}

func (batch RecordBatch) QueryIdentitySHA256() ([sha256.Size]byte, error) {
	if batch.Validate() != nil {
		return [sha256.Size]byte{}, ErrContract
	}
	return batch.binding.identity, nil
}

func (batch RecordBatch) ContinuationSHA256() ([sha256.Size]byte, error) {
	if batch.Validate() != nil {
		return [sha256.Size]byte{}, ErrContract
	}
	return batch.binding.continuation, nil
}

func (batch RecordBatch) ValidatePlan(plan resolved.AuthorizedPlan) error {
	binding, count, err := recordPlanBinding(plan)
	if batch.Validate() != nil || err != nil || binding != batch.binding ||
		count != batch.countMode {
		return ErrContract
	}
	return nil
}
