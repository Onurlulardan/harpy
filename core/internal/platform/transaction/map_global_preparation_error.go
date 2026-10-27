package transaction

import (
	"errors"

	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

func mapGlobalPreparationError(err error) error {
	switch {
	case errors.Is(err, postgres.ErrGlobalIdentityMismatch):
		return ErrGlobalIdentity
	case errors.Is(err, postgres.ErrAuthorityMismatch):
		return ErrGlobalAuthority
	default:
		return ErrGlobalScopeDatabase
	}
}
