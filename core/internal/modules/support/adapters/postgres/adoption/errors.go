package adoption

import (
	"errors"
	"fmt"

	application "github.com/Onurlulardan/harpy/core/internal/modules/support/application/adoption"
)

type sqlStateError interface{ SQLState() string }

func normalize(err error) error {
	if err == nil {
		return nil
	}
	for _, known := range []error{application.ErrInvalid, application.ErrSourceDrift,
		application.ErrConflict, application.ErrLease, application.ErrUnavailable} {
		if errors.Is(err, known) {
			return known
		}
	}
	var state sqlStateError
	if errors.As(err, &state) {
		switch state.SQLState() {
		case "22023", "22P02", "23514":
			return application.ErrInvalid
		case "P6512":
			return application.ErrSourceDrift
		case "P6513":
			return application.ErrLease
		case "P6511", "P6514", "23505", "40001":
			return application.ErrConflict
		}
	}
	return fmt.Errorf("%w: %v", application.ErrUnavailable, err)
}
