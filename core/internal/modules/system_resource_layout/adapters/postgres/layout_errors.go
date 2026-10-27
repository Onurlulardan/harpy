package postgres

import layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"

type sqlStateError interface {
	SQLState() string
}

func mapLayoutError(err error) error {
	if err == nil {
		return nil
	}
	state, ok := err.(sqlStateError)
	if !ok {
		return layoutUnavailable()
	}
	switch state.SQLState() {
	case "P0002":
		return layoutapi.ErrNotFound
	case "23505", "40001":
		return layoutapi.ErrConflict
	case "22023", "23514":
		return layoutapi.ErrInvalid
	default:
		return layoutUnavailable()
	}
}

func layoutUnavailable() error { return layoutapi.ErrUnavailable }
