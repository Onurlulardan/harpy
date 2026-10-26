package application

import "github.com/Onurlulardan/harpy/core/internal/modules/work/ports"

var (
	ErrInvalid         = ports.ErrInvalid
	ErrForbidden       = ports.ErrForbidden
	ErrNotFound        = ports.ErrNotFound
	ErrConflict        = ports.ErrConflict
	ErrUnavailable     = ports.ErrUnavailable
	ErrProjectionStale = ports.ErrProjectionStale
)
