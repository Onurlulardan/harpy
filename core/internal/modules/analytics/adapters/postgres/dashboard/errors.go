package dashboard

import dashboardapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboard"

type sqlStateError interface{ SQLState() string }

func mapError(err error, notFound rowNotFound) error {
	if err == nil {
		return nil
	}
	if notFound != nil && notFound(err) {
		return dashboardapp.ErrNotFound
	}
	state, ok := err.(sqlStateError)
	if !ok {
		return dashboardapp.ErrUnavailable
	}
	switch state.SQLState() {
	case "22023", "22P02", "23514":
		return dashboardapp.ErrInvalid
	case "42501":
		return dashboardapp.ErrForbidden
	case "P6102":
		return dashboardapp.ErrNotFound
	case "40001", "P6103", "P6104", "55000":
		return dashboardapp.ErrConflict
	default:
		return dashboardapp.ErrUnavailable
	}
}
