package adoption

import (
	"errors"

	"github.com/Onurlulardan/harpy/core/internal/modules/usage/domain"
)

type sqlStateError interface{ SQLState() string }

func normalize(err error) error {
	if err == nil {
		return nil
	}
	for _, known := range []error{domain.ErrInvalid, domain.ErrConflict,
		domain.ErrNotFound, domain.ErrSourceDrift, domain.ErrUnavailable} {
		if errors.Is(err, known) {
			return known
		}
	}
	var state sqlStateError
	if errors.As(err, &state) {
		switch state.SQLState() {
		case "22023", "23502", "23514", "P6901":
			return domain.ErrInvalid
		case "23505", "40001", "P6902":
			return domain.ErrConflict
		case "P6904":
			return domain.ErrNotFound
		case "P6906":
			return domain.ErrSourceDrift
		}
	}
	return domain.ErrUnavailable
}
