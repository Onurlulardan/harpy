package knowledgefile

import (
	"errors"

	knowledgeapi "github.com/Onurlulardan/harpy/core/internal/modules/documents/api/knowledgefile"
	downloadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/download"
	folderapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/folder"
	linkapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/link"
	uploadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/upload"
)

var errUnavailable = errors.New("Documents knowledge-file internal dependency failed")

func mapError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, knowledgeapi.ErrInvalid), errors.Is(err, folderapp.ErrInvalid),
		errors.Is(err, uploadapp.ErrInvalid):
		return knowledgeapi.ErrInvalid
	case errors.Is(err, knowledgeapi.ErrForbidden), errors.Is(err, folderapp.ErrForbidden),
		errors.Is(err, uploadapp.ErrForbidden), errors.Is(err, downloadapp.ErrForbidden):
		return knowledgeapi.ErrForbidden
	case errors.Is(err, knowledgeapi.ErrConflict), errors.Is(err, folderapp.ErrConflict),
		errors.Is(err, uploadapp.ErrConflict), errors.Is(err, linkapp.ErrConflict):
		return knowledgeapi.ErrConflict
	default:
		return knowledgeapi.ErrUnavailable
	}
}
