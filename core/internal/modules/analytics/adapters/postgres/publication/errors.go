package publication

import publicationapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/publication"

type sqlStateError interface{ SQLState() string }

func mapError(err error, notFound rowNotFound) error {
	if err == nil {
		return nil
	}
	if notFound != nil && notFound(err) {
		return publicationapp.ErrNotFound
	}
	state, ok := err.(sqlStateError)
	if !ok {
		return publicationapp.ErrUnavailable
	}
	switch state.SQLState() {
	case "22023", "22P02", "23514":
		return publicationapp.ErrInvalid
	case "42501":
		return publicationapp.ErrForbidden
	case "P6102":
		return publicationapp.ErrNotFound
	case "P6105":
		return publicationapp.ErrRateLimited
	case "40001", "P6103", "P6104", "55000":
		return publicationapp.ErrConflict
	default:
		return publicationapp.ErrUnavailable
	}
}
