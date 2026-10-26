package postgres

import (
	"errors"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/application"
)

type sqlStateError interface{ SQLState() string }

func (store *Store) normalize(err error) error {
	if err == nil {
		return nil
	}
	for _, known := range []error{application.ErrAuthentication, application.ErrForbidden,
		application.ErrInvalid, application.ErrNotFound, application.ErrConflict,
		application.ErrRateLimited, application.ErrIdempotencyMismatch,
		application.ErrRevisionStale, application.ErrUnavailable} {
		if errors.Is(err, known) {
			return known
		}
	}
	if store != nil && store.rowNotFound != nil && store.rowNotFound(err) {
		return application.ErrNotFound
	}
	var state sqlStateError
	if errors.As(err, &state) {
		switch state.SQLState() {
		case "P6501", "22023", "23514":
			return application.ErrInvalid
		case "P6502":
			return application.ErrIdempotencyMismatch
		case "P6503":
			return application.ErrRevisionStale
		case "P6504":
			return application.ErrNotFound
		case "40001", "23505":
			return application.ErrConflict
		}
	}
	return application.ErrUnavailable
}
