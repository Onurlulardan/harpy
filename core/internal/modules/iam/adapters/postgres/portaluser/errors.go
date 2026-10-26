package portaluserpostgres

import (
	"errors"

	portaluser "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/portaluser"
)

type sqlStateError interface{ SQLState() string }

func mapError(err error) error {
	if err == nil {
		return nil
	}
	var state sqlStateError
	if !errors.As(err, &state) {
		return portaluser.ErrUnavailable
	}
	switch state.SQLState() {
	case "P0002":
		return portaluser.ErrNotFound
	case "40001", "23505":
		return portaluser.ErrConflict
	case "42501":
		return portaluser.ErrForbidden
	case "22023", "23514", "23503", "23502":
		return portaluser.ErrInvalid
	default:
		return portaluser.ErrUnavailable
	}
}
