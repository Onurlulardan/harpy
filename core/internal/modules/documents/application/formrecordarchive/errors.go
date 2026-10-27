package formrecordarchive

import (
	"errors"

	formrecordarchiveapi "github.com/Onurlulardan/harpy/core/internal/modules/documents/api/formrecordarchive"
	archiveapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/archive"
)

func mapError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, formrecordarchiveapi.ErrInvalid),
		errors.Is(err, archiveapp.ErrInvalid), errors.Is(err, archiveapp.ErrLimit):
		return formrecordarchiveapi.ErrInvalid
	case errors.Is(err, formrecordarchiveapi.ErrForbidden),
		errors.Is(err, archiveapp.ErrForbidden):
		return formrecordarchiveapi.ErrForbidden
	case errors.Is(err, formrecordarchiveapi.ErrReferenceMismatch):
		return formrecordarchiveapi.ErrReferenceMismatch
	case errors.Is(err, formrecordarchiveapi.ErrNotFound),
		errors.Is(err, archiveapp.ErrNotFound):
		return formrecordarchiveapi.ErrNotFound
	case errors.Is(err, archiveapp.ErrConflict), errors.Is(err, archiveapp.ErrCancelled):
		return formrecordarchiveapi.ErrConflict
	default:
		return formrecordarchiveapi.ErrUnavailable
	}
}
