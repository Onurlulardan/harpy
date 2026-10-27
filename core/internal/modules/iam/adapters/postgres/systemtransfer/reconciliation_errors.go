package systemtransfer

import (
	"errors"

	app "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
)

func mapExportReconciliationError(err error) error {
	if err == nil {
		return nil
	}
	var database postgresStateError
	if errors.As(err, &database) {
		switch database.SQLState() {
		case "P6047":
			return app.ErrExportLeaseLost
		case "P6044", "P6045", "P6048", "23505":
			return app.ErrExportStateConflict
		case "22023", "23514", "42501":
			return app.ErrExportReconciliationInvalid
		}
	}
	return app.ErrExportUnavailable
}
