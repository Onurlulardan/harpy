package query

import queryapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/query"

type sqlStateError interface{ SQLState() string }

func mapError(err error, notFound rowNotFound) error {
	if err == nil {
		return nil
	}
	if notFound != nil && notFound(err) {
		return queryapp.ErrNotFound
	}
	state, ok := err.(sqlStateError)
	if !ok {
		return queryapp.ErrUnavailable
	}
	switch state.SQLState() {
	case "22023", "23514":
		return queryapp.ErrInvalid
	case "42501":
		return queryapp.ErrForbidden
	case "P6102":
		return queryapp.ErrNotFound
	case "40001", "P6103", "P6104", "55000":
		return queryapp.ErrConflict
	default:
		return queryapp.ErrUnavailable
	}
}
