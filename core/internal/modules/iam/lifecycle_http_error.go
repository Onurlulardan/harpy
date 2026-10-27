package iam

import (
	"errors"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycle"
)

func mapLifecycleHTTPError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, ErrLifecycleHTTPInvalid), errors.Is(err, lifecycle.ErrInvalid):
		return ErrLifecycleHTTPInvalid
	case errors.Is(err, ErrLifecycleHTTPForbidden), errors.Is(err, lifecycle.ErrDenied):
		return ErrLifecycleHTTPForbidden
	case errors.Is(err, lifecycle.ErrNotFound):
		return ErrLifecycleHTTPNotFound
	case errors.Is(err, lifecycle.ErrConflict), errors.Is(err, lifecycle.ErrStale):
		return ErrLifecycleHTTPConflict
	default:
		return ErrLifecycleHTTPUnavailable
	}
}
