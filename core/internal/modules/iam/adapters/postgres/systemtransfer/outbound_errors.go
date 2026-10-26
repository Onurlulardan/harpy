package systemtransfer

import (
	"errors"

	app "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
)

func mapOutboundError(err error, conflict error) error {
	if err == nil {
		return nil
	}
	var database postgresStateError
	if errors.As(err, &database) {
		switch database.SQLState() {
		case "P6046":
			return app.ErrExportCancelled
		case "P6047":
			return app.ErrExportLeaseLost
		case "P6101":
			return app.ErrExportCapacity
		case "42501":
			return app.ErrExportAuthorizationRevoked
		case "P6044", "P6045", "P6048", "23505", "23514":
			if conflict != nil {
				return conflict
			}
			return app.ErrExportStateConflict
		case "22023":
			return app.ErrExportWorkerInvalid
		}
	}
	return app.ErrExportUnavailable
}
