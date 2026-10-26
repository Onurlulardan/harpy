package systemtransfercleanup

import (
	"errors"

	app "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
)

type sqlStateError interface{ SQLState() string }

func mapError(err error, notFound rowNotFound) error {
	if err == nil {
		return nil
	}
	if notFound != nil && notFound(err) {
		return app.ErrSourceCleanupStale
	}
	var state sqlStateError
	if !errors.As(err, &state) {
		return errors.Join(app.ErrSourceCleanupUnavailable, err)
	}
	switch state.SQLState() {
	case "22023", "23514", "42501":
		return app.ErrSourceCleanupInvalid
	case "P6044", "P6045", "P6046", "P6049":
		return app.ErrSourceCleanupStale
	default:
		return errors.Join(app.ErrSourceCleanupUnavailable, err)
	}
}
