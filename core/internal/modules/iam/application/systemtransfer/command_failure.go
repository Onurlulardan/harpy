package systemtransfer

import (
	"errors"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
)

func (worker *Worker) commandFailure(row SourceRow, err error) (RowReceipt, error) {
	switch {
	case errors.Is(err, publicresource.ErrConflict),
		errors.Is(err, publicresource.ErrNotFound):
		return rejectedReceipt(row, "REVISION_CONFLICT")
	case errors.Is(err, publicresource.ErrContract):
		return rejectedReceipt(row, "ROW_VALIDATION_FAILED")
	case errors.Is(err, publicresource.ErrForbidden):
		return rejectedReceipt(row, "COMMAND_FORBIDDEN")
	default:
		return RowReceipt{}, publicresource.ErrUnavailable
	}
}
