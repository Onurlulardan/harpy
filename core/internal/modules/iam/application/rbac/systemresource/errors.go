package systemresource

import rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"

var (
	ErrConflict        = rbacdomain.ErrConflict
	ErrDenied          = rbacdomain.ErrDenied
	ErrInvalidMutation = rbacdomain.ErrInvalidMutation
	ErrInvalidQuery    = rbacdomain.ErrInvalidQuery
	ErrNotFound        = rbacdomain.ErrNotFound
	ErrStorage         = rbacdomain.ErrStorage
)
