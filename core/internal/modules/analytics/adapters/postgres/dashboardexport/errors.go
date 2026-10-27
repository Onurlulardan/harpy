package dashboardexport

import exportapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboardexport"

type sqlStateError interface{ SQLState() string }

func mapError(err error, notFound rowNotFound) error {
	if err == nil {
		return nil
	}
	if notFound != nil && notFound(err) {
		return exportapp.ErrNotFound
	}
	state, ok := err.(sqlStateError)
	if !ok {
		return exportapp.ErrUnavailable
	}
	switch state.SQLState() {
	case "22023", "22P02", "23514":
		return exportapp.ErrInvalid
	case "42501":
		return exportapp.ErrForbidden
	case "P6102":
		return exportapp.ErrNotFound
	case "55000":
		return exportapp.ErrNotReady
	case "P6103":
		return exportapp.ErrIdempotencyReused
	case "P6110":
		return exportapp.ErrQueueLimit
	case "40001", "P6104", "23505", "55P03":
		return exportapp.ErrConflict
	default:
		return exportapp.ErrUnavailable
	}
}
