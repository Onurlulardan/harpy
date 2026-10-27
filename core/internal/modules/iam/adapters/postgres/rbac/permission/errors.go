package permissionpostgres

import (
	"errors"

	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
)

type postgresStateError interface{ SQLState() string }

func mapError(err error, invalid error) error {
	if rowNotFound(err) {
		return rbacdomain.ErrNotFound
	}
	var state postgresStateError
	if !errors.As(err, &state) {
		return rbacdomain.ErrStorage
	}
	switch state.SQLState() {
	case "P0002":
		return rbacdomain.ErrNotFound
	case "40001", "23505", "55000":
		return rbacdomain.ErrConflict
	case "42501":
		return rbacdomain.ErrDenied
	case "22023", "23502", "23503", "23514":
		return invalid
	default:
		return rbacdomain.ErrStorage
	}
}
