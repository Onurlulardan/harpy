package systemtransfer

import "errors"

import systemtransferapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"

func mapWorkerError(err error) error {
	if err == nil {
		return nil
	}
	var database postgresStateError
	if errors.As(err, &database) {
		switch database.SQLState() {
		case "P6046":
			return systemtransferapp.ErrWorkerLeaseLost
		case "P6044", "P6045", "23505", "23514", "22023":
			return systemtransferapp.ErrWorkerState
		case "42501":
			return systemtransferapp.ErrAuthorizationCheckpoint
		case "P6101", "40001":
			return systemtransferapp.ErrWorkerUnavailable
		}
	}
	return systemtransferapp.ErrWorkerUnavailable
}
