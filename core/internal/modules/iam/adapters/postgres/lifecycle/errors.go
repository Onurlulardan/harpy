package postgres

import (
	"errors"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycle"
)

type postgresStateError interface{ SQLState() string }

func mapError(err error) error {
	if err == nil {
		return nil
	}
	if rowNotFound(err) {
		return lifecycle.ErrNotFound
	}
	var state postgresStateError
	if !errors.As(err, &state) {
		return lifecycle.ErrUnavailable
	}
	switch state.SQLState() {
	case "P0002":
		return lifecycle.ErrNotFound
	case "40001":
		return lifecycle.ErrStale
	case "42501":
		return lifecycle.ErrDenied
	case "23505":
		return lifecycle.ErrConflict
	case "22023", "23514", "23503", "23502":
		return lifecycle.ErrInvalid
	default:
		return lifecycle.ErrUnavailable
	}
}
