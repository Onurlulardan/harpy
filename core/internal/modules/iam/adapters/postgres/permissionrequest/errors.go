package postgres

import (
	"errors"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/permissionrequest"
)

type postgresStateError interface {
	SQLState() string
}

func mapPermissionRequestError(err error) error {
	if err == nil {
		return nil
	}
	var postgresError postgresStateError
	if !errors.As(err, &postgresError) {
		return permissionrequest.ErrUnavailable
	}
	switch postgresError.SQLState() {
	case "P0002":
		return permissionrequest.ErrNotFound
	case "40001":
		return permissionrequest.ErrStale
	case "42501":
		return permissionrequest.ErrDenied
	case "23505":
		return permissionrequest.ErrConflict
	case "22023", "23514", "23503", "23502":
		return permissionrequest.ErrInvalid
	default:
		return permissionrequest.ErrUnavailable
	}
}
