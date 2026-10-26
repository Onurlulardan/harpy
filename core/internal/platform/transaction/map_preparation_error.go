package transaction

import (
	"errors"

	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

func mapPreparationError(err error) error {
	switch {
	case errors.Is(err, postgres.ErrTenantMissing):
		return ErrTenantMissing
	case errors.Is(err, postgres.ErrTenantInactive):
		return ErrTenantInactive
	case errors.Is(err, postgres.ErrTenantSchemaInvalid),
		errors.Is(err, postgres.ErrTenantIdentityMismatch):
		return ErrTenantIdentity
	case errors.Is(err, postgres.ErrAuthorityMismatch):
		return ErrWriterAuthority
	default:
		return ErrScopeDatabase
	}
}
