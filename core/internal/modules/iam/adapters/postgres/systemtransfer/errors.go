package systemtransfer

import (
	"errors"

	app "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
)

type postgresStateError interface{ SQLState() string }

func mapError(err error, notFound func(error) bool) error {
	if err == nil {
		return nil
	}
	if notFound != nil && notFound(err) {
		return app.ErrControlNotFound
	}
	var postgresError postgresStateError
	if errors.As(err, &postgresError) {
		switch postgresError.SQLState() {
		case "P6101":
			return app.ErrControlCapacity
		case "P6102", "23505", "40001":
			return app.ErrControlConflict
		case "P6103":
			return app.ErrControlExpired
		case "22023", "23514", "42501":
			return app.ErrControlInvalid
		case "P0002":
			return app.ErrControlNotFound
		}
	}
	return app.ErrControlUnavailable
}
