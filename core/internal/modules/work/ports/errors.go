package ports

import (
	"errors"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
)

var (
	ErrInvalid         = errors.New("WORK_INVALID")
	ErrForbidden       = errors.New("WORK_FORBIDDEN")
	ErrNotFound        = errors.New("WORK_NOT_FOUND")
	ErrConflict        = errors.New("WORK_VERSION_CONFLICT")
	ErrUnavailable     = errors.New("WORK_UNAVAILABLE")
	ErrProjectionStale = errors.New("WORK_PERMISSION_REVISION_STALE")
	ErrOrderingInvalid = domain.ErrOrderingInvalid
	ErrBoardLimit      = domain.ErrBoardLimit
)
