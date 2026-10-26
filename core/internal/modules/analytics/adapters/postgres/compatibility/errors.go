package compatibility

import compatibilityapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/compatibility"

type sqlStateError interface{ SQLState() string }

func mapError(err error) error {
	if err == nil {
		return nil
	}
	state, ok := err.(sqlStateError)
	if !ok {
		return compatibilityapp.ErrUnavailable
	}
	switch state.SQLState() {
	case "22023", "22P02", "23514", "42501":
		return compatibilityapp.ErrInvalid
	case "P6102":
		return compatibilityapp.ErrNotFound
	case "P6103", "40001", "55000":
		return compatibilityapp.ErrConflict
	case "P6104":
		return compatibilityapp.ErrBlocked
	default:
		return compatibilityapp.ErrUnavailable
	}
}
