package iam

import (
	"errors"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
)

func identitySystemResourceError(err error) error {
	if err == nil {
		return nil
	}
	normalized := identityApplicationError(err)
	switch {
	case errors.Is(normalized, ErrIdentityHTTPForbidden):
		return publicresource.ErrForbidden
	case errors.Is(normalized, ErrIdentityHTTPNotFound):
		return publicresource.ErrNotFound
	case errors.Is(normalized, ErrIdentityHTTPConflict),
		errors.Is(normalized, ErrIdentityHTTPRecoveryRequired):
		return publicresource.ErrConflict
	case errors.Is(normalized, ErrIdentityHTTPInvalid):
		return publicresource.ErrContract
	}
	if status, ok := err.(interface{ HTTPStatus() int }); ok {
		switch status.HTTPStatus() {
		case 400, 422:
			return publicresource.ErrContract
		case 403:
			return publicresource.ErrForbidden
		case 404:
			return publicresource.ErrNotFound
		case 409:
			return publicresource.ErrConflict
		}
	}
	return publicresource.ErrUnavailable
}
