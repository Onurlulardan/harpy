package compiler

import (
	"crypto/sha256"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/google/uuid"
)

func (statement RecordStatement) TenantID() (uuid.UUID, error) {
	if statement.Validate() != nil || statement.binding.scope != resolved.SnapshotScopeTenant {
		return uuid.Nil, ErrContract
	}
	return statement.binding.scopeID, nil
}

func (statement RecordStatement) ActorID() (uuid.UUID, error) {
	if statement.Validate() != nil || statement.binding.scope != resolved.SnapshotScopeTenant {
		return uuid.Nil, ErrContract
	}
	return statement.binding.subject, nil
}

func (statement RecordStatement) AuthorizationContextSHA256() ([sha256.Size]byte, error) {
	if statement.Validate() != nil {
		return [sha256.Size]byte{}, ErrContract
	}
	return statement.binding.authorization, nil
}

func (statement RecordStatement) MandatorySHA256() ([sha256.Size]byte, error) {
	if statement.Validate() != nil {
		return [sha256.Size]byte{}, ErrContract
	}
	return statement.binding.mandatory, nil
}

func (statement RecordStatement) Catalog() (api.CatalogReference, error) {
	if statement.Validate() != nil {
		return api.CatalogReference{}, ErrContract
	}
	return statement.binding.catalog, nil
}

func (statement RecordStatement) Owner() (string, error) {
	if statement.Validate() != nil {
		return "", ErrContract
	}
	owner, err := statement.binding.catalog.Owner()
	if err != nil {
		return "", ErrContract
	}
	return owner, nil
}

func (statement RecordStatement) QueryIdentitySHA256() ([sha256.Size]byte, error) {
	if statement.Validate() != nil {
		return [sha256.Size]byte{}, ErrContract
	}
	return statement.binding.identity, nil
}

func (statement RecordStatement) ContinuationSHA256() ([sha256.Size]byte, error) {
	if statement.Validate() != nil {
		return [sha256.Size]byte{}, ErrContract
	}
	return statement.binding.continuation, nil
}
