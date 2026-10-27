package documents

import (
	"errors"

	documentshttp "github.com/Onurlulardan/harpy/core/internal/modules/documents/adapters/http"
	adoptionapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/adoption"
	archiveapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/archive"
	archivedownloadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/archivedownload"
	copyapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/copy"
	downloadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/download"
	folderapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/folder"
	itemapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/item"
	uploadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/upload"
)

func mapHTTPApplicationError(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, folderapp.ErrInvalid), errors.Is(err, itemapp.ErrInvalid),
		errors.Is(err, uploadapp.ErrInvalid), errors.Is(err, downloadapp.ErrInvalid),
		errors.Is(err, copyapp.ErrInvalid), errors.Is(err, archiveapp.ErrInvalid),
		errors.Is(err, archivedownloadapp.ErrInvalid), errors.Is(err, adoptionapp.ErrInvalid):
		return documentshttp.ErrInvalid
	case errors.Is(err, folderapp.ErrForbidden), errors.Is(err, itemapp.ErrForbidden),
		errors.Is(err, uploadapp.ErrForbidden), errors.Is(err, downloadapp.ErrForbidden),
		errors.Is(err, copyapp.ErrForbidden), errors.Is(err, archiveapp.ErrForbidden),
		errors.Is(err, archivedownloadapp.ErrForbidden):
		return documentshttp.ErrForbidden
	case errors.Is(err, folderapp.ErrNotFound), errors.Is(err, itemapp.ErrNotFound),
		errors.Is(err, uploadapp.ErrNotFound), errors.Is(err, downloadapp.ErrNotFound),
		errors.Is(err, archiveapp.ErrNotFound), errors.Is(err, archivedownloadapp.ErrNotFound),
		errors.Is(err, archivedownloadapp.ErrExpired):
		return documentshttp.ErrNotFound
	case errors.Is(err, folderapp.ErrConflict), errors.Is(err, itemapp.ErrConflict),
		errors.Is(err, itemapp.ErrLegalHold), errors.Is(err, uploadapp.ErrConflict),
		errors.Is(err, uploadapp.ErrInProgress), errors.Is(err, uploadapp.ErrQuota),
		errors.Is(err, copyapp.ErrConflict), errors.Is(err, archiveapp.ErrConflict),
		errors.Is(err, archiveapp.ErrLimit), errors.Is(err, archiveapp.ErrCancelled),
		errors.Is(err, adoptionapp.ErrConflict), errors.Is(err, adoptionapp.ErrBlocked):
		return documentshttp.ErrConflict
	case errors.Is(err, uploadapp.ErrContentRejected):
		return documentshttp.ErrContentRejected
	case errors.Is(err, uploadapp.ErrScanUnavailable):
		return documentshttp.ErrScanUnavailable
	default:
		return documentshttp.ErrUnavailable
	}
}
