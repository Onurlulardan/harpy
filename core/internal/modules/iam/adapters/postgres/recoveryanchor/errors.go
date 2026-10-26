package recoveryanchorpostgres

import (
	"errors"

	anchordomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/recoveryanchor"
)

type stateError interface{ SQLState() string }

func storeError(err error, rowNotFound func(error) bool) error {
	if err == nil {
		return nil
	}
	if rowNotFound != nil && rowNotFound(err) {
		return anchordomain.ErrNotFound
	}
	var state stateError
	if !errors.As(err, &state) {
		return anchordomain.ErrDependency
	}
	switch state.SQLState() {
	case "42501":
		return anchordomain.ErrDenied
	case "40001", "23505", "23514", "55000", "P1401":
		return anchordomain.ErrConflict
	case "P0002", "P1404":
		return anchordomain.ErrNotFound
	default:
		return anchordomain.ErrDependency
	}
}
