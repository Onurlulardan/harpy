package iam

import (
	"errors"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/permissionrequest"
)

func permissionRequestHTTPError(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, ErrPermissionRequestHTTPInvalid),
		errors.Is(err, ErrPermissionRequestHTTPForbidden),
		errors.Is(err, ErrPermissionRequestHTTPNotFound),
		errors.Is(err, ErrPermissionRequestHTTPConflict),
		errors.Is(err, ErrPermissionRequestHTTPUnavailable):
		return err
	case errors.Is(err, permissionrequest.ErrInvalid):
		return ErrPermissionRequestHTTPInvalid
	case errors.Is(err, permissionrequest.ErrDenied):
		return ErrPermissionRequestHTTPForbidden
	case errors.Is(err, permissionrequest.ErrNotFound):
		return ErrPermissionRequestHTTPNotFound
	case errors.Is(err, permissionrequest.ErrStale),
		errors.Is(err, permissionrequest.ErrConflict):
		return ErrPermissionRequestHTTPConflict
	default:
		return ErrPermissionRequestHTTPUnavailable
	}
}
