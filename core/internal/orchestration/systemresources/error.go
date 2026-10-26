package systemresources

import (
	"errors"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
)

func normalizeError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, ErrContract),
		errors.Is(err, publicresource.ErrContract):
		return ErrContract
	case errors.Is(err, ErrForbidden),
		errors.Is(err, publicresource.ErrForbidden):
		return ErrForbidden
	case errors.Is(err, ErrNotFound),
		errors.Is(err, publicresource.ErrNotFound):
		return ErrNotFound
	case errors.Is(err, ErrConflict),
		errors.Is(err, publicresource.ErrConflict):
		return ErrConflict
	case errors.Is(err, layoutapi.ErrInvalid),
		errors.Is(err, layoutapi.ErrNotFound),
		errors.Is(err, layoutapi.ErrConflict),
		errors.Is(err, layoutapi.ErrUnavailable),
		errors.Is(err, publicresource.ErrUnavailable):
		return ErrUnavailable
	default:
		return ErrUnavailable
	}
}
