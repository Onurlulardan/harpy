package postgres

import (
	"errors"

	"github.com/Onurlulardan/harpy/core/internal/modules/usage/domain"
)

type sqlStateError interface{ SQLState() string }

func (store *Store) normalize(err error) error {
	var rowNotFound func(error) bool
	if store != nil {
		rowNotFound = store.rowNotFound
	}
	return normalizePostgresError(err, rowNotFound)
}

func normalizePostgresError(err error, rowNotFound func(error) bool) error {
	if err == nil {
		return nil
	}
	for _, known := range []error{domain.ErrAuthentication, domain.ErrForbidden,
		domain.ErrInvalid, domain.ErrNotFound, domain.ErrConflict,
		domain.ErrRetentionHeld, domain.ErrSourceDrift, domain.ErrUnavailable} {
		if errors.Is(err, known) {
			return known
		}
	}
	if rowNotFound != nil && rowNotFound(err) {
		return domain.ErrNotFound
	}
	var state sqlStateError
	if errors.As(err, &state) {
		switch state.SQLState() {
		case "22023", "23502", "23514", "P6901":
			return domain.ErrInvalid
		case "42501", "P6903":
			return domain.ErrForbidden
		case "P6904":
			return domain.ErrNotFound
		case "P6902", "23505", "40001":
			return domain.ErrConflict
		case "P6905":
			return domain.ErrRetentionHeld
		case "P6906":
			return domain.ErrSourceDrift
		}
	}
	return domain.ErrUnavailable
}
