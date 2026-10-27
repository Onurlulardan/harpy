package iam

import (
	"context"
	"errors"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
)

func systemTransferError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return err
	case errors.Is(err, systemtransfer.ErrControlInvalid),
		errors.Is(err, publicresource.ErrContract):
		return ErrSystemTransferInvalid
	case errors.Is(err, systemtransfer.ErrControlNotFound),
		errors.Is(err, publicresource.ErrNotFound):
		return ErrSystemTransferNotFound
	case errors.Is(err, systemtransfer.ErrControlConflict),
		errors.Is(err, publicresource.ErrConflict):
		return ErrSystemTransferConflict
	case errors.Is(err, systemtransfer.ErrControlCapacity):
		return ErrSystemTransferBusy
	case errors.Is(err, systemtransfer.ErrControlExpired):
		return ErrSystemTransferExpired
	case errors.Is(err, publicresource.ErrForbidden):
		return ErrSystemTransferForbidden
	default:
		return ErrSystemTransferUnavailable
	}
}
