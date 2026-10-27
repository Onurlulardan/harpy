package platformlicense

import (
	"errors"

	platformlicenseapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/platformlicense"
)

type sqlStateError interface{ SQLState() string }

func classify(err error) error {
	if err == nil {
		return nil
	}
	var state sqlStateError
	if errors.As(err, &state) {
		switch state.SQLState() {
		case "02000", "P0002":
			return platformlicenseapp.ErrNotFound
		case "22023":
			return platformlicenseapp.ErrInvalid
		case "23505":
			return platformlicenseapp.ErrIdempotencyReused
		case "55000":
			return platformlicenseapp.ErrConflict
		case "55P03":
			return platformlicenseapp.ErrInProgress
		case "42501":
			return platformlicenseapp.ErrForbidden
		}
	}
	return platformlicenseapp.ErrUnavailable
}
