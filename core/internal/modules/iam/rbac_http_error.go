package iam

import (
	"errors"

	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/rbac/systemresource"
)

func rbacHTTPError(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, ErrRBACHTTPInvalid),
		errors.Is(err, ErrRBACHTTPForbidden),
		errors.Is(err, ErrRBACHTTPNotFound),
		errors.Is(err, ErrRBACHTTPConflict),
		errors.Is(err, ErrRBACHTTPUnavailable):
		return err
	case errors.Is(err, rbacdomain.ErrInvalidMutation),
		errors.Is(err, rbacdomain.ErrInvalidQuery):
		return ErrRBACHTTPInvalid
	case errors.Is(err, rbacdomain.ErrDenied):
		return ErrRBACHTTPForbidden
	case errors.Is(err, rbacdomain.ErrNotFound):
		return ErrRBACHTTPNotFound
	case errors.Is(err, rbacdomain.ErrConflict):
		return ErrRBACHTTPConflict
	default:
		return ErrRBACHTTPUnavailable
	}
}
