package platformiampostgres

import (
	"errors"

	platformdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/platformiam"
)

type postgresStateError interface{ SQLState() string }

func storeError(err error) error {
	if err == nil {
		return nil
	}
	var state postgresStateError
	if !errors.As(err, &state) {
		return platformdomain.ErrStorage
	}
	switch state.SQLState() {
	case "42501":
		return platformdomain.ErrDenied
	case "40001", "23505", "23514", "P1401":
		return platformdomain.ErrConflict
	case "P1404":
		return platformdomain.ErrNotFound
	default:
		return platformdomain.ErrStorage
	}
}
