package iam

import (
	"errors"

	identityfacade "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/systemresource"
	userapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/user"
	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/systemresource"
)

func identityApplicationError(err error) error {
	if err == nil {
		return nil
	}
	var queryFailure queryapp.Failure
	if errors.As(err, &queryFailure) {
		return err
	}
	switch {
	case errors.Is(err, ErrIdentityHTTPInvalid),
		errors.Is(err, ErrIdentityHTTPForbidden),
		errors.Is(err, ErrIdentityHTTPNotFound),
		errors.Is(err, ErrIdentityHTTPConflict),
		errors.Is(err, ErrIdentityHTTPRecoveryRequired),
		errors.Is(err, ErrIdentityHTTPUnavailable):
		return err
	case errors.Is(err, userapp.ErrInvalidCommand),
		errors.Is(err, userapp.ErrInvalidQuery):
		return ErrIdentityHTTPInvalid
	case errors.Is(err, profileapp.ErrRecoveryRequired):
		return ErrIdentityHTTPRecoveryRequired
	case errors.Is(err, identityfacade.ErrBulkDeleteConflict):
		return ErrIdentityHTTPConflict
	default:
		return ErrIdentityHTTPUnavailable
	}
}
