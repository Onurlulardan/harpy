package postgres

import "github.com/Onurlulardan/harpy/core/internal/modules/audit/application"

type sqlStateError interface{ SQLState() string }

func mapError(err error, notFound func(error) bool) error {
	if err == nil {
		return nil
	}
	if notFound != nil && notFound(err) {
		return application.ErrNotFound
	}
	state, ok := err.(sqlStateError)
	if !ok {
		return application.ErrUnavailable
	}
	switch state.SQLState() {
	case "22023", "23502", "23514":
		return application.ErrInvalid
	case "42501":
		return application.ErrForbidden
	case "P6801":
		return application.ErrNotFound
	case "P6802", "23505", "40001":
		return application.ErrConflict
	default:
		return application.ErrUnavailable
	}
}
