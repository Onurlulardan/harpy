package platformops

import platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"

type PermissionCacheFlush = platformopsapi.PermissionCacheFlush
type PermissionCacheResult = platformopsapi.PermissionCacheResult
type PermissionCacheOperations = platformopsapi.PermissionCacheOperations

var (
	ErrInvalid     = platformopsapi.ErrInvalid
	ErrConflict    = platformopsapi.ErrConflict
	ErrUnavailable = platformopsapi.ErrUnavailable
)
